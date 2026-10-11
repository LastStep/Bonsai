package guard

// What step 5.3.3's delete check asks of the one function (design/plan-5.md, 5.3.2 notes 1 and 8; 5.3.3 note 4): a
// delete or move judged as an edit of its path (a change with del set: the floor's .git the project's own top one
// only, another linked project by its top and floor folders alone, what the file holds after it unknown), and the
// folders a shell call's deletes are judged in (judgedRoots), each with its guarded globs.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/testpack"
	"github.com/LastStep/Bonsai/internal/workspace"
)

func TestADeleteIsJudgedAsAnEdit(t *testing.T) {
	dir := withTask(t, "running", "assets/**", "bonsai.yaml")
	writeFile(t, filepath.Join(dir, "vendor", "x", ".git", "HEAD"), "ref: refs/heads/main\n")
	j := judgeIn(dir)
	home, _ := workspace.Home()
	for _, c := range []struct {
		rel  string
		rule string
	}{
		{"free.txt", RuleNotProtected},
		{"assets/a.png", RuleGranted},
		{"bonsai.yaml", RuleGranted},
		{".bonsai/lock.json", RulePersonOnly},
		{"keys/a.key", RulePersonOnly},
		{".bonsai/local", RuleBonsaiFiles},
		{".bonsai/tasks.md", RuleBonsaiFiles},
		{".git", RuleNeverGranted},
		{".claude/settings.json", RuleNeverGranted},
		{"vendor/x/.git", RuleNotProtected},
		{"vendor/x/bonsai.yaml", RulePersonOnly},
		{".bonsai/STATE.md", RuleNotProtected},
	} {
		if d := mayIn(j, dir, filepath.Join(dir, filepath.FromSlash(c.rel)), true); d.Rule != c.rule {
			t.Errorf("a delete of %s: %+v, want %s", c.rel, d, c.rule)
		}
	}
	// Relative to the folder the command runs in.
	if d, _ := j.may(change{tool: "rm", path: "../.git", dir: filepath.Join(dir, "work"), base: dir, del: true}); d.Rule != RuleNeverGranted {
		t.Errorf("rm ../.git from work/: %+v", d)
	}
	// Bonsai's home: the folder itself and anything in it.
	for _, p := range []string{".", "workspaces", filepath.Join("workspaces", "r-0123456789abcdef", "workspace.json")} {
		if d, _ := j.may(change{tool: "rm", path: p, dir: home, base: dir, del: true}); d.Rule != RuleBonsaiFiles {
			t.Errorf("a delete of %s in Bonsai's home: %+v", p, d)
		}
	}
	// command mode: a delete takes an existing task's status line away (refused); a path that is not there yet is
	// a shell write (allowed).
	commandMode(t, dir)
	j = judgeIn(dir)
	if d := mayIn(j, dir, filepath.Join(dir, "work", "tasks", "T-0001-the-task.md"), true); d.Rule != RuleStatusLine {
		t.Errorf("a delete of a task in command mode: %+v", d)
	}
	if d := mayIn(j, dir, filepath.Join(dir, "work", "tasks", "T-0009-new.md"), true); !d.Allow {
		t.Errorf("a move onto a new task's path in command mode: %+v", d)
	}
}

func TestTheFoldersADeleteIsJudgedIn(t *testing.T) {
	tmp := t.TempDir()
	home := testpack.Isolate(t, tmp)
	main, wt := gitProject(t, home, grantYAML, time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC))
	// Another project on this machine's record; a record whose folder no longer holds a bonsai.yaml.
	other := unrecordedProject(t, strings.Replace(grantYAML, testID, "ws-bbbbbbbbbbbbbbbbbbbbbbbbbb", 1))
	if _, err := workspace.RecordCheckout(home, other, "ws-bbbbbbbbbbbbbbbbbbbbbbbbbb", time.Now()); err != nil {
		t.Fatal(err)
	}
	gone := unrecordedProject(t, strings.Replace(grantYAML, testID, "ws-dddddddddddddddddddddddddd", 1))
	if _, err := workspace.RecordCheckout(home, gone, "ws-dddddddddddddddddddddddddd", time.Now()); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(gone, workspace.ConfigFile)); err != nil {
		t.Fatal(err)
	}
	user := t.TempDir()
	t.Setenv("HOME", user)
	t.Setenv("USERPROFILE", user)

	j := judgeIn(wt)
	roots := j.judgedRoots()
	byDir := map[string]judgedRoot{}
	for _, r := range roots {
		byDir[r.Dir] = r
	}
	own, ok := byDir[wt]
	if !ok || !own.Own || !listHas(own.Globs, "assets/**") || !listHas(own.Globs, ".git") || !listHas(own.Globs, "**/bonsai.yaml") ||
		listHas(own.Globs, "**/.git") || listHas(own.Globs, ".bonsai/STATE.md") {
		t.Errorf("the session's own project: %+v", own)
	}
	if m, ok := byDir[main]; !ok || m.Own || !listHas(m.Globs, "keys/**") {
		t.Errorf("the worktree's verified main checkout: %+v (roots %+v)", m, roots)
	}
	if h, ok := byDir[filepath.Clean(home)]; !ok || !h.Whole {
		t.Errorf("Bonsai's home: %+v", roots)
	}
	if h, ok := byDir[filepath.Join(user, ".bonsai")]; !ok || !h.Whole {
		t.Errorf("~/.bonsai: %+v", roots)
	}
	if o, ok := byDir[other]; !ok || o.Own || strings.Join(o.Globs, " ") != ".git .bonsai .claude" {
		t.Errorf("another project on the record: %+v", o)
	}
	if _, ok := byDir[gone]; ok {
		t.Errorf("a record whose folder holds no bonsai.yaml is a root: %+v", roots)
	}
	if len(roots) != 5 {
		t.Errorf("%d roots: %+v", len(roots), roots)
	}
	// A session in no linked project: Bonsai's home and the projects on the record.
	if n := len(judgeIn("").judgedRoots()); n != 4 {
		t.Errorf("an unlinked session's roots: %d", n)
	}
}
