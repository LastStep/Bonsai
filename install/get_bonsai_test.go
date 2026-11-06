package install_test

// The one-command installers, get-bonsai.sh (Linux and WSL) and get-bonsai.ps1 (Windows PowerShell 5.1): each is
// self-documenting and parses; each holds the install lines' repository and workflow (engine's constants); and each
// refuses what it must before it downloads anything, then stops at the first check that fails and never runs the
// release's installer unless every check passed. get-bonsai.sh runs on Linux against stubs of curl and gh on a narrowed
// PATH (no network), with a release built here whose install.sh only leaves a mark. get-bonsai.ps1's runs are on
// Windows only and stop before any download: it calls Windows' own curl.exe by its full path, which no test can stub,
// so the downloads and checks after them are proved by hand against a real release (T-6020's run report).

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/engine"
)

var getScripts = []string{"get-bonsai.sh", "get-bonsai.ps1"}

// Each script is ASCII text with LF line ends, documents itself (what it does, usage from main's raw file, needs, exit
// codes, the way back), names the repository and workflow the install lines hold an archive to, runs gh attestation
// verify with the lines' flags, evaluates no text as code, and names no person's folder.
func TestGetBonsaiScriptsDocumented(t *testing.T) {
	for _, name := range getScripts {
		// .gitattributes keeps both LF on every checkout (*.sh, and get-bonsai.ps1's own line); a CRLF checkout of the
		// .ps1 made before that line, which PowerShell runs as well, is read as LF.
		s := read(t, name)
		if name == "get-bonsai.ps1" {
			s = strings.ReplaceAll(s, "\r\n", "\n")
		}
		for i := 0; i < len(s); i++ {
			if c := s[i]; c == '\r' || c > '~' || (c < ' ' && c != '\n' && c != '\t') {
				t.Fatalf("%s holds a byte that is not ASCII text with LF line ends at offset %d: %q", name, i, c)
			}
		}
		for _, want := range []string{
			name + ": installs a Bonsai release",
			"What it does:", "Usage", "Needs:", "Exit 0", "The way back:",
			"https://raw.githubusercontent.com/" + engine.ReleaseRepo + "/main/install/" + name,
			"attestation", "verify", "--signer-workflow", "--source-ref", "--deny-self-hosted-runners",
			"It has no option that skips a check.",
		} {
			// The header read as prose: sh's comment lines joined.
			if !strings.Contains(strings.ReplaceAll(s, "\n# ", " "), want) {
				t.Errorf("%s does not hold %q", name, want)
			}
		}
		repo := map[string]string{"get-bonsai.sh": "\nrepo=" + engine.ReleaseRepo + "\n", "get-bonsai.ps1": "\n$Repo = '" + engine.ReleaseRepo + "'\n"}[name]
		if !strings.Contains(s, repo) {
			t.Errorf("%s does not hold %q", name, strings.TrimSpace(repo))
		}
		// The scripts spell the workflow as the repository and its path, which is engine's ReleaseWorkflow.
		wf := map[string]string{"get-bonsai.sh": "\nworkflow=$repo/.github/workflows/release.yml\n", "get-bonsai.ps1": "\n$Workflow = \"$Repo/.github/workflows/release.yml\"\n"}[name]
		if engine.ReleaseWorkflow != engine.ReleaseRepo+"/.github/workflows/release.yml" || !strings.Contains(s, wf) {
			t.Errorf("%s does not name the release workflow %s", name, engine.ReleaseWorkflow)
		}
		low := strings.ToLower(s)
		for _, bad := range []string{"eval ", "invoke-expression", "iex ", "/home/", `c:\users`, "bonsai-checks", "--insecure", "-k "} {
			if strings.Contains(low, bad) {
				t.Errorf("%s holds %q", name, bad)
			}
		}
	}
}

// sh -n reads get-bonsai.sh without running it, wherever an sh is found.
func TestGetBonsaiShParses(t *testing.T) {
	sh, err := exec.LookPath("sh")
	if err != nil {
		t.Skip("no sh on the PATH: get-bonsai.sh is a POSIX sh script, parsed where one is (Linux and WSL, CI's linux jobs)")
	}
	if out, err := exec.Command(sh, "-n", "get-bonsai.sh").CombinedOutput(); err != nil {
		t.Fatalf("sh -n get-bonsai.sh: %v\n%s", err, out)
	}
}

