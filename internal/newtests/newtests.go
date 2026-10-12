// Package newtests checks that a climb's new tests fail on the base (spec §9, "New tests must fail on the base", Rohan
// 7 Oct; design/plan-5.md 5.4.5): when a rung's test count rose, it runs the rung on the code from before the change,
// once as it is and once with the branch's changed test files copied in, and tells apart the new tests that failed
// there (they prove something), those that passed there (they prove nothing new) and those whose check was not run.
//
//	r := newtests.Check(newtests.Input{Root: root, Rung: rung, Base: base, Head: report, Rise: 3, Stop: stop})
//	ratchet.NewTests = &r.NewTests   // {base, count, failed_on_base, passed_on_base, not_run}
//	lines := r.Lines()               // the human output: the counts, then the names, at most 20 a group
//	why := r.Summary()               // one line for the Bless ask's why
//
// When it runs is the caller's (note 1, the ratchet's rise step): a local run, green and clean, for a ratchet whose
// count rose; on any other run new_tests is null. Check does the check whenever it is called.
//
// The base (note 2) is the caller's: the merge base of HEAD with the base branch, from the runner; on the base branch
// itself, where that is HEAD, the base the task's last result recorded (ChooseBase, through vergit.RecordedBase). None:
// "check not run: no base", and new_tests.base is "".
//
// The two base runs (note 3), in one temporary worktree at the base: git worktree add --detach in the system's
// temporary folder (a folder bonsai-newtests-<random> holding the worktree, base, and an owner file naming this
// process), with the repository's hooks off and git run by vergit.Runner (its fixed environment, the user's own
// line-end settings kept, so the base is checked out as the user's git would). Each run goes through internal/rungjob
// with the rung's own timeout, in the rung's environment, so nothing a base run starts outlives it:
//  1. base_setup, if the rung has one (installing dependencies). Not green: every new test is "check not run".
//  2. The rung's command as it is: the base's own test names, and that they pass. The new tests are the tests the
//     HEAD run reports passed whose names this run does not report at all (passed, since the ratchet counts passed
//     tests: a test skipped at HEAD adds nothing to the count and proves nothing anywhere). A test that failed here
//     (an old test failing at the base) makes every new test "check not run". When this run cannot name the base's
//     tests (it timed out, was stopped or did not start; a package did not build or failed outside its tests, or the
//     output was not a report; it ended red with no failing test named; its output passed rungjob.OutputCap) the new
//     tests cannot be named, and count is the ratchet's rise (Input.Rise), every one "check not run".
//  3. The branch's changed test files copied in (deleted ones removed), by the form's convention
//     (format.TestFileConventions; vergit.TestFile): go-test-json's *_test.go and testdata/ files, node's default test
//     patterns for tap, and for junit the files the reports' testcase elements name. A file is changed when git's diff
//     of the base against the climbed working tree, or its untracked files, name it. None found: "check not run: the
//     test files cannot be told", and no base run is made.
//  4. The rung again. Failed on the base: a new test this run reports failed by name, while every test of the first
//     run that this run also ran passed (none failed; a test failing only because a subtest failed is judged by its
//     subtests). Proves nothing new: a new test this run reports passed. Check not run: a new test this run does not
//     report, or reports skipped or todo (its package did not build, its file did not load: a Go test calling code the
//     branch adds does not build at the base, which proves nothing about the base), and every one when an old test
//     failed, or the run timed out, was stopped or did not start.
//
// A stop (Input.Stop) ends a running base run at once, as the rung job ends a rung; git's own calls, each quick, are
// let finish (a git worktree add cut short could leave its record behind), and the stop is read after each. Then the
// worktree is removed (git worktree remove --force, the folder removed with Windows' busy retry, git worktree prune
// when that failed), always: after a red run, a timeout, a stop, or a panic during the check, with a context of its
// own so a stop cannot cut the removal short. A worktree left by a check whose process was killed (SIGKILL,
// taskkill /f) is removed by the next check in the same repository, once its owner file names a process that no longer
// runs.
//
// Counts: count = failed_on_base + passed_on_base + not_run, always. The rise is offered for Bless whatever the
// numbers (spec §9: "ratchets are Rohan's"); the caller carries them into the result and the ask.
//
// Known limits, for the verifier:
//   - A new test that reads a file the branch added outside the form's test-file convention (a fixture beside the
//     code, not under testdata/ or test/) fails at the base for want of that file, and counts as failed on the base.
//   - A test that is new by name but was renamed from an old one counts as new; an old test the branch changed to
//     fail at the base makes every new test "check not run" (spec §9: the old tests must run and pass).
//   - Names are the identity: two tests of one name (tap and junit allow it) are one new test, failed when any
//     instance failed at the base.
//   - JUnit has no parent tests: where a reporter lists a parent beside its subtests (Go's converters do), a new
//     subtest under an old test that fails at the base fails its parent too, which reads as an old test failing, so
//     every new test is "check not run". go-test-json and tap say when a parent failed only by its subtests.
//   - The base is checked out with the repository's own configuration (its filters run, as in the user's checkout);
//     the user's global configuration is not read but for the line-end settings and excludes file, so a filter only
//     the user's global configuration defines (Git LFS installed globally) leaves its files as pointers at the base,
//     where the tests reading them fail: an old test failing makes every new test "check not run".
//   - The process that made a worktree is told by its pid alone: a killed check's worktree whose pid now names another
//     process is left until that process ends. A check killed in the instant git worktree add has made its record but
//     not yet the worktree's link can leave that record (a few bytes in the git folder's worktrees/, which git
//     worktree list does not show).
//   - The base runs in the system's temporary folder, not beside the checkout: a build that reaches outside the
//     checkout by a relative path (a go.mod replace of ../x) does not build there, and every new test is "check not
//     run". A folder of the base that is a link is never copied through.
//   - A test whose name changes from run to run (a commit id or a time in a subtest's name) is new at every rise, and
//     "check not run", never failed: the base run names it otherwise.
//
// Every path printed uses forward slashes; the output is byte-stable (names in the HEAD run's order, no map-order
// iteration).
package newtests

