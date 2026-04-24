package updateflow

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/LastStep/Bonsai/internal/generate"
	"github.com/LastStep/Bonsai/internal/tui"
	"github.com/LastStep/Bonsai/internal/tui/hints"
	"github.com/LastStep/Bonsai/internal/tui/initflow"
)

// yieldMode distinguishes the three terminal card variants.
type yieldMode int

const (
	yieldModeSynced   yieldMode = iota // at least one change applied
	yieldModeUpToDate                  // nothing new — maintains legacy "Up to date" string contract
	yieldModeError                     // sync-action returned an error
)

// YieldStage is the terminal completion card at rail position 3 (結 YIELD).
// Three variants, all chromeless — matches addflow.YieldStage presentation
// so the update flow lands its exit with the same typographic beat.
type YieldStage struct {
	initflow.Stage

	mode yieldMode

	// Success-mode inputs.
	wr            *generate.WriteResult
	configChanged bool

	// Error-mode inputs.
	syncErr error

	// Hints block — optional; zero-value renders nothing.
	hintBlock hints.Block
}

// YieldInputs carries the data every yield variant needs at ctor time.
type YieldInputs struct {
	WriteResult   *generate.WriteResult
	ConfigChanged bool
	SyncErr       error
	HintBlock     hints.Block
}

// NewYieldStage constructs the Yield stage, auto-selecting the variant
// based on the inputs.
//
//   - SyncErr != nil              → yieldModeError
//   - no changes + no conflicts   → yieldModeUpToDate
//   - otherwise                   → yieldModeSynced
//
// Caller builds inputs from the cross-stage flow state — Sync's
// WriteResult + configChanged flag, any error bubbled up, and the
// pre-computed hints block.
func NewYieldStage(ctx initflow.StageContext, inputs YieldInputs) *YieldStage {
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
	base.ApplyContextHeader(ctx)
	base.SetRailLabels(StageLabels)

	mode := yieldModeSynced
	switch {
	case inputs.SyncErr != nil:
		mode = yieldModeError
	case !hasAnyChange(inputs.WriteResult, inputs.ConfigChanged):
		mode = yieldModeUpToDate
	}

	return &YieldStage{
		Stage:         base,
		mode:          mode,
		wr:            inputs.WriteResult,
		configChanged: inputs.ConfigChanged,
		syncErr:       inputs.SyncErr,
		hintBlock:     inputs.HintBlock,
	}
}

// hasAnyChange reports whether the WriteResult has at least one
// created/updated file OR the config was mutated during the sync.
// Matches the legacy cmd/update.go `hadChanges` check semantics.
func hasAnyChange(wr *generate.WriteResult, configChanged bool) bool {
	if configChanged {
		return true
	}
	if wr == nil {
		return false
	}
	created, updated, _, _, conflicts := wr.Summary()
	return created > 0 || updated > 0 || conflicts > 0
}

// Chromeless reports true — YieldStage renders its own centred exit card
// without the enso rail.
func (s *YieldStage) Chromeless() bool { return true }

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

// View renders the exit card. Mirrors addflow.YieldStage — chromeless,
// vertically-centred body with inline hint row below.
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
	case yieldModeError:
		return s.renderError()
	case yieldModeUpToDate:
		return s.renderUpToDate()
	default:
		return s.renderSynced()
	}
}

