package nonint

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/LastStep/Bonsai/internal/catalog"
	"github.com/LastStep/Bonsai/internal/config"
	"github.com/LastStep/Bonsai/internal/generate"
)

// projectConfigFile is the canonical .bonsai.yaml basename. RunUpdate takes a
// cwd (not a configPath) per the Plan 41 signature, so it joins this here to
// persist config when a discovery mutated it. Mirrors cmd.configFile.
const projectConfigFile = ".bonsai.yaml"

// RunUpdate reconciles an already-initialised project's workspace: it scans
// for user-created custom files, auto-accepts every valid discovery, and
// re-renders the ability surface so CLAUDE.md / settings.json / catalog
// snapshot stay in sync. It is the SINGLE headless update implementation —
// the cinematic updateflow.Run path keeps its own harness-staged re-render
// pipeline, but the non-TTY fallback (updateflow.RunStatic) is now a thin
// shim that delegates here so there is no duplicate headless logic.
//
// Like RunInit / RunAdd it is a pure headless core: typed options in, a
// structured *Result out, no output of its own. The CLI adapter serialises
// the Result to JSONL on stdout (EmitJSONL) and prints Result.Warnings to
// stderr; the future MCP adapter (Plan 42) consumes the same Result.
//
// Behaviour:
//   - Auto-accept every valid discovery (clean frontmatter): the custom file
//     is tracked in the lock + registered in the agent's ability lists.
//   - Invalid discoveries (missing/malformed frontmatter) are NOT applied —
//     they ride in Result.Warnings (the old RunStatic wrote these to raw
//     os.Stderr; routing them through Warnings keeps stdout pure JSONL and
//     lets the CLI adapter own the stderr write).
//   - Conflicts: when skipConflicts is true, the conflicting files are left
//     untouched, counted as skipped/conflict in Result.Write, and the run
//     exits ExitOK. When skipConflicts is false, the run returns ExitConflict
//     (5) with a Result whose Write.Files STILL lists the conflict entries so
//     the caller (and a downstream agent) can see exactly which files blocked.
//     This replaces the old RunStatic behaviour where conflicts folded into
//     SyncErr and cobra surfaced a generic non-zero error.
//
// Persistence: when at least one discovery mutated the config, the project's
// .bonsai.yaml is saved; the lock file is always saved. Either save failing
// is non-fatal and surfaces in Result.Warnings (mirrors init/add).
//
// version is the Bonsai version string written into .bonsai/catalog.json.
// Kept as a parameter so the core does not import cmd/.
//
// Returns (*Result, exitCode, error):
//   - (res, ExitOK,       nil) — clean run (or skipConflicts with conflicts)
//   - (res, ExitConflict, nil) — unresolved conflicts, skipConflicts=false
//   - (nil, ExitRuntime,  err) — a generator/filesystem error occurred
func RunUpdate(cwd string, cfg *config.ProjectConfig, cat *catalog.Catalog, lock *config.LockFile, version string, skipConflicts bool) (*Result, int, error) {
	res := &Result{Write: &generate.WriteResult{}}

	// Scan — auto-accept valid discoveries, collect invalid ones as warnings.
	// Mirrors the old updateflow.RunStatic scan loop (run.go:193-227) with the
	// invalid-file signal moved from raw os.Stderr into Result.Warnings.
	agentNames := sortedConfigAgentNames(cfg)
	configChanged := false
	for _, agentName := range agentNames {
		installed := cfg.Agents[agentName]
		found, scanErr := generate.ScanCustomFiles(cwd, installed, lock)
		if scanErr != nil || len(found) == 0 {
			continue
		}
		var valid []generate.DiscoveredFile
		for _, d := range found {
			if d.Error == "" {
				valid = append(valid, d)
			} else {
				res.Warnings = append(res.Warnings, fmt.Sprintf("%s — %s", d.RelPath, d.Error))
			}
		}
		if len(valid) == 0 {
			continue
		}
		// Auto-accept: select every valid discovery for this agent.
		selected := make([]string, 0, len(valid))
		for _, d := range valid {
			selected = append(selected, d.Type+":"+d.Name)
		}
		if applyCustomFileSelection(installed, valid, selected, lock, cwd) {
			configChanged = true
		}
	}

	// Re-render pipeline. Same call order / error join as the old RunStatic
	// (run.go:241-255) and the cinematic Sync action.
	var errs []error
	for _, name := range agentNames {
		installed := cfg.Agents[name]
		agentDef := cat.GetAgent(installed.AgentType)
		if agentDef == nil {
			continue
		}
		generate.EnsureRoutineCheckSensor(installed)
		errs = append(errs, generate.AgentWorkspace(cwd, agentDef, installed, cfg, cat, lock, res.Write, false))
	}
	errs = append(errs, generate.PathScopedRules(cwd, cfg, cat, lock, res.Write, false))
	errs = append(errs, generate.WorkflowSkills(cwd, cfg, cat, lock, res.Write, false))
	errs = append(errs, generate.SettingsJSON(cwd, cfg, cat, lock, res.Write, false))
	errs = append(errs, generate.WriteCatalogSnapshot(cwd, version, cat, res.Write))
	if joined := errors.Join(errs...); joined != nil {
		return nil, ExitRuntime, fmt.Errorf("update: generate: %w", joined)
	}

	// Conflict gate. force=false above means user-modified files surface as
	// ActionConflict in Write.Files (never overwritten on disk).
	//
	//   - skipConflicts=true  → those files are intentionally skipped: rewrite
	//     each conflict entry to ActionSkipped (counted in `skipped`, file left
	//     untouched) and exit ExitOK. The agent sees them as deliberate skips.
	//   - skipConflicts=false → hard stop: leave the entries as ActionConflict
	//     so Write.Files lists exactly which files blocked, and exit ExitConflict.
	//
	// Persistence still runs below either way so the lock/config reflect any
	// non-conflicting tracking that did happen.
	exitCode := ExitOK
	if res.Write.HasConflicts() {
		if skipConflicts {
			for i := range res.Write.Files {
				if res.Write.Files[i].Action == generate.ActionConflict {
					res.Write.Files[i].Action = generate.ActionSkipped
				}
			}
		} else {
			exitCode = ExitConflict
		}
	}

	// Persistence — config only when a discovery mutated it; lock always.
	// Save failures are non-fatal (matches init/add); they ride in Warnings.
	if configChanged {
		configPath := filepath.Join(cwd, projectConfigFile)
		if err := cfg.Save(configPath); err != nil {
			res.Warnings = append(res.Warnings, "could not save config: "+err.Error())
		}
	}
	if err := lock.Save(cwd); err != nil {
		res.Warnings = append(res.Warnings, "could not save lock file: "+err.Error())
	}

	return res, exitCode, nil
}

