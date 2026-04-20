package harness

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"

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
