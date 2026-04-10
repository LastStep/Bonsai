package cmd

import (
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh/spinner"
	"github.com/spf13/cobra"

	"github.com/LastStep/Bonsai/internal/generate"
	"github.com/LastStep/Bonsai/internal/tui"
)

func init() {
	rootCmd.AddCommand(removeCmd)
	removeCmd.Flags().BoolP("delete-files", "d", false, "Also delete the generated agent/ directory")
}

var removeCmd = &cobra.Command{
	Use:   "remove [agent]",
	Short: "Remove an installed agent from the project.",
	Args:  cobra.ExactArgs(1),
	RunE:  runRemove,
}

func runRemove(cmd *cobra.Command, args []string) error {
	agentName := args[0]
	cwd, _ := os.Getwd()
	configPath := filepath.Join(cwd, configFile)
	cfg, err := requireConfig(configPath)
	if err != nil {
		return err
	}

	agent, exists := cfg.Agents[agentName]
	if !exists {
		tui.Error("Agent " + agentName + " is not installed.")
		os.Exit(1)
	}

	preview := tui.ItemTree(
		tui.StyleLabel.Render(agentName)+" "+tui.StyleMuted.Render(tui.GlyphArrow+" "+agent.Workspace),
		[]tui.Category{
			{Name: "Skills", Items: agent.Skills},
			{Name: "Workflows", Items: agent.Workflows},
			{Name: "Protocols", Items: agent.Protocols},
			{Name: "Sensors", Items: agent.Sensors},
		},
		nil,
	)

	tui.TitledPanel("Remove", preview, tui.Amber)
	tui.Blank()

	confirmed, err := tui.AskConfirm("Remove "+agentName+"?", false)
	if err != nil || !confirmed {
		return nil
	}

	cat := loadCatalog()

	_ = spinner.New().
		Title("Removing agent...").
		Action(func() {
			delete(cfg.Agents, agentName)
			_ = cfg.Save(configPath)
			_ = generate.RootClaudeMD(cwd, cfg)
			_ = generate.SettingsJSON(cwd, cfg, cat)
		}).
		Run()

	deleteFiles, _ := cmd.Flags().GetBool("delete-files")
	if deleteFiles {
		agentDir := filepath.Join(cwd, agent.Workspace, "agent")
		claudeMD := filepath.Join(cwd, agent.Workspace, "CLAUDE.md")
		if err := os.RemoveAll(agentDir); err == nil {
			tui.Info("Deleted " + agentDir)
		}
		if err := os.Remove(claudeMD); err == nil {
			tui.Info("Deleted " + claudeMD)
		}
	}

	tui.Success("Removed " + agentName)
	return nil
}
