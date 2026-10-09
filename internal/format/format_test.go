package format

// Every format's Go type against its schema and its example (formats/examples): read, and where Bonsai writes the
// format, written back byte for byte; the reading rules every format shares (contract §2.2), each refusal with its
// next step; the registry against the set. The lock's Go type is internal/workspace's, and its example is read and
// written back there (TestLockReadsAndWritesTheSetsExample).

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/formats"
	"github.com/LastStep/Bonsai/internal/schema"
)

// types gives each format's Go type, but the lock's (internal/workspace).
var goTypes = map[string]any{
	"task": Task{}, "labels": Labels{}, "lanes": Lanes{}, "run": Run{}, "state": State{}, "log": Log{}, "ask": Ask{},
	"ladder": Ladder{}, "status": Status{}, "workspace": Workspace{}, "pack": Pack{}, "tasks": Tasks{},
	"sessions": Sessions{}, "memory": Memory{}, "error": ErrorObject{}, "check": Check{}, "changes": Changes{},
}

// readers read each format's example source into its Go type.
var readers = map[string]func([]byte) (any, error){
	"task":      func(b []byte) (any, error) { return ReadTask(b) },
	"labels":    func(b []byte) (any, error) { return ReadLabels(b) },
	"lanes":     func(b []byte) (any, error) { return ReadLanes(b) },
	"run":       func(b []byte) (any, error) { return ReadRun(b) },
	"state":     func(b []byte) (any, error) { return ReadState(b) },
	"log":       func(b []byte) (any, error) { return ReadLog(b) },
	"ask":       func(b []byte) (any, error) { return ReadAsk(b) },
	"ladder":    func(b []byte) (any, error) { return ReadLadder(b) },
	"status":    func(b []byte) (any, error) { return ReadStatus(b) },
	"workspace": func(b []byte) (any, error) { return ReadWorkspace(b) },
	"pack":      func(b []byte) (any, error) { return ReadPack(b) },
	"tasks":     func(b []byte) (any, error) { return ReadTasks(b) },
	"sessions":  func(b []byte) (any, error) { return ReadSessions(b) },
	"memory":    func(b []byte) (any, error) { return ReadMemory(b) },
	"error":     func(b []byte) (any, error) { return ReadErrorObject(b) },
	"check":     func(b []byte) (any, error) { return ReadCheck(b) },
	"changes":   func(b []byte) (any, error) { return ReadChanges(b) },
}

func example(t *testing.T, file string) []byte {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "formats", "examples", file))
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

// source is the file an example's reader reads: the YAML or markdown source when there is one, else the JSON.
func source(f *Format) string {
	switch f.Shape {
	case YAMLFile:
		return f.Name + ".yaml"
	case Markdown, Table:
		return f.Name + ".md"
	}
	return f.Name + ".json"
}

