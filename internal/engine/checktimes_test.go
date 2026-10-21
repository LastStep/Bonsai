package engine

// Contract section 23 (P-T-5795 step 6, the orchestrator's ruling on T-5716's audit, Q2 (c)): bonsai check's half of
// "fresh stamps checked" needs no new word. Every stamp field of Bonsai's own kinds carries its form's pattern in its
// schema (a task's created, started and finished, STATE's and a memory note's updated: YYYY-MM-DD; a run report's
// started and finished: YYYY-MM-DD HH:MM), so a time written in another form, a zone name or an offset after it among
// them, is the finding document; a well-formed value, legacy or new, is none. A pack kind's stamp fields are judged by
// the guard alone (stamps.go in internal/guard); a check word for them is proposed for after 1.0.

import (
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/testpack"
)

func TestAStampInAnotherFormIsADocumentFinding(t *testing.T) {
	e := setup(t)
	root := testpack.Project(t, e.tmp, "stamps")
	e.link(t, root, e.pack.A)
	task := strings.Replace(f1Todo, "id: T-0902", "id: T-0903", 1)
	run := "---\nformat: bonsai.run/1\nid: R-2026-10-21-T-0903\ntask: T-0903\nrole: builder\nmodel: a-model\nstarted: " +
		"2026-10-21 09:16\nfinished: null\noutcome: running\ncommits: []\nlabels: {}\n---\n\n# Run\n"
	state := "---\nformat: bonsai.state/1\nupdated: 2026-10-21\nupdated_by: an agent\nlabels: {}\n---\n\n# Where it stands\n"
	note := "---\nformat: bonsai.memory/1\nid: M-a-note\ntitle: A note\nkind: project\nupdated: 2026-10-21\nsource: a test\n" +
		"labels: {}\n---\n\nThe note.\n"
	files := []struct{ path, good, field string }{
		{"work/tasks/T-0903-a.md", task, "created"},
		{"work/runs/R-2026-10-21-T-0903.md", run, "started"},
		{".bonsai/STATE.md", state, "updated"},
		{"work/memory/M-a-note.md", note, "updated"},
	}
	for _, f := range files {
		writeFile(t, root, f.path, f.good)
	}
	if r, err := checkLocal(t, root, e.home); err != nil || len(r.Findings) != 0 {
		t.Fatalf("well-formed stamps: %v %+v", err, r.Findings)
	}
	for _, f := range files {
		for _, bad := range []string{`"2026-10-21 09:16 IST"`, `"2026-10-21 14:46 +05:30"`, `"21 Oct 2026"`} {
			var broken string
			for _, l := range strings.Split(f.good, "\n") {
				if strings.HasPrefix(l, f.field+": ") {
					broken = strings.Replace(f.good, l, f.field+": "+bad, 1)
					break
				}
			}
			if broken == "" {
				t.Fatalf("%s has no %s line", f.path, f.field)
			}
			writeFile(t, root, f.path, broken)
			r, err := checkLocal(t, root, e.home)
			if err != nil || len(r.Findings) != 1 || r.Findings[0].Code != "document" || r.Findings[0].File != f.path {
				t.Errorf("%s with %s %s: %v %+v", f.path, f.field, bad, err, r.Findings)
			}
			writeFile(t, root, f.path, f.good)
		}
	}
}