// PowerShell's own parser reads get-bonsai.ps1 with no error, where Windows PowerShell (or pwsh) is found.
func TestGetBonsaiPs1Parses(t *testing.T) {
	ps := ""
	if runtime.GOOS == "windows" {
		p := filepath.Join(os.Getenv("SystemRoot"), "System32", "WindowsPowerShell", "v1.0", "powershell.exe")
		if _, err := os.Stat(p); err == nil {
			ps = p
		}
	}
	if ps == "" {
		if p, err := exec.LookPath("pwsh"); err == nil {
			ps = p
		}
	}
	if ps == "" {
		t.Skip("no Windows PowerShell or pwsh here: get-bonsai.ps1 is parsed where one is (CI's windows jobs)")
	}
	abs, err := filepath.Abs("get-bonsai.ps1")
	if err != nil {
		t.Fatal(err)
	}
	code := "$e = $null; [void][Management.Automation.Language.Parser]::ParseFile(" + psQuote(abs) + ", [ref]$null, [ref]$e); " +
		"foreach ($x in $e) { [Console]::Out.WriteLine($x.ToString()) }; exit $e.Count"
	out, err := exec.Command(ps, "-NoProfile", "-NonInteractive", "-Command", code).CombinedOutput()
	if err != nil {
		t.Fatalf("get-bonsai.ps1 does not parse: %v\n%s", err, out)
	}
}

// --- get-bonsai.sh, on Linux, against stubs ---

// getSh is one run's folders: bin (the only folder on the PATH: wrappers of the real tools and stubs of gh and curl),
// rel (the files the curl stub serves, by name), and the logs the stubs write. steps is the order the checks and the
// install ran in, one line each: sha256sum (the SHA-256 check), attestation (gh attestation verify), tar (the
// unpacking) and install.sh (the release's installer).
type getSh struct {
	tmp, bin, rel, curlLog, ghLog, mark, steps string
	env                                        []string
}

// stepped are the tools whose wrapper writes its name to the steps log before it runs.
var stepped = map[string]bool{"sha256sum": true, "tar": true}

// The tools get-bonsai.sh, the stubs and a release's install.sh stub run, found on the real PATH.
var getTools = []string{"sh", "awk", "id", "uname", "mktemp", "sha256sum", "tar", "gzip", "cat"}

const curlStub = `#!/bin/sh
for a in "$@"; do url=$a; done
printf '%s\n' "$url" >>"$GET_CURL_LOG"
name=${url##*/}
[ -f "$GET_REL/$name" ] || { echo "curl: (22) The requested URL returned error: 404" >&2; exit 22; }
cat "$GET_REL/$name" >"$name"
`

const ghStub = `#!/bin/sh
printf '%s\n' "$*" >>"$GET_GH_LOG"
case $1 in
--version) printf 'gh version %s (2026-09-30)\nhttps://github.com/cli/cli/releases/tag/v%s\n' "$GET_GH_VERSION" "$GET_GH_VERSION" ;;
auth) exit "$GET_GH_AUTH" ;;
api) printf '%s\n' "$GET_GH_LATEST" ;;
attestation) echo attestation >>"$GET_STEPS"; exit "$GET_GH_ATTEST" ;;
*) exit 9 ;;
esac
`

