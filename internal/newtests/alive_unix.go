//go:build unix

package newtests

import (
	"errors"
	"syscall"
)

// alive reports whether a process with the pid runs (one this user may not signal counts as running), so a stale
// worktree's owner that may still run is never swept.
func alive(pid int) bool {
	if pid <= 0 {
		return false
	}
	err := syscall.Kill(pid, 0)
	return err == nil || errors.Is(err, syscall.EPERM)
}
