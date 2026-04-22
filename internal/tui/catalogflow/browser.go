package catalogflow

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/LastStep/Bonsai/internal/catalog"
	"github.com/LastStep/Bonsai/internal/tui"
	"github.com/LastStep/Bonsai/internal/tui/initflow"
)

// timeZero returns the zero-value time.Time used for stages that have
// no session-start anchor (catalog is a single-shot browser — no
// ELAPSED row to render, so any zero-value is safe).
func timeZero() time.Time { return time.Time{} }

// category is one of the seven catalog section tabs. The tab's header
// cell renders as "LABEL (N)" where N counts the filtered entries;
// filterHidZero is the agent-filter's greyed-out suffix signal (tabs
// with zero filtered items render with a muted (0) suffix but stay in
// the tab strip so the user sees what their filter excludes).
type category struct {
	key         string
	displayName string
	entries     []Entry
}

// BrowserStage is the BubbleTea model for `bonsai catalog`. Seven tabs
// cover the seven catalog sections; per-tab focus is preserved across
// tab switches so `← → ↑ ↓ ? q` always reads the current tab's state
// verbatim.
//
// Fields:
//   - categories: seven tabs, in catalog-section order.
//   - catIdx: 0..6 — currently-visible tab.
//   - itemIdx: per-tab focus-row index.
//   - expanded: global `?` toggle for the inline-expand detail block.
//   - viewport: scrolls the list when entries overflow body rows.
//
// Embeds initflow.Stage with railHidden=true so the shared chrome
// composer skips the enso rail row (catalog is read-only — it doesn't
// encode a mutation process, so no rail segments to walk).
type BrowserStage struct {
	initflow.Stage

	categories []category
	catIdx     int
	itemIdx    map[int]int
	expanded   bool
	viewport   initflow.Viewport

	// quit flips to true on any quit key; View returns "" afterwards and
	// the tea.Program is told to Quit via the Update return cmd.
	quit bool
}

// shortLabelThreshold is the width (in columns) at or below which the
// tab strip switches to short-label mode. The seven full tabs plus
// "(N)" counts plus 2-space separators render at ~96 visible cells;
// below that they wrap or clip against the 70-col minimum floor. Short
// labels bring the strip inside a 70-col frame.
const shortLabelThreshold = 96

// shortLabel returns the compact form of a tab label for use under
// narrow widths. Target: keep the full 7-tab strip inside the 70-col
// minimum-width floor. All labels collapse to <=5 chars so the strip
// + "(N)" counts + 1-space separators fit with margin.
func shortLabel(full string) string {
	switch full {
	case "AGENTS":
		return "AGENT"
	case "SKILLS":
		return "SKILL"
	case "WORKFLOWS":
		return "FLOWS"
	case "PROTOCOLS":
		return "PROTO"
	case "SENSORS":
		return "SENSE"
	case "ROUTINES":
		return "RTNES"
	case "SCAFFOLDING":
		return "SCAFF"
	default:
		return full
	}
}

// NewBrowser constructs a BrowserStage from a loaded catalog, applying
// the optional -a/--agent filter. An empty agentFilter shows every
// entry in every tab. With a non-empty filter, per-section cat.*For
// accessors drop incompatible items; tabs that end up with zero
// entries render a greyed (0) suffix but are not removed from the
// tab strip.
//
// projectDir is the absolute path to the invoking shell's cwd — passed
// through so the header's right-row-2 renders a useful path fragment
// instead of a bare "./". Resolve at the cmd/catalog.go callsite via
// the existing mustCwd() helper.
//
// Called from cmd/catalog.go when stdout is a TTY. Non-TTY invocations
// stay on the existing static-render path.
func NewBrowser(cat *catalog.Catalog, agentFilter string, projectDir string) *BrowserStage {
	categories := buildCategories(cat, agentFilter)

	itemIdx := make(map[int]int, len(categories))
	for i := range categories {
		itemIdx[i] = 0
	}

	base := initflow.NewStage(
		0,
		initflow.StageLabel{Kanji: "録", Kana: "ロク", English: "CATALOG"},
		"CATALOG",
		"",         // version: catalog's header renders no version chip
		projectDir, // projectDir: threaded through so header right-row-2 renders the cwd
		"",
		"",
		// Zero time — catalog has no elapsed counter. Matches the
		// initflow precedent (StageContext.StartedAt zero-value is safe
		// for stages that never render an ELAPSED row).
		timeZero(),
	)
	base.SetRailHidden(true)
	base.SetHeaderAction("CATALOG")
	// Empty rightLabel hides the right-block row 1 — catalog is global,
	// not scoped to a project path.
	base.SetHeaderRightLabel("")

	return &BrowserStage{
		Stage:      base,
		categories: categories,
		catIdx:     0,
		itemIdx:    itemIdx,
		expanded:   false,
	}
}

