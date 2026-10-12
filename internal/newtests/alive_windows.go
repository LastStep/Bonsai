package newtests

import (
	"errors"
	"syscall"
)

const (
	processQueryLimited = 0x1000 // PROCESS_QUERY_LIMITED_INFORMATION
	stillActive         = 259    // STILL_ACTIVE
)

// alive reports whether a process with the pid runs: one Windows will not open for this user counts as running, so a
// stale worktree's owner that may still run is never swept.
func alive(pid int) bool {
	if pid <= 0 {
		return false
	}
	h, err := syscall.OpenProcess(processQueryLimited, false, uint32(pid))
	if err != nil {
		return errors.Is(err, syscall.ERROR_ACCESS_DENIED)
	}
	defer func() { _ = syscall.CloseHandle(h) }()
	var code uint32
	if err := syscall.GetExitCodeProcess(h, &code); err != nil {
		return true
	}
	return code == stillActive
}
