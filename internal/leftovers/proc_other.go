//go:build !linux && !windows

package leftovers

import "os"

// Elsewhere (macOS and the BSDs build, untested): no reader, so the check says "not checked (<platform>)".
func snapshot() ([]Row, style, error) { return nil, posixStyle, errNotChecked }

func nativeStyle() style { return posixStyle }

func selfPID() int { return os.Getpid() }
