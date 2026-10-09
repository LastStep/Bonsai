package reader

// The formats set's test of this reader (plan part 2, check 7's Bonsai half): every case in formats/expect.json,
// read as raw bytes from formats/ (so the CRLF and BOM cases count), reaches its format-1 outcome: the same
// outcome, the same value (key order included) or the same reason code. And the reason codes this reader reports
// are formats/README.md's table, in its order, and nothing else.

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/schema"
)

// setDir is the formats set, from this package's folder.
var setDir = filepath.Join("..", "..", "formats")

// setCases is how many cases the set this reader is held to has (set 4); a case dropped or added fails the test until
// this follows.
const setCases = 125

// readCase reads one input file of the set the way its extension says.
func readCase(raw []byte, path string) Result {
	if strings.HasSuffix(path, ".md") {
		return ReadMarkdown(raw)
	}
	return ReadYAML(raw)
}

func loadExpect(t *testing.T) []schema.Object {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(setDir, "expect.json"))
	if err != nil {
		t.Fatal(err)
	}
	v, err := schema.Decode(raw)
	if err != nil {
		t.Fatalf("expect.json: %v", err)
	}
	casesV, _ := v.(schema.Object).Get("cases")
	var out []schema.Object
	for _, c := range casesV.([]any) {
		out = append(out, c.(schema.Object))
	}
	return out
}

func TestEveryCaseReachesItsFormat1Outcome(t *testing.T) {
	cases := loadExpect(t)
	reached := 0
	for _, c := range cases {
		path := c.String("path")
		f1v, _ := c.Get("format1")
		want := f1v.(schema.Object)
		raw, err := os.ReadFile(filepath.Join(setDir, filepath.FromSlash(path)))
		if err != nil {
			t.Errorf("%s: %v", path, err)
			continue
		}
		got := readCase(raw, path)
		if got.Outcome.String() != want.String("outcome") {
			t.Errorf("%s: outcome %s, want %s (%v)", path, got.Outcome, want.String("outcome"), got.Err())
			continue
		}
		switch got.Outcome {
		case Accepted:
			wantV, _ := want.Get("value")
			gotV := JSON(got.Value)
			if !schema.Equal(gotV, wantV) || !sameOrder(gotV, wantV) {
				t.Errorf("%s: value\n  %s\nwant\n  %s", path, schema.Show(gotV), schema.Show(wantV))
				continue
			}
		case Refused:
			if got.Refusal.Code != want.String("reason") {
				t.Errorf("%s: refused for %s, want %s (%v)", path, got.Refusal.Code, want.String("reason"), got.Refusal)
				continue
			}
		}
		reached++
	}
	t.Logf("format-1 outcomes reached: %d of %d", reached, len(cases))
	if reached != len(cases) || len(cases) != setCases {
		t.Errorf("reached %d of %d cases; set 4 has %d", reached, len(cases), setCases)
	}
}

// sameOrder holds two equal JSON values to the same key order at every depth (schema.Equal ignores order).
func sameOrder(a, b any) bool {
	switch x := a.(type) {
	case schema.Object:
		y := b.(schema.Object)
		for i := range x {
			if x[i].Key != y[i].Key || !sameOrder(x[i].Value, y[i].Value) {
				return false
			}
		}
	case []any:
		y := b.([]any)
		for i := range x {
			if !sameOrder(x[i], y[i]) {
				return false
			}
		}
	}
	return true
}

