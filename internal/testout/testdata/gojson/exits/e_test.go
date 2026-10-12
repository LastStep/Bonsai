// Package exits is a fixture of package testout's tests: a test that ends the test binary.
package exits

import (
	"os"
	"testing"
)

func TestExits(t *testing.T) { os.Exit(3) }
