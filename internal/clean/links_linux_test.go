package clean

import (
	"syscall"
	"testing"
	"time"
	"unsafe"
)

// lstamp sets a link's own modification time, not its target's (utimensat with AT_SYMLINK_NOFOLLOW, which Go's
// os.Chtimes does not offer), so a link decoy is as old as the files it sits among.
func lstamp(t *testing.T, p string, at time.Time) {
	t.Helper()
	name, err := syscall.BytePtrFromString(p)
	if err != nil {
		t.Fatal(err)
	}
	ts := [2]syscall.Timespec{syscall.NsecToTimespec(at.UnixNano()), syscall.NsecToTimespec(at.UnixNano())}
	dir := -100            // AT_FDCWD: a relative name is the working folder's (the name here is absolute)
	const noFollow = 0x100 // AT_SYMLINK_NOFOLLOW
	if _, _, e := syscall.Syscall6(syscall.SYS_UTIMENSAT, uintptr(dir), uintptr(unsafe.Pointer(name)),
		uintptr(unsafe.Pointer(&ts[0])), noFollow, 0, 0); e != 0 {
		t.Fatal(e)
	}
}
