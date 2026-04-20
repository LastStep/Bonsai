package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
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
	rootCmd.AddCommand(removeCmd)
	removeCmd.Flags().BoolP("delete-files", "d", false, "Also delete the generated agent/ directory")

	removeCmd.AddCommand(removeSkillCmd)
	removeCmd.AddCommand(removeWorkflowCmd)
	removeCmd.AddCommand(removeProtocolCmd)
	removeCmd.AddCommand(removeSensorCmd)
	removeCmd.AddCommand(removeRoutineCmd)
}

var removeCmd = &cobra.Command{
	Use:   "remove <agent | skill|workflow|protocol|sensor|routine <name>>",
	Short: "Remove an installed agent or individual ability from the project.",
	RunE:  runRemove,
}

// ─── Agent removal (existing behavior) ──────────────────────────────────

func runRemove(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return cmd.Help()
	}

	agentName := args[0]
	cwd := mustCwd()
	configPath := filepath.Join(cwd, configFile)
	cfg, err := requireConfig(configPath)
	if err != nil {
		return err
	}

	agent, exists := cfg.Agents[agentName]
	if !exists {
		tui.FatalPanel("Agent not installed", agentName+" is not in the current project.", "Run: bonsai list")
	}

	// Prevent removing tech-lead while other agents depend on it
	if agentName == "tech-lead" && len(cfg.Agents) > 1 {
		tui.ErrorDetail("Tech Lead in use", "Other agents depend on Tech Lead. Remove them first.", "Run: bonsai list")
		return nil
	}

	cat := loadCatalog()

	agentDisplayName := catalog.DisplayNameFrom(agentName)
	if agentDef := cat.GetAgent(agentName); agentDef != nil {
		agentDisplayName = agentDef.DisplayName
	}

	preview := tui.ItemTree(
		tui.StyleLabel.Render(agentDisplayName)+" "+tui.StyleMuted.Render(tui.GlyphArrow+" "+agent.Workspace),
		[]tui.Category{
			{Name: "Skills", Items: agent.Skills},
			{Name: "Workflows", Items: agent.Workflows},
			{Name: "Protocols", Items: agent.Protocols},
			{Name: "Sensors", Items: agent.Sensors},
			{Name: "Routines", Items: agent.Routines},
		},
		nil,
	)

	// Declare lock + wr up-front so the spinner closure and the conflict-picker
	// LazyGroup can both close over them.
	lock, _ := config.LoadLockFile(cwd)
	var wr generate.WriteResult

	steps := []harness.Step{
		harness.NewReview("Confirm removal",
			tui.TitledPanelString("Remove", preview, tui.Amber),
			"Remove "+agentDisplayName+"?",
			false),
		// Spinner runs only if the user confirmed Yes.
		harness.NewConditional(
			harness.NewSpinner("Removing", "Removing agent...", func() error {
				wsPrefix := agent.Workspace
				for relPath := range lock.Files {
					if strings.HasPrefix(relPath, wsPrefix) {
						lock.Untrack(relPath)
					}
				}
				delete(cfg.Agents, agentName)
				var errs []error
				errs = append(errs, cfg.Save(configPath))
				errs = append(errs, generate.SettingsJSON(cwd, cfg, cat, lock, &wr, false))
				return errors.Join(errs...)
			}),
			func(prev []any) bool { return asBool(prev[0]) },
		),
		// Conflict picker — splice in only if Yes + conflicts exist.
		harness.NewLazyGroup("Resolve conflicts", func(prev []any) []harness.Step {
			if len(prev) == 0 || !asBool(prev[0]) {
				return nil
			}
			if !wr.HasConflicts() {
				return nil
			}
			return buildConflictSteps(&wr)
		}),
	}

	bannerLine := "BONSAI"
	if Version != "" && Version != "dev" {
		bannerLine = fmt.Sprintf("BONSAI v%s", Version)
	}

	results, err := harness.Run(bannerLine, "Removing agent", steps)
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

	if !asBool(results[0]) {
		return nil
	}

	// Spinner result at slot 1 — surface any aggregated generator error so
	// the user sees permission / IO failures rather than a silent success.
	if len(results) > 1 {
		if errVal := results[1]; errVal != nil {
			if e, ok := errVal.(error); ok && e != nil {
				tui.Warning("Removal error: " + e.Error())
				return nil
			}
		}
	}

	// Conflict-picker LazyGroup at slot 2 expands to [MultiSelect, Conditional]
	// in place. The MultiSelect (the actual conflict picks) lands at index 2.
	// applyConflictPicks tolerates the slot being absent (LazyGroup spliced
	// nothing when there are no conflicts).
	applyConflictPicks(results, 2, &wr, lock, cwd)

	if err := lock.Save(cwd); err != nil {
		tui.Warning("Could not save lock file: " + err.Error())
	}

	deleteFiles, _ := cmd.Flags().GetBool("delete-files")
	if deleteFiles {
		agentDir := filepath.Join(cwd, agent.Workspace, "agent")
		claudeMD := filepath.Join(cwd, agent.Workspace, "CLAUDE.md")
		claudeDir := filepath.Join(cwd, agent.Workspace, ".claude")
		if err := os.RemoveAll(agentDir); err == nil {
			tui.Info("Deleted " + agentDir)
		}
		if err := os.Remove(claudeMD); err == nil {
			tui.Info("Deleted " + claudeMD)
		}
		if err := os.RemoveAll(claudeDir); err == nil {
			tui.Info("Deleted " + claudeDir)
		}
	}

	tui.Success("Removed " + agentDisplayName)
	tui.Blank()
	return nil
}

