// Package bonsai exposes the embedded catalog and guide content used by the
// bonsai CLI. It exists so //go:embed directives stay at the repo root, where
// the embedded paths (catalog/, docs/custom-files.md) live.
package bonsai

import "embed"

//go:embed all:catalog
var CatalogFS embed.FS

//go:embed docs/custom-files.md
var GuideContent string
