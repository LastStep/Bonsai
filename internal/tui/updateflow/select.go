package updateflow

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/LastStep/Bonsai/internal/generate"
	"github.com/LastStep/Bonsai/internal/tui"
	"github.com/LastStep/Bonsai/internal/tui/initflow"
)

// selectLabel is the kanji triple rendered in the body title.
var selectLabel = initflow.StageLabel{Kanji: "択", Kana: "えらぶ", English: "SELECT"}

// SelectStage is a chromeless per-agent multi-select. Each agent with
// valid discoveries becomes a tab; within a tab each row is a file that
// can be toggled on/off. All rows default to selected (mirrors legacy
// huh.Selected(true) default from the pre-cinematic flow).
//
// Keystrokes:
//
//	←→/h/l      cycle agent tabs
//	↑↓/j/k      move focus within the current tab
//	␣           toggle the focused file
//	a           toggle all files in current tab
//	↵           advance
//
// Result: map[string][]string — keyed by AgentName, value is the list of
// selected "type:name" keys (matches legacy buildCustomFileOptions Value
// shape so applyCustomFileSelection can consume the output unchanged).
type SelectStage struct {
	initflow.Stage

	agents []AgentDiscoveries

	// Tab state — which agent is currently being shown + selection bitset
	// per agent. selected[i] has one bool per agents[i].Valid entry.
	tab      int
	focus    int
	selected [][]bool

	viewport initflow.Viewport
}

// NewSelectStage constructs the chromeless per-agent multi-select.
// agents should contain ONLY entries with at least one valid discovery —
// the caller filters out invalid-only agents upstream.
func NewSelectStage(ctx initflow.StageContext, agents []AgentDiscoveries) *SelectStage {
	label := selectLabel
	base := initflow.NewStage(
		StageIdxSelect,
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

	selected := make([][]bool, len(agents))
	for i, a := range agents {
		row := make([]bool, len(a.Valid))
		for j := range row {
			row[j] = true // default all-selected — legacy contract
		}
		selected[i] = row
	}

	return &SelectStage{
		Stage:    base,
		agents:   agents,
		selected: selected,
	}
}

// Chromeless reports true so the harness yields the stage's View verbatim.
// Plan 31 §F §4 — chromeless per-agent tab strip.
func (s *SelectStage) Chromeless() bool { return true }

// Init implements tea.Model — no cmd on entry.
func (s *SelectStage) Init() tea.Cmd { return nil }

// Update handles tab cycling + in-tab focus + space toggle + enter advance.
func (s *SelectStage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		s.SetSize(m.Width, m.Height)
	case tea.KeyMsg:
		if len(s.agents) == 0 {
			if m.String() == "enter" {
				s.MarkDone()
			}
			return s, nil
		}
		switch m.String() {
		case "left", "h":
			if s.tab > 0 {
				s.tab--
				s.focus = 0
			}
		case "right", "l":
			if s.tab+1 < len(s.agents) {
				s.tab++
				s.focus = 0
			}
		case "up", "k":
			if s.focus > 0 {
				s.focus--
			}
		case "down", "j":
			if s.focus+1 < len(s.agents[s.tab].Valid) {
				s.focus++
			}
		case " ":
			if len(s.selected[s.tab]) > 0 && s.focus < len(s.selected[s.tab]) {
				s.selected[s.tab][s.focus] = !s.selected[s.tab][s.focus]
			}
		case "a":
			// Toggle all in current tab — if any unchecked, select all;
			// otherwise deselect all.
			any := false
			for _, v := range s.selected[s.tab] {
				if !v {
					any = true
					break
				}
			}
			for i := range s.selected[s.tab] {
				s.selected[s.tab][i] = any
			}
		case "enter":
			s.MarkDone()
			return s, nil
		}
	}
	return s, nil
}

