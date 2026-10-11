package clean

// A keep_days past what a time.Duration holds (maxKeepDays, about 292 years): counted back as a duration it would wrap
// round to a cutoff in the future, and every file and row the protections do not hold would go. It cleans nothing by
// age (rule.go).

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/sessions"
)

func TestAHugeKeepDaysCleansNothing(t *testing.T) {
	for _, days := range []string{strconv.FormatInt(maxKeepDays+1, 10), "1000000"} {
		t.Run(days, func(t *testing.T) {
			l := project(t, strings.ReplaceAll("  log:\n    keep_days: D\n  asks:\n    keep_days: D\n  ladder:\n"+
				"    keep_days: D\n  sessions:\n    keep_days: D\n", "D", days))
			fixture(t, l)
			if d := SessionEnd(opts(l)); len(d.Cleaned) != 0 || len(d.Errs) != 0 {
				t.Errorf("the files: cleaned %v, errors %v", targets(d.Cleaned), d.Errs)
			}
			table := &format.Sessions{Sessions: []format.SessionRow{
				row(sid("1"), "T-0001", daysAgo(41), daysAgo(40)), // its task done: would go by an age rule
				row(sid("3"), "none", daysAgo(41), daysAgo(40)),   // no task: would go by an age rule
			}}
			table.Hours = sessions.Hours(table.Sessions)
			if _, items, err := Rows(opts(l), table, nil); err != nil || len(items) != 0 {
				t.Errorf("the rows: cleaned %v, error %v", items, err)
			}
		})
	}
}

// The largest keep_days that counts back still has its cutoff; one more has none.
func TestTheCutoffAtItsLimit(t *testing.T) {
	days := maxKeepDays
	at, byAge := Rule{KeepDays: &days}.cutoff(now)
	if want := now.Add(-time.Duration(days) * 24 * time.Hour); !byAge || !at.Equal(want) || !at.Before(now) {
		t.Errorf("keep_days %d: cutoff %v (by age %v), want %v", days, at, byAge, want)
	}
	more := maxKeepDays + 1
	if _, byAge := (Rule{KeepDays: &more}).cutoff(now); byAge {
		t.Errorf("keep_days %d has a cutoff", more)
	}
	if _, byAge := (Rule{}).cutoff(now); byAge {
		t.Error("no keep_days has a cutoff")
	}
}
