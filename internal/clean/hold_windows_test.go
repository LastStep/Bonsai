package clean

import (
	"syscall"
	"testing"
)

// holdFile opens path with no sharing at all, as an editor or a scanner may: a delete of it is refused with a sharing
// violation until release.
func holdFile(t *testing.T, path string) (release func()) {
	t.Helper()
	p, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		t.Fatal(err)
	}
	h, err := syscall.CreateFile(p, syscall.GENERIC_READ|syscall.GENERIC_WRITE, 0, nil, syscall.OPEN_EXISTING,
		syscall.FILE_ATTRIBUTE_NORMAL, 0)
	if err != nil {
		t.Fatal(err)
	}
	return func() { _ = syscall.CloseHandle(h) }
}
