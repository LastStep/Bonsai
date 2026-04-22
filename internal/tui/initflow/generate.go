package initflow

import (
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/LastStep/Bonsai/internal/tui"
)

// GenerateState is the lifecycle of a GenerateStage. Carried verbatim from
// Plan 22 Phase 5 so the states read identically in plan and code.
type GenerateState int

const (
	// stateRunning — action goroutine is active. Arc draws 0 → 360°.
	stateRunning GenerateState = iota
	// stateMinHold — action finished, but elapsed < minHold. Keep drawing
	// so the "something happened" beat is preserved.
	stateMinHold
	// stateDone — action finished AND minHold reached. Done() returns
	// true; the harness advances on the next tick.
	stateDone
	// stateError — action returned an error. Renders an error panel and
	// waits for a keypress (↵) before advancing so Done() can unblock.
	stateError
)

// minGenerateHold is the floor wait time before stateDone fires, even if
// the underlying action completes instantly. Keeps the cinematic beat on
// tiny catalogs.
const minGenerateHold = 600 * time.Millisecond

// generateTick is the tick interval — 42ms ≈ 24fps, matching the plan.
const generateTick = 42 * time.Millisecond

// GenerateAction is the unit of work Generate drives in its goroutine.
// Must be safe to call exactly once and must be goroutine-safe relative
// to the stage's View (stage reads state via the shared mutex on tick).
//
// Returns a ready-to-display WriteResult (stored by the caller) or an
// error that drives the stage into stateError. Both outputs are optional
// — many callers will capture the WriteResult via a pointer closure and
// return only an error from this signature.
type GenerateAction func() error

// generateDoneMsg is posted by the action goroutine (via tea.Cmd) when the
// action returns. It carries the error (nil on success) so the Update
// handler can transition into stateMinHold or stateError.
type generateDoneMsg struct {
	err     error
	elapsed time.Duration
}

// generateTickMsg is emitted by tea.Tick to drive the arc animation +
// min-hold timer. progress is monotonically increasing on every tick.
type generateTickMsg struct{}

// GenerateStage is the full-screen progress stage that runs the write
// pipeline. Wired between Observe and Planted in cmd.runInit.
type GenerateStage struct {
	Stage

	// The action runs in a goroutine started by Init(). mu guards the
	// state / err / elapsed fields that Update + the tick callback read.
	action GenerateAction

	mu        sync.Mutex
	state     GenerateState
	err       error
	startedAt time.Time // goroutine start for elapsed math
	ticks     int       // frames drawn since start

	// Optional body-title override. Sibling flow packages (addflow) swap
	// "生 · PLANTING" for their own kanji/English pair. Zero value keeps
	// the init-flow default.
	bodyKanji   string
	bodyEnglish string
}

// SetBodyTitle overrides the "生 · PLANTING" body header. Used by addflow's
// GrowStage wrapper so the kanji reads 育 (Grow) instead of the init-flow
// default. Empty kanji/english restore the default.
func (s *GenerateStage) SetBodyTitle(kanji, english string) {
	s.bodyKanji = kanji
	s.bodyEnglish = english
}

// NewGenerateStage constructs the Generate stage. The action closure wraps
// the caller's generate pipeline and is invoked once on Init. Remaining
// ctx fields follow the other stage constructors.
func NewGenerateStage(ctx StageContext, action GenerateAction) *GenerateStage {
	// Generate shares rail position 3 visually (it lives between Observe
	// and Planted; the rail only has 4 checkpoints — we reuse Observe's
	// slot so the progress beat still reads correctly).
	label := StageLabels[3]
	base := NewStage(
		3,
		label,
		label.English,
		ctx.Version,
		ctx.ProjectDir,
		ctx.StationDir,
		ctx.AgentDisplay,
		ctx.StartedAt,
	)
	base.applyContextHeader(ctx)
	return &GenerateStage{
		Stage:  base,
		action: action,
		state:  stateRunning,
	}
}

