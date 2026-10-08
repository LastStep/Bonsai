// Package testpack builds, for Go tests, a local pack repository with four commits, A to D, shaped like Bonsai's
// public test pack (github.com/LastStep/bonsai-test-pack, plan part 4a), so the engine's tests fetch a pack with no
// network. Only tests import it.
//
// What each commit changes, as in the public pack (its README's table):
//
//	A  pack.yaml lists guide.md (kind pack), start.md (kind once), the block and one hook line "echo demo hook A"
//	B  guide.md becomes edition 2; extra.md is new (kind pack); the marker says commit B
//	C  guide.md becomes edition 3
//	D  only the hook line's command: "echo demo hook D"
//
// Every repository and project it makes lives in the test's temporary folder, with git isolated from the
// machine's own configuration (no global or system config, no repository above the folder).
package testpack

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// ID is the pack's id.
const ID = "demo-pack"

// Pack is a built pack repository: its source (a bare repository's folder) and its four commits.
type Pack struct {
	Source     string
	A, B, C, D string
}

// Isolate points git at an empty configuration and no ignore file of the person's, stops it looking above tmp, and
// sets BONSAI_HOME to tmp/home. Every test that runs git or Bonsai calls it first.
func Isolate(t *testing.T, tmp string) (home string) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not on the PATH: the engine fetches packs and finds projects with git")
	}
	empty := filepath.Join(tmp, "gitconfig")
	if err := os.WriteFile(empty, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GIT_CONFIG_GLOBAL", empty)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	// git also reads the person's own ignore file, $XDG_CONFIG_HOME/git/ignore (or ~/.config/git/ignore): none here.
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmp, "xdg"))
	t.Setenv("GIT_CEILING_DIRECTORIES", tmp)
	t.Setenv("CLAUDE_CODE_CHILD_SESSION", "")
	home = filepath.Join(tmp, "home")
	t.Setenv("BONSAI_HOME", home)
	return home
}

// Git runs git in dir, as a fixed test user, and returns its output.
func Git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir, "-c", "user.name=bonsai-test", "-c", "user.email=bonsai-test",
		"-c", "commit.gpgsign=false", "-c", "init.defaultBranch=main", "-c", "core.autocrlf=false"}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

// PackYAML is the pack's bonsai/pack.yaml: hook is the hook line's letter, extra whether extra.md is listed.
func PackYAML(hook string, extra bool) string {
	s := `# bonsai/pack.yaml: a test pack for Bonsai's engine tests.
format: bonsai.pack/1                # the format
id: demo-pack                        # the pack's id
version: "0.1.0"                     # the same at every commit
needs:                               # what a machine needs
  claude_code: "2.1.0"               # the oldest Claude Code
block: block.md                      # the block
files:                               # the files the engine writes
  - path: demo/guide.md              # changed by B and C
    from: guide.md                   # its source
    kind: pack                       # Bonsai's file
  - path: demo/start.md              # written once
    from: start.md                   # its source
    kind: once                       # the project's after init
`
	if extra {
		s += `  - path: demo/extra.md              # new at B
    from: extra.md                   # its source
    kind: pack                       # Bonsai's file
`
	}
	s += `hooks:                               # hook lines
  - event: SessionStart              # a session starts
    matcher: startup                 # new sessions only
    command: "echo demo hook ` + hook + `"     # D changes it
    why: "Prints which demo hook line is in place; it blocks nothing."
deny:                                # deny rules
  - rule: "Edit(demo/never.txt)"     # an example rule
    why: "Agents cannot edit demo/never.txt (the demo pack's rule)."
`
	return s
}

const block = "<!--\nbonsai/block.md: the demo pack's block. This comment stays in the pack.\n-->\nThe demo pack is linked: its role is demo-pack:marker.\n"

// Build makes the pack repository with commits A to D and returns it.
func Build(t *testing.T, tmp string) *Pack {
	t.Helper()
	work := filepath.Join(tmp, "pack-work")
	if err := os.MkdirAll(work, 0o755); err != nil {
		t.Fatal(err)
	}
	Git(t, work, "init", "-q")
	write := func(rel, content string) {
		p := filepath.Join(work, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	commit := func(msg string) string {
		Git(t, work, "add", "-A")
		Git(t, work, "commit", "-q", "-m", msg)
		return Git(t, work, "rev-parse", "HEAD")
	}
	p := &Pack{}
	write(".claude-plugin/plugin.json", "{\"name\": \"demo-pack\", \"description\": \"A test pack.\"}\n")
	write("agents/marker.md", "---\nname: marker\ndescription: says commit A\n---\nSay commit A.\n")
	write("bonsai/pack.yaml", PackYAML("A", false))
	write("bonsai/block.md", block)
	write("bonsai/files/guide.md", "# Guide\n\nEdition 1.\n")
	write("bonsai/files/start.md", "# Start here\n\nThe project's own file once written.\n")
	p.A = commit("A")
	write("agents/marker.md", "---\nname: marker\ndescription: says commit B\n---\nSay commit B.\n")
	write("bonsai/files/guide.md", "# Guide\n\nEdition 2.\n")
	write("bonsai/files/extra.md", "# Extra\n\nNew at commit B.\n")
	write("bonsai/pack.yaml", PackYAML("A", true))
	p.B = commit("B")
	write("bonsai/files/guide.md", "# Guide\n\nEdition 3.\n")
	p.C = commit("C")
	write("bonsai/pack.yaml", PackYAML("D", true))
	p.D = commit("D")
	p.Source = filepath.Join(tmp, "demo-pack.git")
	Git(t, tmp, "clone", "-q", "--bare", "--", work, p.Source)
	return p
}

// Project makes an empty git checkout at tmp/<name> and returns it.
func Project(t *testing.T, tmp, name string) string {
	t.Helper()
	root := filepath.Join(tmp, name)
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	Git(t, root, "init", "-q")
	r, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

// SetRef changes the ref of bonsai.yaml's one pack, as a person would.
func SetRef(t *testing.T, root, from, to string) {
	t.Helper()
	p := filepath.Join(root, "bonsai.yaml")
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), from) {
		t.Fatalf("bonsai.yaml does not hold %s", from)
	}
	if err := os.WriteFile(p, []byte(strings.Replace(string(b), from, to, 1)), 0o644); err != nil {
		t.Fatal(err)
	}
}
