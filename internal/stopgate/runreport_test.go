package stopgate

// The stop gate reads the task's run report as bonsai check does (T-5731; design/pi-bonsai-work.md, B3): a run report
// on which check makes a finding of its own (format0-new, document, label, absolute-path) keeps the session going
// (stop-bad-run-report), its reason check's finding, its next step check's; a report check accepts lets it go as
// before. The project here is linked: it has a lock, whose format0 list holds one legacy report and whose one pack
// declares Bonsai's label bonsai.branch (on tasks only).

import (
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/workspace"
)

// gateLock is the test project's lock: one pack declaring bonsai.branch for tasks, and a format0 list holding one
// legacy run report of T-0901's.
const gateLock = `{
  "format": "bonsai.lock/1",
  "written_by": "0.0.0-test",
  "packs": [
    {
      "id": "base",
      "source": "https://example.invalid/base.git",
      "version": "0.1.0",
      "commit": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
      "sha256": "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
      "declares": {
        "labels": {
          "namespace": "bonsai",
          "version": 1,
          "labels": [
            {
              "name": "bonsai.branch",
              "kind": "text",
              "values": [],
              "items": null,
              "pattern": "^[A-Za-z0-9._/-]{1,100}$",
              "max": null,
              "kinds": ["task"],
              "set_by": "agent",
              "grants": false,
              "description": "The branch the task is built on."
            }
          ]
        }
      },
      "path": null
    }
  ],
  "files": {},
  "format0": {
    "work/runs/R-2026-10-01-T-0901-legacy.md": "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
  }
}
`

// goodRun is a format-1 run report of T-0901 that check accepts, mid-session: finished null, outcome running.
const goodRun = "---\nformat: bonsai.run/1\nid: R-2026-10-11-T-0901\ntask: T-0901\nrole: builder\nmodel: a-model\n" +
	"started: 2026-10-11 09:16\nfinished: null\noutcome: running\ncommits: []\nlabels: {}\n---\n\n# Run\n\nWhat changed.\n"

const reportPath = "work/runs/R-2026-10-11-T-0901.md"

// linkedProject is project(t) with the lock above, its run report the good one.
func linkedProject(t *testing.T) string {
	t.Helper()
	root := project(t)
	write(t, root, workspace.LockFile, gateLock)
	if _, err := workspace.LoadLock(root); err != nil {
		t.Fatalf("the test's lock does not read: %v", err)
	}
	write(t, root, reportPath, goodRun)
	return root
}

func TestARunReportCheckRejectsHoldsTheGate(t *testing.T) {
	ladder := installed + " ladder --task T-0901"
	cases := []struct {
		name, report, why, next string
	}{
		{name: "no frontmatter (format0-new)", report: "# R-2026-10-11-T-0901\n\nWhat changed.\n",
			why: "(format0-new)", next: "format: bonsai.run/1"},
		{name: "a frontmatter with no format: line (format0-new)", report: "---\ntask: T-0901\nstarted: 2026-10-11\n---\n\n# Run\n",
			why: "(format0-new)", next: "format: bonsai.run/1"},
		{name: "started without a time (document)", report: strings.Replace(goodRun, "started: 2026-10-11 09:16", "started: 2026-10-11", 1),
			why: "field started", next: "bonsai check --schema bonsai.run"},
		{name: "an outcome off its list (document)", report: strings.Replace(goodRun, "outcome: running", "outcome: finished", 1),
			why: "field outcome", next: "bonsai check --schema bonsai.run"},
		{name: "a label on a kind it does not go on (label)", report: strings.Replace(goodRun, "labels: {}", "labels:\n  bonsai.branch: T-0901", 1),
			why: "bonsai.branch", next: "take bonsai.branch out"},
		{name: "an absolute path in a field (absolute-path)", report: strings.Replace(goodRun, "model: a-model", "model: \"/opt/model\"", 1),
			why: "(absolute-path)", next: "bonsai check"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			root := linkedProject(t)
			write(t, root, reportPath, c.report)
			code, stderr := gateFor(t, root, "T-0901")
			blocks(t, code, stderr, root, RuleBadRunReport, c.why, c.next)
			if !strings.Contains(stderr, "the run report "+reportPath+" does not pass bonsai check") {
				t.Errorf("the reason does not name the report: %q", stderr)
			}
			if !strings.Contains(stderr, ladder) {
				t.Errorf("the next step does not end with the ladder: %q", stderr)
			}
			r := lastRecord(t, root)
			if r.String("target") != reportPath {
				t.Errorf("the block's target is %q, want %q", r.String("target"), reportPath)
			}
		})
	}
}

// A report check accepts lets the gate go: format 1 mid-session, and a format-0 report the lock's format0 list holds.
func TestARunReportCheckAcceptsLetsTheGateGo(t *testing.T) {
	t.Run("format 1, finished null, outcome running", func(t *testing.T) {
		root := linkedProject(t)
		code, stderr := gateFor(t, root, "T-0901")
		allows(t, code, stderr, root)
	})
	t.Run("format 0 on the lock's format0 list", func(t *testing.T) {
		root := linkedProject(t)
		remove(t, root, reportPath)
		write(t, root, "work/runs/R-2026-10-01-T-0901-legacy.md", "# R-2026-10-01-T-0901\n\nWhat changed.\n")
		code, stderr := gateFor(t, root, "T-0901")
		allows(t, code, stderr, root)
	})
	t.Run("no lock: format 0 passes, as in check", func(t *testing.T) {
		root := project(t)
		code, stderr := gateFor(t, root, "T-0901")
		allows(t, code, stderr, root)
	})
}

// Every run report of the task is judged: a good one beside a bad one does not let the gate go, whichever sorts first;
// another task's bad report does not hold it.
func TestEveryRunReportOfTheTaskIsJudged(t *testing.T) {
	bad := strings.Replace(goodRun, "started: 2026-10-11 09:16", "started: 2026-10-11", 1)
	for _, name := range []string{"work/runs/R-2026-10-10-T-0901.md", "work/runs/R-2026-10-12-T-0901-b.md"} {
		t.Run(name, func(t *testing.T) {
			root := linkedProject(t)
			write(t, root, name, bad)
			code, stderr := gateFor(t, root, "T-0901")
			blocks(t, code, stderr, root, RuleBadRunReport, "the run report "+name+" does not pass bonsai check", "bonsai check --schema bonsai.run")
		})
	}
	t.Run("another task's report", func(t *testing.T) {
		root := linkedProject(t)
		write(t, root, "work/runs/R-2026-10-11-T-0902.md", "# no frontmatter\n")
		code, stderr := gateFor(t, root, "T-0901")
		allows(t, code, stderr, root)
	})
}
