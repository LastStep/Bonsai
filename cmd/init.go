package cmd

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh/spinner"
	"github.com/spf13/cobra"

	"github.com/LastStep/Bonsai/internal/config"
	"github.com/LastStep/Bonsai/internal/generate"
	"github.com/LastStep/Bonsai/internal/tui"
)

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize Bonsai in the current project.",
	RunE:  runInit,
}

func runInit(cmd *cobra.Command, args []string) error {
	cwd, _ := os.Getwd()
	configPath := filepath.Join(cwd, configFile)

	if _, err := os.Stat(configPath); err == nil {
		tui.WarningPanel(configFile + " already exists. Skipping init.")
		return nil
	}

	tui.Banner()
	tui.Heading("Initialize Project")

	projectName, err := tui.AskText("Project name:", "", true)
	if err != nil {
		return err
	}
	description, err := tui.AskText("Description (optional):", "", false)
	if err != nil {
		return err
	}
	docsPath, err := tui.AskText("Docs directory (blank for root, e.g. 'docs/'):", "", false)
	if err != nil {
		return err
	}
	docsPath = strings.TrimSpace(docsPath)
	if docsPath != "" && !strings.HasSuffix(docsPath, "/") {
		docsPath += "/"
	}

	cfg := &config.ProjectConfig{
		ProjectName: strings.TrimSpace(projectName),
		Description: strings.TrimSpace(description),
		DocsPath:    docsPath,
		Agents:      make(map[string]*config.InstalledAgent),
	}

	if err := cfg.Save(configPath); err != nil {
		return err
	}

	var created []string
	_ = spinner.New().
		Title("Generating project files...").
		Action(func() {
			_ = generate.RootClaudeMD(cwd, cfg)
			created, _ = generate.Scaffolding(cwd, cfg, catalogFS)
		}).
		Run()

	if len(created) > 0 {
		root := docsPath
		if root == "" {
			root = "."
		}
		tree := tui.FileTree(created, root)
		tui.TitledPanel("Created Files", tree, tui.Moss)
	}

	tui.Success("Initialized " + cfg.ProjectName)
	tui.Hint("Next: run bonsai add to add an agent.")
	tui.Blank()
	return nil
}
