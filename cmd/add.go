package cmd

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
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
// string is trimmed, then normalised via filepath.Clean (so "./backend",
// "backend", and "backend/" all collapse to the same key) before comparison
// against the existing-set keys.
func workspaceUniqueValidator(existing map[string]bool) func(string) error {
	return func(s string) error {
		v := strings.TrimSpace(s)
		if v == "" {
			return nil // required validator handles empty
		}
		v = strings.TrimRight(filepath.Clean(v), "/") + "/"
		if existing[v] {
			return fmt.Errorf("workspace %q is already in use", v)
		}
		return nil
	}
}

// normaliseWorkspace applies the same trim + Clean + trailing-slash rule the
// validator uses, so the post-harness write path stores keys identical to what
// the validator compared against.
func normaliseWorkspace(s string) string {
	v := strings.TrimSpace(s)
	v = strings.TrimRight(filepath.Clean(v), "/") + "/"
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

	existingWorkspaces := make(map[string]bool)
	for _, a := range cfg.Agents {
		key := strings.TrimRight(filepath.Clean(a.Workspace), "/") + "/"
		existingWorkspaces[key] = true
	}

	bannerLine := "BONSAI"
	if Version != "" && Version != "dev" {
		bannerLine = fmt.Sprintf("BONSAI v%s", Version)
	}

	// Lock + WriteResult shared between the spinner closure and the conflict
	// picker. The spinner closure also surfaces a couple of values the
	// post-harness pipeline needs back via these closure-captured pointers.
	lock, _ := config.LoadLockFile(cwd)
	var wr generate.WriteResult
	// addOutcome captures what the spinner closure decided so the post-harness
	// success banner / write summary can render correctly.
	var addOutcome addOutcome

	steps := []harness.Step{
		harness.NewSelect("Agent", "Agent type:", agentOptions),
		harness.NewLazyGroup("Agent flow", func(prev []any) []harness.Step {
			agentType := asString(prev[0])
			if installed, exists := cfg.Agents[agentType]; exists {
				return buildAddItemsSteps(cat, agentType, installed)
			}
			// Require tech-lead before adding other agents. The user can still
			// pick "tech-lead" here to bootstrap — we only block when the pick
			// is a non-tech-lead agent and no tech-lead is installed yet. The
			// error surfaces post-harness on stdout (see ErrorDetail below) so
			// it persists after AltScreen exits — no in-harness NoteStep.
			if agentType != "tech-lead" {
				if _, hasTechLead := cfg.Agents["tech-lead"]; !hasTechLead {
					return nil
				}
			}
			return buildNewAgentSteps(cat, cfg, agentType, existingWorkspaces)
		}),
		// Spinner — runs only when the splice ended in a Review confirmed=true.
		// The predicate gates on the trailing prev[] slot being a confirmed
		// bool, which handles all branches uniformly:
		//   - all-installed empty splice → last is the agent type (string) → false
		//   - all-installed NoteStep → last is nil (NoteStep result) → false
		//   - tech-lead-required NoteStep → last is nil → false
		//   - new-agent / add-items happy path → last is review confirm bool
		harness.NewConditional(
			harness.NewSpinnerWithPrior("Generating", "Generating files...",
				func(prev []any) error {
					return runAddSpinner(prev, cwd, configPath, cfg, cat, existingWorkspaces, lock, &wr, &addOutcome)
				}),
			func(prev []any) bool {
				if len(prev) == 0 {
					return false
				}
				b, ok := prev[len(prev)-1].(bool)
				return ok && b
			},
		),
		harness.NewLazyGroup("Resolve conflicts", func(prev []any) []harness.Step {
			if !wr.HasConflicts() {
				return nil
			}
			return buildConflictSteps(&wr)
		}),
	}

	results, err := harness.Run(bannerLine, "Adding", steps)
	if err != nil {
		if errors.Is(err, harness.ErrAborted) {
			return nil
		}
		var bpe *harness.BuilderPanicError
		if errors.As(err, &bpe) {
			tui.FatalPanel("Harness builder panic",
				fmt.Sprintf("Step %q: %v", bpe.Step, bpe.Value),
				"This is a bug — please report it.")
			return nil
		}
		return err
	}

	agentType := asString(results[0])
	agentDef := cat.GetAgent(agentType)
	if agentDef == nil {
		tui.FatalPanel("Unknown agent type", agentType+" is not in the catalog.", "Run: bonsai catalog")
	}

	// Post-harness short-circuits for branches the spinner predicate skipped.
	if installed, exists := cfg.Agents[agentType]; exists {
		// "All installed" short-circuit: if every category filtered empty, the
		// splicer returned zero steps — render the durable panel on stdout.
		if availableAddItems(cat, installed).Total() == 0 {
			tui.EmptyPanel("All available abilities are already installed.\nBrowse more with: bonsai catalog")
			return nil
		}
	}

	// Tech-lead guard for the new-agent branch (must surface on stdout because
	// AltScreen tears down on exit).
	if _, exists := cfg.Agents[agentType]; !exists {
		if agentType != "tech-lead" {
			if _, hasTechLead := cfg.Agents["tech-lead"]; !hasTechLead {
				tui.ErrorDetail("Tech Lead required", "No tech-lead agent is installed yet.", "Run: bonsai init")
				return nil
			}
		}
	}

	// If the spinner predicate skipped (user declined review or no review
	// emitted), there's nothing more to do.
	if !addOutcome.ran {
		return nil
	}

	if addOutcome.spinnerErr != nil {
		tui.Warning("Generation error: " + addOutcome.spinnerErr.Error())
		return nil
	}

	// Add-items branch with no picks: show an info line and exit. The spinner
	// closure short-circuited before any generation, so there is nothing to
	// summarise.
	if !addOutcome.newAgent && addOutcome.totalSelected == 0 {
		tui.Info("No new abilities selected.")
		return nil
	}

	// The conflict picker is the trailing LazyGroup; when it splices it adds
	// exactly two slots (MultiSelect, then a Conditional-wrapped Confirm). The
	// agent-flow splice above produces a variable number of slots, so compute
	// the picker index from the tail rather than declaring a fixed slot. When
	// there are no conflicts the LazyGroup splices nothing and we pass -1 so
	// applyConflictPicks no-ops cleanly.
	confIdx := -1
	if wr.HasConflicts() {
		confIdx = len(results) - 2
	}
	applyConflictPicks(results, confIdx, &wr, lock, cwd)

	if err := lock.Save(cwd); err != nil {
		tui.Warning("Could not save lock file: " + err.Error())
	}

	showWriteResults(&wr)

	if addOutcome.newAgent {
		tui.Success(fmt.Sprintf("Added %s at %s", addOutcome.agentDef.DisplayName, addOutcome.workspace))
	} else {
		tui.Success(fmt.Sprintf("Added %d abilities to %s", addOutcome.totalSelected, addOutcome.agentDef.DisplayName))
	}
	tui.Hint("Run: bonsai list to see your full setup.")
	tui.Blank()
	return nil
}

