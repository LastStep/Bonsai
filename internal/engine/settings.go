package engine

// Bonsai's lines in the project's .claude/settings.json (file kind keys, contract §14: "Bonsai's own entries in a
// project-owned JSON file"; spec §5, §6, §7). The rest of the file is the project's and is kept, value for value.
//
// A line is one entry Bonsai writes, as the preview names it (spec §6, "The preview names every settings line"):
//
//	hook         a hook command under hooks.<event>, in a group with its matcher (spec §7; a pack's hooks)
//	deny         a rule in permissions.deny (each never_edit path in bonsai.yaml as Edit(<path>); a pack's deny)
//	key          autoMemoryEnabled: false (spec §10) and disableAllHooks: false (spec §7)
//	marketplace  the inline marketplace under extraKnownMarketplaces, bonsai-<workspace name>-<8 hex>, each pack's
//	             plugin pinned to its locked commit by sha, with no version (spec §5)
//	plugin       <pack id>@<marketplace>: true under enabledPlugins (spec §5)
//
// What Bonsai last wrote. The lock records one fingerprint for the file, the SHA-256 of Bonsai's lines as written
// (linesHash). The lines themselves are not stored: the engine finds them in the file (claim) by what it can tell
// is its own, the lines it would write from the lock's packs at their locked commits and bonsai.yaml, and lines no
// person writes: autoMemoryEnabled and disableAllHooks, a marketplace named bonsai-..., a plugin enabled from one,
// and a hook line that runs `bonsai hook`. Only then is the fingerprint compared, so an edit of one of Bonsai's
// lines reads as an edit, and a line a person added beside them is never Bonsai's. When a path left never_edit in
// bonsai.yaml since the last update, its old Edit rule is found by trying the file's other Edit rules against the
// lock's fingerprint, each set of them in turn: only the set whose fingerprint matches exactly is Bonsai's, so a rule
// a person wrote is never taken for one of Bonsai's.
//
// The file's key order (step 5.1.7). Claude Code writes this file too (a project-scope plugin install, an uninstall),
// with JSON.stringify: two-space indent, a final newline, and the keys it knows in its own order, the ones it does not
// know after them in their order. As measured on Claude Code 2.1.294 and 2.1.295 (records/runs, 5.1.7), that order
// is stable: a second install writes the same bytes. So Bonsai changes only its own entries and keeps every other
// key, value and character where the file has it:
//   - an entry of Bonsai's that stays keeps its place, and its bytes when its value is unchanged (a marketplace in
//     Claude Code's key order is left as it is);
//   - a marketplace or plugin entry whose name changed (a new commit) takes the old entry's place;
//   - a key Bonsai adds goes where Claude Code would put it (claudeKeyOrder): after the last key of the file that
//     Claude Code writes before it; a deny list goes into permissions the same way (claudePermissionsOrder); a
//     marketplace is written in Claude Code's order (claudeMarketplaceOrder);
//   - strings are written as JSON.stringify writes them (schema.EncodeUTF8), so a project's non-ASCII text is kept.
// A new file is therefore written in the order Claude Code writes, and Claude Code's first install leaves it as it is.
// In a file that held settings of the project's own, Claude Code's first install may still move the project's keys
// into its own order once (keys Claude Code knows that claudeKeyOrder does not list, a file not in Claude Code's
// order); from then on the order stays, and Bonsai's writes keep it.

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// SettingsFile is the project's shared Claude Code settings file, the one settings file the engine writes.
const SettingsFile = ".claude/settings.json"

// The engine's own hook line (spec §7's table; the skeleton writes the guard line only, below).
const (
	GuardEvent   = "PreToolUse"
	GuardMatcher = "Edit|Write|MultiEdit|NotebookEdit|Bash|PowerShell"
	GuardCommand = "bonsai hook guard || exit 2"
	GuardTimeout = "10"
)

// ownHooks is this build's table of Bonsai's own hook lines. Spec §7 lists four: the guard (PreToolUse), the stop
// gate (Stop), start (SessionStart) and the recorder (eleven events, async). The walking skeleton builds the guard
// only (plan part 5), so only its line is written: a line that runs a `bonsai hook` word this build does not have
// would fail every time it runs, and the stop line (|| exit 2) would then block every session's end. Step 5.1-5.3
// add the other three as they build them.
var ownHooks = []Line{{
	Kind: "hook", Origin: "bonsai", Own: true, Event: GuardEvent, Matcher: GuardMatcher, Command: GuardCommand, Timeout: GuardTimeout,
	Why: "Checks every file edit and shell command against the task's rights; blocks if bonsai is missing.",
}}

// Line is one of Bonsai's entries in .claude/settings.json.
type Line struct {
	Kind    string // hook, deny, key, marketplace or plugin
	Origin  string // where it comes from: bonsai (this build's own), bonsai.yaml (never_edit), old (an old Bonsai hook line), or a pack id
	Own     bool   // a hook line of Bonsai's own (ownHooks, or a `bonsai hook` line claim finds): set only here, never from a pack
	Slot    string // two lines with one slot are one line changed, not one removed and one added
	Event   string // hook: the Claude Code hook event
	Matcher string // hook: the group's matcher, "" for none
	Command string // hook: the shell line
	Timeout string // hook: the timeout in seconds, as JSON number text; "" for none
	Async   bool   // hook: runs without waiting
	Rule    string // deny: the permission rule
	Name    string // key, marketplace, plugin: its key in its object
	Value   any    // key, marketplace, plugin: its JSON value
	Why     string // the sentence the preview prints
}

