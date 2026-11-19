package rung0

// Rung 0 and bugs (plan P-T-6023, step 9): the bugs page is judged by the marker on the main checkout's copy (a branch
// changing an unmarked page passes, one adding the marker passes, one changing a marked page is refused); a branch
// changing a bug file's status line is refused (status-line); on the base, the page and the bug files' moves are
// bonsai's own, as the tables are.

import (
	"strconv"
	"testing"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/guard"
)

const handKept = "---\ntitle: Known bugs\n---\n\n# Known bugs\n\n| Bug | Title |\n|---|---|\n| TB-001 | One |\n"

func bugMD(status string) string {
	return "---\nformat: bonsai.bug/1\nid: BUG-001\ntitle: One\nstatus: " + status + "\narea: null\nwhere: null\nfound_by: null\n" +
		"filed_by: person\nfixed_by: null\nfixed_in: null\ntask: null\nwas: null\nwhy: null\ncreated: 2026-10-13\nclosed: null\n" +
		"labels: {}\n---\n"
}

func markedPage(t *testing.T, rows ...format.BugRow) string {
	t.Helper()
	p := &format.Bugs{Bugs: append([]format.BugRow{}, rows...)}
	b, err := p.Encode()
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestTheBugsPage(t *testing.T) {
	const page = "work/bugs.md"
	t.Run("an unmarked page changed on a branch", func(t *testing.T) {
		f := newProject(t, map[string]string{page: handKept})
		wt := f.worktree(t)
		write(t, wt, page, handKept+"| TB-002 | Two |\n")
		commit(t, wt, "a row")
		if o := f.climb(t, wt, "T-0901"); len(o.Refused) != 0 {
			t.Fatalf("an unmarked page is refused: %s", o.Reason)
		}
	})
	t.Run("a branch adds the marker", func(t *testing.T) {
		f := newProject(t, map[string]string{page: handKept})
		wt := f.worktree(t)
		write(t, wt, page, markedPage(t))
		commit(t, wt, "the page, generated")
		if o := f.climb(t, wt, "T-0901"); len(o.Refused) != 0 {
			t.Fatalf("a branch adding the marker is refused: %s", o.Reason)
		}
	})
	t.Run("a marked page changed on a branch", func(t *testing.T) {
		f := newProject(t, map[string]string{page: markedPage(t)})
		wt := f.worktree(t)
		write(t, wt, page, handKept) // the marker stripped
		commit(t, wt, "the page by hand")
		o := f.climb(t, wt, "T-0901")
		if got := refusedRules(o); got[page] != guard.RuleBonsaiFiles {
			t.Fatalf("refused %v: %s", got, o.Reason)
		}
		red(t, o, strconv.Quote(page)+" (bonsai-files)")
	})
	t.Run("a marked page on the base", func(t *testing.T) {
		f := newProject(t, map[string]string{page: markedPage(t)})
		write(t, f.main, page, markedPage(t, format.BugRow{ID: "BUG-001", Title: "One", Status: "open"}))
		write(t, f.main, "work/bugs/BUG-001.md", bugMD("open"))
		if o := f.climb(t, f.main, "T-0901"); len(o.Refused) != 0 {
			t.Fatalf("the page rebuilt on the base is refused: %s", o.Reason)
		}
	})
}

func TestABugsStatusLine(t *testing.T) {
	const bug = "work/bugs/BUG-001.md"
	t.Run("on a branch", func(t *testing.T) {
		f := newProject(t, map[string]string{bug: bugMD("open")})
		wt := f.worktree(t)
		write(t, wt, bug, bugMD("fixed"))
		o := f.climb(t, wt, "T-0901")
		if got := refusedRules(o); got[bug] != guard.RuleStatusLine {
			t.Fatalf("refused %v: %s", got, o.Reason)
		}
	})
	t.Run("a body edit on a branch", func(t *testing.T) {
		f := newProject(t, map[string]string{bug: bugMD("open")})
		wt := f.worktree(t)
		write(t, wt, bug, bugMD("open")+"\nHow to repeat it.\n")
		if o := f.climb(t, wt, "T-0901"); len(o.Refused) != 0 {
			t.Fatalf("a body edit is refused: %s", o.Reason)
		}
	})
	t.Run("on the base: bonsai bug's own", func(t *testing.T) {
		f := newProject(t, map[string]string{bug: bugMD("open")})
		write(t, f.main, bug, bugMD("watched"))
		if o := f.climb(t, f.main, "T-0901"); len(o.Refused) != 0 {
			t.Fatalf("a move on the base is refused: %s", o.Reason)
		}
	})
}
