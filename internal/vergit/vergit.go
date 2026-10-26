// Package vergit is the ver-git rung kind, git integrity (spec §9; plan-5 step 5.4, piece 5.4.6): it records the
// task's base commit and finds the git tricks that hide a change, for the verifier.
//
// What it records, in the rung's captures, so the task's next climb can compare: base (the merge base of HEAD with
// the base branch, which the runner works out once per run; on the base branch itself, where that is HEAD, the base
// an earlier climb of the task recorded), stash-count, and exclude-hash (sha256 of the common git folder's
// info/exclude, line ends read as LF).
//
// What it finds, each by its word in format.GitFindings with its count (null when it cannot be judged, the reason
// saying why), and at most 20 names a kind in the reason, paths with forward slashes:
//   - test-file-changed: test files, by the tests forms' conventions (format.TestFileConventions), that differ
//     between the base and the working tree, untracked ones included;
//   - assume-unchanged and skip-worktree: index entries git has been told not to look at (git ls-files -v);
//   - exclude-hidden: files present that only info/exclude or core.excludesFile ignore, by the source git
//     check-ignore -v names; untracked-gitignore: files present that a .gitignore git does not track ignores (one
//     that ignores itself, so neither it nor what it hides shows in git status);
//   - exclude-lines: lines in info/exclude other than comments; exclude-edited: info/exclude changed since the
//     task's last climb;
//   - filter-local: a filter driver defined in the repository's own configuration (a filter can make a changed file
//     read as unchanged);
//   - stash-added: stash entries the task's last climb did not count; base-rewritten: a base the last climb recorded
//     that is no longer an ancestor of HEAD.
//
// Why it informs and never turns red: spec §9 makes the rung "not required at first: it informs the verifier", and
// most of what it finds has an honest reason (an editor's files in a global excludes file, a test changed on
// purpose). Its status is green with its findings listed, or error when git cannot be read at all: an error blocks a
// stop gate whose floor holds the rung, so it never comes from an ordinary repository (a shallow clone, no stash, a
// detached HEAD, a linked worktree, a repository with one commit or none).
//
// Why Hides exists (note 5.4.6 item 3): a test run on a change git does not show would claim a clean commit. An
// assume-unchanged or skip-worktree entry, or a file only info/exclude, core.excludesFile or an untracked .gitignore
// ignores, is exactly such a change, and git status, which the runner's dirty reads, is fooled by it. So Outcome.Hides
// is true for those findings, the runner makes the result's dirty true, and the stop gate, which refuses a dirty
// result, refuses it, while this rung itself stays green and informative. The fix is the repository's own: undo the
// entry, remove the file, or ignore it in a committed .gitignore, which shows in the diff (and which the guard
// protects where a project says so). Chosen over a red rung, which the spec defers, and over leaving dirty to git's
// status alone. One exemption: Claude Code itself ignores its own local files in every project, through the user's
// global excludes (.claude/settings.local.json) and a block in info/exclude (its other .claude/ files), by a list that
// grows with its versions; a file inside a .claude/ folder (at any depth) that only those rules ignore is named under
// exclude-hidden but does not set Hides, so Claude Code's own state never makes a checkout dirty. The other half of
// dirty is the tree's state (tree.go: Tree, Dirty): git status, which settings outside any diff can also fool
// (status.showUntrackedFiles=no or a lying core.fsmonitor in the repository's config; a filter and an attributes file
// in the user's global config), run as this package runs git; and the checkout's contents read against HEAD through a
// fresh index with the repository's own filters off, since git status trusts what its index kept from an earlier look
// (stat times refreshed while a filter made a changed file read as unchanged, then the filter taken out; an index
// flag), and each file under the ident or working-tree-encoding attribute by its bytes, since git reads different
// bytes under those as one blob. Every submodule checked out is read the same way and judged for the findings that hide a change (its own
// assume-unchanged, skip-worktree and ignore rules), its paths named from the top. So the runner's dirty is Dirty(in)
// or Outcome.Hides.
//
// How git runs (git.go): one program, the judgment's deadline, a fixed environment (no user configuration, no GIT_*
// variable from the caller, no pager, no prompt, no network) and settings pinned on the command line, so nothing the
// user or the repository set can point git elsewhere or make it stop looking. Read once with the user's own
// environment: which global excludes file their git uses, since that file is what hides (it can only add a file to
// look through), and their line-end settings, so a touched CRLF checkout reads as git reads it for them.
//
// Known limits, each for the verifier to know:
//   - A test file edited under an assume-unchanged or skip-worktree entry is listed under that entry, not under
//     test-file-changed: git's diff trusts the entry. It still makes the run dirty.
//   - An untracked folder holding both files git shows and files only info/exclude ignores is listed as git shows it
//     (the folder, untracked), not file by file; git status shows that folder, so the run is dirty anyway.
//   - stash-added compares counts: a dropped entry and an added one between two climbs cancel out. The stash is the
//     main checkout's, shared by its worktrees.
//   - A filter defined in the repository's own configuration is named under filter-local, and switched off for the
//     contents' read, so a file it changes reads as it is: a file whose contents differ from its blob without the
//     filter counts as dirty. A file HEAD's committed attributes put under Git LFS is the exception (lfs.go): it reads
//     as unchanged when its bytes hash to the SHA-256 HEAD's committed pointer names, at its size, with HEAD's
//     executable bit where core.fileMode is pinned (or when it holds the pointer itself), trusting neither git-lfs nor
//     the index; which files are tried comes from HEAD's committed attributes only, and none while info/attributes
//     holds a rule. A pointer with an extension line, or in a form other than the spec's version 1, is not read. Any
//     other filter that changes a file honestly (git-crypt, in the repository's configuration or the user's, which
//     these reads do not load) still makes every climb of such a checkout dirty, its reason naming the filter: a proof
//     there needs its filtered files out of the tested tree, or a person's decision to drop the filter.
//   - The contents' read hashes every tracked file at every climb's start and end (about 60 ms on Bonsai's own
//     repository on Linux, against 6 ms for git status), and rung 0 reads them once more; each Git LFS file is hashed
//     once more by its pointer at each of those reads.
//   - A Git LFS file inside a submodule whose index keeps stale stat times (touched since the submodule's index was
//     written, Git LFS not in the submodule's own configuration) makes the top's git status show the submodule, which
//     counts as dirty: the top's status line for a submodule is not read file by file.
//   - Inside a submodule, the hiding findings and the contents are judged; test-file-changed, info/exclude's lines
//     and edits, filter-local, the stash and the base are the top checkout's only. A submodule more than 8 deep is
//     not read, and named. Its own filters are switched off as the top's are.
//   - The repository's own configuration is read by every git call (no variable switches it off); the pinned
//     settings cover the parts of it known to hide a change from these reads (git.go's pins). Three of them describe
//     the file system and git keeps them from when it made the repository: core.ignorecase (true hides an untracked
//     file whose name differs from a tracked one's only in case), core.fileMode (false hides an executable bit) and
//     core.symlinks (false hides a tracked link replaced by a file holding the link's text). Each is pinned where the
//     file system under the repository's common git folder can hold what it hides (probe.go, probed once per run by
//     a temporary file and link there, as git init probes), and left as git has it where it cannot: Windows (NTFS is
//     case-blind, keeps no executable bit, and Git for Windows checks links out as files) and macOS's case-blind
//     volumes, where pinning would read honest files as changed. A working tree on another file system than its git
//     folder (or a Windows folder made case-sensitive under a case-blind git folder) is read by the git folder's
//     answer; macOS's core.precomposeUnicode (a name in two Unicode forms) is not pinned.
//   - Conversions on a file's way into git make different bytes read as one blob. Filters are switched off for the
//     contents' read; the ident and working-tree-encoding attributes are read by the file's bytes against HEAD's blob
//     (convert.go), so an honest checkout under them always counts as dirty, as one under git-crypt does. Line ends are
//     the one conversion left on (the text, eol and crlf attributes, core.autocrlf, core.eol), as the user's git
//     reads them: a CR added before an LF reads as unchanged, which a test reading a file's bytes could see.
//   - Hooks: the one call that writes an index (the contents' fresh index, in a temporary file) runs with
//     core.hooksPath naming a folder that does not exist, so no hook the repository names runs during the read; no
//     other call writes the index (GIT_OPTIONAL_LOCKS=0) or runs a hook. A process the agent left running beside the
//     climb can still change files between the reads.
//   - A file inside a .claude/ folder that only info/exclude or core.excludesFile ignores is named, not made dirty
//     (Claude Code's own local files; no project's test reads them). So a test that reads a file an agent hid inside
//     .claude/ through info/exclude passes here with the file named and the run clean; CI, which lacks the file,
//     still fails it. The exemption is the folder's name, never Claude Code's marker line or its list of files, which
//     an agent could write or Claude Code could change.
//   - In a sparse checkout every file outside the cone is a skip-worktree entry, so such a checkout always counts as
//     dirty: a proof is of the whole tree.
//   - A checkout git refuses as another user's (safe.directory) is an error: the user's global configuration, where
//     that setting lives, is not read.
//   - It judges what git can see; a determined agent able to rewrite .git itself can do more, which the guard's
//     never-granted .git path is for.
package vergit

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/schema"
)

