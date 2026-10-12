// Package mainfail is a fixture of package testout's tests: TestMain fails before any test runs.
package mainfail

import (
	"fmt"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	fmt.Println("setup: cannot open the fixture")
	os.Exit(1)
}

func TestNever(t *testing.T) {}
