package selfcheck

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// The fault-tag pass (bonsai.yaml's rungs 1 and 4, ci.yml's tagged test, vet and lint steps) runs only the packages
// the tag bonsai_test_fault can change: those holding a file whose file set the tag changes, and those that reach
// such a package in their normal or test build. Every other package compiles the same with and without the tag, so
// its tagged run would repeat the plain pass. The list is written into each tagged command; this test works out the
// set from go list and fails on any copy that differs from it, so a new importer of internal/guard, a new tagged file
// or a stale copy cannot drop out of the tagged pass unseen.

// faultTag is the build tag the guard's test faults are compiled under.
const faultTag = "bonsai_test_fault"

// faultCopies are the files that hold a copy of the list, each with how many commands in it name the tag.
var faultCopies = []struct {
	file string
	want int
}{
	{"bonsai.yaml", 2},              // rung 1 (vet) and rung 4 (tests)
	{".github/workflows/ci.yml", 5}, // the test job's test and vet, the windows (fault tag) job's test and vet, lint
}

// goListLine is the go list -test template: the package's import path, its files, and its dependencies.
const goListLine = "{{.ImportPath}}|{{join .GoFiles \",\"}};{{join .CgoFiles \",\"}};{{join .TestGoFiles \",\"}};" +
	"{{join .XTestGoFiles \",\"}}|{{join .Deps \",\"}}"

// goList runs go list -test over the module at root, with extra flags, and returns its lines.
func goList(t *testing.T, root string, flags ...string) []string {
	t.Helper()
	args := append([]string{"list"}, flags...)
	args = append(args, "-test", "-f", goListLine, "./...")
	cmd := exec.Command("go", args...)
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "GOFLAGS=")
	out, err := cmd.Output()
	if err != nil {
		msg := ""
		if ee, ok := err.(*exec.ExitError); ok {
			msg = string(ee.Stderr)
		}
		t.Fatalf("go %s: %v\n%s", strings.Join(args, " "), err, msg)
	}
	return strings.Split(strings.TrimSpace(strings.ReplaceAll(string(out), "\r\n", "\n")), "\n")
}

// basePackage is the package a go list -test entry belongs to: "p", "p [p.test]", "p_test [p.test]" and "p.test" are
// all p.
func basePackage(entry string) string {
	if i := strings.Index(entry, " ["); i >= 0 {
		entry = entry[:i]
	}
	entry = strings.TrimSuffix(entry, ".test")
	return strings.TrimSuffix(entry, "_test")
}

