package engine

// The pack command's second stage in the engine (packpicks.go; T-6011, the plan P-T-6009's stage-2 proof rows): a
// moved pack's commit titles between two commits, limited to its folder, and a read that fails said in the preview;
// the downgrade order; the commit lines. The command's own tests are cmd/bonsai/packpicks_test.go.

import (
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/testpack"
	"github.com/LastStep/Bonsai/internal/textui"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// What changed: from the first commit to the third lists two titles, newest first; a pack in a folder lists only
// the commits that touch it; a commit the clone does not hold is said, never refused, and the preview still prints.
func TestPackChangesBetweenCommits(t *testing.T) {
	tmp := t.TempDir()
	testpack.Isolate(t, tmp)
	home := t.TempDir()
	src, shas := testpack.Fixture(t, tmp, "two-folders",
		map[string]string{"a/one.md": "1\n", "b/one.md": "1\n"},
		map[string]string{"b/one.md": "2\n"},
		map[string]string{"a/one.md": "3\n"})
	c := cache{home: home}
	for _, s := range []string{shas[0], shas[2]} {
		if _, err := c.fetch(src, s); err != nil {
			t.Fatal(err)
		}
	}
	titles, more, err := c.packChanges(src, "", shas[0], shas[2])
	if err != nil || strings.Join(titles, "|") != "commit 3|commit 2" || more != 0 {
		t.Errorf("the whole repository: %q %d %v", titles, more, err)
	}
	titles, _, err = c.packChanges(src, "a", shas[0], shas[2])
	if err != nil || strings.Join(titles, "|") != "commit 3" {
		t.Errorf("the folder a: %q %v", titles, err)
	}
	// A move back lists none; the plan's changes and its preview.
	missing := strings.Repeat("1", 40)
	p := &Plan{Command: "update", Home: home, Config: &workspace.Config{Packs: []workspace.PackRef{{ID: "two", Source: src, Path: "a"}}},
		Packs: []PackMove{{ID: "two", Source: src, Version: "0.1.0", From: shas[0], To: shas[2]}, {ID: "gone", Source: src, From: missing, To: shas[2]},
			{ID: "back", Source: src, From: shas[2], To: shas[0]}, {ID: "same", Source: src, From: shas[2], To: shas[2]}}}
	p.FillPackChanges()
	if len(p.PackChanges) != 3 || strings.Join(p.PackChanges[0].Titles, "|") != "commit 3" || p.PackChanges[0].Why != nil ||
		p.PackChanges[1].Why == nil || !strings.Contains(*p.PackChanges[1].Why, "is not in this machine's clone") || len(p.PackChanges[1].Titles) != 0 ||
		p.PackChanges[2].Why != nil || len(p.PackChanges[2].Titles) != 0 {
		t.Errorf("the plan's changes: %+v", p.PackChanges)
	}
	var d textui.Doc
	p.PreviewTo(&d, false, false)
	if s := strings.Join(strings.Fields(d.String()), " "); !strings.Contains(s, "Changes two commit 3") ||
		!strings.Contains(s, "gone its changes could not be read: the commit 1111111 is not in this machine's clone") ||
		!strings.Contains(s, "back no commit of its folder between the two") || !strings.Contains(s, "Files") {
		t.Errorf("the preview:\n%s", d.String())
	}
	// The changes output and the fingerprint carry them.
	ch := p.Changes("preview", nil)
	if len(ch.PackChanges) != 3 || ch.PackChanges[0].Pack != "two" {
		t.Errorf("pack_changes: %+v", ch.PackChanges)
	}
	fp := p.Fingerprint()
	p.PackChanges[0].Titles = []string{"another"}
	if p.Fingerprint() == fp {
		t.Error("the fingerprint does not cover pack_changes")
	}
}

// The downgrade guard's order: a release tag below the locked version goes back; a commit, a branch and a tag with no
// version number have no order.
func TestDowngrade(t *testing.T) {
	for _, c := range []struct {
		version, ref  string
		down, ordered bool
	}{
		{"1.0.1", "v1.0.0", true, true}, {"1.0.1", "base-v1.0.0", true, true}, {"1.0.1", "base-v0.9.9", true, true},
		{"1.0.1", "v1.0.1", false, true}, {"1.0.1", "v1.0.2", false, true}, {"1.0.1", "workflow/v2.0.0", false, true},
		{"1.0.1", strings.Repeat("a", 40), false, false}, {"1.0.1", "main", false, false}, {"1.0.1", "edge", false, false},
		{"not-a-version", "v1.0.0", false, false},
	} {
		if down, ordered := Downgrade(c.version, c.ref); down != c.down || ordered != c.ordered {
			t.Errorf("Downgrade(%q, %q) = %v, %v; want %v, %v", c.version, c.ref, down, ordered, c.down, c.ordered)
		}
	}
}

// The commit lines: git add with every file written, in the plan's order and the lock last, forward slashes; then
// git commit naming each pack moved; never joined by &&; none when nothing was written.
func TestCommitLines(t *testing.T) {
	p := &Plan{Command: "update", LockWrite: true, Config: &workspace.Config{Packs: []workspace.PackRef{{ID: "base", Ref: "base-v1.0.1"}, {ID: "wf", Ref: "v1.0.2"}}},
		Files: []*FileResult{{Path: workspace.ConfigFile, Result: Updated, write: []byte("x")}, {Path: "docs/a b.md", Result: Created, write: []byte("y")},
			{Path: "kept.md", Result: Unchanged}, {Path: "gone/old.md", Result: Removed, remove: true}},
		Packs: []PackMove{{ID: "base", From: "a", To: "b"}, {ID: "wf", From: "c", To: "d"}, {ID: "same", From: "e", To: "e"}, {ID: "out", From: "f"}}}
	got := p.CommitLines()
	want := []string{"git add -- bonsai.yaml 'docs/a b.md' gone/old.md .bonsai/lock.json",
		`git commit -m 'bonsai: base to base-v1.0.1, wf to v1.0.2, out taken out'`}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("CommitLines:\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
	for _, l := range got {
		if strings.Contains(l, "&&") || strings.Contains(l, `\`) {
			t.Errorf("a line joined or with a backslash: %s", l)
		}
	}
	p.Packs = nil
	if got := p.CommitLines(); got[1] != `git commit -m 'bonsai: update'` {
		t.Errorf("no pack moved: %q", got)
	}
	if got := (&Plan{Command: "update", Files: []*FileResult{{Path: "kept.md", Result: Unchanged}}}).CommitLines(); got != nil {
		t.Errorf("nothing written: %q", got)
	}
	// The changes output: commit_lines only once the command set them.
	p.CommitLinesOut = []string{"git add x", "git commit -m \"y\""}
	if ch := p.Changes("applied", nil); len(ch.CommitLines) != 2 {
		t.Errorf("commit_lines: %v", ch.CommitLines)
	}
}

// strictWords splits a printed command line the way neither sh nor PowerShell 5.1 can read otherwise: words apart by
// one space, each bare ([A-Za-z0-9._/-] only) or in ASCII single quotes holding printable ASCII (a backtick and a
// double quote are literal there in both shells; a doubled single quote is one quote to PowerShell, one word to sh). ok false for a line
// holding anything else (a byte outside printable ASCII, an unquoted ; $ ` & | ( ) or #, a smart quote, a newline):
// such a line could be more than its words.
func strictWords(line string) (words []string, ok bool) {
	bare := func(c byte) bool {
		return c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || strings.IndexByte("._/-", c) >= 0
	}
	for i := 0; i < len(line); {
		switch {
		case line[i] == '\'':
			// In single quotes; '' inside is one quote to PowerShell and two quoted strings run together to sh:
			// one word either way.
			var w strings.Builder
			j := i + 1
			for {
				if j >= len(line) {
					return nil, false
				}
				if line[j] == '\'' {
					if j+1 < len(line) && line[j+1] == '\'' {
						w.WriteByte('\'')
						j += 2
						continue
					}
					break
				}
				if line[j] < 0x20 || line[j] > 0x7e {
					return nil, false
				}
				w.WriteByte(line[j])
				j++
			}
			words = append(words, w.String())
			i = j + 1
		default:
			j := i
			for j < len(line) && bare(line[j]) {
				j++
			}
			if j == i {
				return nil, false
			}
			words = append(words, line[i:j])
			i = j
		}
		if i < len(line) {
			if line[i] != ' ' || i+1 == len(line) || line[i+1] == ' ' {
				return nil, false
			}
			i++
		}
	}
	return words, true
}

// ShellArg, which every printed command line's arguments go through (next steps naming a path, the command lines
// a refusal repeats): a hostile word is always one word in printable ASCII; a printable ASCII one keeps its bytes
// (B3: a backtick and a double quote among them, so a next step names the file that is there); a plain one stays
// bare.
func TestShellArgHostile(t *testing.T) {
	for _, s := range []string{"a;b", "$(id)", "`id`", "docs/a`b.md", "a b", "it's", `"q"`, `docs/a"b.md`, "a&b|c", "x,y", "@x", "a*b",
		"-n.md", "#x", "!x"} {
		got := ShellArg(s)
		if words, ok := strictWords(got); !ok || len(words) != 1 || words[0] != s {
			t.Errorf("ShellArg(%q) = %q: not one word with the same bytes (%q, %v)", s, got, words, ok)
		}
	}
	for _, s := range []string{"docs/x’;Write-Output INJECTED;#.md", "a\nb", "‘x‘", "a b"} {
		got := ShellArg(s)
		if words, ok := strictWords(got); !ok || len(words) != 1 {
			t.Errorf("ShellArg(%q) = %q: not one plain word", s, got)
		}
	}
	for _, s := range []string{"docs/a.md", "base@base-v1.0.1", "--keep", "k=v", "ws-abc"} {
		if got := ShellArg(s); got != s {
			t.Errorf("ShellArg(%q) = %q, want it bare", s, got)
		}
	}
}

// B1 (the verifier of 29 Oct): a hostile pack file path never gives a printed git add line that is more than one
// argument for it, in sh or PowerShell 5.1, and the line stays printable ASCII; a path it cannot print safely leaves
// the commit lines out (the command then says to commit by hand). The commit message is single-quoted, its ref passed
// through the same rule (no ! or $ a shell would read).
func TestCommitLinesHostilePaths(t *testing.T) {
	plan := func(path string) *Plan {
		return &Plan{Command: "update", LockWrite: true, Config: &workspace.Config{}, Files: []*FileResult{{Path: path, Result: Created, write: []byte("x")}}}
	}
	// Printed: each path one argument after --, its bytes unchanged (a plain path among them, so holding every path back
	// fails this; one starting with - is never an option, after --).
	for _, path := range []string{"docs/plain.md", "docs/a;rm -rf x.md", "docs/$(id).md", "docs/a b.md", "docs/a&b|c.md", "-n.md",
		"--pathspec-from-file=docs/list.txt", "docs/#x.md", "docs/a!b.md"} {
		p := plan(path)
		lines := p.CommitLines()
		if lines == nil {
			t.Errorf("%q: no commit lines (%s)", path, p.CommitLinesHeld())
			continue
		}
		words, ok := strictWords(lines[0])
		// update wrote the lock: bonsai.yaml first, then the file, then the lock.
		if !ok || len(words) != 6 || words[0] != "git" || words[1] != "add" || words[2] != "--" || words[3] != workspace.ConfigFile ||
			words[4] != path || words[5] != workspace.LockFile {
			t.Errorf("%q: git add line %q reads as %q (ok %v)", path, lines[0], words, ok)
		}
	}
	// Held: no line, and the reason (a byte outside printable ASCII, a quote character, git's pattern magic).
	for _, path := range []string{"docs/x’;Write-Output INJECTED;#.md", "docs/`id`.md", "docs/a\nb.md", "docs/it's.md", `docs/"q".md`,
		"docs/[ab].md", "docs/*.md", "docs/a?.md", "docs/a]b.md", ":(glob)docs", ":docs/a.md"} {
		p := plan(path)
		if lines := p.CommitLines(); lines != nil || p.CommitLinesHeld() == "" {
			t.Errorf("%q: printed %q, want held", path, lines)
		}
	}
	p := &Plan{Command: "update", LockWrite: true, Config: &workspace.Config{Packs: []workspace.PackRef{{ID: "base", Ref: "v1!$(id)"}, {ID: "wf", Ref: "v1.0.2"}}},
		Files: []*FileResult{{Path: "base/a.md", Result: Updated, write: []byte("x")}},
		Packs: []PackMove{{ID: "base", From: "a", To: "b"}, {ID: "wf", From: "c", To: "d"}}}
	lines := p.CommitLines()
	words, ok := strictWords(lines[1])
	if !ok || len(words) != 4 || words[0] != "git" || words[1] != "commit" || words[2] != "-m" || !strings.HasPrefix(lines[1], "git commit -m '") ||
		strings.ContainsAny(words[3], "!$`\"'") || !strings.Contains(words[3], "wf to v1.0.2") {
		t.Errorf("the commit message: %q reads as %q (ok %v)", lines[1], words, ok)
	}
}
