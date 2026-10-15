package guard

// The stall stops (budget.go): the budget table (each measure crossed in turn; calls side by side losing no count; past
// it, the report in its exact form, its brief and the hand-back allowed and nothing else), the brief's ties (from the
// starting call's prompt, by its read, re-tied by a newer one; another address's, one with no message file, one from a
// role that may not call it and one another agent tied, each tying nothing), red climbs read from the climb's own
// result (behind || true, in the background), the second line, and a live lock edit changing no rule.

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/agents"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/testpack"
)

const (
	b1 = "M-2026-10-15-0000000a" // the session's brief to tiny@T-0001, actions 3
	b2 = "M-2026-10-15-0000000b" // a newer one in its thread, actions 5
	b3 = "M-2026-10-15-0000000c" // another role's
	b4 = "M-2026-10-15-0000000d" // from a helper, which may not call a helper
	b5 = "M-2026-10-15-0000000e" // to tiny, read by a second agent; its record in another session's log file
	b6 = "M-2026-10-15-0000000f" // to tiny, written by hand: no message record in the log
)

func briefProject(t *testing.T) string {
	t.Helper()
	m := func(folder, id string) string { return ".bonsai/local/messages/" + folder + "/" + id + ".md" }
	dir := rolesProject(t, map[string]string{
		m("tiny-T-0001", b1):   briefFile(b1, b1, "orchestrator", "tiny@T-0001", "2026-10-15T03:00:00.000Z", "  minutes: null\n  actions: 3\n  red_runs: null"),
		m("tiny-T-0001", b2):   briefFile(b2, b1, "orchestrator", "tiny@T-0001", "2026-10-15T04:00:00.000Z", "  minutes: null\n  actions: 5\n  red_runs: null"),
		m("helper-T-0002", b3): briefFile(b3, b3, "orchestrator", "helper@T-0002", "2026-10-15T03:00:00.000Z", ""),
		m("tiny-T-0003", b4):   briefFile(b4, b4, "helper@T-0001", "tiny@T-0003", "2026-10-15T03:00:00.000Z", ""),
		m("tiny-T-0004", b5):   briefFile(b5, b5, "orchestrator", "tiny@T-0004", "2026-10-15T03:00:00.000Z", "  actions: 50"),
		m("tiny-T-0005", b6):   briefFile(b6, b6, "orchestrator", "tiny@T-0005", "2026-10-15T03:00:00.000Z", ""),
		// bonsai message's records: in the day file (written outside a session), and b5's in another session's file.
		".bonsai/local/log/w-2026-10-15.ndjson": messageRecords(b1, b2, b3, b4),
		".bonsai/local/log/s-other.ndjson":      messageRecords(b5),
	})
	// The other session's file was last written after b5 was made (a file written before it cannot hold its record).
	later := time.Date(2026, 10, 15, 5, 0, 0, 0, time.UTC)
	if err := os.Chtimes(filepath.Join(dir, ".bonsai", "local", "log", "s-other.ndjson"), later, later); err != nil {
		t.Fatal(err)
	}
	return dir
}

// messageRecords are a log file's lines holding the message record bonsai message writes with each message: its
// event, its target the message's id (contract §8.2).
func messageRecords(ids ...string) string {
	var b strings.Builder
	for _, id := range ids {
		b.WriteString(`{"format":"bonsai.log/1","event":"message","target":"` + id + `","text":"brief to tiny"}` + "\n")
	}
	return b.String()
}

// tinyFrom is a call from a tiny helper.
func tinyFrom(dir, id, tool string, input schema.Object) string {
	return hookCall{session: sessA, agentID: id, agentType: "workflow:tiny", tool: tool, input: input}.payload(dir)
}

// start is the session's Agent call starting a tiny with a prompt.
func start(dir, use, prompt string) string {
	return hookCall{session: sessA, tool: "Agent", use: use, input: in("subagent_type", "workflow:tiny", "prompt", prompt)}.payload(dir)
}

func readOf(dir, rel string) schema.Object {
	return in("file_path", filepath.Join(dir, filepath.FromSlash(rel)))
}

