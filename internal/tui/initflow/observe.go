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

	// Tree viewport — wraps the PLANTING body lines so tight terminals
	// don't bump the CTA off-screen. Mutated by renderBody() on every
	// View() call and by j/k scroll keys in Update().
	treeVP Viewport
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
		case "j", "down":
			s.treeVP.ScrollBy(1)
		case "k", "up":
			s.treeVP.ScrollBy(-1)
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

// keyHints builds the footer key row for this stage. j/k appear only when
// the PLANTING tree viewport has off-screen rows — avoids clutter on wide
// terminals where the full tree already fits.
func (s *ObserveStage) keyHints() []KeyHint {
	hints := []KeyHint{
		{Key: "↵", Desc: "confirm"},
		{Key: "tab", Desc: "toggle"},
		{Key: "y/n", Desc: "plant / cancel"},
	}
	if up, down := s.treeVP.HasMore(); up || down {
		hints = append(hints, KeyHint{Key: "j/k", Desc: "scroll tree"})
	}
	hints = append(hints,
		KeyHint{Key: "esc", Desc: "back"},
		KeyHint{Key: "ctrl-c", Desc: "quit"},
	)
	return hints
}

// renderBody renders the stage intro + PROJECT summary + a unified
// file-hierarchy PLANTING tree + the CANCEL / PLANT CTA. Everything
// lives in a single column aligned to PanelContentWidth so sections
// line up vertically regardless of terminal width — eliminates the
// pre-2026-04-22 grid jitter where wide terminals pushed Branches far
// right while narrow ones stacked them awkwardly.
func (s *ObserveStage) renderBody() string {
	bark := LabelStyle()
	white := lipgloss.NewStyle().Foreground(tui.ColorAccent).Bold(true)
	dim := DimStyle()

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

	project := s.renderProjectSummary()
	cta := s.renderCTA()
	plantingHeader, treeLines := s.renderPlantingParts()

	// Priority layout — PROJECT summary + CTA are mandatory; intro and
	// PLANTING tree degrade first when the terminal is too short to hold
	// the full frame. Two-blank separators collapse to one before sections
	// are dropped entirely. Keeping the CTA visible was the explicit
	// 2026-04-22 complaint — "on smaller window im not able to see the
	// plant options."
	bodyAvail := 10000 // unknown dims → render everything
	if s.height > 0 {
		bodyAvail = s.height - 10 // matches stage.renderFrame chrome budget
	}
	if bodyAvail < 10 {
		bodyAvail = 10
	}

	const (
		introRows   = 3
		projectRows = 5
		treeHdrRows = 1
		ctaRows     = 4
	)
	mustRows := projectRows + 1 /*blank before cta*/ + ctaRows
	slack := bodyAvail - mustRows
	if slack < 0 {
		slack = 0
	}

	// Tree allocation — header + 1 blank + at least one row.
	treeAllocated := 0
	treeRows := 0
	if slack >= treeHdrRows+1+1 {
		base := treeHdrRows + 1 + 1 // header + blank + 1 tree row
		extra := slack - base
		if extra > len(treeLines)-1 {
			extra = len(treeLines) - 1
		}
		if extra < 0 {
			extra = 0
		}
		treeRows = 1 + extra
		treeAllocated = treeHdrRows + 1 + treeRows
		slack -= treeAllocated
	}

	// Intro allocation — title block + 1 blank.
	introAllocated := 0
	if slack >= introRows+1 {
		introAllocated = introRows + 1
		slack -= introAllocated
	}

	// Extra breathing-room blanks spent on top spacers when slack remains.
	breathing := slack
	if breathing > 2 {
		breathing = 2
	}

	lines := make([]string, 0, 32)
	if introAllocated > 0 {
		lines = append(lines, intro)
		lines = append(lines, "")
		if breathing > 0 {
			lines = append(lines, "")
			breathing--
		}
	}
	lines = append(lines, project)
	if treeRows > 0 {
		lines = append(lines, "")
		if breathing > 0 {
			lines = append(lines, "")
			breathing--
		}
		lines = append(lines, plantingHeader)
		if treeRows >= len(treeLines) {
			lines = append(lines, treeLines...)
			s.treeVP.SetLines(treeLines)
			s.treeVP.SetHeight(len(treeLines))
		} else {
			s.treeVP.SetLines(treeLines)
			s.treeVP.SetHeight(treeRows)
			vlines := strings.Split(s.treeVP.View(), "\n")
			if _, down := s.treeVP.HasMore(); down && len(vlines) > 0 {
				vlines[len(vlines)-1] = DimStyle().Render("  … j/k to scroll")
			}
			lines = append(lines, vlines...)
		}
	}
	lines = append(lines, "")
	if treeRows == 0 && breathing > 0 {
		// Without the tree, give the CTA a little extra breathing room
		// so it doesn't sit flush against PROJECT.
		lines = append(lines, "")
	}
	lines = append(lines, cta)

	return centerBlock(strings.Join(lines, "\n"), s.width)
}

