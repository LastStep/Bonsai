package nonint

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/LastStep/Bonsai/internal/catalog"
	"github.com/LastStep/Bonsai/internal/config"
	"github.com/LastStep/Bonsai/internal/generate"
)

// ItemKind describes a removable ability category. It mirrors the cmd-side
// `itemType` descriptor but lives here so the headless core needs no cmd
// import. singular is the machine name ("skill"), dir is the agent/ subtree
// directory ("Skills"), ext is the generated-file extension (".md"/".sh").
type ItemKind struct {
	Singular string
	Dir      string
	Ext      string
}

// The five removable ability kinds. The cmd layer maps its own descriptors
// onto these; tests and the future MCP adapter use them directly.
var (
	KindSkill    = ItemKind{"skill", "Skills", ".md"}
	KindWorkflow = ItemKind{"workflow", "Workflows", ".md"}
	KindProtocol = ItemKind{"protocol", "Protocols", ".md"}
	KindSensor   = ItemKind{"sensor", "Sensors", ".sh"}
	KindRoutine  = ItemKind{"routine", "Routines", ".md"}
)

// KindFor resolves a kind by its singular machine name. Returns the kind and
// true on a match, zero-value + false otherwise. Used by the cmd adapter to
// translate its descriptor and by callers that only have the type string.
func KindFor(singular string) (ItemKind, bool) {
	switch singular {
	case "skill":
		return KindSkill, true
	case "workflow":
		return KindWorkflow, true
	case "protocol":
		return KindProtocol, true
	case "sensor":
		return KindSensor, true
	case "routine":
		return KindRoutine, true
	}
	return ItemKind{}, false
}

