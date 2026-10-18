package stopgate

// The stop gate's record (design/plan-5.md, 5.3.4 note 7): once the gate engages, one guard record (bonsai.log/1,
// contract §8.1) per decision, allow or deny, with agent_event Stop and the rule that decided, written through
// internal/record's one append path (record.New's common fields, the task and role redacted; record.WriteLog, held
// to the log's schema and cut to its 2,048 bytes) into the .bonsai/local/log/ of the verified main checkout (the
// session's own checkout while main is not verified): s-<session>.ndjson, or w-<YYYY-MM-DD>.ndjson (UTC) for a
// payload with no session (contract §8.5).
//
// What, beside the common fields: agent_event Stop; the subagent's id and type when the payload names them; no tool,
// category, tool_use_id or input_hash (a session's end is no tool call); target, what the decision is about (the
// ladder result's path in the main checkout, the run folder or the task folder), or null; text, a block's reason (null
// for an allow); decision and rule. Every free-text field passes the redactor and its cap (target 200, text 300).
//
// bonsai_path and bonsai_sha256 (the binary that answered, and its SHA-256) are written only on the record that made
// its log file, null elsewhere (the log schema): bonsai hook start makes a session's file first, so the gate rarely
// hashes itself. The path is written from ~/ under the user's home folder (contract §2.6), else as it is, with forward
// slashes.

import (
	"os"
	"path/filepath"
	"time"

	"github.com/LastStep/Bonsai/internal/hookin"
	"github.com/LastStep/Bonsai/internal/record"
	"github.com/LastStep/Bonsai/internal/redact"
)

// writeRecord appends the record of decision d at the place. in is nil when the payload has not been read.
func writeRecord(at *place, in *hookin.Payload, d Decision, getenv func(string) string, now time.Time) error {
	if in == nil {
		in = &hookin.Payload{}
	}
	name, err := record.LogFileName(in.Session, now)
	if err != nil {
		return err
	}
	l := record.New("guard", record.Common{Workspace: at.id, Local: at.local, Session: in.Session, Agent: "claude-code",
		Getenv: getenv, Redact: redact.Text, At: now})
	l.AgentEvent = ptr("Stop")
	l.SubagentID = ptr(hookin.ID(in.AgentID))
	l.SubagentType = text(in.AgentType)
	l.Checkout = text(deref(l.Checkout))
	l.Branch = text(deref(l.Branch))
	l.Target = ptr(redact.Capped(d.Target, redact.TargetCap))
	decision := "allow"
	if !d.Allow {
		decision = "deny"
		l.Text = ptr(redact.Capped(safe(oneLine(d.Why)), redact.TextCap))
	}
	l.Decision, l.Rule = ptr(decision), ptr(d.Rule)
	if claimFile(filepath.Join(at.local.Dir(), record.LogFolder, name)) {
		exe, sum := record.Self(nil, true)
		l.BonsaiPath, l.BonsaiSHA256 = ptr(fromHome(exe)), ptr(sum)
	}
	_, err = record.WriteLog(at.local.Main, l)
	return err
}

// claimFile makes the log file, empty, when it is not there yet, and reports whether this call made it: the record
// about to be appended is then the file's first, and carries the binary's path and hash. Two calls side by side
// cannot both make it. The line itself goes through record.Append, which finds the file there.
func claimFile(path string) bool {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return false
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return false
	}
	_ = f.Close()
	return true
}

// userHome is the user's home folder, which fromHome writes a path under from ~/ (a variable for the tests).
var userHome = os.UserHomeDir

// fromHome writes a path from ~/ when it lies under the user's home folder (letter case ignored on Windows), else as
// it is; forward slashes either way: record.FromHome, the rule's one home.
func fromHome(path string) string {
	h, err := userHome()
	if err != nil {
		h = ""
	}
	return record.FromHome(path, h)
}

// text gives a free-text field as written: redacted, or nil for "".
func text(s string) *string {
	if s == "" {
		return nil
	}
	return ptr(redact.Text(s))
}

// ptr gives "" as nil (written null), any other text as a pointer to it.
func ptr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func deref(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
