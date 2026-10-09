package formats

// Contract §13's fixtures (README.md, "The active-task fixtures"): one folder per case under active-task/, and
// answers.json with one section per reader of the active task. Step 5.1.5 builds the active-task function and tests
// it on the function's section; 5.3 tests the guard's and the stop gate's, 5.4 rung 0's. This test holds the
// fixtures and the answers to their documented shape, and to each other: every case has a case.json and the folders
// it names, no case holds a .git, every section answers every case once in the folders' order, every answer's
// fields have their form, and every reason the function gives is in README.md's table (their one home), each used.

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/schema"
)

const fixturesDir = "active-task"

// The sections of answers.json, in order: the readers of contract §13.
var answerSections = []string{"function", "guard", "rung0", "stop_gate"}

var taskID = regexp.MustCompile(`^T-[0-9]{4,6}$`)

// fixtureCases lists the case folders, sorted by name.
func fixtureCases(t *testing.T) []string {
	t.Helper()
	entries, err := os.ReadDir(fixturesDir)
	if err != nil {
		t.Fatalf("%s: %v", fixturesDir, err)
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() {
			out = append(out, e.Name())
		}
	}
	sort.Strings(out)
	if len(out) == 0 {
		t.Fatalf("%s holds no case", fixturesDir)
	}
	return out
}

// activeReasons reads the reasons the active-task function gives for none, from README.md's table under "### Why
// there is none": their one home.
func activeReasons(t *testing.T) map[string]bool {
	t.Helper()
	raw, err := os.ReadFile("README.md")
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
	codes := map[string]bool{}
	for _, m := range regexp.MustCompile("(?m)^\\| `([a-z0-9-]+)` \\|").FindAllStringSubmatch(section, -1) {
		codes[m[1]] = true
	}
	if len(codes) == 0 {
		t.Fatal("README.md's table of reasons is empty")
	}
	return codes
}

// fields holds an answer or a case.json to exactly these fields, in this order.
func fields(t *testing.T, o schema.Object, where string, want ...string) {
	t.Helper()
	if got := strings.Join(o.Keys(), " "); got != strings.Join(want, " ") {
		t.Errorf("%s: fields %q, want %q", where, got, strings.Join(want, " "))
	}
}

// idOrNull checks a value is null or a task id, and returns it ("" for null).
func idOrNull(t *testing.T, v any, where string) string {
	t.Helper()
	if v == nil {
		return ""
	}
	s, ok := v.(string)
	if !ok || !taskID.MatchString(s) {
		t.Errorf("%s: %s is neither null nor a task id", where, schema.Show(v))
		return ""
	}
	return s
}

