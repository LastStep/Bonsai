package reader

// The format-1 grammar's values on one line (contract §2.4, "Structure", "Comments", "Quoted scalars", "Plain
// scalars"): plain and quoted scalars, one-line flow sequences, [] and {}, and the refusals of anchors, aliases,
// tags and flow mappings.

import (
	"regexp"
	"strconv"
	"strings"
)

// inline reads a value that sits whole on its line: s is the text after the key's colon (or the item's "- "),
// leading spaces gone, and not empty.
func inline(s string, l *line) (any, *Refusal) {
	switch s[0] {
	case '&': // Structure: anchors refused.
		return nil, refuse(CodeAnchor, l.no, "an anchor %s", show(s))
	case '*': // Structure: aliases refused.
		return nil, refuse(CodeAlias, l.no, "an alias %s", show(s))
	case '!': // Structure: tags refused.
		return nil, refuse(CodeTag, l.no, "a tag %s", show(s))
	case '[':
		return flowSequence(s, l)
	case '{':
		return emptyMap(s, l)
	case '"', '\'':
		v, end, err := quoted(s, l.no)
		if err != nil {
			return nil, err
		}
		return v, afterQuote(s[0], s[end:], l)
	}
	return plain(plainText(s), false, l)
}

// plainText cuts a plain scalar at its comment and trims its trailing spaces. Comments (§2.4): a comment starts at
// a # after a space; a quote character inside a plain value is just a character and opens nothing.
func plainText(s string) string {
	if i := strings.Index(s, " #"); i >= 0 {
		s = s[:i]
	}
	return strings.TrimRight(s, " ")
}

// afterQuote checks what follows a quoted scalar's closing quote on its line (rest). Quoted scalars (§2.4): nothing
// but a comment may follow the closing quote, and an unescaped " inside double quotes is refused: the first
// unescaped " closes the scalar, so another " before any comment means one was not escaped.
func afterQuote(q byte, rest string, l *line) *Refusal {
	before := rest
	if i := strings.Index(rest, " #"); i >= 0 {
		before = rest[:i]
	}
	if q == '"' && strings.Contains(before, `"`) {
		return refuse(CodeUnescapedQuote, l.no, "an unescaped \" inside double quotes")
	}
	if strings.Trim(before, " ") != "" {
		return refuse(CodeAfterQuote, l.no, "%s after the closing quote", show(strings.Trim(before, " ")))
	}
	return nil
}

// quotedEnd returns the index just after the quoted scalar that starts s, and whether it closes on this line.
func quotedEnd(s string) (int, bool) {
	q := s[0]
	for i := 1; i < len(s); i++ {
		switch {
		case q == '"' && s[i] == '\\':
			i++ // the escaped character never closes the scalar
		case s[i] == q && q == '\'' && i+1 < len(s) && s[i+1] == '\'':
			i++ // '' is one ' inside single quotes
		case s[i] == q:
			return i + 1, true
		}
	}
	return 0, false
}

// quoted reads the quoted scalar that starts s: its text and the index just after its closing quote. Quoted
// scalars (§2.4) are always text, on one line; in double quotes only five escapes exist (\", \\, \n, \r, \t); in
// single quotes two single quotes stand for one (two libraries agree; formats/expect.json settles it).
func quoted(s string, lineNo int) (string, int, *Refusal) {
	q := s[0]
	var b strings.Builder
	for i := 1; i < len(s); i++ {
		c := s[i]
		switch {
		case q == '"' && c == '\\':
			if i+1 == len(s) {
				// A backslash at the line's end continues a double-quoted scalar on the next line in YAML.
				return "", 0, refuse(CodeQuotedMulti, lineNo, "a quoted value that does not close on its line")
			}
			switch s[i+1] {
			case '"':
				b.WriteByte('"')
			case '\\':
				b.WriteByte('\\')
			case 'n':
				b.WriteByte('\n')
			case 'r':
				b.WriteByte('\r')
			case 't':
				b.WriteByte('\t')
			default:
				return "", 0, refuse(CodeBadEscape, lineNo, "the escape %s: double quotes have only \\\", \\\\, \\n, \\r and \\t", show(s[i:i+2]))
			}
			i++
		case c == q && q == '\'' && i+1 < len(s) && s[i+1] == '\'':
			b.WriteByte('\'')
			i++
		case c == q:
			return b.String(), i + 1, nil
		default:
			b.WriteByte(c)
		}
	}
	return "", 0, refuse(CodeQuotedMulti, lineNo, "a quoted value that does not close on its line")
}

