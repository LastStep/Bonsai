package workspace

// Tests of .bonsai/local/'s finder (ReadGitLayout, FindLocal) and EnsureGitignore. The finder reads .git with no git
// process; the tests that make real checkouts hold its answer to git's own (Find) and skip only where git is not on
// the PATH (as TestFind does); the tests that lay the files out by hand run everywhere.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
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

// FindLocal writes to the main checkout only when it is verified: a worktree of the project, with git's link back
// and this machine's record of the main checkout, writes main's local/; a .git file pointing at another project's
// worktree, at a main checkout with no bonsai.yaml, at one with no record on this machine, or a worktree whose link
// back names another, leaves the writer in its own checkout. A main checkout whose bonsai.yaml has a problem away
// from its id still counts. Laid out by hand, so it runs on every system with no git.
func TestFindLocal(t *testing.T) {
	const id, other = "ws-aaaaaaaaaaaaaaaaaaaaaaaaaa", "ws-bbbbbbbbbbbbbbbbbbbbbbbbbb"
	yaml := func(id string) string {
		return "format: bonsai.workspace/1\nid: " + id + "\nname: demo\npacks: []\n"
	}
	tmp := resolved(t, t.TempDir())
	home := filepath.Join(tmp, "home")
	t.Setenv(HomeEnv, home)
	main := filepath.Join(tmp, "main")
	wt := filepath.Join(tmp, "wt")
	write(t, main, ".git/HEAD", "ref: refs/heads/main\n")
	write(t, main, ".git/worktrees/wt/HEAD", "ref: refs/heads/feature\n")
	write(t, main, ".git/worktrees/wt/commondir", "../..\n")
	write(t, main, ".git/worktrees/wt/gitdir", filepath.ToSlash(filepath.Join(wt, ".git"))+"\n")
	write(t, main, "bonsai.yaml", yaml(id))
	write(t, wt, ".git", "gitdir: ../main/.git/worktrees/wt\n")
	write(t, wt, "bonsai.yaml", yaml(id))

	// No record of the main checkout on this machine yet: the worktree writes its own local/.
	if l := FindLocal(wt, id); l.Main != wt || l.Branch != "feature" {
		t.Errorf("a main checkout this machine has no record of: %+v", l)
	}
	at := time.Date(2026, 10, 11, 9, 0, 0, 0, time.UTC)
	if _, err := RecordCheckout(home, main, id, at); err != nil {
		t.Fatal(err)
	}
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
	// The link back, relative to the worktree's git folder, as git 2.48's worktree.useRelativePaths writes it.
	write(t, main, ".git/worktrees/wt/gitdir", "../../../../wt/.git\n")
	if l := FindLocal(wt, id); l.Main != main {
		t.Errorf("a relative link back: %+v", l)
	}
	// The link back naming another folder: git's own record says the .git file is not this worktree's.
	write(t, main, ".git/worktrees/wt/gitdir", filepath.ToSlash(filepath.Join(tmp, "elsewhere", ".git"))+"\n")
	if l := FindLocal(wt, id); l.Main != wt {
		t.Errorf("a link back naming another folder: %+v", l)
	}
	write(t, main, ".git/worktrees/wt/gitdir", filepath.ToSlash(filepath.Join(wt, ".git"))+"\n")
	// Another project's worktree: its main holds another id.
	stranger := filepath.Join(tmp, "stranger")
	write(t, stranger, ".git/worktrees/x/HEAD", "ref: refs/heads/x\n")
	write(t, stranger, ".git/worktrees/x/commondir", "../..\n")
	write(t, stranger, ".git/worktrees/x/gitdir", filepath.ToSlash(filepath.Join(wt, ".git"))+"\n")
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

// mainCase is one project for VerifyMain's tests: a main checkout made by git with bonsai.yaml holding id, and a
// worktree of it, in tmp, recorded on this machine in home at since unless since is zero.
type mainCase struct {
	tmp, home, main, wt string
}

const verifyID = "ws-vvvvvvvvvvvvvvvvvvvvvvvvvv"

var since0 = time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC)

