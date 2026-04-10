package cmd

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/LastStep/Bonsai/internal/tui"
)

func init() {
	rootCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Show installed agents and their components.",
	RunE:  runList,
}

func runList(cmd *cobra.Command, args []string) error {
	cwd, _ := os.Getwd()
	configPath := filepath.Join(cwd, configFile)
	cfg, err := requireConfig(configPath)
	if err != nil {
		return err
	}

	if len(cfg.Agents) == 0 {
		tui.Blank()
		tui.EmptyPanel("No agents installed.\nRun bonsai add to get started.")
		return nil
	}

	cat := loadCatalog()
	tui.Heading(cfg.ProjectName)

	for name, agent := range cfg.Agents {
		displayName := name
		if agentDef := cat.GetAgent(name); agentDef != nil {
			displayName = agentDef.DisplayName
		}

		pairs := [][2]string{{"Workspace", agent.Workspace}}
		if len(agent.Skills) > 0 {
			pairs = append(pairs, [2]string{"Skills", strings.Join(agent.Skills, ", ")})
		}
		if len(agent.Workflows) > 0 {
			pairs = append(pairs, [2]string{"Workflows", strings.Join(agent.Workflows, ", ")})
		}
		if len(agent.Protocols) > 0 {
			pairs = append(pairs, [2]string{"Protocols", strings.Join(agent.Protocols, ", ")})
		}
		if len(agent.Sensors) > 0 {
			pairs = append(pairs, [2]string{"Sensors", strings.Join(agent.Sensors, ", ")})
		}
		if len(agent.Routines) > 0 {
			pairs = append(pairs, [2]string{"Routines", strings.Join(agent.Routines, ", ")})
		}

		content := tui.CardFields(pairs)
		tui.TitledPanel(displayName, content, tui.Water)
	}

	tui.Blank()
	return nil
}
