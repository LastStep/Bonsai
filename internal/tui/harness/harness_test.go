package harness

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// fakeStep is a minimal Step implementation used for reducer tests. It records
// every message it receives and exposes a `done` flag the test can flip to
// signal completion to the harness.
type fakeStep struct {
	title    string
	done     bool
	result   any
	received []tea.Msg
	width    int
	height   int
}

func (f *fakeStep) Title() string { return f.title }
func (f *fakeStep) Done() bool    { return f.done }
func (f *fakeStep) Result() any   { return f.result }
func (f *fakeStep) Init() tea.Cmd { return nil }
func (f *fakeStep) View() string  { return "fake:" + f.title }
func (f *fakeStep) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	f.received = append(f.received, msg)
	if w, ok := msg.(tea.WindowSizeMsg); ok {
		f.width = w.Width
		f.height = w.Height
	}
	return f, nil
}

func newFake(title string) *fakeStep { return &fakeStep{title: title} }

// fakeResetStep satisfies Step + the resetter interface used by the harness
// on Esc-back. It counts Reset() invocations and returns a sentinel tea.Cmd
// so the test can verify the harness folds it into its batch return.
type fakeResetStep struct {
	fakeStep
	resetCount int
	resetCmd   tea.Cmd
}

func (f *fakeResetStep) Reset() tea.Cmd {
	f.resetCount++
	return f.resetCmd
}

// Override Update so the embedded fakeStep's Update returns a pointer to the
// outer fakeResetStep (the harness expects the updated Step to satisfy Step —
// returning *fakeStep would drop the resetter implementation on the next tick).
func (f *fakeResetStep) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	f.received = append(f.received, msg)
	if w, ok := msg.(tea.WindowSizeMsg); ok {
		f.width = w.Width
		f.height = w.Height
	}
	return f, nil
}

// resetSentinelMsg is produced by the fake reset step's returned tea.Cmd so
// TestEscPopReinitsActiveStep can assert it made it into the harness's batch.
type resetSentinelMsg struct{}

// runeKey constructs a tea.KeyMsg the harness will see for a printable rune.
func runeKey(r rune) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
}

func TestHarnessAdvancesOnDone(t *testing.T) {
	a, b := newFake("a"), newFake("b")
	h := New("BANNER", "TEST", []Step{a, b})

	// Flip the active step's done flag, then send any message — the reducer
	// should advance the cursor past it.
	a.done = true
	updated, _ := h.Update(runeKey('x'))
	hh := updated.(*Harness)

	if hh.cursor != 1 {
		t.Fatalf("expected cursor=1 after Done step, got %d", hh.cursor)
	}
}

func TestHarnessQuitsAfterLastStep(t *testing.T) {
	a := newFake("a")
	h := New("BANNER", "TEST", []Step{a})
	a.done = true

	_, cmd := h.Update(runeKey('x'))
	if cmd == nil {
		t.Fatalf("expected tea.Quit cmd after last step completed, got nil")
	}
	// Execute the command and verify it surfaces a tea.QuitMsg somewhere in
	// the batch (Quit returns a QuitMsg when invoked).
	if !cmdContainsQuit(cmd) {
		t.Fatalf("expected returned cmd to contain tea.Quit")
	}
	if !h.quitting {
		t.Fatalf("expected harness.quitting=true after last step")
	}
}

func TestEscPopsCursor(t *testing.T) {
	a, b, c := newFake("a"), newFake("b"), newFake("c")
	h := New("BANNER", "TEST", []Step{a, b, c})
	h.cursor = 2

	updated, _ := h.Update(tea.KeyMsg{Type: tea.KeyEsc})
	hh := updated.(*Harness)

	if hh.cursor != 1 {
		t.Fatalf("expected cursor=1 after Esc, got %d", hh.cursor)
	}
	if hh.aborted {
		t.Fatalf("Esc must not set aborted")
	}
}

func TestEscOnFirstStepIgnored(t *testing.T) {
	a, b := newFake("a"), newFake("b")
	h := New("BANNER", "TEST", []Step{a, b})

	updated, cmd := h.Update(tea.KeyMsg{Type: tea.KeyEsc})
	hh := updated.(*Harness)

	if hh.cursor != 0 {
		t.Fatalf("expected cursor=0 (no-op) after Esc on first step, got %d", hh.cursor)
	}
	if hh.aborted {
		t.Fatalf("Esc on first step must not set aborted")
	}
	if cmdContainsQuit(cmd) {
		t.Fatalf("Esc on first step must not return tea.Quit")
	}
}

