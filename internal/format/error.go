package format

// The error object (bonsai.error, spec §3 and §16 row 29; format review 4.5): what a command's --json carries when
// the command refuses or fails. It has no format line and no major: it is part of status, check and changes, and
// changes with them, only by additions.

import "github.com/LastStep/Bonsai/internal/schema"

// ErrorObject is the error object.
type ErrorObject struct {
	Code    string        `json:"code"`    // a fixed word: an open list whose known words are ErrorWords
	Message string        `json:"message"` // what went wrong, one sentence
	Next    Next          `json:"next"`
	Extra   schema.Object `json:"-"`
}

// Next is a next step: what to do, and whose step it is. The error object, a finding, a warning and the plugin step
// all take this shape (formats/README.md).
type Next struct {
	Do    string        `json:"do"`
	Who   string        `json:"who"` // the closed list: agent, person
	Extra schema.Object `json:"-"`
}

// ErrorWords are the known words of the error object's code: an open list (contract §2.2) whose one home is this
// table (formats/README.md); bonsai check --schema bonsai.error prints it, and step 5.1.10's reference page lists
// it. Step 5.1.4b fills it as it gives every command's refusals their words; a word is added, never renamed.
var ErrorWords = []Word{}

// ReadErrorObject reads an error object on its own (its JSON).
func ReadErrorObject(raw []byte) (*ErrorObject, error) {
	e := &ErrorObject{}
	if err := MustLookup("error").readJSONInto(raw, e); err != nil {
		return nil, err
	}
	return e, nil
}

// Encode writes the error object on its own, held to bonsai.error.
func (e *ErrorObject) Encode() ([]byte, error) { return MustLookup("error").Encode(e) }
