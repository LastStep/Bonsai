package main

import (
	"strings"
	"testing"
	"time"
	_ "time/tzdata" // tests only: Go on Windows ignores TZ, so each test injects its zone
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
		zone, doc, day, off, abbr string
	}{
		{"Asia/Kolkata", "2026-10-21 12:49 +05:30", "2026-10-21", "+05:30", "IST"},
		{"UTC", "2026-10-21 07:19 +00:00", "2026-10-21", "+00:00", "UTC"},
		{"America/New_York", "2026-10-21 03:19 -04:00", "2026-10-21", "-04:00", "EDT"},
	}
	for _, c := range cases {
		withClock(t, at, c.zone)
		for _, w := range []struct {
			args []string
			want string
		}{
			{[]string{"now"}, c.doc},
			{[]string{"now", "--day"}, c.day},
			{[]string{"now", "--utc"}, "2026-10-21T07:19:04.325Z"},
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
		for _, line := range []string{`"format": "bonsai.now/1"`, `"local": "` + strings.TrimSuffix(c.doc, " "+c.off) + `"`, `"utc": "2026-10-21T07:19:04.325Z"`,
			`"day": "` + c.day + `"`, `"offset": "` + c.off + `"`, `"zone": "` + c.abbr + `"`, `"error": null`} {
			if !strings.Contains(out, line) {
				t.Errorf("%s --json lacks %s:\n%s", c.zone, line, out)
			}
		}
		asciiOnly(t, "bonsai now --json", out)
	}
}

// Around a local midnight the local day and the UTC instant name different days.
func TestNowAtLocalMidnight(t *testing.T) {
	withClock(t, time.Date(2026, 10, 21, 18, 30, 0, 0, time.UTC), "Asia/Kolkata")
	if _, out, _ := runArgs("now", "--day"); out != "2026-10-22\n" {
		t.Errorf("--day = %q", out)
	}
	if _, out, _ := runArgs("now", "--utc"); out != "2026-10-21T18:30:00.000Z\n" {
		t.Errorf("--utc = %q", out)
	}
	if _, out, _ := runArgs("now"); out != "2026-10-22 00:00 +05:30\n" {
		t.Errorf("default = %q", out)
	}
}

func TestNowRefusals(t *testing.T) {
	withClock(t, time.Date(2026, 10, 21, 7, 19, 0, 0, time.UTC), "UTC")
	for _, args := range [][]string{{"now", "--utc", "--day"}, {"now", "extra"}, {"now", "--nope"}} {
		code, out, errOut := runArgs(args...)
		if code != 2 || out != "" || !strings.Contains(errOut, "next") {
			t.Errorf("%v: exit %d, out %q, stderr %q", args, code, out, errOut)
		}
		code, out, errOut = runArgs(append(args, "--json")...)
		if code != 2 || errOut != "" || !strings.Contains(out, `"error": {`) || !strings.Contains(out, `"local": null`) {
			t.Errorf("%v --json: exit %d, out %q, stderr %q", args, code, out, errOut)
		}
	}
}
