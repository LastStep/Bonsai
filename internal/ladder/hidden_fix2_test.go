package ladder

// Changes hidden from git status by a setting inside the repository that no diff shows (V2b's reproducers b1-ident,
// b1w-ident-worktree and b2-ident-committed, and its encoding probe), through the runner and the stop gate: the result counts the tree as dirty, names the path, rung 0 judges it by the guard's function,
// and the stop gate refuses the result.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/testpack"
)

// The fixture's code as HEAD holds it, and the edit the ident attribute reads as the same (V2b's b1: a buggy Sub in
// HEAD, the fix inside "$Id: ... $" in the checkout).
const (
	identCode = "the code /*$Id$*/ + 0*b\n"
	identFix  = "the code /*$Id: */ - b /* $*/ + 0*b\n"
)

// identHidden names the hidden file in the climb's human lines.
const identHidden = `    "src/main.txt": its bytes differ from HEAD's, though the ident attribute makes git read them as HEAD's` + "\n"

// hiddenClimb climbs in dir for T-0901 and holds the result to green but dirty, rung 0 having checked checked paths,
// the human lines to each of lines, and the stop gate in dir to its refusal of the dirty result.
func hiddenClimb(t *testing.T, f *fx, dir, checked string, lines ...string) {
	t.Helper()
	a := f.climb(dir, Options{Task: "T-0901"})
	l := f.result("T-0901")
	if a.Exit != 0 || !l.Git.Dirty || showOf(l.Rungs[0].Captures, "checked") != checked {
		t.Fatalf("exit %d, dirty %v, checked %s\n%s", a.Exit, l.Git.Dirty, showOf(l.Rungs[0].Captures, "checked"), f.out.String())
	}
	for _, line := range lines {
		if !strings.Contains(f.out.String(), line) {
			t.Errorf("the human lines lack %q:\n%s", line, f.out.String())
		}
	}
	if code, stderr := gate(t, dir, "T-0901"); code != 2 || !strings.Contains(stderr, "uncommitted changes") {
		t.Errorf("the stop gate: %d %q", code, stderr)
	}
}

// stageHidden writes rel in dir and stages it (the index then holds the edited file's times), and fails when plain
// git status shows anything: the fixture would prove nothing.
func stageHidden(t *testing.T, dir, rel, body string) {
	t.Helper()
	p := filepath.Join(dir, filepath.FromSlash(rel))
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	testpack.Git(t, dir, "add", rel)
	if s := testpack.Git(t, dir, "status", "--porcelain", "--untracked-files=all"); s != "" {
		t.Fatalf("plain git status shows the change, so the fixture proves nothing: %q", s)
	}
}

// infoAttributes writes the repository's info/attributes (the common git folder's, shared by its worktrees).
func infoAttributes(t *testing.T, root, text string) {
	t.Helper()
	dir := filepath.Join(root, ".git", "info")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "attributes"), []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

// b1: the ident attribute in info/attributes, the main checkout.
func TestHiddenByIdentInInfoAttributes(t *testing.T) {
	f := r0Project(t)
	f.write("src/main.txt", identCode)
	f.commit("a buggy Sub")
	infoAttributes(t, f.root, "src/main.txt ident\n")
	stageHidden(t, f.root, "src/main.txt", identFix)
	hiddenClimb(t, f, f.root, "1", identHidden)
}

// b1w: the same in an agent's worktree, the attribute in the common info/attributes; the climb and the gate from the
// worktree.
func TestHiddenByIdentInAWorktree(t *testing.T) {
	f := r0Project(t)
	wt := f.worktree("t0901")
	f.writeIn(wt, "src/main.txt", identCode)
	f.commitIn(wt, "a buggy Sub")
	infoAttributes(t, f.root, "src/main.txt ident\n")
	stageHidden(t, wt, "src/main.txt", identFix)
	hiddenClimb(t, f, wt, "1", identHidden)
}

// b2: the ident attribute committed in .gitattributes (in the diff, and still no clean tree).
func TestHiddenByIdentCommitted(t *testing.T) {
	f := r0Project(t)
	f.write(".gitattributes", "src/main.txt ident\n")
	f.write("src/main.txt", identCode)
	f.commit("a buggy Sub, attributes")
	stageHidden(t, f.root, "src/main.txt", identFix)
	hiddenClimb(t, f, f.root, "1", identHidden)
}

// working-tree-encoding in info/attributes: other bytes than HEAD's read as the same text.
func TestHiddenByAnEncoding(t *testing.T) {
	f := r0Project(t)
	f.write("src/want.txt", "caf\xc3\xa9\n")
	f.commit("golden")
	infoAttributes(t, f.root, "src/want.txt working-tree-encoding=ISO-8859-1\n")
	stageHidden(t, f.root, "src/want.txt", "caf\xe9\n")
	hiddenClimb(t, f, f.root, "1",
		`    "src/want.txt": its bytes differ from HEAD's, though the working-tree-encoding attribute makes git read them as HEAD's`+"\n")
}
