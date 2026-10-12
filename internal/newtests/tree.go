package newtests

// The base worktree: made in the system's temporary folder, the branch's test files copied in, and always removed;
// a worktree a killed check left is swept by the next check in the same repository.

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/LastStep/Bonsai/internal/vergit"
)

// The temporary folder's shape: <system temporary folder>/bonsai-newtests-<random>/, holding the owner file and the
// worktree.
const (
	tempPrefix = "bonsai-newtests-"
	ownerName  = "owner" // "pid <n>": the process whose check made it
	treeName   = "base"  // the worktree
	hooksName  = "hooks" // never made: core.hooksPath points here, so no hook of the repository runs
)

// removeWait is how long the folder's removal keeps trying while Windows reports a file busy (an antivirus scan, an
// indexer, a process the end could not reach).
var removeWait = 5 * time.Second

// tree is one base worktree.
type tree struct {
	parent string // the temporary folder holding it (absolute, system separators)
	dir    string // the worktree
	added  bool   // git worktree add ran (its record in the repository may exist)
}

// makeTree makes the base worktree: the temporary folder and its owner file, then git worktree add --detach at the
// base with the repository's hooks off. A tree is returned even with an error once its folder exists, so the caller
// removes what was made.
func makeTree(g *vergit.Runner, root, base string) (*tree, error) {
	parent, err := os.MkdirTemp("", tempPrefix)
	if err != nil {
		return nil, fmt.Errorf("no temporary folder: %v", err)
	}
	t := &tree{parent: parent, dir: filepath.Join(parent, treeName)}
	if long, err := filepath.EvalSymlinks(parent); err == nil { // Windows' temporary folder may be an 8.3 short name
		t.parent, t.dir = long, filepath.Join(long, treeName)
	}
	owner := "pid " + strconv.Itoa(os.Getpid()) + "\nchecks " + filepath.ToSlash(root) + "\n"
	if err := os.WriteFile(filepath.Join(t.parent, ownerName), []byte(owner), 0o644); err != nil {
		return t, fmt.Errorf("the owner file cannot be written: %v", err)
	}
	t.added = true
	_, err = g.Run("-c", "core.hooksPath="+filepath.ToSlash(filepath.Join(t.parent, hooksName)), "-c", "advice.detachedHead=false",
		"worktree", "add", "--detach", "--quiet", t.dir, base)
	return t, err
}

// copyIn copies the branch's test files into the worktree: each file from the climbed working tree (its bytes and,
// where the system keeps one, its mode), and each file the branch deleted removed.
func (t *tree) copyIn(root string, files []change) error {
	for _, f := range files {
		dst := filepath.Join(t.dir, filepath.FromSlash(f.path))
		if f.deleted {
			if err := os.Remove(dst); err != nil && !errors.Is(err, fs.ErrNotExist) {
				return fmt.Errorf("%s: %v", f.path, err)
			}
			continue
		}
		src := filepath.Join(root, filepath.FromSlash(f.path))
		info, err := os.Stat(src)
		if errors.Is(err, fs.ErrNotExist) {
			if err := os.Remove(dst); err != nil && !errors.Is(err, fs.ErrNotExist) {
				return fmt.Errorf("%s: %v", f.path, err)
			}
			continue
		}
		if err != nil {
			return fmt.Errorf("%s: %v", f.path, err)
		}
		if info.IsDir() { // a submodule's folder: not a file to copy
			continue
		}
		data, err := os.ReadFile(src)
		if err != nil {
			return fmt.Errorf("%s: %v", f.path, err)
		}
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return fmt.Errorf("%s: %v", f.path, err)
		}
		if err := os.Remove(dst); err != nil && !errors.Is(err, fs.ErrNotExist) { // a link at the base is replaced, not followed
			return fmt.Errorf("%s: %v", f.path, err)
		}
		if err := os.WriteFile(dst, data, info.Mode().Perm()|0o200); err != nil {
			return fmt.Errorf("%s: %v", f.path, err)
		}
	}
	return nil
}

// remove removes the worktree and its folder, with a context of its own: git worktree remove --force, the folder
// removed (Windows' busy retry), and git worktree prune when git's removal failed, so git's record of it goes too.
// It gives a note for what it could not do.
func (t *tree) remove(gitProg, root string) []string {
	var notes []string
	var gitErr error
	ctx, cancel := context.WithTimeout(context.Background(), gitWait)
	defer cancel()
	g, _, err := vergit.NewRunner(ctx, gitProg, root)
	if t.added {
		if err == nil {
			_, gitErr = g.Run("worktree", "remove", "--force", "--force", t.dir)
		} else {
			gitErr = err
		}
	}
	if err := removeAll(t.parent); err != nil {
		notes = append(notes, "the base worktree's folder "+filepath.ToSlash(t.parent)+" could not be removed: "+oneLine(err.Error())+
			"; the next check in this repository removes it")
	}
	if t.added && gitErr != nil && g != nil {
		if _, err := g.Run("worktree", "prune"); err != nil {
			notes = append(notes, "git worktree prune failed: "+oneLine(err.Error()))
		}
	}
	return notes
}

