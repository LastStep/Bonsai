---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Backlog Hygiene"
date: 2026-09-13
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
- **Files Modified:** 2 — `/home/user/Bonsai/station/Playbook/Backlog.md` (3 items resolved), `/home/user/Bonsai/station/agent/Core/routines.md` (dashboard updated)
- **Tools Used:** Read, Edit, Write
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Escalate misplaced P0s
- **Action:** Read Backlog.md P0 section; cross-referenced each active P0 item against Status.md.
- **Result:** Both active P0 items are RESOLVED and should not be in the active P0 section:
  - `[bug] Sensor hook commands use $PWD-walk-up` — resolved by v0.4.3 hotfix (2026-05-13, PRs #105/#106). Status.md confirms: "sensor hook commands now bake install-time absolute paths."
  - `[feature] bonsai init / bonsai add need non-interactive flags` — resolved by v0.4.2 (2026-05-13, PR #102). Status.md confirms: `--non-interactive --from-config` shipped for both commands.
  - Both replaced with HTML resolution comments in Backlog.md.
  - Note: sentrux P0 already correctly commented out; tracked in Status.md Pending.
- **Issues:** None — resolved autonomously.

### Step 2: Cross-reference with Status.md
- **Action:** Read Status.md In Progress, Pending, and Recently Done tables; checked each against Backlog active items.
- **Result:**
  - **P1 `[feature] Full agent-drivable (non-interactive) CLI parity`** — resolved by Plan 41 (2026-06-16, PRs #120/#122/#123/#121/#125). Status.md confirms: "Every mutating cmd (init/add/update/remove) has a pure *Result headless core + JSONL/exit contract." Item removed from Backlog.
  - **Pending item cross-ref:** `[research] Trial sentrux on Bonsai repo` — still blocked on Rust toolchain (cargo/rustc). No Backlog item would unblock this; requires user action to install rustup.
  - No Backlog items unblock the only blocked Pending item.
- **Issues:** None.

### Step 3: Cross-reference with Roadmap.md
- **Action:** Read Roadmap.md; compared Phase 2 milestones against Backlog items.
- **Result:**
  - Phase 1 is fully complete (all checkboxes checked).
  - Phase 2 milestones: "Self-update mechanism" maps to Backlog P3 `[improvement] Self-update mechanism`; "Micro-task fast path" maps to Backlog P3 `[improvement] Micro-task fast path`. Both are Phase 2 goals sitting at P3 — candidates for P2 promotion at user's discretion.
  - "Template variables expansion" (Phase 2 milestone) has **no corresponding Backlog entry** — gap worth noting.
  - No items reference deprecated approaches or completed phases that aren't already resolved.
- **Issues:** Template variables expansion gap — flagged for user review.

### Step 4: Flag stale items
- **Action:** Scanned all P0–P3 items for age (30+ days without progress), context gaps, and near-duplicates.
- **Result:**
  - **HIGH URGENCY — HOMEBREW_TAP_TOKEN PAT:** P1 item `[ops] HOMEBREW_TAP_TOKEN PAT expiry calendar reminder` was added 2026-04-22; PAT rotation was due ~2026-07-15. Today is 2026-09-13 — the PAT is ~60 days overdue. Any release cut now will fail at the Homebrew tap step with 401 Bad credentials. Flagged for immediate user action.
  - **P2 Security — website npm vulns:** `[security] Website npm vuln tree — astro upgrade breaks npm run build` added 2026-06-16 (89 days ago). High-severity esbuild + vite vulns unresolved; astro upgrade still broken. Security debt accumulating.
  - **P1 `[debt] Testing infrastructure for triggers and sensors`:** Added 2026-04-16 (150+ days). No progress. Eligible for de-prioritization to P2 if still not planned.
  - **P1 `[debt] Stale agent worktrees + branches`:** Added 2026-04-20 (146+ days). No progress signal in logs.
  - **P1 `[ops] Routine bot PR pile-up`:** Added 2026-05-07 (129 days). Fix (commit-direct-to-main or auto-merge) not implemented.
  - Near-duplicates: None remaining after P0/P1 cleanups.
  - Items with no clear rationale: none found — all items have context.
- **Issues:** HOMEBREW_TAP_TOKEN PAT expiry is the highest-urgency flag.

### Step 5: Check for routine-generated items
- **Action:** Read RoutineLog.md entries since last backlog-hygiene (2026-05-07).
- **Result:**
  - Routines have not been run since the 2026-05-07 cycle (dashboard confirms last-ran dates are all in May). This run is 129 days overdue (7-day routine).
  - Non-routine session activity (Plans 40 + 41, 2026-06-13/16) filed Backlog items directly: "Harden all scaffolding writes," "bonsai validate can't pass," "bonsai validate warn on project.yaml drift," "Plan 40 review nits," "Unify remove business logic," "Website npm vuln tree" — all present in Backlog. No uncaptured findings.
  - No routine flagged issues in the window (no routines ran).
- **Issues:** 129-day routine gap itself is notable context for the user.

### Step 6: Promote ready items via issue-to-implementation
- **Action:** Checked for items approved for implementation or unambiguous P0 requiring dispatch.
- **Result:** No items meet the bar for autonomous i2i routing without user confirmation. HOMEBREW_TAP_TOKEN PAT rotation is urgent but requires user action (PAT creation on GitHub, not code dispatch).
- **Issues:** None.

### Steps 7 & 8: Log and dashboard
- **Action:** Appended to RoutineLog.md; updated routines.md dashboard.
- **Result:** Done.
- **Issues:** None.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | High | P0 `[bug] Sensor hook $PWD-walk-up` was resolved (v0.4.3) but still in active Backlog | Backlog.md P0 | Replaced with HTML resolution comment |
| 2 | High | P0 `[feature] non-interactive flags` was resolved (v0.4.2) but still in active Backlog | Backlog.md P0 | Replaced with HTML resolution comment |
| 3 | High | P1 `[feature] Full agent-drivable CLI parity` was resolved (Plan 41) but still in active Backlog | Backlog.md P1 | Replaced with HTML resolution comment |
| 4 | High | HOMEBREW_TAP_TOKEN PAT was due 2026-07-15 — now 60+ days expired | Backlog.md P1 | Flagged for immediate user action |
| 5 | Medium | P2 website npm security vulns (HIGH esbuild/vite) — 89 days unresolved | Backlog.md P2 | Flagged for user review |
| 6 | Low | Phase 2 Roadmap: "Template variables expansion" has no Backlog entry | Roadmap.md | Flagged for user review |
| 7 | Low | Phase 2 Roadmap items (Self-update, Micro-task fast path) sitting at P3 | Backlog.md P3 | Flagged as P2-promotion candidates |

## Errors & Warnings
No errors encountered.

## Items Flagged for User Review

1. **HOMEBREW_TAP_TOKEN PAT expiry (IMMEDIATE):** The PAT was due ~2026-07-15. It is now 2026-09-13. The next release will fail at the Homebrew formula update step with `401 Bad credentials`. Rotate the PAT at https://github.com/settings/personal-access-tokens and update the `HOMEBREW_TAP_TOKEN` secret on `LastStep/Bonsai` before cutting any release.

2. **Website npm security vulns (P2 debt):** `[security] Website npm vuln tree` (Backlog P2, added 2026-06-16) — esbuild HIGH, vite HIGH+MED vulns still open. PR #108's astro bump fails the website build after rebase. This needs a real upgrade-with-build-fix pass. Consider scheduling.

3. **Template variables expansion (Roadmap gap):** Phase 2 milestone "Template variables expansion" has no Backlog tracking entry. If this is in scope for the current phase, add a Backlog item.

4. **Phase 2 promotion candidates (P3 → P2):** `[improvement] Self-update mechanism` and `[improvement] Micro-task fast path` both map directly to Phase 2 Roadmap milestones but sit at P3. Consider promoting if Phase 2 work is beginning.

5. **Routines 129 days overdue:** All routines were last run 2026-05-07. If the loop dispatch is running again, the other 6 routines (Dependency Audit, Doc Freshness Check, Memory Consolidation, Roadmap Accuracy, Status Hygiene, Vulnerability Scan) are all significantly overdue.

## Notes for Next Run
- P0 section should now be empty of active items (both resolved, sentrux tracked in Status.md Pending).
- P1 CLI parity resolved; next major P1 items are testing infrastructure, worktree cleanup, and routine bot fix.
- Verify HOMEBREW_TAP_TOKEN rotation has happened before next release.
- If Phase 2 work is starting, run Roadmap Accuracy routine to validate phase alignment.
