package guard

// The roles' rules and the stall stops (step 5.5.11), on the payloads Claude Code 2.1.300 sent on WSL (15 Oct, the
// run report's payload table): each fixture in testdata/claude-2.1.300/ is a redacted payload of one route an agent is
// started by, read here with its folders set to the test's own. The fixture pack declares six roles: orchestrator (the
// session's), lead (a lead that writes), verifier (a lead that changes nothing), and helper, tiny and fragile (helpers
// that change nothing; tiny's budget is 3 actions and 1 red run, fragile's 100 actions and 1 red run), so the measured
// types lead and helper name declared roles by their bare names, as a project's own agents of a pack's role names
// would.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/agents"
	"github.com/LastStep/Bonsai/internal/engine"
	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/hookin"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/testpack"
	"github.com/LastStep/Bonsai/internal/workspace"
)

const rolesWS = "ws-bbbbbbbbbbbbbbbbbbbbbbbbbb"

// roleItem is one roles item of a pack.yaml; limits are minutes, actions and red_runs, each "default max".
func roleItem(name, kind string, session, writes bool, limits [3]string) string {
	s := "  - name: " + name + "\n    kind: " + kind + "\n    session: " + fmt.Sprint(session) + "\n    writes: " +
		fmt.Sprint(writes) + "\n    max_turns: null\n"
	for i, l := range []string{"minutes", "actions", "red_runs"} {
		d, m, _ := strings.Cut(limits[i], " ")
		s += "    " + l + ":\n      default: " + d + "\n      max: " + m + "\n"
	}
	return s
}

var fixtureRoles = roleItem("orchestrator", "lead", true, true, [3]string{"null null", "null null", "null null"}) +
	roleItem("lead", "lead", false, true, [3]string{"180 180", "600 600", "4 4"}) +
	roleItem("verifier", "lead", false, false, [3]string{"120 120", "400 400", "null null"}) +
	roleItem("helper", "helper", false, false, [3]string{"30 30", "150 150", "null null"}) +
	roleItem("tiny", "helper", false, false, [3]string{"null null", "3 10", "1 2"}) +
	roleItem("fragile", "helper", false, false, [3]string{"null null", "100 100", "1 2"})

