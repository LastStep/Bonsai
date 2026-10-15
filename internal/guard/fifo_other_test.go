//go:build !unix

package guard

import "testing"

// mkfifo skips the test: Windows has no FIFO in its file system.
func mkfifo(t *testing.T, _ string) {
	t.Helper()
	t.Skip("Windows has no FIFO in its file system")
}