// faultSet works out, from go list -test's lines without and with the tag, the packages the tag can change: each
// package one of whose entries has a different file set in the two runs (or is in one run only), and each package one
// of whose entries depends on such a package in either run. It returns their import paths, sorted.
func faultSet(plain, tagged []string) []string {
	type entry struct{ files, deps string }
	read := func(lines []string) map[string]entry {
		m := map[string]entry{}
		for _, l := range lines {
			parts := strings.SplitN(l, "|", 3)
			for len(parts) < 3 {
				parts = append(parts, "")
			}
			m[parts[0]] = entry{parts[1], parts[2]}
		}
		return m
	}
	p, g := read(plain), read(tagged)
	changed := map[string]bool{}
	for name, e := range p {
		if o, ok := g[name]; !ok || o.files != e.files {
			changed[basePackage(name)] = true
		}
	}
	for name := range g {
		if _, ok := p[name]; !ok {
			changed[basePackage(name)] = true
		}
	}
	set := map[string]bool{}
	for c := range changed {
		set[c] = true
	}
	for _, m := range []map[string]entry{p, g} {
		for name, e := range m {
			if e.deps == "" {
				continue
			}
			for _, d := range strings.Split(e.deps, ",") {
				if changed[basePackage(d)] {
					set[basePackage(name)] = true
					break
				}
			}
		}
	}
	out := make([]string, 0, len(set))
	for s := range set {
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}

// relPackages writes import paths in the module as the ./ paths a go command takes, in the same order.
func relPackages(module string, paths []string) []string {
	out := make([]string, 0, len(paths))
	for _, p := range paths {
		if p == module {
			out = append(out, ".")
			continue
		}
		out = append(out, "./"+strings.TrimPrefix(p, module+"/"))
	}
	return out
}

// faultCopy is one command naming the tag: where it is and the packages it names after the tag.
type faultCopy struct {
	where    string
	packages []string
}

// copiesIn returns every command in content, a file named name, that names the tag, with the packages it names after
// -tags bonsai_test_fault or --build-tags=bonsai_test_fault: the words up to a closing quote, a && or the line's end.
// A comment line is not a command.
func copiesIn(name, content string) []faultCopy {
	var out []faultCopy
	for i, line := range strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "#") || !strings.Contains(line, faultTag) {
			continue
		}
		words := strings.Fields(line)
		c := faultCopy{where: fmt.Sprintf("%s:%d", name, i+1)}
		for j := 0; j < len(words); j++ {
			w := strings.Trim(words[j], `"`)
			var rest []string
			switch {
			case w == "-tags" && j+1 < len(words) && strings.Trim(words[j+1], `"`) == faultTag:
				rest = words[j+2:]
			case w == "--build-tags="+faultTag || w == "-tags="+faultTag:
				rest = words[j+1:]
			default:
				continue
			}
			for _, r := range rest {
				end := strings.HasSuffix(r, `"`)
				r = strings.TrimSuffix(r, `"`)
				if r == "&&" || r == "#" {
					break
				}
				if r != "" {
					c.packages = append(c.packages, r)
				}
				if end {
					break
				}
			}
			break
		}
		out = append(out, c)
	}
	return out
}

// checkCopies returns one problem per copy of the list that differs from want, and one per file whose number of
// copies is not the number expected, each naming the file (and line).
func checkCopies(want []string, files map[string]string) []string {
	var problems []string
	for _, fc := range faultCopies {
		copies := copiesIn(fc.file, files[fc.file])
		if len(copies) != fc.want {
			problems = append(problems, fmt.Sprintf("%s: %d commands name %s, %d expected (rungs 1 and 4 in bonsai.yaml; "+
				"test, vet and lint in ci.yml); a copy was dropped or added, update faultCopies with it", fc.file,
				len(copies), faultTag, fc.want))
		}
		for _, c := range copies {
			if strings.Join(c.packages, " ") != strings.Join(want, " ") {
				problems = append(problems, fmt.Sprintf("%s: the tagged command names %q; the packages the tag can "+
					"change are %q", c.where, strings.Join(c.packages, " "), strings.Join(want, " ")))
			}
		}
	}
	return problems
}

// moduleRoot is the module's folder, two up from this package's.
var moduleRoot = filepath.Join("..", "..")

// modulePath reads the module line of go.mod at root.
func modulePath(t *testing.T, root string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	for _, l := range strings.Split(strings.ReplaceAll(string(b), "\r\n", "\n"), "\n") {
		if f := strings.Fields(l); len(f) == 2 && f[0] == "module" {
			return f[1]
		}
	}
	t.Fatal("go.mod names no module")
	return ""
}

// TestFaultTagPackageLists holds every tagged command in bonsai.yaml and ci.yml to the packages the tag can change,
// worked out from go list -test without and with the tag.
func TestFaultTagPackageLists(t *testing.T) {
	plain := goList(t, moduleRoot)
	tagged := goList(t, moduleRoot, "-tags="+faultTag)
	want := relPackages(modulePath(t, moduleRoot), faultSet(plain, tagged))
	if len(want) == 0 {
		t.Fatalf("go list finds no package the tag %s changes; the guard's fault files are expected", faultTag)
	}
	files := map[string]string{}
	for _, fc := range faultCopies {
		files[fc.file] = read(t, fc.file)
	}
	for _, p := range checkCopies(want, files) {
		t.Error(p)
	}
}

