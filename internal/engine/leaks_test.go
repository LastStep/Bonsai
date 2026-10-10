package engine

// The leaks step 5.1.1's verifier found (B1, S1, B2, B3, S3), one test each: code that a plan wrote, or would have
// written, on --yes alone. Each failed on the build the verifier read and passes on the fix.

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/testpack"
	"github.com/LastStep/Bonsai/internal/workspace"
)

func pluginJSON(id string) string {
	return "{\"name\": \"" + id + "\", \"description\": \"A fixture pack.\"}\n"
}

func packYAML(id, files, hooks string) string {
	return "format: bonsai.pack/1\nid: " + id + "\nversion: \"0.1.0\"\nfiles:" + files + "\nhooks:" + hooks + "\ndeny: []\n"
}

func hookYAML(command, runs string) string {
	s := "\n  - event: SessionStart\n    matcher: startup\n    command: \"" + command + "\"\n"
	if runs != "" {
		s += "    runs: " + runs + "\n"
	}
	return s + "    why: \"A fixture hook line.\""
}

const pluginHooksJSON = "{\"hooks\": {\"SessionStart\": [{\"hooks\": [{\"type\": \"command\", \"command\": \"echo plugin hook\"}]}]}}\n"

// refusedWith checks a plan was refused at input (exit 2) with words in its message.
func refusedWith(t *testing.T, err error, words string) {
	t.Helper()
	var ee *Error
	if !errors.As(err, &ee) || ee.Exit != ExitInput || !strings.Contains(ee.What, words) || ee.Next == "" {
		t.Fatalf("want exit 2 holding %q and a next step, got %v", words, err)
	}
}

// B1: a pack's folder changed in bonsai.yaml (the lock records no folder). The new folder's hook line and plugin
// hook are code, with the commit the same and with it changed too.
func TestFolderChangeIsCode(t *testing.T) {
	e := setup(t)
	src, shas := testpack.Fixture(t, e.tmp, "twin",
		map[string]string{
			"a/.claude-plugin/plugin.json": pluginJSON("twin"),
			"a/bonsai/pack.yaml":           packYAML("twin", " []", " []"),
			"b/.claude-plugin/plugin.json": pluginJSON("twin"),
			"b/hooks/hooks.json":           pluginHooksJSON,
			"b/bonsai/pack.yaml":           packYAML("twin", " []", hookYAML("echo twin b pack hook", "")),
		},
		map[string]string{"b/agents/x.md": "x\n"})
	want := "hook add SessionStart (startup): echo twin b pack hook (twin); plugin add hooks/hooks.json (twin)"
	for i, to := range []string{shas[0], shas[1]} {
		root := testpack.Project(t, e.tmp, "twin-"+string(rune('a'+i)))
		e.apply(t, root, Request{Command: "init", Init: &InitValues{Name: "twin", Source: src, Ref: shas[0], Path: "a"}})
		cfg := read(t, root, "bonsai.yaml")
		writeFile(t, root, "bonsai.yaml", strings.Replace(strings.Replace(cfg, `path: "a"`, `path: "b"`, 1), shas[0], to, 1))
		before := snapshot(t, root)
		p := e.plan(t, root, Request{})
		if got := items(p); got != want {
			t.Fatalf("to %s: runs code\n  %s\nwant\n  %s\n%s", to[:7], got, want, p.Preview(false))
		}
		if err := Apply(p); err == nil {
			t.Fatal("applied without AllowExec")
		}
		sameSnapshot(t, "a refused folder change", before, snapshot(t, root))
	}
}

