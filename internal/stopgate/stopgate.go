// Package stopgate is `bonsai hook stop` (spec §7; contract §13; design/plan-5.md, 5.3.4): the Stop hook that keeps
// a session whose environment names a task from ending until a green local ladder result proves its checkout's HEAD,
// so a builder cannot stop before its proof is green.
//
// When it engages (note 1): only when BONSAI_TASK names a task and the session's own project is linked: the folder
// CLAUDE_PROJECT_DIR names (Claude Code sets it to the folder the session started in; an agent's cd does not move it),
// or the nearest folder above it holding bonsai.yaml (hookin.FindProjects, as every hook finds its project). Otherwise
// it allows and writes no record (the recorder's stop record stands). A hook takes no --task: the environment alone
// names the task, and it counts here, the session's own project (contract §13).
//
// The session's checkout, whose HEAD the result must prove and whose run report is read (contract §13: the gate's
// project is the one holding the session's working folder): the git checkout holding the payload's cwd when git's own
// link ties it to the project (cwdCheckout: the verified main checkout itself, or a worktree whose .git names the main
// checkout's .git/worktrees/<name> and whose gitdir there names it back, with or without a bonsai.yaml of its own),
// else CLAUDE_PROJECT_DIR's. Claude Code keeps CLAUDE_PROJECT_DIR at the folder the session started in while cwd
// follows the session: after its EnterWorktree tool, CLAUDE_PROJECT_DIR is the main checkout and cwd the worktree; in a
// `claude --worktree` session both are the worktree (measured on 2.1.296). A cwd in a checkout of the project that
// git's link does not tie to it (a broken link back, another repository inside the project, a .git Bonsai cannot read)
// fails closed: CLAUDE_PROJECT_DIR's checkout is judged, and the result must prove the HEAD of the checkout holding cwd
// as well. A cwd in no checkout of the project (none, a folder in no checkout, a repository outside the project) leaves
// CLAUDE_PROJECT_DIR's checkout judged alone. An agent can move cwd, as it can move HEAD; either way the result must prove the HEAD of the
// checkout the session is in when it ends.
//
// What it checks once engaged, in this order. Each failure blocks (exit 2) with its reason first and its next step,
// which Claude Code gives the agent as the reason to go on; the rule that decided is the record's (rules below):
//  1. the session's own bonsai.yaml reads (bad-config: blocked with no record, since a record needs the workspace id),
//     and the payload reads as a Stop payload (bad-input);
//  2. the main checkout is verified (workspace.VerifyMain: git's own link back for a worktree, and this machine's
//     record), since the task and the ladder result are read from it and an unverified main could be a decoy
//     (main-not-verified, recorded in the session's own checkout);
//  3. main's bonsai.yaml reads in full (bad-config): its task folder, its run folder and ladder_floor;
//  4. the named task: exactly one file in main's task folder has the id, and it parses under its format (the active-task
//     function, workspace.Active: stop-task-missing, stop-task-twice, stop-task-broken; stop-task-unreadable when the
//     folder cannot be read), and its bonsai.ladder label (format 0's ladder) reads as rung numbers (stop-bad-rungs);
//  5. a run report for the task in the session's own checkout's run folder (main's documents.run), named
//     R-<date>-<task id>[-<more>].md by the run kind's id pattern (stop-no-run-report): today's gate's first check,
//     as contract §13 has the gate read "the run report from the session's own checkout, where the builder writes it";
//  6. the ladder result at main's .bonsai/local/ladder/<task id>.json (contract §11): there (stop-no-result), reading
//     as bonsai.ladder/1 (stop-bad-result), made for the task (its task the named id and its workspace this one's:
//     stop-other-task), mode local (stop-not-local), green (stop-red), proving the commit at the
//     session's checkout's HEAD (stop-other-commit; HEAD read from the checkout's git files with no git process,
//     head.go, and a HEAD it cannot read blocks: stop-head-unreadable), made on a clean tree (stop-dirty), its
//     requested rungs holding the workspace's ladder_floor and the task's bonsai.ladder (contract §5.6:
//     stop-rung-missing), each of them green among its rungs (stop-rung-not-green).
//
// When every check holds it allows (stop-proven). Unlike today's gate, the ladder result is needed whatever the task's
// status says, as spec §7 and the plan's note 2 have it, and a git process is never started.
//
// The frame (internal/hookin's Frame, as the guard's): its own timer first, at Budget, half the line's 10 s timeout (a
// Stop hook that times out does not block, and Claude Code's onFailure setting has no effect on Stop hooks), the work in
// a goroutine, a panic as a block (internal-error, unrecorded, as the guard's), past the budget a block (over-time)
// whose record, once the workspace id is known, is awaited for hookin.RecordWait.
//
// Claude Code's own limit (note 5): Stop hooks keep a turn going eight times in a row at most (a tool call starts the
// count again); the ninth block in a row is overridden and the turn ends (CLAUDE_CODE_STOP_HOOK_BLOCK_CAP raises it;
// measured on 2.1.296: nine records of a block, the ninth overridden). The gate does not read stop_hook_active: it
// blocks each time, and each block is recorded, so the log shows every one.
//
// Its records (note 7): a guard record (bonsai.log/1, agent_event Stop) for every decision once it engages, allow or
// deny, through internal/record's one append path, in the verified main checkout's .bonsai/local/log/ (the session's own
// checkout's while main is not verified); an allow it cannot record is blocked (log-failed), as the guard's is.
//
// What it prints: nothing on an allow; on a block, "bonsai stop gate: <why>.\nnext: <next>.\n" on stderr, one ASCII
// line each, never words that say a file is missing as the operating system says it ("No such file or directory"):
// Claude Code reads a Stop hook's exit 2 with such words on stderr as a hook that failed to start, and lets the
// session end (its hooks reference, read on 2.1.296; measured in a -p session on WSL the same day, while "bonsai: not
// found" still kept the session going).
package stopgate

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/hookin"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// Budget is the stop gate's own time limit: half the 10 s timeout of its hook line (engine.StopTimeout), so it blocks
// before the timeout could let the session end.
const Budget = 5 * time.Second

