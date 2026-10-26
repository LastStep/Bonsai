package vergit

// The tree's state as a ladder result's git.dirty reads it (Tree, Dirty): git status, and the checkout's contents read
// against HEAD with nothing git keeps from an earlier look trusted, inside every submodule too.

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/LastStep/Bonsai/internal/workspace"
)

// maxDepth is how deep submodules inside submodules are read; one deeper cannot be proved and counts as dirty.
const maxDepth = 8

// TreeState is a checkout's tree as git.dirty reads it.
type TreeState struct {
	// Dirty: git status shows a change, or a path's contents differ from HEAD's though git status shows nothing, or a
	// submodule's state cannot be proved.
	Dirty bool
	// Shown are the paths git status shows (forward slashes, relative to the checkout's top), sorted.
	Shown []string
	// Hidden are the changed paths git status does not show, each with why, sorted by path: a tracked file whose
	// contents differ from HEAD's (the index's stat times or an index flag kept git from looking, or a filter of the
	// repository's own made it read as unchanged), the same inside a submodule (its path from the top: lib/x.go), a file
	// in a submodule's folder git does not read, and a submodule whose state cannot be read.
	Hidden []Hidden
	// Verified are the tracked files under Git LFS whose bytes are those HEAD's committed pointer names (lfs.go), so
	// they read as unchanged though their bytes differ from HEAD's blob (the pointer), sorted; inside a submodule, its
	// path from the top. A caller running its own git status drops such a file's " M" entry.
	Verified []string
}

