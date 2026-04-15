package cmd

import (
	"github.com/spf13/cobra"

	"github.com/LastStep/Bonsai/internal/tui"
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

func runCatalog(cmd *cobra.Command, args []string) error {
	cat := loadCatalog()
	agentFilter, _ := cmd.Flags().GetString("agent")

	// Agents
	tui.SectionHeader("Agents")
	var agentRows [][]string
	for _, a := range cat.Agents {
		agentRows = append(agentRows, []string{a.DisplayName, a.Description})
	}
	tui.CatalogTable([]string{"Name", "Description"}, agentRows)

	suffix := ""
	if agentFilter != "" {
		suffix = " " + tui.StyleMuted.Render("(for "+agentFilter+")")
	}

	// Skills
	skills := cat.Skills
	if agentFilter != "" {
		skills = cat.SkillsFor(agentFilter)
	}
	tui.SectionHeader("Skills" + suffix)
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
	tui.SectionHeader("Workflows" + suffix)
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
	tui.SectionHeader("Protocols" + suffix)
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
	tui.SectionHeader("Sensors" + suffix)
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
	tui.SectionHeader("Routines" + suffix)
	var routineRows [][]string
	for _, r := range routines {
		routineRows = append(routineRows, []string{r.DisplayName, r.Description, r.Frequency, r.Agents.String(), r.Required.String()})
	}
	tui.CatalogTable([]string{"Name", "Description", "Frequency", "Agents", "Required"}, routineRows)

	// Scaffolding
	tui.SectionHeader("Scaffolding")
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
