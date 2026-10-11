package guard

// The stop gate's rows of the rules table (rules.go; step 5.3.4, internal/stopgate) in the guard's one test per rule
// (rules_test.go's ruleTests, walked by TestEveryRule): each stop- row answered by bonsai hook stop's own Main on a
// gated project made by hand (its .git written as git writes it), the rule read back from the guard record it writes
// in the project's log (none for stop-not-named, which does not engage). The stop gate's own tests hold every case in
// full, with its reasons and next steps, the fixtures and the HEAD forms; these hold each row to an answer.

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/hookin"
	"github.com/LastStep/Bonsai/internal/stopgate"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// stopYAML is the gated project's bonsai.yaml: its floor is rung 0; stopTask is T-0901, running, adding rung 1.
const (
	stopYAML = "format: bonsai.workspace/1\nid: " + testID + "\nname: stop-rows\npacks: []\ndocuments:\n  task: work/tasks\n" +
		"  run: work/runs\n  answers: work/answers.md\n  memory: work/memory\n  protocols: work/protocols\nprotected: []\n" +
		"person_only: []\nnever_edit: []\nladder_floor: [0]\nladder: []\nratchets: {}\nci_marked_tests: []\ngenerated:\n" +
		"  log:\n    keep_days: 30\n    keep_newest: null\n  asks:\n    keep_days: null\n    keep_newest: null\n" +
		"  ladder:\n    keep_days: 7\n    keep_newest: null\n  run:\n    keep_days: null\n    keep_newest: null\n" +
		"  sessions:\n    keep_days: null\n    keep_newest: null\n"
	stopTask = "---\nformat: bonsai.task/1\nid: T-0901\ntitle: A gated task\nstatus: running\nlane: light\ndone_when:\n" +
		"  - \"It ends\"\ndepends_on: []\nblocked_by:\ncreated: 2026-10-11\nstarted: 2026-10-11\nfinished:\nlabels:\n" +
		"  bonsai.ladder: [1]\n---\n\n# A gated task\n"
	stopHead  = "5d6e7f8091a2b3c4d5e6f708192a3b4c5d6e7f80"
	stopOther = "0123456789abcdef0123456789abcdef01234567"
)