// buildCategories walks the seven catalog sections and packs each into
// the per-tab category shape. Per-section Meta packing:
//
//   - Agents: Meta = nil.
//   - Skills / Workflows / Protocols: Meta = nil.
//   - Sensors: Meta = {"Event": event, "Matcher": matcher} (matcher omitted if empty).
//   - Routines: Meta = {"Frequency": freq}.
//   - Scaffolding: Meta = {"If Removed": affects}.
//
// agentFilter applies via the catalog's per-section *For accessors.
// Agents and Scaffolding are unfiltered — Agents has no per-agent
// compat concept (every agent IS an agent), and Scaffolding is a
// project-level decision independent of agent choice.
func buildCategories(cat *catalog.Catalog, agentFilter string) []category {
	out := make([]category, 0, 7)

	// Agents — always unfiltered.
	{
		entries := make([]Entry, 0, len(cat.Agents))
		for _, a := range cat.Agents {
			display := a.DisplayName
			if display == "" {
				display = catalog.DisplayNameFrom(a.Name)
			}
			entries = append(entries, Entry{
				Name:        a.Name,
				DisplayName: display,
				Description: a.Description,
			})
		}
		out = append(out, category{key: "agents", displayName: "AGENTS", entries: entries})
	}

	// Skills.
	{
		items := cat.Skills
		if agentFilter != "" {
			items = cat.SkillsFor(agentFilter)
		}
		entries := make([]Entry, 0, len(items))
		for _, it := range items {
			entries = append(entries, Entry{
				Name:        it.Name,
				DisplayName: it.DisplayName,
				Description: it.Description,
				Agents:      it.Agents.String(),
				Required:    it.Required.String(),
			})
		}
		out = append(out, category{key: "skills", displayName: "SKILLS", entries: entries})
	}

	// Workflows.
	{
		items := cat.Workflows
		if agentFilter != "" {
			items = cat.WorkflowsFor(agentFilter)
		}
		entries := make([]Entry, 0, len(items))
		for _, it := range items {
			entries = append(entries, Entry{
				Name:        it.Name,
				DisplayName: it.DisplayName,
				Description: it.Description,
				Agents:      it.Agents.String(),
				Required:    it.Required.String(),
			})
		}
		out = append(out, category{key: "workflows", displayName: "WORKFLOWS", entries: entries})
	}

	// Protocols.
	{
		items := cat.Protocols
		if agentFilter != "" {
			items = cat.ProtocolsFor(agentFilter)
		}
		entries := make([]Entry, 0, len(items))
		for _, it := range items {
			entries = append(entries, Entry{
				Name:        it.Name,
				DisplayName: it.DisplayName,
				Description: it.Description,
				Agents:      it.Agents.String(),
				Required:    it.Required.String(),
			})
		}
		out = append(out, category{key: "protocols", displayName: "PROTOCOLS", entries: entries})
	}

	// Sensors — Meta carries Event + Matcher (matcher elided if empty).
	{
		items := cat.Sensors
		if agentFilter != "" {
			items = cat.SensorsFor(agentFilter)
		}
		entries := make([]Entry, 0, len(items))
		for _, it := range items {
			meta := map[string]string{"Event": it.Event}
			if it.Matcher != "" {
				meta["Matcher"] = it.Matcher
			}
			entries = append(entries, Entry{
				Name:        it.Name,
				DisplayName: it.DisplayName,
				Description: it.Description,
				Meta:        meta,
				Agents:      it.Agents.String(),
				Required:    it.Required.String(),
			})
		}
		out = append(out, category{key: "sensors", displayName: "SENSORS", entries: entries})
	}

	// Routines — Meta carries Frequency.
	{
		items := cat.Routines
		if agentFilter != "" {
			items = cat.RoutinesFor(agentFilter)
		}
		entries := make([]Entry, 0, len(items))
		for _, it := range items {
			entries = append(entries, Entry{
				Name:        it.Name,
				DisplayName: it.DisplayName,
				Description: it.Description,
				Meta:        map[string]string{"Frequency": it.Frequency},
				Agents:      it.Agents.String(),
				Required:    it.Required.String(),
			})
		}
		out = append(out, category{key: "routines", displayName: "ROUTINES", entries: entries})
	}

	// Scaffolding — always unfiltered. Required is a simple bool so
	// Required renders as "yes" / "" (matching the static-render path).
	// Meta carries If Removed (Affects).
	{
		entries := make([]Entry, 0, len(cat.Scaffolding))
		for _, s := range cat.Scaffolding {
			req := ""
			if s.Required {
				req = "yes"
			}
			meta := map[string]string{}
			if s.Affects != "" {
				meta["If Removed"] = s.Affects
			}
			entries = append(entries, Entry{
				Name:        s.Name,
				DisplayName: s.DisplayName,
				Description: s.Description,
				Meta:        meta,
				Required:    req,
			})
		}
		out = append(out, category{key: "scaffolding", displayName: "SCAFFOLDING", entries: entries})
	}

	return out
}

