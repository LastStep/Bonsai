---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Backlog Hygiene"
date: 2026-07-20
status: partial
---

# Routine Report — Backlog Hygiene

## Overview
- **Routine:** Backlog Hygiene
- **Frequency:** Every 7 days
- **Last Ran:** 2026-05-07
- **Triggered By:** loop.md autonomous dispatch

## Execution Metadata
- **Status:** partial
- **Duration:** ~10 min
- **Files Read:** 5 — `station/Playbook/Backlog.md`, `station/Playbook/Status.md`, `station/Playbook/Roadmap.md`, `station/Logs/RoutineLog.md`, `station/agent/Core/routines.md`
- **Files Modified:** 3 — `station/Playbook/Backlog.md`, `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`
- **Tools Used:** Read, Edit, Write
- **Errors Encountered:** 0

> **Status is "partial" because:** One item (HOMEBREW_TAP_TOKEN PAT expiry) requires immediate user action that cannot be completed autonomously. Promotion-via-issue-to-implementation was not triggered (no user approval in scope). All audit and cleanup steps completed successfully.

---

## Procedure Walkthrough

### Step 1: Escalate misplaced P0s

- **Action:** Read Backlog.md P0 section; cross-referenced each P0 against Status.md.
- **Result:** 2 P0 items found. Both were RESOLVED — already shipped and present in Status.md as Done. No untracked P0s requiring promotion.
  - `[bug] Sensor hook commands use $PWD-walk-up` — resolved by v0.4.3 hotfix (Status.md 2026-05-13)
  - `[feature] bonsai init / bonsai add need non-interactive flags` — resolved by v0.4.2 (Status.md 2026-05-13)
  - Note: `[research] Trial sentrux` was previously promoted to Status.md Pending (Blocked on Rust toolchain) — confirmed still present, not a new escalation.
- **Issues:** None — both P0s were resolved, not untracked.

### Step 2: Cross-reference with Status.md

