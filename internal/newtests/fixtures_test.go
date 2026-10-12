package newtests

// The check on fixture repositories, for each tests form (plan-5 5.4.5's proof): go test -json, node's TAP (node
// --test) and JUnit (go test -json through a converter that names each test's file). Each repository is built in a
// temporary folder by the machine's git (Windows git on Windows) from testdata/: the base commit on main, then a task
// branch one commit ahead. Its HEAD run is the rung run at HEAD through internal/rungjob and read by internal/testout,
// as the runner does. The system's temporary folder is the test's own, so every case can check that no base worktree
// is left in it and that git's list of worktrees is as before.

import (
	"bytes"
	"context"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/rungjob"
	"github.com/LastStep/Bonsai/internal/testout"
)

// isolate gives the test git with no user or system configuration and no git variable from the caller, and a
// temporary folder of its own (TMPDIR; TMP and TEMP on Windows), which it returns.
func isolate(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not on the PATH: the check makes its base worktree with git")
	}
	global := filepath.Join(t.TempDir(), "gitconfig")
	must(t, os.WriteFile(global, nil, 0o644))
	t.Setenv("GIT_CONFIG_GLOBAL", global)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	for _, k := range []string{"GIT_DIR", "GIT_WORK_TREE", "GIT_INDEX_FILE", "GIT_COMMON_DIR", "GIT_OBJECT_DIRECTORY",
		"GIT_CONFIG_COUNT", "GIT_CONFIG_PARAMETERS", "GIT_CEILING_DIRECTORIES"} {
		t.Setenv(k, "")
		must(t, os.Unsetenv(k))
	}
	tmp, err := filepath.EvalSymlinks(t.TempDir())
	must(t, err)
	if runtime.GOOS == "windows" {
		t.Setenv("TMP", tmp)
		t.Setenv("TEMP", tmp)
	} else {
		t.Setenv("TMPDIR", tmp)
	}
	return tmp
}

// git runs git in dir for a fixture and gives its standard output, trimmed.
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

// layer copies a testdata folder's files into root, over any there.
func layer(t *testing.T, root, testdata string) {
	t.Helper()
	src, err := filepath.Abs(filepath.Join("testdata", filepath.FromSlash(testdata)))
	must(t, err)
	must(t, filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, err := filepath.Rel(src, p)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		dst := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		return os.WriteFile(dst, data, 0o644)
	}))
}

// write writes files into root, by project-relative path; "" removes the file.
func write(t *testing.T, root string, files map[string]string) {
	t.Helper()
	names := make([]string, 0, len(files))
	for n := range files {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		p := filepath.Join(root, filepath.FromSlash(n))
		if files[n] == "" {
			must(t, os.Remove(p))
			continue
		}
		must(t, os.MkdirAll(filepath.Dir(p), 0o755))
		must(t, os.WriteFile(p, []byte(files[n]), 0o644))
	}
}

// form is one tests form's fixture.
type form struct {
	name    string
	needs   string   // the program its rung runs: go or node
	command string   // the rung's command
	base    []string // testdata folders laid down for the base commit, in order
	branch  []string // testdata folders laid over it for the branch's commit
	// the new tests of the three-outcome case, by group
	failed, passed, notRun []string
	broken                 map[string]string // an old test failing at the base, which the branch deletes
	oldChanged             map[string]string // the branch changes an old test so it fails at the base
	oldName                string            // that old test's name
	news                   int               // how many new tests the branch's tests are, with oldChanged laid over
}

const (
	goCalc = "example.com/nt/calc."
	goMul  = "example.com/nt/mul."
	goNeg  = "example.com/nt/neg."
)

var goBroken = map[string]string{"calc/broken_test.go": "package calc\n\nimport \"testing\"\n\n" +
	"func TestBroken(t *testing.T) { t.Fatal(\"broken at the base\") }\n"}

var goOldChanged = map[string]string{"calc/calc_test.go": "package calc\n\nimport \"testing\"\n\n" +
	"func TestAdd(t *testing.T) {\n\tif Add(2, 3) != 5 || Sub(5, 3) != 2 {\n\t\tt.Fatal(\"Add or Sub is wrong\")\n\t}\n}\n\n" +
	"func TestSub(t *testing.T) {\n\tt.Run(\"zero\", func(t *testing.T) {\n\t\tif Sub(5, 0) != 5 {\n\t\t\tt.Fatal(\"Sub(5, 0)\")\n\t\t}\n\t})\n}\n"}