// The rules a stop gate decision names: the record's rule field (contract §8.1). Their one home is the guard's rules
// table (guard.Rules, internal/guard/rules.go), which the reference page lists; this package cannot import the guard,
// whose tests run the stop gate, so it keeps these copies, and TestRuleNames holds them to that table, name and answer.
// The first eight are the guard's own words for the same case.
const (
	RuleBadInput        = "bad-input"         // deny: the payload cannot be read, or it is not a Stop payload
	RuleBadConfig       = "bad-config"        // deny: a bonsai.yaml the gate reads cannot be read
	RuleNoProject       = "no-project"        // deny (no record): CLAUDE_PROJECT_DIR is not set, or its folder cannot be read
	RuleMainNotVerified = "main-not-verified" // deny: the session's main checkout is not verified, so what it would read there is not trusted
	RuleLogFailed       = "log-failed"        // deny: the gate would allow the session's end but cannot record it
	RuleOverTime        = "over-time"         // deny: the gate ran past its own time limit
	RuleInternal        = "internal-error"    // deny: the gate failed inside (a panic)

	RuleProven         = "stop-proven"          // allow: a green local ladder result proves the checkout's HEAD
	RuleTaskMissing    = "stop-task-missing"    // deny: no task file has the named id
	RuleTaskTwice      = "stop-task-twice"      // deny: two or more task files have the named id
	RuleTaskBroken     = "stop-task-broken"     // deny: the named task's file does not parse under its format
	RuleTaskUnreadable = "stop-task-unreadable" // deny: the main checkout's task folder or a file in it cannot be read
	RuleBadRungs       = "stop-bad-rungs"       // deny: the task's bonsai.ladder, or the workspace's ladder_floor, is not a list of rung numbers
	RuleNoRunReport    = "stop-no-run-report"   // deny: the session's checkout holds no run report for the task
	RuleNoResult       = "stop-no-result"       // deny: there is no ladder result for the task in the main checkout
	RuleBadResult      = "stop-bad-result"      // deny: the ladder result cannot be read, or does not read as bonsai.ladder/1
	RuleOtherTask      = "stop-other-task"      // deny: the ladder result at the task's path was made for another task (another id, none, or another workspace)
	RuleNotLocal       = "stop-not-local"       // deny: the ladder result's mode is not local (a CI run is never a task's proof)
	RuleRed            = "stop-red"             // deny: the ladder result is not green
	RuleHeadUnreadable = "stop-head-unreadable" // deny: the commit at the checkout's HEAD cannot be read from its git files
	RuleOtherCommit    = "stop-other-commit"    // deny: the ladder result proves another commit than the checkout's HEAD
	RuleDirty          = "stop-dirty"           // deny: the ladder result was made on a tree with uncommitted changes
	RuleRungMissing    = "stop-rung-missing"    // deny: the result did not ask for a rung of the floor or of the task's bonsai.ladder
	RuleRungNotGreen   = "stop-rung-not-green"  // deny: a rung of the floor or of the task's bonsai.ladder is not green in the result

	// Allowed with no record: the gate does not engage.
	RuleNotNamed  = "stop-not-named" // allow (no record): BONSAI_TASK names no task
	RuleNotLinked = "not-linked"     // allow (no record): the session's folder is in no linked project
)

// Options are what Main needs besides the payload.
type Options struct {
	Stdin         io.Reader           // the payload
	Stderr        io.Writer           // a block's reason and next step
	StdinTerminal bool                // stdin is a terminal: the gate was run by hand with no payload piped in
	Getenv        func(string) string // os.Getenv, or a test's own
	Budget        time.Duration       // the gate's own time limit; 0 means Budget
	RecordWait    time.Duration       // how long the over-time block waits for its record; 0 means hookin.RecordWait
	Now           func() time.Time    // the clock; nil means time.Now
}

// Decision is the gate's answer to one session's end.
type Decision struct {
	Allow  bool   // true: exit 0, the session ends; false: exit 2, it goes on
	Rule   string // the rule that decided (the constants above)
	Target string // what the decision is about, relative to the checkout whose local/ holds the record: the ladder result's path, the run folder; "" for none
	Why    string // a block's reason, said first (spec §7); ASCII, no final full stop
	Next   string // a block's next step (spec §3: every refusal names one); ASCII, no final full stop
}