func newGetSh(t *testing.T, skip ...string) *getSh {
	t.Helper()
	linuxOnly(t)
	if os.Getuid() == 0 {
		t.Skip("get-bonsai.sh refuses to run as root, and these runs are its runs as a person")
	}
	tmp := t.TempDir()
	g := &getSh{tmp: tmp, bin: filepath.Join(tmp, "bin"), rel: filepath.Join(tmp, "release"),
		curlLog: filepath.Join(tmp, "curl.log"), ghLog: filepath.Join(tmp, "gh.log"), mark: filepath.Join(tmp, "installed"),
		steps: filepath.Join(tmp, "steps.log")}
	for _, d := range []string{g.bin, g.rel, filepath.Join(tmp, "tmp")} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	skipped := map[string]bool{}
	for _, s := range skip {
		skipped[s] = true
	}
	for _, tool := range getTools {
		if skipped[tool] {
			continue
		}
		real, err := exec.LookPath(tool)
		if err != nil {
			t.Skip(tool + " is not on the PATH")
		}
		log := ""
		if stepped[tool] {
			log = "echo " + tool + " >>\"$GET_STEPS\"\n"
		}
		g.stub(t, tool, "#!/bin/sh\n"+log+"exec '"+real+"' \"$@\"\n")
	}
	if !skipped["curl"] {
		g.stub(t, "curl", curlStub)
	}
	if !skipped["gh"] {
		g.stub(t, "gh", ghStub)
	}
	g.env = []string{"PATH=" + g.bin, "HOME=" + filepath.Join(tmp, "person"), "TMPDIR=" + filepath.Join(tmp, "tmp"), "LC_ALL=C",
		"GET_CURL_LOG=" + g.curlLog, "GET_GH_LOG=" + g.ghLog, "GET_REL=" + g.rel, "GET_MARK=" + g.mark, "GET_STEPS=" + g.steps,
		"GET_GH_VERSION=2.102.0", "GET_GH_AUTH=0", "GET_GH_ATTEST=0", "GET_GH_LATEST=v9.9.9"}
	return g
}

func (g *getSh) stub(t *testing.T, name, body string) {
	t.Helper()
	write(t, filepath.Join(g.bin, name), []byte(body), 0o755)
}

// release puts a release of version ver for arch in rel: its archive (an install.sh that leaves the mark, beside a
// bonsai) and a checksums.txt listing it, its SHA-256 the archive's unless the bytes given differ.
func (g *getSh) release(t *testing.T, ver, arch string, sums func(name, sum string) string) string {
	t.Helper()
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(zw)
	for name, body := range map[string]string{
		"install.sh": "#!/bin/sh\necho install.sh >>\"$GET_STEPS\"\nprintf 'installed\\n' >\"$GET_MARK\"\necho 'install.sh: done'\n",
		"bonsai":     "not a program\n",
	} {
		if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0o755, Size: int64(len(body)), Typeflag: tar.TypeReg}); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(body)); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	name := "bonsai_" + ver + "_linux_" + arch + ".tar.gz"
	write(t, filepath.Join(g.rel, name), buf.Bytes(), 0o644)
	h := sha256.Sum256(buf.Bytes())
	s := hex.EncodeToString(h[:])
	list := s + "  " + name + "\n" + strings.Repeat("0", 64) + "  bonsai_" + ver + "_windows_" + arch + ".zip\n"
	if sums != nil {
		list = sums(name, s)
	}
	write(t, filepath.Join(g.rel, "checksums.txt"), []byte(list), 0o644)
	return name
}

func (g *getSh) run(t *testing.T, env []string, args ...string) (int, string) {
	t.Helper()
	sh, err := exec.LookPath("sh")
	if err != nil {
		t.Fatal(err)
	}
	abs, err := filepath.Abs("get-bonsai.sh")
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(sh, append([]string{abs}, args...)...)
	cmd.Dir = g.tmp
	cmd.Env = append(append([]string{}, g.env...), env...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			t.Fatal(err)
		}
	}
	return cmd.ProcessState.ExitCode(), string(out)
}

func (g *getSh) log(name string) string {
	b, _ := os.ReadFile(name)
	return string(b)
}

func (g *getSh) installed() bool {
	_, err := os.Stat(g.mark)
	return err == nil
}

