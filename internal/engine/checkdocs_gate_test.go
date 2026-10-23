package engine

// DocumentFindings is check's own reading of one document, for the stop gate (T-5731): for each run report below it
// gives exactly the findings bonsai check makes on that file, so the gate cannot drift from check.

import (
	"reflect"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/workspace"
)

func TestDocumentFindingsAreChecksOwn(t *testing.T) {
	e := setup(t)
	root, _ := e.linkBase(t, "gate")
	cfg, err := workspace.LoadConfigFull(root)
	if err != nil {
		t.Fatal(err)
	}
	kinds, _ := workspace.DocKinds(cfg.Full, nil)
	var run workspace.DocKind
	for _, k := range kinds {
		if k.Kind == "run" {
			run = k
		}
	}
	if run.Path == "" {
		t.Fatal("the fixture's bonsai.yaml names no run folder")
	}
	p := strings.TrimSuffix(run.Path, "/") + "/R-2026-10-21-T-0903.md"
	// A legacy report on the lock's format0 list, at its own hash, as a link records one.
	legacy := strings.TrimSuffix(run.Path, "/") + "/R-2026-10-01-T-0903-legacy.md"
	legacyBody := "# R-2026-10-01-T-0903\n\nWhat changed.\n"
	lock, err := workspace.LoadLock(root)
	if err != nil {
		t.Fatal(err)
	}
	if lock.Format0 == nil {
		lock.Format0 = map[string]string{}
	}
	lock.Format0[legacy] = workspace.HashLF([]byte(legacyBody))
	if err := workspace.WriteLock(root, lock); err != nil {
		t.Fatal(err)
	}
	good := "---\nformat: bonsai.run/1\nid: R-2026-10-21-T-0903\ntask: T-0903\nrole: builder\nmodel: a-model\nstarted: " +
		"2026-10-21 09:16\nfinished: null\noutcome: running\ncommits: []\nlabels: {}\n---\n\n# Run\n"
	cases := []struct{ name, path, body, code string }{
		{"good", p, good, ""},
		{"format 0 on the lock's list, as it was", legacy, legacyBody, ""},
		{"format 0 on the lock's list, changed", legacy, legacyBody + "\nA line added later.\n", "format0"},
		{"no frontmatter", p, "# Run\n\nWhat changed.\n", "format0-new"},
		{"started without a time", p, strings.Replace(good, "2026-10-21 09:16", "2026-10-21", 1), "document"},
		{"a label of tasks", p, strings.Replace(good, "labels: {}", "labels:\n  bonsai.branch: main", 1), "label"},
		{"an absolute path", p, strings.Replace(good, "model: a-model", "model: \"/opt/model\"", 1), "absolute-path"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			writeFile(t, root, c.path, c.body)
			r, err := Check(root, e.home)
			if err != nil {
				t.Fatal(err)
			}
			var want []Finding
			for _, f := range r.Findings {
				if f.File == c.path {
					want = append(want, f)
				}
			}
			got := DocumentFindings(root, r.Main, e.home, run, c.path, []byte(c.body))
			if !reflect.DeepEqual(got, want) {
				t.Errorf("DocumentFindings gives %+v, check %+v", got, want)
			}
			if c.code == "" && len(got) != 0 || c.code != "" && (len(got) != 1 || got[0].Code != c.code) {
				t.Errorf("findings %+v, want one %q", got, c.code)
			}
		})
	}
}
