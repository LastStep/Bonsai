package asks

// Writing an ask record and its log record (design/plan-5.md, 5.2.5 notes 5 and 6; contract §8.2, §9.1).
//
// Where: the main checkout's .bonsai/local/asks/<UTC day>.ndjson (workspace.FindLocal names the checkout), the day
// of the record's own at, so an answer goes into its own day's file; through internal/record's one append path
// (whole lines, Windows' busy retries), at most 8,192 bytes a record, refused when longer, never cut.
//
// Then its log record (contract §8.2: ask, "an ask filed, resolved or answered: target is its key"): one ask record in
// the log for each ask record written, after it, with kind the op and text null (the words stay in the asks file),
// in the asking or answering session's file when there is a session, else in its day's file. An answer that writes
// nothing writes no log record either. The ladder's Bless (step 5.4) goes through Write too.

import (
	"fmt"
	"strings"
	"time"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/record"
	"github.com/LastStep/Bonsai/internal/redact"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// Agent is the log record's agent when the record has a session: the agent program, Claude Code (contract §8.1).
const Agent = "claude-code"

// ensureGitignore and appendLine are the writes Write makes, variables for the tests.
var (
	ensureGitignore = workspace.EnsureGitignore
	appendLine      = record.Append
	writeLog        = record.WriteLog
)

// Write appends rec (its id, at, workspace and session already set) to the asks file of its day and then its ask log
// record. It gives the record as written (bonsai.ask/1's document). A record over 8,192 bytes once written is refused
// (bad-value, exit 2), nothing written; .bonsai/.gitignore that cannot be restored, or an asks file that cannot be
// written, is write-failed (exit 3), nothing written; a log record that cannot be written after its ask record is
// partly-written (exit 3): the ask record stands.
func Write(p Place, rec *format.Ask) (schema.Object, *Error) {
	f := format.MustLookup("ask")
	at, err := time.Parse(record.AtLayout, rec.At)
	if err != nil {
		return nil, unexpected(fmt.Errorf("the ask record's at %q is not %s", rec.At, record.AtLayout))
	}
	doc, err := f.Document(rec)
	if err != nil {
		return nil, unexpected(err)
	}
	if err := f.Check(doc); err != nil {
		return nil, unexpected(err)
	}
	line, err := schema.EncodeLine(doc)
	if err != nil {
		return nil, unexpected(err)
	}
	if max := f.MaxLine(); len(line) > max {
		return nil, badValue(fmt.Sprintf("the %s record is %d bytes once written, over the %d an ask record may hold (nothing is cut)",
			rec.Op, len(line), max), "shorten its text: every character outside ASCII takes 6 bytes or more in the record")
	}
	main := p.Local.Main
	if _, err := ensureGitignore(main); err != nil {
		return nil, &Error{Code: "write-failed", Exit: ExitRuntime,
			What: workspace.GitignoreFile + " is missing and cannot be written, so no ask is written: " + oneLine(err),
			Next: "check that .bonsai/ can be written, then run the command again"}
	}
	rel := record.AsksFolder + "/" + at.UTC().Format("2006-01-02") + ".ndjson"
	if _, err := appendLine(main, rel, line, f.MaxLine()); err != nil {
		return nil, &Error{Code: "write-failed", Exit: ExitRuntime,
			What: ".bonsai/local/" + rel + " cannot be written: " + oneLine(err),
			Next: "check that .bonsai/local/asks/ can be written, then run the command again"}
	}
	if err := writeLogRecord(p, rec, at); err != nil {
		return doc, &Error{Code: "partly-written", Exit: ExitRuntime,
			What: fmt.Sprintf("the %s record of %s was written, but its log record was not: %s", rec.Op, rec.Key, oneLine(err)),
			Next: "check that .bonsai/local/log/ can be written; the ask record stands (bonsai ask --status " + rec.Key + ")"}
	}
	return doc, nil
}

// writeLogRecord writes rec's ask log record: target its key, kind its op, text null; the session's, or the day's.
func writeLogRecord(p Place, rec *format.Ask, at time.Time) error {
	session, agent := p.session(), ""
	if session != "" {
		agent = Agent
	}
	l := record.New("ask", record.Common{Workspace: p.Workspace, Local: p.Local, Session: session, Agent: agent,
		Redact: redact.Text, At: at})
	key, op := rec.Key, rec.Op
	l.Target, l.Kind = &key, &op
	_, err := writeLog(p.Local.Main, l)
	return err
}

func unexpected(err error) *Error {
	return &Error{Code: "unexpected", Exit: ExitRuntime, What: oneLine(err),
		Next: "run the command again; if it fails again, report it to Bonsai's maintainers"}
}

// oneLine is an error's text on one line, in ASCII (ascii).
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