// emptyMap reads a value that starts with {: {} (spaces allowed inside) is an empty map; a flow mapping with
// content is refused (Structure, §2.4).
func emptyMap(s string, l *line) (any, *Refusal) {
	j := 1
	for j < len(s) && s[j] == ' ' {
		j++
	}
	if j == len(s) || s[j] != '}' {
		return nil, refuse(CodeFlowMapping, l.no, "a flow mapping with content %s", show(plainText(s)))
	}
	if rest := plainText(s[j+1:]); rest != "" {
		return nil, refuse(CodeLineNotRead, l.no, "%s after {}", show(rest))
	}
	return &Map{}, nil
}

// flowSequence reads a one-line flow sequence of scalars, s starting at its [. As formats/README.md says, it runs
// from its [ to the last ] on its line (before any comment), and its items are split at commas outside quotes. A
// quote opens a quoted item only at the item's start: elsewhere it is a plain character.
func flowSequence(s string, l *line) (any, *Refusal) {
	var commas, closes []int
	end := len(s) // where a comment starts, or the line's end
	itemStart := true
scan:
	for i := 1; i < len(s); i++ {
		c := s[i]
		if itemStart {
			if c == ' ' {
				continue
			}
			itemStart = false
			if c == '"' || c == '\'' {
				e, ok := quotedEnd(s[i:])
				if !ok {
					// The quoted item, so the sequence, does not close on this line; of the two codes that fit,
					// flow-multiline is listed first.
					return nil, refuse(CodeFlowMultiline, l.no, "a flow sequence that does not close on its line")
				}
				i += e - 1
				continue
			}
		}
		switch {
		case c == ',':
			commas = append(commas, i)
			itemStart = true
		case c == ']':
			closes = append(closes, i)
		case c == '#' && s[i-1] == ' ':
			end = i
			break scan
		}
	}
	// Structure: one-line flow sequences only.
	if len(closes) == 0 {
		return nil, refuse(CodeFlowMultiline, l.no, "a flow sequence that does not close on its line")
	}
	last := closes[len(closes)-1]
	if rest := strings.TrimRight(s[last+1:end], " "); rest != "" {
		return nil, refuse(CodeLineNotRead, l.no, "%s after the flow sequence's closing ]", show(rest))
	}
	inner := s[1:last]
	out := []any{}
	if strings.Trim(inner, " ") == "" {
		return out, nil // [] and [ ]
	}
	from := 1
	for _, c := range append(commas, last) {
		if c > last {
			break
		}
		item := strings.Trim(s[from:c], " ")
		from = c + 1
		if item == "" {
			return nil, refuse(CodeLineNotRead, l.no, "an empty item in the flow sequence %s", show(s[:last+1]))
		}
		v, err := flowItem(item, l)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, nil
}

// flowItem reads one flow sequence item: a quoted or plain scalar. Structure (§2.4): flow sequences of scalars
// only, never nested; a [ or { that opens inside one is refused. A flow mapping with content is refused as such,
// the code listed first of the two that fit.
func flowItem(t string, l *line) (any, *Refusal) {
	switch t[0] {
	case '&':
		return nil, refuse(CodeAnchor, l.no, "an anchor %s", show(t))
	case '*':
		return nil, refuse(CodeAlias, l.no, "an alias %s", show(t))
	case '!':
		return nil, refuse(CodeTag, l.no, "a tag %s", show(t))
	case '{':
		if strings.Trim(t[1:], " ") != "}" {
			return nil, refuse(CodeFlowMapping, l.no, "a flow mapping with content %s", show(t))
		}
		return nil, refuse(CodeFlowNested, l.no, "a {} inside a flow sequence")
	case '"', '\'':
		v, end, err := quoted(t, l.no)
		if err != nil {
			return nil, err
		}
		return v, afterQuote(t[0], t[end:], l)
	}
	for i := 0; i < len(t); i++ {
		switch t[i] {
		case '[', '{':
			return nil, refuse(CodeFlowNested, l.no, "a %s that opens inside a flow sequence", show(t[i:i+1]))
		case ']', '}':
			return nil, refuse(CodeQuoteThisValue, l.no, "the flow sequence item %s holds %s", show(t), show(t[i:i+1]))
		}
	}
	return plain(t, true, l)
}

var (
	intPattern      = regexp.MustCompile(`^-?(0|[1-9][0-9]{0,14})$`)
	decimalPattern  = regexp.MustCompile(`^-?(0|[1-9][0-9]*)\.[0-9]+$`)
	datePattern     = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}$`)
	dateTimePattern = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2} [0-9]{2}:[0-9]{2}$`)
	// The words a plain text value may not be in any case: null, true and false in any case but lower (lower case
	// reads as null or a boolean), and y, n, yes, no, on, off in every case.
	plainWords = map[string]bool{"null": true, "true": true, "false": true,
		"y": true, "n": true, "yes": true, "no": true, "on": true, "off": true}
)

