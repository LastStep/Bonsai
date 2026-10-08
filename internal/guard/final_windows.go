//go:build windows

package guard

// The path Windows itself reaches, asked of Windows (paths.go's package comment says why the guard needs it).

import (
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows"
)

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
// FILE_NAME_NORMALIZED (0). golang.org/x/sys/windows does not name them.
const volumeNameDOS = 0

// finalName opens p and gives Windows' own name for it. opened is false when p does not open.
func finalName(p string) (name string, opened bool, err error) {
	u, err := windows.UTF16PtrFromString(p)
	if err != nil {
		return "", false, nil
	}
	h, err := windows.CreateFile(u, 0, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil, windows.OPEN_EXISTING, windows.FILE_FLAG_BACKUP_SEMANTICS, 0)
	if err != nil {
		return "", false, nil
	}
	defer func() { _ = windows.CloseHandle(h) }()
	buf := make([]uint16, 512)
	for {
		n, err := windows.GetFinalPathNameByHandle(h, &buf[0], uint32(len(buf)), volumeNameDOS)
		if err != nil {
			return "", true, err
		}
		if int(n) < len(buf) {
			return dosPath(windows.UTF16ToString(buf[:n])), true, nil
		}
		buf = make([]uint16, n+1)
	}
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
