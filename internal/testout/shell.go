package testout

import "strings"

// quoteLine joins words into one command line for the rung's shell, each quoted when it needs it (quoteWord).
func quoteLine(words []string, windows bool) string {
	out := make([]string, len(words))
	for i, w := range words {
		out[i] = quoteWord(w, windows)
	}
	return strings.Join(out, " ")
}

// quoteWord quotes one word for /bin/sh, in single quotes, or for cmd, in double quotes with a " inside written \"
// and the backslashes before it doubled, as a Windows program splits its command line; a word needing none is bare.
// A single quote inside a single-quoted word is closed, escaped and reopened:
//
//	it's  ->  'it'\''s'
func quoteWord(w string, windows bool) string {
	if w != "" && !needsQuoting(w) {
		return w
	}
	if !windows {
		return "'" + strings.ReplaceAll(w, "'", `'\''`) + "'"
	}
	var b strings.Builder
	b.WriteByte('"')
	slashes := 0
	for i := 0; i < len(w); i++ {
		switch c := w[i]; c {
		case '\\':
			slashes++
		case '"':
			b.WriteString(strings.Repeat(`\`, 2*slashes+1))
			b.WriteByte('"')
			slashes = 0
		default:
			b.WriteString(strings.Repeat(`\`, slashes))
			b.WriteByte(c)
			slashes = 0
		}
	}
	b.WriteString(strings.Repeat(`\`, 2*slashes))
	b.WriteByte('"')
	return b.String()
}

// needsQuoting is true when a word holds anything but ASCII letters, digits and -_./:=,+@.
func needsQuoting(w string) bool {
	for i := 0; i < len(w); i++ {
		c := w[i]
		switch {
		case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9':
		case strings.IndexByte("-_./:=,+@", c) >= 0:
		default:
			return true
		}
	}
	return false
}

// shellSegments splits a command line into its commands' words, as the rung's shell would group them: a segment ends
// at &&, ||, ;, a pipe, a lone & or a line's end; words split on spaces and tabs. For /bin/sh, single quotes keep
// their text as written, double quotes keep it but for a backslash before " \ $ or `, and a backslash outside quotes
// keeps the next character. For cmd, double quotes keep their text (a \" in them is a "), and a ^ outside them keeps
// the next character. The quotes themselves are dropped.
func shellSegments(s string, windows bool) [][]string {
	var segs [][]string
	var words []string
	var cur strings.Builder
	inWord := false
	endWord := func() {
		if inWord {
			words = append(words, cur.String())
			cur.Reset()
			inWord = false
		}
	}
	endSegment := func() {
		endWord()
		if len(words) > 0 {
			segs = append(segs, words)
			words = nil
		}
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c == ' ' || c == '\t':
			endWord()
		case c == '\n' || c == '\r' || c == ';' || c == '&' || c == '|':
			endSegment()
		case c == '\'' && !windows:
			inWord = true
			j := strings.IndexByte(s[i+1:], '\'')
			if j < 0 {
				j = len(s) - i - 1
			}
			cur.WriteString(s[i+1 : i+1+j])
			i += j + 1
		case c == '"':
			inWord = true
			i++
			for ; i < len(s) && s[i] != '"'; i++ {
				if s[i] == '\\' && i+1 < len(s) {
					next := s[i+1]
					if (windows && next == '"') || (!windows && strings.IndexByte("\"\\$`", next) >= 0) {
						i++
					}
				}
				cur.WriteByte(s[i])
			}
		case (c == '\\' && !windows) || (c == '^' && windows):
			inWord = true
			if i+1 < len(s) {
				i++
				cur.WriteByte(s[i])
			}
		default:
			inWord = true
			cur.WriteByte(c)
		}
	}
	endSegment()
	return segs
}