func TestTheBudgetOfActions(t *testing.T) {
	dir := briefProject(t)
	wantAnswer(t, dir, start(dir, "toolu_01T1", "Your brief: "+b1+"."), 0, RuleCallAllowed)
	for i := 1; i <= 3; i++ {
		wantAnswer(t, dir, tinyFrom(dir, "t1", "Read", readOf(dir, "free.txt")), 0, RuleOtherTool)
	}
	st := stateOf(t, dir, sessA, "t1")
	if w := st.Window(); w.Tie == nil || w.Tie.Brief != b1 || *w.Limits.Actions.Value != 3 || *w.Limits.Actions.Max != 10 {
		t.Fatalf("the window %+v", w)
	}
	// The fourth: spent, its refusal in full; the fifth, short.
	s := wantAnswer(t, dir, tinyFrom(dir, "t1", "Read", readOf(dir, "free.txt")), 2, RuleBudgetSpent)
	// b2, a newer brief to its address in its thread, is already waiting: the refusal names it.
	waiting := "a newer brief to you is waiting: read it with /usr/local/bin/bonsai message --show " + b2
	for _, want := range []string{"budget spent (actions: 4 of 3): report what you have and what is left",
		"bonsai message --kind report --from tiny@T-0001 --to orchestrator --thread " + b1, "<<'EOF'", waiting} {
		if !strings.Contains(s, want) {
			t.Errorf("the first refusal has no %q:\n%s", want, s)
		}
	}
	s = wantAnswer(t, dir, tinyFrom(dir, "t1", "Grep", in("pattern", "x")), 2, RuleBudgetSpent)
	if !strings.Contains(s, "budget spent (actions: 4 of 3): only your report") || !strings.Contains(s, waiting) {
		t.Errorf("the later refusal: %s", s)
	}
	// Past it: every call refused but its report, its brief and the hand-back.
	report := "/usr/local/bin/bonsai message --kind report --from tiny@T-0001 --to orchestrator --thread " + b1 + " --status blocked --body 'what I have'"
	for _, p := range []string{
		tinyFrom(dir, "t1", "Bash", bash("git status")),
		tinyFrom(dir, "t1", "Bash", bash(report+"; echo done")),
		tinyFrom(dir, "t1", "Bash", bash(strings.Replace(report, "tiny@T-0001", "tiny@T-0002", 1))),
		tinyFrom(dir, "t1", "Read", readOf(dir, ".bonsai/local/messages/helper-T-0002/"+b3+".md")),
		tinyFrom(dir, "t1", "Bash", bash("bonsai message --show "+b3)),
		tinyFrom(dir, "t1", "Edit", in("file_path", filepath.Join(dir, "free.txt"), "old_string", "f", "new_string", "g")),
		tinyFrom(dir, "t1", "Agent", in("subagent_type", "Explore", "prompt", "x")),
	} {
		wantAnswer(t, dir, p, 2, RuleBudgetSpent)
	}
	wantAnswer(t, dir, tinyFrom(dir, "t1", "Read", readOf(dir, ".bonsai/local/messages/tiny-T-0001/"+b1+".md")), 0, RuleBudgetExempt)
	wantAnswer(t, dir, tinyFrom(dir, "t1", "Bash", bash("/usr/local/bin/bonsai message --show "+b1+" --json")), 0, RuleBudgetExempt)
	// A bare bonsai runs whatever the PATH finds first: past a budget it reads nothing (VG round 2's S2).
	wantAnswer(t, dir, tinyFrom(dir, "t1", "Bash", bash("bonsai message --show "+b1+" --json")), 2, RuleBudgetSpent)
	wantAnswer(t, dir, tinyFrom(dir, "t1", "SubagentHandback", in("message", "x")), 0, RuleBudgetExempt)
	if stateOf(t, dir, sessA, "t1").Reported() {
		t.Error("reported before the report")
	}
	wantAnswer(t, dir, tinyFrom(dir, "t1", "Bash", bash(report)), 0, RuleBudgetExempt)
	if st := stateOf(t, dir, sessA, "t1"); !st.Reported() || st.Spent() == nil {
		t.Errorf("after the report: spent %v reported %v", st.Spent(), st.Reported())
	}
	// Re-briefed by a newer brief named in the SendMessage that resumes it: a fresh window of 5, working again.
	wantAnswer(t, dir, hookCall{session: sessA, tool: "SendMessage", input: in("to", "t1", "message", "Read "+b2+" and go on")}.payload(dir), 0, RuleCallAllowed)
	for i := 1; i <= 5; i++ {
		wantAnswer(t, dir, tinyFrom(dir, "t1", "Read", readOf(dir, "free.txt")), 0, RuleOtherTool)
	}
	if s := wantAnswer(t, dir, tinyFrom(dir, "t1", "Read", readOf(dir, "free.txt")), 2, RuleBudgetSpent); strings.Contains(s, "waiting") {
		t.Errorf("no newer brief waits, yet the refusal names one: %s", s)
	}
	st = stateOf(t, dir, sessA, "t1")
	if w := st.Window(); w.Tie.Brief != b2 || *w.Limits.Actions.Value != 5 {
		t.Errorf("after the re-brief: %+v", w)
	}
	// Every call counted, refused ones among them.
	if n, _ := (agents.Store{Main: dir}).Calls(sessA, "t1"); n != 3+14+5+1 {
		t.Errorf("%d calls counted", n)
	}
}

