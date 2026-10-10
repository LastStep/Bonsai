package main

// bonsai logs (spec section 4, section 8; contract section 8; plan-5 5.2.3 note 5): the log files of the main
// checkout's .bonsai/local/log/, or one session's or one day's records. It reads and writes nothing else: the records
// are printed as stored (they were redacted when written). With --json it prints bonsai.logs/1 (format.Logs).
//
// The files in --json are oldest first (the schema's order); the plain listing is newest first.

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/LastStep/Bonsai/internal/engine"
	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/record"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/sessions"
	"github.com/LastStep/Bonsai/internal/workspace"
)

func init() { register(logsWord) }

// logsNow is the clock a session's open or ended state is read by (the 24-hour rule); a test sets it.
var logsNow = time.Now

// minSessionPrefix is how short a --session prefix may be: the sessions table's 8 characters.
const minSessionPrefix = 8

var logsWord = &Word{
	Name:    "logs",
	Order:   11,
	Title:   "the log of this project (contract section 8).",
	Summary: "list the log's files, or print one session's or one day's records",
	Args:    "[flags]",
	About: `It reads the log in the main checkout's .bonsai/local/log/ (a worktree reads its main checkout's) and writes
nothing. With no filter it lists the files, newest first: a session's id, its first and last time, whether it ended
or is open, its task and role, and how many records and unreadable lines it holds; a day's file (outside events,
clean records) by its date and records. With --session or --day it prints that file's records, one JSON line each,
as stored: they were redacted when they were written.
`,
	Flags: []Flag{
		{Name: "--json", Help: "print the bonsai.logs/1 document (for programs) instead of plain text"},
		{Name: "--session", Value: "<id>", Need: "a session's id, or a prefix of at least 8 characters",
			Help: "print one session's records; <id> is a full session id or a prefix of at least 8 characters\n(the sessions table's); none or several matching is an error naming them"},
		{Name: "--day", Value: "<date>", Need: "a UTC date (YYYY-MM-DD)",
			Help: "print the records of one UTC day's file: the events outside a session and the clean\nrecords. A day with no file prints none"},
	},
	Exits: []Exit{
		{Code: 0, Means: "the log was read (a log with no files prints none)"},
		{Code: 2, Means: "bad input (a flag logs does not take, --session and --day together, a prefix under 8 characters,\nor a date that is no date); bonsai.yaml refused"},
		{Code: 3, Means: "the log cannot be read, or the answer cannot be printed"},
		{Code: 4, Means: "not in a git checkout, not a linked checkout, or --session matches no session or several (session-not-found)"},
	},
	Examples: []string{"bonsai logs", "bonsai logs --session 6d1e2f3a --json", "bonsai logs --day 2026-10-08"},
	Refused:  func(c *call, e *engine.Error) encoder { return &format.Logs{Error: e.Object()} },
	Run:      runLogs,
}

func runLogs(c *call) int {
	if len(c.rest) > 0 {
		return c.refuse(c.flagError("logs takes no %+q", c.rest[0]))
	}
	if c.has("--session") && c.has("--day") {
		return c.refuse(c.flagError("logs takes --session or --day, not both"))
	}
	if c.has("--session") && len(c.value("--session")) < minSessionPrefix {
		return c.refuse(&engine.Error{Code: "bad-value", Exit: exitInput,
			What: fmt.Sprintf("--session needs a full session id or a prefix of at least %d characters", minSessionPrefix),
			Next: "run `bonsai logs` to list the sessions, then give --session at least their first 8 characters"})
	}
	if c.has("--day") {
		if _, err := time.Parse("2006-01-02", c.value("--day")); err != nil {
			return c.refuse(&engine.Error{Code: "bad-value", Exit: exitInput, What: fmt.Sprintf("--day %+q is not a date", c.value("--day")),
				Next: "give --day a UTC date as YYYY-MM-DD, such as 2026-10-08"})
		}
	}
	dir, err := os.Getwd()
	if err != nil {
		return c.fail(&engine.Error{Code: "read-failed", Exit: exitRuntime, What: "the current folder cannot be read",
			Next: "run it from a folder inside the project"})
	}
	co, err := workspace.Find(dir)
	if err != nil {
		return c.fail(engine.FindError(err))
	}
	cfg, err := workspace.LoadConfig(co.Root)
	if err != nil {
		return c.fail(engine.ConfigError(err))
	}
	local := workspace.FindLocal(co.Root, cfg.ID)
	ref := &format.WorkspaceRef{ID: cfg.ID, Name: cfg.Name, Root: filepath.ToSlash(local.Main)}
	logDir := filepath.Join(local.Dir(), record.LogFolder)
	names, err := record.ListLog(logDir)
	if err != nil {
		return c.fail(readFailed(err))
	}
	now := logsNow()
	doc := &format.Logs{Workspace: ref}
	switch {
	case c.has("--session"):
		id, e := matchSession(names.Sessions, c.value("--session"))
		if e != nil {
			return c.fail(e)
		}
		return c.records(doc, logDir, "s-"+id+".ndjson", now)
	case c.has("--day"):
		name := "w-" + c.value("--day") + ".ndjson"
		if !contains(names.Days, name) {
			doc.Files, doc.Records = []format.LogFile{}, []schema.Object{}
			if c.json {
				return c.printDoc(doc, exitOK)
			}
			return c.print("no log file for " + c.value("--day") + "\n")
		}
		return c.records(doc, logDir, name, now)
	}
	files, err := listFiles(logDir, names, now)
	if err != nil {
		return c.fail(readFailed(err))
	}
	doc.Files = files
	if c.json {
		return c.printDoc(doc, exitOK)
	}
	return c.print(filesText(files))
}

