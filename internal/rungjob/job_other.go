//go:build !linux && !windows

package rungjob

// Elsewhere (macOS and the BSDs build, untested): the rung runs in its own process group, ended at its end. What
// it ended is not listed.

import (
	"os"
	"os/exec"
	"runtime"
	"syscall"
	"time"
)

type rung struct{ pid int }

func newRung() *rung { return &rung{} }

func (r *rung) job() string { return "on" }

func (r *rung) setupCost() time.Duration { return 0 }

func (r *rung) prepare(env []string) []string { return env }

func (r *rung) command(command, dir string, env []string) *exec.Cmd {
	cmd := exec.Command("/bin/sh", "-c", command)
	cmd.Dir = dir
	cmd.Env = env
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	return cmd
}

func (r *rung) launch(cmd *exec.Cmd) (<-chan error, error) {
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	r.pid = cmd.Process.Pid
	exited := make(chan error, 1)
	go func() { exited <- cmd.Wait() }()
	return exited, nil
}

func (r *rung) end(cmd *exec.Cmd, alive bool) ([]Process, []string) {
	// While the shell is not reaped (alive) its group's number is its own; after, only while a member holds it,
	// which signal 0 asks.
	if alive || syscall.Kill(-r.pid, 0) == nil {
		_ = syscall.Kill(-r.pid, syscall.SIGKILL)
	}
	return nil, []string{"the rung's processes are not listed on " + runtime.GOOS}
}

func (r *rung) finish() []string { return nil }

func signalNumber(ps *os.ProcessState) int {
	if ws, ok := ps.Sys().(syscall.WaitStatus); ok && ws.Signaled() {
		return int(ws.Signal())
	}
	return 0
}
