package initflow

import (
	"strings"
	"testing"
)

// TestRenderHeader_CollapsesHome verifies that a project path rooted under
// $HOME is rendered with the tilde-collapsed prefix (~/...) rather than the
// full absolute path.
func TestRenderHeader_CollapsesHome(t *testing.T) {
	t.Setenv("HOME", "/home/alice")

	out := RenderHeader("0.1.2", "/home/alice/voyager-api", "INIT", "PLANTING INTO", 120, true)

	if !strings.Contains(out, "~/voyager-api") {
		t.Fatalf("expected tilde-collapsed project path in header, got:\n%s", out)
	}
	// The project name itself must appear.
	if !strings.Contains(out, "voyager-api") {
		t.Fatalf("expected project name to appear in header, got:\n%s", out)
	}
}

// TestRenderHeader_AbsolutePathOutsideHome verifies that a project path that
// is not under $HOME is rendered verbatim (no spurious ~ substitution).
func TestRenderHeader_AbsolutePathOutsideHome(t *testing.T) {
	t.Setenv("HOME", "/home/bob")

	out := RenderHeader("0.1.2", "/tmp/p", "INIT", "PLANTING INTO", 120, true)

	if !strings.Contains(out, "/tmp/p") {
		t.Fatalf("expected absolute project path in header, got:\n%s", out)
	}
	if strings.Contains(out, "~/") {
		t.Fatalf("expected no tilde substitution for path outside HOME, got:\n%s", out)
	}
}

// TestRenderHeader_NoStationSegment is the regression guard for the Phase-3
// bug fix — the station subdir doesn't exist until Phase 5 generate runs, so
// the header must not claim "station/" in its path row. Covers both safe and
// ASCII-fallback rendering modes.
func TestRenderHeader_NoStationSegment(t *testing.T) {
	t.Setenv("HOME", "/home/alice")

	for _, safe := range []bool{true, false} {
		out := RenderHeader("0.1.2", "/home/alice/voyager-api", "INIT", "PLANTING INTO", 120, safe)
		if strings.Contains(out, "station") {
			t.Fatalf("safe=%v: header must not contain \"station\" substring, got:\n%s", safe, out)
		}
	}
}

// TestRenderHeader_TrailingSlash verifies the project row renders a trailing
// slash after the project name so the path reads as a directory.
func TestRenderHeader_TrailingSlash(t *testing.T) {
	t.Setenv("HOME", "/home/alice")

	out := RenderHeader("0.1.2", "/home/alice/voyager-api", "INIT", "PLANTING INTO", 120, true)

	if !strings.Contains(out, "voyager-api/") {
		t.Fatalf("expected trailing slash after project name, got:\n%s", out)
	}
}

// TestRenderHeader_CustomAction covers the Plan 28 Phase 1 signature
// extension — the `action` parameter is rendered on the left block row 2
// instead of the hardcoded "INIT" literal.
func TestRenderHeader_CustomAction(t *testing.T) {
	t.Setenv("HOME", "/home/alice")

	out := RenderHeader("0.1.2", "/home/alice/voyager-api", "CATALOG", "PLANTING INTO", 120, true)

	if !strings.Contains(out, "CATALOG") {
		t.Fatalf("expected custom action label in header, got:\n%s", out)
	}
	// INIT must not leak in when a different action is supplied.
	if strings.Contains(out, "INIT") {
		t.Fatalf("custom action did not replace default INIT, got:\n%s", out)
	}
}

// TestRenderHeader_EmptyRightLabelHidesRow1 verifies that an empty rightLabel
// hides the right-block row 1 entirely. The project path still renders on
// row 2, but the destination preamble (e.g. "PLANTING INTO") is omitted.
func TestRenderHeader_EmptyRightLabelHidesRow1(t *testing.T) {
	t.Setenv("HOME", "/home/alice")

	out := RenderHeader("0.1.2", "/home/alice/voyager-api", "CATALOG", "", 120, true)

	if strings.Contains(out, "PLANTING INTO") {
		t.Fatalf("expected right-block row 1 to be hidden, got:\n%s", out)
	}
	// Project path must still render.
	if !strings.Contains(out, "voyager-api") {
		t.Fatalf("expected project name to still render on row 2, got:\n%s", out)
	}
}
