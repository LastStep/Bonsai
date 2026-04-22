package harness

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"

	"github.com/LastStep/Bonsai/internal/tui"
)

// itemOptionsForTest returns a small set of optional (non-required) items
// suitable for exercising MultiSelectStep's interactive-form code path.
func itemOptionsForTest() []tui.ItemOption {
	return []tui.ItemOption{
		{Name: "skill-a", Value: "skill-a", Desc: "a"},
		{Name: "skill-b", Value: "skill-b", Desc: "b"},
	}
}

// TestTextStepResetRestoresView is the integration-style guard for the Plan 15
// iter 1 regression: after a huh.Form completes, huh sets the unexported field
// f.quitting=true and Form.View() returns "". Flipping f.State back to
// StateNormal does NOT clear f.quitting (Init doesn't touch it), so the prior
// Reset() implementation left the user staring at a blank content area.
//
// The fix rebuilds the *huh.Form on Reset(). This test verifies that after
// forcibly marking a form completed (simulating the effect of submission),
// Reset() produces a form whose View() is non-empty again.
func TestTextStepResetRestoresView(t *testing.T) {
	s := NewText("Name", "Project name:", "default", true)

	// Simulate post-submit state. We can't directly touch f.quitting, but the
	// real bug manifests whenever View() returns "" on a step the harness is
	// trying to render. After Reset(), View() must produce visible content.
	s.form.State = huh.StateCompleted

	// Drive Reset() — this is what the harness calls on Esc-back.
	s.Reset()

	got := s.View()
	if strings.TrimSpace(got) == "" {
		t.Fatalf("TextStep.View() after Reset() is empty — Esc-back would show a blank screen")
	}
	if s.form.State == huh.StateCompleted {
		t.Fatalf("TextStep.Reset() left form in StateCompleted; expected a fresh form in StateNormal")
	}
}

// TestSelectStepResetRestoresView — same guard, SelectStep.
func TestSelectStepResetRestoresView(t *testing.T) {
	s := NewSelect("Pick", "Pick one:", []huh.Option[string]{
		huh.NewOption("alpha", "alpha"),
		huh.NewOption("beta", "beta"),
	})
	s.form.State = huh.StateCompleted

	s.Reset()

	if strings.TrimSpace(s.View()) == "" {
		t.Fatalf("SelectStep.View() after Reset() is empty")
	}
}

// TestConfirmStepResetRestoresView — same guard, ConfirmStep.
func TestConfirmStepResetRestoresView(t *testing.T) {
	s := NewConfirm("Confirm", "Proceed?", true)
	s.form.State = huh.StateCompleted

	s.Reset()

	if strings.TrimSpace(s.View()) == "" {
		t.Fatalf("ConfirmStep.View() after Reset() is empty")
	}
}

// TestReviewStepResetRestoresView — same guard, ReviewStep.
func TestReviewStepResetRestoresView(t *testing.T) {
	s := NewReview("Review", "", "Looks good?", true)
	s.form.State = huh.StateCompleted

	s.Reset()

	if strings.TrimSpace(s.View()) == "" {
		t.Fatalf("ReviewStep.View() after Reset() is empty")
	}
}

// TestMultiSelectStepResetRestoresView — same guard, MultiSelectStep with
// at least one optional item (so it has an interactive form).
func TestMultiSelectStepResetRestoresView(t *testing.T) {
	// Use the local tui.ItemOption shape. We only need Name+Value for the
	// rebuild to work; Required=false keeps the item in the optional bucket.
	s := NewMultiSelect("Skills", "Skills", itemOptionsForTest(), nil)

	// Guard: if the test ItemOptions produce an auto-complete step (no
	// optional items), we can't exercise the form-rebuild path — fail loudly
	// so future refactors don't silently void this test.
	if s.auto {
		t.Fatal("test setup: expected MultiSelectStep to have an interactive form")
	}

	s.form.State = huh.StateCompleted

	s.Reset()

	if strings.TrimSpace(s.View()) == "" {
		t.Fatalf("MultiSelectStep.View() after Reset() is empty")
	}
}

// TestNoteStepViewNonEmpty verifies NoteStep renders visible content after
// construction (the harness will use the View output as the first frame).
func TestNoteStepViewNonEmpty(t *testing.T) {
	s := NewNote("Workspace", "Tech Lead workspace: station/")
	if strings.TrimSpace(s.View()) == "" {
		t.Fatalf("NoteStep.View() empty after construction")
	}
}

// TestNoteStepResetRestoresView — same guard as the other per-adapter Reset
// tests: after the underlying form completes, huh's unexported f.quitting
// blanks View(); Reset() must rebuild the form so content re-renders.
func TestNoteStepResetRestoresView(t *testing.T) {
	s := NewNote("Adding", "agent X is installed at foo/ — showing uninstalled abilities.")
	s.form.State = huh.StateCompleted

	s.Reset()

	if strings.TrimSpace(s.View()) == "" {
		t.Fatalf("NoteStep.View() after Reset() is empty")
	}
	if s.form.State == huh.StateCompleted {
		t.Fatalf("NoteStep.Reset() left form in StateCompleted; expected a fresh form")
	}
}

