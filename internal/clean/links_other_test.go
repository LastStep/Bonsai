//go:build !linux

package clean

import (
	"testing"
	"time"
)

// lstamp leaves a link's own time as it is off Linux: the fixture's links are Linux's (TestEveryKind says why).
func lstamp(*testing.T, string, time.Time) {}
