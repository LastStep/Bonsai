---
tags: [log, session, plan-19, oss-release, bugfix, harness]
description: Session log — Plan 19 pre-launch bug sweep shipped to PR #27 (pending merge).
date: 2026-04-21
---

# 2026-04-21 — Plan 19: Pre-launch Bug Sweep

## Goal

Ship eight fresh-install blocker + harness polish fixes as one PR before OSS launch, following the `issue-to-implementation` workflow end-to-end under user-granted authority ("act with authority, unsupervised").

## Flow

1. **Intake + analysis.** Read Backlog.md, grouped 8 items (4 Tier 1 blockers in Group F, 4 Tier 2 harness items in Group B) into one bundle. Deferred 10+ items (docs audit, demo GIF, test coverage expansion, etc.) to future sessions.
2. **Research.** Read `cmd/*.go`, `internal/tui/harness/{harness,steps}.go`, `internal/generate/generate.go`. Verified: all 5 cmd files already import `errors` (so `errors.Join` drops in clean), all generator functions return `error`, existing `TestConditionalStepResetReevaluates` exercises the post-Reset Init path (so the Esc-back fix mustn't break it), no `.gitattributes` in repo.
3. **Plan 19 drafted** — Tier 2, 283 lines, 8 steps with explicit file paths / line numbers / code sketches / per-step acceptance criteria. Committed as `90d6ec3` before dispatch.
4. **Worktree-isolated dispatch** via `Agent(subagent_type: general-purpose, isolation: worktree)` with explicit "act with authority" briefing — forbid the agent from committing/pushing, require go vet + gofmt + go test to pass before handing back.
5. **Agent reported 13 files changed, +453 / -73, all tests green** on branch `worktree-agent-ab0d06b6`. Returned under 15 minutes.
6. **Verification pass** — read each modified file at the touched region, confirmed:
    - Step 1 (CRLF): `.gitattributes` matches plan, `normalizeShellLF` applied in both `writeFile` + `writeFileChmod`.
    - Step 2 (showWriteResults): `splitTopSegment` bucket-by-first-segment, alpha-sort, all 4 callers dropped the `rootLabel` param.
    - Step 3 (dedup): `appendUnique` helper, applied across all 5 switch cases in `applyCustomFileSelection`.
    - Step 4 (errors.Join): all 5 spinner sites migrated; callers in init/remove/update now surface via `tui.Warning`.
    - Step 5 (spinner recover): `defer func() { recover() → spinnerDoneMsg{err} }` pattern, `fmt.Errorf("spinner action panic: %v", r)`.
    - Step 6 (NewConditional nil guard): nil → `func(prev []any) bool { return true }` (safer default).
    - Step 7 (Esc-back re-eval): harness `SetPrior(priorResults())` before `Reset()`, Conditional.Reset re-evaluates predicate.
    - Step 8 (add duplicate message): NoteStep replaced with `return nil`, post-harness ErrorDetail preserved.
7. **Re-verified** — `go build`, `go vet`, `gofmt -l`, `go test ./...` all clean in the worktree. 6 new tests + 3 subtests pass.
8. **Rebased** worktree onto main (fast-forward, no conflicts — plan doc commit `90d6ec3` is docs-only), renamed branch to `plan-19-bug-sweep`, pushed.
9. **Pushed main** (`5e9255f..90d6ec3` — plan doc + Plan 15 cleanup log that was stale locally).
10. **Opened PR #27** with full 8-item checklist and file diff-stat in the body.
11. **CI watching** in background via `gh run watch`.

## State at EoS

- **Main:** `90d6ec3` (`origin/main` in sync).
- **Open PR:** [#27](https://github.com/LastStep/Bonsai/pull/27) — plan-19-bug-sweep, 13 files +541 / -73.
- **Worktree:** `/home/rohan/ZenGarden/Bonsai/.claude/worktrees/agent-ab0d06b6` — keep alive until PR merges, then `git worktree remove`.
- **Status.md:** Plan 19 moved to In Progress.

## Carry-Forward / Next Session

- **Watch PR #27 CI** — if lint fails, most likely culprit is a `gofmt` drift or an unused import from the dedup helper (though local gofmt was clean). CI v1.64.8 is stricter than local v2.11.4.
- **On merge** — move Plan 19 to Recently Done in Status.md, update Backlog to strike off the 8 items (Group F items 1-3, Group B items covered), remove worktree, delete local branch.
- **Deferred items** (still in Backlog): pre-release docs audit, demo GIF, `generate.go` split, catalog/cmd test coverage, trigger test infra, PTY smoke test, GO-2026-4602 stdlib monitor, Group D catalog expansion, Group E workspace QoL, stale worktree sweep (15+ under `.claude/worktrees/`).
- **Plan 08 Phase C (new sensors)** still paused — unblocked now that UI/UX + harness series is wrapping.

## Notable Decisions

- **Bundle 8 items into one PR, not 8 PRs.** Rationale: these are all small mechanical fixes, high-overlap test scope (fresh-install golden path), low conflict risk. One PR is cheaper to review end-to-end than 8 interleaved PRs where reviewers lose the forest for the trees. Plan 15 split into 3 iterations because it was a genuine multi-domain rewrite; Plan 19 is a cleanup.
- **Drop `rootLabel` param from `showWriteResults` entirely** rather than making it optional. Rationale: every caller was passing a static string anyway; the function can derive roots from `wr.Files` with zero ambiguity. Dropping the param forces callers off the stale pattern and prevents the bug class (single-root assumption) from recurring.
- **Keep `os.Remove` `_ =` swallows** in `cmd/remove.go` (L451, 457, 463, 464) + backup writes in `cmd/root.go:157`. Rationale: these are best-effort filesystem cleanups where ENOENT is expected / irrelevant. Aggregating them into `errors.Join` would produce user-facing noise for a case that's fine.
- **`NewConditional(nil)` defaults to show, not skip.** Rationale: silent-skip is a footgun — a ConditionalStep that disappears on nil predicate masks the bug and produces a smaller output instead of a louder one. Defaulting to show means the bug is visible (step runs when it shouldn't) → faster fix.
