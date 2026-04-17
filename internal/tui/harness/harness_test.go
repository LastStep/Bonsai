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
