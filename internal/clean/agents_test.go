package clean

// The agents kind (step 5.5.11, note 9): a session's folder of agents' state goes, every file of Bonsai's in it with
// its clean record, once its newest line is older than keep_days or it is beyond the newest keep_newest folders; never
// while its log file is kept or a brief one of its agents tied is kept; a brief's tie file goes once its brief is gone
// and the tie is older than keep_days, never while its brief is there; anything else there is left alone; null keeps;
// the session's end cleans it after the log.

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/agents"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// agentsSession writes one session's folder in l: an agent's state and two call lines, a spawn call and a claim, its newest
// line at at.
func agentsSession(t *testing.T, l workspace.Local, sess string, at time.Time) []string {
	t.Helper()
	s := agents.Store{Main: l.Main}
	if _, err := s.Create(agents.Header{Agent: "a1", Session: sess, Type: "Explore", Kind: "helper", First: agents.Stamp(at.Add(-time.Hour))}); err != nil {
		t.Fatal(err)
	}
	for _, c := range []time.Time{at.Add(-time.Hour), at} {
		if _, err := s.Call(sess, "a1", c, 0); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.AddSpawn(sess, agents.Spawn{At: agents.StampMilli(at.Add(-2 * time.Hour)), Tool: "Agent", Use: "toolu_01A", Parent: "session", Type: "Explore"}); err != nil {
		t.Fatal(err)
	}
	if ok, err := s.Claim(sess, "toolu_01A", "a1"); !ok || err != nil {
		t.Fatal(ok, err)
	}
	stamp(t, filepath.Join(s.SessionDir(sess), "claims", "toolu_01A"), at.Add(-2*time.Hour))
	pre := ".bonsai/local/agents/" + sess + "/"
	return []string{pre + "claims/toolu_01A", pre + "spawns.ndjson", pre + "a1.calls", pre + "a1.state"}
}

// leftIn gives the files there are under .bonsai/local/agents/, by path.
func leftIn(t *testing.T, l workspace.Local) []string {
	t.Helper()
	var out []string
	_ = filepath.Walk(filepath.Join(l.Main, ".bonsai", "local", "agents"), func(p string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			rel, _ := filepath.Rel(l.Main, p)
			out = append(out, filepath.ToSlash(rel))
		}
		return nil
	})
	sort.Strings(out)
	return out
}

