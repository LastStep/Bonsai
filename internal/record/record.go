// Package record is how Bonsai writes and reads the files of a project's .bonsai/local/ (contract §3, §8, §9): one
// append path for every writer, the reader of the log, and the one builder of every log record's common fields.
//
//   - append.go: Append, the one append path every Bonsai writer of .bonsai/local/ uses (the recorder and bonsai hook
//     start, bonsai log append, the asks' records, the cleaner's clean records); WriteLog, a log record through it.
//     The guard keeps its own writer (internal/guard/record.go) until step 5.3 changes guard code.
//   - read.go: a log file's records in order, a torn line skipped and counted; a file's last record without reading
//     it whole; a log folder's session and day files by their names.
//   - record.go (this file): New, the common fields of a log record, and Line, a record cut to fit its 2,048 bytes.
//
// Where a writer writes is workspace.FindLocal's answer (the main checkout's .bonsai/local/, found by reading the
// checkout's .git); the salt that keys input_hash is workspace.Salt. What each record says beyond the common fields
// is its writer's: the recorder (step 5.2.4), bonsai ask (5.2.5), the cleaner (5.2.6b).
package record

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// AtLayout is how a record's at is written: UTC, to the millisecond (contract §8.1).
const AtLayout = "2006-01-02T15:04:05.000Z"

// Common is where a log record's common fields come from (design/plan-5.md, 5.2.2 note 8).
type Common struct {
	// Workspace is the workspace id, bonsai.yaml's.
	Workspace string
	// Local is where the writer runs (workspace.FindLocal): checkout is its Root's folder name, branch its Branch.
	// For a clean record at a session's end, the session's checkout; from check --write, the main checkout.
	Local workspace.Local
	// Session and Agent: on a hook's record, the payload's session and "claude-code"; on an ask record,
	// CLAUDE_CODE_SESSION_ID and "claude-code" when it is set; "" (null) on a log append or clean record, and
	// wherever there is none (contract §8.4).
	Session, Agent string
	// Getenv reads BONSAI_TASK and BONSAI_ROLE, the record's task and role; nil reads the process's environment.
	Getenv func(string) string
	// Redact is the redactor task and role pass through (contract §8.1 gives them raw; Bonsai redacts every
	// free-text string it writes). Nil writes both null: never a value unredacted.
	Redact func(string) string
	// At is when the record happened; the zero time is now.
	At time.Time
}

// New lays out a log record of event with its common fields filled, in the schema's order: format, id (a new random
// UUID), at, workspace, session, agent, checkout, branch, task, role, labels ({}) and remote (null). Every other
// field is null until the caller fills it. The one builder of the recorder's, log append's, the asks' and the
// cleaner's records, so each writes the common fields the same way.
func New(event string, c Common) *format.Log {
	at := c.At
	if at.IsZero() {
		at = time.Now()
	}
	getenv := c.Getenv
	if getenv == nil {
		getenv = os.Getenv
	}
	fromEnv := func(name string) *string {
		v := getenv(name)
		if v == "" || c.Redact == nil {
			return nil
		}
		return ptr(c.Redact(v))
	}
	return &format.Log{
		ID:        NewID(),
		At:        at.UTC().Format(AtLayout),
		Workspace: c.Workspace,
		Event:     event,
		Session:   ptr(c.Session),
		Agent:     ptr(c.Agent),
		Checkout:  ptr(CheckoutName(c.Local.Root)),
		Branch:    ptr(c.Local.Branch),
		Task:      fromEnv("BONSAI_TASK"),
		Role:      fromEnv("BONSAI_ROLE"),
		Labels:    schema.Object{},
	}
}

// CheckoutName is a checkout's folder name, never its path (contract §2.6): "" at a file system's root.
func CheckoutName(root string) string {
	if root == "" {
		return ""
	}
	name := filepath.Base(filepath.Clean(root))
	if name == "." || strings.ContainsAny(name, `/\:`) {
		return ""
	}
	return name
}

// NewID gives a random UUID (version 4), a record's id.
func NewID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	b[6] = b[6]&0x0f | 0x40
	b[8] = b[8]&0x3f | 0x80
	h := hex.EncodeToString(b[:])
	return h[0:8] + "-" + h[8:12] + "-" + h[12:16] + "-" + h[16:20] + "-" + h[20:32]
}

