package workspace

// The active task (contract §13): one definition, one function, every reader's: status --json's
// active_task and status --active, check, the tables (step 5.1.8), and from step 5.3 on the guard, rung 0 and the stop
// gate, each doing its own thing with the answer (the guard honours grants only while the task reads running; rung 0
// judges a named task at any status; the stop gate engages only for a named task).
//
// The steps, in this order (contract §13):
//  1. Named. --task, or BONSAI_TASK when the environment counts (only for the session's own project, so a task named
//     for one project never governs another). If both name a task and they differ, there is none. The named task is
//     active if exactly one task file has its id and that file parses under its format; otherwise there is none, with
//     the reason, and a named task that is missing never falls through to step 2.
//  2. Running. Else the single task whose status is running.
//  3. None. Else none: no task running, two or more running, or a task file that does not parse (that file could be
//     the running one).
//
// Tasks are read from the main checkout's task folder (bonsai.yaml's documents.task), never a worktree's own copies.
// Format-0 and format-1 tasks count alike (format.ReadTask reads both). A task file is one at the folder's top level
// named <id>-<slug>.md or <id>.md (DocFiles); a file that parses has the id its frontmatter gives (else its name's), and
// one that does not parse has its name's id, so a named task whose file is broken is found as such. The reasons
// there is none are ActiveReasons (formats/README.md's table "Why there is none" is their one home: a test holds this
// copy to it, word for word and in order).

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/schema"
)

// TaskEnv names the variable that names the session's task (contract §13; today's TRINETRA_TASK).
const TaskEnv = "BONSAI_TASK"

// How the active task was found (status --json's active_task.how, its closed list).
const (
	HowNamed   = "named"
	HowRunning = "running"
)

// The reasons there is no active task, by code.
const (
	ReasonNamedDiffer    = "named-differ"
	ReasonNamedMissing   = "named-missing"
	ReasonNamedDuplicate = "named-duplicate"
	ReasonNamedBroken    = "named-broken"
	ReasonBrokenFile     = "broken-file"
	ReasonNoneRunning    = "none-running"
	ReasonManyRunning    = "many-running"
)

// Reason is one reason there is no active task: its code, what it means, and what a refusal says first.
type Reason struct {
	Code  string // the short code: status --json's and the fixtures' answers.json's
	Means string // what it means: formats/README.md's words
	Says  string // the words a refusal says it with, after "no active task: "
}

// ActiveReasons are the reasons there is no active task, in the order the function weighs them: a copy of
// formats/README.md's table "Why there is none", their one home, which TestActiveReasonsAreTheREADME holds it to,
// code and meaning, in order. Every refusal an absent active task causes says the reason first (ActiveTask.Sentence).
var ActiveReasons = []Reason{
	{ReasonNamedDiffer, "`--task` and `BONSAI_TASK` name different tasks.", "--task and " + TaskEnv + " name different tasks"},
	{ReasonNamedMissing, "No task file has the named id. A named task that is missing never falls through to the running task.",
		"no task file has the named id"},
	{ReasonNamedDuplicate, "Two or more task files have the named id: an id resolves to exactly one file.",
		"two or more task files have the named id"},
	{ReasonNamedBroken, "The named task's file does not parse under its format.", "the named task's file does not parse"},
	{ReasonBrokenFile, "Nothing is named, and a task file does not parse.", "a task file does not parse"},
	{ReasonNoneRunning, "Nothing is named, and no task reads `running`.", "no task reads running"},
	{ReasonManyRunning, "Nothing is named, and two or more tasks read `running`.", "two or more tasks read running"},
}

// ActiveInput is what the active-task function is given.
type ActiveInput struct {
	Main      string // the project's main checkout (system separators), where tasks are read
	TaskDir   string // its bonsai.yaml's documents.task: the task folder, project-relative
	Task      string // what --task names, "" for none (only a command that takes it: the ladder, status --active)
	Env       string // what BONSAI_TASK names in the environment, "" for none
	EnvCounts bool   // this project is the session's own: only then does Env count (contract §13)
}

// ActiveTask is the answer: the active task and how it was found, or none and why.
type ActiveTask struct {
	ID     string       // the task's id; "" for none
	How    string       // HowNamed or HowRunning; "" for none
	Why    string       // a reason code (ActiveReasons) when there is none; "" when there is one
	Detail string       // what the reason is about, for a person: "T-0901, T-0902", or the files
	File   string       // the task's file, project-relative with forward slashes; "" for none
	Task   *format.Task // the task as read; nil for none
}

// taskFile is one task file as the function read it.
type taskFile struct {
	path   string // project-relative, forward slashes
	id     string
	status string
	task   *format.Task
	broken bool
}

