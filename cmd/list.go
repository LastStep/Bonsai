package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/LastStep/Bonsai/internal/catalog"
	"github.com/LastStep/Bonsai/internal/config"
	"github.com/LastStep/Bonsai/internal/generate"
	"github.com/LastStep/Bonsai/internal/tui/listflow"
)

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().Bool("json", false, "Output installed agents as JSON (agent-consumable, non-interactive)")
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Show installed agents and their components.",
	RunE:  runList,
}

// runList renders the cinematic `bonsai list` surface — a single
// fmt.Print of listflow.RenderAll's pure-function output. The TUI
// pipeline is static (no BubbleTea) so piped invocations produce
// clean non-ANSI output via the existing tui.DisableColor() pathway
// in internal/tui/styles.go.
//
// Plan 41 Phase 4: the --json flag short-circuits the cinematic path and
// emits a single-document ListSnapshot to stdout (mirroring catalog --json
// at cmd/catalog.go:43 and validate --json), so an AI agent / CI script can
// read the installed state non-interactively. The serializer lives in the
// TUI-free internal/generate package (beside SerializeCatalog).
func runList(cmd *cobra.Command, args []string) error {
	cwd := mustCwd()
	configPath := filepath.Join(cwd, configFile)
	cfg, err := requireConfig(configPath)
	if err != nil {
		return err
	}
	cat := loadCatalog()

	// cmd may be nil when runList is invoked directly from a unit test;
	// treat that as the default (cinematic) path.
	if cmd != nil {
		if jsonOut, _ := cmd.Flags().GetBool("json"); jsonOut {
			return renderListJSON(cfg, cat, cwd)
		}
	}

	termW, termH := terminalSize()
	fmt.Print(listflow.RenderAll(cfg, cat, Version, cwd, termW, termH))
	return nil
}

// renderListJSON serializes the installed-agent state to stdout as indent-2
// JSON, reusing generate.SerializeJSON (single source of truth for the
// ListSnapshot contract). Matches the catalog --json / validate --json output
// style: marshaled JSON, stdout-only, exit 0.
func renderListJSON(cfg *config.ProjectConfig, cat *catalog.Catalog, cwd string) error {
	data, err := generate.SerializeJSON(cfg, cat, Version, cwd)
	if err != nil {
		return fmt.Errorf("serialize list: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

// terminalSize queries the live terminal dims for the list renderer.
// Falls back to 80×24 when stdout is not a TTY or the syscall errors —
// `bonsai list` must produce readable piped output, so a non-TTY is a
// valid case (the initflow min-size check won't trigger at 80×24).
func terminalSize() (int, int) {
	w, h, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || w <= 0 || h <= 0 {
		return 80, 24
	}
	return w, h
}
