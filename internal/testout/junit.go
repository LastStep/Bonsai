package testout

import (
	"bytes"
	"encoding/xml"
	"errors"
	"io"
	"strings"

	"github.com/LastStep/Bonsai/internal/format"
)

// A JUnit XML report: a <testsuites> or <testsuite> root, suites nested at any depth, and <testcase> elements with
// their classname, name and file. A <failure> or <error> child makes a case Failed, a <skipped> one Skipped; any
// other case passed. A case with no file of its own takes its nearest suite's.

// junitCase is one <testcase>, the parts read.
type junitCase struct {
	Classname string         `xml:"classname,attr"`
	Name      string         `xml:"name,attr"`
	File      string         `xml:"file,attr"`
	Failures  []junitOutcome `xml:"failure"`
	Errors    []junitOutcome `xml:"error"`
	Skipped   []junitOutcome `xml:"skipped"`
}

// junitOutcome is a <failure>, <error> or <skipped>: its message attribute and its text.
type junitOutcome struct {
	Message string `xml:"message,attr"`
	Text    string `xml:",chardata"`
}

// readJUnit reads a JUnit XML report (Read): output that is not one gives no tests and one Broken saying why.
func readJUnit(stdout []byte, root string) Report {
	r := Report{Form: format.TestsJUnit}
	notReport := func(why string) Report {
		return Report{Form: format.TestsJUnit, Broken: []Broken{{Error: clip("the output is not a JUnit report: " + why)}}}
	}
	if len(bytes.TrimSpace(stdout)) == 0 {
		return notReport("it is empty")
	}
	d := xml.NewDecoder(bytes.NewReader(stdout))
	d.CharsetReader = func(_ string, in io.Reader) (io.Reader, error) { return in, nil }
	var files []string // each open suite's file, innermost last
	rootSeen := false
	for {
		tok, err := d.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return notReport(err.Error())
		}
		switch el := tok.(type) {
		case xml.StartElement:
			name := el.Name.Local
			if !rootSeen {
				if name != "testsuites" && name != "testsuite" {
					return notReport("its root element is <" + name + ">, not <testsuites> or <testsuite>")
				}
				rootSeen = true
			}
			switch name {
			case "testsuites", "testsuite":
				file := attr(el, "file")
				if file == "" && len(files) > 0 {
					file = files[len(files)-1]
				}
				files = append(files, file)
			case "testcase":
				var c junitCase
				if err := d.DecodeElement(&c, &el); err != nil {
					return notReport(err.Error())
				}
				if c.File == "" && len(files) > 0 {
					c.File = files[len(files)-1]
				}
				r.Tests = append(r.Tests, junitTest(c, root))
			default:
				if err := d.Skip(); err != nil {
					return notReport(err.Error())
				}
			}
		case xml.EndElement:
			if n := el.Name.Local; (n == "testsuites" || n == "testsuite") && len(files) > 0 {
				files = files[:len(files)-1]
				if len(files) == 0 {
					return r // the root's end: what follows is not the report's
				}
			}
		}
	}
	if !rootSeen {
		return notReport("it holds no XML element")
	}
	return notReport("it ends before its root element does")
}

// junitTest builds one case's Test.
func junitTest(c junitCase, root string) Test {
	t := Test{Name: c.Name, Path: []string{c.Name}}
	if c.Classname != "" {
		t.Name = c.Classname + "." + c.Name
		t.Path = []string{c.Classname, c.Name}
	}
	t.Unit = len(t.Path)
	if c.File != "" {
		t.File = projectFile(c.File, root)
	}
	switch {
	case len(c.Failures) > 0 || len(c.Errors) > 0:
		var o junitOutcome
		if len(c.Failures) > 0 {
			o = c.Failures[0]
		} else {
			o = c.Errors[0]
		}
		t.Outcome = Failed
		if strings.TrimSpace(o.Message) != "" {
			t.Error = errorLine(o.Message)
		} else {
			t.Error = errorLine(o.Text)
		}
	case len(c.Skipped) > 0:
		t.Outcome = Skipped
		t.Reason = strings.TrimSpace(c.Skipped[0].Message)
		if t.Reason == "" {
			t.Reason = strings.TrimSpace(c.Skipped[0].Text)
		}
	default:
		t.Outcome = Passed
	}
	return t
}

// attr is an element's attribute by its local name, or "".
func attr(el xml.StartElement, name string) string {
	for _, a := range el.Attr {
		if a.Name.Local == name {
			return a.Value
		}
	}
	return ""
}
