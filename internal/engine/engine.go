// Package engine is Bonsai's engine (spec §6, plan part 3): it brings a project to the packs bonsai.yaml names and
// keeps it there without overwriting a person's edits.
//
//   - fetch.go: the plain fetch, git at a commit from the pack's URL into the home's pack cache (spec §5), and
//     reading a pack at a commit;
//   - plan.go: what init and update would do, file by file (spec §6's table of three fingerprints), and the lock
//     they would write;
//   - settings.go: Bonsai's lines in the project's .claude/settings.json (kind keys): hook lines, deny rules, the
//     plugin wiring, autoMemoryEnabled and disableAllHooks, each with the sentence the preview prints;
//   - block.go: the instruction block in CLAUDE.md (kind block): the workspace line, the protocol and memory imports,
//     the packs' labels and block.md texts, at most 40 lines;
//   - apply.go: the all-or-nothing write: every file staged, renames, the lock last;
//   - check.go: bonsai check's findings on the lock and the files, and a tracked or staged .bonsai/local/ file;
//   - plugins.go: this machine's plugins (plan part 4b): the install Claude Code is asked for after init and update
//     (at project scope, the checkout's own .claude/settings.json), and check's drift report against the lock;
//   - consent.go: consent to code (step 5.1.1): what init and update write that runs code, which needs --allow-exec
//     as well as --yes, at a first link too;
//   - declares.go: what a pack declares (lanes, document kinds, labels, protected paths, hook lines, deny rules), read
//     at its commit and copied into the lock's declares, so check, status and the guard read them with no pack at
//     hand (step 5.1.5); fetch.go also refuses a moved tag, and plan.go lists the lock's format0 at a first link;
//   - config.go: bonsai.yaml as init writes it, from the built-in template (every field, a comment on every line),
//     and init --new-id's new id;
//   - render.go and diff.go: the preview and the result, in plain ASCII text and in JSON.
//
// The engine applies only what a person asked for: init and update write project files, the home's pack cache and
// the copies --adopt saves in it, and nothing else (no ~/.claude file, no local settings file, no commit). After
// writing, cmd/bonsai asks Claude Code to install each pack's plugin at project scope (InstallPlugins), which writes
// Claude Code's plugin folder and, the first time, the checkout's .claude/settings.json in Claude Code's own key
// order.
//
// Consent to code, for a reviewer (plan-5, piece 5.1.1, rules 1-8; the verifier reads it):
//   - what is written: settings.go, ownHooks (Bonsai's own line, `bonsai hook guard || exit 2` on PreToolUse, by
//     name, in shell form) and buildLines (each pack's hooks entries, as its pack.yaml gives them); applyLines puts
//     them in the file, one group per event and matcher, beside the project's own hooks;
//   - an old Bonsai line taken out at a first link: settings.go, isOldBonsaiHook and claim;
//   - what runs code: lineChanges marks a hook line added or changed against the lines the lock last consented to
//     (SettingsChange.RunsCode); consent.go turns those, the pack files a hook runs (pack.yaml's runs) and a plugin's
//     own code parts into Plan.RunsCode, leaving out only Bonsai's own hook lines added at a first link (OwnHooks);
//   - the refusal: apply.go refuses a plan with anything in RunsCode unless the request has AllowExec, whatever the
//     caller asks; cmd/bonsai/engine.go prints the refusal (exit 4) naming --allow-exec, at a terminal too;
//   - no way round it by deleting the lock: plan.go refuses an update when bonsai.yaml is there and the lock is not
//     (exit 4, naming git checkout of the lock or bonsai init); init then links again judged against the disk;
//   - the baseline: plan.go reads each locked pack at its locked commit and takes it as the baseline only when
//     bonsai.yaml names the folder the lock records and its content hashes to the lock's sha256 (path_test.go), and
//     the consented hook lines only when they hash to the lock's settings
//     record; else the pack is unverified and its code counts as at a first link. Bonsai's own lines carry Line.Own,
//     which no pack sets, and no pack may take the id bonsai; pack files may not land where Claude Code loads code
//     (workspace.checkPackTarget); runs is checked against the hook commands (checkRuns); a plugin that carries code
//     is installed on this machine only with --allow-exec (plugins.go);
//   - the tests: consent_test.go (the table of every consent case), leaks_test.go (the verifier's B1, S1, B2, B3 and
//     S3), plugins_test.go (TestInstallPluginsWithCode, TestInstallOnlyOurPlugins), engine_test.go (TestHookLineChangeIsRefused,
//     TestUpdateWithoutALockIsRefused, TestCheck1InitIntoADriftedProject); cmd/bonsai/engine_test.go
//     (TestConsentCommand, TestUpdateCommand, TestUpdateWithoutALock).
package engine

import (
	"errors"
	"fmt"
	"strings"

	"github.com/LastStep/Bonsai/internal/format"
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

// Error is a refusal or a failure: its word, its exit code, what is wrong, and the next step (spec §3: every refusal
// and error names the next thing to do), with who takes it. Its text is ASCII. Object gives it as the error object
// of a command's --json (bonsai.error).
type Error struct {
	Code string // the error object's word, one of format.ErrorWords (their one home)
	Exit int
	What string
	Next string
	Who  string // agent or person; "" for the word's usual one (format.ErrorWords)
	// Workspace is the workspace the command had read when it stopped (Build fills it once bonsai.yaml is read), for
	// the changes output of a refusal before a plan; nil before that.
	Workspace *format.WorkspaceRef
}

func (e *Error) Error() string {
	return ascii(strings.TrimSuffix(e.What, ".")) + "; next: " + ascii(e.Next)
}

// errorf makes an Error with its word (format.ErrorWords), exit code and next step.
func errorf(code string, exit int, next, format string, args ...any) *Error {
	return &Error{Code: code, Exit: exit, What: fmt.Sprintf(format, args...), Next: next}
}

// Object is the error as the error object of a command's --json (bonsai.error, spec §3, §16 row 29): its word, the
// sentence and the next step, in ASCII, with who takes it: the refusal's own, else the word's usual one. A next step
// of several lines (the two ways to settle conflicts) is one line here. An Error with no word is unexpected (a bug,
// which cmd/bonsai's tests rule out).
func (e *Error) Object() *format.ErrorObject {
	code, who := e.Code, e.Who
	if code == "" {
		code = "unexpected"
	}
	if who == "" {
		w, _ := format.ErrorWord(code)
		who = w.Who
	}
	if who != "person" {
		who = "agent"
	}
	next := strings.ReplaceAll(strings.TrimSpace(e.Next), "\n  ", " ")
	next = strings.ReplaceAll(next, "\n", " ")
	return &format.ErrorObject{Code: code, Message: ascii(strings.TrimSuffix(e.What, ".")), Next: format.Next{Do: ascii(next), Who: who}}
}

// Unexpected gives any error as an Error: an *Error as it is, anything else as unexpected (exit 3).
func Unexpected(err error) *Error {
	var e *Error
	if errors.As(err, &e) {
		return e
	}
	return &Error{Code: "unexpected", Exit: ExitRuntime, What: err.Error(), Next: "run the command again; if it fails again, report it to Bonsai's maintainers"}
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
