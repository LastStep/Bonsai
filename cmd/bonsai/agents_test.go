package main

// bonsai agents (note 5.5.12) through the command line, in a linked project: refused outside one; no agent; then the
// person's session, a lead and its helper as the guard writes their state, their ends recorded by hook record's own
// code from the transcripts measured on Claude Code 2.1.300 (internal/recorder's test data), and a state file that
// does not read. --json against bonsai.agents/1, byte for byte the same on a second run; the human lines; the usage
// shown equal to what the recorder wrote. internal/agents' tests walk every field on a fuller fixture.

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/agents"
	"github.com/LastStep/Bonsai/internal/hookin"
	"github.com/LastStep/Bonsai/internal/recorder"
	"github.com/LastStep/Bonsai/internal/schema"
)

// agentsDoc runs bonsai agents --json, holds its output to bonsai.agents/1, and gives the document and its bytes.
func (c *cli) agentsDoc(exit int) (schema.Object, string) {
	c.t.Helper()
	code, out, errOut := c.run("", "agents", "--json")
	if code != exit {
		c.t.Fatalf("agents --json: exit %d, want %d\n%s%s", code, exit, out, errOut)
	}
	return fits(c.t, out, "agents"), out
}

// record runs hook record's own code on a payload in the project, at a fixed time.
func (c *cli) record(payload string, at time.Time) {
	c.t.Helper()
	env := map[string]string{hookin.ProjectEnv: c.root}
	if code := recorder.Record(recorder.Options{Stdin: strings.NewReader(payload), Getenv: func(k string) string { return env[k] },
		Now: func() time.Time { return at }}); code != 0 {
		c.t.Fatalf("hook record: exit %d", code)
	}
}

