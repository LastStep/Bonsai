package hints

import (
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/catalog"
)

// fakeCatalogWithHints constructs a Catalog with a single agent def
// carrying the given hints map — avoids embedding fs.FS for unit tests.
func fakeCatalogWithHints(agentType string, hints map[string]catalog.HintSection) *catalog.Catalog {
	c := &catalog.Catalog{
		Agents: []catalog.AgentDef{
			{
				Name:  agentType,
				Hints: hints,
			},
		},
	}
	return c
}

// TestHintsLoad_RendersTemplates — {{ .DocsPath }} substitution fires
// on every field.
func TestHintsLoad_RendersTemplates(t *testing.T) {
	cat := fakeCatalogWithHints("tech-lead", map[string]catalog.HintSection{
		"init": {
			NextCLI:      []string{"bonsai add · open {{ .DocsPath }}CLAUDE.md"},
			NextWorkflow: []string{"Edit {{ .DocsPath }}Playbook/Backlog.md"},
			AIPrompts: []catalog.HintPrompt{
				{Label: "Start", Body: "Read {{ .DocsPath }}CLAUDE.md and summarize."},
			},
		},
	})
	ctx := TemplateContext{DocsPath: "station/"}
	block, err := Load(cat, "tech-lead", "init", ctx)
	if err != nil {
		t.Fatalf("Load err: %v", err)
	}
	if len(block.NextCLI) != 1 || !strings.Contains(block.NextCLI[0], "station/CLAUDE.md") {
		t.Fatalf("NextCLI template not substituted; got %v", block.NextCLI)
	}
	if len(block.NextWorkflow) != 1 || !strings.Contains(block.NextWorkflow[0], "station/Playbook/Backlog.md") {
		t.Fatalf("NextWorkflow template not substituted; got %v", block.NextWorkflow)
	}
	if len(block.AIPrompts) != 1 || !strings.Contains(block.AIPrompts[0].Body, "station/CLAUDE.md") {
		t.Fatalf("AIPrompts template not substituted; got %v", block.AIPrompts)
	}
}

// TestHintsLoad_MissingYamlReturnsEmptyBlock — nil Hints map returns
// zero Block with no error.
func TestHintsLoad_MissingYamlReturnsEmptyBlock(t *testing.T) {
	cat := fakeCatalogWithHints("tech-lead", nil)
	block, err := Load(cat, "tech-lead", "init", TemplateContext{})
	if err != nil {
		t.Fatalf("Load err: %v", err)
	}
	if !block.IsZero() {
		t.Fatalf("missing hints should produce zero Block; got %+v", block)
	}
}

// TestHintsLoad_MissingCommandInYaml — agent has hints.yaml but no
// "add" key → empty block.
func TestHintsLoad_MissingCommandInYaml(t *testing.T) {
	cat := fakeCatalogWithHints("tech-lead", map[string]catalog.HintSection{
		"init": {NextCLI: []string{"bonsai list"}},
	})
	block, err := Load(cat, "tech-lead", "add", TemplateContext{})
	if err != nil {
		t.Fatalf("Load err: %v", err)
	}
	if !block.IsZero() {
		t.Fatalf("missing command key should produce zero Block; got %+v", block)
	}
}

// TestHintsLoad_UnknownAgentTypeReturnsEmpty — Load against an agent
// type not in the catalog is a no-op zero block.
func TestHintsLoad_UnknownAgentTypeReturnsEmpty(t *testing.T) {
	cat := fakeCatalogWithHints("tech-lead", map[string]catalog.HintSection{
		"init": {NextCLI: []string{"bonsai list"}},
	})
	block, err := Load(cat, "no-such-agent", "init", TemplateContext{})
	if err != nil {
		t.Fatalf("Load err: %v", err)
	}
	if !block.IsZero() {
		t.Fatalf("unknown agent should produce zero Block; got %+v", block)
	}
}

