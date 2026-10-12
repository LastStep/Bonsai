// Package rung0 is rung 0 of the ladder (design/plan-5.md, 5.4.3; spec §9; contract §13), the guard rung on every
// floor: what a task's branch and working tree changed, each path judged by the guard's one function against the named
// task's grants, Bonsai's own refusals, and bonsai check's findings about the commit. It runs no rung job and no
// process but git; the runner (step 5.4.2) calls Judge through internal/ladder/kinds.go and copies the outcome's
// status, reason and captures into the result's rung.
//
// What changed (note 1), every path once, sorted:
//   - on a branch other than the base, every path the branch's commits changed since the merge base (git diff
//     --no-renames, so both sides of a rename), the base branch and the merge base given by the runner (5.4.2 note 13:
//     the branch the verified main checkout holds, worked out once per run). With no base (the main checkout's HEAD is
//     detached) or no merge base, the branch's commits are not judged, and the reason says so;
//   - on any branch, what is staged, unstaged or untracked and not ignored (git status --no-renames
//     --untracked-files=all; as today's changedFiles). An untracked nested repository (git names its folder with a
//     trailing slash and does not look inside it) counts as its .git, a checkout planted below the top, which the
//     guard's floor never grants; but a worktree of the same main checkout (Claude Code places its own inside the main
//     checkout, under .claude/worktrees/), known by git's link back, is another checkout of the project and passed over.
//
// Each path (note 2) is judged by guard.ChangeJudge: the guard's one function, re-rooted at the verified main
// checkout, with the named task's grants at any status, its declared documents' tripwires given the base's bytes
// (the merge base on a branch, else HEAD) and what the checkout holds now, a deleted path judged as a delete, and a
// change made in a worktree meeting command mode's task-outside-main. The named task is the active-task function's
// answer when it was named (--task or BONSAI_TASK); a task found running is not named, so it grants nothing here
// (formats/active-task's rung0 section). None named (--ci with no task): nothing is granted, so any protected change
// is red (contract §13: "None fails closed everywhere").
//
// Bonsai's own refusals (note 3):
//   - the generated tables, .bonsai/tasks.md and .bonsai/sessions.md: the guard's floor never grants them (they are
//     Bonsai's files), and so a branch other than the base that changes one is red, as the guard refuses it. On the
//     base itself they are not judged: the main checkout is where bonsai check --write rebuilds them and where they
//     are committed (spec §6: "committed, and never on a branch"), so a change to them there is not a task's. Not on
//     the base means any checkout whose branch is not the base branch, a detached HEAD and a run with no base among
//     them (fail closed);
//   - a file under .bonsai/local/ that git tracks or has staged, anywhere: bonsai check's own local finding (below)
//     counts it, once; such paths are not judged again by the guard's function, so a branch that takes one out of git
//     is not refused for the deletion.
//
// bonsai check's findings (note 4), run in the checkout (engine.CheckWith, with git history): every finding counts but
// those whose word is marked Machine in format.CheckWords (about this machine, not the commit: a proof is of a
// commit), which the reason lists as not judged here. Warnings never count. hooks-off counts from the checkout's
// settings files and from the verified main checkout's .claude/settings.local.json (where a worktree's subagent
// builders take their hooks from): bonsai check reads the main checkout git names; rung 0 reads the verified one too
// when the two differ. A checkout where bonsai check cannot run (no bonsai.yaml of its own: a branch from before the
// link, or one that deletes it) is red: the commit cannot be checked.
//
// The outcome (note 5): green, with the count of changed paths in captures (checked); red, with each refused path and
// its rule, at most MaxListed of them, bonsai check's findings by word and file, and the next step, in the reason; error
// when git cannot say what changed, or when the judgment runs past its limit (Input.Timeout: the rung's timeout_s, else
// DefaultTimeout), its git calls then ended (bonsai check's own few reads are Bonsai's and not cut). The reason also says what was not judged (the branch's commits with no base,
// machine findings), on a green rung too; and since a result keeps a rung's reason only when it is not green
// (bonsai.ladder/1), the same is a text capture, not_judged (each thing not judged, joined by "; "; null when
// everything was judged), so a green result keeps it for the verifiers. For the runner's next step (a red climb names
// each path's fix, 5.4.2 note 12), the outcome also gives the refused paths (Refused, each with what clears it), how a
// revert is made (Revert) and bonsai check's step (Check). The result never holds a whole diff.
//
// Labels (note 6): bonsai.allows is read by the guard's own reader (allowsOf), which reads contract §5.6's meaning
// with no definition in force, as the stop gate reads bonsai.ladder; rung 0 adds no reader of its own.
package rung0

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/LastStep/Bonsai/internal/guard"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// MaxListed is the most refused paths a reason lists (note 5).
const MaxListed = 50

