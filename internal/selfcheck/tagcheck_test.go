package selfcheck

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

// The release workflow's tag check (.github/scripts/release-tag-check.sh), run through sh on a saved answer of GitHub's
// check-runs endpoint (testdata/check-runs-<commit>.json: a real commit on main whose every check passed, saved with
// gh api). Each case changes a copy of that answer in one way and expects the script to pass or to fail naming what
// changed. With --answer the script asks GitHub nothing.

const tagCheckScript = ".github/scripts/release-tag-check.sh"

// tagCheckRepo makes a git repository with one commit that origin/main holds and one on a side branch that it does
// not, and returns the folder and the two commits.
func tagCheckRepo(t *testing.T) (dir, onMain, offMain string) {
	t.Helper()
	dir = t.TempDir()
	home := t.TempDir()
	git := func(args ...string) string {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(), "HOME="+home, "XDG_CONFIG_HOME="+home, "GIT_CONFIG_NOSYSTEM=1",
			"GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@example.invalid",
			"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@example.invalid")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}
	git("init", "-q", "-b", "main")
	git("commit", "-q", "--allow-empty", "-m", "on main")
	onMain = git("rev-parse", "HEAD")
	git("update-ref", "refs/remotes/origin/main", onMain)
	git("checkout", "-q", "-b", "side")
	git("commit", "-q", "--allow-empty", "-m", "on a side branch")
	offMain = git("rev-parse", "HEAD")
	return dir, onMain, offMain
}

// savedAnswer reads the one saved answer in testdata and returns it decoded, with the ids of its check suites.
func savedAnswer(t *testing.T) (map[string]any, []float64) {
	t.Helper()
	files, err := filepath.Glob(filepath.Join("testdata", "check-runs-*.json"))
	if err != nil || len(files) != 1 {
		t.Fatalf("want exactly one testdata/check-runs-*.json, found %v (%v)", files, err)
	}
	b, err := os.ReadFile(files[0])
	if err != nil {
		t.Fatal(err)
	}
	var answer map[string]any
	if err := json.Unmarshal(b, &answer); err != nil {
		t.Fatalf("%s: %v", files[0], err)
	}
	var suites []float64
	for _, r := range checkRuns(t, answer) {
		id := r["check_suite"].(map[string]any)["id"].(float64)
		seen := false
		for _, s := range suites {
			seen = seen || s == id
		}
		if !seen {
			suites = append(suites, id)
		}
	}
	return answer, suites
}

func checkRuns(t *testing.T, answer map[string]any) []map[string]any {
	t.Helper()
	list, ok := answer["check_runs"].([]any)
	if !ok || len(list) == 0 {
		t.Fatal("the saved answer has no check_runs")
	}
	out := make([]map[string]any, len(list))
	for i, r := range list {
		out[i] = r.(map[string]any)
	}
	return out
}

// copyAnswer returns a deep copy of the answer, so each case changes its own.
func copyAnswer(t *testing.T, answer map[string]any) map[string]any {
	t.Helper()
	b, err := json.Marshal(answer)
	if err != nil {
		t.Fatal(err)
	}
	var out map[string]any
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	return out
}

// withRun returns a copy of the answer with one more check run: a copy of its first, renamed, in the suite given, with
// the status and conclusion given (conclusion "" for none).
func withRun(t *testing.T, answer map[string]any, name string, suite float64, status, conclusion string) map[string]any {
	t.Helper()
	out := copyAnswer(t, answer)
	runs := out["check_runs"].([]any)
	run := copyAnswer(t, runs[0].(map[string]any))
	run["name"] = name
	run["status"] = status
	if conclusion == "" {
		run["conclusion"] = nil
	} else {
		run["conclusion"] = conclusion
	}
	run["check_suite"] = map[string]any{"id": suite}
	out["check_runs"] = append(runs, run)
	out["total_count"] = float64(len(runs) + 1)
	return out
}

// without returns a copy of the answer with every check run of that name taken out.
func without(t *testing.T, answer map[string]any, name string) map[string]any {
	t.Helper()
	out := copyAnswer(t, answer)
	var kept []any
	for _, r := range out["check_runs"].([]any) {
		if r.(map[string]any)["name"] != name {
			kept = append(kept, r)
		}
	}
	out["check_runs"] = kept
	out["total_count"] = float64(len(kept))
	return out
}

