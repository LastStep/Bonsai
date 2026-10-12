//go:build !linux && !windows

package rungjob

import (
	"context"
	"testing"
)

// Elsewhere the process-tree tests do not run: rung jobs list and end a rung's processes on Linux and Windows only
// (spec §3: macOS builds, untested). A command still runs and gives its exit.
func TestRunsElsewhere(t *testing.T) {
	r := Run(context.Background(), Spec{Command: "exit 3", Dir: t.TempDir()})
	if r.Exit != 3 {
		t.Errorf("exit %d, reason %q", r.Exit, r.Reason)
	}
}
