package main

import (
	"bytes"
	"debug/buildinfo"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/status"
)

// asciiOnly fails on any byte that is not printable ASCII or a line feed.
func asciiOnly(t *testing.T, what, s string) {
	t.Helper()
	for i := 0; i < len(s); i++ {
		if s[i] > 0x7e || (s[i] < 0x20 && s[i] != '\n') {
			t.Fatalf("%s byte %d is %#x, not printable ASCII: %q", what, i, s[i], s)
		}
	}
}

// The words this build answers, and the refusals of everything else: exit 2, ASCII lines on stderr, a next step.
// The build stands in as one with no commit stamp, so --version's line is the same wherever the test runs.
func TestRun(t *testing.T) {
	defer func(r func() (*debug.BuildInfo, bool)) { readBuildInfo = r }(readBuildInfo)
	readBuildInfo = func() (*debug.BuildInfo, bool) { return &debug.BuildInfo{}, true }
	cases := []struct {
		name     string
		args     []string
		code     int
		stdout   string
		inStderr string
		next     string
	}{
		{"version", []string{"--version"}, 0, "bonsai dev (no commit stamp)\n", "", ""},
		{"no command", nil, 2, "", "bonsai: no command given:", "\nnext: run `bonsai --help`"},
		{"unknown command", []string{"ladder"}, 2, "", `bonsai: "ladder" is not a command yet:`, "\nnext: run `bonsai --help`"},
		{"version with more", []string{"--version", "now"}, 2, "", "bonsai: --version takes no other argument:", "\nnext: run `bonsai --help`"},
		{"help with more", []string{"--help", "status"}, 2, "", "bonsai: --help takes no other argument:", "\nnext: run `bonsai --help`"},
		{"non-ASCII argument", []string{"b\xc3\xb6nsai"}, 2, "", `bonsai: "b\` + `u00f6nsai" is not a command yet:`, "\nnext: run `bonsai --help`"},
		{"status --line", []string{"status", "--line", "--help"}, 2, "", "bonsai: status --line is not built yet", "\nnext: run `bonsai status --help` without --line"},
		{"status --active --full", []string{"status", "--active", "--full"}, 2, "", "bonsai: status --active prints only the active task", "\nnext: run `bonsai status --active`"},
		{"status with a word", []string{"status", "now"}, 2, "", `bonsai: status takes no "now"`, "\nnext: run `bonsai status --help`"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := run(c.args, &stdout, &stderr)
			if code != c.code {
				t.Errorf("exit code %d, want %d", code, c.code)
			}
			if stdout.String() != c.stdout {
				t.Errorf("stdout %q, want %q", stdout.String(), c.stdout)
			}
			if c.inStderr == "" {
				if stderr.Len() != 0 {
					t.Errorf("stderr %q, want nothing", stderr.String())
				}
				return
			}
			out := stderr.String()
			if !strings.HasPrefix(out, c.inStderr) {
				t.Errorf("stderr %q, want it to start with %q", out, c.inStderr)
			}
			if !strings.Contains(out, c.next) {
				t.Errorf("stderr %q names no next step %q", out, c.next)
			}
			asciiOnly(t, "stderr", out)
		})
	}
}

// --version names the commit Go stamped, its first 12 hex characters, with +modified for a tree that had changes;
// a build with no stamp, or one Bonsai does not read, says it has none.
func TestVersionLine(t *testing.T) {
	defer func(r func() (*debug.BuildInfo, bool), v string) { readBuildInfo, version = r, v }(readBuildInfo, version)
	const rev = "0123456789abcdef0123456789abcdef01234567"
	for _, c := range []struct {
		name     string
		settings []debug.BuildSetting
		ok       bool
		want     string
	}{
		{"stamped", []debug.BuildSetting{{Key: "vcs", Value: "git"}, {Key: "vcs.revision", Value: rev},
			{Key: "vcs.modified", Value: "false"}}, true, "bonsai 1.0.0 (commit 0123456789ab)\n"},
		{"modified", []debug.BuildSetting{{Key: "vcs.modified", Value: "true"}, {Key: "vcs.revision", Value: rev}}, true,
			"bonsai 1.0.0 (commit 0123456789ab+modified)\n"},
		{"no stamp", []debug.BuildSetting{{Key: "-buildvcs", Value: "false"}}, true, "bonsai 1.0.0 (no commit stamp)\n"},
		{"no build information", nil, false, "bonsai 1.0.0 (no commit stamp)\n"},
		{"a revision too short", []debug.BuildSetting{{Key: "vcs.revision", Value: "0123abc"}}, true, "bonsai 1.0.0 (no commit stamp)\n"},
		{"a revision not hex", []debug.BuildSetting{{Key: "vcs.revision", Value: "0123456789ab\u00e9" + rev}}, true,
			"bonsai 1.0.0 (no commit stamp)\n"},
	} {
		version = "1.0.0"
		readBuildInfo = func() (*debug.BuildInfo, bool) { return &debug.BuildInfo{Settings: c.settings}, c.ok }
		if got := versionLine(); got != c.want {
			t.Errorf("%s: %q, want %q", c.name, got, c.want)
		}
	}
}

