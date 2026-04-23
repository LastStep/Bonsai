package guideflow

import (
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
