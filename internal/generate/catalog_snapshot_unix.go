//go:build !windows

package generate

import (
	"os"
	"syscall"
)

// openSnapshotFile opens the catalog.json target for writing with
// O_NOFOLLOW so the kernel refuses to follow a symlink at the path.
// Defends against an attacker pre-planting a symlink at the target
// to redirect the write at an arbitrary file (e.g. ~/.ssh/authorized_keys).
func openSnapshotFile(absPath string) (*os.File, error) {
	return os.OpenFile(absPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC|syscall.O_NOFOLLOW, 0644)
}
