package guard

// bonsai pack's second-stage forms in a shell call (commands.go; T-6011, the brief's lesson from stage 1: the guard
// judges a pack subcommand only as written): pack show is a read for every role, as outdated is, outdated --changes
// too; pack set --all and pack set --downgrade are pack set, so with --yes (written, fed by xargs, or a word that may
// stand for it) they are judged as an edit of bonsai.yaml and the lock, which a grant opens; their previews pass.

import (
	"strings"
	"testing"
)

// pack show and outdated --changes pass for every role, one that changes nothing among them, and for an agent with
// no grant; pack set --all and --downgrade previews pass too; with --yes they are writes a role that changes nothing
// may not make.
func TestPackShowAndSetAllForEveryRole(t *testing.T) {
	dir := briefProject(t)
	started(t, dir, sessA, "workflow:lead")
	started(t, dir, sessA, "workflow:verifier")
	agent := ChildSessionEnv
	lead := func(cmd string) string { return shellAs(dir, "l1", "workflow:lead", cmd, false) }
	verifier := func(cmd string) string { return shellAs(dir, "v1", "workflow:verifier", cmd, false) }
	person := func(cmd string) string { return hookCall{session: sessA, tool: "Bash", input: bash(cmd)}.payload(dir) }
	for _, c := range []string{"bonsai pack show base", "bonsai pack show base --json", "bonsai pack outdated --changes",
		"bonsai pack set --all", "bonsai pack set --all --json", "bonsai pack set base@v1.0.0 --downgrade"} {
		wantAnswer(t, dir, person(c), 0, RuleBonsaiCommand, agent, "1")
		wantAnswer(t, dir, lead(c), 0, RuleBonsaiCommand, agent, "1")
		wantAnswer(t, dir, verifier(c), 0, RuleBonsaiCommand, agent, "1")
	}
	for _, c := range []string{"bonsai pack set --all --yes", "bonsai pack set base@v1.0.0 --downgrade --yes", "echo --yes | xargs bonsai pack set --all",
		"bonsai pack set --all $FLAGS", "bonsai pack --all set --yes"} {
		if code, s := runs(t, dir, person(c), agent, "1"); code != 2 {
			t.Errorf("%s: exit %d, want a refusal: %s", c, code, s)
		}
	}
	wantAnswer(t, dir, verifier("bonsai pack set --all --yes"), 2, RuleRoleWritesNothing, agent, "1")
	// show is not open to xargs: its own words settle it.
	if why := ownOpen([]string{"bonsai", "pack", "show", "base"}); why != "" {
		t.Errorf("pack show is open: %s", why)
	}
}

// pack set --all --yes and --downgrade --yes with and without a grant of bonsai.yaml and the lock, in both modes:
// allowed under the grant, refused as an edit of the two without it, as pack set --yes is.
func TestPackSetAllAndTheGrant(t *testing.T) {
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
			for _, argv := range [][]string{{"bonsai", "pack", "set", "--all", "--yes"}, {"bonsai", "pack", "--yes", "set", "--all"},
				{"bonsai", "pack", "set", "base@v1.0.0", "--downgrade", "--yes"}} {
				d := j.ownCommand(dir, argv, true)
				switch {
				case granted && (!d.Allow || d.Rule != RuleBonsaiCommand):
					t.Errorf("%s, granted: %q: %+v, want allowed", mode, argv, d)
				case !granted && (d.Allow || !strings.Contains(d.Why, "bonsai pack set writes bonsai.yaml and .bonsai/lock.json, so it is judged as an edit of them")):
					t.Errorf("%s, no grant: %q: %+v, want refused as an edit of bonsai.yaml and the lock", mode, argv, d)
				}
			}
			if d := j.ownCommand(dir, []string{"bonsai", "pack", "show", "base"}, true); !d.Allow {
				t.Errorf("%s, granted %v: pack show: %+v", mode, granted, d)
			}
		}
	}
}
