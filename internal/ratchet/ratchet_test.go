package ratchet

// The ratchet's judgment (design/plan-5.md, 5.4.4 note 1), walked from a table: a count equal to its floor and above it
// green; below it red with "<name> went backwards: <was> -> <now>", even with every test green; a count that cannot be
// read red, each way it can fail to be read, with who fixes it; a floor that does not compare red; under --ci the
// accepted marked skips named in the reason.

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/schema"
)

func strp(s string) *string { return &s }

// commandRung is a command rung with the ratchet given, its capture and tests form ("" for none).
func commandRung(ratchet any, capture schema.Object, tests string) format.Rung {
	r := format.Rung{Rung: 3, Kind: "command", Command: strp("go test -json ./..."), Ratchet: ratchet}
	if capture != nil {
		r.Capture = capture
	}
	if tests != "" {
		r.Tests = &tests
	}
	return r
}

// skippedTests is a result's tests under --ci with n accepted marked skips and one not accepted.
func skippedTests(n int) schema.Object {
	list := []any{schema.Object{{Key: "name", Value: "TestNo"}, {Key: "accepted", Value: false}}}
	for i := 0; i < n; i++ {
		list = append(list, schema.Object{{Key: "name", Value: "TestYes"}, {Key: "accepted", Value: true}})
	}
	return schema.Object{{Key: "seen", Value: int64(9)}, {Key: "skipped", Value: list}}
}

