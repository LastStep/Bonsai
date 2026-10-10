package engine

// The tasks table and check --write (spec section 6, "The two tables"; contract section 7.5; step 5.1.8).
//
// .bonsai/tasks.md is rebuilt whole from the task files of the main checkout (format 0 and format 1 alike): the
// frontmatter format: bonsai.tasks/1 with its comment, the active task as contract section 13's step 2 finds it (no
// task named: the one task reading running, or none and why; workspace.Active, the one function), then one row per
// task that parses, newest id first (by its number, then by path). The same bytes come out whoever builds them, on
// either side: no map order, forward slashes, LF. A task file that does not parse gives no row (check reports it, and
// the active line says none, naming it).
//
// The sessions table is written empty by init (frontmatter, an empty table, the hours line, an empty table); step
// 5.2.3 fills it from the log and adds it to check --write.
//
// A table never grants anything and is never a finding: a tasks table that differs from a rebuild is the warning
// tables, in a main checkout, a worktree and CI alike (checkTables). It is always the main checkout's table that is
// compared, with a rebuild from the main checkout's task files, the ones --write would use: a branch never changes a
// table, so a worktree's own copy lags by design and is never read for this.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/reader"
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

// checkTables is the warning tables: the main checkout's tasks table (this checkout's own, in the main checkout and
// in CI) differs from a rebuild (or is missing). A table that does not read is the finding document's, and a task folder that cannot be read is document's too: no warning then.
func (r *CheckResult) checkTables() {
	want, err := BuildTasksTable(r.Main, taskDir(r.Root, r.Main, r.Config))
	if err != nil {
		return
	}
	raw, exists, err := readFile(r.Main, workspace.TasksTableFile)
	if err != nil {
		return
	}
	next, whose := run(WriteCommand), workspace.TasksTableFile
	if r.Root != r.Main {
		next = "the tables change only in the main checkout, " + filepath.ToSlash(r.Main) + ": from there, run: " + WriteCommand
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

// TablesResult is what WriteTables did.
type TablesResult struct {
	Written []string // the tables whose bytes changed
	Same    []string // the tables already as a rebuild gives them
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
	res := &TablesResult{}
	target := filepath.Join(co.Root, filepath.FromSlash(workspace.TasksTableFile))
	old, rerr := os.ReadFile(target)
	if rerr == nil && bytes.Equal(bytes.ReplaceAll(old, []byte("\r\n"), []byte("\n")), b) {
		res.Same = append(res.Same, workspace.TasksTableFile)
		return res, nil
	}
	if err := workspace.WriteFileAtomic(target, b); err != nil {
		return nil, errorf("write-failed", ExitRuntime, "check that "+workspace.TasksTableFile+" can be written, then run: "+WriteCommand,
			"%s cannot be written: %v", workspace.TasksTableFile, fmt.Sprint(err))
	}
	res.Written = append(res.Written, workspace.TasksTableFile)
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
	return out
}
