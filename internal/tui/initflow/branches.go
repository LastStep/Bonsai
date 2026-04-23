package initflow

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/LastStep/Bonsai/internal/catalog"
	"github.com/LastStep/Bonsai/internal/tui"
)

// Branches category keys (also the map keys used for per-category state).
const (
	branchCatSkills    = "skills"
	branchCatWorkflows = "workflows"
	branchCatProtocols = "protocols"
	branchCatSensors   = "sensors"
	branchCatRoutines  = "routines"
)

// branchCat is a single tab in the BranchesStage — one of the five ability
// categories. introLine1/introLine2 are the two-line description shown above
// the list when this tab is active.
type branchCat struct {
	key         string       // "skills" etc.
	displayName string       // "SKILLS" uppercase label in tab header
	introLine1  string       // first line of the per-tab intro copy
	introLine2  string       // second line
	items       []branchItem // catalog-ordered item list
}

// branchItem is a single selectable row inside a category tab. filePath is
// the catalog ContentPath — shown in the details panel FILE row.
type branchItem struct {
	name        string
	displayName string
	description string
	required    bool
	isDefault   bool
	filePath    string
}

// BranchesStage is the tabbed category picker covering the five ability
// types (Skills / Workflows / Protocols / Sensors / Routines). Per-category
// multi-select with inline-expand details on the focused row.
//
// State:
//   - categories: the five tabs with their items (built in the ctor from the
//     full catalog + agentDef).
//   - catIdx: currently-visible tab.
//   - expanded: global toggle for the inline-expand panel on the focused row.
//   - itemIdx: per-tab focus row index so switching tabs preserves each tab's
//     focus position.
//   - selected: per-tab set of machine-names that are picked.
type BranchesStage struct {
	Stage

	categories []branchCat
	catIdx     int
	expanded   bool
	itemIdx    map[int]int
	selected   map[int]map[string]bool

	// viewport wraps the item list so tabs with more entries than available
	// body rows can scroll while keeping the focused row visible. Reused
	// across renders — SetLines / SetHeight / Follow on each View() call.
	viewport Viewport
}

// BranchesResult is the advance-payload returned from Result() when the user
// hits Enter. Slices preserve the catalog iteration order (alphabetical per
// catalog.loadItems). Required items are always present.
type BranchesResult struct {
	Skills    []string
	Workflows []string
	Protocols []string
	Sensors   []string
	Routines  []string
}

