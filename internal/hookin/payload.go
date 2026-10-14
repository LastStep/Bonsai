package hookin

// Claude Code's hook payload, read once (Claude Code's hooks reference, read on Claude Code 2.1.296: "Common input
// fields", each event's input, PreToolUse's tool inputs; the Agent SDK's tool input types for Monitor and NotebookEdit;
// MultiEdit, which those pages no longer list, as earlier versions sent it: file_path and its edits).
//
// One JSON object, a leading byte order mark allowed, at most MaxInput bytes. Its top level is read member by member:
//   - A key Bonsai reads (fields, below) has its value read strictly, with the guard's rules: valid UTF-8, no key twice
//     at any depth inside it (tool_input is decoded whole by the strict decoder, schema.Decode), and no NUL in a string
//     that names something (an id, the event, the tool, a path, a command). Text a tool writes (Write's content, Edit's
//     strings) and a person's words (a prompt, a notice) may hold a NUL, as a file may. A value of the wrong type is a
//     fault, read as absent; null is absent, no fault.
//   - Any other key, tool_response and the fields of later versions among them, is passed over unread, held only to
//     being valid JSON, as Go's decoder holds a value: so a payload's unread parts cost nothing but their scan.
//   - A key twice at the top level is a fault, and its last value is the one read (as Go's decoder keeps it).
//
// Then each event's own checks: a tool event (ToolEvent) must name its tool and carry tool_input as an object, a file
// tool its path, Bash and PowerShell their command. An unknown event, or an unknown tool, is passed on by name with
// the common fields read. The spawn tools' fields are read too (step 5.5.11: Agent and its alias Task, SendMessage,
// Workflow), and Read's path, each where a guard of roles finds what it judges (SpawnTool).
//
// A reader that judges only some tools (ReadJudged: the guard, whose second line meets every tool, plan-5 5.5.11
// note 2) has tool_input passed over unread for every other tool, held only to being valid JSON, as the payload's other
// keys are: so a tool Claude Code adds later, or one whose input the strict decoder would refuse (a key twice in an
// MCP tool's arguments), is counted, not blocked. Every other reader (the recorder) reads tool_input as before.
//
// A fault leaves the payload read: each field as Go's lenient decoder reads it (a wrong type absent, bad UTF-8 as
// U+FFFD), which is how bonsai hook record has read payloads since step 5.2.4, so it writes every record it wrote
// then. The guard and the stop gate block on any fault (Payload.Fault).

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/LastStep/Bonsai/internal/schema"
)

