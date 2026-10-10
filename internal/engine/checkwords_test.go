package engine

// check's words in the code, and the pieces of step 5.1.6 a test of their own proves: Claude Code's version read
// from a fake claude (never the real one), the floor, approve_first's rule, the absolute-path and settings-rule
// readers, newer tags, the memory index's import left out of the missing-path finding, update's "nothing to change"
// naming what check finds, and a plugin.json with a version refused.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/testpack"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// Every word the code adds a finding or a warning with (each r.add call's first argument) is in format.CheckWords,
// and every word there is added somewhere: the table and the code cannot drift apart.
func TestCheckWordsInTheCode(t *testing.T) {
	pkgs, err := parser.ParseDir(token.NewFileSet(), ".", func(fi os.FileInfo) bool { return !strings.HasSuffix(fi.Name(), "_test.go") }, 0)
	if err != nil {
		t.Fatal(err)
	}
	used := map[string]string{}
	for _, pkg := range pkgs {
		for name, f := range pkg.Files {
			ast.Inspect(f, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok || len(call.Args) == 0 {
					return true
				}
				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok || sel.Sel.Name != "add" {
					return true
				}
				lit, ok := call.Args[0].(*ast.BasicLit)
				if !ok || lit.Kind != token.STRING {
					t.Errorf("%s: an add whose word is not a literal", name)
					return true
				}
				word, _ := strconv.Unquote(lit.Value)
				used[word] = name
				return true
			})
		}
	}
	for word, where := range used {
		if _, ok := format.CheckWord(word); !ok {
			t.Errorf("%s adds %q, which is not in format.CheckWords", where, word)
		}
	}
	for _, w := range format.CheckWords {
		if _, ok := used[w.Word]; !ok {
			t.Errorf("format.CheckWords has %s, which no code adds", w.Word)
		}
		if w.Kind != "finding" && w.Kind != "warning" {
			t.Errorf("%s: kind %q, want finding or warning", w.Word, w.Kind)
		}
		if w.Who != "agent" && w.Who != "person" {
			t.Errorf("%s: who %q", w.Word, w.Who)
		}
	}
	// The later steps' words are not built yet: none is in the table, and each names its step.
	for _, l := range checkLater {
		if _, ok := format.CheckWord(l.Code); ok || l.Step == "" || l.What == "" {
			t.Errorf("checkLater's %s: in the table already, or no step", l.Code)
		}
	}
}

func TestParseVersionAndFloor(t *testing.T) {
	for in, want := range map[string]string{"2.1.294 (Claude Code)": "2.1.294", "v2.10.3": "2.10.3", "Claude Code 3.0.0-beta": "3.0.0",
		"claude 1.2": "", "": "", "build 2.1.294.5": "", "10.0.0": "10.0.0"} {
		if _, got, _ := ParseVersion(in); got != want {
			t.Errorf("ParseVersion(%q) = %q, want %q", in, got, want)
		}
	}
	if v, from := Floor(nil); v != ClaudeCodeFloor || from != "bonsai" {
		t.Errorf("no lock: %s %s", v, from)
	}
	needs := func(id, cc string) workspace.LockedPack {
		d, _ := (&format.Declares{Needs: &format.PackNeeds{ClaudeCode: &cc}}).Object()
		return workspace.LockedPack{ID: id, Declares: d}
	}
	lock := &workspace.Lock{Packs: []workspace.LockedPack{needs("low", "2.1.0"), needs("high", "2.1.400"), needs("odd", "soon")}}
	if v, from := Floor(lock); v != "2.1.400" || from != "high" {
		t.Errorf("a pack's higher floor: %s %s", v, from)
	}
}