// RunRemoveAgent removes an installed agent and every ability under its
// workspace from an already-initialised project. It is the pure headless core
// behind `bonsai remove <agent> --yes`: typed options in, structured *Result
// out. It performs NO output — the CLI adapter serialises the Result to JSONL
// on stdout and prints Result.Warnings to stderr; the future MCP adapter
// (Plan 42) consumes the same Result.
//
// Business logic lifted from the cinematic removeflow closure (the lock
// untrack-by-workspace-prefix, .bonsai.yaml + settings.json + catalog-snapshot
// regeneration) plus the post-harness --delete-files cleanup. The cinematic
// path now calls this same core for its mutation; the TTY only adds Observe /
// Confirm / Yield chrome around it.
//
// Safety (Plan 41 Phase 3 Security): --yes bypasses the human confirmation
// gate, never validation. An empty agentName or a literal "*" is rejected
// (exit 2, zero mutation). When deleteFiles is set, EACH of the three delete
// targets — agentDir (os.RemoveAll), CLAUDE.md (os.Remove), .claude/
// (os.RemoveAll) — is Lstat'd and the whole operation refuses if ANY is a
// symlink (exit 2, zero deletion). This is a leaf-only mitigation: a symlinked
// parent component still escapes (Backlog P2). It matters here because the
// human confirm gate is gone.
//
// Returns (*Result, exitCode, error):
//   - ExitOK              — success (Result carries the file outcomes + warnings).
//   - ExitInvalidConfig   — empty/"*" target, agent not installed, or removing
//     tech-lead while other agents still depend on it (message contains
//     "tech-lead"), or a symlinked delete target.
//   - ExitRuntime         — a config save or generator error.
//   - ExitWrongCWDForInit — no .bonsai.yaml at cwd.
//
// On the error/reject paths the Result is nil.
func RunRemoveAgent(cwd string, cfg *config.ProjectConfig, cat *catalog.Catalog, lock *config.LockFile, version, agentName string, deleteFiles bool) (*Result, int, error) {
	// --yes bypasses confirmation, NEVER validation. Reject the dangerous
	// no-target and wildcard inputs before any filesystem mutation.
	if reason, bad := rejectUnsafeTarget(agentName); bad {
		return nil, ExitInvalidConfig, fmt.Errorf("remove: %s", reason)
	}

	if cfg == nil || cfg.Agents == nil {
		return nil, ExitInvalidConfig, fmt.Errorf("remove: no agents in project config")
	}
	agent, exists := cfg.Agents[agentName]
	if !exists || agent == nil {
		return nil, ExitInvalidConfig, fmt.Errorf("remove: agent %q is not installed (run `bonsai list`)", agentName)
	}

	// Preserve the tech-lead-in-use guard: tech-lead anchors every other
	// agent's peer-awareness + path-scoped rules, so it can only be removed
	// once it is the sole agent. Message MUST contain "tech-lead".
	if agentName == techLeadType && len(cfg.Agents) > 1 {
		return nil, ExitInvalidConfig, fmt.Errorf("remove: other agents depend on tech-lead — remove them first")
	}

	if lock == nil {
		lock = config.NewLockFile()
	}

	// Pre-flight the delete targets BEFORE any mutation so a symlinked target
	// aborts the whole operation with zero side effects. (The cinematic path
	// could rely on the human confirm gate; the headless path cannot.)
	var deleteTargets []deleteTarget
	if deleteFiles {
		deleteTargets = agentDeleteTargets(cwd, agent.Workspace)
		if t, isLink, err := firstSymlink(deleteTargets); err != nil {
			return nil, ExitRuntime, fmt.Errorf("remove: inspect delete target %s: %w", t, err)
		} else if isLink {
			return nil, ExitInvalidConfig, fmt.Errorf("remove: refusing --delete-files: %s is a symlink (would escape the workspace)", t)
		}
	}

	res := &Result{Write: &generate.WriteResult{}}

	// Untrack every locked file under the agent's workspace prefix, drop the
	// agent from config, then regenerate the shared settings.json + catalog
	// snapshot so the removed agent's sensors/rules disappear from them.
	wsPrefix := agent.Workspace
	for relPath := range lock.Files {
		if strings.HasPrefix(relPath, wsPrefix) {
			lock.Untrack(relPath)
		}
	}
	delete(cfg.Agents, agentName)

	var errs []error
	errs = append(errs, cfg.Save(filepath.Join(cwd, configFileName)))
	errs = append(errs, generate.SettingsJSON(cwd, cfg, cat, lock, res.Write, false))
	errs = append(errs, generate.WriteCatalogSnapshot(cwd, version, cat, res.Write))
	if joined := errors.Join(errs...); joined != nil {
		return nil, ExitRuntime, fmt.Errorf("remove: %w", joined)
	}

	// Lock-save failure is non-fatal (mirrors init/add). Warning rides in
	// Result.Warnings → stderr, never the JSONL stream.
	if err := lock.Save(cwd); err != nil {
		res.Warnings = append(res.Warnings, "could not save lock file: "+err.Error())
	}

	// --delete-files cleanup. Targets were already Lstat-vetted above. ENOENT
	// is swallowed (a file may legitimately already be gone). A non-ENOENT
	// error rides in Warnings — the registration was already removed.
	if deleteFiles {
		for _, t := range deleteTargets {
			var err error
			if t.recursive {
				err = os.RemoveAll(t.path)
			} else {
				err = os.Remove(t.path)
			}
			if err != nil && !errors.Is(err, os.ErrNotExist) {
				res.Warnings = append(res.Warnings, "could not delete "+t.path+": "+err.Error())
			}
		}
	}

	return res, ExitOK, nil
}

