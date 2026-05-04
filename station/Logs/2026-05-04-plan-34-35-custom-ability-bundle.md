---
date: 2026-05-04
plans: [34, 35]
prs: [92, 93]
tags: [session-log, plan-34, plan-35, custom-abilities, validate-cmd, v0.4-prep]
---

# 2026-05-04 — Plan 34 + Plan 35 + dogfood backfill

**Two plans shipped. Foundation laid for v0.4.0 release.**

## Triggering session

User reported bug: tech-lead agent in user's project created custom abilities, manually edited `.bonsai.yaml` to register them, then `bonsai update` silently skipped discovery. Files never lock-tracked.

## Plan 34 — Custom ability discovery bug bundle (PR #92, `1e5affa`)

Tier-1 bundle, 4 fixes:

| Fix | File | Detail |
|-----|------|--------|
| Doc | `catalog/skills/bonsai-model/bonsai-model.md` | Spell out: `.bonsai.yaml` is bonsai-managed; never hand-edit ability lists or `custom_items`. Drop file with frontmatter + `bonsai update` instead. |
| Orphan recovery | `internal/generate/scan.go` | `ScanCustomFiles` now skips only when name in `installed.<Cat>` AND relPath in lock. Pre-registered names with missing lock entries are re-discovered + lock-tracked. New `Orphaned bool` flag. |
| Sensor frontmatter | `internal/generate/frontmatter.go` | `ParseFrontmatter` accepts bash-comment frontmatter (`# ---` after optional shebang). Custom sensor files (`.sh` with executable shebang) finally discoverable. |
| Non-TTY UX | `internal/tui/updateflow/run.go` | `RunStatic` prints `warning: <relPath> — <Error>` to stderr per invalid file. CI/scripts get a signal. |

Tests added: 4 (1 in scan_test.go, 3 in frontmatter_test.go). Sandbox repros: 3 scenarios all passed.

Code review: APPROVE, 2 nits (tab handling in stripBashCommentPrefix, minor lock-source filter precision). No blockers.

## Plan 35 — `bonsai validate` command (PR #93, `878cca2`)

v0.4.0 headline new feature. Read-only audit. Justifies the minor bump.

| Component | LOC | Detail |
|-----------|-----|--------|
| `internal/validate/validate.go` | ~360 | New package. `Run(projectRoot, cfg, cat, lock, agentFilter) (*Report, error)`. Six issue categories: `orphaned_registration`, `missing_file`, `stale_lock_entry`, `untracked_custom_file`, `invalid_frontmatter`, `wrong_extension_in_category`. |
| `internal/validate/validate_test.go` | ~330 | 14 table-driven tests, one per category + clean + filter + multi-category + edge cases (catalog guardrail, top-level-only, nil-cfg, sorted invariants). |
| `cmd/validate.go` | ~100 | Cobra wrapper. Flags: `-a/--agent`, `--json`. Exit 0/1/2 contract. |
| `cmd/validate_test.go` | ~80 | 3 smoke tests (text-no-issues, JSON shape, e2e orphan). |
| Docs | — | `bonsai validate` rows in `bonsai-model.md` + `workspace-guide.md.tmpl`. |

7 sandbox scenarios verified — clean, orphan, missing_file, untracked, invalid_frontmatter, wrong_ext, stale_lock_entry, agent filter — all correct exit codes.

Implementer's design decisions (all approved):
- Catalog-tracked items (lock source not `custom:`) skipped by orphan check. Catalog items have empty `custom_items` by design — flagging would be false positive.
- Stale-lock per-agent scoping prevents double-reporting. Trade-off: lock entries belonging to no installed agent are silently skipped (rare in practice).
- Unused `cat *catalog.Catalog` param kept in `Run` signature for future API stability.

Code review: APPROVE, 3 nits (ownerless lock entries, O(N×M) per-agent scan, Windows path edge — all pre-existing patterns). No blockers.

## Dogfood signal

Right after merge, `./bonsai validate` on this very repo caught 2 real issues:

1. `station/agent/Skills/bubbletea.md` — hand-written custom skill, no frontmatter.
2. `station/agent/Sensors/statusline.sh` — issue #53 prototype sensor, no frontmatter.

Both backfilled inline (`8e21b75`):
- `bubbletea.md`: prepend `---\nname: bubbletea\ndescription: ...\n---`
- `statusline.sh`: bash-comment frontmatter block after shebang (uses Plan 34's new parser).

`bonsai update` registered both. Validate now clean.

## v0.4.0 release path forward

Next session, run Plan 36 — release prep:
- `workflow_dispatch:` trigger on `.github/workflows/release.yml` (P1 backlog item — clean retry path).
- Bump `golang.org/x/net` v0.38 → v0.45+ (P2 security hygiene, GO-2026-4441 + GO-2026-4440).
- Close `[Unreleased]` → `[0.4.0]` in CHANGELOG.md with three sections (Added/Changed/Fixed/Security).
- Tag `v0.4.0`, push, monitor GoReleaser, manual brew formula push if PAT expired.
- HOMEBREW_TAP_TOKEN expiry: rotate before ~2026-07-15.

## Process notes

- **Tier-1 autonomous dispatch worked cleanly in both plans.** Single worktree per plan, single dispatch, independent code-review agent, same-session squash-merge. Memory pattern for "small bug bundles" + "single feature with crisp file-level spec" both held up.
- **Plan 34 → Plan 35 sequencing was natural.** Plan 34 was reactive (recover from bad state during update); Plan 35 was the proactive companion (detect bad state without write side effects). Same pair-shape probably appears elsewhere — `add` (reactive merge) ↔ pre-add validation, `remove` ↔ pre-remove conflict scan.
- **Worktree post-merge cleanup hit memory rule again** (3rd time this month per memory note count). Both PRs needed `git worktree remove -f -f` + `git branch -D` + `git push origin --delete` after `gh pr merge --squash --delete-branch` skipped local+remote cleanup.

## Open items / followups

- v0.4.0 release prep (Plan 36 — drafted in memory, not yet planned).
- Reviewer nit from Plan 35: stale-lock entries belonging to no installed agent silently skipped — file as Backlog P3 if it bites.
- Backlog skills frontmatter convention (P2 Group C) — partially closed by bubbletea backfill; remaining catalog skills unaffected since they ship via `meta.yaml` not file frontmatter (the frontmatter convention question is for file-level metadata only).
