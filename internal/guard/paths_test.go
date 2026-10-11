package guard

// Where the guard reads from (design/plan-5.md, 5.3.1; "5.3 done" checks 6 and 7): the projects holding a path, past
// any .git; the main checkout verified, else no grant and main-not-verified naming the person's step; a worktree's
// lists from its main checkout's bonsai.yaml; the records in the local/ of the project whose rules decided (a verified
// worktree's main's), which the cleaner reads; the admin share's forms. The worktree cases are made with real git (on
// Windows, Windows git), as bonsai init and git worktree make them.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/clean"
	"github.com/LastStep/Bonsai/internal/testpack"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// writeFile makes a file and its folders.
func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// run is one guard run through Main in a session whose project folder is dir, its exit code and stderr.
func runIn(t *testing.T, dir, session, path string) (int, string) {
	t.Helper()
	return guardRun(t, env(dir), strings.NewReader(payload(session, dir, "Write", "file_path", path)), untimed)
}

// recordsIn reads a session's records in the log of the checkout at main, or none when it has no such file.
func recordsIn(t *testing.T, main, session string) []string {
	t.Helper()
	if _, err := os.Stat(filepath.Join(main, filepath.FromSlash(LogDir), "s-"+session+".ndjson")); err != nil {
		return nil
	}
	var out []string
	for _, r := range records(t, main, "s-"+session+".ndjson") {
		out = append(out, r.String("decision")+" "+r.String("rule")+" "+r.String("target")+" "+r.String("checkout")+" "+r.String("workspace"))
	}
	return out
}

// The projects holding a path (5.3.1 note 1): a .git in the folder the session started in, a nested repository, a
// .git or a bonsai.yaml with empty lists placed in a subfolder, or both: the project around them still judges, and
// its record goes to its own log, never to a folder planted inside it.
func TestTheProjectsHoldingAPath(t *testing.T) {
	dir := project(t, testYAML)
	writeFile(t, filepath.Join(dir, "docs", "guide.md"), "guide\n")
	emptyLists := "format: bonsai.workspace/1\nid: ws-pppppppppppppppppppppppppp\nname: planted\npacks: []\nprotected: []\nperson_only: []\nnever_edit: []\n"

	// A .git in the folder the session started in: the session is still the project's.
	sub := filepath.Join(dir, "sub")
	writeFile(t, filepath.Join(sub, ".git", "HEAD"), "ref: refs/heads/x\n")
	if code, stderr := runIn(t, sub, "p1", filepath.Join(dir, "protected.txt")); code != 2 || !strings.Contains(stderr, `"protected.txt" refused`) {
		t.Errorf("a .git in the starting folder: exit %d, %s", code, stderr)
	}
	// A nested repository with no bonsai.yaml: its paths are the project's.
	writeFile(t, filepath.Join(dir, "docs", "vendor", ".git", "HEAD"), "ref: refs/heads/main\n")
	if code, _ := runIn(t, dir, "p2", filepath.Join(dir, "docs", "vendor", "lib.go")); code != 2 {
		t.Errorf("a nested repository inside a protected folder: exit %d", code)
	}
	// A bonsai.yaml with empty lists placed in docs/, then a .git beside it: neither narrows the project's lists.
	writeFile(t, filepath.Join(dir, "docs", "bonsai.yaml"), emptyLists)
	code, stderr := runIn(t, dir, "p3", filepath.Join(dir, "docs", "guide.md"))
	if code != 2 || !strings.Contains(stderr, `"docs/guide.md" (in the project`) {
		t.Errorf("a planted bonsai.yaml: exit %d, %s", code, stderr)
	}
	writeFile(t, filepath.Join(dir, "docs", ".git", "HEAD"), "ref: refs/heads/main\n")
	if code, _ := runIn(t, dir, "p3", filepath.Join(dir, "docs", "guide.md")); code != 2 {
		t.Errorf("a planted bonsai.yaml and .git: exit %d", code)
	}
	// Their records are the project's: in its log, under its id, never in docs/.
	got := recordsIn(t, dir, "p3")
	if len(got) != 2 || !strings.HasPrefix(got[0], "deny protected docs/guide.md "+filepath.Base(dir)+" "+testID) {
		t.Errorf("the records: %q", got)
	}
	if _, err := os.Stat(filepath.Join(dir, "docs", ".bonsai")); err == nil {
		t.Error("a record was written into the planted project")
	}
	// The planted lists add to the project's: a path on them alone is refused too, as main-not-verified (the planted
	// folder is no verified checkout, and the project around it grants from its own task files only).
	writeFile(t, filepath.Join(dir, "docs", "bonsai.yaml"), strings.Replace(emptyLists, "protected: []", `protected: ["notes.md"]`, 1))
	if code, stderr := runIn(t, dir, "p4", filepath.Join(dir, "docs", "notes.md")); code != 2 || !strings.Contains(stderr, `"notes.md" (in the project "docs")`) {
		t.Errorf("a planted list: exit %d, %s", code, stderr)
	}
	// A free path beside them stays free.
	if code, stderr := runIn(t, dir, "p4", filepath.Join(dir, "free.txt")); code != 0 {
		t.Errorf("a free path: exit %d, %s", code, stderr)
	}
}

