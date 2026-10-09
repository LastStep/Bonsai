package format

// bonsai.yaml (bonsai.workspace/1, spec §6): a project's Bonsai settings, the one Bonsai file a person edits. Bonsai
// reads it in full here (the engine and check, through workspace.ReadConfigFull) and writes it (init, step 5.1.5,
// through EncodeYAML). The guard reads only the fields it judges by, through workspace.ReadConfig, never this.

import (
	"github.com/LastStep/Bonsai/internal/reader"
	"github.com/LastStep/Bonsai/internal/schema"
)

// Workspace is bonsai.yaml, every field.
type Workspace struct {
	ID            string          `json:"id"`   // ws- and 26 characters (contract §3)
	Name          string          `json:"name"` // a slug
	Packs         []WorkspacePack `json:"packs"`
	Documents     Documents       `json:"documents"`
	Protected     []string        `json:"protected"`
	PersonOnly    []string        `json:"person_only"`
	NeverEdit     []string        `json:"never_edit"`
	LadderFloor   []int64         `json:"ladder_floor"`
	Ladder        []Rung          `json:"ladder"`
	Ratchets      schema.Object   `json:"ratchets"` // a count's name to its floor (a number)
	CIMarkedTests []string        `json:"ci_marked_tests"`
	Generated     Generated       `json:"generated"`
	Extra         schema.Object   `json:"-"`
}

// WorkspacePack is one entry of packs.
type WorkspacePack struct {
	ID     string        `json:"id"`
	Source string        `json:"source"`
	Path   *string       `json:"path"` // the pack's folder in its repository, or null for the root
	Ref    string        `json:"ref"`
	Extra  schema.Object `json:"-"`
}

// Documents is where the project keeps the documents Bonsai knows (contract §7.3): Bonsai's own five, then each
// pack's document kind by its name, in Extra, in file order (the schema allows any further key there).
type Documents struct {
	Task      string        `json:"task"`
	Run       string        `json:"run"`
	Answers   string        `json:"answers"`
	Memory    string        `json:"memory"`
	Protocols string        `json:"protocols"`
	Extra     schema.Object `json:"-"`
}

// Rung is one rung of the ladder (spec §9).
type Rung struct {
	Rung      int64         `json:"rung"`
	Name      *string       `json:"name"`
	Kind      string        `json:"kind"` // the ladder result's closed list: guard, command, ver-git
	Command   *string       `json:"command"`
	Required  bool          `json:"required"`
	TimeoutS  *int64        `json:"timeout_s"`
	Ratchet   any           `json:"ratchet"` // typed open until step 5.4 (formats/README.md)
	Capture   any           `json:"capture"` // the same
	Means     *string       `json:"means"`
	Tests     *string       `json:"tests"` // TAP, go test -json or JUnit: an open string
	BaseSetup *string       `json:"base_setup"`
	Extra     schema.Object `json:"-"`
}

// Generated is how long each kind of generated file is kept (spec §6, "Generated files"), its one home.
type Generated struct {
	Log      Keep          `json:"log"`
	Asks     Keep          `json:"asks"`
	Ladder   Keep          `json:"ladder"`
	Run      Keep          `json:"run"`
	Sessions Keep          `json:"sessions"`
	Extra    schema.Object `json:"-"`
}

// Keep is one kind's cleaning rule: null keeps.
type Keep struct {
	KeepDays   *int64        `json:"keep_days"`
	KeepNewest *int64        `json:"keep_newest"`
	Extra      schema.Object `json:"-"`
}

// ReadWorkspace reads bonsai.yaml's bytes in full.
func ReadWorkspace(raw []byte) (*Workspace, error) {
	w := &Workspace{}
	if _, err := MustLookup("workspace").readYAMLInto(raw, w); err != nil {
		return nil, err
	}
	return w, nil
}

// WorkspaceFromMap reads bonsai.yaml in full from its mapping as internal/reader read it (workspace.ReadConfig
// keeps it), so the file is parsed once.
func WorkspaceFromMap(m *reader.Map) (*Workspace, error) {
	f := MustLookup("workspace")
	doc, err := f.HoldMap(m)
	if err != nil {
		return nil, err
	}
	w := &Workspace{}
	return w, f.Bind(doc, w)
}

// EncodeYAML writes bonsai.yaml as format 1's YAML, every field in the schema's order (yaml.go). comment gives a
// line's comment by its field's path ("packs[0].ref"), "" for none; nil writes none.
func (w *Workspace) EncodeYAML(comment func(path string) string) ([]byte, error) {
	return MustLookup("workspace").EncodeYAML(w, comment)
}
