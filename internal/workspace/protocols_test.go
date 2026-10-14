package workspace

// A pack file's <protocols>/ start (step 5.5.0, note 1): the pack reader takes it as a path's first segment only, and
// the path it resolves to in a project is held to every rule again.

import (
	"strings"
	"testing"
)

// A files entry's path may start with <protocols>/ (as its first segment only); the rest is held to every rule a path
// is held to, and every other < and > is refused.
func TestReadPackProtocols(t *testing.T) {
	file := func(path string) string {
		return packHead + "files:\n  - path: \"" + path + "\"\n    from: a.md\n    kind: pack\n"
	}
	p, err := ReadPack([]byte(file("<protocols>/session-start.md")))
	if err != nil {
		t.Fatal(err)
	}
	if got := p.Files[0].Path; got != "<protocols>/session-start.md" {
		t.Errorf("path %q", got)
	}
	if _, err := ReadPack([]byte(file("<protocols>/deep/er/start.md"))); err != nil {
		t.Errorf("a file in a folder under <protocols>/: %v", err)
	}
	for _, c := range []struct{ path, want string }{
		{"<protocols>", "holds one of < >"},
		{"<protocols>/", "names the folder <protocols> alone"},
		{"work/<protocols>/a.md", "a placeholder may only start the path"},
		{"<other>/a.md", "a placeholder may only start the path"},
		{"<protocols>/<protocols>/a.md", "holds one of < > \" | ? *, which Windows cannot hold (a placeholder is only ever"},
		{"<protocols>/../a.md", "after <protocols>/: the path \"../a.md\" has a . or .. segment"},
		{"<protocols>/.claude/x.md", "after <protocols>/: the path \".claude/x.md\" is inside a .claude folder"},
		{"<protocols>/.git/x", "is inside a .git folder"},
		{"<protocols>//a.md", "after <protocols>/: the path \"/a.md\" starts with /"},
		{"<PROTOCOLS>/a.md", "a placeholder may only start the path"},
		{"<protocols>/CLAUDE~1/a.md", "short (8.3) name"},
	} {
		_, err := ReadPack([]byte(file(c.path)))
		if err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: %v, want %q", c.path, err, c.want)
		}
	}
	// Two entries that resolve alike are the engine's to refuse (it knows the folder); as written they differ.
	two := packHead + "files:\n  - path: \"<protocols>/a.md\"\n    from: a.md\n    kind: pack\n  - path: work/protocols/a.md\n    from: a.md\n    kind: pack\n"
	if _, err := ReadPack([]byte(two)); err != nil {
		t.Errorf("two entries as written: %v", err)
	}
	// The table, its one home: one placeholder, documented.
	if len(PathPlaceholders) != 1 || PathPlaceholders[0].Word != ProtocolsPlaceholder || PathPlaceholders[0].Means == "" {
		t.Errorf("PathPlaceholders: %+v", PathPlaceholders)
	}
}

// The path a <protocols>/ file is written at: under the project's folder, held to every rule again, and refused with
// no folder at all.
func TestResolveFilePath(t *testing.T) {
	for _, c := range []struct{ path, folder, want string }{
		{"<protocols>/s.md", "work/protocols", "work/protocols/s.md"},
		{"<protocols>/s.md", "studio/protocols/", "studio/protocols/s.md"},
		{"<protocols>/a/s.md", "p", "p/a/s.md"},
		{"docs/s.md", "", "docs/s.md"}, // no placeholder: as written, whatever the folder
	} {
		got, err := ResolveFilePath(c.path, c.folder)
		if err != nil || got != c.want {
			t.Errorf("%s under %q: %q, %v; want %q", c.path, c.folder, got, err, c.want)
		}
	}
	for _, c := range []struct{ folder, want string }{
		{"", "names no documents.protocols folder"},
		{".claude", `puts the pack's "<protocols>/s.md" at ".claude/s.md", which is inside a .claude folder`},
		{".git", "is inside a .git folder"},
		{".bonsai", "is Bonsai's own file"},
		{"..", "has a . or .. segment"},
		{"CLAUDE~1", "short (8.3) name"},
		{"a/../b", "has a . or .. segment"},
		{"/abs", "starts with /"},
		{"docs.", "ending in a dot"},
	} {
		_, err := ResolveFilePath("<protocols>/s.md", c.folder)
		if err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("folder %q: %v, want %q", c.folder, err, c.want)
		}
	}
}
