// Package reader is Bonsai's format-1 YAML reader (contract §2.4): its own grammar, no general YAML library (spec
// §3). It reads a YAML definition file (ReadYAML) or a markdown file's frontmatter (ReadMarkdown) and reaches one of
// the three outcomes formats/README.md defines for every case of the formats set:
//
//   - Accepted, with the value: the top-level mapping, its keys in file order (a *Map);
//   - Refused, with one reason code (Codes) and the line it was found on;
//   - Format0, when the file has no top-level format: key, so the file is format 0 (contract §2.3). This reader
//     reads nothing under format 0: the hand port of yaml.mjs is step 5.1 (plan part 2), so the caller decides.
//
// Dispatch comes first (contract §2.4): the first top-level key decides. format: first means format 1; no top-level
// format: key means format 0; a top-level format: anywhere else is refused (format-not-first), whatever else the
// file holds. A format: value naming a major other than 1 is refused as too new before anything else is read
// (contract §2.2: "the reader says format too new ... and parses nothing else").
//
// Under format 1 the file is read top to bottom by the grammar in parse.go and scalar.go, and the first problem
// met is the one reported: on one line, the key before its value, and of two codes that fit one problem, the one
// Codes lists first. Every rule of §2.4 is enforced in this package's code, each at the place the grammar meets
// it; the comments there name the rule.
//
// Values: a mapping is a *Map, a sequence []any, text string, true and false bool, null nil, an integer int64 (at
// most 15 digits), a decimal Decimal (its exact digits). A date or a time stays text. JSON gives the value as the
// JSON a reader returns (formats/README.md, "value").
package reader

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Outcome is what a read reaches: the three outcomes of formats/README.md.
type Outcome int

const (
	// Accepted: the file was read under format 1; Result.Value holds the mapping.
	Accepted Outcome = iota
	// Refused: the file breaks a rule; Result.Refusal says which, and where.
	Refused
	// Format0: the file has no top-level format: key, so it is format 0 (contract §2.3).
	Format0
)

// String names the outcome as formats/expect.json does.
func (o Outcome) String() string {
	switch o {
	case Accepted:
		return "accepted"
	case Refused:
		return "refused"
	case Format0:
		return "format-0"
	}
	return "unknown"
}

// Result is one read's outcome.
type Result struct {
	Outcome Outcome
	Value   *Map     // the top-level mapping, when Accepted
	Refusal *Refusal // the reason, when Refused
}

// Err returns the refusal as an error, or nil when the file was not refused.
func (r Result) Err() error {
	if r.Outcome == Refused {
		return r.Refusal
	}
	return nil
}

// Refusal says why a file was refused: one reason code, the line (1-based, counted in the whole file, a markdown
// file's opening --- being line 1), what is wrong and what to do. Message and Next are ASCII: a file's own text is
// quoted with Go's ASCII escapes.
type Refusal struct {
	Code    string
	Line    int
	Message string
	Next    string
}

// Error prints the refusal as one ASCII line that names the next step.
func (r *Refusal) Error() string {
	return fmt.Sprintf("line %d: %s (%s); next: %s", r.Line, r.Message, r.Code, r.Next)
}

// The reason codes, in the order formats/README.md lists them: of two codes that fit one problem, the one listed
// first is reported. Two codes are this reader's own, outside the set's table (a test holds the rest equal to it):
//   - format-too-new, contract §2.2's rule for a newer major, checked with dispatch;
//   - not-text, for a line that is not text YAML allows, where a YAML 1.1 and a YAML 1.2 library would read
//     different lines or refuse (§2.4's own test): invalid UTF-8; a control character (any but tab) or a
//     noncharacter; a CR that does not end a line; a character a YAML 1.1 reader takes as a line break (U+0085,
//     U+2028, U+2029); a file that ends inside a | or > block scalar with no line ending.
//     The set has no case for either code; plan part 2's report names them for its next version.
const (
	CodeFormatNotFirst = "format-not-first"
	CodeFormatTooNew   = "format-too-new"
	CodeNotText        = "not-text"
	CodeTabIndent      = "tab-indent"
	CodeDocMarker      = "doc-marker"
	CodeLineNotRead    = "line-not-read"
	CodeKeyComplex     = "key-complex"
	CodeKeyQuoted      = "key-quoted"
	CodeKeyMerge       = "key-merge"
	CodeKeyForm        = "key-form"
	CodeKeyReserved    = "key-reserved"
	CodeKeyTwice       = "key-twice"
	CodeSeqDashSpace   = "seq-dash-space"
	CodeAnchor         = "anchor"
	CodeAlias          = "alias"
	CodeTag            = "tag"
	CodeFlowMapping    = "flow-mapping"
	CodeFlowNested     = "flow-nested"
	CodeFlowMultiline  = "flow-multiline"
	CodeBlockIndicator = "block-indicator"
	CodeBlockHashLine  = "block-hash-line"
	CodeFoldedDeeper   = "folded-deeper"
	CodeQuotedMulti    = "quoted-multiline"
	CodeBadEscape      = "bad-escape"
	CodeUnescapedQuote = "unescaped-quote"
	CodeAfterQuote     = "after-quote"
	CodeQuoteThisValue = "quote-this-value"
)

