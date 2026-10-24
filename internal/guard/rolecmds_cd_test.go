package guard

// BUG-108 (T-5737): a role that changes nothing writes a relative path after a cd the guard cannot follow (cd $S, S a
// folder outside the project). The guard still judges the path in the folder the command started in, so it is still
// refused; its words say that it could not follow the cd, and its next step is to name the file by its full path,
// outside the project. A cd the guard follows, or none, keeps the words it had.

import (
	"path/filepath"
	"strings"
	"testing"
)

// The trial's two shapes (LinkSink, rc.2: a researcher's and a verifier's scratch writes), and others like them.
func TestRoleWriteAfterACdNotFollowed(t *testing.T) {
	dir := rolesProject(t, nil)
	started(t, dir, sessA, "workflow:helper")
	lost := []string{
		"cd $S && cat > b.py <<'EOF'\nprint(1)\nEOF",
		"cd $S/probe && node -e 'console.log(1)' > probe.mjs",
		"cd \"$(mktemp -d)\" && echo x > a.txt",
		"cd - && echo x >> log.txt",
		"cd $S && cp free.txt copy.txt",
	}
	for _, line := range lost {
		s := wantAnswer(t, dir, shellAs(dir, "hx", "workflow:helper", line, false), 2, RuleRoleWritesNothing)
		for _, want := range []string{"could not tell where a cd before it went", "full path, outside the project"} {
			if !strings.Contains(s, want) {
				t.Errorf("%q: the refusal does not say %q:\n%s", line, want, s)
			}
		}
		if strings.Contains(s, "write scratch files outside the project") {
			t.Errorf("%q: the old next step, which does not fit after a cd:\n%s", line, s)
		}
	}
	// The next step it names goes through: the same write by its full path, outside the project.
	out := filepath.ToSlash(t.TempDir())
	for _, line := range []string{"cat > " + out + "/b.py <<'EOF'\nprint(1)\nEOF", "node -e 'console.log(1)' > " + out + "/probe.mjs",
		"cd $S && cat > " + out + "/b.py <<'EOF'\nprint(1)\nEOF"} {
		wantAnswer(t, dir, shellAs(dir, "hx", "workflow:helper", line, false), 0, "")
	}
}

// A cd the guard follows, or none: the words it had.
func TestRoleWriteAfterACdFollowed(t *testing.T) {
	dir := rolesProject(t, nil)
	started(t, dir, sessA, "workflow:helper")
	for _, line := range []string{"echo x > notes.txt", "cd sub && echo x > y.txt", "cd " + filepath.ToSlash(dir) + " && echo x > z.txt",
		"cat > b.py <<'EOF'\nprint(1)\nEOF", "cd sub && cp free.txt copy.txt"} {
		s := wantAnswer(t, dir, shellAs(dir, "hx", "workflow:helper", line, false), 2, RuleRoleWritesNothing)
		if !strings.Contains(s, "inside a linked project") || !strings.Contains(s, "write scratch files outside the project") {
			t.Errorf("%q: not today's words:\n%s", line, s)
		}
		if strings.Contains(s, "could not tell where a cd") {
			t.Errorf("%q: says it could not follow a cd it followed:\n%s", line, s)
		}
	}
}
