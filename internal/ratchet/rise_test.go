package ratchet

// The rise step's call (rise.go): NewTests gives internal/newtests' check what it reads (the base chosen from the merge
// base or the task's last result, the rise, the HEAD run's tests, the environment and the stop) and carries back its
// new_tests, its lines and its summary; the rise as a count of tests.

import (
	"context"
	"encoding/json"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/newtests"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/testout"
)

func TestNewTestsCall(t *testing.T) {
	defer func(saved func(newtests.Input) newtests.Result) { check = saved }(check)
	var got []newtests.Input
	check = func(in newtests.Input) newtests.Result {
		got = append(got, in)
		return newtests.Result{NewTests: format.LadderNewTests{Base: in.Base, Count: 2, FailedOnBase: 1, NotRun: 1}, Named: true,
			Failed: []string{"TestA"}, NotRun: []string{"TestB"}, Took: 3 * time.Second}
	}
	stop, cancel := context.WithCancel(context.Background())
	defer cancel()
	cmd, form := "go test -json ./...", format.TestsGoJSON
	rep := &testout.Report{Form: form, Tests: []testout.Test{{Name: "TestA", Outcome: testout.Passed}}}
	in := RiseInput{Name: "go_tests", Was: "812", Now: "815", Rung: format.Rung{Rung: 1, Command: &cmd, Tests: &form}, Tests: rep,
		Root: "/work/project", Head: headA, Base: "main", MergeBase: headB, Env: []string{"A=1"}, Stop: stop}
	c := NewTests(in)
	if len(got) != 1 || got[0].Root != in.Root || got[0].Base != headB || got[0].Rise != 3 || len(got[0].Head.Tests) != 1 ||
		got[0].Env[0] != "A=1" || got[0].Stop != stop || *got[0].Rung.Command != cmd || got[0].Git != "" {
		t.Fatalf("the check's input: %+v", got)
	}
	if c.NewTests == nil || c.NewTests.Base != headB || c.NewTests.Count != 2 || c.Took != 3*time.Second ||
		c.Summary != "2 new tests, run on the base "+headB[:12]+": 1 failed there, 0 prove nothing new, 1 check not run" ||
		len(c.Lines) != 3 || c.Lines[0] != c.Summary || c.Lines[1] != "  failed on the base: TestA" ||
		!strings.HasPrefix(c.Lines[2], "  check not run (") {
		t.Errorf("the check carried back: %+v", c)
	}
	// On the base branch the merge base is HEAD: the base the task's last result recorded, else none; no tests read.
	in.MergeBase, in.Tests = headA, nil
	in.Prior = &format.Ladder{Rungs: []format.LadderRung{{Rung: 5, Kind: "ver-git", Captures: schema.Object{{Key: "base", Value: headB}}}}}
	NewTests(in)
	in.Prior = nil
	NewTests(in)
	if got[1].Base != headB || got[2].Base != "" || len(got[1].Head.Tests) != 0 {
		t.Errorf("on the base branch: %q, %q", got[1].Base, got[2].Base)
	}
}

func TestRiseOf(t *testing.T) {
	for _, c := range []struct {
		now, was json.Number
		want     int64
	}{
		{"815", "812", 3}, {"812", "812", 0}, {"5", "0", 5}, {"87.5", "87", 1}, {"88", "86.5", 2}, {"3", "4", 0},
		{"many", "1", 0}, {"5", "", 0}, {"1e30", "0", math.MaxInt64}, {"9223372036854775807", "-1", math.MaxInt64},
	} {
		if got := riseOf(c.now, c.was); got != c.want {
			t.Errorf("riseOf(%s, %s) = %d, want %d", c.now, c.was, got, c.want)
		}
	}
}
