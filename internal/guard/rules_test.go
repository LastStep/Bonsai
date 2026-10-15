package guard

// One test per rule (design/plan-5.md, 5.3.2's proof), walked from the rules table: TestEveryRule fails for a rule of
// Rules that has no entry in ruleTests, and for an entry that names no rule. Each entry makes the guard answer with
// its rule (through Main where a record or the hook's answer is the point, else through the judge's own functions)
// and checks the reason comes first with a next step. A rule an older test already proves points at that test.

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/testpack"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// grantYAML is a project with a task folder, a run folder, a protected list and a person-only list.
const grantYAML = `format: bonsai.workspace/1
id: ` + testID + `
name: guard-test
packs: []
documents:
  task: work/tasks
  run: work/runs
protected: [".claude/**", "bonsai.yaml", ".bonsai/lock.json", "assets/**", "docs/plan.md"]
person_only: [".claude/**", "bonsai.yaml", ".bonsai/lock.json", "keys/**"]
never_edit: []
`

// taskText is a format-1 task file: its id, status, lane and grants.
func taskText(id, status, lane string, allows ...string) string {
	q := make([]string, len(allows))
	for i, a := range allows {
		q[i] = `"` + a + `"`
	}
	return "---\nformat: bonsai.task/1\nid: " + id + "\ntitle: A made-up task\nstatus: " + status + "\nlane: " + lane +
		"\ndone_when:\n  - \"It is done\"\ndepends_on: []\nblocked_by:\ncreated: 2026-10-11\nstarted:\nfinished:\nlabels:\n" +
		"  bonsai.allows: [" + strings.Join(q, ", ") + "]\n---\n\n# A made-up task\n"
}

// grantProject makes a verified project (project) holding grantYAML's bonsai.yaml unless yaml is given, and the files
// given (project-relative, forward slashes).
func grantProject(t *testing.T, yaml string, files map[string]string) string {
	t.Helper()
	if yaml == "" {
		yaml = grantYAML
	}
	dir := project(t, yaml)
	for rel, body := range files {
		writeFile(t, filepath.Join(dir, filepath.FromSlash(rel)), body)
	}
	return dir
}

// withTask makes a verified project whose one task, T-0001, reads status and grants allows.
func withTask(t *testing.T, status string, allows ...string) string {
	t.Helper()
	return grantProject(t, "", map[string]string{"work/tasks/T-0001-the-task.md": taskText("T-0001", status, "light", allows...),
		"assets/a.png": "png\n", "keys/a.key": "key\n", "free.txt": "free\n"})
}

// commandMode writes this machine's settings for the project at dir: status_writes command.
func commandMode(t *testing.T, dir string) {
	t.Helper()
	machineSettings(t, dir, `{"status_writes": "command", "status_command": "studio-move"}`)
}

// machineSettings writes this machine's settings file for the project at dir, in the test's Bonsai home.
func machineSettings(t *testing.T, dir, body string) {
	t.Helper()
	home, err := workspace.Home()
	if err != nil {
		t.Fatal(err)
	}
	md, err := workspace.MachineDir(home, dir)
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(md, workspace.MachineSettingsFile), body)
}

// toolPayload builds a PreToolUse payload for tool with its tool_input.
func toolPayload(session, cwd, tool string, input schema.Object) string {
	o := schema.Object{
		{Key: "session_id", Value: session}, {Key: "transcript_path", Value: "/tmp/transcript.jsonl"}, {Key: "cwd", Value: cwd},
		{Key: "permission_mode", Value: "acceptEdits"}, {Key: "hook_event_name", Value: "PreToolUse"},
		{Key: "tool_name", Value: tool}, {Key: "tool_input", Value: input}, {Key: "tool_use_id", Value: "toolu_01TEST"},
	}
	b, err := schema.EncodeLine(o)
	if err != nil {
		panic(err)
	}
	return string(b)
}

// writeOf is a Write of content at rel in dir; editIn an Edit of old to new; bashIn a Bash call.
func writeOf(dir, rel, content string) string {
	return toolPayload("r1", dir, "Write", schema.Object{{Key: "file_path", Value: filepath.Join(dir, filepath.FromSlash(rel))},
		{Key: "content", Value: content}})
}

func editIn(dir, rel, old, repl string) string {
	return toolPayload("r1", dir, "Edit", schema.Object{{Key: "file_path", Value: filepath.Join(dir, filepath.FromSlash(rel))},
		{Key: "old_string", Value: old}, {Key: "new_string", Value: repl}})
}

