package guard

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"unsafe"

	"github.com/LastStep/Bonsai/internal/hookin"
)

// kernel32's procedures these tests use, which the syscall package does not wrap.
var (
	kernel32                 = syscall.NewLazyDLL("kernel32.dll")
	procQueryDosDevice       = kernel32.NewProc("QueryDosDeviceW")
	procGetVolumeNameForRoot = kernel32.NewProc("GetVolumeNameForVolumeMountPointW")
)

// call runs a kernel32 procedure that returns 0 on failure, and gives its result.
func call(t *testing.T, p *syscall.LazyProc, args ...uintptr) uintptr {
	t.Helper()
	r, _, err := p.Call(args...)
	if r == 0 {
		t.Fatalf("%s: %v", p.Name, err)
	}
	return r
}

// windowsProject is project with a protected folder docs holding guide.md.
func windowsProject(t *testing.T) string {
	t.Helper()
	dir := project(t, testYAML)
	if err := os.MkdirAll(filepath.Join(dir, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "docs", "guide.md"), []byte("guide\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

// refused holds a Write of p, in a session of the project at root, to a refusal.
func refused(t *testing.T, root, p string) {
	t.Helper()
	d := decide(&hookin.Payload{Tool: "Write", Path: p}, root)
	if d.Allow {
		t.Errorf("%s allowed (rule %s), want refused", p, d.Rule)
	}
}

// A stream in a folder's name is the folder: secrets::$INDEX_ALLOCATION\key.txt writes secrets\key.txt.
func TestDecideReadsDirectoryStreams(t *testing.T) {
	dir := windowsProject(t)
	for _, p := range []string{
		dir + `\docs::$INDEX_ALLOCATION\guide.md`,
		dir + `\docs::$INDEX_ALLOCATION\new.md`,
		dir + `\docs:$I30:$INDEX_ALLOCATION\guide.md`,
		dir + `\docs:$I30:$INDEX_ALLOCATION\new.md`,
	} {
		refused(t, dir, p)
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
	dir := windowsProject(t)
	outside, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	mklinkJ(t, filepath.Join(outside, "j"), filepath.Join(dir, "docs"))
	mklinkJ(t, filepath.Join(outside, "proj"), dir)
	refused(t, dir, filepath.Join(outside, "j", "guide.md"))
	refused(t, dir, filepath.Join(outside, "j", "new.md"))
	refused(t, dir, filepath.Join(outside, "proj", "protected.txt"))
	// The project reached through a junction, the path by its real folder.
	refused(t, filepath.Join(outside, "proj"), filepath.Join(dir, "protected.txt"))
	if d := decide(&hookin.Payload{Tool: "Write", Path: filepath.Join(outside, "j2.txt")}, dir); !d.Allow {
		t.Errorf("a file beside the junctions refused: %+v", d)
	}
}

// Device and volume forms of a protected path: \\?\GLOBALROOT\Device\HarddiskVolumeN\... and \\?\Volume{...}\....
func TestDecideReadsDevicePaths(t *testing.T) {
	dir := windowsProject(t)
	vol := filepath.VolumeName(dir) // C:
	vp, err := syscall.UTF16PtrFromString(vol)
	if err != nil {
		t.Fatal(err)
	}
	buf := make([]uint16, 1024)
	n := call(t, procQueryDosDevice, uintptr(unsafe.Pointer(vp)), uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	device := syscall.UTF16ToString(buf[:n]) // \Device\HarddiskVolumeN
	if device == "" {
		// BUG-112: an empty device would build a path with no device, which the guard rightly reads as outside the project.
		t.Skipf("QueryDosDevice(%q) gave an empty device: the test cannot build a device path (setup, not the guard)", vol)
	}
	rest := dir[len(vol):]                   // \Users\...
	refused(t, dir, `\\?\GLOBALROOT`+device+rest+`\protected.txt`)
	refused(t, dir, `\\?\GLOBALROOT`+device+rest+`\docs\new.md`)
	mp, err := syscall.UTF16PtrFromString(vol + `\`)
	if err != nil {
		t.Fatal(err)
	}
	call(t, procGetVolumeNameForRoot, uintptr(unsafe.Pointer(mp)), uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	guid := strings.TrimSuffix(syscall.UTF16ToString(buf), `\`) // \\?\Volume{...}
	refused(t, dir, guid+rest+`\protected.txt`)
	refused(t, dir, guid+rest+`\docs\guide.md`)
}

// An NTFS short (8.3) name of a protected file is the protected file.
func TestDecideReadsShortNames(t *testing.T) {
	dir := project(t, testYAML)
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
	d := decide(&hookin.Payload{Tool: "Write", Path: short}, dir)
	if d.Allow || d.Rule != RuleProtected {
		t.Fatalf("%s: %+v, want refused as protected", short, d)
	}
}

// Windows only (junctions are Windows'; mklink /J needs no privilege): a junction inside the project to a protected
// folder is followed; one past 260 characters too; a dangling junction (its target gone) on the way to a path is
// refused as bad-path, since the guard cannot tell where it leads and Windows would fail the write anyway (gate
// report section 5); a junction to a free place stays free.
func TestJunctionLimits(t *testing.T) {
	dir := windowsProject(t)
	outside, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	// Inside the project, to the protected folder.
	mklinkJ(t, filepath.Join(dir, "free-j"), filepath.Join(dir, "docs"))
	refused(t, dir, filepath.Join(dir, "free-j", "guide.md"))
	refused(t, dir, filepath.Join(dir, "free-j", "new.md"))
	// Past 260 characters: a long tail under a junction to the protected folder.
	mklinkJ(t, filepath.Join(outside, "j"), filepath.Join(dir, "docs"))
	long := filepath.Join(outside, "j", strings.Repeat("a", 120), strings.Repeat("b", 120), "new.md")
	if len(long) <= 260 {
		t.Fatalf("the path is %d characters, not past 260", len(long))
	}
	refused(t, dir, long)
	// Dangling: a junction to a protected folder made, then its target removed.
	gone := filepath.Join(dir, "docs", "gone")
	if err := os.MkdirAll(gone, 0o755); err != nil {
		t.Fatal(err)
	}
	mklinkJ(t, filepath.Join(outside, "dangling"), gone)
	if err := os.Remove(gone); err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{filepath.Join(outside, "dangling", "new.md"), filepath.Join(outside, "dangling")} {
		d := decide(&hookin.Payload{Tool: "Write", Path: p}, dir)
		if d.Allow || d.Rule != RuleBadPath || !strings.Contains(d.Why, "leads nowhere that exists") {
			t.Errorf("%s: %+v, want refused as bad-path", p, d)
		}
	}
	// To a free place outside: free.
	free := filepath.Join(outside, "free")
	if err := os.MkdirAll(free, 0o755); err != nil {
		t.Fatal(err)
	}
	mklinkJ(t, filepath.Join(dir, "to-free"), free)
	if d := decide(&hookin.Payload{Tool: "Write", Path: filepath.Join(dir, "to-free", "x.txt")}, dir); !d.Allow {
		t.Errorf("a junction to a free place: %+v", d)
	}
}

// Windows only (administrative shares are Windows'): a path through this machine's own share of the drive, by
// localhost, a loopback address or the machine's name, is judged as the drive's path: a protected one refused, a free
// one allowed (gate report section 5; 5.3.1 note 6). It holds whether or not the share opens for this user: the forms
// are read from the string.
func TestAdminShare(t *testing.T) {
	dir := windowsProject(t)
	vol := filepath.VolumeName(dir) // C:
	rest := dir[len(vol):]          // \Users\...
	share := strings.TrimSuffix(vol, ":") + "$"
	host, err := os.Hostname()
	if err != nil {
		t.Fatal(err)
	}
	for _, server := range []string{"localhost", "127.0.0.1", host, strings.ToUpper(host)} {
		for _, form := range []string{`\\` + server + `\` + share + rest, `\\?\UNC\` + server + `\` + share + rest} {
			// Refused as the drive's path is: protected, in the project this machine recorded (not a second project
			// found through the share, which no record names).
			for _, p := range []string{form + `\protected.txt`, form + `\docs\new.md`} {
				if d := decide(&hookin.Payload{Tool: "Write", Path: p}, dir); d.Allow || d.Rule != RuleProtected {
					t.Errorf("%s: %+v, want refused as protected", p, d)
				}
			}
			if d := decide(&hookin.Payload{Tool: "Write", Path: form + `\free.txt`}, dir); !d.Allow {
				t.Errorf("%s: %+v, want allowed", form+`\free.txt`, d)
			}
		}
	}
}
