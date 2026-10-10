package format

// The preview or the result of bonsai init, update and unlink with --json (bonsai.changes/1, spec §4, §6). Written
// by those commands (step 5.1.4b moves them onto Changes.Encode); read by the studio and agents.

import "github.com/LastStep/Bonsai/internal/schema"

// Changes is the changes document.
type Changes struct {
	Command    string           `json:"command"` // the closed list: init, update, unlink
	Result     string           `json:"result"`  // preview, applied, nothing, conflict, refused, failed: an open list
	Workspace  *WorkspaceRef    `json:"workspace"`
	Packs      []ChangesPack    `json:"packs"`
	Files      []ChangesFile    `json:"files"`
	Lock       *string          `json:"lock"` // written, unchanged: an open list; or null
	Settings   []ChangesSetting `json:"settings"`
	RunsCode   []ChangesCode    `json:"runs_code"`
	OwnHooks   []string         `json:"own_hooks"`
	AllowExec  bool             `json:"allow_exec"`
	LeftHooks  []string         `json:"left_hooks"`
	Unverified []string         `json:"unverified"`
	Conflicts  []string         `json:"conflicts"`
	Plugins    []ChangesPlugin  `json:"plugins"`
	Error      *ErrorObject     `json:"error"`
	Left       []string         `json:"left"` // unlink: what it leaves in place that is not among its files (set 6)
	Extra      schema.Object    `json:"-"`
}

// ChangesPack is one pack and its move.
type ChangesPack struct {
	ID      string        `json:"id"`
	Source  string        `json:"source"`
	Version string        `json:"version"`
	From    *string       `json:"from"` // the locked commit, or null for a new pack
	To      *string       `json:"to"`   // the commit taken, or null for a pack taken out
	Extra   schema.Object `json:"-"`
}

// ChangesFile is one file and what happens to it.
type ChangesFile struct {
	Path   string        `json:"path"`
	Kind   *string       `json:"kind"`
	Pack   *string       `json:"pack"`
	Result string        `json:"result"` // an open list
	Why    *string       `json:"why"`
	Saved  *string       `json:"saved"`
	Extra  schema.Object `json:"-"`
}

// ChangesSetting is one line of .claude/settings.json added, changed or removed.
type ChangesSetting struct {
	File     string        `json:"file"`
	Change   string        `json:"change"` // the closed list: add, change, remove
	Kind     string        `json:"kind"`   // an open list
	Line     string        `json:"line"`
	Was      *string       `json:"was"`
	Why      string        `json:"why"`
	RunsCode bool          `json:"runs_code"`
	Extra    schema.Object `json:"-"`
}

// ChangesCode is one item under "runs code".
type ChangesCode struct {
	Kind   string        `json:"kind"`   // hook, file, plugin: an open list
	Change string        `json:"change"` // the closed list: add, change, remove
	Pack   string        `json:"pack"`
	Item   string        `json:"item"`
	Was    *string       `json:"was"`
	Why    string        `json:"why"`
	Extra  schema.Object `json:"-"`
}

// ChangesPlugin is the plugin step's result for one pack.
type ChangesPlugin struct {
	Pack    string        `json:"pack"`
	Plugin  string        `json:"plugin"`
	Commit  string        `json:"commit"`
	Result  string        `json:"result"` // installed, waiting, failed, skipped: an open list
	Message string        `json:"message"`
	Next    *Next         `json:"next"`
	Extra   schema.Object `json:"-"`
}

// ReadChanges reads the changes output.
func ReadChanges(raw []byte) (*Changes, error) {
	c := &Changes{}
	if err := MustLookup("changes").readJSONInto(raw, c); err != nil {
		return nil, err
	}
	return c, nil
}

// Encode writes the document, held to bonsai.changes/1.
func (c *Changes) Encode() ([]byte, error) { return MustLookup("changes").Encode(c) }