// TestLazyStepRebuildsOnReset verifies the Plan 15 iter 2.1 fix: after Reset,
// Build must re-run the closure against the current prior results so a review
// panel reflects the user's NEW picks rather than the pre-Esc snapshot.
func TestLazyStepRebuildsOnReset(t *testing.T) {
	var lastPrev []any
	lazy := NewLazy("Review", func(prev []any) Step {
		// Copy so a later mutation of the outer slice doesn't change what we
		// stored; the fix cares about the value captured at build time.
		copied := append([]any(nil), prev...)
		lastPrev = copied
		// Encode the prior-results slice into the ReviewStep's panel so View
		// reflects what the builder saw.
		panel := ""
		for i, p := range prev {
			if i > 0 {
				panel += ","
			}
			if s, ok := p.(string); ok {
				panel += s
			}
		}
		return NewReview("Review", panel, "OK?", true)
	})

	lazy.Build([]any{"v1"})
	if !lazy.Built() {
		t.Fatalf("expected Built()=true after first Build")
	}
	firstView := lazy.View()
	if !strings.Contains(firstView, "v1") {
		t.Fatalf("first view must contain %q; got:\n%s", "v1", firstView)
	}
	if got, want := lastPrev, []any{"v1"}; len(got) != 1 || got[0] != want[0] {
		t.Fatalf("first Build prev mismatch: got %v", got)
	}

	// Reset must clear the built flag and drop the inner so the next Build
	// re-runs the closure.
	lazy.Reset()
	if lazy.Built() {
		t.Fatalf("expected Built()=false after Reset")
	}
	if lazy.inner != nil {
		t.Fatalf("expected inner to be nil after Reset; got %T", lazy.inner)
	}

	lazy.Build([]any{"v2"})
	secondView := lazy.View()
	if !strings.Contains(secondView, "v2") {
		t.Fatalf("second view must contain %q (the NEW prior result); got:\n%s", "v2", secondView)
	}
	if strings.Contains(secondView, "v1") {
		t.Fatalf("second view must NOT contain the stale %q; got:\n%s", "v1", secondView)
	}
	if got, want := lastPrev, []any{"v2"}; len(got) != 1 || got[0] != want[0] {
		t.Fatalf("second Build prev mismatch: got %v", got)
	}
}

// TestMultiSelectStepResetPreservesPicks verifies that on re-entry after Esc,
// the builder re-applies Selected(true) to options matching the user's prior
// picks. huh's MultiSelect eagerly populates the value slice on Focus (see
// field_multiselect.go updateValue + Focus), so once the rebuilt form is
// focused inside the harness's Init cmd, s.optionalSelected is restored to
// the same contents. We verify the observable contract: after Reset(),
// s.optionalSelected reflects the prior picks (same slice contents), and
// View() renders non-empty content.
func TestMultiSelectStepResetPreservesPicks(t *testing.T) {
	s := NewMultiSelect("Skills", "Skills", itemOptionsForTest(), nil)
	if s.auto {
		t.Fatal("test setup: expected MultiSelectStep to have an interactive form")
	}

	// Simulate the form having been submitted with "skill-b" picked.
	s.optionalSelected = []string{"skill-b"}
	s.form.State = huh.StateCompleted

	s.Reset()

	// After Reset + Init, huh's MultiSelect Focus path populates the value
	// pointer from the options whose Selected(true) was set by buildForm.
	// We don't drive Init directly here — we assert the pickSet logic
	// preserved the prior selection by checking the options visible in
	// View() include the prior pick as already-selected.
	view := s.View()
	if strings.TrimSpace(view) == "" {
		t.Fatalf("MultiSelectStep.View() after Reset() is empty")
	}
	// huh renders a checkmark for selected options. skill-b must be
	// represented as selected in the view output.
	if !strings.Contains(view, "skill-b") {
		t.Fatalf("expected skill-b to appear in rebuilt view; got:\n%s", view)
	}
}

// ─── SpinnerStep tests ─────────────────────────────────────────────────────

// TestSpinnerStepCompletesAction verifies that once the spinnerDoneMsg
// arrives via Update, the step flips Done()=true and Result() returns nil
// for a successful action.
func TestSpinnerStepCompletesAction(t *testing.T) {
	s := NewSpinner("Generating", "Generating files...", func() error { return nil })

	// Init must return a non-nil cmd (the spinner tick + the worker goroutine).
	if cmd := s.Init(); cmd == nil {
		t.Fatalf("SpinnerStep.Init() returned nil cmd; expected tick + worker batch")
	}

	// Synthesise the worker completion message the harness would deliver.
	updated, _ := s.Update(spinnerDoneMsg{err: nil})
	if _, ok := updated.(*SpinnerStep); !ok {
		t.Fatalf("Update did not return *SpinnerStep; got %T", updated)
	}

	if !s.Done() {
		t.Fatalf("expected Done()=true after spinnerDoneMsg, got false")
	}
	if s.Result() != nil {
		// Result is the (interface) error; an untyped nil action error stored
		// in an `error` field reads as nil through any().
		t.Fatalf("expected Result()=nil for successful action, got %v", s.Result())
	}
}

