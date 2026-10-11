package guard

// One damaged machine record un-verifies every project on the machine (workspace.FirstClaim cannot tell which checkout
// claimed an id first; V1's note on 5.3.1): the refusal names that record, from Bonsai's home, and the person's step.

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/workspace"
)

func TestADamagedRecordIsNamed(t *testing.T) {
	dir := withTask(t, "running", "assets/**")
	home, _ := workspace.Home()
	wantRule(t, dir, writeOf(dir, "assets/a.png", "x\n"), RuleGranted)
	bad := filepath.Join(home, "workspaces", "r-00000000000000aa", workspace.MachineRecordFile)
	writeFile(t, bad, "not json\n")
	s := wantRule(t, dir, writeOf(dir, "assets/a.png", "x\n"), RuleMainNotVerified)
	if !strings.Contains(s, "workspaces/r-00000000000000aa/workspace.json in Bonsai's home") || !strings.Contains(s, "next: a person checks") {
		t.Errorf("the refusal does not name the damaged record and the step: %s", s)
	}
}
