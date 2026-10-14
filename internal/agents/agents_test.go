package agents

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func store(t *testing.T) Store {
	t.Helper()
	main := t.TempDir()
	if err := os.MkdirAll(filepath.Join(main, ".bonsai"), 0o755); err != nil {
		t.Fatal(err)
	}
	return Store{Main: main}
}

func n(v int64) *int64 { return &v }

var t0 = time.Date(2026, 10, 15, 3, 0, 0, 0, time.UTC)

const (
	sess  = "393fe3e7-4666-4e2a-80e7-62cfe58f6a23"
	agent = "ab8125621420b9524"
)

func header() Header {
	return Header{Agent: agent, Session: sess, Type: "workflow:runner", Role: "runner", Pack: "workflow", Kind: "helper",
		Limits: Limits{Minutes: Limit{n(150), n(150)}, Actions: Limit{n(250), n(250)}}, First: Stamp(t0), ReadFrom: "HEAD"}
}

// A header is made once, whoever of calls side by side gets there first; it reads back whole, under .bonsai/local/agents/.
func TestOneHeader(t *testing.T) {
	s := store(t)
	if _, err := s.Read(sess, agent); err != ErrNoState {
		t.Fatalf("no state yet: %v", err)
	}
	var wg sync.WaitGroup
	made := make(chan bool, 16)
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			h := header()
			h.Note = fmt.Sprint(i)
			m, err := s.Create(h)
			if err != nil {
				t.Error(err)
			}
			made <- m
		}(i)
	}
	wg.Wait()
	close(made)
	count := 0
	for m := range made {
		if m {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("%d headers made", count)
	}
	st, err := s.Read(sess, agent)
	if err != nil || st.Header.Role != "runner" || *st.Header.Limits.Actions.Value != 250 || st.Header.Limits.RedRuns.Value != nil {
		t.Fatalf("%v %+v", err, st)
	}
	if !s.Exists() {
		t.Error("the agents folder is not there")
	}
	if _, err := os.Stat(filepath.Join(s.Main, ".bonsai", "local", "agents", sess, agent+".state")); err != nil {
		t.Error(err)
	}
	if _, err := os.Stat(filepath.Join(s.Main, ".bonsai", ".gitignore")); err != nil {
		t.Errorf("no .bonsai/.gitignore restored: %v", err)
	}
	for _, bad := range []string{"", "../x", "a/b", "a.b", strings.Repeat("a", 129)} {
		h := header()
		h.Agent = bad
		if _, err := s.Create(h); err == nil {
			t.Errorf("an agent id %q made a state file", bad)
		}
		if _, err := s.Read(sess, bad); err == nil || err == ErrNoState {
			t.Errorf("an agent id %q was read: %v", bad, err)
		}
	}
}

// Calls made side by side lose no count: the count is the calls file's length.
func TestCallsSideBySide(t *testing.T) {
	s := store(t)
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for k := 0; k < 25; k++ {
				if _, err := s.Call(sess, agent, t0.Add(time.Duration(k)*time.Second), int64(k)); err != nil {
					t.Error(err)
				}
			}
		}()
	}
	wg.Wait()
	if c, err := s.Calls(sess, agent); err != nil || c != 200 {
		t.Fatalf("calls %d %v", c, err)
	}
	raw, _ := os.ReadFile(filepath.Join(s.SessionDir(sess), agent+".calls"))
	if len(raw) != 200*CallLen {
		t.Fatalf("%d bytes", len(raw))
	}
	for i, line := range strings.SplitAfter(string(raw), "\n")[:200] {
		if len(line) != CallLen || line[20] != ' ' {
			t.Fatalf("line %d %q", i, line)
		}
	}
	at, m, ok := s.LastCall(sess, agent)
	if !ok || at.Before(t0) || m < 0 || m > 24 {
		t.Errorf("last call %v %d %v", at, m, ok)
	}
	if string(CallLine(t0, 1234567)) != "2026-10-15T03:00:00Z 999999\n" || string(CallLine(t0, -3)) != "2026-10-15T03:00:00Z 000000\n" {
		t.Error("the minutes figure is not held to six digits")
	}
	if c, err := s.Calls(sess, "a2"); c != 0 || err != nil {
		t.Errorf("no calls file: %d %v", c, err)
	}
}

