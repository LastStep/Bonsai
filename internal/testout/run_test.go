package testout

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/format"
)

func TestRunLines(t *testing.T) {
	gojson := absTestdata(t, "gojson")
	gobench := absTestdata(t, "gobench")
	goT := func(pkg string, parts ...string) Test {
		return Test{Path: append([]string{pkg}, parts...), Unit: len(parts) + 1, Outcome: Failed}
	}
	tapT := func(file string, unit int, path ...string) Test {
		return Test{Path: path, Unit: unit, File: file, Outcome: Failed}
	}
	const good = "example.com/demo/good"
	cases := []struct {
		name      string
		form      string
		test      Test
		command   string
		root      string
		sh, cmd   string // the line for /bin/sh and for cmd
		cmdSameAs bool   // the cmd line's command is the same as sh's (both from command)
	}{
		{name: "a subtest", form: format.TestsGoJSON, test: goT(good, "TestParent", "sub", "two"),
			command: "go test -json -count=1 ./...", root: gojson,
			sh:  `go test -count=1 -run '^TestParent$/^sub$/^two$' ./good`,
			cmd: `go test -count=1 -run "^TestParent$/^sub$/^two$" ./good`},
		{name: "a passing test is given one too", form: format.TestsGoJSON,
			test:    Test{Path: []string{good, "TestPass"}, Unit: 2, Outcome: Passed},
			command: "go test -json ./...", root: gojson,
			sh:  `go test -count=1 -run '^TestPass$' ./good`,
			cmd: `go test -count=1 -run "^TestPass$" ./good`},
		{name: "a name's metacharacters", form: format.TestsGoJSON, test: goT(good, "TestParent", "deep", "leaf_(x)"),
			command: "go test -json ./...", root: gojson,
			sh:  `go test -count=1 -run '^TestParent$/^deep$/^leaf_\(x\)$' ./good`,
			cmd: `go test -count=1 -run "^TestParent$/^deep$/^leaf_\(x\)$" ./good`},
		{name: "quotes in a name", form: format.TestsGoJSON, test: goT(good, "TestA", `it's_"x"\`),
			command: "go test ./...", root: gojson,
			sh:  `go test -count=1 -run '^TestA$/^it'\''s_"x"\\$' ./good`,
			cmd: `go test -count=1 -run "^TestA$/^it's_\"x\"\\$" ./good`},
		{name: "the module's own package", form: format.TestsGoJSON, test: goT("example.com/demo", "TestRoot"),
			command: "go test -json ./...", root: gojson,
			sh: `go test -count=1 -run '^TestRoot$' .`, cmd: `go test -count=1 -run "^TestRoot$" .`},
		{name: "no go.mod: the import path", form: format.TestsGoJSON, test: goT("example.com/x/y", "TestA"),
			command: "go test -json ./...", root: "",
			sh: `go test -count=1 -run '^TestA$' example.com/x/y`, cmd: `go test -count=1 -run "^TestA$" example.com/x/y`},
		{name: "a benchmark", form: format.TestsGoJSON, test: goT("example.com/bp", "BenchmarkFails"),
			command: "go test -json -bench . ./...", root: gobench,
			sh:  `go test -count=1 -run '^$' -bench '^BenchmarkFails$' .`,
			cmd: `go test -count=1 -run "^$" -bench "^BenchmarkFails$" .`},
		{name: "kept flags, in their order", form: format.TestsGoJSON, test: goT(good, "TestA"),
			command: `cd sub && CGO_ENABLED=1 go test -json -tags 'a b' -race -short -mod=mod -v ./... -args -short; echo done`,
			root:    gojson,
			sh:      `go test -count=1 -tags 'a b' -race -short -mod=mod -run '^TestA$' ./good`,
			cmd:     `go test -count=1 -tags "'a" -race -short -mod=mod -run "^TestA$" ./good`},
		{name: "kept flags' other spellings", form: format.TestsGoJSON, test: goT(good, "TestA"),
			command: `make vet || /usr/local/go/bin/go test -tags=x,y --race -mod vendor -short=true -count=3 -run X ./...`,
			root:    gojson,
			sh:      `go test -count=1 -tags=x,y --race -mod vendor -short=true -run '^TestA$' ./good`,
			cmd:     `go test -count=1 -tags=x,y --race -mod vendor -short=true -run "^TestA$" ./good`},
		{name: "cmd's quotes and separators", form: format.TestsGoJSON, test: goT(good, "TestA"),
			command: `go.exe test -json -tags "a b" ./... & echo -race`, root: gojson,
			sh:  `go test -count=1 -tags 'a b' -run '^TestA$' ./good`,
			cmd: `go test -count=1 -tags "a b" -run "^TestA$" ./good`},
		{name: "no go test in the command", form: format.TestsGoJSON, test: goT(good, "TestA"),
			command: "make test -race", root: gojson,
			sh: `go test -count=1 -run '^TestA$' ./good`, cmd: `go test -count=1 -run "^TestA$" ./good`},
		{name: "a package alone names nothing", form: format.TestsGoJSON, test: Test{Path: []string{good}, Unit: 1},
			root: gojson},
		{name: "a test in nested suites", form: format.TestsTAP,
			test: tapT("a.test.mjs", 3, "outer suite", "deeper", "leaf # hash"),
			sh:   `node --test '--test-name-pattern=^outer suite deeper leaf # hash$' a.test.mjs`,
			cmd:  `node --test "--test-name-pattern=^outer suite deeper leaf # hash$" a.test.mjs`},
		{name: "a subtest runs with its test", form: format.TestsTAP,
			test: tapT("a.test.mjs", 1, "top level", "sub b"),
			sh:   `node --test '--test-name-pattern=^top level$' a.test.mjs`,
			cmd:  `node --test "--test-name-pattern=^top level$" a.test.mjs`},
		{name: "every escaped character", form: format.TestsTAP,
			test: tapT("sub dir/x.test.mjs", 2, `a\^$.|?*+(`, `)[]{}/-z it's`),
			sh:   `node --test '--test-name-pattern=^a\\\^\$\.\|\?\*\+\( \)\[\]\{\}\/\-z it'\''s$' 'sub dir/x.test.mjs'`,
			cmd:  `node --test "--test-name-pattern=^a\\\^\$\.\|\?\*\+\( \)\[\]\{\}\/\-z it's$" "sub dir/x.test.mjs"`},
		{name: "a tap test with no file", form: format.TestsTAP, test: tapT("", 1, "x")},
		{name: "junit names none", form: format.TestsJUnit, test: Test{Path: []string{"c", "n"}, Unit: 2, File: "x.ts"}},
		{name: "an unknown form names none", form: "nunit", test: goT(good, "TestA")},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			for _, os := range []struct {
				windows bool
				want    string
			}{{false, c.sh}, {true, c.cmd}} {
				got := Run(c.form, c.test, RunInput{Command: c.command, Root: c.root, Windows: os.windows})
				if got != os.want {
					t.Errorf("Windows %v:\n got %s\nwant %s", os.windows, got, os.want)
				}
			}
		})
	}
}