// With every check passing, the named version's archive and checksums.txt are fetched from its release, the
// provenance check runs with the install lines' flags for its tag, and the release's install.sh runs; the newest
// release is the one gh names when no version is given, and a v before the version is the same version.
func TestGetBonsaiShInstalls(t *testing.T) {
	for _, c := range []struct {
		name string
		args []string
	}{{"named", []string{"9.9.9"}}, {"named with v", []string{"v9.9.9"}}, {"newest", nil}} {
		t.Run(c.name, func(t *testing.T) {
			g := newGetSh(t)
			archive := g.release(t, "9.9.9", "amd64", nil)
			code, out := g.run(t, nil, c.args...)
			if code != 0 || !g.installed() {
				t.Fatalf("exit %d, installed %v; want 0 and the release's install.sh run\n%s", code, g.installed(), out)
			}
			base := "https://github.com/" + engine.ReleaseRepo + "/releases/download/v9.9.9/"
			if got, want := g.log(g.curlLog), base+archive+"\n"+base+"checksums.txt\n"; got != want {
				t.Errorf("curl fetched\n%s\nwant\n%s", got, want)
			}
			attest := "attestation verify " + archive + " -R " + engine.ReleaseRepo + " --signer-workflow " + engine.ReleaseWorkflow +
				" --source-ref refs/tags/v9.9.9 --deny-self-hosted-runners\n"
			if !strings.Contains(g.log(g.ghLog), "\n"+attest) {
				t.Errorf("gh was not called as the install lines call it (%q); its calls:\n%s", attest, g.log(g.ghLog))
			}
			if latest := strings.Contains(g.log(g.ghLog), "api repos/"+engine.ReleaseRepo+"/releases/latest --jq .tag_name"); latest != (c.args == nil) {
				t.Errorf("gh api releases/latest called %v; want %v", latest, c.args == nil)
			}
			for _, want := range []string{archive + ": OK", "Every check passed", "install.sh: done"} {
				if !strings.Contains(out, want) {
					t.Errorf("the output does not hold %q:\n%s", want, out)
				}
			}
			// Both checks, then the unpacking, then the installer: nothing from the archive is read before its origin is.
			if got, want := g.log(g.steps), "sha256sum\nattestation\ntar\ninstall.sh\n"; got != want {
				t.Errorf("the steps ran in the order\n%s\nwant\n%s", got, want)
			}
		})
	}
}

// The archive's name is the machine's: amd64 for x86_64, arm64 for aarch64 and arm64; any other machine, and any
// system but Linux, is refused before anything is fetched.
func TestGetBonsaiShMachines(t *testing.T) {
	unameStub := "#!/bin/sh\ncase $1 in -m) echo \"$GET_UNAME_M\" ;; -s) echo \"$GET_UNAME_S\" ;; *) exit 1 ;; esac\n"
	for _, c := range []struct{ s, m, arch, refused string }{
		{"Linux", "x86_64", "amd64", ""},
		{"Linux", "aarch64", "arm64", ""},
		{"Linux", "arm64", "arm64", ""},
		{"Linux", "mips", "", "no Bonsai build for this machine (mips)"},
		{"Darwin", "arm64", "", "this machine runs Darwin, not Linux"},
	} {
		t.Run(c.s+" "+c.m, func(t *testing.T) {
			g := newGetSh(t, "uname")
			g.stub(t, "uname", unameStub)
			env := []string{"GET_UNAME_S=" + c.s, "GET_UNAME_M=" + c.m}
			if c.refused != "" {
				code, out := g.run(t, env, "9.9.9")
				if code != 1 || !strings.Contains(out, c.refused) || g.log(g.curlLog) != "" || g.installed() {
					t.Fatalf("exit %d, curl called %q; want 1, %q, and nothing fetched\n%s", code, g.log(g.curlLog), c.refused, out)
				}
				return
			}
			archive := g.release(t, "9.9.9", c.arch, nil)
			if archive != "bonsai_9.9.9_linux_"+c.arch+".tar.gz" {
				t.Fatalf("the release's archive is %s", archive)
			}
			code, out := g.run(t, env, "9.9.9")
			if code != 0 || !strings.Contains(g.log(g.curlLog), "/v9.9.9/"+archive+"\n") || !g.installed() {
				t.Fatalf("exit %d, curl fetched %q; want 0 and %s\n%s", code, g.log(g.curlLog), archive, out)
			}
		})
	}
}

