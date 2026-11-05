package engine

// The sessions table check through the sessions cache (BUG-137): bonsai check writes the cache in the main
// checkout's .bonsai/local/ once there is a log, and still finds every row the table lacks after a log file changes
// or is added since its last run, whatever the cache holds. A project with no log gets no cache file.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/sessions"
	"github.com/LastStep/Bonsai/internal/workspace"
)

func TestTheSessionsCheckFindsALogChangedSinceItsLastRun(t *testing.T) {
	e := setup(t)
	root := e.linked(t, "cache")
	clock(t, "2026-10-09 12:00")
	cache := filepath.Join(root, filepath.FromSlash(workspace.LocalDir), filepath.FromSlash(sessions.CacheFile))

	// No log yet: no finding, and no cache file.
	if r, err := checkLocal(t, root, e.home); err != nil || len(tablesWarnings(r)) != 0 {
		t.Fatalf("no log: %v %v", err, tablesWarnings(r))
	}
	if _, err := os.Stat(cache); !os.IsNotExist(err) {
		t.Fatalf("a project with no log has a cache file: %v", err)
	}

	writeSessionLog(t, root, sessA, startEv("09:05", "T-0901", "builder"), endEv("10:05"))
	if _, werr := WriteTables(root); werr != nil {
		t.Fatal(werr)
	}
	if r, err := checkLocal(t, root, e.home); err != nil || len(tablesWarnings(r)) != 0 {
		t.Fatalf("the table written: %v %v", err, tablesWarnings(r))
	}
	held, err := os.ReadFile(cache)
	if err != nil || !strings.Contains(string(held), sessA) {
		t.Fatalf("the cache after a run with a log: %v %q", err, held)
	}

	// The same log file gets a second session span: found, though the cache held the file.
	writeSessionLog(t, root, sessA, startEv("09:05", "T-0901", "builder"), endEv("10:05"),
		startEv("10:30", "T-0901", "builder"), endEv("11:00"))
	r, err := checkLocal(t, root, e.home)
	if w := tablesWarnings(r); err != nil || len(w) != 1 || !strings.Contains(w[0], "lacks 1 row") {
		t.Errorf("a log file changed since the last run: %v %v", err, w)
	}
	// And still found on the run after, from the cache it wrote.
	if r, _ = checkLocal(t, root, e.home); len(tablesWarnings(r)) != 1 {
		t.Errorf("the run after: %v", tablesWarnings(r))
	}

	// A new log file: found too, beside the first.
	writeSessionLog(t, root, sessB, startEv("11:10", "T-0901", "builder"), endEv("11:40"))
	if r, _ = checkLocal(t, root, e.home); len(tablesWarnings(r)) != 1 || !strings.Contains(tablesWarnings(r)[0], "lacks 2 row") {
		t.Errorf("a log file added since the last run: %v", tablesWarnings(r))
	}

	// A cache that says otherwise (written by hand, or by an older run over other bytes) is not trusted.
	if err := os.WriteFile(cache, held, 0o644); err != nil {
		t.Fatal(err)
	}
	if r, _ = checkLocal(t, root, e.home); len(tablesWarnings(r)) != 1 || !strings.Contains(tablesWarnings(r)[0], "lacks 2 row") {
		t.Errorf("a stale cache: %v", tablesWarnings(r))
	}

	// check --write reads through the cache too, and adds both rows.
	if res, werr := WriteTables(root); werr != nil || len(res.Written) != 1 {
		t.Fatalf("write: %v %+v", werr, res)
	}
	got := read(t, root, workspace.SessionsTableFile)
	if strings.Count(got, "\n| 6d1e2f3a |") != 2 || strings.Count(got, "\n| 7a2b3c4d |") != 1 {
		t.Errorf("the table:\n%s", got)
	}
	if r, _ = checkLocal(t, root, e.home); len(tablesWarnings(r)) != 0 {
		t.Errorf("after the write: %v", tablesWarnings(r))
	}
}
