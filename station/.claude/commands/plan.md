---
description: Full planning pipeline — research, draft a Tier-1/2 plan, then grill it via 6 critics to convergence.
argument-hint: <topic or task>
---

# /plan — Plan + Grill Pipeline

Topic: $ARGUMENTS

## Pipeline
1. **Research** — load `station/agent/Workflows/planning.md`; understand the request, research trade-offs, surface architectural questions. If the topic is ambiguous, ask ONE clarifying question before drafting.
2. **Draft** — write `station/Playbook/Plans/Active/NN-kebab-name.md` using `station/agent/Skills/planning-template.md` (Tier 2 for multi-domain work, Tier 1 for patches). Make every step specific (file paths, function names, data shapes) and every verification item concrete + testable. Commit + push the draft (worktree critics need it on `origin/main`).
3. **Grill** — run `/grill NN` (or follow `station/agent/Workflows/plan-grilling.md` directly): dispatch the 6 critics, aggregate, batched user Q&A, convergence loop.
4. **Lock** — append the `## Grilling Pass` section, set `status: ready`, commit.
5. **Hand off** — update `station/Playbook/Status.md`; tell the user the plan is ready for dispatch.

## Constraints
- One plan per invocation.
- Never skip pre-flight (clean tree, committed plan).
- Never auto-resolve a genuine judgment call without the user.
- Grilling cost ≈ 6 agents × 2–3 rounds. For low-impact plans, confirm with the user before dispatching critics.
