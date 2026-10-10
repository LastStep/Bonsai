//go:build !windows

package clean

import (
	"errors"
	"testing"
)

// errHeld stands in for Windows' sharing violation.
var errHeld = errors.New("the process cannot access the file because it is being used by another process")

// holdFile stands in a hold: Linux and macOS have no share modes, and let a file another process holds open be
// deleted, so the delete is made to fail as Windows' does, and busy to know that error, until release. The Windows
// run holds the file for real (hold_windows_test.go).
func holdFile(t *testing.T, path string) (release func()) {
	t.Helper()
	oldRemove, oldBusy := removeFile, busy
	removeFile = func(p string) error {
		if p == path {
			return errHeld
		}
		return oldRemove(p)
	}
	busy = func(err error) bool { return errors.Is(err, errHeld) || oldBusy(err) }
	return func() { removeFile, busy = oldRemove, oldBusy }
}
