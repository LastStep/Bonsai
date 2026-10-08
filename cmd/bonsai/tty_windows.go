//go:build windows

package main

import (
	"os"
	"syscall"
)

// isTerminal reports whether f is a console: GetConsoleMode on it succeeds (NUL, a pipe or a file fails it).
func isTerminal(f *os.File) bool {
	var mode uint32
	return syscall.GetConsoleMode(syscall.Handle(f.Fd()), &mode) == nil
}
