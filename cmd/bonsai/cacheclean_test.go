package main

// update's clean of the pack cache (plan-5 5.6.1, 6; engine.CleanCache): with the home's cache_keep_days set, update
// --yes (written, or with nothing to change) removes each pack clone unused for longer than that, never one the lock
// or bonsai.yaml names, never an --adopt copy, and says so in one line; a preview removes nothing.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/engine"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// agedClone makes a folder in the home's cache shaped as a pack clone, last used at used.
func agedClone(t *testing.T, home, name string, used time.Time) string {
	t.Helper()
	dir := filepath.Join(home, "cache", "git", name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{"HEAD", "bonsai-used"} {
		p := filepath.Join(dir, f)
		if err := os.WriteFile(p, []byte("x\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.Chtimes(p, used, used); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestUpdateCleansTheCache(t *testing.T) {
	c := newCLI(t)
	c.link()
	home, err := workspace.Home()
	if err != nil {
		t.Fatal(err)
	}
	if code, out, errOut := c.run("", "settings", "set", "cache_keep_days=30", "--machine"); code != 0 {
		t.Fatalf("settings set: %d\n%s%s", code, out, errOut)
	}
	old := time.Now().Add(-400 * 24 * time.Hour)
	// Every clone the link fetched is the lock's: aged too, so only the lock keeps it.
	entries, err := os.ReadDir(filepath.Join(home, "cache", "git"))
	if err != nil || len(entries) == 0 {
		t.Fatalf("the link fetched no clone: %v", err)
	}
	var locked []string
	for _, e := range entries {
		dir := filepath.Join(home, "cache", "git", e.Name())
		locked = append(locked, dir)
		_ = os.Chtimes(filepath.Join(dir, "bonsai-used"), old, old)
	}
	stranger := agedClone(t, home, "0123456789abcdef.git", old)
	adopted := filepath.Join(engine.AdoptedDir(home, "ws-x", strings.Repeat("a", 64)), "guide.md")
	if err := os.MkdirAll(filepath.Dir(adopted), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(adopted, []byte("a person's edit\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_ = os.Chtimes(adopted, old, old)

	// A pack added, so the update has something to write.
	src, shas := labelsPack(t, c.tmp, "extra")
	c.write("bonsai.yaml", strings.Replace(c.read("bonsai.yaml"), "\ndocuments:", "\n  - id: extra\n    source: \""+
		filepath.ToSlash(src)+"\"\n    path: null\n    ref: \""+shas[0]+"\"\ndocuments:", 1))

	// A preview (no --yes, no terminal) removes nothing.
	if code, out, errOut := c.run("", "update", "--allow-exec"); code != engine.ExitState || strings.Contains(out, "pack clone") {
		t.Fatalf("preview: %d\n%s%s", code, out, errOut)
	}
	if _, err := os.Stat(stranger); err != nil {
		t.Fatalf("a preview removed a clone: %v", err)
	}

	code, out, errOut := c.run("", "update", "--yes", "--allow-exec")
	if code != 0 {
		t.Fatalf("update: %d\n%s%s", code, out, errOut)
	}
	if !strings.Contains(flat(out), "removed 1 pack clone unused for over 30 days from the cache (cache_keep_days)") {
		t.Errorf("update's text names no clean:\n%s", out)
	}
	if _, err := os.Stat(stranger); err == nil {
		t.Errorf("the unused clone is still there: %s", stranger)
	}
	for _, kept := range append(locked, adopted) {
		if _, err := os.Stat(kept); err != nil {
			t.Errorf("removed, but the lock names it or --adopt saved it: %s", kept)
		}
	}

	// --json cleans too; the changes output has no field for it, and stays as it was (its result applied).
	second := agedClone(t, home, "fedcba9876543210.git", old)
	c.write("bonsai.yaml", strings.Replace(c.read("bonsai.yaml"), "  - id: extra\n    source: \""+filepath.ToSlash(src)+
		"\"\n    path: null\n    ref: \""+shas[0]+"\"\n", "", 1))
	code, out, errOut = c.run("", "update", "--yes", "--allow-exec", "--json")
	if code != 0 || strings.Contains(out, "pack clone") || !strings.Contains(out, `"result": "applied"`) {
		t.Errorf("update --json: %d\n%s%s", code, out, errOut)
	}
	if _, err := os.Stat(second); err == nil {
		t.Errorf("update --json cleaned nothing: %s", second)
	}

	// Nothing to change: update without --yes cleans nothing; update --yes cleans still.
	third := agedClone(t, home, "00112233aabbccdd.git", old)
	if code, out, errOut := c.run("", "update"); code != 0 || strings.Contains(out, "pack clone") {
		t.Fatalf("update, nothing to change: %d\n%s%s", code, out, errOut)
	}
	if _, err := os.Stat(third); err != nil {
		t.Fatalf("update without --yes removed a clone: %v", err)
	}
	code, out, errOut = c.run("", "update", "--yes")
	if code != 0 || !strings.Contains(out, "nothing to change") ||
		!strings.Contains(flat(out), "removed 1 pack clone unused for over 30 days from the cache (cache_keep_days)") {
		t.Errorf("update --yes, nothing to change:\n%d\n%s%s", code, out, errOut)
	}
	if _, err := os.Stat(third); err == nil {
		t.Errorf("update --yes with nothing to change cleaned nothing: %s", third)
	}
}
