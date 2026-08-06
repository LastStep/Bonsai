---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Backlog Hygiene"
date: 2026-08-06
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
- **Duration:** ~5 min
- **Files Read:** 5 — `station/agent/Routines/backlog-hygiene.md`, `station/Playbook/Backlog.md`, `station/Playbook/Status.md`, `station/Playbook/Roadmap.md`, `station/Logs/RoutineLog.md`
- **Files Modified:** 3 — `station/Playbook/Backlog.md`, `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`
- **Tools Used:** Read, Edit, Write, Bash (ls)
- **Errors Encountered:** 0

---

## Procedure Walkthrough

### Step 1: Escalate misplaced P0s
- **Action:** Read Backlog.md P0 section; compared each P0 against Status.md In Progress and Pending tables.
- **Result:** Two P0 items found. Neither is in Status.md Pending or In Progress — but both are **resolved** (shipped):
  - `[bug] Sensor hook commands use $PWD-walk-up` — fixed in v0.4.3 (2026-05-13, Status.md Recently Done). No escalation needed; item is resolved.
  - `[feature] bonsai init / bonsai add need non-interactive flags` — shipped in v0.4.2 (2026-05-13, Status.md Recently Done). No escalation needed; item is resolved.
  - The `[research] Trial sentrux` item remains correctly commented-out in the P0 section (promoted to Status.md Pending, blocked on Rust toolchain).
- **Issues:** None — both active P0 items are resolved. Removed from backlog in Step 2.

### Step 2: Cross-reference with Status.md
- **Action:** Read Status.md; matched Backlog items against In Progress, Pending, and Recently Done tables.
- **Result:**
  - **Removed from P0:** `[bug] Sensor hook commands $PWD-walk-up` — resolved by v0.4.3 hotfix (Status.md Recently Done 2026-05-13). Replaced with audit-trail HTML comment.
  - **Removed from P0:** `[feature] bonsai init/add non-interactive flags` — resolved by v0.4.2 (Status.md Recently Done 2026-05-13). Replaced with audit-trail HTML comment.
  - **Removed from P1:** `[feature] Full agent-drivable (non-interactive) CLI parity: init / update / add / remove` — Plan 41 shipped this completely (all four commands now have headless cores + JSONL/exit contract, Status.md Recently Done 2026-06-16). Replaced with audit-trail HTML comment.
  - **Pending cross-check:** `[research] Trial sentrux` remains in Status.md Pending (blocked on Rust toolchain). No Backlog entry needed; it is already tracked.
  - **No Backlog items unblock Status.md Pending items** — the only Pending item (sentrux) is blocked on an external dependency (Rust toolchain), not a Backlog item.
- **Issues:** None.

### Step 3: Cross-reference with Roadmap.md
- **Action:** Read Roadmap.md; compared phase milestones against P2/P3 Backlog items.
- **Result:**
  - Phase 1 is fully complete (all checkboxes checked).
  - Phase 2 (Extensibility) milestones: `[Self-update mechanism]` and `[Micro-task fast path]` are in Backlog at P3 — appropriate priority for unstarted Phase 2 items.
  - No P2/P3 items ready to promote to P1 based on phase alignment — Phase 1 is done but no Phase 2 plan has been formally started.
  - No items reference deprecated approaches or completed phases after today's P0/P1 removals.
  - The `[debt] Unify remove business logic` P2 item (filed by Plan 41) correctly references a current Phase 2 concern.
- **Issues:** None.

### Step 4: Flag stale items
- **Action:** Scanned all priority tiers for items at the same priority for 30+ days without progress, items lacking context, and near-duplicates.
- **Result:**
  - **URGENT — HOMEBREW_TAP_TOKEN PAT expiry:** The P1 item `[ops] HOMEBREW_TAP_TOKEN PAT expiry calendar reminder` set a reminder for ~2026-07-15. Today is 2026-08-06 — the reminder date has **already passed by ~22 days**. The fine-grained PAT expires 90 days after rotation (rotated 2026-04-22, expiry ~2026-07-21). The PAT has very likely expired. Symptom: GoReleaser brew step fails with 401. Marked `[PAST DUE — rotate now]` in Backlog heading. Flagged for immediate user review.
  - **Long-standing stale items (acceptable):** Most P2/P3 items (Group B testing debt, Group C OSS readiness, Group D catalog expansion, Group E workspace improvements) are 90-120+ days old with no progress. These are intentionally deferred — no action required unless prioritized.
  - **No near-duplicates** identified that weren't already noted in previous runs.
  - **No items lacking context or rationale** — all have clear origin notes.