// Codes lists every code this reader reports, in order.
var Codes = []string{
	CodeFormatNotFirst, CodeFormatTooNew, CodeNotText, CodeTabIndent, CodeDocMarker, CodeLineNotRead,
	CodeKeyComplex, CodeKeyQuoted, CodeKeyMerge, CodeKeyForm, CodeKeyReserved, CodeKeyTwice, CodeSeqDashSpace,
	CodeAnchor, CodeAlias, CodeTag, CodeFlowMapping, CodeFlowNested, CodeFlowMultiline, CodeBlockIndicator,
	CodeBlockHashLine, CodeFoldedDeeper, CodeQuotedMulti, CodeBadEscape, CodeUnescapedQuote, CodeAfterQuote,
	CodeQuoteThisValue,
}

// OwnCodes are the codes in Codes that formats/README.md's table does not list (see above).
var OwnCodes = []string{CodeFormatTooNew, CodeNotText}

// next is what to do about each code: the next step every refusal names (CLAUDE.md, spec §3).
var next = map[string]string{
	CodeFormatNotFirst: "move the format: line to the top of the file (comments and blank lines may come before it)",
	CodeFormatTooNew:   "read this file with a Bonsai that knows its format's major",
	CodeNotText:        "save the file as UTF-8 text with no control characters, its lines ending in LF or CRLF",
	CodeTabIndent:      "indent with spaces, never tabs",
	CodeDocMarker:      "remove the --- or ... line: a YAML file, and a frontmatter, is one document",
	CodeLineNotRead:    "indent the line under the key it belongs to, deeper than that key (a list item too), or remove it",
	CodeKeyComplex:     "write the key plainly, as key: value",
	CodeKeyQuoted:      "write the key without quotes",
	CodeKeyMerge:       "write the merged keys out in full",
	CodeKeyForm:        "use a key of lower-case letters, digits and _ starting with a letter, or a label name <namespace>.<name>",
	CodeKeyReserved:    "rename the key: a YAML 1.1 reader takes y, n, yes, no, on, off, true, false and null as a boolean or null",
	CodeKeyTwice:       "keep one of the two keys",
	CodeSeqDashSpace:   "write exactly one space after the -",
	CodeAnchor:         "write the value out in full: anchors are not read",
	CodeAlias:          "write the value out in full: aliases are not read",
	CodeTag:            "remove the tag; quote the value if it is text",
	CodeFlowMapping:    "write the mapping as indented key: value lines; {} alone is an empty map",
	CodeFlowNested:     "write the list as an indented block sequence, or quote the item",
	CodeFlowMultiline:  "close the [ on the same line, or write the list as an indented block sequence",
	CodeBlockIndicator: "use one of |, |-, > and >-",
	CodeBlockHashLine:  "start the line with another character: inside a block scalar a # line is text to YAML but a comment to format 0",
	CodeFoldedDeeper:   "indent every line of a > block the same, or use | to keep the line breaks",
	CodeQuotedMulti:    "keep a quoted value on one line, or use a block scalar (| or >)",
	CodeBadEscape:      "use only \\\", \\\\, \\n, \\r and \\t inside double quotes, or single quotes for a backslash",
	CodeUnescapedQuote: "write a quote inside double quotes as \\\", or use single quotes",
	CodeAfterQuote:     "put nothing but spaces and a comment after the closing quote",
	CodeQuoteThisValue: "quote the value (\"...\" or '...'): format 1 reads it bare as nothing else",
}

