// Package good is a fixture of package testout's tests: passing, failing, skipped and nested tests.
package good

import (
	"fmt"
	"testing"
)

func TestPass(t *testing.T) {}

func TestFail(t *testing.T) { t.Errorf("want 1, got 2:\nmore detail") }

func TestSkipMarked(t *testing.T) { t.Skip("real-checkouts-only: needs a checkout") }

func TestSkipPlain(t *testing.T) { t.Skip("not on this machine") }

func TestParent(t *testing.T) {
	t.Run("sub one", func(t *testing.T) {})
	t.Run("sub/two", func(t *testing.T) { t.Fatal("sub failed") })
	t.Run("deep", func(t *testing.T) {
		t.Run("leaf (x)", func(t *testing.T) { t.Error("leaf failed") })
	})
}

func TestOwnMessage(t *testing.T) {
	t.Run("child", func(t *testing.T) { t.Error("child failed") })
	t.Error("parent failed too")
}

func Example_hello() {
	fmt.Println("hello")
	// Output: hello
}

func TestNoNewline(t *testing.T) {
	fmt.Print("no newline then fail")
	t.Fail()
}
