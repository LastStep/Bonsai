package engine

// This machine's plugins (plugins.go, plan part 4b), with no network and no Claude Code: a fake PluginCLI stands in
// for `claude plugin list` and `claude plugin install`, and the parsers read output captured from Claude Code 2.1.294.

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/testpack"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// fakeCLI answers as Claude Code would, from its fields, and records what it was asked.
type fakeCLI struct {
	list     []InstalledPlugin
	listErr  error
	install  map[string]InstallResult // by plugin id; missing: ok
	instErr  error
	rewrite  string // a file Install appends a line feed to, as Claude Code rewrites the settings file
	asked    []string
	askedDir []string
}

func (f *fakeCLI) List(dir string) ([]InstalledPlugin, error) {
	f.asked = append(f.asked, "list")
	f.askedDir = append(f.askedDir, dir)
	return f.list, f.listErr
}

func (f *fakeCLI) Install(dir, plugin string) (InstallResult, error) {
	f.asked = append(f.asked, "install "+plugin)
	f.askedDir = append(f.askedDir, dir)
	if f.instErr != nil {
		return InstallResult{}, f.instErr
	}
	if f.rewrite != "" {
		b, err := os.ReadFile(f.rewrite)
		if err != nil {
			return InstallResult{}, err
		}
		if err := os.WriteFile(f.rewrite, append(b, '\n'), 0o644); err != nil {
			return InstallResult{}, err
		}
	}
	if r, ok := f.install[plugin]; ok {
		return r, nil
	}
	return InstallResult{Outcome: "ok"}, nil
}

func localDoc(t *testing.T, root string) schema.Object {
	t.Helper()
	v, err := schema.Decode([]byte(read(t, root, LocalSettingsFile)))
	if err != nil {
		t.Fatal(err)
	}
	return v.(schema.Object)
}

func lockOf(t *testing.T, root string) *workspace.Lock {
	t.Helper()
	l, err := workspace.LoadLock(root)
	if err != nil {
		t.Fatal(err)
	}
	return l
}

// init and update write no .claude/settings.local.json: in a git worktree Claude Code reads the main checkout's too,
// so a plugin turned on there would load in every worktree (plugins.go). The plugin line is the checkout's own
// committed .claude/settings.json's, at the locked commits.
func TestNoLocalSettingsWritten(t *testing.T) {
	e := setup(t)
	root := testpack.Project(t, e.tmp, "local")
	e.link(t, root, e.pack.A)
	testpack.SetRef(t, root, e.pack.A, e.pack.B)
	e.apply(t, root, Request{})
	if _, exists, _ := readFile(root, LocalSettingsFile); exists {
		t.Errorf("%s written", LocalSettingsFile)
	}
	shared, _ := settingsDocOf(t, root).Get("enabledPlugins")
	if schema.Show(shared) != `{"demo-pack@`+MarketplaceName("demo", []string{e.pack.B})+`":true}` {
		t.Errorf("the shared settings' plugin line %s", schema.Show(shared))
	}
	// A person's local file is left as it is, byte for byte.
	writeFile(t, root, LocalSettingsFile, `{"enabledPlugins": {"mine@my-market": true}}`)
	testpack.SetRef(t, root, e.pack.B, e.pack.C)
	e.apply(t, root, Request{Adopt: []string{}})
	if read(t, root, LocalSettingsFile) != `{"enabledPlugins": {"mine@my-market": true}}` {
		t.Errorf("the person's local file changed: %s", read(t, root, LocalSettingsFile))
	}
}

