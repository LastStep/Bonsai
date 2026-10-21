package timefmt

import (
	"testing"
	"time"
	_ "time/tzdata" // tests only: Go on Windows ignores TZ, so each test injects its zone
)

func loc(t *testing.T, name string) *time.Location {
	t.Helper()
	l, err := time.LoadLocation(name)
	if err != nil {
		t.Fatal(err)
	}
	return l
}

func TestFormsInThreeZones(t *testing.T) {
	// 2026-10-21 07:19:04.325 UTC
	at := time.Date(2026, 10, 21, 7, 19, 4, 325_000_000, time.UTC)
	cases := []struct {
		zone, doc, day, off string
	}{
		{"Asia/Kolkata", "2026-10-21 12:49 +05:30", "2026-10-21", "+05:30"},
		{"UTC", "2026-10-21 07:19 +00:00", "2026-10-21", "+00:00"},
		{"America/New_York", "2026-10-21 03:19 -04:00", "2026-10-21", "-04:00"},
		{"Pacific/Kiritimati", "2026-10-21 21:19 +14:00", "2026-10-21", "+14:00"},
	}
	for _, c := range cases {
		l := loc(t, c.zone)
		if got := Doc(at, l); got != c.doc {
			t.Errorf("%s Doc = %q, want %q", c.zone, got, c.doc)
		}
		if got := Day(at, l); got != c.day {
			t.Errorf("%s Day = %q, want %q", c.zone, got, c.day)
		}
		if got := Offset(at, l); got != c.off {
			t.Errorf("%s Offset = %q, want %q", c.zone, got, c.off)
		}
		if got := Show(at, l, false); got != c.doc {
			t.Errorf("%s Show = %q", c.zone, got)
		}
		if got := Show(at, l, true); got != "2026-10-21T07:19:04.325Z" {
			t.Errorf("%s Show utc = %q", c.zone, got)
		}
	}
	if got := UTCSeconds(at); got != "2026-10-21T07:19:04Z" {
		t.Errorf("UTCSeconds = %q", got)
	}
	if Doc(at, nil) != "2026-10-21 07:19 +00:00" {
		t.Errorf("nil location is not UTC")
	}
}

func TestMidnightEdges(t *testing.T) {
	ist := loc(t, "Asia/Kolkata")
	// 18:29:59 UTC is 23:59 IST on the 21st; one second later it is the 22nd in Kolkata, still the 21st in UTC.
	before := time.Date(2026, 10, 21, 18, 29, 59, 0, time.UTC)
	after := before.Add(time.Second)
	if Day(before, ist) != "2026-10-21" || Day(after, ist) != "2026-10-22" {
		t.Errorf("Kolkata days: %s, %s", Day(before, ist), Day(after, ist))
	}
	if Doc(after, ist) != "2026-10-22 00:00 +05:30" {
		t.Errorf("Doc after midnight = %q", Doc(after, ist))
	}
	if Day(after, time.UTC) != "2026-10-21" {
		t.Errorf("UTC day moved")
	}
	// A west zone: 02:30 UTC on the 22nd is still the 21st in New York.
	ny := loc(t, "America/New_York")
	if Day(time.Date(2026, 10, 22, 2, 30, 0, 0, time.UTC), ny) != "2026-10-21" {
		t.Errorf("New York day")
	}
	// The year turns.
	if Day(time.Date(2026, 12, 31, 20, 0, 0, 0, time.UTC), ist) != "2027-01-01" {
		t.Errorf("year edge")
	}
}

