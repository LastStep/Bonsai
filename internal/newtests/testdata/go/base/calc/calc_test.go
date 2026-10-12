package calc

import "testing"

func TestAdd(t *testing.T) {
	if got := Add(2, 3); got != 5 {
		t.Fatalf("Add(2, 3) = %d, want 5", got)
	}
}

func TestSub(t *testing.T) {
	t.Run("zero", func(t *testing.T) {
		if got := Sub(5, 0); got != 5 {
			t.Fatalf("Sub(5, 0) = %d, want 5", got)
		}
	})
}
