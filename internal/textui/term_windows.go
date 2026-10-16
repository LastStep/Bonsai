//go:build windows

package textui

import (
	"os"
	"syscall"
	"unsafe"
)

// kernel32's console calls the syscall package does not have, loaded only when first called.
var (
	kernel32                       = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleScreenBufferInfo = kernel32.NewProc("GetConsoleScreenBufferInfo")
	procSetConsoleMode             = kernel32.NewProc("SetConsoleMode")
)

// enableVirtualTerminal is ENABLE_VIRTUAL_TERMINAL_PROCESSING: the console draws colour codes rather than printing
// them (Windows 10 and later).
const enableVirtualTerminal = 0x0004

// screenBuffer is CONSOLE_SCREEN_BUFFER_INFO.
type screenBuffer struct {
	sizeX, sizeY             int16
	cursorX, cursorY         int16
	attributes               uint16
	left, top, right, bottom int16
	maxX, maxY               int16
}

// termSize reports whether f is a console (GetConsoleMode on it succeeds: NUL, a pipe or a file fails it) and the
// width of its window in columns (0 when it does not say).
func termSize(f *os.File) (bool, int) {
	h := syscall.Handle(f.Fd())
	var mode uint32
	if syscall.GetConsoleMode(h, &mode) != nil {
		return false, 0
	}
	if procGetConsoleScreenBufferInfo.Find() != nil {
		return true, 0
	}
	var info screenBuffer
	if r, _, _ := procGetConsoleScreenBufferInfo.Call(uintptr(h), uintptr(unsafe.Pointer(&info))); r == 0 {
		return true, 0
	}
	return true, int(info.right-info.left) + 1
}

// enableColor turns on the console's virtual terminal processing, so it draws colour codes rather than printing
// them; a console that refuses it (a Windows before 10) gets no colour.
func enableColor(f *os.File) bool {
	h := syscall.Handle(f.Fd())
	var mode uint32
	if syscall.GetConsoleMode(h, &mode) != nil {
		return false
	}
	if mode&enableVirtualTerminal != 0 {
		return true
	}
	if procSetConsoleMode.Find() != nil {
		return false
	}
	r, _, _ := procSetConsoleMode.Call(uintptr(h), uintptr(mode|enableVirtualTerminal))
	return r != 0
}
