//go:build !bonsai_test_fault

package guard

import (
	"strings"
	"testing"
)

// A normal build reads no fault switch: with BONSAI_TEST_FAULT set to each fault, it decides as always. (cmd/bonsai's
// TestNormalBuildHasNoFaultCode checks the built binary for any trace of the switch.)
func TestNormalBuildIgnoresTheFaultSwitch(t *testing.T) {
	if FaultBuild {
		t.Fatal("fault_off.go says this is a fault build")
	}
	dir := project(t, testYAML)
	for _, fault := range []string{"missing", "crash", "slow", "minimal-path", "anything"} {
		code, stderr := guardRun(t, env(dir, "BONSAI_TEST"+"_FAULT", fault), strings.NewReader(editOf("s-n", dir, "free.txt")), 0)
		if code != 0 || stderr != "" {
			t.Errorf("%s: exit %d, stderr %q; want the normal allow", fault, code, stderr)
		}
		code, _ = guardRun(t, env(dir, "BONSAI_TEST"+"_FAULT", fault), strings.NewReader(editOf("s-n", dir, "protected.txt")), 0)
		if code != 2 {
			t.Errorf("%s: protected.txt exit %d, want the normal refusal", fault, code)
		}
	}
}
