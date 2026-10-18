package guard

// A path the guard cannot walk to its end is refused, never judged as written (V1's B1 on 5.6, M-2026-10-18-8ba15229,
// and the audit M-2026-10-18-7704b238): a link out of the project, then more folders not made yet than the old 64-step
// limit, into a protected folder; a link loop; and a shell write target whose forms cannot be read (a too-deep path,
// a path into /proc), refused bad-path, not allowed. The links are symbolic links, which need a privilege on Windows:
// each case skips where none can be made, and TestDeepPathString's half runs everywhere.

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/schema"
)

// guardSymlink skips the test where no symbolic link can be made (on Windows it needs a privilege).
func guardSymlink(t *testing.T, target, link string) {
	t.Helper()
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("no link can be made here (on Windows it needs a privilege): %v", err)
	}
}

// deepUnder is base with n folders below it, none made.
func deepUnder(base string, n int) string {
	parts := []string{base}
	for i := 0; i < n; i++ {
		parts = append(parts, "d")
	}
	return filepath.Join(parts...)
}

func TestPathFormsPastTheOldLimit(t *testing.T) {
	tmp, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	inner := filepath.Join(tmp, "inner")
	if err := os.Mkdir(inner, 0o755); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(tmp, "link")
	guardSymlink(t, inner, link)
	for _, n := range []int{60, 64, 70, 200} {
		forms, err := pathForms(tmp, tmp, deepUnder(link, n))
		if err != nil {
			t.Fatalf("%d folders behind a link: %v", n, err)
		}
		want := deepUnder(inner, n)
		found := false
		for _, f := range forms {
			found = found || f == want
		}
		if !found {
			t.Errorf("%d folders behind a link: forms %q lack the link's target", n, forms)
		}
	}
	// A loop: refused.
	guardSymlink(t, filepath.Join(tmp, "b"), filepath.Join(tmp, "a"))
	guardSymlink(t, filepath.Join(tmp, "a"), filepath.Join(tmp, "b"))
	for _, p := range []string{filepath.Join(tmp, "a"), deepUnder(filepath.Join(tmp, "a"), 70)} {
		if forms, err := pathForms(tmp, tmp, p); !errors.Is(err, errUnwalked) {
			t.Errorf("%s through a link loop: %q, %v; want errUnwalked", p, forms, err)
		}
	}
}

// A Write through a link outside the project, past the old limit, into a protected folder: refused, by the target's
// rule (the link is followed) or as bad-path, never allowed as a path outside every project.
func TestADeepWriteThroughALinkIsJudgedByItsTarget(t *testing.T) {
	dir := withTask(t, "running")
	away, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(away, "L")
	guardSymlink(t, filepath.Join(dir, "keys"), link)
	for _, n := range []int{10, 70} {
		code, stderr := answer(t, dir, writeOfAbs(dir, filepath.Join(deepUnder(link, n), "x.key"), "x\n"))
		if code != 2 {
			t.Errorf("a Write through a link to keys/, then %d folders: exit %d, %s", n, code, stderr)
		}
	}
	// A link loop on the path: bad-path.
	guardSymlink(t, filepath.Join(away, "b"), filepath.Join(away, "a"))
	guardSymlink(t, filepath.Join(away, "a"), filepath.Join(away, "b"))
	wantRule(t, dir, writeOfAbs(dir, filepath.Join(away, "a", "x.txt"), "x\n"), RuleBadPath)
}

// writeOfAbs is a Write of content at the absolute path p, from a session in dir.
func writeOfAbs(dir, p, content string) string {
	return toolPayload("r1", dir, "Write", schema.Object{{Key: "file_path", Value: p}, {Key: "content", Value: content}})
}

// An agent's shell write whose target cannot be walked (a link loop, a path past the limit into Bonsai's records)
// is refused, not allowed: bad-path from a role that writes, role-writes-nothing from one that changes nothing (as
// before); so is one into /proc (Linux), which a role that writes was let through before for the same reason.
func TestAShellWriteTheGuardCannotPlaceIsRefused(t *testing.T) {
	dir := rolesProject(t, nil)
	away, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	guardSymlink(t, filepath.Join(away, "b"), filepath.Join(away, "a"))
	guardSymlink(t, filepath.Join(away, "a"), filepath.Join(away, "b"))
	loop := filepath.ToSlash(filepath.Join(away, "a", "x"))
	// A link to the main checkout's .bonsai/local, then 70 folders: Bonsai's own records, reached past the old limit.
	guardSymlink(t, filepath.Join(dir, ".bonsai", "local"), filepath.Join(away, "rec"))
	far := filepath.ToSlash(filepath.Join(deepUnder(filepath.Join(away, "rec"), 70), "x"))
	lines := []struct{ line, rule string }{
		{"echo x > " + loop, RuleBadPath}, {"printf x | tee " + loop, RuleBadPath}, {"cp free.txt " + loop, RuleBadPath},
		{"chmod 000 " + loop, RuleBadPath},
		{"echo x > " + far, RuleBonsaiFiles}, {"printf x | tee " + far, RuleBonsaiFiles}, {"cp free.txt " + far, RuleBonsaiFiles},
		{"chmod 000 " + far, RuleBonsaiFiles},
	}
	if runtime.GOOS == "linux" {
		lines = append(lines, struct{ line, rule string }{"echo x > /proc/self/cwd/.bonsai/local/x", RuleBadPath},
			struct{ line, rule string }{"chmod 000 /proc/self/cwd/.bonsai/lock.json", RuleBadPath})
	}
	for i, c := range lines {
		started(t, dir, sessA, "workflow:lead") // each a fresh lead, joined to its call: a role that writes
		s := wantAnswer(t, dir, shellAs(dir, "w"+string(rune('a'+i)), "workflow:lead", c.line, false), 2, c.rule)
		if c.rule == RuleBadPath && runtime.GOOS != "windows" && strings.Contains(s, "on Windows") {
			t.Errorf("%s: the refusal names Windows: %s", c.line, s)
		}
	}
	// A role that changes nothing keeps its own refusal.
	started(t, dir, sessA, "workflow:helper")
	wantAnswer(t, dir, shellAs(dir, "hz", "workflow:helper", "echo x > "+loop, false), 2, RuleRoleWritesNothing)
	// A plain path outside the project is still allowed.
	started(t, dir, sessA, "workflow:lead")
	wantAnswer(t, dir, shellAs(dir, "wz", "workflow:lead", "echo x > "+filepath.ToSlash(filepath.Join(away, "plain.txt")), false), 0, "")
}
