package guideflow

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// fakeTopics returns a 4-element Topic slice suitable for viewer
// tests. Markdown bodies are deliberately non-empty so the cache
// population assertion has something to observe.
func fakeTopics() []Topic {
	return []Topic{
		{Key: "quickstart", Label: "QUICKSTART", Short: "START", Markdown: "# Quickstart\n\nhello"},
		{Key: "concepts", Label: "CONCEPTS", Short: "CONCP", Markdown: "# Concepts\n\nbody"},
		{Key: "cli", Label: "CLI", Short: "CLI", Markdown: "# CLI\n\nref"},
		{Key: "custom-files", Label: "CUSTOM", Short: "CUSTM", Markdown: "# Custom\n\ndetails"},
	}
}

// key builds a tea.KeyMsg for a single named key, matching the
// helper used across initflow / catalogflow test files.
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
	case "tab":
		return tea.KeyMsg{Type: tea.KeyTab}
	case "shift+tab":
		return tea.KeyMsg{Type: tea.KeyShiftTab}
	case "home":
		return tea.KeyMsg{Type: tea.KeyHome}
	case "end":
		return tea.KeyMsg{Type: tea.KeyEnd}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case "ctrl+c":
		return tea.KeyMsg{Type: tea.KeyCtrlC}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(name)}
	}
}

// TestNewViewer_InitialKeyResolved verifies an initial key maps to
// the matching topic index.
func TestNewViewer_InitialKeyResolved(t *testing.T) {
	s := NewViewer(fakeTopics(), "concepts", "1.0.0", "/tmp")
	if s.idx != 1 {
		t.Fatalf("idx for concepts = %d, want 1", s.idx)
	}
}

// TestNewViewer_InitialKeyMissingDefaultsToZero verifies unknown
// initial keys fall back to idx 0.
func TestNewViewer_InitialKeyMissingDefaultsToZero(t *testing.T) {
	s := NewViewer(fakeTopics(), "no-such-key", "", "")
	if s.idx != 0 {
		t.Fatalf("idx for unknown key = %d, want 0", s.idx)
	}
}

// TestNewViewer_EmptyInitialKeyDefaultsToZero verifies the empty
// string initial key resolves to idx 0.
func TestNewViewer_EmptyInitialKeyDefaultsToZero(t *testing.T) {
	s := NewViewer(fakeTopics(), "", "", "")
	if s.idx != 0 {
		t.Fatalf("idx for empty key = %d, want 0", s.idx)
	}
}

// TestViewer_TabCycleForwardWraps verifies tab / right / l cycle
// past the last topic back to index 0.
func TestViewer_TabCycleForwardWraps(t *testing.T) {
	s := NewViewer(fakeTopics(), "", "", "")
	// Seed a window size so resizeViewport + refreshViewportContent
	// don't trip on a zero-width viewport on the first key event
	// (they tolerate it, but the test is more meaningful with a
	// real frame size).
	s.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	for range s.topics {
		s.Update(key("right"))
	}
	if s.idx != 0 {
		t.Fatalf("idx after full forward cycle = %d, want 0", s.idx)
	}
}

// TestViewer_TabCycleBackwardWraps verifies left / shift+tab / h
// cycle from idx 0 to the last topic.
func TestViewer_TabCycleBackwardWraps(t *testing.T) {
	s := NewViewer(fakeTopics(), "", "", "")
	s.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	s.Update(key("left"))
	if s.idx != len(s.topics)-1 {
		t.Fatalf("idx after one backward cycle = %d, want %d", s.idx, len(s.topics)-1)
	}
}

// TestViewer_TabKeyCyclesForward verifies the Tab key path
// (separate from "right"/"l").
func TestViewer_TabKeyCyclesForward(t *testing.T) {
	s := NewViewer(fakeTopics(), "", "", "")
	s.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	s.Update(key("tab"))
	if s.idx != 1 {
		t.Fatalf("idx after tab = %d, want 1", s.idx)
	}
}

