---
tags: [core, memory]
description: Tech Lead Agent working memory — flags, work state, notes.
---

# Working Memory

## Flags

<!-- Active flags go here. Format: - [FLAG] description -->

(none)

## Work State

**Current task:** Plan 16 (go install binary name fix) — draft [PR #23](https://github.com/LastStep/Bonsai/pull/23) open, awaiting user merge. Plan 15 (BubbleTea foundation) still drafted, awaiting dispatch.
**Blocked on:** Plan 16 → user merge on PR #23 (subagents lacked `gh`; user will click-merge via web)
**Last completed:** Plan 14 iteration 2 — width-aware TitledPanel, collapsed required-only chip line, `mustCwd()` error surfacing (`63a3709`) (2026-04-17)

## Notes

<!-- Session-to-session notes. Keep concise. -->

- **Plan 14 still open for iteration.** Local-only, no PR. Taste-heavy — user drives iterations batch by batch. Iteration 2 (2026-04-17) shipped A/B/C: `mustCwd()` error surfacing in all cmd files, collapsed required-only chip line in `PickItems`, width-aware `TitledPanel` with `ansi.Truncate`. Deferred: palette rebalance (D), heading rhythm / step separators (E) — pending user steer. Phase 4+ (screen lifecycle, progressive disclosure, go-back nav, flow redesign) still in Plan 14 "Out of Scope".
- **Plan 08 Phase C (new sensors) paused** — moved back to Pending while Plan 14 ships. Resume once UI/UX overhaul series wraps or explicitly requested.
- **Pre-flight learning:** Worktrees inherit only committed HEAD — uncommitted plans/docs in main tree are invisible to dispatched agents. Commit station/ planning artifacts before dispatch.
- **PR review memory hygiene:** "both reviews APPROVE" from prior session was dispatched review agents, not GitHub reviews. `gh pr view --json reviews` returned empty. When noting review status, distinguish agent-dispatched reviews (in `Reports/`) from GitHub formal reviews.
- **Subagent tool inheritance is flaky.** On 2026-04-17, the original executing agent (worktree, `gh` authed) successfully created draft PR #23, but every subsequent subagent I spawned for merge/verification reported `gh: command not found` — the environment wasn't inherited from the spawning agent or from my own shell (Windows Git Bash over WSL UNC path, no `gh` on PATH). Implication: for PR-flow tasks, bundle **all** gh operations (push, create PR, mark ready, merge, delete branch) into a **single** agent dispatch rather than splitting across agents. Second-best: ask the user to click-merge via web UI when a subagent-less step is needed.

## Feedback

<!-- User corrections and confirmed approaches that persist across sessions. Only record what isn't already in CLAUDE.md, workflows, or protocols. -->

- **Backlog.md at session start is a P0 *scan*, not a full read.**
    - **Why:** During the 2026-04-17 context audit, the CLI session-start payload was ~4k heavier than the app's, and the delta was traced to my having `Read` Backlog.md in full (~200 lines, ~4k tokens) when session-start protocol explicitly says "scan, look for P0 items only." Full backlog review is the backlog-hygiene routine's job. On a 200k window that's ~4% of headroom burned for zero benefit.
    - **How to apply:** At session start, use `Grep -n "^## P0" Playbook/Backlog.md` (or read with `limit:` bracketed to the P0 section) instead of `Read` on the whole file. Only full-read Backlog when explicitly running backlog-hygiene or planning across multiple groups.

## References

<!-- Pointers to external resources not documented elsewhere in the project. -->

_(empty)_