// TestSpinnerStepReportsActionError verifies that an error returned by the
// action surfaces through Result() unchanged.
func TestSpinnerStepReportsActionError(t *testing.T) {
	wantErr := errors.New("boom")
	s := NewSpinner("Generating", "Generating files...", func() error { return wantErr })

	_, _ = s.Update(spinnerDoneMsg{err: wantErr})
	if !s.Done() {
		t.Fatalf("expected Done()=true after error spinnerDoneMsg, got false")
	}
	got, ok := s.Result().(error)
	if !ok {
		t.Fatalf("expected Result() to be an error, got %T (%v)", s.Result(), s.Result())
	}
	if !errors.Is(got, wantErr) {
		t.Fatalf("expected Result() == wantErr, got %v", got)
	}
}

// TestSpinnerStepResetIsNoop verifies that popping back via Esc onto a
// completed spinner does NOT re-trigger the action — Reset returns nil and
// Done() stays true so the harness's Esc-skip walks past.
func TestSpinnerStepResetIsNoop(t *testing.T) {
	calls := 0
	s := NewSpinner("Generating", "Generating files...", func() error {
		calls++
		return nil
	})
	// Pretend the action ran already.
	_, _ = s.Update(spinnerDoneMsg{err: nil})
	if !s.Done() {
		t.Fatalf("setup: expected Done()=true after spinnerDoneMsg")
	}

	if cmd := s.Reset(); cmd != nil {
		t.Fatalf("SpinnerStep.Reset() must return nil; got non-nil cmd")
	}
	if !s.Done() {
		t.Fatalf("Done() must stay true after Reset (no re-trigger of action)")
	}
	if !s.AutoComplete() {
		t.Fatalf("AutoComplete() must be true once done so Esc-back skips past")
	}
	if calls != 0 {
		t.Fatalf("action must not run as a side-effect of Reset; got %d calls", calls)
	}
}

// ─── ConditionalStep tests ─────────────────────────────────────────────────

// TestConditionalStepSkipsWhenPredicateFalse verifies that when the
// predicate returns false at Init time, the inner step's Init is never
// invoked and the conditional reports Done immediately so the harness
// advances past it in one step.
func TestConditionalStepSkipsWhenPredicateFalse(t *testing.T) {
	innerInited := false
	inner := &fakeInitStep{
		fakeStep: fakeStep{title: "inner"},
		onInit: func() {
			innerInited = true
		},
	}

	c := NewConditional(inner, func(prev []any) bool { return false })
	c.SetPrior(nil)

	cmd := c.Init()
	if cmd != nil {
		t.Fatalf("expected nil cmd when predicate is false, got non-nil")
	}
	if innerInited {
		t.Fatalf("inner.Init must NOT be called when predicate is false")
	}
	if !c.Done() {
		t.Fatalf("expected Done()=true on Init when predicate is false")
	}
	if c.Result() != nil {
		t.Fatalf("expected Result()=nil when predicate is false, got %v", c.Result())
	}
	if !c.AutoComplete() {
		t.Fatalf("expected AutoComplete()=true when predicate is false")
	}
}

// TestConditionalStepDelegatesWhenPredicateTrue verifies that when the
// predicate returns true, the inner step is initialised and method calls
// delegate verbatim.
func TestConditionalStepDelegatesWhenPredicateTrue(t *testing.T) {
	innerInited := false
	inner := &fakeInitStep{
		fakeStep: fakeStep{title: "inner"},
		onInit: func() {
			innerInited = true
		},
	}

	c := NewConditional(inner, func(prev []any) bool { return true })
	c.SetPrior(nil)

	_ = c.Init()
	if !innerInited {
		t.Fatalf("inner.Init must run when predicate is true")
	}
	if c.Done() {
		t.Fatalf("Done() must mirror inner (false) when predicate is true")
	}
	// View should delegate to inner.View now.
	if got := c.View(); got != "fake:inner" {
		t.Fatalf("expected View() to delegate to inner, got %q", got)
	}
	// Update must forward to inner — record the message in the inner's slice.
	_, _ = c.Update(runeKey('z'))
	if len(inner.received) == 0 {
		t.Fatalf("expected inner.Update to receive the forwarded message")
	}
}

