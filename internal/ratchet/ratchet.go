// Package ratchet is the ladder's floors, ratchets and Bless (spec §9, "Ratchets and Bless"; contract §9.1, §9.2,
// §10.3, §11; design/plan-5.md, 5.4.4): it decides a proof's colour on its counts, and offers each rise to a person for
// Bless. The ladder runner (internal/ladder's kinds.go) calls it after each rung and once the result is written.
//
// Bonsai's code never writes the floors (bonsai.yaml's ratchets). Where the studio manages a project, it applies a
// person's bless tap (contract §10.3); elsewhere an agent may raise a floor to a green, clean climb's count under a task
// that grants bonsai.yaml, since a raised floor only makes the ladder stricter, and lowering one waits for a person
// (note 4, the switch's stricter-only rule).
//
//   - ratchet.go (this file): one rung's ratchet judged (Judge). was is ratchets[name] (0 when it names none), now the
//     rung's count; below it the rung is red ("go_tests went backwards: 812 -> 805"), even with every test green. A
//     count that cannot be read is red too: a ratchet that measures nothing is none. Under --ci the accepted marked
//     skips count (the runner adds them, as today), and a red reason says how many. On a rung with a tests form whose
//     tests were read, a count taken from the capture must equal the tests read as passed from the same output; when
//     the two disagree the rung is red, since one of them read a line the code under test printed.
//   - rise.go: the rise step (NewTests): where a local, green, clean climb finds a ratchet above its floor, the one
//     call that gives the rise's new_tests, its human lines and its line for the Bless ask's why (internal/newtests'
//     check, step 5.4.5).
//   - offer.go: the Bless offer (OfferFor, today's blessOffer): the climb was local and green, in the main checkout
//     on the base branch at its HEAD with a clean tree, and at least one green rung's ratchet has now above its floor,
//     its name a key of ratchets; and the ask's words and data.
//   - bless.go: the Bless step after a climb's result is written (Bless): on the base branch, whatever its task, every
//     open ladder:bless-* ask is worked out again from its task's result and the floors as they now are, and resolved
//     when its offer is gone, with its reason; then the climb's own offer is filed (asks.FileBless), or its task's open
//     ask resolved. Never under --ci, nor on a branch; never filed from a red or dirty climb, nor from a result another
//     climb wrote (note 5).
//
// What a count cannot be kept from: the count is read from the output of the code under test, and a test can always
// print what its runner prints. A Go test that prints "\x16--- PASS: TestX (0.00s)" (the frame go test's own lines
// carry for test2json) makes go test -json report a pass for a test that never ran, so the tests read and any count
// from them grow with it; a node test's console.log("pass 50") reaches node's TAP as the comment "# pass 50", which a
// capture of the summary ("^# pass (\d+)") matches. The runner narrows this, never closes it: a capture takes the last
// line it matches (a runner prints its summary after every test's output), a capture on a rung whose tests were read
// must agree with them (one forged line moves one read, not both), and git integrity's test-file-changed names every
// changed test file for the verifier. A test file that forges both reads is code a reviewer reads: the ladder proves
// the tests a branch's own code runs, and a branch's tests are part of what is reviewed.
package ratchet

import (
	"encoding/json"
	"fmt"
	"math/big"
	"regexp"
	"strconv"
	"sync"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/schema"
)

// namePattern is the form of a ratchet's name (bonsai.workspace/1's ratchet and capture descriptions, as today).
var namePattern = sync.OnceValue(func() *regexp.Regexp { return regexp.MustCompile(`^[a-z_][a-z0-9_]{0,39}$`) })

// IsName reports whether s is a ratchet's name.
func IsName(s string) bool { return namePattern().MatchString(s) }

// Reading is one rung's ratchet as the runner read it (internal/ladder's Count), with what the judgment needs to say
// why no count was read.
type Reading struct {
	Rung     format.Rung   // the rung as bonsai.yaml defines it: its kind, ratchet, capture and tests form
	Name     string        // the ratchet's name as the rung gives it, "" when its ratchet is not text
	Floor    json.Number   // ratchets[name] in bonsai.yaml, 0 when it names none
	Now      json.Number   // the count, "" when none could be read
	From     string        // "capture" or "tests", "" when none could be read
	Captures schema.Object // what the rung's captures read (the result's captures)
	Tests    schema.Object // the result's tests: under --ci its skipped list marks each accepted skip
	CI       bool          // a --ci climb, whose count holds the accepted marked skips
	Parsed   json.Number   // with a count from the capture on a rung whose tests were read: the tests read as passed; else ""
	Form     string        // the tests form Parsed was read in ("tap"), for the reason; "" with no Parsed
}

// Verdict is one rung's ratchet judged.
type Verdict struct {
	Ratchet *format.LadderRatchet // the rung's ratchet in the result (new_tests null: the rise step's); nil when no count was read
	Red     bool                  // the count is below its floor, or none was read: the rung is red
	Reason  string                // why it is red, "" when it is not
	Next    string                // the fix a red climb's next step names, when this is the rung's only fault
	Who     string                // who takes it: "" for an agent, "person" for a change to bonsai.yaml's ladder
}

