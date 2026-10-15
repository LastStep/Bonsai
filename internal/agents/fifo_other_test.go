//go:build !unix

package agents

import "errors"

// mkfifo fails: Windows has no FIFO in its file system.
func mkfifo(string) error { return errors.New("no FIFO on this system") }
