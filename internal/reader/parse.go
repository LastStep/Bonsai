package reader

// The format-1 grammar's structure (contract §2.4, "Structure" and "Keys"): mappings by indentation, block
// sequences deeper than their key with exactly one space after each -, block scalars, and every line read. The
// parser walks the lines top to bottom and never looks back, so the first problem it meets is the first in the
// file; each line it reaches is checked (check, in reader.go) before anything else is asked of it, and what a line
// itself holds is settled (settle, in reader.go) against its first bad character before the parser reads the lines
// under it.

import (
	"regexp"
	"strings"
)

type parser struct {
	lines []line
	i     int // the next line to read
}

// at skips blank and comment lines, checking each, and returns the next content line, checked, without consuming
// it; nil at the end.
func (p *parser) at() (*line, *Refusal) {
	for ; p.i < len(p.lines); p.i++ {
		l := &p.lines[p.i]
		if r := check(l); r != nil {
			return nil, r
		}
		// Comments (§2.4): a # at the start of a line (after its indentation) starts a comment line.
		if l.blank || strings.HasPrefix(l.body, "#") {
			if r := settle(l, nil); r != nil {
				return nil, r
			}
			continue
		}
		return l, nil
	}
	return nil, nil
}

// peek returns the next content line without checking or consuming anything (nil at the end). Only decisions use
// it; the line is checked when at reaches it, before it is read.
func (p *parser) peek() *line {
	for j := p.i; j < len(p.lines); j++ {
		if l := &p.lines[j]; !l.blank && !strings.HasPrefix(l.body, "#") {
			return l
		}
	}
	return nil
}

// document reads the whole file: one mapping at the indentation of its first content line, then nothing more.
func (p *parser) document() (*Map, *Refusal) {
	l, err := p.at()
	if err != nil {
		return nil, err
	}
	m := &Map{}
	if l == nil {
		return m, nil
	}
	if err := p.mapping(m, l.indent); err != nil {
		return nil, err
	}
	// Lines (§2.4): every line is read; a line the grammar did not consume is refused.
	if l, err := p.at(); err != nil {
		return nil, err
	} else if l != nil {
		return nil, settle(l, notRead(l))
	}
	return m, nil
}

func notRead(l *line) *Refusal {
	return refuseAt(CodeLineNotRead, l, l.indent, "the line %s belongs to no key at its indentation", show(l.text))
}

// posOf gives the byte in line l where s, a part of the line that runs to its end, starts.
func posOf(l *line, s string) int { return len(l.text) - len(s) }

// mapping reads key lines at exactly indentation n into m, until a line less indented.
func (p *parser) mapping(m *Map, n int) *Refusal {
	for {
		l, err := p.at()
		if err != nil {
			return err
		}
		if l == nil || l.indent < n {
			return nil
		}
		if l.indent > n {
			// Structure: nothing but a nested value is deeper than its mapping, and a nested value was read whole.
			return settle(l, notRead(l))
		}
		if err := p.keyValue(m, l, l.body, n); err != nil {
			return err
		}
	}
}

// keyValue reads one key: value from body, which starts at indentation n of line l (a key line, or the rest of a
// "- key: value" item), consumes the line and reads the value with anything nested under it.
func (p *parser) keyValue(m *Map, l *line, body string, n int) *Refusal {
	key, rest, err := splitKey(body, l)
	if err != nil {
		return settle(l, err)
	}
	// Keys (§2.4): never twice in one mapping.
	if _, ok := m.Get(key); ok {
		return settle(l, refuseAt(CodeKeyTwice, l, posOf(l, body), "the key %s is already in this mapping", show(key)))
	}
	p.i++ // the line is consumed; the value may read the lines after it
	v, err := p.value(rest, l, n)
	if err != nil {
		return err
	}
	m.add(key, v, l.no)
	return nil
}

var (
	keyPattern   = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	labelPattern = regexp.MustCompile(`^[a-z][a-z0-9-]*\.[a-z][a-z0-9_]*$`) // contract §5.1, the label name pattern
	reservedKeys = map[string]bool{"y": true, "n": true, "yes": true, "no": true, "on": true, "off": true,
		"true": true, "false": true, "null": true}
)

