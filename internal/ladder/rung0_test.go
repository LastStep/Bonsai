package ladder

// Rung 0 through the runner (plan-5 5.4.3's proof, re-run through bonsai ladder; "5.4 done" check 7): fixture projects
// whose floor holds rung 0, climbed with a named task. Green when the task grants what its branch changed; red with each
// refused path, its rule and the grant or revert it needs in the result's reason and the climb's next; a person-only
// path in command mode a person's step; what rung 0 did not judge kept in a green result's captures.

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/guard"
	"github.com/LastStep/Bonsai/internal/testpack"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// r0Lock is a lock with no pack and no file: what a project linked with packs: [] holds, so bonsai check (which rung 0
// runs) finds nothing about the commit.
const r0Lock = "{\n  \"format\": \"bonsai.lock/1\",\n  \"written_by\": \"0.0.0-test\",\n  \"packs\": [],\n  \"files\": {},\n  \"format0\": {}\n}\n"

// r0Run is a format-1 run report, in place of the fixture's format-0 one (bonsai check's format0-new about the commit).
const r0Run = "---\nformat: bonsai.run/1\nid: R-2026-10-13-T-0901\ntask: T-0901\nrole: builder\nmodel: a model\nstarted:\nfinished:\n" +
	"outcome: running\ncommits: []\nlabels: {}\n---\n\n# A run\n"

// r0YAML is a bonsai.yaml whose floor is rung 0 (required) and rung 1 (echo), with assets/** protected and secret/**
// and .claude/** person-only.
func r0YAML() string {
	return strings.Replace(yamlOf("[0, 1]", "{}", "[]", rung{n: 0, name: "guard", kind: "guard", required: true},
		rung{n: 1, name: "one", command: "echo one", required: true}),
		"protected: []\nperson_only: []", `protected: ["assets/**", "secret/**", ".claude/**"]`+"\n"+`person_only: ["secret/**", ".claude/**"]`, 1)
}

// allowsTask is a format-1 task file granting allows (a YAML flow list, "" for none).
func allowsTask(id, status, allows string) string {
	labels := "{}"
	if allows != "" {
		labels = "\n  bonsai.allows: " + allows
	}
	return strings.Replace(taskText(id, status, ""), "labels: {}", "labels: "+labels, 1)
}

// r0Project is a fixture project for rung 0: T-0901 running, granting assets/t-0901/** and secret/granted/**; T-0902
// todo, granting nothing.
func r0Project(t *testing.T) *fx {
	t.Helper()
	return project(t, r0YAML(), map[string]string{
		workspace.LockFile:                 r0Lock,
		"work/runs/R-2026-10-13-T-0901.md": r0Run,
		"work/tasks/T-0901-a.md":           allowsTask("T-0901", "running", `["assets/t-0901/**", "secret/granted/**"]`),
		"work/tasks/T-0902-b.md":           allowsTask("T-0902", "todo", ""),
		"assets/old/a.txt":                 "an asset\n",
		"src/main.txt":                     "the code\n",
	})
}

// worktree adds a task worktree on its own branch beside the project, from main's HEAD.
func (f *fx) worktree(branch string) string {
	f.t.Helper()
	wt := filepath.Join(f.tmp, "wt-"+branch)
	testpack.Git(f.t, f.root, "worktree", "add", "-q", "-b", branch, wt)
	return wt
}

// writeIn writes a file in a checkout.
func (f *fx) writeIn(dir, rel, body string) {
	f.t.Helper()
	saved := f.root
	f.root = dir
	f.write(rel, body)
	f.root = saved
}

// commitIn commits every change in a checkout.
func (f *fx) commitIn(dir, msg string) {
	f.t.Helper()
	testpack.Git(f.t, dir, "add", "-A")
	testpack.Git(f.t, dir, "-c", "user.name=t", "-c", "user.email=t@example.invalid", "commit", "-q", "--allow-empty", "-m", msg)
}

// reasonOf is a rung's reason in a result, "" for null.
func reasonOf(f *fx, task string, n int) string {
	l := f.result(task)
	for _, r := range l.Rungs {
		if r.Rung == int64(n) && r.Reason != nil {
			return *r.Reason
		}
	}
	return ""
}

// The named task grants what its branch changed: rung 0 green, its captures checked (the paths) and not_judged
// (null: the base was read and no machine finding came), its reason null; the climb green and clean.
func TestRung0GreenWhenGranted(t *testing.T) {
	f := r0Project(t)
	wt := f.worktree("t0901")
	f.writeIn(wt, "assets/t-0901/new.txt", "granted\n")
	f.writeIn(wt, "secret/granted/key.txt", "granted, person-only\n")
	f.writeIn(wt, "src/main.txt", "changed\n")
	f.commitIn(wt, "the task's work")
	a := f.climb(wt, Options{Task: "T-0901"})
	if a.Exit != 0 || a.Doc.Next != nil {
		t.Fatalf("exit %d, %+v, next %+v\n%s", a.Exit, a.Error, a.Doc.Next, f.out.String())
	}
	l := f.result("T-0901")
	r := l.Rungs[0]
	if statuses(l) != "0:green 1:green" || r.Reason != nil || r.Command != nil || l.Git.Dirty || l.Git.Branch != "t0901" {
		t.Fatalf("%s, %+v, %+v", statuses(l), r, l.Git)
	}
	if got := showOf(r.Captures, "checked") + " " + showOf(r.Captures, "not_judged"); got != "3 null" {
		t.Errorf("the captures: %s", got)
	}
}

