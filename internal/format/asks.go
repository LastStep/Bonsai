package format

// bonsai asks --json (bonsai.asks/1, contract section 9, spec section 4 and 8): the output of bonsai ask, ask
// --resolve, ask --status, answer and asks: the asks a command read, each key with its state, and the record the
// command appended. Printed, never stored. Written by those commands (step 5.2.5); read by agents, the studio's
// bridge and CI. The records inside are bonsai.ask/1's, as stored: open objects here (schema.Object), because the
// schema checker has no $ref and a full copy would refuse an older or hand-written record.

import "github.com/LastStep/Bonsai/internal/schema"

// Asks is the asks document.
type Asks struct {
	Workspace *WorkspaceRef `json:"workspace"` // null when the command refused before reading one
	Asks      []AskEntry    `json:"asks"`
	Written   schema.Object `json:"written"` // the ask record this run appended, or null
	Error     *ErrorObject  `json:"error"`   // null unless the command refused or failed
	Extra     schema.Object `json:"-"`
}

// AskEntry is one key and where it stands.
type AskEntry struct {
	Key    string        `json:"key"`
	State  string        `json:"state"`  // the closed list: open, answered, resolved
	Filed  schema.Object `json:"filed"`  // the key's latest file record, as stored (bonsai.ask/1)
	Closed schema.Object `json:"closed"` // the answer or resolve record that closed it, or null
	Extra  schema.Object `json:"-"`
}

// ReadAsks reads bonsai.asks/1's document.
func ReadAsks(raw []byte) (*Asks, error) {
	a := &Asks{}
	if err := MustLookup("asks").readJSONInto(raw, a); err != nil {
		return nil, err
	}
	return a, nil
}

// Encode writes the document, held to bonsai.asks/1.
func (a *Asks) Encode() ([]byte, error) { return MustLookup("asks").Encode(a) }
