package neg

import "testing"

func TestAbs(t *testing.T) {
	if got := Abs(-2); got != 2 {
		t.Fatalf("Abs(-2) = %d, want 2", got)
	}
}
