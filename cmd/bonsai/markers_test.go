package main

// Every agent's marker, not only Claude Code's, makes a session an agent's (T-5727; the director, 23 Oct, C4 and C8 of
// design/agent-neutral-walkthrough.md): with any one of the eight set, each other cleared, Bonsai's person-only
// commands refuse (person-command, exit 4, nothing written), naming the marker found and the agent that sets it.
// CLAUDECODE, left out on purpose, refuses nothing. The names are written out here, not read from format.AgentMarkers,
// so a name dropped from the list fails these tests.

import (
	"strings"
	"testing"
)

// markerAgents are the eight markers in the director's order, each with the agent a refusal names.
var markerAgents = []struct{ name, agent string }{
	{"CLAUDE_CODE_CHILD_SESSION", "Claude Code"}, {"AI_AGENT", "an agent"}, {"PI_CODING_AGENT", "pi"},
	{"CODEX_THREAD_ID", "Codex"}, {"GEMINI_CLI", "Gemini CLI"}, {"CURSOR_AGENT", "Cursor"}, {"OPENCODE", "OpenCode"},
	{"GOOSE_TERMINAL", "Goose"},
}

// onlyMarker makes the rest of a test a session where only name holds a value (value), every other marker and
// CLAUDECODE cleared.
func onlyMarker(t *testing.T, name, value string) {
	t.Helper()
	for _, m := range markerAgents {
		t.Setenv(m.name, "")
	}
	t.Setenv("CLAUDECODE", "")
	if name != "" {
		t.Setenv(name, value)
	}
}

// TestEveryMarkerRefusesPersonCommands: update --allow-exec --yes, unlink --yes, and an update that would record
// looser settings for the checks, each refused per marker, the refusal naming it and its agent; nothing written.
func TestEveryMarkerRefusesPersonCommands(t *testing.T) {
	for _, m := range markerAgents {
		t.Run(m.name, func(t *testing.T) {
			c := newCLI(t)
			c.link()
			c.commit("link")
			linked := c.read("bonsai.yaml")
			onlyMarker(t, m.name, "some-agent")
			before := tree(t, c.root)
			for _, args := range [][]string{{"update", "--allow-exec", "--yes"}, {"unlink", "--yes"}} {
				code, out, errOut := c.run("", args...)
				text := flat(out + errOut)
				if code != 4 || !strings.Contains(text, "this is an agent's session ("+m.name+" is set)") ||
					!strings.Contains(text, "a person runs it in a terminal of their own, not one "+m.agent+" started: bonsai ") {
					t.Errorf("%v with %s set: exit %d\n%s%s", args, m.name, code, out, errOut)
				}
			}
			lists := `".claude/**", "bonsai.yaml", ".bonsai/lock.json"`
			c.write("bonsai.yaml", strings.Replace(linked, "protected: ["+lists+"]", `protected: ["bonsai.yaml", ".bonsai/lock.json"]`, 1))
			before = tree(t, c.root)
			code, out, errOut := c.run("", "update", "--yes")
			if code != 4 || !strings.Contains(flat(out+errOut), "settings for the checks looser") ||
				!strings.Contains(flat(out+errOut), m.name+" is set") {
				t.Errorf("a looser update with %s set: exit %d\n%s%s", m.name, code, out, errOut)
			}
			if tree(t, c.root) != before {
				t.Errorf("a refused command wrote, %s set", m.name)
			}
		})
	}
}

// TestEveryMarkerRefusesHomeCommands: settings set and labels attach (personOnlyAt, a plain build) refused per marker,
// naming it.
func TestEveryMarkerRefusesHomeCommands(t *testing.T) {
	const shown, what, line = "bonsai settings set", "it changes this machine's settings", "bonsai settings set status_writes=agents"
	for _, m := range markerAgents {
		getenv := func(k string) string {
			if k == m.name {
				return "1"
			}
			return ""
		}
		e := personOnlyAt("", "", "/srv/bonsai-home", getenv, shown, what, line)
		if e == nil || e.Code != "person-command" || !strings.Contains(e.What, "("+m.name+" is set)") || e.Next != line {
			t.Errorf("%s set: %+v", m.name, e)
		}
	}
	if e := personOnlyAt("", "", "/srv/bonsai-home", func(k string) string {
		return map[string]string{"CLAUDECODE": "1"}[k]
	}, shown, what, line); e != nil {
		t.Errorf("CLAUDECODE alone: %+v", e)
	}
}

// TestClaudeCodeMarkerAloneRefusesNothing: CLAUDECODE, which reaches an editor's terminal too, is no agent's marker: a
// person's unlink runs with it set.
func TestClaudeCodeMarkerAloneRefusesNothing(t *testing.T) {
	c := newCLI(t)
	c.link()
	c.commit("link")
	onlyMarker(t, "CLAUDECODE", "1")
	if code, out, errOut := c.run("", "unlink", "--yes"); code != 0 {
		t.Errorf("unlink --yes with CLAUDECODE alone: exit %d\n%s%s", code, out, errOut)
	}
}

// TestEveryMarkerIsAskedNothing: no y/N question with any marker set (askable, which interactive reads before the
// terminals); CLAUDECODE alone, or none, leaves the question to the terminals.
func TestEveryMarkerIsAskedNothing(t *testing.T) {
	for _, m := range markerAgents {
		if askable(func(k string) string { return map[string]string{m.name: "1"}[k] }) {
			t.Errorf("%s set: askable", m.name)
		}
	}
	for _, env := range []map[string]string{{}, {"CLAUDECODE": "1"}, {"AI_AGENT": ""}} {
		if !askable(func(k string) string { return env[k] }) {
			t.Errorf("%v: not askable", env)
		}
	}
}
