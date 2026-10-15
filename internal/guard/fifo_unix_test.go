//go:build unix

package guard

import (
	"syscall"
	"testing"
)

// mkfifo plants a FIFO at path.
func mkfifo(t *testing.T, path string) {
	t.Helper()
	if err := syscall.Mkfifo(path, 0o644); err != nil {
		t.Fatal(err)
	}
}
