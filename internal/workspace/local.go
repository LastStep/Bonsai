package workspace

// A project's .bonsai/local/ (contract §3, spec §6): the folder that is never committed, its .gitignore, which every
// writer restores when it is missing, and the checkout whose local/ a writer uses, found by reading the checkout's
// .git with no git process.
//
// Contract §3: worktrees share the main checkout's .bonsai/local/; a task worktree has none of its own. Spec §6 and
// contract §3 name git rev-parse --git-common-dir for finding it; ReadGitLayout reads the same files git reads,
// without starting git, so a hook stays fast (design/plan-5.md, 5.2.2 note 3; a test holds its answer equal to
// git's, Find's, on a main checkout, a worktree, a submodule and a separate git folder):
//
//   - .git is a folder: the checkout is its own main, and .git is its git folder.
//   - .git is a file, "gitdir: <path>" (relative to the checkout, or absolute): <path> is the checkout's git folder.
//     Its commondir file names the common git folder (relative to the git folder, or absolute). When the git folder
//     is <common>/worktrees/<name> and the common folder is named .git, the checkout is a worktree and its main
//     checkout is the common folder's parent.
//   - Anything else is its own main: a git folder with no commondir (a submodule, a checkout with a separate git
//     folder), a common folder not named .git (a bare repository's worktree, or a separate git folder's), no .git,
//     or a file that cannot be read. Paths are compared case-blind on Windows.
//
// FindLocal uses that main only when its bonsai.yaml holds the same workspace id as the writer's; else the writer's
// own checkout. It is a tripwire, as the log is (spec §7): an agent can rewrite a worktree's .git file, so at worst a
// record lands in another checkout holding the same id, never in a folder holding none. This finder is the
// recorder's, the asks' and the cleaner's; the guard keeps its own folder until step 5.3 decides what the hook path
// may trust (design/plan-5.md, 5.2.2 note 3).

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// LocalDir is the project's folder that is never committed (contract §3), relative to the main checkout.
const LocalDir = ".bonsai/local"

// GitignoreFile and GitignoreText are .bonsai/.gitignore, written by Bonsai (spec §6, contract §3): bonsai init and
// update write it, bonsai check finds it missing or changed, and every writer of .bonsai/local/ restores it when it
// is missing (EnsureGitignore). This is the text's one home.
const (
	GitignoreFile = ".bonsai/.gitignore"
	GitignoreText = "# .bonsai/.gitignore: written by Bonsai (spec section 6). .bonsai/local/ holds this checkout's log, questions\n" +
		"# and ladder results, never committed. Bonsai writes this file again when it is missing or changed.\n" +
		"local/\n"
)

// EnsureGitignore writes .bonsai/.gitignore in the checkout at root when it is missing (spec §6: "every writer of
// local/ restores a missing .bonsai/.gitignore"), whole or not at all, and reports whether it wrote. A file that is
// there is left as it is, changed or not: bonsai check finds a changed one, and bonsai update writes Bonsai's text
// again. Two writers that both find it missing both write the same bytes; the second rename replaces the first.
func EnsureGitignore(root string) (bool, error) {
	path := filepath.Join(root, filepath.FromSlash(GitignoreFile))
	_, err := os.Lstat(path)
	switch {
	case err == nil:
		return false, nil
	case !errors.Is(err, fs.ErrNotExist):
		return false, err
	}
	if err := WriteFileAtomic(path, []byte(GitignoreText)); err != nil {
		return false, err
	}
	return true, nil
}

// GitLayout is a checkout's git layout, as the files in its .git say (ReadGitLayout). Every path is absolute and
// clean, in the system's own separators: use filepath.ToSlash to store or print one.
type GitLayout struct {
	Root   string // the checkout's top folder, as given
	GitDir string // the checkout's own git folder: Root/.git, or the folder its .git file names; "" when none reads
	Common string // the common git folder its commondir names; GitDir when it names none
	Main   string // the main checkout by the layout alone: Common's parent for a worktree; else Root
}

