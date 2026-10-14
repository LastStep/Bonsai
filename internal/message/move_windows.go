//go:build windows

package message

import (
	"errors"
	"syscall"
)

// The Windows errors MoveFileW gives when its target is there.
const (
	errorFileExists    = syscall.Errno(80)
	errorAlreadyExists = syscall.Errno(183)
)

// moveNew moves from to to with MoveFileW, which never replaces a file that is there (errTaken then).
func moveNew(from, to string) error {
	f, err := syscall.UTF16PtrFromString(from)
	if err != nil {
		return err
	}
	t, err := syscall.UTF16PtrFromString(to)
	if err != nil {
		return err
	}
	err = syscall.MoveFile(f, t)
	var errno syscall.Errno
	if errors.As(err, &errno) && (errno == errorFileExists || errno == errorAlreadyExists) {
		return errTaken
	}
	return err
}