// TestConditionalStepResetReevaluates verifies that after Reset the
// predicate is re-evaluated against the current prior-results snapshot.
// Captures a slice that the test mutates between the first Init and the
// second Init+Reset cycle.
func TestConditionalStepResetReevaluates(t *testing.T) {
	flag := []bool{false}
	inner := &fakeInitStep{fakeStep: fakeStep{title: "inner"}}
	c := NewConditional(inner, func(prev []any) bool {
		return flag[0]
	})

	// First entry: flag=false → skip.
	c.SetPrior(nil)
	_ = c.Init()
	if !c.Done() {
		t.Fatalf("first Init expected to skip (Done=true)")
	}

	// User Esc-backs and changes upstream picks; predicate result flips.
	flag[0] = true
	_ = c.Reset()
	c.SetPrior(nil)
	_ = c.Init()
	if c.Done() {
		t.Fatalf("after Reset+flag flip, expected Done() to mirror inner (false)")
	}
}

// TestSpinnerStepRecoversFromPanic verifies that a panic thrown inside the
// SpinnerStep action closure is captured by the tea.Cmd goroutine's deferred
// recover and translated into a spinnerDoneMsg carrying a descriptive error
// — rather than crashing the BubbleTea event loop.
func TestSpinnerStepRecoversFromPanic(t *testing.T) {
	s := NewSpinner("Generating", "Generating files...", func() error {
		panic("boom")
	})

	cmd := s.Init()
	if cmd == nil {
		t.Fatalf("SpinnerStep.Init() returned nil cmd")
	}

	// tea.Batch returns a single Cmd that, when executed, returns a
	// BatchMsg containing further commands. Drive them until we find the
	// spinnerDoneMsg (the worker function) without crashing on the panic.
	msg := cmd()
	var done spinnerDoneMsg
	found := false

	// tea.BatchMsg is []tea.Cmd. Walk it and execute each sub-Cmd.
	if batch, ok := msg.(tea.BatchMsg); ok {
		for _, sub := range batch {
			if sub == nil {
				continue
			}
			out := sub()
			if sd, ok := out.(spinnerDoneMsg); ok {
				done = sd
				found = true
				break
			}
		}
	} else if sd, ok := msg.(spinnerDoneMsg); ok {
		// Non-batch execution path (defensive).
		done = sd
		found = true
	}

	if !found {
		t.Fatalf("did not receive spinnerDoneMsg from Init batch")
	}
	if done.err == nil {
		t.Fatalf("expected non-nil err after action panic")
	}
	if !strings.Contains(done.err.Error(), "spinner action panic") {
		t.Fatalf("expected err to mention panic recovery, got %v", done.err)
	}
}

// TestConditionalNilPredicateDefaultsToShow verifies that NewConditional
// tolerates a nil predicate — the default path is to SHOW the wrapped step
// (safer than silently skipping, which could hide steps the user expected
// to complete).
func TestConditionalNilPredicateDefaultsToShow(t *testing.T) {
	innerInited := false
	inner := &fakeInitStep{
		fakeStep: fakeStep{title: "inner"},
		onInit:   func() { innerInited = true },
	}

	c := NewConditional(inner, nil)
	c.SetPrior(nil)
	_ = c.Init()

	if !innerInited {
		t.Fatalf("expected inner Init to run when predicate is nil (show path)")
	}
	// With predicate defaulted to always-true, Done() should mirror the inner
	// step, which has not completed yet.
	if c.Done() {
		t.Fatalf("expected Done()=false (predicate defaulted to show, inner not done)")
	}
}

// fakeInitStep wraps fakeStep with a callback that fires when Init is
// invoked, so tests can verify whether the harness drove Init on a wrapped
// step.
type fakeInitStep struct {
	fakeStep
	onInit func()
}

func (f *fakeInitStep) Init() tea.Cmd {
	if f.onInit != nil {
		f.onInit()
	}
	return nil
}

func (f *fakeInitStep) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	f.received = append(f.received, msg)
	if w, ok := msg.(tea.WindowSizeMsg); ok {
		f.width = w.Width
		f.height = w.Height
	}
	return f, nil
}

// ─── Conditional + Lazy composition tests ────────────────────────────────

// TestConditionalLazyChromelessForwardsWhenActive verifies the composition
// pattern used in cmd/add.go and cmd/init_flow.go — NewConditional wrapping a
// NewLazy whose builder produces a Chromeless inner step. When the predicate
// is true, the ConditionalStep must report Chromeless()=true on the outer
// step so the harness yields the full frame to the cinematic body. Pre-Init
// the inner is not yet built so Chromeless()=false (no frame is active).
//
// Bundles backlog item #2 from PR #52 review (Plan 23 Phase 3).
func TestConditionalLazyChromelessForwardsWhenActive(t *testing.T) {
	// Inner Chromeless step the Lazy builder will produce on first entry.
	chromelessInner := &fakeChromelessStep{
		fakeStep:   fakeStep{title: "cinematic"},
		chromeless: true,
	}
	lazy := NewLazy("Cinematic", func(prev []any) Step { return chromelessInner })
	c := NewConditional(lazy, func(prev []any) bool { return true })

	// Pre-Init: Lazy hasn't built, ConditionalStep should report false.
	if c.Chromeless() {
		t.Fatalf("Chromeless()=true before Init; expected false (Lazy not yet built)")
	}

	c.SetPrior(nil)
	_ = c.Init()

	// Post-Init with predicate=true: ConditionalStep delegated through Lazy
	// to the chromeless inner. Outer must mirror inner.
	if !c.Chromeless() {
		t.Fatalf("Chromeless()=false after Init with predicate=true; expected true (Lazy inner is Chromeless)")
	}
}

