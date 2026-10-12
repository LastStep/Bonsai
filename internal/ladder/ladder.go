// Package ladder is Bonsai's ladder runner, bonsai ladder (spec §4, §9; contract §11, §13; design/plan-5.md, 5.4.2):
// it proves a piece of work done by climbing the project's rungs in order, cheapest first, and writes the result where
// the stop gate reads it, with its fingerprint in the log. It is run unattended by agents: every refusal names its next
// step as a command an agent runs as written, and a red climb says how to fix it.
//
//	o := ladder.Run(ladder.Options{Dir: cwd, Task: "T-0901", Out: os.Stdout})
//	exit, doc := o.Exit, o.Doc // bonsai.climb/1, always whole
//
// The order of one climb, each step's refusal before anything runs:
//  1. Where (note 1): the checkout holding --root, else the working folder (git's answer, workspace.Find); its
//     bonsai.yaml's id (a worktree with none of its own takes its main checkout's); a stamped test build refuses any
//     of those folders outside its scratch root (engine.Confine: test-build, exit 4, note 10); the Bonsai home; and
//     the main checkout verified as the stop gate verifies it (workspace.VerifyMain: git's own link back and this
//     machine's record). A local climb in a checkout whose main is not verified refuses (main-not-verified, exit 4):
//     it never falls back to the checkout's own folder, since the stop gate would not read a result there. A --ci climb
//     runs anywhere (a CI runner has no machine record): in a checkout whose main is not verified it reads and writes
//     the checkout's own .bonsai/, and says so.
//  2. The task (note 1, contract §13): read from the main checkout's task folder by the active-task function, with
//     --task and BONSAI_TASK (the environment counts only when the session's own project, CLAUDE_PROJECT_DIR's, is this
//     one). A local climb needs a named task: none named (a single running task is named in the next step, not taken),
//     or a named task the function does not find, is no-task (exit 2). --ci may name none.
//  3. The rungs (note 2): the floor plus the task's bonsai.ladder (stopgate.RungsWanted, the stop gate's own reading,
//     so the result asks for exactly what the gate requires); with --ci and no task, every rung. A requested rung that
//     bonsai.yaml does not define is rung-not-defined (exit 2). The rungs are main's bonsai.yaml's: a branch never
//     climbs by rungs of its own making.
//  4. One ladder at a time per home (note 8, lock.go): the home's locks/ladder.lock, held by a lock the system ends
//     with its holder; a second climb waits while the holder lives inside its budget, then refuses (ladder-busy,
//     exit 4). A nested climb (BONSAI_LADDER=1: a ladder run inside a rung) takes no lock: its outer climb holds it.
//  5. The base, once (note 13): the base branch is the branch the main checkout has checked out (read from its HEAD
//     file), and the merge base git's merge-base of it and HEAD; "" for both when main's HEAD is detached. Every kind
//     gets the same answer (kinds.go).
//  6. The rungs climbed in number order (rungs.go, kinds.go): a command rung through internal/rungjob (one at a time,
//     its package's known limit), its tests read through internal/testout and judged for skips and marks (note 4), its
//     captures and its ratchet's count (note 5); a guard or ver-git rung through kinds.go. A required rung that is not
//     green marks every later rung skipped. The climb is green only when every rung it took up is green.
//  7. The result (note 6, result.go): bonsai.ladder/1 through its writer, every free-text value redacted, written
//     whole (a temporary file renamed into place, Windows' busy retry) at the main checkout's
//     .bonsai/local/ladder/<task>.json (--ci: ci.json); git.dirty is git's status at the start or the end, HEAD moved
//     during the climb, or a finding of git integrity that hides a change. Then the fingerprint (note 7): one ladder
//     record through the log's one append path, naming the runner's own path and SHA-256. Then the cleaner's ladder
//     kind (note 9), never the result just written; then the Bless step (kinds.go, 5.4.4's).
//  8. The leftovers line (internal/leftovers), read-only, never under --ci nor in a nested climb.
//
// Exit codes (note 9): 0 green, 1 red, 2 bad input, 3 could not run or write, 4 wrong state. A stop (a signal) ends the
// running rung as at its timeout: the climb then writes a red result, the rest skipped, and exits 1.
//
// Files: ladder.go (this file: Options, Run, the steps before a rung), climb.go (the climb's state, which kinds.go's
// functions read), kinds.go (the one file through which the runner calls each rung kind and each step after a rung),
// rungs.go (command rungs, tests, captures, ratchets), result.go (the result, the fingerprint, the cleaner), lock.go
// and lock_*.go (the home's lock), git.go (the git reads), next.go (refusals' and a red climb's next steps), text.go
// (the human lines).
package ladder

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/LastStep/Bonsai/internal/engine"
	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/rungjob"
	"github.com/LastStep/Bonsai/internal/stopgate"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// Exit codes (spec §3).
