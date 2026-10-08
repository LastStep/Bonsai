package main

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/engine"
	"github.com/LastStep/Bonsai/internal/schema"
)

// builds holds the two binaries this package's hook tests run, built once: a normal build and a build with the
// tag bonsai_test_fault.
var builds struct {
	once          sync.Once
	dir           string
	normal, fault string
	err           error
}

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
			out, err := exec.Command(goTool, append(args, ".")...).CombinedOutput()
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

// A normal build holds no fault code: none of the switch's words, rules or functions is in its bytes, and each is
// in the fault build's, so the check can see them.
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
	for _, m := range []string{"BONSAI_TEST" + "_FAULT", "fault-" + "crash", "fault-" + "unknown", "minimal" + "-path", "test fault " + "crash",
		"internal/guard." + "crashHard"} {
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

// runLine runs part 3's hook line in the hook shell with a PATH and a payload, as Claude Code does, and gives its
// exit code and stderr.
func runLine(t *testing.T, path, project, fault, stdin string) (int, string, time.Duration) {
	t.Helper()
	cmd := exec.Command(hookShell(t), "-c", engine.GuardCommand)
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
	for _, d := range dirs {
		for _, name := range []string{"bonsai", "bonsai.exe"} {
			if fileExists(filepath.Join(d, name)) {
				t.Skipf("a bonsai is installed in %s, so a bare PATH still finds one", d)
			}
		}
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
func TestFaultsThroughTheHookLine(t *testing.T) {
	normal, fault := bonsaiBuilds(t)
	proj := guardProject(t)
	normalPath, faultPath := filepath.Dir(normal), filepath.Dir(fault)
	free := editPayload(t, "s-free", proj, "free.txt")

	t.Run("normal build", func(t *testing.T) {
		if code, stderr, _ := runLine(t, normalPath, proj, "", free); code != 0 || stderr != "" {
			t.Fatalf("free.txt: exit %d, stderr %q", code, stderr)
		}
		code, stderr, _ := runLine(t, normalPath, proj, "", editPayload(t, "s-prot", proj, "protected.txt"))
		if code != 2 || !strings.Contains(stderr, `Edit of "protected.txt" refused`) {
			t.Fatalf("protected.txt: exit %d, stderr %q", code, stderr)
		}
		if r := lastRecord(t, proj, "s-prot"); r.String("bonsai_path") != filepath.ToSlash(normal) || len(r.String("bonsai_sha256")) != 64 {
			t.Errorf("the record names %s %s, want %s and a hash", r.String("bonsai_path"), r.String("bonsai_sha256"), filepath.ToSlash(normal))
		}
		// The normal build has no switch: set or not, it decides the same.
		if code, stderr, _ := runLine(t, normalPath, proj, "crash", free); code != 0 {
			t.Fatalf("normal build with the crash switch set: exit %d, stderr %q", code, stderr)
		}
	})
	t.Run("missing", func(t *testing.T) {
		code, stderr, _ := runLine(t, t.TempDir(), proj, "missing", free)
		if code != 2 || !strings.Contains(stderr, "bonsai") || !strings.Contains(stderr, "not found") {
			t.Fatalf("exit %d, stderr %q; want 2 and the shell's not found", code, stderr)
		}
	})
	t.Run("minimal-path", func(t *testing.T) {
		code, stderr, _ := runLine(t, minimalPath(t), proj, "minimal-path", free)
		if code != 2 || !strings.Contains(stderr, "bonsai") || !strings.Contains(stderr, "not found") {
			t.Fatalf("exit %d, stderr %q; want 2 and the shell's not found", code, stderr)
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
		code, stderr, _ := runLine(t, faultPath, proj, "crash", editPayload(t, "s-crash", proj, "free.txt"))
		if code != 2 || !strings.Contains(stderr, "test fault crash") {
			t.Fatalf("exit %d, stderr %q", code, stderr)
		}
		if r := lastRecord(t, proj, "s-crash"); r.String("rule") != "fault-"+"crash" || r.String("bonsai_path") != filepath.ToSlash(fault) {
			t.Errorf("record: %s", schema.Show(r))
		}
	})
	t.Run("slow", func(t *testing.T) {
		code, stderr, took := runLine(t, faultPath, proj, "slow", editPayload(t, "s-slow", proj, "free.txt"))
		if code != 2 || !strings.Contains(stderr, "the guard ran past its own 5s limit") {
			t.Fatalf("exit %d, stderr %q", code, stderr)
		}
		if took > 9*time.Second {
			t.Errorf("blocked after %v: past the hook line's 10 s timeout, Claude Code would have let it through", took)
		}
		if r := lastRecord(t, proj, "s-slow"); r.String("rule") != "over-time" {
			t.Errorf("record: %s", schema.Show(r))
		}
	})
	t.Run("a hidden bonsai found anyway", func(t *testing.T) {
		for _, f := range []string{"missing", "minimal-path"} {
			code, stderr, _ := runLine(t, faultPath, proj, f, free)
			if code != 2 || !strings.Contains(stderr, "the launcher did not hide it") {
				t.Fatalf("%s: exit %d, stderr %q", f, code, stderr)
			}
		}
	})
}

func TestHookCommand(t *testing.T) {
	cases := []struct {
		args     []string
		code     int
		inStdout string
		inStderr string
	}{
		{[]string{"hook", "--help"}, 0, "bonsai hook guard ", ""},
		{[]string{"hook", "guard", "--help"}, 0, "bonsai hook guard || exit 2", ""},
		{[]string{"hook"}, 2, "", "bonsai: bonsai hook needs a hook name.\nnext: run `bonsai hook --help`"},
		{[]string{"hook", "guard", "--json"}, 2, "", `bonsai: hook guard takes no "--json".`},
		{[]string{"hook", "stop"}, 2, "", "bonsai hook stop is not built yet"},
		{[]string{"hook", "nap"}, 2, "", `bonsai: "nap" is not a hook.`},
	}
	for _, c := range cases {
		var stdout, stderr bytes.Buffer
		code := run(c.args, &stdout, &stderr)
		if code != c.code || !strings.Contains(stdout.String(), c.inStdout) || !strings.Contains(stderr.String(), c.inStderr) {
			t.Errorf("%v: exit %d, stdout %q, stderr %q", c.args, code, stdout.String(), stderr.String())
		}
		asciiOnly(t, "stdout", stdout.String())
		asciiOnly(t, "stderr", stderr.String())
		if code != 0 && !strings.Contains(stderr.String(), "\nnext: ") {
			t.Errorf("%v: a refusal with no next step", c.args)
		}
	}
}
