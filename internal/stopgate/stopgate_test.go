package stopgate

// The stop gate's tests (design/plan-5.md, 5.3.4 note 9): one per blocking case of note 2 and of note 1, each on a
// project made by hand (its .git written as git writes it, so no git process is needed), the green case that allows,
// the cases that do not engage, the payload's faults, the budget, the records; then the stop gate's section of the
// active-task fixtures (formats/active-task, set 4's), every case built with git as the README says, each case whose
// task is ok given a fixture result proving its session's HEAD.

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/formats"
	"github.com/LastStep/Bonsai/internal/engine"
	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/guard"
	"github.com/LastStep/Bonsai/internal/hookin"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/workspace"
)

const testID = "ws-aaaaaaaaaaaaaaaaaaaaaaaaaa"

// testYAML is the test project's bonsai.yaml: its floor is rungs 0 and 1.
const testYAML = `format: bonsai.workspace/1
id: ` + testID + `
name: gate-test
packs: []
documents:
  task: work/tasks
  run: work/runs
  answers: work/answers.md
  memory: work/memory
  protocols: work/protocols
protected: []
person_only: []
never_edit: []
ladder_floor: [0, 1]
ladder: []
ratchets: {}
ci_marked_tests: []
generated:
  log:
    keep_days: 30
    keep_newest: null
  asks:
    keep_days: null
    keep_newest: null
  ladder:
    keep_days: 7
    keep_newest: null
  run:
    keep_days: null
    keep_newest: null
  sessions:
    keep_days: null
    keep_newest: null
`

// testTask is T-0901, running, whose bonsai.ladder adds rung 2 to the floor.
const testTask = `---
format: bonsai.task/1
id: T-0901
title: The gated task
status: running
lane: light
done_when:
  - "The gate lets it stop"
depends_on: []
blocked_by:
created: 2026-10-11
started: 2026-10-11
finished:
labels:
  bonsai.ladder: [2]
---

# The gated task
`

// otherSHA is a commit other than the test project's HEAD (headSHA, objects_test.go).
const otherSHA = "0123456789abcdef0123456789abcdef01234567"

// testNow is the runs' clock, so a record's day file is known.
var testNow = time.Date(2026, 10, 11, 12, 0, 0, 0, time.UTC)

// TestMain gives the package's tests a Bonsai home of their own, so no test reaches the real one; each project gives
// its test its own home in turn.
func TestMain(m *testing.M) {
	home, err := os.MkdirTemp("", "bonsai-stopgate-home-")
	if err != nil {
		panic(err)
	}
	if err := os.Setenv(workspace.HomeEnv, home); err != nil {
		panic(err)
	}
	code := m.Run()
	_ = os.RemoveAll(home)
	os.Exit(code)
}

// project makes the gated project in a temporary folder and gives its real path: bonsai.yaml, T-0901, a run report
// for it, a .git folder on branch main at headSHA (a loose ref file, as git writes it; the commit holding bonsai.yaml
// written as loose objects, objects_test.go), this machine's record of it in a home of the test's own (so its main
// checkout is verified, main its base branch), and a green local result proving headSHA with rungs 0, 1 and 2.
func project(t *testing.T) string {
	t.Helper()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	write(t, root, "bonsai.yaml", testYAML)
	write(t, root, "work/tasks/T-0901-gated.md", testTask)
	write(t, root, "work/runs/R-2026-10-11-T-0901.md", "# R-2026-10-11-T-0901\n\nWhat changed.\n")
	write(t, root, ".git/HEAD", "ref: refs/heads/main\n")
	write(t, root, ".git/refs/heads/main", commitYAML(t, filepath.Join(root, ".git"), testYAML)+"\n")
	home := t.TempDir()
	t.Setenv(workspace.HomeEnv, home)
	if _, err := workspace.RecordCheckout(home, root, testID, testNow); err != nil {
		t.Fatal(err)
	}
	writeResult(t, root, "T-0901", greenResult(headSHA))
	return root
}

