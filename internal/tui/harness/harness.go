// Package harness provides a single long-lived BubbleTea program that owns the
// screen for the lifetime of an interactive Bonsai command.
//
// The harness drives a linear stack of Steps. Each Step is a thin adapter
// wrapping an existing huh.Form (or other tea.Model) so the visual identity
// established by Plan 11/12/14 (semantic tokens, glyphs, panels, BonsaiTheme)
// is preserved — only the orchestration layer changes.
//
// Lifecycle:
//   - Run() builds a tea.Program with WithAltScreen(), drives it to completion,
//     and returns the per-step results in declaration order.
//   - Esc / Shift+Tab pops the cursor to the previous step (no-op on step 0).
//   - Ctrl-C aborts the flow; Run returns ErrAborted.
//   - When the cursor advances past the last step, the program quits cleanly
//     and the caller resumes normal stdout for spinner / write-result output.
package harness

import (
	"errors"
	"fmt"
	"runtime/debug"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/LastStep/Bonsai/internal/tui"
)

// ErrAborted is returned by Run when the user pressed Ctrl-C.
var ErrAborted = errors.New("flow aborted by user")

// BuilderPanicError is returned by Run when a builder closure (LazyStep.Build
// or LazyGroup.Splice) panicked. The harness recovers, exits AltScreen
// cleanly, and surfaces the panic as a typed error so the caller can render
// a structured FatalPanel instead of a stacktrace dumped mid-AltScreen.
type BuilderPanicError struct {
	Step  string // step title at the time of panic
	Value any    // recovered value
	Stack string // captured at recovery time via debug.Stack()
}

func (e *BuilderPanicError) Error() string {
	return fmt.Sprintf("harness: builder for step %q panicked: %v", e.Step, e.Value)
}

// Step is the contract every harness step satisfies. It is a tea.Model with
// extra metadata so the harness can render breadcrumbs and detect completion.
type Step interface {
	tea.Model
	// Title is the short label shown in the header crumb.
	Title() string
	// Result is the value produced by the step. nil while the step is pending.
	Result() any
	// Done signals the harness that this step is complete and the cursor
	// should advance.
	Done() bool
}

// lazyBuilder is implemented by LazyStep so the harness can build it on entry.
type lazyBuilder interface {
	// Build constructs the inner step using prior results. Called once, the
	// first time the cursor advances onto the lazy step.
	Build(prev []any)
	// Built reports whether Build has already run.
	Built() bool
}

// splicer is implemented by LazyGroup. When the harness cursor advances onto a
// splicer, the group is replaced in-place with the steps it returns, and the
// cursor stays at the same index (now pointing at the first of the new steps).
// Used for multi-step branches where the shape of the sub-sequence depends on
// prior answers.
//
// Limitations:
//   - Nested splicers are NOT supported. If Splice() returns a slice that
//     itself contains another splicer at the cursor position, the inner
//     splicer's Splice() will not run automatically — the harness only calls
//     expandSplicer() once per advance. Either flatten the splice to a single
//     level, or build the splice eagerly in the outer builder.
type splicer interface {
	Splice(prev []any) []Step
	Spliced() bool
}

// priorAware is an optional Step capability. When implemented, the harness
// invokes SetPrior(prev) immediately before Init/Build so the step can capture
// the up-to-date prior-results snapshot. Used by SpinnerStep (action body
// reads upstream picks) and ConditionalStep (predicate evaluates against
// upstream answers).
type priorAware interface {
	SetPrior(prev []any)
}

// resetter is an optional Step capability. The harness calls Reset() on a step
// after the cursor pops back onto it via Esc/Shift+Tab so the underlying form
// returns to StateNormal — otherwise the next keypress would re-trigger
// Done() and immediately advance again, defeating "go back".
type resetter interface {
	// Reset returns the step to a pre-completion state while preserving any
	// values the user already entered.
	Reset() tea.Cmd
}

// Harness is the root tea.Model that frames the active step.
type Harness struct {
	steps        []Step
	cursor       int
	width        int
	height       int
	banner       string // e.g. "BONSAI v0.1.3"
	action       string // e.g. "Initializing new project"
	quitting     bool
	aborted      bool
	builderPanic *BuilderPanicError
}

// New constructs a Harness. The caller should usually invoke Run rather than
// driving the model directly.
func New(banner, action string, steps []Step) *Harness {
	return &Harness{
		banner: banner,
		action: action,
		steps:  steps,
	}
}

// Aborted reports whether the user pressed Ctrl-C.
func (h *Harness) Aborted() bool { return h.aborted }