// A real build's --version names the commit the binary carries, as Go's own reader of a binary finds it (go version
// -m reads the same): from a git checkout, its commit; and a build made with -buildvcs=false says it has none.
func TestVersionNamesTheBuildsCommit(t *testing.T) {
	normal, _ := bonsaiBuilds(t)
	unstamped := filepath.Join(t.TempDir(), "bonsai")
	if runtime.GOOS == "windows" {
		unstamped += ".exe"
	}
	goTool, err := exec.LookPath("go")
	if err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command(goTool, "build", "-buildvcs=false", "-o", unstamped, ".").CombinedOutput(); err != nil {
		t.Fatalf("building with -buildvcs=false: %v\n%s", err, out)
	}
	for _, c := range []struct{ bin, want string }{{normal, ""}, {unstamped, "bonsai dev (no commit stamp)\n"}} {
		bi, err := buildinfo.ReadFile(c.bin)
		if err != nil {
			t.Fatal(err)
		}
		want := c.want
		if want == "" {
			stamp := "no commit stamp"
			rev, modified := "", ""
			for _, s := range bi.Settings {
				switch s.Key {
				case "vcs.revision":
					rev = s.Value
				case "vcs.modified":
					modified = s.Value
				}
			}
			if len(rev) >= 12 {
				stamp = "commit " + rev[:12]
				if modified == "true" {
					stamp += "+modified"
				}
			}
			want = "bonsai dev (" + stamp + ")\n"
			t.Logf("the test build: %s", strings.TrimSpace(want))
		}
		out, err := exec.Command(c.bin, "--version").Output()
		if err != nil || string(out) != want {
			t.Errorf("%s --version: %q, %v; want %q", filepath.Base(filepath.Dir(c.bin)), out, err, want)
		}
	}
}

func TestHelp(t *testing.T) {
	for _, args := range [][]string{{"--help"}, {"-h"}, {"status", "--help"}, {"status", "--json", "-h"}} {
		var stdout, stderr bytes.Buffer
		if code := run(args, &stdout, &stderr); code != 0 || stderr.Len() != 0 {
			t.Errorf("%v: exit %d, stderr %q", args, code, stderr.String())
		}
		out := stdout.String()
		asciiOnly(t, "help", out)
		if !strings.Contains(out, "Exit codes") || !strings.Contains(out, "Example: bonsai status --json") {
			t.Errorf("%v: the help names no exit codes or example:\n%s", args, out)
		}
	}
}

// The help's commands keep their words in one column, at least two spaces after the longest command (bonsai answer
// <key> [flags] is 27 characters).
func TestHelpColumn(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run([]string{"--help"}, &stdout, &stderr); code != 0 {
		t.Fatalf("exit %d", code)
	}
	col, lines := -1, 0
	for _, line := range strings.Split(stdout.String(), "\n") {
		if !strings.HasPrefix(line, "  bonsai ") {
			continue
		}
		lines++
		i := strings.Index(line[2:], "  ")
		if i < 0 {
			t.Errorf("no two spaces before the words: %q", line)
			continue
		}
		rest := line[2+i:]
		at := 2 + i + len(rest) - len(strings.TrimLeft(rest, " "))
		if col == -1 {
			col = at
		} else if at != col {
			t.Errorf("%q: its words at column %d, the first line's at %d", line, at, col)
		}
	}
	if lines < 10 || !strings.Contains(stdout.String(), "  bonsai answer <key> [flags]  answer an open ask") {
		t.Errorf("the help's commands:\n%s", stdout.String())
	}
}

