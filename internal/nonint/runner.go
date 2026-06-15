package nonint

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"

	"github.com/LastStep/Bonsai/internal/catalog"
	"github.com/LastStep/Bonsai/internal/config"
	"github.com/LastStep/Bonsai/internal/generate"
)

// Exit codes shared between the runner and the cobra wiring. Kept here
// rather than in cmd/ so unit tests that only import internal/nonint can
// assert against named constants. These constants are the canonical source
// of truth for the headless mutating-command exit-code contract — the
// docs/agent-interface.md table (Plan 41 Phase 5) points back here.
//
// Per-command reachability (Plan 41 Verification matrix):
//
//	init   : 0, 2, 3, 4
//	add    : 0, 2, 3, 4
//	update : 0, 2, 3, 4, 5   (5 = conflict without --skip-conflicts)
//	remove : 0, 2, 3, 4
const (
	// ExitOK — success. Emitted by every mutating command.
	ExitOK = 0
	// ExitInvalidConfig — caller-supplied config/args rejected before any
	// mutation (bad overlay shape, missing tech-lead, unknown agent type,
	// multi-owner item with no --from, last-tech-lead removal, empty/`*`
	// target). Emitted by init/add/update/remove.
	ExitInvalidConfig = 2
	// ExitRuntime — a generator or filesystem error occurred mid-run.
	// Emitted by init/add/update/remove.
	ExitRuntime = 3
	// ExitWrongCWDForInit — wrong working-directory state: .bonsai.yaml
	// already present (init) or missing (add/update/remove). Emitted by
	// init/add/update/remove.
	ExitWrongCWDForInit = 4
	// ExitConflict — unresolved file conflicts; re-run with --skip-conflicts
	// or interactively. Emitted by update only (init/add force-skip
	// conflicts under --non-interactive and never reach this code).
	ExitConflict = 5
)

// RunInit performs a full `bonsai init` from cfg. cfg is assumed to be the
// output of LoadConfig — already defaulted and validated. The runner mirrors
// cmd.buildGenerateAction's filesystem pipeline exactly so generated output
// is bit-identical to the interactive flow for the same inputs.
//
// It is a pure headless core: typed options in, structured *Result out. It
// performs NO output itself — the CLI adapter serialises the Result to JSONL
// on stdout (via EmitJSONL) and prints Result.Warnings to stderr. The future
// MCP adapter (Plan 42) serialises the same Result to structuredContent.
// Conflicts under --non-interactive are forced to skip (Plan 39 Locked
// Decision Q5 + Decision §3) — the runner never calls ForceSelected or
// ForceConflicts.
//
// Returns (*Result, exitCode, error). On success the Result carries every
// file outcome (created/updated/unchanged/skipped/conflicts) plus any
// non-fatal Warnings (lock-save failure). exitCode is ExitOK on success,
// ExitRuntime on any generator error, ExitInvalidConfig on a rejected cfg
// shape, or ExitWrongCWDForInit when configPath already exists (interactive
// path's "Skipping init" branch made the same choice). On the error/reject
// paths the Result is nil.
//
// version is the Bonsai version string written into .bonsai/catalog.json —
// the cobra entry point passes cmd.Version. Kept as a parameter so the
// runner does not import cmd/.
func RunInit(cwd, configPath string, cfg *config.ProjectConfig, cat *catalog.Catalog, version string) (*Result, int, error) {
	// Pre-flight: matches the interactive path's "configFile already exists"
	// early-exit. Plan 39 §A.2.
	if _, err := os.Stat(configPath); err == nil {
		return nil, ExitWrongCWDForInit, fmt.Errorf("from-config: %s already exists at %s — refusing to overwrite", filepath.Base(configPath), filepath.Dir(configPath))
	}

	agentDef := cat.GetAgent(techLeadType)
	if agentDef == nil {
		return nil, ExitRuntime, fmt.Errorf("from-config: tech-lead agent missing from catalog (binary bug — please report)")
	}
	installed, ok := cfg.Agents[techLeadType]
	if !ok || installed == nil {
		// Defence-in-depth — LoadConfig's init-path caller validates this,
		// but a direct RunInit caller could skip that step.
		return nil, ExitInvalidConfig, fmt.Errorf("from-config: bonsai init requires a 'tech-lead' entry under agents:")
	}
	// Defence-in-depth for the exclusivity rule: `bonsai init` materialises
	// only tech-lead's workspace. Any extra agent entries would be partially
	// installed (registered in .bonsai.yaml + path-scoped rules but with no
	// workspace files) — reject so the failure mode is loud, not silent.
	if got := len(cfg.Agents); got != 1 {
		return nil, ExitInvalidConfig, fmt.Errorf("from-config: bonsai init accepts only a single 'tech-lead' entry under agents:, got %d agents", got)
	}

	lock, _ := config.LoadLockFile(cwd)
	res := &Result{Write: &generate.WriteResult{}}

	if err := cfg.Save(configPath); err != nil {
		return nil, ExitRuntime, fmt.Errorf("from-config: save %s: %w", configPath, err)
	}

	// Pipeline mirrors cmd.buildGenerateAction. force=false everywhere so
	// the writeFile lock-aware policy treats user-modified files as conflicts
	// (rather than silently overwriting them).
	var errs []error
	errs = append(errs, generate.Scaffolding(cwd, cfg, cat, lock, res.Write, false))
	errs = append(errs, generate.AgentWorkspace(cwd, agentDef, installed, cfg, cat, lock, res.Write, false))
	errs = append(errs, generate.PathScopedRules(cwd, cfg, cat, lock, res.Write, false))
	errs = append(errs, generate.WorkflowSkills(cwd, cfg, cat, lock, res.Write, false))
	errs = append(errs, generate.SettingsJSON(cwd, cfg, cat, lock, res.Write, false))
	errs = append(errs, generate.WriteCatalogSnapshot(cwd, version, cat, res.Write))
	if joined := errors.Join(errs...); joined != nil {
		return nil, ExitRuntime, fmt.Errorf("from-config: generate: %w", joined)
	}

	// Lock-save failure is not fatal — mirrors the interactive path's
	// `tui.Warning` swallow. The warning rides in Result.Warnings (NOT the
	// JSONL stream) so the CLI adapter can route it to stderr and stdout
	// stays pure protocol.
	if err := lock.Save(cwd); err != nil {
		res.Warnings = append(res.Warnings, "could not save lock file: "+err.Error())
	}
	return res, ExitOK, nil
}

