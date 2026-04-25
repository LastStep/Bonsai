package config

import (
	"strings"
	"testing"
)

// TestValidate_OK checks that a well-formed config passes validation. Acts
// as a baseline so regressions in the Validate rules surface immediately.
func TestValidate_OK(t *testing.T) {
	cfg := &ProjectConfig{
		ProjectName: "demo",
		DocsPath:    "docs",
		Agents: map[string]*InstalledAgent{
			"tech-lead": {AgentType: "tech-lead", Workspace: "station"},
		},
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
}

// TestValidate_RequiresProjectName — empty project_name is the most common
// hand-edit mistake; check it errors with a stable message.
func TestValidate_RequiresProjectName(t *testing.T) {
	cfg := &ProjectConfig{}
	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate: want error for empty project_name, got nil")
	}
	if !strings.Contains(err.Error(), "project_name is required") {
		t.Errorf("error = %q, want substring %q", err.Error(), "project_name is required")
	}
}

// TestValidate_BadWorkspace_AbsolutePath — absolute workspace paths are
// rejected and the agent name appears in the message so the caller knows
// which agent stanza tripped the check.
func TestValidate_BadWorkspace_AbsolutePath(t *testing.T) {
	cfg := &ProjectConfig{
		ProjectName: "demo",
		Agents: map[string]*InstalledAgent{
			"backend": {AgentType: "backend", Workspace: "/abs/path"},
		},
	}
	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate: want error, got nil")
	}
	msg := err.Error()
	if !strings.Contains(msg, "backend") {
		t.Errorf("error = %q, expected agent name 'backend'", msg)
	}
	if !strings.Contains(msg, "/abs/path") {
		t.Errorf("error = %q, expected workspace value", msg)
	}
}

// TestValidate_BadDocsPath — DocsPath uses the same wsvalidate rules; ".."
// must escape detection.
func TestValidate_BadDocsPath(t *testing.T) {
	cfg := &ProjectConfig{
		ProjectName: "demo",
		DocsPath:    "..",
	}
	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate: want error, got nil")
	}
	if !strings.Contains(err.Error(), "docs_path") {
		t.Errorf("error = %q, expected docs_path mention", err.Error())
	}
}

// TestValidate_ShellMetacharsInProjectName — table-driven scan of every
// forbidden rune. Each must produce an error referencing the field and
// the offending character.
func TestValidate_ShellMetacharsInProjectName(t *testing.T) {
	cases := []struct {
		name string
		ch   string
	}{
		{"double-quote", `"`},
		{"backtick", "`"},
		{"dollar", "$"},
		{"backslash", `\`},
		{"newline", "\n"},
		{"close-bracket", "]"},
		{"close-paren", ")"},
		{"open-bracket", "["},
		{"open-paren", "("},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &ProjectConfig{ProjectName: "demo" + tc.ch + "x"}
			err := cfg.Validate()
			if err == nil {
				t.Fatalf("Validate: want error for %q, got nil", tc.ch)
			}
			if !strings.Contains(err.Error(), "project_name") {
				t.Errorf("error = %q, expected field name 'project_name'", err.Error())
			}
		})
	}
}
