# Plan 04 — Release Pipeline

**Tier:** 2 (Feature)
**Status:** Complete
**Agent:** tech-lead
**Source:** Backlog P1 — "Release pipeline — GoReleaser + GitHub Actions + Homebrew Tap"

## Goal

Bonsai has an automated release pipeline: push a semver tag → GoReleaser builds cross-platform binaries → GitHub Releases publishes them → Homebrew Tap formula is updated. Users can `brew install LastStep/tap/bonsai` or download binaries directly.

## Context

Bonsai currently has no version management, no CI/CD, no release automation. The only install path is `go install github.com/LastStep/Bonsai@latest`, which requires Go toolchain. A release pipeline enables binary distribution to users without Go, and is the prerequisite for public launch.

## Steps

### Step 1: Add version variable and `--version` flag

**File:** `cmd/root.go`

1. Add a package-level variable: `var Version = "dev"`
2. Add a `SetVersion` function that sets `rootCmd.Version`:
   ```go
   func SetVersion(v string) {
       rootCmd.Version = v
   }
   ```
3. Cobra provides `--version` flag automatically when `rootCmd.Version` is set.

**File:** `main.go`

4. Add a package-level variable: `var version = "dev"` (set via ldflags at build time)
5. Call `cmd.SetVersion(version)` before `cmd.Execute(sub, guideContent)`

**File:** `Makefile`

6. Update the `build` target to inject version via ldflags:
   ```makefile
   VERSION ?= dev
   LDFLAGS := -s -w -X main.version=$(VERSION)

   build:
   	go build -ldflags "$(LDFLAGS)" -o bonsai .
   ```
7. Update the `install` target similarly:
   ```makefile
   install:
   	go install -ldflags "$(LDFLAGS)" .
   ```

### Step 2: Create GoReleaser config

**File:** `.goreleaser.yaml` (new, project root)

```yaml
version: 2

project_name: bonsai

builds:
  - id: bonsai
    main: .
    binary: bonsai
    env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w -X main.version={{.Version}}
    mod_timestamp: "{{ .CommitTimestamp }}"

archives:
  - id: default
    formats:
      - tar.gz
    format_overrides:
      - goos: windows
        formats:
          - zip
    name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"
    files:
      - LICENSE*
      - README*

checksum:
  name_template: "checksums.txt"

changelog:
  sort: asc
  filters:
    exclude:
      - "^docs:"
      - "^test:"
      - "^chore:"
      - "^ci:"

brews:
  - name: bonsai
    repository:
      owner: LastStep
      name: homebrew-tap
      token: "{{ .Env.HOMEBREW_TAP_TOKEN }}"
    homepage: "https://github.com/LastStep/Bonsai"
    description: "CLI tool for scaffolding Claude Code agent workspaces"
    license: "MIT"
    directory: Formula
    install: |
      bin.install "bonsai"
    test: |
      system "#{bin}/bonsai", "--version"
```

### Step 3: Create GitHub Actions release workflow

**File:** `.github/workflows/release.yml` (new)

```yaml
name: Release

on:
  push:
    tags:
      - "v*"

permissions:
  contents: write

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version-file: go.mod

      - name: Run tests
        run: go test ./...

      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v6
        with:
          distribution: goreleaser
          version: "~> v2"
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          HOMEBREW_TAP_TOKEN: ${{ secrets.HOMEBREW_TAP_TOKEN }}
```

### Step 4: Add `.goreleaser.yaml` to `.gitignore` exclusions

Check if `.gitignore` exists. If so, ensure `dist/` (GoReleaser output dir) is ignored:

**File:** `.gitignore`

Add `dist/` if not already present (GoReleaser uses `dist/` for build artifacts during `goreleaser release --clean`).

## Dependencies

- **Tap repo:** User must create `LastStep/homebrew-tap` on GitHub (empty repo with a README)
- **Secret:** User must create a GitHub PAT with `repo` scope and add it as `HOMEBREW_TAP_TOKEN` repository secret on `LastStep/Bonsai`
- These are manual one-time setup steps — the pipeline code doesn't depend on them to build/test

## Security

> Refer to `Playbook/Standards/SecurityStandards.md`

- No secrets hardcoded — `GITHUB_TOKEN` is auto-provided by Actions, `HOMEBREW_TAP_TOKEN` comes from repository secrets
- `CGO_ENABLED=0` ensures static binaries with no C dependencies
- Checksum file generated for download verification
- `permissions: contents: write` is minimum required scope — no broader permissions
- No `.env` files, no credentials in config

## Verification

- [ ] `make build` passes with version ldflags
- [ ] `./bonsai --version` prints the version string
- [ ] `go test ./...` passes
- [ ] `goreleaser check` validates `.goreleaser.yaml` (if goreleaser is installed locally)
- [ ] `goreleaser build --single-target --snapshot --clean` builds successfully (if goreleaser is installed locally)
- [ ] `.github/workflows/release.yml` is valid YAML
- [ ] `dist/` is in `.gitignore`
