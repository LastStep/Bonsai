# Plan 20 — Security Scanning Infrastructure

**Tier:** 2
**Status:** Complete
**Agent:** general-purpose (per-PR worktree dispatch)
**Shipped:** 2026-04-21
**Merged PRs:** #30 (pre-flight), #29 (lint v1→v2), #28 (Go 1.25.8 + pin bump), #31 (Dependabot), #40 (govulncheck), #41 (CodeQL)
**gitleaks history audit:** 0 findings across 156 commits

## Goal

Stand up layered automated security scanning for the Bonsai repo — Go toolchain bump (clears 2 reachable stdlib CVEs), Dependabot (dep CVE auto-PRs), govulncheck CI step (per-PR reachability-aware Go scan), CodeQL workflow (SAST). Ship as 4 sequential, individually-mergeable PRs. Close with a gitleaks one-shot history audit.

## Context

**Current state:**
- CI runs `go test`, `go vet`, `golangci-lint` on PR.
- GitGuardian GitHub App integration covers PR secret scanning (per memory).
- No Dependabot, no CodeQL, no govulncheck in CI.
- Last vulnerability-scan routine (2026-04-21) flagged 2 reachable stdlib CVEs on Go 1.24.13 (GO-2026-4602, GO-2026-4601). Both cleared by Go 1.25.8+.
- User has enabled GH Secret Scanning + Push Protection + Dependabot alerts + Private vulnerability reporting in repo Settings → Code security. Branch protection on `main` already configured.

**Why now:** pre-launch security posture for OSS release. No scanner covers all four threat categories (dep CVEs, code SAST, secrets, stdlib CVEs), so layered stack is required. Automating scans closes the 7-day routine feedback gap.

**Scope boundary:** This plan only adds scanning infra. It does NOT attempt to fix any new findings surfaced by the new scanners — those become follow-up tickets in Backlog.md.

## Steps

Each PR is dispatched to a general-purpose agent in an isolated worktree. Each PR branches off fresh `main` after the previous merge.

### PR #1 — Go toolchain bump 1.24.13 → 1.25.8

**Branch:** `security/go-toolchain-bump`

**Files:**
- `go.mod` — change `go 1.24.2` → `go 1.25.0`; change `toolchain go1.24.13` → `toolchain go1.25.8`

**Agent steps:**
1. Edit `go.mod` top directives as above.
2. Run `go mod tidy` (toolchain auto-downloads Go 1.25.8 via `GOTOOLCHAIN=auto` default).
3. Run `make build` — verify clean build.
4. Run `go test ./...` — all tests pass.
5. Run `go vet ./...` — no warnings.
6. If `go.sum` changes during `go mod tidy`, commit those changes alongside `go.mod`.
7. Push branch, create draft PR titled `build(deps): bump Go toolchain to 1.25.8`.

**PR body:**
```
## Summary
Bump Go toolchain from 1.24.13 to 1.25.8 to clear 2 reachable stdlib CVEs flagged by the 2026-04-21 vulnerability-scan routine.

- GO-2026-4602 (`os` FileInfo Root escape) — reachable via `internal/generate/scan.go:44`
- GO-2026-4601 (`net/url` IPv6 host literal parsing) — reachable via `cmd/guide.go:92`

Also clears 9 unreachable stdlib CVEs in the same release line.

## Changes
- `go.mod` — `go 1.25.0`, `toolchain go1.25.8`

## Verification
- [x] `make build` — passes
- [x] `go test ./...` — passes
- [x] `go vet ./...` — clean
```

### PR #2 — Dependabot config

**Branch:** `security/dependabot-config`

**Files:**
- `.github/dependabot.yml` (new)

**Content:**
```yaml
version: 2
updates:
  - package-ecosystem: gomod
    directory: /
    schedule:
      interval: weekly
      day: monday
    open-pull-requests-limit: 5
    labels:
      - dependencies
      - go
    commit-message:
      prefix: "build(deps)"
  - package-ecosystem: github-actions
    directory: /
    schedule:
      interval: weekly
      day: monday
    open-pull-requests-limit: 5
    labels:
      - dependencies
      - github-actions
    commit-message:
      prefix: "build(deps)"
```

**Agent steps:**
1. Create `.github/dependabot.yml` with the content above.
2. Validate YAML parses (`python3 -c "import yaml; yaml.safe_load(open('.github/dependabot.yml'))"` or any YAML linter).
3. Push branch, create draft PR titled `ci: add Dependabot config for gomod + github-actions`.

**PR body:**
```
## Summary
Enable Dependabot weekly version + security PRs for Go modules and GitHub Actions. Auto-opens PRs on new CVE disclosures and weekly dep drift.

## Changes
- `.github/dependabot.yml` — 2 ecosystems, weekly schedule, 5 PR limit each

## Verification
- [x] YAML parses clean
- [ ] Post-merge: Dependabot tab populates with first-run state within 1 hour
```

### PR #3 — govulncheck CI job

**Branch:** `security/govulncheck-ci`

**Files:**
- `.github/workflows/ci.yml` — add new `govulncheck` job

**Job to add** (append under existing `lint` job):
```yaml
  govulncheck:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version-file: go.mod

      - name: Install govulncheck
        run: go install golang.org/x/vuln/cmd/govulncheck@latest

      - name: Run govulncheck
        run: govulncheck ./...
```

