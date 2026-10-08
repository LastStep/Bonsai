package reader

// Fuzzing the reader: whatever the bytes, it returns one of the three outcomes, never panics, refuses only with a
// code in Codes on a line of the file, and gives a value the JSON writer takes. `go test` runs the seeds (every file
// of the formats set and the cases above); `go test -fuzz FuzzRead ./internal/reader/` explores further.

import (
	"bytes"
	"os"
	"path/filepath"
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
