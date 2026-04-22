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

// ConflictsStage is the tabbed conflict-resolution picker at rail position 5
// (衝 CONFLICT). Rendered only when the Grow action produced at least one
// ActionConflict file — the harness gates construction on
// wr.HasConflicts(), so this stage never instantiates for clean writes.
//
// Layout: one tab per conflict file. Each tab shows:
//   - Header  : the conflict file's relative path
//   - Summary : a short "local has edits · source has changes" line (until the
//     generator surfaces an inline diff field, this is a placeholder — see
//     the TODO next to diffSummary below)
//   - Radio   : three rows — Keep local · Overwrite with source · Backup
//     then overwrite. Default Keep.
//
// Keystrokes match the addflow house style:
//   - ← → / h l     cycle tabs (wrap)
//   - ↑ ↓ / j k     cycle radio (no-wrap)
//   - ␣ / ↵         advance (to next tab, or complete if last)
//   - shift+tab / esc   back to prior stage (Grow; harness handles the pop)
//
// Result: map[string]config.ConflictAction keyed by FileResult.RelPath — one
// entry per conflict file. applyCinematicConflictPicks in cmd/add.go
// reads the map and dispatches per-file mutations.
type ConflictsStage struct {
	initflow.Stage

	files   []generate.FileResult
	catIdx  int                              // focused tab
	action  map[string]config.ConflictAction // per-file pick
	radio   map[string]int                   // per-file radio focus row
	toneOrd []config.ConflictAction          // index → action for radio rows
}

// NewConflictsStage constructs the stage over the conflict entries present
// in wr. When wr has zero conflicts the ctor still returns a usable stage —
// it simply renders an empty body with a single "nothing to reconcile" line
// and completes on Enter. Callers should gate on wr.HasConflicts() before
// splicing.
// conflictsLabel is the kanji/kana/English triple shown in the Conflicts
// stage body title. Plan 27 shrunk the rail to four visible stages so the
// Conflicts stage renders off-rail; its rail index is StageIdxOffRail and
// the rail row is suppressed. The body still reads "衝 CONFLICT" so the
// Bonsai-metaphor identity is retained.
var conflictsLabel = initflow.StageLabel{Kanji: "衝", Kana: "しょう", English: "CONFLICT"}

// Rail suppression is implicit: passing StageIdxOffRail (-1) into the base
// ctor trips the negative-index branch in internal/tui/initflow/stage.go
// (railHidden := s.railHidden || s.idx < 0), so no SetRailIndex call is
// needed here.
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
	base.SetRailLabels(StageLabels)

	var conflicts []generate.FileResult
	if wr != nil {
		conflicts = wr.Conflicts()
	}

	action := make(map[string]config.ConflictAction, len(conflicts))
	radio := make(map[string]int, len(conflicts))
	for _, f := range conflicts {
		action[f.RelPath] = config.ConflictActionKeep
		radio[f.RelPath] = 0
	}

	return &ConflictsStage{
		Stage:  base,
		files:  conflicts,
		catIdx: 0,
		action: action,
		radio:  radio,
		toneOrd: []config.ConflictAction{
			config.ConflictActionKeep,
			config.ConflictActionOverwrite,
			config.ConflictActionBackup,
		},
	}
}

// Init implements tea.Model — no cmd on entry.
func (s *ConflictsStage) Init() tea.Cmd { return nil }

// Update handles tab cycle + radio cycle + advance + back.
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
		case "left", "h":
			s.catIdx = (s.catIdx - 1 + len(s.files)) % len(s.files)
		case "right", "l":
			s.catIdx = (s.catIdx + 1) % len(s.files)
		case "up", "k":
			key := s.currentKey()
			if key == "" {
				return s, nil
			}
			cur := s.radio[key] - 1
			if cur < 0 {
				cur = 0
			}
			s.radio[key] = cur
			s.action[key] = s.toneOrd[cur]
		case "down", "j":
			key := s.currentKey()
			if key == "" {
				return s, nil
			}
			cur := s.radio[key] + 1
			if cur >= len(s.toneOrd) {
				cur = len(s.toneOrd) - 1
			}
			s.radio[key] = cur
			s.action[key] = s.toneOrd[cur]
		case " ", "enter":
			// Advance to the next tab; complete on the last.
			if s.catIdx+1 < len(s.files) {
				s.catIdx++
				return s, nil
			}
			s.MarkDone()
			return s, nil
		}
	}
	return s, nil
}

// currentKey returns the RelPath of the focused tab's file, or "" if no tabs.
func (s *ConflictsStage) currentKey() string {
	if s.catIdx < 0 || s.catIdx >= len(s.files) {
		return ""
	}
	return s.files[s.catIdx].RelPath
}