func TestCtrlCSetsAbortedAndQuits(t *testing.T) {
	a := newFake("a")
	h := New("BANNER", "TEST", []Step{a})

	_, cmd := h.Update(tea.KeyMsg{Type: tea.KeyCtrlC})

	if !h.aborted {
		t.Fatalf("expected aborted=true after ctrl-c")
	}
	if !cmdContainsQuit(cmd) {
		t.Fatalf("expected ctrl-c to return tea.Quit")
	}
}

func TestWindowSizeBroadcasts(t *testing.T) {
	a, b := newFake("a"), newFake("b")
	h := New("BANNER", "TEST", []Step{a, b})

	_, _ = h.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	if h.width != 120 || h.height != 40 {
		t.Fatalf("expected harness width/height set to 120/40, got %d/%d", h.width, h.height)
	}
	if a.width != 120 || a.height != 40 {
		t.Fatalf("expected active step to receive WindowSizeMsg with 120/40, got %d/%d", a.width, a.height)
	}
	if b.width != 0 {
		t.Fatalf("only the active step should receive the size; got width=%d on inactive step", b.width)
	}
}

func TestLazyStepBuildsOnEntry(t *testing.T) {
	a := newFake("a")
	a.result = "answer-a"

	var built bool
	var capturedPrev []any
	inner := newFake("inner")

	lazy := NewLazy("Lazy", func(prev []any) Step {
		built = true
		capturedPrev = prev
		return inner
	})

	h := New("BANNER", "TEST", []Step{a, lazy})

	if built {
		t.Fatalf("lazy step must not build before harness reaches it")
	}

	// Mark the first step done and feed a message to drive advancement.
	a.done = true
	_, _ = h.Update(runeKey('x'))

	if !built {
		t.Fatalf("expected lazy step to build on cursor entry")
	}
	if h.cursor != 1 {
		t.Fatalf("expected cursor=1 (lazy active), got %d", h.cursor)
	}
	if got, want := len(capturedPrev), 1; got != want {
		t.Fatalf("expected %d prior results, got %d", want, got)
	}
	if capturedPrev[0] != "answer-a" {
		t.Fatalf("expected prior result %q, got %v", "answer-a", capturedPrev[0])
	}

	// Re-trigger advancement; Build must NOT run a second time.
	built = false
	a.done = true
	inner.done = false
	_, _ = h.Update(runeKey('y'))
	if built {
		t.Fatalf("Build must be invoked exactly once")
	}
}

// TestEscPopReinitsActiveStep verifies that when Esc pops the cursor back, the
// step being popped onto has its Reset() called exactly once and the tea.Cmd
// Reset() returns is included in the harness's batch return. This guards the
// fix for the Plan 15 iter 1 regression where huh's unexported f.quitting=true
// made the popped-onto form render as an empty string.
func TestEscPopReinitsActiveStep(t *testing.T) {
	a := &fakeResetStep{fakeStep: fakeStep{title: "a"}}

	// b is the step being popped onto. Give its Reset() a sentinel cmd so
	// we can verify the harness batches it into the return.
	b := &fakeResetStep{fakeStep: fakeStep{title: "b"}}
	b.resetCmd = func() tea.Msg { return resetSentinelMsg{} }

	c := &fakeResetStep{fakeStep: fakeStep{title: "c"}}

	h := New("BANNER", "TEST", []Step{a, b, c})
	h.cursor = 2

	updated, cmd := h.Update(tea.KeyMsg{Type: tea.KeyEsc})
	hh := updated.(*Harness)

	if hh.cursor != 1 {
		t.Fatalf("expected cursor=1 after Esc, got %d", hh.cursor)
	}
	if b.resetCount != 1 {
		t.Fatalf("expected popped-onto step Reset() count=1, got %d", b.resetCount)
	}
	if a.resetCount != 0 {
		t.Fatalf("step before the new cursor must not be reset; got count=%d", a.resetCount)
	}
	// c was at the original cursor position. Plan 15 iter 2.1: the harness
	// now resets steps in [new_cursor, orig_cursor] (inclusive) so a
	// review-style LazyStep at origCursor rebuilds its closure on re-entry
	// rather than showing stale pre-Esc content. Reset() on a non-lazy step
	// is a cheap form rebuild that preserves the user's value pointer.
	if c.resetCount != 1 {
		t.Fatalf("step at original cursor position must be reset once (iter 2.1 inclusive-bound); got count=%d", c.resetCount)
	}
	if cmd == nil {
		t.Fatalf("expected harness to return a tea.Cmd batching the popped-onto step's Reset() cmd")
	}
	if !cmdContainsSentinel(cmd) {
		t.Fatalf("expected Reset()'s tea.Cmd sentinel to surface in the harness's batch return")
	}
}