// Hidden is one changed path git status does not show.
type Hidden struct {
	Path string // relative to the checkout's top, forward slashes; a folder ends in /
	Why  string // what hid it, in a few ASCII words
	// Filter is the filter HEAD's committed attributes put the file under ("lfs" for Git LFS, "git-crypt"), when its
	// contents read as changed with that filter off and it is not the file Git LFS's pointer names; "" otherwise.
	Filter string
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
//     filter cannot make a changed file read as HEAD's, and with core.hooksPath naming a folder inside the read's own
//     temporary one, so no hook the repository names (post-index-change, run when git writes the fresh index) can mark
//     a file in it as unchanged. A file whose contents with filters off differ from HEAD's counts as changed even where
//     a filter is honest (git-crypt): such a checkout always counts as dirty. The exception is Git LFS (lfs.go): a file
//     HEAD's committed attributes put under it, changed in the working tree only, reads as unchanged when its bytes
//     are those HEAD's committed pointer names (its SHA-256 and size), in this read and in the plain status alike, and
//     is listed in Verified. Then each tracked file the
//     ident or working-tree-encoding attribute applies to, which git reads through a conversion that makes different
//     bytes read as one blob, is read by its bytes against HEAD's blob (convert.go), and counts as changed when they
//     differ: an honest checkout git made under such an attribute (an expanded $Id$, a UTF-16 file) counts as dirty
//     too, as with a filter. Line ends still convert as the user's git converts them (core.autocrlf, core.eol, the
//     text, eol and crlf attributes), which are no program: a CR before an LF reads as unchanged.
//
// Every submodule checked out (a gitlink in HEAD or the index whose folder holds a .git) is read the same way, its
// paths named from the top; a gitlink folder git does not read (no .git) that holds a file counts as dirty, its files
// named, since git status shows nothing there; a submodule that cannot be read, or one past maxDepth, counts as dirty
// and is named. An error means the tree could not be read: the caller counts it dirty.
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
	st, err := j.tree(0)
	if err != nil {
		return TreeState{Dirty: true}, err
	}
	sort.Strings(st.Shown)
	sort.Strings(st.Verified)
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

// tree reads one checkout (the judgment's), and its submodules at depth+1.
func (j *judgment) tree(depth int) (TreeState, error) {
	var st TreeState
	out, err := j.git.run(nil, statusArgs...)
	if err != nil {
		return st, fmt.Errorf("git cannot read the checkout's status: %v", err)
	}
	entries := splitZ(out)
	var modified []string // the paths changed in the working tree only, each a candidate for the Git LFS read
	for _, e := range entries {
		if len(e) < 4 {
			return st, fmt.Errorf("git status gave a line this package does not read: %q", e)
		}
		if e[:2] == " M" {
			modified = append(modified, e[3:])
		}
	}
	changed, verified, err := j.contents(modified)
	if err != nil {
		return st, err
	}
	shown := map[string]bool{}
	for _, e := range entries {
		if e[:2] == " M" && verified[e[3:]] {
			continue // a Git LFS file whose bytes are those HEAD's pointer names
		}
		st.Shown = append(st.Shown, e[3:])
		shown[strings.TrimSuffix(e[3:], "/")] = true
	}
	for _, h := range changed {
		if !shown[h.Path] {
			st.Hidden = append(st.Hidden, h)
		}
	}
	for p := range verified {
		st.Verified = append(st.Verified, p)
	}
	subs, err := j.gitlinks()
	if err != nil {
		return st, err
	}
	for _, s := range subs {
		h, err := j.submodule(s, depth, &st.Verified)
		if err != nil {
			return st, err
		}
		st.Hidden = append(st.Hidden, h...)
	}
	return st, nil
}

// contents gives the tracked paths whose contents differ from HEAD's (Tree's second read), each with why: changed,
// deleted or of another type, by a status against a fresh index read from HEAD, the repository's own filters switched
// off; and a file whose bytes an attribute that converts them on their way into git makes read as HEAD's (converted).
// A path that read changed in the working tree only (" M"), here or in the plain status (modified), is read by Git
// LFS's pointer (lfs.go): verified gives those whose bytes are those HEAD's committed pointer names, which are left
// out of the paths changed; another such path under a filter keeps its place, its reason saying which filter.
func (j *judgment) contents(modified []string) (changed []Hidden, verified map[string]bool, err error) {
	tmp, err := os.MkdirTemp("", "bonsai-index-")
	if err != nil {
		return nil, nil, fmt.Errorf("no temporary folder for a fresh index: %s", osWords(err))
	}
	defer removeTree(tmp)
	off, err := j.filtersOff()
	if err != nil {
		return nil, nil, err
	}
	g := *j.git
	g.env = append(append([]string{}, j.git.env...), "GIT_INDEX_FILE="+filepath.Join(tmp, "index"))
	g.extra = append(off, "core.splitIndex=false", "index.sparse=false",
		"core.hooksPath="+filepath.ToSlash(filepath.Join(tmp, "no-hooks")))
	if j.head != "" {
		if _, err := g.run(nil, "read-tree", j.head); err != nil {
			return nil, nil, fmt.Errorf("git cannot read HEAD into a fresh index: %v", err)
		}
	}
	out, err := g.run(nil, statusArgs...)
	if err != nil {
		return nil, nil, fmt.Errorf("git cannot compare the checkout's contents with HEAD: %v", err)
	}
	var found []Hidden
	seen := map[string]bool{}
	candidates := append([]string{}, modified...)
	for _, e := range splitZ(out) {
		if len(e) < 4 {
			return nil, nil, fmt.Errorf("git status gave a line this package does not read: %q", e)
		}
		if e[:2] == "??" {
			continue // an untracked file shows to plain git status as well (or as staged there, when the index has it)
		}
		found = append(found, Hidden{Path: e[3:], Why: "its contents differ from HEAD's, though git status shows no change"})
		seen[e[3:]] = true
		if e[:2] == " M" {
			candidates = append(candidates, e[3:])
		}
	}
	conv, err := j.converted(&g, seen)
	if err != nil {
		return nil, nil, err
	}
	lfs, err := j.lfs(&g, candidates)
	if err != nil {
		return nil, nil, err
	}
	for _, h := range found {
		if lfs.verified[h.Path] {
			continue
		}
		if why, ok := lfs.why[h.Path]; ok {
			h.Why, h.Filter = why, lfs.filter[h.Path]
		}
		changed = append(changed, h)
	}
	return append(changed, conv...), lfs.verified, nil
}

// removeTree removes a temporary folder this package made and everything in it, each file and then each folder, the
// deepest first, with workspace.RemoveFile, which retries while Windows reports one busy (an antivirus scan or an
// indexer opening a fresh index). Best effort: what still cannot be removed is left in the system's temporary folder,
// where it changes no read.
func removeTree(dir string) {
	var dirs []string
	_ = filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		switch {
		case err != nil:
		case d.IsDir():
			dirs = append(dirs, p)
		default:
			_ = workspace.RemoveFile(p)
		}
		return nil
	})
	for i := len(dirs) - 1; i >= 0; i-- {
		_ = workspace.RemoveFile(dirs[i])
	}
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