// ─── Item type descriptors ──────────────────────────────────────────────

type itemType struct {
	singular string // "skill"
	dir      string // "Skills" — subdirectory under agent/
	ext      string // ".md" — file extension for the generated file
}

var (
	typeSkill    = itemType{"skill", "Skills", ".md"}
	typeWorkflow = itemType{"workflow", "Workflows", ".md"}
	typeProtocol = itemType{"protocol", "Protocols", ".md"}
	typeSensor   = itemType{"sensor", "Sensors", ".sh"}
	typeRoutine  = itemType{"routine", "Routines", ".md"}
)

// ─── Item removal subcommands ───────────────────────────────────────────

var removeSkillCmd = &cobra.Command{
	Use:   "skill <name>",
	Short: "Remove a skill from an agent.",
	Args:  cobra.ExactArgs(1),
	RunE:  func(cmd *cobra.Command, args []string) error { return runRemoveItem(args[0], typeSkill) },
}

var removeWorkflowCmd = &cobra.Command{
	Use:   "workflow <name>",
	Short: "Remove a workflow from an agent.",
	Args:  cobra.ExactArgs(1),
	RunE:  func(cmd *cobra.Command, args []string) error { return runRemoveItem(args[0], typeWorkflow) },
}

var removeProtocolCmd = &cobra.Command{
	Use:   "protocol <name>",
	Short: "Remove a protocol from an agent.",
	Args:  cobra.ExactArgs(1),
	RunE:  func(cmd *cobra.Command, args []string) error { return runRemoveItem(args[0], typeProtocol) },
}

var removeSensorCmd = &cobra.Command{
	Use:   "sensor <name>",
	Short: "Remove a sensor from an agent.",
	Args:  cobra.ExactArgs(1),
	RunE:  func(cmd *cobra.Command, args []string) error { return runRemoveItem(args[0], typeSensor) },
}

var removeRoutineCmd = &cobra.Command{
	Use:   "routine <name>",
	Short: "Remove a routine from an agent.",
	Args:  cobra.ExactArgs(1),
	RunE:  func(cmd *cobra.Command, args []string) error { return runRemoveItem(args[0], typeRoutine) },
}

// ─── Shared item removal logic ──────────────────────────────────────────

type agentMatch struct {
	name  string
	agent *config.InstalledAgent
}