// Payload is one hook payload as Bonsai reads it. A string is "" when its field is absent, null or of the wrong type.
type Payload struct {
	// The common fields.
	Session        string // session_id, held to the log file's rule (ID): "" when absent, or when it cannot name a log file (a fault)
	TranscriptPath string // transcript_path
	Cwd            string // cwd: the session's current folder when the hook ran (an agent's cd moves it)
	PermissionMode string // permission_mode
	Event          string // hook_event_name: any event, known or not
	AgentID        string // agent_id, inside a subagent, as written (the recorder holds it to ID)
	AgentType      string // agent_type, inside a subagent or a session started with --agent

	// A tool event's (ToolEvent): PreToolUse, PermissionRequest, PostToolUse, PostToolUseFailure.
	Tool        string // tool_name: an unknown tool is passed on by name
	ToolUseID   string // tool_use_id, as written (PermissionRequest has none)
	Input       any    // tool_input decoded whole by the strict decoder (an Object for every tool); nil when absent, null or refused
	Path        string // a file tool's path as written, backslashes and all (FileTool: file_path, NotebookEdit's notebook_path)
	Command     string // the command of Bash, PowerShell and Monitor (a Monitor watching a WebSocket has none)
	Content     string // Write's content: what it would write
	Edits       []Edit // Edit's one replacement, or MultiEdit's edits: what they would replace
	IsInterrupt bool   // PostToolUseFailure's is_interrupt

	// The spawn tools' fields and Read's path (SpawnTool; step 5.5.11), each "" when absent. Of these only
	// subagent_type and to are held to being a string (a fault otherwise), since a guard of roles judges by them; a
	// prompt, a message, a script and a path to read are only searched, so a value of another type reads as "".
	SubagentType string // Agent and Task: tool_input.subagent_type, as written ("" for Claude Code's general-purpose agent)
	AgentName    string // Agent and Task: tool_input.name (no field of Claude Code 2.1.300's Agent tool; read when given)
	AgentPrompt  string // Agent and Task: tool_input.prompt, where a brief's id is named
	To           string // SendMessage: tool_input.to (an agent's id, a name it was given, main, or another session)
	MessageText  string // SendMessage: tool_input.message, where a newer brief's id is named
	Script       string // Workflow: tool_input.script
	ReadPath     string // Read: tool_input.file_path, as written

	// InputPassed: tool_input was passed over unread, as a reader that does not judge its tool asked (ReadJudged); Input
	// is then nil, with no fault for its shape.
	InputPassed bool

	// The other events' own.
	Prompt              string // UserPromptSubmit's prompt
	Source              string // SessionStart's source: startup, resume, clear, compact or fork
	Model               string // SessionStart's model: a string, or an object's id
	NotificationType    string // Notification's notification_type
	Message             string // Notification's message
	StopHookActive      bool   // Stop's and SubagentStop's stop_hook_active
	AgentTranscriptPath string // SubagentStop's agent_transcript_path
	Reason              string // SessionEnd's reason

	// Fault is the first value read that broke a rule (above), in ASCII words that follow "the hook input"; nil when
	// every value read cleanly. A payload with a fault is still read.
	Fault error
}

// Edit is one replacement an Edit or MultiEdit call would make.
type Edit struct {
	Old        string // old_string
	New        string // new_string
	ReplaceAll bool   // replace_all
}

// InputObject gives tool_input as an object (redact.Target, redact.Question and the guard read it so), or nil.
func (p *Payload) InputObject() schema.Object {
	o, _ := p.Input.(schema.Object)
	return o
}

// fields are the keys of a payload's top level that Bonsai reads, each with where its value goes: their one home.
// Every other key is passed over unread.
var fields = []struct {
	key  string
	read func(p *Payload, key string, raw []byte)
}{
	{"session_id", func(p *Payload, k string, raw []byte) { p.Session = p.session(k, raw) }},
	{"transcript_path", func(p *Payload, k string, raw []byte) { p.TranscriptPath = p.text(k, raw, true) }},
	{"cwd", func(p *Payload, k string, raw []byte) { p.Cwd = p.text(k, raw, true) }},
	{"permission_mode", func(p *Payload, k string, raw []byte) { p.PermissionMode = p.text(k, raw, true) }},
	{"hook_event_name", func(p *Payload, k string, raw []byte) { p.Event = p.text(k, raw, true) }},
	{"agent_id", func(p *Payload, k string, raw []byte) { p.AgentID = p.text(k, raw, true) }},
	{"agent_type", func(p *Payload, k string, raw []byte) { p.AgentType = p.text(k, raw, true) }},
	{"tool_name", func(p *Payload, k string, raw []byte) { p.Tool = p.text(k, raw, true) }},
	{"tool_use_id", func(p *Payload, k string, raw []byte) { p.ToolUseID = p.text(k, raw, true) }},
	{"tool_input", func(p *Payload, k string, raw []byte) { p.Input = p.input(k, raw) }},
	{"is_interrupt", func(p *Payload, k string, raw []byte) { p.IsInterrupt = p.flag(k, raw) }},
	{"prompt", func(p *Payload, k string, raw []byte) { p.Prompt = p.text(k, raw, false) }},
	{"source", func(p *Payload, k string, raw []byte) { p.Source = p.text(k, raw, true) }},
	{"model", func(p *Payload, k string, raw []byte) { p.Model = p.model(k, raw) }},
	{"notification_type", func(p *Payload, k string, raw []byte) { p.NotificationType = p.text(k, raw, true) }},
	{"message", func(p *Payload, k string, raw []byte) { p.Message = p.text(k, raw, false) }},
	{"stop_hook_active", func(p *Payload, k string, raw []byte) { p.StopHookActive = p.flag(k, raw) }},
	{"agent_transcript_path", func(p *Payload, k string, raw []byte) { p.AgentTranscriptPath = p.text(k, raw, true) }},
	{"reason", func(p *Payload, k string, raw []byte) { p.Reason = p.text(k, raw, true) }},
}

