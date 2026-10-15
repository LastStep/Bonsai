package guard

// T-5505's findings 1 and 2 (16 Oct, 5.5.5's real sessions): calls that write nothing, judged as writes. bonsai
// message --help was refused for its --from "", and a preview of bonsai update (no --yes) as an edit of bonsai.yaml and
// the lock with no running task; check's finding for a tracked .bonsai/local/ named git rm --cached, which the guard
// refused as a delete of a guarded path.

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/engine"
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
	// The sweep: git's own help of a write word, --help first after the subcommand (git turns it into git help <sub>)
	// or -h alone, is no write for a role that changes nothing, and git clean --help no delete; --help elsewhere may be
	// a value or a path, and is judged as before.
	for _, c := range []string{"git commit --help", "git clean --help", "git stash -h", "git rm --help", "git -C . checkout --help",
		"git mv --help -f"} {
		if code, s := runs(t, dir, verifier(c), agent, "1"); code != 0 {
			t.Errorf("a verifier, %s: exit %d: %s", c, code, s)
		}
		if code, s := runs(t, dir, person(c), agent, "1"); code != 0 {
			t.Errorf("the person's session, %s: exit %d: %s", c, code, s)
		}
	}
	for _, c := range []string{"git commit -m --help", "git rm -- --help", "git commit -h -a", "git commit $H", "git commit '--he'*"} {
		wantAnswer(t, dir, verifier(c), 2, RuleRoleWritesNothing, agent, "1")
	}
	wantAnswer(t, dir, person("git clean -f"), 2, RuleDeleteUnnamed, agent, "1")
	wantAnswer(t, dir, lead("bonsai message --kind report --from lead --to orchestrator --body x"), 2, RuleCallRefused, agent, "1")
}

// check's finding for a tracked .bonsai/local/ (who: agent) names git rm -r -q --cached -- .bonsai/local: an index-only
// untracking of .bonsai/local/ is let through; a delete of it, an untracking of any other guarded path (a deletion in
// git once committed), and one the check cannot read plainly are refused as before.
func TestUntrackingBonsaiLocal(t *testing.T) {
	dir := briefProject(t)
	writeFile(t, filepath.Join(dir, "sub", "x.txt"), "x\n")
	started(t, dir, sessA, "workflow:lead")
	lead := func(cmd string) string { return shellAs(dir, "l1", "workflow:lead", cmd, false) }
	person := func(cmd string) string { return hookCall{session: sessA, tool: "Bash", input: bash(cmd)}.payload(dir) }
	for _, c := range []string{engine.UntrackLocal, engine.UntrackLocal + "; git commit -m untrack",
		"git rm --cached .bonsai/local/log/w-2026-10-15.ndjson", "git rm -r --cached .bonsai/local/", "cd .bonsai && git rm -r --cached local",
		"git rm -r --cached ./.bonsai/local", "cd sub && git rm -r --cached ../.bonsai/local", "git -C sub rm -r --cached ../.bonsai/local",
		"git rm -r --cached " + filepath.ToSlash(dir) + "/.bonsai/local"} {
		if code, s := runs(t, dir, person(c)); code != 0 {
			t.Errorf("the person's session, %s: exit %d: %s", c, code, s)
		}
		if code, s := runs(t, dir, lead(c)); code != 0 {
			t.Errorf("a lead, %s: exit %d: %s", c, code, s)
		}
	}
	for _, c := range []string{"git rm -r .bonsai/local", "git rm -r -q -- .bonsai/local", "git rm -r --cached .bonsai",
		"git rm --cached bonsai.yaml", "git rm --cached .bonsai/lock.json", "git rm --cached .bonsai/local/../lock.json",
		"git rm --cached .bonsai/local protected.txt", "git rm -r --cached '.bonsai/loc*'", "git rm -r --cached $L",
		"git rm -r --cached :/.bonsai/lock.json", "git rm -r --cached --dry-run .bonsai; rm -r .bonsai/local",
		"rm -r .bonsai/local", "git rm -r --cached .bonsai/local && rm -rf .bonsai/local"} {
		if code, s := runs(t, dir, person(c)); code != 2 {
			t.Errorf("the person's session, %s: exit %d, want a refusal: %s", c, code, s)
		}
	}
}
