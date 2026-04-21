package initflow

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/LastStep/Bonsai/internal/catalog"
	"github.com/LastStep/Bonsai/internal/tui"
)

// ObserveStage is the pre-plant "one last look" stage at rail position 3.
// It composes a read-only summary of the three prior stages (Vessel, Soil,
// Branches) into a side-by-side review screen with a CANCEL / PLANT CTA.
//
// Result:
//   - true  → user confirmed PLANT — the harness proceeds to Generate.
//   - false → user cancelled — the flow exits without writes.
//
// Prior inputs are captured through harness.priorAware.SetPrior so the
// stage can render its summary against whatever the user last entered,
// even after an Esc-back edit pass upstream.
type ObserveStage struct {
	Stage

	cat      *catalog.Catalog
	agentDef *catalog.AgentDef

	// Captured via SetPrior on entry / after esc-back.
	vessel   map[string]string
	soil     []string
	branches BranchesResult

	// Confirmed / cancelled are mirror-state: `confirmed` is a flip when
	// the user chooses PLANT; `done` on the embedded Stage advances. We
	// split them so Result() can return false when the user hits `n` + ↵.
	confirmed bool

	// Button focus. 0 = CANCEL, 1 = PLANT. Default PLANT (1) so a bare ↵
	// ships the happy path. Tab / ← → toggles.
	btnFocus int
}

// NewObserveStage constructs the Observe stage at rail position 3. cat +
// agentDef are kept for future summary panels that might want DisplayName
// lookups on Soil picks; the current renderer reads only what Vessel/Soil/
// Branches produced, so these are optional today. They're still accepted
// to keep the signature symmetric with the other real-stage constructors.
func NewObserveStage(ctx StageContext, cat *catalog.Catalog, agentDef *catalog.AgentDef) *ObserveStage {
	label := StageLabels[3]
	base := NewStage(
		3,
		label,
		label.English,
		ctx.Version,
		ctx.ProjectDir,
		ctx.StationDir,
		ctx.AgentDisplay,
		ctx.StartedAt,
	)
	return &ObserveStage{
		Stage:    base,
		cat:      cat,
		agentDef: agentDef,
		btnFocus: 1, // default PLANT
	}
}

// SetPrior implements harness.priorAware so we pick up the Vessel/Soil/
// Branches results the moment the cursor lands on us. Called again after
// esc-back so edits propagate to the review summary.
func (s *ObserveStage) SetPrior(prev []any) {
	if len(prev) >= 1 {
		if m, ok := prev[0].(map[string]string); ok {
			s.vessel = m
		}
	}
	if len(prev) >= 2 {
		if sl, ok := prev[1].([]string); ok {
			s.soil = sl
		}
	}
	if len(prev) >= 3 {
		if br, ok := prev[2].(BranchesResult); ok {
			s.branches = br
		}
	}
}

// Init implements tea.Model — no cursor / cmd on entry.
func (s *ObserveStage) Init() tea.Cmd { return nil }

// Update handles button focus cycling, the y/n shortcuts, and ↵ confirm.
func (s *ObserveStage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = m.Width
		s.height = m.Height
	case tea.KeyMsg:
		switch m.String() {
		case "tab", "right", "l":
			s.btnFocus = (s.btnFocus + 1) % 2
		case "shift+tab", "left", "h":
			s.btnFocus = (s.btnFocus + 1) % 2
		case "y", "Y":
			s.confirmed = true
			s.done = true
			return s, nil
		case "n", "N":
			s.confirmed = false
			s.done = true
			return s, nil
		case "enter":
			s.confirmed = s.btnFocus == 1
			s.done = true
			return s, nil
		}
	}
	return s, nil
}

// View composes the Observe stage body inside the shared frame.
func (s *ObserveStage) View() string {
	return s.renderFrame(s.renderBody(), s.keyHints())
}

// keyHints builds the footer key row for this stage.
func (s *ObserveStage) keyHints() []KeyHint {
	return []KeyHint{
		{Key: "↵", Desc: "confirm"},
		{Key: "tab", Desc: "toggle"},
		{Key: "y/n", Desc: "plant / cancel"},
		{Key: "esc", Desc: "back"},
		{Key: "ctrl-c", Desc: "quit"},
	}
}

