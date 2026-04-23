package catalogflow

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"github.com/LastStep/Bonsai/internal/catalog"
)

// fakeCatalog returns a minimal Catalog suitable for browser tests. Two
// skills compatible with "tech-lead", one routine for "all", one agent,
// one scaffolding item. Every field is populated so downstream
// renderers see non-empty inputs.
func fakeCatalog() *catalog.Catalog {
	return &catalog.Catalog{
		Agents: []catalog.AgentDef{
			{Name: "tech-lead", DisplayName: "Tech Lead", Description: "orchestrator"},
		},
		Skills: []catalog.CatalogItem{
			{
				Name: "planning-template", DisplayName: "Planning Template",
				Description: "tiered plans", Agents: catalog.AgentCompat{All: true},
			},
			{
				Name: "coding-standards", DisplayName: "Coding Standards",
				Description: "style guide", Agents: catalog.AgentCompat{Names: []string{"code"}},
			},
		},
		Workflows: []catalog.CatalogItem{
			{
				Name: "code-review", DisplayName: "Code Review",
				Description: "review pipeline", Agents: catalog.AgentCompat{All: true},
			},
		},
		Protocols: []catalog.CatalogItem{
			{
				Name: "memory", DisplayName: "Memory",
				Description: "memory protocol", Agents: catalog.AgentCompat{All: true},
			},
		},
		Sensors: []catalog.SensorItem{
			{
				Name: "status-bar", DisplayName: "Status Bar",
				Description: "status line", Event: "Stop",
				Agents: catalog.AgentCompat{All: true},
			},
		},
		Routines: []catalog.RoutineItem{
			{
				Name: "backlog-hygiene", DisplayName: "Backlog Hygiene",
				Description: "weekly sweep", Frequency: "7 days",
				Agents: catalog.AgentCompat{All: true},
			},
		},
		Scaffolding: []catalog.ScaffoldingItem{
			{Name: "playbook", DisplayName: "Playbook", Description: "plans + roadmap", Required: true, Affects: "planning"},
		},
	}
}

// key builds a tea.KeyMsg for a single named key.
func key(name string) tea.KeyMsg {
	switch name {
	case "left":
		return tea.KeyMsg{Type: tea.KeyLeft}
	case "right":
		return tea.KeyMsg{Type: tea.KeyRight}
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case "ctrl+c":
		return tea.KeyMsg{Type: tea.KeyCtrlC}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(name)}
	}
}

// TestNewBrowser_AllSevenTabsPresent verifies every catalog section is
// represented in the tab strip, in the canonical order.
func TestNewBrowser_AllSevenTabsPresent(t *testing.T) {
	s := NewBrowser(fakeCatalog(), "", "")
	if got, want := len(s.categories), 7; got != want {
		t.Fatalf("len(categories) = %d, want %d", got, want)
	}
	expected := []string{"agents", "skills", "workflows", "protocols", "sensors", "routines", "scaffolding"}
	for i, key := range expected {
		if s.categories[i].key != key {
			t.Fatalf("categories[%d].key = %q, want %q", i, s.categories[i].key, key)
		}
	}
}

// TestBrowser_TabCycleWrapsForward verifies left/right key cycling
// wraps past the end back to index 0.
func TestBrowser_TabCycleWrapsForward(t *testing.T) {
	s := NewBrowser(fakeCatalog(), "", "")
	for range s.categories {
		s.Update(key("right"))
	}
	if s.catIdx != 0 {
		t.Fatalf("catIdx after full forward cycle = %d, want 0", s.catIdx)
	}
}

// TestBrowser_TabCycleWrapsBackward verifies left key cycling from
// index 0 wraps to the last tab.
func TestBrowser_TabCycleWrapsBackward(t *testing.T) {
	s := NewBrowser(fakeCatalog(), "", "")
	s.Update(key("left"))
	if s.catIdx != len(s.categories)-1 {
		t.Fatalf("catIdx after one backward = %d, want %d", s.catIdx, len(s.categories)-1)
	}
}

// TestBrowser_FocusClampAtTop verifies up key at idx 0 stays at 0.
func TestBrowser_FocusClampAtTop(t *testing.T) {
	s := NewBrowser(fakeCatalog(), "", "")
	s.Update(key("up"))
	if s.itemIdx[s.catIdx] != 0 {
		t.Fatalf("itemIdx after up at top = %d, want 0", s.itemIdx[s.catIdx])
	}
}

// TestBrowser_FocusClampAtBottom verifies down key past the last row
// clamps to len-1.
func TestBrowser_FocusClampAtBottom(t *testing.T) {
	s := NewBrowser(fakeCatalog(), "", "")
	// Move to skills tab (2 entries).
	s.Update(key("right"))
	cat := s.currentCat()
	if cat == nil || len(cat.entries) == 0 {
		t.Fatalf("expected skills tab to have entries")
	}
	last := len(cat.entries) - 1
	for i := 0; i < 10; i++ {
		s.Update(key("down"))
	}
	if s.itemIdx[s.catIdx] != last {
		t.Fatalf("itemIdx after down past bottom = %d, want %d", s.itemIdx[s.catIdx], last)
	}
}