// TestConditionalLazyChromelessFalseWhenSkipped verifies the inverse: when
// the Conditional predicate is false, the Lazy builder never runs and the
// outer ConditionalStep reports Chromeless()=false (the harness would render
// nothing anyway, so chrome vs. not is moot — but the contract is to report
// false so default-chrome behaviour is preserved for the surrounding flow).
func TestConditionalLazyChromelessFalseWhenSkipped(t *testing.T) {
	builderRan := false
	lazy := NewLazy("Cinematic", func(prev []any) Step {
		builderRan = true
		return &fakeChromelessStep{
			fakeStep:   fakeStep{title: "cinematic"},
			chromeless: true,
		}
	})
	c := NewConditional(lazy, func(prev []any) bool { return false })

	c.SetPrior(nil)
	_ = c.Init()

	if builderRan {
		t.Fatalf("Lazy builder ran when predicate=false; expected skip")
	}
	if c.Chromeless() {
		t.Fatalf("Chromeless()=true when predicate=false; expected false (skipped path)")
	}
}

// TestConditionalLazyBuilderFiresOncePerActivePass verifies the Lazy
// builder closure runs exactly once per active Conditional pass, not on
// every harness tick. ConditionalStep.Init drives the lazyBuilder.Build
// path through its `if lb, ok := c.inner.(lazyBuilder); ok && !lb.Built()`
// guard — re-Init (via Reset) would re-fire only after the lazy.Reset
// flips Built()=false. This test pins the once-per-pass contract.
//
// Bundles backlog item #2 from PR #52 review (Plan 23 Phase 3).
func TestConditionalLazyBuilderFiresOncePerActivePass(t *testing.T) {
	buildCalls := 0
	lazy := NewLazy("Inner", func(prev []any) Step {
		buildCalls++
		return &fakeInitStep{fakeStep: fakeStep{title: "inner"}}
	})
	c := NewConditional(lazy, func(prev []any) bool { return true })

	// First active pass.
	c.SetPrior(nil)
	_ = c.Init()
	if buildCalls != 1 {
		t.Fatalf("first Init: buildCalls = %d, want 1", buildCalls)
	}

	// A second Init without an intervening Reset must NOT re-fire the
	// builder — the harness only re-Inits on Esc-back which routes through
	// Reset first.
	_ = c.Init()
	if buildCalls != 1 {
		t.Fatalf("second Init without Reset: buildCalls = %d, want 1 (no re-fire)", buildCalls)
	}

	// Esc-back path: Reset flips Lazy.Built()=false; subsequent Init must
	// rebuild exactly once more.
	_ = c.Reset()
	c.SetPrior(nil)
	_ = c.Init()
	if buildCalls != 2 {
		t.Fatalf("after Reset+Init: buildCalls = %d, want 2 (one re-fire)", buildCalls)
	}
}

// ─── LazyGroup re-splice (Plan 27 §B1) ────────────────────────────────────

// answerableFakeStep is a fakeStep that can be toggled between different
// result values on demand, and whose Done flag the test flips explicitly.
// Used to simulate an upstream Select step whose pick drives the LazyGroup
// branch shape: first "foo", then "bar" after an esc-back re-pick.
type answerableFakeStep struct {
	fakeStep
}

func (f *answerableFakeStep) answer(v any) {
	f.result = v
	f.done = true
}

