package main

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/engine"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/selfcheck"
)

// builds holds the two binaries this package's hook tests run, built once: a normal build and a build with the
// tag bonsai_test_fault.
var builds struct {
	once          sync.Once
	dir           string
	normal, fault string
	err           error
}

// pkgDir is this package's folder, where go test starts: the builds are made there, whatever folder a test moved to.
var pkgDir, _ = os.Getwd()

func TestMain(m *testing.M) {
	code := m.Run()
	if builds.dir != "" {
		_ = os.RemoveAll(builds.dir)
	}
	os.Exit(code)
}

// bonsaiBuilds builds this command twice, each binary named bonsai in its own folder, as a PATH would hold it.
func bonsaiBuilds(t *testing.T) (normal, fault string) {
	t.Helper()
	builds.once.Do(func() {
		goTool, err := exec.LookPath("go") // go test puts its own Go first on the PATH
		if err != nil {
			builds.err = err
			return
		}
		dir, err := os.MkdirTemp("", "bonsai-hook-builds-")
		if err != nil {
			builds.err = err
			return
		}
		builds.dir = dir
		exe := "bonsai"
		if runtime.GOOS == "windows" {
			exe = "bonsai.exe"
		}
		builds.normal = filepath.Join(dir, "normal", exe)
		builds.fault = filepath.Join(dir, "fault", exe)
		for _, b := range []struct{ out, tags string }{{builds.normal, ""}, {builds.fault, "bonsai_test_fault"}} {
			args := []string{"build", "-o", b.out}
			if b.tags != "" {
				args = append(args, "-tags", b.tags)
			}
			cmd := exec.Command(goTool, append(args, ".")...)
			cmd.Dir = pkgDir
			out, err := cmd.CombinedOutput()
			if err != nil {
				builds.err = errors.New(err.Error() + ": " + string(out))
				return
			}
		}
	})
	if builds.err != nil {
		t.Fatalf("building bonsai: %v", builds.err)
	}
	return builds.normal, builds.fault
}

// A normal build holds no fault code: none of the switch's words, rules or functions (selfcheck.FaultMarkers, their one
// home, which the release's dist checks read too) is in its bytes, and each is in the fault build's, so the check can
// see them.
func TestNormalBuildHasNoFaultCode(t *testing.T) {
	normal, fault := bonsaiBuilds(t)
	nb, err := os.ReadFile(normal)
	if err != nil {
		t.Fatal(err)
	}
	fb, err := os.ReadFile(fault)
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range selfcheck.FaultMarkers() {
		if bytes.Contains(nb, []byte(m)) {
			t.Errorf("the normal build holds %q", m)
		}
		if !bytes.Contains(fb, []byte(m)) {
			t.Errorf("the fault build lacks %q, so this check could not see it", m)
		}
	}
}