// Input is what one ver-git rung judges. The runner fills it once per run (5.4.2 note 13); tests fill it from their
// own fixture repositories.
type Input struct {
	// Root is the top folder of the checkout the ladder runs in (the runner's, from git rev-parse --show-toplevel):
	// the whole checkout is judged, and git never looks above it for another.
	Root string
	// BaseBranch is the base branch, "" when there is none (a main checkout on a detached HEAD). Only the reason
	// names it.
	BaseBranch string
	// MergeBase is the merge base of HEAD with the base branch, a full commit id, "" when there is none.
	MergeBase string
	// Tests are the tests forms (format.TestForms) whose test-file conventions say which files are test files: the
	// rung's own tests when it names one, else the forms the ladder's other rungs name, in rung order. None, or only
	// forms with no patterns (junit), and test-file-changed is not judged.
	Tests []string
	// Earlier is this rung's captures in the task's earlier result, nil when there is none.
	Earlier schema.Object
	// Git is the git program, "" for git on this process's PATH.
	Git string
	// Timeout is the whole judgment's limit; 0 means DefaultTimeout.
	Timeout time.Duration
}

// Outcome is a rung's outcome, the shape rung 0 (internal/rung0) returns too, with Hides added.
type Outcome struct {
	Status   string        // a format.RungStatuses word: green, or error when git cannot be read
	Reason   string        // one line of ASCII: what was found, what was not judged and why, the base
	Captures schema.Object // format.GitFindings' words, in that order
	Hides    bool          // a finding hides a change from git: the runner makes the result's dirty true
}

