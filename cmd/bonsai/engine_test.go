package main

// bonsai init, update and check from the command line: the flags, the consent rule (no prompt without a terminal,
// y/N at one), the exit codes, --json, ASCII output and a next step on every refusal.

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/engine"
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

// linkArgs links the project to the test pack at ref, with --allow-exec: every commit of the test pack has a hook
// line, a pack's code, which a first link writes only with it (plan-5, piece 5.1.1, rule 6).
func (c *cli) linkArgs(ref string) []string {
	return []string{"init", "--name", "demo", "--source", c.pack.Source, "--ref", ref, "--never-edit", "ledger.json", "--allow-exec"}
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
		!strings.Contains(out, "next: to write it, run: bonsai init --name demo --source ") ||
		!strings.HasSuffix(out, "--never-edit ledger.json --allow-exec --yes\n") {
		t.Errorf("init without --yes: %d\n%s", code, out)
	}
	if len(c.files()) != 0 {
		t.Fatalf("init without --yes wrote %v", c.files())
	}
	// Without --allow-exec, the pack's hook line refuses the link, at a terminal too: no y/N for code.
	noExec := c.linkArgs(c.pack.A)[:len(c.linkArgs(c.pack.A))-1]
	code, out, _ = c.run("y\n", noExec...)
	if code != 4 || strings.Contains(out, "[y/N]") || !strings.Contains(out, "Runs code: 1 item, which needs --allow-exec as well as --yes") ||
		!strings.Contains(out, "add     hook    SessionStart (startup): echo demo hook A  (demo-pack)") ||
		!strings.HasSuffix(out, "--never-edit ledger.json --allow-exec --yes\n") || len(c.files()) != 0 {
		t.Errorf("init without --allow-exec at a terminal: %d\n%s", code, out)
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
	code, out, _ = c.run("", "update", "--diff")
	if code != 4 || !strings.Contains(out, "--- a/demo/guide.md\n+++ b/demo/guide.md\n@@ -1,3 +1,3 @@\n # Guide\n \n-Edition 1.\n+Edition 2.\n") ||
		!strings.Contains(out, "--- /dev/null\n+++ b/demo/extra.md\n") || !strings.Contains(out, "--- a/.claude/settings.json\n") {
		t.Errorf("update --diff: %d\n%s", code, out)
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
	// A conflict: without --yes, the preview names the commands; exit 5 with --yes, and the commands to paste.
	if err := os.WriteFile(filepath.Join(c.root, "demo", "guide.md"), []byte("mine\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	code, out, _ = c.run("", "update")
	if code != 4 || !strings.HasSuffix(out, "Nothing written yet.\nnext: to keep your edits, run: bonsai update --yes --keep demo/guide.md\n"+
		"  or, to take the pack's copies (yours are saved in the Bonsai home's cache, never in the repo), run: bonsai update --yes --adopt demo/guide.md\n") {
		t.Errorf("a conflict without --yes: %d\n%s", code, out)
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
	if code, out, _ := c.run("", "check"); code != 1 || !onlySource(out) {
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
	if code != 4 || !strings.Contains(out, "Refused: this update writes code that runs on this machine (1 item under Runs code), which needs --allow-exec as well as --yes: nothing was written.") ||
		!strings.Contains(out, "next: read each item under Runs code; to write them with the rest, run: bonsai update --allow-exec --yes\n") {
		t.Errorf("a hook-line change: %d\n%s", code, out)
	}
	if after := c.snapshot(); after != before {
		t.Errorf("a refused hook-line change wrote something")
	}
	code, out, _ = c.run("", "update", "--yes", "--allow-exec")
	if code != 0 || !strings.Contains(out, "Runs code: 1 item, consented to with --allow-exec") || !strings.Contains(out, "bonsai update: written") ||
		!strings.Contains(c.snapshot(), "echo demo hook D") {
		t.Errorf("--allow-exec: %d\n%s", code, out)
	}
}

// onlySource reports whether check's text holds one finding, and no warning: the one every test project has, its
// bonsai.yaml naming the test pack by a local folder, an absolute path (absolute-path; spec §14 check 2).
func onlySource(out string) bool {
	return strings.HasPrefix(out, "bonsai check: 1 finding.\n  bonsai.yaml's field packs[0].source holds the absolute path ") &&
		strings.Contains(out, "next: a person names the pack's remote URL") && !strings.Contains(out, "warning:")
}

// codesIn lists the codes of check --json's findings or warnings, space-separated.
func codesIn(doc schema.Object, list string) string {
	v, _ := doc.Get(list)
	var out []string
	for _, f := range v.([]any) {
		out = append(out, f.(schema.Object).String("code"))
	}
	return strings.Join(out, " ")
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
	if code != 1 || !strings.Contains(out, "demo/guide.md was edited") || !strings.Contains(out, "next: to keep the edit, run: bonsai update --yes --keep demo/guide.md") {
		t.Errorf("check: %d\n%s", code, out)
	}
	code, out, _ = c.run("", "check", "--json")
	if code != 1 || codesIn(fits(t, out, "check"), "findings") != "changed absolute-path" {
		t.Errorf("check --json: %d %s", code, out)
	}
	for _, args := range [][]string{{"check", "now"}} {
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

// With bonsai.yaml but no lock, update --yes is refused (exit 4) and names bonsai init: deleting the lock is no way
// round the refusal of a hook-line change (pack commit D changes one).
func TestUpdateWithoutALock(t *testing.T) {
	c := newCLI(t)
	if code, _, _ := c.run("", append(c.linkArgs(c.pack.C), "--yes")...); code != 0 {
		t.Fatal("link failed")
	}
	if err := os.Remove(filepath.Join(c.root, ".bonsai", "lock.json")); err != nil {
		t.Fatal(err)
	}
	testpack.SetRef(t, c.root, c.pack.C, c.pack.D)
	before := c.snapshot()
	code, out, errOut := c.run("", "update", "--yes")
	if code != 4 || !strings.Contains(out+errOut, "bonsai.yaml is here but .bonsai/lock.json is not") ||
		!strings.Contains(out+errOut, "run bonsai init") {
		t.Errorf("update --yes with no lock: %d\n%s%s", code, out, errOut)
	}
	if after := c.snapshot(); after != before {
		t.Errorf("update with no lock wrote:\n%s\n---\n%s", before, after)
	}
}

// fakePlugins stands in for Claude Code's plugin commands (internal/engine/plugins.go).
type fakePlugins struct {
	list    []engine.InstalledPlugin
	install engine.InstallResult
	markets []string
	asked   []string
}

func (f *fakePlugins) List(dir string) ([]engine.InstalledPlugin, error) {
	f.asked = append(f.asked, "list")
	return f.list, nil
}

func (f *fakePlugins) Install(dir, plugin string) (engine.InstallResult, error) {
	f.asked = append(f.asked, "install "+plugin)
	return f.install, nil
}

func (f *fakePlugins) Uninstall(dir, plugin string) (engine.InstallResult, error) {
	f.asked = append(f.asked, "uninstall "+plugin)
	return engine.InstallResult{Outcome: "ok"}, nil
}

// Marketplaces answers with markets; nil: Claude Code could not say.
func (f *fakePlugins) Marketplaces(dir string) ([]string, error) {
	f.asked = append(f.asked, "marketplaces")
	if f.markets == nil {
		return nil, errors.New("not asked here")
	}
	return f.markets, nil
}

// init and update install each pack's plugin after writing (or with nothing to write), never before and never on a
// refusal; the outcome never changes the exit code; check reports drift with exit 1; status shows the offline half.
func TestPluginStep(t *testing.T) {
	c := newCLI(t)
	f := &fakePlugins{install: engine.InstallResult{Outcome: "failed", FailureCode: "not_found"}}
	defer func(p engine.PluginCLI) { pluginCLI = p }(pluginCLI)
	pluginCLI = f
	marketA := engine.MarketplaceName("demo", []string{c.pack.A})
	marketB := engine.MarketplaceName("demo", []string{c.pack.B})

	if code, _, _ := c.run("", c.linkArgs(c.pack.A)...); code != 4 || len(f.asked) != 0 {
		t.Fatalf("a preview asked Claude Code: %d %v", code, f.asked)
	}
	code, out, _ := c.run("", append(c.linkArgs(c.pack.A), "--yes")...)
	if code != 0 || strings.Join(f.asked, ",") != "install demo-pack@"+marketA+",list" ||
		!strings.Contains(out, "This machine's plugins (Claude Code, scope project: this checkout's .claude/settings.json):\n  waiting      demo-pack@"+marketA+": ") ||
		!strings.Contains(out, "\n               next: a person opens Claude Code in this checkout and accepts its trust question") ||
		!strings.Contains(out, "; then run: bonsai update\n") || c.exists(".claude/settings.local.json") {
		t.Errorf("init --yes: %d %v\n%s", code, f.asked, out)
	}
	// Nothing to change: installed now.
	f.asked, f.install = nil, engine.InstallResult{Outcome: "ok"}
	code, out, _ = c.run("", "update", "--json")
	doc, err := schema.Decode([]byte(out))
	if code != 0 || err != nil {
		t.Fatalf("update --json: %d %v %s", code, err, out)
	}
	plugins, _ := doc.(schema.Object).Get("plugins")
	if schema.Show(plugins) != `[{"pack":"demo-pack","plugin":"demo-pack@`+marketA+`","commit":"`+c.pack.A+`","result":"installed","message":"at `+c.pack.A[:12]+`","next":null}]` {
		t.Errorf("update --json's plugins: %s", schema.Show(plugins))
	}
	// A refused update asks nothing.
	testpack.SetRef(t, c.root, c.pack.A, c.pack.B)
	f.asked = nil
	if code, _, _ := c.run("", "update"); code != 4 || len(f.asked) != 0 {
		t.Errorf("update without --yes asked Claude Code: %d %v", code, f.asked)
	}
	// Written to B: B's plugin installed, then A's record for this checkout, under the older marketplace name, removed
	// (it is no longer turned on); another workspace's record, another checkout's and a local one are left.
	f.list = []engine.InstalledPlugin{
		{ID: "demo-pack@" + marketA, Version: c.pack.A[:12], Scope: "project", ProjectPath: c.root},
		{ID: "demo-pack@bonsai-other-0123abcd", Version: c.pack.A[:12], Scope: "project", ProjectPath: c.root},
		{ID: "demo-pack@" + marketA, Version: c.pack.A[:12], Scope: "project", ProjectPath: c.root + "-elsewhere"},
		{ID: "demo-pack@" + marketA, Version: c.pack.A[:12], Scope: "local", ProjectPath: c.root},
	}
	code, out, _ = c.run("", "update", "--yes", "--json")
	plugins, _ = fits(t, out, "changes").Get("plugins")
	if code != 0 || strings.Join(f.asked, ",") != "install demo-pack@"+marketB+",list,uninstall demo-pack@"+marketA ||
		!strings.Contains(schema.Show(plugins), `{"pack":"demo-pack","plugin":"demo-pack@`+marketA+`","commit":"`+c.pack.B+`","result":"uninstalled","message":"a stale record (at `+c.pack.A[:12]+`, from an older marketplace name of this workspace) removed for this checkout; the lock's plugin is demo-pack@`+marketB+`","next":null}`) {
		t.Errorf("update --yes to B: %d %v\n%s", code, f.asked, schema.Show(plugins))
	}
	// Nothing to change: no stale record looked for.
	f.asked = nil
	if code, _, _ := c.run("", "update"); code != 0 || strings.Contains(strings.Join(f.asked, ","), "uninstall") {
		t.Errorf("update with nothing to change: %d %v", code, f.asked)
	}
	// check: A's plugin still on for this checkout is drift (exit 1); B's installed is not.
	f.list = []engine.InstalledPlugin{
		{ID: "demo-pack@" + marketB, Version: c.pack.B[:12], Scope: "local", Enabled: true, ProjectPath: c.root},
		{ID: "demo-pack@" + marketA, Version: c.pack.A[:12], Scope: "local", Enabled: true, ProjectPath: c.root},
	}
	code, out, _ = c.run("", "check")
	if code != 1 || !strings.Contains(out, "Claude Code turns on the plugin demo-pack@"+marketA) {
		t.Errorf("check with drift: %d\n%s", code, out)
	}
	f.list = f.list[:1]
	if code, out, _ := c.run("", "check"); code != 1 || !onlySource(out) {
		t.Errorf("check with no drift: %d %q", code, out)
	}
	// The offline half: a settings.local.json (a local-scope install's, never Bonsai's) turning on A's plugin.
	// status shows it among its problems, with no question to Claude Code.
	if err := os.WriteFile(filepath.Join(c.root, ".claude", "settings.local.json"),
		[]byte(`{"enabledPlugins": {"demo-pack@`+marketA+`": true}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	f.asked = nil
	code, out, _ = c.run("", "status", "--json")
	if code != 0 || !strings.Contains(out, "turns on demo-pack@"+marketA) || len(f.asked) != 0 {
		t.Errorf("status: %d %v\n%s", code, f.asked, out)
	}
}

func (c *cli) exists(rel string) bool {
	_, err := os.Stat(filepath.Join(c.root, filepath.FromSlash(rel)))
	return err == nil
}

// A pack's plugin that carries code (the test pack from F: a hook of the plugin itself) installs on this machine only
// with --allow-exec, on each machine: a git-pulled lock's update with nothing to change leaves it waiting, exit 0
// (the plugin step never changes the exit code), naming the plugin, its code part and the command; with
// --allow-exec it installs; once installed, update asks nothing more.
func TestPluginStepWithCode(t *testing.T) {
	c := newCLI(t)
	f := &fakePlugins{install: engine.InstallResult{Outcome: "ok"}}
	defer func(p engine.PluginCLI) { pluginCLI = p }(pluginCLI)
	pluginCLI = f
	market := engine.MarketplaceName("demo", []string{c.pack.F})
	want := "demo-pack@" + market
	if code, _, _ := c.run("", append(c.linkArgs(c.pack.F), "--yes")...); code != 0 || strings.Join(f.asked, ",") != "install "+want+",list" {
		t.Fatalf("the link with --allow-exec: %d, asked %v", code, f.asked)
	}
	f.asked = nil
	code, out, _ := c.run("", "update")
	if code != 0 || !strings.Contains(out, "  waiting      "+want+": the plugin carries code Claude Code runs on its own (hooks/hooks.json)") ||
		!strings.Contains(out, "next: if you consent to that code running on this machine, run: bonsai update --allow-exec\n") ||
		strings.Join(f.asked, ",") != "list" {
		t.Errorf("update without --allow-exec: %d, asked %v\n%s", code, f.asked, out)
	}
	code, out, _ = c.run("", "update", "--json")
	if doc, err := schema.Decode([]byte(out)); code != 0 || err != nil || !strings.Contains(out, `"result": "waiting"`) {
		t.Errorf("update --json: %d %v %s", code, err, schema.Show(doc))
	}
	f.asked = nil
	if code, out, _ := c.run("", "update", "--allow-exec"); code != 0 || !strings.Contains(out, "installed    "+want) ||
		strings.Join(f.asked, ",") != "install "+want {
		t.Errorf("update --allow-exec: %d, asked %v\n%s", code, f.asked, out)
	}
	f.asked, f.list = nil, []engine.InstalledPlugin{{ID: want, Version: c.pack.F[:12], Scope: "project", Enabled: true, ProjectPath: c.root}}
	if code, out, _ := c.run("", "update"); code != 0 || !strings.Contains(out, "already installed for this checkout") ||
		strings.Join(f.asked, ",") != "list" {
		t.Errorf("update once installed: %d, asked %v\n%s", code, f.asked, out)
	}
}

// A pack taken out of bonsai.yaml (step 5.1.7): update previews it (exit 4 with no --yes), and with --yes takes it
// out, then removes Claude Code's record of its plugin's install for this checkout (after its own write) and installs
// the remaining pack's plugin; check is clean before and after but for the step it names.
func TestUpdateTakesAPackOut(t *testing.T) {
	c := newCLI(t)
	side, sc := testpack.SidePack(t, c.tmp)
	c.write("bonsai.yaml", "format: bonsai.workspace/1\nid: ws-aaaaaaaaaaaaaaaaaaaaaaaaaa\nname: demo\npacks:\n"+
		"  - id: demo-pack\n    source: \""+filepath.ToSlash(c.pack.Source)+"\"\n    ref: \""+c.pack.A+"\"\n"+
		"  - id: side-pack\n    source: \""+filepath.ToSlash(side)+"\"\n    ref: \""+sc[0]+"\"\n")
	if code, out, errOut := c.run("", "init", "--yes", "--allow-exec"); code != 0 {
		t.Fatalf("link: %d\n%s%s", code, out, errOut)
	}
	oldMarket := engine.MarketplaceName("demo", []string{c.pack.A, sc[0]})
	newMarket := engine.MarketplaceName("demo", []string{c.pack.A})
	f := &fakePlugins{install: engine.InstallResult{Outcome: "ok"}, list: []engine.InstalledPlugin{
		{ID: "side-pack@" + oldMarket, Version: sc[0][:12], Scope: "project", ProjectPath: c.root},
		{ID: "demo-pack@" + oldMarket, Version: c.pack.A[:12], Scope: "project", ProjectPath: c.root},
	}}
	defer func(p engine.PluginCLI) { pluginCLI = p }(pluginCLI)
	pluginCLI = f
	cfg := c.read("bonsai.yaml")
	c.write("bonsai.yaml", cfg[:strings.Index(cfg, "  - id: side-pack")])

	code, out, _ := c.run("", "update")
	if code != 4 || !strings.Contains(out, "side-pack 0.2.0  "+sc[0][:7]+" -> taken out") || !strings.Contains(out, "next: to write it, run: bonsai update --yes") ||
		len(f.asked) != 0 || !c.exists("side/a.md") {
		t.Errorf("update with no --yes: %d %v\n%s", code, f.asked, out)
	}
	code, out, _ = c.run("", "update", "--yes", "--json")
	doc := fits(t, out, "changes")
	if code != 0 || doc.String("result") != "applied" || c.exists("side/a.md") ||
		strings.Join(f.asked, ",") != "list,uninstall side-pack@"+oldMarket+",install demo-pack@"+newMarket+",list,uninstall demo-pack@"+oldMarket {
		t.Errorf("update --yes: %d %v\n%s", code, f.asked, out)
	}
	plugins, _ := doc.Get("plugins")
	if !strings.Contains(schema.Show(plugins), `{"pack":"side-pack","plugin":"side-pack@`+oldMarket+`","commit":"`+sc[0]+`","result":"uninstalled",`) {
		t.Errorf("plugins %s", schema.Show(plugins))
	}
	f.list = []engine.InstalledPlugin{{ID: "demo-pack@" + newMarket, Version: c.pack.A[:12], Scope: "project", ProjectPath: c.root}}
	if code, out, _ := c.run("", "check"); code != 1 || !onlySource(out) {
		t.Errorf("check after: %d\n%s", code, out)
	}
}