// gitlinks gives the submodules' paths: every gitlink (mode 160000) in HEAD's tree or in the index, sorted, each once.
func (j *judgment) gitlinks() ([]string, error) {
	seen := map[string]bool{}
	read := func(args ...string) error {
		out, err := j.git.run(nil, args...)
		if err != nil {
			return fmt.Errorf("git cannot list the submodules: %v", err)
		}
		for _, e := range splitZ(out) {
			meta, p, ok := strings.Cut(e, "\t")
			if ok && strings.HasPrefix(meta, "160000 ") {
				seen[p] = true
			}
		}
		return nil
	}
	entries, err := j.headTree()
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if e.mode == "160000" {
			seen[e.path] = true
		}
	}
	if err := read("ls-files", "-z", "--stage"); err != nil {
		return nil, err
	}
	var out []string
	for p := range seen {
		out = append(out, p)
	}
	sort.Strings(out)
	return out, nil
}

// submodule reads one gitlink's folder (rel, from the checkout's top): a submodule checked out is read as Tree reads a
// checkout, its paths named from the top; a folder git does not read that holds a file names its files; one past
// maxDepth, or one that cannot be read, is named as dirty.
func (j *judgment) submodule(rel string, depth int, verified *[]string) ([]Hidden, error) {
	dir := filepath.Join(j.git.dir, filepath.FromSlash(rel))
	fi, err := os.Stat(dir)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return nil, nil // not there: git status shows a deleted gitlink, or none was checked out
	case err != nil:
		return []Hidden{{Path: rel + "/", Why: "a submodule's folder that cannot be read: " + osWords(err)}}, nil
	case !fi.IsDir():
		return nil, nil // a file where a gitlink was: git status shows it
	}
	if _, err := os.Lstat(filepath.Join(dir, ".git")); err != nil {
		var files []Hidden
		walkErr := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() {
				r, _ := filepath.Rel(j.git.dir, p)
				files = append(files, Hidden{Path: filepath.ToSlash(r), Why: "a file in a submodule's folder that is not checked out, which " +
					"git does not read"})
			}
			return nil
		})
		if walkErr != nil {
			return []Hidden{{Path: rel + "/", Why: "a submodule's folder that cannot be read: " + osWords(walkErr)}}, nil
		}
		return files, nil
	}
	if depth+1 > maxDepth {
		return []Hidden{{Path: rel + "/", Why: fmt.Sprintf("a submodule more than %d deep, which is not read", maxDepth)}}, nil
	}
	sub := &judgment{in: Input{Root: dir, Git: j.in.Git}, found: map[string]*finding{}}
	if err := sub.start(j.git.ctx); err != nil {
		if j.git.ctx.Err() != nil {
			return nil, err
		}
		return []Hidden{{Path: rel + "/", Why: "a submodule whose state cannot be read: " + oneLine(err.Error())}}, nil
	}
	if !sameDir(sub.top, filepath.ToSlash(dir)) {
		return []Hidden{{Path: rel + "/", Why: "a submodule's folder whose .git is not a checkout's own"}}, nil
	}
	st, err := sub.tree(depth + 1)
	if err != nil {
		if j.git.ctx.Err() != nil {
			return nil, err
		}
		return []Hidden{{Path: rel + "/", Why: "a submodule whose state cannot be read: " + oneLine(err.Error())}}, nil
	}
	var out []Hidden
	for _, h := range st.Hidden {
		out = append(out, Hidden{Path: path.Join(rel, h.Path) + trailingSlash(h.Path), Why: h.Why + " (inside the submodule " + rel + ")",
			Filter: h.Filter})
	}
	for _, p := range st.Verified {
		*verified = append(*verified, path.Join(rel, p))
	}
	return out, nil
}

