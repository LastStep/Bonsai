package reader

// The formats set's test of this reader (plan part 2, check 7's Bonsai half): every case in formats/expect.json,
// read as raw bytes from formats/ (so the CRLF and BOM cases count), reaches its format-1 outcome: the same
// outcome, the same value (key order included) or the same reason code. And the reason codes this reader reports
// are formats/README.md's table, in its order, plus the two of its own (OwnCodes).

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
	if reached != len(cases) || len(cases) != 102 {
		t.Errorf("reached %d of %d cases; the set has 102", reached, len(cases))
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

// The reason codes have one home, formats/README.md's table; Codes holds them in its order, with this reader's own
// two at fixed places. A code added to the table, or reordered, fails here until Codes follows.
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
	own := map[string]bool{}
	for _, c := range OwnCodes {
		own[c] = true
		for _, x := range table {
			if x == c {
				t.Errorf("%s is in README.md's table now: drop it from OwnCodes", c)
			}
		}
	}
	var ours []string
	for _, c := range Codes {
		if !own[c] {
			ours = append(ours, c)
		}
		if next[c] == "" {
			t.Errorf("code %s names no next step", c)
		}
	}
	if strings.Join(ours, " ") != strings.Join(table, " ") {
		t.Errorf("Codes (less OwnCodes) is\n  %v\nREADME.md's table is\n  %v", ours, table)
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