// splitKey reads the key at the start of body and returns it and the text after its colon. It refuses what §2.4
// refuses of a key, in Codes' order: complex, quoted, merge, form, reserved. A line that is no key: line at all (a
// sequence item where a key belongs, a value alone) is a line the grammar does not consume.
func splitKey(body string, l *line) (key, rest string, err *Refusal) {
	pos := posOf(l, body) // a key's problems are found at its first character: the key comes before its value
	// Keys: never complex (? key, then : value).
	if body == "?" || strings.HasPrefix(body, "? ") || strings.HasPrefix(body, "?\t") {
		return "", "", refuseAt(CodeKeyComplex, l, pos, "a complex key %s", show(body))
	}
	// Keys: never quoted. A quoted scalar followed by its colon is a quoted key; one with no colon is a value alone.
	if body[0] == '"' || body[0] == '\'' {
		if end, ok := quotedEnd(body); ok {
			after := strings.TrimLeft(body[end:], " ")
			if after == ":" || strings.HasPrefix(after, ": ") || strings.HasPrefix(after, ":\t") {
				return "", "", refuseAt(CodeKeyQuoted, l, pos, "a quoted key %s", show(body[:end]))
			}
		}
		return "", "", refuseAt(CodeLineNotRead, l, pos, "the line %s belongs to no key at its indentation", show(l.text))
	}
	if isSeqItem(body) {
		return "", "", refuseAt(CodeLineNotRead, l, pos, "a sequence item %s where a key belongs: a block sequence is indented deeper than its key", show(body))
	}
	sep := keySeparator(body)
	if sep < 0 {
		return "", "", refuseAt(CodeLineNotRead, l, pos, "the line %s belongs to no key at its indentation", show(l.text))
	}
	key, rest = body[:sep], body[sep+1:]
	switch {
	case key == "<<": // Keys: never <<.
		return "", "", refuseAt(CodeKeyMerge, l, pos, "the merge key <<")
	case !keyPattern.MatchString(key) && !labelPattern.MatchString(key):
		// Keys: [a-z][a-z0-9_]* for core and definition fields, <namespace>.<name> for label names.
		return "", "", refuseAt(CodeKeyForm, l, pos, "the key %s is neither [a-z][a-z0-9_]* nor a label name <namespace>.<name>", show(key))
	case len(key) > maxKey:
		// §2.4 fixes a key's form, not its length; YAML 1.1 and 1.2 libraries refuse an implicit key longer than
		// 1024 characters, so §2.4's own test refuses it, as a key of the wrong form (formats/README.md).
		return "", "", refuseAt(CodeKeyForm, l, pos, "the key is %d characters long, more than %d", len(key), maxKey)
	case reservedKeys[key]: // Keys: never y, n, yes, no, on, off, true, false, null.
		return "", "", refuseAt(CodeKeyReserved, l, pos, "the key %s is one a YAML 1.1 reader takes as a boolean or null", show(key))
	}
	return key, rest, nil
}

// maxKey is the longest key a YAML 1.1 and a YAML 1.2 library both read (PyYAML and npm yaml: 1024 characters).
const maxKey = 1024

// keySeparator returns the index of the colon that ends a plain key: the first : followed by a space, a tab or the
// end of the text, before any comment. A tab counts, as it does to format 0's reader and a YAML 1.2 library, so
// format: and a tab is the file's format key; the value after the tab is then refused (value).
func keySeparator(body string) int {
	for i := 0; i < len(body); i++ {
		switch {
		case body[i] == '#' && i > 0 && body[i-1] == ' ':
			return -1
		case body[i] == ':' && (i+1 == len(body) || body[i+1] == ' ' || body[i+1] == '\t'):
			return i
		}
	}
	return -1
}

// isSeqItem reports whether a line's body is a block sequence item: a - followed by whitespace or the line's end.
func isSeqItem(body string) bool {
	return body == "-" || strings.HasPrefix(body, "- ") || strings.HasPrefix(body, "-\t")
}

