package agents

// The view (note 5.5.12, 6) on a fixture project written as the guard, the recorder and bonsai message write it: the
// person's session; a lead it started and two helpers under the lead, one recorded spent and the other with no call
// for twenty minutes; an ended agent; an agent resumed after its end, with two ends' usage records; an agent of a type
// no pack declares; a named session; a session with a log file and no start record; a brief's actions: 5 below the
// role's 600, shown as 5; a report from another address in the lead's thread, not taken; three state files that do
// not read whole, skipped and named. Each entry equal to what was recorded, field by field; the tree rebuilt from
// parent alone; the same answer read twice and with the files made in another order; nothing written.

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/message"
	"github.com/LastStep/Bonsai/internal/record"
	"github.com/LastStep/Bonsai/internal/workspace"
)

const (
	viewWS   = "ws-aaaaaaaaaaaaaaaaaaaaaaaaaa"
	personS  = "6d1e2f3a-4b5c-4d6e-8f7a-9b0c1d2e3f4a"
	namedS   = "7e2f3a4b-5c6d-4e7f-8a9b-0c1d2e3f4a5b"
	noStartS = "8f3a4b5c-6d7e-4f8a-9b0c-1d2e3f4a5b6c"
)

var v0 = time.Date(2026, 10, 16, 9, 0, 0, 0, time.UTC)

func at(m, s int) time.Time {
	return v0.Add(time.Duration(m)*time.Minute + time.Duration(s)*time.Second)
}

func n64(v int64) *int64 { return &v }

func sp(s string) *string { return &s }

// fixture is the project as written, and the ids its messages were given.
type fixture struct {
	main               string
	b1, b2, b3, b4, b5 string // briefs
	r1, r2, r3         string // reports
}

// logRec appends one log record of a session as the recorder or the guard writes it.
func (f *fixture) logRec(t *testing.T, ev, session string, when time.Time, set func(l *format.Log)) {
	t.Helper()
	l := record.New(ev, record.Common{Workspace: viewWS, Session: session, Agent: "claude-code", At: when})
	if set != nil {
		set(l)
	}
	if _, err := record.WriteLog(f.main, l); err != nil {
		t.Fatal(err)
	}
}

func toolStart(sub, tool, category string) func(l *format.Log) {
	return func(l *format.Log) {
		l.AgentEvent, l.Tool, l.Category = sp("PreToolUse"), sp(tool), sp(category)
		if sub != "" {
			l.SubagentID = sp(sub)
		}
	}
}

func usageRec(sub, model string, in, out, read, write int64) func(l *format.Log) {
	return func(l *format.Log) {
		l.AgentEvent = sp("SubagentStop")
		if sub == "" {
			l.AgentEvent = sp("SessionEnd")
		} else {
			l.SubagentID = sp(sub)
		}
		if model == "" {
			l.Kind, l.Reason = sp("unknown"), sp("the transcript is not there")
			return
		}
		l.Model, l.Usage = sp(model), &format.LogUsage{Input: in, Output: out, CacheRead: read, CacheWrite: write}
	}
}

func subagent(sub string) func(l *format.Log) {
	return func(l *format.Log) { l.AgentEvent, l.SubagentID = sp("SubagentStop"), sp(sub) }
}

func spawnRec(sub, typ, parent string) func(l *format.Log) {
	return func(l *format.Log) {
		l.AgentEvent, l.SubagentID, l.SubagentType, l.Tool, l.Category, l.Parent = sp("PreToolUse"), sp(sub), sp(typ), sp("Agent"),
			sp("Agent"), sp(parent)
	}
}

// msg writes one message through bonsai message's own writer, at the time given; its id.
func (f *fixture) msg(t *testing.T, when time.Time, w message.Writing) string {
	t.Helper()
	p := message.Place{Workspace: viewWS, Local: workspace.Local{Main: f.main, Root: f.main}, Now: func() time.Time { return when }}
	done, e := message.Write(p, w, message.DefaultRules())
	if e != nil {
		t.Fatalf("%+v: %v", w, e)
	}
	return done.Message.ID
}

var helperLimits = Limits{Minutes: Limit{Value: n64(30), Max: n64(30)}, Actions: Limit{Value: n64(150), Max: n64(150)}}

var leadLimits = Limits{Minutes: Limit{Value: n64(180), Max: n64(360)}, Actions: Limit{Value: n64(600), Max: n64(600)},
	RedRuns: Limit{Value: n64(4), Max: n64(4)}}

