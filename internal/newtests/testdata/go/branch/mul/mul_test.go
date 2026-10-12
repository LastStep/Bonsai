package mul

import "testing"

// TestMul is new on the branch and passes at the base too: it proves nothing new.
func TestMul(t *testing.T) {
	if got := Mul(2, 3); got != 6 {
		t.Fatalf("Mul(2, 3) = %d, want 6", got)
	}
}
