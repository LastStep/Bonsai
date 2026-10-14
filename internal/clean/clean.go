// Package clean is Bonsai's cleaner (spec §6, "Generated files"; contract §7.5, §8.2, §8.5, §9.1, §11;
// design/plan-5.md, 5.2.6b): it deletes the generated files and table rows that bonsai.yaml's generated: rules say
// may go, and nothing else, and logs each one it deletes as a clean record.
//
// The kinds it cleans (format.GeneratedKinds is their one home; skill base:generated-files says the same in plain
// words): log files (.bonsai/local/log/), asks day files (.bonsai/local/asks/) and ladder results
// (.bonsai/local/ladder/), by file; the messages between agents (.bonsai/local/messages/), by thread; and the sessions
// table's rows (.bonsai/sessions.md), by row. Run reports are never deleted by Bonsai, and the tasks table is a
// rebuild.
//
// The rule, per kind (generated.<kind> in the main checkout's bonsai.yaml, read leanly from the file as read: rule.go):
// a file or row goes when it is older than keep_days, or beyond the newest keep_newest of its kind, either one
// cleaning; null keeps. A kind left out, or generated: left out, takes its default: format.GeneratedKinds holds each
// kind's default numbers, their one home, and skill base:generated-files prints them in words. Age is a log or asks
// file's last record's at (its modification time when no record reads), a ladder result's finished (its modification
// time when the result does not read), a thread's last message's created, and a row's end.
//
// The protections come first and always win, and a protected file or row still counts among the newest:
//   - a log file holding an open span (internal/sessions: the one rule that says "open"), or an ended session or
//     subagent run with no row yet in .bonsai/sessions.md (rows before files: only check --write adds them);
//   - an asks day file holding the file record of an open ask (internal/asks: a key's state), or the answer or
//     withdrawal of an ask whose filing stays (else that ask would read open again);
//   - a ladder result whose task is not done or cut, read from the main checkout's task files: a task not found, or
//     one whose file does not read, counts as not done;
//   - a row whose task is not done or cut (none is never protected), or whose span is still in the log (check --write
//     would only add it again).
//   - an open thread of messages (a brief with no report after it), whatever its age (messages.go).
//
// Only Bonsai's own files: the names it writes (s-<session>.ndjson and w-<date>.ndjson in log/, <date>.ndjson in
// asks/, <task id>.json in ladder/), regular files only, at the top of the kind's folder in the main checkout's
// .bonsai/local/, that folder and the two above it real folders, not links (a link is never followed); and messages
// that read, each <id>.md in its recipient's folder of messages/ (internal/message's reader). Anything else is left
// alone and never named.
//
// When: at a session's end, after its session_end line (internal/recorder, SessionEnd: ladder, asks, messages and log
// within SessionEndBudget, oldest first; the budget is checked before each step of the listing and the judging, a run
// ends within it and one step, and what it leaves waits for the next end: files.go); in check --write, the sessions
// rows (internal/engine, Rows and Record: the table's only writer); and by the ladder runner after each run (step 5.4,
// Ladder). No process outlives its caller. The cleaner reads the main checkout's local/ only (workspace.FindLocal).
//
// The clean record (contract §8.2): one per file or row, written after the delete succeeded (for a row, after the
// table was written), through internal/record's one append path, in today's day file with session null: target the
// project-relative path (a row: .bonsai/sessions.md# and its key), reason the rule (generated.log.keep_days=30). A
// delete that finds the file gone writes nothing (another cleaner took it). A file Windows reports held open by
// another process is skipped after one try and cleaned at a later run; one that changed since it was judged is left
// for a later run too (an asks day file left so stops that run's asks there: files.go). A cleaning killed between a
// delete and its record leaves the file gone with no record: a run ends within its budget and one step, so the session
// end's SessionEndBudget keeps it well inside SessionEnd's line's 5 s (Claude Code sends SIGTERM at the line's
// timeout, measured in step 5.2.4), and a record that cannot be written stops the run, so no further file goes
// unrecorded.
//
//   - clean.go (this file): Options, the run, the delete and its record.
//   - rule.go: generated.<kind> read from bonsai.yaml, and the reasons.
//   - files.go: a kind's folder, its files' ages, the order they are judged in, and the budget.
//   - log.go, asks.go, ladder.go: each file kind's names, clock and protections.
//   - messages.go: the messages kind, by thread.
//   - rows.go: the sessions table's rows, for check --write.
package clean

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/LastStep/Bonsai/internal/record"
	"github.com/LastStep/Bonsai/internal/redact"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// The kinds the cleaner cleans: format.GeneratedKinds' names.
