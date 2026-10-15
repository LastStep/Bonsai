package guard

// VG round 2's findings (15 Oct), each reproducer a test: M1, part 2, a call never passed uncounted (chmod.py: the
// count goes on where it was when a mode is taken away, and a call it cannot count is refused); S1, a fork joined only
// to its own caller's call, an open call expiring (a.py's case A4). M1, part 1 is permissions_test.go's; S2 is
// vgfixes_test.go's D; the can-wait type with spaces round it, backing_test.go's.

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/agents"
	"github.com/LastStep/Bonsai/internal/schema"
)

// modesBite reports a system where taking a file's modes away stops this process writing it: not as root.
func modesBite() bool { return os.Geteuid() != 0 }

// tinyBudget starts a tiny (b1: 3 actions) and makes its first call.
func tinyBudget(t *testing.T, dir, id string) {
	t.Helper()
	wantAnswer(t, dir, start(dir, "toolu_01M1"+id, "Your brief: "+b1+"."), 0, RuleCallAllowed)
	wantAnswer(t, dir, tinyFrom(dir, id, "Bash", bash("echo 1")), 0, RuleShell)
}

// M1, chmod.py: an agent's calls file made read-only (by a route the guard cannot read, here the test's own), or its
// state made unreadable, leaves its count where it was: the guard sets back its own access, counts the call, and the
// budget runs out at its fourth action as before, with no window afresh. (Windows: os.Chmod sets the read-only
// attribute, which stops a write as a mode does; it never stops a read, so the state's case there reads on unaided.)
func TestM1AChangedModeLeavesTheCountAsItWas(t *testing.T) {
	for _, target := range []string{"calls", "state"} {
		t.Run(target, func(t *testing.T) {
			dir := briefProject(t)
			tinyBudget(t, dir, "c1")
			f := filepath.Join(dir, ".bonsai", "local", "agents", sessA, "c1."+target)
			// The readable line is refused (and counted: refused calls count).
			wantAnswer(t, dir, tinyFrom(dir, "c1", "Bash", bash("chmod a-w .bonsai/local/agents/"+sessA+"/c1."+target)), 2, RuleBonsaiFiles)
			mode := os.FileMode(0o444)
			if target == "state" {
				mode = 0
			}
			if err := os.Chmod(f, mode); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = os.Chmod(f, 0o644) })
			wantAnswer(t, dir, tinyFrom(dir, "c1", "Bash", bash("echo 3")), 0, RuleShell)
			s := wantAnswer(t, dir, tinyFrom(dir, "c1", "Bash", bash("echo 4")), 2, RuleBudgetSpent)
			if !strings.Contains(s, "budget spent (actions: 4 of 3)") {
				t.Errorf("%s", s)
			}
			for i := 0; i < 3; i++ {
				wantAnswer(t, dir, tinyFrom(dir, "c1", "Bash", bash("echo after")), 2, RuleBudgetSpent)
			}
			if n, err := (agents.Store{Main: dir}).Calls(sessA, "c1"); err != nil || n != 7 {
				t.Errorf("calls counted %d %v, want 7", n, err)
			}
			if (agents.Store{Main: dir}).Asides(sessA, "c1") != 0 {
				t.Error("the state was moved aside: a window afresh")
			}
			if fi, err := os.Stat(f); err != nil || modesBite() && fi.Mode().Perm()&0o600 != 0o600 {
				t.Errorf("the owner's access was not set back: %v %v", fi.Mode(), err)
			}
			if !modesBite() || target == "state" && runtime.GOOS == "windows" {
				return // nothing was stopped, so nothing needed setting back
			}
			found := false
			for _, r := range sessionRecords(t, dir, tinyFrom(dir, "c1", "Read", in("file_path", "x"))) {
				found = found || r.String("rule") == RuleStateUnreadable && strings.Contains(r.String("text"), "set back its own access")
			}
			if !found {
				t.Error("no state-unreadable record of the access set back")
			}
		})
	}
	// The session's folder made unreachable: set back too, the count goes on.
	dir := briefProject(t)
	tinyBudget(t, dir, "c2")
	folder := filepath.Join(dir, ".bonsai", "local", "agents", sessA)
	if err := os.Chmod(folder, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(folder, 0o755) })
	for i := 2; i <= 3; i++ {
		wantAnswer(t, dir, tinyFrom(dir, "c2", "Bash", bash("echo x")), 0, RuleShell)
	}
	wantAnswer(t, dir, tinyFrom(dir, "c2", "Bash", bash("echo 4")), 2, RuleBudgetSpent)
}

