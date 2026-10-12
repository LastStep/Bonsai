package ladder

// The runner on fixture projects (plan-5 5.4.2's proof): each requested set, a required red skipping the rest, the skip
// and mark rules local and under --ci, each test form, captures and the ratchet's count, a worktree's result landing in
// main's .bonsai/local/ladder/, an unverified main, the result and the climb document held to their schemas, the
// fingerprint record, the cleaner. Every project is a real git repository in a temporary folder with a Bonsai home of
// its own; every rung's command reads the same in sh and cmd (echo, exit, cat or type of a file in the checkout).

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/testpack"
	"github.com/LastStep/Bonsai/internal/workspace"
)

const testID = "ws-aaaaaaaaaaaaaaaaaaaaaaaaaa"

// rung is one rung of a fixture's bonsai.yaml.
type rung struct {
	n                     int
	name, kind, command   string
	required              bool
	timeout               int    // 0 for null
	ratchet, capture, tst string // the YAML values as written ("" for null)
}

// yamlOf writes a fixture's bonsai.yaml: floor, ratchets and ci_marked_tests as YAML values (a mapping with content in
// block style: "\n  name: 5"; bonsai.yaml takes no flow mapping but {}), and the rungs (a capture the same way, deeper).
func yamlOf(floor, ratchets, marked string, rungs ...rung) string {
	var b strings.Builder
	b.WriteString("format: bonsai.workspace/1\nid: " + testID + "\nname: climb-test\npacks: []\ndocuments:\n  task: work/tasks\n" +
		"  run: work/runs\n  answers: work/answers.md\n  memory: work/memory\n  protocols: work/protocols\nprotected: []\n" +
		"person_only: []\nnever_edit: []\nladder_floor: " + floor + "\nladder:\n")
	null := func(s string) string {
		if s == "" {
			return "null"
		}
		return s
	}
	for _, r := range rungs {
		kind := r.kind
		if kind == "" {
			kind = "command"
		}
		cmd := "null"
		if r.command != "" {
			cmd = strconv.Quote(r.command)
		}
		timeout := "null"
		if r.timeout > 0 {
			timeout = strconv.Itoa(r.timeout)
		}
		b.WriteString("  - rung: " + strconv.Itoa(r.n) + "\n    name: " + strconv.Quote(r.name) + "\n    kind: " + kind +
			"\n    command: " + cmd + "\n    required: " + strconv.FormatBool(r.required) + "\n    timeout_s: " + timeout +
			"\n    ratchet: " + null(r.ratchet) + "\n    capture: " + null(r.capture) + "\n    means: \"A rung of the test.\"\n" +
			"    tests: " + null(r.tst) + "\n    base_setup: null\n")
	}
	if len(rungs) == 0 {
		b.WriteString("  []\n")
	}
	b.WriteString("ratchets: " + ratchets + "\nci_marked_tests: " + marked + "\ngenerated:\n")
	for _, k := range []string{"log", "asks", "ladder", "run", "sessions"} {
		b.WriteString("  " + k + ":\n    keep_days: null\n    keep_newest: null\n")
	}
	return strings.Replace(b.String(), "ladder:\n  []\n", "ladder: []\n", 1)
}

// taskText is a format-1 task file: its status and its bonsai.ladder ("" for none).
func taskText(id, status, ladder string) string {
	labels := "{}"
	if ladder != "" {
		labels = "\n  bonsai.ladder: " + ladder
	}
	return "---\nformat: bonsai.task/1\nid: " + id + "\ntitle: A task of the test\nstatus: " + status + "\nlane: null\ndone_when:\n" +
		"  - \"It climbs\"\ndepends_on: []\nblocked_by:\ncreated: 2026-10-13\nstarted:\nfinished:\nlabels: " + labels + "\n---\n\n# A task\n"
}

// show is a file's contents command: cat on Linux and macOS, type in cmd.
func show(file string) string {
	if runtime.GOOS == "windows" {
		return "type " + file
	}
	return "cat " + file
}

// fx is a fixture project.
type fx struct {
	t         *testing.T
	tmp, root string
	home      string
	env       map[string]string
	out       bytes.Buffer
	notice    bytes.Buffer
	lockPoll  time.Duration
	lockGrace time.Duration
}

// project makes a fixture: a git repository holding bonsai.yaml, .bonsai/.gitignore, the tasks and files given, one
// commit on main, and this machine's record of it in a home of its own (so its main checkout is verified).
func project(t *testing.T, yaml string, files map[string]string) *fx {
	t.Helper()
	tmp, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	f := &fx{t: t, tmp: tmp, env: map[string]string{}, lockPoll: 20 * time.Millisecond, lockGrace: 300 * time.Millisecond}
	f.home = testpack.Isolate(t, tmp)
	f.root = testpack.Project(t, tmp, "project")
	f.write("bonsai.yaml", yaml)
	f.write(workspace.GitignoreFile, workspace.GitignoreText)
	f.write("work/runs/R-2026-10-13-T-0901.md", "# R-2026-10-13-T-0901\n\nWhat changed.\n")
	for rel, body := range files {
		f.write(rel, body)
	}
	f.commit("the fixture")
	if _, err := workspace.RecordCheckout(f.home, f.root, testID, time.Now()); err != nil {
		t.Fatal(err)
	}
	return f
}

func (f *fx) write(rel, body string) {
	f.t.Helper()
	p := filepath.Join(f.root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		f.t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		f.t.Fatal(err)
	}
}

func (f *fx) commit(msg string) {
	f.t.Helper()
	testpack.Git(f.t, f.root, "add", "-A")
	testpack.Git(f.t, f.root, "commit", "-q", "--allow-empty", "-m", msg)
}

func (f *fx) head() string { return testpack.Git(f.t, f.root, "rev-parse", "HEAD") }

