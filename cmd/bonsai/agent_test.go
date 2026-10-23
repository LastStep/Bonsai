package main

// Bonsai's own commands refusing an agent (design/plan-5.md, 5.3.6 note 8, under the director's answer (ii)), and a project
// linked by the build before taking the new lines only with --allow-exec --yes (note 2, "Whichever"). Every test here
// sets or clears the agent markers itself (newCLI's testpack.Isolate clears every one; asAgent sets
// CLAUDE_CODE_CHILD_SESSION), so the suite behaves the same inside any agent's session and in CI. Each other agent's
// marker is tested in markers_test.go.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/engine"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/testpack"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// asAgent makes the rest of a test an agent's session: CLAUDE_CODE_CHILD_SESSION=1, as Claude Code sets it.
func asAgent(c *cli) { c.t.Setenv("CLAUDE_CODE_CHILD_SESSION", "1") }

// asPerson makes the rest of a test a person's terminal: every agent marker cleared (testpack.AsPerson).
func asPerson(c *cli) { testpack.AsPerson(c.t) }

// oldGuardMatcher is the guard's matcher every build before step 5.3.6 wrote, without Monitor.
const oldGuardMatcher = "Edit|Write|MultiEdit|NotebookEdit|Bash|PowerShell"

// linkedByName turns a project this build linked into one the build before step 5.3.6 linked: each of Bonsai's own
// hook lines calls bonsai by name (the guard's and the stop gate's with || exit 2), the guard's matcher has no
// Monitor and no onFailure, and the lock's fingerprint is that file's (fingerprint, checked against the engine's
// first by the same reading linkedWithout uses).
func (c *cli) linkedByName() {
	c.t.Helper()
	c.linkedWithout(func(string) bool { return false })
	v, err := schema.Decode([]byte(c.read(".claude/settings.json")))
	if err != nil {
		c.t.Fatal(err)
	}
	doc := v.(schema.Object)
	hooks, _ := doc.Get("hooks")
	for _, ev := range hooks.(schema.Object) {
		for _, g := range ev.Value.([]any) {
			gObj := g.(schema.Object)
			if gObj.String("matcher") == engine.GuardMatcher {
				gObj[gObj.Index("matcher")].Value = oldGuardMatcher
			}
			hs, _ := gObj.Get("hooks")
			for i, h := range hs.([]any) {
				ho := h.(schema.Object)
				cmd := ho.String("command")
				if !ownCommand(cmd) {
					continue
				}
				word := cmd[strings.Index(cmd, `"$b" hook `)+len(`"$b" hook `):]
				word = strings.TrimSuffix(word, " || exit 2")
				byName := "bonsai hook " + word
				if word == "guard" || word == "stop" {
					byName += " || exit 2"
				}
				var out schema.Object
				for _, m := range ho {
					switch m.Key {
					case "onFailure":
					case "command":
						out = append(out, schema.Member{Key: m.Key, Value: byName})
					default:
						out = append(out, m)
					}
				}
				hs.([]any)[i] = out
			}
		}
	}
	raw, err := schema.EncodeUTF8(doc)
	if err != nil {
		c.t.Fatal(err)
	}
	c.write(".claude/settings.json", string(raw))
	lockV, err := schema.Decode([]byte(c.read(workspace.LockFile)))
	if err != nil {
		c.t.Fatal(err)
	}
	lock := lockV.(schema.Object)
	files, _ := lock.Get("files")
	e, _ := files.(schema.Object).Get(".claude/settings.json")
	eo := e.(schema.Object)
	eo[eo.Index("sha256")].Value = fingerprint(doc)
	if raw, err = schema.Encode(lock); err != nil {
		c.t.Fatal(err)
	}
	c.write(workspace.LockFile, string(raw))
	if s := c.read(".claude/settings.json"); !strings.Contains(s, `"bonsai hook guard || exit 2"`) || strings.Contains(s, `\"$b\"`) ||
		strings.Contains(s, "Monitor") || strings.Contains(s, "onFailure") {
		c.t.Fatalf("not as the build before left it:\n%s", s)
	}
}