// M1, part 2: a call whose line cannot be written, with the agent's access set back and still failing (an ACL, an
// immutable flag), is refused as a call past a budget is, never passed uncounted; its report, its brief's read and
// the hand-back still pass; the next call that can be counted is counted as before.
func TestM1ACallThatCannotBeCountedIsRefused(t *testing.T) {
	dir := briefProject(t)
	tinyBudget(t, dir, "f1")
	was := countCall
	t.Cleanup(func() { countCall = was })
	countCall = func(agents.Store, string, string, time.Time, int64) (int64, error) {
		return 0, errors.New("permission denied")
	}
	s := wantAnswer(t, dir, tinyFrom(dir, "f1", "Read", readOf(dir, "free.txt")), 2, RuleBudgetSpent)
	for _, want := range []string{"the guard cannot count this call: this call's line cannot be written (permission denied)",
		"never passes unjudged", "next: try the call once more", hookPlace(false) + " message --kind report --from tiny@T-0001"} {
		if !strings.Contains(s, want) {
			t.Errorf("the refusal has no %q:\n%s", want, s)
		}
	}
	wantAnswer(t, dir, tinyFrom(dir, "f1", "Edit", in("file_path", filepath.Join(dir, "free.txt"), "old_string", "f", "new_string", "g")), 2,
		RuleBudgetSpent)
	// The second line refuses it too: a budget it cannot count is held spent there as well.
	if code, s := runAt(t, dir, tinyFrom(dir, "f1", "Grep", in("pattern", "x")), testNow, true); code != 2 || !strings.Contains(s, "cannot count") {
		t.Errorf("the second line: %d %s", code, s)
	}
	report := hookPlace(false) + " message --kind report --from tiny@T-0001 --to orchestrator --thread " + b1 + " --status blocked --body 'x'"
	wantAnswer(t, dir, tinyFrom(dir, "f1", "Bash", bash(report)), 0, RuleBudgetExempt)
	wantAnswer(t, dir, tinyFrom(dir, "f1", "Read", readOf(dir, ".bonsai/local/messages/tiny-T-0001/"+b1+".md")), 0, RuleBudgetExempt)
	wantAnswer(t, dir, tinyFrom(dir, "f1", "SubagentHandback", in("message", "x")), 0, RuleBudgetExempt)
	countCall = was
	wantAnswer(t, dir, tinyFrom(dir, "f1", "Bash", bash("echo 2")), 0, RuleShell)
	// An agent with no address has its last reply for its report.
	countCall = func(agents.Store, string, string, time.Time, int64) (int64, error) { return 0, errors.New("disk full") }
	s = wantAnswer(t, dir, hookCall{session: sessA, agentID: "x1", agentType: "Explore", tool: "Read", input: readOf(dir, "free.txt")}.payload(dir),
		2, RuleBudgetSpent)
	if !strings.Contains(s, "you have no address") {
		t.Errorf("%s", s)
	}
}

