package addflow

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/LastStep/Bonsai/internal/catalog"
	"github.com/LastStep/Bonsai/internal/tui"
	"github.com/LastStep/Bonsai/internal/tui/initflow"
)

// ObserveStage is the pre-write "one last look" stage at rail position 2
// (観 OBSERVE). Composes a read-only summary of the Select / Ground /
// Branches results into a single review panel with a GRAFT / BACK CTA.
//
// Result:
//   - true  → user confirmed GRAFT — harness proceeds to Grow.
//   - false → user cancelled — flow exits without writes.
//
// Prior inputs arrive via SetPrior so the stage can render fresh summaries
// even after an Esc-back edit pass upstream.
type ObserveStage struct {
	initflow.Stage

	cat *catalog.Catalog

	agent     string         // machine name — prev[0]
	workspace string         // resolved path — prev[1] (tech-lead auto-fills)
	graft     BranchesResult // prev[2]

	// agentDef + agentDisplay resolved from cat on each SetPrior.
	agentDef     *catalog.AgentDef
	agentDisplay string

	confirmed bool
	btnFocus  int // 0 = BACK, 1 = GRAFT. Default GRAFT (1).
}

// NewObserveStage constructs the Observe stage.
func NewObserveStage(ctx initflow.StageContext, cat *catalog.Catalog) *ObserveStage {
	label := StageLabels[StageIdxObserve]
	base := initflow.NewStage(
		StageIdxObserve,
		label,
		label.English,
		ctx.Version,
		ctx.ProjectDir,
		ctx.StationDir,
		ctx.AgentDisplay,
		ctx.StartedAt,
	)
	base.ApplyContextHeader(ctx)
	base.SetRailLabels(StageLabels)
	return &ObserveStage{
		Stage:    base,
		cat:      cat,
		btnFocus: 1,
	}
}

// SetDefaultWorkspace seeds a fallback workspace value rendered when the
// prior-results snapshot does not carry one. Used on the add-items branch,
// where Ground is skipped — the installed agent's existing workspace is
// stamped into the stage before SetPrior runs so the Observe panel reads
// the durable workspace rather than "—".
func (s *ObserveStage) SetDefaultWorkspace(ws string) { s.workspace = ws }

// SetPrior captures Select / Ground / Graft results. Called on entry and
// after every Esc-back edit.
//
// The add-flow splices differ in shape across branches — new-agent writes
// (agent, workspace, graft, ...) while add-items writes (agent, graft, ...)
// because Ground is skipped. Rather than hard-code positional indices (which
// break the add-items branch), SetPrior scans prev[] for the first value of
// each expected type. BranchesResult is uniquely typed, so the match is
// unambiguous; agent / workspace are both strings but agent is always
// prev[0] (from SelectStage) so we grab it first and treat any later string
// as the workspace.
func (s *ObserveStage) SetPrior(prev []any) {
	var firstString, secondString string
	var stringIdx int
	var graftSet bool
	for _, v := range prev {
		switch x := v.(type) {
		case string:
			if stringIdx == 0 {
				firstString = x
			} else if stringIdx == 1 {
				secondString = x
			}
			stringIdx++
		case BranchesResult:
			s.graft = x
			graftSet = true
		}
	}
	s.agent = firstString
	// Workspace only comes from Ground (new-agent branch). On add-items the
	// second string slot does not exist — preserve whatever the splicer
	// pre-stamped via SetDefaultWorkspace (installedAgent.Workspace) rather
	// than clobbering it with "".
	if secondString != "" {
		s.workspace = secondString
	}
	if !graftSet {
		s.graft = BranchesResult{}
	}
	if s.agent != "" && s.cat != nil {
		s.agentDef = s.cat.GetAgent(s.agent)
		if s.agentDef != nil {
			s.agentDisplay = s.agentDef.DisplayName
			if s.agentDisplay == "" {
				s.agentDisplay = catalog.DisplayNameFrom(s.agentDef.Name)
			}
		}
	}
}

// Init implements tea.Model — no cmd on entry.
func (s *ObserveStage) Init() tea.Cmd { return nil }

// Update handles button toggle + Enter confirm + y/n shortcuts.
func (s *ObserveStage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		s.SetSize(m.Width, m.Height)
	case tea.KeyMsg:
		switch m.String() {
		case "tab", "right", "l", "shift+tab", "left", "h":
			s.btnFocus = (s.btnFocus + 1) % 2
		case "y", "Y":
			s.confirmed = true
			s.MarkDone()
			return s, nil
		case "n", "N":
			s.confirmed = false
			s.MarkDone()
			return s, nil
		case "enter":
			s.confirmed = s.btnFocus == 1
			s.MarkDone()
			return s, nil
		}
	}
	return s, nil
}

// View composes the body inside the shared frame.
func (s *ObserveStage) View() string {
	return s.RenderFrame(s.renderBody(), s.keyHints())
}

func (s *ObserveStage) keyHints() []initflow.KeyHint {
	return []initflow.KeyHint{
		{Key: "↵", Desc: "confirm"},
		{Key: "tab", Desc: "toggle"},
		{Key: "y/n", Desc: "graft / cancel"},
		{Key: "esc", Desc: "back"},
		{Key: "ctrl-c", Desc: "quit"},
	}
}

