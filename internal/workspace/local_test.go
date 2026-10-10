package workspace

// Tests of .bonsai/local/'s finder (ReadGitLayout, FindLocal) and EnsureGitignore. The finder reads .git with no git
// process; the tests that make real checkouts hold its answer to git's own (Find) and skip only where git is not on
// the PATH (as TestFind does); the tests that lay the files out by hand run everywhere.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// write makes a file, and its folders, under dir.
func write(t *testing.T, dir, rel, text string) {
	t.Helper()
	path := filepath.Join(dir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

// real gives a path as Find gives it: symbolic links resolved (a temporary folder may sit behind one).
func resolved(t *testing.T, p string) string {
	t.Helper()
	r, err := filepath.EvalSymlinks(p)
	if err != nil {
		t.Fatalf("%s: %v", p, err)
	}
	return r
}

// The finder agrees with git (Find, git rev-parse --git-common-dir) on a main checkout, a worktree, a submodule, a
// checkout with a separate git folder and a worktree of that one, and reads each checkout's branch from its HEAD.
func TestReadGitLayoutAgreesWithGit(t *testing.T) {
	tmp := resolved(t, t.TempDir())
	isolateGit(t, tmp)
	main := filepath.Join(tmp, "main")
	write(t, main, "f.txt", "x\n")
	git(t, main, "init", "-q")
	git(t, main, "add", "f.txt")
	git(t, main, "commit", "-q", "-m", "one")
	wt := filepath.Join(tmp, "main-wt")
	git(t, main, "worktree", "add", "-q", "-b", "feature", wt)
	detached := filepath.Join(tmp, "main-detached")
	git(t, main, "worktree", "add", "-q", "--detach", detached)

	lib := filepath.Join(tmp, "lib")
	write(t, lib, "lib.txt", "lib\n")
	git(t, lib, "init", "-q")
	git(t, lib, "add", "lib.txt")
	git(t, lib, "commit", "-q", "-m", "lib")
	git(t, main, "-c", "protocol.file.allow=always", "submodule", "add", "-q", filepath.ToSlash(lib), "sub")
	sub := filepath.Join(main, "sub")

	sep := filepath.Join(tmp, "sep")
	git(t, tmp, "init", "-q", "--separate-git-dir", filepath.Join(tmp, "store"), sep)
	write(t, sep, "s.txt", "s\n")
	git(t, sep, "add", "s.txt")
	git(t, sep, "commit", "-q", "-m", "s")
	sepWT := filepath.Join(tmp, "sep-wt")
	git(t, sep, "worktree", "add", "-q", sepWT)

	for _, c := range []struct {
		name, root, main, branch string
	}{
		{"a main checkout", main, main, "main"},
		{"a worktree", wt, main, "feature"},
		{"a detached worktree", detached, main, ""},
		{"a submodule", sub, sub, ""},
		{"a separate git folder", sep, sep, "main"},
		{"a worktree of a separate git folder", sepWT, sepWT, "sep-wt"},
	} {
		g := ReadGitLayout(c.root)
		co, err := Find(c.root)
		if err != nil {
			t.Fatalf("%s: git: %v", c.name, err)
		}
		if !samePlace(resolved(t, g.Main), co.Main) || !samePlace(resolved(t, g.Root), co.Root) {
			t.Errorf("%s: the finder says root %s, main %s; git says root %s, main %s", c.name, g.Root, g.Main, co.Root, co.Main)
		}
		if !samePlace(resolved(t, g.Main), c.main) {
			t.Errorf("%s: main %s, want %s", c.name, g.Main, c.main)
		}
		if g.GitDir == "" {
			t.Errorf("%s: no git folder found", c.name)
		}
		if c.name == "a submodule" {
			continue // its HEAD is detached at the submodule's commit
		}
		if b := g.Branch(); b != c.branch {
			t.Errorf("%s: branch %q, want %q", c.name, b, c.branch)
		}
	}
}

// Layouts written by hand, so the rules run on every system with no git: a worktree's .git file and its commondir,
// relative or absolute, with CRLF line endings too; and every way a .git file fails to make a worktree, which leaves
// the checkout its own main.
func TestReadGitLayoutByHand(t *testing.T) {
	tmp := resolved(t, t.TempDir())
	main := filepath.Join(tmp, "main")
	write(t, main, ".git/HEAD", "ref: refs/heads/main\n")
	write(t, main, ".git/worktrees/wt/HEAD", "ref: refs/heads/t0001-thing\r\n")
	write(t, main, ".git/worktrees/wt/commondir", "../..\n")
	write(t, main, ".git/worktrees/abs/HEAD", "0123456789abcdef0123456789abcdef01234567\n")
	write(t, main, ".git/worktrees/abs/commondir", filepath.ToSlash(filepath.Join(main, ".git"))+"\r\n")
	write(t, main, ".git/modules/m/HEAD", "ref: refs/heads/main\n")
	write(t, main, ".git/modules/m/commondir", "../..\n")
	write(t, main, ".git/worktrees/a/b/commondir", "../../..\n")
	bare := filepath.Join(tmp, "bare.git")
	write(t, bare, "worktrees/w/HEAD", "ref: refs/heads/w\n")
	write(t, bare, "worktrees/w/commondir", "../..\n")

	g := ReadGitLayout(main)
	if g.Main != main || g.GitDir != filepath.Join(main, ".git") || g.Common != g.GitDir || g.Branch() != "main" {
		t.Errorf("a main checkout: %+v, branch %q", g, g.Branch())
	}
	for _, c := range []struct {
		name, gitfile, main, branch string
	}{
		{"a relative gitdir", "gitdir: ../main/.git/worktrees/wt\n", main, "t0001-thing"},
		{"an absolute gitdir, CRLF", "gitdir: " + filepath.ToSlash(filepath.Join(main, ".git", "worktrees", "wt")) + "\r\n", main, "t0001-thing"},
		{"an absolute commondir, a detached HEAD", "gitdir: ../main/.git/worktrees/abs\n", main, ""},
		{"a gitdir not under worktrees/", "gitdir: ../main/.git/modules/m\n", "", "main"},
		{"a gitdir two folders under worktrees/", "gitdir: ../main/.git/worktrees/a/b\n", "", ""},
		{"a common folder not named .git", "gitdir: ../bare.git/worktrees/w\n", "", "w"},
		{"a gitdir that is missing", "gitdir: ../main/.git/worktrees/gone\n", "", ""},
		{"no gitdir: line", "hello\n", "", ""},
		{"an empty gitdir", "gitdir:   \n", "", ""},
		{"an empty file", "", "", ""},
	} {
		root := filepath.Join(tmp, "co")
		if err := os.RemoveAll(root); err != nil {
			t.Fatal(err)
		}
		write(t, root, ".git", c.gitfile)
		want := c.main
		if want == "" {
			want = root
		}
		g := ReadGitLayout(root)
		if g.Root != root || g.Main != want {
			t.Errorf("%s: root %s, main %s; want main %s", c.name, g.Root, g.Main, want)
		}
		if b := g.Branch(); b != c.branch {
			t.Errorf("%s: branch %q, want %q", c.name, b, c.branch)
		}
	}
	none := filepath.Join(tmp, "plain")
	if err := os.MkdirAll(none, 0o755); err != nil {
		t.Fatal(err)
	}
	if g := ReadGitLayout(none); g.Main != none || g.GitDir != "" || g.Branch() != "" {
		t.Errorf("no .git: %+v", g)
	}
	// A .git file too long to be git's pointer is not read.
	big := filepath.Join(tmp, "big")
	write(t, big, ".git", "gitdir: ../main/.git/worktrees/wt\n"+strings.Repeat("x", 5000))
	if g := ReadGitLayout(big); g.Main != big {
		t.Errorf("a 5 KB .git file was read: %+v", g)
	}
}

// On Windows the finder compares paths case-blind, as the file system does; elsewhere letter case counts.
func TestFinderComparesCaseBlindOnWindows(t *testing.T) {
	wt, parent := filepath.FromSlash("/Repo/.GIT/Worktrees/wt"), filepath.FromSlash("/repo/.git/worktrees")
	if !childOf(wt, parent, true) || childOf(wt, parent, false) {
		t.Errorf("childOf: folded %v, not folded %v", childOf(wt, parent, true), childOf(wt, parent, false))
	}
	if !sameName(".GIT", ".git", true) || sameName(".GIT", ".git", false) {
		t.Errorf("sameName folds wrongly")
	}
	if childOf(parent, parent, true) || childOf(filepath.Join(parent, "a", "b"), parent, false) {
		t.Errorf("childOf takes the folder itself, or a folder two below")
	}
}

// FindLocal writes to the main checkout only when its bonsai.yaml holds the writer's id: a worktree of the project
// writes main's local/; a .git file pointing at another project's worktree, or at a main checkout with no bonsai.yaml,
// leaves the writer in its own checkout. A main checkout whose bonsai.yaml has a problem elsewhere still counts.
func TestFindLocal(t *testing.T) {
	const id, other = "ws-aaaaaaaaaaaaaaaaaaaaaaaaaa", "ws-bbbbbbbbbbbbbbbbbbbbbbbbbb"
	yaml := func(id string) string {
		return "format: bonsai.workspace/1\nid: " + id + "\nname: demo\npacks: []\n"
	}
	tmp := resolved(t, t.TempDir())
	main := filepath.Join(tmp, "main")
	write(t, main, ".git/HEAD", "ref: refs/heads/main\n")
	write(t, main, ".git/worktrees/wt/HEAD", "ref: refs/heads/feature\n")
	write(t, main, ".git/worktrees/wt/commondir", "../..\n")
	write(t, main, "bonsai.yaml", yaml(id))
	wt := filepath.Join(tmp, "wt")
	write(t, wt, ".git", "gitdir: ../main/.git/worktrees/wt\n")
	write(t, wt, "bonsai.yaml", yaml(id))

	if l := FindLocal(main, id); l.Root != main || l.Main != main || l.Branch != "main" {
		t.Errorf("the main checkout: %+v", l)
	}
	if l := FindLocal(wt, id); l.Root != wt || l.Main != main || l.Branch != "feature" ||
		l.Dir() != filepath.Join(main, ".bonsai", "local") {
		t.Errorf("a worktree: %+v, dir %s", l, l.Dir())
	}
	if l := FindLocal(wt, other); l.Main != wt {
		t.Errorf("a writer of another id: %+v", l)
	}
	if l := FindLocal(wt, ""); l.Main != wt {
		t.Errorf("a writer with no id: %+v", l)
	}
	// Another project's worktree: its main holds another id.
	stranger := filepath.Join(tmp, "stranger")
	write(t, stranger, ".git/worktrees/x/HEAD", "ref: refs/heads/x\n")
	write(t, stranger, ".git/worktrees/x/commondir", "../..\n")
	write(t, stranger, "bonsai.yaml", yaml(other))
	write(t, wt, ".git", "gitdir: ../stranger/.git/worktrees/x\n")
	if l := FindLocal(wt, id); l.Main != wt || l.Branch != "x" {
		t.Errorf("a .git file pointing at another project: %+v", l)
	}
	// A main checkout with no bonsai.yaml, then one whose bonsai.yaml has a problem away from its id.
	write(t, wt, ".git", "gitdir: ../main/.git/worktrees/wt\n")
	if err := os.Remove(filepath.Join(main, "bonsai.yaml")); err != nil {
		t.Fatal(err)
	}
	if l := FindLocal(wt, id); l.Main != wt {
		t.Errorf("a main checkout with no bonsai.yaml: %+v", l)
	}
	write(t, main, "bonsai.yaml", yaml(id)+"protected: 7\n")
	if _, err := LoadConfig(main); err == nil {
		t.Fatal("the broken bonsai.yaml reads")
	}
	if l := FindLocal(wt, id); l.Main != main {
		t.Errorf("a main checkout whose bonsai.yaml has a problem away from its id: %+v", l)
	}
	write(t, main, "bonsai.yaml", "format: [\n")
	if l := FindLocal(wt, id); l.Main != wt {
		t.Errorf("a main checkout whose bonsai.yaml does not read: %+v", l)
	}
}

// EnsureGitignore writes Bonsai's .bonsai/.gitignore when it is missing, and leaves one that is there, changed or not.
func TestEnsureGitignore(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, ".bonsai", ".gitignore")
	if wrote, err := EnsureGitignore(root); err != nil || !wrote {
		t.Fatalf("missing: wrote %v, %v", wrote, err)
	}
	if b, _ := os.ReadFile(path); string(b) != GitignoreText {
		t.Errorf("wrote %q", b)
	}
	if wrote, err := EnsureGitignore(root); err != nil || wrote {
		t.Errorf("present: wrote %v, %v", wrote, err)
	}
	if err := os.WriteFile(path, []byte("local/\nmine/\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if wrote, err := EnsureGitignore(root); err != nil || wrote {
		t.Errorf("changed: wrote %v, %v", wrote, err)
	}
	if b, _ := os.ReadFile(path); string(b) != "local/\nmine/\n" {
		t.Errorf("a changed file was written over: %q", b)
	}
	if !strings.Contains(GitignoreText, "\nlocal/\n") || GitignoreFile != ".bonsai/.gitignore" || LocalDir != ".bonsai/local" {
		t.Errorf("the text no longer ignores local/")
	}
	// .bonsai is a file: an error, and nothing written.
	odd := t.TempDir()
	if err := os.WriteFile(filepath.Join(odd, ".bonsai"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if wrote, err := EnsureGitignore(odd); err == nil || wrote {
		t.Errorf(".bonsai a file: wrote %v, %v", wrote, err)
	}
}
