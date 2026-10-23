package format

import "testing"

// TestAgentMarker: the first marker holding a value, in the list's order, with its agent; an empty value is unset;
// CLAUDECODE is none of them.
func TestAgentMarker(t *testing.T) {
	want := []string{"CLAUDE_CODE_CHILD_SESSION", "AI_AGENT", "PI_CODING_AGENT", "CODEX_THREAD_ID", "GEMINI_CLI",
		"CURSOR_AGENT", "OPENCODE", "GOOSE_TERMINAL"}
	if len(AgentMarkers) != len(want) {
		t.Fatalf("AgentMarkers: %v", AgentMarkers)
	}
	for k, m := range AgentMarkers {
		if m.Word != want[k] || m.Means == "" {
			t.Errorf("AgentMarkers[%d]: %+v, want %s", k, m, want[k])
		}
	}
	env := map[string]string{"CLAUDECODE": "1", "AI_AGENT": "", "OPENCODE": "1", "PI_CODING_AGENT": "x"}
	if name, agent := AgentMarker(func(k string) string { return env[k] }); name != "PI_CODING_AGENT" || agent != "pi" {
		t.Errorf("AgentMarker: %s %s", name, agent)
	}
	if name, _ := AgentMarker(func(k string) string { return map[string]string{"CLAUDECODE": "1"}[k] }); name != "" {
		t.Errorf("CLAUDECODE alone: %s", name)
	}
}
