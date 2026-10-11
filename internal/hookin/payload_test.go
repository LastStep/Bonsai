package hookin

// Payloads written fresh for every event and tool Bonsai reads, from Claude Code's hooks reference on the version in
// use (2.1.296: "Common input fields", each event's input and its example, PreToolUse's tool inputs; the tools
// reference for Monitor and NotebookEdit), with made-up values and the fields of later versions Bonsai passes over.

import (
	"bytes"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/LastStep/Bonsai/internal/schema"
)

// common is the start of a payload: the common fields, as the reference's examples give them, and the fields of
// later versions (prompt_id, scratchpad_dir, effort) Bonsai passes over.
const common = `"session_id":"0b6f3c2e-5d1a-4e8b-9c7f-2a4d6e8f0a1b","prompt_id":"550e8400-e29b-41d4-a716-446655440000",` +
	`"transcript_path":"/home/someone/.claude/projects/p/0b6f3c2e.jsonl","cwd":"/home/someone/proj",` +
	`"scratchpad_dir":"/tmp/claude-1000/p/0b6f3c2e/scratchpad","permission_mode":"default","effort":{"level":"high"}`

func parse(t *testing.T, raw string) *Payload {
	t.Helper()
	p, err := Parse([]byte(raw))
	if err != nil {
		t.Fatalf("%v\n%s", err, raw)
	}
	return p
}

// clean checks a payload read with no fault, and its common fields.
func clean(t *testing.T, p *Payload, event string) {
	t.Helper()
	if p.Fault != nil {
		t.Fatalf("%s: a fault: %v", event, p.Fault)
	}
	if p.Event != event || p.Session != "0b6f3c2e-5d1a-4e8b-9c7f-2a4d6e8f0a1b" || p.Cwd != "/home/someone/proj" ||
		p.TranscriptPath != "/home/someone/.claude/projects/p/0b6f3c2e.jsonl" || p.PermissionMode != "default" {
		t.Errorf("%s: the common fields: %+v", event, p)
	}
}