// state writes an agent's state as the guard writes it: its header, its events and its call lines (each a time and
// the minutes figure the guard judged it by).
func state(t *testing.T, s Store, h Header, events []Event, calls ...[2]int) {
	t.Helper()
	if made, err := s.Create(h); !made || err != nil {
		t.Fatal(made, err)
	}
	for _, e := range events {
		if err := s.Append(h.Session, h.Agent, e); err != nil {
			t.Fatal(err)
		}
	}
	for _, c := range calls {
		if _, err := s.Call(h.Session, h.Agent, at(c[0], c[1]), int64(c[0]-int(mustTime(t, h.First).Sub(v0).Minutes()))); err != nil {
			t.Fatal(err)
		}
	}
}

func mustTime(t *testing.T, s string) time.Time {
	t.Helper()
	v, err := ParseStamp(s)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

// makeFixture writes the project; reversed runs its independent writes (each agent's state, each file) in the other
// order.
func makeFixture(t *testing.T, reversed bool) *fixture {
	t.Helper()
	f := &fixture{main: t.TempDir()}
	s := Store{Main: f.main}
	// The messages first: their ids are made as they are written.
	f.b1 = f.msg(t, at(1, 0), message.Writing{From: "orchestrator", To: "builder@T-0901", Kind: "brief", Task: "T-0901",
		Budget: sp("actions=5"), Body: "Build it."})
	f.b3 = f.msg(t, at(1, 30), message.Writing{From: "orchestrator", To: "verifier@T-0901", Kind: "brief", Thread: f.b1,
		Body: "Check it."})
	f.b2 = f.msg(t, at(2, 30), message.Writing{From: "builder@T-0901", To: "auditor@T-0901", Kind: "brief", Task: "T-0901",
		Budget: sp("actions=3"), Body: "Audit it."})
	f.r3 = f.msg(t, at(3, 50), message.Writing{From: "auditor@T-0901", To: "builder@T-0901", Kind: "report", Thread: f.b2,
		Status: "done", Body: "Audited."})
	f.b4 = f.msg(t, at(5, 50), message.Writing{From: "orchestrator", To: "researcher@T-0901", Kind: "brief", Task: "T-0901",
		Budget: sp("actions=10"), Body: "Find out."})
	f.b5 = f.msg(t, at(29, 0), message.Writing{From: "orchestrator", To: "researcher@T-0901", Kind: "brief", Thread: f.b4,
		Budget: sp("actions=20"), Body: "Go on."})
	f.r1 = f.msg(t, at(11, 0), message.Writing{From: "builder@T-0901", To: "orchestrator", Kind: "report", Thread: f.b1,
		Status: "done", Body: "Built."})
	f.r2 = f.msg(t, at(12, 0), message.Writing{From: "verifier@T-0901", To: "orchestrator", Kind: "report", Thread: f.b1,
		Status: "blocked", Body: "Not mine to say."})
	tie := func(when time.Time, brief, to, thread, from string, calls int64, lim Limits) Event {
		return Event{E: EventTie, At: StampMilli(when), Brief: brief, To: to, Thread: thread, From: from, Calls: calls, Limits: &lim}
	}
	link := func(when time.Time, parent, call string, briefs ...string) Event {
		return Event{E: EventLink, At: Stamp(when), Parent: parent, Call: call, Briefs: briefs}
	}
	leadTied := leadLimits
	leadTied.Actions.Value = n64(5)
	auditTied := helperLimits
	auditTied.Actions.Value = n64(3)
	researchTied := helperLimits
	researchTied.Actions.Value = n64(10)
	researchRetied := helperLimits
	researchRetied.Actions.Value = n64(20)
	writes := []func(){
		// The person's session's log: its start, its own calls, the spawn records, the subagents' ends and usage.
		func() {
			f.logRec(t, "session_start", personS, at(0, 0), func(l *format.Log) {
				l.AgentEvent, l.Source, l.Model = sp("SessionStart"), sp("startup"), sp("claude-opus-5-5")
			})
			f.logRec(t, "tool_start", personS, at(1, 0), toolStart("", "Agent", "Agent"))
			f.logRec(t, "spawn", personS, at(2, 0), spawnRec("a1lead", "workflow:builder", "session"))
			f.logRec(t, "tool_start", personS, at(2, 0), toolStart("a1lead", "Bash", "Shell"))
			f.logRec(t, "spawn", personS, at(3, 0), spawnRec("a2spent", "workflow:auditor", "a1lead"))
			f.logRec(t, "tool_start", personS, at(3, 10), toolStart("a2spent", "Bash", "Shell"))
			f.logRec(t, "tool_start", personS, at(3, 40), toolStart("a2spent", "Read", "Read"))
			f.logRec(t, "spawn", personS, at(4, 0), spawnRec("a3quiet", "workflow:auditor", "a1lead"))
			f.logRec(t, "tool_start", personS, at(4, 30), toolStart("a3quiet", "Grep", "Search"))
			f.logRec(t, "spawn", personS, at(5, 0), spawnRec("a4ended", "workflow:runner", "session"))
			f.logRec(t, "tool_start", personS, at(5, 30), toolStart("a4ended", "Bash", "Ladder"))
			f.logRec(t, "spawn", personS, at(6, 0), spawnRec("a5resumed", "workflow:researcher", "session"))
			f.logRec(t, "tool_start", personS, at(6, 30), toolStart("a5resumed", "WebSearch", "Web"))
			f.logRec(t, "spawn", personS, at(7, 0), spawnRec("a6explore", "Explore", "session"))
			f.logRec(t, "tool_start", personS, at(7, 30), toolStart("a6explore", "Read", "Read"))
			f.logRec(t, "subagent_stop", personS, at(7, 50), subagent("a6explore"))
			f.logRec(t, "usage", personS, at(7, 50), usageRec("a6explore", "", 0, 0, 0, 0))
			f.logRec(t, "subagent_stop", personS, at(8, 0), subagent("a4ended"))
			f.logRec(t, "usage", personS, at(8, 0), usageRec("a4ended", "claude-haiku-4-5", 100, 200, 300, 400))
			f.logRec(t, "subagent_stop", personS, at(9, 0), subagent("a5resumed"))
			f.logRec(t, "usage", personS, at(9, 0), usageRec("a5resumed", "claude-sonnet-4-5", 10, 20, 30, 40))
			f.logRec(t, "tool_start", personS, at(10, 0), toolStart("a1lead", "Agent", "Agent"))
			// The resumed agent's second end: newer records, each its whole transcript, its last reply's model last.
			f.logRec(t, "tool_start", personS, at(30, 0), toolStart("a5resumed", "Bash", "Shell"))
			f.logRec(t, "subagent_stop", personS, at(31, 0), subagent("a5resumed"))
			f.logRec(t, "usage", personS, at(31, 0), usageRec("a5resumed", "claude-haiku-4-5", 1, 2, 3, 4))
			f.logRec(t, "usage", personS, at(31, 0), usageRec("a5resumed", "claude-sonnet-4-5", 15, 25, 35, 45))
			f.logRec(t, "session_end", personS, at(40, 0), func(l *format.Log) { l.AgentEvent, l.Reason = sp("SessionEnd"), sp("prompt_input_exit") })
			f.logRec(t, "usage", personS, at(40, 0), usageRec("", "claude-opus-5-5", 1000, 2000, 3000, 4000))
		},
		// A named session's log, and a session's log with no start record.
		func() {
			f.logRec(t, "session_start", namedS, at(20, 0), func(l *format.Log) {
				l.AgentEvent, l.Source, l.Model, l.Task, l.Role = sp("SessionStart"), sp("startup"), sp("claude-sonnet-4-5"), sp("T-0902"), sp("builder")
			})
			f.logRec(t, "tool_start", namedS, at(21, 0), toolStart("", "Bash", "Shell"))
		},
		func() { f.logRec(t, "tool_start", noStartS, at(50, 0), toolStart("", "Read", "Read")) },
		// The lead: its brief's actions: 5, below the role's 600; two red climbs in a row.
		func() {
			state(t, s, Header{Agent: "a1lead", Session: personS, Type: "workflow:builder", Role: "builder", Pack: "workflow", Kind: "lead",
				Writes: true, MaxTurns: n64(200), Limits: leadLimits, First: Stamp(at(2, 0)), ReadFrom: "HEAD"},
				[]Event{link(at(2, 0), "session", "toolu_01", f.b1), tie(at(2, 0), f.b1, "builder@T-0901", f.b1, "orchestrator", 0, leadTied),
					{E: EventClimb, At: StampMilli(at(5, 0)), Task: "T-0901"}, {E: EventRed, At: Stamp(at(6, 0)), Task: "T-0901", N: 1},
					{E: EventClimb, At: StampMilli(at(7, 0)), Task: "T-0901"}, {E: EventRed, At: Stamp(at(8, 0)), Task: "T-0901", N: 2}},
				[2]int{2, 0}, [2]int{5, 0}, [2]int{10, 0})
		},
		// A helper recorded spent (its brief's 3 actions), and reported.
		func() {
			state(t, s, Header{Agent: "a2spent", Session: personS, Type: "workflow:auditor", Role: "auditor", Pack: "workflow", Kind: "helper",
				Limits: helperLimits, First: Stamp(at(3, 0)), ReadFrom: "HEAD"},
				[]Event{link(at(3, 0), "a1lead", "toolu_02", f.b2), tie(at(3, 0), f.b2, "auditor@T-0901", f.b2, "builder@T-0901", 0, auditTied),
					{E: EventSpent, At: Stamp(at(3, 40)), Measure: "actions", Count: 4, Of: 3}, {E: EventReported, At: Stamp(at(3, 50))}},
				[2]int{3, 0}, [2]int{3, 10}, [2]int{3, 20}, [2]int{3, 40})
		},
		// A helper with no call since 09:04:30, no brief tied: its role's limits.
		func() {
			state(t, s, Header{Agent: "a3quiet", Session: personS, Type: "workflow:auditor", Role: "auditor", Pack: "workflow", Kind: "helper",
				Limits: helperLimits, First: Stamp(at(4, 0)), ReadFrom: "HEAD"}, []Event{link(at(4, 0), "a1lead", "toolu_03")}, [2]int{4, 30})
		},
		// An ended agent.
		func() {
			state(t, s, Header{Agent: "a4ended", Session: personS, Type: "workflow:runner", Role: "runner", Pack: "workflow", Kind: "helper",
				Limits: helperLimits, First: Stamp(at(5, 0)), ReadFrom: "HEAD"}, []Event{link(at(5, 0), "session", "toolu_04")}, [2]int{5, 30})
		},
		// An agent resumed after its end by a newer brief: re-tied, a fresh window after its two calls.
		func() {
			state(t, s, Header{Agent: "a5resumed", Session: personS, Type: "workflow:researcher", Role: "researcher", Pack: "workflow",
				Kind: "helper", Limits: helperLimits, First: Stamp(at(6, 0)), ReadFrom: "HEAD"},
				[]Event{link(at(6, 0), "session", "toolu_05", f.b4), tie(at(6, 0), f.b4, "researcher@T-0901", f.b4, "orchestrator", 0, researchTied),
					{E: EventOffer, At: StampMilli(at(29, 30)), Brief: f.b5, From: "session"},
					tie(at(29, 50), f.b5, "researcher@T-0901", f.b4, "orchestrator", 2, researchRetied)},
				[2]int{6, 30}, [2]int{8, 0}, [2]int{30, 0})
		},
		// An agent of a type no pack declares, held as a helper.
		func() {
			state(t, s, Header{Agent: "a6explore", Session: personS, Type: "Explore", Kind: "helper", Limits: helperLimits,
				First: Stamp(at(7, 0)), ReadFrom: "HEAD"}, []Event{link(at(7, 0), "session", "toolu_06")}, [2]int{7, 30})
		},
		// A named session (BONSAI_ROLE and BONSAI_TASK), counted by its session.
		func() {
			state(t, s, Header{Agent: namedS, Session: namedS, Role: "builder", Pack: "workflow", Kind: "lead", Writes: true,
				Limits: leadLimits, Named: true, Task: "T-0902", First: Stamp(at(21, 0)), ReadFrom: "HEAD"}, nil, [2]int{21, 0})
		},
		// Three state files that do not read whole.
		func() {
			write(t, filepath.Join(s.SessionDir(personS), "a7bad.state"), "not a start line\n")
			h, _ := json.Marshal(Header{Agent: "a8cut", Session: personS, Kind: "helper", First: Stamp(at(9, 0))})
			write(t, filepath.Join(s.SessionDir(personS), "a8cut.state"), string(h))
			state(t, s, Header{Agent: "a9reset", Session: personS, Type: "workflow:auditor", Role: "auditor", Kind: "helper",
				Limits: helperLimits, First: Stamp(at(9, 0))}, []Event{tie(at(9, 0), f.b2, "auditor@T-0901", f.b2, "builder@T-0901", 5, auditTied)},
				[2]int{9, 0}, [2]int{9, 10})
		},
	}
	if reversed {
		for i, j := 0, len(writes)-1; i < j; i, j = i+1, j-1 {
			writes[i], writes[j] = writes[j], writes[i]
		}
	}
	for _, w := range writes {
		w()
	}
	return f
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// snapshot is every path under a folder with its type, and each file's content hash: what a read must leave as it
// was. Times are not compared: NTFS updates a folder's modification time lazily, after the writes that made the
// fixture (measured on Windows, 15 Oct).
func snapshot(t *testing.T, root string) string {
	t.Helper()
	var lines []string
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		line := filepath.ToSlash(p) + " " + d.Type().String()
		if d.Type().IsRegular() {
			raw, err := os.ReadFile(p)
			if err != nil {
				return err
			}
			sum := sha256.Sum256(raw)
			line += " " + hex.EncodeToString(sum[:])
		}
		lines = append(lines, line)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(lines)
	return strings.Join(lines, "\n")
}

// entries are the view's agents, each as its JSON line in bonsai.agents/1's order.
func entries(t *testing.T, v *Viewed) []string {
	t.Helper()
	var out []string
	for _, a := range v.Agents {
		b, err := json.Marshal(a)
		if err != nil {
			t.Fatal(err)
		}
		out = append(out, string(b))
	}
	return out
}

func TestView(t *testing.T) {
	f := makeFixture(t, false)
	st, err := message.Read(f.main)
	if err != nil {
		t.Fatal(err)
	}
	before := snapshot(t, f.main)
	v, err := Store{Main: f.main}.View(st)
	if err != nil {
		t.Fatal(err)
	}
	if after := snapshot(t, f.main); after != before {
		t.Errorf("the view wrote:\n%s\n---\n%s", before, after)
	}
	logOf := func(s string) string { return `".bonsai/local/log/s-` + s + `.ndjson"` }
	helper := `{"minutes":{"value":30,"max":30},"actions":{"value":150,"max":150},"red_runs":{"value":null,"max":null}}`
	none := `"counts":null,"limits":null,"spent":false,"report":null`
	want := []string{
		// The person's session, from its log alone.
		`{"id":"` + personS + `","session":"` + personS + `","parent":null,"role":null,"type":null,"task":null,"brief":null,` +
			`"model":"claude-opus-5-5","started":"2026-10-16T09:00:00.000Z","ended":"2026-10-16T09:40:00.000Z",` +
			`"last_call":{"at":"2026-10-16T09:01:00.000Z","tool":"Agent","category":"Agent"},` + none + `,"log":` + logOf(personS) +
			`,"thread":null,"usage":[{"model":"claude-opus-5-5","input":1000,"output":2000,"cache_read":3000,"cache_write":4000,"kind":null,"reason":null}]}`,
		// The lead: its brief's actions 5 (the role's default 600), its own report (r1), never the verifier's newer one (r2).
		`{"id":"a1lead","session":"` + personS + `","parent":"session","role":"builder","type":"workflow:builder","task":"T-0901",` +
			`"brief":"` + f.b1 + `","model":null,"started":"2026-10-16T09:02:00.000Z","ended":null,` +
			`"last_call":{"at":"2026-10-16T09:10:00.000Z","tool":"Agent","category":"Agent"},"counts":{"minutes":8,"actions":3,"red_runs":2},` +
			`"limits":{"minutes":{"value":180,"max":360},"actions":{"value":5,"max":600},"red_runs":{"value":4,"max":4}},"spent":false,` +
			`"report":{"id":"` + f.r1 + `","status":"done"},"log":` + logOf(personS) + `,"thread":"` + f.b1 + `","usage":null}`,
		// A helper recorded spent, past its brief's 3 actions, with its report.
		`{"id":"a2spent","session":"` + personS + `","parent":"a1lead","role":"auditor","type":"workflow:auditor","task":"T-0901",` +
			`"brief":"` + f.b2 + `","model":null,"started":"2026-10-16T09:03:00.000Z","ended":null,` +
			`"last_call":{"at":"2026-10-16T09:03:40.000Z","tool":"Read","category":"Read"},"counts":{"minutes":0,"actions":4,"red_runs":0},` +
			`"limits":{"minutes":{"value":30,"max":30},"actions":{"value":3,"max":150},"red_runs":{"value":null,"max":null}},"spent":true,` +
			`"report":{"id":"` + f.r3 + `","status":"done"},"log":` + logOf(personS) + `,"thread":"` + f.b2 + `","usage":null}`,
		// A helper whose last call was at 09:04:30: shown as recorded, nothing worked out from it.
		`{"id":"a3quiet","session":"` + personS + `","parent":"a1lead","role":"auditor","type":"workflow:auditor","task":null,` +
			`"brief":null,"model":null,"started":"2026-10-16T09:04:00.000Z","ended":null,` +
			`"last_call":{"at":"2026-10-16T09:04:30.000Z","tool":"Grep","category":"Search"},"counts":{"minutes":0,"actions":1,"red_runs":0},` +
			`"limits":` + helper + `,"spent":false,"report":null,"log":` + logOf(personS) + `,"thread":null,"usage":null}`,
		// An ended agent, its end's usage.
		`{"id":"a4ended","session":"` + personS + `","parent":"session","role":"runner","type":"workflow:runner","task":null,` +
			`"brief":null,"model":"claude-haiku-4-5","started":"2026-10-16T09:05:00.000Z","ended":"2026-10-16T09:08:00.000Z",` +
			`"last_call":{"at":"2026-10-16T09:05:30.000Z","tool":"Bash","category":"Ladder"},"counts":{"minutes":0,"actions":1,"red_runs":0},` +
			`"limits":` + helper + `,"spent":false,"report":null,"log":` + logOf(personS) + `,"thread":null,"usage":[{"model":"claude-haiku-4-5",` +
			`"input":100,"output":200,"cache_read":300,"cache_write":400,"kind":null,"reason":null}]}`,
		// The resumed agent: its newer brief (b5, in b4's thread) with its actions 20 and a fresh window; its newest end and
		// that end's two records, as written (never summed with the first end's).
		`{"id":"a5resumed","session":"` + personS + `","parent":"session","role":"researcher","type":"workflow:researcher",` +
			`"task":"T-0901","brief":"` + f.b5 + `","model":"claude-sonnet-4-5","started":"2026-10-16T09:06:00.000Z",` +
			`"ended":"2026-10-16T09:31:00.000Z","last_call":{"at":"2026-10-16T09:30:00.000Z","tool":"Bash","category":"Shell"},` +
			`"counts":{"minutes":24,"actions":1,"red_runs":0},"limits":{"minutes":{"value":30,"max":30},"actions":{"value":20,"max":150},` +
			`"red_runs":{"value":null,"max":null}},"spent":false,"report":null,"log":` + logOf(personS) + `,"thread":"` + f.b4 + `",` +
			`"usage":[{"model":"claude-haiku-4-5","input":1,"output":2,"cache_read":3,"cache_write":4,"kind":null,"reason":null},` +
			`{"model":"claude-sonnet-4-5","input":15,"output":25,"cache_read":35,"cache_write":45,"kind":null,"reason":null}]}`,
		// A type no pack declares: role null; its end's usage unknown, with the reason.
		`{"id":"a6explore","session":"` + personS + `","parent":"session","role":null,"type":"Explore","task":null,"brief":null,` +
			`"model":null,"started":"2026-10-16T09:07:00.000Z","ended":"2026-10-16T09:07:50.000Z",` +
			`"last_call":{"at":"2026-10-16T09:07:30.000Z","tool":"Read","category":"Read"},"counts":{"minutes":0,"actions":1,"red_runs":0},` +
			`"limits":` + helper + `,"spent":false,"report":null,"log":` + logOf(personS) + `,"thread":null,"usage":[{"model":null,` +
			`"input":null,"output":null,"cache_read":null,"cache_write":null,"kind":"unknown","reason":"the transcript is not there"}]}`,
		// A named session: its task, its model from its start, no parent Bonsai sees.
		`{"id":"` + namedS + `","session":"` + namedS + `","parent":null,"role":"builder","type":null,"task":"T-0902","brief":null,` +
			`"model":"claude-sonnet-4-5","started":"2026-10-16T09:21:00.000Z","ended":null,` +
			`"last_call":{"at":"2026-10-16T09:21:00.000Z","tool":"Bash","category":"Shell"},"counts":{"minutes":0,"actions":1,"red_runs":0},` +
			`"limits":{"minutes":{"value":180,"max":360},"actions":{"value":600,"max":600},"red_runs":{"value":4,"max":4}},"spent":false,` +
			`"report":null,"log":` + logOf(namedS) + `,"thread":null,"usage":null}`,
		// A session with no start record: started null, so last.
		`{"id":"` + noStartS + `","session":"` + noStartS + `","parent":null,"role":null,"type":null,"task":null,"brief":null,` +
			`"model":null,"started":null,"ended":null,"last_call":{"at":"2026-10-16T09:50:00.000Z","tool":"Read","category":"Read"},` +
			none + `,"log":` + logOf(noStartS) + `,"thread":null,"usage":null}`,
	}
	got := entries(t, v)
	if len(got) != len(want) {
		t.Fatalf("%d entries, want %d:\n%s", len(got), len(want), strings.Join(got, "\n"))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("entry %d:\n got %s\nwant %s", i, got[i], want[i])
		}
	}
	skipped := fmt.Sprint(v.Skipped)
	wantSkipped := "[{.bonsai/local/agents/" + personS + "/a7bad.state its first line is not a start line the guard writes []} " +
		"{.bonsai/local/agents/" + personS + "/a8cut.state its last line is not whole (cut short) []} " +
		"{.bonsai/local/agents/" + personS + "/a9reset.state its calls file holds 2 calls, fewer than the 5 its last tie line counts before its window []}]"
	if skipped != wantSkipped {
		t.Errorf("skipped:\n got %s\nwant %s", skipped, wantSkipped)
	}
	// The whole document fits bonsai.agents/1.
	if _, err := (&format.Agents{Agents: v.Agents, Skipped: v.Skipped}).Encode(); err != nil {
		t.Fatal(err)
	}
	// The tree, rebuilt from parent alone: session names the session's own entry, an id the entry with that id.
	if got, want := tree(v.Agents), strings.Join([]string{
		personS, "  a1lead", "    a2spent", "    a3quiet", "  a4ended", "  a5resumed", "  a6explore", namedS, noStartS}, "\n"); got != want {
		t.Errorf("the tree:\n%s\nwant\n%s", got, want)
	}
	// The same answer read again, and from the same files written in the other order.
	again, err := Store{Main: f.main}.View(st)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(entries(t, again), "\n") != strings.Join(got, "\n") || fmt.Sprint(again.Skipped) != skipped {
		t.Errorf("a second read differs")
	}
}

// tree draws the agents' tree from parent alone, each child under its parent in the view's order.
func tree(agents []format.Agent) string {
	children := map[string][]string{}
	var roots []string
	for _, a := range agents {
		key := a.Session + "/" + a.ID
		switch {
		case a.Parent == nil:
			roots = append(roots, key)
		case *a.Parent == "session":
			children[a.Session+"/"+a.Session] = append(children[a.Session+"/"+a.Session], key)
		default:
			children[a.Session+"/"+*a.Parent] = append(children[a.Session+"/"+*a.Parent], key)
		}
	}
	var lines []string
	var walk func(key, indent string)
	walk = func(key, indent string) {
		_, id, _ := strings.Cut(key, "/")
		lines = append(lines, indent+id)
		for _, c := range children[key] {
			walk(c, indent+"  ")
		}
	}
	for _, r := range roots {
		walk(r, "")
	}
	return strings.Join(lines, "\n")
}

// The same files written in the other order give the same answer, byte for byte (the message ids aside, which are
// made fresh each time).
func TestViewOrder(t *testing.T) {
	read := func(reversed bool) string {
		f := makeFixture(t, reversed)
		st, err := message.Read(f.main)
		if err != nil {
			t.Fatal(err)
		}
		v, err := Store{Main: f.main}.View(st)
		if err != nil {
			t.Fatal(err)
		}
		out := strings.Join(entries(t, v), "\n") + fmt.Sprint(v.Skipped)
		for _, id := range []string{f.b1, f.b2, f.b3, f.b4, f.b5, f.r1, f.r2, f.r3} {
			out = strings.ReplaceAll(out, id, "M-id")
		}
		return out
	}
	if a, b := read(false), read(true); a != b {
		t.Errorf("the other order:\n%s\n---\n%s", a, b)
	}
}

// A project with no agents' state and no log: no agent, nothing skipped; an agents folder that is a file is skipped
// and named, never an error.
func TestViewEmpty(t *testing.T) {
	main := t.TempDir()
	v, err := Store{Main: main}.View(&message.Store{})
	if err != nil || len(v.Agents) != 0 || len(v.Skipped) != 0 {
		t.Fatalf("%v %+v", err, v)
	}
	write(t, filepath.Join(main, ".bonsai", "local", "agents"), "a file\n")
	v, err = Store{Main: main}.View(&message.Store{})
	if err != nil || len(v.Skipped) != 1 || v.Skipped[0].Path != ".bonsai/local/agents" {
		t.Fatalf("%v %+v", err, v)
	}
}

// Each way a state file is read strictly: skipped and named, never shown with counts it does not hold.
func TestViewStrict(t *testing.T) {
	cases := []struct {
		name  string
		make  func(t *testing.T, s Store)
		wantR string
	}{
		{"its first line names another agent", func(t *testing.T, s Store) {
			h, _ := json.Marshal(Header{Agent: "other", Session: personS, First: Stamp(v0)})
			write(t, filepath.Join(s.SessionDir(personS), "x1.state"), string(h)+"\n")
		}, "its first line names another agent or session than its file"},
		{"an event line that does not read", func(t *testing.T, s Store) {
			state(t, s, Header{Agent: "x1", Session: personS, Kind: "helper", First: Stamp(v0)}, nil)
			f, _ := os.OpenFile(filepath.Join(s.SessionDir(personS), "x1.state"), os.O_APPEND|os.O_WRONLY, 0)
			_, _ = f.WriteString("{\"e\":\n")
			_ = f.Close()
		}, "its line 2 is not an event line the guard writes"},
		{"a role that is no role's name", func(t *testing.T, s Store) {
			state(t, s, Header{Agent: "x1", Session: personS, Role: "Builder!", First: Stamp(v0)}, nil)
		}, "its first line names a role that is not a role's name"},
		{"a limit below 0", func(t *testing.T, s Store) {
			state(t, s, Header{Agent: "x1", Session: personS, First: Stamp(v0), Limits: Limits{Actions: Limit{Value: n64(-1)}}}, nil)
		}, "its first line holds a limit below 0"},
		{"a tie naming no message", func(t *testing.T, s Store) {
			state(t, s, Header{Agent: "x1", Session: personS, First: Stamp(v0)}, []Event{{E: EventTie, At: Stamp(v0), Brief: "M-x", Thread: "M-x"}})
		}, "its line 2 ties no message id"},
		{"a red count below 0", func(t *testing.T, s Store) {
			state(t, s, Header{Agent: "x1", Session: personS, First: Stamp(v0)}, []Event{{E: EventRed, At: Stamp(v0), N: -1}})
		}, "its line 2 counts red runs below 0"},
		{"a calls file cut short", func(t *testing.T, s Store) {
			state(t, s, Header{Agent: "x1", Session: personS, First: Stamp(v0)}, nil, [2]int{0, 1})
			f, _ := os.OpenFile(filepath.Join(s.SessionDir(personS), "x1.calls"), os.O_APPEND|os.O_WRONLY, 0)
			_, _ = f.WriteString("2026-10-16T09:")
			_ = f.Close()
		}, "its calls file is not whole call lines (42 bytes)"},
		{"a calls file whose last line does not read", func(t *testing.T, s Store) {
			state(t, s, Header{Agent: "x1", Session: personS, First: Stamp(v0)}, nil)
			write(t, filepath.Join(s.SessionDir(personS), "x1.calls"), strings.Repeat("x", CallLen))
		}, "its calls file's last line does not read"},
		{"a state that is a folder's name for a file", func(t *testing.T, s Store) {
			// VG's item K: named in skipped, never passed over unseen, and never opened.
			if err := os.MkdirAll(filepath.Join(s.SessionDir(personS), "x1.state"), 0o755); err != nil {
				t.Fatal(err)
			}
		}, "it is a folder, not a regular file, so it is not opened"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := Store{Main: t.TempDir()}
			c.make(t, s)
			v, err := s.View(&message.Store{})
			if err != nil {
				t.Fatal(err)
			}
			if len(v.Agents) != 0 || len(v.Skipped) != 1 || v.Skipped[0].Reason != c.wantR ||
				v.Skipped[0].Path != ".bonsai/local/agents/"+personS+"/x1.state" {
				t.Errorf("%+v", v)
			}
		})
	}
}
