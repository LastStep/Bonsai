package rung0

// The tests' fixture projects: a linked project as rung 0 meets one, made by hand in a temporary folder and made a
// git repository with real git (Windows git on Windows), its main checkout recorded on a Bonsai home of the test's
// own, so it is verified; a task worktree added with git worktree add on a branch of its own.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/testpack"
	"github.com/LastStep/Bonsai/internal/workspace"
)

const testID = "ws-r0r0r0r0r0r0r0r0r0r0r0r0r0"

// projectYAML is the fixture project's bonsai.yaml: assets/** protected, secret/** and .claude/** person-only.
const projectYAML = `format: bonsai.workspace/1
id: ` + testID + `
name: rung0-test
packs: []
documents:
  task: work/tasks
  run: work/runs
  answers: work/answers.md
  memory: work/memory
  protocols: work/protocols
protected: ["assets/**", "secret/**", ".claude/**"]
person_only: ["secret/**", ".claude/**"]
never_edit: []
ladder_floor: [0]
ladder: []
ratchets: {}
ci_marked_tests: []
generated:
  log:
    keep_days: 30
    keep_newest: null
  asks:
    keep_days: null
    keep_newest: null
  ladder:
    keep_days: 7
    keep_newest: null
  run:
    keep_days: null
    keep_newest: null
  sessions:
    keep_days: null
    keep_newest: null
`

// lockJSON is a lock with no pack and no file: what a project linked with packs: [] and no lines of Bonsai's holds,
// for the tests' purposes.
const lockJSON = `{
  "format": "bonsai.lock/1",
  "written_by": "0.0.0-test",
  "packs": [],
  "files": {},
  "format0": {}
}
`

// taskMD is a format-1 task file: id, status and its bonsai.allows (a YAML flow list, "" for none).
func taskMD(id, status, allows string) string {
	labels := "labels: {}\n"
	if allows != "" {
		labels = "labels:\n  bonsai.allows: " + allows + "\n"
	}
	return "---\nformat: bonsai.task/1\nid: " + id + "\ntitle: A task of rung 0's tests\nstatus: " + status +
		"\nlane: light\ndone_when:\n  - \"Rung 0 judges it\"\ndepends_on: []\nblocked_by:\ncreated: 2026-10-13\nstarted: 2026-10-13\n" +
		"finished:\n" + labels + "---\n\n# " + id + "\n\nA made-up task of rung 0's tests.\n"
}

// fx is one fixture project: its main checkout, the Bonsai home it is recorded in, and its task worktree when made.
type fx struct {
	tmp, home, main, wt string
}

var since = time.Date(2026, 10, 13, 9, 0, 0, 0, time.UTC)

// newProject makes the fixture project's main checkout on branch main, files (a path to its content) written over the
// standard ones, committed, and recorded on the home. A file mapped to "" is left out.
func newProject(t *testing.T, files map[string]string) *fx {
	t.Helper()
	tmp, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	f := &fx{tmp: tmp, home: testpack.Isolate(t, tmp), main: filepath.Join(tmp, "main")}
	all := map[string]string{
		"bonsai.yaml":                   projectYAML,
		workspace.LockFile:              lockJSON,
		workspace.GitignoreFile:         workspace.GitignoreText,
		"work/tasks/T-0901-the-task.md": taskMD("T-0901", "running", `["assets/t-0901/**", "secret/granted/**", ".claude/**"]`),
		"work/tasks/T-0902-todo.md":     taskMD("T-0902", "todo", ""),
		"src/main.txt":                  "the code\n",
		"assets/old/a.txt":              "an asset\n",
	}
	for p, c := range files {
		all[p] = c
	}
	for p, c := range all {
		if c != "" {
			write(t, f.main, p, c)
		}
	}
	testpack.Git(t, f.main, "init", "-q")
	testpack.Git(t, f.main, "add", "-A")
	testpack.Git(t, f.main, "commit", "-q", "-m", "the base")
	if _, err := workspace.RecordCheckout(f.home, f.main, testID, since); err != nil {
		t.Fatal(err)
	}
	return f
}

// worktree adds the task worktree on branch task, from main's HEAD.
func (f *fx) worktree(t *testing.T) string {
	t.Helper()
	f.wt = filepath.Join(f.tmp, "wt")
	testpack.Git(t, f.main, "worktree", "add", "-q", "-b", "task", f.wt)
	return f.wt
}

// commit commits every change in a checkout.
func commit(t *testing.T, dir, msg string) {
	t.Helper()
	testpack.Git(t, dir, "add", "-A")
	testpack.Git(t, dir, "commit", "-q", "--allow-empty", "-m", msg)
}

// climb runs rung 0 in the checkout root of the project, --task naming task ("" for none), the base branch and the
// merge base worked out as the runner does (the branch the main checkout holds, and git merge-base).
func (f *fx) climb(t *testing.T, root, task string) Outcome {
	t.Helper()
	cfg, err := workspace.LoadConfigFull(f.main)
	if err != nil {
		t.Fatal(err)
	}
	a, err := workspace.Active(workspace.ActiveInput{Main: f.main, TaskDir: cfg.Full.Documents.Task, Task: task})
	if err != nil {
		t.Fatal(err)
	}
	base := workspace.ReadGitLayout(f.main).Branch()
	mb := ""
	if base != "" {
		mb = mergeBase(t, root, base)
	}
	return Judge(Input{Root: root, Main: f.main, Home: f.home, Active: a, Base: base, MergeBase: mb})
}

// mergeBase is git merge-base of HEAD and base in root, "" when there is none.
func mergeBase(t *testing.T, root, base string) string {
	t.Helper()
	out, err := gitIn(root).run("merge-base", "HEAD", base)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func write(t *testing.T, root, rel, body string) {
	t.Helper()
	p := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func remove(t *testing.T, root, rel string) {
	t.Helper()
	if err := os.Remove(filepath.Join(root, filepath.FromSlash(rel))); err != nil {
		t.Fatal(err)
	}
}

// refusedRules maps each refused path to its rule.
func refusedRules(o Outcome) map[string]string {
	out := map[string]string{}
	for _, r := range o.Refused {
		out[r.Path] = r.Rule
	}
	return out
}
