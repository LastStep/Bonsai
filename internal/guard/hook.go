package guard

// Main: one run of `bonsai hook guard`, from its own timer to its exit code.

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/LastStep/Bonsai/internal/workspace"
)

// Budget is the guard's own time limit: half the 10 s timeout part 3's hook line gives it (engine.GuardTimeout),
// leaving the rest for the shell, the process start and a slow disk. A hook that times out does not block (spec
// §7), so the guard blocks itself before that can happen.
const Budget = 5 * time.Second

// ProjectEnv names the variable Claude Code sets for a hook to the project's folder, the one the session started in.
// The guard finds the project from it, never from the payload's cwd (an agent's cd moves that) or from git (whose
// answer comes from files an agent may edit).
const ProjectEnv = "CLAUDE_PROJECT_DIR"

// recordWait caps the time the over-time answer spends writing its record.
const recordWait = 500 * time.Millisecond

// Options are what Main needs besides the payload.
type Options struct {
	Stdin         io.Reader           // the payload
	Stderr        io.Writer           // a refusal's reason and next step
	StdinTerminal bool                // stdin is a terminal: the guard was run by hand with no payload piped in
	Getenv        func(string) string // os.Getenv, or a test's own
	Budget        time.Duration       // the guard's own time limit; 0 means Budget
	Now           func() time.Time    // the clock; nil means time.Now
}

// run is one guard run's state, shared by the worker and the timer under mu.
type run struct {
	o    Options
	mu   sync.Mutex
	root string
	cfg  *workspace.Config
	in   *Input
}

// Main answers one PreToolUse call: 0 allows it, 2 blocks it with the reason and the next step on stderr. It
// returns by its budget whatever the work is doing; the caller then exits, which stops the work.
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
	timer := time.NewTimer(o.Budget)
	defer timer.Stop()
	done := make(chan Decision, 1)
	go func() { done <- r.work() }()
	select {
	case d := <-done:
		return r.answer(d)
	case <-timer.C:
		return r.answer(r.overTime())
	}
}

// answer prints a refusal and gives the exit code. A refusal that cannot be printed still blocks: the code is the
// answer.
func (r *run) answer(d Decision) int {
	if d.Allow {
		return 0
	}
	_, _ = fmt.Fprintf(r.o.Stderr, "bonsai guard: %s.\nnext: %s.\n", d.Why, d.Next)
	return 2
}

// work does the guard's whole job and gives its decision. A panic inside becomes a refusal.
func (r *run) work() (d Decision) {
	defer func() {
		if p := recover(); p != nil {
			d = deny(RuleInternal, fmt.Sprintf("the guard failed inside (%s), and a guard that fails blocks the call",
				oneLine(fmt.Sprint(p))), "tell the person; a person runs `bonsai hook guard` by hand with this payload on stdin to see the failure")
		}
	}()
	dir := r.o.Getenv(ProjectEnv)
	if dir == "" {
		return deny(RuleNoProject, ProjectEnv+" is not set, so the guard cannot tell which project this call is in, and it blocks",
			"run the guard as Claude Code's hook (Claude Code sets "+ProjectEnv+"); to run it by hand, set "+ProjectEnv+" to the project's folder")
	}
	root, linked, err := findProject(dir)
	if err != nil {
		return deny(RuleNoProject, fmt.Sprintf("the project's folder %s cannot be read (%s), so the guard blocks",
			quote(filepath.ToSlash(dir), 120), oneLine(err.Error())), "tell the person: check that the folder exists and can be read")
	}
	if !linked {
		r.drain()
		return allow(RuleNotLinked, "")
	}
	cfg, err := workspace.LoadConfig(root)
	if err != nil {
		next := "leave the change and tell the person: a person fixes bonsai.yaml (bonsai check names the line)"
		var we *workspace.Error
		if errors.As(err, &we) {
			next = "tell the person: " + we.Next
		}
		return deny(RuleBadConfig, "bonsai.yaml cannot be read ("+oneLine(err.Error())+"), so the guard cannot tell what "+
			"is protected, and it blocks", oneLine(next))
	}
	r.mu.Lock()
	r.root, r.cfg = root, cfg
	r.mu.Unlock()
	in, err := r.readInput()
	if err != nil {
		d = deny(RuleBadInput, "the hook input "+err.Error()+", and an input the guard cannot read is blocked",
			"tell the person: a person checks the payload Claude Code sends (bonsai hook guard, by hand, with it on stdin)")
		return r.recorded(d)
	}
	r.mu.Lock()
	r.in = in
	r.mu.Unlock()
	if fd := testFault(r); fd != nil {
		return r.recorded(*fd)
	}
	return r.recorded(Decide(in, root, cfg))
}

