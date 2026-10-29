package main

// bonsai pack's second stage (T-6011; the plan P-T-6009's stage-2 proof rows, the director's eight picks of 28 Oct):
// what changed in set's preview and outdated --changes, the session nudge, set --all, the downgrade guard, the
// compatibility check, the old marketplace's hint, the commit lines, pack show. Each test links one scratch project
// to small fixture packs (lightCLI: no test pack is built) and never reaches the network: the sources' tags are
// faked (packTags), and a fake that fails the test stands where nothing may read a source.

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/engine"
	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/recorder"
	"github.com/LastStep/Bonsai/internal/testpack"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// lightCLI is newCLI without the test pack (testpack.Build): an isolated home and an empty project, for the tests
// that link their own fixture packs.
func lightCLI(t *testing.T) *cli {
	t.Helper()
	tmp := t.TempDir()
	if r, err := filepath.EvalSymlinks(tmp); err == nil {
		tmp = r
	}
	testpack.Isolate(t, tmp)
	t.Setenv("BONSAI_TASK", "")
	t.Setenv("BONSAI_ROLE", "")
	c := &cli{t: t, tmp: tmp}
	c.root = testpack.Project(t, tmp, "project")
	t.Chdir(c.root)
	return c
}

// tagPack is a pack named id with three releases, tagged v0.1.0 to v0.3.0, each commit's pack.yaml naming its own
// version and writing <id>/readme.md; the third drops its deny rule when loosen is set. extra commits follow,
// untagged, each a pack.yaml text (the version and its tag are the caller's).
func tagPack(t *testing.T, tmp, id string, loosen bool, extra ...string) (string, []string) {
	t.Helper()
	deny := "deny:\n  - rule: \"Read(" + id + "/secret.txt)\"\n    why: \"Agents cannot read " + id + "/secret.txt.\"\n"
	commits := []map[string]string{{".claude-plugin/plugin.json": "{\"name\": \"" + id + "\", \"description\": \"A fixture pack.\"}\n"}}
	for i, v := range []string{"0.1.0", "0.2.0", "0.3.0"} {
		if i > 0 {
			commits = append(commits, map[string]string{})
		}
		d := deny
		if loosen && i == 2 {
			d = "deny: []\n"
		}
		commits[i]["bonsai/pack.yaml"] = tagPackYAML(id, v, d, "")
		commits[i]["bonsai/files/readme.md"] = "# " + id + "\n\nRelease " + v + ".\n"
	}
	for _, y := range extra {
		commits = append(commits, map[string]string{"bonsai/pack.yaml": y})
	}
	src, shas := testpack.Fixture(t, tmp, id, commits...)
	for i, v := range []string{"v0.1.0", "v0.2.0", "v0.3.0"} {
		testpack.Tag(t, src, v, shas[i])
	}
	return src, shas
}

// tagPackYAML is tagPack's pack.yaml: format line, id, version, its one file, deny and, when given, a needs block.
func tagPackYAML(id, version, deny, needs string) string {
	return "format: bonsai.pack/1\nid: " + id + "\nversion: \"" + version + "\"\nfiles:\n  - path: " + id + "/readme.md\n    from: readme.md\n" +
		"    kind: pack\nhooks: []\n" + deny + "protected: []\n" + needs
}

// linkPacks links the project to each source at its ref, in order, as a person, and commits the link.
func (c *cli) linkPacks(pairs ...string) {
	c.t.Helper()
	args := []string{"init", "--name", "demo"}
	for i := 0; i+1 < len(pairs); i += 2 {
		args = append(args, "--source", pairs[i], "--ref", pairs[i+1])
	}
	if code, out, errOut := c.run("", append(args, "--yes")...); code != 0 {
		c.t.Fatalf("link: %d\n%s%s", code, out, errOut)
	}
	c.commit("linked")
}

// noSourceRead makes any read of a source's tags fail the test.
func noSourceRead(t *testing.T) {
	t.Helper()
	old := packTags
	t.Cleanup(func() { packTags = old })
	packTags = func(s string) ([]string, error) {
		t.Errorf("a source was read: %s", s)
		return nil, errors.New("no read here")
	}
}

// fakeTags gives every source the three releases tagPack makes.
func fakeTags(t *testing.T) {
	t.Helper()
	old := packTags
	t.Cleanup(func() { packTags = old })
	packTags = func(string) ([]string, error) { return []string{"v0.1.0", "v0.2.0", "v0.3.0"}, nil }
}