// canon is the line's identity: two lines are the same line when their canon is.
func (l Line) canon() string {
	switch l.Kind {
	case "hook":
		return strings.Join([]string{"hook", l.Event, l.Matcher, l.Command, l.Timeout, strconv.FormatBool(l.Async)}, "\t")
	case "deny":
		return "deny\t" + l.Rule
	}
	return l.Kind + "\t" + l.Name + "\t" + schema.Show(l.Value)
}

// match is the line's identity when claim looks for it in the file: its canon, with every JSON object's keys in
// sorted order. Claude Code rewrites the settings file in an order of its own when it writes it (a project-scope
// plugin install moves the marketplace's owner after its plugins), and that is no edit of Bonsai's lines.
func (l Line) match() string {
	switch l.Kind {
	case "hook", "deny":
		return l.canon()
	}
	return l.Kind + "\t" + l.Name + "\t" + schema.Show(sortedKeys(l.Value))
}

// sortedKeys gives a JSON value with every object's keys in sorted order.
func sortedKeys(v any) any {
	switch x := v.(type) {
	case schema.Object:
		out := make(schema.Object, len(x))
		for i, m := range x {
			out[i] = schema.Member{Key: m.Key, Value: sortedKeys(m.Value)}
		}
		sort.SliceStable(out, func(i, j int) bool { return out[i].Key < out[j].Key })
		return out
	case []any:
		out := make([]any, len(x))
		for i, e := range x {
			out[i] = sortedKeys(e)
		}
		return out
	}
	return v
}

// Text is the line as the preview shows it.
func (l Line) Text() string {
	switch l.Kind {
	case "hook":
		s := l.Event
		if l.Matcher != "" {
			s += " (" + l.Matcher + ")"
		}
		s += ": " + l.Command
		if l.Async {
			s += " [async]"
		}
		return s
	case "deny":
		return l.Rule
	case "marketplace":
		return l.Name + ": " + marketplaceText(l.Value)
	case "plugin":
		if l.Value == true {
			return l.Name + " enabled"
		}
		return l.Name + ": " + schema.Show(l.Value)
	}
	return l.Name + ": " + schema.Show(l.Value)
}

// marketplaceText names each plugin of an inline marketplace and its pinned commit.
func marketplaceText(v any) string {
	o, _ := v.(schema.Object)
	src, _ := o.Get("source")
	so, _ := src.(schema.Object)
	pl, _ := so.Get("plugins")
	list, _ := pl.([]any)
	var parts []string
	for _, p := range list {
		po, _ := p.(schema.Object)
		ps, _ := po.Get("source")
		pso, _ := ps.(schema.Object)
		parts = append(parts, po.String("name")+" at "+short(pso.String("sha")))
	}
	if len(parts) == 0 {
		return schema.Show(v)
	}
	return strings.Join(parts, ", ")
}

// entry is the line's JSON value as written: a hook's command object.
func (l Line) hookEntry() schema.Object {
	o := schema.Object{{Key: "type", Value: "command"}, {Key: "command", Value: l.Command}}
	if l.Timeout != "" {
		o = append(o, schema.Member{Key: "timeout", Value: json.Number(l.Timeout)})
	}
	if l.Async {
		o = append(o, schema.Member{Key: "async", Value: true})
	}
	return o
}

func objectOf(kv ...any) schema.Object {
	o := schema.Object{}
	for i := 0; i+1 < len(kv); i += 2 {
		o = append(o, schema.Member{Key: kv[i].(string), Value: kv[i+1]})
	}
	return o
}

// packLines is what a pack gives the settings file: its plugin wiring, and its hook lines and deny rules when the
// pack was read at its commit (known).
type packLines struct {
	id, source, folder, commit string
	known                      bool
	hooks                      []workspace.HookEntry
	deny                       []workspace.DenyEntry
}

// MarketplaceName is the inline marketplace's name (spec §5): bonsai-<workspace name>-<8 hex>, the hex the first
// 8 characters of the SHA-256 of the packs' pins (Pin), each followed by a line feed, in bonsai.yaml's order. Two
// checkouts at the same commits and folders share it; a checkout whose update moved a pack to another commit or
// folder gets its own, so a plugin installed from the old folder is never taken for the new one's (step 5.1.1's
// verifier: "already installed" compared the commit, not the folder).
func MarketplaceName(name string, pins []string) string {
	var b strings.Builder
	for _, c := range pins {
		b.WriteString(c + "\n")
	}
	sum := sha256.Sum256([]byte(b.String()))
	return "bonsai-" + name + "-" + hex.EncodeToString(sum[:])[:8]
}

// Pin is what a pack adds to its marketplace's name: its locked commit, and, for a pack in a folder of its
// repository, a space and the folder. A pack at its repository's root pins its commit alone, so its marketplace keeps
// the name spec §5 gives it (the locked commits).
func Pin(commit, folder string) string {
	if folder == "" {
		return commit
	}
	return commit + " " + folder
}

