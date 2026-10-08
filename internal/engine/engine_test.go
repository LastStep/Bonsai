package engine

// The engine's tests: spec §14's checks 1-6 and 12 as Go tests, on projects and pack repositories in temporary
// folders (internal/testpack builds the pack: commits A to D, as the public test pack has them). No network.

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/testpack"
	"github.com/LastStep/Bonsai/internal/workspace"
)

type env struct {
	tmp, home string
	pack      *testpack.Pack
}

func setup(t *testing.T) *env {
	t.Helper()
	tmp := t.TempDir()
	if r, err := filepath.EvalSymlinks(tmp); err == nil {
		tmp = r
	}
	home := testpack.Isolate(t, tmp)
	return &env{tmp: tmp, home: home, pack: testpack.Build(t, tmp)}
}

func (e *env) values(ref string, neverEdit ...string) *InitValues {
	return &InitValues{Name: "demo", Source: e.pack.Source, Ref: ref, NeverEdit: neverEdit}
}

// plan builds a plan, failing the test on an error.
func (e *env) plan(t *testing.T, root string, req Request) *Plan {
	t.Helper()
	p, err := e.try(root, req)
	if err != nil {
		t.Fatalf("plan: %v", err)
	}
	return p
}

func (e *env) try(root string, req Request) (*Plan, error) {
	req.Dir, req.Home, req.Version = root, e.home, "test"
	if req.Command == "" {
		req.Command = "update"
	}
	return Build(req)
}

// apply builds a plan and writes it, failing the test on anything that stops it.
func (e *env) apply(t *testing.T, root string, req Request) *Plan {
	t.Helper()
	p := e.plan(t, root, req)
	if len(p.Conflicts) > 0 || p.HookChange {
		t.Fatalf("the plan has %d conflicts, hook change %v:\n%s", len(p.Conflicts), p.HookChange, p.Preview(false))
	}
	if err := Apply(p); err != nil {
		t.Fatalf("apply: %v", err)
	}
	return p
}

func (e *env) link(t *testing.T, root, ref string, neverEdit ...string) *Plan {
	t.Helper()
	return e.apply(t, root, Request{Command: "init", Init: e.values(ref, neverEdit...)})
}

