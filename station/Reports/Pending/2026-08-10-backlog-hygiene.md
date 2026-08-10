---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Backlog Hygiene"
date: 2026-08-10
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
- **Files Modified:** 3 — `/home/user/Bonsai/station/Playbook/Backlog.md`, `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`
- **Tools Used:** Read, Edit, Write, Bash (ls)
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Escalate misplaced P0s
- **Action:** Read `Backlog.md` P0 section and compared each item against `Status.md` In Progress and Pending.
- **Result:** Found 2 active P0 items. Neither was in Status.md Pending or In Progress — both were in Status.md Recently Done, confirming resolution. Escalation not needed; cleanup performed in Step 2.
  - P0 sensor bug → resolved by v0.4.3 hotfix (PRs #105/#106)
  - P0 non-interactive flags → resolved by v0.4.2 release (PR #102)
- **Issues:** none — no genuinely untracked P0 items requiring immediate escalation

### Step 2: Cross-reference with Status.md
- **Action:** Read `Status.md`, compared all Backlog entries against In Progress, Pending, and Recently Done.
- **Result:**
  - **3 items removed (commented out with audit trail):**
    1. `[bug] Sensor hook commands use $PWD-walk-up` (P0) — resolved by v0.4.3 (2026-05-13)
    2. `[feature] bonsai init / bonsai add need non-interactive flags` (P0) — resolved by v0.4.2 (2026-05-13)
    3. `[feature] Full agent-drivable (non-interactive) CLI parity` (P1) — resolved by Plan 41 (2026-06-16, PRs #120–#125)
  - **1 item already handled (already commented out):** `[research] Trial sentrux` — promoted to Status.md Pending 2026-05-07; still blocked on Rust toolchain (no change needed)
  - **Blocking check:** Status.md Pending has only the sentrux item (blocked on Rust toolchain); no Backlog item can unblock it.
- **Issues:** none

### Step 3: Cross-reference with Roadmap.md
- **Action:** Read `Roadmap.md`, checked Phase alignment.
- **Result:**
  - **Phase 1 is fully complete** (all checkboxes marked). No stale Phase 1 references found in the Backlog requiring cleanup.
  - **Phase 2 (Extensibility) candidates** — several P2 Backlog items align and could be considered for promotion:
    - `[security] Harden all scaffolding writes against symlink substitution` — directly relevant to Phase 2 robustness
    - `[improvement] bonsai validate warn on .bonsai/project.yaml ↔ .bonsai.yaml drift` — Phase 2 extensibility quality
    - `[feature] Integrate plan-grilling as first-class Bonsai catalog ability` — expands catalog (Phase 2)
    - Group D items (unbuilt agents, documentation-standards skill, routines) — squarely Phase 2
  - **Phase 2 item without Backlog tracking:** "Template variables expansion" has no matching Backlog entry. Adding to flags.
  - No items referencing deprecated approaches or completed phases found (after removal of the 3 resolved items).
- **Issues:** Phase 2 "Template variables expansion" has no Backlog tracking entry (flagged for user)

### Step 4: Flag stale items
- **Action:** Scanned all priority tiers for items unchanged 30+ days, missing context, or near-duplicates.
- **Result:**
  - **URGENT — HOMEBREW_TAP_TOKEN PAT reminder expired:** The P1 item `[ops] HOMEBREW_TAP_TOKEN PAT expiry calendar reminder` was filed 2026-04-22 with a target rotation date of ~2026-07-15. Today is 2026-08-10 — that date has passed by 26 days. No Status.md entry confirms the rotation occurred. If the PAT expired, the next release will fail at the Homebrew tap step. **Flagged for immediate user verification.**
  - **Items 30+ days old without progress (sample):**
    - Group A: `[bookkeeping] Retroactively trim Backlog entries to NoteStandards` — added 2026-04-25, no progress. Still valid.
    - Group C: `[improvement] OSS polish — demo GIF/asciinema` — added 2026-04-16, requires user recording, no progress.
    - Group F: `[docs] Document AltScreen behavior change in release notes` — added 2026-04-20 re: Plan 15. Multiple releases shipped since (v0.2.0–v0.5.0). The moment for this note has passed — potentially stale/obsolete. Flagged for user decision.
    - Ungrouped P2: `[improvement] Install semgrep` — added 2026-04-16, narrowed 2026-05-04. Still pending.
  - **Near-duplicates:** No new near-duplicates found. Previously noted CHANGELOG duplication (Group C / Group D) persists; previous routines did not resolve it.
  - **Items missing context/rationale:** None found — all items have source annotations.
- **Issues:** 1 potentially time-critical (PAT expiry), 1 potentially obsolete (AltScreen release note)

### Step 5: Check for routine-generated items
- **Action:** Read `RoutineLog.md` entries since last backlog-hygiene run (2026-05-07).
- **Result:** No routine log entries exist between 2026-05-07 and 2026-08-10. The log shows the 2026-06-13 Plan 40 dispatch entry (not a routine) and then nothing. All 7 routines are severely overdue (95–99 days since last run, vs. 5–14 day frequencies). This means no routine-generated findings exist to check for Backlog capture since the last run.
  - Items added to Backlog from session work (2026-06-13 Plan 40, 2026-06-16 Plan 41) are already captured.
  - No uncaptured routine findings found — but this is because routines haven't run, not because they found nothing.
- **Issues:** Overdue routines are a systemic concern; flagged for user (not auto-adding to backlog)

### Step 6: Promote ready items via issue-to-implementation
- **Action:** Assessed whether any items warrant promotion with user confirmation.
- **Result:** No items are unambiguously ready for autonomous promotion. Items flagged for user decision:
  - `[ops] HOMEBREW_TAP_TOKEN PAT` — needs user to verify rotation status (ops action, not code work)
  - `[security] Website npm vuln tree` (P2) — active vulnerabilities in website npm deps; user may want to promote to P1 and dispatch
  - `[debt] Unify remove business logic` (P2) — filed as Plan 41 debt follow-up; clean implementation candidate
- **Issues:** Presenting to user for decision — no autonomous promotion per procedure

### Steps 7–8: Log and Dashboard (handled post-report)

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | high | P0 sensor bug resolved but still in Backlog as active | Backlog P0 | Commented out with audit trail |
| 2 | high | P0 non-interactive flags resolved but still in Backlog as active | Backlog P0 | Commented out with audit trail |
| 3 | high | P1 full CLI parity resolved by Plan 41 but still in Backlog as active | Backlog P1 | Commented out with audit trail |
| 4 | high | HOMEBREW_TAP_TOKEN PAT reminder date (2026-07-15) has passed — rotation unconfirmed | Backlog P1 | Flagged for user |
| 5 | medium | Website npm vulnerabilities (P2) — active Dependabot alerts, build-breaking astro upgrade | Backlog P2 | Flagged for user promotion decision |
| 6 | medium | All 7 routines overdue 95–99 days — no routine findings since 2026-05-07 | routines.md | Flagged for user |
| 7 | low | Roadmap Phase 2 item "Template variables expansion" has no Backlog entry | Roadmap.md | Flagged for user |
| 8 | low | Group F `[docs] AltScreen behavior change in release notes` likely obsolete (multiple releases shipped) | Backlog Group F | Flagged for user decision |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

1. **HOMEBREW_TAP_TOKEN PAT rotation** — The reminder date 2026-07-15 has passed. Verify PAT was rotated; if not, rotate immediately before the next release attempt. File: `Backlog.md` P1 ops item.

2. **All routines are severely overdue (95–99 days)** — Dependency Audit, Doc Freshness Check, Memory Consolidation, Roadmap Accuracy, Status Hygiene, and Vulnerability Scan are all overdue. Consider running a batch routine-digest session soon.

3. **Website npm vuln tree (P2)** — Active Dependabot alerts on `/website` (esbuild HIGH, vite HIGH+MED, js-yaml MED); astro upgrade breaks build. Promote to P1 or schedule a dedicated npm fix session?

4. **Phase 2 Roadmap item "Template variables expansion" untracked** — No Backlog entry captures this Phase 2 goal. Worth adding if it's still on the roadmap.

5. **Group F AltScreen release note item (obsolete?)** — `[docs] Document AltScreen behavior change in release notes` was filed 2026-04-20 about Plan 15. Six releases have shipped since. Should this be closed/removed as the window has passed?

6. **P0 section is now empty** — Both P0 items were resolved. This is good news. The P0 section now contains only HTML comments. The "Trial sentrux" research item remains in Status.md Pending (blocked on Rust toolchain install) — is there intent to install Rust/cargo?

## Notes for Next Run

- P0 section is now clean — both active P0s were resolved during the 2026-05-07–2026-08-10 gap (v0.4.2, v0.4.3, Plan 41)
- P1 "Full agent-drivable CLI parity" resolved by Plan 41; only remaining P1s are ops items and debt
- HOMEBREW_TAP_TOKEN PAT expiry tracking should be moved from Backlog to an external calendar/reminder system — the Backlog is not a good place for time-sensitive calendar items
- If the sentrux trial is abandoned (Rust toolchain install not planned), remove from Status.md Pending
- After running the overdue routines (especially dependency-audit and vulnerability-scan), check if new backlog items need capturing
