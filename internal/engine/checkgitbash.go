package engine

// check's finding for Windows without Git Bash (git-bash, plan-5 5.3.6 note 6). Claude Code runs a hook's command in
// Git Bash on Windows, and in PowerShell when Git Bash isn't installed (its hooks reference, "Shell form", and the
// shell field's default). Windows PowerShell 5.1 cannot read `||`, and no PowerShell reads a line in Bonsai's form, so
// the line fails, and a hook that fails is a non-blocking error unless it sets onFailure: "block" (Claude Code 2.1.295
// or later): the call goes through unguarded. So check and status give a finding on Windows when Git Bash is not where
// Claude Code looks for it.
//
// Where Claude Code looks (its env-vars reference and setup page, read on 2.1.294 and 2.1.296): the file
// CLAUDE_CODE_GIT_BASH_PATH names, when it exists and is named bash.exe, sh.exe, bash or sh (any other value is
// ignored, and Git Bash found as if it were unset); else Git Bash found from the git on the PATH. The pages say no
// more, and with no Git folder on its PATH Claude Code 2.1.294 still ran hook lines in Git Bash (step 5.3.6's
// sessions), so FindGitBash also looks in Git for Windows' own install folders (for all users, then for one user). A
// lookup is a function, so a test gives one that finds nothing, on both sides (CheckOptions.GitBash).

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// GitBash is what a lookup for Git Bash found.
type GitBash struct {
	Applies bool     // Claude Code would need Git Bash here: native Windows; false elsewhere, and check looks no further
	Path    string   // where Git Bash was found, forward slashes; "" when it was not
	Looked  []string // the places looked in, in order, forward slashes: the finding names them
}

// FindGitBash looks for Git Bash where Claude Code looks for it, on Windows; elsewhere it does not apply.
func FindGitBash() GitBash {
	if runtime.GOOS != "windows" {
		return GitBash{}
	}
	return lookGitBash(os.Getenv, exec.LookPath, gitInstallFolders(os.Getenv))
}

// gitInstallFolders are Git for Windows' install folders, for all users then for one: <Program Files>\Git and
// <LOCALAPPDATA>\Programs\Git.
func gitInstallFolders(getenv func(string) string) []string {
	var out []string
	for _, v := range []string{"ProgramW6432", "ProgramFiles", "ProgramFiles(x86)"} {
		if p := getenv(v); p != "" {
			out = append(out, filepath.Join(p, "Git"))
		}
	}
	if p := getenv("LOCALAPPDATA"); p != "" {
		out = append(out, filepath.Join(p, "Programs", "Git"))
	}
	return out
}

// gitBashNames are the names Claude Code takes in CLAUDE_CODE_GIT_BASH_PATH (its env-vars reference).
var gitBashNames = []string{"bash.exe", "sh.exe", "bash", "sh"}

// lookGitBash is FindGitBash's lookup, its reads given: the environment, the PATH lookup and Git's install folders.
func lookGitBash(getenv func(string) string, lookPath func(string) (string, error), installs []string) GitBash {
	g := GitBash{Applies: true}
	seen := map[string]bool{}
	try := func(p string) bool {
		key := strings.ToLower(filepath.Clean(p))
		if seen[key] {
			return false
		}
		seen[key] = true
		g.Looked = append(g.Looked, filepath.ToSlash(p))
		if fi, err := os.Stat(p); err == nil && !fi.IsDir() {
			g.Path = filepath.ToSlash(p)
			return true
		}
		return false
	}
	if p := getenv("CLAUDE_CODE_GIT_BASH_PATH"); p != "" {
		named := false
		for _, n := range gitBashNames {
			named = named || strings.EqualFold(filepath.Base(p), n)
		}
		if named && try(p) {
			return g
		}
	}
	if git, err := lookPath("git"); err == nil {
		if abs, err := filepath.Abs(git); err == nil {
			git = abs
		}
		// Git for Windows puts git.exe in <Git>\cmd, <Git>\bin or <Git>\mingw64\bin, and Git Bash at <Git>\bin\bash.exe.
		dir := filepath.Dir(git)
		for _, root := range []string{filepath.Dir(dir), filepath.Dir(filepath.Dir(dir))} {
			if try(filepath.Join(root, "bin", "bash.exe")) {
				return g
			}
		}
	}
	for _, root := range installs {
		if try(filepath.Join(root, "bin", "bash.exe")) {
			return g
		}
	}
	return g
}

// checkGitBash gives the git-bash finding when Git Bash applies here and was not found.
func (r *CheckResult) checkGitBash() {
	look := r.gitBash
	if look == nil {
		look = FindGitBash
	}
	g := look()
	if !g.Applies || g.Path != "" {
		return
	}
	where := "CLAUDE_CODE_GIT_BASH_PATH and the git on the PATH"
	if len(g.Looked) > 0 {
		where = strings.Join(g.Looked, ", ")
	}
	r.add("git-bash", "", "", "Git Bash is not where Claude Code looks for it (looked in: "+ascii(where)+"), so Claude Code runs every hook "+
		"line in PowerShell, where Bonsai's lines do not parse, and a hook that fails does not block: every edit and command in this "+
		"project goes through unguarded", "a person installs Git for Windows for all users (it puts Git Bash in C:/Program Files/Git/bin), "+
		"then run: bonsai check")
}
