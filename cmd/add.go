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

func toItemOptions(items []catalog.CatalogItem, agentType string) []tui.ItemOption {
	result := make([]tui.ItemOption, len(items))
	for i, item := range items {
		result[i] = tui.ItemOption{Name: item.DisplayName, Value: item.Name, Desc: item.Description, Required: item.Required.CompatibleWith(agentType)}
	}
	return result
}

func toSensorOptions(items []catalog.SensorItem, agentType string) []tui.ItemOption {
	result := make([]tui.ItemOption, len(items))
	for i, item := range items {
		result[i] = tui.ItemOption{Name: item.DisplayName, Value: item.Name, Desc: item.Description, Required: item.Required.CompatibleWith(agentType)}
	}
	return result
}

func toRoutineOptions(items []catalog.RoutineItem, agentType string) []tui.ItemOption {
	result := make([]tui.ItemOption, len(items))
	for i, item := range items {
		desc := item.Description
		if item.Frequency != "" {
			desc += " (every " + item.Frequency + ")"
		}
		result[i] = tui.ItemOption{Name: item.DisplayName, Value: item.Name, Desc: desc, Required: item.Required.CompatibleWith(agentType)}
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

	// If agent is already installed, pivot to "add items" flow
	if existing, exists := cfg.Agents[agentType]; exists {
		return runAddItems(cwd, configPath, cfg, cat, agentType, existing)
	}

	// Require tech-lead before adding other agents
	if agentType != "tech-lead" {
		if _, hasTechLead := cfg.Agents["tech-lead"]; !hasTechLead {
			tui.ErrorDetail("Tech Lead required", "No tech-lead agent is installed yet.", "Run: bonsai init")
			return nil
		}
	}

	agentDef := cat.GetAgent(agentType)
	if agentDef == nil {
		tui.FatalPanel("Unknown agent type", agentType+" is not in the catalog.", "Run: bonsai catalog")
	}

	// 2. Workspace directory
	existingWorkspaces := make(map[string]bool)
	for _, a := range cfg.Agents {
		existingWorkspaces[a.Workspace] = true
	}

	var workspace string
	if agentType == "tech-lead" {
		// Tech-lead always lives in the station directory
		workspace = cfg.DocsPath
		if workspace == "" {
			workspace = "station/"
		}
		tui.Info("Tech Lead workspace: " + workspace)
	} else {
		workspace, err = tui.AskText("Workspace directory (e.g. backend/):", agentType+"/", true)
		if err != nil {
			return err
		}
		workspace = strings.TrimSpace(workspace)
		workspace = strings.TrimRight(workspace, "/") + "/"
	}

	if existingWorkspaces[workspace] {
		tui.FatalPanel("Workspace conflict", workspace+" is already in use by another agent.", "Choose a different directory.")
	}

	// 3. Pick components
	selectedSkills, err := tui.PickItems("Skills", toItemOptions(cat.SkillsFor(agentType), agentType), agentDef.DefaultSkills)
	if err != nil {
		return err
	}
	selectedWorkflows, err := tui.PickItems("Workflows", toItemOptions(cat.WorkflowsFor(agentType), agentType), agentDef.DefaultWorkflows)
	if err != nil {
		return err
	}
	selectedProtocols, err := tui.PickItems("Protocols", toItemOptions(cat.ProtocolsFor(agentType), agentType), agentDef.DefaultProtocols)
	if err != nil {
		return err
	}
	// Filter out auto-managed sensors (routine-check is added/removed automatically)
	availableSensors := cat.SensorsFor(agentType)
	var userSensors []catalog.SensorItem
	for _, s := range availableSensors {
		if s.Name != "routine-check" {
			userSensors = append(userSensors, s)
		}
	}
	selectedSensors, err := tui.PickItems("Sensors", toSensorOptions(userSensors, agentType), agentDef.DefaultSensors)
	if err != nil {
		return err
	}
	selectedRoutines, err := tui.PickItems("Routines", toRoutineOptions(cat.RoutinesFor(agentType), agentType), agentDef.DefaultRoutines)
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
		if routine := cat.GetRoutine(name); routine != nil {
			return routine.Description
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
			{Name: "Routines", Items: selectedRoutines},
		},
		describer,
	)

	tui.TitledPanel("Review", summary, tui.Water)

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
		Routines:  selectedRoutines,
	}
	generate.EnsureRoutineCheckSensor(installed)
	cfg.Agents[agentType] = installed
	if err := cfg.Save(configPath); err != nil {
		return err
	}

	lock, _ := config.LoadLockFile(cwd)
	var wr generate.WriteResult

	_ = spinner.New().
		Title("Generating workspace...").
		Action(func() {
			_ = generate.AgentWorkspace(cwd, agentDef, installed, cfg, cat, lock, &wr, false)
			_ = generate.PathScopedRules(cwd, cfg, cat, lock, &wr, false)
			_ = generate.WorkflowSkills(cwd, cfg, cat, lock, &wr, false)
			_ = generate.SettingsJSON(cwd, cfg, cat, lock, &wr, false)
		}).
		Run()

	if wr.HasConflicts() {
		resolveConflicts(&wr, lock, cwd)
	}

	if err := lock.Save(cwd); err != nil {
		tui.Warning("Could not save lock file: " + err.Error())
	}

	showWriteResults(&wr, workspace)

	tui.Success(fmt.Sprintf("Added %s at %s", agentDef.DisplayName, workspace))
	tui.Hint("Run: bonsai list to see your full setup.")
	tui.Blank()
	return nil
}

