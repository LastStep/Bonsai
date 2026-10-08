//go:build windows

package workspace

import (
	"errors"
	"syscall"
)

// The Windows errors a rename meets while another process holds its target open.
const (
	errorAccessDenied     = syscall.Errno(5)
	errorSharingViolation = syscall.Errno(32)
	errorLockViolation    = syscall.Errno(33)
)

// isBusy reports whether a rename failed because another process holds the file.
func isBusy(err error) bool {
	var errno syscall.Errno
	if !errors.As(err, &errno) {
		return false
	}
	return errno == errorAccessDenied || errno == errorSharingViolation || errno == errorLockViolation
}
