package guideflow

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
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
	// renderers caches glamour.TermRenderer instances keyed by viewport
	// width. Building a renderer is expensive (termenv OSC 11 background
	// query + goldmark + chroma init — typically 100ms+), so we build
	// once per distinct width and reuse for every tab switch / scroll
	// at that width. Cache writes happen only on the tea.Update loop.
	renderers map[int]*glamour.TermRenderer
	// preWarmedWidth is the width for which preWarmCmd was already
	// dispatched. Protects against re-dispatching on no-op
	// WindowSizeMsg (same width, different height).
	preWarmedWidth int
	width          int
	height         int
	quit           bool
}

// Tab-label fallback is measurement-driven: renderTabs compares the
// rendered full-label strip width against the live panel budget
// (initflow.PanelWidth) and switches to the short labels
// (START · CONCP · CLI · CUSTM) whenever the full strip would
// overflow. No hard threshold — the viewport-relative budget makes
// the decision at every WindowSizeMsg.

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

	// Empty label + empty title — the rail is hidden (SetRailHidden
	// below) and the chromeless viewer has no breadcrumb title slot,
	// so both fields are unused. Header text comes exclusively from
	// the SetHeaderAction("GUIDE") call a few lines down.
	base := initflow.NewStage(
		0,
		initflow.StageLabel{},
		"",
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
		Stage:     base,
		topics:    topics,
		idx:       idx,
		viewport:  vp,
		rendered:  make(map[string]string),
		renderers: make(map[int]*glamour.TermRenderer),
	}
}

// rendererFor returns the cached glamour.TermRenderer for the given
// width, building + caching a fresh one if absent. Width ≤ 0 clamps
// to defaultRenderWidth. Called on the tea.Update loop only — writes
// to s.renderers are not goroutine-safe.
func (s *ViewerStage) rendererFor(width int) (*glamour.TermRenderer, error) {
	if width <= 0 {
		width = defaultRenderWidth
	}
	if r, ok := s.renderers[width]; ok {
		return r, nil
	}
	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return nil, fmt.Errorf("guideflow: create renderer: %w", err)
	}
	s.renderers[width] = r
	return r, nil
}

// preWarmMsg is the result of a background pre-warm pass. results
// maps topic idx → rendered markdown string at the given width.
// Handled in Update where the values are committed to s.rendered on
// the tea loop (no shared-state races).
type preWarmMsg struct {
	width   int
	results map[int]string
}

// preWarmCmd returns a tea.Cmd that renders every topic at the given
// width on a background goroutine and delivers the results as a
// preWarmMsg. The goroutine constructs its own local renderer so it
// never touches s.renderers — the cmd is race-free by construction.
// Renderer construction on the goroutine is one extra build (~100ms)
// but the tea loop is unblocked for the duration.
func (s *ViewerStage) preWarmCmd(width int) tea.Cmd {
	if width <= 0 {
		width = defaultRenderWidth
	}
	// Snapshot topics into a local slice so the goroutine never reads
	// through s.topics (which is never mutated today but would be a
	// latent race if that ever changed).
	topics := make([]Topic, len(s.topics))
	copy(topics, s.topics)
	return func() tea.Msg {
		renderer, err := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(width),
		)
		if err != nil {
			// Silent drop — the synchronous path will retry on the
			// tea loop when the user hits a tab. Returning a msg with
			// empty results is a no-op in the Update handler.
			return preWarmMsg{width: width, results: map[int]string{}}
		}
		results := make(map[int]string, len(topics))
		for i, t := range topics {
			body, err := renderMarkdownWith(t.Markdown, renderer)
			if err != nil {
				continue
			}
			results[i] = body
		}
		return preWarmMsg{width: width, results: results}
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
		// Pre-warm the rest of the topics on first sight of each
		// distinct viewport width, so subsequent tab-switches hit the
		// cache and feel instant. Re-dispatch only when the viewport
		// width actually changes (height-only resizes skip it).
		vw := s.viewport.Width
		if vw <= 0 {
			vw = defaultRenderWidth
		}
		if vw != s.preWarmedWidth {
			s.preWarmedWidth = vw
			return s, s.preWarmCmd(vw)
		}
		return s, nil
	case preWarmMsg:
		for idx, body := range m.results {
			key := fmt.Sprintf("%d:%d", idx, m.width)
			// Preserve any value already written synchronously for the
			// current tab (they're identical, but skipping the write
			// keeps Update idempotent and avoids spurious map churn).
			if _, ok := s.rendered[key]; ok {
				continue
			}
			s.rendered[key] = body
		}
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
// footer. Stage.RenderFrame short-circuits to the min-size floor
// when the terminal falls below the 70×20 threshold, so the check
// isn't duplicated here.
func (s *ViewerStage) View() string {
	if s.quit {
		return ""
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
// label set when the rendered strip fits inside the live panel
// budget (initflow.PanelWidth); otherwise falls back to the compact
// short labels so the 4-tab strip still fits on narrow terminals.
func (s *ViewerStage) renderTabs(width int) string {
	full := s.buildTabStrip(false)
	// Measure the rendered full-label strip against the live panel
	// content budget. PanelWidth clamps to the design target on wide
	// terminals and falls back to (width-4) on narrow ones, so the
	// comparison fires only when full labels would actually overflow.
	budget := initflow.PanelWidth(width)
	if budget > 0 && lipgloss.Width(full) > budget {
		return s.buildTabStrip(true)
	}
	return full
}

// buildTabStrip assembles the tab row from s.topics. When useShort
// is true each cell uses the Topic.Short label.
func (s *ViewerStage) buildTabStrip(useShort bool) string {
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
// terminal size. Body area = height minus chrome (header + footer,
// both from Stage.ChromeHeights) minus the inter-section blanks
// rendered by Stage.renderFrame (2 rows) minus the tab strip (1
// row + 1 blank). Floor at 3 so at least a small scroll window is
// visible on short terminals.
func (s *ViewerStage) resizeViewport() {
	w := initflow.PanelWidth(s.width)
	if w <= 0 {
		w = 80
	}
	headerH, footerH := s.ChromeHeights(s.keyHints())
	// +4 = 2 blank separators around the body (Stage.renderFrame) +
	// tab strip (1 row) + blank below the tab strip (1 row).
	chromeRows := headerH + footerH + 4
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
	r, err := s.rendererFor(w)
	if err != nil {
		s.viewport.SetContent("Render error: " + err.Error())
		return
	}
	rendered, err := renderMarkdownWith(s.topics[s.idx].Markdown, r)
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