import (
	"context"
	"fmt"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/rungjob"
	"github.com/LastStep/Bonsai/internal/testout"
	"github.com/LastStep/Bonsai/internal/vergit"
)

// MaxNames is how many names of one group (failed on the base, proves nothing new, check not run) a Result keeps.
const MaxNames = 20

// gitWait is the longest the git calls before the base runs (finding the files, making the worktree) may take, and
// separately the removal's.
const gitWait = 5 * time.Minute

// Input is one rung's check.
type Input struct {
	// Root is the climbed checkout's top folder, absolute, with the system's separators: its working tree holds the
	// branch's test files, and its repository gives the base worktree.
	Root string
	// Rung is the rung as bonsai.yaml defines it: Command (run at the base as it is), Tests (its form, which names the
	// tests and the test files), BaseSetup (run first at the base, or nil) and TimeoutS (each base run's limit).
	Rung format.Rung
	// Base is the base commit, a full commit id (ChooseBase); "" when there is none.
	Base string
	// Head is the HEAD run's tests, as testout read the rung's standard output in its form.
	Head testout.Report
	// Rise is how far the ratchet's count rose above its floor (now - was): new_tests' count when the new tests cannot
	// be named (no base, the base could not be made or set up, its own run could not name its tests, or the test files
	// cannot be told).
	Rise int64
	// Env is the rung's environment, NAME=value; nil for this process's. rungjob adds BONSAI_LADDER=1.
	Env []string
	// Stop is the stop signal: when done, a running base run is ended as at its timeout and the check is not run;
	// nil for none.
	Stop context.Context
	// Git is the git program, "" for git on the PATH.
	Git string
}

