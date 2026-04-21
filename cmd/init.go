package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/LastStep/Bonsai/internal/catalog"
	"github.com/LastStep/Bonsai/internal/config"
	"github.com/LastStep/Bonsai/internal/generate"
	"github.com/LastStep/Bonsai/internal/tui"
	"github.com/LastStep/Bonsai/internal/tui/harness"
)

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize Bonsai in the current project.",
	RunE:  runInit,
}

// stationDirValidator rejects empty / root paths up-front so the user can
// correct in place inside the harness rather than crashing after collection.
// Trim semantics intentionally match the post-collection normalisation below.
func stationDirValidator(s string) error {
	v := strings.TrimSpace(s)
	if v == "" || v == "/" {
		return fmt.Errorf("must be a subdirectory like: station/")
	}
	return nil
}

// normaliseDocsPath applies the same trim + trailing-slash rule that lived
// inline in the old runInit. Kept as a free function so the post-harness
// pipeline stays small.
func normaliseDocsPath(s string) string {
	v := strings.TrimSpace(s)
	if !strings.HasSuffix(v, "/") {
		v += "/"
	}
	return v
}

// scaffoldingOptions lifts the catalog scaffolding entries into the
// tui.ItemOption shape consumed by MultiSelectStep.
func scaffoldingOptions(cat *catalog.Catalog) []tui.ItemOption {
	out := make([]tui.ItemOption, 0, len(cat.Scaffolding))
	for _, item := range cat.Scaffolding {
		desc := item.Description
		if !item.Required && item.Affects != "" {
			desc += " · if removed: " + item.Affects
		}
		out = append(out, tui.ItemOption{
			Name:     item.DisplayName,
			Value:    item.Name,
			Desc:     desc,
			Required: item.Required,
		})
	}
	return out
}

// userSensorOptions filters out the auto-managed routine-check sensor so the
// user only picks from sensors they actually choose. Mirrors the inline filter
// in the pre-harness runInit.
func userSensorOptions(cat *catalog.Catalog, agentType string) []tui.ItemOption {
	available := cat.SensorsFor(agentType)
	filtered := make([]catalog.SensorItem, 0, len(available))
	for _, s := range available {
		if s.Name != "routine-check" {
			filtered = append(filtered, s)
		}
	}
	return toSensorOptions(filtered, agentType)
}