// ToolEvent reports the four events that describe one tool call, each carrying tool_name and tool_input.
func ToolEvent(event string) bool {
	switch event {
	case "PreToolUse", "PermissionRequest", "PostToolUse", "PostToolUseFailure":
		return true
	}
	return false
}

// pathKey names the tool_input key holding the file a file tool writes: "" for any other tool.
func pathKey(tool string) string {
	switch tool {
	case "Edit", "MultiEdit", "Write":
		return "file_path"
	case "NotebookEdit":
		return "notebook_path"
	}
	return ""
}

// FileTool reports a tool that writes the file its input names: Edit, MultiEdit, Write and NotebookEdit.
func FileTool(tool string) bool { return pathKey(tool) != "" }

var bom = []byte{0xEF, 0xBB, 0xBF}

// Parse reads one payload (this file's comment). Its error, for one that is not a single JSON object, says what is
// wrong in ASCII words that follow "the hook input"; a payload it reads with a fault comes back with Fault set.
func Parse(raw []byte) (*Payload, error) { return parsePayload(raw, nil) }

// ParseJudged is Parse for a reader that judges only the tools judged reports true for: every other tool's tool_input
// is passed over, held only to being valid JSON (this file's comment). judged nil reads every tool's, as Parse.
func ParseJudged(raw []byte, judged func(tool string) bool) (*Payload, error) {
	return parsePayload(raw, judged)
}

func parsePayload(raw []byte, judged func(tool string) bool) (*Payload, error) {
	raw = bytes.TrimPrefix(raw, bom)
	if len(bytes.TrimSpace(raw)) == 0 {
		return nil, errors.New("is empty")
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	tok, err := dec.Token()
	if err != nil {
		return nil, notOne(err)
	}
	if d, ok := tok.(json.Delim); !ok || d != '{' {
		return nil, errors.New("is not a JSON object")
	}
	p := &Payload{}
	seen := make(map[string]bool, 16)
	var input json.RawMessage // tool_input, kept to read once the tool is known, for a reader that judges some tools only
	for dec.More() {
		kt, err := dec.Token()
		if err != nil {
			return nil, notOne(err)
		}
		key, _ := kt.(string)
		if seen[key] {
			p.fault("has the key %s twice", quoted(key, 60))
		}
		seen[key] = true
		read := field(key)
		if read == nil {
			// Passed over: the decoder scans it as valid JSON, and nothing of it is kept.
			if err := dec.Decode(&skipped{}); err != nil {
				return nil, notOne(err)
			}
			continue
		}
		var value json.RawMessage
		if err := dec.Decode(&value); err != nil {
			return nil, notOne(err)
		}
		if key == "tool_input" && judged != nil {
			input = value // its tool may come after it
			continue
		}
		read(p, key, value)
	}
	if _, err := dec.Token(); err != nil {
		return nil, notOne(err)
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return nil, errors.New("is not one JSON document (something follows it)")
	}
	if p.Event == "" {
		p.fault("has no hook_event_name")
	}
	switch {
	case judged != nil && !judged(p.Tool):
		p.InputPassed = true // absent or present, of any shape: not this reader's to read
	case input != nil:
		p.Input = p.input("tool_input", input)
	}
	if ToolEvent(p.Event) {
		p.tool(seen["tool_name"])
	}
	return p, nil
}

// field gives how a key Bonsai reads is read, or nil for a key it passes over.
func field(key string) func(p *Payload, key string, raw []byte) {
	for _, f := range fields {
		if f.key == key {
			return f.read
		}
	}
	return nil
}

// skipped is a value passed over unread.
type skipped struct{}

func (*skipped) UnmarshalJSON([]byte) error { return nil }

// notOne is the error of a payload that is not one JSON document.
func notOne(err error) error {
	if errors.Is(err, io.EOF) {
		err = io.ErrUnexpectedEOF
	}
	return errors.New("is not one JSON document (" + ascii(err.Error()) + ")")
}

// fault keeps the first fault met.
func (p *Payload) fault(format string, a ...any) {
	if p.Fault == nil {
		p.Fault = errors.New(ascii(fmt.Sprintf(format, a...)))
	}
}

// text reads a string: as Go's decoder reads it (wrong type or null: ""), with a fault for a wrong type, bad UTF-8,
// or, when names is true, a NUL.
func (p *Payload) text(key string, raw []byte, names bool) string {
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		p.fault("has a %s that is not a string", key)
		return ""
	}
	if !utf8.Valid(raw) {
		p.fault("has a %s that is not valid UTF-8", key)
	}
	if names && strings.IndexByte(s, 0) >= 0 {
		p.fault("has a %s holding a NUL character", key)
	}
	return s
}