// M1, part 2: an agent whose state cannot be made is refused (fail closed), as is one whose id no state can be kept
// under; one whose state is moved aside a second time has its window held spent (a repeated fault resets no budget),
// while the first is counted afresh (step 5.5's "starts a fresh count").
func TestM1AStateThatCannotBeKeptFailsClosed(t *testing.T) {
	dir := briefProject(t)
	// No session id a state can be kept under (its record goes to the day's file, not a session's).
	if code, s := runs(t, dir, hookCall{session: "", agentID: "n1", agentType: "Explore", tool: "Read", input: readOf(dir, "free.txt")}.payload(dir)); code != 2 ||
		!strings.Contains(s, "the guard cannot count this call: the call names no session or agent id") {
		t.Errorf("no session: %d %s", code, s)
	}
	// A folder planted where the session's folder goes: the state cannot be made.
	writeFile(t, filepath.Join(dir, ".bonsai", "local", "agents", sessB), "a file in the session folder's place\n")
	s := wantAnswer(t, dir, hookCall{session: sessB, agentID: "n2", agentType: "Explore", tool: "Read", input: readOf(dir, "free.txt")}.payload(dir),
		2, RuleBudgetSpent)
	if !strings.Contains(s, "the state of this agent cannot be written") {
		t.Errorf("%s", s)
	}
	// Moved aside once: counted afresh. A second time: held spent, its report still let through, a newer brief from its
	// caller opens a fresh window.
	tinyBudget(t, dir, "a1")
	state := filepath.Join(dir, ".bonsai", "local", "agents", sessA, "a1.state")
	at := func(sec int, payload string, code int, rule string) string {
		t.Helper()
		got, s := runAt(t, dir, payload, testNow.Add(time.Duration(sec)*time.Second), false)
		if got != code || lastRule(t, dir, payload) != rule {
			t.Fatalf("at %ds: exit %d, rule %s: %s", sec, got, lastRule(t, dir, payload), s)
		}
		return s
	}
	writeFile(t, state, "not a header\n")
	at(1, tinyFrom(dir, "a1", "Bash", bash("echo afresh")), 0, RuleShell)
	if st := stateOf(t, dir, sessA, "a1"); st.Spent() != nil || st.Window().Tie == nil || st.Window().Tie.Brief != b1 {
		t.Fatalf("after the first move aside: %+v", st.Window())
	}
	writeFile(t, state, "not a header\n")
	s = at(2, tinyFrom(dir, "a1", "Bash", bash("echo again")), 2, RuleBudgetSpent)
	for _, want := range []string{"budget held spent (this agent's state could not be read a second time", "report what you have",
		hookPlace(false) + " message --kind report --from tiny@T-0001 --to orchestrator --thread " + b1} {
		if !strings.Contains(s, want) {
			t.Errorf("the refusal has no %q:\n%s", want, s)
		}
	}
	if s := at(3, tinyFrom(dir, "a1", "Bash", bash("echo again")), 2, RuleBudgetSpent); !strings.Contains(s, "budget held spent") ||
		!strings.Contains(s, "only your report") {
		t.Errorf("the later refusal: %s", s)
	}
	if n := (agents.Store{Main: dir}).Asides(sessA, "a1"); n != 2 {
		t.Errorf("%d states moved aside", n)
	}
	report := hookPlace(false) + " message --kind report --from tiny@T-0001 --to orchestrator --thread " + b1 + " --status blocked --body 'x'"
	at(4, tinyFrom(dir, "a1", "Bash", bash(report)), 0, RuleBudgetExempt)
	at(5, hookCall{session: sessA, tool: "SendMessage", input: in("to", "a1", "message", "Read "+b2)}.payload(dir), 0, RuleCallAllowed)
	at(6, tinyFrom(dir, "a1", "Bash", bash("echo rebriefed")), 0, RuleShell)
	// A third: held spent again.
	writeFile(t, state, "not a header\n")
	if s := at(7, tinyFrom(dir, "a1", "Bash", bash("echo x")), 2, RuleBudgetSpent); !strings.Contains(s, "budget held spent") {
		t.Errorf("%s", s)
	}
	// One that cannot be moved aside (its aside's name is taken): the call fails closed.
	stamp := testNow.Add(9 * time.Second).UTC().Format("20060102T150405.000Z")
	writeFile(t, filepath.Join(filepath.Dir(state), "a1-aside-"+stamp[:15]+stamp[16:]+".state"), "taken\n")
	writeFile(t, state, "not a header\n")
	if s := at(9, tinyFrom(dir, "a1", "Bash", bash("echo x")), 2, RuleBudgetSpent); !strings.Contains(s, "the guard cannot count this call") {
		t.Errorf("%s", s)
	}
}