const (
	ExitGreen   = 0
	ExitRed     = 1
	ExitInput   = engine.ExitInput
	ExitRuntime = engine.ExitRuntime
	ExitState   = engine.ExitState
)

// The modes a result is written in (contract §11).
const (
	ModeLocal = "local"
	ModeCI    = "ci"
)

// The variables the runner reads.
const (
	// SessionEnv is Claude Code's session id: the ladder record goes to that session's log file.
	SessionEnv = "CLAUDE_CODE_SESSION_ID"
	// ProjectEnv is the folder Claude Code started the session in: BONSAI_TASK counts only in that project.
	ProjectEnv = "CLAUDE_PROJECT_DIR"
)

// Options are one climb's command line and what it runs with.
type Options struct {
	Dir  string // the folder the command runs in (the working folder); the checkout holding it is climbed
	Root string // --root: the checkout to climb, "" for Dir's
	Task string // --task: the task to climb for, "" for none named on the command line
	CI   bool   // --ci: mode ci, never a task's proof

	Getenv func(string) string // reads BONSAI_TASK, CLAUDE_PROJECT_DIR, CLAUDE_CODE_SESSION_ID, BONSAI_LADDER; nil: os.Getenv
	Env    []string            // the rungs' environment; nil: this process's
	Now    func() time.Time    // the clock of the result and the records; nil: time.Now
	Stop   context.Context     // the stop signal: when done, the running rung is ended and the rest skipped; nil: none

	Out    io.Writer // the human lines as the climb goes (one per rung, then the summary); nil for none (--json)
	Notice io.Writer // the lock's waiting line; nil for none

	// LockPoll is how often a waiting climb tries the lock again; 0 means LockPoll. LockGrace is how long a held lock
	// whose holder has not named itself is waited for; 0 means LockGrace. Tests shorten both.
	LockPoll, LockGrace time.Duration
}

// Answer is what one climb gives: the climb document (bonsai.climb/1, always whole), its exit code, and the refusal
// or failure when there was one (its error object is the document's).
type Answer struct {
	Doc   *format.Climb
	Exit  int
	Error *engine.Error
}

// Run climbs once (this package's comment) and gives its answer. It never panics on a project's files: what cannot
// be read is a refusal with its word and next step.
func Run(o Options) *Answer {
	o = o.withDefaults()
	c, e := prepare(o)
	if e != nil {
		return refused(c, e)
	}
	release, e := takeLock(c)
	if e != nil {
		return refused(c, e)
	}
	defer release()
	c.climb()
	return c.finish()
}

func (o Options) withDefaults() Options {
	if o.Getenv == nil {
		o.Getenv = os.Getenv
	}
	if o.Env == nil {
		o.Env = os.Environ()
	}
	if o.Now == nil {
		o.Now = time.Now
	}
	if o.Stop == nil {
		o.Stop = context.Background()
	}
	if o.Out == nil {
		o.Out = io.Discard
	}
	if o.Notice == nil {
		o.Notice = io.Discard
	}
	if o.LockPoll <= 0 {
		o.LockPoll = LockPoll
	}
	if o.LockGrace <= 0 {
		o.LockGrace = LockGrace
	}
	return o
}

// nested reports a climb run inside another climb's rung (rungjob puts BONSAI_LADDER=1 in every rung's environment).
func (o Options) nested() bool { return o.Getenv(rungjob.LadderEnv) == "1" }

