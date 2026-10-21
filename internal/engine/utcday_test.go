package engine

import (
	"testing"
	"time"
	_ "time/tzdata" // tests only: Go on Windows ignores TZ, so a test sets the zone itself
)

// inZone makes zone the machine's local zone for one test (time.Local is a variable, so this holds on Windows too).
func inZone(t *testing.T, zone string) {
	t.Helper()
	loc, err := time.LoadLocation(zone)
	if err != nil {
		t.Fatal(err)
	}
	was := time.Local
	time.Local = loc
	t.Cleanup(func() { time.Local = was })
}

// Bonsai records days in UTC: the day init writes into STATE, and the retention cutoff of run reports, do not follow
// the machine's zone. Run at 20:00 UTC on the 21st it is already the 22nd in Kolkata and still the 21st in New York.
func TestDaysAreUTCWhateverTheZone(t *testing.T) {
	at := time.Date(2026, 10, 21, 20, 0, 0, 0, time.UTC)
	for _, zone := range []string{"Asia/Kolkata", "America/New_York", "UTC", "Pacific/Kiritimati"} {
		inZone(t, zone)
		was := stateClock
		stateClock = func() time.Time { return at.In(time.Local) }
		got := today()
		stateClock = was
		if got != "2026-10-21" {
			t.Errorf("%s: STATE's day = %s, the UTC day is 2026-10-21", zone, got)
		}
		if got := runCutoff(at, 0); got != "2026-10-21" {
			t.Errorf("%s: cutoff for 0 days = %s", zone, got)
		}
		if got := runCutoff(at.In(time.Local), 30); got != "2026-09-21" {
			t.Errorf("%s: cutoff for 30 days = %s", zone, got)
		}
	}
}
