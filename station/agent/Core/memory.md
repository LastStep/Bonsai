---
tags: [core, memory]
description: Tech Lead Agent working memory — flags, work state, notes.
---

# Working Memory

## Flags

<!-- Active flags go here. Format: - [FLAG] description -->

(none)

## Work State

**Current task:** Plan 17 (release prep — Go toolchain + triggerSection + OSS polish) — draft [PR #24](https://github.com/LastStep/Bonsai/pull/24) on `release-prep` branch. Dispatch + independent review PASS. Awaiting CI + merge-to-`release-prep` (NOT main). Plan 16 PR #23 still awaiting user merge.
**Blocked on:** Plan 17 → CI verification on PR #24, then user merge to `release-prep`. Plan 16 → user merge on PR #23.
**Last completed:** Plan 17 dispatch + review (commit `5451196` on `plan-17/release-prep`) (2026-04-17)

## Notes

<!-- Session-to-session notes. Keep concise. -->

- **Plan 14 still open for iteration.** Local-only, no PR. Taste-heavy — user drives iterations batch by batch. Iteration 2 (2026-04-17) shipped A/B/C: `mustCwd()` error surfacing in all cmd files, collapsed required-only chip line in `PickItems`, width-aware `TitledPanel` with `ansi.Truncate`. Deferred: palette rebalance (D), heading rhythm / step separators (E) — pending user steer. Phase 4+ (screen lifecycle, progressive disclosure, go-back nav, flow redesign) still in Plan 14 "Out of Scope".
- **Plan 08 Phase C (new sensors) paused** — moved back to Pending while Plan 14 ships. Resume once UI/UX overhaul series wraps or explicitly requested.
- **Pre-flight learning:** Worktrees inherit only committed HEAD — uncommitted plans/docs in main tree are invisible to dispatched agents. Commit station/ planning artifacts before dispatch.
- **PR review memory hygiene:** "both reviews APPROVE" from prior session was dispatched review agents, not GitHub reviews. `gh pr view --json reviews` returned empty. When noting review status, distinguish agent-dispatched reviews (in `Reports/`) from GitHub formal reviews.
- **Subagent tool inheritance is flaky.** On 2026-04-17, the original executing agent (worktree, `gh` authed) successfully created draft PR #23, but every subsequent subagent I spawned for merge/verification reported `gh: command not found` — the environment wasn't inherited from the spawning agent or from my own shell (Windows Git Bash over WSL UNC path, no `gh` on PATH). Implication: for PR-flow tasks, bundle **all** gh operations (push, create PR, mark ready, merge, delete branch) into a **single** agent dispatch rather than splitting across agents. Second-best: ask the user to click-merge via web UI when a subagent-less step is needed.
- **Plan 17 confirmed the gh-inheritance pain persists even with user's WSL gh auth.** User installed/authed `gh` in their own WSL on 2026-04-17, but dispatched subagents (tested via `which gh`, `/usr/bin/gh`, `/usr/local/bin/gh`, `/mnt/c/Program Files/GitHub CLI/gh.exe`) still couldn't find any `gh` binary. Subagents evidently run in a separate filesystem namespace from the user's WSL — installing tools in user WSL doesn't help subagents. For now: user click-creates PRs via web, or user runs `gh` locally from their terminal. Ongoing research.
- **Parallel session coexistence (2026-04-17).** Two Claude sessions ran concurrently: one on `ui-ux-testing` (Plan 15 iter 1), one on `release-prep` (Plan 17 / PR #24). Branches touch disjoint file sets (UI/UX on `internal/tui/*` + `cmd/*`; release-prep on `go.mod` + `internal/generate/*` + tooling), so no merge conflicts. Memory and Status.md updates MUST stay branch-scoped — don't cross-pollinate Plan 15 notes into release-prep's planning docs or vice versa, or rebase-to-main later gets messy.

## Feedback

<!-- User corrections and confirmed approaches that persist across sessions. Only record what isn't already in CLAUDE.md, workflows, or protocols. -->

- **Backlog.md at session start is a P0 *scan*, not a full read.**
    - **Why:** During the 2026-04-17 context audit, the CLI session-start payload was ~4k heavier than the app's, and the delta was traced to my having `Read` Backlog.md in full (~200 lines, ~4k tokens) when session-start protocol explicitly says "scan, look for P0 items only." Full backlog review is the backlog-hygiene routine's job. On a 200k window that's ~4% of headroom burned for zero benefit.
    - **How to apply:** At session start, use `Grep -n "^## P0" Playbook/Backlog.md` (or read with `limit:` bracketed to the P0 section) instead of `Read` on the whole file. Only full-read Backlog when explicitly running backlog-hygiene or planning across multiple groups.

## References

<!-- Pointers to external resources not documented elsewhere in the project. -->

_(empty)_