**Agent steps:**
1. Edit `.github/workflows/ci.yml` — append the job under existing `lint` job, preserving indentation.
2. Verify YAML parses (`python3 -c "import yaml; yaml.safe_load(open('.github/workflows/ci.yml'))"`).
3. Do NOT run govulncheck locally — CI verifies.
4. Push branch, create draft PR titled `ci: add govulncheck job`.

**PR body:**
```
## Summary
Add `govulncheck` as a CI job on PRs. Reachability-aware Go stdlib + module CVE scanner. Complements Dependabot (which covers modules only) by catching stdlib CVEs.

## Changes
- `.github/workflows/ci.yml` — new `govulncheck` job using `go install golang.org/x/vuln/cmd/govulncheck@latest`

## Verification
- [x] YAML parses
- [ ] Post-merge: CI `govulncheck` job green on main (Go 1.25.8 toolchain should produce zero reachable findings)
```

### PR #4 — CodeQL workflow

**Branch:** `security/codeql-workflow`

**Files:**
- `.github/workflows/codeql.yml` (new)

**Content:**
```yaml
name: CodeQL

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]
  schedule:
    - cron: '30 4 * * 1'  # Monday 04:30 UTC

permissions:
  contents: read
  security-events: write
  actions: read

jobs:
  analyze:
    name: Analyze Go
    runs-on: ubuntu-latest

    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version-file: go.mod

      - name: Initialize CodeQL
        uses: github/codeql-action/init@v3
        with:
          languages: go
          queries: security-and-quality

      - name: Autobuild
        uses: github/codeql-action/autobuild@v3

      - name: Perform CodeQL Analysis
        uses: github/codeql-action/analyze@v3
        with:
          category: "/language:go"
```

**Agent steps:**
1. Create `.github/workflows/codeql.yml` with content above.
2. Validate YAML parses.
3. Push branch, create draft PR titled `ci: add CodeQL workflow for Go SAST`.

**PR body:**
```
## Summary
Add CodeQL SAST workflow for Go. Runs on PR, push to main, and weekly schedule. `security-and-quality` query suite (broader than default `security-extended`).

## Changes
- `.github/workflows/codeql.yml` — Go analysis, PR + push + weekly cron

## Verification
- [x] YAML parses
- [ ] Post-merge: initial scan completes within 15min
- [ ] Security tab shows CodeQL results populated (0 findings expected on current codebase)
```

### Post-PRs — gitleaks history one-shot

**Not a PR.** Tech-lead session operation:

1. Run `docker run --rm -v "$PWD:/repo" zricethezav/gitleaks:latest detect --source=/repo --redact --verbose` (or `go install github.com/gitleaks/gitleaks/v8@latest` if Docker unavailable).
2. Triage output:
   - **Clean** → log result, done.
   - **False positives** (test fixtures, template placeholders) → add `.gitleaksignore` entries, re-run, commit via a small PR.
   - **Real leak** → STOP, escalate to user with finding details. User decides: `git filter-repo` history rewrite vs. rotate-and-accept.

## Dependencies

- Pre-flight working-tree commit (Task 1) — complete.
- GH repo settings (Secret Scanning, Push Protection, branch protection) — complete (user confirmed).

## Security

> [!warning]
> Refer to `Playbook/Standards/SecurityStandards.md` for all security requirements.

Specifically this plan honors:
- **Domain 1** (Secrets Management) — no secrets in any new config; GITHUB_TOKEN is GH-managed.
- **Domain 6** (Dependencies) — Dependabot enforces pin reviews; govulncheck enforces CVE gate.
- **Domain 7** (Embedded Catalog Integrity) — no catalog/embed changes; repo-infra only.

No new secrets, no new runtime dependencies, no user-facing behavior changes.

## Verification

**Per-PR (agent-side, pre-merge):**
- [ ] PR #1 — `make build`, `go test ./...`, `go vet ./...` green locally on 1.25.8
- [ ] PR #2 — YAML parses; dependabot.yml schema valid
- [ ] PR #3 — YAML parses; CI `govulncheck` job present in workflow file
- [ ] PR #4 — YAML parses; codeql.yml present

**Per-PR (tech-lead side, post-dispatch):**
- [ ] `gh pr checks <N>` — all required status checks green (test, lint, govulncheck once added, CodeQL once added, GitGuardian)
- [ ] `gh pr view <N>` — summary/changes/verification sections filled
- [ ] No scope creep — diff matches plan steps exactly

**Merge gate:**
- [ ] All required status checks green
- [ ] Branch up to date with main (required by branch protection)
- [ ] Squash merge with `gh pr merge <N> --squash --delete-branch`

**Post-merge verification (whole plan):**
- [ ] `git pull && make build && go test ./...` — main green
- [ ] GH Security tab shows: Dependabot alerts populated, CodeQL results populated, Secret Scanning active, govulncheck shown as check on recent PRs
- [ ] gitleaks history audit complete with logged outcome

## Out of Scope (Backlog on discovery)

If any new scanner surfaces findings:
1. Do NOT fix inline.
2. Log finding details to `Playbook/Backlog.md` under a new `## Security findings — surfaced by Plan 20` section.
3. Triage priority: reachable-CVE or injection pattern → P0/P1; unreachable or informational → P2/P3.
4. Separate plan per finding cluster.
