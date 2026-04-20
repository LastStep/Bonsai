// Package bonsai exposes the embedded catalog and guide content used by the
// bonsai CLI. It exists so //go:embed directives stay at the repo root, where
// the embedded paths (catalog/, docs/*.md) live.
package bonsai

import "embed"

//go:embed all:catalog
var CatalogFS embed.FS

//go:embed docs/custom-files.md
var GuideCustomFiles string

//go:embed docs/quickstart.md
var GuideQuickstart string

//go:embed docs/concepts.md
var GuideConcepts string

//go:embed docs/cli.md
var GuideCli string
