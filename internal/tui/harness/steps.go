package harness

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"github.com/LastStep/Bonsai/internal/tui"
)

// ─── TextStep ─────────────────────────────────────────────────────────────

// TextStep wraps huh.NewInput. defaultVal is used as a Placeholder and is also
// returned as the result if the user submits an empty string. validators run
// on the trimmed input — first validator failure wins. If required is true,
// the empty-input guard is prepended automatically.
type TextStep struct {
	title      string
	prompt     string
	value      string
	defaultVal string
	required   bool
	validators []func(string) error
	form       *huh.Form
}

// NewText constructs a TextStep.
//
//   - title: the breadcrumb / Huh title shown to the user.
//   - prompt: ignored today; reserved for a separate description line.
//   - defaultVal: placeholder text and fallback value on empty submit.
//   - required: prepend the "must not be empty" validator.
//   - validators: additional Validate funcs run on the trimmed value.
//
// All validators receive the user-entered string (not yet trimmed) so each can
// decide how strict to be — TextStep only trims for the required-empty check.
func NewText(title, prompt, defaultVal string, required bool, validators ...func(string) error) *TextStep {
	step := &TextStep{
		title:      title,
		prompt:     prompt,
		defaultVal: defaultVal,
		required:   required,
		validators: validators,
	}
	step.form = step.buildForm()
	return step
}

// buildForm constructs a fresh *huh.Form wired to this step's value pointer.
// Called from NewText and from Reset() — rebuilding (rather than flipping
// form.State) is required because huh sets the unexported field f.quitting=true
// on submit, and Form.View() returns "" while f.quitting is true. f.quitting
// is not reachable from outside the huh package, so on Esc-back we construct a
// new form with the same value pointer and validators and let it render
// normally. See huh form.go:560,576,649 and huh form.go:505 (Init does not
// clear f.quitting).
func (s *TextStep) buildForm() *huh.Form {
	input := huh.NewInput().
		Title(s.prompt).
		Value(&s.value)

	if s.defaultVal != "" {
		input.Placeholder(s.defaultVal)
	}

	chain := make([]func(string) error, 0, len(s.validators)+1)
	if s.required {
		chain = append(chain, func(v string) error {
			if strings.TrimSpace(v) == "" {
				return fmt.Errorf("required")
			}
			return nil
		})
	}
	chain = append(chain, s.validators...)

	if len(chain) > 0 {
		input.Validate(func(v string) error {
			for _, fn := range chain {
				if err := fn(v); err != nil {
					return err
				}
			}
			return nil
		})
	}

	return huh.NewForm(huh.NewGroup(input)).WithTheme(tui.BonsaiTheme())
}

// Title implements Step.
func (s *TextStep) Title() string { return s.title }

// Done implements Step.
func (s *TextStep) Done() bool { return s.form.State == huh.StateCompleted }

// Result implements Step. Falls back to defaultVal if the user submitted empty.
func (s *TextStep) Result() any {
	if s.value == "" && s.defaultVal != "" {
		return s.defaultVal
	}
	return s.value
}

// Init implements tea.Model.
func (s *TextStep) Init() tea.Cmd { return s.form.Init() }

// Update implements tea.Model.
func (s *TextStep) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	updated, cmd := s.form.Update(msg)
	if f, ok := updated.(*huh.Form); ok {
		s.form = f
	}
	return s, cmd
}

// View implements tea.Model.
func (s *TextStep) View() string { return s.form.View() }

// Reset rebuilds the step's *huh.Form so the user sees the prior input on
// re-entry. The value pointer (&s.value) stays the same, so whatever the user
// typed before Esc-back still shows in the field.
func (s *TextStep) Reset() tea.Cmd {
	s.form = s.buildForm()
	return s.form.Init()
}

// ─── SelectStep ───────────────────────────────────────────────────────────

// SelectStep wraps huh.NewSelect[string].
type SelectStep struct {
	title   string
	prompt  string
	value   string
	options []huh.Option[string]
	form    *huh.Form
}

// NewSelect constructs a SelectStep.
func NewSelect(title, prompt string, options []huh.Option[string]) *SelectStep {
	step := &SelectStep{
		title:   title,
		prompt:  prompt,
		options: options,
	}
	step.form = step.buildForm()
	return step
}

