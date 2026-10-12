package neg

import "testing"

// TestNeg is new on the branch and calls code the branch adds, so its package does not build at the base.
func TestNeg(t *testing.T) {
	if got := Neg(2); got != -2 {
		t.Fatalf("Neg(2) = %d, want -2", got)
	}
}