// In an agent's session, init and update refuse --allow-exec, and unlink refuses itself, before anything is read:
// person-command, exit 4, nothing written, the next step a person's command in a terminal of their own. --allow-exec
// has one spelling. A person runs each.
func TestAgentRefusedPersonCommands(t *testing.T) {
	c := newCLI(t)
	c.link()
	c.commit("link")
	asAgent(c)
	before := tree(t, c.root)
	for _, args := range [][]string{{"update", "--allow-exec", "--yes"}, {"update", "--allow-exec"}, {"init", "--allow-exec", "--yes"},
		{"unlink", "--yes"}, {"unlink"}, {"update", "--yes", "--allow-exec", "--json"}, {"unlink", "--json"}} {
		code, out, errOut := c.run("", args...)
		if code != 4 || !strings.Contains(flat(out+errOut), "a person runs it in a terminal of their own, not one Claude Code started: bonsai ") ||
			!strings.Contains(flat(out+errOut), "CLAUDE_CODE_CHILD_SESSION is set") || !strings.Contains(flat(out+errOut), "status_writes agents") {
			t.Errorf("%v as an agent: exit %d\n%s%s", args, code, out, errOut)
		}
		if args[len(args)-1] == "--json" {
			e := errorIn(fits(t, out, "changes"), "changes")
			if n, _ := e.Get("next"); e.String("code") != "person-command" || n.(schema.Object).String("who") != "person" {
				t.Errorf("%v: error %s", args, schema.Show(e))
			}
		}
	}
	if tree(t, c.root) != before {
		t.Error("a refused command wrote")
	}
	// The flag in another spelling is refused by the parser, agent or person.
	for _, flag := range []string{"--allow-exec=true", "--allow-exec=", "--Allow-Exec", "-allow-exec", "--allow_exec"} {
		if code, _, errOut := c.run("", "update", flag, "--yes"); code != 2 || !strings.Contains(errOut, "update takes no") {
			t.Errorf("update %s: exit %d %s", flag, code, errOut)
		}
	}
	// What is no person's command still runs: update with nothing to change, and init without --allow-exec.
	if code, out, errOut := c.run("", "update", "--yes"); code != 0 || !strings.Contains(flat(out), "nothing to change") {
		t.Errorf("update --yes as an agent, nothing to change: exit %d\n%s%s", code, out, errOut)
	}
	// A person: written.
	asPerson(c)
	if code, out, errOut := c.run("", "unlink", "--yes"); code != 0 {
		t.Errorf("unlink --yes as a person: exit %d\n%s%s", code, out, errOut)
	}
}

// Recording another base branch is a person's command (the machine record's base_branch, workspace/record.go): with
// the main checkout switched to another branch, update in an agent's session is refused, exit 4, nothing written and
// the record's base unchanged; a person's update records the new base. The first record and the same branch again are
// free to an agent.
func TestAgentRecordingAnotherBase(t *testing.T) {
	c := newCLI(t)
	c.link()
	c.commit("link")
	home := os.Getenv(workspace.HomeEnv)
	base := func() string {
		t.Helper()
		rec, err := workspace.LoadMachineRecord(home, c.root)
		if err != nil || rec == nil {
			t.Fatalf("the record: %+v %v", rec, err)
		}
		return rec.BaseBranch
	}
	if got := base(); got != "main" {
		t.Fatalf("the base recorded at the link: %q", got)
	}
	asAgent(c)
	if code, out, errOut := c.run("", "update", "--yes"); code != 0 {
		t.Fatalf("the same base again, as an agent: exit %d\n%s%s", code, out, errOut)
	}
	testpack.Git(t, c.root, "checkout", "-q", "-b", "loose")
	before := tree(t, c.root)
	for _, args := range [][]string{{"update", "--yes"}, {"update", "--yes", "--json"}} {
		code, out, errOut := c.run("", args...)
		text := out + errOut
		if args[len(args)-1] == "--json" {
			e := errorIn(fits(t, out, "changes"), "changes")
			n, _ := e.Get("next")
			next, _ := n.(schema.Object)
			if e.String("code") != "person-command" || next.String("who") != "person" {
				t.Errorf("%v: error %s", args, schema.Show(e))
			}
			text = e.String("message") + "\n" + next.String("do")
		}
		if code != 4 || !strings.Contains(flat(text), `it would record the main checkout's branch "loose" as the workspace's base branch `+
			`in place of "main"`) || !strings.Contains(flat(text), "a person runs it in a terminal of their own, not one Claude Code "+
			"started: bonsai update --yes") {
			t.Errorf("%v as an agent: exit %d\n%s", args, code, text)
		}
	}
	if tree(t, c.root) != before || base() != "main" {
		t.Errorf("a refused update wrote: the base %q", base())
	}
	asPerson(c)
	if code, out, errOut := c.run("", "update", "--yes"); code != 0 || base() != "loose" {
		t.Errorf("a person's update: exit %d, the base %q\n%s%s", code, base(), out, errOut)
	}
}

