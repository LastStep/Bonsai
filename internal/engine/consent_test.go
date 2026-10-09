package engine

// Consent to code (consent.go; plan-5, piece 5.1.1, rules 1-8) at the engine's level: what each plan lists under
// "Runs code" on the test pack's commits A to F and on fixture packs, that Apply writes nothing without AllowExec,
// and a plugin's code parts read from its files. cmd/bonsai/consent_test.go walks the same cases through the flags.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/testpack"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// items gives a plan's Runs code as "kind change item (pack)", one per item, joined by "; ".
func items(p *Plan) string {
	var out []string
	for _, c := range p.RunsCode {
		out = append(out, c.Kind+" "+c.Change+" "+c.Item+" ("+c.Pack+")")
		if c.Why == "" || c.Why[len(c.Why)-1] != '.' {
			out = append(out, "NO SENTENCE")
		}
	}
	return strings.Join(out, "; ")
}

func (e *env) linkTo(t *testing.T, root, name, source, ref string, allow bool) *Plan {
	t.Helper()
	return e.plan(t, root, Request{Command: "init", Init: &InitValues{Name: name, Source: source, Ref: ref}, AllowExec: allow})
}

// Each case: a project brought to its starting state, then one plan; what it lists under Runs code; Apply refuses
// it without AllowExec, writing nothing, and writes it with AllowExec.
func TestConsentCases(t *testing.T) {
	e := setup(t)
	quiet, qc := testpack.QuietPack(t, e.tmp)
	run, rc := testpack.RunPack(t, e.tmp)
	mcp, mc := testpack.MCPPack(t, e.tmp)
	type step struct {
		name  string
		start func(root string) // brings the project to its starting state (linked with --allow-exec, a ref changed)
		req   Request           // the plan's request (AllowExec is set by the test)
		want  string            // items(p)
	}
	move := func(from, to string) func(string) {
		return func(root string) { e.link(t, root, from); testpack.SetRef(t, root, from, to) }
	}
	initReq := Request{Command: "init", Init: &InitValues{}}
	other := func(source string, from, to string) func(string) {
		return func(root string) {
			e.apply(t, root, Request{Command: "init", Init: &InitValues{Name: "other", Source: source, Ref: from}, AllowExec: true})
			testpack.SetRef(t, root, from, to)
		}
	}
	noLock := func(ref, to string) func(string) {
		return func(root string) {
			e.link(t, root, ref)
			if err := os.Remove(filepath.Join(root, filepath.FromSlash(workspace.LockFile))); err != nil {
				t.Fatal(err)
			}
			if to != ref {
				testpack.SetRef(t, root, ref, to)
			}
		}
	}
	firstLink := func(source, ref string) Request {
		return Request{Command: "init", Init: &InitValues{Name: "demo", Source: source, Ref: ref}}
	}
	hookA := "hook add SessionStart (startup): echo demo hook A (demo-pack)"
	cases := []step{
		{"first link to A", nil, firstLink(e.pack.Source, e.pack.A), hookA},
		{"first link to the no-hook fixture", nil, firstLink(quiet, qc[0]), ""},
		{"first link to F", nil, firstLink(e.pack.Source, e.pack.F),
			"hook add SessionStart (startup): echo demo hook E (demo-pack); plugin add hooks/hooks.json (demo-pack)"},
		{"first link to the hook-run fixture", nil, firstLink(run, rc[0]),
			"hook add SessionStart (startup): sh run/hello.sh (run-pack); file add run/hello.sh (run-pack)"},
		{"first link to a plugin with an MCP server", nil, firstLink(mcp, mc[0]), "plugin add .mcp.json (mcp-pack)"},
		{"A to B, no code", move(e.pack.A, e.pack.B), Request{}, ""},
		{"C to D, the hook line alone", move(e.pack.C, e.pack.D), Request{},
			"hook change SessionStart (startup): echo demo hook D (demo-pack)"},
		{"D to E, the mixed update", move(e.pack.D, e.pack.E), Request{},
			"hook change SessionStart (startup): echo demo hook E (demo-pack)"},
		{"E to F, a hook of the plugin itself", move(e.pack.E, e.pack.F), Request{}, "plugin add hooks/hooks.json (demo-pack)"},
		{"F to E, the plugin's hook taken out", move(e.pack.F, e.pack.E), Request{}, ""},
		{"the hook-run file changed alone", other(run, rc[0], rc[1]), Request{}, "file change run/hello.sh (run-pack)"},
		{"a file no hook runs changed", other(run, rc[1], rc[2]), Request{}, ""},
		{"a role of a plugin that carries code", other(mcp, mc[0], mc[1]), Request{},
			"plugin change agents/helper.md (mcp-pack)"},
		{"the plugin's code taken out", other(mcp, mc[1], mc[2]), Request{}, ""},
		{"a relink, the lock deleted, the hook line changed", noLock(e.pack.C, e.pack.D), initReq,
			"hook add SessionStart (startup): echo demo hook D (demo-pack)"},
		{"a relink, the lock deleted, nothing changed", noLock(e.pack.C, e.pack.C), initReq, ""},
		{"a relink, the lock deleted, a plugin with code", noLock(e.pack.F, e.pack.F), initReq, "plugin add hooks/hooks.json (demo-pack)"},
	}
	for i, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			root := testpack.Project(t, e.tmp, "consent-"+string(rune('a'+i)))
			if c.start != nil {
				c.start(root)
			}
			before := snapshot(t, root)
			p := e.plan(t, root, c.req)
			if got := items(p); got != c.want {
				t.Fatalf("runs code\n  %s\nwant\n  %s\n%s", got, c.want, p.Preview(false))
			}
			if c.want == "" {
				if !strings.Contains(p.Preview(false), "Runs code: nothing that needs --allow-exec") {
					t.Errorf("preview:\n%s", p.Preview(false))
				}
				e.apply(t, root, c.req)
				return
			}
			if pv := p.Preview(false); !strings.Contains(pv, "which needs --allow-exec as well as --yes") &&
				!strings.Contains(pv, "which need --allow-exec as well as --yes") {
				t.Errorf("preview:\n%s", pv)
			}
			if err := Apply(p); err == nil {
				t.Fatal("Apply wrote a plan that runs code, without AllowExec")
			}
			sameSnapshot(t, "a refused plan", before, snapshot(t, root))
			req := c.req
			req.AllowExec = true
			p = e.apply(t, root, req)
			if !strings.Contains(p.Preview(false), "consented to with --allow-exec") {
				t.Errorf("preview with --allow-exec:\n%s", p.Preview(false))
			}
			if again := e.plan(t, root, Request{}); !again.Nothing() {
				t.Errorf("after the write, update would write again:\n%s", again.Preview(false))
			}
		})
	}
}