// recoverBuilder installs a deferred recovery for a builder closure call. On
// panic, it stores a BuilderPanicError on the harness and flips quitting=true
// so the next Update returns tea.Quit. aborted stays false — this is a
// distinct exit reason from a user Ctrl-C.
func (h *Harness) recoverBuilder(stepTitle string) {
	if r := recover(); r != nil {
		h.builderPanic = &BuilderPanicError{
			Step:  stepTitle,
			Value: r,
			Stack: string(debug.Stack()),
		}
		h.aborted = false
		h.quitting = true
	}
}

// invokeBuild calls Build on a lazyBuilder under recoverBuilder so a panic
// in the builder closure exits AltScreen cleanly instead of dumping a stack.
func (h *Harness) invokeBuild(lb lazyBuilder, stepTitle string) {
	defer h.recoverBuilder(stepTitle)
	lb.Build(h.priorResults())
}

// Init implements tea.Model.
func (h *Harness) Init() tea.Cmd {
	if len(h.steps) == 0 {
		return tea.Quit
	}
	h.expandSplicer()
	if h.builderPanic != nil || len(h.steps) == 0 {
		return tea.Quit
	}
	if pa, ok := h.steps[0].(priorAware); ok {
		pa.SetPrior(h.priorResults())
	}
	if lb, ok := h.steps[0].(lazyBuilder); ok && !lb.Built() {
		h.invokeBuild(lb, h.steps[0].Title())
		if h.builderPanic != nil {
			return tea.Quit
		}
	}
	return h.steps[0].Init()
	// NOTE: WindowSize re-broadcast on the very first Init is intentionally
	// skipped — h.width/h.height are zero before the first WindowSizeMsg, so
	// there's nothing useful to forward.
}

// expandSplicer replaces the step at h.cursor with its splice expansion if the
// step is a not-yet-spliced splicer. After splice the cursor stays at the same
// index, now pointing at the first of the new steps. Idempotent and guarded by
// Spliced(). On panic in the builder closure, sets h.builderPanic and returns
// without mutating h.steps so the next Update can return tea.Quit cleanly.
func (h *Harness) expandSplicer() {
	if h.cursor >= len(h.steps) {
		return
	}
	sp, ok := h.steps[h.cursor].(splicer)
	if !ok || sp.Spliced() {
		return
	}
	stepTitle := h.steps[h.cursor].Title()
	var inserted []Step
	func() {
		defer h.recoverBuilder(stepTitle)
		inserted = sp.Splice(h.priorResults())
	}()
	if h.builderPanic != nil {
		return
	}
	// Defensive: drop any nil entries so callers can use `nil` or `append`-of-
	// nothing as an "empty splice" signal without tripping a nil-method panic
	// downstream.
	filtered := inserted[:0]
	for _, s := range inserted {
		if s != nil {
			filtered = append(filtered, s)
		}
	}
	inserted = filtered
	head := append([]Step{}, h.steps[:h.cursor]...)
	tail := append([]Step(nil), h.steps[h.cursor+1:]...)
	h.steps = append(append(head, inserted...), tail...)
}

// rebroadcastWindowSize forwards a synthetic WindowSizeMsg with the harness's
// stored width/height to the active step so its first frame computes layout
// against the right dimensions instead of waiting for the next user keystroke.
// No-op if width/height haven't been initialised yet.
func (h *Harness) rebroadcastWindowSize() tea.Cmd {
	if h.width <= 0 || h.height <= 0 {
		return nil
	}
	if h.cursor >= len(h.steps) {
		return nil
	}
	updated, sizeCmd := h.steps[h.cursor].Update(tea.WindowSizeMsg{
		Width:  h.width,
		Height: h.height,
	})
	if step, ok := updated.(Step); ok {
		h.steps[h.cursor] = step
	}
	return sizeCmd
}

