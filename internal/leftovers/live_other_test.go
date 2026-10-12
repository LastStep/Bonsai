//go:build !linux && !windows

package leftovers

// Elsewhere the live test only checks the value says not checked; nothing is started.
func namesDir(int, string) bool { return false }

func endPID(int) {}
