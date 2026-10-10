package main

// check's two rows that wait for the recorder (step 5.2.4), their cases in TestCheckTable's walk: the memory notes'
// secret scan (secret) and Bonsai's own hook lines out of date (own-hooks).

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/engine"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/workspace"
)

func init() {
	// A memory note holding a made-up token: a finding naming the note, the line and the kind, never the value.
	checkCases["secret"] = func(c *cli) {
		c.link()
		c.write("work/memory/M-key.md", "---\nformat: bonsai.memory/1\nid: M-key\ntitle: A note with a key in it\nkind: project\n"+
			"updated: 2026-10-10\nsource: \"a test\"\nlabels: {}\n---\nThe deploy step.\nIt used "+madeUpToken()+" once.\n")
	}
	checkMore["secret"] = func(t *testing.T, c *cli) {
		_, out, _ := c.run("", "check", "--json")
		if strings.Contains(out, madeUpToken()) || strings.Contains(out, madeUpToken()[4:12]) {
			t.Errorf("check names the secret's value:\n%s", out)
		}
		if !strings.Contains(out, `"file": "work/memory/M-key.md"`) || !strings.Contains(out, "on line 11 (github-token)") {
			t.Errorf("the finding does not name the note, the line and the kind:\n%s", out)
		}
		// A clean note is no finding; the index is read too.
		c.write("work/memory/M-key.md", "---\nformat: bonsai.memory/1\nid: M-key\ntitle: A clean note\nkind: project\n"+
			"updated: 2026-10-10\nsource: \"a test\"\nlabels: {}\n---\nThe deploy step reads its token from the vault.\n")
		if _, out, _ := c.run("", "check", "--json"); strings.Contains(out, `"code": "secret"`) {
			t.Errorf("a clean note is a finding:\n%s", out)
		}
		c.write("work/memory/INDEX.md", "---\nformat: bonsai.memory/1\nid: null\ntitle: Index\nkind: index\nupdated: 2026-10-10\n"+
			"source: null\nlabels: {}\n---\n- [Key](M-key.md): password: hunter2-made-up\n")
		if _, out, _ := c.run("", "check", "--json"); !strings.Contains(out, `"file": "work/memory/INDEX.md"`) ||
			!strings.Contains(out, "(secret-named-key)") || strings.Contains(out, "hunter2") {
			t.Errorf("the index's secret:\n%s", out)
		}
	}
	// A project linked by a build before the recorder's lines: its settings file holds the guard's line only, and the
	// lock's fingerprint is that file's, as such a build wrote it. A warning, never a finding; update --allow-exec
	// --yes clears it.
	checkCases["own-hooks"] = func(c *cli) { c.link(); c.linkedBefore() }
	checkMore["own-hooks"] = func(t *testing.T, c *cli) {
		_, out, _ := c.run("", "check", "--json")
		if !strings.Contains(out, "11 to add (SessionStart: bonsai hook start; UserPromptSubmit: bonsai hook record [async]") ||
			!strings.Contains(out, `"do": "run: bonsai update --allow-exec --yes"`) || strings.Contains(out, `"code": "changed"`) {
			t.Errorf("the warning:\n%s", out)
		}
		// update --yes alone is refused (a hook line added runs code), writing nothing.
		before := c.read(".claude/settings.json")
		if code, out, _ := c.run("", "update", "--yes"); code != 4 || !strings.Contains(out, "add     hook    SessionStart: bonsai hook start  (bonsai)") {
			t.Errorf("update --yes: exit %d\n%s", code, out)
		}
		if c.read(".claude/settings.json") != before {
			t.Error("a refused update wrote the settings file")
		}
		if code, out, errOut := c.run("", "update", "--allow-exec", "--yes"); code != 0 {
			t.Fatalf("update --allow-exec --yes: exit %d\n%s%s", code, out, errOut)
		}
		if _, out, _ := c.run("", "check", "--json"); strings.Contains(out, `"code": "own-hooks"`) || strings.Contains(out, `"code": "changed"`) {
			t.Errorf("after the update:\n%s", out)
		}
	}
}

// madeUpToken is a made-up GitHub-shaped token, built at run time so no token-shaped literal sits in the repo.
func madeUpToken() string { return "ghp_" + strings.Repeat("Qz7Wk2", 6) }

