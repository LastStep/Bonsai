//go:build !windows

package message

import (
	"errors"
	"io/fs"
	"os"
	"syscall"
)

// moveNew links to to from, which never replaces a file that is there (errTaken then), and removes from. A file
// system with no hard links (EPERM, or not supported) gets a rename instead, once no file of that name is there.
func moveNew(from, to string) error {
	err := os.Link(from, to)
	switch {
	case err == nil:
		_ = os.Remove(from) // a staged name left behind is passed over by every reader
		return nil
	case errors.Is(err, fs.ErrExist):
		return errTaken
	case !errors.Is(err, errors.ErrUnsupported) && !errors.Is(err, syscall.EPERM):
		return err
	}
	if _, err := os.Lstat(to); err == nil {
		return errTaken
	}
	return os.Rename(from, to)
}