// renderBody renders the stage intro + VESSEL/SOIL/BRANCHES summary +
// the CANCEL / PLANT CTA. Layout is a responsive grid: two columns
// (Vessel+Soil left, Branches right) on widths ≥100; stacked single
// column below that.
func (s *ObserveStage) renderBody() string {
	dim := lipgloss.NewStyle().Foreground(tui.ColorRule2)
	bark := lipgloss.NewStyle().Foreground(tui.ColorSecondary).Bold(true)
	white := lipgloss.NewStyle().Foreground(tui.ColorAccent).Bold(true)

	var title string
	if s.ensoSafe {
		title = bark.Render(s.label.Kanji) + " " + white.Render(s.label.English)
	} else {
		title = white.Render(s.label.English)
	}

	intro := strings.Join([]string{
		title,
		white.Render("One last look before planting."),
		dim.Render("Review your picks — ↵ plants, n cancels."),
	}, "\n")

	vesselBlock := s.renderVesselSummary()
	soilBlock := s.renderSoilSummary()
	branchesBlock := s.renderBranchesSummary()

	var grid string
	if s.width >= 100 {
		left := vesselBlock + "\n\n" + soilBlock
		grid = joinHoriz(left, branchesBlock, 4)
	} else {
		grid = vesselBlock + "\n\n" + branchesBlock + "\n\n" + soilBlock
	}

	cta := s.renderCTA()

	body := strings.Join([]string{
		intro,
		"",
		"",
		grid,
		"",
		"",
		cta,
	}, "\n")

	return centerBlock(body, s.width)
}

// renderVesselSummary renders the NAME / DESCRIPTION / STATION / AGENT
// rows sourced from prev[0] (VesselStage Result) + the stage's agentDisplay.
func (s *ObserveStage) renderVesselSummary() string {
	dim := lipgloss.NewStyle().Foreground(tui.ColorRule2)
	bark := lipgloss.NewStyle().Foreground(tui.ColorSecondary).Bold(true)
	leaf := lipgloss.NewStyle().Foreground(tui.ColorPrimary)
	value := lipgloss.NewStyle().Foreground(tui.ColorAccent)

	header := leaf.Render(strings.Repeat("─", 3)) + " " +
		bark.Render("VESSEL") + " "
	if s.ensoSafe {
		header += bark.Render("器 ")
	}
	header += dim.Render(strings.Repeat("─", 20))

	name := s.vessel["name"]
	if name == "" {
		name = "—"
	}
	desc := s.vessel["description"]
	if desc == "" {
		desc = dim.Render("(none)")
	} else {
		desc = value.Render(desc)
	}
	station := s.vessel["station"]
	if station == "" {
		station = defaultStationDir
	}
	agent := s.agentDisplay
	if agent == "" {
		agent = "—"
	}

	const labelW = 12
	rows := []string{
		header,
		bark.Render(padRight("NAME", labelW)) + value.Render(name),
		bark.Render(padRight("DESCRIPTION", labelW)) + desc,
		bark.Render(padRight("STATION", labelW)) + value.Render(station),
		bark.Render(padRight("AGENT", labelW)) + value.Render(agent),
	}
	return strings.Join(rows, "\n")
}

// renderSoilSummary renders the scaffolding picks from prev[1] as a tiny
// file-tree via tui.RenderFileTree. Preview-only — no on-disk check. Items
// are rendered NodeNew so the user sees which entries will be written.
func (s *ObserveStage) renderSoilSummary() string {
	dim := lipgloss.NewStyle().Foreground(tui.ColorRule2)
	bark := lipgloss.NewStyle().Foreground(tui.ColorSecondary).Bold(true)
	leaf := lipgloss.NewStyle().Foreground(tui.ColorPrimary)

	header := leaf.Render(strings.Repeat("─", 3)) + " " +
		bark.Render("SOIL") + " "
	if s.ensoSafe {
		header += bark.Render("土 ")
	}
	header += dim.Render(strings.Repeat("─", 20))

	if len(s.soil) == 0 {
		return header + "\n  " + dim.Render("(no scaffolding picks)")
	}

	// Render a flat list of picks inline — the full file tree blossoms
	// after Generate; here we just want to confirm what the user ticked.
	lines := []string{header}
	for _, name := range s.soil {
		lines = append(lines, "  "+leaf.Render("◆")+" "+name)
	}
	return strings.Join(lines, "\n")
}

