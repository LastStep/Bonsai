package ladder

// Step 5.3's stop gate reads what the runner writes (start check 2.6): a green local climb at HEAD lets a named session
// end; the same task's result from a --ci climb, a red climb, a climb of another commit and a climb of a dirty tree
// each keep it going, with the gate's own rule.

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/hookin"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/stopgate"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// gate runs the stop gate for a session in root whose BONSAI_TASK names task, and gives its exit code and stderr.
func gate(t *testing.T, root, task string) (int, string) {
	t.Helper()
	payload, err := schema.EncodeLine(schema.Object{{Key: "session_id", Value: "s1"}, {Key: "cwd", Value: root},
		{Key: "hook_event_name", Value: "Stop"}, {Key: "stop_hook_active", Value: false}})
	if err != nil {
		t.Fatal(err)
	}
	vars := map[string]string{hookin.ProjectEnv: root, workspace.TaskEnv: task}
	var stderr bytes.Buffer
	code := stopgate.Main(stopgate.Options{Stdin: bytes.NewReader(payload), Stderr: &stderr, Getenv: func(k string) string { return vars[k] },
		Budget: time.Minute})
	return code, stderr.String()
}

func TestTheStopGateReadsTheResult(t *testing.T) {
	rungs := []rung{{n: 1, name: "one", command: "echo one", required: true}, {n: 2, name: "two", command: show("verdict.txt"), required: true},
		{n: 3, name: "three", command: "echo three", required: true}}
	f := project(t, yamlOf("[1]", "{}", "[]", rungs...), map[string]string{
		"work/tasks/T-0901-a.md": taskText("T-0901", "running", "[2]"),
		"verdict.txt":            "fine\n",
	})
	if code, stderr := gate(t, f.root, "T-0901"); code != 2 || !strings.Contains(stderr, "no ladder result proves T-0901") {
		t.Fatalf("before any climb: %d %q", code, stderr)
	}
	// Green, local, at HEAD, clean, its rungs the floor and the task's: the session may end.
	if a := f.climb("", Options{Task: "T-0901"}); a.Exit != 0 {
		t.Fatalf("the green climb: %d %+v\n%s", a.Exit, a.Error, f.out.String())
	}
	if l := f.result("T-0901"); len(l.Requested) != 2 || l.Requested[0] != 1 || l.Requested[1] != 2 {
		t.Errorf("requested %v, want the floor and the task's: 1, 2", l.Requested)
	}
	if code, stderr := gate(t, f.root, "T-0901"); code != 0 {
		t.Fatalf("the green local result at HEAD is refused: %d %q", code, stderr)
	}
	result := filepath.Join(f.root, ".bonsai", "local", "ladder", "T-0901.json")
	if _, err := os.ReadFile(result); err != nil {
		t.Fatal(err)
	}
	// The same task's --ci result, put where the task's is: a CI run is never a task's proof.
	if a := f.climb("", Options{Task: "T-0901", CI: true}); a.Exit != 0 {
		t.Fatalf("the ci climb: %d", a.Exit)
	}
	ci, _ := os.ReadFile(filepath.Join(f.root, ".bonsai", "local", "ladder", "ci.json"))
	if err := os.WriteFile(result, ci, 0o644); err != nil {
		t.Fatal(err)
	}
	if code, stderr := gate(t, f.root, "T-0901"); code != 2 || !strings.Contains(stderr, `is a "ci" result`) {
		t.Errorf("a ci result: %d %q", code, stderr)
	}
	// A red climb.
	f.write("verdict.txt", "fine\n")
	f.write("bonsai.yaml", strings.Replace(yamlOf("[1]", "{}", "[]", rungs...), show("verdict.txt"), "exit 1", 1))
	f.commit("rung 2 red")
	if a := f.climb("", Options{Task: "T-0901"}); a.Exit != ExitRed {
		t.Fatalf("the red climb: %d", a.Exit)
	}
	if code, stderr := gate(t, f.root, "T-0901"); code != 2 || !strings.Contains(stderr, "is not green") {
		t.Errorf("a red result: %d %q", code, stderr)
	}
	// A green result of the commit before.
	f.write("bonsai.yaml", yamlOf("[1]", "{}", "[]", rungs...))
	f.commit("rung 2 green again")
	if a := f.climb("", Options{Task: "T-0901"}); a.Exit != 0 {
		t.Fatalf("green again: %d", a.Exit)
	}
	f.commit("one more commit")
	if code, stderr := gate(t, f.root, "T-0901"); code != 2 || !strings.Contains(stderr, "proves commit") {
		t.Errorf("another commit: %d %q", code, stderr)
	}
	// A green climb of a dirty tree.
	f.write("loose.txt", "not committed\n")
	if a := f.climb("", Options{Task: "T-0901"}); a.Exit != 0 || !f.result("T-0901").Git.Dirty {
		t.Fatalf("the dirty climb: %d", a.Exit)
	}
	if code, stderr := gate(t, f.root, "T-0901"); code != 2 || !strings.Contains(stderr, "uncommitted changes") {
		t.Errorf("a dirty tree: %d %q", code, stderr)
	}
}
