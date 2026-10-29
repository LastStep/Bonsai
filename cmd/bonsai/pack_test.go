package main

// bonsai pack (pack.go; T-6009's plan, steps 2, 3, 7 and 7b, and its stage-1 proof rows): pack set's one preview and
// its all-or-nothing write, the bonsai.yaml line and its kept comment; a plan that takes protections away refused to
// an agent's --yes, with and without a grant, for update after a hand edit too, while its preview lists them; the
// changes output's set-13 fields; the fingerprint and --expect; pack outdated offline from a faked source read.

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/engine"
	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/testpack"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// wallPack is a pack with no hook line (so --yes alone writes it) whose commits move its walls: 1 holds two deny
// rules and a protected path; 2 the same walls and another edition of its file (a move that drops nothing); 3 drops
// the second deny rule and the protected path.
func wallPack(t *testing.T, tmp string) (string, []string) {
	t.Helper()
	yaml := func(deny, protected string) string {
		return "format: bonsai.pack/1\nid: wall-pack\nversion: \"0.1.0\"\nfiles:\n  - path: wall/readme.md\n    from: readme.md\n    kind: pack\n" +
			"hooks: []\ndeny:\n" + deny + "protected: " + protected + "\n"
	}
	both := "  - rule: \"Read(wall/one.txt)\"\n    why: \"Agents cannot read wall/one.txt.\"\n" +
		"  - rule: \"Read(wall/two.txt)\"\n    why: \"Agents cannot read wall/two.txt.\"\n"
	one := "  - rule: \"Read(wall/one.txt)\"\n    why: \"Agents cannot read wall/one.txt.\"\n"
	return testpack.Fixture(t, tmp, "wall-pack",
		map[string]string{".claude-plugin/plugin.json": "{\"name\": \"wall-pack\", \"description\": \"A fixture pack.\"}\n",
			"bonsai/pack.yaml": yaml(both, `["wall/keep.md"]`), "bonsai/files/readme.md": "# Wall\n\nEdition 1.\n"},
		map[string]string{"bonsai/files/readme.md": "# Wall\n\nEdition 2.\n"},
		map[string]string{"bonsai/pack.yaml": yaml(one, "[]"), "bonsai/files/readme.md": "# Wall\n\nEdition 3.\n"})
}

// wallCLI links a project to wall-pack's first commit, its ref line given a comment of the person's own, committed.
func wallCLI(t *testing.T) (*cli, []string) {
	t.Helper()
	t.Setenv("BONSAI_TASK", "")
	t.Setenv("BONSAI_ROLE", "")
	c := newCLI(t)
	src, shas := wallPack(t, c.tmp)
	if code, out, errOut := c.run("", "init", "--name", "demo", "--source", src, "--ref", shas[0], "--yes"); code != 0 {
		t.Fatalf("link: %d\n%s%s", code, out, errOut)
	}
	y := c.read("bonsai.yaml")
	lines := strings.Split(y, "\n")
	for i, l := range lines {
		if strings.Contains(l, `ref: "`+shas[0]+`"`) {
			lines[i] = `    ref: "` + shas[0] + `"   # the person's own note on the ref`
		}
	}
	c.write("bonsai.yaml", strings.Join(lines, "\n"))
	c.commit("linked")
	return c, shas
}

// changesOf reads a --json output as bonsai.changes/1, failing the test when it does not validate.
func changesOf(t *testing.T, out string) *format.Changes {
	t.Helper()
	ch, err := format.ReadChanges([]byte(out))
	if err != nil {
		t.Fatalf("not bonsai.changes/1: %v\n%s", err, out)
	}
	if _, err := ch.Encode(); err != nil {
		t.Fatalf("does not validate: %v", err)
	}
	return ch
}

