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
// A skill's fork (measured on WSL, 15 Oct, a context: fork skill run by the session) has a file holding agentType and
// spawnDepth 1 and no toolUseId; a subagent's own subagent has parentAgentId, the id of the subagent that started it,
// and spawnDepth 2 (VG round 2's scripted sessions). SpawnInfo gives all four, so the guard joins a fork only to a call
// of its own caller (VG round 2's S1: a helper's fork took the session's stale Skill call).
//
// The file is Claude Code's internal layout, which it does not publish and may change in any release: a file that is
// not there gives an error that is fs.ErrNotExist, and the guard then joins the subagent to the oldest open call of its
// type in its session made a short while before (internal/guard); one that does not read, or names an id that is not
// one, gives another error, and the guard joins it to none. The path is built from the payload's transcript_path,
// session_id and agent_id alone (each id held to ID's rule, so none can leave the folder), and the file is read only
// when it is a regular file, capped at maxMeta bytes (agents.ReadRegular: a FIFO or a device there is named, never
// opened, so it stalls nothing).

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/LastStep/Bonsai/internal/agents"
)

// maxMeta caps the subagent meta file read, in bytes (the measured ones are about 200).
const maxMeta = 64 << 10

// SpawnMeta is what Claude Code's meta file beside a subagent's transcript says of its start (this file's comment).
type SpawnMeta struct {
	Call   string // the starting call's tool_use_id (toolUseId); "" where the file names none (a skill's fork)
	Type   string // its agent type (agentType)
	Parent string // the subagent that started it (parentAgentId); "" where the file names none
	Depth  int    // spawnDepth: 1 for one the top-level session started; 0 where the file gives none
}

// SpawnInfo reads the meta file of the subagent agentID (this file's comment): transcript is the payload's
// transcript_path (the session's transcript), session its session_id. Its error is fs.ErrNotExist when there is no
// file to read (none there, or no transcript, session or agent id to find it by).
func SpawnInfo(transcript, session, agentID string) (SpawnMeta, error) {
	if transcript == "" || ID(session) == "" || ID(agentID) == "" {
		return SpawnMeta{}, fmt.Errorf("no transcript, session or agent id to find it by: %w", fs.ErrNotExist)
	}
	dir := filepath.Dir(filepath.FromSlash(transcript))
	raw, err := agents.ReadRegular(filepath.Join(dir, session, "subagents", "agent-"+agentID+".meta.json"), maxMeta)
	if err != nil {
		return SpawnMeta{}, err
	}
	var m struct {
		ToolUseID *string `json:"toolUseId"`
		AgentType string  `json:"agentType"`
		Parent    *string `json:"parentAgentId"`
		Depth     int     `json:"spawnDepth"`
	}
	if err := json.Unmarshal(raw, &m); err != nil {
		return SpawnMeta{}, err
	}
	out := SpawnMeta{Type: m.AgentType, Depth: m.Depth}
	if m.ToolUseID != nil {
		if out.Call = ID(*m.ToolUseID); out.Call == "" {
			return SpawnMeta{}, errors.New("the meta file's toolUseId is not an id")
		}
	}
	if m.Parent != nil {
		if out.Parent = ID(*m.Parent); out.Parent == "" {
			return SpawnMeta{}, errors.New("the meta file's parentAgentId is not an id")
		}
	}
	return out, nil
}
