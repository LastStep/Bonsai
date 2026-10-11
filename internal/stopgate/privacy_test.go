package stopgate

// What the gate records names no user's folder (contract §2.6: a record carries paths relative to the project, and no
// user folder). A reason names a path relative to the project when it lies inside it, else says where in plain words;
// an operating system's error is said by its kind, never by its text, which names the file by its whole path. The
// printed reason is the recorded one.

import (
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/redact"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// absPathIn finds a whole path in a text: one written from the root (/..., C:\..., C:/..., \\share\...) at the start
// of a word.
var absPathIn = regexp.MustCompile(`(^|[\s"'(=\[])(/[A-Za-z0-9._~-]|[A-Za-z]:[\\/]|\\\\)`)

// holdsNoFolder checks a text holds no whole path, none of the folders given (in either slash), no home folder and no
// user name.
func holdsNoFolder(t *testing.T, what, text string, folders ...string) {
	t.Helper()
	if m := absPathIn.FindString(text); m != "" {
		t.Errorf("%s holds a whole path (%q): %q", what, m, text)
	}
	if home, err := os.UserHomeDir(); err == nil && len(home) > 3 {
		folders = append(folders, home)
	}
	for _, f := range folders {
		for _, form := range []string{f, filepath.ToSlash(f), strings.ReplaceAll(f, `\`, `\\`)} {
			if form != "" && strings.Contains(strings.ToLower(text), strings.ToLower(form)) {
				t.Errorf("%s holds the folder %q: %q", what, form, text)
			}
		}
	}
	if u, err := user.Current(); err == nil {
		name := u.Username[strings.LastIndexAny(u.Username, `\/`)+1:]
		if len(name) >= 3 && strings.Contains(strings.ToLower(text), strings.ToLower(name)) {
			t.Errorf("%s holds the user name %q: %q", what, name, text)
		}
	}
}

// Each blocking case whose reason once carried a whole path: a worktree moved outside the project, a .git outside it
// naming a worktree's git folder, a ladder result that is a folder, a task folder that cannot be read. The session is
// kept going, its record's text and its printed reason hold no folder of this machine, and the two say the same.
func TestTheRecordHoldsNoUserFolder(t *testing.T) {
	for _, c := range []struct {
		name, rule, why string
		set             func(t *testing.T, root string) (cwd string, folders []string)
	}{
		{name: "a worktree moved outside the project", rule: RuleOtherCommit,
			why: "the session's working folder is in a folder outside the main checkout and the session's starting folder, " +
				"a checkout whose .git names this project's main checkout",
			set: func(t *testing.T, root string) (string, []string) {
				wt := worktree(t, root, "x", wtSHA, false)
				moved := wt + "-moved"
				if err := os.Rename(wt, moved); err != nil {
					t.Fatal(err)
				}
				return moved, []string{moved, filepath.Dir(root)}
			}},
		{name: "a .git outside the project naming a worktree's git folder", rule: RuleOtherCommit,
			why: "the session's working folder is in a folder outside the main checkout and the session's starting folder, " +
				"a checkout whose .git names this project's main checkout",
			set: func(t *testing.T, root string) (string, []string) {
				worktree(t, root, "x", wtSHA, true)
				forged, err := filepath.EvalSymlinks(t.TempDir())
				if err != nil {
					t.Fatal(err)
				}
				write(t, forged, ".git", "gitdir: "+filepath.ToSlash(filepath.Join(root, ".git", "worktrees", "x"))+"\n")
				return forged, []string{forged, filepath.Dir(forged), root}
			}},
		{name: "a ladder result that is a folder", rule: RuleBadResult,
			why: "the ladder result .bonsai/local/ladder/T-0901.json cannot be read (it is a folder, not a file)",
			set: func(t *testing.T, root string) (string, []string) {
				remove(t, root, workspace.LocalDir+"/ladder/T-0901.json")
				write(t, root, workspace.LocalDir+"/ladder/T-0901.json/x", "")
				return root, []string{root}
			}},
		{name: "a task folder that cannot be read", rule: RuleTaskUnreadable,
			why: `the main checkout's task folder "work/tasks" cannot be read (work/tasks: a file stands where a folder belongs)`,
			set: func(t *testing.T, root string) (string, []string) {
				if runtime.GOOS == "windows" {
					t.Skip(unreadableOnWindows)
				}
				if err := os.RemoveAll(filepath.Join(root, "work", "tasks")); err != nil {
					t.Fatal(err)
				}
				write(t, root, "work/tasks", "a file where the task folder goes\n")
				return root, []string{root}
			}},
	} {
		t.Run(c.name, func(t *testing.T) {
			root := project(t)
			cwd, folders := c.set(t, root)
			code, stderr := gateIn(t, root, cwd)
			blocks(t, code, stderr, root, c.rule, c.why, "")
			text := lastRecord(t, root).String("text")
			holdsNoFolder(t, "the record's text", text, folders...)
			holdsNoFolder(t, "the printed reason", strings.SplitN(stderr, "\n", 2)[0], folders...)
			// The record keeps the printed reason, cut to the record's text cap.
			printed := strings.TrimSuffix(strings.TrimPrefix(strings.SplitN(stderr, "\n", 2)[0], "bonsai stop gate: "), ".")
			if text != redact.Cut(printed, redact.TextCap) {
				t.Errorf("the recorded reason is not the printed one:\n printed  %q\n recorded %q", printed, text)
			}
		})
	}
}

// noPaths says every whole path in a reason in plain words, a Windows one over its spaces too, and leaves a path
// relative to the project as it is.
func TestNoPaths(t *testing.T) {
	for in, want := range map[string]string{
		"read /home/someone/p/x.json: is a directory":            "read a path on this machine: is a directory",
		`open C:\Users\First Last\p\x.json: Incorrect function.`: "open a path on this machine: Incorrect function.",
		`the folder "C:/Users/someone/p" is gone`:                `the folder "a path on this machine" is gone`,
		`in \\wsl.localhost\Ubuntu\home\someone, and more`:       "in a path on this machine, and more",
		"the ladder result .bonsai/local/ladder/T-0901.json":     "the ladder result .bonsai/local/ladder/T-0901.json",
		"the run folder work/runs (~/x stays)":                   "the run folder work/runs (~/x stays)",
		"/abs at the start; then words":                          "a path on this machine; then words",
	} {
		if got := noPaths(in); got != want {
			t.Errorf("noPaths(%q) = %q, want %q", in, got, want)
		}
	}
}