// S1: a lock edited by hand (its commit moved, bonsai.yaml with it) is no record of consent: D's hook line, and F's
// plugin hook, are code.
func TestHandEditedLockIsNoConsent(t *testing.T) {
	e := setup(t)
	for _, c := range []struct{ from, to, want string }{
		{e.pack.C, e.pack.D, "SessionStart (startup): echo demo hook D (demo-pack)"},
		{e.pack.E, e.pack.F, "plugin add hooks/hooks.json (demo-pack)"},
	} {
		t.Run(c.from[:7]+" to "+c.to[:7], func(t *testing.T) {
			root := testpack.Project(t, e.tmp, "hand-"+c.to[:7])
			e.link(t, root, c.from)
			lock := read(t, root, workspace.LockFile)
			writeFile(t, root, workspace.LockFile, strings.Replace(lock, c.from, c.to, 1))
			testpack.SetRef(t, root, c.from, c.to)
			before := snapshot(t, root)
			reqs := []Request{{}}
			if p := e.plan(t, root, Request{}); len(p.Conflicts) > 0 {
				reqs = append(reqs, Request{Adopt: []string{p.Conflicts[0].Path}})
			}
			for _, req := range reqs {
				p := e.plan(t, root, req)
				if got := items(p); !strings.Contains(got, c.want) || !p.NeedsExec() {
					t.Fatalf("%s to %s, adopt %v: runs code %q, want it to hold %q\n%s", c.from[:7], c.to[:7], req.Adopt, got, c.want, p.Preview(false))
				}
				if err := Apply(p); err == nil {
					t.Fatal("applied without AllowExec")
				}
			}
			sameSnapshot(t, "a hand-edited lock", before, snapshot(t, root))
		})
	}
}

// B2: a pack may not take the id bonsai, and a pack's hook line is never Bonsai's own, whatever its pack's id.
func TestPackNamedBonsai(t *testing.T) {
	e := setup(t)
	src, shas := testpack.Fixture(t, e.tmp, "named-bonsai", map[string]string{
		".claude-plugin/plugin.json": pluginJSON("bonsai"),
		"bonsai/pack.yaml":           packYAML("bonsai", " []", hookYAML("echo a pack named bonsai runs this", "")),
	})
	root := testpack.Project(t, e.tmp, "named")
	_, err := e.try(root, Request{Command: "init", Init: &InitValues{Name: "demo", Source: src, Ref: shas[0]}})
	refusedWith(t, err, "is Bonsai's own")
}

// B2: past the reader (as a pack read some other way would be), a pack whose id is bonsai still has its hook line
// counted as the pack's code at a first link: Bonsai's own lines are marked by what no pack sets.
func TestPackLineIsNeverOwn(t *testing.T) {
	src, sha := "https://example.invalid/pack", "0123456789012345678901234567890123456789"
	cfg := &workspace.Config{Name: "demo"}
	lnew := buildLines(cfg, []packLines{{id: "bonsai", source: src, commit: sha, known: true,
		hooks: []workspace.HookEntry{{Event: "SessionStart", Matcher: "startup", Command: "echo a pack named bonsai runs this", Why: "x"}}}})
	p := &Plan{FirstLink: true}
	p.consent(lineChanges(nil, lnew, nil, nil), nil, nil)
	if got := items(p); got != "hook add SessionStart (startup): echo a pack named bonsai runs this (bonsai)" || len(p.OwnHooks) != len(ownHooks) {
		t.Errorf("runs code %q, own hooks %+v", got, p.OwnHooks)
	}
}

// B3: a pack file aimed where Claude Code loads code on its own (a skills-directory plugin under .claude/, the
// project's .mcp.json) is refused at read, at a first link and at an update.
func TestPackFilesClaudeCodeLoads(t *testing.T) {
	e := setup(t)
	for _, target := range []string{".claude/skills/pc/hooks/hooks.json", ".claude/skills/pc/.claude-plugin/plugin.json", ".mcp.json"} {
		src, shas := testpack.Fixture(t, e.tmp, "loads-"+strings.NewReplacer("/", "-", ".", "").Replace(target),
			map[string]string{
				".claude-plugin/plugin.json": pluginJSON("loads"),
				"bonsai/pack.yaml":           packYAML("loads", " []", " []"),
			},
			map[string]string{
				"bonsai/pack.yaml":    packYAML("loads", "\n  - path: \""+target+"\"\n    from: x.json\n    kind: pack", " []"),
				"bonsai/files/x.json": pluginHooksJSON,
			})
		first := testpack.Project(t, e.tmp, "first-"+strings.NewReplacer("/", "-", ".", "").Replace(target))
		_, err := e.try(first, Request{Command: "init", Init: &InitValues{Name: "demo", Source: src, Ref: shas[1]}})
		refusedWith(t, err, "Claude Code")
		root := testpack.Project(t, e.tmp, "update-"+strings.NewReplacer("/", "-", ".", "").Replace(target))
		e.apply(t, root, Request{Command: "init", Init: &InitValues{Name: "demo", Source: src, Ref: shas[0]}})
		testpack.SetRef(t, root, shas[0], shas[1])
		before := snapshot(t, root)
		_, err = e.try(root, Request{})
		refusedWith(t, err, "Claude Code")
		sameSnapshot(t, "a refused update", before, snapshot(t, root))
	}
}

