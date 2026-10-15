package ladder

// V2b's K1 (its reproducer m2b-base-commit; the director, 13 Oct: fix now): settings for the checks committed on the base
// branch itself are history the runner and the stop gate read, so an agent that lowers the floor there must not make
// a red branch climb green. Three ways to put a looser bonsai.yaml at the main checkout's HEAD: a commit on the base
// branch (from the agent's worktree, git -C), git reset --hard to a looser commit, and the branch's ref moved to one
// (git update-ref, then the file checked out). Each is refused by the runner (config-looser, a person's step) and
// blocked by the gate (stop-config-looser), naming the rung that left the floor; a person's bonsai update then records
// the looser settings as accepted, and the climb goes on. This file uses only what the fixtures held before the fix, so
// it runs on the code before it too, where each case climbs green and the gate lets the session end.

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/testpack"
	"github.com/LastStep/Bonsai/internal/workspace"
)

func TestLooserSettingsOnTheBase(t *testing.T) {
	cases := []struct {
		name  string
		loose func(f *fx, wt string)
	}{
		{"committed on the base branch from the worktree", func(f *fx, wt string) {
			f.write("bonsai.yaml", lowered())
			testpack.Git(t, wt, "-C", f.root, "-c", "user.name=t", "-c", "user.email=t@example.invalid", "commit", "-q", "-am",
				"settings: a lower floor")
		}},
		{"git reset --hard to a looser commit", func(f *fx, wt string) {
			testpack.Git(t, f.root, "checkout", "-q", "-b", "side")
			f.write("bonsai.yaml", lowered())
			f.commit("settings: a lower floor, on a side branch")
			testpack.Git(t, f.root, "checkout", "-q", "main")
			testpack.Git(t, f.root, "reset", "-q", "--hard", "side")
		}},
		{"the branch's ref moved to a looser commit", func(f *fx, wt string) {
			testpack.Git(t, f.root, "checkout", "-q", "-b", "side")
			f.write("bonsai.yaml", lowered())
			f.commit("settings: a lower floor, on a side branch")
			testpack.Git(t, f.root, "checkout", "-q", "main")
			testpack.Git(t, f.root, "update-ref", "refs/heads/main", "side")
			testpack.Git(t, f.root, "reset", "-q")
			testpack.Git(t, f.root, "checkout", "-q", "--", "bonsai.yaml")
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f, wt := cfgProject(t)
			c.loose(f, wt)
			if got := testpack.Git(t, f.root, "show", "HEAD:bonsai.yaml"); !strings.Contains(got, "ladder_floor: [1]") {
				t.Fatalf("main's HEAD does not hold the looser bonsai.yaml:\n%s", got)
			}
			a := f.climb(wt, Options{Task: "T-0902"})
			if a.Exit != ExitState || a.Error == nil || a.Error.Code != "config-looser" || a.Error.Who != "person" ||
				!strings.Contains(a.Error.What, "the main checkout's bonsai.yaml, as committed at its HEAD, sets the checks looser than "+
					"the settings this machine last accepted for the workspace: ladder_floor no longer holds rung 2") ||
				!strings.Contains(a.Error.Next, "a person runs bonsai update in the main checkout") ||
				!strings.HasSuffix(a.Error.Next, "; then run: bonsai ladder --task T-0902 --json") {
				t.Errorf("the red branch's climb on a looser floor: exit %d, %+v\n%s", a.Exit, a.Error, f.out.String())
			}
			if code, stderr := gate(t, wt, "T-0902"); code != 2 || !strings.Contains(stderr, "sets the checks looser than the settings "+
				"this machine last accepted for the workspace: ladder_floor no longer holds rung 2") {
				t.Errorf("the stop gate: %d %q", code, stderr)
			}
			// A person's bonsai update records the looser settings as accepted: the climb goes on, by them.
			if _, err := workspace.RecordCheckout(f.home, f.root, testID, time.Now()); err != nil {
				t.Fatal(err)
			}
			if a := f.climb(wt, Options{Task: "T-0902"}); a.Exit != ExitGreen {
				t.Errorf("accepted by a person: exit %d, %+v\n%s", a.Exit, a.Error, f.out.String())
			}
		})
	}
}

// remove is a file's removal command: rm on Linux and macOS, del in cmd.
func remove(file string) string {
	if runtime.GOOS == "windows" {
		return "del " + file
	}
	return "rm " + file
}

// V2c's break-it (its reproducer k1x-added-rung): a rung added runs its command inside the climb, between the climb's
// two reads of the tree. Two entries committed on the base branch from the worktree, for numbers already on the floor,
// one after rung 1 that writes the file the red rung 2 needs and one after rung 2 that takes it away again, made the
// red branch climb green and clean, and the stop gate let the session end. A rung added is a loosening a person
// records: the climb is refused and the gate blocks, naming each entry, and nothing ran; a person's bonsai update
// records them as accepted, and the climb goes on by them. Like the test above, this uses only what the fixtures held
// before the fix, so it runs on the code before it too, where the red branch climbs green.
func TestAddedRungsOnTheBase(t *testing.T) {
	f, wt := cfgProject(t)
	was := f.accepted()
	f.write("bonsai.yaml", yamlOf("[1, 2]", "{}", "[]", rung{n: 1, name: "one", command: "echo one", required: true},
		rung{n: 1, name: "lint", command: "echo fine> missing.txt", required: true},
		rung{n: 2, name: "two", command: show("missing.txt"), required: true},
		rung{n: 2, name: "tidy", command: remove("missing.txt"), required: true}))
	testpack.Git(t, wt, "-C", f.root, "-c", "user.name=t", "-c", "user.email=t@example.invalid", "commit", "-q", "-am",
		"settings: two more checks")
	added := "rung 1 (its entry 2 in the ladder) was added: its command runs inside the climb; " +
		"rung 2 (its entry 2 in the ladder) was added: its command runs inside the climb"
	f.out.Reset()
	a := f.climb(wt, Options{Task: "T-0902"})
	if a.Exit == ExitGreen {
		t.Errorf("the red branch climbed green by the rungs added on the base\n%s", f.out.String())
	}
	if a.Exit != ExitState || a.Error == nil || a.Error.Code != "config-looser" || a.Error.Who != "person" ||
		!strings.Contains(a.Error.What, "sets the checks looser than the settings this machine last accepted for the workspace: "+added) ||
		!strings.Contains(a.Error.Next, "a person runs bonsai update in the main checkout") {
		t.Errorf("the climb: exit %d, %+v", a.Exit, a.Error)
	}
	if _, err := os.Stat(filepath.Join(wt, "missing.txt")); err == nil {
		t.Error("a rung added ran: missing.txt is there")
	}
	if code, stderr := gate(t, wt, "T-0902"); code != 2 || !strings.Contains(stderr, "sets the checks looser than the settings "+
		"this machine last accepted for the workspace: "+added) {
		t.Errorf("the stop gate: %d %q", code, stderr)
	}
	if f.accepted() != was {
		t.Error("the climb wrote the rungs added into this machine's record")
	}
	// A person's bonsai update records them as accepted: the climb goes on, by them.
	f.accept()
	if a := f.climb(wt, Options{Task: "T-0902"}); a.Exit != ExitGreen {
		t.Errorf("accepted by a person: exit %d, %+v\n%s", a.Exit, a.Error, f.out.String())
	}
}
