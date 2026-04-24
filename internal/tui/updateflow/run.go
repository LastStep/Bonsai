package updateflow

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/LastStep/Bonsai/internal/catalog"
	"github.com/LastStep/Bonsai/internal/config"
	"github.com/LastStep/Bonsai/internal/generate"
	"github.com/LastStep/Bonsai/internal/tui/harness"
	"github.com/LastStep/Bonsai/internal/tui/hints"
	"github.com/LastStep/Bonsai/internal/tui/initflow"
)

// Run drives the cinematic 5-stage `bonsai update` flow end-to-end. The
// caller (cmd/update.go) retains responsibility for config/lock
// persistence — Run just orchestrates the rail and reports the outcome.
//
// Stage order:
//
//	Discover (rail 0) → Select (rail 1, chromeless, gated on valid
//	discoveries) → Sync (rail 2) → Conflict (off-rail, gated on
//	wr.HasConflicts()) → Yield (rail 3)
//
// Returns a populated Result even on Ctrl-C so callers can decide on
// persistence based on Cancelled and SyncErr fields.
func Run(cwd string, cfg *config.ProjectConfig, cat *catalog.Catalog, lock *config.LockFile, version string) (Result, error) {
	startedAt := time.Now()

	ctx := initflow.StageContext{
		Version:          version,
		ProjectDir:       cwd,
		StationDir:       cfg.DocsPath,
		AgentDisplay:     "",
		StartedAt:        startedAt,
		HeaderAction:     "UPDATE",
		HeaderRightLabel: "SYNCING",
	}

	// Shared flow state — populated as stages complete.
	var wr generate.WriteResult
	configChanged := false
	var syncErr error
	// discover stage publishes its result via Result(); select reads
	// that via the harness prev[] slot.

	discoverStage := NewDiscoverStage(ctx, cwd, cfg, cat, lock)

	hasSelect := func(_ []any) bool { return discoverStage.HasValidDiscoveries() }

	// Build the Sync action closure. Reads per-agent selections from the
	// Select stage's Result() (via prev[] when Select ran; defaults to
	// empty map when it was skipped).
	buildSyncAction := func(prev []any) initflow.GenerateAction {
		return func() error {
			picks := extractSelectionMap(prev)
			// Apply user-accepted discoveries: mutate installed + lock
			// for every selected file. Uses legacy applyCustomFileSelection
			// semantics (lifted verbatim from cmd/update.go).
			for _, disc := range discoverStage.Discoveries() {
				selected := picks[disc.AgentName]
				if len(selected) == 0 {
					continue
				}
				if applyCustomFileSelection(disc.Installed, disc.Valid, selected, lock, cwd) {
					configChanged = true
				}
			}
			// Full re-render pipeline. Mirrors legacy cmd/update.go spinner
			// action verbatim — same call order, same error join.
			var errs []error
			// Deterministic agent-name order for stable WriteResult layout.
			agentNames := sortedAgentNames(cfg)
			for _, name := range agentNames {
				installed := cfg.Agents[name]
				agentDef := cat.GetAgent(installed.AgentType)
				if agentDef == nil {
					continue
				}
				generate.EnsureRoutineCheckSensor(installed)
				errs = append(errs, generate.AgentWorkspace(cwd, agentDef, installed, cfg, cat, lock, &wr, false))
			}
			errs = append(errs, generate.PathScopedRules(cwd, cfg, cat, lock, &wr, false))
			errs = append(errs, generate.WorkflowSkills(cwd, cfg, cat, lock, &wr, false))
			errs = append(errs, generate.SettingsJSON(cwd, cfg, cat, lock, &wr, false))
			// Plan 31 Phase C (preserved through Phase F rewrite).
			errs = append(errs, generate.WriteCatalogSnapshot(cwd, version, cat, &wr))
			if err := errors.Join(errs...); err != nil {
				syncErr = err
				return err
			}
			return nil
		}
	}

	steps := []harness.Step{
		discoverStage,
		// Select is conditional — only spliced when at least one valid
		// discovery exists. The wrapped stage runs chromeless.
		harness.NewConditional(
			harness.NewLazy("Select", func(_ []any) harness.Step {
				// Pull only agents with at least one valid entry.
				var eligible []AgentDiscoveries
				for _, d := range discoverStage.Discoveries() {
					if len(d.Valid) > 0 {
						eligible = append(eligible, d)
					}
				}
				return NewSelectStage(ctx, eligible)
			}),
			hasSelect,
		),
		// Sync — on-rail, always runs (no-op when no discoveries / picks).
		harness.NewLazy("Sync", func(prev []any) harness.Step {
			return NewSyncStage(ctx, buildSyncAction(prev))
		}),
		// Conflict stage — off-rail, spliced lazily only when wr has
		// conflicts. Copy of addflow.ConflictsStage (Plan 31 F hard
		// constraint #2).
		harness.NewLazyGroup("Resolve conflicts", func(_ []any) []harness.Step {
			if !wr.HasConflicts() {
				return nil
			}
			return []harness.Step{NewConflictsStage(ctx, &wr)}
		}),
		// Yield — terminal card, built via Lazy so the hint block +
		// yield mode can read live configChanged/syncErr state.
		harness.NewLazy("Yield", func(_ []any) harness.Step {
			// Hints keyed by the tech-lead agent type by convention —
			// bonsai-level commands ask tech-lead for next steps. Falls
			// back to any installed agent if tech-lead is absent.
			agentType := pickAgentForHints(cfg)
			block, _ := hints.Load(cat, agentType, "update", hints.TemplateContext{
				DocsPath:    cfg.DocsPath,
				AgentName:   agentType,
				ProjectName: cfg.ProjectName,
			})
			return NewYieldStage(ctx, YieldInputs{
				WriteResult:   &wr,
				ConfigChanged: configChanged,
				SyncErr:       syncErr,
				HintBlock:     block,
			})
		}),
	}

	bannerLine := "BONSAI"
	if version != "" && version != "dev" {
		bannerLine = fmt.Sprintf("BONSAI v%s", version)
	}

	results, err := harness.Run(bannerLine, "Updating workspace", steps)
	if err != nil {
		if errors.Is(err, harness.ErrAborted) {
			return Result{Cancelled: true}, nil
		}
		var bpe *harness.BuilderPanicError
		if errors.As(err, &bpe) {
			return Result{}, fmt.Errorf("harness builder panic in step %q: %v", bpe.Step, bpe.Value)
		}
		return Result{}, err
	}

	// Post-harness: apply conflict picks if the Conflict stage ran.
	for _, r := range results {
		if picks, ok := r.(map[string]config.ConflictAction); ok {
			applyCinematicConflictPicks(picks, &wr, lock, cwd)
			break
		}
	}

	return Result{
		ConfigChanged: configChanged,
		WriteResult:   &wr,
		SyncErr:       syncErr,
	}, nil
}

