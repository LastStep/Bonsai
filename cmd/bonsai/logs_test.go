package main

// bonsai logs (step 5.2.3): the listing, --session by a prefix, --day, the refusals, and --json against bonsai.logs/1.
// Every log is a fixture written fresh by the test.

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/record"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/testpack"
)

const (
	logWS = "ws-abcdefghijklmnopqrstuvwxyz"
	idA   = "6d1e2f3a-4b5c-4d6e-8f7a-9b0c1d2e3f4a"
	idC   = "6d1e2f3a-9999-4d6e-8f7a-000000000000" // shares A's first 8 characters
	idB   = "7a2b3c4d-5e6f-4a7b-8c9d-0e1f2a3b4c5d"
	idD   = "7a2b3c4d" // a whole id that is also B's prefix
)

// fixture is one record of a fixture log: a clock time on 2026-10-08, its event and what it sets.
type fixture struct {
	at, event string
	set       func(*format.Log)
}

func fstr(s string) *string { return &s }

// writeLogFile writes name in the project's log folder from the fixtures, and adds torn to its end.
func (c *cli) writeLogFile(name, session, torn string, events ...fixture) {
	c.t.Helper()
	var b []byte
	for _, e := range events {
		ts, err := time.Parse(record.AtLayout, "2026-10-08T"+e.at+":00.000Z")
		if err != nil {
			c.t.Fatal(err)
		}
		l := record.New(e.event, record.Common{Workspace: logWS, Session: session, Agent: "claude-code", At: ts})
		if e.set != nil {
			e.set(l)
		}
		line, err := record.Line(l)
		if err != nil {
			c.t.Fatal(err)
		}
		b = append(b, line...)
	}
	c.write(".bonsai/local/log/"+name, string(b)+torn)
}

func sessionStart(at, task, role string) fixture {
	return fixture{at, "session_start", func(l *format.Log) { l.Target, l.Role = fstr(task), fstr(role) }}
}

func logsOf(t *testing.T, out string) (schema.Object, []schema.Object) {
	t.Helper()
	doc := fits(t, out, "logs")
	files, _ := doc.Get("files")
	var list []schema.Object
	if l, ok := files.([]any); ok {
		for _, f := range l {
			list = append(list, f.(schema.Object))
		}
	}
	return doc, list
}

func fileNames(files []schema.Object) []string {
	var out []string
	for _, f := range files {
		out = append(out, f.String("file"))
	}
	return out
}

