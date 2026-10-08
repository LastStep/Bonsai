package main

// bonsai init, update and check from the command line: the flags, the consent rule (no prompt without a terminal,
// y/N at one), the exit codes, --json, ASCII output and a next step on every refusal.

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/testpack"
)

type cli struct {
	t    *testing.T
	tmp  string
	pack *testpack.Pack
	root string
}

func newCLI(t *testing.T) *cli {
	t.Helper()
	tmp := t.TempDir()
	if r, err := filepath.EvalSymlinks(tmp); err == nil {
		tmp = r
	}
	testpack.Isolate(t, tmp)
	c := &cli{t: t, tmp: tmp, pack: testpack.Build(t, tmp)}
	c.root = testpack.Project(t, tmp, "project")
	t.Chdir(c.root)
	return c
}

// run runs bonsai with a non-terminal stdin, unless answer is given: then it runs as at a terminal, answering it.
func (c *cli) run(answer string, args ...string) (int, string, string) {
	c.t.Helper()
	defer func(i func() bool, r io.Reader) { interactive, stdin = i, r }(interactive, stdin)
	interactive = func() bool { return answer != "" }
	stdin = strings.NewReader(answer)
	var stdout, stderr bytes.Buffer
	code := run(args, &stdout, &stderr)
	asciiOnly(c.t, "stdout", stdout.String())
	asciiOnly(c.t, "stderr", stderr.String())
	return code, stdout.String(), stderr.String()
}

func (c *cli) linkArgs(ref string) []string {
	return []string{"init", "--name", "demo", "--source", c.pack.Source, "--ref", ref, "--never-edit", "ledger.json"}
}

func (c *cli) files() []string {
	var out []string
	_ = filepath.Walk(c.root, func(p string, info os.FileInfo, err error) error {
		if err == nil && info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}
		if err == nil && !info.IsDir() {
			rel, _ := filepath.Rel(c.root, p)
			out = append(out, filepath.ToSlash(rel))
		}
		return nil
	})
	return out
}

func TestInitCommand(t *testing.T) {
	c := newCLI(t)
	code, out, errOut := c.run("", "init")
	if code != 2 || out != "" || !strings.Contains(errOut, "init needs --name, --source and --ref") || !strings.Contains(errOut, "\nnext: run bonsai init --name") {
		t.Errorf("init with no values: %d %q %q", code, out, errOut)
	}
	// No terminal, no --yes: the preview, exit 4, nothing written, the command with --yes named.
	code, out, _ = c.run("", c.linkArgs(c.pack.A)...)
	if code != 4 || !strings.Contains(out, "bonsai init: the preview.") || !strings.Contains(out, "add     hook        PreToolUse") ||
		!strings.Contains(out, "next: to write it, run: bonsai init --name demo --source ") || !strings.HasSuffix(out, "--never-edit ledger.json --yes\n") {
		t.Errorf("init without --yes: %d\n%s", code, out)
	}
	if len(c.files()) != 0 {
		t.Fatalf("init without --yes wrote %v", c.files())
	}
	// At a terminal: y/N; n writes nothing.
	code, out, _ = c.run("n\n", c.linkArgs(c.pack.A)...)
	if code != 4 || !strings.Contains(out, "Write these changes? [y/N] Nothing written.\n") || len(c.files()) != 0 {
		t.Errorf("init answered n: %d\n%s", code, out)
	}
	code, out, _ = c.run("y\n", c.linkArgs(c.pack.A)...)
	if code != 0 || !strings.Contains(out, "bonsai init: written; demo-pack at ") || !strings.Contains(out, "Linked demo. Workspace id: ws-") ||
		!strings.HasSuffix(out, "A copy meant as a new project needs its own id: bonsai init --new-id\n") {
		t.Errorf("init answered y: %d\n%s", code, out)
	}
	// Again, with --yes: nothing to change, and the closing words.
	code, out, _ = c.run("", append(c.linkArgs(c.pack.A), "--yes")...)
	if code != 0 || !strings.HasPrefix(out, "bonsai init: nothing to change; demo-pack at ") || !strings.Contains(out, "Linked demo.") {
		t.Errorf("init again: %d\n%s", code, out)
	}
	// --json: one document, ASCII, the same facts.
	code, out, _ = c.run("", "init", "--json")
	doc, err := schema.Decode([]byte(out))
	if code != 0 || err != nil || doc.(schema.Object).String("result") != "nothing" {
		t.Errorf("init --json: %d %v\n%s", code, err, out)
	}
}

