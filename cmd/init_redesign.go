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
	"github.com/LastStep/Bonsai/internal/tui/harness"
	"github.com/LastStep/Bonsai/internal/tui/initflow"
)

// runInit is the entry point for `bonsai init`. It renders the cinematic
// four-stage init flow (Vessel → Soil → Branches → Observe) followed by the
// Generate progress stage, an optional conflict picker spliced in when any
// generated files clash with user edits, and the terminal Planted stage
// (Plan 22 Phase 5B).
func runInit(cmd *cobra.Command, args []string) error {
	// Capture the session start time once so every stage shares the same
	// origin — the Planted stage computes ELAPSED from here.
	startedAt := time.Now()

	cwd := mustCwd()
	configPath := filepath.Join(cwd, configFile)

	// Early-exit with the legacy warning copy so existing configs are
	// respected identically to pre-redesign behaviour.
	if _, err := os.Stat(configPath); err == nil {
		tui.WarningPanel(configFile + " already exists. Skipping init.")
		return nil
	}

	cat := loadCatalog()

	const techLeadType = "tech-lead"
	agentDef := cat.GetAgent(techLeadType)
	if agentDef == nil {
		tui.FatalPanel("Tech Lead agent not found",
			"The built-in catalog is missing the tech-lead agent.",
			"This is a bug — please report it.")
	}

	// Pull the display name with a derive-from-machine-name fallback so the
	// Observe stage can always render AGENT row correctly.
	agentDisplay := agentDef.DisplayName
	if agentDisplay == "" {
		agentDisplay = catalog.DisplayNameFrom(agentDef.Name)
	}

	// Shared context stamped on every stage.
	ctx := initflow.StageContext{
		Version:      Version,
		ProjectDir:   cwd,
		StationDir:   "station/",
		AgentDisplay: agentDisplay,
		StartedAt:    startedAt,
	}

	// Lock + WriteResult shared between the Generate action and the conflict
	// picker. cfg + installed are populated by the Generate closure; the
	// Planted stage reads `installed` through the closure to render counts.
	lock, _ := config.LoadLockFile(cwd)
	var wr generate.WriteResult
	var cfg *config.ProjectConfig
	var installed *config.InstalledAgent

	soilOptions := scaffoldingToSoilOptions(cat)

	// Gate for whether Observe confirmed PLANT. Predicate reads prev[3]
	// (Observe.Result() returns bool). Both Generate and downstream stages
	// depend on this — on cancel, all three skip and the harness exits.
	observeConfirmed := func(prev []any) bool {
		if len(prev) <= 3 {
			return false
		}
		b, _ := prev[3].(bool)
		return b
	}

	// Gate for whether to advance to Planted. Same as observeConfirmed plus:
	// Generate must not have returned an error (prev[4] is either nil on
	// success or an error value on failure). The plan explicitly requires
	// "Generate surfaces errors in an InfoPanel and does not advance to
	// Planted on failure" — this predicate enforces the non-advance piece;
	// the GenerateStage's stateError keypress handler owns the in-frame
	// error display.
	generateSucceeded := func(prev []any) bool {
		if !observeConfirmed(prev) {
			return false
		}
		if len(prev) <= 4 {
			return false
		}
		if err, isErr := prev[4].(error); isErr && err != nil {
			return false
		}
		return true
	}

	steps := []harness.Step{
		initflow.NewVesselStage(ctx),
		initflow.NewSoilStage(ctx, soilOptions),
		initflow.NewBranchesStage(ctx, cat, agentDef),
		initflow.NewObserveStage(ctx, cat, agentDef),
		harness.NewConditional(
			// Generate is wrapped in NewLazy so its action closure can capture
			// prev-results (Vessel / Soil / Branches). The initflow GenerateAction
			// signature is `func() error` — prev is injected via closure capture
			// at Lazy-build time, which fires before the inner stage's Init.
			harness.NewLazy("Generate", func(prev []any) harness.Step {
				action := buildGenerateAction(
					prev, cat, agentDef, techLeadType, cwd, configPath,
					lock, &wr, &cfg, &installed,
				)
				return initflow.NewGenerateStage(ctx, action)
			}),
			observeConfirmed,
		),
		harness.NewLazyGroup("Resolve conflicts", func(prev []any) []harness.Step {
			if !generateSucceeded(prev) {
				return nil
			}
			if !wr.HasConflicts() {
				return nil
			}
			return buildConflictSteps(&wr)
		}),
		harness.NewConditional(
			// Planted is wrapped in NewLazy so the ability counts rendered in
			// the SUMMARY panel reflect the live `installed` struct populated
			// by the Generate action, rather than agent-def defaults captured
			// at step-declaration time. The ctx.StationDir is also refreshed
			// from the Vessel result so SUMMARY + tree render the user's
			// chosen station path instead of the "station/" placeholder.
			harness.NewLazy("Planted", func(prev []any) harness.Step {
				plantedCtx := ctx
				if vessel, ok := prev[0].(map[string]string); ok {
					if s := strings.TrimSpace(vessel["station"]); s != "" {
						if !strings.HasSuffix(s, "/") {
							s += "/"
						}
						plantedCtx.StationDir = s
					}
				}
				summary := plantedSummary(installed)
				return initflow.NewPlantedStage(plantedCtx, &wr, summary)
			}),
			generateSucceeded,
		),
	}

	bannerLine := "BONSAI"
	if Version != "" && Version != "dev" {
		bannerLine = fmt.Sprintf("BONSAI v%s", Version)
	}

	results, err := harness.Run(bannerLine, "Initializing new project", steps)
	if err != nil {
		if errors.Is(err, harness.ErrAborted) {
			// Ctrl-C — cfg.Save runs inside the Generate closure, so an abort
			// before that point leaves no .bonsai.yaml on disk. An abort after
			// means partial files may exist, but the user's intent was to quit
			// cleanly and the lock file was not yet synced.
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

	// If the user cancelled at Observe, no writes happened and there is
	// nothing to clean up.
	if !observeConfirmed(results) {
		return nil
	}

	// Surface any Generate error post-harness. The in-frame stateError view
	// already showed the message; this is a safety belt in case the user
	// exits without acknowledging and for any later telemetry hooks.
	if len(results) > 4 {
		if errVal, isErr := results[4].(error); isErr && errVal != nil {
			tui.Warning("Generation error: " + errVal.Error())
			return nil
		}
	}

	// Apply conflict-picker selections. The LazyGroup splices the MultiSelect
	// + Confirm pair at index 5 when conflicts exist; applyConflictPicks
	// tolerates the slot being absent (spliced nothing) via its length check.
	applyConflictPicks(results, 5, &wr, lock, cwd)

	if err := lock.Save(cwd); err != nil {
		tui.Warning("Could not save lock file: " + err.Error())
	}

	return nil
}

// buildGenerateAction constructs the closure that the GenerateStage invokes
// on Init. Extracted as a free function so runInit stays readable — the
// body mirrors the legacy spinner action but reads prev indices mapped to
// the redesign's stage order (Vessel / Soil / Branches / Observe).
func buildGenerateAction(
	prev []any,
	cat *catalog.Catalog,
	agentDef *catalog.AgentDef,
	agentType string,
	cwd string,
	configPath string,
	lock *config.LockFile,
	wr *generate.WriteResult,
	cfgOut **config.ProjectConfig,
	installedOut **config.InstalledAgent,
) initflow.GenerateAction {
	return func() error {
		vessel, _ := prev[0].(map[string]string)
		soil, _ := prev[1].([]string)
		branches, _ := prev[2].(initflow.BranchesResult)

		projectName := strings.TrimSpace(vessel["name"])
		description := strings.TrimSpace(vessel["description"])
		// Vessel already normalised to a trailing slash — mirror legacy
		// behaviour so downstream callers get a path-shaped value even if a
		// future Vessel refactor relaxes that guarantee.
		docsPath := vessel["station"]
		if !strings.HasSuffix(docsPath, "/") {
			docsPath += "/"
		}

		installed := &config.InstalledAgent{
			AgentType: agentType,
			Workspace: docsPath,
			Skills:    append([]string(nil), branches.Skills...),
			Workflows: append([]string(nil), branches.Workflows...),
			Protocols: append([]string(nil), branches.Protocols...),
			Sensors:   append([]string(nil), branches.Sensors...),
			Routines:  append([]string(nil), branches.Routines...),
		}
		generate.EnsureRoutineCheckSensor(installed)

		cfg := &config.ProjectConfig{
			ProjectName: projectName,
			Description: description,
			DocsPath:    docsPath,
			Scaffolding: append([]string(nil), soil...),
			Agents:      map[string]*config.InstalledAgent{agentType: installed},
		}
		// Save .bonsai.yaml first — Scaffolding depends on it existing, and
		// early-returning here leaves no partial config on disk if Save fails.
		if err := cfg.Save(configPath); err != nil {
			return err
		}

		// Publish the populated structs to the outer scope so the Planted
		// stage can render correct counts and applyConflictPicks has the
		// correct lock reference.
		*cfgOut = cfg
		*installedOut = installed

		var errs []error
		errs = append(errs, generate.Scaffolding(cwd, cfg, cat, lock, wr, false))
		errs = append(errs, generate.AgentWorkspace(cwd, agentDef, installed, cfg, cat, lock, wr, false))
		errs = append(errs, generate.PathScopedRules(cwd, cfg, cat, lock, wr, false))
		errs = append(errs, generate.WorkflowSkills(cwd, cfg, cat, lock, wr, false))
		errs = append(errs, generate.SettingsJSON(cwd, cfg, cat, lock, wr, false))
		return errors.Join(errs...)
	}
}

// plantedSummary returns the ability counts rendered in the Planted stage's
// SUMMARY panel. Reads from the live InstalledAgent populated by the Generate
// closure so the numbers reflect post-routine-check ensure (which may add
// `routine-check` to the sensor list when any routines are installed).
func plantedSummary(installed *config.InstalledAgent) initflow.PlantedSummary {
	if installed == nil {
		return initflow.PlantedSummary{}
	}
	return initflow.PlantedSummary{
		Skills:    len(installed.Skills),
		Workflows: len(installed.Workflows),
		Protocols: len(installed.Protocols),
		Sensors:   len(installed.Sensors),
		Routines:  len(installed.Routines),
	}
}

// scaffoldingToSoilOptions maps catalog scaffolding entries into the
// initflow.ScaffoldingOption shape consumed by SoilStage.
func scaffoldingToSoilOptions(cat *catalog.Catalog) []initflow.ScaffoldingOption {
	out := make([]initflow.ScaffoldingOption, 0, len(cat.Scaffolding))
	for _, item := range cat.Scaffolding {
		desc := item.Description
		if !item.Required && item.Affects != "" {
			desc += " · if removed: " + item.Affects
		}
		out = append(out, initflow.ScaffoldingOption{
			Name:        item.Name,
			DisplayName: item.DisplayName,
			Description: desc,
			Required:    item.Required,
		})
	}
	return out
}
