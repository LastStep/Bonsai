//go:build bonsai_test_fault && windows

package guard

import (
	"os"
	"syscall"
)

// crashHard terminates the process at once with the code of an access violation (0xC0000005), as a crashed
// Windows program ends: no deferred call, no Go panic message.
func crashHard() {
	if h, err := syscall.GetCurrentProcess(); err == nil {
		_ = syscall.TerminateProcess(h, 0xC0000005)
	}
	os.Exit(3)
}
