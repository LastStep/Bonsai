// Package panics is a fixture of package testout's tests: a test that panics.
package panics

import "testing"

func TestBefore(t *testing.T) {}

func TestPanics(t *testing.T) { panic("boom") }