// buildForm constructs a fresh *huh.Form. See TextStep.buildForm for why we
// rebuild instead of toggling form.State on Reset.
func (s *SelectStep) buildForm() *huh.Form {
	sel := huh.NewSelect[string]().
		Title(s.prompt).
		Options(s.options...).
		Value(&s.value)
	return huh.NewForm(huh.NewGroup(sel)).WithTheme(tui.BonsaiTheme())
}

func (s *SelectStep) Title() string { return s.title }
func (s *SelectStep) Done() bool    { return s.form.State == huh.StateCompleted }
func (s *SelectStep) Result() any   { return s.value }
func (s *SelectStep) Init() tea.Cmd { return s.form.Init() }
func (s *SelectStep) View() string  { return s.form.View() }
func (s *SelectStep) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	updated, cmd := s.form.Update(msg)
	if f, ok := updated.(*huh.Form); ok {
		s.form = f
	}
	return s, cmd
}

// Reset rebuilds the form so the selection list is visible on re-entry with
// the prior pick still highlighted (value pointer preserved).
func (s *SelectStep) Reset() tea.Cmd {
	s.form = s.buildForm()
	return s.form.Init()
}

// ─── MultiSelectStep ──────────────────────────────────────────────────────

// MultiSelectStep wraps huh.NewMultiSelect[string] but mirrors the
// required/optional split logic from internal/tui/prompts.go:175-198 so that:
//   - required items are auto-included (always present in Result()),
//   - if the section is required-only, the form is skipped entirely and the
//     step auto-completes (Done()=true on first View),
//   - the visual chip line / per-item display matches the existing PickItems
//     output so users see the same shape inside or outside AltScreen.
type MultiSelectStep struct {
	title            string
	label            string
	required         []tui.ItemOption
	optional         []tui.ItemOption
	defaults         []string
	header           string // pre-rendered required-item display
	selected         []string
	optionalSelected []string // slice the form writes picks into; reused across rebuilds
	form             *huh.Form
	auto             bool // true when no optional items — auto-complete on entry
	autoFlipped      bool // already flipped to Done()
}

// NewMultiSelect constructs a MultiSelectStep.
//
//   - title: breadcrumb label.
//   - label: section heading printed above the form (used for the static
//     section-header line, equivalent to tui.Section(label)).
//   - available: the full list of options. Items with Required=true are split
//     out and auto-included.
//   - defaults: machine identifiers to pre-select among the optional items.
func NewMultiSelect(title, label string, available []tui.ItemOption, defaults []string) *MultiSelectStep {
	step := &MultiSelectStep{
		title:    title,
		label:    label,
		defaults: defaults,
	}

	for _, item := range available {
		if item.Required {
			step.required = append(step.required, item)
		} else {
			step.optional = append(step.optional, item)
		}
	}

	step.header = renderRequiredHeader(label, step.required, len(step.optional) == 0)

	// Seed Result() with required values up-front so consumers reading prior
	// results from a not-yet-displayed step still see the locked-in items.
	for _, r := range step.required {
		step.selected = append(step.selected, valueOf(r))
	}

	if len(step.optional) == 0 {
		step.auto = true
		return step
	}

	step.form = step.buildForm()
	return step
}

// buildForm constructs a fresh *huh.Form for the optional-picks portion of the
// multi-select. Called from NewMultiSelect and from Reset() — rebuilding (vs.
// flipping form.State) is required because huh's form.View() returns "" while
// its unexported f.quitting=true after submit. On re-entry, prior picks (held
// in s.optionalSelected) are re-applied via Selected(true) so the user's
// selections visibly persist.
func (s *MultiSelectStep) buildForm() *huh.Form {
	// Build a set of the user's currently-selected optional values. On first
	// build this starts empty and falls back to the provided defaults; on
	// subsequent rebuilds (Esc-back) it reflects whatever the user picked
	// before popping.
	pickSet := make(map[string]bool, len(s.optionalSelected))
	if len(s.optionalSelected) > 0 {
		for _, v := range s.optionalSelected {
			pickSet[v] = true
		}
	} else {
		for _, d := range s.defaults {
			pickSet[d] = true
		}
	}

	options := make([]huh.Option[string], 0, len(s.optional))
	for _, item := range s.optional {
		labelText := item.Name + " " + tui.StyleMuted.Render(tui.GlyphDash+" "+item.Desc)
		opt := huh.NewOption(labelText, valueOf(item))
		if pickSet[valueOf(item)] {
			opt = opt.Selected(true)
		}
		options = append(options, opt)
	}

	// Reset the slice length but keep the underlying array so the *huh.Form's
	// value pointer continues to point at the same destination across rebuilds.
	s.optionalSelected = s.optionalSelected[:0]

	ms := huh.NewMultiSelect[string]().
		Title("").
		Options(options...).
		Value(&s.optionalSelected)

	form := huh.NewForm(huh.NewGroup(ms)).WithTheme(tui.BonsaiTheme())

	// When the form completes, fold optional picks into selected on top of
	// the required prefix already seeded above.
	requiredCount := len(s.required)
	form.SubmitCmd = func() tea.Msg {
		// Truncate any prior optional picks (re-entry via Esc) and re-append.
		s.selected = s.selected[:requiredCount]
		s.selected = append(s.selected, s.optionalSelected...)
		return nil
	}
	return form
}

