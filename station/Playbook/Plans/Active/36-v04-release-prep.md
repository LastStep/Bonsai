---
tags: [plan, release, v0.4, docs]
description: v0.4.0 release-prep bundle — workflow_dispatch retry hook, x/net + Go toolchain bumps, CHANGELOG entry, doc-drift sweep (root CLAUDE.md tree, INDEX.md arch diagram, code-index.md). Headline shipped via Plan 35; this cuts the tag.
---

# Plan 36 — v0.4.0 Release Prep

**Tier:** 2
**Status:** Draft
**Agent:** general-purpose (single worktree, 4 file-disjoint phases)

## Goal

Cut `v0.4.0` cleanly with the `bonsai validate` headline (Plan 35), zero reachable CVEs, doc-drift fixed, and a retry-friendly release pipeline. Tag annotated `v0.4.0`, GoReleaser publishes binaries + Homebrew formula, doc state matches code state.

## Context

Plan 35 (`bonsai validate`) shipped 2026-05-04 as the v0.4 headline. Plan 34 (custom-ability discovery bug bundle) shipped same day. v0.3.0 last tagged at `ac59f8b` (2026-04-24); 23 commits since. Routine digest 2026-05-04 surfaced compounded doc drift across root `Bonsai/CLAUDE.md`, `INDEX.md` arch diagram, and `code-index.md`. Same digest confirmed 0 reachable CVEs but flagged Go 1.25.8 → 1.25.9 + `golang.org/x/net` v0.38.0 → v0.45.0+ as the cleanest path to a tag-zero security baseline.

## Steps

### Phase A — Dep + toolchain bumps

Touches: `go.mod`, `go.sum`. Optional: `.github/workflows/codeql.yml`.

A1. Bump `golang.org/x/net` v0.38.0 → v0.53.0:
   ```
   go get golang.org/x/net@v0.53.0
   go mod tidy
   ```
   Verify `go.sum` regenerated, no spurious diffs in unrelated modules.

A2. Bump Go toolchain in `go.mod`:
   - Keep `go 1.25.0` (language directive — no API change needed).
   - Update `toolchain go1.25.8` → `toolchain go1.25.9`.
   - All CI workflows use `go-version-file: go.mod` (verified via research) — no separate workflow Go-version edits required.

A3. Optional consistency: `.github/workflows/codeql.yml` — bump `actions/setup-go@v5` → `@v6` (matches `ci.yml` + `release.yml`). One-line edit.

A4. Verify locally:
   ```
   go vet ./...
   go test ./...
   govulncheck ./...
   ```
   Expected: 0 reachable findings, 0 unreachable findings (clears all 8 from 2026-05-04 dependency-audit). `npm audit` not affected (separate scope).

### Phase B — release.yml workflow_dispatch retry hook

Touches: `.github/workflows/release.yml`. No other files.

B1. Add `workflow_dispatch` trigger with `tag` input under existing `on:` block:
   ```yaml
   on:
     push:
       tags:
         - "v*"
     workflow_dispatch:
       inputs:
         tag:
           description: "Existing tag to (re-)release (e.g. v0.4.0). Must already exist."
           required: true
           type: string
   ```

B2. Add ref override on the checkout step so dispatched runs operate on the requested tag:
   ```yaml
         - name: Checkout
           uses: actions/checkout@v4
           with:
             fetch-depth: 0
             ref: ${{ github.event.inputs.tag || github.ref }}
   ```

B3. **Do not** add `--skip=publish` flags or split brew job — keep retry path simple. Document precondition: "delete the failed Release first via `gh release delete v0.X.Y --cleanup-tag=false` before re-running" — add note to a new `RELEASE.md` snippet OR comment inline in the workflow YAML at the dispatch trigger.

### Phase C — Doc-drift sweep

Touches: `Bonsai/CLAUDE.md` (project root, outside `station/`), `station/INDEX.md`, `station/code-index.md`.

> All OLD/NEW string blocks below are exact, copy-paste-ready for the Edit tool. Line numbers + block boundaries verified against current HEAD (`c45c8b8`).

C1. **Root `Bonsai/CLAUDE.md` project-structure tree** — refresh `cmd/`, `internal/`, `internal/tui/` blocks. Add `cmd/init_flow.go`, `cmd/validate.go`; add `internal/validate/` (`validate.go` + `validate_test.go`), `internal/wsvalidate/` (`wsvalidate.go` + `wsvalidate_test.go`); add `internal/generate/{catalog_snapshot.go, catalog_snapshot_test.go, bonsai_reference_test.go, refresh_peer_awareness_test.go}`; add 7 missing `internal/tui/{flow}/` packages (`addflow`, `removeflow`, `updateflow`, `listflow`, `catalogflow`, `guideflow`, `hints`). Use the patch in C-Patch-1 (below).

