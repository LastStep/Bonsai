package clean

// The test fixture: a linked project in a temporary folder, its log, asks and ladder files and its sessions table
// written fresh with made-up values, each file's modification time set to its own clock as Bonsai's writes leave it.

import (
	"bytes"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/record"
	"github.com/LastStep/Bonsai/internal/workspace"
)

const testID = "ws-aaaaaaaaaaaaaaaaaaaaaaaaaa"

// now is the fixture's clock: every age is judged against it.
var now = time.Date(2026, 10, 10, 12, 0, 0, 0, time.UTC)

func daysAgo(n float64) time.Time { return now.Add(-time.Duration(n * 24 * float64(time.Hour))) }

// project makes a linked project (no git: it is its own main checkout) whose bonsai.yaml holds generated, as YAML
// lines under generated: ("" leaves generated: out), and the task folder work/tasks. Its own BONSAI_HOME.
func project(t *testing.T, generated string) workspace.Local {
	t.Helper()
	tmp := t.TempDir()
	t.Setenv("BONSAI_HOME", filepath.Join(tmp, "home"))
	root := filepath.Join(tmp, "proj")
	yaml := "format: bonsai.workspace/1\nid: " + testID + "\nname: demo\ndocuments:\n  task: work/tasks\n"
	if generated != "" {
		yaml += "generated:\n" + generated
	}
	write(t, root, "bonsai.yaml", yaml)
	for _, d := range []string{"log", "asks", "ladder"} {
		if err := os.MkdirAll(filepath.Join(root, ".bonsai", "local", d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	return workspace.Local{Root: root, Main: root}
}

// projectAt makes the folders of a second checkout's .bonsai/local/ at root (no bonsai.yaml): what a link points to.
func projectAt(t *testing.T, root string) workspace.Local {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(root, ".bonsai", "local", "ladder"), 0o755); err != nil {
		t.Fatal(err)
	}
	return workspace.Local{Root: root, Main: root}
}

func write(t *testing.T, root, rel, content string) string {
	t.Helper()
	p := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

// stamp sets a file's modification time.
func stamp(t *testing.T, p string, at time.Time) {
	t.Helper()
	if err := os.Chtimes(p, at, at); err != nil {
		t.Fatal(err)
	}
}

// opts are a run's options in the fixture: its clock, no environment.
func opts(l workspace.Local) Options {
	return Options{Local: l, Now: func() time.Time { return now }, Getenv: func(string) string { return "" }}
}

// rec is one log record to write: its event, when, and what else it fills.
type rec struct {
	event    string
	at       time.Time
	session  string
	subagent string
	target   string
	source   string
}

func logLine(t *testing.T, r rec) []byte {
	t.Helper()
	l := record.New(r.event, record.Common{Workspace: testID, Session: r.session, Agent: agentOf(r.session), At: r.at})
	if r.subagent != "" {
		l.SubagentID = &r.subagent
	}
	if r.target != "" {
		l.Target = &r.target
	}
	if r.source != "" {
		l.Source = &r.source
	}
	b, err := record.Line(l)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func agentOf(session string) string {
	if session == "" {
		return ""
	}
	return "claude-code"
}

// logFile writes a log file of the records, its modification time its last record's.
func logFile(t *testing.T, l workspace.Local, name string, recs ...rec) string {
	t.Helper()
	var b bytes.Buffer
	for _, r := range recs {
		b.Write(logLine(t, r))
	}
	p := write(t, l.Main, ".bonsai/local/log/"+name, b.String())
	if len(recs) > 0 {
		stamp(t, p, recs[len(recs)-1].at)
	}
	return p
}

// session gives the records of one ended session: its start (with task), a tool call, and its end.
func session(id, task string, start, end time.Time) []rec {
	return []rec{
		{event: "session_start", at: start, session: id, target: task, source: "startup"},
		{event: "tool_end", at: start.Add(time.Minute), session: id, target: "src/main.go"},
		{event: "session_end", at: end, session: id},
	}
}

// askLine is one ask record.
func askLine(t *testing.T, op, key string, at time.Time) []byte {
	t.Helper()
	a := &format.Ask{ID: record.NewID(), At: at.UTC().Format(record.AtLayout), Op: op, Key: key, Workspace: testID,
		Source: "agent", Options: []string{}}
	if op == "file" {
		typ, title := "Answer", "A made-up question"
		a.Type, a.Title = &typ, &title
	}
	if op == "answer" {
		words := "yes"
		a.Answer = &format.AskAnswer{By: "terminal", Words: &words}
	}
	b, err := a.Encode()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasSuffix(b, []byte("\n")) {
		b = append(b, '\n')
	}
	return b
}

// asksFile writes an asks day file of the records, its modification time its last record's.
func asksFile(t *testing.T, l workspace.Local, day string, recs ...[]byte) string {
	t.Helper()
	p := write(t, l.Main, ".bonsai/local/asks/"+day+".ndjson", string(bytes.Join(recs, nil)))
	last, _ := time.Parse("2006-01-02", day)
	stamp(t, p, last.Add(12*time.Hour))
	return p
}

// ladderFile writes a task's ladder result, finished at, its task field task (nil for null), its modification time
// finished's.
func ladderFile(t *testing.T, l workspace.Local, name string, task *string, finished time.Time) string {
	t.Helper()
	at := finished.UTC().Format(record.AtLayout)
	r := &format.Ladder{Task: task, Workspace: testID, Mode: "local", Started: at, Finished: at,
		Git: format.LadderGit{SHA: strings.Repeat("ab", 20), Branch: "main"}, Requested: []int64{0},
		Green: true, Rungs: []format.LadderRung{}, Skipped: []format.LadderSkip{}}
	b, err := r.Encode()
	if err != nil {
		t.Fatal(err)
	}
	p := write(t, l.Main, ".bonsai/local/ladder/"+name, string(b))
	stamp(t, p, finished)
	return p
}

// taskFile writes a task file with its status ("broken" writes one that does not read).
func taskFile(t *testing.T, l workspace.Local, id, status string) {
	t.Helper()
	body := "---\nformat: bonsai.task/1\nid: " + id + "\ntitle: A made-up task\nstatus: " + status +
		"\nlane: light\ndone_when:\n  - \"It is made up\"\ndepends_on: []\nblocked_by:\ncreated: 2026-09-01\n" +
		"started:\nfinished:\nlabels: {}\n---\n\n# A made-up task\n"
	if status == "broken" {
		body = "---\nformat: bonsai.task/1\nid: " + id + "\nstatus: [not, a, status\n---\n"
	}
	write(t, l.Main, "work/tasks/"+id+"-made-up.md", body)
}

func ptr(s string) *string { return &s }

// sessionsTable writes .bonsai/sessions.md with the rows.
func sessionsTable(t *testing.T, l workspace.Local, rows ...format.SessionRow) {
	t.Helper()
	tb := &format.Sessions{Sessions: append([]format.SessionRow{}, rows...), Hours: []format.HoursRow{}}
	b, err := tb.Encode()
	if err != nil {
		t.Fatal(err)
	}
	write(t, l.Main, ".bonsai/sessions.md", string(b))
}

// row is a session's row as the table holds it.
func row(session, task string, start, end time.Time) format.SessionRow {
	return format.SessionRow{Session: session[:8], Kind: "session", Task: task, Start: start.UTC().Format(rowLayout),
		End: end.UTC().Format(rowLayout), Minutes: int64(end.Sub(start) / time.Minute)}
}

// names lists a folder's entries under .bonsai/local/, sorted ("x/" for a folder).
func names(t *testing.T, l workspace.Local, folder string) []string {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(l.Main, ".bonsai", "local", folder))
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, e := range entries {
		n := e.Name()
		if e.IsDir() {
			n += "/"
		}
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

// cleanRecords reads today's day file's clean records, as target=reason.
func cleanRecords(t *testing.T, l workspace.Local) []string {
	t.Helper()
	lf, err := record.ReadLog(filepath.Join(l.Main, ".bonsai", "local", "log", "w-"+now.Format("2006-01-02")+".ndjson"))
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, r := range lf.Records {
		if r.Event != "clean" {
			continue
		}
		if r.Session != nil || r.Target == nil || r.Reason == nil {
			t.Errorf("a clean record with session %v, target %v, reason %v", r.Session, r.Target, r.Reason)
			continue
		}
		out = append(out, *r.Target+"="+*r.Reason)
	}
	sort.Strings(out)
	return out
}

func same(a, b []string) bool {
	return strings.Join(a, "\n") == strings.Join(b, "\n")
}
