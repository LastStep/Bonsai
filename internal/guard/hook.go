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

	// Second runs the guard as its second line (bonsai hook guard --second; engine.SecondCommand, where the lock declares
	// roles): on every tool the first line does not match, it refuses only past an agent's budget (budget-spent) and a
	// Skill that forks into a type its caller may not call (call-refused, skill.go), and lets the call through on any
	// fault of its own (second-line-fault), as its line does while bonsai is missing.
	Second bool
}

// run is one guard run's state, shared by the worker and the timer under mu.
type run struct {
	o  Options
	mu sync.Mutex
	at *place // where the record goes: the session's project, then the project whose rules decided; nil until known
	in *hookin.Payload
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
		Work: func() hookin.Answer { return r.work().answer() },
		Panic: func(p any) hookin.Answer {
			if o.Second {
				return hookin.Answer{} // a fault of its own: the second line lets the call through
			}
			return panicked(p).answer()
		},
		OverTime: r.overTime,
	})
}

// secondRefuses reports the refusals the second line answers as refusals: past a budget (budget-spent), and against
// a role rule (call-refused: a Skill whose frontmatter forks into a type its caller may not call, skill.go). Every
// other refusal there is a fault of its own, and lets the call through (secondFault).
func secondRefuses(rule string) bool { return rule == RuleBudgetSpent || rule == RuleCallRefused }

// secondFault is the second line's answer in place of a refusal other than its own two (secondRefuses): the call goes
// through, and the record says why it was not judged (second-line-fault).
func secondFault(d Decision) Decision {
	f := allow(RuleSecondLineFault, d.Target)
	f.Why = "the guard's second line could not judge this call (" + d.Rule + ": " + d.Why + "), so it lets it through"
	return f
}