- **Issues:** One urgent flag (PAT expiry — see Items Flagged for User Review).

### Step 5: Check for routine-generated items
- **Action:** Read RoutineLog.md entries since last backlog-hygiene run (2026-05-07).
- **Result:** No routine runs are logged between 2026-05-07 and 2026-08-06. Log shows Plan 40 and Plan 41 session entries (2026-06-13 and 2026-06-16) but these are session logs, not routine runs. No routine-generated findings are waiting to be captured in the Backlog.
- **Issues:** All routines are significantly overdue (last ran: 2026-05-04 to 2026-05-07 — approximately 90 days ago). This is a scheduling gap, not a missed-capture issue. The gap itself is worth noting for the user.

### Step 6: Promote ready items via issue-to-implementation
- **Action:** Assessed whether any Backlog items are approved or ready for immediate implementation.
- **Result:** No items are explicitly approved for immediate implementation. The HOMEBREW_TAP_TOKEN rotation is urgent but is a user-action (rotating a GitHub secret) rather than an issue-to-implementation candidate. No workflow invoked.
- **Issues:** None.

### Step 7: Log results
- **Action:** Appended entry to `station/Logs/RoutineLog.md`.
- **Result:** Entry written.

### Step 8: Update dashboard
- **Action:** Updated `agent/Core/routines.md` Backlog Hygiene row.
- **Result:** Last Ran → 2026-08-06, Next Due → 2026-08-13, Status → done.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | resolved | P0 `[bug] Sensor hook $PWD-walk-up` was resolved by v0.4.3 but still in Backlog P0 | `Backlog.md` P0 | Removed; replaced with audit-trail comment |
| 2 | resolved | P0 `[feature] non-interactive flags` was resolved by v0.4.2 but still in Backlog P0 | `Backlog.md` P0 | Removed; replaced with audit-trail comment |
| 3 | resolved | P1 `[feature] Full headless CLI parity` was resolved by Plan 41 but still in Backlog P1 | `Backlog.md` P1 | Removed; replaced with audit-trail comment |
| 4 | **urgent** | `[ops] HOMEBREW_TAP_TOKEN PAT` reminder date 2026-07-15 has passed; PAT likely expired | `Backlog.md` P1 | Marked `[PAST DUE — rotate now]`; flagged for user |
| 5 | info | All routines ~90 days overdue — no runs logged since 2026-05-07 | `routines.md` dashboard | Noted; this run updates Backlog Hygiene only |

---

## Errors & Warnings

No errors encountered.

---

## Items Flagged for User Review

### 1. HOMEBREW_TAP_TOKEN PAT — LIKELY EXPIRED (urgent)

The `HOMEBREW_TAP_TOKEN` secret on `LastStep/Bonsai` was rotated 2026-04-22. Fine-grained PATs default to 90-day expiry, placing the expiry around **2026-07-21**. The backlog reminder was set for ~2026-07-15, but today is 2026-08-06 — approximately 22 days past the reminder and ~16 days past likely expiry.

**Symptom if expired:** GoReleaser brew step fails with `401 Bad credentials` from `api.github.com/repos/LastStep/homebrew-tap`. The binary release succeeds; only the Homebrew formula update is missed.

**Action needed:** Rotate the PAT at `github.com/settings/personal-access-tokens` and update `HOMEBREW_TAP_TOKEN` secret at `github.com/LastStep/Bonsai/settings/secrets/actions`. Then update the Backlog item's next reminder to the new expiry date (~90 days from rotation).

---

## Notes for Next Run

- The last routine batch ran 2026-05-04 to 2026-05-07 (~90 days ago). All routines are significantly overdue. Consider running the full routine batch (Dependency Audit, Vulnerability Scan, Doc Freshness Check, Status Hygiene, Memory Consolidation, Roadmap Accuracy) in a dedicated routine-digest session.
- After the PAT rotation, update the HOMEBREW_TAP_TOKEN backlog item with the new expiry date and reminder.
- P0 section is now empty after removing two resolved items. Next run should start with a clean slate.
- The `[research] Trial sentrux` item has been blocked on Rust toolchain since 2026-05-07 (90 days). If Rust is still not installed, consider whether to demote to P2 or close as won't-fix.