func refuse(code string, lineNo int, format string, args ...any) *Refusal {
	return &Refusal{Code: code, Line: lineNo, Message: fmt.Sprintf(format, args...), Next: next[code]}
}

// show quotes a piece of the file for a message, in ASCII, cut to 60 bytes.
func show(s string) string {
	if len(s) > 60 {
		cut := 60
		for cut > 0 && !utf8.RuneStart(s[cut]) {
			cut--
		}
		s = s[:cut] + "..."
	}
	return strconv.QuoteToASCII(s)
}

var bom = []byte{0xEF, 0xBB, 0xBF}

// ReadYAML reads a YAML definition file's bytes. One leading BOM is ignored; LF and CRLF lines are read alike.
func ReadYAML(raw []byte) Result {
	raw = bytes.TrimPrefix(raw, bom)
	return read(splitLines(string(raw), 1))
}

// ReadMarkdown reads a markdown file's frontmatter: the lines between its opening --- (the file's first line, after
// one ignored BOM) and the next line that is --- alone, each line with its own line ending (formats/README.md). The
// body is not read. A file with no such block has no format: key, so it is format 0, as format 0's reader sees it
// (yaml.mjs reads no frontmatter there either).
func ReadMarkdown(raw []byte) Result {
	raw = bytes.TrimPrefix(raw, bom)
	all := splitLines(string(raw), 1)
	if len(all) == 0 || all[0].text != "---" || !all[0].eol {
		return Result{Outcome: Format0}
	}
	for i := 1; i < len(all); i++ {
		if all[i].text == "---" {
			return read(all[1:i])
		}
	}
	return Result{Outcome: Format0}
}

// line is one line of the YAML being read.
type line struct {
	no     int    // 1-based line number in the file
	text   string // the line without its line ending (the CR of a CRLF removed)
	eol    bool   // the line ended in LF or CRLF (only a file's last line may not)
	indent int    // leading spaces
	body   string // the text after the leading spaces
	blank  bool   // nothing but spaces
}

// splitLines cuts text into lines at LF, removing the CR of each CRLF. Any other CR stays in its line, where the
// not-text check refuses it.
func splitLines(s string, firstNo int) []line {
	var out []line
	no := firstNo
	for len(s) > 0 {
		var t string
		eol := false
		if i := strings.IndexByte(s, '\n'); i >= 0 {
			t, s, eol = s[:i], s[i+1:], true
			t = strings.TrimSuffix(t, "\r")
		} else {
			t, s = s, ""
		}
		n := 0
		for n < len(t) && t[n] == ' ' {
			n++
		}
		out = append(out, line{no: no, text: t, eol: eol, indent: n, body: t[n:], blank: n == len(t)})
		no++
	}
	return out
}

// check finds the problems a line has before any grammar reads it, in Codes' order: not-text, tab-indent, then
// doc-marker. Every line the reader reaches is checked first, whatever it turns out to be.
func check(l *line) *Refusal {
	// Lines (§2.4): UTF-8 text, LF or CRLF.
	if !utf8.ValidString(l.text) {
		return refuse(CodeNotText, l.no, "the line is not valid UTF-8")
	}
	for _, r := range l.text {
		if bad := badChar(r); bad != "" {
			return refuse(CodeNotText, l.no, "the line holds %s (%U)", bad, r)
		}
	}
	// Lines: no tab in indentation. A tab anywhere in a line's leading whitespace is refused, a blank line's too.
	for i := 0; i < len(l.text) && (l.text[i] == ' ' || l.text[i] == '\t'); i++ {
		if l.text[i] == '\t' {
			return refuse(CodeTabIndent, l.no, "a tab in the line's indentation")
		}
	}
	// Lines: one document. A --- or ... at the start of a line is a document marker: in a YAML file there is none,
	// and a frontmatter's own two markers are not among its lines.
	if isDocMarker(l.text) {
		return refuse(CodeDocMarker, l.no, "a document marker %s", show(l.text))
	}
	return nil
}