const (
	Log      = "log"
	Asks     = "asks"
	Ladder   = "ladder"
	Messages = "messages"
	Sessions = "sessions"
)

// LadderFolder is the ladder results' folder inside .bonsai/local/ (contract §11), beside record.LogFolder and
// record.AsksFolder.
const LadderFolder = "ladder"

// SessionEndBudget is how long the cleaning at a session's end may take. SessionEnd's line has a timeout of 5 s, at
// which Claude Code sends SIGTERM (measured in step 5.2.4), and the recorder's own budget is 4 s; the session_end
// line takes a few milliseconds, and a run ends within its budget and one step (files.go), so 1 s of cleaning leaves
// about 3 s for that step, its longest single read.
const SessionEndBudget = time.Second

// Options are where and how a run cleans.
type Options struct {
	// Local is where: Local.Main's .bonsai/local/ is cleaned, by the rules of Local.Main's bonsai.yaml, and Local.Root
	// is the checkout the clean records name (the session's checkout at a session's end; the main checkout in check
	// --write). workspace.FindLocal gives it.
	Local workspace.Local
	// Now is the clock ages are judged by and records stamped with; nil reads time.Now.
	Now func() time.Time
	// Getenv reads BONSAI_TASK and BONSAI_ROLE, the records' task and role (redacted); nil reads the process's own.
	Getenv func(string) string
	// Budget is how long the run may take, by the wall clock; 0 for no limit. It is checked before each step (a batch
	// of a folder's listing, a file's age for keep_newest, an asks day file read, a file judged and deleted: files.go),
	// so a run ends within it and one step. Past it, no new step begins, and the rest waits for the next run.
	Budget time.Duration
	// Keep are project-relative paths (forward slashes) this run never cleans: the ladder runner's own result.
	Keep []string
}

// Item is one file or row cleaned, or left for a later run.
type Item struct {
	Kind   string // log, asks, ladder, messages or sessions
	Target string // the project-relative path, forward slashes; a row: .bonsai/sessions.md# and its key
	Reason string // the rule: generated.log.keep_days=30
}

// Done is what a run did.
type Done struct {
	Cleaned []Item  // deleted, each with its clean record written, in the order deleted
	Busy    []Item  // held open by another process (Windows) after one try: left for a later run
	Stopped bool    // the budget ran out: what was not judged waits for the next run (not an error)
	Errs    []error // what could not be read, deleted or recorded; a file that cannot be judged is kept
}

// removeFile deletes one file, once (a variable for the tests: a file gone, or held, at the moment of the delete).
var removeFile = os.Remove

// busy tells an error that says another process holds the file open (Windows' sharing violation; a variable for the
// tests, since no other system has one).
var busy = workspace.IsBusy

// writeLog appends a clean record (a variable for the tests: a record that cannot be written).
var writeLog = record.WriteLog

// clock is the wall clock a run's budget is measured by, apart from Options.Now, the clock of ages (a variable for the
// tests: a clock the test moves, so where a run stops depends on the code, not on the machine).
var clock = time.Now

// run is one run's state: its options, the main checkout's config, and what it read once.
type run struct {
	o        Options
	main     string
	cfg      *workspace.Config
	now      time.Time // ages are judged against this one instant
	deadline time.Time // the zero time for none
	done     *Done
	keep     map[string]bool
	halt     bool // a record could not be written: nothing more is deleted

	rows    map[string]bool // the sessions table's row keys (log.go), read once
	tasks   *taskStates     // the main checkout's task files (ladder.go), read once
	asks    *asksHistory    // the asks folder's records (asks.go), read once
	asksErr error
}

// start sets a run up: the main checkout's bonsai.yaml read leanly (its id, its generated: rules, its task folder).
// nil when it does not read: nothing is cleaned, and bonsai check names the file's problem.
func start(o Options) (*run, *Done) {
	d := &Done{}
	if o.Now == nil {
		o.Now = time.Now
	}
	r := &run{o: o, main: o.Local.Main, now: o.Now(), done: d, keep: map[string]bool{}}
	if o.Budget > 0 {
		r.deadline = clock().Add(o.Budget)
	}
	for _, k := range o.Keep {
		r.keep[k] = true
	}
	if r.main == "" {
		d.Errs = append(d.Errs, errors.New("clean: no main checkout given"))
		return nil, d
	}
	cfg, err := workspace.LoadConfig(r.main)
	if err != nil {
		d.Errs = append(d.Errs, err)
		return nil, d
	}
	r.cfg = cfg
	return r, d
}

