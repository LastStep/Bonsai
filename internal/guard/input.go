package guard

// Reading Claude Code's PreToolUse payload (spec §7: "one hook adapter reads Claude Code's payload (tool names,
// backslash paths on Windows, unknown fields). A payload it cannot read blocks."). The fields are those of Claude
// Code's hooks reference: session_id, transcript_path, cwd, permission_mode, hook_event_name, tool_name, tool_input,
// tool_use_id, and agent_id and agent_type inside a subagent. Fields the guard does not use are ignored.

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/LastStep/Bonsai/internal/schema"
)

// MaxInput caps the payload the guard reads, in bytes. A larger one is not read, and the call is blocked.
const MaxInput = 64 << 20

// Input is what the guard takes from the payload.
type Input struct {
	Session   string // session_id; "" when the payload has none
	Event     string // hook_event_name: always PreToolUse here
	Cwd       string // cwd: the session's current folder, against which a relative path is read; "" when absent
	Tool      string // tool_name
	ToolUseID string // tool_use_id; "" when absent
	AgentID   string // agent_id, inside a subagent; "" otherwise
	AgentType string // agent_type, inside a subagent; "" otherwise
	Path      string // for a file tool: tool_input's file_path (notebook_path for NotebookEdit); "" for other tools
}

// fileTools are the tools the rule judges, each with the tool_input field that names the file it writes. Part 3's
// hook line sends these four, and Bash and PowerShell (engine.GuardMatcher).
var fileTools = map[string]string{
	"Edit":         "file_path",
	"Write":        "file_path",
	"MultiEdit":    "file_path",
	"NotebookEdit": "notebook_path",
}

// sessionPattern is what a session id must look like to name a log file (s-<session>.ndjson): Claude Code's ids
// are UUIDs; anything with a separator or a dot-dot could reach outside the log folder, and is refused.
var sessionPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.-]{0,127}$`)

var bom = []byte{0xEF, 0xBB, 0xBF}

// ParseInput reads one PreToolUse payload. Its error says what is wrong in words that follow "the hook input".
func ParseInput(raw []byte) (*Input, error) {
	raw = bytes.TrimPrefix(raw, bom)
	if len(bytes.TrimSpace(raw)) == 0 {
		return nil, errors.New("is empty")
	}
	v, err := schema.Decode(raw)
	if err != nil {
		return nil, fmt.Errorf("is not one JSON document (%s)", oneLine(err.Error()))
	}
	o, ok := v.(schema.Object)
	if !ok {
		return nil, errors.New("is not a JSON object")
	}
	in := &Input{}
	var bad error
	text := func(o schema.Object, where, key string, required bool) string {
		if bad != nil {
			return ""
		}
		v, ok := o.Get(key)
		if !ok || v == nil {
			if required {
				bad = fmt.Errorf("has no %s%s", where, key)
			}
			return ""
		}
		s, ok := v.(string)
		if !ok {
			bad = fmt.Errorf("has a %s%s that is not a string", where, key)
			return ""
		}
		if strings.ContainsRune(s, 0) {
			bad = fmt.Errorf("has a %s%s holding a NUL character", where, key)
			return ""
		}
		return s
	}
	in.Event = text(o, "", "hook_event_name", true)
	in.Session = text(o, "", "session_id", false)
	in.Cwd = text(o, "", "cwd", false)
	in.Tool = text(o, "", "tool_name", true)
	in.ToolUseID = text(o, "", "tool_use_id", false)
	in.AgentID = text(o, "", "agent_id", false)
	in.AgentType = text(o, "", "agent_type", false)
	if bad != nil {
		return nil, bad
	}
	if in.Event != "PreToolUse" {
		return nil, fmt.Errorf("is a %s payload: the guard answers PreToolUse only", strconv.QuoteToASCII(in.Event))
	}
	if in.Tool == "" {
		return nil, errors.New("names no tool (an empty tool_name)")
	}
	if in.Session != "" && !sessionPattern.MatchString(in.Session) {
		return nil, fmt.Errorf("has a session_id %s that is not a session id", quote(in.Session, 60))
	}
	ti, ok := o.Get("tool_input")
	if !ok {
		return nil, errors.New("has no tool_input")
	}
	tin, ok := ti.(schema.Object)
	if !ok {
		return nil, errors.New("has a tool_input that is not a JSON object")
	}
	if field, ok := fileTools[in.Tool]; ok {
		in.Path = text(tin, "tool_input.", field, true)
		if bad != nil {
			return nil, bad
		}
		if in.Path == "" {
			return nil, fmt.Errorf("has an empty tool_input.%s", field)
		}
	}
	return in, nil
}

// quote gives a value as an ASCII Go string, cut to max runes.
func quote(s string, max int) string {
	r := []rune(s)
	if len(r) > max {
		s = string(r[:max]) + "..."
	}
	return strconv.QuoteToASCII(s)
}

// oneLine keeps a message on one ASCII line: printable ASCII as itself, a line break as a space, any other
// character as a Go escape, so it prints unbroken in PowerShell 5.1 (spec §3).
func oneLine(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r == '\n' || r == '\r' || r == '\t':
			b.WriteByte(' ')
		case r >= 0x20 && r < 0x7f:
			b.WriteRune(r)
		default:
			q := strconv.QuoteRuneToASCII(r)
			b.WriteString(q[1 : len(q)-1])
		}
	}
	return b.String()
}