func write(t *testing.T, root, rel, body string) {
	t.Helper()
	p := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func remove(t *testing.T, root, rel string) {
	t.Helper()
	if err := os.Remove(filepath.Join(root, filepath.FromSlash(rel))); err != nil {
		t.Fatal(err)
	}
}

func rung(n int64, status string) format.LadderRung {
	return format.LadderRung{Rung: n, Name: "rung " + strconv.FormatInt(n, 10), Kind: "command", Required: true,
		Command: ptr("go test ./..."), Means: ptr("A rung of the test."), Status: status}
}

// greenResult is a green local result proving sha with rungs 0, 1 and 2.
func greenResult(sha string) *format.Ladder {
	return &format.Ladder{Task: ptr("T-0901"), Workspace: testID, Mode: "local", Started: "2026-10-11T10:00:00.000Z",
		Finished: "2026-10-11T10:01:00.000Z", Git: format.LadderGit{SHA: sha, Branch: "main"}, Requested: []int64{0, 1, 2},
		Green: true, Rungs: []format.LadderRung{rung(0, "green"), rung(1, "green"), rung(2, "green")},
		Skipped: []format.LadderSkip{}}
}

// writeResult writes a ladder result for a task in a main checkout's .bonsai/local/ladder/, as the runner would.
func writeResult(t *testing.T, main, task string, l *format.Ladder) {
	t.Helper()
	raw, err := l.Encode()
	if err != nil {
		t.Fatal(err)
	}
	write(t, main, workspace.LocalDir+"/ladder/"+task+".json", string(raw))
}

// stopPayload is a Stop payload as Claude Code sends it (its hooks reference, 2.1.296), with the fields Bonsai does
// not read among them.
func stopPayload(session, cwd string) string {
	o := schema.Object{
		{Key: "session_id", Value: session},
		{Key: "transcript_path", Value: "/tmp/transcript.jsonl"},
		{Key: "cwd", Value: cwd},
		{Key: "permission_mode", Value: "default"},
		{Key: "hook_event_name", Value: "Stop"},
		{Key: "stop_hook_active", Value: false},
		{Key: "last_assistant_message", Value: "Done."},
		{Key: "background_tasks", Value: []any{}},
		{Key: "session_crons", Value: []any{}},
	}
	b, err := schema.EncodeLine(o)
	if err != nil {
		panic(err)
	}
	return string(b)
}

// env gives a Getenv with CLAUDE_PROJECT_DIR set to dir and the other variables given as pairs.
func env(dir string, kv ...string) func(string) string {
	m := map[string]string{hookin.ProjectEnv: dir}
	for i := 0; i+1 < len(kv); i += 2 {
		m[kv[i]] = kv[i+1]
	}
	return func(k string) string { return m[k] }
}

// untimed is the budget of every run that ends on the gate's own decision: a minute, which no run comes near, so a
// loaded machine can never turn its answer into over-time. Only the over-time test runs on a short budget.
const untimed = time.Minute

// gate runs Main and gives its exit code and stderr. Its clock is testNow.
func gate(t *testing.T, getenv func(string) string, stdin io.Reader, budget time.Duration) (int, string) {
	t.Helper()
	var stderr bytes.Buffer
	code := Main(Options{Stdin: stdin, Stderr: &stderr, Getenv: getenv, Budget: budget, Now: func() time.Time { return testNow }})
	return code, stderr.String()
}

// gateFor runs the gate for the session s1 in root, with BONSAI_TASK naming task.
func gateFor(t *testing.T, root, task string) (int, string) {
	t.Helper()
	return gate(t, env(root, workspace.TaskEnv, task), strings.NewReader(stopPayload("s1", root)), untimed)
}

// records reads every record of a main checkout's log, each held to log.schema.json and its field order, in file
// order (s-s1.ndjson, then the day files).
func records(t *testing.T, main string) []schema.Object {
	t.Helper()
	raw, err := formats.Schema("log")
	if err != nil {
		t.Fatal(err)
	}
	s, err := schema.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	names, _ := filepath.Glob(filepath.Join(main, filepath.FromSlash(workspace.LocalDir), "log", "*.ndjson"))
	var out []schema.Object
	for _, name := range names {
		b, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		for i, line := range strings.Split(strings.TrimSuffix(string(b), "\n"), "\n") {
			if line == "" {
				continue
			}
			v, err := schema.Decode([]byte(line))
			if err != nil {
				t.Fatalf("%s record %d does not parse: %v", filepath.Base(name), i, err)
			}
			if msgs := schema.Validate(s, v); len(msgs) != 0 {
				t.Errorf("%s record %d against log.schema.json: %v", filepath.Base(name), i, msgs)
			}
			if msgs := schema.CheckOrder(s, v); len(msgs) != 0 {
				t.Errorf("%s record %d field order: %v", filepath.Base(name), i, msgs)
			}
			out = append(out, v.(schema.Object))
		}
	}
	return out
}

// lastRecord is the last record of a main checkout's log, checked as the gate writes every one: a guard record of
// Claude Code's Stop, with no tool, its decision and rule, its text a block's reason.
func lastRecord(t *testing.T, main string) schema.Object {
	t.Helper()
	recs := records(t, main)
	if len(recs) == 0 {
		t.Fatal("no record")
	}
	r := recs[len(recs)-1]
	if r.String("event") != "guard" || r.String("agent_event") != "Stop" || r.String("agent") != "claude-code" ||
		r.String("workspace") == "" {
		t.Errorf("the record's fields: %s", schema.Show(r))
	}
	for _, k := range []string{"tool", "category", "tool_use_id", "input_hash"} {
		if v, _ := r.Get(k); v != nil {
			t.Errorf("a stop record's %s is %v, not null", k, v)
		}
	}
	text, _ := r.Get("text")
	if (r.String("decision") == "deny") != (text != nil) {
		t.Errorf("decision %s with text %v", r.String("decision"), text)
	}
	return r
}

// unsafeOut is what the gate's stderr must never say (safe's comment): Claude Code reads a Stop hook's exit 2 with
// such words as a hook that failed to start.
var unsafeOut = regexp.MustCompile(`(?i)no such file|cannot find|not found|missing|enoent`)

// blocks checks a run kept the session going by rule, its reason first naming want, its next step naming
// next, in one ASCII line each and with no unsafe word, and recorded in main's log.
func blocks(t *testing.T, code int, stderr, main, rule, want, next string) {
	t.Helper()
	if code != 2 {
		t.Fatalf("exit %d, want 2 (%s): %q", code, rule, stderr)
	}
	lines := strings.Split(strings.TrimSuffix(stderr, "\n"), "\n")
	if len(lines) != 2 || !strings.HasPrefix(lines[0], "bonsai stop gate: ") || !strings.HasPrefix(lines[1], "next: ") ||
		!strings.HasSuffix(lines[0], ".") || !strings.HasSuffix(lines[1], ".") {
		t.Fatalf("stderr is not the reason and the next step: %q", stderr)
	}
	if !strings.Contains(lines[0], want) || !strings.Contains(lines[1], next) {
		t.Errorf("%s: stderr %q, want %q in the reason and %q in the next step", rule, stderr, want, next)
	}
	if unsafeOut.MatchString(stderr) {
		t.Errorf("stderr says a file is missing as an operating system says it: %q", stderr)
	}
	for _, r := range stderr {
		if r != '\n' && (r < 0x20 || r >= 0x7f) {
			t.Errorf("stderr is not ASCII: %q", stderr)
			break
		}
	}
	if main == "" {
		return
	}
	r := lastRecord(t, main)
	if r.String("decision") != "deny" || r.String("rule") != rule {
		t.Errorf("the record: decision %s rule %s, want deny %s", r.String("decision"), r.String("rule"), rule)
	}
}

// allows checks a run let the session end, printing nothing, and recorded stop-proven in main's log.
func allows(t *testing.T, code int, stderr, main string) {
	t.Helper()
	if code != 0 || stderr != "" {
		t.Fatalf("exit %d, stderr %q; want 0 and nothing", code, stderr)
	}
	r := lastRecord(t, main)
	if r.String("decision") != "allow" || r.String("rule") != RuleProven || r.String("target") != workspace.LocalDir+"/ladder/T-0901.json" {
		t.Errorf("the allow's record: %s", schema.Show(r))
	}
}

// noLog checks a run wrote no log in the folder.
func noLog(t *testing.T, root string) {
	t.Helper()
	if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(workspace.LocalDir), "log")); err == nil {
		t.Errorf("a log was written in %s", root)
	}
}