func runRemoveItem(name string, it itemType) error {
	// Block auto-managed sensors
	if it.singular == "sensor" && name == "routine-check" {
		tui.ErrorDetail("Auto-managed sensor", "routine-check is added and removed automatically when routines change.", "")
		return nil
	}

	cwd := mustCwd()
	configPath := filepath.Join(cwd, configFile)
	cfg, err := requireConfig(configPath)
	if err != nil {
		return err
	}

	cat := loadCatalog()

	// Find agents that have this item installed (sorted for stable ordering)
	var matches []agentMatch
	var agentNames []string
	for agentName := range cfg.Agents {
		agentNames = append(agentNames, agentName)
	}
	sort.Strings(agentNames)
	for _, agentName := range agentNames {
		agent := cfg.Agents[agentName]
		if itemInList(agentItemList(agent, it), name) {
			matches = append(matches, agentMatch{agentName, agent})
		}
	}

	if len(matches) == 0 {
		tui.ErrorDetail(it.singular+" not installed", fmt.Sprintf("%q is not in any agent.", name), "Run: bonsai list")
		return nil
	}

	// Pre-filter required: if every match has the item as required for that
	// agent type we abort up-front. The picker is only useful when there is
	// at least one viable target left.
	allowedAll := filterRequired(matches, cat, name, it)
	if len(allowedAll) == 0 {
		tui.ErrorDetail("Required item", fmt.Sprintf("%s is required by all agents that have it.", itemDisplayName(cat, name, it)), "")
		return nil
	}

	displayName := itemDisplayName(cat, name, it)

	// needsPicker: only multiple eligible (non-required) matches need a picker.
	needsPicker := len(allowedAll) > 1
	agentOptions := buildAgentOptions(allowedAll, cat)

	// Lock + WriteResult shared across spinner closure and conflict picker.
	lock, _ := config.LoadLockFile(cwd)
	var wr generate.WriteResult

	// capturedTargets is populated by the LazyStep build closure (Step 1) and
	// read by the SpinnerStep closure (Step 2). The closures hold a reference
	// to the variable, so by the time the spinner runs the targets are set.
	var capturedTargets []agentMatch

	steps := []harness.Step{
		// Step 0: optional agent picker. Predicate gates rendering — when only
		// one viable target exists, the Conditional is auto-completed and the
		// LazyStep below resolves a single-target slice.
		harness.NewConditional(
			harness.NewSelect("Agent", "Remove from which agent?", agentOptions),
			func(prev []any) bool { return needsPicker },
		),

		// Step 1: confirm summary panel. resolveTargets handles both the
		// single-match (prev[0]==nil) and multi-match paths.
		harness.NewLazy("Confirm removal", func(prev []any) harness.Step {
			capturedTargets = resolveTargets(prev[0], allowedAll)
			panel := tui.TitledPanelString("Remove Item",
				buildItemSummary(displayName, it, cat, capturedTargets), tui.Amber)
			return harness.NewReview("Confirm removal", panel, "Remove "+displayName+"?", false)
		}),

		// Step 2: spinner — gated by confirm bool at prev[1].
		harness.NewConditional(
			harness.NewSpinner("Removing", "Removing "+it.singular+"...", func() error {
				return runRemoveItemAction(cwd, cfg, cat, lock, &wr, configPath, name, it, capturedTargets)
			}),
			func(prev []any) bool { return asBool(prev[1]) },
		),

		// Step 3: conflict picker — gated by confirm + conflicts existing.
		harness.NewLazyGroup("Resolve conflicts", func(prev []any) []harness.Step {
			if len(prev) <= 1 || !asBool(prev[1]) {
				return nil
			}
			if !wr.HasConflicts() {
				return nil
			}
			return buildConflictSteps(&wr)
		}),
	}

	bannerLine := "BONSAI"
	if Version != "" && Version != "dev" {
		bannerLine = fmt.Sprintf("BONSAI v%s", Version)
	}

	results, err := harness.Run(bannerLine, "Removing "+it.singular, steps)
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

	if !asBool(results[1]) {
		return nil
	}

	// Spinner result at slot 2 — surface any aggregated generator error so
	// the user sees permission / IO failures rather than a silent success.
	if len(results) > 2 {
		if errVal := results[2]; errVal != nil {
			if e, ok := errVal.(error); ok && e != nil {
				tui.Warning("Removal error: " + e.Error())
				return nil
			}
		}
	}

	// Conflict-picker LazyGroup at slot 3 expands to [MultiSelect, Conditional]
	// in place. The MultiSelect (the actual conflict picks) lands at index 3.
	applyConflictPicks(results, 3, &wr, lock, cwd)

	if err := lock.Save(cwd); err != nil {
		tui.Warning("Could not save lock file: " + err.Error())
	}

	tui.Success("Removed " + displayName)
	tui.Blank()
	return nil
}

