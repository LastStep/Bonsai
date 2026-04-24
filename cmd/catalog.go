package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"

	"github.com/LastStep/Bonsai/internal/catalog"
	"github.com/LastStep/Bonsai/internal/generate"
	"github.com/LastStep/Bonsai/internal/tui"
	"github.com/LastStep/Bonsai/internal/tui/catalogflow"
)

func init() {
	rootCmd.AddCommand(catalogCmd)
	catalogCmd.Flags().StringP("agent", "a", "", "Filter to items compatible with this agent type")
	catalogCmd.Flags().Bool("json", false, "Output catalog as JSON (agent-consumable, non-interactive)")
}

var catalogCmd = &cobra.Command{
	Use:   "catalog",
	Short: "Browse available abilities — agents, skills, workflows, protocols, and routines.",
	RunE:  runCatalog,
}

// runCatalog is the catalog command entry point. TTY invocations open
// the cinematic BubbleTea browser (catalogflow.BrowserStage);
// non-TTY invocations (pipes, CI, `> out.txt`) fall back to the
// static-render path so piped output stays clean and ANSI-free.
//
// Plan 31 Phase G: the --json flag short-circuits both paths and emits a
// stable JSON catalog snapshot to stdout, reusing generate.SerializeCatalog
// (single source of truth with WriteCatalogSnapshot, Plan 31 Phase C).
// --json honours the -a <agent> filter the same way the TTY + static paths
// do — abilities with agents: that don't match the filter are excluded.
func runCatalog(cmd *cobra.Command, args []string) error {
	cat := loadCatalog()
	agentFilter, _ := cmd.Flags().GetString("agent")

	if jsonOut, _ := cmd.Flags().GetBool("json"); jsonOut {
		return renderCatalogJSON(cat, agentFilter)
	}

	if !isatty.IsTerminal(os.Stdout.Fd()) {
		return renderCatalogStatic(cat, agentFilter)
	}

	// cwd feeds the header's right-row-2 so the browser shows where it
	// was invoked from (catalog is global but the path anchor is still a
	// useful breadcrumb for multi-terminal users).
	cwd := mustCwd()
	stage := catalogflow.NewBrowser(cat, agentFilter, cwd)
	if _, err := tea.NewProgram(stage, tea.WithAltScreen()).Run(); err != nil {
		return fmt.Errorf("catalog browser: %w", err)
	}
	return nil
}