// Every event Bonsai reads: the ten bonsai hook record writes, SessionStart (bonsai hook start), PreToolUse (the guard)
// and Stop (the stop gate, step 5.3.4), each with its own fields.
func TestEveryEvent(t *testing.T) {
	type want struct {
		event string
		check func(p *Payload) bool
	}
	cases := []struct {
		raw string
		want
	}{
		{`{` + common + `,"hook_event_name":"SessionStart","source":"startup","model":"claude-opus-5","agent_type":"reviewer",` +
			`"session_title":"a title"}`, want{"SessionStart", func(p *Payload) bool {
			return p.Source == "startup" && p.Model == "claude-opus-5" && p.AgentType == "reviewer"
		}}},
		{`{` + common + `,"hook_event_name":"SessionStart","source":"resume","model":"claude-opus-5","seconds_since_last_response":5400,` +
			`"context_tokens":182340,"prompt_cache_likely_expired":true,"estimated_cache_write_usd":1.1396}`,
			want{"SessionStart", func(p *Payload) bool { return p.Source == "resume" && p.Model == "claude-opus-5" }}},
		{`{` + common + `,"hook_event_name":"SessionStart","source":"clear"}`,
			want{"SessionStart", func(p *Payload) bool { return p.Source == "clear" && p.Model == "" }}},
		{`{` + common + `,"hook_event_name":"SessionStart","source":"compact","model":{"id":"claude-made-up","display_name":"x"}}`,
			want{"SessionStart", func(p *Payload) bool { return p.Source == "compact" && p.Model == "claude-made-up" }}},
		{`{` + common + `,"hook_event_name":"UserPromptSubmit","prompt":"Write a function","session_title":"t"}`,
			want{"UserPromptSubmit", func(p *Payload) bool { return p.Prompt == "Write a function" }}},
		{`{` + common + `,"hook_event_name":"PreToolUse","tool_name":"Bash","tool_input":{"command":"npm test",` +
			`"description":"Run test suite","timeout":120000,"run_in_background":false},"tool_use_id":"toolu_01ABC123"}`,
			want{"PreToolUse", func(p *Payload) bool {
				return p.Tool == "Bash" && p.Command == "npm test" && p.ToolUseID == "toolu_01ABC123"
			}}},
		{`{` + common + `,"hook_event_name":"PermissionRequest","tool_name":"Bash","tool_input":{"command":"rm -rf node_modules",` +
			`"description":"Remove node_modules directory"},"permission_suggestions":[{"type":"addRules","rules":[{"toolName":"Bash",` +
			`"ruleContent":"rm -rf node_modules"}],"behavior":"allow","destination":"localSettings"}]}`,
			want{"PermissionRequest", func(p *Payload) bool {
				return p.Tool == "Bash" && p.Command == "rm -rf node_modules" && p.ToolUseID == ""
			}}},
		{`{` + common + `,"hook_event_name":"PostToolUse","tool_name":"Write","tool_input":{"file_path":"/path/to/file.txt",` +
			`"content":"file content"},"tool_response":{"filePath":"/path/to/file.txt","type":"create"},"tool_use_id":"toolu_01ABC123",` +
			`"duration_ms":12}`, want{"PostToolUse", func(p *Payload) bool {
			return p.Tool == "Write" && p.Path == "/path/to/file.txt" && p.Content == "file content"
		}}},
		{`{` + common + `,"hook_event_name":"PostToolUseFailure","tool_name":"Bash","tool_input":{"command":"npm test",` +
			`"description":"Run test suite"},"tool_use_id":"toolu_01ABC123","error":"Exit code 1\nError: Cannot find module",` +
			`"is_interrupt":false,"duration_ms":4187}`, want{"PostToolUseFailure", func(p *Payload) bool {
			return p.Tool == "Bash" && !p.IsInterrupt && p.ToolUseID == "toolu_01ABC123"
		}}},
		{`{` + common + `,"hook_event_name":"PostToolUseFailure","tool_name":"Bash","tool_input":{"command":"sleep 99"},` +
			`"tool_use_id":"toolu_01ABC124","error":"aborted","is_interrupt":true}`,
			want{"PostToolUseFailure", func(p *Payload) bool { return p.IsInterrupt }}},
		{`{` + common + `,"hook_event_name":"Notification","message":"Claude needs your permission","title":"Permission needed",` +
			`"notification_type":"permission_prompt"}`, want{"Notification", func(p *Payload) bool {
			return p.Message == "Claude needs your permission" && p.NotificationType == "permission_prompt"
		}}},
		{`{` + common + `,"hook_event_name":"SubagentStart","agent_id":"agent-abc123","agent_type":"Explore"}`,
			want{"SubagentStart", func(p *Payload) bool { return p.AgentID == "agent-abc123" && p.AgentType == "Explore" }}},
		{`{` + common + `,"hook_event_name":"SubagentStop","stop_hook_active":false,"agent_id":"def456","agent_type":"Explore",` +
			`"agent_transcript_path":"/home/someone/.claude/projects/p/0b6f3c2e/subagents/agent-def456.jsonl",` +
			`"last_assistant_message":"Analysis complete.","background_tasks":[],"session_crons":[]}`,
			want{"SubagentStop", func(p *Payload) bool {
				return p.AgentID == "def456" && !p.StopHookActive &&
					p.AgentTranscriptPath == "/home/someone/.claude/projects/p/0b6f3c2e/subagents/agent-def456.jsonl"
			}}},
		{`{` + common + `,"hook_event_name":"Stop","stop_hook_active":true,"last_assistant_message":"I've completed the refactoring.",` +
			`"background_tasks":[{"id":"task-001","type":"shell","status":"running","description":"tail logs","command":"tail -f x"}],` +
			`"session_crons":[{"id":"cron-001","schedule":"0 9 * * 1-5","recurring":true,"prompt":"check the build"}]}`,
			want{"Stop", func(p *Payload) bool { return p.StopHookActive && p.Command == "" }}},
		{`{` + common + `,"hook_event_name":"SessionEnd","reason":"prompt_input_exit"}`,
			want{"SessionEnd", func(p *Payload) bool { return p.Reason == "prompt_input_exit" }}},
		// An event Bonsai does not read is passed on by name, its common fields read.
		{`{` + common + `,"hook_event_name":"PreCompact","trigger":"manual","custom_instructions":""}`,
			want{"PreCompact", func(p *Payload) bool { return p.Tool == "" && p.Input == nil }}},
	}
	for _, c := range cases {
		p := parse(t, c.raw)
		clean(t, p, c.event)
		if !c.check(p) {
			t.Errorf("%s: %+v", c.event, p)
		}
	}
}