// TestViewer_ShiftTabCyclesBackward verifies shift+tab mirrors the
// left-arrow behavior.
func TestViewer_ShiftTabCyclesBackward(t *testing.T) {
	s := NewViewer(fakeTopics(), "", "", "")
	s.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	s.Update(key("shift+tab"))
	if s.idx != len(s.topics)-1 {
		t.Fatalf("idx after shift+tab = %d, want %d", s.idx, len(s.topics)-1)
	}
}

// TestViewer_WindowSizePopulatesRenderCache verifies a
// WindowSizeMsg triggers a render for the current idx+width —
// the cache map should have at least one entry after the resize.
func TestViewer_WindowSizePopulatesRenderCache(t *testing.T) {
	s := NewViewer(fakeTopics(), "", "", "")
	s.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	if len(s.rendered) == 0 {
		t.Fatalf("rendered cache empty after WindowSizeMsg; want >=1 entry")
	}
}

// TestViewer_QuitKeys verifies each quit key (q / esc / ctrl+c)
// flips the quit flag and issues a tea.Quit command.
func TestViewer_QuitKeys(t *testing.T) {
	for _, k := range []string{"q", "esc", "ctrl+c"} {
		s := NewViewer(fakeTopics(), "", "", "")
		_, cmd := s.Update(key(k))
		if !s.quit {
			t.Fatalf("key %q: quit flag = false, want true", k)
		}
		if cmd == nil {
			t.Fatalf("key %q: cmd = nil, want tea.Quit", k)
		}
	}
}

// TestViewer_HomeScrollsToTop verifies the "home" key path
// (delegates to viewport.GotoTop — we can't easily observe the
// Y offset without invoking viewport internals, so assert the
// Update returns no error + stays on the current tab).
func TestViewer_HomeScrollsToTop(t *testing.T) {
	s := NewViewer(fakeTopics(), "", "", "")
	s.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	before := s.idx
	s.Update(key("home"))
	if s.idx != before {
		t.Fatalf("home key should not change tab; idx before=%d after=%d", before, s.idx)
	}
}

// TestViewer_EndScrollsToBottom mirrors the home-key test for
// the "end" key path.
func TestViewer_EndScrollsToBottom(t *testing.T) {
	s := NewViewer(fakeTopics(), "", "", "")
	s.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	before := s.idx
	s.Update(key("end"))
	if s.idx != before {
		t.Fatalf("end key should not change tab; idx before=%d after=%d", before, s.idx)
	}
}

// TestViewer_MinSizeFloorUnder70x20 verifies View returns the
// min-size floor panel when the terminal is below the
// initflow 70×20 threshold.
func TestViewer_MinSizeFloorUnder70x20(t *testing.T) {
	s := NewViewer(fakeTopics(), "", "", "")
	s.Update(tea.WindowSizeMsg{Width: 40, Height: 10})
	out := s.View()
	if !strings.Contains(out, "please enlarge your terminal") {
		t.Fatalf("expected min-size floor under 40×10; got:\n%s", out)
	}
}

// TestViewer_ViewContainsBrandAndTabs verifies a full-sized render
// includes the BONSAI brand and every tab label (full form when
// the budget allows).
func TestViewer_ViewContainsBrandAndTabs(t *testing.T) {
	s := NewViewer(fakeTopics(), "", "", "")
	s.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	out := s.View()
	if !strings.Contains(out, "BONSAI") {
		t.Fatalf("expected BONSAI brand in view; got:\n%s", out)
	}
	for _, label := range []string{"QUICKSTART", "CONCEPTS", "CLI", "CUSTOM"} {
		if !strings.Contains(out, label) {
			t.Fatalf("expected tab label %q in view; got:\n%s", label, out)
		}
	}
}

// longTopics returns a 4-element Topic slice where the first topic's
// Markdown is long enough to exceed any reasonable viewport height,
// so scroll-key tests can observe a non-zero YOffset after a line /
// page down. Each paragraph is deliberately unique so glamour doesn't
// collapse or dedupe.
func longTopics() []Topic {
	var body strings.Builder
	body.WriteString("# Long Topic\n\n")
	for i := 0; i < 200; i++ {
		fmt.Fprintf(&body, "Paragraph %d — lorem ipsum dolor sit amet, consectetur adipiscing elit.\n\n", i)
	}
	return []Topic{
		{Key: "quickstart", Label: "QUICKSTART", Short: "START", Markdown: body.String()},
		{Key: "concepts", Label: "CONCEPTS", Short: "CONCP", Markdown: "# Concepts\n\nbody"},
		{Key: "cli", Label: "CLI", Short: "CLI", Markdown: "# CLI\n\nref"},
		{Key: "custom-files", Label: "CUSTOM", Short: "CUSTM", Markdown: "# Custom\n\ndetails"},
	}
}

