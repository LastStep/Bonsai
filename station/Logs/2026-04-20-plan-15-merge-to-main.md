---
tags: [log, session, plan-15]
description: Session log — merge ui-ux-testing → main via PR #26 squash.
date: 2026-04-20
---

# 2026-04-20 — Plan 15 Merge to Main (PR #26)

## Goal

Ship Plan 15 (BubbleTea harness migration) from `ui-ux-testing` branch to `main`.

## Flow

1. **Merge `main` → `ui-ux-testing`** to incorporate 6 commits from main (Plan 18, Plan 16, routine archives, docs).
   - Auto-merged: `cmd/root.go`, `Backlog.md`.
   - Conflicts: `Status.md`, `memory.md` — both in the Work-State / In-Progress sections. Kept HEAD's Plan 15 iter 3.2 detail; folded main's Plan 18 completion + post-merge hotfix into Notes.
   - Merge commit: `d060020`.
   - Verified: `make build` + `go test ./...` both green.

2. **Push `ui-ux-testing` to origin** — new remote branch.

3. **Create PR #26** — `feat(tui): BubbleTea harness — migrate init/add/remove/update (Plan 15)`.

4. **CI run 1 — lint failed.** `cmd/root.go:90 resolveConflicts is unused`. Legacy pre-harness function left behind from iter 3 migration (replaced by `buildConflictSteps`). Local `make build`, `go test ./...`, and `go vet ./...` had all passed — only golangci-lint v1.64.8 on CI caught it.
   - Fix: deleted `resolveConflicts` + the now-dead `github.com/charmbracelet/huh` import (`070bcd0`).

5. **CI run 2 — all green** (lint + test + GitGuardian).

6. **Squash-merge.** `gh pr merge --squash --delete-branch` succeeded on the GitHub side (squash `2ce63f6` on main, remote branch deleted eventually) but errored locally with `'main' is already checked out at '/home/rohan/ZenGarden/Bonsai'` (sibling worktree blocks auto-checkout). Followed up:
   - `git fetch --prune origin` — confirmed squash on `origin/main`.
   - `git push origin --delete ui-ux-testing` — remote branch actually gone.
   - Local `ui-ux-testing` branch still checked out here in the `Bonsai-uiux` worktree.

## Learnings (added to memory Notes)

- **Local Go toolchain misses `unused` findings that CI flags** — `go build/test/vet` pass but `golangci-lint v1 unused` doesn't. Local `golangci-lint` here is v2; repo config is v1; local lint errors out with "unsupported version". Either install v1 binary for pre-push OR trust CI to catch post-refactor dead code. Added to memory.
- **`gh pr merge --squash --delete-branch` is not atomic across remote+local when run from a non-main worktree with a sibling worktree holding main.** Remote branch deletion may silently fail; verify after. Added to memory.

## Cleanup Left for User

- **Local `ui-ux-testing` branch + `Bonsai-uiux` worktree** — orphan now that Plan 15 shipped. Remove via:
  ```
  cd /home/rohan/ZenGarden/Bonsai
  git worktree remove /home/rohan/ZenGarden/Bonsai-uiux
  git branch -D ui-ux-testing
  ```
- **Safety branches** `ui-ux-testing-pre-rebase` (`2fa91d0`) + `ui-ux-testing-pre-iter2-rebase` (`2d7a947`) — keep ~30d per prior convention; delete around 2026-05-20.
- **Doc updates on `ui-ux-testing`** (this log, Status.md move-to-done, memory.md post-merge state) — committed on the now-dead branch. Re-apply in main worktree next session, OR cherry-pick the wrap-up commit after fetching.

## State at EoS

- **Main:** `2ce63f6` — Plan 15 squash-merged.
- **This worktree:** `ui-ux-testing` at `070bcd0` + any wrap-up commit.
- **Remote branches deleted:** `ui-ux-testing`, `docs/starlight-scaffold`, `plan-13-actionunchanged-followups`.
- **Next session:** consolidate to main worktree; delete this worktree + dead branch; re-apply wrap-up docs on main if not cherry-picked.