func readFailed(err error) *engine.Error {
	return &engine.Error{Code: "read-failed", Exit: exitRuntime, What: "the log cannot be read: " + engine.ASCII(err.Error()),
		Next: "check that .bonsai/local/log/ can be read, then run: bonsai logs"}
}

func contains(list []string, s string) bool {
	i := sort.SearchStrings(list, s)
	return i < len(list) && list[i] == s
}

// matchSession gives the id of the one session file that a full id or a prefix names: an exact id first, else the one
// session the prefix begins; none or several is session-not-found (exit 4), naming them.
func matchSession(files []string, want string) (string, *engine.Error) {
	var ids, hits []string
	for _, f := range files {
		id := record.SessionOf(f)
		ids = append(ids, id)
		if id == want {
			return id, nil
		}
		if strings.HasPrefix(id, want) {
			hits = append(hits, id)
		}
	}
	if len(hits) == 1 {
		return hits[0], nil
	}
	next := "run `bonsai logs` to list the sessions, then give --session more of the id"
	if len(hits) == 0 {
		what := fmt.Sprintf("no session in the log starts with %+q", want)
		if len(ids) > 0 {
			what += fmt.Sprintf(" (the log holds %d)", len(ids))
		}
		return "", &engine.Error{Code: "session-not-found", Exit: engine.ExitState, What: what, Next: next}
	}
	return "", &engine.Error{Code: "session-not-found", Exit: engine.ExitState,
		What: fmt.Sprintf("%+q matches %d sessions: %s", want, len(hits), strings.Join(hits, ", ")), Next: next}
}

// records prints one file's records: its entry in files, and each record as stored.
func (c *call) records(doc *format.Logs, logDir, name string, now time.Time) int {
	entry, lf, err := fileEntry(logDir, name, now)
	if err != nil {
		return c.fail(readFailed(err))
	}
	doc.Files = []format.LogFile{entry}
	doc.Records = []schema.Object{}
	logf := format.MustLookup("log")
	var text strings.Builder
	for _, r := range lf.Records {
		o, err := logf.Document(r)
		if err != nil {
			return c.fail(engine.Unexpected(err))
		}
		doc.Records = append(doc.Records, o)
		if !c.json {
			line, err := schema.EncodeLine(o)
			if err != nil {
				return c.fail(engine.Unexpected(err))
			}
			text.Write(line)
		}
	}
	if c.json {
		return c.printDoc(doc, exitOK)
	}
	return c.print(text.String())
}

// fileEntry reads one log file and gives its entry for the listing, with the records it was read from.
func fileEntry(logDir, name string, now time.Time) (format.LogFile, *record.LogFile, error) {
	path := filepath.Join(logDir, name)
	lf, err := record.ReadLog(path)
	if err != nil {
		return format.LogFile{}, nil, err
	}
	e := format.LogFile{File: name}
	if id := record.SessionOf(name); id != "" {
		e.Session = &id
		f := sessions.Spans(lf, id, now)
		e.Records, e.Unreadable = int64(f.Records), int64(f.Unreadable)
		e.First, e.Last = text(f.First), text(f.Last)
		e.Ended, e.Task, e.Role = f.Ended, f.Task, f.Role
		return e, lf, nil
	}
	day := record.DayOf(name)
	e.Day = &day
	e.Records, e.Unreadable = int64(len(lf.Records)), int64(lf.Skipped)
	if n := len(lf.Records); n > 0 {
		e.First, e.Last = text(lf.Records[0].At), text(lf.Records[n-1].At)
	}
	return e, lf, nil
}

func text(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// listFiles gives every log file's entry, oldest first: a session's file by its first record, a day's by its day;
// equal times by file name, so the order is the same on every run.
func listFiles(logDir string, names record.LogFiles, now time.Time) ([]format.LogFile, error) {
	type item struct {
		key   string
		entry format.LogFile
	}
	var items []item
	for _, name := range append(append([]string{}, names.Sessions...), names.Days...) {
		e, _, err := fileEntry(logDir, name, now)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue // removed since it was listed
			}
			return nil, err
		}
		key := record.DayOf(name) // a day file by its day
		if e.Session != nil && e.First != nil {
			key = *e.First
		}
		items = append(items, item{key, e})
	}
	sort.Slice(items, func(i, j int) bool { // the file names are distinct
		if items[i].key != items[j].key {
			return items[i].key < items[j].key
		}
		return items[i].entry.File < items[j].entry.File
	})
	out := []format.LogFile{}
	for _, it := range items {
		out = append(out, it.entry)
	}
	return out, nil
}

// filesText is the plain listing, newest first, one line per file.
func filesText(files []format.LogFile) string {
	if len(files) == 0 {
		return "no log files in .bonsai/local/log/\n"
	}
	var b strings.Builder
	for i := len(files) - 1; i >= 0; i-- {
		f := files[i]
		counts := fmt.Sprintf("%d records, %d unreadable", f.Records, f.Unreadable)
		if f.Session != nil {
			state := "-"
			if f.Ended != nil {
				state = map[bool]string{true: "ended", false: "open"}[*f.Ended]
			}
			fmt.Fprintf(&b, "session %s  %s to %s  %s  task %s  role %s  %s\n", *f.Session, dash(f.First), dash(f.Last), state,
				dash(f.Task), dash(f.Role), counts)
			continue
		}
		fmt.Fprintf(&b, "day %s  %s to %s  %s\n", *f.Day, dash(f.First), dash(f.Last), counts)
	}
	return engine.ASCII(b.String())
}

func dash(s *string) string {
	if s == nil || *s == "" {
		return "-"
	}
	return *s
}
