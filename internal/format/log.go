package format

// The log record (bonsai.log/1, contract §8.1): one JSON line per thing an agent did or Bonsai recorded. Bonsai
// writes it (the guard today, by its own code in internal/guard/record.go; the recorder from step 5.2, through
// Log.Encode) and reads it (bonsai logs, the sessions table).
//
// The log's events and categories are open lists whose known words live here, in one Go table each beside this type
// (their one home: LogEvents and LogCategories), registered as Lists in format.go's entry for log, so bonsai check
// --schema bonsai.log prints them and the reference page of lists reads them. The schema's descriptions of event and
// category name those commands and copy no word.

import "github.com/LastStep/Bonsai/internal/schema"

// Log is one log record. Every field but id, at, workspace and event may be null where it does not apply.
type Log struct {
	ID           string        `json:"id"`        // a UUID
	At           string        `json:"at"`        // UTC, to the millisecond: 2026-10-08T14:08:00.000Z
	Workspace    string        `json:"workspace"` // the workspace id
	Session      *string       `json:"session"`
	Agent        *string       `json:"agent"`
	AgentEvent   *string       `json:"agent_event"`
	Event        string        `json:"event"` // Bonsai's event name: an open list
	SubagentID   *string       `json:"subagent_id"`
	SubagentType *string       `json:"subagent_type"`
	ToolUseID    *string       `json:"tool_use_id"`
	InputHash    *string       `json:"input_hash"` // 16 lower-case hex
	Checkout     *string       `json:"checkout"`   // the checkout's folder name
	Branch       *string       `json:"branch"`
	Task         *string       `json:"task"`
	Role         *string       `json:"role"`
	Tool         *string       `json:"tool"`
	Category     *string       `json:"category"` // the tool's category: an open list
	Target       *string       `json:"target"`   // at most 200 characters
	OK           *bool         `json:"ok"`
	Kind         *string       `json:"kind"`
	Text         *string       `json:"text"` // at most 300 characters
	Source       *string       `json:"source"`
	Model        *string       `json:"model"`
	Reason       *string       `json:"reason"`
	Decision     *string       `json:"decision"` // allow, deny or null
	Rule         *string       `json:"rule"`
	Labels       schema.Object `json:"labels"`
	Remote       *string       `json:"remote"`        // the Remote Control id
	BonsaiPath   *string       `json:"bonsai_path"`   // the bonsai binary's path, forward slashes (set 6); the guard writes its absolute form
	BonsaiSHA256 *string       `json:"bonsai_sha256"` // the binary's SHA-256, 64 lower-case hex (set 6)
	Extra        schema.Object `json:"-"`             // fields this Bonsai does not know, kept as read
}

// LogEvents are the known words of a log record's event: an open list (contract section 8.2; a reader shows a word it
// does not know as other), whose one home is this table. bonsai check --schema bonsai.log prints it, and the reference
// page of lists reads it. The first eleven come from the agent's hooks (contract section 8.3), the rest are Bonsai's
// own records.
var LogEvents = []Word{
	{Word: "session_start", Means: "a session began, resumed or was cleared (the agent's SessionStart); it carries the active task found, the model, and the bonsai binary's path and hash"},
	{Word: "prompt", Means: "a person's prompt was submitted (the agent's UserPromptSubmit); its words are never kept, only its kind"},
	{Word: "tool_start", Means: "a tool call is about to run (the agent's PreToolUse)"},
	{Word: "tool_end", Means: "a tool call finished (the agent's PostToolUse)"},
	{Word: "tool_fail", Means: "a tool call failed or was interrupted (the agent's PostToolUseFailure)"},
	{Word: "permission", Means: "the agent asked for a permission (the agent's PermissionRequest)"},
	{Word: "notice", Means: "the agent noticed something to tell a person, such as that it is waiting (the agent's Notification)"},
	{Word: "subagent_start", Means: "a subagent run began (the agent's SubagentStart); it carries the active task found"},
	{Word: "subagent_stop", Means: "a subagent run ended (the agent's SubagentStop)"},
	{Word: "stop", Means: "the agent finished a turn (the agent's Stop)"},
	{Word: "session_end", Means: "a session ended (the agent's SessionEnd); its reason says why"},
	{Word: "guard", Means: "the guard decided on a tool call, allow or deny, by a named rule"},
	{Word: "ladder", Means: "a ladder result was written (contract section 11)"},
	{Word: "ask", Means: "an ask was filed, resolved or answered; the target is its key and the kind its op"},
	{Word: "event", Means: "an outside event, written by bonsai log append with its labels (contract section 8.4)"},
	{Word: "clean", Means: "a generated file Bonsai deleted; the target is its project-relative path and the reason the rule that cleaned it (contract section 8.5)"},
}

// LogCategories are the known words of a log record's category: an open list (contract section 8.2), whose one home
// is this table. A pack declares its own as <namespace>.<Name>, which are not listed; a reader shows a word it does not
// know as other. bonsai check --schema bonsai.log prints this table, and the reference page of lists reads it.
var LogCategories = []Word{
	{Word: "Read", Means: "reading a file or a folder"},
	{Word: "Search", Means: "searching files or their contents (today's Grep and Glob)"},
	{Word: "Edit", Means: "changing a file in place"},
	{Word: "Write", Means: "writing a whole file"},
	{Word: "Shell", Means: "running a command line (today's Bash and PowerShell)"},
	{Word: "Ladder", Means: "running Bonsai's ladder"},
	{Word: "MCP", Means: "calling a tool of an MCP server"},
	{Word: "Agent", Means: "starting a subagent"},
	{Word: "Web", Means: "fetching or searching the web"},
	{Word: "Other", Means: "any other tool, or one the recorder cannot place"},
}

// ReadLog reads one log record (one line, its line feed allowed).
func ReadLog(raw []byte) (*Log, error) {
	l := &Log{}
	if err := MustLookup("log").readJSONInto(raw, l); err != nil {
		return nil, err
	}
	return l, nil
}

// Encode writes the record as one line, held to bonsai.log/1 and its 2,048 bytes.
func (l *Log) Encode() ([]byte, error) { return MustLookup("log").Encode(l) }
