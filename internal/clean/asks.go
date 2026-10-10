package clean

// The asks kind (contract §9.1): <UTC day>.ndjson in .bonsai/local/asks/, aged by its last record that reads, and
// kept while it holds what an ask still open, or one that stays, needs.
//
// A key's state is its latest record (internal/asks, read.go): a file record opens it, and the first answer or
// withdrawal after it closes it. Taking whole day files out of that history must leave every key that stays as it
// was, so two protections, judged over every candidate at once:
//   - open: the day file holding an open ask's latest file record (its state is asks.Read's, the one reader of a key's
//     state) is kept;
//   - closed stays closed: when a day file that stays holds a key's file record, every day file holding that key's
//     answers or withdrawals after it stays too, else the ask would read open again. Kept files can keep more, so this
//     runs until nothing changes.
//
// Records are read here as asks.Read reads them (the day files by name, each file's lines in order, a line that does
// not read skipped, a record glued after a torn line still read), to know which file holds which record; asks.Read
// itself gives the state. An open key whose file record is not found here (a write between the two reads) keeps every
// asks file this run.

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/LastStep/Bonsai/internal/asks"
	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/record"
)

// asksDayName is how asks.Read finds a day file (its dayFile).
var asksDayName = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}\.ndjson$`)

var asksKind = fileKind{
	name:   Asks,
	folder: record.AsksFolder,
	match: func(name string) bool {
		if !asksDayName.MatchString(name) {
			return false
		}
		_, err := time.Parse("2006-01-02.ndjson", name)
		return err == nil
	},
	age:        asksAge,
	protectAll: asksProtected,
}

// askRec is one ask record, and the day file holding it.
type askRec struct {
	name string // the day file's name
	op   string
	key  string
}

// asksHistory is every ask record in the folder, in asks.Read's order, and each file's last record's time.
type asksHistory struct {
	recs []askRec
	last map[string]time.Time // a day file's last record that reads, by name
}

// asksAge is a day file's last record's at, its modification time when none reads.
func asksAge(r *run, f *file) time.Time {
	h, err := r.asksHistory()
	if err == nil {
		if at, ok := h.last[f.name]; ok {
			return at
		}
	}
	return f.mtime
}

// asksHistory reads the asks folder's records once a run.
func (r *run) asksHistory() (*asksHistory, error) {
	if r.asks != nil {
		return r.asks, r.asksErr
	}
	r.asks = &asksHistory{last: map[string]time.Time{}}
	dir := filepath.Join(r.localDir(), record.AsksFolder)
	entries, err := os.ReadDir(dir)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		r.asksErr = err
		return r.asks, err
	}
	for _, de := range entries { // os.ReadDir sorts by name: the days in order
		if !de.Type().IsRegular() || !asksDayName.MatchString(de.Name()) {
			continue
		}
		if err := r.asks.readFile(filepath.Join(dir, de.Name()), de.Name()); err != nil && !errors.Is(err, fs.ErrNotExist) {
			r.asksErr = err
			return r.asks, err
		}
	}
	return r.asks, nil
}

func (h *asksHistory) readFile(path, name string) error {
	fh, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = fh.Close() }()
	br := bufio.NewReaderSize(fh, 64<<10)
	for {
		line, err := br.ReadBytes('\n')
		if rec := readAskLine(line); rec != nil {
			h.recs = append(h.recs, askRec{name: name, op: rec.Op, key: rec.Key})
			if at, perr := time.Parse(time.RFC3339Nano, rec.At); perr == nil {
				h.last[name] = at
			}
		}
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

// askRecordStart is how every ask record Bonsai writes begins.
var askRecordStart = []byte(`{"format":"bonsai.ask/`)

// readAskLine reads one line as an ask record as asks.Read does: the line whole, else the record glued after a torn
// part; nil when neither reads.
func readAskLine(line []byte) *format.Ask {
	line = bytes.TrimSuffix(bytes.TrimSuffix(line, []byte("\n")), []byte("\r"))
	if len(bytes.TrimSpace(line)) == 0 {
		return nil
	}
	if rec := readAsk(line); rec != nil {
		return rec
	}
	if i := bytes.LastIndex(line, askRecordStart); i > 0 {
		return readAsk(line[i:])
	}
	return nil
}

func readAsk(line []byte) *format.Ask {
	f := format.MustLookup("ask")
	doc, err := f.ReadJSON(line)
	if err != nil {
		return nil
	}
	a := &format.Ask{}
	if err := f.Bind(doc, a); err != nil {
		return nil
	}
	return a
}

// asksProtected gives the candidates that stay: the day file of an open ask's latest file record, and every file
// holding an answer or withdrawal after a file record that stays (this file's comment).
func asksProtected(r *run, all, cands []*file) (map[*file]bool, error) {
	h, err := r.asksHistory()
	if err != nil {
		return nil, fmt.Errorf("clean: the asks cannot be read, so none is cleaned: %w", err)
	}
	book, err := asks.Read(r.main, "")
	if err != nil {
		return nil, fmt.Errorf("clean: the asks cannot be read, so none is cleaned: %w", err)
	}
	going := map[string]*file{}
	for _, f := range cands {
		going[f.name] = f
	}
	stays := func(name string) bool { return going[name] == nil }
	kept := map[string]bool{}
	keep := func(name string) bool {
		if stays(name) || kept[name] {
			return false
		}
		kept[name] = true
		return true
	}
	// An open ask's latest file record.
	latest := map[string]string{}
	for _, rec := range h.recs {
		if rec.op == asks.OpFile {
			latest[rec.key] = rec.name
		}
	}
	for _, e := range book.Entries(false) {
		name, ok := latest[e.Key]
		if !ok {
			return nil, fmt.Errorf("clean: the open ask %s has no file record the cleaner can find, so no asks file is cleaned this run", e.Key)
		}
		keep(name)
	}
	// A closed ask stays closed: the closes after a file record that stays, until nothing changes.
	for changed := true; changed; {
		changed = false
		filed := map[string]bool{} // keys with a file record in a file that stays, so far in the history
		for _, rec := range h.recs {
			switch rec.op {
			case asks.OpFile:
				if stays(rec.name) || kept[rec.name] {
					filed[rec.key] = true
				}
			case asks.OpAnswer, asks.OpResolve:
				if filed[rec.key] && keep(rec.name) {
					changed = true
				}
			}
		}
	}
	out := map[*file]bool{}
	for _, f := range cands {
		if kept[f.name] {
			out[f] = true
		}
	}
	return out, nil
}