func allow(rule, target string) Decision { return Decision{Allow: true, Rule: rule, Target: target} }

func deny(rule, target, why, next string) Decision {
	return Decision{Rule: rule, Target: target, Why: why, Next: next}
}

// run is one gate run's state, shared by the worker and the timer under mu.
type run struct {
	o  Options
	mu sync.Mutex
	at *place // where the record goes; nil until the workspace id is known
	in *hookin.Payload
}

// place is where a record goes: the workspace id it is written under, and the checkout whose local/ holds it.
type place struct {
	id    string
	local workspace.Local
}

// Main answers one Stop call: 0 lets the session end, 2 keeps it going with the reason and the next step on stderr.
// It returns by its budget (and the over-time record's wait) whatever the work is doing; the caller then exits, which
// stops the work.
func Main(o Options) int {
	if o.Getenv == nil {
		o.Getenv = os.Getenv
	}
	if o.Budget <= 0 {
		o.Budget = Budget
	}
	if o.Now == nil {
		o.Now = time.Now
	}
	r := &run{o: o}
	f := hookin.Frame{Name: "stop gate", Budget: o.Budget, Wait: o.RecordWait, Stderr: o.Stderr}
	return f.Run(hookin.Job{
		Work:     func() hookin.Answer { return r.work().answer() },
		Panic:    func(p any) hookin.Answer { return panicked(p).answer() },
		OverTime: r.overTime,
	})
}

// answer gives a decision as the frame's answer, its words made safe to print (safe).
func (d Decision) answer() hookin.Answer {
	return hookin.Answer{Block: !d.Allow, Why: safe(oneLine(d.Why)), Next: safe(oneLine(d.Next))}
}

// panicked is the block of a gate that failed inside.
func panicked(p any) Decision {
	return deny(RuleInternal, "", fmt.Sprintf("the stop gate failed inside (%s), and a gate that fails keeps the session going",
		oneLine(fmt.Sprint(p))), "tell the person; a person runs `bonsai hook stop` by hand with this payload on stdin to see the failure")
}

// beforeWork runs at the start of every run's work: a variable for the tests, which hold it to run past the budget
// before anything is read, or panic.
var beforeWork = func() {}

// work does the gate's whole job and gives its decision. A panic inside is the frame's to answer (panicked).
func (r *run) work() Decision {
	beforeWork()
	named := r.o.Getenv(workspace.TaskEnv)
	if named == "" {
		r.drain()
		return allow(RuleNotNamed, "")
	}
	dir := r.o.Getenv(hookin.ProjectEnv)
	if dir == "" {
		r.drain()
		return deny(RuleNoProject, "", hookin.ProjectEnv+" is not set, so the stop gate cannot tell which project this session "+
			"is in, and "+workspace.TaskEnv+" names "+quote(named)+": it keeps the session going",
			"run the gate as Claude Code's hook (Claude Code sets "+hookin.ProjectEnv+"); to run it by hand, set "+hookin.ProjectEnv+
				" to the project's folder")
	}
	roots, _, err := hookin.FindProjects(dir)
	if err != nil {
		r.drain()
		return deny(RuleNoProject, "", "the project's folder "+quote(filepath.ToSlash(dir))+" cannot be read ("+osWords(err)+
			"), so the stop gate keeps the session going", "tell the person: check that the folder exists and can be read")
	}
	if len(roots) == 0 {
		r.drain()
		return allow(RuleNotLinked, "")
	}
	root := roots[0]
	own, err := workspace.LoadConfig(root)
	if err != nil {
		r.drain()
		return deny(RuleBadConfig, "", "bonsai.yaml cannot be read ("+oneLine(err.Error())+"), so the stop gate cannot tell where "+
			"the task "+quote(named)+" and its proof are, and it keeps the session going", nextOf(err,
			"a person fixes bonsai.yaml (bonsai check names the line)"))
	}
	h := home()
	layout, trust := verify(h, root, own.ID)
	at := &place{id: own.ID, local: workspace.Local{Root: layout.Root, Main: layout.Root, Branch: layout.Branch()}}
	if trust.Verified {
		at.local.Main = layout.Main
	}
	r.mu.Lock()
	r.at = at
	r.mu.Unlock()

	in, err := r.readInput()
	if err != nil {
		return r.recorded(deny(RuleBadInput, "", "the hook input "+err.Error()+", and an input the stop gate cannot read keeps the "+
			"session going", "tell the person: a person checks the payload Claude Code sends (bonsai hook stop, by hand, with it on stdin)"))
	}
	r.mu.Lock()
	r.in = in
	r.mu.Unlock()
	if !trust.Verified {
		return r.recorded(deny(RuleMainNotVerified, "", "the main checkout of this checkout is not verified: "+trust.Why+
			"; the task "+quote(named)+" and its ladder result are read from the main checkout, so the stop gate cannot trust them",
			trust.Next+"; then run: bonsai ladder --task "+named))
	}
	co, ok, also := cwdCheckout(in.Cwd, root, at.local.Main)
	if ok {
		layout = co
		at = &place{id: own.ID, local: workspace.Local{Root: co.Root, Main: at.local.Main, Branch: co.Branch()}}
		r.mu.Lock()
		r.at = at
		r.mu.Unlock()
	}
	return r.recorded(judgeFn(at.local, layout, also, named))
}

