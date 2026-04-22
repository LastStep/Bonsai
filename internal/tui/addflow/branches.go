package addflow

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/LastStep/Bonsai/internal/catalog"
	"github.com/LastStep/Bonsai/internal/config"
	"github.com/LastStep/Bonsai/internal/tui"
	"github.com/LastStep/Bonsai/internal/tui/initflow"
)

// Branches category keys — same five ability types as initflow.BranchesStage.
const (
	branchCatSkills    = "skills"
	branchCatWorkflows = "workflows"
	branchCatProtocols = "protocols"
	branchCatSensors   = "sensors"
	branchCatRoutines  = "routines"
)

// branchCat is a single tab in the BranchesStage — mirrors initflow's
// branchCat but kept package-local so addflow does not couple to
// unexported initflow types.
type branchCat struct {
	key         string
	displayName string
	introLine1  string
	introLine2  string
	items       []branchItem
}

type branchItem struct {
	name        string
	displayName string
	description string
	required    bool
	isDefault   bool
	filePath    string
}

// BranchesStage is the tabbed ability picker at rail position 1 (枝 BRANCHES).
// Keystroke model matches initflow.BranchesStage: ← → tab, ↑ ↓ focus, ␣
// toggle, ? details, ↵ advance. Required items are pre-selected and
// immutable; defaults are pre-selected but toggleable.
//
// Result: BranchesResult with per-category slices of selected machine names.
type BranchesStage struct {
	initflow.Stage

	categories []branchCat
	catIdx     int
	expanded   bool
	itemIdx    map[int]int
	selected   map[int]map[string]bool

	viewport initflow.Viewport
}

// BranchesContext bundles everything a Branches ctor needs. Kept separate
// from StageContext so the add-items branch can pass an Installed pointer
// without polluting the shared context type.
type BranchesContext struct {
	Cat       *catalog.Catalog
	AgentType string
	AgentDef  *catalog.AgentDef
	// Installed is nil on the new-agent branch; populated on add-items.
	Installed *config.InstalledAgent
}

// NewNewAgentBranches constructs a Branches stage for the new-agent branch —
// all five tabs, defaults seeded from agentDef.
func NewNewAgentBranches(ctx initflow.StageContext, gctx BranchesContext) *BranchesStage {
	return newBranches(ctx, gctx, false)
}

// NewAddItemsBranches constructs a Branches stage for the add-items branch —
// filters each category to uninstalled items, drops empty tabs.
func NewAddItemsBranches(ctx initflow.StageContext, gctx BranchesContext) *BranchesStage {
	return newBranches(ctx, gctx, true)
}