// An agent with no address (no brief tied: an undeclared type) has no report call: every call is refused past it.
func TestNoAddressNoReport(t *testing.T) {
	dir := briefProject(t)
	for i := 1; i <= 3; i++ {
		wantAnswer(t, dir, hookCall{session: sessA, agentID: "x1", agentType: "Explore", tool: "Read", input: readOf(dir, "free.txt")}.payload(dir), 0, RuleOtherTool)
	}
	s := wantAnswer(t, dir, hookCall{session: sessA, agentID: "x1", agentType: "Explore", tool: "Bash",
		input: bash("bonsai message --kind report --from x --to orchestrator --status done --body 'x'")}.payload(dir), 2, RuleBudgetSpent)
	if !strings.Contains(s, "you have no address") || !strings.Contains(s, "your last reply is your report") {
		t.Errorf("%s", s)
	}
	wantAnswer(t, dir, hookCall{session: sessA, agentID: "x1", agentType: "Explore", tool: "SubagentHandback", input: in()}.payload(dir), 0, RuleBudgetExempt)
}

func TestTheBudgetOfMinutes(t *testing.T) {
	dir := briefProject(t)
	t0 := time.Date(2026, 10, 15, 3, 0, 0, 0, time.UTC)
	startedAt(t, dir, sessA, "workflow:lead", t0)
	lead := func(cmd string) string {
		return hookCall{session: sessA, agentID: "l1", agentType: "workflow:lead", tool: "Bash", input: bash(cmd)}.payload(dir)
	}
	for _, c := range []struct {
		at   time.Duration
		code int
	}{{0, 0}, {179 * time.Minute, 0}, {180*time.Minute + 59*time.Second, 0}, {181 * time.Minute, 2}} {
		code, s := runAt(t, dir, lead("git status"), t0.Add(c.at), false)
		if code != c.code {
			t.Fatalf("at %v: exit %d: %s", c.at, code, s)
		}
		if code == 2 && !strings.Contains(s, "budget spent (minutes: 181 of 180)") {
			t.Errorf("%s", s)
		}
	}
	// Each call's line carries its time and the minutes figure it was judged by.
	if at, m, ok := (agents.Store{Main: dir}).LastCall(sessA, "l1"); !ok || m != 181 || !at.Equal(t0.Add(181*time.Minute)) {
		t.Errorf("the last call line %v %d %v", at, m, ok)
	}
}

