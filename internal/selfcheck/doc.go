// Package selfcheck holds checks of Bonsai's own repository files, run by every climb with the rest of the tests. It
// has no code of its own: each check is a test in this package that reads the files from the module root and fails
// naming the file and line.
package selfcheck