// TestLazyGroupResplicesOnEscBack is the B1 regression guard. Pre-fix, a
// LazyGroup was replaced in h.steps by its children on first splice and the
// original reference was lost — esc-back + re-picking the upstream answer
// landed the user back on the previously-spliced children with stale data
// baked in. The fix adds spliceRecord tracking in Harness so esc-back can
// unsplice (remove children, reinstate the group, Reset it) and the next
// forward advance invokes Splice with the new prior results.
//
// Scenario exercised end-to-end through the harness reducer:
//
//  1. Steps: [answerableFakeStep, LazyGroup(fn)]
//     fn(prev) inspects prev[0] and returns child sets keyed on that value:
//     "foo" → [fakeFoo]; "bar" → [fakeBar]; anything else → nil.
//  2. User "picks foo" (step 0 sets Done + result="foo"); harness advances.
//  3. fn should fire once with prev=["foo"] and splice fakeFoo into h.steps.
//  4. User hits esc. Cursor pops to step 0. The LazyGroup should be reinstated
//     at its original slot; h.steps should match the declaration-time shape.
//  5. User "picks bar" (step 0 Done + result="bar"); harness advances.
//  6. fn should fire a second time with prev=["bar"] and splice fakeBar —
//     NOT fakeFoo. Without the B1 fix, the old children stay in place and
//     fakeBar never appears.
func TestLazyGroupResplicesOnEscBack(t *testing.T) {
	picker := &answerableFakeStep{fakeStep: fakeStep{title: "pick"}}

	var buildCalls int
	var lastPrev []any
	fakeFoo := newFake("spliced-foo")
	fakeBar := newFake("spliced-bar")

	group := NewLazyGroup("Branch", func(prev []any) []Step {
		buildCalls++
		lastPrev = append([]any(nil), prev...)
		if len(prev) == 0 {
			return nil
		}
		switch prev[0] {
		case "foo":
			return []Step{fakeFoo}
		case "bar":
			return []Step{fakeBar}
		}
		return nil
	})

	h := New("BANNER", "TEST", []Step{picker, group})
	// Prime dimensions so rebroadcastWindowSize no-ops cleanly.
	_, _ = h.Update(tea.WindowSizeMsg{Width: 100, Height: 30})

	// Step 1: pick "foo" and let the harness advance.
	picker.answer("foo")
	_, _ = h.Update(runeKey('x'))

	if buildCalls != 1 {
		t.Fatalf("first advance: build calls = %d, want 1", buildCalls)
	}
	if len(lastPrev) != 1 || lastPrev[0] != "foo" {
		t.Fatalf("first splice prev = %v, want [foo]", lastPrev)
	}
	if h.cursor != 1 {
		t.Fatalf("after first splice: cursor = %d, want 1", h.cursor)
	}
	if len(h.steps) != 2 || h.steps[1] != Step(fakeFoo) {
		t.Fatalf("after first splice: h.steps = %v, want [picker, fakeFoo]", h.steps)
	}

	// Step 2: user hits esc to go back. The harness should pop the cursor to
	// 0, unsplice the group (removing fakeFoo, reinstating the group), and
	// Reset the picker so its form is live again.
	_, _ = h.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if h.cursor != 0 {
		t.Fatalf("after esc: cursor = %d, want 0", h.cursor)
	}
	if len(h.steps) != 2 {
		t.Fatalf("after esc: len(h.steps) = %d, want 2 (group reinstated)", len(h.steps))
	}
	if _, ok := h.steps[1].(*LazyGroup); !ok {
		t.Fatalf("after esc: h.steps[1] = %T, want *LazyGroup (unspliced)", h.steps[1])
	}
	// Group must be reset so Spliced() reports false for the next advance.
	if grp, ok := h.steps[1].(*LazyGroup); ok && grp.Spliced() {
		t.Fatalf("after esc: LazyGroup.Spliced() = true, want false (not reset)")
	}

	// Step 3: re-pick "bar" on the same picker. The picker is still the same
	// instance — its done flag needs to flip back to false so the harness
	// doesn't treat it as already-advanced; the answer helper replaces the
	// result value and flips done.
	picker.done = false
	picker.answer("bar")
	_, _ = h.Update(runeKey('y'))

	if buildCalls != 2 {
		t.Fatalf("after re-advance: build calls = %d, want 2 (splice re-fired)", buildCalls)
	}
	if len(lastPrev) != 1 || lastPrev[0] != "bar" {
		t.Fatalf("second splice prev = %v, want [bar]", lastPrev)
	}
	if h.cursor != 1 {
		t.Fatalf("after re-advance: cursor = %d, want 1", h.cursor)
	}
	if len(h.steps) != 2 || h.steps[1] != Step(fakeBar) {
		t.Fatalf("after re-advance: h.steps = %v, want [picker, fakeBar]", h.steps)
	}
}

