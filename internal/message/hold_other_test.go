//go:build !windows

package message

import (
	"os"
	"testing"
)

// holdFile opens path as another process would. Linux and macOS have no share modes, so the hold refuses no one: the
// move is made at once there, which is what TestMoveNeverReplacesAndWaitsOutAHold checks on these systems.
func holdFile(t *testing.T, path string) (release func()) {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	return func() { _ = f.Close() }
}
