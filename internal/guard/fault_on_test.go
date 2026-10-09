//go:build bonsai_test_fault

package guard

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/schema"
)

// untimed is the budget of a run that ends on the guard's own decision: one no run comes near, so the outcome never
// rests on the guard's timer. Nothing in such a run can hang (its payload is a strings.Reader, and a busy log file is
// retried for openWait at most), so it needs no timer as a backstop.
const untimed = time.Hour

// crashStandIn is the panic the test's die ends the work with, in place of the crash fault's hard death.
const crashStandIn = "the test's stand-in for the crash fault's hard death"

// Each fault, in process: every one blocks the free.txt edit the normal guard allows. The real process ends (a
// crash's hard death, the slow fault's sleep past the real budget, the hook line's shell with no bonsai on its PATH)
// are cmd/bonsai's TestFaultsThroughTheHookLine, on the built binaries.
//
// No case races a short budget. Every case ran on a fixed 200 ms budget, and on a loaded machine the work could
// reach its fault after that, so the case got an over-time record instead of its own. Now every fault but slow ends
// the work by itself (it decides, or the crash's stand-in ends it), so those cases run untimed and end on the
// guard's own decision. The slow fault's block is the guard's timer, so that case alone has a budget, and it counts
// only a run whose timer fired after the work reached the fault (slowRun, below).
func TestEachFaultBlocks(t *testing.T) {
	if !FaultBuild {
		t.Fatal("fault_on.go says this is not a fault build")
	}
	// A crash cannot end the test binary, so the test stands in its own die. The old stand-in held the work in die
	// until the budget ran out, which put the crash case's outcome on the timer. This one ends the work with a panic,
	// which the guard's own recover turns into a block (internal-error, as TestBlocksWhatItCannotReadOrDecide's "a
	// panic inside" shows), with no timer. What the case proves of the crash fault itself is what it does before it
	// dies: its record, its message, its call to die, and no decision of its own after that. No work is left in die,
	// so die is restored.
	var died atomic.Int32
	defer func(d func()) { die = d }(die)
	die = func() { died.Add(1); panic(crashStandIn) }
	// Set once for the rest of this test binary, never restored or set again: a slow run leaves its work goroutine in
	// the sleep or on its way there, and a write would race with its read. slowFor never ends while the tests run, so
	// no late record lands in a removed folder.
	if slowFor != time.Hour {
		slowFor = time.Hour
	}

	cases := []struct {
		fault    string
		inStderr []string
		rule     string
	}{
		{"crash", []string{"test fault crash (BONSAI_TEST_FAULT=crash): the guard dies now",
			"the guard failed inside (" + crashStandIn + ")"}, "fault-crash"},
		{"missing", []string{"test fault missing (BONSAI_TEST_FAULT=missing): this session should find no bonsai on its PATH"}, "fault-missing"},
		{"minimal-path", []string{"test fault minimal-path (BONSAI_TEST_FAULT=minimal-path)"}, "fault-minimal-path"},
		{"sideways", []string{`BONSAI_TEST_FAULT="sideways" is not a test fault this build knows`}, "fault-unknown"},
	}
	for _, c := range cases {
		t.Run(c.fault, func(t *testing.T) {
			dir := project(t, testYAML)
			code, stderr := guardRun(t, env(dir, FaultEnv, c.fault), strings.NewReader(editOf("s-f", dir, "free.txt")), untimed)
			if code != 2 || !strings.Contains(stderr, "\nnext: ") || strings.Contains(stderr, "ran past its own") {
				t.Fatalf("exit %d, stderr %q; want 2, a next step and no over-time", code, stderr)
			}
			for _, s := range c.inStderr {
				if !strings.Contains(stderr, s) {
					t.Errorf("stderr %q does not say %q", stderr, s)
				}
			}
			asciiOnly(t, stderr)
			recs := records(t, dir, "s-s-f.ndjson")
			if len(recs) != 1 {
				t.Fatalf("%d records, want one (%s)", len(recs), c.rule)
			}
			checkFaultRecord(t, recs[0], c.rule)
		})
	}
	if n := died.Load(); n != 1 {
		t.Errorf("the crash fault called die %d times, want once", n)
	}

	// The slow fault sleeps (an hour here) and the guard's timer answers it, so this case cannot do without a
	// budget. It starts at 200 ms; a run whose timer fired before the work reached the fault (a loaded machine) is
	// run again, on a new project, with twice the budget. Past the guard's real Budget it stops: a guard whose work
	// cannot reach the fault in that time fails a real run too.
	t.Run("slow", func(t *testing.T) {
		for budget := 200 * time.Millisecond; ; budget *= 2 {
			early := slowRun(t, budget)
			if early == "" {
				return
			}
			if budget > Budget {
				t.Fatalf("budget %v, past the guard's real %v: the timer still fired before the work reached the fault (%s)",
					budget, Budget, early)
			}
			t.Logf("budget %v: the timer fired before the work reached the fault (%s); again with %v", budget, early, 2*budget)
		}
	})
}

