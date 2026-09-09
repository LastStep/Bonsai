---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Roadmap Accuracy"
date: 2026-09-09
status: partial
---

# Routine Report — Roadmap Accuracy

## Overview
- **Routine:** Roadmap Accuracy
- **Frequency:** Every 14 days
- **Last Ran:** 2026-05-07
- **Triggered By:** loop.md autonomous dispatch
- **Gap:** 125 days since last run (8.9 cycles overdue)

## Execution Metadata

| Field | Value |
|-------|-------|
| Executed | 2026-09-09 |
| Mode | Subagent (autonomous, no user interaction) |
| Duration | ~6 min |
| Files Read | Roadmap.md, Status.md, KeyDecisionLog.md, Backlog.md, routines.md, RoutineLog.md, Plans/Active/ listing |
| Files Modified | routines.md (dashboard), RoutineLog.md (append), Reports/Pending/ (this file) |
| Roadmap edits | None — findings flagged for user review per procedure |

## Procedure Walkthrough

### Step 1 — Compare Roadmap against current state

Read `station/Playbook/Roadmap.md`. Compared all Phase 1 items against Status.md and Plans/Archive listing.

**Phase 1 — Foundation & Polish:** All 11 items marked `[x]`. Cross-checked against Status.md shipped work, release history (v0.4.0 → v0.4.3), and archived plans (Plans 30–39). All items are accurate. Phase 1 is complete.

**Phase 2 — Extensibility:** One item marked `[x]` (custom item detection). Three items still `[ ]`:
- Self-update mechanism
- Template variables expansion
- Micro-task fast path

**Phase 3 and Phase 4:** No items checked. Aligns with KDL decision to defer cloud integration.

### Step 2 — Check milestone accuracy

Cross-checked Phase 2 `[ ]` items against Status.md recently-done rows and Plans/Active/:

- **Plan 41 (Headless CLI Contract)** shipped 2026-06-16: pure `*Result` headless cores + JSONL/exit-code contract for all 4 mutating commands + `list --json` + `docs/agent-interface.md`. This is not on the roadmap but represents a substantial Phase 2 deliverable (extensibility surface for agents and external tooling). **MCP server (Plan 42)** is explicitly referenced as its "fast-follow" and is also absent from the roadmap.

- **Micro-task fast path**: The Roadmap.md footnote on "Better trigger sections" already notes "intent-classification prompt-hook deferred to P3 Backlog as Plan 08 C3." If this item is now intended for P3 scope, it should be moved out of Phase 2 in the roadmap.

- **Self-update mechanism**: Roadmap description is "catalog items can flag when they're stale or have issues." The shipped `bonsai validate` command (Phase 1, v0.4.0) already detects orphans, stale lock entries, untracked customs, and frontmatter gaps. Partial overlap — but `validate` is read-only and flags; "self-update" implies auto-remediation or catalog sync. The remainder (auto-fix / catalog sync) is still unbuilt.

- **Template variables expansion**: Not found in any shipped plan. Plan 40 ("Odysseus platform integration") introduced frozen v1 schemas for `.bonsai/project.yaml`, but did not expand template variables in the `.md.tmpl` / `.sh.tmpl` sense. Still unbuilt.

### Step 3 — Cross-check against Key Decision Log

Read `station/Logs/KeyDecisionLog.md`. No KDL entries invalidate existing roadmap items. Relevant cross-checks:

- "Defer Managed Agents cloud integration until local foundation is stable" (2026-04-02) — Phase 3 still marked all `[ ]`, consistent.
- "Go rewrite, embed.FS, text/template, tech-lead required" decisions (2026-04-12/13) — all Phase 1 items already marked complete.

One notable absence: the KDL does not record the decision to add a headless/agent-drivable CLI layer (Plan 41). That was a significant architectural decision. Worth capturing (flagged below for user).

### Step 4 — Report findings

See Findings Summary below. Not modifying Roadmap.md per procedure — all items flagged for user review.

### Step 5 — Update dashboard