// Recording settings for the checks looser than the ones this machine accepted is a person's command (the machine
// record's checks, workspace/checks.go): with a path taken off protected, update in an agent's session is refused,
// exit 4, nothing written and the record's settings unchanged; stricter settings are free to an agent; a person's
// update records looser ones, its text naming each loosening.
func TestAgentRecordingLooserChecks(t *testing.T) {
	c := newCLI(t)
	c.link()
	c.commit("link")
	home := os.Getenv(workspace.HomeEnv)
	accepted := func() string {
		t.Helper()
		ch, tr := workspace.AcceptedChecks(home, c.root)
		if !tr.Verified {
			t.Fatalf("no settings recorded at the link: %+v", tr)
		}
		return schema.Show(ch.JSON())
	}
	linked := c.read("bonsai.yaml")
	lists := `".claude/**", "bonsai.yaml", ".bonsai/lock.json"`
	if !strings.Contains(linked, "protected: ["+lists+"]") {
		t.Fatalf("the linked bonsai.yaml's protected list:\n%s", linked)
	}
	was := accepted()
	asAgent(c)
	// Stricter: free.
	c.write("bonsai.yaml", strings.Replace(linked, "protected: ["+lists+"]", "protected: ["+lists+`, "more/**"]`, 1))
	if code, out, errOut := c.run("", "update", "--yes"); code != 0 || accepted() == was || !strings.Contains(accepted(), `"more/**"`) {
		t.Fatalf("stricter settings, as an agent: exit %d\n%s%s", code, out, errOut)
	}
	stricter := accepted()
	// Looser than recorded: a person's command.
	c.write("bonsai.yaml", strings.Replace(linked, "protected: ["+lists+"]", `protected: ["bonsai.yaml", ".bonsai/lock.json"]`, 1))
	before := tree(t, c.root)
	for _, args := range [][]string{{"update", "--yes"}, {"update", "--yes", "--json"}} {
		code, out, errOut := c.run("", args...)
		text := out + errOut
		if args[len(args)-1] == "--json" {
			e := errorIn(fits(t, out, "changes"), "changes")
			n, _ := e.Get("next")
			next, _ := n.(schema.Object)
			if e.String("code") != "person-command" || next.String("who") != "person" {
				t.Errorf("%v: error %s", args, schema.Show(e))
			}
			text = e.String("message") + "\n" + next.String("do")
		}
		if code != 4 || !strings.Contains(flat(text), "it would record settings for the checks looser than the ones this machine last "+
			`accepted for the workspace: protected no longer holds ".claude/**"; protected no longer holds "more/**"`) ||
			!strings.Contains(flat(text), "a person runs it in a terminal of their own, not one Claude Code started: bonsai update --yes "+
				"(or put bonsai.yaml's settings back)") {
			t.Errorf("%v as an agent: exit %d\n%s", args, code, text)
		}
	}
	if after := tree(t, c.root); after != before || accepted() != stricter {
		t.Error("a refused update wrote the project or the record")
	}
	asPerson(c)
	code, out, errOut := c.run("", "update", "--yes")
	if code != 0 || accepted() == stricter || !strings.Contains(flat(out), "note this machine now accepts bonsai.yaml's settings for the "+
		`checks, looser than the ones it accepted before: protected no longer holds ".claude/**"`) {
		t.Errorf("a person's update: exit %d\n%s%s", code, out, errOut)
	}
}