// trailingSlash keeps a folder's slash through path.Join.
func trailingSlash(p string) string {
	if strings.HasSuffix(p, "/") {
		return "/"
	}
	return ""
}

// sameDir compares two folders as git names them (forward slashes), case-blind on Windows.
func sameDir(a, b string) bool {
	a, b = strings.TrimSuffix(a, "/"), strings.TrimSuffix(b, "/")
	return a == b || (os.PathSeparator == '\\' && strings.EqualFold(a, b))
}

// osWords says why a file could not be read by the error's kind, never its text (which names a whole path):
// workspace.OSWords.
func osWords(err error) string { return workspace.OSWords(err) }

// hidingInSubmodules are the findings a submodule's own index and ignore rules can hide a change by: each submodule
// checked out is judged for them as the top checkout is, and what it finds counts in the top's findings.
var hidingInSubmodules = []string{findAssume, findSkip, findHidden, findUntrackedGI}

// submodules judges every submodule checked out (gitlinks) for the findings that hide a change (hidingInSubmodules),
// their names from the top checkout, at any depth up to maxDepth; a submodule that cannot be judged, or one deeper,
// is a note (Tree counts it dirty).
func (j *judgment) submodules() error {
	subs, err := j.gitlinks()
	if err != nil {
		return err
	}
	for _, rel := range subs {
		dir := filepath.Join(j.git.dir, filepath.FromSlash(rel))
		if fi, err := os.Stat(filepath.Join(dir, ".git")); err != nil || fi == nil {
			continue // not checked out: Tree names any file in its folder
		}
		where := j.prefix + rel
		if j.depth+1 > maxDepth {
			j.notes = append(j.notes, fmt.Sprintf("the submodule %s is more than %d deep and not judged (the run counts as dirty)",
				shown(where), maxDepth))
			continue
		}
		sub := &judgment{in: Input{Root: dir, Git: j.in.Git}, found: map[string]*finding{}, prefix: where + "/", depth: j.depth + 1}
		err := sub.start(j.git.ctx)
		if err == nil && !sameDir(sub.top, filepath.ToSlash(dir)) {
			err = errors.New("its .git is not a checkout's own")
		}
		for _, step := range []func() error{sub.index, sub.ignored, sub.submodules} {
			if err != nil {
				break
			}
			err = step()
		}
		if err != nil {
			if j.git.ctx.Err() != nil {
				return errors.New("git did not answer in time")
			}
			j.notes = append(j.notes, "the submodule "+shown(where)+" is not judged ("+oneLine(err.Error())+"; the run counts as dirty)")
			continue
		}
		for _, w := range hidingInSubmodules {
			f, got := j.found[w], sub.found[w]
			if f == nil || got == nil || f.why != "" {
				continue
			}
			f.names = append(f.names, got.names...)
			f.exempt += got.exempt
		}
		j.notes = append(j.notes, sub.notes...)
	}
	return nil
}

// ErrNoCommit and ErrNotInHead are Committed's answers when HEAD holds no such file: HEAD names no commit yet (a new
// repository's branch), or its commit holds no file at that path. ErrNotAFile: HEAD's commit holds a folder or a
// submodule there. ErrHeadUnreadable: HEAD names a commit, but git cannot read it, its tree or the file it lists (an
// object not in the repository or damaged: a clone whose borrowed objects are gone, a fetch cut short, a fault on the
// disk); git fsck names which.
var (
	ErrNoCommit       = errors.New("HEAD names no commit yet")
	ErrNotInHead      = errors.New("HEAD's commit holds no such file")
	ErrNotAFile       = errors.New("HEAD's commit holds a folder or a submodule there, not a file")
	ErrHeadUnreadable = errors.New("git cannot read what HEAD names")
	// ErrCommitUnreadable is ErrHeadUnreadable's case where HEAD's commit itself is not there: errors.Is matches both.
	ErrCommitUnreadable = fmt.Errorf("git cannot read HEAD's commit: it is not in the repository, or is damaged (%w)", ErrHeadUnreadable)
)