// View composes the body inside the shared frame.
func (s *ConflictsStage) View() string {
	return s.RenderFrame(s.renderBody(), s.keyHints())
}

func (s *ConflictsStage) keyHints() []initflow.KeyHint {
	return []initflow.KeyHint{
		{Key: "←→", Desc: "file"},
		{Key: "↑↓", Desc: "action"},
		{Key: "↵", Desc: "next"},
		{Key: "esc", Desc: "back"},
		{Key: "ctrl-c", Desc: "quit"},
	}
}

func (s *ConflictsStage) renderBody() string {
	dim := initflow.DimStyle()
	bark := initflow.LabelStyle()
	white := lipgloss.NewStyle().Foreground(tui.ColorAccent).Bold(true)

	var title string
	if s.EnsoSafe() {
		title = bark.Render(s.Label().Kanji) + " " + white.Render(s.Label().English)
	} else {
		title = white.Render(s.Label().English)
	}

	intro := strings.Join([]string{
		title,
		white.Render("Reconcile edited files."),
		dim.Render("Pick an action per file — nothing overwritten silently."),
	}, "\n")

	panelW := initflow.PanelWidth(s.Width())
	divider := initflow.RenderSectionHeader("CONFLICTS", panelW)

	if len(s.files) == 0 {
		empty := dim.Render("  (nothing to reconcile)")
		body := []string{intro, "", "", divider, "", empty}
		return initflow.CenterBlock(strings.Join(body, "\n"), s.Width())
	}

	tabRow := s.renderTabs()
	header := s.renderFileHeader()
	summary := s.renderDiffSummary()
	radio := s.renderRadio()
	counter := s.renderCounter()

	body := []string{
		intro,
		"",
		"",
		divider,
		"",
		tabRow,
		"",
		header,
		"",
		summary,
		"",
		radio,
		"",
		dim.Render(counter),
	}
	return initflow.CenterBlock(strings.Join(body, "\n"), s.Width())
}

// renderTabs draws a compact tab strip — one cell per conflict file, keyed
// by file index. File paths are truncated to fit.
func (s *ConflictsStage) renderTabs() string {
	muted := lipgloss.NewStyle().Foreground(tui.ColorMuted)
	leaf := lipgloss.NewStyle().Foreground(tui.ColorPrimary).Bold(true)
	bracket := lipgloss.NewStyle().Foreground(tui.ColorPrimary)
	chevron := lipgloss.NewStyle().Foreground(tui.ColorSecondary).Bold(true)

	cells := make([]string, 0, len(s.files))
	// Use a compact label — index + short basename — so the tab strip does
	// not wrap on terminals with many conflicts. Full path is shown in the
	// header block below.
	for i, f := range s.files {
		label := fmt.Sprintf("%d · %s", i+1, basename(f.RelPath))
		// Clamp label to a conservative width so many conflicts still fit.
		if lipgloss.Width(label) > 20 {
			rr := []rune(label)
			if len(rr) > 19 {
				label = string(rr[:19]) + "…"
			}
		}
		var cell string
		if i == s.catIdx {
			cell = bracket.Render("[ ") + leaf.Render(label) + bracket.Render(" ]")
		} else {
			cell = "  " + muted.Render(label) + "  "
		}
		cells = append(cells, cell)
	}

	row := strings.Join(cells, " ")
	return chevron.Render("‹") + " " + row + " " + chevron.Render("›")
}

// renderFileHeader shows the focused file's full relative path, truncated to
// the panel width if needed.
func (s *ConflictsStage) renderFileHeader() string {
	bark := initflow.LabelStyle()
	value := initflow.ValueStyle()

	key := s.currentKey()
	if key == "" {
		return ""
	}
	panelW := initflow.PanelWidth(s.Width())
	// Reserve space for the "FILE " label prefix (4 + 2 gap).
	labelW := 6
	pathBudget := panelW - labelW - 2
	if pathBudget < 10 {
		pathBudget = 10
	}
	path := key
	if lipgloss.Width(path) > pathBudget {
		rr := []rune(path)
		if len(rr) > pathBudget-1 {
			// Head-truncate paths so the distinctive basename survives.
			path = "…" + string(rr[len(rr)-pathBudget+1:])
		}
	}
	return "  " + bark.Render(initflow.PadRight("FILE", labelW)) + value.Render(path)
}

