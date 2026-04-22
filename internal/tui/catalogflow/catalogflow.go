// Package catalogflow implements the cinematic tabbed browser for
// `bonsai catalog`. The package mirrors internal/tui/initflow's chrome
// shape (header + footer, rail hidden) around a single BubbleTea stage
// that cycles seven tabs over the seven catalog sections (Agents,
// Skills, Workflows, Protocols, Sensors, Routines, Scaffolding).
//
// Every chrome primitive (RenderHeader, RenderFooter,
// RenderMinSizeFloor, Viewport, PanelContentWidth, design tokens,
// WideCharSafe, TerminalTooSmall) is imported from initflow —
// catalogflow does not reimplement any of them. The package's single
// public entry is NewBrowser, consumed by cmd/catalog.go when stdout
// is a TTY; non-TTY invocations fall back to the static-render path
// already in cmd/catalog.go.
//
// Plan 28 reference: station/Playbook/Plans/Active/28-view-cmds-cinematic.md.
package catalogflow

// Entry is the per-row shape rendered in a single tab. Every catalog
// section (Agents, Skills, Workflows, Protocols, Sensors, Routines,
// Scaffolding) packs its items into this shape — per-category extras
// (Event, Matcher, Frequency, If Removed) go into the Meta dict keyed
// by their labelled form so the inline-expand renderer can surface
// them as `LABEL  value` rows.
//
// Agents is the AgentCompat.String() form (e.g. "all" or "tech-lead,
// code"). Required is the same format — empty string signals "not
// required for any agent".
type Entry struct {
	Name        string
	DisplayName string
	Description string
	Meta        map[string]string
	Agents      string
	Required    string
}