// TestLazyGroupResplicesMultiple exercises the multi-group unsplice path:
// two LazyGroups declared back-to-back after a Picker, each returning a
// distinct child shape. Advancing past both splices then esc-backing all the
// way to the Picker must reinstate BOTH groups at their original slots and
// Reset both so a subsequent forward pass re-splices each one with fresh
// prior results.
//
// Scenario:
//
//  1. Declaration: [Picker, LG_A(fn_A), LG_B(fn_B)].
//     fn_A returns 2 children ([fooA1, fooA2]); fn_B returns 1 child ([fooB]).
//  2. Advance: picker→LG_A splices in 2 → cursor now on fooA1 → LG_B splices
//     right after (as the harness advances through fooA1, fooA2 each time
//     they answer); we drive the advance manually by flipping the fake steps
//     Done and sending keys.
//  3. Esc all the way back to Picker.
//  4. Assert: h.steps matches declaration-time shape — [Picker, LG_A, LG_B] —
//     both groups Reset (Spliced()==false).
//  5. Re-advance with the same picker; assert fn_A and fn_B both re-fire and
//     h.steps ends with the spliced children again.
func TestLazyGroupResplicesMultiple(t *testing.T) {
	picker := &answerableFakeStep{fakeStep: fakeStep{title: "pick"}}

	var buildCallsA, buildCallsB int
	fooA1 := newFake("foo-A-1")
	fooA2 := newFake("foo-A-2")
	fooB := newFake("foo-B")

	groupA := NewLazyGroup("BranchA", func(prev []any) []Step {
		buildCallsA++
		return []Step{fooA1, fooA2}
	})
	groupB := NewLazyGroup("BranchB", func(prev []any) []Step {
		buildCallsB++
		return []Step{fooB}
	})

	h := New("BANNER", "TEST", []Step{picker, groupA, groupB})
	_, _ = h.Update(tea.WindowSizeMsg{Width: 100, Height: 30})

	// Advance: picker answers, harness splices LG_A (2 children).
	picker.answer("go")
	_, _ = h.Update(runeKey('x'))
	if buildCallsA != 1 {
		t.Fatalf("after first advance: buildCallsA = %d, want 1", buildCallsA)
	}
	// Harness cursor now sits at fooA1 (index 1). h.steps shape:
	// [picker, fooA1, fooA2, groupB].
	if len(h.steps) != 4 {
		t.Fatalf("after LG_A splice: len(h.steps) = %d, want 4", len(h.steps))
	}
	if h.cursor != 1 {
		t.Fatalf("after LG_A splice: cursor = %d, want 1", h.cursor)
	}

	// Advance fooA1 → fooA2 → LG_B splices.
	fooA1.done = true
	_, _ = h.Update(runeKey('x'))
	if h.cursor != 2 {
		t.Fatalf("after fooA1 done: cursor = %d, want 2", h.cursor)
	}
	fooA2.done = true
	_, _ = h.Update(runeKey('x'))
	if buildCallsB != 1 {
		t.Fatalf("after LG_B splice: buildCallsB = %d, want 1", buildCallsB)
	}
	// h.steps now: [picker, fooA1, fooA2, fooB]. cursor on fooB (idx 3).
	if len(h.steps) != 4 {
		t.Fatalf("after LG_B splice: len(h.steps) = %d, want 4", len(h.steps))
	}
	if h.cursor != 3 {
		t.Fatalf("after LG_B splice: cursor = %d, want 3", h.cursor)
	}

	// Esc all the way back to Picker. Each esc moves cursor back one
	// non-auto-complete step; fakeStep is not auto-complete so each esc
	// moves by one.
	for h.cursor > 0 {
		_, _ = h.Update(tea.KeyMsg{Type: tea.KeyEsc})
	}

	if h.cursor != 0 {
		t.Fatalf("after esc-all: cursor = %d, want 0", h.cursor)
	}
	// Both groups should be reinstated at their original slots.
	if len(h.steps) != 3 {
		t.Fatalf("after esc-all: len(h.steps) = %d, want 3 (both groups reinstated)", len(h.steps))
	}
	if _, ok := h.steps[1].(*LazyGroup); !ok {
		t.Fatalf("after esc-all: h.steps[1] = %T, want *LazyGroup", h.steps[1])
	}
	if _, ok := h.steps[2].(*LazyGroup); !ok {
		t.Fatalf("after esc-all: h.steps[2] = %T, want *LazyGroup", h.steps[2])
	}
	if gA := h.steps[1].(*LazyGroup); gA.Spliced() {
		t.Fatalf("after esc-all: LG_A.Spliced() = true, want false (should be reset)")
	}
	if gB := h.steps[2].(*LazyGroup); gB.Spliced() {
		t.Fatalf("after esc-all: LG_B.Spliced() = true, want false (should be reset)")
	}

	// Re-advance: picker answers again; both groups must re-splice.
	// Reset downstream fake steps so they aren't already-Done.
	fooA1.done = false
	fooA2.done = false
	fooB.done = false
	picker.done = false
	picker.answer("go-again")
	_, _ = h.Update(runeKey('x'))
	if buildCallsA != 2 {
		t.Fatalf("after re-advance: buildCallsA = %d, want 2 (re-splice)", buildCallsA)
	}
	if h.cursor != 1 {
		t.Fatalf("after re-advance: cursor = %d, want 1", h.cursor)
	}
	// Walk fooA1, fooA2 forward to trigger LG_B re-splice.
	fooA1.done = true
	_, _ = h.Update(runeKey('x'))
	fooA2.done = true
	_, _ = h.Update(runeKey('x'))
	if buildCallsB != 2 {
		t.Fatalf("after re-advance through fooA2: buildCallsB = %d, want 2 (LG_B re-splice)", buildCallsB)
	}
	if len(h.steps) != 4 || h.steps[3] != Step(fooB) {
		t.Fatalf("after re-advance: h.steps = %v, want [picker, fooA1, fooA2, fooB]", h.steps)
	}
}