// filterRequired returns the subset of matches where the item is NOT required
// for that agent type. Emits a Warning for each filtered-out match (mirroring
// the legacy inline behavior).
func filterRequired(matches []agentMatch, cat *catalog.Catalog, name string, it itemType) []agentMatch {
	var allowed []agentMatch
	for _, m := range matches {
		if itemIsRequired(cat, name, it, m.agent.AgentType) {
			tui.Warning(fmt.Sprintf("%s is required for %s — skipping.",
				itemDisplayName(cat, name, it), agentDisplayName(cat, m.name)))
		} else {
			allowed = append(allowed, m)
		}
	}
	return allowed
}

// buildAgentOptions builds the agent-picker options from matches, plus an
// "All agents" entry. Returns nil when matches is single-element since the
// picker is then skipped.
func buildAgentOptions(matches []agentMatch, cat *catalog.Catalog) []huh.Option[string] {
	if len(matches) <= 1 {
		return nil
	}
	options := make([]huh.Option[string], 0, len(matches)+1)
	for _, m := range matches {
		label := agentDisplayName(cat, m.name)
		options = append(options,
			huh.NewOption(label+" "+tui.StyleMuted.Render(tui.GlyphArrow+" "+m.agent.Workspace), m.name))
	}
	options = append(options, huh.NewOption("All agents", "_all_"))
	return options
}

// resolveTargets converts the agent-picker result + matches into the target
// slice. Handles: nil/empty (single match — auto-pick), "_all_", or a
// specific name. Falls back to all matches if the picked value doesn't match
// any known agent (defensive — should not happen in practice).
func resolveTargets(picked any, matches []agentMatch) []agentMatch {
	if picked == nil {
		return matches
	}
	selected := asString(picked)
	if selected == "" || selected == "_all_" || len(matches) == 1 {
		return matches
	}
	for _, m := range matches {
		if m.name == selected {
			return []agentMatch{m}
		}
	}
	return matches
}

// buildItemSummary returns the static summary content for the review panel,
// matching the layout produced by the legacy inline tui.CardFields call.
func buildItemSummary(displayName string, it itemType, cat *catalog.Catalog, targets []agentMatch) string {
	fromLabels := make([]string, 0, len(targets))
	for _, t := range targets {
		fromLabels = append(fromLabels, agentDisplayName(cat, t.name)+" ("+t.agent.Workspace+")")
	}
	return tui.CardFields([][2]string{
		{"Item", displayName},
		{"Type", catalog.DisplayNameFrom(it.singular)},
		{"From", strings.Join(fromLabels, ", ")},
	})
}

// runRemoveItemAction is the body of the legacy spinner.Action closure,
// extracted so it can be invoked from a SpinnerStep closure (which expects an
// error-returning function rather than a bare func()). Generator calls are
// aggregated via errors.Join so any IO or permission failure surfaces
// post-harness; os.Remove swallows are kept because ENOENT after a lock
// Untrack is legitimate (the file may already be gone).
func runRemoveItemAction(cwd string, cfg *config.ProjectConfig, cat *catalog.Catalog,
	lock *config.LockFile, wr *generate.WriteResult, configPath, name string,
	it itemType, targets []agentMatch) error {
	var errs []error
	for _, t := range targets {
		removeFromItemList(t.agent, name, it)

		// Untrack and delete the generated file
		relPath := filepath.Join(t.agent.Workspace, "agent", it.dir, name+it.ext)
		lock.Untrack(relPath)
		_ = os.Remove(filepath.Join(cwd, relPath))

		// Clean up generated trigger files
		if it.singular == "skill" {
			rulePath := filepath.Join(t.agent.Workspace, ".claude", "rules", "skill-"+name+".md")
			lock.Untrack(rulePath)
			_ = os.Remove(filepath.Join(cwd, rulePath))
		}
		if it.singular == "workflow" {
			skillDir := filepath.Join(t.agent.Workspace, ".claude", "skills", name)
			skillPath := filepath.Join(skillDir, "SKILL.md")
			lock.Untrack(filepath.Join(t.agent.Workspace, ".claude", "skills", name, "SKILL.md"))
			_ = os.Remove(filepath.Join(cwd, skillPath))
			_ = os.Remove(filepath.Join(cwd, skillDir)) // remove empty dir
		}

		// Routine-specific: update auto-sensor and dashboard
		if it.singular == "routine" {
			generate.EnsureRoutineCheckSensor(t.agent)
			workspaceRoot := filepath.Join(cwd, t.agent.Workspace)
			if len(t.agent.Routines) > 0 {
				errs = append(errs, generate.RoutineDashboard(cwd, workspaceRoot, t.agent, cat, lock, wr, false))
			} else {
				dashPath := filepath.Join(t.agent.Workspace, "agent", "Core", "routines.md")
				lock.Untrack(dashPath)
				_ = os.Remove(filepath.Join(cwd, dashPath))
			}
		}

		// Regenerate workspace CLAUDE.md
		if agentDef := cat.GetAgent(t.name); agentDef != nil {
			workspaceRoot := filepath.Join(cwd, t.agent.Workspace)
			errs = append(errs, generate.WorkspaceClaudeMD(cwd, workspaceRoot, agentDef, t.agent, cfg, cat, lock, wr, false))
		}
	}

	errs = append(errs, cfg.Save(configPath))
	errs = append(errs, generate.SettingsJSON(cwd, cfg, cat, lock, wr, false))
	return errors.Join(errs...)
}

