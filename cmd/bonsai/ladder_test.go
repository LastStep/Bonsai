package main

// bonsai ladder from the command line (plan-5 5.4.2): a climb's --json fits bonsai.climb/1 and its text names each
// rung; every refusal's next step is a command an agent runs as written (agents first, note 12); a stamped test build
// climbs only inside its scratch folder (note 10); bonsai --help --json lists the word. And check's rows for the
// ladder's shape (note 11), one case per word of format.LadderCheckWords, walked as TestCheckTable walks CheckWords.

import (
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/workspace"
)

const climbID = "ws-bbbbbbbbbbbbbbbbbbbbbbbbbb"

// climbYAML is a bonsai.yaml with the floor and the ladder given (each rung a YAML block of its own fields), every other
// field written out.
func climbYAML(floor, ladder, ratchets string) string {
	if ratchets == "" {
		ratchets = "{}"
	}
	gen := ""
	for _, k := range []string{"log", "asks", "ladder", "run", "sessions"} {
		gen += "  " + k + ":\n    keep_days: null\n    keep_newest: null\n"
	}
	return "format: bonsai.workspace/1\nid: " + climbID + "\nname: climb\npacks: []\ndocuments:\n  task: work/tasks\n  run: work/runs\n" +
		"  answers: work/answers.md\n  memory: work/memory\n  protocols: work/protocols\nprotected: []\nperson_only: []\nnever_edit: []\n" +
		"ladder_floor: " + floor + "\nladder:" + ladder + "\nratchets: " + ratchets + "\nci_marked_tests: []\ngenerated:\n" + gen
}

// rungYAML is one rung of climbYAML's ladder.
func rungYAML(n int, name, command, ratchet, capture, tests string) string {
	null := func(s string) string {
		if s == "" {
			return "null"
		}
		return s
	}
	cmd := "null"
	kind := "guard"
	if command != "" {
		cmd, kind = strconv.Quote(command), "command"
	}
	return "\n  - rung: " + strconv.Itoa(n) + "\n    name: " + name + "\n    kind: " + kind + "\n    command: " + cmd + "\n    required: true" +
		"\n    timeout_s: null\n    ratchet: " + null(ratchet) + "\n    capture: " + null(capture) + "\n    means: null\n    tests: " +
		null(tests) + "\n    base_setup: null"
}

// climbTask is a running format-1 task.
func climbTask(id string) string {
	return "---\nformat: bonsai.task/1\nid: " + id + "\ntitle: A task\nstatus: running\nlane: null\ndone_when:\n  - \"x\"\n" +
		"depends_on: []\nblocked_by:\ncreated: 2026-10-13\nstarted:\nfinished:\nlabels: {}\n---\n\n# A task\n"
}

// climbable makes c's project a linked, committed checkout with one running task, T-0901, and this machine's record of it.
func (c *cli) climbable(yaml string) {
	c.t.Helper()
	c.write("bonsai.yaml", yaml)
	c.write(workspace.GitignoreFile, workspace.GitignoreText)
	c.write("work/tasks/T-0901-a.md", climbTask("T-0901"))
	c.commit("climbable")
	if _, err := workspace.RecordCheckout(os.Getenv(workspace.HomeEnv), c.root, climbID, time.Now()); err != nil {
		c.t.Fatal(err)
	}
}