// newBranches is the shared constructor. filter=true excludes already-installed
// items per category and drops empty tabs.
func newBranches(ctx initflow.StageContext, gctx BranchesContext, filter bool) *BranchesStage {
	label := StageLabels[StageIdxBranches]
	base := initflow.NewStage(
		StageIdxBranches,
		label,
		label.English,
		ctx.Version,
		ctx.ProjectDir,
		ctx.StationDir,
		ctx.AgentDisplay,
		ctx.StartedAt,
	)
	base.SetRailLabels(StageLabels)

	agentDef := gctx.AgentDef
	agentType := gctx.AgentType
	cat := gctx.Cat

	stringSet := func(xs []string) map[string]bool {
		out := make(map[string]bool, len(xs))
		for _, s := range xs {
			out[s] = true
		}
		return out
	}

	var (
		installedSkills, installedWorkflows  map[string]bool
		installedProtocols, installedSensors map[string]bool
		installedRoutines                    map[string]bool
	)
	if filter && gctx.Installed != nil {
		installedSkills = stringSet(gctx.Installed.Skills)
		installedWorkflows = stringSet(gctx.Installed.Workflows)
		installedProtocols = stringSet(gctx.Installed.Protocols)
		installedSensors = stringSet(gctx.Installed.Sensors)
		installedRoutines = stringSet(gctx.Installed.Routines)
	}

	skillsDefaults := stringSet(agentDef.DefaultSkills)
	workflowsDefaults := stringSet(agentDef.DefaultWorkflows)
	protocolsDefaults := stringSet(agentDef.DefaultProtocols)
	sensorsDefaults := stringSet(agentDef.DefaultSensors)
	routinesDefaults := stringSet(agentDef.DefaultRoutines)

	skills := make([]branchItem, 0)
	for _, it := range cat.SkillsFor(agentType) {
		if filter && installedSkills[it.Name] {
			continue
		}
		skills = append(skills, branchItem{
			name:        it.Name,
			displayName: it.DisplayName,
			description: it.Description,
			required:    it.Required.CompatibleWith(agentType),
			isDefault:   !filter && skillsDefaults[it.Name],
			filePath:    it.ContentPath,
		})
	}
	workflows := make([]branchItem, 0)
	for _, it := range cat.WorkflowsFor(agentType) {
		if filter && installedWorkflows[it.Name] {
			continue
		}
		workflows = append(workflows, branchItem{
			name:        it.Name,
			displayName: it.DisplayName,
			description: it.Description,
			required:    it.Required.CompatibleWith(agentType),
			isDefault:   !filter && workflowsDefaults[it.Name],
			filePath:    it.ContentPath,
		})
	}
	protocols := make([]branchItem, 0)
	for _, it := range cat.ProtocolsFor(agentType) {
		if filter && installedProtocols[it.Name] {
			continue
		}
		protocols = append(protocols, branchItem{
			name:        it.Name,
			displayName: it.DisplayName,
			description: it.Description,
			required:    it.Required.CompatibleWith(agentType),
			isDefault:   !filter && protocolsDefaults[it.Name],
			filePath:    it.ContentPath,
		})
	}
	sensors := make([]branchItem, 0)
	for _, it := range cat.SensorsFor(agentType) {
		// routine-check is auto-managed — never user-picked.
		if it.Name == "routine-check" {
			continue
		}
		if filter && installedSensors[it.Name] {
			continue
		}
		sensors = append(sensors, branchItem{
			name:        it.Name,
			displayName: it.DisplayName,
			description: it.Description,
			required:    it.Required.CompatibleWith(agentType),
			isDefault:   !filter && sensorsDefaults[it.Name],
			filePath:    it.ContentPath,
		})
	}
	routines := make([]branchItem, 0)
	for _, it := range cat.RoutinesFor(agentType) {
		if filter && installedRoutines[it.Name] {
			continue
		}
		routines = append(routines, branchItem{
			name:        it.Name,
			displayName: it.DisplayName,
			description: it.Description,
			required:    it.Required.CompatibleWith(agentType),
			isDefault:   !filter && routinesDefaults[it.Name],
			filePath:    it.ContentPath,
		})
	}

	all := []branchCat{
		{
			key: branchCatSkills, displayName: "SKILLS",
			introLine1: "Rulebooks for specific domains.",
			introLine2: "Standards the agent consults when doing focused work.",
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

	// Add-items branch: drop empty tabs entirely.
	categories := make([]branchCat, 0, len(all))
	for _, c := range all {
		if filter && len(c.items) == 0 {
			continue
		}
		categories = append(categories, c)
	}

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

// Init implements tea.Model — nothing to fire on entry.
func (s *BranchesStage) Init() tea.Cmd { return nil }

// Update handles tab cycling, row focus, toggle, inline-expand, and Enter.
func (s *BranchesStage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			c := s.currentCat()
			if c == nil || len(c.items) == 0 {
				return s, nil
			}
			cur := s.itemIdx[s.catIdx] - 1
			if cur < 0 {
				cur = 0
			}
			s.itemIdx[s.catIdx] = cur
		case "down", "j":
			c := s.currentCat()
			if c == nil || len(c.items) == 0 {
				return s, nil
			}
			cur := s.itemIdx[s.catIdx] + 1
			if cur >= len(c.items) {
				cur = len(c.items) - 1
			}
			s.itemIdx[s.catIdx] = cur
		case " ":
			c := s.currentCat()
			if c == nil || len(c.items) == 0 {
				return s, nil
			}
			row := s.itemIdx[s.catIdx]
			if row < 0 || row >= len(c.items) {
				return s, nil
			}
			it := c.items[row]
			if it.required {
				return s, nil
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
			s.MarkDone()
			return s, nil
		}
	}
	return s, nil
}

func (s *BranchesStage) currentCat() *branchCat {
	if s.catIdx < 0 || s.catIdx >= len(s.categories) {
		return nil
	}
	return &s.categories[s.catIdx]
}

// View composes the stage body inside the shared frame.
func (s *BranchesStage) View() string {
	return s.RenderFrame(s.renderBody(), s.keyHints())
}

func (s *BranchesStage) keyHints() []initflow.KeyHint {
	return []initflow.KeyHint{
		{Key: "←→", Desc: "tab"},
		{Key: "↑↓", Desc: "move"},
		{Key: "␣", Desc: "toggle"},
		{Key: "?", Desc: "details"},
		{Key: "↵", Desc: "next"},
		{Key: "esc", Desc: "back"},
		{Key: "ctrl-c", Desc: "quit"},
	}
}

func (s *BranchesStage) renderBody() string {
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
		white.Render("Graft the agent's abilities."),
		dim.Render("Five categories — required pinned, defaults pre-picked."),
	}, "\n")

	panelW := initflow.PanelWidth(s.Width())
	divider := initflow.RenderSectionHeader("CATEGORIES", panelW)

	if len(s.categories) == 0 {
		empty := dim.Render("  (nothing to graft — all abilities already installed)")
		body := []string{intro, "", "", divider, "", empty}
		return initflow.CenterBlock(strings.Join(body, "\n"), s.Width())
	}

	tabRow := s.renderTabs()
	tabIntro := s.renderTabIntro()
	list := s.renderList()
	details := s.renderDetails()
	counter := s.renderCounter()

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
		"",
		details,
		"",
		dim.Render(counter),
	}
	return initflow.CenterBlock(strings.Join(body, "\n"), s.Width())
}

