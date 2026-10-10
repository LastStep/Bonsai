package format

import (
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/reader"
)

// WriterProblems (step 5.1.9) holds a YAML file to its format as a writer writes it, listing every problem with its
// field and line: a missing field (a reader would read it as null), a wrong type, a value outside a closed list, a
// field out of order. The format's own example has none.
func TestWriterProblems(t *testing.T) {
	read := func(src string) *reader.Map {
		t.Helper()
		r := reader.ReadYAML([]byte(src))
		if r.Outcome != reader.Accepted {
			t.Fatalf("the reader does not accept:\n%s", src)
		}
		return r.Value
	}
	lanes := MustLookup("lanes")
	if ps := lanes.WriterProblems(read("format: bonsai.lanes/1\nlanes:\n  - name: light\n    approve_first: false\n    close: agent\n    description: \"x\"\n")); len(ps) != 0 {
		t.Errorf("a lanes file as a writer writes it: %+v", ps)
	}
	ps := lanes.WriterProblems(read("format: bonsai.lanes/1\nlanes:\n  - approve_first: 1\n    name: light\n    close: nobody\n"))
	var got []string
	for _, p := range ps {
		got = append(got, p.Field+" line "+itoa(p.Line)+": "+p.Msg)
	}
	want := []string{
		`lanes[0] line 2: required field "description" is missing`, // a list item has no line of its own: its list's
		"lanes[0].approve_first line 3: is integer, want boolean",
		`lanes[0].close line 5: "nobody" is not one of ["person","agent"]`,
		"lanes[0].name line 4: out of the schema's field order",
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("problems:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
	if ps := MustLookup("pack").WriterProblems(read("format: bonsai.pack/1\nid: x\n")); len(ps) == 0 || ps[0].Field != "(the document)" {
		t.Errorf("a pack.yaml with two fields: %+v", ps)
	}
}