func bashIn(dir, tool, command string) string {
	return toolPayload("r1", dir, tool, schema.Object{{Key: "command", Value: command}})
}

// answer runs the guard on a payload in a session whose project is dir (and the variables given as pairs): its exit
// code and stderr.
func answer(t *testing.T, dir, in string, kv ...string) (int, string) {
	t.Helper()
	return guardRun(t, env(dir, kv...), strings.NewReader(in), untimed)
}

// wantRule runs the guard and checks it answers with rule, the reason first and a next step for a refusal, ASCII,
// and records rule as the last record of the session's file in the project at logDir ("" for dir).
func wantRule(t *testing.T, dir, in, rule string, kv ...string) string {
	t.Helper()
	code, stderr := answer(t, dir, in, kv...)
	refuses := !ruleAllows(t, rule)
	if (code == 2) != refuses || (refuses && (!strings.HasPrefix(stderr, "bonsai guard: ") || !strings.Contains(stderr, "\nnext: "))) {
		t.Fatalf("want %s: exit %d, stderr %q", rule, code, stderr)
	}
	asciiOnly(t, stderr)
	recs := records(t, dir, "s-r1.ndjson")
	if got := recs[len(recs)-1].String("rule"); got != rule {
		t.Fatalf("want %s, recorded %s (stderr %q)", rule, got, stderr)
	}
	return stderr
}

// ruleAllows is a rule's row's Allow.
func ruleAllows(t *testing.T, name string) bool {
	for _, r := range Rules {
		if r.Name == name {
			return r.Allow
		}
	}
	t.Fatalf("no rule %q in the table", name)
	return false
}

// judgeIn is a judge for a session whose project is dir, reading this test's home.
func judgeIn(dir string) *judge {
	j := newJudge(guardHome())
	j.getenv = func(string) string { return "" }
	if dir != "" {
		j.session = j.project(dir)
	}
	return j
}

// mayIn judges a change of path (absolute) as the delete check does (mayDelete), or as a file tool with no content.
func mayIn(j *judge, dir, path string, del bool) Decision {
	if del {
		return j.mayDelete(dir, path)
	}
	d, _ := j.may(change{tool: "Edit", path: path, dir: dir, base: dir})
	return d
}

