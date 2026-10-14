// Package message is Bonsai's messages between agents (bonsai.message/1, set 8; contract §22; design/plan-5.md, step
// 5.5, "Messages between agents, written once", note 5.5.10): a brief from a caller to an agent it starts or resumes,
// or a report back, one markdown file each in the main checkout's .bonsai/local/messages/<recipient>/<id>.md, never
// committed. bonsai message and bonsai messages (cmd/bonsai) are thin commands over this package, the one reader every
// command uses; check, status and the cleaning read through it too.
//
//   - message.go (this file): where the messages are kept and who writes them (Place), what a project holds them to
//     (Rules, the seam step 5.5.0's limits and roles and bonsai.yaml's thresholds come in through), the size limits and
//     the refusal (Error).
//   - address.go: an address (<role>, <role>@<task>, <role>@<task>#<n>) and its folder, a name Windows can hold.
//   - read.go: every message file read (Read), a file that does not read reported and skipped, never an error for
//     the rest; one message found by its id.
//   - thread.go: the threads folded from the files (each file new, so no shared file is ever rewritten), their state
//     and their flags, the filters of bonsai messages.
//   - pointer.go: a pointer's file found and fingerprinted (SHA-256, line endings read as LF), a regular file only,
//     streamed and never held whole, at writing and again at --show, where a file changed since shows as changed.
//   - write.go: Write, one message: its checks, the file created new and never overwritten, and its message log
//     record (contract §8.2).
//   - move_windows.go, move_other.go: the rename that never replaces a file, retried while Windows reports it busy.
//
// The format's own rules (every field, its reader and writer, the trust line) are internal/format's (format.Message,
// ReadMessage, EncodeFile, TrustLine); this package adds what a folder of them needs: where each lives, the threads,
// and the rules of writing one. Bonsai never starts, wakes or resumes an agent, never delivers a message and never
// takes one as a grant or as the person's word: these files are data, attributed and never authenticated.
package message

import (
	"fmt"
	"strings"
	"time"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// Folder is the messages' folder inside .bonsai/local/ (contract §22), and Dir the same from the main checkout's top.
const (
	Folder = "messages"
	Dir    = workspace.LocalDir + "/" + Folder
)

// The size limits, in bytes. A body over MaxBody is refused when written, never cut ("long output goes in a file and
// is pointed to, never pasted"); a file over MaxFile is never read, and is listed as malformed: it is larger than any
// file bonsai message writes, so a hand-made one cannot make a reader hold it whole.
const (
	MaxBody = 64 << 10
	MaxFile = 1 << 20
)

// Place is where a workspace's messages are kept, and who writes.
type Place struct {
	Workspace string // the workspace id, bonsai.yaml's
	// Local is workspace.FindLocal's answer: Local.Main's .bonsai/local/messages/ holds the messages (a worktree writes
	// its main checkout's, as the log and asks do); Local.Root is the checkout the command runs in, from which a
	// pointer's project-relative path is read.
	Local workspace.Local
	// Session is the writing session, CLAUDE_CODE_SESSION_ID; "" outside one. It names the log file the message's log
	// record goes to, as an ask's does.
	Session string
	// Now is the clock a message's id and created are read from; nil reads time.Now.
	Now func() time.Time
	// Home gives the person's home folder, where a pointer written from ~/ lies; nil reads os.UserHomeDir.
	Home func() (string, error)
	// Getenv reads BONSAI_ROLE for the log record's role (record.Common); nil reads the process's environment.
	Getenv func(string) string
}

func (p Place) now() time.Time {
	if p.Now != nil {
		return p.Now()
	}
	return time.Now()
}

// Rules are what a project holds its messages to beyond the format: the roles its lock declares, each limit's max in
// force for a role, and the thresholds of bonsai messages' flags. They come from step 5.5.0's code (its one function
// for the limits in force, the lock's roles) and bonsai.yaml's messages; DefaultRules is Bonsai's own, from the Go
// tables, used where a project sets none.
type Rules struct {
	// Roles are the roles the linked packs declare, by name (an address's role); nil or empty where none is declared,
	// and then any role is taken.
	Roles []string
	// Max gives a limit's max in force (one of format.RoleLimits) for a role: nil for no limit. A nil Max holds no
	// brief to any max.
	Max func(role, limit string) *int64
	// UnansweredMinutes and CirclingBriefs are the flags' thresholds (format.MessageThresholds).
	UnansweredMinutes int64
	CirclingBriefs    int64
}

// DefaultRules are Bonsai's own: no role declared, so no address is refused for its role and no brief for its budget
// (no role, no limit in force), and the flags' thresholds at format.MessageThresholds' defaults, their one home.
func DefaultRules() Rules {
	r := Rules{}
	for _, t := range format.MessageThresholds {
		switch t.Setting {
		case "unanswered_minutes":
			r.UnansweredMinutes = t.Default
		case "circling_briefs":
			r.CirclingBriefs = t.Default
		}
	}
	return r
}

// declares reports whether the rules declare role; true where no role is declared at all.
func (r Rules) declares(role string) bool {
	if len(r.Roles) == 0 {
		return true
	}
	for _, d := range r.Roles {
		if d == role {
			return true
		}
	}
	return false
}

// Exit codes (spec §3): 2 bad input, 3 could not read or write, 4 wrong state.
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

func badValue(what, next string) *Error {
	return &Error{Code: "bad-value", Exit: ExitInput, What: what, Next: next}
}

func badFlag(what, next string) *Error {
	return &Error{Code: "bad-flag", Exit: ExitInput, What: what, Next: next}
}

func unexpected(err error) *Error {
	return &Error{Code: "unexpected", Exit: ExitRuntime, What: oneLine(err),
		Next: "run the command again; if it fails again, report it to Bonsai's maintainers"}
}

// oneLine is an error's text on one line, in ASCII.
func oneLine(err error) string { return ascii(err.Error()) }

// ascii writes text on one line in ASCII: a line feed, carriage return or tab as a space, any other character
// outside printable ASCII as a Go escape (spec §3: human output is ASCII).
func ascii(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r == '\n' || r == '\r' || r == '\t':
			b.WriteByte(' ')
		case r >= 0x20 && r < 0x7f:
			b.WriteRune(r)
		default:
			q := fmt.Sprintf("%+q", string(r))
			b.WriteString(q[1 : len(q)-1])
		}
	}
	return b.String()
}

func ptr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