// verify reads the git layout of the checkout at root, whose bonsai.yaml holds the workspace id id, and whether its
// main checkout is verified (workspace.VerifyMain, given the main checkout's own id for a worktree), as the guard asks.
func verify(home, root, id string) (workspace.GitLayout, workspace.Trust) {
	layout := workspace.ReadGitLayout(root)
	mainID := id
	if layout.Worktree() {
		mainID = ""
		if c, err := workspace.LoadConfig(layout.Main); err == nil {
			mainID = c.ID
		}
	}
	return layout, workspace.VerifyMain(home, layout, id, mainID)
}

// cwdCheckout finds the checkout the session works in when it is not CLAUDE_PROJECT_DIR's (contract §13: the stop
// gate's project is the one holding the session's working folder, and the result proves that checkout's HEAD).
// Claude Code keeps CLAUDE_PROJECT_DIR at the folder the session started in, while the payload's cwd follows the
// session: after its EnterWorktree tool, CLAUDE_PROJECT_DIR is still the main checkout and cwd the worktree (measured
// on 2.1.296).
//
// The checkout holding cwd is the git checkout: the nearest folder at or above it holding a .git (checkoutTops), its
// layout read by workspace.ReadGitLayout. It is the session's (ok) when git's own link ties it to main, the session's
// verified main checkout: it is main itself, or a worktree whose .git names main's .git/worktrees/<name> and whose
// gitdir there names it back (GitLayout.LinksBack), whatever its own bonsai.yaml says or whether it has one (a branch
// made before the project was linked, a bonsai.yaml never committed): only its HEAD and its run report are read, and
// the task and the result still come from main. CLAUDE_PROJECT_DIR's checkout (root) stays the one judged otherwise.
// When cwd lies in a checkout git's link does not tie to the project (a worktree whose link back does not hold, another
// repository, a .git Bonsai cannot read) while a checkout of the project holds it further up (root, main, or a folder
// whose .git names main), it fails closed: also names that checkout, whose HEAD the result must prove as well as
// root's. A cwd in no checkout, or in a repository outside every checkout of the project, leaves root's checkout
// judged alone, as before; so do no cwd and a cwd that is not an absolute folder there.
func cwdCheckout(cwd, root, main string) (layout workspace.GitLayout, ok bool, also *untied) {
	var none workspace.GitLayout
	if cwd == "" || !filepath.IsAbs(cwd) {
		return none, false, nil
	}
	if fi, err := os.Stat(cwd); err != nil || !fi.IsDir() {
		return none, false, nil
	}
	tops := checkoutTops(filepath.Clean(cwd))
	if len(tops) == 0 {
		return none, false, nil
	}
	g := workspace.ReadGitLayout(tops[0])
	named := g.Worktree() && sameFolder(g.Main, main) // its .git names main's git folder
	switch {
	case sameFolder(g.Root, root):
		return none, false, nil
	case sameFolder(g.Root, main) && !g.Worktree():
		return g, true, nil
	case named && g.LinksBack():
		return g, true, nil
	}
	for i, top := range tops {
		l := g
		if i > 0 {
			l = workspace.ReadGitLayout(top)
		}
		if !sameFolder(top, root) && !sameFolder(top, main) && !(l.Worktree() && sameFolder(l.Main, main)) {
			continue
		}
		why := "another repository inside this project's checkout"
		switch {
		case g.GitDir == "":
			why = "a checkout whose .git Bonsai cannot read"
		case named:
			why = "a checkout whose .git names this project's main checkout, though git's own link back (gitdir in its git " +
				"folder) does not name it"
		}
		return none, false, &untied{layout: g, at: shown(g.Root, main, root), why: why}
	}
	return none, false, nil
}

// untied is a git checkout holding the session's working folder that git's own link does not tie to the project, inside
// a checkout of the project (cwdCheckout): the result must prove its HEAD as well as the judged checkout's.
type untied struct {
	layout workspace.GitLayout // its git layout, read with no git process
	at     string              // its folder as a reason prints it (shown)
	why    string              // what it is, in words that follow its folder
}

// checkoutTops gives the folders at or above dir that hold a .git, nearest first: the git checkouts holding dir, as git
// finds a checkout's top, each to be read by workspace.ReadGitLayout. Any .git ends a checkout here, one git itself
// would not take as well (an empty folder, a stray file), and so does a folder where .git cannot be looked for: a
// checkout Bonsai cannot read is never passed over (fail closed). It costs one Lstat per folder on the way.
func checkoutTops(dir string) []string {
	var tops []string
	for d := dir; ; {
		if _, err := os.Lstat(filepath.Join(d, ".git")); err == nil || !errors.Is(err, fs.ErrNotExist) {
			tops = append(tops, d)
		}
		parent := filepath.Dir(d)
		if parent == d {
			return tops
		}
		d = parent
	}
}

