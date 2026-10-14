// Command gen writes every file Bonsai's code generates (reference.Generated: the reference page of lists,
// docs/reference/lists.md, and base's generated skills under packs/base/skills/) from the code's tables and templates:
// `go generate ./...` runs it from internal/reference with the repository's root. It is a build tool, never part of
// the bonsai binary.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/LastStep/Bonsai/internal/reference"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: go run ./gen <the repository's root>")
		os.Exit(2)
	}
	for _, g := range reference.Generated() {
		p := filepath.Join(os.Args[1], filepath.FromSlash(g.Path))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			fmt.Fprintln(os.Stderr, "gen:", err)
			os.Exit(1)
		}
		if err := os.WriteFile(p, g.Build(), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, "gen:", err)
			os.Exit(1)
		}
	}
}
