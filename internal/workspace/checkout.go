package workspace

// Finding the checkout a command runs in and its main checkout, through git (contract §3: "a worktree finds its
// main checkout through git (git rev-parse --git-common-dir), as the hooks do today").
//
// Not for the guard: git takes its answer from files an agent may edit (a worktree's .git file, a commondir file),
// so spec step 5.3 must find the main checkout another way for the hook path (design/plan.md, "What still links
// Bonsai and the studio", item 5). Status and the engine use this.

import (
	"bytes"
	"errors"
	"os/exec"
	"path/filepath"
	"strings"
)

// Checkout is where a command runs: the checkout's own top folder, and its main checkout's. In the main checkout
// both are the same; in a worktree, Main is the checkout the worktree was added from. Both are absolute and real
// (symbolic links resolved), in the system's own separators: use filepath.ToSlash to store or print one.
type Checkout struct {
	Root string
	Main string
}

// Find finds the checkout holding dir.
func Find(dir string) (*Checkout, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, &Error{File: "the current folder", Msg: "cannot be made absolute: " + oneLine(err), Err: err,
			Next: "run the command from a folder inside the project"}
	}
	cmd := exec.Command("git", "-C", abs, "rev-parse", "--show-toplevel", "--git-common-dir")
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return nil, &Error{File: "git", Msg: "is not on the PATH: Bonsai finds a project through git", Err: errNoGit,
				Next: "install git, or put it on the PATH, then run the command again"}
		}
		msg := strings.TrimSpace(stderr.String())
		if i := strings.IndexByte(msg, '\n'); i >= 0 {
			msg = msg[:i]
		}
		if strings.Contains(msg, "not a git repository") {
			return nil, &Error{File: filepath.ToSlash(abs), Msg: "is not inside a git checkout", Err: err,
				Next: "run the command inside a project's checkout (a new project: git init, then bonsai init)"}
		}
		return nil, &Error{File: filepath.ToSlash(abs), Msg: "git rev-parse failed: " + asciiOnly(msg), Err: err,
			Next: "run the command inside a project's work tree, not inside its .git folder"}
	}
	out := strings.Split(strings.TrimRight(strings.ReplaceAll(stdout.String(), "\r\n", "\n"), "\n"), "\n")
	if len(out) != 2 || out[0] == "" || out[1] == "" {
		return nil, &Error{File: filepath.ToSlash(abs), Msg: "git rev-parse gave an answer Bonsai does not read: " +
			asciiOnly(stdout.String()), Next: "check that git works in this folder (git status), then run the command again"}
	}
	root := filepath.FromSlash(out[0])
	common := filepath.FromSlash(out[1])
	if !filepath.IsAbs(common) {
		common = filepath.Join(abs, common)
	}
	main := root
	if filepath.Base(common) == ".git" {
		main = filepath.Dir(common)
	}
	c := &Checkout{}
	for _, p := range []struct {
		from string
		to   *string
	}{{root, &c.Root}, {main, &c.Main}} {
		real, err := filepath.EvalSymlinks(p.from)
		if err != nil {
			return nil, &Error{File: filepath.ToSlash(p.from), Msg: "cannot be resolved: " + oneLine(err), Err: err,
				Next: "check that the checkout's folder exists and can be read"}
		}
		*p.to = real
	}
	return c, nil
}