// shown gives a folder as a reason prints it: from the main checkout, or else from root, when it lies inside one of
// them; else whole. Forward slashes.
func shown(p, main, root string) string {
	for _, base := range []string{main, root} {
		rel, err := filepath.Rel(base, p)
		if err == nil && rel != "." && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) &&
			!filepath.IsAbs(rel) {
			return filepath.ToSlash(rel)
		}
	}
	return filepath.ToSlash(p)
}

// sameFolder reports whether two folders are one: the same clean path (letter case ignored on Windows), or one folder
// reached two ways (a link on the path, a short name), as the system tells it.
func sameFolder(a, b string) bool {
	a, b = filepath.Clean(a), filepath.Clean(b)
	if a == b || (runtime.GOOS == "windows" && strings.EqualFold(a, b)) {
		return true
	}
	x, err := os.Stat(a)
	if err != nil {
		return false
	}
	y, err := os.Stat(b)
	return err == nil && os.SameFile(x, y)
}

// judgeFn is judge, a variable for the tests, which hold it to run past the budget once the record's place is known.
var judgeFn = judge

// judge decides an engaged gate's answer once the main checkout is verified: the task, the run report, the ladder
// result (the package's comment, steps 3 to 6). local names the session's checkout (Root) and the verified main
// checkout (Main); layout is the session's checkout's git layout; also, when not nil, a checkout holding the session's
// working folder whose HEAD the result must prove too (cwdCheckout).
func judge(local workspace.Local, layout workspace.GitLayout, also *untied, named string) Decision {
	ladder := "run: bonsai ladder --task " + named
	cfg, err := workspace.LoadConfigFull(local.Main)
	if err != nil {
		return deny(RuleBadConfig, "", "the main checkout's bonsai.yaml cannot be read in full ("+oneLine(err.Error())+"), so the "+
			"stop gate cannot tell where the task "+quote(named)+" is or which rungs it needs", nextOf(err,
			"a person fixes bonsai.yaml (bonsai check names the line)")+"; then "+ladder)
	}
	w := cfg.Full
	taskDir := w.Documents.Task
	act, err := workspace.Active(workspace.ActiveInput{Main: local.Main, TaskDir: taskDir, Env: named, EnvCounts: true})
	if err != nil {
		return deny(RuleTaskUnreadable, "", "the main checkout's task folder "+quote(taskDir)+" cannot be read ("+
			oneLine(err.Error())+"), so the stop gate cannot read the task "+quote(named), nextOf(err,
			"tell the person: check that the task folder can be read")+"; then "+ladder)
	}
	if act.ID == "" {
		return taskRefusal(act, named, taskDir)
	}
	if !taskIDPattern().MatchString(act.ID) {
		return deny(RuleTaskMissing, "", workspace.TaskEnv+" names "+quote(named)+", which is not a task id (T- and 4 to 6 digits)",
			"check that "+workspace.TaskEnv+" names this session's task; then "+ladder)
	}
	want, err := rungsWanted(w.LadderFloor, act.Task)
	if err != nil {
		return deny(RuleBadRungs, act.File, err.Error()+", so the stop gate cannot tell which rungs the task needs",
			"fix the task's bonsai.ladder label to a list of rung numbers, e.g. [1, 2], in the main checkout's "+act.File+
				"; then "+ladder)
	}
	if d, ok := runReport(local.Root, w, named); !ok {
		return d
	}
	return result(local, layout, also, named, cfg.ID, want)
}

// taskRefusal is the block when the named task is not active: missing, named twice, or not parsing (contract §13).
func taskRefusal(act workspace.ActiveTask, named, taskDir string) Decision {
	why := workspace.TaskEnv + " names " + named + ", and " + strings.TrimPrefix(act.Sentence(), "no active task: ") +
		" in the main checkout's task folder " + taskDir
	ladder := "; then run: bonsai ladder --task " + named
	switch act.Why {
	case workspace.ReasonNamedDuplicate:
		return deny(RuleTaskTwice, taskDir, why+": an id names one task file",
			"give each of those files its own id in the main checkout (one keeps "+named+")"+ladder)
	case workspace.ReasonNamedBroken:
		return deny(RuleTaskBroken, taskDir, why+", so the stop gate cannot read the rungs it needs",
			"fix the task file so it parses (bonsai check names the line) in the main checkout"+ladder)
	}
	return deny(RuleTaskMissing, taskDir, why, "check that "+workspace.TaskEnv+" names this session's task; its file belongs in "+
		"the main checkout's "+taskDir+", where bonsai ladder --task "+named+" reads it too")
}

