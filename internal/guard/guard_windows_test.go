package guard

import (
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"github.com/LastStep/Bonsai/internal/workspace"
)

// An NTFS short (8.3) name of a protected file is the protected file.
func TestDecideReadsShortNames(t *testing.T) {
	dir := project(t, testYAML)
	cfg, err := workspace.LoadConfig(dir)
	if err != nil {
		t.Fatal(err)
	}
	long := filepath.Join(dir, "protected.txt")
	in, err := syscall.UTF16PtrFromString(long)
	if err != nil {
		t.Fatal(err)
	}
	buf := make([]uint16, 1024)
	n, err := syscall.GetShortPathName(in, &buf[0], uint32(len(buf)))
	if err != nil || n == 0 {
		t.Skipf("no short name for %s (%v)", long, err)
	}
	short := syscall.UTF16ToString(buf[:n])
	if strings.EqualFold(filepath.Base(short), "protected.txt") {
		t.Skip("short (8.3) names are off on this volume: protected.txt has none")
	}
	d := Decide(&Input{Tool: "Write", Path: short}, dir, cfg)
	if d.Allow || d.Rule != RuleProtected {
		t.Fatalf("%s: %+v, want refused as protected", short, d)
	}
}
