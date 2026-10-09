package format

// bonsai check --json (bonsai.check/1, spec §6): what check found in a project. Printed, never stored. Written by
// bonsai check (step 5.1.4b moves it onto Check.Encode); read by the studio, agents and CI.

import "github.com/LastStep/Bonsai/internal/schema"

// Check is check --json's document.
type Check struct {
	Findings []Finding     `json:"findings"` // each one makes the exit code 1
	Warnings []Finding     `json:"warnings"` // never the exit code
	Error    *ErrorObject  `json:"error"`    // null unless check refused or failed
	Extra    schema.Object `json:"-"`
}

// Finding is one finding or warning.
type Finding struct {
	Code    string        `json:"code"`    // a fixed word: an open list whose known words are CheckWords
	File    *string       `json:"file"`    // the project-relative file it is about, or null
	Message string        `json:"message"` // one sentence
	Next    Next          `json:"next"`
	Extra   schema.Object `json:"-"`
}

// CheckWords are the known words of a finding's and a warning's code: an open list whose one home is this table
// (formats/README.md); bonsai check --schema bonsai.check prints it. Step 5.1.6 fills it with its table of findings
// and warnings, one test per word.
var CheckWords = []Word{}

// ReadCheck reads check --json's output.
func ReadCheck(raw []byte) (*Check, error) {
	c := &Check{}
	if err := MustLookup("check").readJSONInto(raw, c); err != nil {
		return nil, err
	}
	return c, nil
}

// Encode writes the document, held to bonsai.check/1.
func (c *Check) Encode() ([]byte, error) { return MustLookup("check").Encode(c) }
