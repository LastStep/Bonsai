package reader

// Format 0's reader (contract §2.3, §2.4): a hand port of the studio's yaml.mjs at commit 4a05eac (parseYaml and
// parseFrontmatter), the contract's own definition of format 0. "Format 0 reads exactly as [yaml.mjs] reads today,
// forever, with no new refusals": every leniency of that reader is kept (a plain value holding ": ", unquoted hashes,
// the last of two duplicate keys, document markers skipped, a list at its key's indent ending the mapping, a # line
// inside a block scalar read as an empty line), and it refuses only where yaml.mjs throws. Node's own stack overflow
// (a RangeError, at roughly 1,700 to 2,100 levels of nesting) is not one of yaml.mjs's throws, so the port reads such a
// file: refusing it would be a new refusal.
//
// The functions below follow yaml.mjs's one for one, under its names (parser0's methods, or with a 0 added); the
// comments say where JavaScript itself decides something a Go reader would otherwise do differently.
//
// What JavaScript decides, and how this port matches it:
//   - The text. The studio reads a file with readFileSync(path, 'utf8'): bytes that are not UTF-8 become U+FFFD as
//     the WHATWG decoder makes them (decode0), and a BOM is kept (parseYaml then cannot read a first line that starts
//     with one; parseFrontmatter drops one).
//   - Whitespace. JavaScript's \s, trim() and trimEnd() mean the same set (isSpace0), which is wider than ASCII:
//     U+00A0, U+FEFF, U+2028 and others. A regular expression's . matches anything but LF, CR, U+2028 and U+2029.
//   - Numbers. A JavaScript number is a float64: an integer string is read as parseInt does (the nearest float64,
//     Infinity when too large) and a decimal as parseFloat does, so 0755 is 755, 1.0 is 1 and -0 is -0.
//   - Objects. A mapping is a JavaScript object: keys that are array indices (0 to 4294967294, written without
//     leading zeros) come first in ascending order, then the others in the order they were first set; a key set twice
//     keeps its first place and its last value; __proto__ follows Object.prototype's setter (obj0.set).
//   - Strings are UTF-16 in JavaScript. Only one place in yaml.mjs counts UTF-16 units where it matters: a block
//     scalar line less indented than the block's first line is cut at the first line's indent (sliceUnits0).

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Result0 is a format-0 read's outcome: Accepted with its value, or Refused (yaml.mjs threw). A read under format 0
// never returns Format0.
type Result0 struct {
	Outcome Outcome
	// Value, when Accepted, is what yaml.mjs returns: a *Map, or []any when the file's top level is a list. Inside it
	// a mapping is a *Map (keys in JavaScript's order, each Entry's Line the line that set its value), a sequence
	// []any, text string, a number float64 (parseInt's or parseFloat's, so possibly -0 or ±Inf), true and false bool,
	// null nil. A string holds valid UTF-8 except in one case JavaScript allows and Go cannot spell: a lone UTF-16
	// surrogate, which only sliceUnits0 can make, kept as its three WTF-8 bytes.
	Value   any
	Refusal *Refusal // when Refused: Code is empty (format 0 has no reason codes), Line counts in the whole file
}

// Map returns the value when it is a mapping.
func (r Result0) Map() (*Map, bool) {
	m, ok := r.Value.(*Map)
	return m, ok && r.Outcome == Accepted
}

// Err returns the refusal as an error, or nil when the file was not refused.
func (r Result0) Err() error {
	if r.Outcome == Refused {
		return r.Refusal
	}
	return nil
}

// ReadFormat0YAML reads a YAML file's bytes under format 0, as yaml.mjs's parseYaml reads the file's text. A caller
// reaches it when ReadYAML's dispatch returns Format0 and the file is one it reads under format 0.
func ReadFormat0YAML(raw []byte) Result0 {
	return read0(decode0(raw), 0)
}

