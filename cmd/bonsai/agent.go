package main

// Bonsai's own commands refusing an agent (design/plan-5.md, 5.3.6 note 8; the director's answer (ii) to step 5.3's second
// question, 9 Oct: who may consent to code, and take the guard out). In an agent's session, which an agent marks by
// setting one of the agent markers (format.AgentMarkers, their one home: Claude Code's CLAUDE_CODE_CHILD_SESSION, set
// in every process its Bash, PowerShell and Monitor tools and its hooks start, AI_AGENT, pi's PI_CODING_AGENT and
// the other agents'; T-5727) in every process its shell tool starts, so scripts, xargs and subprocesses inherit it,
// init, update and unlink refuse a person's command:
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
// command in a terminal of their own, and the refusal names the marker it found. A marker counts when it holds any
// value, as interactive (engine.go) reads them for the y/N question: Claude Code sets its own to 1, and tests clear
// them all to run as a person (testpack.Isolate). The guard's rule 7 judges the same commands in a shell call, and a
// line that unsets or assigns a marker, as a second layer: a flag in a variable, a script or xargs is hidden from it,
// but not from the command. A test build (engine.Stamped) skips this:
// its scratch root confines it (engine's confine), so builders' links to the test pack keep working.
//
// --allow-exec has one spelling: the word tables' parser takes --allow-exec alone and refuses --allow-exec=true and
// any other spelling, and no variable or file stands for it.

import (
	"os"
	"path/filepath"

	"github.com/LastStep/Bonsai/internal/engine"
	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/guard"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// agentSession gives the agent marker (format.AgentMarkers) a plain build finds set, and the agent that sets it, or ""
// for a person's terminal: a test build never counts one.
func agentSession() (marker, agent string) {
	if engine.Stamped() {
		return "", ""
	}
	return format.AgentMarker(os.Getenv)
}

// personNext is the next step of a person's command refused to an agent: the command, run by a person. agent is the
// one whose marker was found (Claude Code, pi, an agent).
func personNext(agent, command string) string {
	return "a person runs it in a terminal of their own, not one " + agent + " started: " + command
}

// personCommand refuses a person's command in an agent's session, or gives nil when it may run: no agent's session,
// or a person's grant in a project in command mode (guard.PersonGrant). shown is the command as typed, what says why
// it is a person's, and command is what the person runs.
func personCommand(dir, home, shown, what, command string) *engine.Error {
	marker, agent := agentSession()
	if marker == "" {
		return nil
	}
	why := guard.PersonGrant(home, dir, os.Getenv)
	if why == "" {
		return nil
	}
	return &engine.Error{Code: "person-command", Exit: engine.ExitState, Who: "person",
		What: shown + " is a person's command (" + what + "), and this is an agent's session (" + marker +
			" is set): an agent may run it only in a project in command mode, under a running task whose bonsai.allows a " +
			"person gave bonsai.yaml and .bonsai/lock.json, and " + engine.ASCII(why) + "; nothing was written",
		Next: personNext(agent, command)}
}

// The commands that change this machine's Bonsai home (spec section 4: "refused in an agent session, a tripwire,
// decision D"; plan-5 note 5.6.0, 2): settings set, labels attach and labels detach. Unlike init, update and unlink,
// no grant opens them: a machine's settings and attached definitions are a person's, whatever the project's mode.
// They call personOnly, the one check of the variable for them, so neither writes its own.
//
// The stamped build's confinement, carried from the project (engine.Confine) to the home: a test build (stamped with
// hookPlace and scratchRoot) skips the refusal only while Bonsai's home (BONSAI_HOME, else ~/.bonsai) lies inside its
// stamped scratch root, so scripted runs on scratch homes work; with the home outside it, a test build refuses
// (test-build), agent or not, so it never writes a real home's settings. A plain build never skips: in an agent's
// session (any agent marker set) it refuses (person-command), and its sentence names the marker and says where to
// type the command instead.

// personOnly refuses a command that changes this machine's Bonsai home when it may not run here, or gives nil:
// shown is the command as typed (bonsai settings set), what says what it changes, and command is the line a person
// types. It reads the agent markers, the home and the build's stamps as they are now.
func personOnly(shown, what, command string) *engine.Error {
	place, root := engine.TestBuild()
	home, err := workspace.Home()
	if err != nil {
		home = "" // a home that cannot be found is never inside a scratch root; the command's own run reports it
	}
	return personOnlyAt(place, root, home, os.Getenv, shown, what, command)
}

// personOnlyAt is personOnly with the stamps, the home and the variables given, for its tests.
func personOnlyAt(place, root, home string, getenv func(string) string, shown, what, command string) *engine.Error {
	if place != "" || root != "" {
		if place != "" && root != "" && home != "" && filepath.IsAbs(root) && engine.FolderHolds(root, home) {
			return nil
		}
		return &engine.Error{Code: "test-build", Exit: engine.ExitState, Who: "agent",
			What: shown + " is refused: this bonsai is a test build, and " + what + " only in a Bonsai home inside its " +
				"scratch folder " + engine.ASCII(filepath.ToSlash(root)) + ", which the home " +
				engine.ASCII(filepath.ToSlash(home)) + " is not in (or it was stamped with only one of hookPlace and " +
				"scratchRoot); nothing was written",
			Next: "set BONSAI_HOME to a folder inside " + engine.ASCII(filepath.ToSlash(root)) + ", or use the installed bonsai"}
	}
	marker, agent := format.AgentMarker(getenv)
	if marker == "" {
		return nil
	}
	inside := "not inside " + agent
	if marker == guard.ChildSessionEnv {
		inside = "not inside Claude (its ! lines count as an agent's session)"
	}
	return &engine.Error{Code: "person-command", Exit: engine.ExitState, Who: "person",
		What: shown + " is a person's command (" + what + "), and this is an agent's session (" + marker +
			" is set): type it in a terminal of your own, " + inside + "; nothing was written",
		Next: command}
}

// line is the call's command line again, for a person to type: the word, its arguments, then its flags as given,
// each quoted for a shell where it needs it.
func (c *call) line() string {
	out := "bonsai " + c.word.Name
	for _, a := range c.rest {
		out += " " + engine.ShellArg(a)
	}
	for _, g := range c.given {
		out += " " + g.name
		if g.value != "" {
			out += " " + engine.ShellArg(g.value)
		}
	}
	return out
}