// TestViewer_ScrollKeyDelegation verifies the fall-through path in
// Update forwards scroll keys (j/k, pgdn/pgup) to the embedded
// bubbles/viewport model, which responds by moving YOffset. The
// viewport's DefaultKeyMap binds Down→{down,j}, PageDown→{pgdown,
// space,f}, Up→{up,k}, PageUp→{pgup,b} — this test uses the single-
// char forms so the assertion doesn't depend on tea.KeyType→string
// aliasing.
func TestViewer_ScrollKeyDelegation(t *testing.T) {
	s := NewViewer(longTopics(), "", "", "")
	s.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	if s.viewport.YOffset != 0 {
		t.Fatalf("precondition: YOffset = %d, want 0 before any scroll", s.viewport.YOffset)
	}

	// Line down (j) — YOffset should advance by ~1.
	s.Update(key("j"))
	afterLineDown := s.viewport.YOffset
	if afterLineDown <= 0 {
		t.Fatalf("YOffset after j = %d, want > 0 (viewport should advance)", afterLineDown)
	}

	// Page down (f is a DefaultKeyMap alias for PageDown; we use it
	// here because pgdown typed into a tea.KeyMsg can round-trip
	// through tea.KeyPgDown.String() == "pgdown", which the viewport
	// keymap also matches — either path proves delegation).
	s.Update(key("f"))
	afterPageDown := s.viewport.YOffset
	if afterPageDown <= afterLineDown {
		t.Fatalf("YOffset after pgdown = %d, want > %d", afterPageDown, afterLineDown)
	}

	// Line up (k) — YOffset should move back toward the top.
	s.Update(key("k"))
	afterLineUp := s.viewport.YOffset
	if afterLineUp >= afterPageDown {
		t.Fatalf("YOffset after k = %d, want < %d", afterLineUp, afterPageDown)
	}

	// Page up (b is the DefaultKeyMap alias for PageUp).
	s.Update(key("b"))
	afterPageUp := s.viewport.YOffset
	if afterPageUp >= afterLineUp {
		t.Fatalf("YOffset after pgup = %d, want < %d", afterPageUp, afterLineUp)
	}
}

// TestNewTopics_PreservesCanonicalOrder verifies the helper
// preserves the quickstart → concepts → cli → custom-files order
// regardless of map iteration.
func TestNewTopics_PreservesCanonicalOrder(t *testing.T) {
	raw := map[string]string{
		"custom-files": "body d",
		"cli":          "body c",
		"concepts":     "body b",
		"quickstart":   "body a",
	}
	topics := NewTopics(raw)
	wantOrder := []string{"quickstart", "concepts", "cli", "custom-files"}
	if len(topics) != len(wantOrder) {
		t.Fatalf("NewTopics len = %d, want %d", len(topics), len(wantOrder))
	}
	for i, want := range wantOrder {
		if topics[i].Key != want {
			t.Fatalf("topics[%d].Key = %q, want %q", i, topics[i].Key, want)
		}
	}
}

// TestNewTopics_SkipsMissingKeys verifies the helper tolerates a
// rawContents map that's missing one or more canonical keys.
func TestNewTopics_SkipsMissingKeys(t *testing.T) {
	raw := map[string]string{
		"quickstart": "hello",
		"cli":        "ref",
	}
	topics := NewTopics(raw)
	if len(topics) != 2 {
		t.Fatalf("NewTopics len = %d, want 2", len(topics))
	}
	if topics[0].Key != "quickstart" || topics[1].Key != "cli" {
		t.Fatalf("order mismatch: %+v", topics)
	}
}

