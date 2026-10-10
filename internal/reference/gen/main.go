// Command gen writes the reference page of lists (docs/reference/lists.md) from the code's tables: `go generate ./...`
// runs it from internal/reference with the page's path. It is a build tool, never part of the bonsai binary.
package main

import (
	"fmt"
	"os"

	"github.com/LastStep/Bonsai/internal/reference"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: go run ./gen <path of the page>")
		os.Exit(2)
	}
	if err := os.WriteFile(os.Args[1], reference.Page(), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "gen:", err)
		os.Exit(1)
	}
}
