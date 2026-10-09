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
// (InstallPlugins), and only a project's own packs' plugins: PluginCLI can list plugins and marketplaces, install and
// uninstall, and nothing else, and Bonsai never installs, removes or changes a plugin that is not one of the project's
// own packs (Rohan, 9 Oct: "it shouldn't install or remove other plugins").
//
// Taking a pack out (step 5.1.7: unlink, or update with a pack gone from bonsai.yaml) removes Claude Code's record of
// the pack's install for this checkout (UninstallPlugins): `claude plugin uninstall <id> --scope project` in the
// checkout, for each project-scope install Claude Code lists for this checkout of that pack from one of the
// workspace's marketplaces (the lock's, and any older one an update left). As measured on Claude Code 2.1.295, the
// uninstall works before and after Bonsai's entries leave .claude/settings.json, the file deleted too; run before, it
// writes the file itself (an empty enabledPlugins), so Bonsai runs it after its own write, and Claude Code then
// leaves the file alone. A plugin not installed answers failureCode not_installed, which is no failure: nothing to
// remove. A worktree's record is its own (its own checkout's path): unlink in each checkout.
//
// First-time trust stays a person's (step 5.1.7): Bonsai never answers Claude Code's trust question and never
// registers a marketplace behind it. Until a Claude Code session in the trusted folder has registered the workspace's
// marketplace, `claude plugin install` answers not_found and `claude plugin marketplace list` lacks it: the install
// step reports waiting, check warns plugin-trust and status --full says so, each naming the person's step (open
// Claude Code in the checkout and accept its trust question) and then the command. A plugin that carries code parts (consent.go: hooks, MCP and LSP
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
	"sort"
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

// PluginCLI is the part of Claude Code's command line Bonsai uses. Each call runs in dir, a checkout's top folder;
// Install and Uninstall pass --scope project (PluginScope), never another scope.
type PluginCLI interface {
	List(dir string) ([]InstalledPlugin, error)
	Install(dir, plugin string) (InstallResult, error)
	Uninstall(dir, plugin string) (InstallResult, error)
	Marketplaces(dir string) ([]string, error) // the names of the marketplaces Claude Code has registered
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

// Uninstall runs `claude plugin uninstall <plugin> --scope project --json`: Claude Code's record of the plugin's
// install for this checkout goes. Its result line has install's shape (failureCode not_installed: nothing to remove).
func (c ClaudeCLI) Uninstall(dir, plugin string) (InstallResult, error) {
	out, msg, err := c.run(dir, listTimeout, "plugin", "uninstall", plugin, "--scope", PluginScope, "--json")
	if errors.Is(err, ErrNoClaude) {
		return InstallResult{}, err
	}
	if r, ok := ParseInstallResult(out); ok {
		return r, nil
	}
	if err != nil {
		return InstallResult{}, errors.New("claude plugin uninstall failed: " + msg)
	}
	return InstallResult{}, errors.New("claude plugin uninstall printed no result line")
}

// Marketplaces runs `claude plugin marketplace list --json`: the marketplaces Claude Code has registered on this
// machine, by name.
func (c ClaudeCLI) Marketplaces(dir string) ([]string, error) {
	out, msg, err := c.run(dir, listTimeout, "plugin", "marketplace", "list", "--json")
	if err != nil {
		if errors.Is(err, ErrNoClaude) {
			return nil, err
		}
		return nil, errors.New("claude plugin marketplace list --json failed: " + msg)
	}
	return ParseMarketplaces(out)
}

// ParseMarketplaces reads `claude plugin marketplace list --json`: a JSON list of objects with a name.
func ParseMarketplaces(out []byte) ([]string, error) {
	var list []struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(bytes.TrimSpace(out), &list); err != nil {
		return nil, errors.New("claude plugin marketplace list --json printed no list Bonsai reads: " + err.Error())
	}
	names := make([]string, 0, len(list))
	for _, m := range list {
		names = append(names, m.Name)
	}
	return names, nil
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

// LockMarket is lockMarket, for status --full.
func LockMarket(cfg *workspace.Config, lock *workspace.Lock) string { return lockMarket(cfg, lock) }

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

// PluginResult is what InstallPlugins or UninstallPlugins did for one pack's plugin.
type PluginResult struct {
	Pack    string
	Plugin  string // <pack>@<marketplace>
	Commit  string
	Result  string // installed, waiting, failed or skipped; uninstalled (UninstallPlugins)
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
				"in this checkout registers it once a person has trusted the folder (Bonsai never answers that question), " +
				"then update installs the plugin at " + pluginVersion(lp.Commit)
			r.Next, r.Who = TrustNext(len(consent.Code[lp.ID]) > 0), "person"
		default:
			r.Result, r.Message, r.Next, r.Who = "failed", strings.TrimSpace(res.Message), "run: "+cmd, "agent"
		}
		r.Message, r.Next = ascii(r.Message), ascii(r.Next)
		out = append(out, r)
	}
	return out
}

