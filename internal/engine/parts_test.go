package engine

// Tests of the engine's parts: the fetch and the cache, the block, the settings lines, the diff.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/testpack"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// The fetch takes a tag or a commit; a commit already in the cache is read with the source gone (once per
// commit); a ref that is not there is bad input; the pack's content hash is line-ending blind and stable.
func TestFetch(t *testing.T) {
	e := setup(t)
	work := filepath.Join(e.tmp, "pack-work")
	testpack.Git(t, work, "tag", "v1.0.0", e.pack.B)
	testpack.Git(t, e.tmp, "-C", e.pack.Source, "fetch", "-q", work, "refs/tags/v1.0.0:refs/tags/v1.0.0")
	c := cache{home: e.home}
	got, err := c.fetch(e.pack.Source, "v1.0.0")
	if err != nil || got != e.pack.B {
		t.Fatalf("a tag: %s %v", got, err)
	}
	if _, err := c.fetch(e.pack.Source, e.pack.A); err != nil {
		t.Fatal(err)
	}
	pd, err := c.packAt(workspace.PackRef{ID: testpack.ID, Source: e.pack.Source}, e.pack.A)
	if err != nil {
		t.Fatal(err)
	}
	if string(pd.Files["demo/guide.md"]) != "# Guide\n\nEdition 1.\n" || pd.Block != "The demo pack is linked: its role is demo-pack:marker." ||
		len(pd.SHA256) != 64 || pd.Manifest.Hooks[0].Command != "echo demo hook A" {
		t.Errorf("pack at A: %+v", pd)
	}
	moved := e.pack.Source + "-moved"
	if err := os.Rename(e.pack.Source, moved); err != nil {
		t.Fatal(err)
	}
	if got, err := c.fetch(e.pack.Source, e.pack.A); err != nil || got != e.pack.A {
		t.Errorf("a cached commit with its source gone: %s %v", got, err)
	}
	if _, err := c.fetch(e.pack.Source, e.pack.C); err == nil {
		t.Errorf("an uncached commit with its source gone was found")
	}
	if err := os.Rename(moved, e.pack.Source); err != nil {
		t.Fatal(err)
	}
	if _, err := c.fetch(e.pack.Source, "v9"); err == nil || err.(*Error).Exit != ExitInput {
		t.Errorf("a missing tag: %v", err)
	}
	if _, err := c.fetch(e.pack.Source, strings.Repeat("ab", 20)); err == nil || err.(*Error).Exit != ExitInput {
		t.Errorf("a missing commit: %v", err)
	}
	// The content hash: one line per file, line endings made LF.
	entries := map[string]treeEntry{"b": {typ: "blob", sha: "2"}, "a": {typ: "blob", sha: "1"}, "t": {typ: "tree", sha: "3"}}
	h1 := contentHash(entries, map[string][]byte{"1": []byte("x\n"), "2": []byte("y\n")})
	h2 := contentHash(entries, map[string][]byte{"1": []byte("x\r\n"), "2": []byte("y\r\n")})
	h3 := contentHash(entries, map[string][]byte{"1": []byte("y\n"), "2": []byte("x\n")})
	if h1 != h2 || h1 == h3 {
		t.Errorf("content hash: %s %s %s", h1, h2, h3)
	}
}

