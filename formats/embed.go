// Package formats holds Bonsai's formats set: one JSON Schema per format, one example document per format, the
// trick files with their expected outcomes, and a manifest of every file's bytes (README.md says what each is).
//
// This file is the set's only Go code outside its tests: it embeds the schemas so Bonsai's code reads them from here
// and keeps no second copy of any list they hold (contract §2.2; plan parts 0 and 2). Go's embed cannot reach a
// parent folder, so the embedding lives beside the schemas. Go files are outside the manifest.
package formats

import (
	"embed"
	"io/fs"
)

// schemas holds formats/schemas/<name>.schema.json, byte for byte as committed.
//
//go:embed schemas/*.schema.json
var schemas embed.FS

// Names lists the formats the set has a schema for, in the order README.md gives them: the ten of contract §2, then
// the eight set 4 added (bonsai.yaml, pack.yaml, the two tables, memory, the error object, and the --json of check
// and of init, update and unlink).
var Names = []string{"task", "labels", "lanes", "run", "state", "log", "ask", "ladder", "status", "lock",
	"workspace", "pack", "tasks", "sessions", "memory", "error", "check", "changes"}

// Schema returns the committed bytes of one format's schema, by its short name ("lock", "status").
func Schema(name string) ([]byte, error) {
	return schemas.ReadFile("schemas/" + name + ".schema.json")
}

// Schemas gives the embedded schema files, under schemas/, for a test that checks them against the folder.
func Schemas() fs.FS {
	return schemas
}
