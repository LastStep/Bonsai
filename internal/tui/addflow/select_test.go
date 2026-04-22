package addflow

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/LastStep/Bonsai/internal/catalog"
	"github.com/LastStep/Bonsai/internal/tui/initflow"
)

// newTestSelect constructs a SelectStage with a three-option fixture. The
// second option is pre-marked installed so tests can assert the suffix
// rendering.
func newTestSelect() *SelectStage {
	opts := []AgentOption{
		{Name: "tech-lead", DisplayName: "Tech Lead", Description: "orchestrator", Installed: false},
		{Name: "backend", DisplayName: "Backend", Description: "server-side", Installed: true},
		{Name: "frontend", DisplayName: "Frontend", Description: "ui layer", Installed: false},
	}
	return NewSelectStage(initflow.StageContext{
		Version:      "test",
		ProjectDir:   "/tmp/select-test",
		StationDir:   "station/",
		AgentDisplay: "",
		StartedAt:    time.Now(),
	}, opts)
}

func selectPressKey(s *SelectStage, k tea.KeyType) {
	m, _ := s.Update(tea.KeyMsg{Type: k})
	if ss, ok := m.(*SelectStage); ok {
		*s = *ss
	}
}

func selectPressRune(s *SelectStage, r rune) {
	m, _ := s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	if ss, ok := m.(*SelectStage); ok {
		*s = *ss
	}
}

// TestSelect_FocusDownUp verifies ↑↓ clamps at the ends (no wrap).
func TestSelect_FocusDownUp(t *testing.T) {
	s := newTestSelect()
	if s.focus != 0 {
		t.Fatalf("initial focus = %d, want 0", s.focus)
	}
	selectPressKey(s, tea.KeyDown)
	selectPressKey(s, tea.KeyDown)
	if s.focus != 2 {
		t.Fatalf("after 2× down focus = %d, want 2", s.focus)
	}
	// Clamp at bottom.
	selectPressKey(s, tea.KeyDown)
	if s.focus != 2 {
		t.Fatalf("after extra down focus = %d, want clamp at 2", s.focus)
	}
	selectPressKey(s, tea.KeyUp)
	selectPressKey(s, tea.KeyUp)
	selectPressKey(s, tea.KeyUp)
	if s.focus != 0 {
		t.Fatalf("after 3× up focus = %d, want clamp at 0", s.focus)
	}
}

// TestSelect_EnterMarksDone verifies Enter flips done.
func TestSelect_EnterMarksDone(t *testing.T) {
	s := newTestSelect()
	if s.Done() {
		t.Fatal("should not be done before Enter")
	}
	selectPressKey(s, tea.KeyEnter)
	if !s.Done() {
		t.Fatal("should be done after Enter")
	}
}

// TestSelect_ResultReturnsMachineName verifies Result returns the focused
// option's machine Name (not DisplayName).
func TestSelect_ResultReturnsMachineName(t *testing.T) {
	s := newTestSelect()
	selectPressKey(s, tea.KeyDown)
	got, ok := s.Result().(string)
	if !ok {
		t.Fatalf("Result type = %T, want string", s.Result())
	}
	if got != "backend" {
		t.Fatalf("Result = %q, want %q", got, "backend")
	}
}

// TestSelect_ResetPreservesFocus verifies Reset clears done but not focus.
func TestSelect_ResetPreservesFocus(t *testing.T) {
	s := newTestSelect()
	selectPressKey(s, tea.KeyDown)
	selectPressKey(s, tea.KeyEnter)
	s.Reset()
	if s.Done() {
		t.Fatal("Reset should clear done")
	}
	if s.focus != 1 {
		t.Fatalf("Reset should preserve focus — got %d, want 1", s.focus)
	}
}

// TestSelect_ViewRendersInstalledBadge verifies the "(installed)" suffix is
// rendered for options flagged as installed.
func TestSelect_ViewRendersInstalledBadge(t *testing.T) {
	s := newTestSelect()
	s.SetSize(120, 40)
	selectPressKey(s, tea.KeyUp) // harmless; triggers dims captured
	out := s.View()
	if !strings.Contains(out, "(installed)") {
		t.Fatal("expected (installed) badge in rendered view")
	}
}

// TestSelect_JKBindings verifies j/k map to down/up.
func TestSelect_JKBindings(t *testing.T) {
	s := newTestSelect()
	selectPressRune(s, 'j')
	if s.focus != 1 {
		t.Fatalf("j → focus = %d, want 1", s.focus)
	}
	selectPressRune(s, 'k')
	if s.focus != 0 {
		t.Fatalf("k → focus = %d, want 0", s.focus)
	}
}

// TestSelect_EmptyOptions verifies no-panic on empty option list.
func TestSelect_EmptyOptions(t *testing.T) {
	s := NewSelectStage(initflow.StageContext{StartedAt: time.Now()}, nil)
	selectPressKey(s, tea.KeyDown)
	selectPressKey(s, tea.KeyUp)
	got := s.Result()
	if str, _ := got.(string); str != "" {
		t.Fatalf("empty options Result = %q, want empty", str)
	}
}

// TestBuildAgentOptions_FlagsInstalled verifies BuildAgentOptions marks the
// Installed bool on agents present in the installed set.
func TestBuildAgentOptions_FlagsInstalled(t *testing.T) {
	cat := &catalog.Catalog{
		Agents: []catalog.AgentDef{
			{Name: "tech-lead", DisplayName: "Tech Lead", Description: "a"},
			{Name: "backend", DisplayName: "Backend", Description: "b"},
		},
	}
	opts := BuildAgentOptions(cat, map[string]bool{"backend": true})
	if len(opts) != 2 {
		t.Fatalf("len(opts) = %d, want 2", len(opts))
	}
	if opts[0].Installed {
		t.Fatal("tech-lead should not be installed")
	}
	if !opts[1].Installed {
		t.Fatal("backend should be installed")
	}
	if opts[0].DisplayName != "Tech Lead" {
		t.Fatalf("opts[0].DisplayName = %q, want Tech Lead", opts[0].DisplayName)
	}
}