// addOutcome captures the side-effects produced by the runAddSpinner closure
// so the post-harness pipeline can render the right success banner + tree.
type addOutcome struct {
	agentDef      *catalog.AgentDef
	workspace     string
	newAgent      bool  // true = new-agent branch, false = add-items branch
	totalSelected int   // for add-items success message
	spinnerErr    error // surfaced after harness exits
	ran           bool  // did the spinner action actually fire (predicate true)?
}

// runAddSpinner is the unified action body for both runAdd branches. It reads
// prev[] to determine which branch ran, applies all config mutations, calls
// cfg.Save, and runs the generator pipeline. Side-effects flow back through
// the *generate.WriteResult and *addOutcome pointers.
func runAddSpinner(prev []any, cwd, configPath string, cfg *config.ProjectConfig, cat *catalog.Catalog,
	existingWorkspaces map[string]bool, lock *config.LockFile, wr *generate.WriteResult,
	outcome *addOutcome) error {
	outcome.ran = true

	agentType := asString(prev[0])
	agentDef := cat.GetAgent(agentType)
	if agentDef == nil {
		outcome.spinnerErr = fmt.Errorf("unknown agent type %q", agentType)
		return outcome.spinnerErr
	}
	outcome.agentDef = agentDef

	if installed, exists := cfg.Agents[agentType]; exists {
		// Add-items branch. prev layout (after splice replaces LazyGroup):
		//   [0]=agentType, [1]=intro NoteStep (nil), [2..K]=pickers,
		//   [K+1]=review confirm.
		reviewIdx := len(prev) - 1
		picks := make([][]string, 0, reviewIdx-2)
		for i := 2; i < reviewIdx; i++ {
			picks = append(picks, asStringSlice(prev[i]))
		}

		selectedSkills, selectedWorkflows, selectedProtocols, selectedSensors, selectedRoutines := distributeAddItemPicks(cat, installed, picks)
		totalSelected := len(selectedSkills) + len(selectedWorkflows) + len(selectedProtocols) + len(selectedSensors) + len(selectedRoutines)
		if totalSelected == 0 {
			// Nothing to apply — leave outcome.totalSelected=0 and let the
			// post-harness path render an Info message instead of a Success.
			outcome.workspace = installed.Workspace
			return nil
		}

		installed.Skills = append(installed.Skills, selectedSkills...)
		installed.Workflows = append(installed.Workflows, selectedWorkflows...)
		installed.Protocols = append(installed.Protocols, selectedProtocols...)
		installed.Sensors = append(installed.Sensors, selectedSensors...)
		installed.Routines = append(installed.Routines, selectedRoutines...)
		generate.EnsureRoutineCheckSensor(installed)

		if err := cfg.Save(configPath); err != nil {
			outcome.spinnerErr = err
			return err
		}

		var errs []error
		errs = append(errs, generate.AgentWorkspace(cwd, agentDef, installed, cfg, cat, lock, wr, false))
		errs = append(errs, generate.PathScopedRules(cwd, cfg, cat, lock, wr, false))
		errs = append(errs, generate.WorkflowSkills(cwd, cfg, cat, lock, wr, false))
		errs = append(errs, generate.SettingsJSON(cwd, cfg, cat, lock, wr, false))
		if joined := errors.Join(errs...); joined != nil {
			outcome.spinnerErr = joined
			return joined
		}

		outcome.workspace = installed.Workspace
		outcome.totalSelected = totalSelected
		outcome.newAgent = false
		return nil
	}

	// New-agent branch. prev layout (after splice replaces LazyGroup):
	//   [0]=agentType, [1]=workspace, [2..6]=pickers, [7]=review confirm.
	if len(prev) < 8 {
		// Splicer returned a single NoteStep; nothing to do.
		outcome.ran = false
		return nil
	}

	workspace := newAgentWorkspace(cfg, agentType, prev[1:])
	if existingWorkspaces[workspace] {
		outcome.spinnerErr = fmt.Errorf("workspace %q is already in use by another agent", workspace)
		return outcome.spinnerErr
	}

	installed := &config.InstalledAgent{
		AgentType: agentType,
		Workspace: workspace,
		Skills:    asStringSlice(prev[2]),
		Workflows: asStringSlice(prev[3]),
		Protocols: asStringSlice(prev[4]),
		Sensors:   asStringSlice(prev[5]),
		Routines:  asStringSlice(prev[6]),
	}
	generate.EnsureRoutineCheckSensor(installed)
	cfg.Agents[agentType] = installed
	if err := cfg.Save(configPath); err != nil {
		outcome.spinnerErr = err
		return err
	}

	var errs []error
	errs = append(errs, generate.AgentWorkspace(cwd, agentDef, installed, cfg, cat, lock, wr, false))
	errs = append(errs, generate.PathScopedRules(cwd, cfg, cat, lock, wr, false))
	errs = append(errs, generate.WorkflowSkills(cwd, cfg, cat, lock, wr, false))
	errs = append(errs, generate.SettingsJSON(cwd, cfg, cat, lock, wr, false))
	if joined := errors.Join(errs...); joined != nil {
		outcome.spinnerErr = joined
		return joined
	}

	outcome.workspace = workspace
	outcome.newAgent = true
	return nil
}

