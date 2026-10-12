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

// What rung 0 did not judge stays in a green result: its reason is null there, its not_judged capture is not. A
// machine finding (this home's install record) is listed and does not count; with the main checkout on a detached HEAD
// there is no base, so the branch's commits are not judged (a protected change committed on the branch passes), and the
// result says so; the human lines print both as notes.
func TestRung0NotJudgedKept(t *testing.T) {
	f := r0Project(t)
	f.writeIn(f.home, "install.json", "not the installed bonsai's record\n")
	a := f.climb("", Options{Task: "T-0901"})
	r := f.result("T-0901").Rungs[0]
	if a.Exit != 0 || r.Status != "green" || r.Reason != nil ||
		r.Captures.String("not_judged") != "bonsai check's findings about this machine, not the commit (bonsai-path)" {
		t.Fatalf("a machine finding: exit %d, %+v, %s\n%s", a.Exit, r, showOf(r.Captures, "not_judged"), f.out.String())
	}
	if !strings.Contains(f.out.String(), "note: not judged here: bonsai check's findings about this machine") {
		t.Errorf("the human lines:\n%s", f.out.String())
	}
	wt := f.worktree("t0901")
	f.writeIn(wt, "assets/old/a.txt", "changed on the branch\n")
	f.commitIn(wt, "a protected change")
	testpack.Git(t, f.root, "checkout", "-q", "--detach")
	a = f.climb(wt, Options{Task: "T-0901"})
	r = f.result("T-0901").Rungs[0]
	nj := r.Captures.String("not_judged")
	if a.Exit != 0 || r.Status != "green" || !strings.HasPrefix(nj, "the branch's commits (no base branch: the main checkout's HEAD is detached") ||
		!strings.HasSuffix(nj, "; bonsai check's findings about this machine, not the commit (bonsai-path)") {
		t.Errorf("no base: exit %d, %s, not_judged %q", a.Exit, r.Status, nj)
	}
}

// A branch that deletes bonsai.yaml: the climb runs in the worktree by its main checkout's bonsai.yaml, and rung 0 is
// red: the deletion refused (bonsai.yaml is person-only, the floor's) and bonsai check unable to run there; next names
// both fixes.
func TestRung0BranchDeletesBonsaiYAML(t *testing.T) {
	f := r0Project(t)
	wt := f.worktree("t0901")
	testpack.Git(t, wt, "rm", "-q", "bonsai.yaml")
	f.commitIn(wt, "delete bonsai.yaml")
	a := f.climb(wt, Options{Task: "T-0901"})
	if a.Exit != ExitRed || a.Doc.Next == nil {
		t.Fatalf("exit %d, %+v\n%s", a.Exit, a.Error, f.out.String())
	}
	reason := reasonOf(f, "T-0901", 0)
	if !strings.Contains(reason, `- "bonsai.yaml" (person-only)`) || !strings.Contains(reason, "bonsai check cannot run in this checkout") {
		t.Errorf("the reason:\n%s", reason)
	}
	if !strings.Contains(a.Doc.Next.Do, `"bonsai.yaml" (person-only) needs a person's grant in T-0901's bonsai.allows, or a revert`) ||
		!strings.Contains(a.Doc.Next.Do, "rebase the branch on the base branch so it holds the main checkout's bonsai.yaml, or restore it") {
		t.Errorf("next: %s", a.Doc.Next.Do)
	}
}

// A worktree's climb is judged against the main checkout: its branch loosens its own bonsai.yaml (assets/** no longer
// protected) and its own copy of the task (granting assets/**), and changes an asset; rung 0 judges by main's lists and
// main's copy of the task, so the asset and bonsai.yaml are refused, while the task file, free in agents mode, is not.
func TestRung0WorktreeJudgedAgainstMain(t *testing.T) {
	f := r0Project(t)
	wt := f.worktree("t0901")
	f.writeIn(wt, "bonsai.yaml", strings.Replace(r0YAML(), `protected: ["assets/**", "secret/**", ".claude/**"]`, `protected: ["secret/**", ".claude/**"]`, 1))
	f.writeIn(wt, "work/tasks/T-0901-a.md", allowsTask("T-0901", "running", `["assets/**", "secret/granted/**"]`))
	f.writeIn(wt, "assets/old/a.txt", "changed\n")
	f.commitIn(wt, "loosen the lists on the branch")
	a := f.climb(wt, Options{Task: "T-0901"})
	reason := reasonOf(f, "T-0901", 0)
	if a.Exit != ExitRed || !strings.Contains(reason, "2 of 3 changed paths refused") || !strings.Contains(reason, `- "assets/old/a.txt" (protected)`) ||
		!strings.Contains(reason, `- "bonsai.yaml" (person-only)`) || strings.Contains(reason, "T-0901-a.md") {
		t.Errorf("exit %d, the reason:\n%s", a.Exit, reason)
	}
}