// Judge judges one rung's ratchet (this file's comment).
func Judge(r Reading) Verdict {
	if r.Now == "" {
		why, who := unread(r)
		return Verdict{Red: true, Who: who,
			Reason: "the ratchet " + shownName(r) + " measured nothing: " + why + "; a ratchet that measures nothing is none",
			Next:   "make the ratchet " + shownName(r) + " count again: " + why + "; then climb again"}
	}
	floor, okFloor := Number(r.Floor)
	now, okNow := Number(r.Now)
	if !okFloor || !okNow {
		bad, what := string(r.Floor), "its floor in bonsai.yaml's ratchets"
		if !okNow {
			bad, what = string(r.Now), "its count"
		}
		return Verdict{Red: true, Who: "person",
			Reason: fmt.Sprintf("the ratchet %s cannot be judged: %s, %q, does not read as a number Bonsai can compare", shownName(r), what, bad),
			Next:   "a person writes ratchets." + r.Name + " in bonsai.yaml as a plain number; then run: bonsai check --json"}
	}
	v := Verdict{Ratchet: &format.LadderRatchet{Name: r.Name, Was: r.Floor, Now: r.Now, OK: now.Cmp(floor) >= 0}}
	if r.From == "capture" && r.Parsed != "" {
		if parsed, ok := Number(r.Parsed); !ok || parsed.Cmp(now) != 0 {
			v.Red = true
			v.Reason = fmt.Sprintf("the ratchet %s's two counts disagree: its capture read %s, and the tests read as %s name %s passed; a "+
				"count is taken only when both reads of the output agree, since a test can print a line its runner prints", shownName(r),
				r.Now, r.Form, r.Parsed)
			v.Next = "find the line the capture " + r.Name + " took (the last line it matches; a test's own output can print a runner's " +
				"summary line) and the test that printed it, and take that print out; if the capture counts something else than the " +
				"passing tests, a person changes it in bonsai.yaml or takes it out, so the tests read are the count"
			return v
		}
	}
	if v.Ratchet.OK {
		return v
	}
	v.Red = true
	v.Reason = r.Name + " went backwards: " + string(r.Floor) + " -> " + string(r.Now) + " (its floor is bonsai.yaml's ratchets." + r.Name +
		"; a count may only rise)"
	if n := AcceptedSkips(r.Tests); r.CI && r.From == "tests" && n > 0 {
		v.Reason += fmt.Sprintf("; under --ci the count holds %d accepted marked %s", n, plural(n, "skip", "skips"))
	}
	v.Next = "find why " + r.Name + " fell from " + string(r.Floor) + " to " + string(r.Now) + " (a test removed, renamed, skipped or no " +
		"longer run: git integrity's rung lists the changed test files) and bring it back; a floor is lowered only by a person, in " +
		"bonsai.yaml's ratchets"
	return v
}

// unread says why no count was read, and who fixes it ("person" when the fix is bonsai.yaml's ladder).
func unread(r Reading) (string, string) {
	if r.Rung.Kind != "command" {
		return "a " + r.Rung.Kind + " rung counts nothing, so a person moves the ratchet to a command rung in bonsai.yaml", "person"
	}
	if s, isText := r.Rung.Ratchet.(string); !isText || !IsName(s) {
		return "the rung's ratchet is not a ratchet's name (a lower-case letter or _, then lower-case letters, digits or _, at most 40 " +
			"in all), so a person names one in bonsai.yaml", "person"
	}
	capture, who := "", "person"
	pattern, defined := captureOf(r.Rung.Capture, r.Name)
	if defined {
		v, matched := r.Captures.Get(r.Name)
		switch {
		case !pattern || (matched && v == nil):
			capture = "its capture " + r.Name + " is not a pattern that compiles"
		case !matched:
			capture, who = "its capture "+r.Name+" matched no line of the output", ""
		default:
			capture, who = fmt.Sprintf("its capture %s read %s, not a number", r.Name, strconv.QuoteToASCII(schema.Show(v))), ""
		}
	} else {
		capture = "it has no capture named " + r.Name
	}
	// With a tests form Bonsai reads, the passing tests are a count; so a form here is one it does not read.
	tests := "and no tests form to count its passing tests"
	if r.Rung.Tests != nil {
		tests = "and its tests form " + strconv.QuoteToASCII(*r.Rung.Tests) + " is not one Bonsai reads (bonsai check --schema bonsai.workspace " +
			"names them)"
		who = "person"
	}
	return capture + ", " + tests, who
}

// captureOf reports whether a rung's capture names name, and whether its pattern is text.
func captureOf(spec any, name string) (pattern, defined bool) {
	m, ok := spec.(schema.Object)
	if !ok {
		return false, false
	}
	v, ok := m.Get(name)
	if !ok {
		return false, false
	}
	_, isText := v.(string)
	return isText, true
}

// shownName is the ratchet's name as a reason shows it: as written when it is a name, quoted when not.
func shownName(r Reading) string {
	if IsName(r.Name) {
		return r.Name
	}
	return strconv.QuoteToASCII(r.Name)
}

// AcceptedSkips counts the accepted marked skips in a rung's tests (the result's tests.skipped, each with accepted
// under --ci): what a --ci count holds besides the passing tests.
func AcceptedSkips(tests schema.Object) int {
	v, _ := tests.Get("skipped")
	list, _ := v.([]any)
	n := 0
	for _, it := range list {
		if o, ok := it.(schema.Object); ok {
			if a, _ := o.Get("accepted"); a == true {
				n++
			}
		}
	}
	return n
}

// Number reads a count or a floor exactly enough to compare two counts (256 bits); ok is false when it is not a
// number Bonsai can compare.
func Number(n json.Number) (*big.Float, bool) {
	f, ok := new(big.Float).SetPrec(256).SetString(string(n))
	return f, ok && !f.IsInf()
}

// Above reports whether now is above was, both read as numbers; false when either does not read.
func Above(now, was json.Number) bool {
	a, okA := Number(now)
	b, okB := Number(was)
	return okA && okB && a.Cmp(b) > 0
}

func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}
