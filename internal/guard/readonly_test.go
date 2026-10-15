package guard

// T-5505's findings 1 and 2 (16 Oct, 5.5.5's real sessions): calls that write nothing, judged as writes. bonsai
// message --help was refused for its --from "", and a preview of bonsai update (no --yes) as an edit of bonsai.yaml and
// the lock with no running task; check's finding for a tracked .bonsai/local/ named git rm --cached, which the guard
// refused as a delete of a guarded path.

import (
	"strings"
	"testing"
)

// A --help of any command, and a preview of init or update, are not writes: let through to every caller, a role that
// changes nothing among them; the same words with --yes, --allow-exec or a variable that may stand for --yes are judged
// as before.
func TestReadOnlyBonsaiCallsAreNotWrites(t *testing.T) {
	dir := briefProject(t)
	started(t, dir, sessA, "workflow:lead")
	started(t, dir, sessA, "workflow:verifier")
	agent := ChildSessionEnv
	lead := func(cmd string) string { return shellAs(dir, "l1", "workflow:lead", cmd, false) }
	verifier := func(cmd string) string { return shellAs(dir, "v1", "workflow:verifier", cmd, false) }
	person := func(cmd string) string { return hookCall{session: sessA, tool: "Bash", input: bash(cmd)}.payload(dir) }
	// The lead has no address (no brief tied): its message --help was refused for its --from "".
	for _, c := range []string{"bonsai message --help", "bonsai message -h", "bonsai message --kind brief --to x --help",
		"bonsai update --help", "bonsai update --yes --allow-exec --help", "bonsai unlink --help", "bonsai init -h",
		"bonsai settings set x y --help", "bonsai labels attach f.yaml -h", "bonsai check --write --help"} {
		wantAnswer(t, dir, lead(c), 0, RuleBonsaiCommand, agent, "1")
		wantAnswer(t, dir, verifier(c), 0, RuleBonsaiCommand, agent, "1")
		wantAnswer(t, dir, person(c), 0, RuleBonsaiCommand, agent, "1")
	}
	// A preview: no task reads running, so an edit of bonsai.yaml and the lock is refused; a preview is no edit.
	for _, c := range []string{"bonsai update", "bonsai update --json", "bonsai update --diff --json", "bonsai init",
		"bonsai update --keep packs/x.md"} {
		wantAnswer(t, dir, person(c), 0, RuleBonsaiCommand, agent, "1")
		wantAnswer(t, dir, lead(c), 0, RuleBonsaiCommand, agent, "1")
		wantAnswer(t, dir, verifier(c), 0, RuleBonsaiCommand, agent, "1")
	}
	// Judged as before: --yes (in any spelling), --allow-exec, a word a variable stands for, xargs that may add --yes,
	// a preview where the hook's environment names no agent's session (the command could ask y/N at a terminal), the
	// other writing words, and a writing word for a role that changes nothing.
	for _, c := range []string{"bonsai update --yes", "bonsai update --json --yes", "bonsai update -yes", "bonsai update --YES",
		"bonsai update $FLAGS", "bonsai update --keep=$X", "echo --yes | xargs bonsai update", "bonsai update --allow-exec",
		"bonsai unlink", "bonsai settings set x y", "bonsai --json update --yes"} {
		if code, s := runs(t, dir, person(c), agent, "1"); code != 2 {
			t.Errorf("%s: exit %d, want a refusal: %s", c, code, s)
		}
	}
	if code, s := runs(t, dir, person("bonsai update")); code != 2 || !strings.Contains(s, "judged as an edit of them") {
		t.Errorf("a preview with no agent's session named: exit %d: %s", code, s)
	}
	wantAnswer(t, dir, verifier("bonsai check --write"), 2, RuleRoleWritesNothing, agent, "1")
	wantAnswer(t, dir, lead("bonsai message --kind report --from lead --to orchestrator --body x"), 2, RuleCallRefused, agent, "1")
}
