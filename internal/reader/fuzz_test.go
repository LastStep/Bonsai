package reader

// Fuzzing the reader: whatever the bytes, it returns one of the three outcomes, never panics, refuses only with a
// code in Codes on a line of the file, and gives a value the JSON writer takes. `go test` runs the seeds (every file
// of the formats set and the cases above); `go test -fuzz FuzzRead ./internal/reader/` explores further.

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/schema"
)

func FuzzRead(f *testing.F) {
	_ = filepath.WalkDir(setDir, func(p string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() && (strings.HasSuffix(p, ".yaml") || strings.HasSuffix(p, ".md")) {
			if raw, err := os.ReadFile(p); err == nil {
				f.Add(raw, strings.HasSuffix(p, ".md"))
			}
		}
		return nil
	})
	for _, c := range cases {
		f.Add([]byte(c.in), c.md)
	}
	known := map[string]bool{}
	for _, c := range Codes {
		known[c] = true
	}
	f.Fuzz(func(t *testing.T, raw []byte, md bool) {
		read := ReadYAML
		if md {
			read = ReadMarkdown
		}
		r := read(raw)
		lines := bytes.Count(raw, []byte("\n")) + 1
		switch r.Outcome {
		case Format0:
		case Refused:
			if r.Refusal == nil || !known[r.Refusal.Code] {
				t.Fatalf("refused with %v", r.Refusal)
			}
			if r.Refusal.Line < 1 || r.Refusal.Line > lines {
				t.Fatalf("line %d of %d: %v", r.Refusal.Line, lines, r.Refusal)
			}
			for _, b := range []byte(r.Refusal.Error()) {
				if b < 0x20 || b > 0x7e {
					t.Fatalf("not ASCII: %q", r.Refusal.Error())
				}
			}
		case Accepted:
			if r.Value == nil {
				t.Fatal("accepted with no value")
			}
			out, err := schema.Encode(JSON(r.Value))
			if err != nil {
				t.Fatalf("the value does not encode: %v", err)
			}
			if again := read(raw); again.Outcome != Accepted || !schema.Equal(JSON(again.Value), JSON(r.Value)) {
				t.Fatalf("a second read differs: %s", out)
			}
		default:
			t.Fatalf("outcome %d", r.Outcome)
		}
	})
}

// Fuzzing the format-0 reader: whatever the bytes, it accepts or refuses, never panics or hangs; a refusal names a
// line of the file, with no code and a next step, in ASCII; an accepted value is a mapping or a list that the JSON
// writer takes and a second read gives again. And where format 1 accepts a file that holds nothing format 0 is known to
// read otherwise (a quote, a block scalar, a {, which format 1 reads in { } and format 0 only in {}, a comment right
// after a list's dash, which format 0 reads as a null item, a byte past ASCII, which a BOM and every whitespace
// JavaScript adds are),
// format 0 reads the same value, keys in the same order, numbers as the same float64. `go test` runs the seeds; `go
// test -fuzz FuzzFormat0 ./internal/reader/` explores further.
func FuzzFormat0(f *testing.F) {
	_ = filepath.WalkDir(setDir, func(p string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() && (strings.HasSuffix(p, ".yaml") || strings.HasSuffix(p, ".md")) {
			if raw, err := os.ReadFile(p); err == nil {
				f.Add(raw, strings.HasSuffix(p, ".md"))
			}
		}
		return nil
	})
	for _, c := range cases {
		f.Add([]byte(c.in), c.md)
	}
	for _, c := range cases0 {
		f.Add([]byte(c.in), c.md)
	}
	f.Fuzz(func(t *testing.T, raw []byte, md bool) {
		read0, read1 := ReadFormat0YAML, ReadYAML
		if md {
			read0, read1 = ReadFormat0Markdown, ReadMarkdown
		}
		r := read0(raw)
		lines := bytes.Count(raw, []byte("\n")) + 1
		switch r.Outcome {
		case Refused:
			if r.Refusal == nil || r.Refusal.Code != "" || r.Refusal.Next == "" || r.Value != nil {
				t.Fatalf("refused with %#v", r.Refusal)
			}
			if r.Refusal.Line < 1 || r.Refusal.Line > lines {
				t.Fatalf("line %d of %d: %v", r.Refusal.Line, lines, r.Refusal)
			}
			for _, b := range []byte(r.Refusal.Error()) {
				if b < 0x20 || b > 0x7e {
					t.Fatalf("not ASCII: %q", r.Refusal.Error())
				}
			}
		case Accepted:
			switch r.Value.(type) {
			case *Map, []any:
			default:
				t.Fatalf("accepted a %T", r.Value)
			}
			if _, err := schema.Encode(JSON(r.Value)); err != nil {
				t.Fatalf("the value does not encode: %v", err)
			}
			if again := read0(raw); again.Outcome != Accepted || stringify0(again.Value) != stringify0(r.Value) {
				t.Fatalf("a second read differs: %s", stringify0(r.Value))
			}
		default:
			t.Fatalf("outcome %d", r.Outcome)
		}
		if bytes.ContainsAny(raw, `'"|>{`) || bytes.IndexFunc(raw, func(c rune) bool { return c > 0x7e }) >= 0 ||
			dashComment.Match(raw) {
			return
		}
		if r1 := read1(raw); r1.Outcome == Accepted {
			if r.Outcome != Accepted || !sameAs1(r1.Value, r.Value) {
				t.Fatalf("format 1 reads %s, format 0 %s (%v)", schema.Show(JSON(r1.Value)), stringify0(r.Value), r.Err())
			}
		}
	})
}

var dashComment = regexp.MustCompile(`-[ \t]+#`)

// sameAs1 compares a format-1 value with a format-0 one: the same keys in the same order, numbers as float64.
func sameAs1(v1, v0 any) bool {
	switch x := v1.(type) {
	case *Map:
		y, ok := v0.(*Map)
		if !ok || x.Len() != y.Len() {
			return false
		}
		for i, e := range x.Entries() {
			if f := y.Entries()[i]; f.Key != e.Key || !sameAs1(e.Value, f.Value) {
				return false
			}
		}
		return true
	case []any:
		y, ok := v0.([]any)
		if !ok || len(x) != len(y) {
			return false
		}
		for i := range x {
			if !sameAs1(x[i], y[i]) {
				return false
			}
		}
		return true
	case int64:
		return v0 == float64(x)
	case Decimal:
		return v0 == x.Float64()
	}
	return v1 == v0
}
