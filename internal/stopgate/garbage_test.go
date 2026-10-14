package stopgate

// The stop gate meets a named task file whose frontmatter holds nothing (a file holding only "garbage: [") as one that
// does not parse (workspace's readTask; contract §13): it blocks, though a green result for the task proves HEAD. Before
// the fix, such a file read as a task with no status, and the gate let the session end (STATE's "Placed for later",
// the 5.3 end verifier's finding).

import "testing"

func TestTheGateBlocksATaskFileWithNothingInItsFrontmatter(t *testing.T) {
	for _, c := range []struct{ name, junk string }{{"garbage alone", "garbage: [\n"}, {"an empty file", ""},
		{"an empty block", "---\n\n---\n"}} {
		t.Run(c.name, func(t *testing.T) {
			root := project(t)
			write(t, root, "work/tasks/T-0901-gated.md", c.junk)
			code, stderr := gateFor(t, root, "T-0901")
			blocks(t, code, stderr, root, RuleTaskBroken, "the named task's file does not parse (T-0901 in work/tasks/T-0901-gated.md)",
				"run: "+installed+" ladder --task T-0901")
		})
	}
	t.Run("beside the named task, another such file changes nothing", func(t *testing.T) {
		root := project(t)
		write(t, root, "work/tasks/T-0902-junk.md", "garbage: [\n")
		code, stderr := gateFor(t, root, "T-0901")
		allows(t, code, stderr, root)
	})
}