// slowRun runs the slow fault once with the given budget. The answer (exit 2, over-time) is checked on every run,
// since the slow work never answers. The record is checked when the run counts: when the timer's record began after
// the work had reached the fault. The work reads BONSAI_TEST_FAULT at the fault, and a record reads the clock as it
// begins (the slow work records nothing, so the clock is the timer's alone). A run that does not count gives why.
//
// The answer waits recordWait at most for its record (hook.go), and a loaded machine can take longer to hash the
// binary and write it: the record then lands after the answer, as the guard allows ("recorded if that can be done in
// recordWait"). The record is awaited here, not raced; that the answer waits for it is TestOverTimeBlocks's record
// half, whose wait is guard code (step 5.3).
func slowRun(t *testing.T, budget time.Duration) (early string) {
	t.Helper()
	dir := project(t, testYAML)
	base := env(dir, FaultEnv, "slow")
	var atFault atomic.Bool
	getenv := func(k string) string {
		if k == FaultEnv {
			atFault.Store(true)
		}
		return base(k)
	}
	const (
		none        = iota // no record began by the answer: the timer fired before the work had read bonsai.yaml
		afterFault         // the record began with the work at the fault
		beforeFault        // the record began before the work reached the fault
	)
	var began atomic.Int32
	now := func() time.Time {
		at := int32(beforeFault)
		if atFault.Load() {
			at = afterFault
		}
		began.CompareAndSwap(none, at)
		return time.Now()
	}
	var stderr bytes.Buffer
	code := Main(Options{Stdin: strings.NewReader(editOf("s-f", dir, "free.txt")), Stderr: &stderr, Getenv: getenv,
		Budget: budget, Now: now})
	if code != 2 || !strings.Contains(stderr.String(), fmt.Sprintf("the guard ran past its own %s limit", budget)) ||
		!strings.Contains(stderr.String(), "\nnext: ") {
		t.Fatalf("budget %v: exit %d, stderr %q; want 2, over-time and a next step", budget, code, stderr.String())
	}
	asciiOnly(t, stderr.String())
	switch began.Load() {
	case none:
		return "no record began"
	case beforeFault:
		return "its record began before the work reached the fault"
	}
	file, late := awaitRecord(t, dir)
	if late {
		t.Logf("budget %v: the record landed after the answer (the answer waits %v at most)", budget, recordWait)
	}
	// The timer's record takes the payload first and reads the clock next: a work that reached the fault between the
	// two leaves a record with no session, in a day's file.
	if file != "s-s-f.ndjson" {
		return "its record took no payload"
	}
	recs := records(t, dir, file)
	if len(recs) != 1 {
		t.Fatalf("budget %v: %d records, want one (%s)", budget, len(recs), RuleOverTime)
	}
	checkFaultRecord(t, recs[0], RuleOverTime)
	return ""
}

// awaitRecord waits for the one record a slow run writes: a log file holding a whole line, as the guard writes a
// record in one write. It gives the file's name, and whether it was still being written when the wait began. The
// guard's real Budget bounds the wait only to fail a record that never lands; a record that lands ends it.
func awaitRecord(t *testing.T, dir string) (file string, late bool) {
	t.Helper()
	logDir := filepath.Join(dir, filepath.FromSlash(LogDir))
	for deadline := time.Now().Add(Budget); ; late = true {
		names, err := filepath.Glob(filepath.Join(logDir, "*.ndjson"))
		if err != nil {
			t.Fatal(err)
		}
		for _, name := range names {
			if b, err := os.ReadFile(name); err == nil && bytes.HasSuffix(b, []byte("\n")) {
				return filepath.Base(name), late
			}
		}
		if time.Now().After(deadline) {
			for i := range names {
				names[i] = filepath.Base(names[i])
			}
			t.Fatalf("no whole record in %s %v after the answer (files: %v)", LogDir, Budget, names)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// checkFaultRecord checks a fault case's one record: its rule, a refusal, no target, and the binary's hash (it is
// the session's first).
func checkFaultRecord(t *testing.T, r schema.Object, rule string) {
	t.Helper()
	if r.String("rule") != rule || r.String("decision") != "deny" || r.String("target") != "" {
		t.Errorf("record: rule %s decision %s target %q; want %s, deny, none", r.String("rule"), r.String("decision"),
			r.String("target"), rule)
	}
	if r.String("bonsai_sha256") == "" {
		t.Errorf("the session's first record names no hash")
	}
}

// With the switch unset, a fault build decides as a normal one.
func TestFaultBuildWithNoFaultIsNormal(t *testing.T) {
	dir := project(t, testYAML)
	if code, stderr := guardRun(t, env(dir), strings.NewReader(editOf("s-g", dir, "free.txt")), 0); code != 0 || stderr != "" {
		t.Fatalf("free.txt: exit %d, %q", code, stderr)
	}
	if code, _ := guardRun(t, env(dir), strings.NewReader(editOf("s-g", dir, "protected.txt")), 0); code != 2 {
		t.Fatalf("protected.txt: exit %d", code)
	}
}
