package addflow

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/LastStep/Bonsai/internal/tui"
	"github.com/LastStep/Bonsai/internal/tui/initflow"
)

// GroundStage collects the workspace directory for a new non-tech-lead
// agent at rail position 1 (地 GROUND). Single textinput with Vessel-style
// focus underline + unique-workspace validator. For the tech-lead agent
// type the stage auto-completes with the resolved DocsPath (or "station/")
// without any keystroke — AutoComplete() drives the harness's skip-past.
//
// Result: resolved workspace string (always trailing-slashed + cleaned).
type GroundStage struct {
	initflow.Stage

	input              textinput.Model
	agentType          string
	defaultWorkspace   string            // pre-filled value / tech-lead override
	existingWorkspaces map[string]bool   // duplicate guard
	techLead           bool              // true → AutoComplete skips the stage
	validateErr        string            // inline error label under the input
	showError          bool              // only draw errors after first submit attempt
	_                  lipgloss.Position // reserved — keeps imports stable
}

// GroundContext is the ctor bundle for GroundStage. Mirrors the shape used
// by the other addflow ctors so cmd/add.go can stamp everything in
// one place.
type GroundContext struct {
	AgentType          string
	DocsPath           string          // cfg.DocsPath for tech-lead fallback
	ExistingWorkspaces map[string]bool // keys already taken by installed agents
}

// groundLabel is the kanji/kana/English triple shown in the Ground stage's
// body title. Plan 27 shrunk the rail to four visible stages, so Ground no
// longer has a rail tab — its rail index is StageIdxOffRail and the rail row
// is suppressed by the base Stage's renderFrame. The body title still reads
// from this local label so the stage retains its bonsai-metaphor identity.
var groundLabel = initflow.StageLabel{Kanji: "地", Kana: "じ", English: "GROUND"}

// NewGroundStage constructs the Ground stage. When agentType is "tech-lead"
// the stage is in auto-complete mode — AutoComplete() returns (workspace,
// true) and the harness's NewLazy wrapper should skip it.
func NewGroundStage(ctx initflow.StageContext, gc GroundContext) *GroundStage {
	label := groundLabel
	base := initflow.NewStage(
		StageIdxOffRail,
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

	ti := textinput.New()
	ti.Prompt = "❯ "
	ti.PromptStyle = lipgloss.NewStyle().Foreground(tui.ColorPrimary)
	ti.TextStyle = lipgloss.NewStyle().Foreground(tui.ColorAccent).Bold(true)
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(tui.ColorRule2)
	ti.CharLimit = 256
	ti.Width = 60
	ti.Placeholder = gc.AgentType + "/"
	ti.SetValue(gc.AgentType + "/")
	ti.CursorEnd()
	ti.Focus()

	techLead := gc.AgentType == "tech-lead"
	defaultWs := gc.DocsPath
	if defaultWs == "" {
		defaultWs = "station/"
	}

	return &GroundStage{
		Stage:              base,
		input:              ti,
		agentType:          gc.AgentType,
		defaultWorkspace:   defaultWs,
		existingWorkspaces: gc.ExistingWorkspaces,
		techLead:           techLead,
	}
}

// Init kicks the textinput cursor blink. For the tech-lead agent the stage
// is non-interactive — flip Done immediately so the harness forwards past
// the stage without a keystroke.
func (s *GroundStage) Init() tea.Cmd {
	if s.techLead {
		s.MarkDone()
		return nil
	}
	return textinput.Blink
}

// AutoComplete reports true when there is nothing for the user to change on
// this stage (tech-lead path) so the harness's Esc-back walker skips past
// it. Matches the SpinnerStep / MultiSelectStep(all required) convention.
func (s *GroundStage) AutoComplete() bool { return s.techLead }

// Update handles typing + Enter-to-validate-and-advance. Esc propagates to
// the harness (pops back to Select).
func (s *GroundStage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		s.SetSize(m.Width, m.Height)
	case tea.KeyMsg:
		if m.String() == "enter" {
			v := strings.TrimSpace(s.input.Value())
			if v == "" {
				s.validateErr = "workspace required"
				s.showError = true
				return s, nil
			}
			norm := NormaliseWorkspace(v)
			if s.existingWorkspaces[norm] {
				s.validateErr = fmt.Sprintf("workspace %q is already in use", norm)
				s.showError = true
				return s, nil
			}
			s.validateErr = ""
			s.showError = false
			s.input.SetValue(norm)
			s.MarkDone()
			return s, nil
		}
	}
	var cmd tea.Cmd
	s.input, cmd = s.input.Update(msg)
	return s, cmd
}

