// Package selfcheck holds checks of Bonsai's own repository files, run by every climb with the rest of the tests. Its
// only code is the fault markers' one home (markers.go), which cmd/bonsai's tests read too: each check is a test in
// this package that reads the files from the module root, or the release files TestDist is pointed at, and fails
// naming the file and line.
package selfcheck