// rolesLock is a lock whose one pack, workflow, declares roles (none for "").
func rolesLock(t *testing.T, roles string) []byte {
	t.Helper()
	list := " []\n"
	if roles != "" {
		list = "\n" + roles
	}
	pack, err := format.ReadPack([]byte("format: bonsai.pack/1\nid: workflow\nversion: \"1\"\nroles:" + list))
	if err != nil {
		t.Fatal(err)
	}
	decl, err := (&format.Declares{Roles: pack.Roles}).Object()
	if err != nil {
		t.Fatal(err)
	}
	l := &workspace.Lock{WrittenBy: "0.5.0", Files: map[string]workspace.LockedFile{}, Packs: []workspace.LockedPack{{
		ID: "workflow", Source: "https://example.invalid/workflow.git", Version: "1", Commit: strings.Repeat("a", 40),
		SHA256: strings.Repeat("b", 64), Declares: decl, PathSet: true}}}
	raw, err := l.Encode()
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

const rolesYAML = `format: bonsai.workspace/1
id: ` + rolesWS + `
name: roles-test
packs:
  - id: workflow
    source: "https://example.invalid/workflow.git"
    path: null
    ref: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
    roles: null
protected: ["protected.txt"]
person_only: []
never_edit: []
`

// rolesProject makes a linked project whose lock declares the fixture roles, committed in a real git repository and
// recorded on a Bonsai home of the test's own; extra files are written and committed with it. It gives its real path.
func rolesProject(t *testing.T, extra map[string]string) string {
	t.Helper()
	tmp, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	home := testpack.Isolate(t, tmp)
	dir := filepath.Join(tmp, "proj")
	files := map[string]string{"bonsai.yaml": rolesYAML, ".bonsai/lock.json": string(rolesLock(t, fixtureRoles)),
		"free.txt": "free\n", "protected.txt": "protected\n"}
	for k, v := range extra {
		files[k] = v
	}
	for name, body := range files {
		writeFile(t, filepath.Join(dir, filepath.FromSlash(name)), body)
	}
	testpack.Git(t, dir, "init", "-q")
	testpack.Git(t, dir, "add", "-A")
	testpack.Git(t, dir, "commit", "-q", "-m", "the fixture")
	if _, err := workspace.RecordCheckout(home, dir, rolesWS, time.Now()); err != nil {
		t.Fatal(err)
	}
	return dir
}

// fixture reads a measured payload, its folders set: ~/proj to dir, ~/.claude/projects/p to transcripts.
func fixture(t *testing.T, rel, dir, transcripts string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", "claude-2.1.300", filepath.FromSlash(rel)))
	if err != nil {
		t.Fatal(err)
	}
	s := strings.ReplaceAll(string(raw), "~/.claude/projects/p", filepath.ToSlash(transcripts))
	return strings.ReplaceAll(s, "~/proj", filepath.ToSlash(dir))
}

// hookCall is a PreToolUse payload of the routes the fixtures show, made fresh: session, an agent's id and type ("" for
// none), the tool and its input. (Not call: guard_windows_test.go has a func call.)
type hookCall struct {
	session, agentID, agentType, tool, use, transcript string
	input                                              schema.Object
}

func (c hookCall) payload(dir string) string {
	o := schema.Object{{Key: "session_id", Value: c.session}, {Key: "transcript_path", Value: c.transcript},
		{Key: "cwd", Value: dir}, {Key: "permission_mode", Value: "bypassPermissions"}}
	if c.agentID != "" {
		o = append(o, schema.Member{Key: "agent_id", Value: c.agentID})
	}
	if c.agentType != "" || c.agentID != "" {
		o = append(o, schema.Member{Key: "agent_type", Value: c.agentType})
	}
	use := c.use
	switch {
	case use == "" && (hookin.SpawnTool(c.tool) || c.tool == "Skill"):
		// A call that starts an agent: a use of its own, so each agent it starts is joined to its own call.
		use = fmt.Sprintf("toolu_01CALL%06d", atomic.AddInt64(&callUses, 1))
	case use == "":
		use = "toolu_01CALL"
	}
	o = append(o, schema.Member{Key: "hook_event_name", Value: "PreToolUse"}, schema.Member{Key: "tool_name", Value: c.tool},
		schema.Member{Key: "tool_input", Value: c.input}, schema.Member{Key: "tool_use_id", Value: use})
	b, err := schema.EncodeLine(o)
	if err != nil {
		panic(err)
	}
	return string(b)
}

// callUses numbers the uses hookCall gives the calls that start an agent.
var callUses int64

// started runs the person's session's Agent call for an agent of type typ, so the next agent of that type to make
// its first call is joined to a call the guard allowed for its role, and plays it (VG's item A: a declared role's
// rights go only to an agent joined to such a call).
func started(t *testing.T, dir, session, typ string) {
	t.Helper()
	wantAnswer(t, dir, hookCall{session: session, tool: "Agent", input: in("subagent_type", typ, "prompt", "go")}.payload(dir), 0,
		RuleCallAllowed)
}

func in(kv ...string) schema.Object {
	var o schema.Object
	for i := 0; i+1 < len(kv); i += 2 {
		o = append(o, schema.Member{Key: kv[i], Value: kv[i+1]})
	}
	return o
}

const (
	sessA = "393fe3e7-4666-4e2a-80e7-62cfe58f6a23"
	sessB = "4e3cffd5-bd72-4f4a-a089-220fd2fdbff5"
)

// runAt runs the guard on a payload at a time, as the first line or the second, and gives its exit code and stderr.
func runAt(t *testing.T, dir, payload string, now time.Time, second bool, kv ...string) (int, string) {
	t.Helper()
	var stderr bytes.Buffer
	code := Main(Options{Stdin: strings.NewReader(payload), Stderr: &stderr, Getenv: env(dir, kv...), Budget: untimed,
		Now: func() time.Time { return now }, Second: second})
	return code, stderr.String()
}

// runs is runAt at testNow, on the first line.
func runs(t *testing.T, dir, payload string, kv ...string) (int, string) {
	t.Helper()
	return runAt(t, dir, payload, testNow, false, kv...)
}

// wantAnswer checks a run's exit code and, for a refusal, its rule in the session's log's last guard record.
func wantAnswer(t *testing.T, dir, payload string, code int, rule string, kv ...string) string {
	t.Helper()
	got, stderr := runs(t, dir, payload, kv...)
	if got != code {
		t.Fatalf("exit %d, want %d (%s): %s\n%s", got, code, rule, stderr, payload)
	}
	if rule != "" {
		if r := lastRule(t, dir, payload); r != rule {
			t.Fatalf("rule %q, want %q: %s", r, rule, stderr)
		}
	}
	return stderr
}

// lastRule is the rule of the last guard record in the payload's session's log.
func lastRule(t *testing.T, dir, payload string) string {
	t.Helper()
	recs := sessionRecords(t, dir, payload)
	for i := len(recs) - 1; i >= 0; i-- {
		if ev, _ := recs[i].Get("event"); ev == "guard" {
			r, _ := recs[i].Get("rule")
			s, _ := r.(string)
			return s
		}
	}
	return ""
}

// sessionRecords reads the log records of a payload's session.
func sessionRecords(t *testing.T, dir, payload string) []schema.Object {
	t.Helper()
	var p struct {
		Session string `json:"session_id"`
	}
	if err := json.Unmarshal([]byte(payload), &p); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(LogDir), "s-"+p.Session+".ndjson"))
	if err != nil {
		t.Fatalf("the session's log: %v", err)
	}
	var out []schema.Object
	for _, line := range strings.Split(strings.TrimSuffix(string(raw), "\n"), "\n") {
		v, err := schema.Decode([]byte(line))
		if err != nil {
			t.Fatal(err)
		}
		out = append(out, v.(schema.Object))
	}
	return out
}

