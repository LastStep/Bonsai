package guard

// V1's fix round (5.5.5e): git rm's --no-cached undoing --cached, and a command string a runner other than sh -c and
// bash -c runs.

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/agents"
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

// A command string run by script, su, runuser, sg or flock, in their util-linux and BSD spellings, and a program
// run by flock, runuser -u and the like, is read as sh -c's is: a bonsai call, a change of CLAUDE_CODE_CHILD_SESSION
// and a delete in it are judged (V1's M2: script -qec 'bonsai update --yes' /dev/null ran for every role).
func TestCommandStringRunnersAreRead(t *testing.T) {
	dir := briefProject(t)
	started(t, dir, sessA, "workflow:lead")
	person := func(cmd string) string { return hookCall{session: sessA, tool: "Bash", input: bash(cmd)}.payload(dir) }
	verifier := func(cmd string) string { return shellAs(dir, "v1", "workflow:verifier", cmd, false) }
	started(t, dir, sessA, "workflow:verifier")
	for _, inner := range []string{"bonsai update --yes", "env -u CLAUDE_CODE_CHILD_SESSION bonsai update",
		"export CLAUDE_CODE_CHILD_SESSION=", "rm -rf .bonsai"} {
		q := "'" + inner + "'"
		for _, c := range []string{
			"script -c " + q, "script -qec " + q + " /dev/null", "script -qc " + q + " /dev/null",
			"script -q -e -c " + q + " /dev/null", "script --command=" + q + " /dev/null", "script --comm " + q + " /dev/null",
			"script -a -t -q /dev/null -c " + q, "script -c" + q + " /dev/null", "script -q /dev/null " + inner,
			"script -q -t 0 /dev/null " + inner, "echo " + q + " | script -q /dev/null", "script -q <<< " + q,
			"su -c " + q, "su -lc " + q, "su - root -c " + q, "su root -- -c " + q, "su --session-command=" + q + " root",
			"su -s /bin/sh -c " + q + " root", "echo " + q + " | su root",
			"runuser -c " + q + " root", "runuser -lc " + q + " root", "runuser -u root -- " + inner, "runuser -u root " + inner,
			"flock /tmp/l -c " + q, "flock -x /tmp/l --command " + q, "flock /tmp/l " + inner, "flock -n -w 3 /tmp/l " + inner,
			"sg wheel -c " + q, "sg wheel " + q, "sg - wheel " + q,
			"timeout 30 sh -c " + q, "nice -n 5 bash -c " + q, "stdbuf -oL sh -c " + q, "nohup sh -c " + q,
			"setsid -f sh -c " + q, "xargs sh -c " + q, "watch -n 1 " + q,
			"chrt 1 " + inner, "chrt -f 10 " + inner, "taskset 1 " + inner, "taskset -c 0 " + inner,
			"pkexec " + inner, "pkexec --user root " + inner, "unshare -r " + inner, "unshare -m -w . " + inner,
			"chroot / " + inner, "dbus-run-session -- " + inner,
		} {
			if code, s := runs(t, dir, person(c), ChildSessionEnv, "1"); code != 2 {
				t.Errorf("the person's session, %s: exit %d, want a refusal: %s", c, code, s)
			}
			if code, s := runs(t, dir, verifier(c), ChildSessionEnv, "1"); code != 2 {
				t.Errorf("a verifier, %s: exit %d, want a refusal: %s", c, code, s)
			}
		}
	}
	for _, c := range []string{"script -qec 'bonsai status' /dev/null", "script -q /dev/null", "script -qc 'echo hi' /dev/null",
		"su -c 'ls'", "flock /tmp/l echo hi", "flock /tmp/l -c 'bonsai check'", "runuser -u root -- bonsai status",
		"sg wheel 'ls -l'", "taskset -c 0 ls", "chrt 1 bonsai status"} {
		if code, s := runs(t, dir, person(c), ChildSessionEnv, "1"); code != 0 {
			t.Errorf("the person's session, %s: exit %d: %s", c, code, s)
		}
	}
}

// A minutes budget is spent once the time since its window started passes the limit, not a whole minute after it
// (V1's should-fix 1: with minutes: 1, calls passed until 2:00); the refusal counts the minutes begun.
func TestAMinutesBudgetIsSpentOnceTheLimitPasses(t *testing.T) {
	dir := briefProject(t)
	t0 := time.Date(2026, 10, 15, 3, 0, 0, 0, time.UTC)
	startedAt(t, dir, sessA, "workflow:lead", t0)
	lead := func(cmd string) string {
		return hookCall{session: sessA, agentID: "l1", agentType: "workflow:lead", tool: "Bash", input: bash(cmd)}.payload(dir)
	}
	for _, at := range []time.Duration{0, 180 * time.Minute} {
		if code, s := runAt(t, dir, lead("git status"), t0.Add(at), false); code != 0 {
			t.Fatalf("at %v: exit %d: %s", at, code, s)
		}
	}
	code, s := runAt(t, dir, lead("git status"), t0.Add(180*time.Minute+time.Second), false)
	if code != 2 || !strings.Contains(s, "budget spent (minutes: 181 of 180)") {
		t.Fatalf("a second past the limit: exit %d: %s", code, s)
	}
	// The call line keeps the whole minutes elapsed.
	if _, m, ok := (agents.Store{Main: dir}).LastCall(sessA, "l1"); !ok || m != 180 {
		t.Errorf("the last call line's minutes %d %v", m, ok)
	}
}
