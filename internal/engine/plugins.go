package engine

// This machine's plugins (spec §5, "This machine's install follows the lock" and "Drift is reported, never silent";
// plan part 4b). A pack reaches a session as a Claude Code plugin from the workspace's own inline marketplace, which
// settings.go writes into the shared .claude/settings.json: bonsai-<workspace name>-<8 hex>, each plugin pinned to
// its locked commit by sha. What Claude Code does with it, as part 4b found it (Claude Code 2.1.294):
//
//   - A marketplace that only a project's settings declare is registered by a Claude Code session in that folder,
//     and only once the folder is trusted (an interactive session asks; -p and --bg never do). Neither `claude plugin
//     install` nor `claude plugin marketplace update` registers it, so right after init or update moves a commit (a
//     new marketplace name) the install fails "not found" until a session has started there.
//   - Claude Code fetches a plugin by itself only when a file outside git, the user's settings or a flag turns it on,
//     never the shared settings alone. So Bonsai turns each locked plugin on in this checkout's untracked
//     .claude/settings.local.json too (the spec's other route): the first trusted session registers the marketplace
//     and fetches the plugin at its commit.
//   - `claude plugin install --scope project` rewrites the committed .claude/settings.json (it reorders its keys);
//     `--scope local` writes only the untracked .claude/settings.local.json. Bonsai installs at local scope, never
//     user: each checkout (a worktree included) gets its own record and its own enable.
//
// So init and update write the local enable as one more settings line (previewed, written with --yes), then ask
// Claude Code to install each plugin (InstallPlugins); check compares what Claude Code reports with the lock
// (ComparePlugins). Both go through PluginCLI: ClaudeCLI runs the real `claude`, tests use a fake. A nil PluginCLI
// does neither (Bonsai's tests, and a build that has no Claude Code to ask).

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// LocalSettingsFile is this checkout's own Claude Code settings file, never committed (Claude Code's
// .claude/settings.local.json). Bonsai writes one kind of line in it: each locked pack's plugin turned on.
const LocalSettingsFile = ".claude/settings.local.json"

// PluginScope is the scope Bonsai installs plugins at: local, this checkout only. Never user, which writes the
// person's own ~/.claude/settings.json (plan, "Claude Code's own files").
const PluginScope = "local"

// InstalledPlugin is one plugin as `claude plugin list --json` reports it.
type InstalledPlugin struct {
	ID          string `json:"id"`          // <plugin>@<marketplace>
	Version     string `json:"version"`     // with no version in plugin.json: the commit's first 12 characters
	Scope       string `json:"scope"`       // user, project, local or managed
	Enabled     bool   `json:"enabled"`     // turned on in the settings the folder the list ran in reads
	ProjectPath string `json:"projectPath"` // project and local scope: the folder it was installed for
}

// InstallResult is `claude plugin install --json`'s one result line.
type InstallResult struct {
	Outcome     string `json:"outcome"`     // ok or failed
	FailureCode string `json:"failureCode"` // when failed: not_found, ...
	Message     string `json:"message"`
}

// ErrNoClaude is PluginCLI's error when Claude Code is not on the PATH.
var ErrNoClaude = errors.New("claude is not on the PATH")

// PluginCLI is the part of Claude Code's command line Bonsai uses. Each call runs in dir, a checkout's top folder.
type PluginCLI interface {
	List(dir string) ([]InstalledPlugin, error)
	Install(dir, plugin string) (InstallResult, error)
}

// ClaudeCLI runs the `claude` on the PATH, with the process's environment (CLAUDE_CODE_PLUGIN_CACHE_DIR passes
// through) and no input.
type ClaudeCLI struct{}

// The longest a claude command may take: an install clones the pack's repository.
const (
	listTimeout    = 60 * time.Second
	installTimeout = 300 * time.Second
)