// RunRemoveItem removes a single ability (skill/workflow/protocol/sensor/
// routine) from one or all owning agents. It is the pure headless core behind
// `bonsai remove <type> <name> --yes [--from <agent>]`.
//
// Target resolution is re-implemented OUTSIDE the harness (the cinematic path
// resolves it via the SelectStage picker): it computes the owning agents
// directly from cfg, applies the required-item filter over ALL matches first,
// then disambiguates:
//   - fromAgent == "" and exactly one agent owns the item → that agent.
//   - fromAgent == "" and >1 agents own it → ExitInvalidConfig (2) with a
//     message NAMING the owners (the caller must pass --from).
//   - fromAgent set → scoped to that agent only; ExitInvalidConfig (2) if that
//     agent does not own the item (or had it filtered out as required).
//
// The required-item filter (filterRequired) runs on the --from branch too:
// --from is NOT an escape hatch around required-protection. Removing a
// required item via --from → exit 2, zero mutation.
//
// Safety: --yes bypasses confirmation, never validation. An empty itemName or
// a literal "*" is rejected (exit 2, zero mutation). The routine-check sensor
// is auto-managed (added/removed automatically when routines change) and
// cannot be removed directly (exit 2).
//
// Returns (*Result, exitCode, error):
//   - ExitOK              — success.
//   - ExitInvalidConfig   — empty/"*" target, routine-check removal, item not
//     installed (or not owned by --from), required item, or multi-owner item
//     with no --from.
//   - ExitRuntime         — a config save or generator error.
//   - ExitWrongCWDForInit — no .bonsai.yaml at cwd.
//
// On the error/reject paths the Result is nil.
func RunRemoveItem(cwd string, cfg *config.ProjectConfig, cat *catalog.Catalog, lock *config.LockFile, version, itemType, itemName, fromAgent string) (*Result, int, error) {
	kind, ok := KindFor(itemType)
	if !ok {
		return nil, ExitInvalidConfig, fmt.Errorf("remove: unknown item type %q", itemType)
	}

	// --yes bypasses confirmation, NEVER validation.
	if reason, bad := rejectUnsafeTarget(itemName); bad {
		return nil, ExitInvalidConfig, fmt.Errorf("remove: %s", reason)
	}

	// Block the auto-managed routine-check sensor (added/removed automatically
	// when routines change — Plan 31). Preserved from the cinematic path.
	if kind.Singular == "sensor" && itemName == "routine-check" {
		return nil, ExitInvalidConfig, fmt.Errorf("remove: routine-check is auto-managed (added and removed automatically when routines change)")
	}

	if cfg == nil || cfg.Agents == nil {
		return nil, ExitInvalidConfig, fmt.Errorf("remove: no agents in project config")
	}

	// Find every agent that owns the item, in stable (sorted) order.
	matches := ownersOf(cfg, kind, itemName)
	if len(matches) == 0 {
		return nil, ExitInvalidConfig, fmt.Errorf("remove: %s %q is not installed in any agent (run `bonsai list`)", kind.Singular, itemName)
	}

	// Required-item filter over ALL matches FIRST — this runs on the --from
	// branch too (see filterRequiredItem). If every owner has the item as
	// required, abort with zero mutation.
	allowed, requiredSkips := filterRequiredItem(matches, cat, kind, itemName)
	if len(allowed) == 0 {
		return nil, ExitInvalidConfig, fmt.Errorf("remove: %s %q is required by all agents that have it", kind.Singular, itemName)
	}

	// Disambiguate the removal target(s).
	var targets []agentItemMatch
	if fromAgent != "" {
		// Scope to the named owner. It must be in the allowed set — if the
		// agent had the item filtered out as required, --from must not bypass
		// that (the required filter already ran over all matches).
		var found bool
		for _, m := range allowed {
			if m.name == fromAgent {
				targets = []agentItemMatch{m}
				found = true
				break
			}
		}
		if !found {
			// Distinguish "required for that agent" from "not owned at all" for
			// a clearer message, but both are exit 2 zero-mutation.
			for _, name := range requiredSkips {
				if name == fromAgent {
					return nil, ExitInvalidConfig, fmt.Errorf("remove: %s %q is required for agent %q — cannot remove", kind.Singular, itemName, fromAgent)
				}
			}
			return nil, ExitInvalidConfig, fmt.Errorf("remove: agent %q does not have %s %q installed", fromAgent, kind.Singular, itemName)
		}
	} else if len(allowed) > 1 {
		// Multi-owner with no --from: refuse and NAME the owners so the caller
		// can re-run with --from <one>.
		names := make([]string, 0, len(allowed))
		for _, m := range allowed {
			names = append(names, m.name)
		}
		return nil, ExitInvalidConfig, fmt.Errorf("remove: %s %q is installed in multiple agents (%s) — pass --from <agent> to choose one", kind.Singular, itemName, strings.Join(names, ", "))
	} else {
		targets = allowed
	}

	if lock == nil {
		lock = config.NewLockFile()
	}

	res := &Result{Write: &generate.WriteResult{}}

	if err := removeItemFromTargets(cwd, cfg, cat, lock, res.Write, version, kind, itemName, targets); err != nil {
		return nil, ExitRuntime, fmt.Errorf("remove: %w", err)
	}

	if err := lock.Save(cwd); err != nil {
		res.Warnings = append(res.Warnings, "could not save lock file: "+err.Error())
	}

	return res, ExitOK, nil
}

// agentItemMatch pairs an agent's machine name with its installed record.
type agentItemMatch struct {
	name  string
	agent *config.InstalledAgent
}

