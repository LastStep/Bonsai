package addflow

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/LastStep/Bonsai/internal/catalog"
	"github.com/LastStep/Bonsai/internal/config"
	"github.com/LastStep/Bonsai/internal/tui"
	"github.com/LastStep/Bonsai/internal/tui/initflow"
)

// yieldMode distinguishes the three terminal card variants.
type yieldMode int

const (
	yieldModeSuccess yieldMode = iota
	yieldModeAllInstalled
	yieldModeTechLeadRequired
	yieldModeAddItemsDeferred
)

// YieldStage is the terminal completion card at rail position 5 (結 YIELD).
// Three variants selected at construction time:
//
//   - success          — renders the installed ability tree + 3 next-steps.
//   - all-installed    — "already full" panel + `bonsai catalog` CTA.
//   - tech-lead-req    — error panel + `bonsai init` CTA.
//
// The stage is terminal: ↵ / q / esc flip Done and the harness exits.
type YieldStage struct {
	initflow.Stage

	mode yieldMode

	// success-mode inputs.
	installed     *config.InstalledAgent
	cat           *catalog.Catalog
	isNewAgent    bool
	totalSelected int

	// all-installed-mode inputs.
	agentDef *catalog.AgentDef

	// tech-lead-required-mode inputs.
	pickedAgentType string
}

// NewYieldSuccess renders the happy-path completion card. installed is the
// InstalledAgent populated by the Grow action; cat is the loaded catalog;
// isNewAgent distinguishes new-agent vs add-items success messaging;
// totalSelected drives the add-items summary line.
func NewYieldSuccess(ctx initflow.StageContext, installed *config.InstalledAgent, cat *catalog.Catalog, isNewAgent bool, totalSelected int) *YieldStage {
	return newYield(ctx, yieldModeSuccess, func(y *YieldStage) {
		y.installed = installed
		y.cat = cat
		y.isNewAgent = isNewAgent
		y.totalSelected = totalSelected
	})
}

// NewYieldAllInstalled renders the "every ability already installed" card.
func NewYieldAllInstalled(ctx initflow.StageContext, agentDef *catalog.AgentDef) *YieldStage {
	return newYield(ctx, yieldModeAllInstalled, func(y *YieldStage) {
		y.agentDef = agentDef
	})
}

// NewYieldTechLeadRequired renders the "tech-lead must come first" card.
func NewYieldTechLeadRequired(ctx initflow.StageContext, agentType string) *YieldStage {
	return newYield(ctx, yieldModeTechLeadRequired, func(y *YieldStage) {
		y.pickedAgentType = agentType
	})
}

// NewYieldAddItemsDeferred renders the Phase 1 "add-items not yet wired"
// card. Plan 23 ships add-items in Phase 2; until then the user must
// unset BONSAI_ADD_REDESIGN to reach the legacy flow for that path.
func NewYieldAddItemsDeferred(ctx initflow.StageContext, agentDef *catalog.AgentDef) *YieldStage {
	return newYield(ctx, yieldModeAddItemsDeferred, func(y *YieldStage) {
		y.agentDef = agentDef
	})
}

func newYield(ctx initflow.StageContext, mode yieldMode, tail func(*YieldStage)) *YieldStage {
	label := StageLabels[StageIdxYield]
	base := initflow.NewStage(
		StageIdxYield,
		label,
		label.English,
		ctx.Version,
		ctx.ProjectDir,
		ctx.StationDir,
		ctx.AgentDisplay,
		ctx.StartedAt,
	)
	base.SetRailLabels(StageLabels)
	y := &YieldStage{
		Stage: base,
		mode:  mode,
	}
	tail(y)
	return y
}

// Init implements tea.Model — no cmd on entry.
func (s *YieldStage) Init() tea.Cmd { return nil }

// Update waits for ↵ / q / esc acknowledgement.
func (s *YieldStage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		s.SetSize(m.Width, m.Height)
	case tea.KeyMsg:
		switch m.String() {
		case "enter", "q", "esc":
			s.MarkDone()
			return s, nil
		}
	}
	return s, nil
}

// View renders Yield as a chromeless, vertically-centred exit card. Mirrors
// initflow.PlantedStage.View — no header/rail/footer chrome, body centred in
// the live AltScreen height, inline "↵ exit · q quit" hint. The terminal-too-
// small floor still gates via initflow.RenderMinSizeFloor.
func (s *YieldStage) View() string {
	w := s.Width()
	h := s.Height()
	if w <= 0 {
		w = 80
	}
	if h <= 0 {
		h = 24
	}
	if initflow.TerminalTooSmall(s.Width(), s.Height()) {
		return initflow.RenderMinSizeFloor(s.Width(), s.Height())
	}

	dim := initflow.DimStyle()
	hint := dim.Render("↵  exit  ·  q  quit")
	body := s.renderBody() + "\n\n" + initflow.CenterBlock(hint, w)

	rows := strings.Count(body, "\n") + 1
	topPad := (h - rows) / 2
	if topPad < 1 {
		topPad = 1
	}
	bottomPad := h - rows - topPad
	if bottomPad < 0 {
		bottomPad = 0
	}
	return strings.Repeat("\n", topPad) + body + strings.Repeat("\n", bottomPad)
}

