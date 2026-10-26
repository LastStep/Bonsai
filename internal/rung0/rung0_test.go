package rung0

// Rung 0 on fixture repositories, one per refusal and per thing it does not refuse (design/plan-5.md, 5.4.3 note 7;
// "5.4 done" check 7): a protected path granted, not granted, and granted while the task reads verify; the floor's
// never-granted paths under a grant; a table on a branch and on the base; a staged .bonsai/local/ file; a bonsai check
// finding counted and a machine one not; hooks-off in the checkout's and the main checkout's local settings; a branch
// deleting bonsai.yaml; a worktree with no bonsai.yaml of its own; no task named; no base; command mode's tripwires;
// the cap on the paths listed; git that cannot be read.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/engine"
	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/guard"
	"github.com/LastStep/Bonsai/internal/testpack"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// green fails the test unless the outcome is green with checked paths and no reason but notes.
func green(t *testing.T, o Outcome, checked int) {
	t.Helper()
	if o.Status != "green" || len(o.Refused) != 0 {
		t.Fatalf("want green, got %s: %s", o.Status, o.Reason)
	}
	if got := o.Captures.String("checked"); got != "" {
		t.Fatalf("checked is text: %q", got)
	}
	n, _ := o.Captures.Get("checked")
	if n != json.Number(strconv.Itoa(checked)) {
		t.Errorf("checked %v, want %d", n, checked)
	}
	for _, l := range strings.Split(o.Reason, "\n") {
		if l != "" && !strings.HasPrefix(l, "not judged here: ") {
			t.Errorf("a green rung's reason says more than what was not judged: %q", l)
		}
	}
	notJudgedKept(t, o)
}

// notJudgedKept fails the test unless the not_judged capture holds what the reason's "not judged here" lines say,
// joined by "; ", and is null when there are none: a green rung's reason is not in the result, its captures are.
func notJudgedKept(t *testing.T, o Outcome) {
	t.Helper()
	var items []string
	for _, l := range strings.Split(o.Reason, "\n") {
		if rest, ok := strings.CutPrefix(l, "not judged here: "); ok {
			items = append(items, rest)
		}
	}
	got, ok := o.Captures.Get("not_judged")
	switch {
	case !ok:
		t.Errorf("no not_judged capture: %v", o.Captures)
	case len(items) == 0 && got != nil:
		t.Errorf("not_judged %v, want null: %q", got, o.Reason)
	case len(items) > 0 && got != strings.Join(items, "; "):
		t.Errorf("not_judged %v, want %q", got, strings.Join(items, "; "))
	}
}

// red fails the test unless the outcome is red and its reason holds every one of words.
func red(t *testing.T, o Outcome, words ...string) {
	t.Helper()
	if o.Status != "red" {
		t.Fatalf("want red, got %s: %s", o.Status, o.Reason)
	}
	notJudgedKept(t, o)
	for _, w := range words {
		if !strings.Contains(o.Reason, w) {
			t.Errorf("the reason lacks %q:\n%s", w, o.Reason)
		}
	}
}

func TestStatusesAreKnownWords(t *testing.T) {
	for _, s := range []string{"green", "red", "error"} {
		found := false
		for _, w := range format.RungStatuses {
			found = found || w.Word == s
		}
		if !found {
			t.Errorf("%s is not in format.RungStatuses", s)
		}
	}
}

func TestCleanAndUnprotected(t *testing.T) {
	f := newProject(t, nil)
	green(t, f.climb(t, f.main, "T-0901"), 0)
	wt := f.worktree(t)
	green(t, f.climb(t, wt, "T-0901"), 0)
	write(t, wt, "src/main.txt", "changed\n")
	write(t, wt, "src/new.txt", "new\n")
	commit(t, wt, "code")
	write(t, wt, "src/untracked.txt", "untracked\n")
	write(t, wt, "assets/t-0901/granted.txt", "granted\n")
	o := f.climb(t, wt, "T-0901")
	green(t, o, 4)
	if o.Reason != "" {
		t.Errorf("a green rung with everything judged has a reason: %s", o.Reason)
	}
}