// The captures' keys this package reads back, each a word of format.GitFindings.
const (
	keyBase        = "base"
	keyStashCount  = "stash-count"
	keyExcludeHash = "exclude-hash"
)

// The findings, each a word of format.GitFindings.
const (
	findTestFile    = "test-file-changed"
	findAssume      = "assume-unchanged"
	findSkip        = "skip-worktree"
	findExcludeEdit = "exclude-edited"
	findStash       = "stash-added"
	findRewritten   = "base-rewritten"
	findExcludeLine = "exclude-lines"
	findHidden      = "exclude-hidden"
	findUntrackedGI = "untracked-gitignore"
	findFilter      = "filter-local"
)

// hiding are the findings that hide a change from git status: any of them makes Outcome.Hides true.
var hiding = map[string]string{
	findAssume:      "undo each: git update-index --no-assume-unchanged -- <path>",
	findSkip:        "undo each: git update-index --no-skip-worktree -- <path> (or git sparse-checkout disable)",
	findHidden:      "remove each, or ignore it in the repository's .gitignore",
	findUntrackedGI: "remove the untracked .gitignore, or commit it",
}

// maxNames is how many names of one kind the reason gives.
const maxNames = 20

var commitID = regexp.MustCompile(`^([0-9a-f]{40}|[0-9a-f]{64})$`)

// finding is one kind found: its names (their number is its count), and why it was not judged ("" when it was).
type finding struct {
	names  []string
	why    string
	exempt int // names that do not make the run dirty (a hiding finding's files inside a .claude/ folder)
}

// judgment is one run's state.
type judgment struct {
	in       Input
	git      *gitRunner
	top      string // the checkout's top folder, as git gives it
	common   string // the common git folder, absolute
	shallow  bool
	head     string // HEAD's commit, "" when HEAD has none yet
	tracked  map[string]bool
	found    map[string]*finding
	base     string // the base recorded now
	baseHow  string
	notes    []string
	stashes  int
	exclHash string
	prefix   string // a submodule's path from the top checkout, with its slash ("lib/"); "" for the top checkout
	depth    int    // how many submodules deep this checkout lies

	headEntries []treeEntry // HEAD's tree, every entry (headTree), once read
	headRead    bool
}

// name is a path of this checkout as a finding names it: from the top checkout (a submodule's prefix first), as shown.
func (j *judgment) name(p string) string { return shown(j.prefix + p) }

// Judge judges one checkout's git integrity.
func Judge(in Input) Outcome {
	timeout := in.Timeout
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	j := &judgment{in: in, found: map[string]*finding{}}
	if err := j.judge(ctx); err != nil {
		return j.failed(err)
	}
	return j.outcome()
}

// judge runs every check; an error means git could not be read.
func (j *judgment) judge(ctx context.Context) error {
	if err := j.start(ctx); err != nil {
		return err
	}
	for _, step := range []func() error{j.index, j.ignored, j.submodules, j.exclude, j.filters, j.stash, j.setBase, j.testFiles} {
		if err := step(); err != nil {
			return err
		}
	}
	return nil
}

// start finds git and the checkout, and reads the user's settings every call keeps.
func (j *judgment) start(ctx context.Context) error {
	root, err := filepath.Abs(j.in.Root)
	if err == nil {
		root, err = filepath.EvalSymlinks(root)
	}
	if err != nil {
		return fmt.Errorf("the checkout's folder cannot be read: %v", oneLine(err.Error()))
	}
	g, err := newRunner(ctx, j.in.Git, root)
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return errors.New("git is not on the PATH")
		}
		return fmt.Errorf("git cannot be found: %v", err)
	}
	j.git = g
	if err := j.locate(); err != nil {
		return err
	}
	g.fs = fsPins(filepath.FromSlash(j.common))
	note := g.userSettings()
	if g.ctx.Err() != nil {
		return errors.New("git did not answer in time")
	}
	if note != "" {
		j.notes = append(j.notes, note)
	}
	return nil
}

