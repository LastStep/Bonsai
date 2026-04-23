package guideflow

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/LastStep/Bonsai/internal/tui"
	"github.com/LastStep/Bonsai/internal/tui/initflow"
)

// ViewerStage is the BubbleTea model for `bonsai guide`. Four
// topic tabs across the top, a scrollable glamour-rendered body
// below, standard initflow chrome wrapping the whole frame. Embeds
// initflow.Stage with railHidden=true — guide is read-only (no
// mutation process, no rail segments to walk).
//
// Markdown output is cached per (topicIdx, viewport width) so
// resize events only re-render when the width actually changes and
// tab cycles re-use prior renders.
type ViewerStage struct {
	initflow.Stage

	topics   []Topic
	idx      int
	viewport viewport.Model
	rendered map[string]string // key: "idx:width" → glamour output
	width    int
	height   int
	quit     bool
}

// narrowStripThreshold defines the cell budget above which the full
// tab labels (QUICKSTART · CONCEPTS · CLI · CUSTOM) are rendered.
// Below the threshold the short labels (START · CONCP · CLI · CUSTM)
// kick in so the strip still fits inside the 70-col min-size floor.
// The threshold is measured dynamically at render time — this
// constant only supplies the hard minimum below which short labels
// are always used even if they would technically fit.
const narrowStripThreshold = 60

// NewViewer constructs a ViewerStage from the given topics + the
// initial topic key (empty or unknown → idx 0). version and
// projectDir feed the header chrome; projectDir may be empty if the
// caller couldn't resolve it (guide doesn't strictly need it).
func NewViewer(topics []Topic, initialKey, version, projectDir string) *ViewerStage {
	idx := 0
	if initialKey != "" {
		for i, t := range topics {
			if t.Key == initialKey {
				idx = i
				break
			}
		}
	}

	base := initflow.NewStage(
		0,
		initflow.StageLabel{English: "GUIDE"},
		"GUIDE",
		version,
		projectDir,
		"",
		"",
		time.Time{},
	)
	base.SetRailHidden(true)
	base.SetHeaderAction("GUIDE")
	// Guide is scope-global (docs are the same wherever you run it),
	// so omit the destination preamble on the right block — project
	// path still renders on row 2 via the base chrome.
	base.SetHeaderRightLabel("")

	vp := viewport.New(0, 0)

	return &ViewerStage{
		Stage:    base,
		topics:   topics,
		idx:      idx,
		viewport: vp,
		rendered: make(map[string]string),
	}
}

// Init implements tea.Model. No-op — the first render is wired up
// on the first WindowSizeMsg.
func (s *ViewerStage) Init() tea.Cmd { return nil }

// Update implements tea.Model.
//
// Keys:
//   - tab / right / l       — next topic (wraps)
//   - shift+tab / left / h  — prev topic (wraps)
//   - g / home              — viewport top
//   - G / end               — viewport bottom
//   - up/down, j/k, pgup/pgdn, space, u/d — scroll (viewport)
//   - q / esc / ctrl+c      — quit
func (s *ViewerStage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = m.Width
		s.height = m.Height
		s.SetSize(m.Width, m.Height)
		s.resizeViewport()
		s.refreshViewportContent()
		return s, nil
	case tea.KeyMsg:
		switch m.String() {
		case "q", "esc", "ctrl+c":
			s.quit = true
			s.MarkDone()
			return s, tea.Quit
		case "tab", "right", "l":
			if len(s.topics) == 0 {
				return s, nil
			}
			s.idx = (s.idx + 1) % len(s.topics)
			s.refreshViewportContent()
			s.viewport.GotoTop()
			return s, nil
		case "shift+tab", "left", "h":
			if len(s.topics) == 0 {
				return s, nil
			}
			s.idx = (s.idx - 1 + len(s.topics)) % len(s.topics)
			s.refreshViewportContent()
			s.viewport.GotoTop()
			return s, nil
		case "g", "home":
			s.viewport.GotoTop()
			return s, nil
		case "G", "end":
			s.viewport.GotoBottom()
			return s, nil
		}
		// Fall through to viewport for scroll keys (up/down, j/k,
		// pgup/pgdn, space, u/d).
		var cmd tea.Cmd
		s.viewport, cmd = s.viewport.Update(msg)
		return s, cmd
	}
	return s, nil
}

// Title implements harness.Step.
func (s *ViewerStage) Title() string { return "GUIDE" }

// Result implements harness.Step. Guide is read-only — no payload.
func (s *ViewerStage) Result() any { return nil }

