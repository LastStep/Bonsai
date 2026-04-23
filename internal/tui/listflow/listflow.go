// Package listflow renders the cinematic `bonsai list` surface — a
// static, non-interactive string output composed of the shared initflow
// chrome (header + min-size floor) plus per-agent panels and a counts
// footer. Unlike catalogflow/guideflow, this package has no BubbleTea
// model: RenderAll is a pure function returning the complete rendered
// output ready for a single fmt.Print call from cmd/list.go.
//
// Layout (top to bottom):
//
//	<initflow header — action "LIST", rightLabel "">
//	<optional scaffolding row, muted>
//	<per-agent panel + workspace tree/hint stack, alphabetical>
//	<muted counts footer — agents · skills · workflows · protocols · sensors · routines>
//
// When the terminal is below the initflow min-size floor (70×20), the
// output collapses to RenderMinSizeFloor and nothing else.
package listflow

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/LastStep/Bonsai/internal/catalog"
	"github.com/LastStep/Bonsai/internal/config"
	"github.com/LastStep/Bonsai/internal/tui"
	"github.com/LastStep/Bonsai/internal/tui/initflow"
)

// RenderAll builds the complete `bonsai list` output as a single string.
// version and projectDir feed the initflow header. termW / termH are the
// live terminal dims — below the min-size floor, the function short-circuits
// to RenderMinSizeFloor.
//
// cat is consulted for display-name lookups on installed abilities; when
// nil, callers fall back to DisplayNameFrom (still renders, just without
// catalog-side overrides).
func RenderAll(cfg *config.ProjectConfig, cat *catalog.Catalog, version, projectDir string, termW, termH int) string {
	if initflow.TerminalTooSmall(termW, termH) {
		return initflow.RenderMinSizeFloor(termW, termH)
	}

	width := termW
	if width <= 0 {
		width = 80
	}

	var b strings.Builder

	header := initflow.RenderHeader(version, projectDir, "LIST", "", width, initflow.WideCharSafe())
	b.WriteString(header)
	b.WriteString("\n\n")

	// Scaffolding row — single muted line above the agent panels. Only
	// rendered when scaffolding is non-empty; a zero-scaffolding project
	// collapses straight to the agent section.
	if cfg != nil && len(cfg.Scaffolding) > 0 {
		names := make([]string, len(cfg.Scaffolding))
		for i, s := range cfg.Scaffolding {
			names[i] = catalog.DisplayNameFrom(s)
		}
		muted := lipgloss.NewStyle().Foreground(tui.ColorMuted)
		b.WriteString("  " + muted.Render("Scaffolding: "+strings.Join(names, ", ")))
		b.WriteString("\n\n")
	}

	// Empty-state short-circuit — when no agents are installed, show the
	// EmptyPanel CTA + zero counts and stop. Scaffolding may still have
	// rendered above when the user ran `bonsai init` but hasn't added any
	// agents yet; that row stays.
	if cfg == nil || len(cfg.Agents) == 0 {
		b.WriteString(renderEmptyPanel("No agents installed.\nRun bonsai add to get started."))
		b.WriteString("\n\n")
		b.WriteString(renderCountsFooter(0, 0, 0, 0, 0, 0))
		b.WriteString("\n")
		return b.String()
	}

	// Per-agent panels — alphabetical by agent name. Panels and trees are
	// stacked with a blank line between agents for visual breathing room.
	agentNames := make([]string, 0, len(cfg.Agents))
	for name := range cfg.Agents {
		agentNames = append(agentNames, name)
	}
	sort.Strings(agentNames)

	for i, name := range agentNames {
		agent := cfg.Agents[name]
		b.WriteString(RenderAgentPanel(name, agent, cat, projectDir))
		if i < len(agentNames)-1 {
			b.WriteString("\n")
		}
	}

	// Counts footer — aggregate totals across all installed agents.
	totalSkills, totalWorkflows, totalProtocols, totalSensors, totalRoutines := 0, 0, 0, 0, 0
	for _, a := range cfg.Agents {
		totalSkills += len(a.Skills)
		totalWorkflows += len(a.Workflows)
		totalProtocols += len(a.Protocols)
		totalSensors += len(a.Sensors)
		totalRoutines += len(a.Routines)
	}

	b.WriteString("\n")
	b.WriteString(renderCountsFooter(
		len(cfg.Agents),
		totalSkills, totalWorkflows, totalProtocols, totalSensors, totalRoutines,
	))
	b.WriteString("\n")

	return b.String()
}

// renderEmptyPanel builds the bordered "no agents installed" panel as a
// string. Mirrors tui.EmptyPanel's visual (same PanelEmpty style) but
// returns rather than fmt.Println'ing so it can compose into RenderAll's
// single-string output. Every line of the rendered panel is indented two
// columns to match the rail/panel inset used elsewhere in the flow.
func renderEmptyPanel(msg string) string {
	rendered := tui.PanelEmpty.Render(msg)
	lines := strings.Split(rendered, "\n")
	for i, l := range lines {
		lines[i] = "  " + l
	}
	return strings.Join(lines, "\n")
}

// renderCountsFooter renders the aggregate footer line. GlyphDot (mid-dot)
// separates each count; pluralize matches the current cmd/list.go output
// (no change to wording across the Plan 28 Phase 2 rewrite).
func renderCountsFooter(agents, skills, workflows, protocols, sensors, routines int) string {
	muted := lipgloss.NewStyle().Foreground(tui.ColorMuted)
	parts := []string{
		pluralize(agents, "agent", "agents"),
		pluralize(skills, "skill", "skills"),
		pluralize(workflows, "workflow", "workflows"),
		pluralize(protocols, "protocol", "protocols"),
		pluralize(sensors, "sensor", "sensors"),
		pluralize(routines, "routine", "routines"),
	}
	return "  " + muted.Render(strings.Join(parts, " "+tui.GlyphDot+" "))
}

// pluralize picks singular or plural based on count.
func pluralize(n int, singular, plural string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, singular)
	}
	return fmt.Sprintf("%d %s", n, plural)
}