var (
	taskIDOnce sync.Once
	taskIDRe   *regexp.Regexp
)

func taskID() *regexp.Regexp {
	taskIDOnce.Do(func() { taskIDRe = regexp.MustCompile(TaskIDPattern) })
	return taskIDRe
}

// Active finds the active task (contract §13). Its error is a task folder or file that cannot be read at all (not
// one that does not parse, which is an answer: none): a caller fails closed on it.
func Active(in ActiveInput) (ActiveTask, error) {
	env := ""
	if in.EnvCounts {
		env = in.Env
	}
	if in.Task != "" && env != "" && in.Task != env {
		return ActiveTask{Why: ReasonNamedDiffer, Detail: "--task " + in.Task + ", " + TaskEnv + " " + env}, nil
	}
	named := in.Task
	if named == "" {
		named = env
	}
	files, err := readTasks(in.Main, in.TaskDir)
	if err != nil {
		return ActiveTask{}, err
	}
	if named != "" {
		var hits []taskFile
		for _, f := range files {
			if f.id == named {
				hits = append(hits, f)
			}
		}
		switch {
		case len(hits) == 0:
			return ActiveTask{Why: ReasonNamedMissing, Detail: named}, nil
		case len(hits) > 1:
			return ActiveTask{Why: ReasonNamedDuplicate, Detail: named + " in " + paths(hits)}, nil
		case hits[0].broken:
			return ActiveTask{Why: ReasonNamedBroken, Detail: named + " in " + hits[0].path}, nil
		}
		return ActiveTask{ID: named, How: HowNamed, File: hits[0].path, Task: hits[0].task}, nil
	}
	var broken, running []taskFile
	for _, f := range files {
		switch {
		case f.broken:
			broken = append(broken, f)
		case f.status == "running":
			running = append(running, f)
		}
	}
	switch {
	case len(broken) > 0:
		return ActiveTask{Why: ReasonBrokenFile, Detail: paths(broken)}, nil
	case len(running) == 0:
		return ActiveTask{Why: ReasonNoneRunning}, nil
	case len(running) > 1:
		var ids []string
		for _, f := range running {
			ids = append(ids, f.id)
		}
		return ActiveTask{Why: ReasonManyRunning, Detail: strings.Join(ids, ", ")}, nil
	}
	return ActiveTask{ID: running[0].id, How: HowRunning, File: running[0].path, Task: running[0].task}, nil
}

// readTasks reads every task file of the main checkout's task folder, in path order.
func readTasks(main, dir string) ([]taskFile, error) {
	names, err := DocFiles(main, DocKind{Kind: "task", Path: dir, ID: TaskIDPattern})
	if err != nil {
		return nil, &Error{File: dir, Msg: "cannot be read: " + oneLine(err), Err: err,
			Next: "check that the task folder (bonsai.yaml's documents.task) can be read"}
	}
	var out []taskFile
	for _, p := range names {
		raw, err := os.ReadFile(filepath.Join(main, filepath.FromSlash(p)))
		if err != nil {
			return nil, &Error{File: p, Msg: "cannot be read: " + oneLine(err), Err: err, Next: "check the file's permissions"}
		}
		f := taskFile{path: p, id: NameID(filepath.Base(p), taskID())}
		t, err := format.ReadTask(raw)
		if err != nil {
			f.broken = true
		} else {
			f.task, f.status = t, t.Status
			if t.ID != "" {
				f.id = t.ID
			}
		}
		out = append(out, f)
	}
	return out, nil
}

func paths(fs []taskFile) string {
	var out []string
	for _, f := range fs {
		out = append(out, f.path)
	}
	return strings.Join(out, ", ")
}

// Sentence says why there is no active task, the reason first, then what it is about; "" when there is one.
func (a ActiveTask) Sentence() string {
	if a.Why == "" {
		return ""
	}
	s := "no active task: " + a.Why
	for _, r := range ActiveReasons {
		if r.Code == a.Why {
			s = "no active task: " + r.Says
		}
	}
	if a.Detail != "" {
		s += " (" + a.Detail + ")"
	}
	return asciiOnly(s)
}

// Object is the answer as status --json's active_task holds it: id, how, and why (the sentence), each null where it
// does not apply.
func (a ActiveTask) Object() schema.Object {
	null := func(s string) any {
		if s == "" {
			return nil
		}
		return s
	}
	return schema.Object{{Key: "id", Value: null(a.ID)}, {Key: "how", Value: null(a.How)}, {Key: "why", Value: null(a.Sentence())}}
}