// runTagCheck writes the answers one after another (as gh api --paginate writes pages) and runs the script on them.
func runTagCheck(t *testing.T, dir, tag, sha string, own string, answers ...map[string]any) (int, string) {
	t.Helper()
	var b []byte
	for _, a := range answers {
		j, err := json.Marshal(a)
		if err != nil {
			t.Fatal(err)
		}
		b = append(append(b, j...), '\n')
	}
	return runTagCheckRaw(t, dir, tag, sha, own, b)
}

// runTagCheckRaw runs the script on a saved answer of exactly these bytes.
func runTagCheckRaw(t *testing.T, dir, tag, sha string, own string, answer []byte) (int, string) {
	t.Helper()
	file := filepath.Join(t.TempDir(), "answer.json")
	if err := os.WriteFile(file, answer, 0o644); err != nil {
		t.Fatal(err)
	}
	script, err := filepath.Abs(filepath.Join("..", "..", filepath.FromSlash(tagCheckScript)))
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("sh", script, "--answer", file, "--own-suite", own)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GITHUB_REF_NAME="+tag, "GITHUB_SHA="+sha, "GITHUB_REPOSITORY=LastStep/Bonsai",
		"GITHUB_RUN_ID=", "GH_TOKEN=")
	out, err := cmd.CombinedOutput()
	code := 0
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		code = exit.ExitCode()
	} else if err != nil {
		t.Fatalf("sh %s: %v", tagCheckScript, err)
	}
	return code, string(out)
}

func needTagCheckTools(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("the tag check runs in the release workflow on GitHub's ubuntu-latest runner, through sh with git and " +
			"jq; Windows has no sh, so it is tested on Linux only")
	}
	for _, tool := range []string{"sh", "git", "jq"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Fatalf("the tag check's test needs %s on the PATH, as the release workflow's runner has it", tool)
		}
	}
}

// ownSuite is a check suite id no saved check run carries: the release run's own suite when no case adds a run to it.
const ownSuite = "1"

// TestTagCheckOnSavedAnswer runs the script on the saved answer and on changed copies of it.
func TestTagCheckOnSavedAnswer(t *testing.T) {
	needTagCheckTools(t)
	dir, onMain, offMain := tagCheckRepo(t)
	answer, suites := savedAnswer(t)
	const own = 7 // the release run's own suite, in the case that adds a run to it

	cases := []struct {
		name    string
		tag     string
		sha     string
		answers []map[string]any
		code    int
		want    string // in the output
		ownRun  bool   // the added run is in the release run's own suite
	}{
		{"the saved answer passes", "v1.0.0", onMain, []map[string]any{answer}, 0, "every named job completed with success", false},
		{"a pre-release tag passes", "v1.0.0-rc.1", onMain, []map[string]any{answer}, 0, "every named job completed with success", false},
		{"v1.0 fails", "v1.0", onMain, []map[string]any{answer}, 1, "tag v1.0 is not vX.Y.Z", false},
		{"v1.0.0-beta fails", "v1.0.0-beta", onMain, []map[string]any{answer}, 1, "tag v1.0.0-beta is not vX.Y.Z", false},
		{"base-v1.0.0 fails", "base-v1.0.0", onMain, []map[string]any{answer}, 1, "tag base-v1.0.0 is not vX.Y.Z", false},
		{"a commit off main fails", "v1.0.0", offMain, []map[string]any{answer}, 1, "is not on main", false},
		{"a run in progress in the own suite is left out", "v1.0.0", onMain,
			[]map[string]any{withRun(t, answer, "build", own, "in_progress", "")}, 0, "every named job completed with success", true},
		{"a named job's name in progress in the own suite is left out", "v1.0.0", onMain,
			[]map[string]any{withRun(t, answer, "windows", own, "in_progress", "")}, 0, "every named job completed with success", true},
		{"a named job in progress fails", "v1.0.0", onMain,
			[]map[string]any{withRun(t, answer, "windows", suites[0], "in_progress", "")}, 1, "windows: in_progress", false},
		{"a named job failed fails", "v1.0.0", onMain,
			[]map[string]any{withRun(t, answer, "lint", suites[0], "completed", "failure")}, 1, "lint: completed, failure", false},
		{"a named job skipped fails", "v1.0.0", onMain,
			[]map[string]any{withRun(t, answer, "test", suites[0], "completed", "skipped")}, 1, "test: completed, skipped", false},
		{"a named job missing fails", "v1.0.0", onMain,
			[]map[string]any{without(t, answer, "release-check")}, 1, "release-check: no check run on this commit", false},
		{"another check failed fails", "v1.0.0", onMain,
			[]map[string]any{withRun(t, answer, "something else", suites[0], "completed", "failure")}, 1, "something else: failure", false},
		{"another check cancelled fails", "v1.0.0", onMain,
			[]map[string]any{withRun(t, answer, "something else", suites[0], "completed", "cancelled")}, 1, "something else: cancelled", false},
		{"another check timed out fails", "v1.0.0", onMain,
			[]map[string]any{withRun(t, answer, "something else", suites[0], "completed", "timed_out")}, 1, "something else: timed_out", false},
		{"another check neutral passes", "v1.0.0", onMain,
			[]map[string]any{withRun(t, answer, "something else", suites[0], "completed", "neutral")}, 0, "every named job completed with success", false},
		{"another check still running passes", "v1.0.0", onMain,
			[]map[string]any{withRun(t, answer, "something else", suites[0], "queued", "")}, 0, "every named job completed with success", false},
		{"a failure on a second page fails", "v1.0.0", onMain,
			[]map[string]any{answer, withRun(t, answer, "packs", suites[0], "completed", "failure")}, 1, "packs: completed, failure", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ownArg := ownSuite
			if c.ownRun {
				ownArg = strconv.Itoa(own)
			}
			code, out := runTagCheck(t, dir, c.tag, c.sha, ownArg, c.answers...)
			if code != c.code || !strings.Contains(out, c.want) {
				t.Errorf("exit %d, want %d with %q; output:\n%s", code, c.code, c.want, out)
			}
		})
	}
}

