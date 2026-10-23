package format

// The agent markers (the director, 23 Oct, C4 and C8 of design/agent-neutral-walkthrough.md; T-5727): the variables
// by which Bonsai's commands tell an agent's session from a person's terminal. Any of them holding a value (any value:
// an empty one counts as unset) makes a session an agent's, and Bonsai's person-only commands refuse it
// (person-command): init or update with --allow-exec, unlink, an update that changes Bonsai's own lines, records
// another base branch or looser settings for the checks, settings set, labels attach and detach; and no y/N question
// is asked. Each agent sets its marker in every process its shell tool starts, so scripts and subprocesses inherit it.
//
// This list is their one home: cmd/bonsai's refusals, the y/N question, the guard's watch on a shell line that unsets
// or assigns one (internal/guard's shell_env.go) and the test helper that runs a test as a person
// (testpack.Isolate) all read it, and the person-command word's meaning names it, so the reference page shows it. A
// name is added here with its agent, and the reference page regenerated in the same commit.
//
// Left out on purpose: CLAUDECODE, which Claude Code also sets but which reaches the terminal of an editor that hosts
// it, where a person types; and the agents that document no marker (Copilot CLI, Droid, Kiro, Devin, Aider, Amp's
// unconfirmed AGENT=amp), not covered until one is confirmed. A marker is a tripwire for the honest and the accidental
// case: an agent that clears it in a shell no guard reads gets past it.
//
// Word is the variable's name and Means the agent that sets it, as the refusals name it ("not one <Means> started").
var AgentMarkers = []Word{
	{Word: "CLAUDE_CODE_CHILD_SESSION", Means: "Claude Code"},
	{Word: "AI_AGENT", Means: "an agent"}, // Claude Code, pi and OpenCode set it, each to its own name
	{Word: "PI_CODING_AGENT", Means: "pi"},
	{Word: "CODEX_THREAD_ID", Means: "Codex"},
	{Word: "GEMINI_CLI", Means: "Gemini CLI"},
	{Word: "CURSOR_AGENT", Means: "Cursor"},
	{Word: "OPENCODE", Means: "OpenCode"},
	{Word: "GOOSE_TERMINAL", Means: "Goose"},
}

// AgentMarker gives the first of AgentMarkers, in the list's order, that holds a value in getenv, and the agent that
// sets it; "", "" when none does: a person's terminal.
func AgentMarker(getenv func(string) string) (name, agent string) {
	for _, m := range AgentMarkers {
		if getenv(m.Word) != "" {
			return m.Word, m.Means
		}
	}
	return "", ""
}

// agentMarkerNames lists AgentMarkers for the person-command word's meaning.
func agentMarkerNames() string {
	out := ""
	for i, m := range AgentMarkers {
		if i > 0 {
			out += ", "
		}
		out += m.Word
	}
	return out
}
