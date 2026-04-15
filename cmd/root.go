package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"github.com/LastStep/Bonsai/internal/catalog"
	"github.com/LastStep/Bonsai/internal/config"
	"github.com/LastStep/Bonsai/internal/generate"
	"github.com/LastStep/Bonsai/internal/tui"
)

const configFile = ".bonsai.yaml"

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

// Execute is the main entry point for the CLI.
func Execute(fsys fs.FS, guide string) {
	catalogFS = fsys
	guideMarkdown = guide
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// resolveConflicts shows conflict TUI and lets user choose how to handle modified files.
func resolveConflicts(wr *generate.WriteResult, lock *config.LockFile, projectRoot string) {
	conflicts := wr.Conflicts()
	if len(conflicts) == 0 {
		return
	}

	var paths []string
	for _, c := range conflicts {
		paths = append(paths, c.RelPath)
	}

	tree := tui.FileTree(paths, ".")
	tui.TitledPanel("Modified Files", tree, tui.Amber)
	tui.Blank()
	tui.Warning(fmt.Sprintf("%d file(s) have been modified since Bonsai generated them.", len(conflicts)))
	tui.Info("Overwriting will replace your changes.")

	options := []huh.Option[string]{
		huh.NewOption("Skip all "+tui.StyleMuted.Render(tui.GlyphDash+" keep my changes"), "skip"),
		huh.NewOption("Overwrite all "+tui.StyleMuted.Render(tui.GlyphDash+" use Bonsai's version"), "overwrite"),
		huh.NewOption("Back up & overwrite "+tui.StyleMuted.Render(tui.GlyphDash+" save .bak copies first"), "backup"),
	}

	choice, err := tui.AskSelect("How should Bonsai handle these files?", options)
	if err != nil {
		return // user cancelled
	}

	switch choice {
	case "skip":
		return

	case "backup":
		for _, c := range conflicts {
			abs := filepath.Join(projectRoot, c.RelPath)
			bakPath := abs + ".bak"
			data, readErr := os.ReadFile(abs)
			if readErr == nil {
				_ = os.WriteFile(bakPath, data, 0644)
			}
		}
		tui.Info(fmt.Sprintf("Backed up %d file(s) with .bak extension.", len(conflicts)))
		wr.ForceConflicts(projectRoot, lock)

	case "overwrite":
		wr.ForceConflicts(projectRoot, lock)
	}
}

// showWriteResults displays categorized file trees for generation outcomes.
func showWriteResults(wr *generate.WriteResult, rootLabel string) {
	var created, updated, conflicted []string
	for _, f := range wr.Files {
		switch f.Action {
		case generate.ActionCreated:
			created = append(created, f.RelPath)
		case generate.ActionUpdated, generate.ActionForced:
			updated = append(updated, f.RelPath)
		case generate.ActionConflict:
			conflicted = append(conflicted, f.RelPath)
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