// ReadGitLayout reads the git layout of the checkout whose top folder is root, with no git process (the rules are
// this file's comment). It never fails: what cannot be read leaves the checkout its own main.
func ReadGitLayout(root string) GitLayout {
	abs, err := filepath.Abs(root)
	if err != nil {
		abs = filepath.Clean(root)
	}
	g := GitLayout{Root: abs, Main: abs}
	dotgit := filepath.Join(abs, ".git")
	fi, err := os.Stat(dotgit)
	if err != nil {
		return g
	}
	if fi.IsDir() {
		g.GitDir, g.Common = dotgit, dotgit
		return g
	}
	if !fi.Mode().IsRegular() {
		return g
	}
	line, ok := firstLine(dotgit)
	rest, isGitdir := strings.CutPrefix(line, "gitdir:")
	gitdir := strings.TrimSpace(rest)
	if !ok || !isGitdir || gitdir == "" {
		return g
	}
	gitdir = resolve(abs, gitdir)
	if fi, err := os.Stat(gitdir); err != nil || !fi.IsDir() {
		return g
	}
	g.GitDir, g.Common = gitdir, gitdir
	common, ok := firstLine(filepath.Join(gitdir, "commondir"))
	common = strings.TrimSpace(common)
	if !ok || common == "" {
		return g // a submodule, or a checkout with a separate git folder: its own main
	}
	g.Common = resolve(gitdir, common)
	fold := runtime.GOOS == "windows"
	if sameName(filepath.Base(g.Common), ".git", fold) && childOf(gitdir, filepath.Join(g.Common, "worktrees"), fold) {
		g.Main = filepath.Dir(g.Common)
	}
	return g
}

// Branch is the checkout's branch, read from its git folder's HEAD: "" for a detached HEAD, no git folder, or a
// HEAD that cannot be read.
func (g GitLayout) Branch() string {
	if g.GitDir == "" {
		return ""
	}
	line, ok := firstLine(filepath.Join(g.GitDir, "HEAD"))
	branch, isRef := strings.CutPrefix(strings.TrimSpace(line), "ref: refs/heads/")
	if !ok || !isRef {
		return ""
	}
	return branch
}

// Local is where a writer in one checkout writes .bonsai/local/ (FindLocal).
type Local struct {
	Root   string // the checkout the writer runs in: its top folder, absolute
	Main   string // the checkout whose .bonsai/local/ is written: the main checkout, or Root
	Branch string // Root's branch, read from its HEAD; "" when detached or unreadable
}

// FindLocal finds where a writer of workspace id, running in the checkout whose top folder is root, writes: the main
// checkout its .git names (ReadGitLayout) when that checkout's bonsai.yaml holds the same id, else root itself. It
// starts no process and never fails.
func FindLocal(root, id string) Local {
	g := ReadGitLayout(root)
	l := Local{Root: g.Root, Main: g.Root, Branch: g.Branch()}
	if id != "" && !samePath(g.Main, g.Root, runtime.GOOS == "windows") && configID(g.Main) == id {
		l.Main = g.Main
	}
	return l
}

// Dir is the folder the writer writes: Main's .bonsai/local/.
func (l Local) Dir() string { return filepath.Join(l.Main, filepath.FromSlash(LocalDir)) }

// configID reads only the id of root/bonsai.yaml, "" when the file cannot be read or holds none. It does not hold
// the rest of the file to its fields, so a main checkout whose bonsai.yaml has a problem elsewhere still counts.
func configID(root string) string {
	raw, err := os.ReadFile(filepath.Join(root, ConfigFile))
	if err != nil {
		return ""
	}
	m, err := readYAML(ConfigFile, raw, WorkspaceFormat, configNext)
	if err != nil {
		return ""
	}
	id, _ := m.Get("id")
	s, _ := id.(string)
	return s
}

// firstLine reads a small file's first line, without its line ending. A file over 4 KiB is not one of git's
// pointer files and is not read.
func firstLine(path string) (string, bool) {
	f, err := os.Open(path)
	if err != nil {
		return "", false
	}
	defer func() { _ = f.Close() }()
	raw, err := io.ReadAll(io.LimitReader(f, 4097))
	if err != nil || len(raw) > 4096 {
		return "", false
	}
	s := string(raw)
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = s[:i]
	}
	return strings.TrimSuffix(s, "\r"), true
}

// resolve makes p, as a git file names it, absolute: relative to base when it is not absolute already.
func resolve(base, p string) string {
	p = filepath.FromSlash(p)
	if !filepath.IsAbs(p) {
		p = filepath.Join(base, p)
	}
	return filepath.Clean(p)
}

// sameName compares two names, case-blind when fold is set (Windows).
func sameName(a, b string, fold bool) bool {
	if fold {
		return strings.EqualFold(a, b)
	}
	return a == b
}

// samePath compares two clean paths, case-blind when fold is set (Windows).
func samePath(a, b string, fold bool) bool {
	return sameName(filepath.Clean(a), filepath.Clean(b), fold)
}

// childOf reports whether path is exactly one folder below parent (parent/<name>), case-blind when fold is set.
func childOf(path, parent string, fold bool) bool {
	path, parent = filepath.Clean(path), filepath.Clean(parent)
	return !samePath(path, parent, fold) && samePath(filepath.Dir(path), parent, fold)
}
