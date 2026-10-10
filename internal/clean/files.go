package clean

// A file kind's run (log, asks, ladder): its folder listed, each file's age read, and each judged and deleted in
// order.
//
// The order: by modification time, oldest first, then by name, so what a budget leaves for the next run is the
// newest. The modification time only orders the files and spares reads: a file modified within keep_days is not
// read for keep_days, since Bonsai writes each line a moment after its record's at, so such a file's own clock is
// within keep_days too (or, for a file touched since, it is kept until its modification time ages as well: the safe
// side). Every file that is read is judged by its own clock (the kind's age). keep_newest ranks every file by age, so
// with it every file is read first; if the budget runs out before that ranking is whole, the run stops and
// keep_newest waits for a run that finishes it.
//
// Log and ladder files are judged one at a time, each deleted as soon as it is judged, so a run the budget cuts short
// has still cleaned what it judged. Asks are judged all at once (asks.go), since whether one day file may go depends on
// which others stay.

import (
	"errors"
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
		if len(cands) == 0 {
			return
		}
		kept, err := k.protectAll(r, list, cands)
		if err != nil {
			r.fail(err)
			return
		}
		for _, f := range cands {
			if r.out() || r.halt {
				return
			}
			if !kept[f] {
				r.remove(f)
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

// rank reads every file's age and marks those beyond the newest n (newest first by age, then by name). ok is false
// when the budget ran out first.
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

// list lists a kind's folder in the main checkout's .bonsai/local/: the names Bonsai writes there, regular files only,
// oldest modification time first. The folder, and .bonsai and .bonsai/local above it, must be real folders: a link
// (or a Windows junction) is never followed, and then nothing is listed. A folder not there holds nothing.
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
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var out []*file
	for _, de := range entries {
		if !de.Type().IsRegular() || !k.match(de.Name()) {
			continue
		}
		fi, err := de.Info()
		if err != nil || !fi.Mode().IsRegular() {
			continue // gone since the folder was read, or not a regular file after all
		}
		out = append(out, &file{kind: k.name, name: de.Name(), path: filepath.Join(dir, de.Name()),
			target: target(k.folder, de.Name()), mtime: fi.ModTime(), size: fi.Size()})
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