// Green at HEAD allows, with its record; each blocking case of note 2 keeps the session going with its reason first
// and bonsai ladder --task as the next step, and is recorded.
func TestEachCase(t *testing.T) {
	const ladder = "bonsai ladder --task T-0901"
	cases := []struct {
		name, rule, why, next string
		change                func(t *testing.T, root string) // from the green project
		task                  string                          // what BONSAI_TASK names; "" for T-0901
	}{
		{name: "red", rule: RuleRed, why: "is not green (rung 1 rung 1 red: vet failed)", next: ladder,
			change: func(t *testing.T, root string) {
				l := greenResult(headSHA)
				l.Green, l.Rungs[1].Status, l.Rungs[1].Reason = false, "red", ptr("vet failed")
				writeResult(t, root, "T-0901", l)
			}},
		{name: "another commit", rule: RuleOtherCommit, why: `proves commit "0123456789ab", but this checkout's HEAD is ` + headSHA[:12],
			next: ladder + " at this HEAD", change: func(t *testing.T, root string) { writeResult(t, root, "T-0901", greenResult(otherSHA)) }},
		{name: "dirty", rule: RuleDirty, why: "uncommitted changes (git.dirty is true)", next: "commit or put away the changes, then run: " + ladder,
			change: func(t *testing.T, root string) {
				l := greenResult(headSHA)
				l.Git.Dirty = true
				writeResult(t, root, "T-0901", l)
			}},
		{name: "mode ci", rule: RuleNotLocal, why: `is a "ci" result, and only a local run is a task's proof`, next: ladder + " on this machine",
			change: func(t *testing.T, root string) {
				l := greenResult(headSHA)
				l.Mode, l.Proof = "ci", ptr("none: a CI result is never a task's proof")
				writeResult(t, root, "T-0901", l)
			}},
		{name: "a floor rung not asked for", rule: RuleRungMissing, why: "did not ask for rung 1, which the workspace's ladder_floor and the task's bonsai.ladder require (0, 1, 2)",
			next: ladder, change: func(t *testing.T, root string) {
				l := greenResult(headSHA)
				l.Requested, l.Rungs = []int64{0, 2}, []format.LadderRung{rung(0, "green"), rung(2, "green")}
				writeResult(t, root, "T-0901", l)
			}},
		{name: "a task rung not asked for", rule: RuleRungMissing, why: "did not ask for rung 2", next: ladder,
			change: func(t *testing.T, root string) {
				l := greenResult(headSHA)
				l.Requested, l.Rungs = []int64{0, 1}, []format.LadderRung{rung(0, "green"), rung(1, "green")}
				writeResult(t, root, "T-0901", l)
			}},
		{name: "a task rung not green", rule: RuleRungNotGreen, why: `rung 2, which the task needs, (rung 2) is "pending" in the ladder result for T-0901`,
			next: ladder, change: func(t *testing.T, root string) {
				l := greenResult(headSHA)
				l.Rungs[2].Status = "pending" // as today, a pending rung does not make a result red
				writeResult(t, root, "T-0901", l)
			}},
		{name: "a requested rung not among the rungs", rule: RuleRungNotGreen, why: "rung 2, which the task needs, is not among its rungs",
			next: ladder, change: func(t *testing.T, root string) {
				l := greenResult(headSHA)
				l.Rungs = l.Rungs[:2]
				writeResult(t, root, "T-0901", l)
			}},
		{name: "another task's result", rule: RuleOtherTask, why: `the ladder result .bonsai/local/ladder/T-0901.json was made for ` +
			`another task: its task is "T-0902", not T-0901`, next: ladder + ", which writes T-0901's own result there",
			change: func(t *testing.T, root string) {
				l := greenResult(headSHA)
				l.Task = ptr("T-0902")
				writeResult(t, root, "T-0901", l)
			}},
		{name: "a result with no task", rule: RuleOtherTask, why: "was made for another task: its task is null (a run for no task), not T-0901",
			next: ladder, change: func(t *testing.T, root string) {
				l := greenResult(headSHA)
				l.Task = nil
				writeResult(t, root, "T-0901", l)
			}},
		{name: "another workspace's result", rule: RuleOtherTask, why: `was made for another task: its workspace is ` +
			`"ws-bbbbbbbbbbbbbbbbbbbbbbbbbb", not this project's ` + testID, next: ladder,
			change: func(t *testing.T, root string) {
				l := greenResult(headSHA)
				l.Workspace = "ws-bbbbbbbbbbbbbbbbbbbbbbbbbb"
				writeResult(t, root, "T-0901", l)
			}},
		{name: "no result", rule: RuleNoResult, why: "no ladder result proves T-0901: the main checkout has no .bonsai/local/ladder/T-0901.json",
			next: "commit the work, then run: " + ladder, change: func(t *testing.T, root string) {
				remove(t, root, workspace.LocalDir+"/ladder/T-0901.json")
			}},
		{name: "a result that is not JSON", rule: RuleBadResult, why: "does not read as bonsai.ladder/1", next: ladder + ", which writes it again",
			change: func(t *testing.T, root string) {
				write(t, root, workspace.LocalDir+"/ladder/T-0901.json", "{\"format\": ")
			}},
		{name: "a result of another format", rule: RuleBadResult, why: "does not read as bonsai.ladder/1", next: ladder,
			change: func(t *testing.T, root string) {
				write(t, root, workspace.LocalDir+"/ladder/T-0901.json", `{"format": "bonsai.ladder/2", "green": true}`)
			}},
		{name: "a result that is a folder", rule: RuleBadResult, why: "cannot be read", next: ladder,
			change: func(t *testing.T, root string) {
				remove(t, root, workspace.LocalDir+"/ladder/T-0901.json")
				write(t, root, workspace.LocalDir+"/ladder/T-0901.json/x", "")
			}},
		{name: "a task no file holds", rule: RuleTaskMissing, why: "BONSAI_TASK names T-0999, and no task file has the named id (T-0999) in the main checkout's task folder work/tasks",
			next: "where bonsai ladder --task T-0999 reads it too", task: "T-0999"},
		{name: "a task named twice", rule: RuleTaskTwice, why: "two or more task files have the named id", next: ladder,
			change: func(t *testing.T, root string) { write(t, root, "work/tasks/T-0901-again.md", testTask) }},
		{name: "a task that does not parse", rule: RuleTaskBroken, why: "the named task's file does not parse (T-0901 in work/tasks/T-0901-gated.md)",
			next: ladder, change: func(t *testing.T, root string) {
				write(t, root, "work/tasks/T-0901-gated.md", strings.Replace(testTask, "status: running", "status: [", 1))
			}},
		{name: "a task folder that cannot be read", rule: RuleTaskUnreadable, why: `the main checkout's task folder "work/tasks" cannot be read`,
			next: ladder, change: func(t *testing.T, root string) {
				if runtime.GOOS == "windows" {
					t.Skip(unreadableOnWindows)
				}
				if err := os.RemoveAll(filepath.Join(root, "work", "tasks")); err != nil {
					t.Fatal(err)
				}
				write(t, root, "work/tasks", "a file where the task folder goes\n")
			}},
		{name: "a task's rungs that are not numbers", rule: RuleBadRungs, why: `the task's bonsai.ladder holds "two", not a rung number`,
			next: ladder, change: func(t *testing.T, root string) {
				write(t, root, "work/tasks/T-0901-gated.md", strings.Replace(testTask, "bonsai.ladder: [2]", `bonsai.ladder: ["two"]`, 1))
			}},
		{name: "main not verified", rule: RuleMainNotVerified, why: "the main checkout of this checkout is not verified: this machine has no record of the main checkout",
			next:   "a person runs bonsai update in the project's main checkout, which records it on this machine; then run: " + ladder,
			change: func(t *testing.T, root string) { t.Setenv(workspace.HomeEnv, t.TempDir()) }},
		{name: "main's HEAD with no commit yet", rule: RuleConfigUncommitted,
			why:  "the main checkout's HEAD names no commit yet, so it holds no bonsai.yaml, and the ladder's rungs and floor are read from the committed one only",
			next: UncommittedNext + "; then run: " + ladder, change: func(t *testing.T, root string) { remove(t, root, ".git/refs/heads/main") }},
		{name: "main's HEAD not a commit", rule: RuleMainNotVerified, why: "the main checkout is not on its base branch: the main checkout is on a detached HEAD",
			next: "a person checks out \"main\" in the main checkout", change: func(t *testing.T, root string) { write(t, root, ".git/HEAD", "main\n") }},
		{name: "main on a detached HEAD", rule: RuleMainNotVerified, why: `the main checkout is on a detached HEAD, and this machine's record names "main" as the workspace's base branch`,
			next: "(git switch main); then run: " + ladder, change: func(t *testing.T, root string) { write(t, root, ".git/HEAD", headSHA+"\n") }},
		{name: "main on another branch", rule: RuleMainNotVerified, why: `the main checkout is not on its base branch: the main checkout holds the branch "loose", and this machine's record names "main"`,
			next: "or, when \"loose\" is the base from now on, runs bonsai update in the main checkout to record it; then run: " + ladder,
			change: func(t *testing.T, root string) {
				write(t, root, ".git/HEAD", "ref: refs/heads/loose\n")
				write(t, root, ".git/refs/heads/loose", headSHA+"\n")
			}},
		{name: "a record from before Bonsai kept a base", rule: RuleMainNotVerified, why: "names no base branch (it was written before Bonsai kept one)",
			next: "a person checks out the project's base branch in the main checkout and runs bonsai update there, which records it as the base",
			change: func(t *testing.T, root string) {
				dir, err := workspace.MachineDir(os.Getenv(workspace.HomeEnv), root)
				if err != nil {
					t.Fatal(err)
				}
				rec, err := os.ReadFile(filepath.Join(dir, workspace.MachineRecordFile))
				if err != nil {
					t.Fatal(err)
				}
				write(t, dir, workspace.MachineRecordFile, strings.Replace(string(rec), `,
  "base_branch": "main"`, "", 1))
			}},
		{name: "main's bonsai.yaml changed, not committed", rule: RuleConfigUncommitted,
			why:  "the main checkout's bonsai.yaml differs from the one committed at its HEAD, and the ladder's rungs and floor are read from the committed one only",
			next: UncommittedNext + "; then run: " + ladder, change: func(t *testing.T, root string) {
				write(t, root, "bonsai.yaml", strings.Replace(testYAML, "ladder_floor: [0, 1]", "ladder_floor: [0]", 1))
			}},
		{name: "no run report", rule: RuleNoRunReport, why: "this checkout has no run report for T-0901 in its run folder work/runs",
			next: "write this session's run report, work/runs/R-<YYYY-MM-DD>-T-0901.md, in this checkout", change: func(t *testing.T, root string) {
				remove(t, root, "work/runs/R-2026-10-11-T-0901.md")
				write(t, root, "work/runs/R-2026-10-11-T-09012.md", "another task's report\n")
				write(t, root, "work/runs/notes-T-0901.md", "not a run report's name\n")
			}},
		{name: "a main bonsai.yaml that does not read in full", rule: RuleBadConfig, why: "cannot be read in full as committed at its HEAD", next: ladder,
			change: func(t *testing.T, root string) {
				recommit(t, root, strings.Replace(testYAML, "ladder_floor: [0, 1]", "ladder_floor: [zero]", 1))
			}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			root := project(t)
			if c.change != nil {
				c.change(t, root)
			}
			task := c.task
			if task == "" {
				task = "T-0901"
			}
			code, stderr := gateFor(t, root, task)
			blocks(t, code, stderr, root, c.rule, c.why, c.next)
		})
	}

	t.Run("green at HEAD", func(t *testing.T) {
		root := project(t)
		code, stderr := gateFor(t, root, "T-0901")
		allows(t, code, stderr, root)
		recs := records(t, root)
		if len(recs) != 1 || recs[0].String("session") != "s1" || recs[0].String("task") != "T-0901" ||
			recs[0].String("checkout") != filepath.Base(root) || recs[0].String("branch") != "main" {
			t.Errorf("the records: %v", recs)
		}
		if p, _ := recs[0].Get("bonsai_path"); p == nil {
			t.Errorf("the record that made the session's file names no binary: %s", schema.Show(recs[0]))
		}
		// A second end is recorded too, and only the file's first record names the binary.
		code, stderr = gateFor(t, root, "T-0901")
		allows(t, code, stderr, root)
		if recs = records(t, root); len(recs) != 2 {
			t.Fatalf("%d records", len(recs))
		} else if p, _ := recs[1].Get("bonsai_path"); p != nil {
			t.Errorf("a later record names the binary: %s", schema.Show(recs[1]))
		}
	})
	t.Run("green at HEAD through packed-refs", func(t *testing.T) {
		root := project(t)
		remove(t, root, ".git/refs/heads/main")
		write(t, root, ".git/packed-refs", "# pack-refs with: peeled fully-peeled sorted \n"+otherSHA+" refs/heads/feature\n"+
			headSHA+" refs/heads/main\n"+otherSHA+" refs/tags/v1\n^"+headSHA+"\n")
		code, stderr := gateFor(t, root, "T-0901")
		allows(t, code, stderr, root)
	})
	t.Run("green with a CRLF bonsai.yaml in the main checkout", func(t *testing.T) {
		root := project(t)
		write(t, root, "bonsai.yaml", strings.ReplaceAll(testYAML, "\n", "\r\n"))
		code, stderr := gateFor(t, root, "T-0901")
		allows(t, code, stderr, root)
	})
	t.Run("green with another bonsai.yaml, committed", func(t *testing.T) {
		root := project(t)
		recommit(t, root, strings.Replace(testYAML, "ladder_floor: [0, 1]", "ladder_floor: [0, 1, 2]", 1))
		code, stderr := gateFor(t, root, "T-0901")
		allows(t, code, stderr, root)
	})
	t.Run("green with a format-0 task's ladder", func(t *testing.T) {
		root := project(t)
		write(t, root, "work/tasks/T-0901-gated.md", "---\nid: T-0901\ntitle: Today's task\nstatus: running\nladder: [2]\n---\n")
		code, stderr := gateFor(t, root, "T-0901")
		allows(t, code, stderr, root)
		write(t, root, "work/tasks/T-0901-gated.md", "---\nid: T-0901\ntitle: Today's task\nstatus: running\nladder: [3]\n---\n")
		code, stderr = gateFor(t, root, "T-0901")
		blocks(t, code, stderr, root, RuleRungMissing, "did not ask for rung 3", "bonsai ladder --task T-0901")
	})
	t.Run("a session started in a subfolder", func(t *testing.T) {
		root := project(t)
		write(t, root, "docs/x/readme.md", "x\n")
		code, stderr := gate(t, env(filepath.Join(root, "docs", "x"), workspace.TaskEnv, "T-0901"),
			strings.NewReader(stopPayload("s1", root)), untimed)
		allows(t, code, stderr, root)
	})
}

