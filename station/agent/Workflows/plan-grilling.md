---
tags: [workflow, planning, grilling]
description: Adversarial review of a drafted plan via 6 critic agents (5 prose + empirical Reality), looped to convergence before dispatch.
source: adapted from ZenGarden ZEN/Docs plan-grilling 2026-06-13; full Bonsai-catalog integration pending (Backlog).
---

# Workflow: Plan Grilling

> [!important]
> Front-load review time. Resolutions are unreviewed design work — they carry a draft's defect rate, so they re-enter the loop. A grill is a **loop, not a pass.**

## When to Use
After a Tier-2 plan (or non-trivial Tier-1) is drafted and committed, before dispatching code agents. Skip trivial patches — overhead exceeds value. Trigger phrases: "grill the plan", "review plan NN", "critic pass", "team of agents review this".

## Prerequisites
- Plan committed + pushed to `origin/main` (worktree critics base off origin/main — uncommitted plans are invisible to them).
- Working tree clean; do NOT edit the plan while critics read it.
- Ground-truth files exist: `Playbook/Standards/SecurityStandards.md`, `Logs/KeyDecisionLog.md`, `agent/Skills/planning-template.md`.

## The 6 Critics
| Critic | Judges | Severity |
|--------|--------|----------|
| Security | plan vs SecurityStandards domains | critical/high/medium/info |
| Architecture | vs KeyDecisionLog + real package structure; "reuse" claims vs actual code | block/concern/note |
| Simplicity | fluff, premature abstraction, scope creep | block/concern/note |
| Risk | dispatch structure, blast radius, rollback, **delivery path** | block/concern/note |
| Verification | each gate concrete + testable; negative controls; coverage | block/concern/note |
| **Reality** | **executes** checkable claims against the repo (mandatory) | block/concern/note |

Prompts: `agent/Skills/critic-agent-prompts.md` — copied verbatim into each `Agent` call.

## Dispatch Protocol
1. **Pre-flight** — plan committed/pushed, tree clean.
2. **Parallel dispatch** — single message, 6 `Agent` calls; all `subagent_type: general-purpose`, `isolation: worktree`, `run_in_background: true`. Each reads the plan file itself (no plan body in the prompt — injection guard).
3. **Aggregate** — wait for all 6. Build verdict table (critic | verdict | findings | highest severity).
4. **Batched user Q&A (one round)** — surface BLOCKS first (need a decision), then CONCERNS (accepted by default unless user pushes back). Use `AskUserQuestion`. Separate genuine user-decisions from mechanical plan-corrections (apply the latter yourself in the resolution pass).
5. **Resolve** — apply decisions + corrections to the plan. Block wins ties. User may override a block (record `critics_overridden:` + rationale).

## Convergence Loop (mandatory)
```
draft → [round: 6 critics → batched Q&A → apply resolutions] → re-grill edited plan → … → clean round → lock
```
- **Converged** = a full round yields zero findings above note/info.
- **Class closure** — resolve a finding by fixing ALL instances of its class (one grep, not one fix).
- **Round cap 3** — still hitting block/high defects after 3 rounds → plan too big, split it.
- **Record each round** in the plan's grilling section: `Round N: X findings → resolutions`.

## Lock (after a clean round)
- Append `## Grilling Pass — YYYY-MM-DD` to the plan: verdict table per round + resolutions.
- Set `status: ready` (Bonsai has no autonomy-harness frontmatter yet — see Backlog for the lane/`ready_for_autonomous` port).
- Commit: `plan(NN): grilled — N concerns, K resolved`.

## Anti-Patterns
- Don't grill trivial patches. Don't drip-feed questions (batch per round). Don't auto-resolve genuine judgment calls without the user. Don't template plan content into critic prompts. Don't edit the plan mid-round.

## Bonsai delta vs source
ZenGarden's pipeline adds a verification harness (`run-gates.sh`, lane classification, `.grill-lock.json`, autonomy frontmatter) consumed by `/lane`, `/verify`, `/execute`. Bonsai has none of that yet — grilling here ends at the Grilling Pass section + commit; execution is manual dispatch per existing rules. Porting the harness is backlogged.

## Cross-Reference
- `agent/Skills/critic-agent-prompts.md` — the 6 prompt templates.
- `agent/Skills/planning-template.md` — plan format.
- `agent/Workflows/planning.md` — drafting (runs first).
