package ladder

// The one file through which the runner calls each rung kind and each step after a rung (plan-5 5.4, "The order"): a
// piece built beside the runner in a package of its own wires itself in here with one last commit, as 5.3.3 wired into
// the guard. The runner itself (rungs.go) calls only runKind, afterRung, afterRungs and afterResult, and reads only what
// they give.
//
//   - guardKind: rung 0 (step 5.4.3, internal/rung0): rung0.Judge(rung0.Input{Root: c.Root, Main: c.Main, Home:
//     c.Home, Active: the climbed task (*c.Task, the zero ActiveTask when c.Task is nil: no grants), Base: c.Base,
//     MergeBase: c.MergeBase, Timeout: rungTimeout(r)}), its Status, Reason and Captures copied (checked, and
//     not_judged: what a green rung's reason says was not judged, which the result would not keep). A red one's Next
//     names each refused path with its rule and the grant or revert it needs (Refused[].Needs), how a revert is made,
//     and bonsai check's step when its findings count (5.4.2 note 12); NextWho is person when a refused path is
//     person-only and this machine's settings put the workspace in command mode, where only a person's grant opens
//     one. An error (git cannot say what changed, or the limit passed) names git status as the next step.
//   - verGitKind: git integrity (step 5.4.6, internal/vergit): vergit.Judge(vergit.Input{Root: c.Root, BaseBranch:
//     c.Base, MergeBase: c.MergeBase, Tests: the rung's own tests form, else the forms the ladder's other rungs name in
//     rung order (c.Config.Full.Ladder), each once, Earlier: c.PriorCaptures(r.Rung), Timeout: rungTimeout(r)}), its
//     Status, Reason, Captures (base, stash-count, exclude-hash and each finding's count; text redacted) and Hides
//     copied: the runner turns Hides into the result's git.dirty. Its findings never turn it red; error (git cannot be
//     read) names git status as the next step.
//   - treeDirty: whether git's status shows the tree dirty, read at the climb's start and its end: vergit.Dirty, git
//     status as git integrity runs it (a fixed environment and settings pinned, so the repository's own configuration
//     cannot talk it out of a change), an error counting dirty. The result's git.dirty is that, HEAD moved, or a
//     ver-git rung's Hides.
//   - rungTimeout: each rung's limit, used for its run and counted in the lock's budget (lock.go): its timeout_s, else
//     its kind's own default: a command rung the rung job's (rungjob.DefaultTimeout, 600 s), rung 0 rung0.DefaultTimeout
//     and git integrity vergit.DefaultTimeout (120 s each: git reads, which a limit only ends when git does not answer),
//     so a floor of rung 0 and ver-git does not make a waiting climb wait as if each ran tests for ten minutes.
//   - afterRung: the steps after each rung that ran: the ratchet's judgment (internal/ratchet's Judge, step 5.4.4): the
//     count written on the rung (was, now and whether now is at least was), and the rung red when the count is below its
//     floor or none could be read (s.Count.Now is "" then, or the rung's kind counts nothing), or when its floor in
//     ratchets is not a number (never read as 0, so a typo cannot lower a floor), the reason added to the rung's own.
//   - afterRungs: the steps after the last rung, once every rung's line is printed: the rise step. On a local climb
//     that is green and clean (clean at its start and now, HEAD where it started, nothing hidden from git), each ratchet
//     above its floor gets its new_tests from ratchet.NewTests, internal/newtests' check (step 5.4.5): the rung run on
//     the base in a temporary worktree, with the rung's limit, the climb's environment and its stop. Its lines print
//     under the rungs' (a line before it starts, since it runs the rung twice more), and its one-line summary goes to
//     the Bless step for the ask's why. A check stopped or past its limit says "check not run" and never changes a
//     colour. On any other climb new_tests stays null.
//   - afterResult: once the result and its fingerprint are written, the Bless step (ratchet.Bless, step 5.4.4): on the
//     base branch, every open Bless ask whose offer is gone resolved, and the climb's own offer filed, its why giving
//     each rise's new-tests line from the rise step (Result.Checks); it gives the ask records written, the climb's
//     bless, and its lines go under the rungs'.
//
// A kind's Reason goes into the result only when the rung is not green (bonsai.ladder/1: reason is null when green); a
// green rung's reason (what was found, what was not judged) is printed under its line in the human output, and its
// findings are in its captures.

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/LastStep/Bonsai/internal/asks"
	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/guard"
	"github.com/LastStep/Bonsai/internal/ratchet"
	"github.com/LastStep/Bonsai/internal/redact"
	"github.com/LastStep/Bonsai/internal/rung0"
	"github.com/LastStep/Bonsai/internal/rungjob"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/testout"
	"github.com/LastStep/Bonsai/internal/vergit"
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