// A green climb and a red one from the command line: --json fits bonsai.climb/1 with the result inside, the text names
// each rung and the result, and the exit code is the colour.
func TestLadderWord(t *testing.T) {
	c := newCLI(t)
	t.Setenv(workspace.TaskEnv, "")
	t.Setenv("BONSAI_LADDER", "")
	c.climbable(climbYAML("[1]", rungYAML(1, "one", "echo one", "", "", "")+rungYAML(2, "two", "exit 2", "", "", ""), ""))
	code, out, errOut := c.run("", "ladder", "--task", "T-0901", "--json")
	doc := fits(t, out, "climb")
	if code != 0 || errOut != "" || doc.String("result") != ".bonsai/local/ladder/T-0901.json" || doc.String("task") != "T-0901" {
		t.Fatalf("exit %d, %s\n%s", code, out, errOut)
	}
	l, _ := doc.Get("ladder")
	if lo := l.(schema.Object); lo.String("mode") != "local" || schema.Show(must(lo.Get("green"))) != "true" {
		t.Errorf("the result inside: %s", schema.Show(lo))
	}
	code, out, _ = c.run("", "ladder", "--task", "T-0901")
	if code != 0 || !strings.Contains(out, "bonsai ladder: climb, T-0901 (local), rungs 1") || !strings.Contains(out, "GREEN: 1 of 1") {
		t.Errorf("the text: exit %d\n%s", code, out)
	}
	// A task whose bonsai.ladder adds the red rung: exit 1, next names its command and the climb again.
	c.write("work/tasks/T-0902-b.md", strings.Replace(climbTask("T-0902"), "labels: {}", "labels:\n  bonsai.ladder: [2]", 1))
	c.write("work/tasks/T-0901-a.md", strings.Replace(climbTask("T-0901"), "status: running", "status: todo", 1))
	c.commit("a red task")
	code, out, _ = c.run("", "ladder", "--task", "T-0902", "--json")
	doc = fits(t, out, "climb")
	next, _ := doc.Get("next")
	if code != 1 || next == nil || next.(schema.Object).String("do") != "fix the two rung (rung 2): its reason shows what failed; run: exit 2 ; "+
		"then run: bonsai ladder --task T-0902 --json" {
		t.Errorf("the red climb: exit %d, next %s", code, schema.Show(next))
	}
	// Its last step, the climb again, is a command Bonsai has, with its flags; the first is the rung's own command.
	do := next.(schema.Object).String("do")
	if err := runnable(strings.TrimPrefix(do[strings.LastIndex(do, "run: "):], "run: ")); err != nil {
		t.Errorf("the climb again in %q: %v", do, err)
	}
}

func must(v any, _ bool) any { return v }

// Every refusal of bonsai ladder exits with its code and an error whose next step names only commands Bonsai has,
// with its flags (checkNext), and runs as written where it names no placeholder.
func TestLadderRefusalsNameRunnableSteps(t *testing.T) {
	c := newCLI(t)
	t.Setenv(workspace.TaskEnv, "")
	t.Setenv("BONSAI_LADDER", "")
	c.climbable(climbYAML("[1, 7]", rungYAML(1, "one", "echo one", "", "", ""), ""))
	cases := []struct {
		args []string
		exit int
		word string
	}{
		{[]string{"ladder", "--json"}, 2, "no-task"},
		{[]string{"ladder", "--task", "T-0999", "--json"}, 2, "no-task"},
		{[]string{"ladder", "--task", "T-0901", "--json"}, 2, "rung-not-defined"},
		{[]string{"ladder", "--nope", "--json"}, 2, "bad-flag"},
		{[]string{"ladder", "now", "--json"}, 2, "bad-flag"},
	}
	for _, k := range cases {
		code, out, _ := c.run("", k.args...)
		e := errorIn(fits(t, out, "climb"), "climb")
		if code != k.exit || e.String("code") != k.word {
			t.Errorf("%v: exit %d, %s", k.args, code, schema.Show(e))
			continue
		}
		n, _ := e.Get("next")
		checkNext(t, k.word, n.(schema.Object).String("do"))
	}
	// The one task running is named in the next step, as a command that runs as written: it is refused only for its
	// floor's rung 7, which the next step names.
	code, out, _ := c.run("", "ladder", "--json")
	n, _ := errorIn(fits(t, out, "climb"), "climb").Get("next")
	do := n.(schema.Object).String("do")
	if code != 2 || do != "run: bonsai ladder --task T-0901 --json" {
		t.Fatalf("no task named: %d %q", code, do)
	}
	args := shellSplit(strings.TrimPrefix(do, "run: bonsai "))
	if code, out, _ := c.run("", args...); code != 2 || errorIn(fits(t, out, "climb"), "climb").String("code") != "rung-not-defined" {
		t.Errorf("the next step run as written: exit %d\n%s", code, out)
	}
	// Main not verified: a person's step, exit 4.
	if err := os.RemoveAll(filepath.Join(os.Getenv(workspace.HomeEnv), "workspaces")); err != nil {
		t.Fatal(err)
	}
	code, out, _ = c.run("", "ladder", "--task", "T-0901", "--json")
	if e := errorIn(fits(t, out, "climb"), "climb"); code != 4 || e.String("code") != "main-not-verified" || whoOf(e) != "person" {
		t.Errorf("main not verified: exit %d, %s", code, schema.Show(e))
	}
}

