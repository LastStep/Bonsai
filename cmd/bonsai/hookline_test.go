package main

// The hook lines' form (design/plan-5.md, 5.3.6 note 2, under the director's answer (a)): the shipped lines' exact text, the
// line run through the hook shell (sh on Linux, Git Bash on Windows) with its places swapped for temporary folders by
// the function that makes it, and test builds stamped with their own place and scratch folder.

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/engine"
	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/testpack"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// The shipped lines, as note 2 gives the guard's: Windows' place first, then WSL's; the stamps empty in a plain build
// (this test binary is one), so no default stamp is ever committed.
func TestTheShippedLines(t *testing.T) {
	if place, root := engine.TestBuild(); place != "" || root != "" || engine.Stamped() {
		t.Fatalf("a plain build is stamped: %q %q", place, root)
	}
	if strings.Count(versionLine(), "\n") != 1 {
		t.Errorf("a plain build's --version says more than its version: %q", versionLine())
	}
	if engine.PlaceWindows != "/c/Program Files/Bonsai/bonsai.exe" || engine.PlaceLinux != "/usr/local/bin/bonsai" {
		t.Errorf("the installed places: %q %q", engine.PlaceWindows, engine.PlaceLinux)
	}
	want := `b="/c/Program Files/Bonsai/bonsai.exe"; [ -x "$b" ] || b=/usr/local/bin/bonsai; [ -x "$b" ] || { echo ` +
		`"bonsai guard: no bonsai is installed at C:/Program Files/Bonsai/bonsai.exe or /usr/local/bin/bonsai, so this call is ` +
		`refused; next: tell the person: a person installs Bonsai there" >&2; exit 2; }; "$b" hook guard || exit 2`
	if engine.GuardCommand != want {
		t.Errorf("the guard's line:\n%s\nwant\n%s", engine.GuardCommand, want)
	}
	if want := `b="/c/Program Files/Bonsai/bonsai.exe"; [ -x "$b" ] || b=/usr/local/bin/bonsai; [ -x "$b" ] || exit 0; "$b" hook record`; engine.RecordCommand != want {
		t.Errorf("the recorder's line: %s", engine.RecordCommand)
	}
	if !strings.HasSuffix(engine.StartCommand, `[ -x "$b" ] || exit 0; "$b" hook start`) ||
		!strings.HasSuffix(engine.StopCommand, `"$b" hook stop || exit 2`) || !strings.Contains(engine.StopCommand, "bonsai stop: no bonsai is installed at") {
		t.Errorf("the start and stop lines:\n%s\n%s", engine.StartCommand, engine.StopCommand)
	}
	if want := `b="/c/Program Files/Bonsai/bonsai.exe"; [ -x "$b" ] || b=/usr/local/bin/bonsai; [ -x "$b" ] || exit 0; "$b" hook now || exit 0`; engine.NowCommand != want {
		t.Errorf("the time line: %s", engine.NowCommand)
	}
	for _, l := range []string{engine.GuardCommand, engine.StopCommand, engine.StartCommand, engine.RecordCommand, engine.NowCommand} {
		if failedToStart(l) || strings.Contains(l, "PATH") || strings.Contains(l, "bash") {
			t.Errorf("a line holds words read as a hook that failed to start, reads the PATH or names bash: %s", l)
		}
	}
	if engine.GuardMatcher != "Edit|Write|MultiEdit|NotebookEdit|Bash|PowerShell|Monitor" || engine.GuardOnFailure != "block" {
		t.Errorf("the guard's matcher %q, onFailure %q", engine.GuardMatcher, engine.GuardOnFailure)
	}
}

// copyFile copies a binary to dst, making its folder, executable.
func copyFile(t *testing.T, src, dst string) {
	t.Helper()
	b, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dst, b, 0o755); err != nil {
		t.Fatal(err)
	}
}

