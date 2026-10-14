package clean

// The agents kind (format.GeneratedKinds' agents; design/plan-5.md, note 5.5.11, 9): the guard's state of each agent
// in the main checkout's .bonsai/local/agents/ (internal/agents), cleaned by session folder, and each brief's tie file
// with its thread.
//
//   - A session's folder goes once its newest line (agents.LastLine: each agent's last call line, each state's last
//     event, the last spawn call) is older than keep_days, or it is beyond the newest keep_newest folders (by their
//     newest line, then their name), as the log's files go.
//   - Never while that session's log file is kept (.bonsai/local/log/s-<session>.ndjson: the log goes first, so a
//     session's agents stay as long as its log does), nor while a brief one of its agents tied is kept (its message
//     file is still there, read or not: a thread goes whole, internal/clean's messages kind).
//   - A brief's tie file goes once its brief's message file is gone (its thread is cleaned) and the tie is older than
//     keep_days (its modification time, the time of the tie: a tie file is never rewritten); never while the brief is
//     there. keep_newest counts session folders only.
//   - Only Bonsai's own files (agents.SessionFiles, agents.Ties): each a regular file in a plain folder, links never
//     followed; anything else is left alone and never named. A session folder's files go one at a time, each with its
//     clean record, the folder itself (and its claims/) removed once empty, with no record.
//   - The ties go first, then the session folders, oldest first; the budget is checked before each file, so a run ends
//     within it and one step, and a folder it leaves part-way is still old, its log gone: it goes at a later run.

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/LastStep/Bonsai/internal/agents"
	"github.com/LastStep/Bonsai/internal/message"
	"github.com/LastStep/Bonsai/internal/record"
)

// agents runs the agents kind: its rule, its ties and session folders judged, and each that goes deleted.
func (r *run) agents() {
	rule, err := RuleOf(r.cfg.Doc, Agents)
	if err != nil {
		r.fail(err)
		return
	}
	if rule.Keeps() || r.out() {
		return
	}
	store := agents.Store{Main: r.main}
	for _, p := range []string{filepath.Join(r.main, ".bonsai"), r.localPath(""), store.Dir()} {
		if _, err := os.Lstat(p); errors.Is(err, fs.ErrNotExist) {
			return
		}
		if !agents.Plain(p) {
			return // a link, a junction or a file where a folder should be: left alone
		}
	}
	st, err := message.Read(r.main)
	if err != nil {
		r.fail(err)
		return
	}
	briefKept := func(id string) bool {
		e, m := st.Find(id)
		return e != nil || m != nil
	}
	cutoff, byAge := rule.cutoff(r.now)
	held := map[string]bool{} // the sessions one of whose agents tied a brief still kept
	ties, err := store.Ties()
	if err != nil {
		r.fail(err)
		return
	}
	for _, t := range ties {
		if r.out() || r.halt {
			return
		}
		if briefKept(t.Brief) {
			if t.Session != "" {
				held[t.Session] = true
			}
			continue
		}
		path := store.TiePath(t.Brief)
		fi, err := os.Lstat(path)
		if err != nil || !fi.Mode().IsRegular() || !byAge || !fi.ModTime().Before(cutoff) {
			continue
		}
		f := &file{kind: Agents, name: t.Brief, path: path, target: target(agents.Folder, agents.TiesFolder+"/"+t.Brief),
			mtime: fi.ModTime(), size: fi.Size(), reason: reason(Agents, "keep_days", *rule.KeepDays)}
		if !r.keep[f.target] {
			r.remove(f)
		}
	}
	sessions, err := store.Sessions()
	if err != nil {
		r.fail(err)
		return
	}
	type folder struct {
		name string
		age  time.Time
	}
	logDir := r.localPath(record.LogFolder)
	logKept := func(session string) bool {
		_, err := os.Lstat(filepath.Join(logDir, "s-"+session+".ndjson"))
		return !errors.Is(err, fs.ErrNotExist) // kept, or it cannot be told
	}
	var all []folder
	for _, s := range sessions {
		if r.out() {
			return // the ranking is not whole: nothing of this kind is judged on it
		}
		age, ok := store.LastLine(s)
		if !ok {
			// None of Bonsai's files: a folder a run left empty goes once its log has (a session still writing keeps
			// its log, and may be about to write its first state here).
			if !logKept(s) {
				_ = os.Remove(filepath.Join(store.SessionDir(s), agents.ClaimsFolder))
				_ = os.Remove(store.SessionDir(s))
			}
			continue
		}
		all = append(all, folder{s, age})
	}
	sort.Slice(all, func(i, j int) bool { // oldest first; names are distinct
		if !all[i].age.Equal(all[j].age) {
			return all[i].age.Before(all[j].age)
		}
		return all[i].name < all[j].name
	})
	for i, f := range all {
		if r.out() || r.halt {
			return
		}
		var why string
		switch {
		case byAge && f.age.Before(cutoff):
			why = reason(Agents, "keep_days", *rule.KeepDays)
		case rule.KeepNewest != nil && int64(len(all)-1-i) >= *rule.KeepNewest:
			why = reason(Agents, "keep_newest", *rule.KeepNewest)
		default:
			continue
		}
		if held[f.name] || logKept(f.name) {
			continue // a brief one of its agents tied is kept, or its log file is: its agents stay with them
		}
		r.sessionFolder(store, f.name, why)
	}
}

// sessionFolder deletes one session folder's files, each with its clean record, stopping at the first that stays;
// then its claims/ and the folder itself, once empty.
func (r *run) sessionFolder(store agents.Store, session, why string) {
	files, err := store.SessionFiles(session)
	if err != nil {
		r.fail(err)
		return
	}
	dir := store.SessionDir(session)
	for _, rel := range files {
		if r.out() || r.halt {
			return
		}
		path := filepath.Join(dir, filepath.FromSlash(rel))
		fi, err := os.Lstat(path)
		if errors.Is(err, fs.ErrNotExist) {
			continue // another cleaner took it
		}
		if err != nil {
			r.fail(err)
			return
		}
		f := &file{kind: Agents, name: filepath.Base(path), path: path, target: target(agents.Folder, session+"/"+rel),
			mtime: fi.ModTime(), size: fi.Size(), reason: why}
		if r.keep[f.target] || !r.remove(f) || r.halt {
			return
		}
	}
	_ = os.Remove(filepath.Join(dir, agents.ClaimsFolder)) // only once empty; anything else there keeps it
	_ = os.Remove(dir)
}