func TestProtectedGrantedOrNot(t *testing.T) {
	t.Run("granted", func(t *testing.T) {
		f := newProject(t, nil)
		wt := f.worktree(t)
		write(t, wt, "assets/t-0901/new.txt", "x\n")
		write(t, wt, "secret/granted/key.txt", "x\n") // person-only, and granted: agents mode counts it as written
		commit(t, wt, "granted")
		green(t, f.climb(t, wt, "T-0901"), 2)
	})
	t.Run("not granted", func(t *testing.T) {
		f := newProject(t, nil)
		wt := f.worktree(t)
		write(t, wt, "assets/old/a.txt", "changed\n")
		write(t, wt, "secret/other/key.txt", "x\n")
		commit(t, wt, "not granted")
		o := f.climb(t, wt, "T-0901")
		red(t, o, `"assets/old/a.txt" (protected)`, `"secret/other/key.txt" (person-only)`, "the named task T-0901 does not grant it",
			"2 of 2 changed paths refused")
		if want := map[string]string{"assets/old/a.txt": guard.RuleProtected, "secret/other/key.txt": guard.RulePersonOnly}; !reflect.DeepEqual(refusedRules(o), want) {
			t.Errorf("refused %v, want %v", refusedRules(o), want)
		}
		for _, r := range o.Refused {
			if !strings.Contains(r.Needs, "T-0901's bonsai.allows, or a revert") {
				t.Errorf("%s needs %q", r.Path, r.Needs)
			}
		}
	})
	t.Run("granted while the task reads verify", func(t *testing.T) {
		f := newProject(t, map[string]string{
			"work/tasks/T-0901-the-task.md": taskMD("T-0901", "verify", `["assets/t-0901/**"]`)})
		wt := f.worktree(t)
		write(t, wt, "assets/t-0901/new.txt", "x\n")
		commit(t, wt, "granted at verify")
		green(t, f.climb(t, wt, "T-0901"), 1)
		// The guard grants nothing at verify; rung 0 judges finished work with the named task's grants.
		write(t, wt, "assets/old/a.txt", "changed\n")
		red(t, f.climb(t, wt, "T-0901"), `"assets/old/a.txt" (protected)`)
	})
	t.Run("a bonsai.allows that is not a list of text", func(t *testing.T) {
		f := newProject(t, map[string]string{"work/tasks/T-0901-the-task.md": taskMD("T-0901", "running", `"assets/t-0901/**"`)})
		wt := f.worktree(t)
		write(t, wt, "assets/t-0901/new.txt", "x\n")
		red(t, f.climb(t, wt, "T-0901"), "bonsai.allows is not a list of text, so it grants nothing")
	})
}

func TestNeverGrantedUnderAGrant(t *testing.T) {
	f := newProject(t, nil) // T-0901 grants .claude/**
	wt := f.worktree(t)
	write(t, wt, ".claude/settings.json", "{}\n")
	write(t, wt, ".claude/settings.local.json", "{}\n")
	write(t, wt, ".claude/agents/builder.md", "granted\n")
	commit(t, wt, "settings")
	o := f.climb(t, wt, "T-0901")
	want := map[string]string{".claude/settings.json": guard.RuleNeverGranted, ".claude/settings.local.json": guard.RuleNeverGranted}
	if !reflect.DeepEqual(refusedRules(o), want) {
		t.Fatalf("refused %v, want %v:\n%s", refusedRules(o), want, o.Reason)
	}
	red(t, o, "and no grant opens it")
	for _, r := range o.Refused {
		if r.Needs != "a revert: no grant opens it" {
			t.Errorf("%s needs %q", r.Path, r.Needs)
		}
	}
}

func TestTables(t *testing.T) {
	for _, table := range []string{workspace.TasksTableFile, workspace.SessionsTableFile} {
		t.Run(table+" on a branch", func(t *testing.T) {
			f := newProject(t, nil)
			wt := f.worktree(t)
			write(t, wt, table, "---\nformat: bonsai.tasks/1\n---\n")
			red(t, f.climb(t, wt, "T-0901"), strconv.Quote(table)+" (bonsai-files)")
			commit(t, wt, "a table on a branch")
			red(t, f.climb(t, wt, "T-0901"), strconv.Quote(table)+" (bonsai-files)")
		})
		t.Run(table+" on the base", func(t *testing.T) {
			f := newProject(t, nil)
			write(t, f.main, table, "changed\n")
			o := f.climb(t, f.main, "T-0901")
			if len(o.Refused) != 0 {
				t.Fatalf("a table on the base is refused: %s", o.Reason)
			}
			commit(t, f.main, "the table, committed on the base")
			if o := f.climb(t, f.main, "T-0901"); len(o.Refused) != 0 {
				t.Fatalf("a committed table on the base is refused: %s", o.Reason)
			}
		})
	}
	t.Run("on a detached HEAD in the main checkout", func(t *testing.T) {
		f := newProject(t, nil)
		testpack.Git(t, f.main, "checkout", "-q", "--detach")
		write(t, f.main, workspace.TasksTableFile, "changed\n")
		// The runner finds no base: not the base, so the table is judged (fail closed).
		cfg, _ := workspace.LoadConfigFull(f.main)
		a, _ := workspace.Active(workspace.ActiveInput{Main: f.main, TaskDir: cfg.Full.Documents.Task, Task: "T-0901"})
		o := Judge(Input{Root: f.main, Main: f.main, Home: f.home, Active: a})
		red(t, o, strconv.Quote(workspace.TasksTableFile)+" (bonsai-files)", "not judged here: the branch's commits (no base branch")
	})
}

