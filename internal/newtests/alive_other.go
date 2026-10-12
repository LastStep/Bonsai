//go:build !unix && !windows

package newtests

// alive cannot tell here, so every owner counts as running and no stale worktree is swept.
func alive(int) bool { return true }
