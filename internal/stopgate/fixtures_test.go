package stopgate

// The stop gate on contract §13's fixtures (formats/active-task, set 4's; answers.json's stop_gate section): every
// case gives its answer. Each case is copied into a temporary folder, each project made a git repository with every
// file committed and recorded on this machine in a home of the case's own (as bonsai init records a main checkout),
// and a worktree case's worktree added with git, as the README says. The gate runs for the case's session project
// with BONSAI_TASK set to the case's env:
//   - engages false: it lets the session end and writes no record anywhere;
//   - task blocks: it keeps the session going, its rule one of the task's, recorded in the session's main checkout;
//   - task ok: with no run report in the session's checkout it keeps the session going (stop-no-run-report); with
//     one and no result, naming the answer's ladder path (stop-no-result); given at that path a green local result
//     proving the session's HEAD (git rev-parse HEAD in its checkout) with the floor's rungs, it lets the session end
//     (stop-proven), and its record names that path. In a worktree, a commit after the result keeps it going again.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/testpack"
	"github.com/LastStep/Bonsai/internal/workspace"
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

func TestStopGateFixtures(t *testing.T) {
	tmp, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	testpack.Isolate(t, tmp)
	answers := readJSON(t, filepath.Join(fixtures, "answers.json"))
	section, _ := answers.Get("stop_gate")
	list, _ := section.([]any)
	if len(list) < 16 {
		t.Fatalf("answers.json's stop_gate section holds %d cases", len(list))
	}
	for _, a := range list {
		ans := a.(schema.Object)
		name := ans.String("case")
		t.Run(name, func(t *testing.T) {
			c := readJSON(t, filepath.Join(fixtures, name, "case.json"))
			dir := filepath.Join(tmp, name)
			copyTree(t, filepath.Join(fixtures, name), dir)
			home := filepath.Join(dir, "home")
			t.Setenv(workspace.HomeEnv, home)
			place := map[string]string{}
			for _, project := range []string{"main", "other"} {
				p := filepath.Join(dir, project)
				if _, err := os.Stat(p); err != nil {
					continue
				}
				testpack.Git(t, p, "init", "-q")
				testpack.Git(t, p, "add", "-A")
				testpack.Git(t, p, "commit", "-q", "-m", "the fixture")
				cfg, err := workspace.LoadConfig(p)
				if err != nil {
					t.Fatal(err)
				}
				if _, err := workspace.RecordCheckout(home, p, cfg.ID, time.Now()); err != nil {
					t.Fatal(err)
				}
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
			session := place[c.String("session")]
			named, _ := c.Get("env")
			task, _ := named.(string)
			getenv := env(session)
			if task != "" {
				getenv = env(session, workspace.TaskEnv, task)
			}
			run := func() (int, string) {
				t.Helper()
				return gate(t, getenv, strings.NewReader(stopPayload("s1", session)), untimed)
			}
			nothingRecorded := func() {
				t.Helper()
				for _, p := range place {
					noLog(t, p)
				}
			}
			main := place["main"]
			if c.String("layout") == "plain" {
				main = session
			}

			engages, _ := ans.Get("engages")
			if engages != true {
				if code, stderr := run(); code != 0 || stderr != "" {
					t.Errorf("does not engage, yet: exit %d %q", code, stderr)
				}
				nothingRecorded()
				return
			}
			if ans.String("task") == "blocks" {
				code, stderr := run()
				if code != 2 {
					t.Fatalf("a task that blocks: exit %d %q", code, stderr)
				}
				switch r := lastRecord(t, main); r.String("rule") {
				case RuleTaskMissing, RuleTaskTwice, RuleTaskBroken:
					if !strings.Contains(stderr, "BONSAI_TASK names "+task+", and ") || unsafeOut.MatchString(stderr) {
						t.Errorf("the reason: %q", stderr)
					}
				default:
					t.Errorf("rule %q, want one of the task's: %q", r.String("rule"), stderr)
				}
				return
			}
			if ans.String("task") != "ok" {
				t.Fatalf("an answer of task %q", ans.String("task"))
			}
			at := ans.String("ladder") // by path in the case folder
			rel := strings.TrimPrefix(at, "main/")
			if rel == at || !samePlace(place["main"], filepath.Join(dir, "main")) {
				t.Fatalf("the ladder path %q is not in the main checkout", at)
			}
			cfg, err := workspace.LoadConfigFull(main)
			if err != nil {
				t.Fatal(err)
			}
			code, stderr := run()
			blocks(t, code, stderr, main, RuleNoRunReport, "this checkout has no run report for "+task, "bonsai ladder --task "+task)
			// The run report goes in the session's own checkout (a worktree's own), uncommitted: the result says what
			// the tree was when the ladder ran, and the run report is read from the checkout as it is.
			write(t, session, cfg.Full.Documents.Run+"/R-2026-10-11-"+task+".md", "# R-2026-10-11-"+task+"\n")
			code, stderr = run()
			blocks(t, code, stderr, main, RuleNoResult, "the main checkout has no "+rel, "bonsai ladder --task "+task)

			head := testpack.Git(t, session, "rev-parse", "HEAD")
			l := greenResult(head)
			l.Task, l.Workspace = &task, cfg.ID
			l.Requested, l.Rungs = cfg.Full.LadderFloor, nil
			for _, n := range cfg.Full.LadderFloor {
				l.Rungs = append(l.Rungs, rung(n, "green"))
			}
			raw, err := l.Encode()
			if err != nil {
				t.Fatal(err)
			}
			write(t, dir, at, string(raw))
			code, stderr = run()
			if code != 0 || stderr != "" {
				t.Fatalf("a green result proving HEAD: exit %d %q", code, stderr)
			}
			if r := lastRecord(t, main); r.String("decision") != "allow" || r.String("rule") != RuleProven || r.String("target") != rel {
				t.Errorf("the allow's record: %s", schema.Show(r))
			}
			if c.String("layout") == "worktree" {
				noLog(t, session) // a verified worktree's records are its main checkout's
				// The same result at main's HEAD (another commit than the worktree's) does not prove the worktree.
				testpack.Git(t, session, "commit", "-q", "--allow-empty", "-m", "more")
				code, stderr = run()
				blocks(t, code, stderr, main, RuleOtherCommit, "but this checkout's HEAD is", "bonsai ladder --task "+task)

				// Seen as Claude Code's EnterWorktree leaves a session (CLAUDE_PROJECT_DIR the main checkout, the
				// payload's cwd the worktree), the gate judges the worktree all the same, even with a run report in
				// the main checkout and the result proving the main checkout's HEAD; a result at the worktree's HEAD
				// lets it end.
				write(t, main, cfg.Full.Documents.Run+"/R-2026-10-11-"+task+".md", "# R-2026-10-11-"+task+"\n")
				entered := func() (int, string) {
					t.Helper()
					return gate(t, env(main, workspace.TaskEnv, task), strings.NewReader(stopPayload("s1", session)), untimed)
				}
				code, stderr = entered()
				blocks(t, code, stderr, main, RuleOtherCommit, "but this checkout's HEAD is", "bonsai ladder --task "+task)
				l.Git.SHA = testpack.Git(t, session, "rev-parse", "HEAD")
				if raw, err = l.Encode(); err != nil {
					t.Fatal(err)
				}
				write(t, dir, at, string(raw))
				code, stderr = entered()
				if code != 0 || stderr != "" {
					t.Fatalf("entered, a green result proving the worktree's HEAD: exit %d %q", code, stderr)
				}
				if r := lastRecord(t, main); r.String("rule") != RuleProven || r.String("checkout") != "wt" || r.String("branch") != "task" {
					t.Errorf("entered, the allow's record: %s", schema.Show(r))
				}
				noLog(t, session)

				// A worktree with no bonsai.yaml of its own (its branch made before the project was linked, or
				// bonsai.yaml never committed) is the worktree all the same, by git's own link: kept going after a
				// commit there, and allowed on a result proving its new HEAD.
				testpack.Git(t, session, "rm", "-q", "bonsai.yaml")
				testpack.Git(t, session, "commit", "-q", "-m", "no bonsai.yaml")
				code, stderr = entered()
				blocks(t, code, stderr, main, RuleOtherCommit, "but this checkout's HEAD is", "bonsai ladder --task "+task)
				l.Git.SHA = testpack.Git(t, session, "rev-parse", "HEAD")
				if raw, err = l.Encode(); err != nil {
					t.Fatal(err)
				}
				write(t, dir, at, string(raw))
				code, stderr = entered()
				if code != 0 || stderr != "" {
					t.Fatalf("entered, no bonsai.yaml in the worktree, a green result proving its HEAD: exit %d %q", code, stderr)
				}
				if r := lastRecord(t, main); r.String("rule") != RuleProven || r.String("checkout") != "wt" {
					t.Errorf("entered, no bonsai.yaml in the worktree, the allow's record: %s", schema.Show(r))
				}
			}
		})
	}
}

func samePlace(a, b string) bool {
	return strings.EqualFold(filepath.Clean(a), filepath.Clean(b)) && (filepath.Separator == '\\' || filepath.Clean(a) == filepath.Clean(b))
}

// The fixtures' tasks name no bonsai.ladder, so the floor alone is what a result must hold; a format-1 task's
// numbers read as rung numbers (json.Number).
func TestRungsWanted(t *testing.T) {
	task := &format.Task{Labels: schema.Object{{Key: "bonsai.ladder", Value: []any{json.Number("3"), json.Number("1")}}}}
	if got, err := rungsWanted([]int64{1, 0}, task); err != nil || rungList(got) != "0, 1, 3" {
		t.Errorf("rungsWanted %v (%v)", got, err)
	}
	for _, bad := range []any{"2", []any{json.Number("-1")}, []any{json.Number("1.5")}, []any{"x"}, []any{true}} {
		task.Labels = schema.Object{{Key: "bonsai.ladder", Value: bad}}
		if got, err := rungsWanted(nil, task); err == nil {
			t.Errorf("bonsai.ladder %v read as %v", bad, got)
		}
	}
	task.Labels = schema.Object{{Key: "bonsai.ladder", Value: nil}}
	if got, err := rungsWanted([]int64{0}, task); err != nil || rungList(got) != "0" {
		t.Errorf("a null bonsai.ladder: %v (%v)", got, err)
	}
	if _, err := rungsWanted([]int64{-1}, nil); err == nil {
		t.Error("a floor of -1 read")
	}
}
