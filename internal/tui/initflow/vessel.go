package initflow

import (
	"path/filepath"
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
// cmd.runInit indexes per-stage rather than per-field — keeps prev[]
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
	base.applyContextHeader(ctx)

	inputs := [3]textinput.Model{}
	for i := range inputs {
		ti := textinput.New()
		ti.Prompt = "❯ "
		ti.PromptStyle = lipgloss.NewStyle().Foreground(tui.ColorPrimary)
		ti.TextStyle = lipgloss.NewStyle().Foreground(tui.ColorAccent).Bold(true)
		ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(tui.ColorRule2)
		ti.CharLimit = 256
		ti.Width = 60
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
// value. Used on ↵ to gate stage advancement. Rule: STATION rejects empty
// or "/" — Vessel enforces inline so the user corrects without leaving the
// stage. Empty STATION is treated as "use the default station/" rather than
// an error.
func (s *VesselStage) validate() bool {
	if strings.TrimSpace(s.inputs[vesselIdxName].Value()) == "" {
		return false
	}
	station := strings.TrimSpace(s.inputs[vesselIdxStation].Value())
	if station == "" {
		return true
	}
	if station == "/" {
		return false
	}
	// After normalising to trailing slash, reject absolute + path-escape.
	// Project-relative only — defence against accidental writes outside
	// the project root when the user types "../..." or a rooted path.
	norm := station
	if !strings.HasSuffix(norm, "/") {
		norm += "/"
	}
	if filepath.IsAbs(norm) {
		return false
	}
	for _, seg := range strings.Split(strings.TrimRight(norm, "/"), "/") {
		if seg == ".." {
			return false
		}
	}
	return true
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

// renderBody renders the stage intro + the three labelled inputs. The body
// is centred inside the current terminal width via centerBlock so field rows
// sit in a premium, balanced layout rather than flush-left with a 2-col gap.
func (s *VesselStage) renderBody() string {
	// Helper/informational text uses ColorRule2 (dimmer than ColorMuted) so
	// hints/subtitles/captions sit back from the foreground copy — reduces
	// visual noise per design.
	dim := lipgloss.NewStyle().Foreground(tui.ColorRule2)
	bark := lipgloss.NewStyle().Foreground(tui.ColorSecondary).Bold(true)
	leaf := lipgloss.NewStyle().Foreground(tui.ColorPrimary)
	white := lipgloss.NewStyle().Foreground(tui.ColorAccent).Bold(true)
	errStyle := lipgloss.NewStyle().Foreground(tui.ColorDanger)

	// Title: "<gold-bold>器</gold-bold> <white-bold>VESSEL</white-bold>".
	// Drops the prior kana tail and trailing "· VESSEL" duplicate per design.
	var title string
	if s.ensoSafe {
		title = bark.Render(s.label.Kanji) + " " + white.Render(s.label.English)
	} else {
		title = white.Render(s.label.English)
	}

	intro := strings.Join([]string{
		title,
		white.Render("Shape the vessel."),
		dim.Render("Three quick answers — a name, a purpose, a place to grow."),
	}, "\n")

	// Divider: green-tint left rule + Bark "FIELDS" + long dim right rule.
	// Previous mid-segment "入力" dropped per design (kana removed throughout).
	divider := leaf.Render(strings.Repeat("─", 3)) + " " +
		bark.Render("FIELDS") + " " +
		dim.Render(strings.Repeat("─", 55))

	// Per-field rows: LEFT column stacks LABEL + small subtitle; RIGHT column
	// stacks input prompt + underline. The input line is pinned to a fixed
	// cell width via lipgloss.PlaceHorizontal so line1 length stays constant
	// regardless of typed value — bubbles/textinput.View() returns Width
	// cells when empty (placeholder mode) but Width+3 cells when typed
	// (prompt + value + cursor + padding), so a naive padRight can't
	// equalise them. PlaceHorizontal pads-or-truncates to exactly inputCellW.
	const labelColW = 20
	// inputW = clamp(s.width - labelColW - 4, 30, 60). Shrinks on narrow
	// terminals so the label + input pair fits inside the current row,
	// but never below 30 cells (unreadable) or above 60 (original). The
	// same clamp flows into the underline so the focus rule tracks the
	// actual input width. (Scaffolding TODO: if catalog ever exceeds
	// ~8 fields this stage could gain a Viewport — current layout renders
	// three rows only, so no vertical scroll needed today.)
	inputW := s.width - labelColW - 4
	if inputW > 60 {
		inputW = 60
	}
	if inputW < 30 {
		inputW = 30
	}
	inputCellW := inputW + 4 // prompt(2) + cursor(1) + value safety (1)
	// Sync live textinput widths so bubbles/textinput renders at the
	// clamped width rather than its 60-cell default. Each call is cheap
	// and idempotent.
	for i := range s.inputs {
		s.inputs[i].Width = inputW
	}
	row := func(label, subtitle string, input *textinput.Model, focused bool, errMsg string) string {
		labelLine := bark.Render(padRight(label, labelColW))
		subtitleLine := dim.Render(padRight(subtitle, labelColW))

		inputLine := lipgloss.PlaceHorizontal(inputCellW, lipgloss.Left, input.View())
		// Underline: solid ─ under the input field, tinted green on focus.
		underlineStyle := dim
		if focused {
			underlineStyle = lipgloss.NewStyle().Foreground(tui.ColorPrimary)
		}
		underline := underlineStyle.Render(strings.Repeat("─", inputCellW))

		// Compose 2-col grid: LEFT labelCol, RIGHT input stack.
		line1 := labelLine + inputLine
		line2 := subtitleLine + underline

		out := line1 + "\n" + line2
		if errMsg != "" {
			out += "\n" + padRight(" ", labelColW) + errStyle.Render(errMsg)
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

	nameRow := row("NAME", "required",
		&s.inputs[vesselIdxName], s.focus == vesselIdxName, nameErr)
	descRow := row("DESCRIPTION", "optional",
		&s.inputs[vesselIdxDescription], s.focus == vesselIdxDescription, "")
	stationRow := row("STATION", "default station/",
		&s.inputs[vesselIdxStation], s.focus == vesselIdxStation, stationErrMsg)

	block := strings.Join([]string{
		intro,
		"",
		"",
		divider,
		"",
		nameRow,
		"",
		descRow,
		"",
		stationRow,
	}, "\n")

	return centerBlock(block, s.width)
}

// Result returns the three collected fields as a single map keyed by
// "name" / "description" / "station". Description is returned verbatim
// (may be empty — it is optional). Station falls back to "station/" when
// empty, and is normalised to a trailing slash so downstream callers get a
// path-shaped value already suitable for use as DocsPath.
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