// Red runs are read from the climb's own result, written after the climb was noted, never from the shell: behind
// || true, and in the background, a red climb counts; a green one sets the count back.
func TestTheBudgetOfRedRuns(t *testing.T) {
	dir := briefProject(t)
	t0 := time.Date(2026, 10, 15, 3, 0, 0, 0, time.UTC)
	startedAt(t, dir, sessA, "workflow:fragile", t0)
	tiny := func(cmd string) string {
		return hookCall{session: sessA, agentID: "r1", agentType: "workflow:fragile", tool: "Bash", input: bash(cmd)}.payload(dir)
	}
	result := func(green bool, started time.Time) {
		raw := fmt.Sprintf(`{"format":"bonsai.ladder/1","task":"T-0001","workspace":"%s","mode":"local","started":"%s",`+
			`"finished":"%s","git":{"sha":"%s","branch":"main","dirty":false},"requested":[0],"green":%v,"rungs":[],`+
			`"skipped":[],"leftovers":null,"proof":null,"labels":{}}`, rolesWS, started.UTC().Format("2006-01-02T15:04:05.000Z"),
			started.UTC().Format("2006-01-02T15:04:05.000Z"), strings.Repeat("c", 40), green)
		writeFile(t, filepath.Join(dir, ".bonsai", "local", "ladder", "T-0001.json"), raw)
	}
	at := func(m int) time.Time { return t0.Add(time.Duration(m) * time.Second) }
	step := func(m int, cmd string, code int) string {
		t.Helper()
		got, s := runAt(t, dir, tiny(cmd), at(m), false)
		if got != code {
			t.Fatalf("at %ds, %s: exit %d: %s", m, cmd, got, s)
		}
		return s
	}
	// A red result from before the climb was noted counts for nothing.
	result(false, at(0))
	step(1, "bonsai ladder --task T-0001 --json || true", 0)
	step(2, "git status", 0)
	if st := stateOf(t, dir, sessA, "r1"); st.RedRuns() != 0 || len(st.OpenClimbs()) != 1 {
		t.Fatalf("an old result counted: %d %v", st.RedRuns(), st.OpenClimbs())
	}
	// Its own red result, written after the note: one red run (fragile's limit is 1: still within it).
	result(false, at(3))
	step(4, "git status", 0)
	if st := stateOf(t, dir, sessA, "r1"); st.RedRuns() != 1 || len(st.OpenClimbs()) != 0 {
		t.Fatalf("red runs %d open %v", st.RedRuns(), st.OpenClimbs())
	}
	// A green climb sets it back; a climb in the background is counted once its result appears.
	step(5, "bonsai ladder --task T-0001 &", 0)
	result(true, at(6))
	step(7, "git status", 0)
	if st := stateOf(t, dir, sessA, "r1"); st.RedRuns() != 0 {
		t.Fatalf("a green climb did not set the count back: %d", st.RedRuns())
	}
	// Two red climbs in a row: past the limit of 1 at the call after the second's result.
	step(8, "/usr/local/bin/bonsai ladder --task T-0001; true", 0)
	result(false, at(9))
	step(10, "bonsai ladder --task=T-0001 | tail -1", 0)
	result(false, at(11))
	s := step(12, "git status", 2)
	if !strings.Contains(s, "budget spent (red runs: 2 of 1)") {
		t.Errorf("%s", s)
	}
}

