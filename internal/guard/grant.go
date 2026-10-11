package guard

// Grants (design/plan-5.md, 5.3.2 notes 3 and 4; contract §13, §5.5, §5.6): a guarded path that grants can open (on
// a protected list, a pack's declared protected paths, or the floor's person-only paths) is allowed when the active
// task of the nearest verified project holding it (rule.go's grantor) reads running and its bonsai.allows holds a
// glob matching the path, relative to that project's folder. Else it is refused, the reason naming the task or why
// there is none ("no active task: two or more tasks read running (T-0901, T-0902)").
//
// The active task is contract §13's one function (workspace.Active), given the task folder of the grantor's lists
// (its verified main checkout's bonsai.yaml, documents.task) and read in the main checkout; BONSAI_TASK counts only
// when the grantor is the session's own project (a worktree of it counts as it: one main checkout). The guard is a
// hook and takes no --task. A task that is named but missing, named twice or broken, two tasks running, a broken file
// beside one running, or none: no grant. A task file that does not read gives none (the function's answer); a task
// folder that cannot be read gives none too, with its reason.
//
// Who wrote a grant cannot be told here (note 4). In command mode a person-only path counts in bonsai.allows only as a
// person granted it: what stops a forged one is tripwire.go's refusal of an agent's edit that puts one there, and the
// Desk's check (contract §10.6). In agents mode grants count as written, so a person-only path is grantable like any
// protected one: the contract's own gap in a workspace no studio manages, which the floor's never-granted paths do
// not share.
//
// A value of the wrong kind grants nothing (contract §5.2): bonsai.allows (format 0: allows_assets) must be a list of
// text. Rung 0 (step 5.4) judges with the named task's grants at any status: it sets the judge's taskGrants.

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// taskGrant is what a project's active task grants, read once a run.
type taskGrant struct {
	task  string   // the active task's id, "" when none counts
	globs []string // its bonsai.allows while it reads running
	why   string   // why nothing is granted: ASCII words that follow "and"; "" when globs may grant
}

// grants says whether the project's task grants rel (relative to the project's folder); when not, why.
func (g *taskGrant) grants(rel string, fold bool) (bool, string) {
	if g.why != "" {
		return false, g.why
	}
	for _, glob := range g.globs {
		if checkGlob(glob) == nil && matchGlob(glob, rel, fold) {
			return true, ""
		}
	}
	return false, "the running task " + g.task + " does not grant it (its bonsai.allows: " + strings.Join(quoteAll(g.globs, 60), ", ") + ")"
}

// grantOf reads, once, what the project p's active task grants (this file's comment). p is verified (rule.go's
// grantor), so its lists are its main checkout's and its main checkout is local(p).Main.
func (j *judge) grantOf(p *linkedProject) *taskGrant {
	if p.grant != nil {
		return p.grant
	}
	g := j.readGrant(p)
	p.grant = g
	return g
}

func (j *judge) readGrant(p *linkedProject) *taskGrant {
	if j.taskGrants != nil {
		return j.taskGrants(p)
	}
	l := j.listsOf(p)
	if l.err != nil {
		return &taskGrant{why: "its bonsai.yaml or lock cannot be read"}
	}
	dir := docPlace(l.cfg, "task")
	if dir == "" {
		return &taskGrant{why: "no task can grant it: bonsai.yaml names no task folder (documents.task)"}
	}
	at, err := workspace.Active(workspace.ActiveInput{Main: j.local(p).Main, TaskDir: dir, Env: j.getenv(workspace.TaskEnv),
		EnvCounts: j.sameProject(p, j.session)})
	switch {
	case err != nil:
		return &taskGrant{why: "no task can grant it: the task folder " + quote(dir, 60) + " cannot be read (" + oneLine(err.Error()) + ")"}
	case at.ID == "":
		return &taskGrant{why: at.Sentence()}
	case at.Task.Status != "running":
		return &taskGrant{task: at.ID, why: fmt.Sprintf("the active task %s reads %s, and a task's grants count only while it reads running",
			at.ID, strconv.QuoteToASCII(at.Task.Status))}
	}
	globs, ok := allowsOf(at.Task)
	switch {
	case !ok:
		return &taskGrant{task: at.ID, why: "the running task " + at.ID + "'s bonsai.allows is not a list of text, so it grants nothing"}
	case len(globs) == 0:
		return &taskGrant{task: at.ID, why: "the running task " + at.ID + " grants no path (its bonsai.allows is empty)"}
	}
	return &taskGrant{task: at.ID, globs: globs}
}

// allowsOf gives a task's bonsai.allows (contract §5.6), or its format-0 allows_assets (contract §4.3); false when the
// value is there but is not a list of text (contract §5.2: a granting label of the wrong kind grants nothing).
func allowsOf(t *format.Task) ([]string, bool) {
	v, ok := t.Labels.Get("bonsai.allows")
	if !ok && t.Format0 != nil {
		v, ok = t.Format0.Get("allows_assets")
	}
	if !ok || v == nil {
		return nil, true
	}
	return textList(v)
}

// textList reads a list of text, false for anything else.
func textList(v any) ([]string, bool) {
	list, ok := v.([]any)
	if !ok {
		return nil, false
	}
	out := make([]string, 0, len(list))
	for _, e := range list {
		s, ok := e.(string)
		if !ok {
			return nil, false
		}
		out = append(out, s)
	}
	return out, true
}

// docPlace is the place bonsai.yaml's documents names for a kind, "" for none.
func docPlace(cfg *workspace.Config, kind string) string {
	if cfg == nil {
		return ""
	}
	for _, d := range cfg.Documents {
		if d.Kind == kind {
			return strings.TrimSuffix(d.Path, "/")
		}
	}
	return ""
}

// sameProject reports whether two projects are one: the same folder, or two checkouts of one verified main checkout
// (a worktree and its main, or two worktrees). Either may be nil (no project): never the same then.
func (j *judge) sameProject(a, b *linkedProject) bool {
	if a == nil || b == nil {
		return false
	}
	if a == b {
		return true
	}
	return samePlace(j.local(a).Main, j.local(b).Main)
}

// samePlace compares two clean absolute folders, letter case ignored on Windows.
func samePlace(a, b string) bool {
	if caseFoldSystem() {
		return strings.EqualFold(a, b)
	}
	return a == b
}

// quoteAll quotes each of a list, cut to max runes.
func quoteAll(list []string, max int) []string {
	out := make([]string, len(list))
	for i, s := range list {
		out[i] = quote(s, max)
	}
	return out
}
