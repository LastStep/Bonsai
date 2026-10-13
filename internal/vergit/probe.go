package vergit

// What the file system under a repository can hold (V2b's H2): git records it in the repository's own configuration
// when it makes one (core.ignorecase, core.fileMode, core.symlinks) and trusts it from then on, so a value no diff
// shows can make git look past a change the file system holds: an untracked zz_sub.go beside a tracked zz_sub.GO with
// core.ignorecase=true, an executable bit with core.fileMode=false, a tracked link replaced by a file holding the
// link's text with core.symlinks=false. This package probes the file system itself, once per run and folder, as git
// init does, and pins each setting where the file system can hold the difference; where it cannot (Windows, macOS's
// default case-blind volumes), git's own value stands, since pinning it there would read honest files as changed.

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/LastStep/Bonsai/internal/workspace"
)

// probed holds each folder's settings, probed once per process (a climb reads the tree three times).
var probed sync.Map // folder (forward slashes) -> []string

// fsPins gives the settings that make git see what the file system under dir (the repository's common git folder,
// where git init probes) can hold, given on every call after the fixed pins:
//   - core.ignorecase=false where a file's case twin is not the same file (a case-sensitive folder): git then lists a
//     file whose name differs from a tracked one's only in case;
//   - core.fileMode=true where a file's executable bit, once changed, reads back changed (Linux, macOS; never on
//     Windows, whose files keep no such bit);
//   - core.symlinks=true where a symbolic link can be made and reads back as one (Linux, macOS; never on Windows,
//     where Git for Windows checks links out as files by default and a pin would read each as changed).
//
// The probe makes one temporary file and one link in dir and removes them (workspace.RemoveFile, with its busy
// retry); what is left by a process ended in between is a file git does not read. A folder that cannot be probed
// (not there, not writable) is taken at its strictest on Linux and macOS (all three pinned: a file system that cannot
// hold the difference then reads as changed, never as clean), and as git has it on Windows.
func fsPins(dir string) []string {
	key := filepath.ToSlash(filepath.Clean(dir))
	if v, ok := probed.Load(key); ok {
		return v.([]string)
	}
	pins := probeFS(dir)
	probed.Store(key, pins)
	return pins
}

// probeFS probes the folder dir; see fsPins.
func probeFS(dir string) []string {
	windows := runtime.GOOS == "windows"
	strict := []string{"core.ignorecase=false", "core.fileMode=true", "core.symlinks=true"}
	if windows {
		strict = nil
	}
	if dir == "" {
		return strict
	}
	f, err := os.CreateTemp(dir, "bonsai-probe-")
	if err != nil {
		return strict
	}
	name := f.Name()
	_ = f.Close()
	defer func() { _ = workspace.RemoveFile(name) }()
	var pins []string
	one, err := os.Lstat(name)
	if err != nil {
		return strict
	}
	twin, err := os.Lstat(filepath.Join(dir, strings.ToUpper(filepath.Base(name))))
	if errors.Is(err, fs.ErrNotExist) || (err == nil && !os.SameFile(one, twin)) {
		pins = append(pins, "core.ignorecase=false")
	}
	if windows {
		return pins
	}
	if os.Chmod(name, one.Mode().Perm()^0o100) == nil {
		if again, err := os.Lstat(name); err == nil && again.Mode() != one.Mode() {
			pins = append(pins, "core.fileMode=true")
		}
	}
	link := name + "-link"
	if os.Symlink("bonsai-probe", link) == nil {
		if fi, err := os.Lstat(link); err == nil && fi.Mode()&fs.ModeSymlink != 0 {
			pins = append(pins, "core.symlinks=true")
		}
		_ = workspace.RemoveFile(link)
	}
	return pins
}

// FileSystemSettings gives the settings fsPins probes for the checkout whose top folder is root (its common git folder
// found from its .git, with no git call), as "key=value" for git's -c: for a caller that runs plain git status on the
// checkout and must see what this package's reads see (rung 0's list of changed paths: an untracked file a stale
// core.ignorecase hides). On Linux and macOS a checkout whose git folder cannot be found gets all three.
func FileSystemSettings(root string) []string {
	return append([]string(nil), fsPins(workspace.ReadGitLayout(root).Common)...)
}
