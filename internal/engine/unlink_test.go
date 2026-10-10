package engine

// bonsai unlink's plan and write (step 5.1.7): what it removes, what it leaves and names, a second run, a run that
// stopped part-way, init after it, and its refusals.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/testpack"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// ownSettings is a project's own settings file: a deny rule, a hook of its own and a model.
const ownSettings = `{
  "permissions": {"allow": ["Bash(npm test)"], "deny": ["Read(secrets/**)"]},
  "hooks": {"PreToolUse": [{"matcher": "Bash", "hooks": [{"type": "command", "command": "npm run lint-staged"}]}]},
  "model": "opus"
}
`

const ownClaude = "# The project's own instructions\n\nKeep these.\n"

func unlinkPlan(t *testing.T, e *env, root string) *Plan {
	t.Helper()
	p, err := BuildUnlink(root, e.home)
	if err != nil {
		t.Fatalf("unlink: %v", err)
	}
	return p
}

// A linked project with files of its own: unlink takes out exactly what Bonsai wrote, and the checkout is as it was
// before the link (the settings file in Bonsai's encoding, the same JSON), but for .bonsai/local/ and its .gitignore,
// which stay while local/ holds a file, and the files written once.
func TestUnlink(t *testing.T) {
	e := setup(t)
	root := testpack.Project(t, e.tmp, "unlink")
	writeFile(t, root, SettingsFile, ownSettings)
	writeFile(t, root, "CLAUDE.md", ownClaude)
	testpack.Git(t, root, "add", "-A")
	testpack.Git(t, root, "commit", "-q", "-m", "the project's own")
	before := snapshot(t, root)
	e.link(t, root, e.pack.A, "ledger.json")
	writeFile(t, root, ".bonsai/local/log/s-1.ndjson", "{}\n")
	writeFile(t, root, workspace.StateFile, "# STATE\n")
	for _, table := range []string{workspace.TasksTableFile, workspace.SessionsTableFile} { // init wrote them (step 5.1.8)
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(table))); err != nil {
			t.Errorf("init did not write %s: %v", table, err)
		}
	}
	// A person's edit of bonsai.yaml: it goes all the same.
	writeFile(t, root, "bonsai.yaml", read(t, root, "bonsai.yaml")+"# a person's note\n")

	p := unlinkPlan(t, e, root)
	if p.Command != "unlink" || !p.LockRemove || p.LockWrite || len(p.Removed) != 1 || p.NeedsExec() || len(p.Conflicts) != 0 {
		t.Fatalf("the plan: %+v", p)
	}
	for path, want := range map[string]string{"demo/guide.md": Removed, "demo/start.md": Released, "CLAUDE.md": Updated,
		SettingsFile: Updated, workspace.GitignoreFile: Released, "bonsai.yaml": Removed, workspace.TasksTableFile: Removed,
		workspace.SessionsTableFile: Removed} {
		if got := result(t, p, path); got.Result != want {
			t.Errorf("%s: %s (%s), want %s", path, got.Result, got.Why, want)
		}
	}
	if last := p.Files[len(p.Files)-1]; last.Path != "bonsai.yaml" || !strings.Contains(last.Why, "removed even if a person edited it") {
		t.Errorf("bonsai.yaml is not the last file, or its why: %+v", last)
	}
	var lines []string
	for _, c := range p.Settings {
		if c.Change != "remove" {
			t.Errorf("a settings line not removed: %+v", c)
		}
		lines = append(lines, c.Kind+" "+c.Line)
	}
	got := strings.Join(lines, "\n")
	for _, want := range []string{"key autoMemoryEnabled: false", "key disableAllHooks: false", "deny Edit(ledger.json)", "deny Edit(demo/never.txt)",
		"hook PreToolUse (Edit|Write|MultiEdit|NotebookEdit|Bash|PowerShell): bonsai hook guard || exit 2", "hook SessionStart (startup): echo demo hook A",
		"marketplace bonsai-demo-", "plugin demo-pack@bonsai-demo-"} {
		if !strings.Contains(got, want) {
			t.Errorf("the settings lines lack %q:\n%s", want, got)
		}
	}
	if len(lines) != 8 || strings.Contains(got, "secrets") || strings.Contains(got, "lint-staged") {
		t.Errorf("the settings lines:\n%s", got)
	}
	preview := p.Preview(false)
	for _, want := range []string{"  demo-pack 0.1.0  " + e.pack.A[:7] + " -> taken out  (", "  removed      demo/guide.md [pack]: a pack file nobody edited",
		"  released     demo/start.md [once]: the project's file: it stays", "  removed      .bonsai/lock.json: removed last",
		".claude/settings.json: 8 lines\n  remove  key", "Left in place:\n  .bonsai/STATE.md: the project's STATE\n  .bonsai/local/: this checkout's log"} {
		if !strings.Contains(preview, want) {
			t.Errorf("the preview lacks %q:\n%s", want, preview)
		}
	}
	if strings.Contains(preview, "Runs code") {
		t.Errorf("unlink's preview speaks of code:\n%s", preview)
	}
	if err := Apply(p); err != nil {
		t.Fatal(err)
	}

	after := snapshot(t, root)
	for _, path := range []string{"demo/guide.md", "bonsai.yaml", workspace.LockFile, workspace.TasksTableFile, workspace.SessionsTableFile} {
		if _, ok := after[path]; ok {
			t.Errorf("%s is still there", path)
		}
	}
	for _, path := range []string{"demo/start.md", workspace.GitignoreFile, ".bonsai/local/log/s-1.ndjson", workspace.StateFile} {
		if _, ok := after[path]; !ok {
			t.Errorf("%s is gone", path)
		}
	}
	if read(t, root, "CLAUDE.md") != ownClaude {
		t.Errorf("CLAUDE.md is not as it was:\n%q", read(t, root, "CLAUDE.md"))
	}
	want, _ := schema.Decode([]byte(ownSettings))
	if s := settingsDocOf(t, root); schema.Show(s) != schema.Show(want) {
		t.Errorf("the settings are not the project's own:\n%s", read(t, root, SettingsFile))
	}
	for path, h := range before {
		if path != SettingsFile && after[path] != "" && strings.Fields(after[path])[0] != strings.Fields(h)[0] {
			t.Errorf("%s changed", path)
		}
	}

	// A second unlink: nothing to change.
	if p := unlinkPlan(t, e, root); !p.Nothing() || p.Config != nil {
		t.Errorf("a second unlink:\n%s", p.Preview(false))
	}
	// init links again: start.md, left in place, is found and locked as found.
	p = e.link(t, root, e.pack.A, "ledger.json")
	if r := result(t, p, "demo/start.md"); r.Result != Found {
		t.Errorf("start.md at the link again: %s", r.Result)
	}
	for _, table := range []string{workspace.TasksTableFile, workspace.SessionsTableFile} {
		if r := result(t, p, table); r.Result != Created {
			t.Errorf("%s at the link again: %s", table, r.Result)
		}
	}
	if r, err := checkLocal(t, root, e.home); err != nil || len(r.Findings) != 0 {
		t.Errorf("check after the link again: %v %+v", err, r.Findings)
	}
}