// Files cleans the files of the kinds given (Log, Asks, Ladder, Messages), each in turn, within the budget. A kind
// whose rule keeps (null) reads nothing. It never fails: what it could not do is in Done.
func Files(o Options, kinds ...string) *Done {
	r, d := start(o)
	if r == nil {
		return d
	}
	for _, k := range kinds {
		if r.out() || r.halt {
			break
		}
		switch k {
		case Log:
			r.files(logKind)
		case Asks:
			r.files(asksKind)
		case Ladder:
			r.files(ladderKind)
		case Messages:
			r.messages()
		default:
			d.Errs = append(d.Errs, fmt.Errorf("clean: %q is not a kind of file the cleaner cleans", k))
		}
	}
	return d
}

// SessionEnd is the cleaning at a session's end (internal/recorder, after the session_end line): ladder results,
// asks, the messages between agents and the log, the cheap kinds first, within SessionEndBudget unless o.Budget is
// set.
func SessionEnd(o Options) *Done {
	if o.Budget <= 0 {
		o.Budget = SessionEndBudget
	}
	return Files(o, Ladder, Asks, Messages, Log)
}

// LadderResults is the ladder runner's call after each run (step 5.4): the ladder kind alone. The runner names its own
// result in o.Keep, so the result a run wrote is never cleaned by that run.
func LadderResults(o Options) *Done { return Files(o, Ladder) }

// out reports whether the budget has run out, and marks the run stopped when it has. Each step of a run asks it
// first (files.go).
func (r *run) out() bool {
	if r.deadline.IsZero() || clock().Before(r.deadline) {
		return false
	}
	r.done.Stopped = true
	return true
}

// fail notes what the run could not do. The budget run out in the middle of a read (errStopped) is not one: Stopped
// says it.
func (r *run) fail(err error) {
	if errors.Is(err, errStopped) {
		return
	}
	r.done.Errs = append(r.done.Errs, err)
}

// remove deletes one judged file and writes its clean record. The file is looked at once more first: one that is
// gone writes nothing; one that is no longer a regular file, or changed since it was listed (a session writing it),
// is left alone. A file Windows reports held open is skipped after one try. gone reports the file gone, by this delete
// or another cleaner's; false when it is still there.
func (r *run) remove(f *file) (gone bool) {
	fi, err := os.Lstat(f.path)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return true
	case err != nil:
		r.fail(fmt.Errorf("clean: %s cannot be read: %w", f.target, err))
		return false
	case !fi.Mode().IsRegular() || !fi.ModTime().Equal(f.mtime) || fi.Size() != f.size:
		return false
	}
	err = removeFile(f.path)
	switch {
	case err == nil:
	case errors.Is(err, fs.ErrNotExist):
		return true // another session's cleaner took it
	case busy(err):
		r.done.Busy = append(r.done.Busy, Item{Kind: f.kind, Target: f.target, Reason: f.reason})
		return false
	default:
		r.fail(fmt.Errorf("clean: %s cannot be deleted: %w", f.target, err))
		return false
	}
	it := Item{Kind: f.kind, Target: f.target, Reason: f.reason}
	if err := r.record(it); err != nil {
		r.fail(fmt.Errorf("clean: %s was deleted, but its clean record cannot be written: %w", f.target, err))
		r.halt = true
		return true
	}
	r.done.Cleaned = append(r.done.Cleaned, it)
	return true
}

// record writes one clean record: in today's day file (session null), target and reason filled.
func (r *run) record(it Item) error {
	l := record.New("clean", record.Common{Workspace: r.cfg.ID, Local: r.o.Local, Getenv: r.o.Getenv,
		Redact: redact.Text, At: r.o.Now()})
	target, reason := it.Target, it.Reason
	l.Target, l.Reason = &target, &reason
	_, err := writeLog(r.main, l)
	return err
}

// Record writes the clean records of rows check --write took out of the sessions table (Rows), once the table is
// written. It stops at the first record that cannot be written.
func Record(o Options, items []Item) error {
	if len(items) == 0 {
		return nil
	}
	r, d := start(o)
	if r == nil {
		return errors.Join(d.Errs...)
	}
	for _, it := range items {
		if err := r.record(it); err != nil {
			return err
		}
	}
	return nil
}

// target is a file's project-relative path, forward slashes: .bonsai/local/<folder>/<name>.
func target(folder, name string) string {
	return workspace.LocalDir + "/" + folder + "/" + name
}

// localPath is the absolute path of a folder inside the main checkout's .bonsai/local/.
func (r *run) localPath(folder string) string {
	return filepath.Join(r.main, filepath.FromSlash(workspace.LocalDir), folder)
}