// value reads what follows a key's colon (rest) on line l, whose mapping sits at indentation n, and anything
// nested under it on the following lines.
func (p *parser) value(rest string, l *line, n int) (any, *Refusal) {
	s := strings.TrimLeft(rest, " ")
	if strings.HasPrefix(s, "\t") {
		// A tab before the value: a YAML 1.1 library cannot read it, a YAML 1.2 one can (§2.4's own test). The
		// value then starts with a tab, which fits no row of the plain-scalar table.
		r := refuseAt(CodeQuoteThisValue, l, posOf(l, s), "a tab between the key's colon and its value")
		r.Next = "separate the value from the colon with spaces, never a tab"
		return nil, settle(l, r)
	}
	return p.valueOrNested(s, l, n)
}

// valueOrNested reads a value s that starts on line l (leading spaces gone), whose parent sits at indentation n:
// nothing or a comment (the value is nested on the lines below, or null), a block scalar, or a value on the line.
// What the line holds is settled before any line below it is read.
func (p *parser) valueOrNested(s string, l *line, n int) (any, *Refusal) {
	// Comments: a # after a space starts a comment, so a key with only a comment after it has no inline value.
	if s == "" || s[0] == '#' {
		if r := settle(l, nil); r != nil {
			return nil, r
		}
		return p.nested(n)
	}
	if s[0] == '|' || s[0] == '>' {
		h, r := blockHeader(s, l)
		if r = settle(l, r); r != nil {
			return nil, r
		}
		return p.blockScalar(h, l, n)
	}
	v, r := inline(s, l)
	if r = settle(l, r); r != nil {
		return nil, r
	}
	return v, nil
}

// nested reads the value of a key (or item) with nothing after it on its own line: a mapping or a block sequence
// on the following lines, indented deeper than n, or null.
func (p *parser) nested(n int) (any, *Refusal) {
	l := p.peek()
	if l == nil || l.indent <= n || strings.HasPrefix(l.body, "\t") {
		// Structure: a block sequence at its key's own indentation is not its value; the mapping that follows
		// meets it as a line it does not consume (line-not-read).
		return nil, nil
	}
	if isSeqItem(l.body) {
		return p.sequence(l.indent)
	}
	m := &Map{}
	if err := p.mapping(m, l.indent); err != nil {
		return nil, err
	}
	return m, nil
}

