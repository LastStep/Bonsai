package engine

// Consent to code (spec §6: "Code is consented to separately: a change to a hook line or to a file a hook runs is
// listed under 'runs code' and needs --allow-exec as well as --yes"; plan-5, piece 5.1.1, rules 1-8). init and update
// share these rules.
//
// What runs code (rule 1), each an item of Plan.RunsCode:
//   - hook: a hook line added or changed in .claude/settings.json, Bonsai's own kinds of line (spec §7) or a pack's.
//     A removed hook line runs nothing: it is an ordinary settings line in the preview. A line the lock last
//     consented to that only comes back (a person deleted it) is no new code.
//   - file: a pack file that a pack's hook line runs (its hooks entry's runs, rule 4), new or changed: written by the
//     plan, or a conflict that --adopt would write. Rule 5: the hook line may stay the same. The file may be another
//     linked pack's: it counts as the code of the hook that runs it. runs is checked, not trusted (checkRuns): every
//     item is a file a linked pack writes, and a hook command that names a file a linked pack writes, by its path or
//     its file name, must list it, else the plan is refused (exit 2).
//   - plugin: a plugin's own code parts (rule 3): what Claude Code runs on its own, without an agent's call.
//
// The baseline (rules 1 and 3): what the lock says was consented, only once verified. A locked pack is read at its
// locked commit and bonsai.yaml's folder, and is the baseline only when that content hashes to the lock's sha256 for
// the pack; the consented hook lines are the ones Bonsai would have written from those packs, only when they hash to
// the lock's record of the settings file (plan.go). A pack with no verified baseline (a folder changed in bonsai.yaml,
// a lock edited by hand) is listed as unverified, and its hook lines and plugin count as at a first link.
//
// The first link (rule 6): Bonsai's own hook lines are the link's purpose, and the preview names each with its
// sentence, so they are written on --yes (or y at a terminal), and are no item. A pack's hook lines, the files they
// run and a plugin's code parts are a pack's code, not Bonsai's: they need --allow-exec at a first link too.
//
// A relink (rule 7), init with bonsai.yaml but no lock, is judged against what is on disk, as a first link is: a
// hook line already there and equal is no change (lineChanges skips it); one that replaces another of Bonsai's
// lines on disk is "changed" and needs --allow-exec, Bonsai's own included; a pack file a hook runs is code when the
// plan writes it (equal on disk, it is adopted and not written). A plugin's code parts have no locked commit to
// compare with, so they count as at a first link.
//
// All or nothing (rule 2): a plan with any item and no --allow-exec writes nothing (Apply refuses it; cmd/bonsai
// prints the refusal, exit 4, naming --allow-exec), at a terminal too: code is never a y/N question.
//
// A plugin's code parts (rule 3), from Claude Code's plugin reference as read on 9 Oct 2026 against Claude Code
// 2.1.295 (code.claude.com/docs/en/plugins-reference, and the pages it links for mods, monitors and settings):
//   - hooks/ (hooks/hooks.json, and a mod's hooks module, hooks/register.js, which hooks.json names under modules);
//   - .mcp.json (MCP servers) and .lsp.json (LSP servers);
//   - monitors/ (monitors/monitors.json: shell commands Claude Code runs in the background);
//   - settings.json at the plugin's root (its subagentStatusLine runs a command);
//   - in .claude-plugin/plugin.json, the keys that declare such a part inline or at another path: hooks,
//     mcpServers, lspServers, monitors, experimental.monitors, channels (bound to MCP servers), settings, and
//     dependencies (other plugins it turns on); a plugin.json Bonsai cannot read counts as code;
//   - a symbolic link or a submodule anywhere in the plugin, which can stand for any of these.
//
// Paths are matched letter case aside, as Windows finds them. Not counted, as prompts or as code only an agent's call
// runs: agents/, skills/, commands/, output-styles/, themes/, workflows/ and bin/ (files an agent runs by name,
// after the user's own PATH, so they shadow no command).
//
// "The files they name included": a code part can run any file of its plugin (a hook runs a script that runs
// another; a mod imports modules; an MCP server loads the packages package.json lists), and scanning commands for
// paths misses a file run through another (rule 4's reason). So when the new commit's plugin carries a code part,
// every file of the plugin that differs from the locked commit counts as code; a plugin with no code part at the new
// commit runs nothing on its own, and a change to it alone (roles, skills, commands) is no code. At a first link or
// a relink, each code part is an item.

