package testout

import (
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/LastStep/Bonsai/internal/format"
)

// Node's test runner prints one TAP line per test when its output is piped (node --test --test-reporter=tap):
//
//	ok 3 - a name # SKIP real-checkouts-only: why        not ok 4 - another # TODO
//
// Subtests sit under "# Subtest: <name>" headers, four spaces deeper a level, and their own ok line comes after them;
// a test's full name is its suites' and parents' names and its own, joined with " > ". A # inside a name is written
// \#. Each ok line is followed by a YAML block two spaces deeper, whose type says a suite (describe) from a test
// (it, test, t.test); suites are dropped, only tests are judged. A not ok test also keeps from its block its file
// (location:), its failureType and its error:
//
//	location: 'E:\\x\\a.test.mjs:3:1'      error: 'one line'   or   error: |-
//
// A single-line value is quoted the way util.inspect quotes a string; a longer one is a block, two spaces deeper. A
// file that fails to load is reported as a failing test named after the file. This is today's ladder runner's parseTap,
// ported line for line, plus each test's kind and the ancestors that give its Unit.

// tapSubtest matches a "# Subtest: <name>" header: its indent and name.
var tapSubtest = sync.OnceValue(func() *regexp.Regexp { return regexp.MustCompile(`^( *)# Subtest: (.*)$`) })

// tapResult matches an ok or not ok line: its indent, ok or not ok, and the rest after " - ".
var tapResult = sync.OnceValue(func() *regexp.Regexp { return regexp.MustCompile(`^( *)(ok|not ok) \d+ - (.*)$`) })

// tapLineCol matches the ":line:col" a location ends in.
var tapLineCol = sync.OnceValue(func() *regexp.Regexp { return regexp.MustCompile(`:\d+:\d+$`) })

// tapTest is one ok or not ok line and what its block said.
type tapTest struct {
	path        []string
	ancestors   []*tapHeader // the headers of its suites and parents, outermost first, one per path part but its own
	ok          bool
	directive   string // "SKIP", "TODO" or ""
	reason      string
	depth       int
	kind        string // its block's type: "suite", "test", or "" when it said none
	file        string
	failureType string
	err         string
}

// tapHeader is one "# Subtest:" header; its test is the ok line that closes it, once read.
type tapHeader struct {
	name string
	test *tapTest
}

// tapBlock is a not ok test's multi-line error being read.
type tapBlock struct {
	indent int
	lines  []string
}

// readTAP reads node's TAP (Read).
func readTAP(stdout []byte, root string) Report {
	var tests []*tapTest
	var stack []*tapHeader // by depth; nil where no header was read
	var last *tapTest
	var block *tapBlock
	for _, line := range splitLines(stdout) {
		if block != nil {
			if strings.TrimSpace(line) == "" || strings.HasPrefix(line, strings.Repeat(" ", block.indent)) {
				if l := strings.TrimSpace(line); l != "" {
					block.lines = append(block.lines, l)
				}
				continue
			}
			last.err = errorLine(block.lines...)
			block = nil
		}
		if m := tapSubtest().FindStringSubmatch(line); m != nil {
			depth := len(m[1]) / 4
			stack = append(resize(stack, depth), &tapHeader{name: unescapeTAP(m[2])})
			continue
		}
		if m := tapResult().FindStringSubmatch(line); m != nil {
			depth := len(m[1]) / 4
			own, directive, reason := tapDirective(m[3])
			t := &tapTest{ok: m[2] == "ok", directive: directive, reason: reason, depth: depth}
			for _, h := range stack[:min(depth, len(stack))] {
				if h != nil {
					t.path = append(t.path, h.name)
					t.ancestors = append(t.ancestors, h)
				}
			}
			t.path = append(t.path, unescapeTAP(own))
			if depth < len(stack) && stack[depth] != nil {
				stack[depth].test = t
			}
			tests = append(tests, t)
			last = t
			stack = resize(stack, depth)
			continue
		}
		if last == nil {
			continue
		}
		pad := strings.Repeat(" ", last.depth*4+2)
		if v, ok := strings.CutPrefix(line, pad+"type: "); ok {
			last.kind = inspected(v)
			continue
		}
		if last.ok || !strings.HasPrefix(line, pad) {
			continue
		}
		if v, ok := strings.CutPrefix(line, pad+"location: "); ok {
			last.file = projectFile(fromFileURL(tapLineCol().ReplaceAllString(inspected(v), "")), root)
		} else if v, ok := strings.CutPrefix(line, pad+"failureType: "); ok {
			last.failureType = inspected(v)
		} else if v, ok := strings.CutPrefix(line, pad+"error: "); ok {
			if isBlockMark(v) {
				block = &tapBlock{indent: len(pad) + 2}
			} else {
				last.err = errorLine(inspected(v))
			}
		}
	}
	if block != nil {
		last.err = errorLine(block.lines...)
	}
	r := Report{Form: format.TestsTAP}
	for _, t := range tests {
		if t.kind == "suite" {
			continue
		}
		out := Test{Name: strings.Join(t.path, " > "), Path: t.path, Unit: len(t.path), File: t.file, Error: t.err,
			Parent: t.failureType == "subtestsFailed"}
		for i, h := range t.ancestors {
			if h.test == nil || h.test.kind != "suite" { // a test, or a header whose kind was never read
				out.Unit = i + 1
				break
			}
		}
		switch {
		case t.directive == "SKIP":
			out.Outcome, out.Reason = Skipped, t.reason
		case t.directive == "TODO":
			out.Outcome, out.Reason = Todo, t.reason
		case t.ok:
			out.Outcome = Passed
		default:
			out.Outcome = Failed
		}
		r.Tests = append(r.Tests, out)
	}
	return r
}