// What changed (pick 1): set's preview from the first tag to the third lists the two titles between, newest first,
// in text and in pack_changes; outdated --changes lists the same for the pack behind (text only: --json refused).
// The engine's own test holds a path-limited pack and a read that fails (TestPackChangesBetweenCommits).
func TestPackSetListsWhatChanged(t *testing.T) {
	c := lightCLI(t)
	src, _ := tagPack(t, c.tmp, "tag-pack", false)
	c.linkPacks(src, "v0.1.0")
	code, out, _ := c.run("", "pack", "set", "tag-pack@v0.3.0", "--json")
	ch := changesOf(t, out)
	if code != 4 || len(ch.PackChanges) != 1 || ch.PackChanges[0].Pack != "tag-pack" || strings.Join(ch.PackChanges[0].Titles, "|") != "commit 3|commit 2" ||
		ch.PackChanges[0].More != 0 || ch.PackChanges[0].Why != nil {
		t.Errorf("pack_changes: %d %+v", code, ch.PackChanges)
	}
	_, out, _ = c.run("", "pack", "set", "tag-pack@v0.3.0")
	if s := flat(out); !strings.Contains(s, "Changes tag-pack commit 3 tag-pack commit 2") {
		t.Errorf("the text preview:\n%s", out)
	}
	fakeTags(t)
	code, out, errOut := c.run("", "pack", "outdated", "--changes")
	if s := flat(out); code != 0 || !strings.Contains(s, "Changes tag-pack to v0.3.0 commit 3 tag-pack to v0.3.0 commit 2") {
		t.Errorf("outdated --changes: %d\n%s%s", code, out, errOut)
	}
	code, out, _ = c.run("", "pack", "outdated", "--changes", "--json")
	if doc, err := format.ReadPacks([]byte(out)); err != nil || code != 2 || doc.Error == nil || !strings.Contains(doc.Error.Message, "text only") {
		t.Errorf("outdated --changes --json: %d %v\n%s", code, err, out)
	}
}

// The downgrade guard (pick 4): an older release tag is refused pack-downgrade without --downgrade, nothing written,
// and applied with it; a commit and a tag with no version number are never refused as one, the preview saying so.
func TestPackDowngradeGuard(t *testing.T) {
	c := lightCLI(t)
	src, shas := tagPack(t, c.tmp, "tag-pack", false)
	testpack.Tag(t, src, "edge", shas[0])
	c.linkPacks(src, "v0.2.0")
	before := tree(t, c.root)
	code, out, _ := c.run("", "pack", "set", "tag-pack@v0.1.0", "--yes", "--json")
	ch := changesOf(t, out)
	if code != 4 || ch.Error == nil || ch.Error.Code != "pack-downgrade" || ch.Error.Next.Do != "to go back, run: bonsai pack set tag-pack@v0.1.0 --downgrade" ||
		tree(t, c.root) != before {
		t.Errorf("an older tag: %d %+v", code, ch.Error)
	}
	for _, ref := range []string{shas[0], "edge"} {
		code, out, errOut := c.run("", "pack", "set", "tag-pack@"+ref)
		if code != 4 || strings.Contains(out+errOut, "pack-downgrade") || !strings.Contains(flat(out), "has no order and is not judged a downgrade") {
			t.Errorf("%s: %d\n%s%s", ref, code, out, errOut)
		}
	}
	if code, out, errOut := c.run("", "pack", "set", "tag-pack@v0.1.0", "--downgrade", "--yes"); code != 0 ||
		!strings.Contains(c.read(".bonsai/lock.json"), shas[0]) || !strings.Contains(c.read("tag-pack/readme.md"), "Release 0.1.0.") {
		t.Errorf("--downgrade --yes: %d\n%s%s", code, out, errOut)
	}
}