// Every tool whose input Bonsai reads, and some it passes on by name: the file tools' path kept as written, what
// Edit, MultiEdit and Write would write, the command of Bash, PowerShell and Monitor, and tool_input kept whole.
func TestEveryTool(t *testing.T) {
	cases := []struct {
		tool, input string
		want        Payload
	}{
		{"Bash", `{"command":"npm test","description":"Run test suite","timeout":120000,"run_in_background":false,` +
			`"dangerouslyDisableSandbox":false}`, Payload{Command: "npm test"}},
		{"PowerShell", `{"command":"Get-ChildItem -Recurse","description":"List files recursively","timeout":120000}`,
			Payload{Command: "Get-ChildItem -Recurse"}},
		{"Monitor", `{"description":"tail the log","timeout_ms":300000,"command":"tail -f app.log"}`, Payload{Command: "tail -f app.log"}},
		{"Monitor", `{"description":"a feed","timeout_ms":300000,"ws":{"url":"wss://example.invalid/feed","protocols":["v1"]}}`, Payload{}},
		{"Write", `{"file_path":"C:\\project\\src\\index.ts","content":"..."}`, Payload{Path: `C:\project\src\index.ts`, Content: "..."}},
		{"Edit", `{"file_path":"/path/to/file.txt","old_string":"original text","new_string":"replacement text","replace_all":false}`,
			Payload{Path: "/path/to/file.txt", Edits: []Edit{{Old: "original text", New: "replacement text"}}}},
		{"Edit", `{"file_path":"/c/project/a.txt","old_string":"a","new_string":"","replace_all":true}`,
			Payload{Path: "/c/project/a.txt", Edits: []Edit{{Old: "a", ReplaceAll: true}}}},
		{"MultiEdit", `{"file_path":"C:/project/b.go","edits":[{"old_string":"x","new_string":"y"},` +
			`{"old_string":"p","new_string":"q","replace_all":true}]}`,
			Payload{Path: "C:/project/b.go", Edits: []Edit{{Old: "x", New: "y"}, {Old: "p", New: "q", ReplaceAll: true}}}},
		{"NotebookEdit", `{"notebook_path":"\\\\?\\C:\\project\\n.ipynb","cell_id":"c1","new_source":"print(1)",` +
			`"cell_type":"code","edit_mode":"replace"}`, Payload{Path: `\\?\C:\project\n.ipynb`}},
		{"Read", `{"file_path":"/path/to/file.txt","offset":10,"limit":50}`, Payload{}},
		{"Glob", `{"pattern":"**/*.ts","path":"/path/to/dir"}`, Payload{}},
		{"Grep", `{"pattern":"TODO.*fix","path":"/path/to/dir","glob":"*.ts","output_mode":"content","-i":true,"multiline":false}`, Payload{}},
		{"WebFetch", `{"url":"https://example.com/api","prompt":"Extract the API endpoints","offset":100000}`, Payload{}},
		{"WebSearch", `{"query":"react hooks","allowed_domains":["docs.example.com"],"blocked_domains":[]}`, Payload{}},
		{"Agent", `{"prompt":"Find all API endpoints","description":"Find API endpoints","subagent_type":"Explore","model":"sonnet"}`, Payload{}},
		{"AskUserQuestion", `{"questions":[{"question":"Which framework?","header":"Framework","options":[{"label":"React",` +
			`"description":"Component library"}],"multiSelect":false}]}`, Payload{}},
		{"ExitPlanMode", `{"plan":"## Refactor auth\n1. Extract","planFilePath":"/home/someone/plans/a.md"}`, Payload{}},
		{"mcp__tracker__list_items", `{"q":1}`, Payload{}},
		{"SomeToolOfALaterVersion", `{"file_path":"/x","command":"rm -rf /"}`, Payload{}},
	}
	for _, c := range cases {
		raw := `{` + common + `,"hook_event_name":"PreToolUse","tool_name":"` + c.tool + `","tool_input":` + c.input +
			`,"tool_use_id":"toolu_01X","mcp_server":{"name":"tracker","source":"user"}}`
		p := parse(t, raw)
		clean(t, p, "PreToolUse")
		if p.Tool != c.tool || p.Path != c.want.Path || p.Command != c.want.Command || p.Content != c.want.Content ||
			!reflect.DeepEqual(p.Edits, c.want.Edits) {
			t.Errorf("%s %s: path %q command %q content %q edits %+v", c.tool, c.input, p.Path, p.Command, p.Content, p.Edits)
		}
		whole, err := schema.Decode([]byte(c.input))
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(p.Input, whole) || !reflect.DeepEqual(p.InputObject(), whole) {
			t.Errorf("%s: tool_input not kept whole: %#v", c.tool, p.Input)
		}
		if FileTool(c.tool) != (c.want.Path != "") {
			t.Errorf("%s: FileTool %v", c.tool, FileTool(c.tool))
		}
	}
}