// renderProjectSummary renders the NAME / DESCRIPTION / STATION / AGENT
// rows sourced from prev[0] (VesselStage Result) + the stage's agentDisplay.
// Labels are bark-gold bold at a fixed column width; values leaf-green to
// emphasise the living-plant identity. The header uses the shared
// RenderSectionHeader helper so all five stages' sections align.
func (s *ObserveStage) renderProjectSummary() string {
	bark := LabelStyle()
	value := ValueStyle()
	dim := DimStyle()

	panelW := PanelWidth(s.width)
	header := RenderSectionHeader("PROJECT", panelW)

	name := s.vessel["name"]
	if name == "" {
		name = "—"
	}
	descRaw := s.vessel["description"]
	var desc string
	if descRaw == "" {
		desc = dim.Render("(none)")
	} else {
		desc = value.Render(descRaw)
	}
	station := s.vessel["station"]
	if station == "" {
		station = defaultStationDir
	}
	agent := s.agentDisplay
	if agent == "" {
		agent = "—"
	}

	const labelW = 14
	const indent = "  "
	rows := []string{
		header,
		indent + bark.Render(padRight("NAME", labelW)) + value.Render(name),
		indent + bark.Render(padRight("DESCRIPTION", labelW)) + desc,
		indent + bark.Render(padRight("STATION", labelW)) + value.Render(station),
		indent + bark.Render(padRight("AGENT", labelW)) + value.Render(agent),
	}
	return strings.Join(rows, "\n")
}

// renderPlantingParts renders the PLANTING section in two pieces: the
// section header (always rendered) and the tree body lines (viewport-
// wrappable). Splitting them lets renderBody budget the tree within the
// terminal height so the CTA below stays on screen even when the tree
// itself needs scrolling.
//
// Format of the rendered tree body:
//
//	station/
//	├── CLAUDE.md
//	├── INDEX.md         (soil pick)
//	├── Playbook/        (soil pick)
//	└── agent/
//	    ├── Core/
//	    ├── Skills/      4 items
//	    ├── Workflows/   6 items
//	    ...
//
// Scaffolding picks (soil) render as station/-level children; ability
// picks (branches) render as agent/ subdirectories with their selected-
// count on the right. Expanding per-ability filenames would overflow the
// screen on busy configs — the count keeps the preview scannable.
func (s *ObserveStage) renderPlantingParts() (header string, lines []string) {
	bark := LabelStyle()
	leaf := ValueStyle()
	dim := DimStyle()

	panelW := PanelWidth(s.width)
	header = RenderSectionHeader("PLANTING", panelW)

	// station/ always written (the root label for the whole tree). The
	// vessel Result appends a trailing "/" to the value; strip it here
	// before re-appending so the root never renders as "station//".
	station := s.vessel["station"]
	if station == "" {
		station = defaultStationDir
	}
	stationRoot := strings.TrimSuffix(station, "/") + "/"

	// Agent subtree children — fixed order, counts sourced from branches.
	type abilityGroup struct {
		label string
		n     int
	}
	abilityGroups := []abilityGroup{
		{"Core/", -1}, // always present, no count
		{"Skills/", len(s.branches.Skills)},
		{"Workflows/", len(s.branches.Workflows)},
		{"Protocols/", len(s.branches.Protocols)},
		{"Sensors/", len(s.branches.Sensors)},
		{"Routines/", len(s.branches.Routines)},
	}

	lines = make([]string, 0, 2+len(s.soil)+len(abilityGroups))
	// Root row.
	lines = append(lines, "  "+leaf.Render(stationRoot))

	// Scaffolding children (soil picks) — each gets ├── connector. agent/
	// is always the final child → └── connector regardless of soil length.
	rootIndent := "  "
	for _, label := range s.soil {
		lines = append(lines, rootIndent+dim.Render("├── ")+leaf.Render(label))
	}
	lines = append(lines, rootIndent+dim.Render("└── ")+leaf.Render("agent/"))

	// Ability subtree — indent under agent/.
	subIndent := rootIndent + "    " // 4 spaces aligning under "agent/"
	lastIdx := len(abilityGroups) - 1
	for i, g := range abilityGroups {
		connector := "├── "
		if i == lastIdx {
			connector = "└── "
		}
		row := subIndent + dim.Render(connector) + leaf.Render(g.label)
		if g.n >= 0 {
			// Right-align the count inside the panel width. Pad to the
			// panelW so SKILLS/WORKFLOWS/etc counts line up vertically.
			row = padRight(row, panelW-10) + bark.Render(fmt.Sprintf("%d items", g.n))
		}
		lines = append(lines, row)
	}

	return header, lines
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

	line1 := leaf.Render(top) + bark.Render(name) + muted.Render(fmt.Sprintf("  ·  %d conflicts", 0))
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
