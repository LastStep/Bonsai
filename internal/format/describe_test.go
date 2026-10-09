package format

import (
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/schema"
)

// TestDescribe: check --schema's text for every format names every field the schema has, at every depth, with its
// allowed values, in ASCII; an open list's known words come from its Go table.
func TestDescribe(t *testing.T) {
	for _, f := range All {
		text := f.Describe()
		for i := 0; i < len(text); i++ {
			if c := text[i]; c > 0x7e || (c < 0x20 && c != '\n') {
				t.Fatalf("%s: byte %d is %#x, not printable ASCII", f.Name, i, c)
			}
		}
		if !strings.HasPrefix(text, f.Versioned()) || !strings.Contains(text, "\nFields:\n") {
			t.Errorf("%s: no title or field list:\n%s", f.Name, text)
		}
		for _, name := range fieldPaths(f.Schema(), "") {
			if !strings.Contains(text, "\n  "+name+"\n") {
				t.Errorf("%s: the field %s is not described", f.Name, name)
			}
		}
		for _, line := range strings.Split(text, "\n") {
			if len(line) > describeWidth && !strings.Contains(strings.TrimSpace(line), " ") {
				continue // one word longer than the width (an example) stands alone
			}
			if len(line) > describeWidth {
				t.Errorf("%s: a line of %d characters: %q", f.Name, len(line), line)
			}
		}
	}
	task := flat(MustLookup("task").Describe())
	for _, want := range []string{`one of "todo", "plan", "approved", "running", "verify", "done", "blocked", "cut" (a`,
		"Bonsai reads majors 0 and 1", "Bonsai does not write it", "type: text matching ^T-[0-9]{4,6}$"} {
		if !strings.Contains(task, want) {
			t.Errorf("bonsai.task's text does not say %q", want)
		}
	}
	saved := ErrorWords
	defer func() { ErrorWords = saved }()
	ErrorWords = []Word{{Word: "needs-yes", Means: "the command writes, and was not given --yes"}}
	for _, name := range []string{"error", "status", "check", "changes"} {
		if text := flat(MustLookup(name).Describe()); !strings.Contains(text, "needs-yes (the command writes, and was not given --yes)") {
			t.Errorf("%s: the error words' table is not printed", name)
		}
	}
	if text := MustLookup("task").Describe(); strings.Contains(text, "needs-yes") {
		t.Error("the task has no error code, but its text prints the table")
	}
}

// flat gives text with its line breaks and indents as single spaces, so a phrase wrapped across lines is found.
func flat(text string) string { return strings.Join(strings.Fields(text), " ") }

// fieldPaths names every property of a schema at every depth, as Describe names them.
func fieldPaths(s schema.Object, path string) []string {
	var out []string
	for _, p := range props(s) {
		ps, _ := p.Value.(schema.Object)
		name := p.Key
		if path != "" {
			name = path + "." + p.Key
		}
		out = append(out, name)
		if items := sub(ps, "items"); items != nil && props(items) != nil {
			out = append(out, fieldPaths(items, name+"[]")...)
		}
		if props(ps) != nil {
			out = append(out, fieldPaths(ps, name)...)
		}
	}
	return out
}
