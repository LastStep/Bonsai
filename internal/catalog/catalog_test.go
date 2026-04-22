package catalog

import "testing"

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
