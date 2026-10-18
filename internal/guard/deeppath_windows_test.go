package guard

// Windows only (8.3 short names and file links are Windows'): Bonsai's records and home reached by a short (8.3)
// spelling of their folder, as CI's runner spells its temporary folder (C:\Users\RUNNER~1\...), are refused as
// Bonsai's own; and a path through a file link to a folder (os.Symlink makes one where the target is not there yet)
// is refused as bad-path, since Windows itself cannot go through it.

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

// shortName is the folder or file p, which exists, spelled by its short (8.3) names; it skips the test where the
// volume gives p none (8.3 names off).
func shortName(t *testing.T, p string) string {
	t.Helper()
	in, err := syscall.UTF16PtrFromString(p)
	if err != nil {
		t.Fatal(err)
	}
	buf := make([]uint16, 1024)
	n, err := syscall.GetShortPathName(in, &buf[0], uint32(len(buf)))
	if err != nil || n == 0 || int(n) >= len(buf) {
		t.Skipf("no short name for %s (%v)", p, err)
	}
	s := syscall.UTF16ToString(buf[:n])
	if !strings.Contains(s, "~") {
		t.Skipf("short (8.3) names are off on this volume: %s has none", p)
	}
	return s
}

func TestBonsaisFilesByAShortNameAreRefused(t *testing.T) {
	dir := rolesProject(t, nil)
	local := filepath.Join(dir, ".bonsai", "local")
	if err := os.MkdirAll(local, 0o755); err != nil {
		t.Fatal(err)
	}
	short := shortName(t, local)
	home := shortName(t, filepath.Join(filepath.Dir(dir), "home"))
	// A Write by the main session: into the records, a folder not made yet in them, and the home.
	for _, p := range []string{filepath.Join(short, "x.ndjson"), filepath.Join(short, "new", "x.json"), filepath.Join(home, "x.json")} {
		wantRule(t, dir, writeOfAbs(dir, p, "x\n"), RuleBonsaiFiles)
	}
	// A shell write by a lead, past the old limit too.
	far := filepath.ToSlash(filepath.Join(deepUnder(short, 70), "x"))
	for i, line := range []string{"echo x > " + filepath.ToSlash(filepath.Join(short, "x")), "printf x | tee " + far,
		"cp free.txt " + filepath.ToSlash(filepath.Join(home, "x"))} {
		started(t, dir, sessA, "workflow:lead")
		wantAnswer(t, dir, shellAs(dir, "s"+string(rune('a'+i)), "workflow:lead", line, false), 2, RuleBonsaiFiles)
	}
}

// os.Symlink makes a file link where its target is not there yet; Windows cannot go through a file link to a folder
// made later, so neither can a write, and the guard, which cannot open it, refuses a path through it as bad-path.
func TestAFileLinkToAFolderIsBadPath(t *testing.T) {
	dir := rolesProject(t, nil)
	away, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	target, link := filepath.Join(away, "later"), filepath.Join(away, "rec")
	guardSymlink(t, target, link)
	if err := os.Mkdir(target, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(link, "probe.txt"), []byte("x"), 0o644); err == nil {
		t.Fatalf("Windows wrote through a file link to a folder made later: the guard's bad-path here would be wrong")
	}
	wantRule(t, dir, writeOfAbs(dir, filepath.Join(link, "x.txt"), "x\n"), RuleBadPath)
	started(t, dir, sessA, "workflow:lead")
	wantAnswer(t, dir, shellAs(dir, "fa", "workflow:lead", "echo x > "+filepath.ToSlash(filepath.Join(link, "x.txt")), false), 2, RuleBadPath)
}
