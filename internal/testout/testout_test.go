package testout

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/format"
)

// The fixtures under testdata are real outputs (each folder's go.mod or test file says how it was made), their
// home-folder paths made neutral: a checkout at /work/project.
const neutralRoot = "/work/project"

// fixture reads a file under testdata.
func fixture(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", filepath.FromSlash(name)))
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// absTestdata is a folder under testdata as an absolute path, as a rung's root is.
func absTestdata(t *testing.T, name string) string {
	t.Helper()
	dir, err := filepath.Abs(filepath.Join("testdata", filepath.FromSlash(name)))
	if err != nil {
		t.Fatal(err)
	}
	return dir
}

// read reads an output and fails the test when the form is not known.
func read(t *testing.T, form string, out []byte, root string) Report {
	t.Helper()
	r, ok := Read(form, out, root)
	if !ok {
		t.Fatalf("Read(%q) did not know the form", form)
	}
	if r.Form != form {
		t.Errorf("Form = %q, want %q", r.Form, form)
	}
	return r
}

// sameTests compares two test lists entry by entry, naming each difference.
func sameTests(t *testing.T, got, want []Test) {
	t.Helper()
	for i := 0; i < max(len(got), len(want)); i++ {
		switch {
		case i >= len(got):
			t.Errorf("test %d missing: %+v", i, want[i])
		case i >= len(want):
			t.Errorf("test %d not wanted: %+v", i, got[i])
		case !reflect.DeepEqual(got[i], want[i]):
			t.Errorf("test %d:\n got %+v\nwant %+v", i, got[i], want[i])
		}
	}
}

// crlf is an output with every line ended CRLF, as a Windows checkout with core.autocrlf may hold a fixture.
func crlf(b []byte) []byte {
	return bytes.ReplaceAll(bytes.ReplaceAll(b, []byte("\r\n"), []byte("\n")), []byte("\n"), []byte("\r\n"))
}

// goTest builds a go-test-json Test from its package and test name.
func goTestOf(pkg, name, outcome string) Test {
	path := append([]string{pkg}, strings.Split(name, "/")...)
	return Test{Name: pkg + "." + name, Path: path, Unit: len(path), Outcome: outcome}
}

func TestReadGoJSON(t *testing.T) {
	const good = "example.com/demo/good"
	with := func(tt Test, f func(*Test)) Test { f(&tt); return tt }
	want := []Test{
		goTestOf("example.com/demo", "TestRoot", Passed),
		with(goTestOf("example.com/demo/exits", "TestExits", Failed), func(t *Test) { t.Error = unendedError }),
		goTestOf(good, "TestPass", Passed),
		with(goTestOf(good, "TestFail", Failed), func(t *Test) {
			t.File, t.Error = "good/a_test.go", "want 1, got 2: more detail"
		}),
		with(goTestOf(good, "TestSkipMarked", Skipped), func(t *Test) {
			t.File, t.Reason = "good/a_test.go", "real-checkouts-only: needs a checkout"
		}),
		with(goTestOf(good, "TestSkipPlain", Skipped), func(t *Test) {
			t.File, t.Reason = "good/a_test.go", "not on this machine"
		}),
		goTestOf(good, "TestParent/sub_one", Passed),
		with(goTestOf(good, "TestParent/sub/two", Failed), func(t *Test) {
			t.File, t.Error = "good/a_test.go", "sub failed"
		}),
		with(goTestOf(good, "TestParent/deep/leaf_(x)", Failed), func(t *Test) {
			t.File, t.Error = "good/a_test.go", "leaf failed"
		}),
		with(goTestOf(good, "TestParent/deep", Failed), func(t *Test) { t.Parent = true }),
		with(goTestOf(good, "TestParent", Failed), func(t *Test) { t.Parent = true }),
		with(goTestOf(good, "TestOwnMessage/child", Failed), func(t *Test) {
			t.File, t.Error = "good/a_test.go", "child failed"
		}),
		with(goTestOf(good, "TestOwnMessage", Failed), func(t *Test) {
			t.File, t.Error = "good/a_test.go", "parent failed too"
		}),
		with(goTestOf(good, "TestNoNewline", Failed), func(t *Test) { t.Error = "no newline then fail" }),
		goTestOf(good, "Example_hello", Passed),
		goTestOf("example.com/demo/panics", "TestBefore", Passed),
		with(goTestOf("example.com/demo/panics", "TestPanics", Failed), func(t *Test) {
			t.Error = "panic: boom [recovered, repanicked]"
		}),
		with(goTestOf("example.com/demo/slow", "TestSlow", Failed), func(t *Test) {
			t.Error = "panic: test timed out after 3s"
		}),
	}
	wantBroken := []Broken{
		{What: "example.com/demo/bad", Build: true, Error: "bad/b_test.go:6:28: undefined: undefined"},
		{What: "example.com/demo/vetfail", Build: true,
			Error: "vetfail/v_test.go:6:1: ExampleHello refers to unknown identifier: Hello"},
		{What: "example.com/demo/mainfail", Error: "setup: cannot open the fixture"},
	}
	root := absTestdata(t, "gojson")
	out := fixture(t, "gojson/out.json")
	for _, c := range []struct {
		name string
		out  []byte
	}{{"LF", out}, {"CRLF", crlf(out)}} {
		t.Run(c.name, func(t *testing.T) {
			r := read(t, format.TestsGoJSON, c.out, root)
			sameTests(t, r.Tests, want)
			if !reflect.DeepEqual(r.Broken, wantBroken) {
				t.Errorf("Broken:\n got %+v\nwant %+v", r.Broken, wantBroken)
			}
			if n := r.Passed(); n != 5 {
				t.Errorf("Passed() = %d, want 5", n)
			}
		})
	}
}

func TestReadGoJSONBenchmarks(t *testing.T) {
	// go 1.27 starts each benchmark and ends only the one that fails: the others are no tests, and not failed.
	r := read(t, format.TestsGoJSON, fixture(t, "gobench/out.json"), absTestdata(t, "gobench"))
	fails := goTestOf("example.com/bp", "BenchmarkFails", Failed)
	fails.File, fails.Error = "b_test.go", "the benchmark failed"
	sameTests(t, r.Tests, []Test{
		goTestOf("example.com/bp", "FuzzF/seed#0", Passed),
		goTestOf("example.com/bp", "FuzzF", Passed),
		fails,
	})
	if len(r.Broken) != 0 || r.Passed() != 2 {
		t.Errorf("Broken %+v, Passed() %d; want none and 2", r.Broken, r.Passed())
	}
}

func TestReadGoJSONFiles(t *testing.T) {
	// A message's file is made a folder of the module when root/go.mod's module prefixes the import path; else it is
	// the bare name; a full path (go test -fullpath) is made project-relative; another .go file than a test file
	// (a helper) names no file, its prefix still taken off.
	events := func(pkg string, lines ...string) []byte {
		var b strings.Builder
		fmt.Fprintf(&b, `{"Action":"run","Package":%q,"Test":"TestA"}`+"\n", pkg)
		for _, l := range lines {
			fmt.Fprintf(&b, `{"Action":"output","Package":%q,"Test":"TestA","Output":%q}`+"\n", pkg, l+"\n")
		}
		fmt.Fprintf(&b, `{"Action":"fail","Package":%q,"Test":"TestA"}`+"\n", pkg)
		fmt.Fprintf(&b, `{"Action":"fail","Package":%q}`+"\n", pkg)
		return []byte(b.String())
	}
	withMod := func(t *testing.T, mod string) string {
		dir := t.TempDir()
		if mod != "" {
			if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(mod), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		return dir
	}
	cases := []struct {
		name, mod, pkg, line, file, err string
	}{
		{"module prefix", "module example.com/m\n\ngo 1.25\n", "example.com/m/a/b", "    x_test.go:3: no", "a/b/x_test.go", "no"},
		{"module's own package", "module example.com/m\n", "example.com/m", "    x_test.go:3: no", "x_test.go", "no"},
		{"quoted module, a comment", "// the module\nmodule \"example.com/m\" // its path\n", "example.com/m/a",
			"    x_test.go:3: no", "a/x_test.go", "no"},
		{"another module", "module example.com/m\n", "example.com/other/a", "    x_test.go:3: no", "x_test.go", "no"},
		{"a module that only shares a prefix", "module example.com/m\n", "example.com/mm/a", "    x_test.go:3: no", "x_test.go", "no"},
		{"no go.mod", "", "example.com/m/a", "    x_test.go:3: no", "x_test.go", "no"},
		{"full path", "module example.com/m\n", "example.com/m/a", "    " + neutralRoot + "/a/x_test.go:3: no", "a/x_test.go", "no"},
		{"a helper's file", "module example.com/m\n", "example.com/m/a", "    helper.go:9: no", "", "no"},
		{"no prefix", "module example.com/m\n", "example.com/m/a", "plain words:", "", "plain words: next"},
		{"an empty message", "module example.com/m\n", "example.com/m/a", "    x_test.go:3: ", "a/x_test.go", "next"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			root := withMod(t, c.mod)
			if c.name == "full path" {
				root = neutralRoot
			}
			r := read(t, format.TestsGoJSON, events(c.pkg, c.line, "next"), root)
			if len(r.Tests) != 1 {
				t.Fatalf("tests %+v", r.Tests)
			}
			if got := r.Tests[0]; got.File != c.file || got.Error != c.err {
				t.Errorf("File %q Error %q, want %q and %q", got.File, got.Error, c.file, c.err)
			}
		})
	}
}

func TestReadGoJSONEdges(t *testing.T) {
	t.Run("not events", func(t *testing.T) {
		out := "null\n[1,2]\n{not json}\n\n   {\"Action\":\"run\",\"Package\":\"p\",\"Test\":\"TestA\"}\n" +
			`{"Action":"pass","Package":"p","Test":"TestA"}` + "\n" + `{"Action":"pass","Package":"p"}`
		r := read(t, format.TestsGoJSON, []byte(out), "")
		sameTests(t, r.Tests, []Test{goTestOf("p", "TestA", Passed)})
	})
	t.Run("output cut off", func(t *testing.T) {
		// The rung ended before the package did: a test still running is failed, with the panic it printed.
		out := `{"Action":"start","Package":"p"}
{"Action":"run","Package":"p","Test":"TestA"}
{"Action":"output","Package":"p","Test":"TestA","Output":"=== RUN   TestA\n"}
{"Action":"output","Package":"p","Test":"TestA","Output":"panic: runtime error: index out of range [3] with length 1\n"}
`
		r := read(t, format.TestsGoJSON, []byte(out), "")
		want := goTestOf("p", "TestA", Failed)
		want.Error = "panic: runtime error: index out of range [3] with length 1"
		sameTests(t, r.Tests, []Test{want})
		if len(r.Broken) != 0 {
			t.Errorf("Broken %+v", r.Broken)
		}
	})
	t.Run("a panic outside tests", func(t *testing.T) {
		out := `{"Action":"output","Package":"p","Output":"panic: init failed\n"}
{"Action":"output","Package":"p","Output":"goroutine 1 [running]:\n"}
{"Action":"output","Package":"p","Output":"FAIL\tp\t0.01s\n"}
{"Action":"fail","Package":"p"}
{"Action":"output","Package":"q","Output":"FAIL\tq\t0.01s\n","OutputType":"frame"}
{"Action":"fail","Package":"q"}
{"Action":"output","Package":"n","Output":"?   \tn\t[no test files]\n"}
{"Action":"skip","Package":"n"}
`
		r := read(t, format.TestsGoJSON, []byte(out), "")
		want := []Broken{{What: "p", Error: "panic: init failed"},
			{What: "q", Error: "the test binary failed outside any test, printing nothing"}}
		if len(r.Tests) != 0 || !reflect.DeepEqual(r.Broken, want) {
			t.Errorf("Tests %+v Broken %+v", r.Tests, r.Broken)
		}
	})
	t.Run("a build failure with no message", func(t *testing.T) {
		out := `{"ImportPath":"p [p.test]","Action":"build-output","Output":"# p [p.test]\n"}
{"Action":"fail","Package":"p","FailedBuild":"p [p.test]"}
{"Action":"fail","Package":"r","FailedBuild":"dep"}
`
		r := read(t, format.TestsGoJSON, []byte(out), "")
		want := []Broken{{What: "p", Build: true, Error: "the package did not build"},
			{What: "r", Build: true, Error: "the package did not build"}}
		if !reflect.DeepEqual(r.Broken, want) {
			t.Errorf("Broken %+v", r.Broken)
		}
	})
	t.Run("a test run twice (go test -count=2)", func(t *testing.T) {
		out := `{"Action":"run","Package":"p","Test":"TestA"}
{"Action":"pass","Package":"p","Test":"TestA"}
{"Action":"run","Package":"p","Test":"TestA"}
{"Action":"output","Package":"p","Test":"TestA","Output":"    a_test.go:3: failed the second time\n"}
{"Action":"fail","Package":"p","Test":"TestA"}
{"Action":"fail","Package":"p"}
`
		r := read(t, format.TestsGoJSON, []byte(out), "")
		second := goTestOf("p", "TestA", Failed)
		second.File, second.Error = "a_test.go", "failed the second time"
		sameTests(t, r.Tests, []Test{goTestOf("p", "TestA", Passed), second})
		if len(r.Broken) != 0 {
			t.Errorf("Broken %+v", r.Broken)
		}
	})
	t.Run("a parent with its own message is not a parent", func(t *testing.T) {
		out := `{"Action":"run","Package":"p","Test":"TestA"}
{"Action":"run","Package":"p","Test":"TestA/b"}
{"Action":"output","Package":"p","Test":"TestA/b","Output":"    a_test.go:3: b failed\n"}
{"Action":"fail","Package":"p","Test":"TestA/b"}
{"Action":"output","Package":"p","Test":"TestA","Output":"    a_test.go:4: set up\n"}
{"Action":"fail","Package":"p","Test":"TestA"}
{"Action":"run","Package":"p","Test":"TestAB"}
{"Action":"fail","Package":"p","Test":"TestAB"}
{"Action":"fail","Package":"p"}
`
		r := read(t, format.TestsGoJSON, []byte(out), "")
		if len(r.Tests) != 3 || r.Tests[1].Parent || r.Tests[1].Error != "set up" || r.Tests[2].Parent {
			t.Errorf("tests %+v", r.Tests)
		}
	})
}

func TestReadTAP(t *testing.T) {
	type tap struct {
		name      string
		unit      int
		outcome   string
		reason    string
		file, err string
		parent    bool
	}
	const a, b = "a.test.mjs", "test/b.test.mjs"
	const strict = "Expected values to be strictly equal: 1 !== 2"
	rows := []tap{
		{"outer suite > inner one", 2, Failed, "", a, strict, false},
		{"outer suite > inner two", 2, Passed, "", "", "", false},
		{"outer suite > deeper > leaf # hash", 3, Failed, "", a, "boom", false},
		{"top level > sub a", 1, Passed, "", "", "", false},
		{"top level > sub b", 1, Failed, "", a, "nope", false},
		{"top level", 1, Failed, "", a, "1 subtest failed", true},
		{"skipped", 1, Skipped, "real-checkouts-only: needs a real checkout", "", "", false},
		{"todo one", 1, Todo, "later", "", "", false},
		{"plain skip", 1, Skipped, "no reason mark", "", "", false},
		{"broken.test.mjs", 1, Failed, "", "broken.test.mjs", "test failed", false},
		{"outer (suite) > passes", 2, Passed, "", "", "", false},
		{"outer (suite) > inner [x] > fails: a.b + c?", 3, Failed, "", b, strict, false},
		{"outer (suite) > inner [x] > unrelated", 3, Failed, "", b, "unrelated failure", false},
		{"second top > sub a", 1, Passed, "", "", "", false},
		{"second top > sub b > sub b deep", 1, Failed, "", b, "deep failed", false},
		{"second top > sub b", 1, Failed, "", b, "1 subtest failed", true},
		{"second top", 1, Failed, "", b, "1 subtest failed", true},
		{"mixed > holder > child", 2, Failed, "", b, "child failed", false},
		{"mixed > holder", 2, Failed, "", b, "1 subtest failed", true},
		{"other", 1, Failed, "", b, "other failed", false},
		{"a failing todo", 1, Todo, "", b, "not yet", false},
		{"bare skip", 1, Skipped, "", "", "", false},
	}
	want := make([]Test, len(rows))
	for i, w := range rows {
		want[i] = Test{Name: w.name, Path: strings.Split(w.name, " > "), Unit: w.unit, Outcome: w.outcome,
			Reason: w.reason, File: w.file, Error: w.err, Parent: w.parent}
	}
	out := fixture(t, "tap/out.tap")
	for _, c := range []struct {
		name string
		out  []byte
	}{{"LF", out}, {"CRLF", crlf(out)}} {
		t.Run(c.name, func(t *testing.T) {
			r := read(t, format.TestsTAP, c.out, neutralRoot)
			sameTests(t, r.Tests, want)
			if len(r.Broken) != 0 || r.Passed() != 4 { // node's own count: # pass 4
				t.Errorf("Broken %+v, Passed() %d; want none and 4", r.Broken, r.Passed())
			}
		})
	}
	t.Run("another root", func(t *testing.T) {
		r := read(t, format.TestsTAP, out, "/elsewhere")
		if f := r.Tests[0].File; f != neutralRoot+"/a.test.mjs" {
			t.Errorf("File %q, want the path as printed", f)
		}
	})
}

func TestReadTAPEdges(t *testing.T) {
	cases := []struct {
		name, tap, root string
		want            []Test
	}{
		{"a file URL and a Windows path", `# Subtest: x
not ok 1 - x
  ---
  type: 'test'
  location: 'file:///work/project/sub%20dir/x.test.mjs:3:1'
  error: 'first line:\nsecond line\nthird'
  ...
# Subtest: y
not ok 2 - y
  ---
  type: 'test'
  location: 'C:\\work\\project\\y.test.mjs:9:5'
  error: "a \"quoted\" \x41\u0042 word"
  ...
`, `C:\work\project`, []Test{
			{Name: "x", Path: []string{"x"}, Unit: 1, Outcome: Failed, File: "/work/project/sub dir/x.test.mjs",
				Error: "first line: second line"},
			{Name: "y", Path: []string{"y"}, Unit: 1, Outcome: Failed, File: "y.test.mjs",
				Error: `a "quoted" AB word`},
		}},
		{"a file URL inside root", `not ok 1 - x
  ---
  location: 'file:///work/project/sub%20dir/x.test.mjs:3:1'
  ...
`, neutralRoot, []Test{
			{Name: "x", Path: []string{"x"}, Unit: 1, Outcome: Failed, File: "sub dir/x.test.mjs"},
		}},
		{"escapes and directives", `ok 1 - back\\slash \# not a directive # SKIP
ok 2 - lower # skip why: \# kept
ok 3 - name \# SKIP x
ok 4 - word # SKIPPED not a directive
not ok 5 - failing todo # todo later
`, neutralRoot, []Test{
			{Name: `back\slash # not a directive`, Path: []string{`back\slash # not a directive`}, Unit: 1, Outcome: Skipped},
			{Name: "lower", Path: []string{"lower"}, Unit: 1, Outcome: Skipped, Reason: "why: # kept"},
			{Name: "name # SKIP x", Path: []string{"name # SKIP x"}, Unit: 1, Outcome: Passed},
			{Name: "word # SKIPPED not a directive", Path: []string{"word # SKIPPED not a directive"}, Unit: 1, Outcome: Passed},
			{Name: "failing todo", Path: []string{"failing todo"}, Unit: 1, Outcome: Todo, Reason: "later"},
		}},
		{"blocks of no type (an older node): an ancestor counts as a test", `# Subtest: s
    # Subtest: t
    not ok 1 - t
      ---
      error: |-
        multi
        line
    1..1
not ok 1 - s
  ---
  failureType: 'subtestsFailed'
  ...
`, neutralRoot, []Test{
			{Name: "s > t", Path: []string{"s", "t"}, Unit: 1, Outcome: Failed, Error: "multi"},
			{Name: "s", Path: []string{"s"}, Unit: 1, Outcome: Failed, Parent: true},
		}},
		{"an error block at the output's end", `not ok 1 - z
  ---
  type: 'test'
  error: >-
    the last:
    words`, neutralRoot, []Test{
			{Name: "z", Path: []string{"z"}, Unit: 1, Outcome: Failed, Error: "the last: words"},
		}},
		{"nothing", "", neutralRoot, nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := read(t, format.TestsTAP, []byte(c.tap), c.root)
			sameTests(t, r.Tests, c.want)
		})
	}
}

