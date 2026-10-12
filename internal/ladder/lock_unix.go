//go:build unix

package ladder

// The home's ladder lock on Linux and macOS (lock.go): flock on the whole file, which the system lets go of when the
// holder's last descriptor closes, its process ending included. Through package syscall.

import (
	"errors"
	"os"
	"syscall"
)

// tryLock takes the lock without waiting: false when another holds it.
func tryLock(f *os.File) (bool, error) {
	err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	switch {
	case err == nil:
		return true, nil
	case errors.Is(err, syscall.EWOULDBLOCK), errors.Is(err, syscall.EAGAIN), errors.Is(err, syscall.EINTR):
		return false, nil
	}
	return false, err
}

// unlock lets go of the lock.
func unlock(f *os.File) { _ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN) }

// alive reports whether a process with the pid runs (one this user may not signal counts as running).
func alive(pid int) bool {
	if pid <= 0 {
		return false
	}
	err := syscall.Kill(pid, 0)
	return err == nil || errors.Is(err, syscall.EPERM)
}
