---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Backlog Hygiene"
date: 2026-08-04
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
- **Files Read:** 5 — `/home/user/Bonsai/station/agent/Routines/backlog-hygiene.md`, `/home/user/Bonsai/station/Playbook/Backlog.md`, `/home/user/Bonsai/station/Playbook/Status.md`, `/home/user/Bonsai/station/Playbook/Roadmap.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`
- **Files Modified:** 3 — `/home/user/Bonsai/station/Playbook/Backlog.md`, `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`
- **Tools Used:** Read (files), Edit (Backlog.md, routines.md, RoutineLog.md), Write (this report), Bash (directory listing)
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Escalate misplaced P0s
- **Action:** Read Backlog.md P0 section; checked each P0 against Status.md In Progress and Pending tables.
- **Result:** Both P0 items were NOT in Status.md as In Progress or Pending — they are in Recently Done (resolved). No true P0 escalation needed; both items were already resolved and pending removal in Step 2.
  - P0 `[bug] Sensor hook commands use $PWD-walk-up` → resolved in v0.4.3, shipped 2026-05-13 (Status.md Recently Done).
  - P0 `[feature] bonsai init / bonsai add need non-interactive flags` → resolved in v0.4.2, shipped 2026-05-13 (Status.md Recently Done).
- **Issues:** None — resolved items were handled in Step 2.