// locate finds the checkout's top folder, its common git folder, whether it is shallow, and HEAD.
func (j *judgment) locate() error {
	out, err := j.git.run(nil, "rev-parse", "--show-toplevel")
	if err != nil {
		var e *gitError
		if errors.As(err, &e) && e.code == -1 && !errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("git could not be run: %v", err)
		}
		return fmt.Errorf("git cannot read the checkout: %v", err)
	}
	j.top = filepath.ToSlash(strings.TrimRight(string(out), "\r\n"))
	if j.top == "" {
		return errors.New("git rev-parse --show-toplevel gave no folder")
	}
	j.git.dir = filepath.FromSlash(j.top)
	out, err = j.git.run(nil, "rev-parse", "--git-common-dir", "--is-shallow-repository")
	if err != nil {
		return fmt.Errorf("git cannot read the checkout: %v", err)
	}
	lines := strings.Split(strings.TrimRight(strings.ReplaceAll(string(out), "\r\n", "\n"), "\n"), "\n")
	if len(lines) != 2 || lines[0] == "" {
		return fmt.Errorf("git rev-parse gave an answer this rung does not read: %q", firstLine(string(out)))
	}
	common := filepath.FromSlash(lines[0])
	if !filepath.IsAbs(common) {
		common = filepath.Join(j.git.dir, common)
	}
	j.common, j.shallow = filepath.ToSlash(common), lines[1] == "true"
	head, err := j.git.run(nil, "rev-parse", "-q", "--verify", "HEAD^{commit}")
	switch {
	case err == nil:
		j.head = strings.TrimSpace(string(head))
	case exitCode(err) == 1: // HEAD names no commit yet: a new repository
	default:
		return fmt.Errorf("git cannot read HEAD: %v", err)
	}
	return nil
}

// index reads the index: its paths (for the .gitignore files git tracks) and its assume-unchanged and skip-worktree
// entries (git ls-files -v: a lower-case tag is assume-unchanged, S or s skip-worktree).
func (j *judgment) index() error {
	out, err := j.git.run(nil, "ls-files", "-z", "-v")
	if err != nil {
		return fmt.Errorf("git cannot read the index: %v", err)
	}
	j.tracked = map[string]bool{}
	assume, skip := j.newFinding(findAssume), j.newFinding(findSkip)
	seen := map[string]bool{}
	for _, e := range splitZ(out) {
		if len(e) < 3 || e[1] != ' ' {
			return fmt.Errorf("git ls-files -v gave an entry this rung does not read: %q", e)
		}
		tag, p := e[0], e[2:]
		j.tracked[p] = true
		if seen[p] {
			continue // an unmerged path, once per stage
		}
		seen[p] = true
		if tag >= 'a' && tag <= 'z' {
			assume.names = append(assume.names, j.name(p))
		}
		if tag == 'S' || tag == 's' {
			skip.names = append(skip.names, j.name(p))
		}
	}
	return nil
}

// ignored finds the files present that git ignores, and which of them only a rule outside the repository's committed
// .gitignore files ignores: info/exclude or the global excludes file (exclude-hidden), or a .gitignore git does not
// track (untracked-gitignore). A folder git ignores whole counts once, and only when a file is in it: an empty folder
// hides nothing.
//
// git ls-files --ignored --directory lists a folder whose files are all ignored by finer rules (.claude/ when only
// .claude/settings.local.json is ignored) and then what is inside it; with --no-empty-directory it would list the
// folder alone, which no rule matches, and the files inside would go unseen. So empty folders are listed and judged
// here, and a folder no rule matches whose insides git did not list is listed again, file by file.
func (j *judgment) ignored() error {
	hidden, untracked := j.newFinding(findHidden), j.newFinding(findUntrackedGI)
	out, err := j.git.run(nil, ignoredListing...)
	if err != nil {
		return fmt.Errorf("git cannot list the ignored files: %v", err)
	}
	paths := splitZ(out)
	if len(paths) == 0 {
		return nil
	}
	matched, unmatched, err := j.whyIgnored(paths)
	if err != nil {
		return err
	}
	for _, d := range unmatched {
		if !strings.HasSuffix(d, "/") || listsInside(paths, d) {
			continue
		}
		out, err := j.git.run(nil, "ls-files", "-z", "--others", "--ignored", "--exclude-standard", "--", d)
		if err != nil {
			return fmt.Errorf("git cannot list the ignored files in %s: %v", shown(d), err)
		}
		if inside := splitZ(out); len(inside) > 0 {
			more, _, err := j.whyIgnored(inside)
			if err != nil {
				return err
			}
			matched = append(matched, more...)
		}
	}
	var dirs []string // folders git ignores whole: what it lists inside one is not counted again
	for _, e := range matched {
		if strings.HasSuffix(e.path, "/") {
			dirs = append(dirs, e.path)
		}
	}
next:
	for _, e := range matched {
		for _, d := range dirs {
			if e.path != d && strings.HasPrefix(e.path, d) {
				continue next
			}
		}
		if strings.HasSuffix(e.path, "/") && !holdsFile(filepath.Join(j.git.dir, filepath.FromSlash(e.path))) {
			continue
		}
		switch {
		case isPerDirectory(e.src):
			if !j.tracked[e.src] {
				untracked.names = append(untracked.names, j.name(e.path)+" (by "+j.name(e.src)+")")
			}
		default:
			by := "info/exclude"
			if j.git.excludes != "" && filepath.ToSlash(e.src) == j.git.excludes {
				by = "core.excludesFile"
			}
			if insideClaude(e.path) {
				by += "; inside .claude/, not counted as dirty"
				hidden.exempt++
			}
			hidden.names = append(hidden.names, j.name(e.path)+" (by "+by+")")
		}
	}
	return nil
}