// A path in another linked project is judged by that project's lists, and recorded in its log under its id; a path
// in no project is allowed and recorded in the session's; a session in no linked project still has a linked
// project's path judged (and recorded there), and records nothing else.
func TestAPathInAnotherProject(t *testing.T) {
	home := t.TempDir()
	t.Setenv(workspace.HomeEnv, home)
	const otherID = "ws-bbbbbbbbbbbbbbbbbbbbbbbbbb"
	a := unrecordedProject(t, testYAML)
	b := unrecordedProject(t, strings.Replace(strings.Replace(testYAML, testID, otherID, 1), `"protected.txt"`, `"b-secret.txt"`, 1))
	for root, id := range map[string]string{a: testID, b: otherID} {
		if _, err := workspace.RecordCheckout(home, root, id, time.Now()); err != nil {
			t.Fatal(err)
		}
	}
	if code, stderr := runIn(t, a, "x1", filepath.Join(b, "b-secret.txt")); code != 2 || !strings.Contains(stderr, `"b-secret.txt" refused: it is on bonsai.yaml's protected list`) {
		t.Errorf("another project's protected path: exit %d, %s", code, stderr)
	}
	if code, _ := runIn(t, a, "x1", filepath.Join(b, "protected.txt")); code != 0 {
		t.Errorf("a path free in its own project (protected only in the session's): exit %d", code)
	}
	if got := recordsIn(t, b, "x1"); len(got) != 2 || got[0] != "deny protected b-secret.txt "+filepath.Base(b)+" "+otherID ||
		got[1] != "allow not-protected protected.txt "+filepath.Base(b)+" "+otherID {
		t.Errorf("records in the other project: %q", got)
	}
	if got := recordsIn(t, a, "x1"); got != nil {
		t.Errorf("records in the session's project: %q", got)
	}
	outside := filepath.Join(t.TempDir(), "elsewhere.txt")
	if code, _ := runIn(t, a, "x2", outside); code != 0 {
		t.Errorf("a path in no project: exit %d", code)
	}
	if got := recordsIn(t, a, "x2"); len(got) != 1 || !strings.HasPrefix(got[0], "allow outside-project  ") {
		t.Errorf("a path in no project: %q", got)
	}
	// A session in no linked project.
	plain, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if code, _ := runIn(t, plain, "x3", filepath.Join(b, "b-secret.txt")); code != 2 {
		t.Errorf("an unlinked session, a linked project's protected path: exit %d", code)
	}
	if code, _ := runIn(t, plain, "x3", filepath.Join(plain, "anything.txt")); code != 0 {
		t.Errorf("an unlinked session, its own path: exit %d", code)
	}
	if got := recordsIn(t, b, "x3"); len(got) != 1 {
		t.Errorf("an unlinked session's records in the linked project: %q", got)
	}
	if _, err := os.Stat(filepath.Join(plain, ".bonsai")); err == nil {
		t.Error("an unlinked session's folder got a record")
	}
}

