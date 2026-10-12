// Package vetfail is a fixture of package testout's tests: a test file that go vet refuses.
package vetfail

import "fmt"

func ExampleHello() {
	fmt.Println("hello")
	// Output: hello
}
