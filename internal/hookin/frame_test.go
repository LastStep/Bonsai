package hookin

// The frame of a hook that blocks, alone: the guard's own tests (internal/guard) run it with the guard's work, and
// TestOverTimeBlocks there with its record. No case races its budget: a case that answers by itself runs on a budget
// of a minute, and the over-time cases' work never answers. Each answer's time is held only to a ceiling a hang alone
// crosses: the budget, the wait and the guard's 5 s.

import (
	"bytes"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

const (
	untimed = time.Minute     // a budget no case that answers by itself comes near
	hang    = 5 * time.Second // the ceiling's margin: the guard's own budget
)

var (
	allowed = func() Answer { return Answer{} }
	blocked = func() Answer { return Answer{Block: true, Why: "the path is protected", Next: "tell the person"} }
	never   = func() Answer { select {} }
)

func panicAnswer(p any) Answer {
	return Answer{Block: true, Why: "the hook failed inside (" + strings.TrimSpace(ascii(toText(p))) + ")", Next: "tell the person"}
}

func toText(p any) string {
	if s, ok := p.(string); ok {
		return s
	}
	return "?"
}

func overTimeAnswer() (Answer, func()) {
	return Answer{Block: true, Why: "the hook ran past its own limit", Next: "try the call again"}, nil
}

func TestFrameAnswers(t *testing.T) {
	cases := []struct {
		name   string
		work   func() Answer
		code   int
		stderr string
	}{
		{"allowed", allowed, 0, ""},
		{"blocked", blocked, 2, "bonsai test-hook: the path is protected.\nnext: tell the person.\n"},
		{"a panic", func() Answer { panic("boom") }, 2, "bonsai test-hook: the hook failed inside (boom).\nnext: tell the person.\n"},
	}
	for _, c := range cases {
		var stderr bytes.Buffer
		f := Frame{Name: "test-hook", Budget: untimed, Stderr: &stderr}
		if code := f.Run(Job{Work: c.work, Panic: panicAnswer, OverTime: overTimeAnswer}); code != c.code || stderr.String() != c.stderr {
			t.Errorf("%s: exit %d, stderr %q; want %d, %q", c.name, code, stderr.String(), c.code, c.stderr)
		}
	}
	// A block with nowhere to print it still blocks: the code is the answer.
	if code := (Frame{Name: "x", Budget: untimed}).Run(Job{Work: blocked, Panic: panicAnswer, OverTime: overTimeAnswer}); code != 2 {
		t.Errorf("a block with no stderr: exit %d", code)
	}
}

// The budget runs out with the work never answering: the over-time answer counts, with no record to wait for, with
// a record awaited for Wait, and with a record held past Wait (the answer comes back, the record not yet written).
func TestFrameOverTime(t *testing.T) {
	const budget = 50 * time.Millisecond
	run := func(wait time.Duration, write func()) (int, string, time.Duration) {
		t.Helper()
		var stderr bytes.Buffer
		f := Frame{Name: "test-hook", Budget: budget, Wait: wait, Stderr: &stderr}
		start := time.Now()
		code := f.Run(Job{Work: never, Panic: panicAnswer, OverTime: func() (Answer, func()) {
			a, _ := overTimeAnswer()
			return a, write
		}})
		took := time.Since(start)
		if ceiling := budget + wait + hang; took > ceiling {
			t.Fatalf("answered after %v, past %v", took, ceiling)
		}
		return code, stderr.String(), took
	}
	want := "bonsai test-hook: the hook ran past its own limit.\nnext: try the call again.\n"
	if code, stderr, _ := run(0, nil); code != 2 || stderr != want {
		t.Errorf("no record: exit %d, %q", code, stderr)
	}
	// A record awaited: with a wait far past any load, it is written when Run returns.
	var wrote atomic.Bool
	if code, _, _ := run(untimed, func() { time.Sleep(10 * time.Millisecond); wrote.Store(true) }); code != 2 || !wrote.Load() {
		t.Errorf("an awaited record: exit %d, written %v", code, wrote.Load())
	}
	// A record held: the answer comes back within the wait, the record unwritten until it is let go.
	release, done := make(chan struct{}), make(chan struct{})
	if code, stderr, _ := run(20*time.Millisecond, func() { <-release; close(done) }); code != 2 || stderr != want {
		t.Errorf("a held record: exit %d, %q", code, stderr)
	}
	select {
	case <-done:
		t.Error("the held record was written before it was let go")
	default:
	}
	close(release)
	<-done
	// A writer that fails inside leaves the refusal unrecorded, never unanswered.
	if code, stderr, _ := run(untimed, func() { panic("the disk") }); code != 2 || stderr != want {
		t.Errorf("a failing writer: exit %d, %q", code, stderr)
	}
}

// Wait 0 means RecordWait: a held record is waited for that long, then the answer comes back.
func TestFrameDefaultWait(t *testing.T) {
	release := make(chan struct{})
	defer close(release)
	f := Frame{Name: "test-hook", Budget: 10 * time.Millisecond}
	start := time.Now()
	code := f.Run(Job{Work: never, Panic: panicAnswer, OverTime: func() (Answer, func()) {
		a, _ := overTimeAnswer()
		return a, func() { <-release }
	}})
	took := time.Since(start)
	if code != 2 || took < RecordWait || took > 10*time.Millisecond+RecordWait+hang {
		t.Errorf("exit %d after %v; want 2 after RecordWait (%v)", code, took, RecordWait)
	}
}

// A block with a rule word prints it on a line of its own after the next step; one without prints none.
func TestFrameRuleWord(t *testing.T) {
	for _, c := range []struct{ rule, want string }{
		{"bad-path", "bonsai test-hook: no.\nnext: ask.\nrule: bad-path.\n"},
		{"", "bonsai test-hook: no.\nnext: ask.\n"},
	} {
		var stderr bytes.Buffer
		f := Frame{Name: "test-hook", Budget: time.Minute, Stderr: &stderr}
		code := f.Run(Job{Work: func() Answer { return Answer{Block: true, Why: "no", Next: "ask", Rule: c.rule} }})
		if code != 2 || stderr.String() != c.want {
			t.Errorf("rule %q: exit %d, %q; want 2, %q", c.rule, code, stderr.String(), c.want)
		}
	}
}
