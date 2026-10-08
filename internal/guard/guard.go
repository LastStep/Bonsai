// Package guard is `bonsai hook guard` (spec §7, plan part 5): the PreToolUse hook that refuses an agent's edit of
// a protected path, and blocks whenever it cannot read its input or decide.
//
// The walking skeleton's guard has one rule: an Edit, Write, MultiEdit or NotebookEdit of a path on bonsai.yaml's
// protected list, or on its person_only list (a part of the protected one, spec §6), is refused. There are no tasks
// yet, so nothing grants a protected path. Grants, command mode's tripwires, the .bonsai/local/ and table refusals
// and the shell delete check are step 5.3's; Bash and PowerShell calls are allowed here.
//
// For a reviewer, in the order a call runs:
//   - hook.go: Main, the whole run. It arms the guard's own timer first (Budget, under the 10 s timeout part 3's hook
//     line carries: a hook that times out does not block, spec §7), then works in a goroutine: finds the project
//     from CLAUDE_PROJECT_DIR (set by Claude Code, not by any file an agent edits), reads bonsai.yaml and the
//     payload, decides, writes the guard record. What does not finish in time blocks; a panic inside blocks.
//   - input.go: Claude Code's PreToolUse payload. A payload the guard cannot read blocks.
//   - rule.go: the rule; paths.go: every project-relative form an edit's path may take, and the glob match.
//   - record.go: the guard record (bonsai.log/1, contract §8.1) in the project's .bonsai/local/log/, with the
//     binary's own path on every record and its SHA-256 on a session's first (spec §3, §8).
//   - fault_off.go, fault_on.go: the test fault switch, compiled in only with the build tag bonsai_test_fault. A
//     normal build has fault_off.go alone, which holds no fault code and reads no variable for it.
//
// Exit codes (spec §3): 0 allows the call, 2 blocks it, and Claude Code shows a blocked call's stderr to the agent.
// Part 3's hook line, `bonsai hook guard || exit 2`, turns every other ending (a crash, a missing binary) into a block.
package guard

// The rules a guard decision names: the log record's rule field (contract §8.1). This block is the list's one home.
const (
	RuleProtected    = "protected"        // deny: the path is on bonsai.yaml's protected list
	RulePersonOnly   = "person-only"      // deny: the path is on bonsai.yaml's person_only list
	RuleNotProtected = "not-protected"    // allow: a file tool's path inside the project, on neither list
	RuleOutside      = "outside-project"  // allow: a file tool's path outside the project, in every form it takes
	RuleShell        = "shell-not-judged" // allow: Bash or PowerShell (the delete check is step 5.3)
	RuleOtherTool    = "tool-not-judged"  // allow: a tool other than the file tools and the shells
	RuleBadInput     = "bad-input"        // deny: the payload cannot be read
	RuleBadGlob      = "bad-glob"         // deny: a glob on bonsai.yaml's lists is not a glob the guard reads
	RuleLogFailed    = "log-failed"       // deny: the guard would allow the call but cannot record it
	RuleOverTime     = "over-time"        // deny: the guard ran past its own time limit
	RuleInternal     = "internal-error"   // deny: the guard failed inside (a panic)

	// Blocked with no record, since the workspace id a record needs is not known.
	RuleNoProject = "no-project" // deny: CLAUDE_PROJECT_DIR is not set, or its folder cannot be read
	RuleBadConfig = "bad-config" // deny: bonsai.yaml is there but cannot be read

	// Allowed with no record (spec §3, §7: in a project with no bonsai.yaml, bonsai hook exits 0 and records nothing).
	RuleNotLinked = "not-linked"
)

// Decision is the guard's answer to one call.
type Decision struct {
	Allow  bool   // true: exit 0, the call goes on; false: exit 2, the call is blocked
	Rule   string // the rule that decided (the constants above)
	Target string // the project-relative path judged, forward slashes; "" for none or a path outside the project
	Why    string // a refusal's reason, said first (spec §7); ASCII
	Next   string // a refusal's next step (spec §3: every refusal names one); ASCII
}

func allow(rule, target string) Decision {
	return Decision{Allow: true, Rule: rule, Target: target}
}

func deny(rule, why, next string) Decision {
	return Decision{Rule: rule, Why: why, Next: next}
}