func forms() []form {
	return []form{
		{name: format.TestsGoJSON, needs: "go", command: "go test -json ./...",
			base: []string{"go/base"}, branch: []string{"go/branch", "go/subtest"},
			failed: []string{goCalc + "TestSub/five", goCalc + "TestSubMore"}, passed: []string{goMul + "TestMul"},
			notRun: []string{goNeg + "TestNeg"}, broken: goBroken, oldChanged: goOldChanged, oldName: goCalc + "TestAdd", news: 3},
		{name: format.TestsTAP, needs: "node", command: "node --test --test-reporter=tap",
			base: []string{"tap/base"}, branch: []string{"tap/branch"},
			failed: []string{"sub > five", "sub more"}, passed: []string{"mul works"}, notRun: []string{"neg works"},
			broken: map[string]string{"test/broken.test.mjs": "import { test } from 'node:test';\n\n" +
				"test('broken at the base', () => { throw new Error('broken'); });\n"},
			oldChanged: map[string]string{"test/calc.test.mjs": "import { test } from 'node:test';\nimport assert from 'node:assert';\n" +
				"import { add, sub } from '../src/calc.mjs';\n\ntest('add works', () => {\n  assert.strictEqual(add(2, 3), 5);\n" +
				"  assert.strictEqual(sub(5, 3), 2);\n});\n\ntest('sub', async (t) => {\n  await t.test('zero', () => {\n" +
				"    assert.strictEqual(sub(5, 0), 5);\n  });\n});\n"},
			oldName: "add works", news: 3},
		// JUnit lists a parent test beside its subtests and cannot say it failed only by them, so its fixture adds no
		// subtest under an old test (this package's known limits).
		{name: format.TestsJUnit, needs: "go", command: "go test -json ./... | go run ./tojunit",
			base: []string{"go/base", "junit"}, branch: []string{"go/branch"},
			failed: []string{goCalc + "TestSubMore"}, passed: []string{goMul + "TestMul"}, notRun: []string{goNeg + "TestNeg"},
			broken: goBroken, oldChanged: goOldChanged, oldName: goCalc + "TestAdd", news: 3},
	}
}

// needs skips a test when the form's program is not on the PATH.
func (f form) need(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath(f.needs); err != nil {
		t.Skipf("%s is not on the PATH: the %s fixture's rung runs it", f.needs, f.name)
	}
}

// repo builds a fixture repository: the form's base folders and baseFiles committed on main (the base), then a task
// branch holding the form's branch folders and branchFiles ("" deletes a file), one commit ahead. It gives the
// checkout's top folder and the base commit.
func (f form) repo(t *testing.T, baseFiles, branchFiles map[string]string, branch ...string) (root, base string) {
	t.Helper()
	root, err := filepath.EvalSymlinks(t.TempDir())
	must(t, err)
	root = filepath.Join(root, "project")
	must(t, os.MkdirAll(root, 0o755))
	git(t, root, "init", "-q")
	for _, d := range f.base {
		layer(t, root, d)
	}
	write(t, root, baseFiles)
	git(t, root, "add", "-A")
	git(t, root, "commit", "-q", "-m", "base")
	base = git(t, root, "rev-parse", "HEAD")
	git(t, root, "checkout", "-q", "-b", "task")
	if branch == nil {
		branch = f.branch
	}
	for _, d := range branch {
		layer(t, root, d)
	}
	write(t, root, branchFiles)
	git(t, root, "add", "-A")
	git(t, root, "commit", "-q", "-m", "work")
	return root, base
}

// headRun runs the rung at HEAD, as the runner does, and reads its tests: it must be green.
func headRun(t *testing.T, root, command, tests string) testout.Report {
	t.Helper()
	r := rungjob.Run(context.Background(), rungjob.Spec{Command: command, Dir: root, Timeout: 5 * time.Minute})
	if !r.Green() {
		t.Fatalf("the HEAD run of %q is not green (exit %d, %s):\n%s", command, r.Exit, r.Reason, r.Out)
	}
	rep, ok := testout.Read(tests, r.Stdout, root)
	if !ok || len(rep.Tests) == 0 {
		t.Fatalf("the HEAD run named no tests:\n%s", r.Out)
	}
	return rep
}

