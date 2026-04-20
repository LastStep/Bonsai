package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"

	"github.com/LastStep/Bonsai/internal/config"
	"github.com/LastStep/Bonsai/internal/generate"
	"github.com/LastStep/Bonsai/internal/tui"
	"github.com/LastStep/Bonsai/internal/tui/harness"
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
	cwd := mustCwd()
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

	// Sort agent names for deterministic output.
	var agentNames []string
	for name := range cfg.Agents {
		agentNames = append(agentNames, name)
	}
	sort.Strings(agentNames)

	// Pre-flight: scan all agents for discoveries. Warnings stay on stdout
	// (pre-harness), exactly as the legacy flow.
	discoveredByAgent := make(map[string][]generate.DiscoveredFile)
	for _, agentName := range agentNames {
		installed := cfg.Agents[agentName]
		discovered, scanErr := generate.ScanCustomFiles(cwd, installed, lock)
		if scanErr != nil || len(discovered) == 0 {
			continue
		}
		var valid, invalid []generate.DiscoveredFile
		for _, d := range discovered {
			if d.Error != "" {
				invalid = append(invalid, d)
			} else {
				valid = append(valid, d)
			}
		}
		for _, d := range invalid {
			tui.Warning(fmt.Sprintf("Skipping %s: %s", d.RelPath, d.Error))
			tui.Hint("Add frontmatter to track this file. See docs/custom-files.md for format.")
		}
		if len(valid) > 0 {
			discoveredByAgent[agentName] = valid
		}
	}

	// agentsWithDiscoveries preserves the deterministic per-agent ordering for
	// the LazyGroup splice and for the spinner's prev[]-based selection lookup.
	var agentsWithDiscoveries []string
	for _, agentName := range agentNames {
		if len(discoveredByAgent[agentName]) > 0 {
			agentsWithDiscoveries = append(agentsWithDiscoveries, agentName)
		}
	}

	configChanged := false
	var wr generate.WriteResult

	steps := []harness.Step{
		// Per-agent custom-file pickers, spliced in for agents with discoveries.
		harness.NewLazyGroup("Custom files", func(prev []any) []harness.Step {
			out := make([]harness.Step, 0, len(agentsWithDiscoveries))
			for _, agentName := range agentsWithDiscoveries {
				valid := discoveredByAgent[agentName]
				agentDef := cat.GetAgent(cfg.Agents[agentName].AgentType)
				agentLabel := agentName
				if agentDef != nil {
					agentLabel = agentDef.DisplayName
				}
				options := buildCustomFileOptions(valid)
				defaults := buildCustomFileDefaults(valid)
				out = append(out, harness.NewMultiSelect(
					"Custom files — "+agentLabel,
					fmt.Sprintf("Custom files found — %s", agentLabel),
					options, defaults))
			}
			return out
		}),

		// Spinner — re-renders abilities/CLAUDE.md/settings. Reads per-agent
		// selections from prev (which has length == len(agentsWithDiscoveries)
		// after the LazyGroup splice; the LazyGroup itself is replaced in place).
		harness.NewSpinnerWithPrior("Syncing", "Syncing workspace...", func(prev []any) error {
			cursor := 0
			for _, agentName := range agentsWithDiscoveries {
				valid := discoveredByAgent[agentName]
				if cursor >= len(prev) {
					break
				}
				selected := asStringSlice(prev[cursor])
				cursor++
				if applyCustomFileSelection(cfg.Agents[agentName], valid, selected, lock, cwd) {
					configChanged = true
				}
			}

			var errs []error
			for _, agentName := range agentNames {
				installed := cfg.Agents[agentName]
				agentDef := cat.GetAgent(installed.AgentType)
				if agentDef == nil {
					continue
				}
				generate.EnsureRoutineCheckSensor(installed)
				errs = append(errs, generate.AgentWorkspace(cwd, agentDef, installed, cfg, cat, lock, &wr, false))
			}
			errs = append(errs, generate.PathScopedRules(cwd, cfg, cat, lock, &wr, false))
			errs = append(errs, generate.WorkflowSkills(cwd, cfg, cat, lock, &wr, false))
			errs = append(errs, generate.SettingsJSON(cwd, cfg, cat, lock, &wr, false))
			return errors.Join(errs...)
		}),

		harness.NewLazyGroup("Resolve conflicts", func(prev []any) []harness.Step {
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

	results, err := harness.Run(bannerLine, "Updating workspace", steps)
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

	// Spinner slot sits immediately after the per-agent custom-file pickers.
	// Surface any aggregated generator error post-harness so a silent failure
	// no longer masquerades as a successful sync.
	spinnerIdx := len(agentsWithDiscoveries)
	if spinnerIdx < len(results) {
		if errVal := results[spinnerIdx]; errVal != nil {
			if e, ok := errVal.(error); ok && e != nil {
				tui.Warning("Update error: " + e.Error())
				return nil
			}
		}
	}

	// Conflict picker, if it spliced in steps, lands at index
	// len(agentsWithDiscoveries) + 1 (one slot per per-agent picker, plus the
	// spinner). The MultiSelectStep produced by buildConflictSteps lives there.
	conflictIdx := len(agentsWithDiscoveries) + 1
	applyConflictPicks(results, conflictIdx, &wr, lock, cwd)

	if configChanged {
		if err := cfg.Save(configPath); err != nil {
			tui.Warning("Could not save config: " + err.Error())
		}
	}

	if err := lock.Save(cwd); err != nil {
		tui.Warning("Could not save lock file: " + err.Error())
	}

	created, updated, _, _, conflicts := wr.Summary()
	hadChanges := configChanged || created > 0 || updated > 0 || conflicts > 0

	if !hadChanges {
		tui.TitledPanel("Up to date",
			"Workspace is in sync with the catalog.\nNo files needed updating.",
			tui.Moss)
		tui.Blank()
		return nil
	}

	showWriteResults(&wr)

	if configChanged {
		tui.Success("Update complete — custom files tracked")
	} else {
		tui.Success("Update complete — workspace synced")
	}
	tui.Hint("Review changes with: bonsai list")
	tui.Blank()
	return nil
}

// buildCustomFileOptions builds the multi-select options for a single agent's
// discovered custom files. Mirrors the legacy inline construction at
// cmd/update.go:91-100 but returns ItemOption (the harness MultiSelectStep
// shape) rather than huh.Option directly.
func buildCustomFileOptions(valid []generate.DiscoveredFile) []tui.ItemOption {
	options := make([]tui.ItemOption, 0, len(valid))
	for _, d := range valid {
		desc := d.Meta.Description
		name := fmt.Sprintf("[%s] %s", d.Type, d.Meta.DisplayName)
		if d.Meta.DisplayName == "" {
			name = fmt.Sprintf("[%s] %s", d.Type, d.Meta.Description)
			desc = ""
		}
		options = append(options, tui.ItemOption{
			Name:  name,
			Value: d.Type + ":" + d.Name,
			Desc:  desc,
		})
	}
	return options
}

// buildCustomFileDefaults returns the keys that should be pre-selected (all of
// them, mirroring the legacy `Selected(true)` on every option).
func buildCustomFileDefaults(valid []generate.DiscoveredFile) []string {
	defaults := make([]string, 0, len(valid))
	for _, d := range valid {
		defaults = append(defaults, d.Type+":"+d.Name)
	}
	return defaults
}

// appendUnique appends name to slice unless it is already present. The ability
// lists on config.InstalledAgent are small (bounded by catalog size), so a
// linear scan is cheaper than maintaining a parallel set. Guards against
// accumulating duplicates when `bonsai update` is re-run over the same
// user-selected custom files.
func appendUnique(slice []string, name string) []string {
	for _, existing := range slice {
		if existing == name {
			return slice
		}
	}
	return append(slice, name)
}

// applyCustomFileSelection mutates installed + lock for the given agent based on
// the user-selected keys. Returns true if any selections were applied (so the
// caller can flip configChanged). Lifted from the inline body at
// cmd/update.go:107-143.
func applyCustomFileSelection(installed *config.InstalledAgent, valid []generate.DiscoveredFile,
	selected []string, lock *config.LockFile, cwd string) bool {
	selectedSet := make(map[string]bool, len(selected))
	for _, s := range selected {
		selectedSet[s] = true
	}

	changed := false
	for _, d := range valid {
		if !selectedSet[d.Type+":"+d.Name] {
			continue
		}

		switch d.Type {
		case "skill":
			installed.Skills = appendUnique(installed.Skills, d.Name)
		case "workflow":
			installed.Workflows = appendUnique(installed.Workflows, d.Name)
		case "protocol":
			installed.Protocols = appendUnique(installed.Protocols, d.Name)
		case "sensor":
			installed.Sensors = appendUnique(installed.Sensors, d.Name)
		case "routine":
			installed.Routines = appendUnique(installed.Routines, d.Name)
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

		changed = true
	}
	return changed
}