// Note 2's line test: the shipped line's exact text, made by the same function with the two places swapped for
// temporary folders (the first with a space in its name), run through the hook shell: the first place present; only
// the second; neither (exit 2, and the line's own words, which never read as a hook that failed to start); a decoy
// bonsai first on the PATH, ignored. The stop line and the start and recorder lines with neither present too.
func TestTheLineThroughTheShell(t *testing.T) {
	normal, _ := bonsaiBuilds(t)
	proj := guardProject(t)
	exe := filepath.Base(normal)
	tmp := t.TempDir()
	first := filepath.Join(tmp, "place one", exe)
	second := filepath.Join(tmp, "place-two", exe)
	places := []string{shellPath(first), shellPath(second)}

	// A decoy bonsai first on the PATH: a script that leaves a mark, which Git Bash runs too (step 5.3.6a measured it).
	decoyDir := filepath.Join(tmp, "decoy")
	mark := filepath.Join(tmp, "decoy-ran")
	if err := os.MkdirAll(decoyDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(decoyDir, "bonsai"), []byte("#!/bin/sh\necho ran >> '"+shellPath(mark)+"'\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	decoyPath := decoyDir + string(os.PathListSeparator) + os.Getenv("PATH")
	protected := func(name string) string { return editPayload(t, name, proj, "protected.txt") }
	refused := func(name, stderr string, code int, from string) {
		t.Helper()
		if code != 2 || !strings.Contains(stderr, `Edit of "protected.txt" refused`) {
			t.Fatalf("%s: exit %d, stderr %q", name, code, stderr)
		}
		if r := lastRecord(t, proj, name); r.String("bonsai_path") != logPath(t, from) {
			t.Errorf("%s: answered by %s, want %s", name, r.String("bonsai_path"), logPath(t, from))
		}
	}

	t.Run("neither place", func(t *testing.T) {
		for _, path := range []string{hookPath(t), decoyPath} {
			code, stderr, _ := runLine(t, places, path, proj, "", protected("s-none"))
			want := "bonsai guard: no bonsai is installed at " + shownPlace(places[0]) + " or " + shownPlace(places[1]) +
				", so this call is refused; next: tell the person: a person installs Bonsai there\n"
			if code != 2 || stderr != want || failedToStart(stderr) {
				t.Fatalf("exit %d, stderr %q; want 2 and %q", code, stderr, want)
			}
			code, stderr, _ = runHookText(t, engine.HookCommand("stop", true, places...), path, proj, "", "{}")
			if code != 2 || !strings.HasPrefix(stderr, "bonsai stop: no bonsai is installed at ") || failedToStart(stderr) {
				t.Fatalf("the stop line: exit %d, stderr %q", code, stderr)
			}
			for _, w := range []string{"start", "record"} {
				if code, stderr, _ := runHookText(t, engine.HookCommand(w, false, places...), path, proj, "", "{}"); code != 0 || stderr != "" {
					t.Fatalf("the %s line: exit %d, stderr %q; want 0 and nothing", w, code, stderr)
				}
			}
		}
	})
	t.Run("only the second", func(t *testing.T) {
		copyFile(t, normal, second)
		code, stderr, _ := runLine(t, places, decoyPath, proj, "", protected("s-second"))
		refused("s-second", stderr, code, second)
		if code, stderr, _ := runLine(t, places, decoyPath, proj, "", editPayload(t, "s-second-free", proj, "free.txt")); code != 0 || stderr != "" {
			t.Fatalf("free.txt: exit %d, stderr %q", code, stderr)
		}
	})
	t.Run("the first present", func(t *testing.T) {
		copyFile(t, normal, first)
		code, stderr, _ := runLine(t, places, decoyPath, proj, "", protected("s-first"))
		refused("s-first", stderr, code, first)
	})
	if _, err := os.Stat(mark); err == nil {
		t.Error("the decoy bonsai on the PATH ran")
	}
}

// shownPlace is a place as the line's words name it: a Git Bash drive path as C:/... (engine's shownPlace).
func shownPlace(p string) string {
	if len(p) >= 3 && p[0] == '/' && p[2] == '/' && ((p[1] >= 'a' && p[1] <= 'z') || (p[1] >= 'A' && p[1] <= 'Z')) {
		return strings.ToUpper(p[1:2]) + ":" + p[2:]
	}
	return p
}

// stampedBuild builds this command stamped as a test build (engine's hookPlace and scratchRoot), into place.
func stampedBuild(t *testing.T, place, root string) {
	t.Helper()
	goTool, err := exec.LookPath("go")
	if err != nil {
		t.Skip("no go on the PATH to build a stamped binary with")
	}
	pkg := "github.com/LastStep/Bonsai/internal/engine"
	ld := "-X " + pkg + ".hookPlace=" + shellPath(place) + " -X " + pkg + ".scratchRoot=" + root
	cmd := exec.Command(goTool, "build", "-ldflags", ld, "-o", place, ".")
	cmd.Dir = pkgDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("building a stamped bonsai: %v\n%s", err, out)
	}
}

// runBin runs a built bonsai in dir with this process's environment and more, and gives its exit code and output.
func runBin(t *testing.T, bin, dir string, env []string, args ...string) (int, string, string) {
	t.Helper()
	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), env...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	var ee *exec.ExitError
	switch {
	case errors.As(err, &ee):
		return ee.ExitCode(), stdout.String(), stderr.String()
	case err != nil:
		t.Fatalf("running %s: %v", bin, err)
	}
	return 0, stdout.String(), stderr.String()
}

