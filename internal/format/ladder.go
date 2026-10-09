package format

// The ladder result (bonsai.ladder/1, contract §11): the proof that a task's checks passed, at a commit. Bonsai's
// ladder runner writes it (step 5.4, through Ladder.Encode); the stop gate and the studio read it.

import (
	"encoding/json"

	"github.com/LastStep/Bonsai/internal/schema"
)

// Ladder is one ladder result.
type Ladder struct {
	Task      *string       `json:"task"`      // the task's id, or null
	Workspace string        `json:"workspace"` // the workspace id
	Mode      string        `json:"mode"`      // the closed list: local, ci
	Started   string        `json:"started"`   // UTC, to the millisecond
	Finished  string        `json:"finished"`
	Git       LadderGit     `json:"git"`
	Requested []int64       `json:"requested"` // the rungs asked for
	Green     bool          `json:"green"`
	Rungs     []LadderRung  `json:"rungs"`
	Skipped   []LadderSkip  `json:"skipped"`
	Leftovers any           `json:"leftovers"` // an object, a text or null: typed open
	Proof     *string       `json:"proof"`
	Extra     schema.Object `json:"-"`
}

// LadderGit is the commit the result proves.
type LadderGit struct {
	SHA    string        `json:"sha"`
	Branch string        `json:"branch"`
	Dirty  bool          `json:"dirty"`
	Extra  schema.Object `json:"-"`
}

// LadderRung is one rung's outcome.
type LadderRung struct {
	Rung       int64          `json:"rung"`
	Name       string         `json:"name"`
	Kind       string         `json:"kind"` // the closed list: guard, command, ver-git
	Required   bool           `json:"required"`
	Command    *string        `json:"command"`
	Means      *string        `json:"means"`
	Status     string         `json:"status"` // green, red, pending, skipped, error: typed open
	Reason     *string        `json:"reason"`
	DurationMS *int64         `json:"duration_ms"`
	Tests      schema.Object  `json:"tests"`    // typed open, or null
	Captures   schema.Object  `json:"captures"` // name to a number, a text or null; or null
	Ratchet    *LadderRatchet `json:"ratchet"`
	Extra      schema.Object  `json:"-"`
}

// LadderRatchet is a rung's count against its floor.
type LadderRatchet struct {
	Name     string          `json:"name"`
	Was      json.Number     `json:"was"`
	Now      json.Number     `json:"now"`
	OK       bool            `json:"ok"`
	NewTests *LadderNewTests `json:"new_tests"`
	Extra    schema.Object   `json:"-"`
}

// LadderNewTests is the new tests run on the base.
type LadderNewTests struct {
	Base         string        `json:"base"`
	Count        int64         `json:"count"`
	FailedOnBase int64         `json:"failed_on_base"`
	Extra        schema.Object `json:"-"`
}

// LadderSkip is a rung skipped by a mark.
type LadderSkip struct {
	Rung     int64         `json:"rung"`
	Name     string        `json:"name"`
	Mark     *string       `json:"mark"`
	Reason   string        `json:"reason"`
	Accepted bool          `json:"accepted"`
	Extra    schema.Object `json:"-"`
}

// ReadLadder reads a ladder result's bytes.
func ReadLadder(raw []byte) (*Ladder, error) {
	l := &Ladder{}
	if err := MustLookup("ladder").readJSONInto(raw, l); err != nil {
		return nil, err
	}
	return l, nil
}

// Encode writes the result, held to bonsai.ladder/1.
func (l *Ladder) Encode() ([]byte, error) { return MustLookup("ladder").Encode(l) }