// TestLazyGroupSplicesOnEntry verifies that when the cursor advances onto a
// LazyGroup, the group is replaced in-place with its expansion, the cursor
// stays at the same index (now pointing at the first new step), and the
// newly-active step's Init cmd is invoked.
func TestLazyGroupSplicesOnEntry(t *testing.T) {
	a := newFake("a")
	a.result = "answer-a"

	newA := newFake("spliced-1")
	newB := newFake("spliced-2")

	var built bool
	group := NewLazyGroup("Branch", func(prev []any) []Step {
		built = true
		return []Step{newA, newB}
	})

	h := New("BANNER", "TEST", []Step{a, group, newFake("tail")})

	// Flip step 0 done and drive advancement.
	a.done = true
	_, _ = h.Update(runeKey('x'))

	if !built {
		t.Fatalf("expected LazyGroup.build to fire on cursor entry")
	}
	if h.cursor != 1 {
		t.Fatalf("expected cursor=1 after splice (first spliced step), got %d", h.cursor)
	}
	if len(h.steps) != 4 {
		t.Fatalf("expected steps len=4 (3 orig − 1 group + 2 spliced), got %d", len(h.steps))
	}
	if h.steps[1] != Step(newA) {
		t.Fatalf("expected h.steps[1] to be the first spliced step")
	}
	if h.steps[2] != Step(newB) {
		t.Fatalf("expected h.steps[2] to be the second spliced step")
	}
}

// TestLazyGroupRunsOnce verifies the group's Splice builder fires exactly once
// even if the harness re-enters the index across subsequent ticks.
func TestLazyGroupRunsOnce(t *testing.T) {
	a := newFake("a")

	var buildCount int
	group := NewLazyGroup("Branch", func(prev []any) []Step {
		buildCount++
		return []Step{newFake("spliced")}
	})

	h := New("BANNER", "TEST", []Step{a, group})

	a.done = true
	_, _ = h.Update(runeKey('x'))

	if buildCount != 1 {
		t.Fatalf("expected build count=1 after first entry, got %d", buildCount)
	}

	// Drive another message while the spliced step is active — the group is
	// already gone from the list, so its builder must not fire again.
	_, _ = h.Update(runeKey('y'))
	if buildCount != 1 {
		t.Fatalf("expected build count=1 after subsequent ticks, got %d", buildCount)
	}
}

// TestLazyGroupPassesPriorResults verifies the builder sees the prior step's
// Result() in the closure argument so branch selection can depend on the
// earlier answer.
func TestLazyGroupPassesPriorResults(t *testing.T) {
	a := newFake("a")
	a.result = "agent-x"

	var capturedPrev []any
	group := NewLazyGroup("Branch", func(prev []any) []Step {
		capturedPrev = prev
		return []Step{newFake("spliced")}
	})

	h := New("BANNER", "TEST", []Step{a, group})

	a.done = true
	_, _ = h.Update(runeKey('x'))

	if got, want := len(capturedPrev), 1; got != want {
		t.Fatalf("expected %d prior results, got %d", want, got)
	}
	if capturedPrev[0] != "agent-x" {
		t.Fatalf("expected prev[0]=%q, got %v", "agent-x", capturedPrev[0])
	}
}

// TestSpliceRebroadcastsWindowSize verifies that when a LazyGroup splices in
// new steps, the harness immediately re-broadcasts a synthesised
// WindowSizeMsg to the newly-active step using the harness's stored
// width/height — without waiting for another real WindowSizeMsg.
func TestSpliceRebroadcastsWindowSize(t *testing.T) {
	a := newFake("a")
	spliced := newFake("spliced")
	group := NewLazyGroup("Branch", func(prev []any) []Step {
		return []Step{spliced}
	})
	tail := newFake("tail")
	h := New("BANNER", "TEST", []Step{a, group, tail})

	// Seed harness dimensions as if we'd received a WindowSizeMsg already.
	_, _ = h.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	// spliced has not yet been visited, so its width/height must still be 0.
	if spliced.width != 0 {
		t.Fatalf("setup: spliced step must not have received a size before activation")
	}

	// Drive the advance so the cursor moves past `a` and onto the splicer
	// expansion (which puts `spliced` at h.cursor).
	a.done = true
	_, _ = h.Update(runeKey('x'))

	if spliced.width != 120 || spliced.height != 40 {
		t.Fatalf("expected spliced step to receive re-broadcast WindowSize 120/40, got %d/%d",
			spliced.width, spliced.height)
	}
}

