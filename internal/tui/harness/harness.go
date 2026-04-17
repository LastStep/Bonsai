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
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/LastStep/Bonsai/internal/tui"
)

// ErrAborted is returned by Run when the user pressed Ctrl-C.
var ErrAborted = errors.New("flow aborted by user")

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
	steps    []Step
	cursor   int
	width    int
	height   int
	banner   string // e.g. "BONSAI v0.1.3"
	action   string // e.g. "Initializing new project"
	quitting bool
	aborted  bool
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

// Init implements tea.Model.
func (h *Harness) Init() tea.Cmd {
	if len(h.steps) == 0 {
		return tea.Quit
	}
	if lb, ok := h.steps[0].(lazyBuilder); ok && !lb.Built() {
		lb.Build(h.priorResults())
	}
	return h.steps[0].Init()
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
			// Reset every step from the new cursor onward so that:
			//   1. The active step's form leaves StateCompleted (otherwise
			//      the next keypress immediately re-advances).
			//   2. Stepping forward again re-shows downstream steps so the
			//      user can review/edit them, rather than auto-skipping
			//      straight to the review.
			var cmds []tea.Cmd
			for i := h.cursor; i < origCursor && i < len(h.steps); i++ {
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
		// Build LazyStep on entry.
		if lb, ok := h.steps[h.cursor].(lazyBuilder); ok && !lb.Built() {
			lb.Build(h.priorResults())
		}
		// Initialise the newly-active step (Huh forms need this so their
		// first field is focused).
		if init := h.steps[h.cursor].Init(); init != nil {
			cmd = tea.Batch(cmd, init)
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
	if model.aborted {
		return nil, ErrAborted
	}
	results := make([]any, len(model.steps))
	for i, s := range model.steps {
		results[i] = s.Result()
	}
	return results, nil
}