// A payload that is not one JSON object is not read: Parse and Read give an ASCII error that follows "the hook
// input". A leading byte order mark is allowed.
func TestPayloadsThatDoNotRead(t *testing.T) {
	for raw, inErr := range map[string]string{
		"":                                   "is empty",
		" \n\t":                              "is empty",
		"edit protected.txt":                 "is not one JSON document (invalid character 'e'",
		`["PreToolUse"]`:                     "is not a JSON object",
		`"PreToolUse"`:                       "is not a JSON object",
		`null`:                               "is not a JSON object",
		`{"hook_event_name":"Stop"} {}`:      "is not one JSON document (something follows it)",
		`{"hook_event_name":"Stop"`:          "is not one JSON document (unexpected end",
		`{"hook_event_name":"Stop",}`:        "is not one JSON document",
		`{"hook_event_name":"Stop","x":}`:    "is not one JSON document",
		"{\"x\":\"\x00\"}":                   "is not one JSON document",
		"{\"x\":\"a\\q\"}":                   "is not one JSON document (invalid escape",
		"{\"hook_event_name\":\"Stop\"}\xff": "is not one JSON document",
		"\xEF\xBB\xBF\xEF\xBB\xBF{}":         "is not one JSON document",
		"{\"x\":[" + strings.Repeat("[", 10001) + strings.Repeat("]", 10001) + "]}": "is not one JSON document",
	} {
		p, err := Parse([]byte(raw))
		if err == nil || !strings.HasPrefix(err.Error(), inErr) {
			t.Errorf("%.40q: %v %+v, want %q", raw, err, p, inErr)
			continue
		}
		for _, r := range err.Error() {
			if r < 0x20 || r > 0x7e {
				t.Errorf("%.40q: the error is not ASCII: %q", raw, err)
			}
		}
	}
	if p, err := Parse([]byte("\xEF\xBB\xBF" + `{` + common + `,"hook_event_name":"Stop"}`)); err != nil || p.Fault != nil || p.Event != "Stop" {
		t.Errorf("a byte order mark: %+v %v", p, err)
	}
	// Read: a terminal is not waited on; a payload past the cap is not read; a stdin that fails says so.
	if _, err := Read(strings.NewReader(`{}`), true); err == nil || !strings.HasPrefix(err.Error(), "is a terminal") {
		t.Errorf("a terminal: %v", err)
	}
	if _, err := Read(nil, false); err == nil || err.Error() != "is empty" {
		t.Errorf("no stdin: %v", err)
	}
	big := `{"hook_event_name":"Stop","x":"` + strings.Repeat("y", MaxInput) + `"}`
	if _, err := Read(strings.NewReader(big), false); err == nil || err.Error() != "is larger than 64 MiB" {
		t.Errorf("over the cap: %v", err)
	}
	atCap := `{"hook_event_name":"Stop","x":"` + strings.Repeat("y", MaxInput-len(`{"hook_event_name":"Stop","x":""}`)) + `"}`
	if p, err := Read(strings.NewReader(atCap), false); err != nil || p.Event != "Stop" {
		t.Errorf("at the cap: %v", err)
	}
	if _, err := Read(io.MultiReader(strings.NewReader(`{"a":`), failing{}), false); err == nil ||
		!strings.HasPrefix(err.Error(), "cannot be read (the stdin broke)") {
		t.Errorf("a failing stdin: %v", err)
	}
}

