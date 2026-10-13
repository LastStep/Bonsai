package ladder

// New tests must fail on the base, through the runner (plan-5 5.4.5's proof; kinds.go's afterRungs): on fixture
// projects whose rung runs go test -json, and node's test runner in TAP where node is on the PATH, a climb whose count
// rose, green and clean, writes the ratchet's new_tests: a new test failing at the base, one passing there and one that
// does not build or load there, each counted, named in the human lines; on the base branch, the Bless ask's why and its
// data.rises[].new_tests carry them; with no base, "check not run: no base"; a stop during the check leaves the result
// written and every new test "check not run"; a dirty climb runs no check; on the base branch, the base git integrity
// recorded on the task's branch. Every climb leaves no base worktree in the system's temporary folder (the test's own)
// and git's list of worktrees as it was.

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/ratchet"
	"github.com/LastStep/Bonsai/internal/record"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/testpack"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// ntForm is one tests form's fixture project: its base files (committed with the fixture), its branch's files laid
// over them (a new test failing at the base, one passing there, one that cannot build or load there), and those
// three tests' names.
type ntForm struct {
	name, needs, command, ratchet string
	base, branch                  map[string]string
	failed, passed, notRun        string
}

// The go fixture: Sub adds at the base and the branch fixes it; Mul is right at the base; Neg is new on the branch.
var ntGo = ntForm{name: format.TestsGoJSON, needs: "go", command: "go test -json ./...", ratchet: "go_tests",
	base: map[string]string{
		"go.mod":            "module example.com/nt\n\ngo 1.21\n",
		"calc/calc.go":      "package calc\n\nfunc Add(a, b int) int { return a + b }\n\nfunc Sub(a, b int) int { return a + b }\n",
		"calc/calc_test.go": "package calc\n\nimport \"testing\"\n\nfunc TestAdd(t *testing.T) {\n\tif Add(2, 3) != 5 {\n\t\tt.Fatal(\"Add\")\n\t}\n}\n",
		"mul/mul.go":        "package mul\n\nfunc Mul(a, b int) int { return a * b }\n",
		"neg/neg.go":        "package neg\n\nfunc Abs(a int) int {\n\tif a < 0 {\n\t\treturn -a\n\t}\n\treturn a\n}\n",
	},
	branch: map[string]string{
		"calc/calc.go":     "package calc\n\nfunc Add(a, b int) int { return a + b }\n\nfunc Sub(a, b int) int { return a - b }\n",
		"calc/sub_test.go": "package calc\n\nimport \"testing\"\n\nfunc TestSub(t *testing.T) {\n\tif Sub(5, 3) != 2 {\n\t\tt.Fatal(\"Sub\")\n\t}\n}\n",
		"mul/mul_test.go":  "package mul\n\nimport \"testing\"\n\nfunc TestMul(t *testing.T) {\n\tif Mul(2, 3) != 6 {\n\t\tt.Fatal(\"Mul\")\n\t}\n}\n",
		"neg/neg.go": "package neg\n\nfunc Abs(a int) int {\n\tif a < 0 {\n\t\treturn -a\n\t}\n\treturn a\n}\n\n" +
			"func Neg(a int) int { return -a }\n",
		"neg/neg_test.go": "package neg\n\nimport \"testing\"\n\nfunc TestNeg(t *testing.T) {\n\tif Neg(2) != -2 {\n\t\tt.Fatal(\"Neg\")\n\t}\n}\n",
	},
	failed: "example.com/nt/calc.TestSub", passed: "example.com/nt/mul.TestMul", notRun: "example.com/nt/neg.TestNeg"}

