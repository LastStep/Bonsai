package vergit

// The tree's state as a ladder result's git.dirty reads it (Tree, Dirty): git status, and the checkout's contents read
// against HEAD with nothing git keeps from an earlier look trusted.

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// TreeState is a checkout's tree as git.dirty reads it.
type TreeState struct {
	// Dirty: git status shows a change, or a path's contents differ from HEAD's though git status shows nothing.
	Dirty bool
	// Shown are the paths git status shows (forward slashes, relative to the checkout's top), sorted.
	Shown []string
	// Hidden are the changed paths git status does not show, each with why, sorted by path: a tracked file whose
	// contents differ from HEAD's (the index's stat times or an index flag kept git from looking, or a filter of the
	// repository's own made it read as unchanged).
	Hidden []Hidden
}

// Hidden is one changed path git status does not show.
type Hidden struct {
	Path string // relative to the checkout's top, forward slashes; a folder ends in /
	Why  string // what hid it, in a few ASCII words
}

// Tree reads the state of the checkout at in.Root (only in.Root, in.Git and in.Timeout are read).
//
// Two reads, both as this package runs git (its fixed environment and pinned settings, with the user's excludes file
// and line-end settings):
//   - git status, as git shows the tree to anyone who asks (with every submodule's changes);
//   - the contents: the same status run against a fresh index read from HEAD into a temporary file (GIT_INDEX_FILE;
//     git read-tree HEAD), whose entries carry no stat times and no index flags, so git reads and hashes every tracked
//     file and compares its contents with HEAD's; with every filter driver the repository's own configuration
//     defines switched off for that read (filter.<name>.clean, smudge and process emptied, required false), so a
//     filter cannot make a changed file read as HEAD's. A file whose contents with filters off differ from HEAD's counts
//     as changed even where a filter is honest (Git LFS, git-crypt): such a checkout always counts as dirty. Line ends
//     still convert as the user's git converts them (core.autocrlf, core.eol, the text and eol attributes), which are
//     no program.
//
// An error means the tree could not be read: the caller counts it dirty.
func Tree(in Input) (TreeState, error) {
	timeout := in.Timeout
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	j := &judgment{in: in, found: map[string]*finding{}}
	if err := j.start(ctx); err != nil {
		return TreeState{Dirty: true}, err
	}
	st, err := j.tree()
	if err != nil {
		return TreeState{Dirty: true}, err
	}
	sort.Strings(st.Shown)
	sort.Slice(st.Hidden, func(a, b int) bool {
		if st.Hidden[a].Path != st.Hidden[b].Path {
			return st.Hidden[a].Path < st.Hidden[b].Path
		}
		return st.Hidden[a].Why < st.Hidden[b].Why
	})
	st.Dirty = len(st.Shown) > 0 || len(st.Hidden) > 0
	return st, nil
}

// Dirty reports whether the tree is dirty as Tree reads it: the git-status half of a ladder result's git.dirty, the
// other half being Outcome.Hides (dirty = Dirty(in) or Hides). Run plainly, git status can be talked out of a change
// by settings outside any diff (status.showUntrackedFiles=no or a lying core.fsmonitor in the repository's config; a
// filter driver and an attributes file in the user's global configuration), which these reads do not read or pin
// off, and by what the index keeps from an earlier look (its stat times, refreshed while a filter made a changed file
// read as unchanged; its assume-unchanged and skip-worktree flags), which the read of the contents does not trust. On
// an error, the caller counts the run dirty: a tree that cannot be read proves no clean commit.
func Dirty(in Input) (bool, error) {
	st, err := Tree(in)
	return st.Dirty || err != nil, err
}

// statusArgs are git status's arguments for both reads.
var statusArgs = []string{"status", "--porcelain=v1", "-z", "--untracked-files=normal", "--ignore-submodules=none", "--no-renames"}

