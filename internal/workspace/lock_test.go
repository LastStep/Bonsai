package workspace

// Tests of the lock: it reads the set's example, writes it back byte for byte, validates against the lock schema,
// keeps unknown fields, refuses what it should, and writes whole.

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/schema"
)

// schemaAt walks a schema by keys and shows the value there.
func schemaAt(t *testing.T, s schema.Object, keys ...string) string {
	t.Helper()
	var v any = s
	for _, k := range keys {
		o, ok := v.(schema.Object)
		if !ok {
			t.Fatalf("schema: no %v", keys)
		}
		v, _ = o.Get(k)
	}
	return schema.Show(v)
}

func exampleLock(t *testing.T) []byte {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "formats", "examples", "lock.json"))
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestLockReadsAndWritesTheSetsExample(t *testing.T) {
	raw := exampleLock(t)
	l, err := ReadLock(raw)
	if err != nil {
		t.Fatal(err)
	}
	if l.WrittenBy != "1.0.0" || len(l.Packs) != 2 || l.Packs[1].ID != "workflow" || len(l.Files) != 2 ||
		l.Files["CLAUDE.md"].Kind != "block" || len(l.Format0) != 1 || len(l.Packs[1].Declares) != 1 {
		t.Errorf("read %+v", l)
	}
	out, err := l.Encode()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(out, raw) {
		t.Errorf("Encode(ReadLock(example)) differs from formats/examples/lock.json:\n%s", out)
	}
}

// What the writer writes validates against part 0's lock schema and keeps its field order, at every depth.
func TestLockOutputFitsTheSchema(t *testing.T) {
	l := &Lock{
		WrittenBy: "dev",
		Packs: []LockedPack{{ID: "test-pack", Source: "https://example.com/test-pack.git", Version: "0.1.0",
			Commit: strings.Repeat("a", 40), SHA256: strings.Repeat("b", 64)}},
		Files: map[string]LockedFile{
			"z/last.md":             {Kind: "pack", Pack: "test-pack", SHA256: strings.Repeat("c", 64)},
			".claude/settings.json": {Kind: "keys", Pack: "test-pack", SHA256: strings.Repeat("d", 64)},
			"a/first.md":            {Kind: "once", Pack: "test-pack", SHA256: strings.Repeat("e", 64)},
		},
	}
	out, err := l.Encode()
	if err != nil {
		t.Fatal(err)
	}
	doc, err := schema.Decode(out)
	if err != nil {
		t.Fatal(err)
	}
	if msgs := schema.Validate(LockSchema(), doc); len(msgs) != 0 {
		t.Errorf("the written lock does not validate: %v", msgs)
	}
	if msgs := schema.CheckOrder(LockSchema(), doc); len(msgs) != 0 {
		t.Errorf("the written lock is out of the schema's order: %v", msgs)
	}
	files, _ := doc.(schema.Object).Get("files")
	if got := strings.Join(files.(schema.Object).Keys(), " "); got != ".claude/settings.json a/first.md z/last.md" {
		t.Errorf("files not sorted: %s", got)
	}
	if !strings.Contains(string(out), `"declares": {}`) || !strings.Contains(string(out), `"format0": {}`) {
		t.Errorf("empty declares and format0 are not written as {}:\n%s", out)
	}
	for i := 0; i < 10; i++ {
		again, _ := l.Encode()
		if !bytes.Equal(again, out) {
			t.Fatal("two encodings differ")
		}
	}
}

func TestLockKeepsUnknownFields(t *testing.T) {
	in := `{"format": "bonsai.lock/1", "written_by": "9.0.0", "later": {"x": 1},
	"packs": [{"id": "p", "source": "s", "version": "1", "commit": "` + strings.Repeat("0", 40) + `",
	  "sha256": "` + strings.Repeat("1", 64) + `", "new_pack_field": [1, 2], "declares": {"lanes": []}}],
	"files": {"a.md": {"kind": "pack", "new_file_field": true, "pack": "p", "sha256": "` + strings.Repeat("2", 64) + `"}},
	"format0": {}, "last": null}`
	l, err := ReadLock([]byte(in))
	if err != nil {
		t.Fatal(err)
	}
	out, err := l.Encode()
	if err != nil {
		t.Fatal(err)
	}
	doc, _ := schema.Decode(out)
	o := doc.(schema.Object)
	if got := strings.Join(o.Keys(), " "); got != "format written_by packs files format0 later last" {
		t.Errorf("top-level keys %s", got)
	}
	packs, _ := o.Get("packs")
	p := packs.([]any)[0].(schema.Object)
	if got := strings.Join(p.Keys(), " "); got != "id source version commit sha256 declares new_pack_field" {
		t.Errorf("pack keys %s", got)
	}
	files, _ := o.Get("files")
	f, _ := files.(schema.Object).Get("a.md")
	if got := strings.Join(f.(schema.Object).Keys(), " "); got != "kind pack sha256 new_file_field" {
		t.Errorf("file keys %s", got)
	}
	if v, _ := o.Get("later"); !schema.Equal(v, schema.Object{{Key: "x", Value: jsonNumber("1")}}) {
		t.Errorf("later = %s", schema.Show(v))
	}
}