// gitProject makes a project with git, as bonsai init leaves one: bonsai.yaml committed, recorded on this machine at
// since (unless it is zero), and a worktree of it beside it. It gives the main checkout and the worktree.
func gitProject(t *testing.T, home, yaml string, since time.Time) (main, wt string) {
	t.Helper()
	tmp, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("GIT_CEILING_DIRECTORIES", tmp)
	main = testpack.Project(t, tmp, "main")
	writeFile(t, filepath.Join(main, "bonsai.yaml"), yaml)
	writeFile(t, filepath.Join(main, "protected.txt"), "protected\n")
	testpack.Git(t, main, "add", "-A")
	testpack.Git(t, main, "commit", "-q", "-m", "linked")
	wt = filepath.Join(tmp, "wt")
	testpack.Git(t, main, "worktree", "add", "-q", "-b", "t0901", wt)
	if !since.IsZero() {
		if _, err := workspace.RecordCheckout(home, main, testID, since); err != nil {
			t.Fatal(err)
		}
	}
	return main, wt
}

// A worktree session (5.3 done, check 6): verified, it judges by the main checkout's lists (its own bonsai.yaml,
// emptied on its branch, changes nothing) and its records land in the main checkout's local/, which the cleaner (5.2.6b)
// cleans; pointed at a decoy main, with the link back forged, the decoy holding the id and recorded later, no grant
// counts (main-not-verified), the refusal names the step, and the records stay in the worktree's own local/.
func TestAWorktreeSession(t *testing.T) {
	tmp := t.TempDir()
	home := testpack.Isolate(t, tmp)
	since := time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC)
	main, wt := gitProject(t, home, testYAML, since)
	// The branch's own bonsai.yaml protects nothing: it changes nothing while the main checkout is verified.
	writeFile(t, filepath.Join(wt, "bonsai.yaml"), strings.Replace(strings.Replace(testYAML, `"protected.txt", `, "", 1),
		`".claude/**", "bonsai.yaml"`, `"bonsai.yaml"`, 1))
	code, stderr := runIn(t, wt, "w1", filepath.Join(wt, "protected.txt"))
	if code != 2 || !strings.Contains(stderr, "on bonsai.yaml's protected list") || strings.Contains(stderr, "not verified") {
		t.Errorf("a protected path in a verified worktree: exit %d, %s", code, stderr)
	}
	if code, stderr := runIn(t, wt, "w1", filepath.Join(wt, "free.txt")); code != 0 {
		t.Errorf("a free path in a verified worktree: exit %d, %s", code, stderr)
	}
	got := recordsIn(t, main, "w1")
	if len(got) != 2 || got[0] != "deny protected protected.txt wt "+testID || got[1] != "allow not-protected free.txt wt "+testID {
		t.Errorf("the worktree's records in main's local/: %q", got)
	}
	if _, err := os.Stat(filepath.Join(wt, ".bonsai", "local")); err == nil {
		t.Error("the verified worktree has a local/ of its own")
	}
	recs := records(t, main, "s-w1.ndjson")
	if recs[0].String("branch") != "t0901" || recs[0].String("bonsai_sha256") == "" || recs[1].String("bonsai_sha256") != "" {
		t.Errorf("records: %v", recs)
	}

	// The cleaner, as it runs at a session's end in the worktree (5.2.6b, its API as it is): a file of guard records
	// in main's local/, 31 days past its last line, is removed, with its clean record.
	local := workspace.FindLocal(wt, testID)
	if local.Main != main {
		t.Fatalf("FindLocal for the worktree: %+v", local)
	}
	done := clean.Files(clean.Options{Local: local, Now: func() time.Time { return testNow.Add(31 * 24 * time.Hour) }}, clean.Log)
	if len(done.Cleaned) != 1 || done.Cleaned[0].Target != ".bonsai/local/log/s-w1.ndjson" || len(done.Errs) != 0 {
		t.Fatalf("the cleaner: %+v", done)
	}
	if _, err := os.Stat(filepath.Join(main, filepath.FromSlash(LogDir), "s-w1.ndjson")); err == nil {
		t.Error("the old file of guard records is still there")
	}

	// A decoy main: the worktree's .git file rewritten to a folder holding the id, git's link back forged, recorded
	// on this machine later than the main checkout (bonsai init run in it).
	decoy := filepath.Join(tmp, "decoy")
	writeFile(t, filepath.Join(decoy, "bonsai.yaml"), testYAML)
	writeFile(t, filepath.Join(decoy, ".git", "HEAD"), "ref: refs/heads/main\n")
	writeFile(t, filepath.Join(decoy, ".git", "worktrees", "wt", "HEAD"), "ref: refs/heads/t0901\n")
	writeFile(t, filepath.Join(decoy, ".git", "worktrees", "wt", "commondir"), "../..\n")
	writeFile(t, filepath.Join(decoy, ".git", "worktrees", "wt", "gitdir"), filepath.ToSlash(filepath.Join(wt, ".git"))+"\n")
	writeFile(t, filepath.Join(wt, ".git"), "gitdir: "+filepath.ToSlash(filepath.Join(decoy, ".git", "worktrees", "wt"))+"\n")
	for _, c := range []struct {
		name   string
		record bool
		why    string
	}{
		{"a decoy main with no record", false, "no record of the main checkout"},
		{"a decoy main recorded later", true, "claimed this workspace's id before its main checkout did"},
	} {
		if c.record {
			if _, err := workspace.RecordCheckout(home, decoy, testID, since.Add(time.Hour)); err != nil {
				t.Fatal(err)
			}
		}
		// The worktree's own bonsai.yaml judges now (it protects bonsai.yaml, person-only), and no grant counts.
		code, stderr := runIn(t, wt, "w2", filepath.Join(wt, "bonsai.yaml"))
		if code != 2 || !strings.Contains(stderr, "no task can grant it here: its main checkout is not verified, as ") ||
			!strings.Contains(stderr, c.why) || !strings.Contains(stderr, "next: a person") {
			t.Errorf("%s: exit %d, %s", c.name, code, stderr)
		}
		asciiOnly(t, stderr)
		if code, _ := runIn(t, wt, "w2", filepath.Join(wt, "free.txt")); code != 0 {
			t.Errorf("%s, a free path: exit %d", c.name, code)
		}
	}
	got = recordsIn(t, wt, "w2")
	if len(got) != 4 || !strings.HasPrefix(got[0], "deny main-not-verified bonsai.yaml wt ") {
		t.Errorf("the redirected worktree's records, in its own local/: %q", got)
	}
	if got := recordsIn(t, decoy, "w2"); got != nil {
		t.Errorf("records in the decoy: %q", got)
	}
}

