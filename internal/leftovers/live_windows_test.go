//go:build windows

package leftovers

import (
	"strings"
	"syscall"
)

// namesDir reports a live process whose command line names dir.
func namesDir(pid int, dir string) bool {
	h, err := syscall.OpenProcess(QueryLimited|syscall.SYNCHRONIZE, false, uint32(pid))
	if err != nil {
		return false
	}
	defer func() { _ = syscall.CloseHandle(h) }()
	if ev, _ := syscall.WaitForSingleObject(h, 0); ev != 0x102 { // WAIT_TIMEOUT: still running
		return false
	}
	line, ok := CommandLine(h)
	return ok && strings.Contains(strings.ToLower(line), strings.ToLower(dir))
}

func endPID(pid int) {
	h, err := syscall.OpenProcess(0x0001|syscall.SYNCHRONIZE, false, uint32(pid)) // PROCESS_TERMINATE
	if err != nil {
		return
	}
	defer func() { _ = syscall.CloseHandle(h) }()
	_ = syscall.TerminateProcess(h, 1)
	_, _ = syscall.WaitForSingleObject(h, 5000)
}
