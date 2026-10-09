package format

// The task (bonsai.task/1, contract §4), and the two other markdown kinds of Bonsai's own that are read under format
// 0 too: the run report (bonsai.run/1, §7.1) and STATE (bonsai.state/1, §7.2). Agents and people write them; Bonsai
// reads them (checks, the active task, the tables) and writes none.

import (
	"github.com/LastStep/Bonsai/internal/reader"
	"github.com/LastStep/Bonsai/internal/schema"
)

// Task is a task file's frontmatter.
type Task struct {
	ID        string        `json:"id"`         // T- and 4 to 6 digits
	Title     string        `json:"title"`      // one line
	Status    string        `json:"status"`     // the closed list in the schema: todo, plan, ..., cut
	Lane      *string       `json:"lane"`       // a lane's name, or null
	DoneWhen  []string      `json:"done_when"`  // checkable statements
	DependsOn []string      `json:"depends_on"` // task ids
	BlockedBy *string       `json:"blocked_by"` // one sentence while blocked, else null
	Created   *string       `json:"created"`    // YYYY-MM-DD, or null
	Started   *string       `json:"started"`
	Finished  *string       `json:"finished"`
	Labels    schema.Object `json:"labels"` // label name to value (contract §5), in file order
	Extra     schema.Object `json:"-"`      // fields this Bonsai does not know, kept as read

	// Format0 is the frontmatter as read under format 0 (a file with no format: line, contract §2.3), nil for a
	// format-1 file. Such a file is read as it is: each field above holds its format-0 value when that value fits
	// the field (a status of running reads as running), and is empty otherwise; nothing is refused for a field.
	Format0 *reader.Map `json:"-"`
}

// ReadTask reads a task file's bytes, under format 1 or format 0.
func ReadTask(raw []byte) (*Task, error) {
	t := &Task{}
	m0, err := MustLookup("task").readYAMLInto(raw, t)
	if err != nil {
		return nil, err
	}
	t.Format0 = m0
	return t, nil
}

// Run is a run report's frontmatter.
type Run struct {
	ID       string        `json:"id"`       // R-<date>-<task id>, a suffix allowed
	Task     string        `json:"task"`     // the task's id
	Role     string        `json:"role"`     // the role the session ran as
	Model    string        `json:"model"`    // the model
	Started  *string       `json:"started"`  // YYYY-MM-DD HH:MM, or null
	Finished *string       `json:"finished"` // the same
	Outcome  string        `json:"outcome"`  // the closed list in the schema: running, verify, ..., abandoned
	Commits  []string      `json:"commits"`  // the commits the run made
	Labels   schema.Object `json:"labels"`
	Extra    schema.Object `json:"-"`
	Format0  *reader.Map   `json:"-"` // as Task's
}

// ReadRun reads a run report's bytes, under format 1 or format 0.
func ReadRun(raw []byte) (*Run, error) {
	r := &Run{}
	m0, err := MustLookup("run").readYAMLInto(raw, r)
	if err != nil {
		return nil, err
	}
	r.Format0 = m0
	return r, nil
}

// State is STATE's frontmatter (.bonsai/STATE.md).
type State struct {
	Updated   string        `json:"updated"`    // YYYY-MM-DD
	UpdatedBy string        `json:"updated_by"` // who rewrote it
	Labels    schema.Object `json:"labels"`
	Extra     schema.Object `json:"-"`
	Format0   *reader.Map   `json:"-"` // as Task's
}

// ReadState reads STATE's bytes, under format 1 or format 0.
func ReadState(raw []byte) (*State, error) {
	s := &State{}
	m0, err := MustLookup("state").readYAMLInto(raw, s)
	if err != nil {
		return nil, err
	}
	s.Format0 = m0
	return s, nil
}

// Memory is a memory note's frontmatter, or the memory index's (spec §10; contract §7.4).
type Memory struct {
	ID      *string       `json:"id"`      // M- and a slug; null for the index
	Title   string        `json:"title"`   // one line
	Kind    string        `json:"kind"`    // the closed list: project, feedback, reference, index
	Updated string        `json:"updated"` // YYYY-MM-DD
	Source  *string       `json:"source"`  // where the fact came from; null for the index
	Labels  schema.Object `json:"labels"`
	Extra   schema.Object `json:"-"`
}

// ReadMemory reads a memory note's or the index's bytes (format 1 only: memory is new with format 1).
func ReadMemory(raw []byte) (*Memory, error) {
	m := &Memory{}
	if _, err := MustLookup("memory").readYAMLInto(raw, m); err != nil {
		return nil, err
	}
	return m, nil
}