// panicSplicer is a splicer that panics inside Splice() so tests can verify
// the harness's recoverBuilder path.
type panicSplicer struct {
	title    string
	value    any
	splicedF bool
}

func (p *panicSplicer) Title() string                           { return p.title }
func (p *panicSplicer) Done() bool                              { return false }
func (p *panicSplicer) Result() any                             { return nil }
func (p *panicSplicer) Init() tea.Cmd                           { return nil }
func (p *panicSplicer) View() string                            { return "" }
func (p *panicSplicer) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return p, nil }
func (p *panicSplicer) Splice(prev []any) []Step {
	p.splicedF = true
	panic(p.value)
}
func (p *panicSplicer) Spliced() bool { return p.splicedF }

// TestSpliceBuilderPanicReturnsTypedError verifies that a panic inside the
// LazyGroup's Splice closure is recovered, the harness flips quitting=true,
// and a *BuilderPanicError is recorded on the model.
func TestSpliceBuilderPanicReturnsTypedError(t *testing.T) {
	a := newFake("a")
	group := &panicSplicer{title: "boom-group", value: "boom"}
	h := New("BANNER", "TEST", []Step{a, group})

	a.done = true
	_, cmd := h.Update(runeKey('x'))

	if h.builderPanic == nil {
		t.Fatalf("expected harness.builderPanic to be set after Splice panic")
	}
	if h.builderPanic.Step != "boom-group" {
		t.Fatalf("expected panic recorded against step %q, got %q", "boom-group", h.builderPanic.Step)
	}
	if h.builderPanic.Value != "boom" {
		t.Fatalf("expected panic value %q, got %v", "boom", h.builderPanic.Value)
	}
	if !h.quitting {
		t.Fatalf("expected harness.quitting=true after panic")
	}
	if h.aborted {
		t.Fatalf("panic must NOT set aborted (distinct from user Ctrl-C)")
	}
	if !cmdContainsQuit(cmd) {
		t.Fatalf("expected reducer to return tea.Quit after builder panic")
	}
}

// TestLazyBuilderPanicReturnsTypedError — same shape, but the panic is in a
// LazyStep's Build closure rather than a LazyGroup's Splice closure.
func TestLazyBuilderPanicReturnsTypedError(t *testing.T) {
	a := newFake("a")
	lazy := NewLazy("boom-lazy", func(prev []any) Step {
		panic("kaboom")
	})
	h := New("BANNER", "TEST", []Step{a, lazy})

	a.done = true
	_, cmd := h.Update(runeKey('x'))

	if h.builderPanic == nil {
		t.Fatalf("expected harness.builderPanic to be set after Build panic")
	}
	if h.builderPanic.Step != "boom-lazy" {
		t.Fatalf("expected panic recorded against step %q, got %q", "boom-lazy", h.builderPanic.Step)
	}
	if h.builderPanic.Value != "kaboom" {
		t.Fatalf("expected panic value %q, got %v", "kaboom", h.builderPanic.Value)
	}
	if !h.quitting {
		t.Fatalf("expected harness.quitting=true after panic")
	}
	if !cmdContainsQuit(cmd) {
		t.Fatalf("expected reducer to return tea.Quit after lazy build panic")
	}
}