// climb runs the ladder in dir (the project's root when "") and holds its document to bonsai.climb/1.
func (f *fx) climb(dir string, o Options) *Answer {
	f.t.Helper()
	if dir == "" {
		dir = f.root
	}
	o.Dir = dir
	o.Getenv = func(k string) string { return f.env[k] }
	o.Out, o.Notice = &f.out, &f.notice
	o.LockPoll, o.LockGrace = f.lockPoll, f.lockGrace
	a := Run(o)
	if _, err := a.Doc.Encode(); err != nil {
		f.t.Fatalf("the climb document does not fit bonsai.climb/1: %v", err)
	}
	asciiOnly(f.t, f.out.String())
	return a
}

// result reads a result the climb wrote in the project's main checkout, held to bonsai.ladder/1 as a writer writes it.
func (f *fx) result(name string) *format.Ladder {
	f.t.Helper()
	raw, err := os.ReadFile(filepath.Join(f.root, ".bonsai", "local", "ladder", name+".json"))
	if err != nil {
		f.t.Fatalf("the result: %v", err)
	}
	fits(f.t, "ladder", raw)
	l, err := format.ReadLadder(raw)
	if err != nil {
		f.t.Fatal(err)
	}
	return l
}

// fits holds a document to its format's schema as a writer writes it: every field, in order.
func fits(t *testing.T, name string, raw []byte) schema.Object {
	t.Helper()
	doc, err := schema.Decode(raw)
	if err != nil {
		t.Fatalf("not JSON: %v\n%s", err, raw)
	}
	s := format.MustLookup(name).Schema()
	if msgs := append(schema.Validate(s, doc), schema.CheckOrder(s, doc)...); len(msgs) > 0 {
		t.Fatalf("not bonsai.%s as written: %v\n%s", name, msgs, raw)
	}
	return doc.(schema.Object)
}

func asciiOnly(t *testing.T, s string) {
	t.Helper()
	for i := 0; i < len(s); i++ {
		if s[i] > 0x7e || (s[i] < 0x20 && s[i] != '\n') {
			t.Fatalf("a byte that is not printable ASCII (%#x) in:\n%s", s[i], s)
		}
	}
}

func statuses(l *format.Ladder) string {
	var out []string
	for _, r := range l.Rungs {
		out = append(out, strconv.FormatInt(r.Rung, 10)+":"+r.Status)
	}
	return strings.Join(out, " ")
}

// The requested set (note 2): the floor alone; the floor plus the task's bonsai.ladder, sorted, each once; --ci with no
// task every rung. The result is the task's, local, at HEAD, clean, green; exit 0; the human lines name each rung.
func TestRequestedSets(t *testing.T) {
	rungs := []rung{{n: 1, name: "one", command: "echo one", required: true}, {n: 2, name: "two", command: "echo two", required: true},
		{n: 3, name: "three", command: "echo three", required: true}}
	f := project(t, yamlOf("[1]", "{}", "[]", rungs...), map[string]string{
		"work/tasks/T-0901-a.md": taskText("T-0901", "running", ""),
		"work/tasks/T-0902-b.md": taskText("T-0902", "todo", "[3, 1]"),
	})
	a := f.climb("", Options{Task: "T-0901"})
	if a.Exit != 0 || a.Error != nil {
		t.Fatalf("exit %d, %v\n%s", a.Exit, a.Error, f.out.String())
	}
	l := f.result("T-0901")
	if statuses(l) != "1:green" || !l.Green || l.Mode != "local" || *l.Task != "T-0901" || l.Workspace != testID ||
		l.Git.SHA != f.head() || l.Git.Dirty || l.Git.Branch != "main" || l.Proof != nil || len(l.Skipped) != 0 {
		t.Errorf("the floor alone: %s %+v", statuses(l), l.Git)
	}
	if fmt := l.Requested; len(fmt) != 1 || fmt[0] != 1 {
		t.Errorf("requested %v", l.Requested)
	}
	if !strings.Contains(f.out.String(), "GREEN: 1 of 1 rungs green") || !strings.Contains(f.out.String(), "result: .bonsai/local/ladder/T-0901.json") {
		t.Errorf("the human lines:\n%s", f.out.String())
	}
	a = f.climb("", Options{Task: "T-0902"})
	if l := f.result("T-0902"); a.Exit != 0 || statuses(l) != "1:green 3:green" || len(l.Requested) != 2 {
		t.Errorf("the floor plus the task's: exit %d, %s, %v", a.Exit, statuses(l), l.Requested)
	}
	a = f.climb("", Options{CI: true})
	if l := f.result("ci"); a.Exit != 0 || statuses(l) != "1:green 2:green 3:green" || l.Mode != "ci" || l.Task != nil ||
		l.Proof == nil || *l.Proof != ProofSentence || l.Leftovers != nil {
		t.Errorf("--ci with no task: exit %d, %s, mode %s, proof %v, leftovers %v", a.Exit, statuses(l), l.Mode, l.Proof, l.Leftovers)
	}
	if a.Doc.Result == nil || *a.Doc.Result != ".bonsai/local/ladder/ci.json" || *a.Doc.Mode != "ci" || a.Doc.Task != nil {
		t.Errorf("the --ci climb document: %+v", a.Doc)
	}
	// BONSAI_TASK names the task in the session's own project; in another project it does not count.
	f.env[workspace.TaskEnv] = "T-0902"
	if a := f.climb("", Options{}); a.Exit != 0 || *a.Doc.Task != "T-0902" {
		t.Errorf("BONSAI_TASK: exit %d, %v", a.Exit, a.Error)
	}
	f.env[ProjectEnv] = f.tmp
	if a := f.climb("", Options{}); a.Exit != ExitInput || a.Error == nil || a.Error.Code != "no-task" {
		t.Errorf("BONSAI_TASK of another project counted: exit %d", a.Exit)
	}
}

