// Package bad is a fixture of package testout's tests: a test file that does not build.
package bad

import "testing"

func TestX(t *testing.T) { undefined() }
