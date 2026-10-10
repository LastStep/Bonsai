// Package sessions reads a session's spans and its subagent runs from the log (design/plan-5.md, 5.2.3 note 1; spec
// section 6, "The two tables"; contract sections 7.5 and 8.1) and builds the rows and hours of .bonsai/sessions.md
// from them. It reads the log through internal/record and writes nothing: check --write (internal/engine) writes the
// table, and bonsai logs prints the files.
//
// The rules of a span, in one place:
//   - a session_start opens a span, unless its source is compact (a compaction is the same session going on) and a
//     span is open; a session_end closes the open span;
//   - a session_start while a span is open (a lost end line) closes the open one at its last record, the one before
//     the new start;
//   - a subagent_start opens a subagent run, closed by the subagent_stop with its subagent_id; a run with no stop ends
//     when its session's span ends;
//   - a span with no end whose file's last record is more than 24 hours old ends at that last record (a killed or
//     crashed session); otherwise it is open: no row yet, and Open says so, for the table and for the cleaner.
//
// Times are the records' own, UTC. A row's minutes are the end minus the start as the row shows them (both cut to the
// minute), so a person can check a row by eye.
package sessions

import (
	"time"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/record"
)

// Stale is how old a file's last record must be for an unended span in it to count as ended at that record.
const Stale = 24 * time.Hour

// Span is one session span or one subagent run.
type Span struct {
	Session  string // the session's full id
	Subagent string // the subagent run's full id; "" for a session span
	Start    time.Time
	End      time.Time // the zero time while the span is open
	Open     bool
	Task     string  // the start record's target, "none" when it has none
	Role     *string // role for a session span, subagent_type for a subagent run
	Model    *string
}

// File is what one log file's records say.
type File struct {
	Spans      []Span // in the order they opened
	Records    int
	Unreadable int
	First      string // the first record's time as written, "" for none
	Last       string // the last record's time as written
	// Ended is nil when the file holds no session span; else whether its last session span has ended.
	Ended *bool
	Task  *string // the first session_start's target, nil for none
	Role  *string // the first session_start's role
}

// Open reports whether the log file at path holds an open span, as of now: the one function that says "open", for the
// table (an open span has no row yet) and for the cleaner (its file is never cleaned). A file that cannot be read is
// an error, and a missing one is an error whose errors.Is(err, fs.ErrNotExist) holds.
func Open(path string, now time.Time) (bool, error) {
	f, err := Read(path, now)
	if err != nil {
		return false, err
	}
	return f.HasOpen(), nil
}

// HasOpen reports whether any of the file's spans is open.
func (f *File) HasOpen() bool {
	for _, s := range f.Spans {
		if s.Open {
			return true
		}
	}
	return false
}

// Read reads the log file at path and finds its spans as of now.
func Read(path string, now time.Time) (*File, error) {
	lf, err := record.ReadLog(path)
	if err != nil {
		return nil, err
	}
	return Spans(lf, record.SessionOf(baseName(path)), now), nil
}

// Spans finds the spans in a log file's records. fileSession is the session the file's name holds, which names the
// spans; "" takes each record's own session.
func Spans(lf *record.LogFile, fileSession string, now time.Time) *File {
	out := &File{Spans: []Span{}, Records: len(lf.Records), Unreadable: lf.Skipped}
	cur := -1 // the open session span
	started := false
	var last time.Time
	for _, r := range lf.Records {
		at, err := time.Parse(timeLayout, r.At)
		if err != nil {
			out.Unreadable++
			out.Records--
			continue
		}
		if out.First == "" {
			out.First = r.At
		}
		out.Last = r.At
		session := fileSession
		if session == "" && r.Session != nil {
			session = *r.Session
		}
		switch r.Event {
		case "session_start":
			if cur >= 0 && deref(r.Source) == "compact" {
				break // the same session going on
			}
			if cur >= 0 {
				out.close(cur, last) // a lost end line: the span ended at its last record
			}
			out.Spans = append(out.Spans, Span{Session: session, Start: at, Open: true, Task: taskOf(r.Target), Role: nz(r.Role), Model: nz(r.Model)})
			cur = len(out.Spans) - 1
			if !started {
				out.Task, out.Role, started = nz(r.Target), nz(r.Role), true
			}
		case "session_end":
			if cur >= 0 {
				out.close(cur, at)
				cur = -1
			}
		case "subagent_start":
			id := deref(r.SubagentID)
			if id == "" || out.run(id) >= 0 {
				break
			}
			out.Spans = append(out.Spans, Span{Session: session, Subagent: id, Start: at, Open: true, Task: taskOf(r.Target), Role: nz(r.SubagentType), Model: nz(r.Model)})
		case "subagent_stop":
			if i := out.run(deref(r.SubagentID)); i >= 0 {
				s := &out.Spans[i]
				s.End, s.Open = clamp(at, s.Start), false
				if s.Model == nil {
					s.Model = nz(r.Model)
				}
			}
		}
		last = at
	}
	if len(out.Spans) > 0 && now.Sub(last) > Stale {
		for i := range out.Spans {
			if out.Spans[i].Open {
				s := &out.Spans[i]
				s.End, s.Open = clamp(last, s.Start), false
			}
		}
	}
	for i := len(out.Spans) - 1; i >= 0; i-- {
		if out.Spans[i].Subagent == "" {
			ended := !out.Spans[i].Open
			out.Ended = &ended
			break
		}
	}
	return out
}

const timeLayout = "2006-01-02T15:04:05.000Z"

// close ends the session span i, and every subagent run still open, at end.
func (f *File) close(i int, end time.Time) {
	for j := range f.Spans {
		if s := &f.Spans[j]; s.Open && (j == i || s.Subagent != "") {
			s.End, s.Open = clamp(end, s.Start), false
		}
	}
}

// run gives the index of the open subagent run with id, or -1.
func (f *File) run(id string) int {
	for i := range f.Spans {
		if s := &f.Spans[i]; s.Open && s.Subagent == id && id != "" {
			return i
		}
	}
	return -1
}

func clamp(end, start time.Time) time.Time {
	if end.Before(start) {
		return start
	}
	return end
}

func taskOf(target *string) string {
	if t := deref(target); t != "" {
		return t
	}
	return "none"
}

// nz gives nil for a text that is not there or is empty: an empty role or model is none.
func nz(s *string) *string {
	if s == nil || *s == "" {
		return nil
	}
	return s
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func baseName(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' || path[i] == '\\' {
			return path[i+1:]
		}
	}
	return path
}

// stamp is a time as a row shows it: UTC, to the minute.
func stamp(t time.Time) string { return t.UTC().Format("2006-01-02 15:04") }

// rowOf is the row of an ended span. A session or subagent id shorter than 8 characters has no row (the schema's
// session column holds 8).
func rowOf(s Span) (format.SessionRow, bool) {
	if len(s.Session) < 8 {
		return format.SessionRow{}, false
	}
	r := format.SessionRow{Session: s.Session[:8], Kind: "session", Task: s.Task, Role: s.Role, Model: s.Model,
		Start: stamp(s.Start), End: stamp(s.End)}
	r.Minutes = int64(s.End.UTC().Truncate(time.Minute).Sub(s.Start.UTC().Truncate(time.Minute)) / time.Minute)
	if s.Subagent != "" {
		id := s.Subagent
		if len(id) > 8 {
			id = id[:8]
		}
		r.Kind, r.Subagent = "subagent", &id
	}
	return r, true
}