// A required rung that is not green skips the rest; one that is not required lets them run; any red makes the climb
// red (exit 1), and next names the first red rung's fix and the climb again.
func TestRequiredRed(t *testing.T) {
	rungs := []rung{{n: 1, name: "soft", command: "exit 3", required: false}, {n: 2, name: "hard", command: "echo out && exit 4", required: true},
		{n: 3, name: "after", command: "echo after", required: true}}
	f := project(t, yamlOf("[1, 2, 3]", "{}", "[]", rungs...), map[string]string{"work/tasks/T-0901-a.md": taskText("T-0901", "running", "")})
	a := f.climb("", Options{Task: "T-0901"})
	l := f.result("T-0901")
	if a.Exit != ExitRed || l.Green || statuses(l) != "1:red 2:red 3:skipped" {
		t.Fatalf("exit %d, %s", a.Exit, statuses(l))
	}
	if r := l.Rungs[2]; r.Reason == nil || *r.Reason != "an earlier required rung failed (rung 2)" || r.DurationMS != nil {
		t.Errorf("the skipped rung: %+v", r)
	}
	if r := l.Rungs[1]; r.Reason == nil || !strings.Contains(*r.Reason, "exit 4") || !strings.Contains(*r.Reason, "| out") {
		t.Errorf("the red rung's reason: %v", r.Reason)
	}
	want := "fix the soft rung (rung 1): its reason shows what failed; run: exit 3 ; then run: bonsai ladder --task T-0901 --json"
	if a.Doc.Next == nil || a.Doc.Next.Do != want || a.Doc.Next.Who != "agent" || a.Doc.Error != nil {
		t.Errorf("next: %+v", a.Doc.Next)
	}
	if !strings.Contains(f.out.String(), "RED: 0 of 3 rungs green") || !strings.Contains(f.out.String(), "next: fix the soft rung") {
		t.Errorf("the human lines:\n%s", f.out.String())
	}
}

// A rung past its limit is ended red, its reason saying so, and next names its command and the person's longer limit.
func TestTimeout(t *testing.T) {
	sleep := "sleep 5"
	if runtime.GOOS == "windows" {
		sleep = "ping -n 6 127.0.0.1 >NUL"
	}
	f := project(t, yamlOf("[1]", "{}", "[]", rung{n: 1, name: "slow", command: sleep, required: true, timeout: 1}),
		map[string]string{"work/tasks/T-0901-a.md": taskText("T-0901", "running", "")})
	start := time.Now()
	a := f.climb("", Options{Task: "T-0901"})
	l := f.result("T-0901")
	if a.Exit != ExitRed || statuses(l) != "1:red" || l.Rungs[0].Reason == nil || !strings.Contains(*l.Rungs[0].Reason, "timed out after 1 s") {
		t.Fatalf("exit %d, %s, %v", a.Exit, statuses(l), l.Rungs[0].Reason)
	}
	if time.Since(start) > 15*time.Second {
		t.Errorf("the climb took %v past a 1 s limit", time.Since(start))
	}
	if a.Doc.Next == nil || !strings.Contains(a.Doc.Next.Do, "it ran past its limit of 1 s") {
		t.Errorf("next: %+v", a.Doc.Next)
	}
}

// goJSON is go test -json's output for a package with a pass, a failure in a subtest, a marked skip and an unmarked
// skip, and a package that does not build.
const goJSON = `{"Action":"start","Package":"example.com/demo/good"}
{"Action":"run","Package":"example.com/demo/good","Test":"TestPass"}
{"Action":"output","Package":"example.com/demo/good","Test":"TestPass","Output":"=== RUN   TestPass\n"}
{"Action":"pass","Package":"example.com/demo/good","Test":"TestPass","Elapsed":0}
{"Action":"run","Package":"example.com/demo/good","Test":"TestMarked"}
{"Action":"output","Package":"example.com/demo/good","Test":"TestMarked","Output":"    a_test.go:7: real-checkouts-only: needs a checkout\n"}
{"Action":"skip","Package":"example.com/demo/good","Test":"TestMarked","Elapsed":0}
{"Action":"run","Package":"example.com/demo/good","Test":"TestPlainSkip"}
{"Action":"output","Package":"example.com/demo/good","Test":"TestPlainSkip","Output":"    a_test.go:9: not on this system\n"}
{"Action":"skip","Package":"example.com/demo/good","Test":"TestPlainSkip","Elapsed":0}
{"Action":"pass","Package":"example.com/demo/good","Elapsed":0.004}
`

