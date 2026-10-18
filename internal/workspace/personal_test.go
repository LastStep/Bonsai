package workspace

// The personal layer's index and the Windows copy (personal.go; plan-5 5.6.3, 1 and 4), on scratch homes. The copy
// is refreshed on Windows only in Bonsai's commands (personal_from is a Windows key); RefreshPersonal itself reads
// any folder, so these tests run it on every system with a scratch source.

import (
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

var personalDay = time.Date(2026, 10, 18, 9, 0, 0, 0, time.UTC)

func personalClock() time.Time { return personalDay }

func writeTestFile(t *testing.T, p, s string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(s), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readTestFile(t *testing.T, p string) string {
	t.Helper()
	raw, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

// TestEnsurePersonalIndex: the empty index is written once, in the format, and never over what is there.
func TestEnsurePersonalIndex(t *testing.T) {
	home := t.TempDir()
	wrote, err := EnsurePersonalIndex(home, personalDay)
	if err != nil || !wrote {
		t.Fatalf("first: %v %v", wrote, err)
	}
	p := filepath.Join(home, "personal", "INDEX.md")
	got := readTestFile(t, p)
	for _, want := range []string{"format: bonsai.memory/1   # fields:", "id: null\n", "title: Personal memory\n", "kind: index\n",
		"updated: 2026-10-18\n", "source: null\n", "labels: {}\n", "written by the person"} {
		if !strings.Contains(got, want) {
			t.Errorf("the index lacks %q:\n%s", want, got)
		}
	}
	writeTestFile(t, p, "mine\n")
	if wrote, err := EnsurePersonalIndex(home, personalDay); err != nil || wrote || readTestFile(t, p) != "mine\n" {
		t.Errorf("second: wrote %v, %v, %q", wrote, err, readTestFile(t, p))
	}
}

// personalSource makes a source layer with an index and two notes.
func personalSource(t *testing.T) string {
	t.Helper()
	src := filepath.Join(t.TempDir(), "wsl", "personal")
	writeTestFile(t, filepath.Join(src, "INDEX.md"), "index\n")
	writeTestFile(t, filepath.Join(src, "notes", "M-a.md"), "a\n")
	writeTestFile(t, filepath.Join(src, "notes", "M-b.md"), "b\n")
	writeTestFile(t, filepath.Join(src, "notes", "other.txt"), "not a note\n")
	return src
}

// TestRefreshPersonal: a refresh copies the layer and records it; a second finds it the same; an edited copy differs
// from the record; a change at the source is brought over, a note the source dropped removed.
func TestRefreshPersonal(t *testing.T) {
	home, src := t.TempDir(), personalSource(t)
	r := RefreshPersonal(home, src, personalClock)
	if r.Failed != "" || r.Copied != 3 || r.Removed != 0 {
		t.Fatalf("first refresh %+v", r)
	}
	if got := readTestFile(t, filepath.Join(home, "personal", "notes", "M-b.md")); got != "b\n" {
		t.Errorf("M-b %q", got)
	}
	if _, err := os.Stat(filepath.Join(home, "personal", "notes", "other.txt")); err == nil {
		t.Errorf("a file that is not a note was copied")
	}
	rec, err := ReadPersonalRecord(home)
	if err != nil || rec == nil || rec.Source != src || rec.Refreshed == nil || *rec.Refreshed != "2026-10-18T09:00:00Z" ||
		rec.Failed != nil || len(rec.Files) != 3 || rec.Files[0].Path != "INDEX.md" || rec.Files[2].Path != "notes/M-b.md" {
		t.Fatalf("record %+v %v", rec, err)
	}
	if !strings.HasSuffix(readTestFile(t, filepath.Join(home, "cache", "personal-copy.json")), "\n") {
		t.Errorf("the record does not end in a line end")
	}
	if r := RefreshPersonal(home, src, personalClock); !r.Same || r.Failed != "" || !strings.Contains(r.Line(), "is the same as") {
		t.Errorf("second refresh %+v %s", r, r.Line())
	}
	have, _ := PersonalFiles(PersonalPath(home))
	if same, _ := SamePersonal(have, rec.Files); !same {
		t.Errorf("the copy reads as changed right after its refresh")
	}
	writeTestFile(t, filepath.Join(home, "personal", "notes", "M-a.md"), "edited here\n")
	have, _ = PersonalFiles(PersonalPath(home))
	if same, differs := SamePersonal(have, rec.Files); same || differs != "notes/M-a.md" {
		t.Errorf("an edited copy: %v %s", same, differs)
	}
	writeTestFile(t, filepath.Join(home, "personal", "notes", "M-extra.md"), "added here\n")
	have, _ = PersonalFiles(PersonalPath(home))
	if same, differs := SamePersonal(have, rec.Files); same || differs == "" {
		t.Errorf("a note added to the copy: %v %s", same, differs)
	}
	writeTestFile(t, filepath.Join(src, "notes", "M-b.md"), "b, changed at the source\n")
	if err := os.Remove(filepath.Join(src, "notes", "M-a.md")); err != nil {
		t.Fatal(err)
	}
	r = RefreshPersonal(home, src, personalClock)
	if r.Failed != "" || r.Copied != 1 || r.Removed != 2 || !strings.Contains(r.Line(), "1 written, 2 removed") {
		t.Fatalf("refresh after the source changed %+v %s", r, r.Line())
	}
	for _, gone := range []string{"M-a.md", "M-extra.md"} {
		if _, err := os.Stat(filepath.Join(home, "personal", "notes", gone)); err == nil {
			t.Errorf("%s is still in the copy", gone)
		}
	}
	if got := readTestFile(t, filepath.Join(home, "personal", "notes", "M-b.md")); got != "b, changed at the source\n" {
		t.Errorf("M-b %q", got)
	}
}

// TestRefreshPersonalRefuses: the home's own personal/ (or a folder in it) is refused; a note over 64 KB, more than
// 500 notes, or a source with no index copies nothing; each recorded as a failure, the earlier copy kept.
func TestRefreshPersonalRefuses(t *testing.T) {
	home := t.TempDir()
	for _, own := range []string{PersonalPath(home), filepath.Join(PersonalPath(home), "notes")} {
		if r := RefreshPersonal(home, own, personalClock); !strings.Contains(r.Failed, "the copy's target") {
			t.Errorf("%s: %+v", own, r)
		}
	}
	rec, _ := ReadPersonalRecord(home)
	if rec == nil || rec.Failed == nil || rec.Why == nil || !strings.Contains(*rec.Why, "target") {
		t.Fatalf("the refusal is not recorded: %+v", rec)
	}

	src := personalSource(t)
	if r := RefreshPersonal(home, src, personalClock); r.Failed != "" {
		t.Fatal(r.Failed)
	}
	big := filepath.Join(src, "notes", "M-big.md")
	writeTestFile(t, big, strings.Repeat("x", 65*1024))
	r := RefreshPersonal(home, src, personalClock)
	if !strings.Contains(r.Failed, "notes/M-big.md in the source is 66560 bytes, over the copy's 64 KB a file: nothing was copied") {
		t.Errorf("a 65 KB note: %+v", r)
	}
	if _, err := os.Stat(filepath.Join(home, "personal", "notes", "M-big.md")); err == nil {
		t.Errorf("the 65 KB note was copied")
	}
	rec, _ = ReadPersonalRecord(home)
	if rec.Failed == nil || rec.Refreshed == nil || len(rec.Files) != 3 {
		t.Errorf("a failed refresh keeps the last good one's files and time: %+v", rec)
	}
	if err := os.Remove(big); err != nil {
		t.Fatal(err)
	}
	for i := 0; i <= PersonalNotesMax; i++ {
		writeTestFile(t, filepath.Join(src, "notes", "M-n"+itoa(i)+".md"), "n\n")
	}
	if r := RefreshPersonal(home, src, personalClock); !strings.Contains(r.Failed, "notes, over the copy's 500: nothing was copied") {
		t.Errorf("501 notes and more: %+v", r)
	}
	empty := filepath.Join(t.TempDir(), "personal")
	if err := os.MkdirAll(empty, 0o755); err != nil {
		t.Fatal(err)
	}
	if r := RefreshPersonal(home, empty, personalClock); !strings.Contains(r.Failed, "the source holds no INDEX.md") {
		t.Errorf("no index: %+v", r)
	}
	if r := RefreshPersonal(home, filepath.Join(t.TempDir(), "gone"), personalClock); !strings.Contains(r.Failed, "cannot be read") {
		t.Errorf("no source: %+v", r)
	}
}

// TestRefreshPersonalLink: a note that is a link is left out and named, never followed. Making a link needs
// Developer Mode or an administrator on Windows, so there the test skips when the link cannot be made.
func TestRefreshPersonalLink(t *testing.T) {
	home, src := t.TempDir(), personalSource(t)
	outside := filepath.Join(t.TempDir(), "secret.md")
	writeTestFile(t, outside, "outside the layer\n")
	if err := os.Symlink(outside, filepath.Join(src, "notes", "M-link.md")); err != nil {
		if runtime.GOOS == "windows" {
			t.Skip("a link cannot be made here (Windows needs Developer Mode or an administrator):", err)
		}
		t.Fatal(err)
	}
	r := RefreshPersonal(home, src, personalClock)
	if r.Failed != "" || r.Copied != 3 || len(r.Skipped) != 1 || !strings.Contains(r.Line(), "left out: notes/M-link.md (not a regular file") {
		t.Fatalf("%+v %s", r, r.Line())
	}
	if _, err := os.Lstat(filepath.Join(home, "personal", "notes", "M-link.md")); err == nil {
		t.Errorf("the link was copied")
	}
}

// TestRefreshPersonalSlow: a source that never answers stops the refresh within the wait, recorded as a failure.
func TestRefreshPersonalSlow(t *testing.T) {
	home := t.TempDir()
	never := make(chan struct{})
	defer close(never)
	defer func(l func(string) (fs.FileInfo, error), w time.Duration) { sourceLstat, PersonalReadWait = l, w }(sourceLstat, PersonalReadWait)
	sourceLstat = func(string) (fs.FileInfo, error) { <-never; return nil, fs.ErrNotExist }
	PersonalReadWait = 100 * time.Millisecond
	start := time.Now()
	r := RefreshPersonal(home, "//wsl.localhost/Distro/home/someone/.bonsai/personal", personalClock)
	if took := time.Since(start); took > 2*time.Second {
		t.Errorf("the refresh took %v", took)
	}
	if !strings.Contains(r.Failed, "did not answer within 100 ms: nothing was compared or copied") {
		t.Errorf("%+v", r)
	}
	rec, _ := ReadPersonalRecord(home)
	if rec == nil || rec.Failed == nil || rec.Why == nil || !strings.Contains(*rec.Why, "did not answer") {
		t.Errorf("not recorded: %+v", rec)
	}
}