// A main checkout at A and its worktree moved to B: a line turning A's plugin on in the main checkout's
// .claude/settings.local.json (a local-scope install's) is drift in the worktree, which Claude Code makes read it;
// in the main checkout it is the lock's own plugin, no finding.
func TestLocalSettingsDriftInAWorktree(t *testing.T) {
	e := setup(t)
	main := testpack.Project(t, e.tmp, "main")
	e.link(t, main, e.pack.A)
	testpack.Git(t, main, "add", "-A")
	testpack.Git(t, main, "commit", "-q", "-m", "link")
	wt := filepath.Join(e.tmp, "main-worktree")
	testpack.Git(t, main, "worktree", "add", "-q", "-b", "task", wt)
	testpack.SetRef(t, wt, e.pack.A, e.pack.B)
	e.apply(t, wt, Request{})
	marketA := MarketplaceName("demo", []string{e.pack.A})
	writeFile(t, main, LocalSettingsFile, `{"enabledPlugins": {"demo-pack@`+marketA+`": true, "mine@my-market": true}}`)

	r, err := Check(wt, e.home)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Findings) != 1 || r.Findings[0].Code != "plugin" || !strings.Contains(r.Findings[0].Message, "turns on demo-pack@"+marketA) ||
		!strings.Contains(r.Findings[0].Message, "the main checkout's, which Claude Code reads in every worktree") ||
		!strings.Contains(r.Findings[0].Next, "claude plugin uninstall demo-pack@"+marketA+" --scope local") {
		t.Errorf("the worktree: %+v", r.Findings)
	}
	if r, err := Check(main, e.home); err != nil || len(r.Findings) != 0 {
		t.Errorf("the main checkout: %v %+v", err, r.Findings)
	}
	// The worktree's own local file is read too.
	writeFile(t, main, LocalSettingsFile, `{}`)
	writeFile(t, wt, LocalSettingsFile, `{"enabledPlugins": {"demo-pack@`+marketA+`": true}}`)
	if r, err := Check(wt, e.home); err != nil || len(r.Findings) != 1 || !strings.HasPrefix(r.Findings[0].Message, LocalSettingsFile+" turns on") {
		t.Errorf("the worktree's own file: %v %+v", err, r.Findings)
	}
	// A plugin turned off, or the lock's own: no finding. A file Bonsai cannot read: a warning.
	writeFile(t, wt, LocalSettingsFile, `{"enabledPlugins": {"demo-pack@`+marketA+`": false, "demo-pack@`+MarketplaceName("demo", []string{e.pack.B})+`": true}}`)
	if r, err := Check(wt, e.home); err != nil || len(r.Findings) != 0 {
		t.Errorf("turned off: %v %+v", err, r.Findings)
	}
	writeFile(t, wt, LocalSettingsFile, `{`)
	if r, err := Check(wt, e.home); err != nil || len(r.Findings) != 0 || len(r.Warnings) != 1 || r.Warnings[0].Code != "plugin" {
		t.Errorf("unreadable: %v %+v %+v", err, r.Findings, r.Warnings)
	}
}

// InstallPlugins asks Claude Code to install each locked pack's plugin at local scope, in the checkout, and says
// what this machine still needs when it cannot.
func TestInstallPlugins(t *testing.T) {
	e := setup(t)
	root := testpack.Project(t, e.tmp, "install")
	p := e.link(t, root, e.pack.A)
	lock := p.NewLock()
	want := PluginID(testpack.ID, MarketplaceName("demo", []string{e.pack.A}))

	if got := InstallPlugins(root, p.Config, lock, nil); got != nil {
		t.Errorf("a nil CLI: %+v", got)
	}
	f := &fakeCLI{}
	got := InstallPlugins(root, p.Config, lock, f)
	if len(got) != 1 || got[0].Result != "installed" || got[0].Plugin != want || got[0].Commit != e.pack.A ||
		got[0].Message != "at "+e.pack.A[:12] || got[0].Next != "" {
		t.Errorf("installed: %+v", got)
	}
	// Claude Code rewriting the settings file as it installs: said so.
	f = &fakeCLI{rewrite: filepath.Join(root, filepath.FromSlash(SettingsFile))}
	if got := InstallPlugins(root, p.Config, lock, f); len(got) != 1 || got[0].Result != "installed" ||
		!strings.Contains(got[0].Message, "Claude Code wrote .claude/settings.json again in its own key order") {
		t.Errorf("installed, the file rewritten: %+v", got)
	}
	if strings.Join(f.asked, ",") != "install "+want || f.askedDir[0] != root {
		t.Errorf("asked %v in %v", f.asked, f.askedDir)
	}
	f = &fakeCLI{install: map[string]InstallResult{want: {Outcome: "failed", FailureCode: "not_found",
		Message: `Plugin "demo-pack" not found in marketplace`}}}
	if got := InstallPlugins(root, p.Config, lock, f); len(got) != 1 || got[0].Result != "waiting" ||
		!strings.Contains(got[0].Message, "has not registered this checkout's marketplace") || !strings.Contains(got[0].Next, "trust question") ||
		!strings.Contains(got[0].Next, "run bonsai update again") {
		t.Errorf("an unregistered marketplace: %+v", got)
	}
	f = &fakeCLI{install: map[string]InstallResult{want: {Outcome: "failed", FailureCode: "network", Message: "fetch failed\u2026"}}}
	if got := InstallPlugins(root, p.Config, lock, f); len(got) != 1 || got[0].Result != "failed" ||
		got[0].Message != `fetch failed\u2026` || got[0].Next != "run: claude plugin install "+want+" --scope project" {
		t.Errorf("a failed install: %+v", got)
	}
	f = &fakeCLI{instErr: ErrNoClaude}
	if got := InstallPlugins(root, p.Config, lock, f); len(got) != 1 || got[0].Result != "skipped" || got[0].Next == "" {
		t.Errorf("no Claude Code: %+v", got)
	}
	f = &fakeCLI{instErr: errors.New("claude plugin install printed no result line")}
	if got := InstallPlugins(root, p.Config, lock, f); len(got) != 1 || got[0].Result != "failed" ||
		!strings.HasPrefix(got[0].Next, "run: claude plugin install ") {
		t.Errorf("an error: %+v", got)
	}
}

