---
tags: [core, memory]
description: Tech Lead Agent working memory — flags, work state, notes.
---

# Working Memory

## Flags

<!-- Active flags go here. Format: - [FLAG] description -->

(none)

## Work State

**Current task:** Plan 15 iter 3 + 3.1 shipped on `ui-ux-testing`. Pending: final audit pass + merge `ui-ux-testing` → `main` (whole branch as one bundle per Plan 15 convention).
**In flight (other tracks):**
- Plan 15 (BubbleTea foundation) — on `ui-ux-testing` branch. Iter 1 + iter 2 + iter 2.1 + iter 3 + iter 3.1 all shipped locally (no PR yet; whole branch merges to main on completion). Latest commits: `1f55d06` iter-3 plan, `a406908` iter-3 implementation (full harness migration of remove/update/init/add), iter-3.1 conflict-picker index fix pending commit. Safety branches: `ui-ux-testing-pre-rebase` (`2fa91d0`) + `ui-ux-testing-pre-iter2-rebase` (`2d7a947`) — keep ~30d.
**Blocked on:** Nothing.
**Last completed:** Plan 15 iter 3.1 — two independent post-ship reviewers of iter 3 surfaced 1 real regression in `cmd/add.go:250` (literal `3` for conflict-picker index discarded user's selections in `bonsai add`; correct value is `len(results)-2` since the agent-flow splice produces variable-length results). Fixed; `go build`/`vet`/`fmt`/`test` all clean. 5 non-fix nits filed to Backlog Group F: ConditionalStep predicate-not-re-evaluated docstring/code mismatch, NewConditional nil-predicate guard, SpinnerStep action panic-recovery gap, "Tech Lead required" duplicate surface in `bonsai add`, `applyCustomFileSelection` no-dedup. Group F items resolved by iter 3 (commented out): spinner Ctrl-C partial-write, workspace validator normalization, panic recovery around Splice/Build, ConditionalStep adapter.

## Notes

<!-- Session-to-session notes. Keep concise. -->

- **Plan 14 code shipped to main via PR #24 bundle (2026-04-17).** Iterations 1 + 2 (palette tokens, banner, answered-prompt persistence, required-only feedback, category counts, prompt polish, mustCwd error surfacing, width-aware TitledPanel) merged along with Plan 17's release-prep work. Phase 4+ scope (screen lifecycle, progressive disclosure, go-back nav, flow redesign) still deferred — Plan 15's BubbleTea harness is the likely vehicle for those.
- **Parallel session convention.** `ui-ux-testing` branch is where Plan 15 iterations land (separate session). Don't cross-pollinate planning-doc edits between branches — each branch's Status.md/memory.md should stay scoped to its own track. Merging to main collapses everything cleanly.
- **CI lint gotcha.** `golangci/golangci-lint-action@v6` with `version: latest` resolves to v1.x, not v2. `.golangci.yml` must use v1 schema (`linters.disable-all: true`, no `version:`/`formatters:` keys). If we ever move to v2, update both the action pin (`version: v2.x`) and the config schema together. Learned from PR #24 `66a6304`.
- **Plan 08 Phase C (new sensors) paused** — moved back to Pending while Plan 14 ships. Resume once UI/UX overhaul series wraps or explicitly requested.
- **Pre-flight learning:** Worktrees inherit only committed HEAD — uncommitted plans/docs in main tree are invisible to dispatched agents. Commit station/ planning artifacts before dispatch.
- **PR review memory hygiene:** "both reviews APPROVE" from prior session was dispatched review agents, not GitHub reviews. `gh pr view --json reviews` returned empty. When noting review status, distinguish agent-dispatched reviews (in `Reports/`) from GitHub formal reviews.
- **Squash-bundle rebase pattern.** When main has absorbed a feature branch's commits via squash-merge (e.g., Plan 14 individual commits → bundled into Plan 17's `bc565bf` squash), straight `git rebase main feature-branch` will conflict on every doc snapshot because main has moved past those states even though the *code* matches. Cleaner approach: `git reset --hard main && git cherry-pick <only-the-truly-new-commits>` after listing each commit and asking "is this code/content already on main via the squash?" Always preserve old tip as `<branch>-pre-rebase` first. Used 2026-04-20 to rebase `ui-ux-testing` (13 commits → 3 commits actually carried forward).
- **Subagent tool inheritance is flaky.** On 2026-04-17, the original executing agent (worktree, `gh` authed) successfully created draft PR #23, but every subsequent subagent I spawned for merge/verification reported `gh: command not found` — the environment wasn't inherited from the spawning agent or from my own shell (Windows Git Bash over WSL UNC path, no `gh` on PATH). Implication: for PR-flow tasks, bundle **all** gh operations (push, create PR, mark ready, merge, delete branch) into a **single** agent dispatch rather than splitting across agents. Second-best: ask the user to click-merge via web UI when a subagent-less step is needed.

## Feedback

<!-- User corrections and confirmed approaches that persist across sessions. Only record what isn't already in CLAUDE.md, workflows, or protocols. -->

- **Backlog.md at session start is a P0 *scan*, not a full read.**
    - **Why:** During the 2026-04-17 context audit, the CLI session-start payload was ~4k heavier than the app's, and the delta was traced to my having `Read` Backlog.md in full (~200 lines, ~4k tokens) when session-start protocol explicitly says "scan, look for P0 items only." Full backlog review is the backlog-hygiene routine's job. On a 200k window that's ~4% of headroom burned for zero benefit.
    - **How to apply:** At session start, use `Grep -n "^## P0" Playbook/Backlog.md` (or read with `limit:` bracketed to the P0 section) instead of `Read` on the whole file. Only full-read Backlog when explicitly running backlog-hygiene or planning across multiple groups.

## References

<!-- Pointers to external resources not documented elsewhere in the project. -->

_(empty)_
