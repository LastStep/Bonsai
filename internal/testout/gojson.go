package testout

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/LastStep/Bonsai/internal/format"
)

// go test -json (cmd/test2json): one JSON event per line. A test's events carry its package and its name (a subtest's
// levels joined by "/"); its own output events are attributed to it, Go naming each subtest's lines by the subtest.
// Since go 1.24 a build failure is build-output events under the failing ImportPath, then the package's fail event
// naming it in FailedBuild. A benchmark (go 1.27) has run and output events and, unless it fails, no end of its own:
// one that never ends is left out, not failed; one that fails is a Failed test, run alone by -bench.

// unendedError is the Error of a test that started and never ended, with no panic: line after its start.
const unendedError = "the test binary ended while the test ran (a panic, a timeout or os.Exit)"

// goEvent is one go test -json event, the fields read.
type goEvent struct {
	Action      string
	Package     string
	Test        string
	Output      string
	OutputType  string
	FailedBuild string
	ImportPath  string
}

// goPackage is one package's tests and output, as the events arrive.
type goPackage struct {
	path       string
	tests      map[string]*goTest // by the Test field; only looked up, never ranged over
	running    []*goTest          // tests in the order they started
	lines      []string           // every output line, its tests' included: for a panic's line
	own        []string           // its own output lines (no Test), framing left out
	failedTest bool               // a test of its own failed
	ended      bool
}

// goTest is one test of a package.
type goTest struct {
	name    string
	from    int      // len(lines) of its package when it started
	outcome string   // "" until it ends
	unended bool     // it ended with its package, not on its own
	msgs    []string // its message lines: its output, framing and blank lines left out
}

// goPrefix matches a test's message line, "    a_test.go:10: want 1, got 2:": the file and the message after it.
var goPrefix = sync.OnceValue(func() *regexp.Regexp {
	return regexp.MustCompile(`^[ \t]*([^ \t]+\.go):[0-9]+:(?: (.*))?$`)
})

// goReader reads one output's events.
type goReader struct {
	root      string
	module    string
	moduleSet bool
	pkgs      map[string]*goPackage // only looked up, never ranged over
	order     []*goPackage          // packages in the order they first appear
	build     map[string][]string   // build-output lines by ImportPath; only looked up
	done      []doneTest            // tests in the order they ended
	broken    []Broken
}

// doneTest is a test that ended, with its package.
type doneTest struct {
	pkg  *goPackage
	test *goTest
}

// readGoJSON reads go test -json's events (Read).
func readGoJSON(stdout []byte, root string) Report {
	rd := &goReader{root: root, pkgs: map[string]*goPackage{}, build: map[string][]string{}}
	for _, line := range splitLines(stdout) {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "{") {
			continue // not an event: another command of the rung printed it
		}
		var e goEvent
		if json.Unmarshal([]byte(line), &e) != nil {
			continue
		}
		rd.event(e)
	}
	for _, p := range rd.order {
		if !p.ended {
			rd.endRunning(p) // the output stopped before the package's end
		}
	}
	r := Report{Form: format.TestsGoJSON, Broken: rd.broken}
	for _, d := range rd.done {
		r.Tests = append(r.Tests, rd.test(d))
	}
	return r
}

// event takes one event in.
func (rd *goReader) event(e goEvent) {
	if e.Action == "build-output" {
		rd.build[e.ImportPath] = append(rd.build[e.ImportPath], outputLines(e.Output)...)
		return
	}
	if e.Package == "" {
		return
	}
	p := rd.pkgs[e.Package]
	if p == nil {
		p = &goPackage{path: e.Package, tests: map[string]*goTest{}}
		rd.pkgs[e.Package] = p
		rd.order = append(rd.order, p)
	}
	if e.Action == "output" {
		frame := e.OutputType == "frame"
		for _, l := range outputLines(e.Output) {
			p.lines = append(p.lines, l)
			if frame || isFraming(l) || strings.TrimSpace(l) == "" {
				continue
			}
			if e.Test == "" {
				p.own = append(p.own, l)
			} else if !isBenchmarkLine(e.Test, l) {
				t := p.test(e.Test)
				t.msgs = append(t.msgs, l)
			}
		}
		return
	}
	if e.Test == "" {
		switch e.Action {
		case "pass", "fail", "skip":
			rd.endPackage(p, e)
		}
		return
	}
	switch e.Action {
	case "run":
		t := p.test(e.Test)
		if t.outcome != "" { // run again (go test -count): a new run of it
			t = &goTest{name: e.Test}
			p.tests[e.Test] = t
		}
		t.from = len(p.lines)
		p.running = append(p.running, t)
	case "pass", "fail", "skip":
		t := p.test(e.Test)
		if t.outcome != "" {
			return
		}
		switch e.Action {
		case "pass":
			t.outcome = Passed
		case "fail":
			t.outcome = Failed
			p.failedTest = true
		default:
			t.outcome = Skipped
		}
		rd.done = append(rd.done, doneTest{p, t})
	}
}