type failing struct{}

func (failing) Read([]byte) (int, error) { return 0, errors.New("the stdin broke") }

// A value Bonsai reads breaks a rule: the payload is read, Fault says what is wrong (ASCII, after "the hook input"),
// and each field reads as Go's lenient decoder reads it (the recorder's reading since step 5.2.4).
func TestFaults(t *testing.T) {
	pre := `"hook_event_name":"PreToolUse","tool_name":"Edit"`
	cases := []struct {
		name, raw, fault string
		check            func(p *Payload) bool
	}{
		{"no event", `{"session_id":"s"}`, "has no hook_event_name", func(p *Payload) bool { return p.Session == "s" }},
		{"an empty event", `{"hook_event_name":""}`, "has no hook_event_name", nil},
		{"a null event", `{"hook_event_name":null}`, "has no hook_event_name", nil},
		{"an event not a string", `{"hook_event_name":7}`, "has a hook_event_name that is not a string", nil},
		{"a key twice, the last read", `{"hook_event_name":"Stop","session_id":"a","session_id":"b"}`, `has the key "session_id" twice`,
			func(p *Payload) bool { return p.Session == "b" }},
		{"a key twice, the last of the wrong type", `{"hook_event_name":"Stop","reason":"x","reason":3}`, `has the key "reason" twice`,
			func(p *Payload) bool { return p.Reason == "" }},
		{"an unread key twice", `{"hook_event_name":"Stop","x":1,"x":2}`, `has the key "x" twice`, nil},
		{"a session id with a dot", `{"hook_event_name":"Stop","session_id":"a.b"}`, `has a session_id "a.b" that cannot name a log file`,
			func(p *Payload) bool { return p.Session == "" }},
		{"a session id too long", `{"hook_event_name":"Stop","session_id":"` + strings.Repeat("a", 129) + `"}`, "has a session_id",
			func(p *Payload) bool { return p.Session == "" }},
		{"a session id not ASCII", `{"hook_event_name":"Stop","session_id":"caf\u00e9"}`, `has a session_id "caf\u00e9"`, nil},
		{"a session id not a string", `{"hook_event_name":"Stop","session_id":["a"]}`, "has a session_id that is not a string", nil},
		{"bad UTF-8 in a read string", "{\"hook_event_name\":\"Notification\",\"message\":\"caf\xe9\"}", "has a message that is not valid UTF-8",
			func(p *Payload) bool { return p.Message == "caf"+string(utf8.RuneError) }},
		{"a NUL in a name", `{"hook_event_name":"Stop","cwd":"/p\u0000"}`, "has a cwd holding a NUL character",
			func(p *Payload) bool { return p.Cwd == "/p\x00" }},
		{"a flag not a flag", `{"hook_event_name":"PostToolUseFailure","tool_name":"Bash","tool_input":{"command":"x"},"is_interrupt":"yes"}`,
			"has a is_interrupt that is not true or false", func(p *Payload) bool { return !p.IsInterrupt && p.Command == "x" }},
		{"a model of the wrong type", `{"hook_event_name":"SessionStart","model":5}`, "has a model that is neither a string nor an object", nil},
		{"a model object with a key twice", `{"hook_event_name":"SessionStart","model":{"id":"a","id":"b"}}`, "has a model that does not read",
			func(p *Payload) bool { return p.Model == "b" }},
		{"a model id not a string", `{"hook_event_name":"SessionStart","model":{"id":5}}`, "has a model whose id is not a string", nil},
		{"no tool", `{"hook_event_name":"PreToolUse","tool_input":{}}`, "has no tool_name", nil},
		{"a null tool", `{"hook_event_name":"PreToolUse","tool_name":null,"tool_input":{}}`, "names no tool", nil},
		{"no tool_input", `{` + pre + `}`, "has no tool_input", nil},
		{"a null tool_input", `{` + pre + `,"tool_input":null}`, "has no tool_input", nil},
		{"a tool_input not an object", `{` + pre + `,"tool_input":["/p/a"]}`, "has a tool_input that is not a JSON object",
			func(p *Payload) bool { return p.Input != nil && p.InputObject() == nil }},
		{"a key twice in tool_input", `{` + pre + `,"tool_input":{"file_path":"/p/free.txt","file_path":"/p/protected.txt"}}`,
			`has a tool_input that does not read (duplicate key "file_path")`, func(p *Payload) bool { return p.Input == nil && p.Path == "" }},
		{"a key twice deep in tool_input", `{` + pre + `,"tool_input":{"file_path":"/p/a","x":{"y":{"z":1,"z":2}}}}`,
			"has a tool_input that does not read", nil},
		{"bad UTF-8 in tool_input", "{" + pre + ",\"tool_input\":{\"file_path\":\"/p/\xff\"}}", "has a tool_input that does not read (is not valid UTF-8)", nil},
		{"no path", `{` + pre + `,"tool_input":{"path":"/p/a"}}`, "has no tool_input.file_path", nil},
		{"an empty path", `{` + pre + `,"tool_input":{"file_path":""}}`, "has an empty tool_input.file_path", nil},
		{"a path not a string", `{` + pre + `,"tool_input":{"file_path":["/p/a"]}}`, "has a tool_input.file_path that is not a string", nil},
		{"a NUL in a path", `{` + pre + `,"tool_input":{"file_path":"/p/a\u0000b"}}`, "has a tool_input.file_path holding a NUL character",
			func(p *Payload) bool { return p.Path == "/p/a\x00b" }},
		{"a notebook's path in the wrong key", `{"hook_event_name":"PreToolUse","tool_name":"NotebookEdit","tool_input":{"file_path":"/p/n"}}`,
			"has no tool_input.notebook_path", nil},
		{"no command", `{"hook_event_name":"PreToolUse","tool_name":"PowerShell","tool_input":{}}`, "has no tool_input.command", nil},
		{"a NUL in a command", `{"hook_event_name":"PreToolUse","tool_name":"Bash","tool_input":{"command":"rm -rf x\u0000y"}}`,
			"has a tool_input.command holding a NUL character", nil},
		{"a Monitor command not a string", `{"hook_event_name":"PreToolUse","tool_name":"Monitor","tool_input":{"command":1}}`,
			"has a tool_input.command that is not a string", nil},
		{"Write's content not a string", `{"hook_event_name":"PreToolUse","tool_name":"Write","tool_input":{"file_path":"/p/a","content":1}}`,
			"has a tool_input.content that is not a string", nil},
		{"Edit's string not a string", `{` + pre + `,"tool_input":{"file_path":"/p/a","old_string":1}}`,
			"has a tool_input.old_string that is not a string", nil},
		{"Edit's replace_all not a flag", `{` + pre + `,"tool_input":{"file_path":"/p/a","replace_all":"yes"}}`,
			"has a tool_input.replace_all that is not true or false", nil},
		{"MultiEdit's edits not a list", `{"hook_event_name":"PreToolUse","tool_name":"MultiEdit","tool_input":{"file_path":"/p/a","edits":{}}}`,
			"has a tool_input.edits that is not a list", nil},
		{"MultiEdit's edit not an object", `{"hook_event_name":"PreToolUse","tool_name":"MultiEdit","tool_input":{"file_path":"/p/a","edits":[1]}}`,
			"has a tool_input.edits[0] that is not a JSON object", nil},
		{"MultiEdit's edit's string not a string", `{"hook_event_name":"PreToolUse","tool_name":"MultiEdit","tool_input":` +
			`{"file_path":"/p/a","edits":[{"old_string":"a"},{"new_string":2}]}}`, "has a tool_input.edits[1].new_string that is not a string", nil},
	}
	for _, c := range cases {
		p := parse(t, c.raw)
		if p.Fault == nil || !strings.HasPrefix(p.Fault.Error(), c.fault) {
			t.Errorf("%s: fault %v, want %q", c.name, p.Fault, c.fault)
			continue
		}
		for _, r := range p.Fault.Error() {
			if r < 0x20 || r > 0x7e {
				t.Errorf("%s: the fault is not ASCII: %q", c.name, p.Fault)
			}
		}
		if c.check != nil && !c.check(p) {
			t.Errorf("%s: read as %+v", c.name, p)
		}
	}
}