func (s *MultiSelectStep) Title() string { return s.title }

func (s *MultiSelectStep) Done() bool {
	if s.auto {
		return s.autoFlipped
	}
	if s.form == nil {
		return false
	}
	return s.form.State == huh.StateCompleted
}

func (s *MultiSelectStep) Result() any { return s.selected }

func (s *MultiSelectStep) Init() tea.Cmd {
	if s.auto {
		// Auto-complete on first Init — the harness will see Done()=true and
		// advance immediately.
		s.autoFlipped = true
		return nil
	}
	return s.form.Init()
}

func (s *MultiSelectStep) View() string {
	if s.auto {
		return s.header
	}
	if s.header != "" {
		return s.header + "\n" + s.form.View()
	}
	return s.form.View()
}

func (s *MultiSelectStep) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if s.auto {
		// No interactive form to drive; harness should already have advanced.
		s.autoFlipped = true
		return s, nil
	}
	updated, cmd := s.form.Update(msg)
	if f, ok := updated.(*huh.Form); ok {
		s.form = f
	}
	return s, cmd
}

// Reset rebuilds the form so the list of options is visible again on re-entry
// with the user's prior picks still highlighted. Auto-completing steps (no
// optional items) have no interactive form to rebuild — keep them completed.
func (s *MultiSelectStep) Reset() tea.Cmd {
	if s.auto {
		return nil
	}
	s.form = s.buildForm()
	return s.form.Init()
}

// AutoComplete reports whether this step has no interactive form to drive.
// Used by the harness to skip past such steps when popping back via Esc.
func (s *MultiSelectStep) AutoComplete() bool { return s.auto }

// ─── ConfirmStep ──────────────────────────────────────────────────────────

// ConfirmStep wraps huh.NewConfirm.
type ConfirmStep struct {
	title  string
	prompt string
	value  bool
	form   *huh.Form
}

// NewConfirm constructs a ConfirmStep.
func NewConfirm(title, prompt string, defaultVal bool) *ConfirmStep {
	step := &ConfirmStep{title: title, prompt: prompt, value: defaultVal}
	step.form = step.buildForm()
	return step
}

// buildForm constructs a fresh *huh.Form. See TextStep.buildForm for why we
// rebuild instead of toggling form.State on Reset.
func (s *ConfirmStep) buildForm() *huh.Form {
	c := huh.NewConfirm().
		Title(s.prompt).
		Affirmative("Yes").
		Negative("No").
		Value(&s.value)
	return huh.NewForm(huh.NewGroup(c)).WithTheme(tui.BonsaiTheme())
}

func (s *ConfirmStep) Title() string { return s.title }
func (s *ConfirmStep) Done() bool    { return s.form.State == huh.StateCompleted }
func (s *ConfirmStep) Result() any   { return s.value }
func (s *ConfirmStep) Init() tea.Cmd { return s.form.Init() }
func (s *ConfirmStep) View() string  { return s.form.View() }
func (s *ConfirmStep) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	updated, cmd := s.form.Update(msg)
	if f, ok := updated.(*huh.Form); ok {
		s.form = f
	}
	return s, cmd
}

// Reset rebuilds the form so the confirm prompt is visible again on re-entry,
// with the user's prior choice preserved via &s.value.
func (s *ConfirmStep) Reset() tea.Cmd {
	s.form = s.buildForm()
	return s.form.Init()
}

// ─── ReviewStep ───────────────────────────────────────────────────────────