C2. **`station/INDEX.md`** — refresh Architecture Overview ASCII diagram. Add `internal/validate/` + `internal/wsvalidate/` lines. Replace `internal/tui/   ← Huh forms + LipGloss styled output` with a line covering BubbleTea cinematic flows + harness + Huh + LipGloss. Update `cmd/ (Cobra)` line to list all 8 commands. Use C-Patch-2.

C3. **`station/code-index.md`** — three sub-patches:
   - **3a.** CLI Commands table — fix 7 stale line numbers (`cmd/init_flow.go:26→:27`, `cmd/add.go:26→:28`, `cmd/add.go:54→:56`, `cmd/remove.go:32→:34`, `cmd/list.go:19→:18`, `cmd/catalog.go:16→:23`, `cmd/update.go:22→:19`, `cmd/guide.go:34→:27`, `runGuide :42→:44`). Add new row for `bonsai validate` (`cmd/validate.go:23` → `runValidate :43`). Use C-Patch-3a.
   - **3b.** Generator section — append `### catalog_snapshot.go` subsection. Add `## Validate (internal/validate/)` and `## Workspace-path Validation (internal/wsvalidate/)` sections between Generator and TUI. Use C-Patch-3b.
   - **3c.** TUI section — append 6 missing flow-package subsections (RemoveFlow, UpdateFlow, ListFlow, CatalogFlow, GuideFlow, Hints) before `## Generation Flow`. Use C-Patch-3c.

> [!note]
> The full OLD/NEW string content for C-Patch-1 through C-Patch-3c is bundled in the agent dispatch prompt — too large to inline here. Stored in research output at `tasks/ac1817dc730b1df01.output` (TL provides verbatim in prompt).

### Phase D — CHANGELOG entry

Touches: `CHANGELOG.md`. No other files.

D1. Replace empty `## [Unreleased]` placeholder with `## [0.4.0] - 2026-05-04` section. Mirrors the existing v0.3.0 / v0.2.0 Added/Changed/Fixed/Security style. Cover Plans 32 / 33 / 34 / 35 only (Plans 30 / 31 already shipped under v0.3.0):

```markdown
## [Unreleased]

## [0.4.0] - 2026-05-04

> **The "audit your workspace" release.** New `bonsai validate` command surfaces drift between catalog and installed state. Plus a custom-ability discovery fix that recovers orphaned hand-rolled abilities, a website concept-page rewrite, and a chokepoint-hardening followup bundle.

### Added
- **`bonsai validate`** — read-only ability-state audit. Reports orphaned files, missing required items, lock/disk drift; `--json` for CI/agents; exits non-zero on any finding so CI can gate on workspace integrity. Headlines v0.4. (Plan 35, [#93](https://github.com/LastStep/Bonsai/pull/93))
- **NoteStandards skill** in generated workspaces — 3-line cap on log/decision entries with link-out for trackers, wired into memory protocol and session-logging workflow.

### Changed
- **`bonsai update` recovers orphaned custom-ability registrations** — hand-authored abilities under `agent/{Skills,Workflows,Protocols}/` that were missing from `.bonsai.yaml` are now picked up and registered on update. (Plan 34, [#92](https://github.com/LastStep/Bonsai/pull/92))
- **`wsvalidate` extracted** to a dedicated package; `Validate()` is now the single chokepoint for config validation; snapshot writes hardened with `O_NOFOLLOW`. (Plan 32, [#80](https://github.com/LastStep/Bonsai/pull/80))
- **Website concept-page rewrite** — narrative-first explainers replace reference-style pages; clearer mental model for new users. (Plan 33, [#79](https://github.com/LastStep/Bonsai/pull/79))
- **Sensor scripts now ship with shebang-aware frontmatter** so generator preserves `#!/usr/bin/env bash` across re-runs. ([#92](https://github.com/LastStep/Bonsai/pull/92))