// buildLines gives every line Bonsai writes for a workspace and its packs, in the order it writes them: the two
// keys, the deny rules (never_edit, then each pack's), the hook lines (Bonsai's own, then each pack's), the
// marketplace, then each pack's plugin. A line that two sources give is written once.
func buildLines(cfg *workspace.Config, packs []packLines) []Line {
	ls := []Line{
		{Kind: "key", Origin: "bonsai", Name: "autoMemoryEnabled", Value: false,
			Why: "Turns Claude Code's own auto memory off in this project; Bonsai keeps the project's memory itself."},
		{Kind: "key", Origin: "bonsai", Name: "disableAllHooks", Value: false,
			Why: "Keeps hooks on in this project, so Bonsai's guard runs even when a user setting turns hooks off."},
	}
	for _, p := range cfg.NeverEdit {
		ls = append(ls, Line{Kind: "deny", Origin: "bonsai.yaml", Rule: "Edit(" + p + ")",
			Why: "Agents can never edit " + p + ": it is in never_edit in bonsai.yaml."})
	}
	for _, p := range packs {
		for _, d := range p.deny {
			ls = append(ls, Line{Kind: "deny", Origin: p.id, Rule: d.Rule, Why: p.id + ": " + d.Why})
		}
	}
	ls = append(ls, ownHooks...)
	for _, p := range packs {
		for _, h := range p.hooks {
			ls = append(ls, Line{Kind: "hook", Origin: p.id, Event: h.Event, Matcher: h.Matcher, Command: h.Command,
				Why: p.id + ": " + h.Why})
		}
	}
	if len(packs) > 0 {
		var pins []string
		var plugins []any
		for _, p := range packs {
			pins = append(pins, Pin(p.commit, p.folder))
			plugins = append(plugins, objectOf("name", p.id, "source", pluginSource(p.source, p.folder, p.commit)))
		}
		market := MarketplaceName(cfg.Name, pins)
		ls = append(ls, Line{Kind: "marketplace", Origin: "bonsai", Name: market,
			Value: objectOf("source", objectOf("source", "settings", "name", market, "owner", objectOf("name", "Bonsai"),
				"plugins", plugins)),
			Why: "Tells Claude Code where to fetch this workspace's packs as plugins, each pinned to its locked commit."})
		for _, p := range packs {
			ls = append(ls, Line{Kind: "plugin", Origin: p.id, Name: p.id + "@" + market, Value: true,
				Why: "Turns on the " + p.id + " plugin from that marketplace, so its roles and skills load as " + p.id + ":<name>."})
		}
	}
	return slotted(dedupe(ls))
}

func dedupe(ls []Line) []Line {
	seen := map[string]bool{}
	var out []Line
	for _, l := range ls {
		if c := l.canon(); !seen[c] {
			seen[c] = true
			out = append(out, l)
		}
	}
	return out
}

// slotted gives each line its slot: a key, the marketplace and a plugin by name; a hook by its origin, event,
// matcher and place among those; a deny rule by itself.
func slotted(ls []Line) []Line {
	n := map[string]int{}
	for i, l := range ls {
		if l.Slot != "" {
			continue
		}
		switch l.Kind {
		case "key":
			ls[i].Slot = "key:" + l.Name
		case "marketplace":
			ls[i].Slot = "marketplace"
		case "plugin":
			ls[i].Slot = "plugin:" + strings.SplitN(l.Name, "@", 2)[0]
		case "deny":
			ls[i].Slot = "deny:" + l.Rule
		case "hook":
			k := "hook:" + l.Origin + ":" + l.Event + ":" + l.Matcher
			if l.Own {
				k = "hook:@own:" + l.Event + ":" + l.Matcher // no pack id holds an @
			}
			ls[i].Slot = k + ":" + strconv.Itoa(n[k])
			n[k]++
		}
	}
	return ls
}

