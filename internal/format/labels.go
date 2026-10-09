package format

// A pack's label definitions (bonsai.labels/1, contract §5.2) and lanes (bonsai.lanes/1, contract §6): YAML files a
// pack writes and Bonsai reads (the engine copies them into the lock's declares; check --pack holds a pack to them).

import "github.com/LastStep/Bonsai/internal/schema"

// Labels is a pack's labels.yaml: the definitions of one namespace.
type Labels struct {
	Namespace string        `json:"namespace"` // the namespace every label name starts with
	Version   int64         `json:"version"`   // the definitions' own version, from 1
	Labels    []LabelDef    `json:"labels"`    // the definitions
	Extra     schema.Object `json:"-"`
}

// LabelDef is one label's definition.
type LabelDef struct {
	Name        string        `json:"name"`        // <namespace>.<name>
	Kind        string        `json:"kind"`        // the closed list: choice, text, number, list
	Values      []string      `json:"values"`      // a choice's values; [] otherwise
	Items       *string       `json:"items"`       // a list's item kind (text, number), or null
	Pattern     *string       `json:"pattern"`     // a text's pattern, or null
	Max         *int64        `json:"max"`         // a text's or a list's longest, or null
	Kinds       []string      `json:"kinds"`       // the document kinds it may be set on
	SetBy       string        `json:"set_by"`      // the closed list: agent, outside
	Grants      bool          `json:"grants"`      // it grants protected paths (contract §5.5); a missing one reads as false
	Description string        `json:"description"` // one line agents see
	Extra       schema.Object `json:"-"`
}

// ReadLabels reads a labels.yaml's bytes.
func ReadLabels(raw []byte) (*Labels, error) {
	l := &Labels{}
	if _, err := MustLookup("labels").readYAMLInto(raw, l); err != nil {
		return nil, err
	}
	return l, nil
}

// Lanes is a pack's lanes.yaml.
type Lanes struct {
	Lanes []Lane        `json:"lanes"`
	Extra schema.Object `json:"-"`
}

// Lane is one lane and the two rules Bonsai's base understands for it.
type Lane struct {
	Name         string        `json:"name"`          // the lane's name, as a task's lane names it
	ApproveFirst bool          `json:"approve_first"` // a task in it needs a person's approval before it runs
	Close        string        `json:"close"`         // the closed list: person, agent
	Description  string        `json:"description"`   // one line
	Extra        schema.Object `json:"-"`
}

// ReadLanes reads a lanes.yaml's bytes.
func ReadLanes(raw []byte) (*Lanes, error) {
	l := &Lanes{}
	if _, err := MustLookup("lanes").readYAMLInto(raw, l); err != nil {
		return nil, err
	}
	return l, nil
}
