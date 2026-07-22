---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Backlog Hygiene"
date: 2026-07-22
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
- **Files Read:** 6 — `station/Playbook/Backlog.md`, `station/Playbook/Status.md`, `station/Playbook/Roadmap.md`, `station/Logs/RoutineLog.md`, `station/agent/Core/routines.md`, `station/agent/Core/memory.md`
- **Files Modified:** 2 — `station/Playbook/Backlog.md`, `station/agent/Core/routines.md` (dashboard), `station/Logs/RoutineLog.md` (log entry)
- **Tools Used:** Read (files), Edit (Backlog.md modifications, dashboard update, log append), Write (this report)
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Escalate misplaced P0s
- **Action:** Read `Backlog.md` P0 section; identified 2 active P0 items (sensor hook bug, non-interactive flags).
- **Result:** Both P0 items are fully RESOLVED — shipped in prior releases. Neither is in Status.md "In Progress" or "Pending" because both are done. They belong in neither backlog nor Status.md at this point.
  - `[bug] Sensor hook commands use $PWD-walk-up` → shipped v0.4.3 (PRs #105/#106, 2026-05-13). Commented out.
  - `[feature] bonsai init/add non-interactive flags` → shipped v0.4.2 (Plan 39, PR #102, 2026-05-13). Commented out.
- **Issues:** None — both were cleanup items, no missing escalations.

### Step 2: Cross-reference with Status.md
- **Action:** Read Status.md (In Progress: none; Pending: 1 item — sentrux trial; Recently Done: Plans 40, 41, v0.4.x releases).
- **Result:**
  - Two P0 backlog items matched "Recently Done" work in Status.md — removed (commented out) per procedure.
  - P1 item "Full agent-drivable CLI parity" was resolved by Plan 41 (shipped 2026-06-16, all 4 cmds have `*Result` headless cores + JSONL/exit contract). Commented out from P1.
  - No Backlog items match "In Progress" (empty) or "Pending" (sentrux is already in a commented P0 stub).
  - No Status.md Pending "Blocked By" items unblocked by resolving a Backlog item (sentrux is blocked by Rust toolchain install, no Backlog item addresses that).
- **Issues:** None.

### Step 3: Cross-reference with Roadmap.md
- **Action:** Read Roadmap.md; checked all Phase items against backlog entries.
- **Result:**
  - Phase 1 is fully checked — all items shipped. No stale phase-1 references found in backlog.
  - Phase 2 "Template variables expansion" has NO corresponding Backlog entry at any priority. Flagged for user.
  - Phase 2 "Self-update mechanism" maps to a P3 backlog item ("Future Platform"). With Phase 1 complete, this could be promoted to P2.
  - Phase 2 "Micro-task fast path" maps to a P3 backlog item ("Future Platform"). Same promotion candidate.
  - Phase 3 and 4 items have P3 Backlog counterparts (Managed Agents, Greenhouse, Catalog marketplace) — appropriate priority.
  - No Backlog items reference deprecated approaches or completed phases beyond what was already cleaned up.
- **Issues:** Missing Phase 2 backlog entry for "Template variables expansion" — flagged for user.

### Step 4: Flag stale items
- **Action:** Assessed all Backlog items for age at same priority. Last hygiene run was 2026-05-07 (76 days ago). All items added in April–May 2026 qualify as 30+ day stale.
- **Result:**
  - **URGENT: [ops] HOMEBREW_TAP_TOKEN PAT expiry** (P1) — Added 2026-04-22 with a calendar reminder target of ~2026-07-15. We are now 2026-07-22, 7 days past that date. The PAT may already be expired. Flagged as urgent user action.
  - P2 items are all 60–90 days old without status change. Notable candidates for re-triage: `Harden scaffolding writes against symlink substitution`, `bonsai validate warn on project.yaml drift`, `Website npm vuln tree — astro upgrade`, `Unify remove business logic`.
  - P3 items are all 60–100 days old. Most are still valid research/future items. No clear context issues found.
  - Near-duplicates: The Group C "Changelog generation skill + release changelogs" and Group D "Changelog generation skill" items — the Group C item notes `(refiled as good-first-issue via Plan 24 Step E)`, which partially differentiates scope. Not identical but very close. Flagged for user consolidation decision.
- **Issues:** HOMEBREW_TAP_TOKEN PAT may be expired — urgent.

### Step 5: Check for routine-generated items
- **Action:** Read RoutineLog.md entries since last backlog-hygiene run (2026-05-07). Checked for unfiled findings.
- **Result:** No routine runs are logged between 2026-05-07 and 2026-07-22 (gap of 76 days). The only log entry after 2026-05-07 is the Plan 40 dispatch session (2026-06-13), which is a plan execution log, not a routine. No routine findings from this gap period need Backlog capture.
- **Issues:** 76-day routine gap is itself notable — all routines are significantly overdue. Not a Backlog item, but noted for awareness.

### Step 6: Promote ready items via issue-to-implementation
- **Action:** Assessed whether any approved or urgent items should be routed through issue-to-implementation.
- **Result:** No items have received explicit user approval for promotion. The HOMEBREW_TAP_TOKEN PAT item is the most urgent but requires user verification before action. No autonomous routing performed.
- **Issues:** None.

### Step 7 & 8: Log results and update dashboard
- **Action:** Appended to RoutineLog.md; updated routines.md dashboard row for Backlog Hygiene.
- **Result:** Dashboard row updated: Last Ran → 2026-07-22, Next Due → 2026-07-29, Status → done. Log entry appended.
- **Issues:** None.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | high | P0 "[bug] Sensor hook commands" resolved in v0.4.3 — stale in backlog | Backlog.md P0 | Commented out with resolution note |
| 2 | high | P0 "[feature] init/add non-interactive flags" resolved in v0.4.2 — stale in backlog | Backlog.md P0 | Commented out with resolution note |
| 3 | high | P1 "[feature] Full agent-drivable CLI parity" resolved by Plan 41 — stale in backlog | Backlog.md P1 | Commented out with resolution note |
| 4 | high | HOMEBREW_TAP_TOKEN PAT expiry reminder target was ~2026-07-15 — now 7 days past due | Backlog.md P1 | Flagged for user — verify PAT status and rotate if expired |
| 5 | medium | Roadmap Phase 2 "Template variables expansion" has no Backlog entry | Roadmap.md / Backlog.md | Flagged for user — add P2 item or confirm out of scope |
| 6 | medium | 76-day gap since last routine run — all routines significantly overdue | routines.md dashboard | Flagged for user awareness — no Backlog entry needed |
| 7 | low | P1 "Plan 41 in Plans/Active/ needs archiving" per memory.md Work State | Plans/Active/ | Flagged for user (per memory.md note; no autonomous archive) |
| 8 | low | Phase 2 items "Self-update mechanism" + "Micro-task fast path" are P3 despite Phase 1 completion | Backlog.md P3 | Flagged for user — consider promoting to P2 given Phase 2 roadmap is now the active phase |
| 9 | low | Near-duplicate: Group C + Group D both have "Changelog generation skill" items | Backlog.md | Flagged for user consolidation decision |

## Errors & Warnings
No errors encountered.

## Items Flagged for User Review

1. **[URGENT] HOMEBREW_TAP_TOKEN PAT expiry** — The P1 calendar reminder targeted ~2026-07-15; today is 2026-07-22. Verify PAT health immediately. A stale PAT will silently fail the Homebrew formula step on the next release (symptom: GoReleaser fails at brew step with 401; binaries still published). Rotate at `LastStep/Bonsai` → Settings → Secrets. Update `station/Playbook/Backlog.md` P1 item with rotation date after done.

2. **[MEDIUM] Roadmap Phase 2 "Template variables expansion" missing from Backlog** — No Backlog item exists for this Phase 2 milestone. Add as P2 if planning to pursue, or annotate Roadmap.md as deferred.

3. **[MEDIUM] Plans/Active/ contains Plan 40 and Plan 41** — Memory.md explicitly notes "Plan 41 file still in Plans/Active/ — archive to Plans/Archive/ at next wrap-up." Plan 40 (phases 1–3 shipped, Phase 4 held) also still in Active/. User should confirm archiving disposition for both.

4. **[LOW] Phase 2 Backlog promotion candidates** — With Phase 1 complete, "Self-update mechanism" and "Micro-task fast path" (both P3 / Future Platform) align with the now-active Phase 2 roadmap. Consider promoting to P2.

5. **[LOW] Changelog skill near-duplicate** — Group C "Changelog generation skill + release changelogs" and Group D "Changelog generation skill" are very similar. Group C notes "(refiled as good-first-issue via Plan 24 Step E)". User should decide whether to keep both or consolidate.

## Notes for Next Run

- P0 section is now empty (all three items resolved). This is healthy — P0 was a backlog smell.
- If HOMEBREW_TAP_TOKEN is rotated before next run, update/remove the P1 ops reminder item.
- 76-day gap in routine execution means all other routines (Dependency Audit, Vulnerability Scan, Doc Freshness Check, etc.) are also significantly overdue. Consider a routine-digest session after this run.
- "Trial sentrux on Bonsai repo" has been Pending in Status.md since 2026-05-07 (blocked by Rust toolchain). Still valid but aging.
- Plan 41 archive should be done before next backlog-hygiene to keep Plans/Active/ clean.
