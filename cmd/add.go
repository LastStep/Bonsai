package cmd

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/spf13/cobra"

	"github.com/LastStep/Bonsai/internal/catalog"
	"github.com/LastStep/Bonsai/internal/config"
	"github.com/LastStep/Bonsai/internal/generate"
	"github.com/LastStep/Bonsai/internal/tui"
	"github.com/LastStep/Bonsai/internal/tui/harness"
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

// newDescriber returns a lookup that resolves a machine name to its catalog
// description, searching abilities, sensors, and routines in that order.
// Shared between runAdd (both branches) and runInit's review panel.
func newDescriber(cat *catalog.Catalog) func(string) string {
	return func(name string) string {
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
}

// workspaceUniqueValidator rejects workspace paths already in use. The input
// string is trimmed and trailing-slash-normalised the same way the post-harness
// pipeline normalises the accepted value, so "backend" and "backend/" both
// collide if either variant is already installed.
func workspaceUniqueValidator(existing map[string]bool) func(string) error {
	return func(s string) error {
		v := strings.TrimSpace(s)
		if v == "" {
			return nil // required validator handles empty
		}
		v = strings.TrimRight(v, "/") + "/"
		if existing[v] {
			return fmt.Errorf("workspace %q is already in use", v)
		}
		return nil
	}
}

// normaliseWorkspace applies the same trim + trailing-slash rule the legacy
// runAdd used after collection.
func normaliseWorkspace(s string) string {
	v := strings.TrimSpace(s)
	v = strings.TrimRight(v, "/") + "/"
	return v
}

func runAdd(cmd *cobra.Command, args []string) error {
	cwd := mustCwd()
	configPath := filepath.Join(cwd, configFile)
	cfg, err := requireConfig(configPath)
	if err != nil {
		return err
	}

	cat := loadCatalog()

	var agentOptions []huh.Option[string]
	for _, a := range cat.Agents {
		agentOptions = append(agentOptions,
			huh.NewOption(a.DisplayName+" "+tui.StyleMuted.Render(tui.GlyphDash+" "+a.Description), a.Name))
	}

	// Pre-flight: require tech-lead. We check before entering AltScreen so the
	// user isn't invited into the harness only to be rejected.
	if _, hasTechLead := cfg.Agents["tech-lead"]; !hasTechLead {
		// If the only selectable agent is tech-lead, no need for the guard.
		tui.ErrorDetail("Tech Lead required", "No tech-lead agent is installed yet.", "Run: bonsai init")
		return nil
	}

	existingWorkspaces := make(map[string]bool)
	for _, a := range cfg.Agents {
		existingWorkspaces[a.Workspace] = true
	}

	bannerLine := "BONSAI"
	if Version != "" && Version != "dev" {
		bannerLine = fmt.Sprintf("BONSAI v%s", Version)
	}

	steps := []harness.Step{
		harness.NewSelect("Agent", "Agent type:", agentOptions),
		harness.NewLazyGroup("Agent flow", func(prev []any) []harness.Step {
			agentType := asString(prev[0])
			if installed, exists := cfg.Agents[agentType]; exists {
				return buildAddItemsSteps(cat, agentType, installed)
			}
			return buildNewAgentSteps(cat, cfg, agentType, existingWorkspaces)
		}),
	}

	results, err := harness.Run(bannerLine, "Adding", steps)
	if err != nil {
		if errors.Is(err, harness.ErrAborted) {
			return nil
		}
		return err
	}

	agentType := asString(results[0])
	agentDef := cat.GetAgent(agentType)
	if agentDef == nil {
		tui.FatalPanel("Unknown agent type", agentType+" is not in the catalog.", "Run: bonsai catalog")
	}

	if installed, exists := cfg.Agents[agentType]; exists {
		return finaliseAddItems(cwd, configPath, cfg, cat, agentDef, installed, results)
	}
	return finaliseNewAgent(cwd, configPath, cfg, cat, agentDef, agentType, existingWorkspaces, results)
}

// buildNewAgentSteps returns the sub-sequence for configuring a brand-new
// agent: workspace, five picker categories, review.
func buildNewAgentSteps(cat *catalog.Catalog, cfg *config.ProjectConfig, agentType string, existingWorkspaces map[string]bool) []harness.Step {
	agentDef := cat.GetAgent(agentType)
	if agentDef == nil {
		// Return a single NoteStep explaining the failure; the flow exits with
		// no writes. FatalPanel would leave AltScreen in an odd state.
		return []harness.Step{harness.NewNote("Unknown agent type", agentType+" is not in the catalog. Run: bonsai catalog")}
	}

	workspaceStep := harness.NewLazy("Workspace", func(prev []any) harness.Step {
		if agentType == "tech-lead" {
			ws := cfg.DocsPath
			if ws == "" {
				ws = "station/"
			}
			return harness.NewNote("Workspace", "Tech Lead workspace: "+ws)
		}
		return harness.NewText(
			"Workspace",
			"Workspace directory (e.g. backend/):",
			agentType+"/",
			true,
			workspaceUniqueValidator(existingWorkspaces),
		)
	})

	return []harness.Step{
		workspaceStep,
		harness.NewMultiSelect("Skills", "Skills", toItemOptions(cat.SkillsFor(agentType), agentType), agentDef.DefaultSkills),
		harness.NewMultiSelect("Workflows", "Workflows", toItemOptions(cat.WorkflowsFor(agentType), agentType), agentDef.DefaultWorkflows),
		harness.NewMultiSelect("Protocols", "Protocols", toItemOptions(cat.ProtocolsFor(agentType), agentType), agentDef.DefaultProtocols),
		harness.NewMultiSelect("Sensors", "Sensors", userSensorOptions(cat, agentType), agentDef.DefaultSensors),
		harness.NewMultiSelect("Routines", "Routines", toRoutineOptions(cat.RoutinesFor(agentType), agentType), agentDef.DefaultRoutines),
		harness.NewLazy("Review", func(prev []any) harness.Step {
			workspace := newAgentWorkspace(cfg, agentType, prev)
			// prev indices inside the splice: [0]=workspace, [1..5]=pickers.
			selectedSkills := asStringSlice(prev[1])
			selectedWorkflows := asStringSlice(prev[2])
			selectedProtocols := asStringSlice(prev[3])
			selectedSensors := asStringSlice(prev[4])
			selectedRoutines := asStringSlice(prev[5])

			tree := tui.ItemTree(
				tui.StyleLabel.Render(agentDef.DisplayName)+" "+tui.StyleMuted.Render(tui.GlyphArrow+" "+workspace),
				[]tui.Category{
					{Name: "Skills", Items: selectedSkills},
					{Name: "Workflows", Items: selectedWorkflows},
					{Name: "Protocols", Items: selectedProtocols},
					{Name: "Sensors", Items: selectedSensors},
					{Name: "Routines", Items: selectedRoutines},
				},
				newDescriber(cat),
			)

			return harness.NewReview("Review", tui.TitledPanelString("Review", tree, tui.Water), "Generate files?", true)
		}),
	}
}

// newAgentWorkspace resolves the workspace value from the workspace step's
// result slot (prev[0] inside the splice). Tech-lead uses cfg.DocsPath because
// its slot holds a NoteStep (no result).
func newAgentWorkspace(cfg *config.ProjectConfig, agentType string, prev []any) string {
	if agentType == "tech-lead" {
		ws := cfg.DocsPath
		if ws == "" {
			ws = "station/"
		}
		return ws
	}
	return normaliseWorkspace(asString(prev[0]))
}

// buildAddItemsSteps returns the sub-sequence for appending abilities to an
// already-installed agent. Categories with zero uninstalled items are omitted
// entirely so the user doesn't page through empty pickers.
func buildAddItemsSteps(cat *catalog.Catalog, agentType string, installed *config.InstalledAgent) []harness.Step {
	agentDef := cat.GetAgent(agentType)
	if agentDef == nil {
		return []harness.Step{harness.NewNote("Unknown agent type", agentType+" is not in the catalog. Run: bonsai catalog")}
	}

	installedSet := func(items []string) map[string]bool {
		m := make(map[string]bool, len(items))
		for _, item := range items {
			m[item] = true
		}
		return m
	}

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

	total := len(newSkills) + len(newWorkflows) + len(newProtocols) + len(newSensors) + len(newRoutines)
	if total == 0 {
		return []harness.Step{
			harness.NewNote("All installed", "All available abilities are already installed.\nBrowse more with: bonsai catalog"),
		}
	}

	steps := []harness.Step{
		harness.NewNote(
			"Adding to "+agentDef.DisplayName,
			agentDef.DisplayName+" is already installed at "+installed.Workspace+" — showing uninstalled abilities.",
		),
	}

	// Skip categories with zero uninstalled items. This leaves the breadcrumb
	// counter honest about the number of decision points the user has left.
	var noDefaults []string
	if len(newSkills) > 0 {
		steps = append(steps, harness.NewMultiSelect("Skills", "Skills", toItemOptions(newSkills, agentType), noDefaults))
	}
	if len(newWorkflows) > 0 {
		steps = append(steps, harness.NewMultiSelect("Workflows", "Workflows", toItemOptions(newWorkflows, agentType), noDefaults))
	}
	if len(newProtocols) > 0 {
		steps = append(steps, harness.NewMultiSelect("Protocols", "Protocols", toItemOptions(newProtocols, agentType), noDefaults))
	}
	if len(newSensors) > 0 {
		steps = append(steps, harness.NewMultiSelect("Sensors", "Sensors", toSensorOptions(newSensors, agentType), noDefaults))
	}
	if len(newRoutines) > 0 {
		steps = append(steps, harness.NewMultiSelect("Routines", "Routines", toRoutineOptions(newRoutines, agentType), noDefaults))
	}

	steps = append(steps, harness.NewLazy("Review", func(prev []any) harness.Step {
		// prev inside the splice: [0]=intro NoteStep (nil), [1..N]=pickers.
		// Collect by walking from index 1 and matching category order.
		offset := 1
		var selectedSkills, selectedWorkflows, selectedProtocols, selectedSensors, selectedRoutines []string
		if len(newSkills) > 0 {
			selectedSkills = asStringSlice(prev[offset])
			offset++
		}
		if len(newWorkflows) > 0 {
			selectedWorkflows = asStringSlice(prev[offset])
			offset++
		}
		if len(newProtocols) > 0 {
			selectedProtocols = asStringSlice(prev[offset])
			offset++
		}
		if len(newSensors) > 0 {
			selectedSensors = asStringSlice(prev[offset])
			offset++
		}
		if len(newRoutines) > 0 {
			selectedRoutines = asStringSlice(prev[offset])
			offset++
		}

		tree := tui.ItemTree(
			tui.StyleLabel.Render(agentDef.DisplayName)+" "+tui.StyleMuted.Render(tui.GlyphArrow+" "+installed.Workspace),
			[]tui.Category{
				{Name: "Skills", Items: selectedSkills},
				{Name: "Workflows", Items: selectedWorkflows},
				{Name: "Protocols", Items: selectedProtocols},
				{Name: "Sensors", Items: selectedSensors},
				{Name: "Routines", Items: selectedRoutines},
			},
			newDescriber(cat),
		)

		return harness.NewReview("Adding", tui.TitledPanelString("Adding", tree, tui.Water), "Generate files?", true)
	}))

	return steps
}

// finaliseNewAgent runs the post-harness pipeline for the new-agent branch:
// extract picks, build InstalledAgent, save config, generate, conflicts, lock,
// success banner. Matches the legacy post-prompt block from the old runAdd.
func finaliseNewAgent(cwd, configPath string, cfg *config.ProjectConfig, cat *catalog.Catalog, agentDef *catalog.AgentDef, agentType string, existingWorkspaces map[string]bool, results []any) error {
	// results layout: [0]=agentType (Select), [1..]=spliced-branch results.
	// New-agent branch: [1]=workspace, [2..6]=pickers, [7]=review confirm.
	if len(results) < 8 {
		// Splicer returned a single NoteStep (e.g., unknown agent type). Exit
		// without writes.
		return nil
	}

	workspace := newAgentWorkspace(cfg, agentType, results[1:])
	selectedSkills := asStringSlice(results[2])
	selectedWorkflows := asStringSlice(results[3])
	selectedProtocols := asStringSlice(results[4])
	selectedSensors := asStringSlice(results[5])
	selectedRoutines := asStringSlice(results[6])

	if !asBool(results[7]) {
		return nil
	}

	// Defence-in-depth: the validator already blocked duplicates, but a
	// malicious or stale existingWorkspaces map would surface here.
	if existingWorkspaces[workspace] {
		tui.FatalPanel("Workspace conflict", workspace+" is already in use by another agent.", "Choose a different directory.")
	}

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

// finaliseAddItems runs the post-harness pipeline for the add-items branch.
func finaliseAddItems(cwd, configPath string, cfg *config.ProjectConfig, cat *catalog.Catalog, agentDef *catalog.AgentDef, installed *config.InstalledAgent, results []any) error {
	// results layout: [0]=agentType, [1..]=spliced-branch results.
	// Add-items branch: [1]=intro NoteStep (nil), [2..K]=pickers (only those
	// categories with uninstalled items), [K+1]=review confirm. If the splice
	// was just the "All installed" NoteStep, len(results)==2 and we return.
	if len(results) <= 2 {
		return nil
	}

	// Walk from index 2, collect picks (always []string) until we hit the
	// review confirm (bool). The review confirm is always last.
	reviewIdx := len(results) - 1
	confirmed := asBool(results[reviewIdx])
	if !confirmed {
		return nil
	}

	var picks [][]string
	for i := 2; i < reviewIdx; i++ {
		picks = append(picks, asStringSlice(results[i]))
	}

	// Distribute picks to categories in the same order buildAddItemsSteps
	// appended them: Skills, Workflows, Protocols, Sensors, Routines.
	installedSet := func(items []string) map[string]bool {
		m := make(map[string]bool, len(items))
		for _, item := range items {
			m[item] = true
		}
		return m
	}
	agentType := installed.AgentType
	hasNew := func(available []catalog.CatalogItem, existing []string) bool {
		have := installedSet(existing)
		for _, item := range available {
			if !have[item.Name] {
				return true
			}
		}
		return false
	}
	hasNewSensor := func(available []catalog.SensorItem, existing []string) bool {
		have := installedSet(existing)
		for _, item := range available {
			if !have[item.Name] && item.Name != "routine-check" {
				return true
			}
		}
		return false
	}
	hasNewRoutine := func(available []catalog.RoutineItem, existing []string) bool {
		have := installedSet(existing)
		for _, item := range available {
			if !have[item.Name] {
				return true
			}
		}
		return false
	}

	var selectedSkills, selectedWorkflows, selectedProtocols, selectedSensors, selectedRoutines []string
	idx := 0
	take := func() []string {
		if idx >= len(picks) {
			return nil
		}
		p := picks[idx]
		idx++
		return p
	}
	if hasNew(cat.SkillsFor(agentType), installed.Skills) {
		selectedSkills = take()
	}
	if hasNew(cat.WorkflowsFor(agentType), installed.Workflows) {
		selectedWorkflows = take()
	}
	if hasNew(cat.ProtocolsFor(agentType), installed.Protocols) {
		selectedProtocols = take()
	}
	if hasNewSensor(cat.SensorsFor(agentType), installed.Sensors) {
		selectedSensors = take()
	}
	if hasNewRoutine(cat.RoutinesFor(agentType), installed.Routines) {
		selectedRoutines = take()
	}

	totalSelected := len(selectedSkills) + len(selectedWorkflows) + len(selectedProtocols) + len(selectedSensors) + len(selectedRoutines)
	if totalSelected == 0 {
		tui.Info("No new abilities selected.")
		return nil
	}

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