// sweep removes the base worktrees of this repository that a check left when its process was killed: listed by git,
// in a bonsai-newtests- folder directly in the system's temporary folder, whose owner file names a process that no
// longer runs. A worktree whose folder is gone leaves git's record, which git worktree prune takes. It changes no
// count; what it could not do is a note.
func (c *check) sweep() {
	out, err := c.git.Run("worktree", "list", "--porcelain")
	if err != nil {
		c.note("the repository's worktrees could not be listed, so none a killed check left was removed: " + oneLine(err.Error()))
		return
	}
	tmp := os.TempDir()
	if long, err := filepath.EvalSymlinks(tmp); err == nil {
		tmp = long
	}
	prune := false
	for _, line := range strings.Split(strings.ReplaceAll(string(out), "\r\n", "\n"), "\n") {
		p, ok := strings.CutPrefix(line, "worktree ")
		if !ok {
			continue
		}
		dir := filepath.FromSlash(p)
		parent := filepath.Dir(dir)
		if filepath.Base(dir) != treeName || !strings.HasPrefix(filepath.Base(parent), tempPrefix) ||
			!samePath(filepath.Dir(parent), tmp) {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(parent, ownerName))
		switch {
		case errors.Is(err, fs.ErrNotExist):
			if _, err := os.Stat(parent); errors.Is(err, fs.ErrNotExist) {
				prune = true // the folder is gone: only git's record is left
			}
			continue
		case err != nil:
			continue
		}
		pid := ownerPID(string(raw))
		if pid <= 0 || pid == os.Getpid() || alive(pid) {
			continue
		}
		old := &tree{parent: parent, dir: dir, added: true}
		for _, n := range old.remove(c.in.Git, c.in.Root) {
			c.note(n)
		}
		c.note("removed the base worktree " + filepath.ToSlash(dir) + " that a killed check (pid " + strconv.Itoa(pid) + ") left")
	}
	if prune {
		if _, err := c.git.Run("worktree", "prune"); err != nil {
			c.note("git worktree prune failed: " + oneLine(err.Error()))
		}
	}
}

// ownerPID reads an owner file's pid, 0 when it names none.
func ownerPID(text string) int {
	for _, line := range strings.Split(text, "\n") {
		if v, ok := strings.CutPrefix(strings.TrimSpace(line), "pid "); ok {
			n, err := strconv.Atoi(strings.TrimSpace(v))
			if err == nil {
				return n
			}
		}
	}
	return 0
}

// samePath compares two folders as the file system does: ignoring case on Windows.
func samePath(a, b string) bool {
	a, b = filepath.Clean(a), filepath.Clean(b)
	if runtime.GOOS == "windows" {
		return strings.EqualFold(a, b)
	}
	return a == b
}

// removeAll removes a folder and everything in it, trying again while removeWait lasts (a file busy on Windows; a
// read-only folder, or a read-only file on Windows, made writable first). A folder already gone is no error.
func removeAll(dir string) error {
	deadline := time.Now().Add(removeWait)
	delay := 10 * time.Millisecond
	for {
		err := os.RemoveAll(dir)
		if err == nil {
			return nil
		}
		if time.Now().After(deadline) {
			return err
		}
		_ = filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
			switch {
			case err != nil:
			case d.IsDir(): // a read-only folder (a module cache a base_setup filled) keeps its files on Linux
				_ = os.Chmod(p, 0o755)
			case runtime.GOOS == "windows": // a read-only file cannot be removed on Windows
				_ = os.Chmod(p, 0o666)
			}
			return nil
		})
		time.Sleep(delay)
		if delay < 320*time.Millisecond {
			delay *= 2
		}
	}
}

// cleanRel is a project-relative path with forward slashes, "" when it is absolute or leaves the project.
func cleanRel(p string) string {
	p = strings.ReplaceAll(p, `\`, "/")
	if p == "" || strings.HasPrefix(p, "/") || (len(p) > 1 && p[1] == ':') {
		return ""
	}
	p = path.Clean(p)
	if p == "." || p == ".." || strings.HasPrefix(p, "../") {
		return ""
	}
	return p
}