### Fixed
- **Custom-ability discovery bug bundle** — orphaned registrations, sensor shebang frontmatter loss, and non-TTY warning channel resolved. (Plan 34, [#92](https://github.com/LastStep/Bonsai/pull/92))
- **`bonsai update` non-TTY warnings** route to stderr instead of being suppressed. ([#92](https://github.com/LastStep/Bonsai/pull/92))

### Security
- `O_NOFOLLOW` on snapshot writes — defense-in-depth against symlink-substitution races during `.bonsai/catalog.json` materialization. (Plan 32, [#80](https://github.com/LastStep/Bonsai/pull/80))
- **Go toolchain 1.25.9** — clears 6 unreachable stdlib CVEs (GO-2026-4864/4947/4946/4870/4869/4865).
- **`golang.org/x/net` v0.53.0** — clears GO-2026-4441 + GO-2026-4440. Reachable-set vuln count: 0.
```

D2. Update footnote links at bottom of file:
```
[Unreleased]: https://github.com/LastStep/Bonsai/compare/v0.4.0...HEAD
[0.4.0]: https://github.com/LastStep/Bonsai/compare/v0.3.0...v0.4.0
```

### Phase E — Tag + Release (TL action, post-merge)

Not in agent scope — TL executes after PR merged + post-merge tests green.

E1. `git pull origin main && git log --oneline v0.3.0..HEAD | wc -l` — confirm 23+ commits.

E2. Annotate + push tag:
```bash
git tag -a v0.4.0 -m "v0.4.0 — bonsai validate

New: read-only ability-state audit command (\`bonsai validate\`).
Fix: custom-ability discovery — orphaned registrations recovered on update.
Polish: wsvalidate extract + Validate() chokepoint, NoteStandards in workspaces,
website concept-page rewrite. 23 commits since v0.3.0."
git push origin v0.4.0
```

E3. Watch GoReleaser run via `gh run watch`. If brew step fails (PAT expiry or otherwise), use the dispatch trigger from Phase B for retry — first `gh release delete v0.4.0 --cleanup-tag=false`, then `gh workflow run release.yml --ref v0.4.0 -f tag=v0.4.0`.

E4. Verify: `gh release view v0.4.0`, Homebrew formula at `LastStep/homebrew-tap` updated.

E5. Update `Status.md` (move Plan 36 → Recently Done), update memory.md Work State, archive plan to `Plans/Archive/`.

## Dependencies

- Plans 32, 33, 34, 35 already shipped + archived (verified).
- v0.3.0 tagged at `ac59f8b` (verified — memory note about "no v0.3 tag" is stale; flag for memory-consolidation routine, do NOT block this plan).
- HOMEBREW_TAP_TOKEN PAT rotated 2026-04-22 (next rotation due ~2026-07-15 per Backlog P1 — should be valid for v0.4.0).

## Security

> [!warning]
> Refer to [Playbook/Standards/SecurityStandards.md](../../Standards/SecurityStandards.md) for all security requirements.

- **No new code surface** — release prep only. Toolchain + dep bumps reduce CVE exposure (8 unreachable findings → 0).
- **`govulncheck ./...` MUST run clean** post-bump (Phase A4 verification gate).
- **`bonsai validate` MUST run clean** on the dogfood repo post-merge (Plan 35 dogfooding rule from memory).
- **No secrets touched** — release.yml dispatch input is a tag string, not a token.
- **Tag signing** — current convention is unsigned annotated tags (consistent with v0.2.0 / v0.3.0); preserve.

## Verification

Agent must complete ALL of these in the worktree before reporting done:

- [ ] Phase A: `go vet ./...` clean
- [ ] Phase A: `go test ./...` — all tests pass
- [ ] Phase A: `govulncheck ./...` — 0 reachable, 0 unreachable findings
- [ ] Phase A: `make build` clean (binary builds with new toolchain)
- [ ] Phase A: `git diff go.mod go.sum` shows ONLY x/net version + toolchain directive changes (plus indirect dep version bumps from `go mod tidy`)
- [ ] Phase B: `release.yml` YAML lint clean — paste into `yamllint` or `actionlint`
- [ ] Phase C: `bonsai validate` (dogfood) on the worktree → exit 0, no findings
- [ ] Phase C: every link in patched `station/CLAUDE.md` (if any), `station/INDEX.md`, `station/code-index.md` resolves on disk
- [ ] Phase D: `CHANGELOG.md` parses as valid markdown; section ordering 0.4.0 → 0.3.0 → 0.2.0 → 0.1.x preserved
- [ ] Final: `make build && go test ./...` green
- [ ] Draft PR created targeting `main`, body lists all 4 phases.

## Out of Scope

- CodeQL Action v3 → v4 bump (P1 backlog) — v4 not yet released; defer.
- 23-module hygiene refresh (P3 backlog) — separate post-release sweep PR.
- Add root-CLAUDE.md tree-drift check to `doc-freshness-check` routine (P2 backlog) — separate plan; not v0.4 scope.
- Install `semgrep` (P2 backlog) — host-side action, not code change.
- Memory note "v0.3.0 not tagged" correction — handle in next memory-consolidation routine; not release-blocking.

## Dispatch

| Phase | Files Touched | Agent | Dependencies |
|-------|---------------|-------|--------------|
| A — Dep + toolchain | `go.mod`, `go.sum`, `.github/workflows/codeql.yml` | general-purpose | none |
| B — workflow_dispatch | `.github/workflows/release.yml` | general-purpose | none |
| C — Doc sweep | `Bonsai/CLAUDE.md`, `station/INDEX.md`, `station/code-index.md` | general-purpose | none |
| D — CHANGELOG | `CHANGELOG.md` | general-purpose | none |
| E — Tag + Release | (post-merge) | tech-lead | A+B+C+D merged |

All four agent-phases are file-disjoint. Single-agent bundled dispatch recommended (low coordination overhead, all small text edits).
