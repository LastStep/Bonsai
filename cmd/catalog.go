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
	Short: "Browse available agents, skills, workflows, protocols, and routines.",
	RunE:  runCatalog,
}

func runCatalog(cmd *cobra.Command, args []string) error {
	cat := loadCatalog()
	agentFilter, _ := cmd.Flags().GetString("agent")

	// Agents
	tui.SectionHeader("Agents")
	var agentRows [][]string
	for _, a := range cat.Agents {
		agentRows = append(agentRows, []string{a.Name, a.Description})
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
		skillRows = append(skillRows, []string{s.Name, s.Description, s.Agents.String()})
	}
	tui.CatalogTable([]string{"Name", "Description", "Agents"}, skillRows)

	// Workflows
	workflows := cat.Workflows
	if agentFilter != "" {
		workflows = cat.WorkflowsFor(agentFilter)
	}
	tui.SectionHeader("Workflows" + suffix)
	var wfRows [][]string
	for _, w := range workflows {
		wfRows = append(wfRows, []string{w.Name, w.Description, w.Agents.String()})
	}
	tui.CatalogTable([]string{"Name", "Description", "Agents"}, wfRows)

	// Protocols
	protocols := cat.Protocols
	if agentFilter != "" {
		protocols = cat.ProtocolsFor(agentFilter)
	}
	tui.SectionHeader("Protocols" + suffix)
	var protoRows [][]string
	for _, p := range protocols {
		protoRows = append(protoRows, []string{p.Name, p.Description, p.Agents.String()})
	}
	tui.CatalogTable([]string{"Name", "Description", "Agents"}, protoRows)

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
		sensorRows = append(sensorRows, []string{s.Name, s.Description, event, s.Agents.String()})
	}
	tui.CatalogTable([]string{"Name", "Description", "Event", "Agents"}, sensorRows)

	// Routines
	routines := cat.Routines
	if agentFilter != "" {
		routines = cat.RoutinesFor(agentFilter)
	}
	tui.SectionHeader("Routines" + suffix)
	var routineRows [][]string
	for _, r := range routines {
		routineRows = append(routineRows, []string{r.Name, r.Description, r.Frequency, r.Agents.String()})
	}
	tui.CatalogTable([]string{"Name", "Description", "Frequency", "Agents"}, routineRows)

	tui.Blank()
	return nil
}