func TestRegistryIsTheSet(t *testing.T) {
	var names []string
	for _, f := range All {
		names = append(names, f.Name)
		s := f.Schema()
		if s.String("title") != f.Versioned() {
			t.Errorf("%s: the schema's title is %q, the registry's %q", f.Name, s.String("title"), f.Versioned())
		}
		if _, ok := goTypes[f.Name]; !ok && f.Name != "lock" {
			t.Errorf("%s: no Go type in this test's table", f.Name)
		}
		for _, l := range f.Lists {
			if l.Words == nil || !strings.HasPrefix(l.Table, "format.") {
				t.Errorf("%s: the list at %s names no table here", f.Name, l.Field)
			}
		}
		if g, ok := Lookup(f.ID()); !ok || g != f {
			t.Errorf("Lookup(%q) does not find it", f.ID())
		}
		if g, ok := Lookup(f.Versioned()); !ok || g != f {
			t.Errorf("Lookup(%q) does not find it", f.Versioned())
		}
	}
	if strings.Join(names, ",") != strings.Join(formats.Names, ",") {
		t.Errorf("the registry lists %v, the set %v", names, formats.Names)
	}
	entries, err := os.ReadDir(filepath.Join("..", "..", "formats", "schemas"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != len(All) {
		t.Errorf("formats/schemas holds %d files, the registry %d formats", len(entries), len(All))
	}
	for _, bad := range []string{"bonsai.task/2", "bonsai.tasks/1x", "nope", "bonsai.", ""} {
		if _, ok := Lookup(bad); ok {
			t.Errorf("Lookup(%q) found a format", bad)
		}
	}
}

func TestTypesFitTheirSchemas(t *testing.T) {
	for _, f := range All {
		v, ok := goTypes[f.Name]
		if !ok {
			continue
		}
		for _, msg := range fit(f.Schema(), reflect.TypeOf(v), f.Name) {
			t.Error(msg)
		}
	}
}

// TestEveryExampleReads reads each format's example into its Go type, and checks the type gives back the document
// the set says a reader returns (examples/<name>.json), every field kept.
func TestEveryExampleReads(t *testing.T) {
	for _, f := range All {
		if f.Name == "lock" {
			continue
		}
		t.Run(f.Name, func(t *testing.T) {
			v, err := readers[f.Name](example(t, source(f)))
			if err != nil {
				t.Fatalf("the example does not read: %v", err)
			}
			doc, err := f.Document(v)
			if err != nil {
				t.Fatal(err)
			}
			want, err := schema.Decode(example(t, f.Name+".json"))
			if err != nil {
				t.Fatal(err)
			}
			if !schema.Equal(doc, want) {
				t.Errorf("the Go type gives\n%s\nnot examples/%s.json", schema.Show(doc), f.Name)
			}
			// Byte for byte, in the schema's order: the type gives the document as the set stores it.
			if out, err := schema.Encode(doc); err != nil || !bytes.Equal(out, example(t, f.Name+".json")) {
				t.Errorf("the Go type's document is not examples/%s.json byte for byte (%v):\n%s", f.Name, err, out)
			}
			if err := f.Check(doc); err != nil {
				t.Errorf("as a writer writes it: %v", err)
			}
		})
	}
}

// TestWritersWriteTheExamples: where Bonsai writes the format, its writer writes the example back byte for byte, in
// the schema's order, held to the schema.
func TestWritersWriteTheExamples(t *testing.T) {
	for _, f := range All {
		if !f.Writes || f.Name == "lock" {
			continue
		}
		t.Run(f.Name, func(t *testing.T) {
			raw := example(t, source(f))
			v, err := readers[f.Name](raw)
			if err != nil {
				t.Fatal(err)
			}
			var out []byte
			switch x := v.(type) {
			case *Tasks:
				out, err = x.Encode()
			case *Sessions:
				out, err = x.Encode()
			case *Workspace:
				// bonsai.yaml's comments are init's template (step 5.1.5): the writer is held to the value.
				out, err = x.EncodeYAML(nil)
				if err == nil {
					again, err2 := ReadWorkspace(out)
					if err2 != nil {
						t.Fatalf("the YAML written does not read: %v\n%s", err2, out)
					}
					d1, _ := f.Document(again)
					d2, _ := f.Document(x)
					if !schema.Equal(d1, d2) {
						t.Errorf("the YAML written reads as another value:\n%s", out)
					}
					out2, _ := again.EncodeYAML(nil)
					if !bytes.Equal(out, out2) {
						t.Errorf("the YAML writer is not byte-stable")
					}
				}
				return
			default:
				out, err = f.Encode(v)
				raw = example(t, f.Name+".json")
				if f.Shape == JSONLine {
					var c bytes.Buffer
					if err := json.Compact(&c, raw); err != nil {
						t.Fatal(err)
					}
					raw = append(c.Bytes(), '\n')
				}
			}
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(out, raw) {
				t.Errorf("written back:\n%s\nwant:\n%s", out, raw)
			}
		})
	}
}

func TestReadsAMissingFieldAsNull(t *testing.T) {
	ladder, err := ReadLadder([]byte(`{"format": "bonsai.ladder/1", "workspace": "ws-7kq2m4xw5r3t6y2u7p4a5c3e2b", "proof": null, "green": null}`))
	if err != nil {
		t.Fatal(err)
	}
	if ladder.Task != nil || ladder.Proof != nil || ladder.Green || ladder.Rungs != nil {
		t.Errorf("missing fields read as %+v", ladder)
	}
	task, err := ReadTask([]byte("---\nformat: bonsai.task/1\nid: T-0001\ntitle: x\nstatus: todo\ndone_when:\n---\n"))
	if err != nil {
		t.Fatal(err)
	}
	if task.Lane != nil || task.DoneWhen != nil || task.Labels != nil {
		t.Errorf("missing fields read as %+v", task)
	}
}

func TestKeepsAnUnknownField(t *testing.T) {
	line := `{"format":"bonsai.log/1","id":"7d0c1f2e-3a4b-4c5d-8e6f-7a8b9c0d1e2f","at":"2026-10-08T14:08:00.000Z",` +
		`"workspace":"ws-7kq2m4xw5r3t6y2u7p4a5c3e2b","newer":{"b":1.50,"a":[true]},"event":"guard","labels":{}}`
	l, err := ReadLog([]byte(line))
	if err != nil {
		t.Fatal(err)
	}
	if len(l.Extra) != 1 || l.Extra[0].Key != "newer" {
		t.Fatalf("Extra holds %v", l.Extra)
	}
	out, err := l.Encode()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasSuffix(out, []byte(`"remote":null,"newer":{"b":1.50,"a":[true]}}`+"\n")) {
		t.Errorf("the unknown field is not kept, value for value, after the known ones: %s", out)
	}
	ws, err := ReadWorkspace([]byte("format: bonsai.workspace/1\nid: ws-7kq2m4xw5r3t6y2u7p4a5c3e2b\nname: x\nlater: [1, 2]\n" +
		"documents:\n  task: t\n  plan: work/plans\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(ws.Extra) != 1 || ws.Extra[0].Key != "later" || len(ws.Documents.Extra) != 1 {
		t.Errorf("Extra holds %v, documents' %v", ws.Extra, ws.Documents.Extra)
	}
}

// TestRefusals: each reading refusal names the field (and its line in a YAML or markdown file) and the next step.
func TestRefusals(t *testing.T) {
	cases := []struct {
		name   string
		read   func() error
		field  string
		line   int
		tooNew bool
		has    string
	}{
		{"a closed list's value", func() error {
			_, err := ReadTask([]byte("---\nformat: bonsai.task/1\nid: T-0001\ntitle: x\nstatus: finished\n---\n"))
			return err
		}, "status", 5, false, `"finished" is not one of`},
		{"a wrong type, in a list", func() error {
			_, err := ReadWorkspace([]byte("format: bonsai.workspace/1\nid: ws-7kq2m4xw5r3t6y2u7p4a5c3e2b\nname: x\n" +
				"ladder:\n  - rung: 0\n    kind: guard\n    required: true\n  - rung: 1\n    kind: command\n    timeout_s: \"ten\"\n"))
			return err
		}, "ladder[1].timeout_s", 10, false, "is string, want integer or null"},
		{"a pattern", func() error {
			_, err := ReadLog([]byte(`{"format":"bonsai.log/1","id":"x","event":"guard"}`))
			return err
		}, "id", 0, false, "does not match"},
		{"a wrong type, in JSON", func() error {
			_, err := ReadLadder([]byte(`{"format":"bonsai.ladder/1","green":"yes"}`))
			return err
		}, "green", 0, false, "is string, want boolean"},
		{"a newer major, markdown", func() error {
			_, err := ReadTask([]byte("---\nformat: bonsai.task/2\nstatus: [whatever\n---\n"))
			return err
		}, "", 2, true, "format too new"},
		{"a newer major, JSON", func() error {
			_, err := ReadAsk([]byte(`{"format":"bonsai.ask/2","op":12}`))
			return err
		}, "", 0, true, "format too new: \"bonsai.ask/2\""},
		{"another format", func() error {
			_, err := ReadRun([]byte("---\nformat: bonsai.task/1\nid: T-0001\n---\n"))
			return err
		}, "format", 2, false, "not bonsai.run/1"},
		{"format 0 where format 1 is all there is", func() error {
			_, err := ReadMemory([]byte("---\ntitle: x\n---\n"))
			return err
		}, "", 1, false, "reads as format 0"},
		{"bash by name in a hook", func() error {
			_, err := ReadPack([]byte("format: bonsai.pack/1\nid: p\nversion: \"1\"\nhooks:\n  - event: SessionStart\n" +
				"    command: \"bash run/x.sh\"\n    why: \"x\"\n"))
			return err
		}, "hooks[0].command", 6, false, "calls bash by name"},
		{"a table's header", func() error {
			_, err := ReadTasks([]byte("---\nformat: bonsai.tasks/1\n---\nActive task when none is named: none\n\n| Task |\n"))
			return err
		}, "tasks", 6, false, "header"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := c.read()
			var re *ReadError
			if !errors.As(err, &re) {
				t.Fatalf("not refused: %v", err)
			}
			if re.Field != c.field || re.Line != c.line || re.TooNew != c.tooNew || !strings.Contains(re.Msg, c.has) {
				t.Errorf("refused as field %q line %d too new %v: %s", re.Field, re.Line, re.TooNew, re.Msg)
			}
			if re.Next == "" || !strings.Contains(err.Error(), "; next: ") {
				t.Errorf("no next step: %v", err)
			}
			if s := err.Error(); s != ascii(s) {
				t.Errorf("not ASCII: %q", s)
			}
		})
	}
}

// TestReadsFormat0: task, run and state read under format 0 as they are (contract §2.3); nothing is refused for a
// field, and a field that fits reads.
func TestReadsFormat0(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "formats", "active-task", "format0-running", "main", "work", "tasks", "T-0901-format0.md"))
	if err != nil {
		t.Fatal(err)
	}
	task, err := ReadTask(raw)
	if err != nil {
		t.Fatal(err)
	}
	if task.Format0 == nil || task.ID != "T-0901" || task.Status != "running" || task.Lane == nil || *task.Lane != "light" {
		t.Errorf("read as %+v", task)
	}
	task, err = ReadTask([]byte("---\nid: T-0902\nstatus: In Progress\nlane: 3\ncost_usd: 1.5\n---\n"))
	if err != nil {
		t.Fatal(err)
	}
	if task.Status != "" || task.Lane != nil || task.ID != "T-0902" || len(task.Extra) != 1 {
		t.Errorf("a format-0 field that does not fit should read as missing: %+v", task)
	}
	if _, err := ReadRun([]byte("---\nid: R-1\noutcome: merged\n---\n")); err != nil {
		t.Error(err)
	}
	if _, err := ReadState([]byte("---\nupdated: 2026-10-08\n---\n")); err != nil {
		t.Error(err)
	}
	_, err = ReadTask([]byte("---\nid: T-0903\nlane: *light\n---\n"))
	var re *ReadError
	if !errors.As(err, &re) || re.Next == "" || !strings.Contains(re.Format, "format 0") {
		t.Errorf("a format-0 file yaml.mjs refuses is refused, with its next step: %v", err)
	}
}

