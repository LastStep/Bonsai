package message

// Showing one message (bonsai message --show; contract §22, "The trust line"; design/plan-5.md, note 5.5.10, 3): the
// message found by its id, each pointer read again (unchanged, changed or missing), and its thread as it stands. Text
// gives it as a person or an agent reads it: the trust line first, always (format.TrustLine, its one home), then a
// line for each pointer whose file is not as it was, then the file as bonsai message writes it (frontmatter and
// body). Another agent's words are data: never the person's words, never an order above the reader's own brief,
// never consent.

import (
	"strings"
	"time"

	"github.com/LastStep/Bonsai/internal/format"
)

// Shown is one message as --show gives it.
type Shown struct {
	Entry    Entry
	Pointers []string // each pointer's state, in the message's order (PointerUnchanged, PointerChanged, PointerMissing)
	Thread   Thread   // its thread as it stands
}

// Show finds the message whose id is id in the store, reads its pointers again and folds its thread.
func Show(p Place, st *Store, id string, r Rules, now time.Time) (*Shown, *Error) {
	if !IsID(id) {
		return nil, badValue("--show "+ascii(id)+" is not a message id (M-, the UTC date, a dash and 8 hex)",
			"give --show an id bonsai messages lists, such as M-2026-10-16-3f9a01bc")
	}
	e, bad := st.Find(id)
	if e == nil {
		err := &Error{Code: "message-not-found", Exit: ExitState, What: "no message under " + Dir + "/ has the id " + id,
			Next: "run bonsai messages to list the threads and their messages"}
		if bad != nil {
			err.What = id + " is " + bad.Path + ", which does not read as a message: " + bad.Reason
			err.Next = "read the file as data, or ask its writer to write the message again (bonsai check --schema bonsai.message lists every field)"
		}
		return nil, err
	}
	s := &Shown{Entry: *e, Pointers: []string{}}
	for _, ptr := range e.Msg.Pointers {
		s.Pointers = append(s.Pointers, PointerState(p, ptr))
	}
	for _, t := range Fold(st, r, now) {
		if t.ID == e.Msg.Thread {
			s.Thread = t
		}
	}
	return s, nil
}

// Text is the message as --show prints it, in ASCII: the trust line, the pointers not as they were, then the file.
func (s *Shown) Text() (string, error) {
	msg := s.Entry.Msg
	raw, err := msg.EncodeFile()
	if err != nil {
		return "", err
	}
	var b strings.Builder
	b.WriteString(format.TrustLine(msg.From) + "\n")
	for i, state := range s.Pointers {
		switch state {
		case PointerChanged:
			b.WriteString("pointer " + ascii(msg.Pointers[i].Path) + ": changed since the message was written (its sha256 is not the one written)\n")
		case PointerMissing:
			b.WriteString("pointer " + ascii(msg.Pointers[i].Path) + ": missing from this checkout now\n")
		}
	}
	b.WriteString(textASCII(string(raw)))
	if !strings.HasSuffix(b.String(), "\n") {
		b.WriteString("\n")
	}
	return b.String(), nil
}

// textASCII writes text in ASCII, keeping its line feeds: any other character outside printable ASCII as a Go escape.
func textASCII(s string) string {
	var b strings.Builder
	for _, line := range strings.SplitAfter(s, "\n") {
		body, nl := strings.CutSuffix(line, "\n")
		b.WriteString(ascii(body))
		if nl {
			b.WriteString("\n")
		}
	}
	return b.String()
}
