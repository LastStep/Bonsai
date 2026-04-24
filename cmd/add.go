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
	"github.com/LastStep/Bonsai/internal/tui/hints"
	"github.com/LastStep/Bonsai/internal/tui/initflow"
)

func init() {
	rootCmd.AddCommand(addCmd)
}

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add an agent to the project.",
	RunE:  runAdd,
}

// runAdd is the entry point for `bonsai add`. It renders the cinematic
// add-flow (Plan 23) — Select → [Ground] → Graft → Observe → Grow → [Conflicts]
// → Yield — fully replacing the legacy Huh+harness path. Branches:
//
//   - new-agent              : agent type is not in cfg.Agents
//   - add-items              : agent type already exists
//   - all-installed          : add-items short-circuit when nothing to add
//   - tech-lead-required     : non-tech-lead pick with no tech-lead present
//   - unknown-agent          : pick name not in catalog (stale binary)
//
// Step list:
//
//	[0] Select             agent picker
//	[1] LazyGroup          splices one of:
//	      new-agent         [Ground(Lazy), Graft(NewAgent), Observe]
//	      add-items         [Graft(AddItems, Installed), Observe]
//	      all-installed     [YieldAllInstalled]                 (terminal)
//	      tech-lead-req     [YieldTechLeadRequired]             (terminal)
//	      unknown-agent     [YieldUnknownAgent]                 (terminal)
//	[2] Conditional(Lazy(Grow))  gated on observeConfirmed
//	[3] LazyGroup          cinematic ConflictsStage iff wr.HasConflicts()
//	[4] Conditional(Lazy(Yield)) gated on growSucceeded
func runAdd(cmd *cobra.Command, args []string) error {
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
	// here is empty so the header right-block reads "GRAFTING INTO" over the
	// project path until the agent is picked.
	ctx := initflow.StageContext{
		Version:          Version,
		ProjectDir:       cwd,
		StationDir:       "station/",
		AgentDisplay:     "",
		StartedAt:        startedAt,
		HeaderAction:     "ADD",
		HeaderRightLabel: "GRAFTING INTO",
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
	// growSucceeded reads the addflow.Outcome scratchpad via closure capture
	// instead of walking prev[] for an error value. The Grow action's first
	// statement sets outcome.Ran=true and SpinnerErr is the only error source
	// on the happy path, so checking those two fields is sufficient.
	//
	// Safety: ConditionalStep evaluates this predicate at Init time on the
	// Yield slot, which happens after the harness drains the Grow stage's
	// generateDoneMsg — the action goroutine has fully returned by then so
	// the outcome fields are stable.
	growSucceeded := func(prev []any) bool {
		if !observeConfirmed(prev) {
			return false
		}
		return outcome.SpinnerErr == nil && outcome.Ran
	}

	// Step list ---------------------------------------------------------
	steps := []harness.Step{
		addflow.NewSelectStage(ctx, addflow.BuildAgentOptions(cat, installedSet(cfg))),
		harness.NewLazyGroup("Agent flow", func(prev []any) []harness.Step {
			agentType := asString(prev[0])
			agentDef := cat.GetAgent(agentType)
			if agentDef == nil {
				// Unknown agent — render the dedicated unknown-agent variant
				// pointing the user at `bonsai update` to refresh the catalog.
				return []harness.Step{addflow.NewYieldUnknownAgent(ctx, agentType)}
			}

			// Add-items branch. When the agent is already installed, either
			// short-circuit to YieldAllInstalled (nothing uninstalled across
			// any category) or splice the real sub-sequence
			// [AddItemsGraft, Observe]. Ground is skipped — the workspace is
			// already set. AgentDisplay stamps the agent's display name into
			// the chrome header so every downstream stage inherits the
			// correct "GRAFTING INTO" right-block.
			if installedAgent, exists := cfg.Agents[agentType]; exists {
				if availableAddItems(cat, installedAgent).Total() == 0 {
					return []harness.Step{addflow.NewYieldAllInstalled(ctx, agentDef)}
				}
				agentCtx := ctx
				agentCtx.AgentDisplay = agentDef.DisplayName
				if agentCtx.AgentDisplay == "" {
					agentCtx.AgentDisplay = catalog.DisplayNameFrom(agentDef.Name)
				}
				graft := addflow.NewAddItemsBranches(agentCtx, addflow.BranchesContext{
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
			graft := addflow.NewNewAgentBranches(agentCtx, addflow.BranchesContext{
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
				stage := addflow.NewYieldSuccess(ctx, installed, cat, outcome.NewAgent, total)
				// Plan 31 Phase H — render hints for the NEW agent being
				// grafted so the user gets agent-specific onboarding copy
				// (backend hints for a backend add, etc.).
				if installed != nil {
					block, _ := hints.Load(cat, installed.AgentType, "add", hints.TemplateContext{
						DocsPath:    installed.Workspace,
						AgentName:   installed.AgentType,
						ProjectName: cfg.ProjectName,
					})
					stage.SetHintBlock(hints.Render(block, initflow.PanelContentWidth))
				}
				return stage
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
		// GrowStage already surfaced the error in-frame via
		// initflow.GenerateStage.stateError and waited for a keypress;
		// the terminal has been cleared, so no post-harness Warning here.
		return nil
	}

	// Cinematic conflict picker — single stage spliced into the trailing
	// LazyGroup. Locate it by type-scanning results for the
	// map[string]config.ConflictAction shape that ConflictsStage publishes.
	// Type-scan beats positional arithmetic (len(results)-2) because the
	// Yield slot or other tail stages may shift; the Conflicts stage is the
	// only step in this flow that emits this map type so the scan is
	// unambiguous.
	if wr.HasConflicts() {
		for _, r := range results {
			if picks, ok := r.(map[string]config.ConflictAction); ok {
				applyCinematicConflictPicks(picks, &wr, lock, cwd)
				break
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
// Backup-failure handling: when the .bak read OR write step fails for a
// given path, that path is dropped from the overwrite list and a single
// collected tui.Warning is emitted naming all dropped paths. This avoids
// silently overwriting the user's local edits without a recoverable backup.
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
	// Track backup failures so the path can be dropped from toOverwrite. We
	// build a set rather than mutating toOverwrite during the read/write loop
	// so the slice stays stable until we filter at the end.
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
		droppedList := make([]string, 0, len(dropped))
		for _, rel := range toOverwrite {
			if dropped[rel] {
				droppedList = append(droppedList, rel)
				continue
			}
			filtered = append(filtered, rel)
		}
		toOverwrite = filtered
		tui.Warning("Could not write backup for: " + strings.Join(droppedList, ", ") + " — original file left unchanged.")
	}
	if len(toOverwrite) == 0 {
		return false
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
// the legacy add-spinner semantics exactly — same cfg.Save ordering, same
// generator pipeline, same error plumbing — so file output is bit-identical
// to the prior path for the same picks.
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
		var graft addflow.BranchesResult
		var graftFound bool
		for _, v := range prev {
			if g, ok := v.(addflow.BranchesResult); ok {
				graft = g
				graftFound = true
				break
			}
		}
		if !graftFound {
			outcome.SpinnerErr = fmt.Errorf("no graft selection captured")
			return outcome.SpinnerErr
		}

		// Add-items branch. The installed agent already exists; append only
		// the newly-picked items per category, then regenerate.
		// distributeAddItemPicks expects the picks slice to contain one
		// entry per category with at least one uninstalled item, in the
		// same skip-empty-categories order availableAddItems produces, so
		// build picks the same way and the shared helper produces
		// bit-identical filesystem output for the same selection.
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
			// Plan 27 §B2: scope path-rules / workflow-skills / settings
			// regeneration to the single agent being augmented. The all-
			// agents variants (used by `bonsai update`) would regenerate
			// files under every installed workspace and flag any user-edited
			// rule / skill / settings file as a cross-agent conflict.
			errs = append(errs, generate.PathScopedRulesForAgent(cwd, installedAgent, cfg, cat, lock, wr, false))
			errs = append(errs, generate.WorkflowSkillsForAgent(cwd, installedAgent, cfg, cat, lock, wr, false))
			errs = append(errs, generate.SettingsJSONForAgent(cwd, installedAgent, cfg, cat, lock, wr, false))
			// Plan 31 Phase A: re-render peers' OtherAgents-dependent files so
			// already-installed agents' scope-guard / dispatch-guard / identity
			// reflect the newly-augmented agent. No-op on the add-items branch
			// because cfg.Agents set is unchanged here — but kept for symmetry
			// and defence in depth; peers' render is a cheap no-diff write.
			errs = append(errs, generate.RefreshPeerAwareness(cwd, installedAgent.AgentType, cfg, cat, lock, wr, false))
			// Plan 31 Phase C: refresh .bonsai/catalog.json snapshot.
			errs = append(errs, generate.WriteCatalogSnapshot(cwd, Version, cat, wr))
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
		// Plan 27 §B2: see the add-items branch above for rationale — scope
		// to the single newly-installed agent.
		errs = append(errs, generate.PathScopedRulesForAgent(cwd, installed, cfg, cat, lock, wr, false))
		errs = append(errs, generate.WorkflowSkillsForAgent(cwd, installed, cfg, cat, lock, wr, false))
		errs = append(errs, generate.SettingsJSONForAgent(cwd, installed, cfg, cat, lock, wr, false))
		// Plan 31 Phase A: re-render peers' OtherAgents-dependent files
		// (identity.md, scope-guard-files.sh, dispatch-guard.sh) so existing
		// sibling agents pick up the new agent in their awareness lists.
		// Without this, tech-lead's scope-guard has no block for the newly
		// added backend/ workspace and silently fails open — see plan §Phase A.
		errs = append(errs, generate.RefreshPeerAwareness(cwd, installed.AgentType, cfg, cat, lock, wr, false))
		// Plan 31 Phase C: refresh .bonsai/catalog.json snapshot.
		errs = append(errs, generate.WriteCatalogSnapshot(cwd, Version, cat, wr))
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

// distributeAddItemPicks splits the per-category picks slice into the five
// category-typed slices, respecting the same skip-empty-categories rule that
// the AddItemsGraft stage uses when building the tab list. Invariant: picks
// is provided with one entry per non-empty category, in catalog iteration
// order (skills → workflows → protocols → sensors → routines), so each take()
// pulls from the slot that corresponds to the next non-empty category.
func distributeAddItemPicks(cat *catalog.Catalog, installed *config.InstalledAgent, picks [][]string) (skills, workflows, protocols, sensors, routines []string) {
	installedItems := func(items []string) map[string]bool {
		m := make(map[string]bool, len(items))
		for _, item := range items {
			m[item] = true
		}
		return m
	}
	agentType := installed.AgentType
	hasNew := func(available []catalog.CatalogItem, existing []string) bool {
		have := installedItems(existing)
		for _, item := range available {
			if !have[item.Name] {
				return true
			}
		}
		return false
	}
	hasNewSensor := func(available []catalog.SensorItem, existing []string) bool {
		have := installedItems(existing)
		for _, item := range available {
			if !have[item.Name] && item.Name != "routine-check" {
				return true
			}
		}
		return false
	}
	hasNewRoutine := func(available []catalog.RoutineItem, existing []string) bool {
		have := installedItems(existing)
		for _, item := range available {
			if !have[item.Name] {
				return true
			}
		}
		return false
	}

	idx := 0
	take := func() []string {
		if idx >= len(picks) {
			return nil
		}
		p := picks[idx]
		idx++
		return p
	}
	if hasNew(cat.SkillsFor(agentType), installed.Skills) {
		skills = take()
	}
	if hasNew(cat.WorkflowsFor(agentType), installed.Workflows) {
		workflows = take()
	}
	if hasNew(cat.ProtocolsFor(agentType), installed.Protocols) {
		protocols = take()
	}
	if hasNewSensor(cat.SensorsFor(agentType), installed.Sensors) {
		sensors = take()
	}
	if hasNewRoutine(cat.RoutinesFor(agentType), installed.Routines) {
		routines = take()
	}
	return
}

// availableAddSet is the result of filtering the catalog against an installed
// agent: only the items not already installed in each category.
type availableAddSet struct {
	Skills    []catalog.CatalogItem
	Workflows []catalog.CatalogItem
	Protocols []catalog.CatalogItem
	Sensors   []catalog.SensorItem
	Routines  []catalog.RoutineItem
}

// Total reports the total count across all categories — used by the "all
// installed" short-circuit.
func (a availableAddSet) Total() int {
	return len(a.Skills) + len(a.Workflows) + len(a.Protocols) + len(a.Sensors) + len(a.Routines)
}

// availableAddItems computes the uninstalled-per-category slices for an
// already-installed agent. Shared by the LazyGroup splicer's empty-check
// (which routes to YieldAllInstalled) and by buildAddGrowAction (which
// rebuilds the picks ordering for distributeAddItemPicks) so both reach
// the same filter result.
func availableAddItems(cat *catalog.Catalog, installed *config.InstalledAgent) availableAddSet {
	agentType := installed.AgentType

	installedItems := func(items []string) map[string]bool {
		m := make(map[string]bool, len(items))
		for _, item := range items {
			m[item] = true
		}
		return m
	}

	filterItems := func(available []catalog.CatalogItem, existing []string) []catalog.CatalogItem {
		have := installedItems(existing)
		var result []catalog.CatalogItem
		for _, item := range available {
			if !have[item.Name] {
				result = append(result, item)
			}
		}
		return result
	}

	filterSensors := func(available []catalog.SensorItem, existing []string) []catalog.SensorItem {
		have := installedItems(existing)
		var result []catalog.SensorItem
		for _, item := range available {
			if !have[item.Name] && item.Name != "routine-check" {
				result = append(result, item)
			}
		}
		return result
	}

	filterRoutines := func(available []catalog.RoutineItem, existing []string) []catalog.RoutineItem {
		have := installedItems(existing)
		var result []catalog.RoutineItem
		for _, item := range available {
			if !have[item.Name] {
				result = append(result, item)
			}
		}
		return result
	}

	return availableAddSet{
		Skills:    filterItems(cat.SkillsFor(agentType), installed.Skills),
		Workflows: filterItems(cat.WorkflowsFor(agentType), installed.Workflows),
		Protocols: filterItems(cat.ProtocolsFor(agentType), installed.Protocols),
		Sensors:   filterSensors(cat.SensorsFor(agentType), installed.Sensors),
		Routines:  filterRoutines(cat.RoutinesFor(agentType), installed.Routines),
	}
}