// One preview, all or nothing: the bonsai.yaml line and the pack's files previewed, nothing written without --yes,
// everything with it; a fetch that fails writes nothing, bonsai.yaml included.
func TestPackSetOnePreviewAllOrNothing(t *testing.T) {
	c, shas := wallCLI(t)
	before := tree(t, c.root)
	code, out, errOut := c.run("", "pack", "set", "wall-pack@"+shas[1])
	if code != 4 || !strings.Contains(out, "bonsai.yaml line") || !strings.Contains(out, shas[1]) ||
		!strings.Contains(out, "wall/readme.md") || !strings.Contains(out, "only the ref changes") || !strings.Contains(flat(out+errOut), "--yes") {
		t.Errorf("the preview: %d\n%s%s", code, out, errOut)
	}
	if tree(t, c.root) != before {
		t.Fatal("a preview wrote")
	}
	// The --json preview: command update, word {pack, set}, config with the line before and after and its comment.
	code, out, _ = c.run("", "pack", "set", "wall-pack@"+shas[1], "--json")
	ch := changesOf(t, out)
	if code != 4 || ch.Command != "update" || ch.Result != "preview" || ch.Word.Name != "pack" || ch.Word.Sub == nil || *ch.Word.Sub != "set" {
		t.Errorf("--json preview: %d %+v", code, ch)
	}
	if ch.Config == nil || ch.Config.Path != "bonsai.yaml" || ch.Config.Before == nil || ch.Config.After == nil ||
		!strings.Contains(*ch.Config.Before, shas[0]) || !strings.Contains(*ch.Config.After, shas[1]) ||
		ch.Config.Comment == nil || *ch.Config.Comment != "# the person's own note on the ref" ||
		!strings.HasSuffix(*ch.Config.After, "# the person's own note on the ref") {
		t.Errorf("config: %+v", ch.Config)
	}
	if len(ch.Loosening) != 0 || ch.Fingerprint == nil || len(*ch.Fingerprint) != 64 {
		t.Errorf("loosening %v, fingerprint %v", ch.Loosening, ch.Fingerprint)
	}
	if tree(t, c.root) != before {
		t.Fatal("a --json preview wrote")
	}
	// A fetch that fails: nothing written, bonsai.yaml included.
	missing := strings.Repeat("0", 40)
	if code, out, errOut := c.run("", "pack", "set", "wall-pack@"+missing, "--yes"); code == 0 || tree(t, c.root) != before {
		t.Errorf("a ref not in the source: exit %d, wrote %v\n%s%s", code, tree(t, c.root) != before, out, errOut)
	}
	// --yes writes it all: bonsai.yaml (the ref only, its comment kept), the pack's file, the lock.
	raw := c.read("bonsai.yaml")
	if code, out, errOut := c.run("", "pack", "set", "wall-pack@"+shas[1], "--yes"); code != 0 {
		t.Fatalf("--yes: %d\n%s%s", code, out, errOut)
	}
	if got, want := c.read("bonsai.yaml"), strings.Replace(raw, shas[0], shas[1], 1); got != want {
		t.Errorf("bonsai.yaml after --yes:\n%s\nwant\n%s", got, want)
	}
	if !strings.Contains(c.read("wall/readme.md"), "Edition 2.") || !strings.Contains(c.read(".bonsai/lock.json"), shas[1]) {
		t.Errorf("the pack's file or the lock not written")
	}
	// Refused before a plan: an unknown pack, a bad argument.
	if code, out, _ := c.run("", "pack", "set", "nope@"+shas[1], "--json"); code != 2 || changesOf(t, out).Word.Name != "pack" {
		t.Errorf("an unknown pack: %d\n%s", code, out)
	}
	if code, _, errOut := c.run("", "pack", "set", "wall-pack"); code != 2 || !strings.Contains(errOut, "<id>@<ref>") {
		t.Errorf("no @ref: %d %s", code, errOut)
	}
}

