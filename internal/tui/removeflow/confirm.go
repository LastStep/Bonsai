package removeflow

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/LastStep/Bonsai/internal/tui"
	"github.com/LastStep/Bonsai/internal/tui/initflow"
)

// ConfirmStage is the chromeless full-screen destructive-gate at rail
// position 2 (確 CONFIRM). Renders a centered prompt + UPROOT / BACK buttons.
// Default focus is BACK — the destructive action is opt-in, matching the
// addflow.ObserveStage pattern but with inverted defaults (remove is
// destructive).
//
// Keystrokes:
//   - tab / ← → / h / l   toggle focus
//   - y / Y               confirm uproot
//   - n / N               cancel
//   - ↵                   commit focused button
//   - esc                 back (harness pops cursor)
//
// Result: bool. true → proceed to action + Yield; false → abort.
type ConfirmStage struct {
	initflow.Stage

	// Prompt content — stamped at ctor time so the stage renders even if
	// no prior stages populated state (tests).
	heading string // e.g. "Uproot Backend?"
	detail  string // e.g. "Removes the agent and everything it has installed."
	caption string // e.g. "This is destructive — backed-up files are reversible"

	confirmed bool
	btnFocus  int // 0 = BACK (default), 1 = UPROOT
}

// NewConfirmStage constructs a chromeless confirm gate.
func NewConfirmStage(ctx initflow.StageContext, heading, detail, caption string) *ConfirmStage {
	label := StageLabels[StageIdxConfirm]
	base := initflow.NewStage(
		StageIdxConfirm,
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
	return &ConfirmStage{
		Stage:    base,
		heading:  heading,
		detail:   detail,
		caption:  caption,
		btnFocus: 0, // default BACK — destructive action opt-in
	}
}

// Chromeless reports true so the harness yields View() verbatim without its
// default header/footer. Matches addflow.ConflictsStage pattern.
func (s *ConfirmStage) Chromeless() bool { return true }

// Init implements tea.Model — no cmd on entry.
func (s *ConfirmStage) Init() tea.Cmd { return nil }

// Update handles focus toggle + y/n shortcuts + Enter commit.
func (s *ConfirmStage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

// View renders the chromeless full-screen frame. Mirrors ConflictsStage and
// YieldStage layouts: body centered vertically in the AltScreen.
func (s *ConfirmStage) View() string {
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

	body := s.renderBody()
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

func (s *ConfirmStage) renderBody() string {
	dim := initflow.DimStyle()
	white := lipgloss.NewStyle().Foreground(tui.ColorAccent).Bold(true)
	danger := lipgloss.NewStyle().Foreground(tui.ColorDanger).Bold(true)

	var titleRow string
	if s.EnsoSafe() {
		titleRow = danger.Render(s.Label().Kanji + " · CONFIRM")
	} else {
		titleRow = danger.Render("CONFIRM")
	}

	heading := s.heading
	if heading == "" {
		heading = "Proceed?"
	}
	detail := s.detail

	panelW := initflow.PanelWidth(s.Width())
	divider := initflow.RenderSectionHeader("COMMIT", panelW)

	intro := []string{
		titleRow,
		white.Render(heading),
	}
	if detail != "" {
		intro = append(intro, dim.Render(detail))
	}

	cta := s.renderCTA()
	hint := s.renderKeyHints()

	body := []string{
		strings.Join(intro, "\n"),
		"",
		"",
		divider,
		"",
		cta,
	}
	if s.caption != "" {
		body = append(body, "", dim.Render(s.caption))
	}
	body = append(body, "", hint)
	return initflow.CenterBlock(strings.Join(body, "\n"), s.Width())
}

func (s *ConfirmStage) renderCTA() string {
	danger := lipgloss.NewStyle().Foreground(tui.ColorDanger).Bold(true)
	muted := lipgloss.NewStyle().Foreground(tui.ColorMuted)
	accent := lipgloss.NewStyle().Foreground(tui.ColorAccent).Bold(true)

	backLabel := "[ BACK ]"
	uprootLabel := "[ ⏎  UPROOT ]"
	if !s.EnsoSafe() {
		uprootLabel = "[ Enter  UPROOT ]"
	}

	var backBtn, uprootBtn string
	if s.btnFocus == 0 {
		backBtn = accent.Render(backLabel)
		uprootBtn = muted.Render(uprootLabel)
	} else {
		backBtn = muted.Render(backLabel)
		uprootBtn = danger.Render(uprootLabel)
	}

	return "  " + backBtn + "   " + uprootBtn
}

func (s *ConfirmStage) renderKeyHints() string {
	dim := initflow.DimStyle()
	hint := "tab toggle  ·  y/n uproot/back  ·  ↵ commit  ·  esc back  ·  ctrl-c quit"
	return dim.Render(hint)
}

// Result returns the user's pick as a bool.
func (s *ConfirmStage) Result() any { return s.confirmed }

// Reset clears completion + confirmation but preserves focus so re-entry
// after an esc-back lands where the user left off.
func (s *ConfirmStage) Reset() tea.Cmd {
	s.ClearDone()
	s.confirmed = false
	return nil
}
