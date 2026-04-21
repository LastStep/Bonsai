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

// runInitRedesign is the Phase-2 stub entry point for Plan 22's cinematic
// `bonsai init` flow. It renders the persistent chrome (header + enso rail
// + footer) around four placeholder stages that advance on Enter and pop
// on Esc. No files are written — the generate + planted pipeline lands in
// Phases 3–5.
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
	// until Phase 3's VesselStage captures a user-entered value.
	ctx := initflow.StageContext{
		Version:      Version,
		ProjectDir:   cwd,
		StationDir:   "station/",
		AgentDisplay: agentDisplay,
		StartedAt:    startedAt,
	}

	steps := []harness.Step{
		initflow.NewStubStage(0, ctx.Version, ctx.ProjectDir, ctx.StationDir, ctx.AgentDisplay, ctx),
		initflow.NewStubStage(1, ctx.Version, ctx.ProjectDir, ctx.StationDir, ctx.AgentDisplay, ctx),
		initflow.NewStubStage(2, ctx.Version, ctx.ProjectDir, ctx.StationDir, ctx.AgentDisplay, ctx),
		initflow.NewStubStage(3, ctx.Version, ctx.ProjectDir, ctx.StationDir, ctx.AgentDisplay, ctx),
	}

	bannerLine := "BONSAI"
	if Version != "" && Version != "dev" {
		bannerLine = fmt.Sprintf("BONSAI v%s", Version)
	}

	_, err := harness.Run(bannerLine, "Initializing new project (redesign)", steps)
	if err != nil {
		if errors.Is(err, harness.ErrAborted) {
			// Ctrl-C — no config / files written in Phase 2 either way.
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

	// Phase 2 does not write files, save config, or run a conflict picker —
	// those wire up in Phases 3–5. Returning cleanly after the flow exits
	// AltScreen is the expected behaviour for this phase.
	return nil
}
