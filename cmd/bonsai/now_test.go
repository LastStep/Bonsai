package main

import (
	"strings"
	"testing"
	"time"
	_ "time/tzdata" // tests only: Go on Windows ignores TZ, so each test injects its zone

	"github.com/LastStep/Bonsai/internal/format"
)

// withClock fixes bonsai now's clock and zone for one test.
func withClock(t *testing.T, at time.Time, zone string) {
	t.Helper()
	loc, err := time.LoadLocation(zone)
	if err != nil {
		t.Fatal(err)
	}
	oc, oz := nowClock, nowZone
	nowClock, nowZone = func() time.Time { return at }, func() *time.Location { return loc }
	t.Cleanup(func() { nowClock, nowZone = oc, oz })
}

func TestNowForms(t *testing.T) {
	at := time.Date(2026, 10, 21, 7, 19, 4, 325_000_000, time.UTC)
	cases := []struct {
		zone, local, off, abbr string
	}{
		{"Asia/Kolkata", "2026-10-21 12:49 +05:30", "+05:30", "IST"},
		{"UTC", "2026-10-21 07:19 +00:00", "+00:00", "UTC"},
		{"America/New_York", "2026-10-21 03:19 -04:00", "-04:00", "EDT"},
	}
	for _, c := range cases {
		withClock(t, at, c.zone)
		// What is recorded is UTC whatever the zone; --local alone follows it.
		for _, w := range []struct {
			args []string
			want string
		}{
			{[]string{"now"}, "2026-10-21 07:19"},
			{[]string{"now", "--day"}, "2026-10-21"},
			{[]string{"now", "--iso"}, "2026-10-21T07:19:04.325Z"},
			{[]string{"now", "--local"}, c.local},
		} {
			code, out, errOut := runArgs(w.args...)
			if code != 0 || errOut != "" || out != w.want+"\n" {
				t.Errorf("%s %v: exit %d, out %q, stderr %q; want %q", c.zone, w.args, code, out, errOut, w.want)
			}
		}
		code, out, errOut := runArgs("now", "--json")
		if code != 0 || errOut != "" {
			t.Fatalf("%s --json: exit %d, stderr %q", c.zone, code, errOut)
		}
		for _, line := range []string{`"format": "bonsai.now/1"`, `"utc": "2026-10-21 07:19"`, `"day": "2026-10-21"`,
			`"iso": "2026-10-21T07:19:04.325Z"`, `"local": "` + c.local + `"`, `"offset": "` + c.off + `"`, `"zone": "` + c.abbr + `"`,
			`"error": null`} {
			if !strings.Contains(out, line) {
				t.Errorf("%s --json lacks %s:\n%s", c.zone, line, out)
			}
		}
		asciiOnly(t, "bonsai now --json", out)
		// Held to bonsai.now/1 as the schema says it, every field in order, and read back by the format's reader.
		fits(t, out, "now")
		n, err := format.ReadNow([]byte(out))
		if err != nil || n.Error != nil || n.UTC == nil || *n.UTC != "2026-10-21 07:19" || *n.Local != c.local {
			t.Errorf("%s: bonsai.now/1 reads as %+v, %v", c.zone, n, err)
		}
	}
}

// Around a local midnight the recorded day stays the UTC day while the local display moves to the next.
func TestNowAtLocalMidnight(t *testing.T) {
	withClock(t, time.Date(2026, 10, 21, 18, 30, 0, 0, time.UTC), "Asia/Kolkata")
	for args, want := range map[string]string{"--day": "2026-10-21", "--iso": "2026-10-21T18:30:00.000Z", "--local": "2026-10-22 00:00 +05:30", "": "2026-10-21 18:30"} {
		argv := []string{"now"}
		if args != "" {
			argv = append(argv, args)
		}
		if _, out, _ := runArgs(argv...); out != want+"\n" {
			t.Errorf("%v = %q, want %q", argv, out, want)
		}
	}
}

func TestNowRefusals(t *testing.T) {
	withClock(t, time.Date(2026, 10, 21, 7, 19, 0, 0, time.UTC), "UTC")
	for _, args := range [][]string{{"now", "--iso", "--day"}, {"now", "--local", "--iso"}, {"now", "--day", "--local"}, {"now", "extra"}, {"now", "--utc"}} {
		code, out, errOut := runArgs(args...)
		if code != 2 || out != "" || !strings.Contains(errOut, "next") {
			t.Errorf("%v: exit %d, out %q, stderr %q", args, code, out, errOut)
		}
		code, out, errOut = runArgs(append(args, "--json")...)
		if code != 2 || errOut != "" || !strings.Contains(out, `"error": {`) || !strings.Contains(out, `"utc": null`) {
			t.Errorf("%v --json: exit %d, out %q, stderr %q", args, code, out, errOut)
		}
		fits(t, out, "now")
	}
}
