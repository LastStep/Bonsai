package workspace

// Writing a file whole or not at all, with Windows' busy renames retried (CLAUDE.md's Windows rules; contract §2.5:
// "with today's retry on Windows' busy errors").

import (
	"os"
	"path/filepath"
	"time"
)

// WriteFileAtomic writes data to path through a temporary file in the same folder and a rename, so a reader sees
// the old bytes or the new ones, never part. It makes the folder when it is missing. The temporary file is removed
// on any failure.
func WriteFileAtomic(path string, data []byte) error {
	tmp, err := StageFile(path, data)
	if err != nil {
		return err
	}
	if err := CommitStaged(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

// StageFile writes data to a temporary file beside path (".<name>.tmp-<random>", in the same folder, so a rename
// later replaces path whole) and returns its name: the first half of an all-or-nothing write (spec §6: "every write
// is staged first; renames follow"). It makes the folder when it is missing. On failure nothing is left but the
// folder.
func StageFile(path string, data []byte) (string, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	f, err := os.CreateTemp(dir, "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return "", err
	}
	tmp := f.Name()
	ok := false
	defer func() {
		if !ok {
			_ = os.Remove(tmp)
		}
	}()
	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		return "", err
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		return "", err
	}
	if err := f.Close(); err != nil {
		return "", err
	}
	if err := os.Chmod(tmp, 0o644); err != nil {
		return "", err
	}
	ok = true
	return tmp, nil
}

// CommitStaged renames a staged file over its target, retrying while Windows reports the target busy.
func CommitStaged(tmp, path string) error {
	return renameRetry(tmp, path)
}

// RemoveFile removes a file, retrying while Windows reports it busy, as renameRetry does; a file already gone is
// no error.
func RemoveFile(path string) error {
	deadline := time.Now().Add(renameWait)
	delay := 5 * time.Millisecond
	for {
		err := remove(path)
		if err == nil || os.IsNotExist(err) {
			return nil
		}
		if !busy(err) || time.Now().After(deadline) {
			return err
		}
		time.Sleep(delay)
		if delay < 160*time.Millisecond {
			delay *= 2
		}
	}
}

// rename and busy are variables so a test can stand in a busy Windows target on any system.
var (
	rename = os.Rename
	remove = os.Remove
	busy   = isBusy
)

// renameWait is how long renameRetry keeps trying while the target is busy (a variable for the tests).
var renameWait = 2 * time.Second

// renameRetry renames, retrying while the error says another process holds the target (an antivirus scan, a
// search indexer, an editor), for up to renameWait, waiting 5 ms and doubling to 160 ms. Any other error returns
// at once.
func renameRetry(from, to string) error {
	deadline := time.Now().Add(renameWait)
	delay := 5 * time.Millisecond
	for {
		err := rename(from, to)
		if err == nil || !busy(err) || time.Now().After(deadline) {
			return err
		}
		time.Sleep(delay)
		if delay < 160*time.Millisecond {
			delay *= 2
		}
	}
}
