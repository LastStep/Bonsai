//go:build !windows

package message

import (
	"os"
	"syscall"
)

// openNoWait opens a file to read without waiting: O_NONBLOCK, so a FIFO put where a regular file stood never holds
// the open until a writer comes (fingerprint then refuses it by its open handle's Stat, before any read). A regular
// file reads as ever with the flag.
func openNoWait(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_RDONLY|syscall.O_NONBLOCK, 0)
}
