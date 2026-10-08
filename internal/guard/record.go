package guard

// The guard record: one bonsai.log/1 line per decision (spec §7: "every guard decision is a guard record in the
// log"; contract §8.1), appended to the project's .bonsai/local/log/: s-<session>.ndjson for a session's calls,
// w-<YYYY-MM-DD>.ndjson (UTC) for a payload with no session (contract §8.5). Lines are only appended, each in one
// write, so records from calls running side by side never mix.
//
// Every field of contract §8.1 is written, in its order, null where it does not apply here: input_hash is null
// (the machine's salt that keys it is the recorder's, step 5.2), ok is null (the call has not run), and target is
// null for a shell call or a path outside the project (no redaction yet, step 5.2, so no command line is kept).
//
// Two fields follow them: bonsai_path, the binary that answered (its own path, forward slashes), on every record,
// and bonsai_sha256, its SHA-256, on the record of the call that made the log file (a session's first call), else
// null. Spec §3
// and §8 have `bonsai hook start` log both once per session, and the formats set leaves their names open
// (formats/README.md, "A name not invented"); until that hook exists the guard writes them under these two names, as
// fields outside log.schema.json, which a reader keeps (contract §2.2) and the studio's bridge does not forward (it
// forwards by allowlist, contract §2.6). Hashing the binary takes about 3 ms, twice the hook's whole start-up, so it
// is done once per session file and not on every call. The path is absolute and stays on the machine: the log is
// never committed.
//
// Where: the project's own folder, the one holding bonsai.yaml. In a worktree that is the worktree's folder, not the
// main checkout's as contract §3 has it: finding the main checkout in a way no agent-editable file can redirect is
// step 5.3's (design/plan.md, "What still links Bonsai and the studio", item 5), and git's answer is such a file.

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// LogDir is the folder of the project's log, relative to the project's folder (contract §3).
const LogDir = ".bonsai/local/log"

// MaxRecord caps one record's line, newline included (contract §8.1: at most 2,048 bytes).
const MaxRecord = 2048

// openWait is how long an append keeps retrying while Windows reports the log file busy (a variable for tests).
var openWait = time.Second

// recordFile names the log file a record goes to.
func recordFile(session string, at time.Time) string {
	if session != "" {
		return "s-" + session + ".ndjson"
	}
	return "w-" + at.UTC().Format("2006-01-02") + ".ndjson"
}

// writeRecord appends the record of decision d to the project's log. in is nil when the payload could not be read.
func writeRecord(root string, cfg *workspace.Config, in *Input, d Decision, getenv func(string) string, at time.Time) error {
	if in == nil {
		in = &Input{}
	}
	dir := filepath.Join(root, filepath.FromSlash(LogDir))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	file := filepath.Join(dir, recordFile(in.Session, at))
	f, err := openRetry(file, os.O_WRONLY|os.O_CREATE|os.O_EXCL|os.O_APPEND)
	first := err == nil
	if errors.Is(err, fs.ErrExist) {
		f, err = openRetry(file, os.O_WRONLY|os.O_APPEND)
	}
	if err != nil {
		return err
	}
	exe, sum := self(first)
	line, err := recordLine(buildRecord(root, cfg, in, d, getenv, at, exe, sum))
	if err != nil {
		_ = f.Close()
		return err
	}
	_, werr := f.Write(line)
	cerr := f.Close()
	if werr != nil {
		return werr
	}
	return cerr
}