// View composes the full frame: header + tab strip + viewport +
// footer. Falls back to the min-size floor when the terminal is
// below the 70×20 threshold.
func (s *ViewerStage) View() string {
	if s.quit {
		return ""
	}
	if initflow.TerminalTooSmall(s.width, s.height) {
		return initflow.RenderMinSizeFloor(s.width, s.height)
	}

	width := s.width
	if width <= 0 {
		width = 80
	}

	tabRow := s.renderTabs(width)
	body := tabRow + "\n\n" + s.viewport.View()
	return s.RenderFrame(body, s.keyHints())
}

// keyHints returns the footer key row.
func (s *ViewerStage) keyHints() []initflow.KeyHint {
	return []initflow.KeyHint{
		{Key: "←→", Desc: "tabs"},
		{Key: "↑↓", Desc: "scroll"},
		{Key: "g/G", Desc: "top/bot"},
		{Key: "q", Desc: "quit"},
	}
}

// renderTabs renders the single-row topic tab strip. Active tab is
// bold ColorPrimary; inactive tabs are ColorMuted. Chooses the full
// label set when the rendered strip fits inside
// initflow.ClampColumns(width).Total; otherwise falls back to the
// compact short labels so the 4-tab strip fits inside the 70-col
// min-size floor.
func (s *ViewerStage) renderTabs(width int) string {
	full := s.buildTabStrip(width, false)
	// Measure full-label strip against the available content budget.
	// ClampColumns returns (nameW, descW, tagW); sum is the effective
	// panel budget in cells. The +4 adds back the init-flow sidePad
	// margin so the threshold compares like-for-like against the
	// rendered strip (which itself has no margin baked in).
	nameW, descW, tagW := initflow.ClampColumns(width - 4)
	budget := nameW + descW + tagW
	if budget < narrowStripThreshold {
		budget = narrowStripThreshold
	}
	if lipgloss.Width(full) > budget {
		return s.buildTabStrip(width, true)
	}
	return full
}

// buildTabStrip assembles the tab row from s.topics. When useShort
// is true each cell uses the Topic.Short label.
func (s *ViewerStage) buildTabStrip(_ int, useShort bool) string {
	active := lipgloss.NewStyle().Foreground(tui.ColorPrimary).Bold(true)
	muted := lipgloss.NewStyle().Foreground(tui.ColorMuted)

	cells := make([]string, 0, len(s.topics))
	for i, t := range s.topics {
		label := t.Label
		if useShort {
			label = t.Short
		}
		if i == s.idx {
			cells = append(cells, active.Render(label))
		} else {
			cells = append(cells, muted.Render(label))
		}
	}
	sep := "  "
	return strings.Join(cells, sep)
}

// resizeViewport recomputes the viewport dims from the current
// terminal size. Body area = height minus chrome (header 2 + 2
// blanks + footer 2) minus tab strip (1 + 1 blank). Floor at 3 so
// at least a small scroll window is visible on short terminals.
func (s *ViewerStage) resizeViewport() {
	w := initflow.PanelWidth(s.width)
	if w <= 0 {
		w = 80
	}
	// Chrome: header (2 rows) + 2 blank + footer (2 rows, rule +
	// brand row) = 6. Tab strip row (1) + 1 blank = 2. Total 8.
	const chromeRows = 8
	h := s.height - chromeRows
	if h < 3 {
		h = 3
	}
	s.viewport.Width = w
	s.viewport.Height = h
}

// refreshViewportContent resolves the rendered markdown for the
// current topic at the current width, pushing it into the
// viewport. Uses the cache keyed by "idx:width" so repeated
// resize-to-same-width and tab-revisits skip the glamour call.
func (s *ViewerStage) refreshViewportContent() {
	if len(s.topics) == 0 {
		s.viewport.SetContent("")
		return
	}
	w := s.viewport.Width
	if w <= 0 {
		w = defaultRenderWidth
	}
	key := fmt.Sprintf("%d:%d", s.idx, w)
	if cached, ok := s.rendered[key]; ok {
		s.viewport.SetContent(cached)
		return
	}
	rendered, err := renderMarkdown(s.topics[s.idx].Markdown, w)
	if err != nil {
		// Surface the render failure inside the viewport itself
		// so the user sees something rather than a blank pane;
		// quit key still exits cleanly.
		s.viewport.SetContent("Render error: " + err.Error())
		return
	}
	s.rendered[key] = rendered
	s.viewport.SetContent(rendered)
}
