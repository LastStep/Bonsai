package guard

// command mode's tripwires, each in both modes ("5.3 done", check 5): in command mode refused, the reason first and
// the workspace's status_command as the next step; in agents mode the same edit is the agent's own, allowed.

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/testpack"
	"github.com/LastStep/Bonsai/internal/workspace"
)

func TestTripwiresOnlyInCommandMode(t *testing.T) {
	task := "work/tasks/T-0001-the-task.md"
	labels := func(t *testing.T, dir string) {
		home, _ := workspace.Home()
		md, _ := workspace.MachineDir(home, dir)
		writeFile(t, filepath.Join(md, workspace.MachineLabelsDir, "acme.yaml"), "format: bonsai.labels/1\nnamespace: acme\n"+
			"version: 1\nlabels:\n  - name: acme.cost\n    kind: number\n    values: []\n    items:\n    pattern:\n    max:\n"+
			"    kinds: [task]\n    set_by: outside\n    grants: false\n    description: \"Dollars this work cost.\"\n")
	}
	cases := []struct {
		rule  string
		setup func(t *testing.T, dir string)
		in    func(dir string) string
	}{
		{RuleStatusLine, nil, func(dir string) string { return editIn(dir, task, "status: todo", "status: running") }},
		{RuleLaneLine, nil, func(dir string) string { return editIn(dir, task, "lane: light", "lane: none") }},
		{RuleGrantPersonOnly, nil, func(dir string) string { return editIn(dir, task, `["assets/**"]`, `["assets/**", "keys/**"]`) }},
		{RuleOutsideLabel, labels, func(dir string) string { return editIn(dir, task, "labels:\n", "labels:\n  acme.cost: 12\n") }},
		{RuleBadFormat, nil, func(dir string) string {
			return writeOf(dir, task, "---\nformat: bonsai.task/1\nid: T-0001\nstatus: [todo]\n---\n")
		}},
	}
	for _, c := range cases {
		t.Run(c.rule, func(t *testing.T) {
			for _, mode := range []string{"agents", "command"} {
				dir := withTask(t, "todo", "assets/**")
				if c.setup != nil {
					c.setup(t, dir)
				}
				want := RuleNotProtected
				if mode == "command" {
					commandMode(t, dir)
					want = c.rule
				}
				s := wantRule(t, dir, c.in(dir), want)
				if mode == "command" && (!strings.Contains(s, "refused: in command mode (this machine's status_writes), ") ||
					!strings.Contains(s, `"studio-move"`)) {
					t.Errorf("%s in command mode: the reason first and the status command next: %s", c.rule, s)
				}
			}
		})
	}
	// A task file outside the main checkout: a worktree's copy, refused in command mode only.
	t.Run(RuleTaskOutsideMain, func(t *testing.T) {
		for _, mode := range []string{"agents", "command"} {
			tmp := t.TempDir()
			home := testpack.Isolate(t, tmp)
			main, wt := gitProject(t, home, grantYAML, time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC))
			want := 0
			if mode == "command" {
				commandMode(t, main)
				want = 2
			}
			code, s := answer(t, wt, editIn(wt, "work/tasks/T-0005-x.md", "", taskText("T-0005", "todo", "light")))
			if code != want {
				t.Errorf("%s mode: exit %d, %s", mode, code, s)
			}
		}
	})
}

// A pack's declared protected path is guarded, and a running task's grant opens it ("5.3 done", check 4).
func TestAPacksDeclaredPathIsGrantable(t *testing.T) {
	yaml := strings.Replace(grantYAML, "packs: []", "packs:\n  - id: base\n    source: \"https://example.com/base.git\"\n    ref: v1.0.0", 1)
	for _, c := range []struct {
		allows []string
		rule   string
	}{{nil, RuleProtected}, {[]string{"work/protocols/**"}, RuleGranted}} {
		dir := grantProject(t, yaml, map[string]string{
			".bonsai/lock.json":               lockWith(`{"protected":["work/protocols/**"]}`, `{}`),
			"work/tasks/T-0001-the-task.md":   taskText("T-0001", "running", "light", c.allows...),
			"work/protocols/session-start.md": "# Start\n",
		})
		s := wantRule(t, dir, editIn(dir, "work/protocols/session-start.md", "# Start", "# Begin"), c.rule)
		if c.rule == RuleProtected && !strings.Contains(s, `the protected paths the pack "base" declares ("work/protocols/**")`) {
			t.Errorf("the refusal names the pack: %s", s)
		}
	}
}