// openRetry opens a log file, retrying while Windows reports it busy, for up to openWait.
func openRetry(name string, flag int) (*os.File, error) {
	deadline := time.Now().Add(openWait)
	delay := 5 * time.Millisecond
	for {
		f, err := os.OpenFile(name, flag, 0o644)
		if err == nil || !workspace.IsBusy(err) || time.Now().After(deadline) {
			return f, err
		}
		time.Sleep(delay)
		if delay < 80*time.Millisecond {
			delay *= 2
		}
	}
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

// buildRecord lays out one guard record, every field of contract §8.1 in its order, then bonsai_path and
// bonsai_sha256.
func buildRecord(root string, cfg *workspace.Config, in *Input, d Decision, getenv func(string) string, at time.Time,
	exe, sum string) schema.Object {
	decision, text := "deny", d.Why
	if d.Allow {
		decision, text = "allow", ""
	}
	return schema.Object{
		{Key: "format", Value: "bonsai.log/1"},
		{Key: "id", Value: newID()},
		{Key: "at", Value: at.UTC().Format("2006-01-02T15:04:05.000Z")},
		{Key: "workspace", Value: cfg.ID},
		{Key: "session", Value: orNull(in.Session)},
		{Key: "agent", Value: "claude-code"},
		{Key: "agent_event", Value: "PreToolUse"},
		{Key: "event", Value: "guard"},
		{Key: "subagent_id", Value: orNull(in.AgentID)},
		{Key: "subagent_type", Value: orNull(in.AgentType)},
		{Key: "tool_use_id", Value: orNull(in.ToolUseID)},
		{Key: "input_hash", Value: nil},
		{Key: "checkout", Value: orNull(checkoutName(root))},
		{Key: "branch", Value: orNull(branch(root))},
		{Key: "task", Value: orNull(getenv("BONSAI_TASK"))},
		{Key: "role", Value: orNull(getenv("BONSAI_ROLE"))},
		{Key: "tool", Value: orNull(in.Tool)},
		{Key: "category", Value: orNull(category(in.Tool))},
		{Key: "target", Value: orNull(cut(d.Target, 200))},
		{Key: "ok", Value: nil},
		{Key: "kind", Value: nil},
		{Key: "text", Value: orNull(cut(text, 300))},
		{Key: "source", Value: nil},
		{Key: "model", Value: nil},
		{Key: "reason", Value: nil},
		{Key: "decision", Value: decision},
		{Key: "rule", Value: d.Rule},
		{Key: "labels", Value: schema.Object{}},
		{Key: "remote", Value: nil},
		{Key: "bonsai_path", Value: orNull(exe)},
		{Key: "bonsai_sha256", Value: orNull(sum)},
	}
}

// shrinkable are the fields a record too long for MaxRecord gives up length from, longest first.
var shrinkable = []string{"text", "target", "bonsai_path", "task", "role", "branch", "checkout", "subagent_type",
	"subagent_id", "tool_use_id", "tool"}

// recordLine encodes a record as one line of at most MaxRecord bytes, halving its longest text field until it fits.
func recordLine(rec schema.Object) ([]byte, error) {
	for i := 0; i < 64; i++ {
		line, err := schema.EncodeLine(rec)
		if err != nil {
			return nil, err
		}
		if len(line) <= MaxRecord {
			return line, nil
		}
		longest, n := -1, 0
		for _, k := range shrinkable {
			j := rec.Index(k)
			if s, ok := rec[j].Value.(string); ok && len(s) > n {
				longest, n = j, len(s)
			}
		}
		if longest < 0 {
			break
		}
		s := rec[longest].Value.(string)
		rec[longest].Value = cut(s, utf8.RuneCountInString(s)/2)
	}
	return nil, fmt.Errorf("a guard record does not fit in %d bytes", MaxRecord)
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

// checkoutName is the working tree's folder name, never its path (contract §2.6); "" at a file system's root.
func checkoutName(root string) string {
	name := filepath.Base(root)
	if name == "" || strings.ContainsAny(name, `/\`) || name == "." {
		return ""
	}
	return name
}

// branch reads the checkout's branch from .git/HEAD when .git is a folder (no git process: the hook stays fast);
// "" for a detached HEAD, a worktree (whose .git is a file) or anything it cannot read.
func branch(root string) string {
	raw, err := os.ReadFile(filepath.Join(root, ".git", "HEAD"))
	if err != nil {
		return ""
	}
	s := strings.TrimSpace(string(raw))
	if !strings.HasPrefix(s, "ref: refs/heads/") {
		return ""
	}
	return strings.TrimPrefix(s, "ref: refs/heads/")
}

// newID gives a random UUID (version 4).
func newID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	b[6] = b[6]&0x0f | 0x40
	b[8] = b[8]&0x3f | 0x80
	h := hex.EncodeToString(b[:])
	return h[0:8] + "-" + h[8:12] + "-" + h[12:16] + "-" + h[16:20] + "-" + h[20:32]
}

// cut keeps at most max runes of s, ending a cut one with "...".
func cut(s string, max int) string {
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	r := []rune(s)
	if max <= 3 {
		return string(r[:max])
	}
	return string(r[:max-3]) + "..."
}

// orNull writes "" as null.
func orNull(s string) any {
	if s == "" {
		return nil
	}
	return s
}