// ReadFormat0Markdown reads a markdown file's frontmatter under format 0, as yaml.mjs's parseFrontmatter(text).data:
// one BOM dropped, then the text between a first line of --- and the next line of ---, each line ending in LF or
// CRLF (frontmatter0). A file with no such block reads as an empty mapping, as it does in yaml.mjs. A refusal's line
// counts in the whole file, the opening --- being line 1.
func ReadFormat0Markdown(raw []byte) Result0 {
	s := strings.TrimPrefix(decode0(raw), "\uFEFF")
	content, ok := frontmatter0(s)
	if !ok {
		return Result0{Outcome: Accepted, Value: &Map{}}
	}
	return read0(content, 1)
}

// read0 runs parseYaml over text; offset is the number of file lines before the text's first.
func read0(text string, offset int) (res Result0) {
	p := &parser0{offset: offset}
	defer func() {
		if x := recover(); x != nil {
			r, ok := x.(*Refusal)
			if !ok {
				panic(x)
			}
			res = Result0{Outcome: Refused, Refusal: r}
		}
	}()
	return Result0{Outcome: Accepted, Value: finish0(p.parseYaml(text))}
}

// frontmatter0 is parseFrontmatter's /^---\r?\n([\s\S]*?)\r?\n---(?:\r?\n|$)/: the opening --- and its line end,
// then the shortest text followed by a line end, ---, and a line end or the end of the file.
func frontmatter0(s string) (string, bool) {
	var o int
	switch {
	case strings.HasPrefix(s, "---\r\n"):
		o = 5
	case strings.HasPrefix(s, "---\n"):
		o = 4
	default:
		return "", false
	}
	for p := o; p < len(s); p++ {
		q := p
		if s[q] == '\r' {
			q++
		}
		if q >= len(s) || s[q] != '\n' || !strings.HasPrefix(s[q+1:], "---") {
			continue
		}
		rest := s[q+4:]
		if rest == "" || strings.HasPrefix(rest, "\n") || strings.HasPrefix(rest, "\r\n") {
			return s[o:p], true
		}
	}
	return "", false
}

// decode0 turns bytes into text as Node's readFileSync(path, 'utf8') does, by the WHATWG UTF-8 decoder: valid UTF-8 as
// it is; one U+FFFD for each byte that cannot start a character, and one for each character cut short (its bytes so
// far), the byte that cut it then read again.
func decode0(raw []byte) string {
	if utf8.Valid(raw) {
		return string(raw)
	}
	var b strings.Builder
	needed, seen := 0, 0
	var cp rune
	lower, upper := byte(0x80), byte(0xBF)
	for i := 0; i < len(raw); {
		c := raw[i]
		if needed == 0 {
			switch {
			case c <= 0x7F:
				b.WriteByte(c)
			case c >= 0xC2 && c <= 0xDF:
				needed, cp = 1, rune(c&0x1F)
			case c >= 0xE0 && c <= 0xEF:
				if c == 0xE0 {
					lower = 0xA0
				} else if c == 0xED {
					upper = 0x9F
				}
				needed, cp = 2, rune(c&0xF)
			case c >= 0xF0 && c <= 0xF4:
				if c == 0xF0 {
					lower = 0x90
				} else if c == 0xF4 {
					upper = 0x8F
				}
				needed, cp = 3, rune(c&0x7)
			default:
				b.WriteRune(utf8.RuneError)
			}
			i++
			continue
		}
		if c < lower || c > upper {
			// The run so far is one U+FFFD; this byte is read again from the start.
			needed, seen, cp = 0, 0, 0
			lower, upper = 0x80, 0xBF
			b.WriteRune(utf8.RuneError)
			continue
		}
		lower, upper = 0x80, 0xBF
		cp = cp<<6 | rune(c&0x3F)
		seen++
		i++
		if seen == needed {
			b.WriteRune(cp)
			needed, seen, cp = 0, 0, 0
		}
	}
	if needed != 0 {
		b.WriteRune(utf8.RuneError)
	}
	return b.String()
}

// isSpace0 is JavaScript's whitespace: what \s matches and trim() removes.
func isSpace0(r rune) bool {
	switch r {
	case '\t', '\n', '\v', '\f', '\r', ' ', 0xA0, 0x1680, 0x2028, 0x2029, 0x202F, 0x205F, 0x3000, 0xFEFF:
		return true
	}
	return r >= 0x2000 && r <= 0x200A
}

