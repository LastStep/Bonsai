//go:build windows

package rungjob

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

// A job that cannot be made: the rung runs without one and says so, in Job and in its reason, never silently. At a
// timeout only its shell is ended, and the notes say what is not known.
func TestJobOff(t *testing.T) {
	saved := makeJob
	makeJob = func() (syscall.Handle, error) { return 0, errors.New("CreateJobObjectW: refused for the test") }
	t.Cleanup(func() { makeJob = saved })
	dir := testDir(t)
	r := Run(context.Background(), Spec{Command: self("print", "a", "b", "0"), Dir: dir})
	if !r.Green() || r.Job != "off: CreateJobObjectW: refused for the test" ||
		r.Reason != "job: off: CreateJobObjectW: refused for the test" {
		t.Errorf("exit %d, job %q, reason %q", r.Exit, r.Job, r.Reason)
	}
	if r.Ended != nil {
		t.Errorf("ended %+v: with no job, what was left is not known", r.Ended)
	}
	// The sleeper outlives the shell and holds the rung's pipes (a child of cmd inherits them, whatever its own
	// output), so they are cut: after a second here, PipeCutoff in a real run.
	savedCut := pipeCutoff
	pipeCutoff = time.Second
	t.Cleanup(func() { pipeCutoff = savedCut })
	p := filepath.Join(dir, "left.pid")
	r = Run(context.Background(), Spec{Command: self("sleep", p), Dir: dir, Timeout: time.Second})
	if !r.TimedOut || r.Reason != "timed out after 1 s; ended 1 processes; job: off: CreateJobObjectW: refused for the test" {
		t.Errorf("timed out %v, reason %q", r.TimedOut, r.Reason)
	}
	if len(r.Notes) != 2 || !strings.Contains(r.Notes[0], "only the rung's shell was ended") ||
		!strings.HasPrefix(r.Notes[1], "output cut 1 s after the command ended") {
		t.Errorf("notes %q", r.Notes)
	}
	readPID(t, p) // left running: the test's cleanup ends it
}

// A child asking to leave the job is refused by the system.
func TestBreakawayRefused(t *testing.T) {
	dir := testDir(t)
	p := filepath.Join(dir, "left.pid")
	r := Run(context.Background(), Spec{Command: self("spawn", "breakaway", p), Dir: dir, Timeout: time.Minute})
	b, err := os.ReadFile(p + ".refused")
	if err != nil {
		t.Fatalf("the child was not refused (exit %d): %s", r.Exit, r.Out)
	}
	if _, err := os.Stat(p); err == nil {
		t.Errorf("a process left the job: %d", readPID(t, p))
	}
	if r.Exit != 5 {
		t.Errorf("exit %d", r.Exit)
	}
	t.Logf("refused: %s", b)
}
