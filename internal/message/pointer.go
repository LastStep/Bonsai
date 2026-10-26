package message

// Pointers (contract §22: "files it points to", so long output travels as a file and is never pasted). A --pointer
// is given as a path relative to the folder the command runs in, an absolute path, or one written from ~/ under the
// person's home folder. It must name a regular file that is there, inside the checkout the command runs in (written
// project-relative) or under the home (written from ~/), never elsewhere, and its path must hold no colon or backslash
// (forward slashes everywhere, a name Windows can hold). Its sha256 is the file's bytes with every CRLF read as LF
// (workspace.HashLF's rule), so a Windows checkout's line endings never read as a change.
//
// Only a regular file is ever read (fingerprint): the path is looked at first (os.Stat, which follows links), and
// anything else, a folder, a device (/dev/zero, Windows' nul), a pipe (a FIFO) or a socket, or a link to any of them,
// is refused as pointer-missing, saying what it is, and never opened; on Windows a name the system keeps for a device
// (nul, con, com1, nul.txt, in any case) is refused before anything is looked at. The file is then opened without
// waiting (a FIFO put in its place since never holds the open), the open file must be the same regular file the path
// named, and its bytes are hashed as they stream through a fixed buffer: no file, however large, is ever held whole,
// and the stream is the only bound (no plan or contract names a size).
//
// At --show a pointer is read again the same way, from the checkout --show runs in (or the home): unchanged, changed
// (its fingerprint is not the one written) or missing (no file there, or one that is not a regular file now). A
// project-relative pointer is read in the checkout of the command that reads it (one into .bonsai/local/ in the main
// checkout, where Bonsai keeps those files), so a builder's report read in the
// main checkout before its branch has landed shows its files as they stand there.

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/LastStep/Bonsai/internal/format"
)

// The states of a pointer's file at --show.
const (
	PointerUnchanged = "unchanged"
	PointerChanged   = "changed"
	PointerMissing   = "missing"
)

func (p Place) home() (string, error) {
	if p.Home != nil {
		return p.Home()
	}
	return os.UserHomeDir()
}

