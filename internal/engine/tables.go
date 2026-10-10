package engine

// The two tables and check --write (spec section 6, "The two tables"; contract section 7.5; steps 5.1.8 and 5.2.3).
//
// .bonsai/tasks.md is rebuilt whole from the task files of the main checkout (format 0 and format 1 alike): the
// frontmatter format: bonsai.tasks/1 with its comment, the active task as contract section 13's step 2 finds it (no
// task named: the one task reading running, or none and why; workspace.Active, the one function), then one row per
// task that parses, newest id first (by its number, then by path). The same bytes come out whoever builds them, on
// either side: no map order, forward slashes, LF. A task file that does not parse gives no row (check reports it, and
// the active line says none, naming it).
//
// .bonsai/sessions.md is written empty by init (frontmatter, an empty table, the hours line, an empty table), and
// check --write adds to it: a row for every ended session and subagent run in the main checkout's log that the table
// lacks (internal/sessions: the spans, the rows, the hours). Rows are only added, never rewritten; the one way a row
// leaves is bonsai.yaml's generated.sessions rule (internal/clean, Rows: by default none), with a clean record for
// each in the log once the table is written. A table that does not read back is refused (exit 3, naming the line) and
// nothing is written, so no row is lost. A table written before set 6 (eight columns, no Subagent) is written again
// with nine, every row kept.
//
// A table never grants anything and is never a finding: a tasks table that differs from a rebuild, or a sessions table
// that lacks a row for an ended span in the log, is the warning tables, in a main checkout, a worktree and CI alike
// (checkTables). It is always the main checkout's table that is compared, with a rebuild from the main checkout's task
// files and log, the ones --write would use: a branch never changes a table, so a worktree's own copy lags by design
// and is never read for this.

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/LastStep/Bonsai/internal/clean"
	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/reader"
	"github.com/LastStep/Bonsai/internal/record"
	"github.com/LastStep/Bonsai/internal/sessions"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// WriteCommand is the exact command that rebuilds the tables.
const WriteCommand = "bonsai check --write"

// BuildTasksTable gives the bytes of .bonsai/tasks.md for the task files in main's folder dir (bonsai.yaml's
// documents.task, project-relative).
func BuildTasksTable(main, dir string) ([]byte, error) {
	entries, err := workspace.ReadTaskFiles(main, dir)
	if err != nil {
		return nil, err
	}
	active, err := workspace.Active(workspace.ActiveInput{Main: main, TaskDir: dir})
	if err != nil {
		return nil, err
	}
	type row struct {
		n   int
		p   string
		row format.TaskRow
	}
	var rows []row
	idp := regexp.MustCompile(workspace.TaskIDPattern)
	for _, e := range entries {
		if e.Broken || e.Task == nil {
			continue
		}
		id := e.ID
		if !idp.MatchString(id) {
			id = workspace.NameID(path.Base(e.Path), idp)
		}
		if id == "" {
			continue
		}
		n, _ := strconv.Atoi(strings.TrimPrefix(id, "T-"))
		rows = append(rows, row{n: n, p: e.Path, row: taskRow(id, e.Task)})
	}
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].n != rows[j].n {
			return rows[i].n > rows[j].n
		}
		return rows[i].p < rows[j].p
	})
	t := &format.Tasks{Tasks: []format.TaskRow{}}
	for _, r := range rows {
		t.Tasks = append(t.Tasks, r.row)
	}
	if active.ID != "" {
		id := active.ID
		t.Active.ID = &id
	} else {
		why := strings.TrimPrefix(active.Sentence(), "no active task: ")
		t.Active.Why = &why
	}
	return t.Encode()
}

// taskRow is a task's row. A format-0 task is shown as its file says (the schema: "a format-0 task's own word, shown
// as it is"): a value the format-1 field could not hold is read from the mapping as read.
func taskRow(id string, t *format.Task) format.TaskRow {
	text := func(key, fit string) string {
		if t.Format0 == nil {
			return oneLineCell(fit)
		}
		v, ok := t.Format0.Get(key)
		if !ok {
			return ""
		}
		switch x := reader.JSON(v).(type) {
		case string:
			return oneLineCell(x)
		case json.Number:
			return string(x)
		case bool:
			return strconv.FormatBool(x)
		}
		return ""
	}
	opt := func(key string, fit *string) *string {
		s := ""
		if fit != nil {
			s = *fit
		}
		if s = text(key, s); s == "" {
			return nil
		}
		return &s
	}
	return format.TaskRow{ID: id, Title: text("title", t.Title), Status: text("status", t.Status),
		Lane: opt("lane", t.Lane), Started: opt("started", t.Started), Finished: opt("finished", t.Finished)}
}

// oneLineCell keeps a table cell on one line.
func oneLineCell(s string) string {
	return strings.Join(strings.Fields(strings.ReplaceAll(strings.ReplaceAll(s, "\r", " "), "\n", " ")), " ")
}

// EmptySessionsTable gives the bytes of .bonsai/sessions.md with no rows, as init writes it.
func EmptySessionsTable() ([]byte, error) {
	return (&format.Sessions{Sessions: []format.SessionRow{}, Hours: []format.HoursRow{}}).Encode()
}

