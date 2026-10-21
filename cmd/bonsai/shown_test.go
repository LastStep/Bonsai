package main

// Printed times in the person's zone (T-5714; design/contract.md section 23): `when`, --utc, logs --day as a local
// day over UTC-day files, and the trust line of message --show. The zone is injected (nowZone): Go on Windows ignores TZ.

import (
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/record"
	"github.com/LastStep/Bonsai/internal/schema"
)

// TestWhen: a stored instant is shown in the zone with its offset, a day crossing UTC midnight changes its date, --utc
// leaves it as stored, and a string that is no stored instant is printed as it is.
func TestWhen(t *testing.T) {
	defer func(f func() *time.Location) { nowZone = f }(nowZone)
	nowZone = func() *time.Location { return time.FixedZone("west", -(7 * 3600)) }
	local := &call{}
	utc := &call{given: []given{{name: "--utc"}}}
	for in, want := range map[string]string{
		"2026-10-21T03:30:00.000Z":  "2026-10-20 20:30 -07:00", // the local date is the day before
		"2026-10-21T09:00:00Z":      "2026-10-21 02:00 -07:00",
		"not a time":                "not a time",
		"2026-10-21T09:00:00+02:00": "2026-10-21T09:00:00+02:00", // the stored form is UTC: a numeric offset is not one
		"":                          "",
	} {
		if got := local.when(in); got != want {
			t.Errorf("when(%q) = %q, want %q", in, got, want)
		}
		if got := utc.when(in); got != in {
			t.Errorf("--utc when(%q) = %q, want it as stored", in, got)
		}
	}
}

// writeDayFile writes the day file of a UTC day holding one record at each clock time of that day.
func (c *cli) writeDayFile(day string, clocks ...string) {
	c.t.Helper()
	var b []byte
	for _, at := range clocks {
		ts, err := time.Parse(record.AtLayout, day+"T"+at+":00.000Z")
		if err != nil {
			c.t.Fatal(err)
		}
		line, err := record.Line(record.New("event", record.Common{Workspace: logWS, Agent: "claude-code", At: ts}))
		if err != nil {
			c.t.Fatal(err)
		}
		b = append(b, line...)
	}
	c.write(".bonsai/local/log/w-"+day+".ndjson", string(b))
}

// TestLogsLocalDay: in a zone 5:30 ahead, the local day 2026-10-09 starts at 18:30 UTC on the 8th, so it reads two UTC-day
// files and keeps the records of its instants; with --utc the day is the one file.
func TestLogsLocalDay(t *testing.T) {
	c := newCLI(t)
	c.link()
	c.commit("link")
	c.writeDayFile("2026-10-08", "17:00", "19:00") // local 22:30 on the 8th, 00:30 on the 9th
	c.writeDayFile("2026-10-09", "03:00", "19:00") // local 08:30 on the 9th, 00:30 on the 10th
	count := func(args ...string) (int, []string) {
		code, out, _ := c.run("", append([]string{"logs", "--json"}, args...)...)
		if code != 0 {
			t.Fatalf("logs %v: %d\n%s", args, code, out)
		}
		doc, files := logsOf(t, out)
		recs, _ := doc.Get("records")
		var ats []string
		list, _ := recs.([]any)
		for _, r := range list {
			ats = append(ats, r.(schema.Object).String("at"))
		}
		if code != 0 {
			t.Fatalf("logs %v: %d\n%s", args, code, out)
		}
		_ = files
		return len(files), ats
	}
	if n, ats := count("--day", "2026-10-09"); n != 2 || strings.Join(ats, ",") != "2026-10-08T19:00:00.000Z,2026-10-09T03:00:00.000Z" {
		t.Errorf("the local 9th: %d files %v", n, ats)
	}
	if n, ats := count("--day", "2026-10-08"); n != 1 || strings.Join(ats, ",") != "2026-10-08T17:00:00.000Z" {
		t.Errorf("the local 8th: %d files %v", n, ats)
	}
	if n, ats := count("--day", "2026-10-09", "--utc"); n != 1 || strings.Join(ats, ",") != "2026-10-09T03:00:00.000Z,2026-10-09T19:00:00.000Z" {
		t.Errorf("the UTC 9th: %d files %v", n, ats)
	}
	if n, ats := count("--day", "2026-10-11"); n != 0 || len(ats) != 0 {
		t.Errorf("a day with no file: %d files %v", n, ats)
	}
	// The listing shows each file's first and last in the zone; --utc as stored.
	_, out, _ := c.run("", "logs")
	if !strings.Contains(out, "2026-10-08 22:30 +05:30 to 2026-10-09 00:30 +05:30") {
		t.Errorf("the listing is not local:\n%s", out)
	}
	if _, out, _ = c.run("", "logs", "--utc"); !strings.Contains(out, "2026-10-08T17:00:00.000Z to 2026-10-08T19:00:00.000Z") {
		t.Errorf("--utc listing is not as stored:\n%s", out)
	}
	if code, out, errOut := c.run("", "logs", "--day", "10-09"); code != 2 || !strings.Contains(out+errOut, "your zone") {
		t.Errorf("a bad day: %d %q %q", code, out, errOut)
	}
}