// Update implements tea.Model.
func (h *Harness) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		h.width = m.Width
		h.height = m.Height
		// Broadcast to the active step so embedded forms can resize.
		if h.cursor < len(h.steps) {
			updated, cmd := h.steps[h.cursor].Update(msg)
			if step, ok := updated.(Step); ok {
				h.steps[h.cursor] = step
			}
			return h, cmd
		}
		return h, nil

	case tea.KeyMsg:
		switch m.String() {
		case "ctrl+c":
			h.aborted = true
			h.quitting = true
			return h, tea.Quit
		case "esc", "shift+tab":
			// Pop to previous step if possible. Esc on step 0 is a no-op
			// (NOT a quit). Skip past steps that have no interactive form
			// (e.g. a MultiSelectStep where every item is required) — there
			// is nothing for the user to change there.
			origCursor := h.cursor
			for h.cursor > 0 {
				h.cursor--
				if !h.stepIsAutoComplete(h.cursor) {
					break
				}
			}
			if h.cursor == origCursor {
				return h, nil
			}
			// Reset every step from the new cursor through origCursor
			// (inclusive) so that:
			//   1. The active step's form leaves StateCompleted (otherwise
			//      the next keypress immediately re-advances).
			//   2. Stepping forward again re-shows downstream steps so the
			//      user can review/edit them, rather than auto-skipping
			//      straight to the review.
			//   3. A LazyStep at origCursor (typical for a review-panel step
			//      that captured prior results at build time) rebuilds on
			//      re-entry so its panel reflects the user's NEW picks
			//      rather than the stale pre-Esc snapshot.
			var cmds []tea.Cmd
			for i := h.cursor; i <= origCursor && i < len(h.steps); i++ {
				if r, ok := h.steps[i].(resetter); ok {
					if cmd := r.Reset(); cmd != nil && i == h.cursor {
						// Only the active step's Init cmd needs to fire now.
						cmds = append(cmds, cmd)
					}
				}
			}
			if len(cmds) > 0 {
				return h, tea.Batch(cmds...)
			}
			return h, nil
		}
	}

	if h.cursor >= len(h.steps) {
		h.quitting = true
		return h, tea.Quit
	}

	updated, cmd := h.steps[h.cursor].Update(msg)
	if step, ok := updated.(Step); ok {
		h.steps[h.cursor] = step
	}

	// Advance through any number of consecutively-Done steps. Steps that
	// auto-complete (e.g. a MultiSelectStep where every item is required)
	// should not block the flow.
	for h.cursor < len(h.steps) && h.steps[h.cursor].Done() {
		h.cursor++
		if h.cursor >= len(h.steps) {
			h.quitting = true
			return h, tea.Batch(cmd, tea.Quit)
		}
		// Expand a LazyGroup in place before any further setup. The cursor
		// stays at the same index, now pointing at the first spliced step.
		h.expandSplicer()
		if h.builderPanic != nil {
			h.quitting = true
			return h, tea.Batch(cmd, tea.Quit)
		}
		if h.cursor >= len(h.steps) {
			h.quitting = true
			return h, tea.Batch(cmd, tea.Quit)
		}
		// Forward prior-results into priorAware steps before Build/Init so the
		// closure body can read upstream picks.
		if pa, ok := h.steps[h.cursor].(priorAware); ok {
			pa.SetPrior(h.priorResults())
		}
		// Build LazyStep on entry.
		if lb, ok := h.steps[h.cursor].(lazyBuilder); ok && !lb.Built() {
			h.invokeBuild(lb, h.steps[h.cursor].Title())
			if h.builderPanic != nil {
				h.quitting = true
				return h, tea.Batch(cmd, tea.Quit)
			}
		}
		// Initialise the newly-active step (Huh forms need this so their
		// first field is focused).
		if init := h.steps[h.cursor].Init(); init != nil {
			cmd = tea.Batch(cmd, init)
		}
		// Re-broadcast WindowSize so the spliced/lazy step computes layout at
		// the harness's known dimensions instead of waiting for huh's next
		// update cycle. No-op until the first WindowSizeMsg has arrived.
		if sizeCmd := h.rebroadcastWindowSize(); sizeCmd != nil {
			cmd = tea.Batch(cmd, sizeCmd)
		}
	}

	return h, cmd
}

// View implements tea.Model.
func (h *Harness) View() string {
	if h.quitting {
		// Empty view on quit so AltScreen exits cleanly without leaving a
		// final frame burned into normal stdout.
		return ""
	}
	if h.cursor >= len(h.steps) {
		return ""
	}

	header := h.renderHeader()
	footer := h.renderFooter()
	body := h.steps[h.cursor].View()

	// Per BubbleTea Golden Rule #1, subtract the chrome height before clipping
	// the body. Header is two lines (banner row + blank separator), footer is
	// one line preceded by a blank — total chrome is 4 lines. We don't render
	// bordered panels here, but downstream Huh forms may; the height budget is
	// surfaced via WindowSizeMsg, so this calculation is informational only.
	const chromeLines = 4
	avail := h.height - chromeLines
	// Small-terminal guard: below the minimum viable height, the chrome would
	// overlap the body or clip everything. Show a plain notice instead of
	// panicking or rendering a broken frame. h.height=0 before the first
	// WindowSizeMsg, so this only triggers for genuinely tiny terminals.
	if h.height > 0 && avail < 3 {
		return lipgloss.NewStyle().
			Foreground(tui.ColorMuted).
			Render("Terminal too small — please resize to at least 8 rows.")
	}
	if avail > 0 {
		bodyLines := strings.Split(body, "\n")
		if len(bodyLines) > avail {
			bodyLines = bodyLines[:avail]
			body = strings.Join(bodyLines, "\n")
		}
	}

	return header + "\n\n" + body + "\n\n" + footer
}

