package ladder

// The rungs climbed (notes 2 to 5): in number order, a command rung through its rung job, its tests read and judged,
// its captures and its ratchet's count; a guard or ver-git rung through kinds.go; each followed by the steps after a
// rung (kinds.go's afterRung), and after the last, the steps after the rungs (kinds.go's afterRungs: the rise step). A
// required rung that is not green marks every later rung skipped.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/leftovers"
	"github.com/LastStep/Bonsai/internal/redact"
	"github.com/LastStep/Bonsai/internal/rungjob"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/testout"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// The rung statuses Bonsai writes (format.RungStatuses, their one home, holds what each means).
const (
	statusGreen   = "green"
	statusRed     = "red"
	statusSkipped = "skipped"
	statusError   = "error"
)

// FailedKept is how many failing tests a rung's tests keep in the result (today's 50); failed_count says how many in
// all. FailedPrinted is how many the human lines print.
const (
	FailedKept    = 50
	FailedPrinted = 20
)

// markPattern reads a test's mark: the start of its skip reason, <word>-only: (real-checkouts-only: ...), as today.
var markPattern = sync.OnceValue(func() *regexp.Regexp { return regexp.MustCompile(`^([a-z0-9]+(?:-[a-z0-9]+)*-only):`) })

// ratchetName is the form of a ratchet's name (bonsai.workspace/1's ratchet and capture descriptions).
var ratchetName = sync.OnceValue(func() *regexp.Regexp { return regexp.MustCompile(`^[a-z_][a-z0-9_]{0,39}$`) })

// jsonNumber is the form of a JSON number: a capture's value that reads as one is a number.
var jsonNumber = sync.OnceValue(func() *regexp.Regexp {
	return regexp.MustCompile(`^-?(0|[1-9][0-9]*)(\.[0-9]+)?([eE][+-]?[0-9]+)?$`)
})

// climb climbs the rungs and reads git at the start and the end (this package's comment, steps 6 and 8).
func (c *Climb) climb() {
	c.started = c.Now()
	c.git = gitState(c.Root)
	c.git.Dirty = treeDirty(c)
	c.printHeader()
	var entries []*format.LadderRung
	stopWhy := ""
	for _, r := range c.rungs {
		e := describe(r)
		entries = append(entries, e)
		if stopWhy == "" && c.o.Stop.Err() != nil {
			c.stopped = true
			stopWhy = "the climb was stopped before this rung (a signal)"
		}
		if stopWhy != "" {
			e.Status, e.Reason = statusSkipped, &stopWhy
			c.steps = append(c.steps, &Step{Rung: r, Entry: e})
			c.printRung(r, e, nil)
			continue
		}
		start := time.Now()
		var s *Step
		var notes []string
		if r.Kind == KindCommand {
			s, notes = c.runCommand(r, e)
		} else {
			s = c.runOther(r, e)
			if s.note != "" {
				notes = append(notes, s.note)
			}
		}
		ms := time.Since(start).Milliseconds()
		e.DurationMS = &ms
		afterRung(c, s)
		c.steps = append(c.steps, s)
		c.printRung(r, e, notes)
		if e.Status != statusGreen && r.Required {
			stopWhy = "an earlier required rung failed (rung " + strconv.FormatInt(r.Rung, 10) + ")"
		}
		if c.o.Stop.Err() != nil {
			c.stopped = true
			if stopWhy == "" {
				stopWhy = "the climb was stopped before this rung (a signal)"
			}
		}
	}
	afterRungs(c)
	for _, e := range entries {
		c.results = append(c.results, *e)
	}
	c.headAfter = head(c.Root)
	c.dirtyAfter = treeDirty(c)
	c.configMoved = !bytes.Equal(mainConfigBytes(c.Main), c.configBytes)
	if !c.CI && !c.o.nested() {
		v := leftovers.Check(c.Root)
		c.leftovers, c.leftLines = v.Record(), v.Lines()
	}
}

// describe is a rung's entry in the result before it runs: what bonsai.yaml says of it.
func describe(r format.Rung) *format.LadderRung {
	e := &format.LadderRung{Rung: r.Rung, Kind: r.Kind, Required: r.Required, Means: r.Means}
	if r.Name != nil {
		e.Name = *r.Name
	}
	if r.Kind == KindCommand {
		e.Command = r.Command
	}
	return e
}