func TestWritersRefuse(t *testing.T) {
	l := &Log{ID: "not-a-uuid", At: "2026-10-08T14:08:00.000Z", Workspace: "ws-7kq2m4xw5r3t6y2u7p4a5c3e2b", Event: "guard"}
	if _, err := l.Encode(); err == nil || !strings.Contains(err.Error(), "#/id") {
		t.Errorf("a log record with a bad id is written: %v", err)
	}
	l.ID = "7d0c1f2e-3a4b-4c5d-8e6f-7a8b9c0d1e2f"
	reason := strings.Repeat("x", 3000)
	l.Reason = &reason
	if _, err := l.Encode(); err == nil || !strings.Contains(err.Error(), "over the format's 2048") {
		t.Errorf("a log record over 2,048 bytes is written: %v", err)
	}
	l.Reason = nil
	out, err := l.Encode()
	if err != nil {
		t.Fatal(err)
	}
	back, err := ReadLog(out)
	if err != nil {
		t.Fatal(err)
	}
	again, _ := back.Encode()
	if !bytes.Equal(out, again) {
		t.Errorf("not byte-stable:\n%s\n%s", out, again)
	}
	if _, err := MustLookup("task").Encode(&Task{}); err == nil {
		t.Error("Bonsai writes no task, but the task has a writer")
	}
	e := &ErrorObject{Code: "x", Message: "m", Next: Next{Do: "d", Who: "nobody"}}
	if _, err := e.Encode(); err == nil || !strings.Contains(err.Error(), "next/who") {
		t.Errorf("an error object with who nobody is written: %v", err)
	}
	ex := &Check{Extra: schema.Object{{Key: "findings", Value: []any{}}}}
	if _, err := ex.Encode(); err == nil {
		t.Error("an Extra naming a known field is written")
	}
}

