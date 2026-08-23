---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Backlog Hygiene"
date: 2026-08-23
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
- **Files Read:** 6 — `/home/user/Bonsai/station/Playbook/Backlog.md`, `/home/user/Bonsai/station/Playbook/Status.md`, `/home/user/Bonsai/station/Playbook/Roadmap.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`, `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/agent/Core/memory.md`
- **Files Modified:** 3 — `Playbook/Backlog.md` (2 P0 items commented out), `agent/Core/routines.md` (dashboard updated), `Logs/RoutineLog.md` (entry appended)
- **Tools Used:** Read, Edit, Write, Glob
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1 — Escalate misplaced P0s
Read `Backlog.md` P0 section. Found 2 active P0 items (plus 1 already-commented sentrux entry):

1. `[bug] Sensor hook commands use $PWD-walk-up` — Cross-checked against Status.md. **RESOLVED**: v0.4.3 hotfix shipped 2026-05-13 (PRs #105/#106). Absolute-path baking shipped; `bonsai update` refreshes existing projects. Item not in Status.md as In Progress or Pending — it is in Status.md Recently Done ("v0.4.3 hotfix shipped"). Commented out in Backlog.

2. `[feature] bonsai init / bonsai add need non-interactive flags` — Cross-checked against Status.md. **RESOLVED**: v0.4.2 shipped 2026-05-13 (PR #102) with `--non-interactive --from-config`. Superseded by P1 "Full agent-drivable CLI parity" item. Commented out in Backlog.

Neither P0 was legitimately unresolved — both were fixed after the 2026-05-07 last-hygiene run and never cleaned up. **P0 section is now empty of active items (all commented).**

### Step 2 — Cross-reference with Status.md
- Reviewed Status.md In Progress (none), Pending (sentrux trial — already commented out in Backlog), and Recently Done.
- Confirmed the 2 P0 removals above match Status.md Recently Done entries.
- `[research] Trial sentrux` is correctly in Status.md Pending and correctly commented out in Backlog — no change needed.
- Plan 40 Phase 4 remains held; Plan 41 is shipped. Both are in Status.md Recently Done. No new Backlog items need removal from these.
- Checked for Pending "Blocked By" items that Backlog items could unblock: sentrux is blocked by Rust toolchain (not a Backlog item) — no Backlog item can unblock it.

### Step 3 — Cross-reference with Roadmap.md
- Phase 1 is complete (all checkboxes checked).
- Phase 2 — Extensibility: 3 remaining items:
  - "Self-update mechanism" → in Backlog P3 as `[improvement] Self-update mechanism` — Phase 2 is now the active phase, could promote P3→P2.
  - "Template variables expansion" — no Backlog entry at all; minor gap.
  - "Micro-task fast path" → in Backlog P3 as `[improvement] Micro-task fast path` — same as above, could promote.
- Phase 3 — Cloud & Orchestration: both items in Backlog P3 Big Bets section (Managed Agents, Greenhouse app). No deprecated references found.
- No items reference completed phases in misleading ways. Group F heading notes "Group F essentially closed for init" which remains accurate.
- **Flagging for user**: Phase 2 is the active frontier; 2 P3 items map directly to Phase 2 milestones — consider promoting "Self-update mechanism" and "Micro-task fast path" to P2 at next user review.

### Step 4 — Flag stale items
Run was last executed 2026-05-07. Today is 2026-08-23 — a **108-day gap**, well above the 30-day staleness threshold. Many items are technically stale by duration; most are legitimately backlogged rather than forgotten. Key staleness flags:

**HIGH — requires immediate user action:**
- `[ops] HOMEBREW_TAP_TOKEN PAT expiry calendar reminder` (P1, added 2026-04-22) — reminder was set for ~2026-07-15. Today is 2026-08-23, which is **35 days past the reminder date**. Fine-grained PATs default to 90 days; PAT created ~2026-04-22 + 90 days = ~2026-07-21 → **PAT is almost certainly expired**. Next release will fail at Homebrew brew step with 401 unless rotated. Flag for immediate action.

**MEDIUM — consider re-prioritizing:**
- `[debt] Testing infrastructure for triggers and sensors` (P1, added 2026-04-16) — 4+ months old with no progress. The trigger system hasn't been the focus recently. Still valid but could slide to P2 if MCP/headless-parity is the current track.
- `[debt] Stale agent worktrees + branches accumulating` (P1, added 2026-04-20) — Recurring worktree cleanup issue. No dedicated sweep has been done. Consider a one-session Tier-1 sweep.
- `[feature] Full agent-drivable CLI parity` (P1, added 2026-06-13) — 71 days old. Memory confirms MCP server (Plan 42) is the current next track. This item feeds Plan 42. Still active — not stale in the "forgotten" sense.
- `[ops] Routine bot PR pile-up` (P1, added 2026-05-07) — 108 days old. No fix applied. PRs may still be accumulating. Should be resolved as part of next session.
- `[security] Website npm vuln tree` (P2, added 2026-06-16) — 68 days old. PR #108 astro bump still broken. esbuild HIGH and vite HIGH unresolved. Vulnerability-scan routine has not run since 2026-05-04. Flag.

**INFO — acknowledged, no change needed:**
- Group F items (UI/UX testing findings) — mostly from 2026-04-17 dogfooding. Two docs items remain (AltScreen behavior change docs, "Deviations from Plan" reminder). Valid low-priority open items.
- Group B items — long-running technical debt queue; no new urgency surfaced.
- Group C OSS readiness — demo GIF still pending user recording; tab completion still pending.

**Near-duplicates scan:**
- No new near-duplicates identified. Prior known near-duplicate (Group C changelog vs Group D changelog generation skill) is still in place; user was aware from the 2026-04-21 run. Still a valid consolidation candidate.

### Step 5 — Check for routine-generated items since last run
- Reviewed RoutineLog entries after 2026-05-07. The only entry is 2026-06-13 Plan 40 dispatch (a session log, not a routine log).
- **No routines have run since 2026-05-07** (108-day gap across all routines). Vulnerability Scan, Dependency Audit, Doc Freshness Check, Memory Consolidation, Roadmap Accuracy, and Status Hygiene are all massively overdue.
- No routine-generated findings to verify for Backlog capture because no routines have run.
- The existing `[security] Website npm vuln tree` (P2) was filed from the 2026-06-16 Plan 41 session, not a routine — already in Backlog.
- **Flag for user**: 6 other routines are overdue (some by 90+ days). The loop.md dispatcher should prioritize running them.

### Step 6 — Promote ready items via issue-to-implementation
- No user directive to promote any item to implementation this run.
- The `[feature] Full agent-drivable CLI parity` (P1) is the natural next plan target per memory.md ("MCP server = Plan 42"). No autonomous promotion without user confirmation.
- The HOMEBREW_TAP_TOKEN PAT item may need a quick fix session but doesn't require issue-to-implementation workflow.

### Steps 7-8 — Log + Dashboard Update
Completed as required.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | P0-resolved | `[bug] Sensor hook $PWD-walk-up` resolved in v0.4.3 — item left in Backlog past its resolution date | Backlog P0 | Commented out in Backlog.md |
| 2 | P0-resolved | `[feature] bonsai init/add non-interactive flags` resolved in v0.4.2 — item left in Backlog past its resolution date | Backlog P0 | Commented out in Backlog.md |
| 3 | HIGH | HOMEBREW_TAP_TOKEN PAT expiry reminder was 2026-07-15 — today is 2026-08-23; PAT is almost certainly expired (~90-day TTL from 2026-04-22) | Backlog P1 | Flagged for user — immediate rotation needed before next release |
| 4 | MEDIUM | No routines have run since 2026-05-07 (108-day gap) — 6 other routines massively overdue | Routines dashboard | Flagged for user — loop dispatcher should prioritize other routines |
| 5 | MEDIUM | `[security] Website npm vuln tree` (esbuild HIGH, vite HIGH) still unresolved at 68 days old | Backlog P2 | Flagged for user — vulnerability-scan routine overdue |
| 6 | MEDIUM | `[ops] Routine bot PR pile-up` still unresolved at 108 days — PRs may continue accumulating | Backlog P1 | Flagged for user |
| 7 | LOW | Plan 41 file remains in `Plans/Active/` per memory note — should be archived | Plans/Active/ | Flagged for user — archive at next wrap-up |
| 8 | INFO | Phase 2 is active frontier; 2 P3 Backlog items ("Self-update mechanism", "Micro-task fast path") map to Phase 2 milestones | Backlog P3 | Flagged for user — consider promoting to P2 |
| 9 | INFO | No near-duplicate items — prior known duplicate (changelog entries) still present and user-aware | Backlog | No action |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

1. **[URGENT] HOMEBREW_TAP_TOKEN PAT** — Almost certainly expired (set 2026-04-22, 90-day TTL, reminder was 2026-07-15). Rotate now via GitHub → Settings → Developer settings → Personal access tokens. Set secret on `LastStep/Bonsai` repo. Calendar-remind ~90 days forward.

2. **[URGENT] Other routines massively overdue** — Dependency Audit, Vulnerability Scan, Doc Freshness Check, Memory Consolidation, Roadmap Accuracy, Status Hygiene all last ran 2026-05-04 or 2026-05-07 (90–108 days ago). Loop dispatcher should run them. Vulnerability Scan in particular is relevant given the open npm HIGH vulns.

3. **[MEDIUM] Website npm vulns unresolved** — `[security] Website npm vuln tree` (P2, added 2026-06-16): esbuild HIGH, vite HIGH open. PR #108 astro bump still broken. Needs a real upgrade-with-build-fix pass.

4. **[MEDIUM] Routine bot PR pile-up** — Still unresolved 108 days after filing (P1). Fix options: (a) commit-direct-to-main, (b) auto-merge if green + no conflicts, or (c) skip PR when local digest absorbed. Pick one and ship.

5. **[LOW] Plan 41 still in Plans/Active/** — Memory.md notes "archive to Plans/Archive/ at next wrap-up." Plan is shipped (Status.md Recently Done 2026-06-16). Archive it.

6. **[INFO] Promote Phase-2-aligned P3 items** — "Self-update mechanism" and "Micro-task fast path" map directly to Roadmap Phase 2 milestones. Now that Phase 1 is complete, consider P3→P2 promotion.

## Notes for Next Run

- P0 section is now clean — all 3 entries are comments. Future P0 items should be escalated to Status.md within days, not left to accumulate.
- The 108-day gap since last run means all routines are far behind. Prioritize the full routine suite before the next development cycle.
- The HOMEBREW_TAP_TOKEN pattern has now hit twice (v0.2.0 + now). Consider automating PAT expiry tracking (GitHub Actions secret rotation reminder via cron) as a P2 improvement.
- Next backlog-hygiene run should check if `[ops] Routine bot PR pile-up` has been resolved.