func TestReadJUnit(t *testing.T) {
	want := []Test{
		{Name: "math.add.adds two numbers", Path: []string{"math.add", "adds two numbers"}, Unit: 2, Outcome: Passed,
			File: "src/math.test.ts"},
		{Name: "math.add.adds a negative", Path: []string{"math.add", "adds a negative"}, Unit: 2, Outcome: Failed,
			File: "src/math.test.ts", Error: "expected -1 to equal 1: at add (src/math.ts:3)"},
		{Name: "math.divide.by zero", Path: []string{"math.divide", "by zero"}, Unit: 2, Outcome: Failed,
			File: "src/divide.test.ts", Error: "Cannot read properties of undefined (reading 'x')"},
		{Name: "math.divide.slow path", Path: []string{"math.divide", "slow path"}, Unit: 2, Outcome: Skipped,
			File: "src/math.test.ts", Reason: "real-checkouts-only: needs a real checkout"},
		{Name: "math.add.later", Path: []string{"math.add", "later"}, Unit: 2, Outcome: Skipped,
			File: "src/math.test.ts", Reason: "not on this machine"},
		{Name: "no classname", Path: []string{"no classname"}, Unit: 1, Outcome: Passed},
		{Name: "loose.outside the checkout", Path: []string{"loose", "outside the checkout"}, Unit: 2,
			Outcome: Failed, File: "/elsewhere/loose.test.ts", Error: "boom"},
		{Name: "loose.with output", Path: []string{"loose", "with output"}, Unit: 2, Outcome: Passed},
	}
	out := fixture(t, "junit/report.xml")
	for _, c := range []struct {
		name string
		out  []byte
	}{{"LF", out}, {"CRLF", crlf(out)}, {"behind other lines", append([]byte("Running the suite\n"), out...)}} {
		t.Run(c.name, func(t *testing.T) {
			r := read(t, format.TestsJUnit, c.out, neutralRoot)
			sameTests(t, r.Tests, want)
			if len(r.Broken) != 0 || r.Passed() != 3 {
				t.Errorf("Broken %+v, Passed() %d; want none and 3", r.Broken, r.Passed())
			}
		})
	}
	t.Run("a testsuite root", func(t *testing.T) {
		r := read(t, format.TestsJUnit, []byte(`<testsuite name="s"><testcase classname="c" name="n"/></testsuite>trailing words`), "")
		sameTests(t, r.Tests, []Test{{Name: "c.n", Path: []string{"c", "n"}, Unit: 2, Outcome: Passed}})
	})
}

