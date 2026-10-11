package guard

// The guard's section of contract §13's fixtures (formats/active-task/, set 4; "5.3 done", check 3), every case, in
// agents mode and in command mode (grants do not change with the mode), each case built as formats/README.md says:
// every project folder copied into a temporary folder and made a git repository with real git (Windows git on
// Windows), its main checkout recorded on a Bonsai home of the case's own so it is verified; a worktree case's main/
// committed and a worktree added on a branch of its own with git worktree add, the case's worktree/ files copied over
// it. The guard then runs through Main for a session whose project folder is the case's session, BONSAI_TASK as its
// env, on Writes of a path under each glob any of the case's tasks grant, and of one no task grants:
//
//   - grants names a task: the paths its bonsai.allows (read from the main checkout's copy) holds are allowed
//     (granted); every other is refused, the reason naming that task;
//   - grants is null: every one is refused (protected), the reason saying there is no active task, or that the one
//     named does not read running.

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/testpack"
	"github.com/LastStep/Bonsai/internal/workspace"
)

var fixtureDir = filepath.Join("..", "..", "formats", "active-task")

// fixtureCase is a case's case.json.
type fixtureCase struct {
	Layout   string  `json:"layout"`
	Checkout string  `json:"checkout"`
	Session  string  `json:"session"`
	Env      *string `json:"env"`
}

