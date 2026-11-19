package main

// bonsai bug through the command line, in a linked project: add, list, show, the four moves and edit, each --json
// against bonsai.buglist/1; a carry-over; fix --commit refusing a commit the repository lacks; a worktree writing the
// main checkout's folder; who writes from BONSAI_ROLE and --by; every refusal with its word and exit code; one bug
// record in the log per write; and bonsai check finding nothing in the files written. internal/bugs's tests walk the
// rules (the lock, two adds at once, every move); these hold the command to them.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/asks"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/testpack"
)

// bugDoc runs a bug command with --json, holds its output to bonsai.buglist/1, and gives the document.
func (c *cli) bugDoc(exit int, args ...string) schema.Object {
	c.t.Helper()
	code, out, errOut := c.run("", append(args, "--json")...)
	if code != exit {
		c.t.Fatalf("%v: exit %d, want %d\n%s%s", args, code, exit, out, errOut)
	}
	return fits(c.t, out, "buglist")
}

// bugOf is the frontmatter of a buglist document's i-th bug.
func bugOf(t *testing.T, doc schema.Object, i int) schema.Object {
	t.Helper()
	list, _ := doc.Get("bugs")
	items, _ := list.([]any)
	if i >= len(items) {
		t.Fatalf("no bug %d in %s", i, schema.Show(doc))
	}
	b, _ := items[i].(schema.Object).Get("bug")
	return b.(schema.Object)
}

func bugClockAt(t *testing.T) {
	t.Helper()
	old := bugClock
	t.Cleanup(func() { bugClock = old })
	bugClock = func() time.Time { return time.Date(2026, 11, 19, 8, 0, 0, 0, time.UTC) }
}

