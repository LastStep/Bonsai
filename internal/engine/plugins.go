package engine

// This machine's plugins (spec §5, "This machine's install follows the lock" and "Drift is reported, never silent";
// plan part 4b). A pack reaches a session as a Claude Code plugin from the workspace's own inline marketplace, which
// settings.go writes into the checkout's .claude/settings.json: bonsai-<workspace name>-<8 hex>, each plugin pinned to
// its locked commit by sha, and turned on there. What Claude Code does with it, as part 4b found it (2.1.294):
//
//   - A marketplace that only a project's settings declare is registered by a Claude Code session in that folder,
//     and only once the folder is trusted (an interactive session asks; -p and --bg never do: in an untrusted folder
//     the entry is ignored, and --bg exits "Workspace not trusted"). Neither `claude plugin install` nor `claude
//     plugin marketplace update` registers it, so right after init, or an update that moved a commit (a new
//     marketplace name), the install fails "not found" until such a session has started in the checkout.
//   - That session fetches no plugin that only the project's settings turn on; `claude plugin install` does.
//   - Scope local is not per checkout: in a git worktree, Claude Code reads the main checkout's
//     .claude/settings.local.json as well as the worktree's own, a line there turning a plugin on cannot be turned off
//     from the worktree, and `claude plugin install --scope local` run in a worktree writes the main checkout's file.
//     A pack's plugin turned on there would load in every worktree beside the worktree's own commit: the two fight.
//   - Scope project is per checkout: the plugin line is the checkout's own committed .claude/settings.json, which
//     Bonsai writes already. `claude plugin install --scope project` records the install for the checkout, and the
//     first time rewrites that file in Claude Code's own key order (no value changes; claim reads Bonsai's lines in
//     any order).
//
// So after init and update write, Bonsai asks Claude Code to install each locked plugin at project scope, never user
// (InstallPlugins), and only a project's own packs' plugins: PluginCLI can list and install, and nothing else, so
// Bonsai never installs, removes or changes a plugin that is not one of the project's own packs (Rohan, 9 Oct: "it
// shouldn't install or remove other plugins"). A plugin that carries code parts (consent.go: hooks, MCP and LSP
// servers, monitors, mods) runs code on this machine once installed, so it is installed only with --allow-exec, on
// each machine (Rohan, 9 Oct, S2: "Ask on each machine"): without it the result is waiting, naming the plugin and its
// code parts, with the same command and --allow-exec as the person's next step; a plugin Claude Code already reports
// installed at the locked commit for the checkout is not asked about again. Like every result of this step, waiting
// never changes the command's exit code. A plugin with no code part (roles, skills, commands) installs with no flag; check compares what Claude Code reports with the lock (ComparePlugins), and reads, offline, the
// .claude/settings.local.json files Claude Code reads for the checkout for a line turning on another commit's plugin
// (checkLocalPlugins). Bonsai never writes a local settings file. The calls go through PluginCLI: ClaudeCLI runs the
// real `claude`, tests use a fake; a nil PluginCLI asks nothing (Bonsai's tests).

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

// LocalSettingsFile is Claude Code's settings file outside git, .claude/settings.local.json. Bonsai never writes it;
// check reads it for drift.
const LocalSettingsFile = ".claude/settings.local.json"

// PluginScope is the scope Bonsai installs plugins at: project, this checkout's .claude/settings.json. Never user,
// which writes the person's own ~/.claude/settings.json (plan, "Claude Code's own files"); never local, which every
// worktree of the repository shares.
const PluginScope = "project"

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

// Install runs `claude plugin install <plugin> --scope project --json`. A refusal Claude Code reports on its result
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

