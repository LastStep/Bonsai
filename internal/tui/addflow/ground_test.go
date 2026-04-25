package addflow

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/LastStep/Bonsai/internal/tui/initflow"
)

func newTestGround(agentType string, existing map[string]bool) *GroundStage {
	return NewGroundStage(initflow.StageContext{
		Version:      "test",
		ProjectDir:   "/tmp/ground-test",
		StationDir:   "station/",
		AgentDisplay: "",
		StartedAt:    time.Now(),
	}, GroundContext{
		AgentType:          agentType,
		DocsPath:           "station/",
		ExistingWorkspaces: existing,
	})
}

func groundPressKey(s *GroundStage, k tea.KeyType) {
	m, _ := s.Update(tea.KeyMsg{Type: k})
	if gs, ok := m.(*GroundStage); ok {
		*s = *gs
	}
}

func groundType(s *GroundStage, text string) {
	s.input.SetValue(text)
}

// TestGround_TechLeadAutoCompletes verifies tech-lead agent flips done at
// Init and AutoComplete returns true.
func TestGround_TechLeadAutoCompletes(t *testing.T) {
	s := newTestGround("tech-lead", nil)
	s.Init()
	if !s.Done() {
		t.Fatal("tech-lead Init should MarkDone")
	}
	if !s.AutoComplete() {
		t.Fatal("tech-lead AutoComplete should return true")
	}
	got, ok := s.Result().(string)
	if !ok || got != "station/" {
		t.Fatalf("Result = %v, want station/", s.Result())
	}
}

// TestGround_TechLeadUsesDocsPath verifies the DocsPath passed via context is
// returned verbatim when non-empty.
func TestGround_TechLeadUsesDocsPath(t *testing.T) {
	s := NewGroundStage(initflow.StageContext{StartedAt: time.Now()}, GroundContext{
		AgentType: "tech-lead",
		DocsPath:  "docs/",
	})
	got, _ := s.Result().(string)
	if got != "docs/" {
		t.Fatalf("Result = %q, want docs/", got)
	}
}

// TestGround_NewAgentDoesNotAutoComplete verifies non-tech-lead paths don't
// auto-advance.
func TestGround_NewAgentDoesNotAutoComplete(t *testing.T) {
	s := newTestGround("backend", nil)
	s.Init()
	if s.Done() {
		t.Fatal("backend Init should not MarkDone")
	}
	if s.AutoComplete() {
		t.Fatal("backend AutoComplete should return false")
	}
}

// TestGround_EnterValidatesEmpty verifies empty workspace is rejected.
func TestGround_EnterValidatesEmpty(t *testing.T) {
	s := newTestGround("backend", nil)
	s.input.SetValue("")
	groundPressKey(s, tea.KeyEnter)
	if s.Done() {
		t.Fatal("empty value should not advance")
	}
	if !strings.Contains(s.validateErr, "required") {
		t.Fatalf("validateErr = %q, want 'required'", s.validateErr)
	}
}

// TestGround_EnterRejectsDuplicate verifies a workspace already in use is
// flagged.
func TestGround_EnterRejectsDuplicate(t *testing.T) {
	existing := map[string]bool{"backend/": true}
	s := newTestGround("backend", existing)
	groundType(s, "backend/")
	groundPressKey(s, tea.KeyEnter)
	if s.Done() {
		t.Fatal("duplicate should not advance")
	}
	if !strings.Contains(s.validateErr, "already in use") {
		t.Fatalf("validateErr = %q, want 'already in use'", s.validateErr)
	}
}

// TestGround_EnterAdvancesOnValid verifies valid input normalises + advances.
func TestGround_EnterAdvancesOnValid(t *testing.T) {
	s := newTestGround("backend", nil)
	groundType(s, "./services/api")
	groundPressKey(s, tea.KeyEnter)
	if !s.Done() {
		t.Fatal("valid input should advance")
	}
	got, _ := s.Result().(string)
	if got != "services/api/" {
		t.Fatalf("Result = %q, want services/api/", got)
	}
}

// TestGround_ResultNormalises verifies Result normalises whitespace/clean.
func TestGround_ResultNormalises(t *testing.T) {
	s := newTestGround("backend", nil)
	groundType(s, "  custom//  ")
	got, _ := s.Result().(string)
	if got != "custom/" {
		t.Fatalf("Result = %q, want custom/", got)
	}
}

// TestGround_ResetPreservesValue verifies Reset clears done but not the
// entered value.
func TestGround_ResetPreservesValue(t *testing.T) {
	s := newTestGround("backend", nil)
	groundType(s, "services/api/")
	groundPressKey(s, tea.KeyEnter)
	if !s.Done() {
		t.Fatal("should be done")
	}
	s.Reset()
	if s.Done() {
		t.Fatal("Reset should clear done")
	}
	if s.input.Value() != "services/api/" {
		t.Fatalf("Reset should preserve value — got %q", s.input.Value())
	}
}

// TestGround_RejectsAbsolutePath verifies a rooted workspace input is
// rejected with a user-facing error. Defence against accidental writes
// outside the project root (Plan 29 §H).
func TestGround_RejectsAbsolutePath(t *testing.T) {
	s := newTestGround("backend", nil)
	groundType(s, "/etc/foo")
	groundPressKey(s, tea.KeyEnter)
	if s.Done() {
		t.Fatal("absolute path should not advance")
	}
	if !strings.Contains(s.validateErr, "absolute paths not allowed") {
		t.Fatalf("validateErr = %q, want 'absolute paths not allowed'", s.validateErr)
	}
}

// TestGround_RejectsParentEscape verifies "../..." is rejected as escaping
// the project root.
func TestGround_RejectsParentEscape(t *testing.T) {
	s := newTestGround("backend", nil)
	groundType(s, "../foo")
	groundPressKey(s, tea.KeyEnter)
	if s.Done() {
		t.Fatal("../foo should not advance")
	}
	if !strings.Contains(s.validateErr, "escape") {
		t.Fatalf("validateErr = %q, want 'escape'", s.validateErr)
	}
}

// TestGround_RejectsHiddenParentEscape verifies a filepath.Clean-reduced
// escape ("nested/../..") is rejected after normalisation.
func TestGround_RejectsHiddenParentEscape(t *testing.T) {
	s := newTestGround("backend", nil)
	groundType(s, "nested/../..")
	groundPressKey(s, tea.KeyEnter)
	if s.Done() {
		t.Fatal("nested/../.. should not advance")
	}
	if !strings.Contains(s.validateErr, "escape") {
		t.Fatalf("validateErr = %q, want 'escape'", s.validateErr)
	}
}

// TestGround_AcceptsNestedRelative is the positive companion to the
// rejects-* tests above. Verifies clean relative inputs (including ones
// that Clean reduces to a safe nested path) advance the stage.
func TestGround_AcceptsNestedRelative(t *testing.T) {
	cases := []string{"./foo", "foo/../bar"}
	for _, in := range cases {
		s := newTestGround("backend", nil)
		groundType(s, in)
		groundPressKey(s, tea.KeyEnter)
		if !s.Done() {
			t.Errorf("input %q should advance — validateErr=%q", in, s.validateErr)
		}
	}
}