// TestLazyStepInsideLazyGroupBuilds exercises the runtime pattern used by
// cmd/add.go and the iter-3 flows: a LazyGroup whose Splice() returns a slice
// that itself contains a LazyStep. The harness must, after splice, build the
// LazyStep on entry with the correct prior results, and Reset must rebuild
// the inner step against fresh prior results on re-entry.
func TestLazyStepInsideLazyGroupBuilds(t *testing.T) {
	a := newFake("a")
	a.result = "agent-x"

	first := newFake("first-spliced")

	var innerBuildCount int
	var innerCapturedPrev []any
	innerStep := newFake("inner-of-lazy")
	lazy := NewLazy("nested-lazy", func(prev []any) Step {
		innerBuildCount++
		// Snapshot prev so a later mutation by the harness doesn't affect what
		// we asserted on.
		innerCapturedPrev = append([]any(nil), prev...)
		return innerStep
	})

	group := NewLazyGroup("Branch", func(prev []any) []Step {
		return []Step{first, lazy}
	})

	h := New("BANNER", "TEST", []Step{a, group, newFake("tail")})

	// Drive past `a`. The splice expansion runs, cursor stays at index 1
	// (now `first-spliced`).
	a.done = true
	_, _ = h.Update(runeKey('x'))
	if h.cursor != 1 {
		t.Fatalf("expected cursor=1 after splice, got %d", h.cursor)
	}
	if innerBuildCount != 0 {
		t.Fatalf("inner LazyStep must NOT build until cursor advances onto it")
	}

	// Drive past `first-spliced` so cursor lands on the LazyStep inside the
	// splice expansion.
	first.done = true
	_, _ = h.Update(runeKey('y'))
	if h.cursor != 2 {
		t.Fatalf("expected cursor=2 (LazyStep active), got %d", h.cursor)
	}
	if innerBuildCount != 1 {
		t.Fatalf("expected inner LazyStep to build exactly once, got %d", innerBuildCount)
	}
	// prev at this point should contain results from `a` (slot 0) and
	// `first-spliced` (slot 1).
	if got := len(innerCapturedPrev); got != 2 {
		t.Fatalf("expected 2 prior results visible to the inner build, got %d", got)
	}
	if innerCapturedPrev[0] != "agent-x" {
		t.Fatalf("expected prev[0]=%q, got %v", "agent-x", innerCapturedPrev[0])
	}
}

// TestEscBackReevaluatesConditional covers Plan 19 Step 7 — when the user
// Esc-backs past a ConditionalStep to a preceding step, changes the upstream
// selection, and advances forward again, the conditional's predicate must
// re-evaluate against the NEW selection. Without the SetPrior-before-Reset
// fix in the harness + the predicate re-eval in ConditionalStep.Reset, the
// conditional would still use its stale initPrev snapshot and render (or
// skip) the wrong branch.
func TestEscBackReevaluatesConditional(t *testing.T) {
	// Step 0: a fake Select whose Result() returns the "picked" value. We
	// simulate the user changing the selection by mutating pick.result.
	pick := &fakeStep{title: "Select"}
	pick.result = "a"
	pick.done = true // gate forward advancement

	// Step 1: a ConditionalStep wrapping an inner fakeInitStep. Predicate
	// fires when the latest selection at prev[0] == "a".
	innerInitCount := 0
	inner := &fakeInitStep{
		fakeStep: fakeStep{title: "inner"},
		onInit:   func() { innerInitCount++ },
	}
	cond := NewConditional(inner, func(prev []any) bool {
		if len(prev) == 0 {
			return false
		}
		s, _ := prev[0].(string)
		return s == "a"
	})

	h := New("BANNER", "TEST", []Step{pick, cond})

	// Drive the harness forward off step 0 onto the conditional. With
	// pick.result == "a", predicate is true → inner shows.
	_, _ = h.Update(runeKey('x'))
	if h.cursor != 1 {
		t.Fatalf("setup: expected cursor=1 after forward advance, got %d", h.cursor)
	}
	if innerInitCount != 1 {
		t.Fatalf("setup: expected inner Init to run once on first entry (pick='a'), got %d", innerInitCount)
	}
	if cond.skip {
		t.Fatalf("setup: conditional must not skip when predicate true")
	}

	// User Esc-backs to the Select step. The inner step is a fakeInitStep
	// (NOT a resetter) so only the SetPrior/Reset walk for the Conditional
	// itself runs here. The harness's Esc-back loop calls SetPrior on every
	// step in [newCursor, origCursor] before Reset.
	_, _ = h.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if h.cursor != 0 {
		t.Fatalf("after Esc: expected cursor=0, got %d", h.cursor)
	}

	// User changes the pick to "b" → predicate should now return false and
	// the conditional should skip on the next forward pass.
	pick.result = "b"
	pick.done = true

	// Advance forward again. The Conditional is now active; its Reset ran
	// during the Esc loop with the fresh prior snapshot (pick="b"), so
	// c.skip should already be true. Done() is true via skipDone, and the
	// harness's auto-advance loop should sail past the conditional to the
	// end-of-steps quit path.
	_, cmd := h.Update(runeKey('y'))

	// With only two steps and the conditional skipping, the harness should
	// have advanced past both and started quitting.
	if !h.quitting {
		t.Fatalf("expected harness to quit after advancing past skipped conditional; cursor=%d, skip=%v",
			h.cursor, cond.skip)
	}
	if !cond.skip {
		t.Fatalf("expected conditional to skip with new pick='b', skip=%v", cond.skip)
	}
	// Inner Init must NOT have fired a second time (because the conditional
	// skipped on this forward pass).
	if innerInitCount != 1 {
		t.Fatalf("expected inner Init to NOT re-run after skip re-eval (total %d, want 1)", innerInitCount)
	}
	if !cmdContainsQuit(cmd) {
		t.Fatalf("expected tea.Quit in batch once harness runs off the end")
	}
}

