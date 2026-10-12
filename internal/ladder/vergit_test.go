package ladder

// Git integrity through the runner (plan-5 5.4.6, re-run through bonsai ladder; "5.4 done" check 11): a clean
// fixture's ver-git rung green with its records; a stash added since the task's last climb found from that result's
// records; the test files judged by the rung's own tests form, else the other rungs'; error only when git cannot be
// read.

import (
	"regexp"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/testpack"
)

// vgProject is a fixture whose floor is rung 1 (echo) and rung 5 (git integrity, not required), T-0901 running; more
// rungs may follow, outside the floor.
func vgProject(t *testing.T, more ...rung) *fx {
	t.Helper()
	rungs := append([]rung{{n: 1, name: "one", command: "echo one", required: true}, {n: 5, name: "git", kind: "ver-git"}}, more...)
	return project(t, yamlOf("[1, 5]", "{}", "[]", rungs...), map[string]string{
		"work/tasks/T-0901-a.md": taskText("T-0901", "running", ""),
		"src/main.txt":           "the code\n",
	})
}

// gitRung is the ver-git rung's entry in the task's result.
func gitRung(t *testing.T, f *fx) format.LadderRung {
	t.Helper()
	for _, r := range f.result("T-0901").Rungs {
		if r.Kind == KindVerGit {
			return r
		}
	}
	t.Fatal("no ver-git rung in the result")
	return format.LadderRung{}
}

// A clean fixture: the ver-git rung green, its reason not in the result but printed as a note, its captures the
// records (base, the merge base; stash-count; exclude-hash) and every finding's count, 0; the result clean, and the
// stop gate lets the session end on it.
func TestVerGitClean(t *testing.T) {
	f := vgProject(t)
	a := f.climb("", Options{Task: "T-0901"})
	if a.Exit != 0 {
		t.Fatalf("exit %d, %+v\n%s", a.Exit, a.Error, f.out.String())
	}
	r := gitRung(t, f)
	if r.Status != "green" || r.Reason != nil || f.result("T-0901").Git.Dirty || !strings.Contains(f.out.String(), "note: nothing found") {
		t.Fatalf("%+v, dirty %v\n%s", r, f.result("T-0901").Git.Dirty, f.out.String())
	}
	if got := showOf(r.Captures, "base"); got != `"`+f.head()+`"` {
		t.Errorf("base %s, want HEAD (the merge base on main)", got)
	}
	if !regexp.MustCompile(`^"sha256:[0-9a-f]{64}"$`).MatchString(showOf(r.Captures, "exclude-hash")) || showOf(r.Captures, "stash-count") != "0" {
		t.Errorf("the records: %s", schema.Show(r.Captures))
	}
	for _, w := range []string{"assume-unchanged", "skip-worktree", "exclude-lines", "exclude-hidden", "untracked-gitignore", "filter-local"} {
		if got := showOf(r.Captures, w); got != "0" {
			t.Errorf("%s %s", w, got)
		}
	}
	if code, stderr := gate(t, f.root, "T-0901"); code != 0 {
		t.Errorf("the stop gate refused a clean climb: %d %q", code, stderr)
	}
}

// A stash added since the task's last climb: the climb compares with the records that result holds (its stash count,
// its exclude-hash, read back from the result as written), and finds the stash and nothing else.
func TestVerGitStashAddedSinceTheLastClimb(t *testing.T) {
	f := vgProject(t)
	if a := f.climb("", Options{Task: "T-0901"}); a.Exit != 0 {
		t.Fatalf("the first climb: %d", a.Exit)
	}
	if r := gitRung(t, f); showOf(r.Captures, "stash-added") != "null" || showOf(r.Captures, "exclude-edited") != "null" {
		t.Fatalf("the first climb, with no earlier records, judged: %s", schema.Show(r.Captures))
	}
	f.write("src/main.txt", "put aside\n")
	testpack.Git(t, f.root, "stash", "-q")
	if a := f.climb("", Options{Task: "T-0901"}); a.Exit != 0 {
		t.Fatalf("the second climb: %d", a.Exit)
	}
	r := gitRung(t, f)
	if showOf(r.Captures, "stash-added") != "1" || showOf(r.Captures, "stash-count") != "1" || showOf(r.Captures, "exclude-edited") != "0" ||
		!strings.Contains(f.out.String(), "stash-added 1") {
		t.Errorf("the second climb: %s\n%s", schema.Show(r.Captures), f.out.String())
	}
}

// Which files are test files: the rung's own tests form when it names one, else the forms of the ladder's other
// rungs (any of bonsai.yaml's, climbed or not), in rung order.
func TestVerGitTestForms(t *testing.T) {
	f := vgProject(t, rung{n: 7, name: "go", command: "echo go", tst: "go-test-json"})
	wt := f.worktree("t0901")
	f.writeIn(wt, "pkg/a_test.go", "package pkg\n")
	f.writeIn(wt, "web/x.test.mjs", "// a test\n")
	f.commitIn(wt, "two test files")
	if a := f.climb(wt, Options{Task: "T-0901"}); a.Exit != 0 {
		t.Fatalf("exit %d\n%s", a.Exit, f.out.String())
	}
	if got := showOf(gitRung(t, f).Captures, "test-file-changed"); got != "1" || !strings.Contains(f.out.String(), "test-file-changed 1: pkg/a_test.go") {
		t.Errorf("by the other rung's go-test-json: %s\n%s", got, f.out.String())
	}
	tap := yamlOf("[1, 5]", "{}", "[]", rung{n: 1, name: "one", command: "echo one", required: true},
		rung{n: 5, name: "git", kind: "ver-git", tst: "tap"}, rung{n: 7, name: "go", command: "echo go", tst: "go-test-json"})
	f.write("bonsai.yaml", tap)
	f.out.Reset()
	if a := f.climb(wt, Options{Task: "T-0901"}); a.Exit != 0 {
		t.Fatalf("exit %d\n%s", a.Exit, f.out.String())
	}
	if got := showOf(gitRung(t, f).Captures, "test-file-changed"); got != "1" || !strings.Contains(f.out.String(), "test-file-changed 1: web/x.test.mjs") {
		t.Errorf("by its own tap: %s\n%s", got, f.out.String())
	}
}

// Git that cannot be read (a broken index) is the rung's only error: the climb red, the tree counted dirty (its status
// cannot be read), next naming git status; the rung's records carried from the task's last climb.
func TestVerGitErrorWhenGitCannotBeRead(t *testing.T) {
	f := vgProject(t)
	if a := f.climb("", Options{Task: "T-0901"}); a.Exit != 0 {
		t.Fatalf("the first climb: %d", a.Exit)
	}
	base := showOf(gitRung(t, f).Captures, "base")
	f.write(".git/index", "not an index")
	a := f.climb("", Options{Task: "T-0901"})
	r := gitRung(t, f)
	if a.Exit != ExitRed || r.Status != "error" || r.Reason == nil || !strings.Contains(*r.Reason, "git cannot read the index") ||
		!f.result("T-0901").Git.Dirty || showOf(r.Captures, "base") != base {
		t.Fatalf("exit %d, %+v, dirty %v\n%s", a.Exit, r, f.result("T-0901").Git.Dirty, f.out.String())
	}
	if a.Doc.Next == nil || a.Doc.Next.Do != "fix the git rung (rung 5): git cannot be read in this checkout (the rung's reason says why): "+
		"run: git status, and fix what it reports ; then run: bonsai ladder --task T-0901 --json" {
		t.Errorf("next: %+v", a.Doc.Next)
	}
}