// DefaultTimeout is rung 0's limit when Input gives none (a rung 0 with no timeout_s): 120 seconds, as git integrity's
// (vergit.DefaultTimeout). Rung 0 runs git reads of the branch's changes and bonsai check, the same order of work, so
// the limit only ends a git that does not answer; the runner counts it in its lock's budget.
const DefaultTimeout = 120 * time.Second

// maxRead is the largest file whose bytes rung 0 reads as what a declared document holds now: a bigger file is no
// document, and if it were, what it holds counts as unknown (its status line then refused in command mode).
const maxRead = 1 << 20

// Input is what rung 0 judges.
type Input struct {
	Root      string               // the checkout the ladder runs in: its top folder (a folder inside it is read as its top)
	Main      string               // the verified main checkout's top folder (Root itself in the main checkout)
	Home      string               // this machine's Bonsai home, "" for none
	Active    workspace.ActiveTask // the active-task function's answer for the run, with --task and BONSAI_TASK; its task grants only when How is named
	Base      string               // the base branch: the branch the verified main checkout holds; "" when it holds none
	MergeBase string               // the merge base of HEAD and the base branch, a commit; "" when there is none
	Timeout   time.Duration        // the judgment's limit: a git call still running then is ended; 0 means DefaultTimeout
}

// Outcome is rung 0's answer: the rung's status, reason and captures, which the runner copies into the result's rung.
type Outcome struct {
	Status   string        // a word of format.RungStatuses: green, red or error
	Reason   string        // why it is not green, and what was not judged; "" when green and everything was judged
	Captures schema.Object // checked: the count of changed paths; not_judged: what was not judged, or null; nil on error
	Refused  []Refusal     // the refused paths, at most MaxListed, in path order: what the reason lists
	More     int           // refused paths past MaxListed
	Revert   string        // how a refused path goes back to the base's copy, a command with <path> for the path; "" with no commit
	Check    string        // bonsai check's part of the next step when its findings make the rung red; "" when they do not
}

// Refusal is one changed path the guard's function refuses.
type Refusal struct {
	Path  string // relative to the checkout, forward slashes
	Rule  string // the guard's rule (guard.Rules)
	Why   string // the guard's reason, from its first words after "refused: "
	Needs string // what clears it: a grant, or a revert to the base's copy
}

// Judge runs rung 0 (the package comment).
func Judge(in Input) Outcome {
	if in.Root == "" || in.Main == "" {
		return errorOutcome("rung 0 was given no checkout or no main checkout")
	}
	timeout := in.Timeout
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	o := judge(ctx, in)
	if ctx.Err() != nil {
		return errorOutcome(fmt.Sprintf("rung 0 did not finish within its limit of %s (bonsai.yaml's timeout_s, else %s): "+
			"git did not answer in time, so the changes are not judged", timeout, DefaultTimeout))
	}
	return o
}

