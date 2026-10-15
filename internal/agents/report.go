package agents

// What the guard and the stop gate read of the messages and the log about an agent (design/plan-5.md, step 5.5,
// "Past a budget"): its own address, whether bonsai message wrote a message (its message record in the log), and the
// report it wrote past its budget, found as the message file itself (VG's item C, 15 Oct: the stop gate let a named
// session stop whose report command had failed, on a flag the guard set when the call was let through).

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/message"
	"github.com/LastStep/Bonsai/internal/record"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// Address gives an agent's own address as its state holds it (step 5.5's "an agent's own address"): its tied brief's
// to; else, for a named session, <role>@<task> (its role alone with no task); "" for none (an undeclared type, an agent
// with no tied brief, the person's session).
func (st *State) Address() string {
	h := st.Header
	if h.Role == "" || h.SessRole {
		return ""
	}
	if t := st.Window().Tie; t != nil {
		return t.To
	}
	if h.Named {
		if h.Task == "" {
			return h.Role
		}
		return h.Role + "@" + h.Task
	}
	return ""
}

// ReportAfter finds a report written from addr at or after the time after: a message file under the main checkout's
// .bonsai/local/messages/ (5.5.10's paths: <recipient folder>/<id>.md, read only when it is a regular file of at most
// message.MaxFile bytes, through message.Parse), of the kind report, from addr, created at or after it, whose message
// record bonsai message wrote in the log (MessageRecorded; session is the writer's session, read first). It gives the
// report's file, project-relative with forward slashes, "" for none. A file named for a day before after's is not read.
func (s Store) ReportAfter(session, addr string, after time.Time) string {
	if addr == "" {
		return ""
	}
	root := filepath.Join(s.Main, filepath.FromSlash(message.Dir))
	folders, err := os.ReadDir(root)
	if err != nil {
		return ""
	}
	day := after.UTC().Format("2006-01-02")
	type report struct {
		m   *format.Message
		rel string
	}
	var found []report
	for _, f := range folders {
		if strings.HasPrefix(f.Name(), ".") || f.Type() != fs.ModeDir {
			continue
		}
		files, err := os.ReadDir(filepath.Join(root, f.Name()))
		if err != nil {
			continue
		}
		for _, e := range files {
			id, ok := strings.CutSuffix(e.Name(), ".md")
			if !ok || !message.IsID(id) || id[2:12] < day || !e.Type().IsRegular() {
				continue
			}
			raw, err := ReadRegular(filepath.Join(root, f.Name(), e.Name()), message.MaxFile)
			if err != nil {
				continue
			}
			m, _, why := message.Parse(raw, f.Name(), e.Name())
			if why != "" || m.Kind != format.KindReport || m.From != addr {
				continue
			}
			created, err := time.Parse(time.RFC3339Nano, m.Created)
			if err != nil || created.Before(after.Truncate(time.Millisecond)) {
				continue
			}
			found = append(found, report{m, message.Dir + "/" + f.Name() + "/" + e.Name()})
		}
	}
	sort.Slice(found, func(i, j int) bool { return found[i].rel < found[j].rel })
	for _, r := range found {
		created, _ := time.Parse(time.RFC3339Nano, r.m.Created)
		if MessageRecorded(s.Main, session, r.m.ID, created) {
			return r.rel
		}
	}
	return ""
}

// recordScanCap is the most of the log MessageRecorded reads to find a message's record, in bytes.
const recordScanCap = 64 << 20

// MessageRecorded finds the message record bonsai message wrote with a message (contract §8.2: event message, target
// its id), created at created: the session's own log file first, then the day file of its created (a message written
// outside a session), then each other log file written to since it was created, newest first, at most recordScanCap
// bytes in all, each a regular file.
func MessageRecorded(main, session, id string, created time.Time) bool {
	dir := filepath.Join(main, filepath.FromSlash(workspace.LocalDir), record.LogFolder)
	var order []string
	if name, err := record.LogFileName(session, created); err == nil && session != "" {
		order = append(order, name)
	}
	day, _ := record.LogFileName("", created)
	order = append(order, day)
	if entries, err := os.ReadDir(dir); err == nil {
		type cand struct {
			name string
			mod  time.Time
		}
		var rest []cand
		for _, e := range entries {
			n := e.Name()
			if listHas(order, n) || !strings.HasSuffix(n, ".ndjson") || !(strings.HasPrefix(n, "s-") || strings.HasPrefix(n, "w-")) {
				continue
			}
			fi, err := e.Info()
			if err != nil || !fi.Mode().IsRegular() || fi.ModTime().Before(created.Add(-time.Second)) {
				continue // a file last written before the message was made cannot hold its record
			}
			rest = append(rest, cand{n, fi.ModTime()})
		}
		sort.Slice(rest, func(i, j int) bool {
			if !rest[i].mod.Equal(rest[j].mod) {
				return rest[i].mod.After(rest[j].mod)
			}
			return rest[i].name < rest[j].name
		})
		for _, c := range rest {
			order = append(order, c.name)
		}
	}
	budget := int64(recordScanCap)
	needle := []byte(`"` + id + `"`)
	for _, name := range order {
		if budget <= 0 {
			return false
		}
		found, read := scanForRecord(filepath.Join(dir, name), needle, id, budget)
		if found {
			return true
		}
		budget -= read
	}
	return false
}

// listHas reports a string in a list.
func listHas(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

// scanForRecord reads one log file, a regular one, at most budget bytes, for a message record whose target is id:
// found, and the bytes read.
func scanForRecord(path string, needle []byte, id string, budget int64) (found bool, read int64) {
	f, _, err := openRegular(path)
	if err != nil {
		return false, 0
	}
	defer func() { _ = f.Close() }()
	r := bufio.NewReaderSize(io.LimitReader(f, budget), 64<<10)
	for {
		line, err := r.ReadSlice('\n')
		read += int64(len(line))
		if errors.Is(err, bufio.ErrBufferFull) {
			for errors.Is(err, bufio.ErrBufferFull) { // a line longer than any record: passed over
				line, err = r.ReadSlice('\n')
				read += int64(len(line))
			}
			line = nil
		}
		if len(line) > 0 && bytes.Contains(line, needle) {
			var rec struct {
				Event  string  `json:"event"`
				Target *string `json:"target"`
			}
			if json.Unmarshal(line, &rec) == nil && rec.Event == "message" && rec.Target != nil && *rec.Target == id {
				return true, read
			}
		}
		if err != nil {
			return false, read
		}
	}
}
