package engine

// The lock's format0 list (contract §2.3, §14): at a first link, every task, run report and STATE file with no
// format: line, with the SHA-256 of its bytes with line endings made LF; update keeps the list as it is.

import (
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/testpack"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// documentsYAML is a bonsai.yaml linking the test pack at ref, with its documents in work/.
func (e *env) documentsYAML(ref string) string {
	return "format: bonsai.workspace/1\nid: ws-aaaaaaaaaaaaaaaaaaaaaaaaaa\nname: demo\npacks:\n  - id: demo-pack\n    source: \"" +
		strings.ReplaceAll(e.pack.Source, `\`, "/") + "\"\n    ref: \"" + ref + "\"\ndocuments:\n  task: work/tasks\n  run: work/runs\n" +
		"  answers: work/answers.md\n  memory: work/memory\n  protocols: work/protocols\n"
}

func TestFormat0ListedAtAFirstLink(t *testing.T) {
	e := setup(t)
	root := testpack.Project(t, e.tmp, "format0")
	writeFile(t, root, "bonsai.yaml", e.documentsYAML(e.pack.A))
	old := "---\r\nid: T-0901\r\ntitle: An old task\r\nstatus: done\r\n---\r\n\r\nBody.\r\n"
	writeFile(t, root, "work/tasks/T-0901-old.md", old)
	writeFile(t, root, "work/tasks/T-0902-new.md", "---\nformat: bonsai.task/1\nid: T-0902\n---\n")
	writeFile(t, root, "work/tasks/T-0903-refused.md", "---\nformat: bonsai.task/1\nmode: 0755\n---\n")
	writeFile(t, root, "work/tasks/README.md", "# Tasks\n")
	writeFile(t, root, "work/runs/R-2026-10-01-T-0901.md", "# A run report with no frontmatter\n")
	writeFile(t, root, ".bonsai/STATE.md", "---\nupdated: 2026-10-01\n---\n# State\n")
	writeFile(t, root, "work/memory/M-note.md", "---\ntitle: not a format-0 kind\n---\n")
	p := e.apply(t, root, Request{Command: "init", AllowExec: true})
	if p.Format0 != 3 || !strings.Contains(p.Preview(false), "listed       3 task, run report and STATE files with no format: line") {
		t.Errorf("format0 %d\n%s", p.Format0, p.Preview(false))
	}
	l, err := workspace.LoadLock(root)
	if err != nil {
		t.Fatal(err)
	}
	got := sortedPaths(l.Format0)
	if strings.Join(got, " ") != ".bonsai/STATE.md work/runs/R-2026-10-01-T-0901.md work/tasks/T-0901-old.md" {
		t.Errorf("format0 lists %v", got)
	}
	if l.Format0["work/tasks/T-0901-old.md"] != workspace.HashLF([]byte(strings.ReplaceAll(old, "\r\n", "\n"))) {
		t.Errorf("a CRLF file's hash is not its LF bytes'")
	}
	// update keeps the list as it is: a format-0 file new since the link is not added (check finds it, step 5.1.6).
	writeFile(t, root, "work/tasks/T-0904-later.md", "---\nid: T-0904\n---\n")
	before := read(t, root, workspace.LockFile)
	if p := e.apply(t, root, Request{}); p.Format0 != 0 || read(t, root, workspace.LockFile) != before {
		t.Errorf("update changed the format0 list")
	}
	if r, err := Check(root, e.home); err != nil || len(r.Findings) != 0 {
		t.Fatalf("check: %v %+v", err, r.Findings)
	}
	writeFile(t, root, "work/tasks/T-0901-old.md", old+"more\r\n")
	if r, _ := Check(root, e.home); len(r.Findings) != 1 || r.Findings[0].Code != "format0" {
		t.Errorf("a changed format-0 file: %+v", r.Findings)
	}
}
