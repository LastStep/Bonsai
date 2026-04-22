package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/LastStep/Bonsai/internal/catalog"
	"github.com/LastStep/Bonsai/internal/config"
	"github.com/LastStep/Bonsai/internal/generate"
	"github.com/LastStep/Bonsai/internal/tui"
	"github.com/LastStep/Bonsai/internal/tui/addflow"
	"github.com/LastStep/Bonsai/internal/tui/harness"
	"github.com/LastStep/Bonsai/internal/tui/initflow"
)

// runAddRedesign is the cinematic `bonsai add` entry point, gated by the
// BONSAI_ADD_REDESIGN=1 env flag during Plan 23 Phase 1+2. Phase 3 flips
// the default and deletes the legacy runAdd body.
//
// Step list:
//
//	[0] Select             agent picker
//	[1] LazyGroup          splices one of:
//	      new-agent         [Ground(Lazy), Graft(NewAgent), Observe]
//	      add-items         [Graft(AddItems, Installed), Observe]
//	      all-installed     [YieldAllInstalled]                 (terminal)
//	      tech-lead-req     [YieldTechLeadRequired]             (terminal)
//	[2] Conditional(Lazy(Grow))  gated on observeConfirmed
//	[3] LazyGroup          cinematic ConflictsStage iff wr.HasConflicts()
//	[4] Conditional(Lazy(Yield)) gated on growSucceeded
func runAddRedesign(cmd *cobra.Command, args []string) error {
	startedAt := time.Now()

	cwd := mustCwd()
	configPath := filepath.Join(cwd, configFile)
	cfg, err := requireConfig(configPath)
	if err != nil {
		return err
	}
	cat := loadCatalog()

	existingWorkspaces := make(map[string]bool)
	for _, a := range cfg.Agents {
		key := strings.TrimRight(filepath.Clean(a.Workspace), "/") + "/"
		existingWorkspaces[key] = true
	}

	// Shared context stamped onto every stage. AgentDisplay is resolved lazily
	// inside the splicer once the user picks an agent — the initial value
	// here is empty so the header right-block reads "PLANTING INTO" over the
	// project path (same convention as initflow).
	ctx := initflow.StageContext{
		Version:      Version,
		ProjectDir:   cwd,
		StationDir:   "station/",
		AgentDisplay: "",
		StartedAt:    startedAt,
	}

	lock, _ := config.LoadLockFile(cwd)
	var wr generate.WriteResult
	var installed *config.InstalledAgent
	var outcome addflow.Outcome

	// Predicates --------------------------------------------------------
	observeConfirmed := func(prev []any) bool {
		// ObserveStage is the only stage returning bool. Walk prev for it so
		// the predicate holds whether evaluated at Grow's Conditional (last
		// slot = observe bool) or Yield's Conditional (last slot = Grow's nil
		// result after Grow ran).
		for _, v := range prev {
			if b, ok := v.(bool); ok {
				return b
			}
		}
		return false
	}
	growSucceeded := func(prev []any) bool {
		if !observeConfirmed(prev) {
			return false
		}
		for _, v := range prev {
			if e, ok := v.(error); ok && e != nil {
				return false
			}
		}
		return true
	}

	// Step list ---------------------------------------------------------
	steps := []harness.Step{
		addflow.NewSelectStage(ctx, addflow.BuildAgentOptions(cat, installedSet(cfg))),
		harness.NewLazyGroup("Agent flow", func(prev []any) []harness.Step {
			agentType := asString(prev[0])
			agentDef := cat.GetAgent(agentType)
			if agentDef == nil {
				// Unknown agent — render the tech-lead-required variant with
				// the unknown name so the user has a durable next-step.
				return []harness.Step{addflow.NewYieldTechLeadRequired(ctx, agentType)}
			}

			// Add-items branch — Phase 2 wired. When the agent is already
			// installed, either short-circuit to YieldAllInstalled (nothing
			// uninstalled across any category) or splice the real sub-
			// sequence [AddItemsGraft, Observe]. Ground is skipped — the
			// workspace is already set. AgentDisplay stamps the agent's
			// display name into the chrome header so every downstream
			// stage inherits the correct "GRAFTING INTO" right-block.
			if installedAgent, exists := cfg.Agents[agentType]; exists {
				if availableAddItems(cat, installedAgent).Total() == 0 {
					return []harness.Step{addflow.NewYieldAllInstalled(ctx, agentDef)}
				}
				agentCtx := ctx
				agentCtx.AgentDisplay = agentDef.DisplayName
				if agentCtx.AgentDisplay == "" {
					agentCtx.AgentDisplay = catalog.DisplayNameFrom(agentDef.Name)
				}
				graft := addflow.NewAddItemsGraft(agentCtx, addflow.GraftContext{
					Cat:       cat,
					AgentType: agentType,
					AgentDef:  agentDef,
					Installed: installedAgent,
				})
				observe := addflow.NewObserveStage(agentCtx, cat)
				// Ground is skipped on add-items — seed the workspace from
				// the installed agent so Observe renders the real path.
				observe.SetDefaultWorkspace(installedAgent.Workspace)
				return []harness.Step{graft, observe}
			}

			// Tech-lead guard — non-tech-lead pick with no tech-lead installed.
			if agentType != "tech-lead" {
				if _, hasTechLead := cfg.Agents["tech-lead"]; !hasTechLead {
					return []harness.Step{addflow.NewYieldTechLeadRequired(ctx, agentType)}
				}
			}

			// New-agent branch.
			agentCtx := ctx
			agentCtx.AgentDisplay = agentDef.DisplayName
			if agentCtx.AgentDisplay == "" {
				agentCtx.AgentDisplay = catalog.DisplayNameFrom(agentDef.Name)
			}
			ground := addflow.NewGroundStage(agentCtx, addflow.GroundContext{
				AgentType:          agentType,
				DocsPath:           cfg.DocsPath,
				ExistingWorkspaces: existingWorkspaces,
			})
			graft := addflow.NewNewAgentGraft(agentCtx, addflow.GraftContext{
				Cat:       cat,
				AgentType: agentType,
				AgentDef:  agentDef,
			})
			observe := addflow.NewObserveStage(agentCtx, cat)
			return []harness.Step{ground, graft, observe}
		}),
		harness.NewConditional(
			harness.NewLazy("Grow", func(prev []any) harness.Step {
				action := buildAddGrowAction(prev, cwd, configPath, cfg, cat, existingWorkspaces, lock, &wr, &installed, &outcome)
				return addflow.NewGrowStage(ctx, action)
			}),
			observeConfirmed,
		),
		harness.NewLazyGroup("Resolve conflicts", func(prev []any) []harness.Step {
			if !growSucceeded(prev) {
				return nil
			}
			if !wr.HasConflicts() {
				return nil
			}
			// Cinematic path — one tabbed stage per conflict, Keep / Overwrite
			// / Backup radio per tab. Returns map[string]config.ConflictAction
			// consumed post-harness by applyCinematicConflictPicks.
			return []harness.Step{addflow.NewConflictsStage(ctx, &wr)}
		}),
		harness.NewConditional(
			harness.NewLazy("Yield", func(_ []any) harness.Step {
				total := outcome.TotalSelected
				if total == 0 && installed != nil {
					total = len(installed.Skills) + len(installed.Workflows) +
						len(installed.Protocols) + len(installed.Sensors) + len(installed.Routines)
				}
				return addflow.NewYieldSuccess(ctx, installed, cat, outcome.NewAgent, total)
			}),
			growSucceeded,
		),
	}

	bannerLine := "BONSAI"
	if Version != "" && Version != "dev" {
		bannerLine = fmt.Sprintf("BONSAI v%s", Version)
	}

	results, err := harness.Run(bannerLine, "Adding", steps)
	if err != nil {
		if errors.Is(err, harness.ErrAborted) {
			return nil
		}
		var bpe *harness.BuilderPanicError
		if errors.As(err, &bpe) {
			tui.FatalPanel("Harness builder panic",
				fmt.Sprintf("Step %q: %v", bpe.Step, bpe.Value),
				"This is a bug — please report it.")
			return nil
		}
		return err
	}

	// Post-harness cleanup when Grow ran successfully: apply conflict picks
	// and save the lock. When the user cancelled at Observe or the splice
	// terminated at a Yield variant there is nothing to persist.
	if !observeConfirmed(results) {
		return nil
	}
	if outcome.SpinnerErr != nil {
		tui.Warning("Generation error: " + outcome.SpinnerErr.Error())
		return nil
	}

	// Cinematic conflict picker — single stage sitting just before the
	// trailing Yield slot. Result is a per-file map of user picks. When
	// nothing was picked the helper no-ops; when some files were flagged
	// Overwrite/Backup it writes .bak files (Backup only) and dispatches to
	// WriteResult.ForceSelected.
	if wr.HasConflicts() {
		confIdx := len(results) - 2
		if confIdx >= 0 && confIdx < len(results) {
			if picks, ok := results[confIdx].(map[string]config.ConflictAction); ok {
				applyCinematicConflictPicks(picks, &wr, lock, cwd)
			}
		}
	}

	if err := lock.Save(cwd); err != nil {
		tui.Warning("Could not save lock file: " + err.Error())
	}
	return nil
}

