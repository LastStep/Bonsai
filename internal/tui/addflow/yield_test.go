package addflow

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/LastStep/Bonsai/internal/catalog"
	"github.com/LastStep/Bonsai/internal/config"
	"github.com/LastStep/Bonsai/internal/tui/initflow"
)

func yieldPressKey(s *YieldStage, k tea.KeyType) {
	m, _ := s.Update(tea.KeyMsg{Type: k})
	if ys, ok := m.(*YieldStage); ok {
		*s = *ys
	}
}

func yieldPressRune(s *YieldStage, r rune) {
	m, _ := s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	if ys, ok := m.(*YieldStage); ok {
		*s = *ys
	}
}

// TestYield_SuccessRendersAgent verifies the success-mode view contains the
// agent display name.
func TestYield_SuccessRendersAgent(t *testing.T) {
	cat := &catalog.Catalog{
		Agents: []catalog.AgentDef{
			{Name: "backend", DisplayName: "Backend"},
		},
	}
	installed := &config.InstalledAgent{
		AgentType: "backend",
		Workspace: "services/api/",
		Skills:    []string{"s1", "s2"},
	}
	ctx := initflow.StageContext{StartedAt: time.Now()}
	s := NewYieldSuccess(ctx, installed, cat, true, 2)
	s.SetSize(120, 40)
	out := s.View()
	if !strings.Contains(out, "Backend") {
		t.Fatal("success view should include agent display name")
	}
	if !strings.Contains(out, "services/api/") {
		t.Fatal("success view should include workspace")
	}
}

// TestYield_AllInstalledRendersCatalogCTA verifies the all-installed mode
// renders the `bonsai catalog` next-step.
func TestYield_AllInstalledRendersCatalogCTA(t *testing.T) {
	agentDef := &catalog.AgentDef{Name: "backend", DisplayName: "Backend"}
	ctx := initflow.StageContext{StartedAt: time.Now()}
	s := NewYieldAllInstalled(ctx, agentDef)
	s.SetSize(120, 40)
	out := s.View()
	if !strings.Contains(out, "bonsai catalog") {
		t.Fatal("all-installed view should include bonsai catalog CTA")
	}
	if !strings.Contains(out, "ALREADY FULL") {
		t.Fatal("all-installed view should include ALREADY FULL hero")
	}
}

// TestYield_TechLeadRequiredRendersInitCTA verifies the tech-lead-required
// mode renders the `bonsai init` next-step.
func TestYield_TechLeadRequiredRendersInitCTA(t *testing.T) {
	ctx := initflow.StageContext{StartedAt: time.Now()}
	s := NewYieldTechLeadRequired(ctx, "backend")
	s.SetSize(120, 40)
	out := s.View()
	if !strings.Contains(out, "bonsai init") {
		t.Fatal("tech-lead-required view should include bonsai init CTA")
	}
	if !strings.Contains(out, "TECH-LEAD REQUIRED") {
		t.Fatal("tech-lead-required view should include hero")
	}
	if !strings.Contains(out, "backend") {
		t.Fatal("tech-lead-required view should include picked agent type")
	}
}

// TestYield_EnterMarksDone verifies Enter flips done on any variant.
func TestYield_EnterMarksDone(t *testing.T) {
	ctx := initflow.StageContext{StartedAt: time.Now()}
	s := NewYieldTechLeadRequired(ctx, "backend")
	if s.Done() {
		t.Fatal("should not be done before Enter")
	}
	yieldPressKey(s, tea.KeyEnter)
	if !s.Done() {
		t.Fatal("Enter should MarkDone")
	}
}

// TestYield_QKeyMarksDone verifies q also exits the terminal card.
func TestYield_QKeyMarksDone(t *testing.T) {
	ctx := initflow.StageContext{StartedAt: time.Now()}
	s := NewYieldAllInstalled(ctx, &catalog.AgentDef{Name: "backend"})
	yieldPressRune(s, 'q')
	if !s.Done() {
		t.Fatal("q should MarkDone")
	}
}

// TestYield_EscMarksDone verifies esc also exits the terminal card.
func TestYield_EscMarksDone(t *testing.T) {
	ctx := initflow.StageContext{StartedAt: time.Now()}
	s := NewYieldSuccess(ctx, &config.InstalledAgent{AgentType: "backend", Workspace: "backend/"}, nil, true, 0)
	yieldPressKey(s, tea.KeyEsc)
	if !s.Done() {
		t.Fatal("esc should MarkDone")
	}
}

// TestYield_ResultNil verifies Yield returns nil (terminal — no payload).
func TestYield_ResultNil(t *testing.T) {
	ctx := initflow.StageContext{StartedAt: time.Now()}
	s := NewYieldSuccess(ctx, nil, nil, false, 0)
	if s.Result() != nil {
		t.Fatalf("Result = %v, want nil", s.Result())
	}
}

// TestYield_UnknownAgentRendersUpdateCTA verifies the unknown-agent variant
// surfaces a `bonsai update` next-step (catalog stale guidance) and includes
// the picked agent type in the body so the user can correlate the failure
// with their config entry.
func TestYield_UnknownAgentRendersUpdateCTA(t *testing.T) {
	ctx := initflow.StageContext{StartedAt: time.Now()}
	s := NewYieldUnknownAgent(ctx, "frontend")
	s.SetSize(120, 40)
	out := s.View()
	if !strings.Contains(out, "UNKNOWN AGENT") {
		t.Fatal("unknown-agent view should include UNKNOWN AGENT hero")
	}
	if !strings.Contains(out, "bonsai update") {
		t.Fatal("unknown-agent view should direct user to run bonsai update")
	}
	if !strings.Contains(out, "frontend") {
		t.Fatal("unknown-agent view should include the picked agent type")
	}
}

// TestYield_SuccessAddItemsMessaging verifies the isNewAgent=false branch
// renders the "Grafted N abilities" wording.
func TestYield_SuccessAddItemsMessaging(t *testing.T) {
	cat := &catalog.Catalog{
		Agents: []catalog.AgentDef{
			{Name: "backend", DisplayName: "Backend"},
		},
	}
	installed := &config.InstalledAgent{
		AgentType: "backend",
		Workspace: "services/api/",
	}
	ctx := initflow.StageContext{StartedAt: time.Now()}
	s := NewYieldSuccess(ctx, installed, cat, false, 3)
	s.SetSize(120, 40)
	out := s.View()
	if !strings.Contains(out, "Grafted 3") {
		t.Fatal("add-items success view should include 'Grafted 3'")
	}
}