// No task named under --ci: nothing is granted, so a protected change is red (contract §13: none fails closed).
func TestRung0CINoTask(t *testing.T) {
	f := r0Project(t)
	f.write("assets/t-0901/new.txt", "granted to T-0901, but no task is named\n")
	a := f.climb("", Options{CI: true})
	l := f.result("ci")
	reason := reasonOf(f, "ci", 0)
	if a.Exit != ExitRed || l.Task != nil || statuses(l) != "0:red 1:skipped" ||
		!strings.Contains(reason, "judged as the guard judges an edit by no named task (nothing granted)") ||
		!strings.Contains(reason, `- "assets/t-0901/new.txt" (protected)`) {
		t.Errorf("exit %d, %s:\n%s", a.Exit, statuses(l), reason)
	}
	if a.Doc.Next == nil || !strings.Contains(a.Doc.Next.Do, `needs a named task (--task) whose bonsai.allows grants it, or a revert`) ||
		!strings.HasSuffix(a.Doc.Next.Do, "then run: bonsai ladder --ci --json") {
		t.Errorf("next: %+v", a.Doc.Next)
	}
}

// 5.3's stop gate with rung 0 in the floor: a green climb of the worktree's HEAD lets the named session end; a red
// one (a refused path committed) keeps it going; the path reverted and climbed green again, it ends.
func TestRung0TheStopGate(t *testing.T) {
	f := r0Project(t)
	wt := f.worktree("t0901")
	f.writeIn(wt, "assets/t-0901/new.txt", "granted\n")
	f.commitIn(wt, "the task's work")
	if a := f.climb(wt, Options{Task: "T-0901"}); a.Exit != 0 {
		t.Fatalf("the green climb: exit %d\n%s", a.Exit, f.out.String())
	}
	if code, stderr := gate(t, wt, "T-0901"); code != 0 {
		t.Fatalf("a green climb with rung 0 refused: %d %q", code, stderr)
	}
	f.writeIn(wt, "assets/old/a.txt", "changed\n")
	f.commitIn(wt, "beyond the grants")
	if a := f.climb(wt, Options{Task: "T-0901"}); a.Exit != ExitRed {
		t.Fatalf("the red climb: exit %d", a.Exit)
	}
	if code, stderr := gate(t, wt, "T-0901"); code != 2 || !strings.Contains(stderr, "is not green") {
		t.Errorf("a red rung 0 let the session end: %d %q", code, stderr)
	}
	testpack.Git(t, wt, "checkout", "main", "--", "assets/old/a.txt")
	f.commitIn(wt, "reverted")
	if a := f.climb(wt, Options{Task: "T-0901"}); a.Exit != 0 {
		t.Fatalf("green again: exit %d", a.Exit)
	}
	if code, stderr := gate(t, wt, "T-0901"); code != 0 {
		t.Errorf("green again, refused: %d %q", code, stderr)
	}
}

// A rung 0 that cannot judge (git cannot read the checkout's index) is error, and next names git status.
func TestRung0ErrorNamesGitStatus(t *testing.T) {
	f := r0Project(t)
	f.write(".git/index", "not an index")
	a := f.climb("", Options{Task: "T-0901"})
	l := f.result("T-0901")
	if a.Exit != ExitRed || statuses(l) != "0:error 1:skipped" || !strings.Contains(reasonOf(f, "T-0901", 0), "git cannot say what this checkout changed") {
		t.Fatalf("exit %d, %s: %s", a.Exit, statuses(l), reasonOf(f, "T-0901", 0))
	}
	if a.Doc.Next == nil || !strings.HasPrefix(a.Doc.Next.Do, "fix the guard rung (rung 0): rung 0 could not judge this checkout's changes "+
		"(the rung's reason says why): run: git status") {
		t.Errorf("next: %+v", a.Doc.Next)
	}
}