// A project linked by the build before (Bonsai's lines by name, the guard's matcher without Monitor): check warns; it
// takes the new lines (Monitor and onFailure on the guard's) only with --allow-exec --yes: --yes alone is refused,
// exit 4, nothing written. In an agent's session the same update is a person's command whatever its flags: it takes
// out the guard's line (its matcher changed) and changes the stop gate's.
func TestNewLinesForAProjectLinkedBefore(t *testing.T) {
	c := newCLI(t)
	c.link()
	c.linkedByName()
	c.commit("linked by the build before")
	before := tree(t, c.root)

	if code, out, _ := c.run("", "check", "--json"); code > 1 || !strings.Contains(out, `"code": "own-hooks"`) ||
		!strings.Contains(out, "15 to add") || !strings.Contains(out, "15 to take out") || strings.Contains(out, `"code": "changed"`) {
		t.Errorf("check: exit %d\n%s", code, out)
	}

	// An agent: refused, the preview shown, the guard's and stop gate's lines named, the person's command the next step.
	asAgent(c)
	for _, args := range [][]string{{"update", "--yes"}, {"update"}, {"init", "--yes"}} {
		code, out, errOut := c.run("", args...)
		if code != 4 || !strings.HasPrefix(out, "bonsai "+args[0]+"  the preview\n") ||
			!strings.Contains(flat(out), "changes or takes out Bonsai's own guard or stop line (change Stop; remove PreToolUse ("+oldGuardMatcher+"))") ||
			nextOf(out) != "a person runs it in a terminal of their own, not one Claude Code started: bonsai "+args[0]+" --allow-exec --yes" {
			t.Errorf("%v as an agent: exit %d\n%s%s", args, code, out, errOut)
		}
	}
	code, out, _ := c.run("", "update", "--yes", "--json")
	doc := fits(t, out, "changes")
	if e := errorIn(doc, "changes"); code != 4 || doc.String("result") != "refused" || e.String("code") != "person-command" {
		t.Errorf("update --yes --json as an agent: exit %d\n%s", code, out)
	}
	if tree(t, c.root) != before {
		t.Fatal("an agent's refused update wrote")
	}

	// A person: --yes alone is the consent refusal; --allow-exec --yes writes every new line.
	asPerson(c)
	if code, out, _ := c.run("", "update", "--yes"); code != 4 || !strings.Contains(flat(out), "needs --allow-exec as well as --yes") ||
		!strings.Contains(flat(out), "add hook "+flat(engine.OwnHookLines()[0])+" (bonsai)") {
		t.Errorf("update --yes as a person: exit %d\n%s", code, out)
	}
	if tree(t, c.root) != before {
		t.Fatal("a refused update wrote")
	}
	if code, out, errOut := c.run("", "update", "--allow-exec", "--yes"); code != 0 {
		t.Fatalf("update --allow-exec --yes: exit %d\n%s%s", code, out, errOut)
	}
	s := c.read(".claude/settings.json")
	for _, cmd := range []string{engine.GuardCommand, engine.StopCommand, engine.StartCommand, engine.RecordCommand} {
		if !strings.Contains(s, jsonCmd(cmd)) {
			t.Errorf("not written: %s", cmd)
		}
	}
	if !strings.Contains(s, `"matcher": "`+engine.GuardMatcher+`"`) || !strings.Contains(s, `"onFailure": "block"`) ||
		strings.Contains(s, `"bonsai hook `) || strings.Contains(s, `"matcher": "`+oldGuardMatcher+`"`) {
		t.Errorf("after the update:\n%s", s)
	}
	if _, out, _ := c.run("", "check", "--json"); strings.Contains(out, `"code": "own-hooks"`) || strings.Contains(out, `"code": "changed"`) {
		t.Errorf("check after the update:\n%s", out)
	}
}

