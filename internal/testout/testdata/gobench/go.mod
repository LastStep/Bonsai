// A fixture project of package testout's tests (internal/testout/testdata/gobench). out.json is what
//
//	go test -json -run '^$|Fuzz' -bench . -benchtime 10x .
//
// printed here with go1.27.2 (since trimmed: its cpu line made neutral, Time fields dropped): a fuzz test's seed, two
// benchmarks that pass, which go test -json starts and never ends, and one that fails.
module example.com/bp

go 1.25