func newMainCase(t *testing.T) mainCase {
	t.Helper()
	tmp := resolved(t, t.TempDir())
	isolateGit(t, tmp)
	c := mainCase{tmp: tmp, home: filepath.Join(tmp, "home"), main: filepath.Join(tmp, "main"), wt: filepath.Join(tmp, "wt")}
	write(t, c.main, "bonsai.yaml", "format: bonsai.workspace/1\nid: "+verifyID+"\nname: demo\npacks: []\n")
	git(t, c.main, "init", "-q")
	git(t, c.main, "add", "bonsai.yaml")
	git(t, c.main, "commit", "-q", "-m", "linked")
	git(t, c.main, "worktree", "add", "-q", "-b", "feature", c.wt)
	if _, err := RecordCheckout(c.home, c.main, verifyID, since0); err != nil {
		t.Fatal(err)
	}
	return c
}

// verify is VerifyMain on the checkout at root, as FindLocal asks it: its own id, and its main checkout's.
func (c mainCase) verify(root string) Trust {
	g := ReadGitLayout(root)
	return VerifyMain(c.home, g, configID(root), configID(g.Main))
}

// decoy makes a folder that looks like a main checkout of the project (bonsai.yaml with its id unless id says
// another, and a git folder holding the worktree's entry), and points the worktree's .git file at it, as an agent
// could; linkBack forges git's link back too. It gives the decoy's folder.
func (c mainCase) decoy(t *testing.T, id string, linkBack bool) string {
	t.Helper()
	d := filepath.Join(c.tmp, "decoy")
	write(t, d, "bonsai.yaml", "format: bonsai.workspace/1\nid: "+id+"\nname: demo\npacks: []\n")
	write(t, d, ".git/HEAD", "ref: refs/heads/main\n")
	write(t, d, ".git/worktrees/wt/HEAD", "ref: refs/heads/feature\n")
	write(t, d, ".git/worktrees/wt/commondir", "../..\n")
	if linkBack {
		write(t, d, ".git/worktrees/wt/gitdir", filepath.ToSlash(filepath.Join(c.wt, ".git"))+"\n")
	}
	write(t, c.wt, ".git", "gitdir: "+filepath.ToSlash(filepath.Join(d, ".git", "worktrees", "wt"))+"\n")
	return d
}

func wantTrust(t *testing.T, name string, got Trust, verified bool, why string) {
	t.Helper()
	if got.Verified != verified || !strings.Contains(got.Why, why) || (!verified && got.Next == "") {
		t.Errorf("%s: %+v; want verified %v, why with %q, and a next step", name, got, verified, why)
	}
	if verified && (got.Why != "" || got.Next != "") {
		t.Errorf("%s: verified with a reason: %+v", name, got)
	}
	for _, r := range got.Why + got.Next {
		if r > 0x7e {
			t.Errorf("%s: not ASCII: %q", name, got.Why+got.Next)
			break
		}
	}
}