// A brief ties only if it is sound; each that is not ties nothing, with a record saying why.
func TestABriefTiesOnlyIfSound(t *testing.T) {
	dir := briefProject(t)
	notes := func(payload string) []string {
		var out []string
		for _, r := range sessionRecords(t, dir, payload) {
			if r.String("rule") == RuleBriefNotTied {
				out = append(out, r.String("text"))
			}
		}
		return out
	}
	// Named in the prompt: another address's brief (b3), one with no message file, one from a role that may not call
	// tiny (b4), one bonsai message did not write (b6, no message record): none ties, and each says why.
	wantAnswer(t, dir, start(dir, "toolu_01T2", "Briefs: "+b3+", M-2026-10-15-ffffffff, "+b4+" and "+b6+"."), 0, RuleCallAllowed)
	p := tinyFrom(dir, "t2", "Read", readOf(dir, "free.txt"))
	wantAnswer(t, dir, p, 0, RuleOtherTool)
	if st := stateOf(t, dir, sessA, "t2"); st.Window().Tie != nil {
		t.Errorf("an unsound brief tied: %+v", st.Window().Tie)
	}
	got := strings.Join(notes(p), "\n")
	for _, want := range []string{b3 + ", named by the call that started it, ties nothing: it is to \"helper@T-0002\", not to this agent's role \"tiny\"",
		"M-2026-10-15-ffffffff, named by the call that started it, ties nothing: no message file holds it",
		b4 + ", named by the call that started it, ties nothing: its from \"helper@T-0001\" is a role that may not call this agent's",
		b6 + ", named by the call that started it, ties nothing: the log holds no message record of it, so bonsai message did not write it",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("no note %q in:\n%s", want, got)
		}
	}
	// Tied by its read (bonsai message --show): b5, to tiny@T-0004; then a second agent's read of b5 ties nothing.
	wantAnswer(t, dir, tinyFrom(dir, "t2", "Bash", bash("bonsai message --show "+b5)), 0, RuleBonsaiCommand)
	if w := stateOf(t, dir, sessA, "t2").Window(); w.Tie == nil || w.Tie.Brief != b5 || *w.Limits.Actions.Value != 10 {
		t.Fatalf("the read did not tie b5 (its 50 actions held to the max of 10): %+v", w)
	}
	wantAnswer(t, dir, start(dir, "toolu_01T3", "go"), 0, RuleCallAllowed)
	p = tinyFrom(dir, "t3", "Read", readOf(dir, ".bonsai/local/messages/tiny-T-0004/"+b5+".md"))
	wantAnswer(t, dir, p, 0, RuleOtherTool)
	if !strings.Contains(strings.Join(notes(p), "\n"), b5+", named by this agent's read of it, ties nothing: another agent tied it") {
		t.Errorf("a second agent tied b5: %v", notes(p))
	}
	// A brief older than the tied one, or in another thread, does not re-tie (t2 holds b5, thread b5).
	wantAnswer(t, dir, tinyFrom(dir, "t2", "Bash", bash("bonsai message --show "+b1)), 0, RuleBonsaiCommand)
	if w := stateOf(t, dir, sessA, "t2").Window(); w.Tie.Brief != b5 {
		t.Errorf("another address's brief re-tied: %+v", w.Tie)
	}
}

// Calls side by side lose no count, and of an agent's first calls side by side one alone makes its state.
func TestCallsSideBySideLoseNoCount(t *testing.T) {
	dir := briefProject(t)
	started(t, dir, sessA, "workflow:lead")
	var wg sync.WaitGroup
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			runs(t, dir, hookCall{session: sessA, agentID: "p1", agentType: "workflow:lead", tool: "Read", input: readOf(dir, "free.txt")}.payload(dir))
		}()
	}
	wg.Wait()
	if n, _ := (agents.Store{Main: dir}).Calls(sessA, "p1"); n != 12 {
		t.Errorf("%d calls counted, want 12", n)
	}
	raw, _ := os.ReadFile(filepath.Join(dir, ".bonsai", "local", "agents", sessA, "p1.state"))
	if strings.Count(string(raw), "\n") != 2 || strings.Count(string(raw), `"e":"link"`) != 1 {
		t.Errorf("the state holds %d lines, want its header and its one link line:\n%s", strings.Count(string(raw), "\n"), raw)
	}
}

