package hookin

// The frame of a hook that blocks (design/plan-5.md, 5.3.0 note 4): `bonsai hook guard` now, `bonsai hook stop` from
// step 5.3.4, each with its own budget. Spec §7: "The guard keeps its own timer under the hook's timeout, since a hook
// that times out does not block" (Claude Code lets a timed-out hook's call through), so the frame answers by its budget
// whatever the work is doing, and the caller then exits, which stops the work.
//
// The order of a run: the timer is armed first; the work runs in its own goroutine, a panic inside it caught and
// answered as the hook's internal-error refusal; whichever comes first, the work's answer or the timer, is the answer.
// Past the budget the hook's over-time refusal answers, and the writer of its record is awaited for Wait at most (the
// log then holds the refusal that counted; a record not written by then may land after the answer, or never if the
// process has exited). The answer is the exit code (spec §3): 0 allows, 2 blocks, with the reason and the next step
// on stderr, which Claude Code shows the agent, and the rule word when the hook gives one (the guard does), so an
// agent can act on the refusal by its rule without reading the log. A refusal that cannot be printed still blocks:
// the code is the answer.

import (
	"fmt"
	"io"
	"time"
)

// RecordWait is how long an over-time answer waits for its record when the frame is given no wait of its own.
const RecordWait = 500 * time.Millisecond

// Frame is the run of one call of a hook that blocks.
type Frame struct {
	Name   string        // the hook in its refusal: "guard" prints "bonsai guard: <why>.\nnext: <next>.\n" (and "rule: <rule>.\n")
	Budget time.Duration // the hook's own time limit, under its hook line's timeout; past it the over-time answer counts
	Wait   time.Duration // how long the over-time answer waits for its record; 0 means RecordWait
	Stderr io.Writer     // where a block's reason and next step go
}

// Answer is a blocking hook's answer to one call.
type Answer struct {
	Block bool   // true: exit 2, the call is blocked; false: exit 0, it goes on
	Why   string // a block's reason, said first (spec §7); ASCII, no final full stop
	Next  string // a block's next step (spec §3: every refusal names one); ASCII, no final full stop
	Rule  string // a block's rule word, as its record keeps it, printed on a line of its own after the next step; "" for none
}

// Job is the hook's part of a run.
type Job struct {
	Work     func() Answer           // the hook's whole job, in its own goroutine; it writes its own record
	Panic    func(p any) Answer      // the answer when Work panics: the hook's internal-error refusal
	OverTime func() (Answer, func()) // the answer when the budget runs out first, and the writer of its record (nil for none)
}

// Run answers one call and gives the exit code (this file's comment). It returns by the budget and the wait whatever
// Work is doing.
func (f Frame) Run(j Job) int {
	timer := time.NewTimer(f.Budget)
	defer timer.Stop()
	done := make(chan Answer, 1)
	go func() {
		var a Answer
		defer func() {
			if p := recover(); p != nil {
				a = j.Panic(p)
			}
			done <- a
		}()
		a = j.Work()
	}()
	select {
	case a := <-done:
		return f.answer(a)
	case <-timer.C:
	}
	a, write := j.OverTime()
	if write != nil {
		wrote := make(chan struct{})
		go func() {
			defer close(wrote)
			defer func() { _ = recover() }() // a writer that fails inside leaves the refusal unrecorded, never unanswered
			write()
		}()
		wait := f.Wait
		if wait <= 0 {
			wait = RecordWait
		}
		waited := time.NewTimer(wait)
		defer waited.Stop()
		select {
		case <-wrote:
		case <-waited.C:
		}
	}
	return f.answer(a)
}

// answer prints a block and gives the exit code.
func (f Frame) answer(a Answer) int {
	if !a.Block {
		return 0
	}
	if f.Stderr != nil {
		_, _ = fmt.Fprintf(f.Stderr, "bonsai %s: %s.\nnext: %s.\n", f.Name, a.Why, a.Next)
		if a.Rule != "" {
			_, _ = fmt.Fprintf(f.Stderr, "rule: %s.\n", a.Rule)
		}
	}
	return 2
}
