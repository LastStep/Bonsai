package guard

// V1's fix round (5.5.5e): git rm's --no-cached undoing --cached, and a command string a runner other than sh -c and
// bash -c runs.

import (
	"path/filepath"
	"testing"
)

// git rm's last --cached or --no-cached wins, so --no-cached (or a prefix git takes for it) after --cached deletes
// the files from disk: only engine.UntrackLocal's own flags (-r, -q, --cached and a prefix of it, --) let an
// untracking of .bonsai/local/ through; anything else is judged as a delete, as before (V1's TestV1Adversarial).
func TestUntrackingBonsaiLocalTakesOnlyItsOwnFlags(t *testing.T) {
	dir := briefProject(t)
	writeFile(t, filepath.Join(dir, "sub", "x.txt"), "x\n")
	started(t, dir, sessA, "workflow:lead")
	lead := func(cmd string) string { return shellAs(dir, "l1", "workflow:lead", cmd, false) }
	person := func(cmd string) string { return hookCall{session: sessA, tool: "Bash", input: bash(cmd)}.payload(dir) }
	for _, c := range []string{
		"git rm -r --cached --no-cached .bonsai/local",
		"git rm -r --no-cached --cached .bonsai/local",
		"git rm -r .bonsai/local --cached --no-cached",
		"git rm -r --cached --no-cache .bonsai/local",
		"git rm -r --cached --no-c .bonsai/local",
		"git rm -r --cached --no-ca .bonsai/local/agents",
		"git rm -r -q --cached --no-cached -- .bonsai/local",
		"git rm -r -q --cached --no-cached -- .bonsai/local/agents",
		"git add -f .bonsai/local && git rm -r -q --cached --no-cached -- .bonsai/local",
		"bonsai update --yes --help; git rm -r --cached --no-cached .bonsai/local",
		"git rm -r --cached -f .bonsai/local",
		"git rm -r --cached --ignore-unmatch .bonsai/local",
		"git rm -r --cached --cached=x .bonsai/local",
		"git rm -r --cached $NC .bonsai/local",
		"git rm -r --cached \"$NC\" -- .bonsai/local",
	} {
		if code, s := runs(t, dir, person(c), ChildSessionEnv, "1"); code != 2 {
			t.Errorf("the person's session, %s: exit %d, want a refusal: %s", c, code, s)
		}
		if code, s := runs(t, dir, lead(c), ChildSessionEnv, "1"); code != 2 {
			t.Errorf("a lead, %s: exit %d, want a refusal: %s", c, code, s)
		}
	}
	for _, c := range []string{"git rm -r -q --cached -- .bonsai/local", "git rm -rq --cached .bonsai/local",
		"git rm -r --cache .bonsai/local", "git rm --cached -r -q .bonsai/local/agents", "git -C .bonsai rm -r --cached local"} {
		if code, s := runs(t, dir, person(c), ChildSessionEnv, "1"); code != 0 {
			t.Errorf("the person's session, %s: exit %d: %s", c, code, s)
		}
		if code, s := runs(t, dir, lead(c), ChildSessionEnv, "1"); code != 0 {
			t.Errorf("a lead, %s: exit %d: %s", c, code, s)
		}
	}
}