// renderDiffSummary emits a short status line describing the conflict. TODO:
// surface a real inline diff once FileResult carries one. Until then we read
// Source (the catalog path the file came from) and the RelPath to compose a
// one-line provenance summary — no partial diff leaks, no placeholder data
// masquerades as content.
func (s *ConflictsStage) renderDiffSummary() string {
	dim := initflow.DimStyle()
	bark := initflow.LabelStyle()
	value := initflow.ValueStyle()

	if s.catIdx < 0 || s.catIdx >= len(s.files) {
		return ""
	}
	f := s.files[s.catIdx]

	labelW := 6
	source := f.Source
	if source == "" {
		source = "—"
	}

	panelW := initflow.PanelWidth(s.Width())
	srcBudget := panelW - labelW - 2
	if srcBudget < 10 {
		srcBudget = 10
	}
	if lipgloss.Width(source) > srcBudget {
		rr := []rune(source)
		if len(rr) > srcBudget-1 {
			source = string(rr[:srcBudget-1]) + "…"
		}
	}

	// TODO(plan-23-phase2): replace the placeholder status with a real inline
	// diff summary once generate.FileResult exposes one. Today the write
	// pipeline only records Action=ActionConflict, not a per-file diff — so
	// we surface provenance (source) + a stable status instead of mocking
	// content deltas.
	statusLine := dim.Render("local has edits · source has changes")
	srcLine := "  " + bark.Render(initflow.PadRight("SOURCE", labelW)) + value.Render(source)
	return "  " + bark.Render(initflow.PadRight("WHAT", labelW)) + statusLine + "\n" + srcLine
}

// renderRadio draws the three Keep / Overwrite / Backup rows for the focused
// tab, using the colour-coded ConflictRowStyle tokens from initflow/design.
func (s *ConflictsStage) renderRadio() string {
	key := s.currentKey()
	if key == "" {
		return ""
	}

	rowFocus := s.radio[key]
	rows := []struct {
		tone  initflow.ConflictActionTone
		label string
		desc  string
	}{
		{initflow.ConflictToneKeep, "Keep local", "skip this file — your edits win"},
		{initflow.ConflictToneOverwrite, "Overwrite with source", "discard edits and pull new content"},
		{initflow.ConflictToneBackup, "Backup then overwrite", "save .bak and pull new content"},
	}

	out := make([]string, 0, len(rows))
	for i, r := range rows {
		style := initflow.ConflictRowStyle(r.tone)
		glyph := initflow.ConflictActionGlyph(r.tone)

		focused := i == rowFocus
		border := initflow.UnfocusBorder()
		if focused {
			border = initflow.FocusBorder()
		}

		nameStyle := style
		if focused {
			nameStyle = nameStyle.Bold(true)
		}

		desc := initflow.UnfocusedDescStyle().Render(r.desc)
		if focused {
			desc = initflow.FocusedDescStyle().Render(r.desc)
		}
		line := border + style.Render(glyph) + " " + nameStyle.Render(r.label) + "  " + desc
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}

// renderCounter emits "file N of M — action: <label>" so the user always has
// a sticky summary of where they are in the picker and what they picked.
func (s *ConflictsStage) renderCounter() string {
	if len(s.files) == 0 {
		return ""
	}
	key := s.currentKey()
	act := s.action[key]
	return fmt.Sprintf("file %d of %d · action: %s", s.catIdx+1, len(s.files), actionLabel(act))
}

// actionLabel renders a ConflictAction as its user-facing label. Centralised
// here so the counter + (future) confirm lines read from a single source.
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

// basename returns the last path segment of a forward-slash path. Kept local
// so conflicts.go has no dependency on path/filepath on the hot render path.
func basename(p string) string {
	if p == "" {
		return ""
	}
	idx := strings.LastIndex(p, "/")
	if idx < 0 {
		return p
	}
	return p[idx+1:]
}

// Result returns the per-file ConflictAction map. Keys are RelPath values;
// every conflict entry is guaranteed to appear in the map (default Keep if
// untouched). When the stage was aborted via esc the harness calls Reset and
// discards the result — callers receiving a nil map on the read path should
// fall through to a no-op.
func (s *ConflictsStage) Result() any {
	// Return a copy so later mutations on the stage (shouldn't happen, but
	// defensive) cannot mutate the harness-held snapshot.
	if len(s.action) == 0 {
		return map[string]config.ConflictAction{}
	}
	out := make(map[string]config.ConflictAction, len(s.action))
	for k, v := range s.action {
		out[k] = v
	}
	return out
}

// Reset clears the completion flag but preserves per-file picks, tab focus,
// and radio focus so Esc-back → re-entry reads exactly where the user left
// off. Mirrors the Graft stage's preserve-everything-but-done policy.
func (s *ConflictsStage) Reset() tea.Cmd {
	s.ClearDone()
	return nil
}
