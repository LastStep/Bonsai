// Package testout reads a ladder rung's test output (plan 5.4.2): the standard output of a rung whose tests field in
// bonsai.yaml names its form (format.TestForms: go-test-json, tap, junit), read into one shape the runner judges. Each
// test comes with its outcome, a skip's or todo's reason, its file and a failure's message; a failure outside any test
// (a package that did not build, an output that is not a report) is a Broken. Run gives a test the command that runs
// it alone, quoted for the shell its rung runs in.
//
//	r, ok := testout.Read(format.TestsGoJSON, stdout, root)
//	for _, t := range r.Tests {
//		if t.Outcome == testout.Failed && !t.Parent {
//			run := testout.Run(r.Form, t, testout.RunInput{Command: rungCommand, Root: root})
//			// go test -count=1 -run '^TestParent$/^sub_two$' ./good
//		}
//	}
//
// It runs nothing, starts no process and writes nothing. It reads one file: root/go.mod's module line, at most once
// per Read of a go-test-json output and once per Run of a go test. Its regular expressions are compiled on first use,
// never when the program starts. Its output is byte-stable: tests in the order the output finished them, broken
// packages in the order they ended, no map-order iteration.
package testout

import (
	"bytes"
	"path"
	"runtime"
	"strings"
	"unicode/utf8"

	"github.com/LastStep/Bonsai/internal/format"
)

// The outcomes of one test.
const (
	Passed  = "passed"
	Failed  = "failed"
	Skipped = "skipped"
	Todo    = "todo" // TAP's # TODO; never failed, never passed
)

// maxText is the most characters an Error keeps; a longer one ends in "..." within it.
const maxText = 300

// Test is one test the output names.
type Test struct {
	// Name is the full name: go-test-json's "<package>.<test>" ("example.com/demo/good.TestParent/sub_one"), tap's
	// names joined by " > " ("outer suite > inner one"), junit's "<classname>.<name>" (its name alone without a
	// classname). A go-test-json name is unique within one output, unless its tests ran more than once (go test
	// -count); node and JUnit reporters allow two tests of one name, which are then read as two tests of that name.
	Name string
	// Path is the name's parts, outermost first, unescaped: go's package then each level of the test's name
	// (["example.com/demo/good", "TestParent", "sub_one"]); tap's suites and parent tests, then its own; junit's
	// classname then name.
	Path []string
	// Unit is how many leading parts of Path name the smallest unit the form's runner can run alone: len(Path) for
	// go-test-json and junit; for tap, len(Path) when every ancestor is a suite (describe), else the parts up to and
	// including the outermost ancestor that is a test (node cannot run a t.test subtest without its enclosing test).
	Unit int
	// Outcome is Passed, Failed, Skipped or Todo.
	Outcome string
	// Reason is a skip's or todo's reason as the test printed it ("real-checkouts-only: needs a checkout"), "" for
	// none.
	Reason string
	// File is the test's file, project-relative with forward slashes when it lies inside root ("good/a_test.go"),
	// else as printed (with forward slashes); "" when the output names none.
	File string
	// Error is a failure's message: its first line, and the next when the first ends in ':' ("want 1, got 2: more
	// detail"), at most 300 characters (a longer one cut at a character's end and ended "..."); "" for none.
	Error string
	// Parent is true when the test failed only because a subtest failed: go-test-json, a failing subtest and no
	// message of its own; tap, failureType 'subtestsFailed'.
	Parent bool
}

// Broken is a failure outside any test: a Go package that did not build, or whose test binary failed outside a test
// (TestMain, a panic in init); a JUnit output that is not a report.
type Broken struct {
	What  string // the Go package's import path; "" for the whole output
	Build bool   // a build failure
	Error string // its first message line, at most 300 characters ("bad/b_test.go:5:28: undefined: undefined")
}

// Report is what one rung's standard output says, read in one form.
type Report struct {
	Form   string   // the form it was read in (format.TestForms)
	Tests  []Test   // in the order the output finished them
	Broken []Broken // in the order the output ended them
}

// Passed counts the tests (subtests included) whose outcome is Passed: a ratchet's count when the rung has no capture
// of the ratchet's name.
func (r Report) Passed() int {
	n := 0
	for _, t := range r.Tests {
		if t.Outcome == Passed {
			n++
		}
	}
	return n
}

// Read reads a rung's standard output in form. root is the checkout the rung ran in (absolute, system separators): a
// file is made project-relative against it, and go-test-json reads root/go.mod's module line once (to turn an import
// path into a folder). ok is false for a form Read does not know (the tests are then not read).
func Read(form string, stdout []byte, root string) (r Report, ok bool) {
	stdout = bytes.TrimPrefix(stdout, []byte("\xef\xbb\xbf"))
	switch form {
	case format.TestsGoJSON:
		return readGoJSON(stdout, root), true
	case format.TestsTAP:
		return readTAP(stdout, root), true
	case format.TestsJUnit:
		return readJUnit(stdout, root), true
	}
	return Report{Form: form}, false
}

