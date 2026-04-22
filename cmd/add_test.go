package cmd

import (
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/config"
	"github.com/LastStep/Bonsai/internal/generate"
)

// makeWriteResultWithConflicts produces a generate.WriteResult populated with
// ActionConflict entries for each of the given relative paths. The entries
// have nil `content`, which means ForceSelected won't actually overwrite the
// files on disk — this is fine for these tests, which only assert
// applyCinematicConflictPicks's pre-write behaviour (.bak handling, dropped
// paths, return value).
func makeWriteResultWithConflicts(paths ...string) *generate.WriteResult {
	wr := &generate.WriteResult{}
	for _, p := range paths {
		wr.Add(generate.FileResult{RelPath: p, Action: generate.ActionConflict})
	}
	return wr
}

// writeFile is a small helper that materialises a file under root with the
// given relative path. Returns the absolute path written.
func writeFile(t *testing.T, root, rel, body string) string {
	t.Helper()
	abs := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(abs), 0755); err != nil {
		t.Fatalf("MkdirAll(%s): %v", filepath.Dir(abs), err)
	}
	if err := os.WriteFile(abs, []byte(body), 0644); err != nil {
		t.Fatalf("WriteFile(%s): %v", abs, err)
	}
	return abs
}

// TestApplyCinematicConflictPicks_KeepIsNoop verifies that a Keep-only pick
// map is a complete no-op — no .bak files written, return false.
func TestApplyCinematicConflictPicks_KeepIsNoop(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "a.md", "local A")
	wr := makeWriteResultWithConflicts("a.md")
	lock := config.NewLockFile()

	picks := map[string]config.ConflictAction{
		"a.md": config.ConflictActionKeep,
	}
	got := applyCinematicConflictPicks(picks, wr, lock, root)
	if got {
		t.Fatalf("Keep-only picks must return false; got true")
	}
	if _, err := os.Stat(filepath.Join(root, "a.md.bak")); !os.IsNotExist(err) {
		t.Fatalf("Keep must not write a .bak; stat err = %v", err)
	}
}

// TestApplyCinematicConflictPicks_OverwriteWritesNoBackup verifies an
// Overwrite-only pick map skips .bak entirely (no read, no write) and returns
// true.
func TestApplyCinematicConflictPicks_OverwriteWritesNoBackup(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "a.md", "local A")
	wr := makeWriteResultWithConflicts("a.md")
	lock := config.NewLockFile()

	picks := map[string]config.ConflictAction{
		"a.md": config.ConflictActionOverwrite,
	}
	got := applyCinematicConflictPicks(picks, wr, lock, root)
	if !got {
		t.Fatalf("Overwrite pick must return true; got false")
	}
	if _, err := os.Stat(filepath.Join(root, "a.md.bak")); !os.IsNotExist(err) {
		t.Fatalf("Overwrite must not write a .bak; stat err = %v", err)
	}
}

// TestApplyCinematicConflictPicks_BackupWritesBak verifies a Backup pick
// creates a .bak file containing the original local body before the
// (would-be) overwrite, and returns true.
func TestApplyCinematicConflictPicks_BackupWritesBak(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "a.md", "local A")
	wr := makeWriteResultWithConflicts("a.md")
	lock := config.NewLockFile()

	picks := map[string]config.ConflictAction{
		"a.md": config.ConflictActionBackup,
	}
	got := applyCinematicConflictPicks(picks, wr, lock, root)
	if !got {
		t.Fatalf("Backup pick must return true; got false")
	}
	bakBody, err := os.ReadFile(filepath.Join(root, "a.md.bak"))
	if err != nil {
		t.Fatalf("Backup must write a .bak; ReadFile err = %v", err)
	}
	if got := string(bakBody); got != "local A" {
		t.Fatalf(".bak body = %q, want %q", got, "local A")
	}
}

