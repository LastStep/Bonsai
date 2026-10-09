package engine

// check's findings on Claude Code's settings in the project (spec §5, §6, §7): the shared .claude/settings.json and
// the .claude/settings.local.json files Claude Code reads for this checkout (its own and, in a worktree, the main
// checkout's, which Claude Code reads in every worktree: plugins.go).
//
//   - settings-rule: each permission rule (permissions.allow, deny and ask) valid on its own: a string, a tool's name
//     (letters, digits, _ and -, as mcp__server__tool) and, when it has one, a specifier in one pair of brackets at its
//     end, not empty, no space around it. A rule Claude Code cannot read is ignored, so a deny rule written wrong
//     guards nothing;
//   - hooks-off: disableAllHooks true, which turns off every hook, Bonsai's guard among them (spec §7). Bonsai writes
//     disableAllHooks: false into .claude/settings.json itself;
//   - plugin-version: a version in a plugin entry of one of the workspace's marketplaces (bonsai-<name>-<8 hex>):
//     a pack's plugin is pinned by its commit, which Claude Code takes as its version only when plugin.json and the
//     entry carry none (spec §5). A pack whose own plugin.json carries one is refused when it is linked (fetch.go), so
//     a locked pack never has one and the lock alone answers.
//
// A settings file Bonsai cannot read gives no finding here: .claude/settings.json's is the changed finding when the
// lock lists Bonsai's lines in it, and a local file's is the plugin-unchecked warning (plugins.go).

import (
	"bytes"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/LastStep/Bonsai/internal/schema"
)

var (
	ruleName = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_-]*$`)
	ruleFull = regexp.MustCompile(`^([A-Za-z][A-Za-z0-9_-]*)\((.*)\)$`)
)

// ruleProblem says why a permission rule is not valid on its own, "" when it is.
func ruleProblem(v any) string {
	s, ok := v.(string)
	switch {
	case !ok:
		return "it is " + schema.Show(v) + ", not text"
	case s == "":
		return "it is empty"
	case strings.TrimSpace(s) != s:
		return "it has spaces around it"
	case ruleName.MatchString(s):
		return ""
	}
	m := ruleFull.FindStringSubmatch(s)
	switch {
	case m == nil:
		return "it is not Tool or Tool(specifier)"
	case strings.TrimSpace(m[2]) == "":
		return "its specifier in brackets is empty"
	case strings.TrimSpace(m[2]) != m[2]:
		return "its specifier has spaces around it"
	}
	return ""
}

// settingsFile is one settings file check reads: where it is, how a finding names it, and whether Bonsai writes
// lines into it.
type settingsFile struct {
	dir, rel, shown string
	bonsais         bool // .claude/settings.json: Bonsai's own lines are in it
}

// checkSettingsFiles checks the settings files Claude Code reads for this checkout.
func (r *CheckResult) checkSettingsFiles() {
	files := []settingsFile{{r.Root, SettingsFile, SettingsFile, true}, {r.Root, LocalSettingsFile, LocalSettingsFile, false}}
	if r.Main != "" && !samePath(r.Main, r.Root) {
		files = append(files, settingsFile{r.Main, LocalSettingsFile,
			filepath.ToSlash(filepath.Join(r.Main, filepath.FromSlash(LocalSettingsFile))) + " (the main checkout's, which Claude Code reads in every worktree)", false})
	}
	locked := r.Lock != nil
	if locked {
		_, locked = r.Lock.Files[SettingsFile]
	}
	market := bonsaiMarket(r.Config.Name)
	for _, f := range files {
		raw, exists, err := readFile(f.dir, f.rel)
		if err != nil || !exists {
			continue
		}
		v, err := schema.Decode(bytes.TrimPrefix(raw, []byte("\xef\xbb\xbf")))
		o, ok := v.(schema.Object)
		if err != nil || !ok {
			continue
		}
		adopt := run("bonsai update --yes --adopt " + SettingsFile)
		if perms, ok := o.Get("permissions"); ok {
			if po, ok := perms.(schema.Object); ok {
				for _, list := range []string{"allow", "deny", "ask"} {
					rules, _ := po.Get(list)
					items, isList := rules.([]any)
					if rules != nil && !isList {
						r.add("settings-rule", f.rel, "", f.shown+"'s permissions."+list+" is "+schema.Show(rules)+", not a list of rules",
							"a person makes permissions."+list+" in "+f.shown+" a list of rules, such as [\"Read(./secrets/**)\"]")
						continue
					}
					for _, it := range items {
						if why := ruleProblem(it); why != "" {
							r.add("settings-rule", f.rel, "", f.shown+"'s permissions."+list+" rule "+ascii(schema.Show(it))+" is not valid on its own: "+
								why+", so Claude Code ignores it", "a person fixes or takes out the rule in "+f.shown+
								": Tool or Tool(specifier), such as Read(./secrets/**) or Bash(npm run test:*)")
						}
					}
				}
			}
		}
		if off, _ := o.Get("disableAllHooks"); off == true {
			next := "a person takes disableAllHooks out of " + f.shown + ", or sets it to false"
			if f.bonsais && locked {
				next = "Bonsai writes disableAllHooks: false; to write Bonsai's lines back (your copy is saved in the Bonsai home), " + adopt
			}
			r.add("hooks-off", f.rel, "", f.shown+" sets disableAllHooks: true, which turns off every hook, Bonsai's guard among them "+
				"(spec section 7)", next)
		}
		mk, _ := o.Get("extraKnownMarketplaces")
		markets, _ := mk.(schema.Object)
		for _, m := range markets {
			if !market.MatchString(m.Key) {
				continue
			}
			mo, _ := m.Value.(schema.Object)
			src, _ := mo.Get("source")
			so, _ := src.(schema.Object)
			pl, _ := so.Get("plugins")
			list, _ := pl.([]any)
			for _, p := range list {
				po, _ := p.(schema.Object)
				if ver, has := po.Get("version"); has {
					next := "a person takes version out of that plugin entry in " + f.shown
					if f.bonsais && locked {
						next = "Bonsai writes the entry with no version; to write Bonsai's lines back (your copy is saved in the Bonsai home), " + adopt
					}
					r.add("plugin-version", f.rel, "", f.shown+"'s marketplace "+m.Key+" gives the plugin "+ascii(po.String("name"))+" the version "+
						ascii(schema.Show(ver))+", so Claude Code would take it by that version, not by its locked commit (spec section 5)", next)
				}
			}
		}
	}
}

// bonsaiMarket matches the names of a workspace's marketplaces, bonsai-<name>-<8 hex> (settings.go).
func bonsaiMarket(name string) *regexp.Regexp {
	return regexp.MustCompile(`^bonsai-` + regexp.QuoteMeta(name) + `-[0-9a-f]{8}$`)
}
