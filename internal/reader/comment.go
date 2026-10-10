package reader

// Comments on key lines, as the format-1 reader sees them (contract §2.4, "Comments": a comment starts at # at the
// start of a line or after a space; a quote character inside a plain value is just a character). The reader keeps no
// comment in what it returns; bonsai check --pack (plan-5 5.1.9) asks this of a pack's YAML files, which give every key
// a # comment "at the end of its line or on the line just above" (spec §5). A comment line, one whose first character
// after its indentation is #, needs nothing here: inside a block scalar such a line is refused (§2.4), so in a file
// the reader read whole every such line is a comment.

import "strings"

// KeyLineComment reports whether a key line of a file the format-1 reader read whole carries a comment after its key
// and value: text is the line ("key: value", or a "- key: value" item line), without its line ending. What counts is
// what the reader skips as a comment where the value ends:
//   - after a plain value, {} or a block scalar's header (| or >): a # after a space;
//   - after a quoted value: a # after its closing quote (nothing else may follow one);
//   - after a flow sequence: a # after a space outside its quoted items;
//   - a key with nothing after its colon but a # (its value nested on the lines below, or null).
//
// A line that is no key line gives false.
func KeyLineComment(text string) bool {
	s := strings.TrimLeft(strings.TrimSuffix(text, "\r"), " ")
	for isSeqItem(s) {
		s = strings.TrimLeft(s[1:], " ")
	}
	sep := keySeparator(s)
	if sep < 0 {
		return false
	}
	return valueComment(strings.TrimLeft(s[sep+1:], " "))
}

// valueComment reports whether a value's text, from its first character to the line's end, ends in a comment.
func valueComment(s string) bool {
	if s == "" {
		return false
	}
	switch s[0] {
	case '#':
		return true
	case '"', '\'':
		end, ok := quotedEnd(s)
		return ok && strings.Contains(s[end:], "#")
	case '[':
		itemStart := true
		for i := 1; i < len(s); i++ {
			c := s[i]
			if itemStart {
				if c == ' ' {
					continue
				}
				itemStart = false
				if c == '"' || c == '\'' {
					end, ok := quotedEnd(s[i:])
					if !ok {
						return false
					}
					i += end - 1
					continue
				}
			}
			switch {
			case c == ',':
				itemStart = true
			case c == '#' && s[i-1] == ' ':
				return true
			}
		}
		return false
	}
	return strings.Contains(s, " #")
}