// rungsWanted gives the rungs the result must hold green: the workspace's ladder_floor and the task's bonsai.ladder
// label (format 0's ladder), sorted, each once (contract §5.6: the label adds to the floor and never drops a rung).
func rungsWanted(floor []int64, t *format.Task) ([]int64, error) {
	seen := map[int64]bool{}
	var out []int64
	add := func(n int64) {
		if !seen[n] {
			seen[n] = true
			out = append(out, n)
		}
	}
	for _, n := range floor {
		if n < 0 {
			return nil, fmt.Errorf("the workspace's ladder_floor holds %d, not a rung number", n)
		}
		add(n)
	}
	var v any
	label, has := "bonsai.ladder", false
	if t != nil {
		v, has = t.Labels.Get(label)
		if !has && t.Format0 != nil {
			label = "ladder"
			v, has = t.Format0.Get(label)
		}
	}
	if has && v != nil {
		list, isList := v.([]any)
		if !isList {
			return nil, fmt.Errorf("the task's %s is not a list of rung numbers", label)
		}
		for _, e := range list {
			n, ok := rungNumber(e)
			if !ok {
				return nil, fmt.Errorf("the task's %s holds %s, not a rung number", label, quote(fmt.Sprint(e)))
			}
			add(n)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out, nil
}

// rungNumber reads one rung number as a task's label holds it: a whole number from 0 (json.Number from format 1, a
// float64 from format 0).
func rungNumber(v any) (int64, bool) {
	switch x := v.(type) {
	case int64:
		return x, x >= 0
	case int:
		return int64(x), x >= 0
	case float64:
		if x >= 0 && x == float64(int64(x)) && x < 1<<53 {
			return int64(x), true
		}
	case interface{ String() string }:
		n, err := strconv.ParseInt(x.String(), 10, 64)
		return n, err == nil && n >= 0
	}
	return 0, false
}

// runReport finds the task's run report in the session's own checkout (root): a file of the run kind (the run folder,
// documents.run, and the run id pattern R-<date>-<task id>, workspace.DocKinds' one home) whose id names the task.
// ok is false, with the block, when there is none.
func runReport(root string, w *format.Workspace, named string) (Decision, bool) {
	kinds, _ := workspace.DocKinds(w, nil)
	var kind workspace.DocKind
	for _, k := range kinds {
		if k.Kind == "run" {
			kind = k
		}
	}
	folder := kind.Path
	next := "write this session's run report, " + folder + "/R-<YYYY-MM-DD>-" + named + ".md, in this checkout (what changed, " +
		"what did not work, what is left; an honest empty session says so); then run: bonsai ladder --task " + named
	if folder == "" {
		return deny(RuleNoRunReport, "", "the main checkout's bonsai.yaml names no run folder (documents.run), so this "+
			"checkout holds no run report for "+named, "a person sets documents.run in bonsai.yaml; then "+
			"write this session's run report there and run: bonsai ladder --task "+named), false
	}
	files, err := workspace.DocFiles(root, kind)
	if err != nil {
		return deny(RuleNoRunReport, folder, "the run folder "+quote(folder)+" of this checkout cannot be read ("+osWords(err)+
			"), so the stop gate cannot find the run report for "+named, "tell the person: check that the folder can be read"), false
	}
	idp := regexp.MustCompile(kind.ID)
	for _, f := range files {
		id := workspace.NameID(filepath.Base(f), idp)
		if len(id) > len("R-YYYY-MM-DD-") && id[len("R-YYYY-MM-DD-"):] == named {
			return Decision{}, true
		}
	}
	return deny(RuleNoRunReport, folder, "this checkout has no run report for "+named+" in its run folder "+folder+
		" (a file named R-<date>-"+named+".md)", next), false
}

// maxResult caps the ladder result the gate reads, in bytes: a larger one does not read.
const maxResult = 32 << 20

// result judges the task's ladder result in the main checkout against the session's checkout's HEAD (the package's
// comment, step 6), and the HEAD of also when it is not nil, and gives the decision. id is the workspace id.
func result(local workspace.Local, layout workspace.GitLayout, also *untied, named, id string, want []int64) Decision {
	rel := workspace.LocalDir + "/ladder/" + named + ".json"
	ladder := "run: bonsai ladder --task " + named
	raw, err := readCapped(filepath.Join(local.Main, filepath.FromSlash(rel)), maxResult)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return deny(RuleNoResult, rel, "no ladder result proves "+named+": the main checkout has no "+rel,
			"commit the work, then "+ladder+" in this checkout (it writes "+rel+" in the main checkout)")
	case err != nil:
		return deny(RuleBadResult, rel, "the ladder result "+rel+" cannot be read ("+osWords(err)+")", ladder+", which writes it again")
	}
	l, err := format.ReadLadder(raw)
	if err != nil {
		return deny(RuleBadResult, rel, "the ladder result "+rel+" does not read as bonsai.ladder/1 ("+oneLine(err.Error())+")",
			ladder+", which writes it again")
	}
	if d, ok := theTasks(l, rel, named, id); !ok {
		return d
	}
	if l.Mode != "local" {
		return deny(RuleNotLocal, rel, "the ladder result for "+named+" is a "+quote(l.Mode)+" result, and only a local run is a "+
			"task's proof (a CI run may skip tests this machine runs)", ladder+" on this machine")
	}
	if !l.Green {
		return deny(RuleRed, rel, "the ladder result for "+named+" is not green"+redRungs(l),
			"fix what the red rungs name, commit, then "+ladder)
	}
	head, err := Head(layout)
	if err != nil {
		return deny(RuleHeadUnreadable, rel, "this checkout's HEAD cannot be read from its git files ("+err.Error()+"), so the "+
			"stop gate cannot tell whether the ladder result proves the code here", "tell the person: a person checks this "+
			"checkout's git files (git status, git rev-parse HEAD); then "+ladder)
	}
	if !strings.EqualFold(l.Git.SHA, head) {
		return deny(RuleOtherCommit, rel, "the ladder result for "+named+" proves commit "+quote(short(l.Git.SHA))+", but this "+
			"checkout's HEAD is "+short(head)+": it does not describe the code here now", "commit the work, then "+ladder+" at this HEAD")
	}
	if also != nil {
		where := "the session's working folder is in " + also.at + ", " + also.why + ", so the ladder result for " + named +
			" must prove its HEAD as well as this project's checkout's"
		next := "cd back into this project's own checkout (the folder the session started in) and end the session there; a " +
			"person runs git worktree repair in the main checkout when that folder is meant as one of its worktrees; then " + ladder
		h, err := Head(also.layout)
		if err != nil {
			return deny(RuleHeadUnreadable, rel, where+", and that HEAD cannot be read from its git files ("+err.Error()+")", next)
		}
		if !strings.EqualFold(l.Git.SHA, h) {
			return deny(RuleOtherCommit, rel, where+": it proves commit "+quote(short(l.Git.SHA))+", and that HEAD is "+short(h), next)
		}
	}
	if l.Git.Dirty {
		return deny(RuleDirty, rel, "the ladder result for "+named+" was made on a tree with uncommitted changes (git.dirty is "+
			"true), so it does not prove the commit alone", "commit or put away the changes, then "+ladder)
	}
	requested := map[int64]bool{}
	for _, n := range l.Requested {
		requested[n] = true
	}
	var lacks []string
	for _, n := range want {
		if !requested[n] {
			lacks = append(lacks, strconv.FormatInt(n, 10))
		}
	}
	if len(lacks) > 0 {
		return deny(RuleRungMissing, rel, "the ladder result for "+named+" did not ask for rung "+strings.Join(lacks, ", ")+
			", which the workspace's ladder_floor and the task's bonsai.ladder require ("+rungList(want)+")",
			ladder+" (it asks for the floor and the task's rungs in one run, on the final code)")
	}
	for _, n := range want {
		if status, ok := rungStatus(l, n); !ok {
			return deny(RuleRungNotGreen, rel, "rung "+strconv.FormatInt(n, 10)+", which the task needs, "+status+
				" in the ladder result for "+named, "fix it, commit, then "+ladder)
		}
	}
	return allow(RuleProven, rel)
}

// theTasks checks the result at the task's path was made for the task: its task is the named id and its workspace this
// workspace's (contract §11: a result names the task it proves and the workspace it ran in). A result moved or copied
// there from another task, a run for no task (task null), or another workspace's result for a task of the same id
// proves nothing of this one. ok is false, with the block, when it was made for another.
func theTasks(l *format.Ladder, rel, named, id string) (Decision, bool) {
	var why string
	switch {
	case l.Task == nil:
		why = "its task is null (a run for no task), not " + named
	case *l.Task != named:
		why = "its task is " + quote(*l.Task) + ", not " + named
	case l.Workspace != id:
		why = "its workspace is " + quote(l.Workspace) + ", not this project's " + id
	default:
		return Decision{}, true
	}
	return deny(RuleOtherTask, rel, "the ladder result "+rel+" was made for another task: "+why,
		"run: bonsai ladder --task "+named+", which writes "+named+"'s own result there"), false
}

// rungStatus says whether rung n is green in the result: every entry for it green, and at least one. When it is not,
// the words say how ("is red", "is not among its rungs").
func rungStatus(l *format.Ladder, n int64) (string, bool) {
	found := false
	for _, r := range l.Rungs {
		if r.Rung != n {
			continue
		}
		found = true
		if r.Status != "green" {
			s := "is " + quote(r.Status)
			if r.Name != "" {
				s = "(" + oneLine(r.Name) + ") " + s
			}
			if r.Reason != nil && *r.Reason != "" {
				s += ": " + cut(oneLine(*r.Reason), 120)
			}
			return s, false
		}
	}
	if !found {
		return "is not among its rungs", false
	}
	return "", true
}

// redRungs names a red result's rungs that are not green, at most five, for a block's reason.
func redRungs(l *format.Ladder) string {
	var parts []string
	for _, r := range l.Rungs {
		if r.Status == "green" || r.Status == "skipped" {
			continue
		}
		if len(parts) == 5 {
			parts = append(parts, "and more")
			break
		}
		s := "rung " + strconv.FormatInt(r.Rung, 10)
		if r.Name != "" {
			s += " " + oneLine(r.Name)
		}
		s += " " + oneLine(r.Status)
		if r.Reason != nil && *r.Reason != "" {
			s += ": " + cut(oneLine(*r.Reason), 80)
		}
		parts = append(parts, s)
	}
	if len(parts) == 0 {
		return ""
	}
	return " (" + strings.Join(parts, "; ") + ")"
}

// rungList writes rung numbers as "0, 1, 2".
func rungList(ns []int64) string {
	var s []string
	for _, n := range ns {
		s = append(s, strconv.FormatInt(n, 10))
	}
	return strings.Join(s, ", ")
}

// readCapped reads a file of at most max bytes; a larger one is an error.
func readCapped(path string, max int64) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	raw, err := io.ReadAll(io.LimitReader(f, max+1))
	if err != nil {
		return nil, err
	}
	if int64(len(raw)) > max {
		return nil, fmt.Errorf("it is larger than %d MiB", max>>20)
	}
	return raw, nil
}