// recorded writes d's record and gives the decision to answer: an allow that cannot be recorded becomes a refusal.
func (r *run) recorded(d Decision) Decision {
	err := r.record(d)
	if err == nil {
		return d
	}
	if d.Allow {
		return deny(RuleLogFailed, fmt.Sprintf("the guard would allow this call but cannot write its record in %s (%s), "+
			"and a call it cannot record is blocked", LogDir, oneLine(err.Error())),
			"tell the person: a person makes "+LogDir+"/ writable (or moves aside what is in its way), then the call is tried again")
	}
	d.Why += fmt.Sprintf(" (its record in %s could not be written either: %s)", LogDir, oneLine(err.Error()))
	return d
}

// record writes d's record with what the run has read so far.
func (r *run) record(d Decision) error {
	r.mu.Lock()
	root, cfg, in := r.root, r.cfg, r.in
	r.mu.Unlock()
	if cfg == nil {
		return errors.New("no workspace id to record")
	}
	return writeRecord(root, cfg, in, d, r.o.Getenv, r.o.Now())
}

// overTime is the answer when the budget runs out first: a refusal, recorded if that can be done in recordWait.
func (r *run) overTime() Decision {
	d := deny(RuleOverTime, fmt.Sprintf("the guard ran past its own %s limit, so it blocks the call rather than let "+
		"the hook's timeout pass it through", r.o.Budget),
		"try the call again; if it is blocked again, tell the person (a slow disk, or a stdin that never ends)")
	r.mu.Lock()
	known := r.cfg != nil
	r.mu.Unlock()
	if known {
		wrote := make(chan struct{})
		go func() {
			_ = r.record(d)
			close(wrote)
		}()
		select {
		case <-wrote:
		case <-time.After(recordWait):
		}
	}
	return d
}

// readInput reads and parses the payload. Its error follows "the hook input".
func (r *run) readInput() (*Input, error) {
	if r.o.StdinTerminal {
		return nil, errors.New("is a terminal, not a payload (the guard reads Claude Code's PreToolUse payload on stdin)")
	}
	raw, err := io.ReadAll(io.LimitReader(r.o.Stdin, MaxInput+1))
	if err != nil {
		return nil, fmt.Errorf("cannot be read (%s)", oneLine(err.Error()))
	}
	if len(raw) > MaxInput {
		return nil, fmt.Errorf("is larger than %d MiB", MaxInput>>20)
	}
	return ParseInput(raw)
}

// drain reads and drops the payload of a project that is not linked, so Claude Code never writes into a closed
// pipe.
func (r *run) drain() {
	if r.o.StdinTerminal || r.o.Stdin == nil {
		return
	}
	_, _ = io.Copy(io.Discard, io.LimitReader(r.o.Stdin, MaxInput))
}

// findProject finds the project a session's folder belongs to: the folder itself or the nearest one above it that
// holds bonsai.yaml, looking no higher than a checkout's top (a folder holding .git). linked is false when there is
// none (spec §7: in a project with no bonsai.yaml the guard allows and records nothing).
func findProject(dir string) (root string, linked bool, err error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", false, err
	}
	if fi, err := os.Stat(abs); err != nil {
		return "", false, err
	} else if !fi.IsDir() {
		return "", false, errors.New("it is not a folder")
	}
	for d := abs; ; {
		if _, err := os.Lstat(filepath.Join(d, workspace.ConfigFile)); err == nil {
			return d, true, nil
		} else if !errors.Is(err, fs.ErrNotExist) {
			return "", false, err
		}
		if _, err := os.Lstat(filepath.Join(d, ".git")); err == nil {
			return d, false, nil
		} else if !errors.Is(err, fs.ErrNotExist) {
			return "", false, err
		}
		parent := filepath.Dir(d)
		if parent == d {
			return abs, false, nil
		}
		d = parent
	}
}
