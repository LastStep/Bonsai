package main

// The personal memory layer through the commands (plan-5 5.6.3): its five warnings' cases in TestCheckTable's walk,
// each from a scratch home (testpack.Isolate's BONSAI_HOME), never the exit code; init --yes and update --yes write the
// empty index once, never in a preview; a preview and check write nothing in the home.

import (
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/workspace"
)

// homeFile writes a file in the scratch Bonsai home.
func (c *cli) homeFile(rel, content string) {
	c.t.Helper()
	home, err := workspace.Home()
	if err != nil {
		c.t.Fatal(err)
	}
	p := filepath.Join(home, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		c.t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		c.t.Fatal(err)
	}
}

func personalNoteText(id, body string) string {
	return "---\nformat: bonsai.memory/1\nid: " + id + "\ntitle: A note\nkind: feedback\nupdated: 2026-10-18\nsource: \"the person\"\nlabels: {}\n---\n\n" + body
}

func personalIndexText(lines ...string) string {
	return "---\nformat: bonsai.memory/1\nid: null\ntitle: Personal memory\nkind: index\nupdated: 2026-10-18\nsource: null\nlabels: {}\n---\n\n" +
		strings.Join(lines, "\n") + "\n"
}

func init() {
	checkCases["personal-format"] = func(c *cli) {
		c.link()
		c.homeFile("personal/INDEX.md", "not the format\n")
	}
	checkCases["personal-budget"] = func(c *cli) {
		c.link()
		c.homeFile("personal/INDEX.md", personalIndexText("- [Long](notes/M-long.md): a long one"))
		c.homeFile("personal/notes/M-long.md", personalNoteText("M-long", strings.Repeat("words ", 800)))
	}
	checkCases["personal-index"] = func(c *cli) {
		c.link()
		c.homeFile("personal/INDEX.md", personalIndexText("- [Gone](notes/M-gone.md): a note that is not there"))
	}
	checkCases["personal-secret"] = func(c *cli) {
		c.link()
		c.homeFile("personal/INDEX.md", personalIndexText("- [Key](notes/M-key.md): a deploy"))
		c.homeFile("personal/notes/M-key.md", personalNoteText("M-key", "A deploy.\nIt used "+madeUpToken()+" once.\n"))
	}
	checkMore["personal-secret"] = func(t *testing.T, c *cli) {
		for _, args := range [][]string{{"check", "--json"}, {"check"}, {"status"}, {"status", "--json"}} {
			_, out, errOut := c.run("", args...)
			if strings.Contains(out+errOut, madeUpToken()) || strings.Contains(out+errOut, madeUpToken()[4:12]) {
				t.Errorf("bonsai %s names the secret's value:\n%s", strings.Join(args, " "), out)
			}
		}
		_, out, _ := c.run("", "check", "--json")
		if !strings.Contains(out, "notes/M-key.md holds a secret-shaped string on line 12 (github-token)") || !strings.Contains(out, `"file": null`) {
			t.Errorf("the warning does not name the file, the line and the kind, with no project file:\n%s", out)
		}
	}
	checkCases["personal-copy"] = func(c *cli) {
		c.link()
		c.homeFile("cache/personal-copy.json", `{"source": "//wsl.localhost/Distro/home/someone/.bonsai/personal", "refreshed": null, `+
			`"files": [], "failed": "2026-10-18T09:00:00Z", "why": "the source did not answer within 2 s"}`)
	}
}

// homeTree hashes every file under the home (path and bytes) and lists its folders, to show a command wrote nothing.
func homeTree(t *testing.T, sub string) string {
	t.Helper()
	home, err := workspace.Home()
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(home, sub)
	var lines []string
	err = filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
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

// TestPersonalIndexByInitAndUpdate: a preview writes no personal layer; init --yes writes the empty index and says
// so; a second init --yes and update --yes leave it as it is and say nothing of it; update --yes with nothing to
// change writes it when it is gone; update's preview and check change nothing in the home.
func TestPersonalIndexByInitAndUpdate(t *testing.T) {
	c := newCLI(t)
	home, _ := workspace.Home()
	index := filepath.Join(home, "personal", "INDEX.md")
	if _, out, _ := c.run("", c.linkArgs(c.pack.A)...); !strings.Contains(out, "needs-yes") && !strings.Contains(out, "--yes") {
		t.Fatalf("the preview:\n%s", out)
	}
	if _, err := os.Stat(filepath.Join(home, "personal")); err == nil {
		t.Fatalf("init's preview wrote the personal layer")
	}
	code, out, errOut := c.run("", append(c.linkArgs(c.pack.A), "--yes")...)
	if code != 0 || strings.Count(flat(out), "wrote the empty personal memory index") != 1 {
		t.Fatalf("init --yes: %d\n%s%s", code, out, errOut)
	}
	raw, err := os.ReadFile(index)
	if err != nil || !strings.Contains(string(raw), "kind: index\n") {
		t.Fatalf("the index: %v\n%s", err, raw)
	}
	if err := os.WriteFile(index, append(raw, []byte("- a line the person wrote\n")...), 0o644); err != nil {
		t.Fatal(err)
	}
	mine := homeTree(t, "personal")
	for _, args := range [][]string{append(c.linkArgs(c.pack.A), "--yes"), {"update", "--yes"}} {
		if code, out, _ := c.run("", args...); code != 0 || strings.Contains(flat(out), "personal memory index") {
			t.Errorf("%s: %d\n%s", strings.Join(args, " "), code, out)
		}
	}
	if homeTree(t, "personal") != mine {
		t.Errorf("a second init --yes or update --yes changed the personal layer")
	}
	all := homeTree(t, "")
	for _, args := range [][]string{{"update"}, {"update", "--json"}, {"check"}, {"check", "--json"}, {"status"}, {"status", "--json"}} {
		c.run("", args...)
		if got := homeTree(t, ""); got != all {
			t.Errorf("bonsai %s changed the home:\n%s\n---\n%s", strings.Join(args, " "), all, got)
		}
	}
	if err := os.RemoveAll(filepath.Join(home, "personal")); err != nil {
		t.Fatal(err)
	}
	if code, out, _ := c.run("", "update", "--yes"); code != 0 || !strings.Contains(flat(out), "wrote the empty personal memory index") {
		t.Errorf("update --yes with nothing to change, the index gone: %d\n%s", code, out)
	}
	if _, err := os.Stat(index); err != nil {
		t.Errorf("update --yes did not write the index: %v", err)
	}
}
