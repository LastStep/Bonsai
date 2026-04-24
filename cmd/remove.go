package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/LastStep/Bonsai/internal/catalog"
	"github.com/LastStep/Bonsai/internal/config"
	"github.com/LastStep/Bonsai/internal/generate"
	"github.com/LastStep/Bonsai/internal/tui"
	"github.com/LastStep/Bonsai/internal/tui/harness"
	"github.com/LastStep/Bonsai/internal/tui/initflow"
	"github.com/LastStep/Bonsai/internal/tui/removeflow"
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

// ─── Agent removal (cinematic) ─────────────────────────────────────────

// runRemove handles `bonsai remove <agent>`. Plan 31 Phase E replaces the
// legacy raw-harness + tui.FatalPanel path with the cinematic removeflow
// stages (Observe → Confirm → [Conflicts] → Yield) while keeping the
// business logic (tech-lead guard, lock-untrack, SettingsJSON regeneration,
// optional --delete-files cleanup) in cmd/.
func runRemove(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return cmd.Help()
	}

	startedAt := time.Now()

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

	// Preserve existing tech-lead guard behaviour — silently print and bail.
	if agentName == "tech-lead" && len(cfg.Agents) > 1 {
		tui.ErrorDetail("Tech Lead in use", "Other agents depend on Tech Lead. Remove them first.", "Run: bonsai list")
		return nil
	}

	cat := loadCatalog()

	agentDisplay := agentDisplayName(cat, agentName)

	// Shared context stamped onto every removeflow stage. HeaderAction /
	// HeaderRightLabel render as "REMOVE" + "UPROOTING FROM" so the header
	// right-block reads with the remove metaphor.
	ctx := initflow.StageContext{
		Version:          Version,
		ProjectDir:       cwd,
		StationDir:       "station/",
		AgentDisplay:     agentDisplay,
		StartedAt:        startedAt,
		HeaderAction:     "REMOVE",
		HeaderRightLabel: "UPROOTING FROM",
	}

	lock, _ := config.LoadLockFile(cwd)
	var wr generate.WriteResult
	var outcome removeflow.Outcome
	counts := removeflow.AbilityCounts{
		Skills:    len(agent.Skills),
		Workflows: len(agent.Workflows),
		Protocols: len(agent.Protocols),
		Sensors:   len(agent.Sensors),
		Routines:  len(agent.Routines),
	}

	observe := removeflow.NewObserveAgent(ctx, agentName, agentDisplay, agent.Workspace,
		agent.Skills, agent.Workflows, agent.Protocols, agent.Sensors, agent.Routines)
	confirm := removeflow.NewConfirmStage(ctx,
		fmt.Sprintf("Uproot %s?", agentDisplay),
		"Removes the agent and every ability installed under its workspace.",
		"Modified files trigger a conflict prompt — nothing overwritten silently.")

	confirmed := func(prev []any) bool {
		for _, v := range prev {
			if b, ok := v.(bool); ok {
				return b
			}
		}
		return false
	}
	actionSucceeded := func(prev []any) bool {
		return confirmed(prev) && outcome.Ran && outcome.Err == nil
	}

	steps := []harness.Step{
		observe,
		confirm,
		harness.NewConditional(
			harness.NewSpinner("Removing", "Uprooting "+agentDisplay+"...", func() error {
				outcome.Ran = true
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
				// Plan 31 Phase C: refresh .bonsai/catalog.json snapshot.
				errs = append(errs, generate.WriteCatalogSnapshot(cwd, Version, cat, &wr))
				joined := errors.Join(errs...)
				outcome.Err = joined
				return joined
			}),
			confirmed,
		),
		harness.NewLazyGroup("Resolve conflicts", func(prev []any) []harness.Step {
			if !actionSucceeded(prev) {
				return nil
			}
			if !wr.HasConflicts() {
				return nil
			}
			return []harness.Step{removeflow.NewConflictsStage(ctx, &wr)}
		}),
		harness.NewConditional(
			harness.NewLazy("Yield", func(_ []any) harness.Step {
				return removeflow.NewYieldAgentSuccess(ctx, agentDisplay, agent.Workspace, counts)
			}),
			actionSucceeded,
		),
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

	if !confirmed(results) {
		return nil
	}

	if outcome.Err != nil {
		tui.Warning("Removal error: " + outcome.Err.Error())
		return nil
	}

	// Cinematic conflict picker at a trailing slot — type-scan for the
	// ConflictsStage result (map[string]config.ConflictAction).
	if wr.HasConflicts() {
		for _, r := range results {
			if picks, ok := r.(map[string]config.ConflictAction); ok {
				applyCinematicConflictPicks(picks, &wr, lock, cwd)
				break
			}
		}
	}

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

// runRemoveItem handles `bonsai remove <type> <name>`. Plan 31 Phase E
// replaces the legacy raw-harness path with the cinematic removeflow stages.
// Business-logic helpers (filterRequired / itemIsRequired / resolveTargets /
// runRemoveItemAction / agentItemList / removeFromItemList / itemInList /
// itemDisplayName / agentDisplayName) stay in cmd/ so this file is the only
// place that mutates cfg / lock / generated files.
func runRemoveItem(name string, it itemType) error {
	// Block auto-managed sensors
	if it.singular == "sensor" && name == "routine-check" {
		tui.ErrorDetail("Auto-managed sensor", "routine-check is added and removed automatically when routines change.", "")
		return nil
	}

	startedAt := time.Now()

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

	// Pre-filter required: abort up-front if every match has the item as
	// required. Legacy behaviour — keep the same Warning output.
	allowedAll := filterRequired(matches, cat, name, it)
	if len(allowedAll) == 0 {
		tui.ErrorDetail("Required item", fmt.Sprintf("%s is required by all agents that have it.", itemDisplayName(cat, name, it)), "")
		return nil
	}

	displayName := itemDisplayName(cat, name, it)
	typeDisplay := catalog.DisplayNameFrom(it.singular)

	// Build cinematic agent options from the allowed matches + aggregate row.
	needsPicker := len(allowedAll) > 1
	options := buildRemoveOptions(allowedAll, cat)

	ctx := initflow.StageContext{
		Version:          Version,
		ProjectDir:       cwd,
		StationDir:       "station/",
		AgentDisplay:     "",
		StartedAt:        startedAt,
		HeaderAction:     "REMOVE",
		HeaderRightLabel: "UPROOTING FROM",
	}

	lock, _ := config.LoadLockFile(cwd)
	var wr generate.WriteResult
	var outcome removeflow.Outcome

	// capturedTargets is populated by the Observe lazy builder (after the
	// Select stage fires) and read by the action spinner closure. The
	// closures hold a reference to the variable, so by the time the spinner
	// runs the targets are set.
	var capturedTargets []agentMatch

	confirmed := func(prev []any) bool {
		for _, v := range prev {
			if b, ok := v.(bool); ok {
				return b
			}
		}
		return false
	}
	actionSucceeded := func(prev []any) bool {
		return confirmed(prev) && outcome.Ran && outcome.Err == nil
	}

	// Resolve targets up-front when the picker is skipped so Observe renders
	// the correct FROM list even when SelectStage is gated off.
	if !needsPicker {
		capturedTargets = allowedAll
	}

	confirmStage := removeflow.NewConfirmStage(ctx,
		fmt.Sprintf("Uproot %s?", displayName),
		fmt.Sprintf("Removes the %s and any generated files under the target agent's workspace.", it.singular),
		"Modified files trigger a conflict prompt — nothing overwritten silently.")

	steps := []harness.Step{
		// Step 0: optional agent picker. Predicate gates rendering — when only
		// one viable target exists, the Conditional auto-completes and the
		// Observe LazyStep reads capturedTargets (pre-seeded above).
		harness.NewConditional(
			removeflow.NewSelectStage(ctx, displayName, it.singular, options),
			func(prev []any) bool { return needsPicker },
		),

		// Step 1: Observe preview — lazy so the picker result (if any) can be
		// resolved into a concrete target slice before the panel renders.
		harness.NewLazy("Observe", func(prev []any) harness.Step {
			if needsPicker && len(prev) > 0 {
				capturedTargets = resolveRemoveTargets(prev[0], allowedAll)
			}
			targets := buildTargetOptions(capturedTargets, cat)
			return removeflow.NewObserveItem(ctx, displayName, typeDisplay, targets)
		}),

		// Step 2: Confirm gate.
		confirmStage,

		// Step 3: action spinner — gated by confirm bool.
		harness.NewConditional(
			harness.NewSpinner("Removing", "Uprooting "+it.singular+"...", func() error {
				outcome.Ran = true
				err := runRemoveItemAction(cwd, cfg, cat, lock, &wr, configPath, name, it, capturedTargets)
				// Plan 31 Phase C: refresh .bonsai/catalog.json snapshot.
				snapErr := generate.WriteCatalogSnapshot(cwd, Version, cat, &wr)
				joined := errors.Join(err, snapErr)
				outcome.Err = joined
				return joined
			}),
			confirmed,
		),

		// Step 4: conflict picker — gated by confirm + conflicts existing.
		harness.NewLazyGroup("Resolve conflicts", func(prev []any) []harness.Step {
			if !actionSucceeded(prev) {
				return nil
			}
			if !wr.HasConflicts() {
				return nil
			}
			return []harness.Step{removeflow.NewConflictsStage(ctx, &wr)}
		}),

		// Step 5: Yield.
		harness.NewConditional(
			harness.NewLazy("Yield", func(_ []any) harness.Step {
				targets := buildTargetOptions(capturedTargets, cat)
				return removeflow.NewYieldItemSuccess(ctx, displayName, typeDisplay, targets)
			}),
			actionSucceeded,
		),
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

	if !confirmed(results) {
		return nil
	}

	if outcome.Err != nil {
		tui.Warning("Removal error: " + outcome.Err.Error())
		return nil
	}

	// Cinematic conflict picker — type-scan for ConflictsStage result.
	if wr.HasConflicts() {
		for _, r := range results {
			if picks, ok := r.(map[string]config.ConflictAction); ok {
				applyCinematicConflictPicks(picks, &wr, lock, cwd)
				break
			}
		}
	}

	if err := lock.Save(cwd); err != nil {
		tui.Warning("Could not save lock file: " + err.Error())
	}

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

// buildRemoveOptions maps allowed matches onto removeflow.AgentOption rows
// plus an aggregate "All agents" entry. Returns nil when there's a single
// match (picker is then skipped by caller).
func buildRemoveOptions(matches []agentMatch, cat *catalog.Catalog) []removeflow.AgentOption {
	if len(matches) <= 1 {
		return nil
	}
	out := make([]removeflow.AgentOption, 0, len(matches)+1)
	for _, m := range matches {
		out = append(out, removeflow.AgentOption{
			Name:        m.name,
			DisplayName: agentDisplayName(cat, m.name),
			Workspace:   m.agent.Workspace,
		})
	}
	out = append(out, removeflow.AgentOption{
		Name:        "_all_",
		DisplayName: "All agents",
		All:         true,
	})
	return out
}

// buildTargetOptions converts resolved target agentMatches into a
// removeflow.AgentOption slice for the Observe / Yield panels. Never
// includes an "All" row — the aggregate is expanded at selection time.
func buildTargetOptions(targets []agentMatch, cat *catalog.Catalog) []removeflow.AgentOption {
	out := make([]removeflow.AgentOption, 0, len(targets))
	for _, t := range targets {
		out = append(out, removeflow.AgentOption{
			Name:        t.name,
			DisplayName: agentDisplayName(cat, t.name),
			Workspace:   t.agent.Workspace,
		})
	}
	return out
}

// resolveRemoveTargets converts the SelectStage result into a concrete
// match slice. "_all_" → all allowed matches; specific name → that match
// only; unknown name (defensive) → all matches.
func resolveRemoveTargets(picked any, matches []agentMatch) []agentMatch {
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

// runRemoveItemAction is the body of the action spinner — mutates config,
// lockfile, and generated files for each target. Generator calls are
// aggregated via errors.Join so any IO failure surfaces post-harness;
// os.Remove ENOENT swallowing is legitimate (the file may already be gone).
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