// Init starts the action goroutine + fires the first tick. The goroutine
// posts a generateDoneMsg via the returned tea.Cmd when the action ends.
func (s *GenerateStage) Init() tea.Cmd {
	s.mu.Lock()
	s.startedAt = time.Now()
	s.mu.Unlock()

	return tea.Batch(
		s.runAction(),
		s.tickCmd(),
	)
}

// runAction wraps the action in a tea.Cmd so its result arrives as a
// normal tea.Msg.
func (s *GenerateStage) runAction() tea.Cmd {
	return func() tea.Msg {
		start := time.Now()
		err := s.action()
		return generateDoneMsg{err: err, elapsed: time.Since(start)}
	}
}

// tickCmd schedules the next frame via tea.Tick.
func (s *GenerateStage) tickCmd() tea.Cmd {
	return tea.Tick(generateTick, func(time.Time) tea.Msg {
		return generateTickMsg{}
	})
}

// Update handles the action-done msg, tick animation, and the final
// keypress on stateError.
func (s *GenerateStage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = m.Width
		s.height = m.Height
		return s, nil
	case generateDoneMsg:
		s.mu.Lock()
		defer s.mu.Unlock()
		if m.err != nil {
			s.err = m.err
			s.state = stateError
			return s, nil
		}
		elapsed := time.Since(s.startedAt)
		if elapsed < minGenerateHold {
			s.state = stateMinHold
			// Keep ticking until minHold fires.
			return s, s.tickCmd()
		}
		s.state = stateDone
		s.done = true
		return s, nil
	case generateTickMsg:
		s.mu.Lock()
		s.ticks++
		// Promote stateMinHold → stateDone once the floor has elapsed.
		if s.state == stateMinHold {
			if time.Since(s.startedAt) >= minGenerateHold {
				s.state = stateDone
				s.done = true
				s.mu.Unlock()
				return s, nil
			}
		}
		keepTicking := s.state == stateRunning || s.state == stateMinHold
		s.mu.Unlock()
		if keepTicking {
			return s, s.tickCmd()
		}
		return s, nil
	case tea.KeyMsg:
		s.mu.Lock()
		st := s.state
		s.mu.Unlock()
		if st == stateError {
			switch m.String() {
			case "enter", "esc", "q":
				s.done = true
				return s, nil
			}
		}
	}
	return s, nil
}

// View composes the Generate stage body inside the shared frame.
func (s *GenerateStage) View() string {
	return s.renderFrame(s.renderBody(), s.keyHints())
}

// keyHints builds the footer key row. The key set is minimal while
// running; stateError swaps to an acknowledgement hint.
func (s *GenerateStage) keyHints() []KeyHint {
	s.mu.Lock()
	st := s.state
	s.mu.Unlock()
	if st == stateError {
		return []KeyHint{
			{Key: "↵", Desc: "acknowledge"},
			{Key: "ctrl-c", Desc: "quit"},
		}
	}
	return []KeyHint{
		{Key: "ctrl-c", Desc: "quit"},
	}
}

// renderBody renders the tree, title, and progress caption. Bonsai scales
// down from the wide 10-row template to a compact 7-row template at widths
// <90 so the stage still fits 80-col terminals.
func (s *GenerateStage) renderBody() string {
	s.mu.Lock()
	state := s.state
	err := s.err
	ticks := s.ticks
	s.mu.Unlock()

	white := lipgloss.NewStyle().Foreground(tui.ColorAccent).Bold(true)
	dim := lipgloss.NewStyle().Foreground(tui.ColorRule2)
	bark := lipgloss.NewStyle().Foreground(tui.ColorSecondary).Bold(true)
	danger := lipgloss.NewStyle().Foreground(tui.ColorDanger).Bold(true)

	if state == stateError {
		msg := "—"
		if err != nil {
			msg = err.Error()
		}
		errKanji := "生"
		errEnglish := "GENERATE FAILED"
		if s.bodyKanji != "" {
			errKanji = s.bodyKanji
		}
		if s.bodyEnglish != "" {
			// Addflow's Grow stage reads "育 · GROW FAILED" rather than
			// GENERATE FAILED when the body title has been overridden.
			errEnglish = s.bodyEnglish + " FAILED"
		}
		errTitle := errKanji + " · " + errEnglish
		if !s.ensoSafe {
			errTitle = errEnglish
		}
		block := strings.Join([]string{
			danger.Render(errTitle),
			"",
			dim.Render(msg),
			"",
			white.Render("Press ↵ to dismiss."),
		}, "\n")
		return centerBlock(block, s.width)
	}

	tree := renderBonsai(ticks, s.width >= 90, s.ensoSafe)

	label := progressLabel(ticks, s.ensoSafe)

	kanji := "生"
	english := "PLANTING"
	if s.bodyKanji != "" {
		kanji = s.bodyKanji
	}
	if s.bodyEnglish != "" {
		english = s.bodyEnglish
	}
	title := kanji + " · " + english
	if !s.ensoSafe {
		title = english
	}

	block := strings.Join([]string{
		bark.Render(title),
		"",
		tree,
		"",
		white.Render(label),
	}, "\n")
	return centerBlock(block, s.width)
}

