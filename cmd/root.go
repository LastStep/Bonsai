package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"github.com/LastStep/Bonsai/internal/catalog"
	"github.com/LastStep/Bonsai/internal/config"
	"github.com/LastStep/Bonsai/internal/generate"
	"github.com/LastStep/Bonsai/internal/tui"
)

const configFile = ".bonsai.yaml"

// Version is the current CLI version, set via SetVersion at startup.
var Version = "dev"

var catalogFS fs.FS
var guideMarkdown string

var rootCmd = &cobra.Command{
	Use:   "bonsai",
	Short: "Scaffold Claude Code agent workspaces for your project.",
}

func loadCatalog() *catalog.Catalog {
	cat, err := catalog.New(catalogFS)
	if err != nil {
		tui.ErrorPanel("Failed to load catalog: " + err.Error())
		os.Exit(1)
	}
	return cat
}

func requireConfig(configPath string) (*config.ProjectConfig, error) {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		tui.ErrorPanel("No " + configFile + " found.\nRun bonsai init first.")
		os.Exit(1)
	}
	return config.Load(configPath)
}

// SetVersion sets the version string on the root command.
func SetVersion(v string) {
	rootCmd.Version = v
}

// Execute is the main entry point for the CLI.
func Execute(fsys fs.FS, guide string) {
	catalogFS = fsys
	guideMarkdown = guide
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// resolveConflicts shows a multi-select picker for user-modified files.
// Users check the files they want to update (overwrite) and uncheck to skip.
func resolveConflicts(wr *generate.WriteResult, lock *config.LockFile, projectRoot string) {
	conflicts := wr.Conflicts()
	if len(conflicts) == 0 {
		return
	}

	tui.Blank()
	tui.Warning(fmt.Sprintf("%d file(s) modified since Bonsai generated them.", len(conflicts)))
	tui.Info("Select which files to update. Unchecked files keep your changes.")
	tui.Blank()

	// Build multi-select options — all pre-selected for update
	var options []huh.Option[string]
	for _, c := range conflicts {
		options = append(options, huh.NewOption(c.RelPath, c.RelPath).Selected(true))
	}

	selected, err := tui.AskMultiSelect("Update these files?", options)
	if err != nil || len(selected) == 0 {
		return // user cancelled or unchecked everything
	}

	// Offer backup for the selected files
	backup, err := tui.AskConfirm("Create .bak backups before overwriting?", false)
	if err != nil {
		return
	}
	if backup {
		for _, relPath := range selected {
			abs := filepath.Join(projectRoot, relPath)
			data, readErr := os.ReadFile(abs)
			if readErr == nil {
				_ = os.WriteFile(abs+".bak", data, 0644)
			}
		}
		tui.Info(fmt.Sprintf("Backed up %d file(s) with .bak extension.", len(selected)))
	}

	wr.ForceSelected(selected, projectRoot, lock)
}

// showWriteResults displays categorized file trees for generation outcomes.
func showWriteResults(wr *generate.WriteResult, rootLabel string) {
	// Normalize prefix for stripping (e.g. "station/" → "station")
	prefix := strings.TrimRight(rootLabel, "/")

	var created, updated, conflicted []string
	for _, f := range wr.Files {
		// Strip the workspace prefix so the tree doesn't double it
		rel := strings.TrimPrefix(f.RelPath, prefix+"/")
		switch f.Action {
		case generate.ActionCreated:
			created = append(created, rel)
		case generate.ActionUpdated, generate.ActionForced:
			updated = append(updated, rel)
		case generate.ActionConflict:
			conflicted = append(conflicted, rel)
		}
	}
	if rootLabel == "" {
		rootLabel = "."
	}
	if len(created) > 0 {
		tree := tui.FileTree(created, rootLabel)
		tui.TitledPanel("Created", tree, tui.Moss)
	}
	if len(updated) > 0 {
		tree := tui.FileTree(updated, rootLabel)
		tui.TitledPanel("Updated", tree, tui.Water)
	}
	if len(conflicted) > 0 {
		tree := tui.FileTree(conflicted, rootLabel)
		tui.TitledPanel("Skipped (user modified)", tree, tui.Amber)
	}
}