// copyTree copies a folder's files into another.
func copyTree(t *testing.T, from, to string) {
	t.Helper()
	err := filepath.WalkDir(from, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, err := filepath.Rel(from, p)
		if err != nil {
			return err
		}
		raw, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		writeFile(t, filepath.Join(to, rel), string(raw))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

var allowsLine = regexp.MustCompile(`(?m)^  bonsai\.allows: \[(.*)\]\r?$`)

// allowsIn gives each task file's id and bonsai.allows globs in a project folder's work/tasks.
func allowsIn(t *testing.T, dir string) map[string][]string {
	t.Helper()
	out := map[string][]string{}
	names, _ := filepath.Glob(filepath.Join(dir, "work", "tasks", "*.md"))
	for _, n := range names {
		raw, err := os.ReadFile(n)
		if err != nil {
			t.Fatal(err)
		}
		id := strings.SplitN(filepath.Base(n), "-", 3)
		m := allowsLine.FindStringSubmatch(string(raw))
		if m == nil {
			out[id[0]+"-"+id[1]] = append(out[id[0]+"-"+id[1]], []string{}...)
			continue
		}
		for _, g := range strings.Split(m[1], ",") {
			out[id[0]+"-"+id[1]] = append(out[id[0]+"-"+id[1]], strings.Trim(strings.TrimSpace(g), `"`))
		}
	}
	return out
}

func TestTheGuardFixtures(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join(fixtureDir, "answers.json"))
	if err != nil {
		t.Fatal(err)
	}
	var answers struct {
		Guard []struct {
			Case   string  `json:"case"`
			Grants *string `json:"grants"`
		} `json:"guard"`
	}
	if err := json.Unmarshal(raw, &answers); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(fixtureDir)
	if err != nil {
		t.Fatal(err)
	}
	var cases []string
	for _, e := range entries {
		if e.IsDir() {
			cases = append(cases, e.Name())
		}
	}
	sort.Strings(cases)
	if len(cases) != len(answers.Guard) || len(cases) != 16 {
		t.Fatalf("%d case folders, %d guard answers; want 16 of each", len(cases), len(answers.Guard))
	}
	for i, name := range cases {
		a := answers.Guard[i]
		if a.Case != name {
			t.Fatalf("answer %d is for %s, the folder is %s", i, a.Case, name)
		}
		for _, mode := range []string{"agents", "command"} {
			t.Run(name+"/"+mode, func(t *testing.T) { guardFixture(t, name, mode, a.Grants) })
		}
	}
}

// guardFixture builds one case and checks the guard's answer (this file's comment).
func guardFixture(t *testing.T, name, mode string, grants *string) {
	raw, err := os.ReadFile(filepath.Join(fixtureDir, name, "case.json"))
	if err != nil {
		t.Fatal(err)
	}
	var c fixtureCase
	if err := json.Unmarshal(raw, &c); err != nil {
		t.Fatal(err)
	}
	tmp, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	home := testpack.Isolate(t, tmp)
	since := time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC)
	folders := map[string]string{}
	for _, p := range []string{"main", "other"} {
		src := filepath.Join(fixtureDir, name, p)
		if _, err := os.Stat(src); err != nil {
			continue
		}
		dir := filepath.Join(tmp, p)
		copyTree(t, src, dir)
		testpack.Git(t, dir, "init", "-q")
		if c.Layout == "worktree" && p == "main" {
			testpack.Git(t, dir, "add", "-A")
			testpack.Git(t, dir, "commit", "-q", "-m", "the fixture")
		}
		cfg, err := workspace.LoadConfig(dir)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := workspace.RecordCheckout(home, dir, cfg.ID, since); err != nil {
			t.Fatal(err)
		}
		if mode == "command" {
			machineSettings(t, dir, `{"status_writes": "command", "status_command": "studio-move"}`)
		}
		folders[p] = dir
	}
	if c.Layout == "worktree" {
		wt := filepath.Join(tmp, "worktree")
		testpack.Git(t, folders["main"], "worktree", "add", "-q", "-b", "fixture-wt", wt)
		if src := filepath.Join(fixtureDir, name, "worktree"); fileExists(src) {
			copyTree(t, src, wt)
		}
		folders["worktree"] = wt
	}
	checkout, session := folders[c.Checkout], folders[c.Session]
	var kv []string
	if c.Env != nil {
		kv = []string{workspace.TaskEnv, *c.Env}
	}
	// The globs every task of the case grants (main's, other's and the worktree's copies), and one no task grants.
	globs := []string{"assets/granted-by-no-task/**"}
	for _, dir := range folders {
		for _, list := range allowsIn(t, dir) {
			for _, g := range list {
				if !listHas(globs, g) {
					globs = append(globs, g)
				}
			}
		}
	}
	sort.Strings(globs)
	if len(globs) < 2 {
		t.Fatalf("no task of the case grants anything the test can probe: %v", globs)
	}
	var want []string // the globs the expected task grants, read from the main checkout of the checkout's project
	if grants != nil {
		from := folders["main"]
		if c.Checkout == "other" {
			from = folders["other"]
		}
		want = allowsIn(t, from)[*grants]
	}
	for _, g := range globs {
		rel := strings.TrimSuffix(g, "**") + "probe.txt"
		in := toolPayload("fx", session, "Write", schemaFile(filepath.Join(checkout, filepath.FromSlash(rel)), "probe\n"))
		code, stderr := guardRun(t, env(session, kv...), strings.NewReader(in), untimed)
		switch {
		case listHas(want, g):
			if code != 0 {
				t.Errorf("%s (granted by %s): exit %d, %s", rel, *grants, code, stderr)
			}
		case code != 2:
			t.Errorf("%s: exit %d, want 2 (grants %v)", rel, code, deref(grants))
		case grants != nil && !strings.Contains(stderr, "the running task "+*grants+" "):
			t.Errorf("%s: the refusal does not name %s: %s", rel, *grants, stderr)
		case grants == nil && !strings.Contains(stderr, "no active task: ") && !strings.Contains(stderr, "and a task's grants count only while it reads running"):
			t.Errorf("%s: the refusal does not say why no task grants it: %s", rel, stderr)
		}
	}
}

// schemaFile is a Write's tool_input.
func schemaFile(path, content string) schema.Object {
	return schema.Object{{Key: "file_path", Value: path}, {Key: "content", Value: content}}
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}