func (s *YieldStage) renderBody() string {
	switch s.mode {
	case yieldModeAllInstalled:
		return s.renderAllInstalled()
	case yieldModeTechLeadRequired:
		return s.renderTechLeadRequired()
	case yieldModeAddItemsDeferred:
		return s.renderAddItemsDeferred()
	default:
		return s.renderSuccess()
	}
}

func (s *YieldStage) renderSuccess() string {
	dim := initflow.DimStyle()
	bark := initflow.LabelStyle()
	white := lipgloss.NewStyle().Foreground(tui.ColorAccent).Bold(true)
	leaf := lipgloss.NewStyle().Foreground(tui.ColorPrimary).Bold(true)
	value := initflow.ValueStyle()

	var heroTitle string
	if s.EnsoSafe() {
		heroTitle = leaf.Render(s.Label().Kanji + " · YIELDED")
	} else {
		heroTitle = leaf.Render("YIELDED")
	}

	agentDisplay := ""
	workspace := ""
	if s.installed != nil {
		workspace = s.installed.Workspace
		if s.cat != nil {
			if def := s.cat.GetAgent(s.installed.AgentType); def != nil {
				agentDisplay = def.DisplayName
				if agentDisplay == "" {
					agentDisplay = catalog.DisplayNameFrom(def.Name)
				}
			}
		}
	}
	if agentDisplay == "" {
		agentDisplay = "agent"
	}

	var heroSub string
	if s.isNewAgent {
		heroSub = white.Render(fmt.Sprintf("%s is rooted.", agentDisplay))
	} else {
		heroSub = white.Render(fmt.Sprintf("Grafted %d new abilities onto %s.", s.totalSelected, agentDisplay))
	}

	summaryHeader := initflow.RenderSectionHeader("SUMMARY", initflow.PanelWidth(s.Width()))
	const labelW = 14
	const indent = "  "
	totalAbilities := 0
	if s.installed != nil {
		totalAbilities = len(s.installed.Skills) + len(s.installed.Workflows) +
			len(s.installed.Protocols) + len(s.installed.Sensors) + len(s.installed.Routines)
	}
	summaryRows := []string{
		summaryHeader,
		indent + bark.Render(initflow.PadRight("AGENT", labelW)) + value.Render(agentDisplay) +
			dim.Render(" → "+workspace),
		indent + bark.Render(initflow.PadRight("ABILITIES", labelW)) + value.Render(fmt.Sprintf("%d wired", totalAbilities)),
	}
	if s.installed != nil {
		summaryRows = append(summaryRows,
			indent+strings.Repeat(" ", labelW)+dim.Render(fmt.Sprintf(
				"%d skills · %d workflows · %d protocols · %d sensors · %d routines",
				len(s.installed.Skills), len(s.installed.Workflows), len(s.installed.Protocols),
				len(s.installed.Sensors), len(s.installed.Routines),
			)))
	}

	nextHeader := initflow.RenderSectionHeader("NEXT", initflow.PanelWidth(s.Width()))
	steps := []struct {
		num, cmd, caption string
	}{
		{"1", "$ bonsai list", "see every agent + ability in one view"},
		{"2", fmt.Sprintf("$ cd %s", workspace), "open the workspace in your shell"},
		{"3", "$ claude", "say \"hi, get started\" to warm the session"},
	}
	if !s.isNewAgent {
		steps[1] = struct{ num, cmd, caption string }{"2", "$ claude", "restart the session so the new abilities load"}
		steps[2] = struct{ num, cmd, caption string }{"3", "$ bonsai catalog", "browse more abilities any time"}
	}
	nextLines := []string{nextHeader}
	for _, st := range steps {
		nextLines = append(nextLines,
			"  "+bark.Render(st.num)+"  "+white.Render(st.cmd))
		nextLines = append(nextLines,
			"     "+dim.Render(st.caption))
	}

	body := []string{
		heroTitle,
		heroSub,
		"",
		"",
		strings.Join(summaryRows, "\n"),
		"",
		"",
		strings.Join(nextLines, "\n"),
	}
	return initflow.CenterBlock(strings.Join(body, "\n"), s.Width())
}