// ─── Item list helpers ──────────────────────────────────────────────────

func agentItemList(agent *config.InstalledAgent, it itemType) []string {
	switch it.singular {
	case "skill":
		return agent.Skills
	case "workflow":
		return agent.Workflows
	case "protocol":
		return agent.Protocols
	case "sensor":
		return agent.Sensors
	case "routine":
		return agent.Routines
	}
	return nil
}

func itemInList(list []string, name string) bool {
	for _, item := range list {
		if item == name {
			return true
		}
	}
	return false
}

func removeFromItemList(agent *config.InstalledAgent, name string, it itemType) {
	filter := func(list []string) []string {
		var result []string
		for _, item := range list {
			if item != name {
				result = append(result, item)
			}
		}
		return result
	}
	switch it.singular {
	case "skill":
		agent.Skills = filter(agent.Skills)
	case "workflow":
		agent.Workflows = filter(agent.Workflows)
	case "protocol":
		agent.Protocols = filter(agent.Protocols)
	case "sensor":
		agent.Sensors = filter(agent.Sensors)
	case "routine":
		agent.Routines = filter(agent.Routines)
	}
}

func itemIsRequired(cat *catalog.Catalog, name string, it itemType, agentType string) bool {
	switch it.singular {
	case "skill":
		if item := cat.GetSkill(name); item != nil {
			return item.Required.CompatibleWith(agentType)
		}
	case "workflow":
		if item := cat.GetWorkflow(name); item != nil {
			return item.Required.CompatibleWith(agentType)
		}
	case "protocol":
		if item := cat.GetProtocol(name); item != nil {
			return item.Required.CompatibleWith(agentType)
		}
	case "sensor":
		if item := cat.GetSensor(name); item != nil {
			return item.Required.CompatibleWith(agentType)
		}
	case "routine":
		if item := cat.GetRoutine(name); item != nil {
			return item.Required.CompatibleWith(agentType)
		}
	}
	return false
}

func itemDisplayName(cat *catalog.Catalog, name string, it itemType) string {
	switch it.singular {
	case "skill":
		if item := cat.GetSkill(name); item != nil {
			return item.DisplayName
		}
	case "workflow":
		if item := cat.GetWorkflow(name); item != nil {
			return item.DisplayName
		}
	case "protocol":
		if item := cat.GetProtocol(name); item != nil {
			return item.DisplayName
		}
	case "sensor":
		if item := cat.GetSensor(name); item != nil {
			return item.DisplayName
		}
	case "routine":
		if item := cat.GetRoutine(name); item != nil {
			return item.DisplayName
		}
	}
	return catalog.DisplayNameFrom(name)
}

func agentDisplayName(cat *catalog.Catalog, name string) string {
	if agentDef := cat.GetAgent(name); agentDef != nil {
		return agentDef.DisplayName
	}
	return catalog.DisplayNameFrom(name)
}
