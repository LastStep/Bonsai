// Package workspace reads and writes the files Bonsai keeps for a project, and finds the places they live:
//
//   - bonsai.yaml (bonsai.workspace/1, spec §6): config.go;
//   - a pack's bonsai/pack.yaml (bonsai.pack/1, spec §5): pack.go;
//   - the lock, .bonsai/lock.json (bonsai.lock/1, contract §14), read and written: lock.go;
//   - the Bonsai home and a workspace's machine folder (contract §3): home.go;
//   - the checkout and its main checkout, found through git (contract §3): checkout.go;
//   - project-relative paths and atomic writes, with Windows' busy renames retried: paths.go, write.go.
//
// The YAML files are read by internal/reader under format 1 only (format 0 is step 5.1); the lock is held to its
// schema in formats/schemas, embedded by package formats, so its lists (file kinds) have one home. Each schema of
// bonsai.yaml and pack.yaml is step 5.1's in full: this package reads what the walking skeleton's parts 2 to 5 use
// and keeps every other key as read (contract §2.2: a reader keeps an unknown field), and each file's Go type says
// where it stops.
package workspace

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/LastStep/Bonsai/internal/reader"
)

// Error is a problem with one of Bonsai's files. Its message names the file and line and the next step, in ASCII.
type Error struct {
	File string // the file, as a person finds it: "bonsai.yaml", ".bonsai/lock.json", "bonsai/pack.yaml"
	Line int    // the line, or 0 when the problem has none
	Code string // the reader's reason code when the reader refused the file, else ""
	Msg  string // what is wrong
	Next string // what to do
	Err  error  // the cause, when there is one (an os error: errors.Is(err, fs.ErrNotExist) works)
}

func (e *Error) Error() string {
	where := e.File
	if e.Line > 0 {
		where += " line " + strconv.Itoa(e.Line)
	}
	code := ""
	if e.Code != "" {
		code = " (" + e.Code + ")"
	}
	return asciiOnly(fmt.Sprintf("%s: %s%s; next: %s", where, e.Msg, code, e.Next))
}

func (e *Error) Unwrap() error { return e.Err }

// readYAML reads a format-1 YAML file and checks its format: line names want. Format 0 and other formats are
// refused: bonsai.yaml and pack.yaml are new with format 1.
func readYAML(file string, raw []byte, want, next string) (*reader.Map, error) {
	r := reader.ReadYAML(raw)
	switch r.Outcome {
	case reader.Refused:
		return nil, &Error{File: file, Line: r.Refusal.Line, Code: r.Refusal.Code, Msg: r.Refusal.Message, Next: r.Refusal.Next}
	case reader.Format0:
		return nil, &Error{File: file, Line: 1, Msg: "no format: line first, so it reads as format 0, which this file never had",
			Next: "start the file with format: " + want}
	}
	if v, _ := r.Value.Get("format"); v != want {
		e, _ := r.Value.Entry("format")
		return nil, &Error{File: file, Line: e.Line, Msg: fmt.Sprintf("format is %s, not %s", showValue(v), want), Next: next}
	}
	return r.Value, nil
}

func showValue(v any) string {
	if s, ok := v.(string); ok {
		return strconv.QuoteToASCII(s)
	}
	return fmt.Sprintf("%v", v)
}

// fields reads typed values out of a mapping, keeping the first problem, so a reader of one file can ask for each
// field in turn and check the error once.
type fields struct {
	file string
	next string // the next step for a value of the wrong kind
	err  *Error
}

func (f *fields) fail(line int, format string, args ...any) {
	if f.err == nil {
		f.err = &Error{File: f.file, Line: line, Msg: fmt.Sprintf(format, args...), Next: f.next}
	}
}

// text returns a text field, "" when it is missing or null; required fails on those.
func (f *fields) text(m *reader.Map, key, where string, line int, required bool) string {
	e, ok := m.Entry(key)
	if !ok || e.Value == nil {
		if required {
			f.fail(line, "%s has no %s", where, key)
		}
		return ""
	}
	s, isText := e.Value.(string)
	if !isText {
		f.fail(e.Line, "%s's %s is %s, not text (quote it)", where, key, kindOf(e.Value))
		return ""
	}
	if required && s == "" {
		f.fail(e.Line, "%s's %s is empty", where, key)
	}
	return s
}

// list returns a list field's items as entries (null or missing: none); each item keeps the field's line.
func (f *fields) list(m *reader.Map, key, where string) ([]any, int) {
	e, ok := m.Entry(key)
	if !ok || e.Value == nil {
		return nil, e.Line
	}
	l, isList := e.Value.([]any)
	if !isList {
		f.fail(e.Line, "%s's %s is %s, not a list", where, key, kindOf(e.Value))
		return nil, e.Line
	}
	return l, e.Line
}

// texts returns a list field of text items.
func (f *fields) texts(m *reader.Map, key, where string) []string {
	items, line := f.list(m, key, where)
	out := []string{}
	for i, it := range items {
		s, ok := it.(string)
		if !ok || s == "" {
			f.fail(line, "%s's %s item %d is %s, not text (quote it)", where, key, i+1, kindOf(it))
			continue
		}
		out = append(out, s)
	}
	return out
}

// mapping returns a mapping field, nil when it is missing or null.
func (f *fields) mapping(m *reader.Map, key, where string) *reader.Map {
	e, ok := m.Entry(key)
	if !ok || e.Value == nil {
		return nil
	}
	mm, isMap := e.Value.(*reader.Map)
	if !isMap {
		f.fail(e.Line, "%s's %s is %s, not a mapping", where, key, kindOf(e.Value))
	}
	return mm
}

func kindOf(v any) string {
	switch x := v.(type) {
	case nil:
		return "null"
	case string:
		if x == "" {
			return "empty text"
		}
		return "text"
	case bool:
		return "a boolean"
	case int64, reader.Decimal:
		return "a number"
	case []any:
		return "a list"
	case *reader.Map:
		return "a mapping"
	}
	return "unknown"
}

var errNoGit = errors.New("git is not on the PATH")