// judge is Judge within its deadline: a git call still running when ctx ends is ended, and Judge then gives error.
func judge(ctx context.Context, in Input) Outcome {
	top, err := gitIn(in.Root).within(ctx).top()
	if err != nil {
		return errorOutcome("git cannot say what this checkout changed (" + oneLine(err.Error()) + ")")
	}
	g := gitIn(top).within(ctx)
	branch := g.branch()
	head := g.hasHead()
	onBase := in.Base != "" && branch == in.Base
	var notJudged []string
	fromBranch := ""
	switch {
	case onBase:
	case in.Base == "":
		notJudged = append(notJudged, "the branch's commits (no base branch: the main checkout's HEAD is detached, so "+
			"only what the working tree changed was judged)")
	case in.MergeBase == "" || !head:
		notJudged = append(notJudged, "the branch's commits (no merge base with "+quote(in.Base)+", so only what the "+
			"working tree changed was judged)")
	default:
		fromBranch = in.MergeBase
	}
	changed, err := g.changed(fromBranch, in.Main)
	if err != nil {
		return errorOutcome("git cannot say what this checkout changed (" + oneLine(err.Error()) + ")")
	}
	from := fromBranch
	if from == "" && head {
		from = "HEAD"
	}
	named := grantsOf(in.Active)
	j := guard.NewChangeJudge(in.Home, in.Main, named, !samePlace(top, in.Main))
	var refused []Refusal
	for _, c := range changed {
		if underLocal(c.path) || (onBase && table(c.path)) {
			continue
		}
		d := j.Judge(g.change(c, from))
		if !d.Allow {
			refused = append(refused, refusal(c.path, d, named))
		}
	}
	found, machine, checkErr := checkFindings(in)

	o := Outcome{Status: "green", Captures: schema.Object{{Key: "checked", Value: json.Number(strconv.Itoa(len(changed)))}}}
	if len(refused) > MaxListed {
		o.More = len(refused) - MaxListed
		refused = refused[:MaxListed]
	}
	o.Refused = refused
	var lines []string
	if len(refused) > 0 {
		o.Status = "red"
		lines = append(lines, fmt.Sprintf("%d of %d changed paths refused, judged as the guard judges an edit by %s:",
			len(refused)+o.More, len(changed), whose(named)))
		for _, r := range refused {
			lines = append(lines, "- "+quote(r.Path)+" ("+r.Rule+"): "+r.Why)
		}
		if o.More > 0 {
			lines = append(lines, fmt.Sprintf("- and %d more", o.More))
		}
		lines = append(lines, "next: "+nextFor(refused, from))
		o.Revert = revertWith(from)
	}
	switch {
	case checkErr != "":
		o.Status = "red"
		lines = append(lines, "bonsai check cannot run in this checkout ("+checkErr+"), so the commit's findings cannot be "+
			"read; next: rebase the branch on the base branch so it holds the main checkout's bonsai.yaml, or restore it")
		o.Check = "bonsai check cannot run in this checkout (" + checkErr + "): rebase the branch on the base branch so it " +
			"holds the main checkout's bonsai.yaml, or restore it"
	case len(found) > 0:
		o.Status = "red"
		lines = append(lines, fmt.Sprintf("bonsai check finds %d problem(s) about the commit: %s; next: run bonsai check, "+
			"which names each one's next step", len(found), strings.Join(found, ", ")))
		o.Check = fmt.Sprintf("bonsai check finds %d problem(s) about the commit (%s): run: bonsai check --json, which names "+
			"each one's next step", len(found), strings.Join(found, ", "))
	}
	if len(machine) > 0 {
		notJudged = append(notJudged, "bonsai check's findings about this machine, not the commit ("+strings.Join(machine, ", ")+")")
	}
	for _, n := range notJudged {
		lines = append(lines, "not judged here: "+n)
	}
	var nj any
	if len(notJudged) > 0 {
		nj = strings.Join(notJudged, "; ")
	}
	o.Captures = append(o.Captures, schema.Member{Key: "not_judged", Value: nj})
	o.Reason = strings.Join(lines, "\n")
	return o
}

// errorOutcome is rung 0 unable to run: status error, the reason why.
func errorOutcome(why string) Outcome {
	return Outcome{Status: "error", Reason: why}
}

// grantsOf gives the task whose grants rung 0 judges with: the function's task when it was named, at any status
// (contract §13); else none, and why.
func grantsOf(a workspace.ActiveTask) guard.NamedGrants {
	switch {
	case a.ID != "" && a.How == workspace.HowNamed && a.Task != nil:
		return guard.NamedGrants{ID: a.ID, Task: a.Task}
	case a.ID != "":
		return guard.NamedGrants{Why: "no task is named (" + a.ID + " reads running, but rung 0 judges with a named task's " +
			"grants only: --task or " + workspace.TaskEnv + ")"}
	case a.Why != "":
		return guard.NamedGrants{Why: a.Sentence()}
	}
	return guard.NamedGrants{Why: "no task is named"}
}

// whose names the grants a refusal was judged with, for the reason's first line.
func whose(n guard.NamedGrants) string {
	if n.ID != "" {
		return "the named task " + n.ID + " (its grants at any status)"
	}
	return "no named task (nothing granted)"
}

// refusal gives a refused change its reason and what clears it.
func refusal(p string, d guard.Decision, n guard.NamedGrants) Refusal {
	why := d.Why
	if _, rest, ok := strings.Cut(why, " refused: "); ok {
		why = rest
	}
	r := Refusal{Path: p, Rule: d.Rule, Why: why}
	task := "the named task's"
	if n.ID != "" {
		task = n.ID + "'s"
	}
	switch d.Rule {
	case guard.RuleProtected:
		r.Needs = "a grant in " + task + " bonsai.allows, or a revert"
		if n.ID == "" {
			r.Needs = "a named task (--task) whose bonsai.allows grants it, or a revert"
		}
	case guard.RulePersonOnly:
		r.Needs = "a person's grant in " + task + " bonsai.allows, or a revert"
	case guard.RuleNeverGranted, guard.RuleBonsaiFiles:
		r.Needs = "a revert: no grant opens it"
	case guard.RuleTaskOutsideMain, guard.RuleStatusLine, guard.RuleLaneLine, guard.RuleGrantPersonOnly,
		guard.RuleOutsideLabel, guard.RuleBadFormat:
		r.Needs = "a revert: in command mode the workspace's status command makes this change"
	default:
		r.Needs = d.Next
	}
	return r
}

