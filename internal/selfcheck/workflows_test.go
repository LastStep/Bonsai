package selfcheck

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// workflowFolders are the folders whose workflow files this test holds, relative to the module root.
var workflowFolders = []string{".github/workflows", "packs/template/.github/workflows"}

// unpinned lists the workflow files the pin rule does not yet hold, each with its reason. A file here is read by no
// check below; take it out when its reason is gone.
var unpinned = map[string]string{}

var (
	usesLine = regexp.MustCompile(`^(\s*)(?:-\s+)?uses:\s*(\S+)(.*)$`)
	// pinned is an action at a full commit: owner/repo or owner/repo/path, then @ and 40 lowercase hex characters.
	pinned = regexp.MustCompile(`^[A-Za-z0-9._-]+/[A-Za-z0-9._/-]+@[0-9a-f]{40}$`)
	// version is the comment after a pin: a # and the release it is, vX.Y.Z.
	version  = regexp.MustCompile(`^\s+# v\d+\.\d+\.\d+\s*$`)
	stepLine = regexp.MustCompile(`^\s*-\s`)
)

// check returns one problem per line that breaks the rules, each "file:line: what". Every uses: is a local action
// (./...) or owner/repo@<40 hex> followed by "# vX.Y.Z"; every actions/checkout step has persist-credentials: false.
func check(name, content string) []string {
	var out []string
	lines := strings.Split(content, "\n")
	for i, l := range lines {
		m := usesLine.FindStringSubmatch(l)
		if m == nil {
			continue
		}
		ref, rest := m[2], m[3]
		at := fmt.Sprintf("%s:%d", name, i+1)
		if strings.HasPrefix(ref, "./") {
			continue
		}
		if !pinned.MatchString(ref) {
			out = append(out, at+": "+ref+" is not pinned by a full commit (owner/repo@<40 hex>)")
			continue
		}
		if !version.MatchString(rest) {
			out = append(out, at+": "+ref+" has no \"# vX.Y.Z\" comment naming its release")
		}
		if strings.HasPrefix(ref, "actions/checkout@") && !stepHasPersistFalse(lines, i) {
			out = append(out, at+": "+ref+" has no persist-credentials: false")
		}
	}
	return out
}

// stepHasPersistFalse reads the lines of the step that holds lines[i], from i to the next step's dash.
func stepHasPersistFalse(lines []string, i int) bool {
	for j := i + 1; j < len(lines) && !stepLine.MatchString(lines[j]); j++ {
		if strings.TrimSpace(strings.SplitN(lines[j], "#", 2)[0]) == "persist-credentials: false" {
			return true
		}
	}
	return false
}

// workflowFiles returns the module-relative path (forward slashes) of every .yml and .yaml file in the folders.
func workflowFiles(t *testing.T) []string {
	t.Helper()
	var files []string
	for _, dir := range workflowFolders {
		entries, err := os.ReadDir(filepath.Join("..", "..", filepath.FromSlash(dir)))
		if err != nil {
			t.Fatalf("read %s: %v", dir, err)
		}
		for _, e := range entries {
			if ext := filepath.Ext(e.Name()); !e.IsDir() && (ext == ".yml" || ext == ".yaml") {
				files = append(files, dir+"/"+e.Name())
			}
		}
	}
	return files
}

func read(t *testing.T, rel string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", filepath.FromSlash(rel)))
	if err != nil {
		t.Fatal(err)
	}
	return strings.ReplaceAll(string(b), "\r\n", "\n")
}

// TestWorkflowActionsArePinned holds every workflow file in the folders, but the listed exceptions, to the rules.
func TestWorkflowActionsArePinned(t *testing.T) {
	files := workflowFiles(t)
	seen := 0
	for _, f := range files {
		if _, ok := unpinned[f]; ok {
			continue
		}
		seen++
		for _, p := range check(f, read(t, f)) {
			t.Error(p)
		}
	}
	if seen < 3 {
		t.Errorf("only %d workflow files were checked; ci.yml, codeql.yml and the pack template's pack.yml are expected", seen)
	}
	for f, why := range unpinned {
		if why == "" {
			t.Errorf("%s is excepted with no reason", f)
		}
		found := false
		for _, g := range files {
			found = found || g == f
		}
		if !found {
			t.Errorf("%s is excepted but is not a workflow file here; take it out of the list", f)
		}
	}
}