// A link again with the lock missing names the hook lines it leaves as the project's own: here the pack's earlier
// line, which Bonsai cannot tell from a person's without the lock.
func TestRelinkNamesTheHookLinesItLeaves(t *testing.T) {
	e := setup(t)
	root := testpack.Project(t, e.tmp, "relink")
	e.link(t, root, e.pack.C)
	if err := os.Remove(filepath.Join(root, filepath.FromSlash(workspace.LockFile))); err != nil {
		t.Fatal(err)
	}
	p := e.plan(t, root, Request{Command: "init", Init: &InitValues{}})
	if len(p.LeftHooks) != 0 {
		t.Errorf("nothing changed, yet left: %v", p.LeftHooks)
	}
	testpack.SetRef(t, root, e.pack.C, e.pack.D)
	p = e.plan(t, root, Request{Command: "init", Init: &InitValues{}})
	if strings.Join(p.LeftHooks, "; ") != "SessionStart (startup): echo demo hook A" ||
		!strings.Contains(p.Preview(false), "Left in place as the project's own (the lock is missing") ||
		!strings.Contains(p.Preview(false), "  keep    hook        SessionStart (startup): echo demo hook A\n") {
		t.Errorf("left %v:\n%s", p.LeftHooks, p.Preview(false))
	}
	// A first link (no bonsai.yaml) names none: a project's own hooks are its own there.
	fresh := testpack.Project(t, e.tmp, "fresh")
	writeFile(t, fresh, SettingsFile, drifted)
	if p := e.plan(t, fresh, Request{Command: "init", Init: e.values(e.pack.A)}); len(p.LeftHooks) != 0 {
		t.Errorf("a first link left %v", p.LeftHooks)
	}
}

