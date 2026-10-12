//go:build windows

package ladder

// The home's ladder lock on Windows (lock.go): LockFileEx on one byte at 4 GiB, far past the holder's words at the
// file's start, so a waiting climb still reads them. The system lets go of the lock when the holder's handle closes,
// its process ending included. LockFileEx and UnlockFileEx are not in package syscall: they are called through kernel32
// (always loaded, a known DLL), loaded lazily as internal/guard/final_windows.go loads its call, so golang.org/x/sys
// stays out of the program.

import (
	"errors"
	"os"
	"syscall"
	"unsafe"
)

var (
	kernel32         = syscall.NewLazyDLL("kernel32.dll")
	procLockFileEx   = kernel32.NewProc("LockFileEx")
	procUnlockFileEx = kernel32.NewProc("UnlockFileEx")
)

const (
	lockfileFailImmediately = 0x1
	lockfileExclusiveLock   = 0x2
	errorLockViolation      = syscall.Errno(33)
	errorIOPending          = syscall.Errno(997)
	stillActive             = 259
	processQueryLimited     = 0x1000
)

// lockedByte is the byte the lock covers: offset 1<<32 (OffsetHigh 1).
func lockedByte() *syscall.Overlapped { return &syscall.Overlapped{OffsetHigh: 1} }

// tryLock takes the lock without waiting: false when another holds it.
func tryLock(f *os.File) (bool, error) {
	if err := procLockFileEx.Find(); err != nil {
		return false, err
	}
	r, _, err := procLockFileEx.Call(f.Fd(), lockfileExclusiveLock|lockfileFailImmediately, 0, 1, 0, uintptr(unsafe.Pointer(lockedByte())))
	if r != 0 {
		return true, nil
	}
	if errors.Is(err, errorLockViolation) || errors.Is(err, errorIOPending) {
		return false, nil
	}
	return false, err
}

// unlock lets go of the lock.
func unlock(f *os.File) {
	if procUnlockFileEx.Find() != nil {
		return
	}
	_, _, _ = procUnlockFileEx.Call(f.Fd(), 0, 1, 0, uintptr(unsafe.Pointer(lockedByte())))
}

// alive reports whether a process with the pid runs: one Windows will not open for this user counts as running.
func alive(pid int) bool {
	if pid <= 0 {
		return false
	}
	h, err := syscall.OpenProcess(processQueryLimited, false, uint32(pid))
	if err != nil {
		return errors.Is(err, syscall.ERROR_ACCESS_DENIED)
	}
	defer func() { _ = syscall.CloseHandle(h) }()
	var code uint32
	if err := syscall.GetExitCodeProcess(h, &code); err != nil {
		return true
	}
	return code == stillActive
}
