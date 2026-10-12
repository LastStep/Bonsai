// Package neg is a fixture of the new-tests check: the branch adds Neg and its test, which does not build at the base.
package neg

// Abs is the absolute value.
func Abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}