// An update that takes out a stop line of Bonsai's (an old one beside this build's, as if a build had written both):
// a person's --yes alone writes it (a removed line runs nothing); an agent's is refused.
func TestAgentUpdateTakingOutAStopLine(t *testing.T) {
	c := newCLI(t)
	c.link()
	// An old stop line beside the new one, in the same group, and the lock's fingerprint taking it in.
	v, err := schema.Decode([]byte(c.read(".claude/settings.json")))
	if err != nil {
		t.Fatal(err)
	}
	doc := v.(schema.Object)
	hooks, _ := doc.Get("hooks")
	stop, _ := hooks.(schema.Object).Get("Stop")
	g := stop.([]any)[0].(schema.Object)
	hs, _ := g.Get("hooks")
	g[g.Index("hooks")].Value = append(hs.([]any), schema.Object{{Key: "type", Value: "command"},
		{Key: "command", Value: "bonsai hook stop || exit 2"}})
	raw, err := schema.EncodeUTF8(doc)
	if err != nil {
		t.Fatal(err)
	}
	c.write(".claude/settings.json", string(raw))
	lockV, _ := schema.Decode([]byte(c.read(workspace.LockFile)))
	lock := lockV.(schema.Object)
	files, _ := lock.Get("files")
	e, _ := files.(schema.Object).Get(".claude/settings.json")
	eo := e.(schema.Object)
	eo[eo.Index("sha256")].Value = fingerprint(doc)
	raw, _ = schema.Encode(lock)
	c.write(workspace.LockFile, string(raw))
	c.commit("two stop lines")
	before := tree(t, c.root)

	asAgent(c)
	if code, out, _ := c.run("", "update", "--yes"); code != 4 || !strings.Contains(flat(out), "(remove Stop)") || !strings.Contains(flat(out), "person-command") &&
		!strings.Contains(flat(out), "a person runs it in a terminal of their own") {
		t.Errorf("an agent's update taking out a stop line: exit %d\n%s", code, out)
	}
	if tree(t, c.root) != before {
		t.Fatal("an agent's refused update wrote")
	}
	asPerson(c)
	if code, out, errOut := c.run("", "update", "--yes"); code != 0 || !strings.Contains(flat(out), "remove hook Stop: bonsai hook stop || exit 2") {
		t.Errorf("a person's update --yes: exit %d\n%s%s", code, out, errOut)
	}
	if strings.Contains(c.read(".claude/settings.json"), `"bonsai hook stop || exit 2"`) || !strings.Contains(c.read(".claude/settings.json"), jsonCmd(engine.StopCommand)) {
		t.Errorf("after the update:\n%s", c.read(".claude/settings.json"))
	}
}

// The second question's (ii): in a project in command mode whose running task's bonsai.allows holds bonsai.yaml and
// .bonsai/lock.json (a person's tap), an agent's update --allow-exec --yes and unlink --yes run; with the lock left out
// of the grant, refused, the reason naming it.
func TestAgentUnderAPersonsGrant(t *testing.T) {
	t.Setenv("BONSAI_TASK", "") // the caller's named task is not this test's: the running task grants
	t.Setenv("BONSAI_ROLE", "")
	c := newCLI(t)
	c.link()
	c.linkedByName()
	home, err := workspace.Home()
	if err != nil {
		t.Fatal(err)
	}
	md, err := workspace.MachineDir(home, c.root)
	if err != nil {
		t.Fatal(err)
	}
	c.write(filepath.ToSlash(mustRel(t, c.root, filepath.Join(md, workspace.MachineSettingsFile))), `{"status_writes": "command", "status_command": "studio-move"}`)
	c.write("work/tasks/T-0001-the-task.md", taskText("T-0001", "running", "bonsai.yaml"))
	c.commit("command mode")
	asAgent(c)
	before := tree(t, c.root)
	if code, out, errOut := c.run("", "update", "--allow-exec", "--yes"); code != 4 ||
		!strings.Contains(flat(errOut), "for .bonsai/lock.json: the running task T-0001 does not grant it") {
		t.Errorf("update --allow-exec with the lock not granted: exit %d\n%s%s", code, out, errOut)
	}
	if tree(t, c.root) != before {
		t.Fatal("a refused update wrote")
	}
	c.write("work/tasks/T-0001-the-task.md", taskText("T-0001", "running", "bonsai.yaml", ".bonsai/lock.json"))
	if code, out, errOut := c.run("", "update", "--allow-exec", "--yes"); code != 0 || !strings.Contains(c.read(".claude/settings.json"), jsonCmd(engine.GuardCommand)) {
		t.Errorf("update --allow-exec --yes under a person's grant: exit %d\n%s%s", code, out, errOut)
	}
	if code, out, errOut := c.run("", "unlink", "--yes"); code != 0 {
		t.Errorf("unlink --yes under a person's grant: exit %d\n%s%s", code, out, errOut)
	}
}