// NewBranchesStage constructs the real Branches stage at rail position 2.
// The ctor walks the five categories of the catalog (filtered to the
// tech-lead agent), seeds required items as pre-selected + immutable, and
// marks each default-list entry as pre-selected with an isDefault flag so
// the renderer can surface the DEFAULT tag.
func NewBranchesStage(ctx StageContext, cat *catalog.Catalog, agentDef *catalog.AgentDef) *BranchesStage {
	label := StageLabels[2]
	base := NewStage(
		2,
		label,
		label.English,
		ctx.Version,
		ctx.ProjectDir,
		ctx.StationDir,
		ctx.AgentDisplay,
		ctx.StartedAt,
	)
	base.applyContextHeader(ctx)

	const agentType = "tech-lead"

	// Build the five category tabs. Each mapper pulls from its own catalog
	// accessor because the per-type shapes don't share a common interface
	// (CatalogItem vs SensorItem vs RoutineItem) — but the fields we need
	// are identical: Name, DisplayName, Description, Required, ContentPath.
	stringSet := func(xs []string) map[string]bool {
		out := make(map[string]bool, len(xs))
		for _, s := range xs {
			out[s] = true
		}
		return out
	}

	skillsDefaults := stringSet(agentDef.DefaultSkills)
	workflowsDefaults := stringSet(agentDef.DefaultWorkflows)
	protocolsDefaults := stringSet(agentDef.DefaultProtocols)
	sensorsDefaults := stringSet(agentDef.DefaultSensors)
	routinesDefaults := stringSet(agentDef.DefaultRoutines)

	skills := make([]branchItem, 0)
	for _, it := range cat.SkillsFor(agentType) {
		skills = append(skills, branchItem{
			name:        it.Name,
			displayName: it.DisplayName,
			description: it.Description,
			required:    it.Required.CompatibleWith(agentType),
			isDefault:   skillsDefaults[it.Name],
			filePath:    it.ContentPath,
		})
	}

	workflows := make([]branchItem, 0)
	for _, it := range cat.WorkflowsFor(agentType) {
		workflows = append(workflows, branchItem{
			name:        it.Name,
			displayName: it.DisplayName,
			description: it.Description,
			required:    it.Required.CompatibleWith(agentType),
			isDefault:   workflowsDefaults[it.Name],
			filePath:    it.ContentPath,
		})
	}

	protocols := make([]branchItem, 0)
	for _, it := range cat.ProtocolsFor(agentType) {
		protocols = append(protocols, branchItem{
			name:        it.Name,
			displayName: it.DisplayName,
			description: it.Description,
			required:    it.Required.CompatibleWith(agentType),
			isDefault:   protocolsDefaults[it.Name],
			filePath:    it.ContentPath,
		})
	}

	sensors := make([]branchItem, 0)
	for _, it := range cat.SensorsFor(agentType) {
		sensors = append(sensors, branchItem{
			name:        it.Name,
			displayName: it.DisplayName,
			description: it.Description,
			required:    it.Required.CompatibleWith(agentType),
			isDefault:   sensorsDefaults[it.Name],
			filePath:    it.ContentPath,
		})
	}

	routines := make([]branchItem, 0)
	for _, it := range cat.RoutinesFor(agentType) {
		routines = append(routines, branchItem{
			name:        it.Name,
			displayName: it.DisplayName,
			description: it.Description,
			required:    it.Required.CompatibleWith(agentType),
			isDefault:   routinesDefaults[it.Name],
			filePath:    it.ContentPath,
		})
	}

	categories := []branchCat{
		{
			key: branchCatSkills, displayName: "SKILLS",
			introLine1: "Rulebooks for specific domains.",
			introLine2: "Standards the Tech Lead consults when doing focused work.",
			items:      skills,
		},
		{
			key: branchCatWorkflows, displayName: "WORKFLOWS",
			introLine1: "Activity-level procedures.",
			introLine2: "Playbooks for multi-phase tasks from intake to ship.",
			items:      workflows,
		},
		{
			key: branchCatProtocols, displayName: "PROTOCOLS",
			introLine1: "Always-on guardrails.",
			introLine2: "Rules every session follows, regardless of task.",
			items:      protocols,
		},
		{
			key: branchCatSensors, displayName: "SENSORS",
			introLine1: "Hook-triggered automations.",
			introLine2: "Event scripts the harness runs without prompting.",
			items:      sensors,
		},
		{
			key: branchCatRoutines, displayName: "ROUTINES",
			introLine1: "Periodic self-maintenance.",
			introLine2: "Recurring checks on a time-based schedule.",
			items:      routines,
		},
	}

	// Seed selected maps from required + default lists. Required items are
	// always-selected; default items start selected but can be toggled off.
	selected := make(map[int]map[string]bool, len(categories))
	itemIdx := make(map[int]int, len(categories))
	for i, c := range categories {
		picks := make(map[string]bool)
		for _, it := range c.items {
			if it.required || it.isDefault {
				picks[it.name] = true
			}
		}
		selected[i] = picks
		itemIdx[i] = 0
	}

	return &BranchesStage{
		Stage:      base,
		categories: categories,
		catIdx:     0,
		itemIdx:    itemIdx,
		selected:   selected,
	}
}

// Init implements tea.Model — no cursor/cmd to fire on entry.
func (s *BranchesStage) Init() tea.Cmd { return nil }

// Update handles tab switching, item focus movement, item toggling, the
// global inline-expand toggle, and Enter-to-advance.
func (s *BranchesStage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = m.Width
		s.height = m.Height
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
			if cat == nil || len(cat.items) == 0 {
				return s, nil
			}
			cur := s.itemIdx[s.catIdx]
			cur--
			if cur < 0 {
				cur = 0 // clamp, no wrap (per plan spec)
			}
			s.itemIdx[s.catIdx] = cur
		case "down", "j":
			cat := s.currentCat()
			if cat == nil || len(cat.items) == 0 {
				return s, nil
			}
			cur := s.itemIdx[s.catIdx]
			cur++
			if cur >= len(cat.items) {
				cur = len(cat.items) - 1 // clamp, no wrap
			}
			s.itemIdx[s.catIdx] = cur
		case " ":
			cat := s.currentCat()
			if cat == nil || len(cat.items) == 0 {
				return s, nil
			}
			row := s.itemIdx[s.catIdx]
			if row < 0 || row >= len(cat.items) {
				return s, nil
			}
			it := cat.items[row]
			if it.required {
				return s, nil // required items cannot be toggled
			}
			picks := s.selected[s.catIdx]
			if picks[it.name] {
				delete(picks, it.name)
			} else {
				picks[it.name] = true
			}
		case "?":
			s.expanded = !s.expanded
		case "enter":
			s.done = true
			return s, nil
		}
	}
	return s, nil
}

