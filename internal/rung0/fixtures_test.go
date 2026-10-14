package rung0

// Rung 0's section of contract §13's fixtures (formats/active-task/, set 4; "5.4 done", check 7), every case, built as
// formats/README.md says: each project folder copied into a temporary folder and made a git repository with real git
// (Windows git on Windows), every file committed, its main checkout recorded on a Bonsai home of the case's own so it
// is verified; a worktree case's worktree added on a branch of its own with git worktree add, the case's worktree/
// files copied over it. The active-task function is given the case's --task and BONSAI_TASK (the environment counting
// only when the session's project is the checkout's), as the runner gives it, and rung 0 then judges a new file under
// each glob any of the case's tasks grant, and one no task grants, in the case's checkout:
//
//   - grants names a task: the paths its bonsai.allows (the main checkout's copy) holds pass, at any status; every
//     other is refused (protected), the reason naming that task;
//   - grants is null: every one is refused (protected), the reason saying no task is named, or why there is none.

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/guard"
	"github.com/LastStep/Bonsai/internal/testpack"
	"github.com/LastStep/Bonsai/internal/workspace"
)

var fixtureDir = filepath.Join("..", "..", "formats", "active-task")

// fixtureCase is a case's case.json.
type fixtureCase struct {
	Layout   string  `json:"layout"`
	Checkout string  `json:"checkout"`
	Session  string  `json:"session"`
	Task     *string `json:"task"`
	Env      *string `json:"env"`
}

func TestTheRung0Fixtures(t *testing.T) {
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
	if len(cases) != len(answers.Rung0) || len(cases) != 17 {
		t.Fatalf("%d case folders, %d rung0 answers; want 17 of each (set 8 added named-garbage)", len(cases), len(answers.Rung0))
	}
	for i, name := range cases {
		a := answers.Rung0[i]
		if a.Case != name {
			t.Fatalf("answer %d is for %s, the folder is %s", i, a.Case, name)
		}
		t.Run(name, func(t *testing.T) { rung0Fixture(t, name, a.Grants) })
	}
}

// rung0Fixture builds one case and checks rung 0's answer (this file's comment).
func rung0Fixture(t *testing.T, name string, grants *string) {
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
	folders := map[string]string{}
	for _, p := range []string{"main", "other"} {
		src := filepath.Join(fixtureDir, name, p)
		if _, err := os.Stat(src); err != nil {
			continue
		}
		dir := filepath.Join(tmp, p)
		copyTree(t, src, dir)
		testpack.Git(t, dir, "init", "-q")
		commit(t, dir, "the fixture")
		cfg, err := workspace.LoadConfig(dir)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := workspace.RecordCheckout(home, dir, cfg.ID, since); err != nil {
			t.Fatal(err)
		}
		folders[p] = dir
	}
	mainOf := map[string]string{"main": folders["main"], "other": folders["other"], "worktree": folders["main"]}
	if c.Layout == "worktree" {
		wt := filepath.Join(tmp, "worktree")
		testpack.Git(t, folders["main"], "worktree", "add", "-q", "-b", "fixture-wt", wt)
		if src := filepath.Join(fixtureDir, name, "worktree"); fileExists(src) {
			copyTree(t, src, wt)
		}
		folders["worktree"] = wt
	}
	checkout, main := folders[c.Checkout], mainOf[c.Checkout]
	cfg, err := workspace.LoadConfigFull(main)
	if err != nil {
		t.Fatal(err)
	}
	a, err := workspace.Active(workspace.ActiveInput{Main: main, TaskDir: cfg.Full.Documents.Task, Task: deref(c.Task),
		Env: deref(c.Env), EnvCounts: mainOf[c.Session] == main})
	if err != nil {
		t.Fatal(err)
	}
	// Every glob any task of the case grants (main's, other's and the worktree's copies), and one no task grants.
	globs := []string{"assets/granted-by-no-task/**"}
	for _, dir := range folders {
		for _, list := range allowsIn(t, dir) {
			for _, g := range list {
				if !contains(globs, g) {
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
		want = allowsIn(t, main)[*grants]
	}
	probes := map[string]string{} // a probe's path to its glob
	for _, g := range globs {
		rel := strings.TrimSuffix(g, "**") + "probe.txt"
		write(t, checkout, rel, "probe\n")
		probes[rel] = g
	}
	base := workspace.ReadGitLayout(main).Branch()
	o := Judge(Input{Root: checkout, Main: main, Home: home, Active: a, Base: base, MergeBase: mergeBase(t, checkout, base)})
	refused := refusedRules(o)
	for rel, g := range probes {
		switch {
		case contains(want, g):
			if _, no := refused[rel]; no {
				t.Errorf("%s (granted by %s) is refused:\n%s", rel, *grants, o.Reason)
			}
		case refused[rel] != guard.RuleProtected:
			t.Errorf("%s: refused as %q, want protected (grants %v):\n%s", rel, refused[rel], deref(grants), o.Reason)
		}
	}
	for rel := range refused {
		if _, probe := probes[rel]; !probe {
			t.Errorf("%s, not a probe, is refused:\n%s", rel, o.Reason)
		}
	}
	switch {
	case grants != nil && len(refused) > 0 && !strings.Contains(o.Reason, "the named task "+*grants+" does not grant it"):
		t.Errorf("the refusal does not name %s:\n%s", *grants, o.Reason)
	case grants == nil && !strings.Contains(o.Reason, "no task is named") && !strings.Contains(o.Reason, "no active task: "):
		t.Errorf("the refusal does not say why no task grants it:\n%s", o.Reason)
	}
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
		write(t, to, filepath.ToSlash(rel), string(raw))
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
		parts := strings.SplitN(filepath.Base(n), "-", 3)
		id := parts[0] + "-" + parts[1]
		m := allowsLine.FindStringSubmatch(string(raw))
		if m == nil {
			out[id] = append(out[id], []string{}...)
			continue
		}
		for _, g := range strings.Split(m[1], ",") {
			out[id] = append(out[id], strings.Trim(strings.TrimSpace(g), `"`))
		}
	}
	return out
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}
