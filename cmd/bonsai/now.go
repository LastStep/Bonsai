package main

// bonsai now (contract section 23, "Time"): the current time in each form Bonsai writes, so no agent invents one.
// It reads the clock and the machine's zone only: no workspace, no file, no network, so it runs anywhere, linked or
// not. The forms are internal/timefmt's. Its table of flags and exit codes is nowWord.

import (
	"encoding/json"
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
	About: `Bonsai is the only clock, and it records in UTC: take every time you write in a file from here, never from
memory or a guess. By default it prints the time to record, in UTC to the minute ("2026-10-21 08:19"): a run
report's started and finished. --day prints the UTC day ("2026-10-21"): a task's created, started and finished,
STATE's and a memory note's updated. --iso prints the instant as a machine record stores it, UTC with
milliseconds ("2026-10-21T08:19:04.325Z"). --local prints the same instant for a person to read, in this
machine's zone with its numeric offset ("2026-10-21 13:49 +05:30"); it is never what you write in a file. At most
one of the four. --json prints every form in the bonsai.now/1 document: utc, day, iso, local, offset (the local
offset) and zone (the zone's abbreviation, for a person's eye: never write it in a file, and never read a time
from it). It reads the clock and this machine's zone and nothing else: no workspace, no file, no network.
`,
	Flags: []Flag{
		{Name: "--day", Help: "print the UTC day, YYYY-MM-DD (not with --iso or --local)"},
		{Name: "--iso", Help: "print the instant in UTC with milliseconds, the form machine records store (not with --day or --local)"},
		{Name: "--local", Help: "print the time in this machine's zone with its numeric offset, for a person to read (not with --day\nor --iso)"},
		{Name: "--json", Help: "print the bonsai.now/1 document (for programs) instead of plain text; it holds every form, so\n--day, --iso and --local change nothing in it"},
	},
	Exits: []Exit{
		{Code: 0, Means: "the time was printed"},
		{Code: 2, Means: "bad input (a flag now does not take, or two of --day, --iso and --local)"},
		{Code: 3, Means: "the answer cannot be printed"},
	},
	Examples: []string{"bonsai now", "bonsai now --day", "bonsai now --iso", "bonsai now --local", "bonsai now --json"},
	Refused: func(c *call, e *engine.Error) encoder {
		return &nowDoc{Error: e.Object()}
	},
	Run: runNow,
}

func runNow(c *call) int {
	if len(c.rest) > 0 {
		return c.refuse(c.flagError("now takes no %+q", c.rest[0]))
	}
	asked := 0
	for _, f := range []string{"--day", "--iso", "--local"} {
		if c.has(f) {
			asked++
		}
	}
	if asked > 1 {
		e := c.flagError("now prints one form: --day, --iso and --local cannot be asked together")
		e.Next = "run `bonsai now`, `bonsai now --day`, `bonsai now --iso` or `bonsai now --local`, one at a time"
		return c.refuse(e)
	}
	t, loc := nowClock(), nowZone()
	if c.json {
		abbr, _ := t.In(loc).Zone()
		return c.printDoc(&nowDoc{
			UTC: strp(timefmt.Record(t)), Day: strp(timefmt.RecordDay(t)), ISO: strp(timefmt.UTC(t)),
			Local: strp(timefmt.Doc(t, loc)), Offset: strp(timefmt.Offset(t, loc)), Zone: strp(engine.ASCII(abbr)),
		}, exitOK)
	}
	switch {
	case c.has("--iso"):
		return c.print(timefmt.UTC(t) + "\n")
	case c.has("--day"):
		return c.print(timefmt.RecordDay(t) + "\n")
	case c.has("--local"):
		return c.print(timefmt.Doc(t, loc) + "\n")
	}
	return c.print(timefmt.Record(t) + "\n")
}

func strp(s string) *string { return &s }

// nowDoc is bonsai.now/1's document, as written until format.Now takes its place (piece A's schema).
type nowDoc struct {
	UTC    *string             `json:"utc"`
	Day    *string             `json:"day"`
	ISO    *string             `json:"iso"`
	Local  *string             `json:"local"`
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
