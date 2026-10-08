// Command bonsai is Bonsai's one program. During the rebuild it is a stub: it answers `bonsai --version` and refuses
// everything else, so `go vet`, CodeQL's autobuild and GoReleaser find a main package at ./cmd/bonsai. The commands
// come with the walking skeleton's parts 2 to 5 (design/plan.md).
//
// Exit codes follow the spec's (design/bonsai-spec.md, section 3): 0 ok, 2 bad input, 3 when the answer cannot be
// written.
package main

import (
	"fmt"
	"io"
	"os"
)

// version is the build's version. GoReleaser sets it with `-ldflags "-X main.version=<version>"`
// (.goreleaser.yaml), as does `make build VERSION=<version>`; a plain `go build` leaves "dev".
var version = "dev"

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run answers one invocation and returns its exit code. Output is ASCII and names the next step on a refusal.
func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 1 && args[0] == "--version" {
		if _, err := fmt.Fprintf(stdout, "bonsai %s\n", version); err != nil {
			return 3
		}
		return 0
	}
	what := "no command given"
	switch {
	case len(args) > 0 && args[0] == "--version":
		what = "--version takes no other argument"
	case len(args) > 0:
		what = fmt.Sprintf("%+q is not a command yet", args[0]) // %+q keeps the line ASCII
	}
	// A refusal that cannot be written still exits 2: the code is the answer.
	_, _ = fmt.Fprintf(stderr, "bonsai: %s: this build is the rebuild's stub and answers only --version.\n"+
		"next: run `bonsai --version`, or use the old product at tag v0.4.3.\n", what)
	return 2
}