// taskDir is the task folder of the checkout holding the task files: main's bonsai.yaml, else the checkout's own.
func taskDir(root, main string, cfg *workspace.Config) string {
	if main != root {
		if m, err := workspace.LoadConfigFull(main); err == nil {
			return m.Full.Documents.Task
		}
	}
	return cfg.Full.Documents.Task
}

// tablesNow is the clock the sessions table reads the log's spans by (the 24-hour rule); a test sets it.
var tablesNow = time.Now

// checkTables is the warning tables: the main checkout's tasks table (this checkout's own, in the main checkout and
// in CI) differs from a rebuild (or is missing), or its sessions table lacks a row for an ended session or subagent
// run in the log (or is missing). A table that does not read is the finding document's, and a task folder or a log
// that cannot be read is no warning.
func (r *CheckResult) checkTables() {
	next := run(WriteCommand)
	if r.Root != r.Main {
		next = "the tables change only in the main checkout, " + filepath.ToSlash(r.Main) + ": from there, run: " + WriteCommand
	}
	r.checkTasksTable(next)
	r.checkSessionsTable(next)
}

func (r *CheckResult) checkTasksTable(next string) {
	want, err := BuildTasksTable(r.Main, taskDir(r.Root, r.Main, r.Config))
	if err != nil {
		return
	}
	raw, exists, err := readFile(r.Main, workspace.TasksTableFile)
	if err != nil {
		return
	}
	whose := workspace.TasksTableFile
	if r.Root != r.Main {
		whose = "the main checkout's " + workspace.TasksTableFile
	}
	switch {
	case !exists:
		r.add("tables", workspace.TasksTableFile, "", whose+" is missing", next)
	default:
		if _, err := format.ReadTasks(raw); err != nil {
			return
		}
		if !bytes.Equal(bytes.ReplaceAll(raw, []byte("\r\n"), []byte("\n")), want) {
			r.add("tables", workspace.TasksTableFile, "", whose+" differs from a rebuild of the task files (the tables lag between moves; a table grants nothing)", next)
		}
	}
}

func (r *CheckResult) checkSessionsTable(next string) {
	found, err := sessions.Found(logDir(r.Main), tablesNow())
	if err != nil {
		return
	}
	raw, exists, err := readFile(r.Main, workspace.SessionsTableFile)
	if err != nil {
		return
	}
	whose := workspace.SessionsTableFile
	if r.Root != r.Main {
		whose = "the main checkout's " + workspace.SessionsTableFile
	}
	if !exists {
		r.add("tables", workspace.SessionsTableFile, "", whose+" is missing", next)
		return
	}
	t, err := format.ReadSessions(raw)
	if err != nil {
		return
	}
	if n := len(sessions.Missing(t.Sessions, found)); n > 0 {
		r.add("tables", workspace.SessionsTableFile, "", fmt.Sprintf("%s lacks %d row(s) for ended sessions or subagent runs in the log (the tables lag between moves; a table grants nothing)", whose, n), next)
	}
}

// logDir is the log folder of the main checkout main.
func logDir(main string) string {
	return filepath.Join(main, filepath.FromSlash(workspace.LocalDir), record.LogFolder)
}

// BuildSessionsTable gives the bytes of .bonsai/sessions.md for the main checkout main: its table now (none yet: an
// empty one) with a row added for every ended session and subagent run in the log that it lacks, then the rows
// bonsai.yaml's generated.sessions cleans taken out (internal/clean, Rows: never a row of a task not done or cut, nor
// one whose span is still in the log), and the hours rebuilt from the rows that stay. It gives the rows cleaned too,
// whose clean records are written once the table is (WriteTables). A table that does not read is the error (exit 3,
// naming the line) and nothing is built. local names the main checkout for the cleaner (workspace.FindLocal).
func BuildSessionsTable(main string, local workspace.Local) ([]byte, []clean.Item, error) {
	table := &format.Sessions{Sessions: []format.SessionRow{}, Hours: []format.HoursRow{}}
	raw, exists, err := readFile(main, workspace.SessionsTableFile)
	if err != nil {
		return nil, nil, err
	}
	if exists {
		if table, err = format.ReadSessions(raw); err != nil {
			return nil, nil, unreadTable(err)
		}
	}
	found, err := sessions.Found(logDir(main), tablesNow())
	if err != nil {
		return nil, nil, errorf("read-failed", ExitRuntime, "check that the log folder (.bonsai/local/log/) can be read, then run: "+WriteCommand,
			"the log cannot be read: %v", err)
	}
	// A rule that does not read cleans nothing; bonsai check names it as bonsai.yaml's problem.
	kept, cleaned, _ := clean.Rows(cleanOptions(local), sessions.Merge(table, found), found)
	b, err := kept.Encode()
	return b, cleaned, err
}

// cleanOptions are the cleaner's options for check --write: the main checkout, the tables' clock.
func cleanOptions(local workspace.Local) clean.Options {
	return clean.Options{Local: local, Now: tablesNow}
}