func (s *BranchesStage) renderTabIntro() string {
	c := s.currentCat()
	if c == nil {
		return ""
	}
	dim := initflow.DimStyle()
	white := lipgloss.NewStyle().Foreground(tui.ColorAccent)
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
	return white.Render(clamp(c.introLine1)) + "\n" + dim.Render(clamp(c.introLine2))
}

func (s *BranchesStage) renderTabs() string {
	muted := lipgloss.NewStyle().Foreground(tui.ColorMuted)
	leaf := lipgloss.NewStyle().Foreground(tui.ColorPrimary).Bold(true)
	bracket := lipgloss.NewStyle().Foreground(tui.ColorPrimary)
	chevron := lipgloss.NewStyle().Foreground(tui.ColorSecondary).Bold(true)

	const colW = 18

	cells := make([]string, 0, len(s.categories))
	for i, c := range s.categories {
		label := fmt.Sprintf("%s (%d)", c.displayName, len(s.selected[i]))
		var cell string
		if i == s.catIdx {
			cell = bracket.Render("[ ") + leaf.Render(label) + bracket.Render(" ]")
		} else {
			cell = "  " + muted.Render(label) + "  "
		}
		cells = append(cells, lipgloss.PlaceHorizontal(colW, lipgloss.Center, cell))
	}

	row := strings.Join(cells, " ")
	return chevron.Render("‹") + " " + row + " " + chevron.Render("›")
}

