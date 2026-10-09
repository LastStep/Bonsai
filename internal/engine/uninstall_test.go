package engine

// Claude Code's own records (step 5.1.7): the uninstall of a pack taken out, at project scope, for this checkout only;
// and the test that Bonsai's every `claude plugin install` and `uninstall` passes --scope project and nothing else.

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/workspace"
)

// UninstallPlugins removes every project-scope install of the pack for this checkout from one of the workspace's
// marketplaces (the lock's and an older one), and nothing else: not another checkout's, not a local or user scope
// install, not another pack's, not another marketplace's.
func TestUninstallPlugins(t *testing.T) {
	root := t.TempDir()
	other := t.TempDir()
	market := MarketplaceName("demo", []string{strings.Repeat("a", 40)})
	old := "bonsai-demo-0123abcd"
	lp := workspace.LockedPack{ID: "demo-pack", Commit: strings.Repeat("a", 40)}
	slash := filepath.ToSlash(root)
	list := []InstalledPlugin{
		{ID: "demo-pack@" + market, Version: "aaaaaaaaaaaa", Scope: "project", ProjectPath: slash},
		{ID: "demo-pack@" + old, Version: "bbbbbbbbbbbb", Scope: "project", ProjectPath: root},
		{ID: "demo-pack@" + market, Scope: "project", ProjectPath: other},
		{ID: "demo-pack@" + market, Scope: "local", ProjectPath: root},
		{ID: "demo-pack@" + market, Scope: "user"},
		{ID: "other-pack@" + market, Scope: "project", ProjectPath: root},
		{ID: "demo-pack@someone-else", Scope: "project", ProjectPath: root},
		{ID: "demo-pack@bonsai-another-0123abcd", Scope: "project", ProjectPath: root},
	}
	f := &fakeCLI{list: list}
	got := UninstallPlugins(root, "demo", market, []workspace.LockedPack{lp}, f)
	if strings.Join(f.asked, ",") != "list,uninstall demo-pack@"+old+",uninstall demo-pack@"+market {
		t.Errorf("asked %v (in the order of their ids)", f.asked)
	}
	for _, d := range f.askedDir {
		if d != root {
			t.Errorf("asked in %s, not the checkout", d)
		}
	}
	if len(got) != 2 || got[0].Result != "uninstalled" || got[1].Result != "uninstalled" || got[1].Plugin != "demo-pack@"+market ||
		got[1].Commit != lp.Commit || got[1].Next != "" || !strings.Contains(got[1].Message, "record of the install for this checkout removed") {
		t.Errorf("results %+v", got)
	}

	// The resumed unlink (bonsai.yaml gone): any workspace name's marketplace, for this pack.
	f = &fakeCLI{list: list}
	got = UninstallPlugins(root, "", "", []workspace.LockedPack{lp}, f)
	if strings.Join(f.asked, ",") != "list,uninstall demo-pack@bonsai-another-0123abcd,uninstall demo-pack@"+old+",uninstall demo-pack@"+market {
		t.Errorf("no name: asked %v (%+v)", f.asked, got)
	}

	// Nothing installed here: one result, nothing to remove.
	f = &fakeCLI{list: list[2:5]}
	got = UninstallPlugins(root, "demo", market, []workspace.LockedPack{lp}, f)
	if strings.Join(f.asked, ",") != "list" || len(got) != 1 || got[0].Result != "uninstalled" || got[0].Message != "not installed for this checkout: nothing to remove" ||
		got[0].Plugin != "demo-pack@"+market {
		t.Errorf("none installed: asked %v, %+v", f.asked, got)
	}
	// Claude Code says not installed: no failure.
	f = &fakeCLI{list: list[:1], uninstall: map[string]InstallResult{"demo-pack@" + market: {Outcome: "failed", FailureCode: "not_installed"}}}
	if got := UninstallPlugins(root, "demo", market, []workspace.LockedPack{lp}, f); len(got) != 1 || got[0].Result != "uninstalled" ||
		!strings.Contains(got[0].Message, "nothing to remove") {
		t.Errorf("not_installed: %+v", got)
	}
	// A failure names the exact command, an agent's step; it never stops the rest.
	two := []workspace.LockedPack{lp, {ID: "other-pack", Commit: lp.Commit}}
	f = &fakeCLI{list: list, uninstall: map[string]InstallResult{"demo-pack@" + market: {Outcome: "failed", FailureCode: "boom", Message: "it broke"}}}
	got = UninstallPlugins(root, "demo", market, two, f)
	if len(got) != 3 || got[1].Result != "failed" || got[1].Message != "it broke" || got[1].Who != "agent" ||
		got[1].Next != "run: claude plugin uninstall demo-pack@"+market+" --scope project" ||
		got[2].Pack != "other-pack" || got[2].Result != "uninstalled" {
		t.Errorf("a failure: %+v", got)
	}
	// Claude Code not on the PATH: skipped, a person's step; a list that fails: failed, with the command.
	f = &fakeCLI{listErr: ErrNoClaude}
	if got := UninstallPlugins(root, "demo", market, []workspace.LockedPack{lp}, f); len(got) != 1 || got[0].Result != "skipped" || got[0].Who != "person" {
		t.Errorf("no claude: %+v", got)
	}
	f = &fakeCLI{listErr: errors.New("claude plugin list --json failed: boom")}
	if got := UninstallPlugins(root, "demo", market, []workspace.LockedPack{lp}, f); len(got) != 1 || got[0].Result != "failed" ||
		!strings.HasSuffix(got[0].Next, "claude plugin uninstall demo-pack@"+market+" --scope project") {
		t.Errorf("a list that fails: %+v", got)
	}
	// No packs, or no Claude Code to ask: nothing asked.
	if got := UninstallPlugins(root, "demo", market, nil, &fakeCLI{}); got != nil {
		t.Errorf("no packs: %+v", got)
	}
	if got := UninstallPlugins(root, "demo", market, []workspace.LockedPack{lp}, nil); got != nil {
		t.Errorf("no cli: %+v", got)
	}
}