// applyCinematicConflictPicks mutates wr + lock to apply the per-file
// resolution picks produced by addflow.ConflictsStage. Mirrors the legacy
// applyConflictPicks behaviour (wr.ForceSelected + optional .bak writes)
// but reads a map[string]config.ConflictAction rather than the two-step
// MultiSelect + Confirm shape legacy buildConflictSteps produces — written
// fresh so the legacy helper stays unchanged (it is shared with remove /
// update / legacy add).
//
// Selection semantics:
//
//   - ConflictActionKeep        — no write, no backup.
//   - ConflictActionOverwrite   — ForceSelected(path), no backup.
//   - ConflictActionBackup      — write <path>.bak, then ForceSelected(path).
//
// Returns true when at least one file mutated (for symmetry with the legacy
// helper's return shape — currently unused by callers but kept for
// diagnostic parity).
func applyCinematicConflictPicks(picks map[string]config.ConflictAction,
	wr *generate.WriteResult, lock *config.LockFile, projectRoot string) bool {
	if len(picks) == 0 {
		return false
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
		return false
	}
	for _, rel := range toBackup {
		abs := filepath.Join(projectRoot, rel)
		data, readErr := os.ReadFile(abs)
		if readErr == nil {
			_ = os.WriteFile(abs+".bak", data, 0644)
		}
	}
	wr.ForceSelected(toOverwrite, projectRoot, lock)
	return true
}

