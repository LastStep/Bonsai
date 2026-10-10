package record

// Tests of the reader: a file's records in order, a torn line skipped and counted (alone, and glued to the record
// after it), the last record read from the end only, and a log folder's files by their names. The records are written
// by New and Line, so no test here holds the log's list of fields.

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/workspace"
)

// lines gives n log records' lines and their ids, each with its own target.
func lines(t *testing.T, n int) ([][]byte, []string) {
	t.Helper()
	var out [][]byte
	var ids []string
	for i := 0; i < n; i++ {
		l := New("event", Common{Workspace: raceWorkspace, Session: raceSession, Agent: "claude-code"})
		target := "file-" + strings.Repeat("x", i%40)
		l.Target = &target
		line, err := Line(l)
		if err != nil {
			t.Fatal(err)
		}
		out = append(out, line)
		ids = append(ids, l.ID)
	}
	return out, ids
}

func writeFile(t *testing.T, parts ...[]byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "s-"+raceSession+".ndjson")
	if err := os.WriteFile(path, bytes.Join(parts, nil), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func idsOf(f *LogFile) []string {
	out := []string{}
	for _, r := range f.Records {
		out = append(out, r.ID)
	}
	return out
}

func TestReadLog(t *testing.T) {
	ls, ids := lines(t, 4)
	torn := ls[1][:len(ls[1])/2]
	crlf := append(bytes.Clone(bytes.TrimSuffix(ls[3], []byte("\n"))), '\r', '\n')
	future := append(bytes.Clone(bytes.TrimSuffix(ls[0], []byte("}\n"))), []byte(`,"zz_future":"kept"}`+"\n")...)
	for _, c := range []struct {
		name    string
		parts   [][]byte
		ids     []string
		skipped int
	}{
		{"whole lines", [][]byte{ls[0], ls[1], ls[2]}, ids[:3], 0},
		{"a torn last line", [][]byte{ls[0], ls[2], torn}, []string{ids[0], ids[2]}, 1},
		{"a torn line, then an append glued to it", [][]byte{ls[0], torn, ls[2], ls[3]}, []string{ids[0], ids[2], ids[3]}, 1},
		{"a line written by hand, an empty line", [][]byte{[]byte("hello\n"), ls[0], []byte("\n"), ls[2]}, []string{ids[0], ids[2]}, 2},
		{"a torn line glued to a line that does not read", [][]byte{torn, []byte(`{"format":"bonsai.log/1","id":` + "\n"), ls[2]}, []string{ids[2]}, 1},
		{"a CRLF line", [][]byte{ls[0], crlf}, []string{ids[0], ids[3]}, 0},
		{"a last record with no line feed", [][]byte{ls[0], bytes.TrimSuffix(ls[2], []byte("\n"))}, []string{ids[0], ids[2]}, 0},
		{"an empty file", nil, []string{}, 0},
	} {
		f, err := ReadLog(writeFile(t, c.parts...))
		if err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		if got := idsOf(f); strings.Join(got, " ") != strings.Join(c.ids, " ") || f.Skipped != c.skipped {
			t.Errorf("%s: read %v, skipped %d; want %v, skipped %d", c.name, got, f.Skipped, c.ids, c.skipped)
		}
	}
	// A field this Bonsai does not know is kept (contract §2.2).
	f, err := ReadLog(writeFile(t, future))
	if err != nil || len(f.Records) != 1 || f.Skipped != 0 {
		t.Fatalf("a record with a field from a later minor: %+v, %v", f, err)
	}
	if v, ok := f.Records[0].Extra.Get("zz_future"); !ok || v != "kept" {
		t.Errorf("the unknown field was not kept: %v", f.Records[0].Extra)
	}
	if _, err := ReadLog(filepath.Join(t.TempDir(), "missing.ndjson")); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("a missing file: %v", err)
	}
}

// countingReader counts the bytes LastLog reads.
type countingReader struct {
	r    *bytes.Reader
	read int64
}

func (c *countingReader) ReadAt(p []byte, off int64) (int, error) {
	n, err := c.r.ReadAt(p, off)
	c.read += int64(n)
	return n, err
}