// plain reads a plain scalar, s cut at its comment and trimmed, by §2.4's table, row by row. inFlow is set inside a
// flow sequence.
func plain(s string, inFlow bool, l *line) (any, *Refusal) {
	switch {
	case s == "" || s == "null" || s == "~": // empty, null, ~: null
		return nil, nil
	case s == "true": // true, false (lower case only): boolean
		return true, nil
	case s == "false":
		return false, nil
	case intPattern.MatchString(s): // -?(0|[1-9][0-9]{0,14}): an integer of at most 15 digits
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return nil, refuse(CodeQuoteThisValue, l.no, "%s does not read as an integer", show(s))
		}
		return n, nil
	case decimalPattern.MatchString(s): // -?(0|[1-9][0-9]*)\.[0-9]+: a decimal
		return Decimal(s), nil
	case datePattern.MatchString(s), dateTimePattern.MatchString(s): // YYYY-MM-DD, YYYY-MM-DD HH:MM: text
		return s, nil
	}
	if why := notText(s, inFlow); why != "" {
		return nil, refuse(CodeQuoteThisValue, l.no, "the plain value %s %s", show(s), why)
	}
	return s, nil
}

// notText says why a plain scalar fits no row of §2.4's text row, or returns "" when it is text: it starts with
// an ASCII letter or _; holds no ": " and no " #"; does not end in ':'; is not one of the words above in another
// case, nor y, n, yes, no, on, off; inside a flow sequence holds no , [ ] { }. A tab is refused too: a YAML 1.1
// library cannot read one in a plain scalar, so the two libraries §2.4 names would disagree.
func notText(s string, inFlow bool) string {
	c := s[0]
	switch {
	case !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c == '_'):
		return "does not start with a letter or _"
	case strings.Contains(s, ": "):
		return "holds \": \""
	case strings.Contains(s, " #"):
		return "holds \" #\""
	case strings.HasSuffix(s, ":"):
		return "ends in ':'"
	case plainWords[strings.ToLower(s)]:
		return "is a word a YAML 1.1 reader takes as a boolean or null"
	case strings.Contains(s, "\t"):
		return "holds a tab"
	case inFlow && strings.ContainsAny(s, ",[]{}"):
		return "holds one of , [ ] { } inside a flow sequence"
	}
	return ""
}