func TestDST(t *testing.T) {
	ny := loc(t, "America/New_York")
	// 2026-11-01 is the fall-back day: 05:30 UTC is 01:30 EDT (-04:00), 06:30 UTC is 01:30 EST (-05:00).
	a := time.Date(2026, 11, 1, 5, 30, 0, 0, time.UTC)
	b := a.Add(time.Hour)
	if Doc(a, ny) != "2026-11-01 01:30 -04:00" || Doc(b, ny) != "2026-11-01 01:30 -05:00" {
		t.Errorf("fall back: %q, %q", Doc(a, ny), Doc(b, ny))
	}
	// The same wall time is two instants, told apart by the offset: both parse back to their own instant.
	for _, x := range []time.Time{a, b} {
		got, known, err := ParseDoc(Doc(x, ny))
		if err != nil || !known || !got.Equal(x) {
			t.Errorf("round trip of %s: %v %v %v", x, got, known, err)
		}
	}
	// Spring forward 2026-03-08: 07:30 UTC is 03:30 EDT; 06:59 UTC is 01:59 EST.
	if got := Doc(time.Date(2026, 3, 8, 7, 30, 0, 0, time.UTC), ny); got != "2026-03-08 03:30 -04:00" {
		t.Errorf("spring forward: %q", got)
	}
	if got := Doc(time.Date(2026, 3, 8, 6, 59, 0, 0, time.UTC), ny); got != "2026-03-08 01:59 -05:00" {
		t.Errorf("before spring forward: %q", got)
	}
	// A day that skips midnight (Sao Paulo skipped it in past years; Cuba still shifts at 00:00): ParseDay gives a
	// real first instant of that day.
	hav := loc(t, "America/Havana")
	d, err := ParseDay("2026-03-08", hav)
	if err != nil || Day(d, hav) != "2026-03-08" || Day(d.Add(-time.Minute), hav) != "2026-03-07" || Doc(d, hav) != "2026-03-08 01:00 -04:00" {
		t.Errorf("ParseDay in Havana: %v %v", d, err)
	}
	// Sydney's day lengths: 23 hours at spring forward.
	syd := loc(t, "Australia/Sydney")
	s, _ := ParseDay("2026-10-04", syd)
	n, _ := ParseDay("2026-10-05", syd)
	if n.Sub(s) != 23*time.Hour {
		t.Errorf("Sydney 4 Oct 2026 lasts %v", n.Sub(s))
	}
}

func TestParseDoc(t *testing.T) {
	got, known, err := ParseDoc("2026-10-21 12:49 +05:30")
	if err != nil || !known {
		t.Fatalf("with offset: %v %v", known, err)
	}
	if want := time.Date(2026, 10, 21, 7, 19, 0, 0, time.UTC); !got.Equal(want) {
		t.Errorf("instant = %v, want %v", got, want)
	}
	got, known, err = ParseDoc("2026-10-21 12:49")
	if err != nil || known {
		t.Fatalf("legacy: known %v, err %v", known, err)
	}
	if want := time.Date(2026, 10, 21, 12, 49, 0, 0, time.UTC); !got.Equal(want) {
		t.Errorf("legacy digits read as UTC = %v", got)
	}
	for _, bad := range []string{"", "2026-10-21", "2026-10-21 12:49 IST", "2026-10-21 12:49 Z", "2026-10-21T12:49 +05:30",
		"2026-10-21 12:49 +0530", "2026-10-21 12:49:00 +05:30", "2026-10-21 24:00 +05:30", "2026-13-01 12:49", "2026-10-21 12:49 +05:30 ", "26-10-21 12:49"} {
		if _, _, err := ParseDoc(bad); err == nil {
			t.Errorf("ParseDoc(%q) was accepted", bad)
		}
	}
}

func TestParseSessionLegacyIsUTC(t *testing.T) {
	old, err := ParseSession("2026-10-21 07:19")
	if err != nil {
		t.Fatal(err)
	}
	neu, err := ParseSession("2026-10-21 12:49 +05:30")
	if err != nil {
		t.Fatal(err)
	}
	if !old.Equal(neu) {
		t.Errorf("an old row and a new row of the same instant differ: %v, %v", old, neu)
	}
	if _, err := ParseSession("yesterday"); err == nil {
		t.Error("garbage accepted")
	}
}

func TestParseDay(t *testing.T) {
	ist := loc(t, "Asia/Kolkata")
	d, err := ParseDay("2026-10-21", ist)
	if err != nil {
		t.Fatal(err)
	}
	if UTC(d) != "2026-10-20T18:30:00.000Z" {
		t.Errorf("local midnight = %s", UTC(d))
	}
	for _, bad := range []string{"", "2026-1-21", "2026-10-21 12:49", "2026-02-30", "21-10-2026"} {
		if _, err := ParseDay(bad, ist); err == nil {
			t.Errorf("ParseDay(%q) was accepted", bad)
		}
	}
}

func TestParseUTC(t *testing.T) {
	for in, want := range map[string]string{
		"2026-10-21T07:19:04.325Z":    "2026-10-21T07:19:04.325Z",
		"2026-10-21T07:19:04Z":        "2026-10-21T07:19:04.000Z",
		"2026-10-21T07:19:04.123456Z": "2026-10-21T07:19:04.123Z",
	} {
		got, err := ParseUTC(in)
		if err != nil || UTC(got) != want {
			t.Errorf("ParseUTC(%q) = %v, %v; want %s", in, got, err, want)
		}
	}
	for _, bad := range []string{"", "2026-10-21T12:49:04+05:30", "2026-10-21 07:19", "2026-10-21T07:19:04", "2026-10-21T07:19Z"} {
		if _, err := ParseUTC(bad); err == nil {
			t.Errorf("ParseUTC(%q) was accepted", bad)
		}
	}
}