// Init implements tea.Model.
func (s *BrowserStage) Init() tea.Cmd { return nil }

// Update handles tab cycling, focus movement, expand toggle, and quit.
// Keys:
//
//   - left / h / right / l — tab cycle (wraps both directions).
//   - up / k / down / j     — focus clamp (no wrap).
//   - ?                     — inline-expand toggle.
//   - q / esc / ctrl-c / enter — quit.
func (s *BrowserStage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		s.SetSize(m.Width, m.Height)
	case tea.KeyMsg:
		switch m.String() {
		case "left", "h":
			if len(s.categories) == 0 {
				return s, nil
			}
			s.catIdx = (s.catIdx - 1 + len(s.categories)) % len(s.categories)
		case "right", "l":
			if len(s.categories) == 0 {
				return s, nil
			}
			s.catIdx = (s.catIdx + 1) % len(s.categories)
		case "up", "k":
			cat := s.currentCat()
			if cat == nil || len(cat.entries) == 0 {
				return s, nil
			}
			cur := s.itemIdx[s.catIdx]
			cur--
			if cur < 0 {
				cur = 0
			}
			s.itemIdx[s.catIdx] = cur
		case "down", "j":
			cat := s.currentCat()
			if cat == nil || len(cat.entries) == 0 {
				return s, nil
			}
			cur := s.itemIdx[s.catIdx]
			cur++
			if cur >= len(cat.entries) {
				cur = len(cat.entries) - 1
			}
			s.itemIdx[s.catIdx] = cur
		case "?":
			s.expanded = !s.expanded
		case "q", "esc", "ctrl+c", "enter":
			s.quit = true
			s.MarkDone()
			return s, tea.Quit
		}
	}
	return s, nil
}

// Title implements harness.Step.
func (s *BrowserStage) Title() string { return "CATALOG" }

// Result implements harness.Step. Catalog is read-only — no payload.
func (s *BrowserStage) Result() any { return nil }

// View composes the full frame: header + tab strip + list + optional
// inline-expand detail block + footer. No rail (SetRailHidden(true)).
//
// Routing: header/footer go through the embedded Stage.RenderFrame so
// the stored headerAction / headerRightLabel / projectDir (set in
// NewBrowser) drive the render. This keeps the catalog's chrome
// consistent with every other initflow-embedded stage and removes a
// duplicated copy of the padding/truncation math previously inlined
// here. The body we hand to RenderFrame is just the tab strip + list +
// optional details block.
func (s *BrowserStage) View() string {
	if s.quit {
		return ""
	}

	width := s.Width()
	if width <= 0 {
		width = 80
	}

	// Min-size floor short-circuit is redundant with RenderFrame's own
	// guard, but keep it here so View returns the floor panel directly
	// without constructing a body that would be discarded.
	if initflow.TerminalTooSmall(s.Width(), s.Height()) {
		return initflow.RenderMinSizeFloor(s.Width(), s.Height())
	}

	tabRow := s.renderTabs()
	list := s.renderList()
	details := s.renderDetails()

	// Build body rows — tab strip + list + optional details. centerBlock
	// matches the visual rhythm of the init-flow stages.
	bodyParts := []string{
		tabRow,
		"",
		list,
	}
	if s.expanded {
		bodyParts = append(bodyParts, "", details)
	}
	body := initflow.CenterBlock(strings.Join(bodyParts, "\n"), width)

	return s.RenderFrame(body, s.keyHints())
}

// keyHints returns the footer key row for the browser. Single-stage
// flow — no back/next, only tab/focus/details/quit.
func (s *BrowserStage) keyHints() []initflow.KeyHint {
	return []initflow.KeyHint{
		{Key: "←→", Desc: "tabs"},
		{Key: "↑↓", Desc: "focus"},
		{Key: "?", Desc: "details"},
		{Key: "q", Desc: "quit"},
	}
}