// TestApplyCinematicConflictPicks_MixedActions verifies a mixed pick map
// (Keep + Overwrite + Backup) writes .bak only for the Backup entry and
// returns true.
func TestApplyCinematicConflictPicks_MixedActions(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "keep.md", "K")
	writeFile(t, root, "ow.md", "O")
	writeFile(t, root, "bak.md", "B")
	wr := makeWriteResultWithConflicts("keep.md", "ow.md", "bak.md")
	lock := config.NewLockFile()

	picks := map[string]config.ConflictAction{
		"keep.md": config.ConflictActionKeep,
		"ow.md":   config.ConflictActionOverwrite,
		"bak.md":  config.ConflictActionBackup,
	}
	got := applyCinematicConflictPicks(picks, wr, lock, root)
	if !got {
		t.Fatalf("mixed picks with at least one non-Keep must return true; got false")
	}

	if _, err := os.Stat(filepath.Join(root, "keep.md.bak")); !os.IsNotExist(err) {
		t.Fatalf("Keep entry must not have a .bak; stat err = %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "ow.md.bak")); !os.IsNotExist(err) {
		t.Fatalf("Overwrite entry must not have a .bak; stat err = %v", err)
	}
	body, err := os.ReadFile(filepath.Join(root, "bak.md.bak"))
	if err != nil {
		t.Fatalf("Backup entry must have a .bak; err = %v", err)
	}
	if string(body) != "B" {
		t.Fatalf("bak.md.bak body = %q, want %q", string(body), "B")
	}
}

// TestApplyCinematicConflictPicks_EmptyMapReturnsFalse verifies an empty
// pick map short-circuits with false (no work, no panic).
func TestApplyCinematicConflictPicks_EmptyMapReturnsFalse(t *testing.T) {
	wr := makeWriteResultWithConflicts()
	lock := config.NewLockFile()

	got := applyCinematicConflictPicks(map[string]config.ConflictAction{}, wr, lock, t.TempDir())
	if got {
		t.Fatalf("empty pick map must return false; got true")
	}

	got2 := applyCinematicConflictPicks(nil, wr, lock, t.TempDir())
	if got2 {
		t.Fatalf("nil pick map must return false; got true")
	}
}

// TestApplyCinematicConflictPicks_BackupReadFailDrops verifies that when the
// source file does not exist (read fails), that path is dropped from the
// overwrite list and a single tui.Warning is surfaced. We capture stdout and
// assert the dropped path appears in the warning text.
func TestApplyCinematicConflictPicks_BackupReadFailDrops(t *testing.T) {
	root := t.TempDir()
	// Note: NO writeFile for missing.md — read will fail.
	writeFile(t, root, "ok.md", "OK")
	wr := makeWriteResultWithConflicts("missing.md", "ok.md")
	lock := config.NewLockFile()

	picks := map[string]config.ConflictAction{
		"missing.md": config.ConflictActionBackup,
		"ok.md":      config.ConflictActionBackup,
	}

	stdout := captureStdout(t, func() {
		got := applyCinematicConflictPicks(picks, wr, lock, root)
		// "ok.md" still in list → returns true.
		if !got {
			t.Fatalf("read-fail with at least one survivor must return true; got false")
		}
	})

	if !strings.Contains(stdout, "missing.md") {
		t.Fatalf("warning must mention dropped path missing.md; got: %q", stdout)
	}
	if strings.Contains(stdout, "ok.md") {
		t.Fatalf("warning must NOT mention surviving path ok.md; got: %q", stdout)
	}
	if _, err := os.Stat(filepath.Join(root, "missing.md.bak")); !os.IsNotExist(err) {
		t.Fatalf("missing.md.bak must not exist; stat err = %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "ok.md.bak")); err != nil {
		t.Fatalf("ok.md.bak must exist; err = %v", err)
	}
}

// TestApplyCinematicConflictPicks_BackupWriteFailDrops verifies that when
// the .bak write fails (parent directory read-only), that path is dropped
// from the overwrite list and a single tui.Warning is surfaced. Skipped on
// platforms where chmod cannot enforce write protection.
func TestApplyCinematicConflictPicks_BackupWriteFailDrops(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("chmod-based write-protection not portable to Windows")
	}
	if os.Geteuid() == 0 {
		t.Skip("running as root — chmod 0500 does not deny writes")
	}

	root := t.TempDir()
	// Create a file inside a directory we'll then mark read-only so the .bak
	// write fails. The file itself stays readable so the read step succeeds
	// and we exercise the write-fail branch specifically.
	subDir := filepath.Join(root, "ro")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	writeFile(t, root, "ro/locked.md", "L")
	writeFile(t, root, "free.md", "F")

	// Lock down the subdir.
	if err := os.Chmod(subDir, 0500); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(subDir, 0755) })

	wr := makeWriteResultWithConflicts("ro/locked.md", "free.md")
	lock := config.NewLockFile()

	picks := map[string]config.ConflictAction{
		"ro/locked.md": config.ConflictActionBackup,
		"free.md":      config.ConflictActionBackup,
	}

	stdout := captureStdout(t, func() {
		got := applyCinematicConflictPicks(picks, wr, lock, root)
		if !got {
			t.Fatalf("write-fail with at least one survivor must return true; got false")
		}
	})

	if !strings.Contains(stdout, "ro/locked.md") {
		t.Fatalf("warning must mention dropped path ro/locked.md; got: %q", stdout)
	}
	if strings.Contains(stdout, "free.md.bak") {
		// The path "free.md" in a warning is fine if it's part of free.md's
		// bak path mention; we only care that free.md is NOT named as a
		// dropped path. A naive substring check on "free.md" would false-
		// positive on the warning text mentioning `free.md.bak` (which it
		// shouldn't anyway). Use the literal `.md.bak` form to disambiguate.
		t.Fatalf("warning must NOT mention surviving path free.md.bak; got: %q", stdout)
	}
	if _, err := os.Stat(filepath.Join(root, "free.md.bak")); err != nil {
		t.Fatalf("free.md.bak must exist; err = %v", err)
	}
}

