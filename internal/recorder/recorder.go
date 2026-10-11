// Package recorder is Bonsai's recorder (spec §8; contract §8; design/plan-5.md, 5.2.4): `bonsai hook record`, which
// writes one bonsai.log/1 record for each of ten of Claude Code's hook events, and `bonsai hook start`, which prints a
// session's opening context and writes its session_start record, with the bonsai binary's path and SHA-256.
//
// Where: the project is found from CLAUDE_PROJECT_DIR, else the payload's cwd (never a later cd), as the guard finds
// it (hookin.FindProject): the folder itself or the nearest one above it holding bonsai.yaml, walking past any .git.
// In a folder with no bonsai.yaml both hooks exit 0 at once and write nothing (spec §3, §7). bonsai.yaml is read
// leanly (workspace.LoadConfig: the id, and documents.task for the active task, from the file as read). The records go
// to the main checkout's .bonsai/local/log/ when it is verified (workspace.FindLocal; contract §3: a worktree writes
// main's), else the checkout's own, through internal/record's one append path, so a missing .bonsai/.gitignore is
// restored first.
//
// What each event fills is the table in record.go (eventNames). Every free-text string passes the redactor
// (internal/redact), then its cap (target 200, text 300), then the record's 2,048 bytes (record.Line). Ids are held to
// their pattern instead, and remote to its own; bonsai_path is the binary's own path, from ~/ under the user's home,
// and is not redacted. No prompt's words are kept: a prompt record has its kind only.
//
// They never interfere: `hook record` prints nothing, ever (an async hook's stdout can reach the conversation), and
// both exit 0 whatever happens, a reader of their output gone first among it (pipe_unix.go: SIGPIPE ignored). Each
// runs under its own time limit (Budget), past which it exits 0 with what it has done; neither runs in the frame of
// the hooks that block, since neither blocks.
//
// The payload is read by the one hook adapter (internal/hookin), as every hook reads it. A payload that does not read
// (empty, broken, too large), names no event the recorder writes, or has no session id that can name a log file,
// writes nothing. A payload read with a fault is recorded: each field as Go's lenient decoder reads it (a wrong type
// absent), as the recorder has read payloads since step 5.2.4; a tool input the strict decoder refuses has no hash
// and no target.
//
//   - recorder.go (this file): Options, Record and Start, the project and its config.
//   - record.go: one payload as one record: the events, the categories, the input hash.
//   - start.go: bonsai hook start's opening context and the binary's path and hash.
//   - pipe_unix.go, pipe_other.go: SIGPIPE ignored, so a write to a closed pipe is an error, not the process's end.
//
// At a session's end, after its session_end line is written, `hook record` runs the cleaner (internal/clean,
// SessionEnd: the main checkout's ladder results, asks and log by bonsai.yaml's generated: rules, within 1 s of the
// line's 5 s timeout and the hook's own 4 s budget). A session_end line that cannot be written cleans nothing.
package recorder

import (
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/LastStep/Bonsai/internal/clean"
	"github.com/LastStep/Bonsai/internal/hookin"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// Budget is each hook's own time limit. SessionEnd's line has a timeout of 5 s, which also raises Claude Code's 1.5 s
// budget for SessionEnd hooks; hook start's has 10 s. Past Budget the hook exits 0 with what it has done.
const Budget = 4 * time.Second

// Options are what a hook needs besides its payload.
type Options struct {
	Stdin         io.Reader           // the payload
	StdinTerminal bool                // stdin is a terminal: run by hand with no payload piped in
	Getenv        func(string) string // os.Getenv, or a test's own
	Now           func() time.Time    // the clock; nil means time.Now
	Budget        time.Duration       // the hook's own time limit; 0 means Budget
	Executable    func() (string, error)
	UserHome      func() (string, error)
}

func (o *Options) defaults() {
	if o.Getenv == nil {
		o.Getenv = os.Getenv
	}
	if o.Now == nil {
		o.Now = time.Now
	}
	if o.Budget <= 0 {
		o.Budget = Budget
	}
	if o.Executable == nil {
		o.Executable = os.Executable
	}
	if o.UserHome == nil {
		o.UserHome = os.UserHomeDir
	}
}

// Record is one run of `bonsai hook record`: it writes the payload's record and returns 0, always. It writes nothing
// on stdout (it is given none).
func Record(o Options) int {
	ignoreBrokenPipe()
	o.defaults()
	within(o.Budget, func() { work(o) })
	return 0
}

// within runs work under the budget, a panic inside it caught: the hook returns either way.
func within(budget time.Duration, work func()) {
	done := make(chan struct{})
	go func() {
		defer close(done)
		defer func() { _ = recover() }()
		work()
	}()
	timer := time.NewTimer(budget)
	defer timer.Stop()
	select {
	case <-done:
	case <-timer.C:
	}
}

// work does Record's whole job.
func work(o Options) {
	p, proj, ok := payloadAndProject(o)
	if !ok || p == nil || !recorded(p.Event) {
		return
	}
	cfg, err := workspace.LoadConfig(proj)
	if err != nil {
		return // no workspace id to record under; bonsai check names the file's problem
	}
	local := workspace.FindLocal(proj, cfg.ID)
	l := build(o, p, proj, cfg, local, &active{o: o, cfg: cfg, local: local})
	if l == nil {
		return
	}
	if _, err := writeLog(local.Main, l); err != nil {
		return // a log that cannot be written takes no clean record either: nothing is cleaned
	}
	if p.Event == "SessionEnd" {
		// After the session_end line, the cleaning (internal/clean): ladder results, asks and the log of the main
		// checkout's local/, within clean.SessionEndBudget; the clean records name this session's checkout.
		_ = cleanAtEnd(clean.Options{Local: local, Now: o.Now, Getenv: o.Getenv})
	}
}

// cleanAtEnd is the cleaning at a session's end (a variable for the tests).
var cleanAtEnd = clean.SessionEnd

// payloadAndProject reads the payload and finds the project, in the order spec §7 asks: with CLAUDE_PROJECT_DIR set,
// the project first, so an unlinked folder costs nothing but draining stdin; without it, the payload's cwd. ok is
// false when there is no linked project; p is nil when the payload does not read (readStdin).
func payloadAndProject(o Options) (p *hookin.Payload, proj string, ok bool) {
	if dir := o.Getenv(hookin.ProjectEnv); dir != "" {
		root, linked, err := hookin.FindProject(dir)
		if err != nil || !linked {
			hookin.Drain(o.Stdin, o.StdinTerminal)
			return nil, "", false
		}
		return readStdin(o), root, true
	}
	p = readStdin(o)
	if p == nil || p.Cwd == "" || !filepath.IsAbs(p.Cwd) {
		return nil, "", false
	}
	root, linked, err := hookin.FindProject(p.Cwd)
	return p, root, linked && err == nil
}

// readStdin reads the payload through the hook adapter, as the recorder takes it: nil when it does not read, or names
// no event, or has no session id that can name a log file (there is nothing to record); a fault is passed over
// (this package's comment).
func readStdin(o Options) *hookin.Payload {
	p, err := hookin.Read(o.Stdin, o.StdinTerminal)
	if err != nil || p.Event == "" || p.Session == "" {
		return nil
	}
	return p
}
