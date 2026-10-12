package calc

import "testing"

// TestSubMore is new on the branch and fails at the base, where Sub adds.
func TestSubMore(t *testing.T) {
	if got := Sub(5, 3); got != 2 {
		t.Fatalf("Sub(5, 3) = %d, want 2", got)
	}
}