// S3: runs is checked, not trusted. A hook command naming a pack file its runs does not list is refused; a file
// another linked pack writes, run by a hook that lists it, is that hook's code when it changes; a runs item no linked
// pack writes is refused.
func TestRunsIsChecked(t *testing.T) {
	e := setup(t)
	undeclared, us := testpack.Fixture(t, e.tmp, "undeclared", map[string]string{
		".claude-plugin/plugin.json": pluginJSON("undeclared"),
		"bonsai/pack.yaml": packYAML("undeclared", "\n  - path: fxu/run.sh\n    from: run.sh\n    kind: pack",
			hookYAML("sh fxu/run.sh", "")),
		"bonsai/files/run.sh": "echo run 1\n",
	})
	root := testpack.Project(t, e.tmp, "undeclared")
	_, err := e.try(root, Request{Command: "init", Init: &InitValues{Name: "demo", Source: undeclared, Ref: us[0]}, AllowExec: true})
	refusedWith(t, err, "names fxu/run.sh, a file the pack undeclared writes, but its runs does not list it")

	// Two packs: writer writes fxr/notes.md; runner's hook runs it and lists it.
	writer, ws := testpack.Fixture(t, e.tmp, "writer",
		map[string]string{
			".claude-plugin/plugin.json": pluginJSON("writer"),
			"bonsai/pack.yaml":           packYAML("writer", "\n  - path: fxr/notes.md\n    from: notes.md\n    kind: pack", " []"),
			"bonsai/files/notes.md":      "echo notes 1\n",
		},
		map[string]string{"bonsai/files/notes.md": "echo notes 2\n"})
	runner, rs := testpack.Fixture(t, e.tmp, "runner",
		map[string]string{
			".claude-plugin/plugin.json": pluginJSON("runner"),
			"bonsai/pack.yaml":           packYAML("runner", " []", hookYAML("sh fxr/notes.md", `["fxr/notes.md"]`)),
		},
		map[string]string{"bonsai/pack.yaml": packYAML("runner", " []", hookYAML("sh fxr/notes.md", ""))},
		map[string]string{"bonsai/pack.yaml": packYAML("runner", " []", hookYAML("echo hi", `["fxr/other.md"]`))})
	two := func(name, runnerRef string) string {
		root := testpack.Project(t, e.tmp, name)
		writeFile(t, root, "bonsai.yaml", "format: bonsai.workspace/1\nid: ws-aaaaaaaaaaaaaaaaaaaaaaaaaa\nname: demo\npacks:\n"+
			"  - id: writer\n    source: \""+filepath.ToSlash(writer)+"\"\n    ref: \""+ws[0]+"\"\n"+
			"  - id: runner\n    source: \""+filepath.ToSlash(runner)+"\"\n    ref: \""+runnerRef+"\"\n")
		return root
	}
	root = two("two", rs[0])
	e.apply(t, root, Request{Command: "init", Init: &InitValues{}, AllowExec: true})
	testpack.SetRef(t, root, ws[0], ws[1])
	p := e.plan(t, root, Request{})
	if got := items(p); got != "file change fxr/notes.md (runner)" || !strings.Contains(p.Preview(false), "(the pack writer writes it)") {
		t.Errorf("runs code %q\n%s", got, p.Preview(false))
	}
	_, err = e.try(two("two-undeclared", rs[1]), Request{Command: "init", Init: &InitValues{}, AllowExec: true})
	refusedWith(t, err, "names fxr/notes.md, a file the pack writer writes, but its runs does not list it")
	_, err = e.try(two("two-unknown", rs[2]), Request{Command: "init", Init: &InitValues{}, AllowExec: true})
	refusedWith(t, err, "runs names fxr/other.md, which no linked pack writes")
	if _, err := os.Stat(filepath.Join(root, "fxr", "notes.md")); err != nil {
		t.Fatal(err)
	}
}