// The gate engages only when BONSAI_TASK names a task and the session's project is linked: otherwise it lets the
// session end and records nothing, whatever the payload (note 1). A task named with no project folder to tell keeps
// the session going, with no record.
func TestWhenItEngages(t *testing.T) {
	root := project(t)
	remove(t, root, workspace.LocalDir+"/ladder/T-0901.json") // engaged, it would block
	for _, in := range []string{stopPayload("s1", root), "", "{not json", `{"hook_event_name": "PreToolUse"}`} {
		if code, stderr := gate(t, env(root), strings.NewReader(in), untimed); code != 0 || stderr != "" {
			t.Errorf("nothing named, payload %q: exit %d %q", in, code, stderr)
		}
	}
	noLog(t, root)
	unlinked, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if code, stderr := gate(t, env(unlinked, workspace.TaskEnv, "T-0901"), strings.NewReader(stopPayload("s1", unlinked)), untimed); code != 0 || stderr != "" {
		t.Errorf("an unlinked folder: exit %d %q", code, stderr)
	}
	noLog(t, unlinked)
	noLog(t, root)
	// Run by hand at a terminal, nothing named: it answers at once.
	if code := Main(Options{StdinTerminal: true, Getenv: env(root), Budget: untimed}); code != 0 {
		t.Errorf("by hand, nothing named: exit %d", code)
	}

	code, stderr := gate(t, func(k string) string {
		if k == workspace.TaskEnv {
			return "T-0901"
		}
		return ""
	}, strings.NewReader(stopPayload("s1", root)), untimed)
	blocks(t, code, stderr, "", RuleNoProject, "CLAUDE_PROJECT_DIR is not set", "Claude Code sets CLAUDE_PROJECT_DIR")
	code, stderr = gate(t, env(filepath.Join(root, "gone"), workspace.TaskEnv, "T-0901"), strings.NewReader(""), untimed)
	blocks(t, code, stderr, "", RuleNoProject, "cannot be read (it is not there)", "tell the person")
	noLog(t, root)

	write(t, root, "bonsai.yaml", "format: bonsai.workspace/1\nid: [\n")
	code, stderr = gateFor(t, root, "T-0901")
	blocks(t, code, stderr, "", RuleBadConfig, "bonsai.yaml cannot be read", "")
	noLog(t, root)
}

