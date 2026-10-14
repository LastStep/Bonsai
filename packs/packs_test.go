package packs

// The tests of base, Bonsai's own pack in packs/base/ (design/plan-5.md, note 5.5.2 item 5), reading the folder as
// committed: base stays free of code, so linking it never adds code that runs (5.4's promise, kept while base grows);
// its labels stay contract section 5.6's four; and its walls repeat none of the engine's own deny rules, which the
// engine writes into every linked project whatever its packs (internal/engine/settings.go, ownDeny). bonsai check
// --pack and claude plugin validate check the rest, in Bonsai's packs CI job.
//
// What "free of code" holds: pack.yaml's hooks and files are written as [] (a hook line runs code; a file the engine
// writes may be one a hook runs); the folder holds only the plugin manifest, the README, bonsai/'s three files and
// one SKILL.md per skill folder, so no hooks/ (a mod's register.js among them), .mcp.json, .lsp.json, monitors,
// commands or agents, and no script beside a skill that it could run; plugin.json names no component (hooks, servers,
// monitors) and no version. And every wall is a Read or Edit rule: a Bash rule would take the PowerShell tool away on
// Windows (spec section 7).

import (
	"encoding/json"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// baseDir is base's folder, from this package's folder (go test runs in it).
const baseDir = "base"

// contractFile is the contract, whose section 5.6 defines Bonsai's own labels.
const contractFile = "../design/contract.md"

// readBase reads one of base's files, forward slashes from its folder.
func readBase(t *testing.T, rel string) []byte {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(baseDir, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatalf("base's %s cannot be read: %v", rel, err)
	}
	return raw
}

// basePack reads base's bonsai/pack.yaml as the engine reads it.
func basePack(t *testing.T) *workspace.Pack {
	t.Helper()
	p, err := workspace.ReadPack(readBase(t, workspace.PackFile))
	if err != nil {
		t.Fatalf("base's %s does not read as the engine reads it: %v", workspace.PackFile, err)
	}
	return p
}

// TestBaseWritesNoCode: pack.yaml's hooks and files are written, each as [].
func TestBaseWritesNoCode(t *testing.T) {
	p := basePack(t)
	if p.ID != "base" {
		t.Errorf("pack.yaml's id is %q, not base", p.ID)
	}
	for _, key := range []string{"hooks", "files"} {
		v, ok := p.Doc.Get(key)
		list, isList := v.([]any)
		switch {
		case !ok:
			t.Errorf("pack.yaml has no %s: base writes %s: [] (its writer writes every field)", key, key)
		case !isList:
			t.Errorf("pack.yaml's %s is not a list: base writes %s: []", key, key)
		case len(list) > 0:
			t.Errorf("pack.yaml's %s is not [] (it holds %d): base carries no code, so linking it never asks for --allow-exec",
				key, len(list))
		}
	}
}

// allowedFile reports whether a file of base (its path from base's folder, forward slashes) is one base may hold.
func allowedFile(rel string) bool {
	switch rel {
	case ".claude-plugin/plugin.json", "README.md", workspace.PackFile, "bonsai/labels.yaml", "bonsai/block.md":
		return true
	}
	return skillFile().MatchString(rel)
}

// skillFile matches a skill's one file: skills/<name>/SKILL.md, the name a skill's (a lower-case letter, then
// lower-case letters, digits and dashes).
func skillFile() *regexp.Regexp { return regexp.MustCompile(`^skills/[a-z][a-z0-9-]*/SKILL\.md$`) }

// TestBaseHoldsOnlyItsFiles: no file that runs code, or that a skill could run, anywhere in base; and every skill
// folder holds its SKILL.md and nothing else.
func TestBaseHoldsOnlyItsFiles(t *testing.T) {
	var files []string
	err := filepath.WalkDir(baseDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(baseDir, p)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if !d.IsDir() {
			files = append(files, rel)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("base's folder cannot be walked: %v", err)
	}
	sort.Strings(files)
	skills := map[string]bool{}
	for _, rel := range files {
		if strings.HasPrefix(rel, "skills/") {
			skills[strings.SplitN(rel, "/", 3)[1]] = true
		}
		if allowedFile(rel) {
			continue
		}
		why := "a file base does not hold: base carries no code (its plugin manifest, README.md, bonsai/'s pack.yaml, " +
			"labels.yaml and block.md, and one SKILL.md per skill folder, nothing else)"
		switch first := strings.SplitN(rel, "/", 2)[0]; {
		case first == "hooks":
			why = "a plugin hook (or a mod's register.js), code Claude Code runs on its own"
		case first == ".mcp.json" || first == ".lsp.json":
			why = "an MCP or LSP server, code Claude Code starts on its own"
		case strings.Contains(strings.ToLower(rel), "monitor"):
			why = "a monitor, code Claude Code runs on its own"
		case first == "skills":
			why = "a file beside a skill's SKILL.md, which the skill could run: each skill folder holds its SKILL.md alone"
		case first == "bonsai" && strings.HasPrefix(rel, "bonsai/files/"):
			why = "a file the engine would write into a project; base's files: is []"
		}
		t.Errorf("packs/base/%s: %s", rel, why)
	}
	if len(skills) == 0 {
		t.Errorf("base holds no skill: its templates are skills (skills/<name>/SKILL.md)")
	}
	for name := range skills {
		if _, err := os.Stat(filepath.Join(baseDir, "skills", name, "SKILL.md")); err != nil {
			t.Errorf("packs/base/skills/%s/ holds no SKILL.md: each skill folder holds its SKILL.md alone", name)
		}
	}
}

// pluginKeys are the keys base's plugin.json may hold, each documented in base's README.md ("plugin.json"): none names
// a component (hooks, MCP or LSP servers, monitors, commands, agents) or a version.
var pluginKeys = []string{"name", "description", "author", "repository"}

// TestBasePluginNamesNoCode: plugin.json's name is base's id, and it holds no key but those its README documents:
// no component that runs, no version.
func TestBasePluginNamesNoCode(t *testing.T) {
	var m map[string]any
	if err := json.Unmarshal(readBase(t, ".claude-plugin/plugin.json"), &m); err != nil {
		t.Fatalf("base's plugin.json is not a JSON object: %v", err)
	}
	if m["name"] != "base" {
		t.Errorf("plugin.json's name is %v, not base (pack.yaml's id)", m["name"])
	}
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if !slices.Contains(pluginKeys, k) {
			t.Errorf("plugin.json holds %q, which base never holds (only %s): a component may run code, and a version would "+
				"keep every machine on its cached copy", k, strings.Join(pluginKeys, ", "))
		}
	}
}

// contractLabels reads the label definitions in contract section 5.6, its first yaml block.
func contractLabels(t *testing.T) *format.Labels {
	t.Helper()
	raw, err := os.ReadFile(filepath.FromSlash(contractFile))
	if err != nil {
		t.Fatalf("the contract cannot be read: %v", err)
	}
	text := strings.ReplaceAll(string(raw), "\r\n", "\n")
	const heading = "\n### 5.6 Bonsai's own labels\n"
	i := strings.Index(text, heading)
	if i < 0 {
		t.Fatalf("%s has no section %q", contractFile, strings.TrimSpace(heading))
	}
	section := text[i+len(heading):]
	if j := strings.Index(section, "\n## "); j >= 0 {
		section = section[:j]
	}
	start := strings.Index(section, "```yaml\n")
	if start < 0 {
		t.Fatalf("contract section 5.6 holds no yaml block")
	}
	block := section[start+len("```yaml\n"):]
	end := strings.Index(block, "\n```")
	if end < 0 {
		t.Fatalf("contract section 5.6's yaml block does not end")
	}
	l, err := format.ReadLabels([]byte(block[:end+1]))
	if err != nil {
		t.Fatalf("contract section 5.6's yaml block does not read as bonsai.labels/1: %v", err)
	}
	return l
}

// TestBaseLabelsAreTheContracts: bonsai/labels.yaml holds contract section 5.6's four labels, field by field, in its
// order. The contract leaves out the fields that do not apply to a label (values for a list, pattern for a list): a
// reader takes a missing one as null ([] for values, false for grants), and base writes each, so the two compare as read.
func TestBaseLabelsAreTheContracts(t *testing.T) {
	want := contractLabels(t)
	got, err := format.ReadLabels(readBase(t, "bonsai/labels.yaml"))
	if err != nil {
		t.Fatalf("base's labels.yaml does not read: %v", err)
	}
	if got.Namespace != want.Namespace || got.Version != want.Version {
		t.Errorf("labels.yaml's namespace and version are %s %d, the contract's %s %d", got.Namespace, got.Version, want.Namespace, want.Version)
	}
	if len(want.Labels) != 4 {
		t.Fatalf("contract section 5.6 defines %d labels, not Bonsai's four: read it again", len(want.Labels))
	}
	if len(got.Labels) != len(want.Labels) {
		t.Fatalf("labels.yaml defines %d labels, the contract %d", len(got.Labels), len(want.Labels))
	}
	text := func(p *string) string {
		if p == nil {
			return "null"
		}
		return *p
	}
	number := func(p *int64) string {
		if p == nil {
			return "null"
		}
		b, _ := json.Marshal(*p)
		return string(b)
	}
	for i, w := range want.Labels {
		g := got.Labels[i]
		fields := []struct{ name, got, want string }{
			{"name", g.Name, w.Name},
			{"kind", g.Kind, w.Kind},
			{"values", strings.Join(g.Values, ","), strings.Join(w.Values, ",")},
			{"items", text(g.Items), text(w.Items)},
			{"pattern", text(g.Pattern), text(w.Pattern)},
			{"max", number(g.Max), number(w.Max)},
			{"kinds", strings.Join(g.Kinds, ","), strings.Join(w.Kinds, ",")},
			{"set_by", g.SetBy, w.SetBy},
			{"grants", map[bool]string{true: "true", false: "false"}[g.Grants], map[bool]string{true: "true", false: "false"}[w.Grants]},
			{"description", g.Description, w.Description},
		}
		for _, f := range fields {
			if f.got != f.want {
				t.Errorf("labels.yaml's label %d (%s): %s is %q, the contract's %q", i+1, w.Name, f.name, f.got, f.want)
			}
		}
		if len(g.Extra) > 0 {
			t.Errorf("labels.yaml's label %d (%s) holds fields the contract does not define: %v", i+1, w.Name, g.Extra)
		}
	}
}

// ruleForm is a permission rule: its tool and its path.
var ruleForm = regexp.MustCompile(`^([A-Za-z]+)\((.+)\)$`)

// homeTwins are the places a wall's twins start (pack.yaml's deny docs): the Windows home from WSL, and WSL's home
// from Windows by its two network paths. A twin's original starts ~/.
var homeTwins = []string{"//mnt/c/Users/*/", "//wsl.localhost/*/home/*/", "//wsl$/*/home/*/"}

// TestBaseWalls: every wall is a Read or Edit rule, each once; none is the engine's (a rule over Bonsai's own places,
// .bonsai/, which the engine writes into every linked project whatever its packs: internal/engine/settings.go's
// ownDeny), but the Read walls on Bonsai's salt; and every twin has its original in the session's own home, ~/.
func TestBaseWalls(t *testing.T) {
	p := basePack(t)
	if len(p.Deny) == 0 {
		t.Fatalf("pack.yaml's deny is empty: base's walls are its deny rules")
	}
	seen := map[string]bool{}
	for _, d := range p.Deny {
		seen[d.Rule] = true
	}
	twice := map[string]bool{}
	for _, d := range p.Deny {
		m := ruleForm.FindStringSubmatch(d.Rule)
		if m == nil || (m[1] != "Read" && m[1] != "Edit") {
			t.Errorf("pack.yaml line %d: %s is not a Read or Edit rule: base walls files with those two only (a Bash rule takes "+
				"the PowerShell tool away on Windows, spec section 7)", d.Line, d.Rule)
			continue
		}
		if twice[d.Rule] {
			t.Errorf("pack.yaml line %d: %s is written twice", d.Line, d.Rule)
		}
		twice[d.Rule] = true
		if strings.TrimSpace(d.Why) == "" {
			t.Errorf("pack.yaml line %d: %s has no why", d.Line, d.Rule)
		}
		tool, place := m[1], m[2]
		if bonsaiPlace(place) && !(tool == "Read" && path.Base(place) == "salt") {
			t.Errorf("pack.yaml line %d: %s is a rule over Bonsai's own places, which the engine writes into every linked "+
				"project (ownDeny in internal/engine/settings.go): base never repeats one, and keeps only the Read walls on "+
				"Bonsai's salt", d.Line, d.Rule)
		}
		for _, h := range homeTwins {
			if rest, ok := strings.CutPrefix(place, h); ok {
				if original := tool + "(~/" + rest + ")"; !seen[original] {
					t.Errorf("pack.yaml line %d: %s is a twin with no original: %s is not among the walls", d.Line, d.Rule, original)
				}
			}
		}
	}
}

// bonsaiPlace reports whether a rule's path names a place in Bonsai's own folder, .bonsai, as a segment.
func bonsaiPlace(place string) bool {
	for _, seg := range strings.Split(place, "/") {
		if seg == ".bonsai" || (strings.HasPrefix(seg, ".bonsai") && strings.ContainsAny(seg, "*?[")) {
			return true
		}
	}
	return false
}

// maxBlockLines is base's block.md's own cap (design/plan-5.md, note 5.5.2 item 1): the lines the engine writes into
// every linked project's CLAUDE.md, which loads in every session.
const maxBlockLines = 4

// TestBaseBlockIsShort: block.md's text, without its leading HTML comment (which the engine leaves out), is at most
// four lines, and each skill it names is one of base's.
func TestBaseBlockIsShort(t *testing.T) {
	s := strings.ReplaceAll(string(readBase(t, "bonsai/block.md")), "\r\n", "\n")
	t2 := strings.TrimLeft(s, " \n\t")
	if !strings.HasPrefix(t2, "<!--") {
		t.Fatalf("block.md does not start with the HTML comment that documents it")
	}
	end := strings.Index(t2, "-->")
	if end < 0 {
		t.Fatalf("block.md's leading comment does not end")
	}
	body := strings.Trim(t2[end+3:], "\n")
	lines := strings.Split(body, "\n")
	if body == "" || len(lines) > maxBlockLines {
		t.Errorf("block.md's text is %d lines; base's part of every project's block is 1 to %d", len(lines), maxBlockLines)
	}
	for _, m := range regexp.MustCompile(`\bbase:([a-z][a-z0-9-]*)`).FindAllStringSubmatch(body, -1) {
		if _, err := os.Stat(filepath.Join(baseDir, "skills", m[1], "SKILL.md")); err != nil {
			t.Errorf("block.md names base:%s, which is not one of base's skills (skills/%s/SKILL.md)", m[1], m[1])
		}
	}
}
