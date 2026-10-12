// A fixture project of package testout's tests (internal/testout/testdata/gojson). out.json is what
//
//	echo "building the demo" && go test -json -timeout 3s ./...
//
// printed here with go1.27.2 (since trimmed: home-folder paths made neutral, Time fields dropped): one line that is
// not JSON, then passing, failing, skipped, nested and panicking tests, a package that does not build, one that fails
// vet, one failing in TestMain, one with no test files, a test that calls os.Exit and one that times out. The run test
// runs single tests of package good by the commands package testout gives.
module example.com/demo

go 1.22
