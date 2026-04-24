package hints

import (
	"bytes"
	"text/template"

	"github.com/LastStep/Bonsai/internal/catalog"
)

// Load returns the rendered hints Block for the given agent type +
// command. Lookup order:
//
//  1. Resolve agent hints map from cat (populated by catalog.LoadCatalog
//     via catalog/agents/<type>/hints.yaml).
//  2. Pick the sub-block for `command` ("init", "add", "update", "remove").
//  3. Render every field through text/template against `ctx`.
//
// Missing hints.yaml ⇒ zero Block, no error — hints are optional.
// Missing command key ⇒ zero Block, no error.
// Template parse/exec failures on any single entry ⇒ that entry is
// dropped silently; other entries in the block still render so a typo
// in one line doesn't nuke the whole panel.
func Load(cat *catalog.Catalog, agentType, command string, ctx TemplateContext) (Block, error) {
	if cat == nil {
		return Block{}, nil
	}
	def := cat.GetAgent(agentType)
	if def == nil || def.Hints == nil {
		return Block{}, nil
	}
	raw, ok := def.Hints[command]
	if !ok {
		return Block{}, nil
	}

	out := Block{
		NextCLI:      renderStrings(raw.NextCLI, ctx),
		NextWorkflow: renderStrings(raw.NextWorkflow, ctx),
		AIPrompts:    renderPrompts(raw.AIPrompts, ctx),
	}
	return out, nil
}

// renderStrings substitutes Go-template vars in each string, silently
// dropping entries that fail to parse/execute so a single typo doesn't
// blank the whole section.
func renderStrings(in []string, ctx TemplateContext) []string {
	if len(in) == 0 {
		return nil
	}
	out := make([]string, 0, len(in))
	for _, s := range in {
		rendered, ok := renderOne(s, ctx)
		if !ok {
			continue
		}
		out = append(out, rendered)
	}
	return out
}

func renderPrompts(in []catalog.HintPrompt, ctx TemplateContext) []Prompt {
	if len(in) == 0 {
		return nil
	}
	out := make([]Prompt, 0, len(in))
	for _, p := range in {
		label, okL := renderOne(p.Label, ctx)
		body, okB := renderOne(p.Body, ctx)
		if !okL || !okB {
			continue
		}
		out = append(out, Prompt{Label: label, Body: body})
	}
	return out
}

// renderOne parses s as a Go text/template and executes against ctx.
// Returns (rendered, false) on any error — caller filters dropped
// entries.
func renderOne(s string, ctx TemplateContext) (string, bool) {
	tmpl, err := template.New("hint").Parse(s)
	if err != nil {
		return "", false
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, ctx); err != nil {
		return "", false
	}
	return buf.String(), true
}