// The main checkout, verified (design/plan-5.md, 5.3.1 note 2), on checkouts made by git: a worktree and its main
// checkout, recorded; the same with no record (a clone that never ran bonsai update here); a submodule, its own main.
func TestVerifyMain(t *testing.T) {
	c := newMainCase(t)
	wantTrust(t, "the main checkout", c.verify(c.main), true, "")
	wantTrust(t, "a worktree", c.verify(c.wt), true, "")
	wantTrust(t, "no Bonsai home", VerifyMain("", ReadGitLayout(c.wt), verifyID, verifyID), false, "no Bonsai home")
	wantTrust(t, "a folder with no .git", VerifyMain(c.home, ReadGitLayout(c.home), verifyID, ""), false, "not a checkout's top")
	wantTrust(t, "no id", VerifyMain(c.home, ReadGitLayout(c.main), "", ""), false, "no workspace id")
	// No record: neither verifies, and the step is a person's bonsai update in the main checkout.
	if err := os.RemoveAll(filepath.Join(c.home, "workspaces")); err != nil {
		t.Fatal(err)
	}
	wantTrust(t, "the main checkout, no record", c.verify(c.main), false, "no record of the main checkout")
	got := c.verify(c.wt)
	wantTrust(t, "a worktree, no record", got, false, "no record of the main checkout")
	if !strings.Contains(got.Next, "a person runs bonsai update in the project's main checkout") {
		t.Errorf("the step: %q", got.Next)
	}
	// A record of the main checkout holding another id (bonsai update has not run since the id changed).
	if _, err := RecordCheckout(c.home, c.main, "ws-oooooooooooooooooooooooooo", since0); err != nil {
		t.Fatal(err)
	}
	wantTrust(t, "a record of another id", c.verify(c.wt), false, "holds another workspace id")

	// A submodule: its own main, verified by its own record alone.
	c = newMainCase(t)
	lib := filepath.Join(c.tmp, "lib")
	write(t, lib, "bonsai.yaml", "format: bonsai.workspace/1\nid: ws-ssssssssssssssssssssssssss\nname: lib\npacks: []\n")
	git(t, lib, "init", "-q")
	git(t, lib, "add", "bonsai.yaml")
	git(t, lib, "commit", "-q", "-m", "lib")
	git(t, c.main, "-c", "protocol.file.allow=always", "submodule", "add", "-q", filepath.ToSlash(lib), "sub")
	sub := filepath.Join(c.main, "sub")
	if g := ReadGitLayout(sub); g.Worktree() || g.GitDir == "" {
		t.Fatalf("the submodule's layout: %+v", g)
	}
	wantTrust(t, "a submodule, no record", c.verify(sub), false, "no record of the main checkout")
	if _, err := RecordCheckout(c.home, sub, "ws-ssssssssssssssssssssssssss", since0); err != nil {
		t.Fatal(err)
	}
	wantTrust(t, "a submodule, recorded", c.verify(sub), true, "")
	wantTrust(t, "the superproject still", c.verify(c.main), true, "")
}

