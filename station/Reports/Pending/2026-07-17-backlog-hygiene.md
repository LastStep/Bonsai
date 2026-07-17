---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Backlog Hygiene"
date: 2026-07-17
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
- **Duration:** ~8 minutes
- **Files Read:** 5 — `/home/user/Bonsai/station/agent/Routines/backlog-hygiene.md`, `/home/user/Bonsai/station/Playbook/Backlog.md`, `/home/user/Bonsai/station/Playbook/Status.md`, `/home/user/Bonsai/station/Playbook/Roadmap.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`
- **Files Modified:** 2 — `/home/user/Bonsai/station/Playbook/Backlog.md`, `/home/user/Bonsai/station/agent/Core/routines.md`
- **Tools Used:** Read, Edit, Write, Bash (ls), Glob
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Escalate misplaced P0s
- **Action:** Read Backlog.md P0 section; identified 2 active P0 items and cross-referenced with Status.md.
- **Result:** Both P0 items are fully RESOLVED — no unescalated P0s remain.
  - `[bug] Sensor hook commands use $PWD-walk-up` — RESOLVED by v0.4.3 hotfix (PRs #105/#106, 2026-05-13). Absolute install-time paths now baked in.
  - `[feature] bonsai init/add need non-interactive flags` — RESOLVED by v0.4.2 release (2026-05-13). `--non-interactive` + `--from-config` shipped.
  - Both were commented out in Backlog.md using the established HTML comment pattern.
- **Issues:** None — no active P0s require escalation to Status.md.

### Step 2: Cross-reference with Status.md
- **Action:** Read Status.md In Progress, Pending, and Recently Done sections; compared against all Backlog items.
- **Result:** 3 Backlog items found to be fully resolved by recently completed plans:
  1. P0 bug (sensor hooks) → v0.4.3 hotfix 2026-05-13
  2. P0 feature (non-interactive flags) → v0.4.2 release 2026-05-13
  3. P1 feature (full agent-drivable CLI parity: init/update/add/remove) → Plan 41 shipped all 5 phases 2026-06-16 with headless `*Result` cores + JSONL/exit contract + `list --json` + `docs/agent-interface.md`
  
  All three were commented out in Backlog.md. No items in Status.md Pending are blocked by a Backlog item that could be unblocked.
- **Issues:** None.

### Step 3: Cross-reference with Roadmap.md
- **Action:** Read Roadmap.md; checked P2/P3 Backlog items against phase milestones; looked for deprecated references.
- **Result:**
  - Phase 1 is fully complete (all items checked). No stale references.
  - Phase 2 (Extensibility) items are appropriately tracked in P3 Backlog: `self-update mechanism`, `micro-task fast path`.
  - One Phase 2 milestone — "Template variables expansion" — has NO Backlog entry. This is a minor gap; no active work is in flight.
  - Phase 3 items (Managed Agents, Greenhouse) are correctly at P3/Big Bets.
  - No items reference deprecated approaches or completed phases.
- **Issues:** "Template variables expansion" (Phase 2 milestone) is not tracked in the Backlog — flagged for user review.

### Step 4: Flag stale items
- **Action:** Reviewed item add dates against today (2026-07-17); identified items 30+ days old without progress; checked for near-duplicates.
- **Result:**
  - Most Backlog items were added in April–May 2026 (70–90+ days old). This is expected for a P2/P3 backlog — these items are tracked ideas, not active work.
  - **Critical time-sensitive staleness found:** `[ops] HOMEBREW_TAP_TOKEN PAT expiry` (P1) — added 2026-04-22, reminder was set for ~2026-07-15. **Today is 2026-07-17 — the PAT is 2 days overdue for rotation.** If not already rotated, the next release will fail at the Homebrew formula update step with `401 Bad credentials`. This requires immediate user action.
  - No new near-duplicates found. Known coupling between `[improvement] Plan archiving` and `[improvement] Plans Index file` (both Group E) remains appropriate.
  - Group B items (testing/code quality, 6 items) have been sitting untouched since April 2026 — no change in priority, but flagging the cluster for re-prioritization consideration.
- **Issues:** HOMEBREW_TAP_TOKEN PAT expiry — see Items Flagged for User Review.

### Step 5: Check for routine-generated items
- **Action:** Read RoutineLog.md entries since last backlog-hygiene run (2026-05-07); checked whether any flagged findings were captured in the Backlog.
- **Result:** The 2026-05-07 Backlog Hygiene itself flagged 5 items. Items 1 and 2 were resolved (sentrux promoted to Status.md, "Better trigger sections" quick-fixed in routine-digest). Items 3/4/5 remain uncaptured in the Backlog:
  - `code-index.md` staleness (medium drift, flagged 2026-05-04 + 2026-05-07)
  - Broken nav link `agent/Skills/bonsai-model.md` (flagged 2026-05-07)
  - INDEX.md arch diagram drift (low, flagged 2026-05-07)
  
  No routine ran between 2026-05-07 and today (only Plan 40 dispatch on 2026-06-13, which is not a routine). These 3 items are documentation findings belonging to the doc-freshness domain — not adding to Backlog autonomously per step 5 rule.
- **Issues:** 3 uncaptured doc-freshness findings from 2026-05-07 — flagged for user review.

### Step 6: Promote ready items via issue-to-implementation
- **Action:** Checked whether any Backlog items are approved for immediate implementation or represent P0 emergencies.
- **Result:** No items are user-approved for immediate pickup. The HOMEBREW_TAP_TOKEN PAT item requires user action (manual GitHub PAT rotation) rather than an agent-executed plan.
- **Issues:** None requiring issue-to-implementation dispatch.

### Step 7: Log results
- **Action:** Appended entry to `station/Logs/RoutineLog.md`.
- **Result:** Done.

### Step 8: Update dashboard
- **Action:** Updated `agent/Core/routines.md` Backlog Hygiene row.
- **Result:** Last Ran → 2026-07-17, Next Due → 2026-07-24, Status → done.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | resolved | P0 bug `sensor hook commands $PWD-walk-up` was already fixed by v0.4.3 | Backlog.md P0 | Commented out — resolved 2026-05-13 |
| 2 | resolved | P0 feature `bonsai init/add non-interactive flags` was already shipped in v0.4.2 | Backlog.md P0 | Commented out — resolved 2026-05-13 |
| 3 | resolved | P1 feature `full agent-drivable CLI parity` was shipped by Plan 41 | Backlog.md P1 | Commented out — resolved 2026-06-16 |
| 4 | high | HOMEBREW_TAP_TOKEN PAT expiry overdue by 2 days (~2026-07-15 reminder, today 2026-07-17) | Backlog.md P1 | Flagged for immediate user action |
| 5 | low | "Template variables expansion" (Phase 2 Roadmap milestone) has no Backlog entry | Roadmap.md Phase 2 | Flagged for user review |
| 6 | low | 3 doc-freshness findings from 2026-05-07 uncaptured in Backlog (code-index staleness, bonsai-model nav link, INDEX.md arch drift) | RoutineLog.md | Flagged — not auto-added per step 5 rule |
| 7 | info | Group B testing/quality items (6 items) untouched since April 2026 (90+ days) | Backlog.md P1-P2 | No action — noted for re-prioritization |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

- **[URGENT] HOMEBREW_TAP_TOKEN PAT likely expired.** The PAT on `LastStep/Bonsai` was rotated 2026-04-22. The reminder in the Backlog P1 entry was set for ~2026-07-15 — today is 2026-07-17, 2 days past the reminder. Rotate the PAT in GitHub Settings now and update the `HOMEBREW_TAP_TOKEN` secret in the `LastStep/Bonsai` repo before the next release. Symptom of expired PAT: GoReleaser release succeeds (binaries published) but brew formula update fails with `401 Bad credentials`.

- **[LOW] "Template variables expansion"** (Phase 2 Roadmap milestone) is not tracked anywhere in the Backlog. Decide whether to add a P3 entry or defer until Phase 2 work begins.

- **[LOW] 3 uncaptured doc-freshness findings from the 2026-05-07 Backlog Hygiene run** were never added to the Backlog: (1) `code-index.md` medium drift, (2) broken nav link `agent/Skills/bonsai-model.md`, (3) INDEX.md arch diagram drift. These are doc-only findings. Decide whether to add them to the Backlog or route directly to the doc-freshness-check routine on next run.

## Notes for Next Run

- P0 section is now clean — all items commented out or promoted. The only active items near the top are legitimate P1s.
- The HOMEBREW_TAP_TOKEN PAT needs to be rotated. If done, update or remove the `[ops] HOMEBREW_TAP_TOKEN PAT expiry` Backlog P1 entry and add a new reminder for the new expiry date (~90 days from rotation).
- Plan 41 resolved the major headless CLI parity work. MCP server (Plan 42) is the referenced fast-follow — check if that's been picked up and tracked.
- Last ran: 2026-05-07 (70-day gap). Consider increasing routine cadence monitoring — 70 days between runs means several resolved items accumulated.