// A test build (note 2): stamped with its own place and a scratch folder, it says so in --version, writes its place
// alone in its lines, and links, updates and unlinks only inside its scratch folder (test-build, exit 4, nothing
// written, elsewhere); it skips the refusal of an agent's person's commands (note 8), since its folder confines it. A
// build stamped with one of the two links nowhere.
func TestStampedBuild(t *testing.T) {
	c := newCLI(t)
	scratch := filepath.Join(c.tmp, "scratch")
	exe := "bonsai"
	if runtime.GOOS == "windows" {
		exe = "bonsai.exe"
	}
	place := filepath.Join(c.tmp, "stamped", exe)
	stampedBuild(t, place, scratch)

	if code, out, _ := runBin(t, place, c.root, nil, "--version"); code != 0 ||
		!strings.Contains(out, "\ntest build: its hook lines run "+shellPath(place)+"; it links, updates, unlinks and climbs the ladder only inside "+filepath.ToSlash(scratch)+"\n") {
		t.Errorf("--version: %d %q", code, out)
	}
	initArgs := append(c.linkArgs(c.pack.A), "--yes")

	// Outside its scratch folder: refused, nothing written, at --json too.
	before := tree(t, c.root)
	for _, args := range [][]string{initArgs, append(initArgs, "--json"), {"update", "--yes"}, {"unlink", "--yes", "--json"}} {
		code, out, errOut := runBin(t, place, c.root, nil, args...)
		if code != 4 || !strings.Contains(out+errOut, "test build") {
			t.Errorf("%v outside its scratch folder: exit %d\n%s%s", args, code, out, errOut)
		}
		if strings.Contains(strings.Join(args, " "), "--json") {
			if e := errorIn(fits(t, out, "changes"), "changes"); e.String("code") != "test-build" {
				t.Errorf("%v: error %s", args, schema.Show(e))
			}
		}
	}
	if tree(t, c.root) != before {
		t.Error("a stamped build wrote outside its scratch folder")
	}

	// Inside it, in an agent's session too: linked, its place alone in the lines.
	inside := testpack.Project(t, scratch, "inside")
	if code, out, errOut := runBin(t, place, inside, []string{"CLAUDE_CODE_CHILD_SESSION=1"}, initArgs...); code != 0 {
		t.Fatalf("init inside its scratch folder: exit %d\n%s%s", code, out, errOut)
	}
	raw, err := os.ReadFile(filepath.Join(inside, ".claude", "settings.json"))
	if err != nil {
		t.Fatal(err)
	}
	guardLine := engine.HookCommand("guard", true, shellPath(place))
	if !strings.Contains(string(raw), jsonCmd(guardLine)) || strings.Contains(string(raw), engine.PlaceLinux) ||
		strings.Contains(string(raw), "Program Files") {
		t.Errorf("the stamped build's lines:\n%s", raw)
	}
	if code, out, errOut := runBin(t, place, inside, []string{"CLAUDE_CODE_CHILD_SESSION=1"}, "unlink", "--yes"); code != 0 {
		t.Errorf("unlink inside its scratch folder, in an agent's session: exit %d\n%s%s", code, out, errOut)
	}

	// A plain build takes a stamped build's lines as Bonsai's (the form), and replaces them only with --allow-exec.
	if code, out, errOut := runBin(t, place, inside, nil, initArgs...); code != 0 {
		t.Fatalf("init again: exit %d\n%s%s", code, out, errOut)
	}
	normal, _ := bonsaiBuilds(t)
	if code, out, _ := runBin(t, normal, inside, []string{"CLAUDE_CODE_CHILD_SESSION="}, "update", "--yes"); code != 4 ||
		!strings.Contains(flat(out), "change hook Stop: "+flat(engine.StopCommand)) || !strings.Contains(flat(out), "needs --allow-exec") {
		t.Errorf("a plain update of a stamped link, --yes alone: exit %d\n%s", code, out)
	}
	if code, out, errOut := runBin(t, normal, inside, []string{"CLAUDE_CODE_CHILD_SESSION="}, "update", "--allow-exec", "--yes"); code != 0 {
		t.Errorf("a plain update --allow-exec: exit %d\n%s%s", code, out, errOut)
	}
	if raw, _ := os.ReadFile(filepath.Join(inside, ".claude", "settings.json")); strings.Contains(string(raw), shellPath(place)) ||
		!strings.Contains(string(raw), jsonCmd(engine.GuardCommand)) {
		t.Errorf("after a plain update:\n%s", raw)
	}

	// One stamp alone: it links nowhere.
	half := filepath.Join(c.tmp, "half", exe)
	goTool, _ := exec.LookPath("go")
	cmd := exec.Command(goTool, "build", "-ldflags", "-X github.com/LastStep/Bonsai/internal/engine.scratchRoot="+scratch, "-o", half, ".")
	cmd.Dir = pkgDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("building: %v\n%s", err, out)
	}
	other := testpack.Project(t, scratch, "other")
	if code, out, errOut := runBin(t, half, other, nil, initArgs...); code != 4 || !strings.Contains(errOut, "only one of hookPlace and scratchRoot") {
		t.Errorf("a half-stamped build: exit %d\n%s%s", code, out, errOut)
	}
}