// currentCat returns a pointer to the active branchCat or nil if the index is
// out of bounds (defensive — should never happen in practice).
func (s *BranchesStage) currentCat() *branchCat {
	if s.catIdx < 0 || s.catIdx >= len(s.categories) {
		return nil
	}
	return &s.categories[s.catIdx]
}

// View composes the Branches stage body inside the shared frame.
func (s *BranchesStage) View() string {
	return s.renderFrame(s.renderBody(), s.keyHints())
}

// keyHints builds the footer key row for this stage.
func (s *BranchesStage) keyHints() []KeyHint {
	return []KeyHint{
		{Key: "←→", Desc: "tab"},
		{Key: "↑↓", Desc: "move"},
		{Key: "␣", Desc: "toggle"},
		{Key: "?", Desc: "details"},
		{Key: "↵", Desc: "next"},
		{Key: "esc", Desc: "back"},
		{Key: "ctrl-c", Desc: "quit"},
	}
}

// renderBody renders the stage intro + the tab row + the item list +
// the counter summary. Body is centred via centerBlock to match the Vessel
// / Soil visual rhythm.
func (s *BranchesStage) renderBody() string {
	dim := DimStyle()
	bark := LabelStyle()
	white := lipgloss.NewStyle().Foreground(tui.ColorAccent).Bold(true)

	// Title row — mirrors Vessel / Soil pattern.
	var title string
	if s.ensoSafe {
		title = bark.Render(s.label.Kanji) + " " + white.Render(s.label.English)
	} else {
		title = white.Render(s.label.English)
	}

	intro := strings.Join([]string{
		title,
		white.Render("Shape the branches of the Tech Lead."),
		dim.Render("Five categories of abilities — required pinned, defaults pre-picked."),
	}, "\n")

	panelW := PanelWidth(s.width)
	divider := RenderSectionHeader("CATEGORIES", panelW)

	tabRow := s.renderTabs()
	tabIntro := s.renderTabIntro()
	list := s.renderList()
	details := s.renderDetails()
	counter := s.renderCounter()

	// Pin the list to its full budget so `details` lands on the same row
	// regardless of how many items the active tab has. Without this pad,
	// SKILLS (5 items) and SENSORS (9 items) leave the details panel at
	// different vertical positions — tab switching jitters the whole body.
	listH := s.listHeight()
	if listH > 0 {
		rendered := strings.Count(list, "\n") + 1
		if list == "" {
			rendered = 0
		}
		if rendered < listH {
			list = list + strings.Repeat("\n", listH-rendered)
		}
	}

	body := []string{
		intro,
		"",
		"",
		divider,
		"",
		tabRow,
		"",
		tabIntro,
		"",
		list,
		"",
		"", // extra blank pushes DETAILS lower and buys breathing room
		details,
		"",
		dim.Render(counter),
	}
	return centerBlock(strings.Join(body, "\n"), s.width)
}

// renderTabIntro renders the current tab's two-line description block.
// Dim styling keeps it supportive — the ability list is the focal element.
func (s *BranchesStage) renderTabIntro() string {
	cat := s.currentCat()
	if cat == nil {
		return ""
	}
	dim := lipgloss.NewStyle().Foreground(tui.ColorRule2)
	white := lipgloss.NewStyle().Foreground(tui.ColorAccent)

	// Truncate intro lines to the available row width so narrow terminals
	// don't wrap into the list below. Both intro lines are short enough at
	// 80+ cols; the clamp only kicks in at <70 cols where the floor panel
	// is already showing, but keeping the guard is cheap insurance.
	w := s.availableWidth()
	if w < 10 {
		w = 10
	}
	clamp := func(t string) string {
		if lipgloss.Width(t) <= w {
			return t
		}
		rr := []rune(t)
		if len(rr) > w-1 {
			return string(rr[:w-1]) + "…"
		}
		return t
	}
	return white.Render(clamp(cat.introLine1)) + "\n" + dim.Render(clamp(cat.introLine2))
}