// The TAP fixture, node's test runner: the same three, in .mjs files under test/.
var ntTAP = ntForm{name: format.TestsTAP, needs: "node", command: "node --test --test-reporter=tap", ratchet: "node_tests",
	base: map[string]string{
		"src/calc.mjs":       "export const add = (a, b) => a + b;\nexport const sub = (a, b) => a + b;\n",
		"src/mul.mjs":        "export const mul = (a, b) => a * b;\n",
		"src/neg.mjs":        "export const abs = (a) => (a < 0 ? -a : a);\n",
		"test/calc.test.mjs": tapTest("add works", "add", "calc", "add(2, 3), 5"),
	},
	branch: map[string]string{
		"src/calc.mjs":      "export const add = (a, b) => a + b;\nexport const sub = (a, b) => a - b;\n",
		"src/neg.mjs":       "export const abs = (a) => (a < 0 ? -a : a);\nexport const neg = (a) => -a;\n",
		"test/sub.test.mjs": tapTest("sub works", "sub", "calc", "sub(5, 3), 2"),
		"test/mul.test.mjs": tapTest("mul works", "mul", "mul", "mul(2, 3), 6"),
		"test/neg.test.mjs": tapTest("neg works", "neg", "neg", "neg(2), -2"),
	},
	failed: "sub works", passed: "mul works", notRun: "neg works"}

// tapTest is a node test file: one test named name, importing fn from ../src/<file>.mjs and asserting args equal.
func tapTest(name, fn, file, args string) string {
	return "import { test } from 'node:test';\nimport assert from 'node:assert';\nimport { " + fn + " } from '../src/" + file + ".mjs';\n\n" +
		"test('" + name + "', () => {\n  assert.strictEqual(" + args + ");\n});\n"
}

// need skips the test when the form's program is not on the PATH.
func (n ntForm) need(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath(n.needs); err != nil {
		t.Skipf("%s is not on the PATH: the %s fixture's rung runs it", n.needs, n.name)
	}
}

// rung is the fixture's one rung, its ratchet's floor 1 (the base's one test), and its base_setup given.
func (n ntForm) rung(setup string) rung {
	return rung{n: 1, name: "tests", command: n.command, required: true, ratchet: n.ratchet, tst: n.name, setup: setup}
}

// project makes the fixture project on main with the base files, its floor rung 1 and the more rungs given, and gives
// it with the system's temporary folder, which is the test's own.
func (n ntForm) project(t *testing.T, setup string, more ...rung) (*fx, string) {
	t.Helper()
	files := map[string]string{"work/tasks/T-0901-a.md": taskText("T-0901", "running", "")}
	for k, v := range n.base {
		files[k] = v
	}
	floor := "[1]"
	for _, r := range more {
		floor = strings.TrimSuffix(floor, "]") + ", " + strconv.Itoa(r.n) + "]"
	}
	f := project(t, yamlOf(floor, "\n  "+n.ratchet+": 1", "[]", append([]rung{n.rung(setup)}, more...)...), files)
	tmp, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS == "windows" {
		t.Setenv("TMP", tmp)
		t.Setenv("TEMP", tmp)
	} else {
		t.Setenv("TMPDIR", tmp)
	}
	return f, tmp
}