func TestUpdateCommand(t *testing.T) {
	c := newCLI(t)
	if code, _, _ := c.run("", append(c.linkArgs(c.pack.A), "--yes")...); code != 0 {
		t.Fatal("link failed")
	}
	testpack.SetRef(t, c.root, c.pack.A, c.pack.B)
	code, out, _ := c.run("", "update")
	if code != 4 || !strings.Contains(out, "change  marketplace bonsai-demo-") || !strings.HasSuffix(out, "next: to write it, run: bonsai update --yes\n") {
		t.Errorf("update without --yes: %d\n%s", code, out)
	}
	code, out, _ = c.run("", "update", "--json")
	doc, err := schema.Decode([]byte(out))
	if code != 4 || err != nil {
		t.Fatalf("update --json: %d %v %s", code, err, out)
	}
	settings, _ := doc.(schema.Object).Get("settings")
	if list := settings.([]any); len(list) != 2 || list[0].(schema.Object).String("file") != ".claude/settings.json" ||
		list[0].(schema.Object).String("why") == "" {
		t.Errorf("settings lines in the JSON: %s", schema.Show(settings))
	}
	// A conflict: exit 5 with --yes, and the commands to paste.
	if err := os.WriteFile(filepath.Join(c.root, "demo", "guide.md"), []byte("mine\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	code, out, _ = c.run("", "update", "--yes")
	if code != 5 || !strings.Contains(out, "Stopped: 1 conflict (demo/guide.md): nothing was written.") ||
		!strings.Contains(out, "run: bonsai update --yes --keep demo/guide.md\n") {
		t.Errorf("a conflict: %d\n%s", code, out)
	}
	// At a terminal: the preview ends in y/N; y then stops on the conflict.
	code, out, _ = c.run("y\n", "update")
	if code != 5 || !strings.Contains(out, "Write these changes? [y/N] Stopped: 1 conflict") {
		t.Errorf("a conflict at a terminal: %d\n%s", code, out)
	}
	// The pasted command works.
	code, out, _ = c.run("", "update", "--yes", "--keep", "demo/guide.md")
	if code != 0 || !strings.Contains(out, "kept         demo/guide.md") {
		t.Errorf("--keep: %d\n%s", code, out)
	}
	if code, out, _ := c.run("", "check"); code != 0 || out != "bonsai check: no findings.\n" {
		t.Errorf("check: %d %q", code, out)
	}
	// C to D changes only the hook line: refused, exit 4, --allow-exec named, nothing written.
	testpack.SetRef(t, c.root, c.pack.B, c.pack.C)
	if code, _, _ := c.run("", "update", "--yes", "--adopt", "demo/guide.md"); code != 0 {
		t.Fatal("to C failed")
	}
	testpack.SetRef(t, c.root, c.pack.C, c.pack.D)
	before := c.snapshot()
	code, out, _ = c.run("", "update", "--yes")
	if code != 4 || !strings.Contains(out, "Refused: this update changes a hook line") ||
		!strings.Contains(out, "next: a hook-line change needs --allow-exec as well as --yes: bonsai update --yes --allow-exec") {
		t.Errorf("a hook-line change: %d\n%s", code, out)
	}
	if after := c.snapshot(); after != before {
		t.Errorf("a refused hook-line change wrote something")
	}
	code, _, errOut := c.run("", "update", "--yes", "--allow-exec")
	if code != 2 || !strings.Contains(errOut, "--allow-exec is not built yet: it comes with step 5.1") || !strings.Contains(errOut, "\nnext: ") {
		t.Errorf("--allow-exec: %d %q", code, errOut)
	}
}

func (c *cli) snapshot() string {
	var b strings.Builder
	for _, f := range c.files() {
		raw, _ := os.ReadFile(filepath.Join(c.root, filepath.FromSlash(f)))
		b.WriteString(f + "\n" + string(raw) + "\n")
	}
	return b.String()
}

func TestCheckCommand(t *testing.T) {
	c := newCLI(t)
	code, _, errOut := c.run("", "check")
	if code != 4 || !strings.Contains(errOut, "next: run bonsai init") {
		t.Errorf("check unlinked: %d %q", code, errOut)
	}
	c.run("", append(c.linkArgs(c.pack.A), "--yes")...)
	if err := os.WriteFile(filepath.Join(c.root, "demo", "guide.md"), []byte("mine\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	code, out, _ := c.run("", "check")
	if code != 1 || !strings.Contains(out, "demo/guide.md was edited") || !strings.Contains(out, "next: keep the edit: bonsai update --yes --keep demo/guide.md") {
		t.Errorf("check: %d\n%s", code, out)
	}
	code, out, _ = c.run("", "check", "--json")
	doc, err := schema.Decode([]byte(out))
	if code != 1 || err != nil || doc.(schema.Object).String("result") != "findings" {
		t.Errorf("check --json: %d %v %s", code, err, out)
	}
	for _, args := range [][]string{{"check", "--write"}, {"check", "now"}} {
		if code, _, errOut := c.run("", args...); code != 2 || !strings.Contains(errOut, "\nnext: ") {
			t.Errorf("%v: %d %q", args, code, errOut)
		}
	}
}

func TestEngineHelpAndFlags(t *testing.T) {
	for _, word := range []string{"init", "update", "check"} {
		var stdout, stderr bytes.Buffer
		if code := run([]string{word, "--help"}, &stdout, &stderr); code != 0 || stderr.Len() != 0 {
			t.Errorf("%s --help: %d", word, code)
		}
		out := stdout.String()
		asciiOnly(t, word+" --help", out)
		if !strings.Contains(out, "Exit codes:") || !strings.Contains(out, "Example: bonsai "+word) || !strings.Contains(out, "--json") {
			t.Errorf("%s --help:\n%s", word, out)
		}
	}
	for _, args := range [][]string{{"update", "--name", "x"}, {"init", "--name"}, {"update", "--yes=1"}, {"update", "now"}, {"init", "--keep"}} {
		var stdout, stderr bytes.Buffer
		if code := run(args, &stdout, &stderr); code != 2 || !strings.Contains(stderr.String(), "\nnext: ") {
			t.Errorf("%v: %d %q", args, code, stderr.String())
		}
	}
}