// runOther runs a guard or ver-git rung through kinds.go and copies its outcome into the entry.
func (c *Climb) runOther(r format.Rung, e *format.LadderRung) *Step {
	o := runKind(c, r)
	e.Status = o.Status
	if e.Status == "" {
		e.Status = statusError
	}
	s := &Step{Rung: r, Entry: e, next: o.Next, who: o.NextWho}
	if e.Status != statusGreen {
		why := o.Reason
		if why == "" {
			why = "the rung ended " + e.Status + " and gave no reason"
		}
		e.Reason = text(why)
	} else {
		s.note = o.Reason
	}
	e.Captures = o.Captures
	if o.Hides {
		c.hides = append(c.hides, strconv.FormatInt(r.Rung, 10))
	}
	return s
}

// sessionVars are the variables that name a session's task and role. No rung sees them: a test run under a climb must
// never read the caller's own task (the climb's own reading of BONSAI_TASK, c.Getenv, is not changed by this).
var sessionVars = []string{workspace.TaskEnv, "BONSAI_ROLE"}

// rungEnv is the environment a rung runs in: the caller's, less sessionVars (the rest passes through unchanged, in
// order). Windows names variables in any case.
func rungEnv(env []string) []string {
	out := make([]string, 0, len(env))
next:
	for _, kv := range env {
		name, _, _ := strings.Cut(kv, "=")
		for _, v := range sessionVars {
			if name == v || (runtime.GOOS == "windows" && strings.EqualFold(name, v)) {
				continue next
			}
		}
		out = append(out, kv)
	}
	return out
}

// runCommand runs a command rung through its rung job (note 2), reads its tests (note 3) and judges their skips and
// marks (note 4), reads its captures and its ratchet's count (note 5). notes are what the job could not do (never a
// colour), for the human lines.
func (c *Climb) runCommand(r format.Rung, e *format.LadderRung) (*Step, []string) {
	s := &Step{Rung: r, Entry: e}
	if r.Command == nil || strings.TrimSpace(*r.Command) == "" {
		e.Status, e.Reason = statusError, text("the rung's kind is command, but bonsai.yaml gives it no command")
		return s, nil
	}
	res := rungjob.Run(c.o.Stop, rungjob.Spec{Command: *r.Command, Dir: c.Root, Env: rungEnv(c.o.Env), Timeout: rungTimeout(r)})
	s.Output, s.timed = res.Out, res.TimedOut
	var reason []string
	switch {
	case res.Green():
		e.Status = statusGreen
	case res.Exit == -1 && !res.TimedOut && !res.Stopped:
		e.Status = statusError // it could not start
	default:
		e.Status = statusRed
	}
	if res.Reason != "" && e.Status != statusGreen {
		reason = append(reason, res.Reason)
	}
	if res.Exit > 0 {
		reason = append(reason, "exit "+strconv.Itoa(res.Exit))
	}
	accepted := 0
	named := false // the tests read explain the red: a failing test, a broken package, or a test result not accepted
	if r.Tests != nil {
		if rep, ok := testout.Read(*r.Tests, res.Stdout, c.Root); ok {
			s.Tests = &rep
			var problems []string
			e.Tests, problems, accepted = c.judgeTests(r, rep)
			for _, b := range rep.Broken {
				named = true
				what := "the output is not read as " + *r.Tests
				if b.What != "" {
					what = "the package " + b.What + " failed outside its tests"
					if b.Build {
						what = "the package " + b.What + " did not build"
					}
				}
				problems = append(problems, what+": "+b.Error)
				if e.Status == statusGreen {
					e.Status = statusRed
				}
			}
			if n := failedCount(rep); n > 0 {
				named = true
				reason = append(reason, fmt.Sprintf("%d %s failed (each in tests.failed, with the command that runs it alone)", n,
					plural(n, "test", "tests")))
			}
			if len(problems) > 0 {
				named = true
				if e.Status == statusGreen {
					e.Status = statusRed
				}
				reason = append(reason, fmt.Sprintf("%d test %s this climb cannot accept:", len(problems), plural(len(problems), "result", "results")))
				for _, p := range problems {
					reason = append(reason, "  - "+p)
				}
			}
			if res.StdoutSize > int64(len(res.Stdout)) {
				reason = append(reason, fmt.Sprintf("the standard output was %d bytes, past the %d kept: the tests past that were not read",
					res.StdoutSize, rungjob.OutputCap))
			}
		}
	}
	e.Captures = captures(r.Capture, res.Out)
	s.Count = c.count(r, e.Captures, s.Tests, accepted)
	if e.Status != statusGreen {
		if !named && len(res.Tail) > 0 {
			reason = append(reason, "the output's last lines:")
			for _, l := range res.Tail {
				reason = append(reason, "  | "+l)
			}
		}
		if len(reason) == 0 {
			reason = append(reason, "the command did not end green")
		}
		e.Reason = text(strings.Join(reason, "\n"))
	}
	return s, res.Notes
}

