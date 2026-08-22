---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Backlog Hygiene"
date: 2026-08-22
status: success
---

# Routine Report — Backlog Hygiene

## Overview
- **Routine:** Backlog Hygiene
- **Frequency:** Every 7 days
- **Last Ran:** 2026-05-07
- **Triggered By:** loop.md autonomous dispatch

## Execution Metadata
- **Status:** success
- **Duration:** ~8 min
- **Files Read:** 5 — `station/Playbook/Backlog.md`, `station/Playbook/Status.md`, `station/Playbook/Roadmap.md`, `station/Logs/RoutineLog.md`, `station/agent/Core/routines.md`
- **Files Modified:** 4 — `station/Playbook/Backlog.md` (2 resolved P0 items commented out), `station/Reports/Pending/2026-08-22-backlog-hygiene.md` (this report), `station/agent/Core/routines.md` (dashboard updated), `station/Logs/RoutineLog.md` (log entry appended)
- **Tools Used:** Read, Edit, Write
- **Errors Encountered:** 0

---

## Procedure Walkthrough

### Step 1 — Escalate Misplaced P0s
Scanned the P0 section. Found **2 P0 items** — both already resolved per Status.md Recently Done:
- `[bug] Sensor hook commands use $PWD-walk-up` — resolved via v0.4.3 (2026-05-13)
- `[feature] bonsai init / bonsai add need non-interactive flags` — resolved via v0.4.2 (2026-05-13)

Neither needed escalation — both are done. P0 section is now clean after removal.

### Step 2 — Cross-reference with Status.md
Read `Status.md`. In Progress: empty. Pending: 1 item (sentrux trial, blocked). Recently Done: 9 items.

Matches found — removed from Backlog:
1. P0 `[bug] Sensor hook commands` → v0.4.3 hotfix (Status Recently Done, 2026-05-13)
2. P0 `[feature] non-interactive flags` → v0.4.2 release (Status Recently Done, 2026-05-13)

Flagged for user (ambiguous — needs confirmation):
- P1 `[feature] Full agent-drivable (non-interactive) CLI parity` — Plan 41 shipped headless cores for all 4 mutating commands (init/add/update/remove) with JSONL/exit-code contract. This item's core requirement appears fully satisfied. User should confirm and remove if so.

Blocked-by check: The sentrux trial in Status.md Pending is blocked on Rust toolchain. No Backlog item directly unblocks it.

### Step 3 — Cross-reference with Roadmap.md
Read `Roadmap.md`. Phase 1 is **fully complete** (all checkboxes checked). Phase 2 is now the next active phase.

Phase 2 milestones (`[ ]` unchecked):
- `Self-update mechanism` — matches P3 Backlog item "Self-update mechanism" (Big Bets). Phase 1 done → this is now a relevant Phase 2 target.
- `Template variables expansion` — not in Backlog. No active tracking item.
- `Micro-task fast path` — matches P3 Backlog item "Micro-task fast path". Phase 1 done → Phase 2 target.

Flagged for user: Consider promoting "Self-update mechanism" and "Micro-task fast path" from P3 to P2 given Phase 1 is complete and Phase 2 is now the active horizon. Also consider adding `Template variables expansion` to Backlog (currently untracked).

No deprecated-approach items found — no Backlog items reference Phase 1 milestones as blockers.

### Step 4 — Flag Stale Items

**Priority stale (30+ days at same priority, no progress):**

| Item | Priority | Age | Concern |
|------|----------|-----|---------|
| [ops] HOMEBREW_TAP_TOKEN PAT expiry | P1 | 122 days | **Calendar reminder was 2026-07-15 — already 38 days past. PAT may be expired.** |
| [ops] Routine bot PR pile-up | P1 | 107 days | Process improvement, no resolution noted. Still relevant if cloud routines run. |
| [debt] Testing infrastructure for triggers and sensors | P1 | 128 days | No progress visible. High-value item; consider scheduling. |
| [debt] Stale agent worktrees + branches | P1 | 124 days | Recurring cleanup mentioned in RoutineLog but no formal resolution. |
| [bookkeeping] Retroactively trim Backlog entries to NoteStandards | Group A | 119 days | No progress. Low effort, backlog hygiene meta-task. |

**Items with unclear context or no clear next action:** None — all items have rationale and source.

**Near-duplicates:** The "Changelog generation skill" (Group D) and "CHANGELOG.md + richer release notes" in Group C overlap. This was previously noted (2026-04-21 RoutineLog). Still unresolved. Flag for user decision.

### Step 5 — Check Routine-Generated Items Since 2026-05-07