// An answer that cannot be written exits 3 (runtime), not 0.
func TestRunWriteFails(t *testing.T) {
	var stderr bytes.Buffer
	for _, args := range [][]string{{"--version"}, {"--help"}, {"status", "--help"}} {
		if code := run(args, failingWriter{}, &stderr); code != 3 {
			t.Errorf("%v: exit code %d, want 3", args, code)
		}
	}
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("closed") }

// project makes a git checkout (with bonsai.yaml unless yaml is "") and a temporary Bonsai home, and moves into it.
func project(t *testing.T, yaml string) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not on the PATH: status finds a project through git")
	}
	tmp := t.TempDir()
	empty := filepath.Join(tmp, "gitconfig")
	if err := os.WriteFile(empty, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GIT_CONFIG_GLOBAL", empty)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	t.Setenv("GIT_CEILING_DIRECTORIES", tmp)
	t.Setenv("BONSAI_HOME", filepath.Join(tmp, "home"))
	root := filepath.Join(tmp, "project")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command("git", "-C", root, "init", "-q").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v %s", err, out)
	}
	if yaml != "" {
		if err := os.WriteFile(filepath.Join(root, "bonsai.yaml"), []byte(yaml), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	t.Chdir(root)
	return root
}

const linked = "format: bonsai.workspace/1\nid: ws-aaaaaaaaaaaaaaaaaaaaaaaaaa\nname: demo\nperson_only: [\"bonsai.yaml\"]\n"

func TestStatus(t *testing.T) {
	project(t, linked)
	var stdout, stderr bytes.Buffer
	if code := run([]string{"status", "--json"}, &stdout, &stderr); code != 0 || stderr.Len() != 0 {
		t.Fatalf("exit %d, stderr %q", code, stderr.String())
	}
	asciiOnly(t, "status --json", stdout.String())
	doc, err := schema.Decode(stdout.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if msgs := schema.Validate(status.Schema(), doc); len(msgs) != 0 {
		t.Errorf("status --json does not validate: %v", msgs)
	}
	stdout.Reset()
	if code := run([]string{"status"}, &stdout, &stderr); code != 0 || stderr.Len() != 0 {
		t.Fatalf("exit %d, stderr %q", code, stderr.String())
	}
	asciiOnly(t, "status", stdout.String())
	if !strings.HasPrefix(stdout.String(), "Workspace demo, id ws-aaaaaaaaaaaaaaaaaaaaaaaaaa") {
		t.Errorf("status:\n%s", stdout.String())
	}
}

// Exit 3: --json prints the document with format, bonsai and problems on stdout; plain status prints the problem
// on stderr. Either way the problem names its next step.
func TestStatusExit3(t *testing.T) {
	project(t, "")
	var stdout, stderr bytes.Buffer
	if code := run([]string{"status", "--json"}, &stdout, &stderr); code != 3 || stderr.Len() != 0 {
		t.Fatalf("exit %d, stderr %q", code, stderr.String())
	}
	doc, err := schema.Decode(stdout.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if msgs := schema.Validate(status.Schema(), doc); len(msgs) != 0 {
		t.Errorf("the exit-3 document does not validate: %v", msgs)
	}
	if p, _ := doc.(schema.Object).Get("problems"); !strings.Contains(schema.Show(p), "next: run bonsai init") {
		t.Errorf("problems %s", schema.Show(p))
	}
	stdout.Reset()
	if code := run([]string{"status"}, &stdout, &stderr); code != 3 || stdout.Len() != 0 {
		t.Fatalf("exit %d, stdout %q", code, stdout.String())
	}
	if !strings.HasPrefix(stderr.String(), "bonsai status: cannot read this workspace.\n") ||
		!strings.Contains(stderr.String(), "next: run bonsai init") {
		t.Errorf("stderr %q", stderr.String())
	}
	asciiOnly(t, "stderr", stderr.String())
	if code := run([]string{"status", "--json"}, failingWriter{}, &stderr); code != 3 {
		t.Errorf("an unwritable status exits %d, want 3", code)
	}
}