// RunAdd appends a single agent (overlay.Agents has exactly one entry) or
// extra abilities to an existing agent inside an already-initialised project.
// cfg here is the OVERLAY config (from --from-config), not the project's
// own `.bonsai.yaml`. The runner loads the existing `.bonsai.yaml`, validates
// that overlay non-`agents` fields match (Plan 39 Locked Decision §3),
// and dispatches to either the new-agent or add-items branch.
//
// Like RunInit, it is a pure headless core that performs no output: it builds
// and returns a *Result the CLI adapter renders to JSONL (stdout) + stderr
// warnings. The future MCP adapter consumes the same Result.
//
// Returns (*Result, exitCode, error):
//   - ExitOK             — success (Result carries file outcomes + warnings;
//     the all-installed short-circuit returns a zero-count Result)
//   - ExitInvalidConfig  — overlay shape rejected (>1 agent, non-matching
//     project_name / docs_path / scaffolding, unknown
//     agent type, tech-lead-required violation)
//   - ExitRuntime        — generator error
//   - ExitWrongCWDForInit — no `.bonsai.yaml` at cwd
//
// On the error/reject paths the Result is nil.
func RunAdd(cwd, configPath string, overlay *config.ProjectConfig, cat *catalog.Catalog, version string) (*Result, int, error) {
	if _, err := os.Stat(configPath); errors.Is(err, os.ErrNotExist) {
		return nil, ExitWrongCWDForInit, fmt.Errorf("from-config: %s not found at %s — run `bonsai init` first", filepath.Base(configPath), filepath.Dir(configPath))
	} else if err != nil {
		return nil, ExitRuntime, fmt.Errorf("from-config: stat %s: %w", configPath, err)
	}

	existing, err := config.Load(configPath)
	if err != nil {
		return nil, ExitRuntime, fmt.Errorf("from-config: load existing %s: %w", configPath, err)
	}

	// Locked Decision §2 — exactly one agent in the overlay. Multi-agent
	// setups must invoke RunAdd in a loop (Bonsai-Eval rung-3 solver does).
	if got := len(overlay.Agents); got != 1 {
		return nil, ExitInvalidConfig, fmt.Errorf("from-config: bonsai add expects exactly one agent in overlay, got %d", got)
	}

	// Locked Decision §3 — non-`agents` fields must match (or be empty).
	if err := assertOverlayMatchesExisting(overlay, existing); err != nil {
		return nil, ExitInvalidConfig, err
	}

	// Pull the single agent out of the overlay. range-over-map is safe here
	// because we've already asserted len == 1.
	var agentType string
	var overlayAgent *config.InstalledAgent
	for k, v := range overlay.Agents {
		agentType, overlayAgent = k, v
	}
	if overlayAgent == nil {
		// LoadConfig coerces nil to non-nil via applyDefaults, but a direct
		// caller could bypass that — defensive guard.
		return nil, ExitInvalidConfig, fmt.Errorf("from-config: overlay agent %q has no body", agentType)
	}

	agentDef := cat.GetAgent(agentType)
	if agentDef == nil {
		return nil, ExitInvalidConfig, fmt.Errorf("from-config: unknown agent type %q (not in catalog)", agentType)
	}

	// Locked Decision §4 — tech-lead-required guard, exit 2 (not 3).
	if agentType != techLeadType {
		if _, ok := existing.Agents[techLeadType]; !ok {
			return nil, ExitInvalidConfig, fmt.Errorf("from-config: agent %s requires a tech-lead agent in the existing project", agentType)
		}
	}

	lock, _ := config.LoadLockFile(cwd)
	res := &Result{Write: &generate.WriteResult{}}

	_, isNewAgent, total, runErr := mergeAndGenerate(cwd, configPath, version, agentType, overlayAgent, agentDef, existing, cat, lock, res.Write)
	if runErr != nil {
		return nil, ExitRuntime, runErr
	}

	// All-installed short-circuit (Decision §6): add-items overlay against
	// an already-installed agent picked zero new abilities. The empty Write
	// renders as a single zero-count summary line via EmitJSONL — identical
	// to the old EmitSummary(w,0,0,0,0,0) call. Mirrors the interactive
	// `YieldAllInstalled` semantics. No lock save: nothing was written.
	if !isNewAgent && total == 0 {
		return res, ExitOK, nil
	}

	if err := lock.Save(cwd); err != nil {
		res.Warnings = append(res.Warnings, "could not save lock file: "+err.Error())
	}
	return res, ExitOK, nil
}

