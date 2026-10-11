package main

// check's two findings of step 5.3.6b, their cases in TestCheckTable's walk: a damaged machine record in Bonsai's home
// (machine-record), which un-verifies every main checkout on the machine (step 5.3.1's verifier), and on Windows a
// Git for Windows installed for one user (git-user), whose /usr/local/bin, the lines' second place as Git Bash names
// it, that user can write (plan-5 5.3.6 note 2). Both are proved on both sides: the Git Bash lookup is injected.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/engine"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// damagedRecord writes a machine record Bonsai does not read into the test's Bonsai home, and gives its path.
func damagedRecord(c *cli) string {
	c.t.Helper()
	home, err := workspace.Home()
	if err != nil {
		c.t.Fatal(err)
	}
	p := filepath.Join(home, "workspaces", "r-00000000000000aa", workspace.MachineRecordFile)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		c.t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(`{"path": "/srv/x", "ids": [{"id": "ws-aaaaaaaaaaaaaaaaaaaaaaaaaa", "since": "yesterday"}]}`), 0o644); err != nil {
		c.t.Fatal(err)
	}
	return p
}

// perUserGit is a lookup on Windows that found a Git for Windows installed for one user.
var perUserGit = engine.GitBash{Applies: true, Path: "C:/Users/someone/AppData/Local/Programs/Git/bin/bash.exe",
	Looked: []string{"C:/Users/someone/AppData/Local/Programs/Git/bin/bash.exe"}, PerUser: true}

func init() {
	checkCases["machine-record"] = func(c *cli) { c.link(); damagedRecord(c) }
	// The finding names the record, why it does not read and the person's step; status carries it; mended (deleted),
	// it is gone. The checkout's own record, which reads, is never one.
	checkMore["machine-record"] = func(t *testing.T, c *cli) {
		p := filepath.ToSlash(damagedRecord(c))
		_, out, _ := c.run("", "check", "--json")
		if !strings.Contains(out, "this machine's record "+p+" in Bonsai's home the ids entry of") ||
			!strings.Contains(out, "no main checkout on this machine is verified") ||
			!strings.Contains(out, `"do": "a person checks `+p+` (this machine's record of a checkout) and deletes it`) ||
			strings.Count(out, `"code": "machine-record"`) != 1 {
			t.Errorf("the finding:\n%s", out)
		}
		_, st, _ := c.run("", "status", "--json")
		if !hasProblem(fits(t, st, "status"), "this machine's record "+p, "; next: a person checks") {
			t.Errorf("status --json does not carry the finding:\n%s", st)
		}
		if err := os.Remove(filepath.FromSlash(p)); err != nil {
			t.Fatal(err)
		}
		if _, out, _ := c.run("", "check", "--json"); strings.Contains(out, `"code": "machine-record"`) {
			t.Errorf("a mended home still gives the finding:\n%s", out)
		}
	}
	checkCases["git-user"] = func(c *cli) { c.link(); setGitBash(c.t, perUserGit) }
	// status carries it; Git Bash installed for all users, or no Windows, gives none.
	checkMore["git-user"] = func(t *testing.T, c *cli) {
		_, st, _ := c.run("", "status", "--json")
		if !hasProblem(fits(t, st, "status"), "the Git Bash Claude Code runs hook lines in, "+perUserGit.Path+", is a Git for Windows installed for one user",
			"; next: a person installs Git for Windows for all users") {
			t.Errorf("status --json does not carry the finding:\n%s", st)
		}
		for _, g := range []engine.GitBash{gitBashFound, {}} {
			setGitBash(t, g)
			if _, out, _ := c.run("", "check", "--json"); strings.Contains(out, `"code": "git-user"`) {
				t.Errorf("%+v: a finding all the same:\n%s", g, out)
			}
		}
	}
}
