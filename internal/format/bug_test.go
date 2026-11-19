package format

// The bug file and the bugs page (set 14): the page's sections are bonsai.bug/1's statuses, its marker is read from
// the bytes, and a page written from rows reads back as the same rows.

import (
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/schema"
)

// Every status of bonsai.bug/1 has one section on the page, and every section is a status: the schema is the list's
// one home.
func TestBugSectionsAreTheStatuses(t *testing.T) {
	props, _ := MustLookup("bug").Schema().Get("properties")
	st, _ := props.(schema.Object).Get("status")
	enum, _ := st.(schema.Object).Get("enum")
	var want []string
	for _, v := range enum.([]any) {
		want = append(want, v.(string))
	}
	var got []string
	for _, s := range BugSections {
		got = append(got, s.Status)
	}
	if len(got) != len(want) {
		t.Fatalf("the page's sections are %v, bonsai.bug/1's statuses %v", got, want)
	}
	for _, w := range want {
		if !bugSection(w) {
			t.Errorf("the status %s has no section on the page", w)
		}
	}
}

func TestBugsMarked(t *testing.T) {
	for _, c := range []struct {
		raw  string
		want bool
	}{
		{"---\nformat: bonsai.bugs/1   # generated\n---\n", true},
		{"---\r\nformat: bonsai.bugs/1\r\n---\r\n", true},
		{"\xef\xbb\xbf---\nformat: bonsai.bugs/1\n---\n", true},
		{"---\nformat: bonsai.bugs/2\n---\n", true},
		{"---\ntitle: Known bugs   # the bugs register; skill workflow:bugs\nupdated: 2026-11-06\n---\n", false},
		{"# Known bugs\n\n## Open\n", false},
		{"---\ntitle: x\nformat: bonsai.bugs/1\n---\n", false},
		{"", false},
		{"format: bonsai.bugs/1\n", false},
	} {
		if got := BugsMarked([]byte(c.raw)); got != c.want {
			t.Errorf("BugsMarked(%q) = %v, want %v", c.raw, got, c.want)
		}
	}
}

func TestBugsPageRoundTrip(t *testing.T) {
	s := func(v string) *string { return &v }
	p := &Bugs{Bugs: []BugRow{
		{ID: "BUG-003", Title: "A pipe | in a title", Status: "open", Where: s("a.go:1")},
		{ID: "BUG-002", Title: "Two", Status: "wontfix", Why: s("by design"), Closed: s("2026-11-01")},
		{ID: "BUG-001", Title: "One", Status: "fixed", FixedIn: s("0a1b2c3d"), Task: s("T-0001"), Closed: s("2026-11-02")},
	}}
	out, err := p.Encode()
	if err != nil {
		t.Fatal(err)
	}
	if !BugsMarked(out) {
		t.Errorf("the page written carries no marker:\n%s", out)
	}
	back, err := ReadBugs(out)
	if err != nil {
		t.Fatalf("the page written does not read back: %v\n%s", err, out)
	}
	// Read back in the page's order: open, then fixed, then wontfix.
	var ids []string
	for _, r := range back.Bugs {
		ids = append(ids, r.ID+" "+r.Status)
	}
	if got := strings.Join(ids, ", "); got != "BUG-003 open, BUG-001 fixed, BUG-002 wontfix" {
		t.Errorf("read back as %s", got)
	}
	if back.Bugs[0].Title != "A pipe | in a title" {
		t.Errorf("a pipe in a cell reads back as %q", back.Bugs[0].Title)
	}
	again, err := back.Encode()
	if err != nil || string(again) != string(out) {
		t.Errorf("the page is not byte-stable (%v):\n%s\nthen\n%s", err, out, again)
	}
	bad := &Bugs{Bugs: []BugRow{{ID: "BUG-004", Title: "x", Status: "closed"}}}
	if _, err := bad.Encode(); err == nil {
		t.Error("a row whose status has no section was written")
	}
}
