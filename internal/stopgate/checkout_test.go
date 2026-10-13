package stopgate

// The session's checkout (contract §13: the stop gate's project is the one holding the session's working folder, and
// its ladder result must prove that checkout's HEAD). Claude Code keeps CLAUDE_PROJECT_DIR at the folder a session
// started in, while the payload's cwd follows the session: after its EnterWorktree tool, CLAUDE_PROJECT_DIR is still
// the main checkout and cwd the worktree (measured on 2.1.296), and in a `claude --worktree` session both are the
// worktree.
// The gate judges the git checkout holding cwd (the nearest folder at or above it holding a .git) when git's own link
// ties it to the session's verified main checkout: that main checkout itself, or a worktree of it whose .git names the
// main checkout's .git/worktrees/<name> and whose gitdir there names it back, with or without a bonsai.yaml of its own.
// A cwd in a checkout of the project that git's link does not tie to it (a broken link back, another repository inside
// the project, a .git Bonsai cannot read) fails closed: the result must prove that checkout's HEAD as well as the
// project's. Any other cwd (none, one in no checkout or in a repository outside the project) leaves
// CLAUDE_PROJECT_DIR's checkout judged, as before.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// wtSHA is the commit at the HEAD of a test worktree: the worktree's own commit, after the main checkout's headSHA.
const wtSHA = "9a8b7c6d5e4f30211203f4e5d6c7b8a9f0e1d2c3"

// worktree adds a worktree of the project whose main checkout is main, made by hand as git worktree add makes one: at
// <main>/.claude/worktrees/<name>, where Claude Code's EnterWorktree puts it, when inside is true, else beside main.
// Its .git file names its git folder, <main>/.git/worktrees/<name>, which holds its HEAD (branch worktree-<name>, at
// sha in main's refs), the commondir back to main's .git, and git's link back to its .git file (gitdir). It holds the
// files a checkout of the project's commit holds: bonsai.yaml and the task's run report. It gives the worktree's
// folder.
func worktree(t *testing.T, main, name, sha string, inside bool) string {
	t.Helper()
	wt := filepath.Join(main, ".claude", "worktrees", name)
	if !inside {
		wt = filepath.Join(filepath.Dir(main), filepath.Base(main)+"-"+name)
	}
	gitdir := filepath.Join(main, ".git", "worktrees", name)
	write(t, wt, ".git", "gitdir: "+filepath.ToSlash(gitdir)+"\n")
	write(t, gitdir, "HEAD", "ref: refs/heads/worktree-"+name+"\n")
	write(t, gitdir, "commondir", "../..\n")
	write(t, gitdir, "gitdir", filepath.ToSlash(filepath.Join(wt, ".git"))+"\n")
	write(t, main, ".git/refs/heads/worktree-"+name, sha+"\n")
	write(t, wt, "bonsai.yaml", testYAML)
	write(t, wt, "work/runs/R-2026-10-11-T-0901.md", "# R-2026-10-11-T-0901\n\nWhat changed.\n")
	return wt
}

// gateIn runs the gate for the session s1 with CLAUDE_PROJECT_DIR set to dir, BONSAI_TASK to T-0901, and the payload's
// cwd set to cwd ("" leaves it out).
func gateIn(t *testing.T, dir, cwd string) (int, string) {
	t.Helper()
	in := stopPayload("s1", cwd)
	if cwd == "" {
		in = `{"session_id": "s1", "hook_event_name": "Stop", "stop_hook_active": false}`
	}
	return gate(t, env(dir, workspace.TaskEnv, "T-0901"), strings.NewReader(in), untimed)
}

// checkoutIs checks the last record of main's log names the checkout judged: its folder's name and its branch.
func checkoutIs(t *testing.T, main, checkout, branch string) {
	t.Helper()
	r := lastRecord(t, main)
	if r.String("checkout") != checkout || r.String("branch") != branch {
		t.Errorf("the record names checkout %q on %q, want %q on %q: %s", r.String("checkout"), r.String("branch"),
			checkout, branch, schema.Show(r))
	}
}