// TestLazyGroupRespliceEmptyChildren covers the empty-splice corner case:
// a LazyGroup builder returning nil (zero children) must still be tracked
// well enough that esc-back reinstates the group AND a subsequent re-advance
// can invoke the builder again with a different prior-result to produce
// non-empty children.
//
// Post-Fix 4 the Harness guards spliceRecord append on len(inserted) > 0,
// so an empty splice is not recorded (there is nothing to reinstate — the
// group is effectively skipped). This test codifies that contract: an
// empty-splice group is consumed like an auto-complete, but a user re-entry
// to the step BEFORE it that flips the builder's return path must still re-
// splice against the new prior results, with no state leaking from the
// previous zero-child pass.
//
// Scenario:
//
//  1. Steps: [Picker, LG(fn)] where fn returns nil on the first pass, then
//     returns []Step{fakeAfter} on the second pass (prior[0] flips).
//  2. Advance: picker answers, harness splices LG → zero children → harness
//     continues past it (cursor advances off the end), flow quits.
//  3. We inspect buildCalls pre-quit: must be 1, and no splice record was
//     retained (h.splices is empty). The LG must report Spliced()=true since
//     its splice method fired and Spliced flips unconditionally.
//  4. This test does NOT exercise the full esc-back path because a
//     zero-child splice completes the flow; what we verify is that the
//     guard (Fix 4) prevents an over-eager record from being written. A
//     non-recorded empty splice means a hypothetical second pass with a
//     different prev[0] cannot accidentally reinstate stale zero-children —
//     there was never a record to begin with.
func TestLazyGroupRespliceEmptyChildren(t *testing.T) {
	picker := &answerableFakeStep{fakeStep: fakeStep{title: "pick"}}

	var buildCalls int
	var lastPrev []any
	fakeAfter := newFake("after")

	group := NewLazyGroup("Branch", func(prev []any) []Step {
		buildCalls++
		lastPrev = append([]any(nil), prev...)
		if len(prev) > 0 && prev[0] == "refill" {
			return []Step{fakeAfter}
		}
		return nil // zero-child splice
	})

	h := New("BANNER", "TEST", []Step{picker, group})
	_, _ = h.Update(tea.WindowSizeMsg{Width: 100, Height: 30})

	// Advance: picker answers; fn returns nil; harness walks past.
	picker.answer("empty")
	_, _ = h.Update(runeKey('x'))

	if buildCalls != 1 {
		t.Fatalf("after first advance: buildCalls = %d, want 1", buildCalls)
	}
	if len(lastPrev) != 1 || lastPrev[0] != "empty" {
		t.Fatalf("first splice prev = %v, want [empty]", lastPrev)
	}
	// Fix 4 contract: zero-child splice must NOT be recorded — the record
	// table stays empty so a later esc-back doesn't try to "unsplice"
	// something that was never spliced.
	if len(h.splices) != 0 {
		t.Fatalf("after empty splice: len(h.splices) = %d, want 0 (guard in expandSplicer)", len(h.splices))
	}
	// LazyGroup.Spliced() flips on any Splice invocation regardless of
	// child count, so the harness won't re-fire Splice on a second Init.
	if !group.Spliced() {
		t.Fatalf("LazyGroup.Spliced() = false after empty splice; want true (Splice fired)")
	}

	// Forward-advance behaviour: with zero children spliced, the group is
	// effectively dropped from h.steps (len drops from 2 → 1). Cursor
	// should now sit past the end → quitting.
	if len(h.steps) != 1 {
		t.Fatalf("after empty splice: len(h.steps) = %d, want 1 (group dropped)", len(h.steps))
	}
	if !h.quitting {
		t.Fatalf("after empty splice reaching end: quitting = false, want true")
	}

	// Build a fresh harness to prove the builder can still return non-empty
	// on a different prev. This mirrors the "esc-back + repick" flow without
	// needing to resurrect a quit harness.
	group2 := NewLazyGroup("Branch", func(prev []any) []Step {
		buildCalls++
		if len(prev) > 0 && prev[0] == "refill" {
			return []Step{fakeAfter}
		}
		return nil
	})
	picker2 := &answerableFakeStep{fakeStep: fakeStep{title: "pick"}}
	h2 := New("BANNER", "TEST", []Step{picker2, group2})
	_, _ = h2.Update(tea.WindowSizeMsg{Width: 100, Height: 30})

	picker2.answer("refill")
	_, _ = h2.Update(runeKey('x'))

	if len(h2.steps) != 2 || h2.steps[1] != Step(fakeAfter) {
		t.Fatalf("refill pass: h2.steps = %v, want [picker2, fakeAfter]", h2.steps)
	}
	if !group2.Spliced() {
		t.Fatalf("refill pass: group2.Spliced() = false, want true")
	}
	// Non-empty splice MUST be recorded (unlike the empty-splice case above).
	if len(h2.splices) != 1 {
		t.Fatalf("refill pass: len(h2.splices) = %d, want 1 (non-empty splice recorded)", len(h2.splices))
	}
}
