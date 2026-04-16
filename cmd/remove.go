package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
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
	cwd, _ := os.Getwd()
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
		tui.ErrorPanel("Cannot remove Tech Lead while other agents are installed.\nRemove other agents first.")
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

	tui.TitledPanel("Remove", preview, tui.Amber)

	confirmed, err := tui.AskConfirm("Remove "+agentDisplayName+"?", false)
	if err != nil || !confirmed {
		return nil
	}

	lock, _ := config.LoadLockFile(cwd)
	var wr generate.WriteResult

	_ = spinner.New().
		Title("Removing agent...").
		Action(func() {
			// Untrack all files for this agent's workspace from lock
			wsPrefix := agent.Workspace
			for relPath := range lock.Files {
				if strings.HasPrefix(relPath, wsPrefix) {
					lock.Untrack(relPath)
				}
			}
			delete(cfg.Agents, agentName)
			_ = cfg.Save(configPath)
			_ = generate.SettingsJSON(cwd, cfg, cat, lock, &wr, false)
		}).
		Run()

	if wr.HasConflicts() {
		resolveConflicts(&wr, lock, cwd)
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
		tui.ErrorPanel("routine-check is auto-managed.\nIt is added/removed automatically when routines change.")
		return nil
	}

	cwd, _ := os.Getwd()
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
		tui.ErrorPanel(fmt.Sprintf("%s %q is not installed in any agent.", it.singular, name))
		return nil
	}

	// If multiple agents have this item, ask which one
	var targets []agentMatch
	if len(matches) == 1 {
		targets = matches
	} else {
		var options []huh.Option[string]
		for _, m := range matches {
			label := agentDisplayName(cat, m.name)
			options = append(options,
				huh.NewOption(label+" "+tui.StyleMuted.Render(tui.GlyphArrow+" "+m.agent.Workspace), m.name))
		}
		options = append(options, huh.NewOption("All agents", "_all_"))

		selected, err := tui.AskSelect("Remove from which agent?", options)
		if err != nil {
			return nil
		}
		if selected == "_all_" {
			targets = matches
		} else {
			for _, m := range matches {
				if m.name == selected {
					targets = append(targets, m)
					break
				}
			}
		}
	}

	// Check required status — filter out agents where the item is required
	var allowed []agentMatch
	for _, t := range targets {
		if itemIsRequired(cat, name, it, t.agent.AgentType) {
			tui.Warning(fmt.Sprintf("%s is required for %s — skipping.",
				itemDisplayName(cat, name, it), agentDisplayName(cat, t.name)))
		} else {
			allowed = append(allowed, t)
		}
	}
	if len(allowed) == 0 {
		tui.ErrorPanel(fmt.Sprintf("%s is required and cannot be removed.", itemDisplayName(cat, name, it)))
		return nil
	}
	targets = allowed

	displayName := itemDisplayName(cat, name, it)

	// Build summary
	var fromLabels []string
	for _, t := range targets {
		fromLabels = append(fromLabels, agentDisplayName(cat, t.name)+" ("+t.agent.Workspace+")")
	}

	content := tui.CardFields([][2]string{
		{"Item", displayName},
		{"Type", catalog.DisplayNameFrom(it.singular)},
		{"From", strings.Join(fromLabels, ", ")},
	})
	tui.TitledPanel("Remove Item", content, tui.Amber)

	confirmed, err := tui.AskConfirm("Remove "+displayName+"?", false)
	if err != nil || !confirmed {
		return nil
	}

	lock, _ := config.LoadLockFile(cwd)
	var wr generate.WriteResult

	_ = spinner.New().
		Title("Removing " + it.singular + "...").
		Action(func() {
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
						_ = generate.RoutineDashboard(cwd, workspaceRoot, t.agent, cat, lock, &wr, false)
					} else {
						dashPath := filepath.Join(t.agent.Workspace, "agent", "Core", "routines.md")
						lock.Untrack(dashPath)
						_ = os.Remove(filepath.Join(cwd, dashPath))
					}
				}

				// Regenerate workspace CLAUDE.md
				if agentDef := cat.GetAgent(t.name); agentDef != nil {
					workspaceRoot := filepath.Join(cwd, t.agent.Workspace)
					_ = generate.WorkspaceClaudeMD(cwd, workspaceRoot, agentDef, t.agent, cfg, cat, lock, &wr, false)
				}
			}

			_ = cfg.Save(configPath)
			_ = generate.SettingsJSON(cwd, cfg, cat, lock, &wr, false)
		}).
		Run()

	if wr.HasConflicts() {
		resolveConflicts(&wr, lock, cwd)
	}

	if err := lock.Save(cwd); err != nil {
		tui.Warning("Could not save lock file: " + err.Error())
	}

	tui.Success("Removed " + displayName)
	tui.Blank()
	return nil
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
