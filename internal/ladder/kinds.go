package ladder

// The one file through which the runner calls each rung kind and each step after a rung (plan-5 5.4, "The order"): a
// piece built beside the runner in a package of its own wires itself in here with one last commit, as 5.3.3 wired into
// the guard. The runner itself (rungs.go) calls only runKind, afterRung and afterResult, and reads only what they give.
//
//   - guardKind: rung 0 (step 5.4.3). Its wiring replaces the body with rung0.Judge(rung0.Input{Root: c.Root, Main:
//     c.Main, Home: c.Home, Active: the climbed task (*c.Task, the zero ActiveTask when c.Task is nil: no grants), Base:
//     c.Base, MergeBase: c.MergeBase}), copying Status, Reason and Captures into the Outcome, and its Refused[].Needs
//     into Next (and NextWho person for a person's grant), so a red climb's next names each path's fix.
//   - verGitKind: git integrity (step 5.4.6). Its wiring replaces the body with vergit.Judge(vergit.Input{Root: c.Root,
//     BaseBranch: c.Base, MergeBase: c.MergeBase, Tests: the rung's own tests form, else the other rungs' forms in rung
//     order (c.Config.Full.Ladder), Earlier: c.PriorCaptures(r.Rung), Timeout: rungTimeout(r)}), copying Status, Reason, Captures
//     and Hides: the runner already turns Hides into the result's git.dirty.
//   - treeDirty: whether git's status shows the tree dirty, read at the climb's start and its end. Today git status
//     (git.go); 5.4.6's wiring replaces it with vergit.Dirty (its error counting dirty), so the result's git.dirty is
//     that or a ver-git rung's Hides.
//   - afterRung: the steps after each rung that ran, in order. Today it writes the ratchet's count on the rung
//     (note 5: was, now and whether now is at least was) and judges nothing. Step 5.4.4 judges it there (a count below
//     its floor red, a count that cannot be read red: s.Count.Now is "" then), and 5.4.5 hangs the new-tests check on
//     5.4.4's rise step.
//   - afterResult: once the result and its fingerprint are written, the Bless step (5.4.4: file one ask, resolve those
//     whose offer is gone), giving the ask records it wrote for the climb's bless; [] until then.
//
// Until their wiring lands, guard and ver-git rungs end in error, "not built yet", so a floor holding one climbs red
// (the runner lands before them, "The order" item 3). A kind's Reason goes into the result only when the rung is not
// green (bonsai.ladder/1: reason is null when green); a green rung's reason (what was found, what was not judged) is
// printed under its line in the human output, and its findings are in its captures.

import (
	"encoding/json"
	"strconv"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/testout"
)

// The rung kinds bonsai.yaml names (bonsai.workspace/1's closed list).
const (
	KindGuard   = "guard"
	KindCommand = "command"
	KindVerGit  = "ver-git"
)

// Outcome is a guard or ver-git rung's outcome, which the runner copies into the rung's entry in the result.
type Outcome struct {
	Status   string        // a format.RungStatuses word: green, red or error
	Reason   string        // why it is not green, "" when it is (the runner redacts it)
	Captures schema.Object // what it found, by name, ordered and byte-stable; nil for none
	Hides    bool          // ver-git: a finding hides a change from git, so the result's git.dirty is true
	Next     string        // the fix a red climb's next step names for this rung ("" for the runner's own words)
	NextWho  string        // who takes that step: "" or agent, or person (a person-only path in command mode, 5.4.3)
}

// runKind runs a rung that is not a command: rung 0 or git integrity.
func runKind(c *Climb, r format.Rung) Outcome {
	switch r.Kind {
	case KindGuard:
		return guardKind(c, r)
	case KindVerGit:
		return verGitKind(c, r)
	}
	return Outcome{Status: "error", Reason: "the kind " + strconv.Quote(r.Kind) + " is not one this Bonsai runs (guard, command, ver-git)"}
}

// guardKind is rung 0 (step 5.4.3 wires it in).
func guardKind(c *Climb, r format.Rung) Outcome {
	return notBuilt(r, "the guard's judgment of the task's changes comes with step 5.4.3")
}

// verGitKind is git integrity (step 5.4.6 wires it in).
func verGitKind(c *Climb, r format.Rung) Outcome {
	return notBuilt(r, "git integrity comes with step 5.4.6")
}

// treeDirty reports whether the climbed checkout's tree is dirty by git's status (this file's comment).
func treeDirty(c *Climb) bool { return dirtyNow(c.Root) }

// notBuilt is a kind this build does not run yet.
func notBuilt(r format.Rung, when string) Outcome {
	return Outcome{Status: "error", Reason: "not built yet: this Bonsai does not run " + r.Kind + " rungs (" + when + ")",
		Next: "a " + r.Kind + " rung needs a Bonsai that runs it (" + when + "): use that build, or climb without the rung"}
}

// Step is one rung that ran, as the steps after a rung see it.
type Step struct {
	Rung   format.Rung        // the rung as bonsai.yaml defines it
	Entry  *format.LadderRung // its entry in the result: a step may change its status, reason and ratchet
	Tests  *testout.Report    // its tests as read (a command rung with a tests form Bonsai knows), or nil
	Count  *Count             // its ratchet and the count read for it, or nil when the rung has no ratchet
	Output []byte             // a command rung's output, both streams (up to rungjob.OutputCap); nil for other kinds
	next   string             // the kind's own fix for a red climb's next step
	note   string             // a green kind's reason: printed under its line, never in the result
	who    string             // who takes it ("" for agent)
	timed  bool               // the rung's limit ended it
}

// Count is a rung's ratchet and the count this climb read for it (5.4.0 note 2): the rung's capture of the ratchet's
// name, else, on a rung whose tests were read, the tests the output names as passed (under --ci with the accepted
// marked skips added, as today).
type Count struct {
	Name  string      // the ratchet's name
	Floor json.Number // the floor, ratchets[name] in bonsai.yaml; 0 when it names none
	Now   json.Number // the count; "" when none could be read
	From  string      // "capture" or "tests"; "" when none could be read
}

// afterRung runs the steps after a rung that ran, in order (this file's comment).
func afterRung(c *Climb, s *Step) {
	writeRatchet(s)
}

// writeRatchet writes the ratchet's count on the rung's entry (note 5): its name, its floor (was), the count (now) and
// whether now is at least was; new_tests stays null (5.4.5). It judges nothing: with no count the entry's ratchet stays
// null, and a count below its floor leaves the rung's status as it is (5.4.4 judges both).
func writeRatchet(s *Step) {
	if s.Count == nil || s.Count.Now == "" {
		return
	}
	s.Entry.Ratchet = &format.LadderRatchet{Name: s.Count.Name, Was: s.Count.Floor, Now: s.Count.Now,
		OK: number(s.Count.Now) >= number(s.Count.Floor)}
}

// Written is the result as written: its project-relative path, the SHA-256 of its bytes and the result itself.
type Written struct {
	Rel    string
	SHA256 string
	Result *format.Ladder
}

// afterResult runs once the result and its record are written, and gives the ask records it wrote (the climb's
// bless): [] until step 5.4.4's Bless.
func afterResult(c *Climb, w *Written) []schema.Object {
	return []schema.Object{}
}

// number reads a count, 0 when it does not read.
func number(n json.Number) float64 {
	f, err := strconv.ParseFloat(string(n), 64)
	if err != nil {
		return 0
	}
	return f
}
