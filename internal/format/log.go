package format

// The log record (bonsai.log/1, contract §8.1): one JSON line per thing an agent did or Bonsai recorded. Bonsai
// writes it (the guard today, by its own code in internal/guard/record.go; the recorder from step 5.2, through
// Log.Encode) and reads it (bonsai logs, the sessions table).
//
// Step 5.2.0 puts the log's events and categories here, in one Go table each beside this type (their one home), and
// registers them as Lists in format.go's entry for log, so bonsai check --schema bonsai.log prints them.

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
	Remote       *string       `json:"remote"` // the Remote Control id
	Extra        schema.Object `json:"-"`      // fields this Bonsai does not know (the guard's bonsai_path, bonsai_sha256 until set 5)
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
