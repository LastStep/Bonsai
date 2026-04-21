package initflow

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/LastStep/Bonsai/internal/tui"
)

// VesselStage collects the three project-identity fields (NAME, DESCRIPTION,
// STATION) on a single page using three composed bubbles/textinput models.
// Replaces the first three legacy harness steps with a unified stage so the
// cinematic flow can own the full frame.
//
// Result() returns map[string]string{"name", "description", "station"} so
// runInitRedesign indexes per-stage rather than per-field — keeps prev[]
// indexing stable across the 4-stage design.
type VesselStage struct {
	Stage

	inputs [3]textinput.Model
	focus  int // 0..2 — which input is focused

	// showErrors is flipped on when the user attempts to submit with an
	// invalid NAME or STATION — inline error labels render under those
	// fields until the input is valid.
	showErrors bool
}

const (
	vesselIdxName = iota
	vesselIdxDescription
	vesselIdxStation
)

// defaultStationDir is the placeholder + fallback value for the STATION
// input — matches legacy runInit behaviour where empty resolved to
// "station/".
const defaultStationDir = "station/"

// NewVesselStage constructs the real Vessel stage, replacing the Phase-2 stub
// at rail position 0. Inputs default to empty; the station input shows
// "station/" as a placeholder (but does not pre-fill — empty submit falls
// back to the default).
func NewVesselStage(ctx StageContext) *VesselStage {
	label := StageLabels[0]
	base := NewStage(
		0,
		label,
		label.English,
		ctx.Version,
		ctx.ProjectDir,
		ctx.StationDir,
		ctx.AgentDisplay,
		ctx.StartedAt,
	)

	inputs := [3]textinput.Model{}
	for i := range inputs {
		ti := textinput.New()
		ti.Prompt = "❯ "
		ti.PromptStyle = lipgloss.NewStyle().Foreground(tui.ColorPrimary)
		ti.TextStyle = lipgloss.NewStyle().Foreground(tui.ColorSecondary)
		ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(tui.ColorMuted)
		ti.CharLimit = 256
		ti.Width = 48
		inputs[i] = ti
	}

	inputs[vesselIdxName].Placeholder = "my-project"
	inputs[vesselIdxDescription].Placeholder = "one line · shown in agent prompts"
	inputs[vesselIdxStation].Placeholder = defaultStationDir

	inputs[vesselIdxName].Focus()

	return &VesselStage{
		Stage:  base,
		inputs: inputs,
		focus:  vesselIdxName,
	}
}

// Init kicks the textinput cursor blink on the initially-focused input.
func (s *VesselStage) Init() tea.Cmd { return textinput.Blink }

// focusAt shifts focus to the input at idx (mod 3, positive). Called from
// Tab/Shift-Tab/↑/↓ key handling so exactly one input is focused at a time.
func (s *VesselStage) focusAt(idx int) tea.Cmd {
	idx = ((idx % len(s.inputs)) + len(s.inputs)) % len(s.inputs)
	for i := range s.inputs {
		if i == idx {
			s.inputs[i].Focus()
		} else {
			s.inputs[i].Blur()
		}
	}
	s.focus = idx
	return textinput.Blink
}

// validate returns true when every required field has a non-empty, valid
// value. Used on ↵ to gate stage advancement. Rules mirror the existing
// cmd.stationDirValidator (empty or "/" rejected for STATION) — Vessel
// enforces inline so the user corrects without leaving the stage. Empty
// STATION is treated as "use the default station/" rather than an error.
func (s *VesselStage) validate() bool {
	if strings.TrimSpace(s.inputs[vesselIdxName].Value()) == "" {
		return false
	}
	station := strings.TrimSpace(s.inputs[vesselIdxStation].Value())
	if station == "" {
		return true
	}
	return station != "/"
}

// Update handles key input for focus cycling + Enter-to-advance.
func (s *VesselStage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = m.Width
		s.height = m.Height
	case tea.KeyMsg:
		switch m.String() {
		case "tab", "down":
			return s, s.focusAt(s.focus + 1)
		case "up":
			return s, s.focusAt(s.focus - 1)
		case "shift+tab":
			// Shift+Tab on the first field: propagate to the harness so it
			// pops back one stage. Otherwise, move focus up.
			if s.focus == 0 {
				return s, nil // harness handles the pop
			}
			return s, s.focusAt(s.focus - 1)
		case "enter":
			// On the last input, attempt submit. On earlier inputs, advance
			// focus — matches the "one ↵ at the last field" idiom.
			if s.focus < len(s.inputs)-1 {
				return s, s.focusAt(s.focus + 1)
			}
			if !s.validate() {
				s.showErrors = true
				// Jump focus to the first invalid input so the user's next
				// keystroke edits the offender.
				if strings.TrimSpace(s.inputs[vesselIdxName].Value()) == "" {
					return s, s.focusAt(vesselIdxName)
				}
				return s, s.focusAt(vesselIdxStation)
			}
			s.done = true
			return s, nil
		}
	}

	// Forward all other input to the focused textinput.
	var cmd tea.Cmd
	s.inputs[s.focus], cmd = s.inputs[s.focus].Update(msg)
	return s, cmd
}