// insideClaude reports whether a path lies inside a folder named .claude, at any depth: Claude Code's own local files,
// which Claude Code itself ignores through info/exclude and the user's global excludes in every project, by a list
// that grows with its versions. Such a file found under exclude-hidden is named but does not make the run dirty: no
// project's test reads Claude Code's local state. Keyed on the folder, never on Claude Code's marker line in
// info/exclude (an agent can write it) or on a list of its file names (which would chase Claude Code). A folder
// listed whole (.claude/) counts as inside; a file named .claude does not.
func insideClaude(p string) bool {
	segs := strings.Split(strings.TrimSuffix(p, "/"), "/")
	if !strings.HasSuffix(p, "/") {
		segs = segs[:len(segs)-1] // a file: only the folders holding it count
	}
	for _, s := range segs {
		if s == ".claude" {
			return true
		}
	}
	return false
}

// ignoredListing lists the ignored files and folders present (a test adds --no-empty-directory, to prove the
// listing again file by file).
var ignoredListing = []string{"ls-files", "-z", "--others", "--ignored", "--directory", "--exclude-standard"}

// ignoredEntry is a path git ignores and the source of the rule that decides it.
type ignoredEntry struct{ src, path string }

// whyIgnored asks git check-ignore which rule decides each path: the paths a rule ignores (a negation decides that a
// path is not ignored), and the rest.
func (j *judgment) whyIgnored(paths []string) (matched []ignoredEntry, unmatched []string, err error) {
	out, err := j.git.run([]byte(strings.Join(paths, "\x00")+"\x00"), "check-ignore", "-z", "--verbose",
		"--non-matching", "--no-index", "--stdin")
	if err != nil && exitCode(err) != 1 {
		return nil, nil, fmt.Errorf("git cannot say why files are ignored: %v", err)
	}
	f := splitZ(out)
	if len(f)%4 != 0 {
		return nil, nil, fmt.Errorf("git check-ignore gave an answer this rung does not read (%d fields)", len(f))
	}
	for i := 0; i < len(f); i += 4 {
		src, pattern, p := f[i], f[i+2], f[i+3]
		if src == "" || strings.HasPrefix(pattern, "!") {
			unmatched = append(unmatched, p)
			continue
		}
		matched = append(matched, ignoredEntry{src, p})
	}
	return matched, unmatched, nil
}

// listsInside reports whether a listing holds a path inside the folder d.
func listsInside(paths []string, d string) bool {
	for _, p := range paths {
		if p != d && strings.HasPrefix(p, d) {
			return true
		}
	}
	return false
}

// holdsFile reports whether a folder holds a file at any depth (anything but a folder counts). A folder that cannot be
// read counts as holding one: what cannot be looked at is not taken as empty.
func holdsFile(dir string) bool {
	found := errors.New("found")
	err := filepath.WalkDir(dir, func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			return found
		}
		if !d.IsDir() {
			return found
		}
		return nil
	})
	return err != nil
}

// isPerDirectory reports whether check-ignore's source is a .gitignore file in the working tree (a path relative to
// the top folder), not info/exclude or the global excludes file.
func isPerDirectory(src string) bool {
	s := filepath.ToSlash(src)
	return path.Base(s) == ".gitignore" && !path.IsAbs(s) && !filepath.IsAbs(src) && !strings.HasPrefix(s, ".git/")
}

