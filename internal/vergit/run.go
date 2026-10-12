package vergit

// Runner: git run in this package's fixed environment, for other packages. It is the one way Bonsai runs git where
// nothing the user or the repository has set may steer what git reads (git.go: the variables kept and added, the
// settings pinned on every command line, the user's own excludes file and line-end settings read once and kept); a
// later caller with that need uses it rather than building its own. New tests must fail on the base (internal/newtests,
// 5.4.5) is its first: it makes and removes a temporary worktree at the base with it, and lists the branch's changed
// test files.
//
//	r, note, err := vergit.NewRunner(ctx, "", root)
//	out, err := r.Run("diff", "--name-status", "-z", base, "--")
//	if vergit.ExitCode(err) == 1 { ... }

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
)

// Runner runs git in one checkout, in this package's fixed environment, within its context's deadline.
type Runner struct {
	g *gitRunner
}

// NewRunner finds the git program (prog, or git on this process's PATH when prog is ""), builds the fixed environment
// for the checkout whose top folder is dir (git never looks above it for another), and reads once, with the user's
// own environment, the settings every call keeps: the user's global excludes file and core.autocrlf and core.eol, so a
// file git checks out or compares reads as the user's git reads it. note is set when those settings could not be read
// (git's defaults then stand for them). Every call ends with ctx: its deadline, or its cancel.
func NewRunner(ctx context.Context, prog, dir string) (r *Runner, note string, err error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, "", fmt.Errorf("the checkout's folder cannot be read: %v", oneLine(err.Error()))
	}
	if prog != "" {
		if _, err := exec.LookPath(prog); err != nil {
			return nil, "", fmt.Errorf("git cannot be found: %v", oneLine(err.Error()))
		}
	}
	g, err := newRunner(ctx, prog, abs)
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return nil, "", errors.New("git is not on the PATH")
		}
		return nil, "", fmt.Errorf("git cannot be found: %v", err)
	}
	note = g.userSettings()
	if ctx.Err() != nil {
		return nil, "", errors.New("git did not answer in time")
	}
	return &Runner{g: g}, note, nil
}

// Run runs git with these arguments after the pinned settings (git.go's pins, the excludes file and the user's
// line-end settings) and gives its standard output. An argument list may start with more -c settings of its own,
// which git reads before the command. A failed call's error names the call and the first line git printed; ExitCode
// reads its exit code.
func (r *Runner) Run(args ...string) ([]byte, error) {
	return r.g.run(nil, args...)
}

// ExitCode is a failed Run's exit code: -1 when git did not run to an exit (it could not start, or the context ended
// it), or when err is not a Run's.
func ExitCode(err error) int {
	return exitCode(err)
}
