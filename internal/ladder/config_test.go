package ladder

// The ladder's settings come from the main checkout's bonsai.yaml as committed at its HEAD, on the base branch this
// machine recorded (V2's reproducer m2-main-config): an uncommitted edit there that lowers the floor, or the main
// checkout switched to a branch whose committed bonsai.yaml is looser, refuses the climb and blocks the stop gate,
// each with a person's step; bonsai.yaml changed during a climb makes its result dirty.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/testpack"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// cfgProject is a project whose floor is rung 1 (echo) and rung 2 (red: it shows a file no checkout holds), with
// T-0902 running, and a worktree on branch t0902 holding a commit and the task's run report: its climb is red.
func cfgProject(t *testing.T) (*fx, string) {
	t.Helper()
	rungs := []rung{{n: 1, name: "one", command: "echo one", required: true}, {n: 2, name: "two", command: show("missing.txt"), required: true}}
	f := project(t, yamlOf("[1, 2]", "{}", "[]", rungs...), map[string]string{
		"work/tasks/T-0902-a.md": taskText("T-0902", "running", ""),
	})
	wt := f.worktree("t0902")
	f.writeIn(wt, "work/runs/R-2026-10-13-T-0902.md", "# R-2026-10-13-T-0902\n\nWhat changed.\n")
	f.commitIn(wt, "the branch's work")
	if a := f.climb(wt, Options{Task: "T-0902"}); a.Exit != ExitRed {
		t.Fatalf("the branch climbs %d, want red\n%s", a.Exit, f.out.String())
	}
	return f, wt
}

// lowered is the project's bonsai.yaml with rung 2 out of the floor.
func lowered() string {
	return yamlOf("[1]", "{}", "[]", rung{n: 1, name: "one", command: "echo one", required: true},
		rung{n: 2, name: "two", command: show("missing.txt"), required: true})
}

func TestLadderSettingsAsCommitted(t *testing.T) {
	t.Run("an uncommitted edit in the main checkout", func(t *testing.T) {
		f, wt := cfgProject(t)
		f.write("bonsai.yaml", lowered())
		a := f.climb(wt, Options{Task: "T-0902"})
		if a.Exit != ExitState || a.Error == nil || a.Error.Code != "config-uncommitted" || a.Error.Who != "person" ||
			!strings.Contains(a.Error.What, "the main checkout's bonsai.yaml differs from the one committed at its HEAD") ||
			a.Error.Next != "a person commits the change to bonsai.yaml in the main checkout (git commit), or reverts it (git checkout "+
				"HEAD -- bonsai.yaml); then run: bonsai ladder --task T-0902 --json" {
			t.Fatalf("exit %d, %+v", a.Exit, a.Error)
		}
		if code, stderr := gate(t, wt, "T-0902"); code != 2 || !strings.Contains(stderr, "differs from the one committed at its HEAD") {
			t.Errorf("the stop gate: %d %q", code, stderr)
		}
		// Line ends alone are no change: a CRLF copy of the committed file climbs.
		f.write("bonsai.yaml", strings.ReplaceAll(testpack.Git(t, f.root, "show", "HEAD:bonsai.yaml")+"\n", "\n", "\r\n"))
		if a := f.climb(wt, Options{Task: "T-0902"}); a.Exit != ExitRed {
			t.Errorf("a CRLF copy: exit %d, %+v", a.Exit, a.Error)
		}
	})
	t.Run("the main checkout switched to a branch with a looser bonsai.yaml, committed", func(t *testing.T) {
		f, wt := cfgProject(t)
		testpack.Git(t, f.root, "checkout", "-q", "-b", "loose")
		f.write("bonsai.yaml", lowered())
		f.commit("a looser floor")
		a := f.climb(wt, Options{Task: "T-0902"})
		if a.Exit != ExitState || a.Error == nil || a.Error.Code != "main-not-verified" || a.Error.Who != "person" ||
			!strings.Contains(a.Error.What, `the main checkout holds the branch "loose", and this machine's record names "main" as the `+
				"workspace's base branch") || !strings.Contains(a.Error.Next, "a person switches the main checkout back (git switch main)") {
			t.Fatalf("exit %d, %+v", a.Exit, a.Error)
		}
		if code, stderr := gate(t, wt, "T-0902"); code != 2 || !strings.Contains(stderr, "the main checkout is not on its base branch") {
			t.Errorf("the stop gate: %d %q", code, stderr)
		}
		testpack.Git(t, f.root, "checkout", "-q", "main")
		if a := f.climb(wt, Options{Task: "T-0902"}); a.Exit != ExitRed {
			t.Errorf("back on main: exit %d, %+v", a.Exit, a.Error)
		}
	})
	t.Run("a record written before Bonsai kept a base", func(t *testing.T) {
		f, wt := cfgProject(t)
		dir, err := workspace.MachineDir(f.home, f.root)
		if err != nil {
			t.Fatal(err)
		}
		p := filepath.Join(dir, workspace.MachineRecordFile)
		raw, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		old := strings.Replace(string(raw), ",\n  \"base_branch\": \"main\"", "", 1)
		if old == string(raw) {
			t.Fatalf("no base_branch in the record:\n%s", raw)
		}
		if err := os.WriteFile(p, []byte(old), 0o644); err != nil {
			t.Fatal(err)
		}
		a := f.climb(wt, Options{Task: "T-0902"})
		if a.Exit != ExitState || a.Error == nil || a.Error.Code != "main-not-verified" ||
			!strings.Contains(a.Error.What, "names no base branch (it was written before Bonsai kept one)") ||
			!strings.Contains(a.Error.Next, "runs bonsai update there, which records it as the base") {
			t.Fatalf("exit %d, %+v", a.Exit, a.Error)
		}
		// The record written again (bonsai update's step) keeps main as the base.
		if _, err := workspace.RecordCheckout(f.home, f.root, testID, time.Now()); err != nil {
			t.Fatal(err)
		}
		if a := f.climb(wt, Options{Task: "T-0902"}); a.Exit != ExitRed {
			t.Errorf("recorded again: exit %d, %+v", a.Exit, a.Error)
		}
	})
	t.Run("bonsai.yaml changed during the climb", func(t *testing.T) {
		// The rung, run in the worktree beside the main checkout (testpack's project folder), appends a line to the main
		// checkout's bonsai.yaml.
		touch := "echo touched>> " + filepath.Join("..", "project", "bonsai.yaml")
		f := project(t, yamlOf("[1]", "{}", "[]", rung{n: 1, name: "one", command: touch, required: true}),
			map[string]string{"work/tasks/T-0901-a.md": taskText("T-0901", "running", "")})
		wt := f.worktree("t0901")
		f.writeIn(wt, "work/runs/R-2026-10-13-T-0901.md", "# R\n")
		f.commitIn(wt, "work")
		a := f.climb(wt, Options{Task: "T-0901"})
		if a.Exit != 0 {
			t.Fatalf("exit %d, %+v\n%s", a.Exit, a.Error, f.out.String())
		}
		if !f.result("T-0901").Git.Dirty || !strings.Contains(f.out.String(), "the main checkout's bonsai.yaml changed during the climb") {
			t.Errorf("dirty %v\n%s", f.result("T-0901").Git.Dirty, f.out.String())
		}
	})
}
