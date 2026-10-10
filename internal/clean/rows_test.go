package clean

// The sessions table's rows, as check --write cleans them (Rows, then Record once the table is written).

import (
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/sessions"
)

func TestRows(t *testing.T) {
	l := project(t, "  sessions:\n    keep_days: 30\n")
	taskFile(t, l, "T-0001", "done")
	taskFile(t, l, "T-0002", "running")
	taskFile(t, l, "T-0003", "cut")
	old, older := daysAgo(40), daysAgo(41)
	sub := row(sid("6"), "T-0001", older, old)
	sub.Kind, sub.Subagent = "subagent", ptr("b0000006")
	table := &format.Sessions{Sessions: []format.SessionRow{
		row(sid("1"), "T-0001", older, old),                 // done, old, its log gone: goes
		row(sid("2"), "T-0002", older, old),                 // its task running: kept
		row(sid("3"), "none", older, old),                   // no task, old: goes
		row(sid("4"), "T-0001", older, old),                 // its span still in the log: kept
		row(sid("5"), "T-0001", daysAgo(3), daysAgo(2)),     // new: stays
		row(sid("7"), "T-0099", older, old),                 // its task not found: kept
		sub,                                                 // a subagent run's row, done, old: goes
		row(sid("8"), "T-0003", daysAgo(31.1), daysAgo(31)), // cut, just past the rule: goes
	}}
	table.Hours = sessions.Hours(table.Sessions)
	found := []format.SessionRow{row(sid("4"), "T-0001", older, old)}
	kept, items, err := Rows(opts(l), table, found)
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, it := range items {
		got = append(got, it.Target+"="+it.Reason)
	}
	want := []string{
		".bonsai/sessions.md#a0000001@" + older.Format(rowLayout) + "=generated.sessions.keep_days=30",
		".bonsai/sessions.md#a0000003@" + older.Format(rowLayout) + "=generated.sessions.keep_days=30",
		".bonsai/sessions.md#a0000006/b0000006@" + older.Format(rowLayout) + "=generated.sessions.keep_days=30",
		".bonsai/sessions.md#a0000008@" + daysAgo(31.1).Format(rowLayout) + "=generated.sessions.keep_days=30",
	}
	if !same(got, want) {
		t.Errorf("cleaned\n  %s\nwant\n  %s", strings.Join(got, "\n  "), strings.Join(want, "\n  "))
	}
	var stay []string
	for _, r := range kept.Sessions {
		stay = append(stay, r.Session)
	}
	if strings.Join(stay, " ") != "a0000002 a0000004 a0000005 a0000007" {
		t.Errorf("rows kept %v", stay)
	}
	if want := sessions.Hours(kept.Sessions); len(kept.Hours) != len(want) || len(want) == len(table.Hours) {
		t.Errorf("hours %v, want %v (rebuilt from the rows that stay)", kept.Hours, want)
	}
	if b, err := kept.Encode(); err != nil || !strings.Contains(string(b), "a0000005") {
		t.Errorf("the table does not encode: %v", err)
	}
	// Record writes one clean record per row, in today's day file.
	if err := Record(opts(l), items); err != nil {
		t.Fatal(err)
	}
	if recs := cleanRecords(t, l); len(recs) != 4 {
		t.Errorf("records %v", recs)
	}
	// A second pass over what stays cleans nothing.
	if _, again, err := Rows(opts(l), kept, found); err != nil || len(again) != 0 {
		t.Errorf("a second pass cleaned %v (%v)", again, err)
	}
}

// keep_newest on rows: by end, newest first; a protected row counts among the newest.
func TestRowsKeepNewest(t *testing.T) {
	l := project(t, "  sessions:\n    keep_newest: 2\n")
	taskFile(t, l, "T-0001", "done")
	taskFile(t, l, "T-0002", "running")
	table := &format.Sessions{Sessions: []format.SessionRow{
		row(sid("1"), "T-0002", daysAgo(1.1), daysAgo(1)), // newest, running: kept
		row(sid("2"), "T-0001", daysAgo(2.1), daysAgo(2)), // second: stays
		row(sid("3"), "T-0001", daysAgo(3.1), daysAgo(3)), // third: goes
	}}
	kept, items, err := Rows(opts(l), table, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || !strings.Contains(items[0].Target, "a0000003") || items[0].Reason != "generated.sessions.keep_newest=2" {
		t.Errorf("cleaned %v", items)
	}
	if len(kept.Sessions) != 2 {
		t.Errorf("kept %v", kept.Sessions)
	}
}

// The default keeps every row, and a table it keeps is given back as it was.
func TestRowsDefaultKeeps(t *testing.T) {
	l := project(t, "")
	table := &format.Sessions{Sessions: []format.SessionRow{row(sid("1"), "none", daysAgo(400), daysAgo(399))}}
	kept, items, err := Rows(opts(l), table, nil)
	if err != nil || len(items) != 0 || kept != table {
		t.Errorf("the default cleaned %v (%v)", items, err)
	}
}
