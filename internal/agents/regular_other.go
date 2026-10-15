//go:build !unix

package agents

// openFlags are none where the file system holds no FIFO or device and the first look sees every link (regular.go).
const openFlags = 0