func TestLocalFiles(t *testing.T) {
	t.Run("staged", func(t *testing.T) {
		f := newProject(t, nil)
		wt := f.worktree(t)
		write(t, wt, ".bonsai/local/log/w-2026-10-13.ndjson", "{}\n")
		testpack.Git(t, wt, "add", "-f", ".bonsai/local/log/w-2026-10-13.ndjson")
		o := f.climb(t, wt, "T-0901")
		red(t, o, "bonsai check finds 1 problem(s) about the commit: local (.bonsai/local)")
		if len(o.Refused) != 0 {
			t.Errorf("a local/ file is reported twice: %v", o.Refused)
		}
		commit(t, wt, "committed")
		red(t, f.climb(t, wt, "T-0901"), "local (.bonsai/local)")
	})
	t.Run("ignored, as .bonsai/.gitignore keeps it", func(t *testing.T) {
		f := newProject(t, nil)
		write(t, f.main, ".bonsai/local/ladder/T-0901.json", "{}\n")
		green(t, f.climb(t, f.main, "T-0901"), 0)
	})
	t.Run("a branch taking one out of git", func(t *testing.T) {
		f := newProject(t, nil)
		write(t, f.main, ".bonsai/local/asks/a.ndjson", "{}\n")
		testpack.Git(t, f.main, "add", "-f", ".bonsai/local/asks/a.ndjson")
		commit(t, f.main, "a tracked local file on the base")
		red(t, f.climb(t, f.main, "T-0901"), "local (.bonsai/local)")
		wt := f.worktree(t)
		testpack.Git(t, wt, "rm", "-q", "--cached", ".bonsai/local/asks/a.ndjson")
		commit(t, wt, "untracked")
		green(t, f.climb(t, wt, "T-0901"), 1)
	})
}

func TestCheckFindings(t *testing.T) {
	t.Run("a finding about the commit counts", func(t *testing.T) {
		f := newProject(t, nil)
		wt := f.worktree(t)
		write(t, wt, "work/runs/R-2026-10-13-T-0901.md", "---\nformat: bonsai.run/1\nid: [\n---\n")
		commit(t, wt, "a run report that does not read")
		o := f.climb(t, wt, "T-0901")
		red(t, o, "bonsai check finds 1 problem(s) about the commit: document (work/runs/R-2026-10-13-T-0901.md)",
			"next: run bonsai check")
		if len(o.Refused) != 0 {
			t.Errorf("refused %v", o.Refused)
		}
	})
	t.Run("a finding about this machine does not", func(t *testing.T) {
		f := newProject(t, nil)
		write(t, f.home, "install.json", "not the installed bonsai's record\n")
		o := f.climb(t, f.main, "T-0901")
		green(t, o, 0)
		if !strings.Contains(o.Reason, "not judged here: bonsai check's findings about this machine, not the commit (bonsai-path)") {
			t.Errorf("the machine finding is not listed as not judged: %q", o.Reason)
		}
		// On a base with no upstream (BUG-124), the base's commits come first.
		if got := o.Captures.String("not_judged"); got != `the base branch's commits (no upstream for "main", so only what `+
			"the working tree changed was judged); bonsai check's findings about this machine, not the commit (bonsai-path)" {
			t.Errorf("not_judged %q", got)
		}
	})
	t.Run("warnings never count", func(t *testing.T) {
		f := newProject(t, nil) // no generated table yet: bonsai check's tables warning
		res, err := engine.CheckWith(f.main, f.home, engine.CheckOptions{History: true})
		if err != nil || len(res.Warnings) == 0 || len(res.Findings) != 0 {
			t.Fatalf("the fixture: %v, %+v", err, res)
		}
		green(t, f.climb(t, f.main, "T-0901"), 0)
	})
}

