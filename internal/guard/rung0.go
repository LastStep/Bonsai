package guard

// Rung 0's call into the guard (design/plan-5.md, 5.4.3 notes 2 and 6; contract §13; spec §9): rung 0
// (internal/rung0) judges each path a checkout changed by this package's one function (may.go), so rung 0 and the
// guard never disagree. Nothing here holds a rule of its own; what rung 0 gives the function differs from what a hook
// call gives it in five ways, each set on the judge (rule.go) and unset for the guard:
//
//   - The grants (taskGrants): the named task's bonsai.allows at any status, contract §13's "rung 0 judges the diff
//     against the named task's grants at any status: it judges finished work", read by the guard's own reader of the
//     label (allowsOf, which reads contract §5.6's meaning with no definition in force, as the stop gate's rungsWanted
//     reads bonsai.ladder). No task named, nothing is granted ("None fails closed everywhere"). They count for the
//     main checkout's project alone: a path in another linked project inside it is granted by no task of rung 0's.
//   - Where a path is judged: re-rooted at the verified main checkout (the path relative to the checkout, joined to the
//     main checkout's folder), so the main checkout's bonsai.yaml, lock and the floor judge it, never a bonsai.yaml
//     found up the tree from the checkout's own copy: a worktree whose branch deleted it, or that never had one (made
//     before the link), is judged all the same (the guard's own limit, guard.go's last paragraph).
//   - What a declared document held (before): the guard reads it from the disk; rung 0 gives the base's bytes, read
//     only when command mode's tripwires need them. What it holds after is what the checkout holds now, given as a
//     Write of those bytes; a deleted path is judged as a delete (may's del), what it holds after unknown.
//   - Where the change was made (outside): a change made in a worktree meets command mode's task-outside-main for a
//     task file, as the guard's edit there would; one made in the main checkout, on any branch, does not.
//   - The clock (noClock): the fresh-stamp check (stamps.go, time-not-now) is not run. It judges a stamp against the
//     clock as it is written; rung 0 judges finished work hours later, and each kind's format holds the form.
//
// Nothing is recorded: rung 0's answer is a rung of the ladder result (step 5.4.2), not a guard record.

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/hookin"
)

// NamedGrants is the task whose grants rung 0 judges changes with (contract §13): the named task at any status, or
// none.
type NamedGrants struct {
	ID   string       // the named task's id; "" when no task is named
	Task *format.Task // the task as read from the main checkout's task folder; nil when ID is ""
	Why  string       // when ID is "": why no task's grants count, ASCII words that follow "and" in a refusal
}

// Change is one path a checkout changed, as rung 0 hands it to the one function.
type Change struct {
	Path       string                 // relative to the checkout's top folder, forward slashes, as git names it
	Deleted    bool                   // the checkout no longer holds the path
	After      []byte                 // what the checkout holds at the path now, when AfterKnown
	AfterKnown bool                   // false when what it holds cannot be told (deleted, not a file, too large to read)
	Before     func() ([]byte, error) // what it held at the base, read only when a judgment needs it; an error wrapping fs.ErrNotExist when it was not there
}

// ChangeTool is the word a decision on a change names its call by ("a change of "x" refused: ..."), as a hook's
// names the tool.
const ChangeTool = "a change"

// ChangeJudge is the one function as rung 0 calls it (this file's comment), for every change of one climb: what it
// reads (the main checkout's lists, its trust on this machine, the mode) is read once.
type ChangeJudge struct {
	j     *judge
	main  string
	named NamedGrants
}

// NewChangeJudge makes the judge for the changes of a checkout whose verified main checkout is main (its top folder),
// with this machine's Bonsai home (home, "" for none) and the named task's grants. outside says the checkout is not
// the main checkout itself (a worktree).
func NewChangeJudge(home, main string, named NamedGrants, outside bool) *ChangeJudge {
	main = filepath.Clean(main)
	if abs, err := filepath.Abs(main); err == nil {
		main = abs
	}
	j := newJudge(home)
	j.getenv = func(string) string { return "" } // the environment names nothing: the grants are rung 0's own
	j.outside = outside
	j.noClock = true // a fresh stamp is judged against the clock when it is written, not at the climb
	c := &ChangeJudge{j: j, main: main, named: named}
	j.session = j.project(main)
	j.taskGrants = c.grantsOf
	return c
}

// grantsOf is what a project's task grants for rung 0: the named task's bonsai.allows at any status, for the main
// checkout's project; nothing for any other.
func (c *ChangeJudge) grantsOf(p *linkedProject) *taskGrant {
	if !c.j.sameProject(p, c.j.session) {
		return &taskGrant{why: "no task of that project is named (rung 0 judges with the named task of the project it climbs)"}
	}
	n := c.named
	if n.ID == "" || n.Task == nil {
		why := n.Why
		if why == "" {
			why = "no task is named"
		}
		return &taskGrant{why: why}
	}
	globs, ok := allowsOf(n.Task)
	switch {
	case !ok:
		return &taskGrant{task: n.ID, why: "the named task " + n.ID + "'s bonsai.allows is not a list of text, so it grants nothing"}
	case len(globs) == 0:
		return &taskGrant{task: n.ID, why: "the named task " + n.ID + " grants no path (its bonsai.allows is empty)"}
	}
	return &taskGrant{task: n.ID, globs: globs}
}

// Judge judges one change by the one function (this file's comment). A refusal's Why names the change by
// ChangeTool, and the grant it lacks by the named task.
func (c *ChangeJudge) Judge(ch Change) Decision {
	abs := filepath.Join(c.main, filepath.FromSlash(ch.Path))
	var in *hookin.Payload
	if ch.AfterKnown && !ch.Deleted {
		in = &hookin.Payload{Tool: "Write", Path: abs, Content: string(ch.After)}
	}
	read, got, gotErr := false, []byte(nil), error(nil)
	c.j.before = func() ([]byte, error) {
		if !read {
			read = true
			if ch.Before == nil {
				gotErr = fs.ErrNotExist
			} else {
				got, gotErr = ch.Before()
			}
		}
		return got, gotErr
	}
	defer func() { c.j.before = nil }()
	d, _ := c.j.may(change{tool: ChangeTool, path: abs, dir: c.main, base: c.main, del: ch.Deleted, in: in})
	if c.named.ID != "" {
		// grant.go words a grant that does not hold for the guard's running task; rung 0's is the named one, at any status.
		d.Why = strings.Replace(d.Why, "the running task "+c.named.ID+" ", "the named task "+c.named.ID+" ", 1)
	}
	return d
}

// readBefore is what the path judged held before the change: rung 0's base bytes (before), else the file on disk.
func (j *judge) readBefore(abs string) ([]byte, error) {
	if j.before != nil {
		return j.before()
	}
	return os.ReadFile(abs)
}
