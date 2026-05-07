---
tags: [plan, tier-1, docs]
status: Active
tier: 1
agent: tl
---

# Plan 37 — Doc Refresh Bundle (code-index + INDEX verify)

**Tier:** 1
**Status:** Active
**Agent:** tl (inline; no dispatch needed)

## Goal

`station/code-index.md` line numbers match current Go source. `station/INDEX.md` architecture diagram + cmd count verified current. Both files reflect post-v0.4.0 state.

## Source

Routine-digest 2026-05-07 — both items flagged 2026-05-04 Doc Freshness Check; Plan 36 docs sweep (PR #94) refreshed structure but did not refresh line numbers in `code-index.md`. Drift verified 2026-05-07: `cmd/init_flow.go` helpers off by 21-24 lines, `cmd/add.go` helpers off by 16-39 lines.

## Steps

1. **`station/code-index.md` — refresh CLI command table line numbers**
   - For each row in "CLI Commands" + "Shared Helpers" + "Init Helpers" + "Add Helpers" + "Remove Helpers" tables: re-grep `^func <name>` in the referenced `.go` file, replace `:NNN` with actual line number.
   - Files to spot-check: `cmd/root.go`, `cmd/init.go`, `cmd/init_flow.go`, `cmd/add.go`, `cmd/remove.go`, `cmd/list.go`, `cmd/catalog.go`, `cmd/update.go`, `cmd/guide.go`, `cmd/validate.go`.
2. **`station/code-index.md` — refresh `internal/catalog/catalog.go` line numbers**
   - Re-grep for: `DisplayNameFrom`, `New`, `loadItems`, `loadSensors`, `loadRoutines`, `loadScaffolding`, `loadAgents`. Update `:NNN`.
3. **`station/code-index.md` — refresh `internal/config/` line numbers**
   - Re-grep `config.go` + `lockfile.go` for the 13 listed entries. Update `:NNN`.
4. **`station/code-index.md` — refresh `internal/generate/` line numbers**
   - Largest file (`generate.go` 1357+ lines). Re-grep all listed functions across `generate.go`, `frontmatter.go`, `scan.go`, `catalog_snapshot.go`. Update `:NNN`.
5. **`station/code-index.md` — refresh `internal/tui/` line numbers**
   - `styles.go`, `prompts.go`, `harness/harness.go`, `harness/steps.go`, `initflow/*.go`, `addflow/*.go`. Update `:NNN`.
6. **`station/INDEX.md` — verify architecture diagram + key metrics**
   - Confirm 6 internal pkgs (catalog, config, generate, validate, wsvalidate, tui) listed.
   - Confirm 8 CLI cmds (init, add, remove, list, catalog, update, guide, validate) in Key Metrics + Architecture Overview.
   - Confirm Tech Stack rows match `go.mod` Go version (`1.24+`) and dependencies.
   - **No edit if clean.** Document verification result in plan completion note.

## Verification

- [ ] `grep -n ":[0-9]" station/code-index.md` — every `:NNN` reference resolves to a real line in target file (sample 5 random rows: `grep -n "^func <name>" <file>` → matches indexed line).
- [ ] `station/INDEX.md` Key Metrics CLI count = output of `ls cmd/*.go | grep -v _test | grep -v root.go | grep -v main.go | wc -l` plus 1 for root.go derived cmd; or simpler: `bonsai --help` subcommand list.
- [ ] `station/INDEX.md` "internal pkgs" listing in Architecture Overview matches `ls internal/`.
- [ ] No new functions/types in `cmd/` or `internal/` that aren't represented in `code-index.md` (manual scan: count `^func ` in each file vs row count in matching table).

## Security

> [!warning]
> Refer to SecurityStandards.md. Doc-only edits — no runtime/security surface change.

## Notes

- Pure documentation refresh. No source code change. No tests needed.
- Inline execution by tech lead — no agent dispatch, no worktree.
- Estimated effort: 20-30 min mechanical.
- After completion: archive plan to `Plans/Archive/37-doc-refresh-bundle.md`, append Status row.