func TestHooksOff(t *testing.T) {
	const off = "{\n  \"disableAllHooks\": true\n}\n"
	t.Run("the checkout's local settings", func(t *testing.T) {
		f := newProject(t, nil)
		wt := f.worktree(t)
		write(t, wt, engine0LocalSettings, off)
		red(t, f.climb(t, wt, "T-0901"), "hooks-off (.claude/settings.local.json)")
	})
	t.Run("the main checkout's local settings", func(t *testing.T) {
		f := newProject(t, nil)
		wt := f.worktree(t)
		write(t, f.main, engine0LocalSettings, off)
		red(t, f.climb(t, wt, "T-0901"), "hooks-off (.claude/settings.local.json)")
	})
	t.Run("a verified main other than the one git names", func(t *testing.T) {
		// bonsai check reads the main checkout git names; rung 0 reads the verified one too when they differ.
		f := newProject(t, nil)
		other := filepath.Join(f.tmp, "verified-main")
		write(t, other, engine0LocalSettings, off)
		found, _, cannot := checkFindings(Input{Root: f.main, Main: other, Home: f.home})
		if cannot != "" || !reflect.DeepEqual(found, []string{"hooks-off (the main checkout's .claude/settings.local.json)"}) {
			t.Errorf("found %v (%q)", found, cannot)
		}
		write(t, other, engine0LocalSettings, "{\"disableAllHooks\": false}\n")
		if found, _, _ := checkFindings(Input{Root: f.main, Main: other, Home: f.home}); len(found) != 0 {
			t.Errorf("found %v", found)
		}
	})
}

const engine0LocalSettings = ".claude/settings.local.json"

func TestBonsaiYAMLGone(t *testing.T) {
	t.Run("a branch deleting it", func(t *testing.T) {
		f := newProject(t, nil)
		wt := f.worktree(t)
		remove(t, wt, "bonsai.yaml")
		commit(t, wt, "delete bonsai.yaml")
		o := f.climb(t, wt, "T-0901")
		red(t, o, `"bonsai.yaml" (person-only)`, "bonsai check cannot run in this checkout")
		if refusedRules(o)["bonsai.yaml"] != guard.RulePersonOnly {
			t.Errorf("refused %v", refusedRules(o))
		}
	})
	t.Run("a worktree with no bonsai.yaml of its own", func(t *testing.T) {
		// A branch from before the link: its commits never held bonsai.yaml. Its changes are judged against the
		// verified main checkout's lists all the same (the guard there judges them by none: its known limit).
		tmp, err := filepath.EvalSymlinks(t.TempDir())
		if err != nil {
			t.Fatal(err)
		}
		f := &fx{tmp: tmp, home: testpack.Isolate(t, tmp), main: filepath.Join(tmp, "main")}
		write(t, f.main, "src/main.txt", "before the link\n")
		write(t, f.main, "assets/old/a.txt", "an asset\n")
		testpack.Git(t, f.main, "init", "-q")
		commit(t, f.main, "before the link")
		testpack.Git(t, f.main, "branch", "old")
		write(t, f.main, "bonsai.yaml", projectYAML)
		write(t, f.main, workspace.LockFile, lockJSON)
		write(t, f.main, workspace.GitignoreFile, workspace.GitignoreText)
		write(t, f.main, "work/tasks/T-0901-the-task.md", taskMD("T-0901", "running", `["assets/t-0901/**"]`))
		commit(t, f.main, "the link")
		if _, err := workspace.RecordCheckout(f.home, f.main, testID, since); err != nil {
			t.Fatal(err)
		}
		f.wt = filepath.Join(tmp, "wt")
		testpack.Git(t, f.main, "worktree", "add", "-q", f.wt, "old")
		write(t, f.wt, "assets/old/a.txt", "changed\n")
		write(t, f.wt, ".claude/settings.json", "{}\n")
		write(t, f.wt, "assets/t-0901/ok.txt", "granted\n")
		commit(t, f.wt, "work on an old branch")
		o := f.climb(t, f.wt, "T-0901")
		want := map[string]string{"assets/old/a.txt": guard.RuleProtected, ".claude/settings.json": guard.RuleNeverGranted}
		if !reflect.DeepEqual(refusedRules(o), want) {
			t.Fatalf("refused %v, want %v:\n%s", refusedRules(o), want, o.Reason)
		}
		red(t, o, "bonsai check cannot run in this checkout", "rebase the branch on the base branch")
	})
}