// installedSet builds the "already installed" lookup for BuildAgentOptions.
func installedSet(cfg *config.ProjectConfig) map[string]bool {
	out := make(map[string]bool, len(cfg.Agents))
	for name := range cfg.Agents {
		out[name] = true
	}
	return out
}

// buildAddGrowAction returns the GenerateAction closure for the Grow stage.
// Reads the upstream prev[] snapshot to decide which branch ran and applies
// the appropriate config + filesystem mutation. The cinematic flow reuses
// runAddSpinner's semantics exactly — same cfg.Save ordering, same generator
// pipeline, same error plumbing — so file output is bit-identical to the
// legacy path for the same picks.
//
// Branch detection is type-based rather than positional so the same helper
// handles both prev shapes:
//
//	new-agent : [agent(string), workspace(string), graft(GraftResult), ...]
//	add-items : [agent(string), graft(GraftResult), ...]
//
// cfg.Agents[agentType] exists ⇒ add-items branch; otherwise new-agent.
func buildAddGrowAction(
	prev []any,
	cwd, configPath string,
	cfg *config.ProjectConfig,
	cat *catalog.Catalog,
	existingWorkspaces map[string]bool,
	lock *config.LockFile,
	wr *generate.WriteResult,
	installedOut **config.InstalledAgent,
	outcome *addflow.Outcome,
) initflow.GenerateAction {
	return func() error {
		outcome.Ran = true

		agentType, _ := prev[0].(string)
		agentDef := cat.GetAgent(agentType)
		if agentDef == nil {
			outcome.SpinnerErr = fmt.Errorf("unknown agent type %q", agentType)
			return outcome.SpinnerErr
		}
		outcome.AgentDef = agentDef

		// Locate the GraftResult by type. Both branches emit exactly one.
		var graft addflow.GraftResult
		var graftFound bool
		for _, v := range prev {
			if g, ok := v.(addflow.GraftResult); ok {
				graft = g
				graftFound = true
				break
			}
		}
		if !graftFound {
			outcome.SpinnerErr = fmt.Errorf("no graft selection captured")
			return outcome.SpinnerErr
		}

		// Add-items branch — mirror runAddSpinner's add-items body. The
		// installed agent already exists; append only the newly-picked items
		// per category, then regenerate. distributeAddItemPicks expects the
		// picks slice to contain one entry per category with at least one
		// uninstalled item (same order the legacy buildAddItemsSteps emits
		// MultiSelect steps). Build picks the same way so both paths reach
		// the shared helper with the same shape and produce bit-identical
		// filesystem output.
		if installedAgent, exists := cfg.Agents[agentType]; exists {
			avail := availableAddItems(cat, installedAgent)
			picks := make([][]string, 0, 5)
			if len(avail.Skills) > 0 {
				picks = append(picks, graft.Skills)
			}
			if len(avail.Workflows) > 0 {
				picks = append(picks, graft.Workflows)
			}
			if len(avail.Protocols) > 0 {
				picks = append(picks, graft.Protocols)
			}
			if len(avail.Sensors) > 0 {
				picks = append(picks, graft.Sensors)
			}
			if len(avail.Routines) > 0 {
				picks = append(picks, graft.Routines)
			}
			selectedSkills, selectedWorkflows, selectedProtocols, selectedSensors, selectedRoutines :=
				distributeAddItemPicks(cat, installedAgent, picks)
			totalSelected := len(selectedSkills) + len(selectedWorkflows) +
				len(selectedProtocols) + len(selectedSensors) + len(selectedRoutines)
			if totalSelected == 0 {
				outcome.Workspace = installedAgent.Workspace
				outcome.NewAgent = false
				outcome.TotalSelected = 0
				*installedOut = installedAgent
				return nil
			}

			installedAgent.Skills = append(installedAgent.Skills, selectedSkills...)
			installedAgent.Workflows = append(installedAgent.Workflows, selectedWorkflows...)
			installedAgent.Protocols = append(installedAgent.Protocols, selectedProtocols...)
			installedAgent.Sensors = append(installedAgent.Sensors, selectedSensors...)
			installedAgent.Routines = append(installedAgent.Routines, selectedRoutines...)
			generate.EnsureRoutineCheckSensor(installedAgent)

			if err := cfg.Save(configPath); err != nil {
				outcome.SpinnerErr = err
				return err
			}

			var errs []error
			errs = append(errs, generate.AgentWorkspace(cwd, agentDef, installedAgent, cfg, cat, lock, wr, false))
			errs = append(errs, generate.PathScopedRules(cwd, cfg, cat, lock, wr, false))
			errs = append(errs, generate.WorkflowSkills(cwd, cfg, cat, lock, wr, false))
			errs = append(errs, generate.SettingsJSON(cwd, cfg, cat, lock, wr, false))
			if joined := errors.Join(errs...); joined != nil {
				outcome.SpinnerErr = joined
				return joined
			}

			outcome.Workspace = installedAgent.Workspace
			outcome.NewAgent = false
			outcome.TotalSelected = totalSelected
			*installedOut = installedAgent
			return nil
		}

		// New-agent branch. Workspace sits at prev[1] (from Ground). Tech-
		// lead overrides — its Ground stage auto-completed with DocsPath.
		workspace, _ := prev[1].(string)
		if workspace == "" {
			workspace = addflow.NormaliseWorkspace(agentType + "/")
		}
		if agentType == "tech-lead" {
			workspace = cfg.DocsPath
			if workspace == "" {
				workspace = "station/"
			}
		}
		if existingWorkspaces[workspace] {
			outcome.SpinnerErr = fmt.Errorf("workspace %q is already in use by another agent", workspace)
			return outcome.SpinnerErr
		}

		installed := &config.InstalledAgent{
			AgentType: agentType,
			Workspace: workspace,
			Skills:    append([]string(nil), graft.Skills...),
			Workflows: append([]string(nil), graft.Workflows...),
			Protocols: append([]string(nil), graft.Protocols...),
			Sensors:   append([]string(nil), graft.Sensors...),
			Routines:  append([]string(nil), graft.Routines...),
		}
		generate.EnsureRoutineCheckSensor(installed)
		cfg.Agents[agentType] = installed
		if err := cfg.Save(configPath); err != nil {
			outcome.SpinnerErr = err
			return err
		}

		var errs []error
		errs = append(errs, generate.AgentWorkspace(cwd, agentDef, installed, cfg, cat, lock, wr, false))
		errs = append(errs, generate.PathScopedRules(cwd, cfg, cat, lock, wr, false))
		errs = append(errs, generate.WorkflowSkills(cwd, cfg, cat, lock, wr, false))
		errs = append(errs, generate.SettingsJSON(cwd, cfg, cat, lock, wr, false))
		if joined := errors.Join(errs...); joined != nil {
			outcome.SpinnerErr = joined
			return joined
		}

		outcome.Workspace = workspace
		outcome.NewAgent = true
		outcome.TotalSelected = graft.Total()
		*installedOut = installed
		return nil
	}
}
