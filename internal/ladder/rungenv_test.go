package ladder

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/ratchet"
	"github.com/LastStep/Bonsai/internal/testpack"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// A rung never sees the caller's task or role: the runner clears BONSAI_TASK and BONSAI_ROLE from its environment and
// passes the rest through. The fixture rung prints its environment into a file the climb ignores (.bonsai/local/).
func TestRungsEnvironmentHasNoTaskOrRole(t *testing.T) {
	dump := "env > .bonsai/local/rungenv.txt"
	if runtime.GOOS == "windows" {
		dump = `set > .bonsai\local\rungenv.txt`
	}
	f := project(t, yamlOf("[1]", "{}", "[]", rung{n: 1, name: "env", command: dump, required: true}), map[string]string{
		"work/tasks/T-0901-demo.md": taskText("T-0901", "running", ""),
	})
	if err := os.MkdirAll(filepath.Join(f.root, ".bonsai", "local"), 0o755); err != nil {
		t.Fatal(err)
	}
	f.env[workspace.TaskEnv] = "T-0901"
	f.env["BONSAI_ROLE"] = "builder"
	caller := append(os.Environ(), workspace.TaskEnv+"=T-0901", "BONSAI_ROLE=builder", "BONSAI_PASSES_THROUGH=yes")
	a := f.climb("", Options{Task: "T-0901", Env: caller})
	if a.Exit != 0 {
		t.Fatalf("exit %d\n%s", a.Exit, f.out.String())
	}
	raw, err := os.ReadFile(filepath.Join(f.root, ".bonsai", "local", "rungenv.txt"))
	if err != nil {
		t.Fatal(err)
	}
	got := strings.ToUpper(string(raw))
	for _, name := range []string{workspace.TaskEnv, "BONSAI_ROLE"} {
		if strings.Contains(got, name+"=") {
			t.Errorf("the rung's environment holds %s:\n%s", name, raw)
		}
	}
	if !strings.Contains(got, "BONSAI_PASSES_THROUGH=YES") {
		t.Errorf("the rest of the environment did not pass through:\n%s", raw)
	}
	if !strings.Contains(string(raw), "BONSAI_LADDER=1") {
		t.Errorf("the runner's own LadderEnv is missing:\n%s", raw)
	}
}

// The new-tests check's reruns on the base (kinds.go's riseStep) run with the rungs' own environment too: neither
// BONSAI_TASK nor BONSAI_ROLE while the caller's holds both, the rest passed through (VG's item H, 5.5.8 item 4).
func TestNewTestsRerunsHaveNoTaskOrRole(t *testing.T) {
	ntGo.need(t)
	f, tmp := ntGo.project(t, "")
	wt := filepath.Join(f.tmp, "wt")
	testpack.Git(t, f.root, "worktree", "add", "-q", "-b", "t0901", wt)
	ntGo.lay(t, wt)
	defer func(saved func(ratchet.RiseInput) ratchet.Checked) { newTests = saved }(newTests)
	var seen [][]string
	newTests = func(in ratchet.RiseInput) ratchet.Checked {
		seen = append(seen, append([]string(nil), in.Env...))
		return ratchet.NewTests(in)
	}
	caller := append(os.Environ(), workspace.TaskEnv+"=T-0901", "BONSAI_ROLE=builder", "BONSAI_PASSES_THROUGH=yes")
	a := climbClean(t, f, tmp, wt, Options{Task: "T-0901", Env: caller})
	if a.Exit != 0 || len(seen) != 1 {
		t.Fatalf("exit %d, %d new-tests checks\n%s", a.Exit, len(seen), f.out.String())
	}
	through := false
	for _, kv := range seen[0] {
		name, _, _ := strings.Cut(strings.ToUpper(kv), "=")
		if name == workspace.TaskEnv || name == "BONSAI_ROLE" {
			t.Errorf("the reruns' environment holds %s", kv)
		}
		through = through || kv == "BONSAI_PASSES_THROUGH=yes"
	}
	if !through {
		t.Error("the rest of the caller's environment did not reach the reruns")
	}
}

func TestRungEnvKeepsOrderAndOtherNames(t *testing.T) {
	in := []string{"A=1", "BONSAI_TASK=T-1", "BONSAI_TASK_X=2", "BONSAI_ROLE=r", "B=BONSAI_TASK=3"}
	want := []string{"A=1", "BONSAI_TASK_X=2", "B=BONSAI_TASK=3"}
	got := rungEnv(in)
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Errorf("rungEnv = %q, want %q", got, want)
	}
}