func TestNoTaskNamed(t *testing.T) {
	f := newProject(t, nil)
	wt := f.worktree(t)
	write(t, wt, "assets/t-0901/new.txt", "x\n")
	write(t, wt, "src/new.txt", "x\n")
	// T-0901 reads running, but nothing names it: it grants nothing here (--ci with no task).
	o := f.climb(t, wt, "")
	red(t, o, `"assets/t-0901/new.txt" (protected)`, "no task is named (T-0901 reads running", "no named task (nothing granted)")
	if len(o.Refused) != 1 || !strings.Contains(o.Refused[0].Needs, "a named task (--task)") {
		t.Errorf("refused %+v", o.Refused)
	}
	// None at all: the function's reason.
	remove(t, f.main, "work/tasks/T-0901-the-task.md")
	red(t, f.climb(t, wt, ""), "no active task: no task reads running")
}

func TestNoBase(t *testing.T) {
	f := newProject(t, nil)
	wt := f.worktree(t)
	write(t, wt, "assets/old/a.txt", "committed\n")
	commit(t, wt, "a protected change, committed")
	write(t, wt, "assets/old/b.txt", "uncommitted\n")
	cfg, _ := workspace.LoadConfigFull(f.main)
	a, _ := workspace.Active(workspace.ActiveInput{Main: f.main, TaskDir: cfg.Full.Documents.Task, Task: "T-0901"})
	for _, in := range []Input{
		{Root: wt, Main: f.main, Home: f.home, Active: a},                              // the main checkout's HEAD detached
		{Root: wt, Main: f.main, Home: f.home, Active: a, Base: "main", MergeBase: ""}, // no merge base
	} {
		o := Judge(in)
		if want := map[string]string{"assets/old/b.txt": guard.RuleProtected}; !reflect.DeepEqual(refusedRules(o), want) {
			t.Errorf("refused %v, want %v", refusedRules(o), want)
		}
		red(t, o, "not judged here: the branch's commits (no ")
	}
	// With the base, the committed one is judged too.
	if o := f.climb(t, wt, "T-0901"); len(o.Refused) != 2 {
		t.Errorf("with the base: %s", o.Reason)
	}
}

func TestRenameJudgesBothSides(t *testing.T) {
	f := newProject(t, nil)
	wt := f.worktree(t)
	testpack.Git(t, wt, "mv", "assets/old/a.txt", "src/a.txt")
	commit(t, wt, "a rename out of a protected folder")
	o := f.climb(t, wt, "T-0901")
	if want := map[string]string{"assets/old/a.txt": guard.RuleProtected}; !reflect.DeepEqual(refusedRules(o), want) {
		t.Errorf("refused %v, want %v", refusedRules(o), want)
	}
	n, _ := o.Captures.Get("checked")
	if n != json.Number("2") {
		t.Errorf("checked %v, want both sides of the rename", n)
	}
}