var (
	taskIDOnce sync.Once
	taskIDRe   *regexp.Regexp
)

// taskIDPattern is a task id's pattern (workspace.TaskIDPattern): the id names the result's file, so it is held to it.
func taskIDPattern() *regexp.Regexp {
	taskIDOnce.Do(func() { taskIDRe = regexp.MustCompile(workspace.TaskIDPattern) })
	return taskIDRe
}

// home is the Bonsai home the gate reads this machine's records from (workspace.Home: BONSAI_HOME as the hook sees
// it, else ~/.bonsai); "" when there is none, which verifies no main checkout.
func home() string {
	h, err := workspace.Home()
	if err != nil {
		return ""
	}
	return h
}

// recorded writes d's record and gives the decision to answer: an allow that cannot be recorded becomes a block.
func (r *run) recorded(d Decision) Decision {
	err := r.record(d)
	if err == nil {
		return d
	}
	if d.Allow {
		return deny(RuleLogFailed, d.Target, "the stop gate would let the session end but cannot write its record in "+
			workspace.LocalDir+"/log ("+osWords(err)+"), and an end it cannot record is not allowed",
			"tell the person: a person makes "+workspace.LocalDir+"/log/ writable (or moves aside what is in its way), then the session ends again")
	}
	d.Why += " (its record in " + workspace.LocalDir + "/log could not be written either: " + osWords(err) + ")"
	return d
}