// View renders the chromeless body — vertically-centred inside AltScreen,
// same pattern as addflow.ConflictsStage.
func (s *SelectStage) View() string {
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

func (s *SelectStage) renderBody() string {
	white := lipgloss.NewStyle().Foreground(tui.ColorAccent).Bold(true)
	bark := initflow.LabelStyle()
	dim := initflow.DimStyle()

	var title string
	if s.EnsoSafe() {
		title = bark.Render(selectLabel.Kanji) + " " + white.Render(selectLabel.English)
	} else {
		title = white.Render(selectLabel.English)
	}

	intro := strings.Join([]string{
		title,
		white.Render("Promote discovered files into the lockfile."),
		dim.Render("Tab between agents · toggle per-file · ↵ to sync."),
	}, "\n")

	tabs := s.renderTabs()
	divider := initflow.RenderSectionHeader("FILES", initflow.PanelWidth(s.Width()))
	list := s.renderList()
	counter := s.renderCounter()
	hint := s.renderKeyHints()

	body := []string{
		intro,
		"",
		"",
		tabs,
		"",
		divider,
		"",
		list,
		"",
		"  " + bark.Render(counter),
		"",
		hint,
	}
	return initflow.CenterBlock(strings.Join(body, "\n"), s.Width())
}

// renderTabs draws a horizontal tab strip — active tab bolded / bracketed,
// inactive tabs dim. Counts render after each label as "(n/m)".
func (s *SelectStage) renderTabs() string {
	if len(s.agents) == 0 {
		return ""
	}
	leaf := lipgloss.NewStyle().Foreground(tui.ColorPrimary).Bold(true)
	dim := initflow.DimStyle()
	white := lipgloss.NewStyle().Foreground(tui.ColorAccent).Bold(true)

	cells := make([]string, 0, len(s.agents))
	for i, a := range s.agents {
		picked := 0
		for _, v := range s.selected[i] {
			if v {
				picked++
			}
		}
		label := fmt.Sprintf("%s (%d/%d)", a.AgentLabel, picked, len(a.Valid))
		if i == s.tab {
			cells = append(cells, leaf.Render("▸")+" "+white.Render(label))
		} else {
			cells = append(cells, "  "+dim.Render(label))
		}
	}
	return "  " + strings.Join(cells, "    ")
}

// renderList renders one row per valid file in the current tab.
func (s *SelectStage) renderList() string {
	if len(s.agents) == 0 {
		dim := initflow.DimStyle()
		return "  " + dim.Render("(nothing to promote)")
	}
	rows := make([]string, len(s.agents[s.tab].Valid))
	for i := range s.agents[s.tab].Valid {
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

func (s *SelectStage) listHeight() int {
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

// renderRow renders a single file row. Selected = ◆ leaf, unselected = ◇ dim.
func (s *SelectStage) renderRow(idx int) string {
	df := s.agents[s.tab].Valid[idx]
	focused := idx == s.focus
	sel := s.selected[s.tab][idx]

	leaf := lipgloss.NewStyle().Foreground(tui.ColorPrimary).Bold(true)
	dim := initflow.DimStyle()

	border := initflow.UnfocusBorder()
	if focused {
		border = initflow.FocusBorder()
	}

	glyph := "◇"
	glyphStyle := dim
	if sel {
		glyph = "◆"
		glyphStyle = leaf
	}

	nameStyle := initflow.UnfocusedNameStyle()
	descStyle := initflow.UnfocusedDescStyle()
	if focused {
		nameStyle = initflow.FocusedNameStyle()
		descStyle = initflow.FocusedDescStyle()
	}

	name := fmt.Sprintf("[%s] %s", df.Type, resolveLabel(df))
	desc := ""
	if df.Meta != nil {
		desc = df.Meta.Description
	}

	panelW := initflow.PanelWidth(s.Width())
	const nameColW = 32
	const gap = 2
	if lipgloss.Width(name) > nameColW {
		rr := []rune(name)
		if len(rr) > nameColW-1 {
			name = string(rr[:nameColW-1]) + "…"
		}
	}
	namePadded := initflow.PadRight(nameStyle.Render(name), nameColW)

	descBudget := panelW - 2 - 1 - 1 - nameColW - gap
	if descBudget < 10 {
		descBudget = 10
	}
	if lipgloss.Width(desc) > descBudget {
		rr := []rune(desc)
		if len(rr) > descBudget-1 {
			desc = string(rr[:descBudget-1]) + "…"
		}
	}
	return border + glyphStyle.Render(glyph) + " " + namePadded + strings.Repeat(" ", gap) + descStyle.Render(desc)
}

func (s *SelectStage) renderCounter() string {
	if len(s.agents) == 0 {
		return ""
	}
	cur := s.agents[s.tab]
	picked := 0
	for _, v := range s.selected[s.tab] {
		if v {
			picked++
		}
	}
	return fmt.Sprintf("%s · %d of %d selected", cur.AgentLabel, picked, len(cur.Valid))
}

func (s *SelectStage) renderKeyHints() string {
	dim := initflow.DimStyle()
	hint := "←→ tabs  ·  ↑↓ focus  ·  ␣ toggle  ·  a toggle-all  ·  ↵ sync"
	return dim.Render(hint)
}

// SelectedKeys returns the per-agent list of selected "type:name" keys.
// Caller-side applyCustomFileSelection consumes this directly.
func (s *SelectStage) SelectedKeys() map[string][]string {
	out := make(map[string][]string, len(s.agents))
	for i, a := range s.agents {
		keys := make([]string, 0, len(a.Valid))
		for j, df := range a.Valid {
			if s.selected[i][j] {
				keys = append(keys, df.Type+":"+df.Name)
			}
		}
		out[a.AgentName] = keys
	}
	return out
}

// Result returns the SelectedKeys map so the harness can expose the
// picks to downstream stages / post-harness callers via prev[].
func (s *SelectStage) Result() any { return s.SelectedKeys() }

// Reset preserves focus + selection across Esc-back / re-entry.
func (s *SelectStage) Reset() tea.Cmd {
	s.ClearDone()
	return nil
}

// resolveLabel returns df.Meta.DisplayName if populated, otherwise df.Name.
func resolveLabel(df generate.DiscoveredFile) string {
	if df.Meta != nil && df.Meta.DisplayName != "" {
		return df.Meta.DisplayName
	}
	return df.Name
}
