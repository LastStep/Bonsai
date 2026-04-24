package catalog

import (
	"os"
	"testing"
)

func TestDisplayNameFrom(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		// Plain kebab — existing behavior preserved.
		{"scope-guard-files", "Scope Guard Files"},
		{"memory", "Memory"},
		{"coding-standards", "Coding Standards"},
		{"design-guide", "Design Guide"},
		{"status-bar", "Status Bar"},

		// Acronyms.
		{"api-design-standards", "API Design Standards"},
		{"api-development", "API Development"},
		{"api-security-check", "API Security Check"},
		{"cli-conventions", "CLI Conventions"},
		{"iac-conventions", "IaC Conventions"},
		{"iac-safety-guard", "IaC Safety Guard"},
		{"pr-creation", "PR Creation"},
		{"pr-review", "PR Review"},

		// Articles lowercase in non-leading positions.
		{"issue-to-implementation", "Issue to Implementation"},
		{"scope-of-the-agent", "Scope of the Agent"},

		// Article as first token stays capitalized.
		{"the-plan", "The Plan"},

		// Edge cases.
		{"", ""},
		{"single", "Single"},
		{"api", "API"},
	}
	for _, tc := range cases {
		got := DisplayNameFrom(tc.in)
		if got != tc.want {
			t.Errorf("DisplayNameFrom(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// TestLoadAgentHints_PopulatesFromYAML — Plan 31 Phase H: the New()
// loader picks up catalog/agents/<type>/hints.yaml for every agent that
// ships a hints file. Relies on the repo's embedded catalog at ../../catalog.
func TestLoadAgentHints_PopulatesFromYAML(t *testing.T) {
	c, err := New(os.DirFS("../../catalog"))
	if err != nil {
		t.Fatalf("catalog.New: %v", err)
	}
	// Every shipped agent should have a hints map covering all 4 commands.
	wanted := []string{"tech-lead", "backend", "frontend", "fullstack", "devops", "security"}
	for _, name := range wanted {
		def := c.GetAgent(name)
		if def == nil {
			t.Errorf("agent %q not found in catalog", name)
			continue
		}
		if def.Hints == nil {
			t.Errorf("agent %q has nil Hints — hints.yaml not loaded", name)
			continue
		}
		for _, cmd := range []string{"init", "add", "remove", "update"} {
			if _, ok := def.Hints[cmd]; !ok {
				t.Errorf("agent %q missing hints for command %q", name, cmd)
			}
		}
	}
}

// TestLoadAgentHints_RenderTemplateFires — tech-lead init hints contain
// a {{ .DocsPath }} template var; verify it parses at load time by
// checking the raw string shape (Load-time template exec happens in
// tui/hints.Load).
func TestLoadAgentHints_RenderTemplateFires(t *testing.T) {
	c, err := New(os.DirFS("../../catalog"))
	if err != nil {
		t.Fatalf("catalog.New: %v", err)
	}
	def := c.GetAgent("tech-lead")
	if def == nil {
		t.Fatal("tech-lead not found")
	}
	init, ok := def.Hints["init"]
	if !ok {
		t.Fatal("tech-lead init hints missing")
	}
	// At least one entry references the DocsPath template var so Phase H's
	// contract is enforced at load time.
	anyTemplate := false
	for _, s := range init.NextWorkflow {
		if containsTemplate(s) {
			anyTemplate = true
			break
		}
	}
	if !anyTemplate {
		t.Error("tech-lead init hints.NextWorkflow missing {{ .DocsPath }} template var")
	}
}

func containsTemplate(s string) bool {
	return len(s) > 0 && stringContains(s, "{{")
}

func stringContains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
