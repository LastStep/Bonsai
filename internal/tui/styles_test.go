package tui

import (
	"io"
	"os"
	"strings"
	"testing"
)

// captureStdout redirects os.Stdout for the duration of fn and returns the
// text written. Uses os.Pipe so writes through fmt.Println are captured.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stdout = w

	done := make(chan string, 1)
	go func() {
		b, _ := io.ReadAll(r)
		done <- string(b)
	}()

	fn()

	_ = w.Close()
	os.Stdout = orig
	return <-done
}

func TestBannerIncludesAction(t *testing.T) {
	out := captureStdout(t, func() {
		Banner("0.1.3", "Initializing new project")
	})
	for _, want := range []string{"BONSAI", "agent scaffolder", "v0.1.3", "Initializing new project"} {
		if !strings.Contains(out, want) {
			t.Errorf("Banner output missing %q; got:\n%s", want, out)
		}
	}
}

func TestBannerHidesVersionWhenDev(t *testing.T) {
	out := captureStdout(t, func() {
		Banner("dev", "")
	})
	if strings.Contains(out, "vdev") {
		t.Errorf("Banner should not contain %q when version is dev; got:\n%s", "vdev", out)
	}
	// Also should not contain a bare "v" version line
	// (we check for "\nv" which would indicate a version line in our format)
	if strings.Contains(out, "\nv ") || strings.Contains(out, "\nv\n") {
		t.Errorf("Banner should not render an empty version line; got:\n%s", out)
	}
}

func TestBannerHidesActionWhenEmpty(t *testing.T) {
	out := captureStdout(t, func() {
		Banner("0.1.3", "")
	})
	// When action is empty, we do not want a dangling blank line + action
	// text. A positive check: output should not contain a two-newline gap
	// followed by any "Initializing" action hint artifact — since action is
	// empty, no marker text exists. Assert action-specific phrase absent.
	if strings.Contains(out, "Initializing") {
		t.Errorf("Banner should not contain action text when action is empty; got:\n%s", out)
	}
}

func TestItemTreeShowsCategoryCounts(t *testing.T) {
	cats := []Category{
		{Name: "Skills", Items: []string{"a", "b", "c"}},
		{Name: "Workflows", Items: []string{"w1", "w2", "w3", "w4", "w5"}},
	}
	out := ItemTree("root", cats, nil)
	if !strings.Contains(out, "Skills") || !strings.Contains(out, "(3)") {
		t.Errorf("ItemTree should show %q with count; got:\n%s", "Skills (3)", out)
	}
	if !strings.Contains(out, "Workflows") || !strings.Contains(out, "(5)") {
		t.Errorf("ItemTree should show %q with count; got:\n%s", "Workflows (5)", out)
	}
}

func TestAnswerRendersKeyValue(t *testing.T) {
	out := captureStdout(t, func() {
		Answer("Project name", "my-project")
	})
	for _, want := range []string{"Project name", "my-project"} {
		if !strings.Contains(out, want) {
			t.Errorf("Answer output missing %q; got:\n%s", want, out)
		}
	}
}

func TestAnswerShowsSkippedForEmpty(t *testing.T) {
	out := captureStdout(t, func() {
		Answer("Description", "")
	})
	if !strings.Contains(out, "(skipped)") {
		t.Errorf("Answer with empty value should render %q; got:\n%s", "(skipped)", out)
	}
}

func TestTitledPanelStringIncludesTitle(t *testing.T) {
	got := TitledPanelString("Review", "alpha\nbeta", Water)
	if !strings.Contains(got, "Review") {
		t.Errorf("TitledPanelString output missing title %q; got:\n%s", "Review", got)
	}
}

func TestTitledPanelStringMultilineBody(t *testing.T) {
	got := TitledPanelString("Review", "alpha\nbeta", Water)
	for _, want := range []string{"alpha", "beta"} {
		if !strings.Contains(got, want) {
			t.Errorf("TitledPanelString should preserve body line %q; got:\n%s", want, got)
		}
	}
}

func TestTitledPanelPrintsSameAsString(t *testing.T) {
	want := TitledPanelString("Review", "alpha\nbeta", Water) + "\n"
	got := captureStdout(t, func() {
		TitledPanel("Review", "alpha\nbeta", Water)
	})
	if got != want {
		t.Errorf("TitledPanel stdout output differs from TitledPanelString.\nwant:\n%q\ngot:\n%q", want, got)
	}
}