// tree reads one checkout (the judgment's).
func (j *judgment) tree() (TreeState, error) {
	var st TreeState
	out, err := j.git.run(nil, statusArgs...)
	if err != nil {
		return st, fmt.Errorf("git cannot read the checkout's status: %v", err)
	}
	shown := map[string]bool{}
	for _, e := range splitZ(out) {
		if len(e) < 4 {
			return st, fmt.Errorf("git status gave a line this package does not read: %q", e)
		}
		st.Shown = append(st.Shown, e[3:])
		shown[strings.TrimSuffix(e[3:], "/")] = true
	}
	changed, err := j.contents()
	if err != nil {
		return st, err
	}
	for _, p := range changed {
		if !shown[p] {
			st.Hidden = append(st.Hidden, Hidden{Path: p, Why: "its contents differ from HEAD's, though git status shows no change"})
		}
	}
	return st, nil
}

// contents gives the tracked paths whose contents differ from HEAD's (Tree's second read): changed, deleted or of
// another type, by a status against a fresh index read from HEAD, the repository's own filters switched off.
func (j *judgment) contents() ([]string, error) {
	tmp, err := os.MkdirTemp("", "bonsai-index-")
	if err != nil {
		return nil, fmt.Errorf("no temporary folder for a fresh index: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmp) }()
	off, err := j.filtersOff()
	if err != nil {
		return nil, err
	}
	g := *j.git
	g.env = append(append([]string{}, j.git.env...), "GIT_INDEX_FILE="+filepath.Join(tmp, "index"))
	g.extra = append(off, "core.splitIndex=false", "index.sparse=false")
	if j.head != "" {
		if _, err := g.run(nil, "read-tree", j.head); err != nil {
			return nil, fmt.Errorf("git cannot read HEAD into a fresh index: %v", err)
		}
	}
	out, err := g.run(nil, statusArgs...)
	if err != nil {
		return nil, fmt.Errorf("git cannot compare the checkout's contents with HEAD: %v", err)
	}
	var changed []string
	for _, e := range splitZ(out) {
		if len(e) < 4 {
			return nil, fmt.Errorf("git status gave a line this package does not read: %q", e)
		}
		if e[:2] == "??" {
			continue // an untracked file shows to plain git status as well (or as staged there, when the index has it)
		}
		changed = append(changed, e[3:])
	}
	return changed, nil
}

// filterKey is a filter driver's setting in git's configuration.
var filterKey = regexp.MustCompile(`^filter\.(.+)\.(clean|smudge|process|required)$`)

// filtersOff gives the settings that switch off every filter driver the repository's own configuration defines (the
// fixed environment reads no other configuration): its clean, smudge and process emptied, and required false, so a
// file under it reads as it is. A driver whose name cannot be given on git's command line is an error: what a filter
// does cannot then be kept out of the read.
func (j *judgment) filtersOff() ([]string, error) {
	out, err := j.git.run(nil, "config", "-z", "--includes", "--name-only", "--get-regexp", `^filter\.`)
	if err != nil {
		if exitCode(err) == 1 {
			return nil, nil // none
		}
		return nil, fmt.Errorf("git cannot read the repository's configuration: %v", err)
	}
	seen := map[string]bool{}
	var names []string
	for _, key := range splitZ(out) {
		m := filterKey.FindStringSubmatch(strings.ToLower(key))
		if m == nil {
			continue
		}
		name := key[len("filter.") : len(key)-len(m[2])-1]
		if strings.ContainsAny(name, "=\n\x00") {
			return nil, fmt.Errorf("the repository defines a filter whose name (%s) cannot be switched off on git's command line, so its "+
				"files cannot be read as they are", shown(name))
		}
		if !seen[name] {
			seen[name] = true
			names = append(names, name)
		}
	}
	sort.Strings(names)
	var off []string
	for _, n := range names {
		off = append(off, "filter."+n+".clean=", "filter."+n+".smudge=", "filter."+n+".process=", "filter."+n+".required=false")
	}
	return off, nil
}
