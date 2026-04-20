---
tags: [session-log, plan-15]
description: Plan 15 iter 3 + 3.1 + 3.2 ship — full BubbleTea harness migration.
---

# 2026-04-20 — Plan 15 iter 3 ship + 3.1/3.2 fixes

## What shipped (branch `ui-ux-testing`)

- **iter 3** (`a406908`) — Full harness migration of remove/update/init/add. New harness primitives: `ConditionalStep`, `SpinnerStep` (+ `NewSpinnerWithPrior`), `priorAware` interface, panic recovery for `Build`/`Splice`, WindowSize re-broadcast on splice. ~2.5k LOC including 444 lines of new tests. 5 implementer-flagged deviations: conflict-picker index correction in runRemove (3→2), unified `runAddSpinner`/`addOutcome` factoring, `cfg.Save` moved inside init's spinner closure (atomicity), "no picks" guard in add-items, helper extractions.
- **iter 3.1** code (`39d9da3`) + docs (`b4cb17f`) — Two parallel post-ship reviewers surfaced one real regression: `cmd/add.go:250` used literal `3` for the conflict-picker MultiSelect index, but the agent-flow LazyGroup splices a variable number of slots above it. User selections were silently discarded on every `bonsai add` that hit a real conflict. Fixed by computing `len(results)-2` when `wr.HasConflicts()`. 5 non-fix nits routed to Backlog Group F (ConditionalStep predicate-not-re-evaluated docstring drift, NewConditional nil-predicate guard, SpinnerStep action panic-recovery gap, "Tech Lead required" duplicate surface, `applyCustomFileSelection` no-dedup append).
- **iter 3.2** (`4b35971`) — User dogfooded on `~/Apps/Bonsai-Test` and reported "all files updated" on a clean `bonsai update`. Root cause was the long-standing Group F bug: `AgentWorkspace` at `internal/generate/generate.go:1213` iterated `cfg.Agents` (a map) to build `ctx.OtherAgents`. Eight templates `range .OtherAgents` (all six agent identities + scope-guard + dispatch-guard) → nondeterministic byte order → every re-render flagged `ActionUpdated`. Fixed with `sort.Slice` by AgentType after the build loop.

## Process notes

- **Two parallel review agents** for iter 3 (harness+remove vs update/init/add) replicated the iter 2 pattern that surfaced the iter 2.1 fixes. Worked again — caught a silent data-loss bug that no test would have caught (no cmd-level integration tests exist).
- **Group F resolved by iter 3**: spinner Ctrl-C partial-write, workspace validator normalization, Splice/Build panic recovery, ConditionalStep adapter (4 items commented out, not deleted, in Backlog).
- **Squash-bundle pattern** noted again: branch is now ~17 commits ahead of main, but main has 6 commits (Plan 18 + cleanups) ahead of the merge-base. Need to merge main into branch before the squash-merge to main.

## User feedback this session

- "Why no spinner?" — generation completes in <80ms (one spinner tick), so it's working but visually invisible. Optional artificial-minimum-display-time tweak deferred. Not filed to Backlog; will revisit if user mentions again.
- "Why are all files being updated?" — root-caused to OtherAgents nondeterminism, fixed in iter 3.2.

## Pending

- Merge `main` into `ui-ux-testing` (handles Plan 18's docs/cli.md → docs/triggers.md churn).
- Push branch, open PR, squash-merge to main.
- Overdue routines (Memory Consolidation, Status Hygiene — 6d overdue, 5d cadence) — defer to next session.