// TestHintsLoad_NilCatalogReturnsEmpty — safe against nil cat (test
// harness / bootstrapping paths).
func TestHintsLoad_NilCatalogReturnsEmpty(t *testing.T) {
	block, err := Load(nil, "tech-lead", "init", TemplateContext{})
	if err != nil {
		t.Fatalf("Load err: %v", err)
	}
	if !block.IsZero() {
		t.Fatal("nil cat should produce zero Block")
	}
}

// TestHintsLoad_BadTemplateDropsEntry — a malformed template in one
// line is dropped silently; other lines render.
func TestHintsLoad_BadTemplateDropsEntry(t *testing.T) {
	cat := fakeCatalogWithHints("tech-lead", map[string]catalog.HintSection{
		"init": {
			NextCLI: []string{
				"good line",
				"{{ .BadField.Subfield }}", // executes against missing field
			},
		},
	})
	block, err := Load(cat, "tech-lead", "init", TemplateContext{})
	if err != nil {
		t.Fatalf("Load err: %v", err)
	}
	// Bad entry dropped — only the good one survives.
	if len(block.NextCLI) != 1 {
		t.Fatalf("NextCLI len = %d, want 1 (bad entry dropped)", len(block.NextCLI))
	}
	if block.NextCLI[0] != "good line" {
		t.Fatalf("NextCLI[0] = %q, want 'good line'", block.NextCLI[0])
	}
}

// TestHintsRender_ThreeSectionsPresent — a fully-populated block emits
// NEXT / TRY / ASK markers in the output.
func TestHintsRender_ThreeSectionsPresent(t *testing.T) {
	block := Block{
		NextCLI:      []string{"bonsai list"},
		NextWorkflow: []string{"edit Backlog.md"},
		AIPrompts: []Prompt{
			{Label: "Start", Body: "Hi"},
		},
	}
	out := Render(block, 84)
	if !strings.Contains(out, "NEXT STEPS") {
		t.Errorf("missing NEXT STEPS; got:\n%s", out)
	}
	if !strings.Contains(out, "TRY THIS") {
		t.Errorf("missing TRY THIS; got:\n%s", out)
	}
	if !strings.Contains(out, "ASK YOUR AGENT") {
		t.Errorf("missing ASK YOUR AGENT; got:\n%s", out)
	}
}

// TestHintsRender_ZeroBlockReturnsEmpty — zero Block renders to "" so
// callers without a hints source can stack Render unconditionally.
func TestHintsRender_ZeroBlockReturnsEmpty(t *testing.T) {
	out := Render(Block{}, 84)
	if out != "" {
		t.Fatalf("zero Block should render empty; got %q", out)
	}
}

// TestHintsRender_PartialBlockOnlyEmitsPresentSections — a block with
// only NextCLI set does NOT emit TRY / ASK headers.
func TestHintsRender_PartialBlockOnlyEmitsPresentSections(t *testing.T) {
	block := Block{
		NextCLI: []string{"bonsai list"},
	}
	out := Render(block, 84)
	if !strings.Contains(out, "NEXT STEPS") {
		t.Errorf("missing NEXT STEPS; got:\n%s", out)
	}
	if strings.Contains(out, "TRY THIS") {
		t.Errorf("TRY THIS should NOT appear when NextWorkflow empty; got:\n%s", out)
	}
	if strings.Contains(out, "ASK YOUR AGENT") {
		t.Errorf("ASK YOUR AGENT should NOT appear when AIPrompts empty; got:\n%s", out)
	}
}

// TestHintsRender_AIPromptBodyInBoxed — rendered prompt body is
// wrapped in a border (rounded corners) so terminal-users can
// select-copy the content. Smoke test: check the output contains
// the body text AND a border char.
func TestHintsRender_AIPromptBodyInBoxed(t *testing.T) {
	block := Block{
		AIPrompts: []Prompt{
			{Label: "Start", Body: "hello world"},
		},
	}
	out := Render(block, 84)
	if !strings.Contains(out, "hello world") {
		t.Fatalf("prompt body missing; got:\n%s", out)
	}
	// Rounded border chars — one of ╭ ╰ ─ │.
	if !strings.ContainsAny(out, "╭╰─│") {
		t.Fatalf("prompt body should be boxed (border chars missing); got:\n%s", out)
	}
}
