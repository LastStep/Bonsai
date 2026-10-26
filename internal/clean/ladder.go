package clean

// The ladder kind (contract §11): <task id>.json in .bonsai/local/ladder/ (a task's result; ci.json, a --ci run's,
// names no task and is left alone), and the copies of a task's results, <task id>/<commit>.json, one per commit climbed
// (BUG-126); each aged by the result's finished, and kept while its task is not done or cut.
//
// The task's status is read from the main checkout's task files (bonsai.yaml's documents.task, through
// workspace.ReadTaskFiles, the one reader of the task folder), once a run. A task is done only when a task file with
// its id reads, and every task file with that id reads done or cut: a task not found, a task file that does not read,
// or a task folder that cannot be read counts as not done. A result names its task twice, by its file's name and by
// its task field; both must be done.

import (
	"errors"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/reader"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// ladderName is a task's result's name: the task id pattern (workspace.TaskIDPattern, the task schema's) and .json.
var ladderName = regexp.MustCompile(strings.TrimSuffix(workspace.TaskIDPattern, "$") + `\.json$`)

var ladderKind = fileKind{
	name:    Ladder,
	folder:  LadderFolder,
	match:   ladderName.MatchString,
	age:     ladderAge,
	protect: ladderProtected,
}

// ladderCopyName is a copy's name in its task's folder (BUG-126): the commit's full id (SHA-1's 40 hex digits, or
// SHA-256's 64) and .json.
var ladderCopyName = regexp.MustCompile(`^(?:[0-9a-f]{40}|[0-9a-f]{64})\.json$`)

// ladderTaskFolder is a task's folder of copies: the task id pattern alone.
var ladderTaskFolder = regexp.MustCompile(workspace.TaskIDPattern)

// ladderCopyKind is the copies of a task's results, one per commit climbed, at <task>/<sha>.json (BUG-126): the same
// rule (generated.ladder), the same clock and the same protection as the results, the task read from the folder's
// name and the copy's task field. keep_newest ranks the copies among themselves, apart from the results.
var ladderCopyKind = fileKind{
	name:    Ladder,
	folder:  LadderFolder,
	match:   ladderCopyName.MatchString,
	sub:     ladderTaskFolder.MatchString,
	age:     ladderAge,
	protect: ladderProtected,
}

// ladderResult is what the cleaner reads of a result: its task field.
type ladderResult struct {
	task *string
}

// ladderAge is a result's finished, its modification time when the result does not read.
func ladderAge(_ *run, f *file) time.Time {
	raw, err := os.ReadFile(f.path)
	if err != nil {
		return f.mtime
	}
	l, err := format.ReadLadder(raw)
	if err != nil {
		return f.mtime
	}
	f.ladder = &ladderResult{task: l.Task}
	at, err := time.Parse(time.RFC3339Nano, l.Finished)
	if err != nil {
		return f.mtime
	}
	return at
}

// ladderProtected keeps a result whose task is not done or cut.
func ladderProtected(r *run, f *file) (bool, error) {
	ts := r.taskStates()
	id := strings.TrimSuffix(f.name, ".json")
	if folder, _, ok := strings.Cut(f.name, "/"); ok { // a copy: <task>/<sha>.json
		id = folder
	}
	if !ts.done(id) {
		return true, nil
	}
	if f.ladder != nil && f.ladder.task != nil && !ts.done(*f.ladder.task) {
		return true, nil
	}
	return false, nil
}

// taskStates are the main checkout's task files' statuses by id; err when the folder could not be read (no task
// counts as done).
type taskStates struct {
	status map[string][]string // "" for a task file that does not read
	err    error
}

// done reports a task that is done or cut: found, and every file with its id reading done or cut.
func (ts *taskStates) done(id string) bool {
	if ts.err != nil {
		return false
	}
	list := ts.status[id]
	if len(list) == 0 {
		return false
	}
	for _, s := range list {
		if s != "done" && s != "cut" {
			return false
		}
	}
	return true
}

// taskStates reads the main checkout's task files once a run.
func (r *run) taskStates() *taskStates {
	if r.tasks != nil {
		return r.tasks
	}
	r.tasks = &taskStates{status: map[string][]string{}}
	dir, ok := taskFolder(r.cfg.Doc)
	if !ok {
		r.tasks.err = errors.New("bonsai.yaml's documents.task is not a folder")
		return r.tasks
	}
	entries, err := workspace.ReadTaskFiles(r.main, dir)
	if err != nil {
		r.tasks.err = err
		return r.tasks
	}
	for _, e := range entries {
		s := ""
		if !e.Broken && e.Task != nil {
			s = e.Task.Status
		}
		r.tasks.status[e.ID] = append(r.tasks.status[e.ID], s)
	}
	return r.tasks
}

// taskFolder is bonsai.yaml's documents.task, as read.
func taskFolder(doc *reader.Map) (string, bool) {
	d, ok := doc.Get("documents")
	if !ok {
		return "", false
	}
	dm, isMap := d.(*reader.Map)
	if !isMap {
		return "", false
	}
	t, _ := dm.Get("task")
	s, isText := t.(string)
	return s, isText && s != ""
}
