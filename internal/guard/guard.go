// Package guard is `bonsai hook guard` (spec §7, plan part 5): the PreToolUse hook that refuses an agent's edit of
// a protected path, and blocks whenever it cannot read its input or decide.
//
// What it judges (design/plan-5.md, 5.3.2; spec §7, "What the guard judges"): an Edit, Write, MultiEdit or
// NotebookEdit of a path is judged by one function, "may this agent change this path now" (may.go), which step
// 5.3.3's delete check and rung 0 (step 5.4) call too: the floor of paths guarded whatever a project's lists say
// (floor.go), the lists and the packs' declared paths, the active task's grants, read from the main checkout and
// honoured only while it reads running (grant.go), Bonsai's own files in every mode (own.go and the floor), and in
// command mode the tripwires of contract §10.1 (tripwire.go). A Bash, PowerShell or Monitor call is judged for
// Bonsai's own commands (commands.go): a line that unsets CLAUDE_CODE_CHILD_SESSION here, each bonsai call once
// step 5.3.3's splitter finds it. Every rule a decision names is a row of the rules table (rules.go), its one home.
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
//     the guard's own timer first (Budget, under the hook line's 10 s timeout, engine.GuardTimeout: a hook that times
//     out does not block, spec §7), then runs the guard's work in a goroutine: it finds the session's project from
//     CLAUDE_PROJECT_DIR (set by Claude Code, not by any file an agent edits), reads its bonsai.yaml and the payload,
//     decides, writes the guard record. What does not finish in time blocks, its over-time record awaited for
//     RecordWait; a panic inside blocks.
//   - The payload: Claude Code's PreToolUse payload is read by the one hook adapter (internal/hookin). A payload that
//     does not read, any fault in a value the guard reads, or an event other than PreToolUse blocks.
//   - rule.go: the projects holding a path, the lists each judges by (read leanly, once per run), which call goes to
//     which judgment, and the project a record goes in; paths.go: every form an edit's path may take, and the glob
//     match; may.go and the files it names: the judgment itself; rules.go: the rules table.
//   - record.go: the guard record (bonsai.log/1, contract §8.1), through internal/record's one append path, in the
//     .bonsai/local/log/ of the project whose rules decided.
//   - fault_off.go, fault_on.go: the test fault switch, compiled in only with the build tag bonsai_test_fault. A
//     normal build has fault_off.go alone, which holds no fault code and reads no variable for it.
//
// Exit codes (spec §3): 0 allows the call, 2 blocks it, and Claude Code shows a blocked call's stderr to the agent.
// The hook line (engine.GuardCommand, plan-5 5.3.6) runs the installed bonsai at its fixed place, never one on the
// PATH: it refuses the call in its own words when bonsai is at neither place, its || exit 2 turns every other ending
// (a crash) into a block, and its onFailure: block (Claude Code 2.1.295 and later) blocks a line that cannot start.
//
// A session started in a folder below a project's top, or moved by a person's /cd, loads none of the project's
// settings (Claude Code reads .claude/settings.json from the folder a session starts in), so no hook line, the guard's
// among them, runs there: bonsai init's closing words and this hook's help say so.
//
// Known limit, placed for step 5.4: a worktree with no bonsai.yaml of its own. Inside the main checkout's
// .claude/worktrees/, its edits are judged as the main checkout's paths under its lists, so a person_only .claude/**
// refuses every one of them; outside the main checkout, it is judged by no lists at all.
package guard

// Decision is the guard's answer to one call.
type Decision struct {
	Allow  bool   // true: exit 0, the call goes on; false: exit 2, the call is blocked
	Rule   string // the rule that decided: a name of the rules table (rules.go)
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
