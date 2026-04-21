package initflow

import (
	"testing"
)

// TestWideCharSafe_OptOut verifies BONSAI_ASCII_ONLY=1 forces WideCharSafe
// to report false regardless of the rest of the environment.
func TestWideCharSafe_OptOut(t *testing.T) {
	t.Setenv("BONSAI_ASCII_ONLY", "1")
	if WideCharSafe() {
		t.Fatalf("BONSAI_ASCII_ONLY=1 must force WideCharSafe=false")
	}
}

// TestStageLabel_RenderASCIIFallback verifies that with safe=false StageLabel
// renders an English-only primary and an empty secondary — no kanji / kana
// leaks into the output path on constrained terminals.
func TestStageLabel_RenderASCIIFallback(t *testing.T) {
	t.Setenv("BONSAI_ASCII_ONLY", "1")
	// WideCharSafe should now report false; Render mirrors that contract.
	safe := WideCharSafe()
	if safe {
		t.Fatalf("WideCharSafe should be false under BONSAI_ASCII_ONLY=1")
	}

	label := StageLabels[2] // Branches
	primary, secondary := label.Render(safe)

	if primary != "BRANCHES" {
		t.Fatalf("ascii primary = %q, want %q", primary, "BRANCHES")
	}
	if secondary != "" {
		t.Fatalf("ascii secondary = %q, want empty string", secondary)
	}
	// Confirm no kanji/kana slipped through.
	if containsAny(primary, "枝えだ") {
		t.Fatalf("ascii primary must not contain kanji/kana: %q", primary)
	}
}

// TestStageLabel_RenderSafe verifies the happy path: safe=true yields
// "<kanji> <English>" + kana secondary.
func TestStageLabel_RenderSafe(t *testing.T) {
	label := StageLabels[0] // Vessel
	primary, secondary := label.Render(true)
	if primary != "器 VESSEL" {
		t.Fatalf("safe primary = %q, want %q", primary, "器 VESSEL")
	}
	if secondary != "うつわ" {
		t.Fatalf("safe secondary = %q, want %q", secondary, "うつわ")
	}
}

// containsAny reports whether s contains any rune from chars.
func containsAny(s, chars string) bool {
	for _, c := range chars {
		for _, r := range s {
			if r == c {
				return true
			}
		}
	}
	return false
}