// guardKind is rung 0 (this file's comment).
func guardKind(c *Climb, r format.Rung) Outcome {
	var active workspace.ActiveTask // none named: no grants
	if c.Task != nil {
		active = *c.Task
	}
	o := rung0.Judge(rung0.Input{Root: c.Root, Main: c.Main, Home: c.Home, Active: active, Base: c.Base, MergeBase: c.MergeBase,
		Timeout: rungTimeout(r)})
	out := Outcome{Status: o.Status, Reason: o.Reason, Captures: redactCaptures(o.Captures)}
	switch o.Status {
	case statusRed:
		out.Next, out.NextWho = rung0Next(c, o)
	case statusError:
		out.Next = "rung 0 could not judge this checkout's changes (the rung's reason says why): run: git status, and fix " +
			"what it reports (a rung 0 past its limit needs a longer timeout_s in bonsai.yaml, a person's change)"
	}
	return out
}

// rung0Next is a red rung 0's fix, as a red climb's next names it: each refused path with its rule and what clears it,
// how a revert is made, and bonsai check's step; who is person when a person-only path is refused in command mode.
func rung0Next(c *Climb, o rung0.Outcome) (string, string) {
	var parts []string
	person := false
	for _, f := range o.Refused {
		parts = append(parts, strconv.QuoteToASCII(f.Path)+" ("+f.Rule+") needs "+f.Needs)
		person = person || f.Rule == guard.RulePersonOnly
	}
	if o.More > 0 {
		parts = append(parts, "and "+strconv.Itoa(o.More)+" more paths past these, which the climb after them names")
	}
	if len(o.Refused) > 0 && o.Revert != "" {
		parts = append(parts, "a revert: "+o.Revert)
	}
	if o.Check != "" {
		parts = append(parts, o.Check)
	}
	if len(parts) == 0 {
		return "", ""
	}
	who := ""
	if person && commandMode(c) {
		who = "person"
	}
	return strings.Join(parts, "; "), who
}

// commandMode reports whether this machine's settings for the workspace say command mode (status_writes: command),
// read from the verified main checkout's machine folder as the guard reads them; settings that cannot be read say
// agents here (the guard refuses the edit itself then).
func commandMode(c *Climb) bool {
	if c.Home == "" {
		return false
	}
	ms, err := workspace.LoadMachineSettings(c.Home, c.Main)
	return err == nil && ms.StatusWrites == workspace.StatusWritesCommand
}

// redactCaptures passes a kind's text captures through the redactor (note 6: every free-text value of a result).
func redactCaptures(caps schema.Object) schema.Object {
	if caps == nil {
		return nil
	}
	out := make(schema.Object, 0, len(caps))
	for _, m := range caps {
		if s, ok := m.Value.(string); ok {
			m.Value = redact.Text(s)
		}
		out = append(out, m)
	}
	return out
}

// verGitKind is git integrity (this file's comment).
func verGitKind(c *Climb, r format.Rung) Outcome {
	o := vergit.Judge(vergit.Input{Root: c.Root, BaseBranch: c.Base, MergeBase: c.MergeBase, Tests: testForms(c, r),
		Earlier: c.PriorCaptures(r.Rung), Timeout: rungTimeout(r)})
	out := Outcome{Status: o.Status, Reason: o.Reason, Captures: redactCaptures(o.Captures), Hides: o.Hides}
	if o.Status != statusGreen {
		out.Next = "git cannot be read in this checkout (the rung's reason says why): run: git status, and fix what it reports"
	}
	return out
}

// testForms are the tests forms whose test-file conventions git integrity reads: the rung's own tests form when it
// names one, else every form the ladder's other rungs name, in rung order, each once.
func testForms(c *Climb, r format.Rung) []string {
	if r.Tests != nil && *r.Tests != "" {
		return []string{*r.Tests}
	}
	rungs := append([]format.Rung{}, c.Config.Full.Ladder...)
	sortRungs(rungs)
	var forms []string
	for _, o := range rungs {
		if o.Tests != nil && *o.Tests != "" && !contains(forms, *o.Tests) {
			forms = append(forms, *o.Tests)
		}
	}
	return forms
}

// contains reports whether a list holds a word.
func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

// rungTimeout is a rung's limit (this file's comment): its timeout_s, else its kind's default.
func rungTimeout(r format.Rung) time.Duration {
	if r.TimeoutS != nil && *r.TimeoutS > 0 {
		return time.Duration(*r.TimeoutS) * time.Second
	}
	switch r.Kind {
	case KindGuard:
		return rung0.DefaultTimeout
	case KindVerGit:
		return vergit.DefaultTimeout
	}
	return rungjob.DefaultTimeout
}

// treeDirty reports whether the climbed checkout's tree is dirty by git's status (this file's comment).
func treeDirty(c *Climb) bool {
	dirty, err := vergit.Dirty(vergit.Input{Root: c.Root})
	return dirty || err != nil
}