// The window, the red runs, spent and reported, the open climbs and the offers, as the events say.
func TestTheEvents(t *testing.T) {
	s := store(t)
	if _, err := s.Create(header()); err != nil {
		t.Fatal(err)
	}
	add := func(e Event) {
		t.Helper()
		if err := s.Append(sess, agent, e); err != nil {
			t.Fatal(err)
		}
	}
	read := func() *State {
		t.Helper()
		st, err := s.Read(sess, agent)
		if err != nil {
			t.Fatal(err)
		}
		return st
	}
	st := read()
	if w := st.Window(); !w.Start.Equal(t0) || w.Calls != 0 || w.Tie != nil || *w.Limits.Actions.Value != 250 {
		t.Errorf("the first window %+v", w)
	}
	add(Event{E: EventLink, At: Stamp(t0), Parent: "session", Call: "toolu_01A", Briefs: []string{"M-2026-10-15-0000000a"}})
	add(Event{E: EventClimb, At: Stamp(t0), Task: "T-0001"})
	add(Event{E: EventClimb, At: Stamp(t0), Task: "T-0002"})
	if o := read().OpenClimbs(); len(o) != 2 {
		t.Errorf("open climbs %+v", o)
	}
	add(Event{E: EventRed, At: Stamp(t0), Task: "T-0001", N: 1})
	add(Event{E: EventRed, At: Stamp(t0), Task: "T-0002", N: 2})
	st = read()
	if st.RedRuns() != 2 || len(st.OpenClimbs()) != 0 || st.Link().Call != "toolu_01A" {
		t.Errorf("red %d open %v link %+v", st.RedRuns(), st.OpenClimbs(), st.Link())
	}
	add(Event{E: EventSpent, At: Stamp(t0.Add(time.Hour)), Measure: "actions", Count: 251, Of: 250})
	add(Event{E: EventSpent, At: Stamp(t0.Add(2 * time.Hour)), Measure: "minutes", Count: 151, Of: 150})
	st = read()
	if sp := st.Spent(); sp == nil || sp.Measure != "actions" || st.Reported() {
		t.Errorf("spent %+v reported %v", sp, st.Reported())
	}
	add(Event{E: EventOffer, At: Stamp(t0), Brief: "M-2026-10-15-0000000b", From: "builder"})
	add(Event{E: EventReported, At: Stamp(t0)})
	st = read()
	if !st.Reported() || len(st.Offers()) != 1 {
		t.Errorf("reported %v offers %v", st.Reported(), st.Offers())
	}
	// A tie opens a fresh window: its limits, counted from its calls; spent, reported and offers start again; red runs
	// stay (the failures are the work's, not the window's).
	add(Event{E: EventTie, At: Stamp(t0.Add(3 * time.Hour)), Brief: "M-2026-10-15-0000000b", To: "runner@T-0001",
		Limits: &Limits{Actions: Limit{n(3), n(250)}}, Calls: 40})
	st = read()
	w := st.Window()
	if !w.Start.Equal(t0.Add(3*time.Hour)) || w.Calls != 40 || *w.Limits.Actions.Value != 3 || w.Tie.To != "runner@T-0001" ||
		st.Spent() != nil || st.Reported() || len(st.Offers()) != 0 || st.RedRuns() != 2 {
		t.Errorf("after a tie: %+v spent %v reported %v offers %v red %d", w, st.Spent(), st.Reported(), st.Offers(), st.RedRuns())
	}
	add(Event{E: EventOffer, At: StampMilli(t0), Brief: "M-2026-10-15-0000000c", From: "builder"})
	add(Event{E: EventOffer, At: StampMilli(t0), Brief: "M-2026-10-15-0000000d", From: "builder"})
	add(Event{E: EventTried, At: StampMilli(t0), Brief: "M-2026-10-15-0000000c"})
	if o := read().Offers(); len(o) != 1 || o[0].Brief != "M-2026-10-15-0000000d" {
		t.Errorf("offers after one tried: %+v", o)
	}
	if at, err := ParseStamp(StampMilli(t0.Add(1500 * time.Millisecond))); err != nil || !at.Equal(t0.Add(1500*time.Millisecond)) {
		t.Errorf("a stamp to the millisecond: %v %v", at, err)
	}
	if false {
		t.Errorf("after a tie: %+v spent %v reported %v offers %v red %d", w, st.Spent(), st.Reported(), st.Offers(), st.RedRuns())
	}
	add(Event{E: EventReset, At: Stamp(t0), Task: "T-0001"})
	if read().RedRuns() != 0 {
		t.Error("a green climb did not set the count back")
	}
	// A line that does not read is passed over.
	f, _ := os.OpenFile(filepath.Join(s.SessionDir(sess), agent+".state"), os.O_APPEND|os.O_WRONLY, 0o644)
	_, _ = f.WriteString("{not json\n")
	_ = f.Close()
	if st := read(); len(st.Events) != 14 {
		t.Errorf("%d events", len(st.Events))
	}
	if names, err := s.Agents(sess); err != nil || len(names) != 1 || names[0] != agent {
		t.Errorf("agents %v %v", names, err)
	}
}