// rungOf is a fixture's rung.
func rungOf(command, tests, setup string, timeout int64) format.Rung {
	r := format.Rung{Rung: 3, Kind: "command", Command: &command, Tests: &tests, Required: true}
	if setup != "" {
		r.BaseSetup = &setup
	}
	if timeout > 0 {
		r.TimeoutS = &timeout
	}
	return r
}

// checked runs the check and holds what every result must hold: count is the sum of the three groups; each group's
// names are at most MaxNames and no more than its count; new_tests encodes in a ladder result held to bonsai.ladder/1;
// the lines are ASCII; no base worktree is left in the temporary folder, and git's worktrees are as before.
func checked(t *testing.T, tmp string, in Input) Result {
	t.Helper()
	before := ""
	if in.Root != "" {
		before = git(t, in.Root, "worktree", "list", "--porcelain")
	}
	r := Check(in)
	nt := r.NewTests
	if nt.Count != nt.FailedOnBase+nt.PassedOnBase+nt.NotRun {
		t.Errorf("count %d is not %d + %d + %d", nt.Count, nt.FailedOnBase, nt.PassedOnBase, nt.NotRun)
	}
	for _, g := range []struct {
		names []string
		n     int64
	}{{r.Failed, nt.FailedOnBase}, {r.Passed, nt.PassedOnBase}, {r.NotRun, nt.NotRun}} {
		if len(g.names) > MaxNames || int64(len(g.names)) > g.n {
			t.Errorf("%d names for a count of %d", len(g.names), g.n)
		}
	}
	held(t, nt)
	for _, l := range r.Lines() {
		if ascii(l) != l {
			t.Errorf("a line is not ASCII: %q", l)
		}
	}
	if left, _ := filepath.Glob(filepath.Join(tmp, tempPrefix+"*")); len(left) > 0 {
		t.Errorf("left in the temporary folder: %q", left)
	}
	if in.Root != "" {
		if after := git(t, in.Root, "worktree", "list", "--porcelain"); after != before {
			t.Errorf("git's worktrees changed:\n%s\nwas\n%s", after, before)
		}
	}
	return r
}

// held encodes new_tests in a ladder result, held to bonsai.ladder/1, and reads it back.
func held(t *testing.T, nt format.LadderNewTests) {
	t.Helper()
	task, cmd := "T-0901", "go test -json ./..."
	l := &format.Ladder{Task: &task, Workspace: "ws-7kq2m4xw5r3t6y2u7p4a5c3e2b", Mode: "local",
		Started: "2026-10-13T01:00:00.000Z", Finished: "2026-10-13T01:02:00.000Z",
		Git: format.LadderGit{SHA: strings.Repeat("a", 40), Branch: "main"}, Requested: []int64{3}, Green: true,
		Rungs: []format.LadderRung{{Rung: 3, Name: "tests", Kind: "command", Required: true, Command: &cmd, Status: "green",
			Ratchet: &format.LadderRatchet{Name: "go_tests", Was: "5", Now: "8", OK: true, NewTests: &nt}}},
		Skipped: []format.LadderSkip{}}
	raw, err := l.Encode()
	if err != nil {
		t.Fatalf("new_tests %+v is not held to bonsai.ladder/1: %v", nt, err)
	}
	back, err := format.ReadLadder(raw)
	if err != nil {
		t.Fatal(err)
	}
	got := *back.Rungs[0].Ratchet.NewTests
	got.Extra = nil
	if !reflect.DeepEqual(got, nt) {
		t.Errorf("read back %+v, wrote %+v", got, nt)
	}
}

// same compares two new_tests.
func same(a, b format.LadderNewTests) bool { return reflect.DeepEqual(a, b) }

// sameNames compares names without their order (node runs its files side by side).
func sameNames(t *testing.T, what string, got, want []string) {
	t.Helper()
	g, w := append([]string(nil), got...), append([]string(nil), want...)
	sort.Strings(g)
	sort.Strings(w)
	if len(g) == 0 && len(w) == 0 {
		return
	}
	if !reflect.DeepEqual(g, w) {
		t.Errorf("%s: %q, want %q", what, g, w)
	}
}

// whyHas checks the reason every test was not run.
func whyHas(t *testing.T, r Result, parts ...string) {
	t.Helper()
	for _, p := range parts {
		if !strings.Contains(r.Why, p) {
			t.Errorf("the reason lacks %q: %q", p, r.Why)
		}
	}
}

