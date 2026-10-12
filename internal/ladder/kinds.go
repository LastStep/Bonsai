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
//   - afterRung: the steps after each rung that ran, in order. First the ratchet's judgment (internal/ratchet's Judge,
//     step 5.4.4): the count written on the rung (was, now and whether now is at least was), and the rung red when the
//     count is below its floor or none could be read (s.Count.Now is "" then, or the rung's kind counts nothing), the
//     reason added to the rung's own. Then, after the last rung, the rise step: on a local climb that is green and clean,
//     each ratchet above its floor gets its new_tests from ratchet.NewTests, which step 5.4.5's check replaces.
//   - afterResult: once the result and its fingerprint are written, the Bless step (ratchet.Bless, step 5.4.4): on the
//     base branch, every open Bless ask whose offer is gone resolved, and the climb's own offer filed; it gives the ask
//     records written, the climb's bless, and its lines go under the rungs'.
//
// Until their wiring lands, guard and ver-git rungs end in error, "not built yet", so a floor holding one climbs red
// (the runner lands before them, "The order" item 3). A kind's Reason goes into the result only when the rung is not
// green (bonsai.ladder/1: reason is null when green); a green rung's reason (what was found, what was not judged) is
// printed under its line in the human output, and its findings are in its captures.

import (
	"encoding/json"
	"strconv"

	"github.com/LastStep/Bonsai/internal/asks"
	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/ratchet"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/testout"
	"github.com/LastStep/Bonsai/internal/workspace"
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
	judgeRatchet(c, s)
	if len(c.steps)+1 == len(c.rungs) {
		riseStep(c, s)
	}
}

// judgeRatchet judges the rung's ratchet (ratchet.Judge): its count written on the entry, and when the count is below
// its floor or none was read, the rung red and the reason added to its own. The ratchet's fix is the climb's next step
// when it is the rung's only fault; a rung already red or in error keeps its status and its own fix.
func judgeRatchet(c *Climb, s *Step) {
	if s.Rung.Ratchet == nil {
		return
	}
	in := ratchet.Reading{Rung: s.Rung, Captures: s.Entry.Captures, Tests: s.Entry.Tests, CI: c.CI}
	in.Name, _ = s.Rung.Ratchet.(string)
	in.Floor = floorOf(c.Config.Full.Ratchets, in.Name)
	if s.Count != nil {
		in.Floor, in.Now, in.From = s.Count.Floor, s.Count.Now, s.Count.From
	}
	v := ratchet.Judge(in)
	s.Entry.Ratchet = v.Ratchet
	if !v.Red {
		return
	}
	if s.Entry.Status == statusGreen {
		s.Entry.Status = statusRed
		s.next, s.who = v.Next, v.Who
	}
	why := *text(v.Reason)
	if s.Entry.Reason != nil {
		why = *s.Entry.Reason + "\n" + why
	}
	s.Entry.Reason = &why
}

// newTests is the rise step's call (ratchet.NewTests), a variable for the tests.
var newTests = ratchet.NewTests

// riseStep is the rise step, after the last rung (internal/ratchet's rise.go): on a local climb whose every rung is
// green, each ratchet above its floor gets its new_tests from ratchet.NewTests, once the tree is seen clean (clean at
// the start and now, HEAD where it started, nothing hidden from git). On any other climb new_tests stays null.
func riseStep(c *Climb, last *Step) {
	if c.CI {
		return
	}
	steps := append(append([]*Step{}, c.steps...), last)
	var risen []*Step
	for _, s := range steps {
		if s.Entry.Status != statusGreen {
			return
		}
		if rt := s.Entry.Ratchet; rt != nil && ratchet.Above(rt.Now, rt.Was) {
			risen = append(risen, s)
		}
	}
	if len(risen) == 0 || c.git.Dirty || len(c.hides) > 0 || head(c.Root) != c.git.SHA || treeDirty(c) {
		return
	}
	for _, s := range risen {
		rt := s.Entry.Ratchet
		rt.NewTests = newTests(ratchet.RiseInput{Name: rt.Name, Was: rt.Was, Now: rt.Now, Rung: s.Rung, Tests: s.Tests, Root: c.Root,
			Main: c.Main, Task: c.TaskID(), Head: c.git.SHA, Base: c.Base, MergeBase: c.MergeBase, Prior: c.Prior, Env: c.o.Env,
			Timeout: rungTimeout(s.Rung), Stop: c.o.Stop})
	}
}

// Written is the result as written: its project-relative path, the SHA-256 of its bytes and the result itself.
type Written struct {
	Rel    string
	SHA256 string
	Result *format.Ladder
}

// afterResult runs once the result and its record are written (this file's comment): the Bless step, whose lines it
// prints under the rungs', and whose ask records it gives for the climb's bless. A --ci climb has none.
func afterResult(c *Climb, w *Written) []schema.Object {
	if c.CI {
		return []schema.Object{}
	}
	d := ratchet.Bless(ratchet.Step{
		Place:  asks.Place{Workspace: c.Config.ID, Local: c.local(), Session: c.Getenv(SessionEnv), Now: c.Now},
		Task:   c.TaskID(),
		Result: ratchet.Result{Ladder: w.Result, SHA256: w.SHA256, Elsewhere: !samePlace(c.Root, c.Main)},
		Where: ratchet.Where{Workspace: c.Config.ID, Base: c.Base, Branch: workspace.ReadGitLayout(c.Main).Branch(), Head: head(c.Main),
			Floors: c.Config.Full.Ratchets},
	})
	for _, line := range d.Lines {
		c.say("  %s", line)
	}
	for _, n := range d.Notes {
		c.printNote("note: " + n)
	}
	if d.Records == nil {
		return []schema.Object{}
	}
	return d.Records
}