// What Bonsai does not read is held only to being valid JSON: a key twice, bad UTF-8 or a NUL inside it is no fault,
// so a tool's response (a binary file read, a command's raw output) never stops a record.
func TestUnreadValuesAreOnlyJSON(t *testing.T) {
	raw := "{" + common + ",\"hook_event_name\":\"PostToolUse\",\"tool_name\":\"Read\",\"tool_input\":{\"file_path\":\"/p/a\"}," +
		"\"tool_response\":{\"content\":\"\xff\xfe\\u0000\\ud800\",\"content\":2,\"deep\":[{\"a\":1,\"a\":2}]},\"later\":\"\\u0000\"}"
	p := parse(t, raw)
	clean(t, p, "PostToolUse")
	if p.Tool != "Read" || p.InputObject().String("file_path") != "/p/a" {
		t.Errorf("%+v", p)
	}
	// Text a tool would write may hold a NUL, as a file may.
	p = parse(t, `{"hook_event_name":"PreToolUse","tool_name":"Write","tool_input":{"file_path":"/p/a","content":"a\u0000b"}}`)
	if p.Fault != nil || p.Content != "a\x00b" {
		t.Errorf("a NUL in content: %v %q", p.Fault, p.Content)
	}
}

// A path is kept as written, backslashes and all: its forms are the guard's to read (paths.go).
func TestPathsAreKeptAsWritten(t *testing.T) {
	for _, path := range []string{`C:\project\a.txt`, `\\?\C:\project\a.txt`, `\\srv\share\a.txt`, `/c/project/a.txt`,
		`C:/project/a.txt`, `~/a.txt`, `a.txt`, `../a.txt`, `C:\p\a.txt::$DATA`, `/p/a.txt `, `/p/caf` + "\u00e9" + `.txt`} {
		raw, err := schema.EncodeLine(schema.Object{{Key: "hook_event_name", Value: "PreToolUse"}, {Key: "tool_name", Value: "Write"},
			{Key: "tool_input", Value: schema.Object{{Key: "file_path", Value: path}}}})
		if err != nil {
			t.Fatal(err)
		}
		if p := parse(t, string(raw)); p.Fault != nil || p.Path != path {
			t.Errorf("%q read as %q (%v)", path, p.Path, p.Fault)
		}
	}
}

// An id as Bonsai keeps it: the log file's rule, one definition (record.LogFileName).
func TestID(t *testing.T) {
	for id, ok := range map[string]bool{
		"0b6f3c2e-5d1a-4e8b-9c7f-2a4d6e8f0a1b": true, "toolu_01ABC123": true, "agent-abc123": true, "a": true,
		strings.Repeat("a", 128): true, strings.Repeat("a", 129): false, "": false, "a.b": false, "a/b": false, `a\b`: false,
		"..": false, "a b": false, "caf\u00e9": false, "a\x00": false,
	} {
		if got := ID(id); (got == id && id != "") != ok {
			t.Errorf("ID(%q) = %q, want kept %v", id, got, ok)
		}
	}
}

// The payload's read is one pass: a large unread value is scanned, not kept.
func TestALargeUnreadValue(t *testing.T) {
	big := bytes.Repeat([]byte("x"), 8<<20)
	raw := `{` + common + `,"hook_event_name":"PostToolUse","tool_name":"Read","tool_input":{"file_path":"/p/a"},"tool_response":"` +
		string(big) + `"}`
	p := parse(t, raw)
	clean(t, p, "PostToolUse")
}