// sleepFor is a command that waits about n seconds, in sh and in cmd.
func sleepFor(n int) string {
	if runtime.GOOS == "windows" {
		return "ping -n " + strconv.Itoa(n+1) + " 127.0.0.1 >NUL"
	}
	return "sleep " + strconv.Itoa(n)
}

const rise = 7 // the ratchet's rise every fixture passes: the count when the new tests cannot be named

// stopOnceStarted stops a check once its base_setup has written started.txt in the base worktree (a minute at most).
func stopOnceStarted(tmp string, stop func()) {
	defer stop()
	for end := time.Now().Add(time.Minute); time.Now().Before(end); time.Sleep(20 * time.Millisecond) {
		if m, _ := filepath.Glob(filepath.Join(tmp, tempPrefix+"*", treeName, "started.txt")); len(m) > 0 {
			return
		}
	}
}

func TestForms(t *testing.T) {
	for _, f := range forms() {
		t.Run(f.name, func(t *testing.T) {
			f.need(t)
			tmp := isolate(t)
			root, base := f.repo(t, nil, nil)
			head := headRun(t, root, f.command, f.name)
			in := Input{Root: root, Rung: rungOf(f.command, f.name, "", 0), Base: base, Head: head, Rise: rise}

			t.Run("a new test failing at the base, one passing there, one that cannot load there", func(t *testing.T) {
				r := checked(t, tmp, in)
				n := int64(len(f.failed) + len(f.passed) + len(f.notRun))
				want := format.LadderNewTests{Base: base, Count: n, FailedOnBase: int64(len(f.failed)),
					PassedOnBase: int64(len(f.passed)), NotRun: int64(len(f.notRun))}
				if !r.Named || r.Why != "" || !same(r.NewTests, want) {
					t.Fatalf("got %+v (named %v, why %q), want %+v\n%s", r.NewTests, r.Named, r.Why, want, strings.Join(r.Lines(), "\n"))
				}
				t.Log("\n" + strings.Join(r.Lines(), "\n"))
				sameNames(t, "failed on the base", r.Failed, f.failed)
				sameNames(t, "passed on the base", r.Passed, f.passed)
				sameNames(t, "not run", r.NotRun, f.notRun)
				if s := r.Summary(); !strings.Contains(s, "run on the base "+base[:12]) {
					t.Errorf("summary %q", s)
				}
			})

			t.Run("a broken base_setup", func(t *testing.T) {
				in := in
				in.Rung = rungOf(f.command, f.name, "exit 3", 0)
				r := checked(t, tmp, in)
				if r.Named || !same(r.NewTests, format.LadderNewTests{Base: base, Count: rise, NotRun: rise}) {
					t.Fatalf("got %+v, named %v", r.NewTests, r.Named)
				}
				whyHas(t, r, "base_setup failed (exit 3)")
			})

			t.Run("no base", func(t *testing.T) {
				in := in
				in.Base = ""
				r := checked(t, tmp, in)
				if !same(r.NewTests, format.LadderNewTests{Base: "", Count: rise, NotRun: rise}) || r.Why != "no base" {
					t.Fatalf("got %+v, why %q", r.NewTests, r.Why)
				}
				if s := r.Summary(); s != "7 new tests (the count's rise; they could not be named): check not run: no base" {
					t.Errorf("summary %q", s)
				}
			})

			t.Run("an old test failing at the base", func(t *testing.T) {
				deleted := map[string]string{}
				for p := range f.broken {
					deleted[p] = ""
				}
				root, base := f.repo(t, f.broken, deleted)
				head := headRun(t, root, f.command, f.name)
				r := checked(t, tmp, Input{Root: root, Rung: in.Rung, Base: base, Head: head, Rise: rise})
				n := int64(len(f.failed) + len(f.passed) + len(f.notRun))
				if !r.Named || !same(r.NewTests, format.LadderNewTests{Base: base, Count: n, NotRun: n}) {
					t.Fatalf("got %+v, named %v, why %q", r.NewTests, r.Named, r.Why)
				}
				whyHas(t, r, "an old test failed at the base, before the branch's test files were copied in", "roken")
			})

			t.Run("an old test the branch changed failing at the base", func(t *testing.T) {
				branch := f.branch
				if f.name == format.TestsGoJSON {
					branch = []string{"go/branch"} // its own calc_test.go replaces the subtest's
				}
				root, base := f.repo(t, nil, f.oldChanged, branch...)
				head := headRun(t, root, f.command, f.name)
				r := checked(t, tmp, Input{Root: root, Rung: in.Rung, Base: base, Head: head, Rise: rise})
				n := int64(f.news)
				if !r.Named || !same(r.NewTests, format.LadderNewTests{Base: base, Count: n, NotRun: n}) {
					t.Fatalf("got %+v, named %v, why %q", r.NewTests, r.Named, r.Why)
				}
				whyHas(t, r, "an old test failed at the base with the branch's test files copied in: "+f.oldName)
			})

			t.Run("a timeout", func(t *testing.T) {
				in := in
				in.Rung = rungOf(f.command, f.name, sleepFor(30), 1)
				start := time.Now()
				r := checked(t, tmp, in)
				if r.Named || r.NewTests.NotRun != rise {
					t.Fatalf("got %+v, named %v", r.NewTests, r.Named)
				}
				whyHas(t, r, "base_setup timed out after 1 s")
				if took := time.Since(start); took > 20*time.Second {
					t.Errorf("the check took %s past a 1 s timeout", took)
				}
			})

			t.Run("a stop", func(t *testing.T) {
				in := in
				in.Rung = rungOf(f.command, f.name, "echo started > started.txt && "+sleepFor(30), 60)
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				in.Stop = ctx
				go stopOnceStarted(tmp, cancel)
				r := checked(t, tmp, in)
				if r.Named || r.NewTests.NotRun != rise {
					t.Fatalf("got %+v, named %v", r.NewTests, r.Named)
				}
				whyHas(t, r, "base_setup was stopped")
			})
		})
	}
}

