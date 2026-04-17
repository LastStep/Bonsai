# Plan 16 — Fix `go install` binary name

**Tier:** 1 (Patch)
**Status:** Active
**Agent:** general-purpose
**Source:** `Playbook/Backlog.md` P1 bug (added 2026-04-17)

---

## Goal

After this plan lands, `go install github.com/LastStep/Bonsai/cmd/bonsai@latest` produces a `bonsai` binary (lowercase) on `$PATH` — fixing the release-blocker where the current `go install github.com/LastStep/Bonsai@latest` names the binary `Bonsai` (capital), which Linux's case-sensitive PATH resolution can't find.

## Context

Go derives the `go install` binary name from the last segment of the install path. The module is `github.com/LastStep/Bonsai`, so `go install github.com/LastStep/Bonsai@latest` → `$GOPATH/bin/Bonsai`. The `Makefile` dodges this with `-o bonsai`, but `go install` has no such override.

**Chosen fix (from backlog Option 2):** move `main.go` to `cmd/bonsai/main.go` so the install path ends in `/bonsai`. Non-breaking — module path stays `github.com/LastStep/Bonsai`.

**Embed constraint:** `//go:embed` directives cannot use `..` in paths. The current `main.go` embeds `catalog/` and `docs/custom-files.md` from the repo root. Moving `main.go` into `cmd/bonsai/` would break these embeds (the target paths live above `cmd/bonsai/`). Fix: introduce a thin root library package (`embed.go`) that declares the embed vars and is imported by `cmd/bonsai/main.go`.

## Steps

### Step 1 — Create root library package `embed.go`

Create new file `embed.go` at the repo root:

```go
// Package bonsai exposes the embedded catalog and guide content used by the
// bonsai CLI. It exists so //go:embed directives stay at the repo root, where
// the embedded paths (catalog/, docs/custom-files.md) live.
package bonsai

import "embed"

//go:embed all:catalog
var CatalogFS embed.FS

//go:embed docs/custom-files.md
var GuideContent string
```

- Package name: `bonsai` (matches module import path final segment, case-folded)
- Exports: `CatalogFS` (embed.FS) and `GuideContent` (string)
- No other declarations — keep this file minimal

### Step 2 — Create `cmd/bonsai/main.go`

Create directory `cmd/bonsai/` and new file `cmd/bonsai/main.go`:

```go
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
	cmd.Execute(sub, bonsai.GuideContent)
}
```

- Import the root library package with alias `bonsai` (import path has capital `B`, Go identifier needs a valid one)
- `var version = "dev"` stays in `package main` so `-X main.version=...` ldflag still works
- Behavior identical to the current `main.go`

### Step 3 — Delete the old root `main.go`

`git rm main.go` — the file's responsibilities are now split between `embed.go` (at root) and `cmd/bonsai/main.go`.

### Step 4 — Update `Makefile`

Change the two build lines to target `./cmd/bonsai`:

```makefile
build:
	go build -ldflags "$(LDFLAGS)" -o bonsai ./cmd/bonsai

install:
	go install -ldflags "$(LDFLAGS)" ./cmd/bonsai
```

`LDFLAGS`, `clean`, and `VERSION` stay unchanged. The `-X main.version=$(VERSION)` flag still targets the `main` package — now `cmd/bonsai`.

### Step 5 — Update `.goreleaser.yaml`

Change `main: .` to `main: ./cmd/bonsai` in the `builds` section (line 7):

```yaml
builds:
  - id: bonsai
    main: ./cmd/bonsai
    binary: bonsai
    ...
```

`binary: bonsai`, ldflags `-X main.version={{.Version}}`, Homebrew config, archive naming — all unchanged.

### Step 6 — Update `README.md`

Replace the `go install` command at line 68:

```bash
go install github.com/LastStep/Bonsai/cmd/bonsai@latest
```

(Adds `/cmd/bonsai` to the module path.)

### Step 7 — Update repo-root `CLAUDE.md`

Replace line 118 in `/CLAUDE.md` (the repo-root CLAUDE.md, NOT `station/CLAUDE.md`):

```bash
go install ./cmd/bonsai    # install to $GOPATH/bin
```

### Out of scope

- Do NOT modify `station/` (Tech Lead workspace) — tracking-only updates happen after merge.
- Do NOT modify `.github/workflows/ci.yml` or `.github/workflows/release.yml` — they use `go-version-file: go.mod` and `go test ./...`, both unaffected by the path move.
- Do NOT rename the existing `cmd/` package (containing Cobra commands). It stays at `github.com/LastStep/Bonsai/cmd`. The new `cmd/bonsai/` is a sibling subdirectory.
- Do NOT change `go.mod` module path.
- Do NOT move `catalog/` or `docs/` — they stay at the repo root.

## Security

> [!warning]
> Refer to `station/Playbook/Standards/SecurityStandards.md` for all security requirements.

This plan introduces no new attack surface. No new dependencies, no user input changes, no file-permission changes. Embedded data paths (`catalog/`, `docs/custom-files.md`) remain identical — just declared from a different file.

## Verification

Run from the repo root:

- [ ] `make build` succeeds and produces `./bonsai` (lowercase) binary
- [ ] `./bonsai --version` prints a version string (confirms ldflags still wire up)
- [ ] `./bonsai --help` lists all existing subcommands (confirms `cmd/` package still wires up)
- [ ] `go test ./...` passes (no regressions in `internal/generate` or elsewhere)
- [ ] `go vet ./...` clean
- [ ] `go install ./cmd/bonsai` installs to `$GOPATH/bin/bonsai` (lowercase); verify `which bonsai` resolves
- [ ] `goreleaser build --snapshot --clean --single-target` succeeds (cross-build sanity — skip if goreleaser not installed locally; CI will catch it on next tag)
- [ ] `git grep -n "go install .* Bonsai@latest"` returns no matches (no stale uppercase install instruction in tracked files)
- [ ] Root `main.go` is deleted; `embed.go` exists at root; `cmd/bonsai/main.go` exists