// A first link writes Bonsai's own hook line on --yes alone, naming it apart from Runs code; a pack's needs
// --allow-exec. A changed own line is code even at a first link (rule 7).
func TestOwnHookLines(t *testing.T) {
	e := setup(t)
	quiet, qc := testpack.QuietPack(t, e.tmp)
	root := testpack.Project(t, e.tmp, "own")
	p := e.linkTo(t, root, "quiet", quiet, qc[0], false)
	if len(p.OwnHooks) != 1 || p.OwnHooks[0].Line != GuardEvent+" ("+GuardMatcher+"): "+GuardCommand || p.NeedsExec() {
		t.Fatalf("own hooks %+v, runs code %+v", p.OwnHooks, p.RunsCode)
	}
	if pv := p.Preview(false); !strings.Contains(pv, "Bonsai's own hook line, written with --yes (linking the project is your consent to it):\n"+
		"  add     hook        PreToolUse") {
		t.Errorf("preview:\n%s", pv)
	}
	if err := Apply(p); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(read(t, root, SettingsFile), GuardCommand) {
		t.Error("Bonsai's own line was not written")
	}
	// Bonsai's line edited on disk to an older form, the lock deleted: init again changes a hook line: code.
	s := read(t, root, SettingsFile)
	writeFile(t, root, SettingsFile, strings.Replace(s, `"bonsai hook guard || exit 2"`, `"bonsai hook guard"`, 1))
	if err := os.Remove(filepath.Join(root, filepath.FromSlash(workspace.LockFile))); err != nil {
		t.Fatal(err)
	}
	p = e.plan(t, root, Request{Command: "init", Init: &InitValues{}})
	if got := items(p); got != "hook change "+GuardEvent+" ("+GuardMatcher+"): "+GuardCommand+" (bonsai)" || len(p.OwnHooks) != 0 {
		t.Errorf("a changed own line at a relink: %s, own %+v", got, p.OwnHooks)
	}
}

// A hook-run file in conflict counts as code (--adopt would write it); kept, it is not.
func TestHookRunFileConflict(t *testing.T) {
	e := setup(t)
	run, rc := testpack.RunPack(t, e.tmp)
	root := testpack.Project(t, e.tmp, "conflict")
	e.apply(t, root, Request{Command: "init", Init: &InitValues{Name: "demo", Source: run, Ref: rc[0]}, AllowExec: true})
	writeFile(t, root, "run/hello.sh", "echo my own hello\n")
	testpack.SetRef(t, root, rc[0], rc[1])
	p := e.plan(t, root, Request{})
	if got := items(p); got != "file change run/hello.sh (run-pack)" || len(p.Conflicts) != 1 {
		t.Fatalf("runs code %s, conflicts %d", got, len(p.Conflicts))
	}
	if p := e.plan(t, root, Request{Keep: []string{"run/hello.sh"}}); p.NeedsExec() {
		t.Errorf("a kept hook-run file is code: %s", items(p))
	}
	p = e.apply(t, root, Request{Adopt: []string{"run/hello.sh"}, AllowExec: true})
	if items(p) != "file change run/hello.sh (run-pack)" || read(t, root, "run/hello.sh") != "echo hello 2\n" {
		t.Errorf("--adopt: %s %q", items(p), read(t, root, "run/hello.sh"))
	}
}

// The settings file in conflict: its hook-line change is named before --adopt, so one refusal names both steps.
func TestSettingsConflictNamesTheCode(t *testing.T) {
	e := setup(t)
	root := testpack.Project(t, e.tmp, "settings")
	e.link(t, root, e.pack.C)
	s := read(t, root, SettingsFile)
	writeFile(t, root, SettingsFile, strings.Replace(s, `"Edit(demo/never.txt)"`, `"Edit(demo/never.md)"`, 1))
	testpack.SetRef(t, root, e.pack.C, e.pack.D)
	p := e.plan(t, root, Request{})
	if len(p.Conflicts) != 1 || p.Conflicts[0].Path != SettingsFile || items(p) != "hook change SessionStart (startup): echo demo hook D (demo-pack)" {
		t.Fatalf("conflicts %d, runs code %s\n%s", len(p.Conflicts), items(p), p.Preview(false))
	}
	if err := Apply(p); err == nil {
		t.Fatal("applied")
	}
	e.apply(t, root, Request{Adopt: []string{SettingsFile}, AllowExec: true})
	if !strings.Contains(read(t, root, SettingsFile), "echo demo hook D") {
		t.Errorf("settings:\n%s", read(t, root, SettingsFile))
	}
}