// linkedBefore turns a project this build linked into one a build before step 5.2.4 linked: Bonsai's own hook lines
// in .claude/settings.json are the guard's only, and the lock's fingerprint of Bonsai's lines is the one such a build
// wrote (the SHA-256 of the lines' canon forms, sorted, each followed by a line feed: engine's linesHash). The
// fingerprint of the file as linked is computed the same way first and must equal the lock's, so this reading of
// the rule is checked against the engine's.
func (c *cli) linkedBefore() {
	c.t.Helper()
	v, err := schema.Decode([]byte(c.read(".claude/settings.json")))
	if err != nil {
		c.t.Fatal(err)
	}
	doc := v.(schema.Object)
	lockV, err := schema.Decode([]byte(c.read(workspace.LockFile)))
	if err != nil {
		c.t.Fatal(err)
	}
	lock := lockV.(schema.Object)
	files, _ := lock.Get("files")
	entry, _ := files.(schema.Object).Get(".claude/settings.json")
	if got := fingerprint(doc); got != entry.(schema.Object).String("sha256") {
		c.t.Fatalf("the fingerprint read here (%s) is not the lock's (%s): engine's linesHash changed", got, entry.(schema.Object).String("sha256"))
	}
	// Take out every hook line but the guard's that runs bonsai, and the events left empty.
	hooks, _ := doc.Get("hooks")
	var kept schema.Object
	for _, ev := range hooks.(schema.Object) {
		var groups []any
		for _, g := range ev.Value.([]any) {
			gObj := g.(schema.Object)
			hs, _ := gObj.Get("hooks")
			var keep []any
			for _, h := range hs.([]any) {
				cmd := h.(schema.Object).String("command")
				if strings.HasPrefix(cmd, "bonsai hook ") && cmd != engine.GuardCommand {
					continue
				}
				keep = append(keep, h)
			}
			if len(keep) > 0 {
				gObj[gObj.Index("hooks")].Value = keep
				groups = append(groups, gObj)
			}
		}
		if len(groups) > 0 {
			kept = append(kept, schema.Member{Key: ev.Key, Value: groups})
		}
	}
	doc[doc.Index("hooks")].Value = kept
	raw, err := schema.EncodeUTF8(doc)
	if err != nil {
		c.t.Fatal(err)
	}
	c.write(".claude/settings.json", string(raw))
	e := entry.(schema.Object)
	e[e.Index("sha256")].Value = fingerprint(doc)
	raw, err = schema.Encode(lock)
	if err != nil {
		c.t.Fatal(err)
	}
	c.write(workspace.LockFile, string(raw))
}

// fingerprint is the lock's fingerprint of Bonsai's lines in a settings file that holds Bonsai's lines only (a
// fresh link of the test pack): each line's canon form (engine's Line.canon), sorted, each once, each followed by a
// line feed, hashed with SHA-256.
func fingerprint(doc schema.Object) string {
	set := map[string]bool{}
	for _, k := range []string{"autoMemoryEnabled", "disableAllHooks"} {
		if v, ok := doc.Get(k); ok {
			set["key\t"+k+"\t"+schema.Show(v)] = true
		}
	}
	if perms, ok := doc.Get("permissions"); ok {
		deny, _ := perms.(schema.Object).Get("deny")
		for _, r := range deny.([]any) {
			set["deny\t"+r.(string)] = true
		}
	}
	if hooks, ok := doc.Get("hooks"); ok {
		for _, ev := range hooks.(schema.Object) {
			for _, g := range ev.Value.([]any) {
				gObj := g.(schema.Object)
				hs, _ := gObj.Get("hooks")
				for _, h := range hs.([]any) {
					ho := h.(schema.Object)
					timeout := ""
					if t, ok := ho.Get("timeout"); ok {
						timeout = string(t.(json.Number))
					}
					async, _ := ho.Get("async")
					set[strings.Join([]string{"hook", ev.Key, gObj.String("matcher"), ho.String("command"), timeout,
						strconv.FormatBool(async == true)}, "\t")] = true
				}
			}
		}
	}
	if m, ok := doc.Get("extraKnownMarketplaces"); ok {
		for _, mk := range m.(schema.Object) {
			// Bonsai's order of the source object (source, name, owner, plugins), which its fingerprint is taken over.
			o := mk.Value.(schema.Object)
			src, _ := o.Get("source")
			so := src.(schema.Object)
			var ordered schema.Object
			for _, k := range []string{"source", "name", "owner", "plugins"} {
				if v, ok := so.Get(k); ok {
					ordered = append(ordered, schema.Member{Key: k, Value: v})
				}
			}
			set["marketplace\t"+mk.Key+"\t"+schema.Show(schema.Object{{Key: "source", Value: ordered}})] = true
		}
	}
	if p, ok := doc.Get("enabledPlugins"); ok {
		for _, pl := range p.(schema.Object) {
			set["plugin\t"+pl.Key+"\t"+schema.Show(pl.Value)] = true
		}
	}
	keys := make([]string, 0, len(set))
	for k := range set {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	sum := sha256.Sum256([]byte(strings.Join(keys, "\n") + "\n"))
	return hex.EncodeToString(sum[:])
}