- **Action:** Read Status.md In Progress and Recently Done tables; matched against all Backlog items.
- **Result:**
  - **Removed (commented out) 2 P0 items**: sensor hooks bug and non-interactive flags feature — both confirmed Done in Status.md.
  - **Removed (commented out) 1 P1 item**: "Full agent-drivable CLI parity: init/update/add/remove" — Plan 41 shipped 2026-06-16 (PRs #120–125) delivering exactly this: all 4 cmds with headless cores + JSONL/exit contract.
  - **No Status.md Pending items were found to be unblockable** by current Backlog items. The only Pending item (sentrux trial) is blocked on Rust toolchain, which has no matching Backlog resolution.
- **Issues:** None.

### Step 3: Cross-reference with Roadmap.md

- **Action:** Read Roadmap.md; compared Phase 2/3/4 milestones against Backlog P2/P3 items.
- **Result:**
  - Phase 1: Fully checked [x]. No stale Backlog references to Phase 1 items remain.
  - Phase 2 — Extensibility: 3 unchecked milestones. "Template variables expansion" has NO corresponding Backlog entry (gap). "Self-update mechanism" and "Micro-task fast path" both have P3 Backlog entries.
  - Phase 3 — Cloud & Orchestration: Both items (Managed Agents, Greenhouse companion app) have P3 Backlog entries under "Big Bets".
  - Phase 4 — Ecosystem: No Backlog entries for catalog marketplace, plugin system, or cross-project coordination.
  - No Backlog items reference deprecated approaches or completed phases (resolved items have been commented out).
- **Issues:** 1 gap — Roadmap Phase 2 "Template variables expansion" has no Backlog entry. Flagged for user.

### Step 4: Flag stale items

- **Action:** Reviewed all P0–P3 items for age (30+ days without progress), clarity, and near-duplicates.
- **Result:**

  **Critical staleness (urgent):**
  - `[ops] HOMEBREW_TAP_TOKEN PAT expiry` (P1, added 2026-04-22): The PAT was rotated 2026-04-22 with 90-day default expiry. **90 days from 2026-04-22 = 2026-07-21 — expiry is TOMORROW.** Next release will fail at the brew step with 401 if not rotated today. Immediate user action required.

  **Long-standing stale items (90+ days, no progress):**
  - `[debt] Testing infrastructure for triggers and sensors` (P1, added 2026-04-16, 95 days): No progress visible. Valid long-term item.
  - `[debt] Stale agent worktrees + branches accumulating` (P1, added 2026-04-20/21, ~91 days): Root accumulation issue persists. Not flagged as resolved anywhere.
  - `[debt] Break up generate.go` (Group B, P2, added 2026-04-16, 95 days): Now 1,357+ lines. No progress.
  - `[improvement] Plan archiving — Active/Archive folder structure` (Group E, added 2026-04-16, 95 days): Plans/Archive/ appears to exist (Status.md references it), but the scaffolding manifest item and workflow updates remain pending.
  - `[improvement] Plans Index file` (Group E, added 2026-04-21, 90 days): Still no Plans/INDEX.md.

  **Near-duplicates:** None new found beyond previously noted relationships (Plan archiving + Plans Index are intentionally related as a sub-task).

  **Items with no clear rationale:** All items have sufficient context.

- **Issues:** 1 critical — PAT expiry tomorrow. 5 long-stale P1/P2 items with no progress over 90 days.

### Step 5: Check for routine-generated items

- **Action:** Read RoutineLog.md entries since last backlog-hygiene run (2026-05-07) to find unfiled findings.
- **Result:** No routine executions occurred after 2026-05-07. Entries since that date are plan dispatch notes (Plan 40 on 2026-06-13, Plan 41 on 2026-06-16). All findings from those plans that warranted Backlog entries were already captured:
  - P2: `bonsai validate` ↔ `.bonsai.yaml` drift warning (added 2026-06-13) ✓
  - P2: Symlink hardening (added 2026-06-13) ✓
  - P2: Plan 40 review nits (added 2026-06-13) ✓
  - P2: `bonsai validate` can't pass on Bonsai repo (added 2026-06-13) ✓
  - P2: Website npm vuln tree (added 2026-06-16) ✓
  - P2: Unify remove business logic (added 2026-06-16) ✓
- **Issues:** None — all routine-adjacent findings already captured.

### Step 6: Promote ready items via issue-to-implementation

- **Action:** Assessed whether any item qualifies for autonomous promotion.
- **Result:** No user-approved items in scope. No P0s remaining (both resolved). The HOMEBREW_TAP_TOKEN item requires user action (PAT rotation), not code implementation. Deferred all promotion decisions to user.
- **Issues:** None.

### Step 7: Log results

- **Action:** Appended entry to `station/Logs/RoutineLog.md`.
- **Result:** Done.

### Step 8: Update dashboard

- **Action:** Updated `station/agent/Core/routines.md` dashboard row for Backlog Hygiene.
- **Result:** `Last Ran` → 2026-07-20, `Next Due` → 2026-07-27, `Status` → `done`.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | CRITICAL | HOMEBREW_TAP_TOKEN PAT expires TOMORROW (2026-07-21, 90 days after 2026-04-22 rotation). Next release will 401 at brew step. | `Backlog.md` P1 | Flagged for immediate user action — rotate PAT at github.com/settings/tokens before 2026-07-21 |
| 2 | Resolved (removed) | P0 sensor hooks bug — shipped in v0.4.3 (2026-05-13), already in Status.md Done | `Backlog.md` P0 | Commented out with resolution note |
| 3 | Resolved (removed) | P0 non-interactive flags feature — shipped in v0.4.2 (2026-05-13), already in Status.md Done | `Backlog.md` P0 | Commented out with resolution note |
| 4 | Resolved (removed) | P1 Full agent-drivable CLI parity — shipped in Plan 41 (2026-06-16), already in Status.md Done | `Backlog.md` P1 | Commented out with resolution note |
| 5 | Low | Roadmap Phase 2 "Template variables expansion" has no Backlog entry | `Roadmap.md` Phase 2 | Flagged for user — add Backlog entry or confirm not planned |
| 6 | Low | 5 P1/P2 items 90+ days old with no visible progress (testing infra, stale worktrees, generate.go split, plan archiving, plans index) | `Backlog.md` P1/P2 | Flagged for user review — re-prioritize, defer, or close |

---

## Errors & Warnings

No errors encountered.

---

## Items Flagged for User Review

- **[URGENT] Rotate HOMEBREW_TAP_TOKEN PAT TODAY (2026-07-20)** — The PAT was rotated 2026-04-22 with 90-day expiry. Expiry date is 2026-07-21 (tomorrow). Without rotation, the next release will fail at the Homebrew formula step with `401 Bad credentials`. Steps: go to github.com/settings/tokens → create new fine-grained PAT with Bonsai homebrew-tap write scope → `gh secret set HOMEBREW_TAP_TOKEN --repo LastStep/Bonsai` → update the Backlog P1 item with the new rotation date.

- **[GAP] Roadmap Phase 2 "Template variables expansion" has no Backlog entry** — Decide: add a Backlog P2/P3 entry tracking this milestone, or confirm it's intentionally out of scope and update the Roadmap.

- **[STALE] 5 long-standing items (90+ days) with no progress** — Consider: re-prioritize them to P2/P3, defer indefinitely with a note, or remove if no longer relevant:
  - P1: Testing infrastructure for triggers and sensors (95 days)
  - P1: Stale agent worktrees + branches accumulating (91 days)
  - P2/Group B: Break up generate.go (95 days)
  - P2/Group E: Plan archiving — Active/Archive folder structure (95 days)
  - P2/Group E: Plans Index file (90 days)

---

## Notes for Next Run

- P0 section is now clean (both items commented out with resolution notes). If new P0s are added, they should be verified against Status.md immediately.
- The HOMEBREW_TAP_TOKEN item in P1 should be updated with the new rotation date once rotated.
- No routine executions have occurred since 2026-05-07 (75 days). Multiple routines are significantly overdue (Dependency Audit, Doc Freshness Check, Vulnerability Scan last ran 2026-05-04; Memory Consolidation, Status Hygiene last ran 2026-05-07). The next routine digest should process all pending routine results.
- Plan 41 resolved the P1 CLI parity item cleanly — the P2 "Unify remove business logic" debt (cinematic path vs headless core) is now the outstanding follow-up from that plan.
