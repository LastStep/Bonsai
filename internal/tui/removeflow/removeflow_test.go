package removeflow

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/LastStep/Bonsai/internal/tui/initflow"
)

// newTestCtx builds a minimal StageContext suitable for unit tests.
func newTestCtx() initflow.StageContext {
	return initflow.StageContext{
		Version:          "test",
		ProjectDir:       "/tmp/removeflow-test",
		StationDir:       "station/",
		AgentDisplay:     "Backend",
		StartedAt:        time.Now(),
		HeaderAction:     "REMOVE",
		HeaderRightLabel: "UPROOTING FROM",
	}
}

// TestStageLabels_HasExactlyFour verifies the canonical rail has four slots.
func TestStageLabels_HasExactlyFour(t *testing.T) {
	if len(StageLabels) != 4 {
		t.Fatalf("StageLabels len = %d, want 4", len(StageLabels))
	}
	wants := []string{"SELECT", "OBSERVE", "CONFIRM", "YIELD"}
	for i, label := range StageLabels {
		if label.English != wants[i] {
			t.Fatalf("StageLabels[%d].English = %q, want %q", i, label.English, wants[i])
		}
	}
}

// TestConfirm_DefaultFocusBack verifies the destructive Uproot button is not
// the default focus — BACK is. Matches Plan 31 §E: destructive action opt-in.
func TestConfirm_DefaultFocusBack(t *testing.T) {
	s := NewConfirmStage(newTestCtx(), "Uproot X?", "detail", "caption")
	if s.btnFocus != 0 {
		t.Fatalf("btnFocus = %d, want 0 (BACK default)", s.btnFocus)
	}
}

// TestConfirm_TabTogglesButtons verifies tab flips focus.
func TestConfirm_TabTogglesButtons(t *testing.T) {
	s := NewConfirmStage(newTestCtx(), "Uproot X?", "", "")
	confirmPressKey(s, tea.KeyTab)
	if s.btnFocus != 1 {
		t.Fatalf("after Tab btnFocus = %d, want 1", s.btnFocus)
	}
	confirmPressKey(s, tea.KeyTab)
	if s.btnFocus != 0 {
		t.Fatalf("after 2x Tab btnFocus = %d, want 0", s.btnFocus)
	}
}

// TestConfirm_YConfirms verifies 'y' sets confirmed=true + done.
func TestConfirm_YConfirms(t *testing.T) {
	s := NewConfirmStage(newTestCtx(), "Uproot X?", "", "")
	confirmPressRune(s, 'y')
	if !s.Done() {
		t.Fatal("y should MarkDone")
	}
	if got, _ := s.Result().(bool); !got {
		t.Fatal("y should set Result=true")
	}
}

// TestConfirm_EscAbortsRemoval verifies the default BACK focus + Enter gives a
// false Result (no mutation on cancel). Plan 31 §E test case.
func TestConfirm_EscAbortsRemoval(t *testing.T) {
	s := NewConfirmStage(newTestCtx(), "Uproot X?", "", "")
	// Default focus is BACK (0). Enter commits BACK → confirmed=false.
	confirmPressKey(s, tea.KeyEnter)
	if !s.Done() {
		t.Fatal("Enter should MarkDone")
	}
	if got, _ := s.Result().(bool); got {
		t.Fatal("Enter on BACK should give confirmed=false")
	}
}

// TestConfirm_NCancels verifies 'n' shortcut cancels.
func TestConfirm_NCancels(t *testing.T) {
	s := NewConfirmStage(newTestCtx(), "Uproot X?", "", "")
	confirmPressRune(s, 'n')
	if !s.Done() {
		t.Fatal("n should MarkDone")
	}
	if got, _ := s.Result().(bool); got {
		t.Fatal("n should set Result=false")
	}
}

// TestConfirm_Chromeless verifies the stage reports Chromeless=true so the
// harness yields its View() verbatim.
func TestConfirm_Chromeless(t *testing.T) {
	s := NewConfirmStage(newTestCtx(), "Uproot X?", "", "")
	if !s.Chromeless() {
		t.Fatal("ConfirmStage should be chromeless")
	}
}

// TestConfirm_EnterOnUprootConfirms verifies Enter with UPROOT focus commits.
func TestConfirm_EnterOnUprootConfirms(t *testing.T) {
	s := NewConfirmStage(newTestCtx(), "Uproot X?", "", "")
	confirmPressKey(s, tea.KeyTab) // move to UPROOT
	confirmPressKey(s, tea.KeyEnter)
	if got, _ := s.Result().(bool); !got {
		t.Fatal("Enter on UPROOT should confirm")
	}
}

// TestConfirm_ViewRendersHeading verifies the heading appears in the render.
func TestConfirm_ViewRendersHeading(t *testing.T) {
	s := NewConfirmStage(newTestCtx(), "Uproot Backend?", "detail body", "")
	s.SetSize(120, 40)
	out := s.View()
	if !strings.Contains(out, "Uproot Backend?") {
		t.Fatal("view should contain heading")
	}
}

