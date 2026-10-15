//go:build unix

package agents

import "syscall"

// mkfifo plants a FIFO at path.
func mkfifo(path string) error { return syscall.Mkfifo(path, 0o644) }
