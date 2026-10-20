// Package install_test holds Bonsai's two installers (install.sh for Linux and WSL, install.ps1 for Windows) to the
// installed places and to install.json as check reads them, and runs each on scratch targets and a scratch Bonsai home
// (design/plan-5.md, 5.6.6): install, the record read by check, again, again while a copy runs, remove, and each
// refusal. install.sh's tests run on Linux and install.ps1's on Windows, each skipped elsewhere: each script is for its
// own system's shell. No test runs either script at its real place (spec section 3: no agent installs Bonsai); CI's
// install job does, on throwaway machines.
package install_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/engine"
	"github.com/LastStep/Bonsai/internal/testpack"
)

// windowsPlace is engine.PlaceWindows as Windows names it: /c/Program Files/... is C:\Program Files\....
func windowsPlace() string {
	p := engine.PlaceWindows
	return strings.ToUpper(p[1:2]) + ":" + strings.ReplaceAll(p[2:], "/", `\`)
}

// Each script holds the two installed places as engine's constants spell them, and the record's file and keys as
// check's reader reads them (the round trip through engine.Check is in the run tests): ASCII only, LF line ends.
func TestScriptsHoldPlacesAndRecord(t *testing.T) {
	sh := read(t, "install.sh")
	ps := read(t, "install.ps1")
	for name, s := range map[string]string{"install.sh": sh, "install.ps1": ps} {
		for i := 0; i < len(s); i++ {
			if c := s[i]; c == '\r' || (c > '~' || (c < ' ' && c != '\n' && c != '\t')) {
				t.Fatalf("%s holds a byte that is not ASCII text with LF line ends at offset %d: %q", name, i, c)
			}
		}
		for _, k := range []string{`"path"`, `"version"`, `"sha256"`, engine.InstallFile} {
			if !strings.Contains(s, k) {
				t.Errorf("%s does not write %s", name, k)
			}
		}
	}
	if !strings.Contains(sh, "\nplace="+engine.PlaceLinux+"\n") {
		t.Errorf("install.sh's place is not %s", engine.PlaceLinux)
	}
	if strings.Count(sh, engine.PlaceLinux) != strings.Count(sh, "/usr/local/bin/") {
		t.Errorf("install.sh names a /usr/local/bin path other than %s", engine.PlaceLinux)
	}
	w := windowsPlace()
	for _, line := range []string{"\n$Place = '" + w + "'\n", "\n    [string]$Target = '" + w + "',\n"} {
		if !strings.Contains(ps, line) {
			t.Errorf("install.ps1 does not hold %q", strings.TrimSpace(line))
		}
	}
}

func read(t *testing.T, name string) string {
	t.Helper()
	raw, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func exe(name string) string {
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}
	return name
}

var built struct {
	once sync.Once
	dir  string
	path string
	err  error
}

// TestMain removes the shared build once every test has run.
func TestMain(m *testing.M) {
	code := m.Run()
	if built.dir != "" {
		_ = os.RemoveAll(built.dir)
	}
	os.Exit(code)
}

// bonsai builds Bonsai once (a plain build, as a release's is) and returns its path.
func bonsai(t *testing.T) string {
	t.Helper()
	built.once.Do(func() {
		goTool, err := exec.LookPath("go") // go test puts its own Go first on the PATH
		if err != nil {
			built.err = err
			return
		}
		if built.dir, built.err = os.MkdirTemp("", "bonsai-install-build-"); built.err != nil {
			return
		}
		built.path = filepath.Join(built.dir, exe("bonsai"))
		cmd := exec.Command(goTool, "build", "-o", built.path, "./cmd/bonsai")
		cmd.Dir = ".."
		if out, err := cmd.CombinedOutput(); err != nil {
			built.err = errors.New(err.Error() + ": " + string(out))
		}
	})
	if built.err != nil {
		t.Fatalf("building bonsai: %v", built.err)
	}
	return built.path
}

// scratch is one test's folders: dist (the release's folder: the script and the bonsai beside it), bin (the target's
// folder), home (the Bonsai home) and root (a project check reads), with git and BONSAI_HOME isolated.
type scratch struct {
	tmp, dist, bin, home, root, script, src, target string
}

func newScratch(t *testing.T, script string, src []byte) *scratch {
	t.Helper()
	tmp := t.TempDir()
	s := &scratch{tmp: tmp, dist: filepath.Join(tmp, "dist"), bin: filepath.Join(tmp, "bin it's"), home: testpack.Isolate(t, tmp)}
	for _, d := range []string{s.dist, s.bin} {
		if err := os.MkdirAll(d, 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(d, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	s.script = filepath.Join(s.dist, script)
	s.src = filepath.Join(s.dist, exe("bonsai"))
	s.target = filepath.Join(s.bin, exe("bonsai"))
	write(t, s.script, []byte(read(t, script)), 0o644)
	write(t, s.src, src, 0o755)
	s.root = testpack.Project(t, tmp, "project")
	write(t, filepath.Join(s.root, "bonsai.yaml"), []byte("format: bonsai.workspace/1\nid: ws-aaaaaaaaaaaaaaaaaaaaaaaaaa\nname: demo\npacks: []\n"), 0o644)
	return s
}

func write(t *testing.T, p string, b []byte, mode os.FileMode) {
	t.Helper()
	if err := os.WriteFile(p, b, mode); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, p string) []byte {
	t.Helper()
	raw, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func sum(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

// record reads install.json: exactly the three keys, the path with forward slashes. install.sh records the target in
// its folder as that folder resolves (pwd -P); install.ps1 records its full path with an 8.3 short name in it made long
// (Windows PowerShell's GetFullPath does), the path a person sees. CI's Windows runner reaches its temporary folder as
// RUNNER~1, so the path wanted is the target's resolved form, engine.RealPath's, on both.
func (s *scratch) record(t *testing.T) map[string]string {
	t.Helper()
	var m map[string]string
	if err := json.Unmarshal(readFile(t, filepath.Join(s.home, engine.InstallFile)), &m); err != nil {
		t.Fatalf("install.json: %v", err)
	}
	want := s.recorded()
	if len(m) != 3 || m["path"] != want || m["sha256"] != sum(readFile(t, s.src)) || m["version"] == "" {
		t.Fatalf("install.json is %v; want path %s, the source's sha256 and a version", m, want)
	}
	return m
}

// recorded is the target's path as its installer records it in install.json (record's comment says why).
func (s *scratch) recorded() string {
	real, _ := engine.RealPath(s.target) // "" when the walk fails, which no record holds: the comparison fails
	return filepath.ToSlash(real)
}

// checkReads proves check's reader reads the record: with the target first on the PATH, no bonsai-path finding and no
// note that there is no record; with another bonsai first, a bonsai-path finding naming the installed path.
func (s *scratch) checkReads(t *testing.T) {
	t.Helper()
	path := os.Getenv("PATH")
	defer func() { t.Setenv("PATH", path) }()
	t.Setenv("PATH", s.bin+string(os.PathListSeparator)+path)
	r, err := engine.Check(s.root, s.home)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range r.Findings {
		if f.Code == "bonsai-path" {
			t.Fatalf("check found the installed bonsai wrong: %s", f.Sentence())
		}
	}
	for _, n := range r.Notes {
		if strings.Contains(n, engine.InstallFile) {
			t.Fatalf("check did not read the record: %s", n)
		}
	}
	other := filepath.Join(s.tmp, "other")
	if err := os.MkdirAll(other, 0o755); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(other, exe("bonsai")), []byte("not bonsai\n"), 0o755)
	t.Setenv("PATH", other+string(os.PathListSeparator)+s.bin+string(os.PathListSeparator)+path)
	if r, err = engine.Check(s.root, s.home); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, f := range r.Findings {
		found = found || (f.Code == "bonsai-path" && strings.Contains(f.Message, s.recorded()))
	}
	if !found {
		t.Fatalf("check gave no bonsai-path finding naming %s with another bonsai first on the PATH: %+v", s.target, r.Findings)
	}
}

// running starts the installed bonsai's hook guard waiting on its input (inside its 5 s budget) and returns a stop
// that ends its input and says how it ended.
func (s *scratch) running(t *testing.T) (alive func() bool, stop func() error) {
	t.Helper()
	cmd := exec.Command(s.target, "hook", "guard")
	cmd.Dir = s.root
	cmd.Env = append(os.Environ(), "CLAUDE_PROJECT_DIR="+s.root) // else the guard refuses at once
	in, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	cmd.Stdout, cmd.Stderr = io.Discard, io.Discard
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	var res error
	ended := false
	t.Cleanup(func() {
		_ = in.Close()
		if !ended {
			_ = cmd.Process.Kill()
			<-done
		}
	})
	time.Sleep(300 * time.Millisecond)
	alive = func() bool {
		select {
		case res = <-done:
			ended = true
		default:
		}
		return !ended
	}
	stop = func() error {
		_ = in.Close()
		if !ended {
			select {
			case res = <-done:
			case <-time.After(20 * time.Second):
				return errors.New("the running bonsai did not end")
			}
			ended = true
		}
		var ee *exec.ExitError
		if errors.As(res, &ee) && ee.ExitCode() < 0 {
			return errors.New("the running bonsai was killed: " + res.Error())
		}
		return nil
	}
	return alive, stop
}

// --- install.sh, on Linux ---

func linuxOnly(t *testing.T) {
	t.Helper()
	if runtime.GOOS != "linux" {
		t.Skip("install.sh is Linux's and WSL's installer, a POSIX sh script; install.ps1's tests hold Windows")
	}
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("no sh on the PATH")
	}
}

// stubs puts a folder first on the PATH holding a sudo that records it was called and fails, so no test reaches the
// real one, and the extra stubs given (name to script). It returns the file sudo's stub writes.
func (s *scratch) stubs(t *testing.T, extra map[string]string) string {
	t.Helper()
	dir := filepath.Join(s.tmp, "stubs")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	called := filepath.Join(s.tmp, "sudo-called")
	write(t, filepath.Join(dir, "sudo"), []byte("#!/bin/sh\necho \"$@\" >>'"+called+"'\nexit 1\n"), 0o755)
	for name, body := range extra {
		write(t, filepath.Join(dir, name), []byte(body), 0o755)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("HOME", filepath.Join(s.tmp, "person"))
	return called
}

func (s *scratch) sh(t *testing.T, args ...string) (int, string) {
	t.Helper()
	cmd := exec.Command("sh", append([]string{s.script, "--target", s.target}, args...)...)
	cmd.Dir = s.tmp
	out, err := cmd.CombinedOutput()
	var ee *exec.ExitError
	if err != nil && !errors.As(err, &ee) {
		t.Fatal(err)
	}
	return cmd.ProcessState.ExitCode(), string(out)
}

func noFile(t *testing.T, paths ...string) {
	t.Helper()
	for _, p := range paths {
		if _, err := os.Lstat(p); err == nil {
			t.Errorf("%s is there; want none", p)
		}
	}
}

// empty fails the test unless the folder holds nothing.
func empty(t *testing.T, dir string) {
	t.Helper()
	des, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(des) > 0 {
		t.Errorf("%s holds %d entries (the first %s); want none", dir, len(des), des[0].Name())
	}
}

func leftovers(t *testing.T, dir, pattern string) []string {
	t.Helper()
	m, err := filepath.Glob(filepath.Join(dir, pattern))
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func TestShInstallAgainRunningRemove(t *testing.T) {
	linuxOnly(t)
	s := newScratch(t, "install.sh", readFile(t, bonsai(t)))
	called := s.stubs(t, nil)

	code, out := s.sh(t)
	if code != 0 {
		t.Fatalf("install: exit %d\n%s", code, out)
	}
	if !bytes.Equal(readFile(t, s.target), readFile(t, s.src)) {
		t.Fatal("the installed file is not the source")
	}
	if fi, err := os.Stat(s.target); err != nil || fi.Mode().Perm() != 0o755 {
		t.Fatalf("the installed file's mode: %v %v", fi, err)
	}
	for _, want := range []string{sum(readFile(t, s.src)), s.target, "bonsai dev"} {
		if !strings.Contains(out, want) {
			t.Errorf("the install's output does not name %q:\n%s", want, out)
		}
	}
	s.record(t)
	rec := readFile(t, filepath.Join(s.home, engine.InstallFile))
	s.checkReads(t)

	if code, out := s.sh(t); code != 0 || !bytes.Equal(rec, readFile(t, filepath.Join(s.home, engine.InstallFile))) {
		t.Fatalf("a second install: exit %d, or its record changed\n%s", code, out)
	}

	alive, stop := s.running(t)
	if !alive() {
		t.Fatal("the installed bonsai's hook guard did not stay running on an open input")
	}
	if code, out := s.sh(t); code != 0 {
		t.Fatalf("an install while a copy runs: exit %d\n%s", code, out)
	}
	if !alive() {
		t.Fatal("the running bonsai ended during the install")
	}
	if err := stop(); err != nil {
		t.Fatal(err)
	}
	s.record(t)
	if l := leftovers(t, s.bin, ".bonsai.new.*"); len(l) > 0 {
		t.Fatalf("temporary copies left: %v", l)
	}

	write(t, filepath.Join(s.bin, ".bonsai.new.1"), []byte("left by a stopped install"), 0o644)
	if code, out := s.sh(t, "--remove"); code != 0 {
		t.Fatalf("remove: exit %d\n%s", code, out)
	}
	noFile(t, s.target, filepath.Join(s.bin, ".bonsai.new.1"), filepath.Join(s.home, engine.InstallFile))
	noFile(t, called)
}

// The target is read as its folder resolves, so another spelling of a place is that place (the real place's checks
// then hold however it is typed): the record holds the clean path.
func TestShTargetSpellings(t *testing.T) {
	linuxOnly(t)
	src := readFile(t, bonsai(t))
	for _, spell := range []string{"%s//bonsai", "%s/./bonsai", "%s/../bin it's/bonsai"} {
		t.Run(strings.ReplaceAll(spell, "%s", "bin"), func(t *testing.T) {
			s := newScratch(t, "install.sh", src)
			s.stubs(t, nil)
			clean := s.target
			s.target = strings.Replace(spell, "%s", s.bin, 1)
			if code, out := s.sh(t); code != 0 {
				t.Fatalf("install through %s: exit %d\n%s", s.target, code, out)
			}
			s.target = clean
			s.record(t)
		})
	}
}

func TestShRefusals(t *testing.T) {
	linuxOnly(t)
	real := readFile(t, bonsai(t))
	for _, c := range []struct {
		name, want string
		src        []byte
		stubs      map[string]string
		prepare    func(t *testing.T, s *scratch)
		after      func(t *testing.T, s *scratch)
	}{
		{name: "a bonsai that does not run", want: "does not run", src: []byte("#!/bin/sh\nexit 1\n")},
		{name: "a target folder others can write", want: "can be written by a group or by other users", src: real,
			prepare: func(t *testing.T, s *scratch) {
				if err := os.Chmod(s.bin, 0o777); err != nil {
					t.Fatal(err)
				}
			}},
		{name: "a system that is not Linux", want: "runs Darwin, not Linux", src: real,
			stubs: map[string]string{"uname": "#!/bin/sh\necho Darwin\n"}},
		{name: "run as root", want: "runs as root", src: real,
			stubs: map[string]string{"id": "#!/bin/sh\necho 0\n"}},
		{name: "a source changed between the hash and the copy", want: "SHA-256 differs", src: real,
			stubs: map[string]string{"install": "#!/bin/sh\nfor a in \"$@\"; do last=$a; done\nprintf 'another file' >\"$last\"\n"}},
		{name: "a relative target", want: "not an absolute path", src: real,
			prepare: func(t *testing.T, s *scratch) { s.target = "bin/bonsai" }},
		// The link is judged as given, before the folder is resolved: a linked folder is refused, its target untouched.
		{name: "a target folder that is a link", want: "is a link", src: real,
			prepare: func(t *testing.T, s *scratch) {
				dest := filepath.Join(s.tmp, "linked to")
				if err := os.Mkdir(dest, 0o700); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(dest, filepath.Join(s.tmp, "link")); err != nil {
					t.Fatal(err)
				}
				s.target = filepath.Join(s.tmp, "link", "bonsai")
			},
			after: func(t *testing.T, s *scratch) { empty(t, filepath.Join(s.tmp, "linked to")) }},
		{name: "a target that is a folder", want: "is a folder", src: real,
			prepare: func(t *testing.T, s *scratch) {
				if err := os.Mkdir(s.target, 0o700); err != nil {
					t.Fatal(err)
				}
			}},
	} {
		t.Run(c.name, func(t *testing.T) {
			s := newScratch(t, "install.sh", c.src)
			called := s.stubs(t, c.stubs)
			if c.prepare != nil {
				c.prepare(t, s)
			}
			code, out := s.sh(t)
			if code != 1 || !strings.Contains(out, c.want) || !strings.Contains(out, "\nnext: ") {
				t.Fatalf("exit %d; want 1 with %q and its next step:\n%s", code, c.want, out)
			}
			if fi, err := os.Stat(s.target); err == nil && fi.IsDir() {
				empty(t, s.target) // a target that is a folder: nothing copied into it
				code, out := s.sh(t, "--remove")
				if code != 1 || !strings.Contains(out, c.want) || !strings.Contains(out, "\nnext: ") {
					t.Fatalf("--remove: exit %d; want 1 with %q and its next step:\n%s", code, c.want, out)
				}
				empty(t, s.target)
			} else {
				noFile(t, filepath.Join(s.tmp, "bin it's", "bonsai"))
			}
			noFile(t, filepath.Join(s.home, engine.InstallFile), called)
			if c.after != nil {
				c.after(t, s)
			}
			if l := leftovers(t, filepath.Join(s.tmp, "bin it's"), ".bonsai.new.*"); len(l) > 0 {
				t.Fatalf("temporary copies left: %v", l)
			}
		})
	}
}

// --- install.ps1, on Windows ---

func windowsOnly(t *testing.T) string {
	t.Helper()
	if runtime.GOOS != "windows" {
		t.Skip("install.ps1 is Windows' installer, for Windows PowerShell 5.1; install.sh's tests hold Linux")
	}
	ps := filepath.Join(os.Getenv("SystemRoot"), "System32", "WindowsPowerShell", "v1.0", "powershell.exe")
	if _, err := os.Stat(ps); err != nil {
		t.Skip("no Windows PowerShell at " + ps)
	}
	// Windows' own folders only on the PATH: no claude.exe or other program of the person's.
	root := os.Getenv("SystemRoot")
	t.Setenv("PATH", strings.Join([]string{filepath.Join(root, "System32"), root, filepath.Dir(ps), filepath.Dir(gitPath(t))}, ";"))
	return ps
}

// gitPath is git's path, found before the PATH is narrowed (testpack.Isolate and Project run git).
func gitPath(t *testing.T) string {
	t.Helper()
	g, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git is not on the PATH")
	}
	return g
}

func (s *scratch) ps1(t *testing.T, ps string, args ...string) (int, string) {
	t.Helper()
	cmd := exec.Command(ps, append([]string{"-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-File", s.script, "-Target", s.target}, args...)...)
	cmd.Dir = s.tmp
	out, err := cmd.CombinedOutput()
	var ee *exec.ExitError
	if err != nil && !errors.As(err, &ee) {
		t.Fatal(err)
	}
	return cmd.ProcessState.ExitCode(), string(out)
}

// psFunctions runs PowerShell code with install.ps1's functions defined (read from the script's syntax tree, so the
// script's own main part never runs) and returns its exit code and output.
func psFunctions(t *testing.T, ps, code string) (int, string) {
	t.Helper()
	dir := t.TempDir()
	abs, err := filepath.Abs("install.ps1")
	if err != nil {
		t.Fatal(err)
	}
	text := "$ErrorActionPreference = 'Stop'\n" +
		"$ast = [Management.Automation.Language.Parser]::ParseFile('" + strings.ReplaceAll(abs, "'", "''") + "', [ref]$null, [ref]$null)\n" +
		"foreach ($f in $ast.FindAll({ param($n) $n -is [Management.Automation.Language.FunctionDefinitionAst] }, $false)) { . ([scriptblock]::Create($f.Extent.Text)) }\n" +
		code + "\n"
	p := filepath.Join(dir, "run.ps1")
	write(t, p, []byte(text), 0o644)
	cmd := exec.Command(ps, "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-File", p)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	var ee *exec.ExitError
	if err != nil && !errors.As(err, &ee) {
		t.Fatal(err)
	}
	return cmd.ProcessState.ExitCode(), strings.ReplaceAll(string(out), "\r\n", "\n")
}

func psQuote(s string) string { return "'" + strings.ReplaceAll(s, "'", "''") + "'" }

func TestPs1InstallAgainRunningRemove(t *testing.T) {
	src := readFile(t, bonsai(t))
	ps := windowsOnly(t)
	s := newScratch(t, "install.ps1", src)

	code, out := s.ps1(t, ps)
	if code != 0 {
		t.Fatalf("install: exit %d\n%s", code, out)
	}
	if !bytes.Equal(readFile(t, s.target), src) {
		t.Fatal("the installed file is not the source")
	}
	if !strings.Contains(out, sum(src)) || !strings.Contains(out, "bonsai dev") {
		t.Errorf("the install's output does not name the fingerprint and the version:\n%s", out)
	}
	s.record(t)
	rec := readFile(t, filepath.Join(s.home, engine.InstallFile))
	s.checkReads(t)
	if l := leftovers(t, s.bin, "bonsai.exe.*"); len(l) > 0 {
		t.Fatalf("copies left after an install with nothing running: %v", l)
	}

	if code, out := s.ps1(t, ps); code != 0 || !bytes.Equal(rec, readFile(t, filepath.Join(s.home, engine.InstallFile))) {
		t.Fatalf("a second install: exit %d, or its record changed\n%s", code, out)
	}

	alive, stop := s.running(t)
	if !alive() {
		t.Fatal("the installed bonsai's hook guard did not stay running on an open input")
	}
	if code, out := s.ps1(t, ps); code != 0 {
		t.Fatalf("an install while a copy runs: exit %d\n%s", code, out)
	}
	if !alive() {
		t.Fatal("the running bonsai ended during the install")
	}
	if l := leftovers(t, s.bin, "bonsai.exe.old-*"); len(l) != 1 {
		t.Fatalf("the running copy set aside: %v; want one bonsai.exe.old-<n>", l)
	}
	if err := stop(); err != nil {
		t.Fatal(err)
	}
	if code, out := s.ps1(t, ps); code != 0 {
		t.Fatalf("the install after: exit %d\n%s", code, out)
	}
	if l := leftovers(t, s.bin, "bonsai.exe.*"); len(l) > 0 {
		t.Fatalf("the next install left the copy set aside: %v", l)
	}
	s.record(t)

	if code, out := s.ps1(t, ps, "-Remove"); code != 0 {
		t.Fatalf("remove: exit %d\n%s", code, out)
	}
	noFile(t, s.target, filepath.Join(s.home, engine.InstallFile))
	if l := leftovers(t, s.bin, "bonsai.exe*"); len(l) > 0 {
		t.Fatalf("remove left: %v", l)
	}
}

func TestPs1Refusals(t *testing.T) {
	src := readFile(t, bonsai(t))
	ps := windowsOnly(t)
	t.Run("a bonsai.exe that does not run", func(t *testing.T) {
		s := newScratch(t, "install.ps1", []byte("not a program\n"))
		code, out := s.ps1(t, ps)
		if code != 1 || !strings.Contains(out, "does not run") || !strings.Contains(out, "next: ") {
			t.Fatalf("exit %d:\n%s", code, out)
		}
		noFile(t, s.target, filepath.Join(s.home, engine.InstallFile))
	})
	t.Run("-ForceChild in a PowerShell that is not elevated", func(t *testing.T) {
		s := newScratch(t, "install.ps1", src)
		code, out := s.ps1(t, ps, "-ForceChild")
		if strings.Contains(out, "Installed ") {
			t.Skip("this test runs elevated, where -ForceChild is allowed: CI's install job runs it")
		}
		if code != 1 || !strings.Contains(out, "-ForceChild is for tests") {
			t.Fatalf("exit %d:\n%s", code, out)
		}
		noFile(t, s.target, filepath.Join(s.home, engine.InstallFile))
	})
	t.Run("a target that is a folder", func(t *testing.T) {
		s := newScratch(t, "install.ps1", src)
		if err := os.Mkdir(s.target, 0o700); err != nil {
			t.Fatal(err)
		}
		for _, args := range [][]string{nil, {"-Remove"}} {
			if code, out := s.ps1(t, ps, args...); code != 1 || !strings.Contains(out, "is a folder") || !strings.Contains(out, "next: ") {
				t.Fatalf("%v: exit %d:\n%s", args, code, out)
			}
			empty(t, s.target)
		}
		noFile(t, filepath.Join(s.home, engine.InstallFile))
	})
	t.Run("a relative target", func(t *testing.T) {
		s := newScratch(t, "install.ps1", src)
		s.target = `bin\bonsai.exe`
		if code, out := s.ps1(t, ps); code != 1 || !strings.Contains(out, "not an absolute path") {
			t.Fatalf("exit %d:\n%s", code, out)
		}
	})
	t.Run("a declined Windows prompt", func(t *testing.T) {
		code, out := psFunctions(t, ps, "function Start-Process { throw 'The operation was canceled by the user.' }\n"+
			"$c = Invoke-Child 'exit 0' $true\n'not refused'")
		if code != 1 || !strings.Contains(out, "the Windows prompt was declined, so nothing was installed") ||
			!strings.Contains(out, "next: run the last line again and choose Yes") || strings.Contains(out, "not refused") {
			t.Fatalf("exit %d:\n%s", code, out)
		}
	})
	t.Run("a source changed between the hash and the copy", func(t *testing.T) {
		s := newScratch(t, "install.ps1", src)
		wrong := strings.Repeat("0", 64)
		code, out := psFunctions(t, ps, "$t = New-ChildText -Source "+psQuote(s.src)+" -Target "+psQuote(s.target)+" -Hash '"+wrong+
			"' -Real $false -Remove $false\nexit (Invoke-Child $t $false)")
		if code != 12 {
			t.Fatalf("the child with a wrong expected hash: exit %d; want 12 (the copy's check)\n%s", code, out)
		}
		if l := leftovers(t, s.bin, "bonsai.exe*"); len(l) > 0 {
			t.Fatalf("the child left: %v", l)
		}
	})
}

// install.ps1 refuses a target that is a folder in its main part, before the child's text is built for an install or
// a remove (read here on every system; TestPs1Refusals runs it on Windows).
func TestPs1FolderTargetRefusedFirst(t *testing.T) {
	ps := read(t, "install.ps1")
	check := strings.Index(ps, "[IO.Directory]::Exists($Target)")
	main := strings.Index(ps, "\n# Main.")
	if check < 0 || main < 0 || check < main {
		t.Fatal("install.ps1's main part does not refuse a target that is a folder ([IO.Directory]::Exists($Target))")
	}
	for _, call := range []string{"New-ChildText -Source $Target", "New-ChildText -Source $src"} {
		if i := strings.Index(ps, call); i < 0 || i < check {
			t.Errorf("install.ps1 builds the child's text (%s) before it refuses a folder target", call)
		}
	}
}

// childBody is the elevated child's fixed text in install.ps1 (New-ChildText's here-string), read on every system.
func childBody(t *testing.T) string {
	t.Helper()
	ps := read(t, "install.ps1")
	start := strings.Index(ps, "$body = @'\n")
	end := strings.Index(ps, "\n'@")
	if start < 0 || end < start {
		t.Fatal("install.ps1 holds no child's text ($body = @' ... '@)")
	}
	return ps[start+len("$body = @'\n") : end]
}

// The child checks the folder's rights before it copies anything in, and the copy's before it renames it into place,
// so a folder others can write never gets the file or the PATH entry; every right that writes counts.
func TestPs1ChildRightsBeforeCopy(t *testing.T) {
	b := childBody(t)
	folder, copyIn := strings.Index(b, "Test-Rights $dir"), strings.Index(b, "[IO.File]::Copy($src")
	file, swap := strings.Index(b, "Test-Rights $new"), strings.Index(b, "Move-Busy $new $dst")
	if folder < 0 || copyIn < 0 || folder > copyIn {
		t.Error("the child does not check the folder's rights (Test-Rights $dir) before its copy")
	}
	if file < 0 || swap < 0 || file < copyIn || file > swap {
		t.Error("the child does not check the copy's rights (Test-Rights $new) between the copy and the rename")
	}
	line := ""
	for _, l := range strings.Split(b, "\n") {
		if strings.Contains(l, "$mask = ") {
			line = l
		}
	}
	for _, bit := range []string{"0x2", "0x4", "0x10", "0x40", "0x100", "0x10000", "0x40000", "0x80000", "0x10000000", "0x40000000"} {
		found := false
		for _, f := range strings.Fields(line) {
			found = found || f == bit
		}
		if !found {
			t.Errorf("the child's write mask (%q) leaves out %s", strings.TrimSpace(line), bit)
		}
	}
}

// The elevated child compiles and loads no code: Windows PowerShell 5.1's Add-Type writes its source and its DLL in
// the person's %TEMP%, a folder their unelevated programs can write, and then loads the DLL. Neither script uses it,
// so the elevated child reads no user-writable file but the one it copies (plan 5.6.6, note 3).
func TestPs1CompilesNothing(t *testing.T) {
	for _, word := range []string{"Add-Type", "CompileAssembly", "csc.exe", "[Reflection.Assembly]::Load"} {
		if strings.Contains(strings.ToLower(read(t, "install.ps1")), strings.ToLower(word)) {
			t.Errorf("install.ps1 holds %s", word)
		}
	}
}

// The child runs no cmdlet: one from a module (Start-Sleep, Set-Location, New-Object) is loaded from the first folder
// on PSModulePath that holds a module of its name, and the person's Documents\WindowsPowerShell\Modules comes first.
// Only the functions its own text defines, and .NET calls, run in it.
func TestPs1ChildRunsNoCmdlet(t *testing.T) {
	text := childBody(t)
	ps := read(t, "install.ps1")
	if i := strings.Index(ps, "\nfunction Merge-MachinePath {"); i >= 0 {
		end := strings.Index(ps[i+1:], "\n}\n")
		text += ps[i : i+1+end] // the child holds Merge-MachinePath's body too
	}
	own := map[string]bool{}
	for _, m := range regexp.MustCompile(`(?m)^\s*function ([A-Za-z]+-[A-Za-z]+)`).FindAllStringSubmatch(text, -1) {
		own[strings.ToLower(m[1])] = true
	}
	for _, m := range regexp.MustCompile(`\b([A-Z][a-z]+-[A-Z][A-Za-z]+)\b`).FindAllStringSubmatch(text, -1) {
		if !own[strings.ToLower(m[1])] {
			t.Errorf("the child's text runs %s, a cmdlet, not a function it defines", m[1])
		}
	}
}

// The child's text holds each value as a single-quoted literal, a path with a ' and a space among them, and refuses
// a bad hash or a path with no drive before building anything.
func TestPs1ChildText(t *testing.T) {
	ps := windowsOnly(t)
	h := strings.Repeat("ab", 32)
	code, out := psFunctions(t, ps, "$t = New-ChildText -Source 'C:\\a b''c\\bonsai.exe' -Target 'C:\\Program Files\\Bonsai\\bonsai.exe' -Hash '"+h+
		"' -Real $true -Remove $false\n$e = $null\n[void][Management.Automation.Language.Parser]::ParseInput($t, [ref]$null, [ref]$e)\n"+
		"if ($e.Count -ne 0) { 'parse errors: ' + ($e -join '; '); exit 3 }\n$t.Split(\"`n\")[0..4] -join \"`n\"")
	want := "$src = 'C:\\a b''c\\bonsai.exe'\n$dst = 'C:\\Program Files\\Bonsai\\bonsai.exe'\n$want = '" + h + "'\n$real = $true\n$remove = $false\n"
	if code != 0 || out != want {
		t.Fatalf("exit %d; the child's first lines:\n%s\nwant:\n%s", code, out, want)
	}
	for _, bad := range []string{
		"-Source 'C:\\a\\bonsai.exe' -Target 'C:\\b\\bonsai.exe' -Hash 'ABCD'",
		"-Source 'C:\\a\\bonsai.exe' -Target 'C:\\b\\bonsai.exe' -Hash '" + strings.ToUpper(h) + "'",
		"-Source 'bonsai.exe' -Target 'C:\\b\\bonsai.exe' -Hash '" + h + "'",
		"-Source 'C:\\a\\bonsai.exe' -Target '\\\\server\\share\\bonsai.exe' -Hash '" + h + "'",
	} {
		code, out := psFunctions(t, ps, "try { $t = New-ChildText "+bad+" -Real $false -Remove $false; 'built' } catch { 'refused: ' + $_.Exception.Message }")
		if code != 0 || !strings.HasPrefix(out, "refused: ") {
			t.Errorf("New-ChildText %s: exit %d, %s; want refused", bad, code, out)
		}
	}
}

