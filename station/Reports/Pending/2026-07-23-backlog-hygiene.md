---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Backlog Hygiene"
date: 2026-07-23
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
- **Duration:** ~10 min
- **Files Read:** 5 — `station/agent/Routines/backlog-hygiene.md`, `station/Playbook/Backlog.md`, `station/Playbook/Status.md`, `station/Playbook/Roadmap.md`, `station/Logs/RoutineLog.md`
- **Files Modified:** 2 — `station/Playbook/Backlog.md`, `station/agent/Core/routines.md`
- **Tools Used:** Read, Edit, Write, Bash
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Escalate misplaced P0s
- **Action:** Read Backlog.md P0 section; cross-referenced each active P0 item against Status.md.
- **Result:** Both active P0 items are already resolved and shipped — see items removed in Step 2. P0 section is now empty of active items (two HTML comments remain as audit trail). The commented-out sentrux P0 was already promoted to Status.md Pending in the 2026-05-07 routine-digest.
- **Issues:** None — no misplaced P0s requiring escalation. All former P0s have shipped.

### Step 2: Cross-reference with Status.md
- **Action:** Read Status.md In Progress, Pending, and Recently Done tables; matched against every Backlog item.
- **Result:** Three items confirmed resolved by Status.md recently-done entries and removed from the active backlog (commented out with resolution notes):
  1. **P0 [bug] Sensor hook $PWD-walk-up** — Resolved by v0.4.3 hotfix (PRs #105/#106, 2026-05-13). Status.md entry: "sensor hook commands now bake install-time absolute paths."
  2. **P0 [feature] Non-interactive flags for init/add** — Resolved by v0.4.2 release (PR #102, 2026-05-13). Status.md entry: "`--non-interactive --from-config <path>` (JSONL stdout, hard-skip conflicts, exit codes 0/2/3/4)."
  3. **P1 [feature] Full agent-drivable CLI parity (init/update/add/remove)** — Resolved by Plan 41 (PRs #120/#122/#123/#121/#125, 2026-06-16). Status.md entry: "Every mutating cmd (init/add/update/remove) has a pure `*Result` headless core + JSONL/exit contract (`ExitConflict=5`); `list --json`; `docs/agent-interface.md` contract doc."
- **Cross-ref Blocked items:** Status.md Pending has one item: "[research] Trial sentrux" — blocked by Rust toolchain not installed. No Backlog item can unblock this directly.
- **Issues:** None — all resolutions were unambiguous.

### Step 3: Cross-reference with Roadmap.md
- **Action:** Read Roadmap.md; checked P2/P3 Backlog items against current phase milestones.
- **Result:** Phase 1 (Foundation & Polish) is 100% complete — all checkboxes marked. Phase 2 (Extensibility) is active. Relevant Backlog items:
  - P3 "Self-update mechanism" and P3 "Micro-task fast path" align with Phase 2 milestones but are aspirational, no clear trigger to promote. No promotions warranted at this time.
  - Phase 3 (Cloud & Orchestration) items (Managed Agents, Greenhouse) correctly sit at P3 Big Bets.
  - No items reference deprecated or completed-phase approaches that are no longer applicable.
- **Issues:** None found.

### Step 4: Flag stale items
- **Action:** Scanned all backlog items for 30+ day stagnation, unclear rationale, and near-duplicates.
- **Result:** Findings below (flagged for user review — not auto-edited):

  **URGENT — HOMEBREW_TAP_TOKEN PAT expiry (P1 [ops]):** Added 2026-04-22 with a reminder to rotate by ~2026-07-15. Today is 2026-07-23 — that date passed 8 days ago. Fine-grained PATs default to 90-day expiry from rotation date. If the PAT has expired, the next release's GoReleaser brew step will fail with `401 Bad credentials`. User should verify and rotate immediately.

  **Stale — P1 [ops] Routine bot PR pile-up (77 days):** Added 2026-05-07. The 9 stale PRs were already closed; the underlying fix (change cloud routine to commit-direct or auto-merge) has not been implemented. At 77 days without action, user should decide: implement the fix, downgrade to P2, or close as won't-fix.

  **Stale — Group B testing items (90+ days):** The testing infrastructure bundle (trigger/sensor tests, PTY smoke tests, catalog/ test coverage, cmd/ test coverage, generate.go split) was added 2026-04-16 through 2026-04-24. All remain at P1/P2 with no progress in 90+ days. These are valid debt items but clearly deprioritized. No action taken — flag for user acknowledgment.

  **Stale — P3 [debt] Batch refresh outdated Go modules:** Last updated 2026-05-04 (80 days ago). The 23-module count has likely grown since then. Worth sweeping before the next release.

  **Possibly stale — P3 [debt] Rung-3 .bonsai.yaml round-trip:** Added 2026-05-08, references Plan 38 P2 followup. Plan 38 was handed off to the Bonsai-Eval repo (Status.md 2026-05-13: "P2/P3 owned there going forward"). This item may now belong in Bonsai-Eval's backlog rather than here.

  **Near-duplicate check:** No actionable near-duplicates found after removing the resolved items. The remaining items are distinct in scope.

- **Issues:** None requiring immediate auto-edit; all items above flagged for user review.

### Step 5: Check for routine-generated items
- **Action:** Read RoutineLog.md for entries since last backlog-hygiene run (2026-05-07); checked whether any routine findings were not captured in Backlog.
- **Result:** Since 2026-05-07, no new routine runs appear in the log. The 2026-05-07 Routine Digest processed all pending reports (backlog-hygiene, memory-consolidation, roadmap-accuracy, status-hygiene) and applied fixes. Work sessions (Plans 40 and 41) added their own items directly to the backlog.
  - One potential gap: Status.md mentions "MCP server = fast-follow Plan 42" (from Plan 41 entry). No Plan 42 item or MCP server feature exists in the backlog. This was not added by a routine but may warrant a backlog entry.
- **Issues:** No uncaptured routine findings. One possible omission flagged for user (MCP server / Plan 42).

### Step 6: Promote ready items (SKIPPED)
- **Action:** Skipped per task instructions — no user present to confirm promotions.
- **Issues:** None.

### Step 7: Log results
- **Action:** Appended entry to `station/Logs/RoutineLog.md`.
- **Issues:** None.

### Step 8: Update dashboard
- **Action:** Updated `station/agent/Core/routines.md` — Backlog Hygiene row: Last Ran → 2026-07-23, Next Due → 2026-07-30, Status → done.
- **Issues:** None.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Resolved | P0 [bug] Sensor hook $PWD-walk-up — shipped in v0.4.3 | Backlog.md P0 | Commented out with resolution note |
| 2 | Resolved | P0 [feature] Non-interactive init/add flags — shipped in v0.4.2 | Backlog.md P0 | Commented out with resolution note |
| 3 | Resolved | P1 [feature] Full agent-drivable CLI parity — shipped in Plan 41 | Backlog.md P1 | Commented out with resolution note |
| 4 | **URGENT** | HOMEBREW_TAP_TOKEN PAT reminder date (2026-07-15) has passed — may have expired | Backlog.md P1 | Flagged for user |
| 5 | Medium | P1 [ops] Routine bot PR pile-up — 77 days without fix | Backlog.md P1 | Flagged for user |
| 6 | Low | P3 [debt] Rung-3 round-trip belongs in Bonsai-Eval repo, not here | Backlog.md P3 | Flagged for user |
| 7 | Low | MCP server / Plan 42 not tracked in backlog | Backlog.md (missing) | Flagged for user |
| 8 | Info | Group B testing items — 90+ days at same priority, no progress | Backlog.md P1/P2 | Flagged for user awareness |
| 9 | Info | P3 Go modules 80+ days stale since last update | Backlog.md P3 | Flagged for user |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

- **URGENT: HOMEBREW_TAP_TOKEN PAT** — The calendar reminder date (~2026-07-15) passed 8 days ago. Rotate `HOMEBREW_TAP_TOKEN` in GitHub repo secrets now to avoid `401 Bad credentials` on the next GoReleaser brew step. (Backlog P1 `[ops] HOMEBREW_TAP_TOKEN PAT expiry`)

- **P1 [ops] Routine bot PR pile-up** — 77 days at P1 without the fix being applied. Decide: implement the cloud-routine commit/merge fix, downgrade to P2, or close as won't-fix.

- **P3 [debt] Rung-3 `.bonsai.yaml` round-trip** — Plan 38 was explicitly handed off to Bonsai-Eval repo ("P2/P3 owned there going forward"). This item likely belongs in that repo's backlog. Consider removing from here.

- **Missing backlog entry: MCP server / Plan 42** — Status.md (Plan 41 entry) says "MCP server = fast-follow Plan 42." No corresponding backlog item exists. Add a P1 feature item if this is intended to be tracked here.

- **Group B testing debt (90+ days)** — The full testing-infrastructure bundle (trigger tests, PTY smoke, catalog coverage, cmd/ coverage, generate.go split) remains at P1/P2 with no progress since April 2026. Consider: (a) keep as-is (long-term debt), (b) explicitly downgrade to P2/P3, or (c) pick one item for a near-term plan.

- **P3 Go module hygiene** — 23+ modules behind as of 2026-05-04 (80 days ago); count has likely grown. Sweep before next release to avoid accumulating CVE exposure.

## Notes for Next Run

- P0 section is now empty of active items — only two resolution comments remain. If a new P0 surfaces it will stand out clearly.
- 77 days have elapsed since the last backlog-hygiene run (2026-05-07 → 2026-07-23). This exceeds the 7-day frequency significantly. The routine is overdue and the routing lag shows in the stale items found.
- The HOMEBREW_TAP_TOKEN PAT is the most time-sensitive finding — it requires action before the next release regardless of backlog prioritization.
- If the user adds the MCP server / Plan 42 item to the backlog, next run should verify it's properly scoped.
