package format

// Set 10's three command outputs (plan-5 5.6.0), each printed, never stored, and written by its word with --json:
//   - bonsai settings show and set (bonsai.settings/1, step 5.6.1): this machine's settings, for the workspace or, with
//     --machine, for the Bonsai home, every key of the scope with its value in force and where it came from;
//   - bonsai labels attach and detach (bonsai.attach/1, step 5.6.2): label definitions attached on this machine;
//   - bonsai status --line (bonsai.line/1, step 5.6.4): the statusline's workspace half as fields, with the line.
// The settings files themselves (the machine folder's and the home's settings.json) are Bonsai's own, read by Bonsai
// alone, and have no format: the Go file that writes them documents them. The keys, their rules and defaults are one
// Go table, their one home (step 5.6.1), not a list in the schema.

import "github.com/LastStep/Bonsai/internal/schema"

// Settings is the settings document.
type Settings struct {
	Scope    string         `json:"scope"`    // the closed list: workspace, machine
	File     *string        `json:"file"`     // the settings file's absolute place, forward slashes; null when not known
	Settings []SettingValue `json:"settings"` // every key of the scope, in the keys' table order; [] on a refusal
	Changed  []string       `json:"changed"`  // the keys settings set changed; [] for show, a set that changed nothing, a refusal
	Error    *ErrorObject   `json:"error"`
	Extra    schema.Object  `json:"-"`
}

// SettingValue is one key with its value in force and where the value came from.
type SettingValue struct {
	Key   string        `json:"key"`
	Value any           `json:"value"` // a text, a whole number, or nil for none
	From  string        `json:"from"`  // the closed list: file, default
	Extra schema.Object `json:"-"`
}

// ReadSettings reads bonsai.settings/1's document.
func ReadSettings(raw []byte) (*Settings, error) {
	s := &Settings{}
	if err := MustLookup("settings").readJSONInto(raw, s); err != nil {
		return nil, err
	}
	return s, nil
}

// Encode writes the document, held to bonsai.settings/1.
func (s *Settings) Encode() ([]byte, error) { return MustLookup("settings").Encode(s) }

// Attach is the document of bonsai labels attach and detach.
type Attach struct {
	Command   string        `json:"command"`   // the closed list: attach, detach
	Namespace *string       `json:"namespace"` // null when the command refused before it read one
	Version   *int64        `json:"version"`   // the definitions' own version, or null
	File      *string       `json:"file"`      // the machine folder's labels/<namespace>.yaml, absolute, forward slashes
	Result    *string       `json:"result"`    // attached, updated, unchanged, detached, nothing; null on a refusal
	Labels    []string      `json:"labels"`    // the names attached after the command; [] after a detach
	Error     *ErrorObject  `json:"error"`
	Extra     schema.Object `json:"-"`
}

// ReadAttach reads bonsai.attach/1's document.
func ReadAttach(raw []byte) (*Attach, error) {
	a := &Attach{}
	if err := MustLookup("attach").readJSONInto(raw, a); err != nil {
		return nil, err
	}
	return a, nil
}

// Encode writes the document, held to bonsai.attach/1.
func (a *Attach) Encode() ([]byte, error) { return MustLookup("attach").Encode(a) }

// Line is the document of bonsai status --line --json.
type Line struct {
	Workspace *LineWorkspace `json:"workspace"` // null when the folder is not linked
	Task      *LineTask      `json:"task"`      // the active task, or null
	Running   *int64         `json:"running"`   // tasks reading running when none is active; else null
	Ladder    *string        `json:"ladder"`    // green, red, old, none; null with no task
	Verify    int64          `json:"verify"`
	Asks      int64          `json:"asks"`
	Release   *string        `json:"release"` // a newer release's version, or null
	Line      string         `json:"line"`    // the text; "" when not linked
	Error     *ErrorObject   `json:"error"`
	Extra     schema.Object  `json:"-"`
}

// LineWorkspace is the workspace's id and name.
type LineWorkspace struct {
	ID    string        `json:"id"`
	Name  string        `json:"name"`
	Extra schema.Object `json:"-"`
}

// LineTask is the active task's id, title and status.
type LineTask struct {
	ID     string        `json:"id"`
	Title  string        `json:"title"`
	Status string        `json:"status"`
	Extra  schema.Object `json:"-"`
}

// ReadLine reads bonsai.line/1's document.
func ReadLine(raw []byte) (*Line, error) {
	l := &Line{}
	if err := MustLookup("line").readJSONInto(raw, l); err != nil {
		return nil, err
	}
	return l, nil
}

// Encode writes the document, held to bonsai.line/1.
func (l *Line) Encode() ([]byte, error) { return MustLookup("line").Encode(l) }
