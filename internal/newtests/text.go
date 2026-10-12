package newtests

// The check's words: one line for the Bless ask's why (Summary), and the human output's lines (Lines): the counts,
// then the names, at most MaxNames a group. ASCII; names that are not printable ASCII are quoted with Go's escapes.

import (
	"strconv"
	"strings"
)

// notReported says why a new test the second base run did not report (or reported skipped) is "check not run".
const notReported = "the base run did not report them, or skipped them: their package did not build or their file " +
	"did not load there"

// Summary is the check in one line, for the Bless ask's why and the head of the human lines:
//
//	3 new tests, run on the base 3b4c5d6e7f80: 1 failed there, 1 proves nothing new, 1 check not run
//	3 new tests: check not run: an old test failed at the base, ...: example.com/demo/calc.TestOld
//	3 new tests (the count's rise; they could not be named): check not run: no base
//	no new tests: every test the climb passed ran on the base 3b4c5d6e7f80 too
func (r Result) Summary() string {
	nt := r.NewTests
	switch {
	case !r.Named:
		return tests(nt.Count) + " (the count's rise; they could not be named): check not run: " + r.Why
	case nt.Count == 0:
		return "no new tests: every test the climb passed ran on the base " + short(nt.Base) + " too"
	case r.Why != "":
		return tests(nt.Count) + ": check not run: " + r.Why
	}
	return tests(nt.Count) + ", run on the base " + short(nt.Base) + ": " + strconv.FormatInt(nt.FailedOnBase, 10) +
		" failed there, " + strconv.FormatInt(nt.PassedOnBase, 10) + " " + plural(nt.PassedOnBase, "proves", "prove") +
		" nothing new, " + strconv.FormatInt(nt.NotRun, 10) + " check not run"
}

// Lines are the human output: the summary, then each group's names (at most MaxNames, then how many more), then the
// notes.
func (r Result) Lines() []string {
	lines := []string{r.Summary()}
	nt := r.NewTests
	group := func(head string, names []string, n int64) {
		if n == 0 || len(names) == 0 {
			return
		}
		shownNames := make([]string, len(names))
		for i, x := range names {
			shownNames[i] = shown(x)
		}
		line := "  " + head + ": " + strings.Join(shownNames, ", ")
		if more := n - int64(len(names)); more > 0 {
			line += ", and " + strconv.FormatInt(more, 10) + " more"
		}
		lines = append(lines, line)
	}
	group("failed on the base", r.Failed, nt.FailedOnBase)
	group("proves nothing new (passed on the base)", r.Passed, nt.PassedOnBase)
	notRun := "check not run"
	if r.Why == "" {
		notRun += " (" + notReported + ")"
	}
	group(notRun, r.NotRun, nt.NotRun)
	for _, n := range r.Notes {
		lines = append(lines, "  note: "+n)
	}
	return lines
}

// tests is "1 new test" or "n new tests".
func tests(n int64) string {
	return strconv.FormatInt(n, 10) + " new " + plural(n, "test", "tests")
}

func plural(n int64, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}