// ruleTests is one test per rule of the table.
var ruleTests = map[string]func(t *testing.T){
	RuleProtected: func(t *testing.T) {
		dir := withTask(t, "todo", "assets/**")
		stderr := wantRule(t, dir, editIn(dir, "assets/a.png", "png", "jpg"), RuleProtected)
		if !strings.Contains(stderr, `on bonsai.yaml's protected list ("assets/**"), and no active task: no task reads running`) {
			t.Errorf("the reason names no active task: %s", stderr)
		}
		// Another entry of the list (a pack's declared protected path: TestGuardReadOfTheLockIsLean).
		wantRule(t, dir, writeOf(dir, "docs/plan.md", "plan\n"), RuleProtected)
	},
	RulePersonOnly: func(t *testing.T) {
		dir := withTask(t, "running", "assets/**")
		stderr := wantRule(t, dir, editIn(dir, "keys/a.key", "key", "k"), RulePersonOnly)
		if !strings.Contains(stderr, `the running task T-0001 does not grant it (its bonsai.allows: "assets/**")`) {
			t.Errorf("the reason names the running task's grants: %s", stderr)
		}
		// The floor's person-only paths, whatever the lists say: a nested bonsai.yaml or .bonsai/, the lock, anything
		// else in .bonsai/ (5.3 done, check 4: an agent's Write of docs/bonsai.yaml refused).
		for _, rel := range []string{"docs/bonsai.yaml", "a/b/.bonsai/lock.json", ".bonsai/extra.json", "bonsai.yaml",
			".bonsai/lock.json", "docs/bonsai.yaml/x"} {
			if s := wantRule(t, dir, writeOf(dir, rel, "x\n"), RulePersonOnly); !strings.Contains(s, "refused: it is on") &&
				!strings.Contains(s, "refused: Bonsai guards it in every project (") {
				t.Errorf("%s: %s", rel, s)
			}
		}
	},
	RuleNeverGranted: func(t *testing.T) {
		// A running task granting everything opens none of them (5.3 done, check 4).
		dir := withTask(t, "running", "**", ".claude/**", ".git/**")
		for _, rel := range []string{".claude/settings.json", ".claude/settings.local.json", ".git/config", ".git",
			"docs/.git", "vendor/lib/.git/HEAD"} {
			wantRule(t, dir, writeOf(dir, rel, "x\n"), RuleNeverGranted)
		}
		if code, s := answer(t, dir, writeOf(dir, ".claude/agents/x.md", "x\n")); code != 0 {
			t.Errorf("a granted .claude/ path other than the settings files: exit %d, %s", code, s)
		}
		// For a delete, the floor's .git is the project's own top one only (5.3.3 note 4).
		j := judgeIn(dir)
		if d := mayIn(j, dir, filepath.Join(dir, "vendor", "lib", ".git"), true); !d.Allow {
			t.Errorf("a delete of a nested .git: %+v", d)
		}
		if d := mayIn(j, dir, filepath.Join(dir, ".git"), true); d.Rule != RuleNeverGranted {
			t.Errorf("a delete of the project's .git: %+v", d)
		}
	},
	RuleBonsaiFiles: func(t *testing.T) {
		dir := withTask(t, "running", "**", ".bonsai/**")
		for _, rel := range []string{".bonsai/local/log/s-x.ndjson", ".bonsai/local/ladder/T-0001.json", ".bonsai/.gitignore",
			".bonsai/sessions.md"} {
			wantRule(t, dir, writeOf(dir, rel, "x\n"), RuleBonsaiFiles)
		}
		if s := wantRule(t, dir, writeOf(dir, ".bonsai/tasks.md", "x\n"), RuleBonsaiFiles); !strings.Contains(s, "bonsai check --write") {
			t.Errorf("a table's refusal does not name bonsai check --write: %s", s)
		}
		// Bonsai's home, in every mode, granted or not: a forged machine record among them (5.3 done, check 4).
		home, err := workspace.Home()
		if err != nil {
			t.Fatal(err)
		}
		key, err := workspace.MachineKey(dir)
		if err != nil {
			t.Fatal(err)
		}
		forged := filepath.Join(home, "workspaces", key, workspace.MachineRecordFile)
		for _, p := range []string{forged, filepath.Join(home, "settings.json"), filepath.Join(home, "new", "x.json")} {
			in := toolPayload("r1", dir, "Write", schema.Object{{Key: "file_path", Value: p}, {Key: "content", Value: "{}"}})
			if s := wantRule(t, dir, in, RuleBonsaiFiles); !strings.Contains(s, "Bonsai's home") {
				t.Errorf("%s: %s", p, s)
			}
		}
		commandMode(t, dir)
		wantRule(t, dir, toolPayload("r1", dir, "Edit", schema.Object{{Key: "file_path", Value: filepath.Join(home, "workspaces",
			key, workspace.MachineSettingsFile)}, {Key: "old_string", Value: "command"}, {Key: "new_string", Value: "agents"}}), RuleBonsaiFiles)
		// ~/.bonsai as well, when BONSAI_HOME names another folder (the user's home folder is the test's own here).
		user := t.TempDir()
		t.Setenv("HOME", user)
		t.Setenv("USERPROFILE", user)
		in := toolPayload("r1", dir, "Write", schema.Object{{Key: "file_path", Value: filepath.Join(user, ".bonsai", "salt")},
			{Key: "content", Value: "x"}})
		wantRule(t, dir, in, RuleBonsaiFiles)
		// On Linux, a Windows user's Bonsai home as WSL mounts it (read on the string alone, so on every system).
		for p, want := range map[string]string{
			"/mnt/c/Users/nobody-here/.bonsai/workspaces/r-0/workspace.json": "/mnt/c/Users/nobody-here/.bonsai",
			"/mnt/D/users/x/.BONSAI": "/mnt/D/users/x/.BONSAI", "/mnt/c/Users/x/.bonsai2/y": "", "/mnt/c/Users/.bonsai/y": "",
			"/mnt/cc/Users/x/.bonsai/y": "", "/home/x/mnt/c/Users/x/.bonsai/y": "",
		} {
			if got := wslWindowsHome(p, "linux"); got != want {
				t.Errorf("wslWindowsHome(%s) = %q, want %q", p, got, want)
			}
			if got := wslWindowsHome(p, "windows"); got != "" {
				t.Errorf("wslWindowsHome(%s) on Windows = %q", p, got)
			}
		}
		if runtime.GOOS == "linux" {
			in := toolPayload("r1", dir, "Write", schema.Object{{Key: "file_path",
				Value: "/mnt/c/Users/nobody-here/.bonsai/workspaces/r-0123456789abcdef/workspace.json"}, {Key: "content", Value: "{}"}})
			wantRule(t, dir, in, RuleBonsaiFiles)
		}
	},
	RuleStandIn: func(t *testing.T) {
		dir := withTask(t, "running", "**")
		out := t.TempDir()
		for _, name := range []string{"bonsai", "bonsai.exe", "Bonsai.CMD", "bonsai.bat", "bonsai.ps1", "bonsai.com"} {
			in := toolPayload("r1", dir, "Write", schema.Object{{Key: "file_path", Value: filepath.Join(out, "bin", name)},
				{Key: "content", Value: "#!/bin/sh\nexit 0\n"}})
			wantRule(t, dir, in, RuleStandIn)
		}
		for _, name := range []string{"bonsai-notes.txt", "bonsai.yaml.bak"} {
			in := toolPayload("r1", dir, "Write", schema.Object{{Key: "file_path", Value: filepath.Join(out, name)}, {Key: "content", Value: "x"}})
			if code, s := answer(t, dir, in); code != 0 {
				t.Errorf("%s outside every project: exit %d, %s", name, code, s)
			}
		}
		// Inside a project, a file named bonsai is the lists' to judge.
		if code, s := answer(t, dir, writeOf(dir, "bin/bonsai", "x\n")); code != 0 {
			t.Errorf("bin/bonsai inside the project: exit %d, %s", code, s)
		}
	},
	RuleMainNotVerified: TestAnUnrecordedClone,
	RuleOtherProject: func(t *testing.T) {
		dir := withTask(t, "running", "**")
		home, _ := workspace.Home()
		other := unrecordedProject(t, strings.Replace(grantYAML, testID, "ws-bbbbbbbbbbbbbbbbbbbbbbbbbb", 1))
		if _, err := workspace.RecordCheckout(home, other, "ws-bbbbbbbbbbbbbbbbbbbbbbbbbb", time.Now()); err != nil {
			t.Fatal(err)
		}
		scratch := unrecordedProject(t, strings.Replace(grantYAML, testID, "ws-cccccccccccccccccccccccccc", 1))
		j := judgeIn(dir)
		for _, f := range floorFolders {
			if d := mayIn(j, dir, filepath.Join(other, f), true); d.Rule != RuleOtherProject || d.Why == "" || d.Next == "" {
				t.Errorf("a delete of another project's %s: %+v", f, d)
			}
			if d := mayIn(j, dir, filepath.Join(scratch, f), true); !d.Allow {
				t.Errorf("a delete of a project on no record of this machine's (%s): %+v", f, d)
			}
		}
		for _, rel := range []string{"bonsai.yaml", ".bonsai/lock.json", "assets/a.png", "src/x.go"} {
			if d := mayIn(j, dir, filepath.Join(other, filepath.FromSlash(rel)), true); !d.Allow {
				t.Errorf("a named delete below another project's floor folders (%s): %+v", rel, d)
			}
		}
		// A file tool's edit in the other project is judged by its own lists (5.3.1), not by this rule.
		if d := mayIn(j, dir, filepath.Join(other, "bonsai.yaml"), false); d.Rule != RulePersonOnly {
			t.Errorf("an edit of another project's bonsai.yaml: %+v", d)
		}
	},
	RuleGranted: func(t *testing.T) {
		dir := withTask(t, "running", "assets/**", "keys/**", "docs/bonsai.yaml")
		for _, rel := range []string{"assets/a.png", "assets/new/b.png", "keys/a.key", "docs/bonsai.yaml"} {
			wantRule(t, dir, writeOf(dir, rel, "x\n"), RuleGranted)
		}
		// The path closes when the task reads verify (5.3 done, check 4).
		writeFile(t, filepath.Join(dir, "work", "tasks", "T-0001-the-task.md"), taskText("T-0001", "verify", "light", "assets/**"))
		if s := wantRule(t, dir, writeOf(dir, "assets/a.png", "x\n"), RuleProtected); !strings.Contains(s, "no active task: no task reads running") {
			t.Errorf("the reason, nothing named: %s", s)
		}
		if s := wantRule(t, dir, writeOf(dir, "assets/a.png", "x\n"), RuleProtected, workspace.TaskEnv, "T-0001"); !strings.Contains(s,
			`the active task T-0001 reads "verify", and a task's grants count only while it reads running`) {
			t.Errorf("the reason, the task named: %s", s)
		}
	},
	RuleNotProtected: func(t *testing.T) {
		dir := withTask(t, "todo")
		wantRule(t, dir, editIn(dir, "free.txt", "free", "freer"), RuleNotProtected)
		wantRule(t, dir, writeOf(dir, ".bonsai/STATE.md", "# State\n"), RuleNotProtected)
	},
	RuleOutside: func(t *testing.T) {
		dir := withTask(t, "todo")
		in := toolPayload("r1", dir, "Write", schema.Object{{Key: "file_path", Value: filepath.Join(t.TempDir(), "x.txt")},
			{Key: "content", Value: "x"}})
		wantRule(t, dir, in, RuleOutside)
	},
	RuleTaskOutsideMain: func(t *testing.T) {
		tmp := t.TempDir()
		home := testpack.Isolate(t, tmp)
		main, wt := gitProject(t, home, grantYAML, time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC))
		commandMode(t, main)
		in := writeOf(wt, "work/tasks/T-0002-new.md", taskText("T-0002", "todo", "light"))
		code, stderr := answer(t, wt, in)
		if code != 2 || !strings.Contains(stderr, "a task file outside the main checkout") || !strings.Contains(stderr, `"studio-move"`) {
			t.Fatalf("exit %d, %s", code, stderr)
		}
		if got := recordsIn(t, main, "r1"); len(got) != 1 || !strings.HasPrefix(got[0], "deny "+RuleTaskOutsideMain+" work/tasks/T-0002-new.md wt ") {
			t.Fatalf("records %q", got)
		}
		// The same file in the main checkout is a new task at its first status: allowed.
		if code, s := answer(t, main, writeOf(main, "work/tasks/T-0002-new.md", taskText("T-0002", "todo", "light"))); code != 0 {
			t.Errorf("in the main checkout: exit %d, %s", code, s)
		}
	},
	RuleStatusLine: func(t *testing.T) {
		dir := withTask(t, "running")
		commandMode(t, dir)
		task := "work/tasks/T-0001-the-task.md"
		if s := wantRule(t, dir, editIn(dir, task, "status: running", "status: done"), RuleStatusLine); !strings.Contains(s,
			`changes the document's status line ("running" to "done")`) || !strings.Contains(s, "next: run this workspace's status command") {
			t.Errorf("the reason and step: %s", s)
		}
		// A new task starts at its first status, todo.
		if s := wantRule(t, dir, writeOf(dir, "work/tasks/T-0002-new.md", taskText("T-0002", "running", "light")), RuleStatusLine); !strings.Contains(s, `first status, "todo"`) {
			t.Errorf("a new task: %s", s)
		}
		wantRule(t, dir, writeOf(dir, "work/tasks/T-0002-new.md", taskText("T-0002", "todo", "light")), RuleNotProtected)
		// An edit whose old_string is not in the file leaves the status line unknown: refused.
		wantRule(t, dir, editIn(dir, task, "status: nowhere", "status: done"), RuleStatusLine)
		// Line endings and other lines: a CRLF file edited with LF text; the title edited.
		writeFile(t, filepath.Join(dir, filepath.FromSlash(task)), strings.ReplaceAll(taskText("T-0001", "running", "light"), "\n", "\r\n"))
		wantRule(t, dir, editIn(dir, task, "status: running\nlane", "status: verify\nlane"), RuleStatusLine)
		wantRule(t, dir, editIn(dir, task, "title: A made-up task", "title: A better title"), RuleNotProtected)
		// In agents mode, the same edit is the agent's own.
		machineSettings(t, dir, `{"status_writes": "agents"}`)
		wantRule(t, dir, editIn(dir, task, "status: running", "status: done"), RuleNotProtected)
	},
	RuleLaneLine: func(t *testing.T) {
		dir := withTask(t, "running")
		commandMode(t, dir)
		wantRule(t, dir, editIn(dir, "work/tasks/T-0001-the-task.md", "lane: light", "lane: full"), RuleLaneLine)
	},
	RuleGrantPersonOnly: func(t *testing.T) {
		dir := withTask(t, "todo", "assets/**")
		commandMode(t, dir)
		task := "work/tasks/T-0001-the-task.md"
		for _, g := range []string{"keys/a.key", "keys/**", "**", "*.yaml", "bonsai.yaml", "src/bonsai.yaml", ".claude/agents/x.md", ".bonsai/**"} {
			wantRule(t, dir, editIn(dir, task, `["assets/**"]`, `["assets/**", "`+g+`"]`), RuleGrantPersonOnly)
		}
		for _, g := range []string{"src/**", "docs/plan.md", ".bonsai/STATE.md"} {
			wantRule(t, dir, editIn(dir, task, `["assets/**"]`, `["assets/**", "`+g+`"]`), RuleNotProtected)
		}
	},
	RuleOutsideLabel: func(t *testing.T) {
		dir := withTask(t, "todo")
		commandMode(t, dir)
		home, _ := workspace.Home()
		md, _ := workspace.MachineDir(home, dir)
		def := func(name, kind, setBy, about string) string {
			return "  - name: " + name + "\n    kind: " + kind + "\n    values: []\n    items:\n    pattern:\n    max:\n" +
				"    kinds: [task]\n    set_by: " + setBy + "\n    grants: false\n    description: \"" + about + "\"\n"
		}
		writeFile(t, filepath.Join(md, workspace.MachineLabelsDir, "acme.yaml"), "format: bonsai.labels/1\nnamespace: acme\n"+
			"version: 1\nlabels:\n"+def("acme.cost", "number", "outside", "Dollars this work cost. Agents never write it.")+
			def("acme.owner", "text", "agent", "Who owns it."))
		task := "work/tasks/T-0001-the-task.md"
		wantRule(t, dir, editIn(dir, task, "labels:\n", "labels:\n  acme.cost: 3\n"), RuleOutsideLabel)
		wantRule(t, dir, editIn(dir, task, "labels:\n", "labels:\n  acme.owner: builder\n"), RuleNotProtected)
	},
	RuleBadFormat: func(t *testing.T) {
		dir := withTask(t, "todo")
		commandMode(t, dir)
		s := wantRule(t, dir, writeOf(dir, "work/tasks/T-0003-broken.md", strings.Replace(taskText("T-0003", "todo", "light"),
			"status: todo", "status: finished", 1)), RuleBadFormat)
		if !strings.Contains(s, "bonsai check --schema bonsai.task") {
			t.Errorf("the next step: %s", s)
		}
		wantRule(t, dir, writeOf(dir, ".bonsai/STATE.md", "---\nformat: bonsai.state/1\nupdated: never\n---\n"), RuleBadFormat)
		// A task file whose frontmatter holds nothing is no task, as every reader of the task folder reads it
		// (workspace.ReadTask): format 0 would read it as an empty mapping, and it would make the active task none for
		// everyone. Refused new, and refused over a task that reads.
		for _, body := range []string{"garbage: [\n", "---\n---\n\n# Notes\n", "# Notes alone\n", "---\n# a comment\n---\n"} {
			s := wantRule(t, dir, writeOf(dir, "work/tasks/T-0004-empty.md", body), RuleBadFormat)
			if !strings.Contains(s, "its frontmatter holds nothing") {
				t.Errorf("%q: the reason: %s", body, s)
			}
		}
		wantRule(t, dir, writeOf(dir, "work/tasks/T-0001-the-task.md", "garbage: [\n"), RuleBadFormat)
	},
	RuleBadSettings: func(t *testing.T) {
		dir := withTask(t, "todo")
		machineSettings(t, dir, "not json")
		wantRule(t, dir, editIn(dir, "work/tasks/T-0001-the-task.md", "title: A made-up task", "title: Another"), RuleBadSettings)
		// A path that is no document never reads the mode.
		wantRule(t, dir, editIn(dir, "free.txt", "free", "freer"), RuleNotProtected)
	},
	RulePersonCommand: func(t *testing.T) {
		dir := withTask(t, "running", "bonsai.yaml", ".bonsai/lock.json")
		j := judgeIn(dir)
		for _, argv := range [][]string{{"bonsai", "update", "--allow-exec", "--yes"}, {"bonsai", "unlink", "--yes"},
			{"bonsai", "init", "--allow-exec"}, {"/usr/local/bin/bonsai", "update", "--ALLOW-EXEC=true"}, {"bonsai", "check", "-allow-exec"}} {
			if d := j.ownCommand(dir, argv, true); d.Rule != RulePersonCommand || !strings.Contains(d.Why, "status_writes agents") ||
				!strings.HasPrefix(d.Next, "a person runs this command") {
				t.Errorf("%q in agents mode: %+v", argv, d)
			}
		}
		for _, line := range []string{"unset CLAUDE_CODE_CHILD_SESSION; bonsai update --yes",
			"CLAUDE_CODE_CHILD_SESSION= bonsai unlink --yes", "env -u CLAUDE_CODE_CHILD_SESSION ./link.sh",
			"env -i PATH=/usr/bin ./script.sh", "export CLAUDE_CODE_CHILD_SESSION=", "bash -c 'unset CLAUDE_CODE_CHILD_SESSION'",
			"sudo env -i ./x.sh", `bash -c "env - ./x.sh"`, "env -u HOME --ignore-env ./x.sh", "exec -c ./x.sh", "X=1 /usr/bin/env -vi ./x.sh"} {
			wantRule(t, dir, bashIn(dir, "Bash", line), RulePersonCommand)
		}
		for _, line := range []string{"$env:CLAUDE_CODE_CHILD_SESSION = ''", "Remove-Item Env:CLAUDE_CODE_CHILD_SESSION",
			"[Environment]::SetEnvironmentVariable('CLAUDE_CODE_CHILD_SESSION', $null)", "Start-Process bonsai -UseNewEnvironment"} {
			wantRule(t, dir, bashIn(dir, "PowerShell", line), RulePersonCommand)
		}
		// Under a person's grant in a project in command mode (the second question's (ii)): allowed.
		commandMode(t, dir)
		j = judgeIn(dir)
		for _, argv := range [][]string{{"bonsai", "update", "--allow-exec", "--yes"}, {"bonsai", "unlink", "--yes"}} {
			if d := j.ownCommand(dir, argv, true); !d.Allow {
				t.Errorf("%q under a person's grant in command mode: %+v", argv, d)
			}
		}
		wantRule(t, dir, bashIn(dir, "Bash", "unset CLAUDE_CODE_CHILD_SESSION"), RuleShell)
		// In command mode with no such grant: refused, the reason naming the grant missing.
		writeFile(t, filepath.Join(dir, "work", "tasks", "T-0001-the-task.md"), taskText("T-0001", "running", "light", "bonsai.yaml"))
		j = judgeIn(dir)
		if d := j.ownCommand(dir, []string{"bonsai", "unlink", "--yes"}, true); d.Rule != RulePersonCommand ||
			!strings.Contains(d.Why, "for .bonsai/lock.json: the running task T-0001 does not grant it") {
			t.Errorf("unlink with bonsai.yaml alone granted: %+v", d)
		}
	},
	RuleHomeCommand: func(t *testing.T) {
		dir := withTask(t, "running", "**")
		commandMode(t, dir)
		j := judgeIn(dir)
		for _, argv := range [][]string{{"bonsai", "settings", "set", "status_writes=agents"}, {"bonsai", "settings", "--machine", "set", "x=1"},
			{"bonsai", "labels", "attach", "acme.yaml"}, {"bonsai.exe", "labels", "detach", "acme"}} {
			if d := j.ownCommand(dir, argv, true); d.Rule != RuleHomeCommand || d.Why == "" || d.Next == "" {
				t.Errorf("%q: %+v", argv, d)
			}
		}
		if d := j.ownCommand(dir, []string{"bonsai", "settings", "show", "--json"}, true); !d.Allow {
			t.Errorf("settings show: %+v", d)
		}
	},
	RuleBonsaiCommand: func(t *testing.T) {
		dir := withTask(t, "todo")
		j := judgeIn(dir)
		for _, argv := range [][]string{{"bonsai", "status", "--json"}, {"bonsai", "check", "--write"}, {"bonsai", "hook", "guard"},
			{"bonsai"}, {"bonsai", "--version"}} {
			if d := j.ownCommand(dir, argv, true); !d.Allow || d.Rule != RuleBonsaiCommand {
				t.Errorf("%q: %+v", argv, d)
			}
		}
		// update and init are judged as an edit of bonsai.yaml and the lock: no grant, refused; granted, allowed.
		if d := j.ownCommand(filepath.Join(dir, "work"), []string{"bonsai", "update", "--yes"}, true); d.Rule != RulePersonOnly ||
			!strings.HasPrefix(d.Why, "bonsai update writes bonsai.yaml and .bonsai/lock.json, so it is judged as an edit of them") {
			t.Errorf("update with no grant: %+v", d)
		}
		dir = withTask(t, "running", "bonsai.yaml", ".bonsai/lock.json")
		if d := judgeIn(dir).ownCommand(dir, []string{"bonsai", "update", "--yes"}, true); !d.Allow || d.Rule != RuleBonsaiCommand {
			t.Errorf("update under a grant: %+v", d)
		}
		// init in a folder no linked project holds: nothing to judge.
		fresh := t.TempDir()
		if d := judgeIn(dir).ownCommand(fresh, []string{"bonsai", "init", "--yes"}, true); !d.Allow {
			t.Errorf("init outside every project: %+v", d)
		}
	},
	RuleShell: func(t *testing.T) {
		dir := withTask(t, "todo")
		for _, c := range []struct{ tool, line string }{{"Bash", "ls -la"}, {"PowerShell", "Get-ChildItem"}, {"Monitor", "tail -f x.log"},
			{"Bash", "echo $CLAUDE_CODE_CHILD_SESSION"}, {"Bash", "grep -rn CLAUDE_CODE_CHILD_SESSION internal/"},
			{"PowerShell", "if ($env:CLAUDE_CODE_CHILD_SESSION) { 'agent' }"}, {"Bash", "printenv CLAUDE_CODE_CHILD_SESSION"},
			{"PowerShell", "Get-ChildItem Env:CLAUDE_CODE_CHILD_SESSION"}, {"Bash", "env | grep CLAUDE_CODE_CHILD_SESSION"},
			{"Bash", "grep env -i notes.txt"}, {"Bash", "env -u HOME ./x.sh"}, {"Bash", "env X=1 -- ./x.sh -i"}} {
			wantRule(t, dir, bashIn(dir, c.tool, c.line), RuleShell)
		}
	},
	RuleOtherTool: func(t *testing.T) {
		dir := withTask(t, "todo")
		wantRule(t, dir, toolPayload("r1", dir, "Read", schema.Object{{Key: "file_path", Value: filepath.Join(dir, "keys", "a.key")}}), RuleOtherTool)
	},
	RuleBadInput:  TestBlocksWhatItCannotReadOrDecide,
	RuleBadGlob:   TestBadGlobBlocks,
	RuleBadPath:   ruleBadPath,
	RuleBadLock:   TestGuardReadOfTheLockIsLean,
	RuleBadConfig: TestBadConfigBlocksWithNoRecord,
	RuleNoProject: TestBlocksWhatItCannotReadOrDecide,
	RuleLogFailed: TestAnAllowItCannotRecordIsBlocked,
	RuleOverTime:  TestOverTimeBlocks,
	RuleInternal:  TestBlocksWhatItCannotReadOrDecide,
	RuleNotLinked: TestNotLinkedAllowsAndRecordsNothing,
}