// The skip and mark rules (note 4), locally and under --ci, on go test -json's output.
func TestSkipsAndMarks(t *testing.T) {
	r := rung{n: 1, name: "tests", command: show("out.json"), required: true, ratchet: "go_tests", tst: "go-test-json"}
	files := map[string]string{"out.json": goJSON, "work/tasks/T-0901-a.md": taskText("T-0901", "running", "")}
	f := project(t, yamlOf("[1]", "\n  go_tests: 5", "[]", r), files)
	// Locally a marked skip is red; an unmarked one passes.
	a := f.climb("", Options{Task: "T-0901"})
	l := f.result("T-0901")
	reason := ""
	if l.Rungs[0].Reason != nil {
		reason = *l.Rungs[0].Reason
	}
	if a.Exit != ExitRed || !strings.Contains(reason, `"example.com/demo/good.TestMarked" skipped (real-checkouts-only: needs a checkout): it is marked real-checkouts-only`) ||
		strings.Contains(reason, "TestPlainSkip") {
		t.Fatalf("local: exit %d, reason %q", a.Exit, reason)
	}
	tests := l.Rungs[0].Tests
	if s, _ := tests.Get("seen"); schema.Show(s) != "3" || len(l.Skipped) != 0 {
		t.Errorf("tests %s, skipped %v", schema.Show(tests), l.Skipped)
	}
	if want := `[{"name":"example.com/demo/good.TestMarked","mark":"real-checkouts-only","reason":"real-checkouts-only: needs a checkout"},` +
		`{"name":"example.com/demo/good.TestPlainSkip","mark":null,"reason":"not on this system"}]`; showOf(tests, "skipped") != want {
		t.Errorf("tests.skipped %s", showOf(tests, "skipped"))
	}
	// The count: the tests named as passed (1), the floor 5: written, not judged (5.4.4 judges it).
	if rt := l.Rungs[0].Ratchet; rt == nil || rt.Name != "go_tests" || rt.Was != "5" || rt.Now != "1" || rt.OK || rt.NewTests != nil {
		t.Errorf("the ratchet: %+v", rt)
	}
	// Under --ci a marked skip passes only when ci_marked_tests names it, and every other skip is red.
	a = f.climb("", Options{CI: true})
	l = f.result("ci")
	if a.Exit != ExitRed || len(l.Skipped) != 2 || l.Skipped[0].Accepted || l.Skipped[1].Accepted {
		t.Fatalf("--ci, none named: exit %d, %+v", a.Exit, l.Skipped)
	}
	f.write("bonsai.yaml", yamlOf("[1]", "\n  go_tests: 5", `["example.com/demo/good.TestMarked"]`, r))
	f.commit("name the marked test")
	a = f.climb("", Options{CI: true})
	l = f.result("ci")
	reason = ""
	if l.Rungs[0].Reason != nil {
		reason = *l.Rungs[0].Reason
	}
	if a.Exit != ExitRed || !l.Skipped[0].Accepted || l.Skipped[1].Accepted || !strings.Contains(reason, "TestPlainSkip\" skipped without a mark") ||
		strings.Contains(reason, "TestMarked") {
		t.Fatalf("--ci, the marked one named: exit %d, %+v, %q", a.Exit, l.Skipped, reason)
	}
	// Under --ci the accepted skip counts toward the ratchet: 1 passed and 1 accepted.
	if rt := l.Rungs[0].Ratchet; rt == nil || rt.Now != "2" {
		t.Errorf("the --ci count: %+v", rt)
	}
	// Locally the marked test is red, named in ci_marked_tests or not.
	a = f.climb("", Options{Task: "T-0901"})
	if a.Exit != ExitRed {
		t.Errorf("local with the marked test named: exit %d", a.Exit)
	}
	// Unmarked, but named in ci_marked_tests: still red locally. Unmarked and named nowhere: green.
	f.write("out.json", strings.ReplaceAll(goJSON, "real-checkouts-only: needs a checkout", "needs a checkout"))
	f.commit("no mark")
	if a := f.climb("", Options{Task: "T-0901"}); a.Exit != ExitRed {
		t.Errorf("local, an unmarked skip named in ci_marked_tests: exit %d", a.Exit)
	}
	f.write("bonsai.yaml", yamlOf("[1]", "\n  go_tests: 1", "[]", r))
	f.commit("none named")
	if a := f.climb("", Options{Task: "T-0901"}); a.Exit != 0 {
		t.Errorf("local, unmarked skips only: exit %d\n%s", a.Exit, f.out.String())
	}
}

func showOf(o schema.Object, key string) string {
	v, _ := o.Get(key)
	return schema.Show(v)
}

// tapOut is node's TAP: a suite with a failing test, a passing subtest, a skip with a mark and a todo.
const tapOut = `TAP version 13
# Subtest: outer
    # Subtest: inner fails
    not ok 1 - inner fails
      ---
      duration_ms: 1
      type: 'test'
      location: '/work/project/a.test.mjs:4:3'
      failureType: 'testCodeFailure'
      error: 'want 1, got 2'
      ...
    # Subtest: inner passes
    ok 2 - inner passes
      ---
      duration_ms: 1
      type: 'test'
      ...
    1..2
not ok 1 - outer
  ---
  duration_ms: 3
  type: 'suite'
  failureType: 'subtestsFailed'
  error: '1 subtest failed'
  ...
# Subtest: later
ok 2 - later # TODO not yet
  ---
  duration_ms: 1
  type: 'test'
  ...
1..2
`

// junitOut is a JUnit report with a pass, a failure, an error and a skip.
const junitOut = `<?xml version="1.0" encoding="UTF-8"?>
<testsuites>
  <testsuite name="s">
    <testcase classname="pkg.A" name="passes"/>
    <testcase classname="pkg.A" name="fails"><failure message="want 1, got 2">detail</failure></testcase>
    <testcase classname="pkg.B" name="errs"><error message="boom"/></testcase>
    <testcase classname="pkg.B" name="skips"><skipped message="not here"/></testcase>
  </testsuite>
</testsuites>
`

