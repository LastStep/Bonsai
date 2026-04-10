package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/spf13/cobra"

	"github.com/LastStep/Bonsai/internal/catalog"
	"github.com/LastStep/Bonsai/internal/config"
	"github.com/LastStep/Bonsai/internal/generate"
	"github.com/LastStep/Bonsai/internal/tui"
)

func init() {
	rootCmd.AddCommand(addCmd)
}

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add an agent to the project.",
	RunE:  runAdd,
}

func toItemOptions(items []catalog.CatalogItem) []tui.ItemOption {
	result := make([]tui.ItemOption, len(items))
	for i, item := range items {
		result[i] = tui.ItemOption{Name: item.Name, Desc: item.Description}
	}
	return result
}

func toSensorOptions(items []catalog.SensorItem) []tui.ItemOption {
	result := make([]tui.ItemOption, len(items))
	for i, item := range items {
		result[i] = tui.ItemOption{Name: item.Name, Desc: item.Description}
	}
	return result
}

func runAdd(cmd *cobra.Command, args []string) error {
	cwd, _ := os.Getwd()
	configPath := filepath.Join(cwd, configFile)
	cfg, err := requireConfig(configPath)
	if err != nil {
		return err
	}

	cat := loadCatalog()

	tui.Heading("Add Agent")

	// 1. Pick agent type
	var agentOptions []huh.Option[string]
	for _, a := range cat.Agents {
		agentOptions = append(agentOptions,
			huh.NewOption(a.DisplayName+" "+tui.StyleMuted.Render(tui.GlyphDash+" "+a.Description), a.Name))
	}
	agentType, err := tui.AskSelect("Agent type:", agentOptions)
	if err != nil {
		return err
	}

	if _, exists := cfg.Agents[agentType]; exists {
		tui.WarningPanel("Agent " + agentType + " is already installed.")
		return nil
	}

	agentDef := cat.GetAgent(agentType)
	if agentDef == nil {
		tui.Error("Unknown agent type: " + agentType)
		os.Exit(1)
	}

	// 2. Workspace directory
	existingWorkspaces := make(map[string]bool)
	for _, a := range cfg.Agents {
		existingWorkspaces[a.Workspace] = true
	}

	workspace, err := tui.AskText("Workspace directory (e.g. backend/):", agentType+"/", true)
	if err != nil {
		return err
	}
	workspace = strings.TrimSpace(workspace)
	workspace = strings.TrimRight(workspace, "/") + "/"

	if existingWorkspaces[workspace] {
		tui.ErrorPanel("Workspace " + workspace + " is already in use.")
		os.Exit(1)
	}

	// 3. Pick components
	selectedSkills, err := tui.PickItems("Skills", toItemOptions(cat.SkillsFor(agentType)), agentDef.DefaultSkills)
	if err != nil {
		return err
	}
	selectedWorkflows, err := tui.PickItems("Workflows", toItemOptions(cat.WorkflowsFor(agentType)), agentDef.DefaultWorkflows)
	if err != nil {
		return err
	}
	selectedProtocols, err := tui.PickItems("Protocols", toItemOptions(cat.ProtocolsFor(agentType)), agentDef.DefaultProtocols)
	if err != nil {
		return err
	}
	selectedSensors, err := tui.PickItems("Sensors", toSensorOptions(cat.SensorsFor(agentType)), agentDef.DefaultSensors)
	if err != nil {
		return err
	}

	// 4. Review summary
	describer := func(name string) string {
		if item := cat.GetItem(name); item != nil {
			return item.Description
		}
		if sensor := cat.GetSensor(name); sensor != nil {
			return sensor.Description
		}
		return ""
	}

	summary := tui.ItemTree(
		tui.StyleLabel.Render(agentDef.DisplayName)+" "+tui.StyleMuted.Render(tui.GlyphArrow+" "+workspace),
		[]tui.Category{
			{Name: "Skills", Items: selectedSkills},
			{Name: "Workflows", Items: selectedWorkflows},
			{Name: "Protocols", Items: selectedProtocols},
			{Name: "Sensors", Items: selectedSensors},
		},
		describer,
	)

	tui.TitledPanel("Review", summary, tui.Water)
	tui.Blank()

	confirmed, err := tui.AskConfirm("Generate files?", true)
	if err != nil || !confirmed {
		return nil
	}

	// 5. Generate
	installed := &config.InstalledAgent{
		AgentType: agentType,
		Workspace: workspace,
		Skills:    selectedSkills,
		Workflows: selectedWorkflows,
		Protocols: selectedProtocols,
		Sensors:   selectedSensors,
	}
	cfg.Agents[agentType] = installed
	if err := cfg.Save(configPath); err != nil {
		return err
	}

	_ = spinner.New().
		Title("Generating workspace...").
		Action(func() {
			_ = generate.AgentWorkspace(cwd, agentDef, installed, cfg, cat)
			_ = generate.RootClaudeMD(cwd, cfg)
			_ = generate.SettingsJSON(cwd, cfg, cat)
		}).
		Run()

	tui.Success(fmt.Sprintf("Added %s at %s", agentDef.DisplayName, workspace))
	return nil
}