// faultFor gives the decision a run answers with: the second line turns a refusal other than its own two
// (secondRefuses) into second-line-fault.
func (r *run) faultFor(d Decision) Decision {
	if r.o.Second && !d.Allow && !secondRefuses(d.Rule) {
		return secondFault(d)
	}
	return d
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
//
// The order (spec §7): the session's project first, from CLAUDE_PROJECT_DIR (set by Claude Code, not by any file an
// agent edits), so a bonsai.yaml that does not read blocks before anything else; then the payload; then the decision,
// by the projects holding the path (rule.go). When the session's folder is in no linked project, the guard still
// judges a file tool's path in a linked one, and allows and records nothing otherwise (spec §7).
func (r *run) work() (d Decision) {
	defer func() {
		if r.o.Second && !d.Allow && !secondRefuses(d.Rule) {
			d = allow(RuleSecondLineFault, "") // a fault before any record could be written: let through, unrecorded
		}
	}()
	dir := r.o.Getenv(hookin.ProjectEnv)
	if dir == "" {
		return deny(RuleNoProject, hookin.ProjectEnv+" is not set, so the guard cannot tell which project this call is in, and it blocks",
			"run the guard as Claude Code's hook (Claude Code sets "+hookin.ProjectEnv+"); to run it by hand, set "+hookin.ProjectEnv+" to the project's folder")
	}
	roots, base, err := hookin.FindProjects(dir)
	if err != nil {
		return deny(RuleNoProject, fmt.Sprintf("the project's folder %s cannot be read (%s), so the guard blocks",
			quote(filepath.ToSlash(dir), 120), oneLine(err.Error())), "tell the person: check that the folder exists and can be read")
	}
	j := newJudge(guardHome())
	j.getenv, j.now, j.second = r.o.Getenv, r.o.Now, r.o.Second
	if len(roots) > 0 {
		s := j.project(roots[0])
		if err := s.own.err; err != nil {
			next := "leave the change and tell the person: a person fixes bonsai.yaml (bonsai check names the line)"
			var we *workspace.Error
			if errors.As(err, &we) {
				next = "tell the person: " + we.Next
			}
			return deny(RuleBadConfig, "bonsai.yaml cannot be read ("+oneLine(err.Error())+"), so the guard cannot tell what "+
				"is protected, and it blocks", oneLine(next))
		}
		j.session, base = s, s.root
		r.setPlace(j, s)
	}
	in, err := r.readInput()
	if err != nil {
		if j.session == nil {
			return allow(RuleNotLinked, "")
		}
		d = deny(RuleBadInput, "the hook input "+err.Error()+", and an input the guard cannot read is blocked",
			"tell the person: a person checks the payload Claude Code sends (bonsai hook guard, by hand, with it on stdin)")
		return r.recorded(d)
	}
	r.mu.Lock()
	r.in = in
	r.mu.Unlock()
	if j.session != nil {
		if fd := testFault(r); fd != nil {
			return r.recorded(*fd)
		}
	}
	d, rec := j.decide(in, base)
	r.extraRecords(j)
	if rec == nil {
		if j.session == nil && d.Rule == RuleBadPath {
			return allow(RuleNotLinked, "") // in no linked project that can be told: a write through it fails anyway
		}
		return d
	}
	r.setPlace(j, rec)
	return r.recorded(d)
}

// extraRecords writes the records the roles' part kept beside the call's own (roles.go: a brief that ties nothing, a
// state that cannot be read, roles read from the files, a spawn record), in the session's project. One that cannot be
// written is dropped: the call's own record still says what was decided.
func (r *run) extraRecords(j *judge) {
	if j.role == nil || j.session == nil || len(j.role.records) == 0 {
		return
	}
	at := &place{id: j.session.id(), local: j.local(j.session)}
	for _, x := range j.role.records {
		if x.spawn != nil {
			_ = writeSpawn(at, x.in, x.spawn, r.o.Getenv, x.atTime)
			continue
		}
		_ = writeRecord(at, x.in, x.d, r.o.Getenv, x.atTime)
	}
}

// setPlace makes p the project the run's record goes in.
func (r *run) setPlace(j *judge, p *linkedProject) {
	at := &place{id: p.id(), local: j.local(p)}
	r.mu.Lock()
	r.at = at
	r.mu.Unlock()
}

// guardHome is the Bonsai home the guard reads this machine's records from (workspace.Home: BONSAI_HOME as the hook
// sees it, else ~/.bonsai); "" when there is none, which verifies no main checkout.
func guardHome() string {
	home, err := workspace.Home()
	if err != nil {
		return ""
	}
	return home
}

// recorded writes d's record and gives the decision to answer: an allow that cannot be recorded becomes a refusal
// (on the first line; the second lets it through).
func (r *run) recorded(d Decision) Decision {
	d = r.faultFor(d)
	err := r.record(d)
	if err == nil {
		return d
	}
	if d.Allow {
		if r.o.Second {
			return d // the second line lets a call through on a fault of its own, its record's among them
		}
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
	at, in := r.at, r.in
	r.mu.Unlock()
	if at == nil {
		return errors.New("no workspace id to record")
	}
	return writeRecord(at, in, d, r.o.Getenv, r.o.Now())
}

// overTime is the answer when the budget runs out first: a refusal, and, once bonsai.yaml has been read, the writer
// of its record, which the frame awaits for the run's RecordWait. If the work was itself writing a record just then,
// the log holds both, and the later over-time one is the answer that counted.
func (r *run) overTime() (hookin.Answer, func()) {
	d := deny(RuleOverTime, fmt.Sprintf("the guard ran past its own %s limit, so it blocks the call rather than let "+
		"the hook's timeout pass it through", r.o.Budget),
		"try the call again; if it is blocked again, tell the person (a slow disk, or a stdin that never ends)")
	if r.o.Second {
		d = secondFault(d)
	}
	r.mu.Lock()
	known := r.at != nil
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
	p, err := hookin.ReadJudged(r.o.Stdin, r.o.StdinTerminal, judgedTool)
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