// A session that entered a worktree (CLAUDE_PROJECT_DIR the main checkout, cwd the worktree or a folder in it) is
// judged by the worktree, whatever its own bonsai.yaml says or whether it has one: the main checkout's green result at
// its own HEAD does not let it end; a result proving the worktree's HEAD does, and the worktree's run report is the one
// read. A `claude --worktree` session (both the worktree) is judged the same way.
func TestTheWorktreeTheSessionIsIn(t *testing.T) {
	noYAML := func(t *testing.T, wt string) { remove(t, wt, "bonsai.yaml") }
	for _, c := range []struct {
		name   string
		inside bool   // the worktree inside the main checkout, as EnterWorktree makes it, or beside it
		sub    string // the cwd's folder below the worktree's top
		fromWT bool   // CLAUDE_PROJECT_DIR the worktree too: a claude --worktree session
		change func(t *testing.T, wt string)
	}{
		{name: "entered: cwd the worktree", inside: true},
		{name: "entered: cwd a folder in the worktree", inside: true, sub: "notes/deep"},
		{name: "entered: a worktree beside the main checkout", inside: false},
		{name: "entered: a worktree with no bonsai.yaml of its own", inside: true, change: noYAML},
		{name: "entered: a folder in a worktree with no bonsai.yaml", inside: true, sub: "notes", change: noYAML},
		{name: "entered: a worktree beside with no bonsai.yaml", inside: false, change: noYAML},
		{name: "entered: a worktree whose bonsai.yaml holds another id", inside: true, change: func(t *testing.T, wt string) {
			write(t, wt, "bonsai.yaml", strings.Replace(testYAML, testID, "ws-cccccccccccccccccccccccccc", 1))
		}},
		{name: "entered: a worktree whose bonsai.yaml does not read", inside: true, change: func(t *testing.T, wt string) {
			write(t, wt, "bonsai.yaml", "format: bonsai.workspace/1\nid: [\n")
		}},
		{name: "a claude --worktree session", inside: true, fromWT: true},
		{name: "a claude --worktree session in a folder of it", inside: true, fromWT: true, sub: "notes"},
	} {
		t.Run(c.name, func(t *testing.T) {
			root := project(t) // the main checkout: a green result proving its own HEAD, headSHA, and its run report
			wt := worktree(t, root, "x", wtSHA, c.inside)
			if c.change != nil {
				c.change(t, wt)
			}
			cwd := wt
			if c.sub != "" {
				cwd = filepath.Join(wt, filepath.FromSlash(c.sub))
				write(t, cwd, "readme.md", "x\n")
			}
			dir := root
			if c.fromWT {
				dir = wt
			}
			code, stderr := gateIn(t, dir, cwd)
			blocks(t, code, stderr, root, RuleOtherCommit, `proves commit "`+headSHA[:12]+`", but this checkout's HEAD is 9a8b7c6d5e4f`,
				"bonsai ladder --task T-0901 at this HEAD")
			checkoutIs(t, root, filepath.Base(wt), "worktree-x")

			writeResult(t, root, "T-0901", greenResult(wtSHA))
			code, stderr = gateIn(t, dir, cwd)
			allows(t, code, stderr, root)
			checkoutIs(t, root, filepath.Base(wt), "worktree-x")
			noLog(t, wt) // a verified worktree's records are its main checkout's

			// The run report read is the worktree's, not the main checkout's.
			remove(t, wt, "work/runs/R-2026-10-11-T-0901.md")
			code, stderr = gateIn(t, dir, cwd)
			blocks(t, code, stderr, root, RuleNoRunReport, "this checkout has no run report for T-0901", "in this checkout")
		})
	}
}