// stateOf reads an agent's state.
func stateOf(t *testing.T, dir, session, agent string) *agents.State {
	t.Helper()
	st, err := (agents.Store{Main: dir}).Read(session, agent)
	if err != nil {
		t.Fatalf("the state of %s: %v", agent, err)
	}
	return st
}

// bash is a Bash call's input.
func bash(command string) schema.Object { return in("command", command, "description", "a test") }

// The measured routes: who the caller is, and who it may call. Each fixture is run as Claude Code sent it.
func TestTheMeasuredRoutes(t *testing.T) {
	dir := rolesProject(t, nil)
	tr := t.TempDir()
	fx := func(rel string) string { return fixture(t, rel, dir, tr) }

	// The main session (no agent_id, no agent_type): the person's own, never counted; it may call a helper (c2) and
	// a lead (c3's first call).
	wantAnswer(t, dir, fx("c1-main/02-PreToolUse-Bash.json"), 0, RuleShell)
	wantAnswer(t, dir, fx("c2-subagent/02-PreToolUse-Agent.json"), 0, RuleCallAllowed)
	// The helper it started: its first call is counted, joined to that Agent call, its brief named nowhere sound.
	wantAnswer(t, dir, fx("c2-subagent/04-PreToolUse-Bash.json"), 0, RuleShell)
	st := stateOf(t, dir, sessA, "ab8125621420b9524")
	if st.Header.Role != "helper" || st.Header.Kind != "helper" || st.Header.Writes || *st.Header.Limits.Actions.Value != 150 ||
		st.Header.ReadFrom != "HEAD" {
		t.Errorf("the helper's header %+v", st.Header)
	}
	if l := st.Link(); l == nil || l.Parent != "session" || l.Call != "toolu_016Z7tzdB6g9U9Z1bWHtsAah" {
		t.Errorf("the helper's link %+v", l)
	}
	if n, _ := (agents.Store{Main: dir}).Calls(sessA, "ab8125621420b9524"); n != 1 {
		t.Errorf("%d calls counted", n)
	}
	var spawn schema.Object
	for _, r := range sessionRecords(t, dir, fx("c2-subagent/04-PreToolUse-Bash.json")) {
		if ev, _ := r.Get("event"); ev == "spawn" {
			spawn = r
		}
	}
	if spawn == nil || spawn.String("parent") != "session" || spawn.String("subagent_id") != "ab8125621420b9524" ||
		spawn.String("subagent_type") != "helper" || spawn.String("tool_use_id") != "toolu_016Z7tzdB6g9U9Z1bWHtsAah" ||
		spawn.String("agent_event") != "PreToolUse" || spawn.String("category") != "Agent" {
		t.Errorf("the spawn record %v", spawn)
	}
	// Every record of the session, the spawn record among them, holds to log.schema.json and its field order.
	sch, err := schema.Parse(mustSchema(t))
	if err != nil {
		t.Fatal(err)
	}
	for i, r := range sessionRecords(t, dir, fx("c2-subagent/04-PreToolUse-Bash.json")) {
		if msgs := schema.Validate(sch, r); len(msgs) != 0 {
			t.Errorf("record %d against log.schema.json: %v", i, msgs)
		}
		if msgs := schema.CheckOrder(sch, r); len(msgs) != 0 {
			t.Errorf("record %d field order: %v", i, msgs)
		}
	}

	// c3: the session starts a lead; the lead starts a helper (allowed: a lead calls a helper); the helper's call.
	wantAnswer(t, dir, fx("c3-nested/02-PreToolUse-Agent.json"), 0, RuleCallAllowed)
	wantAnswer(t, dir, fx("c3-nested/04-PreToolUse-Agent.json"), 0, RuleCallAllowed)
	wantAnswer(t, dir, fx("c3-nested/06-PreToolUse-Bash.json"), 0, RuleShell)

	// c4: an --agent lead session (agent_type lead, no agent_id): a named session of the lead role, counted by its
	// session id.
	wantAnswer(t, dir, fx("c4-agent/02-PreToolUse-Bash.json"), 0, RuleShell)
	var c4 struct {
		Session string `json:"session_id"`
	}
	_ = json.Unmarshal([]byte(fx("c4-agent/02-PreToolUse-Bash.json")), &c4)
	if st := stateOf(t, dir, c4.Session, c4.Session); !st.Header.Named || st.Header.Role != "lead" {
		t.Errorf("the --agent session's header %+v", st.Header)
	}
	// c3b: that session starts a helper (allowed: a lead calls a helper).
	wantAnswer(t, dir, fx("c3b-agent-sub/02-PreToolUse-Agent.json"), 0, RuleCallAllowed)
	wantAnswer(t, dir, fx("c3b-agent-sub/04-PreToolUse-Bash.json"), 0, RuleShell)

	// c5: a fork (no subagent_type): Claude Code's general-purpose agent, a type no pack declares, held as a helper.
	wantAnswer(t, dir, fx("c5-fork/02-PreToolUse-Agent.json"), 0, RuleCallAllowed)
	wantAnswer(t, dir, fx("c5-fork/04-PreToolUse-Bash.json"), 0, RuleShell)

	// c8: top-level sessions started from an agent's shell. With no --agent and no BONSAI_ROLE it reads as the person's
	// own (a known limit: the shell judge refuses the start itself on an agent's line); with BONSAI_ROLE it plays that
	// role; with --agent its type.
	wantAnswer(t, dir, fx("c8-plain/02-PreToolUse-Bash.json"), 0, RuleShell)
	wantAnswer(t, dir, fx("c8-role/02-PreToolUse-Bash.json"), 0, RuleShell, RoleEnv, "helper")
	wantAnswer(t, dir, fx("c8-agent/02-PreToolUse-Bash.json"), 0, RuleShell)
	wantAnswer(t, dir, fx("c8-agentrole/02-PreToolUse-Bash.json"), 0, RuleShell, RoleEnv, "lead")
	for _, c := range []struct{ rel, role string }{{"c8-role/02-PreToolUse-Bash.json", "helper"},
		{"c8-agent/02-PreToolUse-Bash.json", "helper"}, {"c8-agentrole/02-PreToolUse-Bash.json", "helper"}} {
		var p struct {
			Session string `json:"session_id"`
		}
		_ = json.Unmarshal([]byte(fx(c.rel)), &p)
		if st := stateOf(t, dir, p.Session, p.Session); st.Header.Role != c.role || !st.Header.Named {
			t.Errorf("%s: %+v (the payload's type wins over BONSAI_ROLE)", c.rel, st.Header)
		}
	}

	// c10: a Workflow call from an --agent lead session: refused (only the person's session runs a Workflow).
	s := wantAnswer(t, dir, fx("c10-workflow/02-PreToolUse-Workflow.json"), 2, RuleCallRefused)
	if !strings.Contains(s, "only the person's session runs a Workflow") || !strings.Contains(s, "next: hand the work back") {
		t.Errorf("%s", s)
	}
	// c10b: a Workflow from the person's session is allowed, and its agent (workflow-subagent) runs as an undeclared
	// type, joined to that call, parent session.
	wf := strings.Replace(fx("c10b-workflow-bash/02-PreToolUse-Workflow.json"), `"agent_type": "lead",`, "", 1)
	wantAnswer(t, dir, wf, 0, RuleCallAllowed)
	wantAnswer(t, dir, fx("c10b-workflow-bash/06-PreToolUse-Bash.json"), 0, RuleShell)
	if st := stateOf(t, dir, "a7a8f330-78c4-4398-a265-2230b7719405", "a71418c4916da1718"); st.Header.Role != "" ||
		st.Header.Writes || st.Link() == nil || st.Link().Parent != "session" {
		t.Errorf("the Workflow's agent: %+v %+v", st.Header, st.Link())
	}

	// c6: SendMessage from the session to its helper's id resumes it (allowed); the resumed helper's call is counted on.
	wantAnswer(t, dir, fx("c6-sendmsg/02-PreToolUse-Agent.json"), 0, RuleCallAllowed)
	wantAnswer(t, dir, fx("c6-sendmsg/05-PreToolUse-ToolSearch.json"), 0, RuleOtherTool)
	wantAnswer(t, dir, fx("c6-sendmsg/11-PreToolUse-Bash.json"), 0, RuleShell)
	wantAnswer(t, dir, fx("c6-sendmsg/08-PreToolUse-SendMessage.json"), 0, RuleCallAllowed)
	wantAnswer(t, dir, fx("c6-sendmsg/13-PreToolUse-SendMessage.json"), 0, RuleCallAllowed)
	// c6b: the helper messages main: refused (a helper calls no one, and only the person's session writes to main).
	wantAnswer(t, dir, fx("c6b-tomain/02-PreToolUse-Agent.json"), 0, RuleCallAllowed)
	s = wantAnswer(t, dir, fx("c6b-tomain/04-PreToolUse-SendMessage.json"), 2, RuleCallRefused)
	if !strings.Contains(s, "may not call main") {
		t.Errorf("%s", s)
	}

	// c6b's second: the helper messages another session's id: refused as well.
	wantAnswer(t, dir, fx("c6b-tomain/06-PreToolUse-SendMessage.json"), 2, RuleCallRefused)

	// c7: a background Agent call, as any.
	wantAnswer(t, dir, fx("c7-bg/02-PreToolUse-Agent.json"), 0, RuleCallAllowed)
	wantAnswer(t, dir, fx("c7-bg/06-PreToolUse-Bash.json"), 0, RuleShell)

	// c12: the call a measured hook refused (exit 2): the guard allows the session's call to a helper, and refuses the
	// same call from a helper, which Claude Code then shows the caller (its stderr), no agent starting.
	wantAnswer(t, dir, fx("c12-refuse/02-PreToolUse-Agent.json"), 0, RuleCallAllowed)
	fromHelper := strings.Replace(fx("c12-refuse/02-PreToolUse-Agent.json"), `"hook_event_name"`,
		`"agent_id": "a0000000000000001", "agent_type": "helper", "hook_event_name"`, 1)
	wantAnswer(t, dir, fromHelper, 2, RuleCallRefused)

	// c9: Claude Code's own internal agents (a prompt suggestion, /btw) fire SubagentStop alone, with an empty type, and
	// make no tool call, so the guard never sees them and keeps no state for them. One that did call a tool would run
	// as an undeclared type: its read let through, never refused for its type.
	raw, err := os.ReadFile(filepath.Join("testdata", "claude-2.1.300", "c9-interactive", "05-SubagentStop.json"))
	if err != nil {
		t.Fatal(err)
	}
	var stop struct {
		Session string `json:"session_id"`
		Agent   string `json:"agent_id"`
		Type    string `json:"agent_type"`
	}
	if err := json.Unmarshal(raw, &stop); err != nil || stop.Type != "" || stop.Agent == "" {
		t.Fatalf("the internal agent's SubagentStop: %v %+v", err, stop)
	}
	if _, err := (agents.Store{Main: dir}).Read(stop.Session, stop.Agent); err != agents.ErrNoState {
		t.Errorf("state for an internal agent: %v", err)
	}
	internal := hookCall{session: stop.Session, agentID: stop.Agent, tool: "Read", input: in("file_path", filepath.Join(dir, "free.txt"))}.payload(dir)
	wantAnswer(t, dir, internal, 0, RuleOtherTool)
	if st := stateOf(t, dir, stop.Session, stop.Agent); st.Header.Role != "" || st.Header.Type != "" {
		t.Errorf("an internal agent's header %+v", st.Header)
	}
}