// S1, a.py's case A4: the session ran a forked skill whose fork made no call (its Skill call stays open); a helper's
// fork of a skill the guard cannot read (its meta file names the helper as its parent) does not take that call, so it
// plays no role; the session's own fork (spawnDepth 1, no parent) does; a fork whose file names no parent at another
// depth joins none.
func TestS1AForkJoinsOnlyItsOwnCallersCall(t *testing.T) {
	dir := skillProject(t)
	tr := t.TempDir()
	transcript := filepath.ToSlash(filepath.Join(tr, sessA+".jsonl"))
	as := func(id, typ, tool string, input schema.Object) string {
		return hookCall{session: sessA, agentID: id, agentType: typ, tool: tool, transcript: transcript, input: input}.payload(dir)
	}
	edit := in("file_path", filepath.Join(dir, "free.txt"), "old_string", "free", "new_string", "stolen")
	started(t, dir, sessA, "helper")
	metaFile(t, tr, sessA, "a4helper000000001", "helper", "", "")
	wantAnswer(t, dir, as("a4helper000000001", "helper", "Read", in("file_path", "x")), 0, RuleOtherTool)
	// The session's forky, whose fork never calls.
	wantSecond(t, dir, hookCall{session: sessA, tool: "Skill", use: "toolu_01STALE", transcript: transcript, input: in("skill", "forky")}.payload(dir),
		0, RuleCallAllowed)
	// The helper runs a plugin's skill (not read): let through; its fork, a lead by type, takes nothing.
	wantSecond(t, dir, as("a4helper000000001", "helper", "Skill", in("skill", "workflow:forky")), 0, RuleOtherTool)
	metaFile(t, tr, sessA, "a4fork00000000001", "lead", "a4helper000000001", "")
	wantAnswer(t, dir, as("a4fork00000000001", "lead", "Edit", edit), 2, RuleRoleWritesNothing)
	if st := stateOf(t, dir, sessA, "a4fork00000000001"); !st.Header.Unbacked || st.Link() != nil {
		t.Errorf("the helper's fork: %+v %+v", st.Header, st.Link())
	}
	// A fork whose file names no parent, at depth 2: no caller to match, so none.
	writeFile(t, filepath.Join(tr, sessA, "subagents", "agent-a4fork00000000002.meta.json"), `{"agentType":"lead","spawnDepth":2}`)
	wantAnswer(t, dir, as("a4fork00000000002", "lead", "Edit", edit), 2, RuleRoleWritesNothing)
	// The session's own fork takes the session's call, and plays the lead.
	metaFile(t, tr, sessA, "a4fork00000000003", "lead", "", "")
	wantAnswer(t, dir, as("a4fork00000000003", "lead", "Edit", in("file_path", filepath.Join(dir, "free.txt"), "old_string", "free",
		"new_string", "forked")), 0, RuleNotProtected)
	if st := stateOf(t, dir, sessA, "a4fork00000000003"); st.Link() == nil || st.Link().Call != "toolu_01STALE" {
		t.Errorf("the session's fork: %+v", st.Link())
	}
	// A meta file that is there but does not read joins none.
	started(t, dir, sessA, "workflow:lead")
	writeFile(t, filepath.Join(tr, sessA, "subagents", "agent-a4fork00000000004.meta.json"), "not json")
	wantAnswer(t, dir, as("a4fork00000000004", "workflow:lead", "Edit", edit), 2, RuleRoleWritesNothing)
}

