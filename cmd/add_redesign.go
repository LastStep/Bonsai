package cmd

import (
	"errors"
	"fmt"
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
//	      all-installed     [YieldAllInstalled]                 (terminal)
//	      tech-lead-req     [YieldTechLeadRequired]             (terminal)
//	      add-items         (Phase 2 — nil for now, legacy handles it)
//	[2] Conditional(Lazy(Grow))  gated on observeConfirmed
//	[3] LazyGroup          conflicts placeholder (Phase 2 no-op)
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
		// prev layout after splice:
		//   new-agent happy path:  [0]=agent, [1]=ws, [2]=graft, [3]=observe-bool
		//   all-installed / tech-lead-req terminal splices: last slot is nil (Yield.Result)
		//   add-items (Phase 2 TBD)
		if len(prev) == 0 {
			return false
		}
		b, ok := prev[len(prev)-1].(bool)
		return ok && b
	}
	growSucceeded := func(prev []any) bool {
		if !observeConfirmed(prev) {
			return false
		}
		// grow slot is cursor-1 vs the predicate-eval point. Walk prev from the
		// tail looking for an error — Grow publishes nil on success.
		for i := len(prev) - 1; i >= 0; i-- {
			if e, isErr := prev[i].(error); isErr && e != nil {
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

			// Add-items branch — Phase 1 terminal splices only. Plan 23
			// Phase 2 wires the real filtered Graft + Observe path; until
			// then this splicer either terminates at the "already full"
			// card or directs the user to the legacy flow via the
			// AddItemsDeferred variant. Never falls through silently.
			if installedAgent, exists := cfg.Agents[agentType]; exists {
				if availableAddItems(cat, installedAgent).Total() == 0 {
					return []harness.Step{addflow.NewYieldAllInstalled(ctx, agentDef)}
				}
				return []harness.Step{addflow.NewYieldAddItemsDeferred(ctx, agentDef)}
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
			return buildConflictSteps(&wr)
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

	// Conflict picker, when present, sits just before the trailing Yield slot.
	confIdx := -1
	if wr.HasConflicts() {
		confIdx = len(results) - 2
	}
	applyConflictPicks(results, confIdx, &wr, lock, cwd)

	if err := lock.Save(cwd); err != nil {
		tui.Warning("Could not save lock file: " + err.Error())
	}
	return nil
}

// installedSet builds the "already installed" lookup for BuildAgentOptions.
func installedSet(cfg *config.ProjectConfig) map[string]bool {
	out := make(map[string]bool, len(cfg.Agents))
	for name := range cfg.Agents {
		out[name] = true
	}
	return out
}

// buildAddGrowAction mirrors runAddSpinner's new-agent body but closes over
// the cinematic flow's prev indices (agent at [0], workspace at [1], Graft
// at [2], observe bool at [3]). Duplicated — not reused — so the cinematic
// path has no back-import into legacy helpers beyond loadCatalog et al.
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

		if _, exists := cfg.Agents[agentType]; exists {
			// Phase 1: add-items path terminates at the Yield splice before
			// reaching Grow. Defensive no-op.
			return nil
		}

		// New-agent path.
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

		graft, _ := prev[2].(addflow.GraftResult)

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
