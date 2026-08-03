---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Backlog Hygiene"
date: 2026-08-03
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
- **Files Read:** 6 — `Playbook/Backlog.md`, `Playbook/Status.md`, `Playbook/Roadmap.md`, `Logs/RoutineLog.md`, `agent/Core/routines.md`, `Reports/Archive/2026-05-07-backlog-hygiene.md`
- **Files Modified:** 4 — `Playbook/Backlog.md` (3 items commented out), `agent/Core/routines.md` (dashboard row), `Logs/RoutineLog.md` (append entry), this report file
- **Tools Used:** Read, Write, Edit
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Escalate misplaced P0s
- **Action:** Read P0 section of `Backlog.md`, cross-checked each item against `Status.md` In Progress and Pending tables.
- **Result:** 2 active P0 bullet items found. Both are already resolved — they appear in Status.md Recently Done, not In Progress or Pending. No escalation needed; cleanup handled in Step 2.
  - `[bug] Sensor hook commands use $PWD-walk-up` — shipped as v0.4.3 (2026-05-13, PRs #105/#106)
  - `[feature] bonsai init / bonsai add need non-interactive flags` — shipped as v0.4.2 (2026-05-13, PR #102)
- **Issues:** No active P0 remains in Backlog after cleanup. P0 section is now empty (both entries commented out). The previously-flagged `[research] Trial sentrux on Bonsai repo` remains correctly placed in Status.md Pending (Blocked By: Rust toolchain).

### Step 2: Cross-reference with Status.md
- **Action:** Walked Recently Done table against all active Backlog entries; checked Pending for unblocking opportunities.
- **Result:**
  - **3 resolved items identified for removal:** (1) P0 sensor hook bug → v0.4.3 shipped; (2) P0 non-interactive flags → v0.4.2 shipped; (3) P1 "Full agent-drivable CLI parity" → Plan 41 shipped all 5 phases (PRs #120–#125, 2026-06-16), all 4 mutating commands now have headless `*Result` cores + JSONL/exit contract. All 3 commented out from Backlog.
  - **Pending unblocking:** Status.md Pending has one item (sentrux trial, blocked on Rust toolchain). No Backlog items directly unblock this — it requires user action to install rustup.
- **Issues:** None beyond the 3 removals applied.

### Step 3: Cross-reference with Roadmap.md
- **Action:** Read `Roadmap.md`, compared Phase 1–4 milestones against Backlog priority tiers. Scanned for deprecated approaches or completed phases with stale Backlog references.
- **Result:**
  - Phase 1: All items checked. Clean.
  - Phase 2 (Extensibility): 3 unchecked items. All have P3 Backlog representations (Self-update mechanism, Micro-task fast path — both Big Bets/P3). Template variables expansion has no Backlog entry (acceptable — it's a P3-level idea not yet surfaced).
  - Phase 3 (Cloud & Orchestration): Both items (Managed Agents integration, Greenhouse companion app) have P3 Big Bets entries.
  - Phase 4 (Ecosystem): No items in Backlog yet (far-future).
  - No P2/P3 items strong enough to warrant promotion to P1 solely on Roadmap alignment — current P1 items (HOMEBREW_TAP_TOKEN expiry, bot PR pile-up, testing infra, worktree cleanup) take priority.
  - No deprecated approaches or stale phase references found.
- **Issues:** None.

### Step 4: Flag stale items
- **Action:** Computed item ages relative to 2026-08-03. Cutoff for "30+ days at same priority" = added before 2026-07-04.
- **Result:** The last backlog-hygiene was 2026-05-07 (~87 days ago). Nearly all P1 and P2 items are now 30+ days at their current priority. Key staleness flags:

  **P1 items (all stale):**
  - `[ops] HOMEBREW_TAP_TOKEN PAT expiry` (added 2026-04-22): **CRITICAL — PAST DUE.** Reminder was set for ~2026-07-15. Today is 2026-08-03 — the reminder date has passed by 19 days. If the PAT hasn't been rotated, the next release will fail at the Homebrew tap step with a 401. Immediate user action required.
  - `[ops] Routine bot PR pile-up` (added 2026-05-07): 87 days. The 9 stale PRs were closed in May 2026, but the underlying fix (change cloud routine push behavior) has not shipped. Still valid.
  - `[debt] Testing infrastructure for triggers and sensors` (added 2026-04-16): 109 days. No movement. Consider a dedicated plan to knock out the unit test layer at minimum.
  - `[debt] Stale agent worktrees + branches accumulating` (added 2026-04-20): 105 days. The worktree leak pattern continued through Plans 40/41. Likely more accumulation since the 2026-04-21 audit.

  **P2 items (newly stale or urgency-elevated):**
  - `[security] Website npm vuln tree` (added 2026-06-16): 48 days. Has real security CVEs — esbuild HIGH, vite HIGH+MED, js-yaml MED open in `/website`. Astro upgrade PR #108 fails build after rebase. This is the oldest un-addressed security item after P0 cleanup.
  - `[debt] Unify remove business logic` (added 2026-06-16): 48 days. Plan 41 follow-up. Moderate drift risk.

  **Near-duplicates identified:**
  - `[Plan-29-security-hardening] Phase H validator hardening` (Group B): The one remaining item (Unicode lookalikes — "Purely speculative") is labeled as speculative by its author. With 3+ months of no action, consider demoting to P3 or removing entirely.
  - Group F `[docs]` items (`AltScreen behavior change in release notes`, `Fill "Deviations from Plan"`) were added 2026-04-20 (~105 days ago). Plan 15 shipped long ago; these release notes are now stale opportunities. The "Deviations" item is a meta-process note — may be better served in the planning-template skill directly.

- **Issues:** HOMEBREW_TAP_TOKEN PAT expiry is now past its reminder date — flagging as high urgency.

### Step 5: Check for routine-generated items
- **Action:** Walked `RoutineLog.md` for entries since last backlog-hygiene (2026-05-07).
- **Result:** Only one entry post-2026-05-07 exists: `2026-06-13 — Plan 40 dispatch`. It noted two deferred items ("dogfood deferred," "validate can't pass on Bonsai repo") — both are captured in Backlog as P2 items (`[bug] bonsai validate can't pass on the Bonsai repo itself` and `[improvement] bonsai validate warn on .bonsai/project.yaml drift`). Confirmed.

  **CRITICAL FINDING:** No routine executions are logged between 2026-05-07 and today (2026-08-03) — a gap of approximately 87 days. All 7 routines are significantly overdue:
  - Dependency Audit: overdue since 2026-05-11 (84 days late)
  - Doc Freshness Check: overdue since 2026-05-11 (84 days late)
  - Memory Consolidation: overdue since 2026-05-12 (83 days late)
  - Status Hygiene: overdue since 2026-05-12 (83 days late)
  - Vulnerability Scan: overdue since 2026-05-11 (84 days late)
  - Roadmap Accuracy: overdue since 2026-05-21 (74 days late)
  - Backlog Hygiene (this run): overdue since 2026-05-14 (81 days late)

  No uncaptured routine findings to add to Backlog (per procedure, don't auto-add).

- **Issues:** Routine system has been inactive for ~3 months. Flag for user to schedule a routine-digest pass to process the overdue cycle.

### Step 6: Promote ready items via issue-to-implementation
- **Action:** Reviewed remaining P0/P1 items for promotion candidacy.
- **Result:** P0 section is now empty. Top P1 item is HOMEBREW_TAP_TOKEN PAT expiry — this requires user action (rotate PAT on GitHub), not an implementation workflow. No items are approved for autonomous promotion. Deferring to user per procedure.
- **Issues:** None.

### Step 7: Log results
- **Action:** Appended structured entry to `Logs/RoutineLog.md`.
- **Result:** Done.
- **Issues:** None.

### Step 8: Update dashboard
- **Action:** Updated `agent/Core/routines.md` Backlog Hygiene row — `Last Ran` 2026-05-07 → 2026-08-03, `Next Due` 2026-05-14 → 2026-08-10, `Status` done.
- **Result:** Dashboard updated.
- **Issues:** None.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | **Resolved (cleanup)** | P0 `[bug] Sensor hook commands use $PWD-walk-up` was in Backlog but shipped in v0.4.3 (2026-05-13). | `Backlog.md` P0 | Commented out with resolution note. |
| 2 | **Resolved (cleanup)** | P0 `[feature] bonsai init/add non-interactive flags` was in Backlog but shipped in v0.4.2 (2026-05-13). | `Backlog.md` P0 | Commented out with resolution note. |
| 3 | **Resolved (cleanup)** | P1 `[feature] Full agent-drivable CLI parity` was in Backlog but shipped via Plan 41 (2026-06-16). | `Backlog.md` P1 | Commented out with resolution note. |
| 4 | **HIGH — PAST DUE** | `[ops] HOMEBREW_TAP_TOKEN PAT expiry` — reminder was 2026-07-15, now 19 days overdue. Next release will fail Homebrew step with 401 if PAT not rotated. | `Backlog.md` P1 | Flagged for immediate user action. |
| 5 | **HIGH — routine gap** | No routine runs logged since 2026-05-07 (~87 days). All 7 routines overdue. Dependency Audit, Vulnerability Scan, Doc Freshness, Memory Consolidation, Status Hygiene most urgent. | `Logs/RoutineLog.md` | Flagged to user; recommend scheduling routine-digest pass. |
| 6 | **Medium** | `[security] Website npm vuln tree — astro upgrade breaks build` (P2, 48 days). Real CVEs open in `/website`. | `Backlog.md` P2 | Flagged to user; remains P2, consider promoting to P1. |
| 7 | **Medium** | `[debt] Stale agent worktrees + branches` (P1, 105 days). Pattern continued through Plans 40/41. | `Backlog.md` P1 | Flagged to user; no autonomous action. |
| 8 | **Low** | `[Plan-29-security-hardening]` Unicode lookalikes item labeled "Purely speculative" — 3+ months, no action. | `Backlog.md` Group B | Flagged to user; candidate for P3 demotion or removal. |
| 9 | **Low** | Group F `[docs]` items (AltScreen release notes, Deviations from Plan) from 2026-04-20 — 105 days stale, release moment long past. | `Backlog.md` Group F | Flagged to user; candidates for removal or closure. |
| 10 | clean | Roadmap cross-reference — Phase 1 complete, Phase 2/3/4 items have appropriate Backlog coverage. | `Roadmap.md` | None. |
| 11 | clean | Status.md Pending (sentrux trial) — correctly placed, blocked by user-action (rustup install). | `Status.md` | None. |
| 12 | clean | All Plan 40 dispatch findings captured in Backlog P2 (2026-06-13 entries). | `Backlog.md` | None. |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

1. **[IMMEDIATE] HOMEBREW_TAP_TOKEN PAT expiry** — The reminder date was ~2026-07-15. Today is 2026-08-03. If the PAT has not been rotated, the next `bonsai` release will fail at the Homebrew tap step with `401 Bad credentials`. Rotate via GitHub → Settings → Developer Settings → Fine-grained PATs, then `gh secret set HOMEBREW_TAP_TOKEN --repo LastStep/Bonsai`. After rotation, update the next expiry calendar reminder (~90 days from rotation date).

2. **[URGENT] Routine gap — all 7 routines overdue.** No routine runs have been logged since 2026-05-07. Recommend scheduling a full routine-digest pass to cover: Vulnerability Scan, Dependency Audit, Doc Freshness Check, Memory Consolidation, Status Hygiene, and Roadmap Accuracy. The ~3-month gap means security and dependency state is unknown.

3. **[Security] Website npm vulns (P2)** — `/website` has open CVEs (esbuild HIGH, vite HIGH+MED, js-yaml MED). Astro upgrade PR #108 fails build after rebase and needs a manual upgrade-with-build-fix pass. Consider promoting to P1 for next session given the HIGH severity items.

4. **[Housekeeping] Stale worktrees** — The worktree leak pattern continued through Plans 40/41 sessions. A sweep of `.claude/worktrees/` and stale remote branches on `origin/` is overdue (105 days since last audit).

5. **[Stale speculation] Unicode lookalikes (Group B)** — The remaining `[Plan-29-security-hardening]` item is explicitly labeled "Purely speculative." No filesystem traversal risk. Recommend demoting to P3 or removing from Group B entirely to reduce noise.

6. **[Stale docs] Group F release-note items** — Two items added 2026-04-20 about Plan 15 AltScreen release notes and "Deviations from Plan" documentation. The release moment for Plan 15 is long past. Recommend removing from Backlog (the AltScreen behavior is now standard; the Deviations guidance could be incorporated into planning-template skill if still wanted).

## Notes for Next Run

- P0 section is now empty — if a true P0 surfaces before the next run, it should be added directly to Status.md per the priority guide, not to Backlog.
- Consider scheduling a backlog consolidation session: the P3 section has ~30 items, several research items from 2026-04-13 are now 3.5 months old with no traction. Some may warrant archival to `Research/` notes.
- The routine system needs to be re-activated. If the loop.md dispatch was interrupted, check that it's still running.
- Group B testing-infra items (P1) have been 30+ days for the first time this cycle — they will be 30+ days again next run. A dedicated "test coverage sprint" plan may be warranted.
- Website npm vuln (P2) is now the top active security item after P0 cleanup. Flag again next cycle if no action taken.