// stopProject makes the gated project, recorded on this machine in a home of the test's own (so its main checkout is
// verified), with a run report for T-0901 and a green local result proving stopHead with rungs 0 and 1; change, when
// given, then edits the result before it is written.
func stopProject(t *testing.T, change func(l *format.Ladder)) string {
	t.Helper()
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	put := func(rel, body string) {
		p := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	put("bonsai.yaml", stopYAML)
	put("work/tasks/T-0901-gated.md", stopTask)
	put("work/runs/R-2026-10-11-T-0901.md", "# R-2026-10-11-T-0901\n")
	put(".git/HEAD", "ref: refs/heads/main\n")
	put(".git/refs/heads/main", stopHead+"\n")
	home := t.TempDir()
	t.Setenv(workspace.HomeEnv, home)
	if _, err := workspace.RecordCheckout(home, dir, testID, testNow); err != nil {
		t.Fatal(err)
	}
	str := func(s string) *string { return &s }
	rung := func(n int64) format.LadderRung {
		return format.LadderRung{Rung: n, Name: "rung", Kind: "command", Required: true, Command: str("true"), Means: str("A rung."), Status: "green"}
	}
	l := &format.Ladder{Task: str("T-0901"), Workspace: testID, Mode: "local", Started: "2026-10-11T10:00:00.000Z",
		Finished: "2026-10-11T10:01:00.000Z", Git: format.LadderGit{SHA: stopHead, Branch: "main"}, Requested: []int64{0, 1},
		Green: true, Rungs: []format.LadderRung{rung(0), rung(1)}, Skipped: []format.LadderSkip{}}
	if change != nil {
		change(l)
	}
	raw, err := l.Encode()
	if err != nil {
		t.Fatal(err)
	}
	put(workspace.LocalDir+"/ladder/T-0901.json", string(raw))
	return dir
}

// stopRule runs the stop gate for a session in dir with BONSAI_TASK set to task ("" for unset) and gives its exit code
// and the rule of the last record in the project's log ("" when it wrote none).
func stopRule(t *testing.T, dir, task string) (int, string) {
	t.Helper()
	vars := map[string]string{hookin.ProjectEnv: dir, workspace.TaskEnv: task}
	in := `{"session_id": "s1", "hook_event_name": "Stop", "stop_hook_active": false}`
	var stderr bytes.Buffer
	code := stopgate.Main(stopgate.Options{Stdin: strings.NewReader(in), Stderr: &stderr, Budget: untimed,
		Getenv: func(k string) string { return vars[k] }, Now: func() time.Time { return testNow }})
	names, _ := filepath.Glob(filepath.Join(dir, filepath.FromSlash(LogDir), "*.ndjson"))
	rule := ""
	for _, name := range names {
		raw, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		lines := strings.Split(strings.TrimSuffix(string(raw), "\n"), "\n")
		var rec struct {
			Event      string `json:"event"`
			AgentEvent string `json:"agent_event"`
			Rule       string `json:"rule"`
		}
		if err := json.Unmarshal([]byte(lines[len(lines)-1]), &rec); err != nil {
			t.Fatal(err)
		}
		if rec.Event != "guard" || rec.AgentEvent != "Stop" {
			t.Errorf("the stop gate's record is %s/%s, not guard/Stop", rec.Event, rec.AgentEvent)
		}
		rule = rec.Rule
	}
	return code, rule
}

func init() {
	cases := []struct {
		rule   string
		task   string                         // BONSAI_TASK; "" for unset
		result func(l *format.Ladder)         // a change to the green result
		change func(t *testing.T, dir string) // a change to the project
	}{
		{rule: RuleStopProven, task: "T-0901"},
		{rule: RuleStopNotNamed},
		{rule: RuleStopTaskMissing, task: "T-0999"},
		{rule: RuleStopTaskTwice, task: "T-0901", change: func(t *testing.T, dir string) {
			writeFile(t, filepath.Join(dir, "work", "tasks", "T-0901-again.md"), stopTask)
		}},
		{rule: RuleStopTaskBroken, task: "T-0901", change: func(t *testing.T, dir string) {
			writeFile(t, filepath.Join(dir, "work", "tasks", "T-0901-gated.md"), strings.Replace(stopTask, "status: running", "status: [", 1))
		}},
		{rule: RuleStopTaskUnreadable, task: "T-0901", change: func(t *testing.T, dir string) {
			if runtime.GOOS == "windows" {
				// A file where the folder goes reads as no folder there (stop-task-missing), and any other unreadable
				// folder needs a permission change; the stop gate's own tests say so too, and the Linux run holds it.
				t.Skip("on Windows a file where the task folder goes reads as no folder; the Linux run holds this rule")
			}
			tasks := filepath.Join(dir, "work", "tasks")
			if err := os.RemoveAll(tasks); err != nil {
				t.Fatal(err)
			}
			writeFile(t, tasks, "a file where the task folder goes\n")
		}},
		{rule: RuleStopBadRungs, task: "T-0901", change: func(t *testing.T, dir string) {
			writeFile(t, filepath.Join(dir, "work", "tasks", "T-0901-gated.md"), strings.Replace(stopTask, "[1]", `["one"]`, 1))
		}},
		{rule: RuleStopNoRunReport, task: "T-0901", change: func(t *testing.T, dir string) {
			if err := os.Remove(filepath.Join(dir, "work", "runs", "R-2026-10-11-T-0901.md")); err != nil {
				t.Fatal(err)
			}
		}},
		{rule: RuleStopNoResult, task: "T-0901", change: func(t *testing.T, dir string) {
			if err := os.Remove(filepath.Join(dir, filepath.FromSlash(workspace.LocalDir), "ladder", "T-0901.json")); err != nil {
				t.Fatal(err)
			}
		}},
		{rule: RuleStopBadResult, task: "T-0901", change: func(t *testing.T, dir string) {
			writeFile(t, filepath.Join(dir, filepath.FromSlash(workspace.LocalDir), "ladder", "T-0901.json"), "{")
		}},
		{rule: RuleStopOtherTask, task: "T-0901", result: func(l *format.Ladder) { other := "T-0902"; l.Task = &other }},
		{rule: RuleStopNotLocal, task: "T-0901", result: func(l *format.Ladder) { l.Mode = "ci" }},
		{rule: RuleStopRed, task: "T-0901", result: func(l *format.Ladder) { l.Green, l.Rungs[1].Status = false, "red" }},
		{rule: RuleStopHeadUnreadable, task: "T-0901", change: func(t *testing.T, dir string) {
			if err := os.Remove(filepath.Join(dir, ".git", "refs", "heads", "main")); err != nil {
				t.Fatal(err)
			}
		}},
		{rule: RuleStopOtherCommit, task: "T-0901", result: func(l *format.Ladder) { l.Git.SHA = stopOther }},
		{rule: RuleStopDirty, task: "T-0901", result: func(l *format.Ladder) { l.Git.Dirty = true }},
		{rule: RuleStopRungMissing, task: "T-0901", result: func(l *format.Ladder) { l.Requested = []int64{0} }},
		{rule: RuleStopRungNotGreen, task: "T-0901", result: func(l *format.Ladder) { l.Rungs[1].Status = "pending" }},
	}
	for _, c := range cases {
		c := c
		ruleTests[c.rule] = func(t *testing.T) {
			dir := stopProject(t, c.result)
			if c.change != nil {
				c.change(t, dir)
			}
			code, rule := stopRule(t, dir, c.task)
			want := ruleAllows(t, c.rule)
			if c.rule == RuleStopNotNamed {
				rule = c.rule // allowed with no record
				if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(LogDir))); err == nil {
					t.Error("a gate that does not engage wrote a log")
				}
			}
			if rule != c.rule || (code == 0) != want {
				t.Errorf("the stop gate answered exit %d by %q, want %q (allow %v)", code, rule, c.rule, want)
			}
		}
	}
}
