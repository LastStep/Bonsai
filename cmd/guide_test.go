package cmd

import (
	"strings"
	"testing"
)

// fakeGuides populates the package-level guideContents map with
// minimal markdown for each canonical topic so runGuide-adjacent
// tests can validate lookups and the static-render path.
func fakeGuides() map[string]string {
	return map[string]string{
		"quickstart":   "# Quickstart\n\nstart here.\n",
		"concepts":     "# Concepts\n\nthe mental model.\n",
		"cli":          "# CLI\n\ncommand reference.\n",
		"custom-files": "# Custom Files\n\nadd your own.\n",
	}
}

// withGuides swaps in the fake content map for the duration of fn
// and restores the original on completion so tests don't leak
// state between runs.
func withGuides(t *testing.T, fn func()) {
	t.Helper()
	prev := guideContents
	guideContents = fakeGuides()
	defer func() { guideContents = prev }()
	fn()
}

// TestRenderStatic_ProducesNonEmptyOutput verifies the non-TTY
// static render path emits glamour output to stdout.
func TestRenderStatic_ProducesNonEmptyOutput(t *testing.T) {
	withGuides(t, func() {
		out := captureStdout(t, func() {
			if err := renderStatic(guideContents["quickstart"]); err != nil {
				t.Fatalf("renderStatic: %v", err)
			}
		})
		if strings.TrimSpace(out) == "" {
			t.Fatalf("renderStatic produced empty output")
		}
		// The glamour-rendered output should contain some form of
		// the "Quickstart" heading text (glamour may reformat it
		// with styling glyphs, but the word survives).
		if !strings.Contains(out, "Quickstart") {
			t.Fatalf("renderStatic output missing 'Quickstart' heading:\n%s", out)
		}
	})
}

// TestRenderStatic_StripsFrontmatter verifies the static path
// strips YAML frontmatter before handing to glamour, preserving
// the pre-Plan-28 behavior.
func TestRenderStatic_StripsFrontmatter(t *testing.T) {
	withGuides(t, func() {
		content := "---\ndescription: x\n---\n# Heading\n\nbody\n"
		out := captureStdout(t, func() {
			if err := renderStatic(content); err != nil {
				t.Fatalf("renderStatic: %v", err)
			}
		})
		if strings.Contains(out, "description:") {
			t.Fatalf("renderStatic leaked frontmatter:\n%s", out)
		}
	})
}

// TestRunGuide_UnknownTopicErrors verifies passing an invalid topic
// key produces the pre-existing error message (unchanged by the
// Plan 28 Phase 3 rewire).
func TestRunGuide_UnknownTopicErrors(t *testing.T) {
	withGuides(t, func() {
		err := runGuide(guideCmd, []string{"bogus-topic"})
		if err == nil {
			t.Fatalf("expected error for unknown topic; got nil")
		}
		if !strings.Contains(err.Error(), "unknown topic") {
			t.Fatalf("expected 'unknown topic' in error; got: %v", err)
		}
	})
}

// TestRunGuide_NoArgNonTTYExactMessage verifies the decision-D4
// error: running `bonsai guide` with no arg while stdout is not
// a TTY must produce the exact canonical message.
//
// Relies on the fact that `go test` redirects stdout into the
// test harness (not a TTY), so isatty.IsTerminal returns false
// without any mocking required.
func TestRunGuide_NoArgNonTTYExactMessage(t *testing.T) {
	withGuides(t, func() {
		err := runGuide(guideCmd, nil)
		if err == nil {
			t.Fatalf("expected error for no-arg non-TTY invocation; got nil")
		}
		want := "bonsai guide: specify a topic when piping output (quickstart, concepts, cli, custom-files)"
		if err.Error() != want {
			t.Fatalf("error mismatch:\n got=%q\nwant=%q", err.Error(), want)
		}
	})
}

// TestRunGuide_ArgNonTTYRendersStatic verifies passing a valid
// topic with stdout-not-a-TTY (the `go test` default) falls
// through to the static glamour render — no BubbleTea program
// is started, so the test returns cleanly and stdout contains
// the rendered markdown.
func TestRunGuide_ArgNonTTYRendersStatic(t *testing.T) {
	withGuides(t, func() {
		var runErr error
		out := captureStdout(t, func() {
			runErr = runGuide(guideCmd, []string{"concepts"})
		})
		if runErr != nil {
			t.Fatalf("runGuide with valid topic: %v", runErr)
		}
		if !strings.Contains(out, "Concepts") {
			t.Fatalf("runGuide concepts output missing 'Concepts':\n%s", out)
		}
	})
}