// resize gives the header stack n entries: cut, or grown with nil (a depth with no header read).
func resize(stack []*tapHeader, n int) []*tapHeader {
	if len(stack) >= n {
		return stack[:n]
	}
	for len(stack) < n {
		stack = append(stack, nil)
	}
	return stack
}

// tapDirective splits an ok line's text at its first " # SKIP" or " # TODO" (any case, a whole word, its # not
// escaped): the name before it, the directive in upper case, and its reason, unescaped and trimmed. With none, the
// whole text is the name.
func tapDirective(s string) (name, directive, reason string) {
	for i := 0; i+3 <= len(s); i++ {
		if s[i:i+3] != " # " || (i > 0 && s[i-1] == '\\') {
			continue
		}
		rest := s[i+3:]
		if len(rest) < 4 {
			continue
		}
		word := strings.ToUpper(rest[:4])
		if word != "SKIP" && word != "TODO" {
			continue
		}
		after := rest[4:]
		if after != "" && isWordByte(after[0]) {
			continue
		}
		return s[:i], word, unescapeTAP(strings.TrimSpace(strings.TrimPrefix(after, " ")))
	}
	return s, "", ""
}

// isWordByte is true for a byte a regular expression's \b counts as a word's: a letter, a digit or _.
func isWordByte(c byte) bool {
	return c == '_' || (c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

// unescapeTAP reads a name's \\ and \# as \ and #.
func unescapeTAP(s string) string {
	if !strings.Contains(s, `\`) {
		return s
	}
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) && (s[i+1] == '\\' || s[i+1] == '#') {
			i++
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

// isBlockMark is true for a YAML block scalar's mark (| or >, then - or + or nothing): the value is the lines below.
func isBlockMark(v string) bool {
	if v == "" || (v[0] != '|' && v[0] != '>') {
		return false
	}
	return len(v) == 1 || (len(v) == 2 && (v[1] == '-' || v[1] == '+'))
}

// inspected turns a string as util.inspect printed it ('...', "..." or `...`, with backslash escapes) back to its
// text; a value not quoted that way is returned trimmed.
func inspected(v string) string {
	s := strings.TrimSpace(v)
	if len(s) < 2 {
		return s
	}
	q := s[0]
	if (q != '\'' && q != '"' && q != '`') || s[len(s)-1] != q {
		return s
	}
	body := s[1 : len(s)-1]
	var b strings.Builder
	for i := 0; i < len(body); i++ {
		c := body[i]
		if c != '\\' || i+1 == len(body) {
			b.WriteByte(c)
			continue
		}
		e := body[i+1]
		if (e == 'x' && i+4 <= len(body)) || (e == 'u' && i+6 <= len(body)) {
			digits := body[i+2 : i+4]
			if e == 'u' {
				digits = body[i+2 : i+6]
			}
			if n, err := strconv.ParseUint(digits, 16, 32); err == nil {
				b.WriteRune(rune(n))
				i += 1 + len(digits)
				continue
			}
		}
		switch e {
		case 'n':
			b.WriteByte('\n')
		case 't':
			b.WriteByte('\t')
		case 'r':
			b.WriteByte('\r')
		case 'b':
			b.WriteByte('\b')
		case 'f':
			b.WriteByte('\f')
		case 'v':
			b.WriteByte('\v')
		case '0':
			b.WriteByte(0)
		default:
			b.WriteByte(e)
		}
		i++
	}
	return b.String()
}

// fromFileURL turns a file:// URL into its path (file:///C:/x/a.mjs is C:/x/a.mjs); anything else is returned as is.
func fromFileURL(s string) string {
	rest, ok := strings.CutPrefix(s, "file://")
	if !ok {
		return s
	}
	if u, err := url.PathUnescape(rest); err == nil {
		rest = u
	}
	if len(rest) >= 3 && rest[0] == '/' && rest[2] == ':' && isLetter(rest[1]) {
		rest = rest[1:]
	}
	return rest
}

// isLetter is true for an ASCII letter.
func isLetter(c byte) bool { return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') }

// tapRun is the node command line that runs t alone (Run): "" when the output named no file for it.
func tapRun(t Test, in RunInput) string {
	if t.File == "" || len(t.Path) == 0 {
		return ""
	}
	parts := make([]string, unit(t))
	for i := range parts {
		parts[i] = jsEscape(t.Path[i])
	}
	pattern := "^" + strings.Join(parts, " ") + "$"
	return quoteLine([]string{"node", "--test", "--test-name-pattern=" + pattern, t.File}, in.Windows)
}

// jsEscape escapes a name for a JavaScript regular expression: each of \^$.|?*+()[]{}/ and - gets a backslash.
func jsEscape(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if strings.IndexByte(`\^$.|?*+()[]{}/-`, s[i]) >= 0 {
			b.WriteByte('\\')
		}
		b.WriteByte(s[i])
	}
	return b.String()
}
