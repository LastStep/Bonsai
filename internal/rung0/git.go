package rung0

// What a checkout changed, through git (the package comment, note 1): the only process rung 0 starts.

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/LastStep/Bonsai/internal/guard"
	"github.com/LastStep/Bonsai/internal/vergit"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// waitDelay is how long a git call's pipes may stay open after it ends or is ended at the deadline (a helper it
// started holding them), before the call gives up on them.
const waitDelay = 2 * time.Second

// git runs git in one checkout, within the judgment's deadline.
type git struct {
	root     string
	ctx      context.Context // the deadline; nil for none
	settings []string        // more -c settings for every call (key=value)
}

func gitIn(root string) git { return git{root: root} }

// deadline is the judgment's deadline, when it has one.
func (g git) deadline() (time.Time, bool) {
	if g.ctx == nil {
		return time.Time{}, false
	}
	return g.ctx.Deadline()
}

// within gives the same runner with a deadline: a git call still running when ctx ends is ended.
func (g git) within(ctx context.Context) git {
	g.ctx = ctx
	return g
}

// run runs git with args in the checkout, git's own environment variables that would point it elsewhere taken out,
// and gives its standard output.
func (g git) run(args ...string) ([]byte, error) {
	ctx := g.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	full := []string{"-C", g.root, "-c", "core.quotepath=off", "--literal-pathspecs"}
	for _, s := range g.settings {
		full = append(full, "-c", s)
	}
	cmd := exec.CommandContext(ctx, "git", append(full, args...)...)
	cmd.Env = gitEnv()
	cmd.WaitDelay = waitDelay
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("git %s: ended at the rung's limit", args[0])
		}
		msg := strings.TrimSpace(stderr.String())
		if i := strings.IndexAny(msg, "\r\n"); i >= 0 {
			msg = msg[:i]
		}
		if msg == "" {
			msg = err.Error()
		}
		return nil, fmt.Errorf("git %s: %s", args[0], msg)
	}
	return stdout.Bytes(), nil
}

// gitEnv is the process's environment without the variables that point git at another repository, index or work
// tree, and with git's prompt off (a hook or a script may have set them; engine's own list).
func gitEnv() []string {
	drop := map[string]bool{"GIT_DIR": true, "GIT_WORK_TREE": true, "GIT_INDEX_FILE": true, "GIT_OBJECT_DIRECTORY": true,
		"GIT_ALTERNATE_OBJECT_DIRECTORIES": true, "GIT_COMMON_DIR": true, "GIT_NAMESPACE": true, "GIT_TERMINAL_PROMPT": true}
	var env []string
	for _, kv := range os.Environ() {
		k, _, _ := strings.Cut(kv, "=")
		if runtime.GOOS == "windows" {
			k = strings.ToUpper(k)
		}
		if !drop[k] {
			env = append(env, kv)
		}
	}
	return append(env, "GIT_TERMINAL_PROMPT=0")
}

// top is the checkout's top folder, as git names it (the folder git's paths are relative to), in the system's
// separators.
func (g git) top() (string, error) {
	out, err := g.run("rev-parse", "--show-toplevel")
	if err != nil {
		return "", err
	}
	t := strings.TrimSpace(string(out))
	if t == "" {
		return "", fmt.Errorf("git rev-parse --show-toplevel names no folder (a bare repository?)")
	}
	return filepath.FromSlash(t), nil
}