// A version that is not one (or the old product's 0.x, or more than one) is refused before gh or curl is called.
func TestGetBonsaiShRefusesVersions(t *testing.T) {
	for _, c := range []struct {
		args    []string
		refused string
	}{
		{[]string{"latest"}, "'latest' is not a release version"},
		{[]string{"x"}, "'x' is not a release version"},
		{[]string{"1.0"}, "'1.0' is not a release version"},
		{[]string{"1.0.3-rc1"}, "'1.0.3-rc1' is not a release version"},
		{[]string{"vv1.0.3"}, "'vv1.0.3' is not a release version"},
		{[]string{"0.4.3"}, "the old product (0.x)"},
		{[]string{""}, "the version given is not a release version"},
		{[]string{"1.0.3; touch x"}, "the version given is not a release version"},
		{[]string{"1.0.3\n2.0.0"}, "the version given is not a release version"},
		{[]string{"1.0.3", "1.0.2"}, "one version at most"},
	} {
		t.Run(strings.Join(c.args, " "), func(t *testing.T) {
			g := newGetSh(t)
			code, out := g.run(t, nil, c.args...)
			if code != 1 || !strings.Contains(out, c.refused) || !strings.Contains(out, "\nnext: ") {
				t.Fatalf("exit %d; want 1 and %q with its next step\n%s", code, c.refused, out)
			}
			if g.log(g.ghLog) != "" || g.log(g.curlLog) != "" || g.installed() {
				t.Fatalf("gh or curl was called: %q %q", g.log(g.ghLog), g.log(g.curlLog))
			}
		})
	}
}

// What the machine lacks is refused, naming its next step, before anything is fetched: a tool missing, gh older than
// 2.97 or not logged in, and a newest release gh cannot name.
func TestGetBonsaiShRefusesMachine(t *testing.T) {
	for _, c := range []struct {
		name    string
		skip    []string
		env     []string
		args    []string
		refused string
	}{
		{"no gh", []string{"gh"}, nil, []string{"9.9.9"}, "gh, GitHub's command line, is not on the PATH"},
		{"no curl", []string{"curl"}, nil, []string{"9.9.9"}, "curl is not on the PATH"},
		{"no sha256sum", []string{"sha256sum"}, nil, []string{"9.9.9"}, "sha256sum is not on the PATH"},
		{"no tar", []string{"tar"}, nil, []string{"9.9.9"}, "tar is not on the PATH"},
		{"gh 2.96", nil, []string{"GET_GH_VERSION=2.96.1"}, []string{"9.9.9"}, "gh is 2.96.1, older than 2.97"},
		{"gh 1.99", nil, []string{"GET_GH_VERSION=1.99.0"}, []string{"9.9.9"}, "older than 2.97"},
		{"gh unnamed", nil, []string{"GET_GH_VERSION="}, []string{"9.9.9"}, "older than 2.97"},
		{"gh logged out", nil, []string{"GET_GH_AUTH=1"}, []string{"9.9.9"}, "gh is not logged in"},
		{"newest unread", nil, []string{"GET_GH_LATEST="}, nil, "could not read the newest release's version"},
		{"newest not a version", nil, []string{"GET_GH_LATEST=nightly"}, nil, "could not read the newest release's version"},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := newGetSh(t, c.skip...)
			g.release(t, "9.9.9", "amd64", nil)
			code, out := g.run(t, c.env, c.args...)
			if code != 1 || !strings.Contains(out, c.refused) || !strings.Contains(out, "\nnext: ") {
				t.Fatalf("exit %d; want 1 and %q with its next step\n%s", code, c.refused, out)
			}
			if g.log(g.curlLog) != "" || g.installed() {
				t.Fatalf("curl was called (%q), or install.sh ran", g.log(g.curlLog))
			}
		})
	}
	t.Run("as root", func(t *testing.T) {
		g := newGetSh(t, "id")
		g.stub(t, "id", "#!/bin/sh\necho 0\n")
		code, out := g.run(t, nil, "9.9.9")
		if code != 1 || !strings.Contains(out, "it runs as root") || g.log(g.ghLog) != "" || g.log(g.curlLog) != "" {
			t.Fatalf("exit %d; want 1, refused as root before gh or curl\n%s", code, out)
		}
	})
}