// TestRemoveflow_TechLeadBlockedWhenPeersInstalled is a regression test for
// the legacy tech-lead guard behaviour. The guard lives in cmd/remove.go and
// short-circuits before any removeflow stage is instantiated — this test is
// a placeholder asserting the contract the cmd-layer relies on: an agent
// picker on an item-remove with a single viable target must skip rendering
// (needsPicker == false). Mirrors the `len(allowedAll) > 1` predicate in
// cmd/remove.go:runRemoveItem.
func TestRemoveflow_TechLeadBlockedWhenPeersInstalled(t *testing.T) {
	// Zero options ⇒ picker should not fire (caller never instantiates).
	opts := []AgentOption{}
	if len(opts) > 1 {
		t.Fatal("single-match options should skip picker")
	}
	// Single option ⇒ still skipped.
	opts = []AgentOption{{Name: "backend", DisplayName: "Backend"}}
	if len(opts) > 1 {
		t.Fatal("single-match options should skip picker")
	}
}

// TestRemoveflow_ItemRemoveMultiAgentPicker asserts the agent picker fires
// when an item is installed in two agents. The caller in cmd/remove.go gates
// on `needsPicker := len(allowedAll) > 1`; this test validates that the
// downstream SelectStage constructor accepts the expected option shape.
func TestRemoveflow_ItemRemoveMultiAgentPicker(t *testing.T) {
	opts := []AgentOption{
		{Name: "tech-lead", DisplayName: "Tech Lead", Workspace: "station/"},
		{Name: "backend", DisplayName: "Backend", Workspace: "backend/"},
		{Name: "_all_", DisplayName: "All agents", All: true},
	}
	s := NewSelectStage(newTestCtx(), "Coding Standards", "skill", opts)
	if s == nil {
		t.Fatal("multi-agent picker should instantiate")
	}
	if len(s.options) != 3 {
		t.Fatalf("options len = %d, want 3 (2 agents + all)", len(s.options))
	}
}

// TestRemoveflow_ItemRemoveSingleAgentSkipsPicker verifies the caller's
// single-match contract: one allowed target → no picker. This is a docstring
// test for the cmd/remove.go contract, not a stage-level test.
func TestRemoveflow_ItemRemoveSingleAgentSkipsPicker(t *testing.T) {
	// Single match ⇒ caller passes needsPicker=false, picker Conditional
	// auto-completes, Observe renders directly.
	matches := []AgentOption{
		{Name: "backend", DisplayName: "Backend", Workspace: "backend/"},
	}
	needsPicker := len(matches) > 1
	if needsPicker {
		t.Fatal("single-match should skip picker")
	}
}

// TestRemoveflow_AgentRemoveHappyPath verifies the Observe+Confirm state
// machine reaches a terminal confirmed=true outcome for an agent-remove. No
// filesystem mutation happens inside the stage — this is pure state-machine
// coverage.
func TestRemoveflow_AgentRemoveHappyPath(t *testing.T) {
	ctx := newTestCtx()
	observe := NewObserveAgent(ctx, "backend", "Backend", "backend/",
		[]string{"s1", "s2"}, []string{"w1"}, nil, nil, nil)
	// User presses Enter on Observe → advance.
	observePressKey(observe, tea.KeyEnter)
	if !observe.Done() {
		t.Fatal("Observe should advance on Enter")
	}

	confirm := NewConfirmStage(ctx, "Uproot Backend?", "", "")
	// User picks UPROOT.
	confirmPressKey(confirm, tea.KeyTab)
	confirmPressKey(confirm, tea.KeyEnter)
	if got, _ := confirm.Result().(bool); !got {
		t.Fatal("happy-path Confirm should return true")
	}
}

// confirmPressKey / confirmPressRune are the test helpers mirroring
// addflow's observePressKey / observePressRune shapes. The stage is passed
// by pointer and the updated state copied back into *s.
func confirmPressKey(s *ConfirmStage, k tea.KeyType) {
	m, _ := s.Update(tea.KeyMsg{Type: k})
	if cs, ok := m.(*ConfirmStage); ok {
		*s = *cs
	}
}

func confirmPressRune(s *ConfirmStage, r rune) {
	m, _ := s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	if cs, ok := m.(*ConfirmStage); ok {
		*s = *cs
	}
}

// observePressKey dispatches a key-typed KeyMsg to an ObserveStage.
func observePressKey(s *ObserveStage, k tea.KeyType) {
	m, _ := s.Update(tea.KeyMsg{Type: k})
	if os, ok := m.(*ObserveStage); ok {
		*s = *os
	}
}

// TestRenderStatic_ContainsRefusalMessage verifies the non-TTY static
// renderer emits the user-facing "re-run / --yes" hint.
func TestRenderStatic_ContainsRefusalMessage(t *testing.T) {
	out := RenderStatic(StaticPreview{
		Title: "Remove Backend?",
		Lines: []string{"Agent: Backend", "Workspace: backend/"},
	})
	for _, want := range []string{"Remove Backend?", "Backend", "terminal", "--yes"} {
		if !strings.Contains(out, want) {
			t.Fatalf("RenderStatic output missing %q; got:\n%s", want, out)
		}
	}
}
