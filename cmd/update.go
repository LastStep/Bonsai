package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/spf13/cobra"

	"github.com/LastStep/Bonsai/internal/config"
	"github.com/LastStep/Bonsai/internal/generate"
	"github.com/LastStep/Bonsai/internal/tui"
)

func init() {
	rootCmd.AddCommand(updateCmd)
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Sync workspace — detect custom files, re-render abilities, refresh CLAUDE.md.",
	RunE:  runUpdate,
}

func runUpdate(cmd *cobra.Command, args []string) error {
	cwd, _ := os.Getwd()
	configPath := filepath.Join(cwd, configFile)
	cfg, err := requireConfig(configPath)
	if err != nil {
		return err
	}

	cat := loadCatalog()
	lock, _ := config.LoadLockFile(cwd)
	if lock == nil {
		lock = config.NewLockFile()
	}

	tui.Heading("Update")

	configChanged := false

	// Sort agent names for deterministic output
	var agentNames []string
	for name := range cfg.Agents {
		agentNames = append(agentNames, name)
	}
	sort.Strings(agentNames)

	// 1. Scan for custom files across all agents
	for _, agentName := range agentNames {
		installed := cfg.Agents[agentName]
		discovered, scanErr := generate.ScanCustomFiles(cwd, installed, lock)
		if scanErr != nil || len(discovered) == 0 {
			continue
		}

		// Separate valid from invalid
		var valid, invalid []generate.DiscoveredFile
		for _, d := range discovered {
			if d.Error != "" {
				invalid = append(invalid, d)
			} else {
				valid = append(valid, d)
			}
		}

		// Show invalid files with warnings
		for _, d := range invalid {
			tui.Warning(fmt.Sprintf("Skipping %s: %s", d.RelPath, d.Error))
			tui.Hint("Add frontmatter to track this file. See docs/custom-files.md for format.")
		}

		if len(valid) == 0 {
			continue
		}

		agentDef := cat.GetAgent(installed.AgentType)
		agentLabel := agentName
		if agentDef != nil {
			agentLabel = agentDef.DisplayName
		}

		tui.Blank()
		tui.Section(fmt.Sprintf("Custom files found — %s", agentLabel))

		// Build multi-select options grouped by type
		var options []huh.Option[string]
		for _, d := range valid {
			displayName := d.Meta.Description
			if d.Meta.DisplayName != "" {
				displayName = d.Meta.DisplayName + " " + tui.StyleMuted.Render(tui.GlyphDash+" "+d.Meta.Description)
			}
			label := fmt.Sprintf("[%s] %s", d.Type, displayName)
			key := d.Type + ":" + d.Name
			options = append(options, huh.NewOption(label, key).Selected(true))
		}

		selected, selectErr := tui.AskMultiSelect("Track these custom files?", options)
		if selectErr != nil {
			return selectErr
		}

		selectedSet := make(map[string]bool)
		for _, s := range selected {
			selectedSet[s] = true
		}

		// Add selected items to config
		for _, d := range valid {
			if !selectedSet[d.Type+":"+d.Name] {
				continue
			}

			switch d.Type {
			case "skill":
				installed.Skills = append(installed.Skills, d.Name)
			case "workflow":
				installed.Workflows = append(installed.Workflows, d.Name)
			case "protocol":
				installed.Protocols = append(installed.Protocols, d.Name)
			case "sensor":
				installed.Sensors = append(installed.Sensors, d.Name)
			case "routine":
				installed.Routines = append(installed.Routines, d.Name)
			}

			if installed.CustomItems == nil {
				installed.CustomItems = make(map[string]*config.CustomItemMeta)
			}
			installed.CustomItems[d.Name] = d.Meta

			// Track in lock file
			data, readErr := os.ReadFile(filepath.Join(cwd, d.RelPath))
			if readErr == nil {
				lock.Track(d.RelPath, data, "custom:"+d.Type+"s/"+d.Name)
			}

			configChanged = true
		}
	}

	// 2. Re-render abilities + CLAUDE.md + settings.json for all agents
	var wr generate.WriteResult

	_ = spinner.New().
		Title("Syncing workspace...").
		Action(func() {
			for _, agentName := range agentNames {
				installed := cfg.Agents[agentName]
				agentDef := cat.GetAgent(installed.AgentType)
				if agentDef == nil {
					continue
				}
				generate.EnsureRoutineCheckSensor(installed)
				_ = generate.AgentWorkspace(cwd, agentDef, installed, cfg, cat, lock, &wr, false)
			}
			_ = generate.PathScopedRules(cwd, cfg, cat, lock, &wr, false)
			_ = generate.WorkflowSkills(cwd, cfg, cat, lock, &wr, false)
			_ = generate.SettingsJSON(cwd, cfg, cat, lock, &wr, false)
		}).
		Run()

	if wr.HasConflicts() {
		resolveConflicts(&wr, lock, cwd)
	}

	// 3. Save config + lock
	if configChanged {
		if err := cfg.Save(configPath); err != nil {
			tui.Warning("Could not save config: " + err.Error())
		}
	}

	if err := lock.Save(cwd); err != nil {
		tui.Warning("Could not save lock file: " + err.Error())
	}

	showWriteResults(&wr, ".")

	if configChanged {
		tui.Success("Update complete — custom files tracked")
	} else {
		tui.Success("Update complete — workspace synced")
	}
	tui.Blank()
	return nil
}