// The reason codes have one home, formats/README.md's table; Codes holds exactly its codes, in its order. A code
// added to the table, removed or reordered fails here until Codes follows, and the reader has no code of its own.
func TestCodesAreTheSetsTableInOrder(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join(setDir, "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	text := strings.ReplaceAll(string(raw), "\r\n", "\n")
	start := strings.Index(text, "\n## Reason codes\n")
	if start < 0 {
		t.Fatal("README.md has no Reason codes section")
	}
	section := text[start+1:]
	if end := strings.Index(section[3:], "\n## "); end >= 0 {
		section = section[:end+3]
	}
	var table []string
	for _, m := range regexp.MustCompile("(?m)^\\| `([a-z0-9-]+)` \\|").FindAllStringSubmatch(section, -1) {
		table = append(table, m[1])
	}
	if strings.Join(Codes, " ") != strings.Join(table, " ") {
		t.Errorf("Codes is\n  %v\nREADME.md's table is\n  %v", Codes, table)
	}
	for _, c := range Codes {
		if next[c] == "" {
			t.Errorf("code %s names no next step", c)
		}
	}
	if len(next) != len(Codes) {
		t.Errorf("next has %d entries, Codes %d", len(next), len(Codes))
	}
}

// The reader's label-key pattern is contract §5.1's, whose one home in the set is labels.schema.json.
func TestLabelKeyPatternIsTheSchemas(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join(setDir, "schemas", "labels.schema.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"pattern": "`+strings.ReplaceAll(labelPattern.String(), `\`, `\\`)+`"`) {
		t.Errorf("labels.schema.json holds no pattern %s", labelPattern)
	}
}

// Contract §13's fixtures (formats/active-task, formats/README.md): every bonsai.yaml and task file reads as its case
// says. A file case.json lists under does_not_parse is refused by its format; every other one is accepted under
// format 1, or is format 0 and accepted by the format-0 reader. So the fixtures' answers rest on files that are what
// they claim.
func TestActiveTaskFixturesRead(t *testing.T) {
	dir := filepath.Join(setDir, "active-task")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	read, refused, format0 := 0, 0, 0
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(dir, e.Name(), "case.json"))
		if err != nil {
			t.Fatal(err)
		}
		v, err := schema.Decode(raw)
		if err != nil {
			t.Fatalf("%s/case.json: %v", e.Name(), err)
		}
		broken := map[string]bool{}
		list, _ := v.(schema.Object).Get("does_not_parse")
		for _, p := range list.([]any) {
			broken[p.(string)] = true
		}
		root := filepath.Join(dir, e.Name())
		err = filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() || (!strings.HasSuffix(p, ".md") && !strings.HasSuffix(p, ".yaml")) {
				return err
			}
			rel, _ := filepath.Rel(root, p)
			rel = filepath.ToSlash(rel)
			raw, err := os.ReadFile(p)
			if err != nil {
				return err
			}
			got := readCase(raw, rel)
			switch {
			case broken[rel]:
				if got.Outcome != Refused {
					t.Errorf("%s/%s: %s, but case.json says it does not parse", e.Name(), rel, got.Outcome)
				}
				refused++
			case got.Outcome == Format0:
				if r0 := readCase0(raw, rel); r0.Outcome != Accepted {
					t.Errorf("%s/%s: format 0, and the format-0 reader refuses it: %v", e.Name(), rel, r0.Err())
				}
				format0++
			case got.Outcome != Accepted:
				t.Errorf("%s/%s: %v", e.Name(), rel, got.Err())
			default:
				// A format-1 file is valid under its schema: bonsai.yaml under the workspace's, a task under the task's.
				name := "task"
				if strings.HasSuffix(rel, "bonsai.yaml") {
					name = "workspace"
				}
				s, err := schema.Parse(mustSchema(t, name))
				if err != nil {
					t.Fatal(err)
				}
				for _, msg := range schema.Validate(s, JSON(got.Value)) {
					t.Errorf("%s/%s: not a valid bonsai.%s/1: %s", e.Name(), rel, name, msg)
				}
			}
			read++
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("fixture files read: %d (%d refused as their case says, %d format 0)", read, refused, format0)
	if read == 0 || refused == 0 || format0 == 0 {
		t.Errorf("the fixtures hold %d files, %d refused and %d format 0: each kind is expected", read, refused, format0)
	}
}

// mustSchema reads one of the set's schemas, by its short name.
func mustSchema(t *testing.T, name string) []byte {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(setDir, "schemas", name+".schema.json"))
	if err != nil {
		t.Fatal(err)
	}
	return raw
}
