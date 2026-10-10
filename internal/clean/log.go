package clean

// The log kind (contract §8.5): s-<session>.ndjson and w-<date>.ndjson in .bonsai/local/log/ (internal/record's
// names, their one home), aged by the last record that reads, and kept while it holds an open span or an ended span
// with no row yet in .bonsai/sessions.md.
//
// The spans are internal/sessions' (sessions.Spans: the rules of a span, and of "open", in one place), over the
// file's span records and its last record. Reading a file through sessions.Read parses every line, about 60
// microseconds each, so a long session's file of 20,000 lines takes over a second: more than a session end's whole
// budget, and it would never be cleaned. So readSpans parses in full only the lines that can hold a span's record (a
// line naming one of the four span events, or holding an escape that could spell one: a superset), and the file's
// last record that reads, found from the same pass. Every other record changes no span: sessions.Spans reads them only
// for the time of the file's last record (the 24-hour rule) and for the end of a span a lost end line closed, which no
// row's key holds. TestReadSpansIsSessionsRead holds the two equal, open spans and row keys alike, on every span rule.

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/record"
	"github.com/LastStep/Bonsai/internal/sessions"
	"github.com/LastStep/Bonsai/internal/workspace"
)

var logKind = fileKind{
	name:    Log,
	folder:  record.LogFolder,
	match:   func(name string) bool { return record.SessionOf(name) != "" || record.DayOf(name) != "" },
	age:     logAge,
	protect: logProtected,
}

// logAge is a log file's last record's at (record.LastLog: a torn last line is passed over to the record before it,
// as every reader of the log does), its modification time when no record reads.
func logAge(_ *run, f *file) time.Time {
	l, err := record.LastLog(f.path)
	if err == nil && l != nil {
		if at, err := time.Parse(record.AtLayout, l.At); err == nil {
			return at
		}
	}
	return f.mtime
}

// logProtected keeps a file holding an open span, or an ended span whose row is not yet in .bonsai/sessions.md.
func logProtected(r *run, f *file) (bool, error) {
	sp, err := readSpans(f.path, f.name, r.now)
	if errors.Is(err, fs.ErrNotExist) {
		return true, nil // gone since it was listed: another session's cleaner took it
	}
	if err != nil {
		return true, fmt.Errorf("clean: %s is kept, its spans cannot be read: %w", f.target, err)
	}
	if sp.HasOpen() {
		return true, nil
	}
	rows := r.tableRows()
	for _, s := range sp.Spans {
		if key, ok := rowKey(s); ok && !rows[key] {
			return true, nil
		}
	}
	return false, nil
}

// tableRows are the keys of the rows in the main checkout's .bonsai/sessions.md, read once a run. A table that is
// missing or does not read has none, so every file holding an ended span is kept until check --write writes it.
func (r *run) tableRows() map[string]bool {
	if r.rows != nil {
		return r.rows
	}
	r.rows = map[string]bool{}
	raw, err := os.ReadFile(sessionsTablePath(r.main))
	if err != nil {
		return r.rows
	}
	t, err := format.ReadSessions(raw)
	if err != nil {
		return r.rows
	}
	for _, row := range t.Sessions {
		r.rows[sessions.Key(row)] = true
	}
	return r.rows
}

// rowKey is the key (sessions.Key) of the row an ended span has in the table: the session id's first 8 characters,
// the subagent run's first 8, the start to the minute (UTC). ok is false for a span that has no row (a session id
// under 8 characters, which the table's column cannot hold), or one still open. internal/sessions builds its rows the
// same way; TestReadSpansIsSessionsRead holds the keys equal to sessions.Found's.
func rowKey(s sessions.Span) (string, bool) {
	if s.Open || len(s.Session) < 8 {
		return "", false
	}
	row := format.SessionRow{Session: s.Session[:8], Start: s.Start.UTC().Format("2006-01-02 15:04")}
	if s.Subagent != "" {
		id := s.Subagent
		if len(id) > 8 {
			id = id[:8]
		}
		row.Subagent = &id
	}
	return sessions.Key(row), true
}

