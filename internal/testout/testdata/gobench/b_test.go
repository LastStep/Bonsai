// Package bp is a fixture of package testout's tests: benchmarks, which go test -json starts and, unless they fail,
// never ends, and a fuzz test's seed.
package bp

import "testing"

func BenchmarkQuiet(b *testing.B) {
	for b.Loop() {
	}
}

func BenchmarkLogs(b *testing.B) {
	b.Log("hello")
	for b.Loop() {
	}
}

func BenchmarkFails(b *testing.B) {
	b.Fatal("the benchmark failed")
}

func FuzzF(f *testing.F) {
	f.Add(1)
	f.Fuzz(func(t *testing.T, n int) {})
}