// Step is one rung that ran, as the steps after a rung see it.
type Step struct {
	Rung   format.Rung        // the rung as bonsai.yaml defines it
	Entry  *format.LadderRung // its entry in the result: a step may change its status, reason and ratchet
	Tests  *testout.Report    // its tests as read (a command rung with a tests form Bonsai knows), or nil
	Count  *Count             // its ratchet and the count read for it, or nil when the rung has no ratchet
	Output []byte             // a command rung's output, both streams (up to rungjob.OutputCap); nil for other kinds
	rise   *ratchet.Checked   // the rise step's check of its ratchet's new tests, nil when none ran
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

// afterRung runs the steps after a rung that ran (this file's comment).
func afterRung(c *Climb, s *Step) {
	judgeRatchet(c, s)
}

// afterRungs runs the steps after the last rung, once every rung's line is printed (this file's comment).
func afterRungs(c *Climb) {
	riseStep(c)
}

// judgeRatchet judges the rung's ratchet (ratchet.Judge): its count written on the entry, and when the count is below
// its floor or none was read, or its floor is not a number (badFloor), the rung red and the reason added to its own. The ratchet's fix is the climb's next step
// when it is the rung's only fault; a rung already red or in error keeps its status and its own fix.
func judgeRatchet(c *Climb, s *Step) {
	if s.Rung.Ratchet == nil {
		return
	}
	in := ratchet.Reading{Rung: s.Rung, Captures: s.Entry.Captures, Tests: s.Entry.Tests, CI: c.CI}
	in.Name, _ = s.Rung.Ratchet.(string)
	floor, bad := floorOf(c.Config.Full.Ratchets, in.Name)
	in.Floor = floor
	if s.Count != nil {
		in.Floor, in.Now, in.From = s.Count.Floor, s.Count.Now, s.Count.From
	}
	v := ratchet.Judge(in)
	if bad != "" {
		v = badFloor(in.Name, bad)
	}
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

// badFloor is the judgment of a ratchet whose floor in bonsai.yaml's ratchets is not a number (floorOf): red, never
// judged against 0, and the fix a person's, in bonsai.yaml.
func badFloor(name, written string) ratchet.Verdict {
	shown := name
	if !ratchet.IsName(name) {
		shown = strconv.QuoteToASCII(name)
	}
	return ratchet.Verdict{Red: true, Who: "person",
		Reason: "the ratchet " + shown + " cannot be judged: its floor in bonsai.yaml's ratchets is " + written + ", which is not a " +
			"number, and a floor is never read as 0 (a typo would lower it): the floor in bonsai.yaml's ratchets must be a number, " +
			"the count the rung may not fall below",
		Next: "a person writes ratchets." + name + " in bonsai.yaml as a plain number (the count it may not fall below, such as 812); " +
			"then run: bonsai check --json"}
}

// newTests is the rise step's call (ratchet.NewTests), a variable for the tests.
var newTests = ratchet.NewTests

// riseStep is the rise step, after the last rung (internal/ratchet's rise.go): on a local climb whose every rung is
// green, each ratchet above its floor gets its new_tests from ratchet.NewTests, once the tree is seen clean (clean at
// the start and now, HEAD where it started, nothing hidden from git), its lines printed. On any other climb new_tests
// stays null.
func riseStep(c *Climb) {
	if c.CI || len(c.steps) == 0 {
		return
	}
	var risen []*Step
	for _, s := range c.steps {
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
		what := strings.TrimSpace(fmt.Sprintf("rung %d %s", s.Entry.Rung, s.Entry.Name)) + fmt.Sprintf(" (%s %s->%s)", rt.Name, rt.Was, rt.Now)
		c.say("")
		c.say("  new tests, %s: checking them on the base (the rung runs there twice, in a temporary worktree)", what)
		got := newTests(ratchet.RiseInput{Name: rt.Name, Was: rt.Was, Now: rt.Now, Rung: s.Rung, Tests: s.Tests, Root: c.Root,
			Main: c.Main, Task: c.TaskID(), Head: c.git.SHA, Base: c.Base, MergeBase: c.MergeBase, Prior: c.Prior, Env: c.o.Env,
			Timeout: rungTimeout(s.Rung), Stop: c.o.Stop})
		s.rise = &got
		rt.NewTests = got.NewTests
		for i, line := range got.Lines {
			if i == 0 {
				c.say("  new tests, %s, %.1fs: %s", what, got.Took.Seconds(), redact.Text(line))
				continue
			}
			c.say("  %s", redact.Text(line))
		}
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
		Result: ratchet.Result{Ladder: w.Result, SHA256: w.SHA256, Elsewhere: !samePlace(c.Root, c.Main), Checks: c.checks()},
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

// checks are the rise step's one-line summaries, one per rung climbed in the result's order ("" for a rung with none),
// for the Bless ask's why; nil when the rise step checked nothing.
func (c *Climb) checks() []string {
	var out []string
	for i, s := range c.steps {
		if s.rise == nil {
			continue
		}
		if out == nil {
			out = make([]string, len(c.steps))
		}
		out[i] = s.rise.Summary
	}
	return out
}
