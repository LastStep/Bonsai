//go:build windows

package rungjob

import (
	"context"
	"errors"
	"os"
	"os/exec"
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
	// output), so they are cut: after a second here, PipeCutoff in a real run. The shell must have started the
	// sleeper before the limit ends it; a run in which it had not (the sleeper, which outlives the shell, never
	// writes its pid file) is run again with retryLimit, as in TestTimeout, and that run is checked.
	savedCut := pipeCutoff
	pipeCutoff = time.Second
	t.Cleanup(func() { pipeCutoff = savedCut })
	for _, limit := range []time.Duration{time.Second, retryLimit} {
		sub := testDir(t)
		p := filepath.Join(sub, "left.pid")
		r = Run(context.Background(), Spec{Command: self("sleep", p), Dir: sub, Timeout: limit})
		if limit < retryLimit && !waitFile(p, 3*time.Second) {
			t.Logf("the shell had not started the sleeper within the %s s limit; again with %s s", seconds(limit),
				seconds(retryLimit))
			continue
		}
		if want := "timed out after " + seconds(limit) + " s; ended 1 processes; job: off: CreateJobObjectW: " +
			"refused for the test"; !r.TimedOut || r.Reason != want {
			t.Errorf("timed out %v, reason %q, want %q", r.TimedOut, r.Reason, want)
		}
		if len(r.Notes) != 2 || !strings.Contains(r.Notes[0], "only the rung's shell was ended") ||
			!strings.HasPrefix(r.Notes[1], "output cut 1 s after the command ended") {
			t.Errorf("notes %q", r.Notes)
		}
		readPID(t, p) // left running: the test's cleanup ends it
		return
	}
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

// A pid whose process is gone (it exited between the job's listing and the open) is no member: OpenProcess gives it
// ERROR_INVALID_PARAMETER, and member skips it. A pid the open refuses for another reason (access denied) is a process
// that is there, so it is still reported.
func TestMemberGonePid(t *testing.T) {
	h, err := makeJob()
	if err != nil {
		t.Skipf("no job object here: %v", err)
	}
	t.Cleanup(func() { _ = syscall.CloseHandle(h) })
	r := &rung{handle: h}

	cmd := exec.Command(os.Getenv("COMSPEC"), "/c", "exit 0")
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
	pid := uint32(cmd.Process.Pid)
	// Run has waited and closed the process's handle: its number is held by nothing, until Windows hands it out again.
	if _, err := syscall.OpenProcess(0x1000, false, pid); err == nil {
		t.Skipf("pid %d was taken by another process already", pid)
	} else if !goneProcess(err) {
		t.Fatalf("OpenProcess of a gone pid gave %v, not ERROR_INVALID_PARAMETER", err)
	}
	if p, ok := r.member(pid); ok {
		t.Errorf("a gone pid was reported as a member: %+v", p)
	}
	if goneProcess(syscall.Errno(5)) || goneProcess(errors.New("other")) {
		t.Error("an error other than ERROR_INVALID_PARAMETER read as a gone process")
	}
}
