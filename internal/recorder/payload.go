package recorder

// Claude Code's hook payload, as the recorder reads it (Claude Code's hooks reference, read on Claude Code 2.1.296:
// the common fields session_id, transcript_path, cwd, permission_mode and hook_event_name, agent_id and agent_type
// inside a subagent, and each event's own). Read leniently: a field of the wrong type reads as absent, a field it does
// not use is skipped unread, and the payload is one JSON object or nothing. It must name an event and a session id
// that can name a log file (record.LogFileName's rule): without one there is nothing to record.

import (
	"bytes"
	"encoding/json"
	"errors"
	"regexp"
)

// Payload is what the recorder takes from a hook payload.
type Payload struct {
	Event            string          // hook_event_name
	Session          string          // session_id
	Cwd              string          // cwd: the session's current folder
	AgentID          string          // agent_id, inside a subagent
	AgentType        string          // agent_type, inside a subagent or a session started as an agent
	Tool             string          // tool_name: the tool events
	ToolInput        json.RawMessage // tool_input, as sent: the tool events; nil when absent
	ToolUseID        string          // tool_use_id: PreToolUse, PostToolUse, PostToolUseFailure
	IsInterrupt      bool            // is_interrupt: PostToolUseFailure
	Prompt           string          // prompt: UserPromptSubmit (only its kind is kept)
	NotificationType string          // notification_type: Notification
	Message          string          // message: Notification
	Source           string          // source: SessionStart
	Model            string          // model: SessionStart (a string, or an object's id)
	Reason           string          // reason: SessionEnd
}

// idPattern is what a session, subagent or tool-use id must be to be kept: letters, digits, _ and -, at most 128
// (record.LogFileName's rule for a session; Claude Code's ids are UUIDs, short hex and toolu_ ids).
var idPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,128}$`)

// ParsePayload reads one hook payload. Its error says why there is nothing to record.
func ParsePayload(raw []byte) (*Payload, error) {
	raw = bytes.TrimPrefix(raw, []byte{0xEF, 0xBB, 0xBF})
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil || m == nil {
		return nil, errors.New("not one JSON object")
	}
	text := func(key string) string {
		var s string
		if v, ok := m[key]; ok && json.Unmarshal(v, &s) == nil {
			return s
		}
		return ""
	}
	p := &Payload{
		Event:            text("hook_event_name"),
		Session:          text("session_id"),
		Cwd:              text("cwd"),
		AgentID:          text("agent_id"),
		AgentType:        text("agent_type"),
		Tool:             text("tool_name"),
		ToolUseID:        text("tool_use_id"),
		Prompt:           text("prompt"),
		NotificationType: text("notification_type"),
		Message:          text("message"),
		Source:           text("source"),
		Model:            text("model"),
		Reason:           text("reason"),
	}
	if v, ok := m["tool_input"]; ok && !bytes.Equal(bytes.TrimSpace(v), []byte("null")) {
		p.ToolInput = v
	}
	var b bool
	if v, ok := m["is_interrupt"]; ok && json.Unmarshal(v, &b) == nil {
		p.IsInterrupt = b
	}
	if p.Model == "" {
		var o struct {
			ID string `json:"id"`
		}
		if v, ok := m["model"]; ok && json.Unmarshal(v, &o) == nil {
			p.Model = o.ID
		}
	}
	switch {
	case p.Event == "":
		return nil, errors.New("names no hook_event_name")
	case !idPattern.MatchString(p.Session):
		return nil, errors.New("has no session_id that can name a log file")
	}
	return p, nil
}

// id gives an id as kept: itself when it matches idPattern, else "" (null).
func id(s string) string {
	if idPattern.MatchString(s) {
		return s
	}
	return ""
}