// S1: an agent with no meta file joins only a call made at most openCallLife before its first call; a stale one (its
// child never called) is taken by no one.
func TestS1OpenCallsExpire(t *testing.T) {
	dir := rolesProject(t, nil)
	t0 := time.Date(2026, 10, 15, 3, 0, 0, 0, time.UTC)
	edit := func(id string) string {
		return hookCall{session: sessA, agentID: id, agentType: "workflow:lead", tool: "Edit",
			input: in("file_path", filepath.Join(dir, "free.txt"), "old_string", "f", "new_string", "g")}.payload(dir)
	}
	startedAt(t, dir, sessA, "workflow:lead", t0)
	if code, s := runAt(t, dir, edit("e1"), t0.Add(openCallLife+time.Second), false); code != 2 || !strings.Contains(s, "changes nothing") {
		t.Errorf("a stale call taken: %d %s", code, s)
	}
	startedAt(t, dir, sessA, "workflow:lead", t0.Add(time.Hour))
	if code, s := runAt(t, dir, edit("e2"), t0.Add(time.Hour+openCallLife), false); code != 0 {
		t.Errorf("a fresh call not taken: %d %s", code, s)
	}
}

// M1, part 2: a state made read-only (red-run and climb lines could not be written, so no red run counted) is set back
// and counted on; where its lines still cannot be written, a climb is refused (it could not be noted) and the call
// after a climb whose result cannot be counted is refused, never passed.
func TestM1RedRunsStillCountWhenTheStateIsReadOnly(t *testing.T) {
	dir := briefProject(t)
	t0 := time.Date(2026, 10, 15, 3, 0, 0, 0, time.UTC)
	startedAt(t, dir, sessA, "workflow:fragile", t0)
	call := func(sec int, cmd string, code int) string {
		t.Helper()
		got, s := runAt(t, dir, hookCall{session: sessA, agentID: "rr1", agentType: "workflow:fragile", tool: "Bash", input: bash(cmd)}.payload(dir),
			t0.Add(time.Duration(sec)*time.Second), false)
		if got != code {
			t.Fatalf("at %ds, %s: exit %d: %s", sec, cmd, got, s)
		}
		return s
	}
	state := filepath.Join(dir, ".bonsai", "local", "agents", sessA, "rr1.state")
	call(1, "git status", 0)
	readOnly := func() {
		t.Helper()
		if err := os.Chmod(state, 0o444); err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() { _ = os.Chmod(state, 0o644) })
	readOnly()
	call(2, "bonsai ladder --task T-0001 || true", 0)
	ladderResult(t, dir, "T-0001", false, t0.Add(3*time.Second))
	readOnly()
	call(4, "git status", 0)
	call(5, "bonsai ladder --task T-0001; true", 0)
	ladderResult(t, dir, "T-0001", false, t0.Add(6*time.Second))
	readOnly()
	if s := call(7, "git status", 2); !strings.Contains(s, "budget spent (red runs: 2 of 1)") {
		t.Errorf("%s", s)
	}
	// Lines that cannot be written even then.
	was := stateAppend
	t.Cleanup(func() { stateAppend = was })
	stateAppend = func(agents.Store, string, string, agents.Event) error { return errors.New("read-only file system") }
	startedAt(t, dir, sessA, "workflow:fragile", t0)
	other := func(cmd string) (int, string) {
		return runAt(t, dir, hookCall{session: sessA, agentID: "rr2", agentType: "workflow:fragile", tool: "Bash", input: bash(cmd)}.payload(dir),
			t0.Add(10*time.Second), false)
	}
	if code, s := other("bonsai ladder --task T-0002"); code != 2 || !strings.Contains(s, "its climb of T-0002 cannot be noted") {
		t.Errorf("a climb that cannot be noted: %d %s", code, s)
	}
	stateAppend = was
	if code, s := other("bonsai ladder --task T-0002"); code != 0 {
		t.Fatalf("%d %s", code, s)
	}
	ladderResult(t, dir, "T-0002", false, t0.Add(11*time.Second))
	stateAppend = func(agents.Store, string, string, agents.Event) error { return errors.New("read-only file system") }
	if code, s := runAt(t, dir, hookCall{session: sessA, agentID: "rr2", agentType: "workflow:fragile", tool: "Read", input: readOf(dir, "free.txt")}.payload(dir),
		t0.Add(12*time.Second), false); code != 2 || !strings.Contains(s, "its red line, for the climb of T-0002, cannot be written") {
		t.Errorf("a red run that cannot be counted: %d %s", code, s)
	}
}