func TestJudge(t *testing.T) {
	counted := commandRung("go_tests", nil, "go-test-json")
	captured := commandRung("found", schema.Object{{Key: "found", Value: "passed ([0-9]+)"}}, "")
	for _, c := range []struct {
		name   string
		in     Reading
		red    bool
		ok     bool   // the ratchet written, with ok
		none   bool   // no ratchet written (no count read)
		reason string // the reason holds this ("" for none)
		next   string // the next step holds this
		who    string
	}{
		{name: "equal", in: Reading{Rung: counted, Name: "go_tests", Floor: "812", Now: "812", From: "tests"}, ok: true},
		{name: "above", in: Reading{Rung: counted, Name: "go_tests", Floor: "812", Now: "815", From: "tests"}, ok: true},
		{name: "no floor and no count", in: Reading{Rung: counted, Name: "go_tests", Floor: "0", Now: "0", From: "tests"}, ok: true},
		{name: "below", in: Reading{Rung: counted, Name: "go_tests", Floor: "812", Now: "805", From: "tests"}, red: true,
			reason: "go_tests went backwards: 812 -> 805 (its floor is bonsai.yaml's ratchets.go_tests; a count may only rise)",
			next:   "find why go_tests fell from 812 to 805"},
		{name: "below by a capture", in: Reading{Rung: captured, Name: "found", Floor: "7", Now: "6.5", From: "capture"}, red: true,
			reason: "found went backwards: 7 -> 6.5"},
		{name: "below past float64", in: Reading{Rung: counted, Name: "go_tests", Floor: "100000000000000000000000000000001",
			Now: "100000000000000000000000000000000", From: "tests"}, red: true, reason: "went backwards"},
		{name: "below under --ci, with accepted skips", in: Reading{Rung: counted, Name: "go_tests", Floor: "812", Now: "810", From: "tests",
			Tests: skippedTests(2), CI: true}, red: true, reason: "; under --ci the count holds 2 accepted marked skips"},
		{name: "below under --ci, one accepted skip", in: Reading{Rung: counted, Name: "go_tests", Floor: "812", Now: "810", From: "tests",
			Tests: skippedTests(1), CI: true}, red: true, reason: "; under --ci the count holds 1 accepted marked skip"},
		{name: "unread: no capture and no tests form", in: Reading{Rung: commandRung("go_tests", nil, ""), Name: "go_tests", Floor: "812"},
			red: true, none: true, who: "person",
			reason: "the ratchet go_tests measured nothing: it has no capture named go_tests, and no tests form to count its passing " +
				"tests; a ratchet that measures nothing is none"},
		{name: "unread: the capture matched no line", in: Reading{Rung: captured, Name: "found", Floor: "7"}, red: true, none: true,
			reason: "its capture found matched no line of the output, and no tests form", next: "make the ratchet found count again"},
		{name: "unread: the capture read text", in: Reading{Rung: captured, Name: "found", Floor: "7",
			Captures: schema.Object{{Key: "found", Value: "many"}}}, red: true, none: true, reason: `its capture found read "\"many\"", not a number`},
		{name: "unread: the capture does not compile", in: Reading{Rung: commandRung("found", schema.Object{{Key: "found", Value: "("}}, ""),
			Name: "found", Captures: schema.Object{{Key: "found", Value: nil}}}, red: true, none: true, who: "person",
			reason: "its capture found is not a pattern that compiles"},
		{name: "unread: the capture is not text", in: Reading{Rung: commandRung("found", schema.Object{{Key: "found", Value: json.Number("3")}}, ""),
			Name: "found", Captures: schema.Object{{Key: "found", Value: nil}}}, red: true, none: true, who: "person",
			reason: "is not a pattern that compiles"},
		{name: "unread: a tests form Bonsai does not read", in: Reading{Rung: commandRung("go_tests", nil, "nunit"), Name: "go_tests"},
			red: true, none: true, who: "person", reason: `its tests form "nunit" is not one Bonsai reads`},
		{name: "unread: a guard rung", in: Reading{Rung: format.Rung{Rung: 0, Kind: "guard", Ratchet: "checked"}, Name: "checked"},
			red: true, none: true, who: "person", reason: "the ratchet checked measured nothing: a guard rung counts nothing"},
		{name: "unread: not a name", in: Reading{Rung: commandRung("Go-Tests", nil, "go-test-json"), Name: "Go-Tests"}, red: true, none: true,
			who: "person", reason: `the ratchet "Go-Tests" measured nothing: the rung's ratchet is not a ratchet's name`},
		{name: "unread: not text", in: Reading{Rung: commandRung(json.Number("3"), nil, "go-test-json")}, red: true, none: true,
			who: "person", reason: `the ratchet "" measured nothing: the rung's ratchet is not a ratchet's name`},
		{name: "a capture agreeing with the tests read", in: Reading{Rung: commandRung("tap_pass", schema.Object{{Key: "tap_pass",
			Value: `^# pass (\d+)`}}, "tap"), Name: "tap_pass", Floor: "3", Now: "3", From: "capture", Parsed: "3", Form: "tap"}, ok: true},
		{name: "a capture disagreeing with the tests read", in: Reading{Rung: commandRung("tap_pass", schema.Object{{Key: "tap_pass",
			Value: `^# pass (\d+)`}}, "tap"), Name: "tap_pass", Floor: "3", Now: "50", From: "capture", Parsed: "3", Form: "tap"}, red: true,
			ok: true, reason: "the ratchet tap_pass's two counts disagree: its capture read 50, and the tests read as tap name 3 passed; a " +
				"count is taken only when both reads of the output agree, since a test can print a line its runner prints",
			next: "find the line the capture tap_pass took"},
		{name: "a capture below its floor, agreeing", in: Reading{Rung: commandRung("tap_pass", schema.Object{{Key: "tap_pass",
			Value: `^# pass (\d+)`}}, "tap"), Name: "tap_pass", Floor: "4", Now: "3", From: "capture", Parsed: "3", Form: "tap"}, red: true,
			reason: "tap_pass went backwards: 4 -> 3"},
		{name: "a count from the tests is not held to itself", in: Reading{Rung: counted, Name: "go_tests", Floor: "3", Now: "5", From: "tests",
			Parsed: "4"}, ok: true},
		{name: "a floor past compare", in: Reading{Rung: counted, Name: "go_tests", Floor: "1e999999999", Now: "5", From: "tests"}, red: true,
			none: true, who: "person", reason: `its floor in bonsai.yaml's ratchets, "1e999999999", does not read as a number`},
	} {
		t.Run(c.name, func(t *testing.T) {
			v := Judge(c.in)
			if v.Red != c.red {
				t.Fatalf("red %v, want %v: %+v", v.Red, c.red, v)
			}
			switch {
			case c.none && v.Ratchet != nil:
				t.Errorf("a ratchet written with no count: %+v", v.Ratchet)
			case !c.none && (v.Ratchet == nil || v.Ratchet.Name != c.in.Name || v.Ratchet.Was != c.in.Floor || v.Ratchet.Now != c.in.Now ||
				v.Ratchet.OK != c.ok || v.Ratchet.NewTests != nil):
				t.Errorf("the ratchet: %+v", v.Ratchet)
			}
			if c.reason == "" && (v.Reason != "" || v.Next != "") || !strings.Contains(v.Reason, c.reason) {
				t.Errorf("reason %q, want %q", v.Reason, c.reason)
			}
			if c.red && (v.Next == "" || !strings.Contains(v.Next, c.next)) {
				t.Errorf("next %q, want %q", v.Next, c.next)
			}
			if v.Who != c.who {
				t.Errorf("who %q, want %q", v.Who, c.who)
			}
			if !c.in.CI && strings.Contains(v.Reason, "--ci") {
				t.Errorf("a local reason names --ci: %q", v.Reason)
			}
		})
	}
}

// Above and Number compare counts exactly: decimals, exponents, past float64, and what is no number.
func TestAbove(t *testing.T) {
	for _, c := range []struct {
		now, was string
		above    bool
	}{
		{"815", "812", true}, {"812", "812", false}, {"805", "812", false}, {"1.5", "1.25", true}, {"1e3", "999", true},
		{"100000000000000000000000000000001", "100000000000000000000000000000000", true}, {"0", "-0", false}, {"-1", "0", false},
		{"5", "", false}, {"", "0", false}, {"abc", "1", false}, {"1e999999999", "1", false},
	} {
		if got := Above(json.Number(c.now), json.Number(c.was)); got != c.above {
			t.Errorf("Above(%s, %s) = %v", c.now, c.was, got)
		}
	}
}

// AcceptedSkips counts the skips marked accepted, and nothing in tests that hold none.
func TestAcceptedSkips(t *testing.T) {
	if n := AcceptedSkips(skippedTests(3)); n != 3 {
		t.Errorf("%d", n)
	}
	for _, tests := range []schema.Object{nil, {{Key: "skipped", Value: "no"}}, {{Key: "skipped", Value: []any{"x", schema.Object{}}}}} {
		if n := AcceptedSkips(tests); n != 0 {
			t.Errorf("%s: %d", schema.Show(tests), n)
		}
	}
}