// A clone that never ran bonsai update on this machine: no grant could count, so a protected path is refused as
// main-not-verified and names the person's step; a free path is allowed. Recorded later than another checkout of the
// same id, the same.
func TestAnUnrecordedClone(t *testing.T) {
	home := t.TempDir()
	t.Setenv(workspace.HomeEnv, home)
	dir := unrecordedProject(t, testYAML)
	code, stderr := guardRun(t, env(dir), strings.NewReader(editOf("u1", dir, "protected.txt")), untimed)
	if code != 2 || !strings.Contains(stderr, `"protected.txt" refused: it is on bonsai.yaml's protected list ("protected.txt"), and no task can grant it here`) ||
		!strings.Contains(stderr, "next: a person runs bonsai update in the project's main checkout") {
		t.Errorf("exit %d, %s", code, stderr)
	}
	if code, _ := guardRun(t, env(dir), strings.NewReader(editOf("u1", dir, "free.txt")), untimed); code != 0 {
		t.Errorf("a free path: exit %d", code)
	}
	if recs := records(t, dir, "s-u1.ndjson"); len(recs) != 2 || recs[0].String("rule") != RuleMainNotVerified {
		t.Errorf("records %v", recs)
	}
	other := unrecordedProject(t, testYAML)
	if _, err := workspace.RecordCheckout(home, other, testID, testNow.Add(-time.Hour)); err != nil {
		t.Fatal(err)
	}
	if _, err := workspace.RecordCheckout(home, dir, testID, testNow); err != nil {
		t.Fatal(err)
	}
	code, stderr = guardRun(t, env(dir), strings.NewReader(editOf("u2", dir, "protected.txt")), untimed)
	if code != 2 || !strings.Contains(stderr, "another checkout on this machine claimed this workspace's id before") ||
		!strings.Contains(stderr, filepath.ToSlash(other)) {
		t.Errorf("a second clone: exit %d, %s", code, stderr)
	}
}

