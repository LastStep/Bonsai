package clean

// The messages kind (contract §22; design/plan-5.md, "Messages between agents", note 5.5.10, 5): the messages between
// agents in the main checkout's .bonsai/local/messages/, cleaned by thread, never by file, since a thread's messages
// lie in its two ends' folders. They are read and folded by internal/message, the one reader every command uses.
//
//   - A thread goes once its last message (its newest created, the messages' own clock) is older than keep_days, or it
//     is beyond the newest keep_newest threads (by their last message, then their id), and it is not open.
//   - An open thread (its newest brief with no report after it) is never cleaned, whatever its age, and still counts
//     among the newest (the director, 14 Oct: "A. Local, 30 days"): it stays listed until a report answers it.
//   - Only messages that read are cleaned, each its own file under its recipient's folder. A file that does not read
//     (bonsai messages lists it as malformed, check warns of it), a staged write (.<id>.md.tmp-*), and anything else
//     there are left alone and never named, as are the recipients' folders themselves (a writer may be about to write
//     into one).
//   - A thread's briefs go first, then its reports, each oldest first, so a run stopped part-way (a file held open, one
//     changed, a record that cannot be written) leaves a thread that still reads answered, to go at a later run.
//   - Reading the messages is one step of the budget (files.go's comment): a thread spans folders, so the folder is
//     read whole; each thread judged and deleted is another.
//
// Each file deleted writes its clean record (target its project-relative path; reason generated.messages.keep_days=30
// or keep_newest), through remove, as every kind's.

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/message"
	"github.com/LastStep/Bonsai/internal/record"
)

// messages runs the messages kind: its rule, its threads judged, and the files of each that goes deleted.
func (r *run) messages() {
	rule, err := RuleOf(r.cfg.Doc, Messages)
	if err != nil {
		r.fail(err)
		return
	}
	if rule.Keeps() || r.out() {
		return
	}
	for _, p := range []string{filepath.Join(r.main, ".bonsai"), r.localPath(""), r.localPath(message.Folder)} {
		fi, err := os.Lstat(p)
		switch {
		case errors.Is(err, fs.ErrNotExist):
			return
		case err != nil:
			r.fail(err)
			return
		case fi.Mode().Type() != fs.ModeDir:
			return // a link, a junction or a file where a folder should be: left alone
		}
	}
	st, err := message.Read(r.main)
	if err != nil {
		r.fail(err)
		return
	}
	threads := message.Fold(st, message.DefaultRules(), r.now) // the state alone is read: no threshold matters
	sort.Slice(threads, func(i, j int) bool {                  // oldest last message first; ids are distinct
		if threads[i].LastAt != threads[j].LastAt {
			return threads[i].LastAt < threads[j].LastAt
		}
		return threads[i].ID < threads[j].ID
	})
	beyond := map[string]bool{}
	if rule.KeepNewest != nil {
		for i := range threads {
			if int64(len(threads)-1-i) >= *rule.KeepNewest {
				beyond[threads[i].ID] = true
			}
		}
	}
	cutoff, byAge := rule.cutoff(r.now)
	for _, t := range threads {
		if r.out() || r.halt {
			return
		}
		if t.State == message.Open {
			continue
		}
		var why string
		switch last, err := time.Parse(record.AtLayout, t.LastAt); {
		case byAge && err == nil && last.Before(cutoff):
			why = reason(Messages, "keep_days", *rule.KeepDays)
		case beyond[t.ID]:
			why = reason(Messages, "keep_newest", *rule.KeepNewest)
		default:
			continue
		}
		r.thread(t, why)
	}
}

// thread deletes one thread's files, briefs first and then reports, each oldest first (Entries' order), stopping at
// the first that stays.
func (r *run) thread(t message.Thread, why string) {
	var order []message.Entry
	for _, kind := range []string{format.KindBrief, format.KindReport} {
		for _, e := range t.Entries {
			if e.Msg.Kind == kind {
				order = append(order, e)
			}
		}
	}
	for _, e := range order {
		path := filepath.Join(r.main, filepath.FromSlash(e.Path))
		fi, err := os.Lstat(path)
		if errors.Is(err, fs.ErrNotExist) {
			continue // another cleaner took it
		}
		if err != nil {
			r.fail(err)
			return
		}
		f := &file{kind: Messages, name: filepath.Base(path), path: path, target: e.Path, mtime: fi.ModTime(), size: fi.Size(), reason: why}
		if r.keep[f.target] || !r.remove(f) || r.halt {
			return
		}
	}
}
