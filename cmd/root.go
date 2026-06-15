package cmd

import (
	"fmt"
	"io/fs"
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

const configFile = ".bonsai.yaml"

// Version is the current CLI version, set via SetVersion at startup.
var Version = "dev"

var catalogFS fs.FS
var guideContents map[string]string

var rootCmd = &cobra.Command{
	Use:   "bonsai",
	Short: "Scaffold Claude Code agent workspaces for your project.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if noColor, _ := cmd.Flags().GetBool("no-color"); noColor {
			tui.DisableColor()
		}
	},
}

func init() {
	rootCmd.PersistentFlags().Bool("no-color", false, "Disable colored output")
	// The explicit `completion` command lives in cmd/completion.go and
	// supersedes Cobra's auto-generated one. DisableDefaultCmd suppresses
	// the auto-generated default so AddCommand("completion") in
	// completion.go is the only registered child with that name.
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}

func loadCatalog() *catalog.Catalog {
	cat, err := catalog.New(catalogFS)
	if err != nil {
		tui.FatalPanel("Failed to load catalog", err.Error(), "This is a bug — please report it.")
	}
	return cat
}

func requireConfig(configPath string) (*config.ProjectConfig, error) {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		tui.FatalPanel("No "+configFile+" found", "This command requires an initialized project.", "Run: bonsai init")
	}
	return config.Load(configPath)
}

// mustCwd returns the current working directory or aborts with a structured error.
// Getwd can fail if the cwd was deleted or is otherwise unreadable; silently
// dropping that error produces a relative path in downstream writes that surfaces
// as a confusing "no such file or directory" message.
func mustCwd() string {
	cwd, err := os.Getwd()
	if err != nil || cwd == "" {
		detail := "Could not resolve current directory."
		if err != nil {
			detail = err.Error()
		}
		tui.FatalPanel("Cannot determine working directory", detail, "cd into a valid directory and retry.")
	}
	return cwd
}

// SetVersion sets the version string on the root command.
func SetVersion(v string) {
	Version = v
	rootCmd.Version = v
}

// Execute is the main entry point for the CLI.
func Execute(fsys fs.FS, guides map[string]string) {
	catalogFS = fsys
	guideContents = guides
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// buildConflictSteps returns the harness steps for the conflict-resolution
// picker. Returns nil when wr.HasConflicts() is false so callers can splice
// the result into a LazyGroup without a wrapper conditional.
//
// Layout:
//   - [0] MultiSelectStep — list of conflict paths, all pre-selected (the
//     user unchecks files they want to KEEP their changes on).
//   - [1] ConditionalStep wrapping a ConfirmStep — only asks about backups
//     when the user picked at least one file to overwrite.
//
// The harness only captures the picks. Apply them with applyConflictPicks
// in the post-harness block — wr.ForceSelected mutates lock state and the
// optional .bak writes happen there too.
func buildConflictSteps(wr *generate.WriteResult) []harness.Step {
	conflicts := wr.Conflicts()
	if len(conflicts) == 0 {
		return nil
	}

	available := make([]tui.ItemOption, 0, len(conflicts))
	defaults := make([]string, 0, len(conflicts))
	for _, c := range conflicts {
		available = append(available, tui.ItemOption{
			Name:  c.RelPath,
			Value: c.RelPath,
			Desc:  "modified since last generate",
		})
		defaults = append(defaults, c.RelPath)
	}

	return []harness.Step{
		harness.NewMultiSelect("Conflicts",
			fmt.Sprintf("%d file(s) modified since Bonsai generated them. Select which to update — unchecked files keep your changes.", len(conflicts)),
			available, defaults),
		harness.NewConditional(
			harness.NewConfirm("Backup", "Create .bak backups before overwriting?", false),
			func(prev []any) bool {
				if len(prev) == 0 {
					return false
				}
				picks := asStringSlice(prev[len(prev)-1])
				return len(picks) > 0
			},
		),
	}
}

// applyConflictPicks consumes the harness results from buildConflictSteps
// (the trailing two slots in the results slice) and runs the file mutations
// the legacy resolveConflicts did inline. confIdx is the index of the
// MultiSelectStep result in the results slice; backupIdx is confIdx+1.
//
// Tolerates the slot being absent (LazyGroup spliced nothing) by returning
// false, so callers can pass a sentinel index without checking length first.
//
// Backup-failure handling: when the .bak read OR write step fails for a given
// path, that path is dropped from the overwrite list and a single collected
// tui.Warning is emitted naming all dropped paths. This avoids silently
// overwriting the user's local edits without a recoverable backup.
func applyConflictPicks(results []any, confIdx int, wr *generate.WriteResult,
	lock *config.LockFile, projectRoot string) bool {
	if confIdx < 0 || confIdx >= len(results) {
		return false
	}
	selected := asStringSlice(results[confIdx])
	if len(selected) == 0 {
		return false
	}
	backupIdx := confIdx + 1
	backup := backupIdx < len(results) && asBool(results[backupIdx])
	if backup {
		dropped := make(map[string]bool)
		for _, relPath := range selected {
			abs := filepath.Join(projectRoot, relPath)
			data, readErr := os.ReadFile(abs)
			if readErr != nil {
				dropped[relPath] = true
				continue
			}
			if writeErr := os.WriteFile(abs+".bak", data, 0644); writeErr != nil {
				dropped[relPath] = true
				continue
			}
		}
		if len(dropped) > 0 {
			filtered := make([]string, 0, len(selected)-len(dropped))
			droppedList := make([]string, 0, len(dropped))
			for _, relPath := range selected {
				if dropped[relPath] {
					droppedList = append(droppedList, relPath)
					continue
				}
				filtered = append(filtered, relPath)
			}
			selected = filtered
			tui.Warning("Could not write backup for: " + strings.Join(droppedList, ", ") + " — original file left unchanged.")
		}
	}
	if len(selected) == 0 {
		return false
	}
	wr.ForceSelected(selected, projectRoot, lock)
	return true
}
