package main

import (
	"embed"
	"fmt"
	"io/fs"
	"os"

	"github.com/LastStep/Bonsai/cmd"
)

//go:embed all:catalog
var catalogFS embed.FS

func main() {
	sub, err := fs.Sub(catalogFS, "catalog")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	cmd.Execute(sub)
}