func TestStatusFormats(t *testing.T) {
	s := MustLookup("status").Schema()
	fs := sub(props(s), "formats")
	got := StatusFormats()
	if msgs := schema.Validate(fs, got); len(msgs) > 0 {
		t.Fatal(msgs)
	}
	if len(got) != len(All)-1 {
		t.Errorf("formats lists %d, want every format but the error object", len(got))
	}
	task, _ := got.Get("bonsai.task")
	if schema.Show(task) != `{"read":[0,1],"write":1}` {
		t.Errorf("bonsai.task: %s", schema.Show(task))
	}
}

func TestCallsBash(t *testing.T) {
	for cmd, want := range map[string]bool{
		"bash run.sh": true, "BASH run.sh": true, "/usr/bin/bash -c x": true, "env bash x": true, "sh -c \"bash x\"": true,
		`"C:\Program Files\Git\bin\bash.exe" x`: true, "a && bash b": true, "x|bash": true, "$(bash x)": true,
		"echo `bash`": true, "sh run.sh": false, "bonsai hook guard || exit 2": false, "bash-completion x": false,
		"rebash": false, "sh bashful.sh": false, "./run/bash.sh": false,
	} {
		if got := CallsBash(cmd); got != want {
			t.Errorf("CallsBash(%q) = %v", cmd, got)
		}
	}
}

func TestTableCells(t *testing.T) {
	title := "a | b"
	lone := "x\xed\xa0\x80y" // a lone surrogate as format 0 keeps it (WTF-8)
	tb := &Tasks{Active: TasksActive{Why: strPtr("many-running")}, Tasks: []TaskRow{
		{ID: "T-0002", Title: title, Status: "todo"}, {ID: "T-0001", Title: lone, Status: "done"}}}
	out, err := tb.Encode()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(out, []byte("| T-0002 | a \\| b | todo | | | |\n")) || !bytes.Contains(out, []byte("x\uFFFDy")) ||
		!bytes.Contains(out, []byte("none (many-running)")) {
		t.Errorf("written:\n%s", out)
	}
	back, err := ReadTasks(out)
	if err != nil {
		t.Fatal(err)
	}
	if back.Tasks[0].Title != title || back.Active.ID != nil || *back.Active.Why != "many-running" {
		t.Errorf("read back as %+v", back)
	}
	crlf := bytes.ReplaceAll(example(t, "tasks.md"), []byte("\n"), []byte("\r\n"))
	if _, err := ReadTasks(crlf); err != nil {
		t.Errorf("a CRLF checkout's table does not read: %v", err)
	}
}

func strPtr(s string) *string { return &s }