func TestAgentsCommand(t *testing.T) {
	c := newCLI(t)
	// Not linked: exit 4, not-linked naming bonsai init, the workspace null.
	if code, _, errOut := c.run("", "agents"); code != 4 || !strings.Contains(nextOf(errOut), "run bonsai init") {
		t.Errorf("not linked: %d %q", code, errOut)
	}
	doc, _ := c.agentsDoc(4)
	if w, _ := doc.Get("workspace"); w != nil || errorIn(doc, "agents").String("code") != "not-linked" {
		t.Errorf("not linked: %s", schema.Show(doc))
	}
	c.link()
	// No agent recorded yet.
	if code, out, _ := c.run("", "agents"); code != 0 || out != "bonsai agents  no agents recorded in .bonsai/local/agents/ or .bonsai/local/log/\n" {
		t.Errorf("no agents: %d %q", code, out)
	}
	doc, _ = c.agentsDoc(0)
	if a, _ := doc.Get("agents"); len(a.([]any)) != 0 {
		t.Errorf("no agents: %s", schema.Show(doc))
	}
	// A lead the person's session started, and a helper the lead started, as the guard writes their state.
	sess, helper := "393fe3e7-4666-4e2a-80e7-62cfe58f6a23", "ab8125621420b9524"
	start := time.Date(2026, 10, 16, 9, 0, 0, 0, time.UTC)
	s := agents.Store{Main: c.root}
	n := func(v int64) *int64 { return &v }
	lim := agents.Limits{Minutes: agents.Limit{Value: n(30), Max: n(30)}, Actions: agents.Limit{Value: n(150), Max: n(150)}}
	for _, h := range []agents.Header{
		{Agent: "a1", Session: sess, Type: "workflow:builder", Role: "builder", Kind: "lead", Writes: true, Limits: lim, First: agents.Stamp(start)},
		{Agent: helper, Session: sess, Type: "workflow:auditor", Role: "auditor", Kind: "helper", Limits: lim, First: agents.Stamp(start.Add(time.Minute))},
	} {
		if _, err := s.Create(h); err != nil {
			t.Fatal(err)
		}
	}
	for _, l := range [][2]string{{"a1", "session"}, {helper, "a1"}} {
		if err := s.Append(sess, l[0], agents.Event{E: agents.EventLink, At: agents.Stamp(start), Parent: l[1]}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := s.Call(sess, helper, start.Add(2*time.Minute), 1); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(s.SessionDir(sess), "a2.state"), []byte("{\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Their ends, recorded by hook record from the measured transcripts: the helper's SubagentStop, then the session's
	// SessionEnd.
	fixture := func(name string) string {
		b, _ := json.Marshal(filepath.ToSlash(filepath.Join(packageDir, "..", "..", "internal", "recorder", "testdata", "claude-2.1.300", name)))
		return string(b)
	}
	cwd, _ := json.Marshal(filepath.ToSlash(c.root))
	c.record(`{"session_id":"`+sess+`","cwd":`+string(cwd)+`,"hook_event_name":"SubagentStop","agent_id":"`+helper+
		`","agent_type":"workflow:auditor","agent_transcript_path":`+fixture("subagent.jsonl")+`}`, start.Add(3*time.Minute))
	c.record(`{"session_id":"`+sess+`","cwd":`+string(cwd)+`,"hook_event_name":"SessionEnd","reason":"other","transcript_path":`+
		fixture("session.jsonl")+`}`, start.Add(4*time.Minute))
	doc, first := c.agentsDoc(0)
	if _, again := c.agentsDoc(0); again != first {
		t.Errorf("a second run differs:\n%s\n%s", first, again)
	}
	got := map[string]string{}
	list, _ := doc.Get("agents")
	for _, a := range list.([]any) {
		o := a.(schema.Object)
		got[o.String("id")] = schema.Show(o)
	}
	opus := func(in, out, read, write string) string {
		return `[{"model":"claude-opus-5-5","input":` + in + `,"output":` + out + `,"cache_read":` + read + `,"cache_write":` + write +
			`,"kind":null,"reason":null}]`
	}
	logFile := `".bonsai/local/log/s-` + sess + `.ndjson"`
	want := map[string]string{
		sess: `{"id":"` + sess + `","session":"` + sess + `","parent":null,"role":null,"type":null,"task":null,"brief":null,` +
			`"model":"claude-opus-5-5","started":null,"ended":"2026-10-16T09:04:00.000Z","last_call":null,"counts":null,"limits":null,` +
			`"spent":false,"report":null,"log":` + logFile + `,"thread":null,"usage":` + opus("4", "383", "38638", "14439") + `}`,
		"a1": `{"id":"a1","session":"` + sess + `","parent":"session","role":"builder","type":"workflow:builder","task":null,"brief":null,` +
			`"model":null,"started":"2026-10-16T09:00:00.000Z","ended":null,"last_call":null,"counts":{"minutes":null,"actions":0,` +
			`"red_runs":0},"limits":{"minutes":{"value":30,"max":30},"actions":{"value":150,"max":150},"red_runs":{"value":null,"max":null}},` +
			`"spent":false,"report":null,"log":` + logFile + `,"thread":null,"usage":null}`,
		helper: `{"id":"` + helper + `","session":"` + sess + `","parent":"a1","role":"auditor","type":"workflow:auditor","task":null,` +
			`"brief":null,"model":"claude-opus-5-5","started":"2026-10-16T09:01:00.000Z","ended":"2026-10-16T09:03:00.000Z",` +
			`"last_call":null,"counts":{"minutes":1,"actions":1,"red_runs":0},"limits":{"minutes":{"value":30,"max":30},"actions":` +
			`{"value":150,"max":150},"red_runs":{"value":null,"max":null}},"spent":false,"report":null,"log":` + logFile +
			`,"thread":null,"usage":` + opus("4", "186", "6637", "9363") + `}`,
	}
	if len(got) != len(want) {
		t.Fatalf("%d agents:\n%s", len(got), first)
	}
	for id, w := range want {
		if got[id] != w {
			t.Errorf("%s:\n got %s\nwant %s", id, got[id], w)
		}
	}
	ids := []string{}
	for _, a := range list.([]any) {
		ids = append(ids, a.(schema.Object).String("id"))
	}
	if strings.Join(ids, " ") != "a1 "+helper+" "+sess {
		t.Errorf("the order (started, null last): %v", ids)
	}
	skipped, _ := doc.Get("skipped")
	if schema.Show(skipped) != `[{"path":".bonsai/local/agents/`+sess+`/a2.state","reason":"its first line is not a start line the guard writes"}]` {
		t.Errorf("skipped: %s", schema.Show(skipped))
	}
	// The human lines: one an agent, every field as recorded, then the skipped file.
	_, text, _ := c.run("", "agents")
	// Each agent a card: its id in the gutter (on a line of its own when wider than the gutter can hold), every field
	// a row, name then value.
	g := len(helper) + 2 // the gutter: the widest label it holds, the helper's id
	card := func(label string, fields ...string) string {
		var b strings.Builder
		b.WriteString("\n")
		if len(label) > g-2 {
			b.WriteString(label + "\n")
			label = ""
		}
		for i, f := range fields {
			name, value, _ := strings.Cut(f, " ")
			if i > 0 {
				label = ""
			}
			fmt.Fprintf(&b, "%-*s%-11s%s\n", g, label, name, value)
		}
		return b.String()
	}
	wantText := "bonsai agents  3 recorded\n" +
		card("a1", "session "+sess, "parent session", "role builder", "type workflow:builder", "task -", "brief -", "model -",
			"started 2026-10-16T09:00:00.000Z", "ended -", "last_call -", "counts minutes - actions 0 red_runs 0",
			"limits minutes 30 (max 30) actions 150 (max 150) red_runs - (max -)", "spent false", "report -",
			"log .bonsai/local/log/s-"+sess+".ndjson", "thread -", "usage -") +
		card(helper, "session "+sess, "parent a1", "role auditor", "type workflow:auditor", "task -", "brief -", "model claude-opus-5-5",
			"started 2026-10-16T09:01:00.000Z", "ended 2026-10-16T09:03:00.000Z", "last_call -", "counts minutes 1 actions 1 red_runs 0",
			"limits minutes 30 (max 30) actions 150 (max 150) red_runs - (max -)", "spent false", "report -",
			"log .bonsai/local/log/s-"+sess+".ndjson", "thread -", "usage claude-opus-5-5 input 4 output 186 cache_read 6637 cache_write 9363") +
		card(sess, "session "+sess, "parent -", "role -", "type -", "task -", "brief -", "model claude-opus-5-5", "started -",
			"ended 2026-10-16T09:04:00.000Z", "last_call -", "counts -", "limits -", "spent false", "report -",
			"log .bonsai/local/log/s-"+sess+".ndjson", "thread -", "usage claude-opus-5-5 input 4 output 383 cache_read 38638 cache_write 14439") +
		fmt.Sprintf("%-*s%s\n%s%s\n", g, "skipped", ".bonsai/local/agents/"+sess+"/a2.state", strings.Repeat(" ", g+4),
			"its first line is not a start line the guard writes")
	if text != wantText {
		t.Errorf("the human lines:\n%s\nwant\n%s", text, wantText)
	}
	// Bad input: a word it does not take.
	wantRefusal(t, agentsWord, []string{"agents", "now"}, 2, "bad-flag")
}
