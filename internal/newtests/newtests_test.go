package newtests

// The check's judgments on reports made by hand: which tests are new, how a parent failing only by its subtests is
// judged, names reported twice, skips at the base, when the base's own run cannot name its tests, the base chosen,
// the inputs refused before anything runs, and the words.

import (
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/rungjob"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/testout"
)

// rep makes a report: each test as "name=outcome", a trailing ^ marking a parent failing only by its subtests.
func rep(tests ...string) testout.Report {
	r := testout.Report{Form: format.TestsGoJSON}
	for _, s := range tests {
		name, outcome, _ := strings.Cut(s, "=")
		parent := strings.HasSuffix(outcome, "^")
		r.Tests = append(r.Tests, testout.Test{Name: name, Outcome: strings.TrimSuffix(outcome, "^"), Parent: parent})
	}
	return r
}

func TestNewTests(t *testing.T) {
	head := rep("a=passed", "b=passed", "c=skipped", "d=todo", "e=passed", "b=passed", "f=passed")
	base := rep("a=passed", "e=failed", "x=passed")
	if got, want := newTests(head, base), []string{"b", "f"}; !reflect.DeepEqual(got, want) {
		t.Errorf("newTests = %q, want %q: passed at HEAD, not reported at the base in any outcome, once each", got, want)
	}
}

func TestJudge(t *testing.T) {
	first := rep("p=passed", "p/old=passed", "q=passed")
	cases := []struct {
		name                 string
		news                 []string
		second               testout.Report
		failed, passed, nrun []string
		why                  string
	}{
		{name: "a new subtest under an old parent: the parent's failure is its subtest's",
			news:   []string{"p/new", "n"},
			second: rep("p/old=passed", "p/new=failed", "p=failed^", "q=passed", "n=passed"),
			failed: []string{"p/new"}, passed: []string{"n"}},
		{name: "an old subtest failing: every new test not run",
			news:   []string{"p/new"},
			second: rep("p/old=failed", "p/new=failed", "p=failed^"),
			nrun:   []string{"p/new"}, why: "an old test failed at the base with the branch's test files copied in: p/old"},
		{name: "an old test failing on its own",
			news:   []string{"n"},
			second: rep("q=failed", "n=failed"),
			nrun:   []string{"n"}, why: "copied in: q"},
		{name: "an old test the second run did not run is not judged",
			news:   []string{"n"},
			second: rep("n=failed"),
			failed: []string{"n"}},
		{name: "skipped or todo at the base, or not reported: not run",
			news:   []string{"s", "d", "gone"},
			second: rep("s=skipped", "d=todo"),
			nrun:   []string{"s", "d", "gone"}},
		{name: "a name reported twice: failed when either failed",
			news:   []string{"twice", "both"},
			second: rep("twice=passed", "twice=failed", "both=passed", "both=passed"),
			failed: []string{"twice"}, passed: []string{"both"}},
		{name: "a new parent failing only by its new subtest counts failed",
			news:   []string{"n/a", "n/b", "n"},
			second: rep("n/a=failed", "n/b=passed", "n=failed^"),
			failed: []string{"n/a", "n"}, passed: []string{"n/b"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := &Result{}
			(&check{res: r}).judge(c.news, first, c.second)
			if !reflect.DeepEqual(r.Failed, c.failed) || !reflect.DeepEqual(r.Passed, c.passed) || !reflect.DeepEqual(r.NotRun, c.nrun) {
				t.Errorf("failed %q passed %q not run %q; want %q %q %q", r.Failed, r.Passed, r.NotRun, c.failed, c.passed, c.nrun)
			}
			nt := r.NewTests
			if nt.Count != int64(len(c.news)) || nt.FailedOnBase != int64(len(c.failed)) || nt.PassedOnBase != int64(len(c.passed)) ||
				nt.NotRun != int64(len(c.nrun)) {
				t.Errorf("counts %+v", nt)
			}
			if c.why == "" && r.Why != "" || !strings.Contains(r.Why, c.why) {
				t.Errorf("why %q, want %q", r.Why, c.why)
			}
		})
	}
}