import (
	"sort"
	"strings"

	"github.com/LastStep/Bonsai/internal/schema"
)

// The kinds of item under "Runs code".
const (
	CodeHook   = "hook"
	CodeFile   = "file"
	CodePlugin = "plugin"
)

// CodeItem is one thing under "Runs code": it needs --allow-exec as well as --yes.
type CodeItem struct {
	Kind   string // hook, file or plugin
	Change string // add, change or remove (remove: a file of a plugin that still carries code)
	Pack   string // the pack it comes from; bonsai for Bonsai's own hook line
	Item   string // the hook line as the preview shows it; the file's path in the project; the path in the plugin
	Was    string // a hook line changed: the line before
	Why    string // why it is code, for a person
}

// CodePart is one of a plugin's own code parts.
type CodePart struct {
	Path string // its path in the plugin, or ".claude-plugin/plugin.json: <key>" for a key of the manifest
	Why  string // what Claude Code runs from it
}

// codeFolders and codeFiles are where a plugin keeps the parts Claude Code runs on its own (lower case).
var (
	codeFolders = map[string]string{
		"hooks":    "The plugin's own hooks (or a mod's hooks module), which Claude Code runs on its events.",
		"monitors": "The plugin's monitors: shell commands Claude Code runs in the background.",
	}
	codeFiles = map[string]string{
		".mcp.json":     "The plugin's MCP servers, which Claude Code starts on its own.",
		".lsp.json":     "The plugin's LSP servers, which Claude Code starts on its own.",
		"settings.json": "The plugin's settings, whose subagentStatusLine runs a command.",
	}
	codeKeys = []string{"hooks", "mcpServers", "lspServers", "monitors", "channels", "settings", "dependencies"}
)

// manifestPath is the plugin manifest's path in the plugin.
const manifestPath = ".claude-plugin/plugin.json"

// pluginCode lists a plugin's own code parts from its files (entries by path in the plugin, blobs by object id).
func pluginCode(entries map[string]treeEntry, blobs map[string][]byte) []CodePart {
	var parts []CodePart
	for p, e := range entries {
		l := strings.ToLower(p)
		top := l
		if i := strings.IndexByte(l, '/'); i >= 0 {
			top = l[:i]
		}
		switch {
		case e.mode == "120000":
			parts = append(parts, CodePart{p, "A symbolic link in the plugin, which can stand for a part Claude Code runs."})
		case e.typ == "commit":
			parts = append(parts, CodePart{p, "A git submodule in the plugin, whose files Bonsai cannot read."})
		case codeFolders[top] != "":
			parts = append(parts, CodePart{p, codeFolders[top]})
		case codeFiles[l] != "":
			parts = append(parts, CodePart{p, codeFiles[l]})
		case l == manifestPath && e.typ == "blob":
			parts = append(parts, manifestCode(p, blobs[e.sha])...)
		}
	}
	sort.Slice(parts, func(i, j int) bool { return parts[i].Path < parts[j].Path })
	return parts
}

// manifestCode lists the keys of plugin.json that declare a code part; a manifest Bonsai cannot read is one.
func manifestCode(path string, raw []byte) []CodePart {
	v, err := schema.Decode(raw)
	o, isObject := v.(schema.Object)
	if err != nil || !isObject {
		return []CodePart{{path, "The plugin's manifest is not JSON Bonsai reads, so Bonsai cannot tell what it runs."}}
	}
	var parts []CodePart
	for _, k := range codeKeys {
		if x, ok := o.Get(k); ok && x != nil {
			parts = append(parts, CodePart{path + ": " + k, "The manifest declares " + k + ", a part Claude Code runs on its own."})
		}
	}
	if x, ok := o.Get("experimental"); ok {
		if eo, isObj := x.(schema.Object); isObj {
			if m, ok := eo.Get("monitors"); ok && m != nil {
				parts = append(parts, CodePart{path + ": experimental.monitors", "The manifest declares monitors: shell commands Claude Code runs in the background."})
			}
		} else if x != nil {
			parts = append(parts, CodePart{path + ": experimental", "The manifest's experimental key is not an object, so Bonsai cannot tell what it runs."})
		}
	}
	return parts
}

