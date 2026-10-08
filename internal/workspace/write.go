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
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
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
		return err
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmp, 0o644); err != nil {
		return err
	}
	if err := renameRetry(tmp, path); err != nil {
		return err
	}
	ok = true
	return nil
}

// rename and busy are variables so a test can stand in a busy Windows target on any system.
var (
	rename = os.Rename
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