// ownersOf returns every agent that has the named item installed for the given
// kind, sorted by agent name for deterministic output.
func ownersOf(cfg *config.ProjectConfig, kind ItemKind, itemName string) []agentItemMatch {
	names := make([]string, 0, len(cfg.Agents))
	for name := range cfg.Agents {
		names = append(names, name)
	}
	sort.Strings(names)

	var matches []agentItemMatch
	for _, name := range names {
		agent := cfg.Agents[name]
		if agent == nil {
			continue
		}
		if itemInSlice(itemList(agent, kind), itemName) {
			matches = append(matches, agentItemMatch{name, agent})
		}
	}
	return matches
}

// filterRequiredItem splits matches into the subset where the item is NOT
// required for the agent (allowed) and the names of agents that have it as
// required (requiredSkips). Mirrors the cmd-side filterRequired so the
// required-protection is identical in headless and cinematic paths.
func filterRequiredItem(matches []agentItemMatch, cat *catalog.Catalog, kind ItemKind, itemName string) (allowed []agentItemMatch, requiredSkips []string) {
	for _, m := range matches {
		if itemIsRequiredFor(cat, kind, itemName, m.agent.AgentType) {
			requiredSkips = append(requiredSkips, m.name)
		} else {
			allowed = append(allowed, m)
		}
	}
	return allowed, requiredSkips
}

// removeItemFromTargets performs the actual mutation for each target agent:
// drop the item from the config list, untrack + delete the generated file and
// any companion trigger files, handle the routine auto-sensor + dashboard, and
// regenerate the workspace CLAUDE.md. Finally persists config + settings.json
// + catalog snapshot. os.Remove ENOENT is swallowed (the file may already be
// gone); generator errors are aggregated via errors.Join.
func removeItemFromTargets(cwd string, cfg *config.ProjectConfig, cat *catalog.Catalog, lock *config.LockFile, wr *generate.WriteResult, version string, kind ItemKind, itemName string, targets []agentItemMatch) error {
	var errs []error
	for _, t := range targets {
		removeFromList(t.agent, kind, itemName)

		// Untrack + delete the generated ability file.
		relPath := filepath.Join(t.agent.Workspace, "agent", kind.Dir, itemName+kind.Ext)
		lock.Untrack(relPath)
		_ = os.Remove(filepath.Join(cwd, relPath))

		// Clean up generated trigger files.
		switch kind.Singular {
		case "skill":
			rulePath := filepath.Join(t.agent.Workspace, ".claude", "rules", "skill-"+itemName+".md")
			lock.Untrack(rulePath)
			_ = os.Remove(filepath.Join(cwd, rulePath))
		case "workflow":
			skillDir := filepath.Join(t.agent.Workspace, ".claude", "skills", itemName)
			skillPath := filepath.Join(skillDir, "SKILL.md")
			lock.Untrack(skillPath)
			_ = os.Remove(filepath.Join(cwd, skillPath))
			_ = os.Remove(filepath.Join(cwd, skillDir)) // remove now-empty dir
		case "routine":
			// Routine removal re-syncs the auto-managed routine-check sensor and
			// the routines dashboard. Preserved from the cinematic path.
			generate.EnsureRoutineCheckSensor(t.agent)
			workspaceRoot := filepath.Join(cwd, t.agent.Workspace)
			if len(t.agent.Routines) > 0 {
				errs = append(errs, generate.RoutineDashboard(cwd, workspaceRoot, t.agent, cat, lock, wr, false))
			} else {
				dashPath := filepath.Join(t.agent.Workspace, "agent", "Core", "routines.md")
				lock.Untrack(dashPath)
				_ = os.Remove(filepath.Join(cwd, dashPath))
			}
		}

		// Regenerate the workspace CLAUDE.md so the removed item disappears
		// from its navigation table. Look up by the agent's machine name (map
		// key) — matches the cinematic runRemoveItemAction exactly.
		if agentDef := cat.GetAgent(t.name); agentDef != nil {
			workspaceRoot := filepath.Join(cwd, t.agent.Workspace)
			errs = append(errs, generate.WorkspaceClaudeMD(cwd, workspaceRoot, agentDef, t.agent, cfg, cat, lock, wr, false))
		}
	}

	errs = append(errs, cfg.Save(filepath.Join(cwd, configFileName)))
	errs = append(errs, generate.SettingsJSON(cwd, cfg, cat, lock, wr, false))
	errs = append(errs, generate.WriteCatalogSnapshot(cwd, version, cat, wr))
	return errors.Join(errs...)
}