// linesHash is the fingerprint the lock records for the settings file: the SHA-256 of the lines' canon forms,
// sorted and each once, each followed by a line feed. The order lines sit in the file does not count.
func linesHash(ls []Line) string {
	set := map[string]bool{}
	for _, l := range ls {
		set[l.canon()] = true
	}
	keys := make([]string, 0, len(set))
	for k := range set {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	sum := sha256.Sum256([]byte(strings.Join(keys, "\n") + "\n"))
	return hex.EncodeToString(sum[:])
}

// settingsDoc is the settings file as read.
type settingsDoc struct {
	root   schema.Object // the document; empty when the file is missing
	exists bool
	raw    []byte
}

func readSettings(root string) (*settingsDoc, error) {
	p := filepath.Join(root, filepath.FromSlash(SettingsFile))
	raw, err := os.ReadFile(p)
	if errors.Is(err, fs.ErrNotExist) {
		return &settingsDoc{root: schema.Object{}}, nil
	}
	next := "fix .claude/settings.json (Claude Code reads it too), then run the command again"
	if err != nil {
		return nil, errorf("read-failed", ExitRuntime, "check the file's permissions, then run the command again",
			"%s cannot be read: %v", SettingsFile, err)
	}
	v, err := schema.Decode(bytes.TrimPrefix(raw, []byte("\xef\xbb\xbf")))
	if err != nil {
		return nil, errorf("bad-file", ExitInput, next, "%s is not JSON Bonsai reads: %v", SettingsFile, err)
	}
	o, ok := v.(schema.Object)
	if !ok {
		return nil, errorf("bad-file", ExitInput, next, "%s is not a JSON object", SettingsFile)
	}
	for _, k := range []string{"permissions", "hooks", "extraKnownMarketplaces", "enabledPlugins"} {
		if x, ok := o.Get(k); ok {
			if _, isObj := x.(schema.Object); !isObj {
				return nil, errorf("bad-file", ExitInput, next, "%s: %s is %s, not an object", SettingsFile, k, schema.Show(x))
			}
		}
	}
	if perms, ok := o.Get("permissions"); ok {
		if d, ok := perms.(schema.Object).Get("deny"); ok {
			if _, isList := d.([]any); !isList {
				return nil, errorf("bad-file", ExitInput, next, "%s: permissions.deny is not a list", SettingsFile)
			}
		}
	}
	if hooks, ok := o.Get("hooks"); ok {
		for _, m := range hooks.(schema.Object) {
			if _, isList := m.Value.([]any); !isList {
				return nil, errorf("bad-file", ExitInput, next, "%s: hooks.%s is not a list", SettingsFile, m.Key)
			}
		}
	}
	return &settingsDoc{root: o, exists: true, raw: raw}, nil
}

// ownedKeys are the top-level keys that are Bonsai's whatever their value.
var ownedKeys = map[string]bool{"autoMemoryEnabled": true, "disableAllHooks": true}

// claudeKeyOrder is the order Claude Code writes the top-level keys of a settings file in, for the keys measured on
// Claude Code 2.1.294 and 2.1.295 (step 5.1.7: a project-scope install rewrote a file holding each of them, given in
// another order): Bonsai's six keys and the keys around them. A key not listed is never an anchor (placeKey).
var claudeKeyOrder = []string{"$schema", "respectGitignore", "cleanupPeriodDays", "env", "includeCoAuthoredBy",
	"includeGitInstructions", "permissions", "model", "enableAllProjectMcpServers", "enabledMcpjsonServers",
	"disabledMcpjsonServers", "hooks", "disableAllHooks", "enabledPlugins", "extraKnownMarketplaces", "outputStyle",
	"spinnerTipsEnabled", "alwaysThinkingEnabled", "autoMemoryEnabled"}

// claudePermissionsOrder is the order Claude Code writes permissions' keys in (measured as claudeKeyOrder).
var claudePermissionsOrder = []string{"allow", "deny", "ask", "defaultMode", "additionalDirectories"}

// claudeMarketplaceOrder is the order Claude Code writes an inline marketplace's source object in (measured as
// claudeKeyOrder): its plugins before its owner. Bonsai's Line.Value keeps its own order, which the lock's fingerprint
// was taken over; the file gets this one (claudeOrdered).
var claudeMarketplaceOrder = []string{"source", "name", "plugins", "owner"}

// placeKey gives key its value in o: in place when o holds it, else inserted where Claude Code writes it (order):
// after the last key of o that order puts before key, or first when there is none. A key order does not list is
// appended.
func placeKey(o schema.Object, key string, v any, order []string) schema.Object {
	if i := o.Index(key); i >= 0 {
		o[i].Value = v
		return o
	}
	rank := map[string]int{}
	for i, k := range order {
		rank[k] = i
	}
	r, known := rank[key]
	if !known {
		return append(o, schema.Member{Key: key, Value: v})
	}
	at := 0
	for i, m := range o {
		if mr, ok := rank[m.Key]; ok && mr < r {
			at = i + 1
		}
	}
	out := make(schema.Object, 0, len(o)+1)
	out = append(out, o[:at]...)
	out = append(out, schema.Member{Key: key, Value: v})
	return append(out, o[at:]...)
}

// claudeOrdered gives a line's value as the file holds it: a marketplace's source object in Claude Code's order
// (claudeMarketplaceOrder), every other value as it is. It never changes the line's own value.
func claudeOrdered(l Line) any {
	o, ok := l.Value.(schema.Object)
	if l.Kind != "marketplace" || !ok {
		return l.Value
	}
	src, ok := o.Get("source")
	so, isObj := src.(schema.Object)
	if !ok || !isObj {
		return l.Value
	}
	var inner schema.Object
	for _, k := range claudeMarketplaceOrder {
		if v, ok := so.Get(k); ok {
			inner = append(inner, schema.Member{Key: k, Value: v})
		}
	}
	for _, m := range so {
		if inner.Index(m.Key) < 0 {
			inner = append(inner, m)
		}
	}
	out := cloneObject(o)
	out[out.Index("source")].Value = inner
	return out
}

// bonsaiMarketplaceOrder is the order buildLines writes an inline marketplace's source object in, which the lock's
// fingerprint of Bonsai's lines is taken over.
var bonsaiMarketplaceOrder = []string{"source", "name", "owner", "plugins"}

// bonsaiOrdered gives a marketplace's value read from the file with its source object in Bonsai's own order
// (bonsaiMarketplaceOrder), other keys after them as the file has them; any other value as it is.
func bonsaiOrdered(v any) any {
	o, ok := v.(schema.Object)
	if !ok {
		return v
	}
	i := o.Index("source")
	if i < 0 {
		return v
	}
	so, ok := o[i].Value.(schema.Object)
	if !ok {
		return v
	}
	var inner schema.Object
	for _, k := range bonsaiMarketplaceOrder {
		if x, ok := so.Get(k); ok {
			inner = append(inner, schema.Member{Key: k, Value: x})
		}
	}
	for _, m := range so {
		if inner.Index(m.Key) < 0 {
			inner = append(inner, m)
		}
	}
	out := cloneObject(o)
	out[i].Value = inner
	return out
}

// sameValue reports whether two JSON values are equal but for the order of their objects' keys.
func sameValue(a, b any) bool { return schema.Show(sortedKeys(a)) == schema.Show(sortedKeys(b)) }

// diskLines lists every line the file holds that Bonsai could own: the two keys, each deny rule, each command hook,
// each marketplace and each enabled plugin.
func diskLines(root schema.Object) []Line {
	var ls []Line
	for _, m := range root {
		if ownedKeys[m.Key] {
			ls = append(ls, Line{Kind: "key", Name: m.Key, Value: m.Value})
		}
	}
	if perms, ok := root.Get("permissions"); ok {
		d, _ := perms.(schema.Object).Get("deny")
		list, _ := d.([]any)
		for _, r := range list {
			if s, ok := r.(string); ok {
				ls = append(ls, Line{Kind: "deny", Rule: s})
			}
		}
	}
	if hooks, ok := root.Get("hooks"); ok {
		for _, ev := range hooks.(schema.Object) {
			groups, _ := ev.Value.([]any)
			for _, g := range groups {
				go_, ok := g.(schema.Object)
				if !ok {
					continue
				}
				matcher := go_.String("matcher")
				hs, _ := go_.Get("hooks")
				list, _ := hs.([]any)
				for _, h := range list {
					if l, ok := hookLine(ev.Key, matcher, h); ok {
						ls = append(ls, l)
					}
				}
			}
		}
	}
	for _, key := range []string{"extraKnownMarketplaces", "enabledPlugins"} {
		if x, ok := root.Get(key); ok {
			kind := "marketplace"
			if key == "enabledPlugins" {
				kind = "plugin"
			}
			for _, m := range x.(schema.Object) {
				ls = append(ls, Line{Kind: kind, Name: m.Key, Value: m.Value})
			}
		}
	}
	return ls
}

// hookLine reads one hook entry as a line; an entry that is not a command hook is the project's and no line.
func hookLine(event, matcher string, h any) (Line, bool) {
	o, ok := h.(schema.Object)
	if !ok || o.String("type") != "command" {
		return Line{}, false
	}
	cmd, ok := o.Get("command")
	c, isText := cmd.(string)
	if !ok || !isText {
		return Line{}, false
	}
	l := Line{Kind: "hook", Event: event, Matcher: matcher, Command: c}
	if t, ok := o.Get("timeout"); ok {
		if n, isNum := t.(json.Number); isNum {
			l.Timeout = string(n)
		}
	}
	if a, ok := o.Get("async"); ok && a == true {
		l.Async = true
	}
	return l, true
}

// isBonsaiHook reports a hook line that runs bonsai by name, as Bonsai's lines do (spec §3).
func isBonsaiHook(command string) bool {
	c := strings.TrimSpace(command)
	return c == "bonsai hook" || strings.HasPrefix(c, "bonsai hook ")
}

// isOldBonsaiHook reports an old Bonsai hook line, one that calls Bonsai by an absolute path: a bonsai binary by
// its full path (/.../bonsai, C:\...\bonsai.exe, ~/.../bonsai), or Bonsai 0.4.3's sensor scripts
// (bash "<absolute>/agent/Sensors/<name>.sh" "<absolute>"). init takes such a line out when it links a project
// (spec §14, check 1: "the old absolute line is gone"); Bonsai's own lines call bonsai by name.
func isOldBonsaiHook(command string) bool {
	words := shellWords(command)
	if len(words) == 0 {
		return false
	}
	if absolutePath(words[0]) {
		seg := words[0]
		if i := strings.LastIndexAny(seg, `/\`); i >= 0 {
			seg = seg[i+1:]
		}
		if s := strings.ToLower(seg); s == "bonsai" || s == "bonsai.exe" {
			return true
		}
	}
	if (words[0] == "bash" || words[0] == "sh") && len(words) > 1 && absolutePath(words[1]) &&
		strings.Contains(filepath.ToSlash(words[1]), "/agent/Sensors/") && strings.HasSuffix(words[1], ".sh") {
		return true
	}
	return false
}

func absolutePath(p string) bool {
	if strings.HasPrefix(p, "/") || strings.HasPrefix(p, "~/") || strings.HasPrefix(p, "$HOME/") ||
		strings.HasPrefix(strings.ToUpper(p), "%USERPROFILE%") {
		return true
	}
	return len(p) >= 3 && ((p[0] >= 'A' && p[0] <= 'Z') || (p[0] >= 'a' && p[0] <= 'z')) && p[1] == ':' &&
		(p[2] == '\\' || p[2] == '/')
}

// shellWords splits a command line into words the way a shell would for plain words and quoted ones; enough to
// find a command's program and first argument.
func shellWords(s string) []string {
	var words []string
	var cur strings.Builder
	in, quote := false, byte(0)
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case quote != 0:
			if c == quote {
				quote = 0
			} else {
				cur.WriteByte(c)
			}
		case c == '"' || c == '\'':
			quote, in = c, true
		case c == ' ' || c == '\t':
			if in {
				words = append(words, cur.String())
				cur.Reset()
				in = false
			}
		default:
			cur.WriteByte(c)
			in = true
		}
	}
	if in {
		words = append(words, cur.String())
	}
	return words
}

// claim finds Bonsai's lines in the file: those the reference lines (what Bonsai would have written from the
// lock's commits) hold, the lines no person writes, and, at a first link, old Bonsai hook lines. It says whether
// they are exactly what the lock says Bonsai last wrote (lockHash, "" for none), trying the file's other Edit
// rules, in each combination, as never_edit rules a person has since taken out of bonsai.yaml.
func claim(disk, ref []Line, firstLink bool, lockHash string) ([]Line, bool) {
	refs := map[string]Line{}
	for _, r := range ref {
		refs[r.match()] = r
	}
	var claimed, spare []Line
	seen := map[string]bool{}
	add := func(l Line) {
		if c := l.canon(); !seen[c] {
			seen[c] = true
			claimed = append(claimed, l)
		}
	}
	for _, d := range disk {
		if r, ok := refs[d.match()]; ok {
			add(r)
			continue
		}
		switch {
		case d.Kind == "key":
			d.Origin, d.Slot, d.Why = "bonsai", "key:"+d.Name, "Bonsai's own setting."
			add(d)
		case d.Kind == "marketplace" && strings.HasPrefix(d.Name, "bonsai-"):
			// Its value as Bonsai writes it (bonsaiOrdered): Claude Code moves the owner after the plugins, which is no
			// edit, and the fingerprint is taken over Bonsai's order.
			d.Origin, d.Slot, d.Value = "bonsai", "marketplace", bonsaiOrdered(d.Value)
			d.Why = "An old plugin marketplace of Bonsai's for this workspace: the one above replaces it."
			add(d)
		case d.Kind == "plugin" && strings.Contains(d.Name, "@bonsai-"):
			d.Origin, d.Slot = "bonsai", "plugin:"+strings.SplitN(d.Name, "@", 2)[0]
			d.Why = "Enables a plugin from an old marketplace of Bonsai's."
			add(d)
		case d.Kind == "hook" && isBonsaiHook(d.Command):
			d.Origin, d.Own, d.Slot = "bonsai", true, "hook:@own:"+d.Event+":"+d.Matcher+":0"
			d.Why = "A Bonsai hook line this build does not write."
			add(d)
		case d.Kind == "hook" && firstLink && isOldBonsaiHook(d.Command):
			d.Origin, d.Slot = "old", "old:"+d.canon()
			d.Why = "An old Bonsai hook line that calls bonsai by an absolute path; Bonsai's lines call it by name."
			add(d)
		case d.Kind == "deny" && strings.HasPrefix(d.Rule, "Edit(") && strings.HasSuffix(d.Rule, ")"):
			d.Origin, d.Slot = "bonsai.yaml", "deny:"+d.Rule
			d.Why = "This file is no longer in never_edit in bonsai.yaml, so agents may edit it again."
			spare = append(spare, d)
		}
	}
	if lockHash == "" {
		return claimed, false
	}
	if linesHash(claimed) == lockHash {
		return claimed, true
	}
	var rest []Line
	for _, s := range spare {
		if !seen[s.canon()] {
			seen[s.canon()] = true
			rest = append(rest, s)
		}
	}
	// Each set of the other Edit rules, smallest first, while there are few enough to try them all (2^12 hashes);
	// past that, only all of them. A match is exact: the fingerprint is a SHA-256.
	if len(rest) > maxSpare {
		if wider := append(append([]Line{}, claimed...), rest...); linesHash(wider) == lockHash {
			return wider, true
		}
		return claimed, false
	}
	masks := make([]int, 0, 1<<len(rest))
	for m := 1; m < 1<<len(rest); m++ {
		masks = append(masks, m)
	}
	sort.SliceStable(masks, func(i, j int) bool { return bits(masks[i]) < bits(masks[j]) })
	for _, m := range masks {
		wider := append([]Line{}, claimed...)
		for i, s := range rest {
			if m&(1<<i) != 0 {
				wider = append(wider, s)
			}
		}
		if linesHash(wider) == lockHash {
			return wider, true
		}
	}
	return claimed, false
}

// maxSpare is how many of the file's other Edit rules claim tries in every combination.
const maxSpare = 12

func bits(m int) int {
	n := 0
	for ; m > 0; m &= m - 1 {
		n++
	}
	return n
}

// SettingsChange is one line the preview names: added, changed or removed, with its sentence (spec §6).
type SettingsChange struct {
	Change   string // add, change or remove
	Kind     string // hook, deny, key, marketplace or plugin
	Line     string // the line as written after the change (or, removed, as it was)
	Was      string // change: the line before
	Why      string
	RunsCode bool   // a hook line added or changed, and not one the lock last consented to: it runs code (spec §6)
	Origin   string // where the line comes from (Line.Origin): bonsai for Bonsai's own, or a pack id
	Own      bool   // Bonsai's own hook line (Line.Own), which no pack can set
}

// lineChanges lists the changes from Bonsai's lines in the file (claimed) to the new lines: a new line the file
// already holds (a person wrote the same line) is no change. reference holds the lines the lock last consented to,
// so a hook line that only comes back is not new code.
func lineChanges(claimed, lnew []Line, disk []Line, consented []Line) []SettingsChange {
	inNew, inClaimed, onDisk, ok := map[string]bool{}, map[string]bool{}, map[string]bool{}, map[string]bool{}
	for _, l := range lnew {
		inNew[l.canon()] = true
	}
	for _, l := range claimed {
		inClaimed[l.canon()] = true
	}
	for _, l := range disk {
		onDisk[l.canon()] = true
	}
	for _, l := range consented {
		ok[l.canon()] = true
	}
	var removed []Line
	for _, l := range claimed {
		if !inNew[l.canon()] {
			removed = append(removed, l)
		}
	}
	var out []SettingsChange
	for _, l := range lnew {
		c := l.canon()
		if inClaimed[c] || onDisk[c] {
			continue
		}
		sc := SettingsChange{Change: "add", Kind: l.Kind, Line: l.Text(), Why: l.Why,
			RunsCode: l.Kind == "hook" && !ok[c], Origin: l.Origin, Own: l.Own}
		for i, r := range removed {
			if r.Slot == l.Slot {
				sc.Change, sc.Was = "change", r.Text()
				removed = append(removed[:i], removed[i+1:]...)
				break
			}
		}
		out = append(out, sc)
	}
	for _, r := range removed {
		out = append(out, SettingsChange{Change: "remove", Kind: r.Kind, Line: r.Text(), Why: r.Why, Origin: r.Origin, Own: r.Own})
	}
	return out
}

// applyLines writes the new lines into the document in place of the claimed ones: a claimed line the new lines
// still hold stays where it is; the others go; a new line is added, the keys, marketplaces and plugins in place
// where their key was (else at the end), the deny rules and hook groups at the end of their lists. Every other
// entry of the file stays as it was, in its place. A list or object that held only Bonsai's lines goes with them.
func applyLines(root schema.Object, claimed, lnew []Line) schema.Object {
	doc := cloneObject(root)
	stays := map[string]bool{}
	for _, l := range lnew {
		stays[l.canon()] = true
	}
	gone := map[string]bool{}
	for _, l := range claimed {
		if !stays[l.canon()] {
			gone[l.canon()] = true
		}
	}
	newKeyed := map[string]map[string]any{"key": {}, "marketplace": {}, "plugin": {}}
	for _, l := range lnew {
		if m, ok := newKeyed[l.Kind]; ok {
			m[l.Name] = l.Value
		}
	}

	// Keys: set in place, or placed where Claude Code writes them; a claimed key no new line holds goes.
	for _, l := range claimed {
		if l.Kind == "key" {
			if _, keep := newKeyed["key"][l.Name]; !keep {
				doc = without(doc, l.Name)
			}
		}
	}
	for _, l := range lnew {
		if l.Kind == "key" {
			doc = placeKey(doc, l.Name, l.Value, claudeKeyOrder)
		}
	}

	// Deny rules.
	perms, _ := getObject(doc, "permissions")
	denyList, _ := perms.Get("deny")
	list, _ := denyList.([]any)
	var kept []any
	removedAny := false
	for _, r := range list {
		if s, ok := r.(string); ok && gone[(Line{Kind: "deny", Rule: s}).canon()] {
			removedAny = true
			continue
		}
		kept = append(kept, r)
	}
	present := map[string]bool{}
	for _, r := range kept {
		if s, ok := r.(string); ok {
			present[s] = true
		}
	}
	for _, l := range lnew {
		if l.Kind == "deny" && !present[l.Rule] {
			kept = append(kept, l.Rule)
			present[l.Rule] = true
		}
	}
	if len(kept) > 0 {
		perms = placeKey(perms, "deny", kept, claudePermissionsOrder)
	} else if removedAny {
		perms = without(perms, "deny")
	}
	doc = putObject(doc, "permissions", perms, removedAny)

	// Hook lines: claimed entries leave their groups; the new lines come as one group per event and matcher.
	hooks, _ := getObject(doc, "hooks")
	hooksTouched := false
	for i, ev := range hooks {
		groups, _ := ev.Value.([]any)
		var keptGroups []any
		for _, g := range groups {
			gObj, ok := g.(schema.Object)
			if !ok {
				keptGroups = append(keptGroups, g)
				continue
			}
			hs, _ := gObj.Get("hooks")
			hl, isList := hs.([]any)
			if !isList {
				keptGroups = append(keptGroups, g)
				continue
			}
			var keptHooks []any
			for _, h := range hl {
				if l, ok := hookLine(ev.Key, gObj.String("matcher"), h); ok && gone[l.canon()] {
					continue
				}
				keptHooks = append(keptHooks, h)
			}
			switch {
			case len(keptHooks) == len(hl):
				keptGroups = append(keptGroups, g)
			case len(keptHooks) > 0:
				keptGroups = append(keptGroups, set(gObj, "hooks", keptHooks))
				hooksTouched = true
			default:
				hooksTouched = true
			}
		}
		if keptGroups == nil {
			keptGroups = []any{}
		}
		hooks[i].Value = keptGroups
	}
	// Events left empty by the removal go; an event that was empty before stays.
	var keptEvents schema.Object
	for _, ev := range hooks {
		if g, _ := ev.Value.([]any); len(g) == 0 {
			if before, ok := getList(root, "hooks", ev.Key); ok && len(before) > 0 {
				continue
			}
		}
		keptEvents = append(keptEvents, ev)
	}
	hooks = keptEvents
	presentHook := map[string]bool{}
	for _, l := range diskLines(schema.Object{{Key: "hooks", Value: hooks}}) {
		presentHook[l.canon()] = true
	}
	type groupKey struct{ event, matcher string }
	var order []groupKey
	groups := map[groupKey][]any{}
	for _, l := range lnew {
		if l.Kind != "hook" || presentHook[l.canon()] {
			continue
		}
		k := groupKey{l.Event, l.Matcher}
		if _, ok := groups[k]; !ok {
			order = append(order, k)
		}
		groups[k] = append(groups[k], l.hookEntry())
	}
	for _, k := range order {
		g := schema.Object{}
		if k.matcher != "" {
			g = append(g, schema.Member{Key: "matcher", Value: k.matcher})
		}
		g = append(g, schema.Member{Key: "hooks", Value: groups[k]})
		existing, _ := hooks.Get(k.event)
		el, _ := existing.([]any)
		hooks = set(hooks, k.event, append(append([]any{}, el...), g))
	}
	doc = putObject(doc, "hooks", hooks, hooksTouched)

	// Marketplaces and plugins: a claimed entry no new line holds goes, and a new line of the same slot (the
	// marketplace, or a pack's plugin, under a new name) takes its place; an entry that stays keeps its place, and its
	// bytes when its value is the same but for key order; any other new line is appended.
	for _, kind := range []string{"marketplace", "plugin"} {
		key := map[string]string{"marketplace": "extraKnownMarketplaces", "plugin": "enabledPlugins"}[kind]
		obj, _ := getObject(doc, key)
		var fresh []Line
		for _, l := range lnew {
			if l.Kind == kind {
				fresh = append(fresh, l)
			}
		}
		placed := map[string]bool{}
		goneSlot := map[string]string{} // a claimed entry's name that goes -> its slot
		for _, l := range claimed {
			if _, keep := newKeyed[kind][l.Name]; l.Kind == kind && !keep {
				goneSlot[l.Name] = l.Slot
			}
		}
		touched := false
		var out schema.Object
		for _, m := range obj {
			if slot, gone := goneSlot[m.Key]; gone {
				touched = true
				for _, l := range fresh {
					if l.Slot == slot && !placed[l.Name] && obj.Index(l.Name) < 0 {
						out = append(out, schema.Member{Key: l.Name, Value: claudeOrdered(l)})
						placed[l.Name] = true
						break
					}
				}
				continue
			}
			for _, l := range fresh {
				if l.Name == m.Key {
					placed[l.Name] = true
					if !sameValue(m.Value, l.Value) {
						m.Value = claudeOrdered(l)
					}
				}
			}
			out = append(out, m)
		}
		for _, l := range fresh {
			if !placed[l.Name] {
				out = set(out, l.Name, claudeOrdered(l))
				placed[l.Name] = true
			}
		}
		doc = putObject(doc, key, out, touched)
	}
	return doc
}

func cloneObject(o schema.Object) schema.Object {
	out := make(schema.Object, len(o))
	for i, m := range o {
		out[i] = schema.Member{Key: m.Key, Value: cloneValue(m.Value)}
	}
	return out
}

func cloneValue(v any) any {
	switch x := v.(type) {
	case schema.Object:
		return cloneObject(x)
	case []any:
		out := make([]any, len(x))
		for i, e := range x {
			out[i] = cloneValue(e)
		}
		return out
	}
	return v
}

func getObject(o schema.Object, key string) (schema.Object, bool) {
	v, ok := o.Get(key)
	obj, _ := v.(schema.Object)
	return obj, ok
}

func getList(o schema.Object, key, inner string) ([]any, bool) {
	obj, ok := getObject(o, key)
	if !ok {
		return nil, false
	}
	v, ok := obj.Get(inner)
	l, _ := v.([]any)
	return l, ok
}

// set gives a key its value in place, or appends it.
func set(o schema.Object, key string, v any) schema.Object {
	if i := o.Index(key); i >= 0 {
		o[i].Value = v
		return o
	}
	return append(o, schema.Member{Key: key, Value: v})
}

func without(o schema.Object, key string) schema.Object {
	if i := o.Index(key); i >= 0 {
		return append(o[:i:i], o[i+1:]...)
	}
	return o
}

// putObject stores an inner object, in place or where Claude Code writes it (placeKey): an empty one is left out
// when the removal emptied it or it was not there.
func putObject(doc schema.Object, key string, inner schema.Object, removed bool) schema.Object {
	_, had := doc.Get(key)
	if len(inner) == 0 && (removed || !had) {
		return without(doc, key)
	}
	if len(inner) == 0 {
		return placeKey(doc, key, schema.Object{}, claudeKeyOrder)
	}
	return placeKey(doc, key, inner, claudeKeyOrder)
}
