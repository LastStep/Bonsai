package guard

// The stop gate's rows of the rules table (rules.go; step 5.3.4, internal/stopgate) in the guard's one test per rule
// (rules_test.go's ruleTests, walked by TestEveryRule): each stop- row answered by bonsai hook stop's own Main on a
// gated project made by hand (its .git written as git writes it), the rule read back from the guard record it writes
// in the project's log (none for stop-not-named, which does not engage). The stop gate's own tests hold every case in
// full, with its reasons and next steps, the fixtures and the HEAD forms; these hold each row to an answer.

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/agents"
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
	stopOther = "0123456789abcdef0123456789abcdef01234567"
)

// stopHead is the commit the gated project's HEAD names: bonsai.yaml as stopYAML, committed (stopCommit).
var stopHead = stopCommitObjects(stopYAML)[2].id

// stopObject is one git object as git stores it loose before compression ("<kind> <size>\0<body>") and its id.
type stopObject struct {
	id  string
	raw []byte
}

// stopCommitObjects are bonsai.yaml's blob holding yaml, a tree holding it, and a commit of that tree with a fixed
// author and time, so its id is known before any test runs.
func stopCommitObjects(yaml string) []stopObject {
	obj := func(kind string, body []byte) stopObject {
		raw := append([]byte(fmt.Sprintf("%s %d\x00", kind, len(body))), body...)
		sum := sha1.Sum(raw)
		return stopObject{id: hex.EncodeToString(sum[:]), raw: raw}
	}
	blob := obj("blob", []byte(yaml))
	id, _ := hex.DecodeString(blob.id)
	tree := obj("tree", append([]byte("100644 bonsai.yaml\x00"), id...))
	commit := obj("commit", []byte("tree "+tree.id+"\nauthor t <t> 1760184000 +0000\ncommitter t <t> 1760184000 +0000\n\nthe fixture\n"))
	return []stopObject{blob, tree, commit}
}

// stopCommit writes the commit holding bonsai.yaml with yaml into dir/.git as loose objects (the stop gate reads
// bonsai.yaml as committed at the main checkout's HEAD, with git), with the config git needs, and gives its id.
func stopCommit(t *testing.T, dir, yaml string) string {
	t.Helper()
	writeFile(t, filepath.Join(dir, ".git", "config"), "[core]\n\trepositoryformatversion = 0\n\tbare = false\n")
	objs := stopCommitObjects(yaml)
	for _, o := range objs {
		var b bytes.Buffer
		z := zlib.NewWriter(&b)
		if _, err := z.Write(o.raw); err != nil {
			t.Fatal(err)
		}
		if err := z.Close(); err != nil {
			t.Fatal(err)
		}
		writeFile(t, filepath.Join(dir, ".git", "objects", o.id[:2], o.id[2:]), b.String())
	}
	return objs[2].id
}

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
	put(".git/refs/heads/main", stopCommit(t, dir, stopYAML)+"\n")
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

// stopRule runs the stop gate for a session in session (dir when "") with BONSAI_TASK set to task ("" for unset) and
// gives its exit code and the rule of the last record in the project's log, in dir ("" when it wrote none).
func stopRule(t *testing.T, dir, session, task string) (int, string) {
	t.Helper()
	if session == "" {
		session = dir
	}
	vars := map[string]string{hookin.ProjectEnv: session, workspace.TaskEnv: task}
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
		rule    string
		task    string                                // BONSAI_TASK; "" for unset
		result  func(l *format.Ladder)                // a change to the green result
		change  func(t *testing.T, dir string)        // a change to the project
		session func(t *testing.T, dir string) string // the session's folder, made for it; nil for the project's own
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
		{rule: RuleStopHeadUnreadable, task: "T-0901", session: func(t *testing.T, dir string) string {
			// A session in a worktree whose branch has no commit yet: its HEAD cannot be read, while the main
			// checkout's, which the gate reads bonsai.yaml at, can.
			wt := dir + "-wt"
			gitdir := filepath.Join(dir, ".git", "worktrees", "wt")
			writeFile(t, filepath.Join(wt, ".git"), "gitdir: "+filepath.ToSlash(gitdir)+"\n")
			writeFile(t, filepath.Join(gitdir, "HEAD"), "ref: refs/heads/unborn\n")
			writeFile(t, filepath.Join(gitdir, "commondir"), "../..\n")
			writeFile(t, filepath.Join(gitdir, "gitdir"), filepath.ToSlash(filepath.Join(wt, ".git"))+"\n")
			writeFile(t, filepath.Join(wt, "bonsai.yaml"), stopYAML)
			writeFile(t, filepath.Join(wt, "work", "runs", "R-2026-10-11-T-0901.md"), "# R-2026-10-11-T-0901\n")
			return wt
		}},
		{rule: RuleStopConfigUncommitted, task: "T-0901", change: func(t *testing.T, dir string) {
			writeFile(t, filepath.Join(dir, "bonsai.yaml"), strings.Replace(stopYAML, "ladder_floor: [0]", "ladder_floor: []", 1))
		}},
		{rule: RuleStopConfigLooser, task: "T-0901", change: func(t *testing.T, dir string) {
			// This machine accepted stricter settings than the ones committed (a path on protected): the committed
			// ones read as looser.
			writeFile(t, filepath.Join(dir, "bonsai.yaml"), strings.Replace(stopYAML, "protected: []", `protected: ["x/**"]`, 1))
			if _, err := workspace.RecordCheckout(os.Getenv(workspace.HomeEnv), dir, testID, testNow); err != nil {
				t.Fatal(err)
			}
			writeFile(t, filepath.Join(dir, "bonsai.yaml"), stopYAML)
		}},
		{rule: RuleStopOtherCommit, task: "T-0901", result: func(l *format.Ladder) { l.Git.SHA = stopOther }},
		{rule: RuleStopDirty, task: "T-0901", result: func(l *format.Ladder) { l.Git.Dirty = true }},
		{rule: RuleStopRungMissing, task: "T-0901", result: func(l *format.Ladder) { l.Requested = []int64{0} }},
		{rule: RuleStopRungNotGreen, task: "T-0901", result: func(l *format.Ladder) { l.Rungs[1].Status = "pending" }},
		{rule: RuleStopBudget, task: "T-0901", result: func(l *format.Ladder) { l.Green, l.Rungs[1].Status = false, "red" },
			change: func(t *testing.T, dir string) {
				// The named session s1, past its budget, its report written: it may stop, its ladder red.
				s := agents.Store{Main: dir}
				at := agents.Stamp(testNow)
				if _, err := s.Create(agents.Header{Agent: "s1", Session: "s1", Role: "builder", Kind: "lead", Named: true, First: at}); err != nil {
					t.Fatal(err)
				}
				for _, e := range []agents.Event{{E: agents.EventSpent, At: at}, {E: agents.EventReported, At: at}} {
					if err := s.Append("s1", "s1", e); err != nil {
						t.Fatal(err)
					}
				}
			}},
	}
	for _, c := range cases {
		c := c
		ruleTests[c.rule] = func(t *testing.T) {
			dir := stopProject(t, c.result)
			if c.change != nil {
				c.change(t, dir)
			}
			session := ""
			if c.session != nil {
				session = c.session(t, dir)
			}
			code, rule := stopRule(t, dir, session, c.task)
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