// renderBranchesSummary groups prev[2] (BranchesResult) into five
// kanji-labelled rows. Items are joined with "·" and wrapped via the
// existing wrapToWidth helper to respect the terminal width.
func (s *ObserveStage) renderBranchesSummary() string {
	dim := lipgloss.NewStyle().Foreground(tui.ColorRule2)
	bark := lipgloss.NewStyle().Foreground(tui.ColorSecondary).Bold(true)
	leaf := lipgloss.NewStyle().Foreground(tui.ColorPrimary)
	value := lipgloss.NewStyle().Foreground(tui.ColorAccent)

	total := len(s.branches.Skills) + len(s.branches.Workflows) +
		len(s.branches.Protocols) + len(s.branches.Sensors) + len(s.branches.Routines)

	header := leaf.Render(strings.Repeat("─", 3)) + " " +
		bark.Render("BRANCHES") + " "
	if s.ensoSafe {
		header += bark.Render("枝 ")
	}
	header += dim.Render(fmt.Sprintf("· %d abilities ", total))
	header += dim.Render(strings.Repeat("─", 10))

	// Per-group width budget: right column is ~half the terminal when
	// we're in the 2-col layout, full width when stacked. Floor at 24.
	var valueW int
	if s.width >= 100 {
		valueW = (s.width - 4) / 2
	} else {
		valueW = s.width - 18 // left label 10 + padding budget
	}
	if valueW < 24 {
		valueW = 24
	}

	groups := []struct {
		kanji, label string
		items        []string
	}{
		{"技", "SKILLS", s.branches.Skills},
		{"流", "FLOWS", s.branches.Workflows},
		{"律", "RULES", s.branches.Protocols},
		{"感", "SENSE", s.branches.Sensors},
		{"習", "HABIT", s.branches.Routines},
	}

	lines := []string{header}
	for _, g := range groups {
		var leftLabel string
		if s.ensoSafe {
			leftLabel = bark.Render(g.kanji+" "+g.label) + " "
		} else {
			leftLabel = bark.Render(g.label) + " "
		}
		// Compose value — joined picks or "(none)" dim placeholder.
		var rendered string
		if len(g.items) == 0 {
			rendered = dim.Render("(none)")
			lines = append(lines, "  "+leftLabel+rendered)
			continue
		}
		joined := strings.Join(g.items, " · ")
		wrapped := wrapToWidth(joined, valueW)
		for i, w := range wrapped {
			if i == 0 {
				lines = append(lines, "  "+leftLabel+value.Render(w))
			} else {
				// Indent continuation rows so they align under the value col.
				indent := strings.Repeat(" ", 2+lipgloss.Width(leftLabel))
				lines = append(lines, indent+value.Render(w))
			}
		}
	}
	return strings.Join(lines, "\n")
}

// renderCTA renders the confirm banner — one row with file count +
// conflict count + the CANCEL / PLANT buttons. File count is an
// upper-bound approximation (soil picks + branches picks + "…" suffix)
// because the exact count only lands after Generate runs. Conflict
// count is always 0 pre-generate — we've not checked the disk yet.
func (s *ObserveStage) renderCTA() string {
	dim := lipgloss.NewStyle().Foreground(tui.ColorRule2)
	bark := lipgloss.NewStyle().Foreground(tui.ColorSecondary).Bold(true)
	leaf := lipgloss.NewStyle().Foreground(tui.ColorPrimary)
	muted := lipgloss.NewStyle().Foreground(tui.ColorMuted)
	accent := lipgloss.NewStyle().Foreground(tui.ColorAccent).Bold(true)

	name := s.vessel["name"]
	if name == "" {
		name = "project"
	}
	approx := len(s.soil) + len(s.branches.Skills) + len(s.branches.Workflows) +
		len(s.branches.Protocols) + len(s.branches.Sensors) + len(s.branches.Routines)

	top := fmt.Sprintf("Plant ~%d files into ", approx)
	prompt := dim.Render("Existing files will be offered for merge · nothing overwritten without your say-so")

	cancelLabel := "[ CANCEL ]"
	plantLabel := "[ ⏎  PLANT ]"
	if !s.ensoSafe {
		plantLabel = "[ Enter  PLANT ]"
	}

	var cancelBtn, plantBtn string
	if s.btnFocus == 0 {
		cancelBtn = accent.Render(cancelLabel)
		plantBtn = muted.Render(plantLabel)
	} else {
		cancelBtn = muted.Render(cancelLabel)
		plantBtn = leaf.Bold(true).Render(plantLabel)
	}

	line1 := muted.Render(top) + bark.Render(name) + muted.Render(fmt.Sprintf("?  %d conflicts", 0))
	line2 := prompt
	line3 := cancelBtn + "   " + plantBtn
	return strings.Join([]string{line1, line2, "", line3}, "\n")
}

// Result returns a bool: true = proceed to Generate, false = cancel.
func (s *ObserveStage) Result() any { return s.confirmed }

// Reset clears the completion flag so re-entry behaves correctly but keeps
// the captured prior snapshot + button-focus choice.
func (s *ObserveStage) Reset() tea.Cmd {
	s.done = false
	s.confirmed = false
	return nil
}

// joinHoriz joins two multi-line blocks side-by-side with gap spaces of
// horizontal padding. Each line of `left` is padded to its own max width
// before the gap + corresponding `right` line is appended. Missing right
// lines render as blanks; missing left lines get a width-matching space
// prefix so the right column still lines up.
func joinHoriz(left, right string, gap int) string {
	ls := strings.Split(left, "\n")
	rs := strings.Split(right, "\n")

	maxLW := 0
	for _, l := range ls {
		if w := lipgloss.Width(l); w > maxLW {
			maxLW = w
		}
	}

	n := len(ls)
	if len(rs) > n {
		n = len(rs)
	}
	gapStr := strings.Repeat(" ", gap)
	out := make([]string, n)
	for i := 0; i < n; i++ {
		var l, r string
		if i < len(ls) {
			l = ls[i]
		}
		if i < len(rs) {
			r = rs[i]
		}
		lPad := maxLW - lipgloss.Width(l)
		if lPad < 0 {
			lPad = 0
		}
		out[i] = l + strings.Repeat(" ", lPad) + gapStr + r
	}
	return strings.Join(out, "\n")
}
