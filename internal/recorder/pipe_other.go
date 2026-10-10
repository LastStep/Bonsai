//go:build !unix

package recorder

// ignoreBrokenPipe does nothing here: Windows has no SIGPIPE, and a write to a closed pipe is already an error the
// hook passes over.
func ignoreBrokenPipe() {}