// Result is one rung's check: what the ratchet's new_tests carries, and the names for the human output and the
// Bless ask.
type Result struct {
	// NewTests is the ratchet's new_tests (bonsai.ladder/1): the base ("" when none), the count, and how many failed
	// on the base, passed there and were not run. Count is always the sum of the three.
	NewTests format.LadderNewTests
	// Named is true when the base's own run named the new tests (Count is how many); false when they could not be
	// named, and Count is Input.Rise.
	Named bool
	// Failed, Passed and NotRun are the new tests' names in each group, in the HEAD run's order, at most MaxNames each
	// (the counts are NewTests'). NotRun is empty when the tests were not named.
	Failed, Passed, NotRun []string
	// Why says why every new test is "check not run" ("no base", "base_setup failed (exit 3): ..."); "" when the
	// check ran and judged each test.
	Why string
	// Notes are what the check could not do that changes no count (a worktree it could not remove, a git setting it
	// could not read), one line of ASCII each.
	Notes []string
	// Took is the check's whole time: the worktree, base_setup, the two base runs and the removal.
	Took time.Duration
}

// commitID is a full commit id, SHA-1 or SHA-256.
var commitID = regexp.MustCompile(`^([0-9a-f]{40}|[0-9a-f]{64})$`)

// ChooseBase is the base the check runs on (note 2): the merge base of HEAD with the base branch when it is not HEAD
// itself; else (on the base branch itself, or with no merge base) the base the task's last result recorded, read from
// its ver-git rungs' captures through vergit.RecordedBase (the first such rung, in the result's order, that recorded
// one); else "". head is HEAD's commit; prior is the task's last result, or nil.
func ChooseBase(mergeBase, head string, prior *format.Ladder) string {
	mergeBase = strings.ToLower(mergeBase)
	if commitID.MatchString(mergeBase) && mergeBase != strings.ToLower(head) {
		return mergeBase
	}
	if prior == nil {
		return ""
	}
	for _, r := range prior.Rungs {
		if r.Kind != "ver-git" {
			continue
		}
		if b := vergit.RecordedBase(r.Captures); b != "" {
			return b
		}
	}
	return ""
}

// Check runs the check for one rung (this package's comment). It never panics on a project's files and never leaves
// its worktree behind; what it cannot do is "check not run" with its reason.
func Check(in Input) (r Result) {
	start := time.Now()
	defer func() { r.Took = time.Since(start) }()
	c := &check{in: in, stop: in.Stop, res: &r}
	if c.stop == nil {
		c.stop = context.Background()
	}
	defer c.removeTree()
	c.run()
	return r
}

// check is one check's state.
type check struct {
	in   Input
	stop context.Context
	res  *Result
	form string
	git  *vergit.Runner // in Root, for the steps before the base runs (gitWait, ended by a stop)
	tree *tree          // the base worktree, nil until made
}