// TestPinCheckFailsOnBadCopies shows the check failing on changed copies of the real ci.yml: a floating tag, a pin
// with no version comment, a pin with a comment that is not a release, and a checkout that keeps its credentials.
func TestPinCheckFailsOnBadCopies(t *testing.T) {
	good := read(t, ".github/workflows/ci.yml")
	if p := check("ci.yml", good); len(p) != 0 {
		t.Fatalf("the real ci.yml has problems: %v", p)
	}
	const pin = "actions/checkout@d23441a48e516b6c34aea4fa41551a30e30af803  # v6.1.0"
	if !strings.Contains(good, pin) {
		t.Fatalf("ci.yml no longer holds %q; update this test with the pin", pin)
	}
	cases := []struct{ name, from, to, want string }{
		{"floating tag", pin, "actions/checkout@v6", "is not pinned"},
		{"no comment", pin, "actions/checkout@d23441a48e516b6c34aea4fa41551a30e30af803", "no \"# vX.Y.Z\" comment"},
		{"comment not a release", pin, "actions/checkout@d23441a48e516b6c34aea4fa41551a30e30af803  # latest", "no \"# vX.Y.Z\" comment"},
		{"short hash", pin, "actions/checkout@d23441a  # v6.1.0", "is not pinned"},
		{"credentials kept", "persist-credentials: false", "persist-credentials: true", "persist-credentials: false"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			bad := strings.Replace(good, c.from, c.to, 1)
			got := strings.Join(check("ci.yml", bad), "\n")
			if !strings.Contains(got, c.want) || !regexp.MustCompile(`ci\.yml:\d+:`).MatchString(got) {
				t.Errorf("want a problem naming file and line with %q, got %q", c.want, got)
			}
		})
	}
	if p := check("x.yml", "    steps:\n      - uses: ./.github/actions/mine\n"); len(p) != 0 {
		t.Errorf("a local action must pass: %v", p)
	}
}

// releaseLookalikes returns every name in a workflows folder, other than release.yml itself, that starts with
// "release.yml". gh attestation verify's --signer-workflow matches the signing workflow's path by its start, so a
// workflow named, say, release.yml.extra.yml could make records the install lines' check would take for the release
// workflow's.
func releaseLookalikes(names []string) []string {
	var out []string
	for _, n := range names {
		if n != "release.yml" && strings.HasPrefix(n, "release.yml") {
			out = append(out, n)
		}
	}
	return out
}

// workflowNames lists every entry of a folder by name, files of any kind and folders alike.
func workflowNames(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read %s: %v", dir, err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names
}

// TestNoReleaseWorkflowLookalike holds .github/workflows/ to one name starting with release.yml: release.yml.
func TestNoReleaseWorkflowLookalike(t *testing.T) {
	names := workflowNames(t, filepath.Join("..", "..", ".github", "workflows"))
	found := false
	for _, n := range names {
		found = found || n == "release.yml"
	}
	if !found {
		t.Fatal(".github/workflows/release.yml is missing; the install lines name it as the signer")
	}
	for _, n := range releaseLookalikes(names) {
		t.Errorf(".github/workflows/%s starts with release.yml: gh attestation verify --signer-workflow "+
			"LastStep/Bonsai/.github/workflows/release.yml would match records it signs; rename it", n)
	}
}

// TestReleaseLookalikeCheckFailsOnACopy shows the check failing on a scratch copy of the folder with a lookalike
// added, and passing on names that only resemble it.
func TestReleaseLookalikeCheckFailsOnACopy(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join("..", "..", ".github", "workflows")
	for _, n := range workflowNames(t, src) {
		b, err := os.ReadFile(filepath.Join(src, n))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, n), b, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if got := releaseLookalikes(workflowNames(t, dir)); len(got) != 0 {
		t.Fatalf("the copy of the real folder has lookalikes: %v", got)
	}
	for _, n := range []string{"release.yml.yml", "release.yml-old.yaml", "release.ymlx"} {
		if err := os.WriteFile(filepath.Join(dir, n), []byte("name: x\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, n := range []string{"release-check.yml", "releases.yml", "a-release.yml"} {
		if err := os.WriteFile(filepath.Join(dir, n), []byte("name: x\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	got := strings.Join(releaseLookalikes(workflowNames(t, dir)), " ")
	if want := "release.yml-old.yaml release.yml.yml release.ymlx"; got != want {
		t.Errorf("lookalikes found: %q, want %q", got, want)
	}
}
