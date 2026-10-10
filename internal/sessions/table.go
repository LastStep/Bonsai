package sessions

// The rows and hours of .bonsai/sessions.md (spec section 6, "The two tables"; plan-5 5.2.3 notes 2 and 3).
//
// A row's key is its session (the id's first 8 characters), its subagent (its first 8, "" for a session's row) and its
// start. A key already in the table is never added again or rewritten, and a row whose log file is gone stays: rows
// are only added. Rows are sorted by start, then session, then subagent: the same bytes come out whoever builds them,
// on either side, with no map order in any of it.
//
// The hours add up the minutes of the rows as they are shown, per task, role and kind, in tenths of an hour rounded
// half up; a session's hours and a subagent run's are never added together (a run lies inside its session's row).
// bonsai.sessions/1's hours table has no place for a task's total of each kind (a row with a null role is "rows with
// no role"), so none is written.

import (
	"encoding/json"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/record"
)

// Found gives the rows of every ended span in the session files of the log folder dir, as of now, sorted and with no
// key twice. A folder that is not there holds none. A file that cannot be read is an error.
func Found(dir string, now time.Time) ([]format.SessionRow, error) {
	files, err := record.ListLog(dir)
	if err != nil {
		return nil, err
	}
	var rows []format.SessionRow
	for _, name := range files.Sessions {
		f, err := Read(filepath.Join(dir, name), now)
		if err != nil {
			return nil, err
		}
		for _, s := range f.Spans {
			if s.Open {
				continue
			}
			if r, ok := rowOf(s); ok {
				rows = append(rows, r)
			}
		}
	}
	return sortRows(dedupe(rows, nil)), nil
}

// Key is a row's key: its session, subagent and start.
func Key(r format.SessionRow) string {
	sub := ""
	if r.Subagent != nil {
		sub = *r.Subagent
	}
	return r.Session + "\x00" + sub + "\x00" + r.Start
}

// dedupe keeps the first row of each key, and none whose key is in have.
func dedupe(rows, have []format.SessionRow) []format.SessionRow {
	seen := map[string]bool{}
	for _, r := range have {
		seen[Key(r)] = true
	}
	var out []format.SessionRow
	for _, r := range rows {
		if k := Key(r); !seen[k] {
			seen[k] = true
			out = append(out, r)
		}
	}
	return out
}

// Missing gives the rows of found that the table's rows lack.
func Missing(table, found []format.SessionRow) []format.SessionRow { return dedupe(found, table) }

// sortRows sorts by start, then session, then subagent. The sort is stable, so rows a hand edit left with one key
// keep their order.
func sortRows(rows []format.SessionRow) []format.SessionRow {
	sort.SliceStable(rows, func(i, j int) bool {
		a, b := rows[i], rows[j]
		if a.Start != b.Start {
			return a.Start < b.Start
		}
		if a.Session != b.Session {
			return a.Session < b.Session
		}
		return subagentOf(a) < subagentOf(b)
	})
	return rows
}

func subagentOf(r format.SessionRow) string {
	if r.Subagent == nil {
		return ""
	}
	return *r.Subagent
}

// Merge gives the table with the rows of found it lacks added, its own rows untouched, sorted, and the hours below
// them rebuilt from every row. Fields the table holds that this Bonsai does not know (Extra) are kept.
func Merge(table *format.Sessions, found []format.SessionRow) *format.Sessions {
	rows := append(append([]format.SessionRow{}, table.Sessions...), Missing(table.Sessions, found)...)
	rows = sortRows(rows)
	return &format.Sessions{Sessions: rows, Hours: Hours(rows), Extra: table.Extra}
}

// Hours adds up the rows' minutes per task, role and kind, sorted by task, then role (none first), then kind.
func Hours(rows []format.SessionRow) []format.HoursRow {
	type key struct{ task, role, kind string }
	type sum struct {
		role    *string
		minutes int64
	}
	totals := map[key]*sum{}
	var keys []key
	for _, r := range rows {
		role := ""
		if r.Role != nil {
			role = *r.Role
		}
		k := key{r.Task, role, r.Kind}
		if totals[k] == nil {
			totals[k] = &sum{role: r.Role}
			keys = append(keys, k)
		}
		totals[k].minutes += r.Minutes
	}
	sort.Slice(keys, func(i, j int) bool { // the keys are distinct
		a, b := keys[i], keys[j]
		if a.task != b.task {
			return a.task < b.task
		}
		if a.role != b.role {
			return a.role < b.role
		}
		return a.kind < b.kind
	})
	out := []format.HoursRow{}
	for _, k := range keys {
		s := totals[k]
		out = append(out, format.HoursRow{Task: k.task, Role: s.role, Kind: k.kind, Hours: tenths(s.minutes)})
	}
	return out
}

// tenths gives minutes as hours to one decimal, half up, with integer arithmetic only.
func tenths(minutes int64) json.Number {
	t := (minutes*10 + 30) / 60
	return json.Number(strconv.FormatInt(t/10, 10) + "." + strconv.FormatInt(t%10, 10))
}
