package ladder

// The rung0 section of contract §13's active-task fixtures (formats/active-task/, set 4) climbed through bonsai ladder,
// every case where it fits a climb, built as internal/rung0's fixture test builds them (each project folder copied into
// a temporary folder and made a git repository with real git, its main checkout recorded on a home of the case's own; a
// worktree case's worktree added with git worktree add and its files copied over it). Each copy's bonsai.yaml gets the
// one rung its floor names, rung 0 (the fixtures' ladder is empty), and the lock and .bonsai/.gitignore a linked
// project holds, so bonsai check finds nothing about the commit. The case's --task, BONSAI_TASK and session (as
// CLAUDE_PROJECT_DIR) are the climb's. A probe is a new file under each glob any of the case's tasks grant, and one no
// task grants:
//
//   - grants names a task (the active-task function names it): the climb with the probes that task grants is green;
//     with every probe, red, each probe it does not grant refused (protected) and named in next;
//   - grants is null: a local climb is no-task (exit 2), since no task is named; a --ci climb runs with nothing
//     granted, every probe refused, unless the case names a task that is none (named-differ, -duplicate, -missing,
//     -broken), which --ci refuses too.

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

	"github.com/LastStep/Bonsai/internal/testpack"
	"github.com/LastStep/Bonsai/internal/workspace"
)

var fixtureDir = filepath.Join("..", "..", "formats", "active-task")

// rung0Rung is rung 0's definition, put in each copy's empty ladder.
const rung0Rung = "ladder:\n  - rung: 0\n    name: \"guard\"\n    kind: guard\n    command: null\n    required: true\n    timeout_s: null\n" +
	"    ratchet: null\n    capture: null\n    means: \"Rung 0 of the test.\"\n    tests: null\n    base_setup: null\n"

func TestRung0Fixtures(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join(fixtureDir, "answers.json"))
	if err != nil {
		t.Fatal(err)
	}
	var answers struct {
		Rung0 []struct {
			Case   string  `json:"case"`
			Grants *string `json:"grants"`
		} `json:"rung0"`
	}
	if err := json.Unmarshal(raw, &answers); err != nil {
		t.Fatal(err)
	}
	if len(answers.Rung0) != 16 {
		t.Fatalf("%d rung0 answers, want 16", len(answers.Rung0))
	}
	climbed := map[string]int{}
	for _, a := range answers.Rung0 {
		t.Run(a.Case, func(t *testing.T) { climbed[rung0FixtureClimb(t, a.Case, a.Grants)]++ })
	}
	if climbed["local"] != 4 || climbed["ci"] != 8 || climbed["refused"] != 4 {
		t.Errorf("the cases climbed: %v", climbed)
	}
}