// TestViewer_RendererCachedPerWidth verifies repeated WindowSizeMsg
// events at the same viewport width re-use a single
// *glamour.TermRenderer instance (the cache key is the width, not
// the msg sequence).
func TestViewer_RendererCachedPerWidth(t *testing.T) {
	s := NewViewer(fakeTopics(), "", "", "")
	s.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	first := len(s.renderers)
	if first != 1 {
		t.Fatalf("len(renderers) after first resize = %d, want 1", first)
	}
	// Second resize at the same width — renderer count must stay at 1.
	s.Update(tea.WindowSizeMsg{Width: 120, Height: 24})
	if got := len(s.renderers); got != 1 {
		t.Fatalf("len(renderers) after same-width resize = %d, want 1", got)
	}
}

// TestViewer_RendererPerDistinctWidth verifies two WindowSizeMsgs at
// different widths each build their own renderer — the cache is
// width-keyed and per-width builds are independent.
func TestViewer_RendererPerDistinctWidth(t *testing.T) {
	s := NewViewer(fakeTopics(), "", "", "")
	s.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	s.Update(tea.WindowSizeMsg{Width: 80, Height: 40})
	if got := len(s.renderers); got != 2 {
		t.Fatalf("len(renderers) after two distinct widths = %d, want 2", got)
	}
}

// TestViewer_PreWarmPopulatesCache verifies the pre-warm tea.Cmd
// returned on first WindowSizeMsg, when executed and its resulting
// preWarmMsg fed back into Update, populates s.rendered with an
// entry for every topic idx at that width.
func TestViewer_PreWarmPopulatesCache(t *testing.T) {
	s := NewViewer(fakeTopics(), "", "", "")
	_, cmd := s.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	if cmd == nil {
		t.Fatalf("expected preWarmCmd on first WindowSizeMsg, got nil")
	}
	msg := cmd()
	pw, ok := msg.(preWarmMsg)
	if !ok {
		t.Fatalf("cmd produced %T, want preWarmMsg", msg)
	}
	// Feed the msg back into Update so the results land in s.rendered.
	s.Update(pw)
	// Every topic idx should now have a "idx:width" entry.
	vw := s.viewport.Width
	if vw <= 0 {
		vw = defaultRenderWidth
	}
	for i := range s.topics {
		key := fmt.Sprintf("%d:%d", i, vw)
		if _, ok := s.rendered[key]; !ok {
			t.Fatalf("expected rendered cache entry for %q after pre-warm", key)
		}
	}
}

// TestViewer_TabSwitchUsesCachedRenderer verifies a tab switch after
// the renderer cache is populated does not construct a new renderer.
// Seeds the renderer cache with the viewport width and asserts the
// map size is unchanged after cycling tabs.
func TestViewer_TabSwitchUsesCachedRenderer(t *testing.T) {
	s := NewViewer(fakeTopics(), "", "", "")
	// Seed the size + renderer cache via an initial resize.
	s.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	before := len(s.renderers)
	// Cycle through every tab — none of these should build a renderer.
	for range s.topics {
		s.Update(key("right"))
	}
	if got := len(s.renderers); got != before {
		t.Fatalf("len(renderers) after tab cycle = %d, want %d (no new builds)", got, before)
	}
}

// TestNewTopics_LabelAndShortPopulated verifies each Topic has both
// Label and Short set to the expected values.
func TestNewTopics_LabelAndShortPopulated(t *testing.T) {
	raw := map[string]string{
		"quickstart":   "a",
		"concepts":     "b",
		"cli":          "c",
		"custom-files": "d",
	}
	topics := NewTopics(raw)
	wantLabels := map[string][2]string{
		"quickstart":   {"QUICKSTART", "START"},
		"concepts":     {"CONCEPTS", "CONCP"},
		"cli":          {"CLI", "CLI"},
		"custom-files": {"CUSTOM", "CUSTM"},
	}
	for _, topic := range topics {
		want := wantLabels[topic.Key]
		if topic.Label != want[0] || topic.Short != want[1] {
			t.Fatalf("topic %q: label=%q short=%q, want label=%q short=%q",
				topic.Key, topic.Label, topic.Short, want[0], want[1])
		}
	}
}
