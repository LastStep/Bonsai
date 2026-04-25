// Package wsvalidate centralises workspace-path validation rules used by
// addflow + initflow + cmd. Both flows previously duplicated the
// trim/Clean/trailing-slash + IsAbs + ".." segment scan; this package is
// the single source of truth so adding a defence (backslash, pure-root)
// updates every caller at once.
//
// The package depends only on stdlib path/filepath + strings to keep it
// importable from any TUI flow without cycles.
package wsvalidate

import (
	"path/filepath"
	"strings"
)

// Normalise applies the shared trim + filepath.Clean + trailing-slash rule
// used to canonicalise workspace strings before validation or storage.
// Empty input returns "".
func Normalise(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return ""
	}
	v = strings.TrimRight(filepath.Clean(v), "/") + "/"
	return v
}

// InvalidReason returns a user-facing error string when the normalised
// workspace escapes the project root, is absolute, contains a backslash,
// or reduces to the project root itself. Returns "" when the workspace
// is a safe project-relative subdirectory. Called after Normalise has
// cleaned the input.
//
// Project-relative only. Defence against accidental writes outside the
// project root when the user types "../..." or a rooted path. Not an
// adversarial boundary — the user already has write access to their own
// filesystem — but prevents silent surprises in test harnesses and
// dogfooding sessions.
func InvalidReason(ws string) string {
	// filepath.IsAbs catches "/etc/" on POSIX and "C:\..." on Windows.
	if filepath.IsAbs(ws) {
		return "absolute paths not allowed (no leading / or drive letter)"
	}
	// Reject backslash explicitly — POSIX treats `\` as a legal literal in
	// file names, so "foo\bar" sneaks past IsAbs but is almost certainly a
	// user error (Windows-style separator) we don't want to silently accept.
	if strings.ContainsRune(ws, '\\') {
		return "backslash not allowed (use forward slash)"
	}
	// After filepath.Clean, any remaining ".." component means the path
	// escapes the project root. Split on "/" (Normalise always emits
	// forward slashes) and check each segment.
	for _, seg := range strings.Split(strings.TrimRight(ws, "/"), "/") {
		if seg == ".." {
			return "workspace must not escape project root (no ..)"
		}
	}
	// Reject pure root — input that Cleans to "." (e.g. "", ".", "foo/..")
	// would install at the project root, which is almost never intended.
	// After Normalise the value is "./", so the trimmed form is ".".
	if strings.TrimRight(ws, "/") == "." {
		return "workspace cannot be project root"
	}
	return ""
}
