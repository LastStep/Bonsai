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

// renderBody renders the arc, title, and progress caption. Size scales
// down to 8x16 at widths <90 (plan spec).
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
		block := strings.Join([]string{
			danger.Render("生 · GENERATE FAILED"),
			"",
			dim.Render(msg),
			"",
			white.Render("Press ↵ to dismiss."),
		}, "\n")
		return centerBlock(block, s.width)
	}

	rows, cols := 12, 24
	if s.width < 90 {
		rows, cols = 8, 16
	}

	arc := renderArc(rows, cols, ticks, s.ensoSafe)

	// Progress label cycles 種 SEED → 苗 SPROUT → 盆栽 BONSAI every
	// ~60 ticks (≈2.5s) so longer catalogs still get the three-beat
	// crescendo.
	label := progressLabel(ticks, s.ensoSafe)

	title := "生 · PLANTING"
	if !s.ensoSafe {
		title = "PLANTING"
	}

	block := strings.Join([]string{
		bark.Render(title),
		"",
		arc,
		"",
		white.Render(label),
	}, "\n")
	return centerBlock(block, s.width)
}

// renderArc draws an ASCII ring of dots that light up progressively with
// tick count. The ring is rectangular (rows x cols) with the `生` kanji
// centred inside. Progress is ticks mod (ring perimeter) so the arc
// eventually circles back on very long operations.
func renderArc(rows, cols, ticks int, safe bool) string {
	if rows < 3 || cols < 3 {
		return ""
	}

	lit := "●"
	dim := "○"
	centre := "生"
	if !safe {
		lit = "#"
		dim = "."
		centre = "O"
	}

	// Build grid of dim cells.
	grid := make([][]string, rows)
	for r := 0; r < rows; r++ {
		row := make([]string, cols)
		for c := 0; c < cols; c++ {
			row[c] = " "
		}
		grid[r] = row
	}

	// Compute perimeter indices. Walk: top (L→R), right (T→B), bottom
	// (R→L), left (B→T).
	perim := make([][2]int, 0, 2*(rows+cols)-4)
	for c := 0; c < cols; c++ {
		perim = append(perim, [2]int{0, c})
	}
	for r := 1; r < rows; r++ {
		perim = append(perim, [2]int{r, cols - 1})
	}
	for c := cols - 2; c >= 0; c-- {
		perim = append(perim, [2]int{rows - 1, c})
	}
	for r := rows - 2; r >= 1; r-- {
		perim = append(perim, [2]int{r, 0})
	}

	// Fill all perimeter slots with the dim dot first.
	for _, p := range perim {
		grid[p[0]][p[1]] = dim
	}

	// Light up the prefix of length `ticks mod len(perim)` (or all if the
	// animation has lapped once + state advanced).
	n := len(perim)
	lightCount := ticks
	if lightCount > n {
		lightCount = lightCount % n
		if lightCount == 0 {
			lightCount = n
		}
	}
	for i := 0; i < lightCount; i++ {
		grid[perim[i][0]][perim[i][1]] = lit
	}

	// Place the centre kanji. Kanji is 2-cell, so inserting at centre col-1
	// means the glyph visually centres. On ASCII fallback the centre is a
	// single cell, so we don't need to shift.
	midR := rows / 2
	midC := cols / 2
	if safe {
		grid[midR][midC-1] = centre
		grid[midR][midC] = ""
	} else {
		grid[midR][midC] = centre
	}

	leafStyle := lipgloss.NewStyle().Foreground(tui.ColorPrimary).Bold(true)
	muted := lipgloss.NewStyle().Foreground(tui.ColorRule2)
	bark := lipgloss.NewStyle().Foreground(tui.ColorSecondary).Bold(true)

	var buf strings.Builder
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			cell := grid[r][c]
			switch cell {
			case lit:
				buf.WriteString(leafStyle.Render(cell))
			case dim:
				buf.WriteString(muted.Render(cell))
			case centre:
				buf.WriteString(bark.Render(cell))
			case "":
				// Placeholder when a wide-char glyph consumed this cell.
			default:
				buf.WriteString(cell)
			}
		}
		if r < rows-1 {
			buf.WriteString("\n")
		}
	}
	return buf.String()
}

// progressLabel cycles 種 SEED → 苗 SPROUT → 盆栽 BONSAI on a ~60-tick
// rhythm. ASCII fallback drops the kanji.
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

// Reset is a no-op: once Generate has finished it cannot re-run from the
// same instance. The harness should build a new GenerateStage on retry.
func (s *GenerateStage) Reset() tea.Cmd {
	s.done = false
	return nil
}