// judgeTests judges a rung's skips and todos (note 4, today's judgeTests) and gives its tests object (today's shape),
// the problems that make it red, and how many skips count toward its ratchet (under --ci, the accepted ones).
//   - A local climb is a task's proof and runs every test: a skip with a mark, or of a test named in ci_marked_tests, is
//     red; an unmarked skip passes.
//   - Under --ci a skip counts only when its test is named in ci_marked_tests and its reason carries a mark; any other
//     skip, and any todo, is red. A rung with a ratchet that printed no test is red too: its skips cannot be judged.
func (c *Climb) judgeTests(r format.Rung, rep testout.Report) (schema.Object, []string, int) {
	pinned := map[string]bool{}
	for _, n := range c.Config.Full.CIMarkedTests {
		pinned[n] = true
	}
	skipped, todo := []any{}, []any{}
	var problems []string
	counted := 0
	for _, t := range rep.Tests {
		switch t.Outcome {
		case testout.Todo:
			todo = append(todo, schema.Object{{Key: "name", Value: t.Name}, {Key: "reason", Value: redact.Text(t.Reason)}})
			if c.CI {
				problems = append(problems, quoteName(t.Name)+" is a todo"+inParens(t.Reason)+": a todo never passes --ci")
			}
		case testout.Skipped:
			mark := ""
			if m := markPattern().FindStringSubmatch(t.Reason); m != nil {
				mark = m[1]
			}
			accepted := false
			switch {
			case !c.CI:
				if mark != "" || pinned[t.Name] {
					by := "marked " + mark
					if mark == "" {
						by = "named in bonsai.yaml's ci_marked_tests"
					}
					problems = append(problems, quoteName(t.Name)+" skipped"+inParens(t.Reason)+": it is "+by+", and a local climb, a task's "+
						"proof, runs every marked test")
				}
			case mark == "":
				problems = append(problems, quoteName(t.Name)+" skipped without a mark"+inParens(t.Reason)+": under --ci only a test named in "+
					"bonsai.yaml's ci_marked_tests whose reason carries a mark may skip")
			case !pinned[t.Name]:
				problems = append(problems, quoteName(t.Name)+" is marked "+mark+" but not named in bonsai.yaml's ci_marked_tests")
			default:
				accepted = true
				counted++
			}
			var markV any
			if mark != "" {
				markV = mark
			}
			item := schema.Object{{Key: "name", Value: t.Name}, {Key: "mark", Value: markV}, {Key: "reason", Value: redact.Text(t.Reason)}}
			if c.CI {
				item = append(item, schema.Member{Key: "accepted", Value: accepted})
				c.skipped = append(c.skipped, format.LadderSkip{Rung: r.Rung, Name: t.Name, Mark: strPtr(mark), Reason: redact.Text(t.Reason),
					Accepted: accepted})
			}
			skipped = append(skipped, item)
		}
	}
	if c.CI && r.Ratchet != nil && len(rep.Tests) == 0 {
		problems = append(problems, "it printed no tests, so its skips cannot be judged (under --ci a rung with a ratchet must name its tests)")
	}
	tests := schema.Object{{Key: "seen", Value: int64(len(rep.Tests))}, {Key: "skipped", Value: skipped}, {Key: "todo", Value: todo}}
	if n := failedCount(rep); n > 0 {
		in := testout.RunInput{Command: *r.Command, Root: c.Root, Windows: windows()}
		failed := []any{}
		for _, t := range rep.Tests {
			if t.Outcome != testout.Failed || t.Parent {
				continue
			}
			if len(failed) == FailedKept {
				break
			}
			var file, run any
			if t.File != "" {
				file = t.File
			}
			if cmd := testout.Run(rep.Form, t, in); cmd != "" {
				run = cmd
			}
			failed = append(failed, schema.Object{{Key: "name", Value: t.Name}, {Key: "file", Value: file},
				{Key: "error", Value: redact.Text(t.Error)}, {Key: "run", Value: run}})
		}
		tests = append(tests, schema.Member{Key: "failed", Value: failed}, schema.Member{Key: "failed_count", Value: int64(n)})
	}
	return tests, problems, counted
}

// failedCount is how many tests failed on their own (a parent failing only by its subtests is not counted).
func failedCount(rep testout.Report) int {
	n := 0
	for _, t := range rep.Tests {
		if t.Outcome == testout.Failed && !t.Parent {
			n++
		}
	}
	return n
}