// bonsaiTemplate pairs a character grid with a per-row reveal-layer array.
// Each rune position maps to a glyph class (leaf / trunk / pot / heart /
// blank) and the row's layer determines when it lights up in the reveal
// animation.
type bonsaiTemplate struct {
	grid   []string // fully-lit template, each rune is one cell
	layers []int    // parallel to grid: reveal layer for the row
}

// wideBonsai is the 10-row × 21-col template shown at ≥90 col terminals.
// Reveal order grows bottom-up: pot → trunk → canopy from inner to crown.
var wideBonsai = bonsaiTemplate{
	grid: []string{
		"         LLL         ", // layer 7 — crown
		"       LLLLLLL       ", // layer 6
		"     LLLLLLLLLLL     ", // layer 5
		"   LLLLLLLLLLLLLLL   ", // layer 4
		"     LLLLLLLLLLL     ", // layer 3
		"         lHl         ", // layer 2 — heart + trunk collar
		"          T          ", // layer 1 — trunk
		"          T          ", // layer 1
		"     PPPPPPPPPPP     ", // layer 0 — pot rim
		"      _________      ", // layer 0 — pot base
	},
	layers: []int{7, 6, 5, 4, 3, 2, 1, 1, 0, 0},
}

// narrowBonsai is the 7-row × 13-col template for <90-col terminals.
var narrowBonsai = bonsaiTemplate{
	grid: []string{
		"     LLL     ", // layer 4
		"   LLLLLLL   ", // layer 3
		" LLLLLLLLLLL ", // layer 2
		"     lHl     ", // layer 1 — heart
		"      T      ", // layer 1
		"   PPPPPPP   ", // layer 0 — pot
		"    _____    ", // layer 0
	},
	layers: []int{4, 3, 2, 1, 1, 0, 0},
}