func TestBugCommands(t *testing.T) {
	c := newCLI(t)
	t.Setenv(asks.SessionEnv, "")
	t.Setenv("BONSAI_ROLE", "")
	t.Setenv("BONSAI_TASK", "")
	bugClockAt(t)
	// Not linked: exit 4, not-linked, the workspace null.
	doc := c.bugDoc(4, "bug", "list")
	if w, _ := doc.Get("workspace"); w != nil || errorIn(doc, "buglist").String("code") != "not-linked" || doc.String("command") != "list" {
		t.Errorf("not linked: %s", schema.Show(doc))
	}
	c.link()
	c.commit("link")
	folder := "work/bugs"

	// add: the id and the file printed; the file at open, filed by a person in a person's terminal.
	code, out, errOut := c.run("", "bug", "add", "--title", "The page forgets the dark mode switch", "--area", "settings",
		"--where", "src/settings/store.ts:88", "--body", "How to repeat it: restart.\n")
	if code != 0 || out != "BUG-001\n"+folder+"/BUG-001.md\n" {
		t.Fatalf("add: %d %q %q", code, out, errOut)
	}
	raw := c.read(folder + "/BUG-001.md")
	if !strings.HasPrefix(raw, "---\nformat: bonsai.bug/1   # fields: bonsai check --schema bonsai.bug; how to fill: skill base:bug\nid: BUG-001\n") ||
		!strings.Contains(raw, "\nstatus: open\n") || !strings.Contains(raw, "\nfiled_by: person\n") || !strings.Contains(raw, "\ncreated: 2026-11-19\n") ||
		!strings.HasSuffix(raw, "---\n\nHow to repeat it: restart.\n") {
		t.Errorf("the file written:\n%s", raw)
	}
	// An agent's role and task from the environment.
	t.Setenv("BONSAI_ROLE", "builder")
	t.Setenv("BONSAI_TASK", "T-0901")
	doc = c.bugDoc(0, "bug", "add", "--title", "Second")
	if b := bugOf(t, doc, 0); b.String("id") != "BUG-002" || b.String("filed_by") != "builder@T-0901" || doc.String("command") != "add" {
		t.Errorf("add --json: %s", schema.Show(doc))
	}
	t.Setenv("BONSAI_ROLE", "")
	t.Setenv("BONSAI_TASK", "")

	// list and show.
	code, out, _ = c.run("", "bug", "list")
	if code != 0 || out != "BUG-001  open     The page forgets the dark mode switch\nBUG-002  open     Second\n" {
		t.Errorf("list: %d %q", code, out)
	}
	doc = c.bugDoc(0, "bug", "list", "--status", "fixed")
	if l, _ := doc.Get("bugs"); len(l.([]any)) != 0 {
		t.Errorf("list --status fixed: %s", schema.Show(doc))
	}
	doc = c.bugDoc(0, "bug", "show", "BUG-001")
	list, _ := doc.Get("bugs")
	if body, _ := list.([]any)[0].(schema.Object).Get("body"); body != "How to repeat it: restart.\n" {
		t.Errorf("show's body: %s", schema.Show(doc))
	}

	// fix: a commit the repository lacks is refused; one it holds is written, with who and the day.
	doc = c.bugDoc(2, "bug", "fix", "BUG-001", "--commit", "0123abcd")
	if errorIn(doc, "buglist").String("code") != "commit-not-found" {
		t.Errorf("fix with a commit not there: %s", schema.Show(doc))
	}
	head := strings.TrimSpace(testpack.Git(t, c.root, "rev-parse", "HEAD"))
	doc = c.bugDoc(0, "bug", "fix", "BUG-001", "--commit", head[:10], "--how", "saved before the page closes", "--by", "builder@T-0904", "--task", "T-0904")
	if b := bugOf(t, doc, 0); b.String("status") != "fixed" || b.String("fixed_in") != head[:10]+" (saved before the page closes)" ||
		b.String("fixed_by") != "builder@T-0904" || b.String("closed") != "2026-11-19" || b.String("task") != "T-0904" {
		t.Errorf("fix: %s", schema.Show(doc))
	}
	if doc := c.bugDoc(4, "bug", "fix", "BUG-001", "--commit", head[:10]); errorIn(doc, "buglist").String("code") != "bug-move" {
		t.Errorf("fix twice: %s", schema.Show(doc))
	}
	if doc := c.bugDoc(2, "bug", "wontfix", "BUG-002"); errorIn(doc, "buglist").String("code") != "missing-value" {
		t.Errorf("wontfix with no --why: %s", schema.Show(doc))
	}
	c.bugDoc(0, "bug", "watch", "BUG-002", "--why", "seen once")
	doc = c.bugDoc(0, "bug", "reopen", "BUG-001", "--why", "seen again")
	if b := bugOf(t, doc, 0); b.String("status") != "open" || b.String("why") != "seen again" {
		t.Errorf("reopen: %s", schema.Show(doc))
	}
	if b, _ := bugOf(t, doc, 0).Get("closed"); b != nil {
		t.Errorf("reopen keeps closed: %s", schema.Show(doc))
	}
	doc = c.bugDoc(0, "bug", "edit", "BUG-001", "--title", "The switch is not saved", "--area", "")
	if b := bugOf(t, doc, 0); b.String("title") != "The switch is not saved" {
		t.Errorf("edit: %s", schema.Show(doc))
	}
	if !strings.HasSuffix(c.read(folder+"/BUG-001.md"), "---\n\nHow to repeat it: restart.\n") {
		t.Errorf("edit lost the body:\n%s", c.read(folder+"/BUG-001.md"))
	}

	// A carry-over, then shown by its old id.
	code, out, errOut = c.run("", "bug", "add", "--was", "TB-012", "--number", "12", "--status", "fixed", "--commit", "0a1b2c3d", "--title", "Old one")
	if code != 0 || out != "BUG-012\n"+folder+"/BUG-012.md\n" {
		t.Fatalf("carry-over: %d %q %q", code, out, errOut)
	}
	if doc := c.bugDoc(0, "bug", "show", "TB-012"); bugOf(t, doc, 0).String("id") != "BUG-012" {
		t.Errorf("show by the old id: %s", schema.Show(doc))
	}
	if doc := c.bugDoc(0, "bug", "add", "--title", "After"); bugOf(t, doc, 0).String("id") != "BUG-013" {
		t.Errorf("the next id after a carry-over: %s", schema.Show(doc))
	}

	// Each write is one bug record in the log.
	day := c.read(".bonsai/local/log/w-2026-11-19.ndjson")
	if n := strings.Count(day, `"event":"bug"`); n != 8 || strings.Count(day, "\n") != 8 || !strings.Contains(day, `"reason":"open to fixed"`) {
		t.Errorf("bug records: %d, want 8 (two adds, fix, watch, reopen, edit, a carry-over, an add)\n%s", n, day)
	}
	// bonsai check finds nothing in the files bonsai bug wrote (the test pack's own absolute source is its finding).
	if _, out, errOut := c.run("", "check", "--json"); strings.Contains(out, `"bug-`) || strings.Contains(out, `"file": "work/bugs/`) {
		t.Errorf("check after the writes:\n%s%s", out, errOut)
	}
}

// From a worktree, bonsai bug writes and reads the main checkout's folder, never the worktree's.
func TestBugFromAWorktree(t *testing.T) {
	c := newCLI(t)
	t.Setenv(asks.SessionEnv, "")
	bugClockAt(t)
	c.link()
	c.commit("link")
	wt := filepath.Join(c.tmp, "branch")
	testpack.Git(t, c.root, "worktree", "add", "-q", "-b", "branch", wt)
	t.Chdir(wt)
	code, out, errOut := c.run("", "bug", "add", "--title", "From the worktree")
	if code != 0 {
		t.Fatalf("from a worktree: %d %s", code, errOut)
	}
	if _, err := os.Stat(filepath.Join(c.root, "work", "bugs", "BUG-001.md")); err != nil {
		t.Errorf("not in the main checkout's folder: %v (%s)", err, out)
	}
	if _, err := os.Stat(filepath.Join(wt, "work", "bugs")); !os.IsNotExist(err) {
		t.Errorf("the worktree has a bug folder of its own: %v", err)
	}
	if code, out, _ := c.run("", "bug", "list"); code != 0 || !strings.Contains(out, "BUG-001") {
		t.Errorf("list from the worktree: %d %q", code, out)
	}
}