// Engaged, a payload the gate cannot read keeps the session going, recorded: empty, broken, a fault in a value it
// reads, another event; a payload with nothing Bonsai reads beyond its event is read.
func TestPayloadFaults(t *testing.T) {
	root := project(t)
	for _, c := range []struct{ name, in, why string }{
		{"empty", "", "the hook input is empty"},
		{"broken", "{\"hook_event_name\": ", "the hook input"},
		{"a fault", `{"hook_event_name": "Stop", "session_id": 7}`, "the hook input"},
		{"a duplicate key", `{"hook_event_name": "Stop", "hook_event_name": "Stop"}`, "the hook input"},
		{"another event", `{"hook_event_name": "SubagentStop", "session_id": "s1"}`, `is a "SubagentStop" payload: the stop gate answers Stop only`},
	} {
		t.Run(c.name, func(t *testing.T) {
			code, stderr := gate(t, env(root, workspace.TaskEnv, "T-0901"), strings.NewReader(c.in), untimed)
			blocks(t, code, stderr, root, RuleBadInput, c.why, "tell the person")
		})
	}
	if code, stderr := gate(t, env(root, workspace.TaskEnv, "T-0901"), strings.NewReader(`{"hook_event_name": "Stop"}`), untimed); code != 0 {
		t.Errorf("a Stop payload with nothing else: exit %d %q", code, stderr)
	}
	if code := Main(Options{StdinTerminal: true, Getenv: env(root, workspace.TaskEnv, "T-0901"), Budget: untimed,
		Stderr: io.Discard}); code != 2 {
		t.Errorf("by hand at a terminal, a task named: exit %d", code)
	}
}

