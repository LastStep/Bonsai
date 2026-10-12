package vergit

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/schema"
)

// Every test builds its own repositories with the machine's git (Windows git on Windows) in temporary folders, with
// a temporary home and global configuration, so neither the machine's settings nor its global excludes reach them.

// isolate gives the test a home, an empty global git configuration and no git variable from the caller, and returns
// the home.
func isolate(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not on the PATH: the ver-git rung reads repositories with it")
	}
	home := t.TempDir()
	global := filepath.Join(home, "gitconfig")
	if err := os.WriteFile(global, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("GIT_CONFIG_GLOBAL", global)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	for _, k := range []string{"GIT_DIR", "GIT_WORK_TREE", "GIT_INDEX_FILE", "GIT_COMMON_DIR", "GIT_OBJECT_DIRECTORY",
		"GIT_CONFIG_COUNT", "GIT_CONFIG_PARAMETERS", "GIT_CEILING_DIRECTORIES"} {
		t.Setenv(k, "")
		must(t, os.Unsetenv(k))
	}
	return home
}

// git runs git in dir for a fixture and returns its standard output, trimmed.
func git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-c", "user.name=bonsai-test", "-c", "user.email=bonsai-test",
		"-c", "commit.gpgsign=false", "-c", "init.defaultBranch=main", "-c", "core.autocrlf=false",
		"-c", "advice.detachedHead=false"}, args...)...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, stderr.String())
	}
	return strings.TrimSpace(stdout.String())
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func write(t *testing.T, root, rel, text string) {
	t.Helper()
	p := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

// repo makes a repository: main's one commit (the base, with a tracked .gitignore ignoring build/, a file, a Go
// package and its test), then a branch task one commit ahead. It returns the top folder and the base.
func repo(t *testing.T) (root, base string) {
	t.Helper()
	root = filepath.Join(t.TempDir(), "r")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	git(t, root, "init", "-q")
	write(t, root, ".gitignore", "build/\n")
	write(t, root, "a.txt", "a\n")
	write(t, root, "pkg/a.go", "package pkg\n")
	write(t, root, "pkg/a_test.go", "package pkg\n")
	git(t, root, "add", ".")
	git(t, root, "commit", "-q", "-m", "base")
	base = git(t, root, "rev-parse", "HEAD")
	git(t, root, "checkout", "-q", "-b", "task")
	write(t, root, "b.txt", "b\n")
	git(t, root, "add", ".")
	git(t, root, "commit", "-q", "-m", "work")
	return root, base
}

func input(root, base string) Input {
	return Input{Root: root, BaseBranch: "main", MergeBase: base, Tests: []string{format.TestsGoJSON}}
}

// judge judges and checks what every outcome must hold: the captures are GitFindings' words in order with values a
// result holds (a number, a text or null), a ladder result holding them encodes, and the reason is one line of ASCII.
func judge(t *testing.T, in Input) Outcome {
	t.Helper()
	o := Judge(in)
	if len(o.Captures) != len(format.GitFindings) {
		t.Fatalf("%d captures for %d words: %v", len(o.Captures), len(format.GitFindings), o.Captures)
	}
	for i, w := range format.GitFindings {
		m := o.Captures[i]
		if m.Key != w.Word {
			t.Errorf("capture %d is %s, want %s", i, m.Key, w.Word)
		}
		switch m.Value.(type) {
		case nil, string, int64:
		default:
			t.Errorf("capture %s holds a %T", m.Key, m.Value)
		}
	}
	for i := 0; i < len(o.Reason); i++ {
		if o.Reason[i] < 0x20 || o.Reason[i] > 0x7e {
			t.Errorf("the reason is not one line of ASCII: %q", o.Reason)
			break
		}
	}
	if o.Status != "green" && o.Status != "error" {
		t.Errorf("status %q", o.Status)
	}
	return o
}

// count is a finding's count; ok is false when it was not judged.
func count(t *testing.T, o Outcome, word string) (int, bool) {
	t.Helper()
	v, held := o.Captures.Get(word)
	if !held {
		t.Fatalf("no capture %s", word)
	}
	if v == nil {
		return 0, false
	}
	n, isNum := v.(int64)
	if !isNum {
		t.Fatalf("capture %s is %v", word, v)
	}
	return int(n), true
}

// only checks the outcome is green and finds exactly these counts among the findings (every other judged one 0),
// with Hides as given.
func only(t *testing.T, o Outcome, hides bool, want map[string]int) {
	t.Helper()
	if o.Status != "green" {
		t.Fatalf("status %s: %s", o.Status, o.Reason)
	}
	if o.Hides != hides {
		t.Errorf("Hides %v, want %v: %s", o.Hides, hides, o.Reason)
	}
	for _, w := range format.GitFindings[3:] {
		n, judged := count(t, o, w.Word)
		if !judged {
			if _, wanted := want[w.Word]; wanted {
				t.Errorf("%s not judged: %s", w.Word, o.Reason)
			}
			continue
		}
		if n != want[w.Word] {
			t.Errorf("%s = %d, want %d: %s", w.Word, n, want[w.Word], o.Reason)
		}
	}
}

func has(t *testing.T, o Outcome, parts ...string) {
	t.Helper()
	for _, p := range parts {
		if !strings.Contains(o.Reason, p) {
			t.Errorf("the reason lacks %q: %s", p, o.Reason)
		}
	}
}

// gitTemplateExclude is git's own info/exclude, comments only.
func excludeHash(text string) string {
	sum := sha256.Sum256([]byte(strings.ReplaceAll(text, "\r\n", "\n")))
	return "sha256:" + hex.EncodeToString(sum[:])
}

func TestCleanRepositoryFindsNothing(t *testing.T) {
	isolate(t)
	root, base := repo(t)
	write(t, root, "build/out.bin", "x") // ignored by the committed .gitignore: changes nothing
	o := judge(t, input(root, base))
	only(t, o, false, nil)
	has(t, o, "nothing found", "Base "+base[:12]+", the merge base with main", "0 stash entries")
	if v, _ := o.Captures.Get("base"); v != base {
		t.Errorf("base %v, want %s", v, base)
	}
	if n, _ := count(t, o, "stash-count"); n != 0 {
		t.Errorf("stash-count %d", n)
	}
	raw, err := os.ReadFile(filepath.Join(root, ".git", "info", "exclude"))
	if err != nil {
		raw = nil // a git that writes no template: an absent file reads as empty
	}
	if v, _ := o.Captures.Get("exclude-hash"); v != excludeHash(string(raw)) {
		t.Errorf("exclude-hash %v", v)
	}
	// The first climb compares nothing.
	for _, w := range []string{"exclude-edited", "stash-added", "base-rewritten"} {
		if _, judged := count(t, o, w); judged {
			t.Errorf("%s judged with no earlier climb", w)
		}
	}
	has(t, o, "Not judged: exclude-edited (no earlier climb")
	// Byte-stable: the same checkout judged again gives the same bytes.
	a, err := schema.Encode(o.Captures)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := schema.Encode(judge(t, input(root, base)).Captures)
	if !bytes.Equal(a, b) {
		t.Errorf("two judgments differ:\n%s\n%s", a, b)
	}
}

// The ordinary shapes a repository takes are green, find nothing and hide nothing: never an error.
func TestOrdinaryShapesAreGreen(t *testing.T) {
	isolate(t)
	t.Run("one commit", func(t *testing.T) {
		root := filepath.Join(t.TempDir(), "one")
		must(t, os.MkdirAll(root, 0o755))
		git(t, root, "init", "-q")
		write(t, root, "a.txt", "a\n")
		git(t, root, "add", ".")
		git(t, root, "commit", "-q", "-m", "one")
		head := git(t, root, "rev-parse", "HEAD")
		o := judge(t, input(root, head))
		only(t, o, false, nil)
		has(t, o, "HEAD itself, on main")
	})
	t.Run("no commit yet", func(t *testing.T) {
		root := filepath.Join(t.TempDir(), "new")
		must(t, os.MkdirAll(root, 0o755))
		git(t, root, "init", "-q")
		write(t, root, "a_test.go", "package a\n")
		o := judge(t, Input{Root: root, Tests: []string{format.TestsGoJSON}})
		only(t, o, false, nil)
		has(t, o, "HEAD has no commit yet", "test-file-changed (no base)")
		if v, _ := o.Captures.Get("base"); v != nil {
			t.Errorf("base %v", v)
		}
	})
	t.Run("detached HEAD", func(t *testing.T) {
		root, base := repo(t)
		git(t, root, "checkout", "-q", "--detach")
		only(t, judge(t, input(root, base)), false, nil)
	})
	t.Run("no base branch", func(t *testing.T) {
		root, _ := repo(t)
		o := judge(t, Input{Root: root, Tests: []string{format.TestsGoJSON}})
		only(t, o, false, nil)
		has(t, o, "Base none: the runner found no base branch")
	})
	t.Run("linked worktree", func(t *testing.T) {
		root, base := repo(t)
		wt := filepath.Join(filepath.Dir(root), "wt")
		git(t, root, "worktree", "add", "-q", "-b", "other", wt)
		only(t, judge(t, input(wt, base)), false, nil)
	})
	t.Run("shallow clone", func(t *testing.T) {
		root, _ := repo(t)
		write(t, root, "c.txt", "c\n")
		git(t, root, "add", ".")
		git(t, root, "commit", "-q", "-m", "more")
		u := filepath.ToSlash(root)
		if !strings.HasPrefix(u, "/") {
			u = "/" + u
		}
		clone := filepath.Join(filepath.Dir(root), "shallow")
		git(t, filepath.Dir(root), "clone", "-q", "--depth", "1", "--branch", "task", "file://"+u, clone)
		if s := git(t, clone, "rev-parse", "--is-shallow-repository"); s != "true" {
			t.Fatalf("the clone is not shallow: %s", s)
		}
		head := git(t, clone, "rev-parse", "HEAD")
		// The base an earlier climb recorded is not in the clone: not judged, never an error.
		earlier := schema.Object{{Key: "base", Value: strings.Repeat("ab", 20)}}
		in := Input{Root: clone, MergeBase: head, Tests: []string{format.TestsGoJSON}, Earlier: earlier}
		o := judge(t, in)
		only(t, o, false, nil)
		has(t, o, "base-rewritten (the recorded base is not in this shallow clone)",
			"test-file-changed (the base is not in this clone)")
	})
	t.Run("stash present before the first climb", func(t *testing.T) {
		root, base := repo(t)
		write(t, root, "a.txt", "changed\n")
		git(t, root, "stash", "-q")
		o := judge(t, input(root, base))
		only(t, o, false, nil)
		has(t, o, "1 stash entries")
	})
}

func TestTestFileChanged(t *testing.T) {
	isolate(t)
	root, base := repo(t)
	write(t, root, "pkg/a_test.go", "package pkg // weakened\n")
	write(t, root, "pkg/b_test.go", "package pkg\n") // untracked
	write(t, root, "pkg/testdata/in.txt", "x\n")
	git(t, root, "add", "pkg/testdata/in.txt")
	git(t, root, "commit", "-q", "-m", "data")
	write(t, root, "pkg/a.go", "package pkg // code, not a test\n")
	o := judge(t, input(root, base))
	only(t, o, false, map[string]int{"test-file-changed": 3})
	has(t, o, "test-file-changed 3: pkg/a_test.go (changed), pkg/b_test.go (added, untracked), pkg/testdata/in.txt (added)")
	if strings.Contains(o.Reason, "pkg/a.go") {
		t.Errorf("a file that is not a test is listed: %s", o.Reason)
	}
	must(t, os.Remove(filepath.Join(root, "pkg", "a_test.go")))
	has(t, judge(t, input(root, base)), "pkg/a_test.go (deleted)")

	// Not judged: no tests form, a form with no patterns, an unknown form, no base.
	for _, tests := range [][]string{nil, {format.TestsJUnit}, {"pytest"}} {
		in := input(root, base)
		in.Tests = tests
		o := judge(t, in)
		if _, judged := count(t, o, "test-file-changed"); judged {
			t.Errorf("judged with tests %v", tests)
		}
		has(t, o, "test-file-changed (no tests form with a test-file convention")
	}
	o = judge(t, Input{Root: root, Tests: []string{format.TestsGoJSON}})
	if _, judged := count(t, o, "test-file-changed"); judged {
		t.Errorf("judged with no base")
	}
}

// assume-unchanged and skip-worktree make git status show a changed file as clean: the rung names the entry, and
// Hides is true.
func TestIndexEntriesHide(t *testing.T) {
	isolate(t)
	for _, c := range []struct{ flag, word string }{{"--assume-unchanged", "assume-unchanged"}, {"--skip-worktree", "skip-worktree"}} {
		t.Run(c.word, func(t *testing.T) {
			root, base := repo(t)
			git(t, root, "update-index", c.flag, "a.txt")
			write(t, root, "a.txt", "changed behind git's back\n")
			if s := git(t, root, "status", "--porcelain"); s != "" {
				t.Fatalf("git status shows the change, so the fixture proves nothing: %s", s)
			}
			o := judge(t, input(root, base))
			only(t, o, true, map[string]int{c.word: 1})
			has(t, o, c.word+" 1: a.txt (hides a change from git; undo each: git update-index --no-"+c.word,
				"1 kinds found, 1 hiding a change from git status, so the run counts as dirty")
		})
	}
	t.Run("both on one file", func(t *testing.T) {
		root, base := repo(t)
		git(t, root, "update-index", "--assume-unchanged", "a.txt")
		git(t, root, "update-index", "--skip-worktree", "a.txt")
		only(t, judge(t, input(root, base)), true, map[string]int{"assume-unchanged": 1, "skip-worktree": 1})
	})
	t.Run("a test file under the entry", func(t *testing.T) {
		root, base := repo(t)
		git(t, root, "update-index", "--assume-unchanged", "pkg/a_test.go")
		write(t, root, "pkg/a_test.go", "package pkg // weakened\n")
		// git's diff trusts the entry, so the file is listed under the entry (a known limit), and the run is dirty.
		only(t, judge(t, input(root, base)), true, map[string]int{"assume-unchanged": 1})
	})
}

// A file only info/exclude or the global excludes ignore is hidden; one the repository's own .gitignore ignores is
// not, whatever else ignores it too.
func TestExcludeHidden(t *testing.T) {
	home := isolate(t)
	t.Run("info/exclude", func(t *testing.T) {
		root, base := repo(t)
		appendExclude(t, root, "secret.txt\nlocal/\n")
		write(t, root, "secret.txt", "x\n")
		write(t, root, "local/deep/code.go", "package x\n")
		if s := git(t, root, "status", "--porcelain"); s != "" {
			t.Fatalf("git status shows the files: %s", s)
		}
		o := judge(t, input(root, base))
		only(t, o, true, map[string]int{"exclude-hidden": 2, "exclude-lines": 2})
		has(t, o, "exclude-hidden 2: local/ (by info/exclude), secret.txt (by info/exclude) (hides a change from git;",
			"exclude-lines 2: local/, secret.txt")
	})
	t.Run("a file in a folder git shows", func(t *testing.T) {
		root, base := repo(t)
		appendExclude(t, root, "secret.txt\n")
		write(t, root, "pkg/secret.txt", "x\n") // pkg/ is tracked: the file is listed on its own
		only(t, judge(t, input(root, base)), true, map[string]int{"exclude-hidden": 1, "exclude-lines": 1})
	})
	excludes := filepath.Join(home, "my-ignore")
	write(t, home, "my-ignore", "secret.txt\n")
	for _, c := range []struct {
		name string
		set  func(t *testing.T, root string)
	}{
		{"global core.excludesFile", func(t *testing.T, root string) {
			write(t, home, "gitconfig", "[core]\n\texcludesFile = "+filepath.ToSlash(excludes)+"\n")
		}},
		{"git's default global file", func(t *testing.T, root string) {
			write(t, home, ".config/git/ignore", "secret.txt\n")
		}},
		{"the repository's core.excludesFile", func(t *testing.T, root string) {
			git(t, root, "config", "core.excludesFile", filepath.ToSlash(excludes))
		}},
		{"core.excludesFile from the environment", func(t *testing.T, root string) {
			t.Setenv("GIT_CONFIG_COUNT", "1")
			t.Setenv("GIT_CONFIG_KEY_0", "core.excludesFile")
			t.Setenv("GIT_CONFIG_VALUE_0", filepath.ToSlash(excludes))
		}},
	} {
		t.Run(c.name, func(t *testing.T) {
			root, base := repo(t)
			write(t, home, "gitconfig", "")
			must(t, os.RemoveAll(filepath.Join(home, ".config")))
			c.set(t, root)
			write(t, root, "secret.txt", "x\n")
			if s := git(t, root, "status", "--porcelain"); s != "" {
				t.Fatalf("git status shows the file: %s", s)
			}
			o := judge(t, input(root, base))
			only(t, o, true, map[string]int{"exclude-hidden": 1})
			has(t, o, "exclude-hidden 1: secret.txt (by core.excludesFile)")
		})
	}
	t.Run("also in the repository's .gitignore", func(t *testing.T) {
		root, base := repo(t)
		write(t, home, "gitconfig", "[core]\n\texcludesFile = "+filepath.ToSlash(excludes)+"\n")
		appendExclude(t, root, "build/\n")
		write(t, root, "secret.txt", "x\n")
		write(t, root, ".gitignore", "build/\nsecret.txt\n")
		git(t, root, "commit", "-q", "-am", "ignore it")
		write(t, root, "build/out.bin", "x")
		only(t, judge(t, input(root, base)), false, map[string]int{"exclude-lines": 1})
	})
	t.Run("a negation in info/exclude", func(t *testing.T) {
		root, base := repo(t)
		write(t, home, "gitconfig", "[core]\n\texcludesFile = "+filepath.ToSlash(excludes)+"\n")
		appendExclude(t, root, "!secret.txt\n")
		write(t, root, "secret.txt", "x\n") // git status shows it: nothing hidden
		only(t, judge(t, input(root, base)), false, map[string]int{"exclude-lines": 1})
	})
	write(t, home, "gitconfig", "")
}

func appendExclude(t *testing.T, root, lines string) {
	t.Helper()
	p := filepath.Join(root, ".git", "info", "exclude")
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	f, err := os.OpenFile(p, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.WriteString(lines)
	must(t, err)
	must(t, f.Close())
}

// A .gitignore git does not track, which ignores itself, hides itself and what else it ignores from git status.
func TestUntrackedGitignoreHides(t *testing.T) {
	isolate(t)
	t.Run("a new folder", func(t *testing.T) {
		root, base := repo(t)
		write(t, root, "newpkg/.gitignore", "*\n")
		write(t, root, "newpkg/code.go", "package newpkg\n")
		if s := git(t, root, "status", "--porcelain"); s != "" {
			t.Fatalf("git status shows the folder: %s", s)
		}
		o := judge(t, input(root, base))
		only(t, o, true, map[string]int{"untracked-gitignore": 1})
		has(t, o, "untracked-gitignore 1: newpkg/ (by newpkg/.gitignore)")
	})
	t.Run("a folder git tracks", func(t *testing.T) {
		root, base := repo(t)
		write(t, root, "pkg/.gitignore", "helper.go\n.gitignore\n")
		write(t, root, "pkg/helper.go", "package pkg\n")
		if s := git(t, root, "status", "--porcelain"); s != "" {
			t.Fatalf("git status shows the files: %s", s)
		}
		o := judge(t, input(root, base))
		only(t, o, true, map[string]int{"untracked-gitignore": 2})
		has(t, o, "untracked-gitignore 2: pkg/.gitignore (by pkg/.gitignore), pkg/helper.go (by pkg/.gitignore)")
	})
	t.Run("a tracked .gitignore in a folder", func(t *testing.T) {
		root, base := repo(t)
		write(t, root, "pkg/.gitignore", "*.tmp\n")
		git(t, root, "add", "pkg/.gitignore")
		git(t, root, "commit", "-q", "-m", "ignore tmp")
		write(t, root, "pkg/x.tmp", "x\n")
		only(t, judge(t, input(root, base)), false, nil)
	})
}

func TestExcludeLinesAndEdits(t *testing.T) {
	isolate(t)
	root, base := repo(t)
	first := judge(t, input(root, base))
	only(t, first, false, nil)

	// Lines other than comments, matching no file: listed, hiding nothing. Comments and blank lines are not lines.
	appendExclude(t, root, "# a comment\n\n   \n*.swp\n")
	in := input(root, base)
	in.Earlier = first.Captures
	o := judge(t, in)
	only(t, o, false, map[string]int{"exclude-lines": 1, "exclude-edited": 1})
	has(t, o, "exclude-lines 1: *.swp", "exclude-edited 1: info/exclude")

	// The same file again: not edited since.
	in.Earlier = o.Captures
	only(t, judge(t, in), false, map[string]int{"exclude-lines": 1})

	// Line ends read as LF: the same lines with CRLF hash the same.
	p := filepath.Join(root, ".git", "info", "exclude")
	raw, _ := os.ReadFile(p)
	if err := os.WriteFile(p, bytes.ReplaceAll(raw, []byte("\n"), []byte("\r\n")), 0o644); err != nil {
		t.Fatal(err)
	}
	crlf := judge(t, in)
	if a, b := capture(crlf, "exclude-hash"), capture(o, "exclude-hash"); a != b {
		t.Errorf("CRLF hash %v, LF hash %v", a, b)
	}
	only(t, crlf, false, map[string]int{"exclude-lines": 1})

	// An absent file reads as empty, and is no error.
	must(t, os.Remove(p))
	if v := capture(judge(t, input(root, base)), "exclude-hash"); v != excludeHash("") {
		t.Errorf("absent info/exclude hashes as %v", v)
	}
}

func capture(o Outcome, key string) any {
	v, _ := o.Captures.Get(key)
	return v
}

func TestFilterLocal(t *testing.T) {
	isolate(t)
	root, base := repo(t)
	git(t, root, "config", "filter.hide.clean", "cat")
	git(t, root, "config", "filter.hide.smudge", "cat")
	o := judge(t, input(root, base))
	only(t, o, false, map[string]int{"filter-local": 1})
	has(t, o, "filter-local 1: hide")
}

func TestStashAdded(t *testing.T) {
	isolate(t)
	root, base := repo(t)
	first := judge(t, input(root, base))
	write(t, root, "a.txt", "a failing change, set aside\n")
	git(t, root, "stash", "-q")
	in := input(root, base)
	in.Earlier = first.Captures
	o := judge(t, in)
	only(t, o, false, map[string]int{"stash-added": 1})
	has(t, o, "stash-added 1: stash@{0}", "1 stash entries")
	if n, _ := count(t, o, "stash-count"); n != 1 {
		t.Errorf("stash-count %d", n)
	}
	// Counted against the last climb: nothing new since.
	in.Earlier = o.Captures
	only(t, judge(t, in), false, nil)
	// A stash in a linked worktree is the repository's: it counts there too.
	wt := filepath.Join(filepath.Dir(root), "wt")
	git(t, root, "worktree", "add", "-q", "-b", "other", wt)
	write(t, wt, "a.txt", "another\n")
	git(t, wt, "stash", "-q")
	in.Earlier = o.Captures
	only(t, judge(t, in), false, map[string]int{"stash-added": 1})
}

func TestBaseRewritten(t *testing.T) {
	isolate(t)
	t.Run("in history", func(t *testing.T) {
		root, base := repo(t)
		in := input(root, base)
		in.Earlier = judge(t, in).Captures
		write(t, root, "c.txt", "c\n")
		git(t, root, "add", ".")
		git(t, root, "commit", "-q", "-m", "more")
		only(t, judge(t, in), false, nil)
	})
	t.Run("rewritten", func(t *testing.T) {
		root, base := repo(t)
		in := input(root, base)
		in.Earlier = judge(t, in).Captures
		// main's base commit amended, the task branch moved onto it: the recorded base left HEAD's history.
		git(t, root, "checkout", "-q", "main")
		git(t, root, "commit", "-q", "--amend", "-m", "base, rewritten")
		newBase := git(t, root, "rev-parse", "HEAD")
		git(t, root, "checkout", "-q", "task")
		git(t, root, "rebase", "-q", "--onto", "main", base, "task")
		in.MergeBase = newBase
		o := judge(t, in)
		only(t, o, false, map[string]int{"base-rewritten": 1})
		has(t, o, "base-rewritten 1: "+base)
		if v := capture(o, "base"); v != newBase {
			t.Errorf("base %v, want the new merge base %s", v, newBase)
		}
	})
	t.Run("made to look whole by a replace ref or a graft", func(t *testing.T) {
		root, base := repo(t)
		in := input(root, base)
		in.Earlier = judge(t, in).Captures
		git(t, root, "checkout", "-q", "--orphan", "fresh")
		git(t, root, "commit", "-q", "-m", "unrelated")
		orphan := git(t, root, "rev-parse", "HEAD")
		in.MergeBase = ""
		only(t, judge(t, in), false, map[string]int{"base-rewritten": 1})
		// A replace ref showing the orphan as the base's child: plain git now calls the base an ancestor.
		fake := git(t, root, "commit-tree", "-p", base, "-m", "unrelated", orphan+"^{tree}")
		git(t, root, "replace", orphan, fake)
		if err := exec.Command("git", "-C", root, "merge-base", "--is-ancestor", base, "HEAD").Run(); err != nil {
			t.Fatalf("the replace ref does not fool plain git, so the fixture proves nothing: %v", err)
		}
		only(t, judge(t, in), false, map[string]int{"base-rewritten": 1})
		git(t, root, "replace", "-d", orphan)
		write(t, root, ".git/info/grafts", orphan+" "+base+"\n")
		only(t, judge(t, in), false, map[string]int{"base-rewritten": 1})
	})
	t.Run("gone from the repository", func(t *testing.T) {
		root, base := repo(t)
		in := input(root, base)
		in.Earlier = schema.Object{{Key: "base", Value: strings.Repeat("0123456789", 4)}}
		o := judge(t, in)
		only(t, o, false, map[string]int{"base-rewritten": 1})
		has(t, o, "(gone from the repository)")
	})
	t.Run("kept on the base branch", func(t *testing.T) {
		root, base := repo(t)
		in := input(root, base)
		in.Earlier = judge(t, in).Captures
		// The task lands on main: on main the merge base is HEAD, so the base the task's climbs recorded is kept.
		git(t, root, "checkout", "-q", "main")
		git(t, root, "merge", "-q", "--ff-only", "task")
		head := git(t, root, "rev-parse", "HEAD")
		in.MergeBase = head
		o := judge(t, in)
		only(t, o, false, nil)
		if v := capture(o, "base"); v != base {
			t.Errorf("base %v, want the recorded %s", v, base)
		}
		has(t, o, "kept from the task's earlier climb (HEAD is on main)")
		if RecordedBase(o.Captures) != base {
			t.Errorf("RecordedBase %q", RecordedBase(o.Captures))
		}
	})
}

// The earlier result is read back from a ladder result on disk, as the runner hands it on: its numbers are
// json.Number, and the round trip compares as the first climb's did.
func TestEarlierResultRoundTrip(t *testing.T) {
	isolate(t)
	root, base := repo(t)
	first := judge(t, input(root, base))
	raw, err := os.ReadFile(filepath.Join("..", "..", "formats", "examples", "ladder.json"))
	if err != nil {
		t.Fatal(err)
	}
	l, err := format.ReadLadder(raw)
	if err != nil {
		t.Fatal(err)
	}
	l.Rungs[0].Kind, l.Rungs[0].Captures = "ver-git", first.Captures
	enc, err := l.Encode()
	if err != nil {
		t.Fatalf("a result holding the captures does not encode: %v", err)
	}
	back, err := format.ReadLadder(enc)
	if err != nil {
		t.Fatal(err)
	}
	write(t, root, "a.txt", "x\n")
	git(t, root, "stash", "-q")
	in := input(root, base)
	in.Earlier = back.Rungs[0].Captures
	only(t, judge(t, in), false, map[string]int{"stash-added": 1})

	// Values of the wrong kind are not compared, and are no error.
	in.Earlier = schema.Object{{Key: "base", Value: int64(3)}, {Key: "stash-count", Value: "two"},
		{Key: "exclude-hash", Value: false}}
	o := judge(t, in)
	only(t, o, false, nil)
	for _, w := range []string{"exclude-edited", "stash-added", "base-rewritten"} {
		if _, judged := count(t, o, w); judged {
			t.Errorf("%s judged against a malformed earlier capture", w)
		}
	}
}

// The caller's git variables and a bad merge base cannot point the rung at another repository or into git's
// options.
func TestCallerCannotSteerGit(t *testing.T) {
	isolate(t)
	root, base := repo(t)
	git(t, root, "update-index", "--assume-unchanged", "a.txt")
	other, _ := repo(t)
	t.Setenv("GIT_DIR", filepath.Join(other, ".git"))
	t.Setenv("GIT_WORK_TREE", other)
	t.Setenv("GIT_INDEX_FILE", filepath.Join(other, ".git", "index"))
	t.Setenv("GIT_CONFIG_PARAMETERS", "'core.fsmonitor'='true'")
	o := judge(t, input(root, base))
	only(t, o, true, map[string]int{"assume-unchanged": 1})
	for _, k := range []string{"GIT_DIR", "GIT_WORK_TREE", "GIT_INDEX_FILE", "GIT_CONFIG_PARAMETERS"} {
		must(t, os.Unsetenv(k))
	}
	in := input(root, "--output="+filepath.Join(root, "x"))
	o = judge(t, in)
	only(t, o, true, map[string]int{"assume-unchanged": 1})
	has(t, o, "is not a commit id, so it is not used")
	if _, err := os.Stat(filepath.Join(root, "x")); err == nil {
		t.Errorf("the merge base reached git as an option")
	}
}

// error only when git cannot be read: no git, a folder that is no checkout, a broken repository, no time.
func TestErrorOnlyWhenGitCannotBeRead(t *testing.T) {
	isolate(t)
	root, base := repo(t)
	earlier := judge(t, input(root, base)).Captures
	check := func(t *testing.T, in Input, why string) {
		t.Helper()
		in.Earlier = earlier
		o := judge(t, in)
		if o.Status != "error" || o.Hides {
			t.Fatalf("status %s, Hides %v: %s", o.Status, o.Hides, o.Reason)
		}
		has(t, o, "git cannot be read, so git integrity is not judged: ", why,
			"(the records are the task's earlier climb's, carried)")
		if RecordedBase(o.Captures) != base {
			t.Errorf("the earlier base is not carried: %v", o.Captures)
		}
		for _, w := range format.GitFindings[3:] {
			if _, judged := count(t, o, w.Word); judged {
				t.Errorf("%s judged in an error", w.Word)
			}
		}
	}
	t.Run("no git on the PATH", func(t *testing.T) {
		t.Setenv("PATH", t.TempDir())
		check(t, input(root, base), "git is not on the PATH")
	})
	t.Run("a git program that is not there", func(t *testing.T) {
		in := input(root, base)
		in.Git = filepath.Join(t.TempDir(), "no-git")
		check(t, in, "git could not be run")
	})
	t.Run("no checkout", func(t *testing.T) {
		check(t, input(t.TempDir(), base), "git cannot read the checkout")
	})
	t.Run("a broken index", func(t *testing.T) {
		r, b := repo(t)
		write(t, r, ".git/index", "not an index")
		check(t, input(r, b), "git cannot read the index")
	})
	t.Run("a broken HEAD", func(t *testing.T) {
		r, b := repo(t)
		write(t, r, ".git/HEAD", "garbage\n")
		check(t, input(r, b), "git cannot read the checkout")
	})
	t.Run("no time", func(t *testing.T) {
		in := input(root, base)
		in.Timeout = time.Nanosecond
		check(t, in, "git did not answer in time")
	})
}

func TestReasonNamesAtMost20(t *testing.T) {
	isolate(t)
	root, base := repo(t)
	var names []string
	for i := 0; i < 25; i++ {
		n := fmt.Sprintf("f%02d.txt", i)
		write(t, root, n, "x\n")
		names = append(names, n)
	}
	git(t, root, "add", ".")
	git(t, root, "commit", "-q", "-m", "files")
	git(t, root, append([]string{"update-index", "--assume-unchanged"}, names...)...)
	o := judge(t, input(root, base))
	only(t, o, true, map[string]int{"assume-unchanged": 25})
	has(t, o, "f19.txt, and 5 more")
	if strings.Contains(o.Reason, "f20.txt") {
		t.Errorf("more than 20 names: %s", o.Reason)
	}
}

// A name that is not printable ASCII is quoted, so the reason stays one line of ASCII.
func TestNamesStayASCII(t *testing.T) {
	isolate(t)
	root, base := repo(t)
	appendExclude(t, root, "caf*\n")
	write(t, root, "caf\u00e9.txt", "x\n")
	o := judge(t, input(root, base))
	only(t, o, true, map[string]int{"exclude-hidden": 1, "exclude-lines": 1})
	has(t, o, `exclude-hidden 1: "caf\u00e9.txt" (by info/exclude)`)
}

func TestTestFile(t *testing.T) {
	go1, tap := []string{format.TestsGoJSON}, []string{format.TestsTAP}
	for _, c := range []struct {
		forms []string
		path  string
		want  bool
	}{
		{go1, "a_test.go", true},
		{go1, "internal/x/a_test.go", true},
		{go1, "internal/x/testdata/in.txt", true},
		{go1, "testdata/a/b.json", true},
		{go1, "internal/x/a.go", false},
		{go1, "a_test.go.orig", false},
		{go1, "src/x.test.js", false},
		{tap, "src/x.test.js", true},
		{tap, "test/helpers/h.mjs", true},
		{tap, "test.ts", true},
		{tap, "src/x.js", false},
		{[]string{format.TestsJUnit}, "a_test.go", false},
		{[]string{"unknown", format.TestsGoJSON}, "a_test.go", true},
		{nil, "a_test.go", false},
	} {
		if got := TestFile(c.forms, c.path); got != c.want {
			t.Errorf("TestFile(%v, %s) = %v", c.forms, c.path, got)
		}
	}
}

// Every word the package reads or writes is a word of format.GitFindings, its one home; every hiding finding is one.
func TestWordsHaveOneHome(t *testing.T) {
	words := map[string]bool{}
	for _, w := range format.GitFindings {
		words[w.Word] = true
	}
	for _, k := range []string{keyBase, keyStashCount, keyExcludeHash, findTestFile, findAssume, findSkip,
		findExcludeEdit, findStash, findRewritten, findExcludeLine, findHidden, findUntrackedGI, findFilter} {
		if !words[k] {
			t.Errorf("%s is not a word of format.GitFindings", k)
		}
	}
	for k := range hiding {
		if !words[k] {
			t.Errorf("hiding finding %s is not a word of format.GitFindings", k)
		}
		for _, w := range format.GitFindings {
			if w.Word == k && !strings.Contains(w.Means, "it hides a change, and the run counts as dirty") {
				t.Errorf("%s hides, but its line does not say so", k)
			}
		}
	}
	if len(words) != 13 {
		t.Errorf("format.GitFindings has %d words; this package knows 13", len(words))
	}
	if runtime.GOOS == "" {
		t.Skip()
	}
}