// check's plugin findings: what Claude Code reports installed against the lock (ComparePlugins), and this checkout's
// settings.local.json against it (Check, offline, so status shows it too).
func TestPluginDrift(t *testing.T) {
	e := setup(t)
	root := testpack.Project(t, e.tmp, "drift")
	e.link(t, root, e.pack.A)
	testpack.SetRef(t, root, e.pack.A, e.pack.B)
	e.apply(t, root, Request{})
	marketA := MarketplaceName("demo", []string{e.pack.A})
	marketB := MarketplaceName("demo", []string{e.pack.B})
	atB := InstalledPlugin{ID: "demo-pack@" + marketB, Version: e.pack.B[:12], Scope: "local", Enabled: true, ProjectPath: root}
	atA := InstalledPlugin{ID: "demo-pack@" + marketA, Version: e.pack.A[:12], Scope: "local", Enabled: true, ProjectPath: root}
	other := filepath.Join(e.tmp, "elsewhere")

	check := func(cli PluginCLI) *CheckResult {
		t.Helper()
		r, err := Check(root, e.home)
		if err != nil {
			t.Fatal(err)
		}
		ComparePlugins(r, cli)
		return r
	}
	codes := func(fs []Finding) string {
		var out []string
		for _, f := range fs {
			out = append(out, f.Code)
			if f.Message == "" || f.Next == "" {
				t.Errorf("a finding without its sentence or next step: %+v", f)
			}
		}
		return strings.Join(out, ",")
	}

	// The lock's plugin installed here: nothing to say. Claude Code was asked in the checkout.
	f := &fakeCLI{list: []InstalledPlugin{atB, {ID: "mine@my-market", Version: "1.0.0", Scope: "user", Enabled: true}}}
	if r := check(f); codes(r.Findings) != "" || codes(r.Warnings) != "" || f.askedDir[0] != root {
		t.Errorf("installed at the lock: %+v %+v", r.Findings, r.Warnings)
	}
	// A nil CLI compares nothing (Check alone never runs Claude Code).
	if r := check(nil); codes(r.Findings) != "" || codes(r.Warnings) != "" {
		t.Errorf("a nil CLI: %+v %+v", r.Findings, r.Warnings)
	}
	// An older commit's plugin turned on here beside it: drift, a finding naming both commits.
	r := check(&fakeCLI{list: []InstalledPlugin{atB, atA}})
	if codes(r.Findings) != "plugin" || !strings.Contains(r.Findings[0].Message, marketA+" at "+e.pack.A[:12]) ||
		!strings.Contains(r.Findings[0].Message, "demo-pack at "+e.pack.B[:12]) ||
		!strings.Contains(r.Findings[0].Next, "claude plugin uninstall demo-pack@"+marketA+" --scope local") {
		t.Errorf("an older commit turned on: %+v", r.Findings)
	}
	// Installed but turned off, or installed for another checkout: no drift. The lock's plugin missing: a warning.
	off := atA
	off.Enabled = false
	elsewhere := atB
	elsewhere.ProjectPath = other
	r = check(&fakeCLI{list: []InstalledPlugin{off, elsewhere}})
	if codes(r.Findings) != "" || codes(r.Warnings) != "plugin" || !strings.Contains(r.Warnings[0].Message, "not installed for this checkout") ||
		!strings.HasPrefix(r.Warnings[0].Next, "run bonsai update") {
		t.Errorf("not installed here: %+v %+v", r.Findings, r.Warnings)
	}
	// The plugin installed at the lock's name but another version: drift.
	wrong := atB
	wrong.Version = "0123456789ab"
	if r := check(&fakeCLI{list: []InstalledPlugin{wrong}}); codes(r.Findings) != "plugin" || codes(r.Warnings) != "plugin" {
		t.Errorf("another version under the lock's name: %+v %+v", r.Findings, r.Warnings)
	}
	// Claude Code missing, or failing: a warning, never a finding.
	if r := check(&fakeCLI{listErr: ErrNoClaude}); codes(r.Findings) != "" || codes(r.Warnings) != "plugin" ||
		!strings.Contains(r.Warnings[0].Message, "not on the PATH") {
		t.Errorf("no Claude Code: %+v %+v", r.Findings, r.Warnings)
	}
	if r := check(&fakeCLI{listErr: errors.New("boom")}); codes(r.Findings) != "" || codes(r.Warnings) != "plugin" {
		t.Errorf("a failing list: %+v %+v", r.Findings, r.Warnings)
	}

	// A settings.local.json turning on A's plugin (a local-scope install's): a finding from Check itself, offline,
	// named once even when Claude Code reports the same plugin.
	writeFile(t, root, LocalSettingsFile, `{"enabledPlugins": {"demo-pack@`+marketA+`": true, "mine@my-market": true}}`)
	r = check(nil)
	if codes(r.Findings) != "plugin" || !strings.Contains(r.Findings[0].Message, "turns on demo-pack@"+marketA) {
		t.Errorf("a stale local line: %+v", r.Findings)
	}
	if r := check(&fakeCLI{list: []InstalledPlugin{atA}}); codes(r.Findings) != "plugin" {
		t.Errorf("a stale local line, and Claude Code reporting it: %+v", r.Findings)
	}
	// The line taken out: no drift.
	writeFile(t, root, LocalSettingsFile, `{"enabledPlugins": {"mine@my-market": true}}`)
	if r := check(&fakeCLI{list: []InstalledPlugin{atB}}); codes(r.Findings) != "" || codes(r.Warnings) != "" {
		t.Errorf("the line taken out: %+v %+v", r.Findings, r.Warnings)
	}
}

