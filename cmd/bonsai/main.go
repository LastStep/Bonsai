package main

import (
	"fmt"
	"io/fs"
	"os"

	bonsai "github.com/LastStep/Bonsai"
	"github.com/LastStep/Bonsai/cmd"
)

// version is set via ldflags at build time.
var version = "dev"

func main() {
	sub, err := fs.Sub(bonsai.CatalogFS, "catalog")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	cmd.SetVersion(version)
	cmd.Execute(sub, map[string]string{
		"quickstart":   bonsai.GuideQuickstart,
		"concepts":     bonsai.GuideConcepts,
		"cli":          bonsai.GuideCli,
		"custom-files": bonsai.GuideCustomFiles,
	})
}
