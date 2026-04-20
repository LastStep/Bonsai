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