// isLineEnd0 is what a JavaScript regular expression's . does not match.
func isLineEnd0(r rune) bool {
	return r == '\n' || r == '\r' || r == 0x2028 || r == 0x2029
}

// trimEnd0 is s.replace(/\s+$/, "").
func trimEnd0(s string) string {
	for len(s) > 0 {
		r, size := utf8.DecodeLastRuneInString(s)
		if !isSpace0(r) || (r == utf8.RuneError && size == 1) {
			break
		}
		s = s[:len(s)-size]
	}
	return s
}

// trim0 is s.trim().
func trim0(s string) string {
	for len(s) > 0 {
		r, size := utf8.DecodeRuneInString(s)
		if !isSpace0(r) || (r == utf8.RuneError && size == 1) {
			break
		}
		s = s[size:]
	}
	return trimEnd0(s)
}

// sliceUnits0 is s.slice(n) on a JavaScript string: it drops n UTF-16 units. When the cut falls inside a character
// above U+FFFF, JavaScript keeps that character's second unit, a lone low surrogate, written here in WTF-8.
func sliceUnits0(s string, n int) string {
	i := 0
	for n > 0 {
		if i >= len(s) {
			return ""
		}
		r, size := utf8.DecodeRuneInString(s[i:])
		i += size
		if r < 0x10000 {
			n--
			continue
		}
		if n == 1 {
			lo := 0xDC00 + (r-0x10000)&0x3FF
			return string([]byte{0xED, byte(0x80 | (lo>>6)&0x3F), byte(0x80 | lo&0x3F)}) + s[i:]
		}
		n -= 2
	}
	return s[i:]
}

// The refusals: yaml.mjs's own throws, each worded here with a next step (spec §3). A format-0 file is frozen by git,
// so the step names the line to change, which yaml.mjs refuses today too; contract §15 moves open tasks to format 1.
const (
	nextTab0         = "indent with spaces, never tabs"
	nextIndent0      = "line it up with the keys or list items beside it, or move it under a key that has nothing after its colon"
	nextKey0         = "write the line as key: value, a space after the colon, the key made of letters, digits, _ . $ - / or quoted"
	nextFlowEnd0     = "close the [ on the same line, or write the list as an indented block sequence"
	nextFlowNested0  = "write the list as an indented block sequence, or quote the item"
	nextFlowQuote0   = "close the quote inside the [ ], or write the list as an indented block sequence"
	nextAnchor0      = "quote the value if it is text, or write it out in full: anchors, aliases and tags are not read"
	nextFlowMapping0 = "write the mapping as indented key: value lines; {} alone is an empty mapping"
)

type parser0 struct {
	offset int // the file lines before the text's first: 1 in a frontmatter (its opening ---), else 0
}

// throw is yaml.mjs's throw new YamlError(n, ...): read0 recovers it as the refusal, its line counted in the file.
func (p *parser0) throw(n int, step, format string, args ...any) {
	panic(&Refusal{Line: n + p.offset, Message: fmt.Sprintf(format, args...), Next: step})
}

// tok0 is one line as tokenize gives it.
type tok0 struct {
	n      int    // the line's number in the text, from 1
	skip   bool   // blank, a comment or a document marker
	doc    bool   // a document marker, --- or ...
	indent int    // leading spaces (not set on a skip line)
	text   string // the line after its indent, without its trailing whitespace
	raw    string // the line as split
}

// split0 is String(text).split(/\r?\n/).
func split0(text string) []string {
	parts := strings.Split(text, "\n")
	for i := 0; i < len(parts)-1; i++ {
		parts[i] = strings.TrimSuffix(parts[i], "\r")
	}
	return parts
}

