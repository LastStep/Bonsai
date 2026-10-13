package ladder

// A ratchet's count read from the output of the code under test (V2's S2): a capture takes the last line it matches,
// so a test's own line printed before the runner's summary is not the count; and on a rung whose tests were read, a
// count from the capture must equal the tests read as passed, else the rung is red.

import (
	"strings"
	"testing"
)

// tapTwo is node's TAP for two passing tests, the first printing "pass 50" (node gives a test's output as a comment
// before its result), and the summary node prints after them, its pass count given by summary.
func tapTwo(summary string) string {
	return "TAP version 13\n# pass 50\n# Subtest: a\nok 1 - a\n  ---\n  duration_ms: 1\n  type: 'test'\n  ...\n" +
		"# Subtest: b\nok 2 - b\n  ---\n  duration_ms: 1\n  type: 'test'\n  ...\n1..2\n# tests 2\n# suites 0\n# pass " + summary +
		"\n# fail 0\n# cancelled 0\n# skipped 0\n# todo 0\n"
}

func TestCountFromTheOutput(t *testing.T) {
	r := rung{n: 1, name: "node", command: show("out.tap"), required: true, ratchet: "tap_pass", tst: "tap",
		capture: "\n      tap_pass: \"^# pass ([0-9]+)\""}
	for _, c := range []struct {
		name, summary, status, now, reason string
	}{
		{name: "the summary, not a test's line before it", summary: "2", status: "green", now: "2"},
		{name: "a summary that disagrees with the tests read", summary: "9", status: "red", now: "9",
			reason: "the ratchet tap_pass's two counts disagree: its capture read 9, and the tests read as tap name 2 passed"},
	} {
		t.Run(c.name, func(t *testing.T) {
			f := project(t, yamlOf("[1]", "\n  tap_pass: 2", "[]", r), map[string]string{"out.tap": tapTwo(c.summary),
				"work/tasks/T-0901-a.md": taskText("T-0901", "running", "")})
			f.climb("", Options{Task: "T-0901"})
			e := f.result("T-0901").Rungs[0]
			if e.Status != c.status || e.Ratchet == nil || string(e.Ratchet.Now) != c.now || showOf(e.Captures, "tap_pass") != c.now {
				t.Fatalf("%s, ratchet %+v, captures %s\n%s", e.Status, e.Ratchet, showOf(e.Captures, "tap_pass"), f.out.String())
			}
			if c.reason != "" && (e.Reason == nil || !strings.Contains(*e.Reason, c.reason)) {
				t.Errorf("reason %v, want %q", e.Reason, c.reason)
			}
		})
	}
}
