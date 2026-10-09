package workspace

// This machine's record of a workspace's main checkout, <machine folder>/workspace.json (contract §3: "the main
// checkout's path, and the workspace ids it has held"). bonsai init and update write it, in the main checkout only,
// each time they end without a refusal (written or nothing to change); nothing else writes it, and it never enters a
// project. bonsai check reads it for two things: bonsai.yaml's id no longer the one recorded (the id-changed finding)
// and another checkout on this machine holding the same id (the same-id warning). From step 5.3 the guard trusts a
// main checkout only when this record names it and is the first on this machine to claim its id (design/plan-5.md,
// 5.3.1), which is why each id keeps the time this machine first recorded it.
//
// The file, a JSON object written byte-stable (two-space indent, LF, ASCII), every field always written:
//
//	{
//	  "path": "/srv/projects/example",       the main checkout's real path, forward slashes (it never leaves the machine)
//	  "ids": [                               the ids the checkout has held here, oldest first; the last is the current one
//	    {"id": "ws-7kq2m4xw5r3t6y2u7p4a5c3e2b", "since": "2026-10-09T12:00:00Z"}
//	  ]
//	}
//
// since is when this machine first recorded the id for the checkout, UTC, to the second (RFC 3339). An id is added at
// the end when bonsai.yaml's differs from the last one (init --new-id; a person's change of the id, once update runs);
// an entry is never changed or taken out. A key this Bonsai does not know is kept as read. It has no format line and
// no schema in formats/: like the machine folder's settings.json it is Bonsai's own file on one machine, read by
// Bonsai alone.

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/LastStep/Bonsai/internal/schema"
)

// MachineRecordFile is the machine folder's record of the main checkout.
const MachineRecordFile = "workspace.json"

// MachineRecord is <machine folder>/workspace.json.
type MachineRecord struct {
	Path  string        // the main checkout's real path, forward slashes
	IDs   []RecordedID  // the ids it has held on this machine, oldest first
	Extra schema.Object // keys this Bonsai does not know, kept as read
}

// RecordedID is one id a checkout has held on this machine, and when this machine first recorded it.
type RecordedID struct {
	ID    string
	Since string // RFC 3339, UTC
}

// Current is the id the record holds now (its last), "" for none.
func (m *MachineRecord) Current() string {
	if m == nil || len(m.IDs) == 0 {
		return ""
	}
	return m.IDs[len(m.IDs)-1].ID
}

// readMachineRecord reads a record file; a missing one is nil, no error.
func readMachineRecord(path string) (*MachineRecord, error) {
	raw, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	show := filepath.ToSlash(path)
	next := "delete " + show + ": bonsai update in the main checkout writes it again"
	if err != nil {
		return nil, &Error{File: show, Msg: "cannot be read: " + oneLine(err), Err: err, Next: "check the file's permissions"}
	}
	v, err := schema.Decode(bytes.TrimPrefix(raw, []byte("\xef\xbb\xbf")))
	o, isObject := v.(schema.Object)
	if err != nil || !isObject {
		return nil, &Error{File: show, Msg: "is not a JSON object Bonsai reads", Next: next}
	}
	m := &MachineRecord{}
	for _, f := range o {
		switch f.Key {
		case "path":
			s, ok := f.Value.(string)
			if !ok || s == "" {
				return nil, &Error{File: show, Msg: "its path is " + schema.Show(f.Value) + ", not a folder", Next: next}
			}
			m.Path = s
		case "ids":
			list, ok := f.Value.([]any)
			if !ok {
				return nil, &Error{File: show, Msg: "its ids is " + schema.Show(f.Value) + ", not a list", Next: next}
			}
			for _, it := range list {
				io, _ := it.(schema.Object)
				id, since := io.String("id"), io.String("since")
				if id == "" {
					return nil, &Error{File: show, Msg: "an ids entry " + schema.Show(it) + " has no id", Next: next}
				}
				m.IDs = append(m.IDs, RecordedID{ID: id, Since: since})
			}
		default:
			m.Extra = append(m.Extra, f)
		}
	}
	return m, nil
}

// LoadMachineRecord reads this machine's record of the main checkout main: nil when there is none.
func LoadMachineRecord(home, main string) (*MachineRecord, error) {
	dir, err := MachineDir(home, main)
	if err != nil {
		return nil, err
	}
	return readMachineRecord(filepath.Join(dir, MachineRecordFile))
}

// Encode writes the record byte-stable.
func (m *MachineRecord) Encode() ([]byte, error) {
	ids := []any{}
	for _, id := range m.IDs {
		ids = append(ids, schema.Object{{Key: "id", Value: id.ID}, {Key: "since", Value: id.Since}})
	}
	o := append(schema.Object{{Key: "path", Value: m.Path}, {Key: "ids", Value: ids}}, m.Extra...)
	return schema.Encode(o)
}

// RecordCheckout writes this machine's record of the main checkout main: its real path, and id at the end of its
// ids unless it is already the last, since at. It writes nothing when the record already says so. It reports
// whether it wrote.
func RecordCheckout(home, main, id string, at time.Time) (bool, error) {
	dir, err := MachineDir(home, main)
	if err != nil {
		return false, err
	}
	key := filepath.Join(dir, MachineRecordFile)
	m, err := readMachineRecord(key)
	if err != nil {
		var we *Error
		if !errors.As(err, &we) || we.Err != nil {
			return false, err // the file cannot be read at all
		}
		m = nil // a record Bonsai does not read is written again: it is Bonsai's own
	}
	real, err := filepath.EvalSymlinks(main)
	if err != nil {
		return false, &Error{File: filepath.ToSlash(main), Msg: "cannot be resolved: " + oneLine(err), Err: err,
			Next: "check that the main checkout exists and can be read"}
	}
	if abs, err := filepath.Abs(real); err == nil {
		real = abs
	}
	path := filepath.ToSlash(real)
	if m == nil {
		m = &MachineRecord{}
	}
	if m.Path == path && m.Current() == id {
		return false, nil
	}
	m.Path = path
	if m.Current() != id {
		m.IDs = append(m.IDs, RecordedID{ID: id, Since: at.UTC().Format(time.RFC3339)})
	}
	raw, err := m.Encode()
	if err != nil {
		return false, err
	}
	if err := WriteFileAtomic(key, raw); err != nil {
		return false, &Error{File: filepath.ToSlash(key), Msg: "cannot be written: " + oneLine(err), Err: err,
			Next: "check that the Bonsai home can be written"}
	}
	return true, nil
}

// KeyedRecord is one workspace's record on this machine, with its machine folder's name.
type KeyedRecord struct {
	Key    string // r-<16 hex>
	Record *MachineRecord
}

// MachineRecords reads every workspace's record in the home, by machine folder name; a folder with no record, or one
// Bonsai does not read, is left out.
func MachineRecords(home string) []KeyedRecord {
	entries, err := os.ReadDir(filepath.Join(home, "workspaces"))
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() && strings.HasPrefix(e.Name(), "r-") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	var out []KeyedRecord
	for _, n := range names {
		m, err := readMachineRecord(filepath.Join(home, "workspaces", n, MachineRecordFile))
		if err == nil && m != nil {
			out = append(out, KeyedRecord{Key: n, Record: m})
		}
	}
	return out
}
