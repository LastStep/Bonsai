package main

// Helpers for the tests that read a word's human output (internal/textui draws it): it wraps to its width, so a
// sentence is asserted in flat text, and every refusal is one block whose next step has its own row.

import "strings"

// flat is text with its line breaks and runs of spaces made single spaces.
func flat(s string) string { return strings.Join(strings.Fields(s), " ") }

// hasNext reports whether the text holds a refusal block's next step.
func hasNext(s string) bool { return strings.Contains(s, "\n  next  ") }

// endsWith reports whether the text's last words are tail, on whole words, and its last row holds only the end of
// them: nothing, not even a row the wrapping would hide in flat text, follows. The text must end in a line break.
func endsWith(s, tail string) bool {
	tail = strings.TrimSpace(tail)
	rows := strings.Split(strings.TrimSuffix(s, "\n"), "\n")
	last := strings.TrimSpace(rows[len(rows)-1])
	return strings.HasSuffix(s, "\n") && last != "" && (strings.HasSuffix(last, tail) || strings.HasSuffix(tail, last)) && strings.HasSuffix(" "+flat(s), " "+tail)
}

// nextOf is a refusal block's next step in flat text: what follows its "next" label, to the end of the block.
func nextOf(s string) string {
	i := strings.LastIndex(s, "\n  next  ")
	if i < 0 {
		return ""
	}
	return flat(s[i+len("\n  next  "):])
}