// sortedConfigAgentNames returns the installed-agent names sorted
// alphabetically for deterministic rendering order. Mirrors
// updateflow.sortedAgentNames; duplicated here so the headless core has zero
// TUI dependency.
func sortedConfigAgentNames(cfg *config.ProjectConfig) []string {
	names := make([]string, 0, len(cfg.Agents))
	for name := range cfg.Agents {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// applyCustomFileSelection mutates installed + lock for the given agent based
// on user-selected keys ("<type>:<name>"). Lifted verbatim from
// updateflow.applyCustomFileSelection (which itself was lifted from the
// pre-cinematic cmd/update.go) — business logic is unchanged, only the home
// package moves so the headless core does not import updateflow's chrome.
func applyCustomFileSelection(installed *config.InstalledAgent, valid []generate.DiscoveredFile,
	selected []string, lock *config.LockFile, cwd string) bool {
	selectedSet := make(map[string]bool, len(selected))
	for _, s := range selected {
		selectedSet[s] = true
	}

	changed := false
	for _, d := range valid {
		if !selectedSet[d.Type+":"+d.Name] {
			continue
		}

		switch d.Type {
		case "skill":
			installed.Skills = appendUniqueName(installed.Skills, d.Name)
		case "workflow":
			installed.Workflows = appendUniqueName(installed.Workflows, d.Name)
		case "protocol":
			installed.Protocols = appendUniqueName(installed.Protocols, d.Name)
		case "sensor":
			installed.Sensors = appendUniqueName(installed.Sensors, d.Name)
		case "routine":
			installed.Routines = appendUniqueName(installed.Routines, d.Name)
		}

		if installed.CustomItems == nil {
			installed.CustomItems = make(map[string]*config.CustomItemMeta)
		}
		installed.CustomItems[d.Name] = d.Meta

		data, readErr := os.ReadFile(filepath.Join(cwd, d.RelPath))
		if readErr == nil {
			lock.Track(d.RelPath, data, "custom:"+d.Type+"s/"+d.Name)
		}

		changed = true
	}
	return changed
}

// appendUniqueName appends name to slice unless already present.
func appendUniqueName(slice []string, name string) []string {
	for _, existing := range slice {
		if existing == name {
			return slice
		}
	}
	return append(slice, name)
}
