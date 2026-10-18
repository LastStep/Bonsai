package main

// The help of a word whose sub-words are all not built yet (step 5.6.0's open items, built in 5.6.1): its --help
// lists them under "Not built" alone, with no empty "Answers" heading, and `<word> <sub> --help` shows the sub-word's
// own help (exit 0) where any other call of it is refused as not-built. A word of the test's own, registered for the
// test and taken out after it, so the case holds whichever real words are built.

import (
	"bytes"
	"strings"
	"testing"
)

func TestHelpOfLaterSubWords(t *testing.T) {
	sub := &Word{Name: "later-test show", Title: "show something (a test word).", Summary: "show something",
		Args: "[flags]", About: "It shows something.\n", Exits: []Exit{{Code: 0, Means: "shown"}, {Code: 2, Means: "not built"}},
		Examples: []string{"bonsai later-test show"}, Later: "step 9.9", Run: func(c *call) int { return exitOK }}
	w := &Word{Name: "later-test", Order: 999, Title: "a word whose sub-words are all not built (a test word).",
		Summary: "a test word", Args: "<sub-word>", About: "A test word.\n", Exits: []Exit{{Code: 0, Means: "shown"}, {Code: 2, Means: "not built"}},
		Examples: []string{"bonsai later-test --help"}, Subs: []*Word{sub}}
	register(w)
	t.Cleanup(func() { delete(words, w.Name) })

	help := w.Help()
	if strings.Contains(help, "Answers") || !strings.Contains(flat(help), "not built yet: show (step 9.9).") {
		t.Errorf("the word's help:\n%s", help)
	}

	var stdout, stderr bytes.Buffer
	if code := run([]string{"later-test", "show", "--help"}, &stdout, &stderr); code != exitOK ||
		!strings.Contains(stdout.String(), "bonsai later-test show") || !strings.Contains(stdout.String(), "show something") {
		t.Errorf("later-test show --help: exit %d\n%s%s", code, stdout.String(), stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	if code := run([]string{"later-test", "show"}, &stdout, &stderr); code != exitInput ||
		!strings.Contains(stdout.String()+stderr.String(), "not-built") {
		t.Errorf("later-test show: exit %d\n%s%s", code, stdout.String(), stderr.String())
	}
}
