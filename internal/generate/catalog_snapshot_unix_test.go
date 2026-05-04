//go:build !windows

package generate

import (
	"os"
	"path/filepath"
	"testing"
)

// TestWriteCatalogSnapshot_RefusesSymlink — pre-create a symlink at the
// target path; the writer must error out (O_NOFOLLOW) and leave the link
// intact so an attacker-planted symlink cannot be used to overwrite an
// arbitrary file.
func TestWriteCatalogSnapshot_RefusesSymlink(t *testing.T) {
	cat := buildMinimalCatalog(t)
	tmpDir := t.TempDir()

	bonsaiDir := filepath.Join(tmpDir, ".bonsai")
	if err := os.MkdirAll(bonsaiDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	target := filepath.Join(bonsaiDir, "catalog.json")
	if err := os.Symlink("/dev/null", target); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	var wr WriteResult
	err := WriteCatalogSnapshot(tmpDir, "v0.0.0", cat, &wr)
	if err == nil {
		t.Fatal("WriteCatalogSnapshot: want error for symlink target, got nil")
	}

	// Symlink must still be a symlink (Lstat, not Stat — Stat follows).
	info, lerr := os.Lstat(target)
	if lerr != nil {
		t.Fatalf("lstat after refused write: %v", lerr)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Errorf("target is no longer a symlink (mode=%v)", info.Mode())
	}
}
