package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/LastStep/Bonsai/internal/catalog"
	"github.com/LastStep/Bonsai/internal/config"
	"github.com/LastStep/Bonsai/internal/generate"
	"github.com/LastStep/Bonsai/internal/nonint"
	"github.com/LastStep/Bonsai/internal/tui"
	"github.com/LastStep/Bonsai/internal/tui/harness"
	"github.com/LastStep/Bonsai/internal/tui/hints"
	"github.com/LastStep/Bonsai/internal/tui/initflow"
)

// techLeadInitKey is the agent map key used by the non-interactive branch's
// presence check. Mirrors the `techLeadType` constant inside the function
// body — duplicated as a file-scope constant so the early branch can read
// it without depending on later locals.
const techLeadInitKey = "tech-lead"

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

	// Non-interactive branch — Plan 39 §B. Delegated to runInitNonInteractive
	// so the unit test (cmd/init_nonint_test.go) can drive the same code
	// path without spawning a subprocess to observe os.Exit.
	if initNonInteractive || initFromConfig != "" {
		code, err := runInitNonInteractive(cwd, configPath, initNonInteractive, initFromConfig, os.Stdout, os.Stderr)
		if err != nil && code == 0 {
			// Either-alone usage error — let cobra surface the message with
			// its usage block. The contract: non-zero codes go through
			// os.Exit (skipping the cobra "Error:" prefix + usage banner);
			// a zero code with non-nil err is the soft-fail path where the
			// caller still expects a tidy "Error: ..." line.
			return err
		}
		if code != nonint.ExitOK {
			os.Exit(code)
		}
		return nil
	}

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

	// Shared context stamped on every stage. HeaderAction + HeaderRightLabel
	// are passed explicitly so the per-command chrome reads correctly even
	// as Plan 28's signature extension rolls out across sibling flows.
	ctx := initflow.StageContext{
		Version:          Version,
		ProjectDir:       cwd,
		StationDir:       "station/",
		AgentDisplay:     agentDisplay,
		StartedAt:        startedAt,
		HeaderAction:     "INIT",
		HeaderRightLabel: "PLANTING INTO",
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
				docsPath := "station/"
				if vessel, ok := prev[0].(map[string]string); ok {
					if s := strings.TrimSpace(vessel["station"]); s != "" {
						if !strings.HasSuffix(s, "/") {
							s += "/"
						}
						plantedCtx.StationDir = s
						docsPath = s
					}
				}
				summary := plantedSummary(installed)
				stage := initflow.NewPlantedStage(plantedCtx, &wr, summary)
				// Plan 31 Phase H — render the 3-layer hints block for the
				// tech-lead agent (init always installs tech-lead as its
				// primary agent).
				projectName := ""
				if vessel, ok := prev[0].(map[string]string); ok {
					projectName = vessel["name"]
				}
				block, _ := hints.Load(cat, techLeadType, "init", hints.TemplateContext{
					DocsPath:    docsPath,
					AgentName:   techLeadType,
					ProjectName: projectName,
				})
				stage.SetHintBlock(hints.Render(block, initflow.PanelContentWidth))
				return stage
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

	// Apply conflict-picker selections. The LazyGroup splices the MultiSelect
	// + Confirm pair at index 5 when conflicts exist; applyConflictPicks
	// tolerates the slot being absent (spliced nothing) via its length check.
	applyConflictPicks(results, 5, &wr, lock, cwd)

	if err := lock.Save(cwd); err != nil {
		tui.Warning("Could not save lock file: " + err.Error())
	}

	return nil
}

// runInitNonInteractive is the headless `bonsai init` driver. Returns
// (exitCode, error):
//
//   - (0,   non-nil err) — usage error: only one flag was set. Caller
//     surfaces err via cobra's RunE so the user sees the "Error:" prefix
//     and the usage block.
//   - (>0,  nil err)     — runner reported a non-zero exit code; caller
//     calls os.Exit(code) and the stderr message was already written.
//   - (>0,  non-nil err) — runner reported a non-zero exit code AND a
//     diagnostic message that the test can inspect. Wire-equivalent to
//     above; both fields populated so tests can assert on err.Error().
//   - (0,   nil err)     — success.
//
// The split lets the cobra RunE path use cobra's usage-printer for genuine
// usage mistakes while routing operational errors through os.Exit so the
// JSONL stdout stream isn't polluted with cobra's Error: prefix.
func runInitNonInteractive(cwd, configPath string, nonInt bool, fromConfig string, stdout, stderr io.Writer) (int, error) {
	if !nonInt || fromConfig == "" {
		return 0, fmt.Errorf("--non-interactive and --from-config must be set together")
	}
	cat := loadCatalog()
	cfg, err := nonint.LoadConfig(fromConfig, cwd, cat)
	if err != nil {
		_, _ = fmt.Fprintln(stderr, err)
		return nonint.ExitInvalidConfig, err
	}
	// Plan 39 Locked Decision §1: init requires a tech-lead entry. The
	// guard lives here (not inside RunInit) so the error message is
	// init-specific — RunInit defends in depth with the same check.
	if tl, ok := cfg.Agents[techLeadInitKey]; !ok || tl == nil {
		msg := "from-config: bonsai init requires a 'tech-lead' entry under agents:"
		_, _ = fmt.Fprintln(stderr, msg)
		return nonint.ExitInvalidConfig, fmt.Errorf("%s", msg)
	}
	// Plan 39 Locked Decision §1 (exclusivity): `bonsai init` installs only
	// the tech-lead agent — extra entries would be silently dropped by the
	// runner's tech-lead-only AgentWorkspace call, leaving them registered
	// in .bonsai.yaml and settings.json without a workspace materialised.
	// Reject up front; callers can chain `bonsai add` for additional agents.
	if got := len(cfg.Agents); got != 1 {
		msg := fmt.Sprintf("from-config: bonsai init accepts only a single 'tech-lead' entry under agents:, got %d agents (use `bonsai add` for additional agents after init)", got)
		_, _ = fmt.Fprintln(stderr, msg)
		return nonint.ExitInvalidConfig, fmt.Errorf("%s", msg)
	}
	result, code, runErr := nonint.RunInit(cwd, configPath, cfg, cat, Version)
	if runErr != nil {
		_, _ = fmt.Fprintln(stderr, runErr)
		return code, runErr
	}
	// Data → stdout (pure JSONL); warnings → stderr (plain text). This stream
	// split is a tested invariant — see cmd/init_nonint_test.go.
	if err := nonint.EmitJSONL(stdout, result); err != nil {
		_, _ = fmt.Fprintln(stderr, err)
		return nonint.ExitRuntime, err
	}
	for _, warn := range result.Warnings {
		_, _ = fmt.Fprintln(stderr, "warning:", warn)
	}
	return code, nil
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
		// Plan 31 Phase C: write .bonsai/catalog.json — filesystem-discoverable
		// catalog listing for agent consumption (pi-style convention).
		errs = append(errs, generate.WriteCatalogSnapshot(cwd, Version, cat, wr))
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
