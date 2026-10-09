package format

// bonsai status --json (bonsai.status/1, contract §12). Bonsai writes it (internal/status builds it); the studio's
// bridge and assistants read it. On exit 3 only format, bonsai, problems and error are filled, so every other field
// may be null.

import "github.com/LastStep/Bonsai/internal/schema"

// Status is status --json's document.
type Status struct {
	Bonsai        string           `json:"bonsai"` // this Bonsai's version
	Mode          *string          `json:"mode"`   // offline by default: typed open
	Workspace     *WorkspaceRef    `json:"workspace"`
	Home          *StatusHome      `json:"home"`
	Local         *StatusLocal     `json:"local"`
	Formats       schema.Object    `json:"formats"`   // format name to {read, write}: StatusFormats
	Documents     []StatusDocument `json:"documents"` // the declared document kinds
	Packs         []StatusPack     `json:"packs"`
	Files         *StatusFiles     `json:"files"`
	Labels        []StatusLabels   `json:"labels"`
	Lanes         []StatusLane     `json:"lanes"`
	StatusWrites  *string          `json:"status_writes"` // agents, command or null
	StatusCommand *string          `json:"status_command"`
	PersonOnly    []string         `json:"person_only"`
	ActiveTask    *ActiveTask      `json:"active_task"`
	Needs         []StatusNeed     `json:"needs"`
	Problems      []string         `json:"problems"`
	Checks        schema.Object    `json:"checks"` // with --full only: typed open
	Error         *ErrorObject     `json:"error"`
	Extra         schema.Object    `json:"-"`
}

// WorkspaceRef names a workspace: in status --json and in the changes output.
type WorkspaceRef struct {
	ID    string        `json:"id"`
	Name  string        `json:"name"`
	Root  string        `json:"root"` // the main checkout, absolute, forward slashes
	Extra schema.Object `json:"-"`
}

// StatusHome is Bonsai's home and this workspace's machine folder.
type StatusHome struct {
	Path  string        `json:"path"`
	Key   string        `json:"key"`
	Extra schema.Object `json:"-"`
}

// StatusLocal is the three folders of .bonsai/local/.
type StatusLocal struct {
	Log    string        `json:"log"`
	Asks   string        `json:"asks"`
	Ladder string        `json:"ladder"`
	Extra  schema.Object `json:"-"`
}

// StatusDocument is one declared document kind (contract §7.3).
type StatusDocument struct {
	Kind      string        `json:"kind"`
	From      string        `json:"from"`
	Path      *string       `json:"path"`
	File      *string       `json:"file"`
	ID        *string       `json:"id"`
	Format    *string       `json:"format"`
	Statuses  []string      `json:"statuses"`
	Person    [][]string    `json:"person"`
	Agent     [][]string    `json:"agent"`
	Stamp     schema.Object `json:"stamp"`
	TaskField *string       `json:"task_field"`
	Extra     schema.Object `json:"-"`
}

// StatusPack is one locked pack.
type StatusPack struct {
	ID      string        `json:"id"`
	Version string        `json:"version"`
	Commit  string        `json:"commit"`
	State   string        `json:"state"` // the closed list: ok, changed, missing
	Extra   schema.Object `json:"-"`
}

// StatusFiles counts the files that differ from the lock.
type StatusFiles struct {
	Changed        int64         `json:"changed"`
	Missing        int64         `json:"missing"`
	Format0Changed int64         `json:"format0_changed"`
	Extra          schema.Object `json:"-"`
}

// StatusLabels is one label namespace in force.
type StatusLabels struct {
	Namespace string        `json:"namespace"`
	From      string        `json:"from"`
	Version   int64         `json:"version"`
	Extra     schema.Object `json:"-"`
}

// StatusLane is one lane in force.
type StatusLane struct {
	Name         string        `json:"name"`
	ApproveFirst bool          `json:"approve_first"`
	Close        string        `json:"close"`
	From         string        `json:"from"`
	Extra        schema.Object `json:"-"`
}

// ActiveTask is the active task (contract §13).
type ActiveTask struct {
	ID    *string       `json:"id"`
	How   *string       `json:"how"` // named, running or null
	Why   *string       `json:"why"`
	Extra schema.Object `json:"-"`
}

// StatusNeed is one thing the workspace needs.
type StatusNeed struct {
	Kind    string        `json:"kind"` // pack, tool, mcp, shell, plugin: an open list
	ID      *string       `json:"id"`
	Name    *string       `json:"name"`
	Source  *string       `json:"source"`
	Version *string       `json:"version"`
	Extra   schema.Object `json:"-"`
}

// ReadStatus reads status --json's output.
func ReadStatus(raw []byte) (*Status, error) {
	s := &Status{}
	if err := MustLookup("status").readJSONInto(raw, s); err != nil {
		return nil, err
	}
	return s, nil
}

// Encode writes the document, held to bonsai.status/1.
func (s *Status) Encode() ([]byte, error) { return MustLookup("status").Encode(s) }