// Every helper refusing its calls, a lead refusing a lead, no one calling the session's role, a lead calling an
// undeclared type (allowed, held as a helper: its Edit refused), the session calling an undeclared type the same.
func TestWhoMayCallWhom(t *testing.T) {
	dir := rolesProject(t, nil)
	session := func(tool string, input schema.Object) string {
		return hookCall{session: sessA, tool: tool, input: input}.payload(dir)
	}
	agent := func(id, typ, tool string, input schema.Object) string {
		return hookCall{session: sessA, agentID: id, agentType: typ, tool: tool, input: input}.payload(dir)
	}
	cases := []struct {
		name    string
		payload string
		code    int
		rule    string
	}{
		{"the session calls a lead", session("Agent", in("subagent_type", "workflow:lead", "prompt", "go")), 0, RuleCallAllowed},
		{"the session calls a helper by its bare name", session("Task", in("subagent_type", "helper", "prompt", "go")), 0, RuleCallAllowed},
		{"the session calls Explore", session("Agent", in("subagent_type", "Explore", "prompt", "go")), 0, RuleCallAllowed},
		{"the session calls its own role", session("Agent", in("subagent_type", "workflow:orchestrator", "prompt", "go")), 2, RuleCallRefused},
		{"the session runs a Workflow", session("Workflow", in("script", "await agent('x')")), 0, RuleCallAllowed},
		{"the session writes to main", session("SendMessage", in("to", "main", "message", "hi")), 0, RuleCallAllowed},
		{"the session writes to another session", session("SendMessage", in("to", sessB, "message", "hi")), 0, RuleCallAllowed},
		{"a lead calls a helper", agent("a1", "workflow:lead", "Agent", in("subagent_type", "workflow:helper", "prompt", "go")), 0, RuleCallAllowed},
		{"a lead calls an undeclared type", agent("a1", "workflow:lead", "Agent", in("subagent_type", "Explore", "prompt", "go")), 0, RuleCallAllowed},
		{"a lead calls a fork", agent("a1", "workflow:lead", "Agent", in("prompt", "go")), 0, RuleCallAllowed},
		{"a lead calls a lead", agent("a1", "workflow:lead", "Agent", in("subagent_type", "workflow:verifier", "prompt", "go")), 2, RuleCallRefused},
		{"a lead calls the session's role", agent("a1", "workflow:lead", "Task", in("subagent_type", "orchestrator", "prompt", "go")), 2, RuleCallRefused},
		{"a lead runs a Workflow", agent("a1", "workflow:lead", "Workflow", in("script", "x")), 2, RuleCallRefused},
		{"a lead writes to main", agent("a1", "workflow:lead", "SendMessage", in("to", "main", "message", "x")), 2, RuleCallRefused},
		{"a lead writes to a name never started", agent("a1", "workflow:lead", "SendMessage", in("to", "worker9", "message", "x")), 2, RuleCallRefused},
		{"a helper calls a helper", agent("h1", "workflow:helper", "Agent", in("subagent_type", "workflow:helper", "prompt", "go")), 2, RuleCallRefused},
		{"a helper calls Explore", agent("h1", "workflow:helper", "Agent", in("subagent_type", "Explore", "prompt", "go")), 2, RuleCallRefused},
		{"Explore calls Explore", agent("e1", "Explore", "Agent", in("subagent_type", "Explore", "prompt", "go")), 2, RuleCallRefused},
		{"a subagent naming the session's role calls a helper", agent("s1", "workflow:orchestrator", "Agent", in("subagent_type", "helper", "prompt", "go")), 2, RuleCallRefused},
		{"Explore's Edit", agent("e1", "Explore", "Edit", in("file_path", filepath.Join(dir, "free.txt"), "old_string", "a", "new_string", "b")), 2, RuleRoleWritesNothing},
		{"Explore's Read", agent("e1", "Explore", "Read", in("file_path", filepath.Join(dir, "free.txt"))), 0, RuleOtherTool},
		{"a lead's Edit of a free file", agent("a1", "workflow:lead", "Edit", in("file_path", filepath.Join(dir, "free.txt"), "old_string", "a", "new_string", "b")), 0, RuleNotProtected},
		{"a verifier's Write", agent("v1", "workflow:verifier", "Write", in("file_path", filepath.Join(dir, "new.txt"), "content", "x")), 2, RuleRoleWritesNothing},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) { wantAnswer(t, dir, c.payload, c.code, c.rule) })
	}
	// A lead resumes the helper it started (by its id, once the helper has made a call); a helper of another's too.
	wantAnswer(t, dir, agent("a1", "workflow:lead", "Agent", in("subagent_type", "workflow:helper", "prompt", "go")), 0, RuleCallAllowed)
	wantAnswer(t, dir, agent("h7", "workflow:helper", "Read", in("file_path", filepath.Join(dir, "free.txt"))), 0, RuleOtherTool)
	wantAnswer(t, dir, agent("a1", "workflow:lead", "SendMessage", in("to", "h7", "message", "go on")), 0, RuleCallAllowed)
	wantAnswer(t, dir, agent("a1", "workflow:lead", "SendMessage", in("to", "a1", "message", "me")), 2, RuleCallRefused)
	// A named Agent call (a later Claude Code's name): a SendMessage to that name resolves to the role it started.
	wantAnswer(t, dir, session("Agent", in("subagent_type", "workflow:lead", "prompt", "go", "name", "lead-1")), 0, RuleCallAllowed)
	wantAnswer(t, dir, agent("a1", "workflow:lead", "SendMessage", in("to", "lead-1", "message", "x")), 2, RuleCallRefused)
	wantAnswer(t, dir, session("SendMessage", in("to", "lead-1", "message", "x")), 0, RuleCallAllowed)
}