func TestUnnamed(t *testing.T) {
	green := rungjob.Result{Exit: 0, Stdout: []byte("x"), StdoutSize: 1}
	red := rungjob.Result{Exit: 1, Tail: []string{"FAIL", "exit status 1"}}
	cases := []struct {
		name string
		res  rungjob.Result
		rep  testout.Report
		want string
	}{
		{"green", green, rep("a=passed"), ""},
		{"red by its failing tests alone", red, rep("a=failed"), ""},
		{"red with no failing test", red, rep("a=passed"), "ended with exit 1 and named no failing test: exit status 1"},
		{"a parent failing alone is no failing test", red, rep("a=failed^"), "ended with exit 1 and named no failing test"},
		{"a package that did not build", red, testout.Report{Broken: []testout.Broken{{What: "x/y", Build: true, Error: "undefined: Z"}}},
			"did not build the package x/y: undefined: Z"},
		{"a package failing outside its tests", red, testout.Report{Broken: []testout.Broken{{What: "x/y", Error: "panic: init"}}},
			"failed outside its tests in the package x/y: panic: init"},
		{"not a report", red, testout.Report{Broken: []testout.Broken{{Error: "the output is not a JUnit report: it is empty"}}},
			"printed no report it can read: the output is not a JUnit report"},
		{"timed out", rungjob.Result{Exit: -1, TimedOut: true}, rep(), "timed out after 30 s"},
		{"stopped", rungjob.Result{Exit: -1, Stopped: true}, rep(), "was stopped"},
		{"did not start", rungjob.Result{Exit: -1, Reason: "could not start: exec: no shell"}, rep(), "could not start: exec: no shell"},
		{"its output past the cap", rungjob.Result{Exit: 0, Stdout: []byte("x"), StdoutSize: 2}, rep("a=passed"), "past the"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := unnamed(c.res, c.rep, 30e9)
			if (c.want == "") != (got == "") || !strings.Contains(got, c.want) {
				t.Errorf("unnamed = %q, want %q", got, c.want)
			}
		})
	}
}

func TestChooseBase(t *testing.T) {
	merge, head, kept := strings.Repeat("a", 40), strings.Repeat("b", 40), strings.Repeat("c", 40)
	verGit := func(captures schema.Object) format.LadderRung {
		return format.LadderRung{Kind: "ver-git", Captures: captures}
	}
	prior := &format.Ladder{Rungs: []format.LadderRung{{Kind: "command", Captures: schema.Object{{Key: "base", Value: head}}},
		verGit(schema.Object{{Key: "base", Value: nil}}), verGit(schema.Object{{Key: "base", Value: strings.ToUpper(kept)}})}}
	cases := []struct {
		name, merge, head string
		prior             *format.Ladder
		want              string
	}{
		{"the merge base, on a branch", merge, head, prior, merge},
		{"on the base branch itself: the base the task's last result recorded", head, head, prior, kept},
		{"no merge base: the recorded one", "", head, prior, kept},
		{"a merge base that is no commit id", "--output=x", head, prior, kept},
		{"none recorded", head, head, &format.Ladder{Rungs: []format.LadderRung{verGit(nil)}}, ""},
		{"no earlier result", head, head, nil, ""},
	}
	for _, c := range cases {
		if got := ChooseBase(c.merge, c.head, c.prior); got != c.want {
			t.Errorf("%s: %q, want %q", c.name, got, c.want)
		}
	}
}

// Inputs the check refuses before it runs anything: no git call, no folder made.
func TestNotRunBeforeAnything(t *testing.T) {
	tmp := isolate(t)
	cmd, goJSON, other := "go test -json ./...", format.TestsGoJSON, "nunit"
	sha := strings.Repeat("a", 40)
	cases := []struct {
		name string
		rung format.Rung
		base string
		why  string
	}{
		{"no tests form", format.Rung{Command: &cmd}, sha, "the rung names no tests form"},
		{"a form Bonsai does not read", format.Rung{Command: &cmd, Tests: &other}, sha, `the rung's tests form "nunit" is not one Bonsai reads`},
		{"no command", format.Rung{Tests: &goJSON}, sha, "the rung has no command"},
		{"no base", format.Rung{Command: &cmd, Tests: &goJSON}, "", "no base"},
		{"a base that is no commit id", format.Rung{Command: &cmd, Tests: &goJSON}, "--output=x", "no base: --output=x is not a commit id"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := checked(t, tmp, Input{Root: "", Rung: c.rung, Base: c.base, Rise: 4})
			if r.Named || r.NewTests.Count != 4 || r.NewTests.NotRun != 4 || r.Why != c.why && !strings.HasPrefix(r.Why, c.why) {
				t.Errorf("got %+v, why %q", r.NewTests, r.Why)
			}
			if c.base != sha && r.NewTests.Base != "" {
				t.Errorf("base %q, want \"\"", r.NewTests.Base)
			}
		})
	}
	r := Check(Input{Rung: format.Rung{Command: &cmd, Tests: &goJSON}, Base: "", Rise: -2})
	if r.NewTests.Count != 0 || r.NewTests.NotRun != 0 {
		t.Errorf("a negative rise: %+v", r.NewTests)
	}
}