// RunStatic is the non-TTY fallback. Auto-accepts every valid discovery,
// runs the sync pipeline, and surfaces conflicts as a returned error
// (no interactive picker). The caller handles persistence identically to
// the interactive path based on the returned Result.
func RunStatic(cwd string, cfg *config.ProjectConfig, cat *catalog.Catalog, lock *config.LockFile, version string) (Result, error) {
	// Scan — same logic as DiscoverStage.scan but inlined so RunStatic
	// has zero TUI dependency.
	var discoveries []AgentDiscoveries
	agentNames := sortedAgentNames(cfg)
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
			}
		}
		if len(valid) > 0 {
			discoveries = append(discoveries, AgentDiscoveries{
				AgentName: agentName,
				Installed: installed,
				Valid:     valid,
			})
		}
	}

	var wr generate.WriteResult
	configChanged := false
	for _, disc := range discoveries {
		selected := make([]string, 0, len(disc.Valid))
		for _, d := range disc.Valid {
			selected = append(selected, d.Type+":"+d.Name)
		}
		if applyCustomFileSelection(disc.Installed, disc.Valid, selected, lock, cwd) {
			configChanged = true
		}
	}

	var errs []error
	for _, name := range agentNames {
		installed := cfg.Agents[name]
		agentDef := cat.GetAgent(installed.AgentType)
		if agentDef == nil {
			continue
		}
		generate.EnsureRoutineCheckSensor(installed)
		errs = append(errs, generate.AgentWorkspace(cwd, agentDef, installed, cfg, cat, lock, &wr, false))
	}
	errs = append(errs, generate.PathScopedRules(cwd, cfg, cat, lock, &wr, false))
	errs = append(errs, generate.WorkflowSkills(cwd, cfg, cat, lock, &wr, false))
	errs = append(errs, generate.SettingsJSON(cwd, cfg, cat, lock, &wr, false))
	errs = append(errs, generate.WriteCatalogSnapshot(cwd, version, cat, &wr))
	syncErr := errors.Join(errs...)

	res := Result{
		ConfigChanged: configChanged,
		WriteResult:   &wr,
		SyncErr:       syncErr,
	}
	if wr.HasConflicts() {
		res.SyncErr = errors.Join(syncErr, fmt.Errorf("update has unresolved conflicts; run interactively to resolve"))
	}
	return res, nil
}