// Three Agent calls side by side: each child is joined to its own call by Claude Code's meta file, whatever order
// their first calls come in (measured: the children's first calls followed the calls' order in 11 of 12 runs, the
// meta file named the right call in 36 of 36); with no meta file, the oldest call of its type not yet joined.
func TestParallelSpawnsAreJoinedByTheMetaFile(t *testing.T) {
	dir := briefProject(t)
	tr := t.TempDir()
	transcript := filepath.Join(tr, sessA+".jsonl")
	uses := []string{"toolu_01PA", "toolu_01PB", "toolu_01PC"}
	for _, u := range uses {
		wantAnswer(t, dir, start(dir, u, "go"), 0, RuleCallAllowed)
	}
	meta := func(agent, use string) {
		writeFile(t, filepath.Join(tr, sessA, "subagents", "agent-"+agent+".meta.json"),
			`{"agentType":"workflow:tiny","toolUseId":"`+use+`","spawnDepth":1}`)
	}
	// The children's first calls in the order A, C, B (the measured mismatch, P1-3).
	meta("ca", uses[0])
	meta("cc", uses[2])
	meta("cb", uses[1])
	for _, a := range []string{"ca", "cc", "cb"} {
		p := hookCall{session: sessA, agentID: a, agentType: "workflow:tiny", tool: "Read", transcript: filepath.ToSlash(transcript),
			input: readOf(dir, "free.txt")}.payload(dir)
		wantAnswer(t, dir, p, 0, RuleOtherTool)
	}
	for a, u := range map[string]string{"ca": uses[0], "cb": uses[1], "cc": uses[2]} {
		if l := stateOf(t, dir, sessA, a).Link(); l == nil || l.Call != u || l.Parent != "session" {
			t.Errorf("%s joined to %+v, want %s", a, l, u)
		}
	}
	// No meta file: the oldest open call of its type.
	wantAnswer(t, dir, start(dir, "toolu_01PD", "go"), 0, RuleCallAllowed)
	wantAnswer(t, dir, start(dir, "toolu_01PE", "go"), 0, RuleCallAllowed)
	wantAnswer(t, dir, tinyFrom(dir, "cx", "Read", readOf(dir, "free.txt")), 0, RuleOtherTool)
	if l := stateOf(t, dir, sessA, "cx").Link(); l == nil || l.Call != "toolu_01PD" {
		t.Errorf("cx joined to %+v", l)
	}
}

// The second line: within the budget a read goes through, past it a read is refused; a tool the first line judges is
// passed at once, uncounted; a fault of its own lets the call through.
func TestTheSecondLine(t *testing.T) {
	dir := briefProject(t)
	read := tinyFrom(dir, "s1", "Read", readOf(dir, "free.txt"))
	for i := 0; i < 3; i++ {
		if code, s := runAt(t, dir, read, testNow, true); code != 0 {
			t.Fatalf("a read within the budget: %s", s)
		}
	}
	if code, s := runAt(t, dir, read, testNow, true); code != 2 || !strings.Contains(s, "budget spent (actions: 4 of 3)") {
		t.Fatalf("a read past the budget: %d %s", code, s)
	}
	n, _ := (agents.Store{Main: dir}).Calls(sessA, "s1")
	if code, _ := runAt(t, dir, tinyFrom(dir, "s1", "Bash", bash("rm -rf /")), testNow, true); code != 0 {
		t.Error("the second line judged a Bash call")
	}
	if m, _ := (agents.Store{Main: dir}).Calls(sessA, "s1"); m != n {
		t.Errorf("the second line counted a Bash call: %d, was %d", m, n)
	}
	// Faults: a payload that does not read, a bonsai.yaml that does not read, a state that does not read.
	for _, p := range []string{`{"hook_event_name":`, `{"hook_event_name":"PreToolUse","tool_name":"Read","tool_input":{"file_path":"a","file_path":"b"},"session_id":"` + sessA + `"}`} {
		if code, s := runAt(t, dir, p, testNow, true); code != 0 {
			t.Errorf("a fault blocked: %s", s)
		}
	}
	// An MCP tool's input the strict decoder would refuse is passed over, counted.
	mcp := `{"session_id":"` + sessA + `","cwd":"` + filepath.ToSlash(dir) + `","hook_event_name":"PreToolUse","agent_id":"m1",` +
		`"agent_type":"workflow:helper","tool_name":"mcp__x__y","tool_input":{"a":1,"a":2},"tool_use_id":"toolu_01M"}`
	if code, s := runAt(t, dir, mcp, testNow, true); code != 0 {
		t.Errorf("an MCP call: %s", s)
	}
	if n, _ := (agents.Store{Main: dir}).Calls(sessA, "m1"); n != 1 {
		t.Errorf("the MCP call was not counted: %d", n)
	}
	writeFile(t, filepath.Join(dir, "bonsai.yaml"), "not: [yaml")
	if code, s := runAt(t, dir, read, testNow, true); code != 0 {
		t.Errorf("a bonsai.yaml that does not read blocked the second line: %s", s)
	}
	if code, _ := runAt(t, dir, read, testNow, false); code != 2 {
		t.Error("a bonsai.yaml that does not read let the first line through")
	}
}