### Step 2: Cross-reference with Status.md
- **Action:** Read Status.md; compared all Backlog entries against In Progress, Pending, and Recently Done tables.
- **Result:** Found 3 Backlog items resolved and present in Status.md Recently Done:
  1. P0 `[bug] Sensor hook commands use $PWD-walk-up` — resolved v0.4.3, PRs #105/#106. **Removed from Backlog** (replaced with HTML comment).
  2. P0 `[feature] bonsai init / bonsai add need non-interactive flags` — resolved v0.4.2 (--non-interactive + --from-config). **Removed from Backlog** (replaced with HTML comment).
  3. P1 `[feature] Full agent-drivable CLI parity: init / update / add / remove` — resolved via Plan 41 (2026-06-16, PRs #120/#122/#123/#121/#125; all 4 mutating cmds have headless *Result cores + JSONL/exit contract). **Removed from Backlog** (replaced with HTML comment).
- **Blocked items check:** Status.md Pending has one item: "Trial sentrux on Bonsai repo" (blocked on Rust toolchain). No Backlog item exists specifically to unblock it (Rust toolchain install is an operational step, not a code item).
- **Issues:** None.

### Step 3: Cross-reference with Roadmap.md
- **Action:** Read Roadmap.md; checked P2/P3 Backlog items against current-phase milestones; flagged deprecated-phase references.
- **Result:**
  - **Phase 1 is fully complete** — all items are `[x]`. This cycle's sweeps plus prior routine-digests (2026-05-07) confirmed the last two unchecked items were ticked in May. Phase 1 should be formally annotated as done.
  - **Phase 2 milestone gap:** Roadmap Phase 2 lists "Template variables expansion" as a goal — no Backlog entry exists for this item. Flagged for user.
  - **P3 items aligned with Phase 2:** `[improvement] Self-update mechanism` and `[improvement] Micro-task fast path` are Phase 2 roadmap items sitting at P3. Given Phase 1 is now complete, consider promoting to P2.
  - **No deprecated-phase references found** in existing Backlog items.
- **Issues:** Missing Backlog tracking entry for "Template variables expansion" (Roadmap Phase 2 goal).

### Step 4: Flag stale items
- **Action:** Checked all items for 30+ days at same priority without progress; checked for unclear context; scanned for near-duplicates.
- **Result:**
  - **Critical stale flag:** P1 `[ops] HOMEBREW_TAP_TOKEN PAT expiry calendar reminder` — added 2026-04-22, reminder set for **2026-07-15** to rotate before next release. Today is **2026-08-04**, 20 days past that reminder date. If not already rotated, the PAT has very likely expired (fine-grained PATs default to 90-day expiry; 90 days from 2026-04-22 = 2026-07-21, which has passed). Symptom: GoReleaser brew step fails with `401 Bad credentials`. **Flagged for immediate user action.**
  - **Stale P1 items (100+ days, no progress):**
    - `[ops] Routine bot PR pile-up` (added 2026-05-07, 89 days) — still relevant; no resolution path in sight.
    - `[debt] Testing infrastructure for triggers and sensors` (added 2026-04-16, 110 days) — still relevant, Group B ordering constraint applies.
    - `[debt] Stale agent worktrees + branches accumulating` (added 2026-04-20, 106 days) — still relevant; no automated cleanup yet.
  - **Stale P2 items (30–110 days, no progress):**
    - Group B items: Plan-29-test-gap, Plan-29-security-hardening, Plan-31-cosmetic, Plan-31-security-hardening — all 100+ days old, small nits from past reviews. Low urgency but growing in age.
    - `[security] Harden all scaffolding writes against symlink substitution` (52 days).
    - `[bug] bonsai validate can't pass on Bonsai repo itself` (52 days) — noted as "Blocks proper Plan 40 dogfood"; Plan 40 shipped without dogfood, this remains open.
    - `[security] Website npm vuln tree — astro upgrade breaks npm run build` (49 days) — active security debt; the vulnerability-scan routine should pick this up.
    - `[debt] Unify remove business logic` (49 days).
  - **Near-duplicates:** None found. Group D changelog item was already noted as "refiled as good-first-issue via Plan 24 Step E." Group E plan-archiving and Plans-Index items are related but distinct.
  - **Unclear context:** No items found with entirely missing rationale.
- **Issues:** PAT expiry overdue (high urgency).

### Step 5: Check for routine-generated items
- **Action:** Read RoutineLog.md entries since 2026-05-07 (last backlog-hygiene run) for routine findings that should be in Backlog.
- **Result:** Only one RoutineLog entry after 2026-05-07: "2026-06-13 — Plan 40 dispatch" (not a routine, a plan execution). No routines ran in the 89-day gap since 2026-05-07. All routines are severely overdue:
  - Backlog Hygiene: 89 days overdue
  - Dependency Audit: 92 days overdue
  - Doc Freshness Check: 92 days overdue
  - Memory Consolidation: 89 days overdue
  - Roadmap Accuracy: 89 days overdue
  - Status Hygiene: 89 days overdue
  - Vulnerability Scan: 92 days overdue
- No uncaptured findings from recent routines because no routines ran. The routine dispatch system (loop.md or equivalent cron) appears to have been inactive since May 2026. **Flagged for user — routine dispatch may need to be restarted.**
- **Issues:** Routine execution gap is a systemic concern; other routines (vulnerability scan, dependency audit) likely have new findings that have gone uncaptured.

### Step 6: Promote ready items via issue-to-implementation
- **Action:** Reviewed whether any item is approved for implementation or requires immediate routing.
- **Result:** Running autonomously — no user is in the loop to confirm promotion. PAT expiry item (P1 ops) is the most urgent but requires user action (rotating a PAT), not code implementation. No items routed through issue-to-implementation workflow.
- **Issues:** None.

### Step 7 & 8: Log results and update dashboard
- **Action:** Appended to RoutineLog.md and updated routines.md dashboard.
- **Result:** Done — see below in report.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | HIGH | HOMEBREW_TAP_TOKEN PAT expiry reminder date (2026-07-15) has passed — PAT likely expired | Backlog P1 | Flagged for user; item kept in Backlog |
| 2 | HIGH | Routine execution 89–92 day gap — all 7 routines severely overdue; dispatch system may be inactive | routines.md dashboard | Flagged for user |
| 3 | MEDIUM | P0 `[bug] Sensor hook $PWD-walk-up` was resolved in v0.4.3 but still in Backlog | Backlog P0 | Removed (HTML comment with audit trail) |
| 4 | MEDIUM | P0 `[feature] non-interactive flags` was resolved in v0.4.2 but still in Backlog | Backlog P0 | Removed (HTML comment with audit trail) |
| 5 | MEDIUM | P1 `[feature] Full agent-drivable CLI parity` was resolved by Plan 41 (2026-06-16) but still in Backlog | Backlog P1 | Removed (HTML comment with audit trail) |
| 6 | MEDIUM | Roadmap Phase 2 goal "Template variables expansion" has no Backlog tracking entry | Roadmap.md / Backlog.md | Flagged for user |
| 7 | LOW | Roadmap Phase 1 fully complete — no formal annotation; Phase 2 now the current phase | Roadmap.md | Flagged for user |
| 8 | LOW | P3 items "Self-update mechanism" + "Micro-task fast path" are Phase 2 roadmap goals sitting at lowest priority | Backlog P3 | Flagged for user — consider promoting to P2 |
| 9 | LOW | P1 ops + debt items 89–110 days stale with no progress — PAT, bot PR pile-up, testing infra, worktree debt | Backlog P1 | Noted; kept in place (still valid, not resolvable autonomously) |
| 10 | LOW | P2 security items (website npm vuln tree, scaffolding symlink hardening) 49–52 days without progress | Backlog P2 | Noted; vulnerability-scan routine should pick up the npm vuln when run |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

- **[URGENT] HOMEBREW_TAP_TOKEN PAT expiry** — The calendar reminder set for 2026-07-15 has passed. Fine-grained PATs have a 90-day default expiry; if the PAT rotated 2026-04-22 has not been refreshed, it expired ~2026-07-21. Next `goreleaser` run will fail at the Homebrew tap step with `401 Bad credentials`. Action: verify PAT status in GitHub settings, rotate if expired, update `HOMEBREW_TAP_TOKEN` repo secret.

- **[URGENT] Routine dispatch gap** — All 7 routines are 89–92 days overdue. The loop.md/cron dispatch appears to have been inactive since May 2026. This run is the first since then. Consider restarting the dispatch mechanism and running the remaining 6 overdue routines (vulnerability scan, dependency audit, and doc freshness check are highest priority given time elapsed).

- **[MEDIUM] Roadmap Phase 2 tracking gap** — "Template variables expansion" is a listed Phase 2 goal in Roadmap.md but has no corresponding Backlog entry. Decide whether to add a P2 entry or scope it into an existing plan.

- **[LOW] Phase 1 complete, Phase 2 current** — All Roadmap Phase 1 items are `[x]`. Consider annotating Roadmap.md with a completion date and formally noting Phase 2 as the active phase.

- **[LOW] P3 Phase-2 goal items** — `[improvement] Self-update mechanism` and `[improvement] Micro-task fast path` are listed as Roadmap Phase 2 goals but sit at P3 in Backlog. Given Phase 1 is done, consider promoting to P2.

## Notes for Next Run

- 3 items removed from Backlog this cycle (both P0s + one P1): all were resolved via v0.4.2/v0.4.3/Plan 41 during the 89-day gap; the gap caused them to linger.
- If routine dispatch is restarted, the next backlog-hygiene run should expect findings from vulnerability-scan, dependency-audit, and doc-freshness-check routines (all 90+ days overdue) to need Backlog captures.
- The PAT expiry item can be removed from Backlog once the user confirms rotation.
- After vulnerability-scan runs, check whether the website npm vuln tree (P2) has been resolved or worsened.
