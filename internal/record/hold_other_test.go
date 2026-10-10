//go:build !windows

package record

import (
	"os"
	"testing"
)

// holdFile opens path as another process would. Linux and macOS have no share modes, so the hold refuses no one:
// TestAppendWaitsOutAHeldFile's append succeeds at once there, which is what it checks on these systems.
func holdFile(t *testing.T, path string) (release func()) {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	return func() { _ = f.Close() }
}
