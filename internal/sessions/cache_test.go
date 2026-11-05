package sessions

// The sessions cache (BUG-137): FoundCached gives Found's rows whatever the cache holds; a file the cache holds at
// its size and modification time is not read again, and any other file is; a cache that does not read, is of another
// version, or is stale is never trusted; the 24-hour rule is applied on every run, cache or not; the cache's bytes are
// stable.

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

// foundCached runs FoundCached and checks its rows are Found's, the same folder at the same time.
func foundCached(t *testing.T, dir string, cache []byte, at time.Time) (string, []byte) {
	t.Helper()
	rows, next, err := FoundCached(dir, cache, at)
	if err != nil {
		t.Fatal(err)
	}
	plain, err := Found(dir, at)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rows, plain) {
		t.Fatalf("FoundCached's rows differ from Found's:\n%+v\n%+v", rows, plain)
	}
	return rowsShown(rows), next
}

func rowsShown(rows interface{}) string {
	b, _ := json.Marshal(rows)
	return string(b)
}

// sameSizeGarbage overwrites the file at path with bytes of its own length that hold no record, and puts its
// modification time back: a file the cache holds as it is must not be read again, and this shows whether it was.
func sameSizeGarbage(t *testing.T, path string) {
	t.Helper()
	fi, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, bytes.Repeat([]byte("x"), int(fi.Size())), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(path, fi.ModTime(), fi.ModTime()); err != nil {
		t.Fatal(err)
	}
}

func TestTheCacheIsUsedForAnUnchangedFile(t *testing.T) {
	dir := t.TempDir()
	p := writeLog(t, dir, sid, start("09:00", "T-1", "builder", "opus"), end("09:10"))
	first, cache := foundCached(t, dir, nil, now)
	if cache == nil || !strings.Contains(first, "T-1") {
		t.Fatalf("first run: rows %s, cache %q", first, cache)
	}
	// Nothing changed: the same rows, and no cache to write.
	again, next := foundCached(t, dir, cache, now)
	if again != first || next != nil {
		t.Errorf("an unchanged run: rows %s, next %q", again, next)
	}
	// The file's bytes change but its size and time do not: the cache is trusted, so the file was not read. (Found,
	// which reads it, finds no row: foundCached's comparison is not made here.)
	sameSizeGarbage(t, p)
	rows, next, err := FoundCached(dir, cache, now)
	if err != nil || rowsShown(rows) != first || next != nil {
		t.Errorf("a file held at its size and time was read again: %v %s %q", err, rowsShown(rows), next)
	}
}

