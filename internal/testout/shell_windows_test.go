//go:build windows

package testout

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

// shellCommand runs a command line the way a rung runs on Windows: %ComSpec% /d /s /c "<line>", the line given to
// cmd as written (Go's own quoting of an argument would not be cmd's).
func shellCommand(ctx context.Context, line string) *exec.Cmd {
	comspec := os.Getenv("ComSpec")
	if comspec == "" || !filepath.IsAbs(comspec) {
		comspec = filepath.Join(os.Getenv("SystemRoot"), "System32", "cmd.exe")
	}
	cmd := exec.CommandContext(ctx, comspec)
	cmd.SysProcAttr = &syscall.SysProcAttr{CmdLine: `"` + comspec + `" /d /s /c "` + line + `"`}
	return cmd
}
