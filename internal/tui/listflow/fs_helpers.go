package listflow

import (
	"io/fs"
	"os"
	"path/filepath"
)

// Filesystem indirection kept in a single file so the workspace-scan code
// in agent_panel.go stays readable. Thin wrappers around the stdlib — the
// indirection also gives tests a single seam to stub in future (none do
// today; every test uses t.TempDir and real syscalls).

// osStat is a thin wrapper around os.Stat used by renderWorkspaceBlock's
// existence check. Kept isolated so the caller doesn't import "os".
func osStat(path string) (fs.FileInfo, error) { return os.Stat(path) }

// evalSymlinks wraps filepath.EvalSymlinks so scanWorkspace can resolve
// the workspace root + each symlink's target without the caller pulling
// in path/filepath directly alongside its own walk logic.
func evalSymlinks(path string) (string, error) { return filepath.EvalSymlinks(path) }
