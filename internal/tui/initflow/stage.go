package initflow

import (
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Stage is the shared base type for every init-flow stage (Vessel, Soil,
// Branches, Observe). It composes the persistent chrome (header + enso
// progress rail + footer) around a per-stage body, and satisfies both
// harness.Step and harness.Chromeless so the harness yields the entire frame.
//
// Subclasses embed Stage by value and call renderFrame(body, keys) from
// their View() to compose the final frame. Stage provides no-op defaults
// for Done / Result / Update / Init that subclasses override.
//
// The fields are carried on Stage (not the subclass) so each stage ctor can
// stamp them uniformly from cmd.runInit's context bundle.
type Stage struct {
	// Rendering context.
	title    string     // breadcrumb title — unused today but kept for Step contract
	idx      int        // 0..3 — stage index in the rail
	label    StageLabel // kanji/kana/English triple
	width    int        // last seen terminal width
	height   int        // last seen terminal height
	ensoSafe bool       // WideCharSafe() snapshot captured at ctor time

	// Project context — identical across all four stages, stamped by
	// cmd.runInit at entry so each stage renders the same header.
	projectDir   string    // absolute path to project root
	stationDir   string    // "station/" by default — can be updated post-Vessel
	version      string    // cmd.Version; blank/"dev" hides the version chip
	agentDisplay string    // agentDef.DisplayName — rendered by Observe's AGENT row
	startedAt    time.Time // captured at cmd.runInit entry for Planted's ELAPSED

	// State.
	done bool // set by subclass when Enter advances
}

// NewStage constructs the shared Stage bundle used by every subclass. idx is
// the 0-based rail position; label is typically StageLabels[idx]. Remaining
// fields are the project context captured by cmd.runInit.
func NewStage(
	idx int,
	label StageLabel,
	title string,
	version string,
	projectDir string,
	stationDir string,
	agentDisplay string,
	startedAt time.Time,
) Stage {
	return Stage{
		idx:          idx,
		label:        label,
		title:        title,
		version:      version,
		projectDir:   projectDir,
		stationDir:   stationDir,
		agentDisplay: agentDisplay,
		startedAt:    startedAt,
		ensoSafe:     WideCharSafe(),
	}
}

// Chromeless opts every stage out of the harness's default header/footer.
// The harness yields Stage.View() verbatim and Stage composes its own frame
// via renderFrame().
func (s *Stage) Chromeless() bool { return true }

// Title implements harness.Step. Used only if the harness ever had to
// surface a breadcrumb — which it does not in the cinematic flow.
func (s *Stage) Title() string { return s.title }

// Done implements harness.Step. Flipped by subclasses when Enter advances.
func (s *Stage) Done() bool { return s.done }

// Result implements harness.Step. Overridden by each subclass (Vessel returns
// a map, Soil returns []string, etc.). The base returns nil.
func (s *Stage) Result() any { return nil }

// Init implements tea.Model. No-op default; subclasses override when they
// need to kick a textinput cursor or spinner.
func (s *Stage) Init() tea.Cmd { return nil }

// Update implements tea.Model. The Phase-2 stub stages handle Enter to
// advance and Esc is consumed by the harness itself (pops the cursor).
// Subclasses will override this in Phase 3+ to drive their own key handling.
func (s *Stage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = m.Width
		s.height = m.Height
	case tea.KeyMsg:
		switch m.String() {
		case "enter":
			s.done = true
			return s, nil
		}
	}
	return s, nil
}

// View implements tea.Model. Base returns an empty frame — subclasses
// override to call renderFrame with the appropriate body and key hints.
func (s *Stage) View() string { return "" }

// Reset implements the harness.resetter contract. When the user Esc-backs
// onto a stage, the harness flips Done=false so the cursor doesn't
// immediately re-advance. Subclasses that hold extra state (textinput
// cursors, list indices) override this to preserve their values.
func (s *Stage) Reset() tea.Cmd {
	s.done = false
	return nil
}

// renderFrame composes header + enso rail + body + footer into the final
// AltScreen frame. body is the per-stage payload (rendered inside a padded
// content area); keys are the footer hints.
//
// The frame height is filled out to s.height (minus 1 for a terminal cursor
// row) so the AltScreen doesn't leave earlier frames visible at the bottom.
func (s *Stage) renderFrame(body string, keys []KeyHint) string {
	width := s.width
	if width <= 0 {
		width = 80
	}
	height := s.height
	if height <= 0 {
		height = 24
	}

	// Below the min-size floor, stage bodies would clip regardless of how
	// we lay them out — render a single "please enlarge" panel and skip
	// the rest of the frame composition. Gated on the live dims so the
	// pre-WindowSizeMsg default (80x24) still renders normally.
	if TerminalTooSmall(s.width, s.height) {
		return RenderMinSizeFloor(s.width, s.height)
	}

	header := RenderHeader(s.version, s.projectDir, width, s.ensoSafe)
	rail := RenderEnsoRail(s.idx, width, s.ensoSafe)
	footer := RenderFooter(keys, width)

	// Count rendered rows so we can pad the middle to fill the AltScreen.
	count := func(s string) int { return strings.Count(s, "\n") + 1 }
	headerRows := count(header)
	railRows := count(rail)
	footerRows := count(footer)
	bodyRows := count(body)

	// Compose: header, blank, rail, blank, body, <pad>, footer.
	// Separators contribute 5 blank lines (2 after header, 2 after rail, 1
	// before footer) — subtract to land the footer at the terminal bottom.
	padRows := height - headerRows - railRows - bodyRows - footerRows - 5
	if padRows < 1 {
		padRows = 1
	}
	pad := strings.Repeat("\n", padRows)

	return header + "\n\n" + rail + "\n\n" + body + "\n" + pad + footer
}

// DefaultKeys is the base footer key hint set used by the Phase-2 stubs.
// Phase 3+ subclasses supply their own slice that includes stage-specific
// hints (e.g. "␣ toggle").
func DefaultKeys(canGoBack bool) []KeyHint {
	if canGoBack {
		return []KeyHint{
			{Key: "↵", Desc: "continue"},
			{Key: "esc", Desc: "back"},
			{Key: "ctrl-c", Desc: "quit"},
		}
	}
	return []KeyHint{
		{Key: "↵", Desc: "continue"},
		{Key: "ctrl-c", Desc: "quit"},
	}
}

// StageContext is a small record used by cmd.runInit to stamp each stage
// with the shared project context at construction time. Keeps the call-site
// symmetric and obvious across all stage constructors.
type StageContext struct {
	Version      string
	ProjectDir   string
	StationDir   string
	AgentDisplay string
	StartedAt    time.Time
}

// homeDir returns $HOME via os.UserHomeDir. Kept wrapped so chrome.go can
// call it without importing os itself (keeping imports small on files that
// grow).
func homeDir() (string, error) { return os.UserHomeDir() }