// fakeClaudeLogging puts a stand-in claude first and alone on the PATH (a shell script, or a .cmd file on Windows: no
// file mode is needed there) that writes each command line it is given to a log and answers as Claude Code would: list
// with one project-scope install for root, marketplace list with none, install and uninstall with ok. It returns the
// log's path. No test reaches the real claude: the PATH holds only the stand-in's folder.
func fakeClaudeLogging(t *testing.T, root string, plugins ...string) string {
	t.Helper()
	dir := t.TempDir()
	log := filepath.Join(dir, "claude.log")
	var items []string
	for _, plugin := range plugins {
		items = append(items, `{"id":"`+plugin+`","version":"aaaaaaaaaaaa","scope":"project","enabled":true,"projectPath":"`+filepath.ToSlash(root)+`"}`)
	}
	list := "[" + strings.Join(items, ",") + "]"
	if runtime.GOOS == "windows" {
		body := "@echo off\r\necho %*>>\"" + log + "\"\r\nif \"%2\"==\"list\" goto list\r\nif \"%2\"==\"marketplace\" goto market\r\n" +
			"echo {\"outcome\":\"ok\"}\r\nexit /b 0\r\n:list\r\necho " + list + "\r\nexit /b 0\r\n:market\r\necho []\r\nexit /b 0\r\n"
		if err := os.WriteFile(filepath.Join(dir, "claude.cmd"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	} else {
		body := "#!/bin/sh\nprintf '%s\\n' \"$*\" >> '" + log + "'\ncase \"$2\" in\n  list) echo '" + list + "';;\n" +
			"  marketplace) echo '[]';;\n  *) echo '{\"outcome\":\"ok\"}';;\nesac\n"
		if err := os.WriteFile(filepath.Join(dir, "claude"), []byte(body), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", dir)
	return log
}

// Project scope only (plan-5, 5.1.7; gate report section 5: local scope leaks across worktrees, user scope writes the
// person's own settings): every `claude plugin install` and `uninstall` Bonsai runs, through the real command line
// code (ClaudeCLI) and the install and uninstall steps, passes --scope project once, and no other scope; and Bonsai
// runs no claude command but plugin list, marketplace list, install and uninstall.
func TestPluginScopeOnly(t *testing.T) {
	root := t.TempDir()
	cfg := &workspace.Config{Name: "demo"}
	lp := workspace.LockedPack{ID: "demo-pack", Commit: strings.Repeat("a", 40)}
	lock := &workspace.Lock{Packs: []workspace.LockedPack{lp}}
	market := lockMarket(cfg, lock)
	log := fakeClaudeLogging(t, root, "demo-pack@"+market, "demo-pack@bonsai-demo-0123abcd")

	cli := ClaudeCLI{}
	if _, err := cli.Install(root, "demo-pack@"+market); err != nil {
		t.Fatal(err)
	}
	if _, err := cli.Uninstall(root, "demo-pack@"+market); err != nil {
		t.Fatal(err)
	}
	if _, err := cli.List(root); err != nil {
		t.Fatal(err)
	}
	if names, err := cli.Marketplaces(root); err != nil || len(names) != 0 {
		t.Fatal(names, err)
	}
	InstallPlugins(root, cfg, lock, cli, PluginConsent{AllowExec: true})
	if got := PruneStalePlugins(root, cfg, lock, cli); len(got) != 1 || got[0].Plugin != "demo-pack@bonsai-demo-0123abcd" || got[0].Result != "uninstalled" {
		t.Errorf("the stale record through the stand-in: %+v", got)
	}
	if got := UninstallPlugins(root, "demo", market, lock.Packs, cli); len(got) != 2 || got[0].Result != "uninstalled" || got[1].Result != "uninstalled" {
		t.Errorf("uninstall through the stand-in: %+v", got)
	}
	r := &CheckResult{Root: root, Config: cfg, Lock: lock}
	ComparePlugins(r, cli)

	raw, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(strings.ReplaceAll(string(raw), "\r\n", "\n")), "\n")
	writes := 0
	for _, l := range lines {
		args := strings.Fields(l)
		if len(args) < 2 || args[0] != "plugin" {
			t.Errorf("Bonsai ran claude %s: not a plugin command", l)
			continue
		}
		switch args[1] {
		case "list":
		case "marketplace":
			if len(args) < 3 || args[2] != "list" {
				t.Errorf("Bonsai ran claude %s", l)
			}
		case "install", "uninstall":
			writes++
			scopes := 0
			for i, a := range args {
				switch {
				case a == "--scope" && i+1 < len(args) && args[i+1] == "project":
					scopes++
				case a == "--scope", a == "-s", strings.HasPrefix(a, "--scope="), strings.HasPrefix(a, "-s="):
					t.Errorf("claude %s passes another scope", l)
				}
			}
			if scopes != 1 {
				t.Errorf("claude %s does not pass --scope project once", l)
			}
		default:
			t.Errorf("Bonsai ran claude %s", l)
		}
	}
	if writes != 6 {
		t.Errorf("%d installs and uninstalls logged, want 6:\n%s", writes, raw)
	}
}

// After update writes, the stale records go (PruneStalePlugins): this checkout's project-scope install of a locked
// pack's plugin under an older marketplace name of this workspace. Kept: the lock's own marketplace's record, another
// workspace's (another name), another checkout's, a local or user scope install, another pack's.
func TestPruneStalePlugins(t *testing.T) {
	root := t.TempDir()
	cfg := &workspace.Config{Name: "demo"}
	lp := workspace.LockedPack{ID: "demo-pack", Commit: strings.Repeat("b", 40)}
	lock := &workspace.Lock{Packs: []workspace.LockedPack{lp}}
	market := lockMarket(cfg, lock)
	list := []InstalledPlugin{
		{ID: "demo-pack@" + market, Version: "bbbbbbbbbbbb", Scope: "project", ProjectPath: root},
		{ID: "demo-pack@bonsai-demo-0123abcd", Version: "aaaaaaaaaaaa", Scope: "project", ProjectPath: root},
		{ID: "demo-pack@bonsai-demo-4567cdef", Version: "cccccccccccc", Scope: "project", ProjectPath: filepath.ToSlash(root)},
		{ID: "demo-pack@bonsai-demo-two-0123abcd", Scope: "project", ProjectPath: root},
		{ID: "demo-pack@bonsai-other-0123abcd", Scope: "project", ProjectPath: root},
		{ID: "demo-pack@bonsai-demo-0123abcd", Scope: "project", ProjectPath: t.TempDir()},
		{ID: "demo-pack@bonsai-demo-0123abcd", Scope: "local", ProjectPath: root},
		{ID: "demo-pack@bonsai-demo-0123abcd", Scope: "user"},
		{ID: "side-pack@bonsai-demo-0123abcd", Scope: "project", ProjectPath: root},
		{ID: "demo-pack@someone-else", Scope: "project", ProjectPath: root},
	}
	f := &fakeCLI{list: list}
	got := PruneStalePlugins(root, cfg, lock, f)
	if strings.Join(f.asked, ",") != "list,uninstall demo-pack@bonsai-demo-0123abcd,uninstall demo-pack@bonsai-demo-4567cdef" {
		t.Errorf("asked %v", f.asked)
	}
	if len(got) != 2 || got[0].Result != "uninstalled" || got[0].Commit != lp.Commit || got[0].Next != "" ||
		got[0].Message != "a stale record (at aaaaaaaaaaaa, from an older marketplace name of this workspace) removed for this checkout; the lock's plugin is demo-pack@"+market {
		t.Errorf("results %+v", got)
	}
	// Only the current record: nothing asked but the list, nothing reported.
	f = &fakeCLI{list: list[:1]}
	if got := PruneStalePlugins(root, cfg, lock, f); got != nil || strings.Join(f.asked, ",") != "list" {
		t.Errorf("the current record only: %v %+v", f.asked, got)
	}
	// A failure: failed, its exact run: line, an agent's step.
	f = &fakeCLI{list: list[:2], uninstall: map[string]InstallResult{"demo-pack@bonsai-demo-0123abcd": {Outcome: "failed", FailureCode: "boom", Message: "it broke"}}}
	if got := PruneStalePlugins(root, cfg, lock, f); len(got) != 1 || got[0].Result != "failed" || got[0].Who != "agent" ||
		got[0].Next != "run: claude plugin uninstall demo-pack@bonsai-demo-0123abcd --scope project" {
		t.Errorf("a failure: %+v", got)
	}
	// Claude Code cannot be asked, no cli, no lock: nothing.
	for _, cli := range []PluginCLI{&fakeCLI{listErr: ErrNoClaude}, &fakeCLI{listErr: errors.New("boom")}, nil} {
		if got := PruneStalePlugins(root, cfg, lock, cli); got != nil {
			t.Errorf("%v: %+v", cli, got)
		}
	}
	if got := PruneStalePlugins(root, cfg, &workspace.Lock{}, &fakeCLI{list: list}); got != nil {
		t.Errorf("no packs: %+v", got)
	}
}
