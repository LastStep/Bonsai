package updateflow

import (
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/LastStep/Bonsai/internal/catalog"
	"github.com/LastStep/Bonsai/internal/config"
	"github.com/LastStep/Bonsai/internal/generate"
	"github.com/LastStep/Bonsai/internal/tui"
	"github.com/LastStep/Bonsai/internal/tui/initflow"
)

// DiscoverStage is the first on-rail stage — it runs a read-only custom-file
// scan across every installed agent, splits the hits into valid (promotable)
// and invalid (frontmatter-missing) groups, and renders a preview panel so
// the user sees exactly what the flow is about to offer before committing.
//
// The stdout-warning pre-harness code from the legacy cmd/update.go:68-75 is
// moved into this stage's "warnings" body row — no stdout churn before the
// AltScreen paint, and the user gets the context inside the cinematic frame.
//
// Result: []AgentDiscoveries — one entry per agent with any discovery
// (valid or invalid). Empty slice when the workspace is in sync with the
// catalog; the downstream Sync stage handles that as a no-op.
type DiscoverStage struct {
	initflow.Stage

	// Inputs captured at ctor time; the scan runs on first Init.
	cfg  *config.ProjectConfig
	cat  *catalog.Catalog
	lock *config.LockFile
	cwd  string

	discoveries []AgentDiscoveries
}

// NewDiscoverStage constructs the Discover stage. The scan is deferred to
// the first Init call so the stage's ctor remains cheap.
func NewDiscoverStage(ctx initflow.StageContext, cwd string, cfg *config.ProjectConfig, cat *catalog.Catalog, lock *config.LockFile) *DiscoverStage {
	label := StageLabels[StageIdxDiscover]
	base := initflow.NewStage(
		StageIdxDiscover,
		label,
		label.English,
		ctx.Version,
		ctx.ProjectDir,
		ctx.StationDir,
		ctx.AgentDisplay,
		ctx.StartedAt,
	)
	base.ApplyContextHeader(ctx)
	base.SetRailLabels(StageLabels)
	s := &DiscoverStage{
		Stage: base,
		cfg:   cfg,
		cat:   cat,
		lock:  lock,
		cwd:   cwd,
	}
	s.scan()
	return s
}

// scan runs the per-agent custom-file discovery, splitting each agent's
// hits into valid + invalid groups. Matches legacy cmd/update.go:51-84
// behaviour verbatim — same sort order, same known-file filter — so the
// cinematic port is a pure visual wrapper over the existing logic.
func (s *DiscoverStage) scan() {
	var agentNames []string
	for name := range s.cfg.Agents {
		agentNames = append(agentNames, name)
	}
	sort.Strings(agentNames)

	var out []AgentDiscoveries
	for _, agentName := range agentNames {
		installed := s.cfg.Agents[agentName]
		discovered, scanErr := generate.ScanCustomFiles(s.cwd, installed, s.lock)
		if scanErr != nil || len(discovered) == 0 {
			continue
		}
		var valid, invalid []generate.DiscoveredFile
		for _, d := range discovered {
			if d.Error != "" {
				invalid = append(invalid, d)
			} else {
				valid = append(valid, d)
			}
		}
		if len(valid) == 0 && len(invalid) == 0 {
			continue
		}
		label := agentName
		if def := s.cat.GetAgent(installed.AgentType); def != nil {
			label = def.DisplayName
			if label == "" {
				label = catalog.DisplayNameFrom(def.Name)
			}
		}
		out = append(out, AgentDiscoveries{
			AgentName:  agentName,
			AgentLabel: label,
			Installed:  installed,
			Valid:      valid,
			Invalid:    invalid,
		})
	}
	s.discoveries = out
}

// Discoveries returns the scan result. Exposed for tests + downstream
// stages that consume the payload without going through the harness
// Result() path (e.g. the Sync stage reads this pointer directly).
func (s *DiscoverStage) Discoveries() []AgentDiscoveries { return s.discoveries }

// HasValidDiscoveries reports whether any agent has at least one valid
// (user-selectable) custom file. Used by the splicer to gate the Select
// stage — no selections ⇒ jump straight to Sync.
func (s *DiscoverStage) HasValidDiscoveries() bool {
	for _, a := range s.discoveries {
		if len(a.Valid) > 0 {
			return true
		}
	}
	return false
}

// Init implements tea.Model — no cmd on entry.
func (s *DiscoverStage) Init() tea.Cmd { return nil }

// Update handles Enter-to-advance. Esc is consumed by the harness root.
func (s *DiscoverStage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		s.SetSize(m.Width, m.Height)
	case tea.KeyMsg:
		switch m.String() {
		case "enter":
			s.MarkDone()
			return s, nil
		}
	}
	return s, nil
}

// View composes the Discover stage body inside the shared frame.
func (s *DiscoverStage) View() string {
	return s.RenderFrame(s.renderBody(), s.keyHints())
}

func (s *DiscoverStage) keyHints() []initflow.KeyHint {
	return []initflow.KeyHint{
		{Key: "↵", Desc: "continue"},
		{Key: "ctrl-c", Desc: "quit"},
	}
}

