package engine

// The personal layer's warnings (checkpersonal.go) on scratch homes, the memory caps' one home (MemoryCaps) against
// every text that writes a cap's number, and init's and update's step (PersonalAfterWrite).

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// madeUpKey is a secret-shaped decoy, made up: the redactor reads it as a GitHub token.
func madeUpKey() string { return "ghp_" + strings.Repeat("Hx4Pm9", 6) }

func personalNote(id, kind, body string) string {
	return "---\nformat: bonsai.memory/1\nid: " + id + "\ntitle: A note\nkind: " + kind +
		"\nupdated: 2026-10-18\nsource: \"the person\"\nlabels: {}\n---\n\n" + body
}

func personalIndex(lines ...string) string {
	return "---\nformat: bonsai.memory/1\nid: null\ntitle: Personal memory\nkind: index\nupdated: 2026-10-18\nsource: null\nlabels: {}\n---\n\n" +
		strings.Join(lines, "\n") + "\n"
}

func putFile(t *testing.T, p, s string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(s), 0o644); err != nil {
		t.Fatal(err)
	}
}

// personalWarnings runs the personal layer's check on a home and gives its warnings, and fails on any finding.
func personalWarnings(t *testing.T, home string) []Finding {
	t.Helper()
	r := &CheckResult{Home: home}
	r.checkPersonal()
	if len(r.Findings) > 0 {
		t.Errorf("the personal layer gave findings, which change the exit code: %+v", r.Findings)
	}
	return r.Warnings
}

func codesOf(fs []Finding) string {
	var out []string
	for _, f := range fs {
		out = append(out, f.Code)
	}
	return strings.Join(out, " ")
}

// homeSum hashes every file under a folder, path and bytes, to show a call wrote nothing.
func homeSum(t *testing.T, dir string) string {
	t.Helper()
	var lines []string
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(dir, p)
		if d.IsDir() {
			lines = append(lines, "d "+filepath.ToSlash(rel))
			return nil
		}
		raw, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		h := sha256.Sum256(raw)
		lines = append(lines, "f "+filepath.ToSlash(rel)+" "+hex.EncodeToString(h[:]))
		return nil
	})
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	sort.Strings(lines)
	return strings.Join(lines, "\n")
}