// renderHeader builds the persistent crumb row at the top of every frame.
// Layout: BANNER (left) · ACTION (centre, muted) · [N/M] TITLE (right, accent).
func (h *Harness) renderHeader() string {
	width := h.width
	if width <= 0 {
		width = 80
	}

	left := tui.StyleTitle.Render(h.banner)
	centre := tui.StyleMuted.Render(h.action)

	crumbText := ""
	if h.cursor < len(h.steps) {
		crumbText = fmt.Sprintf("[%d/%d] %s",
			h.cursor+1, len(h.steps), h.steps[h.cursor].Title())
	}
	right := tui.HarnessCrumb.Render(crumbText)

	leftW := lipgloss.Width(left)
	centreW := lipgloss.Width(centre)
	rightW := lipgloss.Width(right)

	// Pad layout: [left] gap1 [centre] gap2 [right] within width.
	// Account for the HarnessHeader padding (2 cols each side).
	const horizPad = 4
	usable := width - horizPad
	if usable < leftW+centreW+rightW {
		// Not enough room for all three; collapse to just left + right.
		gap := usable - leftW - rightW
		if gap < 1 {
			gap = 1
		}
		return tui.HarnessHeader.Render(left + strings.Repeat(" ", gap) + right)
	}

	// Centre the action by computing left/right gaps.
	totalGap := usable - leftW - centreW - rightW
	leftGap := totalGap / 2
	rightGap := totalGap - leftGap
	if leftGap < 1 {
		leftGap = 1
	}
	if rightGap < 1 {
		rightGap = 1
	}

	row := left + strings.Repeat(" ", leftGap) + centre + strings.Repeat(" ", rightGap) + right
	return tui.HarnessHeader.Render(row)
}

// renderFooter builds the persistent key-hint row at the bottom of every frame.
func (h *Harness) renderFooter() string {
	hints := []string{
		tui.StyleMuted.Render("↵") + " continue",
		tui.StyleMuted.Render("esc") + " back",
		tui.StyleMuted.Render("ctrl-c") + " quit",
	}
	// Step 0 has no "back" target, so suppress the hint there.
	if h.cursor == 0 {
		hints = []string{
			tui.StyleMuted.Render("↵") + " continue",
			tui.StyleMuted.Render("ctrl-c") + " quit",
		}
	}
	sep := tui.StyleMuted.Render("  " + tui.GlyphDot + "  ")
	return tui.HarnessFooter.Render(strings.Join(hints, sep))
}

// stepIsAutoComplete reports whether the step at idx has no interactive form
// — used by the Esc-to-back logic so the user doesn't bounce forward through
// a section with nothing to change.
func (h *Harness) stepIsAutoComplete(idx int) bool {
	if idx < 0 || idx >= len(h.steps) {
		return false
	}
	type autoChecker interface{ AutoComplete() bool }
	if a, ok := h.steps[idx].(autoChecker); ok {
		return a.AutoComplete()
	}
	return false
}

// priorResults returns the slice of results for steps that have completed.
// Used by LazyStep on entry to construct itself based on earlier answers.
func (h *Harness) priorResults() []any {
	out := make([]any, 0, h.cursor)
	for i := 0; i < h.cursor && i < len(h.steps); i++ {
		out = append(out, h.steps[i].Result())
	}
	return out
}

// Run drives the harness to completion under tea.WithAltScreen.
//
// Returns the per-step results in declaration order, or ErrAborted if the
// user pressed Ctrl-C. Caller is responsible for any post-flow stdout output
// (spinner, success banner, etc.) — those render to normal stdout once the
// program has exited AltScreen.
func Run(banner, action string, steps []Step) ([]any, error) {
	h := New(banner, action, steps)
	prog := tea.NewProgram(h, tea.WithAltScreen())
	final, err := prog.Run()
	if err != nil {
		return nil, err
	}
	model, ok := final.(*Harness)
	if !ok {
		return nil, fmt.Errorf("harness: unexpected final model type %T", final)
	}
	if model.builderPanic != nil {
		return nil, model.builderPanic
	}
	if model.aborted {
		return nil, ErrAborted
	}
	results := make([]any, len(model.steps))
	for i, s := range model.steps {
		results[i] = s.Result()
	}
	return results, nil
}
