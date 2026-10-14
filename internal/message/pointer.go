package message

// Pointers (contract §22: "files it points to", so long output travels as a file and is never pasted). A --pointer
// is given as a path relative to the folder the command runs in, an absolute path, or one written from ~/ under the
// person's home folder. It must name a file that is there (a folder is not one), inside the checkout the command
// runs in (written project-relative) or under the home (written from ~/), never elsewhere, and its path must hold no
// colon or backslash (forward slashes everywhere, a name Windows can hold). Its sha256 is the file's bytes with every
// CRLF read as LF (workspace.HashLF), so a Windows checkout's line endings never read as a change.
//
// At --show a pointer is read again, from the checkout --show runs in (or the home): unchanged, changed (its
// fingerprint is not the one written) or missing. A project-relative pointer is read in the checkout of the command
// that reads it, so a builder's report read in the main checkout before its branch has landed shows its files as they
// stand there.

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/workspace"
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
	switch {
	case given == "~" || strings.HasPrefix(given, "~/"):
		if herr != nil || home == "" {
			return format.MessagePointer{}, missing("is written from ~/, but the home folder cannot be found")
		}
		abs = filepath.Join(home, filepath.FromSlash(strings.TrimPrefix(strings.TrimPrefix(given, "~"), "/")))
	case !filepath.IsAbs(given):
		abs = filepath.Join(cwd, given)
	}
	abs = filepath.Clean(abs)
	fi, err := os.Stat(abs)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return format.MessagePointer{}, missing("names no file that is there")
	case err != nil:
		return format.MessagePointer{}, &Error{Code: "read-failed", Exit: ExitRuntime,
			What: "--pointer " + ascii(given) + " cannot be read: " + oneLine(err), Next: "check the file's permissions, then run the command again"}
	case fi.IsDir():
		return format.MessagePointer{}, missing("is a folder: a pointer names a file")
	}
	rel, ok := within(p.Local.Root, abs)
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
	raw, err := os.ReadFile(abs)
	if err != nil {
		return format.MessagePointer{}, &Error{Code: "read-failed", Exit: ExitRuntime,
			What: "--pointer " + ascii(given) + " cannot be read: " + oneLine(err), Next: "check the file's permissions, then run the command again"}
	}
	return format.MessagePointer{Path: rel, SHA256: workspace.HashLF(raw)}, nil
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

// PointerState reads a pointer's file again, from the checkout the reading command runs in or the home: unchanged,
// changed or missing. A path with a .. segment (which bonsai message never writes) is missing, never read.
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
	}
	fi, err := os.Stat(path)
	if err != nil || fi.IsDir() {
		return PointerMissing
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return PointerMissing
	}
	if workspace.HashLF(raw) != ptr.SHA256 {
		return PointerChanged
	}
	return PointerUnchanged
}