// Edited files stay and are named: a pack file and Bonsai's block. Files that held only Bonsai's part go, and the
// folders that leaves empty; .bonsai/.gitignore goes with an empty .bonsai/local/.
func TestUnlinkLeavesEdits(t *testing.T) {
	e := setup(t)
	root := testpack.Project(t, e.tmp, "edits")
	e.link(t, root, e.pack.B)
	writeFile(t, root, "demo/guide.md", "mine\n")
	claude := read(t, root, "CLAUDE.md")
	writeFile(t, root, "CLAUDE.md", strings.Replace(claude, "The demo pack is linked", "The demo pack (edited) is linked", 1))
	p := unlinkPlan(t, e, root)
	for path, want := range map[string]string{"demo/guide.md": Released, "demo/extra.md": Removed, "CLAUDE.md": Released,
		SettingsFile: Removed, workspace.GitignoreFile: Removed} {
		if got := result(t, p, path); got.Result != want {
			t.Errorf("%s: %s (%s), want %s", path, got.Result, got.Why, want)
		}
	}
	if err := Apply(p); err != nil {
		t.Fatal(err)
	}
	if read(t, root, "demo/guide.md") != "mine\n" || !strings.Contains(read(t, root, "CLAUDE.md"), "(edited)") {
		t.Errorf("an edit was not left")
	}
	for _, gone := range []string{".claude", ".bonsai", "demo/extra.md"} {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(gone))); err == nil {
			t.Errorf("%s is still there", gone)
		}
	}
	if _, err := os.Stat(filepath.Join(root, "demo")); err != nil {
		t.Errorf("demo/, which holds files, is gone")
	}
}