// assertOverlayMatchesExisting enforces the §3 contract: the overlay's
// project_name / docs_path / scaffolding fields are either empty (so applyDefaults
// filled them or left them zero-valued) or sorted-equal to the existing
// project's. Description is cosmetic and ignored.
//
// "Empty" semantics:
//   - ProjectName: empty after Load means user omitted it AND applyDefaults
//     populated it from filepath.Base(cwd). We compare directly here rather
//     than re-deriving — if applyDefaults happened to produce a value that
//     differs from `existing.ProjectName`, the user's overlay implicitly
//     disagrees and we reject.
//   - DocsPath: similar — applyDefaults gives "station/" when omitted. We
//     accept either ""-after-load OR the value-after-default-equal-to-existing.
//     LoadConfig runs applyDefaults BEFORE this call, so we see post-default
//     values; the equality check is enough.
//   - Scaffolding: applyDefaults sets the required-only fallback list. The
//     check accepts that fallback when sorted-equal to existing, and accepts
//     a nil/empty list (user-overlay opted out entirely) only when existing's
//     is also empty.
func assertOverlayMatchesExisting(overlay, existing *config.ProjectConfig) error {
	if overlay.ProjectName != "" && overlay.ProjectName != existing.ProjectName {
		return fmt.Errorf("from-config: overlay field 'project_name' (%q) does not match existing .bonsai.yaml (%q); leave empty or match exactly", overlay.ProjectName, existing.ProjectName)
	}
	if overlay.DocsPath != "" && overlay.DocsPath != existing.DocsPath {
		return fmt.Errorf("from-config: overlay field 'docs_path' (%q) does not match existing .bonsai.yaml (%q); leave empty or match exactly", overlay.DocsPath, existing.DocsPath)
	}
	// Scaffolding: accept nil/empty as "match". Otherwise require sorted-equal.
	if len(overlay.Scaffolding) > 0 {
		a := append([]string(nil), overlay.Scaffolding...)
		b := append([]string(nil), existing.Scaffolding...)
		sort.Strings(a)
		sort.Strings(b)
		if !reflect.DeepEqual(a, b) {
			return fmt.Errorf("from-config: overlay field 'scaffolding' (%v) does not match existing .bonsai.yaml (%v); leave empty or match exactly", overlay.Scaffolding, existing.Scaffolding)
		}
	}
	return nil
}