// Over its budget the gate keeps the session going (it does not wait for its work), with the over-time record once
// it knows where the record goes; a panic inside keeps it going too. A held run is let go into a panic, so it writes
// nothing after its test.
func TestOverTimeAndPanicBlock(t *testing.T) {
	root := project(t)
	defer func(f func()) { beforeWork = f }(beforeWork)
	defer func(f func(workspace.Local, workspace.GitLayout, *untied, string) Decision) { judgeFn = f }(judgeFn)

	// Past the budget before anything is read: blocked, with no record (no workspace id to record under yet). The
	// held run says it started before the test goes on, so the test's later changes to the hooks come after its read.
	started, release1 := make(chan struct{}, 1), make(chan struct{})
	beforeWork = func() { started <- struct{}{}; <-release1; panic("let go") }
	start := time.Now()
	var stderr bytes.Buffer
	code := Main(Options{Stdin: strings.NewReader(stopPayload("s1", root)), Stderr: &stderr, Budget: 50 * time.Millisecond,
		RecordWait: time.Minute, Getenv: env(root, workspace.TaskEnv, "T-0901"), Now: func() time.Time { return testNow }})
	took := time.Since(start)
	<-started
	close(release1)
	if code != 2 || !strings.Contains(stderr.String(), "ran past its own 50ms limit") || took > 30*time.Second {
		t.Fatalf("over time: exit %d after %v, %q", code, took, stderr.String())
	}
	blocks(t, code, stderr.String(), "", RuleOverTime, "ran past its own 50ms limit", "end the session again")
	noLog(t, root)

	// Past the budget once bonsai.yaml was read and the payload too: the over-time block is recorded. As the guard's
	// test does, a run whose timer fires before its work knows where the record goes (no record, rightly) is tried
	// again on a new project with twice the budget, up to the gate's real Budget. Every held run is let go into a
	// panic only once each has reached its judgment, so none writes after its test.
	beforeWork = func() {}
	release2, entered := make(chan struct{}), make(chan struct{}, 64)
	judgeFn = func(workspace.Local, workspace.GitLayout, *untied, string) Decision {
		entered <- struct{}{}
		<-release2
		panic("let go")
	}
	runs := 0
	for budget := 150 * time.Millisecond; ; budget *= 2 {
		dir := project(t)
		stderr.Reset()
		start := time.Now()
		code = Main(Options{Stdin: strings.NewReader(stopPayload("s1", dir)), Stderr: &stderr, Budget: budget,
			RecordWait: time.Minute, Getenv: env(dir, workspace.TaskEnv, "T-0901"), Now: func() time.Time { return testNow }})
		runs++
		limit := fmt.Sprintf("ran past its own %s limit", budget)
		if took := time.Since(start); code != 2 || !strings.Contains(stderr.String(), limit) || took > budget+time.Minute+Budget {
			t.Fatalf("budget %v: exit %d after %v, %q", budget, code, took, stderr.String())
		}
		if len(records(t, dir)) > 0 {
			blocks(t, code, stderr.String(), dir, RuleOverTime, limit, "end the session again")
			break
		}
		if budget > Budget {
			t.Fatalf("budget %v, past the gate's real %v: the timer still fired before the work had read bonsai.yaml", budget, Budget)
		}
		t.Logf("budget %v: the timer fired before the work knew where its record goes; again with %v", budget, 2*budget)
	}
	for i := 0; i < runs; i++ {
		select {
		case <-entered:
		case <-time.After(time.Minute):
			t.Fatalf("%d of %d held runs reached their judgment within a minute", i, runs)
		}
	}
	close(release2)

	judgeFn = judge
	beforeWork = func() { panic("a test panic") }
	code, out := gateFor(t, root, "T-0901")
	blocks(t, code, out, "", RuleInternal, "the stop gate failed inside (a test panic)", "tell the person")
}

