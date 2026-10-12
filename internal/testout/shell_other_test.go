//go:build !windows

package testout

import (
	"context"
	"os/exec"
)

// shellCommand runs a command line the way a rung runs on Linux and macOS: /bin/sh -c.
func shellCommand(ctx context.Context, line string) *exec.Cmd {
	return exec.CommandContext(ctx, "/bin/sh", "-c", line)
}
