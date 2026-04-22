# Plan 24 — Pre-Launch Polish Bundle

**Tier:** 2
**Status:** Active
**Agent:** general-purpose (single dispatch, multi-file changes, no architectural deps between items)

## Goal

Ship four OSS-launch-readiness items as one PR + one batch of GitHub issues:

1. **CHANGELOG.md** at repo root — keep-a-changelog format, curated backfill of v0.1.0–v0.1.3.
2. **Astro PR build check** — docs.yml fires on `pull_request` with build-only (deploy stays push-to-main).
3. **Root `Bonsai/CLAUDE.md` tree drift** — refresh `internal/tui/` block to match current layout (Plan 15 harness/, Plan 22 initflow/).
4. **Seed 5 GitHub issues** labeled `good first issue` — drives contributor on-ramp for launch post.

## Context

Pre-public-announce punch list. User has locked scope, picked options, no further decisions pending.

- Item 4 deprecates Backlog Group C "CHANGELOG.md + richer release notes" (kept as Group D "Changelog generation skill + `bonsai changelog` CLI" for future) — consolidation is part of this plan's backlog cleanup.
- Item 5 is in-conversation GitHub issue creation, not repo changes — does not ride the PR.
- Item 6 & 7 are mechanical file edits — one agent dispatch for all three code changes (4/6/7).

## Dependencies

- `gh` CLI authed (verified in prior turns)
- `main` clean + ahead pushed (done: `018966d` now at origin)

## Steps

### Step A — CHANGELOG.md backfill (Item 4)

Create `/CHANGELOG.md` at repo root using keep-a-changelog 1.1.0 format with SemVer headers.

**Format:**

```markdown
# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.3] — 2026-04-15
### Changed
- Apply `gofmt -s` across all Go source files for canonical formatting.

## [0.1.2] — 2026-04-15
### Fixed
- Restore Go Reference badge now that case-collision module-proxy issue is resolved.
- Retract v0.1.0 release notice (see v0.1.1 fix).

## [0.1.1] — 2026-04-15
### Added
- Obsidian-compatible markdown links across generated workspace files (`[[link]]` syntax in nav tables and cross-references).
- `workspace-guide` skill — installed with every agent to onboard humans into the generated workspace.
- "How to Work" decision heuristics section in generated `CLAUDE.md`.
- PR-triggered CI workflow — tests + `go vet` as required status checks.
### Fixed
- Case-insensitive file collision on Windows/macOS — `station/index.md` renamed to `station/code-index.md` to avoid collision with `INDEX.md`.
- CI badge and Go version badge corrections in README.
- Reverted stray `INSTRUCT` reference to backtick in memory protocol.

## [0.1.0] — 2026-04-15

First public release. Go CLI for scaffolding Claude Code agent workspaces.

### Added
- Core commands — `bonsai init`, `add`, `remove`, `update`, `list`, `catalog`, `guide`.
- Six agent types — tech-lead (required), frontend, backend, fullstack, devops, security.
- Catalog of skills, workflows, protocols, sensors, routines — each with `meta.yaml` + content file; `required` and `agents` compatibility fields.
- Routines — periodic self-maintenance with auto-managed `routine-check` sensor + dashboard at `agent/Core/routines.md`.
- Sensors — auto-enforced hook scripts wired into `.claude/settings.json` at generation.
- Awareness framework — `status-bar` and `context-guard` sensors for live context monitoring and guardrails.
- Lock file (`.bonsai-lock.yaml`) — content hashing with conflict detection on re-run; user-modified files trigger skip/overwrite/backup prompt.
- Selective file update with multi-select conflict picker via `bonsai update`.
- Individual item removal — `bonsai remove <kind> <name>`.
- `display_name` catalog field — decouples human-readable labels from machine identifiers; auto-derived from `name` when omitted.
- Backlog system across catalog — intake queue scaffolding for every agent.
- PR-driven subagent workflow — dispatched agents create draft PRs in isolated worktrees; tech-lead reviews and merges.
- Release pipeline — GoReleaser v2, GitHub Actions tag trigger, Homebrew tap (`LastStep/homebrew-tap`).
- Dogfooding — Bonsai generates its own `station/` workspace with tech-lead agent.

[Unreleased]: https://github.com/LastStep/Bonsai/compare/v0.1.3...HEAD
[0.1.3]: https://github.com/LastStep/Bonsai/compare/v0.1.2...v0.1.3
[0.1.2]: https://github.com/LastStep/Bonsai/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/LastStep/Bonsai/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/LastStep/Bonsai/releases/tag/v0.1.0
```

**Constraints:**
- No inline commit SHAs — human-readable prose only.
- v0.1.0 section = curated highlights (~18 bullets covering architecture) — not raw log.
- `## [Unreleased]` stub left empty for next-release bookkeeping.

### Step B — Astro PR build check (Item 6)

Edit `/.github/workflows/docs.yml` — add `pull_request` trigger, gate `deploy` job with `if:` guard.

**Change the `on:` block to:**

```yaml
on:
  push:
    branches: [main]
    paths:
      - 'website/**'
      - 'catalog/**'
      - 'docs/**'
      - 'README.md'
  pull_request:
    paths:
      - 'website/**'
      - 'catalog/**'
      - 'docs/**'
      - 'README.md'
  workflow_dispatch:
```

**Gate the `deploy` job by adding `if:`:**

