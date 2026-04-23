package addflow

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/LastStep/Bonsai/internal/config"
	"github.com/LastStep/Bonsai/internal/generate"
	"github.com/LastStep/Bonsai/internal/tui"
	"github.com/LastStep/Bonsai/internal/tui/initflow"
)

// ConflictsStage is the chromeless full-screen conflict-resolution surface
// (Plan 27 PR2 §C1-C5). Rendered only when the Grow action produced at least
// one ActionConflict file — the harness gates construction on
// wr.HasConflicts(), so this stage never instantiates for clean writes.
//
// Layout: one row per conflict file in a vertical list. Each row carries:
//
//	[focus glyph] [action glyph · coloured by pick] [relative path] [action label]
//
// Below the list a batch-resolve row renders three cells (Keep all /
// Overwrite all / Backup all) that apply the chosen action to every row via
// uppercase K/O/B.
//
// The stage is chromeless — View() returns the full AltScreen frame without
// the header/enso-rail/footer chrome used by the four on-rail stages. The
// rail visible to the user (anchored on OBSERVE) stays unchanged while the
// conflict picker runs — no rail churn between Observe and Conflict.
//
// Keystrokes (plan-27 PR2 §C2 + §C4):
//
//   - ↑ ↓ / j k         move focus row (no wrap)
//   - 1 / 2 / 3         set focused row's action to Keep / Overwrite / Backup
//   - ␣                 cycle focused row's action (Keep → Overwrite → Backup → Keep)
//   - K / O / B         batch-resolve — apply action to every row
//   - ↵                 advance (complete stage)
//   - shift+tab / esc   back to Observe (harness pops cursor)
//
// Result: map[string]config.ConflictAction keyed by FileResult.RelPath — one
// entry per conflict file. applyCinematicConflictPicks in cmd/add.go reads
// the map and dispatches per-file mutations.
type ConflictsStage struct {
	initflow.Stage

	files   []generate.FileResult
	focus   int                              // focused row (0..len(files)-1)
	action  map[string]config.ConflictAction // per-file pick
	toneOrd []config.ConflictAction          // cycle order for ␣ keypress

	viewport initflow.Viewport // used when conflict count exceeds visible rows
}

// conflictsLabel is the kanji/kana/English triple shown in the Conflicts
// stage body title. Plan 27 shrunk the rail to four visible stages so the
// Conflicts stage renders off-rail; its rail index is StageIdxOffRail and
// the rail row is suppressed. The body still reads "衝 CONFLICT" so the
// Bonsai-metaphor identity is retained.
var conflictsLabel = initflow.StageLabel{Kanji: "衝", Kana: "しょう", English: "CONFLICT"}

// NewConflictsStage constructs the stage over the conflict entries present
// in wr. When wr has zero conflicts the ctor still returns a usable stage —
// it simply renders an empty body with a single "nothing to reconcile" line
// and completes on Enter. Callers should gate on wr.HasConflicts() before
// splicing.
func NewConflictsStage(ctx initflow.StageContext, wr *generate.WriteResult) *ConflictsStage {
	label := conflictsLabel
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

	var conflicts []generate.FileResult
	if wr != nil {
		conflicts = wr.Conflicts()
	}

	action := make(map[string]config.ConflictAction, len(conflicts))
	for _, f := range conflicts {
		action[f.RelPath] = config.ConflictActionKeep
	}

	return &ConflictsStage{
		Stage:  base,
		files:  conflicts,
		focus:  0,
		action: action,
		toneOrd: []config.ConflictAction{
			config.ConflictActionKeep,
			config.ConflictActionOverwrite,
			config.ConflictActionBackup,
		},
	}
}

// Chromeless reports true so the harness yields View() verbatim without its
// default header/footer. Embedded initflow.Stage.Chromeless() already
// returns true, but ConflictsStage declares the method explicitly so the
// contract is obvious at call-sites that type-scan for Chromeless.
func (s *ConflictsStage) Chromeless() bool { return true }

// Init implements tea.Model — no cmd on entry.
func (s *ConflictsStage) Init() tea.Cmd { return nil }

// Update handles focus movement + per-row action + batch-resolve + advance +
// back.
func (s *ConflictsStage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		s.SetSize(m.Width, m.Height)
	case tea.KeyMsg:
		if len(s.files) == 0 {
			// Empty stage — ↵/␣ completes, everything else no-ops.
			switch m.String() {
			case "enter", " ":
				s.MarkDone()
			}
			return s, nil
		}
		switch m.String() {
		case "up", "k":
			if s.focus > 0 {
				s.focus--
			}
		case "down", "j":
			if s.focus+1 < len(s.files) {
				s.focus++
			}
		case "1":
			s.setFocusedAction(config.ConflictActionKeep)
		case "2":
			s.setFocusedAction(config.ConflictActionOverwrite)
		case "3":
			s.setFocusedAction(config.ConflictActionBackup)
		case " ":
			// Cycle focused row: Keep → Overwrite → Backup → Keep.
			key := s.currentKey()
			if key == "" {
				return s, nil
			}
			cur := s.action[key]
			next := cycleAction(cur, s.toneOrd)
			s.action[key] = next
		case "K":
			s.setAllActions(config.ConflictActionKeep)
		case "O":
			s.setAllActions(config.ConflictActionOverwrite)
		case "B":
			s.setAllActions(config.ConflictActionBackup)
		case "enter":
			s.MarkDone()
			return s, nil
		}
	}
	return s, nil
}

