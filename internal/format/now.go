package format

// Set 11's command output (contract §23; plan P-T-5795 step 2 and its Amendment 2): bonsai now --json (bonsai.now/1),
// the time now in the form every record takes (UTC) and, for a person, in the local zone, so an agent copies a time
// instead of guessing one. Printed, never stored. This file holds its type, reader and writer; internal/timefmt fills it
// and cmd/bonsai prints it (piece B).

import "github.com/LastStep/Bonsai/internal/schema"

// Now is the document of bonsai now --json. On a refusal every field but Error is null.
type Now struct {
	UTC    *string       `json:"utc"`    // the time now in UTC, the record form: "2026-10-21 08:19" (a run report's started)
	Day    *string       `json:"day"`    // the UTC day: "2026-10-21" (a task's dates, STATE's updated)
	ISO    *string       `json:"iso"`    // the same instant, ISO-8601 UTC with milliseconds: "2026-10-21T08:19:04.325Z"
	Local  *string       `json:"local"`  // the display form, never recorded: "2026-10-21 13:49 +05:30"
	Offset *string       `json:"offset"` // the local offset from UTC, for display: "+05:30"
	Zone   *string       `json:"zone"`   // the zone's abbreviation as Go prints it, for a person's eye; never parsed
	Error  *ErrorObject  `json:"error"`
	Extra  schema.Object `json:"-"`
}

// ReadNow reads bonsai.now/1's document.
func ReadNow(raw []byte) (*Now, error) {
	n := &Now{}
	if err := MustLookup("now").readJSONInto(raw, n); err != nil {
		return nil, err
	}
	return n, nil
}

// Encode writes the document, held to bonsai.now/1.
func (n *Now) Encode() ([]byte, error) { return MustLookup("now").Encode(n) }