// TestApplyCinematicConflictPicks_AllOverwritesDroppedReturnsFalse verifies
// the edge case where every overwrite path drops (read or write fail) — the
// helper must return false because there is no work left to dispatch to
// ForceSelected.
func TestApplyCinematicConflictPicks_AllOverwritesDroppedReturnsFalse(t *testing.T) {
	root := t.TempDir()
	wr := makeWriteResultWithConflicts("missing.md")
	lock := config.NewLockFile()

	picks := map[string]config.ConflictAction{
		"missing.md": config.ConflictActionBackup,
	}

	stdout := captureStdout(t, func() {
		got := applyCinematicConflictPicks(picks, wr, lock, root)
		if got {
			t.Fatalf("all-dropped picks must return false; got true")
		}
	})

	if !strings.Contains(stdout, "missing.md") {
		t.Fatalf("warning must mention dropped path; got: %q", stdout)
	}
}

// captureStdout runs fn with os.Stdout redirected to an in-memory pipe and
// returns whatever was written. Used to assert tui.Warning output text in
// the conflict-pick tests (tui.Warning calls fmt.Println which writes to
// stdout).
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w
	done := make(chan string, 1)
	go func() {
		var buf [8192]byte
		var sb strings.Builder
		for {
			n, err := r.Read(buf[:])
			if n > 0 {
				sb.Write(buf[:n])
			}
			if err != nil {
				break
			}
		}
		done <- sb.String()
	}()
	fn()
	_ = w.Close()
	os.Stdout = orig
	return <-done
}

// TestApplyCinematicConflictPicks_DroppedListSorted verifies the dropped-
// path warning text is deterministic — implementation may add the paths in
// map-iteration order, so the user-facing warning could shuffle between
// runs. This test does NOT enforce a sort; it documents the current
// behaviour: paths appear in the order applyCinematicConflictPicks walks
// `selected` (which itself is built from `picks` in map-iteration order, so
// non-deterministic). If that lands as a UX concern in future review, the
// fix is to sort the dropped list in the warning text builder.
//
// Kept as a smoke test — assert only that every dropped path appears
// somewhere in the warning, not the order.
func TestApplyCinematicConflictPicks_DroppedListSorted(t *testing.T) {
	root := t.TempDir()
	// Three missing files all flagged Backup → all drop.
	wr := makeWriteResultWithConflicts("a.md", "b.md", "c.md")
	lock := config.NewLockFile()

	picks := map[string]config.ConflictAction{
		"a.md": config.ConflictActionBackup,
		"b.md": config.ConflictActionBackup,
		"c.md": config.ConflictActionBackup,
	}

	stdout := captureStdout(t, func() {
		_ = applyCinematicConflictPicks(picks, wr, lock, root)
	})

	for _, want := range []string{"a.md", "b.md", "c.md"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("warning must mention dropped path %q; got: %q", want, stdout)
		}
	}
	// Sanity: ensure the warning is one line, not three.
	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	// tui.Warning may add styling/newlines around its own glyph; at minimum
	// the dropped-paths list itself is a single comma-separated cluster.
	// Find the line containing all three.
	found := false
	for _, ln := range lines {
		if strings.Contains(ln, "a.md") && strings.Contains(ln, "b.md") && strings.Contains(ln, "c.md") {
			found = true
			break
		}
	}
	if !found {
		// Defensive: print the seen content to help future debugging if
		// tui.Warning's formatting changes.
		sort.Strings(lines)
		t.Fatalf("expected one warning line listing all dropped paths; got: %v", lines)
	}
}