// prepare does the steps before a rung (this package's comment, steps 1 to 3 and 5): where, the task, the rungs and
// the base. It returns the climb, partly filled when it refuses.
func prepare(o Options) (*Climb, *engine.Error) {
	c := &Climb{o: o, CI: o.CI, Getenv: o.Getenv, Now: o.Now, started: o.Now()}
	dir := o.Dir
	if o.Root != "" {
		dir = o.Root
		if !filepath.IsAbs(dir) && o.Dir != "" {
			dir = filepath.Join(o.Dir, dir)
		}
	}
	if fi, err := os.Stat(dir); err != nil || !fi.IsDir() {
		return c, &engine.Error{Code: "bad-value", Exit: ExitInput, What: "--root " + shown(dir) + " is not a folder",
			Next: "name a checkout's folder: run: bonsai ladder --root <folder> " + c.taskFlag() + " --json"}
	}
	co, err := workspace.Find(dir)
	if err != nil {
		return c, engine.FindError(err)
	}
	c.Root = co.Root
	layout := workspace.ReadGitLayout(co.Root)
	id, mainID := "", ""
	own, ownErr := workspace.LoadConfig(co.Root)
	if layout.Worktree() {
		if m, err := workspace.LoadConfig(layout.Main); err == nil {
			mainID = m.ID
		}
	}
	switch {
	case ownErr == nil:
		id = own.ID
	case errors.Is(ownErr, os.ErrNotExist) && layout.Worktree() && mainID != "":
		id = mainID // a worktree of a branch made before the link: its main checkout's id, git's link proving the tie
	default:
		return c, engine.ConfigError(ownErr)
	}
	if !layout.Worktree() {
		mainID = id
	}
	if err := engine.Confine(dir, co.Root, layout.Main); err != nil {
		e := err.(*engine.Error)
		e.Next = "run: " + installedBonsai() + " ladder " + c.taskFlag() + " --json; that is the installed bonsai, since a test build " +
			"climbs only inside its scratch folder"
		return c, e
	}
	home, err := workspace.Home()
	if err != nil {
		return c, engine.HomeError(err)
	}
	c.Home = home
	trust := workspace.VerifyMain(home, layout, id, mainID)
	c.Verified = trust.Verified
	c.Main = co.Root
	switch {
	case trust.Verified:
		c.Main = layout.Main
	case !o.CI:
		return c, &engine.Error{Code: "main-not-verified", Exit: ExitState, Who: "person",
			What: "this checkout's main checkout is not verified on this machine: " + trust.Why + "; the result goes where the " +
				"stop gate reads it, in the verified main checkout's .bonsai/local/ladder/, so nothing was run",
			Next: trust.Next + "; then run: bonsai ladder " + c.taskFlag() + " --json"}
	default:
		c.unverified = trust.Why
	}
	cfg, err := workspace.LoadConfigFull(c.Main)
	if err != nil {
		return c, engine.ConfigError(err)
	}
	c.Config = cfg
	c.workspace = &format.WorkspaceRef{ID: cfg.ID, Name: cfg.Name, Root: filepath.ToSlash(c.Main)}
	if e := c.findTask(); e != nil {
		return c, e
	}
	if e := c.findRungs(); e != nil {
		return c, e
	}
	c.Base = workspace.ReadGitLayout(c.Main).Branch()
	if c.Base != "" {
		c.MergeBase = mergeBase(c.Root, c.Base)
	}
	c.Prior = c.prior()
	return c, nil
}

// findTask reads the task to climb for (note 1).
func (c *Climb) findTask() *engine.Error {
	counts := true
	if p := c.Getenv(ProjectEnv); p != "" {
		s, err := workspace.Find(p)
		counts = err == nil && samePlace(s.Main, c.Main)
	}
	named := c.o.Task
	env := c.Getenv(workspace.TaskEnv)
	a, err := workspace.Active(workspace.ActiveInput{Main: c.Main, TaskDir: c.Config.Full.Documents.Task, Task: named, Env: env,
		EnvCounts: counts})
	if err != nil {
		return &engine.Error{Code: "no-task", Exit: ExitInput, What: "the main checkout's task folder cannot be read (" +
			oneLine(err.Error()) + "), so no task can be climbed for", Next: "check that the folder " +
			shownRel(c.Config.Full.Documents.Task) + " can be read; then run: bonsai ladder " + c.taskFlag() + " --json"}
	}
	if a.How == workspace.HowNamed {
		c.Task = &a
		return nil
	}
	if c.CI && a.Why != workspace.ReasonNamedMissing && a.Why != workspace.ReasonNamedDuplicate &&
		a.Why != workspace.ReasonNamedBroken && a.Why != workspace.ReasonNamedDiffer {
		return nil // --ci names none: every rung, no task
	}
	return noTask(c, a, named != "" || (counts && env != ""))
}

