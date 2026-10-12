package calc

import "testing"

func TestAdd(t *testing.T) {
	if got := Add(2, 3); got != 5 {
		t.Fatalf("Add(2, 3) = %d, want 5", got)
	}
}

// TestSub gains a subtest on the branch, five, which fails at the base: its old parent then fails only by it.
func TestSub(t *testing.T) {
	t.Run("zero", func(t *testing.T) {
		if got := Sub(5, 0); got != 5 {
			t.Fatalf("Sub(5, 0) = %d, want 5", got)
		}
	})
	t.Run("five", func(t *testing.T) {
		if got := Sub(5, 3); got != 2 {
			t.Fatalf("Sub(5, 3) = %d, want 2", got)
		}
	})
}