// Each test form read through the runner (note 3): tests.failed holds each failing test with its file, error and the
// command that runs it alone (null where the form names none), failed_count, and the result's shape is today's.
func TestTestForms(t *testing.T) {
	rungs := []rung{
		{n: 1, name: "go", command: show("go.json") + " && exit 1", tst: "go-test-json"},
		{n: 2, name: "tap", command: show("out.tap") + " && exit 1", tst: "tap"},
		{n: 3, name: "junit", command: show("out.xml") + " && exit 1", tst: "junit"},
		{n: 4, name: "unknown", command: "echo x", tst: "nunit"},
	}
	goFail := goJSON + `{"Action":"run","Package":"example.com/demo/good","Test":"TestParent/sub_one"}
{"Action":"output","Package":"example.com/demo/good","Test":"TestParent/sub_one","Output":"    a_test.go:10: sub failed\n"}
{"Action":"fail","Package":"example.com/demo/good","Test":"TestParent/sub_one","Elapsed":0}
{"Action":"fail","Package":"example.com/demo/good","Test":"TestParent","Elapsed":0}
`
	goFail = strings.ReplaceAll(goFail, "real-checkouts-only: ", "")
	f := project(t, yamlOf("[1, 2, 3, 4]", "{}", "[]", rungs...), map[string]string{"go.json": goFail, "out.tap": tapOut,
		"out.xml": junitOut, "work/tasks/T-0901-a.md": taskText("T-0901", "running", ""),
		"go.mod": "module example.com/demo\n\ngo 1.25.0\n"})
	a := f.climb("", Options{Task: "T-0901"})
	l := f.result("T-0901")
	if a.Exit != ExitRed || statuses(l) != "1:red 2:red 3:red 4:green" {
		t.Fatalf("exit %d, %s\n%s", a.Exit, statuses(l), f.out.String())
	}
	goTests := l.Rungs[0].Tests
	if showOf(goTests, "failed_count") != "1" || !strings.Contains(showOf(goTests, "failed"), `"name":"example.com/demo/good.TestParent/sub_one"`) ||
		!strings.Contains(showOf(goTests, "failed"), `"file":"good/a_test.go"`) || !strings.Contains(showOf(goTests, "failed"), `"error":"sub failed"`) ||
		!strings.Contains(showOf(goTests, "failed"), "-run") {
		t.Errorf("go's tests: %s", schema.Show(goTests))
	}
	tapTests := l.Rungs[1].Tests
	if showOf(tapTests, "failed_count") != "1" || !strings.Contains(showOf(tapTests, "failed"), `"name":"outer > inner fails"`) ||
		showOf(tapTests, "todo") != `[{"name":"later","reason":"not yet"}]` {
		t.Errorf("tap's tests: %s", schema.Show(tapTests))
	}
	junit := l.Rungs[2].Tests
	if showOf(junit, "failed_count") != "2" || !strings.Contains(showOf(junit, "failed"), `"run":null`) ||
		!strings.Contains(showOf(junit, "skipped"), `"reason":"not here"`) {
		t.Errorf("junit's tests: %s", schema.Show(junit))
	}
	if l.Rungs[3].Tests != nil {
		t.Errorf("an unknown form's tests are not read: %s", schema.Show(l.Rungs[3].Tests))
	}
	if a.Doc.Next == nil || !strings.Contains(a.Doc.Next.Do, "fix the go rung (rung 1): the test example.com/demo/good.TestParent/sub_one fails; "+
		"it runs alone with: run: go test -count=1 -run") {
		t.Errorf("next: %+v", a.Doc.Next)
	}
}

// A package that does not build names no test and makes the rung red, its reason naming the package; next names go vet.
func TestBuildFailure(t *testing.T) {
	build := `{"ImportPath":"example.com/demo/bad [example.com/demo/bad.test]","Action":"build-output","Output":"# example.com/demo/bad\n"}
{"ImportPath":"example.com/demo/bad [example.com/demo/bad.test]","Action":"build-output","Output":"bad/b_test.go:5:28: undefined: undefined\n"}
{"ImportPath":"example.com/demo/bad [example.com/demo/bad.test]","Action":"build-fail"}
{"Action":"start","Package":"example.com/demo/bad"}
{"Action":"fail","Package":"example.com/demo/bad","Elapsed":0,"FailedBuild":"example.com/demo/bad [example.com/demo/bad.test]"}
`
	f := project(t, yamlOf("[1]", "{}", "[]", rung{n: 1, name: "go", command: show("go.json"), required: true, tst: "go-test-json"}),
		map[string]string{"go.json": build, "work/tasks/T-0901-a.md": taskText("T-0901", "running", "")})
	a := f.climb("", Options{Task: "T-0901"})
	l := f.result("T-0901")
	if a.Exit != ExitRed || statuses(l) != "1:red" || l.Rungs[0].Reason == nil ||
		!strings.Contains(*l.Rungs[0].Reason, "the package example.com/demo/bad did not build: bad/b_test.go:5:28: undefined: undefined") {
		t.Fatalf("exit %d, %s, %v", a.Exit, statuses(l), l.Rungs[0].Reason)
	}
	if showOf(l.Rungs[0].Tests, "seen") != "0" || a.Doc.Next == nil || !strings.Contains(a.Doc.Next.Do, "run: go vet example.com/demo/bad") {
		t.Errorf("tests %s, next %+v", schema.Show(l.Rungs[0].Tests), a.Doc.Next)
	}
}

// Captures: a pattern's first group, a number when it reads as one; a pattern that does not compile is null; the
// ratchet's count is the capture of its name over the passed tests.
func TestCaptures(t *testing.T) {
	r := rung{n: 1, name: "c", command: "echo passed 7 of 9 && echo version v1.2", required: true, ratchet: "found",
		capture: "\n      found: \"passed ([0-9]+) of\"\n      version: \"version (v[0-9.]+)\"\n      broken: \"(\"\n      whole: \"of [0-9]+\""}
	f := project(t, yamlOf("[1]", "\n  found: 3", "[]", r), map[string]string{"work/tasks/T-0901-a.md": taskText("T-0901", "running", "")})
	if a := f.climb("", Options{Task: "T-0901"}); a.Exit != 0 {
		t.Fatalf("exit %d, %+v\n%s", a.Exit, a.Error, f.out.String())
	}
	l := f.result("T-0901")
	if got := schema.Show(l.Rungs[0].Captures); got != `{"found":7,"version":"v1.2","broken":null,"whole":"of 9"}` {
		t.Errorf("captures %s", got)
	}
	if rt := l.Rungs[0].Ratchet; rt == nil || rt.Was != "3" || rt.Now != "7" || !rt.OK {
		t.Errorf("the ratchet %+v", rt)
	}
}

// A guard or ver-git rung ends in error until its piece wires into kinds.go: the floor holding one climbs red.
func TestKindsNotBuilt(t *testing.T) {
	f := project(t, yamlOf("[0, 5]", "{}", "[]", rung{n: 0, name: "guard", kind: "guard", required: true},
		rung{n: 5, name: "git", kind: "ver-git", required: false}), map[string]string{"work/tasks/T-0901-a.md": taskText("T-0901", "running", "")})
	a := f.climb("", Options{Task: "T-0901"})
	l := f.result("T-0901")
	if a.Exit != ExitRed || statuses(l) != "0:error 5:skipped" || l.Rungs[0].Command != nil || !strings.Contains(*l.Rungs[0].Reason, "not built yet") {
		t.Errorf("exit %d, %s, %+v", a.Exit, statuses(l), l.Rungs[0])
	}
}

