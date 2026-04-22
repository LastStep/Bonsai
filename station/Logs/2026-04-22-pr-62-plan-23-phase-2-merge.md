---
tags: [log, session, plan-23, phase-2, pr-review, merge]
date: 2026-04-22
plan: 23
phase: 2
pr: 62
squash: 61777c0
---

# PR #62 merge — Plan 23 Phase 2 (`bonsai add` cinematic add-items branch + ConflictsStage)

## What shipped

PR #62 squash `61777c0`. Net +1099 / −57 across 9 files.

### Add-items branch wired end-to-end

- `cmd/add_redesign.go` splicer — installed-agent arm now returns `[AddItemsGraft, Observe]` (Ground skipped). `AgentDisplay` + `SetDefaultWorkspace` stamped from `installedAgent.Workspace`.
- `buildAddGrowAction` handles both branches via type-scan for `GraftResult` in `prev[]`. Add-items arm mirrors legacy `runAddSpinner` add-items body — calls `availableAddItems` to shape picks, funnels through shared `distributeAddItemPicks` so category slices are bit-identical to the legacy path.
- `internal/tui/addflow/graft.go` — `NewAddItemsGraft` filter body already landed Phase 1; Phase 2 adds regression coverage.

### New `ConflictsStage`

- `internal/tui/addflow/conflicts.go` — 438-line new file. One tab per conflict file, 3-way radio (Keep / Overwrite / Backup, default Keep). `Result()` returns `map[string]config.ConflictAction`. Rail extended: `addflow.StageLabels` grows to 7 with `衝 しょう CONFLICT` at new idx 5; Yield moves to idx 6. `Result()` returns defensive copy.
- `internal/tui/initflow/design.go` — real `ConflictRowStyle(tone)` body (Success/Danger/Warning palette tokens per action), new `ConflictActionGlyph(tone)` helper. Phase 1 stub replaced.
- `internal/config/lockfile.go` — new `ConflictAction` enum (Keep/Overwrite/Backup). In-memory transport only; no yaml struct tag, never persisted.
- `cmd/add_redesign.go` slot [3] LazyGroup — switched from legacy `buildConflictSteps` to `[NewConflictsStage(ctx, &wr)]`. Post-harness helper `applyCinematicConflictPicks` consumes the map and drives `.bak` writes + `wr.ForceSelected`. Legacy `applyConflictPicks` untouched (still used by remove/update/legacy add).

### Observe refactor

- `SetPrior` switched from positional indexing to type-scan so it handles both prev shapes (new-agent `[string, string, GraftResult, bool]`, add-items `[string, GraftResult, bool]`) without a branch-specific ctor.
- Added `SetDefaultWorkspace` so the splicer can seed `installedAgent.Workspace` where Ground is skipped.

### Tests

9 new `conflicts_test.go` cases (tab count, default=Keep, tab wrap, radio clamp, Enter advance+complete, Result map population, empty WriteResult, Reset preservation, multi-dim render smoke) + 2 graft regression tests (live tab count update, new-agent unaffected by filter) + 1 observe regression test (add-items prev shape).

## Review passes

**Independent code review — PASS with MINOR_ISSUES.** Plan compliance ticked, no blockers. 3 MINOR + 4 NIT deferred to Backlog. Test gap flagged: `applyCinematicConflictPicks` has no direct unit test (branches exercised transitively only).

**Security review — MINOR_FINDINGS.** 3 items, all MINOR, all filed to Backlog:
1. `.bak` write errors silently discarded in `applyCinematicConflictPicks:281-285` — **same bug in legacy `applyConflictPicks` at `cmd/root.go:158`**, carried forward, not a regression.
2. TOCTOU on `os.ReadFile` for Backup picks — file can disappear between detection and completion.
3. Missing defensive intersect of picks-keys ∩ `wr.Conflicts()` — currently safe (keys built strictly from `wr.Conflicts()`), hardening only.

Confirmed safe: zero dep changes, no new env reads beyond `BONSAI_ADD_REDESIGN`, bounds safe, zero-value `ConflictAction` is Keep (non-destructive default), `ConflictsStage.Update` all slice accesses bounded, `distributeAddItemPicks` receives picks in matching category order so bit-identity holds.

## CI

6/6 green: test / lint / govulncheck / CodeQL / GitGuardian / Analyze Go. Push-CI re-ran on main squash — also green.

## Post-merge cleanup

- `gh pr merge --delete-branch` removed remote branch server-side but `failed to delete local branch` because worktree `agent-aaf027ba` held it (documented pattern in memory — happens every squash-merge from worktree).
- Manual cleanup: `git worktree remove -f -f`, `git branch -D`, `git push origin --delete` — verified remote gone via `git ls-remote | grep`.

## Deferred — filed to Backlog

- **[security/MINOR]** Shared `.bak` write-error silent-discard in `applyCinematicConflictPicks` (add) + legacy `applyConflictPicks` (add/remove/update) — user can end up with overwritten file + no backup. Remediation: drop path from `toOverwrite` on backup failure, surface `tui.Warning`.
- **[debt/MINOR]** `confIdx := len(results) - 2` arithmetic in post-harness cleanup fragile — switch to type-scan for `map[string]config.ConflictAction` (same pattern Observe uses).
- **[debt/MINOR]** No direct unit test for `applyCinematicConflictPicks` — three action branches + mixed-action input only tested transitively.
- **[improvement/MINOR]** `generate.FileResult` has no inline-diff field; `ConflictsStage.renderDiffSummary` emits placeholder with TODO. Expose real diff for better UX when Phase 3 lands.

## Remaining Plan 23 work

- **Phase 3** — flip default (remove `BONSAI_ADD_REDESIGN` env gate) + delete legacy `runAdd` body + delete dead `yieldModeAddItemsDeferred` + `NewYieldAddItemsDeferred` + `runAddSpinner` + `buildNewAgentSteps` + `buildAddItemsSteps` + rename `cmd/init_redesign.go` → `cmd/init_flow.go`.

## Cross-reference

- Status.md: Recently Done top row.
- memory.md: Work State + Main-at refreshed.
- Backlog.md: 4 new items in Group B.
- Plan file: `Playbook/Plans/Active/23-uiux-phase2-add.md` stays Active — Phase 3 pending.
