// Package hints provides the 3-layer yield-stage hints renderer (Plan 31
// Phase H). Every cinematic yield stage — init's Planted, add's Yield,
// update's Yield (and, post-merge, remove's Yield) — consumes this
// package to render a three-section block at the base of the exit card:
//
//  1. NEXT STEPS     — mechanical CLI commands to run next
//  2. TRY THIS       — in-workspace workflow suggestions
//  3. ASK YOUR AGENT — copy-paste AI prompts
//
// Content is catalog-driven (catalog/agents/<type>/hints.yaml) and
// template-rendered against a TemplateContext so values like
// `{{ .DocsPath }}` resolve to the project's actual workspace path.
//
// Missing hints.yaml or missing per-command entries fall back to a zero
// Block silently — hints are optional and never block a flow on load
// failure. Zero-value Block renders to the empty string, safe for
// callers without a hints source (e.g. test fixtures).
package hints

// Block is the loaded hints payload for one (agent type, command) pair.
// Zero value renders to the empty string.
type Block struct {
	NextCLI      []string
	NextWorkflow []string
	AIPrompts    []Prompt
}

// Prompt is a single copy-paste AI prompt entry.
type Prompt struct {
	Label string
	Body  string
}

// IsZero reports whether the block has no content. Used by Render to
// short-circuit on caller paths that lack a hints source.
func (b Block) IsZero() bool {
	return len(b.NextCLI) == 0 && len(b.NextWorkflow) == 0 && len(b.AIPrompts) == 0
}

// TemplateContext is the substitution payload passed to Load. Fields are
// kept minimal; add to this struct when a new hints.yaml template var
// is needed.
type TemplateContext struct {
	DocsPath    string
	AgentName   string
	ProjectName string
}