// A climb in a linked worktree writes the main checkout's .bonsai/local/ladder/<task>.json, proving the worktree's
// HEAD and branch; its record goes to the main checkout's log.
func TestWorktreeResultLandsInMain(t *testing.T) {
	f := project(t, yamlOf("[1]", "{}", "[]", rung{n: 1, name: "one", command: "echo one", required: true}),
		map[string]string{"work/tasks/T-0901-a.md": taskText("T-0901", "running", "")})
	wt := filepath.Join(f.tmp, "wt")
	testpack.Git(t, f.root, "worktree", "add", "-q", "-b", "t0901", wt)
	testpack.Git(t, wt, "-c", "user.name=t", "-c", "user.email=t", "commit", "-q", "--allow-empty", "-m", "work")
	a := f.climb(wt, Options{Task: "T-0901"})
	if a.Exit != 0 {
		t.Fatalf("exit %d, %v\n%s", a.Exit, a.Error, f.out.String())
	}
	l := f.result("T-0901")
	if l.Git.SHA != testpack.Git(t, wt, "rev-parse", "HEAD") || l.Git.Branch != "t0901" || l.Git.Dirty {
		t.Errorf("the result's git: %+v", l.Git)
	}
	if _, err := os.Stat(filepath.Join(wt, ".bonsai", "local")); err == nil {
		t.Error("the worktree has a .bonsai/local/ of its own")
	}
	if a.Doc.Workspace == nil || a.Doc.Workspace.Root != filepath.ToSlash(f.root) {
		t.Errorf("the climb's workspace root: %+v", a.Doc.Workspace)
	}
	recs := ladderRecords(t, f.root)
	if len(recs) != 1 || recs[0].String("checkout") != "wt" || recs[0].String("branch") != "t0901" {
		t.Errorf("the ladder record in main's log: %v", recs)
	}
}

// ladderRecords reads the ladder records of a main checkout's log, each held to bonsai.log/1.
func ladderRecords(t *testing.T, main string) []schema.Object {
	t.Helper()
	var out []schema.Object
	files, _ := filepath.Glob(filepath.Join(main, ".bonsai", "local", "log", "*.ndjson"))
	for _, p := range files {
		raw, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		for _, line := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
			if line == "" {
				continue
			}
			doc := fits(t, "log", []byte(line))
			if doc.String("event") == "ladder" {
				out = append(out, doc)
			}
		}
	}
	return out
}

// The fingerprint (note 7): one ladder record per climb, its target the result's path, its text the sha256 of the
// file's bytes (the climb document's too), ok its green, task its task, kind its mode, category Ladder, and the
// runner's own path and SHA-256; in the session's file when CLAUDE_CODE_SESSION_ID names one.
func TestFingerprint(t *testing.T) {
	f := project(t, yamlOf("[1]", "{}", "[]", rung{n: 1, name: "one", command: "echo one", required: true}),
		map[string]string{"work/tasks/T-0901-a.md": taskText("T-0901", "running", "")})
	f.env[SessionEnv] = "sess-1"
	a := f.climb("", Options{Task: "T-0901"})
	raw, err := os.ReadFile(filepath.Join(f.root, ".bonsai", "local", "ladder", "T-0901.json"))
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256Hex(raw)
	if a.Doc.SHA256 == nil || *a.Doc.SHA256 != sum {
		t.Errorf("the document's sha256 %v, the file's %s", a.Doc.SHA256, sum)
	}
	recs := ladderRecords(t, f.root)
	if len(recs) != 1 {
		t.Fatalf("%d ladder records", len(recs))
	}
	r := recs[0]
	exe, _ := os.Executable()
	exeRaw, _ := os.ReadFile(exe)
	home, _ := os.UserHomeDir()
	if r.String("target") != ".bonsai/local/ladder/T-0901.json" || r.String("text") != "sha256:"+sum || showOf(r, "ok") != "true" ||
		r.String("task") != "T-0901" || r.String("kind") != "local" || r.String("category") != "Ladder" || r.String("session") != "sess-1" ||
		r.String("bonsai_sha256") != sha256Hex(exeRaw) || r.String("bonsai_path") != fromHomeForTest(exe, home) {
		t.Errorf("the record: %s", schema.Show(r))
	}
	if _, err := os.Stat(filepath.Join(f.root, ".bonsai", "local", "log", "s-sess-1.ndjson")); err != nil {
		t.Errorf("not in the session's file: %v", err)
	}
	// With no session, the day's file.
	delete(f.env, SessionEnv)
	f.climb("", Options{Task: "T-0901"})
	if files, _ := filepath.Glob(filepath.Join(f.root, ".bonsai", "local", "log", "w-*.ndjson")); len(files) != 1 {
		t.Errorf("the day file: %v", files)
	}
}

// fromHomeForTest writes a path from ~/ under home, worked out here on its own.
func fromHomeForTest(p, home string) string {
	p = filepath.ToSlash(filepath.Clean(p))
	h := strings.TrimSuffix(filepath.ToSlash(filepath.Clean(home)), "/") + "/"
	if home != "" && len(p) > len(h) && (p[:len(h)] == h || (runtime.GOOS == "windows" && strings.EqualFold(p[:len(h)], h))) {
		return "~/" + p[len(h):]
	}
	return p
}