func TestAChangedOrAddedFileIsReadAgain(t *testing.T) {
	dir := t.TempDir()
	p := writeLog(t, dir, sid, start("09:00", "T-1", "builder", "opus"), end("09:10"))
	_, cache := foundCached(t, dir, nil, now)

	// Appended to (its size changes): the new span is found.
	f, err := os.OpenFile(p, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	more := writeLog(t, t.TempDir(), sid, start("10:00", "T-2", "builder", "opus"), end("10:30"))
	b, err := os.ReadFile(more)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Write(b); err != nil {
		t.Fatal(err)
	}
	_ = f.Close()
	rows, next := foundCached(t, dir, cache, now)
	if !strings.Contains(rows, "T-2") || next == nil {
		t.Fatalf("an appended file: rows %s, next %q", rows, next)
	}
	cache = next

	// A new file: found.
	writeLog(t, dir, sid2, start("11:00", "T-3", "builder", "opus"), end("11:05"))
	rows, next = foundCached(t, dir, cache, now)
	if !strings.Contains(rows, "T-3") || next == nil {
		t.Fatalf("an added file: rows %s, next %q", rows, next)
	}
	cache = next

	// Rewritten to the same size at another time: read again (its rows follow its bytes).
	q := filepath.Join(dir, "s-"+sid2+".ndjson")
	old, err := os.ReadFile(q)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(q, bytes.Replace(old, []byte("T-3"), []byte("T-4"), 1), 0o644); err != nil {
		t.Fatal(err)
	}
	later := time.Now().Add(time.Hour)
	if err := os.Chtimes(q, later, later); err != nil {
		t.Fatal(err)
	}
	rows, next = foundCached(t, dir, cache, now)
	if strings.Contains(rows, "T-3") || !strings.Contains(rows, "T-4") || next == nil {
		t.Fatalf("a file rewritten at another time: rows %s, next %q", rows, next)
	}
	cache = next

	// Gone: its rows go with it, and so does its entry.
	if err := os.Remove(q); err != nil {
		t.Fatal(err)
	}
	rows, next = foundCached(t, dir, cache, now)
	if strings.Contains(rows, "T-4") || next == nil || bytes.Contains(next, []byte(sid2)) {
		t.Fatalf("a removed file: rows %s, next %q", rows, next)
	}
}

func TestAStaleOrForeignCacheIsNeverTrusted(t *testing.T) {
	dir := t.TempDir()
	p := writeLog(t, dir, sid, start("09:00", "T-1", "builder", "opus"), end("09:10"))
	want, cache := foundCached(t, dir, nil, now)
	fi, err := os.Stat(p)
	if err != nil {
		t.Fatal(err)
	}
	var c cacheData
	if err := json.Unmarshal(cache, &c); err != nil || len(c.Files) != 1 {
		t.Fatalf("the cache: %v %q", err, cache)
	}
	// A cache whose spans lie, under each way it can be wrong.
	lie := func(edit func(*cacheData)) []byte {
		var d cacheData
		_ = json.Unmarshal(cache, &d)
		d.Files[0].Spans[0].Task = "T-LIE"
		edit(&d)
		b, _ := json.Marshal(d)
		return b
	}
	for name, bad := range map[string][]byte{
		"another version":    lie(func(d *cacheData) { d.Version++ }),
		"another size":       lie(func(d *cacheData) { d.Files[0].Size++ }),
		"another time":       lie(func(d *cacheData) { d.Files[0].MTime = fi.ModTime().Add(-time.Second).UnixNano() }),
		"another file's":     lie(func(d *cacheData) { d.Files[0].Name = "s-" + sid2 + ".ndjson" }),
		"not json":           []byte("{\"version\":1,\"files\":["),
		"not a cache":        []byte("[1,2,3]\n"),
		"empty":              {},
		"a torn copy of one": cache[:len(cache)/2],
	} {
		rows, next := foundCached(t, dir, bad, now)
		if rows != want || strings.Contains(rows, "T-LIE") || !bytes.Equal(next, cache) {
			t.Errorf("%s: rows %s, next %q", name, rows, next)
		}
	}
	// The same lie at the file's true size and time is trusted: what makes the cases above safe is the check, not
	// chance.
	if rows, _, _ := FoundCached(dir, lie(func(*cacheData) {}), now); !strings.Contains(rowsShown(rows), "T-LIE") {
		t.Errorf("a cache held at the file's size and time was not used: %s", rowsShown(rows))
	}
}

func TestTheTwentyFourHourRuleIsAppliedWithTheCache(t *testing.T) {
	dir := t.TempDir()
	writeLog(t, dir, sid, start("09:00", "T-1", "builder", "opus"), tool("09:30")) // no end: open while recent
	soon := time.Date(2026, 10, 8, 12, 0, 0, 0, time.UTC)
	rows, cache := foundCached(t, dir, nil, soon)
	if rows != "null" || cache == nil {
		t.Fatalf("an open span has no row: %s %q", rows, cache)
	}
	// A day later, the file unchanged: the span ended at its last record, from the cache.
	rows, next := foundCached(t, dir, cache, now)
	if !strings.Contains(rows, `"end":"2026-10-08 09:30"`) || next != nil {
		t.Errorf("the 24-hour rule over a cached file: rows %s, next %q", rows, next)
	}
}

func TestTheCacheIsByteStable(t *testing.T) {
	dir := t.TempDir()
	writeLog(t, dir, sid2, start("10:00", "T-2", "", ""), subStart("10:01", "agent-1", "builder", "T-2"), subStop("10:20", "agent-1"), end("10:30"))
	writeLog(t, dir, sid, start("09:00", "T-1", "builder", "opus"), end("09:10"))
	_, a := foundCached(t, dir, nil, now)
	_, b := foundCached(t, dir, []byte("not a cache"), now)
	if a == nil || !bytes.Equal(a, b) || !bytes.HasSuffix(a, []byte("\n")) || bytes.Contains(a, []byte("\r")) {
		t.Fatalf("two builds of the cache:\n%s\n%s", a, b)
	}
	if i, j := bytes.Index(a, []byte(sid)), bytes.Index(a, []byte(sid2)); i < 0 || j < 0 || i > j {
		t.Errorf("files not in name order:\n%s", a)
	}
}

func TestNoLogNoCache(t *testing.T) {
	for _, dir := range []string{filepath.Join(t.TempDir(), "nope"), t.TempDir()} {
		rows, next, err := FoundCached(dir, nil, now)
		if err != nil || len(rows) != 0 || next != nil {
			t.Errorf("%s: %v %v %q", dir, err, rows, next)
		}
	}
	// A cache with files the log no longer has is written again, empty.
	rows, next, err := FoundCached(t.TempDir(), []byte(`{"version":1,"files":[{"name":"s-x.ndjson"}]}`), now)
	if err != nil || len(rows) != 0 || string(next) != "{\"version\":1,\"files\":[]}\n" {
		t.Errorf("a cache over an emptied log: %v %v %q", err, rows, next)
	}
}