// A stamped test build refuses to climb outside its scratch folder (note 10: a test build never makes a proof in a
// real project), naming the installed bonsai's command; inside it, it climbs.
func TestLadderStampedBuild(t *testing.T) {
	c := newCLI(t)
	t.Setenv(workspace.TaskEnv, "")
	t.Setenv("BONSAI_LADDER", "")
	c.climbable(climbYAML("[1]", rungYAML(1, "one", "echo one", "", "", ""), ""))
	scratch := filepath.Join(c.tmp, "scratch")
	exe := "bonsai"
	if runtime.GOOS == "windows" {
		exe = "bonsai.exe"
	}
	place := filepath.Join(c.tmp, "stamped", exe)
	stampedBuild(t, place, scratch)
	code, out, _ := runBin(t, place, c.root, nil, "ladder", "--task", "T-0901", "--json")
	e := errorIn(fits(t, out, "climb"), "climb")
	n, _ := e.Get("next")
	installed := "/usr/local/bin/bonsai"
	if runtime.GOOS == "windows" {
		installed = `"C:/Program Files/Bonsai/bonsai.exe"`
	}
	if code != 4 || e.String("code") != "test-build" || !strings.HasPrefix(n.(schema.Object).String("do"), "run: "+installed+" ladder --task T-0901 --json") {
		t.Errorf("outside its scratch folder: exit %d, %s", code, schema.Show(e))
	}
	if _, err := os.Stat(filepath.Join(c.root, ".bonsai", "local", "ladder")); err == nil {
		t.Error("a stamped build wrote a result outside its scratch folder")
	}
	// Inside it: a project of its own, recorded on this machine, climbs green.
	inside := filepath.Join(scratch, "inside")
	copyTree(t, c.root, inside)
	if _, err := workspace.RecordCheckout(os.Getenv(workspace.HomeEnv), inside, climbID, time.Now().Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(filepath.Join(os.Getenv(workspace.HomeEnv), "workspaces", mustKey(t, c.root))); err != nil {
		t.Fatal(err)
	}
	if code, out, errOut := runBin(t, place, inside, nil, "ladder", "--task", "T-0901", "--json"); code != 0 {
		t.Errorf("inside its scratch folder: exit %d\n%s%s", code, out, errOut)
	}
}

func mustKey(t *testing.T, root string) string {
	t.Helper()
	k, err := workspace.MachineKey(root)
	if err != nil {
		t.Fatal(err)
	}
	return k
}

// bonsai --help --json lists ladder with its flags and exit codes.
func TestHelpJSONListsLadder(t *testing.T) {
	h, _ := helpOut(t, "--help", "--json")
	for _, w := range h.Words {
		if w.Name != "ladder" {
			continue
		}
		var flags []string
		for _, f := range w.Flags {
			flags = append(flags, f.Name)
		}
		if strings.Join(flags, " ") != "--task --root --ci --json" || len(w.Exits) != 5 {
			t.Errorf("ladder in the help: %v, %d exits", flags, len(w.Exits))
		}
		return
	}
	t.Error("bonsai --help --json does not list ladder")
}

// ladderCheckCases make a project where check finds each word of format.LadderCheckWords.
var ladderCheckCases = map[string]string{
	"rung-twice":       climbYAML("[1]", rungYAML(1, "one", "echo one", "", "", "")+rungYAML(1, "again", "echo again", "", "", ""), ""),
	"floor-no-rung":    climbYAML("[1, 4]", rungYAML(1, "one", "echo one", "", "", ""), ""),
	"ratchet-no-rung":  climbYAML("[1]", rungYAML(1, "one", "echo one", "", "", ""), "\n  tests: 4"),
	"capture-pattern":  climbYAML("[1]", rungYAML(1, "one", "echo one", "", "\n      found: \"(\"", ""), ""),
	"ratchet-no-count": climbYAML("[1]", rungYAML(1, "one", "echo one", "tests", "", ""), ""),
	"rung-bash":        climbYAML("[1]", rungYAML(1, "one", "bash -c true", "", "", ""), ""),
	"tests-unknown":    climbYAML("[1]", rungYAML(1, "one", "echo one", "", "", "nunit"), ""),
}

// Check's rows for the ladder's shape, walked from their table: each word has a case, in its own list (a finding, or
// a warning that never changes the exit code), the output fits bonsai.check/1, and each next step names only commands
// Bonsai has; a ladder with none of them is clean of them.
func TestLadderCheckTable(t *testing.T) {
	for code := range ladderCheckCases {
		if w, ok := format.CheckWord(code); !ok || !inLadderWords(code) {
			t.Errorf("a case for %s (%v), which is not in format.LadderCheckWords", code, w)
		}
	}
	for _, w := range format.LadderCheckWords {
		t.Run(w.Word, func(t *testing.T) {
			yaml, ok := ladderCheckCases[w.Word]
			if !ok {
				t.Fatalf("format.LadderCheckWords has %s, and TestLadderCheckTable no case for it", w.Word)
			}
			c := newCLI(t)
			c.write("bonsai.yaml", yaml)
			code, out, _ := c.run("", "check", "--json")
			doc := fits(t, out, "check")
			findings, _ := doc.Get("findings")
			warnings, _ := doc.Get("warnings")
			mine, others := findings.([]any), warnings.([]any)
			if w.Kind == "warning" {
				mine, others = others, mine
			}
			if !hasCode(mine, w.Word) || hasCode(others, w.Word) {
				t.Fatalf("check --json has no %s %s, or has it in the other list:\n%s", w.Kind, w.Word, out)
			}
			if want := map[bool]int{true: 1, false: 0}[len(findings.([]any)) > 0]; code != want {
				t.Errorf("exit %d with %d findings", code, len(findings.([]any)))
			}
			for _, list := range []any{findings, warnings} {
				for _, f := range list.([]any) {
					fo := f.(schema.Object)
					n, _ := fo.Get("next")
					checkNext(t, fo.String("code"), n.(schema.Object).String("do"))
					if inLadderWords(fo.String("code")) && (fo.String("file") != "bonsai.yaml" || whoOf(fo) != "person") {
						t.Errorf("%s: file %s, who %s", fo.String("code"), fo.String("file"), whoOf(fo))
					}
				}
			}
		})
	}
	// A ladder of every form check knows has none of them.
	c := newCLI(t)
	c.write("bonsai.yaml", climbYAML("[1, 2]", rungYAML(1, "one", "echo 3 passed", "found", "\n      found: \"([0-9]+) passed\"", "")+
		rungYAML(2, "two", "go test -json ./...", "go_tests", "", "go-test-json"), "\n  found: 2\n  go_tests: 1"))
	_, out, _ := c.run("", "check", "--json")
	doc := fits(t, out, "check")
	for _, list := range []string{"findings", "warnings"} {
		v, _ := doc.Get(list)
		for _, f := range v.([]any) {
			if code := f.(schema.Object).String("code"); inLadderWords(code) {
				t.Errorf("a sound ladder has %s:\n%s", code, out)
			}
		}
	}
}

func inLadderWords(code string) bool {
	for _, w := range format.LadderCheckWords {
		if w.Word == code {
			return true
		}
	}
	return false
}