// distributeAddItemPicks splits the per-category picks slice into the five
// category-typed slices, respecting the same skip-empty-categories rule that
// buildAddItemsSteps uses when constructing the picker step list. Mirrors the
// inline walk in the legacy finaliseAddItems.
func distributeAddItemPicks(cat *catalog.Catalog, installed *config.InstalledAgent, picks [][]string) (skills, workflows, protocols, sensors, routines []string) {
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
		skills = take()
	}
	if hasNew(cat.WorkflowsFor(agentType), installed.Workflows) {
		workflows = take()
	}
	if hasNew(cat.ProtocolsFor(agentType), installed.Protocols) {
		protocols = take()
	}
	if hasNewSensor(cat.SensorsFor(agentType), installed.Sensors) {
		sensors = take()
	}
	if hasNewRoutine(cat.RoutinesFor(agentType), installed.Routines) {
		routines = take()
	}
	return
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

// availableAddSet is the result of filtering the catalog against an installed
// agent: only the items not already installed in each category.
type availableAddSet struct {
	Skills    []catalog.CatalogItem
	Workflows []catalog.CatalogItem
	Protocols []catalog.CatalogItem
	Sensors   []catalog.SensorItem
	Routines  []catalog.RoutineItem
}

// Total reports the total count across all categories — used by the "all
// installed" short-circuit.
func (a availableAddSet) Total() int {
	return len(a.Skills) + len(a.Workflows) + len(a.Protocols) + len(a.Sensors) + len(a.Routines)
}

