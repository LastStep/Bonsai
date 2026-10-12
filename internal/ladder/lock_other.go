//go:build !unix && !windows

package ladder

// Elsewhere (no flock, no LockFileEx; Bonsai is built for Linux, macOS and Windows) the lock is not held: every try
// takes it, and the holder's words are still written.

import "os"

func tryLock(*os.File) (bool, error) { return true, nil }

func unlock(*os.File) {}

func alive(int) bool { return false }
