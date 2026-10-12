package rung0

// The guard and rung 0 agree (design/plan-5.md, 5.4.3 note 7; "5.4 done" check 7): a table of paths, each judged by
// the guard, through its own hook entry (guard.Main, a Write or a shell rm of the path in a session in the checkout,
// BONSAI_TASK naming the running task, before the change is made), and by rung 0 after the change (committed on the
// task worktree's branch, or left in the main checkout's working tree on the base), with the same answer: allowed or
// refused, and for a file tool's edit the same rule (a shell delete is the guard's delete check, whose rule is its
// own, delete-protected, so only the answer is compared). In agents mode and in command mode.
//
// What rung 0 does not share with the guard by design is left out here and tested in rung0_test.go: a table on the
// base, a .bonsai/local/ file (bonsai check's), and a task at verify (rung 0 judges finished work).

import (
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/guard"
	"github.com/LastStep/Bonsai/internal/hookin"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// row is one path of the table: what the change leaves there, or its delete.
type row struct {
	path  string
	after string
	del   bool
	where string // "" for both places; "worktree" or "main" for one
}

// stateMD is a STATE that reads; badState one that does not.
const (
	stateMD  = "---\nformat: bonsai.state/1\nupdated: 2026-10-13\nupdated_by: a test\nlabels: {}\n---\n\nWhere it stands.\n"
	badState = "---\nformat: bonsai.state/1\nupdated: [\n---\n"
	runMD    = "---\nformat: bonsai.run/1\nid: R-2026-10-13-T-0901\ntask: T-0901\nrole: builder\nmodel: a model\nstarted:\nfinished:\n" +
		"outcome: running\ncommits: []\nlabels: {}\n---\n\n# A run\n"
)

func agreeRows() []row {
	return []row{
		{path: "src/main.txt", after: "changed\n"},
		{path: "src/new.txt", after: "new\n"},
		{path: "assets/t-0901/new.txt", after: "granted\n"},
		{path: "assets/old/a.txt", after: "changed\n"},
		{path: "assets/new/b.txt", after: "new\n"},
		{path: "secret/granted/k.txt", after: "granted, person-only\n"},
		{path: "secret/other/k.txt", after: "person-only\n"},
		{path: ".claude/agents/x.md", after: "granted, person-only\n"},
		{path: ".claude/settings.json", after: "{}\n"},
		{path: ".claude/settings.local.json", after: "{}\n"},
		{path: "bonsai.yaml", after: projectYAML + "# a comment\n", where: "main"},
		{path: workspace.LockFile, after: lockJSON + "\n"},
		{path: ".bonsai/other.txt", after: "x\n"},
		{path: workspace.StateFile, after: badState},
		{path: workspace.TasksTableFile, after: "---\nformat: bonsai.tasks/1\n---\n", where: "worktree"},
		{path: workspace.SessionsTableFile, after: "---\nformat: bonsai.sessions/1\n---\n", where: "worktree"},
		{path: workspace.GitignoreFile, after: workspace.GitignoreText + "# more\n"},
		{path: "docs/bonsai.yaml", after: strings.Replace(projectYAML, testID, "ws-d0d0d0d0d0d0d0d0d0d0d0d0d0", 1)},
		{path: "docs/.bonsai/x.txt", after: "x\n"},
		{path: "work/tasks/T-0902-todo.md", after: taskMD("T-0902", "todo", "") + "More words.\n"},
		{path: "work/tasks/T-0903-todo.md", after: taskMD("T-0903", "plan", "")},
		{path: "work/tasks/T-0904-new.md", after: taskMD("T-0904", "running", "")},
		{path: "work/tasks/T-0905-todo.md", after: strings.Replace(taskMD("T-0905", "todo", ""), "lane: light", "lane: full", 1)},
		{path: "work/tasks/T-0906-new.md", after: taskMD("T-0906", "todo", `["secret/**"]`)},
		{path: "work/runs/R-2026-10-13-T-0901.md", after: runMD},
		{path: "src/old.txt", del: true},
		{path: "assets/old/c.txt", del: true},
		{path: "bonsai.yaml", del: true, where: "worktree"}, // in the main checkout it would leave no lists to judge by
	}
}

func TestTheGuardAndRung0Agree(t *testing.T) {
	for _, mode := range []string{"agents", "command"} {
		for _, place := range []string{"worktree", "main"} {
			t.Run(mode+"/"+place, func(t *testing.T) { agree(t, mode, place) })
		}
	}
}

func agree(t *testing.T, mode, place string) {
	f := newProject(t, map[string]string{
		workspace.StateFile:         stateMD,
		"work/tasks/T-0903-todo.md": taskMD("T-0903", "todo", ""),
		"work/tasks/T-0905-todo.md": taskMD("T-0905", "todo", ""),
		"src/old.txt":               "old\n",
		"assets/old/c.txt":          "old asset\n",
	})
	if mode == "command" {
		md, err := workspace.MachineDir(f.home, f.main)
		if err != nil {
			t.Fatal(err)
		}
		write(t, md, workspace.MachineSettingsFile, `{"status_writes": "command", "status_command": "studio-move"}`)
	}
	root := f.main
	if place == "worktree" {
		root = f.worktree(t)
	}
	var rows []row
	for _, r := range agreeRows() {
		if r.where == "" || r.where == place {
			rows = append(rows, r)
		}
	}
	// The guard, before each change is made.
	type answer struct {
		allow bool
		rule  string
	}
	byGuard := map[int]answer{}
	for i, r := range rows {
		session := "agree" + strconv.Itoa(i)
		var in string
		if r.del {
			in = payload(session, root, "Bash", schema.Object{{Key: "command", Value: "rm " + r.path}})
		} else {
			in = payload(session, root, "Write", schema.Object{{Key: "file_path", Value: filepath.Join(root, filepath.FromSlash(r.path))},
				{Key: "content", Value: r.after}})
		}
		var stderr bytes.Buffer
		code := guard.Main(guard.Options{Stdin: strings.NewReader(in), Stderr: &stderr, Getenv: func(k string) string {
			switch k {
			case hookin.ProjectEnv:
				return root
			case workspace.TaskEnv:
				return "T-0901"
			}
			return ""
		}, Budget: time.Minute, Now: func() time.Time { return since }})
		if code != 0 && code != 2 {
			t.Fatalf("%s: the guard exits %d: %s", r.path, code, stderr.String())
		}
		byGuard[i] = answer{allow: code == 0, rule: guardRule(t, f.main, session)}
	}
	// The changes, then rung 0.
	for _, r := range rows {
		if r.del {
			remove(t, root, r.path)
		} else {
			write(t, root, r.path, r.after)
		}
	}
	if place == "worktree" {
		commit(t, root, "the table's changes")
	}
	o := f.climb(t, root, "T-0901")
	if o.Status == "error" || o.More != 0 {
		t.Fatalf("rung 0: %s", o.Reason)
	}
	refused := refusedRules(o)
	seen := map[string]bool{}
	for i, r := range rows {
		g := byGuard[i]
		rule, no := refused[r.path]
		switch {
		case g.allow == no:
			t.Errorf("%s (delete %v): the guard allows %v (%s), rung 0 refuses %v (%s)", r.path, r.del, g.allow, g.rule, no, rule)
		case !g.allow && !r.del && g.rule != rule:
			t.Errorf("%s: the guard refuses as %s, rung 0 as %s", r.path, g.rule, rule)
		}
		if testing.Verbose() {
			t.Logf("%-34s del=%-5v guard: %-5v %-18s rung 0: %v %s", r.path, r.del, g.allow, g.rule, !no, rule)
		}
		if !g.allow {
			seen[g.rule] = true
		} else {
			seen["allowed"] = true
		}
	}
	// The table means something: each answer it is for shows up.
	want := []string{"allowed", guard.RuleProtected, guard.RulePersonOnly, guard.RuleNeverGranted, guard.RuleBonsaiFiles, guard.RuleDeleteProtected}
	switch {
	case mode == "command" && place == "worktree":
		want = append(want, guard.RuleTaskOutsideMain, guard.RuleBadFormat)
	case mode == "command":
		want = append(want, guard.RuleStatusLine, guard.RuleLaneLine, guard.RuleGrantPersonOnly, guard.RuleBadFormat)
	}
	for _, w := range want {
		if !seen[w] {
			t.Errorf("no row of the table meets %s (seen %v)", w, seen)
		}
	}
}

// payload is a PreToolUse payload for tool with its tool_input, as Claude Code sends it.
func payload(session, cwd, tool string, input schema.Object) string {
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

// guardRule reads the rule of a session's one guard record in the main checkout's log.
func guardRule(t *testing.T, main, session string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(main, filepath.FromSlash(guard.LogDir), "s-"+session+".ndjson"))
	if err != nil {
		t.Fatalf("the guard's record of %s: %v", session, err)
	}
	lines := strings.Split(strings.TrimSpace(string(raw)), "\n")
	v, err := schema.Decode([]byte(lines[len(lines)-1]))
	if err != nil {
		t.Fatal(err)
	}
	o, _ := v.(schema.Object)
	return o.String("rule")
}