// exclude reads the common git folder's info/exclude: its hash (LF line ends; an absent file reads as empty), its
// lines other than comments, and whether it changed since the task's last climb.
func (j *judgment) exclude() error {
	raw, err := os.ReadFile(filepath.Join(filepath.FromSlash(j.common), "info", "exclude"))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("info/exclude cannot be read: %v", err)
	}
	text := strings.ReplaceAll(string(raw), "\r\n", "\n")
	sum := sha256.Sum256([]byte(text))
	j.exclHash = "sha256:" + hex.EncodeToString(sum[:])
	lines := j.newFinding(findExcludeLine)
	for _, l := range strings.Split(text, "\n") {
		if strings.TrimSpace(l) == "" || strings.HasPrefix(l, "#") {
			continue
		}
		lines.names = append(lines.names, shown(l))
	}
	edited := j.newFinding(findExcludeEdit)
	was, ok := j.in.Earlier.Get(keyExcludeHash)
	if s, isStr := was.(string); !ok || !isStr || s == "" {
		edited.notJudged("no earlier climb of the task recorded it")
	} else if s != j.exclHash {
		edited.names = append(edited.names, "info/exclude")
	}
	return nil
}

// filters finds filter drivers the repository's own configuration defines (the fixed environment reads no other).
func (j *judgment) filters() error {
	f := j.newFinding(findFilter)
	out, err := j.git.run(nil, "config", "-z", "--includes", "--get-regexp", `^filter\..+\.(clean|smudge|process)$`)
	if err != nil {
		if exitCode(err) == 1 {
			return nil // none
		}
		return fmt.Errorf("git cannot read the repository's configuration: %v", err)
	}
	seen := map[string]bool{}
	for _, e := range splitZ(out) {
		key, _, _ := strings.Cut(e, "\n")
		name := strings.TrimPrefix(key[:strings.LastIndexByte(key, '.')], "filter.")
		if !seen[name] {
			seen[name] = true
			f.names = append(f.names, shown(name))
		}
	}
	return nil
}

// stash counts the stash's entries and compares the count with the task's last climb's.
func (j *judgment) stash() error {
	f := j.newFinding(findStash)
	if _, err := j.git.run(nil, "rev-parse", "-q", "--verify", "refs/stash"); err == nil {
		out, err := j.git.run(nil, "rev-list", "--walk-reflogs", "--count", "refs/stash", "--")
		if err != nil {
			return fmt.Errorf("git cannot count the stash: %v", err)
		}
		n, err := strconv.Atoi(strings.TrimSpace(string(out)))
		if err != nil || n < 0 {
			return fmt.Errorf("git counted the stash as %q", firstLine(string(out)))
		}
		j.stashes = n
	} else if exitCode(err) != 1 {
		return fmt.Errorf("git cannot read the stash: %v", err)
	}
	was, ok := number(j.in.Earlier, keyStashCount)
	if !ok {
		f.notJudged("no earlier climb of the task counted it")
		return nil
	}
	for i := 0; i < j.stashes-was; i++ {
		f.names = append(f.names, fmt.Sprintf("stash@{%d}", i))
	}
	return nil
}

// setBase settles the base this climb records, and whether the base the task's last climb recorded is still in
// HEAD's history.
func (j *judgment) setBase() error {
	merge := strings.ToLower(j.in.MergeBase)
	if merge != "" && !commitID.MatchString(merge) {
		j.notes = append(j.notes, fmt.Sprintf("the runner's merge base %s is not a commit id, so it is not used", shown(j.in.MergeBase)))
		merge = ""
	}
	earlier := RecordedBase(j.in.Earlier)
	branch := "the base branch"
	if j.in.BaseBranch != "" {
		branch = shown(j.in.BaseBranch)
	}
	switch {
	case merge != "" && merge != j.head:
		j.base, j.baseHow = merge, "the merge base with "+branch
	case earlier != "":
		j.base, j.baseHow = earlier, "kept from the task's earlier climb"
		if merge == "" {
			j.baseHow += " (no merge base now)"
		} else {
			j.baseHow += " (HEAD is on " + branch + ")"
		}
	case merge != "":
		j.base, j.baseHow = merge, "HEAD itself, on "+branch
	default:
		j.baseHow = "none: the runner found no base branch or merge base"
	}
	rw := j.newFinding(findRewritten)
	switch {
	case earlier == "":
		rw.notJudged("no earlier climb of the task recorded a base")
	case j.head == "":
		rw.names = append(rw.names, earlier+" (HEAD has no commit)")
	default:
		present, err := j.present(earlier)
		if err != nil {
			return err
		}
		if !present {
			if j.shallow {
				rw.notJudged("the recorded base is not in this shallow clone")
			} else {
				rw.names = append(rw.names, earlier+" (gone from the repository)")
			}
			return nil
		}
		_, err = j.git.run(nil, "merge-base", "--is-ancestor", earlier, j.head)
		switch {
		case err == nil:
		case exitCode(err) == 1 && j.shallow:
			rw.notJudged("this shallow clone does not hold HEAD's whole history")
		case exitCode(err) == 1:
			rw.names = append(rw.names, earlier)
		default:
			return fmt.Errorf("git cannot read HEAD's history: %v", err)
		}
	}
	return nil
}

