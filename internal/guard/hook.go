package guard

// Main: one run of `bonsai hook guard`, from its own timer to its exit code, in the frame of the hooks that block
// (internal/hookin: the timer first, the work in a goroutine, a panic as a block, the answer and its exit code).

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/LastStep/Bonsai/internal/hookin"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// Budget is the guard's own time limit: half the 10 s timeout part 3's hook line gives it (engine.GuardTimeout),
// leaving the rest for the shell, the process start and a slow disk. A hook that times out does not block (spec
// §7), so the guard blocks itself before that can happen.
const Budget = 5 * time.Second

// Options are what Main needs besides the payload.
type Options struct {
	Stdin         io.Reader           // the payload
	Stderr        io.Writer           // a refusal's reason and next step
	StdinTerminal bool                // stdin is a terminal: the guard was run by hand with no payload piped in
	Getenv        func(string) string // os.Getenv, or a test's own
	Budget        time.Duration       // the guard's own time limit; 0 means Budget
	RecordWait    time.Duration       // how long the over-time refusal waits for its record; 0 means hookin.RecordWait
	Now           func() time.Time    // the clock; nil means time.Now
}

// run is one guard run's state, shared by the worker and the timer under mu.
type run struct {
	o    Options
	mu   sync.Mutex
	root string
	cfg  *workspace.Config
	in   *hookin.Payload
}

// Main answers one PreToolUse call: 0 allows it, 2 blocks it with the reason and the next step on stderr. It
// returns by its budget (and the over-time record's wait) whatever the work is doing; the caller then exits, which
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
	f := hookin.Frame{Name: "guard", Budget: o.Budget, Wait: o.RecordWait, Stderr: o.Stderr}
	return f.Run(hookin.Job{
		Work:     func() hookin.Answer { return r.work().answer() },
		Panic:    func(p any) hookin.Answer { return panicked(p).answer() },
		OverTime: r.overTime,
	})
}

// answer gives a decision as the frame's answer.
func (d Decision) answer() hookin.Answer {
	return hookin.Answer{Block: !d.Allow, Why: d.Why, Next: d.Next}
}

// panicked is the refusal of a guard that failed inside.
func panicked(p any) Decision {
	return deny(RuleInternal, fmt.Sprintf("the guard failed inside (%s), and a guard that fails blocks the call",
		oneLine(fmt.Sprint(p))), "tell the person; a person runs `bonsai hook guard` by hand with this payload on stdin to see the failure")
}

// work does the guard's whole job and gives its decision. A panic inside is the frame's to answer (panicked).
func (r *run) work() (d Decision) {
	dir := r.o.Getenv(hookin.ProjectEnv)
	if dir == "" {
		return deny(RuleNoProject, hookin.ProjectEnv+" is not set, so the guard cannot tell which project this call is in, and it blocks",
			"run the guard as Claude Code's hook (Claude Code sets "+hookin.ProjectEnv+"); to run it by hand, set "+hookin.ProjectEnv+" to the project's folder")
	}
	root, linked, err := hookin.FindProject(dir)
	if err != nil {
		return deny(RuleNoProject, fmt.Sprintf("the project's folder %s cannot be read (%s), so the guard blocks",
			quote(filepath.ToSlash(dir), 120), oneLine(err.Error())), "tell the person: check that the folder exists and can be read")
	}
	if !linked {
		hookin.Drain(r.o.Stdin, r.o.StdinTerminal)
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

// overTime is the answer when the budget runs out first: a refusal, and, once bonsai.yaml has been read, the writer
// of its record, which the frame awaits for the run's RecordWait. If the work was itself writing a record just then,
// the log holds both, and the later over-time one is the answer that counted.
func (r *run) overTime() (hookin.Answer, func()) {
	d := deny(RuleOverTime, fmt.Sprintf("the guard ran past its own %s limit, so it blocks the call rather than let "+
		"the hook's timeout pass it through", r.o.Budget),
		"try the call again; if it is blocked again, tell the person (a slow disk, or a stdin that never ends)")
	r.mu.Lock()
	known := r.cfg != nil
	r.mu.Unlock()
	if !known {
		return d.answer(), nil
	}
	return d.answer(), recordOverTime(func() { _ = r.record(d) })
}

// recordOverTime hands the over-time refusal's record writer to the frame (a variable for the tests, which see it
// handed over, and hold it to see the answer come back within its wait).
var recordOverTime = func(write func()) func() { return write }

// readInput reads the payload through the hook adapter. Its error follows "the hook input": a payload that does not
// read, a fault in a value the guard reads, or an event other than PreToolUse.
func (r *run) readInput() (*hookin.Payload, error) {
	p, err := hookin.Read(r.o.Stdin, r.o.StdinTerminal)
	if err != nil {
		return nil, err
	}
	if p.Fault != nil {
		return nil, p.Fault
	}
	if p.Event != "PreToolUse" {
		return nil, fmt.Errorf("is a %s payload: the guard answers PreToolUse only", strconv.QuoteToASCII(p.Event))
	}
	return p, nil
}

// quote gives a value as an ASCII Go string, cut to max runes.
func quote(s string, max int) string {
	r := []rune(s)
	if len(r) > max {
		s = string(r[:max]) + "..."
	}
	return strconv.QuoteToASCII(s)
}

// oneLine keeps a message on one ASCII line: printable ASCII as itself, a line break as a space, any other
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