// Each check that fails after the download stops the run there: the release's install.sh never runs, and no later
// check is reached (a SHA-256 that fails never reaches gh attestation verify).
func TestGetBonsaiShStopsAtFirstFailure(t *testing.T) {
	for _, c := range []struct {
		name    string
		sums    func(name, sum string) string
		env     []string
		noRel   bool
		refused string
		attest  bool
	}{
		{name: "no such release", noRel: true, refused: "downloading https://github.com/" + engine.ReleaseRepo + "/releases/download/v9.9.9/bonsai_9.9.9_linux_amd64.tar.gz failed"},
		{name: "wrong sha256", sums: func(n, s string) string { return strings.Repeat("a", 64) + "  " + n + "\n" }, refused: "SHA-256 is not the one checksums.txt holds"},
		{name: "not listed", sums: func(n, s string) string { return s + "  other.tar.gz\n" }, refused: "does not hold exactly one line for"},
		{name: "listed twice", sums: func(n, s string) string { return s + "  " + n + "\n" + s + "  " + n + "\n" }, refused: "does not hold exactly one line for"},
		{name: "listed by a longer name", sums: func(n, s string) string { return s + "  x" + n + "\n" }, refused: "does not hold exactly one line for"},
		{name: "provenance fails", env: []string{"GET_GH_ATTEST=1"}, refused: "gh attestation verify did not pass", attest: true},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := newGetSh(t)
			if !c.noRel {
				g.release(t, "9.9.9", "amd64", c.sums)
			}
			code, out := g.run(t, c.env, "9.9.9")
			if code != 1 || !strings.Contains(out, c.refused) || !strings.Contains(out, "nothing was installed") {
				t.Fatalf("exit %d; want 1 and %q, nothing installed\n%s", code, c.refused, out)
			}
			if g.installed() || strings.Contains(out, "Every check passed") {
				t.Fatalf("the release's install.sh ran after a failed check\n%s", out)
			}
			if got := strings.Contains(g.log(g.ghLog), "attestation verify"); got != c.attest {
				t.Fatalf("gh attestation verify called %v; want %v", got, c.attest)
			}
			if steps := g.log(g.steps); strings.Contains(steps, "tar\n") || strings.Contains(steps, "install.sh\n") {
				t.Fatalf("the archive was unpacked or its installer run after a failed check; the steps:\n%s", steps)
			}
		})
	}
}

// In each script's text the unpacking and the installer come after both checks, and the download after every check
// of the machine (the refusal of root or an elevated PowerShell among them): get-bonsai.ps1's order after its
// download is read here, since no test can stub the curl.exe it runs by its full path.
func TestGetBonsaiOrder(t *testing.T) {
	for name, marks := range map[string][]string{
		"get-bonsai.sh": {`if [ "$(id -u)" = 0 ]; then`, "gh auth status", "\ncurl -fsSLO \"$base/$archive\"", "\nsha256sum -c archive.sha256",
			"\ngh attestation verify ", "\ntar -xzf ", "\nsh ./install.sh"},
		"get-bonsai.ps1": {"gh @('auth', 'status')", "\nif ($admin) {", "(Invoke-Native $curl @('-fsSLO', \"$base/$archive\"))", "\nif ($got -cne $want[0]) {",
			"(Invoke-Native $gh @('attestation', 'verify'", "\n    Expand-Archive ", "(Invoke-Native $powershell @("},
	} {
		s := strings.ReplaceAll(read(t, name), "\r\n", "\n")
		at := -1
		for _, m := range marks {
			if strings.Count(s, m) != 1 {
				t.Errorf("%s holds %q %d times; want once", name, m, strings.Count(s, m))
				continue
			}
			i := strings.Index(s, m)
			if i < at {
				t.Errorf("%s: %q comes before the step that should precede it", name, m)
			}
			at = i
		}
	}
}

// --- get-bonsai.ps1, on Windows, before any download ---

