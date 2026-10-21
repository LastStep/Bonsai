package main

import (
	"time"

	"github.com/LastStep/Bonsai/internal/timefmt"
)

// utcFlag is the flag every command that prints a stored time takes: the stored (UTC) form instead of the person's
// local one (design/contract.md section 23).
var utcFlag = Flag{Name: "--utc", Help: "print each time as stored (UTC, ending in Z) instead of in your zone with its offset;\n--json always prints it as stored"}

// shownZone is the zone printed times are shown in: nowZone's (time.Local unless a test replaces it).
func shownZone() *time.Location { return nowZone() }

// when is a stored instant as printed for the person: the document form in their zone (timefmt.Show), or as stored
// with --utc. A string that is not a stored instant is printed as it is.
func (c *call) when(s string) string {
	if s == "" || c.has("--utc") {
		return s
	}
	t, err := timefmt.ParseUTC(s)
	if err != nil {
		return s
	}
	return timefmt.Show(t, shownZone(), false)
}