// present reports whether the repository holds a commit.
func (j *judgment) present(id string) (bool, error) {
	_, err := j.git.run(nil, "cat-file", "-e", id+"^{commit}")
	switch {
	case err == nil:
		return true, nil
	case exitCode(err) == 1 || exitCode(err) == 128:
		return false, nil
	default:
		return false, fmt.Errorf("git cannot read the object database: %v", err)
	}
}

// testFiles lists the test files that differ between the base and the working tree: changed, added or deleted, by
// git's diff, and untracked ones git status would show.
func (j *judgment) testFiles() error {
	f := j.newFinding(findTestFile)
	var forms []string
	for _, form := range j.in.Tests {
		if c, ok := format.TestFileConventionOf(form); ok && len(c.Patterns) > 0 {
			forms = append(forms, form)
		}
	}
	switch {
	case len(forms) == 0:
		f.notJudged("no tests form with a test-file convention names the test files")
		return nil
	case j.base == "":
		f.notJudged("no base")
		return nil
	}
	present, err := j.present(j.base)
	if err != nil {
		return err
	}
	if !present {
		f.notJudged("the base is not in this clone")
		return nil
	}
	out, err := j.git.run(nil, "diff", "--name-status", "--no-renames", "--no-ext-diff", "--no-textconv",
		"--ignore-submodules=all", "--no-relative", "-z", j.base, "--")
	if err != nil {
		return fmt.Errorf("git cannot compare the working tree with the base: %v", err)
	}
	words := map[byte]string{'A': "added", 'D': "deleted", 'M': "changed", 'T': "changed"}
	d := splitZ(out)
	if len(d)%2 != 0 {
		return fmt.Errorf("git diff gave an answer this rung does not read (%d fields)", len(d))
	}
	for i := 0; i < len(d); i += 2 {
		if TestFile(forms, d[i+1]) {
			how := words[d[i][0]]
			if how == "" {
				how = "changed"
			}
			f.names = append(f.names, shown(d[i+1])+" ("+how+")")
		}
	}
	out, err = j.git.run(nil, "ls-files", "-z", "--others", "--exclude-standard")
	if err != nil {
		return fmt.Errorf("git cannot list the untracked files: %v", err)
	}
	for _, p := range splitZ(out) {
		if TestFile(forms, p) {
			f.names = append(f.names, shown(p)+" (added, untracked)")
		}
	}
	return nil
}

// TestFile reports whether a path (relative to the checkout's top, forward slashes) is a test file by any of the
// tests forms' conventions (format.TestFileConventions): segments match as path.Match does, ** stands for any number
// of whole segments, none included.
func TestFile(forms []string, rel string) bool {
	segs := strings.Split(rel, "/")
	for _, form := range forms {
		c, ok := format.TestFileConventionOf(form)
		if !ok {
			continue
		}
		for _, p := range c.Patterns {
			if matchSegs(strings.Split(p, "/"), segs) {
				return true
			}
		}
	}
	return false
}

func matchSegs(g, r []string) bool {
	if len(g) == 0 {
		return len(r) == 0
	}
	if g[0] == "**" {
		for i := 0; i <= len(r); i++ {
			if matchSegs(g[1:], r[i:]) {
				return true
			}
		}
		return false
	}
	if len(r) == 0 {
		return false
	}
	if ok, err := path.Match(g[0], r[0]); err != nil || !ok {
		return false
	}
	return matchSegs(g[1:], r[1:])
}

// RecordedBase is the base a ver-git rung's captures recorded (a result's, read back), "" when they hold none:
// what new tests must fail on (5.4.5) reads on the base branch itself.
func RecordedBase(captures schema.Object) string {
	s, _ := captureValue(captures, keyBase).(string)
	if s = strings.ToLower(s); commitID.MatchString(s) {
		return s
	}
	return ""
}

// captureValue is a capture's value, nil when absent.
func captureValue(captures schema.Object, key string) any {
	v, _ := captures.Get(key)
	return v
}

// number reads a capture as a count: a whole number at least 0.
func number(captures schema.Object, key string) (int, bool) {
	var s string
	switch v := captureValue(captures, key).(type) {
	case json.Number:
		s = v.String()
	case int:
		return v, v >= 0
	case int64:
		return int(v), v >= 0
	case float64:
		if v >= 0 && v == float64(int(v)) {
			return int(v), true
		}
		return 0, false
	default:
		return 0, false
	}
	n, err := strconv.Atoi(s)
	return n, err == nil && n >= 0
}

