package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/LastStep/Bonsai/internal/tui/listflow"
)

func init() {
	rootCmd.AddCommand(listCmd)
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
func runList(cmd *cobra.Command, args []string) error {
	cwd := mustCwd()
	configPath := filepath.Join(cwd, configFile)
	cfg, err := requireConfig(configPath)
	if err != nil {
		return err
	}
	cat := loadCatalog()

	termW, termH := terminalSize()
	fmt.Print(listflow.RenderAll(cfg, cat, Version, cwd, termW, termH))
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