func (s *YieldStage) renderSynced() string {
	dim := initflow.DimStyle()
	bark := initflow.LabelStyle()
	white := lipgloss.NewStyle().Foreground(tui.ColorAccent).Bold(true)
	leaf := lipgloss.NewStyle().Foreground(tui.ColorPrimary).Bold(true)
	value := initflow.ValueStyle()

	var heroTitle string
	if s.EnsoSafe() {
		heroTitle = leaf.Render(s.Label().Kanji + " · SYNCED")
	} else {
		heroTitle = leaf.Render("SYNCED")
	}

	var heroSub string
	if s.configChanged {
		heroSub = white.Render("Custom files promoted — workspace is in sync.")
	} else {
		heroSub = white.Render("Workspace synced with the catalog.")
	}

	created, updated, _, _, conflicts := s.counts()
	heroStats := dim.Render(fmt.Sprintf(
		"%d files written · %d conflicts · lock synced",
		created+updated, conflicts,
	))

	panelW := initflow.PanelWidth(s.Width())
	summaryHeader := initflow.RenderSectionHeader("SUMMARY", panelW)
	const labelW = 14
	const indent = "  "
	summaryRows := []string{
		summaryHeader,
		indent + bark.Render(initflow.PadRight("CREATED", labelW)) + value.Render(fmt.Sprintf("%d", created)),
		indent + bark.Render(initflow.PadRight("UPDATED", labelW)) + value.Render(fmt.Sprintf("%d", updated)),
		indent + bark.Render(initflow.PadRight("CONFLICTS", labelW)) + value.Render(fmt.Sprintf("%d", conflicts)),
	}

	nextBlock := hints.Render(s.hintBlock, panelW)

	body := []string{
		heroTitle,
		heroSub,
		heroStats,
		"",
		"",
		strings.Join(summaryRows, "\n"),
	}
	if nextBlock != "" {
		body = append(body, "", "", nextBlock)
	}
	return initflow.CenterBlock(strings.Join(body, "\n"), s.Width())
}

func (s *YieldStage) renderUpToDate() string {
	dim := initflow.DimStyle()
	white := lipgloss.NewStyle().Foreground(tui.ColorAccent).Bold(true)
	leaf := lipgloss.NewStyle().Foreground(tui.ColorPrimary).Bold(true)

	var heroTitle string
	if s.EnsoSafe() {
		heroTitle = leaf.Render(s.Label().Kanji + " · UP TO DATE")
	} else {
		heroTitle = leaf.Render("UP TO DATE")
	}

	// Preserve the legacy "Up to date" string contract — test harness &
	// existing dogfood expects these exact lines to appear in output.
	intro := white.Render("Workspace is in sync with the catalog.")
	helper := dim.Render("No files needed updating.")

	panelW := initflow.PanelWidth(s.Width())
	nextBlock := hints.Render(s.hintBlock, panelW)

	body := []string{
		heroTitle,
		intro,
		helper,
	}
	if nextBlock != "" {
		body = append(body, "", "", nextBlock)
	}
	return initflow.CenterBlock(strings.Join(body, "\n"), s.Width())
}

func (s *YieldStage) renderError() string {
	dim := initflow.DimStyle()
	white := lipgloss.NewStyle().Foreground(tui.ColorAccent).Bold(true)
	danger := lipgloss.NewStyle().Foreground(tui.ColorDanger).Bold(true)

	var heroTitle string
	if s.EnsoSafe() {
		heroTitle = danger.Render(s.Label().Kanji + " · SYNC ERROR")
	} else {
		heroTitle = danger.Render("SYNC ERROR")
	}

	intro := white.Render("Update did not complete.")
	helper := dim.Render("The write pipeline reported the following error:")
	errText := danger.Render("  " + s.syncErr.Error())

	body := []string{
		heroTitle,
		intro,
		helper,
		"",
		errText,
	}
	return initflow.CenterBlock(strings.Join(body, "\n"), s.Width())
}

// counts returns WriteResult totals, or zeros when wr is nil.
func (s *YieldStage) counts() (created, updated, unchanged, skipped, conflicts int) {
	if s.wr == nil {
		return
	}
	return s.wr.Summary()
}

// Result returns nil — terminal stage.
func (s *YieldStage) Result() any { return nil }

// Reset clears the completion flag so re-entry renders fresh.
func (s *YieldStage) Reset() tea.Cmd {
	s.ClearDone()
	return nil
}