func TestReadJUnitNotAReport(t *testing.T) {
	cases := []struct {
		name string
		out  []byte
		why  string
	}{
		{"not XML", fixture(t, "junit/notxml.txt"), "it holds no XML element"},
		{"empty", nil, "it is empty"},
		{"blank", []byte("  \n\n"), "it is empty"},
		{"another root", []byte(`<html><body/></html>`), "its root element is <html>, not <testsuites> or <testsuite>"},
		{"cut off", []byte(`<testsuites><testsuite><testcase name="a">`), "XML syntax error on line 1: unexpected EOF"},
		{"broken XML", []byte("<testsuites>\n<testcase name=\"a\"></testsuite>"),
			"XML syntax error on line 2: element <testcase> closed by </testsuite>"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := read(t, format.TestsJUnit, c.out, neutralRoot)
			want := []Broken{{Error: "the output is not a JUnit report: " + c.why}}
			if len(r.Tests) != 0 || !reflect.DeepEqual(r.Broken, want) {
				t.Errorf("Tests %+v Broken %+v, want %+v", r.Tests, r.Broken, want)
			}
		})
	}
}

func TestReadUnknownForm(t *testing.T) {
	r, ok := Read("nunit", []byte("<x/>"), neutralRoot)
	if ok || r.Form != "nunit" || r.Tests != nil || r.Broken != nil {
		t.Errorf("Read(nunit) = %+v, %v; want an empty report and false", r, ok)
	}
	for _, w := range format.TestForms { // every form the table names is read
		if _, ok := Read(w.Word, nil, ""); !ok {
			t.Errorf("Read does not know the form %q", w.Word)
		}
	}
}