// snapshot is every file of a checkout but .git, with its bytes' hash and its modification time.
func snapshot(t *testing.T, root string) map[string]string {
	t.Helper()
	out := map[string]string{}
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && d.Name() == ".git" {
			return filepath.SkipDir
		}
		if d.IsDir() {
			return nil
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		info, _ := d.Info()
		sum := sha256.Sum256(b)
		rel, _ := filepath.Rel(root, p)
		out[filepath.ToSlash(rel)] = hex.EncodeToString(sum[:]) + " " + info.ModTime().Format(time.RFC3339Nano)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func sameSnapshot(t *testing.T, what string, a, b map[string]string) {
	t.Helper()
	keys := map[string]bool{}
	for k := range a {
		keys[k] = true
	}
	for k := range b {
		keys[k] = true
	}
	var diff []string
	for k := range keys {
		if a[k] != b[k] {
			diff = append(diff, k)
		}
	}
	sort.Strings(diff)
	if len(diff) > 0 {
		t.Errorf("%s changed %v", what, diff)
	}
}

func read(t *testing.T, root, rel string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func writeFile(t *testing.T, root, rel, content string) {
	t.Helper()
	p := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func result(t *testing.T, p *Plan, path string) *FileResult {
	t.Helper()
	for _, f := range p.Files {
		if f.Path == path {
			return f
		}
	}
	t.Fatalf("the plan has no %s", path)
	return nil
}

func settingsDocOf(t *testing.T, root string) schema.Object {
	t.Helper()
	v, err := schema.Decode([]byte(read(t, root, SettingsFile)))
	if err != nil {
		t.Fatal(err)
	}
	return v.(schema.Object)
}

// drifted is a project's settings shaped like a drifted one: one old absolute hook line and one hook of its own.
const drifted = `{
  "permissions": {"allow": ["Bash(npm test)"], "deny": ["Read(secrets/**)"]},
  "hooks": {
    "PreToolUse": [
      {"matcher": "Edit|Write", "hooks": [{"type": "command", "command": "/opt/bonsai-0.4/bin/bonsai hook guard", "timeout": 5}]},
      {"matcher": "Bash", "hooks": [{"type": "command", "command": "npm run lint-staged"}]}
    ]
  },
  "model": "opus"
}
`

// Check 1: init into a drifted project. Without the required values it exits 2 naming them; with them it writes
// bonsai.yaml (a comment on every line), the rest, and the lock; the project's own hook stays, the old absolute line
// is gone, Bonsai's lines are in shell form, the deny rules carry never_edit, the plugin wiring holds a
// 40-character commit; the closing words say where everything lives.
func TestCheck1InitIntoADriftedProject(t *testing.T) {
	e := setup(t)
	root := testpack.Project(t, e.tmp, "drifted")
	writeFile(t, root, SettingsFile, drifted)
	writeFile(t, root, "CLAUDE.md", "# The project's own instructions\n\nKeep these.\n")
	testpack.Git(t, root, "add", "-A")
	testpack.Git(t, root, "commit", "-q", "-m", "before")

	_, err := e.try(root, Request{Command: "init", Init: &InitValues{Source: e.pack.Source}})
	var ee *Error
	if !errors.As(err, &ee) || ee.Exit != ExitInput || !strings.Contains(ee.What, "init needs --name and --ref") || ee.Next == "" {
		t.Fatalf("init without values: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "bonsai.yaml")); !os.IsNotExist(err) {
		t.Fatal("a refused init wrote bonsai.yaml")
	}

	p := e.link(t, root, e.pack.A, "ledger.json")
	if !p.FirstLink || p.Config.Name != "demo" || !regexp.MustCompile(`^ws-[a-z2-7]{26}$`).MatchString(p.Config.ID) {
		t.Errorf("first link %v, name %q, id %q", p.FirstLink, p.Config.Name, p.Config.ID)
	}
	// bonsai.yaml: a comment on every line (at its end, or the line above).
	lines := strings.Split(strings.TrimSuffix(read(t, root, "bonsai.yaml"), "\n"), "\n")
	for i, l := range lines {
		if !strings.Contains(l, "#") && (i == 0 || !strings.HasPrefix(lines[i-1], "#")) {
			t.Errorf("bonsai.yaml line %d has no comment: %q", i+1, l)
		}
	}
	cfg, err := workspace.LoadConfig(root)
	if err != nil || cfg.Packs[0].Ref != e.pack.A || cfg.Packs[0].ID != testpack.ID || strings.Join(cfg.NeverEdit, ",") != "ledger.json" {
		t.Fatalf("bonsai.yaml reads back as %+v, %v", cfg, err)
	}
	lock, err := workspace.LoadLock(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(lock.Packs) != 1 || lock.Packs[0].Commit != e.pack.A || lock.Packs[0].Version != "0.1.0" {
		t.Errorf("lock packs %+v", lock.Packs)
	}
	var kinds []string
	for _, path := range sortedPaths(lock.Files) {
		kinds = append(kinds, path+"="+lock.Files[path].Kind)
	}
	if got := strings.Join(kinds, " "); got != ".claude/settings.json=keys CLAUDE.md=block demo/guide.md=pack demo/start.md=once" {
		t.Errorf("lock files %s", got)
	}
	if read(t, root, GitignoreFile) != GitignoreText {
		t.Errorf(".bonsai/.gitignore %q", read(t, root, GitignoreFile))
	}
	if got := read(t, root, "demo/guide.md"); got != "# Guide\n\nEdition 1.\n" {
		t.Errorf("guide.md %q", got)
	}

	// The settings: the project's own entries kept, the old absolute line gone, Bonsai's lines in.
	s := settingsDocOf(t, root)
	text := schema.Show(s)
	for _, want := range []string{`"allow":["Bash(npm test)"]`, `"Read(secrets/**)"`, `"npm run lint-staged"`, `"model":"opus"`,
		`"command":"bonsai hook guard || exit 2"`, `"Edit(ledger.json)"`, `"Edit(demo/never.txt)"`, `"echo demo hook A"`,
		`"autoMemoryEnabled":false`, `"disableAllHooks":false`} {
		if !strings.Contains(text, want) {
			t.Errorf("settings lack %s:\n%s", want, text)
		}
	}
	if strings.Contains(text, "/opt/bonsai-0.4") {
		t.Errorf("the old absolute hook line is still there:\n%s", text)
	}
	// The plugin wiring: a marketplace named for the workspace and its commits, the plugin pinned by sha.
	market := MarketplaceName("demo", []string{e.pack.A})
	if !regexp.MustCompile(`^bonsai-demo-[0-9a-f]{8}$`).MatchString(market) ||
		!strings.Contains(text, `"enabledPlugins":{"demo-pack@`+market+`":true}`) ||
		!strings.Contains(text, `"sha":"`+e.pack.A+`"`) || strings.Contains(text, `"version"`) {
		t.Errorf("plugin wiring:\n%s", text)
	}
	// The order of the project's keys is kept; Bonsai's come after.
	if got := strings.Join(s.Keys(), " "); got != "permissions hooks model autoMemoryEnabled disableAllHooks extraKnownMarketplaces enabledPlugins" {
		t.Errorf("keys %s", got)
	}
	// CLAUDE.md: the project's text, then the block.
	claude := read(t, root, "CLAUDE.md")
	if !strings.HasPrefix(claude, "# The project's own instructions\n\nKeep these.\n\n<!-- bonsai:block start") ||
		!strings.Contains(claude, "The demo pack is linked: its role is demo-pack:marker.\n<!-- bonsai:block end -->\n") ||
		strings.Contains(claude, "This comment stays in the pack") {
		t.Errorf("CLAUDE.md:\n%s", claude)
	}
	// The preview named every settings line, with its sentence, and the hook lines under runs code.
	var named []string
	for _, c := range p.Settings {
		if c.Why == "" {
			t.Errorf("%s %s has no sentence", c.Change, c.Line)
		}
		named = append(named, c.Change+" "+c.Kind)
	}
	if got := strings.Join(named, ", "); got != "add key, add key, add deny, add deny, add hook, add hook, add marketplace, add plugin, remove hook" {
		t.Errorf("settings lines %s", got)
	}
	if !strings.Contains(p.Preview(false), "remove  hook        PreToolUse (Edit|Write): /opt/bonsai-0.4/bin/bonsai hook guard") {
		t.Errorf("preview:\n%s", p.Preview(false))
	}
	words := ClosingWords(p.Config, e.home)
	if !strings.HasPrefix(words, "Linked demo. Workspace id: "+p.Config.ID) || !strings.Contains(words, filepath.ToSlash(e.home)) ||
		!strings.HasSuffix(words, "A copy meant as a new project needs its own id: bonsai init --new-id\n") {
		t.Errorf("closing words:\n%s", words)
	}
}

// Check 2: nothing written holds an absolute path (the project's or the home's); the lock's source is the
// pack's source as bonsai.yaml names it.
func TestCheck2NoAbsolutePath(t *testing.T) {
	e := setup(t)
	root := testpack.Project(t, e.tmp, "fresh")
	e.link(t, root, e.pack.A)
	for _, rel := range []string{"bonsai.yaml", workspace.LockFile, SettingsFile, "CLAUDE.md", GitignoreFile, "demo/guide.md", "demo/start.md"} {
		s := read(t, root, rel)
		for _, abs := range []string{root, filepath.ToSlash(root), e.home, filepath.ToSlash(e.home)} {
			if strings.Contains(s, abs) {
				t.Errorf("%s holds %s", rel, abs)
			}
		}
	}
	lock, _ := workspace.LoadLock(root)
	if lock.Packs[0].Source != e.pack.Source {
		t.Errorf("lock source %s", lock.Packs[0].Source)
	}
}

// Check 3 (its Go half): the lock's keys use forward slashes, and the lock and every file verify after the
// checkout's line endings become CRLF.
func TestCheck3CRLF(t *testing.T) {
	e := setup(t)
	root := testpack.Project(t, e.tmp, "crlf")
	writeFile(t, root, "CLAUDE.md", "# Own\n")
	e.link(t, root, e.pack.A)
	lock, _ := workspace.LoadLock(root)
	for path := range lock.Files {
		if strings.Contains(path, `\`) {
			t.Errorf("lock key %q", path)
		}
	}
	for _, rel := range []string{"bonsai.yaml", workspace.LockFile, SettingsFile, "CLAUDE.md", GitignoreFile, "demo/guide.md", "demo/start.md"} {
		writeFile(t, root, rel, strings.ReplaceAll(read(t, root, rel), "\n", "\r\n"))
	}
	r, err := Check(root, e.home)
	if err != nil || len(r.Findings) != 0 || len(r.Warnings) != 0 {
		t.Fatalf("check after CRLF: %v %+v %+v", err, r.Findings, r.Warnings)
	}
	before := snapshot(t, root)
	if p := e.plan(t, root, Request{}); !p.Nothing() {
		t.Errorf("update after CRLF would write:\n%s", p.Preview(false))
	}
	sameSnapshot(t, "an update plan", before, snapshot(t, root))
}

// Check 4: init again, and update to the same commit, change no byte (nor touch a file).
func TestCheck4InitAgainChangesNoByte(t *testing.T) {
	e := setup(t)
	root := testpack.Project(t, e.tmp, "again")
	writeFile(t, root, SettingsFile, drifted)
	e.link(t, root, e.pack.A, "ledger.json")
	before := snapshot(t, root)
	for _, req := range []Request{{Command: "init", Init: e.values(e.pack.A, "ledger.json")}, {Command: "init", Init: &InitValues{}}, {Command: "update"}} {
		p := e.plan(t, root, req)
		if !p.Nothing() {
			t.Errorf("%s again would write:\n%s", req.Command, p.Preview(false))
		}
		if err := Apply(p); err != nil {
			t.Fatal(err)
		}
	}
	sameSnapshot(t, "init again and update", before, snapshot(t, root))
	// Values that differ from bonsai.yaml are refused, not applied.
	if _, err := e.try(root, Request{Command: "init", Init: e.values(e.pack.B)}); err == nil || err.(*Error).Exit != ExitInput {
		t.Errorf("init with another ref: %v", err)
	}
}

// Check 5: update to a second commit. Without --yes the plan names every settings line it would change and
// writes nothing; applied, one file is updated, one created, the rest unchanged; check then finds nothing.
func TestCheck5UpdateAToB(t *testing.T) {
	e := setup(t)
	root := testpack.Project(t, e.tmp, "update")
	e.link(t, root, e.pack.A)
	testpack.SetRef(t, root, e.pack.A, e.pack.B)
	before := snapshot(t, root)
	p := e.plan(t, root, Request{})
	sameSnapshot(t, "a plan", before, snapshot(t, root))
	var got []string
	for _, f := range p.Files {
		got = append(got, f.Result+" "+f.Path)
	}
	want := "created demo/extra.md, updated demo/guide.md, unchanged demo/start.md, unchanged CLAUDE.md, updated .claude/settings.json, " +
		"updated .claude/settings.local.json, unchanged .bonsai/.gitignore"
	if strings.Join(got, ", ") != want {
		t.Errorf("files\n  %s\nwant\n  %s", strings.Join(got, ", "), want)
	}
	// This checkout's plugin turned on at B in place of A's (plugins.go).
	if len(p.Local) != 1 || p.Local[0].Change != "change" || p.Local[0].File != LocalSettingsFile ||
		!strings.HasPrefix(p.Local[0].Line, "demo-pack@bonsai-demo-") || !strings.HasPrefix(p.Local[0].Was, "demo-pack@bonsai-demo-") ||
		p.Local[0].Line == p.Local[0].Was {
		t.Errorf("local lines %+v", p.Local)
	}
	var lines []string
	for _, c := range p.Settings {
		lines = append(lines, c.Change+" "+c.Kind)
		if c.RunsCode || c.Why == "" || c.Was == "" {
			t.Errorf("settings line %+v", c)
		}
	}
	if strings.Join(lines, ", ") != "change marketplace, change plugin" || p.HookChange {
		t.Errorf("settings lines %v, hook change %v", lines, p.HookChange)
	}
	if pv := p.Preview(false); !strings.Contains(pv, "was: bonsai-demo-") || !strings.Contains(pv, "Runs code: no hook line is added or changed.") {
		t.Errorf("preview:\n%s", pv)
	}
	if err := Apply(p); err != nil {
		t.Fatal(err)
	}
	if read(t, root, "demo/guide.md") != "# Guide\n\nEdition 2.\n" || read(t, root, "demo/extra.md") != "# Extra\n\nNew at commit B.\n" {
		t.Errorf("files not at B")
	}
	if s := schema.Show(settingsDocOf(t, root)); !strings.Contains(s, e.pack.B) || strings.Contains(s, e.pack.A) {
		t.Errorf("the wiring is not at B: %s", s)
	}
	r, err := Check(root, e.home)
	if err != nil || len(r.Findings) != 0 {
		t.Errorf("check: %v %+v", err, r.Findings)
	}
}

// Check 6: an edited pack file meets a pack change: a conflict, nothing written, the file named; --keep applies
// the rest; the next commit changing the file is a conflict again; --adopt saves the project's copy in the home's
// cache, not in the repo.
func TestCheck6ConflictKeepAdopt(t *testing.T) {
	e := setup(t)
	root := testpack.Project(t, e.tmp, "conflict")
	e.link(t, root, e.pack.A)
	writeFile(t, root, "demo/guide.md", "# Guide\n\nMy own edition.\n")
	testpack.SetRef(t, root, e.pack.A, e.pack.B)
	before := snapshot(t, root)
	p := e.plan(t, root, Request{})
	if len(p.Conflicts) != 1 || p.Conflicts[0].Path != "demo/guide.md" {
		t.Fatalf("conflicts %+v", p.Conflicts)
	}
	if err := Apply(p); err == nil {
		t.Fatal("a plan with a conflict was applied")
	}
	sameSnapshot(t, "a conflict", before, snapshot(t, root))
	if next := p.ConflictNext("bonsai update"); !strings.Contains(next, "bonsai update --yes --keep demo/guide.md") ||
		!strings.Contains(next, "bonsai update --yes --adopt demo/guide.md") {
		t.Errorf("next: %s", next)
	}

	p = e.apply(t, root, Request{Keep: []string{"demo/guide.md"}})
	if f := result(t, p, "demo/guide.md"); f.Result != Kept || f.Kind != "kept" {
		t.Errorf("guide.md %+v", f)
	}
	if read(t, root, "demo/guide.md") != "# Guide\n\nMy own edition.\n" || read(t, root, "demo/extra.md") == "" {
		t.Errorf("--keep did not keep the edit and apply the rest")
	}
	lock, _ := workspace.LoadLock(root)
	if lock.Files["demo/guide.md"].Kind != "kept" || lock.Packs[0].Commit != e.pack.B {
		t.Errorf("lock %+v %s", lock.Files["demo/guide.md"], lock.Packs[0].Commit)
	}
	if r, _ := Check(root, e.home); len(r.Findings) != 0 {
		t.Errorf("a kept file is a finding: %+v", r.Findings)
	}
	// Still at B, the kept file stays kept.
	if p := e.plan(t, root, Request{}); !p.Nothing() {
		t.Errorf("at B again:\n%s", p.Preview(false))
	}

	testpack.SetRef(t, root, e.pack.B, e.pack.C)
	p = e.plan(t, root, Request{})
	if len(p.Conflicts) != 1 || p.Conflicts[0].Path != "demo/guide.md" {
		t.Fatalf("at C: conflicts %+v", p.Conflicts)
	}
	p = e.apply(t, root, Request{Adopt: []string{"demo/guide.md"}})
	f := result(t, p, "demo/guide.md")
	if f.Result != Replaced || read(t, root, "demo/guide.md") != "# Guide\n\nEdition 3.\n" {
		t.Errorf("--adopt: %+v %q", f, read(t, root, "demo/guide.md"))
	}
	saved, err := os.ReadFile(filepath.FromSlash(f.Saved))
	if err != nil || string(saved) != "# Guide\n\nMy own edition.\n" || !strings.HasPrefix(f.Saved, filepath.ToSlash(e.home)+"/cache/adopted/") {
		t.Errorf("the project's copy: %s %q %v", f.Saved, saved, err)
	}
	if out := testpack.Git(t, root, "status", "--porcelain", "--untracked-files=all"); strings.Contains(out, "My own") ||
		strings.Contains(out, "adopted") {
		t.Errorf("the copy went into the repo: %s", out)
	}
	lock, _ = workspace.LoadLock(root)
	if lock.Files["demo/guide.md"].Kind != "pack" {
		t.Errorf("after --adopt: %+v", lock.Files["demo/guide.md"])
	}
}

// A hook-line change (C to D: only the hook line differs) is refused: the plan says so, Apply refuses it, and
// nothing is written.
func TestHookLineChangeIsRefused(t *testing.T) {
	e := setup(t)
	root := testpack.Project(t, e.tmp, "hook")
	e.link(t, root, e.pack.C)
	testpack.SetRef(t, root, e.pack.C, e.pack.D)
	before := snapshot(t, root)
	p := e.plan(t, root, Request{})
	if !p.HookChange {
		t.Fatalf("no hook change:\n%s", p.Preview(false))
	}
	var runs []string
	for _, c := range p.Settings {
		if c.RunsCode {
			runs = append(runs, c.Change+" "+c.Line+" (was "+c.Was+")")
		}
	}
	if strings.Join(runs, "; ") != "change SessionStart (startup): echo demo hook D (was SessionStart (startup): echo demo hook A)" {
		t.Errorf("runs code: %v", runs)
	}
	if !strings.Contains(p.Preview(false), "needs --allow-exec as well as --yes") {
		t.Errorf("preview:\n%s", p.Preview(false))
	}
	if err := Apply(p); err == nil {
		t.Fatal("a hook-line change was applied")
	}
	sameSnapshot(t, "a refused hook-line change", before, snapshot(t, root))
	// A deleted settings file comes back with the same hook line, which runs no new code.
	if err := os.Remove(filepath.Join(root, filepath.FromSlash(SettingsFile))); err != nil {
		t.Fatal(err)
	}
	testpack.SetRef(t, root, e.pack.D, e.pack.C)
	p = e.apply(t, root, Request{})
	if f := result(t, p, SettingsFile); f.Result != Restored {
		t.Errorf("settings %+v", f)
	}
}

// bonsai.yaml with no lock: update refuses (exit 4, nothing written, bonsai init named), so deleting the lock is no
// way round the hook-line refusal; init links again, previewing the hook lines as a first link.
func TestUpdateWithoutALockIsRefused(t *testing.T) {
	e := setup(t)
	root := testpack.Project(t, e.tmp, "nolock")
	e.link(t, root, e.pack.C)
	if err := os.Remove(filepath.Join(root, filepath.FromSlash(workspace.LockFile))); err != nil {
		t.Fatal(err)
	}
	testpack.SetRef(t, root, e.pack.C, e.pack.D)
	before := snapshot(t, root)
	_, err := e.try(root, Request{Command: "update"})
	var ee *Error
	if !errors.As(err, &ee) || ee.Exit != ExitState || !strings.Contains(ee.Next, "bonsai init") ||
		!strings.Contains(ee.Next, "git checkout -- .bonsai/lock.json") {
		t.Fatalf("update with no lock: %v", err)
	}
	sameSnapshot(t, "update with no lock", before, snapshot(t, root))
	p := e.plan(t, root, Request{Command: "init", Init: &InitValues{}})
	if !p.FirstLink || !strings.Contains(p.Preview(false), "echo demo hook D") {
		t.Errorf("init with no lock is not a first link naming the hook line:\n%s", p.Preview(false))
	}
}

// init --new-id: a copy gets a new id (only the id line of bonsai.yaml changes) and an empty .bonsai/local/.
func TestNewID(t *testing.T) {
	e := setup(t)
	root := testpack.Project(t, e.tmp, "original")
	p := e.link(t, root, e.pack.A)
	writeFile(t, root, ".bonsai/local/log/s-1.ndjson", "{}\n")
	writeFile(t, root, ".bonsai/local/asks/2026-10-08.ndjson", "{}\n")
	old := read(t, root, "bonsai.yaml")
	p2 := e.apply(t, root, Request{Command: "init", NewID: true, Init: &InitValues{}})
	if p2.OldID != p.Config.ID || p2.Config.ID == p.Config.ID || p2.EmptyLocal != 2 {
		t.Errorf("new id %s (was %s), emptied %d", p2.Config.ID, p2.OldID, p2.EmptyLocal)
	}
	if d := Diff("bonsai.yaml", result(t, p2, "bonsai.yaml").old, result(t, p2, "bonsai.yaml").write); !strings.Contains(d, "-id: "+p.Config.ID) ||
		!strings.Contains(d, "+id: "+p2.Config.ID) {
		t.Errorf("the diff of the new id:\n%s", d)
	}
	now := read(t, root, "bonsai.yaml")
	if strings.Replace(old, p.Config.ID, p2.Config.ID, 1) != now {
		t.Errorf("bonsai.yaml changed beyond its id:\n%s", now)
	}
	if entries, err := os.ReadDir(filepath.Join(root, ".bonsai", "local")); err != nil || len(entries) != 0 {
		t.Errorf(".bonsai/local/ holds %v (%v)", entries, err)
	}
	// In a worktree, --new-id is refused.
	testpack.Git(t, root, "add", "-A")
	testpack.Git(t, root, "commit", "-q", "-m", "link")
	wt := filepath.Join(e.tmp, "wt")
	testpack.Git(t, root, "worktree", "add", "-q", wt)
	if _, err := e.try(wt, Request{Command: "init", NewID: true, Init: &InitValues{}}); err == nil || err.(*Error).Exit != ExitState {
		t.Errorf("--new-id in a worktree: %v", err)
	}
}

// Check 12: a plain git revert of the link commit restores the project's old state, with no Bonsai binary.
func TestCheck12RevertTheLink(t *testing.T) {
	e := setup(t)
	root := testpack.Project(t, e.tmp, "revert")
	writeFile(t, root, SettingsFile, drifted)
	writeFile(t, root, "CLAUDE.md", "# Own\n")
	testpack.Git(t, root, "add", "-A")
	testpack.Git(t, root, "commit", "-q", "-m", "before")
	beforeTree := testpack.Git(t, root, "rev-parse", "HEAD^{tree}")
	before := snapshot(t, root)
	e.link(t, root, e.pack.A, "ledger.json")
	// The link commit holds what Bonsai wrote but this checkout's own .claude/settings.local.json, never committed.
	testpack.Git(t, root, "add", "-A", "--", ".", ":(exclude)"+LocalSettingsFile)
	testpack.Git(t, root, "commit", "-q", "-m", "link")
	testpack.Git(t, root, "revert", "--no-edit", "HEAD")
	if tree := testpack.Git(t, root, "rev-parse", "HEAD^{tree}"); tree != beforeTree {
		t.Errorf("the tree after the revert is %s, before the link %s", tree, beforeTree)
	}
	// The revert leaves only that file, machine state outside git (whether git shows it as untracked or ignored
	// depends on the machine's own ignore rules). Bonsai's unlink (step 5.1) takes it away.
	for _, l := range strings.Split(testpack.Git(t, root, "status", "--porcelain", "--untracked-files=all", "--ignored"), "\n") {
		if l != "" && l != "?? "+LocalSettingsFile && l != "!! "+LocalSettingsFile {
			t.Errorf("left after the revert: %s", l)
		}
	}
	if err := os.Remove(filepath.Join(root, filepath.FromSlash(LocalSettingsFile))); err != nil {
		t.Errorf("no %s after the link: %v", LocalSettingsFile, err)
	}
	after := snapshot(t, root)
	for k, v := range before {
		if strings.Fields(after[k])[0] != strings.Fields(v)[0] {
			t.Errorf("%s differs after the revert", k)
		}
	}
	if len(after) != len(before) {
		t.Errorf("files before %d, after %d", len(before), len(after))
	}
}

// What update does to files a person deleted or the pack dropped, and to the files the lock no longer matches.
func TestMissingFiles(t *testing.T) {
	e := setup(t)
	root := testpack.Project(t, e.tmp, "missing")
	e.link(t, root, e.pack.A)
	for _, rel := range []string{"demo/guide.md", "demo/start.md", GitignoreFile} {
		if err := os.Remove(filepath.Join(root, filepath.FromSlash(rel))); err != nil {
			t.Fatal(err)
		}
	}
	r, _ := Check(root, e.home)
	var codes []string
	for _, f := range r.Findings {
		codes = append(codes, f.Code+" "+f.File)
	}
	if strings.Join(codes, ", ") != "missing demo/guide.md, gitignore .bonsai/.gitignore" || r.Missing != 1 {
		t.Errorf("findings %v", codes)
	}
	p := e.apply(t, root, Request{})
	if result(t, p, "demo/guide.md").Result != Restored || result(t, p, "demo/start.md").Result != LeftMissing ||
		result(t, p, GitignoreFile).Result != Restored {
		t.Errorf("plan:\n%s", p.Preview(false))
	}
	if _, err := os.Stat(filepath.Join(root, "demo", "start.md")); !os.IsNotExist(err) {
		t.Errorf("a deleted once file came back")
	}
}

// A pack that drops a file: an unedited file goes, an edited one is a conflict.
func TestPackDropsAFile(t *testing.T) {
	e := setup(t)
	root := testpack.Project(t, e.tmp, "drop")
	e.link(t, root, e.pack.B)
	testpack.SetRef(t, root, e.pack.B, e.pack.A) // A has no extra.md
	p := e.plan(t, root, Request{})
	if f := result(t, p, "demo/extra.md"); f.Result != Removed || !f.Writes() {
		t.Errorf("extra.md %+v", f)
	}
	writeFile(t, root, "demo/extra.md", "edited\n")
	p = e.plan(t, root, Request{})
	if f := result(t, p, "demo/extra.md"); f.Result != Conflict {
		t.Errorf("edited extra.md %+v", f)
	}
	p = e.apply(t, root, Request{Keep: []string{"demo/extra.md"}})
	if f := result(t, p, "demo/extra.md"); f.Result != Released || read(t, root, "demo/extra.md") != "edited\n" {
		t.Errorf("--keep of a dropped file %+v", f)
	}
}

// check's findings: an edited pack file, an edited block, an edited settings line, a tracked .bonsai/local/ file.
func TestCheckFindings(t *testing.T) {
	e := setup(t)
	root := testpack.Project(t, e.tmp, "check")
	e.link(t, root, e.pack.A, "ledger.json")
	writeFile(t, root, "demo/guide.md", "edited\n")
	writeFile(t, root, "CLAUDE.md", strings.Replace(read(t, root, "CLAUDE.md"), "demo-pack:marker", "someone-else", 1))
	writeFile(t, root, SettingsFile, strings.Replace(read(t, root, SettingsFile), `"Edit(ledger.json)",`, "", 1))
	writeFile(t, root, ".bonsai/local/log/x.ndjson", "{}\n")
	testpack.Git(t, root, "add", "-f", ".bonsai/local/log/x.ndjson")
	r, err := Check(root, e.home)
	if err != nil {
		t.Fatal(err)
	}
	var codes []string
	for _, f := range r.Findings {
		codes = append(codes, f.Code+" "+f.File)
		if f.Next == "" {
			t.Errorf("finding %+v names no next step", f)
		}
	}
	want := "changed .claude/settings.json, changed CLAUDE.md, changed demo/guide.md, local .bonsai/local"
	if strings.Join(codes, ", ") != want || r.Changed != 3 || r.Packs[0].State != "changed" {
		t.Errorf("findings %v (want %s), changed %d, packs %+v", codes, want, r.Changed, r.Packs)
	}
	// The settings file's edit is no conflict while Bonsai's lines stay as they are, and --adopt takes them back.
	p := e.plan(t, root, Request{})
	if f := result(t, p, SettingsFile); f.Result != Changed {
		t.Errorf("settings %+v", f)
	}
	p = e.apply(t, root, Request{Adopt: []string{SettingsFile, "CLAUDE.md", "demo/guide.md"}})
	if !strings.Contains(read(t, root, SettingsFile), `"Edit(ledger.json)"`) {
		t.Errorf("--adopt did not take Bonsai's lines back")
	}
	testpack.Git(t, root, "rm", "-q", "--cached", ".bonsai/local/log/x.ndjson")
	if r, _ := Check(root, e.home); len(r.Findings) != 0 {
		t.Errorf("after --adopt: %+v", r.Findings)
	}
	// Without the pack in this machine's cache, the settings file is a warning, not a finding.
	if err := os.RemoveAll(filepath.Join(e.home, "cache", "git")); err != nil {
		t.Fatal(err)
	}
	r, _ = Check(root, e.home)
	if len(r.Findings) != 0 || len(r.Warnings) != 1 || r.Warnings[0].Code != "cache" {
		t.Errorf("with no cache: %+v %+v", r.Findings, r.Warnings)
	}
	if _, err := os.Stat(filepath.Join(e.home, "cache", "git")); !os.IsNotExist(err) {
		t.Errorf("check wrote the cache")
	}
}

// A path taken out of never_edit: its old deny rule goes, with the spec's sentence; a deny rule the project wrote
// itself stays.
func TestNeverEditChange(t *testing.T) {
	e := setup(t)
	root := testpack.Project(t, e.tmp, "never")
	e.link(t, root, e.pack.A, "ledger.json", "old-ledger.json")
	cfg := read(t, root, "bonsai.yaml")
	writeFile(t, root, "bonsai.yaml", strings.Replace(cfg, `["ledger.json", "old-ledger.json"]`, `["ledger.json", "new.json"]`, 1))
	p := e.apply(t, root, Request{})
	var got []string
	for _, c := range p.Settings {
		got = append(got, c.Change+" "+c.Line+": "+c.Why)
	}
	want := "add Edit(new.json): Agents can never edit new.json: it is in never_edit in bonsai.yaml.; " +
		"remove Edit(old-ledger.json): This file is no longer in never_edit in bonsai.yaml, so agents may edit it again."
	if strings.Join(got, "; ") != want {
		t.Errorf("settings lines\n  %s\nwant\n  %s", strings.Join(got, "; "), want)
	}
	// With a deny rule of the project's own beside them, its rule stays whatever never_edit does.
	s := read(t, root, SettingsFile)
	writeFile(t, root, SettingsFile, strings.Replace(s, `"Edit(new.json)"`, `"Edit(new.json)", "Edit(mine.txt)"`, 1))
	writeFile(t, root, "bonsai.yaml", strings.Replace(read(t, root, "bonsai.yaml"), `"new.json"`, `"other.json"`, 1))
	p = e.plan(t, root, Request{})
	if !strings.Contains(schema.Show(func() any { b, _ := schema.Decode(result(t, p, SettingsFile).write); return b }()), "Edit(mine.txt)") {
		t.Errorf("the project's own rule went:\n%s", p.Preview(false))
	}
}

// Old Bonsai's workspace, a pack two packs write, and a pack taken out of bonsai.yaml are refused.
func TestRefusals(t *testing.T) {
	e := setup(t)
	root := testpack.Project(t, e.tmp, "old")
	writeFile(t, root, ".bonsai.yaml", "agents: {}\n")
	if _, err := e.try(root, Request{Command: "init", Init: e.values(e.pack.A)}); err == nil || err.(*Error).Exit != ExitState ||
		!strings.Contains(err.Error(), "Bonsai 0.4.3") {
		t.Errorf("an old workspace: %v", err)
	}
	root = testpack.Project(t, e.tmp, "unlinked")
	if _, err := e.try(root, Request{Command: "update"}); err == nil || err.(*Error).Exit != ExitState ||
		!strings.Contains(err.Error(), "next: link the project first: bonsai init") {
		t.Errorf("update unlinked: %v", err)
	}
	for _, bad := range []*InitValues{
		{Name: "Bad_Name", Source: e.pack.Source, Ref: e.pack.A},
		{Name: "x", Source: "-oops", Ref: e.pack.A},
		{Name: "x", Source: e.pack.Source, Ref: e.pack.A, NeverEdit: []string{"/etc/passwd"}},
		{Name: "x", Source: e.pack.Source, Ref: e.pack.A, Path: "../out"},
	} {
		if _, err := e.try(root, Request{Command: "init", Init: bad}); err == nil || err.(*Error).Exit != ExitInput {
			t.Errorf("init %+v: %v", bad, err)
		}
	}
	if _, err := e.try(root, Request{Command: "init", Init: e.values("no-such-tag")}); err == nil || err.(*Error).Exit != ExitInput {
		t.Errorf("a missing ref: %v", err)
	}
	handwritten := "format: bonsai.workspace/1\nid: ws-aaaaaaaaaaaaaaaaaaaaaaaaaa\nname: demo\npacks:\n  - id: other-pack\n    source: \"" +
		filepath.ToSlash(e.pack.Source) + "\"\n    ref: \"" + e.pack.A + "\"\n"
	writeFile(t, root, "bonsai.yaml", handwritten)
	// init links from a hand-written bonsai.yaml (update, with no lock yet, refuses: TestUpdateWithoutALockIsRefused).
	if _, err := e.try(root, Request{Command: "init", Init: &InitValues{}}); err == nil || !strings.Contains(err.Error(), "says its id is demo-pack") {
		t.Errorf("a pack under another id: %v", err)
	}
	if err := os.Remove(filepath.Join(root, "bonsai.yaml")); err != nil {
		t.Fatal(err)
	}
	e.link(t, root, e.pack.A)
	writeFile(t, root, "bonsai.yaml", "format: bonsai.workspace/1\nid: ws-aaaaaaaaaaaaaaaaaaaaaaaaaa\nname: demo\n")
	if _, err := e.try(root, Request{}); err == nil || err.(*Error).Exit != ExitState || !strings.Contains(err.Error(), "step 5.1") {
		t.Errorf("a pack taken out: %v", err)
	}
	if _, err := e.try(root, Request{Keep: []string{"demo/start.md"}}); err == nil {
		t.Errorf("--keep of a file with no conflict was taken")
	}
}
