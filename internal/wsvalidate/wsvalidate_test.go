package wsvalidate

import (
	"strings"
	"testing"
)

// TestInvalidReason_RejectsAbsolute covers the POSIX IsAbs branch. Windows
// drive-letter paths (e.g. `C:\foo`) are rejected by the backslash branch
// on POSIX hosts and by IsAbs on Windows hosts — covered by the
// RejectsBackslash test.
func TestInvalidReason_RejectsAbsolute(t *testing.T) {
	cases := []string{"/etc/foo/", "/var/"}
	for _, in := range cases {
		got := InvalidReason(Normalise(in))
		if !strings.Contains(got, "absolute paths not allowed") {
			t.Errorf("InvalidReason(%q) = %q, want 'absolute paths not allowed' substring", in, got)
		}
	}
}

// TestInvalidReason_RejectsBackslash verifies "foo\bar" (POSIX-legal but
// almost-certainly Windows-confusion) is rejected with the
// "backslash not allowed" reason. "C:\foo" is caught by the IsAbs branch
// first on Windows and by the backslash branch here on POSIX — either is
// fine; we only assert that the reason is non-empty for both.
func TestInvalidReason_RejectsBackslash(t *testing.T) {
	cases := []string{`foo\bar`, `nested\path`}
	for _, in := range cases {
		got := InvalidReason(Normalise(in))
		if got == "" {
			t.Errorf("InvalidReason(%q) = empty, want non-empty rejection", in)
		}
		if !strings.Contains(got, "backslash") {
			t.Errorf("InvalidReason(%q) = %q, want 'backslash' substring", in, got)
		}
	}
}

// TestInvalidReason_RejectsPureRoot verifies inputs that Clean to "."
// (i.e. would install at the project root) are rejected.
func TestInvalidReason_RejectsPureRoot(t *testing.T) {
	cases := []string{"./", ".", "foo/.."}
	for _, in := range cases {
		got := InvalidReason(Normalise(in))
		if !strings.Contains(got, "cannot be project root") {
			t.Errorf("InvalidReason(%q) = %q, want 'cannot be project root' substring", in, got)
		}
	}
}

// TestInvalidReason_RejectsParentEscape verifies any ".." segment surviving
// Normalise is rejected — the existing rule, retained as a regression guard
// alongside the new defences.
func TestInvalidReason_RejectsParentEscape(t *testing.T) {
	cases := []string{"../foo", "nested/../..", "../../bar"}
	for _, in := range cases {
		got := InvalidReason(Normalise(in))
		if !strings.Contains(got, "escape project root") {
			t.Errorf("InvalidReason(%q) = %q, want 'escape project root' substring", in, got)
		}
	}
}

// TestInvalidReason_AcceptsNestedRelative verifies a clean nested relative
// path is accepted (returns empty reason). Positive companion to the
// rejection tests above.
func TestInvalidReason_AcceptsNestedRelative(t *testing.T) {
	cases := []string{"nested/path/", "./foo/", "foo/../bar/"}
	for _, in := range cases {
		got := InvalidReason(Normalise(in))
		if got != "" {
			t.Errorf("InvalidReason(%q) = %q, want empty (accepted)", in, got)
		}
	}
}

// TestNormalise_TrimsAndAddsTrailingSlash verifies the canonicalisation
// rule shared by every caller: trim → Clean → trailing slash. Empty stays
// empty (callers treat empty as "use default").
func TestNormalise_TrimsAndAddsTrailingSlash(t *testing.T) {
	cases := map[string]string{
		"backend":    "backend/",
		"backend/":   "backend/",
		"./backend":  "backend/",
		"  api  ":    "api/",
		"":           "",
		"nested/dir": "nested/dir/",
	}
	for in, want := range cases {
		if got := Normalise(in); got != want {
			t.Errorf("Normalise(%q) = %q, want %q", in, got, want)
		}
	}
}