// View composes the Vessel stage body inside the shared frame.
func (s *VesselStage) View() string {
	return s.renderFrame(s.renderBody(), s.keyHints())
}

// keyHints returns the footer key row for this stage.
func (s *VesselStage) keyHints() []KeyHint {
	return []KeyHint{
		{Key: "↵", Desc: "next"},
		{Key: "tab", Desc: "cycle"},
		{Key: "esc", Desc: "back"},
		{Key: "ctrl-c", Desc: "quit"},
	}
}

// renderBody renders the stage intro + the three labelled inputs.
func (s *VesselStage) renderBody() string {
	muted := lipgloss.NewStyle().Foreground(tui.ColorMuted)
	bark := lipgloss.NewStyle().Foreground(tui.ColorSecondary).Bold(true)
	leaf := lipgloss.NewStyle().Foreground(tui.ColorPrimary).Bold(true)
	errStyle := lipgloss.NewStyle().Foreground(tui.ColorLeafDim)

	primary, secondary := s.label.Render(s.ensoSafe)
	title := leaf.Render(primary)
	if secondary != "" {
		title += muted.Render("  " + secondary)
	}
	title += muted.Render("  ·  VESSEL")

	intro := strings.Join([]string{
		title,
		bark.Render("Shape the vessel."),
		muted.Render("Every Bonsai begins with a small decision — what will this one carry?"),
	}, "\n")

	divider := muted.Render(strings.Repeat("─", 3)) + " " +
		bark.Render("FIELDS") + " " +
		muted.Render(strings.Repeat("─", 3)+" 入力 "+strings.Repeat("─", 20))

	// Per-field rows: LABEL (Bark bold) + input + optional help caption + error.
	row := func(label, help string, input *textinput.Model, errMsg string) string {
		labelCol := bark.Render(padRight(label, 14))
		inputCol := input.View()
		line := labelCol + inputCol
		helpLine := muted.Render(padRight(" ", 14) + help)
		out := line + "\n" + helpLine
		if errMsg != "" {
			errLine := errStyle.Render(padRight(" ", 14) + errMsg)
			out += "\n" + errLine
		}
		return out
	}

	nameErr := ""
	if s.showErrors && strings.TrimSpace(s.inputs[vesselIdxName].Value()) == "" {
		nameErr = "required"
	}
	stationErrMsg := ""
	if s.showErrors {
		v := strings.TrimSpace(s.inputs[vesselIdxStation].Value())
		if v == "/" {
			stationErrMsg = "must be a subdirectory like: station/"
		}
	}

	nameRow := row("NAME", "required · used as .bonsai.yaml project_name",
		&s.inputs[vesselIdxName], nameErr)
	descRow := row("DESCRIPTION", "optional · one line · shown in agent prompts",
		&s.inputs[vesselIdxDescription], "")
	stationRow := row("STATION", "where agent files live · default "+defaultStationDir,
		&s.inputs[vesselIdxStation], stationErrMsg)

	return strings.Join([]string{
		"  " + intro,
		"",
		"  " + divider,
		"",
		"  " + strings.ReplaceAll(nameRow, "\n", "\n  "),
		"",
		"  " + strings.ReplaceAll(descRow, "\n", "\n  "),
		"",
		"  " + strings.ReplaceAll(stationRow, "\n", "\n  "),
	}, "\n")
}

// Result returns the three collected fields as a single map keyed by
// "name" / "description" / "station". Description is returned verbatim
// (may be empty — it is optional). Station falls back to "station/" when
// empty, and is normalised to a trailing slash so downstream callers get a
// path-shaped value. The shape matches cmd.normaliseDocsPath so the caller
// can treat the result as already-normalised.
func (s *VesselStage) Result() any {
	name := strings.TrimSpace(s.inputs[vesselIdxName].Value())
	description := strings.TrimSpace(s.inputs[vesselIdxDescription].Value())
	station := strings.TrimSpace(s.inputs[vesselIdxStation].Value())
	if station == "" {
		station = defaultStationDir
	}
	if !strings.HasSuffix(station, "/") {
		station += "/"
	}
	return map[string]string{
		"name":        name,
		"description": description,
		"station":     station,
	}
}

// Reset clears the completion flag so re-entry behaves correctly but keeps
// the entered values so the user's work is preserved when popping back.
func (s *VesselStage) Reset() tea.Cmd {
	s.done = false
	return s.focusAt(s.focus)
}
