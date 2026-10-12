//go:build linux

package leftovers

import (
	"strings"
	"syscall"
	"time"
)

// namesDir reports a live process whose arguments name dir.
func namesDir(pid int, dir string) bool {
	if s, ok := ReadStat(pid); !ok || s.Dead() {
		return false
	}
	for _, a := range Argv(pid) {
		if strings.Contains(a, dir) {
			return true
		}
	}
	return false
}

func endPID(pid int) {
	_ = syscall.Kill(pid, syscall.SIGKILL)
	for i := 0; i < 200; i++ {
		var ws syscall.WaitStatus
		_, _ = syscall.Wait4(pid, &ws, syscall.WNOHANG, nil)
		if s, ok := ReadStat(pid); !ok || s.Dead() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
}