// run is the check's steps, each ending it with "check not run" when it cannot go on.
func (c *check) run() {
	in := c.in
	if commitID.MatchString(strings.ToLower(in.Base)) {
		c.res.NewTests.Base = strings.ToLower(in.Base)
	}
	if in.Rung.Tests == nil {
		c.notRun("the rung names no tests form, so its new tests cannot be named")
		return
	}
	c.form = *in.Rung.Tests
	if _, known := format.TestFileConventionOf(c.form); !known {
		c.notRun("the rung's tests form " + strconv.Quote(c.form) + " is not one Bonsai reads, so its new tests cannot be named")
		return
	}
	if in.Rung.Command == nil || strings.TrimSpace(*in.Rung.Command) == "" {
		c.notRun("the rung has no command to run at the base")
		return
	}
	switch {
	case in.Base == "":
		c.notRun("no base")
		return
	case c.res.NewTests.Base == "":
		c.notRun("no base: " + shown(in.Base) + " is not a commit id")
		return
	}
	if c.stopped() {
		return
	}
	// git's calls are not ended by a stop, which comes between them (a git worktree add cut short could leave its
	// record behind): each is quick, and the stop is read after them.
	ctx, cancel := context.WithTimeout(context.Background(), gitWait)
	defer cancel()
	g, note, err := vergit.NewRunner(ctx, in.Git, in.Root)
	if err != nil {
		c.notRun("the base could not be made: " + oneLine(err.Error()))
		return
	}
	c.git = g
	c.note(note)
	base := c.res.NewTests.Base
	if _, err := g.Run("cat-file", "-e", base+"^{commit}"); err != nil {
		if code := vergit.ExitCode(err); code == 1 || code == 128 {
			c.notRun("the base " + short(base) + " is not in this clone (a shallow clone, or history rewritten)")
		} else {
			c.notRun("the base could not be made: " + oneLine(err.Error()))
		}
		return
	}
	files, why, err := c.changedFiles(nil)
	if err != nil {
		c.notRun("the branch's test files could not be listed: " + oneLine(err.Error()))
		return
	}
	if why != "" {
		c.notRun("the test files cannot be told: " + why)
		return
	}
	c.sweep()
	if c.stopped() {
		return
	}
	t, err := makeTree(g, in.Root, base)
	c.tree = t
	if err != nil {
		c.notRun("the base could not be made: " + oneLine(err.Error()))
		return
	}
	if c.stopped() {
		return
	}

	if in.Rung.BaseSetup != nil && strings.TrimSpace(*in.Rung.BaseSetup) != "" {
		res := c.runAt(*in.Rung.BaseSetup)
		if !res.Green() {
			c.notRun("base_setup " + ended(res, c.timeout()))
			return
		}
	}

	first := c.runAt(*in.Rung.Command)
	firstRep, _ := testout.Read(c.form, first.Stdout, t.dir)
	if why := unnamed(first, firstRep, c.timeout()); why != "" {
		c.notRun("the base's own run " + why)
		return
	}
	news := newTests(in.Head, firstRep)
	c.res.Named = true
	if len(news) == 0 {
		return // no new tests: every test the climb passed ran on the base too
	}
	if old := firstFailed(firstRep); old != "" {
		c.allNotRun(news, "an old test failed at the base, before the branch's test files were copied in: "+shown(old))
		return
	}
	if c.form == format.TestsJUnit {
		// the base's report names the test files the branch deleted
		files, why, err = c.changedFiles(&firstRep)
		if err != nil {
			c.allNotRun(news, "the branch's test files could not be listed: "+oneLine(err.Error()))
			return
		}
		if why != "" {
			c.allNotRun(news, "the test files cannot be told: "+why)
			return
		}
	}
	if err := t.copyIn(in.Root, files); err != nil {
		c.allNotRun(news, "the branch's test files could not be copied in: "+oneLine(err.Error()))
		return
	}

	second := c.runAt(*in.Rung.Command)
	if second.TimedOut || second.Stopped || second.Exit == -1 {
		c.allNotRun(news, "the base run with the branch's test files "+ended(second, c.timeout()))
		return
	}
	secondRep, _ := testout.Read(c.form, second.Stdout, t.dir)
	c.judge(news, firstRep, secondRep)
}

// runAt runs one command in the base worktree through its rung job, with the rung's timeout and environment.
func (c *check) runAt(command string) rungjob.Result {
	return rungjob.Run(c.stop, rungjob.Spec{Command: command, Dir: c.tree.dir, Env: c.in.Env, Timeout: c.timeout()})
}

// timeout is each base run's limit: the rung's timeout_s, else the rung job's default (600 s), as the runner's.
func (c *check) timeout() time.Duration {
	if t := c.in.Rung.TimeoutS; t != nil && *t > 0 {
		return time.Duration(*t) * time.Second
	}
	return rungjob.DefaultTimeout
}

// stopped ends the check when the stop signal has come.
func (c *check) stopped() bool {
	if c.stop.Err() == nil {
		return false
	}
	c.notRun("the climb was stopped")
	return true
}