// codeParts names a plugin's code parts in a few words: "hooks/hooks.json, .mcp.json".
func codeParts(parts []CodePart) string {
	var names []string
	for _, c := range parts {
		names = append(names, c.Path)
	}
	return strings.Join(names, ", ")
}

// pluginItems lists what a pack's plugin brings that runs code: at a first link (old is nil), each code part; at an
// update, when the new commit carries a code part, each file of the plugin that differs from the locked commit.
func pluginItems(pd, old *PackData) []CodeItem {
	if len(pd.Code) == 0 {
		return nil
	}
	id := pd.Ref.ID
	var items []CodeItem
	if old == nil {
		for _, c := range pd.Code {
			items = append(items, CodeItem{Kind: CodePlugin, Change: "add", Pack: id, Item: c.Path, Why: c.Why})
		}
		return items
	}
	why := map[string]string{}
	for _, c := range pd.Code {
		why[c.Path] = c.Why
	}
	paths := map[string]bool{}
	for p := range pd.Tree {
		paths[p] = true
	}
	for p := range old.Tree {
		paths[p] = true
	}
	sorted := make([]string, 0, len(paths))
	for p := range paths {
		sorted = append(sorted, p)
	}
	sort.Strings(sorted)
	carries := "The plugin carries code parts (" + codeParts(pd.Code) + "), which can run any of its files."
	for _, p := range sorted {
		was, had := old.Tree[p]
		now, has := pd.Tree[p]
		change := ""
		switch {
		case !had:
			change = "add"
		case !has:
			change = "remove"
		case was != now:
			change = "change"
		default:
			continue
		}
		w := why[p]
		if w == "" {
			w = carries
		}
		items = append(items, CodeItem{Kind: CodePlugin, Change: change, Pack: id, Item: p, Why: w})
	}
	return items
}

// consent works out the plan's items under "Runs code" (and, at a first link, Bonsai's own hook lines, which --yes
// writes). It runs after --keep and --adopt are settled. settings is what the settings file's lines would become:
// the plan's own changes when it writes the file, or, for a conflict on the file, the changes --adopt would make.
func (p *Plan) consent(settings []SettingsChange, newPacks []*PackData, oldPacks map[string]*PackData) {
	p.RunsCode, p.OwnHooks = nil, nil
	for _, c := range settings {
		if !c.RunsCode {
			continue
		}
		// Bonsai's own lines are marked Own by settings.go (ownHooks, and claim for a `bonsai hook` line on disk),
		// which no pack can set: a pack's line carries its pack id as Origin, and no pack may take the id bonsai.
		if p.FirstLink && c.Own && c.Change == "add" {
			p.OwnHooks = append(p.OwnHooks, c)
			continue
		}
		pack := c.Origin
		if c.Own {
			pack = "bonsai"
		}
		why := "A hook line of the pack " + pack + ", which Claude Code runs on its events."
		switch {
		case c.Own && c.Change == "change":
			why = "Bonsai's own hook line, changed: Claude Code runs it on its events."
		case c.Own:
			why = "Bonsai's own hook line, new since the lock: Claude Code runs it on its events."
		case c.Change == "add" && p.FirstLink:
			why = "A hook line of the pack " + pack + ", which Claude Code runs on its events; a pack's hook line needs --allow-exec at a first link too."
		}
		p.RunsCode = append(p.RunsCode, CodeItem{Kind: CodeHook, Change: c.Change, Pack: pack, Item: c.Line, Was: c.Was, Why: why})
	}

	// The pack files a pack's hook line runs, new or changed.
	runBy := map[string]string{} // path in the project -> the pack whose hook runs it
	for _, pd := range newPacks {
		for _, h := range pd.Manifest.Hooks {
			for _, r := range h.Runs {
				runBy[r] = pd.Ref.ID
			}
		}
	}
	for _, f := range p.Files {
		pack, ok := runBy[f.Path]
		if !ok || f.Kind == "block" || f.Kind == "keys" {
			continue
		}
		why := "A file the pack " + pack + "'s hook line runs"
		if f.Pack != "" && f.Pack != pack {
			why += " (the pack " + f.Pack + " writes it)"
		}
		change := ""
		switch {
		case f.write != nil && f.old == nil:
			change = "add"
		case f.write != nil:
			change = "change"
		case f.Result == Conflict && f.newH != "":
			change, why = "change", why+"; a conflict: --adopt would write the pack's copy"
		default:
			continue
		}
		why += "."
		p.RunsCode = append(p.RunsCode, CodeItem{Kind: CodeFile, Change: change, Pack: pack, Item: f.Path, Why: why})
	}

	// Each plugin's own code parts.
	for _, pd := range newPacks {
		old := oldPacks[pd.Ref.ID]
		if p.FirstLink {
			old = nil
		}
		p.RunsCode = append(p.RunsCode, pluginItems(pd, old)...)
	}
}