// The guard's second line (plan-5 5.5.11 note 2), run through the hook shell as Claude Code runs it: with bonsai at
// neither place it lets a read through, printing nothing; with bonsai there it refuses a read past an agent's budget,
// in a project whose committed lock declares a role of 3 actions; a tool the first line judges it passes at once.
func TestTheSecondLineThroughTheShell(t *testing.T) {
	normal, _ := bonsaiBuilds(t)
	tmp := t.TempDir()
	testpack.Isolate(t, tmp)
	proj := filepath.Join(tmp, "proj")
	pack, err := format.ReadPack([]byte("format: bonsai.pack/1\nid: workflow\nversion: \"1\"\nroles:\n  - name: tiny\n" +
		"    kind: helper\n    session: false\n    writes: false\n    max_turns: null\n    minutes:\n      default: null\n" +
		"      max: null\n    actions:\n      default: 3\n      max: 3\n    red_runs:\n      default: null\n      max: null\n"))
	if err != nil {
		t.Fatal(err)
	}
	decl, err := (&format.Declares{Roles: pack.Roles}).Object()
	if err != nil {
		t.Fatal(err)
	}
	lock, err := (&workspace.Lock{WrittenBy: "0.5.0", Files: map[string]workspace.LockedFile{}, Packs: []workspace.LockedPack{{
		ID: "workflow", Source: "https://example.invalid/workflow.git", Version: "1", Commit: strings.Repeat("a", 40),
		SHA256: strings.Repeat("b", 64), Declares: decl, PathSet: true}}}).Encode()
	if err != nil {
		t.Fatal(err)
	}
	for name, body := range map[string]string{
		"bonsai.yaml": "format: bonsai.workspace/1\nid: ws-aaaaaaaaaaaaaaaaaaaaaaaaaa\nname: roles\npacks:\n  - id: workflow\n" +
			"    source: \"https://example.invalid/workflow.git\"\n    path: null\n    ref: \"" + strings.Repeat("a", 40) + "\"\n" +
			"    roles: null\nprotected: []\nperson_only: []\nnever_edit: []\n",
		".bonsai/lock.json": string(lock), "free.txt": "free\n"} {
		p := filepath.Join(proj, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	testpack.Git(t, proj, "init", "-q")
	testpack.Git(t, proj, "add", "-A")
	testpack.Git(t, proj, "commit", "-q", "-m", "roles")
	read := func(tool string) string {
		b, err := schema.EncodeLine(schema.Object{{Key: "session_id", Value: "s-second"}, {Key: "cwd", Value: proj},
			{Key: "agent_id", Value: "a1"}, {Key: "agent_type", Value: "workflow:tiny"}, {Key: "hook_event_name", Value: "PreToolUse"},
			{Key: "tool_name", Value: tool}, {Key: "tool_input", Value: schema.Object{{Key: "file_path", Value: filepath.Join(proj, "free.txt")}}},
			{Key: "tool_use_id", Value: "toolu_01S"}})
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	place := filepath.Join(tmp, "place", filepath.Base(normal))
	line := engine.HookCommand(engine.SecondWord, false, shellPath(place))
	// Neither place: let through, nothing printed, nothing counted.
	for i := 0; i < 5; i++ {
		if code, stderr, _ := runHookText(t, line, hookPath(t), proj, "", read("Read")); code != 0 || stderr != "" {
			t.Fatalf("with bonsai missing: exit %d, stderr %q", code, stderr)
		}
	}
	copyFile(t, normal, place)
	for i := 0; i < 3; i++ {
		if code, stderr, _ := runHookText(t, line, hookPath(t), proj, "", read("Read")); code != 0 {
			t.Fatalf("read %d within the budget: exit %d, %s", i+1, code, stderr)
		}
	}
	code, stderr, _ := runHookText(t, line, hookPath(t), proj, "", read("Read"))
	if code != 2 || !strings.Contains(stderr, "budget spent (actions: 4 of 3)") {
		t.Fatalf("a read past the budget: exit %d, stderr %q", code, stderr)
	}
	// A tool the first line judges: passed at once by the second.
	if code, stderr, _ := runHookText(t, line, hookPath(t), proj, "", read("Edit")); code != 0 {
		t.Errorf("an Edit through the second line: exit %d, %s", code, stderr)
	}
}