// branch is the checkout's branch, "" for a detached HEAD.
func (g git) branch() string {
	out, err := g.run("symbolic-ref", "-q", "--short", "HEAD")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// hasHead reports whether HEAD names a commit (not a branch with no commit yet).
func (g git) hasHead() bool {
	_, err := g.run("rev-parse", "-q", "--verify", "HEAD^{commit}")
	return err == nil
}

// entry is one changed path.
type entry struct{ path string }

// changed lists every path the checkout changed (note 1), sorted, each once: the branch's commits since from when
// from is not "", what is staged, unstaged or untracked and not ignored, and what the checkout's contents read
// against HEAD show changed though git status does not (vergit.Tree's hidden paths: a file whose stat times the index
// kept from a look while a filter made it read as unchanged, one under an index flag, one whose bytes the ident or
// working-tree-encoding attribute reads as HEAD's, one inside a submodule, a file in a submodule's folder git does not
// read), so a change hidden from git status is judged by the guard's function, not only made dirty. git status runs
// with what the file system can hold pinned (vergit.FileSystemSettings), so a stale core.ignorecase, core.fileMode or
// core.symlinks in the repository's configuration hides no untracked file or change from it. An untracked nested
// repository (git names its folder with a trailing slash and does not look inside) is its .git, a checkout planted
// below the top (the guard's floor never grants one), unless it is a worktree of the main checkout main (Claude Code
// places its own worktrees inside the main checkout): another checkout of the project, whose changes are its own
// climb's.
func (g git) changed(from, main string) ([]entry, error) {
	var paths []string
	if from != "" {
		out, err := g.run("diff", "--name-only", "--no-renames", "--no-ext-diff", "-z", from, "HEAD", "--")
		if err != nil {
			return nil, err
		}
		paths = append(paths, splitZ(out)...)
	}
	pinned := g
	pinned.settings = vergit.FileSystemSettings(g.root)
	out, err := pinned.run("status", "--porcelain=v1", "-z", "--no-renames", "--untracked-files=all")
	if err != nil {
		return nil, err
	}
	for _, e := range splitZ(out) {
		if len(e) < 4 || e[2] != ' ' {
			return nil, fmt.Errorf("git status gave a line rung 0 does not read: %q", e)
		}
		p := e[3:]
		if e[:2] == "??" && strings.HasSuffix(p, "/") {
			if ownWorktree(filepath.Join(g.root, filepath.FromSlash(p)), main) {
				continue
			}
			p += ".git"
		}
		paths = append(paths, p)
	}
	in := vergit.Input{Root: g.root}
	if dl, ok := g.deadline(); ok {
		if in.Timeout = time.Until(dl); in.Timeout <= 0 {
			return nil, errors.New("git: ended at the rung's limit")
		}
	}
	tree, err := vergit.Tree(in)
	if err != nil {
		return nil, fmt.Errorf("the checkout's contents cannot be read against HEAD: %v", err)
	}
	for _, h := range tree.Hidden {
		paths = append(paths, strings.TrimSuffix(h.Path, "/"))
	}
	var out2 []entry
	for _, p := range sortedUnique(paths) {
		out2 = append(out2, entry{path: p})
	}
	return out2, nil
}

// ownWorktree reports whether the folder dir is a worktree of the main checkout main, by git's own files and its
// link back (workspace.ReadGitLayout, GitLayout.LinksBack: the link back lies in the main checkout's .git, which the
// floor never grants).
func ownWorktree(dir, main string) bool {
	l := workspace.ReadGitLayout(dir)
	return l.Worktree() && samePlace(l.Main, main) && l.LinksBack()
}

// splitZ splits git's NUL-ended list.
func splitZ(b []byte) []string {
	var out []string
	for _, s := range bytes.Split(b, []byte{0}) {
		if len(s) > 0 {
			out = append(out, string(s))
		}
	}
	return out
}

// change is one changed path as the guard's function takes it: what the checkout holds now, and what it held at
// from (the base: the merge base on a branch, else HEAD; "" when there is none), read only when asked.
func (g git) change(e entry, from string) guard.Change {
	after, known, deleted := readNow(g.root, e.path)
	return guard.Change{Path: e.path, Deleted: deleted, After: after, AfterKnown: known,
		Before: func() ([]byte, error) { return g.blob(from, e.path) }}
}

// blob reads a path's bytes at a commit: an error wrapping fs.ErrNotExist when the commit does not hold it.
func (g git) blob(commit, p string) ([]byte, error) {
	if commit == "" {
		return nil, fs.ErrNotExist
	}
	out, err := g.run("ls-tree", "-z", commit, "--", p)
	if err != nil {
		return nil, err
	}
	fields := splitZ(out)
	if len(fields) == 0 {
		return nil, fmt.Errorf("%s at %s: %w", p, commit, fs.ErrNotExist)
	}
	// "<mode> <type> <object>\t<path>"
	meta, _, _ := strings.Cut(fields[0], "\t")
	parts := strings.Fields(meta)
	if len(parts) != 3 || parts[1] != "blob" {
		return nil, errors.New(p + " at " + commit + " is not a file")
	}
	return g.run("cat-file", "blob", parts[2])
}