// flag reads true or false (null: false), with a fault for any other value.
func (p *Payload) flag(key string, raw []byte) bool {
	var b bool
	if err := json.Unmarshal(raw, &b); err != nil {
		p.fault("has a %s that is not true or false", key)
		return false
	}
	return b
}

// session reads session_id, held to the log file's rule.
func (p *Payload) session(key string, raw []byte) string {
	s := p.text(key, raw, true)
	if s != "" && ID(s) == "" {
		p.fault("has a session_id %s that cannot name a log file (letters, digits, _ and -, at most 128)", quoted(s, 60))
		return ""
	}
	return s
}

// model reads SessionStart's model: a string, or an object holding the id (earlier versions sent one), read as Go's
// decoder reads it; a fault for any other value, or an object that does not read strictly.
func (p *Payload) model(key string, raw []byte) string {
	t := bytes.TrimSpace(raw)
	if len(t) == 0 || t[0] != '{' {
		if len(t) > 0 && t[0] == '"' || bytes.Equal(t, []byte("null")) {
			return p.text(key, raw, true)
		}
		p.fault("has a %s that is neither a string nor an object", key)
		return ""
	}
	if _, err := schema.Decode(raw); err != nil {
		p.fault("has a %s that does not read (%s)", key, err.Error())
	}
	var o struct {
		ID string `json:"id"`
	}
	if json.Unmarshal(raw, &o) != nil {
		p.fault("has a %s whose id is not a string", key)
		return ""
	}
	if strings.IndexByte(o.ID, 0) >= 0 {
		p.fault("has a %s holding a NUL character", key)
	}
	return o.ID
}

// input reads tool_input whole, by the strict decoder: nil for null, or for a value it refuses (a fault).
func (p *Payload) input(key string, raw []byte) any {
	v, err := schema.Decode(raw)
	if err != nil {
		p.fault("has a %s that does not read (%s)", key, err.Error())
		return nil
	}
	return v
}