// TestFaultTagCheckFailsOnBadCopies shows the set and the copy check failing where they must: a package that only
// imports the changed one is in the set, and a stale, short, long, reordered or missing copy is a problem.
func TestFaultTagCheckFailsOnBadCopies(t *testing.T) {
	plain := []string{
		"m/g|off.go;;;|",
		"m/g [m/g.test]|off.go;;off_test.go;|",
		"m/g.test||m/g [m/g.test],testing",
		"m/a|a.go;;;|m/g",
		"m/b|b.go;;;|",
		"m/b [m/b.test]|b.go;;b_test.go;|m/g",
		"m/c|c.go;;;|m/b",
		"m/d|d.go;;;|fmt",
	}
	tagged := []string{
		"m/g|on.go;;;|",
		"m/g [m/g.test]|on.go;;on_test.go;|",
		"m/g.test||m/g [m/g.test],testing",
		"m/a|a.go;;;|m/g",
		"m/b|b.go;;;|",
		"m/b [m/b.test]|b.go;;b_test.go;|m/g",
		"m/c|c.go;;;|m/b",
		"m/d|d.go;;;|fmt",
	}
	got := strings.Join(relPackages("m", faultSet(plain, tagged)), " ")
	if got != "./a ./b ./g" {
		t.Fatalf("faultSet = %q, want \"./a ./b ./g\" (g's files change; a imports it; b's test does; c and d do not)", got)
	}
	want := []string{"./a", "./b", "./g"}
	yaml := "    command: \"go vet ./... && go vet -tags bonsai_test_fault ./a ./b ./g\"   # vet\n" +
		"    command: \"go test -json -count=1 -tags bonsai_test_fault ./a ./b ./g\"   # tests\n"
	ci := "      # a comment naming bonsai_test_fault is no command\n" +
		"        run: go test -timeout 20m -tags bonsai_test_fault ./a ./b ./g\n" +
		"        run: go vet -tags bonsai_test_fault ./a ./b ./g\n" +
		"        run: go test -timeout 20m -tags bonsai_test_fault ./a ./b ./g\n" +
		"        run: go vet -tags bonsai_test_fault ./a ./b ./g\n" +
		"          args: --build-tags=bonsai_test_fault ./a ./b ./g\n"
	if p := checkCopies(want, map[string]string{"bonsai.yaml": yaml, ".github/workflows/ci.yml": ci}); len(p) != 0 {
		t.Fatalf("good copies have problems: %v", p)
	}
	cases := []struct{ name, file, from, to, wantProblem string }{
		{"stale ./... in a rung", "bonsai.yaml", "-count=1 -tags bonsai_test_fault ./a ./b ./g", "-count=1 -tags bonsai_test_fault ./...", `bonsai.yaml:2: the tagged command names "./..."`},
		{"one missing", ".github/workflows/ci.yml", "go vet -tags bonsai_test_fault ./a ./b ./g\n        run: go test", "go vet -tags bonsai_test_fault ./a ./g\n        run: go test", `ci.yml:3: the tagged command names "./a ./g"`},
		{"one extra", ".github/workflows/ci.yml", "--build-tags=bonsai_test_fault ./a ./b ./g", "--build-tags=bonsai_test_fault ./a ./b ./c ./g", `ci.yml:6: the tagged command names "./a ./b ./c ./g"`},
		{"reordered", "bonsai.yaml", "go vet -tags bonsai_test_fault ./a ./b ./g", "go vet -tags bonsai_test_fault ./b ./a ./g", `bonsai.yaml:1: the tagged command names "./b ./a ./g"`},
		{"a copy dropped", ".github/workflows/ci.yml", "        run: go vet -tags bonsai_test_fault ./a ./b ./g\n", "", "ci.yml: 4 commands name bonsai_test_fault, 5 expected"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			files := map[string]string{"bonsai.yaml": yaml, ".github/workflows/ci.yml": ci}
			if !strings.Contains(files[c.file], c.from) {
				t.Fatalf("the fixture holds no %q", c.from)
			}
			files[c.file] = strings.Replace(files[c.file], c.from, c.to, 1)
			p := strings.Join(checkCopies(want, files), "\n")
			if !strings.Contains(p, c.wantProblem) {
				t.Errorf("problems %q do not hold %q", p, c.wantProblem)
			}
		})
	}
}
