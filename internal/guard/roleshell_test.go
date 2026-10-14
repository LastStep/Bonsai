package guard

import (
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/agents"
	"github.com/LastStep/Bonsai/internal/hookin"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// roleJudge is a judge for a project whose roles are read once from its lock (no git call), with the shell part of a
// caller of the role given: what the shell judge asks of the roles, outside the guard's run.
func roleJudge(t *testing.T, dir string, h agents.Header) *judge {
	lock, err := workspace.LoadLock(dir)
	if err != nil {
		t.Fatal(err)
	}
	roles, err := workspace.RolesInForce(lock, nil)
	if err != nil {
		t.Fatal(err)
	}
	j := newJudge(guardHome())
	j.session = j.project(dir)
	rr := &roleRun{j: j, in: &hookin.Payload{Session: sessA}, main: dir, store: agents.Store{Main: dir}, now: time.Now(),
		rs: &roleSet{roles: roles}}
	j.role, j.sh = rr, &shellRole{rr: rr, w: &who{key: "a1", h: h, counted: true}}
	return j
}

// The shell judge with the roles' part (a role that changes nothing, and a lead), over every line of the delete check's
// table and more: it never panics, answers the same twice, and says a refusal's reason and next step on one ASCII line
// each. (FuzzShellDecide holds the splitter and the delete check to random lines; the roles' part reads only what
// they give it.)
func TestRoleShellOnTheTable(t *testing.T) {
	type line struct {
		text string
		ps   bool
	}
	var lines []line
	for _, row := range delTable {
		lines = append(lines, line{row.line, row.tool == tP})
	}
	for _, s := range []string{"echo x > a", "cp a b", "claude -p x", "bonsai message --kind report --from a --to b --body 'x'",
		"sed -i s/a/b/ f", "tee a < b", "git commit -m x", "npx @anthropic-ai/claude-code", "dd of=x", "install -d x",
		"Set-Content -Path a -Value b", "cmd /c copy a b", "cmd /c copy", "cp", "cp -t", "ln a", "sed -i", "tee", "copy a", "dd",
		"> a", "2>&1 >", "git", "git -C", "npm exec", "node", "bonsai message --kind brief --to", "bonsai ladder --task",
		"{a,b} > {c,d}", "cp {a,} b", "echo > ''", "echo >", "cp -t '' a", "New-Item -Name x", "Copy-Item -Destination",
	} {
		lines = append(lines, line{s, false}, line{s, true})
	}
	dir := rolesProject(t, nil)
	quiet := roleJudge(t, dir, agents.Header{Role: "verifier", Kind: "lead", Writes: false})
	lead := roleJudge(t, dir, agents.Header{Role: "lead", Kind: "lead", Writes: true})
	for _, l := range lines {
		tool := tB
		if l.ps {
			tool = tP
		}
		for _, j := range []*judge{quiet, lead} {
			a := shellDecide(tool, l.text, dir, dir, j)
			if b := shellDecide(tool, l.text, dir, dir, j); a != b {
				t.Fatalf("%q: two answers: %+v and %+v", l.text, a, b)
			}
			if !a.Allow {
				for _, text := range []string{a.Why, a.Next} {
					if text == "" || oneLine(text) != text {
						t.Fatalf("%q: a refusal's reason and next step are one ASCII line each: %q / %q", l.text, a.Why, a.Next)
					}
				}
			}
		}
	}
}
