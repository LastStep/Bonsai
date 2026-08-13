---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Backlog Hygiene"
date: 2026-08-13
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
- **Files Read:** 5 — `/home/user/Bonsai/station/Playbook/Backlog.md`, `/home/user/Bonsai/station/Playbook/Status.md`, `/home/user/Bonsai/station/Playbook/Roadmap.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`, `/home/user/Bonsai/station/agent/Core/routines.md`
- **Files Modified:** 2 — `/home/user/Bonsai/station/Playbook/Backlog.md` (3 items removed), `/home/user/Bonsai/station/agent/Core/routines.md` (dashboard updated), `/home/user/Bonsai/station/Logs/RoutineLog.md` (entry appended)
- **Tools Used:** Read, Edit, Write, Bash (ls, grep)
- **Errors Encountered:** 0

---

## Procedure Walkthrough

### Step 1: Escalate misplaced P0s

- **Action:** Read Backlog P0 section; for each P0 item, checked whether it appears in Status.md In Progress or Pending.
- **Result:** Found 2 active P0 items in Backlog. Both are RESOLVED — confirmed against Status.md Recently Done:
  - `[bug] Sensor hook commands use $PWD-walk-up` — Resolved by v0.4.3 hotfix (PRs #105/#106, 2026-05-13). Status.md: "sensor hook commands now bake install-time absolute paths."
  - `[feature] bonsai init / bonsai add need non-interactive flags` — Resolved by v0.4.2 (PR #102, 2026-05-13). Status.md: "--non-interactive + --from-config shipped."
  - A third item `[research] Trial sentrux on Bonsai repo` was already promoted to Status.md Pending (2026-05-07, blocked on Rust toolchain). Correctly reflected as an HTML comment.
- **Action taken:** Commented out both resolved P0 items with resolution notes.
- **Issues:** None — P0 section now has zero active items.

### Step 2: Cross-reference with Status.md

- **Action:** Read Status.md In Progress, Pending, and Recently Done tables.
- **Result:**
  - In Progress: empty.
  - Pending: Sentrux trial (blocked by Rust toolchain). No Backlog item can unblock it; requires user to install rustup.
  - Recently Done: confirmed v0.4.3 and v0.4.2 hotfixes; also confirmed Plan 41 shipped ALL headless CLI parity (init/add/update/remove) per "Plan 41 — Headless CLI Contract + MCP-ready cores SHIPPED — all 5 phases merged (PRs #120/#122/#123/#121/#125, main ab202c3)."
  - Found Backlog P1 item "Full agent-drivable (non-interactive) CLI parity" — RESOLVED by Plan 41 (2026-06-16). All four mutating commands now have headless cores + JSONL/exit contract.
- **Action taken:** Commented out P1 "Full CLI parity" item with resolution note; residual debt (unify remove cinematic/headless logic) is already captured as a P2 item.
- **Issues:** Plans 40 and 41 are still in `Plans/Active/` despite both being shipped per Status.md. This is bookkeeping for the Status Hygiene routine (currently 94 days overdue).

### Step 3: Cross-reference with Roadmap.md

- **Action:** Read Roadmap.md; compared Phase milestone items against current Backlog P2/P3 entries.
- **Result:**
  - Phase 1: All items checked. No action needed.
  - Phase 2 "Extensibility" milestones:
    - `[ ] Self-update mechanism` → tracked in P3 Backlog. ✓
    - `[ ] Template variables expansion` → **no backlog entry found**. Gap.
    - `[ ] Micro-task fast path` → tracked in P3 Backlog. ✓
  - Phase 3/4: Future, appropriately not tracked in active backlog.
  - No P2/P3 items are strong enough candidates for immediate P1 promotion without user input.
- **Action taken:** Flagging "Template variables expansion" gap for user review (low priority — Phase 2 is not the current focus).
- **Issues:** One minor roadmap tracking gap noted.

### Step 4: Flag stale items

- **Action:** Scanned all Backlog items for staleness (30+ days at same priority without progress, or no clear context).
- **Result:**
  - **URGENT FLAG — HOMEBREW_TAP_TOKEN PAT**: P1 item (`[ops] HOMEBREW_TAP_TOKEN PAT expiry calendar reminder`, added 2026-04-22) notes "set calendar reminder for ~2026-07-15." Today is 2026-08-13 — approximately 4 weeks past the rotation deadline. The PAT has very likely expired. Symptom on next release: GoReleaser fails brew step with 401. **Action required: rotate PAT now.**
  - **STALE SECURITY — Website npm vuln**: P2 item (`[security] Website npm vuln tree — astro upgrade breaks npm run build`, added 2026-06-16) is 58 days old. PR #108 that would fix it breaks the website build. Has received no attention since filing. Consider escalating to P1.
  - **STALE P1 items** (3+ months without progress): Routine bot PR pile-up (2026-05-07), Testing infrastructure for triggers/sensors (2026-04-16), Stale worktrees/branches (2026-04-20). All remain relevant but no new urgency.
  - **STALE Group B items**: All 5+ Group B items dated 2026-04-16 to 2026-04-24 (4 months old). No near-term resolution expected without user prioritization. Items are distinct and clear; no ambiguity.
  - No near-duplicates detected (the former CHANGELOG duplication was previously noted and partially resolved by the note about "refiled as good-first-issue").
- **Issues:** HOMEBREW_TAP_TOKEN PAT expiry is the most actionable urgent item.

### Step 5: Check for routine-generated items

- **Action:** Read RoutineLog.md entries since last backlog-hygiene run (2026-05-07).
- **Result:** No routine entries exist after 2026-05-07. All 7 routines are severely overdue (94–101 days since last run). The Plan 40 dispatch entry (2026-06-13) is a plan log, not a routine. No routine-generated findings since last run were captured or uncaptured — the maintenance gap itself is the finding.
- **Action taken:** Noting the routine gap as a flag for user review. No uncaptured backlog items from routine output.
- **Issues:** 3-month maintenance gap across all routines. The routine-check sensor should surface these at next session start.

### Step 6: Promote ready items via issue-to-implementation

- **Action:** Assessed whether any items are ready for autonomous promotion.
- **Result:** No items are approved for immediate implementation. The HOMEBREW_TAP_TOKEN PAT rotation requires direct user action (not agent-implementable). The website npm vuln fix requires user decision on build break strategy. All other flagged items require user prioritization before dispatch.
- **Action taken:** None — all promotion candidates flagged for user review in report.
- **Issues:** None.

### Step 7 & 8: Log results + Update dashboard

- **Action:** Appended RoutineLog entry; updated routines dashboard.
- **Result:** Both completed.
- **Issues:** None.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Resolved | P0 sensor $PWD-walk-up bug was already fixed by v0.4.3 | Backlog P0 | Commented out (resolved 2026-05-13) |
| 2 | Resolved | P0 non-interactive flags already fixed by v0.4.2 | Backlog P0 | Commented out (resolved 2026-05-13) |
| 3 | Resolved | P1 Full CLI parity already shipped by Plan 41 | Backlog P1 | Commented out (resolved 2026-06-16) |
| 4 | **Critical** | HOMEBREW_TAP_TOKEN PAT rotation deadline (~2026-07-15) has passed — PAT likely expired | Backlog P1 item | Flagged for user |
| 5 | Medium | Website npm vuln security item (P2) is 58 days stale; PR #108 breaks build | Backlog P2 | Flagged for user |
| 6 | Low | Roadmap Phase 2 "Template variables expansion" has no backlog tracking entry | Roadmap.md | Flagged for user |
| 7 | Info | Plans 40 and 41 still in Plans/Active/ despite both being shipped | Plans/Active/ | Flagged for Status Hygiene |
| 8 | Info | All 7 routines are 94–101 days overdue (last run 2026-05-04 to 2026-05-07) | routines.md | Noted — routine-check sensor should surface on next session start |

---

## Errors & Warnings

No errors encountered.

---

## Items Flagged for User Review

- **[URGENT] Rotate HOMEBREW_TAP_TOKEN PAT now.** The calendar reminder in Backlog P1 was set for ~2026-07-15. Today is 2026-08-13 — 4 weeks overdue. Without rotation, the next release will fail at the Homebrew formula update step (GoReleaser 401 Bad credentials). Action: rotate the fine-grained PAT on GitHub and update the `HOMEBREW_TAP_TOKEN` secret on `LastStep/Bonsai`.

- **[Security/P2] Website npm vuln (astro/esbuild/vite/js-yaml) has been stale for 58 days.** PR #108 was the fix attempt but breaks the website build after rebase. This needs a dedicated upgrade pass. Consider escalating to P1 or assigning to next available session.

- **[Low] Roadmap Phase 2 "Template variables expansion" has no corresponding backlog tracking entry.** If this feature is planned, add a tracking item. If it's been superseded or deprioritized, annotate the Roadmap.

- **[Bookkeeping] Run Status Hygiene and plan archiving.** Plans 40 and 41 are shipped but still in `Plans/Active/`. All routines are 94–101 days overdue. A routine-digest session would clear the backlog of pending reports and catch accumulated drift.

---

## Notes for Next Run

- P0 section is now empty of active items (all resolved). If sentrux trial completes and is dropped, remove its comment too.
- The 3-month routine gap means Doc Freshness, Dependency Audit, Vulnerability Scan, Memory Consolidation, Status Hygiene, and Roadmap Accuracy all have significant accumulated drift. Suggest running a full routine-digest session before next feature dispatch.
- HOMEBREW_TAP_TOKEN PAT likely expired — flag prominently if still unrotated at next run.
- The Group B (Code Quality & Testing) items have been stale for 4+ months; consider scheduling a dedicated debt-reduction session to clear the queue.
