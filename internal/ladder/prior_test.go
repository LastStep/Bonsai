package ladder

// The task's earlier result (V2's S3): read only when its bytes are the ones the log's last ladder record for the task
// fingerprints, since git integrity compares with it and the rise step takes a recorded base from it. A result changed
// by hand, an older genuine result put back, or a result with no record in the log is not read, and the human lines say
// so; the climb then compares with no earlier one.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/testpack"
)

func TestPriorReadOnlyWithItsFingerprint(t *testing.T) {
	resultPath := func(f *fx) string { return filepath.Join(f.root, ".bonsai", "local", "ladder", "T-0901.json") }
	read := func(t *testing.T, p string) []byte {
		t.Helper()
		raw, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		return raw
	}
	put := func(t *testing.T, p string, raw []byte) {
		t.Helper()
		if err := os.WriteFile(p, raw, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	stash := func(f *fx) {
		f.write("src/main.txt", "put aside\n")
		testpack.Git(f.t, f.root, "stash", "-q")
	}
	// judged climbs once more and reports whether git integrity compared with an earlier result (its stash-added
	// judged) and the human lines.
	judged := func(t *testing.T, f *fx) (string, string) {
		t.Helper()
		f.out.Reset()
		if a := f.climb("", Options{Task: "T-0901"}); a.Exit != 0 {
			t.Fatalf("exit %d\n%s", a.Exit, f.out.String())
		}
		return showOf(gitRung(t, f).Captures, "stash-added"), f.out.String()
	}
	notRead := "so it is not read: this climb compares with no earlier one"

	t.Run("the last climb's own result is read", func(t *testing.T) {
		f := vgProject(t)
		judged(t, f)
		stash(f)
		if got, out := judged(t, f); got != "1" || strings.Contains(out, "earlier result") {
			t.Errorf("stash-added %s\n%s", got, out)
		}
	})
	t.Run("a result changed by hand", func(t *testing.T) {
		f := vgProject(t)
		judged(t, f)
		raw := read(t, resultPath(f))
		forged := strings.Replace(string(raw), `"stash-count": 0`, `"stash-count": 1`, 1)
		if forged == string(raw) {
			t.Fatalf("no stash-count in the result:\n%s", raw)
		}
		put(t, resultPath(f), []byte(forged))
		stash(f)
		got, out := judged(t, f)
		if got != "null" || !strings.Contains(out, "note: the task's earlier result .bonsai/local/ladder/T-0901.json is not the one its "+
			"last ladder record in the log fingerprints (another file was put there), "+notRead) {
			t.Errorf("stash-added %s (the forged count would hide the stash)\n%s", got, out)
		}
	})
	t.Run("an older genuine result put back", func(t *testing.T) {
		f := vgProject(t)
		judged(t, f)
		older := read(t, resultPath(f))
		stash(f)
		if got, _ := judged(t, f); got != "1" {
			t.Fatalf("the second climb: stash-added %s", got)
		}
		put(t, resultPath(f), older)
		got, out := judged(t, f)
		if got != "null" || !strings.Contains(out, "is not the one its last ladder record in the log fingerprints") {
			t.Errorf("stash-added %s\n%s", got, out)
		}
	})
	t.Run("a result with no record in the log", func(t *testing.T) {
		f := vgProject(t)
		judged(t, f)
		if err := os.RemoveAll(filepath.Join(f.root, ".bonsai", "local", "log")); err != nil {
			t.Fatal(err)
		}
		stash(f)
		got, out := judged(t, f)
		if got != "null" || !strings.Contains(out, "note: the log holds no ladder record for the task's earlier result "+
			".bonsai/local/ladder/T-0901.json, "+notRead) {
			t.Errorf("stash-added %s\n%s", got, out)
		}
	})
}
