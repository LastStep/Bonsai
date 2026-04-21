package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/LastStep/Bonsai/internal/catalog"
	"github.com/LastStep/Bonsai/internal/tui"
	"github.com/LastStep/Bonsai/internal/tui/harness"
	"github.com/LastStep/Bonsai/internal/tui/initflow"
)

// runInitRedesign is the Phase-3 entry point for Plan 22's cinematic
// `bonsai init` flow. It renders the persistent chrome (header + enso rail
// + footer) around four stages — Vessel and Soil are real input stages;
// Branches and Observe remain stubs until Phases 4–5. No files are written
// yet; the generate + planted pipeline lands in Phase 5.
//
// Routing: runInit at cmd/init.go:121 branches here when BONSAI_REDESIGN=1.
// Without the env flag, the legacy flow runs unchanged.
func runInitRedesign(cmd *cobra.Command, args []string) error {
	// Capture the session start time once so every stage shares the same
	// origin — the Planted stage (Phase 5) computes ELAPSED from here.
	startedAt := time.Now()

	cwd := mustCwd()
	configPath := filepath.Join(cwd, configFile)

	// Early-exit with the legacy warning copy so both flows respond
	// identically to an existing config.
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
	// Observe stage (Phase 5) can always render AGENT row correctly.
	agentDisplay := agentDef.DisplayName
	if agentDisplay == "" {
		agentDisplay = catalog.DisplayNameFrom(agentDef.Name)
	}

	// Shared context stamped on every stage. Station defaults to "station/"
	// until VesselStage captures a user-entered value; Phase 5's Planted
	// stage will read the post-Vessel value out of prev[0] when rendering
	// the generated file tree.
	ctx := initflow.StageContext{
		Version:      Version,
		ProjectDir:   cwd,
		StationDir:   "station/",
		AgentDisplay: agentDisplay,
		StartedAt:    startedAt,
	}

	// Phase 3 wires real Vessel + Soil stages; Branches + Observe remain
	// stubs until Phases 4–5 land. Legacy generate / conflict tail is still
	// skipped — runInitRedesign does NOT write files yet.
	soilOptions := scaffoldingToSoilOptions(cat)

	steps := []harness.Step{
		initflow.NewVesselStage(ctx),
		initflow.NewSoilStage(ctx, soilOptions),
		initflow.NewBranchesStage(ctx, cat, agentDef),
		initflow.NewObserveStage(ctx, cat, agentDef),
	}

	bannerLine := "BONSAI"
	if Version != "" && Version != "dev" {
		bannerLine = fmt.Sprintf("BONSAI v%s", Version)
	}

	_, err := harness.Run(bannerLine, "Initializing new project (redesign)", steps)
	if err != nil {
		if errors.Is(err, harness.ErrAborted) {
			// Ctrl-C — no config / files written in Phase 3 either way.
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

	// Phase 3 does not write files, save config, or run a conflict picker —
	// the generate + planted pipeline wires up in Phases 4–5. Returning
	// cleanly after the flow exits AltScreen is the expected behaviour.
	return nil
}

// scaffoldingToSoilOptions maps catalog scaffolding entries into the
// initflow.ScaffoldingOption shape consumed by SoilStage. Parallels the
// legacy `scaffoldingOptions` helper (which returns tui.ItemOption for the
// MultiSelectStep) but keeps the redesign path decoupled from the tui
// package's option type.
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
