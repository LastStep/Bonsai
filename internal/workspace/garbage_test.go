package workspace

// A task file whose frontmatter holds nothing does not parse as a task (active.go's ReadTask; contract §13: "a task
// file that does not parse makes the answer none"): beside a good running task it makes the active task none
// (broken-file), --task naming it gives none (named-broken), and the tasks table reads it as broken. A format-0 task
// with fields still reads (format 0 refuses nothing for a field).

import (
	"testing"
)

const runningTask = "---\nformat: bonsai.task/1\nid: T-0901\ntitle: x\nstatus: running\nlane: null\ndone_when: []\n" +
	"depends_on: []\nblocked_by: null\ncreated: null\nstarted: null\nfinished: null\nlabels: {}\n---\n"

func TestATaskFileWithNothingInItsFrontmatterIsBroken(t *testing.T) {
	cases := []struct{ name, text string }{
		{"garbage alone, no frontmatter block", "garbage: [\n"},
		{"an empty file", ""},
		{"an empty block", "---\n\n---\n"},
		{"a block of comments alone", "---\n# nothing here\n---\n\n# A body\n"},
		{"two marker lines and nothing between", "---\n---\n"},
		{"a body with no block", "# A task\n\nstatus: running\n"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			root := t.TempDir()
			write(t, root, "tasks/T-0901-running.md", runningTask)
			write(t, root, "tasks/T-0903-junk.md", c.text)
			a, err := Active(ActiveInput{Main: root, TaskDir: "tasks"})
			if err != nil || a.ID != "" || a.Why != ReasonBrokenFile || a.Detail != "tasks/T-0903-junk.md" {
				t.Errorf("beside a running task: %+v %v", a, err)
			}
			if got := a.Sentence(); got != "no active task: a task file does not parse (tasks/T-0903-junk.md)" {
				t.Errorf("the sentence: %q", got)
			}
			a, err = Active(ActiveInput{Main: root, TaskDir: "tasks", Task: "T-0903"})
			if err != nil || a.ID != "" || a.Why != ReasonNamedBroken {
				t.Errorf("--task naming it: %+v %v", a, err)
			}
			if got := a.Sentence(); got != "no active task: the named task's file does not parse (T-0903 in tasks/T-0903-junk.md)" {
				t.Errorf("--task naming it, the sentence: %q", got)
			}
			a, _ = Active(ActiveInput{Main: root, TaskDir: "tasks", Env: "T-0903", EnvCounts: true})
			if a.Why != ReasonNamedBroken {
				t.Errorf("BONSAI_TASK naming it: %+v", a)
			}
			if a, _ := Active(ActiveInput{Main: root, TaskDir: "tasks", Task: "T-0901"}); a.ID != "T-0901" || a.How != HowNamed {
				t.Errorf("--task naming the good one: %+v", a)
			}
			entries, err := ReadTaskFiles(root, "tasks")
			if err != nil || len(entries) != 2 || entries[1].ID != "T-0903" || !entries[1].Broken || entries[1].Task != nil {
				t.Errorf("the tasks table's read: %+v %v", entries, err)
			}
		})
	}
	t.Run("a format-0 task with fields still reads", func(t *testing.T) {
		root := t.TempDir()
		write(t, root, "tasks/T-0901-old.md", "---\nid: T-0901\ntitle: An old task\nstatus: running\n---\n")
		if a, _ := Active(ActiveInput{Main: root, TaskDir: "tasks"}); a.ID != "T-0901" || a.How != HowRunning {
			t.Errorf("%+v", a)
		}
		write(t, root, "tasks/T-0902-bare.md", "---\ntitle: A task with no status\n---\n")
		if a, _ := Active(ActiveInput{Main: root, TaskDir: "tasks"}); a.ID != "T-0901" {
			t.Errorf("a format-0 task with no status beside it: %+v", a)
		}
	})
}