// Each new row's test in TestEveryRule (rules_test.go's ruleTests); the tests above hold every case in full.
func init() {
	ruleTests[RuleCallAllowed] = func(t *testing.T) {
		dir := rolesProject(t, nil)
		wantAnswer(t, dir, hookCall{session: sessA, tool: "Agent", input: in("subagent_type", "workflow:helper", "prompt", "go")}.payload(dir), 0, RuleCallAllowed)
		recs := sessionRecords(t, dir, hookCall{session: sessA}.payload(dir))
		if last := recs[len(recs)-1]; last.String("target") != "workflow:helper" || !strings.HasPrefix(last.String("tool_use_id"), "toolu_01CALL") {
			t.Errorf("the record does not name the type asked for and the call: %v", last)
		}
	}
	ruleTests[RuleCallRefused] = func(t *testing.T) {
		dir := rolesProject(t, nil)
		s := wantAnswer(t, dir, hookCall{session: sessA, agentID: "h1", agentType: "workflow:helper", tool: "Agent",
			input: in("subagent_type", "workflow:helper", "prompt", "go")}.payload(dir), 2, RuleCallRefused)
		if !strings.Contains(s, "a helper, and every type no pack declares, calls no one") || !strings.Contains(s, "next: ") {
			t.Errorf("%s", s)
		}
		wantAnswer(t, dir, shellAs(dir, "h1", "workflow:helper", "claude -p hi", false), 2, RuleCallRefused)
	}
	ruleTests[RuleRoleWritesNothing] = func(t *testing.T) {
		dir := rolesProject(t, nil)
		wantAnswer(t, dir, hookCall{session: sessA, agentID: "v1", agentType: "workflow:verifier", tool: "Write",
			input: in("file_path", filepath.Join(dir, "a.txt"), "content", "x")}.payload(dir), 2, RuleRoleWritesNothing)
		wantAnswer(t, dir, shellAs(dir, "v1", "workflow:verifier", "echo x > a.txt", false), 2, RuleRoleWritesNothing)
	}
	ruleTests[RuleBudgetSpent] = func(t *testing.T) {
		dir := rolesProject(t, nil)
		p := hookCall{session: sessA, agentID: "t1", agentType: "workflow:tiny", tool: "Glob", input: in("pattern", "*")}.payload(dir)
		for i := 0; i < 3; i++ {
			wantAnswer(t, dir, p, 0, RuleOtherTool)
		}
		wantAnswer(t, dir, p, 2, RuleBudgetSpent)
	}
	ruleTests[RuleBudgetExempt] = func(t *testing.T) {
		dir := rolesProject(t, nil)
		p := hookCall{session: sessA, agentID: "t1", agentType: "workflow:tiny", tool: "Glob", input: in("pattern", "*")}.payload(dir)
		for i := 0; i < 4; i++ {
			runs(t, dir, p)
		}
		wantAnswer(t, dir, hookCall{session: sessA, agentID: "t1", agentType: "workflow:tiny", tool: "SubagentHandback",
			input: in("message", "done")}.payload(dir), 0, RuleBudgetExempt)
	}
	ruleTests[RuleBriefNotTied] = func(t *testing.T) {
		dir := rolesProject(t, nil)
		wantAnswer(t, dir, hookCall{session: sessA, tool: "Agent", use: "toolu_01X", input: in("subagent_type", "workflow:lead",
			"prompt", "M-2026-10-15-aaaaaaaa")}.payload(dir), 0, RuleCallAllowed)
		p := hookCall{session: sessA, agentID: "l1", agentType: "workflow:lead", tool: "Glob", input: in("pattern", "*")}.payload(dir)
		wantAnswer(t, dir, p, 0, RuleOtherTool)
		found := false
		for _, r := range sessionRecords(t, dir, p) {
			found = found || r.String("rule") == RuleBriefNotTied && strings.Contains(r.String("text"), "no message file holds it") &&
				r.String("decision") == "allow"
		}
		if !found {
			t.Error("no brief-not-tied record")
		}
	}
	ruleTests[RuleStateUnreadable] = func(t *testing.T) {
		dir := rolesProject(t, nil)
		writeFile(t, filepath.Join(dir, ".bonsai", "local", "agents", sessA, "l1.state"), "not a header")
		p := hookCall{session: sessA, agentID: "l1", agentType: "workflow:lead", tool: "Glob", input: in("pattern", "*")}.payload(dir)
		wantAnswer(t, dir, p, 0, RuleOtherTool)
		found := false
		for _, r := range sessionRecords(t, dir, p) {
			found = found || r.String("rule") == RuleStateUnreadable
		}
		if !found {
			t.Error("no state-unreadable record")
		}
	}
	ruleTests[RuleRolesFromFiles] = func(t *testing.T) {
		tmp, _ := filepath.EvalSymlinks(t.TempDir())
		home := testpack.Isolate(t, tmp)
		dir := filepath.Join(tmp, "uncommitted")
		writeFile(t, filepath.Join(dir, "bonsai.yaml"), rolesYAML)
		writeFile(t, filepath.Join(dir, ".bonsai", "lock.json"), string(rolesLock(t, fixtureRoles)))
		testpack.Git(t, dir, "init", "-q")
		if _, err := workspace.RecordCheckout(home, dir, rolesWS, time.Now()); err != nil {
			t.Fatal(err)
		}
		p := hookCall{session: sessA, agentID: "h1", agentType: "workflow:helper", tool: "Write",
			input: in("file_path", filepath.Join(dir, "a.txt"), "content", "x")}.payload(dir)
		wantAnswer(t, dir, p, 2, RuleRoleWritesNothing)
		found := false
		for _, r := range sessionRecords(t, dir, p) {
			found = found || r.String("rule") == RuleRolesFromFiles
		}
		if st := stateOf(t, dir, sessA, "h1"); !found || st.Header.ReadFrom != "files" {
			t.Errorf("roles read from the files: record %v, header %+v", found, st.Header)
		}
	}
	ruleTests[RuleSecondLineFault] = func(t *testing.T) {
		dir := rolesProject(t, nil)
		bad := `{"session_id":"` + sessA + `","hook_event_name":"PreToolUse","tool_name":"Read","tool_input":{"file_path":"a","file_path":"b"}}`
		if code, s := runAt(t, dir, bad, testNow, true); code != 0 {
			t.Fatalf("the second line blocked on its own fault: %s", s)
		}
		// A payload that does not read names no session: its record is the day's.
		recs := dayRecords(t, dir)
		if last := recs[len(recs)-1]; last.String("rule") != RuleSecondLineFault || last.String("decision") != "allow" ||
			!strings.Contains(last.String("text"), "bad-input") {
			t.Errorf("the record %v", last)
		}
		if code, _ := runAt(t, dir, bad, testNow, false); code != 2 {
			t.Error("the first line let a payload it cannot read through")
		}
	}
}

// The first line's tools as the guard reads them are engine.GuardTools' names.
func TestFirstLineToolsAreTheEngines(t *testing.T) {
	for _, tool := range engine.GuardTools {
		if !firstLineTool(tool.Name) {
			t.Errorf("%s: the guard does not read it as a first line's tool", tool.Name)
		}
	}
	for _, tool := range []string{"Read", "Grep", "SubagentHandback", "mcp__x__y", ""} {
		if firstLineTool(tool) {
			t.Errorf("%s read as a first line's tool", tool)
		}
	}
}
