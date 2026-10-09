package format

// The ask record (bonsai.ask/1, contract §9.1): an agent or the ladder asks a person something typed; a person
// answers. Bonsai writes it (bonsai ask and answer, the ladder's Bless, step 5.2.5) and reads it (bonsai asks).

import "github.com/LastStep/Bonsai/internal/schema"

// Ask is one ask record.
type Ask struct {
	ID        string        `json:"id"`        // a UUID
	At        string        `json:"at"`        // UTC, milliseconds optional
	Op        string        `json:"op"`        // the closed list: file, resolve, answer
	Key       string        `json:"key"`       // agent:... or ladder:...
	Workspace string        `json:"workspace"` // the workspace id
	Source    string        `json:"source"`    // the closed list: agent, ladder
	Session   *string       `json:"session"`
	Type      *string       `json:"type"` // Answer, Decide, Look, Play, Bless or a pack's: an open list
	Task      *string       `json:"task"`
	Doc       *string       `json:"doc"`
	Title     *string       `json:"title"`     // at most 300 characters
	Why       *string       `json:"why"`       // at most 600
	ThenText  *string       `json:"then_text"` // at most 300
	Options   []string      `json:"options"`   // at most 4, each at most 200 characters
	Verdict   *string       `json:"verdict"`   // pass, fail or null
	Data      schema.Object `json:"data"`      // a pack type's payload, or null
	Answer    *AskAnswer    `json:"answer"`    // on an answer record, else null
	Extra     schema.Object `json:"-"`
}

// AskAnswer is a person's answer.
type AskAnswer struct {
	By      string        `json:"by"`      // who answered, at most 60 characters
	Via     *string       `json:"via"`     // at most 30
	Choice  *string       `json:"choice"`  // the option chosen, or null
	Verdict *string       `json:"verdict"` // pass, fail or null
	Words   *string       `json:"words"`   // at most 2,000 characters
	Extra   schema.Object `json:"-"`
}

// ReadAsk reads one ask record (one line, its line feed allowed).
func ReadAsk(raw []byte) (*Ask, error) {
	a := &Ask{}
	if err := MustLookup("ask").readJSONInto(raw, a); err != nil {
		return nil, err
	}
	return a, nil
}

// Encode writes the record as one line, held to bonsai.ask/1 and its 8,192 bytes.
func (a *Ask) Encode() ([]byte, error) { return MustLookup("ask").Encode(a) }