func (p *parser0) tokenize(text string) []tok0 {
	raw := split0(text)
	out := make([]tok0, 0, len(raw))
	for i, r := range raw {
		line := trimEnd0(r)
		n := i + 1
		if line == "" || isComment0(line) {
			out = append(out, tok0{n: n, skip: true, raw: r})
			continue
		}
		// /^---\s*$/ and /^\.\.\.\s*$/ on a line whose trailing whitespace is gone.
		if line == "---" || line == "..." {
			out = append(out, tok0{n: n, skip: true, doc: true, raw: r})
			continue
		}
		indent := 0
		for indent < len(line) && line[indent] == ' ' {
			indent++
		}
		if indent < len(line) && line[indent] == '\t' {
			p.throw(n, nextTab0, "a tab in the line's indentation")
		}
		out = append(out, tok0{n: n, indent: indent, text: line[indent:], raw: r})
	}
	return out
}

// isComment0 is /^\s*#/.
func isComment0(line string) bool {
	return strings.HasPrefix(trim0(line), "#")
}

// stripComment0 cuts s at the first # that starts s or follows whitespace, outside quotes. A quote opens wherever it
// stands, as in yaml.mjs.
func stripComment0(s string) string {
	var q byte
	for i := 0; i < len(s); i++ {
		c := s[i]
		if q != 0 {
			if c == q {
				q = 0
			}
			continue
		}
		if c == '"' || c == '\'' {
			q = c
			continue
		}
		if c == '#' {
			if i == 0 {
				return ""
			}
			if r, _ := utf8.DecodeLastRuneInString(s[:i]); isSpace0(r) {
				return s[:i]
			}
		}
	}
	return s
}

// flowSequence0 reads [a, b, "c, d"] on one line: scalars only.
func (p *parser0) flowSequence(s string, n int) []any {
	if s[len(s)-1] != ']' {
		p.throw(n, nextFlowEnd0, "a flow sequence that does not end with ] on its line: %s", show(s))
	}
	inner := ""
	if len(s) >= 2 {
		inner = trim0(s[1 : len(s)-1])
	}
	if inner == "" {
		return []any{}
	}
	// The items are cut at commas outside quotes; every special character is ASCII, so bytes walk as characters do.
	var items []string
	var buf strings.Builder
	var q byte
	for i := 0; i < len(inner); i++ {
		c := inner[i]
		if q != 0 {
			buf.WriteByte(c)
			if c == q {
				q = 0
			}
			continue
		}
		switch c {
		case '"', '\'':
			q = c
			buf.WriteByte(c)
		case '[', '{':
			p.throw(n, nextFlowNested0, "a [ or { inside a flow sequence: %s", show(s))
		case ',':
			items = append(items, buf.String())
			buf.Reset()
		default:
			buf.WriteByte(c)
		}
	}
	if q != 0 {
		p.throw(n, nextFlowQuote0, "a quote left open in a flow sequence: %s", show(s))
	}
	items = append(items, buf.String())
	out := make([]any, len(items))
	for i, x := range items {
		out[i] = p.scalar(x, n)
	}
	return out
}

func (p *parser0) scalar(rawValue string, n int) any {
	s := trim0(stripComment0(rawValue))
	if s == "" {
		return nil
	}
	last := s[len(s)-1]
	if len(s) > 1 && s[0] == '\'' && last == '\'' {
		return strings.ReplaceAll(s[1:len(s)-1], "''", "'")
	}
	if len(s) > 1 && s[0] == '"' && last == '"' {
		return unescape0(s[1 : len(s)-1])
	}
	switch s {
	case "[]":
		return []any{}
	case "{}":
		return newObj0()
	}
	switch s[0] {
	case '&', '*', '!':
		p.throw(n, nextAnchor0, "an anchor, alias or tag: %s", show(s))
	case '[':
		return p.flowSequence(s, n)
	case '{':
		p.throw(n, nextFlowMapping0, "a flow mapping: %s", show(s))
	}
	switch s {
	case "true", "True":
		return true
	case "false", "False":
		return false
	case "null", "Null", "~":
		return nil
	}
	if isInt0(s) || isFloat0(s) {
		// parseInt(s, 10) and parseFloat(s) on these forms both give the float64 nearest the decimal, as ParseFloat
		// does, ±Inf past the largest; the error then is only ErrRange.
		f, _ := strconv.ParseFloat(s, 64)
		return f
	}
	return s
}

