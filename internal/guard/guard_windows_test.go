package guard

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"golang.org/x/sys/windows"

	"github.com/LastStep/Bonsai/internal/workspace"
)

// windowsProject is project with a protected folder docs holding guide.md, and its config.
func windowsProject(t *testing.T) (string, *workspace.Config) {
	t.Helper()
	dir := project(t, testYAML)
	if err := os.MkdirAll(filepath.Join(dir, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "docs", "guide.md"), []byte("guide\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := workspace.LoadConfig(dir)
	if err != nil {
		t.Fatal(err)
	}
	return dir, cfg
}

func refused(t *testing.T, cfg *workspace.Config, root, p string) {
	t.Helper()
	d := Decide(&Input{Tool: "Write", Path: p}, root, cfg)
	if d.Allow {
		t.Errorf("%s allowed (rule %s), want refused", p, d.Rule)
	}
}

// A stream in a folder's name is the folder: secrets::$INDEX_ALLOCATION\key.txt writes secrets\key.txt.
func TestDecideReadsDirectoryStreams(t *testing.T) {
	dir, cfg := windowsProject(t)
	for _, p := range []string{
		dir + `\docs::$INDEX_ALLOCATION\guide.md`,
		dir + `\docs::$INDEX_ALLOCATION\new.md`,
		dir + `\docs:$I30:$INDEX_ALLOCATION\guide.md`,
		dir + `\docs:$I30:$INDEX_ALLOCATION\new.md`,
	} {
		refused(t, cfg, dir, p)
	}
}

// mklinkJ makes a directory junction with cmd's mklink /J, which needs no privilege (unlike a symbolic link).
func mklinkJ(t *testing.T, link, target string) {
	t.Helper()
	out, err := exec.Command("cmd.exe", "/c", "mklink", "/J", link, target).CombinedOutput()
	if err != nil {
		t.Fatalf("mklink /J %s %s: %v %s", link, target, err, out)
	}
}

// A junction to a protected folder, or to the project, is followed: Go no longer follows junctions itself.
func TestDecideFollowsJunctions(t *testing.T) {
	dir, cfg := windowsProject(t)
	outside, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	mklinkJ(t, filepath.Join(outside, "j"), filepath.Join(dir, "docs"))
	mklinkJ(t, filepath.Join(outside, "proj"), dir)
	refused(t, cfg, dir, filepath.Join(outside, "j", "guide.md"))
	refused(t, cfg, dir, filepath.Join(outside, "j", "new.md"))
	refused(t, cfg, dir, filepath.Join(outside, "proj", "protected.txt"))
	// The project reached through a junction, the path by its real folder.
	refused(t, cfg, filepath.Join(outside, "proj"), filepath.Join(dir, "protected.txt"))
	if d := Decide(&Input{Tool: "Write", Path: filepath.Join(outside, "j2.txt")}, dir, cfg); !d.Allow {
		t.Errorf("a file beside the junctions refused: %+v", d)
	}
}

// Device and volume forms of a protected path: \\?\GLOBALROOT\Device\HarddiskVolumeN\... and \\?\Volume{...}\....
func TestDecideReadsDevicePaths(t *testing.T) {
	dir, cfg := windowsProject(t)
	vol := filepath.VolumeName(dir) // C:
	vp, err := windows.UTF16PtrFromString(vol)
	if err != nil {
		t.Fatal(err)
	}
	buf := make([]uint16, 1024)
	n, err := windows.QueryDosDevice(vp, &buf[0], uint32(len(buf)))
	if err != nil || n == 0 {
		t.Fatalf("QueryDosDevice(%s): %v", vol, err)
	}
	device := windows.UTF16ToString(buf[:n]) // \Device\HarddiskVolumeN
	rest := dir[len(vol):]                   // \Users\...
	refused(t, cfg, dir, `\\?\GLOBALROOT`+device+rest+`\protected.txt`)
	refused(t, cfg, dir, `\\?\GLOBALROOT`+device+rest+`\docs\new.md`)
	mp, err := windows.UTF16PtrFromString(vol + `\`)
	if err != nil {
		t.Fatal(err)
	}
	if err := windows.GetVolumeNameForVolumeMountPoint(mp, &buf[0], uint32(len(buf))); err != nil {
		t.Fatalf("GetVolumeNameForVolumeMountPoint(%s): %v", vol, err)
	}
	guid := strings.TrimSuffix(windows.UTF16ToString(buf), `\`) // \\?\Volume{...}
	refused(t, cfg, dir, guid+rest+`\protected.txt`)
	refused(t, cfg, dir, guid+rest+`\docs\guide.md`)
}

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