// A call is claimed by one agent, a brief tied by one agent; asking again for the same agent is still its own.
func TestClaimsAndTies(t *testing.T) {
	s := store(t)
	var wg sync.WaitGroup
	wins := make(chan string, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			who := fmt.Sprintf("a%d", i)
			if ok, err := s.Claim(sess, "toolu_01A", who); err != nil {
				t.Error(err)
			} else if ok {
				wins <- who
			}
		}(i)
	}
	wg.Wait()
	close(wins)
	var won []string
	for w := range wins {
		won = append(won, w)
	}
	if len(won) != 1 {
		t.Fatalf("claimed by %v", won)
	}
	if ok, _ := s.Claim(sess, "toolu_01A", won[0]); !ok || !s.Claimed(sess, "toolu_01A") || s.Claimed(sess, "toolu_01B") {
		t.Error("the claim is not its holder's")
	}
	if ok, err := s.Tie("M-2026-10-15-0000000a", sess, agent); !ok || err != nil {
		t.Fatalf("tie %v %v", ok, err)
	}
	if ok, _ := s.Tie("M-2026-10-15-0000000a", sess, agent); !ok {
		t.Error("a tie asked again by its agent is not its own")
	}
	if ok, _ := s.Tie("M-2026-10-15-0000000a", sess, "another"); ok {
		t.Error("a brief tied twice")
	}
	if s.TiedBy("M-2026-10-15-0000000a") != sess+"/"+agent || s.TiedBy("M-2026-10-15-0000000b") != "" {
		t.Error("TiedBy")
	}
	if _, err := s.Tie("../x", sess, agent); err == nil {
		t.Error("a brief id that leaves the folder was tied")
	}
}

func TestSpawns(t *testing.T) {
	s := store(t)
	if sp, err := s.Spawns(sess); sp != nil || err != nil {
		t.Fatalf("%v %v", sp, err)
	}
	for i, typ := range []string{"workflow:runner", "Explore", "workflow:runner"} {
		if err := s.AddSpawn(sess, Spawn{At: Stamp(t0), Tool: "Agent", Use: fmt.Sprintf("toolu_0%d", i), Parent: "session",
			Type: typ, Briefs: []string{"M-2026-10-15-0000000a"}}); err != nil {
			t.Fatal(err)
		}
	}
	sp, err := s.Spawns(sess)
	if err != nil || len(sp) != 3 || sp[1].Type != "Explore" || sp[2].Use != "toolu_02" {
		t.Errorf("%v %+v", err, sp)
	}
	long := make([]string, 200)
	for i := range long {
		long[i] = fmt.Sprintf("M-2026-10-15-%08x", i)
	}
	if err := s.AddSpawn(sess, Spawn{Use: "toolu_09", Briefs: long}); err != nil {
		t.Errorf("a spawn naming many briefs: %v", err)
	}
	if sp, _ := s.Spawns(sess); len(sp) != 4 || len(sp[3].Briefs) != 4 {
		t.Errorf("%+v", sp[len(sp)-1])
	}
}