// RunInput is what Run needs besides the test.
type RunInput struct {
	Command string // the rung's command line from bonsai.yaml (its build flags are kept: below)
	Root    string // the checkout (as Read's)
	Windows bool   // the rung runs in cmd (%ComSpec% /d /s /c), not /bin/sh: quote for cmd
}

// Run gives the command that runs t alone, as one command line quoted for the shell the rung runs in, or "" when the
// form cannot name one (junit; a tap test with no file; a test whose outcome is not Failed is still given one).
//
// go-test-json: go test -count=1 [kept] -run <pattern> <pkg>. Kept, in their order, are the -tags, -race, -short and
// -mod flags (with their values) of the rung's command's segment that runs go test. The pattern is each part of the
// test's name, regexp-quoted between ^ and $, joined by "/" (a benchmark's goes to -bench, after -run '^$'); pkg is
// ./<folder> (or .) inside root/go.mod's module, else the import path.
//
// tap: node --test --test-name-pattern=<pattern> <file>, the pattern ^ + the first Unit parts of Path, each escaped
// for a JavaScript regular expression, joined by a space, + $ (node matches a test by its ancestors' names and its
// own joined by spaces).
//
// Quoting: a word holding anything but letters, digits and -_./:=,+@ is quoted, in single quotes for /bin/sh and in
// double quotes for cmd, where the backslashes before a quote are doubled, as Windows programs read their arguments:
//
//	it's      sh: 'it'\''s'      cmd: "it's"
//	say "x"   sh: 'say "x"'     cmd: "say \"x\""
func Run(form string, t Test, in RunInput) string {
	switch form {
	case format.TestsGoJSON:
		return goRun(t, in)
	case format.TestsTAP:
		return tapRun(t, in)
	}
	return ""
}

// unit is t.Unit within 1..len(t.Path); a Unit out of that range reads as the whole path.
func unit(t Test) int {
	if t.Unit < 1 || t.Unit > len(t.Path) {
		return len(t.Path)
	}
	return t.Unit
}

// splitLines splits an output into lines, a CRLF line read as LF; the empty piece after a final newline is dropped.
func splitLines(b []byte) []string {
	s := strings.TrimSuffix(string(b), "\n")
	if s == "" {
		return nil
	}
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = strings.TrimSuffix(l, "\r")
	}
	return lines
}

// errorLine is a message's first line, and the next when the first ends in ':' (node's assert prints "...strictly
// equal:" and then "1 !== 2"), each trimmed, blank lines left out, at most maxText characters.
func errorLine(lines ...string) string {
	var l []string
	for _, x := range lines {
		for _, y := range strings.Split(x, "\n") {
			if y = strings.TrimSpace(y); y != "" {
				l = append(l, y)
			}
		}
	}
	if len(l) == 0 {
		return ""
	}
	first := l[0]
	if len(l) > 1 && strings.HasSuffix(first, ":") {
		first += " " + l[1]
	}
	return clip(first)
}

// clip keeps at most maxText characters of s: a longer text is cut at a character's end and ended with "...".
func clip(s string) string {
	if utf8.RuneCountInString(s) <= maxText {
		return s
	}
	n := 0
	for i := range s {
		if n == maxText-3 {
			return s[:i] + "..."
		}
		n++
	}
	return s
}

// projectFile makes a file the output names project-relative when it lies inside root, with forward slashes either
// way. root is absolute with the system's separators; on Windows the comparison ignores case, as the file system does.
func projectFile(file, root string) string {
	f := cleanSlashes(file)
	r := strings.TrimSuffix(cleanSlashes(root), "/")
	if f == "" || r == "" || r == "." {
		return f
	}
	if len(f) > len(r) && f[len(r)] == '/' && samePath(f[:len(r)], r) {
		return f[len(r)+1:]
	}
	return f
}

// cleanSlashes writes a path with forward slashes and no "." or ".." parts it can drop.
func cleanSlashes(p string) string {
	if p == "" {
		return ""
	}
	p = strings.ReplaceAll(p, `\`, "/")
	lead := ""
	if strings.HasPrefix(p, "//") { // a Windows share keeps its two slashes
		lead, p = "/", p[1:]
	}
	return lead + path.Clean(p)
}

// samePath compares two paths as the file system would: ignoring case on Windows.
func samePath(a, b string) bool {
	if runtime.GOOS == "windows" {
		return strings.EqualFold(a, b)
	}
	return a == b
}