func TestLastLog(t *testing.T) {
	ls, ids := lines(t, 300)
	all := bytes.Join(ls, nil)
	torn := ls[5][:40]
	for _, chunk := range []int64{16 << 10, 100, 7} { // the default; smaller than a line; tiny
		lastChunk = chunk
		for _, c := range []struct {
			name  string
			parts [][]byte
			want  string
		}{
			{"many records", [][]byte{all}, ids[299]},
			{"a torn last line", [][]byte{all, torn}, ids[299]},
			{"a torn line glued to the last record", [][]byte{ls[0], torn, ls[1]}, ids[1]},
			{"junk after the records", [][]byte{ls[0], []byte("junk\n\n"), torn}, ids[0]},
			{"no line feed at the end", [][]byte{ls[0], bytes.TrimSuffix(ls[1], []byte("\n"))}, ids[1]},
			{"one record", [][]byte{ls[7]}, ids[7]},
			{"nothing but junk", [][]byte{[]byte("a\nb\n"), torn}, ""},
			{"an empty file", nil, ""},
		} {
			rec, err := LastLog(writeFile(t, c.parts...))
			got := ""
			if rec != nil {
				got = rec.ID
			}
			if err != nil || got != c.want {
				t.Errorf("chunk %d, %s: %q, %v; want %q", chunk, c.name, got, err, c.want)
			}
		}
	}
	lastChunk = 16 << 10
	// The file is read back from its end only as far as the last record: of a 64 KB file, one chunk.
	big := bytes.Repeat(all, 1+(64<<10)/len(all))
	cr := &countingReader{r: bytes.NewReader(big)}
	rec, err := lastRecord(cr, int64(len(big)))
	if err != nil || rec == nil || rec.ID != ids[299] {
		t.Fatalf("the big file: %v, %v", rec, err)
	}
	if cr.read > lastChunk || int64(len(big)) <= 2*lastChunk {
		t.Errorf("read %d bytes of %d to find the last record, want at most %d", cr.read, len(big), lastChunk)
	}
	if _, err := LastLog(filepath.Join(t.TempDir(), "missing.ndjson")); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("a missing file: %v", err)
	}
}

// ListLog lists session and day files by their names, regular files only; nothing else in the folder counts.
func TestListLog(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "log")
	if got, err := ListLog(dir); err != nil || len(got.Sessions) != 0 || len(got.Days) != 0 {
		t.Errorf("a missing folder: %+v, %v", got, err)
	}
	for _, name := range []string{"s-" + raceSession + ".ndjson", "s-b_2.ndjson", "w-2026-10-10.ndjson", "w-2026-01-31.ndjson",
		"w-2026-13-01.ndjson", "w-2026-02-30.ndjson", "notes.txt", "s-.ndjson", "s-a b.ndjson", "s-x.json", "S-y.ndjson",
		"w-2026-10-10.ndjson.tmp", ".s-z.ndjson"} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, name), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Mkdir(filepath.Join(dir, "s-folder.ndjson"), 0o755); err != nil {
		t.Fatal(err)
	}
	// A symbolic link named like a session file is not one of Bonsai's files. Made only off Windows, where a link
	// needs no privilege; the rest of the test runs everywhere.
	if runtime.GOOS != "windows" {
		if err := os.Symlink(filepath.Join(dir, "notes.txt"), filepath.Join(dir, "s-link.ndjson")); err != nil {
			t.Fatal(err)
		}
	}
	got, err := ListLog(dir)
	if err != nil {
		t.Fatal(err)
	}
	if s := strings.Join(got.Sessions, " "); s != "s-"+raceSession+".ndjson s-b_2.ndjson" {
		t.Errorf("sessions: %s", s)
	}
	if d := strings.Join(got.Days, " "); d != "w-2026-01-31.ndjson w-2026-10-10.ndjson" {
		t.Errorf("days: %s", d)
	}
	if SessionOf("s-"+raceSession+".ndjson") != raceSession || SessionOf("w-2026-10-10.ndjson") != "" {
		t.Errorf("SessionOf")
	}
	if DayOf("w-2026-10-10.ndjson") != "2026-10-10" || DayOf("w-2026-02-30.ndjson") != "" || DayOf("s-x.ndjson") != "" {
		t.Errorf("DayOf")
	}
}

// LogFileName names a session's file and a day's, and refuses a session id that could leave the folder.
func TestLogFileName(t *testing.T) {
	at := time.Date(2026, 10, 10, 23, 30, 0, 0, time.FixedZone("east", 5*3600))
	if n, err := LogFileName("", at); err != nil || n != "w-2026-10-10.ndjson" {
		t.Errorf("no session: %q, %v (the day is UTC's)", n, err)
	}
	if n, err := LogFileName(raceSession, at); err != nil || n != "s-"+raceSession+".ndjson" {
		t.Errorf("a session: %q, %v", n, err)
	}
	for _, s := range []string{"../x", "a/b", `a\b`, "a b", "a.b", "con:x", strings.Repeat("a", 129), "é"} {
		if n, err := LogFileName(s, at); err == nil {
			t.Errorf("%q named %q", s, n)
		}
	}
	// WriteLog refuses the same, and a record whose at is not Bonsai's form.
	main := t.TempDir()
	l := New("event", Common{Workspace: raceWorkspace, Session: "../escape"})
	if _, err := WriteLog(main, l); err == nil {
		t.Errorf("a session id holding ../ was written")
	}
	l = New("event", Common{Workspace: raceWorkspace})
	l.At = "yesterday"
	if _, err := WriteLog(main, l); err == nil {
		t.Errorf("an at of %q was written", l.At)
	}
	if _, err := os.Stat(filepath.Join(main, filepath.FromSlash(workspace.LocalDir))); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("a refused record made .bonsai/local/: %v", err)
	}
}