func (ClaudeCLI) run(dir string, timeout time.Duration, args ...string) ([]byte, string, error) {
	bin, err := exec.LookPath("claude")
	if err != nil {
		return nil, "", ErrNoClaude
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err = cmd.Run()
	if ctx.Err() != nil {
		return stdout.Bytes(), "", errors.New("claude " + strings.Join(args, " ") + " took longer than " + timeout.String())
	}
	return stdout.Bytes(), firstLine(stderr.String()), err
}

// List runs `claude plugin list --json`.
func (c ClaudeCLI) List(dir string) ([]InstalledPlugin, error) {
	out, msg, err := c.run(dir, listTimeout, "plugin", "list", "--json")
	if err != nil {
		if errors.Is(err, ErrNoClaude) {
			return nil, err
		}
		return nil, errors.New("claude plugin list --json failed: " + msg)
	}
	return ParsePluginList(out)
}

// Install runs `claude plugin install <plugin> --scope local --json`. A refusal Claude Code reports on its result
// line (exit 1) is a result, not an error.
func (c ClaudeCLI) Install(dir, plugin string) (InstallResult, error) {
	out, msg, err := c.run(dir, installTimeout, "plugin", "install", plugin, "--scope", PluginScope, "--json")
	if errors.Is(err, ErrNoClaude) {
		return InstallResult{}, err
	}
	if r, ok := ParseInstallResult(out); ok {
		return r, nil
	}
	if err != nil {
		return InstallResult{}, errors.New("claude plugin install failed: " + msg)
	}
	return InstallResult{}, errors.New("claude plugin install printed no result line")
}

// ParsePluginList reads `claude plugin list --json`: a JSON list of installed plugins.
func ParsePluginList(out []byte) ([]InstalledPlugin, error) {
	var list []InstalledPlugin
	if err := json.Unmarshal(bytes.TrimSpace(out), &list); err != nil {
		return nil, errors.New("claude plugin list --json printed no list Bonsai reads: " + err.Error())
	}
	return list, nil
}

// ParseInstallResult finds `claude plugin install --json`'s result line: the last stdout line that is a JSON object
// with an outcome.
func ParseInstallResult(out []byte) (InstallResult, bool) {
	lines := strings.Split(strings.ReplaceAll(string(out), "\r\n", "\n"), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		l := strings.TrimSpace(lines[i])
		if !strings.HasPrefix(l, "{") {
			continue
		}
		var r InstallResult
		if json.Unmarshal([]byte(l), &r) == nil && r.Outcome != "" {
			return r, true
		}
	}
	return InstallResult{}, false
}

// PluginID is a pack's plugin as Claude Code names it: <pack id>@<marketplace>.
func PluginID(pack, market string) string { return pack + "@" + market }

// lockMarket is the marketplace name a lock's packs give, in the lock's order (bonsai.yaml's).
func lockMarket(cfg *workspace.Config, lock *workspace.Lock) string {
	var commits []string
	for _, lp := range lock.Packs {
		commits = append(commits, lp.Commit)
	}
	return MarketplaceName(cfg.Name, commits)
}

// pluginVersion is the version Claude Code computes for a plugin with no version in plugin.json: the commit's first
// 12 characters (a git-subdir source adds a hash of its folder, so only the start is compared).
func pluginVersion(commit string) string {
	if len(commit) > 12 {
		return commit[:12]
	}
	return commit
}

// bonsaiPlugin matches a plugin turned on from one of a workspace's marketplaces: <pack id>@bonsai-<name>-<8 hex>.
func bonsaiPlugin(name string) *regexp.Regexp {
	return regexp.MustCompile(`^([a-z0-9][a-z0-9-]*)@bonsai-` + regexp.QuoteMeta(name) + `-[0-9a-f]{8}$`)
}

// localLines works out Bonsai's lines in .claude/settings.local.json: each pack's plugin turned on from the
// marketplace of the commits the plan locks. A line of Bonsai's from another of the workspace's marketplaces (an
// earlier commit) goes; a person's own lines stay, and so does one that turns the plugin off (theirs to decide).
func localLines(root, name string, packIDs []string, market string) (raw []byte, doc schema.Object, changes []SettingsChange, err error) {
	raw, exists, err := readFile(root, LocalSettingsFile)
	if err != nil {
		return nil, nil, nil, err
	}
	next := "fix " + LocalSettingsFile + " (Claude Code reads it too), or delete it, then run the command again"
	doc = schema.Object{}
	if exists {
		v, derr := schema.Decode(bytes.TrimPrefix(raw, []byte("\xef\xbb\xbf")))
		if derr != nil {
			return nil, nil, nil, errorf(ExitInput, next, "%s is not JSON Bonsai reads: %v", LocalSettingsFile, derr)
		}
		o, ok := v.(schema.Object)
		if !ok {
			return nil, nil, nil, errorf(ExitInput, next, "%s is not a JSON object", LocalSettingsFile)
		}
		doc = o
	} else {
		raw = nil
	}
	ep, has := doc.Get("enabledPlugins")
	enabled, isObj := ep.(schema.Object)
	if has && !isObj {
		return nil, nil, nil, errorf(ExitInput, next, "%s: enabledPlugins is %s, not an object", LocalSettingsFile, schema.Show(ep))
	}
	wanted := map[string]string{} // pack id -> its plugin at the locked commits
	for _, id := range packIDs {
		wanted[id] = PluginID(id, market)
	}
	ours := bonsaiPlugin(name)
	addWhy := func(id string) string {
		return "Turns the " + id + " plugin on in this checkout only, so Claude Code fetches it at the locked commit; " +
			"it does not fetch a plugin that only the shared settings turn on. Never committed."
	}
	// A stale line of a pack still locked gives its place to the new one (a change); any other stale line goes.
	var kept schema.Object
	placed := map[string]bool{}
	for _, m := range enabled {
		sub := ours.FindStringSubmatch(m.Key)
		if sub == nil || wanted[sub[1]] == m.Key {
			kept = append(kept, m)
			if sub != nil {
				placed[sub[1]] = true
			}
			continue
		}
		was := (Line{Kind: "plugin", Name: m.Key, Value: m.Value}).Text()
		if key, ok := wanted[sub[1]]; ok && !placed[sub[1]] && enabled.Index(key) < 0 {
			placed[sub[1]] = true
			kept = append(kept, schema.Member{Key: key, Value: true})
			changes = append(changes, SettingsChange{File: LocalSettingsFile, Change: "change", Kind: "plugin",
				Line: key + " enabled", Was: was, Why: addWhy(sub[1])})
			continue
		}
		changes = append(changes, SettingsChange{File: LocalSettingsFile, Change: "remove", Kind: "plugin", Line: was,
			Why: "Turned on the " + sub[1] + " plugin at commits this checkout no longer locks."})
	}
	for _, id := range packIDs {
		if placed[id] {
			continue
		}
		key := wanted[id]
		kept = append(kept, schema.Member{Key: key, Value: true})
		changes = append(changes, SettingsChange{File: LocalSettingsFile, Change: "add", Kind: "plugin", Line: key + " enabled",
			Why: addWhy(id)})
	}
	if len(changes) == 0 {
		return raw, nil, nil, nil
	}
	if len(kept) == 0 {
		doc = without(cloneObject(doc), "enabledPlugins")
	} else {
		doc = set(cloneObject(doc), "enabledPlugins", kept)
	}
	return raw, doc, changes, nil
}

// PluginResult is what InstallPlugins did for one pack.
type PluginResult struct {
	Pack    string
	Plugin  string // <pack>@<marketplace>
	Commit  string
	Result  string // installed, waiting, failed or skipped
	Message string
	Next    string
}

// InstallPlugins asks Claude Code to install each locked pack's plugin for the checkout at root, at local scope
// (spec §5: `claude plugin install` is a no-op once installed, and the marketplace name is new whenever a locked
// commit changed). It never fails the command: the project's files are written; what this machine still needs is
// in each result's next step. A nil cli installs nothing.
func InstallPlugins(root string, cfg *workspace.Config, lock *workspace.Lock, cli PluginCLI) []PluginResult {
	if cli == nil || cfg == nil || lock == nil || len(lock.Packs) == 0 {
		return nil
	}
	market := lockMarket(cfg, lock)
	var out []PluginResult
	for _, lp := range lock.Packs {
		id := PluginID(lp.ID, market)
		cmd := "claude plugin install " + id + " --scope " + PluginScope
		r := PluginResult{Pack: lp.ID, Plugin: id, Commit: lp.Commit}
		res, err := cli.Install(root, id)
		switch {
		case errors.Is(err, ErrNoClaude):
			r.Result, r.Message = "skipped", "Claude Code is not on the PATH, so the plugin was not installed on this machine"
			r.Next = "where sessions run, install Claude Code: its first session in this checkout fetches the plugin (" +
				LocalSettingsFile + " turns it on)"
		case err != nil:
			r.Result, r.Message, r.Next = "failed", err.Error(), "run: "+cmd
		case res.Outcome == "ok":
			r.Result, r.Message = "installed", "at "+pluginVersion(lp.Commit)
		case res.FailureCode == "not_found":
			r.Result = "waiting"
			r.Message = "Claude Code does not know the marketplace " + market + " yet: a session registers it, once this folder is trusted"
			r.Next = "open Claude Code in this checkout and accept its trust question if it asks: the session fetches the plugin at " +
				pluginVersion(lp.Commit) + " (" + LocalSettingsFile + " turns it on); bonsai check then shows it installed"
		default:
			r.Result, r.Message, r.Next = "failed", strings.TrimSpace(res.Message), "run: "+cmd
		}
		r.Message, r.Next = ascii(r.Message), ascii(r.Next)
		out = append(out, r)
	}
	return out
}

// ComparePlugins adds check's plugin findings and warnings (spec §5, "Drift is reported, never silent"): it asks
// Claude Code which plugins it has for this checkout and compares their commits with the lock's. A plugin of a
// locked pack, turned on here from one of the workspace's marketplaces at another commit, is drift: sessions here
// load it instead of the lock's (a finding). The lock's plugin not installed for this checkout is a warning: the
// project is right, and the next trusted session fetches it. A nil cli compares nothing.
func ComparePlugins(r *CheckResult, cli PluginCLI) {
	if cli == nil || r == nil || r.Config == nil || r.Lock == nil || len(r.Lock.Packs) == 0 {
		return
	}
	list, err := cli.List(r.Root)
	if err != nil {
		w := Finding{Code: "plugin", File: LocalSettingsFile,
			Message: "this machine's plugins were not compared with the lock: " + err.Error(),
			Next:    "run claude plugin list --json in this checkout to see why, then check again"}
		if errors.Is(err, ErrNoClaude) {
			w.Message = "Claude Code is not on the PATH, so this machine's plugins were not compared with the lock"
			w.Next = "where sessions run, install Claude Code and check again; where none run (CI), nothing is needed"
		}
		r.Warnings = append(r.Warnings, w)
		return
	}
	market := lockMarket(r.Config, r.Lock)
	ours := bonsaiPlugin(r.Config.Name)
	for _, lp := range r.Lock.Packs {
		want := PluginID(lp.ID, market)
		version := pluginVersion(lp.Commit)
		installed := false
		for _, p := range list {
			sub := ours.FindStringSubmatch(p.ID)
			if sub == nil || sub[1] != lp.ID || !forCheckout(p, r.Root) {
				continue
			}
			sameCommit := strings.HasPrefix(p.Version, version)
			if p.ID == want && sameCommit {
				installed = true
				continue
			}
			if !p.Enabled || r.named(p.ID) {
				continue
			}
			at := p.Version
			if at == "" {
				at = "an unknown commit"
			}
			r.find("plugin", LocalSettingsFile, "Claude Code loads the plugin "+p.ID+" at "+at+" in this checkout, but the lock holds "+
				lp.ID+" at "+version, "run bonsai update: it turns on the locked commit's plugin in "+LocalSettingsFile+
				" and the other one off")
		}
		if !installed {
			r.Warnings = append(r.Warnings, Finding{Code: "plugin", File: LocalSettingsFile,
				Message: "Claude Code reports the plugin " + want + " (" + lp.ID + " at " + version + ") not installed for this checkout on this machine",
				Next: "run bonsai update: it installs it once Claude Code knows this project's marketplace, which a Claude Code session here " +
					"registers (accept its trust question if it asks); sessions here fetch it on their own meanwhile, as " + LocalSettingsFile +
					" turns it on"})
		}
	}
}

// forCheckout reports whether an installed plugin applies to the checkout at root: a user-scope install applies
// everywhere; a project or local one only to the folder it was installed for.
func forCheckout(p InstalledPlugin, root string) bool {
	if p.Scope == "user" || p.Scope == "managed" || p.ProjectPath == "" {
		return true
	}
	return samePath(p.ProjectPath, root)
}

// samePath compares two folders as this system does: cleaned, links resolved where they exist, and on Windows
// without regard to case or slash direction.
func samePath(a, b string) bool {
	norm := func(p string) string {
		p = filepath.Clean(filepath.FromSlash(p))
		if r, err := filepath.EvalSymlinks(p); err == nil {
			p = r
		}
		if abs, err := filepath.Abs(p); err == nil {
			p = abs
		}
		return p
	}
	x, y := norm(a), norm(b)
	if runtime.GOOS == "windows" {
		return strings.EqualFold(x, y)
	}
	return x == y
}

// checkLocalPlugins reads this checkout's .claude/settings.local.json, offline: a line of Bonsai's turning on a
// locked pack's plugin from one of the workspace's marketplaces other than the lock's is drift, as sessions here load
// that commit's plugin when Claude Code has it (a finding). A missing file or line is no finding: a fresh clone or
// worktree has none until bonsai update, and CI never has one.
func (r *CheckResult) checkLocalPlugins() {
	if r.Config == nil || r.Lock == nil || len(r.Lock.Packs) == 0 {
		return
	}
	raw, exists, err := readFile(r.Root, LocalSettingsFile)
	if err != nil || !exists {
		return
	}
	v, err := schema.Decode(bytes.TrimPrefix(raw, []byte("\xef\xbb\xbf")))
	o, ok := v.(schema.Object)
	if err != nil || !ok {
		r.Warnings = append(r.Warnings, Finding{Code: "plugin", File: LocalSettingsFile,
			Message: LocalSettingsFile + " is not a JSON object Bonsai reads, so the plugins it turns on were not checked",
			Next:    "fix it (Claude Code reads it too), or delete it and run bonsai update"})
		return
	}
	ep, _ := o.Get("enabledPlugins")
	enabled, _ := ep.(schema.Object)
	market := lockMarket(r.Config, r.Lock)
	locked := map[string]string{}
	for _, lp := range r.Lock.Packs {
		locked[lp.ID] = lp.Commit
	}
	ours := bonsaiPlugin(r.Config.Name)
	for _, m := range enabled {
		sub := ours.FindStringSubmatch(m.Key)
		if sub == nil || m.Value != true || m.Key == PluginID(sub[1], market) {
			continue
		}
		if commit, ok := locked[sub[1]]; ok {
			r.find("plugin", LocalSettingsFile, LocalSettingsFile+" turns on "+m.Key+", from other commits than the lock's "+
				PluginID(sub[1], market)+" ("+sub[1]+" at "+pluginVersion(commit)+"), so sessions here may load another commit of "+sub[1],
				"run bonsai update: it turns the locked commit's plugin on there in its place")
		}
	}
}

// named reports whether a finding already names a plugin.
func (r *CheckResult) named(plugin string) bool {
	for _, f := range r.Findings {
		if f.Code == "plugin" && strings.Contains(f.Message, plugin) {
			return true
		}
	}
	return false
}