// lay writes the branch's files into dir and commits them there.
func (n ntForm) lay(t *testing.T, dir string) {
	t.Helper()
	for rel, body := range n.branch {
		p := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	testpack.Git(t, dir, "add", "-A")
	testpack.Git(t, dir, "commit", "-q", "-m", "the work")
}

// climbClean climbs and checks that the climb left no base worktree in tmp and git's worktrees as they were.
func climbClean(t *testing.T, f *fx, tmp, dir string, o Options) *Answer {
	t.Helper()
	before := testpack.Git(t, f.root, "worktree", "list", "--porcelain")
	a := f.climb(dir, o)
	if left, _ := filepath.Glob(filepath.Join(tmp, "bonsai-newtests-*")); len(left) > 0 {
		t.Errorf("left in the temporary folder: %q", left)
	}
	if after := testpack.Git(t, f.root, "worktree", "list", "--porcelain"); after != before {
		t.Errorf("git's worktrees changed:\n%s\nwas\n%s", after, before)
	}
	return a
}

// newTestsOf is the result's first rung's new_tests, shown as JSON ("null" for none).
func newTestsOf(t *testing.T, f *fx, name string) string {
	t.Helper()
	rt := f.result(name).Rungs[0].Ratchet
	if rt == nil {
		t.Fatalf("no ratchet:\n%s", f.out.String())
	}
	if rt.NewTests == nil {
		return "null"
	}
	nt := rt.NewTests
	return schema.Show(schema.Object{{Key: "base", Value: nt.Base}, {Key: "count", Value: nt.Count},
		{Key: "failed_on_base", Value: nt.FailedOnBase}, {Key: "passed_on_base", Value: nt.PassedOnBase}, {Key: "not_run", Value: nt.NotRun}})
}

// counts is new_tests as JSON.
func counts(base string, count, failed, passed, notRun int) string {
	return schema.Show(schema.Object{{Key: "base", Value: base}, {Key: "count", Value: int64(count)},
		{Key: "failed_on_base", Value: int64(failed)}, {Key: "passed_on_base", Value: int64(passed)}, {Key: "not_run", Value: int64(notRun)}})
}

// hasLines checks the human lines hold each part.
func hasLines(t *testing.T, f *fx, parts ...string) {
	t.Helper()
	for _, p := range parts {
		if !strings.Contains(f.out.String(), p) {
			t.Errorf("the human lines lack %q:\n%s", p, f.out.String())
		}
	}
}

// A task's branch, climbed green and clean in its worktree: the base is the merge base with main; the new test
// failing at the base, the one passing there and the one that cannot build or load there each counted and named. A
// branch climb files no Bless ask.
func TestNewTestsBranch(t *testing.T) {
	for _, n := range []ntForm{ntGo, ntTAP} {
		t.Run(n.name, func(t *testing.T) {
			n.need(t)
			f, tmp := n.project(t, "")
			base := f.head()
			wt := filepath.Join(f.tmp, "wt")
			testpack.Git(t, f.root, "worktree", "add", "-q", "-b", "t0901", wt)
			n.lay(t, wt)
			a := climbClean(t, f, tmp, wt, Options{Task: "T-0901"})
			if a.Exit != 0 || len(a.Doc.Bless) != 0 {
				t.Fatalf("exit %d, bless %s\n%s", a.Exit, blessOps(a), f.out.String())
			}
			if got, want := newTestsOf(t, f, "T-0901"), counts(base, 3, 1, 1, 1); got != want {
				t.Errorf("new_tests %s, want %s\n%s", got, want, f.out.String())
			}
			hasLines(t, f, "new tests, rung 1 tests ("+n.ratchet+" 1->4): checking them on the base",
				": 3 new tests, run on the base "+base[:12]+": 1 failed there, 1 proves nothing new, 1 check not run",
				"    failed on the base: "+n.failed+"\n", "    proves nothing new (passed on the base): "+n.passed+"\n",
				"    check not run (the base run did not report them, or skipped them: their package did not build or their file did "+
					"not load there): "+n.notRun+"\n")
		})
	}
}

// On the base branch, where the merge base is HEAD itself: the base the task's last result recorded (a ver-git rung's
// base capture), and the Bless ask's why and data.rises[].new_tests carry the numbers; with none recorded, "check not
// run: no base", the count the rise.
func TestNewTestsOnBase(t *testing.T) {
	for _, n := range []ntForm{ntGo, ntTAP} {
		t.Run(n.name, func(t *testing.T) {
			n.need(t)
			f, tmp := n.project(t, "")
			base := f.head()
			n.lay(t, f.root)

			a := climbClean(t, f, tmp, "", Options{Task: "T-0901"})
			if a.Exit != 0 || blessOps(a) != "file ladder:bless-T-0901" {
				t.Fatalf("no base: exit %d, bless %s\n%s", a.Exit, blessOps(a), f.out.String())
			}
			if got, want := newTestsOf(t, f, "T-0901"), counts("", 3, 0, 0, 3); got != want {
				t.Errorf("no base: new_tests %s, want %s", got, want)
			}
			noBase := "3 new tests (the count's rise; they could not be named): check not run: no base"
			hasLines(t, f, ": "+noBase+"\n")
			if why := a.Doc.Bless[0].String("why"); !strings.Contains(why, n.ratchet+" rose from 1 to 4, "+noBase+".") {
				t.Errorf("no base: the why %q", why)
			}

			prior(t, f, base)
			f.out.Reset()
			a = climbClean(t, f, tmp, "", Options{Task: "T-0901"})
			if a.Exit != 0 || blessOps(a) != "file ladder:bless-T-0901" {
				t.Fatalf("exit %d, bless %s\n%s", a.Exit, blessOps(a), f.out.String())
			}
			want := counts(base, 3, 1, 1, 1)
			if got := newTestsOf(t, f, "T-0901"); got != want {
				t.Errorf("new_tests %s, want %s\n%s", got, want, f.out.String())
			}
			checked := "3 new tests, run on the base " + base[:12] + ": 1 failed there, 1 proves nothing new, 1 check not run"
			hasLines(t, f, ": "+checked+"\n", "    failed on the base: "+n.failed+"\n")
			rec := a.Doc.Bless[0]
			data, _ := rec.Get("data")
			if why := rec.String("why"); why != "A green, clean climb of main at "+f.head()[:8]+": "+n.ratchet+" rose from 1 to 4, "+checked+"." {
				t.Errorf("the why %q", why)
			}
			if !strings.Contains(schema.Show(data), `"rises":[{"name":"`+n.ratchet+`","was":1,"now":4,"new_tests":`+want+`}]`) {
				t.Errorf("the data %s", schema.Show(data))
			}
			// The ask as stored is the climb's record.
			if lines := askLines(t, f.root); len(lines) != 2 || schema.Show(lines[1]) != schema.Show(rec) {
				t.Errorf("the asks file: %v", lines)
			}
		})
	}
}

// prior writes the task's last result as a ver-git rung recording base would leave it (5.4.6's base capture), so a
// climb on the base branch finds its base.
func prior(t *testing.T, f *fx, base string) {
	t.Helper()
	task, cmd := "T-0901", "echo"
	l := &format.Ladder{Task: &task, Workspace: testID, Mode: ModeLocal, Started: "2026-10-13T01:00:00.000Z", Finished: "2026-10-13T01:01:00.000Z",
		Git: format.LadderGit{SHA: base, Branch: "t0901"}, Requested: []int64{1, 5}, Green: true, Skipped: []format.LadderSkip{},
		Rungs: []format.LadderRung{{Rung: 1, Name: "tests", Kind: "command", Required: true, Command: &cmd, Status: "green"},
			{Rung: 5, Name: "git", Kind: "ver-git", Status: "green", Captures: schema.Object{{Key: "base", Value: base}}}}}
	raw, err := l.Encode()
	if err != nil {
		t.Fatal(err)
	}
	f.write(".bonsai/local/ladder/T-0901.json", string(raw))
	fingerprinted(t, f.root, "T-0901", raw)
}

// fingerprinted appends the ladder record a climb writes after its result (the runner's fingerprint), so a result a
// test writes by hand reads as the task's last climb's (an earlier result is read only with its record).
func fingerprinted(t *testing.T, main, task string, raw []byte) {
	t.Helper()
	sum := sha256.Sum256(raw)
	l := record.New("ladder", record.Common{Workspace: testID, Local: workspace.Local{Root: main, Main: main}, At: time.Now()})
	rel, kind, text, ok := workspace.LocalDir+"/ladder/"+task+".json", ModeLocal, "sha256:"+hex.EncodeToString(sum[:]), true
	l.Target, l.Kind, l.Text, l.OK, l.Task = &rel, &kind, &text, &ok, &task
	if _, err := record.WriteLog(main, l); err != nil {
		t.Fatal(err)
	}
}

// A stop during the check, here while base_setup runs in the base worktree: the base run is ended at once, the result
// is still written, green, every new test "check not run", and no worktree is left.
func TestNewTestsStopped(t *testing.T) {
	ntGo.need(t)
	setup := "echo started> started.txt && sleep 60"
	if runtime.GOOS == "windows" {
		setup = "echo started> started.txt && ping -n 61 127.0.0.1 >NUL"
	}
	f, tmp := ntGo.project(t, setup)
	wt := filepath.Join(f.tmp, "wt")
	testpack.Git(t, f.root, "worktree", "add", "-q", "-b", "t0901", wt)
	ntGo.lay(t, wt)
	ctx, stop := context.WithCancel(context.Background())
	defer stop()
	defer func(saved func(ratchet.RiseInput) ratchet.Checked) { newTests = saved }(newTests)
	newTests = func(in ratchet.RiseInput) ratchet.Checked {
		go func() {
			defer stop()
			for end := time.Now().Add(time.Minute); time.Now().Before(end); time.Sleep(20 * time.Millisecond) {
				if m, _ := filepath.Glob(filepath.Join(tmp, "bonsai-newtests-*", "base", "started.txt")); len(m) > 0 {
					return
				}
			}
		}()
		return ratchet.NewTests(in)
	}
	start := time.Now()
	a := climbClean(t, f, tmp, wt, Options{Task: "T-0901", Stop: ctx})
	if a.Exit != 0 || a.Error != nil || a.Doc.Result == nil {
		t.Fatalf("exit %d, %v\n%s", a.Exit, a.Error, f.out.String())
	}
	if took := time.Since(start); took > 45*time.Second {
		t.Errorf("the climb took %v: the stop did not end base_setup", took)
	}
	if got, want := newTestsOf(t, f, "T-0901"), counts(f.head(), 3, 0, 0, 3); got != want {
		t.Errorf("new_tests %s, want %s", got, want)
	}
	hasLines(t, f, "3 new tests (the count's rise; they could not be named): check not run: base_setup was stopped (the climb was stopped)")
}

// A dirty climb, on a branch or on the base branch, runs no check: new_tests null, no check's lines.
func TestNewTestsDirty(t *testing.T) {
	ntGo.need(t)
	f, tmp := ntGo.project(t, "")
	base := f.head()
	wt := filepath.Join(f.tmp, "wt")
	testpack.Git(t, f.root, "worktree", "add", "-q", "-b", "t0901", wt)
	ntGo.lay(t, wt)
	if err := os.WriteFile(filepath.Join(wt, "loose.txt"), []byte("not committed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	climbClean(t, f, tmp, wt, Options{Task: "T-0901"})
	if got := newTestsOf(t, f, "T-0901"); got != "null" || strings.Contains(f.out.String(), "new tests") {
		t.Errorf("a dirty branch: new_tests %s\n%s", got, f.out.String())
	}
	ntGo.lay(t, f.root)
	prior(t, f, base)
	f.write("loose.txt", "not committed\n")
	f.out.Reset()
	climbClean(t, f, tmp, "", Options{Task: "T-0901"})
	if got := newTestsOf(t, f, "T-0901"); got != "null" || strings.Contains(f.out.String(), "new tests") {
		t.Errorf("a dirty base branch: new_tests %s\n%s", got, f.out.String())
	}
}

// The base on the base branch, as git integrity records it: the task's branch climbed in its worktree (its ver-git
// rung recording the merge base), the branch merged into main, and main climbed: its check runs on the base the
// branch's result recorded, and the Bless ask carries the numbers.
func TestNewTestsRecordedBase(t *testing.T) {
	n := ntGo
	n.need(t)
	f, tmp := n.project(t, "", rung{n: 5, name: "git", kind: KindVerGit})
	base := f.head()
	wt := filepath.Join(f.tmp, "wt")
	testpack.Git(t, f.root, "worktree", "add", "-q", "-b", "t0901", wt)
	n.lay(t, wt)
	if a := climbClean(t, f, tmp, wt, Options{Task: "T-0901"}); a.Exit != 0 {
		t.Fatalf("the branch: exit %d\n%s", a.Exit, f.out.String())
	}
	if got := schema.Show(f.result("T-0901").Rungs[1].Captures); !strings.Contains(got, `"base":"`+base+`"`) {
		t.Fatalf("the ver-git rung recorded no base: %s\n%s", got, f.out.String())
	}
	testpack.Git(t, f.root, "merge", "-q", "--ff-only", "t0901")
	f.out.Reset()
	a := climbClean(t, f, tmp, "", Options{Task: "T-0901"})
	if a.Exit != 0 || blessOps(a) != "file ladder:bless-T-0901" {
		t.Fatalf("main: exit %d, bless %s\n%s", a.Exit, blessOps(a), f.out.String())
	}
	if got, want := newTestsOf(t, f, "T-0901"), counts(base, 3, 1, 1, 1); got != want {
		t.Errorf("new_tests %s, want %s\n%s", got, want, f.out.String())
	}
	if why := a.Doc.Bless[0].String("why"); !strings.Contains(why, "3 new tests, run on the base "+base[:12]+": 1 failed there") {
		t.Errorf("the why %q", why)
	}
}