// Loosening is the person's: a move that drops a deny rule and a protected path, pack remove, and update after a hand
// edit taking a pack out, each refused to an agent's --yes with and without a grant of bonsai.yaml and the lock,
// their preview listing what goes; a move that drops nothing passes for an agent; a person applies all of them.
func TestPackLooseningIsThePersons(t *testing.T) {
	for _, granted := range []bool{false, true} {
		c, shas := wallCLI(t)
		if granted {
			home, _ := workspace.Home()
			md, err := workspace.MachineDir(home, c.root)
			if err != nil {
				t.Fatal(err)
			}
			c.write(filepath.ToSlash(mustRel(t, c.root, filepath.Join(md, workspace.MachineSettingsFile))), `{"status_writes": "command", "status_command": "studio-move"}`)
			c.write("work/tasks/T-0001-the-task.md", taskText("T-0001", "running", "bonsai.yaml", ".bonsai/lock.json"))
			c.commit("a person's grant")
		}
		asAgent(c)
		before := tree(t, c.root)
		// The preview, with no --yes: listed, and a preview (needs-yes), never refused for it.
		code, out, _ := c.run("", "pack", "set", "wall-pack@"+shas[2], "--json")
		ch := changesOf(t, out)
		if code != 4 || ch.Error == nil || ch.Error.Code != "needs-yes" || !hasLoose(ch, "deny", "Read(wall/two.txt)") ||
			!hasLoose(ch, "protected", "wall/keep.md") || hasLoose(ch, "deny", "Read(wall/one.txt)") {
			t.Errorf("granted %v: the loosening preview: %d %+v %+v", granted, code, ch.Error, ch.Loosening)
		}
		code, out, errOut := c.run("", "pack", "set", "wall-pack@"+shas[2])
		if code != 4 || !strings.Contains(out, "Loosens") || !strings.Contains(out, "Read(wall/two.txt)") || strings.Contains(errOut, "person-command") {
			t.Errorf("granted %v: the text preview: %d\n%s%s", granted, code, out, errOut)
		}
		// --yes: refused, the rule named; nothing written.
		code, out, errOut = c.run("", "pack", "set", "wall-pack@"+shas[2], "--yes")
		if code != 4 || !strings.Contains(flat(out+errOut), "person-command") || !strings.Contains(flat(out+errOut), "the deny rule Read(wall/two.txt) (wall-pack)") {
			t.Errorf("granted %v: a loosening set --yes: %d\n%s%s", granted, code, out, errOut)
		}
		code, out, errOut = c.run("", "pack", "remove", "wall-pack", "--yes", "--json")
		if ch := changesOf(t, out); code != 4 || ch.Error == nil || ch.Error.Code != "person-command" || !hasLoose(ch, "pack", "wall-pack") {
			t.Errorf("granted %v: pack remove --yes: %d %s %s", granted, code, out, errOut)
		}
		// update after a hand edit that takes the pack out.
		if now := tree(t, c.root); now != before {
			t.Fatalf("granted %v: a refused run wrote:\n%s\nwas\n%s", granted, now, before)
		}
		y, lock := c.read("bonsai.yaml"), c.read(".bonsai/lock.json")
		c.write("bonsai.yaml", handRemoved(t, y))
		code, out, errOut = c.run("", "update", "--yes")
		if code != 4 || !strings.Contains(flat(out+errOut), "person-command") || !strings.Contains(flat(out+errOut), "the pack wall-pack taken out") ||
			c.read(".bonsai/lock.json") != lock || c.read("wall/readme.md") == "" {
			t.Errorf("granted %v: update --yes after a hand edit: %d\n%s%s", granted, code, out, errOut)
		}
		// --expect opens no gate: the loosening plan's own fingerprint is still refused.
		c.write("bonsai.yaml", y)
		before = tree(t, c.root)
		_, out, _ = c.run("", "pack", "set", "wall-pack@"+shas[2], "--json")
		fp := *changesOf(t, out).Fingerprint
		if code, out, errOut := c.run("", "pack", "set", "wall-pack@"+shas[2], "--yes", "--expect", fp); code != 4 || !strings.Contains(flat(out+errOut), "person-command") {
			t.Errorf("granted %v: --expect with a loosening plan: %d\n%s%s", granted, code, out, errOut)
		}
		if now := tree(t, c.root); now != before {
			t.Fatalf("granted %v: a refused run wrote:\n%s\nwas\n%s", granted, now, before)
		}
		// A move that drops nothing: written for the agent.
		if code, out, errOut := c.run("", "pack", "set", "wall-pack@"+shas[1], "--yes"); code != 0 {
			t.Errorf("granted %v: a set that drops nothing: %d\n%s%s", granted, code, out, errOut)
		}
		// A person applies the rest.
		asPerson(c)
		if code, out, errOut := c.run("", "pack", "set", "wall-pack@"+shas[2], "--yes"); code != 0 || !strings.Contains(c.read("wall/readme.md"), "Edition 3.") {
			t.Errorf("granted %v: a person's loosening set: %d\n%s%s", granted, code, out, errOut)
		}
		if code, out, errOut := c.run("", "pack", "remove", "wall-pack", "--yes"); code != 0 || strings.Contains(c.read("bonsai.yaml"), "wall-pack") {
			t.Errorf("granted %v: a person's pack remove: %d\n%s%s", granted, code, out, errOut)
		}
	}
}

// handRemoved is bonsai.yaml with its one pack entry taken out by hand.
func handRemoved(t *testing.T, y string) string {
	t.Helper()
	edit, err := engine.RemovePack([]byte(y), "wall-pack")
	if err != nil {
		t.Fatal(err)
	}
	return string(edit.Raw)
}

func hasLoose(ch *format.Changes, kind, value string) bool {
	for _, l := range ch.Loosening {
		if l.Kind == kind && l.Value == value {
			return true
		}
	}
	return false
}