// mustRel gives target relative to base, forward slashes (cli.write takes a project-relative path; the machine
// settings file lies outside, so it is written through ../).
func mustRel(t *testing.T, base, target string) string {
	t.Helper()
	r, err := filepath.Rel(base, target)
	if err != nil {
		t.Fatal(err)
	}
	return filepath.ToSlash(r)
}

// taskText is a format-1 task file: its id, status and grants.
func taskText(id, status string, allows ...string) string {
	q := make([]string, len(allows))
	for i, a := range allows {
		q[i] = `"` + a + `"`
	}
	return "---\nformat: bonsai.task/1\nid: " + id + "\ntitle: A made-up task\nstatus: " + status + "\nlane: light" +
		"\ndone_when:\n  - \"It is done\"\ndepends_on: []\nblocked_by:\ncreated: 2026-10-11\nstarted:\nfinished:\nlabels:\n" +
		"  bonsai.allows: [" + strings.Join(q, ", ") + "]\n---\n\n# A made-up task\n"
}

// Through the shell, as an agent's tool runs it, with a real build: the flag in a variable, through xargs, in a
// script, and in a shell a shell started; each refused with the variable set (exit 4, nothing written), and written
// with it cleared. The hook shell runs them (sh, or Git Bash on Windows), never bash by name.
func TestAgentRefusedThroughTheShell(t *testing.T) {
	normal, _ := bonsaiBuilds(t)
	c := newCLI(t)
	c.link()
	c.linkedByName()
	c.commit("linked by the build before")
	b := shellPath(normal)
	c.write("link.sh", `"`+b+`" update --allow-exec --yes`+"\n")
	c.write("unlink.sh", `"`+b+`" unlink --yes`+"\n")
	c.commit("scripts")
	lines := []string{
		`F=--allow-exec; "` + b + `" update $F --yes`,
		`echo --allow-exec --yes | xargs "` + b + `" update`,
		`sh ./link.sh`,
		`sh -c 'sh ./unlink.sh'`,
	}
	before := tree(t, c.root)
	for i, line := range lines {
		want := 4
		if i == 1 {
			want = 123 // xargs's own exit when the command it ran ended with 1 to 125
		}
		code, out, errOut := runShell(t, c.root, line, "CLAUDE_CODE_CHILD_SESSION=1")
		if code != want || !strings.Contains(flat(out+errOut), "is a person's command") {
			t.Errorf("%s as an agent: exit %d\n%s%s", line, code, out, errOut)
		}
	}
	if tree(t, c.root) != before {
		t.Fatal("a refused command wrote")
	}
	if code, out, errOut := runShell(t, c.root, lines[0], testpack.PersonEnv()...); code != 0 ||
		!strings.Contains(c.read(".claude/settings.json"), jsonCmd(engine.GuardCommand)) {
		t.Errorf("%s as a person: exit %d\n%s%s", lines[0], code, out, errOut)
	}
	if code, out, errOut := runShell(t, c.root, lines[3], testpack.PersonEnv()...); code != 0 {
		t.Errorf("%s as a person: exit %d\n%s%s", lines[3], code, out, errOut)
	}
	if _, err := os.Stat(filepath.Join(c.root, "bonsai.yaml")); err == nil {
		t.Error("unlink as a person left bonsai.yaml")
	}
}

// runShell runs a command line in the hook shell in dir, with this process's environment and more.
func runShell(t *testing.T, dir, line string, env ...string) (int, string, string) {
	t.Helper()
	return runBin(t, hookShell(t), dir, env, "-c", line)
}