// Every refusal: its exit code and word, in a document that fits bonsai.buglist/1, and nothing written.
func TestBugRefusals(t *testing.T) {
	c := newCLI(t)
	t.Setenv(asks.SessionEnv, "")
	t.Setenv("BONSAI_ROLE", "")
	bugClockAt(t)
	c.link()
	c.commit("link")
	c.bugDoc(0, "bug", "add", "--title", "One")
	before := c.read("work/bugs/BUG-001.md")
	for _, r := range []struct {
		args []string
		exit int
		code string
	}{
		{[]string{"bug", "add"}, 2, "missing-value"},
		{[]string{"bug", "add", "--title", "x", "--number", "3"}, 2, "bad-flag"},
		{[]string{"bug", "add", "--title", "x", "--was", "TB-1", "--number", "1"}, 4, "bug-id-taken"},
		{[]string{"bug", "add", "--title", "x", "--was", "TB-1", "--status", "closed"}, 2, "bad-value"},
		{[]string{"bug", "add", "--title", "x", "--task", "T-1"}, 2, "bad-value"},
		{[]string{"bug", "fix", "BUG-001"}, 2, "missing-value"},
		{[]string{"bug", "fix", "BUG-001", "--commit", "--all"}, 2, "bad-flag"},
		{[]string{"bug", "fix", "BUG-001", "--commit", "main"}, 2, "bad-value"},
		{[]string{"bug", "fix", "BUG-404", "--commit", "abcd1234"}, 2, "commit-not-found"},
		{[]string{"bug", "watch", "BUG-404", "--why", "x"}, 4, "bug-not-found"},
		{[]string{"bug", "reopen", "BUG-001"}, 4, "bug-move"},
		{[]string{"bug", "edit", "BUG-001", "--title", "One"}, 2, "bad-flag"},
		{[]string{"bug", "show"}, 2, "bad-flag"},
		{[]string{"bug", "list", "--status", "closed"}, 2, "bad-value"},
		{[]string{"bug", "nothing"}, 2, "bad-value"},
	} {
		code, out, errOut := c.run("", append(r.args, "--json")...)
		if code != r.exit {
			t.Errorf("%v: exit %d, want %d\n%s%s", r.args, code, r.exit, out, errOut)
			continue
		}
		if r.args[1] == "nothing" {
			continue // not a bug word: bonsai bug's own refusal
		}
		doc := fits(t, out, "buglist")
		if got := errorIn(doc, "buglist").String("code"); got != r.code {
			t.Errorf("%v: %s, want %s", r.args, got, r.code)
		}
		code, _, errOut = c.run("", r.args...)
		if code != r.exit || nextOf(errOut) == "" {
			t.Errorf("%v as text: %d, no next step\n%s", r.args, code, errOut)
		}
	}
	// An agent's session must say who writes.
	t.Setenv("CLAUDE_CODE_CHILD_SESSION", "1")
	if doc := c.bugDoc(2, "bug", "add", "--title", "x"); errorIn(doc, "buglist").String("code") != "missing-value" {
		t.Errorf("an agent with no --by: %s", schema.Show(doc))
	}
	if doc := c.bugDoc(2, "bug", "add", "--title", "x", "--by", "person"); errorIn(doc, "buglist").String("code") != "bad-value" {
		t.Errorf("an agent as person: %s", schema.Show(doc))
	}
	if after := c.read("work/bugs/BUG-001.md"); after != before {
		t.Errorf("a refusal changed the bug:\n%s", after)
	}
	if files, _ := filepath.Glob(filepath.Join(c.root, "work", "bugs", "*")); len(files) != 1 {
		t.Errorf("a refusal wrote a file: %v", files)
	}
}

// In command mode (this machine's status_writes: command) bonsai bug writes every move itself, as in agents mode: the
// switch is for tasks only, and no refusal names the status command.
func TestBugMovesInCommandMode(t *testing.T) {
	c := newCLI(t)
	t.Setenv(asks.SessionEnv, "")
	t.Setenv("BONSAI_ROLE", "")
	bugClockAt(t)
	c.link()
	c.commit("link")
	if code, out, errOut := c.run("", "settings", "set", "status_writes=command", "status_command=studio task move"); code != exitOK {
		t.Fatalf("command mode: %d %s%s", code, out, errOut)
	}
	c.bugDoc(0, "bug", "add", "--title", "One")
	c.bugDoc(0, "bug", "wontfix", "BUG-001", "--why", "by design")
	doc := c.bugDoc(0, "bug", "reopen", "BUG-001")
	if b := bugOf(t, doc, 0); b.String("status") != "open" {
		t.Errorf("reopen in command mode: %s", schema.Show(doc))
	}
	if raw := c.read("work/bugs/BUG-001.md"); strings.Contains(raw, "studio") {
		t.Errorf("the status command reached a bug file:\n%s", raw)
	}
}
