// Package engine is Bonsai's engine (spec §6, plan part 3): it brings a project to the packs bonsai.yaml names and
// keeps it there without overwriting a person's edits.
//
//   - fetch.go: the plain fetch, git at a commit from the pack's URL into the home's pack cache (spec §5), and
//     reading a pack at a commit;
//   - plan.go: what init and update would do, file by file (spec §6's table of three fingerprints), and the lock
//     they would write;
//   - settings.go: Bonsai's lines in the project's .claude/settings.json (kind keys): hook lines, deny rules, the
//     plugin wiring, autoMemoryEnabled and disableAllHooks, each with the sentence the preview prints;
//   - block.go: the instruction block in CLAUDE.md (kind block);
//   - apply.go: the all-or-nothing write: every file staged, renames, the lock last;
//   - check.go: bonsai check's findings on the lock and the files, and a tracked or staged .bonsai/local/ file;
//   - config.go: bonsai.yaml as init writes it, a comment on every line, and init --new-id's new id;
//   - render.go and diff.go: the preview and the result, in plain ASCII text and in JSON.
//
// The engine applies only what a person asked for: init and update write project files, the home's pack cache and
// the copies --adopt saves in it, and nothing else (no ~/.claude file, no commit, no plugin install: installing is
// plan part 4). A change to a hook line runs code, and this build only refuses it: --allow-exec is step 5.1.
//
// The hook lines, for a reviewer (plan part 5's verifier reads them):
//   - what is written: settings.go, ownHooks (Bonsai's own line, `bonsai hook guard || exit 2` on PreToolUse, by
//     name, in shell form) and buildLines (each pack's hooks entries, as its pack.yaml gives them); applyLines puts
//     them in the file, one group per event and matcher, beside the project's own hooks;
//   - an old Bonsai line taken out at a first link: settings.go, isOldBonsaiHook and claim;
//   - the refusal: lineChanges marks a hook line added or changed against the lines the lock last consented to
//     (RunsCode); plan.go sets Plan.HookChange when an update (not a first link) would write one; apply.go refuses
//     such a plan whatever the caller asks; cmd/bonsai/engine.go prints the refusal (exit 4) naming --allow-exec,
//     and refuses --allow-exec itself (exit 2) until step 5.1;
//   - the tests: engine_test.go, TestHookLineChangeIsRefused and TestCheck1InitIntoADriftedProject;
//     cmd/bonsai/engine_test.go, TestUpdateCommand.
package engine

import (
	"fmt"
	"strings"
)

// Exit codes (spec §3): 0 ok, 1 check findings, 2 bad input, 3 runtime, 4 wrong state or no --yes, 5 conflicts.
const (
	ExitOK       = 0
	ExitFindings = 1
	ExitInput    = 2
	ExitRuntime  = 3
	ExitState    = 4
	ExitConflict = 5
)

// Error is a refusal or a failure: its exit code, what is wrong, and the next step (spec §3: every refusal and
// error names the next thing to do). Its text is ASCII.
type Error struct {
	Exit int
	What string
	Next string
}

func (e *Error) Error() string {
	return ascii(strings.TrimSuffix(e.What, ".")) + "; next: " + ascii(e.Next)
}

func errorf(exit int, next, format string, args ...any) *Error {
	return &Error{Exit: exit, What: fmt.Sprintf(format, args...), Next: next}
}

// ascii keeps printable ASCII and writes any other character as a Go escape (é), so text prints unbroken in
// PowerShell 5.1 (spec §3). A line feed is kept.
func ascii(s string) string {
	var b strings.Builder
	for _, r := range s {
		if (r >= 0x20 && r < 0x7f) || r == '\n' {
			b.WriteRune(r)
			continue
		}
		q := fmt.Sprintf("%+q", string(r))
		b.WriteString(q[1 : len(q)-1])
	}
	return b.String()
}

// short gives a commit's first 7 characters, as people read them; "" stays "".
func short(commit string) string {
	if len(commit) > 7 {
		return commit[:7]
	}
	return commit
}
