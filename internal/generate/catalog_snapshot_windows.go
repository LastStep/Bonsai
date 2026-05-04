//go:build windows

package generate

import "os"

// openSnapshotFile opens the catalog.json target for writing.
//
// Windows lacks `syscall.O_NOFOLLOW`; symlink-following defense degraded
// on this platform. Acceptable: catalog.json is regenerable + non-secret.
func openSnapshotFile(absPath string) (*os.File, error) {
	return os.OpenFile(absPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
}