// A run that stopped after bonsai.yaml went, the lock still there: the next unlink works from the lock alone, the
// settings lines among them (never_edit's rule found by the lock's fingerprint), and finishes.
func TestUnlinkFinishesFromTheLock(t *testing.T) {
	e := setup(t)
	root := testpack.Project(t, e.tmp, "resume")
	writeFile(t, root, SettingsFile, ownSettings)
	e.link(t, root, e.pack.A, "ledger.json")
	if err := os.Remove(filepath.Join(root, "bonsai.yaml")); err != nil {
		t.Fatal(err)
	}
	p := unlinkPlan(t, e, root)
	if p.Config != nil || p.OldMarket != "" || !p.LockRemove || p.lists("bonsai.yaml") {
		t.Fatalf("the plan: %+v", p)
	}
	if len(p.Settings) != 8 {
		t.Errorf("%d settings lines, want 8:\n%s", len(p.Settings), p.Preview(false))
	}
	if err := Apply(p); err != nil {
		t.Fatal(err)
	}
	want, _ := schema.Decode([]byte(ownSettings))
	if s := settingsDocOf(t, root); schema.Show(s) != schema.Show(want) {
		t.Errorf("the settings:\n%s", read(t, root, SettingsFile))
	}
	if _, err := os.Stat(filepath.Join(root, ".bonsai")); err == nil {
		t.Errorf(".bonsai/ is still there")
	}
}

// Refusals: bonsai.yaml with no lock (Bonsai cannot tell its files from the project's), a lock or bonsai.yaml it does
// not read, a 0.4.3 workspace; each writes nothing.
func TestUnlinkRefusals(t *testing.T) {
	e := setup(t)
	root := testpack.Project(t, e.tmp, "refused")
	e.link(t, root, e.pack.A)
	lock := read(t, root, workspace.LockFile)
	cases := []struct {
		name, file, content, word string
		exit                      int
	}{
		{"no lock", workspace.LockFile, "", "no-lock", ExitState},
		{"a lock Bonsai does not read", workspace.LockFile, "{", "bad-lock", ExitState},
		{"a bonsai.yaml Bonsai does not read", "bonsai.yaml", "id: 0755\n", "bad-config", ExitInput},
		{"a settings file that is not an object", SettingsFile, "[]", "bad-file", ExitInput},
		{"a 0.4.3 workspace", ".bonsai.yaml", "agents: {}\n", "old-workspace", ExitState},
	}
	for _, c := range cases {
		saved, had, _ := readFile(root, c.file)
		if c.content == "" {
			_ = os.Remove(filepath.Join(root, filepath.FromSlash(c.file)))
		} else {
			writeFile(t, root, c.file, c.content)
		}
		snap := snapshot(t, root)
		_, err := BuildUnlink(root, e.home)
		var ue *Error
		if err == nil || !asError(err, &ue) || ue.Code != c.word || ue.Exit != c.exit || (c.word != "old-workspace" && !strings.Contains(ue.Next, "run: ")) {
			t.Errorf("%s: %v", c.name, err)
		}
		sameSnapshot(t, c.name, snap, snapshot(t, root))
		if had {
			writeFile(t, root, c.file, string(saved))
		} else {
			_ = os.Remove(filepath.Join(root, filepath.FromSlash(c.file)))
		}
	}
	if read(t, root, workspace.LockFile) != lock {
		t.Fatal("the lock was not restored")
	}
	// Not in a checkout: not-a-checkout.
	if _, err := BuildUnlink(e.tmp, e.home); err == nil || err.(*Error).Code != "not-a-checkout" {
		t.Errorf("outside a checkout: %v", err)
	}
}

func asError(err error, e **Error) bool {
	x, ok := err.(*Error)
	*e = x
	return ok
}

// withoutBlock gives back the file as it was before withBlock added a block at its end, or put one in place.
func TestWithoutBlock(t *testing.T) {
	for _, own := range []string{"", "# Mine\n", "# Mine\n\n", "# Mine\r\nmore\r\n"} {
		d := &blockDoc{raw: []byte(own), exists: own != ""}
		body := blockStart + "\nx\n" + blockEnd + "\n"
		with := d.withBlock(body)
		root := t.TempDir()
		writeFile(t, root, BlockFile, string(with))
		bd, err := readBlock(root)
		if err != nil || !bd.found {
			t.Fatalf("%q: %v", own, err)
		}
		want := own
		if own == "# Mine\n\n" {
			want = "# Mine\n" // withBlock added no second blank line, so one blank line goes with the block
		}
		if got := string(bd.withoutBlock()); got != want {
			t.Errorf("%q: %q, want %q", own, got, want)
		}
	}
	// A block in the middle: the lines around it stay.
	root := t.TempDir()
	writeFile(t, root, BlockFile, "a\n"+blockStart+"\nx\n"+blockEnd+"\nb\n")
	bd, _ := readBlock(root)
	if got := string(bd.withoutBlock()); got != "a\nb\n" {
		t.Errorf("in the middle: %q", got)
	}
}