func (s *YieldStage) renderAllInstalled() string {
	dim := initflow.DimStyle()
	bark := initflow.LabelStyle()
	white := lipgloss.NewStyle().Foreground(tui.ColorAccent).Bold(true)
	leaf := lipgloss.NewStyle().Foreground(tui.ColorPrimary).Bold(true)

	var heroTitle string
	if s.EnsoSafe() {
		heroTitle = leaf.Render(s.Label().Kanji + " · ALREADY FULL")
	} else {
		heroTitle = leaf.Render("ALREADY FULL")
	}
	agentName := "this agent"
	if s.agentDef != nil {
		agentName = s.agentDef.DisplayName
		if agentName == "" {
			agentName = catalog.DisplayNameFrom(s.agentDef.Name)
		}
	}

	intro := white.Render(fmt.Sprintf("%s already has every compatible ability installed.", agentName))
	helper := dim.Render("Nothing left to graft — browse the catalog for ideas to iterate on later.")

	nextHeader := initflow.RenderSectionHeader("NEXT", initflow.PanelWidth(s.Width()))
	nextLines := []string{
		nextHeader,
		"  " + bark.Render("1") + "  " + white.Render("$ bonsai catalog"),
		"     " + dim.Render("survey every ability bundled in the current binary"),
		"  " + bark.Render("2") + "  " + white.Render("$ bonsai list"),
		"     " + dim.Render("inspect what's already installed per agent"),
	}

	body := []string{
		heroTitle,
		intro,
		helper,
		"",
		"",
		strings.Join(nextLines, "\n"),
	}
	return initflow.CenterBlock(strings.Join(body, "\n"), s.Width())
}

func (s *YieldStage) renderTechLeadRequired() string {
	dim := initflow.DimStyle()
	bark := initflow.LabelStyle()
	white := lipgloss.NewStyle().Foreground(tui.ColorAccent).Bold(true)
	danger := lipgloss.NewStyle().Foreground(tui.ColorDanger).Bold(true)

	var heroTitle string
	if s.EnsoSafe() {
		heroTitle = danger.Render(s.Label().Kanji + " · TECH-LEAD REQUIRED")
	} else {
		heroTitle = danger.Render("TECH-LEAD REQUIRED")
	}
	agentType := s.pickedAgentType
	if agentType == "" {
		agentType = "this agent"
	}

	intro := white.Render("No tech-lead agent is installed yet.")
	helper := dim.Render(fmt.Sprintf("Bonsai roots every project on a tech-lead before grafting %q-type agents.", agentType))

	nextHeader := initflow.RenderSectionHeader("NEXT", initflow.PanelWidth(s.Width()))
	nextLines := []string{
		nextHeader,
		"  " + bark.Render("1") + "  " + white.Render("$ bonsai init"),
		"     " + dim.Render("bootstrap the project scaffold + tech-lead agent"),
		"  " + bark.Render("2") + "  " + white.Render(fmt.Sprintf("$ bonsai add %s", agentType)),
		"     " + dim.Render("re-run this flow once the tech-lead is planted"),
	}

	body := []string{
		heroTitle,
		intro,
		helper,
		"",
		"",
		strings.Join(nextLines, "\n"),
	}
	return initflow.CenterBlock(strings.Join(body, "\n"), s.Width())
}

func (s *YieldStage) renderAddItemsDeferred() string {
	dim := initflow.DimStyle()
	bark := initflow.LabelStyle()
	white := lipgloss.NewStyle().Foreground(tui.ColorAccent).Bold(true)
	warn := lipgloss.NewStyle().Foreground(tui.ColorWarning).Bold(true)

	var heroTitle string
	if s.EnsoSafe() {
		heroTitle = warn.Render(s.Label().Kanji + " · ADD-ITEMS COMING IN PHASE 2")
	} else {
		heroTitle = warn.Render("ADD-ITEMS COMING IN PHASE 2")
	}
	agentName := "this agent"
	if s.agentDef != nil {
		agentName = s.agentDef.DisplayName
		if agentName == "" {
			agentName = catalog.DisplayNameFrom(s.agentDef.Name)
		}
	}

	intro := white.Render(fmt.Sprintf("%s already exists — adding more abilities to it is Phase 2 work.", agentName))
	helper := dim.Render("Phase 1 of the cinematic flow only wires the new-agent path. Use the legacy flow until Phase 2 lands.")

	nextHeader := initflow.RenderSectionHeader("NEXT", initflow.PanelWidth(s.Width()))
	nextLines := []string{
		nextHeader,
		"  " + bark.Render("1") + "  " + white.Render("$ unset BONSAI_ADD_REDESIGN"),
		"     " + dim.Render("drop the cinematic-flow gate for this shell"),
		"  " + bark.Render("2") + "  " + white.Render("$ bonsai add"),
		"     " + dim.Render("re-run via the legacy flow — add-items works there today"),
	}

	body := []string{
		heroTitle,
		intro,
		helper,
		"",
		"",
		strings.Join(nextLines, "\n"),
	}
	return initflow.CenterBlock(strings.Join(body, "\n"), s.Width())
}

// Result returns nil — Yield is the terminal stage.
func (s *YieldStage) Result() any { return nil }

// Reset clears the completion flag.
func (s *YieldStage) Reset() tea.Cmd {
	s.ClearDone()
	return nil
}