// ReviewStep renders a static panel (typically a tui.ItemTree wrapped in a
// TitledPanel) above a confirm prompt. The panel content is supplied at
// construction time; ReviewStep itself does not synthesise the tree.
type ReviewStep struct {
	title  string
	prompt string
	panel  string
	value  bool
	form   *huh.Form
}

// NewReview constructs a ReviewStep.
//
//   - title: breadcrumb label.
//   - panel: pre-rendered review block (e.g. tui.ItemTree wrapped). Rendered
//     verbatim above the confirm prompt.
//   - prompt: confirm question.
//   - defaultVal: yes/no default.
func NewReview(title, panel, prompt string, defaultVal bool) *ReviewStep {
	step := &ReviewStep{title: title, prompt: prompt, panel: panel, value: defaultVal}
	step.form = step.buildForm()
	return step
}

// buildForm constructs a fresh *huh.Form for the confirm prompt. See
// TextStep.buildForm for why we rebuild instead of toggling form.State.
func (s *ReviewStep) buildForm() *huh.Form {
	c := huh.NewConfirm().
		Title(s.prompt).
		Affirmative("Yes").
		Negative("No").
		Value(&s.value)
	return huh.NewForm(huh.NewGroup(c)).WithTheme(tui.BonsaiTheme())
}

func (s *ReviewStep) Title() string { return s.title }
func (s *ReviewStep) Done() bool    { return s.form.State == huh.StateCompleted }
func (s *ReviewStep) Result() any   { return s.value }
func (s *ReviewStep) Init() tea.Cmd { return s.form.Init() }
func (s *ReviewStep) View() string {
	if s.panel == "" {
		return s.form.View()
	}
	return s.panel + "\n" + s.form.View()
}
func (s *ReviewStep) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	updated, cmd := s.form.Update(msg)
	if f, ok := updated.(*huh.Form); ok {
		s.form = f
	}
	return s, cmd
}

// Reset rebuilds the confirm form so it renders on re-entry, with prior
// choice preserved via &s.value.
func (s *ReviewStep) Reset() tea.Cmd {
	s.form = s.buildForm()
	return s.form.Init()
}

// ─── LazyStep ─────────────────────────────────────────────────────────────

// LazyStep defers construction of its inner Step until the harness cursor
// advances onto it. This lets later steps (e.g. a review panel) read prior
// answers without leaving AltScreen.
//
// The closure is invoked exactly once. Subsequent re-entries via Esc reuse the
// already-built inner step, so any partial state the user entered is preserved.
type LazyStep struct {
	title string
	build func(prev []any) Step
	inner Step
	built bool
}

// NewLazy constructs a LazyStep.
func NewLazy(title string, build func(prev []any) Step) *LazyStep {
	return &LazyStep{title: title, build: build}
}

// Build is invoked by the harness when the cursor advances onto the lazy step.
// It runs the closure and stores the inner step. Idempotent.
func (l *LazyStep) Build(prev []any) {
	if l.built {
		return
	}
	l.inner = l.build(prev)
	l.built = true
}

// Built reports whether Build has run.
func (l *LazyStep) Built() bool { return l.built }

// Title returns the lazy step's title before build, falling back to the inner
// step's title once built (which usually matches).
func (l *LazyStep) Title() string {
	if l.inner != nil {
		return l.inner.Title()
	}
	return l.title
}

func (l *LazyStep) Done() bool {
	if l.inner == nil {
		return false
	}
	return l.inner.Done()
}

func (l *LazyStep) Result() any {
	if l.inner == nil {
		return nil
	}
	return l.inner.Result()
}

func (l *LazyStep) Init() tea.Cmd {
	if l.inner == nil {
		return nil
	}
	return l.inner.Init()
}

func (l *LazyStep) View() string {
	if l.inner == nil {
		return ""
	}
	return l.inner.View()
}

func (l *LazyStep) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if l.inner == nil {
		return l, nil
	}
	updated, cmd := l.inner.Update(msg)
	if step, ok := updated.(Step); ok {
		l.inner = step
	}
	return l, cmd
}

// Reset clears the built flag and drops the inner step so the next entry
// rebuilds the closure against current prior results. Necessary because the
// review panel content is captured at build time — without a fresh rebuild,
// Esc-back + edit-picks would show stale content on re-entry.
func (l *LazyStep) Reset() tea.Cmd {
	l.built = false
	l.inner = nil
	return nil
}

