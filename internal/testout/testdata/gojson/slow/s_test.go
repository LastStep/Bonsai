// Package slow is a fixture of package testout's tests: a test that outlives the -timeout.
package slow

import (
	"testing"
	"time"
)

func TestSlow(t *testing.T) { time.Sleep(30 * time.Second) }
