// Package asks is Bonsai's asks and answers (contract §9; spec §4, §8; design/plan-5.md, 5.2.5): an agent asks a
// person something typed, a person answers, and every record of it is kept in the main checkout's
// .bonsai/local/asks/, one file per UTC day, never committed. bonsai ask, answer and asks (cmd/bonsai) are thin
// commands over this package; the ladder's Bless (step 5.4) writes through Write, the same function.
//
//   - asks.go (this file): where the asks are kept and who writes them (Place), the limits, the refusal (Error).
//   - text.go: the rules every free-text field goes through: CRLF made LF, a hidden character refused (not
//     stripped), the redactor (internal/redact), then the limit, refused when over (never cut).
//   - read.go: the records read back, and each key's state (open, answered or resolved).
//   - file.go: filing, withdrawing (resolve), answering and reading one ask, each with the checks of 5.2.5's note.
//   - write.go: Write, one ask record and then its ask log record (contract §8.2), through internal/record's one
//     append path.
//
// What a record holds is bonsai.ask/1 (formats/schemas/ask.schema.json, format.Ask); what the commands print with
// --json is bonsai.asks/1 (format.Asks). An answer answers; it never grants (contract §9.3): it writes its record and
// its log record, nothing else.
package asks

import (
	"strings"
	"time"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/record"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// SessionEnv is the variable Claude Code names its session by (contract §9.3, Claude Code only): the asking
// session's id goes into its ask record, and an answer from that session is refused.
const SessionEnv = "CLAUDE_CODE_SESSION_ID"

// The limits, in characters (Unicode code points), as today's (contract §9.1, format review 4.3); changing one is an
// addition only. Each holds after redaction, and a field over its limit is refused, never cut.
const (
	TitleMax   = 300  // --title
	WhyMax     = 600  // --why
	ThenMax    = 300  // --then
	OptionMax  = 200  // each --option, and an answer's --choice
	OptionsMax = 4    // how many --option a Decide takes
	WordsMax   = 2000 // an answer's --words
	ByMax      = 60   // an answer's --by
	ViaMax     = 30   // an answer's --via
	IDMax      = 120  // a --task or --doc id
)

// LadderType is the one ask type the ladder files and no agent may (contract §9.2): bonsai ladder's Bless, step 5.4.
const LadderType = "Bless"

// AgentTypes are the types an agent files: format.AskTypes (their one home) but the ladder's. A type a pack defines
// is refused until a pack can declare one (design/plan-5.md, "Stale or in tension in the spec, for 5.2").
func AgentTypes() []string {
	var out []string
	for _, w := range format.AskTypes {
		if w.Word != LadderType {
			out = append(out, w.Word)
		}
	}
	return out
}

// The states of a key (bonsai.asks/1's closed list).
const (
	Open     = "open"
	Answered = "answered"
	Resolved = "resolved"
)

// The ops of an ask record (bonsai.ask/1's closed list).
const (
	OpFile    = "file"
	OpResolve = "resolve"
	OpAnswer  = "answer"
)

// The sources of an ask (bonsai.ask/1's closed list): the first part of its key.
const (
	SourceAgent  = "agent"
	SourceLadder = "ladder"
)

// Terminal is an answer's by when none is given (contract §9.1).
const Terminal = "terminal"

// Place is where a workspace's asks are kept, and who writes.
type Place struct {
	Workspace string          // the workspace id, bonsai.yaml's
	Local     workspace.Local // workspace.FindLocal's answer: Local.Main's .bonsai/local/asks/ holds the asks
	// Session is the asking or answering session, CLAUDE_CODE_SESSION_ID; "" outside one. An id that cannot name a
	// log file (internal/record's rule) is taken as none: such an id is not one Claude Code gives.
	Session string
	// Now is the clock a record's at is read from; nil reads time.Now.
	Now func() time.Time
}

// session is the place's session, "" when there is none or it cannot name a log file.
func (p Place) session() string {
	if p.Session == "" {
		return ""
	}
	if _, err := record.LogFileName(p.Session, time.Time{}); err != nil {
		return ""
	}
	return p.Session
}

// at is the time a record written now carries, as the records write it (UTC, to the millisecond).
func (p Place) at() string {
	now := time.Now
	if p.Now != nil {
		now = p.Now
	}
	return now().UTC().Format(record.AtLayout)
}

// Exit codes (spec §3): 2 bad input, 3 could not write, 4 wrong state.
const (
	ExitInput   = 2
	ExitRuntime = 3
	ExitState   = 4
)

// Error is a refusal or a failure: its word (one of format.ErrorWords, their one home), the exit code, what went
// wrong and the next step, in ASCII. Who names who takes the next step when it is not the word's usual one.
type Error struct {
	Code string
	Exit int
	What string
	Next string
	Who  string
}

func (e *Error) Error() string { return strings.TrimSuffix(e.What, ".") + "; next: " + e.Next }

// ptr gives "" as nil (written null), any other text as a pointer to it.
func ptr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// val gives a pointer's text, "" for nil.
func val(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
