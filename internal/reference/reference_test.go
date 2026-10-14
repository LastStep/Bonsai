package reference

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/engine"
	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/guard"
)

// root is the repository's root, from this package's folder.
var root = filepath.Join("..", "..")

// The committed page is exactly what the code builds now. Read with CRLF made LF, so a Windows checkout that made the
// file CRLF still passes (.gitattributes keeps it LF there too); any other difference fails, naming the fix.
func TestPageIsCurrent(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(Path)))
	if err != nil {
		t.Fatalf("%s cannot be read: %v. Run `go generate ./...` from the repository's root to write it.", Path, err)
	}
	have := bytes.ReplaceAll(raw, []byte("\r\n"), []byte("\n"))
	want := Page()
	if bytes.Equal(have, want) {
		return
	}
	hl, wl := strings.Split(string(have), "\n"), strings.Split(string(want), "\n")
	for i := 0; i < len(hl) || i < len(wl); i++ {
		var h, w string
		if i < len(hl) {
			h = hl[i]
		}
		if i < len(wl) {
			w = wl[i]
		}
		if h != w {
			t.Fatalf("%s is not what the code builds: first difference at line %d\n  committed: %q\n  built:     %q\n"+
				"A list changed in its home without the page: run `go generate ./...` from the repository's root and commit the page.", Path, i+1, h, w)
		}
	}
}

// The page is byte-stable and plain: built twice the same, LF only, ASCII, no trailing space, one final newline.
func TestPageIsPlainAndStable(t *testing.T) {
	a, b := Page(), Page()
	if !bytes.Equal(a, b) {
		t.Fatalf("two builds of the page differ")
	}
	if bytes.Contains(a, []byte("\r")) {
		t.Errorf("the page holds a CR")
	}
	for i, l := range strings.Split(strings.TrimSuffix(string(a), "\n"), "\n") {
		if strings.TrimRight(l, " \t") != l {
			t.Errorf("line %d ends in a space", i+1)
		}
		for _, r := range l {
			if r > 126 || (r < 32 && r != '\t') {
				t.Errorf("line %d holds %q, which is not ASCII", i+1, r)
				break
			}
		}
	}
	if !bytes.HasSuffix(a, []byte("\n")) || bytes.HasSuffix(a, []byte("\n\n")) {
		t.Errorf("the page must end in exactly one newline")
	}
	if strings.Contains(string(a), "/home/") || strings.Contains(string(a), "C:\\") {
		t.Errorf("the page names a home-folder path")
	}
}

// Every list the plan names is on the page, and every word of the code's open tables is in its section.
func TestPageHoldsEveryList(t *testing.T) {
	page := string(Page())
	for _, name := range []string{"Exit codes", "Error words", "Check words", "Check --pack words", "Ask types", "Ask ops",
		"Active task: reasons there is none", "Generated kinds", "Task statuses", "Run outcomes", "Label value kinds",
		"Lane rules", "Log events", "Log categories", "Lock file kinds", "Pack file kinds", "Bonsai's document kinds",
		"Needs kinds", "Claude Code states", "Plugin results", "Check words later steps build", "Guard rules", "The guard's floor",
		"Stand-in names", "Delete words", "Words that run another program"} {
		if !strings.Contains(page, "\n## "+name+"\n") {
			t.Errorf("the page has no list %q", name)
		}
	}
	for _, tbl := range [][]format.Word{format.ErrorWords, format.CheckWords, format.PackCheckWords, format.AskTypes, format.LogEvents, format.LogCategories, engine.PluginResults, engine.ClaudeStates} {
		for _, w := range tbl {
			if !strings.Contains(page, "| `"+w.Word+"` |") {
				t.Errorf("the word %s of a table is not on the page", w.Word)
			}
		}
	}
	// The guard's rules table and its floor (plan-5 5.3.2 notes 2 and 9): every row on the page.
	for _, r := range guard.Rules {
		if !strings.Contains(page, "| `"+r.Name+"` |") {
			t.Errorf("the guard rule %s is not on the page", r.Name)
		}
	}
	for _, f := range guard.Floor {
		if !strings.Contains(page, "| `"+f.Glob+"` | "+f.At+" | `"+f.Class+"` |") {
			t.Errorf("the floor's %s is not on the page", f.Glob)
		}
	}
	// The delete check's two tables (plan-5 5.3.3): every row on the page.
	for _, w := range guard.DeleteWords {
		if !strings.Contains(page, "| `"+w.Word+"` | ") {
			t.Errorf("the delete word %s is not on the page", w.Word)
		}
	}
	for _, w := range guard.RunnerWords {
		if !strings.Contains(page, "| `"+w.Word+"` | "+w.Shells()+" |") {
			t.Errorf("the runner word %s is not on the page", w.Word)
		}
	}
	for _, l := range engine.CheckLaterWords() {
		if !strings.Contains(page, "| `"+l.Word+"` |") {
			t.Errorf("the later check word %s is not on the page", l.Word)
		}
	}
	for _, k := range format.GeneratedKinds {
		if !strings.Contains(page, "| `"+k.Kind+"` |") {
			t.Errorf("the generated kind %s is not on the page", k.Kind)
		}
	}
}

