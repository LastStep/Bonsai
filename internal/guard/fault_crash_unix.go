//go:build bonsai_test_fault && !windows

package guard

import (
	"os"
	"syscall"
)

// crashHard kills the process with SIGKILL: no deferred call, no Go panic message, no exit code of its own (a shell
// sees 137).
func crashHard() {
	_ = syscall.Kill(os.Getpid(), syscall.SIGKILL)
	os.Exit(137)
}