// headError is an error of Committed's that wraps one of its answers (errors.Is) and says, in git's words, what failed.
type headError struct {
	is  error
	msg string
}

func (e *headError) Error() string { return e.msg }
func (e *headError) Unwrap() error { return e.is }

// Committed reads the file rel (forward slashes, from the top) as committed at HEAD in the checkout whose top folder
// is root: git's object, never the working file, read with one git call (git cat-file blob HEAD:<rel>) through this
// package's fixed environment, so neither the user's nor the repository's settings can show another file. When that
// call fails, up to three more tell why: ErrNoCommit (HEAD names no commit), ErrNotInHead (its commit holds no such
// path), ErrNotAFile (it holds a folder or a submodule there), or an error wrapping ErrHeadUnreadable (HEAD names a
// commit git cannot read, ErrCommitUnreadable too; or one whose tree or file git cannot read). Any other error says git
// could not run, answer in time, or read HEAD or the repository at all. The ladder runner and the stop gate read
// bonsai.yaml this way (stopgate.CommittedConfig).
func Committed(root, rel string, timeout time.Duration) ([]byte, error) {
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	g, err := newRunner(ctx, "", root)
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return nil, errors.New("git is not on the PATH")
		}
		return nil, fmt.Errorf("git cannot be found: %v", err)
	}
	out, err := g.run(nil, "cat-file", "blob", "HEAD:"+rel)
	if err == nil {
		return out, nil
	}
	late := func() bool { return ctx.Err() != nil }
	if late() {
		return nil, errors.New("git did not answer in time")
	}
	if _, herr := g.run(nil, "rev-parse", "-q", "--verify", "HEAD^{commit}"); herr != nil {
		if late() {
			return nil, errors.New("git did not answer in time")
		}
		_, nerr := g.run(nil, "rev-parse", "-q", "--verify", "HEAD")
		switch {
		case late():
			return nil, errors.New("git did not answer in time")
		case nerr == nil: // HEAD names an object id, which is no commit git can read
			msg := "git cannot read HEAD's commit: it is not in the repository, or is damaged"
			if exitCode(herr) != 1 {
				msg += " (" + herr.Error() + ")"
			}
			return nil, &headError{is: ErrCommitUnreadable, msg: msg}
		case exitCode(herr) == 1 && exitCode(nerr) == 1:
			return nil, ErrNoCommit
		}
		return nil, fmt.Errorf("git cannot read HEAD: %v", herr) // HEAD itself, or the repository, cannot be read
	}
	listed, lerr := g.run(nil, "ls-tree", "-z", "HEAD", "--", rel)
	switch {
	case lerr != nil && late():
		return nil, errors.New("git did not answer in time")
	case lerr != nil:
		return nil, &headError{is: ErrHeadUnreadable, msg: fmt.Sprintf("git cannot read HEAD's tree: %v", lerr)}
	case len(listed) == 0:
		return nil, ErrNotInHead
	}
	meta, _, _ := strings.Cut(strings.TrimSuffix(string(listed), "\x00"), "\t")
	if f := strings.Fields(meta); len(f) != 3 || f[1] != "blob" {
		return nil, &headError{is: ErrNotAFile, msg: fmt.Sprintf("HEAD's %s is a folder or a submodule, not a file", shown(rel))}
	}
	return nil, &headError{is: ErrHeadUnreadable, msg: fmt.Sprintf("HEAD's %s is not a file git can read: %v", shown(rel), err)}
}