func TestEveryRule(t *testing.T) {
	names := map[string]bool{}
	kebab := regexp.MustCompile(`^[a-z]+(-[a-z0-9]+)*$`)
	for _, r := range Rules {
		if names[r.Name] {
			t.Errorf("rule %s is in the table twice", r.Name)
		}
		names[r.Name] = true
		if !kebab.MatchString(r.Name) || r.Line == "" || strings.TrimSpace(r.Line) != r.Line {
			t.Errorf("rule %q: not kebab-case, or no line", r.Name)
		}
		asciiOnly(t, r.Line)
	}
	var extra []string
	for name := range ruleTests {
		if !names[name] {
			extra = append(extra, name)
		}
	}
	sort.Strings(extra)
	if len(extra) > 0 {
		t.Errorf("tests for rules the table does not hold: %v", extra)
	}
	for _, r := range Rules {
		test, ok := ruleTests[r.Name]
		if !ok {
			t.Errorf("rule %s has no test", r.Name)
			continue
		}
		t.Run(r.Name, test)
	}
}

// Every rule name the guard's code gives a decision is the table's: no rule outside it.
func TestEveryRuleInTheCodeIsInTheTable(t *testing.T) {
	names := map[string]bool{}
	for _, r := range Rules {
		names[r.Name] = true
	}
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	decl := regexp.MustCompile(`(?m)^\s*(Rule[A-Za-z]+)\s*=\s*"([^"]+)"`)
	found := 0
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		raw, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		for _, m := range decl.FindAllStringSubmatch(string(raw), -1) {
			found++
			if !names[m[2]] {
				t.Errorf("%s: %s = %q is not in the rules table", f, m[1], m[2])
			}
		}
	}
	if found != len(Rules) {
		t.Errorf("%d rule names declared, %d rows in the table", found, len(Rules))
	}
}