// Pointer reads one --pointer as given, from the folder cwd: its path as a message holds it and its fingerprint now.
func Pointer(p Place, given, cwd string) (format.MessagePointer, *Error) {
	missing := func(why string) *Error {
		return &Error{Code: "pointer-missing", Exit: ExitInput, What: "--pointer " + ascii(given) + " " + why,
			Next: "point to a file inside this checkout or under the home folder that is there, or leave the --pointer out"}
	}
	if strings.TrimSpace(given) == "" {
		return format.MessagePointer{}, missing("is empty")
	}
	home, herr := p.home()
	abs := given
	base := p.Local.Root // the checkout a project-relative path is written against
	switch {
	case given == "~" || strings.HasPrefix(given, "~/"):
		if herr != nil || home == "" {
			return format.MessagePointer{}, missing("is written from ~/, but the home folder cannot be found")
		}
		abs = filepath.Join(home, filepath.FromSlash(strings.TrimPrefix(strings.TrimPrefix(given, "~"), "/")))
	case !filepath.IsAbs(given):
		abs = filepath.Join(cwd, given)
		if inLocal(given) && p.Local.Main != "" {
			// Bonsai keeps .bonsai/local/ in the main checkout only: a linked worktree's pointer into it is read there.
			abs, base = filepath.Join(p.Local.Main, filepath.FromSlash(cleanSlash(given))), p.Local.Main
		}
	}
	abs = filepath.Clean(abs)
	if name := deviceIn(abs); name != "" {
		return format.MessagePointer{}, missing("names " + ascii(name) + ", a name Windows keeps for a device, not a file")
	}
	fi, err := os.Stat(abs)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return format.MessagePointer{}, missing("names no file that is there")
	case err != nil:
		return format.MessagePointer{}, &Error{Code: "read-failed", Exit: ExitRuntime,
			What: "--pointer " + ascii(given) + " cannot be read: " + oneLine(err), Next: "check the file's permissions, then run the command again"}
	case !fi.Mode().IsRegular():
		return format.MessagePointer{}, missing(notRegular(fi))
	}
	rel, ok := within(base, abs)
	if !ok && herr == nil && home != "" {
		if r, inHome := within(home, abs); inHome {
			rel, ok = "~/"+r, true
		}
	}
	if !ok {
		return format.MessagePointer{}, missing("is outside the project and the home folder")
	}
	if strings.ContainsAny(rel, `:\`) {
		return format.MessagePointer{}, badValue("--pointer "+ascii(given)+"'s path holds a colon or a backslash, which a message's path may not",
			"rename the file, or point to another")
	}
	sum, why, err := fingerprint(abs, fi)
	switch {
	case err != nil:
		return format.MessagePointer{}, &Error{Code: "read-failed", Exit: ExitRuntime,
			What: "--pointer " + ascii(given) + " cannot be read: " + oneLine(err), Next: "check the file's permissions, then run the command again"}
	case why != "":
		return format.MessagePointer{}, missing(why)
	}
	return format.MessagePointer{Path: rel, SHA256: sum}, nil
}

// cleanSlash is a relative path written with forward slashes and cleaned.
func cleanSlash(given string) string {
	return filepath.ToSlash(filepath.Clean(filepath.FromSlash(given)))
}

// inLocal reports whether a relative path's first folders are .bonsai/local: Bonsai's own files, kept in the main
// checkout (a ladder result, the messages), where a linked worktree reads them from.
func inLocal(given string) bool {
	return strings.HasPrefix(cleanSlash(given), ".bonsai/local/")
}

// notRegular says what a path that is not a regular file is (its mode, links followed), for a pointer refused.
func notRegular(fi fs.FileInfo) string {
	switch m := fi.Mode(); {
	case m.IsDir():
		return "is a folder: a pointer names a file"
	case m&fs.ModeNamedPipe != 0:
		return "is a pipe (a FIFO), or a link to one, not a regular file: a pointer names a file"
	case m&fs.ModeSocket != 0:
		return "is a socket, or a link to one, not a regular file: a pointer names a file"
	case m&fs.ModeDevice != 0:
		return "is a device, or a link to one, not a regular file: a pointer names a file"
	}
	return "is not a regular file (a device, a pipe or a socket, or a link to one): a pointer names a file"
}

// fingerprint gives the sha256 of the regular file at path, its line endings read as LF, streamed through a fixed
// buffer (hashLF), never read whole. fi is the path's os.Stat, already a regular file's. The file is opened without
// waiting (openNoWait) and must be, once open, the regular file fi is: one put in its place since is never read, and
// why then says so. err is a read that failed.
func fingerprint(path string, fi fs.FileInfo) (sum, why string, err error) {
	f, err := openNoWait(path)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return "", "names no file that is there", nil
	case err != nil:
		return "", "", err
	}
	defer func() { _ = f.Close() }()
	open, err := f.Stat()
	switch {
	case err != nil:
		return "", "", err
	case !open.Mode().IsRegular():
		return "", "was replaced while it was read: it " + notRegular(open), nil
	case !os.SameFile(fi, open):
		return "", "was replaced by another file while it was read: point to it again", nil
	}
	sum, err = hashLF(f)
	return sum, "", err
}

// hashBuffer is the bytes hashLF reads at a time: the most it ever holds of a file.
const hashBuffer = 64 << 10

// hashLF gives the sha256 of what r holds with every CRLF read as LF (workspace.HashLF's rule: a lone CR is kept, as
// no line ending git makes), as 64 lower-case hex characters, streamed through a buffer of hashBuffer bytes: a CR at
// the end of one read is held until the next shows whether an LF follows it.
func hashLF(r io.Reader) (string, error) {
	h := sha256.New()
	buf := make([]byte, hashBuffer)
	held := false // a CR from the end of the last read, not yet written
	for {
		n, err := r.Read(buf)
		if chunk := buf[:n]; n > 0 {
			if held && chunk[0] != '\n' {
				h.Write([]byte{'\r'})
			}
			held = chunk[len(chunk)-1] == '\r'
			if held {
				chunk = chunk[:len(chunk)-1]
			}
			for {
				i := bytes.Index(chunk, []byte("\r\n"))
				if i < 0 {
					break
				}
				h.Write(chunk[:i])
				chunk = chunk[i+1:] // from the LF
			}
			h.Write(chunk)
		}
		switch {
		case errors.Is(err, io.EOF):
			if held {
				h.Write([]byte{'\r'})
			}
			return hex.EncodeToString(h.Sum(nil)), nil
		case err != nil:
			return "", err
		}
	}
}

// pointerDevices are the names Windows keeps for devices beside an address's (deviceNames): the console's own and
// the ports written with a superscript digit.
var pointerDevices = map[string]bool{"conin$": true, "conout$": true, "clock$": true,
	"com\u00b9": true, "com\u00b2": true, "com\u00b3": true, "lpt\u00b9": true, "lpt\u00b2": true, "lpt\u00b3": true}

// deviceIn gives the first element of path that Windows keeps for a device (nul, NUL, nul.txt, com1, CONIN$: the
// name before its first dot, its case and trailing spaces not counted), and "" for none. Only on Windows, where such a
// name opens a device wherever it stands; elsewhere it is a file's name like any other, read only if it is a regular
// file.
func deviceIn(path string) string {
	if runtime.GOOS != "windows" {
		return ""
	}
	for _, seg := range strings.Split(filepath.ToSlash(path[len(filepath.VolumeName(path)):]), "/") {
		base, _, _ := strings.Cut(seg, ".")
		base = strings.ToLower(strings.TrimRight(base, " "))
		if deviceNames[base] || pointerDevices[base] {
			return seg
		}
	}
	return ""
}

// within gives path relative to base, forward slashes, and whether it lies inside base: compared as written first,
// then with links and Windows' short names resolved on both sides.
func within(base, path string) (string, bool) {
	if base == "" {
		return "", false
	}
	if rel, ok := inside(base, path); ok {
		return rel, true
	}
	b, err1 := filepath.EvalSymlinks(base)
	q, err2 := filepath.EvalSymlinks(path)
	if err1 != nil || err2 != nil {
		return "", false
	}
	return inside(b, q)
}

func inside(base, path string) (string, bool) {
	rel, err := filepath.Rel(base, path)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return "", false
	}
	return filepath.ToSlash(rel), true
}

// PointerState reads a pointer's file again, from the checkout the reading command runs in or the home, as Pointer
// reads one (fingerprint: a regular file only, streamed): unchanged, changed, or missing (no file there, or one that is
// not a regular file now: a folder, a device, a pipe, a link to one). A path with a .. segment (which bonsai message
// never writes) is missing, never read.
func PointerState(p Place, ptr format.MessagePointer) string {
	for _, seg := range strings.Split(ptr.Path, "/") {
		if seg == ".." { // a hand-made file's pointer out of the project or the home is never read
			return PointerMissing
		}
	}
	var path string
	if rest, ok := strings.CutPrefix(ptr.Path, "~/"); ok {
		home, err := p.home()
		if err != nil || home == "" {
			return PointerMissing
		}
		path = filepath.Join(home, filepath.FromSlash(rest))
	} else {
		path = filepath.Join(p.Local.Root, filepath.FromSlash(ptr.Path))
		if inLocal(ptr.Path) && p.Local.Main != "" {
			path = filepath.Join(p.Local.Main, filepath.FromSlash(ptr.Path))
		}
	}
	if deviceIn(path) != "" {
		return PointerMissing
	}
	fi, err := os.Stat(path)
	if err != nil || !fi.Mode().IsRegular() {
		return PointerMissing
	}
	sum, why, err := fingerprint(path, fi)
	switch {
	case err != nil || why != "":
		return PointerMissing
	case sum != ptr.SHA256:
		return PointerChanged
	}
	return PointerUnchanged
}
