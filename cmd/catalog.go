package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"

	"github.com/LastStep/Bonsai/internal/catalog"
	"github.com/LastStep/Bonsai/internal/tui"
	"github.com/LastStep/Bonsai/internal/tui/catalogflow"
)

func init() {
	rootCmd.AddCommand(catalogCmd)
	catalogCmd.Flags().StringP("agent", "a", "", "Filter to items compatible with this agent type")
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
func runCatalog(cmd *cobra.Command, args []string) error {
	cat := loadCatalog()
	agentFilter, _ := cmd.Flags().GetString("agent")

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