// revertWith is how a path goes back to its copy at from (the merge base on a branch, else HEAD): "" when there is no
// commit to go back to.
func revertWith(from string) string {
	if from == "" {
		return ""
	}
	return "git checkout " + from + " -- <path> (a path " + from + " does not hold: delete it)"
}

// nextFor is a red reason's next step: a grant for what a grant opens, a revert for the rest.
func nextFor(refused []Refusal, from string) string {
	revert := "revert each path to the base's copy"
	if from != "" {
		revert += " (git checkout " + from + " -- <path>; a path the base did not hold: delete it)"
	}
	grantable, other := false, false
	for _, r := range refused {
		switch r.Rule {
		case guard.RuleProtected, guard.RulePersonOnly:
			grantable = true
		case guard.RuleNeverGranted, guard.RuleBonsaiFiles, guard.RuleTaskOutsideMain, guard.RuleStatusLine,
			guard.RuleLaneLine, guard.RuleGrantPersonOnly, guard.RuleOutsideLabel, guard.RuleBadFormat:
		default:
			other = true
		}
	}
	s := revert
	if grantable {
		s = "a protected path needs a grant in the named task's bonsai.allows (a person-only one, a person's grant) or a " +
			"revert, and any other a revert: " + revert
	}
	if other {
		s += "; for the others, each path's own next step: " + firstOther(refused)
	}
	return s
}

// firstOther is the next step of the first refusal that is neither grantable nor a revert (main not verified, a list
// that cannot be read): the guard's own.
func firstOther(refused []Refusal) string {
	for _, r := range refused {
		switch r.Rule {
		case guard.RuleProtected, guard.RulePersonOnly, guard.RuleNeverGranted, guard.RuleBonsaiFiles, guard.RuleTaskOutsideMain,
			guard.RuleStatusLine, guard.RuleLaneLine, guard.RuleGrantPersonOnly, guard.RuleOutsideLabel, guard.RuleBadFormat:
			continue
		}
		return r.Needs
	}
	return ""
}

// underLocal reports a path under .bonsai/local/, as git's pathspec .bonsai/local matches it (bonsai check's local
// finding counts those it tracks or has staged).
func underLocal(p string) bool {
	return p == workspace.LocalDir || strings.HasPrefix(p, workspace.LocalDir+"/")
}

// table reports one of the generated tables (spec §6), as git names it.
func table(p string) bool {
	return p == workspace.TasksTableFile || p == workspace.SessionsTableFile
}

// readNow reads what the checkout holds at a path now: deleted when nothing is there; known when it is a file small
// enough to read.
func readNow(root, p string) (after []byte, known, deleted bool) {
	full := filepath.Join(root, filepath.FromSlash(p))
	fi, err := os.Lstat(full)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, false, true
	}
	if err != nil || !fi.Mode().IsRegular() || fi.Size() > maxRead {
		return nil, false, false
	}
	b, err := os.ReadFile(full)
	if err != nil {
		return nil, false, false
	}
	return b, true, false
}

// quote quotes a path or word for the reason, ASCII.
func quote(s string) string { return strconv.QuoteToASCII(s) }

// oneLine keeps a message on one ASCII line: printable ASCII as itself, a line break or tab as a space, any other
// character as a Go escape (a path or git's words may hold any), so a reason prints unbroken in PowerShell 5.1.
func oneLine(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r == '\n' || r == '\r' || r == '\t':
			b.WriteByte(' ')
		case r >= 0x20 && r < 0x7f:
			b.WriteRune(r)
		default:
			q := strconv.QuoteRuneToASCII(r)
			b.WriteString(q[1 : len(q)-1])
		}
	}
	return b.String()
}

// samePlace compares two folders as places: links resolved where they resolve, letter case ignored on Windows and
// macOS.
func samePlace(a, b string) bool {
	real := func(p string) string {
		p = filepath.Clean(p)
		if r, err := filepath.EvalSymlinks(p); err == nil {
			return r
		}
		return p
	}
	a, b = real(a), real(b)
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		return strings.EqualFold(a, b)
	}
	return a == b
}

// sortedUnique sorts paths and keeps each once.
func sortedUnique(list []string) []string {
	sort.Strings(list)
	out := list[:0]
	for i, s := range list {
		if i == 0 || s != list[i-1] {
			out = append(out, s)
		}
	}
	return out
}
