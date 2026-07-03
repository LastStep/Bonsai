---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Backlog Hygiene"
date: 2026-07-03
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
- **Files Modified:** 4 — `station/Playbook/Backlog.md`, `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`, `station/Reports/Pending/2026-07-03-backlog-hygiene.md`
- **Tools Used:** Read, Edit, Write
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1 — Escalate misplaced P0s
Scanned the P0 section. Found **2 P0 items both resolved** and not yet cleaned up:

1. `[bug] Sensor hook commands use $PWD-walk-up` — Resolved by v0.4.3 hotfix (2026-05-13). Status.md "v0.4.3 hotfix shipped" confirms the fix baked absolute install-time paths. Commented out with resolution note.
2. `[feature] bonsai init / bonsai add need non-interactive flags` — Resolved by v0.4.2 (2026-05-13). Status.md "v0.4.2 release shipped" confirms `--non-interactive --from-config` flags with JSONL stdout and exit codes. Commented out with resolution note.

After cleanup, the P0 section contains **zero active items**.

### Step 2 — Cross-reference with Status.md
Read Status.md. In Progress: empty. Pending: only sentrux trial (already commented out in Backlog as promoted). Recently Done: Plans 40 + 41, v0.4.2, v0.4.3, PR triage.

Found **1 additional resolved item** in P1:
- `[feature] Full agent-drivable (non-interactive) CLI parity` (added 2026-06-13) — Plan 41 (2026-06-16) shipped headless `*Result` cores + JSONL/exit-code contract for all 4 mutating cmds. Commented out with resolution note.

Checked Pending items for "Blocked By" dependencies that a Backlog item could unblock: the sentrux trial is blocked by Rust toolchain install — no Backlog item addresses this; no change needed.

### Step 3 — Cross-reference with Roadmap.md
Phase 1 is fully complete (all checkboxes checked). Phase 2 milestones align correctly with P2/P3 backlog items (self-update mechanism, micro-task fast path). Phase 3 Big Bets (Managed Agents, Greenhouse) map to P3. No deprecated approach references or misaligned phase tags found.

No backlog item promotions required from Roadmap alignment.

### Step 4 — Flag stale items

**Critical time-sensitive item found:**
- `[ops] HOMEBREW_TAP_TOKEN PAT expiry calendar reminder` (P1) — The PAT was rotated 2026-04-22. Fine-grained PATs default to 90-day expiry (~2026-07-20). The item already called for a reminder at ~2026-07-15. **Today is 2026-07-03 — only 12 days until the recommended rotation date.** Updated the item with `[ACTION NEEDED — due ~2026-07-15, 12 days away]` tag. **Flagged for user review.**

**General staleness observation:**
Most P2/P3 items were added in April 2026 (60+ days ago) and remain at the same priority. However, most are either intentionally long-term (P3 research, Big Bets) or waiting on user prioritization decisions (Group D, Group E). No items appear to have rationale or context gaps that would warrant removal. The P2 backlog items added on 2026-06-13 and 2026-06-16 (during Plans 40/41) are recent and fully contextual.

No near-duplicates detected beyond those already commented out.

### Step 5 — Check for routine-generated items
Read RoutineLog.md. The last routine runs were **2026-05-07**, over 57 days ago. No routine runs between 2026-05-07 and today. The 2026-05-07 Routine Digest resolved all prior flags from the Backlog Hygiene that day.

**Important finding:** All 7 routines in the dashboard are significantly overdue (7-day routines last ran 57 days ago; 5-day routines same). The RoutineLog shows no entries between 2026-05-07 and now. The gap covers Plans 40 and 41, which added several P2 backlog items — these were correctly captured inline by the session agents; no orphaned findings identified.

No uncaptured routine findings to flag.

### Step 6 — Promote ready items via issue-to-implementation
No items are pre-approved for autonomous implementation. The HOMEBREW_TAP_TOKEN item is urgent but requires user action (rotating the PAT in GitHub Secrets) — not an issue-to-implementation candidate. Flagged for user review instead.

### Steps 7–8 — Log and dashboard update
Handled in the post-procedure steps below.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | High | P0 sensor hook bug was resolved by v0.4.3 but still active in Backlog | Backlog.md P0 section | Commented out with resolution note |
| 2 | High | P0 non-interactive flags feature was resolved by v0.4.2 but still active in Backlog | Backlog.md P0 section | Commented out with resolution note |
| 3 | High | P1 CLI parity feature was resolved by Plan 41 but still active in Backlog | Backlog.md P1 section | Commented out with resolution note |
| 4 | High | HOMEBREW_TAP_TOKEN PAT due for rotation in 12 days (~2026-07-15) | Backlog.md P1 item | Added urgency tag; flagged for user |
| 5 | Info | All 7 routines are overdue (57-day gap since last run on 2026-05-07) | routines.md dashboard | Flagged for user — not Backlog action |
| 6 | Info | P0 section is now empty after cleanup | Backlog.md | No action needed — expected outcome |

## Errors & Warnings
No errors encountered.

## Items Flagged for User Review

### [URGENT] HOMEBREW_TAP_TOKEN PAT rotation — due in ~12 days
The PAT was rotated 2026-04-22. 90-day expiry puts hard expiry at ~2026-07-20, with a recommended rotation before ~2026-07-15.

**Action required:** Rotate the `HOMEBREW_TAP_TOKEN` fine-grained PAT on GitHub (`LastStep/Bonsai` → Settings → Secrets). Set a new 90-day or 1-year PAT with the same `homebrew-tap` write scope. Update the expiry note in the Backlog item after rotation.

Symptom of missing this: next release will publish binaries + checksums but GoReleaser brew step fails with `401 Bad credentials` — formula stays on old version until manually patched.

### [INFO] All routines are overdue — 57-day gap
The last routine batch ran 2026-05-07. All 7 routines are now 50+ days overdue. This run covered Backlog Hygiene. The other 6 (Dependency Audit, Doc Freshness Check, Memory Consolidation, Roadmap Accuracy, Status Hygiene, Vulnerability Scan) are all due. A full routine-digest pass is recommended.

## Notes for Next Run
- P0 section is now clean. If a new P0 is added, it will be the only item and easy to spot.
- P1 HOMEBREW_TAP_TOKEN item should be removed or updated to reflect rotation date after the user acts on it.
- Several new P2 items were added during Plans 40/41 work (2026-06-13 to 2026-06-16) — these are recent and well-contextualized; no hygiene needed on them yet.
- Next run: check if the v0.5.0 tag was ever cut (Plan 40 note says "tag held by user"). If v0.5.0 shipped, several Backlog items referencing it may be stale.
- The `[ops] Routine bot PR pile-up` P1 item remains unresolved — the underlying fix (changing cloud routine behavior) was filed but not implemented. Worth noting for the next planning session.