// shrinkable are the log's text fields that give up length when a record is over its 2,048 bytes, by name. The
// longest is halved first, a label's value among them; the fields that name or pair a record (format, id, at,
// workspace, session, event, input_hash, decision, remote, bonsai_sha256) are never cut. bonsai_path is the binary's
// path (formats set 5).
var shrinkable = map[string]bool{"text": true, "target": true, "bonsai_path": true, "task": true, "role": true,
	"branch": true, "checkout": true, "subagent_type": true, "subagent_id": true, "tool_use_id": true, "tool": true,
	"category": true, "kind": true, "source": true, "model": true, "reason": true, "rule": true, "agent": true,
	"agent_event": true}

// Line encodes a log record as its one line, LF-ended, held to bonsai.log/1 by format.Log's writer. This is where a
// record is cut to fit (format.Log.Encode refuses, never cuts): a field longer than its schema's maxLength (target,
// text) is cut to it first, then, while the line is over the log's 2,048 bytes, the longest shrinkable text (the
// shrinkable table) is halved, a cut text ending in "...". The record l itself is left as it is.
func Line(l *format.Log) ([]byte, error) {
	f := format.MustLookup("log")
	doc, err := f.Document(l)
	if err != nil {
		return nil, err
	}
	props, _ := f.Schema().Get("properties")
	for i, m := range doc {
		s, isText := m.Value.(string)
		if !isText {
			continue
		}
		if max, ok := maxLength(props, m.Key); ok && utf8.RuneCountInString(s) > max {
			doc[i].Value = Cut(s, max)
		}
	}
	max := f.MaxLine()
	for i := 0; i < 64; i++ {
		line, err := schema.EncodeLine(doc)
		if err != nil {
			return nil, err
		}
		if len(line) <= max {
			return f.Encode(doc)
		}
		if !halveLongest(doc) {
			break
		}
	}
	return nil, fmt.Errorf("a %s record does not fit in %d bytes even with its text cut", f.Versioned(), max)
}

// halveLongest halves the shrinkable text that is longest on the line (lineLen), a label's value among them; false
// when none has more than three characters left to give up.
func halveLongest(doc schema.Object) bool {
	var best string
	bestLen := 0
	var set func(string)
	consider := func(s string, put func(string)) {
		if n := lineLen(s); n > bestLen && utf8.RuneCountInString(s) > 3 {
			best, bestLen, set = s, n, put
		}
	}
	for i := range doc {
		switch v := doc[i].Value.(type) {
		case string:
			if shrinkable[doc[i].Key] {
				consider(v, func(s string) { doc[i].Value = s })
			}
		case schema.Object:
			if doc[i].Key != "labels" {
				continue
			}
			for j := range v {
				if s, ok := v[j].Value.(string); ok {
					consider(s, func(s string) { v[j].Value = s })
				}
			}
		}
	}
	if set == nil {
		return false
	}
	set(Cut(best, utf8.RuneCountInString(best)/2))
	return true
}

// lineLen is how many bytes a text takes inside a record's line, as schema.EncodeLine writes it: printable ASCII one
// byte (a quote or backslash two), a control character two or six, any other character six, or twelve above U+FFFF.
func lineLen(s string) int {
	n := 0
	for _, r := range s {
		switch {
		case r == '"' || r == '\\' || r == '\n' || r == '\r' || r == '\t':
			n += 2
		case r >= 0x20 && r < 0x7f:
			n++
		case r > 0xffff:
			n += 12
		default:
			n += 6
		}
	}
	return n
}

// maxLength reads a property's maxLength from the schema's properties.
func maxLength(props any, key string) (int, bool) {
	p, _ := props.(schema.Object)
	ps, _ := p.Get(key)
	o, _ := ps.(schema.Object)
	v, ok := o.Get("maxLength")
	if !ok {
		return 0, false
	}
	switch n := v.(type) {
	case json.Number:
		i, err := strconv.Atoi(string(n))
		return i, err == nil
	case int:
		return n, true
	case int64:
		return int(n), true
	case float64:
		return int(n), true
	}
	return 0, false
}

// Cut keeps at most max characters of s, a cut text ending in "...".
func Cut(s string, max int) string {
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	r := []rune(s)
	if max <= 3 {
		return string(r[:max])
	}
	return string(r[:max-3]) + "..."
}

// ptr gives "" as nil (written null), any other text as a pointer to it.
func ptr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