// availableAddItems computes the uninstalled-per-category slices for an
// already-installed agent. Shared by buildAddItemsSteps, the LazyGroup
// splicer's empty-check, and the post-harness "all installed" renderer so all
// three see the same filter result.
func availableAddItems(cat *catalog.Catalog, installed *config.InstalledAgent) availableAddSet {
	agentType := installed.AgentType

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

	return availableAddSet{
		Skills:    filterItems(cat.SkillsFor(agentType), installed.Skills),
		Workflows: filterItems(cat.WorkflowsFor(agentType), installed.Workflows),
		Protocols: filterItems(cat.ProtocolsFor(agentType), installed.Protocols),
		Sensors:   filterSensors(cat.SensorsFor(agentType), installed.Sensors),
		Routines:  filterRoutines(cat.RoutinesFor(agentType), installed.Routines),
	}
}

// buildAddItemsSteps returns the sub-sequence for appending abilities to an
// already-installed agent. Categories with zero uninstalled items are omitted
// entirely so the user doesn't page through empty pickers. If every category
// is empty the function returns nil — the splicer collapses to zero steps and
// the post-harness pipeline renders the "All installed" panel to stdout.
func buildAddItemsSteps(cat *catalog.Catalog, agentType string, installed *config.InstalledAgent) []harness.Step {
	agentDef := cat.GetAgent(agentType)
	if agentDef == nil {
		return []harness.Step{harness.NewNote("Unknown agent type", agentType+" is not in the catalog. Run: bonsai catalog")}
	}

	avail := availableAddItems(cat, installed)
	newSkills := avail.Skills
	newWorkflows := avail.Workflows
	newProtocols := avail.Protocols
	newSensors := avail.Sensors
	newRoutines := avail.Routines

	if avail.Total() == 0 {
		// Empty splice — harness expands to zero steps. The "All installed"
		// panel is rendered post-harness on normal stdout so the user has a
		// durable record after AltScreen tears down.
		return nil
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