func TestCommandMode(t *testing.T) {
	setup := func(t *testing.T) *fx {
		f := newProject(t, nil)
		md, err := workspace.MachineDir(f.home, f.main)
		if err != nil {
			t.Fatal(err)
		}
		write(t, md, workspace.MachineSettingsFile, `{"status_writes": "command", "status_command": "studio-move"}`)
		return f
	}
	// The main checkout is on the base: what its working tree changed is judged against HEAD's bytes, though the
	// file on disk already holds the change (the guard would read "before" there).
	t.Run("on the base: the status line against HEAD", func(t *testing.T) {
		f := setup(t)
		write(t, f.main, "work/tasks/T-0902-todo.md", taskMD("T-0902", "plan", ""))
		o := f.climb(t, f.main, "T-0901")
		if want := map[string]string{"work/tasks/T-0902-todo.md": guard.RuleStatusLine}; !reflect.DeepEqual(refusedRules(o), want) {
			t.Fatalf("refused %v, want %v:\n%s", refusedRules(o), want, o.Reason)
		}
		red(t, o, `it changes the document's status line ("todo" to "plan")`)
		if o.Refused[0].Needs != "a revert: in command mode the workspace's status command makes this change" {
			t.Errorf("needs %q", o.Refused[0].Needs)
		}
	})
	t.Run("on the base: a body edit passes", func(t *testing.T) {
		f := setup(t)
		write(t, f.main, "work/tasks/T-0902-todo.md", taskMD("T-0902", "todo", "")+"More words.\n")
		green(t, f.climb(t, f.main, "T-0901"), 1)
	})
	t.Run("on the base: a new task at a later status, a lane, a person-only grant, a broken STATE, a deleted task", func(t *testing.T) {
		f := setup(t)
		write(t, f.main, "work/tasks/T-0905-gone.md", taskMD("T-0905", "todo", ""))
		commit(t, f.main, "T-0905")
		remove(t, f.main, "work/tasks/T-0905-gone.md")
		write(t, f.main, "work/tasks/T-0903-new.md", taskMD("T-0903", "running", ""))
		write(t, f.main, "work/tasks/T-0902-todo.md", strings.Replace(taskMD("T-0902", "todo", ""), "lane: light", "lane: full", 1))
		write(t, f.main, "work/tasks/T-0904-new.md", taskMD("T-0904", "todo", `["secret/**"]`))
		write(t, f.main, workspace.StateFile, "---\nformat: bonsai.state/1\nupdated: [\n---\n")
		o := f.climb(t, f.main, "T-0901")
		want := map[string]string{"work/tasks/T-0903-new.md": guard.RuleStatusLine, "work/tasks/T-0902-todo.md": guard.RuleLaneLine,
			"work/tasks/T-0904-new.md": guard.RuleGrantPersonOnly, workspace.StateFile: guard.RuleBadFormat,
			"work/tasks/T-0905-gone.md": guard.RuleStatusLine}
		if !reflect.DeepEqual(refusedRules(o), want) {
			t.Fatalf("refused %v, want %v:\n%s", refusedRules(o), want, o.Reason)
		}
	})
	t.Run("in a worktree: any task file", func(t *testing.T) {
		f := setup(t)
		wt := f.worktree(t)
		write(t, wt, "work/tasks/T-0902-todo.md", taskMD("T-0902", "todo", "")+"More words.\n")
		o := f.climb(t, wt, "T-0901")
		if want := map[string]string{"work/tasks/T-0902-todo.md": guard.RuleTaskOutsideMain}; !reflect.DeepEqual(refusedRules(o), want) {
			t.Fatalf("refused %v, want %v:\n%s", refusedRules(o), want, o.Reason)
		}
		commit(t, wt, "committed on the branch")
		if r := refusedRules(f.climb(t, wt, "T-0901")); r["work/tasks/T-0902-todo.md"] != guard.RuleTaskOutsideMain {
			t.Errorf("refused %v", r)
		}
	})
	t.Run("in a worktree: STATE against the merge base", func(t *testing.T) {
		f := setup(t)
		write(t, f.main, workspace.StateFile, "---\nformat: bonsai.state/1\nupdated: 2026-10-13\nupdated_by: a test\nlabels: {}\n---\n\nWhere it stands.\n")
		commit(t, f.main, "STATE")
		wt := f.worktree(t)
		write(t, wt, workspace.StateFile, "---\nformat: bonsai.state/1\nupdated: 2026-10-14\nupdated_by: a test\nlabels: {}\n---\n\nMoved on.\n")
		commit(t, wt, "STATE rewritten")
		green(t, f.climb(t, wt, "T-0901"), 1)
		write(t, wt, workspace.StateFile, "---\nformat: bonsai.state/1\nupdated: [\n---\n")
		commit(t, wt, "STATE broken")
		if r := refusedRules(f.climb(t, wt, "T-0901")); r[workspace.StateFile] != guard.RuleBadFormat {
			t.Errorf("refused %v", r)
		}
	})
}

func TestAtMostFiftyListed(t *testing.T) {
	f := newProject(t, nil)
	wt := f.worktree(t)
	for i := 0; i < MaxListed+12; i++ {
		write(t, wt, "assets/many/"+strconv.Itoa(1000+i)+".txt", "x\n")
	}
	o := f.climb(t, wt, "T-0901")
	if len(o.Refused) != MaxListed || o.More != 12 {
		t.Fatalf("%d listed, %d more", len(o.Refused), o.More)
	}
	red(t, o, "62 of 62 changed paths refused", "- and 12 more")
	if got := strings.Count(o.Reason, " (protected): "); got != MaxListed {
		t.Errorf("the reason lists %d paths", got)
	}
	if strings.Contains(o.Reason, "1061.txt") || !strings.Contains(o.Reason, "1049.txt") {
		t.Errorf("not the first fifty in path order")
	}
}

