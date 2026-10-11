package clean

// A file kind's run (log, asks, ladder): its folder listed, each file's age read, and each judged and deleted in
// order.
//
// The order: by modification time, oldest first, then by name, so what a budget leaves for the next run is the
// newest. The modification time only orders the files and spares reads: a file modified within keep_days is not
// read for keep_days, since Bonsai writes each line a moment after its record's at, so such a file's own clock is
// within keep_days too (or, for a file touched since, it is kept until its modification time ages as well: the safe
// side). Every file that is read is judged by its own clock (the kind's age).
//
// Log and ladder files are judged one at a time, each deleted as soon as it is judged, so a run the budget cuts short
// has still cleaned what it judged. Asks are judged all at once (asks.go), since whether one day file may go depends on
// which others stay; they go in day order, and a run stops deleting them at the first that does not go (the budget, a
// file held open or changed since it was listed, a delete that fails), so the day files that stay are always a set
// the protections allow: an ask whose filing stays keeps its answer, wherever the run stopped.
//
// The budget (Options.Budget) is checked before each step, and a step once begun is finished. The steps: a batch of
// listBatch entries of a folder (a folder is read a batch at a time, never whole at once); one file's age read for
// keep_newest's ranking; one asks day file read for the asks' history (asks.go); one file judged, deleted and
// recorded. A read a run makes once is part of the step that first needs it (the task files, the sessions table, the
// asks as internal/asks reads them). So a run ends within its budget and one step. A run the budget stops leaves each
// file deleted with its clean record, or as it was: nothing is half-done. It says Stopped, which is not an error, and
// the next run (the next session's end) lists the folder again from the start, oldest first, and goes on from there.
//
// keep_newest ranks every file of the kind by its own clock, so every file's age is read before any file is judged.
// A ranking the budget cuts short is dropped, and nothing of that kind is deleted by that run, by either rule: no
// file is judged on a ranking that is not whole, so none that might be among the newest goes. A folder too large to
// rank within one run's budget is therefore not cleaned of that kind while it stays that large (at a session's end,
// each run stops in the ranking): keep_newest suits a number of files a run can read within its budget.

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/LastStep/Bonsai/internal/workspace"
)

// file is one of Bonsai's files in a kind's folder.
type file struct {
	kind   string
	name   string
	path   string // absolute
	target string // project-relative, forward slashes
	mtime  time.Time
	size   int64
	age    time.Time // the kind's own clock, once read
	aged   bool
	reason string // the rule that cleans it, once judged
	ladder *ladderResult
}

// fileKind is one kind of file the cleaner cleans.
type fileKind struct {
	name   string
	folder string
	// match reports one of the names Bonsai writes in the folder.
	match func(name string) bool
	// age reads a file's age by the kind's own clock, its modification time when that does not read.
	age func(r *run, f *file) time.Time
	// protect judges one file (log, ladder): true keeps it. An error keeps it too, and is noted.
	protect func(r *run, f *file) (bool, error)
	// protectAll judges every candidate at once (asks): the set it gives is kept. An error keeps them all.
	protectAll func(r *run, all, cands []*file) (map[*file]bool, error)
}

// files runs one kind: its rule, its folder, its files judged and cleaned.
func (r *run) files(k fileKind) {
	rule, err := RuleOf(r.cfg.Doc, k.name)
	if err != nil {
		r.fail(err)
		return
	}
	if rule.Keeps() {
		return
	}
	list, err := r.list(k)
	if err != nil {
		r.fail(err)
		return
	}
	var beyond map[*file]bool
	if rule.KeepNewest != nil {
		var ok bool
		if beyond, ok = r.rank(k, list, *rule.KeepNewest); !ok {
			return
		}
	}
	var cutoff time.Time
	if rule.KeepDays != nil {
		cutoff = r.now.Add(-time.Duration(*rule.KeepDays) * 24 * time.Hour)
	}
	judge := func(f *file) bool {
		if r.keep[f.target] {
			return false
		}
		if rule.KeepDays != nil && (f.aged || f.mtime.Before(cutoff)) {
			if !f.aged {
				f.age, f.aged = k.age(r, f), true
			}
			if f.age.Before(cutoff) {
				f.reason = reason(k.name, "keep_days", *rule.KeepDays)
				return true
			}
		}
		if beyond[f] {
			f.reason = reason(k.name, "keep_newest", *rule.KeepNewest)
			return true
		}
		return false
	}
	if k.protectAll != nil {
		var cands []*file
		for _, f := range list {
			if r.out() {
				return
			}
			if judge(f) {
				cands = append(cands, f)
			}
		}
		if len(cands) == 0 || r.out() {
			return
		}
		kept, err := k.protectAll(r, list, cands)
		if err != nil {
			r.fail(err)
			return
		}
		// In day order (the names), stopping at the first that does not go: this file's comment.
		sort.Slice(cands, func(i, j int) bool { return cands[i].name < cands[j].name }) // names are distinct
		for _, f := range cands {
			if kept[f] {
				continue
			}
			if r.out() || r.halt || !r.remove(f) {
				return
			}
		}
		return
	}
	for _, f := range list {
		if r.out() || r.halt {
			return
		}
		if !judge(f) {
			continue
		}
		protected, err := k.protect(r, f)
		if err != nil {
			r.fail(err)
		}
		if protected || err != nil {
			continue
		}
		r.remove(f)
	}
}