// unescape0 is .replace(/\\(["\\nrt])/g, ...): left to right, each of the five escapes once.
func unescape0(s string) string {
	if !strings.Contains(s, `\`) {
		return s
	}
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) {
			switch s[i+1] {
			case '"', '\\':
				b.WriteByte(s[i+1])
				i++
				continue
			case 'n':
				b.WriteByte('\n')
				i++
				continue
			case 'r':
				b.WriteByte('\r')
				i++
				continue
			case 't':
				b.WriteByte('\t')
				i++
				continue
			}
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

// isInt0 is /^-?\d+$/ (\d is ASCII).
func isInt0(s string) bool {
	s = strings.TrimPrefix(s, "-")
	return s != "" && strings.Trim(s, "0123456789") == ""
}

// isFloat0 is /^-?\d*\.\d+$/.
func isFloat0(s string) bool {
	s = strings.TrimPrefix(s, "-")
	dot := strings.IndexByte(s, '.')
	if dot < 0 || dot == len(s)-1 {
		return false
	}
	return strings.Trim(s[:dot], "0123456789") == "" && strings.Trim(s[dot+1:], "0123456789") == ""
}

// key0 is yaml.mjs's KEY, /^((?:[A-Za-z0-9_.$][A-Za-z0-9_.\-$/]*)|'[^']*'|"[^"]*"):(?:\s+(.*))?$/: it returns the
// key as written (quotes and all) and what follows the whitespace after the colon.
func key0(text string) (key, rest string, ok bool) {
	if text == "" {
		return "", "", false
	}
	var end int // the colon
	switch c := text[0]; {
	case c == '\'' || c == '"':
		k := strings.IndexByte(text[1:], c)
		if k < 0 {
			return "", "", false
		}
		end = k + 2
	case keyStart0(c):
		end = 1
		for end < len(text) && keyChar0(text[end]) {
			end++
		}
	default:
		return "", "", false
	}
	if end >= len(text) || text[end] != ':' {
		return "", "", false
	}
	key, after := text[:end], text[end+1:]
	if after == "" {
		return key, "", true
	}
	// \s+ takes all the whitespace (a line end among it too); (.*) must then reach the end without a line end.
	i := 0
	for i < len(after) {
		r, size := utf8.DecodeRuneInString(after[i:])
		if !isSpace0(r) || (r == utf8.RuneError && size == 1) {
			break
		}
		i += size
	}
	if i == 0 {
		return "", "", false
	}
	for _, r := range after[i:] {
		if isLineEnd0(r) {
			return "", "", false
		}
	}
	return key, after[i:], true
}

func keyStart0(c byte) bool {
	return c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '_' || c == '.' || c == '$'
}

func keyChar0(c byte) bool {
	return keyStart0(c) || c == '-' || c == '/'
}

func unquoteKey0(k string) string {
	if k == "" {
		return k
	}
	first, last := k[0], k[len(k)-1]
	if first == '\'' && last == '\'' || first == '"' && last == '"' {
		if len(k) < 2 {
			return ""
		}
		return k[1 : len(k)-1]
	}
	return k
}

func next0(toks []tok0, i int) int {
	for i < len(toks) && toks[i].skip {
		i++
	}
	return i
}

// isItem0 is t.text === '-' || t.text.startsWith('- ').
func isItem0(text string) bool {
	return text == "-" || strings.HasPrefix(text, "- ")
}

// blockScalar0: every line more indented than the parent key belongs to it; blank and comment lines are empty lines,
// and a document marker ends it.
func (p *parser0) blockScalar(toks []tok0, i, parentIndent int, style, chomp byte) (any, int) {
	var lines []string
	base := -1
	for i < len(toks) {
		t := toks[i]
		if t.skip && !t.doc {
			lines = append(lines, "")
			i++
			continue
		}
		if t.doc || t.indent <= parentIndent {
			break
		}
		if base < 0 {
			base = t.indent
		}
		lines = append(lines, sliceUnits0(trimEnd0(t.raw), base))
		i++
	}
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	var value string
	if style == '|' {
		value = strings.Join(lines, "\n")
	} else {
		var b strings.Builder
		// A blank line is a line break; other lines join with a space, except right after a break.
		lastBreak, started := false, false
		for _, l := range lines {
			if l == "" {
				b.WriteByte('\n')
				lastBreak, started = true, true
				continue
			}
			if started && !lastBreak {
				b.WriteByte(' ')
			}
			b.WriteString(l)
			lastBreak, started = false, true
		}
		value = b.String()
	}
	if chomp != '-' {
		value += "\n"
	}
	return value, i
}

// block0 is yaml.mjs's BLOCK, /^([|>])([-+]?)$/: the style and the chomping indicator (0 for none).
func block0(rest string) (style, chomp byte, ok bool) {
	if rest == "" || (rest[0] != '|' && rest[0] != '>') {
		return 0, 0, false
	}
	switch rest[1:] {
	case "":
		return rest[0], 0, true
	case "-", "+":
		return rest[0], rest[1], true
	}
	return 0, 0, false
}

func (p *parser0) parseValueAfterKey(toks []tok0, i, indent int, rest string, n int) (any, int) {
	if style, chomp, ok := block0(rest); ok {
		return p.blockScalar(toks, i, indent, style, chomp)
	}
	if rest != "" {
		return p.scalar(rest, n), i
	}
	// Nothing after the colon: a nested collection, or null.
	j := next0(toks, i)
	if j >= len(toks) || toks[j].indent <= indent {
		return nil, i
	}
	return p.parseCollection(toks, j, toks[j].indent)
}

func (p *parser0) parseMap(toks []tok0, i, indent int) (*obj0, int) {
	obj := newObj0()
	for {
		i = next0(toks, i)
		if i >= len(toks) {
			break
		}
		t := toks[i]
		if t.indent < indent {
			break
		}
		if t.indent > indent {
			p.throw(t.n, nextIndent0, "unexpected indentation: %s", show(t.text))
		}
		if isItem0(t.text) {
			break
		}
		k, rest, ok := key0(t.text)
		if !ok {
			p.throw(t.n, nextKey0, "cannot read %s as key: value", show(t.text))
		}
		value, ni := p.parseValueAfterKey(toks, i+1, indent, trim0(stripComment0(rest)), t.n)
		obj.set(unquoteKey0(k), value, t.n+p.offset)
		i = ni
	}
	return obj, i
}

func (p *parser0) parseList(toks []tok0, i, indent int) ([]any, int) {
	arr := []any{}
	for {
		i = next0(toks, i)
		if i >= len(toks) {
			break
		}
		t := toks[i]
		if t.indent < indent {
			break
		}
		if t.indent > indent {
			p.throw(t.n, nextIndent0, "unexpected indentation: %s", show(t.text))
		}
		if !isItem0(t.text) {
			break
		}
		rest := ""
		if t.text != "-" {
			rest = trim0(t.text[2:])
		}
		if rest == "" {
			j := next0(toks, i+1)
			if j < len(toks) && toks[j].indent > indent {
				v, ni := p.parseCollection(toks, j, toks[j].indent)
				arr = append(arr, v)
				i = ni
				continue
			}
			arr = append(arr, nil)
			i++
			continue
		}
		if k, kRest, ok := key0(rest); ok {
			// "- key: value" opens a mapping indented two past the dash, whatever the spaces after it.
			itemIndent := indent + 2
			value, ni := p.parseValueAfterKey(toks, i+1, itemIndent, trim0(stripComment0(kRest)), t.n)
			more, nk := p.parseMap(toks, ni, itemIndent)
			// { [key]: value, ...more }: a new object; both define own keys, so __proto__ is an own key here.
			item := newObj0()
			item.define(unquoteKey0(k), value, t.n+p.offset)
			for _, mk := range more.ownKeys() {
				pr := more.props[mk]
				item.define(mk, pr.value, pr.line)
			}
			arr = append(arr, item)
			i = nk
			continue
		}
		if style, chomp, ok := block0(rest); ok {
			v, ni := p.blockScalar(toks, i+1, indent, style, chomp)
			arr = append(arr, v)
			i = ni
			continue
		}
		arr = append(arr, p.scalar(rest, t.n))
		i++
	}
	return arr, i
}

func (p *parser0) parseCollection(toks []tok0, i, indent int) (any, int) {
	if isItem0(toks[i].text) {
		return p.parseList(toks, i, indent)
	}
	return p.parseMap(toks, i, indent)
}

func (p *parser0) parseYaml(text string) any {
	toks := p.tokenize(text)
	i := next0(toks, 0)
	if i >= len(toks) {
		return newObj0()
	}
	v, _ := p.parseCollection(toks, i, toks[i].indent)
	return v
}

// obj0 is a JavaScript object as yaml.mjs builds one: own keys in the order they were first set, and whether
// assigning __proto__ reaches Object.prototype's setter.
type obj0 struct {
	keys  []string
	props map[string]*prop0
	// setter: the prototype chain reaches Object.prototype before any object with an own __proto__. A new object's
	// prototype is Object.prototype; obj[key] = value with key __proto__ then sets the prototype instead of a key.
	setter bool
}

type prop0 struct {
	value any
	line  int
}

func newObj0() *obj0 {
	return &obj0{props: map[string]*prop0{}, setter: true}
}

// set is obj[key] = value. For __proto__ with no own __proto__ key and the setter reached: an object or array
// becomes the prototype, null makes it null, anything else is ignored; no key is set either way. Otherwise it
// defines the key.
func (o *obj0) set(key string, v any, line int) {
	if _, own := o.props[key]; !own && key == "__proto__" && o.setter {
		switch x := v.(type) {
		case nil:
			o.setter = false
		case *obj0:
			_, xOwn := x.props["__proto__"]
			o.setter = !xOwn && x.setter
		case []any:
			o.setter = true // an array's chain is Array.prototype, then Object.prototype
		}
		return
	}
	o.define(key, v, line)
}

// define is CreateDataProperty: a new key goes last, a key already there keeps its place and takes the value.
func (o *obj0) define(key string, v any, line int) {
	if pr, ok := o.props[key]; ok {
		pr.value, pr.line = v, line
		return
	}
	o.keys = append(o.keys, key)
	o.props[key] = &prop0{v, line}
}

// ownKeys is [[OwnPropertyKeys]]: array-index keys in ascending order, then the rest in the order first set.
func (o *obj0) ownKeys() []string {
	var idx, other []string
	for _, k := range o.keys {
		if _, ok := arrayIndex0(k); ok {
			idx = append(idx, k)
		} else {
			other = append(other, k)
		}
	}
	if len(idx) == 0 {
		return other
	}
	sort.Slice(idx, func(a, b int) bool {
		x, _ := arrayIndex0(idx[a])
		y, _ := arrayIndex0(idx[b])
		return x < y
	})
	return append(idx, other...)
}

// arrayIndex0 reports whether k is an array index: an integer from 0 to 2^32-2 written as JavaScript writes it.
func arrayIndex0(k string) (uint64, bool) {
	if k == "" || len(k) > 10 || (k[0] == '0' && len(k) > 1) {
		return 0, false
	}
	n, err := strconv.ParseUint(k, 10, 64)
	if err != nil || n > math.MaxUint32-1 {
		return 0, false
	}
	return n, true
}

// finish0 turns what the parser built into the values Result0 gives: each obj0 a *Map in JavaScript's key order.
func finish0(v any) any {
	switch x := v.(type) {
	case *obj0:
		m := &Map{}
		for _, k := range x.ownKeys() {
			pr := x.props[k]
			m.add(k, finish0(pr.value), pr.line)
		}
		return m
	case []any:
		out := make([]any, len(x))
		for i, e := range x {
			out[i] = finish0(e)
		}
		return out
	}
	return v
}