// captures reads a rung's capture patterns against its output (bonsai.workspace/1's capture): each pattern, a Go
// (RE2) regular expression, is matched line by line; the last line it matches gives its first group (the whole match
// when it has none), a number when it reads as one, else text (redacted). The last, since a runner prints its summary
// after every test's own output: a test that prints a line like the summary's ("pass 50", which node's TAP gives as
// "# pass 50") prints it before the runner's own. A pattern that does not compile gives null; one that matches no line
// is left out. null when the rung captures nothing, or nothing matched.
func captures(spec any, out []byte) schema.Object {
	m, ok := spec.(schema.Object)
	if !ok || len(m) == 0 {
		return nil
	}
	var lines []string
	got := schema.Object{}
	for _, kv := range m {
		pat, isText := kv.Value.(string)
		if !isText {
			got = append(got, schema.Member{Key: kv.Key, Value: nil})
			continue
		}
		re, err := regexp.Compile(pat)
		if err != nil {
			got = append(got, schema.Member{Key: kv.Key, Value: nil})
			continue
		}
		if lines == nil {
			lines = strings.Split(strings.ReplaceAll(string(out), "\r\n", "\n"), "\n")
		}
		for i := len(lines) - 1; i >= 0; i-- {
			sm := re.FindStringSubmatch(lines[i])
			if sm == nil {
				continue
			}
			v := sm[0]
			if len(sm) > 1 {
				v = sm[1]
			}
			if t := strings.TrimSpace(v); jsonNumber().MatchString(t) {
				got = append(got, schema.Member{Key: kv.Key, Value: json.Number(t)})
			} else {
				got = append(got, schema.Member{Key: kv.Key, Value: redact.Text(v)})
			}
			break
		}
	}
	if len(got) == 0 {
		return nil
	}
	return got
}

// count reads a rung's ratchet and its count (5.4.0 note 2): the capture of the ratchet's name when it is a number,
// else, on a rung whose tests were read, the tests named as passed plus the accepted skips (under --ci; 0 locally).
// On a rung whose tests were read, a count from the capture comes with the tests named as passed (Parsed), which
// internal/ratchet's Judge holds it to: two reads of one output must agree. nil when the rung has no ratchet; Now ""
// when none could be read (or the ratchet is not a name).
func (c *Climb) count(r format.Rung, caps schema.Object, rep *testout.Report, accepted int) *Count {
	if r.Ratchet == nil {
		return nil
	}
	name, _ := r.Ratchet.(string)
	if !ratchetName().MatchString(name) {
		return &Count{Name: name, Floor: "0"}
	}
	floor, _ := floorOf(c.Config.Full.Ratchets, name) // a floor that is not a number makes the rung red (judgeRatchet)
	k := &Count{Name: name, Floor: floor}
	if v, ok := caps.Get(name); ok {
		if n, isNum := v.(json.Number); isNum {
			k.Now, k.From = n, "capture"
			if rep != nil {
				k.Parsed, k.Form = json.Number(strconv.Itoa(rep.Passed())), rep.Form
			}
			return k
		}
	}
	if rep != nil {
		k.Now, k.From = json.Number(strconv.Itoa(rep.Passed()+accepted)), "tests"
	}
	return k
}

// floorOf is a ratchet's floor in bonsai.yaml's ratchets, 0 when it names none. bad is the floor as written when it is
// not a number: it is never read as 0, which would let a typo lower the floor, and judgeRatchet makes the rung red
// (bonsai.yaml's full read refuses such a floor before a climb, so only a Config built otherwise holds one).
func floorOf(ratchets schema.Object, name string) (floor json.Number, bad string) {
	v, ok := ratchets.Get(name)
	if !ok {
		return "0", ""
	}
	switch n := v.(type) {
	case json.Number:
		if jsonNumber().MatchString(string(n)) {
			return n, ""
		}
	case string:
		if jsonNumber().MatchString(n) {
			return json.Number(n), ""
		}
	case int64:
		return json.Number(strconv.FormatInt(n, 10)), ""
	case int:
		return json.Number(strconv.Itoa(n)), ""
	case float64:
		if !math.IsNaN(n) && !math.IsInf(n, 0) {
			return json.Number(strconv.FormatFloat(n, 'f', -1, 64)), ""
		}
	}
	return "0", schema.Show(v)
}

// text redacts a free-text value for the result (note 6), nil for "".
func text(s string) *string {
	if s == "" {
		return nil
	}
	r := redact.Text(s)
	return &r
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func quoteName(n string) string { return strconv.QuoteToASCII(n) }

func inParens(s string) string {
	if s == "" {
		return " (no reason)"
	}
	return " (" + s + ")"
}

func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}
