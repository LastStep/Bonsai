package sessions

// Fixture logs written fresh by each test (no Claude Code): one test per span rule of plan-5 5.2.3 note 1, then the
// rows, their key and order, the hours, and the table's bytes.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/record"
)

const (
	wsID = "ws-abcdefghijklmnopqrstuvwxyz"
	sid  = "6d1e2f3a-4b5c-4d6e-8f7a-9b0c1d2e3f4a"
	sid2 = "7a2b3c4d-5e6f-4a7b-8c9d-0e1f2a3b4c5d"
)

// ev is one record of a fixture: its time (a clock text, today's date), its event and what it sets.
type ev struct {
	at, event string
	set       func(*format.Log)
}

func str(s string) *string { return &s }

func start(at, task, role, model string) ev {
	return ev{at, "session_start", func(l *format.Log) { l.Target, l.Role, l.Model = str(task), str(role), str(model) }}
}
func compact(at string) ev {
	return ev{at, "session_start", func(l *format.Log) { l.Source = str("compact") }}
}
func end(at string) ev  { return ev{at, "session_end", nil} }
func tool(at string) ev { return ev{at, "tool_end", nil} }
func subStart(at, id, typ, task string) ev {
	return ev{at, "subagent_start", func(l *format.Log) { l.SubagentID, l.SubagentType, l.Target = str(id), str(typ), str(task) }}
}
func subStop(at, id string) ev {
	return ev{at, "subagent_stop", func(l *format.Log) { l.SubagentID = str(id) }}
}

// day is the fixtures' date; a clock text "09:05" is 09:05 on it. The tests' "now" is the day after, 12:00.
const day = "2026-10-08"

var now = time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC)