// unreadTable is the refusal of a table that does not read back: the line, and that no row was dropped.
func unreadTable(err error) *Error {
	what := workspace.SessionsTableFile + " does not read back"
	var re *format.ReadError
	if errors.As(err, &re) {
		what += ", at line " + strconv.Itoa(re.Line) + ": " + re.Msg
	} else {
		what += ": " + err.Error()
	}
	return errorf("bad-file", ExitRuntime,
		"mend that line by hand, or put the file back from git (git restore "+workspace.SessionsTableFile+"), then run: "+WriteCommand,
		"%s; nothing was written, and no row was dropped", ASCII(what))
}

// TablesResult is what WriteTables did.
type TablesResult struct {
	Written []string // the tables whose bytes changed
	Same    []string // the tables already as a rebuild gives them
	Cleaned []string // the sessions rows generated.sessions cleaned, each a clean record in the log
}

// WriteTables is check --write: it rebuilds the tables in the main checkout holding dir. A worktree is refused
// (not-main-checkout, exit 4, naming the main checkout); a folder in no linked checkout is the refusal check gives
// (exit 4); anything that stops the write is exit 3.
func WriteTables(dir string) (*TablesResult, *Error) {
	co, err := workspace.Find(dir)
	if err != nil {
		return nil, findError(err).(*Error)
	}
	cfg, err := workspace.LoadConfigFull(co.Root)
	if err != nil {
		e := fileError(err, "not-linked", "bad-config", ExitInput).(*Error)
		if e.Code == "bad-config" {
			e.Exit = ExitRuntime
		}
		return nil, e
	}
	if co.Root != co.Main {
		return nil, errorf("not-main-checkout", ExitState, "run it in the main checkout, "+filepath.ToSlash(co.Main)+": "+WriteCommand,
			"check --write rebuilds the tables in the main checkout only; this is a worktree of %s", filepath.ToSlash(co.Main)).whose("agent")
	}
	b, err := BuildTasksTable(co.Root, cfg.Full.Documents.Task)
	if err != nil {
		e := Unexpected(err)
		if _, ok := err.(*Error); !ok {
			e = errorf("read-failed", ExitRuntime, "check that the task folder (bonsai.yaml's documents.task) can be read, then run: "+WriteCommand,
				"the tasks table cannot be built: %v", err)
		}
		return nil, e
	}
	// Both tables are built before either is written, so a sessions table that does not read back stops the write
	// with nothing changed.
	local := workspace.FindLocal(co.Root, cfg.ID)
	sb, cleaned, err := BuildSessionsTable(co.Root, local)
	if err != nil {
		if e, ok := err.(*Error); ok {
			return nil, e
		}
		return nil, errorf("read-failed", ExitRuntime, "check that "+workspace.SessionsTableFile+" can be read, then run: "+WriteCommand,
			"the sessions table cannot be built: %v", err)
	}
	res := &TablesResult{}
	for _, t := range []struct {
		rel   string
		bytes []byte
	}{{workspace.TasksTableFile, b}, {workspace.SessionsTableFile, sb}} {
		target := filepath.Join(co.Root, filepath.FromSlash(t.rel))
		old, rerr := os.ReadFile(target)
		if rerr == nil && bytes.Equal(bytes.ReplaceAll(old, []byte("\r\n"), []byte("\n")), t.bytes) {
			res.Same = append(res.Same, t.rel)
			continue
		}
		if err := workspace.WriteFileAtomic(target, t.bytes); err != nil {
			return res, errorf("write-failed", ExitRuntime, "check that "+t.rel+" can be written, then run: "+WriteCommand,
				"%s cannot be written: %v", t.rel, fmt.Sprint(err))
		}
		res.Written = append(res.Written, t.rel)
	}
	// The rows cleaned are gone once the table is written: then each row's clean record, as a file's follows its
	// delete.
	if len(cleaned) > 0 {
		if err := clean.Record(cleanOptions(local), cleaned); err != nil {
			return res, errorf("partly-written", ExitRuntime, "check that .bonsai/local/log/ can be written; the rows are out of the table already, so nothing needs running again",
				"%s was written without %d row(s) generated.sessions cleans, but their clean records cannot be written: %v",
				workspace.SessionsTableFile, len(cleaned), fmt.Sprint(err))
		}
		for _, it := range cleaned {
			res.Cleaned = append(res.Cleaned, it.Target)
		}
	}
	return res, nil
}

// Notes are what check says to a person about the write: one line per table.
func (t *TablesResult) Notes() []string {
	if t == nil {
		return nil
	}
	var out []string
	for _, p := range t.Written {
		out = append(out, "wrote "+p)
	}
	for _, p := range t.Same {
		out = append(out, p+" is already as a rebuild gives it")
	}
	if n := len(t.Cleaned); n > 0 {
		out = append(out, fmt.Sprintf("cleaned %d row(s) of %s by bonsai.yaml's generated.sessions, each a clean record in the log", n, workspace.SessionsTableFile))
	}
	return out
}
