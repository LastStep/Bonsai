package engine

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// fakeClone makes a clone folder in the home's cache for source, last used at used (its mark's time).
func fakeClone(t *testing.T, home, source string, used time.Time) string {
	t.Helper()
	dir := cache{home: home}.dir(source)
	if err := os.MkdirAll(filepath.Join(dir, "objects", "pack"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{"HEAD", "objects/pack/pack-1.pack"} {
		if err := os.WriteFile(filepath.Join(dir, filepath.FromSlash(f)), []byte("x\n"), 0o444); err != nil {
			t.Fatal(err)
		}
	}
	defer func(c func() time.Time) { cacheClock = c }(cacheClock)
	cacheClock = func() time.Time { return used }
	markUsed(dir)
	return dir
}

func setKeepDays(t *testing.T, home, body string) {
	t.Helper()
	if err := os.MkdirAll(home, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, "settings.json"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// Two projects sharing the home's clones: a clone past its days and in neither project's lock is removed, one past
// its days that this project's lock names is kept, one used today is kept, and an --adopt copy years old is never
// touched; with cache_keep_days none (no settings file) nothing is removed.
func TestCleanCache(t *testing.T) {
	home := t.TempDir()
	now := time.Now()
	old := now.Add(-40 * 24 * time.Hour)
	stale := fakeClone(t, home, "https://example.invalid/stale.git", old)
	locked := fakeClone(t, home, "https://example.invalid/locked.git", old)
	fresh := fakeClone(t, home, "https://example.invalid/fresh.git", now.Add(-2*time.Hour))
	adopted := filepath.Join(AdoptedDir(home, "ws-x", strings.Repeat("a", 64)), "CLAUDE.md")
	if err := os.MkdirAll(filepath.Dir(adopted), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(adopted, []byte("a person's edit\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	years := now.Add(-5 * 365 * 24 * time.Hour)
	_ = os.Chtimes(adopted, years, years)
	_ = os.Chtimes(filepath.Dir(adopted), years, years)
	other := filepath.Join(home, "cache", "git", "not-a-clone")
	if err := os.MkdirAll(other, 0o755); err != nil {
		t.Fatal(err)
	}
	_ = os.Chtimes(other, years, years)

	if cc := CleanCache(home, nil); cc != nil {
		t.Fatalf("no settings: %+v", cc)
	}
	setKeepDays(t, home, `{"cache_keep_days": 30}`)
	// The other project's update: its lock names fresh only.
	cc := CleanCache(home, []string{"https://example.invalid/locked.git"})
	if cc == nil || len(cc.Removed) != 1 || cc.Removed[0] != filepath.Base(stale) || len(cc.Failed) != 0 {
		t.Fatalf("cleaned: %+v", cc)
	}
	if !strings.Contains(cc.Line(), "removed 1 pack clone unused for over 30 days") {
		t.Errorf("its line: %q", cc.Line())
	}
	for _, kept := range []string{locked, fresh, adopted, other} {
		if _, err := os.Stat(kept); err != nil {
			t.Errorf("%s was removed", kept)
		}
	}
	if _, err := os.Stat(stale); err == nil {
		t.Errorf("%s was kept", stale)
	}
	// Nothing past its days: no line.
	if cc := CleanCache(home, []string{"https://example.invalid/locked.git"}); cc == nil || cc.Line() != "" {
		t.Errorf("again: %+v %q", cc, cc.Line())
	}
	// A settings file outside its rule: nothing cleaned, the line says why.
	setKeepDays(t, home, `{"cache_keep_days": 0}`)
	if cc := CleanCache(home, nil); cc == nil || !strings.Contains(cc.Line(), "not cleaned") {
		t.Errorf("a bad file: %+v", cc)
	}
	if _, err := os.Stat(locked); err != nil {
		t.Errorf("a bad file removed %s", locked)
	}
}

// The fetch and the read mark a clone used: a clone with no mark counts its newest entry, and one marked now is
// never older than a day.
func TestMarkUsed(t *testing.T) {
	home := t.TempDir()
	dir := cache{home: home}.dir("https://example.invalid/p.git")
	markUsed(dir) // not there yet: nothing made
	if _, err := os.Stat(dir); err == nil {
		t.Fatal("markUsed made a clone folder")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "HEAD"), []byte("ref: refs/heads/main\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	past := time.Now().Add(-100 * 24 * time.Hour)
	_ = os.Chtimes(filepath.Join(dir, "HEAD"), past, past)
	_ = os.Chtimes(dir, past, past)
	if used, ok := lastUse(dir); !ok || used.After(past.Add(time.Minute)) {
		t.Errorf("no mark: %v %v", used, ok)
	}
	markUsed(dir)
	if used, ok := lastUse(dir); !ok || time.Since(used) > time.Hour {
		t.Errorf("marked: %v %v", used, ok)
	}
}