// TrustNext is the person's step while Claude Code has not registered a checkout's marketplace (first-time trust,
// spec section 5): open Claude Code in the checkout and accept its trust question, then the command that installs
// the plugin, with --allow-exec when it carries code (a person's consent on each machine).
func TrustNext(code bool) string {
	cmd := "bonsai update"
	if code {
		cmd += " --allow-exec"
	}
	return "a person opens Claude Code in this checkout and accepts its trust question (Bonsai never answers it), then " +
		"quits it; then run: " + cmd
}

// UninstallPlugins removes Claude Code's record of each pack's plugin install for the checkout at root (step 5.1.7:
// the packs unlink takes out, or update takes out of bonsai.yaml), at project scope: every project-scope install
// `claude plugin list` reports for this checkout of the pack from one of the workspace's marketplaces
// (bonsai-<name>-<8 hex>; any workspace name when name is ""), the lock's (market) and any older one an update left.
// Run it after the project's files are written, so Claude Code leaves .claude/settings.json alone. One result per
// install removed, or one per pack with none: uninstalled (also when Claude Code had no record: nothing to remove),
// failed or skipped, each failure naming the exact command; it never fails the command. A nil cli asks nothing.
func UninstallPlugins(root, name, market string, packs []workspace.LockedPack, cli PluginCLI) []PluginResult {
	if cli == nil || len(packs) == 0 {
		return nil
	}
	pattern := regexp.QuoteMeta(name)
	if name == "" {
		pattern = `[a-z][a-z0-9-]*`
	}
	ours := regexp.MustCompile(`^([a-z0-9][a-z0-9-]*)@bonsai-` + pattern + `-[0-9a-f]{8}$`)
	plugin := func(pack string) string {
		if market == "" {
			return pack + "@bonsai-" + name
		}
		return PluginID(pack, market)
	}
	command := func(id string) string { return "claude plugin uninstall " + id + " --scope " + PluginScope }
	list, err := cli.List(root)
	var out []PluginResult
	for _, lp := range packs {
		r := PluginResult{Pack: lp.ID, Plugin: plugin(lp.ID), Commit: lp.Commit}
		switch {
		case errors.Is(err, ErrNoClaude):
			r.Result, r.Message = "skipped", "Claude Code is not on the PATH, so its record of the plugin was not removed (a machine without Claude Code has none)"
			r.Next, r.Who = "if Claude Code is installed here off the PATH, run: "+command(r.Plugin), "person"
			out = append(out, asciiResult(r))
			continue
		case err != nil:
			r.Result, r.Message = "failed", "Claude Code's plugins could not be listed: "+err.Error()
			r.Next, r.Who = "run: "+command(r.Plugin), "agent"
			out = append(out, asciiResult(r))
			continue
		}
		var ids []string
		for _, p := range list {
			sub := ours.FindStringSubmatch(p.ID)
			if sub == nil || sub[1] != lp.ID || p.Scope != PluginScope || p.ProjectPath == "" || !samePath(p.ProjectPath, root) {
				continue
			}
			ids = append(ids, p.ID)
		}
		sort.Strings(ids)
		if len(ids) == 0 {
			r.Result, r.Message = "uninstalled", "not installed for this checkout: nothing to remove"
			out = append(out, asciiResult(r))
			continue
		}
		for i, id := range ids {
			if i > 0 && ids[i-1] == id {
				continue
			}
			r := PluginResult{Pack: lp.ID, Plugin: id, Commit: lp.Commit}
			res, err := cli.Uninstall(root, id)
			switch {
			case errors.Is(err, ErrNoClaude):
				r.Result, r.Message = "skipped", "Claude Code is not on the PATH, so its record of the plugin was not removed"
				r.Next, r.Who = "if Claude Code is installed here off the PATH, run: "+command(id), "person"
			case err != nil:
				r.Result, r.Message, r.Next, r.Who = "failed", err.Error(), "run: "+command(id), "agent"
			case res.Outcome == "ok":
				r.Result, r.Message = "uninstalled", "Claude Code's record of the install for this checkout removed"
			case res.FailureCode == "not_installed":
				r.Result, r.Message = "uninstalled", "not installed for this checkout: nothing to remove"
			default:
				r.Result, r.Message, r.Next, r.Who = "failed", strings.TrimSpace(res.Message), "run: "+command(id), "agent"
			}
			out = append(out, asciiResult(r))
		}
	}
	return out
}