func TestLockRefuses(t *testing.T) {
	good := string(exampleLock(t))
	cases := []struct {
		name, in, want string
	}{
		{"a duplicate key", strings.Replace(good, `"written_by": "1.0.0",`, `"written_by": "1.0.0", "written_by": "1.0.0",`, 1), "duplicate key"},
		{"a BOM", "\xEF\xBB\xBF" + good, "byte order mark"},
		{"trailing data", good + "{}", "something follows"},
		{"not an object", "[]", "is not a JSON object"},
		{"a newer major", strings.Replace(good, "bonsai.lock/1", "bonsai.lock/2", 1), "format too new"},
		{"another format", strings.Replace(good, "bonsai.lock/1", "bonsai.task/1", 1), `its format is "bonsai.task/1"`},
		{"a kind not in the schema", strings.Replace(good, `"kind": "block"`, `"kind": "copy"`, 1), `"copy" is not one of`},
		{"a short commit", strings.Replace(good, `"0a1b2c3d4e5f60718293a4b5c6d7e8f901a2b3c4"`, `"0a1b2c3"`, 1), "does not match"},
		{"a missing field", strings.Replace(good, `"written_by": "1.0.0",`, ``, 1), `required field "written_by" is missing`},
		{"an absolute path", strings.Replace(good, `"CLAUDE.md": {`, `"/etc/CLAUDE.md": {`, 1), "does not match"},
		{"a path out", strings.Replace(good, `"CLAUDE.md": {`, `"a/../../CLAUDE.md": {`, 1), "files: the path"},
		{"a format0 path out", strings.Replace(good, `"work/runs/`, `"../runs/`, 1), "format0: the path"},
		{"a pack twice", strings.Replace(good, `"id": "workflow"`, `"id": "base"`, 1), `the pack "base" is locked twice`},
		{"declares not an object", strings.Replace(good, `"declares": {}`, `"declares": []`, 1), "is array, want object"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := ReadLock([]byte(c.in))
			if err == nil {
				t.Fatal("read with no error")
			}
			if msg := err.Error(); !strings.HasPrefix(msg, ".bonsai/lock.json: ") || !strings.Contains(msg, c.want) || !strings.Contains(msg, "; next: ") {
				t.Errorf("error %q, want it to hold %q and name a next step", msg, c.want)
			}
			if msg := err.Error(); strings.Contains(msg, `\u`) && c.name != "a BOM" {
				t.Errorf("an ASCII file gave a message with an escape (our own text is not ASCII): %q", msg)
			}
		})
	}
	// The writer refuses what the reader would.
	bad := []*Lock{
		{WrittenBy: "dev", Packs: []LockedPack{{ID: "p", Commit: "short", SHA256: strings.Repeat("0", 64)}}},
		{WrittenBy: "dev", Files: map[string]LockedFile{"a.md": {Kind: "copy", Pack: "p", SHA256: strings.Repeat("0", 64)}}},
		{WrittenBy: "dev", Files: map[string]LockedFile{"a\\b.md": {Kind: "pack", Pack: "p", SHA256: strings.Repeat("0", 64)}}},
		{WrittenBy: "dev", Format0: map[string]string{"nul.md": strings.Repeat("0", 64)}},
	}
	for i, l := range bad {
		if _, err := l.Encode(); err == nil {
			t.Errorf("bad lock %d encoded", i)
		}
		if err := WriteLock(t.TempDir(), l); err == nil {
			t.Errorf("bad lock %d written", i)
		}
	}
}

func TestWriteLockAndLoadLock(t *testing.T) {
	root := t.TempDir()
	if _, err := LoadLock(root); !errors.Is(err, fs.ErrNotExist) || !strings.Contains(err.Error(), "next: run bonsai update") {
		t.Fatalf("a missing lock: %v", err)
	}
	l, err := ReadLock(exampleLock(t))
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ { // the second write replaces the first
		if err := WriteLock(root, l); err != nil {
			t.Fatal(err)
		}
	}
	raw, err := os.ReadFile(filepath.Join(root, ".bonsai", "lock.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(raw, exampleLock(t)) {
		t.Errorf("the written lock differs from what was read")
	}
	entries, _ := os.ReadDir(filepath.Join(root, ".bonsai"))
	if len(entries) != 1 {
		t.Errorf(".bonsai holds %d files, want the lock alone (no temporary file left)", len(entries))
	}
	back, err := LoadLock(root)
	if err != nil {
		t.Fatal(err)
	}
	again, _ := back.Encode()
	if !bytes.Equal(again, raw) {
		t.Errorf("LoadLock does not read back what WriteLock wrote")
	}
}
