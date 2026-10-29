package engine

// cache/packs.json (packcache.go): written sorted by source, a failed read keeping the last good one, a broken file
// not read, and the verdict worked out against the locked version at each read, never stored.

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestPackCacheRecordAndVerdict(t *testing.T) {
	home := t.TempDir()
	at := time.Date(2026, 10, 29, 8, 0, 0, 0, time.UTC)
	fakeClock(t, &at)
	c, err := RecordPackTags(home, []PackTagRead{
		{Source: "https://example.org/z.git", Tags: []string{"v1.0.1", "v1.0.0"}},
		{Source: "https://example.org/a.git", Tags: []string{"base-v1.0.0", "base-v1.0.2", "workflow-v3.0.0"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Sources) != 2 || c.Sources[0].Source != "https://example.org/a.git" || strings.Join(c.Sources[1].Tags, ",") != "v1.0.0,v1.0.1" {
		t.Fatalf("not sorted: %+v", c)
	}
	raw, _ := os.ReadFile(filepath.Join(home, filepath.FromSlash(PackCacheFile)))
	if !strings.Contains(string(raw), `"read": "2026-10-29T08:00:00Z"`) {
		t.Errorf("the file:\n%s", raw)
	}
	// The verdict follows the lock: base at 1.0.0 is behind, at 1.0.2 current; nothing is stored about either.
	if n := PackNewerOf(c, "https://example.org/a.git", "base", "base-v1.0.0", "packs/base", "1.0.0"); n == nil || n.State != PackNewer || n.Tag != "base-v1.0.2" {
		t.Errorf("behind: %+v", n)
	}
	if n := PackNewerOf(c, "https://example.org/a.git", "base", "base-v1.0.2", "packs/base", "1.0.2"); n == nil || n.State != PackCurrent || n.Tag != "base-v1.0.2" {
		t.Errorf("current: %+v", n)
	}
	// A commit ref: the pack's id prefix counts, the tag is the ref itself when none is newer.
	commit := "0123456789abcdef0123456789abcdef01234567"
	if n := PackNewerOf(c, "https://example.org/a.git", "workflow", commit, "packs/workflow", "3.0.0"); n == nil || n.State != PackCurrent || n.Tag != commit {
		t.Errorf("commit ref: %+v", n)
	}
	// A failed read keeps the good one.
	later := at.Add(time.Hour)
	at = later
	c, err = RecordPackTags(home, []PackTagRead{{Source: "https://example.org/z.git", Err: errors.New("git ls-remote failed: no network")}})
	if err != nil {
		t.Fatal(err)
	}
	r := c.Record("https://example.org/z.git")
	if r == nil || r.Failed == nil || r.Failed.At != "2026-10-29T09:00:00Z" || *r.Read != "2026-10-29T08:00:00Z" || len(r.Tags) != 2 {
		t.Errorf("after a failure: %+v", r)
	}
	// Only failures: no verdict.
	c, _ = RecordPackTags(home, []PackTagRead{{Source: "https://example.org/new.git", Err: errors.New("no")}})
	if n := PackNewerOf(c, "https://example.org/new.git", "x", "v1.0.0", "", "1.0.0"); n != nil {
		t.Errorf("only a failed read: %+v, want nil", n)
	}
	// A broken file is not read, and the next write replaces it.
	if err := os.WriteFile(filepath.Join(home, filepath.FromSlash(PackCacheFile)), []byte(`{"sources": [{"source": "x", "tags": [], "read": "yesterday", "failed": null}]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadPackCache(home); err == nil {
		t.Error("a broken record was read")
	}
	if c, err := RecordPackTags(home, []PackTagRead{{Source: "y", Tags: []string{}}}); err != nil || len(c.Sources) != 1 {
		t.Errorf("rewrite of a broken record: %+v, %v", c, err)
	}
}