Reviewed RoutineLog.md entries after 2026-05-07:
- 2026-06-13: Plan 40 dispatch filed 7 Backlog items (P2 security hardening, identity drift, review nits, validate lock issue, website npm vulns, unify remove logic, plan-grilling) — **all present in Backlog ✓**
- 2026-06-16 (from Plan 41 Status.md row): Filed "Unify remove business logic" (P2) and "Website npm vuln tree" (P2) — **both present in Backlog ✓**

No uncaptured findings from recent routine entries.

### Step 6 — Promote Ready Items via Issue-to-Implementation
Skipped — no user present for confirmation. No autonomous promotions made.

### Steps 7–8 — Log + Dashboard
Log entry appended and dashboard updated (see below).

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | high | P0 `[bug] Sensor hook commands` already resolved via v0.4.3 | Backlog P0 | Removed — replaced with resolution comment |
| 2 | high | P0 `[feature] non-interactive flags` already resolved via v0.4.2 | Backlog P0 | Removed — replaced with resolution comment |
| 3 | high | HOMEBREW_TAP_TOKEN PAT: reminder date 2026-07-15 passed 38 days ago | Backlog P1 | Flagged for user — rotate immediately |
| 4 | medium | P1 "Full agent-drivable CLI parity" likely resolved by Plan 41 | Backlog P1 | Flagged for user — confirm and remove |
| 5 | medium | P1 "[debt] Testing infrastructure" — 128 days old, no progress | Backlog P1 | Flagged for re-prioritization |
| 6 | medium | P1 "[debt] Stale agent worktrees" — 124 days old, no formal resolution | Backlog P1 | Flagged for re-prioritization |
| 7 | medium | P1 "[ops] Routine bot PR pile-up" — 107 days old | Backlog P1 | Flagged for re-prioritization |
| 8 | low | Phase 2 now active; P3 items "Self-update" + "Micro-task fast path" are Phase 2 milestones | Backlog P3, Roadmap | Flagged for promotion to P2 |
| 9 | low | "Template variables expansion" (Phase 2 milestone) has no Backlog tracking item | Roadmap Phase 2 | Flagged — consider adding to Backlog |
| 10 | low | Changelog near-duplicate (Group C vs Group D) persists unresolved | Backlog | Flagged for user decision |

---

## Errors & Warnings

No errors encountered.

---

## Items Flagged for User Review

- **[URGENT] HOMEBREW_TAP_TOKEN PAT expiry** — PAT was rotated 2026-04-22, reminder set for 2026-07-15. Today is 2026-08-22 — 38 days past the reminder. The PAT has very likely expired. Next GoReleaser run will fail at the brew step with `401 Bad credentials`. Rotate the PAT immediately at GitHub → Settings → Developer settings → Fine-grained personal access tokens, then update the `HOMEBREW_TAP_TOKEN` secret on `LastStep/Bonsai`. Update the Backlog P1 item with the new rotation date and set a new calendar reminder for ~90 days out.

- **P1 "Full agent-drivable CLI parity" — confirm resolution** — Plan 41 shipped headless cores for init/add/update/remove with `*Result` pure cores, JSONL/exit-code contract, `list --json`, and `docs/agent-interface.md`. This satisfies the P1 item's stated need. If confirmed, remove the Backlog item. Note: the related debt item "Unify remove business logic" (P2) is a separate follow-on and should stay.

- **P1 stale items — re-prioritization needed:** "[debt] Testing infrastructure for triggers and sensors" (128d), "[debt] Stale agent worktrees + branches" (124d), "[ops] Routine bot PR pile-up" (107d). Recommend either scheduling these or downgrading to P2 if not blocking current work.

- **Phase 2 milestones — promotion candidates:** "Self-update mechanism" and "Micro-task fast path" are both Phase 2 Roadmap items currently sitting at P3. Phase 1 is fully complete. Consider promoting both to P2. Also: "Template variables expansion" (Phase 2 milestone) has no Backlog tracking item — consider adding one.

- **Changelog near-duplicate decision:** "Changelog generation skill + release changelogs" (Group D) and any changelog-related Group C item. User noted this in 2026-04-21 RoutineLog but no decision was made. Pick one location or merge.

---

## Notes for Next Run

- P0 section is now clean — both items resolved and commented out. If the sentrux trial completes or is abandoned, remove/update the Status.md Pending row and the HTML comment in the P0 section.
- Verify HOMEBREW_TAP_TOKEN rotation was completed before next backlog-hygiene run.
- If Plan 41 P1 item confirmed resolved by user, next run can verify the Backlog no longer has it.
- The NoteStandards sweep (Group A bookkeeping) is 119 days old — low-effort, could be done as a quick fix in a regular session without a formal plan.