// Claude Code rewrites the shared settings file in its own key order when it writes it (a project-scope plugin
// install moves the marketplace's owner after its plugins, and the top-level keys around): not an edit of Bonsai's
// lines.
func TestSettingsRewrittenByClaudeCode(t *testing.T) {
	e := setup(t)
	root := testpack.Project(t, e.tmp, "reordered")
	e.link(t, root, e.pack.A)
	doc := settingsDocOf(t, root)
	var reordered schema.Object
	for i := len(doc) - 1; i >= 0; i-- {
		reordered = append(reordered, schema.Member{Key: doc[i].Key, Value: sortedKeys(doc[i].Value)})
	}
	b, err := schema.Encode(reordered)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) == read(t, root, SettingsFile) {
		t.Fatal("the rewrite changed nothing")
	}
	writeFile(t, root, SettingsFile, string(b))
	r, err := Check(root, e.home)
	if err != nil || len(r.Findings) != 0 {
		t.Errorf("check after Claude Code's rewrite: %v %+v", err, r.Findings)
	}
	if p := e.plan(t, root, Request{}); !p.Nothing() {
		t.Errorf("update after Claude Code's rewrite:\n%s", p.Preview(false))
	}
}

// Output captured from Claude Code 2.1.294 (paths shortened).
const (
	listOutput = `[
  {
    "id": "test-pack@bonsai-exp-one-f004a1d0",
    "version": "506205354b75",
    "scope": "local",
    "enabled": true,
    "installPath": "/x/plugin-cache/cache/bonsai-exp-one-f004a1d0/test-pack/506205354b75",
    "installedAt": "2026-10-08T15:01:22.562Z",
    "lastUpdated": "2026-10-08T15:01:22.562Z",
    "projectPath": "/x/p4b-exp/x1",
    "projectEnabled": true
  },
  {
    "id": "warp@claude-code-warp",
    "version": "2.3.0",
    "scope": "user",
    "enabled": true,
    "installPath": "/x/plugin-cache/cache/claude-code-warp/warp/2.3.0",
    "installedAt": "2026-10-08T12:57:18.712Z",
    "lastUpdated": "2026-10-08T12:57:18.712Z",
    "projectEnabled": false
  }
]
`
	installOK       = "{\"command\":\"install\",\"outcome\":\"ok\",\"plugin\":\"test-pack@bonsai-project-guard-f004a1d0\",\"pluginId\":\"test-pack@bonsai-project-guard-f004a1d0\",\"scope\":\"local\",\"message\":\"Successfully installed plugin: test-pack@bonsai-project-guard-f004a1d0 (scope: local)\"}\r\n"
	installNotFound = "{\"command\":\"install\",\"outcome\":\"failed\",\"plugin\":\"test-pack@bonsai-exp-one-f004a1d0\",\"scope\":\"project\",\"message\":\"Plugin \\\"test-pack\\\" not found in marketplace \\\"bonsai-exp-one-f004a1d0\\\". Your local copy may be out of date \u2014 try `claude plugin marketplace update bonsai-exp-one-f004a1d0`.\",\"failureCode\":\"not_found\"}\n"
)