// rung0FixtureClimb builds one case and climbs it (this file's comment); it gives how: local, ci or refused.
func rung0FixtureClimb(t *testing.T, name string, grants *string) string {
	raw, err := os.ReadFile(filepath.Join(fixtureDir, name, "case.json"))
	if err != nil {
		t.Fatal(err)
	}
	var c struct {
		Layout, Checkout, Session string
		Task, Env                 *string
	}
	if err := json.Unmarshal(raw, &c); err != nil {
		t.Fatal(err)
	}
	tmp, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	f := &fx{t: t, tmp: tmp, env: map[string]string{}, lockPoll: 20 * time.Millisecond, lockGrace: 300 * time.Millisecond}
	f.home = testpack.Isolate(t, tmp)
	folders := map[string]string{}
	for _, p := range []string{"main", "other"} {
		src := filepath.Join(fixtureDir, name, p)
		if _, err := os.Stat(src); err != nil {
			continue
		}
		dir := filepath.Join(tmp, p)
		copyFolder(t, src, dir)
		yaml, err := os.ReadFile(filepath.Join(dir, "bonsai.yaml"))
		if err != nil || !strings.Contains(string(yaml), "\nladder: []\n") {
			t.Fatalf("%s's bonsai.yaml: %v", p, err)
		}
		f.writeIn(dir, "bonsai.yaml", strings.Replace(string(yaml), "ladder: []\n", rung0Rung, 1))
		f.writeIn(dir, workspace.LockFile, r0Lock)
		f.writeIn(dir, workspace.GitignoreFile, workspace.GitignoreText)
		testpack.Git(t, dir, "init", "-q", "-b", "main")
		f.commitIn(dir, "the fixture")
		cfg, err := workspace.LoadConfig(dir)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := workspace.RecordCheckout(f.home, dir, cfg.ID, time.Now()); err != nil {
			t.Fatal(err)
		}
		folders[p] = dir
	}
	mainOf := map[string]string{"main": folders["main"], "other": folders["other"], "worktree": folders["main"]}
	if c.Layout == "worktree" {
		wt := filepath.Join(tmp, "worktree")
		testpack.Git(t, folders["main"], "worktree", "add", "-q", "-b", "fixture-wt", wt)
		if src := filepath.Join(fixtureDir, name, "worktree"); exists(src) {
			copyFolder(t, src, wt)
		}
		folders["worktree"] = wt
	}
	checkout, main := folders[c.Checkout], mainOf[c.Checkout]
	f.root = main
	if c.Env != nil {
		f.env[workspace.TaskEnv] = *c.Env
	}
	f.env[ProjectEnv] = folders[c.Session]
	task := ""
	if c.Task != nil {
		task = *c.Task
	}
	// The probes: a file under each glob any task of the case grants, and one no task grants.
	globs := []string{"assets/granted-by-no-task/**"}
	for _, dir := range folders {
		for _, list := range grantsIn(t, dir) {
			for _, g := range list {
				if !contains(globs, g) {
					globs = append(globs, g)
				}
			}
		}
	}
	sort.Strings(globs)
	probe := func(g string) string { return strings.TrimSuffix(g, "**") + "probe.txt" }

	if grants != nil {
		granted := grantsIn(t, main)[*grants]
		for _, g := range granted {
			f.writeIn(checkout, probe(g), "probe\n")
		}
		a := f.climb(checkout, Options{Task: task})
		if a.Exit != 0 || a.Doc.Task == nil || *a.Doc.Task != *grants || statuses(f.result(*grants)) != "0:green" {
			t.Fatalf("the granted probes: exit %d, %+v\n%s", a.Exit, a.Error, f.out.String())
		}
		for _, g := range globs {
			f.writeIn(checkout, probe(g), "probe\n")
		}
		a = f.climb(checkout, Options{Task: task})
		reason := reasonOf(f, *grants, 0)
		if a.Exit != ExitRed || a.Doc.Next == nil {
			t.Fatalf("every probe: exit %d, %+v\n%s", a.Exit, a.Error, f.out.String())
		}
		for _, g := range globs {
			p := strings.TrimSuffix(g, "**") + "probe.txt"
			refused := strings.Contains(reason, `- "`+p+`" (protected): `)
			named := strings.Contains(a.Doc.Next.Do, `"`+p+`" (protected) needs a grant in `+*grants+`'s bonsai.allows, or a revert`)
			if contains(granted, g) == (refused || named) {
				t.Errorf("%s: granted %v, refused %v, in next %v:\n%s\n%s", p, contains(granted, g), refused, named, reason, a.Doc.Next.Do)
			}
		}
		if !strings.Contains(reason, "the named task "+*grants+" does not grant it") {
			t.Errorf("the refusal does not name %s:\n%s", *grants, reason)
		}
		return "local"
	}
	for _, g := range globs {
		f.writeIn(checkout, probe(g), "probe\n")
	}
	if a := f.climb(checkout, Options{Task: task}); a.Exit != ExitInput || a.Error == nil || a.Error.Code != "no-task" {
		t.Fatalf("a local climb with no task named: exit %d, %+v", a.Exit, a.Error)
	}
	cfg, err := workspace.LoadConfigFull(main)
	if err != nil {
		t.Fatal(err)
	}
	fn, err := workspace.Active(workspace.ActiveInput{Main: main, TaskDir: cfg.Full.Documents.Task, Task: task, Env: f.env[workspace.TaskEnv],
		EnvCounts: mainOf[c.Session] == main})
	if err != nil {
		t.Fatal(err)
	}
	a := f.climb(checkout, Options{Task: task, CI: true})
	if strings.HasPrefix(fn.Why, "named-") {
		if a.Exit != ExitInput || a.Error == nil || a.Error.Code != "no-task" {
			t.Errorf("--ci naming a task that is none (%s): exit %d, %+v", fn.Why, a.Exit, a.Error)
		}
		return "refused"
	}
	reason := reasonOf(f, "ci", 0)
	if a.Exit != ExitRed || a.Doc.Task != nil {
		t.Fatalf("--ci with nothing named: exit %d, %+v\n%s", a.Exit, a.Error, f.out.String())
	}
	for _, g := range globs {
		if p := probe(g); !strings.Contains(reason, `- "`+p+`" (protected): `) {
			t.Errorf("%s is not refused:\n%s", p, reason)
		}
	}
	if !strings.Contains(reason, "judged as the guard judges an edit by no named task (nothing granted)") {
		t.Errorf("the reason:\n%s", reason)
	}
	return "ci"
}

// copyFolder copies a folder's files into another.
func copyFolder(t *testing.T, from, to string) {
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
		dst := filepath.Join(to, rel)
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		return os.WriteFile(dst, raw, 0o644)
	})
	if err != nil {
		t.Fatal(err)
	}
}

var allowsLine = regexp.MustCompile(`(?m)^  bonsai\.allows: \[(.*)\]\r?$`)

// grantsIn gives each task file's id and bonsai.allows globs in a project folder's work/tasks.
func grantsIn(t *testing.T, dir string) map[string][]string {
	t.Helper()
	out := map[string][]string{}
	names, _ := filepath.Glob(filepath.Join(dir, "work", "tasks", "*.md"))
	for _, n := range names {
		raw, err := os.ReadFile(n)
		if err != nil {
			t.Fatal(err)
		}
		parts := strings.SplitN(filepath.Base(n), "-", 3)
		id := parts[0] + "-" + parts[1]
		m := allowsLine.FindStringSubmatch(string(raw))
		if m == nil {
			continue
		}
		for _, g := range strings.Split(m[1], ",") {
			out[id] = append(out[id], strings.Trim(strings.TrimSpace(g), `"`))
		}
	}
	return out
}

func exists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}
