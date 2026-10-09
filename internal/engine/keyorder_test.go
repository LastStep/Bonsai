package engine

// .claude/settings.json's key order (step 5.1.7): Bonsai changes only its own entries in the file, keeps every other
// key, value and character where the file has it, and writes a new file in the order Claude Code writes, so Claude
// Code's first project-scope install leaves the file as it is and a later update's diff holds only Bonsai's lines.

import (
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/testpack"
)

// What Claude Code 2.1.295 wrote when a project-scope install rewrote a settings file that Bonsai had written (the
// run report of 5.1.7, measurement 1; 2.1.294 wrote the same in part 4b): the top-level keys, and an inline
// marketplace's source object.
const (
	claudeTopKeys    = "permissions hooks disableAllHooks enabledPlugins extraKnownMarketplaces autoMemoryEnabled"
	claudeMarketKeys = "source name plugins owner"
)

func marketKeys(t *testing.T, doc schema.Object, market string) string {
	t.Helper()
	ekm, _ := doc.Get("extraKnownMarketplaces")
	m, _ := ekm.(schema.Object).Get(market)
	src, _ := m.(schema.Object).Get("source")
	return strings.Join(src.(schema.Object).Keys(), " ")
}

// A new file is written in Claude Code's order; an update to the same commit writes nothing; an update to a new
// commit changes only the lines naming the marketplace and the commit, in place.
func TestSettingsKeyOrder(t *testing.T) {
	e := setup(t)
	root := testpack.Project(t, e.tmp, "order")
	e.link(t, root, e.pack.A)
	marketA := MarketplaceName("demo", []string{e.pack.A})
	marketB := MarketplaceName("demo", []string{e.pack.B})
	doc := settingsDocOf(t, root)
	if got := strings.Join(doc.Keys(), " "); got != claudeTopKeys {
		t.Errorf("a new file's keys: %s, want Claude Code's %s", got, claudeTopKeys)
	}
	if got := marketKeys(t, doc, marketA); got != claudeMarketKeys {
		t.Errorf("the marketplace's keys: %s, want Claude Code's %s", got, claudeMarketKeys)
	}
	before := read(t, root, SettingsFile)

	// The same commit: nothing written.
	p := e.apply(t, root, Request{})
	if r := result(t, p, SettingsFile); r.Result != Unchanged || read(t, root, SettingsFile) != before {
		t.Errorf("update to the same commit: %s\n%s", r.Result, read(t, root, SettingsFile))
	}

	// A new commit: only the marketplace's name, the plugin's name and the pinned commit change, each in its place.
	testpack.SetRef(t, root, e.pack.A, e.pack.B)
	e.apply(t, root, Request{})
	want := strings.ReplaceAll(strings.ReplaceAll(before, marketA, marketB), e.pack.A, e.pack.B)
	if got := read(t, root, SettingsFile); got != want {
		t.Errorf("update to B changed more than Bonsai's lines:\n%s\nwant\n%s", got, want)
	}
	if n := changedLines(before, read(t, root, SettingsFile)); n != 4 {
		t.Errorf("update to B changed %d lines, want 4 (the plugin's key, the marketplace's key and name, the sha)", n)
	}
}

// changedLines counts the lines that differ between two texts of the same line count.
func changedLines(a, b string) int {
	al, bl := strings.Split(a, "\n"), strings.Split(b, "\n")
	if len(al) != len(bl) {
		return -1
	}
	n := 0
	for i := range al {
		if al[i] != bl[i] {
			n++
		}
	}
	return n
}

// A file of the project's own, in Claude Code's order (as its install leaves it, measurement 4): Bonsai's keys go
// where Claude Code writes them, and every key, value and byte of the project's stays as it was, its non-ASCII text
// and an unknown key at the end among them.
func TestSettingsKeepTheProjectsBytes(t *testing.T) {
	e := setup(t)
	root := testpack.Project(t, e.tmp, "own")
	own := `{
  "$schema": "https://json.schemastore.org/claude-code-settings.json",
  "env": {
    "ZETA": "1",
    "GREETING": "caf` + "\u00e9" + ` \u0001 \t"
  },
  "permissions": {
    "allow": [
      "Bash(ls)"
    ],
    "ask": [
      "Bash(git push)"
    ]
  },
  "model": "haiku",
  "outputStyle": "default",
  "zzzProjectKey": 1
}
`
	writeFile(t, root, SettingsFile, own)
	e.link(t, root, e.pack.A)
	doc := settingsDocOf(t, root)
	if got := strings.Join(doc.Keys(), " "); got != "$schema env permissions model hooks disableAllHooks enabledPlugins extraKnownMarketplaces outputStyle autoMemoryEnabled zzzProjectKey" {
		t.Errorf("keys: %s", got)
	}
	perms, _ := doc.Get("permissions")
	if got := strings.Join(perms.(schema.Object).Keys(), " "); got != "allow deny ask" {
		t.Errorf("permissions' keys: %s", got)
	}
	got := read(t, root, SettingsFile)
	for _, keep := range []string{"\"GREETING\": \"caf\u00e9 \\u0001 \\t\"", "  \"model\": \"haiku\",\n  \"hooks\": {", "  \"zzzProjectKey\": 1\n}\n"} {
		if !strings.Contains(got, keep) {
			t.Errorf("the file lacks %q:\n%s", keep, got)
		}
	}
}

// placeKey puts a key where Claude Code writes it: after the last key Claude Code writes before it, first when there
// is none, and appends a key it does not know.
func TestPlaceKey(t *testing.T) {
	obj := func(keys ...string) schema.Object {
		o := schema.Object{}
		for _, k := range keys {
			o = append(o, schema.Member{Key: k, Value: true})
		}
		return o
	}
	cases := []struct {
		have []string
		key  string
		want string
	}{
		{nil, "hooks", "hooks"},
		{[]string{"zzz"}, "permissions", "permissions zzz"},
		{[]string{"env", "zzz"}, "permissions", "env permissions zzz"},
		{[]string{"permissions", "model"}, "hooks", "permissions model hooks"},
		{[]string{"autoMemoryEnabled"}, "disableAllHooks", "disableAllHooks autoMemoryEnabled"},
		{[]string{"hooks", "extraKnownMarketplaces"}, "enabledPlugins", "hooks enabledPlugins extraKnownMarketplaces"},
		{[]string{"hooks"}, "notAKey", "hooks notAKey"},
		{[]string{"hooks", "permissions"}, "permissions", "hooks permissions"},
	}
	for _, c := range cases {
		if got := strings.Join(placeKey(obj(c.have...), c.key, true, claudeKeyOrder).Keys(), " "); got != c.want {
			t.Errorf("%v + %s: %s, want %s", c.have, c.key, got, c.want)
		}
	}
}

// The encoder for the settings file writes strings as JSON.stringify does.
func TestEncodeUTF8(t *testing.T) {
	b, err := schema.EncodeUTF8(schema.Object{{Key: "k\u00e9", Value: "a\u00e9\b\f\n\r\t\u0001\u007f\"\\/<>&\U0001F600"}})
	if err != nil {
		t.Fatal(err)
	}
	if want := "{\n  \"k\u00e9\": \"a\u00e9\\b\\f\\n\\r\\t\\u0001\u007f\\\"\\\\/<>&\U0001F600\"\n}\n"; string(b) != want {
		t.Errorf("%q, want %q", b, want)
	}
}
