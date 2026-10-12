// Command tojunit is a fixture of the new-tests check: it reads go test -json on its standard input and prints a JUnit
// XML report, one testcase per test with its package as classname and the _test.go file that declares its top-level
// test as file (as go-junit-report and gotestsum do). A package that does not build prints no testcase. It exits 1 when
// a test failed or a package did not build. With -nofile, no testcase names a file.
//
//	go test -json ./... | go run ./tojunit
package main

import (
	"bufio"
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type event struct {
	Action      string
	Package     string
	Test        string
	FailedBuild string
}

func main() {
	noFile := flag.Bool("nofile", false, "name no file")
	flag.Parse()
	module := moduleOf("go.mod")
	failed := false
	var cases []string
	in := bufio.NewScanner(os.Stdin)
	in.Buffer(make([]byte, 1<<20), 1<<20)
	for in.Scan() {
		var e event
		if json.Unmarshal(in.Bytes(), &e) != nil {
			continue
		}
		if e.Test == "" {
			if e.Action == "fail" {
				failed = true
			}
			continue
		}
		var body string
		switch e.Action {
		case "pass":
		case "fail":
			failed = true
			body = `<failure message="failed"></failure>`
		case "skip":
			body = `<skipped message="skipped"></skipped>`
		default:
			continue
		}
		file := ""
		if !*noFile {
			file = fileOf(module, e.Package, e.Test)
		}
		cases = append(cases, fmt.Sprintf(`  <testcase classname="%s" name="%s"%s>%s</testcase>`, esc(e.Package), esc(e.Test),
			fileAttr(file), body))
	}
	fmt.Println(`<?xml version="1.0" encoding="UTF-8"?>`)
	fmt.Println("<testsuites>")
	for _, c := range cases {
		fmt.Println(c)
	}
	fmt.Println("</testsuites>")
	if failed {
		os.Exit(1)
	}
}

// moduleOf reads go.mod's module line.
func moduleOf(name string) string {
	raw, _ := os.ReadFile(name)
	for _, l := range strings.Split(string(raw), "\n") {
		if v, ok := strings.CutPrefix(strings.TrimSpace(l), "module "); ok {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

// fileOf finds the _test.go file of the package's folder that declares the test's top-level function.
func fileOf(module, pkg, test string) string {
	dir := strings.TrimPrefix(strings.TrimPrefix(pkg, module), "/")
	if dir == "" {
		dir = "."
	}
	top, _, _ := strings.Cut(test, "/")
	decl := regexp.MustCompile(`(?m)^func ` + regexp.QuoteMeta(top) + `\(`)
	files, _ := filepath.Glob(filepath.Join(dir, "*_test.go"))
	for _, f := range files {
		raw, err := os.ReadFile(f)
		if err == nil && decl.Match(raw) {
			return filepath.ToSlash(f)
		}
	}
	return ""
}

func fileAttr(file string) string {
	if file == "" {
		return ""
	}
	return ` file="` + esc(file) + `"`
}

func esc(s string) string {
	var b strings.Builder
	_ = xml.EscapeText(&b, []byte(s))
	return b.String()
}
