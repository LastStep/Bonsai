package guard

// The guard record: one bonsai.log/1 line per decision (spec §7: "every guard decision is a guard record in the
// log"; contract §8.1), written through internal/record's one append path (design/plan-5.md, 5.3.1 note 4): its
// common fields from record.New (task and role redacted), appended whole by record.Append, which retries a file
// Windows reports busy and restores a missing .bonsai/.gitignore. Lines from calls running side by side never mix.
//
// The line itself is laid out here (line: every field in the schema's order) and cut to fit as record.Line cuts, not
// by record.Line: that one holds every record to the log schema through internal/format, which parses the 12 KB
// schema at a process's first record, about 0.25 ms of a hook whose whole start-up must stay under 5 ms (spec §3),
// and the guard writes one record per process. The tests hold every guard record to the schema (each field, its type
// and the order), so the two cannot drift.
//
// Where: the .bonsai/local/log/ of the project whose rules decided (rule.go's recordIn: a file tool's path's project,
// a shell call's session project), which for a verified worktree is its main checkout's (contract §3: worktrees share
// the main checkout's local/, so bonsai logs, the sessions table and the cleaner see them); s-<session>.ndjson for a
// session's calls, w-<YYYY-MM-DD>.ndjson (UTC) for a payload with no session (contract §8.5).
//
// What, beside the common fields: agent_event PreToolUse; the subagent's id and type, and tool_use_id, when the
// payload names them; input_hash, the salt's HMAC over the tool input (hookin.InputHash, as the recorder's tool_start
// record has it, so a call's two records pair; null with no salt or no input read); tool and category; target, the
// project-relative path judged (null for a shell call or a path outside every project); text, a refusal's reason (null
// for an allow); decision and rule. Every free-text field passes the redactor and its cap (target 200, text 300).
//
// bonsai_path and bonsai_sha256 (the binary that answered, and its SHA-256) are written only on the record that made
// its log file, null elsewhere (the log schema). bonsai hook start makes a session's file first, with both, so the
// guard rarely hashes itself; it still does wherever no hook start made the file (a worktree session's first call
// when the file is main's and new, a project without the start line, a payload with no session). The path is written
// from ~/ under the user's home folder (contract §2.6: no user name in the log), else as it is, forward slashes.

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/hookin"
	"github.com/LastStep/Bonsai/internal/record"
	"github.com/LastStep/Bonsai/internal/redact"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// LogDir is the folder of the log, relative to the main checkout whose .bonsai/local/ holds it (contract §3).
const LogDir = workspace.LocalDir + "/" + record.LogFolder

// MaxRecord caps one record's line, newline included (contract §8.1: at most 2,048 bytes; record.Line cuts to it).
const MaxRecord = 2048

// place is where a guard record goes: the workspace id it is written under, and the checkout it names with the main
// checkout whose .bonsai/local/ holds it.
type place struct {
	id    string
	local workspace.Local
}

// salt gives the input hash's key (a variable for the tests).
var salt = workspace.Salt

// writeRecord appends the record of decision d to the log at the place. in is nil when the payload could not be read.
func writeRecord(at *place, in *hookin.Payload, d Decision, getenv func(string) string, now time.Time) error {
	if in == nil {
		in = &hookin.Payload{}
	}
	name, err := record.LogFileName(in.Session, now)
	if err != nil {
		return err
	}
	l := buildRecord(at, in, d, getenv, now)
	if claimFile(filepath.Join(at.local.Dir(), record.LogFolder, name)) {
		exe, sum := self(true)
		l.BonsaiPath, l.BonsaiSHA256 = ptr(fromHome(exe)), ptr(sum)
	}
	b, err := line(l)
	if err != nil {
		return err
	}
	_, err = record.Append(at.local.Main, record.LogFolder+"/"+name, b, MaxRecord)
	return err
}