// Paths the named task does not grant: rung 0 red, the required rung after it skipped; the reason lists each path with
// its rule, and the climb's next names each path's rule and the grant or revert it needs, how to revert, and the climb
// again, who agent (agents mode: an agent may write a grant in its task).
func TestRung0RedNamesEachFix(t *testing.T) {
	f := r0Project(t)
	wt := f.worktree("t0901")
	f.writeIn(wt, "assets/old/a.txt", "changed\n")
	f.writeIn(wt, "secret/other/x.txt", "person-only\n")
	f.writeIn(wt, ".claude/settings.json", "{}\n")
	f.writeIn(wt, "assets/t-0901/ok.txt", "granted\n")
	f.commitIn(wt, "work beyond the grants")
	mb := testpack.Git(t, wt, "merge-base", "main", "HEAD")
	a := f.climb(wt, Options{Task: "T-0901"})
	l := f.result("T-0901")
	if a.Exit != ExitRed || statuses(l) != "0:red 1:skipped" {
		t.Fatalf("exit %d, %s\n%s", a.Exit, statuses(l), f.out.String())
	}
	reason := reasonOf(f, "T-0901", 0)
	for _, w := range []string{`3 of 4 changed paths refused, judged as the guard judges an edit by the named task T-0901`,
		`- ".claude/settings.json" (never-granted)`, `- "assets/old/a.txt" (protected)`, `- "secret/other/x.txt" (person-only)`} {
		if !strings.Contains(reason, w) {
			t.Errorf("the reason lacks %q:\n%s", w, reason)
		}
	}
	want := `fix the guard rung (rung 0): ".claude/settings.json" (never-granted) needs a revert: no grant opens it; ` +
		`"assets/old/a.txt" (protected) needs a grant in T-0901's bonsai.allows, or a revert; ` +
		`"secret/other/x.txt" (person-only) needs a person's grant in T-0901's bonsai.allows, or a revert; ` +
		`a revert: git checkout ` + mb + ` -- <path> (a path ` + mb + ` does not hold: delete it) ; ` +
		`then run: bonsai ladder --task T-0901 --json`
	if a.Doc.Next == nil || a.Doc.Next.Do != want || a.Doc.Next.Who != "agent" {
		t.Errorf("next:\n%+v\nwant\n%s", a.Doc.Next, want)
	}
	if !strings.Contains(f.out.String(), "next: fix the guard rung (rung 0): \".claude/settings.json\" (never-granted)") {
		t.Errorf("the human lines:\n%s", f.out.String())
	}
	// The revert as next names it, run as written for each path, and the climb again: green.
	for _, p := range []string{"assets/old/a.txt"} {
		testpack.Git(t, wt, "checkout", mb, "--", p)
	}
	testpack.Git(t, wt, "rm", "-q", "secret/other/x.txt", ".claude/settings.json")
	f.commitIn(wt, "reverted")
	if a := f.climb(wt, Options{Task: "T-0901"}); a.Exit != 0 {
		t.Errorf("after the reverts: exit %d\n%s", a.Exit, reasonOf(f, "T-0901", 0))
	}
}

// A person-only path in command mode (this machine's settings): only a person's grant opens it, so next's who is
// person; a protected path alone in command mode stays the agent's (a grant its task may carry, or a revert).
func TestRung0PersonOnlyInCommandMode(t *testing.T) {
	f := r0Project(t)
	md, err := workspace.MachineDir(f.home, f.root)
	if err != nil {
		t.Fatal(err)
	}
	f.writeIn(md, workspace.MachineSettingsFile, `{"status_writes": "command", "status_command": "studio-move"}`)
	wt := f.worktree("t0901")
	f.writeIn(wt, "assets/old/a.txt", "changed\n")
	f.commitIn(wt, "a protected path")
	a := f.climb(wt, Options{Task: "T-0901"})
	if a.Exit != ExitRed || a.Doc.Next == nil || a.Doc.Next.Who != "agent" || !strings.Contains(a.Doc.Next.Do, `"assets/old/a.txt" (protected)`) {
		t.Fatalf("a protected path: exit %d, %+v", a.Exit, a.Doc.Next)
	}
	f.writeIn(wt, "secret/other/x.txt", "person-only\n")
	f.commitIn(wt, "a person-only path")
	a = f.climb(wt, Options{Task: "T-0901"})
	if a.Exit != ExitRed || a.Doc.Next == nil || a.Doc.Next.Who != "person" ||
		!strings.Contains(a.Doc.Next.Do, `"secret/other/x.txt" (`+guard.RulePersonOnly+`) needs a person's grant in T-0901's bonsai.allows, or a revert`) {
		t.Errorf("a person-only path: exit %d, %+v", a.Exit, a.Doc.Next)
	}
	if d := a.Doc.Next; d != nil && !strings.HasPrefix(d.Do, "fix the guard rung (rung 0): ") {
		t.Errorf("next: %s", d.Do)
	}
}