// asString safely extracts a string result from a harness step. Returns ""
// for nil to keep call-site logic short — empty input is already meaningful
// (e.g., optional Description field).
func asString(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

// asStringSlice safely extracts a []string result. Returns nil for absent
// results so downstream nil checks behave as before the harness migration.
func asStringSlice(v any) []string {
	if v == nil {
		return nil
	}
	if s, ok := v.([]string); ok {
		return s
	}
	return nil
}

// asBool safely extracts a bool result.
func asBool(v any) bool {
	if v == nil {
		return false
	}
	if b, ok := v.(bool); ok {
		return b
	}
	return false
}

func runInit(cmd *cobra.Command, args []string) error {
	// BONSAI_REDESIGN=1 routes to the in-progress cinematic flow (Plan 22).
	// Default behaviour is unchanged until the redesign ships Phase 5.
	if os.Getenv("BONSAI_REDESIGN") == "1" {
		return runInitRedesign(cmd, args)
	}

	cwd := mustCwd()
	configPath := filepath.Join(cwd, configFile)

	if _, err := os.Stat(configPath); err == nil {
		tui.WarningPanel(configFile + " already exists. Skipping init.")
		return nil
	}

	cat := loadCatalog()

	const techLeadType = "tech-lead"
	agentDef := cat.GetAgent(techLeadType)
	if agentDef == nil {
		tui.FatalPanel("Tech Lead agent not found", "The built-in catalog is missing the tech-lead agent.", "This is a bug — please report it.")
	}

	// Lock + WriteResult shared between the spinner closure and the conflict
	// picker. cfg + installed are populated by the spinner closure (which runs
	// inside the harness, gated by the review confirm).
	lock, _ := config.LoadLockFile(cwd)
	var wr generate.WriteResult
	var cfg *config.ProjectConfig
	var installed *config.InstalledAgent

	// Build the step stack. The review step is lazy so it can read every
	// prior selection without leaving AltScreen. The spinner + conflict picker
	// also live in-harness so the entire flow stays in AltScreen.
	steps := []harness.Step{
		harness.NewText("Project name", "Project name:", "", true),
		harness.NewText("Description", "Description (optional):", "", false),
		harness.NewText("Station directory", "Station directory:", "station/", true, stationDirValidator),
		harness.NewMultiSelect("Scaffolding", "Project Scaffolding", scaffoldingOptions(cat), nil),
		harness.NewMultiSelect("Skills", "Skills", toItemOptions(cat.SkillsFor(techLeadType), techLeadType), agentDef.DefaultSkills),
		harness.NewMultiSelect("Workflows", "Workflows", toItemOptions(cat.WorkflowsFor(techLeadType), techLeadType), agentDef.DefaultWorkflows),
		harness.NewMultiSelect("Protocols", "Protocols", toItemOptions(cat.ProtocolsFor(techLeadType), techLeadType), agentDef.DefaultProtocols),
		harness.NewMultiSelect("Sensors", "Sensors", userSensorOptions(cat, techLeadType), agentDef.DefaultSensors),
		harness.NewMultiSelect("Routines", "Routines", toRoutineOptions(cat.RoutinesFor(techLeadType), techLeadType), agentDef.DefaultRoutines),
		harness.NewLazy("Review", func(prev []any) harness.Step {
			workspace := normaliseDocsPath(asString(prev[2]))
			panel := buildReviewPanel(cat, agentDef, workspace, prev)
			return harness.NewReview("Review", panel, "Generate project?", true)
		}),
		// Spinner — gated on the review confirm bool. Builds cfg + installed
		// from prev[] and runs the generators. cfg.Save lives in here too so
		// a Ctrl-C before this point leaves no .bonsai.yaml on disk.
		harness.NewConditional(
			harness.NewSpinnerWithPrior("Generating", "Generating project files...",
				func(prev []any) error {
					projectName := asString(prev[0])
					description := asString(prev[1])
					docsPath := normaliseDocsPath(asString(prev[2]))
					selectedScaffolding := asStringSlice(prev[3])
					selectedSkills := asStringSlice(prev[4])
					selectedWorkflows := asStringSlice(prev[5])
					selectedProtocols := asStringSlice(prev[6])
					selectedSensors := asStringSlice(prev[7])
					selectedRoutines := asStringSlice(prev[8])

					installed = &config.InstalledAgent{
						AgentType: techLeadType,
						Workspace: docsPath,
						Skills:    selectedSkills,
						Workflows: selectedWorkflows,
						Protocols: selectedProtocols,
						Sensors:   selectedSensors,
						Routines:  selectedRoutines,
					}
					generate.EnsureRoutineCheckSensor(installed)

					cfg = &config.ProjectConfig{
						ProjectName: strings.TrimSpace(projectName),
						Description: strings.TrimSpace(description),
						DocsPath:    docsPath,
						Scaffolding: selectedScaffolding,
						Agents:      map[string]*config.InstalledAgent{techLeadType: installed},
					}
					// Save .bonsai.yaml first — Scaffolding depends on it
					// existing, and early-returning here leaves no partial
					// config on disk if Save fails.
					if err := cfg.Save(configPath); err != nil {
						return err
					}
					var errs []error
					errs = append(errs, generate.Scaffolding(cwd, cfg, cat, lock, &wr, false))
					errs = append(errs, generate.AgentWorkspace(cwd, agentDef, installed, cfg, cat, lock, &wr, false))
					errs = append(errs, generate.PathScopedRules(cwd, cfg, cat, lock, &wr, false))
					errs = append(errs, generate.WorkflowSkills(cwd, cfg, cat, lock, &wr, false))
					errs = append(errs, generate.SettingsJSON(cwd, cfg, cat, lock, &wr, false))
					return errors.Join(errs...)
				}),
			func(prev []any) bool {
				if len(prev) == 0 {
					return false
				}
				b, ok := prev[len(prev)-1].(bool)
				return ok && b
			},
		),
		harness.NewLazyGroup("Resolve conflicts", func(prev []any) []harness.Step {
			if !wr.HasConflicts() {
				return nil
			}
			return buildConflictSteps(&wr)
		}),
	}

	bannerLine := "BONSAI"
	if Version != "" && Version != "dev" {
		bannerLine = fmt.Sprintf("BONSAI v%s", Version)
	}

	results, err := harness.Run(bannerLine, "Initializing new project", steps)
	if err != nil {
		if errors.Is(err, harness.ErrAborted) {
			// Ctrl-C — exit cleanly with no .bonsai.yaml or partial files.
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

	// Review confirm lives at index 9 (the LazyStep wrapping the ReviewStep).
	if len(results) <= 9 || !asBool(results[9]) {
		return nil
	}

	// Spinner ran at index 10. Surface any error it returned.
	if len(results) > 10 {
		if errVal := results[10]; errVal != nil {
			if e, ok := errVal.(error); ok && e != nil {
				tui.Warning("Generation error: " + e.Error())
				return nil
			}
		}
	}

	// Conflict-picker LazyGroup at index 11, when spliced, produces the
	// MultiSelect at index 11 (LazyGroup is replaced in place).
	applyConflictPicks(results, 11, &wr, lock, cwd)

	if err := lock.Save(cwd); err != nil {
		tui.Warning("Could not save lock file: " + err.Error())
	}

	showWriteResults(&wr)

	if cfg != nil {
		tui.Success(fmt.Sprintf("Initialized %s with %s", cfg.ProjectName, agentDef.DisplayName))
	}
	tui.Hint("Next: run bonsai add to add code agents (backend, frontend, etc.).")
	tui.Blank()
	return nil
}

// buildReviewPanel composes the static review block shown by the lazy
// ReviewStep. Uses the string-returning TitledPanelString so the harness can
// render the full bordered panel inside AltScreen.
func buildReviewPanel(cat *catalog.Catalog, agentDef *catalog.AgentDef, workspace string, prev []any) string {
	// prev indices match the step declaration order in runInit.
	selectedSkills := asStringSlice(prev[4])
	selectedWorkflows := asStringSlice(prev[5])
	selectedProtocols := asStringSlice(prev[6])
	selectedSensors := asStringSlice(prev[7])
	selectedRoutines := asStringSlice(prev[8])

	tree := tui.ItemTree(
		tui.StyleLabel.Render(agentDef.DisplayName)+" "+tui.StyleMuted.Render(tui.GlyphArrow+" "+workspace),
		[]tui.Category{
			{Name: "Skills", Items: selectedSkills},
			{Name: "Workflows", Items: selectedWorkflows},
			{Name: "Protocols", Items: selectedProtocols},
			{Name: "Sensors", Items: selectedSensors},
			{Name: "Routines", Items: selectedRoutines},
		},
		newDescriber(cat),
	)

	return tui.TitledPanelString("Review", tree, tui.Water)
}