// Claude Code's own worktrees sit inside the main checkout (.claude/worktrees/), where the main checkout's .claude/**
// would refuse every edit: verified, the worktree judges alone, so its free paths are free.
func TestAWorktreeInsideTheMainCheckout(t *testing.T) {
	tmp := t.TempDir()
	home := testpack.Isolate(t, tmp)
	main, _ := gitProject(t, home, testYAML, time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC))
	inner := filepath.Join(main, ".claude", "worktrees", "agent-1")
	testpack.Git(t, main, "worktree", "add", "-q", "-b", "agent-1", inner)
	if code, stderr := runIn(t, inner, "c1", filepath.Join(inner, "free.txt")); code != 0 {
		t.Errorf("a free path in Claude Code's worktree: exit %d, %s", code, stderr)
	}
	if code, _ := runIn(t, inner, "c1", filepath.Join(inner, "protected.txt")); code != 2 {
		t.Errorf("a protected path in Claude Code's worktree: exit %d", code)
	}
	// Its link back broken (git's record names another folder): no longer verified, the main checkout's lists apply
	// too, and .claude/** refuses every edit in it.
	gitdirFile := filepath.Join(main, ".git", "worktrees", "agent-1", "gitdir")
	writeFile(t, gitdirFile, filepath.ToSlash(filepath.Join(tmp, "elsewhere", ".git"))+"\n")
	if code, stderr := runIn(t, inner, "c2", filepath.Join(inner, "free.txt")); code != 2 || !strings.Contains(stderr, `".claude/**"`) {
		t.Errorf("an unverified worktree inside the main checkout: exit %d, %s", code, stderr)
	}
}

// projectPaths gives a path's forms relative to one project's folder (the delete check, step 5.3.3, reads a shell
// operand so): as written and resolved, each once, none for a path outside it.
func TestProjectPaths(t *testing.T) {
	dir := project(t, testYAML)
	got, err := projectPaths(dir, dir, filepath.Join(dir, "sub", "..", "protected.txt"))
	if err != nil || len(got) != 1 || got[0] != "protected.txt" {
		t.Errorf("a path inside: %q %v", got, err)
	}
	if got, err := projectPaths(dir, filepath.Join(dir, "sub"), "../free.txt"); err != nil || len(got) != 1 || got[0] != "free.txt" {
		t.Errorf("a path relative to the cwd: %q %v", got, err)
	}
	if got, err := projectPaths(dir, dir, filepath.Join(filepath.Dir(dir), "x.txt")); err != nil || len(got) != 0 {
		t.Errorf("a path outside: %q %v", got, err)
	}
}
