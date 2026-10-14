//go:build windows

package message

import "os"

// openNoWait opens a file to read. On Windows an open never waits for a writer, and a device's name is refused
// before it is opened (deviceIn); a device reached another way is refused by its open handle's Stat (GetFileType),
// before any read.
func openNoWait(path string) (*os.File, error) {
	return os.Open(path)
}