// The changes output's set-13 fields for update itself: word {update, null}, no config, the memory and cache lines
// present when update prints them (BUG-017).
func TestUpdateChangesSet13(t *testing.T) {
	c, shas := wallCLI(t)
	home, _ := workspace.Home()
	if err := os.RemoveAll(workspace.PersonalPath(home)); err != nil {
		t.Fatal(err)
	}
	c.write("bonsai.yaml", strings.Replace(c.read("bonsai.yaml"), shas[0], shas[1], 1))
	code, out, errOut := c.run("", "update", "--yes", "--json")
	ch := changesOf(t, out)
	if code != 0 || ch.Command != "update" || ch.Word.Name != "update" || ch.Word.Sub != nil || ch.Config != nil || ch.Result != "applied" {
		t.Errorf("update --yes --json: %d %+v\n%s", code, ch, errOut)
	}
	found := false
	for _, m := range ch.Memory {
		found = found || strings.Contains(m.Line, "wrote the empty personal memory index")
	}
	if !found || ch.PackChanges == nil || ch.CommitLines == nil {
		t.Errorf("memory %+v, pack_changes %v, commit_lines %v", ch.Memory, ch.PackChanges, ch.CommitLines)
	}
}

// The fingerprint: the same plan twice gives the same; a preview without --allow-exec then --yes --allow-exec --expect
// applies (the plan does not depend on the flag); --yes --expect applies; a file edited between the preview and the
// apply refuses plan-changed and writes nothing; update --expect with a pack preview's fingerprint is refused.
func TestPackFingerprintAndExpect(t *testing.T) {
	c, shas := wallCLI(t)
	fpOf := func(args ...string) string {
		t.Helper()
		_, out, _ := c.run("", append(args, "--json")...)
		return *changesOf(t, out).Fingerprint
	}
	a, b := fpOf("pack", "set", "wall-pack@"+shas[1]), fpOf("pack", "set", "wall-pack@"+shas[1])
	if a != b {
		t.Fatalf("the same plan twice: %s and %s", a, b)
	}
	if fpOf("pack", "set", "wall-pack@"+shas[1], "--allow-exec") != a || fpOf("pack", "set", "wall-pack@"+shas[2]) == a {
		t.Errorf("the fingerprint follows a flag, or not the plan")
	}
	// update's preview of the same change, by hand: another word, another fingerprint.
	y := c.read("bonsai.yaml")
	c.write("bonsai.yaml", strings.Replace(y, shas[0], shas[1], 1))
	upd := fpOf("update")
	c.write("bonsai.yaml", y)
	if upd == a {
		t.Errorf("update's and pack's fingerprints of one change are the same")
	}
	before := tree(t, c.root)
	if code, out, errOut := c.run("", "update", "--yes", "--expect", a); code != 4 || !strings.Contains(flat(out+errOut), "plan-changed") || tree(t, c.root) != before {
		t.Errorf("update --expect with pack's fingerprint: %d\n%s%s", code, out, errOut)
	}
	// A file edited between the preview and the apply.
	c.write("wall/readme.md", "# Wall\n\nEdited here.\n")
	edited := tree(t, c.root)
	code, out, _ := c.run("", "pack", "set", "wall-pack@"+shas[1], "--yes", "--expect", a, "--json")
	if ch := changesOf(t, out); code != 4 || ch.Error == nil || ch.Error.Code != "plan-changed" || ch.Error.Next.Who != "person" || tree(t, c.root) != edited {
		t.Errorf("an edit between the preview and the apply: %d %s", code, out)
	}
	c.write("wall/readme.md", "# Wall\n\nEdition 1.\n")
	// The normal flow: preview without --allow-exec, then --yes --allow-exec --expect.
	if code, out, errOut := c.run("", "pack", "set", "wall-pack@"+shas[1], "--yes", "--allow-exec", "--expect", a); code != 0 ||
		!strings.Contains(c.read("wall/readme.md"), "Edition 2.") {
		t.Errorf("--yes --allow-exec --expect: %d\n%s%s", code, out, errOut)
	}
	// And plain --yes --expect.
	fp := fpOf("pack", "set", "wall-pack@"+shas[2])
	if code, out, errOut := c.run("", "pack", "set", "wall-pack@"+shas[2], "--yes", "--expect", fp); code != 0 {
		t.Errorf("--yes --expect: %d\n%s%s", code, out, errOut)
	}
}