// TestPersonalWarnings: each warning from a fixture home, each a warning naming no project file and a person's
// step; a clean layer and a home with no personal/ give none; the secret's value is in no output; check writes
// nothing in the home.
func TestPersonalWarnings(t *testing.T) {
	if got := personalWarnings(t, t.TempDir()); len(got) != 0 {
		t.Errorf("a home with no personal/: %+v", got)
	}
	clean := t.TempDir()
	putFile(t, filepath.Join(clean, "personal", "INDEX.md"), personalIndex("- [Short replies](notes/M-short.md): how they like answers"))
	putFile(t, filepath.Join(clean, "personal", "notes", "M-short.md"), personalNote("M-short", "feedback", "Short replies.\n"))
	if got := personalWarnings(t, clean); len(got) != 0 {
		t.Errorf("a clean layer: %+v", got)
	}

	cases := []struct {
		name, want, says string
		setup            func(home string)
	}{
		{"index not the format", "personal-format", "INDEX.md does not read as bonsai.memory/1", func(h string) {
			putFile(t, filepath.Join(h, "personal", "INDEX.md"), "just text\n")
		}},
		{"index kind", "personal-format", "has kind feedback, not index", func(h string) {
			putFile(t, filepath.Join(h, "personal", "INDEX.md"), strings.Replace(personalIndex(), "kind: index", "kind: feedback", 1))
		}},
		{"a note not the format", "personal-format", "notes/M-bad.md does not read as bonsai.memory/1", func(h string) {
			putFile(t, filepath.Join(h, "personal", "INDEX.md"), personalIndex("- [Bad](notes/M-bad.md)"))
			putFile(t, filepath.Join(h, "personal", "notes", "M-bad.md"), "---\nformat: bonsai.memory/1\nkind: [\n---\n")
		}},
		{"index over its lines", "personal-budget", "is 51 lines", func(h string) {
			putFile(t, filepath.Join(h, "personal", "INDEX.md"), personalIndex(strings.Repeat("- a line\n", 40)))
		}},
		{"index over its bytes", "personal-budget", "over its 40 lines or 4 KB", func(h string) {
			putFile(t, filepath.Join(h, "personal", "INDEX.md"), personalIndex(strings.Repeat("x", 4200)))
		}},
		{"a note over its bytes", "personal-budget", "over its 4 KB", func(h string) {
			putFile(t, filepath.Join(h, "personal", "INDEX.md"), personalIndex("- [Long](notes/M-long.md)"))
			putFile(t, filepath.Join(h, "personal", "notes", "M-long.md"), personalNote("M-long", "reference", strings.Repeat("y", 4200)))
		}},
		{"a line naming no note", "personal-index", "names the note notes/M-gone.md, which does not exist", func(h string) {
			putFile(t, filepath.Join(h, "personal", "INDEX.md"), personalIndex("- [Gone](notes/M-gone.md)"))
		}},
		{"a note with no line", "personal-index", "notes/M-lone.md has no line in the personal memory index", func(h string) {
			putFile(t, filepath.Join(h, "personal", "INDEX.md"), personalIndex())
			putFile(t, filepath.Join(h, "personal", "notes", "M-lone.md"), personalNote("M-lone", "feedback", "Alone.\n"))
		}},
		{"a secret", "personal-secret", "holds a secret-shaped string on line 12 (github-token)", func(h string) {
			putFile(t, filepath.Join(h, "personal", "INDEX.md"), personalIndex("- [Key](notes/M-key.md)"))
			putFile(t, filepath.Join(h, "personal", "notes", "M-key.md"), personalNote("M-key", "reference", "A deploy.\nIt used "+madeUpKey()+" once.\n"))
		}},
		{"a refresh that failed", "personal-copy", "failed at 2026-10-18T09:00:00Z: the source did not answer", func(h string) {
			putFile(t, filepath.Join(h, "cache", "personal-copy.json"), `{"source": "//wsl.localhost/D/home/u/.bonsai/personal", "refreshed": null, "files": [], "failed": "2026-10-18T09:00:00Z", "why": "the source did not answer"}`)
		}},
		{"a copy changed since its refresh", "personal-copy", "changed since its refresh", func(h string) {
			putFile(t, filepath.Join(h, "personal", "INDEX.md"), personalIndex())
			putFile(t, filepath.Join(h, "cache", "personal-copy.json"), `{"source": "//wsl.localhost/D/home/u/.bonsai/personal", "refreshed": "2026-10-18T09:00:00Z", "files": [], "failed": null, "why": null}`)
		}},
		{"a record that does not read", "personal-copy", "cannot be read", func(h string) {
			putFile(t, filepath.Join(h, "cache", "personal-copy.json"), "{")
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			home := t.TempDir()
			c.setup(home)
			before := homeSum(t, home)
			got := personalWarnings(t, home)
			if after := homeSum(t, home); after != before {
				t.Errorf("check wrote in the home:\n%s\n---\n%s", before, after)
			}
			found := false
			for _, w := range got {
				out := w.Message + " " + w.Next
				if strings.Contains(out, madeUpKey()) || strings.Contains(out, madeUpKey()[4:12]) {
					t.Errorf("the secret's value is printed: %s", out)
				}
				if w.File != "" || w.Who != "person" || ascii(out) != out {
					t.Errorf("%s: file %q, who %s, or not ASCII", w.Code, w.File, w.Who)
				}
				if w.Code == c.want && strings.Contains(w.Message, c.says) {
					found = true
				}
			}
			if !found {
				t.Errorf("no %s saying %q: %+v", c.want, c.says, got)
			}
		})
	}
}

// TestPersonalCopyNeverRefreshed: with personal_from set (read on Windows only) and no record of a refresh from it,
// check warns; the Linux half of this form is the end run's (personal_from is not read off Windows).
func TestPersonalCopyFromSettings(t *testing.T) {
	home := t.TempDir()
	putFile(t, filepath.Join(home, "settings.json"), `{"personal_from": "//wsl.localhost/D/home/u/.bonsai/personal"}`+"\n")
	got := codesOf(personalWarnings(t, home))
	if hs, _ := workspace.LoadHomeSettings(home); hs.PersonalFrom != "" {
		if got != "personal-copy" {
			t.Errorf("personal_from set, never refreshed: %q", got)
		}
	} else if got != "" {
		t.Errorf("personal_from is not read here, so nothing is said: %q", got)
	}
}

