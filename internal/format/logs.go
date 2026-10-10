package format

// bonsai logs --json (bonsai.logs/1, contract section 8, spec section 8): the output of bonsai logs and bonsai log
// append: the log files of the main checkout's .bonsai/local/log/, or one file's records, or the record log append
// just wrote. Printed, never stored. Written by those commands (steps 5.2.3 and 5.2.5); read by agents, the studio's
// forwarder and bridge, and CI. The records inside are bonsai.log/1's, as written: open objects here (schema.Object),
// because the schema checker has no $ref and a full copy would refuse an older or hand-written record.

import "github.com/LastStep/Bonsai/internal/schema"

// Logs is the logs document.
type Logs struct {
	Workspace *WorkspaceRef   `json:"workspace"` // null when the command refused before reading one
	Files     []LogFile       `json:"files"`     // the log files, or null when the command gave no list
	Records   []schema.Object `json:"records"`   // one file's records as written, or null
	Written   schema.Object   `json:"written"`   // the record log append wrote, or null
	Error     *ErrorObject    `json:"error"`     // null unless the command refused or failed
	Extra     schema.Object   `json:"-"`
}

// LogFile is one log file and what it holds.
type LogFile struct {
	File       string        `json:"file"`    // the file's name in .bonsai/local/log/
	Session    *string       `json:"session"` // a session's file; null for a day file
	Day        *string       `json:"day"`     // a day file's UTC date; null for a session's file
	First      *string       `json:"first"`   // the first readable record's time, or null
	Last       *string       `json:"last"`
	Records    int64         `json:"records"`    // how many records were read
	Unreadable int64         `json:"unreadable"` // how many lines were skipped
	Ended      *bool         `json:"ended"`      // a session's file: whether its last span ended; null for a day file
	Task       *string       `json:"task"`
	Role       *string       `json:"role"`
	Extra      schema.Object `json:"-"`
}

// ReadLogs reads bonsai.logs/1's document.
func ReadLogs(raw []byte) (*Logs, error) {
	l := &Logs{}
	if err := MustLookup("logs").readJSONInto(raw, l); err != nil {
		return nil, err
	}
	return l, nil
}

// Encode writes the document, held to bonsai.logs/1.
func (l *Logs) Encode() ([]byte, error) { return MustLookup("logs").Encode(l) }