// newFinding starts a finding, judged with nothing found until told otherwise.
func (j *judgment) newFinding(word string) *finding {
	f := &finding{}
	j.found[word] = f
	return f
}

func (f *finding) notJudged(why string) { f.why = why }

// outcome is the judgment as a rung's outcome.
func (j *judgment) outcome() Outcome {
	o := Outcome{Status: "green"}
	var found, notJudged []string
	kinds, hides := 0, 0
	for _, w := range format.GitFindings {
		switch w.Word {
		case keyBase:
			if j.base == "" {
				o.Captures = append(o.Captures, schema.Member{Key: w.Word, Value: nil})
			} else {
				o.Captures = append(o.Captures, schema.Member{Key: w.Word, Value: j.base})
			}
			continue
		case keyStashCount:
			o.Captures = append(o.Captures, schema.Member{Key: w.Word, Value: int64(j.stashes)})
			continue
		case keyExcludeHash:
			o.Captures = append(o.Captures, schema.Member{Key: w.Word, Value: j.exclHash})
			continue
		}
		f := j.found[w.Word]
		if f == nil || f.why != "" {
			o.Captures = append(o.Captures, schema.Member{Key: w.Word, Value: nil})
			if f != nil {
				notJudged = append(notJudged, w.Word+" ("+f.why+")")
			}
			continue
		}
		sort.Strings(f.names)
		o.Captures = append(o.Captures, schema.Member{Key: w.Word, Value: int64(len(f.names))})
		if len(f.names) == 0 {
			continue
		}
		kinds++
		s := fmt.Sprintf("%s %d: %s", w.Word, len(f.names), names(f.names))
		if fix, ok := hiding[w.Word]; ok {
			if len(f.names) > f.exempt {
				hides++
				o.Hides = true
				s += " (hides a change from git; " + fix + ")"
			} else {
				s += " (all inside a .claude/ folder: Claude Code's local files, named, not counted as dirty)"
			}
		}
		found = append(found, s)
	}
	var b strings.Builder
	switch {
	case kinds == 0:
		b.WriteString("nothing found")
	case hides > 0:
		fmt.Fprintf(&b, "%s found, %d hiding a change from git status, so the run counts as dirty", plural(kinds, "kind"), hides)
	default:
		fmt.Fprintf(&b, "%s found, none making the run dirty", plural(kinds, "kind"))
	}
	for _, s := range found {
		b.WriteString("; " + s)
	}
	if len(notJudged) > 0 {
		b.WriteString(". Not judged: " + strings.Join(notJudged, "; "))
	}
	if j.base != "" {
		fmt.Fprintf(&b, ". Base %s, %s", j.base[:12], j.baseHow)
	} else {
		b.WriteString(". Base " + j.baseHow)
	}
	b.WriteString("; " + plural(j.stashes, "stash entry"))
	if j.head == "" {
		b.WriteString("; HEAD has no commit yet")
	}
	for _, n := range j.notes {
		b.WriteString("; " + n)
	}
	o.Reason = asciiOnly(b.String())
	return o
}

// plural is a count and its noun.
func plural(n int, noun string) string {
	switch {
	case n == 1:
		return "1 " + noun
	case strings.HasSuffix(noun, "y"):
		return fmt.Sprintf("%d %sies", n, noun[:len(noun)-1])
	}
	return fmt.Sprintf("%d %ss", n, noun)
}

// names lists at most maxNames names (each already as shown), then how many more.
func names(all []string) string {
	var out []string
	for i, n := range all {
		if i == maxNames {
			out = append(out, fmt.Sprintf("and %d more", len(all)-maxNames))
			break
		}
		out = append(out, n)
	}
	return strings.Join(out, ", ")
}

// failed is the outcome when git cannot be read: error, with the earlier climb's records carried, so the task's next
// climb still compares with the last ones read.
func (j *judgment) failed(err error) Outcome {
	o := Outcome{Status: "error", Reason: "git cannot be read, so git integrity is not judged: " + asciiOnly(err.Error())}
	for _, w := range format.GitFindings {
		var v any
		switch w.Word {
		case keyBase:
			if b := RecordedBase(j.in.Earlier); b != "" {
				v = b
			}
		case keyStashCount:
			if n, ok := number(j.in.Earlier, keyStashCount); ok {
				v = int64(n)
			}
		case keyExcludeHash:
			if s, ok := captureValue(j.in.Earlier, keyExcludeHash).(string); ok && strings.HasPrefix(s, "sha256:") {
				v = s
			}
		}
		o.Captures = append(o.Captures, schema.Member{Key: w.Word, Value: v})
	}
	if RecordedBase(j.in.Earlier) != "" || captureValue(j.in.Earlier, keyStashCount) != nil {
		o.Reason += " (the records are the task's earlier climb's, carried)"
	}
	return o
}