// notRun ends the check before the new tests were named: count is the rise, every one not run.
func (c *check) notRun(why string) {
	why = c.stopWhy(why)
	n := c.in.Rise
	if n < 0 {
		n = 0
	}
	c.res.Named = false
	c.res.NewTests.Count, c.res.NewTests.NotRun = n, n
	c.res.NewTests.FailedOnBase, c.res.NewTests.PassedOnBase = 0, 0
	c.res.Failed, c.res.Passed, c.res.NotRun = nil, nil, nil
	c.res.Why = why
}

// allNotRun ends the check once the new tests were named: every one not run.
func (c *check) allNotRun(news []string, why string) {
	why = c.stopWhy(why)
	c.res.NewTests.Count, c.res.NewTests.NotRun = int64(len(news)), int64(len(news))
	c.res.NotRun = keep(news)
	c.res.Why = why
}

// stopWhy is why, or once the stop signal has come "the climb was stopped" in its place: a git call or a run that
// failed after a stop failed by it ("git did not answer in time" would mislead).
func (c *check) stopWhy(why string) string {
	if c.stop != nil && c.stop.Err() != nil && !strings.Contains(why, "stopped") {
		return "the climb was stopped"
	}
	return why
}

// judge sorts the named new tests by what the base run with the branch's test files reported (step 4).
func (c *check) judge(news []string, first, second testout.Report) {
	if old := oldFailed(first, second); old != "" {
		c.allNotRun(news, "an old test failed at the base with the branch's test files copied in: "+shown(old))
		return
	}
	outcome := outcomes(second)
	var failed, passed, notRun []string
	for _, n := range news {
		switch outcome[n] {
		case testout.Failed:
			failed = append(failed, n)
		case testout.Passed:
			passed = append(passed, n)
		default:
			notRun = append(notRun, n)
		}
	}
	nt := &c.res.NewTests
	nt.Count = int64(len(news))
	nt.FailedOnBase, nt.PassedOnBase, nt.NotRun = int64(len(failed)), int64(len(passed)), int64(len(notRun))
	c.res.Failed, c.res.Passed, c.res.NotRun = keep(failed), keep(passed), keep(notRun)
}

// note keeps a note, one line of ASCII.
func (c *check) note(s string) {
	if s = oneLine(s); s != "" {
		c.res.Notes = append(c.res.Notes, s)
	}
}

// removeTree removes the base worktree, with a context of its own (a stop has no say over the removal).
func (c *check) removeTree() {
	if c.tree == nil {
		return
	}
	for _, n := range c.tree.remove(c.in.Git, c.in.Root) {
		c.note(n)
	}
	c.tree = nil
}

// unnamed says why a base run as it is cannot name the base's tests, "" when it can (step 2): it ran to its end and
// its output was read whole, with no package broken, and it ended green or red by its failing tests alone.
func unnamed(res rungjob.Result, rep testout.Report, limit time.Duration) string {
	switch {
	case res.TimedOut || res.Stopped || res.Exit == -1:
		return ended(res, limit)
	case res.StdoutSize > int64(len(res.Stdout)):
		return fmt.Sprintf("printed %d bytes, past the %d kept, so its tests cannot all be named", res.StdoutSize, rungjob.OutputCap)
	case len(rep.Broken) > 0:
		b := rep.Broken[0]
		what := "printed no report it can read"
		if b.What != "" {
			what = "failed outside its tests in the package " + shown(b.What)
			if b.Build {
				what = "did not build the package " + shown(b.What)
			}
		}
		return what + ": " + oneLine(b.Error)
	case !res.Green() && failedCount(rep) == 0:
		return "ended with exit " + strconv.Itoa(res.Exit) + " and named no failing test" + lastLine(res)
	}
	return ""
}

// ended says how a base run that was not green ended: "failed (exit 3): <its last line>", "timed out after 600 s",
// "was stopped (the climb was stopped)", "could not start: <why>".
func ended(res rungjob.Result, limit time.Duration) string {
	switch {
	case res.TimedOut:
		return "timed out after " + strconv.FormatFloat(limit.Seconds(), 'f', -1, 64) + " s"
	case res.Stopped:
		return "was stopped (the climb was stopped)"
	case res.Exit == -1:
		why := strings.TrimPrefix(res.Reason, "could not start: ")
		return "could not start: " + oneLine(why)
	}
	return "failed (exit " + strconv.Itoa(res.Exit) + ")" + lastLine(res)
}

