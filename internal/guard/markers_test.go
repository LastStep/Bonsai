package guard

// The guard's watch on a shell line that unsets or assigns an agent marker (T-5727; the director, 23 Oct, C4 and C8
// of design/agent-neutral-walkthrough.md, the watch's choice C): every exact form the watch reads for Claude Code's
// CLAUDE_CODE_CHILD_SESSION refuses for each of the other agents' markers too, its reason naming the marker; the
// catch-all text match (the backstop, and a name held as data) stays Claude Code's alone, so ordinary lines that hold
// another marker's name as part of a word pass. The names are written out here, not read from format.AgentMarkers,
// so a name dropped from the list fails these tests.

import (
	"strings"
	"testing"
)

// otherMarkers are the agent markers beside Claude Code's own, in the director's order.
var otherMarkers = []string{"AI_AGENT", "PI_CODING_AGENT", "CODEX_THREAD_ID", "GEMINI_CLI", "CURSOR_AGENT", "OPENCODE", "GOOSE_TERMINAL"}

// TestEveryMarkerWatched: each exact form, per marker, is a person's step (person-command) naming the marker; under
// a person's grant in command mode, allowed as Claude Code's is.
func TestEveryMarkerWatched(t *testing.T) {
	dir := withTask(t, "running", "bonsai.yaml", ".bonsai/lock.json")
	for _, m := range otherMarkers {
		lower := strings.ToLower(m)
		for _, c := range []struct{ tool, line string }{
			{"Bash", "unset " + m + "; bonsai update --yes --allow-exec"},
			{"Bash", m + "= bonsai unlink --yes"},
			{"Bash", m + "+=x ./link.sh"},
			{"Bash", "env -u " + m + " bonsai unlink --yes"},
			{"Bash", "env --unset=" + m + " ./link.sh"},
			{"Bash", "env " + m + "= ./link.sh"},
			{"Bash", "export " + m + "="},
			{"Bash", "export " + lower},
			{"Bash", "declare -x " + m + "=''"},
			{"Bash", "bash -c 'unset " + m + "'"},
			{"PowerShell", "$env:" + m + " = ''"},
			{"PowerShell", "${env:" + lower + "}=$null"},
			{"PowerShell", "Remove-Item Env:" + m},
			{"PowerShell", `Remove-Item Env:\` + m},
			{"PowerShell", "Set-Item -Path Microsoft.PowerShell.Core\\Environment::" + m + " -Value ''"},
			{"PowerShell", "[Environment]::SetEnvironmentVariable('" + m + "', $null)"},
		} {
			t.Run(c.tool+" "+c.line, func(t *testing.T) {
				stderr := wantRule(t, dir, bashIn(dir, c.tool, c.line), RulePersonCommand)
				if !strings.Contains(stderr, m) {
					t.Errorf("the refusal does not name %s: %s", m, stderr)
				}
			})
		}
	}
	commandMode(t, dir)
	wantRule(t, dir, bashIn(dir, "Bash", "unset PI_CODING_AGENT"), RuleShell)
}

// TestOtherMarkersNoWolf: a read of another agent's marker, its name inside a longer word or held as data, and
// CLAUDECODE (left out on purpose) are no person's step.
func TestOtherMarkersNoWolf(t *testing.T) {
	dir := withTask(t, "todo")
	for _, c := range []struct{ tool, line string }{
		{"Bash", "mkdir -p opencode-tests && cd opencode-tests"},
		{"Bash", "cd opencode-tests"},
		{"Bash", "echo $AI_AGENT ${PI_CODING_AGENT}"},
		{"Bash", "printenv GEMINI_CLI"},
		{"Bash", "grep -rn CURSOR_AGENT internal/"},
		{"Bash", "env | grep -E '^(AI_AGENT|PI_CODING_AGENT|OPENCODE)='"},
		{"Bash", "X=opencode ./x.sh"},
		{"Bash", "export OPENCODE_HOME=/tmp/x"},
		{"Bash", "unset CLAUDECODE"},
		{"Bash", "env -u CLAUDECODE ./x.sh"},
		{"Bash", "git commit -m 'guard: reads GOOSE_TERMINAL'"},
		{"PowerShell", "Write-Output $env:CODEX_THREAD_ID"},
		{"PowerShell", "Get-ChildItem Env:OPENCODE"},
		{"PowerShell", "$env:OPENCODE_HOME = 'x'"},
	} {
		t.Run(c.tool+" "+c.line, func(t *testing.T) { wantRule(t, dir, bashIn(dir, c.tool, c.line), RuleShell) })
	}
}
