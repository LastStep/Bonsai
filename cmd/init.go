package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh/spinner"
	"github.com/spf13/cobra"

	"github.com/LastStep/Bonsai/internal/catalog"
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

	cat := loadCatalog()

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
	docsPath, err := tui.AskText("Station directory:", "station/", true)
	if err != nil {
		return err
	}
	docsPath = strings.TrimSpace(docsPath)
	if docsPath == "" || docsPath == "/" {
		tui.ErrorPanel("Station directory cannot be empty or root. Use a subdirectory like station/.")
		os.Exit(1)
	}
	if !strings.HasSuffix(docsPath, "/") {
		docsPath += "/"
	}

	// Scaffolding selection
	var scaffoldingOptions []tui.ItemOption
	for _, item := range cat.Scaffolding {
		desc := item.Description
		if !item.Required && item.Affects != "" {
			desc += " · if removed: " + item.Affects
		}
		scaffoldingOptions = append(scaffoldingOptions, tui.ItemOption{
			Name:     item.DisplayName,
			Value:    item.Name,
			Desc:     desc,
			Required: item.Required,
		})
	}
	selectedScaffolding, err := tui.PickItems("Project Scaffolding", scaffoldingOptions, nil)
	if err != nil {
		return err
	}

	// Tech Lead setup (required — primary agent for all projects)
	tui.Heading("Tech Lead Agent")
	tui.Info("Tech Lead is your project's primary agent — it architects the system and dispatches work to other agents.")

	const techLeadType = "tech-lead"
	agentDef := cat.GetAgent(techLeadType)
	if agentDef == nil {
		tui.ErrorPanel("Tech Lead agent not found in catalog.")
		os.Exit(1)
	}

	workspace := docsPath

	selectedSkills, err := tui.PickItems("Skills", toItemOptions(cat.SkillsFor(techLeadType), techLeadType), agentDef.DefaultSkills)
	if err != nil {
		return err
	}
	selectedWorkflows, err := tui.PickItems("Workflows", toItemOptions(cat.WorkflowsFor(techLeadType), techLeadType), agentDef.DefaultWorkflows)
	if err != nil {
		return err
	}
	selectedProtocols, err := tui.PickItems("Protocols", toItemOptions(cat.ProtocolsFor(techLeadType), techLeadType), agentDef.DefaultProtocols)
	if err != nil {
		return err
	}
	availableSensors := cat.SensorsFor(techLeadType)
	var userSensors []catalog.SensorItem
	for _, s := range availableSensors {
		if s.Name != "routine-check" {
			userSensors = append(userSensors, s)
		}
	}
	selectedSensors, err := tui.PickItems("Sensors", toSensorOptions(userSensors, techLeadType), agentDef.DefaultSensors)
	if err != nil {
		return err
	}
	selectedRoutines, err := tui.PickItems("Routines", toRoutineOptions(cat.RoutinesFor(techLeadType), techLeadType), agentDef.DefaultRoutines)
	if err != nil {
		return err
	}

	// Review summary
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
	tui.Blank()

	confirmed, err := tui.AskConfirm("Generate project?", true)
	if err != nil || !confirmed {
		return nil
	}

	// Build config with tech-lead agent
	installed := &config.InstalledAgent{
		AgentType: techLeadType,
		Workspace: workspace,
		Skills:    selectedSkills,
		Workflows: selectedWorkflows,
		Protocols: selectedProtocols,
		Sensors:   selectedSensors,
		Routines:  selectedRoutines,
	}
	generate.EnsureRoutineCheckSensor(installed)

	cfg := &config.ProjectConfig{
		ProjectName: strings.TrimSpace(projectName),
		Description: strings.TrimSpace(description),
		DocsPath:    docsPath,
		Scaffolding: selectedScaffolding,
		Agents:      map[string]*config.InstalledAgent{techLeadType: installed},
	}

	if err := cfg.Save(configPath); err != nil {
		return err
	}

	lock, _ := config.LoadLockFile(cwd)
	var wr generate.WriteResult

	_ = spinner.New().
		Title("Generating project files...").
		Action(func() {
			_ = generate.Scaffolding(cwd, cfg, cat, lock, &wr, false)
			_ = generate.AgentWorkspace(cwd, agentDef, installed, cfg, cat, lock, &wr, false)
			_ = generate.SettingsJSON(cwd, cfg, cat, lock, &wr, false)
		}).
		Run()

	if wr.HasConflicts() {
		resolveConflicts(&wr, lock, cwd)
	}

	if err := lock.Save(cwd); err != nil {
		tui.Warning("Could not save lock file: " + err.Error())
	}

	root := docsPath
	if root == "" {
		root = "."
	}
	showWriteResults(&wr, root)

	tui.Success(fmt.Sprintf("Initialized %s with %s", cfg.ProjectName, agentDef.DisplayName))
	tui.Hint("Next: run bonsai add to add code agents (backend, frontend, etc.).")
	tui.Blank()
	return nil
}