// rank reads every file's age and marks those beyond the newest n (newest first by age, then by name), the budget
// checked before each read. ok is false when the budget ran out first: the ranking is dropped, and the kind waits
// (this file's comment).
func (r *run) rank(k fileKind, list []*file, n int64) (map[*file]bool, bool) {
	for _, f := range list {
		if r.out() {
			return nil, false
		}
		if !f.aged {
			f.age, f.aged = k.age(r, f), true
		}
	}
	byAge := append([]*file{}, list...)
	sort.Slice(byAge, func(i, j int) bool { // names are distinct
		if !byAge[i].age.Equal(byAge[j].age) {
			return byAge[i].age.After(byAge[j].age)
		}
		return byAge[i].name < byAge[j].name
	})
	beyond := map[*file]bool{}
	for i, f := range byAge {
		if int64(i) >= n {
			beyond[f] = true
		}
	}
	return beyond, true
}

// listBatch is how many of a folder's entries a run reads at a time: the budget is checked between batches, so a huge
// folder's listing stops at the budget too.
const listBatch = 256

// errStopped is the budget run out in the middle of a read. The run stops, with Stopped set; it is not an error.
var errStopped = errors.New("clean: the budget ran out")

// readBatch reads a folder's next n entries at most (a variable for the tests: a folder slowed to a fixed pace,
// standing in for a huge one).
var readBatch = func(fh *os.File, n int) ([]fs.DirEntry, error) { return fh.ReadDir(n) }

// readDir reads a folder listBatch entries at a time, in the folder's own order, giving each entry to each. The budget
// is checked before each batch: errStopped when it ran out first. A folder not there is fs.ErrNotExist.
func (r *run) readDir(dir string, each func(fs.DirEntry)) error {
	fh, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer func() { _ = fh.Close() }()
	for {
		if r.out() {
			return errStopped
		}
		batch, err := readBatch(fh, listBatch)
		for _, de := range batch {
			each(de)
		}
		switch {
		case errors.Is(err, io.EOF), err == nil && len(batch) == 0:
			return nil
		case err != nil:
			return err
		}
	}
}

// list lists a kind's folder in the main checkout's .bonsai/local/: the names Bonsai writes there, regular files only,
// oldest modification time first, read a batch at a time (readDir: errStopped when the budget runs out first). The
// folder, and .bonsai and .bonsai/local above it, must be real folders: a link (or a Windows junction) is never
// followed, and then nothing is listed. A folder not there holds nothing.
func (r *run) list(k fileKind) ([]*file, error) {
	dir := r.localPath(k.folder)
	for _, p := range []string{filepath.Join(r.main, ".bonsai"), filepath.Dir(dir), dir} {
		fi, err := os.Lstat(p)
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		if err != nil {
			return nil, err
		}
		if fi.Mode().Type() != fs.ModeDir {
			return nil, nil // a link, a junction or a file where a folder should be: left alone
		}
	}
	var out []*file
	err := r.readDir(dir, func(de fs.DirEntry) {
		if !de.Type().IsRegular() || !k.match(de.Name()) {
			return
		}
		fi, err := de.Info()
		if err != nil || !fi.Mode().IsRegular() {
			return // gone since the folder was read, or not a regular file after all
		}
		out = append(out, &file{kind: k.name, name: de.Name(), path: filepath.Join(dir, de.Name()),
			target: target(k.folder, de.Name()), mtime: fi.ModTime(), size: fi.Size()})
	})
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	sort.Slice(out, func(i, j int) bool { // names are distinct
		if !out[i].mtime.Equal(out[j].mtime) {
			return out[i].mtime.Before(out[j].mtime)
		}
		return out[i].name < out[j].name
	})
	return out, nil
}

// localDir is the main checkout's .bonsai/local/ (for the kinds' readers).
func (r *run) localDir() string { return filepath.Join(r.main, filepath.FromSlash(workspace.LocalDir)) }
