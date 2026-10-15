package main

// Bonsai's own commands refusing an agent (design/plan-5.md, 5.3.6 note 8; the director's answer (ii) to step 5.3's second
// question, 9 Oct: who may consent to code, and take the guard out). In an agent's session, which Claude Code marks by
// setting CLAUDE_CODE_CHILD_SESSION in every process its Bash, PowerShell and Monitor tools and its hooks start (its
// env-vars reference), so scripts, xargs and subprocesses inherit it, init, update and unlink refuse a person's
// command:
//   - --allow-exec, the consent to code that runs on this machine (init and update);
//   - unlink, which takes Bonsai's guard and stop gate out of the project;
//   - an update (or init in a linked project) that changes or takes out Bonsai's own guard line or stop gate's line
//     (engine.Plan.OwnGuardChanges): counted as a person's command here, though for consent a removed line runs
//     nothing (5.1.1 rule 1), so a person's own update that takes one out still needs only --yes;
//   - an update or init that would record, in this machine's record of the main checkout, another base branch than the
//     one recorded (engine.BaseChange), or settings for the checks looser than the ones recorded (engine.ChecksLooser,
//     workspace/checks.go): the ladder and the stop gate hold the project to both, and an agent that could record its
//     own would make them moot. The first record of either is free.
//
// The one exception, (ii): a project in command mode whose running task's bonsai.allows holds bonsai.yaml and
// .bonsai/lock.json, which in command mode only a person's tap puts there (guard.PersonGrant, the guard's own
// judgment of rule 7, its one home). The refusal (person-command, exit 4) names the next step: a person runs the
// command in a terminal of their own. The variable counts when it holds any value, as interactive (engine.go) reads it
// for the y/N question: Claude Code sets it to 1, and tests clear it to run as a person (testpack.Isolate). The guard's
// rule 7 judges the same commands in a shell call, and a line that unsets the variable, as a second layer: a flag in a
// variable, a script or xargs is hidden from it, but not from the command. A test build (engine.Stamped) skips this:
// its scratch root confines it (engine's confine), so builders' links to the test pack keep working.
//
// --allow-exec has one spelling: the word tables' parser takes --allow-exec alone and refuses --allow-exec=true and
// any other spelling, and no variable or file stands for it.

import (
	"os"

	"github.com/LastStep/Bonsai/internal/engine"
	"github.com/LastStep/Bonsai/internal/guard"
)

// agentSession reports a session Claude Code started (CLAUDE_CODE_CHILD_SESSION holding a value) in a plain build: a
// test build never counts one.
func agentSession() bool {
	return !engine.Stamped() && os.Getenv(guard.ChildSessionEnv) != ""
}

// personNext is the next step of a person's command refused to an agent: the command, run by a person.
func personNext(command string) string {
	return "a person runs it in a terminal of their own, not one Claude Code started: " + command
}

// personCommand refuses a person's command in an agent's session, or gives nil when it may run: no agent's session,
// or a person's grant in a project in command mode (guard.PersonGrant). shown is the command as typed, what says why
// it is a person's, and command is what the person runs.
func personCommand(dir, home, shown, what, command string) *engine.Error {
	if !agentSession() {
		return nil
	}
	why := guard.PersonGrant(home, dir, os.Getenv)
	if why == "" {
		return nil
	}
	return &engine.Error{Code: "person-command", Exit: engine.ExitState, Who: "person",
		What: shown + " is a person's command (" + what + "), and this is an agent's session (" + guard.ChildSessionEnv +
			" is set): an agent may run it only in a project in command mode, under a running task whose bonsai.allows a " +
			"person gave bonsai.yaml and .bonsai/lock.json, and " + engine.ASCII(why) + "; nothing was written",
		Next: personNext(command)}
}
