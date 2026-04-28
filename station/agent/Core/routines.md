---
tags: [core, routines]
description: Periodic self-maintenance routines — schedules, dashboard, execution tracking.
---

# Routines

> [!note]
> Routines are periodic maintenance tasks the agent checks at session start. The session-start hook flags overdue routines — the user decides whether to run them now or defer. Routines are **opt-in per session**, never automatic.

---

## How Routines Work

1. **Session start:** Hook parses this file, compares `last_ran` against `frequency`, flags overdue routines.
2. **User decides:** Run now, defer, or skip. Agent never runs a routine without user approval.
3. **Execution:** Read the routine's definition file in `agent/Routines/`, follow the procedure step by step.
4. **Log:** Append results to `Logs/RoutineLog.md` (date, routine, outcome, notes).
5. **Update:** Set `last_ran` to today's date in this file.

### Rules

- Every routine must be **idempotent** — safe to re-run if interrupted mid-session.
- When validating facts against codebase, **mark stale entries as outdated** rather than deleting — preserves audit trail.
- Consolidation decisions follow four options: **keep** (still accurate), **update** (merge new info), **archive** (outdated but historically useful), **insert_new** (truly unique fact).

---

## Dashboard

<!-- ROUTINE_DASHBOARD_START — session-start hook parses this table -->

| Routine | Frequency | Last Ran | Next Due | Status |
|---------|-----------|----------|----------|--------|
| Backlog Hygiene | 7 days | 2026-04-28 | 2026-05-05 | done |
| Dependency Audit | 7 days | 2026-04-21 | 2026-04-28 | done |
| Doc Freshness Check | 7 days | 2026-04-21 | 2026-04-28 | done |
| Memory Consolidation | 5 days | 2026-04-25 | 2026-04-30 | done |
| Roadmap Accuracy | 14 days | 2026-04-14 | 2026-04-28 | done |
| Status Hygiene | 5 days | 2026-04-25 | 2026-04-30 | done |
| Vulnerability Scan | 7 days | 2026-04-21 | 2026-04-28 | done |

<!-- ROUTINE_DASHBOARD_END -->

---

## Routine Definitions

| Routine | File |
|---------|------|
| Backlog Hygiene | `agent/Routines/backlog-hygiene.md` |
| Dependency Audit | `agent/Routines/dependency-audit.md` |
| Doc Freshness Check | `agent/Routines/doc-freshness-check.md` |
| Memory Consolidation | `agent/Routines/memory-consolidation.md` |
| Roadmap Accuracy | `agent/Routines/roadmap-accuracy.md` |
| Status Hygiene | `agent/Routines/status-hygiene.md` |
| Vulnerability Scan | `agent/Routines/vulnerability-scan.md` |