// AutoComplete delegates to the inner step if available.
func (l *LazyStep) AutoComplete() bool {
	if l.inner == nil {
		return false
	}
	type autoChecker interface{ AutoComplete() bool }
	if a, ok := l.inner.(autoChecker); ok {
		return a.AutoComplete()
	}
	return false
}

// ─── LazyGroup ────────────────────────────────────────────────────────────

// LazyGroup is a placeholder step that, on first entry, expands into a slice
// of steps spliced into the harness at its position. Used for multi-step
// branches (e.g. "configure new agent" vs "add items to existing agent").
// The builder runs once with prior results in scope.
//
// LazyGroup satisfies Step so it can live in the declaration-time step slice,
// but its View/Update/Init are never driven — the harness splices it out
// before the user sees a frame. The Done/Result methods likewise return zero
// values because the group itself never produces a result; the steps it
// expands into do.
type LazyGroup struct {
	title   string
	build   func(prev []any) []Step
	spliced bool
}

// NewLazyGroup constructs a LazyGroup.
func NewLazyGroup(title string, build func(prev []any) []Step) *LazyGroup {
	return &LazyGroup{title: title, build: build}
}

// Splice runs the builder with prior results and returns the sub-sequence to
// splice in. The harness calls this exactly once, guarded by Spliced().
func (g *LazyGroup) Splice(prev []any) []Step {
	if g.spliced {
		return nil
	}
	g.spliced = true
	return g.build(prev)
}

// Spliced reports whether Splice has already run.
func (g *LazyGroup) Spliced() bool { return g.spliced }

// Title implements Step. Only surfaces in a breadcrumb if the harness somehow
// fails to splice before View() runs — in normal operation the user never sees
// this title.
func (g *LazyGroup) Title() string                           { return g.title }
func (g *LazyGroup) Done() bool                              { return false }
func (g *LazyGroup) Result() any                             { return nil }
func (g *LazyGroup) Init() tea.Cmd                           { return nil }
func (g *LazyGroup) View() string                            { return "" }
func (g *LazyGroup) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return g, nil }

// ─── NoteStep ─────────────────────────────────────────────────────────────

// NoteStep wraps huh.NewNote — a static information block the user advances
// past by pressing Enter. Produces no result. Used for tech-lead workspace
// info panels and "add items" intro lines.
type NoteStep struct {
	title string
	body  string
	form  *huh.Form
}

// NewNote constructs a NoteStep. title is the breadcrumb/huh title shown at
// the top of the note; body is the description paragraph rendered beneath it.
func NewNote(title, body string) *NoteStep {
	step := &NoteStep{title: title, body: body}
	step.form = step.buildForm()
	return step
}

// buildForm constructs a fresh *huh.Form. See TextStep.buildForm for why we
// rebuild instead of toggling form.State on Reset.
func (s *NoteStep) buildForm() *huh.Form {
	note := huh.NewNote().
		Title(s.title).
		Description(s.body).
		Next(true)
	return huh.NewForm(huh.NewGroup(note)).WithTheme(tui.BonsaiTheme())
}

func (s *NoteStep) Title() string { return s.title }
func (s *NoteStep) Done() bool    { return s.form.State == huh.StateCompleted }
func (s *NoteStep) Result() any   { return nil }
func (s *NoteStep) Init() tea.Cmd { return s.form.Init() }
func (s *NoteStep) View() string  { return s.form.View() }
func (s *NoteStep) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	updated, cmd := s.form.Update(msg)
	if f, ok := updated.(*huh.Form); ok {
		s.form = f
	}
	return s, cmd
}

// Reset rebuilds the form so the note renders again on re-entry via Esc-back.
func (s *NoteStep) Reset() tea.Cmd {
	s.form = s.buildForm()
	return s.form.Init()
}

// ─── SpinnerStep ──────────────────────────────────────────────────────────