// TestTagCheckUnreadableAnswer shows an answer that is not the check-runs endpoint's refused, naming why: GitHub's
// error object, text that is not JSON, nothing at all, and a page with no check_runs list among good pages.
func TestTagCheckUnreadableAnswer(t *testing.T) {
	needTagCheckTools(t)
	dir, onMain, _ := tagCheckRepo(t)
	good, err := os.ReadFile(filepath.Join("testdata", "check-runs-137e419f.json"))
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name   string
		answer []byte
		want   string
	}{
		{"not found", []byte(`{"message":"Not Found","documentation_url":"https://docs.github.com/rest","status":"404"}` + "\n"),
			"is not an answer of the check-runs endpoint"},
		{"not JSON", []byte("<html>502 Bad Gateway</html>\n"), "is not JSON"},
		{"empty", nil, "is empty"},
		{"a bad page after a good one", append(append([]byte{}, good...), []byte("\n{\"message\":\"Not Found\"}\n")...),
			"is not an answer of the check-runs endpoint"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			code, out := runTagCheckRaw(t, dir, "v1.0.0", onMain, ownSuite, c.answer)
			if code != 1 || !strings.Contains(out, c.want) {
				t.Errorf("exit %d, want 1 with %q; output:\n%s", code, c.want, out)
			}
		})
	}
}

// TestTagCheckWrongCall shows a wrong call refused with exit 2: --answer without --own-suite, an unknown argument.
func TestTagCheckWrongCall(t *testing.T) {
	needTagCheckTools(t)
	dir, onMain, _ := tagCheckRepo(t)
	script, err := filepath.Abs(filepath.Join("..", "..", filepath.FromSlash(tagCheckScript)))
	if err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"--answer", "x.json"}, {"--bogus"}, {"--own-suite"}} {
		cmd := exec.Command("sh", append([]string{script}, args...)...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(), "GITHUB_REF_NAME=v1.0.0", "GITHUB_SHA="+onMain, "GITHUB_REPOSITORY=LastStep/Bonsai")
		out, err := cmd.CombinedOutput()
		var exit *exec.ExitError
		if !errors.As(err, &exit) || exit.ExitCode() != 2 || !strings.Contains(string(out), "usage:") {
			t.Errorf("%v: want exit 2 with the usage, got %v:\n%s", args, err, out)
		}
	}
}