// fakeClaudeBin writes a stand-in claude that prints out (a shell script, or a .cmd file on Windows: no file mode is
// needed there) and returns its path. It is run by its path, never looked up on the PATH, so no test reaches the
// real claude.
func fakeClaudeBin(t *testing.T, out string, fail bool) string {
	t.Helper()
	dir := t.TempDir()
	if runtime.GOOS == "windows" {
		p := filepath.Join(dir, "claude.cmd")
		body := "@echo off\r\necho " + out + "\r\n"
		if fail {
			body += "exit /b 1\r\n"
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		return p
	}
	p := filepath.Join(dir, "claude")
	body := "#!/bin/sh\necho '" + out + "'\n"
	if fail {
		body += "exit 1\n"
	}
	if err := os.WriteFile(p, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	return p
}

// A fake claude --version older than the floor, newer, and unreadable: the warning, nothing, and a warning (plan-5,
// 5.1 done, check 9); a claude that fails or is not on the PATH: a warning. Warnings only: no finding.
func TestClaudeCodeVersion(t *testing.T) {
	cases := []struct {
		out  string
		fail bool
		want string
	}{
		{"2.1.200 (Claude Code)", false, "claude-code-old"},
		{"2.1.294 (Claude Code)", false, ""},
		{"2.2.0 (Claude Code)", false, ""},
		{"Claude Code, the latest", false, "claude-code-unknown"},
		{"2.1.300 (Claude Code)", true, "claude-code-unknown"},
	}
	for _, c := range cases {
		bin := fakeClaudeBin(t, c.out, c.fail)
		r := &CheckResult{Config: &workspace.Config{}}
		CompareClaudeCode(r, func() (string, error) { return ClaudeVersion(bin) })
		got := ""
		for _, w := range r.Warnings {
			got += w.Code
		}
		if got != c.want || len(r.Findings) != 0 {
			t.Errorf("%q (fails %v): warnings %q, want %q; findings %v", c.out, c.fail, got, c.want, r.Findings)
		}
	}
	r := &CheckResult{Config: &workspace.Config{}}
	CompareClaudeCode(r, func() (string, error) { return "", ErrNoClaude })
	if len(r.Warnings) != 1 || r.Warnings[0].Code != "claude-code-unknown" || !strings.Contains(r.Warnings[0].Message, "not on the PATH") {
		t.Errorf("not on the PATH: %+v", r.Warnings)
	}
}

func TestApprovedFirst(t *testing.T) {
	for seq, want := range map[string]bool{
		"todo approved running":              true,
		"todo running":                       false,
		"running":                            false,
		"approved running done":              true,
		"todo approved running plan running": false,
		"plan approved blocked running":      true,
		"plan running approved running":      false,
		"todo approved running plan approved running verify done": true,
	} {
		if got, _ := approvedFirst(strings.Fields(seq)); got != want {
			t.Errorf("%s: %v, want %v", seq, got, want)
		}
	}
}

func TestAbsoluteIn(t *testing.T) {
	for in, want := range map[string]string{"/srv/x": "/srv/x", `C:\Users\x`: `C:\Users\x`, "C:/work": "C:/work", `\\host\share`: `\\host\share`,
		"see /home/example/a.txt": "/home/example/a.txt", "file:///tmp/x": "file:///tmp/x", "the /api/users endpoint": "",
		"~/notes": "", "work/tasks": "", "https://example.com/a": "", "a/b:c": "", "": ""} {
		if got, _ := absoluteIn(in); got != want {
			t.Errorf("absoluteIn(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestRuleProblem(t *testing.T) {
	for _, ok := range []any{"Read", "Read(./secrets/**)", "Bash(npm run test:*)", "mcp__my-server__tool", "Edit(a (b))", "WebFetch(domain:example.com)"} {
		if why := ruleProblem(ok); why != "" {
			t.Errorf("%v: %s", ok, why)
		}
	}
	for _, bad := range []any{"", "Read(", "Read()", " Read", "Read( x)", "read me", "Bash(x) extra", 3.0, nil} {
		if why := ruleProblem(bad); why == "" {
			t.Errorf("%v: valid", bad)
		}
	}
}

func TestNewerTags(t *testing.T) {
	tags := []string{"v0.1.0", "v0.2.0", "v0.10.0", "base-v2.0.0", "base-v0.9.0", "nightly", "v1.0.0-rc1", "0.3.0"}
	for _, c := range []struct{ ref, id, version, want string }{
		{"v0.1.0", "demo", "0.1.0", "v0.2.0 0.3.0 v0.10.0"},
		{"0123456789012345678901234567890123456789", "demo", "0.1.0", "v0.2.0 0.3.0 v0.10.0"},
		{"base-v1.0.0", "base", "1.0.0", "base-v2.0.0"},
		{"v0.10.0", "demo", "0.10.0", ""},
		{"v1", "demo", "not a version", ""},
	} {
		if got := strings.Join(NewerTags(tags, c.ref, c.id, c.version), " "); got != c.want {
			t.Errorf("%s at %s: %q, want %q", c.id, c.ref, got, c.want)
		}
	}
}

// The block imports the memory index before it exists (spec §6, §10): a project linked a moment ago has no
// missing-path finding for it. The same import written outside Bonsai's block is the project's own, and is found.
func TestIndexImportLeftOut(t *testing.T) {
	e := setup(t)
	root := testpack.Project(t, e.tmp, "index")
	e.link(t, root, e.pack.A)
	if !strings.Contains(read(t, root, BlockFile), "@work/memory/INDEX.md") {
		t.Fatalf("the block does not import the index:\n%s", read(t, root, BlockFile))
	}
	if r, err := checkLocal(t, root, e.home); err != nil || findingsOf(r) != "" {
		t.Fatalf("a fresh link: %v %s", err, findingsOf(r))
	}
	writeFile(t, root, BlockFile, read(t, root, BlockFile)+"\nOur notes: @work/memory/INDEX.md\n")
	if r, _ := checkLocal(t, root, e.home); findingsOf(r) != "missing-path CLAUDE.md" {
		t.Errorf("the project's own import: %s", findingsOf(r))
	}
	writeFile(t, root, "work/memory/INDEX.md", "---\nformat: bonsai.memory/1\nid: null\ntitle: Index\nkind: index\nupdated: 2026-10-09\nsource: null\nlabels: {}\n---\n")
	if r, _ := checkLocal(t, root, e.home); findingsOf(r) != "" {
		t.Errorf("the index written: %s", findingsOf(r))
	}
}

// update and check agree (step 5.1.1's verifier: a forged lock left update saying only "nothing to change" while
// check found Bonsai's lines edited): with nothing to write, update names each file it leaves edited, as check does.
func TestUpdateNamesWhatCheckFinds(t *testing.T) {
	e := setup(t)
	root := testpack.Project(t, e.tmp, "forged")
	e.link(t, root, e.pack.C)
	other := testpack.Project(t, e.tmp, "at-d")
	e.link(t, other, e.pack.D)
	lc, _ := workspace.LoadLock(root)
	ld, _ := workspace.LoadLock(other)
	forged := *lc
	forged.Packs = ld.Packs
	files := map[string]workspace.LockedFile{}
	for k, v := range lc.Files {
		files[k] = v
	}
	files[SettingsFile] = ld.Files[SettingsFile]
	forged.Files = files
	if err := workspace.WriteLock(root, &forged); err != nil {
		t.Fatal(err)
	}
	testpack.SetRef(t, root, e.pack.C, e.pack.D)
	p := e.plan(t, root, Request{})
	r, _ := Check(root, e.home)
	if !p.Nothing() || !strings.Contains(p.LeftAlone(), SettingsFile+" [keys]") || !strings.Contains(p.LeftAlone(), "--adopt "+SettingsFile) ||
		!strings.Contains(findingsOf(r), "changed "+SettingsFile) {
		t.Errorf("update:\n%s\ncheck: %s", p.LeftAlone(), findingsOf(r))
	}
	if p := e.plan(t, other, Request{}); p.LeftAlone() != "" {
		t.Errorf("nothing edited, yet: %s", p.LeftAlone())
	}
}

// A pack whose plugin.json carries a version is refused when it is linked (spec §5): a plugin pinned by its commit
// carries none.
func TestPluginVersionRefused(t *testing.T) {
	e := setup(t)
	src, shas := testpack.Fixture(t, e.tmp, "versioned", map[string]string{
		".claude-plugin/plugin.json": `{"name": "versioned", "version": "1.0.0", "description": "A fixture pack."}` + "\n",
		"bonsai/pack.yaml":           packYAML("versioned", " []", " []"),
	})
	root := testpack.Project(t, e.tmp, "versioned-project")
	_, err := e.try(root, Request{Command: "init", Init: &InitValues{Name: "demo", Source: src, Ref: shas[0]}})
	refusedWith(t, err, "carries the version \"1.0.0\"")
}

// The machine record: written by init in the main checkout, not by a worktree's update; a new id is added at its end.
func TestInitRecordsTheCheckout(t *testing.T) {
	e := setup(t)
	root := testpack.Project(t, e.tmp, "recorded")
	p := e.link(t, root, e.pack.A)
	if err := RecordCheckout(p); err != nil {
		t.Fatal(err)
	}
	rec, err := workspace.LoadMachineRecord(e.home, root)
	if err != nil || rec == nil || rec.Path != filepath.ToSlash(root) || rec.Current() != p.Config.ID {
		t.Fatalf("record %+v %v", rec, err)
	}
	testpack.Git(t, root, "add", "-A")
	testpack.Git(t, root, "commit", "-q", "-m", "link")
	wt := filepath.Join(e.tmp, "recorded-wt")
	testpack.Git(t, root, "worktree", "add", "-q", wt)
	wp := e.plan(t, wt, Request{})
	before := read(t, e.home, filepath.Join("workspaces", mustKey(t, root), workspace.MachineRecordFile))
	if err := RecordCheckout(wp); err != nil {
		t.Fatal(err)
	}
	if after := read(t, e.home, filepath.Join("workspaces", mustKey(t, root), workspace.MachineRecordFile)); after != before {
		t.Errorf("a worktree's update wrote the record")
	}
	np := e.apply(t, root, Request{Command: "init", NewID: true, AllowExec: true})
	if err := RecordCheckout(np); err != nil {
		t.Fatal(err)
	}
	rec, _ = workspace.LoadMachineRecord(e.home, root)
	if len(rec.IDs) != 2 || rec.IDs[0].ID != p.Config.ID || rec.Current() != np.Config.ID {
		t.Errorf("after --new-id: %+v", rec.IDs)
	}
	raw, _ := rec.Encode()
	if v, err := schema.Decode(raw); err != nil || !strings.HasPrefix(string(raw), "{\n  \"path\": ") || v == nil {
		t.Errorf("encoded %s %v", raw, err)
	}
}

func mustKey(t *testing.T, main string) string {
	t.Helper()
	k, err := workspace.MachineKey(main)
	if err != nil {
		t.Fatal(err)
	}
	return k
}