// A cwd in CLAUDE_PROJECT_DIR's own checkout, or in no checkout of the project at all, leaves CLAUDE_PROJECT_DIR's
// checkout judged, as before: the main checkout's green result at its HEAD lets the session end, and a result proving
// the worktree's HEAD does not.
func TestACwdOutsideTheProjectKeepsItsCheckout(t *testing.T) {
	for _, c := range []struct {
		name string
		cwd  func(t *testing.T, root, wt string) string // the payload's cwd; "" leaves it out
	}{
		{name: "no cwd", cwd: func(*testing.T, string, string) string { return "" }},
		{name: "the main checkout itself", cwd: func(_ *testing.T, root, _ string) string { return root }},
		{name: "a folder in the main checkout", cwd: func(t *testing.T, root, _ string) string {
			write(t, root, "docs/x/readme.md", "x\n")
			return filepath.Join(root, "docs", "x")
		}},
		{name: "a folder in no project", cwd: func(t *testing.T, _, _ string) string {
			d, err := filepath.EvalSymlinks(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			return d
		}},
		{name: "a folder that is not there", cwd: func(_ *testing.T, root, _ string) string { return filepath.Join(root, "gone", "away") }},
		{name: "a file, not a folder", cwd: func(_ *testing.T, root, _ string) string { return filepath.Join(root, "bonsai.yaml") }},
		{name: "a relative folder", cwd: func(*testing.T, string, string) string { return filepath.Join(".claude", "worktrees", "x") }},
		{name: "another linked project", cwd: func(t *testing.T, _, _ string) string {
			other, err := filepath.EvalSymlinks(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			const otherID = "ws-bbbbbbbbbbbbbbbbbbbbbbbbbb"
			write(t, other, "bonsai.yaml", strings.Replace(testYAML, testID, otherID, 1))
			write(t, other, "work/runs/R-2026-10-11-T-0901.md", "# R-2026-10-11-T-0901\n")
			write(t, other, ".git/HEAD", wtSHA+"\n")
			if _, err := workspace.RecordCheckout(os.Getenv(workspace.HomeEnv), other, otherID, testNow); err != nil {
				t.Fatal(err)
			}
			return other
		}},
		{name: "a repository outside the project", cwd: func(t *testing.T, _, _ string) string {
			other, err := filepath.EvalSymlinks(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			write(t, other, ".git/HEAD", headSHA+"\n")
			write(t, other, "src/readme.md", "x\n")
			return filepath.Join(other, "src")
		}},
		{name: "a folder holding a bonsai.yaml and no .git", cwd: func(t *testing.T, root, _ string) string {
			write(t, root, "planted/bonsai.yaml", testYAML)
			return filepath.Join(root, "planted")
		}},
	} {
		t.Run(c.name, func(t *testing.T) {
			root := project(t)
			wt := worktree(t, root, "x", wtSHA, true)
			cwd := c.cwd(t, root, wt)
			code, stderr := gateIn(t, root, cwd)
			allows(t, code, stderr, root)
			checkoutIs(t, root, filepath.Base(root), "main")

			writeResult(t, root, "T-0901", greenResult(wtSHA))
			code, stderr = gateIn(t, root, cwd)
			blocks(t, code, stderr, root, RuleOtherCommit, `proves commit "9a8b7c6d5e4f", but this checkout's HEAD is `+headSHA[:12],
				"bonsai ladder --task T-0901 at this HEAD")
		})
	}
}

// A cwd in a checkout of the project that git's own link does not tie to it fails closed: CLAUDE_PROJECT_DIR's
// checkout is judged, and the result must prove the HEAD of the checkout holding cwd as well. A worktree git does not
// link back to, a .git file naming the worktree's git folder from elsewhere, another repository inside the main
// checkout or a worktree (a nested clone, a submodule), a worktree of another main checkout placed inside this one: the
// main checkout's green result at its own HEAD does not let the session end; once that checkout's HEAD is the same
// commit, it does. A .git Bonsai cannot read keeps it going.
func TestAnUntiedCheckoutMustBeProvenToo(t *testing.T) {
	for _, c := range []struct {
		name string
		cwd  func(t *testing.T, root, wt string) string // the payload's cwd: in the untied checkout
		same func(t *testing.T, root, wt string)        // makes the untied checkout's HEAD headSHA; nil when it cannot be read
		at   string                                     // the untied checkout's folder, as the reason prints it
	}{
		{name: "a worktree git does not link back to", at: ".claude/worktrees/x",
			cwd: func(t *testing.T, root, wt string) string {
				write(t, root, ".git/worktrees/x/gitdir", filepath.ToSlash(filepath.Join(root, "elsewhere", ".git"))+"\n")
				return wt
			},
			same: func(t *testing.T, root, _ string) { write(t, root, ".git/refs/heads/worktree-x", headSHA+"\n") }},
		{name: "a .git file naming a worktree's git folder from elsewhere", at: "forged",
			cwd: func(t *testing.T, root, _ string) string {
				write(t, root, "forged/.git", "gitdir: "+filepath.ToSlash(filepath.Join(root, ".git", "worktrees", "x"))+"\n")
				write(t, root, "forged/notes/readme.md", "x\n")
				return filepath.Join(root, "forged", "notes")
			},
			same: func(t *testing.T, root, _ string) { write(t, root, ".git/refs/heads/worktree-x", headSHA+"\n") }},
		{name: "a repository inside the main checkout", at: "vendor/lib",
			cwd: func(t *testing.T, root, _ string) string {
				write(t, root, "vendor/lib/.git/HEAD", otherSHA+"\n")
				return filepath.Join(root, "vendor", "lib")
			},
			same: func(t *testing.T, root, _ string) { write(t, root, "vendor/lib/.git/HEAD", headSHA+"\n") }},
		{name: "a repository inside a worktree", at: ".claude/worktrees/x/vendor/lib",
			cwd: func(t *testing.T, _, wt string) string {
				write(t, wt, "vendor/lib/.git/HEAD", otherSHA+"\n")
				write(t, wt, "vendor/lib/src/readme.md", "x\n")
				return filepath.Join(wt, "vendor", "lib", "src")
			},
			same: func(t *testing.T, _, wt string) { write(t, wt, "vendor/lib/.git/HEAD", headSHA+"\n") }},
		{name: "a submodule's checkout", at: "sub/mod",
			cwd: func(t *testing.T, root, _ string) string {
				write(t, root, "sub/mod/.git", "gitdir: ../../.git/modules/mod\n")
				write(t, root, ".git/modules/mod/HEAD", otherSHA+"\n")
				return filepath.Join(root, "sub", "mod")
			},
			same: func(t *testing.T, root, _ string) { write(t, root, ".git/modules/mod/HEAD", headSHA+"\n") }},
		{name: "a worktree of another main checkout, inside this one", at: ".claude/worktrees/x",
			cwd: func(t *testing.T, _, wt string) string {
				decoy, err := filepath.EvalSymlinks(t.TempDir())
				if err != nil {
					t.Fatal(err)
				}
				write(t, decoy, "bonsai.yaml", testYAML)
				write(t, decoy, ".git/HEAD", "ref: refs/heads/main\n")
				gitdir := filepath.Join(decoy, ".git", "worktrees", "x")
				write(t, wt, ".git", "gitdir: "+filepath.ToSlash(gitdir)+"\n")
				write(t, gitdir, "HEAD", wtSHA+"\n")
				write(t, gitdir, "commondir", "../..\n")
				write(t, gitdir, "gitdir", filepath.ToSlash(filepath.Join(wt, ".git"))+"\n")
				return wt
			},
			same: func(t *testing.T, _, wt string) {
				l := workspace.ReadGitLayout(wt)
				write(t, l.GitDir, "HEAD", headSHA+"\n")
			}},
		{name: "a .git Bonsai cannot read", at: "odd",
			cwd: func(t *testing.T, root, _ string) string {
				write(t, root, "odd/.git", "not a pointer to a git folder\n")
				return filepath.Join(root, "odd")
			}},
	} {
		t.Run(c.name, func(t *testing.T) {
			root := project(t)
			wt := worktree(t, root, "x", wtSHA, true)
			cwd := c.cwd(t, root, wt)
			code, stderr := gateIn(t, root, cwd)
			if c.same == nil {
				blocks(t, code, stderr, root, RuleHeadUnreadable, "the session's working folder is in "+c.at+", ", "cd back into")
				return
			}
			blocks(t, code, stderr, root, RuleOtherCommit, "the session's working folder is in "+c.at+", ", "cd back into")
			if !strings.Contains(stderr, `so the ladder result for T-0901 must prove its HEAD as well as this project's checkout's: `+
				`it proves commit "`+headSHA[:12]+`", and that HEAD is `) {
				t.Errorf("the reason: %q", stderr)
			}
			checkoutIs(t, root, filepath.Base(root), "main")

			c.same(t, root, wt)
			code, stderr = gateIn(t, root, cwd)
			allows(t, code, stderr, root)
			checkoutIs(t, root, filepath.Base(root), "main")
		})
	}
}

// A worktree session whose cwd is its main checkout (a claude --worktree session that changed folder to it) is judged
// by the main checkout: the checkout holding the session's working folder (contract §13).
func TestAWorktreeSessionInItsMainCheckout(t *testing.T) {
	root := project(t)
	wt := worktree(t, root, "x", wtSHA, true)
	code, stderr := gateIn(t, wt, root)
	allows(t, code, stderr, root)
	checkoutIs(t, root, filepath.Base(root), "main")
	writeResult(t, root, "T-0901", greenResult(wtSHA))
	code, stderr = gateIn(t, wt, root)
	blocks(t, code, stderr, root, RuleOtherCommit, "but this checkout's HEAD is "+headSHA[:12], "bonsai ladder --task T-0901")
	// With no cwd, CLAUDE_PROJECT_DIR's checkout, the worktree, is judged, as before.
	code, stderr = gateIn(t, wt, "")
	allows(t, code, stderr, root)
	checkoutIs(t, root, filepath.Base(wt), "worktree-x")
}

// A main checkout that is not verified blocks whatever the cwd: the session's checkout is chosen only once it is.
func TestAnUnverifiedMainBlocksWhateverTheCwd(t *testing.T) {
	root := project(t)
	wt := worktree(t, root, "x", wtSHA, true)
	writeResult(t, root, "T-0901", greenResult(wtSHA))
	t.Setenv(workspace.HomeEnv, t.TempDir())
	for _, dir := range []string{root, wt} {
		code, stderr := gateIn(t, dir, wt)
		blocks(t, code, stderr, "", RuleMainNotVerified, "is not verified", "bonsai ladder --task T-0901")
	}
}
