package guard

// Crying wolf, by test (design/plan-5.md, 5.3.2 note 10): a guard that refuses ordinary work gets worked around, so
// the ordinary calls must pass: an edit in a running task's worktree, a granted path, a free path in a project with a
// long protected list, a new file in a new free folder, .bonsai/STATE.md; in command mode, an edit of a task's body or
// title, a new run report, STATE rewritten in its format; ordinary shell lines, those that name
// CLAUDE_CODE_CHILD_SESSION to read it among them; Bonsai's own read-only words.

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/testpack"
)

func TestOrdinaryWorkPasses(t *testing.T) {
	t.Run("a running task's worktree", func(t *testing.T) {
		tmp := t.TempDir()
		home := testpack.Isolate(t, tmp)
		main, wt := gitProject(t, home, grantYAML, time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC))
		writeFile(t, filepath.Join(main, "work", "tasks", "T-0001-the-task.md"), taskText("T-0001", "running", "light", "assets/**"))
		writeFile(t, filepath.Join(wt, "src", "main.go"), "package main\n")
		for _, in := range []string{
			editIn(wt, "src/main.go", "package main", "package main // edited"),
			writeOf(wt, "assets/logo.png", "png\n"),
			writeOf(wt, "brand/new/folder/notes.md", "# Notes\n"),
			writeOf(wt, ".bonsai/STATE.md", "# Where it stands\n"),
		} {
			if code, s := answer(t, wt, in, "BONSAI_TASK", "T-0001"); code != 0 {
				t.Errorf("exit %d, %s", code, s)
			}
		}
	})
	t.Run("a long protected list", func(t *testing.T) {
		var globs []string
		for i := 0; i < 300; i++ {
			globs = append(globs, fmt.Sprintf(`"area%03d/**"`, i), fmt.Sprintf(`"docs/page%03d.md"`, i))
		}
		yaml := strings.Replace(grantYAML, `protected: [`, "protected: ["+strings.Join(globs, ", ")+", ", 1)
		dir := grantProject(t, yaml, map[string]string{"src/app.go": "package app\n"})
		for _, in := range []string{editIn(dir, "src/app.go", "package app", "package app // x"), writeOf(dir, "docs/guide.md", "x\n"),
			writeOf(dir, "new/deep/file.txt", "x\n")} {
			if code, s := answer(t, dir, in); code != 0 {
				t.Errorf("exit %d, %s", code, s)
			}
		}
	})
	t.Run("command mode's ordinary edits", func(t *testing.T) {
		dir := withTask(t, "running", "assets/**")
		commandMode(t, dir)
		task := "work/tasks/T-0001-the-task.md"
		run := "---\nformat: bonsai.run/1\nid: R-2026-10-11-T-0001\ntask: T-0001\nrole: builder\nmodel: a-model\nstarted: " +
			"2026-10-11 09:00\nfinished:\noutcome: running\ncommits: []\nlabels: {}\n---\n\n# Run\n"
		state := "---\nformat: bonsai.state/1\nupdated: 2026-10-11\nupdated_by: an agent\nlabels: {}\n---\n\n# Where it stands\n"
		for _, in := range []string{
			editIn(dir, task, "# A made-up task", "# A made-up task\n\nNotes from the builder."),
			editIn(dir, task, "title: A made-up task", "title: A clearer title"),
			editIn(dir, task, `  - "It is done"`, `  - "It is done"`+"\n"+`  - "And tested"`),
			writeOf(dir, "work/runs/R-2026-10-11-T-0001.md", run),
			writeOf(dir, ".bonsai/STATE.md", state),
			writeOf(dir, "work/runs/README.md", "# Runs\n"),
			editIn(dir, "free.txt", "free", "freer"),
			writeOf(dir, "assets/new.png", "png\n"),
		} {
			if code, s := answer(t, dir, in); code != 0 {
				t.Errorf("exit %d, %s", code, s)
			}
		}
	})
	t.Run("ordinary shell lines", func(t *testing.T) {
		dir := withTask(t, "todo")
		for _, c := range []struct{ tool, line string }{
			{"Bash", "git status && go test ./... && go vet ./..."}, {"Bash", "npm install && npm run build"},
			{"Bash", "rm -rf node_modules"}, {"Bash", "echo \"session: $CLAUDE_CODE_CHILD_SESSION\""},
			{"Bash", `grep -rn "CLAUDE_CODE_CHILD_SESSION" cmd/ internal/`}, {"Bash", "env | sort"}, {"Bash", "env GOOS=windows go build ./..."},
			{"Bash", "export PATH=$PATH:/usr/local/go/bin"}, {"Bash", "git commit -m 'guard: reads CLAUDE_CODE_CHILD_SESSION'"},
			{"PowerShell", "Get-ChildItem -Recurse | Select-String CLAUDE_CODE_CHILD_SESSION"}, {"PowerShell", "$env:GOOS = 'linux'"},
			{"PowerShell", "Write-Output $env:CLAUDE_CODE_CHILD_SESSION"}, {"Monitor", "tail -f build.log"},
			{"Bash", "bonsai status --json && bonsai check && bonsai logs --json"},
		} {
			if code, s := answer(t, dir, bashIn(dir, c.tool, c.line)); code != 0 {
				t.Errorf("%s %q: exit %d, %s", c.tool, c.line, code, s)
			}
		}
	})
	t.Run("Bonsai's own read-only words", func(t *testing.T) {
		dir := withTask(t, "todo")
		j := judgeIn(dir)
		for _, argv := range [][]string{{"bonsai", "status", "--json"}, {"bonsai", "check", "--json"}, {"bonsai", "logs"},
			{"bonsai", "asks", "--json"}, {"bonsai", "ask", "--status", "k"}, {"bonsai", "settings", "show"}, {"bonsai", "--help"}} {
			if d := j.ownCommand(dir, argv, true); !d.Allow {
				t.Errorf("%q: %+v", argv, d)
			}
		}
	})
}