// Every enum of every schema is on the page (its home named) or skipped for a stated reason, and the page's own names
// and skips name enums that exist.
func TestEverySchemaEnumIsOnThePage(t *testing.T) {
	page := string(Page())
	seen := map[string]bool{}
	for _, f := range format.All {
		for _, e := range enums(f.Name, f.Schema()) {
			key := e.format + ":" + e.path
			seen[key] = true
			if _, skip := skipped[key]; skip {
				continue
			}
			home := "formats/schemas/" + f.Name + ".schema.json (`" + e.path + "`)"
			if !strings.Contains(page, home) && key != "lanes:lanes[].close" {
				t.Errorf("the enum %s is not on the page", key)
			}
		}
	}
	for key := range listNames {
		if !seen[key] {
			t.Errorf("listNames names %s, which is no enum of any schema", key)
		}
	}
	for key, why := range skipped {
		if !seen[key] {
			t.Errorf("skipped names %s, which is no enum of any schema", key)
		}
		if why == "" {
			t.Errorf("skipped %s has no reason", key)
		}
	}
}

// Every file go generate writes (Generated: the lists page and base's three generated skills) is exactly what the code
// builds now, a CRLF checkout's line ends read as LF: a changed template, schema or table without its file fails here,
// naming the fix.
func TestGeneratedAreCurrent(t *testing.T) {
	for _, s := range Stale(root) {
		t.Errorf("%s\nRun `go generate ./...` from the repository's root and commit what it writes.", s)
	}
	if len(Generated()) != 4 {
		t.Errorf("Generated lists %d files", len(Generated()))
	}
}

// Stale finds a generated file that differs from the code: a temporary copy of every generated file passes, a CRLF
// copy too, and the copy fails once a template in it is changed or a file is gone.
func TestStaleFindsAChange(t *testing.T) {
	tmp := t.TempDir()
	write := func(rel string, b []byte) {
		t.Helper()
		p := filepath.Join(tmp, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, b, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, g := range Generated() {
		write(g.Path, bytes.ReplaceAll(g.Build(), []byte("\n"), []byte("\r\n")))
	}
	if s := Stale(tmp); len(s) != 0 {
		t.Fatalf("a CRLF copy of what the code builds: %v", s)
	}
	state := string(engine.StateSkill())
	write(engine.StateSkillPath, []byte(strings.Replace(state, "updated_by: bonsai init\n", "updated_by: someone\n", 1)))
	s := Stale(tmp)
	if len(s) != 1 || !strings.HasPrefix(s[0], engine.StateSkillPath+" is not what the code builds") || !strings.Contains(s[0], "updated_by: someone") {
		t.Errorf("a changed template: %v", s)
	}
	if err := os.Remove(filepath.Join(tmp, filepath.FromSlash(GeneratedFilesSkillPath))); err != nil {
		t.Fatal(err)
	}
	if s := Stale(tmp); len(s) != 2 || !strings.HasPrefix(s[1], GeneratedFilesSkillPath+" cannot be read") {
		t.Errorf("a file gone: %v", s)
	}
}

// The generated-files skill is byte-stable and plain, a skill's frontmatter first, and names every kind with its
// writer and when it is cleaned.
func TestGeneratedFilesSkillIsPlain(t *testing.T) {
	a := GeneratedFilesSkill()
	if !bytes.Equal(a, GeneratedFilesSkill()) {
		t.Fatalf("two builds of the skill differ")
	}
	if !strings.HasPrefix(string(a), "---\nname: generated-files\ndescription: \"") {
		t.Errorf("no skill frontmatter:\n%s", a)
	}
	for i, l := range strings.Split(strings.TrimSuffix(string(a), "\n"), "\n") {
		if strings.TrimRight(l, " \t") != l {
			t.Errorf("line %d ends in a space", i+1)
		}
		for _, r := range l {
			if r > 126 || r < 32 {
				t.Errorf("line %d holds %q, which is not ASCII", i+1, r)
			}
		}
	}
	page := string(a)
	for _, k := range format.GeneratedKinds {
		for _, want := range []string{"### `" + k.Kind + "`", "Written by: " + k.Writer, "Cleaned: " + k.When, "Never cleaned: " + k.Never} {
			if !strings.Contains(page, want) {
				t.Errorf("the skill lacks %q", want)
			}
		}
	}
}

// The hook lines' two lists (plan-5 5.3.6 notes 2 and 5): the installed places, in their order, and every tool of the
// guard's matcher.
func TestPageHoldsTheHookLinesLists(t *testing.T) {
	page := string(Page())
	for _, name := range []string{"Installed places", "The guard line's tools"} {
		if !strings.Contains(page, "\n## "+name+"\n") {
			t.Errorf("the page has no list %q", name)
		}
	}
	if !strings.Contains(page, "| 1 | `"+engine.PlaceWindows+"` |") || !strings.Contains(page, "| 2 | `"+engine.PlaceLinux+"` |") {
		t.Error("the installed places are not on the page, in their order")
	}
	for _, tool := range engine.GuardTools {
		if !strings.Contains(page, "| `"+tool.Name+"` | "+tool.Why+" |") {
			t.Errorf("the guard line's tool %s is not on the page", tool.Name)
		}
	}
}
