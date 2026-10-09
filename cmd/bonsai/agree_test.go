package main

// update and check agree (step 5.1.1's verifier's loose end, settled in 5.1.6): when update has nothing to write but
// leaves a file edited here, its "nothing to change" names that file and how to settle it, the finding check reports;
// and init and update record the checkout on this machine (the record check's id-changed and same-id read).

import (
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/workspace"
)

func TestUpdateAgreesWithCheck(t *testing.T) {
	c := newCLI(t)
	c.link()
	c.write("demo/guide.md", "mine\n")
	code, out, _ := c.run("", "update", "--yes")
	if code != 0 || !strings.HasPrefix(out, "bonsai update: nothing to change; demo-pack at ") ||
		!strings.Contains(out, "Left as they are, edited here (bonsai check reports each as changed until it is settled):\n  demo/guide.md [pack]: ") ||
		!strings.Contains(out, "next: to keep the edit, run: bonsai update --yes --keep demo/guide.md; to take the pack's copy back, run: bonsai update --yes --adopt demo/guide.md") {
		t.Errorf("update with an edited file: %d\n%s", code, out)
	}
	if code, out, _ := c.run("", "check"); code != 1 || !strings.Contains(out, "demo/guide.md was edited") {
		t.Errorf("check: %d\n%s", code, out)
	}
	// The JSON says the same: result nothing, the file changed.
	code, out, _ = c.run("", "update", "--yes", "--json")
	doc := fits(t, out, "changes")
	files, _ := doc.Get("files")
	if code != 0 || doc.String("result") != "nothing" || !strings.Contains(out, `"path": "demo/guide.md"`) || len(files.([]any)) == 0 {
		t.Errorf("update --json: %d\n%s", code, out)
	}
	// The record: this checkout's path and id, written by the link.
	cfg, _ := workspace.LoadConfig(c.root)
	home, _ := workspace.Home()
	if rec, err := workspace.LoadMachineRecord(home, c.root); err != nil || rec == nil || rec.Current() != cfg.ID {
		t.Errorf("no record of the checkout: %+v %v", rec, err)
	}
}