// A live edit of the lock (an agent's, through what the guard cannot read) changes no rule: the roles are read as
// committed, and an agent keeps the facts it started with.
func TestALiveLockEditChangesNoRule(t *testing.T) {
	dir := briefProject(t)
	edit := func(id string) string {
		return hookCall{session: sessA, agentID: id, agentType: "workflow:helper", tool: "Edit",
			input: in("file_path", filepath.Join(dir, "free.txt"), "old_string", "f", "new_string", "g")}.payload(dir)
	}
	wantAnswer(t, dir, edit("h1"), 2, RuleRoleWritesNothing)
	// helper made a writer in the lock as it stands, and every role taken out of it: no rule changes, for the agent
	// that started before and for one starting now.
	writeFile(t, filepath.Join(dir, ".bonsai", "lock.json"), string(rolesLock(t, strings.Replace(fixtureRoles,
		"name: helper\n    kind: helper\n    session: false\n    writes: false", "name: helper\n    kind: helper\n    session: false\n    writes: true", 1))))
	wantAnswer(t, dir, edit("h1"), 2, RuleRoleWritesNothing)
	wantAnswer(t, dir, edit("h2"), 2, RuleRoleWritesNothing)
	writeFile(t, filepath.Join(dir, ".bonsai", "lock.json"), string(rolesLock(t, "")))
	wantAnswer(t, dir, edit("h3"), 2, RuleRoleWritesNothing)
	// Committed, it counts for agents started after it.
	testpack.Git(t, dir, "commit", "-q", "-am", "roles out")
	wantAnswer(t, dir, edit("h4"), 0, RuleNotProtected)
	wantAnswer(t, dir, edit("h1"), 2, RuleRoleWritesNothing)
}

// A brief's message record is found in the session's own log file, the day file of its created, or another file
// written since it was made; a line longer than any record is passed over, and a file last written before the
// message was made is not read.
func TestMessageRecordFound(t *testing.T) {
	main := t.TempDir()
	logDir := filepath.Join(main, ".bonsai", "local", "log")
	created := time.Date(2026, 10, 15, 3, 0, 0, 0, time.UTC)
	b := &brief{id: "M-2026-10-15-00000001", created: created}
	if messageRecorded(main, sessA, b) {
		t.Fatal("found with no log at all")
	}
	long := strings.Repeat("x", 200<<10) + `"` + b.id + `"` + "\n"
	writeFile(t, filepath.Join(logDir, "s-"+sessA+".ndjson"), long+`{"event":"tool_start","target":"`+b.id+`"}`+"\n")
	if messageRecorded(main, sessA, b) {
		t.Error("found by a record of another event, or a line no record is")
	}
	writeFile(t, filepath.Join(logDir, "s-"+sessA+".ndjson"), long+messageRecords(b.id))
	if !messageRecorded(main, sessA, b) {
		t.Error("not found in the session's own file, after a line longer than any record")
	}
	// In another session's file: found only while that file was written to since the message was made.
	other := filepath.Join(logDir, "s-other.ndjson")
	writeFile(t, filepath.Join(logDir, "s-"+sessA+".ndjson"), "")
	writeFile(t, other, messageRecords(b.id))
	for _, c := range []struct {
		mod  time.Time
		want bool
	}{{created.Add(time.Hour), true}, {created.Add(-time.Hour), false}} {
		if err := os.Chtimes(other, c.mod, c.mod); err != nil {
			t.Fatal(err)
		}
		if got := messageRecorded(main, sessA, b); got != c.want {
			t.Errorf("another session's file written at %v: found %v, want %v", c.mod, got, c.want)
		}
	}
	// In the day file of its created (a message written outside a session).
	writeFile(t, filepath.Join(logDir, "w-2026-10-15.ndjson"), messageRecords(b.id))
	if !messageRecorded(main, "", b) {
		t.Error("not found in the day file")
	}
}
