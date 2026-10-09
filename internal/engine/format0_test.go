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
	// check finds the new format-0 file (format0-new), and the two files that do not read under their formats: a
	// format-1 task the reader refuses, and a memory note with no format: line (memory never had format 0).
	want := "document work/tasks/T-0903-refused.md, format0-new work/tasks/T-0904-later.md, document work/memory/M-note.md"
	if r, err := Check(root, e.home); err != nil || findingsOf(r) != want {
		t.Fatalf("check: %v\n%s\nwant %s", err, findingsOf(r), want)
	}
	writeFile(t, root, "work/tasks/T-0901-old.md", old+"more\r\n")
	if r, _ := Check(root, e.home); findingsOf(r) != "format0 work/tasks/T-0901-old.md, "+want {
		t.Errorf("a changed format-0 file: %s", findingsOf(r))
	}
	// Given a format: line, a listed file is held to format 1, no longer to its format-0 hash.
	writeFile(t, root, "work/tasks/T-0901-old.md", "---\nformat: bonsai.task/1\nid: T-0901\ntitle: An old task\nstatus: done\n---\n")
	if r, _ := Check(root, e.home); findingsOf(r) != want {
		t.Errorf("a format-0 file given a format: line: %s", findingsOf(r))
	}
}

// findingsOf lists a check's findings as "code file", in order.
func findingsOf(r *CheckResult) string {
	if r == nil {
		return "<nil>"
	}
	var out []string
	for _, f := range r.Findings {
		out = append(out, f.Code+" "+f.File)
	}
	return strings.Join(out, ", ")
}
