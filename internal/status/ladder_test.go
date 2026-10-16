package status

// status --json's ladder (set 7; plan-5 5.4.2 note 12): the active task's last result, fresh at HEAD and stale after a
// new commit; null with no active task, no result, or a result made for another task.

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/schema"
)

const ladderCfg = `format: bonsai.workspace/1
id: ws-7kq2m4xw5r3t6y2u7p4a5c3e2b
name: example
documents:
  task: work/tasks
`

const runningTask = "---\nformat: bonsai.task/1\nid: T-0901\ntitle: A task\nstatus: running\nlane: null\ndone_when:\n  - \"x\"\n" +
	"depends_on: []\nblocked_by:\ncreated: 2026-10-13\nstarted:\nfinished:\nlabels: {}\n---\n\n# A task\n"

func headOf(t *testing.T, dir string) string {
	t.Helper()
	out, err := exec.Command("git", "-C", dir, "rev-parse", "HEAD").Output()
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimSpace(string(out))
}

func writeLadder(t *testing.T, root, name, task, sha string) {
	t.Helper()
	l := &format.Ladder{Task: &task, Workspace: "ws-7kq2m4xw5r3t6y2u7p4a5c3e2b", Mode: "local", Started: "2026-10-13T01:00:00.000Z",
		Finished: "2026-10-13T01:02:00.000Z", Git: format.LadderGit{SHA: sha, Branch: "main"}, Requested: []int64{1}, Green: true,
		Rungs: []format.LadderRung{}, Skipped: []format.LadderSkip{}}
	raw, err := l.Encode()
	if err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(root, ".bonsai", "local", "ladder", name+".json")
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, raw, 0o644); err != nil {
		t.Fatal(err)
	}
}

func ladderOf(t *testing.T, root string) string {
	t.Helper()
	doc, code := Build(root, "dev")
	if code != ExitOK {
		t.Fatalf("exit %d: %s", code, schema.Show(doc))
	}
	checkShape(t, doc, false)
	v, _ := doc.Get("ladder")
	return schema.Show(v)
}

func TestStatusLadder(t *testing.T) {
	t.Setenv("BONSAI_TASK", "") // the caller's named task is not this test's
	t.Setenv("BONSAI_ROLE", "")
	root, _ := project(t, ladderCfg)
	if err := os.MkdirAll(filepath.Join(root, "work", "tasks"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "work", "tasks", "T-0901-a.md"), []byte(runningTask), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".gitignore"), []byte(".bonsai/local/\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitRun(t, root, "add", "-A")
	gitRun(t, root, "commit", "-q", "-m", "x")
	if got := ladderOf(t, root); got != "null" {
		t.Errorf("no result yet: %s", got)
	}
	head := headOf(t, root)
	writeLadder(t, root, "T-0902", "T-0902", head)
	writeLadder(t, root, "T-0901", "T-0902", head) // made for another task
	if got := ladderOf(t, root); got != "null" {
		t.Errorf("a result made for another task: %s", got)
	}
	writeLadder(t, root, "T-0901", "T-0901", head)
	want := `{"task":"T-0901","path":".bonsai/local/ladder/T-0901.json","green":true,"mode":"local","sha":"` + head +
		`","finished":"2026-10-13T01:02:00.000Z","stale":false}`
	if got := ladderOf(t, root); got != want {
		t.Errorf("fresh: %s\nwant %s", got, want)
	}
	doc, _ := Build(root, "dev")
	if text := Text(doc); !strings.Contains(flat(text), "Last ladder green at commit "+head[:12]+" (local), finished ") || strings.Contains(text, "stale") {
		t.Errorf("the text:\n%s", text)
	}
	gitRun(t, root, "commit", "-q", "--allow-empty", "-m", "y")
	if got := ladderOf(t, root); got != strings.Replace(want, `"stale":false`, `"stale":true`, 1) {
		t.Errorf("stale: %s", got)
	}
	doc, _ = Build(root, "dev")
	if text := Text(doc); !strings.Contains(flat(text), " stale it proves another commit than HEAD, so climb again") {
		t.Errorf("the stale text:\n%s", text)
	}
}
