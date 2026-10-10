package record

// The one append path for every Bonsai writer of .bonsai/local/ (design/plan-5.md, 5.2.2 note 1; contract §2.5:
// "records are appended whole, one line per write, with today's retry on Windows' busy errors").
//
// A line is written whole, in one Write of the line with its LF, to a file opened with O_APPEND, never longer than
// its format's cap (2,048 bytes for the log, 8,192 for asks). The file is first opened with O_CREATE|O_EXCL, so the
// writer knows when it made it (bonsai hook start makes a session's file first). On Linux an O_APPEND write of a line
// this size lands whole at the end; on Windows Go opens an O_APPEND file for appending only (FILE_APPEND_DATA), so
// each write lands whole at the end too. TestAppendsFromEightProcesses shows both, on each system: eight processes
// appending 250 records each to one file, every line whole. A file Windows reports busy (an antivirus scan, an
// indexer, a process holding it with no write sharing) is retried for up to 1 s, waiting 5 ms and doubling to 80 ms,
// as the guard's own writer does. Chosen over a lock file, which is slower on every hook and leaves a stale lock
// behind a crash.
//
// Not tested, and so not promised: a project on a Windows drive mounted in WSL (/mnt/c/...), written through WSL's
// bridge to that drive.

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// LogFolder and AsksFolder are the log's and the asks' folders inside .bonsai/local/ (contract §3).
const (
	LogFolder  = "log"
	AsksFolder = "asks"
)

// openWait is how long an append keeps retrying while Windows reports its file busy; openFile opens it; busy tells a
// busy error; ensureGitignore restores .bonsai/.gitignore. Variables for the tests.
var (
	openWait        = time.Second
	openFile        = os.OpenFile
	busy            = workspace.IsBusy
	ensureGitignore = workspace.EnsureGitignore
)

// Append appends one line to the file rel (forward slashes, inside .bonsai/local/: "log/s-<session>.ndjson") of the
// checkout at main, the checkout workspace.FindLocal names. The line is one JSON record: one LF, at its end, and at
// most max bytes with it (format.Format.MaxLine). It restores a missing .bonsai/.gitignore first (spec §6), makes
// the folders it needs, and reports whether this append made the file.
//
// A .gitignore that cannot be restored does not stop the line: the line is still written, and the error returned
// names both when both failed.
func Append(main, rel string, line []byte, max int) (made bool, err error) {
	if err := workspace.CheckRelPath(rel); err != nil {
		return false, fmt.Errorf("record: %v", err)
	}
	switch {
	case max <= 0:
		return false, fmt.Errorf("record: no cap given for %s", rel)
	case len(line) > max:
		return false, fmt.Errorf("record: a line of %d bytes is over the cap of %d for %s", len(line), max, rel)
	case len(line) < 2 || line[len(line)-1] != '\n' || bytes.IndexByte(line, '\n') != len(line)-1:
		return false, fmt.Errorf("record: a line for %s must be one record ending in its only line feed", rel)
	}
	_, gerr := ensureGitignore(main)
	if gerr != nil {
		gerr = fmt.Errorf("%s cannot be restored: %w", workspace.GitignoreFile, gerr)
	}
	made, err = appendLine(filepath.Join(main, filepath.FromSlash(workspace.LocalDir), filepath.FromSlash(rel)), line)
	return made, errors.Join(err, gerr)
}

// appendLine opens path for appending (making it, and its folder, when missing) and writes line in one Write.
func appendLine(path string, line []byte) (bool, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return false, err
	}
	var f *os.File
	made := false
	for try := 0; ; try++ {
		var err error
		f, err = openRetry(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL|os.O_APPEND)
		if err == nil {
			made = true
			break
		}
		if !errors.Is(err, fs.ErrExist) {
			return false, err
		}
		f, err = openRetry(path, os.O_WRONLY|os.O_APPEND)
		if err == nil {
			break
		}
		// The file went between the two opens (a cleaner removed it): make it again, once.
		if !errors.Is(err, fs.ErrNotExist) || try > 0 {
			return false, err
		}
	}
	n, werr := f.Write(line)
	cerr := f.Close()
	switch {
	case werr != nil:
		return made, werr
	case n != len(line):
		return made, fmt.Errorf("record: %d of %d bytes written to %s", n, len(line), filepath.ToSlash(path))
	}
	return made, cerr
}

// openRetry opens a file, retrying while Windows reports it busy, for up to openWait: 5 ms, doubling to 80 ms.
func openRetry(name string, flag int) (*os.File, error) {
	deadline := time.Now().Add(openWait)
	delay := 5 * time.Millisecond
	for {
		f, err := openFile(name, flag, 0o644)
		if err == nil || !busy(err) || time.Now().After(deadline) {
			return f, err
		}
		time.Sleep(delay)
		if delay < 80*time.Millisecond {
			delay *= 2
		}
	}
}

// sessionPattern is what a session id must be to name its log file: letters, digits, _ and -, at most 128 (Claude
// Code's are UUIDs). Anything else could reach outside the log folder or name a file Windows cannot hold.
const sessionChars = `[A-Za-z0-9_-]{1,128}`

var sessionPattern = regexp.MustCompile(`^` + sessionChars + `$`)

// LogFileName names the log file a record goes to (contract §8.5): s-<session>.ndjson for a session's records,
// w-<YYYY-MM-DD>.ndjson, the record's UTC day, for one with no session. A session id that cannot name a file is
// refused.
func LogFileName(session string, at time.Time) (string, error) {
	if session == "" {
		return "w-" + at.UTC().Format("2006-01-02") + ".ndjson", nil
	}
	if !sessionPattern.MatchString(session) {
		return "", fmt.Errorf("record: the session id %q cannot name a log file: it is not letters, digits, _ and -", session)
	}
	return "s-" + session + ".ndjson", nil
}

// WriteLog appends a log record to the main checkout's log (Append): to its session's file, or to its day's when it
// has no session. The record is encoded by Line, so it is cut to fit. It reports whether this append made the file.
func WriteLog(main string, l *format.Log) (bool, error) {
	at, err := time.Parse(AtLayout, l.At)
	if err != nil {
		return false, fmt.Errorf("record: the record's at %q is not %s", l.At, AtLayout)
	}
	session := ""
	if l.Session != nil {
		session = *l.Session
	}
	name, err := LogFileName(session, at)
	if err != nil {
		return false, err
	}
	line, err := Line(l)
	if err != nil {
		return false, err
	}
	return Append(main, LogFolder+"/"+name, line, format.MustLookup("log").MaxLine())
}