// test finds or adds a package's test by name.
func (p *goPackage) test(name string) *goTest {
	t := p.tests[name]
	if t == nil {
		t = &goTest{name: name, from: len(p.lines)}
		p.tests[name] = t
	}
	return t
}

// endPackage ends a package: its tests still running failed with it, and a failure outside its tests is a Broken.
func (rd *goReader) endPackage(p *goPackage, e goEvent) {
	if p.ended {
		return
	}
	rd.endRunning(p)
	p.ended = true
	if e.Action != "fail" {
		return
	}
	switch {
	case e.FailedBuild != "":
		msg := "the package did not build"
		for _, l := range rd.build[e.FailedBuild] {
			if l = strings.TrimSpace(l); l != "" && !strings.HasPrefix(l, "#") {
				msg = l
				break
			}
		}
		rd.broken = append(rd.broken, Broken{What: p.path, Build: true, Error: clip(msg)})
	case !p.failedTest:
		msg := firstPanic(p.lines, 0)
		for i := len(p.own) - 1; i >= 0 && msg == ""; i-- {
			if l := strings.TrimSpace(p.own[i]); l != "PASS" && !strings.HasPrefix(l, "FAIL") {
				msg = l
			}
		}
		if msg == "" {
			msg = "the test binary failed outside any test, printing nothing"
		}
		rd.broken = append(rd.broken, Broken{What: p.path, Error: clip(msg)})
	}
}

// endRunning fails a package's tests that started and have not ended, in the order they started; a benchmark that
// never ended is left out (go 1.27 ends none that passes).
func (rd *goReader) endRunning(p *goPackage) {
	for _, t := range p.running {
		if t.outcome == "" && !isBenchmark(t.name) {
			t.outcome = Failed
			t.unended = true
			p.failedTest = true
			rd.done = append(rd.done, doneTest{p, t})
		}
	}
}

// test builds the Test of an ended test.
func (rd *goReader) test(d doneTest) Test {
	p, g := d.pkg, d.test
	path := append([]string{p.path}, strings.Split(g.name, "/")...)
	t := Test{Name: p.path + "." + g.name, Path: path, Unit: len(path), Outcome: g.outcome}
	var first, second string
	if len(g.msgs) > 0 {
		var file string
		file, first = rd.message(p, g.msgs[0])
		t.File = file
	}
	if len(g.msgs) > 1 {
		_, second = rd.message(p, g.msgs[1])
	}
	switch g.outcome {
	case Skipped:
		t.Reason = strings.TrimSpace(first)
	case Failed:
		if g.unended {
			t.Error = clip(firstPanic(p.lines, g.from))
			if t.Error == "" {
				t.Error = unendedError
			}
		} else {
			t.Error = errorLine(first, second)
		}
		t.Parent = len(g.msgs) == 0 && rd.failedBelow(p, g.name)
	}
	return t
}