// checkRuns holds every pack's hook entries to the files the linked packs write (targets, by path in the project):
// each runs item is one of them, and a hook command that names one, by its path or by its file name (letter case
// aside, a backslash read as a slash), lists it in its runs. Scanning a command cannot find every file it runs (one
// run through another), so runs stays the record; this catches a pack that forgets to declare what it plainly names.
func checkRuns(packs []*PackData, targets map[string]target) error {
	paths := make([]string, 0, len(targets))
	for path := range targets {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	next := func(id string) string {
		return "the pack's maker fixes the pack " + id + "'s bonsai/pack.yaml and releases it again; until then, keep " +
			"bonsai.yaml at a ref of the pack without that hook line"
	}
	for _, pd := range packs {
		id := pd.Ref.ID
		for i, h := range pd.Manifest.Hooks {
			listed := map[string]bool{}
			for _, r := range h.Runs {
				if _, ok := targets[r]; !ok {
					return errorf(ExitInput, next(id), "the pack %s: hooks item %d's runs names %s, which no linked pack writes "+
						"(runs lists pack files by their path in the project)", id, i+1, r)
				}
				listed[r] = true
			}
			for _, path := range paths {
				if !listed[path] && namesPath(h.Command, path) {
					return errorf(ExitInput, next(id), "the pack %s: hooks item %d's command (%s) names %s, a file the pack %s "+
						"writes, but its runs does not list it, so a change to that file would run unseen", id, i+1, h.Command, path,
						targets[path].pack)
				}
			}
		}
	}
	return nil
}

// namesPath reports whether a shell command names a project path, in full or by its file name, as a word of its own:
// letter case aside (Windows), a backslash read as a slash.
func namesPath(command, path string) bool {
	cmd := strings.ToLower(strings.ReplaceAll(command, `\`, "/"))
	path = strings.ToLower(path)
	words := []string{path}
	if i := strings.LastIndexByte(path, '/'); i >= 0 {
		words = append(words, path[i+1:])
	}
	inPath := func(c byte) bool {
		return c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '_' || c == '.' || c == '-'
	}
	for _, w := range words {
		for from := 0; ; {
			i := strings.Index(cmd[from:], w)
			if i < 0 {
				break
			}
			i += from
			end := i + len(w)
			if (i == 0 || !inPath(cmd[i-1])) && (end == len(cmd) || (!inPath(cmd[end]) && cmd[end] != '/')) {
				return true
			}
			from = i + 1
		}
	}
	return false
}

// NeedsExec reports whether the plan writes anything that runs code, which needs --allow-exec as well as --yes.
func (p *Plan) NeedsExec() bool { return len(p.RunsCode) > 0 }
