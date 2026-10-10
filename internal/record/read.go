package record

// Reading the log (design/plan-5.md, 5.2.2 note 6): a file's records in order, each through format.ReadLog, the log
// type's reader; a file's last record without reading the file whole; a log folder's session and day files by their
// names, nothing else. bonsai logs, the sessions table and the cleaner read the log through these.
//
// A line that does not read as a record (a torn last line after a crash, or a line written by hand) is skipped and
// counted, never fatal. When a torn line was followed by an append, the two share one line: the torn part is
// skipped and counted, and the record after it is still read. Every record Bonsai writes starts with
// {"format":"bonsai.log/, and no text inside a record can hold that run of bytes (a quote in a JSON string is
// written \"), so the reader starts again at its last place in such a line.

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"io/fs"
	"os"
	"regexp"
	"time"

	"github.com/LastStep/Bonsai/internal/format"
)

// recordStart is how every log record Bonsai writes begins.
var recordStart = []byte(`{"format":"bonsai.log/`)

// LogFile is what one log file holds, as read.
type LogFile struct {
	Records []*format.Log // its records, in the file's order
	Skipped int           // the lines, or parts of a line, that did not read as a record
}

// ReadLog reads the log file at path. A missing file is an error whose errors.Is(err, fs.ErrNotExist) holds.
func ReadLog(path string) (*LogFile, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	out := &LogFile{Records: []*format.Log{}}
	r := bufio.NewReaderSize(f, 64<<10)
	for {
		line, err := r.ReadBytes('\n')
		if len(line) > 0 {
			rec, skipped := readLine(line)
			out.Skipped += skipped
			if rec != nil {
				out.Records = append(out.Records, rec)
			}
		}
		if errors.Is(err, io.EOF) {
			return out, nil
		}
		if err != nil {
			return nil, err
		}
	}
}

// readLine reads one line (its LF, or CRLF, allowed) as a record: the record, or nil, and whether part of the line
// was skipped (1) or none (0). A torn line glued to a record skips the torn part and gives the record.
func readLine(line []byte) (*format.Log, int) {
	line = bytes.TrimSuffix(bytes.TrimSuffix(line, []byte("\n")), []byte("\r"))
	if rec, err := format.ReadLog(line); err == nil {
		return rec, 0
	}
	if i := bytes.LastIndex(line, recordStart); i > 0 {
		if rec, err := format.ReadLog(line[i:]); err == nil {
			return rec, 1
		}
	}
	return nil, 1
}

// lastChunk is how much of a file LastLog reads at a time, from its end: several of the log's 2,048-byte lines.
var lastChunk int64 = 16 << 10

// LastLog reads the last record of the log file at path, reading back from its end only as far as it must: nil when
// the file holds none. A torn last line is passed over to the record before it. A missing file is an error whose
// errors.Is(err, fs.ErrNotExist) holds.
func LastLog(path string) (*format.Log, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}
	return lastRecord(f, fi.Size())
}

// lastRecord is LastLog over the first size bytes of r, read back from the end lastChunk bytes at a time.
func lastRecord(r io.ReaderAt, size int64) (*format.Log, error) {
	pos := size
	var tail []byte // the bytes from pos to the end of the last line not yet tried
	for pos > 0 {
		n := lastChunk
		if n > pos {
			n = pos
		}
		pos -= n
		buf := make([]byte, int(n)+len(tail))
		if _, err := r.ReadAt(buf[:n], pos); err != nil && !errors.Is(err, io.EOF) {
			return nil, err
		}
		copy(buf[n:], tail)
		tail = buf
		for {
			line, rest, ok := lastLine(tail, pos == 0)
			if !ok {
				break
			}
			if rec, _ := readLine(line); rec != nil {
				return rec, nil
			}
			tail = rest
		}
	}
	return nil, nil
}

// lastLine splits b's last line from what comes before it. A line is whole when a LF comes before it, or when b
// starts at the file's start; else ok is false and more of the file is needed.
func lastLine(b []byte, atStart bool) (line, rest []byte, ok bool) {
	if len(b) == 0 {
		return nil, nil, false
	}
	body := bytes.TrimSuffix(b, []byte("\n"))
	i := bytes.LastIndexByte(body, '\n')
	if i < 0 {
		if !atStart {
			return nil, nil, false
		}
		return body, nil, true
	}
	return body[i+1:], b[:i+1], true
}

// The names of a log folder's files (contract §8.5).
var (
	sessionFile = regexp.MustCompile(`^s-(` + sessionChars + `)\.ndjson$`)
	dayFile     = regexp.MustCompile(`^w-([0-9]{4}-[0-9]{2}-[0-9]{2})\.ndjson$`)
)

// LogFiles are a log folder's files, by name, each list sorted.
type LogFiles struct {
	Sessions []string // s-<session>.ndjson: one session's records
	Days     []string // w-<YYYY-MM-DD>.ndjson: a UTC day's records outside a session
}

// ListLog lists the log folder dir's session and day files by their names: regular files only, nothing else (no
// folder, link or other name). A missing folder holds none.
func ListLog(dir string) (LogFiles, error) {
	out := LogFiles{Sessions: []string{}, Days: []string{}}
	entries, err := os.ReadDir(dir)
	if errors.Is(err, fs.ErrNotExist) {
		return out, nil
	}
	if err != nil {
		return out, err
	}
	for _, e := range entries { // os.ReadDir sorts by name
		if !e.Type().IsRegular() {
			continue
		}
		name := e.Name()
		switch {
		case sessionFile.MatchString(name):
			out.Sessions = append(out.Sessions, name)
		case DayOf(name) != "":
			out.Days = append(out.Days, name)
		}
	}
	return out, nil
}

// SessionOf gives the session a session file's name holds, "" for any other name.
func SessionOf(name string) string {
	if m := sessionFile.FindStringSubmatch(name); m != nil {
		return m[1]
	}
	return ""
}

// DayOf gives the day (YYYY-MM-DD) a day file's name holds, "" for any other name or a day that is no date.
func DayOf(name string) string {
	m := dayFile.FindStringSubmatch(name)
	if m == nil {
		return ""
	}
	if _, err := time.Parse("2006-01-02", m[1]); err != nil {
		return ""
	}
	return m[1]
}