// message takes a message line's indent and file prefix off: its text, and the file the prefix names (a _test.go
// file, made project-relative; "" for another file or none).
func (rd *goReader) message(p *goPackage, line string) (file, text string) {
	m := goPrefix().FindStringSubmatch(line)
	if m == nil {
		return "", strings.TrimSpace(line)
	}
	text = m[2]
	name := m[1]
	if !strings.HasSuffix(name, "_test.go") {
		return "", text
	}
	if strings.ContainsAny(name, `/\`) { // a full path (go test -fullpath)
		return projectFile(name, rd.root), text
	}
	if !rd.moduleSet {
		rd.module, rd.moduleSet = readModule(rd.root), true
	}
	if dir, ok := moduleFolder(p.path, rd.module); ok {
		if dir == "" {
			return name, text
		}
		return dir + "/" + name, text
	}
	return name, text
}

// failedBelow is true when a test of the package under name (name + "/...") failed.
func (rd *goReader) failedBelow(p *goPackage, name string) bool {
	for _, d := range rd.done {
		if d.pkg == p && d.test.outcome == Failed && strings.HasPrefix(d.test.name, name+"/") {
			return true
		}
	}
	return false
}

// firstPanic is the first line from lines[from:] that starts "panic:" (after its indent), or "".
func firstPanic(lines []string, from int) string {
	for _, l := range lines[min(from, len(lines)):] {
		if l = strings.TrimSpace(l); strings.HasPrefix(l, "panic:") {
			return l
		}
	}
	return ""
}

// outputLines splits one output event's text into lines; a text with no newline is one line.
func outputLines(s string) []string {
	return splitLines([]byte(s))
}

// isFraming is true for the lines go test prints around a test: === RUN, PAUSE, CONT, NAME (and the other ===
// updates test2json knows) and --- PASS, FAIL, SKIP, BENCH reports, indented or not.
func isFraming(l string) bool {
	if strings.HasPrefix(l, "=== ") {
		for _, w := range []string{"RUN", "PAUSE", "CONT", "NAME", "PASS", "FAIL", "SKIP", "ATTR", "ARTIFACTS"} {
			if strings.HasPrefix(l[4:], w+" ") || l[4:] == w {
				return true
			}
		}
		return false
	}
	l = strings.TrimLeft(l, " ")
	for _, w := range []string{"--- PASS: ", "--- FAIL: ", "--- SKIP: ", "--- BENCH: "} {
		if strings.HasPrefix(l, w) {
			return true
		}
	}
	return false
}

// isBenchmark is true for a benchmark's name, by go test's rule: its top level is "Benchmark" and a name not starting
// in lower case.
func isBenchmark(test string) bool {
	top, _, _ := strings.Cut(test, "/")
	rest, ok := strings.CutPrefix(top, "Benchmark")
	if !ok {
		return false
	}
	r, _ := utf8.DecodeRuneInString(rest)
	return rest == "" || !unicode.IsLower(r)
}

// isBenchmarkLine is true for the lines go test prints for a benchmark that are not its messages: its name, and its
// result ("BenchmarkLogs-16  10  21.70 ns/op").
func isBenchmarkLine(test, l string) bool {
	if !isBenchmark(test) {
		return false
	}
	f := strings.Fields(l)
	if len(f) == 0 {
		return false
	}
	rest, ok := strings.CutPrefix(f[0], test)
	if !ok {
		return false
	}
	return rest == "" || (rest[0] == '-' && strings.Trim(rest[1:], "0123456789") == "")
}

// readModule reads the module path from root/go.mod, or "" when it cannot.
func readModule(root string) string {
	if root == "" {
		return ""
	}
	b, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		return ""
	}
	for _, line := range splitLines(b) {
		rest, ok := strings.CutPrefix(strings.TrimSpace(line), "module")
		if !ok || (rest != "" && rest[0] != ' ' && rest[0] != '\t' && rest[0] != '"') {
			continue
		}
		rest, _, _ = strings.Cut(rest, "//")
		rest = strings.TrimSpace(rest)
		if u, err := strconv.Unquote(rest); err == nil {
			rest = u
		}
		return rest
	}
	return ""
}

// moduleFolder is an import path's folder inside module, project-relative ("" for the module's own package); ok is
// false when the module does not prefix it.
func moduleFolder(importPath, module string) (dir string, ok bool) {
	if module == "" {
		return "", false
	}
	if importPath == module {
		return "", true
	}
	if rest, found := strings.CutPrefix(importPath, module+"/"); found {
		return rest, true
	}
	return "", false
}

// goRun is the go test command line that runs t alone (Run).
func goRun(t Test, in RunInput) string {
	n := unit(t)
	if n < 2 {
		return ""
	}
	pats := make([]string, 0, n-1)
	for _, part := range t.Path[1:n] {
		pats = append(pats, "^"+regexp.QuoteMeta(part)+"$")
	}
	pkg := t.Path[0]
	if dir, ok := moduleFolder(pkg, readModule(in.Root)); ok {
		pkg = "./" + dir
		if dir == "" {
			pkg = "."
		}
	}
	words := []string{"go", "test", "-count=1"}
	words = append(words, keptGoFlags(in.Command, in.Windows)...)
	if isBenchmark(t.Path[1]) {
		words = append(words, "-run", "^$", "-bench", strings.Join(pats, "/"), pkg)
	} else {
		words = append(words, "-run", strings.Join(pats, "/"), pkg)
	}
	return quoteLine(words, in.Windows)
}

// keptGoFlags are the build flags of the command's first segment that runs go test, in their order: -tags, -race,
// -short and -mod, with a value given as the next word kept too; written with one dash or two, as the command wrote
// them. Words after -args belong to the test binary and are not read.
func keptGoFlags(command string, windows bool) []string {
	for _, words := range shellSegments(command, windows) {
		at := -1
		for i := 0; i+1 < len(words); i++ {
			if isGoCommand(words[i]) && words[i+1] == "test" {
				at = i + 2
				break
			}
		}
		if at < 0 {
			continue
		}
		var kept []string
		for i := at; i < len(words); i++ {
			w := words[i]
			if w == "-args" || w == "--args" {
				break
			}
			name, ok := strings.CutPrefix(w, "-")
			if !ok {
				continue
			}
			name = strings.TrimPrefix(name, "-")
			name, _, hasValue := strings.Cut(name, "=")
			switch name {
			case "race", "short":
				kept = append(kept, w)
			case "tags", "mod":
				kept = append(kept, w)
				if !hasValue && i+1 < len(words) {
					kept = append(kept, words[i+1])
					i++
				}
			}
		}
		return kept
	}
	return nil
}

// isGoCommand is true for the go command's word: go, go.exe or a path ending in either.
func isGoCommand(w string) bool {
	w = w[strings.LastIndexAny(w, `/\`)+1:]
	return strings.EqualFold(w, "go") || strings.EqualFold(w, "go.exe")
}
