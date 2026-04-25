---
date: 2026-04-25
plan: 32
pr: 80
commit: 99e9948
tags: [session-log, plan-32, knock-off-bundle, dispatch-rate-limit-recovery]
---

# Plan 32 — Plan 31+29 review followup bundle

**Tier-1 knock-off, single-PR bundle. 6 phases, 13 of 17 review items closed. PR #80 squash `99e9948`.**

## Phase summary

| Phase | Commit | Closes | Why |
|-------|--------|--------|-----|
| A | `20a8f76` | Plan-29-cosmetic 3, Plan-29-sec-hardening 1 | Extract `internal/wsvalidate/` shared package; migrate addflow/ground.go + initflow/vessel.go validate(); error string clarity |
| B | `7c792c1` | Plan-29-sec-hardening 2, 3 | Backslash + pure-root rejection in `wsvalidate.InvalidReason` + 6 unit tests |
| C | `053c6bc` | Plan-29-cosmetic 1, 2; Plan-29-test-gap 1 | Keep-vs-Backup tone assertion + positive companion tests + shortName→conflictsShortName rename |
| D | `d7b6fde` | Plan-31-cosmetic 4, 5 | hasAbility → slices.Contains; agentsToSlice + requiredToSlice → compatToSlice |
| E | `12707d0` | Plan-31-test-gap 1, 2 | Snapshot trailing-newline assertion + version pass-through table test |
| F | `42252b4` | Plan-31-sec-hardening 1, 2, 4 | ProjectConfig.Validate() chokepoint + O_NOFOLLOW symlink-resistant write + 5 config tests |
| (lint-fix) | `bd9b12f` | n/a | `_ = f.Close()` errcheck + gofmt blank-line strip |

## Process notes

**Dispatch agent #1 crashed on rate limit.** Mid-Phase-C, agent `a04bedaaeb91a4c3d` returned `result: "You've hit your limit · resets 7:30pm (Asia/Calcutta)"` after 2 commits + uncommitted-passing tests. Worktree branch state inspected via `git log --oneline main..HEAD` + `git status`: A + B committed clean, Phase C tests written + passing but unstaged.

**Recovery pattern that worked.** Dispatched continuation agent #2 with explicit "Step 0: commit Phase C as-is" instruction targeting the existing worktree path (no new worktree creation). Agent shipped Phase C-F + 4 commits clean. This pattern is reusable: when a long-running dispatch hits rate limit mid-flight, inspect the worktree, identify completed-but-uncommitted work, and brief continuation agent with explicit commit-then-continue steps.

**CI lint failed first push.** golangci-lint v2.11.4 caught two issues local `go test ./...` did not: errcheck on `f.Close()` in error path of `WriteCatalogSnapshot`, and gofmt blank line in `ground_test.go:131`. Local lint disabled (config v1, binary v2 mismatch — see memory.md Notes). Lesson: trust CI for lint, OR install matching golangci-lint version locally before claiming green.

**Auto-merge fired immediately on CodeQL green.** `gh pr merge 80 --squash --delete-branch --auto` set up the queue; squash-merge happened ~1min later when Analyze Go finished. Worktree-held-branch cleanup pattern hit 18× — manual `git worktree remove -f -f` + `git branch -D` after merge.

## Code review notes (filed informational, not blockers)

1. **Co-Authored-By trailer split.** Phase A+B (first agent) included the trailer; Phase C-F (continuation agent, my prompt told it not to) did not. Inconsistent within single PR but functional. Future dispatch prompts should match repo convention (CLAUDE.md: include trailer).
2. **O_NOFOLLOW protects only final path component.** `.bonsai/` dir itself unprotected against symlink-via-parent-dir attacks. Closing this requires `openat`-style traversal — out of scope for v0.4. Acceptable per threat model.
3. **`agent.AgentType` + ability-name lists not metachar-scanned.** Validate() only scans ProjectName, agent map keys, agent.Workspace, DocsPath. Catalog-controlled identifiers (kebab-case enforced at install time) so safe.

## Out-of-scope items remaining (filed in Backlog)

- Plan-29-test-gap 2 — `TestGenerateStage_BodyOnlyDropsChrome` inverse companion
- Plan-29-sec-hardening 4 — Unicode lookalike (NFKC normalisation before `..` scan)
- Plan-31-cosmetic 1, 2, 3, 6 — octal style, WriteCatalogSnapshot/writeFile dedup, bonsai_reference link labels, render-cost benchmark
- Plan-31-sec-hardening 3 — TOCTOU on `.bonsai/` dir perms

All low-value, non-blocking, can age out or pick up in next knock-off sweep.