// asciiResult keeps a result's text ASCII.
func asciiResult(r PluginResult) PluginResult {
	r.Message, r.Next = ascii(r.Message), ascii(r.Next)
	return r
}

// registered asks Claude Code once whether it has registered a marketplace: true or false, and false with ok false
// when it could not be asked.
func registered(root, market string, cli PluginCLI) (yes, ok bool) {
	names, err := cli.Marketplaces(root)
	if err != nil {
		return false, false
	}
	for _, n := range names {
		if n == market {
			return true, true
		}
	}
	return false, true
}

// MarketplaceRegistered is registered for status --full: whether Claude Code has registered the lock's marketplace
// (first-time trust, spec section 5), and whether it could be asked.
func MarketplaceRegistered(r *CheckResult, cli PluginCLI) (yes, ok bool) {
	if cli == nil || r == nil || r.Config == nil || r.Lock == nil || len(r.Lock.Packs) == 0 {
		return false, false
	}
	return registered(r.Root, lockMarket(r.Config, r.Lock), cli)
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
	askedMarket, known, askedOK := false, false, false
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
			if !askedMarket {
				askedMarket = true
				known, askedOK = registered(r.Root, market, cli)
			}
			if askedOK && !known {
				r.add("plugin-trust", SettingsFile, "", "Claude Code has not registered this checkout's marketplace "+market+" yet, so the plugin "+
					want+" ("+lp.ID+" at "+version+") cannot be installed: a Claude Code session in this checkout registers it once a "+
					"person has trusted the folder", "a person opens Claude Code in this checkout and accepts its trust question (Bonsai "+
					"never answers it), then quits it; a plugin that carries code also needs --allow-exec, a person's consent, which "+
					"update names; then run: bonsai update")
				continue
			}
			r.add("plugin-missing", SettingsFile, "", "Claude Code reports the plugin "+want+" ("+lp.ID+" at "+version+") not installed for this checkout",
				"a plugin that carries code needs --allow-exec, a person's consent, which update names; to install it, run: bonsai update --yes")
		}
	}
}

// PluginsInstalled asks Claude Code which of the lock's plugins it has installed for this checkout at their locked
// commits (status --full: a pack not installed on this machine is a needs entry of kind plugin, spec §5): by pack id.
// A nil cli, or a list that fails, gives the error and no answer.
func PluginsInstalled(r *CheckResult, cli PluginCLI) (map[string]bool, error) {
	if r == nil || r.Config == nil || r.Lock == nil {
		return map[string]bool{}, nil
	}
	if cli == nil {
		return nil, ErrNoClaude
	}
	list, err := cli.List(r.Root)
	if err != nil {
		return nil, err
	}
	market := lockMarket(r.Config, r.Lock)
	out := map[string]bool{}
	for _, lp := range r.Lock.Packs {
		want := PluginID(lp.ID, market)
		for _, p := range list {
			if p.ID == want && strings.HasPrefix(p.Version, pluginVersion(lp.Commit)) && forCheckout(p, r.Root) {
				out[lp.ID] = true
			}
		}
		if !out[lp.ID] {
			out[lp.ID] = false
		}
	}
	return out, nil
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
