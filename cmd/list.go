package cmd

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/LastStep/Bonsai/internal/catalog"
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

	tui.Heading(cfg.ProjectName)

	if len(cfg.Scaffolding) > 0 {
		scaffoldDisplay := make([]string, len(cfg.Scaffolding))
		for i, s := range cfg.Scaffolding {
			scaffoldDisplay[i] = catalog.DisplayNameFrom(s)
		}
		tui.Info("Scaffolding: " + strings.Join(scaffoldDisplay, ", "))
	}

	if len(cfg.Agents) == 0 {
		tui.Blank()
		tui.EmptyPanel("No agents installed.\nRun bonsai add to get started.")
		return nil
	}

	cat := loadCatalog()

	displayNames := func(names []string, lookup func(string) string) string {
		display := make([]string, len(names))
		for i, n := range names {
			if dn := lookup(n); dn != "" {
				display[i] = dn
			} else {
				display[i] = catalog.DisplayNameFrom(n)
			}
		}
		return strings.Join(display, ", ")
	}

	for name, agent := range cfg.Agents {
		displayName := name
		if agentDef := cat.GetAgent(name); agentDef != nil {
			displayName = agentDef.DisplayName
		}

		pairs := [][2]string{{"Workspace", agent.Workspace}}
		if len(agent.Skills) > 0 {
			pairs = append(pairs, [2]string{"Skills", displayNames(agent.Skills, func(n string) string {
				if s := cat.GetSkill(n); s != nil { return s.DisplayName }; return ""
			})})
		}
		if len(agent.Workflows) > 0 {
			pairs = append(pairs, [2]string{"Workflows", displayNames(agent.Workflows, func(n string) string {
				if w := cat.GetWorkflow(n); w != nil { return w.DisplayName }; return ""
			})})
		}
		if len(agent.Protocols) > 0 {
			pairs = append(pairs, [2]string{"Protocols", displayNames(agent.Protocols, func(n string) string {
				if p := cat.GetProtocol(n); p != nil { return p.DisplayName }; return ""
			})})
		}
		if len(agent.Sensors) > 0 {
			pairs = append(pairs, [2]string{"Sensors", displayNames(agent.Sensors, func(n string) string {
				if s := cat.GetSensor(n); s != nil { return s.DisplayName }; return ""
			})})
		}
		if len(agent.Routines) > 0 {
			pairs = append(pairs, [2]string{"Routines", displayNames(agent.Routines, func(n string) string {
				if r := cat.GetRoutine(n); r != nil { return r.DisplayName }; return ""
			})})
		}

		content := tui.CardFields(pairs)
		tui.TitledPanel(displayName, content, tui.Water)
	}

	tui.Blank()
	return nil
}