// The compatibility check (pick 5, the director's "Refuse, change the rule"): a version needing a newer Claude Code
// than the faked one, and one in a pack format this Bonsai does not read, are refused pack-needs before anything is
// written; a Claude Code version that cannot be read is only noted; update, check and status still only warn.
func TestPackNeedsRefused(t *testing.T) {
	c := lightCLI(t)
	deny := "deny: []\n"
	src, shas := tagPack(t, c.tmp, "tag-pack", false,
		tagPackYAML("tag-pack", "0.4.0", deny, "needs:\n  claude_code: \"99.0.0\"\n"),
		strings.Replace(tagPackYAML("tag-pack", "0.5.0", deny, ""), "bonsai.pack/1", "bonsai.pack/2", 1))
	testpack.Tag(t, src, "v0.4.0", shas[3])
	testpack.Tag(t, src, "v0.5.0", shas[4])
	c.linkPacks(src, "v0.1.0")
	fakeClaude(t, "2.1.300 (Claude Code)", nil)
	before := tree(t, c.root)
	code, out, _ := c.run("", "pack", "set", "tag-pack@v0.4.0", "--yes", "--json")
	if ch := changesOf(t, out); code != 4 || ch.Error == nil || ch.Error.Code != "pack-needs" || ch.Error.Next.Who != "person" ||
		!strings.Contains(ch.Error.Message, "needs Claude Code 99.0.0, and this machine runs 2.1.300") || tree(t, c.root) != before {
		t.Errorf("a newer Claude Code needed: %d %+v", code, ch.Error)
	}
	code, out, errOut := c.run("", "pack", "set", "tag-pack@v0.5.0", "--yes")
	if code != 4 || !strings.Contains(flat(errOut), "pack-needs") || !strings.Contains(flat(errOut), "bonsai.pack/2, a pack format this Bonsai does not read") ||
		tree(t, c.root) != before {
		t.Errorf("a pack format not read: %d\n%s%s", code, out, errOut)
	}
	// Claude Code's version not read: noted, not refused (a preview, then the write).
	fakeClaude(t, "", engine.ErrNoClaude)
	if code, out, errOut := c.run("", "pack", "set", "tag-pack@v0.4.0"); code != 4 || strings.Contains(out+errOut, "pack-needs") ||
		!strings.Contains(flat(out), "needs Claude Code 99.0.0, and Claude Code is not on the PATH, so it was not compared (nothing refuses for it)") {
		t.Errorf("a version not read: %d\n%s%s", code, out, errOut)
	}
	// update applies the same move (it never refuses for a need); check then warns, its exit as before; status is 0.
	fakeClaude(t, "2.1.300 (Claude Code)", nil)
	checkBefore, _, _ := c.run("", "check", "--json")
	c.write("bonsai.yaml", strings.Replace(c.read("bonsai.yaml"), "v0.1.0", "v0.4.0", 1))
	if code, out, errOut := c.run("", "update", "--yes"); code != 0 {
		t.Fatalf("update to a version with a need: %d\n%s%s", code, out, errOut)
	}
	c.commit("moved by hand")
	code, out, errOut = c.run("", "check", "--json")
	if code != checkBefore || !strings.Contains(out, `"claude-code-old"`) {
		t.Errorf("check after the move: %d (was %d)\n%s%s", code, checkBefore, out, errOut)
	}
	if code, _, _ := c.run("", "status"); code != 0 {
		t.Errorf("status after the move: %d", code)
	}
}