// lastLine is ": " and a run's last output line, or "".
func lastLine(res rungjob.Result) string {
	if len(res.Tail) == 0 {
		return ""
	}
	return ": " + oneLine(res.Tail[len(res.Tail)-1])
}

// newTests are the tests the HEAD run reports passed whose names the base's own run does not report, distinct, in
// the HEAD run's order.
func newTests(head, base testout.Report) []string {
	there := map[string]bool{}
	for _, t := range base.Tests {
		there[t.Name] = true
	}
	seen := map[string]bool{}
	var out []string
	for _, t := range head.Tests {
		if t.Outcome != testout.Passed || there[t.Name] || seen[t.Name] {
			continue
		}
		seen[t.Name] = true
		out = append(out, t.Name)
	}
	return out
}

// firstFailed is the first test of a run that failed on its own (not only by a subtest), "" for none.
func firstFailed(rep testout.Report) string {
	for _, t := range rep.Tests {
		if t.Outcome == testout.Failed && !t.Parent {
			return t.Name
		}
	}
	return ""
}

// failedCount is how many tests of a run failed on their own.
func failedCount(rep testout.Report) int {
	n := 0
	for _, t := range rep.Tests {
		if t.Outcome == testout.Failed && !t.Parent {
			n++
		}
	}
	return n
}

// oldFailed is the first test of the second run that the first run also reported and that failed on its own (a
// parent failing only by its subtests is judged by them), "" for none.
func oldFailed(first, second testout.Report) string {
	old := map[string]bool{}
	for _, t := range first.Tests {
		old[t.Name] = true
	}
	for _, t := range second.Tests {
		if old[t.Name] && t.Outcome == testout.Failed && !t.Parent {
			return t.Name
		}
	}
	return ""
}

// outcomes is each name's outcome in a run: failed when any test of the name failed, else passed when any passed,
// else what it reported (skipped, todo).
func outcomes(rep testout.Report) map[string]string {
	out := map[string]string{}
	for _, t := range rep.Tests {
		switch {
		case out[t.Name] == testout.Failed:
		case t.Outcome == testout.Failed, t.Outcome == testout.Passed:
			out[t.Name] = t.Outcome
		case out[t.Name] == "":
			out[t.Name] = t.Outcome
		}
	}
	return out
}

// keep is a group's names as a Result keeps them: at most MaxNames.
func keep(names []string) []string {
	if len(names) > MaxNames {
		names = names[:MaxNames]
	}
	return append([]string(nil), names...)
}

// short is a commit's first 12 characters.
func short(sha string) string {
	if len(sha) > 12 {
		return sha[:12]
	}
	return sha
}

// oneLine is a text's first non-empty line, trimmed and made ASCII (every other byte a ?); on Windows its
// backslashes are made forward slashes, for the paths it names.
func oneLine(s string) string {
	for _, l := range strings.Split(strings.ReplaceAll(s, "\r\n", "\n"), "\n") {
		if l = strings.TrimSpace(l); l != "" {
			if runtime.GOOS == "windows" {
				l = strings.ReplaceAll(l, `\`, "/")
			}
			return ascii(l)
		}
	}
	return ""
}

// ascii replaces every byte outside printable ASCII with ?.
func ascii(s string) string {
	b := []byte(s)
	for i := range b {
		if b[i] < 0x20 || b[i] > 0x7e {
			b[i] = '?'
		}
	}
	return string(b)
}

// shown is a name as a line prints it: as it is when it is printable ASCII, else quoted with Go's escapes.
func shown(name string) string {
	for i := 0; i < len(name); i++ {
		if name[i] < 0x20 || name[i] > 0x7e {
			return strconv.QuoteToASCII(name)
		}
	}
	return name
}
