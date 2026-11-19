package bugs

// internal/bugs: the next free id under the lock, two adds at once never taking one id (on every system), a stale lock
// taken over and a live one refused, every move the bug kind makes and none it does not, a carry-over, an edit, and
// one bug record in the log for each write.

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/workspace"
)

func place(t *testing.T) Place {
	t.Helper()
	main := t.TempDir()
	return Place{Main: main, Folder: "work/bugs", Workspace: "ws-7kq2m4xw5r3t6y2u7p4a5c3e2b",
		Local:  workspace.Local{Root: main, Main: main},
		Getenv: func(string) string { return "" },
		Now:    func() time.Time { return time.Date(2026, 11, 19, 8, 0, 0, 0, time.UTC) }}
}

func s(v string) *string { return &v }

func add(t *testing.T, p Place, title string) *Written {
	t.Helper()
	w, e := Add(p, AddInput{Fields: Fields{Title: s(title)}, By: "builder@T-0901"})
	if e != nil {
		t.Fatalf("add %q: %v (%s)", title, e, e.Code)
	}
	return w
}

func TestAddTakesTheNextFreeID(t *testing.T) {
	p := place(t)
	if w := add(t, p, "One"); w.Entry.Bug.ID != "BUG-001" || w.Entry.Path != "work/bugs/BUG-001.md" {
		t.Fatalf("the first bug: %s %s", w.Entry.Bug.ID, w.Entry.Path)
	}
	// A file the reader refuses still holds its number: the next one is above it.
	if err := os.WriteFile(filepath.Join(p.Main, "work", "bugs", "BUG-007.md"), []byte("---\nformat: bonsai.bug/1\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	w := add(t, p, "Two")
	if w.Entry.Bug.ID != "BUG-008" {
		t.Errorf("past a malformed BUG-007: %s", w.Entry.Bug.ID)
	}
	b := w.Entry.Bug
	if b.Status != "open" || b.Created != "2026-11-19" || b.FiledBy == nil || *b.FiledBy != "builder@T-0901" || b.Closed != nil {
		t.Errorf("a new bug's fields: %+v", b)
	}
	if _, err := os.Stat(filepath.Join(p.Main, filepath.FromSlash(LockFile))); !os.IsNotExist(err) {
		t.Errorf("the lock is left behind: %v", err)
	}
}

// Two adds at once (and ten) never take one id: the lock holds the read of the folder and the write together.
func TestAddsAtOnceTakeDistinctIDs(t *testing.T) {
	p := place(t)
	const n = 10
	var wg sync.WaitGroup
	ids := make([]string, n)
	errs := make([]*Error, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			w, e := Add(p, AddInput{Fields: Fields{Title: s("at once")}, By: "builder@T-0901"})
			if e != nil {
				errs[i] = e
				return
			}
			ids[i] = w.Entry.Bug.ID
		}(i)
	}
	wg.Wait()
	seen := map[string]bool{}
	for i, id := range ids {
		if errs[i] != nil {
			t.Fatalf("add %d refused: %v", i, errs[i])
		}
		if seen[id] {
			t.Errorf("two adds took %s", id)
		}
		seen[id] = true
	}
	entries, bad, err := Read(p)
	if err != nil || len(entries) != n || len(bad) != 0 {
		t.Errorf("the folder after ten adds: %d bugs, %d malformed, %v", len(entries), len(bad), err)
	}
}

func TestTheLock(t *testing.T) {
	p := place(t)
	file := filepath.Join(p.Main, filepath.FromSlash(LockFile))
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		t.Fatal(err)
	}
	// A live lock: refused after the wait, bug-busy, nothing written.
	if err := os.WriteFile(file, []byte("pid 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	old := LockWait
	LockWait = 50 * time.Millisecond
	t.Cleanup(func() { LockWait = old })
	if _, e := Add(p, AddInput{Fields: Fields{Title: s("x")}, By: "person"}); e == nil || e.Code != "bug-busy" || e.Exit != 3 {
		t.Fatalf("a live lock: %v", e)
	}
	if entries, _, _ := Read(p); len(entries) != 0 {
		t.Errorf("written under a live lock: %d", len(entries))
	}
	// A stale one (older than a minute): taken over.
	past := time.Now().Add(-2 * time.Minute)
	if err := os.Chtimes(file, past, past); err != nil {
		t.Fatal(err)
	}
	if w := add(t, p, "after a stale lock"); w.Entry.Bug.ID != "BUG-001" {
		t.Errorf("after a stale lock: %s", w.Entry.Bug.ID)
	}
	if _, err := os.Stat(file); !os.IsNotExist(err) {
		t.Errorf("the stale lock is left: %v", err)
	}
}

// Every move of the bug kind is made, with its stamp and fields; every other is refused (bug-move).
func TestMoves(t *testing.T) {
	p := place(t)
	add(t, p, "One")
	k, _ := workspace.BonsaiDocKind("bug")
	allowed := map[string]bool{}
	for _, m := range k.Agent {
		allowed[m[0]+">"+m[1]] = true
	}
	sub := map[string]string{"fixed": "fix", "wontfix": "wontfix", "watched": "watch", "open": "reopen"}
	in := func(to string) MoveInput {
		return MoveInput{By: "builder@T-0902", Why: s("a reason")}
	}
	set := func(status string) {
		raw, _ := os.ReadFile(filepath.Join(p.Main, "work", "bugs", "BUG-001.md"))
		b, body, err := format.ReadBugFile(raw)
		if err != nil {
			t.Fatal(err)
		}
		b.Status = status
		out, _ := b.EncodeFile(body)
		if err := os.WriteFile(filepath.Join(p.Main, "work", "bugs", "BUG-001.md"), out, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, from := range k.Statuses {
		for _, to := range k.Statuses {
			set(from)
			mi := in(to)
			if to == "fixed" {
				mi.Commits = nil // the commit check needs a repository: the move is judged first only when allowed
			}
			_, e := Move(p, sub[to], "BUG-001", to, mi)
			switch {
			case to == "fixed":
				if e == nil || e.Code != "missing-value" {
					t.Errorf("fix with no --commit: %v", e)
				}
			case allowed[from+">"+to] && e != nil:
				t.Errorf("%s to %s refused: %v", from, to, e)
			case !allowed[from+">"+to] && (e == nil || e.Code != "bug-move" || e.Exit != 4):
				t.Errorf("%s to %s: %v, want bug-move", from, to, e)
			}
		}
	}
	// The stamp: wontfix sets closed and why; reopen clears closed, fixed_by and fixed_in.
	set("open")
	w, e := Move(p, "wontfix", "BUG-001", "wontfix", MoveInput{By: "person", Why: s("by design")})
	if e != nil || w.Entry.Bug.Closed == nil || *w.Entry.Bug.Closed != "2026-11-19" || *w.Entry.Bug.Why != "by design" || w.Reason != "open to wontfix" {
		t.Fatalf("wontfix: %v %+v", e, w)
	}
	w, e = Move(p, "reopen", "BUG-001", "open", MoveInput{By: "person"})
	if e != nil || w.Entry.Bug.Closed != nil || w.Entry.Bug.Why != nil || w.Entry.Bug.FixedIn != nil {
		t.Fatalf("reopen: %v %+v", e, w)
	}
	if _, e := Move(p, "watch", "BUG-001", "watched", MoveInput{By: "person"}); e == nil || e.Code != "missing-value" {
		t.Errorf("watch with no --why: %v", e)
	}
	if _, e := Move(p, "watch", "BUG-404", "watched", MoveInput{By: "person", Why: s("x")}); e == nil || e.Code != "bug-not-found" {
		t.Errorf("no such bug: %v", e)
	}
}

func TestCarryOverAndEdit(t *testing.T) {
	p := place(t)
	if _, e := Add(p, AddInput{Fields: Fields{Title: s("x")}, By: "person", Number: 12}); e == nil || e.Code != "bad-flag" {
		t.Errorf("--number without --was: %v", e)
	}
	w, e := Add(p, AddInput{Fields: Fields{Title: s("The switch is not saved")}, By: "builder@T-6028", Was: "TB-012", Number: 12,
		Status: "fixed", Commit: "0a1b2c3d", Created: "2026-10-09"})
	if e != nil {
		t.Fatal(e)
	}
	b := w.Entry.Bug
	if b.ID != "BUG-012" || b.Status != "fixed" || *b.Was != "TB-012" || *b.FixedIn != "0a1b2c3d" || b.Created != "2026-10-09" ||
		b.Closed != nil || w.Reason != "carried over from TB-012 (fixed)" {
		t.Errorf("the carry-over: %+v %q", b, w.Reason)
	}
	if _, e := Add(p, AddInput{Fields: Fields{Title: s("x")}, By: "person", Was: "TB-099", Number: 12}); e == nil || e.Code != "bug-id-taken" {
		t.Errorf("a number held: %v", e)
	}
	if _, e := Add(p, AddInput{Fields: Fields{Title: s("x")}, By: "person", Was: "TB-012", Number: 40}); e == nil || e.Code != "bug-id-taken" {
		t.Errorf("an old id carried over twice: %v", e)
	}
	if w := add(t, p, "Next"); w.Entry.Bug.ID != "BUG-013" {
		t.Errorf("after a carry-over at 12: %s", w.Entry.Bug.ID)
	}
	// Found by its old id; an edit changes only what is given, the body kept.
	entries, _, _ := Read(p)
	if en, e := Find(entries, "TB-012"); e != nil || en.Bug.ID != "BUG-012" {
		t.Errorf("found by its was: %v", e)
	}
	w, e = Edit(p, "TB-012", Fields{Where: s("src/settings/store.ts:88"), Body: s("How to repeat it.\n")})
	if e != nil || *w.Entry.Bug.Where != "src/settings/store.ts:88" || w.Entry.Body != "How to repeat it.\n" || w.Reason != "edited where, body" {
		t.Fatalf("edit: %v %+v", e, w)
	}
	w, e = Edit(p, "BUG-012", Fields{Where: s("")})
	if e != nil || w.Entry.Bug.Where != nil || w.Entry.Body != "How to repeat it.\n" {
		t.Errorf("an edit to none: %v %+v", e, w)
	}
	if _, e := Edit(p, "BUG-012", Fields{Title: s("The switch is not saved")}); e == nil || e.Code != "bad-flag" {
		t.Errorf("an edit that changes nothing: %v", e)
	}
}

// Each write appends one bug record: target the id, kind the subcommand, reason the change.
func TestEachWriteIsLogged(t *testing.T) {
	p := place(t)
	add(t, p, "One")
	if _, e := Move(p, "wontfix", "BUG-001", "wontfix", MoveInput{By: "person", Why: s("by design")}); e != nil {
		t.Fatal(e)
	}
	files, _ := filepath.Glob(filepath.Join(p.Main, ".bonsai", "local", "log", "*.ndjson"))
	var lines []string
	for _, f := range files {
		raw, _ := os.ReadFile(f)
		lines = append(lines, strings.Split(strings.TrimSpace(string(raw)), "\n")...)
	}
	if len(lines) != 2 {
		t.Fatalf("records: %q", lines)
	}
	for i, want := range []struct{ kind, reason string }{{"add", "filed (open)"}, {"wontfix", "open to wontfix"}} {
		l, err := format.ReadLog([]byte(lines[i]))
		if err != nil {
			t.Fatal(err)
		}
		if l.Event != "bug" || *l.Target != "BUG-001" || *l.Kind != want.kind || *l.Reason != want.reason || *l.Text != "One" {
			t.Errorf("record %d: %s", i, lines[i])
		}
	}
}

func TestBy(t *testing.T) {
	env := func(m map[string]string) func(string) string { return func(k string) string { return m[k] } }
	for _, c := range []struct {
		by   string
		env  map[string]string
		want string
		code string
	}{
		{"", map[string]string{}, "person", ""},
		{"", map[string]string{"BONSAI_ROLE": "builder", "BONSAI_TASK": "T-0901"}, "builder@T-0901", ""},
		{"", map[string]string{"BONSAI_ROLE": "orchestrator"}, "orchestrator", ""},
		{"", map[string]string{"CLAUDE_CODE_CHILD_SESSION": "1"}, "", "missing-value"},
		{"builder@T-0901", map[string]string{"CLAUDE_CODE_CHILD_SESSION": "1"}, "builder@T-0901", ""},
		{"person", map[string]string{"AI_AGENT": "x"}, "", "bad-value"},
		{"a\nb", map[string]string{}, "", "bad-value"},
	} {
		got, e := By(c.by, env(c.env))
		if got != c.want || (e == nil) != (c.code == "") || (e != nil && e.Code != c.code) {
			t.Errorf("By(%q, %v) = %q, %v; want %q, %s", c.by, c.env, got, e, c.want, c.code)
		}
	}
}
