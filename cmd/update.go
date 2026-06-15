package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"

	"github.com/LastStep/Bonsai/internal/catalog"
	"github.com/LastStep/Bonsai/internal/config"
	"github.com/LastStep/Bonsai/internal/nonint"
	"github.com/LastStep/Bonsai/internal/tui"
	"github.com/LastStep/Bonsai/internal/tui/updateflow"
)

// Headless flags for `bonsai update`. --non-interactive forces the headless
// JSONL core even on a TTY; --skip-conflicts turns unresolved file conflicts
// from a hard stop (exit 5) into a counted skip (exit 0). No --json — headless
// mode always emits JSONL (Plan 41 flag-surface decision).
var (
	updateNonInteractive bool
	updateSkipConflicts  bool
)

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.Flags().BoolVar(&updateNonInteractive, "non-interactive", false,
		"Skip the cinematic flow; emit JSONL and exit codes (auto-enabled when stdin is not a TTY)")
	updateCmd.Flags().BoolVar(&updateSkipConflicts, "skip-conflicts", false,
		"Skip (and count) files that conflict with user edits instead of exiting 5")
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Sync workspace — detect custom files, re-render abilities, refresh CLAUDE.md.",
	RunE:  runUpdate,
	// SilenceUsage keeps cobra's usage block off stderr when the headless
	// adapter returns an operational error — stdout stays pure JSONL.
	SilenceUsage: true,
}

// runUpdate is the cobra entry point. It routes between two surfaces:
//
//   - Headless (--non-interactive OR stdin is not a TTY): nonint.RunUpdate +
//     EmitJSONL on stdout + warnings on stderr + os.Exit(code). stdout is pure
//     JSONL protocol; the explicit flag forces this even on a TTY.
//   - Cinematic (interactive TTY, no flag): updateflow.Run — unchanged.
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

	// Headless gate: explicit flag OR no TTY on stdin.
	if updateNonInteractive || !isTerminal() {
		code := runUpdateNonInteractive(cwd, cfg, cat, lock, updateSkipConflicts, os.Stdout, os.Stderr)
		os.Exit(code)
	}

	// Cinematic path — unchanged. The Yield stage renders its own exit card,
	// so persistence happens here based on the returned flow-control Result.
	result, err := updateflow.Run(cwd, cfg, cat, lock, Version)
	if err != nil {
		return err
	}
	if result.Cancelled {
		// User Ctrl-C'd before Sync — no persistence, no summary print.
		return nil
	}
	if result.ConfigChanged {
		if err := cfg.Save(configPath); err != nil {
			tui.Warning("Could not save config: " + err.Error())
		}
	}
	if err := lock.Save(cwd); err != nil {
		tui.Warning("Could not save lock file: " + err.Error())
	}
	return nil
}

// runUpdateNonInteractive is the headless `bonsai update` adapter. It calls
// the pure core, serialises the Result to JSONL on stdout, routes warnings to
// stderr as plain text, and returns the exit code for the caller to os.Exit.
// Split out (with injectable writers) so cmd/update_nonint_test.go can drive
// it without trapping os.Exit and assert stream separation.
//
// On a runner error the diagnostic is written to stderr and stdout stays
// empty — never partial JSONL.
func runUpdateNonInteractive(cwd string, cfg *config.ProjectConfig, cat *catalog.Catalog,
	lock *config.LockFile, skipConflicts bool, stdout, stderr io.Writer) int {
	result, code, runErr := nonint.RunUpdate(cwd, cfg, cat, lock, Version, skipConflicts)
	if runErr != nil {
		_, _ = fmt.Fprintln(stderr, runErr)
		return code
	}
	// Data → stdout (pure JSONL); warnings → stderr (plain text). This stream
	// split is a tested invariant — see cmd/update_nonint_test.go.
	if err := nonint.EmitJSONL(stdout, result); err != nil {
		_, _ = fmt.Fprintln(stderr, err)
		return nonint.ExitRuntime
	}
	for _, warn := range result.Warnings {
		_, _ = fmt.Fprintln(stderr, "warning:", warn)
	}
	return code
}

// isTerminal reports whether stdin is a TTY. Used to pick between the
// cinematic interactive flow and the headless JSONL fallback.
func isTerminal() bool {
	return isatty.IsTerminal(os.Stdin.Fd()) || isatty.IsCygwinTerminal(os.Stdin.Fd())
}