func TestFixturesAndAnswers(t *testing.T) {
	cases := fixtureCases(t)
	// No case holds a .git: a case that needs git is a layout a reader's test builds.
	err := filepath.WalkDir(fixturesDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Name() == ".git" {
			t.Errorf("%s: a fixture holds no .git (a reader's test builds the repository)", filepath.ToSlash(p))
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	// Each case.json, and the folders it names.
	inputs := map[string]schema.Object{}
	for _, c := range cases {
		where := fixturesDir + "/" + c + "/case.json"
		o := mustObject(t, readJSON(t, where), where)
		inputs[c] = o
		fields(t, o, where, "about", "layout", "checkout", "session", "task", "env", "does_not_parse")
		if o.String("about") == "" {
			t.Errorf("%s: no about", where)
		}
		layout := o.String("layout")
		if layout != "plain" && layout != "worktree" {
			t.Errorf("%s: layout %q, want plain or worktree", where, layout)
		}
		for _, k := range []string{"checkout", "session"} {
			folder := o.String(k)
			switch {
			case folder == "worktree" && layout != "worktree":
				t.Errorf("%s: %s is worktree in a plain layout", where, k)
			case folder != "main" && folder != "other" && folder != "worktree":
				t.Errorf("%s: %s %q, want main, other or worktree", where, k, folder)
			case folder != "worktree":
				if _, err := os.Stat(filepath.Join(fixturesDir, c, folder, "bonsai.yaml")); err != nil {
					t.Errorf("%s: its %s, %s, holds no bonsai.yaml", where, k, folder)
				}
			}
		}
		if _, err := os.Stat(filepath.Join(fixturesDir, c, "main", "bonsai.yaml")); err != nil {
			t.Errorf("%s: the case has no main/bonsai.yaml", where)
		}
		for _, k := range []string{"task", "env"} {
			v, _ := o.Get(k)
			idOrNull(t, v, where+" "+k)
		}
		list, _ := o.Get("does_not_parse")
		files, ok := list.([]any)
		if !ok {
			t.Errorf("%s: does_not_parse is not a list", where)
		}
		for _, f := range files {
			p, _ := f.(string)
			if _, err := os.Stat(filepath.Join(fixturesDir, c, filepath.FromSlash(p))); err != nil || p == "" {
				t.Errorf("%s: does_not_parse names %s, which is not in the case", where, schema.Show(f))
			}
		}
	}
	// answers.json: four sections, each answering every case once, in the folders' order.
	answers := mustObject(t, readJSON(t, fixturesDir+"/answers.json"), "answers.json")
	fields(t, answers, "answers.json", append([]string{"about"}, answerSections...)...)
	reasons := activeReasons(t)
	// how's values have their one home in status --json's active_task (contract §12, §13).
	howEnum, _ := schemaAt(t, loadSchema(t, "status"), "status", "properties", "active_task", "properties", "how").Get("enum")
	used := map[string]bool{}
	function := map[string]schema.Object{}
	for _, section := range answerSections {
		v, _ := answers.Get(section)
		list, ok := v.([]any)
		if !ok {
			t.Errorf("answers.json: %s is not a list", section)
			continue
		}
		var named []string
		for i, a := range list {
			where := fmt.Sprintf("answers.json %s[%d]", section, i)
			o := mustObject(t, a, where)
			c := o.String("case")
			named = append(named, c)
			where += " (" + c + ")"
			switch section {
			case "function":
				fields(t, o, where, "case", "id", "how", "why")
				function[c] = o
				idV, _ := o.Get("id")
				id := idOrNull(t, idV, where+" id")
				how, _ := o.Get("how")
				why, _ := o.Get("why")
				inEnum := false
				hows, _ := howEnum.([]any)
				for _, e := range hows {
					inEnum = inEnum || schema.Equal(e, how)
				}
				switch {
				case !inEnum:
					t.Errorf("%s: how %s is not one of status's active_task.how %s", where, schema.Show(how), schema.Show(howEnum))
				case id != "" && (how == nil || why != nil):
					t.Errorf("%s: an active task has a how and why null", where)
				case id == "" && how != nil:
					t.Errorf("%s: no active task has how null", where)
				case id == "":
					w, _ := why.(string)
					if !reasons[w] {
						t.Errorf("%s: why %s is not in README.md's table of reasons", where, schema.Show(why))
					}
					used[w] = true
				}
			case "guard", "rung0":
				fields(t, o, where, "case", "grants")
				g, _ := o.Get("grants")
				idOrNull(t, g, where+" grants")
			case "stop_gate":
				fields(t, o, where, "case", "engages", "task", "ladder")
				engages, _ := o.Get("engages")
				task, _ := o.Get("task")
				ladder, _ := o.Get("ladder")
				lp, _ := ladder.(string)
				switch {
				case engages == false && (task != nil || ladder != nil):
					t.Errorf("%s: a stop gate that does not engage judges no task and reads no ladder result", where)
				case engages == true && task == "blocks" && ladder != nil:
					t.Errorf("%s: a stop gate that blocks on the task reads no ladder result", where)
				case engages == true && task == "ok" && !regexp.MustCompile(`^main/\.bonsai/local/ladder/T-[0-9]{4,6}\.json$`).MatchString(lp):
					t.Errorf("%s: ladder %s is not the main checkout's .bonsai/local/ladder/<task id>.json", where, schema.Show(ladder))
				case engages != true && engages != false, engages == true && task != "ok" && task != "blocks":
					t.Errorf("%s: engages %s, task %s", where, schema.Show(engages), schema.Show(task))
				}
			}
		}
		if strings.Join(named, " ") != strings.Join(cases, " ") {
			t.Errorf("answers.json: %s answers\n  %v\nwant every case folder once, in order\n  %v", section, named, cases)
		}
	}
	for _, code := range sortedKeys(reasons) {
		if !used[code] {
			t.Errorf("README.md lists the reason %s, which no case's function answer gives", code)
		}
	}
	// Rung 0 judges against the named task's grants: its answer is the function's id when the function names it.
	rung0, _ := answers.Get("rung0")
	if list, ok := rung0.([]any); ok {
		for _, a := range list {
			o, _ := a.(schema.Object)
			f := function[o.String("case")]
			g, _ := o.Get("grants")
			want, _ := f.Get("id")
			if f.String("how") != "named" {
				want = nil
			}
			if f != nil && !schema.Equal(g, want) {
				t.Errorf("answers.json rung0 (%s): grants %s, want the function's named task %s", o.String("case"),
					schema.Show(g), schema.Show(want))
			}
		}
	}
}
