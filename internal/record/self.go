package record

// The running bonsai, as a log record names it (bonsai_path and bonsai_sha256, set 6; spec §16 row 27): the one
// reader of the binary's own path and hash, which bonsai hook start (internal/recorder), the guard, the stop gate and
// the ladder runner each write on the records that carry them.

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Self gives the running binary's own path, forward slashes, and, when hash is true, its SHA-256 in hex. executable
// finds the binary (os.Executable when nil; a test gives its own). What cannot be found is "": no binary found gives
// both "", and a binary that cannot be read gives its path alone. The path is as the system gives it: a record writes
// it from ~/ under the user's home folder (FromHome, or the writer's own copy of the rule).
func Self(executable func() (string, error), hash bool) (path, sum string) {
	if executable == nil {
		executable = os.Executable
	}
	exe, err := executable()
	if err != nil || exe == "" {
		return "", ""
	}
	path = filepath.ToSlash(exe)
	if !hash {
		return path, ""
	}
	f, err := os.Open(exe)
	if err != nil {
		return path, ""
	}
	defer func() { _ = f.Close() }()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return path, ""
	}
	return path, hex.EncodeToString(h.Sum(nil))
}

// FromHome writes path as a record keeps a binary's path (the log schema's bonsai_path; contract §2.6: no user name
// in the log): from ~/ when it lies under home (letter case ignored on Windows), else as it is, forward slashes either
// way. home "" writes it as it is. The recorder, the guard and the stop gate keep copies of this rule of their own
// (STATE.md, "One home, later"); the ladder runner writes through this one.
func FromHome(path, home string) string {
	if path == "" {
		return ""
	}
	p := filepath.ToSlash(filepath.Clean(path))
	h := strings.TrimSuffix(filepath.ToSlash(filepath.Clean(home)), "/")
	if home == "" || h == "" {
		return p
	}
	prefix := h + "/"
	if len(p) > len(prefix) {
		head := p[:len(prefix)]
		if head == prefix || (runtime.GOOS == "windows" && strings.EqualFold(head, prefix)) {
			return "~/" + p[len(prefix):]
		}
	}
	return p
}