// mergeAndGenerate mutates `existing` to include the overlay agent's items
// then drives the generator chain. Returns (installedAgent, isNewAgent,
// totalChanges, err).
//
// totalChanges semantics:
//   - new-agent: total == len of every ability list in the new agent
//   - add-items: total == count of items newly added to the existing agent
//     across all categories
//
// The caller uses totalChanges to detect the add-items zero-add short-circuit
// (Decision §6).
func mergeAndGenerate(
	cwd, configPath, version, agentType string,
	overlayAgent *config.InstalledAgent,
	agentDef *catalog.AgentDef,
	existing *config.ProjectConfig,
	cat *catalog.Catalog,
	lock *config.LockFile,
	wr *generate.WriteResult,
) (*config.InstalledAgent, bool, int, error) {
	installedAgent, alreadyInstalled := existing.Agents[agentType]

	if !alreadyInstalled {
		// new-agent branch. Build the InstalledAgent from the overlay's
		// ability lists (applyDefaults already filled them with agent.yaml
		// defaults when the user omitted them), wire EnsureRoutineCheckSensor,
		// then drive the per-agent generator pipeline. Mirrors
		// cmd.buildAddGrowAction's new-agent branch.
		newAgent := &config.InstalledAgent{
			AgentType: agentType,
			Workspace: overlayAgent.Workspace,
			Skills:    append([]string(nil), overlayAgent.Skills...),
			Workflows: append([]string(nil), overlayAgent.Workflows...),
			Protocols: append([]string(nil), overlayAgent.Protocols...),
			Sensors:   append([]string(nil), overlayAgent.Sensors...),
			Routines:  append([]string(nil), overlayAgent.Routines...),
		}
		generate.EnsureRoutineCheckSensor(newAgent)
		existing.Agents[agentType] = newAgent

		if err := existing.Save(configPath); err != nil {
			return nil, true, 0, fmt.Errorf("from-config: save %s: %w", configPath, err)
		}
		var errs []error
		errs = append(errs, generate.AgentWorkspace(cwd, agentDef, newAgent, existing, cat, lock, wr, false))
		errs = append(errs, generate.PathScopedRulesForAgent(cwd, newAgent, existing, cat, lock, wr, false))
		errs = append(errs, generate.WorkflowSkillsForAgent(cwd, newAgent, existing, cat, lock, wr, false))
		errs = append(errs, generate.SettingsJSONForAgent(cwd, newAgent, existing, cat, lock, wr, false))
		errs = append(errs, generate.RefreshPeerAwareness(cwd, newAgent.AgentType, existing, cat, lock, wr, false))
		errs = append(errs, generate.WriteCatalogSnapshot(cwd, version, cat, wr))
		if joined := errors.Join(errs...); joined != nil {
			return newAgent, true, 0, fmt.Errorf("from-config: generate: %w", joined)
		}
		total := len(newAgent.Skills) + len(newAgent.Workflows) + len(newAgent.Protocols) +
			len(newAgent.Sensors) + len(newAgent.Routines)
		return newAgent, true, total, nil
	}

	// add-items branch — agent is already installed. Append only newly-named
	// abilities per category. EnsureRoutineCheckSensor runs again afterwards
	// in case the overlay added a routine to a previously routine-less agent.
	added := mergeNewAbilities(installedAgent, overlayAgent)
	generate.EnsureRoutineCheckSensor(installedAgent)
	if added == 0 {
		// Short-circuit: no new abilities. Defensive — caller checks this
		// total and emits the zero-summary line. Skip the save+generate so
		// the lock + .bonsai.yaml stay byte-identical.
		return installedAgent, false, 0, nil
	}

	if err := existing.Save(configPath); err != nil {
		return installedAgent, false, added, fmt.Errorf("from-config: save %s: %w", configPath, err)
	}
	var errs []error
	errs = append(errs, generate.AgentWorkspace(cwd, agentDef, installedAgent, existing, cat, lock, wr, false))
	errs = append(errs, generate.PathScopedRulesForAgent(cwd, installedAgent, existing, cat, lock, wr, false))
	errs = append(errs, generate.WorkflowSkillsForAgent(cwd, installedAgent, existing, cat, lock, wr, false))
	errs = append(errs, generate.SettingsJSONForAgent(cwd, installedAgent, existing, cat, lock, wr, false))
	errs = append(errs, generate.RefreshPeerAwareness(cwd, installedAgent.AgentType, existing, cat, lock, wr, false))
	errs = append(errs, generate.WriteCatalogSnapshot(cwd, version, cat, wr))
	if joined := errors.Join(errs...); joined != nil {
		return installedAgent, false, added, fmt.Errorf("from-config: generate: %w", joined)
	}
	return installedAgent, false, added, nil
}

// mergeNewAbilities appends every overlay ability not already present in
// `installed` and returns the count added. Each category is independently
// deduplicated. Order within a category preserves the overlay's order for
// the additions — pre-existing items keep their position at the head.
func mergeNewAbilities(installed, overlay *config.InstalledAgent) int {
	added := 0
	add := func(have []string, want []string) ([]string, int) {
		set := make(map[string]bool, len(have))
		for _, s := range have {
			set[s] = true
		}
		n := 0
		for _, s := range want {
			if set[s] {
				continue
			}
			have = append(have, s)
			set[s] = true
			n++
		}
		return have, n
	}
	var n int
	installed.Skills, n = add(installed.Skills, overlay.Skills)
	added += n
	installed.Workflows, n = add(installed.Workflows, overlay.Workflows)
	added += n
	installed.Protocols, n = add(installed.Protocols, overlay.Protocols)
	added += n
	installed.Sensors, n = add(installed.Sensors, overlay.Sensors)
	added += n
	installed.Routines, n = add(installed.Routines, overlay.Routines)
	added += n
	return added
}
