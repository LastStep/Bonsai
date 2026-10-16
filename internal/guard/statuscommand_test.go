package guard

// The guard on a status_command outside its rule (plan-5 5.6.1, 1): workspace.LoadMachineSettings refuses a file
// whose status_command holds ;, $, a backtick, a quote, a line break or more than 100 characters, written straight into
// a scratch home's file, so the guard meets it on its fail-closed path: a call whose judgment needs the mode is refused
// (bad-settings), naming the file, and the command is never printed in a refusal or handed to an agent.

import (
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/workspace"
)

func TestStatusCommandOutsideItsRule(t *testing.T) {
	for _, command := range []string{"move; rm -rf ~", "move $(id)", "move `id`", `move "x"`, "move\nnext", strings.Repeat("a", 101)} {
		dir := withTask(t, "todo")
		machineSettings(t, dir, `{"status_writes": "command", "status_command": `+strconv.Quote(command)+`}`)
		home, err := workspace.Home()
		if err != nil {
			t.Fatal(err)
		}
		md, err := workspace.MachineDir(home, dir)
		if err != nil {
			t.Fatal(err)
		}
		stderr := wantRule(t, dir, editIn(dir, "work/tasks/T-0001-the-task.md", "title: A made-up task", "title: Another"), RuleBadSettings)
		if !strings.Contains(stderr, filepath.ToSlash(filepath.Join(md, workspace.MachineSettingsFile))) || !strings.Contains(stderr, "status_command") {
			t.Errorf("%q: the refusal names neither the file nor the key: %s", command, stderr)
		}
		if strings.Contains(stderr, "rm -rf") || strings.Contains(stderr, "$(id)") || strings.Contains(stderr, "`id`") {
			t.Errorf("%q: the refusal prints the command: %s", command, stderr)
		}
	}
}
