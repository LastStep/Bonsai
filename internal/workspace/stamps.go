package workspace

// The stamp fields of each document kind, their one home (design/contract.md section 23, "Time"): the guard's
// fresh-stamp rule (internal/guard/stamps.go, time-not-now) and bonsai check's future-stamp finding (BUG-123,
// internal/engine/checkdocs.go) both read them here, with the same window for a time.
//
//   - a task: created, started, finished (days);
//   - a run report: started, finished (record times, YYYY-MM-DD HH:MM in UTC);
//   - STATE: updated (a day);
//   - a memory note: updated (a day);
//   - a pack's kind: each field named as a value of its stamp (contract section 7.3: set to the day's date when a
//     document reaches a status), a day.

import (
	"fmt"
	"time"

	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/timefmt"
)

// StampWindow is how far a time stamp may lie from Bonsai's clock: the guard's window for a fresh time, either way, and
// check's for a time ahead of the UTC now.
const StampWindow = 10 * time.Minute

// StampField is one stamp field: its name, whether it holds a time (else a day), and whether every change of it counts
// as fresh to the guard (STATE's and a memory note's updated, "the day it was rewritten").
type StampField struct {
	Name   string
	Time   bool
	Always bool
}

// BonsaiStamps gives the stamp fields of one of Bonsai's own kinds (task, run, state, memory); none for another.
func BonsaiStamps(kind string) []StampField {
	switch kind {
	case "task":
		return []StampField{{Name: "created"}, {Name: "started"}, {Name: "finished"}}
	case "run":
		return []StampField{{Name: "started", Time: true}, {Name: "finished", Time: true}}
	case "state", "memory":
		return []StampField{{Name: "updated", Always: true}}
	}
	return nil
}

// PackStamps gives the stamp fields a pack's kind names in its stamp ({status: field}), each a day, in the stamp's
// order with none twice.
func PackStamps(stamp schema.Object) []StampField {
	var out []StampField
	seen := map[string]bool{}
	for _, m := range stamp {
		if f, ok := m.Value.(string); ok && f != "" && !seen[f] {
			seen[f] = true
			out = append(out, StampField{Name: f})
		}
	}
	return out
}

// KindStamps gives a document kind's stamp fields: Bonsai's own by its name, a pack's from its stamp.
func KindStamps(k DocKind) []StampField {
	if k.From == "bonsai" {
		return BonsaiStamps(k.Kind)
	}
	return PackStamps(k.Stamp)
}

// FutureStamp says how a stamp's value lies ahead of now (taken in UTC): a time more than StampWindow after it, a day
// after the UTC today; "" when it does not, and for a value that reads as neither (its kind's format judges that).
func FutureStamp(f StampField, v string, now time.Time) string {
	now = now.UTC()
	if f.Time {
		t, err := timefmt.ParseRecord(v)
		if err != nil {
			return ""
		}
		if d := t.Sub(now); d > StampWindow {
			return fmt.Sprintf("%s is %d minutes ahead of the UTC now (%s), more than the %d a time may be", v,
				int(d/time.Minute), now.Format("2006-01-02 15:04"), int(StampWindow/time.Minute))
		}
		return ""
	}
	day, err := time.Parse("2006-01-02", v)
	if err != nil {
		return ""
	}
	if today := now.Format("2006-01-02"); day.After(time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)) {
		return fmt.Sprintf("%s is after the UTC today (%s)", v, today)
	}
	return ""
}