// The test files cannot be told: a branch that changed no test file by the form's convention, and a JUnit report
// that names no file. No base worktree is made.
func TestFilesCannotBeTold(t *testing.T) {
	for _, f := range forms() {
		t.Run(f.name, func(t *testing.T) {
			f.need(t)
			tmp := isolate(t)
			code := map[string]string{"src/calc.mjs": "export const add = (a, b) => a + b;\nexport const sub = (a, b) => a - b;\n"}
			if f.needs == "go" {
				code = map[string]string{"calc/calc.go": "package calc\n\n// Add adds.\nfunc Add(a, b int) int { return a + b }\n\n" +
					"// Sub subtracts.\nfunc Sub(a, b int) int { return a - b }\n"}
			}
			root, base := f.repo(t, nil, code, []string{}...)
			head := headRun(t, root, f.command, f.name)
			r := checked(t, tmp, Input{Root: root, Rung: rungOf(f.command, f.name, "", 0), Base: base, Head: head, Rise: rise})
			if r.Named || !same(r.NewTests, format.LadderNewTests{Base: base, Count: rise, NotRun: rise}) {
				t.Fatalf("got %+v, named %v, why %q", r.NewTests, r.Named, r.Why)
			}
			whyHas(t, r, "the test files cannot be told: the branch changed no")
			if f.name != format.TestsJUnit {
				return
			}
			root, base = f.repo(t, nil, nil)
			command := "go test -json ./... | go run ./tojunit -nofile"
			head = headRun(t, root, command, f.name)
			r = checked(t, tmp, Input{Root: root, Rung: rungOf(command, f.name, "", 0), Base: base, Head: head, Rise: rise})
			if r.Named || r.NewTests.NotRun != rise {
				t.Fatalf("got %+v, named %v, why %q", r.NewTests, r.Named, r.Why)
			}
			whyHas(t, r, "the test files cannot be told: the JUnit report names no file")
		})
	}
}

// A new test that never ends at the base: the base run with the branch's test files times out, and every new test is
// "check not run" (node: its start is quick enough for a short timeout).
func TestSecondRunTimesOut(t *testing.T) {
	f := forms()[1]
	f.need(t)
	tmp := isolate(t)
	hang := map[string]string{"test/hang.test.mjs": "import { test } from 'node:test';\nimport { sub } from '../src/calc.mjs';\n\n" +
		"test('waits for sub', async () => {\n  while (sub(5, 3) !== 2) await new Promise((r) => setTimeout(r, 50));\n});\n"}
	root, base := f.repo(t, nil, hang)
	head := headRun(t, root, f.command, f.name)
	start := time.Now()
	r := checked(t, tmp, Input{Root: root, Rung: rungOf(f.command, f.name, "", 8), Base: base, Head: head, Rise: rise})
	if !r.Named || !same(r.NewTests, format.LadderNewTests{Base: base, Count: 5, NotRun: 5}) {
		t.Fatalf("got %+v, named %v, why %q", r.NewTests, r.Named, r.Why)
	}
	whyHas(t, r, "the base run with the branch's test files timed out after 8 s")
	if took := time.Since(start); took > 60*time.Second {
		t.Errorf("the check took %s past an 8 s timeout", took)
	}
}