// A main checkout that is not verified on this machine (no record of it here) writes no result and runs nothing:
// main-not-verified, exit 4, a person's step. --ci runs anyway, in the checkout's own .bonsai/.
func TestMainNotVerified(t *testing.T) {
	f := project(t, yamlOf("[1]", "{}", "[]", rung{n: 1, name: "one", command: "echo one > ran.txt", required: true}),
		map[string]string{"work/tasks/T-0901-a.md": taskText("T-0901", "running", "")})
	if err := os.RemoveAll(filepath.Join(f.home, "workspaces")); err != nil {
		t.Fatal(err)
	}
	a := f.climb("", Options{Task: "T-0901"})
	if a.Exit != ExitState || a.Error == nil || a.Error.Code != "main-not-verified" || a.Doc.Error.Next.Who != "person" ||
		!strings.Contains(a.Error.Next, "bonsai update") || !strings.Contains(a.Error.Next, "then run: bonsai ladder --task T-0901 --json") {
		t.Fatalf("exit %d, %+v", a.Exit, a.Error)
	}
	if _, err := os.Stat(filepath.Join(f.root, "ran.txt")); err == nil {
		t.Error("a rung ran")
	}
	if _, err := os.Stat(filepath.Join(f.root, ".bonsai", "local", "ladder")); err == nil {
		t.Error("a result was written")
	}
	a = f.climb("", Options{CI: true})
	if a.Exit != 0 || !strings.Contains(f.out.String(), "not verified") {
		t.Errorf("--ci: exit %d\n%s", a.Exit, f.out.String())
	}
	f.result("ci")
}

// The refusals before anything runs, each with its word, its exit code and a next step with the values filled in.
func TestRefusals(t *testing.T) {
	rungs := []rung{{n: 1, name: "one", command: "echo one > ran.txt", required: true}}
	cases := []struct {
		name  string
		yaml  string
		tasks map[string]string
		o     Options
		word  string
		exit  int
		next  string
	}{
		{"no task named, one running", yamlOf("[1]", "{}", "[]", rungs...), map[string]string{"work/tasks/T-0901-a.md": taskText("T-0901", "running", "")},
			Options{}, "no-task", ExitInput, "run: bonsai ladder --task T-0901 --json"},
		{"no task named, none running", yamlOf("[1]", "{}", "[]", rungs...), map[string]string{"work/tasks/T-0901-a.md": taskText("T-0901", "todo", "")},
			Options{}, "no-task", ExitInput, "run: bonsai status --active --json; then, with its id, run: bonsai ladder --task <id> --json"},
		{"the named task missing", yamlOf("[1]", "{}", "[]", rungs...), map[string]string{"work/tasks/T-0901-a.md": taskText("T-0901", "todo", "")},
			Options{Task: "T-0999"}, "no-task", ExitInput, "check the id"},
		{"the named task broken", yamlOf("[1]", "{}", "[]", rungs...), map[string]string{"work/tasks/T-0901-a.md": "---\nformat: bonsai.task/1\nid: [\n---\n"},
			Options{Task: "T-0901"}, "no-task", ExitInput, "find the line: run: bonsai check --json;"},
		{"--ci naming a missing task", yamlOf("[1]", "{}", "[]", rungs...), nil, Options{CI: true, Task: "T-0999"}, "no-task", ExitInput, "check the id"},
		{"a floor rung not defined", yamlOf("[1, 4]", "{}", "[]", rungs...), map[string]string{"work/tasks/T-0901-a.md": taskText("T-0901", "todo", "")},
			Options{Task: "T-0901"}, "rung-not-defined", ExitInput, "run: bonsai check --json"},
		{"a task rung not defined", yamlOf("[1]", "{}", "[]", rungs...), map[string]string{"work/tasks/T-0901-a.md": taskText("T-0901", "todo", "[9]")},
			Options{Task: "T-0901"}, "rung-not-defined", ExitInput, "run: bonsai check --json"},
		{"no ladder", yamlOf("[]", "{}", "[]"), map[string]string{"work/tasks/T-0901-a.md": taskText("T-0901", "todo", "")},
			Options{Task: "T-0901"}, "rung-not-defined", ExitInput, "bonsai check --json"},
		{"a task's bonsai.ladder that is not rung numbers", yamlOf("[1]", "{}", "[]", rungs...),
			map[string]string{"work/tasks/T-0901-a.md": taskText("T-0901", "todo", "[one]")},
			Options{Task: "T-0901"}, "rung-not-defined", ExitInput, "work/tasks/T-0901-a.md"},
		{"--root not a folder", yamlOf("[1]", "{}", "[]", rungs...), nil, Options{Root: "no-such-folder", Task: "T-0901"}, "bad-value", ExitInput,
			"run: bonsai ladder --root <folder> --task T-0901 --json"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := project(t, c.yaml, c.tasks)
			a := f.climb("", c.o)
			if a.Exit != c.exit || a.Error == nil || a.Error.Code != c.word || a.Doc.Error == nil || a.Doc.Error.Code != c.word {
				t.Fatalf("exit %d, %+v; want %d %s", a.Exit, a.Error, c.exit, c.word)
			}
			if !strings.Contains(a.Error.Next, c.next) {
				t.Errorf("next %q, want it to hold %q", a.Error.Next, c.next)
			}
			if a.Doc.Result != nil || a.Doc.Ladder != nil || a.Doc.Next != nil || len(a.Doc.Bless) != 0 || len(a.Doc.Cleaned) != 0 {
				t.Errorf("a refusal's document: %+v", a.Doc)
			}
			if _, err := os.Stat(filepath.Join(f.root, "ran.txt")); err == nil {
				t.Error("a rung ran")
			}
		})
	}
	// A folder in no checkout, and a checkout with no bonsai.yaml.
	tmp := t.TempDir()
	testpack.Isolate(t, tmp)
	if a := Run(Options{Dir: tmp, Task: "T-0901", Getenv: func(string) string { return "" }}); a.Exit != ExitState || a.Error.Code != "not-a-checkout" {
		t.Errorf("no checkout: %d %+v", a.Exit, a.Error)
	}
	plain := testpack.Project(t, tmp, "plain")
	if a := Run(Options{Dir: plain, Task: "T-0901", Getenv: func(string) string { return "" }}); a.Exit != ExitState || a.Error.Code != "not-linked" {
		t.Errorf("not linked: %d %+v", a.Exit, a.Error)
	}
}