// renderTabs renders the single-row tab header. Active tab is bold
// ColorPrimary; inactive tabs are ColorMuted. Count suffix `(N)` is
// attached to every tab — greyed out when N==0 to signal that an
// agent filter excluded all entries in that tab (but the tab still
// renders so the user sees what's being filtered).
//
// Narrow-width adaptation: when terminal width is below
// shortLabelThreshold (96 cols), swaps full labels for compact forms
// (see shortLabel). The min-size floor is 70×20 — short labels keep
// the full 7-tab strip inside 70 cols without wrapping or clipping.
func (s *BrowserStage) renderTabs() string {
	active := lipgloss.NewStyle().Foreground(tui.ColorPrimary).Bold(true)
	muted := lipgloss.NewStyle().Foreground(tui.ColorMuted)
	dim := lipgloss.NewStyle().Foreground(tui.ColorRule2)

	useShort := s.Width() > 0 && s.Width() < shortLabelThreshold

	cells := make([]string, 0, len(s.categories))
	for i, c := range s.categories {
		label := c.displayName
		if useShort {
			label = shortLabel(label)
		}
		countSuffix := fmt.Sprintf(" (%d)", len(c.entries))
		var cell string
		switch {
		case i == s.catIdx:
			cell = active.Render(label) + dim.Render(countSuffix)
		case len(c.entries) == 0:
			cell = dim.Render(label + countSuffix)
		default:
			cell = muted.Render(label) + dim.Render(countSuffix)
		}
		cells = append(cells, cell)
	}

	sep := "  "
	if useShort {
		sep = " "
	}
	return strings.Join(cells, sep)
}

// renderList renders the entries of the current tab. Each row goes
// through the per-entry renderer (renderEntry). Rows that overflow
// the visible body area are clamped by the embedded viewport — focus
// stays visible.
func (s *BrowserStage) renderList() string {
	cat := s.currentCat()
	if cat == nil {
		return ""
	}
	if len(cat.entries) == 0 {
		dim := lipgloss.NewStyle().Foreground(tui.ColorRule2)
		return dim.Render("  (no entries in this category)")
	}

	rows := make([]string, 0, len(cat.entries))
	for i := range cat.entries {
		rows = append(rows, renderEntry(cat.entries[i], i == s.itemIdx[s.catIdx]))
	}

	listH := s.listHeight()
	if listH <= 0 || listH >= len(rows) {
		return strings.Join(rows, "\n")
	}
	s.viewport.SetLines(rows)
	s.viewport.SetHeight(listH)
	s.viewport.Follow(s.itemIdx[s.catIdx])
	return s.viewport.View()
}

// renderDetails renders the inline-expand block for the focused
// entry. Shows Agents, Required, and per-category Meta keys (Event,
// Matcher, Frequency, If Removed). Entries whose tab carries no
// focused row (empty tab) render a muted "nothing to show".
func (s *BrowserStage) renderDetails() string {
	cat := s.currentCat()
	if cat == nil || len(cat.entries) == 0 {
		dim := lipgloss.NewStyle().Foreground(tui.ColorRule2)
		return dim.Render("  (nothing to show)")
	}
	row := s.itemIdx[s.catIdx]
	if row < 0 || row >= len(cat.entries) {
		return ""
	}
	return renderDetailsBlock(cat.entries[row], initflow.PanelWidth(s.Width()))
}

// listHeight computes the visible-row budget for the list. Reserves
// rows for chrome (header 2 + 2 blanks + footer 2 = 6), tab strip (1
// + 1 blank), and optional details (5 rows when expanded). Floor at
// 3 so at least a tiny window is visible on short terminals.
func (s *BrowserStage) listHeight() int {
	h := s.Height()
	if h <= 0 {
		return 0
	}
	const chromeRows = 6
	const tabRows = 2
	bodyBelow := 0
	if s.expanded {
		bodyBelow = 6 // 1 blank + details block (≈5 rows)
	}
	out := h - chromeRows - tabRows - bodyBelow
	if out < 3 {
		out = 3
	}
	return out
}

// currentCat returns a pointer to the active category or nil if the
// index is out of bounds.
func (s *BrowserStage) currentCat() *category {
	if s.catIdx < 0 || s.catIdx >= len(s.categories) {
		return nil
	}
	return &s.categories[s.catIdx]
}