// Windows' own folders only on the PATH (windowsOnly), with a stub gh.cmd where one is given: a bad version, a machine
// with no build, gh missing, older than 2.97 or logged out, and an elevated PowerShell (where the test runs elevated),
// each refused with its next step before curl.exe runs.
func TestGetBonsaiPs1Refusals(t *testing.T) {
	ps := windowsOnly(t)
	abs, err := filepath.Abs("get-bonsai.ps1")
	if err != nil {
		t.Fatal(err)
	}
	ghStubCmd := func(auth string) string {
		return "@echo off\r\nif \"%1\"==\"--version\" goto v\r\nexit /b " + auth + "\r\n:v\r\necho gh version 2.102.0 (2026-09-30)\r\nexit /b 0\r\n"
	}
	ghOK := ghStubCmd("0")
	for _, c := range []struct {
		name, gh, arch, refused string
		args                    []string
		elevated                bool
	}{
		{name: "latest", gh: ghOK, args: []string{"-Version", "latest"}, refused: "'latest' is not a release version"},
		{name: "x", gh: ghOK, args: []string{"x"}, refused: "'x' is not a release version"},
		{name: "0.x", gh: ghOK, args: []string{"-Version", "0.4.3"}, refused: "the old product (0.x)"},
		{name: "bad characters", gh: ghOK, args: []string{"-Version", "1.0.3;x"}, refused: "the version given is not a release version"},
		{name: "x86", gh: ghOK, arch: "x86", args: []string{"9.9.9"}, refused: "no Bonsai build for this machine (x86)"},
		{name: "no gh", args: []string{"9.9.9"}, refused: "gh, GitHub's command line, is not on the PATH"},
		{name: "gh 2.96", gh: "@echo gh version 2.96.0 (2026-01-01)\r\n", args: []string{"9.9.9"}, refused: "gh is 2.96, older than 2.97"},
		{name: "gh logged out", gh: ghStubCmd("1"), args: []string{"9.9.9"}, refused: "gh is not logged in"},
		{name: "elevated", gh: ghOK, args: []string{"9.9.9"}, refused: "it runs in an elevated PowerShell (as administrator)", elevated: true},
	} {
		t.Run(c.name, func(t *testing.T) {
			if c.elevated && !psElevated(t, ps) {
				// Not elevated, every check passes and the run would go on to download: no test raises a Windows prompt.
				t.Skip("this PowerShell is not elevated, so the refusal cannot show; CI's windows jobs run elevated")
			}
			tmp := t.TempDir()
			if c.gh != "" {
				bin := filepath.Join(tmp, "bin")
				if err := os.MkdirAll(bin, 0o755); err != nil {
					t.Fatal(err)
				}
				write(t, filepath.Join(bin, "gh.cmd"), []byte(c.gh), 0o755)
				t.Setenv("PATH", bin+";"+os.Getenv("PATH"))
			}
			arch := c.arch
			if arch == "" {
				arch = "AMD64"
			}
			t.Setenv("PROCESSOR_ARCHITECTURE", arch)
			// Unset, not empty: an empty PROCESSOR_ARCHITEW6432 in its environment leaves Windows PowerShell 5.1 reading
			// PROCESSOR_ARCHITECTURE as empty too.
			t.Setenv("PROCESSOR_ARCHITEW6432", "")
			if err := os.Unsetenv("PROCESSOR_ARCHITEW6432"); err != nil {
				t.Fatal(err)
			}
			t.Setenv("TEMP", tmp)
			cmd := exec.Command(ps, append([]string{"-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-File", abs}, c.args...)...)
			cmd.Dir = tmp
			raw, err := cmd.CombinedOutput()
			if err != nil {
				if _, ok := err.(*exec.ExitError); !ok {
					t.Fatal(err)
				}
			}
			out := strings.ReplaceAll(string(raw), "\r\n", "\n")
			if code := cmd.ProcessState.ExitCode(); code != 1 || !strings.Contains(out, c.refused) || !strings.Contains(out, "\nnext: ") {
				t.Fatalf("exit %d; want 1 and %q with its next step\n%s", code, c.refused, out)
			}
			if strings.Contains(out, "Downloading") {
				t.Fatalf("it reached the download\n%s", out)
			}
		})
	}
}

// psElevated reports whether a PowerShell started by this test runs as administrator, read as get-bonsai.ps1 reads it.
func psElevated(t *testing.T, ps string) bool {
	t.Helper()
	out, err := exec.Command(ps, "-NoProfile", "-NonInteractive", "-Command",
		"([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)").CombinedOutput()
	if err != nil {
		t.Fatalf("reading whether PowerShell is elevated: %v\n%s", err, out)
	}
	return strings.TrimSpace(string(out)) == "True"
}
