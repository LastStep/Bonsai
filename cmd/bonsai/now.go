package main

// bonsai now (contract section 23, "Time"): the current time in each form Bonsai writes, so no agent invents one.
// It reads the clock and the machine's zone only: no workspace, no file, no network, so it runs anywhere, linked or
// not. The forms are internal/timefmt's. Its table of flags and exit codes is nowWord.

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/LastStep/Bonsai/internal/engine"
	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/timefmt"
)

func init() { register(nowWord) }

// nowClock and nowZone are where the word reads the time and the zone; a test replaces them (Go on Windows ignores
// TZ, so a test cannot set the zone through the environment).
var (
	nowClock = time.Now
	nowZone  = func() *time.Location { return time.Local }
)

var nowWord = &Word{
	Name:    "now",
	Order:   18,
	Title:   "the current time, in the form Bonsai writes it (contract section 23).",
	Summary: "print the current time to write in a document",
	Args:    "[flags]",
	About: `Bonsai is the only clock: take every time you write in a file from here, never from memory or a guess. By
default it prints the time for a person to read, your zone with its numeric offset ("2026-10-21 12:49 +05:30").
A run report's started and finished hold the time alone ("2026-10-21 12:49") and the report's offset field the
offset ("+05:30"): take them apart from --json (local and offset). --day prints the local day ("2026-10-21"), the
form of a task's created, started and finished, STATE's and a memory note's updated. --utc prints the instant as a
machine record stores it, UTC with milliseconds ("2026-10-21T07:19:04.325Z"). --json prints every form in the
bonsai.now/1 document: local (the time with no offset), utc, day, offset (its own field) and zone (the zone's
abbreviation, for a person's eye: never write it in a file, and never read a time from it). It reads the clock
and this machine's zone and nothing else: no workspace, no file, no network.
`,
	Flags: []Flag{
		{Name: "--utc", Help: "print the instant in UTC with milliseconds, the form machine records store"},
		{Name: "--day", Help: "print the local day, YYYY-MM-DD (not with --utc)"},
		{Name: "--json", Help: "print the bonsai.now/1 document (for programs) instead of plain text; it holds every form, so\n--utc and --day change nothing in it"},
	},
	Exits: []Exit{
		{Code: 0, Means: "the time was printed"},
		{Code: 2, Means: "bad input (a flag now does not take, or --utc with --day)"},
		{Code: 3, Means: "the answer cannot be printed"},
	},
	Examples: []string{"bonsai now", "bonsai now --day", "bonsai now --utc", "bonsai now --json"},
	Refused: func(c *call, e *engine.Error) encoder {
		return &nowDoc{Error: e.Object()}
	},
	Run: runNow,
}

func runNow(c *call) int {
	if len(c.rest) > 0 {
		return c.refuse(c.flagError("now takes no %+q", c.rest[0]))
	}
	if c.has("--utc") && c.has("--day") {
		e := c.flagError("now --utc prints an instant and --day a day: they cannot be asked together")
		e.Next = "run `bonsai now --utc` or `bonsai now --day`, not both"
		return c.refuse(e)
	}
	t, loc := nowClock(), nowZone()
	if c.json {
		abbr, _ := t.In(loc).Zone()
		return c.printDoc(&nowDoc{
			Local: strp(timefmt.DocTime(t, loc)), UTC: strp(timefmt.UTC(t)), Day: strp(timefmt.Day(t, loc)),
			Offset: strp(timefmt.Offset(t, loc)), Zone: strp(engine.ASCII(abbr)),
		}, exitOK)
	}
	switch {
	case c.has("--utc"):
		return c.print(timefmt.UTC(t) + "\n")
	case c.has("--day"):
		return c.print(timefmt.Day(t, loc) + "\n")
	}
	return c.print(strings.TrimSpace(timefmt.Doc(t, loc)) + "\n")
}

func strp(s string) *string { return &s }

// nowDoc is bonsai.now/1's document, as written until format.Now takes its place (piece A's schema).
type nowDoc struct {
	Local  *string             `json:"local"`
	UTC    *string             `json:"utc"`
	Day    *string             `json:"day"`
	Offset *string             `json:"offset"`
	Zone   *string             `json:"zone"`
	Error  *format.ErrorObject `json:"error"`
}

func (d *nowDoc) Encode() ([]byte, error) {
	b, err := json.MarshalIndent(struct {
		Format string `json:"format"`
		*nowDoc
	}{"bonsai.now/1", d}, "", "  ")
	return append(b, '\n'), err
}
