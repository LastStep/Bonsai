//go:build unix

package recorder

import (
	"os/signal"
	"syscall"
)

// ignoreBrokenPipe makes a write to a pipe whose reader is gone an error the hook passes over, instead of the
// process's end: by default a Go program writing to a closed pipe on its stdout or stderr dies by SIGPIPE (signal 13),
// which would break the hooks' promise to exit 0 whatever happens (bonsai hook start prints its context after its
// record is written; Claude Code may have closed the pipe by then).
func ignoreBrokenPipe() { signal.Ignore(syscall.SIGPIPE) }