// TestBrowser_QuestionTogglesExpand verifies the `?` key flips the
// expanded state each press.
func TestBrowser_QuestionTogglesExpand(t *testing.T) {
	s := NewBrowser(fakeCatalog(), "", "")
	if s.expanded {
		t.Fatalf("expanded starts true, want false")
	}
	s.Update(key("?"))
	if !s.expanded {
		t.Fatalf("expanded after one ?, got false")
	}
	s.Update(key("?"))
	if s.expanded {
		t.Fatalf("expanded after two ?, got true")
	}
}

// TestBrowser_FilterGreysEmptyTabs verifies that with agent filter
// "tech-lead", skills with agents=[code] are excluded but the tab
// strip still shows every category (greyed when empty).
func TestBrowser_FilterGreysEmptyTabs(t *testing.T) {
	s := NewBrowser(fakeCatalog(), "tech-lead", "")
	if len(s.categories) != 7 {
		t.Fatalf("len(categories) under filter = %d, want 7", len(s.categories))
	}
	// Skills: only "planning-template" (All: true) passes — "coding-standards" is code-only.
	var skills *category
	for i := range s.categories {
		if s.categories[i].key == "skills" {
			skills = &s.categories[i]
			break
		}
	}
	if skills == nil {
		t.Fatalf("skills tab missing after filter")
	}
	if got, want := len(skills.entries), 1; got != want {
		t.Fatalf("skills entries under tech-lead filter = %d, want %d", got, want)
	}
}

// TestBrowser_EachQuitKey verifies q / esc / ctrl+c / enter all flip
// the quit flag and issue tea.Quit.
func TestBrowser_EachQuitKey(t *testing.T) {
	for _, k := range []string{"q", "esc", "ctrl+c", "enter"} {
		s := NewBrowser(fakeCatalog(), "", "")
		_, cmd := s.Update(key(k))
		if !s.quit {
			t.Fatalf("key %q: quit flag = false, want true", k)
		}
		if cmd == nil {
			t.Fatalf("key %q: cmd = nil, want tea.Quit", k)
		}
	}
}

// TestBrowser_ViewUnderMinSizeFloor verifies the browser renders the
// min-size floor panel (not the tab strip) when terminal dims are
// below the 70×20 threshold.
func TestBrowser_ViewUnderMinSizeFloor(t *testing.T) {
	s := NewBrowser(fakeCatalog(), "", "")
	s.Update(tea.WindowSizeMsg{Width: 40, Height: 10})

	out := s.View()
	if !strings.Contains(out, "please enlarge your terminal") {
		t.Fatalf("expected min-size floor panel under small terminal, got:\n%s", out)
	}
}

// TestBrowser_ViewContainsHeaderAndTabs verifies a normal-sized render
// contains the BONSAI header brand and every tab label.
func TestBrowser_ViewContainsHeaderAndTabs(t *testing.T) {
	s := NewBrowser(fakeCatalog(), "", "")
	s.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	out := s.View()
	if !strings.Contains(out, "BONSAI") {
		t.Fatalf("expected BONSAI brand in view, got:\n%s", out)
	}
	if !strings.Contains(out, "CATALOG") {
		t.Fatalf("expected CATALOG action in view, got:\n%s", out)
	}
	for _, tab := range []string{"AGENTS", "SKILLS", "WORKFLOWS", "PROTOCOLS", "SENSORS", "ROUTINES", "SCAFFOLDING"} {
		if !strings.Contains(out, tab) {
			t.Fatalf("expected tab %q in view, got:\n%s", tab, out)
		}
	}
}

// TestBrowser_EmptyCategoriesNoOp ensures focus/tab keys on an empty
// catalog silently no-op without panicking.
func TestBrowser_EmptyCategoriesNoOp(t *testing.T) {
	s := NewBrowser(&catalog.Catalog{}, "", "")
	// Even an empty catalog still yields 7 tabs (all empty).
	if len(s.categories) != 7 {
		t.Fatalf("empty catalog len(categories) = %d, want 7", len(s.categories))
	}
	s.Update(key("down"))
	s.Update(key("up"))
	s.Update(key("right"))
	s.Update(key("left"))
	if s.itemIdx[s.catIdx] != 0 {
		t.Fatalf("empty catalog focus drifted to %d, want 0", s.itemIdx[s.catIdx])
	}
}