// ─── small shared helpers ───────────────────────────────────────────────

// configFileName is the project config filename. Duplicated from cmd's
// `configFile` const so the core needs no cmd import.
const configFileName = ".bonsai.yaml"

// rejectUnsafeTarget enforces the --yes safety contract: --yes bypasses the
// human confirmation gate, never validation. An empty or whitespace-only
// target and a literal "*" wildcard are rejected before any filesystem
// mutation. Returns (reason, true) when the target is unsafe.
func rejectUnsafeTarget(name string) (string, bool) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return "empty target name (--yes bypasses confirmation, not validation)", true
	}
	if trimmed == "*" {
		return "wildcard target \"*\" is not allowed (name an explicit target)", true
	}
	return "", false
}

// deleteTarget is one path the --delete-files cleanup operates on, plus
// whether it is removed recursively (a directory).
type deleteTarget struct {
	path      string
	recursive bool
}

// agentDeleteTargets returns the three --delete-files targets for an agent
// workspace, in the same order the cinematic path deleted them: agentDir
// (recursive), CLAUDE.md (single file), .claude/ (recursive).
func agentDeleteTargets(cwd, workspace string) []deleteTarget {
	return []deleteTarget{
		{filepath.Join(cwd, workspace, "agent"), true},
		{filepath.Join(cwd, workspace, "CLAUDE.md"), false},
		{filepath.Join(cwd, workspace, ".claude"), true},
	}
}

// firstSymlink Lstats each delete target and returns the first one that is a
// symlink. A non-existent target is fine (nothing to delete). Returns
// (path, isSymlink, statErr) — statErr non-nil only on a genuine stat error
// (not ErrNotExist). This is the leaf-only symlink mitigation: it inspects the
// final component, not parent directories.
func firstSymlink(targets []deleteTarget) (string, bool, error) {
	for _, t := range targets {
		info, err := os.Lstat(t.path)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return t.path, false, err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return t.path, true, nil
		}
	}
	return "", false, nil
}

// itemList returns the agent's installed-item slice for the given kind.
func itemList(agent *config.InstalledAgent, kind ItemKind) []string {
	switch kind.Singular {
	case "skill":
		return agent.Skills
	case "workflow":
		return agent.Workflows
	case "protocol":
		return agent.Protocols
	case "sensor":
		return agent.Sensors
	case "routine":
		return agent.Routines
	}
	return nil
}

// removeFromList drops every occurrence of itemName from the agent's slice for
// the given kind, in place.
func removeFromList(agent *config.InstalledAgent, kind ItemKind, itemName string) {
	filter := func(list []string) []string {
		var out []string
		for _, item := range list {
			if item != itemName {
				out = append(out, item)
			}
		}
		return out
	}
	switch kind.Singular {
	case "skill":
		agent.Skills = filter(agent.Skills)
	case "workflow":
		agent.Workflows = filter(agent.Workflows)
	case "protocol":
		agent.Protocols = filter(agent.Protocols)
	case "sensor":
		agent.Sensors = filter(agent.Sensors)
	case "routine":
		agent.Routines = filter(agent.Routines)
	}
}

// itemInSlice reports whether name is present in list.
func itemInSlice(list []string, name string) bool {
	for _, item := range list {
		if item == name {
			return true
		}
	}
	return false
}

// itemIsRequiredFor reports whether the named item is marked required for the
// given agent type in the catalog. Unknown items / kinds are not required.
func itemIsRequiredFor(cat *catalog.Catalog, kind ItemKind, name, agentType string) bool {
	switch kind.Singular {
	case "skill":
		if item := cat.GetSkill(name); item != nil {
			return item.Required.CompatibleWith(agentType)
		}
	case "workflow":
		if item := cat.GetWorkflow(name); item != nil {
			return item.Required.CompatibleWith(agentType)
		}
	case "protocol":
		if item := cat.GetProtocol(name); item != nil {
			return item.Required.CompatibleWith(agentType)
		}
	case "sensor":
		if item := cat.GetSensor(name); item != nil {
			return item.Required.CompatibleWith(agentType)
		}
	case "routine":
		if item := cat.GetRoutine(name); item != nil {
			return item.Required.CompatibleWith(agentType)
		}
	}
	return false
}