func TestByteStable(t *testing.T) {
	f := newProject(t, nil)
	wt := f.worktree(t)
	for _, p := range []string{"assets/z.txt", "assets/a.txt", ".claude/settings.json", "secret/x/y.txt", "assets/m/n.txt"} {
		write(t, wt, p, "x\n")
	}
	commit(t, wt, "several")
	write(t, wt, "assets/b.txt", "x\n")
	first := f.climb(t, wt, "T-0901")
	for i := 0; i < 3; i++ {
		if again := f.climb(t, wt, "T-0901"); !reflect.DeepEqual(first, again) {
			t.Fatalf("two climbs differ:\n%s\n---\n%s", first.Reason, again.Reason)
		}
	}
	var paths []string
	for _, r := range first.Refused {
		paths = append(paths, r.Path)
	}
	if want := []string{".claude/settings.json", "assets/a.txt", "assets/b.txt", "assets/m/n.txt", "assets/z.txt", "secret/x/y.txt"}; !reflect.DeepEqual(paths, want) {
		t.Errorf("refused in order %v, want %v", paths, want)
	}
	for _, l := range strings.Split(first.Reason, "\n") {
		for _, r := range l {
			if r < 0x20 || r > 0x7e {
				t.Errorf("a reason line is not ASCII: %q", l)
				break
			}
		}
	}
}

func TestErrors(t *testing.T) {
	if o := Judge(Input{}); o.Status != "error" {
		t.Errorf("no checkout: %+v", o)
	}
	tmp, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	testpack.Isolate(t, tmp)
	dir := filepath.Join(tmp, "not-a-repository")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	o := Judge(Input{Root: dir, Main: dir})
	if o.Status != "error" || !strings.Contains(o.Reason, "git cannot say what this checkout changed") || o.Captures != nil {
		t.Errorf("not a repository: %+v", o)
	}
}

// Past its limit, rung 0 is error, whatever its git calls had found: its changes are not judged. The same climb with
// the default limit is green.
func TestTimeout(t *testing.T) {
	f := newProject(t, nil)
	in := Input{Root: f.main, Main: f.main, Home: f.home, Timeout: time.Nanosecond}
	o := Judge(in)
	if o.Status != "error" || o.Captures != nil || o.Reason != "rung 0 did not finish within its limit of 1ns (bonsai.yaml's "+
		"timeout_s, else 2m0s): git did not answer in time, so the changes are not judged" {
		t.Errorf("past the limit: %+v", o)
	}
	in.Timeout = 0
	if o := Judge(in); o.Status != "green" {
		t.Errorf("the default limit: %+v", o)
	}
}

func TestNestedRepositories(t *testing.T) {
	t.Run("a checkout planted below the top", func(t *testing.T) {
		f := newProject(t, nil)
		wt := f.worktree(t)
		planted := filepath.Join(wt, "docs", "planted")
		write(t, planted, "readme.txt", "a nested repository\n")
		testpack.Git(t, planted, "init", "-q")
		o := f.climb(t, wt, "T-0901")
		if want := map[string]string{"docs/planted/.git": guard.RuleNeverGranted}; !reflect.DeepEqual(refusedRules(o), want) {
			t.Fatalf("refused %v, want %v:\n%s", refusedRules(o), want, o.Reason)
		}
	})
	t.Run("Claude Code's own worktree inside the main checkout", func(t *testing.T) {
		f := newProject(t, nil)
		inside := filepath.Join(f.main, ".claude", "worktrees", "wt1")
		testpack.Git(t, f.main, "worktree", "add", "-q", "-b", "wt1", inside)
		green(t, f.climb(t, f.main, "T-0901"), 0)
		// Its own climb judges its own changes, against the main checkout's lists.
		write(t, inside, "assets/old/a.txt", "changed\n")
		red(t, f.climb(t, inside, "T-0901"), `"assets/old/a.txt" (protected)`)
		// A worktree of another repository there is a planted checkout all the same.
		other := filepath.Join(f.tmp, "other")
		write(t, other, "x.txt", "x\n")
		testpack.Git(t, other, "init", "-q")
		commit(t, other, "another repository")
		testpack.Git(t, other, "worktree", "add", "-q", "-b", "wt2", filepath.Join(f.main, ".claude", "worktrees", "wt2"))
		if r := refusedRules(f.climb(t, f.main, "T-0901")); r[".claude/worktrees/wt2/.git"] != guard.RuleNeverGranted || len(r) != 1 {
			t.Errorf("refused %v", r)
		}
	})
}

