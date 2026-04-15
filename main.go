package main

import (
	"embed"
	"fmt"
	"io/fs"
	"os"

	"github.com/LastStep/Bonsai/cmd"
)

// version is set via ldflags at build time.
var version = "dev"

//go:embed all:catalog
var catalogFS embed.FS

//go:embed docs/custom-files.md
var guideContent string

func main() {
	sub, err := fs.Sub(catalogFS, "catalog")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	cmd.SetVersion(version)
	cmd.Execute(sub, guideContent)
}
