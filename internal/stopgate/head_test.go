package stopgate

// Head, the commit at a checkout's HEAD read with no git process: held to git's own answer (git rev-parse HEAD) on a
// real repository, its worktree, its refs packed and a detached HEAD; then each form of git's files made by hand, and
// each HEAD it cannot read.

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/testpack"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// headOf reads the HEAD of the checkout at root.
func headOf(root string) (string, error) { return Head(workspace.ReadGitLayout(root)) }

func TestHeadIsGitsAnswer(t *testing.T) {
	tmp, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	testpack.Isolate(t, tmp)
	main := filepath.Join(tmp, "main")
	if err := os.MkdirAll(main, 0o755); err != nil {
		t.Fatal(err)
	}
	testpack.Git(t, main, "init", "-q")
	write(t, main, "a.txt", "a\n")
	testpack.Git(t, main, "add", "-A")
	testpack.Git(t, main, "commit", "-q", "-m", "one")
	testpack.Git(t, main, "tag", "-a", "-m", "a tag", "v1")
	write(t, main, "a.txt", "b\n")
	testpack.Git(t, main, "commit", "-q", "-am", "two")
	wt := filepath.Join(tmp, "wt")
	testpack.Git(t, main, "worktree", "add", "-q", "-b", "task", wt)
	write(t, wt, "b.txt", "b\n")
	testpack.Git(t, wt, "add", "-A")
	testpack.Git(t, wt, "commit", "-q", "-m", "three")

	same := func(what, root string) {
		t.Helper()
		want := testpack.Git(t, root, "rev-parse", "HEAD")
		got, err := headOf(root)
		if err != nil || got != want {
			t.Errorf("%s: Head %q (%v), git %q", what, got, err, want)
		}
	}
	same("main, a loose ref", main)
	same("a worktree, its branch in the common folder", wt)
	testpack.Git(t, main, "pack-refs", "--all")
	if _, err := os.Stat(filepath.Join(main, ".git", "refs", "heads", "main")); err == nil {
		t.Fatal("pack-refs left the loose ref")
	}
	same("main, its refs packed", main)
	same("a worktree, its refs packed", wt)
	testpack.Git(t, main, "checkout", "-q", "--detach", "HEAD~1")
	same("a detached HEAD", main)
	testpack.Git(t, main, "checkout", "-q", "v1")
	same("a tag checked out", main)
	testpack.Git(t, wt, "checkout", "-q", "--detach")
	same("a worktree's detached HEAD", wt)
}