func TestQuoteWord(t *testing.T) {
	cases := []struct{ word, sh, cmd string }{
		{"plain-word_1.2/x:y=z,+@", "plain-word_1.2/x:y=z,+@", "plain-word_1.2/x:y=z,+@"},
		{"", "''", `""`},
		{"a b", "'a b'", `"a b"`},
		{"it's", `'it'\''s'`, `"it's"`},
		{`say "hi"`, `'say "hi"'`, `"say \"hi\""`},
		{`a\b`, `'a\b'`, `"a\b"`},
		{`a\"b`, `'a\"b'`, `"a\\\"b"`},
		{`trail\`, `'trail\'`, `"trail\\"`},
		{"$HOME", "'$HOME'", `"$HOME"`},
		{"~x", "'~x'", `"~x"`},
		{"\u00e9", "'\u00e9'", "\"\u00e9\""},
	}
	for _, c := range cases {
		if got := quoteWord(c.word, false); got != c.sh {
			t.Errorf("sh: quoteWord(%q) = %s, want %s", c.word, got, c.sh)
		}
		if got := quoteWord(c.word, true); got != c.cmd {
			t.Errorf("cmd: quoteWord(%q) = %s, want %s", c.word, got, c.cmd)
		}
	}
}

func TestShellSegments(t *testing.T) {
	cases := []struct {
		line    string
		windows bool
		want    [][]string
	}{
		{"a && b || c; d | e & f\ng", false, [][]string{{"a"}, {"b"}, {"c"}, {"d"}, {"e"}, {"f"}, {"g"}}},
		{`go test -tags 'a b' "c d" e\ f`, false, [][]string{{"go", "test", "-tags", "a b", "c d", "e f"}}},
		{`x "a \"q\" \$y \z" a'b'"c" "a && b"`, false, [][]string{{"x", `a "q" $y \z`, "abc", "a && b"}}},
		{`x 'unclosed`, false, [][]string{{"x", "unclosed"}}},
		{`go test -tags "a b" 'c d' x^&y "q\"r"`, true, [][]string{{"go", "test", "-tags", "a b", "'c", "d'", "x&y", `q"r`}}},
		{"   ", false, nil},
	}
	for _, c := range cases {
		if got := shellSegments(c.line, c.windows); !reflect.DeepEqual(got, c.want) {
			t.Errorf("shellSegments(%q, %v) = %q, want %q", c.line, c.windows, got, c.want)
		}
	}
}

// copyProject copies a fixture project under testdata into a new temporary folder, named by its long path (Windows'
// temporary folder may be a short 8.3 name, which a runner's output would not print).
func copyProject(t *testing.T, name string) string {
	t.Helper()
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := os.CopyFS(dir, os.DirFS(absTestdata(t, name))); err != nil {
		t.Fatal(err)
	}
	return dir
}

// runLine runs a command line in dir by the shell a rung runs in (shellCommand), and gives its standard output. A
// command whose tests fail exits non-zero, which is not an error here.
func runLine(t *testing.T, dir, line string) []byte {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	cmd := shellCommand(ctx, line)
	cmd.Dir = dir
	cmd.WaitDelay = 10 * time.Second // a process the line left holding its output does not hold the test
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	var exit *exec.ExitError
	if err != nil && (!errors.As(err, &exit) || ctx.Err() != nil) {
		t.Fatalf("%s: %v\n%s", line, err, stderr.String())
	}
	return out
}

// names lists a report's test names, in its order.
func names(r Report) []string {
	out := make([]string, len(r.Tests))
	for i, t := range r.Tests {
		out[i] = t.Name
	}
	return out
}

// find finds a test by name in a report.
func find(t *testing.T, r Report, name string) Test {
	t.Helper()
	for _, x := range r.Tests {
		if x.Name == name {
			return x
		}
	}
	t.Fatalf("no test %q in %q", name, names(r))
	return Test{}
}

func TestRunGoForReal(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go is not on the PATH, so the go test lines Run gives cannot be run here")
	}
	fixtureReport := read(t, format.TestsGoJSON, fixture(t, "gojson/out.json"), absTestdata(t, "gojson"))
	dir := copyProject(t, "gojson")
	in := RunInput{Command: "go test -json -short ./...", Root: dir, Windows: runtime.GOOS == "windows"}
	const good = "example.com/demo/good."
	cases := []struct {
		test string
		ran  []string // every test the line runs, in the order it ends them
	}{
		{good + "TestParent/sub/two", []string{good + "TestParent/sub/two", good + "TestParent"}},
		{good + "TestParent/deep/leaf_(x)",
			[]string{good + "TestParent/deep/leaf_(x)", good + "TestParent/deep", good + "TestParent"}},
	}
	for _, c := range cases {
		t.Run(c.test, func(t *testing.T) {
			want := find(t, fixtureReport, c.test)
			line := Run(format.TestsGoJSON, want, in)
			rest, ok := strings.CutPrefix(line, "go test ")
			if !ok {
				t.Fatalf("Run gave %q", line)
			}
			r := read(t, format.TestsGoJSON, runLine(t, dir, "go test -json "+rest), dir)
			if got := names(r); !reflect.DeepEqual(got, c.ran) {
				t.Errorf("%s ran %q, want %q", line, got, c.ran)
			}
			if got := find(t, r, c.test); !reflect.DeepEqual(got, want) {
				t.Errorf("%s:\n got %+v\nwant %+v", line, got, want)
			}
		})
	}
}

