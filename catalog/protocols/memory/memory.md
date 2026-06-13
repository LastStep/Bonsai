---
tags: [protocol, memory]
description: How to read, write, and clean working memory.
---

# Protocol: Working Memory

---

## Three Surfaces — Where a Write Belongs

Persistent project state lives on three distinct surfaces. Before writing
anything, decide which one it belongs to — they are not interchangeable:

| Surface | What it holds | Where it lives |
|---------|---------------|----------------|
| **Working memory** | Session/ephemeral agent state — flags, work-in-progress, cross-session notes | `agent/Core/memory.md` (this protocol) |
| **Durable memory graph** | Long-lived, linkable record — **decisions**, facts, durable notes | `{memory_dir}/{decisions,notes}/<permalink>.md` |
| **Logs** | Narrative/process history — what happened, in order | `Logs/` |

Routing rules:

- **A decision was made** → write a durable **decision note** to
  `{memory_dir}/decisions/<permalink>.md`, following the frozen note schema in
  `Playbook/Standards/NoteStandards.md` (frontmatter: `schema_version`,
  `permalink`, `scope`, etc. + Observations / Relations). Link it from the
  `MEMORY.md` index. This is the durable record other tooling and future
  sessions read back — not a one-liner in working memory.
- **A durable fact or reference** the project should keep → a **note** under
  `{memory_dir}/notes/<permalink>.md`, same schema.
- **A session needs to know X next time** (a flag, work-in-progress, a transient
  gotcha) → working memory (`agent/Core/memory.md`), per the rest of this
  protocol. Working memory is ephemeral — prune it; the durable graph is not.
- **Process narrative** (how a saga unfolded, a rebase recovery, session
  history) → `Logs/`, not memory.

> `{memory_dir}` defaults to `station/Memory` (set in the project manifest,
> `.bonsai/project.yaml`). The durable graph is untouched by working-memory
> cleanup — a resolved flag is pruned from `memory.md`, but the decision note
> it referenced stays.

---

## Reading Memory

At session start, read [agent/Core/memory.md](../Core/memory.md) and act on any flags:

- **Flags** — unresolved items from prior sessions. Address or escalate each one.
- **Work state** — what was in progress. Resume or confirm completion.
- **Notes** — context for ongoing work.

## Writing Memory

> **Brevity rule:** every memory write follows `Playbook/Standards/NoteStandards.md` — 3 lines max per entry, link out for detail. Work State = one-liner + plan/PR links. Notes = one line per durable gotcha. Phase walkthroughs go in the plan; commit walkthroughs in the PR; process narrative in `Logs/`.

Update `agent/Core/memory.md` (working memory) when:

- You encounter something the next session needs to know
- A task is partially complete and will continue later
- You're blocked and need to flag it

A **decision** that affects future work is not working memory — it goes to the
durable memory graph (`{memory_dir}/decisions/<permalink>.md`) per the routing
table above. Leave at most a one-line pointer in working memory if the next
session needs to act on it.

## Cleaning Memory

- Remove flags that have been resolved
- Clear work state when a task is complete
- Keep notes concise — if it's no longer relevant, remove it

## Rules

- **Do NOT use Claude Code's auto-memory system** (`~/.claude/projects/*/memory/`). All persistent memory goes in `agent/Core/memory.md` — version-controlled, auditable, inside the project.
- Memory is for cross-session continuity, not session logs
- Keep it short — if memory exceeds 30 lines, prune aggressively
- Never store secrets or credentials in memory
