package guard

// bonsai pack in a shell call (commands.go; T-6009's plan, step 7, and its proof rows "Loosening is the person's",
// "Previews pass for every role" and "A fed --yes is judged"): pack outdated and any pack line with no --yes pass for
// every role, a role that changes nothing among them; pack set and pack add with --yes are judged as an edit of
// bonsai.yaml and the lock, so a grant opens them; pack remove with --yes is a person's command no grant opens; a fed
// --yes (xargs) and a line the check cannot split are judged as their written form.

import (
	"strings"
	"testing"
)

// agentJudge is judgeIn in an agent's session (CLAUDE_CODE_CHILD_SESSION set).
func agentJudge(dir string) *judge {
	j := judgeIn(dir)
	j.getenv = func(k string) string {
		if k == ChildSessionEnv {
			return "1"
		}
		return ""
	}
	return j
}

func TestPackPreviewsPassForEveryRole(t *testing.T) {
	dir := briefProject(t)
	started(t, dir, sessA, "workflow:lead")
	started(t, dir, sessA, "workflow:verifier")
	agent := ChildSessionEnv
	lead := func(cmd string) string { return shellAs(dir, "l1", "workflow:lead", cmd, false) }
	verifier := func(cmd string) string { return shellAs(dir, "v1", "workflow:verifier", cmd, false) }
	person := func(cmd string) string { return hookCall{session: sessA, tool: "Bash", input: bash(cmd)}.payload(dir) }
	// No task reads running here, so an edit of bonsai.yaml and the lock is refused to everyone: a preview is no edit.
	for _, c := range []string{"bonsai pack outdated", "bonsai pack outdated --json", "bonsai pack set base@base-v1.0.1",
		"bonsai pack set base@base-v1.0.1 --json", "bonsai pack add --source https://example.org/x.git --ref v1.0.0",
		"bonsai pack remove base", "bonsai pack remove base --json", "bonsai pack set base@x --diff --explain"} {
		wantAnswer(t, dir, person(c), 0, RuleBonsaiCommand, agent, "1")
		wantAnswer(t, dir, lead(c), 0, RuleBonsaiCommand, agent, "1")
		wantAnswer(t, dir, verifier(c), 0, RuleBonsaiCommand, agent, "1")
	}
	// With --yes, a variable that may stand for it, or xargs that may add it: judged as writes.
	for _, c := range []string{"bonsai pack set base@x --yes", "bonsai pack set base@x -yes", "bonsai pack add --source s --ref r --yes",
		"bonsai pack set base@x $FLAGS", "echo --yes | xargs bonsai pack set base@x", "echo --yes | xargs bonsai pack add --source s --ref r",
		"bonsai pack remove base --yes", "echo --yes | xargs bonsai pack remove base", "echo set base@x --yes | xargs bonsai pack",
		"bonsai pack outdated --allow-exec", "bonsai pack set base@x --allow-exec"} {
		if code, s := runs(t, dir, person(c), agent, "1"); code != 2 {
			t.Errorf("%s: exit %d, want a refusal: %s", c, code, s)
		}
	}
	// A role that changes nothing: a pack write is refused as such, its preview and outdated passed above.
	wantAnswer(t, dir, verifier("bonsai pack set base@x --yes"), 2, RuleRoleWritesNothing, agent, "1")
	wantAnswer(t, dir, verifier("bonsai pack add --source s --ref r --yes"), 2, RuleRoleWritesNothing, agent, "1")
}

// "with and without a grant of bonsai.yaml and the lock": set and add open under a grant, remove never does.
func TestPackWritesAndTheGrant(t *testing.T) {
	for _, granted := range []bool{false, true} {
		dir := withTask(t, "running")
		if granted {
			dir = withTask(t, "running", "bonsai.yaml", ".bonsai/lock.json")
		}
		for _, mode := range []string{"agents", "command"} {
			if mode == "command" {
				commandMode(t, dir)
			}
			j := agentJudge(dir)
			for _, argv := range [][]string{{"bonsai", "pack", "set", "base@base-v1.0.1", "--yes"},
				{"bonsai", "pack", "add", "--source", "https://example.org/x.git", "--ref", "v1.0.0", "--yes"},
				{"bonsai", "--json", "pack", "--yes", "set", "base@x"}} {
				d := j.ownCommand(dir, argv, true)
				switch {
				case granted && (!d.Allow || d.Rule != RuleBonsaiCommand):
					t.Errorf("%s, granted: %q: %+v, want allowed", mode, argv, d)
				case !granted && (d.Allow || d.Rule != RulePersonOnly || !strings.Contains(d.Why, "bonsai pack ")):
					t.Errorf("%s, no grant: %q: %+v, want refused as an edit of bonsai.yaml and the lock", mode, argv, d)
				case !granted && !strings.Contains(d.Why, "writes bonsai.yaml and .bonsai/lock.json, so it is judged as an edit of them"):
					t.Errorf("%s, no grant: %q: %s", mode, argv, d.Why)
				}
			}
			for _, argv := range [][]string{{"bonsai", "pack", "remove", "base", "--yes"}, {"bonsai", "pack", "--yes", "remove", "base"},
				{"/usr/local/bin/bonsai", "pack", "remove", "base", "--YES"}} {
				if d := j.ownCommand(dir, argv, true); d.Allow || d.Rule != RulePersonCommand || !strings.Contains(d.Why, "no task's grant opens") ||
					!strings.HasPrefix(d.Next, "a person runs this command") {
					t.Errorf("%s, granted %v: %q: %+v, want person-command", mode, granted, argv, d)
				}
			}
			// pack remove whose words a variable may stand for: judged as with --yes.
			if d := j.ownCommand(dir, []string{"bonsai", "pack", "remove", "base"}, false); d.Rule != RulePersonCommand {
				t.Errorf("%s: pack remove, not literal: %+v", mode, d)
			}
			// The preview and outdated, in an agent's session: allowed whatever the grant.
			for _, argv := range [][]string{{"bonsai", "pack", "remove", "base"}, {"bonsai", "pack", "set", "base@x"}, {"bonsai", "pack", "outdated"}} {
				if d := j.ownCommand(dir, argv, true); !d.Allow {
					t.Errorf("%s, granted %v: %q: %+v, want allowed", mode, granted, argv, d)
				}
			}
		}
	}
}

