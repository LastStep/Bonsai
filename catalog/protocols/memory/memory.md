---
tags: [protocol, memory]
description: How to read, write, and clean working memory.
---

# Protocol: Working Memory

---

## Reading Memory

At session start, read [agent/Core/memory.md](../Core/memory.md) and act on any flags:

- **Flags** — unresolved items from prior sessions. Address or escalate each one.
- **Work state** — what was in progress. Resume or confirm completion.
- **Notes** — context for ongoing work.

## Writing Memory

> **Brevity rule:** every memory write follows `Playbook/Standards/NoteStandards.md` — 3 lines max per entry, link out for detail. Work State = one-liner + plan/PR links. Notes = one line per durable gotcha. Phase walkthroughs go in the plan; commit walkthroughs in the PR; process narrative in `Logs/`.

Update `agent/Core/memory.md` when:

- You encounter something the next session needs to know
- A task is partially complete and will continue later
- A decision was made that affects future work
- You're blocked and need to flag it

## Cleaning Memory

- Remove flags that have been resolved
- Clear work state when a task is complete
- Keep notes concise — if it's no longer relevant, remove it

## Rules

- **Do NOT use Claude Code's auto-memory system** (`~/.claude/projects/*/memory/`). All persistent memory goes in `agent/Core/memory.md` — version-controlled, auditable, inside the project.
- Memory is for cross-session continuity, not session logs
- Keep it short — if memory exceeds 30 lines, prune aggressively
- Never store secrets or credentials in memory
