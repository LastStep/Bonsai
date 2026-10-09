package main

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
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
func TestRun(t *testing.T) {
	cases := []struct {
		name     string
		args     []string
		code     int
		stdout   string
		inStderr string
		next     string
	}{
		{"version", []string{"--version"}, 0, "bonsai dev\n", "", ""},
		{"no command", nil, 2, "", "bonsai: no command given:", "\nnext: run `bonsai --help`"},
		{"unknown command", []string{"unlink"}, 2, "", `bonsai: "unlink" is not a command yet:`, "\nnext: run `bonsai --help`"},
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