// renderDetails renders the fixed-height details block below the list.
// Height is constant across both states (header + 3 ABOUT rows + 1 FILE row
// = 5 visible lines) so toggling `?` never shifts the counter below —
// eliminates the viewport jitter that both inline-expand and variable-height
// collapsed states would cause. ABOUT word-wraps to 3 lines with a trailing
// "…" on overflow (3 × 70 = 210 cells absorbs every current catalog
// description). FILE is tail-truncated with a leading "…". ABOUT + FILE
// values render in ColorAccent (white) for maximum legibility against the
// dim surround.
func (s *BranchesStage) renderDetails() string {
	dim := DimStyle()
	bark := LabelStyle()
	value := lipgloss.NewStyle().Foreground(tui.ColorAccent)

	header := RenderSectionHeader("DETAILS", PanelWidth(s.width))

	const labelW = 10
	const indent = "    "
	const aboutRows = 3
	// contentW clamps to min(s.width - 10, 70). Keeps ABOUT + FILE columns
	// inside the terminal on narrow widths while preserving the 70-cell
	// target (3 rows × 70 = 210 absorbs every current catalog description)
	// on wide ones.
	contentW := s.width - 10
	if contentW > 70 {
		contentW = 70
	}
	if contentW < 20 {
		contentW = 20
	}

	if !s.expanded {
		hint := indent + dim.Render("press ? to reveal ABOUT + FILE on the focused ability")
		blank := indent
		return header + "\n" + hint + "\n" + blank + "\n" + blank + "\n" + blank
	}

	cat := s.currentCat()
	if cat == nil || len(cat.items) == 0 {
		empty := indent + dim.Render("(nothing to show)")
		blank := indent
		return header + "\n" + empty + "\n" + blank + "\n" + blank + "\n" + blank
	}
	row := s.itemIdx[s.catIdx]
	if row < 0 || row >= len(cat.items) {
		empty := indent + dim.Render("(nothing to show)")
		blank := indent
		return header + "\n" + empty + "\n" + blank + "\n" + blank + "\n" + blank
	}
	it := cat.items[row]

	about := it.description
	if about == "" {
		about = "—"
	}
	aboutLines := wrapToWidth(about, contentW)
	if len(aboutLines) > aboutRows {
		last := aboutLines[aboutRows-1]
		rr := []rune(last)
		if len(rr) > contentW-1 {
			last = string(rr[:contentW-1]) + "…"
		} else {
			last = last + "…"
		}
		aboutLines = append(aboutLines[:aboutRows-1], last)
	}
	for len(aboutLines) < aboutRows {
		aboutLines = append(aboutLines, "")
	}

	file := it.filePath
	if file == "" {
		file = "—"
	}
	fileRR := []rune(file)
	if len(fileRR) > contentW {
		file = "…" + string(fileRR[len(fileRR)-contentW+1:])
	}

	aboutRow1 := indent + bark.Render(padRight("ABOUT", labelW)) + value.Render(aboutLines[0])
	aboutRow2 := indent + strings.Repeat(" ", labelW) + value.Render(aboutLines[1])
	aboutRow3 := indent + strings.Repeat(" ", labelW) + value.Render(aboutLines[2])
	fileRow := indent + bark.Render(padRight("FILE", labelW)) + value.Render(file)
	return header + "\n" + aboutRow1 + "\n" + aboutRow2 + "\n" + aboutRow3 + "\n" + fileRow
}

// renderTabs renders the single-row tab header. Active tab is wrapped in
// leaf-green `[  LABEL  ]` brackets; inactive tabs render muted. Leading `‹`
// and trailing `›` chevrons bookend the row to signal the whole strip is
// left/right navigable — 2026-04-22 UX pass dropped the "N / Total" subtitle
// (selection count already lives in the footer counter) in favour of a
// stronger navigation affordance.
func (s *BranchesStage) renderTabs() string {
	muted := lipgloss.NewStyle().Foreground(tui.ColorMuted)
	leaf := lipgloss.NewStyle().Foreground(tui.ColorPrimary).Bold(true)
	bracket := lipgloss.NewStyle().Foreground(tui.ColorPrimary)
	chevron := lipgloss.NewStyle().Foreground(tui.ColorSecondary).Bold(true)

	const colW = 16

	cells := make([]string, 0, len(s.categories))
	for i, c := range s.categories {
		var cell string
		if i == s.catIdx {
			cell = bracket.Render("[ ") + leaf.Render(c.displayName) + bracket.Render(" ]")
		} else {
			cell = "  " + muted.Render(c.displayName) + "  "
		}
		cells = append(cells, lipgloss.PlaceHorizontal(colW, lipgloss.Center, cell))
	}

	row := strings.Join(cells, " ")
	return chevron.Render("‹") + " " + row + " " + chevron.Render("›")
}