func (s *ObserveStage) renderBody() string {
	bark := initflow.LabelStyle()
	white := lipgloss.NewStyle().Foreground(tui.ColorAccent).Bold(true)
	dim := initflow.DimStyle()

	var title string
	if s.EnsoSafe() {
		title = bark.Render(s.Label().Kanji) + " " + white.Render(s.Label().English)
	} else {
		title = white.Render(s.Label().English)
	}

	intro := strings.Join([]string{
		title,
		white.Render("One last look before grafting."),
		dim.Render("Review your picks — ↵ grafts, n cancels."),
	}, "\n")

	agentBlock := s.renderAgentBlock()
	abilitiesBlock := s.renderAbilitiesBlock()
	cta := s.renderCTA()

	body := []string{
		intro,
		"",
		"",
		agentBlock,
		"",
		"",
		abilitiesBlock,
		"",
		"",
		cta,
	}
	return initflow.CenterBlock(strings.Join(body, "\n"), s.Width())
}

func (s *ObserveStage) renderAgentBlock() string {
	bark := initflow.LabelStyle()
	value := initflow.ValueStyle()

	panelW := initflow.PanelWidth(s.Width())
	header := initflow.RenderSectionHeader("AGENT", panelW)

	name := s.agentDisplay
	if name == "" {
		name = s.agent
	}
	if name == "" {
		name = "—"
	}
	workspace := s.workspace
	if workspace == "" {
		workspace = "—"
	}
	agentType := s.agent
	if agentType == "" {
		agentType = "—"
	}

	const labelW = 14
	const indent = "  "
	rows := []string{
		header,
		indent + bark.Render(initflow.PadRight("NAME", labelW)) + value.Render(name),
		indent + bark.Render(initflow.PadRight("WORKSPACE", labelW)) + value.Render(workspace),
		indent + bark.Render(initflow.PadRight("TYPE", labelW)) + value.Render(agentType),
	}
	return strings.Join(rows, "\n")
}

func (s *ObserveStage) renderAbilitiesBlock() string {
	bark := initflow.LabelStyle()
	dim := initflow.DimStyle()
	value := initflow.ValueStyle()

	panelW := initflow.PanelWidth(s.Width())
	header := initflow.RenderSectionHeader("ABILITIES", panelW)

	const labelW = 14
	const indent = "  "

	renderRow := func(label string, items []string) string {
		row := indent + bark.Render(initflow.PadRight(label, labelW))
		if len(items) == 0 {
			return row + dim.Render("(none)")
		}
		listStr := strings.Join(items, ", ")
		const maxList = 36
		if lipgloss.Width(listStr) > maxList {
			runes := []rune(listStr)
			cut := maxList - 1
			if cut > len(runes) {
				cut = len(runes)
			}
			listStr = string(runes[:cut]) + "…"
		}
		return row + value.Render(fmt.Sprintf("%d", len(items))) + dim.Render("  "+listStr)
	}

	rows := []string{
		header,
		renderRow("SKILLS", s.graft.Skills),
		renderRow("WORKFLOWS", s.graft.Workflows),
		renderRow("PROTOCOLS", s.graft.Protocols),
		renderRow("SENSORS", s.graft.Sensors),
		renderRow("ROUTINES", s.graft.Routines),
	}
	return strings.Join(rows, "\n")
}

func (s *ObserveStage) renderCTA() string {
	dim := initflow.DimStyle()
	bark := lipgloss.NewStyle().Foreground(tui.ColorSecondary).Bold(true)
	leaf := lipgloss.NewStyle().Foreground(tui.ColorPrimary)
	muted := lipgloss.NewStyle().Foreground(tui.ColorMuted)
	accent := lipgloss.NewStyle().Foreground(tui.ColorAccent).Bold(true)

	total := s.graft.Total()
	ws := s.workspace
	if ws == "" {
		ws = "—"
	}

	backLabel := "[ BACK ]"
	graftLabel := fmt.Sprintf("[ ⏎  GRAFT ~%d items ]", total)
	if !s.EnsoSafe() {
		graftLabel = fmt.Sprintf("[ Enter  GRAFT ~%d items ]", total)
	}

	var backBtn, graftBtn string
	if s.btnFocus == 0 {
		backBtn = accent.Render(backLabel)
		graftBtn = muted.Render(graftLabel)
	} else {
		backBtn = muted.Render(backLabel)
		graftBtn = leaf.Bold(true).Render(graftLabel)
	}

	line1 := leaf.Render("Graft into ") + bark.Render(ws)
	line2 := dim.Render("Conflicts will prompt — nothing overwritten silently.")
	line3 := backBtn + "   " + graftBtn
	return strings.Join([]string{line1, line2, "", line3}, "\n")
}

// Result returns a bool: true = proceed to Grow, false = cancel.
func (s *ObserveStage) Result() any { return s.confirmed }

// Reset clears completion + confirmation while preserving the captured
// prior snapshot and button focus.
func (s *ObserveStage) Reset() tea.Cmd {
	s.ClearDone()
	s.confirmed = false
	return nil
}