// extractSelectionMap finds the map[string][]string result emitted by
// SelectStage in the harness prev slice. Returns an empty map when no
// Select stage ran (the conditional pruned it).
func extractSelectionMap(prev []any) map[string][]string {
	for _, v := range prev {
		if m, ok := v.(map[string][]string); ok {
			return m
		}
	}
	return map[string][]string{}
}

// sortedAgentNames returns the installed-agent names sorted
// alphabetically for deterministic rendering order.
func sortedAgentNames(cfg *config.ProjectConfig) []string {
	names := make([]string, 0, len(cfg.Agents))
	for name := range cfg.Agents {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// pickAgentForHints returns the agent type whose hints.yaml should
// supply the Yield hint block. Tech-lead is the canonical choice —
// bonsai-level commands default to tech-lead's guidance. Falls back to
// any installed agent when tech-lead is absent (defensive — should be
// rare post-Plan 31 Phase A).
func pickAgentForHints(cfg *config.ProjectConfig) string {
	if _, ok := cfg.Agents["tech-lead"]; ok {
		return "tech-lead"
	}
	for name := range cfg.Agents {
		return name
	}
	return "tech-lead"
}

// applyCustomFileSelection mutates installed + lock for the given agent
// based on user-selected keys. Lifted verbatim from cmd/update.go (pre-
// cinematic flow) — business logic is unchanged, only the call site
// moves.
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
			installed.Skills = appendUnique(installed.Skills, d.Name)
		case "workflow":
			installed.Workflows = appendUnique(installed.Workflows, d.Name)
		case "protocol":
			installed.Protocols = appendUnique(installed.Protocols, d.Name)
		case "sensor":
			installed.Sensors = appendUnique(installed.Sensors, d.Name)
		case "routine":
			installed.Routines = appendUnique(installed.Routines, d.Name)
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

// appendUnique appends name to slice unless already present.
func appendUnique(slice []string, name string) []string {
	for _, existing := range slice {
		if existing == name {
			return slice
		}
	}
	return append(slice, name)
}

// applyCinematicConflictPicks mirrors cmd/add.go's identically-named
// helper — writes per-file conflict resolutions (Keep / Overwrite /
// Backup) against wr + lock. Backup failures drop the path from the
// overwrite list with a silent skip (caller can surface the wr.Files
// count as signal); we don't emit tui.Warning here because the stage
// has already torn down the AltScreen.
func applyCinematicConflictPicks(picks map[string]config.ConflictAction,
	wr *generate.WriteResult, lock *config.LockFile, projectRoot string) {
	if len(picks) == 0 {
		return
	}
	toOverwrite := make([]string, 0, len(picks))
	toBackup := make([]string, 0, len(picks))
	for path, act := range picks {
		switch act {
		case config.ConflictActionBackup:
			toBackup = append(toBackup, path)
			toOverwrite = append(toOverwrite, path)
		case config.ConflictActionOverwrite:
			toOverwrite = append(toOverwrite, path)
		}
	}
	if len(toOverwrite) == 0 {
		return
	}
	dropped := make(map[string]bool)
	for _, rel := range toBackup {
		abs := filepath.Join(projectRoot, rel)
		data, readErr := os.ReadFile(abs)
		if readErr != nil {
			dropped[rel] = true
			continue
		}
		if writeErr := os.WriteFile(abs+".bak", data, 0644); writeErr != nil {
			dropped[rel] = true
			continue
		}
	}
	if len(dropped) > 0 {
		filtered := make([]string, 0, len(toOverwrite)-len(dropped))
		for _, rel := range toOverwrite {
			if dropped[rel] {
				continue
			}
			filtered = append(filtered, rel)
		}
		toOverwrite = filtered
	}
	if len(toOverwrite) == 0 {
		return
	}
	wr.ForceSelected(toOverwrite, projectRoot, lock)
}
