package cmd

import (
	"io/fs"
	"os"

	"github.com/LastStep/Bonsai/internal/catalog"
	"github.com/LastStep/Bonsai/internal/config"
	"github.com/LastStep/Bonsai/internal/tui"
	"github.com/spf13/cobra"
)

const configFile = ".bonsai.yaml"

var catalogFS fs.FS

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
func Execute(fsys fs.FS) {
	catalogFS = fsys
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
