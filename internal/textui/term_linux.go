//go:build linux

package textui

import (
	"os"
	"syscall"
	"unsafe"
)

// termSize reports whether f is a terminal (a TCGETS on it succeeds: /dev/null, a pipe or a file fails it) and its
// width in columns (TIOCGWINSZ; 0 when it does not say).
func termSize(f *os.File) (bool, int) {
	fd := f.Fd()
	var t syscall.Termios
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, syscall.TCGETS, uintptr(unsafe.Pointer(&t))); errno != 0 {
		return false, 0
	}
	var ws struct{ Row, Col, X, Y uint16 }
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, syscall.TIOCGWINSZ, uintptr(unsafe.Pointer(&ws))); errno != 0 {
		return true, 0
	}
	return true, int(ws.Col)
}

// enableColor reports that a terminal here draws colour codes as they are.
func enableColor(*os.File) bool { return true }