// A main checkout redirected (design/plan-5.md, 5.3.1 note 2; "5.3 done" check 6): the worktree's .git file pointed at
// a decoy main, with and without git's link back forged, with and without the decoy's bonsai.yaml holding the id,
// and with a decoy record written later than the main checkout's, and in the same second; a forged commondir; a
// worktree made to look like its own main. None is verified, and FindLocal keeps the worktree's records in its own
// folder.
func TestVerifyMainRefusesARedirectedMain(t *testing.T) {
	c := newMainCase(t)
	d := c.decoy(t, verifyID, false)
	wantTrust(t, "a decoy, no link back", c.verify(c.wt), false, "link back")
	c.decoy(t, verifyID, true)
	wantTrust(t, "a decoy with the link back forged, no record", c.verify(c.wt), false, "no record of the main checkout")
	c.decoy(t, "ws-dddddddddddddddddddddddddd", true)
	wantTrust(t, "a decoy holding another id", c.verify(c.wt), false, "does not hold this workspace's id")
	c.decoy(t, verifyID, true)
	if _, err := RecordCheckout(c.home, d, verifyID, since0.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	got := c.verify(c.wt)
	wantTrust(t, "a decoy recorded later", got, false, "claimed this workspace's id before its main checkout did")
	if !strings.Contains(got.Next, filepath.ToSlash(c.main)) {
		t.Errorf("the step does not name the checkout that claimed first: %q", got.Next)
	}
	t.Setenv(HomeEnv, c.home)
	if l := FindLocal(c.wt, verifyID); l.Main != c.wt {
		t.Errorf("a worktree pointed at a recorded decoy writes %s", l.Main)
	}
	// The real main checkout is still verified, and a worktree pointed back at it too.
	wantTrust(t, "the main checkout beside a decoy", c.verify(c.main), true, "")
	write(t, c.wt, ".git", "gitdir: "+filepath.ToSlash(filepath.Join(c.main, ".git", "worktrees", "wt"))+"\n")
	wantTrust(t, "the worktree pointed back", c.verify(c.wt), true, "")

	// A tie: the decoy recorded in the same second as the main checkout. Neither is trusted.
	c = newMainCase(t)
	d = c.decoy(t, verifyID, true)
	if _, err := RecordCheckout(c.home, d, verifyID, since0); err != nil {
		t.Fatal(err)
	}
	wantTrust(t, "a tie, through the decoy", c.verify(c.wt), false, "in the same second")
	wantTrust(t, "a tie, the main checkout itself", c.verify(c.main), false, "in the same second")

	// A forged commondir: the worktree's git folder names the decoy's common folder. The worktree then reads as its
	// own main, which this machine has no record of; recorded later by bonsai init run in it, it is still not first.
	c = newMainCase(t)
	d = c.decoy(t, verifyID, true)
	write(t, c.wt, ".git", "gitdir: "+filepath.ToSlash(filepath.Join(c.main, ".git", "worktrees", "wt"))+"\n")
	write(t, c.main, ".git/worktrees/wt/commondir", filepath.ToSlash(filepath.Join(d, ".git"))+"\n")
	if g := ReadGitLayout(c.wt); g.Worktree() {
		t.Fatalf("a forged commondir still reads as the main's worktree: %+v", g)
	}
	wantTrust(t, "a forged commondir", c.verify(c.wt), false, "no record of the main checkout")
	if _, err := RecordCheckout(c.home, c.wt, verifyID, since0.Add(2*time.Hour)); err != nil {
		t.Fatal(err)
	}
	wantTrust(t, "a forged commondir, the worktree recorded later", c.verify(c.wt), false, "claimed this workspace's id before")

	// A worktree made to look like its own main: its .git file replaced by a folder.
	c = newMainCase(t)
	if err := os.Remove(filepath.Join(c.wt, ".git")); err != nil {
		t.Fatal(err)
	}
	write(t, c.wt, ".git/HEAD", "ref: refs/heads/main\n")
	wantTrust(t, "a worktree made its own main", c.verify(c.wt), false, "no record of the main checkout")
}

// A damaged record (one Bonsai cannot read): the main checkout's own is not verified while it is damaged, nor is a
// decoy recorded later, since the damaged record may hold the earliest claim; bonsai update does not write over it.
func TestVerifyMainWithADamagedRecord(t *testing.T) {
	c := newMainCase(t)
	d := c.decoy(t, verifyID, true)
	if _, err := RecordCheckout(c.home, d, verifyID, since0.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	dir, err := MachineDir(c.home, c.main)
	if err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(dir, MachineRecordFile)
	for name, damage := range map[string]string{
		"not JSON":         "{\"path\": ",
		"a since not time": `{"path": "` + filepath.ToSlash(c.main) + `", "ids": [{"id": "` + verifyID + `", "since": "yesterday"}]}`,
		"ids not a list":   `{"path": "` + filepath.ToSlash(c.main) + `", "ids": 7}`,
	} {
		if err := os.WriteFile(file, []byte(damage), 0o644); err != nil {
			t.Fatal(err)
		}
		wantTrust(t, name+": the decoy", c.verify(c.wt), false, "a record in the Bonsai home cannot be read")
		wantTrust(t, name+": the main checkout", c.verify(c.main), false, "cannot be read")
		if wrote, err := RecordCheckout(c.home, c.main, verifyID, since0.Add(3*time.Hour)); wrote || err == nil ||
			!strings.Contains(err.Error(), "not written over") || !strings.Contains(err.Error(), "a person checks") {
			t.Errorf("%s: bonsai update wrote over it: %v %v", name, wrote, err)
		}
		if raw, _ := os.ReadFile(file); string(raw) != damage {
			t.Errorf("%s: the damaged record changed: %s", name, raw)
		}
	}
	// A person deletes it; update writes it again, its claim starting now: the decoy, recorded earlier, wins nothing
	// either, since the main checkout's worktree no longer points at it, and the main checkout is not first.
	if err := os.Remove(file); err != nil {
		t.Fatal(err)
	}
	if _, err := RecordCheckout(c.home, c.main, verifyID, since0.Add(3*time.Hour)); err != nil {
		t.Fatal(err)
	}
	wantTrust(t, "the main checkout recorded again after the decoy", c.verify(c.main), false, "claimed this workspace's id before")
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
