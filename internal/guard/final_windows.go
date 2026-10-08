//go:build windows

package guard

// The path Windows itself reaches, asked of Windows (paths.go's package comment says why the guard needs it).

import (
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

// GetFinalPathNameByHandleW is not in the syscall package. It is called through kernel32 (always loaded, and a
// known DLL, so no search path is involved) rather than through golang.org/x/sys/windows: importing that package
// alone added about 3 ms to every start of bonsai.exe (measured 8 Oct: 8.3 ms to 11.2 ms for an empty program),
// and the hook starts once per tool call.
var procGetFinalPathNameByHandle = syscall.NewLazyDLL("kernel32.dll").NewProc("GetFinalPathNameByHandleW")

// finalPath gives the path a write to p reaches on Windows: the longest part of p that exists, opened (following
// every junction, symbolic link and mount point) and named by GetFinalPathNameByHandle in its drive-letter form and
// its case on disk, then the rest of p as written. ok is false when no part of p opens. err is set when a part opens
// but Windows cannot give it a drive-letter name (a volume with none): the guard then cannot tell where p leads.
// It opens each part with no access asked for, so it reads and locks nothing.
func finalPath(p string) (final string, ok bool, err error) {
	var rest []string
	cur := p
	for i := 0; i < 256; i++ {
		name, opened, err := finalName(cur)
		if opened {
			if err != nil {
				return "", false, err
			}
			return filepath.Join(append([]string{name}, rest...)...), true, nil
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			break
		}
		rest = append([]string{filepath.Base(cur)}, rest...)
		cur = parent
	}
	return "", false, nil
}

// volumeNameDOS is GetFinalPathNameByHandle's flags for a drive-letter name, normalized: VOLUME_NAME_DOS (0) with
// FILE_NAME_NORMALIZED (0).
const volumeNameDOS = 0

// finalName opens p and gives Windows' own name for it. opened is false when p does not open.
func finalName(p string) (name string, opened bool, err error) {
	u, err := syscall.UTF16PtrFromString(p)
	if err != nil {
		return "", false, nil
	}
	h, err := syscall.CreateFile(u, 0, syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE|syscall.FILE_SHARE_DELETE,
		nil, syscall.OPEN_EXISTING, syscall.FILE_FLAG_BACKUP_SEMANTICS, 0)
	if err != nil {
		return "", false, nil
	}
	defer func() { _ = syscall.CloseHandle(h) }()
	buf := make([]uint16, 512)
	for {
		n, err := getFinalPathNameByHandle(h, buf, volumeNameDOS)
		if err != nil {
			return "", true, err
		}
		if int(n) < len(buf) {
			return dosPath(syscall.UTF16ToString(buf[:n])), true, nil
		}
		buf = make([]uint16, n+1)
	}
}

// getFinalPathNameByHandle calls GetFinalPathNameByHandleW: the name's length in buf, or, when buf is too short, the
// length it needs.
func getFinalPathNameByHandle(h syscall.Handle, buf []uint16, flags uint32) (uint32, error) {
	if err := procGetFinalPathNameByHandle.Find(); err != nil {
		return 0, err
	}
	r, _, e := procGetFinalPathNameByHandle.Call(uintptr(h), uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)), uintptr(flags))
	if r == 0 {
		if errno, ok := e.(syscall.Errno); ok && errno != 0 {
			return 0, errno
		}
		return 0, syscall.EINVAL
	}
	return uint32(r), nil
}

// dosPath drops the \\?\ prefix GetFinalPathNameByHandle puts on a drive path (\\?\C:\x is C:\x) or a share
// (\\?\UNC\server\share is \\server\share).
func dosPath(s string) string {
	switch {
	case strings.HasPrefix(s, `\\?\UNC\`):
		return `\\` + s[len(`\\?\UNC\`):]
	case strings.HasPrefix(s, `\\?\`):
		return s[len(`\\?\`):]
	}
	return s
}
