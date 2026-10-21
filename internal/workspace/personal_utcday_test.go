package workspace

import (
	"strings"
	"testing"
	"time"
	_ "time/tzdata" // tests only: Go on Windows ignores TZ
)

// The personal index's updated day is the UTC day, whatever zone the time is held in: 00:30 on the 22nd in Kolkata is
// still the 21st in UTC.
func TestPersonalIndexDayIsUTC(t *testing.T) {
	ist, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		t.Fatal(err)
	}
	at := time.Date(2026, 10, 22, 0, 30, 0, 0, ist)
	if got := string(PersonalIndexText(at)); !strings.Contains(got, "\nupdated: 2026-10-21\n") {
		t.Errorf("the index reads:\n%s", got)
	}
}
