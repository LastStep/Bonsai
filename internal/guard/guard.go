// Package guard is `bonsai hook guard` (spec §7, plan part 5): the PreToolUse hook that refuses an agent's edit of
// a protected path, and blocks whenever it cannot read its input or decide.
//
// What it judges (design/plan-5.md, 5.3.2; spec §7, "What the guard judges"): an Edit, Write, MultiEdit or
// NotebookEdit of a path is judged by one function, "may this agent change this path now" (may.go), which step
// 5.3.3's delete check and rung 0 (step 5.4) call too: the floor of paths guarded whatever a project's lists say
// (floor.go), the lists and the packs' declared paths, the active task's grants, read from the main checkout and
// honoured only while it reads running (grant.go), Bonsai's own files in every mode (own.go and the floor), and in
// command mode the tripwires of contract §10.1 (tripwire.go). A Bash, PowerShell or Monitor call is judged for
// Bonsai's own commands (commands.go): a line that unsets an agent marker here, each bonsai call once
// step 5.3.3's splitter finds it. Every rule a decision names is a row of the rules table (rules.go), its one home.
//
// Where the project's lock declares roles (step 5.5.11), the roles' rules and the stall stops come first (roles.go,
// budget.go, rolecmds.go; their state in internal/agents): who may call whom on Agent, Task, SendMessage and Workflow;
// the file tools and the shell writes it can read refused to a role that changes nothing; an agent's readable shell
// writes into Bonsai's own records and the lock refused; a start of Claude Code on an agent's shell line refused;
// bonsai message's --from held to its writer's address; and each agent's budget, counted on every call. There the
// engine writes a second line, bonsai hook guard --second (Options.Second), on every other tool: it refuses only past
// a budget, and lets a call through on a fault of its own.
//
// Where it reads from (design/plan-5.md, 5.3.1): a path is judged by the projects holding it, walking past any .git to
// the volume's root, the nearest alone only when it is a verified checkout; a worktree's main checkout is found from
// its .git with no git process and verified against git's own link back and this machine's record of the main
// checkout (workspace.VerifyMain), and a verified worktree judges by its main checkout's bonsai.yaml and lock and
// writes its records into the main checkout's .bonsai/local/. Where nothing holding a refused path is verified, the
// refusal is main-not-verified and names the person's step that verifies it.
//
// The lists are read live, from the main checkout's bonsai.yaml as the file holds it now (5.4's fix rounds left this
// as it was): a shell write there that takes a path off protected, person_only or never_edit (question A leaves a shell
// write unjudged in agents mode) opens that path to a file-tool edit during the work. No such work climbs green or ends
// a session with a task: the ladder runner and the stop gate read bonsai.yaml as committed at the main checkout's HEAD,
// refusing it while the file differs from it (config-uncommitted, stop-config-uncommitted), and once it is committed,
// while its settings for the checks are looser than the ones this machine last accepted (config-looser,
// stop-config-looser; stopgate.Accepted, workspace/checks.go), so a person sees the change before it counts.
//
// For a reviewer, in the order a call runs:
//   - hook.go: Main, the whole run, in the frame of the hooks that block (internal/hookin's Frame). The frame arms
//     the guard's own timer first (Budget, under the hook line's 10 s timeout, engine.GuardTimeout: a hook that times
//     out does not block, spec §7), then runs the guard's work in a goroutine: it finds the session's project from
//     CLAUDE_PROJECT_DIR (set by Claude Code, not by any file an agent edits), reads its bonsai.yaml and the payload,
//     decides, writes the guard record. What does not finish in time blocks, its over-time record awaited for
//     RecordWait; a panic inside blocks.
//   - The payload: Claude Code's PreToolUse payload is read by the one hook adapter (internal/hookin), each tool's input
//     the guard does not judge passed over unread (judgedTool). A payload that does not read, any fault in a value the
//     guard reads, or an event other than PreToolUse blocks (on the second line, lets the call through).
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
// Known limit: a worktree with no bonsai.yaml of its own (its branch made before the link, or one that deleted it).
// Inside the main checkout's .claude/worktrees/, its edits are judged as the main checkout's paths under its lists, so
// a person_only .claude/** refuses every one of them; outside the main checkout, it is judged by no lists at all. Rung
// 0 (internal/rung0, step 5.4.3) catches that case at the climb: it judges every path such a worktree changed against
// the verified main checkout's lists and the named task's grants, wherever the worktree lies, and it is red while
// bonsai check cannot run there (no bonsai.yaml of its own), so such a worktree's work never climbs green on a floor
// that holds rung 0.
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