// SpinnerStep displays a spinner while a blocking action runs on a worker
// goroutine. When the action returns the step completes; Result() is the
// returned error (nil on success). Ctrl-C is handled by the harness — the
// worker keeps running until the underlying call returns, so this does NOT
// fix the pre-existing "Ctrl-C during generate.* leaves partial files"
// issue at the I/O level. What it does fix is the AltScreen exit path:
// cancelling no longer leaves the terminal in spinner-frame state.
//
// Two construction shapes are exposed:
//   - NewSpinner takes a func() error; the action runs without seeing prior
//     step results.
//   - NewSpinnerWithPrior takes a func(prev []any) error; the harness invokes
//     SetPrior(prev) (priorAware hook) before Init so the action body can
//     read upstream picks captured by previous steps. Required by
//     cmd/update.go to thread per-agent custom-file selections into the
//     re-render closure.
type SpinnerStep struct {
	title    string
	label    string                 // text shown next to the spinner
	action   func() error           // simple blocking work
	actionP  func(prev []any) error // alternative: prior-aware work; one of action/actionP is nil
	sp       spinner.Model
	err      error
	done     bool
	started  bool
	initPrev []any // captured by SetPrior before Init
}

// spinnerDoneMsg is dispatched by the SpinnerStep's Init cmd once the action
// goroutine returns; the reducer flips done=true and stores the error.
type spinnerDoneMsg struct{ err error }

// newSpinnerCommon constructs a SpinnerStep with the shared spinner config.
func newSpinnerCommon(title, label string) *SpinnerStep {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(tui.ColorAccent)
	return &SpinnerStep{title: title, label: label, sp: s}
}

// NewSpinner constructs a SpinnerStep.
//   - title: breadcrumb label.
//   - label: text rendered to the right of the spinner glyph.
//   - action: the blocking work; errors are stored in Result().
func NewSpinner(title, label string, action func() error) *SpinnerStep {
	s := newSpinnerCommon(title, label)
	s.action = action
	return s
}

// NewSpinnerWithPrior constructs a SpinnerStep whose action receives the
// harness's prior-results snapshot at Init time. Use this when the action
// body needs to read upstream picks captured by earlier steps.
func NewSpinnerWithPrior(title, label string, action func(prev []any) error) *SpinnerStep {
	s := newSpinnerCommon(title, label)
	s.actionP = action
	return s
}

// SetPrior implements the priorAware hook. Called by the harness immediately
// before Init so the action closure can capture upstream results.
//
// Note on prev indexing post-splice: when a LazyGroup splices in N steps, the
// SpinnerStep that follows sees prev with length == N (the LazyGroup's own
// nil result is no longer in the list because the group was replaced). Plan
// 15 iter 3 documents this in cmd/update.go where the per-agent picker
// results land at prev[0..N-1].
func (s *SpinnerStep) SetPrior(prev []any) { s.initPrev = prev }

func (s *SpinnerStep) Title() string { return s.title }
func (s *SpinnerStep) Done() bool    { return s.done }
func (s *SpinnerStep) Result() any   { return s.err }

func (s *SpinnerStep) Init() tea.Cmd {
	s.started = true
	runner := s.action
	if s.actionP != nil {
		prev := s.initPrev
		runner = func() error { return s.actionP(prev) }
	}
	return tea.Batch(
		s.sp.Tick,
		func() tea.Msg { return spinnerDoneMsg{err: runner()} },
	)
}

func (s *SpinnerStep) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case spinnerDoneMsg:
		s.err = msg.(spinnerDoneMsg).err
		s.done = true
		return s, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		s.sp, cmd = s.sp.Update(msg)
		return s, cmd
	}
	return s, nil
}

func (s *SpinnerStep) View() string {
	return "  " + s.sp.View() + " " + tui.StyleMuted.Render(s.label)
}

// Reset is a no-op — once a SpinnerStep has run, popping back to it must NOT
// re-trigger the action (which would re-write files). The harness's Esc-back
// reset loop walks past completed spinners because AutoComplete() returns
// true once done is set.
func (s *SpinnerStep) Reset() tea.Cmd { return nil }

// AutoComplete reports true once the action has finished, so Esc-back skips
// over a completed spinner instead of trying to re-render its (gone) action.
func (s *SpinnerStep) AutoComplete() bool { return s.done }

// ─── ConditionalStep ──────────────────────────────────────────────────────

// ConditionalStep wraps another Step with a predicate. When the predicate
// returns false at the moment the harness advances onto this step, the
// wrapped step never renders — Done()=true on Init, AutoComplete()=true so
// Esc-back skips past, and Result()=nil. When the predicate returns true,
// every Step method delegates to the inner step verbatim.
//
// The predicate evaluates against prior results captured at Init time. If
// the user later Esc-backs and changes upstream picks, Reset() re-evaluates
// the predicate so the conditional re-checks correctly.
type ConditionalStep struct {
	inner     Step
	predicate func(prev []any) bool
	skip      bool  // set at Init based on predicate
	skipDone  bool  // flips true once the harness has seen Done()=true once
	initPrev  []any // prior results captured for the most recent (re-)Init
}