func TestRunNodeForReal(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not on the PATH, so the node lines Run gives cannot be run here")
	}
	fixtureReport := read(t, format.TestsTAP, fixture(t, "tap/out.tap"), neutralRoot)
	dir := copyProject(t, "tap")
	in := RunInput{Command: "node --test", Root: dir, Windows: runtime.GOOS == "windows"}
	cases := []struct {
		test string
		ran  []string // every test the line runs, suites left out, in the order node reports them
	}{
		{"outer suite > deeper > leaf # hash", []string{"outer suite > deeper > leaf # hash"}},
		{"outer (suite) > inner [x] > fails: a.b + c?", []string{"outer (suite) > inner [x] > fails: a.b + c?"}},
		{"top level > sub b", []string{"top level > sub a", "top level > sub b", "top level"}},
		{"second top > sub b > sub b deep",
			[]string{"second top > sub a", "second top > sub b > sub b deep", "second top > sub b", "second top"}},
		{"mixed > holder > child", []string{"mixed > holder > child", "mixed > holder"}},
	}
	for _, c := range cases {
		t.Run(c.test, func(t *testing.T) {
			want := find(t, fixtureReport, c.test)
			line := Run(format.TestsTAP, want, in)
			rest, ok := strings.CutPrefix(line, "node --test ")
			if !ok {
				t.Fatalf("Run gave %q", line)
			}
			r := read(t, format.TestsTAP, runLine(t, dir, "node --test --test-reporter=tap "+rest), dir)
			if got := names(r); !reflect.DeepEqual(got, c.ran) {
				t.Errorf("%s ran %q, want %q", line, got, c.ran)
			}
			if got := find(t, r, c.test); !reflect.DeepEqual(got, want) {
				t.Errorf("%s:\n got %+v\nwant %+v", line, got, want)
			}
		})
	}
}