func TestAgentsByAge(t *testing.T) {
	l := project(t, "")
	old := agentsSession(t, l, "s-old", daysAgo(31))
	logged := agentsSession(t, l, "s-logged", daysAgo(40))
	write(t, l.Main, ".bonsai/local/log/s-s-logged.ndjson", "{}\n")
	stamp(t, filepath.Join(l.Main, ".bonsai", "local", "log", "s-s-logged.ndjson"), daysAgo(1))
	recent := agentsSession(t, l, "s-recent", daysAgo(29))
	held := agentsSession(t, l, "s-held", daysAgo(40))
	stray := agentsSession(t, l, "s-stray", daysAgo(40))
	notes := write(t, l.Main, ".bonsai/local/agents/s-stray/notes.txt", "not Bonsai's\n")
	stamp(t, notes, daysAgo(40))
	// A thread whose brief s-held's agent tied, still there; and two ties whose briefs are gone.
	paths := thread(t, l, daysAgo(40), true)
	kept := strings.TrimSuffix(filepath.Base(paths[0]), ".md")
	s := agents.Store{Main: l.Main}
	tie := func(brief, sess string, at time.Time) string {
		t.Helper()
		if ok, err := s.Tie(brief, sess, "a1"); !ok || err != nil {
			t.Fatal(ok, err)
		}
		stamp(t, s.TiePath(brief), at)
		return ".bonsai/local/agents/ties/" + brief
	}
	keptTie := tie(kept, "s-held", daysAgo(40))
	goneOld := tie("M-2026-08-01-0000000a", "s-old", daysAgo(31))
	goneNew := tie("M-2026-09-20-0000000b", "s-recent", daysAgo(20))
	d := Files(opts(l), Agents)
	if len(d.Errs) != 0 || d.Stopped {
		t.Fatalf("errors %v, stopped %v", d.Errs, d.Stopped)
	}
	there := leftIn(t, l)
	for _, p := range append(append([]string{goneOld}, old...), stray...) {
		if has(there, p) {
			t.Errorf("%s is still there", p)
		}
	}
	stay := append([]string{keptTie, goneNew, ".bonsai/local/agents/s-stray/notes.txt"}, logged...)
	stay = append(append(stay, recent...), held...)
	for _, p := range stay {
		if !has(there, p) {
			t.Errorf("%s is gone: a folder whose log is kept, one 29 days old, one holding a kept brief's tie, a tie of a "+
				"kept brief, a tie 20 days old and a file not Bonsai's stay", p)
		}
	}
	if _, err := os.Stat(filepath.Join(l.Main, ".bonsai", "local", "agents", "s-old")); !os.IsNotExist(err) {
		t.Errorf("s-old's folder, emptied, is still there: %v", err)
	}
	// Each file gone has its clean record: the tie first, then each folder's files, oldest folder first.
	var got []string
	for _, c := range d.Cleaned {
		got = append(got, c.Target)
		if c.Kind != Agents || c.Reason != "generated.agents.keep_days=30" {
			t.Errorf("cleaned %+v", c)
		}
	}
	want := append(append([]string{goneOld}, stray...), old...)
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Errorf("cleaned:\n got %v\nwant %v", got, want)
	}
	if recs := cleanRecords(t, l); len(recs) != len(want) {
		t.Errorf("clean records: %v", recs)
	}
	// The log gone, s-logged goes at the next run; the brief's thread gone, s-held and its tie go too.
	if err := os.Remove(filepath.Join(l.Main, ".bonsai", "local", "log", "s-s-logged.ndjson")); err != nil {
		t.Fatal(err)
	}
	for _, p := range paths {
		if err := os.Remove(filepath.Join(l.Main, filepath.FromSlash(p))); err != nil {
			t.Fatal(err)
		}
	}
	d = Files(opts(l), Agents)
	there = leftIn(t, l)
	for _, p := range append(append([]string{keptTie}, logged...), held...) {
		if has(there, p) {
			t.Errorf("%s is still there at the second run (cleaned %+v)", p, d.Cleaned)
		}
	}
}

// keep_newest ranks the session folders by their newest line; a protected folder still counts among the newest.
func TestAgentsKeepNewest(t *testing.T) {
	l := project(t, "  agents:\n    keep_days: null\n    keep_newest: 2\n")
	a := agentsSession(t, l, "s-a", daysAgo(10))
	b := agentsSession(t, l, "s-b", daysAgo(8))
	write(t, l.Main, ".bonsai/local/log/s-s-b.ndjson", "{}\n")
	c := agentsSession(t, l, "s-c", daysAgo(6))
	dd := agentsSession(t, l, "s-d", daysAgo(4))
	d := Files(opts(l), Agents)
	there := leftIn(t, l)
	// The newest two are s-d and s-c; s-b, its log kept, is beyond them and stays; s-a goes.
	if has(there, a[0]) || has(there, a[3]) || !has(there, b[3]) || !has(there, c[3]) || !has(there, dd[3]) {
		t.Errorf("kept: %v (cleaned %+v)", there, d.Cleaned)
	}
	if len(d.Cleaned) != 4 || d.Cleaned[0].Reason != "generated.agents.keep_newest=2" {
		t.Errorf("cleaned: %+v", d.Cleaned)
	}
}

// generated.agents null keeps every folder; the session's end cleans the kind after the log, so a session whose old
// log file goes in that run has its agents' folder go too.
func TestAgentsNullAndSessionEnd(t *testing.T) {
	l := project(t, "  agents: null\n")
	old := agentsSession(t, l, "s-old", daysAgo(60))
	if d := Files(opts(l), Agents); len(d.Cleaned) != 0 || !has(leftIn(t, l), old[0]) {
		t.Errorf("agents: null cleaned %+v", d.Cleaned)
	}
	l = project(t, "")
	old = agentsSession(t, l, "s-old", daysAgo(60))
	lf := write(t, l.Main, ".bonsai/local/log/s-s-old.ndjson", "")
	stamp(t, lf, daysAgo(60))
	d := SessionEnd(opts(l))
	if has(leftIn(t, l), old[3]) {
		t.Errorf("the session's end: %+v, left %v", d, leftIn(t, l))
	}
	if _, err := os.Stat(lf); !os.IsNotExist(err) {
		t.Errorf("the old log file stays: %v", err)
	}
}