// The base is not in the repository (a shallow clone, history rewritten): not run, and nothing made.
func TestBaseNotInClone(t *testing.T) {
	tmp := isolate(t)
	f := forms()[0]
	root, _ := f.repo(t, nil, nil)
	gone := strings.Repeat("ab", 20)
	r := checked(t, tmp, Input{Root: root, Rung: rungOf(f.command, f.name, "", 0), Base: gone, Rise: rise})
	if r.Named || !same(r.NewTests, format.LadderNewTests{Base: gone, Count: rise, NotRun: rise}) {
		t.Fatalf("got %+v", r.NewTests)
	}
	whyHas(t, r, "the base abababababab is not in this clone")
}

// A base worktree a killed check left is removed by the next check in the repository once its owner no longer runs; a
// running owner's is left; a folder already gone leaves only git's record, which is pruned.
func TestSweep(t *testing.T) {
	tmp := isolate(t)
	f := forms()[0]
	root, base := f.repo(t, nil, nil)
	before := git(t, root, "worktree", "list", "--porcelain")
	dead := exec.Command("git", "--version")
	must(t, dead.Run())
	stale := func(pid int) string {
		parent, err := os.MkdirTemp(tmp, tempPrefix)
		must(t, err)
		must(t, os.WriteFile(filepath.Join(parent, ownerName), []byte("pid "+strconv.Itoa(pid)+"\n"), 0o644))
		git(t, root, "worktree", "add", "--detach", "--quiet", filepath.Join(parent, treeName), base)
		return parent
	}
	killed := stale(dead.Process.Pid)
	running := stale(os.Getppid())
	vanished := stale(dead.Process.Pid)
	must(t, os.RemoveAll(vanished))

	r := Check(Input{Root: root, Rung: rungOf(f.command, f.name, "exit 3", 0), Base: base, Rise: rise})
	whyHas(t, r, "base_setup failed (exit 3)")
	if _, err := os.Stat(killed); !os.IsNotExist(err) {
		t.Errorf("the killed check's worktree is still there: %v", err)
	}
	if _, err := os.Stat(filepath.Join(running, treeName, "go.mod")); err != nil {
		t.Errorf("the running check's worktree was touched: %v", err)
	}
	notes := strings.Join(r.Notes, "\n")
	if !strings.Contains(notes, "that a killed check (pid "+strconv.Itoa(dead.Process.Pid)+") left") {
		t.Errorf("no note of the sweep: %q", notes)
	}
	list := git(t, root, "worktree", "list", "--porcelain")
	for _, p := range []string{killed, vanished} {
		if strings.Contains(list, filepath.ToSlash(p)) {
			t.Errorf("git still lists %s:\n%s", p, list)
		}
	}
	git(t, root, "worktree", "remove", "--force", filepath.Join(running, treeName))
	must(t, os.RemoveAll(running))
	if after := git(t, root, "worktree", "list", "--porcelain"); after != before {
		t.Errorf("git's worktrees:\n%s\nwas\n%s", after, before)
	}
}

// A stop at any moment of the check (making the worktree, either base run, the removal) leaves no worktree behind.
func TestStopAtAnyMoment(t *testing.T) {
	f := forms()[0]
	f.need(t)
	tmp := isolate(t)
	root, base := f.repo(t, nil, nil)
	head := headRun(t, root, f.command, f.name)
	ms := time.Millisecond
	for _, d := range []time.Duration{0, 5 * ms, 10 * ms, 20 * ms, 40 * ms, 60 * ms, 100 * ms, 150 * ms, 250 * ms, 400 * ms,
		700 * ms, time.Second, 2 * time.Second} {
		t.Run(d.String(), func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			time.AfterFunc(d, cancel)
			r := checked(t, tmp, Input{Root: root, Rung: rungOf(f.command, f.name, "", 0), Base: base, Head: head, Rise: rise, Stop: ctx})
			if r.Why != "" && !strings.Contains(r.Why, "stopped") {
				t.Errorf("why %q", r.Why)
			}
			for _, n := range r.Notes {
				t.Logf("note: %s", n)
			}
		})
	}
}