// NewConditional constructs a ConditionalStep.
func NewConditional(inner Step, predicate func(prev []any) bool) *ConditionalStep {
	return &ConditionalStep{inner: inner, predicate: predicate}
}

// SetPrior implements the priorAware hook. Called by the harness before Init
// so the predicate has the up-to-date prior-results snapshot.
func (c *ConditionalStep) SetPrior(prev []any) { c.initPrev = prev }

func (c *ConditionalStep) Title() string { return c.inner.Title() }

func (c *ConditionalStep) Done() bool {
	if c.skip {
		return c.skipDone
	}
	return c.inner.Done()
}

func (c *ConditionalStep) Result() any {
	if c.skip {
		return nil
	}
	return c.inner.Result()
}

func (c *ConditionalStep) Init() tea.Cmd {
	c.skip = !c.predicate(c.initPrev)
	if c.skip {
		c.skipDone = true
		return nil
	}
	// Forward prior results into the inner step too if it's prior-aware (e.g.
	// a SpinnerStep wrapped in a Conditional needs the same prev snapshot).
	if pa, ok := c.inner.(interface{ SetPrior(prev []any) }); ok {
		pa.SetPrior(c.initPrev)
	}
	return c.inner.Init()
}

func (c *ConditionalStep) View() string {
	if c.skip {
		return ""
	}
	return c.inner.View()
}

func (c *ConditionalStep) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if c.skip {
		return c, nil
	}
	updated, cmd := c.inner.Update(msg)
	if step, ok := updated.(Step); ok {
		c.inner = step
	}
	return c, cmd
}

func (c *ConditionalStep) Reset() tea.Cmd {
	c.skip = false
	c.skipDone = false
	if r, ok := c.inner.(resetter); ok {
		return r.Reset()
	}
	return nil
}

func (c *ConditionalStep) AutoComplete() bool {
	if c.skip {
		return true
	}
	type autoChecker interface{ AutoComplete() bool }
	if a, ok := c.inner.(autoChecker); ok {
		return a.AutoComplete()
	}
	return false
}

// ─── Internal helpers ─────────────────────────────────────────────────────

func valueOf(item tui.ItemOption) string {
	if item.Value != "" {
		return item.Value
	}
	return item.Name
}

// renderRequiredHeader mirrors the required/optional split rendering from
// internal/tui/prompts.go:175-198 so the visual behavior is preserved when a
// section is required-only OR mixed.
//
// onlyRequired=true means there are no optional items in the section; the
// section collapses to a single chip line summarising what's been auto-included.
// Otherwise each required item gets its own line so the user knows what's
// already in the bundle before they pick the rest.
func renderRequiredHeader(label string, required []tui.ItemOption, onlyRequired bool) string {
	var buf strings.Builder

	// Section heading (matches tui.Section).
	if label != "" {
		buf.WriteString("  " + tui.StyleLabel.Render("▸ "+label) + "\n")
	}

	if len(required) == 0 {
		return buf.String()
	}

	if onlyRequired {
		names := make([]string, len(required))
		for i, r := range required {
			names[i] = r.Name
		}
		plural := "s"
		if len(required) == 1 {
			plural = ""
		}
		head := tui.StyleSuccess.Render(tui.GlyphCheck) + " " +
			tui.StyleSand.Render(fmt.Sprintf("%d item%s auto-included", len(required), plural))
		chips := tui.StyleMuted.Render(strings.Join(names, "  "+tui.GlyphDot+"  "))
		buf.WriteString("    " + head + "\n")
		buf.WriteString("    " + chips + "\n")
		return buf.String()
	}

	for _, r := range required {
		line := "    " + tui.StyleSuccess.Render(tui.GlyphCheck) + " " + r.Name
		if r.Desc != "" {
			line += " " + tui.StyleMuted.Render(tui.GlyphDash+" "+r.Desc)
		}
		line += " " + tui.StyleAccent.Render("(required)")
		buf.WriteString(line + "\n")
	}
	return buf.String()
}