func TestBlockText(t *testing.T) {
	for in, want := range map[string]string{
		"<!--\ndocs\n-->\nThe block.\n":             "The block.",
		"\ufeff<!-- docs -->\r\n\r\nA\r\nB\r\n\r\n": "A\nB",
		"No comment.\n":                             "No comment.",
	} {
		if got := blockText([]byte(in)); got != want {
			t.Errorf("blockText(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestBlockInPlace(t *testing.T) {
	root := t.TempDir()
	body := blockStart + "\nnew\n" + blockEnd + "\n"
	cases := []struct{ name, before, after string }{
		{"no file", "", body},
		{"text without a newline", "# Own", "# Own\n\n" + body},
		{"text with a blank line", "# Own\n\n", "# Own\n\n" + body},
		{"a block in the middle", "top\n" + blockStart + "\nold\n" + blockEnd + "\nbottom\n", "top\n" + body + "bottom\n"},
		{"CRLF", "# Own\r\n", "# Own\r\n\r\n" + strings.ReplaceAll(body, "\n", "\r\n")},
	}
	for _, c := range cases {
		path := filepath.Join(root, BlockFile)
		_ = os.Remove(path)
		if c.before != "" {
			if err := os.WriteFile(path, []byte(c.before), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		d, err := readBlock(root)
		if err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		if got := string(d.withBlock(body)); got != c.after {
			t.Errorf("%s:\n%q\nwant\n%q", c.name, got, c.after)
		}
		if err := os.WriteFile(path, d.withBlock(body), 0o644); err != nil {
			t.Fatal(err)
		}
		d, _ = readBlock(root)
		if !d.found || regionHash(d.region()) != regionHash(body) {
			t.Errorf("%s: the block reads back as %q", c.name, d.region())
		}
	}
	for _, broken := range []string{blockStart + "\n", blockEnd + "\n" + blockStart + "\n", blockStart + "\n" + blockStart + "\n" + blockEnd + "\n"} {
		if err := os.WriteFile(filepath.Join(root, BlockFile), []byte(broken), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := readBlock(root); err == nil || err.(*Error).Exit != ExitInput {
			t.Errorf("%q: %v", broken, err)
		}
	}
}

func TestOldBonsaiHook(t *testing.T) {
	for cmd, want := range map[string]bool{
		"/opt/bonsai-0.4/bin/bonsai hook guard":                       true,
		`"/srv/tools/go/bin/bonsai" guard`:                            true,
		`C:\Users\x\go\bin\bonsai.exe hook guard`:                     true,
		`bash "/srv/proj/station/agent/Sensors/scope.sh" "/srv/proj"`: true,
		"bonsai hook guard || exit 2":                                 false,
		"/usr/bin/npm run lint":                                       false,
		"bash ./agent/Sensors/scope.sh":                               false,
		`node "${CLAUDE_PROJECT_DIR}/tools/hooks/guard.mjs"`:          false,
	} {
		if got := isOldBonsaiHook(cmd); got != want {
			t.Errorf("isOldBonsaiHook(%q) = %v", cmd, got)
		}
	}
}

// applyLines writes Bonsai's lines and keeps every other entry, in its place; applied twice it gives the same
// document; taking every line out leaves the project's file as it was.
func TestApplyLines(t *testing.T) {
	v, err := schema.Decode([]byte(drifted))
	if err != nil {
		t.Fatal(err)
	}
	root := v.(schema.Object)
	cfg := &workspace.Config{Name: "demo", NeverEdit: []string{"ledger.json"}}
	lnew := buildLines(cfg, []packLines{{id: "p", source: "https://github.com/o/p", commit: strings.Repeat("a", 40),
		hooks: []workspace.HookEntry{{Event: "PreToolUse", Matcher: "Bash", Command: "echo p", Why: "x"}},
		deny:  []workspace.DenyEntry{{Rule: "Read(secrets/**)", Why: "y"}}}})
	claimed, _ := claim(diskLines(root), nil, true, "")
	once := applyLines(root, claimed, lnew)
	claimed2, same := claim(diskLines(once), lnew, false, linesHash(lnew))
	if !same {
		t.Fatalf("the lines written do not read back as Bonsai's")
	}
	twice := applyLines(once, claimed2, lnew)
	a, _ := schema.Encode(once)
	b, _ := schema.Encode(twice)
	if string(a) != string(b) {
		t.Errorf("applied twice:\n%s\n%s", a, b)
	}
	got := schema.Show(once)
	for _, want := range []string{`"deny":["Read(secrets/**)","Edit(ledger.json)"]`, `{"matcher":"Bash","hooks":[{"type":"command","command":"npm run lint-staged"}]}`,
		`{"matcher":"Bash","hooks":[{"type":"command","command":"echo p"}]}`, `"repo":"o/p"`} {
		if !strings.Contains(got, want) {
			t.Errorf("lacks %s:\n%s", want, got)
		}
	}
	// Taking Bonsai's lines out again leaves the project's own entries; the project's Read(secrets/**) equals a
	// pack's rule, so once written it is Bonsai's line too, and goes with them.
	back := schema.Show(applyLines(once, claimed2, nil))
	want := `{"permissions":{"allow":["Bash(npm test)"]},"hooks":{"PreToolUse":[{"matcher":"Bash","hooks":[{"type":"command","command":"npm run lint-staged"}]}]},"model":"opus"}`
	if back != want {
		t.Errorf("after taking the lines out:\n%s\nwant\n%s", back, want)
	}
}

func TestPluginSource(t *testing.T) {
	c := strings.Repeat("c", 40)
	for _, x := range []struct{ source, folder, want string }{
		{"https://github.com/LastStep/bonsai-test-pack", "", `{"source":"github","repo":"LastStep/bonsai-test-pack","sha":"` + c + `"}`},
		{"https://github.com/LastStep/bonsai-test-pack.git", "", `{"source":"github","repo":"LastStep/bonsai-test-pack","sha":"` + c + `"}`},
		{"https://github.com/LastStep/Bonsai.git", "packs/base", `{"source":"git-subdir","url":"https://github.com/LastStep/Bonsai.git","path":"packs/base","sha":"` + c + `"}`},
		{"https://git.example.com/p.git", "", `{"source":"url","url":"https://git.example.com/p.git","sha":"` + c + `"}`},
	} {
		if got := schema.Show(pluginSource(x.source, x.folder, c)); got != x.want {
			t.Errorf("%s %s: %s", x.source, x.folder, got)
		}
	}
	if u := gitURL(t.TempDir()); !strings.HasPrefix(u, "file:///") || strings.Contains(u, `\`) {
		t.Errorf("a local folder: %s", u)
	}
}

func TestDiff(t *testing.T) {
	if Diff("a", []byte("x\n"), []byte("x\n")) != "" {
		t.Error("equal files differ")
	}
	got := Diff("f.md", []byte("1\n2\n3\n4\n5\n6\n7\n8\n9\n10\n"), []byte("1\n2\n3\n4\nfive\n6\n7\n8\n9\n10\n11\n"))
	want := "--- a/f.md\n+++ b/f.md\n@@ -2,9 +2,10 @@\n 2\n 3\n 4\n-5\n+five\n 6\n 7\n 8\n 9\n 10\n+11\n"
	if got != want {
		t.Errorf("diff:\n%s\nwant\n%s", got, want)
	}
	if got := Diff("n.md", nil, []byte("caf\xc3\xa9\n")); got != "--- /dev/null\n+++ b/n.md\n@@ -0,0 +1 @@\n+caf\\u00e9\n" {
		t.Errorf("a new file: %q", got)
	}
	if got := Diff("b", []byte("a\x00"), []byte("b")); !strings.Contains(got, "binary") {
		t.Errorf("binary: %q", got)
	}
}

func TestNewIDForm(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 50; i++ {
		id, err := NewID()
		if err != nil || len(id) != 29 || !strings.HasPrefix(id, "ws-") || seen[id] {
			t.Fatalf("%q %v", id, err)
		}
		seen[id] = true
		if _, err := workspace.ReadConfig([]byte("format: bonsai.workspace/1\nid: " + id + "\nname: x\n")); err != nil {
			t.Fatal(err)
		}
	}
}