// cycleAction returns the next action in order, wrapping around.
func cycleAction(cur config.ConflictAction, order []config.ConflictAction) config.ConflictAction {
	for i, a := range order {
		if a == cur {
			return order[(i+1)%len(order)]
		}
	}
	return order[0]
}

func (s *ConflictsStage) setFocusedAction(a config.ConflictAction) {
	key := s.currentKey()
	if key == "" {
		return
	}
	s.action[key] = a
}

func (s *ConflictsStage) setAllActions(a config.ConflictAction) {
	for _, f := range s.files {
		s.action[f.RelPath] = a
	}
}

// currentKey returns the RelPath of the focused row's file, or "" if no rows.
func (s *ConflictsStage) currentKey() string {
	if s.focus < 0 || s.focus >= len(s.files) {
		return ""
	}
	return s.files[s.focus].RelPath
}

// View returns the full AltScreen frame. Chromeless — no header/rail/footer.
// Mirrors initflow.PlantedStage.View for layout parity: centre vertically,
// compose title + divider + list + batch row + inline key hints.
func (s *ConflictsStage) View() string {
	h := s.Height()
	if h <= 0 {
		h = 24
	}
	if initflow.TerminalTooSmall(s.Width(), s.Height()) {
		return initflow.RenderMinSizeFloor(s.Width(), s.Height())
	}

	body := s.renderBody()

	// Vertically centre inside the AltScreen.
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

// renderBody composes title row + divider + list + batch-resolve row +
// inline key-hint footer. No surrounding chrome — the stage owns the frame.
func (s *ConflictsStage) renderBody() string {
	dim := initflow.DimStyle()
	bark := initflow.LabelStyle()
	white := lipgloss.NewStyle().Foreground(tui.ColorAccent).Bold(true)
	danger := lipgloss.NewStyle().Foreground(tui.ColorDanger).Bold(true)

	// Dedicated title row (replaces RenderHeader chrome). Title mixes the
	// kanji + English so ASCII fallback still reads "CONFLICT".
	var titleRow string
	if s.EnsoSafe() {
		titleRow = danger.Render(s.Label().Kanji + " · CONFLICT")
	} else {
		titleRow = danger.Render("CONFLICT")
	}

	panelW := initflow.PanelWidth(s.Width())
	divider := initflow.RenderSectionHeader("RECONCILE", panelW)

	intro := strings.Join([]string{
		titleRow,
		white.Render("Reconcile edited files."),
		dim.Render("Pick an action per file — nothing overwritten silently."),
	}, "\n")

	if len(s.files) == 0 {
		empty := dim.Render("  (nothing to reconcile)")
		hint := s.renderKeyHints()
		body := []string{intro, "", "", divider, "", empty, "", "", hint}
		return initflow.CenterBlock(strings.Join(body, "\n"), s.Width())
	}

	list := s.renderList()
	batchCaption := "  " + dim.Render("batch resolve:")
	batchRow := s.renderBatchRow()
	counter := s.renderCounter()
	hint := s.renderKeyHints()

	body := []string{
		intro,
		"",
		"",
		divider,
		"",
		list,
		"",
		batchCaption,
		batchRow,
		"",
		"  " + bark.Render(counter),
		"",
		hint,
	}
	return initflow.CenterBlock(strings.Join(body, "\n"), s.Width())
}

// renderList renders one row per conflict file in a vertical list. When the
// row count exceeds the available budget the list is rendered through a
// Viewport that follows the focus.
func (s *ConflictsStage) renderList() string {
	rows := make([]string, len(s.files))
	for i := range s.files {
		rows[i] = s.renderRow(i)
	}

	listH := s.listHeight()
	if listH > 0 && listH < len(rows) {
		s.viewport.SetLines(rows)
		s.viewport.SetHeight(listH)
		s.viewport.Follow(s.focus)
		return s.viewport.View()
	}
	return strings.Join(rows, "\n")
}

// listHeight returns the visible row budget for the conflict list. Fixed
// body rows match renderBody's body slice: intro (3) + blank (1) + blank (1) +
// divider (1) + blank (1) + blank-after-list (1) + batch caption (1) +
// batch row (1) + blank (1) + counter (1) + blank (1) + key hint (1) =
// 14 rows. Leaves at least 3 rows for the list on 20-row terminals.
func (s *ConflictsStage) listHeight() int {
	h := s.Height()
	if h <= 0 {
		return 0
	}
	const fixedRows = 14
	v := h - fixedRows
	if v < 3 {
		v = 3
	}
	return v
}

// renderRow renders a single conflict entry. Layout:
//
//	[border 2] [focus glyph 1] [sp 1] [action glyph 1] [sp 1] [relative path] [sp 2] [action label]
//
// Colour family is driven by the CURRENT action for the row — Keep green,
// Overwrite red, Backup amber — so the palette updates live as the user
// cycles ␣ / 1 / 2 / 3 / K / O / B. The focused row uses FocusBorder + bold
// emphasis; unfocused rows are dim.
func (s *ConflictsStage) renderRow(idx int) string {
	f := s.files[idx]
	focused := idx == s.focus

	act := s.action[f.RelPath]
	tone := toneFor(act)
	rowStyle := initflow.ConflictRowStyle(tone)
	glyph := initflow.ConflictActionGlyph(tone)

	border := initflow.UnfocusBorder()
	if focused {
		border = initflow.FocusBorder()
	}

	focusGlyph := " "
	leaf := lipgloss.NewStyle().Foreground(tui.ColorPrimary).Bold(true)
	if focused {
		focusGlyph = leaf.Render("▸")
	}

	// Path column — budget = panelW - (border 2 + focusGlyph 1 + sp 1 +
	// actionGlyph 1 + sp 1 + actionLabel ~18 + sp 2) ≈ panelW - 26.
	panelW := initflow.PanelWidth(s.Width())
	labelText := actionLabel(act)
	labelW := 18 // generous budget so "backup + overwrite" fits
	pathBudget := panelW - 2 - 1 - 1 - 1 - 1 - labelW - 2
	if pathBudget < 10 {
		pathBudget = 10
	}
	path := f.RelPath
	if lipgloss.Width(path) > pathBudget {
		rr := []rune(path)
		if len(rr) > pathBudget-1 {
			path = "…" + string(rr[len(rr)-pathBudget+1:])
		}
	}

	pathStyle := initflow.UnfocusedNameStyle()
	if focused {
		pathStyle = initflow.FocusedNameStyle()
	}

	actionCell := rowStyle.Render(glyph) + " " + rowStyle.Render(labelText)
	return border + focusGlyph + " " + pathStyle.Render(initflow.PadRight(path, pathBudget)) + "  " + actionCell
}

// renderBatchRow draws the three labelled cells below the list. Uppercase
// K/O/B trigger each cell; lowercase k/o/b are no-ops (Plan 27 §C4).
func (s *ConflictsStage) renderBatchRow() string {
	keepStyle := initflow.ConflictRowStyle(initflow.ConflictToneKeep)
	overStyle := initflow.ConflictRowStyle(initflow.ConflictToneOverwrite)
	backStyle := initflow.ConflictRowStyle(initflow.ConflictToneBackup)
	bark := initflow.LabelStyle()

	cell := func(key string, style lipgloss.Style, label string) string {
		return style.Render("[ ") + bark.Render(key) + style.Render(" "+label+" ]")
	}

	row := cell("K", keepStyle, "Keep all") + "  " +
		cell("O", overStyle, "Overwrite all") + "  " +
		cell("B", backStyle, "Backup all")
	return "  " + row
}

// renderCounter emits a sticky summary: "file N of M · focus: <action>".
func (s *ConflictsStage) renderCounter() string {
	if len(s.files) == 0 {
		return ""
	}
	key := s.currentKey()
	act := s.action[key]
	return fmt.Sprintf("file %d of %d · focus: %s", s.focus+1, len(s.files), actionLabel(act))
}

// renderKeyHints renders the inline key hint row (since the stage is
// chromeless there is no RenderFooter). Keeps the hint set compact.
func (s *ConflictsStage) renderKeyHints() string {
	dim := initflow.DimStyle()
	hint := "↑↓ focus  ·  1/2/3 action  ·  ␣ cycle  ·  K/O/B batch  ·  ↵ next  ·  esc back"
	return dim.Render(hint)
}

// toneFor maps a ConflictAction to its visual tone.
func toneFor(a config.ConflictAction) initflow.ConflictActionTone {
	switch a {
	case config.ConflictActionOverwrite:
		return initflow.ConflictToneOverwrite
	case config.ConflictActionBackup:
		return initflow.ConflictToneBackup
	default:
		return initflow.ConflictToneKeep
	}
}

// actionLabel renders a ConflictAction as its user-facing label.
func actionLabel(a config.ConflictAction) string {
	switch a {
	case config.ConflictActionKeep:
		return "keep local"
	case config.ConflictActionOverwrite:
		return "overwrite"
	case config.ConflictActionBackup:
		return "backup + overwrite"
	default:
		return "?"
	}
}

// Result returns the per-file ConflictAction map. Keys are RelPath values;
// every conflict entry is guaranteed to appear in the map (default Keep if
// untouched).
func (s *ConflictsStage) Result() any {
	if len(s.action) == 0 {
		return map[string]config.ConflictAction{}
	}
	out := make(map[string]config.ConflictAction, len(s.action))
	for k, v := range s.action {
		out[k] = v
	}
	return out
}

// Reset clears the completion flag but preserves per-file picks and focus so
// Esc-back → re-entry reads exactly where the user left off.
func (s *ConflictsStage) Reset() tea.Cmd {
	s.ClearDone()
	return nil
}