// Each form of git's files made by hand, and each HEAD that cannot be read, with its words.
func TestHeadForms(t *testing.T) {
	const sha = headSHA
	sha256 := strings.Repeat("ab", 32)
	cases := []struct {
		name  string
		files map[string]string // under the checkout; "" removes nothing, every file is written
		want  string            // the commit, or "" when it cannot be read
		why   string            // in the error
	}{
		{"a loose ref", map[string]string{".git/HEAD": "ref: refs/heads/main\n", ".git/refs/heads/main": sha + "\n"}, sha, ""},
		{"CRLF and spaces", map[string]string{".git/HEAD": "ref:  refs/heads/main \r\n", ".git/refs/heads/main": sha + "\r\n"}, sha, ""},
		{"upper case read as lower", map[string]string{".git/HEAD": strings.ToUpper(sha) + "\n"}, sha, ""},
		{"a SHA-256 repository", map[string]string{".git/HEAD": "ref: refs/heads/main\n", ".git/refs/heads/main": sha256 + "\n"}, sha256, ""},
		{"a branch with a slash", map[string]string{".git/HEAD": "ref: refs/heads/t/0901\n", ".git/refs/heads/t/0901": sha}, sha, ""},
		{"a symbolic ref", map[string]string{".git/HEAD": "ref: refs/heads/alias\n", ".git/refs/heads/alias": "ref: refs/heads/main\n",
			".git/refs/heads/main": sha}, sha, ""},
		{"packed, the loose ref first", map[string]string{".git/HEAD": "ref: refs/heads/main\n", ".git/refs/heads/main": sha,
			".git/packed-refs": otherSHA + " refs/heads/main\n"}, sha, ""},
		{"packed, with its header and a peeled tag", map[string]string{".git/HEAD": "ref: refs/heads/main\n",
			".git/packed-refs": "# pack-refs with: peeled fully-peeled sorted \r\n" + otherSHA + " refs/tags/main\n^" + otherSHA + "\n" +
				sha + " refs/heads/main\r\n"}, sha, ""},
		{"no git folder", map[string]string{"x": ""}, "", "the checkout holds no .git that Bonsai reads"},
		{"no HEAD", map[string]string{".git/config": ""}, "", "HEAD in its git folder cannot be read (it is not there)"},
		{"an empty HEAD", map[string]string{".git/HEAD": ""}, "", "holds neither a commit nor a ref"},
		{"a short commit", map[string]string{".git/HEAD": sha[:12] + "\n"}, "", "holds neither a commit nor a ref"},
		{"a new branch", map[string]string{".git/HEAD": "ref: refs/heads/main\n"}, "", `HEAD names "refs/heads/main", which has no commit yet`},
		{"a ref file that is not a commit", map[string]string{".git/HEAD": "ref: refs/heads/main\n", ".git/refs/heads/main": "zz\n"},
			"", `the ref file "refs/heads/main" holds neither a commit nor a ref`},
		{"a ref out of the git folder", map[string]string{".git/HEAD": "ref: refs/../../x\n", "x": sha}, "", "which is not a ref name git writes"},
		{"a ref not under refs/", map[string]string{".git/HEAD": "ref: ORIG_HEAD\n", ".git/ORIG_HEAD": sha}, "", "which is not a ref name git writes"},
		{"a ref with a backslash", map[string]string{".git/HEAD": "ref: refs\\heads\\main\n"}, "", "which is not a ref name git writes"},
		{"an absolute ref", map[string]string{".git/HEAD": "ref: /etc/x\n"}, "", "which is not a ref name git writes"},
		{"a loop of symbolic refs", map[string]string{".git/HEAD": "ref: refs/heads/a\n", ".git/refs/heads/a": "ref: refs/heads/b\n",
			".git/refs/heads/b": "ref: refs/heads/a\n"}, "", "more than five deep"},
		{"a reftable", map[string]string{".git/HEAD": "ref: refs/heads/.invalid\n", ".git/reftable/tables.list": ""}, "", "keeps its refs in a reftable"},
		{"a reftable with a plain branch name", map[string]string{".git/HEAD": "ref: refs/heads/main\n", ".git/reftable/tables.list": ""},
			"", "keeps its refs in a reftable"},
		{"a ref file too large", map[string]string{".git/HEAD": "ref: refs/heads/main\n", ".git/refs/heads/main": sha + strings.Repeat("\n", 5000)},
			"", "larger than any file git writes there"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			root := t.TempDir()
			for rel, body := range c.files {
				write(t, root, rel, body)
			}
			got, err := headOf(root)
			switch {
			case c.want != "" && (err != nil || got != c.want):
				t.Errorf("Head %q (%v), want %q", got, err, c.want)
			case c.want == "" && (err == nil || !strings.Contains(err.Error(), c.why)):
				t.Errorf("Head %q, error %v; want an error saying %q", got, err, c.why)
			case err != nil && unsafeOut.MatchString(safe(err.Error())):
				t.Errorf("an unsafe word: %v", err)
			}
		})
	}
}

// A worktree's HEAD is its own (in its git folder under the common one), its per-worktree refs in its own folder and
// its branches in the common folder.
func TestHeadOfAWorktreeByHand(t *testing.T) {
	tmp := t.TempDir()
	main, wt := filepath.Join(tmp, "main"), filepath.Join(tmp, "wt")
	write(t, main, ".git/HEAD", "ref: refs/heads/main\n")
	write(t, main, ".git/refs/heads/main", otherSHA+"\n")
	write(t, main, ".git/refs/heads/task", headSHA+"\n")
	write(t, main, ".git/worktrees/wt/HEAD", "ref: refs/heads/task\n")
	write(t, main, ".git/worktrees/wt/commondir", "../..\n")
	write(t, main, ".git/worktrees/wt/gitdir", filepath.ToSlash(filepath.Join(wt, ".git"))+"\n")
	write(t, wt, ".git", "gitdir: "+filepath.ToSlash(filepath.Join(main, ".git", "worktrees", "wt"))+"\n")
	if got, err := headOf(wt); err != nil || got != headSHA {
		t.Errorf("the worktree's HEAD %q (%v)", got, err)
	}
	if got, err := headOf(main); err != nil || got != otherSHA {
		t.Errorf("main's HEAD %q (%v)", got, err)
	}
	write(t, main, ".git/worktrees/wt/HEAD", "ref: refs/worktree/mine\n")
	write(t, main, ".git/worktrees/wt/refs/worktree/mine", headSHA+"\n")
	if got, err := headOf(wt); err != nil || got != headSHA {
		t.Errorf("a per-worktree ref: %q (%v)", got, err)
	}
}

// Head starts no git process: with git off the PATH it still reads.
func TestHeadStartsNoGit(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not on the PATH, so there is nothing to take off it")
	}
	root := t.TempDir()
	write(t, root, ".git/HEAD", "ref: refs/heads/main\n")
	write(t, root, ".git/packed-refs", headSHA+" refs/heads/main\n")
	t.Setenv("PATH", "")
	if got, err := headOf(root); err != nil || got != headSHA {
		t.Errorf("Head with no PATH: %q (%v)", got, err)
	}
}
