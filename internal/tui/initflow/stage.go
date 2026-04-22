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
	title    string       // breadcrumb title — unused today but kept for Step contract
	idx      int          // 0..3 — stage index in the rail
	label    StageLabel   // kanji/kana/English triple
	labels   []StageLabel // full rail label set (nil = fall back to init's StageLabels)
	width    int          // last seen terminal width
	height   int          // last seen terminal height
	ensoSafe bool         // WideCharSafe() snapshot captured at ctor time

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

// SetRailLabels installs a non-default label slice for the enso rail. Used by
// flows outside initflow (e.g. addflow's 6-stage set) that embed Stage but
// need their own rail segment count + kanji. A nil/empty slice restores the
// init-flow default (see RenderEnsoRail).
func (s *Stage) SetRailLabels(labels []StageLabel) { s.labels = labels }

// SetRailIndex overrides the 0-based rail position baked in at ctor time.
// Used by sibling flow packages that reuse an initflow stage at a different
// rail slot — e.g. addflow's Grow stage wraps GenerateStage (ctor idx=3) but
// actually sits at rail position 4 (育 GROW) in the 6-segment add rail.
func (s *Stage) SetRailIndex(idx int) { s.idx = idx }

// SetLabel overrides the stage's rendered kanji/kana/English triple. Used
// in tandem with SetRailIndex when a sibling package reuses an initflow
// stage struct but needs the body title to read from its own StageLabels
// slice.
func (s *Stage) SetLabel(label StageLabel) { s.label = label }

// SetSize stores the live terminal dims. Sibling packages that embed Stage
// (e.g. addflow) call this from their own Update on WindowSizeMsg because the
// fields are unexported.
func (s *Stage) SetSize(w, h int) { s.width = w; s.height = h }

// Width returns the last-seen terminal width. Zero until the first
// WindowSizeMsg lands.
func (s *Stage) Width() int { return s.width }

// Height returns the last-seen terminal height.
func (s *Stage) Height() int { return s.height }

// EnsoSafe returns the WideCharSafe() snapshot captured at ctor time. Stage
// bodies gate kanji glyphs on this so ASCII-only terminals get safe
// substitutes.
func (s *Stage) EnsoSafe() bool { return s.ensoSafe }

// Label returns the kanji/kana/English triple for this stage.
func (s *Stage) Label() StageLabel { return s.label }

// MarkDone flips the completion flag. Subclasses call this when Enter
// advances; the harness reads Done() to drive the cursor forward.
func (s *Stage) MarkDone() { s.done = true }

// ClearDone resets the completion flag. Used by Reset overrides that need to
// preserve other stage state.
func (s *Stage) ClearDone() { s.done = false }

// RenderFrame is the exported entry into the shared frame composer. Sibling
// packages that embed Stage call this from their View() so every flow's
// persistent chrome (header + enso rail + footer) renders identically. The
// unexported renderFrame remains for internal callers.
func (s *Stage) RenderFrame(body string, keys []KeyHint) string {
	return s.renderFrame(body, keys)
}

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
// Chrome is fixed-height and positioned rigidly: header always occupies
// rows 1-2, rail row 4, footer the last two rows. The body slot between
// rail and footer is padded (or truncated) to a fixed row count so per-
// stage content length never nudges the chrome. Eliminates the pre-
// 2026-04-22 jitter where inline error rows / varying list lengths
// shifted the footer up and down between renders.
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
	rail := RenderEnsoRail(s.idx, s.labels, width, s.ensoSafe)
	footer := RenderFooter(keys, width)

	// Chrome budget: header(2) + blank + rail(2) + blank + blank + footer(2) = 9 rows.
	// Rail is 2 rows (dots + labels); pad accordingly.
	count := func(s string) int { return strings.Count(s, "\n") + 1 }
	chromeRows := count(header) + 1 + count(rail) + 1 + 1 + count(footer)
	bodyTarget := height - chromeRows
	if bodyTarget < 1 {
		bodyTarget = 1
	}

	bodyRows := count(body)
	if bodyRows < bodyTarget {
		body = body + strings.Repeat("\n", bodyTarget-bodyRows)
	} else if bodyRows > bodyTarget {
		lines := strings.Split(body, "\n")
		body = strings.Join(lines[:bodyTarget], "\n")
	}

	return header + "\n\n" + rail + "\n\n" + body + "\n\n" + footer
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
