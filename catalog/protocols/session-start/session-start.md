---
tags: [protocol, session]
description: Ordered startup sequence — always steps + conditional loading by task type.
---

# Protocol: Session Start

> [!warning]
> This is a Protocol — follow it every session, no exceptions.

---

## Always (every session)

1. Read `agent/Core/identity.md` — confirm role and mindset
2. Read `agent/Core/memory.md` — surface pending flags, check work state
3. Read `agent/Core/self-awareness.md` — refresh context monitoring rules
4. Read `INDEX.md` — get project snapshot
5. Read `Playbook/Status.md` — see what's in progress and pending
6. Scan `Playbook/Backlog.md` — check for P0 items not yet in Status.md (escalate immediately if found)
7. Read `Logs/FieldNotes.md` — check for user updates since last session
8. Check `Reports/Pending/` — process any unreviewed agent reports
9. Read `agent/Protocols/security.md` — refresh security constraints (if installed)
10. Read `agent/Protocols/scope-boundaries.md` — refresh what you own (if installed)

> [!note]
> Paths like `INDEX.md`, `Playbook/`, `Logs/`, and `Reports/` refer to the project docs location configured during `bonsai init`. Check your workspace CLAUDE.md → External References for the exact paths.
> Backlog.md is a scan, not a full read — look for P0 items only. Full backlog review is handled by the backlog-hygiene routine.

---

## Conditional (by task type)

### If executing a plan

- Read the assigned plan in full before writing any code
- Read `Playbook/Standards/SecurityStandards.md`
- Read relevant coding standards from `agent/Skills/`

### If starting new work

- Check if a plan exists in `Playbook/Plans/Active/` — if not, ask the user
- Read scope boundaries before touching any files

### If reviewing or reporting

- Read the relevant plan or prior report
- Read `Playbook/Standards/SecurityStandards.md`
- Submit reports to `Reports/Pending/` using the report template
