package engine

// The Git Bash lookup (checkgitbash.go, plan-5 5.3.6 note 6): where it looks, in order, run on both sides with its
// reads given (a scratch Git layout in t.TempDir, a fake environment and PATH lookup), and the real lookup on each.

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func touch(t *testing.T, p string) string {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestLookGitBash(t *testing.T) {
	tmp := t.TempDir()
	gitRoot := filepath.Join(tmp, "Program Files", "Git")
	bash := touch(t, filepath.Join(gitRoot, "bin", "bash.exe"))
	cmdGit := touch(t, filepath.Join(gitRoot, "cmd", "git.exe"))
	mingwGit := touch(t, filepath.Join(gitRoot, "mingw64", "bin", "git.exe"))
	other := touch(t, filepath.Join(tmp, "elsewhere", "bash.exe"))
	wrongName := touch(t, filepath.Join(tmp, "elsewhere", "wrapper.exe"))
	userGit := filepath.Join(tmp, "Local", "Programs", "Git")
	userBash := touch(t, filepath.Join(userGit, "bin", "bash.exe"))
	noGit := func(string) (string, error) { return "", errors.New("not on the PATH") }
	gitAt := func(p string) func(string) (string, error) { return func(string) (string, error) { return p, nil } }
	env := func(v string) func(string) string {
		return func(k string) string {
			if k == "CLAUDE_CODE_GIT_BASH_PATH" {
				return v
			}
			return ""
		}
	}
	slash := filepath.ToSlash
	for _, c := range []struct {
		name     string
		getenv   func(string) string
		lookPath func(string) (string, error)
		installs []string
		want     string
	}{
		{"the variable names a bash.exe", env(other), gitAt(cmdGit), nil, other},
		{"the variable names a file of another name: ignored, git on the PATH", env(wrongName), gitAt(cmdGit), nil, bash},
		{"the variable names nothing there: ignored", env(filepath.Join(tmp, "gone", "bash.exe")), gitAt(cmdGit), nil, bash},
		{"git in Git's cmd folder", env(""), gitAt(cmdGit), nil, bash},
		{"git in Git's mingw64/bin folder", env(""), gitAt(mingwGit), nil, bash},
		{"no git on the PATH: Git's install folder for all users", env(""), noGit, []string{gitRoot}, bash},
		{"no git on the PATH: a per-user install", env(""), noGit, []string{filepath.Join(tmp, "nowhere"), userGit}, userBash},
		{"nowhere", env(""), noGit, []string{filepath.Join(tmp, "nowhere")}, ""},
		{"a git with no Git Bash beside it", env(""), gitAt(touch(t, filepath.Join(tmp, "plain", "cmd", "git.exe"))), nil, ""},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := lookGitBash(c.getenv, c.lookPath, c.installs)
			if !g.Applies {
				t.Errorf("a lookup on Windows always applies")
			}
			want := ""
			if c.want != "" {
				want = slash(c.want)
			}
			if g.Path != want {
				t.Errorf("found %q, want %q (looked in %v)", g.Path, want, g.Looked)
			}
			if len(g.Looked) == 0 {
				t.Errorf("it looked nowhere")
			}
			for _, p := range append(g.Looked, g.Path) {
				if strings.Contains(p, `\`) {
					t.Errorf("%q: not forward slashes", p)
				}
			}
		})
	}
}

// The finding names the places looked in; a lookup that found Git Bash, or one that does not apply, gives none.
func TestCheckGitBash(t *testing.T) {
	for _, c := range []struct {
		g    GitBash
		want bool
	}{
		{GitBash{Applies: true, Looked: []string{"C:/Program Files/Git/bin/bash.exe"}}, true},
		{GitBash{Applies: true, Path: "C:/Program Files/Git/bin/bash.exe"}, false},
		{GitBash{}, false},
	} {
		g := c.g
		r := &CheckResult{gitBash: func() GitBash { return g }}
		r.checkGitBash()
		if got := len(r.Findings) == 1 && r.Findings[0].Code == "git-bash"; got != c.want {
			t.Errorf("%+v: %+v, want a finding %v", c.g, r.Findings, c.want)
		}
		if c.want && (!strings.Contains(r.Findings[0].Message, "looked in: C:/Program Files/Git/bin/bash.exe") ||
			r.Findings[0].Who != "person" || r.Findings[0].File != "") {
			t.Errorf("the finding: %+v", r.Findings[0])
		}
	}
}

// The real lookup: elsewhere than Windows it does not apply; on Windows it applies, and finds the Git Bash that this PC
// and CI's Windows runner have (Git for Windows, for all users). The Windows half runs on Windows only: the lookup
// reads Windows' own places.
func TestFindGitBash(t *testing.T) {
	g := FindGitBash()
	if runtime.GOOS != "windows" {
		if g.Applies || g.Path != "" || len(g.Looked) != 0 {
			t.Errorf("off Windows: %+v, want nothing", g)
		}
		return
	}
	if !g.Applies || !strings.HasSuffix(strings.ToLower(g.Path), "/bin/bash.exe") {
		t.Errorf("on Windows: %+v, want Git Bash found", g)
	}
}
