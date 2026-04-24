package cmd

import (
	"os"
	"path/filepath"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"

	"github.com/LastStep/Bonsai/internal/config"
	"github.com/LastStep/Bonsai/internal/tui"
	"github.com/LastStep/Bonsai/internal/tui/updateflow"
)

func init() {
	rootCmd.AddCommand(updateCmd)
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Sync workspace — detect custom files, re-render abilities, refresh CLAUDE.md.",
	RunE:  runUpdate,
}

// runUpdate is the cobra entry point. Thin wrapper over
// updateflow.Run — the cinematic port (Plan 31 Phase F) moved all
// interactive surface, discovery scanning, and the re-render pipeline
// into internal/tui/updateflow so this command stays tiny.
//
// Non-TTY callers (CI, piped stdin) are routed through updateflow.RunStatic
// which auto-accepts every valid discovery and returns errors for any
// conflicts rather than showing the interactive picker.
func runUpdate(cmd *cobra.Command, args []string) error {
	cwd := mustCwd()
	configPath := filepath.Join(cwd, configFile)
	cfg, err := requireConfig(configPath)
	if err != nil {
		return err
	}

	cat := loadCatalog()
	lock, _ := config.LoadLockFile(cwd)
	if lock == nil {
		lock = config.NewLockFile()
	}

	var result updateflow.Result
	if isTerminal() {
		result, err = updateflow.Run(cwd, cfg, cat, lock, Version)
	} else {
		result, err = updateflow.RunStatic(cwd, cfg, cat, lock, Version)
	}
	if err != nil {
		return err
	}
	if result.Cancelled {
		// User Ctrl-C'd before Sync — no persistence, no summary print.
		return nil
	}

	// Persist config when the user accepted at least one discovery.
	if result.ConfigChanged {
		if err := cfg.Save(configPath); err != nil {
			tui.Warning("Could not save config: " + err.Error())
		}
	}
	if err := lock.Save(cwd); err != nil {
		tui.Warning("Could not save lock file: " + err.Error())
	}

	// Non-TTY success line — the TTY path renders its own exit card
	// inside the cinematic Yield stage, so we only print here when the
	// flow ran without a terminal.
	if !isTerminal() {
		if result.SyncErr != nil {
			tui.Warning("Update error: " + result.SyncErr.Error())
			return nil
		}
		if result.WriteResult != nil {
			created, updated, _, _, conflicts := result.WriteResult.Summary()
			if created == 0 && updated == 0 && conflicts == 0 && !result.ConfigChanged {
				tui.Success("Workspace already in sync.")
				return nil
			}
			showWriteResults(result.WriteResult)
		}
		if result.ConfigChanged {
			tui.Success("Update complete — custom files tracked")
		} else {
			tui.Success("Update complete — workspace synced")
		}
	}
	return nil
}

// isTerminal reports whether stdin is a TTY. Used to pick between the
// cinematic interactive flow and the RunStatic fallback.
func isTerminal() bool {
	return isatty.IsTerminal(os.Stdin.Fd()) || isatty.IsCygwinTerminal(os.Stdin.Fd())
}