func TestErrorLineAndClip(t *testing.T) {
	long := strings.Repeat("\u00e9", 400)
	cases := []struct {
		lines []string
		want  string
	}{
		{nil, ""},
		{[]string{"", "  "}, ""},
		{[]string{"one", "two"}, "one"},
		{[]string{"  one:  ", "", "  two  ", "three"}, "one: two"},
		{[]string{"one:\ntwo"}, "one: two"},
		{[]string{"only:"}, "only:"},
		{[]string{long}, strings.Repeat("\u00e9", 297) + "..."},
		{[]string{strings.Repeat("a", 300)}, strings.Repeat("a", 300)},
		{[]string{strings.Repeat("a", 301)}, strings.Repeat("a", 297) + "..."},
	}
	for i, c := range cases {
		if got := errorLine(c.lines...); got != c.want {
			t.Errorf("case %d: errorLine(%q) = %q, want %q", i, c.lines, got, c.want)
		}
	}
}

func TestProjectFile(t *testing.T) {
	cases := []struct{ file, root, want string }{
		{"/work/project/a/b.js", "/work/project", "a/b.js"},
		{"/work/project/a/../b.js", "/work/project/", "b.js"},
		{"/work/projects/b.js", "/work/project", "/work/projects/b.js"},
		{"/work/project", "/work/project", "/work/project"},
		{`C:\work\project\a\b.js`, `C:\work\project`, "a/b.js"},
		{`D:\elsewhere\b.js`, `C:\work\project`, "D:/elsewhere/b.js"},
		{`\\server\share\b.js`, `C:\work\project`, "//server/share/b.js"},
		{"rel/b.js", "/work/project", "rel/b.js"},
		{"b.js", "", "b.js"},
		{"", "/work/project", ""},
	}
	for _, c := range cases {
		if got := projectFile(c.file, c.root); got != c.want {
			t.Errorf("projectFile(%q, %q) = %q, want %q", c.file, c.root, got, c.want)
		}
	}
}