// record writes d's record with what the run has read so far.
func (r *run) record(d Decision) error {
	r.mu.Lock()
	at, in := r.at, r.in
	r.mu.Unlock()
	if at == nil {
		return errors.New("no workspace id to record")
	}
	return writeRecord(at, in, d, r.o.Getenv, r.o.Now())
}

// overTime is the answer when the budget runs out first: a block, and, once bonsai.yaml has been read, the writer of
// its record, which the frame awaits for the run's RecordWait.
func (r *run) overTime() (hookin.Answer, func()) {
	d := deny(RuleOverTime, "", fmt.Sprintf("the stop gate ran past its own %s limit, so it keeps the session going rather "+
		"than let the hook's timeout end it unchecked", r.o.Budget),
		"end the session again; if it is kept going again, tell the person (a slow disk, or a stdin that never ends)")
	r.mu.Lock()
	known := r.at != nil
	r.mu.Unlock()
	if !known {
		return d.answer(), nil
	}
	return d.answer(), func() { _ = r.record(d) }
}

// readInput reads the payload through the hook adapter. Its error follows "the hook input": a payload that does not
// read, a fault in a value it holds, or an event other than Stop.
func (r *run) readInput() (*hookin.Payload, error) {
	p, err := hookin.Read(r.o.Stdin, r.o.StdinTerminal)
	if err != nil {
		return nil, err
	}
	if p.Fault != nil {
		return nil, p.Fault
	}
	if p.Event != "Stop" {
		return nil, fmt.Errorf("is a %s payload: the stop gate answers Stop only", strconv.QuoteToASCII(p.Event))
	}
	return p, nil
}

// drain reads and drops the payload of a run that does not read it (hookin.Drain).
func (r *run) drain() { hookin.Drain(r.o.Stdin, r.o.StdinTerminal) }

// nextOf gives an error's own next step when it is a *workspace.Error that names one, else def.
func nextOf(err error, def string) string {
	var we *workspace.Error
	if errors.As(err, &we) && we.Next != "" {
		return oneLine(we.Next)
	}
	return def
}

// osWords says why a file operation failed in the gate's own words for the common cases, never the operating
// system's (safe says why).
func osWords(err error) string {
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return "it is not there"
	case errors.Is(err, fs.ErrPermission):
		return "permission denied"
	}
	return oneLine(err.Error())
}

// unsafeWords are what a refusal never says, in any letter case: the words an operating system uses for a file that is
// not there. Claude Code reads a Stop hook's exit 2 with such words on stderr as a hook that failed to start, which
// does not block (its hooks reference: "its stderr says a file is missing, such as No such file or directory").
var unsafeWords = regexp.MustCompile(`(?i)no such file or directory|no such file|(the system )?cannot find the (file|path)` +
	`( specified)?|\bnot found\b|\benoent\b|\bmissing\b`)

// safe gives a refusal's words with each unsafe phrase (unsafeWords) said as "not there".
func safe(s string) string { return unsafeWords.ReplaceAllString(s, "not there") }

// short gives a commit as its first 12 characters.
func short(sha string) string {
	if len(sha) > 12 {
		return sha[:12]
	}
	return sha
}

// quote gives a value as an ASCII Go string, cut to 120 runes.
func quote(s string) string { return strconv.QuoteToASCII(cut(s, 120)) }

// cut keeps at most max runes of s, a cut text ending in "...".
func cut(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max-3]) + "..."
}

// oneLine keeps a message on one ASCII line: printable ASCII as itself, a line break or tab as a space, any other
// character as a Go escape, so it prints unbroken in PowerShell 5.1 (spec §3).
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