// line lays out a guard record as its one line: every field of bonsai.log/1 in the schema's order (format.Log's), then
// cut to fit MaxRecord as record.Line cuts it: the longest text the record may give up halved, a cut text ending in
// "...", until the line fits. A guard record has no labels and nothing unknown, so the fields are all there is.
func line(l *format.Log) ([]byte, error) {
	rec := schema.Object{
		{Key: "format", Value: "bonsai.log/1"}, {Key: "id", Value: l.ID}, {Key: "at", Value: l.At},
		{Key: "workspace", Value: l.Workspace}, {Key: "session", Value: orNull(l.Session)}, {Key: "agent", Value: orNull(l.Agent)},
		{Key: "agent_event", Value: orNull(l.AgentEvent)}, {Key: "event", Value: l.Event},
		{Key: "subagent_id", Value: orNull(l.SubagentID)}, {Key: "subagent_type", Value: orNull(l.SubagentType)},
		{Key: "tool_use_id", Value: orNull(l.ToolUseID)}, {Key: "input_hash", Value: orNull(l.InputHash)},
		{Key: "checkout", Value: orNull(l.Checkout)}, {Key: "branch", Value: orNull(l.Branch)}, {Key: "task", Value: orNull(l.Task)},
		{Key: "role", Value: orNull(l.Role)}, {Key: "tool", Value: orNull(l.Tool)}, {Key: "category", Value: orNull(l.Category)},
		{Key: "target", Value: orNull(l.Target)}, {Key: "ok", Value: nil}, {Key: "kind", Value: orNull(l.Kind)},
		{Key: "text", Value: orNull(l.Text)}, {Key: "source", Value: orNull(l.Source)}, {Key: "model", Value: orNull(l.Model)},
		{Key: "reason", Value: orNull(l.Reason)}, {Key: "decision", Value: orNull(l.Decision)}, {Key: "rule", Value: orNull(l.Rule)},
		{Key: "labels", Value: schema.Object{}}, {Key: "remote", Value: orNull(l.Remote)},
		{Key: "bonsai_path", Value: orNull(l.BonsaiPath)}, {Key: "bonsai_sha256", Value: orNull(l.BonsaiSHA256)},
	}
	for i := 0; i < 64; i++ {
		b, err := schema.EncodeLine(rec)
		if err != nil {
			return nil, err
		}
		if len(b) <= MaxRecord {
			return b, nil
		}
		longest, n := -1, 3
		for _, k := range shrinkable {
			j := rec.Index(k)
			if s, ok := rec[j].Value.(string); ok && utf8.RuneCountInString(s) > n {
				longest, n = j, utf8.RuneCountInString(s)
			}
		}
		if longest < 0 {
			break
		}
		rec[longest].Value = record.Cut(rec[longest].Value.(string), n/2)
	}
	return nil, fmt.Errorf("a guard record does not fit in %d bytes", MaxRecord)
}

// shrinkable are the fields a guard record too long for MaxRecord gives up length from, the longest first: its text
// fields, as record.Line's; never one that names or pairs a record (format, id, at, workspace, session, event,
// input_hash, decision, bonsai_sha256).
var shrinkable = []string{"text", "target", "bonsai_path", "task", "role", "branch", "checkout", "subagent_type",
	"subagent_id", "tool_use_id", "tool", "category", "rule", "agent", "agent_event"}

// orNull writes a missing text as null.
func orNull(p *string) any {
	if p == nil {
		return nil
	}
	return *p
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

// buildRecord lays out one guard record (this file's comment), all but bonsai_path and bonsai_sha256.
func buildRecord(at *place, in *hookin.Payload, d Decision, getenv func(string) string, now time.Time) *format.Log {
	l := record.New("guard", record.Common{Workspace: at.id, Local: at.local, Session: in.Session, Agent: "claude-code",
		Getenv: getenv, Redact: redact.Text, At: now})
	l.AgentEvent = ptr("PreToolUse")
	l.SubagentID = ptr(hookin.ID(in.AgentID))
	l.SubagentType = text(in.AgentType)
	l.ToolUseID = ptr(hookin.ID(in.ToolUseID))
	if in.Input != nil {
		if key, err := salt(); err == nil {
			l.InputHash = ptr(hookin.InputHash(key, in.Input))
		}
	}
	l.Checkout = text(deref(l.Checkout))
	l.Branch = text(deref(l.Branch))
	l.Tool = text(in.Tool)
	l.Category = ptr(category(in.Tool))
	l.Target = ptr(redact.Capped(d.Target, redact.TargetCap))
	decision := "allow"
	if !d.Allow {
		decision = "deny"
		l.Text = ptr(redact.Capped(d.Why, redact.TextCap))
	}
	l.Decision, l.Rule = ptr(decision), ptr(d.Rule)
	return l
}

// self gives the running binary's own path (forward slashes) and, when hash is true, its SHA-256 in hex. What
// cannot be found is "" (written as null).
func self(hash bool) (string, string) {
	exe, err := os.Executable()
	if err != nil {
		return "", ""
	}
	if !hash {
		return filepath.ToSlash(exe), ""
	}
	f, err := os.Open(exe)
	if err != nil {
		return filepath.ToSlash(exe), ""
	}
	defer func() { _ = f.Close() }()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return filepath.ToSlash(exe), ""
	}
	return filepath.ToSlash(exe), hex.EncodeToString(h.Sum(nil))
}

// userHome is the user's home folder, which fromHome writes a path under from ~/ (a variable for the tests).
var userHome = os.UserHomeDir

// fromHome writes a path from ~/ when it lies under the user's home folder (letter case ignored on Windows), else as
// it is; forward slashes either way. The recorder writes hook start's bonsai_path the same way.
func fromHome(path string) string {
	if path == "" {
		return ""
	}
	p := filepath.ToSlash(filepath.Clean(path))
	home, err := userHome()
	if err != nil || home == "" {
		return p
	}
	prefix := strings.TrimSuffix(filepath.ToSlash(filepath.Clean(home)), "/") + "/"
	if len(p) > len(prefix) {
		head := p[:len(prefix)]
		if head == prefix || (runtime.GOOS == "windows" && strings.EqualFold(head, prefix)) {
			return "~/" + p[len(prefix):]
		}
	}
	return p
}

// category maps a tool to its log category (contract §8.2: today's Bash becomes Shell).
func category(tool string) string {
	switch tool {
	case "":
		return ""
	case "Edit", "MultiEdit", "NotebookEdit":
		return "Edit"
	case "Write":
		return "Write"
	case "Bash", "PowerShell":
		return "Shell"
	}
	return "Other"
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