// An end the gate would allow but cannot record is kept going (log-failed), as the guard's call is.
func TestAnAllowItCannotRecordIsBlocked(t *testing.T) {
	root := project(t)
	write(t, root, workspace.LocalDir+"/log", "a file where the log folder goes\n")
	code, stderr := gateFor(t, root, "T-0901")
	if code != 2 || !strings.Contains(stderr, "would let the session end but cannot write its record") || unsafeOut.MatchString(stderr) {
		t.Errorf("exit %d %q", code, stderr)
	}
}

// unreadableOnWindows is why a task folder that cannot be read is tried on Linux alone: on Windows a file where the
// folder goes reads as no folder at all (the system reports the path not found, which the task folder's reader takes
// as an empty folder, and the gate blocks with stop-task-missing), and any other way needs a permission change, which
// no test may need there.
const unreadableOnWindows = "on Windows a file where the task folder goes reads as no folder (stop-task-missing), and " +
	"any other unreadable folder needs a permission change; the Linux run holds this case"

// The gate's rule names are copies of rows of the guard's rules table, their one home: each the same name with the
// same answer (allow or block), the shared eight the guard's own constants, and every stop- row of the table one of
// them. The gate's names are read from its own Go source (ruleConsts), so a constant added there and not to the table
// fails here.
func TestRuleNames(t *testing.T) {
	shared := map[string]string{RuleBadInput: guard.RuleBadInput, RuleBadConfig: guard.RuleBadConfig,
		RuleNoProject: guard.RuleNoProject, RuleMainNotVerified: guard.RuleMainNotVerified, RuleLogFailed: guard.RuleLogFailed,
		RuleOverTime: guard.RuleOverTime, RuleInternal: guard.RuleInternal, RuleNotLinked: guard.RuleNotLinked}
	for mine, theirs := range shared {
		if mine != theirs {
			t.Errorf("%q is the guard's %q", mine, theirs)
		}
	}
	mine := ruleConsts(t)
	for _, r := range []string{RuleProven, RuleNotNamed, RuleBadInput, RuleOtherTask} {
		if _, ok := mine[r]; !ok { // the reader found the constants at all
			t.Fatalf("the package's source read gives no constant for %q: %v", r, mine)
		}
	}
	table := map[string]bool{}
	for _, r := range guard.Rules {
		table[r.Name] = r.Allow
		if _, ok := mine[r.Name]; strings.HasPrefix(r.Name, "stop-") && !ok {
			t.Errorf("the rules table's %q is no rule of the stop gate", r.Name)
		}
	}
	for r, allow := range mine {
		theirs, ok := table[r]
		switch {
		case !ok:
			t.Errorf("the stop gate's %q is not in the guard's rules table", r)
		case theirs != allow:
			t.Errorf("%q: the table says allow %v, the stop gate's constant %v", r, theirs, allow)
		}
	}
}