// A fed --yes: echo --yes | xargs bonsai pack set base@x is judged as the written form, refused for an agent with no
// grant as update is; a line the check cannot split that holds pack remove is refused outright.
func TestPackFedYesIsJudged(t *testing.T) {
	dir := withTask(t, "running")
	for _, line := range []string{"echo --yes | xargs bonsai pack set base@x", "echo --yes | xargs bonsai update",
		"echo set base@x --yes | xargs bonsai pack", "echo --yes | xargs bonsai pack add --source s --ref r"} {
		code, s := answer(t, dir, bashIn(dir, "Bash", line), ChildSessionEnv, "1")
		if code != 2 || !strings.Contains(s, "rule: person-command.") && !strings.Contains(s, "rule: person-only.") {
			t.Errorf("%s: exit %d: %s", line, code, s)
		}
	}
	// Not open: outdated's own words settle it.
	if why := ownOpen([]string{"bonsai", "pack", "outdated"}); why != "" {
		t.Errorf("pack outdated is open: %s", why)
	}
	for _, argv := range [][]string{{"bonsai", "pack"}, {"bonsai", "pack", "set", "x@y"}, {"bonsai", "pack", "add"}, {"bonsai", "pack", "remove", "x"}} {
		if ownOpen(argv) == "" {
			t.Errorf("%q is not open to xargs", argv)
		}
	}
	if k := delHoldsBonsai(`bonsai pack remove base --yes "`); k != "person-only" {
		t.Errorf("an unsplittable pack remove: %q", k)
	}
	wantRule(t, dir, bashIn(dir, "Bash", `bonsai pack remove base --yes "`), RulePersonCommand, ChildSessionEnv, "1")
}

// S1 (the verifier of 29 Oct): a pack call whose subcommand, or a word before it, is not literal (a variable, a
// substitution, a glob) could be pack remove: refused person-command with --yes on the line or a word that could
// stand for it, with and without a grant of bonsai.yaml and the lock; the literal forms are judged as before.
func TestPackSubcommandNotLiteral(t *testing.T) {
	lines := []string{"bonsai pack $S base@x --yes", "S=set; bonsai pack $S base@x --yes", "bonsai pack se* base@x --yes",
		"bonsai pack $(echo remove) base --yes", "bonsai pack $S base@x", "bonsai pack `echo remove` base"}
	for _, granted := range []bool{false, true} {
		dir := withTask(t, "running")
		if granted {
			dir = withTask(t, "running", "bonsai.yaml", ".bonsai/lock.json")
		}
		for _, line := range lines {
			code, s := answer(t, dir, bashIn(dir, "Bash", line), ChildSessionEnv, "1")
			if code != 2 || !strings.Contains(s, "rule: person-command.") {
				t.Errorf("granted %v: %s: exit %d: %s", granted, line, code, s)
			}
		}
		// The literal forms, unchanged: set with --yes judged as an edit (allowed under the grant), remove --yes a
		// person's command, previews free.
		code, s := answer(t, dir, bashIn(dir, "Bash", "bonsai pack set base@x --yes"), ChildSessionEnv, "1")
		if granted != (code == 0) {
			t.Errorf("granted %v: a literal set --yes: exit %d: %s", granted, code, s)
		}
		wantRule(t, dir, bashIn(dir, "Bash", "bonsai pack remove base --yes"), RulePersonCommand, ChildSessionEnv, "1")
		wantRule(t, dir, bashIn(dir, "Bash", "bonsai pack set base@x"), RuleBonsaiCommand, ChildSessionEnv, "1")
		wantRule(t, dir, bashIn(dir, "Bash", "bonsai pack outdated"), RuleBonsaiCommand, ChildSessionEnv, "1")
	}
	// A role that changes nothing: refused as such.
	dir := briefProject(t)
	started(t, dir, sessA, "workflow:verifier")
	wantAnswer(t, dir, shellAs(dir, "v1", "workflow:verifier", "bonsai pack $S base@x --yes", false), 2, RuleRoleWritesNothing, ChildSessionEnv, "1")
	// xargs: a glob subcommand is open to its input.
	if ownOpen([]string{"bonsai", "pack", "se*", "base@x"}) == "" {
		t.Error("a glob subcommand is not open to xargs")
	}
}
