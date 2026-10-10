package clean

// The sessions kind (contract §7.5): the rows of .bonsai/sessions.md, cleaned only by check --write, the table's one
// writer (internal/engine), after it has added the rows the log lacks.
//
// Age is a row's end. The protections: a row whose task is not done or cut, as the ladder reads a task (ladder.go;
// none is never protected), and a row whose span is still in the log (in found, sessions.Found's rows): rows are only
// added, so check --write would add it again at its next run, and the log file holding it, judged by its rows, would
// then wait for it. Such a row goes once its log file has gone. A row whose end does not read is kept.
//
// The table is written first, then each row's clean record (Record), so a row's record follows its delete as a file's
// does. Its target is .bonsai/sessions.md# and the row's key as a person reads it: the session's 8 characters, a
// subagent run's 8 after a slash, and the start after an @ (.bonsai/sessions.md#a1b2c3d4/e5f6a7b8@2026-09-01 10:00).

import (
	"errors"
	"sort"
	"time"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/sessions"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// rowLayout is how a row writes its start and end (UTC, to the minute).
const rowLayout = "2006-01-02 15:04"

// Rows cleans the sessions table's rows by bonsai.yaml's generated.sessions: it gives the table to write, with the
// rows cleaned taken out and the hours below rebuilt from the rows that stay, and the rows cleaned, whose records
// Record writes once the table is written. found are the rows of every ended span in the log (sessions.Found). A rule
// that keeps, or does not read, cleans nothing (bonsai check names a rule that does not read).
func Rows(o Options, table *format.Sessions, found []format.SessionRow) (*format.Sessions, []Item, error) {
	r, d := start(o)
	if r == nil {
		return table, nil, errors.Join(d.Errs...)
	}
	rule, err := RuleOf(r.cfg.Doc, Sessions)
	if err != nil || rule.Keeps() || len(table.Sessions) == 0 {
		return table, nil, err
	}
	inLog := map[string]bool{}
	for _, row := range found {
		inLog[sessions.Key(row)] = true
	}
	type entry struct {
		i      int
		key    string
		end    time.Time
		ok     bool // the end reads
		reason string
	}
	entries := make([]*entry, len(table.Sessions))
	for i, row := range table.Sessions {
		end, err := time.Parse(rowLayout, row.End)
		entries[i] = &entry{i: i, key: sessions.Key(row), end: end, ok: err == nil}
	}
	if rule.KeepNewest != nil {
		byEnd := append([]*entry{}, entries...)
		sort.SliceStable(byEnd, func(i, j int) bool {
			a, b := byEnd[i], byEnd[j]
			if a.ok != b.ok {
				return !a.ok // a row whose end does not read counts as newest: it is kept anyway
			}
			if !a.end.Equal(b.end) {
				return a.end.After(b.end)
			}
			return a.key < b.key
		})
		for i, e := range byEnd {
			if int64(i) >= *rule.KeepNewest {
				e.reason = reason(Sessions, "keep_newest", *rule.KeepNewest)
			}
		}
	}
	if rule.KeepDays != nil {
		cutoff := r.now.Add(-time.Duration(*rule.KeepDays) * 24 * time.Hour)
		for _, e := range entries {
			if e.ok && e.end.Before(cutoff) {
				e.reason = reason(Sessions, "keep_days", *rule.KeepDays)
			}
		}
	}
	out := &format.Sessions{Sessions: []format.SessionRow{}, Extra: table.Extra}
	var items []Item
	for _, e := range entries {
		row := table.Sessions[e.i]
		if e.reason == "" || !e.ok || inLog[e.key] || (row.Task != "none" && !r.taskStates().done(row.Task)) {
			out.Sessions = append(out.Sessions, row)
			continue
		}
		items = append(items, Item{Kind: Sessions, Target: rowTarget(row), Reason: e.reason})
	}
	if len(items) == 0 {
		return table, nil, nil
	}
	out.Hours = sessions.Hours(out.Sessions)
	return out, items, nil
}

// rowTarget is a row's target: .bonsai/sessions.md#<session>[/<subagent>]@<start>.
func rowTarget(row format.SessionRow) string {
	key := row.Session
	if row.Subagent != nil {
		key += "/" + *row.Subagent
	}
	return workspace.SessionsTableFile + "#" + key + "@" + row.Start
}