// TestPersonalAfterWrite: init's and update's step writes the empty index once, never over one, and says so.
func TestPersonalAfterWrite(t *testing.T) {
	home := t.TempDir()
	clock := func() time.Time { return time.Date(2026, 10, 18, 9, 0, 0, 0, time.UTC) }
	lines := PersonalAfterWrite(home, clock)
	if len(lines) != 1 || lines[0].Warn || !strings.Contains(lines[0].Text, "wrote the empty personal memory index") {
		t.Fatalf("first: %+v", lines)
	}
	raw, err := os.ReadFile(filepath.Join(home, "personal", "INDEX.md"))
	if err != nil {
		t.Fatal(err)
	}
	if m, err := format.ReadMemory(raw); err != nil || m.Kind != "index" || m.ID != nil || m.Updated != "2026-10-18" {
		t.Errorf("the empty index does not read as bonsai.memory/1's index: %+v %v", m, err)
	}
	if got := personalWarnings(t, home); len(got) != 0 {
		t.Errorf("the empty index gives warnings: %+v", got)
	}
	before := homeSum(t, home)
	if lines := PersonalAfterWrite(home, clock); len(lines) != 0 {
		t.Errorf("second: %+v", lines)
	}
	if homeSum(t, home) != before {
		t.Errorf("the second run changed the home")
	}
	if lines := PersonalAfterWrite("", clock); lines != nil {
		t.Errorf("no home: %+v", lines)
	}
	bad := t.TempDir()
	putFile(t, filepath.Join(bad, "settings.json"), "[1]\n")
	if lines := PersonalAfterWrite(bad, clock); len(lines) != 2 || !lines[0].Warn {
		t.Errorf("an unreadable home settings file: %+v", lines)
	}
}

// TestMemoryCapsHeld: every text that writes a cap's number writes the table's: base's memory skill, the check words'
// meanings, and the findings' texts (by their builders, which read the table).
func TestMemoryCapsHeld(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "packs", "base", "skills", "memory", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	skill := strings.Join(strings.Fields(string(raw)), " ")
	pi, ni, note := memoryCap("project-index"), memoryCap("personal-index"), memoryCap("note")
	for _, want := range []string{
		fmt.Sprintf("a note at most %s; the index at most %d lines and %s", KB(note.Bytes), pi.Lines, KB(pi.Bytes)),
		fmt.Sprintf("the personal index at most %d lines and %s, a note at most %s", ni.Lines, KB(ni.Bytes), KB(note.Bytes)),
	} {
		if !strings.Contains(skill, want) {
			t.Errorf("base's memory skill does not say %q", want)
		}
	}
	// Every number written next to "lines" or "KB" in the skill is one of the table's.
	known := map[string]bool{}
	for _, c := range MemoryCaps {
		known[fmt.Sprint(c.Lines)+" lines"] = c.Lines > 0
		known[KB(c.Bytes)] = true
	}
	for _, m := range regexp.MustCompile(`\b\d+ (?:lines|KB)\b`).FindAllString(skill, -1) {
		if !known[m] {
			t.Errorf("base's memory skill writes %q, which is no memory cap", m)
		}
	}
	words := map[string]string{
		"memory-index-size": fmt.Sprintf("over its fixed %d lines or %s", pi.Lines, KB(pi.Bytes)),
		"memory-note-size":  "over " + KB(note.Bytes),
	}
	for code, want := range words {
		w, _ := format.CheckWord(code)
		if !strings.Contains(w.Means, want) {
			t.Errorf("the check word %s's meaning %q does not say %q", code, w.Means, want)
		}
	}
	for _, w := range format.CheckWords {
		if strings.HasPrefix(w.Word, "personal-") && regexp.MustCompile(`\d+ (lines|KB)`).MatchString(w.Means) {
			t.Errorf("%s writes a cap's number: the personal layer's caps are named on the reference page only", w.Word)
		}
	}
}
