package workspace

// The active-task function on contract §13's fixtures (formats/active-task/, set 4): every case of answers.json's
// function section gives its answer. Each case is copied into a temporary folder and each project made a git
// repository, as the fixtures' README asks; a worktree case's worktree is added with git and given the case's stale
// copies, so tasks read from main, never the worktree, is what is tested.

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/testpack"
)

var fixtures = filepath.Join("..", "..", "formats", "active-task")

// copyTree copies a folder's files into dst.
func copyTree(t *testing.T, src, dst string) {
	t.Helper()
	err := filepath.WalkDir(src, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(src, p)
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		to := filepath.Join(dst, rel)
		if err := os.MkdirAll(filepath.Dir(to), 0o755); err != nil {
			return err
		}
		return os.WriteFile(to, b, 0o644)
	})
	if err != nil {
		t.Fatal(err)
	}
}

func readJSON(t *testing.T, path string) schema.Object {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	v, err := schema.Decode(raw)
	if err != nil {
		t.Fatal(err)
	}
	return v.(schema.Object)
}

func text(v any) string {
	s, _ := v.(string)
	return s
}

func samePlace(a, b string) bool {
	if runtime.GOOS == "windows" {
		return strings.EqualFold(filepath.Clean(a), filepath.Clean(b))
	}
	return filepath.Clean(a) == filepath.Clean(b)
}

func TestActiveTaskFixtures(t *testing.T) {
	tmp := t.TempDir()
	if r, err := filepath.EvalSymlinks(tmp); err == nil {
		tmp = r
	}
	testpack.Isolate(t, tmp)
	answers := readJSON(t, filepath.Join(fixtures, "answers.json"))
	section, _ := answers.Get("function")
	list := section.([]any)
	if len(list) < 16 {
		t.Fatalf("answers.json's function section holds %d cases", len(list))
	}
	for _, a := range list {
		ans := a.(schema.Object)
		name := ans.String("case")
		t.Run(name, func(t *testing.T) {
			c := readJSON(t, filepath.Join(fixtures, name, "case.json"))
			dir := filepath.Join(tmp, name)
			copyTree(t, filepath.Join(fixtures, name), dir)
			place := map[string]string{}
			for _, project := range []string{"main", "other"} {
				p := filepath.Join(dir, project)
				if _, err := os.Stat(p); err != nil {
					continue
				}
				testpack.Git(t, p, "init", "-q")
				testpack.Git(t, p, "add", "-A")
				testpack.Git(t, p, "commit", "-q", "-m", "the fixture")
				place[project] = p
			}
			if c.String("layout") == "worktree" {
				wt := filepath.Join(dir, "wt")
				testpack.Git(t, place["main"], "worktree", "add", "-q", "-b", "task", wt)
				if _, err := os.Stat(filepath.Join(dir, "worktree")); err == nil {
					copyTree(t, filepath.Join(dir, "worktree"), wt)
				}
				place["worktree"] = wt
			}
			checkout, err := Find(place[c.String("checkout")])
			if err != nil {
				t.Fatal(err)
			}
			session, err := Find(place[c.String("session")])
			if err != nil {
				t.Fatal(err)
			}
			cfg, err := LoadConfigFull(checkout.Main)
			if err != nil {
				t.Fatal(err)
			}
			task, _ := c.Get("task")
			env, _ := c.Get("env")
			got, err := Active(ActiveInput{Main: checkout.Main, TaskDir: cfg.Full.Documents.Task, Task: text(task), Env: text(env),
				EnvCounts: samePlace(checkout.Main, session.Main)})
			if err != nil {
				t.Fatal(err)
			}
			id, _ := ans.Get("id")
			how, _ := ans.Get("how")
			why, _ := ans.Get("why")
			if got.ID != text(id) || got.How != text(how) || got.Why != text(why) {
				t.Errorf("got id %q how %q why %q (%s), want %v %v %v", got.ID, got.How, got.Why, got.Sentence(), id, how, why)
			}
			if (got.ID == "") == (got.Sentence() == "") || (got.ID != "") != (got.Task != nil && got.File != "") {
				t.Errorf("an answer's parts disagree: %+v %q", got, got.Sentence())
			}
			if doc := got.Object(); doc.String("id") != text(id) {
				t.Errorf("status's object %s", schema.Show(doc))
			}
		})
	}
}

// ActiveReasons is a copy of formats/README.md's table "Why there is none", their one home: the same codes, in the
// same order, with the same meanings.
func TestActiveReasonsAreTheREADME(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "formats", "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	text := strings.ReplaceAll(string(raw), "\r\n", "\n")
	start := strings.Index(text, "\n### Why there is none\n")
	if start < 0 {
		t.Fatal(`README.md has no "### Why there is none" section`)
	}
	section := text[start+1:]
	if end := strings.Index(section[4:], "\n#"); end >= 0 {
		section = section[:end+4]
	}
	var readme []string
	for _, m := range regexp.MustCompile("(?m)^\\| `([a-z0-9-]+)` \\| (.*) \\|$").FindAllStringSubmatch(section, -1) {
		readme = append(readme, m[1]+": "+m[2])
	}
	var ours []string
	for _, r := range ActiveReasons {
		ours = append(ours, r.Code+": "+r.Means)
		if r.Says == "" {
			t.Errorf("%s says nothing", r.Code)
		}
	}
	if strings.Join(ours, "\n") != strings.Join(readme, "\n") {
		t.Errorf("ActiveReasons:\n%s\nREADME.md:\n%s", strings.Join(ours, "\n"), strings.Join(readme, "\n"))
	}
	if got := (ActiveTask{Why: ReasonManyRunning, Detail: "T-0901, T-0902"}).Sentence(); got != "no active task: two or more tasks read running (T-0901, T-0902)" {
		t.Errorf("the sentence: %q", got)
	}
}

// The environment counts only for the session's own project, and a task folder that is not there gives none.
func TestActiveTaskEnvAndFolder(t *testing.T) {
	root := t.TempDir()
	task := "---\nformat: bonsai.task/1\nid: T-0901\ntitle: x\nstatus: todo\nlane: null\ndone_when: []\ndepends_on: []\n" +
		"blocked_by: null\ncreated: null\nstarted: null\nfinished: null\nlabels: {}\n---\n"
	if err := os.MkdirAll(filepath.Join(root, "tasks"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "tasks", "T-0901-x.md"), []byte(task), 0o644); err != nil {
		t.Fatal(err)
	}
	if a, _ := Active(ActiveInput{Main: root, TaskDir: "tasks", Env: "T-0901", EnvCounts: true}); a.ID != "T-0901" || a.How != HowNamed {
		t.Errorf("the session's own project: %+v", a)
	}
	if a, _ := Active(ActiveInput{Main: root, TaskDir: "tasks", Env: "T-0901"}); a.Why != ReasonNoneRunning {
		t.Errorf("another project: %+v", a)
	}
	if a, _ := Active(ActiveInput{Main: root, TaskDir: "tasks", Task: "T-0901", Env: "T-0902"}); a.ID != "T-0901" {
		t.Errorf("--task with another project's environment: %+v", a)
	}
	if a, err := Active(ActiveInput{Main: root, TaskDir: "nowhere"}); err != nil || a.Why != ReasonNoneRunning {
		t.Errorf("no task folder: %+v %v", a, err)
	}
}