// spanEvents are the events a span is made of (internal/sessions).
var spanEvents = map[string]bool{"session_start": true, "session_end": true, "subagent_start": true, "subagent_stop": true}

// spanMarks are what a line holding a span's record must contain: one of the events as written, or an escape, which
// could spell one (JSON lets "session_start" stand for it). Bonsai writes neither escape, so it parses only the
// span records and the odd line with a quote or a backslash in its text.
var spanMarks = [][]byte{[]byte("session_start"), []byte("session_end"), []byte("subagent_start"), []byte("subagent_stop"), []byte(`\`)}

// logRecordStart is how every log record Bonsai writes begins (internal/record's reader starts again there on a line
// where a torn write was glued to the next record).
var logRecordStart = []byte(`{"format":"bonsai.log/`)

// tailLines is how many of a file's last lines readSpans keeps to find its last record that reads: more torn lines
// than that at a file's end and the file is kept, unjudged.
const tailLines = 8

// readSpans finds a log file's spans as sessions.Read does, as of now, parsing only the lines that can hold a span's
// record and the file's last record that reads (this file's comment). One pass, so what it judges is one view of the
// file.
func readSpans(path, name string, now time.Time) (*sessions.File, error) {
	fh, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = fh.Close() }()
	type held struct {
		no   int
		line []byte
	}
	lf := &record.LogFile{Records: []*format.Log{}}
	lastSpan := -1 // the line of the last span record kept
	var tail []held
	br := bufio.NewReaderSize(fh, 64<<10)
	for no := 0; ; no++ {
		line, err := br.ReadBytes('\n')
		if len(line) > 0 {
			if len(bytes.TrimSpace(line)) > 0 {
				if len(tail) == tailLines {
					tail = tail[1:]
				}
				tail = append(tail, held{no, line})
			}
			if mayHoldSpan(line) {
				if rec := readLogLine(line); rec != nil && spanEvents[rec.Event] {
					lf.Records = append(lf.Records, rec)
					lastSpan = no
				}
			}
		}
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
	}
	// The file's last record that reads, from the end of the same pass.
	var last *format.Log
	lastNo := -1
	for i := len(tail) - 1; i >= 0; i-- {
		if rec := readLogLine(tail[i].line); rec != nil {
			last, lastNo = rec, tail[i].no
			break
		}
	}
	switch {
	case last == nil && len(tail) == tailLines:
		return nil, fmt.Errorf("its last %d lines do not read as records", tailLines)
	case last == nil:
		// No record reads at all, so no span either.
	case lastNo == lastSpan:
		// The last record is a span's, already kept.
	case lastNo > lastSpan && !spanEvents[last.Event]:
		lf.Records = append(lf.Records, last)
	default:
		return nil, errors.New("its last record and its span records do not agree")
	}
	if last != nil {
		if _, err := time.Parse(record.AtLayout, last.At); err != nil {
			return nil, errors.New("its last record's time does not read")
		}
	}
	return sessions.Spans(lf, record.SessionOf(name), now), nil
}

// mayHoldSpan reports a line that could hold a span's record.
func mayHoldSpan(line []byte) bool {
	for _, m := range spanMarks {
		if bytes.Contains(line, m) {
			return true
		}
	}
	return false
}

// readLogLine reads one line as a log record as internal/record's reader does: the line whole (its LF, or CRLF,
// allowed), else the record glued after a torn part; nil when neither reads.
func readLogLine(line []byte) *format.Log {
	line = bytes.TrimSuffix(bytes.TrimSuffix(line, []byte("\n")), []byte("\r"))
	if rec, err := format.ReadLog(line); err == nil {
		return rec
	}
	if i := bytes.LastIndex(line, logRecordStart); i > 0 {
		if rec, err := format.ReadLog(line[i:]); err == nil {
			return rec
		}
	}
	return nil
}

// sessionsTablePath is the main checkout's .bonsai/sessions.md.
func sessionsTablePath(main string) string {
	return filepath.Join(main, filepath.FromSlash(workspace.SessionsTableFile))
}