// What a hook command names: a pack file by its path or its file name, as a word of its own, letter case aside, a
// backslash read as a slash.
func TestNamesPath(t *testing.T) {
	for _, cmd := range []string{"sh fxu/run.sh", "cd fxu && sh run.sh", `sh FXU\RUN.SH`, `sh "$CLAUDE_PROJECT_DIR"/fxu/run.sh`,
		"sh ./fxu/run.sh; echo done", "'fxu/run.sh'"} {
		if !namesPath(cmd, "fxu/run.sh") {
			t.Errorf("%q is not seen to name fxu/run.sh", cmd)
		}
	}
	for _, cmd := range []string{"sh fxu/run.sh.bak", "echo fxu/run.shx", "sh other-run.sh", "sh fxu/run.sh/x", "echo hi"} {
		if namesPath(cmd, "fxu/run.sh") {
			t.Errorf("%q is seen to name fxu/run.sh", cmd)
		}
	}
}

// A plugin's code parts, read from its files: each kind the plugin reference lists, letter case aside, a symbolic
// link, a submodule, the manifest's keys, and a manifest Bonsai cannot read.
func TestPluginCode(t *testing.T) {
	blob := func(sha string) treeEntry { return treeEntry{mode: "100644", typ: "blob", sha: sha} }
	cases := []struct {
		name     string
		files    map[string]treeEntry
		manifest string
		want     string
	}{
		{"prompts only", map[string]treeEntry{"agents/a.md": blob("1"), "skills/s/SKILL.md": blob("2"), "commands/c.md": blob("3"),
			"output-styles/o.md": blob("4"), "themes/t.json": blob("5"), "workflows/w.js": blob("6"), "bin/tool": blob("7"),
			"README.md": blob("8"), "bonsai/pack.yaml": blob("9"), "sub/.mcp.json": blob("a"), "sub/settings.json": blob("b")},
			`{"name": "p", "agents": ["./agents/a.md"], "skills": "./skills/", "commands": "./commands/", "outputStyles": "./o/"}`, ""},
		{"hooks", map[string]treeEntry{"hooks/hooks.json": blob("1"), "hooks/register.js": blob("2")}, "",
			"hooks/hooks.json, hooks/register.js"},
		{"letter case", map[string]treeEntry{"Hooks/Hooks.json": blob("1"), ".MCP.json": blob("2"), "Settings.json": blob("3")}, "",
			".MCP.json, Hooks/Hooks.json, Settings.json"},
		{"servers and monitors", map[string]treeEntry{".mcp.json": blob("1"), ".lsp.json": blob("2"), "monitors/monitors.json": blob("3"),
			"settings.json": blob("4")}, "", ".lsp.json, .mcp.json, monitors/monitors.json, settings.json"},
		{"a link and a submodule", map[string]treeEntry{"scripts": {mode: "120000", typ: "blob", sha: "1"},
			"vendor/x": {mode: "160000", typ: "commit", sha: "2"}}, "", "scripts, vendor/x"},
		{"the manifest's keys", nil, `{"name": "p", "hooks": "./h.json", "mcpServers": {}, "lspServers": [], "monitors": null,
			"channels": [], "settings": {"agent": "a"}, "dependencies": ["x"], "experimental": {"monitors": "./m.json", "themes": "./t/"}}`,
			manifestPath + ": channels, " + manifestPath + ": dependencies, " + manifestPath + ": experimental.monitors, " +
				manifestPath + ": hooks, " + manifestPath + ": lspServers, " + manifestPath + ": mcpServers, " + manifestPath + ": settings"},
		{"a manifest Bonsai cannot read", nil, `{"name": "p", "name": "q"}`, manifestPath},
		{"a manifest not an object", nil, `["p"]`, manifestPath},
		{"experimental not an object", nil, `{"name": "p", "experimental": 1}`, manifestPath + ": experimental"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			entries := map[string]treeEntry{}
			blobs := map[string][]byte{}
			for p, e := range c.files {
				entries[p] = e
			}
			if c.manifest != "" {
				entries[manifestPath] = blob("m")
				blobs["m"] = []byte(c.manifest)
			}
			parts := pluginCode(entries, blobs)
			if got := codeParts(parts); got != c.want {
				t.Errorf("code parts\n  %s\nwant\n  %s", got, c.want)
			}
			for _, p := range parts {
				if p.Why == "" || !strings.HasSuffix(p.Why, ".") {
					t.Errorf("%s has no sentence", p.Path)
				}
			}
		})
	}
}