// renderList renders the item rows for the current tab. Details for the
// focused ability render in a fixed-height block below the list (see
// renderDetails) rather than inline — the list never jitters when `?`
// toggles expansion.
//
// When the tab has more items than the available body height allows, the
// list is wrapped in a Viewport and scrolls focus-follows-cursor. Height
// budget is computed from s.height less chrome and non-list body rows;
// floor at 3 so at least a tiny window is visible even on short terminals.
func (s *BranchesStage) renderList() string {
	cat := s.currentCat()
	if cat == nil {
		return ""
	}
	if len(cat.items) == 0 {
		dim := lipgloss.NewStyle().Foreground(tui.ColorRule2)
		return dim.Render("  (no items in this category)")
	}

	rows := make([]string, 0, len(cat.items))
	for i := range cat.items {
		rows = append(rows, s.renderRow(i))
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

// listHeight computes the visible-row budget for the item list. The
// Branches stage body has a fixed footprint above + below the list:
//
//	intro (3) + blanks (2) + divider (1) + blank (1)
//	  + tabs (1) + blank (1) + tabIntro (2) + blank (1)       = 12 rows above
//	blanks (2) + details (5) + blank (1) + counter (1)         = 9 rows below
//
// (Two blanks sit between list and details so the DETAILS panel drops low
// enough to feel like a distinct block, with breathing room between it
// and the counter too.)
//
// Total non-list body rows = 21. Chrome (header 2 + blank + rail 2 + blank
// + blank + footer 2 = 10) lives outside renderBody; renderFrame's body
// padding absorbs the remainder. Floor at 3 so something always shows.
func (s *BranchesStage) listHeight() int {
	if s.height <= 0 {
		return 0 // unknown — render all rows, let harness clip
	}
	const chromeRows = 10    // header + rail + footer + separators
	const fixedBodyRows = 21 // see comment above
	h := s.height - chromeRows - fixedBodyRows
	if h < 3 {
		h = 3
	}
	return h
}

// renderRow renders a single ability row at index idx within the current
// tab. Focused rows get a Leaf "│ " left border; selected rows use ◆ (Leaf),
// unselected use ◇ (dim). Required items show an inline gold "*" glyph
// after the ability name; the green dot (◆) already conveys "pre-selected"
// so no separate DEFAULT tag.
//
// Column widths come from ClampColumns(availableW). The old tag column
// (12 cells) is reclaimed by the description column since there is no
// trailing tag; nameW caps at 24, descW absorbs the remainder. When descW
// floors to 0 (very tight terminal) the description column is dropped
// entirely.
//
// Focus colouring: ability name switches to white-bold (ColorAccent),
// description lifts from ColorRule2 (dim) to ColorMuted so the focused
// row reads brighter than its unfocused neighbours.
func (s *BranchesStage) renderRow(idx int) string {
	cat := s.currentCat()
	if cat == nil || idx < 0 || idx >= len(cat.items) {
		return ""
	}
	it := cat.items[idx]
	focused := idx == s.itemIdx[s.catIdx]
	selected := s.selected[s.catIdx][it.name]

	leaf := lipgloss.NewStyle().Foreground(tui.ColorPrimary)
	dim := lipgloss.NewStyle().Foreground(tui.ColorRule2)

	// Glyph: ◆ when selected (Leaf), ◇ when not (dim).
	glyph := "◇"
	glyphStyle := dim
	if selected {
		glyph = "◆"
		glyphStyle = leaf
	}

	// Left border — Leaf "│ " for focused row, two spaces otherwise
	// (mirrors soil.go:210-215 AdaptiveColor-friendly pattern).
	border := "  "
	if focused {
		border = lipgloss.NewStyle().Foreground(tui.ColorPrimary).Render("│ ")
	}

	nameColW, descColW, tagColW := ClampColumns(s.availableWidth())
	// Tag column now carries a single "*" (or blank) instead of
	// "(required)" / "DEFAULT" — reclaim the spare space for desc.
	descColW += tagColW - 2
	tagColW = 2

	// Required asterisk lives after the ability name. Truncate the name
	// against (nameColW - 2) when required so the " *" always fits.
	name := it.displayName
	if name == "" {
		name = it.name
	}
	nameBudget := nameColW
	if it.required {
		nameBudget = nameColW - 2
	}
	if lipgloss.Width(name) > nameBudget {
		rr := []rune(name)
		if len(rr) > nameBudget-1 && nameBudget > 1 {
			name = string(rr[:nameBudget-1]) + "…"
		}
	}
	// Name colour encodes two orthogonal states:
	//   selected (any focus)    → leaf-green (bold when focused)
	//   unselected + focused    → white bold  (FocusedNameStyle)
	//   unselected + unfocused  → subtle      (UnfocusedNameStyle)
	// Green for "picked" matches the glyph colour so selection reads as
	// a single leaf-coloured block rather than a gold+white mix.
	var nameStyle lipgloss.Style
	switch {
	case selected && focused:
		nameStyle = leaf.Bold(true)
	case selected:
		nameStyle = leaf
	case focused:
		nameStyle = FocusedNameStyle()
	default:
		nameStyle = UnfocusedNameStyle()
	}
	nameText := nameStyle.Render(name)
	if it.required {
		nameText += " " + RequiredGlyph()
	}
	nameCol := padRight(nameText, nameColW)

	// Description column — ColorMuted when focused (brighter), ColorRule2
	// otherwise. descColW=0 is the "tight-terminal" signal: drop entirely.
	var descCol string
	if descColW > 0 {
		desc := it.description
		if lipgloss.Width(desc) > descColW {
			rr := []rune(desc)
			if len(rr) > descColW-1 {
				desc = string(rr[:descColW-1]) + "…"
			}
		}
		descStyle := dim
		if focused {
			descStyle = FocusedDescStyle()
		}
		descCol = " " + descStyle.Render(padRight(desc, descColW))
	}

	// Trailing tag slot stays empty — reserved for column alignment only.
	_ = tagColW

	return border + glyphStyle.Render(glyph) + " " + nameCol + descCol
}

// availableWidth returns the per-row budget after subtracting side-padding
// used by centerBlock. Floor at 0 so ClampColumns returns zeros on degenerate
// terminals (rendered only from renderFrame's post-floor path, so in practice
// always ≥ MinTerminalWidth - 4).
func (s *BranchesStage) availableWidth() int {
	w := s.width - 4
	if w < 0 {
		return 0
	}
	return w
}

// renderCounter renders the summary line at the bottom of the stage body.
// Format: "N abilities selected · across 5 categories".
func (s *BranchesStage) renderCounter() string {
	total := 0
	for i := range s.categories {
		total += len(s.selected[i])
	}
	return fmt.Sprintf("%d abilities selected · across %d categories", total, len(s.categories))
}

// Result returns a BranchesResult with per-category slices of selected
// machine-names, preserving catalog order. Required items are always in the
// result (they're pre-selected + immutable).
func (s *BranchesStage) Result() any {
	pick := func(idx int) []string {
		picks := s.selected[idx]
		cat := s.categories[idx]
		out := make([]string, 0, len(picks))
		for _, it := range cat.items {
			if picks[it.name] {
				out = append(out, it.name)
			}
		}
		return out
	}

	res := BranchesResult{}
	for i, c := range s.categories {
		slice := pick(i)
		switch c.key {
		case branchCatSkills:
			res.Skills = slice
		case branchCatWorkflows:
			res.Workflows = slice
		case branchCatProtocols:
			res.Protocols = slice
		case branchCatSensors:
			res.Sensors = slice
		case branchCatRoutines:
			res.Routines = slice
		}
	}
	return res
}

// Reset clears the completion flag so re-entry behaves correctly. The
// per-tab selections, focus cursor, expand toggle, and current tab index
// are all preserved so Esc-and-return restores the user's state verbatim.
func (s *BranchesStage) Reset() tea.Cmd {
	s.done = false
	return nil
}

// wrapToWidth word-wraps text into lines whose visible cell-width is ≤ width,
// breaking on spaces. A single word wider than width is hard-wrapped by rune
// so an oversized token still fits. Returns at least one line; an empty input
// yields []string{""}.
func wrapToWidth(text string, width int) []string {
	if width <= 0 {
		return []string{text}
	}
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{""}
	}
	var lines []string
	cur := ""
	for _, w := range words {
		if cur == "" {
			if lipgloss.Width(w) > width {
				rr := []rune(w)
				for len(rr) > width {
					lines = append(lines, string(rr[:width]))
					rr = rr[width:]
				}
				cur = string(rr)
			} else {
				cur = w
			}
			continue
		}
		candidate := cur + " " + w
		if lipgloss.Width(candidate) <= width {
			cur = candidate
		} else {
			lines = append(lines, cur)
			cur = w
		}
	}
	if cur != "" {
		lines = append(lines, cur)
	}
	return lines
}