// writeLog writes the session's file in dir from the events, and gives its path.
func writeLog(t *testing.T, dir, session string, events ...ev) string {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	var b []byte
	for _, e := range events {
		at := e.at
		if len(at) == len("09:05") {
			at = day + "T" + at + ":00.000Z"
		} else if len(at) == len("09:05:30") {
			at = day + "T" + at + ".000Z"
		}
		ts, err := time.Parse(record.AtLayout, at)
		if err != nil {
			t.Fatal(err)
		}
		l := record.New(e.event, record.Common{Workspace: wsID, Session: session, Agent: "claude-code", At: ts})
		if e.set != nil {
			e.set(l)
		}
		line, err := record.Line(l)
		if err != nil {
			t.Fatal(err)
		}
		b = append(b, line...)
	}
	path := filepath.Join(dir, "s-"+session+".ndjson")
	if err := os.WriteFile(path, b, 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func read(t *testing.T, path string, at time.Time) *File {
	t.Helper()
	f, err := Read(path, at)
	if err != nil {
		t.Fatal(err)
	}
	return f
}

// shown is a span as a test compares it: start-end and whether it is open.
func shown(s Span) string {
	if s.Open {
		return stamp(s.Start)[11:] + " open"
	}
	return stamp(s.Start)[11:] + "-" + stamp(s.End)[11:]
}

func shownAll(f *File) string {
	var out []string
	for _, s := range f.Spans {
		kind := "session"
		if s.Subagent != "" {
			kind = "sub " + s.Subagent
		}
		out = append(out, kind+" "+shown(s))
	}
	return strings.Join(out, "; ")
}

func TestSpanEndedBySessionEnd(t *testing.T) {
	dir := t.TempDir()
	p := writeLog(t, dir, sid, start("09:05", "T-0901", "builder", "claude-sonnet"), tool("09:30"), end("10:35"))
	f := read(t, p, now)
	if got := shownAll(f); got != "session 09:05-10:35" {
		t.Errorf("spans: %s", got)
	}
	s := f.Spans[0]
	if s.Session != sid || s.Task != "T-0901" || *s.Role != "builder" || *s.Model != "claude-sonnet" || s.Open {
		t.Errorf("span: %+v", s)
	}
	if f.Ended == nil || !*f.Ended || *f.Task != "T-0901" || f.Records != 3 || f.First != day+"T09:05:00.000Z" || f.Last != day+"T10:35:00.000Z" {
		t.Errorf("file: %+v", f)
	}
	if open, err := Open(p, now); err != nil || open {
		t.Errorf("Open of an ended file: %v %v", open, err)
	}
}

func TestACompactionIsTheSameSession(t *testing.T) {
	p := writeLog(t, t.TempDir(), sid, start("09:00", "T-0901", "builder", "m"), compact("09:40"), tool("09:50"), compact("10:10"), end("10:30"))
	f := read(t, p, now)
	if got := shownAll(f); got != "session 09:00-10:30" {
		t.Errorf("a compaction opened a span: %s", got)
	}
	// A compaction with no span open (a file that begins with one) opens the span.
	p = writeLog(t, t.TempDir(), sid, compact("09:00"), end("09:30"))
	if got := shownAll(read(t, p, now)); got != "session 09:00-09:30" {
		t.Errorf("a first compaction: %s", got)
	}
}

func TestALostEndLineClosesAtTheLastRecord(t *testing.T) {
	p := writeLog(t, t.TempDir(), sid, start("09:00", "T-0901", "builder", "m"), tool("09:20"), tool("09:44"),
		start("11:00", "T-0902", "builder", "m"), end("11:30"))
	f := read(t, p, now)
	if got := shownAll(f); got != "session 09:00-09:44; session 11:00-11:30" {
		t.Errorf("a lost end: %s", got)
	}
	if f.Spans[0].Task != "T-0901" || f.Spans[1].Task != "T-0902" || f.Task == nil || *f.Task != "T-0901" {
		t.Errorf("tasks: %+v", f.Spans)
	}
}

func TestASubagentRunWithAndWithoutItsStop(t *testing.T) {
	p := writeLog(t, t.TempDir(), sid,
		start("09:00", "T-0901", "orchestrator", "m"),
		subStart("09:10", "agent-aaaaaaaa-1", "builder", "T-0901"), subStart("09:12", "agent-bbbbbbbb-2", "verifier", "T-0901"),
		subStop("09:40", "agent-aaaaaaaa-1"), tool("09:50"), end("10:00"))
	f := read(t, p, now)
	if got := shownAll(f); got != "session 09:00-10:00; sub agent-aaaaaaaa-1 09:10-09:40; sub agent-bbbbbbbb-2 09:12-10:00" {
		t.Errorf("subagent runs: %s", got)
	}
	if *f.Spans[1].Role != "builder" || *f.Spans[2].Role != "verifier" {
		t.Errorf("roles come from the subagent type: %+v", f.Spans)
	}
	// A stop for a run never started is ignored; a run's stop closes only that run.
	p = writeLog(t, t.TempDir(), sid, start("09:00", "none", "x", "m"), subStop("09:05", "agent-unknown"), subStart("09:10", "agent-cccccccc-3", "builder", "T-1"), end("09:30"))
	if got := shownAll(read(t, p, now)); got != "session 09:00-09:30; sub agent-cccccccc-3 09:10-09:30" {
		t.Errorf("an unknown stop: %s", got)
	}
	// A lost end line closes a run still open at the session's last record.
	p = writeLog(t, t.TempDir(), sid, start("09:00", "T-1", "x", "m"), subStart("09:10", "agent-dddddddd-4", "builder", "T-1"), tool("09:30"), start("10:00", "T-1", "x", "m"), end("10:10"))
	if got := shownAll(read(t, p, now)); got != "session 09:00-09:30; sub agent-dddddddd-4 09:10-09:30; session 10:00-10:10" {
		t.Errorf("a run with a lost end: %s", got)
	}
}

func TestTheTwentyFourHourRule(t *testing.T) {
	events := []ev{start("09:00", "T-0901", "builder", "m"), subStart("09:10", "agent-eeeeeeee-5", "builder", "T-0901"), tool("09:44")}
	p := writeLog(t, t.TempDir(), sid, events...)
	// 24 hours or less after the last record: open (the span and its run), no end yet.
	for _, at := range []time.Time{time.Date(2026, 10, 8, 10, 0, 0, 0, time.UTC), time.Date(2026, 10, 9, 9, 44, 0, 0, time.UTC)} {
		f := read(t, p, at)
		if got := shownAll(f); got != "session 09:00 open; sub agent-eeeeeeee-5 09:10 open" || f.Ended == nil || *f.Ended {
			t.Errorf("at %v: %s", at, got)
		}
		if open, err := Open(p, at); err != nil || !open {
			t.Errorf("at %v: Open %v %v", at, open, err)
		}
	}
	// More than 24 hours: the span ends at its last record (a killed or crashed session), the run with it.
	late := time.Date(2026, 10, 9, 9, 44, 1, 0, time.UTC)
	f := read(t, p, late)
	if got := shownAll(f); got != "session 09:00-09:44; sub agent-eeeeeeee-5 09:10-09:44" || !*f.Ended {
		t.Errorf("after 24 hours: %s", got)
	}
	if open, _ := Open(p, late); open {
		t.Errorf("a stale span is still open")
	}
}

func TestAnOpenSpanHasNoRow(t *testing.T) {
	dir := t.TempDir()
	writeLog(t, dir, sid, start("09:00", "T-0901", "builder", "m"), end("09:30"), start("11:00", "T-0901", "builder", "m"), tool("11:20"))
	rows, err := Found(dir, time.Date(2026, 10, 8, 12, 0, 0, 0, time.UTC))
	if err != nil || len(rows) != 1 || rows[0].End != day+" 09:30" {
		t.Fatalf("rows: %v %+v", err, rows)
	}
	// The same file a day later: the open span has ended at its last record, and has its row.
	rows, _ = Found(dir, now)
	if len(rows) != 2 || rows[1].Start != day+" 11:00" || rows[1].End != day+" 11:20" || rows[1].Minutes != 20 {
		t.Errorf("rows a day later: %+v", rows)
	}
}

func TestRows(t *testing.T) {
	dir := t.TempDir()
	writeLog(t, dir, sid, start("09:05:50", "T-0901", "orchestrator", "claude-opus"), subStart("09:10", "agent-aaaaaaaa-1", "builder", "T-0902"),
		subStop("10:30:59", "agent-aaaaaaaa-1"), end("10:35:10"))
	writeLog(t, dir, sid2, start("08:00", "", "", ""), end("08:20"))
	rows, err := Found(dir, now)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 3 {
		t.Fatalf("rows: %+v", rows)
	}
	// Sorted by start: the other session first. Minutes are the shown end minus the shown start.
	a, b, c := rows[0], rows[1], rows[2]
	if a.Session != "7a2b3c4d" || a.Kind != "session" || a.Task != "none" || a.Start != day+" 08:00" || a.Minutes != 20 || a.Subagent != nil {
		t.Errorf("first row: %+v", a)
	}
	if b.Session != "6d1e2f3a" || b.Kind != "session" || b.Task != "T-0901" || *b.Role != "orchestrator" || *b.Model != "claude-opus" ||
		b.Start != day+" 09:05" || b.End != day+" 10:35" || b.Minutes != 90 || b.Subagent != nil {
		t.Errorf("session row: %+v", b)
	}
	if c.Session != "6d1e2f3a" || c.Kind != "subagent" || c.Task != "T-0902" || *c.Role != "builder" || c.Model != nil ||
		c.Start != day+" 09:10" || c.End != day+" 10:30" || c.Minutes != 80 || *c.Subagent != "agent-aa" {
		t.Errorf("subagent row: %+v", c)
	}
	if a.Role != nil || a.Model != nil {
		t.Errorf("a session with no role or model has null cells: %+v", a)
	}
}

func row(session, sub, task, role, start, end string, minutes int64) format.SessionRow {
	r := format.SessionRow{Session: session, Kind: "session", Task: task, Start: day + " " + start, End: day + " " + end, Minutes: minutes}
	if role != "" {
		r.Role = str(role)
	}
	if sub != "" {
		r.Kind, r.Subagent = "subagent", str(sub)
	}
	return r
}

func TestMergeKeepsKeysAndRowsOfGoneFiles(t *testing.T) {
	// The table holds a row whose log file is gone, and a row the log now says otherwise about (same key).
	table := &format.Sessions{Sessions: []format.SessionRow{
		row("gonegone", "", "T-0001", "builder", "07:00", "08:00", 60),
		row("6d1e2f3a", "", "T-0901", "builder", "09:00", "09:10", 10),
	}}
	found := []format.SessionRow{
		row("6d1e2f3a", "", "T-0901", "builder", "09:00", "10:00", 60), // the same key: never rewritten
		row("6d1e2f3a", "agent-aa", "T-0901", "builder", "09:00", "09:50", 50),
		row("4d4d4d4d", "", "none", "", "08:30", "08:40", 10),
	}
	if m := Missing(table.Sessions, found); len(m) != 2 {
		t.Fatalf("missing: %+v", m)
	}
	got := Merge(table, found)
	var keys []string
	for _, r := range got.Sessions {
		keys = append(keys, r.Start[11:]+" "+r.Session+" "+subagentOf(r)+" "+r.End[11:])
	}
	want := []string{"07:00 gonegone  08:00", "08:30 4d4d4d4d  08:40", "09:00 6d1e2f3a  09:10", "09:00 6d1e2f3a agent-aa 09:50"}
	if strings.Join(keys, "|") != strings.Join(want, "|") {
		t.Errorf("rows:\n%s\nwant\n%s", strings.Join(keys, "\n"), strings.Join(want, "\n"))
	}
	// A second merge adds nothing and changes nothing.
	again := Merge(got, found)
	a, _ := got.Encode()
	b, _ := again.Encode()
	if string(a) != string(b) {
		t.Errorf("a second merge changed the table:\n%s\n%s", a, b)
	}
}

func TestHoursKeepSessionsAndSubagentsApart(t *testing.T) {
	rows := []format.SessionRow{
		row("aaaaaaaa", "", "T-0901", "orchestrator", "09:00", "11:00", 120),
		row("aaaaaaaa", "agent-aa", "T-0901", "builder", "09:10", "10:10", 60),
		row("aaaaaaaa", "agent-bb", "T-0901", "builder", "10:10", "10:40", 30),
		row("bbbbbbbb", "", "T-0901", "orchestrator", "13:00", "13:10", 10),
		row("cccccccc", "", "none", "", "14:00", "14:04", 4),
		row("cccccccc", "agent-cc", "T-0901", "orchestrator", "14:00", "14:05", 5),
	}
	var got []string
	for _, h := range Hours(rows) {
		role := "-"
		if h.Role != nil {
			role = *h.Role
		}
		got = append(got, h.Task+" "+role+" "+h.Kind+" "+string(h.Hours))
	}
	want := "T-0901 builder subagent 1.5|T-0901 orchestrator session 2.2|T-0901 orchestrator subagent 0.1|none - session 0.1"
	if strings.Join(got, "|") != want {
		t.Errorf("hours:\n%s\nwant\n%s", strings.Join(got, "|"), want)
	}
	for minutes, want := range map[int64]string{0: "0.0", 2: "0.0", 3: "0.1", 90: "1.5", 125: "2.1", 126: "2.1", 128: "2.1", 129: "2.2", 600: "10.0", 6: "0.1"} {
		if got := string(tenths(minutes)); got != want {
			t.Errorf("tenths(%d) = %s, want %s", minutes, got, want)
		}
	}
}

func TestTheTableIsByteStable(t *testing.T) {
	dir := t.TempDir()
	writeLog(t, dir, sid, start("09:05", "T-0901", "orchestrator", "claude-opus"), subStart("09:10", "agent-aaaaaaaa-1", "builder", "T-0901"),
		subStop("10:00", "agent-aaaaaaaa-1"), end("10:35"))
	writeLog(t, dir, sid2, start("08:00", "", "", ""), end("08:20"))
	var first string
	for i := 0; i < 2; i++ {
		found, err := Found(dir, now)
		if err != nil {
			t.Fatal(err)
		}
		b, err := Merge(&format.Sessions{}, found).Encode()
		if err != nil {
			t.Fatal(err)
		}
		if i == 0 {
			first = string(b)
		} else if string(b) != first {
			t.Errorf("the second run differs:\n%s\n%s", first, b)
		}
		if strings.Contains(string(b), "\r") {
			t.Errorf("a CR in the table")
		}
		if _, err := format.ReadSessions(b); err != nil {
			t.Errorf("the table does not read back: %v", err)
		}
	}
	want := "---\nformat: bonsai.sessions/1   # generated by bonsai check --write from the log; never edit by hand\n---\n" +
		"| Session | Kind | Task | Role | Model | Start | End | Minutes | Subagent |\n|---|---|---|---|---|---|---|---|---|\n" +
		"| 7a2b3c4d | session | none | | | 2026-10-08 08:00 | 2026-10-08 08:20 | 20 | |\n" +
		"| 6d1e2f3a | session | T-0901 | orchestrator | claude-opus | 2026-10-08 09:05 | 2026-10-08 10:35 | 90 | |\n" +
		"| 6d1e2f3a | subagent | T-0901 | builder | | 2026-10-08 09:10 | 2026-10-08 10:00 | 50 | agent-aa |\n" +
		"\nHours per task and role (a subagent run lies inside its session's row, so the two are never added):\n\n" +
		"| Task | Role | Kind | Hours |\n|---|---|---|---|\n" +
		"| T-0901 | builder | subagent | 0.8 |\n| T-0901 | orchestrator | session | 1.5 |\n| none | | session | 0.3 |\n"
	if first != want {
		t.Errorf("the table:\n%s\nwant\n%s", first, want)
	}
}

func TestFoundIgnoresOtherFilesAndAMissingFolder(t *testing.T) {
	if rows, err := Found(filepath.Join(t.TempDir(), "nope"), now); err != nil || len(rows) != 0 {
		t.Errorf("a missing folder: %v %+v", err, rows)
	}
	dir := t.TempDir()
	writeLog(t, dir, sid, start("09:00", "none", "x", "m"), end("09:10"))
	if err := os.WriteFile(filepath.Join(dir, "notes.ndjson"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if rows, err := Found(dir, now); err != nil || len(rows) != 1 {
		t.Errorf("rows: %v %+v", err, rows)
	}
}

func TestATornLineIsCountedNotFatal(t *testing.T) {
	dir := t.TempDir()
	p := writeLog(t, dir, sid, start("09:00", "T-1", "x", "m"), end("09:10"))
	f, err := os.OpenFile(p, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = f.WriteString(`{"format":"bonsai.log/1","id":"torn`)
	_ = f.Close()
	got := read(t, p, now)
	if got.Unreadable != 1 || got.Records != 2 || len(got.Spans) != 1 {
		t.Errorf("a torn line: %+v", got)
	}
}