// sequence reads block sequence items at exactly indentation m.
func (p *parser) sequence(m int) ([]any, *Refusal) {
	out := []any{}
	for {
		l, err := p.at()
		if err != nil {
			return nil, err
		}
		if l == nil || l.indent < m {
			return out, nil
		}
		if l.indent > m {
			return nil, settle(l, notRead(l))
		}
		if !isSeqItem(l.body) {
			return out, nil // the mapping around the sequence meets this line
		}
		// Structure: exactly one space after -.
		if l.body == "-" || l.body[1] != ' ' || len(l.body) > 2 && (l.body[2] == ' ' || l.body[2] == '\t') {
			return nil, settle(l, refuseAt(CodeSeqDashSpace, l, l.indent, "the item's - is not followed by exactly one space"))
		}
		v, err := p.item(l.body[2:], l, m)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
}

// item reads one sequence item: content is the text after "- " on line l, the sequence at indentation m. A
// "key: value" item opens a mapping whose keys sit at m+2; anything else is a value.
func (p *parser) item(content string, l *line, m int) (any, *Refusal) {
	if !itemIsKey(content) {
		p.i++
		return p.valueOrNested(content, l, m)
	}
	mm := &Map{}
	if err := p.keyValue(mm, l, content, m+2); err != nil {
		return nil, err
	}
	if err := p.mapping(mm, m+2); err != nil {
		return nil, err
	}
	return mm, nil
}

// itemIsKey reports whether an item's content is a key: value (or a key form §2.4 refuses) rather than a value.
func itemIsKey(c string) bool {
	if c == "" {
		return false
	}
	switch c[0] {
	case '?':
		return c == "?" || strings.HasPrefix(c, "? ") || strings.HasPrefix(c, "?\t")
	case '"', '\'':
		end, ok := quotedEnd(c)
		if !ok {
			return false
		}
		after := strings.TrimLeft(c[end:], " ")
		return after == ":" || strings.HasPrefix(after, ": ") || strings.HasPrefix(after, ":\t")
	case '[', '{', '|', '>', '#':
		return false
	}
	return keySeparator(c) >= 0
}

// blockHeader reads a block scalar's header, s from its indicator to the line's end. Structure (§2.4): block
// scalars |, |-, > and >- only; a comment may follow the header. A wrong header is found at its first character.
func blockHeader(s string, l *line) (string, *Refusal) {
	h := s
	if i := strings.Index(h, " #"); i >= 0 {
		h = h[:i]
	}
	h = strings.TrimRight(h, " ")
	if h != "|" && h != "|-" && h != ">" && h != ">-" {
		return "", refuseAt(CodeBlockIndicator, l, posOf(l, s), "the block scalar header %s", show(h))
	}
	return h, nil
}

// blockScalar reads a block scalar: h is its header (blockHeader) on line hl, whose parent sits at indentation n;
// its content is the following lines indented deeper than n.
func (p *parser) blockScalar(h string, hl *line, n int) (string, *Refusal) {
	folded, strip := h[0] == '>', strings.HasSuffix(h, "-")

	type bline struct {
		text  string
		empty bool
		eol   bool
	}
	var lines []bline
	ind := -1           // the block's indentation, from its first content line
	var leading []*line // empty lines before the first content line
	for p.i < len(p.lines) {
		l := &p.lines[p.i]
		if err := check(l); err != nil {
			return "", err
		}
		if l.blank {
			if ind < 0 {
				leading = append(leading, l)
			} else if len(l.text) > ind {
				// Both libraries keep the spaces past the indentation, format 0 drops them, and a reader cannot see
				// them (formats/README.md): refused.
				return "", refuseAt(CodeLineNotRead, l, ind, "a line of only spaces, deeper than the block scalar's indentation")
			}
			lines = append(lines, bline{empty: true, eol: l.eol})
			p.i++
			continue
		}
		if ind < 0 {
			if l.indent <= n {
				break // no content: the block scalar is empty
			}
			ind = l.indent
			for _, e := range leading {
				if len(e.text) > ind {
					// YAML 1.2 refuses this; a YAML 1.1 library reads another indentation.
					return "", refuseAt(CodeLineNotRead, e, ind, "a line of only spaces before the block scalar's first line, deeper than it")
				}
			}
		} else if l.indent < ind {
			break
		}
		// Structure: no line inside a block scalar starts with #.
		if strings.HasPrefix(l.body, "#") {
			return "", settle(l, refuseAt(CodeBlockHashLine, l, l.indent, "a line inside a block scalar starts with #"))
		}
		// Structure: in > no line is indented deeper than the first.
		if folded && l.indent > ind {
			return "", settle(l, refuseAt(CodeFoldedDeeper, l, l.indent, "a line deeper than the first line of a > block scalar"))
		}
		if r := settle(l, nil); r != nil {
			return "", r
		}
		lines = append(lines, bline{text: l.text[ind:], eol: l.eol})
		p.i++
	}
	// Trailing empty lines are not part of the value (clip and strip alike).
	last := len(lines) - 1
	for last >= 0 && lines[last].empty {
		last--
	}
	if last < 0 {
		return "", nil
	}
	if !strip && !lines[last].eol {
		// Lines (§2.4): LF or CRLF. A | or > block whose last line ends the file with no line ending reads as "a"
		// to a YAML 1.1 library and "a\n" to a YAML 1.2 one, so §2.4's own test refuses it.
		last := &p.lines[p.i-1]
		r := refuseAt(CodeNotText, last, len(last.text), "the file ends inside a | or > block scalar with no line ending")
		r.Next = "end the file with a line ending, or use |- or >-"
		return "", r
	}
	var b strings.Builder
	sawText := false
	empties := 0
	for _, bl := range lines[:last+1] {
		if bl.empty {
			empties++
			if !folded || !sawText {
				b.WriteByte('\n')
			}
			continue
		}
		if sawText {
			switch {
			case !folded:
				b.WriteByte('\n')
			case empties == 0:
				b.WriteByte(' ')
			default:
				b.WriteString(strings.Repeat("\n", empties))
			}
		}
		empties = 0
		sawText = true
		b.WriteString(bl.text)
	}
	if !strip && lines[last].eol {
		b.WriteByte('\n')
	}
	return b.String(), nil
}
