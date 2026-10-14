package hookin

// Which Agent call started a subagent (design/plan-5.md, 5.5.11 notes 1 and 5): Claude Code's own file beside the
// subagent's transcript, read at the subagent's first tool call. Measured on Claude Code 2.1.300 (WSL, 15 Oct; the run
// report's payload table and its parallel-calls measurement): no payload ties a new subagent's id to the Agent call
// that started it before that call's PostToolUse, which for a foreground call comes only after the subagent has
// finished; SubagentStart carries the subagent's id and type but no tool_use_id. Claude Code writes
// <the session's transcript folder>/<session id>/subagents/agent-<agent id>.meta.json, holding toolUseId (the starting
// call's tool_use_id) and agentType, after SubagentStart and before the subagent's first PreToolUse: absent at 36 of
// 36 SubagentStarts, present with the right toolUseId at 36 of 36 first calls, three Agent calls side by side in one
// message among them. An agent a Workflow script starts has its file under subagents/workflows/<run>/ with no
// toolUseId; Claude Code's own internal agents have none read here (they make no tool call).
//
// The file is Claude Code's internal layout, which it does not publish and may change in any release: a file that is
// not there, does not read, or names no call gives "", and the guard then joins the subagent to the oldest open Agent
// call of its type in its session (internal/guard). The path is built from the payload's transcript_path, session_id
// and agent_id alone (each id held to ID's rule, so none can leave the folder), and the file is read capped at
// maxMeta bytes.

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
)

// maxMeta caps the subagent meta file read, in bytes (the measured ones are about 200).
const maxMeta = 64 << 10

// SpawnCall gives the tool_use_id of the call that started the subagent agentID, and its agent type, as Claude Code's
// meta file beside its transcript holds them (this file's comment); "" and an error when it cannot be told.
// transcript is the payload's transcript_path (the session's transcript), session its session_id.
func SpawnCall(transcript, session, agentID string) (toolUseID, agentType string, err error) {
	if transcript == "" || ID(session) == "" || ID(agentID) == "" {
		return "", "", errors.New("no transcript, session or agent id to find it by")
	}
	dir := filepath.Dir(filepath.FromSlash(transcript))
	p := filepath.Join(dir, session, "subagents", "agent-"+agentID+".meta.json")
	f, err := os.Open(p)
	if err != nil {
		return "", "", err
	}
	defer func() { _ = f.Close() }()
	raw, err := io.ReadAll(io.LimitReader(f, maxMeta+1))
	if err != nil {
		return "", "", err
	}
	if len(raw) > maxMeta {
		return "", "", errors.New("the meta file is too large")
	}
	var m struct {
		ToolUseID string `json:"toolUseId"`
		AgentType string `json:"agentType"`
	}
	if err := json.Unmarshal(raw, &m); err != nil {
		return "", "", err
	}
	if ID(m.ToolUseID) == "" {
		return "", "", errors.New("the meta file names no call")
	}
	return m.ToolUseID, m.AgentType, nil
}
