---
description: Re-grill an existing plan via 6 critic agents (5 prose + Reality), looped to convergence.
argument-hint: <plan-NN>
---

# /grill — Grill Existing Plan

Plan to grill: $ARGUMENTS

## Procedure
1. **Pre-flight** — resolve `station/Playbook/Plans/Active/NN-*.md` (refuse if it resolves to ≠1 file; never grill Archive plans). Confirm the plan is committed + pushed to `origin/main` (worktree critics base off origin/main). Confirm working tree clean. If the plan already has a `Grilling Pass` section, warn and ask before re-running.
2. **Load workflow** — read `station/agent/Workflows/plan-grilling.md` and follow it.
3. **Dispatch 6 critics** — single message, 6 `Agent` calls (security, architecture, simplicity, risk, verification, reality). All `subagent_type: general-purpose`, `isolation: worktree`, `run_in_background: true`. Copy each prompt verbatim from `station/agent/Skills/critic-agent-prompts.md`, repointing the plan path. Never template plan body into a prompt.
4. **Aggregate** — wait for all 6; build the verdict table.
5. **Batched user Q&A** — one round via `AskUserQuestion`: blocks first, then concerns. Separate genuine user-decisions from mechanical corrections (apply the latter yourself).
6. **Convergence loop** — apply resolutions, re-grill the edited plan, repeat until a round yields zero findings above note/info. Round cap 3 → split the plan.
7. **Lock** — append `## Grilling Pass — YYYY-MM-DD` (per-round verdict tables + resolutions), set `status: ready`, commit `plan(NN): grilled — N concerns, K resolved`, push if user OK.

## Skip cases
Trivial Tier-1 patches (single-file fix with test). If invoked on one, surface the rule and ask whether to proceed.