Updated `agent/Core/routines.md` Roadmap Accuracy row: Last Ran → 2026-09-09, Next Due → 2026-09-23, Status → done.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | HIGH | Plan 41 (Headless CLI Contract) shipped 2026-06-16 — not on roadmap. Agent-drivable CLI with `*Result` cores, JSONL/exit contract, and `docs/agent-interface.md` is a Phase 2 deliverable that has no entry in Roadmap.md | Roadmap.md Phase 2 | Flagged for user — suggest adding `[x] Headless CLI contract — agent-drivable cores + JSONL/exit-code contract` to Phase 2 |
| 2 | MEDIUM | MCP server (Plan 42) planned as Plan 41 fast-follow — not on roadmap | Roadmap.md Phase 2 | Flagged for user — suggest adding `[ ] MCP server — Plan 42` to Phase 2 or Phase 3 |
| 3 | MEDIUM | "Micro-task fast path" (Phase 2) is already noted in the roadmap as deferred to "P3 Backlog as Plan 08 C3" in a footnote of an unrelated item. If it's been demoted, it should be explicitly moved to Phase 3 rather than left as an unchecked Phase 2 item | Roadmap.md Phase 2 | Flagged for user — clarify whether to move this item to Phase 3 |
| 4 | LOW | Phase 2 "Self-update mechanism" description overlaps with shipped `bonsai validate`. The unbuilt remainder (auto-remediation / catalog sync) is distinct — but the description may confuse whether this is partially done | Roadmap.md Phase 2 | Flagged for user — suggest clarifying the description to distinguish validate (done, Phase 1) from auto-sync (still unbuilt) |
| 5 | LOW | KDL does not record the headless/agent-drivable CLI decision from Plan 41 (2026-06-16). Architectural decisions at that scope belong in the Structural section | KeyDecisionLog.md | Flagged for user — suggest adding a KDL entry for the headless-CLI architectural decision |
| 6 | INFO | HOMEBREW_TAP_TOKEN PAT expiry deadline (2026-07-15) has passed — Backlog P1 row flags this as 55 days overdue. Not a roadmap issue but operationally urgent. Not actioned here | Backlog.md P1 | Flagged for user — verify PAT was rotated; if not, rotate now |

---

## Errors & Warnings

None. All source files were readable and parseable.

---

## Items Flagged for User Review

1. **Add Plan 41 deliverable to Phase 2 roadmap** (Finding #1)
   - Suggested text: `- [x] Headless CLI contract — agent-drivable *Result cores, JSONL/exit-code contract, docs/agent-interface.md (Plan 41, shipped 2026-06-16)`
   - Impact: Without this, the roadmap understates Phase 2 progress.

2. **Add MCP server to roadmap** (Finding #2)
   - Suggested text: `- [ ] MCP server — exposes headless cores via MCP protocol (Plan 42, fast-follow to Plan 41)`
   - Placement: Phase 2 (extensibility) or Phase 3 (cloud) depending on scope.

3. **Clarify micro-task fast path placement** (Finding #3)
   - Currently in Phase 2 `[ ]` but footnoted as deferred to P3. Move to Phase 3 or explicitly keep in Phase 2 with a "low priority" note.

4. **Clarify self-update mechanism scope** (Finding #4)
   - Current description conflates detection (covered by validate) with remediation (still unbuilt). Suggest: `- [ ] Self-update mechanism — auto-sync/apply updated catalog items (validate covers detection; remediation path still unbuilt)`

5. **Log Plan 41 architectural decision in KDL** (Finding #5)
   - Suggested entry under Structural: `**2026-06-16** — All mutating commands expose pure *Result headless cores with JSONL/exit-code contract. Rationale: enables agent-drivable orchestration, MCP integration, and non-interactive scripting without re-implementing business logic outside the binary. See docs/agent-interface.md.`

6. **Verify HOMEBREW_TAP_TOKEN PAT rotation** (Finding #6)
   - Backlog P1 row notes deadline passed 2026-07-15 and is now 55 days overdue.

---

## Notes for Next Run

- Roadmap is structurally sound but has drifted from shipped work over the 4-month gap since last run.
- Phase 1 is fully complete and accurately marked.
- Phase 2 has two shipped items (custom item detection + headless CLI), but only one is checked.
- No KDL decisions invalidate any roadmap phases — cloud deferral policy still stands.
- If Plan 42 (MCP server) ships before the next run, it should be added to the roadmap and checked.
- Next run (2026-09-23): check status of Plan 42 and whether v0.5.0 tag has been applied.