// The commit lines (pick 7) and the old marketplace's hint (pick 6): after --yes, git add with every file written and
// git commit, one per line, never &&, in text and commit_lines; for update as well; and, the marketplace name having
// moved, the old one named with the command that removes it, a person's step, nothing removed and nothing more asked
// of Claude Code.
func TestPackCommitLinesAndOldMarketplace(t *testing.T) {
	c := lightCLI(t)
	src, shas := tagPack(t, c.tmp, "tag-pack", false)
	c.linkPacks(src, "v0.1.0")
	f := &fakePlugins{install: engine.InstallResult{Outcome: "ok"}}
	defer func(p engine.PluginCLI) { pluginCLI = p }(pluginCLI)
	pluginCLI = f
	oldMarket := engine.MarketplaceName("demo", []string{shas[0]})
	code, out, errOut := c.run("", "pack", "set", "tag-pack@v0.2.0", "--yes", "--json")
	ch := changesOf(t, out)
	if code != 0 || len(ch.CommitLines) != 2 || !strings.HasPrefix(ch.CommitLines[0], "git add bonsai.yaml ") ||
		!strings.Contains(ch.CommitLines[0], " tag-pack/readme.md ") || !strings.HasSuffix(ch.CommitLines[0], " .bonsai/lock.json") ||
		ch.CommitLines[1] != `git commit -m "bonsai: tag-pack to v0.2.0"` {
		t.Errorf("commit_lines: %d %q\n%s", code, ch.CommitLines, errOut)
	}
	for _, l := range ch.CommitLines {
		if strings.Contains(l, "&&") || strings.Contains(l, `\`) {
			t.Errorf("a commit line joined or with a backslash: %s", l)
		}
	}
	var hint *format.ChangesPlugin
	for i, p := range ch.Plugins {
		if p.Result == engine.OldMarket {
			hint = &ch.Plugins[i]
		}
	}
	if hint == nil || hint.Plugin != "tag-pack@"+oldMarket || hint.Next == nil || hint.Next.Who != "person" ||
		hint.Next.Do != "once no other checkout of this project uses it, run: claude plugin marketplace remove "+oldMarket {
		t.Errorf("the old marketplace's hint: %+v", ch.Plugins)
	}
	for _, a := range f.asked {
		if strings.HasPrefix(a, "marketplaces") || strings.HasPrefix(a, "uninstall") {
			t.Errorf("Claude Code asked %q: %v", a, f.asked)
		}
	}
	// The text, by update after a hand edit: the Commit lines.
	c.commit("moved")
	c.write("bonsai.yaml", strings.Replace(c.read("bonsai.yaml"), "v0.2.0", "v0.3.0", 1))
	code, out, errOut = c.run("", "update", "--yes")
	if s := flat(out); code != 0 || !strings.Contains(s, "Commit to commit what was written, run each line: git add bonsai.yaml") ||
		!strings.Contains(s, `git commit -m "bonsai: tag-pack to v0.3.0"`) || !strings.Contains(s, "claude plugin marketplace remove") {
		t.Errorf("update --yes's text: %d\n%s%s", code, out, errOut)
	}
	// A preview has none.
	c.commit("moved again")
	if _, out, _ := c.run("", "pack", "set", "tag-pack@v0.2.0", "--downgrade", "--json"); len(changesOf(t, out).CommitLines) != 0 {
		t.Errorf("a preview's commit_lines: %s", out)
	}
}

// Update all (pick 3): two packs behind and one at a commit: one preview, the two ref lines changed and the pinned
// pack named and left; a loosening in one of them refuses the agent's --yes under a grant; a person applies it, only
// the two refs changing in bonsai.yaml; then nothing is behind.
func TestPackSetAll(t *testing.T) {
	c := lightCLI(t)
	a, _ := tagPack(t, c.tmp, "a-pack", true)
	b, _ := tagPack(t, c.tmp, "b-pack", false)
	p, pin := tagPack(t, c.tmp, "pin-pack", false)
	c.linkPacks(a, "v0.1.0", b, "v0.2.0", p, pin[0])
	fakeTags(t)
	home, _ := workspace.Home()
	md, err := workspace.MachineDir(home, c.root)
	if err != nil {
		t.Fatal(err)
	}
	c.write(filepath.ToSlash(mustRel(t, c.root, filepath.Join(md, workspace.MachineSettingsFile))), `{"status_writes": "command", "status_command": "studio-move"}`)
	c.write("work/tasks/T-0001-the-task.md", taskText("T-0001", "running", "bonsai.yaml", ".bonsai/lock.json"))
	c.commit("a person's grant")
	y := c.read("bonsai.yaml")
	asAgent(c)
	code, out, errOut := c.run("", "pack", "set", "--all")
	if s := flat(out); code != 4 || strings.Count(s, "+ bonsai.yaml line") != 2 || !strings.Contains(s, "v0.3.0") ||
		!strings.Contains(s, "the pack pin-pack is left at "+pin[0]) || !strings.Contains(s, "Loosens") {
		t.Errorf("the --all preview: %d\n%s%s", code, out, errOut)
	}
	before := tree(t, c.root)
	code, out, errOut = c.run("", "pack", "set", "--all", "--yes")
	if code != 4 || !strings.Contains(flat(out+errOut), "person-command") || !strings.Contains(flat(out+errOut), "bonsai pack set --all --yes") ||
		tree(t, c.root) != before {
		t.Errorf("an agent's --all --yes with a loosening: %d\n%s%s", code, out, errOut)
	}
	asPerson(c)
	if code, out, errOut := c.run("", "pack", "set", "--all", "--yes"); code != 0 {
		t.Fatalf("a person's --all --yes: %d\n%s%s", code, out, errOut)
	}
	want := strings.Replace(strings.Replace(y, `ref: "v0.1.0"`, `ref: "v0.3.0"`, 1), `ref: "v0.2.0"`, `ref: "v0.3.0"`, 1)
	if got := c.read("bonsai.yaml"); got != want || !strings.Contains(got, pin[0]) {
		t.Errorf("bonsai.yaml after --all:\n%s\nwant\n%s", got, want)
	}
	code, out, _ = c.run("", "pack", "set", "--all", "--json")
	if ch := changesOf(t, out); code != 0 || ch.Result != "nothing" || ch.Error != nil || ch.Word.Sub == nil || *ch.Word.Sub != "set" {
		t.Errorf("nothing behind: %d %+v", code, ch)
	}
	if code, out, _ := c.run("", "pack", "set", "--all"); code != 0 || !strings.Contains(out, "no pack has a newer release") {
		t.Errorf("nothing behind, text: %d\n%s", code, out)
	}
}

// Pack details (pick 8): pack show prints one pack from the lock, the home's clone and the cache, offline; its --json
// is bonsai.packs/1 with command show; a pack the lock does not hold is bad input. And the session nudge (pick 2):
// hook start's context names the packs the cache says are behind, nothing with no cache or none behind, and reads
// no source.
func TestPackShowAndTheSessionNudge(t *testing.T) {
	c := lightCLI(t)
	src, shas := tagPack(t, c.tmp, "tag-pack", false)
	c.linkPacks(src, "v0.1.0")
	noSourceRead(t)
	home, _ := workspace.Home()
	start := func() string {
		t.Helper()
		var b bytes.Buffer
		recorder.Start(recorder.Options{Stdin: strings.NewReader(""), More: startMore,
			Getenv: func(k string) string {
				if k == "CLAUDE_PROJECT_DIR" {
					return c.root
				}
				return os.Getenv(k)
			}}, &b)
		return b.String()
	}
	if s := start(); !strings.Contains(s, "Bonsai: this project is the workspace demo") || strings.Contains(s, "Packs behind") {
		t.Errorf("no cache:\n%s", s)
	}
	code, out, _ := c.run("", "pack", "show", "tag-pack", "--json")
	doc, err := format.ReadPacks([]byte(out))
	if err != nil || code != 0 || doc.Command != "show" || len(doc.Packs) != 1 || doc.Packs[0].Commit != shas[0] || doc.Packs[0].Newer != nil ||
		strings.Join(doc.Packs[0].Deny, ",") != "Read(tag-pack/secret.txt)" || doc.Packs[0].Ref == nil || *doc.Packs[0].Ref != "v0.1.0" {
		t.Fatalf("pack show --json: %d %v\n%s", code, err, out)
	}
	if _, err := doc.Encode(); err != nil {
		t.Errorf("pack show --json does not validate: %v", err)
	}
	// The cache, as pack outdated writes it: one release current, then newer ones.
	if _, err := engine.RecordPackTags(home, []engine.PackTagRead{{Source: src, Tags: []string{"v0.1.0"}}}); err != nil {
		t.Fatal(err)
	}
	if s := start(); strings.Contains(s, "Packs behind") {
		t.Errorf("none behind:\n%s", s)
	}
	if _, err := engine.RecordPackTags(home, []engine.PackTagRead{{Source: src, Tags: []string{"v0.1.0", "v0.2.0", "v0.3.0"}}}); err != nil {
		t.Fatal(err)
	}
	if s := start(); !strings.Contains(s, "\nPacks behind: tag-pack v0.1.0 -> v0.3.0 (bonsai pack outdated)\n") {
		t.Errorf("one behind:\n%s", s)
	}
	code, out, _ = c.run("", "pack", "show", "tag-pack")
	if s := flat(out); code != 0 || !strings.Contains(s, "newer v0.3.0 is released") || !strings.Contains(s, "deny Read(tag-pack/secret.txt)") ||
		!strings.Contains(s, "commit "+shas[0]) {
		t.Errorf("pack show:\n%s", out)
	}
	if code, out, _ := c.run("", "pack", "show", "nope", "--json"); code != 2 {
		t.Errorf("an unknown pack: %d\n%s", code, out)
	}
}