// findRungs works out the requested rungs (note 2) and refuses one bonsai.yaml does not define.
func (c *Climb) findRungs() *engine.Error {
	w := c.Config.Full
	defined := map[int64]bool{}
	for _, r := range w.Ladder {
		defined[r.Rung] = true
	}
	if c.Task == nil {
		for _, r := range w.Ladder {
			c.Requested = appendOnce(c.Requested, r.Rung)
		}
		sortInts(c.Requested)
	} else {
		want, err := stopgate.RungsWanted(w.LadderFloor, c.Task.Task)
		if err != nil {
			return &engine.Error{Code: "rung-not-defined", Exit: ExitInput, What: err.Error() + ", so the rungs to climb cannot be told",
				Next: "fix the task's bonsai.ladder label in the main checkout's " + c.Task.File + " to a list of rung numbers, e.g. [1, 2]; " +
					"then run: bonsai ladder " + c.taskFlag() + " --json"}
		}
		c.Requested = want
	}
	if len(w.Ladder) == 0 {
		return &engine.Error{Code: "rung-not-defined", Exit: ExitInput, What: "bonsai.yaml defines no ladder (its ladder: is empty), so " +
			"there is nothing to climb", Next: "a person writes the project's rungs in bonsai.yaml's ladder; then run: bonsai check --json"}
	}
	var missing []string
	for _, n := range c.Requested {
		if !defined[n] {
			missing = append(missing, strconv.FormatInt(n, 10))
		}
	}
	if len(missing) > 0 {
		where := "the workspace's ladder_floor"
		if c.Task != nil {
			where += " or the task's bonsai.ladder"
		}
		var have []string
		for _, r := range w.Ladder {
			have = append(have, strconv.FormatInt(r.Rung, 10))
		}
		return &engine.Error{Code: "rung-not-defined", Exit: ExitInput,
			What: "rung " + strings.Join(missing, ", ") + ", which " + where + " asks for, is not in bonsai.yaml's ladder (it defines " +
				strings.Join(have, ", ") + "), so nothing was run",
			Next: "see what bonsai.yaml's ladder lacks: run: bonsai check --json; then fix the rung list (the floor and bonsai.yaml " +
				"are a person's, a task's bonsai.ladder its own) and run: bonsai ladder " + c.taskFlag() + " --json"}
	}
	for _, r := range w.Ladder {
		for _, n := range c.Requested {
			if r.Rung == n {
				c.rungs = append(c.rungs, r)
				break
			}
		}
	}
	sortRungs(c.rungs)
	return nil
}

// taskFlag is the --task part of a next step's command line: the task's id when it is known, else a placeholder.
func (c *Climb) taskFlag() string {
	switch {
	case c.Task != nil:
		return "--task " + c.Task.ID
	case c.o.Task != "":
		return "--task " + c.o.Task
	case c.CI:
		return "--ci"
	}
	return "--task <id>"
}

// installedBonsai is the installed bonsai's path on this system, as a next step names it.
func installedBonsai() string {
	if runtime.GOOS == "windows" {
		return `"C:/Program Files/Bonsai/bonsai.exe"`
	}
	return engine.PlaceLinux
}

// samePlace compares two folders, case-blind on Windows.
func samePlace(a, b string) bool {
	if runtime.GOOS == "windows" {
		return strings.EqualFold(filepath.Clean(a), filepath.Clean(b))
	}
	return filepath.Clean(a) == filepath.Clean(b)
}

// shown gives a path as a message prints it: forward slashes, ASCII.
func shown(p string) string { return engine.ASCII(filepath.ToSlash(p)) }

// shownRel gives a project-relative path as a message prints it.
func shownRel(p string) string { return engine.ASCII(p) }

// oneLine keeps the first line of a text.
func oneLine(s string) string {
	if i := strings.IndexAny(s, "\r\n"); i >= 0 {
		s = s[:i]
	}
	return strings.TrimSpace(s)
}

func appendOnce(list []int64, n int64) []int64 {
	for _, x := range list {
		if x == n {
			return list
		}
	}
	return append(list, n)
}

// unexpected wraps an error Bonsai does not expect.
func unexpected(what string, err error) *engine.Error {
	return &engine.Error{Code: "unexpected", Exit: ExitRuntime, What: fmt.Sprintf("%s: %v", what, err),
		Next: "run the same command again, and report it to Bonsai's maintainers if it repeats"}
}