// tool reads a tool event's own: the tool, its input as an object, and the input's fields Bonsai reads for its tool.
// named is whether the payload has the key tool_name.
func (p *Payload) tool(named bool) {
	switch {
	case !named:
		p.fault("has no tool_name")
	case p.Tool == "":
		p.fault("names no tool (tool_name is empty or null)")
	}
	if p.InputPassed {
		return
	}
	if p.Input == nil {
		p.fault("has no tool_input")
		return
	}
	in, ok := p.Input.(schema.Object)
	if !ok {
		p.fault("has a tool_input that is not a JSON object")
		return
	}
	if k := pathKey(p.Tool); k != "" {
		p.Path = p.inText(in, k, true, true)
		if p.Path == "" {
			p.fault("has an empty tool_input.%s", k)
		}
	}
	switch p.Tool {
	case "Bash", "PowerShell":
		p.Command = p.inText(in, "command", true, true)
	case "Monitor":
		p.Command = p.inText(in, "command", false, true)
	case "Write":
		p.Content = p.inText(in, "content", false, false)
	case "Edit":
		p.Edits = []Edit{p.edit(in, "tool_input.")}
	case "MultiEdit":
		p.Edits = p.edits(in)
	case "Agent", "Task":
		p.SubagentType = p.inText(in, "subagent_type", false, true)
		p.AgentName = p.inText(in, "name", false, true)
		p.AgentPrompt = looseText(in, "prompt")
	case "SendMessage":
		p.To = p.inText(in, "to", false, true)
		p.MessageText = looseText(in, "message")
	case "Workflow":
		p.Script = looseText(in, "script")
	case "Read":
		p.ReadPath = looseText(in, "file_path")
	}
}

// SpawnTool reports the tools that start or resume an agent (step 5.5.11): Agent, its alias Task, SendMessage (which
// resumes one) and Workflow (whose script starts its own).
func SpawnTool(tool string) bool {
	switch tool {
	case "Agent", "Task", "SendMessage", "Workflow":
		return true
	}
	return false
}

// looseText reads a string of tool_input that is only searched: any other type, or absent, is "", with no fault.
func looseText(in schema.Object, key string) string {
	v, _ := in.Get(key)
	s, _ := v.(string)
	return s
}

// inText reads a string of tool_input: absent or null is "" (a fault when required), any other type a fault; a NUL is
// a fault when names is true.
func (p *Payload) inText(in schema.Object, key string, required, names bool) string {
	v, ok := in.Get(key)
	if !ok || v == nil {
		if required {
			p.fault("has no tool_input.%s", key)
		}
		return ""
	}
	s, ok := v.(string)
	if !ok {
		p.fault("has a tool_input.%s that is not a string", key)
		return ""
	}
	if names && strings.IndexByte(s, 0) >= 0 {
		p.fault("has a tool_input.%s holding a NUL character", key)
	}
	return s
}

// edit reads one replacement: old_string, new_string and replace_all, each optional; where names its object for a
// fault ("tool_input." or "tool_input.edits[2].").
func (p *Payload) edit(o schema.Object, where string) Edit {
	var e Edit
	for _, f := range []struct {
		key string
		to  *string
	}{{"old_string", &e.Old}, {"new_string", &e.New}} {
		v, _ := o.Get(f.key)
		switch s := v.(type) {
		case nil:
		case string:
			*f.to = s
		default:
			p.fault("has a %s%s that is not a string", where, f.key)
		}
	}
	v, _ := o.Get("replace_all")
	switch b := v.(type) {
	case nil:
	case bool:
		e.ReplaceAll = b
	default:
		p.fault("has a %sreplace_all that is not true or false", where)
	}
	return e
}

// edits reads MultiEdit's edits: a list of replacements (absent or null: none).
func (p *Payload) edits(in schema.Object) []Edit {
	v, _ := in.Get("edits")
	if v == nil {
		return nil
	}
	list, ok := v.([]any)
	if !ok {
		p.fault("has a tool_input.edits that is not a list")
		return nil
	}
	out := make([]Edit, 0, len(list))
	for i, x := range list {
		o, ok := x.(schema.Object)
		if !ok {
			p.fault("has a tool_input.edits[%d] that is not a JSON object", i)
			continue
		}
		out = append(out, p.edit(o, fmt.Sprintf("tool_input.edits[%d].", i)))
	}
	return out
}