func TestParseClaudeOutput(t *testing.T) {
	list, err := ParsePluginList([]byte(listOutput))
	if err != nil || len(list) != 2 || list[0] != (InstalledPlugin{ID: "test-pack@bonsai-exp-one-f004a1d0", Version: "506205354b75",
		Scope: "local", Enabled: true, ProjectPath: "/x/p4b-exp/x1"}) || list[1].Scope != "user" || list[1].ProjectPath != "" {
		t.Errorf("list: %v %+v", err, list)
	}
	if _, err := ParsePluginList([]byte("Installed plugins:\n")); err == nil {
		t.Error("a text list read as JSON")
	}
	if r, ok := ParseInstallResult([]byte(installOK)); !ok || r.Outcome != "ok" {
		t.Errorf("install ok: %v %+v", ok, r)
	}
	if r, ok := ParseInstallResult([]byte("some text\n" + installNotFound)); !ok || r.Outcome != "failed" || r.FailureCode != "not_found" {
		t.Errorf("install not found: %v %+v", ok, r)
	}
	if _, ok := ParseInstallResult([]byte("{\"no\": \"outcome\"}\n")); ok {
		t.Error("a line with no outcome read as a result")
	}
}

// A plugin installed for a checkout is matched by its folder as this system compares folders.
func TestForCheckout(t *testing.T) {
	root := t.TempDir()
	if r, err := filepath.EvalSymlinks(root); err == nil {
		root = r
	}
	cases := []struct {
		p    InstalledPlugin
		want bool
	}{
		{InstalledPlugin{Scope: "user"}, true},
		{InstalledPlugin{Scope: "local", ProjectPath: root}, true},
		{InstalledPlugin{Scope: "local", ProjectPath: root + string(filepath.Separator)}, true},
		{InstalledPlugin{Scope: "project", ProjectPath: filepath.Join(root, "sub")}, false},
		{InstalledPlugin{Scope: "local", ProjectPath: filepath.Dir(root)}, false},
	}
	if runtime.GOOS == "windows" {
		cases = append(cases,
			struct {
				p    InstalledPlugin
				want bool
			}{InstalledPlugin{Scope: "local", ProjectPath: strings.ToUpper(filepath.ToSlash(root))}, true})
	}
	for _, c := range cases {
		if got := forCheckout(c.p, root); got != c.want {
			t.Errorf("%+v: %v", c.p, got)
		}
	}
}