func TestWords(t *testing.T) {
	sha := "3b4c5d6e7f8091a2b3c4d5e6f708192a3b4c5d6e"
	many := make([]string, MaxNames)
	for i := range many {
		many[i] = "pkg.Test" + strconv.Itoa(i)
	}
	cases := []struct {
		r     Result
		lines []string
	}{
		{Result{Named: true, NewTests: format.LadderNewTests{Base: sha, Count: 3, FailedOnBase: 1, PassedOnBase: 1, NotRun: 1},
			Failed: []string{"a"}, Passed: []string{"b"}, NotRun: []string{"cé"}, Notes: []string{"a note"}},
			[]string{"3 new tests, run on the base 3b4c5d6e7f80: 1 failed there, 1 proves nothing new, 1 check not run",
				"  failed on the base: a", "  proves nothing new (passed on the base): b",
				"  check not run (" + notReported + "): \"c\\u00e9\"", "  note: a note"}},
		{Result{Named: true, NewTests: format.LadderNewTests{Base: sha, Count: 25, FailedOnBase: 25}, Failed: many},
			[]string{"25 new tests, run on the base 3b4c5d6e7f80: 25 failed there, 0 prove nothing new, 0 check not run",
				"  failed on the base: " + strings.Join(many, ", ") + ", and 5 more"}},
		{Result{Named: true, NewTests: format.LadderNewTests{Base: sha, Count: 2, NotRun: 2}, NotRun: []string{"a", "b"},
			Why: "an old test failed at the base: x"},
			[]string{"2 new tests: check not run: an old test failed at the base: x", "  check not run: a, b"}},
		{Result{NewTests: format.LadderNewTests{Count: 1, NotRun: 1}, Why: "no base"},
			[]string{"1 new test (the count's rise; they could not be named): check not run: no base"}},
		{Result{Named: true, NewTests: format.LadderNewTests{Base: sha}},
			[]string{"no new tests: every test the climb passed ran on the base 3b4c5d6e7f80 too"}},
	}
	for _, c := range cases {
		if got := c.r.Lines(); !reflect.DeepEqual(got, c.lines) {
			t.Errorf("Lines:\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(c.lines, "\n"))
		}
		if got := c.r.Summary(); got != c.lines[0] {
			t.Errorf("Summary %q", got)
		}
	}
}

// Copying the branch's test files in: a file copied over the base's, a deleted one removed, one gone from the working
// tree removed too; never through a link or a file the base holds where a folder should be.
func TestCopyIn(t *testing.T) {
	root, wt, outside := t.TempDir(), t.TempDir(), t.TempDir()
	tr := &tree{dir: wt}
	put := func(dir, rel, text string) {
		p := filepath.Join(dir, filepath.FromSlash(rel))
		must(t, os.MkdirAll(filepath.Dir(p), 0o755))
		must(t, os.WriteFile(p, []byte(text), 0o644))
	}
	read := func(dir, rel string) string {
		b, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(rel)))
		if err != nil {
			return "<none>"
		}
		return string(b)
	}
	put(root, "a/x_test.go", "new")
	put(wt, "a/x_test.go", "old")
	put(wt, "a/gone_test.go", "old")
	put(wt, "a/missing_test.go", "old")
	must(t, tr.copyIn(root, []change{{path: "a/x_test.go"}, {path: "a/gone_test.go", deleted: true},
		{path: "a/missing_test.go"}, {path: "b/deep/new_test.go"}}))
	if read(wt, "a/x_test.go") != "new" || read(wt, "a/gone_test.go") != "<none>" || read(wt, "a/missing_test.go") != "<none>" {
		t.Errorf("copied %q, deleted %q, missing %q", read(wt, "a/x_test.go"), read(wt, "a/gone_test.go"), read(wt, "a/missing_test.go"))
	}

	put(wt, "f", "a file where a folder should be")
	put(root, "f/y_test.go", "new")
	if err := tr.copyIn(root, []change{{path: "f/y_test.go"}}); err == nil || !strings.Contains(err.Error(), "not a folder") {
		t.Errorf("through a file: %v", err)
	}

	if err := os.Symlink(outside, filepath.Join(wt, "link")); err != nil {
		t.Skipf("a symbolic link cannot be made here (Windows without the privilege): %v", err)
	}
	put(root, "link/z_test.go", "new")
	put(outside, "kept_test.go", "outside")
	for _, c := range []change{{path: "link/z_test.go"}, {path: "link/kept_test.go", deleted: true}} {
		if err := tr.copyIn(root, []change{c}); err == nil || !strings.Contains(err.Error(), "as a link") {
			t.Errorf("%+v through a link: %v", c, err)
		}
	}
	if read(outside, "z_test.go") != "<none>" || read(outside, "kept_test.go") != "outside" {
		t.Errorf("a copy or a removal went through the link")
	}
}