func TestLogs(t *testing.T) {
	c := newCLI(t)
	// The fixtures are of 2026-10-08, and it is 13:00 that day: a session with no end line since 10:00 is open.
	defer func(f func() time.Time) { logsNow = f }(logsNow)
	logsNow = func() time.Time { return time.Date(2026, 10, 8, 13, 0, 0, 0, time.UTC) }
	// Not linked: exit 4, the error alone in the document, the workspace null.
	code, out, _ := c.run("", "logs", "--json")
	doc := fits(t, out, "logs")
	if w, _ := doc.Get("workspace"); code != 4 || w != nil || errorIn(doc, "logs").String("code") != "not-linked" {
		t.Errorf("not linked: %d\n%s", code, out)
	}
	c.link()
	c.commit("link")

	// A log with no files.
	code, out, _ = c.run("", "logs")
	if code != 0 || out != "no log files in .bonsai/local/log/\n" {
		t.Errorf("no files: %d %q", code, out)
	}
	code, out, _ = c.run("", "logs", "--json")
	if _, files := logsOf(t, out); code != 0 || files == nil && !strings.Contains(out, `"files": []`) {
		t.Errorf("no files, json: %d\n%s", code, out)
	}

	c.writeLogFile("s-"+idA+".ndjson", idA, "", sessionStart("09:05", "T-0901", "builder"), fixture{"09:30", "tool_end", nil}, fixture{"10:35", "session_end", nil})
	c.writeLogFile("s-"+idC+".ndjson", idC, `{"format":"bonsai.log/1","id":"torn`, sessionStart("10:00", "none", "verifier"))
	c.writeLogFile("s-"+idB+".ndjson", idB, "", sessionStart("11:00", "T-0902", "orchestrator"), fixture{"11:20", "tool_end", nil})
	c.writeLogFile("s-"+idD+".ndjson", idD, "", sessionStart("12:00", "none", "x"), fixture{"12:01", "session_end", nil})
	c.writeLogFile("w-2026-10-08.ndjson", "", "", fixture{"16:00", "event", nil})
	c.write(".bonsai/local/log/notes.txt", "not a log file\n")
	before := c.files()

	// The listing: newest first in the text, one line per file; oldest first in --json, which fits the schema.
	code, out, _ = c.run("", "logs")
	lines := strings.Split(strings.TrimSuffix(out, "\n"), "\n")
	if code != 0 || len(lines) != 5 {
		t.Fatalf("the listing: %d\n%s", code, out)
	}
	if !strings.HasPrefix(lines[0], "session "+idD) || !strings.HasPrefix(lines[1], "session "+idB+"  ") || !strings.HasPrefix(lines[2], "session "+idC) ||
		!strings.HasPrefix(lines[3], "session "+idA) || !strings.HasPrefix(lines[4], "day 2026-10-08") {
		t.Errorf("the listing is not newest first:\n%s", out)
	}
	if !strings.Contains(lines[1], "open  task T-0902  role orchestrator  2 records, 0 unreadable") ||
		!strings.Contains(lines[2], "open  task none  role verifier  1 records, 1 unreadable") ||
		!strings.Contains(lines[3], "ended  task T-0901  role builder  3 records, 0 unreadable") ||
		!strings.Contains(lines[4], "2026-10-08T16:00:00.000Z to 2026-10-08T16:00:00.000Z  1 records, 0 unreadable") {
		t.Errorf("the listing's lines:\n%s", out)
	}
	code, out, _ = c.run("", "logs", "--json")
	doc, files := logsOf(t, out)
	if code != 0 || !reflect.DeepEqual(fileNames(files), []string{"w-2026-10-08.ndjson", "s-" + idA + ".ndjson", "s-" + idC + ".ndjson", "s-" + idB + ".ndjson", "s-" + idD + ".ndjson"}) {
		t.Fatalf("the files: %d %v\n%s", code, fileNames(files), out)
	}
	if r, _ := doc.Get("records"); r != nil {
		t.Errorf("records of a listing: %v", r)
	}
	if ended, _ := files[1].Get("ended"); ended != true {
		t.Errorf("A ended: %v", ended)
	}
	if ended, _ := files[3].Get("ended"); ended != false {
		t.Errorf("B is open: %v", ended)
	}
	if day, _ := files[0].Get("day"); day != "2026-10-08" {
		t.Errorf("day file: %v", files[0])
	}
	if un, _ := files[2].Get("unreadable"); un.(interface{ String() string }).String() != "1" {
		t.Errorf("unreadable: %v", un)
	}

	// --session: a prefix of 8 characters, a longer prefix, a whole id that is also a prefix.
	code, out, _ = c.run("", "logs", "--session", "7a2b3c4d-5")
	if code != 0 || strings.Count(out, "\n") != 2 || !strings.HasPrefix(out, `{"format":"bonsai.log/1","id":"`) || !strings.Contains(out, `"event":"session_start"`) {
		t.Errorf("--session B: %d\n%s", code, out)
	}
	code, out, _ = c.run("", "logs", "--session", "7a2b3c4d", "--json")
	doc, files = logsOf(t, out)
	recs, _ := doc.Get("records")
	if code != 0 || len(files) != 1 || files[0].String("file") != "s-"+idD+".ndjson" || len(recs.([]any)) != 2 {
		t.Errorf("--session with the whole id: %d\n%s", code, out)
	}
	code, out, _ = c.run("", "logs", "--session", "6d1e2f3a-4", "--json")
	doc, files = logsOf(t, out)
	recs, _ = doc.Get("records")
	if code != 0 || len(files) != 1 || files[0].String("file") != "s-"+idA+".ndjson" || len(recs.([]any)) != 3 {
		t.Errorf("--session A: %d\n%s", code, out)
	}
	if first := recs.([]any)[0].(schema.Object); first.String("event") != "session_start" || first.String("target") != "T-0901" || first.String("format") != "bonsai.log/1" {
		t.Errorf("a record as stored: %v", first)
	}
	// An ambiguous prefix and one that matches none: exit 4, session-not-found, naming the matches.
	code, out, errOut := c.run("", "logs", "--session", "6d1e2f3a")
	if code != 4 || out != "" || !strings.Contains(errOut, idA) || !strings.Contains(errOut, idC) || !strings.Contains(errOut, "next: ") {
		t.Errorf("an ambiguous prefix: %d %q %q", code, out, errOut)
	}
	code, out, _ = c.run("", "logs", "--session", "6d1e2f3a", "--json")
	doc = fits(t, out, "logs")
	e := errorIn(doc, "logs")
	if w, _ := doc.Get("workspace"); code != 4 || e.String("code") != "session-not-found" || !strings.Contains(e.String("message"), idA) || !strings.Contains(e.String("message"), idC) || w != nil {
		t.Errorf("an ambiguous prefix, json: %d\n%s", code, out)
	}
	code, out, _ = c.run("", "logs", "--session", "ffffffff", "--json")
	if e = errorIn(fits(t, out, "logs"), "logs"); code != 4 || e.String("code") != "session-not-found" {
		t.Errorf("no match: %d\n%s", code, out)
	}
	// A prefix under 8 characters, both filters, a bad date, a flag without its value, an unknown flag.
	for _, args := range [][]string{{"--session", "6d1e"}, {"--session", idA, "--day", "2026-10-08"}, {"--day", "2026-13-40"}, {"--day"}, {"--since", "x"}, {"extra"}} {
		code, out, errOut = c.run("", append([]string{"logs"}, args...)...)
		if code != 2 || out != "" || !strings.Contains(errOut, "next: ") {
			t.Errorf("logs %v: %d %q %q", args, code, out, errOut)
		}
	}
	code, out, _ = c.run("", "logs", "--session", "6d1e", "--json")
	if e = errorIn(fits(t, out, "logs"), "logs"); code != 2 || e.String("code") != "bad-value" {
		t.Errorf("a short prefix, json: %d\n%s", code, out)
	}

	// --day: that UTC day's file; a day with no file prints none.
	code, out, _ = c.run("", "logs", "--day", "2026-10-08", "--json")
	doc, files = logsOf(t, out)
	recs, _ = doc.Get("records")
	if code != 0 || len(files) != 1 || files[0].String("file") != "w-2026-10-08.ndjson" || len(recs.([]any)) != 1 || recs.([]any)[0].(schema.Object).String("event") != "event" {
		t.Errorf("--day: %d\n%s", code, out)
	}
	code, out, _ = c.run("", "logs", "--day", "2026-10-08")
	if code != 0 || strings.Count(out, "\n") != 1 || !strings.Contains(out, `"event":"event"`) {
		t.Errorf("--day, text: %d\n%s", code, out)
	}
	code, out, _ = c.run("", "logs", "--day", "2026-10-07", "--json")
	doc, files = logsOf(t, out)
	if recs, _ = doc.Get("records"); code != 0 || len(files) != 0 || len(recs.([]any)) != 0 {
		t.Errorf("--day with no file: %d\n%s", code, out)
	}
	if code, out, _ = c.run("", "logs", "--day", "2026-10-07"); code != 0 || out != "no log file for 2026-10-07\n" {
		t.Errorf("--day with no file, text: %d %q", code, out)
	}

	// It wrote nothing, and it is stable.
	if after := c.files(); !reflect.DeepEqual(before, after) {
		t.Errorf("logs changed the files:\n%v\n%v", before, after)
	}
	_, again, _ := c.run("", "logs", "--json")
	_, first, _ := c.run("", "logs", "--json")
	if again != first {
		t.Errorf("two runs differ")
	}

	// A day later the 24-hour rule has ended them.
	logsNow = func() time.Time { return time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC) }
	if _, out, _ = c.run("", "logs"); strings.Contains(out, "  open") {
		t.Errorf("a span still open after 24 hours:\n%s", out)
	}
	logsNow = func() time.Time { return time.Date(2026, 10, 8, 13, 0, 0, 0, time.UTC) }

	// A worktree reads the main checkout's log.
	wt := filepath.Join(c.tmp, "task-branch")
	testpack.Git(t, c.root, "worktree", "add", "-q", "-b", "task-branch", wt)
	t.Chdir(wt)
	code, out, _ = c.run("", "logs", "--json")
	doc, files = logsOf(t, out)
	ws, _ := doc.Get("workspace")
	if code != 0 || len(files) != 5 || ws.(schema.Object).String("root") != filepath.ToSlash(c.root) {
		t.Errorf("from a worktree: %d %v\n%s", code, ws, out)
	}
	if _, err := os.Stat(filepath.Join(wt, ".bonsai", "local", "log")); err == nil {
		t.Errorf("the worktree got a log folder of its own")
	}
}