// renderCatalogJSON serializes the catalog to stdout as JSON. Respects the
// -a agent filter: when set, abilities not compatible with that agent are
// excluded. Plan 31 Phase G — agent-consumable output for CI / scripts /
// downstream tooling. Scaffolding is intentionally excluded from the JSON
// shape (it's project-config data, not catalog data — the stable contract
// lives in generate.CatalogSnapshot).
//
// Single source of truth with WriteCatalogSnapshot (Plan 31 Phase C).
func renderCatalogJSON(cat *catalog.Catalog, agentFilter string) error {
	filtered := filterCatalog(cat, agentFilter)
	data, err := generate.SerializeCatalog(filtered, Version)
	if err != nil {
		return fmt.Errorf("serialize catalog: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

// filterCatalog returns a catalog view with abilities reduced to those
// compatible with agentFilter. When agentFilter is "" the original catalog
// is returned verbatim. The returned Catalog is a shallow-copied value —
// lookup maps (skillsByName, etc.) are NOT rebuilt because the JSON
// serializer only iterates the slices.
func filterCatalog(cat *catalog.Catalog, agentFilter string) *catalog.Catalog {
	if agentFilter == "" || cat == nil {
		return cat
	}
	out := &catalog.Catalog{
		Agents:      cat.Agents,
		Skills:      cat.SkillsFor(agentFilter),
		Workflows:   cat.WorkflowsFor(agentFilter),
		Protocols:   cat.ProtocolsFor(agentFilter),
		Sensors:     cat.SensorsFor(agentFilter),
		Routines:    cat.RoutinesFor(agentFilter),
		Scaffolding: cat.Scaffolding,
	}
	return out
}

// renderCatalogStatic renders the seven catalog sections as a flat,
// one-shot block — the pre-Plan-28 output preserved verbatim for
// non-TTY invocations (piped output, CI consumers). The TTY path
// launches the BubbleTea browser instead.
func renderCatalogStatic(cat *catalog.Catalog, agentFilter string) error {
	// Agents
	tui.SectionHeader(fmt.Sprintf("Agents (%d)", len(cat.Agents)))
	var agentRows [][]string
	for _, a := range cat.Agents {
		agentRows = append(agentRows, []string{a.DisplayName, a.Description})
	}
	tui.CatalogTable([]string{"Name", "Description"}, agentRows)

	// Skills
	skills := cat.Skills
	if agentFilter != "" {
		skills = cat.SkillsFor(agentFilter)
	}
	skillsSuffix := fmt.Sprintf(" (%d)", len(skills))
	if agentFilter != "" {
		skillsSuffix = fmt.Sprintf(" (%d for %s)", len(skills), agentFilter)
	}
	tui.SectionHeader("Skills" + skillsSuffix)
	var skillRows [][]string
	for _, s := range skills {
		skillRows = append(skillRows, []string{s.DisplayName, s.Description, s.Agents.String(), s.Required.String()})
	}
	tui.CatalogTable([]string{"Name", "Description", "Agents", "Required"}, skillRows)

	// Workflows
	workflows := cat.Workflows
	if agentFilter != "" {
		workflows = cat.WorkflowsFor(agentFilter)
	}
	workflowsSuffix := fmt.Sprintf(" (%d)", len(workflows))
	if agentFilter != "" {
		workflowsSuffix = fmt.Sprintf(" (%d for %s)", len(workflows), agentFilter)
	}
	tui.SectionHeader("Workflows" + workflowsSuffix)
	var wfRows [][]string
	for _, w := range workflows {
		wfRows = append(wfRows, []string{w.DisplayName, w.Description, w.Agents.String(), w.Required.String()})
	}
	tui.CatalogTable([]string{"Name", "Description", "Agents", "Required"}, wfRows)

	// Protocols
	protocols := cat.Protocols
	if agentFilter != "" {
		protocols = cat.ProtocolsFor(agentFilter)
	}
	protocolsSuffix := fmt.Sprintf(" (%d)", len(protocols))
	if agentFilter != "" {
		protocolsSuffix = fmt.Sprintf(" (%d for %s)", len(protocols), agentFilter)
	}
	tui.SectionHeader("Protocols" + protocolsSuffix)
	var protoRows [][]string
	for _, p := range protocols {
		protoRows = append(protoRows, []string{p.DisplayName, p.Description, p.Agents.String(), p.Required.String()})
	}
	tui.CatalogTable([]string{"Name", "Description", "Agents", "Required"}, protoRows)

	// Sensors
	sensors := cat.Sensors
	if agentFilter != "" {
		sensors = cat.SensorsFor(agentFilter)
	}
	sensorsSuffix := fmt.Sprintf(" (%d)", len(sensors))
	if agentFilter != "" {
		sensorsSuffix = fmt.Sprintf(" (%d for %s)", len(sensors), agentFilter)
	}
	tui.SectionHeader("Sensors" + sensorsSuffix)
	var sensorRows [][]string
	for _, s := range sensors {
		event := s.Event
		if s.Matcher != "" {
			event += " (" + s.Matcher + ")"
		}
		sensorRows = append(sensorRows, []string{s.DisplayName, s.Description, event, s.Agents.String(), s.Required.String()})
	}
	tui.CatalogTable([]string{"Name", "Description", "Event", "Agents", "Required"}, sensorRows)

	// Routines
	routines := cat.Routines
	if agentFilter != "" {
		routines = cat.RoutinesFor(agentFilter)
	}
	routinesSuffix := fmt.Sprintf(" (%d)", len(routines))
	if agentFilter != "" {
		routinesSuffix = fmt.Sprintf(" (%d for %s)", len(routines), agentFilter)
	}
	tui.SectionHeader("Routines" + routinesSuffix)
	var routineRows [][]string
	for _, r := range routines {
		routineRows = append(routineRows, []string{r.DisplayName, r.Description, r.Frequency, r.Agents.String(), r.Required.String()})
	}
	tui.CatalogTable([]string{"Name", "Description", "Frequency", "Agents", "Required"}, routineRows)

	// Scaffolding
	tui.SectionHeader(fmt.Sprintf("Scaffolding (%d)", len(cat.Scaffolding)))
	var scaffoldRows [][]string
	for _, s := range cat.Scaffolding {
		req := ""
		if s.Required {
			req = "yes"
		}
		scaffoldRows = append(scaffoldRows, []string{s.DisplayName, s.Description, req, s.Affects})
	}
	tui.CatalogTable([]string{"Name", "Description", "Required", "If Removed"}, scaffoldRows)

	tui.Blank()
	return nil
}