// lockMarket is the marketplace name a lock's packs give, in the lock's order (bonsai.yaml's): each pack's commit
// and the folder the lock records (bonsai.yaml's for a lock written before formats set 4).
func lockMarket(cfg *workspace.Config, lock *workspace.Lock) string {
	folders := map[string]string{}
	for _, r := range cfg.Packs {
		folders[r.ID] = r.Path
	}
	var pins []string
	for _, lp := range lock.Packs {
		pins = append(pins, Pin(lp.Commit, lockedFolder(lp, folders[lp.ID])))
	}
	return MarketplaceName(cfg.Name, pins)
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

// PluginResult is what InstallPlugins did for one pack.
type PluginResult struct {
	Pack    string
	Plugin  string // <pack>@<marketplace>
	Commit  string
	Result  string // installed, waiting, failed or skipped
	Message string
	Next    string
	Who     string // who takes Next: agent or person
}

// PluginConsent is what the install step needs to keep a plugin that carries code off this machine until a person
// consents (Rohan, 9 Oct: "Ask on each machine"): each locked pack's plugin code parts at its locked commit (by pack
// id; Plan.PluginConsent fills them), whether --allow-exec was given, and the command that repeats the run with it.
type PluginConsent struct {
	Code      map[string][]CodePart
	AllowExec bool
	Again     string
}

// InstallPlugins asks Claude Code to install each locked pack's plugin for the checkout at root, at project scope
// (spec §5: `claude plugin install` is a no-op once installed, and the marketplace name is new whenever a locked
// commit or a pack's folder changed, so "already installed" below, which matches the plugin's id and commit, compares
// the folder too). A plugin that carries code parts is installed only with consent.AllowExec; without it, one Claude
// Code already reports installed at the locked commit for the checkout is left as it is, and any other waits for the
// person's --allow-exec. It never fails the command: the project's files are written; what this machine still needs
// is in each result's next step. A nil cli installs nothing.
func InstallPlugins(root string, cfg *workspace.Config, lock *workspace.Lock, cli PluginCLI, consent PluginConsent) []PluginResult {
	if cli == nil || cfg == nil || lock == nil || len(lock.Packs) == 0 {
		return nil
	}
	market := lockMarket(cfg, lock)
	var listed []InstalledPlugin
	listedOnce := false
	installedHere := func(id, commit string) bool {
		if !listedOnce {
			listedOnce = true
			listed, _ = cli.List(root)
		}
		for _, p := range listed {
			if p.ID == id && strings.HasPrefix(p.Version, pluginVersion(commit)) && forCheckout(p, root) {
				return true
			}
		}
		return false
	}
	var out []PluginResult
	for _, lp := range lock.Packs {
		id := PluginID(lp.ID, market)
		cmd := "claude plugin install " + id + " --scope " + PluginScope
		r := PluginResult{Pack: lp.ID, Plugin: id, Commit: lp.Commit}
		if parts := consent.Code[lp.ID]; len(parts) > 0 && !consent.AllowExec {
			if installedHere(id, lp.Commit) {
				r.Result, r.Message = "installed", "at "+pluginVersion(lp.Commit)+" (already installed for this checkout)"
			} else {
				again := consent.Again
				if again == "" {
					again = "bonsai update --allow-exec"
				}
				r.Result = "waiting"
				r.Message = "the plugin carries code Claude Code runs on its own (" + codeParts(parts) + "), so Bonsai installs it " +
					"on this machine only with --allow-exec"
				r.Next, r.Who = "if you consent to that code running on this machine, run: "+again, "person"
			}
			r.Message, r.Next = ascii(r.Message), ascii(r.Next)
			out = append(out, r)
			continue
		}
		before, _, _ := readFile(root, SettingsFile)
		res, err := cli.Install(root, id)
		switch {
		case errors.Is(err, ErrNoClaude):
			r.Result, r.Message = "skipped", "Claude Code is not on the PATH, so the plugin was not installed on this machine"
			r.Next, r.Who = "where sessions run, install Claude Code, then run bonsai update again", "person"
		case err != nil:
			r.Result, r.Message, r.Next, r.Who = "failed", err.Error(), "run: "+cmd, "agent"
		case res.Outcome == "ok":
			r.Result, r.Message = "installed", "at "+pluginVersion(lp.Commit)
			if after, _, _ := readFile(root, SettingsFile); !bytes.Equal(before, after) {
				r.Message += "; Claude Code wrote " + SettingsFile + " again in its own key order (Bonsai's lines are unchanged)"
			}
		case res.FailureCode == "not_found":
			r.Result = "waiting"
			r.Message = "Claude Code has not registered this checkout's marketplace " + market + " yet: a Claude Code session " +
				"here does that, once the folder is trusted"
			r.Next = "open Claude Code in this checkout (accept its trust question if it asks), leave it, then run bonsai update " +
				"again: it installs the plugin at " + pluginVersion(lp.Commit) + ", and sessions after that load it"
			r.Who = "person"
		default:
			r.Result, r.Message, r.Next, r.Who = "failed", strings.TrimSpace(res.Message), "run: "+cmd, "agent"
		}
		r.Message, r.Next = ascii(r.Message), ascii(r.Next)
		out = append(out, r)
	}
	return out
}

// ComparePlugins adds check's plugin findings and warnings (spec §5, "Drift is reported, never silent"): it asks
// Claude Code which plugins it has for this checkout and compares their commits with the lock's. A plugin of a
// locked pack, turned on here from one of the workspace's marketplaces at another commit, is drift: sessions here
// load it beside or instead of the lock's (a finding). The lock's plugin not installed for this checkout is a
// warning: the project is right, and bonsai update installs it. A nil cli compares nothing.
func ComparePlugins(r *CheckResult, cli PluginCLI) {
	if cli == nil || r == nil || r.Config == nil || r.Lock == nil || len(r.Lock.Packs) == 0 {
		return
	}
	list, err := cli.List(r.Root)
	if err != nil {
		if errors.Is(err, ErrNoClaude) {
			r.add("plugin-unchecked", SettingsFile, "", "Claude Code is not on the PATH, so this machine's plugins were not compared with the lock",
				"where sessions run, a person installs Claude Code, then run: bonsai check; where none run (CI), nothing is needed")
			return
		}
		r.add("plugin-unchecked", SettingsFile, "agent", "this machine's plugins were not compared with the lock: "+err.Error(),
			"to see why, run: claude plugin list --json; then run: bonsai check")
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
			if sub == nil || sub[1] != lp.ID {
				continue
			}
			sameCommit := strings.HasPrefix(p.Version, version)
			if p.ID == want && sameCommit && forCheckout(p, r.Root) {
				installed = true
				continue
			}
			// Claude Code reports a plugin enabled when the settings it reads for this folder turn it on, whichever
			// checkout the install was recorded for (a worktree reads its main checkout's local settings too).
			if !p.Enabled || (p.ID == want && sameCommit) || r.named(p.ID) {
				continue
			}
			at := p.Version
			if at == "" {
				at = "an unknown commit"
			}
			r.add("plugin", SettingsFile, "", "Claude Code turns on the plugin "+p.ID+" at "+at+" in this checkout beside the lock's "+
				want+" ("+lp.ID+" at "+version+"), so sessions here may load another commit of "+lp.ID,
				"a person finds the setting that turns "+p.ID+" on (a .claude/settings.local.json here or in the main checkout, or "+
					"their own settings) and takes it out; for a local one, in the main checkout, run: claude plugin uninstall "+p.ID+" --scope local")
		}
		if !installed {
			r.add("plugin-missing", SettingsFile, "", "Claude Code reports the plugin "+want+" ("+lp.ID+" at "+version+") not installed for this checkout",
				"to install it, run: bonsai update --yes (when the plugin carries code, update names --allow-exec, a person's consent; "+
					"Claude Code first registers this checkout's marketplace in a session here, once a person accepts its trust question)")
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

// checkLocalPlugins reads, offline, the .claude/settings.local.json files Claude Code reads for this checkout: its
// own and, in a worktree, the main checkout's. A line there turning on a locked pack's plugin from one of the
// workspace's marketplaces at other commits than the lock's is drift: sessions here load that commit's plugin beside
// the lock's when Claude Code has it, and nothing in the worktree can turn it off (a finding). Bonsai writes no local
// settings file; such a line comes from a local-scope install. A missing file is no finding.
func (r *CheckResult) checkLocalPlugins() {
	if r.Config == nil || r.Lock == nil || len(r.Lock.Packs) == 0 {
		return
	}
	market := lockMarket(r.Config, r.Lock)
	locked := map[string]string{}
	for _, lp := range r.Lock.Packs {
		locked[lp.ID] = lp.Commit
	}
	ours := bonsaiPlugin(r.Config.Name)
	dirs := []string{r.Root}
	if r.Main != "" && !samePath(r.Main, r.Root) {
		dirs = append(dirs, r.Main)
	}
	for _, dir := range dirs {
		where := LocalSettingsFile
		if dir != r.Root {
			where = filepath.ToSlash(filepath.Join(dir, filepath.FromSlash(LocalSettingsFile))) + " (the main checkout's, which Claude Code reads in every worktree)"
		}
		raw, exists, err := readFile(dir, LocalSettingsFile)
		if err != nil || !exists {
			continue
		}
		v, err := schema.Decode(bytes.TrimPrefix(raw, []byte("\xef\xbb\xbf")))
		o, ok := v.(schema.Object)
		if err != nil || !ok {
			r.add("plugin-unchecked", LocalSettingsFile, "", where+" is not a JSON object Bonsai reads, so the plugins it turns on were not checked",
				"a person fixes the file (Claude Code reads it too), then run: bonsai check")
			continue
		}
		ep, _ := o.Get("enabledPlugins")
		enabled, _ := ep.(schema.Object)
		for _, m := range enabled {
			sub := ours.FindStringSubmatch(m.Key)
			if sub == nil || m.Value != true || m.Key == PluginID(sub[1], market) {
				continue
			}
			if commit, ok := locked[sub[1]]; ok {
				r.add("plugin", LocalSettingsFile, "", where+" turns on "+m.Key+", not the lock's "+PluginID(sub[1], market)+" ("+sub[1]+
					" at "+pluginVersion(commit)+"), so sessions here may load another commit of "+sub[1],
					"take the line out of that file (Bonsai installs plugins at project scope, in each checkout's own "+SettingsFile+
						"): in the main checkout, run: claude plugin uninstall "+m.Key+" --scope local")
			}
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
