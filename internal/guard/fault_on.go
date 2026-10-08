//go:build bonsai_test_fault

package guard

// The test fault switch (plan part 5; spec §14 and §17 step 6): the variable BONSAI_TEST_FAULT, read only by a build
// made with the tag bonsai_test_fault (go build -tags bonsai_test_fault). A normal build compiles fault_off.go in its
// place and has no code for it. Every fault only blocks: none can let through a call the normal guard refuses.
//
//	crash         the guard writes its record (rule fault-crash), says on stderr that it dies now, and dies at once,
//	              killed by the system (Linux, macOS) or terminated with an access-violation code (Windows): it ends
//	              with no decision and an exit code other than 0 and 2, and the hook line's `|| exit 2` blocks the call.
//	slow          the guard stops before it decides and sleeps past its own budget and past the hook's 10 s timeout;
//	              its timer (hook.go, the code a normal build runs too) answers first and blocks (rule over-time).
//	missing       the launcher (claude-here) starts the session with no bonsai anywhere on its PATH, so the hook
//	minimal-path  line's shell finds none and `|| exit 2` blocks; for minimal-path the PATH is a bare one (the
//	              system's folders and git's). If a bonsai built with this tag answers anyway, the launcher did not
//	              hide it: it blocks, naming itself (rule fault-missing or fault-minimal-path).
//	other values  blocked, naming the four (rule fault-unknown).
//
// The variable is read once bonsai.yaml and the payload are read, so a crash's or a slow run's record has its
// context, and a project with no bonsai.yaml stays allowed with no record, as in a normal build.

import (
	"fmt"
	"time"
)

// FaultBuild reports whether this build holds the test fault switch.
const FaultBuild = true

// FaultEnv names the test fault switch.
const FaultEnv = "BONSAI_TEST_FAULT"

// faultNext is the next step every fault's refusal names.
const faultNext = "this is a test fault: unset " + FaultEnv + " (or close the window that set it) to end it"

// slowFor is how long the slow fault sleeps: past the budget and past the hook's 10 s timeout (a variable for tests).
var slowFor = 60 * time.Second

// die ends the process the hard way (fault_crash_unix.go, fault_crash_windows.go); a test stands in its own.
var die = crashHard

// testFault acts out the fault BONSAI_TEST_FAULT names. nil means none: the guard decides as a normal build does.
func testFault(r *run) *Decision {
	v := r.o.Getenv(FaultEnv)
	var d Decision
	switch v {
	case "":
		return nil
	case "crash":
		d = deny("fault-crash", "test fault crash ("+FaultEnv+"=crash): the guard died before it decided, and the "+
			"hook line's `|| exit 2` blocked the call", faultNext)
		_ = r.record(d)
		_, _ = fmt.Fprintf(r.o.Stderr, "bonsai guard: test fault crash (%s=crash): the guard dies now, before it "+
			"decides; the hook line's `|| exit 2` blocks the call.\nnext: %s.\n", FaultEnv, faultNext)
		die()
		select {} // a real die never returns; a test's waits here for the timer
	case "slow":
		time.Sleep(slowFor)
		// Reached only when slowFor is shorter than the budget (a test): a real run's timer has answered long before.
		d = deny("fault-slow", "test fault slow ("+FaultEnv+"=slow): the guard was slow on purpose", faultNext)
	case "missing", "minimal-path":
		exe, _ := self(false)
		d = deny("fault-"+v, fmt.Sprintf("test fault %s (%s=%s): this session should find no bonsai on its PATH, yet "+
			"%s answered, so the launcher did not hide it; the call is blocked", v, FaultEnv, v, quote(exe, 200)),
			"start the session through claude-here, which hides every bonsai for this fault")
	default:
		d = deny("fault-unknown", fmt.Sprintf("%s=%s is not a test fault this build knows (missing, crash, slow, "+
			"minimal-path), so the call is blocked", FaultEnv, quote(v, 40)),
			"set "+FaultEnv+" to missing, crash, slow or minimal-path, or unset it")
	}
	return &d
}
