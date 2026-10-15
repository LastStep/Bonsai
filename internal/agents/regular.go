package agents

// Reading only regular files (VG's item F, 15 Oct: a FIFO planted under .bonsai/local/agents/ hung bonsai status and
// bonsai messages, and stalled the guard 5 s a call). Every reader of this package opens a file only when it is a
// regular file: a FIFO, a device, a folder, a socket or a link (to anything) is never opened, and the reader says what
// it found instead (ErrNotRegular), so the guard, the stop gate and the views skip and name it. On Linux and macOS the
// open itself never waits (O_NONBLOCK: a FIFO swapped in after the first look opens at once, then fails the second
// look) and follows no link (O_NOFOLLOW); the file opened is held to the one first looked at (os.SameFile). Windows has
// no FIFO or device in its file system, and a link or junction is seen by the first look. Every append (Append, Call,
// AddSpawn) looks first too, so no write waits on a FIFO either.

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
)

// ErrNotRegular is a readers' answer for a file that is there but is not a regular file.
var ErrNotRegular = errors.New("not a regular file")

// kindOf names what a path that is not a regular file is, for a reader's error.
func kindOf(m fs.FileMode) string {
	switch {
	case m&fs.ModeSymlink != 0:
		return "a link"
	case m.IsDir():
		return "a folder"
	case m&fs.ModeNamedPipe != 0:
		return "a FIFO"
	case m&fs.ModeDevice != 0 || m&fs.ModeCharDevice != 0:
		return "a device"
	case m&fs.ModeSocket != 0:
		return "a socket"
	}
	return "not a regular file"
}

// notRegular is ErrNotRegular naming what the file is.
func notRegular(m fs.FileMode) error {
	return fmt.Errorf("it is %s, %w", kindOf(m), ErrNotRegular)
}

// openRegular opens a file for reading only when it is a regular file (this file's comment), with its size.
func openRegular(path string) (*os.File, int64, error) {
	first, err := os.Lstat(path)
	if err != nil {
		return nil, 0, err
	}
	if !first.Mode().IsRegular() {
		return nil, 0, notRegular(first.Mode())
	}
	f, err := os.OpenFile(path, os.O_RDONLY|openFlags, 0)
	if err != nil {
		return nil, 0, err
	}
	fi, err := f.Stat()
	switch {
	case err != nil:
		_ = f.Close()
		return nil, 0, err
	case !fi.Mode().IsRegular():
		_ = f.Close()
		return nil, 0, notRegular(fi.Mode())
	case !os.SameFile(first, fi):
		_ = f.Close()
		return nil, 0, errors.New("it was replaced while it was read")
	}
	return f, fi.Size(), nil
}

// ReadRegular reads a file whole, at most max bytes, only when it is a regular file (this file's comment): the
// guard's and the stop gate's reads of Bonsai's own files beside this package's (a claim, a brief, a report, a skill's
// file). A file larger than max is an error, read no further.
func ReadRegular(path string, max int64) ([]byte, error) {
	f, size, err := openRegular(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	if size > max {
		return nil, fmt.Errorf("it is larger than %d bytes", max)
	}
	raw, err := io.ReadAll(io.LimitReader(f, max+1))
	if err != nil {
		return nil, err
	}
	if int64(len(raw)) > max {
		return nil, fmt.Errorf("it is larger than %d bytes", max)
	}
	return raw, nil
}

// appendable reports whether a file of this package may be appended to: it is not there yet, or it is a regular file.
func appendable(path string) error {
	fi, err := os.Lstat(path)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return nil
	case err != nil:
		return err
	case !fi.Mode().IsRegular():
		return notRegular(fi.Mode())
	}
	return nil
}