// The machine PATH's merge, on strings only (no test writes the registry): %SystemRoot% entries kept as written, the
// folder once, empty entries dropped only on a change, removal taking every copy out.
func TestPs1MergeMachinePath(t *testing.T) {
	ps := windowsOnly(t)
	f := `C:\Program Files\Bonsai`
	for _, c := range []struct {
		current, want string
		remove        bool
	}{
		{`%SystemRoot%\system32;%SystemRoot%;C:\Tools`, `%SystemRoot%\system32;%SystemRoot%;C:\Tools;` + f, false},
		{`%SystemRoot%\system32;c:\program files\bonsai\;C:\Tools`, `%SystemRoot%\system32;c:\program files\bonsai\;C:\Tools`, false},
		{`%SystemRoot%;;C:\Tools;`, `%SystemRoot%;C:\Tools;` + f, false},
		{`%SystemRoot%;;C:\Tools;`, `%SystemRoot%;;C:\Tools;`, true},
		{f + `;%SystemRoot%;` + f + `\`, f + `;%SystemRoot%`, false},
		{f + `;%SystemRoot%;` + f + `\`, `%SystemRoot%`, true},
		{``, f, false},
	} {
		code, out := psFunctions(t, ps, "[Console]::Out.Write((Merge-MachinePath -Current "+psQuote(c.current)+" -Folder "+psQuote(f)+
			" -Remove $"+map[bool]string{true: "true", false: "false"}[c.remove]+"))")
		if code != 0 || out != c.want {
			t.Errorf("Merge-MachinePath %q (remove %v): exit %d, %q; want %q", c.current, c.remove, code, out, c.want)
		}
	}
}