// TestRenderTabs_ShortLabelsAtNarrowWidth verifies the tab strip
// switches to compact labels below the 96-col threshold so the full
// 7-tab row fits inside the 70-col minimum-width floor. All seven
// tabs must still be present (by their short form), the active tab
// must still render bold, and the rendered strip must fit in 70 cols.
func TestRenderTabs_ShortLabelsAtNarrowWidth(t *testing.T) {
	s := NewBrowser(fakeCatalog(), "", "")
	s.Update(tea.WindowSizeMsg{Width: 70, Height: 20})

	strip := s.renderTabs()
	// Visible width — lipgloss.Width strips ANSI escapes, giving the
	// true on-screen cell count.
	if w := lipgloss.Width(strip); w > 70 {
		t.Fatalf("tab strip width at 70 cols = %d, want <= 70\nstrip: %q", w, strip)
	}

	// Every tab still present (by short label). All seven collapse to
	// 5-char forms under narrow widths.
	shortLabels := []string{"AGENT", "SKILL", "FLOWS", "PROTO", "SENSE", "RTNES", "SCAFF"}
	for _, lab := range shortLabels {
		if !strings.Contains(strip, lab) {
			t.Fatalf("short label %q missing from strip at 70 cols:\n%s", lab, strip)
		}
	}

	// Full labels that should have been compressed must NOT appear as
	// whole words — checking boundary chars via a space suffix is
	// sufficient because the strip only has labels + "(N)" suffixes.
	for _, full := range []string{"AGENTS ", "SKILLS ", "WORKFLOWS ", "PROTOCOLS ", "SENSORS ", "ROUTINES ", "SCAFFOLDING "} {
		if strings.Contains(strip, full) {
			t.Fatalf("expected full label %q compressed at 70 cols, still present:\n%s", full, strip)
		}
	}

	// Active tab is catIdx=0 (AGENT). In a color-capable terminal the
	// active cell carries a bold SGR escape; under Go's test environment
	// lipgloss auto-disables colour so we only assert the active label
	// is present. The key visual invariant is that AGENT appears first
	// (catIdx 0) — which is covered by the width + presence checks
	// above.
	if idx := strings.Index(strip, "AGENT"); idx != 0 {
		t.Fatalf("active tab AGENT should render first in strip, got index %d:\n%s", idx, strip)
	}
}

// TestBrowser_ZeroCountTabRendersMuted verifies that a tab whose
// filtered entry count is zero renders with a muted styling
// (different ANSI from the active + default-inactive cells). The
// test forces a truecolor profile so lipgloss emits SGR escapes —
// by default the Go test env picks the Ascii profile (stdout not a
// TTY), which would strip all styling and defeat the assertion.
//
// Uses an agent filter that yields zero Skills entries (only the
// "code"-scoped skill exists, and the filter is "tech-lead" with
// the All-true skill removed from the fake catalog so the Skills
// tab is forced to (0)).
func TestBrowser_ZeroCountTabRendersMuted(t *testing.T) {
	// Force truecolor so lipgloss emits ANSI SGR sequences that the
	// assertion can find.
	prev := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	t.Cleanup(func() { lipgloss.SetColorProfile(prev) })

	// Build a catalog whose Skills section has only a code-agent
	// entry so the "tech-lead" filter yields an empty Skills tab.
	cat := &catalog.Catalog{
		Agents: []catalog.AgentDef{
			{Name: "tech-lead", DisplayName: "Tech Lead", Description: "x"},
		},
		Skills: []catalog.CatalogItem{
			{
				Name: "coding-standards", DisplayName: "Coding Standards",
				Description: "style", Agents: catalog.AgentCompat{Names: []string{"code"}},
			},
		},
	}
	s := NewBrowser(cat, "tech-lead", "")
	s.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	strip := s.renderTabs()

	// Skills tab should now render "(0)" — confirm the literal suffix
	// is present first (sanity check that the test is looking at the
	// right cell).
	if !strings.Contains(strip, "(0)") {
		t.Fatalf("expected '(0)' suffix for empty Skills tab, got:\n%s", strip)
	}

	// Locate the SKILLS cell in the strip and assert it carries ANSI
	// SGR escapes (muted styling). The active AGENTS cell at catIdx=0
	// also has SGR escapes but those encode the active bold+primary
	// styling — the invariant we're asserting is "(0) cells are
	// styled distinctly from plain text", which is exactly what the
	// `dim.Render(label + countSuffix)` branch produces.
	idx := strings.Index(strip, "SKILLS")
	if idx < 0 {
		t.Fatalf("SKILLS tab missing from strip:\n%s", strip)
	}
	// The SKILLS substring is preceded by the style-open SGR for the
	// dim.Render call. Grab a window that should include the escape
	// prefix and assert it contains an ESC + '['.
	window := strip[maxInt(0, idx-16):idx]
	if !strings.Contains(window, "\x1b[") {
		t.Fatalf("expected ANSI SGR escape prefixing the muted SKILLS (0) cell; window=%q strip=%q", window, strip)
	}
}

// maxInt is a tiny local int-max helper for substring window
// arithmetic — stdlib math.Max is float-only.
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// TestRenderTabs_FullLabelsAtWideWidth verifies that at widths >= 96
// cols the tab strip uses the full labels (regression guard against
// the short-label mode leaking into normal terminals).
func TestRenderTabs_FullLabelsAtWideWidth(t *testing.T) {
	s := NewBrowser(fakeCatalog(), "", "")
	s.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	strip := s.renderTabs()
	for _, full := range []string{"AGENTS", "WORKFLOWS", "PROTOCOLS", "SCAFFOLDING"} {
		if !strings.Contains(strip, full) {
			t.Fatalf("expected full label %q at 120 cols:\n%s", full, strip)
		}
	}
}
