// Package timefmt owns every form a time takes in Bonsai (design/contract.md, section 23, "Time"; P-T-5795's
// Amendment 2: every recorded time is UTC, and display converts). The recorded forms are UTC: the time in a file
// ("2026-10-21 08:19", Record), the day ("2026-10-21", RecordDay) and the instant a machine record holds
// ("2026-10-21T08:19:04.325Z", UTC). The display form converts to a person's zone with its offset
// ("2026-10-21 13:49 +05:30", Doc, Show). Bonsai is the only clock: the guard checks fresh stamps against what
// `bonsai now` prints, which is this package's output.
//
// Every function that depends on a zone takes a *time.Location: production code passes time.Local, a test passes one
// from time.LoadLocation (Go on Windows ignores TZ, so a test cannot set the zone through the environment). A zone
// is always shown as a numeric offset, never an abbreviation: Go reads an unknown abbreviation as offset zero and
// names are ambiguous, so a parser refuses one.
//
// Legacy forms still read and are never rewritten: a document time with no offset (written before formats set 11).
// In a run report it means "offset unknown" (ParseDoc says so); in .bonsai/sessions.md it is UTC (ParseSession).
package timefmt

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	docLayout    = "2006-01-02 15:04 -07:00" // a document time with its offset
	legacyLayout = "2006-01-02 15:04"        // a document time before set 11: no offset
	dayLayout    = "2006-01-02"
	milliLayout  = "2006-01-02T15:04:05.000Z"
	secondLayout = "2006-01-02T15:04:05Z"
)

// Doc is the document time of t in loc, to the minute, with a numeric offset: "2026-10-21 12:49 +05:30". A nil loc
// is UTC ("+00:00").
func Doc(t time.Time, loc *time.Location) string { return t.In(zone(loc)).Format(docLayout) }

// Day is the day of t in loc: "2026-10-21". A day Bonsai records is the UTC day (RecordDay); this one is for
// display and for reading a person's local day.
func Day(t time.Time, loc *time.Location) string { return t.In(zone(loc)).Format(dayLayout) }

// Record is the time Bonsai records in a file a person opens, always UTC, to the minute: "2026-10-21 08:19". A run
// report's started and finished and a sessions row's start and end hold it; Doc shows the same instant for a person.
func Record(t time.Time) string { return t.UTC().Format(legacyLayout) }

// RecordDay is the day Bonsai records, the UTC day: "2026-10-21" (a task's dates, STATE's and a memory note's updated).
func RecordDay(t time.Time) string { return t.UTC().Format(dayLayout) }

// ParseRecord reads a recorded time, "2026-10-21 08:19", as UTC. A time written before formats set 11 was local with
// the offset unknown and reads the same way: its digits as UTC, so a caller that needs to say "unknown" checks the
// report's date against the cutover.
func ParseRecord(s string) (time.Time, error) {
	t, err := time.Parse(legacyLayout, s)
	if err != nil {
		return time.Time{}, fmt.Errorf("%q is not a recorded time (YYYY-MM-DD HH:MM, UTC)", s)
	}
	return t, nil
}

// UTC is the instant form of t with milliseconds, in UTC: "2026-10-21T07:19:04.325Z".
func UTC(t time.Time) string { return t.UTC().Format(milliLayout) }

// UTCSeconds is the instant form of t in whole seconds, for the records that hold seconds today:
// "2026-10-21T07:19:04Z".
func UTCSeconds(t time.Time) string { return t.UTC().Format(secondLayout) }

// Offset is loc's offset from UTC at t as "+05:30" (the document form's last part).
func Offset(t time.Time, loc *time.Location) string { return t.In(zone(loc)).Format("-07:00") }

// Show is the time as printed for a person: the document form in loc, or with utc the stored instant form.
func Show(t time.Time, loc *time.Location, utc bool) string {
	if utc {
		return UTC(t)
	}
	return Doc(t, loc)
}

// ParseDoc reads a document time. A time with an offset gives its instant and known true. A legacy time (no offset)
// gives its digits read as UTC and known false: the caller decides what "offset unknown" means (a run report's
// started line is shown as written; the sessions table reads it as UTC, ParseSession). Anything else, including a
// zone name ("12:49 IST") or a "Z", is refused.
func ParseDoc(s string) (t time.Time, known bool, err error) {
	if t, err = time.Parse(docLayout, s); err == nil {
		return t, true, nil
	}
	if t, err = time.Parse(legacyLayout, s); err == nil {
		return t, false, nil
	}
	return time.Time{}, false, fmt.Errorf("%q is not a document time (YYYY-MM-DD HH:MM +HH:MM, a numeric offset; or the legacy YYYY-MM-DD HH:MM)", s)
}

// ParseSession reads a time of .bonsai/sessions.md: an old row has no offset and is UTC, a new one carries its
// offset. Rows compare by the instant this gives.
func ParseSession(s string) (time.Time, error) {
	t, _, err := ParseDoc(s)
	return t, err
}

// ParseDay reads a day and gives its first instant in loc (local midnight; a zone that skips midnight gives the
// first instant of the day).
func ParseDay(s string, loc *time.Location) (time.Time, error) {
	if len(s) != len(dayLayout) {
		return time.Time{}, fmt.Errorf("%q is not a day (YYYY-MM-DD)", s)
	}
	d, err := time.Parse(dayLayout, s)
	if err != nil {
		return time.Time{}, fmt.Errorf("%q is not a day (YYYY-MM-DD)", s)
	}
	l := zone(loc)
	t := time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, l)
	// A zone whose clock skips midnight (a spring-forward at 00:00): time.Date gives an instant of the day before,
	// so step to the day's first minute.
	for i := 0; i < 180 && t.In(l).Format(dayLayout) < s; i++ {
		t = t.Add(time.Minute)
	}
	return t, nil
}

// ParseUTC reads an instant of a machine record: RFC 3339 in UTC ("Z"), with or without fractional seconds. A
// numeric offset is refused, since the stored form is always UTC.
func ParseUTC(s string) (time.Time, error) {
	if !strings.HasSuffix(s, "Z") {
		return time.Time{}, errors.New(`"` + s + `" is not an instant (RFC 3339 in UTC, ending in Z)`)
	}
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		return time.Time{}, fmt.Errorf("%q is not an instant (RFC 3339 in UTC, ending in Z)", s)
	}
	return t.UTC(), nil
}

func zone(loc *time.Location) *time.Location {
	if loc == nil {
		return time.UTC
	}
	return loc
}