```yaml
  deploy:
    needs: build
    if: github.event_name == 'push'
    runs-on: ubuntu-latest
    environment:
      name: github-pages
      url: ${{ steps.deployment.outputs.page_url }}
    steps:
      - id: deployment
        uses: actions/deploy-pages@v4
```

**Also update `build` job so `upload-pages-artifact` only runs on `push`** (PRs have no pages artifact to upload):

```yaml
      - uses: actions/upload-pages-artifact@v5
        if: github.event_name == 'push'
        with:
          path: website/dist
```

**Constraints:**
- Don't touch the `concurrency:` block.
- No changes to `ci.yml`, `codeql.yml`, `release.yml`.

### Step C — Root `Bonsai/CLAUDE.md` tree drift (Item 7)

Edit `/CLAUDE.md` — replace the `internal/` block in the Project Structure tree.

**Current (lines 35–50):**

```
├── internal/
│   ├── catalog/
│   │   └── catalog.go       ← loads YAML metadata from embedded catalog/
│   ├── config/
│   │   ├── config.go        ← ProjectConfig, InstalledAgent + YAML I/O
│   │   └── lockfile.go      ← LockFile, content hashing, conflict detection
│   ├── generate/
│   │   ├── generate.go      ← renders templates, writes files to target project
│   │   ├── generate_test.go ← tests for core file generation
│   │   ├── frontmatter.go   ← BONSAI marker parsing for generated files
│   │   ├── frontmatter_test.go ← tests for frontmatter parsing
│   │   ├── scan.go          ← custom file discovery (user-created abilities)
│   │   └── scan_test.go     ← tests for custom file scanning
│   └── tui/
│       ├── styles.go         ← LipGloss styles, panels, trees, display helpers
│       └── prompts.go        ← Huh form wrappers (text, select, multi-select, confirm)
```

**Replace the `tui/` sub-block only (last 3 lines) with:**

```
│   └── tui/
│       ├── styles.go         ← LipGloss styles, palette tokens, panels, trees
│       ├── styles_test.go    ← tests for palette + display helpers
│       ├── prompts.go        ← Huh form wrappers (text, select, multi-select, confirm)
│       ├── filetree.go       ← RenderFileTree widget for scaffold previews
│       ├── filetree_test.go  ← tests for file tree renderer
│       ├── harness/          ← BubbleTea step/reducer harness (Plan 15)
│       └── initflow/         ← `bonsai init` cinematic flow — stages + chrome (Plan 22)
```

**Constraints:**
- Depth: list subdirs only; do not enumerate the 4 harness/ files or 22 initflow/ files.
- Preserve surrounding whitespace and box-drawing characters.
- No other changes to CLAUDE.md.

### Step D — Backlog consolidation (Item 4 cleanup)

Edit `/station/Playbook/Backlog.md`:
- **Remove** the Group C entry `- **[improvement] CHANGELOG.md + richer release notes**` (now resolved by Plan 24).
- **Remove** the Group C entry `- **[improvement] Consolidate or delineate CHANGELOG backlog items**` (resolved — Plan 24 keeps only the Group D future work item).
- **Keep** the Group D `- **[feature] Changelog generation skill + release changelogs**` entry — add `(refiled as good-first-issue via Plan 24 Step E)` suffix.
- **Add removal comments** in HTML-comment style (`<!-- "CHANGELOG.md..." — resolved 2026-04-22 via Plan 24 -->`) per Backlog convention.

### Step E — Seed 5 GitHub issues (Item 5)

**Not part of the PR** — handled by tech-lead directly post-dispatch via `gh issue create`.

Issues to file (titles + brief bodies drafted at dispatch time):

| Label | Title |
|-------|-------|
| A | Add shell completion via `bonsai completion [bash\|zsh\|fish]` |
| B | Add `bonsai changelog` command + changelog-generation skill |
| E | (re-label existing issue #53) Port statusLine prototype to catalog sensor |
| G | Umbrella — propose and add a new catalog skill/workflow/protocol |
| H | Post-update `.bak` merge hint in `bonsai update` |

**Labels to apply:**
- `good first issue` on all 5
- `help wanted` on all 5
- Prereq: ensure both labels exist in repo (create via `gh api` if missing)

## Security

> [!warning]
> Refer to `Playbook/Standards/SecurityStandards.md` for all security requirements.

No security-sensitive changes: docs-only file edits + workflow YAML trigger changes + changelog text + GitHub issue creation. No secrets, no new deps, no executable code changes.

## Verification

- [ ] `cat CHANGELOG.md | head -5` — shows keep-a-changelog header.
- [ ] CHANGELOG.md validated as valid markdown (no missing link references at bottom).
- [ ] `make build` still passes — no Go source touched.
- [ ] `go test ./...` still passes.
- [ ] `gh workflow view docs.yml` (or `cat .github/workflows/docs.yml`) shows `pull_request` trigger + `if: github.event_name == 'push'` on deploy job.
- [ ] Root `Bonsai/CLAUDE.md` tree block for `internal/tui/` matches actual `ls internal/tui/` output.
- [ ] Backlog Group C has 2 fewer live entries + 2 removal comments; Group D entry updated.
- [ ] 5 GitHub issues created, all labeled `good first issue` + `help wanted`. (Counted via `gh issue list --label "good first issue"`.)
- [ ] PR created as draft, links to Plan 24.
- [ ] After merge: `gh release list` unchanged (no release cut); CHANGELOG.md visible on repo homepage.