// fakeChromelessStep satisfies Step plus Chromeless so the harness short-
// circuits its View() composition and yields the step's raw frame.
type fakeChromelessStep struct {
	fakeStep
	chromeless bool
}

func (f *fakeChromelessStep) Chromeless() bool { return f.chromeless }

// Override Update so the harness receives a value that still satisfies
// Chromeless on the next tick (dropping to *fakeStep would shed the capability).
func (f *fakeChromelessStep) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	f.received = append(f.received, msg)
	if w, ok := msg.(tea.WindowSizeMsg); ok {
		f.width = w.Width
		f.height = w.Height
	}
	return f, nil
}

// TestChromelessStepBypassesHarnessChrome verifies that when a step opts into
// the Chromeless capability, Harness.View returns the step's View() verbatim
// without prepending the default header/footer. Regression guard for the
// `internal/tui/initflow` cinematic stages that own the whole frame.
func TestChromelessStepBypassesHarnessChrome(t *testing.T) {
	const frame = "CINEMATIC-FRAME-BODY"
	cs := &fakeChromelessStep{
		fakeStep:   fakeStep{title: "cinematic"},
		chromeless: true,
	}
	// Override View via the embedded fakeStep field by replacing title — easier:
	// use a dedicated helper type below. Instead override inline with a wrapper.
	csFrame := &chromelessFrameStep{fakeChromelessStep: *cs, frame: frame}
	h := New("BANNER", "TEST", []Step{csFrame})
	// Seed a window size so the chrome-full path would otherwise have enough
	// room to render its header/footer — this isolates the Chromeless bypass.
	_, _ = h.Update(tea.WindowSizeMsg{Width: 120, Height: 32})

	got := h.View()
	if got != frame {
		t.Fatalf("expected Chromeless View() verbatim %q, got %q", frame, got)
	}
}

// chromelessFrameStep is a fakeChromelessStep that returns a canned View() so
// the test can assert the harness didn't add chrome around it.
type chromelessFrameStep struct {
	fakeChromelessStep
	frame string
}

func (f *chromelessFrameStep) View() string { return f.frame }

// Update must also return the outer type so the harness retains the custom
// View() across ticks.
func (f *chromelessFrameStep) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	f.received = append(f.received, msg)
	if w, ok := msg.(tea.WindowSizeMsg); ok {
		f.width = w.Width
		f.height = w.Height
	}
	return f, nil
}

// cmdContainsSentinel walks a tea.Cmd (expanding tea.BatchMsg one level) and
// returns true if any produced message is a resetSentinelMsg.
func cmdContainsSentinel(cmd tea.Cmd) bool {
	if cmd == nil {
		return false
	}
	msg := cmd()
	if msg == nil {
		return false
	}
	if _, ok := msg.(resetSentinelMsg); ok {
		return true
	}
	if batch, ok := msg.(tea.BatchMsg); ok {
		for _, sub := range batch {
			if cmdContainsSentinel(sub) {
				return true
			}
		}
	}
	return false
}

// cmdContainsQuit walks the result of executing a tea.Cmd and returns true if
// any produced message is tea.QuitMsg. tea.Batch returns a BatchMsg holding a
// slice of nested commands; we fan out one level to handle the common case.
func cmdContainsQuit(cmd tea.Cmd) bool {
	if cmd == nil {
		return false
	}
	msg := cmd()
	if msg == nil {
		return false
	}
	if _, ok := msg.(tea.QuitMsg); ok {
		return true
	}
	if batch, ok := msg.(tea.BatchMsg); ok {
		for _, sub := range batch {
			if cmdContainsQuit(sub) {
				return true
			}
		}
	}
	return false
}