// ruleConsts reads the stop gate's rule names from its own Go source, every file of the package but its tests: each
// constant named Rule..., by its value, with its answer as its line comment says it ("allow..." or "deny..."). A Rule
// constant it cannot read (a value that is not a string literal, a comment that says neither, a value twice) fails
// the test, so none is passed over.
func ruleConsts(t *testing.T) map[string]bool {
	t.Helper()
	files, err := filepath.Glob("*.go")
	if err != nil || len(files) == 0 {
		t.Fatalf("the package's Go files: %v %v", files, err)
	}
	out := map[string]bool{}
	fset := token.NewFileSet()
	for _, name := range files {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, name, nil, parser.ParseComments)
		if err != nil {
			t.Fatal(err)
		}
		for _, decl := range f.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok || gd.Tok != token.CONST {
				continue
			}
			for _, spec := range gd.Specs {
				vs := spec.(*ast.ValueSpec)
				for i, id := range vs.Names {
					if !strings.HasPrefix(id.Name, "Rule") {
						continue
					}
					var lit *ast.BasicLit
					if i < len(vs.Values) {
						lit, _ = vs.Values[i].(*ast.BasicLit)
					}
					if lit == nil || lit.Kind != token.STRING {
						t.Errorf("%s: %s is not a string literal", name, id.Name)
						continue
					}
					value, err := strconv.Unquote(lit.Value)
					if err != nil {
						t.Errorf("%s: %s: %v", name, id.Name, err)
						continue
					}
					says := strings.TrimSpace(vs.Comment.Text())
					if _, twice := out[value]; twice {
						t.Errorf("%s: %s: the rule %q is named twice", name, id.Name, value)
					}
					switch {
					case strings.HasPrefix(says, "allow"):
						out[value] = true
					case strings.HasPrefix(says, "deny"):
						out[value] = false
					default:
						t.Errorf("%s: %s's comment says neither allow nor deny: %q", name, id.Name, says)
					}
				}
			}
		}
	}
	return out
}

// The gate's own limit is half the timeout of its hook line.
func TestBudgetIsHalfTheHookTimeout(t *testing.T) {
	secs, err := strconv.Atoi(engine.StopTimeout)
	if err != nil {
		t.Fatalf("engine.StopTimeout %q: %v", engine.StopTimeout, err)
	}
	if 2*Budget != time.Duration(secs)*time.Second {
		t.Fatalf("Budget %v is not half of the hook line's timeout, %d s", Budget, secs)
	}
}

// What a refusal never says (safe): the words an operating system uses for a file that is not there.
func TestSafeWords(t *testing.T) {
	for in, want := range map[string]string{
		"open x: no such file or directory":                   "open x: not there",
		"The system cannot find the file specified.":          "not there.",
		"The system cannot find the path specified":           "not there",
		"sh: 1: bonsai: not found; ENOENT; a Missing rung":    "sh: 1: bonsai: not there; not there; a not there rung",
		"dismissing nothing; a notfound word stays; founding": "dismissing nothing; a notfound word stays; founding",
	} {
		if got := safe(in); got != want {
			t.Errorf("safe(%q) = %q, want %q", in, got, want)
		}
	}
}