// The cleaner's ladder kind runs after the result is written (note 9): a done task's old result goes, and is listed
// in cleaned; the result just written never does, even when its own rule would clean it.
func TestCleaning(t *testing.T) {
	yaml := strings.Replace(yamlOf("[1]", "{}", "[]", rung{n: 1, name: "one", command: "echo one", required: true}),
		"  ladder:\n    keep_days: null", "  ladder:\n    keep_days: 0", 1)
	f := project(t, yaml, map[string]string{"work/tasks/T-0901-a.md": taskText("T-0901", "done", ""),
		"work/tasks/T-0800-old.md": taskText("T-0800", "done", "")})
	old := &format.Ladder{Task: strPtr("T-0800"), Workspace: testID, Mode: "local", Started: "2026-01-01T00:00:00.000Z",
		Finished: "2026-01-01T00:01:00.000Z", Git: format.LadderGit{SHA: f.head(), Branch: "main"}, Requested: []int64{1}, Green: true,
		Rungs: []format.LadderRung{}, Skipped: []format.LadderSkip{}}
	raw, err := old.Encode()
	if err != nil {
		t.Fatal(err)
	}
	f.write(".bonsai/local/ladder/T-0800.json", string(raw))
	a := f.climb("", Options{Task: "T-0901"})
	if a.Exit != 0 || len(a.Doc.Cleaned) != 1 || a.Doc.Cleaned[0] != ".bonsai/local/ladder/T-0800.json" {
		t.Fatalf("exit %d, cleaned %v", a.Exit, a.Doc.Cleaned)
	}
	f.result("T-0901")
}

// The climb document holds the result as written, and its every field in order; bless is [] (5.4.4 fills it).
func TestClimbDocument(t *testing.T) {
	f := project(t, yamlOf("[1]", "{}", "[]", rung{n: 1, name: "one", command: "echo one", required: true}),
		map[string]string{"work/tasks/T-0901-a.md": taskText("T-0901", "running", "")})
	a := f.climb("", Options{Task: "T-0901"})
	b, err := a.Doc.Encode()
	if err != nil {
		t.Fatal(err)
	}
	doc := fits(t, "climb", b)
	written, _ := os.ReadFile(filepath.Join(f.root, ".bonsai", "local", "ladder", "T-0901.json"))
	in, _ := doc.Get("ladder")
	inRaw, _ := schema.Encode(in)
	if !bytes.Equal(inRaw, written) {
		t.Errorf("the document's ladder differs from the file:\n%s\n%s", inRaw, written)
	}
	if showOf(doc, "bless") != "[]" || showOf(doc, "next") != "null" || showOf(doc, "error") != "null" ||
		showOf(doc, "workspace") != `{"id":"`+testID+`","name":"climb-test","root":`+schema.Show(filepath.ToSlash(f.root))+`}` {
		t.Errorf("the document: %s", b)
	}
	if !json.Valid(b) {
		t.Error("not JSON")
	}
}

func sha256Hex(b []byte) string {
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:])
}

// A result that cannot be written, and a record that cannot be written, are write-failed (exit 3): the first proves
// nothing and leaves no result; the second leaves the result, named in the document, with the error saying its record
// is missing, so no climb counts as a proof without its fingerprint.
func TestWriteFailures(t *testing.T) {
	f := project(t, yamlOf("[1]", "{}", "[]", rung{n: 1, name: "one", command: "echo one", required: true}),
		map[string]string{"work/tasks/T-0901-a.md": taskText("T-0901", "running", "")})
	saved := writeFile
	writeFile = func(string, []byte) error { return os.ErrPermission }
	a := f.climb("", Options{Task: "T-0901"})
	writeFile = saved
	if a.Exit != ExitRuntime || a.Error == nil || a.Error.Code != "write-failed" || a.Doc.Result != nil || a.Doc.Ladder != nil {
		t.Errorf("the result not written: exit %d, %+v", a.Exit, a.Error)
	}
	if _, err := os.Stat(filepath.Join(f.root, ".bonsai", "local", "ladder", "T-0901.json")); err == nil {
		t.Error("a result was written")
	}
	savedLog := writeLog
	writeLog = func(string, *format.Log) (bool, error) { return false, os.ErrPermission }
	a = f.climb("", Options{Task: "T-0901"})
	writeLog = savedLog
	if a.Exit != ExitRuntime || a.Error == nil || a.Error.Code != "write-failed" || a.Doc.Result == nil || a.Doc.SHA256 == nil ||
		a.Doc.Next != nil || !strings.Contains(a.Error.What, "its ladder record cannot be written") {
		t.Errorf("the record not written: exit %d, %+v, %+v", a.Exit, a.Error, a.Doc)
	}
	if len(ladderRecords(t, f.root)) != 0 {
		t.Error("a record was written")
	}
}

// The result's dirty is git's status at the start or at the end, or HEAD moved during the climb: a rung that leaves a
// file, or commits, makes a climb that started clean dirty, so it proves no commit.
func TestDirtyDuringTheClimb(t *testing.T) {
	for _, c := range []struct{ name, command, says string }{
		{"a file left", "echo made > made.txt", "the tree was dirty"},
		{"a commit made", "git -c user.name=a -c user.email=a@example.invalid commit -q --allow-empty -m moved", "HEAD moved during the climb"},
	} {
		t.Run(c.name, func(t *testing.T) {
			f := project(t, yamlOf("[1]", "{}", "[]", rung{n: 1, name: "one", command: c.command, required: true}),
				map[string]string{"work/tasks/T-0901-a.md": taskText("T-0901", "running", "")})
			before := f.head()
			if a := f.climb("", Options{Task: "T-0901"}); a.Exit != 0 {
				t.Fatalf("exit %d, %+v\n%s", a.Exit, a.Error, f.out.String())
			}
			l := f.result("T-0901")
			if !l.Git.Dirty || l.Git.SHA != before || !strings.Contains(f.out.String(), c.says) {
				t.Errorf("git %+v, before %s\n%s", l.Git, before, f.out.String())
			}
		})
	}
}
