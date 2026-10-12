package format

// bonsai ladder --json (bonsai.climb/1, contract section 11, spec section 4 and 9): one climb of the ladder, the
// output of the ladder runner (step 5.4.2). Printed, never stored: the result it stores is bonsai.ladder/1. Named for
// the act, since bonsai.ladder/1 is the result's name. The records inside (the result, the Bless asks) are open
// objects here, as asks and logs hold theirs: the schema checker has no $ref, and a full copy would refuse an older
// or hand-written record.

import "github.com/LastStep/Bonsai/internal/schema"

// Climb is the climb document.
type Climb struct {
	Workspace *WorkspaceRef   `json:"workspace"` // null when the command refused before reading one
	Task      *string         `json:"task"`      // the task climbed for, or null
	Mode      *string         `json:"mode"`      // the closed list: local, ci; null when the command refused before it knew
	Result    *string         `json:"result"`    // the result's project-relative path, or null when none was written
	SHA256    *string         `json:"sha256"`    // the result file's SHA-256, 64 lower-case hex, or null
	Ladder    schema.Object   `json:"ladder"`    // the result as written (bonsai.ladder/1), or null
	Bless     []schema.Object `json:"bless"`     // the ask records this run wrote (bonsai.ask/1, a file or a resolve), [] for none
	Cleaned   []string        `json:"cleaned"`   // the project-relative paths cleaned after the run
	Next      *Next           `json:"next"`      // the step after a red climb, or null when green
	Error     *ErrorObject    `json:"error"`     // null unless the command refused or failed
	Extra     schema.Object   `json:"-"`
}

// ReadClimb reads bonsai.climb/1's document.
func ReadClimb(raw []byte) (*Climb, error) {
	c := &Climb{}
	if err := MustLookup("climb").readJSONInto(raw, c); err != nil {
		return nil, err
	}
	return c, nil
}

// Encode writes the document, held to bonsai.climb/1.
func (c *Climb) Encode() ([]byte, error) { return MustLookup("climb").Encode(c) }
