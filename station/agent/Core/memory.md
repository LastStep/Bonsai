---
tags: [core, memory]
description: Tech Lead Agent working memory — flags, work state, notes.
---

# Working Memory

## Flags

<!-- Active flags go here. Format: - [FLAG] description -->

(none)

## Work State

**Current task:** Idle. Plan 15 iter 3 + 3.1 + 3.2 all shipped on `ui-ux-testing`; user confirmed dogfooding clean. Next session: merge `main` into branch (incorporates 6 commits including Plan 18) → push → PR squash-merge to main.
**In flight (other tracks):**
- Plan 15 (BubbleTea foundation) — on `ui-ux-testing` branch. All iters shipped locally. Latest: `4b35971` iter-3.2 OtherAgents sort, `b4cb17f` iter-3.1 doc reconcile, `39d9da3` iter-3.1 add.go index fix, `a406908` iter-3 full migration. Safety branches: `ui-ux-testing-pre-rebase` (`2fa91d0`) + `ui-ux-testing-pre-iter2-rebase` (`2d7a947`) — keep ~30d. Main merged into branch 2026-04-20 (pulls in Plan 18). Pending: push → PR squash-merge to main.
**Blocked on:** Nothing.
**Last completed:** Plan 15 iter 3.2 — user dogfooded iter 3 on Bonsai-Test and reported "all files being updated" on `bonsai update` even after a fresh init+add. Root cause: `AgentWorkspace` at `internal/generate/generate.go:1213` iterated `cfg.Agents` (a map) to build `ctx.OtherAgents`, producing nondeterministic order. 8 templates `range .OtherAgents` (every agent identity + scope-guard + dispatch-guard sensors) → re-renders had different bytes → every file flagged ActionUpdated. Pre-existing Group F bug, surfaced again. Fixed via `sort.Slice` by AgentType after the build loop. User also asked "where's the spinner" — answered: generation completes in ~50–150ms which is sub-frame for the 80ms spinner tick; working as designed but visually invisible at this speed. Optional follow-up: artificial minimum display time (deferred — no Backlog entry, raise if user mentions again).

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
- **PR CI ≠ main CI.** The `CI` workflow (test + lint + GitGuardian) runs on pull_request, but `Deploy Docs` only runs on `push` to main. So a broken `website/**` change can pass PR CI and then fail on main — happened 2026-04-20 with PR #25's MDX autolink. When reviewing PRs that touch `website/`, run `cd website && npm run build` locally before merging, or add a `website/**` path trigger to Deploy Docs on pull_request (separate backlog item).
- **MDX autolink gotcha.** `<https://example.com>` is valid GitHub-flavored markdown but MDX parses `<` as JSX — Astro build fails with "Unexpected character `/`". Always use `[label](url)` inside `.mdx` files. Incident: PR #25 `guide.mdx:20`, fixed in `e336ccb`.

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
