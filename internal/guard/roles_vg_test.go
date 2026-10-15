package guard

// VG's items I and K (15 Oct), each reproducer a test: a state that does not read (a FIFO among them, VG's item F) is
// moved aside and counted afresh, and nothing hangs (I); a lead resumes only its own helpers, and calls side by side
// write one spent line (K).

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

// F and I, VG's reproducers: a state that does not read (a header garbled, an empty file, a folder, a FIFO, a calls
// file that is a folder) is moved aside, kept and named, and the agent counted afresh from that call, so it runs into
// its budget again (tiny's 3 actions) and nothing hangs.
func TestIAStateThatDoesNotReadIsCountedAfresh(t *testing.T) {
	cases := []struct {
		name  string
		plant func(t *testing.T, state, calls string)
		unix  bool
	}{
		{"a garbled header", func(t *testing.T, state, _ string) {
			raw, _ := os.ReadFile(state)
			writeFile(t, state, "garbage"+string(raw))
		}, false},
		{"an empty state", func(t *testing.T, state, _ string) { writeFile(t, state, "") }, false},
		{"a folder", func(t *testing.T, state, _ string) {
			if err := os.Remove(state); err != nil {
				t.Fatal(err)
			}
			if err := os.Mkdir(state, 0o755); err != nil {
				t.Fatal(err)
			}
		}, false},
		{"a calls file that is a folder", func(t *testing.T, _, calls string) {
			if err := os.Remove(calls); err != nil {
				t.Fatal(err)
			}
			if err := os.Mkdir(calls, 0o755); err != nil {
				t.Fatal(err)
			}
		}, false},
		{"a FIFO", func(t *testing.T, state, _ string) {
			if err := os.Remove(state); err != nil {
				t.Fatal(err)
			}
			mkfifo(t, state)
		}, true},
		{"a FIFO for its calls", func(t *testing.T, _, calls string) {
			if err := os.Remove(calls); err != nil {
				t.Fatal(err)
			}
			mkfifo(t, calls)
		}, true},
	}
	for i, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if c.unix && runtime.GOOS == "windows" {
				t.Skip("Windows has no FIFO in its file system")
			}
			dir := briefProject(t)
			id := fmt.Sprint("i", i)
			wantAnswer(t, dir, start(dir, "toolu_01I"+id, "go"), 0, RuleCallAllowed)
			call := tinyFrom(dir, id, "Read", readOf(dir, "free.txt"))
			wantAnswer(t, dir, call, 0, RuleOtherTool)
			wantAnswer(t, dir, call, 0, RuleOtherTool)
			folder := filepath.Join(dir, ".bonsai", "local", "agents", sessA)
			c.plant(t, filepath.Join(folder, id+".state"), filepath.Join(folder, id+".calls"))
			done := make(chan struct{})
			go func() {
				defer close(done)
				for n := 1; n <= 3; n++ {
					if code, s := runs(t, dir, call); code != 0 {
						t.Errorf("call %d after the plant: %s", n, s)
					}
				}
				if code, s := runs(t, dir, call); code != 2 || !strings.Contains(s, "budget spent (actions: 4 of 3)") {
					t.Errorf("counted afresh, the fourth call: %d %s", code, s)
				}
			}()
			select {
			case <-done:
			case <-time.After(30 * time.Second):
				t.Fatal("the guard hung on the planted state")
			}
			aside, _ := filepath.Glob(filepath.Join(folder, id+"-aside-*.state"))
			if len(aside) != 1 {
				t.Fatalf("moved aside: %v", aside)
			}
			found := false
			for _, r := range sessionRecords(t, dir, call) {
				found = found || r.String("rule") == RuleStateUnreadable && strings.Contains(r.String("text"), "moved aside, kept, to "+
					".bonsai/local/agents/"+sessA+"/"+filepath.Base(aside[0]))
			}
			if !found {
				t.Error("no record naming where the state was moved")
			}
			if st := stateOf(t, dir, sessA, id); st.Link() == nil || st.Header.Role != "tiny" {
				t.Errorf("the fresh state: %+v %+v", st.Header, st.Link())
			}
		})
	}
}

// K, VG's items: a lead resumes (and so re-briefs) only an agent it started itself; calls side by side past a budget
// write the window's spent line once.
func TestKALeadResumesOnlyItsOwnAndOneSpentLine(t *testing.T) {
	dir := briefProject(t)
	started(t, dir, sessA, "workflow:lead")
	started(t, dir, sessA, "workflow:lead")
	l1 := func(tool string, kv ...string) string {
		return hookCall{session: sessA, agentID: "l1", agentType: "workflow:lead", tool: tool, input: in(kv...)}.payload(dir)
	}
	l2 := func(tool string, kv ...string) string {
		return hookCall{session: sessA, agentID: "l2", agentType: "workflow:lead", tool: tool, input: in(kv...)}.payload(dir)
	}
	wantAnswer(t, dir, l1("Agent", "subagent_type", "workflow:helper", "prompt", "go"), 0, RuleCallAllowed)
	wantAnswer(t, dir, hookCall{session: sessA, agentID: "h1", agentType: "workflow:helper", tool: "Read", input: readOf(dir, "free.txt")}.payload(dir),
		0, RuleOtherTool)
	wantAnswer(t, dir, l2("Read", "file_path", "x"), 0, RuleOtherTool)
	s := wantAnswer(t, dir, l2("SendMessage", "to", "h1", "message", "your newer brief is "+b2), 2, RuleCallRefused)
	if !strings.Contains(s, "a lead resumes only an agent it started itself") {
		t.Errorf("%s", s)
	}
	wantAnswer(t, dir, l1("SendMessage", "to", "h1", "message", "go on"), 0, RuleCallAllowed)
	wantAnswer(t, dir, hookCall{session: sessA, tool: "SendMessage", input: in("to", "h1", "message", "go on")}.payload(dir), 0, RuleCallAllowed)

	// Twelve calls side by side, each past tiny's 3 actions: one spent line in the state, one full refusal.
	wantAnswer(t, dir, start(dir, "toolu_01K", "Your brief: "+b1+"."), 0, RuleCallAllowed)
	call := tinyFrom(dir, "k1", "Read", readOf(dir, "free.txt"))
	for i := 0; i < 3; i++ {
		wantAnswer(t, dir, call, 0, RuleOtherTool)
	}
	var wg sync.WaitGroup
	var mu sync.Mutex
	full := 0
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, s := runs(t, dir, call); strings.Contains(s, "report what you have and what is left") {
				mu.Lock()
				full++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	raw, _ := os.ReadFile(filepath.Join(dir, ".bonsai", "local", "agents", sessA, "k1.state"))
	if n := strings.Count(string(raw), `"e":"spent"`); n != 1 || full != 1 {
		t.Errorf("%d spent lines and %d full refusals, want one each:\n%s", n, full, raw)
	}
}