func isDocMarker(t string) bool {
	if !strings.HasPrefix(t, "---") && !strings.HasPrefix(t, "...") {
		return false
	}
	return len(t) == 3 || t[3] == ' ' || t[3] == '\t'
}

// badChar names a character YAML text may not hold here, or returns "".
func badChar(r rune) string {
	switch {
	case r == '\t':
		return ""
	case r == '\r':
		return "a CR that does not end a line"
	case r < 0x20 || r == 0x7f || (r >= 0x80 && r <= 0x9f && r != 0x85):
		return "a control character"
	case r == 0x85 || r == 0x2028 || r == 0x2029:
		return "a character a YAML 1.1 reader takes as a line break"
	case r == 0xfffe || r == 0xffff:
		return "a noncharacter"
	}
	return ""
}

// formatKey matches a format: value that names a format and its major.
var formatValue = regexp.MustCompile(`^[a-z][a-z0-9_.-]*/([1-9][0-9]*)$`)

// read dispatches, then reads under format 1.
func read(lines []line) Result {
	first, at, ok := dispatch(lines)
	if !ok {
		return Result{Outcome: Format0}
	}
	if first != at {
		return Result{Outcome: Refused, Refusal: refuse(CodeFormatNotFirst, lines[at].no,
			"format: is a top-level key here but not the first one (the first is on line %d)", lines[first].no)}
	}
	// Contract §2.2: a newer major is refused before anything else is read.
	if v, ok := formatLineValue(&lines[at]); ok {
		if m := formatValue.FindStringSubmatch(v); m != nil && m[1] != "1" {
			return Result{Outcome: Refused, Refusal: refuse(CodeFormatTooNew, lines[at].no,
				"format too new: %s (this Bonsai reads major 1)", show(v))}
		}
	}
	p := &parser{lines: lines}
	m, err := p.document()
	if err != nil {
		return Result{Outcome: Refused, Refusal: err}
	}
	return Result{Outcome: Accepted, Value: m}
}

// dispatch finds the top-level keys the way format 0's reader (yaml.mjs) tokenizes a file: blank lines, comment
// lines, document markers and tab-indented lines are skipped; the first other line sets the top level's
// indentation; a top-level key is a line at that indentation that reads as key: (a plain key up to its colon, or a
// quoted one). It returns the index of the first top-level key and of the first format: key, and whether there is
// a format: key at all.
func dispatch(lines []line) (first, format int, ok bool) {
	first, format, root := -1, -1, -1
	for i := range lines {
		l := &lines[i]
		if l.blank || strings.HasPrefix(l.body, "#") || isDocMarker(l.text) || strings.HasPrefix(l.body, "\t") {
			continue
		}
		if root < 0 {
			root = l.indent
		}
		if l.indent != root {
			continue
		}
		name, isKey := dispatchKey(l.body)
		if !isKey {
			continue
		}
		if first < 0 {
			first = i
		}
		if name == "format" {
			return first, i, true
		}
	}
	return -1, -1, false
}

// dispatchKey reads a line as a key: line for dispatch: a quoted key (its text between the quotes) or a plain key
// (the text before the first colon followed by a space or the line's end, before any comment).
func dispatchKey(body string) (string, bool) {
	if body == "" || body[0] == '-' && (len(body) == 1 || body[1] == ' ' || body[1] == '\t') || body[0] == '?' {
		return "", false
	}
	if body[0] == '"' || body[0] == '\'' {
		end := strings.IndexByte(body[1:], body[0])
		if end < 0 {
			return "", false
		}
		rest := strings.TrimLeft(body[end+2:], " ")
		if rest == ":" || strings.HasPrefix(rest, ": ") {
			return body[1 : end+1], true
		}
		return "", false
	}
	sep := keySeparator(body)
	if sep < 0 {
		return "", false
	}
	return body[:sep], true
}

// formatLineValue reads the value of a plain format: key line, when it is a plain or quoted scalar.
func formatLineValue(l *line) (string, bool) {
	if !strings.HasPrefix(l.body, "format:") {
		return "", false
	}
	rest := strings.TrimLeft(l.body[len("format:"):], " ")
	if rest == "" {
		return "", false
	}
	if rest[0] == '"' || rest[0] == '\'' {
		v, _, err := quoted(rest, l.no)
		return v, err == nil
	}
	return plainText(rest), true
}