// pack add and remove through the command: one entry in, one out, the rest of bonsai.yaml byte for byte.
func TestPackAddAndRemove(t *testing.T) {
	c, _ := wallCLI(t)
	quiet, qs := testpack.QuietPack(t, c.tmp)
	y := c.read("bonsai.yaml")
	if code, out, errOut := c.run("", "pack", "add", "--source", quiet, "--ref", qs[0]); code != 4 || !strings.Contains(out, "a new packs entry") {
		t.Errorf("pack add preview: %d\n%s%s", code, out, errOut)
	}
	if code, out, errOut := c.run("", "pack", "add", "--source", quiet, "--ref", qs[0], "--yes"); code != 0 {
		t.Fatalf("pack add --yes: %d\n%s%s", code, out, errOut)
	}
	added := c.read("bonsai.yaml")
	i, j := strings.Index(added, "  - id: quiet-pack"), strings.Index(added, "# where this project keeps")
	if i < 0 || j < i || added[:i]+added[j:] != y || !strings.Contains(c.read(".bonsai/lock.json"), qs[0]) {
		t.Errorf("after pack add:\n%s", added)
	}
	if code, out, errOut := c.run("", "pack", "add", "--source", quiet, "--ref", qs[0], "--yes"); code != 2 {
		t.Errorf("a pack already named: %d\n%s%s", code, out, errOut)
	}
	if code, out, errOut := c.run("", "pack", "remove", "quiet-pack", "--yes"); code != 0 || c.read("bonsai.yaml") != y {
		t.Errorf("pack remove --yes: %d\n%s%s\n%s", code, out, errOut, c.read("bonsai.yaml"))
	}
}

// pack outdated, offline: the source's tags faked (never the network), the cache written, plain status then showing
// the same; --json validates as bonsai.packs/1; a failed read is named and still exits 0.
func TestPackOutdated(t *testing.T) {
	c, _ := wallCLI(t)
	source := sourceOf(t, c)
	reads := 0
	old := packTags
	t.Cleanup(func() { packTags = old })
	packTags = func(s string) ([]string, error) {
		reads++
		if s != source {
			t.Errorf("read %s, want the locked source %s", s, source)
		}
		return []string{"v0.1.0", "v0.2.0", "other-v9.0.0"}, nil
	}
	code, out, errOut := c.run("", "pack", "outdated", "--json")
	doc, err := format.ReadPacks([]byte(out))
	if err != nil || code != 0 || reads != 1 || len(doc.Packs) != 1 {
		t.Fatalf("pack outdated --json: %d %v (reads %d)\n%s%s", code, err, reads, out, errOut)
	}
	p := doc.Packs[0]
	if p.ID != "wall-pack" || p.Newer == nil || p.Newer.State != engine.PackNewer || p.Newer.Tag != "v0.2.0" || p.Why != nil ||
		strings.Join(p.Deny, ",") != "Read(wall/one.txt),Read(wall/two.txt)" || strings.Join(p.Protected, ",") != "wall/keep.md" || p.Skills == nil {
		t.Errorf("the pack: %+v", p)
	}
	if code, out, _ := c.run("", "pack", "outdated"); code != 0 || !strings.Contains(flat(out), "v0.2.0 is released: bonsai pack set wall-pack@v0.2.0") {
		t.Errorf("the text: %d\n%s", code, out)
	}
	// Plain status shows it, offline.
	_, st, _ := c.run("", "status", "--json")
	if !strings.Contains(st, `"newer": {`) || !strings.Contains(st, `"tag": "v0.2.0"`) {
		t.Errorf("status after pack outdated:\n%s", st)
	}
	// A failed read: named, the cache's last good read kept, exit 0.
	packTags = func(string) ([]string, error) { return nil, errors.New("git ls-remote failed: no network") }
	code, out, _ = c.run("", "pack", "outdated", "--json")
	doc, err = format.ReadPacks([]byte(out))
	if err != nil || code != 0 || doc.Packs[0].Why == nil || doc.Packs[0].Newer == nil {
		t.Errorf("a failed read: %d %v\n%s", code, err, out)
	}
	if v, _ := schema.Decode([]byte(out)); v == nil {
		t.Error("not JSON")
	}
}

// sourceOf is the locked pack's source.
func sourceOf(t *testing.T, c *cli) string {
	t.Helper()
	lock, err := workspace.LoadLock(c.root)
	if err != nil || len(lock.Packs) == 0 {
		t.Fatalf("no lock: %v", err)
	}
	return lock.Packs[0].Source
}