// runAddItems handles adding new catalog items to an already-installed agent.
func runAddItems(cwd, configPath string, cfg *config.ProjectConfig, cat *catalog.Catalog, agentType string, installed *config.InstalledAgent) error {
	agentDef := cat.GetAgent(agentType)
	if agentDef == nil {
		tui.FatalPanel("Unknown agent type", agentType+" is not in the catalog.", "Run: bonsai catalog")
	}

	tui.Info(fmt.Sprintf("%s is already installed at %s — showing uninstalled abilities.", agentDef.DisplayName, installed.Workspace))
	tui.Blank()

	installedSet := func(items []string) map[string]bool {
		m := make(map[string]bool, len(items))
		for _, item := range items {
			m[item] = true
		}
		return m
	}

	// Filter each category to exclude already-installed items
	filterItems := func(available []catalog.CatalogItem, existing []string) []catalog.CatalogItem {
		have := installedSet(existing)
		var result []catalog.CatalogItem
		for _, item := range available {
			if !have[item.Name] {
				result = append(result, item)
			}
		}
		return result
	}

	filterSensors := func(available []catalog.SensorItem, existing []string) []catalog.SensorItem {
		have := installedSet(existing)
		var result []catalog.SensorItem
		for _, item := range available {
			if !have[item.Name] && item.Name != "routine-check" {
				result = append(result, item)
			}
		}
		return result
	}

	filterRoutines := func(available []catalog.RoutineItem, existing []string) []catalog.RoutineItem {
		have := installedSet(existing)
		var result []catalog.RoutineItem
		for _, item := range available {
			if !have[item.Name] {
				result = append(result, item)
			}
		}
		return result
	}

	newSkills := filterItems(cat.SkillsFor(agentType), installed.Skills)
	newWorkflows := filterItems(cat.WorkflowsFor(agentType), installed.Workflows)
	newProtocols := filterItems(cat.ProtocolsFor(agentType), installed.Protocols)
	newSensors := filterSensors(cat.SensorsFor(agentType), installed.Sensors)
	newRoutines := filterRoutines(cat.RoutinesFor(agentType), installed.Routines)

	totalAvailable := len(newSkills) + len(newWorkflows) + len(newProtocols) + len(newSensors) + len(newRoutines)
	if totalAvailable == 0 {
		tui.EmptyPanel("All available abilities are already installed.\nBrowse more with: bonsai catalog")
		return nil
	}

	// Show pickers — nothing pre-selected since these are additions
	var noDefaults []string

	selectedSkills, err := tui.PickItems("Skills", toItemOptions(newSkills, agentType), noDefaults)
	if err != nil {
		return err
	}
	selectedWorkflows, err := tui.PickItems("Workflows", toItemOptions(newWorkflows, agentType), noDefaults)
	if err != nil {
		return err
	}
	selectedProtocols, err := tui.PickItems("Protocols", toItemOptions(newProtocols, agentType), noDefaults)
	if err != nil {
		return err
	}
	selectedSensors, err := tui.PickItems("Sensors", toSensorOptions(newSensors, agentType), noDefaults)
	if err != nil {
		return err
	}
	selectedRoutines, err := tui.PickItems("Routines", toRoutineOptions(newRoutines, agentType), noDefaults)
	if err != nil {
		return err
	}

	totalSelected := len(selectedSkills) + len(selectedWorkflows) + len(selectedProtocols) + len(selectedSensors) + len(selectedRoutines)
	if totalSelected == 0 {
		tui.Info("No new abilities selected.")
		return nil
	}

	// Review summary — show only the new abilities being added
	describer := func(name string) string {
		if item := cat.GetItem(name); item != nil {
			return item.Description
		}
		if sensor := cat.GetSensor(name); sensor != nil {
			return sensor.Description
		}
		if routine := cat.GetRoutine(name); routine != nil {
			return routine.Description
		}
		return ""
	}

	summary := tui.ItemTree(
		tui.StyleLabel.Render(agentDef.DisplayName)+" "+tui.StyleMuted.Render(tui.GlyphArrow+" "+installed.Workspace),
		[]tui.Category{
			{Name: "Skills", Items: selectedSkills},
			{Name: "Workflows", Items: selectedWorkflows},
			{Name: "Protocols", Items: selectedProtocols},
			{Name: "Sensors", Items: selectedSensors},
			{Name: "Routines", Items: selectedRoutines},
		},
		describer,
	)

	tui.TitledPanel("Adding", summary, tui.Water)

	confirmed, err := tui.AskConfirm("Generate files?", true)
	if err != nil || !confirmed {
		return nil
	}

	// Append new selections to existing config
	installed.Skills = append(installed.Skills, selectedSkills...)
	installed.Workflows = append(installed.Workflows, selectedWorkflows...)
	installed.Protocols = append(installed.Protocols, selectedProtocols...)
	installed.Sensors = append(installed.Sensors, selectedSensors...)
	installed.Routines = append(installed.Routines, selectedRoutines...)

	generate.EnsureRoutineCheckSensor(installed)

	if err := cfg.Save(configPath); err != nil {
		return err
	}

	lock, _ := config.LoadLockFile(cwd)
	var wr generate.WriteResult

	_ = spinner.New().
		Title("Generating files...").
		Action(func() {
			_ = generate.AgentWorkspace(cwd, agentDef, installed, cfg, cat, lock, &wr, false)
			_ = generate.PathScopedRules(cwd, cfg, cat, lock, &wr, false)
			_ = generate.WorkflowSkills(cwd, cfg, cat, lock, &wr, false)
			_ = generate.SettingsJSON(cwd, cfg, cat, lock, &wr, false)
		}).
		Run()

	if wr.HasConflicts() {
		resolveConflicts(&wr, lock, cwd)
	}

	if err := lock.Save(cwd); err != nil {
		tui.Warning("Could not save lock file: " + err.Error())
	}

	showWriteResults(&wr, installed.Workspace)

	tui.Success(fmt.Sprintf("Added %d abilities to %s", totalSelected, agentDef.DisplayName))
	tui.Hint("Run: bonsai list to see your full setup.")
	tui.Blank()
	return nil
}