func TestASubfolderReadsAsTheTop(t *testing.T) {
	f := newProject(t, nil)
	wt := f.worktree(t)
	write(t, wt, "assets/old/a.txt", "changed\n")
	write(t, wt, "src/deep/new.txt", "new\n")
	at := f.climb(t, wt, "T-0901")
	below := f.climb(t, filepath.Join(wt, "src"), "T-0901")
	if !reflect.DeepEqual(at, below) {
		t.Errorf("from the top:\n%s\nfrom src/:\n%s", at.Reason, below.Reason)
	}
	red(t, below, `"assets/old/a.txt" (protected)`)
}

// TestBaseUpstream is BUG-124: work built straight on the base branch has its unpushed commits judged, those after the
// base's upstream (found offline, from git's own refs), as a branch's are after its merge base; a base whose commits
// are all pushed judges nothing more; a base with no upstream says its commits were not judged.
func TestBaseUpstream(t *testing.T) {
	withOrigin := func(t *testing.T) *fx {
		f := newProject(t, nil)
		origin := filepath.Join(f.tmp, "origin.git")
		testpack.Git(t, f.tmp, "init", "-q", "--bare", origin)
		testpack.Git(t, f.main, "remote", "add", "origin", origin)
		testpack.Git(t, f.main, "push", "-q", "-u", "origin", "main")
		return f
	}
	t.Run("an unpushed commit to a protected path is refused", func(t *testing.T) {
		f := withOrigin(t)
		write(t, f.main, "assets/old/a.txt", "changed on the base\n")
		write(t, f.main, "src/main.txt", "changed too\n")
		commit(t, f.main, "a protected change, committed on the base, not pushed")
		o := f.climb(t, f.main, "T-0901")
		if want := map[string]string{"assets/old/a.txt": guard.RuleProtected}; !reflect.DeepEqual(refusedRules(o), want) {
			t.Errorf("refused %v, want %v", refusedRules(o), want)
		}
		red(t, o, `"assets/old/a.txt" (protected)`, "1 of 2 changed paths refused")
		if strings.Contains(o.Reason, "not judged here") {
			t.Errorf("a base with an upstream says something was not judged: %s", o.Reason)
		}
		// Pushed, nothing is left to judge.
		testpack.Git(t, f.main, "push", "-q", "origin", "main")
		o = f.climb(t, f.main, "T-0901")
		green(t, o, 0)
		if o.Reason != "" {
			t.Errorf("a pushed base has a reason: %s", o.Reason)
		}
	})
	t.Run("a pushed base judges only the working tree", func(t *testing.T) {
		f := withOrigin(t)
		write(t, f.main, "src/new.txt", "uncommitted\n")
		o := f.climb(t, f.main, "T-0901")
		green(t, o, 1)
		if o.Reason != "" {
			t.Errorf("a pushed base has a reason: %s", o.Reason)
		}
	})
	t.Run("an upstream not an ancestor: the merge base", func(t *testing.T) {
		f := withOrigin(t)
		// Someone else's commit on the upstream, fetched; a local commit beside it.
		other := filepath.Join(f.tmp, "other")
		testpack.Git(t, f.tmp, "clone", "-q", filepath.Join(f.tmp, "origin.git"), other)
		write(t, other, "src/theirs.txt", "theirs\n")
		commit(t, other, "theirs")
		testpack.Git(t, other, "push", "-q", "origin", "main")
		testpack.Git(t, f.main, "fetch", "-q", "origin")
		write(t, f.main, "assets/old/a.txt", "mine\n")
		commit(t, f.main, "mine")
		o := f.climb(t, f.main, "T-0901")
		if want := map[string]string{"assets/old/a.txt": guard.RuleProtected}; !reflect.DeepEqual(refusedRules(o), want) {
			t.Errorf("refused %v, want %v", refusedRules(o), want)
		}
		red(t, o, "1 of 1 changed paths refused")
	})
	t.Run("no upstream: the base's commits are not judged, and it says so", func(t *testing.T) {
		f := newProject(t, nil)
		write(t, f.main, "src/main.txt", "committed on the base\n")
		commit(t, f.main, "a commit on a base with no upstream")
		o := f.climb(t, f.main, "T-0901")
		green(t, o, 0)
		if !strings.Contains(o.Reason, `not judged here: the base branch's commits (no upstream for "main", so only what `+
			"the working tree changed was judged)") {
			t.Errorf("no upstream: the reason does not say the base's commits were not judged: %q", o.Reason)
		}
	})
}
