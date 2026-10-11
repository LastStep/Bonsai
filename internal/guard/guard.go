// Package guard is `bonsai hook guard` (spec §7, plan part 5): the PreToolUse hook that refuses an agent's edit of
// a protected path, and blocks whenever it cannot read its input or decide.
//
// This build's guard refuses an Edit, Write, MultiEdit or NotebookEdit of a path on the protected or person_only list
// (a part of the protected one, spec §6) of a project holding it, or among the protected paths a pack of that project
// declares. Nothing grants a protected path yet: the active task's grants, the floor, command mode's tripwires and
// the .bonsai/local/ and table refusals are step 5.3.2's, the shell delete check 5.3.3's; Bash and PowerShell calls
// are allowed here.
//
// Where it reads from (design/plan-5.md, 5.3.1): a path is judged by the projects holding it, walking past any .git to
// the volume's root, the nearest alone only when it is a verified checkout; a worktree's main checkout is found from
// its .git with no git process and verified against git's own link back and this machine's record of the main
// checkout (workspace.VerifyMain), and a verified worktree judges by its main checkout's bonsai.yaml and lock and
// writes its records into the main checkout's .bonsai/local/. Where nothing holding a refused path is verified, the
// refusal is main-not-verified and names the person's step that verifies it.
//
// For a reviewer, in the order a call runs:
//   - hook.go: Main, the whole run, in the frame of the hooks that block (internal/hookin's Frame). The frame arms
//     the guard's own timer first (Budget, under the 10 s timeout part 3's hook line carries: a hook that times out
//     does not block, spec §7), then runs the guard's work in a goroutine: it finds the session's project from
//     CLAUDE_PROJECT_DIR (set by Claude Code, not by any file an agent edits), reads its bonsai.yaml and the payload,
//     decides, writes the guard record. What does not finish in time blocks, its over-time record awaited for
//     RecordWait; a panic inside blocks.
//   - The payload: Claude Code's PreToolUse payload is read by the one hook adapter (internal/hookin). A payload that
//     does not read, any fault in a value the guard reads, or an event other than PreToolUse blocks.
//   - rule.go: the projects holding a path, the lists each judges by (read leanly, once per run), the decision, and
//     the project a record goes in; paths.go: every form an edit's path may take, and the glob match.
//   - record.go: the guard record (bonsai.log/1, contract §8.1), through internal/record's one append path, in the
//     .bonsai/local/log/ of the project whose rules decided.
//   - fault_off.go, fault_on.go: the test fault switch, compiled in only with the build tag bonsai_test_fault. A
//     normal build has fault_off.go alone, which holds no fault code and reads no variable for it.
//
// Exit codes (spec §3): 0 allows the call, 2 blocks it, and Claude Code shows a blocked call's stderr to the agent.
// Part 3's hook line, `bonsai hook guard || exit 2`, turns every other ending (a crash, a missing binary) into a block.
//
// A session started in a folder below a project's top, or moved by a person's /cd, loads none of the project's
// settings (Claude Code reads .claude/settings.json from the folder a session starts in), so no hook line, the guard's
// among them, runs there: bonsai init's closing words and this hook's help say so.
package guard

// The rules a guard decision names: the log record's rule field (contract §8.1). This block is the list's one home.
const (
	RuleProtected       = "protected"         // deny: the path is on a protected list (bonsai.yaml's, or a pack's declared paths)
	RulePersonOnly      = "person-only"       // deny: the path is on bonsai.yaml's person_only list
	RuleMainNotVerified = "main-not-verified" // deny: the path is on a list, and no project holding it is verified, so no grant counts
	RuleNotProtected    = "not-protected"     // allow: a file tool's path inside a project, on no list
	RuleOutside         = "outside-project"   // allow: a file tool's path in no linked project, in every form it takes
	RuleShell           = "shell-not-judged"  // allow: Bash or PowerShell (the delete check is step 5.3.3)
	RuleOtherTool       = "tool-not-judged"   // allow: a tool other than the file tools and the shells
	RuleBadInput        = "bad-input"         // deny: the payload cannot be read
	RuleBadGlob         = "bad-glob"          // deny: a glob on a list is not a glob the guard reads
	RuleBadPath         = "bad-path"          // deny: Windows cannot name where the path leads, or it runs through a dangling junction or link
	RuleBadLock         = "bad-lock"          // deny: the lock's declares cannot be read (or the lock is missing while bonsai.yaml names packs)
	RuleLogFailed       = "log-failed"        // deny: the guard would allow the call but cannot record it
	RuleOverTime        = "over-time"         // deny: the guard ran past its own time limit
	RuleInternal        = "internal-error"    // deny: the guard failed inside (a panic)

	// Blocked with no record when it is the session's own bonsai.yaml, since the workspace id a record needs is not
	// known; recorded in the session's project when it is the bonsai.yaml of another project holding the path.
	RuleNoProject = "no-project" // deny: CLAUDE_PROJECT_DIR is not set, or its folder cannot be read
	RuleBadConfig = "bad-config" // deny: bonsai.yaml is there but cannot be read

	// Allowed with no record (spec §3, §7: in a project with no bonsai.yaml, bonsai hook exits 0 and records nothing):
	// neither the session's folder nor the path is in a linked project.
	RuleNotLinked = "not-linked"
)

// Decision is the guard's answer to one call.
type Decision struct {
	Allow  bool   // true: exit 0, the call goes on; false: exit 2, the call is blocked
	Rule   string // the rule that decided (the constants above)
	Target string // the path judged, relative to the project its record goes in, forward slashes; "" for none
	Why    string // a refusal's reason, said first (spec §7); ASCII
	Next   string // a refusal's next step (spec §3: every refusal names one); ASCII
}

func allow(rule, target string) Decision {
	return Decision{Allow: true, Rule: rule, Target: target}
}

func deny(rule, why, next string) Decision {
	return Decision{Rule: rule, Why: why, Next: next}
}
