# Plan 02 — `bonsai guide` Command

**Tier:** 1 (Patch)
**Status:** Active
**Agent:** tech-lead (single domain: cli)

## Goal

Add a `bonsai guide` command that renders `docs/custom-files.md` as styled terminal output, so users can read the custom files guide without leaving the CLI.

## Steps

### 1. Embed the guide content in `main.go`

**File:** `main.go`

Add a separate embed directive for the docs directory. Add a new `docsFS` variable and pass it to `cmd.Execute()`.

**Current code:**
```go
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
```

**Replace with:**
```go
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
	cmd.Execute(sub, guideContent)
}
```

Using `//go:embed` as a string for the single file — cleaner than a full FS for one file.

### 2. Update `cmd.Execute()` to accept guide content

**File:** `cmd/root.go`

Add a module-level `guideMarkdown` variable. Update the `Execute` function signature:

**Current:**
```go
var catalogFS fs.FS

func Execute(fsys fs.FS) {
	catalogFS = fsys
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
```

**Replace with:**
```go
var catalogFS fs.FS
var guideMarkdown string

func Execute(fsys fs.FS, guide string) {
	catalogFS = fsys
	guideMarkdown = guide
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
```

### 3. Add `glamour` dependency

Run:
```bash
go get github.com/charmbracelet/glamour
```

This is from the same Charm ecosystem (LipGloss, Huh, BubbleTea) already used in the project.

### 4. Create `cmd/guide.go`

**File:** `cmd/guide.go` (new file)

```go
package cmd

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(guideCmd)
}

var guideCmd = &cobra.Command{
	Use:   "guide",
	Short: "View the custom files guide.",
	Long:  "Display the guide for creating custom skills, workflows, protocols, sensors, and routines.",
	RunE:  runGuide,
}

func runGuide(cmd *cobra.Command, args []string) error {
	content := guideMarkdown

	// Strip YAML frontmatter if present
	if strings.HasPrefix(content, "---") {
		if idx := strings.Index(content[3:], "---"); idx >= 0 {
			content = strings.TrimSpace(content[idx+6:])
		}
	}

	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(100),
	)
	if err != nil {
		return fmt.Errorf("failed to create renderer: %w", err)
	}

	out, err := renderer.Render(content)
	if err != nil {
		return fmt.Errorf("failed to render guide: %w", err)
	}

	fmt.Print(out)
	return nil
}
```

**Key details:**
- Strips YAML frontmatter (the guide doesn't have any currently, but defensive)
- Uses `glamour.WithAutoStyle()` for dark/light terminal detection
- Uses `glamour.WithWordWrap(100)` for readable line lengths
- No flags needed — single-purpose command

### 5. Verification

- `make build` — must compile with no errors
- `go test ./...` — all existing tests pass
- `./bonsai guide` — renders the custom files guide with styled headings, tables, code blocks
- `./bonsai --help` — `guide` appears in the command list
- `./bonsai guide --help` — shows the Long description

## Security

> Refer to `Playbook/Standards/SecurityStandards.md` — no security-sensitive changes in this patch. The guide content is embedded at build time from a local file; no user input, no file system access at runtime.

## Verification

- [ ] `make build` passes
- [ ] `go test ./...` passes
- [ ] `./bonsai guide` renders styled markdown in terminal
- [ ] `./bonsai --help` lists the `guide` command
- [ ] Guide content matches `docs/custom-files.md` (minus frontmatter)
