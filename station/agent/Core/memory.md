---
tags: [core, memory]
description: Tech Lead Agent working memory — flags, work state, notes.
---

# Working Memory

## Flags

<!-- Active flags go here. Format: - [FLAG] description -->

(none)

## Work State

**Current task:** None in-session. Awaiting next pickup.
**In flight (other tracks):**
- Plan 15 (BubbleTea foundation) — on `ui-ux-testing` branch, parallel session driving iter 2/3. **Rebased onto current main 2026-04-20.** Branch is now `main + 3 commits` (Plan 15 iter 1 only: harness `8aab2fd`, Esc-back fix `86e57f0`, reports `2d7a947`). Plan 14 commits dropped (already on main via Plan 17 squash bundle). Old tip preserved as `ui-ux-testing-pre-rebase` (`2fa91d0`) — keep ~30d for safety. Build + tests verified post-rebase. **Dedicated worktree at `/home/rohan/ZenGarden/Bonsai-uiux` (created 2026-04-20)** — Plan 15 iter 2/3 runs in a separate Claude session there; main-branch work runs here in `/home/rohan/ZenGarden/Bonsai`.
**Blocked on:** Nothing in this session.
**Last completed:** Plan 16 merged to main via PR #23 (`28d181e`, 2026-04-20). Moved `main.go` → `cmd/bonsai/main.go` + introduced root `embed.go` package (keeps `//go:embed catalog/` and `docs/custom-files.md` resolving from repo root). `go install github.com/LastStep/Bonsai/cmd/bonsai@latest` now produces lowercase `bonsai` binary on PATH. Verified locally before merge.

## Notes

<!-- Session-to-session notes. Keep concise. -->

- **Plan 14 code shipped to main via PR #24 bundle (2026-04-17).** Iterations 1 + 2 (palette tokens, banner, answered-prompt persistence, required-only feedback, category counts, prompt polish, mustCwd error surfacing, width-aware TitledPanel) merged along with Plan 17's release-prep work. Phase 4+ scope (screen lifecycle, progressive disclosure, go-back nav, flow redesign) still deferred — Plan 15's BubbleTea harness is the likely vehicle for those.
- **Parallel session convention.** `ui-ux-testing` branch is where Plan 15 iterations land (separate session). Don't cross-pollinate planning-doc edits between branches — each branch's Status.md/memory.md should stay scoped to its own track. Merging to main collapses everything cleanly.
- **CI lint gotcha.** `golangci/golangci-lint-action@v6` with `version: latest` resolves to v1.x, not v2. `.golangci.yml` must use v1 schema (`linters.disable-all: true`, no `version:`/`formatters:` keys). If we ever move to v2, update both the action pin (`version: v2.x`) and the config schema together. Learned from PR #24 `66a6304`.
- **Plan 08 Phase C (new sensors) paused** — moved back to Pending while Plan 14 ships. Resume once UI/UX overhaul series wraps or explicitly requested.
- **Pre-flight learning:** Worktrees inherit only committed HEAD — uncommitted plans/docs in main tree are invisible to dispatched agents. Commit station/ planning artifacts before dispatch.
- **PR review memory hygiene:** "both reviews APPROVE" from prior session was dispatched review agents, not GitHub reviews. `gh pr view --json reviews` returned empty. When noting review status, distinguish agent-dispatched reviews (in `Reports/`) from GitHub formal reviews.
- **Squash-bundle rebase pattern.** When main has absorbed a feature branch's commits via squash-merge (e.g., Plan 14 individual commits → bundled into Plan 17's `bc565bf` squash), straight `git rebase main feature-branch` will conflict on every doc snapshot because main has moved past those states even though the *code* matches. Cleaner approach: `git reset --hard main && git cherry-pick <only-the-truly-new-commits>` after listing each commit and asking "is this code/content already on main via the squash?" Always preserve old tip as `<branch>-pre-rebase` first. Used 2026-04-20 to rebase `ui-ux-testing` (13 commits → 3 commits actually carried forward).
- **Worktree cwd gotcha from Claude Code.** `git worktree add ../<name> <branch>` runs in the bash tool's cwd, which is often a subdirectory (for us: `station/`), not the repo root. So `../Bonsai-uiux` resolves to `/home/rohan/ZenGarden/Bonsai/Bonsai-uiux` (nested inside the main worktree — violates the no-nesting rule), not the intended sibling `/home/rohan/ZenGarden/Bonsai-uiux`. **Always use absolute paths when creating worktrees from this session.** Hit this 2026-04-20 setting up the Plan 15 parallel-session worktree; caught, removed, recreated at the correct sibling path.
- **Subagent tool inheritance is flaky.** On 2026-04-17, the original executing agent (worktree, `gh` authed) successfully created draft PR #23, but every subsequent subagent I spawned for merge/verification reported `gh: command not found` — the environment wasn't inherited from the spawning agent or from my own shell (Windows Git Bash over WSL UNC path, no `gh` on PATH). Implication: for PR-flow tasks, bundle **all** gh operations (push, create PR, mark ready, merge, delete branch) into a **single** agent dispatch rather than splitting across agents. Second-best: ask the user to click-merge via web UI when a subagent-less step is needed.

## Feedback

<!-- User corrections and confirmed approaches that persist across sessions. Only record what isn't already in CLAUDE.md, workflows, or protocols. -->

- **Backlog.md at session start is a P0 *scan*, not a full read.**
    - **Why:** During the 2026-04-17 context audit, the CLI session-start payload was ~4k heavier than the app's, and the delta was traced to my having `Read` Backlog.md in full (~200 lines, ~4k tokens) when session-start protocol explicitly says "scan, look for P0 items only." Full backlog review is the backlog-hygiene routine's job. On a 200k window that's ~4% of headroom burned for zero benefit.
    - **How to apply:** At session start, use `Grep -n "^## P0" Playbook/Backlog.md` (or read with `limit:` bracketed to the P0 section) instead of `Read` on the whole file. Only full-read Backlog when explicitly running backlog-hygiene or planning across multiple groups.

## References

<!-- Pointers to external resources not documented elsewhere in the project. -->

- **Foundational research docs** — Anchor for methodology/concept decisions. Reference these when improvements need grounding in the core philosophy (ambient behavior, meta-layer, talents, evals).
    - [Research/RESEARCH-landscape-analysis.md](../../Research/RESEARCH-landscape-analysis.md) — How Bonsai compares to GSD/ECC/others; Bonsai's identity/coordination layer positioning
    - [Research/RESEARCH-concept-decisions.md](../../Research/RESEARCH-concept-decisions.md) — Ambient vs. command-driven, authority hierarchy, catalog ownership, talents taxonomy
    - [Research/RESEARCH-eval-system.md](../../Research/RESEARCH-eval-system.md) — Eval system concept: scenarios, evaluators, benchmarks for methodology rigor
    - [Research/RESEARCH-trigger-system.md](../../Research/RESEARCH-trigger-system.md) — Trigger section design research
    - [Research/RESEARCH-uiux-overhaul.md](../../Research/RESEARCH-uiux-overhaul.md) — UI/UX overhaul research (Plan 14/15 origin)
