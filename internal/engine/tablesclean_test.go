package engine

// check --write cleans the sessions table's rows by bonsai.yaml's generated.sessions (internal/clean, step 5.2.6b):
// a row of a done task, past the rule and with its log file gone, leaves the table, and its clean record follows
// the write; a row whose span is still in the log, or whose task is not done, stays; a second write cleans nothing.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/record"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// setSessionsRule sets generated.sessions.keep_days in the project's bonsai.yaml, as a person would.
func setSessionsRule(t *testing.T, root, days string) {
	t.Helper()
	p := filepath.Join(root, workspace.ConfigFile)
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	i := strings.Index(s, "\n  sessions:")
	j := strings.Index(s[i+1:], "keep_days: null")
	if i < 0 || j < 0 {
		t.Fatalf("bonsai.yaml has no generated.sessions.keep_days:\n%s", s)
	}
	j += i + 1
	s = s[:j] + "keep_days: " + days + s[j+len("keep_days: null"):]
	if err := os.WriteFile(p, []byte(s), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestWriteTablesCleansSessionRows(t *testing.T) {
	e := setup(t)
	root := e.linked(t, "cleanrows")
	setSessionsRule(t, root, "1")
	writeFile(t, root, "work/tasks/T-0901-running.md", strings.Replace(f1Todo, "T-0902", "T-0901", 1))
	writeFile(t, root, "work/tasks/T-100000-done.md", f1Done)
	clock(t, "2026-10-08 12:00")
	logA := writeSessionLog(t, root, sessA, startEv("09:05", "T-100000", "orchestrator"), endEv("10:35"))
	writeSessionLog(t, root, sessB, startEv("09:00", "T-0901", "builder"), endEv("09:30"))
	if res, werr := WriteTables(root); werr != nil || len(res.Cleaned) != 0 {
		t.Fatalf("the first write: %v %+v", werr, res)
	}
	table := read(t, root, workspace.SessionsTableFile)
	if !strings.Contains(table, "| 6d1e2f3a |") || !strings.Contains(table, "| 7a2b3c4d |") {
		t.Fatalf("the rows were not added:\n%s", table)
	}

	// Two days on, A's row is past the rule but its log file is still there: it stays (else it would come back).
	clock(t, "2026-10-10 12:00")
	if res, werr := WriteTables(root); werr != nil || len(res.Cleaned) != 0 || len(res.Written) != 0 {
		t.Errorf("with the log file there: %v %+v", werr, res)
	}

	// Its log file gone (the session end's cleaner took it): A's row goes, B's (a running task) stays.
	if err := os.Remove(filepath.Join(root, filepath.FromSlash(logA))); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(root, ".bonsai", "local", "log", "s-"+sessB+".ndjson")); err != nil {
		t.Fatal(err)
	}
	res, werr := WriteTables(root)
	if werr != nil || len(res.Cleaned) != 1 || len(res.Written) != 1 || res.Written[0] != workspace.SessionsTableFile {
		t.Fatalf("the cleaning write: %v %+v", werr, res)
	}
	target := ".bonsai/sessions.md#6d1e2f3a@2026-10-08 09:05"
	if res.Cleaned[0] != target {
		t.Errorf("cleaned %v, want %s", res.Cleaned, target)
	}
	table = read(t, root, workspace.SessionsTableFile)
	if strings.Contains(table, "| 6d1e2f3a |") || !strings.Contains(table, "| 7a2b3c4d |") || strings.Contains(table, "orchestrator") {
		t.Errorf("the table after the cleaning:\n%s", table)
	}
	notes := strings.Join(res.Notes(), "\n")
	if !strings.Contains(notes, "cleaned 1 row(s) of .bonsai/sessions.md by bonsai.yaml's generated.sessions") {
		t.Errorf("notes %q", notes)
	}
	lf, err := record.ReadLog(filepath.Join(root, ".bonsai", "local", "log", "w-2026-10-10.ndjson"))
	if err != nil {
		t.Fatal(err)
	}
	if len(lf.Records) != 1 || lf.Records[0].Event != "clean" || lf.Records[0].Session != nil ||
		*lf.Records[0].Target != target || *lf.Records[0].Reason != "generated.sessions.keep_days=1" ||
		*lf.Records[0].Checkout != filepath.Base(root) {
		t.Errorf("the clean record: %+v", lf.Records)
	}
	// No stale warning: the row cleaned is no span in the log.
	if r, err := checkLocal(t, root, e.home); err != nil || len(tablesWarnings(r)) != 0 {
		t.Errorf("after the cleaning: %v %v", err, tablesWarnings(r))
	}
	// A second write cleans nothing and writes no record.
	if res, werr = WriteTables(root); werr != nil || len(res.Cleaned) != 0 || len(res.Written) != 0 {
		t.Errorf("the second write: %v %+v", werr, res)
	}
	if lf, _ = record.ReadLog(filepath.Join(root, ".bonsai", "local", "log", "w-2026-10-10.ndjson")); len(lf.Records) != 1 {
		t.Errorf("records after the second write: %d", len(lf.Records))
	}
}
