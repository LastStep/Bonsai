// Command gen writes every file Bonsai's code generates (reference.Generated: the reference page of lists,
// docs/reference/lists.md, and base's generated skills under packs/base/skills/) from the code's tables and templates:
// `go generate ./...` runs it from internal/reference with the repository's root. It is a build tool, never part of
// the bonsai binary.
//
// It takes one argument, the repository's root, a folder that is there: anything else (no argument, more than one,
// a flag such as -h, or a path that is not a folder) prints the usage and exits 2, writing nothing.
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/LastStep/Bonsai/internal/reference"
)

const usage = "usage: go run ./gen <the repository's root, a folder>"

func main() { os.Exit(run(os.Args[1:], os.Stderr)) }

// run writes the generated files under the root args names and gives the exit code: 0 written, 1 a write failed,
// 2 the arguments refused (nothing written).
func run(args []string, stderr io.Writer) int {
	if len(args) != 1 || args[0] == "" || strings.HasPrefix(args[0], "-") {
		_, _ = fmt.Fprintln(stderr, usage)
		return 2
	}
	root := args[0]
	if fi, err := os.Stat(root); err != nil || !fi.IsDir() {
		_, _ = fmt.Fprintf(stderr, "gen: %s is not a folder\n%s\n", filepath.ToSlash(root), usage)
		return 2
	}
	for _, g := range reference.Generated() {
		p := filepath.Join(root, filepath.FromSlash(g.Path))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			_, _ = fmt.Fprintln(stderr, "gen:", err)
			return 1
		}
		if err := os.WriteFile(p, g.Build(), 0o644); err != nil {
			_, _ = fmt.Fprintln(stderr, "gen:", err)
			return 1
		}
	}
	return 0
}