func (s *BranchesStage) renderList() string {
	c := s.currentCat()
	if c == nil {
		return ""
	}
	if len(c.items) == 0 {
		dim := initflow.DimStyle()
		return dim.Render("  (no items in this category)")
	}
	rows := make([]string, 0, len(c.items))
	for i := range c.items {
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

func (s *BranchesStage) renderRow(idx int) string {
	c := s.currentCat()
	if c == nil || idx < 0 || idx >= len(c.items) {
		return ""
	}
	it := c.items[idx]
	focused := idx == s.itemIdx[s.catIdx]
	selected := s.selected[s.catIdx][it.name]

	leaf := lipgloss.NewStyle().Foreground(tui.ColorPrimary)
	dim := initflow.DimStyle()

	glyph := "◇"
	glyphStyle := dim
	if selected {
		glyph = "◆"
		glyphStyle = leaf
	}

	border := initflow.UnfocusBorder()
	if focused {
		border = initflow.FocusBorder()
	}

	nameColW, descColW, tagColW := initflow.ClampColumns(s.availableWidth())
	descColW += tagColW - 2

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

	var nameStyle lipgloss.Style
	switch {
	case selected && focused:
		nameStyle = leaf.Bold(true)
	case selected:
		nameStyle = leaf
	case focused:
		nameStyle = initflow.FocusedNameStyle()
	default:
		nameStyle = initflow.UnfocusedNameStyle()
	}
	nameText := nameStyle.Render(name)
	if it.required {
		nameText += " " + initflow.RequiredGlyph()
	}
	nameCol := initflow.PadRight(nameText, nameColW)

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
			descStyle = initflow.FocusedDescStyle()
		}
		descCol = " " + descStyle.Render(initflow.PadRight(desc, descColW))
	}
	return border + glyphStyle.Render(glyph) + " " + nameCol + descCol
}

func (s *BranchesStage) renderDetails() string {
	dim := initflow.DimStyle()
	bark := initflow.LabelStyle()
	value := lipgloss.NewStyle().Foreground(tui.ColorAccent)

	header := initflow.RenderSectionHeader("DETAILS", initflow.PanelWidth(s.Width()))
	const labelW = 10
	const indent = "    "
	const aboutRows = 3
	contentW := s.Width() - 10
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

	c := s.currentCat()
	if c == nil || len(c.items) == 0 {
		empty := indent + dim.Render("(nothing to show)")
		blank := indent
		return header + "\n" + empty + "\n" + blank + "\n" + blank + "\n" + blank
	}
	row := s.itemIdx[s.catIdx]
	if row < 0 || row >= len(c.items) {
		empty := indent + dim.Render("(nothing to show)")
		blank := indent
		return header + "\n" + empty + "\n" + blank + "\n" + blank + "\n" + blank
	}
	it := c.items[row]

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

	aboutRow1 := indent + bark.Render(initflow.PadRight("ABOUT", labelW)) + value.Render(aboutLines[0])
	aboutRow2 := indent + strings.Repeat(" ", labelW) + value.Render(aboutLines[1])
	aboutRow3 := indent + strings.Repeat(" ", labelW) + value.Render(aboutLines[2])
	fileRow := indent + bark.Render(initflow.PadRight("FILE", labelW)) + value.Render(file)
	return header + "\n" + aboutRow1 + "\n" + aboutRow2 + "\n" + aboutRow3 + "\n" + fileRow
}

func (s *BranchesStage) renderCounter() string {
	total := 0
	for i := range s.categories {
		total += len(s.selected[i])
	}
	return fmt.Sprintf("%d abilities selected · across %d categories", total, len(s.categories))
}

// listHeight mirrors BranchesStage.listHeight — same chrome + non-list body
// row accounting.
func (s *BranchesStage) listHeight() int {
	if s.Height() <= 0 {
		return 0
	}
	const chromeRows = 10
	const fixedBodyRows = 21
	h := s.Height() - chromeRows - fixedBodyRows
	if h < 3 {
		h = 3
	}
	return h
}

func (s *BranchesStage) availableWidth() int {
	w := s.Width() - 4
	if w < 0 {
		return 0
	}
	return w
}

// Result returns a BranchesResult with per-category slices of selected machine
// names, preserving catalog iteration order. Required items are always
// present.
func (s *BranchesStage) Result() any {
	pick := func(idx int) []string {
		picks := s.selected[idx]
		c := s.categories[idx]
		out := make([]string, 0, len(picks))
		for _, it := range c.items {
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

// Reset clears the completion flag. Per-tab selections, cursor positions,
// and the inline-expand toggle are all preserved across Esc-back.
func (s *BranchesStage) Reset() tea.Cmd {
	s.ClearDone()
	return nil
}

// wrapToWidth word-wraps text so no line exceeds width cells. Hard-wraps
// oversized single words. Mirrors the helper in initflow/branches.go —
// duplicated so addflow has no dependency on unexported initflow helpers.
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
