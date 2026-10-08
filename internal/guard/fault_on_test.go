//go:build bonsai_test_fault

package guard

import (
	"strings"
	"testing"
	"time"
)

// Each fault, in process: every one blocks the free.txt edit the normal guard allows. The real process ends (a
// crash's hard death, the slow fault's sleep past the real budget, the hook line's shell with no bonsai on its PATH)
// are cmd/bonsai's TestFaultsThroughTheHookLine, on the built binaries.
func TestEachFaultBlocks(t *testing.T) {
	if !FaultBuild {
		t.Fatal("fault_on.go says this is not a fault build")
	}
	// Set for the rest of this test binary, not restored: the crash and slow runs leave their work goroutines
	// waiting in die and in the sleep, and restoring would race with them. slowFor never ends while the tests run, so
	// no late record lands in a removed folder.
	died := make(chan struct{}, 4)
	die = func() { died <- struct{}{}; select {} }
	slowFor = time.Hour

	cases := []struct {
		fault, inStderr string
		rules           []string
	}{
		{"crash", "test fault crash (BONSAI_TEST_FAULT=crash): the guard dies now", []string{"fault-crash", RuleOverTime}},
		{"slow", "the guard ran past its own 200ms limit", []string{RuleOverTime}},
		{"missing", "test fault missing (BONSAI_TEST_FAULT=missing): this session should find no bonsai on its PATH", []string{"fault-missing"}},
		{"minimal-path", "test fault minimal-path (BONSAI_TEST_FAULT=minimal-path)", []string{"fault-minimal-path"}},
		{"sideways", `BONSAI_TEST_FAULT="sideways" is not a test fault this build knows`, []string{"fault-unknown"}},
	}
	for _, c := range cases {
		t.Run(c.fault, func(t *testing.T) {
			dir := project(t, testYAML)
			code, stderr := guardRun(t, env(dir, FaultEnv, c.fault), strings.NewReader(editOf("s-f", dir, "free.txt")), 200*time.Millisecond)
			if code != 2 || !strings.Contains(stderr, c.inStderr) || !strings.Contains(stderr, "\nnext: ") {
				t.Fatalf("exit %d, stderr %q; want 2 with %q", code, stderr, c.inStderr)
			}
			asciiOnly(t, stderr)
			recs := records(t, dir, "s-s-f.ndjson")
			if len(recs) != len(c.rules) {
				t.Fatalf("%d records, want %v", len(recs), c.rules)
			}
			for i, r := range recs {
				if r.String("rule") != c.rules[i] || r.String("decision") != "deny" || r.String("target") != "" {
					t.Errorf("record %d: rule %s decision %s target %q", i, r.String("rule"), r.String("decision"), r.String("target"))
				}
			}
			if recs[0].String("bonsai_sha256") == "" {
				t.Errorf("the session's first record names no hash")
			}
		})
	}
	select {
	case <-died:
	default:
		t.Errorf("the crash fault never called die")
	}
}

// With the switch unset, a fault build decides as a normal one.
func TestFaultBuildWithNoFaultIsNormal(t *testing.T) {
	dir := project(t, testYAML)
	if code, stderr := guardRun(t, env(dir), strings.NewReader(editOf("s-g", dir, "free.txt")), 0); code != 0 || stderr != "" {
		t.Fatalf("free.txt: exit %d, %q", code, stderr)
	}
	if code, _ := guardRun(t, env(dir), strings.NewReader(editOf("s-g", dir, "protected.txt")), 0); code != 2 {
		t.Fatalf("protected.txt: exit %d", code)
	}
}