// renderBonsai draws a bonsai tree that reveals layer-by-layer with tick
// count. ticksPerLayer controls the cadence — each layer lights up every
// ~2 ticks so the full animation completes inside ~14 ticks (≈600ms at
// 42ms/tick), matching minGenerateHold. Once all layers are revealed the
// canopy drops its dim "outer leaf" cells and renders fully bright.
func renderBonsai(ticks int, wide, safe bool) string {
	tpl := wideBonsai
	if !wide {
		tpl = narrowBonsai
	}

	leaf := lipgloss.NewStyle().Foreground(tui.ColorPrimary).Bold(true)
	leafDim := lipgloss.NewStyle().Foreground(tui.ColorPrimary)
	trunk := lipgloss.NewStyle().Foreground(tui.ColorSecondary).Bold(true)
	pot := lipgloss.NewStyle().Foreground(tui.ColorSecondary)
	base := lipgloss.NewStyle().Foreground(tui.ColorRule2)
	heart := lipgloss.NewStyle().Foreground(tui.ColorAccent).Bold(true)
	hidden := lipgloss.NewStyle().Foreground(tui.ColorRule2)

	const ticksPerLayer = 2
	revealedLayer := ticks / ticksPerLayer
	maxLayer := 0
	for _, l := range tpl.layers {
		if l > maxLayer {
			maxLayer = l
		}
	}
	allLit := revealedLayer >= maxLayer

	leafLit := "✦"
	leafOff := "·"
	trunkCh := "║"
	potCh := "═"
	baseCh := "─"
	heartCh := "生"
	heartW := 2
	if !safe {
		leafLit = "*"
		leafOff = "."
		trunkCh = "|"
		potCh = "="
		baseCh = "_"
		heartCh = "O"
		heartW = 1
	}

	var buf strings.Builder
	for r, row := range tpl.grid {
		layer := tpl.layers[r]
		lit := layer <= revealedLayer
		// Track column position explicitly so the wide heart glyph can
		// consume two cells cleanly on safe (kanji) terminals.
		for c := 0; c < len(row); c++ {
			ch := row[c]
			switch ch {
			case ' ':
				buf.WriteString(" ")
			case 'L':
				switch {
				case !lit:
					buf.WriteString(hidden.Render(leafOff))
				case allLit:
					buf.WriteString(leaf.Render(leafLit))
				default:
					buf.WriteString(leafDim.Render(leafLit))
				}
			case 'l':
				if lit {
					buf.WriteString(leafDim.Render(leafLit))
				} else {
					buf.WriteString(hidden.Render(leafOff))
				}
			case 'T':
				if lit {
					buf.WriteString(trunk.Render(trunkCh))
				} else {
					buf.WriteString(hidden.Render(leafOff))
				}
			case 'P':
				if lit {
					buf.WriteString(pot.Render(potCh))
				} else {
					buf.WriteString(hidden.Render(leafOff))
				}
			case '_':
				if lit {
					buf.WriteString(base.Render(baseCh))
				} else {
					buf.WriteString(" ")
				}
			case 'H':
				if lit {
					buf.WriteString(heart.Render(heartCh))
					// Safe heart is 2-cell — swallow the next column so
					// the row width stays correct. Narrow/ascii heart
					// is 1-cell so no swallow needed.
					if heartW == 2 && c+1 < len(row) {
						c++
					}
				} else {
					buf.WriteString(hidden.Render(leafOff))
				}
			default:
				buf.WriteString(string(ch))
			}
		}
		if r < len(tpl.grid)-1 {
			buf.WriteString("\n")
		}
	}
	return buf.String()
}

// progressLabel cycles 種 SEED → 苗 SPROUT → 盆栽 BONSAI on a ~20-tick
// rhythm so every planting (at ≥600ms / 14 ticks) shows at least the SEED
// beat, and longer operations cycle through all three. ASCII fallback
// drops the kanji.
func progressLabel(ticks int, safe bool) string {
	stages := [][2]string{
		{"種", "SEED"},
		{"苗", "SPROUT"},
		{"盆栽", "BONSAI"},
	}
	idx := (ticks / 20) % len(stages)
	if safe {
		return stages[idx][0] + " " + stages[idx][1]
	}
	return stages[idx][1]
}

// Result returns the stage's error (or nil). Callers typically consume the
// WriteResult through a closure captured in action; this return value is
// just the proceed/abort signal for downstream Conditional steps.
func (s *GenerateStage) Result() any {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.err != nil {
		return s.err
	}
	return nil
}

// Reset flips the completion flag so harness bookkeeping stays consistent
// after an Esc-back, but leaves the internal state alone. Once generation
// has run (stateDone or stateError) the action is not re-invoked — the
// harness should skip past this stage via AutoComplete.
func (s *GenerateStage) Reset() tea.Cmd {
	s.done = false
	return nil
}

// AutoComplete reports true once the generate action has finished — whether
// it succeeded or errored. This lets the harness's Esc-back loop skip past
// Generate so the user isn't stranded on a post-run frame after pressing
// Esc on the subsequent Planted stage. The one-shot nature of file writes
// means we never want re-entry into this stage's interactive state; the
// only way "back" is to the prior confirmation stage (Observe).
func (s *GenerateStage) AutoComplete() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.state == stateDone || s.state == stateError
}