// Chromeless reports true so the harness yields View() verbatim without its
// default header/footer wrap. Plan 27 PR2 §C6 — GroundStage renders a
// centred off-rail form panel without the enso rail / header / footer
// chrome used by the four on-rail stages. The rail visible to the user
// stays anchored on SELECT while Ground collects the workspace.
func (s *GroundStage) Chromeless() bool { return true }

// View returns the full AltScreen frame for the Ground stage. Chromeless —
// body centred vertically, inline key-hint row below the form. Mirrors the
// layout rhythm of initflow.PlantedStage (centre body + inline hint) so the
// off-rail Ground reads as part of the same cinematic.
func (s *GroundStage) View() string {
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

	body := s.renderBody() + "\n\n" + initflow.CenterBlock(s.renderInlineHints(), w)

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

// renderInlineHints renders the key-hint row inline (replaces RenderFooter).
func (s *GroundStage) renderInlineHints() string {
	dim := initflow.DimStyle()
	return dim.Render("↵  next  ·  esc  back  ·  ctrl-c  quit")
}

func (s *GroundStage) renderBody() string {
	dim := initflow.DimStyle()
	bark := initflow.LabelStyle()
	leaf := lipgloss.NewStyle().Foreground(tui.ColorPrimary)
	white := lipgloss.NewStyle().Foreground(tui.ColorAccent).Bold(true)
	errStyle := lipgloss.NewStyle().Foreground(tui.ColorDanger)

	var title string
	if s.EnsoSafe() {
		title = bark.Render(s.Label().Kanji) + " " + white.Render(s.Label().English)
	} else {
		title = white.Render(s.Label().English)
	}

	intro := strings.Join([]string{
		title,
		white.Render("Ground the agent."),
		dim.Render("Where on disk should this agent live?"),
	}, "\n")

	divider := initflow.RenderSectionHeader("WORKSPACE", initflow.PanelWidth(s.Width()))

	const labelColW = 20
	inputW := s.Width() - labelColW - 4
	if inputW > 60 {
		inputW = 60
	}
	if inputW < 30 {
		inputW = 30
	}
	inputCellW := inputW + 4
	s.input.Width = inputW

	labelLine := bark.Render(initflow.PadRight("WORKSPACE", labelColW))
	subtitleLine := dim.Render(initflow.PadRight("e.g. backend/", labelColW))
	inputLine := lipgloss.PlaceHorizontal(inputCellW, lipgloss.Left, s.input.View())
	underline := leaf.Render(strings.Repeat("─", inputCellW))

	line1 := labelLine + inputLine
	line2 := subtitleLine + underline

	row := line1 + "\n" + line2
	if s.showError && s.validateErr != "" {
		row += "\n" + initflow.PadRight(" ", labelColW) + errStyle.Render(s.validateErr)
	}

	block := strings.Join([]string{
		intro,
		"",
		"",
		divider,
		"",
		row,
	}, "\n")
	return initflow.CenterBlock(block, s.Width())
}

// Result returns the resolved workspace string (normalised trailing slash)
// or the tech-lead default when in auto-complete mode.
func (s *GroundStage) Result() any {
	if s.techLead {
		return s.defaultWorkspace
	}
	return NormaliseWorkspace(s.input.Value())
}

// Reset clears completion so re-entry doesn't auto-advance. Preserves the
// entered value.
func (s *GroundStage) Reset() tea.Cmd {
	s.ClearDone()
	return nil
}

// NormaliseWorkspace applies the shared trim + filepath.Clean + trailing-
// slash rule used by cmd/add.go. Duplicated here (not referenced via cmd/)
// so addflow has no back-import into cmd. Phase 3 can drop the cmd copy.
func NormaliseWorkspace(s string) string {
	v := strings.TrimSpace(s)
	if v == "" {
		return ""
	}
	v = strings.TrimRight(filepath.Clean(v), "/") + "/"
	return v
}
