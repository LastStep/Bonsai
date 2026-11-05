package sessions

// The sessions cache (BUG-137): what each log file's records say, kept between runs so that bonsai check and bonsai
// status read again only the log files that changed since the last run. A large log (tens of MB) took seconds to read
// on every run; with the cache, a run with nothing changed reads none of it.
//
// The cache is .bonsai/local/cache/sessions.json in the main checkout (CacheFile), never committed (.bonsai/local/ is
// ignored whole). It is a convenience and never a record: Bonsai writes it, nothing reads it but FoundCached, and a
// cache that is missing, does not read, was written by another version of this code, or names a file whose size or
// modification time is not the one it holds is passed over and the log file read again. Deleting it at any time is
// safe; the next run writes it again.
//
// What it keeps for each session file is parse's answer, which depends only on the file's bytes and its name: the
// spans as the records alone say (before the 24-hour rule, which needs the clock and is applied on every run) and
// the last record's time. Its bytes are stable: files in name order, fields in a fixed order, no map.
//
// An example, one file:
//
//	{"version":1,"files":[{"name":"s-6d1e2f3a-4b5c-4d6e-8f7a-9b0c1d2e3f4a.ndjson","size":2048,
//	 "mtime":1760000000000000000,"last":"2026-10-08T09:30:00Z","spans":[{"session":"6d1e2f3a-...",
//	 "subagent":"","start":"2026-10-08T09:05:00Z","end":"2026-10-08T09:30:00Z","open":false,"task":"T-1234",
//	 "role":"builder","model":"opus"}]}]}

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/record"
)

// CacheFile is the sessions cache's path, relative to the main checkout's .bonsai/local/.
const CacheFile = "cache/sessions.json"

// cacheVersion is the cache's own version. A change to what parse gives, or to the cache's fields, raises it, and a
// cache of another version is passed over whole.
const cacheVersion = 1

type cacheData struct {
	Version int         `json:"version"`
	Files   []cacheFile `json:"files"`
}

// cacheFile is one log file: its name, the size and modification time (in nanoseconds since 1970, UTC) it had when it
// was read, and what parse gave.
type cacheFile struct {
	Name  string      `json:"name"`
	Size  int64       `json:"size"`
	MTime int64       `json:"mtime"`
	Last  time.Time   `json:"last"`
	Spans []cacheSpan `json:"spans"`
}

type cacheSpan struct {
	Session  string    `json:"session"`
	Subagent string    `json:"subagent"`
	Start    time.Time `json:"start"`
	End      time.Time `json:"end"`
	Open     bool      `json:"open"`
	Task     string    `json:"task"`
	Role     *string   `json:"role"`
	Model    *string   `json:"model"`
}

// FoundCached is Found with the sessions cache: cache is the cache file's bytes (nil or anything that does not read
// as a cache of this version: none), and next is the cache to write for the next run, nil when it would be the same
// bytes as cache, or when there is neither a session file nor a cache (a checkout with no log, as in CI, gets no
// cache file). The rows are Found's, the same in every case: a log file is read again unless the cache holds it
// at the size and modification time it has now, which is read before the file is.
func FoundCached(dir string, cache []byte, now time.Time) (rows []format.SessionRow, next []byte, err error) {
	files, err := record.ListLog(dir)
	if err != nil {
		return nil, nil, err
	}
	held := map[string]cacheFile{}
	var old cacheData
	if json.Unmarshal(cache, &old) == nil && old.Version == cacheVersion {
		for _, f := range old.Files {
			held[f.Name] = f
		}
	}
	out := cacheData{Version: cacheVersion, Files: []cacheFile{}}
	for _, name := range files.Sessions { // in name order
		path := filepath.Join(dir, name)
		fi, err := os.Stat(path)
		if err != nil {
			return nil, nil, err
		}
		cf, ok := held[name]
		if !ok || cf.Size != fi.Size() || cf.MTime != fi.ModTime().UnixNano() {
			lf, err := record.ReadLog(path)
			if err != nil {
				return nil, nil, err
			}
			f, last := parse(lf, record.SessionOf(name))
			cf = toCache(name, fi, f, last)
		}
		out.Files = append(out.Files, cf)
		f := fromCache(cf)
		f.settle(cf.Last, now)
		for _, s := range f.Spans {
			if s.Open {
				continue
			}
			if r, ok := rowOf(s); ok {
				rows = append(rows, r)
			}
		}
	}
	rows = sortRows(dedupe(rows, nil))
	b, err := json.Marshal(out)
	if err != nil {
		return nil, nil, err
	}
	b = append(b, '\n')
	if bytes.Equal(b, cache) || (len(out.Files) == 0 && len(cache) == 0) {
		b = nil
	}
	return rows, b, nil
}

func toCache(name string, fi os.FileInfo, f *File, last time.Time) cacheFile {
	cf := cacheFile{Name: name, Size: fi.Size(), MTime: fi.ModTime().UnixNano(), Last: last.UTC(), Spans: []cacheSpan{}}
	for _, s := range f.Spans {
		cf.Spans = append(cf.Spans, cacheSpan{Session: s.Session, Subagent: s.Subagent, Start: s.Start.UTC(), End: s.End.UTC(),
			Open: s.Open, Task: s.Task, Role: s.Role, Model: s.Model})
	}
	return cf
}

// fromCache gives the file parse gave, from its cache entry: its spans, a fresh copy, so settle changes no entry.
func fromCache(cf cacheFile) *File {
	f := &File{Spans: make([]Span, 0, len(cf.Spans))}
	for _, s := range cf.Spans {
		f.Spans = append(f.Spans, Span{Session: s.Session, Subagent: s.Subagent, Start: s.Start, End: s.End, Open: s.Open,
			Task: s.Task, Role: s.Role, Model: s.Model})
	}
	return f
}
