package clean

// readSpans against sessions.Read, the one reader that says "open": the same open spans, the same ended spans and the
// same row keys as sessions.Found's, on every span rule (a compaction, a lost end line, a subagent run with and without
// its stop, the 24-hour rule both ways, a torn line, a record glued after one, a day file) and on many files made at
// random from those parts.

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/sessions"
)

// spanView is what the cleaner judges by in a file: its open spans and its ended spans' keys.
func spanView(f *sessions.File) string {
	var out []string
	for _, s := range f.Spans {
		state := "ended"
		if s.Open {
			state = "open"
		}
		out = append(out, fmt.Sprintf("%s|%s|%s|%s|%s", s.Session, s.Subagent, s.Start.UTC().Format(time.RFC3339Nano), state, s.Task))
	}
	sort.Strings(out)
	return strings.Join(out, "\n")
}

func TestReadSpansIsSessionsRead(t *testing.T) {
	l := project(t, "")
	dir := filepath.Join(l.Main, ".bonsai", "local", "log")
	base := now.Add(-30 * time.Hour)
	at := func(m int) time.Time { return base.Add(time.Duration(m) * time.Minute) }
	cases := map[string][]rec{
		"ended": session(sid("1"), "T-0001", at(0), at(30)),
		"compaction": {{event: "session_start", at: at(0), session: sid("2"), source: "startup"},
			{event: "session_start", at: at(5), session: sid("2"), source: "compact"},
			{event: "session_end", at: at(9), session: sid("2")}},
		"lost end": {{event: "session_start", at: at(0), session: sid("3"), source: "startup"},
			{event: "tool_end", at: at(3), session: sid("3")},
			{event: "session_start", at: at(10), session: sid("3"), source: "resume"},
			{event: "session_end", at: at(20), session: sid("3")}},
		"subagents": {{event: "session_start", at: at(0), session: sid("4"), source: "startup"},
			{event: "subagent_start", at: at(1), session: sid("4"), subagent: "c0000001-y"},
			{event: "subagent_stop", at: at(2), session: sid("4"), subagent: "c0000001-y"},
			{event: "subagent_start", at: at(3), session: sid("4"), subagent: "c0000002-y"},
			{event: "session_end", at: at(4), session: sid("4")}},
		"open, recent": {{event: "session_start", at: now.Add(-time.Hour), session: sid("5"), source: "startup"},
			{event: "tool_end", at: now.Add(-time.Minute), session: sid("5")}},
		"open, stale": {{event: "session_start", at: at(0), session: sid("6"), source: "startup"},
			{event: "tool_end", at: at(5), session: sid("6")}},
		"open, its last record recent": {{event: "session_start", at: at(0), session: sid("7"), source: "startup"},
			{event: "tool_end", at: now.Add(-23 * time.Hour), session: sid("7")}},
	}
	write := func(name string, recs []rec, extra string) {
		var b strings.Builder
		for _, r := range recs {
			b.Write(logLine(t, r))
		}
		b.WriteString(extra)
		if err := os.WriteFile(filepath.Join(dir, name), []byte(b.String()), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	i := 0
	for _, name := range sortedKeys(cases) {
		i++
		write(fmt.Sprintf("s-%s-case%d.ndjson", sid("9")[:8], i), cases[name], "")
	}
	// A torn last line after an open span, and a record glued after a torn line.
	write("s-"+sid("8")+"-torn.ndjson", cases["open, recent"], `{"format":"bonsai.log/1","id":"tor`)
	glued := strings.TrimSuffix(string(logLine(t, rec{event: "session_start", at: at(0), session: sid("8"), source: "startup"})), "\n")
	write("s-"+sid("8")+"-glued.ndjson", nil, `{"format":"bonsai.log/1","id":"tor`+glued+"\n"+
		string(logLine(t, rec{event: "session_end", at: at(2), session: sid("8")})))
	// Many files made at random from the parts, each with torn lines here and there.
	rng := rand.New(rand.NewSource(7))
	events := []string{"session_start", "session_end", "subagent_start", "subagent_stop", "tool_end", "guard", "prompt"}
	for n := 0; n < 300; n++ {
		id := fmt.Sprintf("d%07d-0000-4000-8000-000000000000", n)
		var b strings.Builder
		m := 0
		for k := 0; k < 1+rng.Intn(30); k++ {
			m += rng.Intn(600)
			r := rec{event: events[rng.Intn(len(events))], at: base.Add(time.Duration(m) * time.Minute), session: id}
			switch r.event {
			case "session_start":
				r.source = []string{"startup", "resume", "compact", "clear"}[rng.Intn(4)]
				r.target = []string{"", "T-0001"}[rng.Intn(2)]
			case "subagent_start", "subagent_stop":
				r.subagent = fmt.Sprintf("e%07d-z", rng.Intn(3))
			}
			b.Write(logLine(t, r))
			if rng.Intn(15) == 0 {
				b.WriteString(`{"format":"bonsai.log/1","id":"to`) // a torn write, glued to the next record
			}
		}
		if rng.Intn(6) == 0 {
			b.WriteString(`{"format":"bonsai.log/1","id":"to`) // a torn last line
		}
		if err := os.WriteFile(filepath.Join(dir, "s-"+id+".ndjson"), []byte(b.String()), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	ourKeys := map[string]bool{}
	opened := 0
	for _, e := range entries {
		p := filepath.Join(dir, e.Name())
		want, err := sessions.Read(p, now)
		if err != nil {
			t.Fatal(err)
		}
		got, err := readSpans(p, e.Name(), now)
		if err != nil {
			t.Errorf("%s: %v", e.Name(), err)
			continue
		}
		if spanView(got) != spanView(want) || got.HasOpen() != want.HasOpen() {
			t.Errorf("%s:\nread by the cleaner\n%s\nby sessions.Read\n%s", e.Name(), spanView(got), spanView(want))
		}
		if want.HasOpen() {
			opened++
		}
		for _, s := range got.Spans {
			if k, ok := rowKey(s); ok {
				ourKeys[k] = true
			}
		}
	}
	found, err := sessions.Found(dir, now)
	if err != nil {
		t.Fatal(err)
	}
	theirs := map[string]bool{}
	for _, r := range found {
		theirs[sessions.Key(r)] = true
	}
	if len(theirs) != len(ourKeys) {
		t.Errorf("%d row keys, sessions.Found's %d", len(ourKeys), len(theirs))
	}
	for k := range theirs {
		if !ourKeys[k] {
			t.Errorf("sessions.Found's row %q is not one of the cleaner's", k)
		}
	}
	if opened == 0 || len(theirs) < 100 {
		t.Errorf("the files hold %d open and %d ended: too few to compare", opened, len(theirs))
	}
}

func sortedKeys[V any](m map[string]V) []string {
	var out []string
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// A file whose last lines do not read cannot be judged, and is kept.
func TestUnreadableTailIsKept(t *testing.T) {
	l := project(t, "")
	p := filepath.Join(l.Main, ".bonsai", "local", "log", "s-"+sid("1")+".ndjson")
	body := string(logLine(t, rec{event: "guard", at: daysAgo(40), session: sid("1")})) + strings.Repeat("not a record\n", tailLines)
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := readSpans(p, filepath.Base(p), now); err == nil {
		t.Error("a file whose last lines do not read was judged")
	}
}
