package initflow

import (
	tea "github.com/charmbracelet/bubbletea"
)

// StubStage is a Phase-2 placeholder that renders the persistent chrome
// (header + enso rail + footer) around a stub body, and advances on Enter.
// Each of the four stages in runInitRedesign constructs one StubStage so
// the flow can be verified end-to-end before Phase 3+ replaces them with
// real inputs.
//
// StubStage composes Stage and delegates Chromeless/Title/Done/Result to
// the embedded base. It overrides View to paint the placeholder body and
// Update to advance on Enter (Esc is consumed by the harness).
type StubStage struct {
	Stage
}

// NewStubStage constructs a Phase-2 placeholder stage at rail position idx.
// All context fields flow through Stage; behaviour is fixed (Enter advances,
// no body interaction).
func NewStubStage(
	idx int,
	version string,
	projectDir string,
	stationDir string,
	agentDisplay string,
	startedAt StageContext,
) *StubStage {
	label := StageLabels[idx]
	s := NewStage(
		idx,
		label,
		label.English,
		version,
		projectDir,
		stationDir,
		agentDisplay,
		startedAt.StartedAt,
	)
	return &StubStage{Stage: s}
}

// Init kicks off a no-op tea.Cmd. No focus/cursor to manage at this stage.
func (s *StubStage) Init() tea.Cmd { return nil }

// Update advances on Enter. Esc / Ctrl-C are consumed by the harness.
func (s *StubStage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok {
		switch k.String() {
		case "enter":
			s.done = true
			return s, nil
		}
	}
	if w, ok := msg.(tea.WindowSizeMsg); ok {
		s.width = w.Width
		s.height = w.Height
	}
	return s, nil
}

// View draws the shared chrome around a placeholder body. The string body
// is deliberately minimal — Phase 3+ replaces this with real stage content.
func (s *StubStage) View() string {
	body := "  (stage body goes here)"
	canGoBack := s.idx > 0
	return s.renderFrame(body, DefaultKeys(canGoBack))
}

// Reset ensures popping back onto a stub stage clears its completed flag
// so the next Enter re-advances (matches harness.resetter expectations).
func (s *StubStage) Reset() tea.Cmd {
	s.done = false
	return nil
}
