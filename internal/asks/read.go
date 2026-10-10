package asks

// Reading the asks back (contract §9.1; design/plan-5.md, 5.2.5 note 4): every record of the day files in
// .bonsai/local/asks/, the files by name (their UTC day) and each file's lines in order, which is the order they were
// appended. A line that does not read as a bonsai.ask/1 record (a torn last line after a crash, a line written by
// hand, a later major) is skipped and counted, never fatal; a record glued after a torn line is still read, at its
// {"format":"bonsai.ask/, as internal/record's reader does for the log.
//
// A key's state is its latest record: a file record opens it (filing an open key again, the newest wording stands;
// filing a closed key opens it again), and a resolve or answer record closes it, resolved or answered. A record that
// would close a key already closed changes nothing, so when two answers race, the first answer stands. A key with no
// file record (a hand-written answer, say) is not an ask and is not listed.

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"time"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/record"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// dayFile is how an asks file is named: its UTC day (contract §9.1).
var dayFile = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}\.ndjson$`)

// recordStart is how every ask record Bonsai writes begins.
var recordStart = []byte(`{"format":"bonsai.ask/`)

// Entry is one key and where it stands.
type Entry struct {
	Key    string
	State  string           // Open, Answered or Resolved
	Filed  *format.Ask      // the key's latest file record
	Closed *format.Ask      // the answer or resolve record that closed it; nil while open
	Docs   [2]schema.Object // Filed and Closed as stored (bonsai.asks/1 prints them so); Docs[1] nil while open
}

// Entry gives the entry as bonsai.asks/1 prints it.
func (e *Entry) Entry() format.AskEntry {
	return format.AskEntry{Key: e.Key, State: e.State, Filed: e.Docs[0], Closed: e.Docs[1]}
}

// Book is what the asks folder holds: each key's entry, and how many lines did not read.
type Book struct {
	byKey      map[string]*Entry
	Unreadable int
}

// Get gives a key's entry, nil when the key has no ask.
func (b *Book) Get(key string) *Entry { return b.byKey[key] }

// Entries gives every key's entry (only the open ones unless all), newest first: by the time of their latest file
// record, latest first, then by key.
func (b *Book) Entries(all bool) []*Entry {
	out := []*Entry{}
	for _, e := range b.byKey {
		if all || e.State == Open {
			out = append(out, e)
		}
	}
	sort.Slice(out, func(i, j int) bool { // keys are distinct
		if a, b := when(out[i].Filed.At), when(out[j].Filed.At); !a.Equal(b) {
			return a.After(b)
		}
		return out[i].Key < out[j].Key
	})
	return out
}

// when reads a record's at (milliseconds optional, as the schema allows); the zero time when it does not read.
func when(at string) time.Time {
	t, _ := time.Parse(time.RFC3339Nano, at)
	return t
}

// Folder is the asks folder of the main checkout main.
func Folder(main string) string {
	return filepath.Join(main, filepath.FromSlash(workspace.LocalDir), record.AsksFolder)
}

// Read reads the asks folder of the main checkout main. With only set, it reads that key's records alone (a line that
// does not hold the key in quotes is passed over unread, and not counted). A folder that is not there holds no asks.
func Read(main, only string) (*Book, error) {
	b := &Book{byKey: map[string]*Entry{}}
	dir := Folder(main)
	entries, err := os.ReadDir(dir)
	if errors.Is(err, fs.ErrNotExist) {
		return b, nil
	}
	if err != nil {
		return nil, err
	}
	var mark []byte
	if only != "" {
		mark = []byte(`"` + only + `"`) // a key needs no escaping in JSON
	}
	for _, de := range entries { // os.ReadDir sorts by name: the days in order
		if !de.Type().IsRegular() || !dayFile.MatchString(de.Name()) {
			continue
		}
		if err := b.readFile(filepath.Join(dir, de.Name()), mark); err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue // removed since it was listed
			}
			return nil, err
		}
	}
	return b, nil
}

func (b *Book) readFile(path string, mark []byte) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	r := bufio.NewReaderSize(f, 64<<10)
	for {
		line, err := r.ReadBytes('\n')
		if len(line) > 0 && (mark == nil || bytes.Contains(line, mark)) {
			b.take(line)
		}
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

// take reads one line as a record and moves its key's state on.
func (b *Book) take(line []byte) {
	line = bytes.TrimSuffix(bytes.TrimSuffix(line, []byte("\n")), []byte("\r"))
	if len(bytes.TrimSpace(line)) == 0 {
		return
	}
	rec, doc := readRecord(line)
	if rec == nil {
		if i := bytes.LastIndex(line, recordStart); i > 0 {
			rec, doc = readRecord(line[i:])
		}
		b.Unreadable++
		if rec == nil {
			return
		}
	}
	e := b.byKey[rec.Key]
	switch rec.Op {
	case OpFile:
		if e == nil {
			e = &Entry{Key: rec.Key}
			b.byKey[rec.Key] = e
		}
		e.State, e.Filed, e.Closed, e.Docs = Open, rec, nil, [2]schema.Object{doc, nil}
	case OpResolve, OpAnswer:
		if e == nil || e.State != Open {
			return // no ask yet, or closed already: the first close stands
		}
		e.State, e.Closed, e.Docs[1] = Resolved, rec, doc
		if rec.Op == OpAnswer {
			e.State = Answered
		}
	}
}

// readRecord reads one line as an ask record and its document as stored; nil when it is not one.
func readRecord(line []byte) (*format.Ask, schema.Object) {
	f := format.MustLookup("ask")
	doc, err := f.ReadJSON(line)
	if err != nil {
		return nil, nil
	}
	a := &format.Ask{}
	if err := f.Bind(doc, a); err != nil {
		return nil, nil
	}
	return a, doc
}