// renderBody composes the intro + FINDINGS block + (optional) WARNINGS
// block. Zero-discovery case renders the "nothing to promote" placeholder.
func (s *DiscoverStage) renderBody() string {
	white := lipgloss.NewStyle().Foreground(tui.ColorAccent).Bold(true)
	bark := initflow.LabelStyle()
	dim := initflow.DimStyle()

	var title string
	if s.EnsoSafe() {
		title = bark.Render(s.Label().Kanji) + " " + white.Render(s.Label().English)
	} else {
		title = white.Render(s.Label().English)
	}

	intro := strings.Join([]string{
		title,
		white.Render("Scanning for custom files to promote."),
		dim.Render("New abilities picked up from disk are promoted into the lockfile."),
	}, "\n")

	findings := s.renderFindings()
	warnings := s.renderWarnings()

	parts := []string{intro, "", "", findings}
	if warnings != "" {
		parts = append(parts, "", "", warnings)
	}
	return initflow.CenterBlock(strings.Join(parts, "\n"), s.Width())
}

func (s *DiscoverStage) renderFindings() string {
	bark := initflow.LabelStyle()
	dim := initflow.DimStyle()
	value := initflow.ValueStyle()

	panelW := initflow.PanelWidth(s.Width())
	header := initflow.RenderSectionHeader("FINDINGS", panelW)

	const labelW = 16
	const indent = "  "

	if len(s.discoveries) == 0 {
		return header + "\n" + indent + dim.Render("(workspace already in sync — nothing new to promote)")
	}

	rows := []string{header}
	totalValid := 0
	totalInvalid := 0
	for _, a := range s.discoveries {
		totalValid += len(a.Valid)
		totalInvalid += len(a.Invalid)
		count := fmt.Sprintf("%d valid · %d invalid", len(a.Valid), len(a.Invalid))
		rows = append(rows,
			indent+bark.Render(initflow.PadRight(a.AgentLabel, labelW))+value.Render(count))
		// Inline sample of first few valid entries — users get a taste
		// before the select stage reveals the full tab list.
		sample := sampleNames(a.Valid, 4)
		if sample != "" {
			rows = append(rows, indent+strings.Repeat(" ", labelW)+dim.Render(sample))
		}
	}
	summary := fmt.Sprintf("%d total discoveries across %d agent(s)", totalValid+totalInvalid, len(s.discoveries))
	rows = append(rows, "", indent+dim.Render(summary))
	return strings.Join(rows, "\n")
}

// renderWarnings renders the per-file invalid-frontmatter messages. Only
// emits when at least one agent has invalid files — otherwise returns ""
// so renderBody can omit the block.
func (s *DiscoverStage) renderWarnings() string {
	// Collect invalid files across every agent so the panel shows the
	// full set in one place.
	type inv struct {
		agent string
		df    generate.DiscoveredFile
	}
	var all []inv
	for _, a := range s.discoveries {
		for _, d := range a.Invalid {
			all = append(all, inv{agent: a.AgentLabel, df: d})
		}
	}
	if len(all) == 0 {
		return ""
	}

	danger := lipgloss.NewStyle().Foreground(tui.ColorWarning).Bold(true)
	dim := initflow.DimStyle()
	bark := initflow.LabelStyle()

	panelW := initflow.PanelWidth(s.Width())
	header := initflow.RenderSectionHeader("WARNINGS", panelW)
	const indent = "  "

	rows := []string{header}
	for _, e := range all {
		rows = append(rows,
			indent+danger.Render(tui.GlyphWarn)+" "+bark.Render(e.df.RelPath))
		rows = append(rows,
			indent+"  "+dim.Render(e.df.Error))
	}
	rows = append(rows, "", indent+dim.Render("Add frontmatter to track these files. See docs/custom-files.md."))
	return strings.Join(rows, "\n")
}

// sampleNames renders up to n names from the valid slice as a comma-joined
// preview. Truncates with an ellipsis when more remain.
func sampleNames(valid []generate.DiscoveredFile, n int) string {
	if len(valid) == 0 {
		return ""
	}
	names := make([]string, 0, n)
	limit := n
	if limit > len(valid) {
		limit = len(valid)
	}
	for i := 0; i < limit; i++ {
		label := valid[i].Meta.DisplayName
		if label == "" {
			label = valid[i].Name
		}
		names = append(names, fmt.Sprintf("[%s] %s", valid[i].Type, label))
	}
	out := strings.Join(names, ", ")
	if len(valid) > n {
		out += fmt.Sprintf(" · +%d more", len(valid)-n)
	}
	return out
}

// Result returns the discovery bundle. The downstream Select stage reads
// this through a SetDiscoveries shim rather than the harness prev[] slot
// to avoid the type-erasure cost; Result() is implemented for tests.
func (s *DiscoverStage) Result() any { return s.discoveries }

// Reset preserves the scan result; Esc-back redisplays the same findings
// without re-running the filesystem walk.
func (s *DiscoverStage) Reset() tea.Cmd {
	s.ClearDone()
	return nil
}