// guardProject makes a linked project in a temporary folder: bonsai.yaml protecting protected.txt, a .git folder,
// protected.txt and free.txt.
func guardProject(t *testing.T) string {
	t.Helper()
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		"bonsai.yaml": "format: bonsai.workspace/1\nid: ws-aaaaaaaaaaaaaaaaaaaaaaaaaa\nname: guard\npacks: []\n" +
			"protected: [\".claude/**\", \"bonsai.yaml\", \"protected.txt\"]\nperson_only: [\"bonsai.yaml\"]\nnever_edit: []\n",
		".git/HEAD":     "ref: refs/heads/main\n",
		"protected.txt": "protected\n",
		"free.txt":      "free\n",
	}
	for name, body := range files {
		p := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func editPayload(t *testing.T, session, dir, rel string) string {
	t.Helper()
	b, err := schema.EncodeLine(schema.Object{
		{Key: "session_id", Value: session},
		{Key: "cwd", Value: dir},
		{Key: "hook_event_name", Value: "PreToolUse"},
		{Key: "tool_name", Value: "Edit"},
		{Key: "tool_input", Value: schema.Object{{Key: "file_path", Value: filepath.Join(dir, rel)},
			{Key: "old_string", Value: "a"}, {Key: "new_string", Value: "b"}}},
		{Key: "tool_use_id", Value: "toolu_01HOOK"},
	})
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// hookShell is the shell Claude Code runs a hook line in: sh on Linux and macOS, Git Bash on Windows (by its full
// path, never `bash` by name: on Windows that can be WSL's launcher).
func hookShell(t *testing.T) string {
	t.Helper()
	if runtime.GOOS != "windows" {
		if _, err := os.Stat("/bin/sh"); err != nil {
			t.Skip("no /bin/sh to run the hook line in")
		}
		return "/bin/sh"
	}
	if p := os.Getenv("CLAUDE_CODE_GIT_BASH_PATH"); p != "" {
		return p
	}
	git, err := exec.LookPath("git")
	if err != nil {
		t.Skip("no git on the PATH, so no Git Bash: Claude Code runs Windows hooks in Git Bash, and so does this test")
	}
	for d := filepath.Dir(git); ; {
		if p := filepath.Join(d, "bin", "bash.exe"); fileExists(p) {
			return p
		}
		parent := filepath.Dir(d)
		if parent == d {
			break
		}
		d = parent
	}
	t.Skipf("no Git Bash beside %s: Claude Code runs Windows hooks in Git Bash, and so does this test", git)
	return ""
}

func fileExists(p string) bool {
	fi, err := os.Stat(p)
	return err == nil && !fi.IsDir()
}

// hookEnv is this process's environment without PATH, the project variable and the fault switch, plus vars.
func hookEnv(vars ...string) []string {
	var out []string
	for _, kv := range os.Environ() {
		k := kv
		if i := strings.IndexByte(kv, '='); i > 0 {
			k = kv[:i]
		}
		switch strings.ToUpper(k) {
		case "PATH", "CLAUDE_PROJECT_DIR", "BONSAI_TEST" + "_FAULT", "BONSAI_HOME":
			continue
		}
		out = append(out, kv)
	}
	return append(out, vars...)
}

// runLine runs the guard's hook line, in Bonsai's form with places as its places (engine.HookCommand, the function
// that makes the shipped line), in the hook shell with a PATH and a payload, as Claude Code does, and gives its exit
// code and stderr.
func runLine(t *testing.T, places []string, path, project, fault, stdin string) (int, string, time.Duration) {
	t.Helper()
	return runHookText(t, engine.HookCommand("guard", true, places...), path, project, fault, stdin)
}

// runHookText runs a hook line's text in the hook shell, as runLine does.
func runHookText(t *testing.T, line, path, project, fault, stdin string) (int, string, time.Duration) {
	t.Helper()
	cmd := exec.Command(hookShell(t), "-c", line)
	cmd.Env = hookEnv("PATH="+path, "CLAUDE_PROJECT_DIR="+project, "BONSAI_HOME="+t.TempDir(), "BONSAI_TEST"+"_FAULT="+fault)
	cmd.Stdin = strings.NewReader(stdin)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	start := time.Now()
	err := cmd.Run()
	took := time.Since(start)
	code := 0
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		code = ee.ExitCode()
	} else if err != nil {
		t.Fatalf("running the hook line: %v", err)
	}
	return code, stderr.String(), took
}

// shellPath is a path as the hook shell names it: as it is on Linux and macOS; on Windows with forward slashes and its
// drive as Git Bash names one, C:\x as /c/x, the form Bonsai's lines name Windows' installed place in.
func shellPath(p string) string {
	p = filepath.ToSlash(p)
	if runtime.GOOS == "windows" && len(p) >= 3 && p[1] == ':' && p[2] == '/' {
		return "/" + strings.ToLower(p[:1]) + p[2:]
	}
	return p
}

// hookPath is the PATH a hook line gets in these tests: the test's own, which may hold a bonsai of its own (an old
// build, a decoy): Bonsai's line never reads it.
func hookPath(t *testing.T) string {
	t.Helper()
	return os.Getenv("PATH")
}

// minimalPath is a bare PATH, the system's own folders only, as a session started by a service might have.
func minimalPath(t *testing.T) string {
	t.Helper()
	dirs := []string{"/usr/bin", "/bin"}
	if runtime.GOOS == "windows" {
		root := os.Getenv("SystemRoot")
		if root == "" {
			root = `C:\Windows`
		}
		dirs = []string{filepath.Join(root, "System32"), root}
	}
	return strings.Join(dirs, string(os.PathListSeparator))
}

func lastRecord(t *testing.T, project, session string) schema.Object {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(project, ".bonsai", "local", "log", "s-"+session+".ndjson"))
	if err != nil {
		t.Fatalf("the log: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(raw)), "\n")
	v, err := schema.Decode([]byte(lines[len(lines)-1]))
	if err != nil {
		t.Fatal(err)
	}
	return v.(schema.Object)
}

// Check 11's fail-closed half, on the built binaries and part 3's real hook line: the guard refuses the protected
// edit and allows the free one, and each of the four faults blocks the free edit.
// logPath is a binary's path as a guard record names it (the log schema's bonsai_path; contract §2.6): forward slashes,
// written from ~/ when it lies under the user's home folder (letter case ignored on Windows), else as it is. Worked
// out here on its own, from os.UserHomeDir, so a wrong conversion in the guard fails the test. On Windows the test's
// builds lie in %TEMP%, under the home; on Linux in /tmp, outside it.
func logPath(t *testing.T, bin string) string {
	t.Helper()
	p := filepath.ToSlash(filepath.Clean(bin))
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return p
	}
	prefix := strings.TrimSuffix(filepath.ToSlash(filepath.Clean(home)), "/") + "/"
	if len(p) > len(prefix) && (p[:len(prefix)] == prefix || (runtime.GOOS == "windows" && strings.EqualFold(p[:len(prefix)], prefix))) {
		return "~/" + p[len(prefix):]
	}
	return p
}

func TestFaultsThroughTheHookLine(t *testing.T) {
	normal, fault := bonsaiBuilds(t)
	proj := guardProject(t)
	nowhere := shellPath(filepath.Join(t.TempDir(), "bonsai"))
	free := editPayload(t, "s-free", proj, "free.txt")

	t.Run("normal build", func(t *testing.T) {
		at := []string{shellPath(normal)}
		if code, stderr, _ := runLine(t, at, hookPath(t), proj, "", free); code != 0 || stderr != "" {
			t.Fatalf("free.txt: exit %d, stderr %q", code, stderr)
		}
		code, stderr, _ := runLine(t, at, hookPath(t), proj, "", editPayload(t, "s-prot", proj, "protected.txt"))
		if code != 2 || !strings.Contains(stderr, `Edit of "protected.txt" refused`) {
			t.Fatalf("protected.txt: exit %d, stderr %q", code, stderr)
		}
		if r := lastRecord(t, proj, "s-prot"); r.String("bonsai_path") != logPath(t, normal) || len(r.String("bonsai_sha256")) != 64 {
			t.Errorf("the record names %s %s, want %s and a hash", r.String("bonsai_path"), r.String("bonsai_sha256"), logPath(t, normal))
		}
		// The normal build has no switch: set or not, it decides the same.
		if code, stderr, _ := runLine(t, at, hookPath(t), proj, "crash", free); code != 0 {
			t.Fatalf("normal build with the crash switch set: exit %d, stderr %q", code, stderr)
		}
	})
	t.Run("missing", func(t *testing.T) {
		// No bonsai at either place: the line refuses itself, in its own words, and starts no bonsai.
		code, stderr, _ := runLine(t, []string{nowhere, nowhere + "-too"}, hookPath(t), proj, "missing", free)
		if code != 2 || !strings.HasPrefix(stderr, "bonsai guard: no bonsai is installed at ") || failedToStart(stderr) {
			t.Fatalf("exit %d, stderr %q; want 2 and the line's own words", code, stderr)
		}
	})
	t.Run("minimal-path", func(t *testing.T) {
		// A bare PATH changes nothing: the line never reads it.
		at := []string{nowhere, shellPath(normal)}
		if code, stderr, _ := runLine(t, at, minimalPath(t), proj, "minimal-path", free); code != 0 || stderr != "" {
			t.Fatalf("free.txt: exit %d, stderr %q", code, stderr)
		}
		if code, stderr, _ := runLine(t, at, minimalPath(t), proj, "minimal-path", editPayload(t, "s-min", proj, "protected.txt")); code != 2 ||
			!strings.Contains(stderr, `Edit of "protected.txt" refused`) {
			t.Fatalf("protected.txt: exit %d, stderr %q", code, stderr)
		}
	})
	t.Run("crash", func(t *testing.T) {
		// Run alone, the crashed guard ends with neither 0 nor 2: only the hook line's `|| exit 2` blocks.
		cmd := exec.Command(fault, "hook", "guard")
		cmd.Env = hookEnv("CLAUDE_PROJECT_DIR="+proj, "BONSAI_HOME="+t.TempDir(), "BONSAI_TEST"+"_FAULT=crash")
		cmd.Stdin = strings.NewReader(editPayload(t, "s-crash-alone", proj, "free.txt"))
		err := cmd.Run()
		var ee *exec.ExitError
		if !errors.As(err, &ee) || ee.ExitCode() == 0 || ee.ExitCode() == 2 {
			t.Fatalf("the crashed guard alone: %v; want an exit other than 0 and 2", err)
		}
		code, stderr, _ := runLine(t, []string{shellPath(fault)}, hookPath(t), proj, "crash", editPayload(t, "s-crash", proj, "free.txt"))
		if code != 2 || !strings.Contains(stderr, "test fault crash") {
			t.Fatalf("exit %d, stderr %q", code, stderr)
		}
		if r := lastRecord(t, proj, "s-crash"); r.String("rule") != "fault-"+"crash" || r.String("bonsai_path") != logPath(t, fault) {
			t.Errorf("record: %s", schema.Show(r))
		}
	})
	t.Run("slow", func(t *testing.T) {
		// The block must come before the hook line's timeout, past which Claude Code lets the call through: held to that
		// timeout (engine.GuardTimeout, the line's own constant) less a second for the shell and the process start, so
		// the two cannot drift. The guard needs its 5 s budget and the record's wait.
		secs, err := strconv.Atoi(engine.GuardTimeout)
		if err != nil {
			t.Fatalf("engine.GuardTimeout %q: %v", engine.GuardTimeout, err)
		}
		limit := time.Duration(secs)*time.Second - time.Second
		// The session's log file is made first, by a call the fault build decides at once, as bonsai hook start makes it
		// in a real session: the over-time record is then an append, not the file's first record, which hashes the
		// whole binary inside the frame's 500 ms wait for it (hookin.RecordWait). Under a loaded machine (every test of a
		// climb's rung at once) that wait can still pass before the record lands; the frame then answers without it, by
		// design (a record not written by then may land after the answer, or never), so the record is looked for over
		// three slow calls, each of which must block in time.
		if code, stderr, _ := runLine(t, []string{shellPath(fault)}, hookPath(t), proj, "", editPayload(t, "s-slow", proj, "free.txt")); code != 0 {
			t.Fatalf("the call that makes the session's file: exit %d, stderr %q", code, stderr)
		}
		var r schema.Object
		for try := 0; try < 3; try++ {
			code, stderr, took := runLine(t, []string{shellPath(fault)}, hookPath(t), proj, "slow", editPayload(t, "s-slow", proj, "free.txt"))
			if code != 2 || !strings.Contains(stderr, "the guard ran past its own 5s limit") {
				t.Fatalf("exit %d, stderr %q", code, stderr)
			}
			if took > limit {
				t.Errorf("blocked after %v, past %v: so near the hook line's %s s timeout, Claude Code could have let it through", took, limit, engine.GuardTimeout)
			}
			if r = lastRecord(t, proj, "s-slow"); r.String("rule") == "over-time" {
				break
			}
		}
		if r.String("rule") != "over-time" {
			t.Errorf("no over-time record in three slow calls: %s", schema.Show(r))
		}
	})
	t.Run("a fault build answering missing", func(t *testing.T) {
		// A fault build at the place, with the switch saying no bonsai should answer: it blocks, naming itself.
		for _, f := range []string{"missing", "minimal-path"} {
			code, stderr, _ := runLine(t, []string{shellPath(fault)}, hookPath(t), proj, f, free)
			if code != 2 || !strings.Contains(stderr, "the launcher did not hide it") {
				t.Fatalf("%s: exit %d, stderr %q", f, code, stderr)
			}
		}
	})
}

// failedToStart reports words a shell or an operating system uses for a file that is not there: Claude Code reads a
// Stop hook's exit 2 with such words on stderr as a hook that failed to start, and lets the session end (its hooks
// reference; step 5.3.4).
func failedToStart(stderr string) bool {
	l := strings.ToLower(stderr)
	for _, w := range []string{"no such file", "not found", "missing", "cannot find", "enoent", "does not exist", "failed to start"} {
		if strings.Contains(l, w) {
			return true
		}
	}
	return false
}

func TestHookCommand(t *testing.T) {
	cases := []struct {
		args     []string
		code     int
		inStdout string
		inStderr string
	}{
		{[]string{"hook", "--help"}, 0, "bonsai hook guard ", ""},
		{[]string{"hook", "guard", "--help"}, 0, engine.ShownLine(engine.GuardCommand) + "on PreToolUse (Edit, Write, MultiEdit, NotebookEdit, Bash, PowerShell, Monitor), timeout 10 s, onFailure block.\n", ""},
		{[]string{"hook", "guard", "--help"}, 0, "A Bash, PowerShell or Monitor call is judged too", ""},
		{[]string{"hook", "start", "--help"}, 0, engine.ShownLine(engine.StartCommand) + "on SessionStart", ""},
		{[]string{"hook", "record", "--help"}, 0, engine.ShownLine(engine.RecordCommand) + "on UserPromptSubmit", ""},
		{[]string{"hook", "--help"}, 0, "never one on the PATH", ""},
		{[]string{"hook"}, 2, "", "bonsai hook: refused  [missing-value]\n  what  exit 2, block\n  why   bonsai hook needs a hook name\n  next  run `bonsai hook --help`\n"},
		{[]string{"hook", "guard", "--json"}, 2, "", "bonsai hook guard: refused  [bad-flag]\n  what  exit 2, block\n  why   hook guard takes no \"--json\"\n"},
		{[]string{"hook", "stop", "--help"}, 0, engine.ShownLine(engine.StopCommand) + "on Stop (no matcher), timeout 10 s.\n", ""},
		{[]string{"hook", "stop", "--help"}, 0, engine.HookPlaceWords(), ""},
		{[]string{"hook", "stop", "now"}, 2, "", "bonsai hook stop: refused  [bad-flag]\n  what  exit 2, block\n  why   hook stop takes no \"now\"\n"},
		{[]string{"hook", "nap"}, 2, "", "bonsai hook: refused  [bad-value]\n  what  exit 2, block\n  why   \"nap\" is not a hook\n"},
	}
	for _, c := range cases {
		var stdout, stderr bytes.Buffer
		code := run(c.args, &stdout, &stderr)
		if code != c.code || !strings.Contains(stdout.String(), c.inStdout) || !strings.Contains(stderr.String(), c.inStderr) {
			t.Errorf("%v: exit %d, stdout %q, stderr %q", c.args, code, stdout.String(), stderr.String())
		}
		asciiOnly(t, "stdout", stdout.String())
		asciiOnly(t, "stderr", stderr.String())
		if code != 0 && !hasNext(stderr.String()) {
			t.Errorf("%v: a refusal with no next step", c.args)
		}
	}
}
