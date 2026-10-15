//go:build unix

package agents

import "syscall"

// openFlags make a reader's open never wait on a FIFO and follow no link (regular.go).
const openFlags = syscall.O_NONBLOCK | syscall.O_NOFOLLOW
