---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-08-31
status: success
---

# Routine Report — Doc Freshness Check

## Overview
- **Routine:** Doc Freshness Check
- **Frequency:** Every 7 days
- **Last Ran:** 2026-05-04
- **Triggered By:** loop.md autonomous dispatch

## Execution Metadata
- **Status:** success
- **Duration:** ~6 min
- **Files Read:** 9 —
  - `station/agent/Routines/doc-freshness-check.md`
  - `station/INDEX.md`
  - `station/CLAUDE.md` (via system-reminder)
  - `station/agent/Core/routines.md`
  - `station/Playbook/Status.md`
  - `station/Logs/RoutineLog.md`
  - `docs/agent-interface.md` (preview)
  - `station/agent/Workflows/plan-grilling.md` (preview)
  - `station/agent/Skills/critic-agent-prompts.md` (preview)
- **Files Modified:** 2 — `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`
- **Tools Used:** `git log --oneline -20`, `git log --since/--until`, `ls` (directory listings), `grep` (Backlog search), `head` (file previews)
- **Errors Encountered:** 0

---

## Procedure Walkthrough

### Step 1: Scan project documentation against recent git history
- **Action:** Read `station/INDEX.md`, `station/CLAUDE.md`, `station/Playbook/Status.md`, and ran `git log` to identify commits since last run (2026-05-04). Total span: ~119 days.
- **Result:** Three significant feature deliveries since last check:
  - **Plan 40** — `bonsai validate` command, freeze schemas, root-relative scaffolding, `docs/formats.md`
  - **Plan 41** — Headless CLI contract for all mutating commands; `list --json`; `docs/agent-interface.md` contract document (Phase 5, merged in PR #125)
  - **v0.4.3 hotfix** — sensor hooks now bake absolute paths in `.claude/settings.json`
- **Issues:** `docs/agent-interface.md` is a new top-level project document with no entry in `station/INDEX.md` Document Registry.

### Step 2: Check INDEX.md accuracy
- **Action:** Verified tech stack table, key metrics, architecture overview, and document registry against actual project state.
- **Result:**
  - Tech stack (Go 1.25+, Cobra, Huh, LipGloss, BubbleTea, YAML, text/template) — **accurate**
  - Key metrics (6 agent types, ~50 catalog items, 8 CLI commands) — **accurate**
  - Architecture overview (`cmd/`, `internal/catalog|config|generate|validate|wsvalidate|tui/`, `catalog/`) — **accurate**
  - Document Registry — **one omission**: `docs/agent-interface.md` created in Plan 41 Phase 5 is not listed.
- **Issues:** Document Registry needs a new row for `docs/agent-interface.md`.

### Step 3: Check navigation links
- **Action:** Cross-checked every link in `station/CLAUDE.md` navigation tables (Core, Protocols, Workflows, Skills, Sensors) against actual files in `station/agent/`.
- **Result:**
  - Core (3 files): all resolve ✓
  - Protocols (4 files): all resolve ✓
  - Sensors (10 files): all resolve ✓
  - Workflows (9 linked): all linked files resolve ✓ — but **`plan-grilling.md` exists and is not linked**
  - Skills (6 linked): all linked files resolve ✓ — but **`critic-agent-prompts.md` exists and is not linked**
- **Issues:** Two station-local files installed outside catalog (plan-grilling workflow + critic-agent-prompts skill) are missing from CLAUDE.md navigation tables. Both have `source: adapted from ZenGarden` headers and are already tracked in Backlog for full catalog integration.

### Step 4: Report findings
- **Action:** Compiled three findings (below), checked Backlog for pre-existing tracking.
- **Result:** Finding 1 (docs/agent-interface.md) is new and untracked. Findings 2–3 (plan-grilling/critic-agent-prompts) are partially addressed — the catalog-integration work is in Backlog P2, but the interim CLAUDE.md nav gap is not flagged. All three flagged for user decision per routine's "propose, don't execute" rule.
- **Issues:** none

### Step 5: Update dashboard
- **Action:** Updated `station/agent/Core/routines.md` dashboard row for Doc Freshness Check: Last Ran → 2026-08-31, Next Due → 2026-09-07, Status → done.
- **Result:** Done.
- **Issues:** none

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Medium | `docs/agent-interface.md` (Plan 41 Phase 5 contract doc) has no entry in the Document Registry | `station/INDEX.md` Document Registry table | Flagged for user — proposed fix below |
| 2 | Low | `station/agent/Workflows/plan-grilling.md` exists and is active but not linked in CLAUDE.md Workflows nav table | `station/CLAUDE.md` Workflows section | Flagged for user — proposed fix below |
| 3 | Low | `station/agent/Skills/critic-agent-prompts.md` exists and is active but not linked in CLAUDE.md Skills nav table | `station/CLAUDE.md` Skills section | Flagged for user — proposed fix below |

---

## Proposed Fixes (flag for user decision — not executed)

### Fix 1 — Add `docs/agent-interface.md` to INDEX.md Document Registry

Add this row to the Document Registry table in `station/INDEX.md`:

```
| `docs/agent-interface.md` | Headless CLI contract — flags, JSONL/exit-code serializations for driving Bonsai non-interactively from an agent, CI script, or MCP wrapper | Plan 42 (MCP server) work; any automated Bonsai integration |
```

### Fix 2 — Link `plan-grilling.md` in CLAUDE.md Workflows table

Add this row to the Workflows table in `station/CLAUDE.md`:

```
| Running an adversarial plan review via 6 critic agents before dispatch | [agent/Workflows/plan-grilling.md](agent/Workflows/plan-grilling.md) |
```

### Fix 3 — Link `critic-agent-prompts.md` in CLAUDE.md Skills table

Add this row to the Skills table in `station/CLAUDE.md`:

```
| Prompt templates for the 6 plan-grilling critic agents (5 prose + Reality). Consumed by the plan-grilling workflow. | [agent/Skills/critic-agent-prompts.md](agent/Skills/critic-agent-prompts.md) |
```

---

## Errors & Warnings

No errors encountered.

---

## Items Flagged for User Review

- **Finding 1 (Medium):** Add `docs/agent-interface.md` to `station/INDEX.md` Document Registry — low effort, clear value, recommend applying now.
- **Finding 2 (Low):** Link `plan-grilling.md` in `station/CLAUDE.md` Workflows nav — interim measure while catalog integration (Backlog P2) is pending; recommend applying to avoid nav drift until the full catalog item ships.
- **Finding 3 (Low):** Link `critic-agent-prompts.md` in `station/CLAUDE.md` Skills nav — same rationale as Finding 2.

---

## Notes for Next Run

- Last run spanned 119 days (well overdue). Routine should run on the 7-day cadence going forward.
- Backlog Hygiene (run earlier today) flagged Plans 40 and 41 still sitting in `Plans/Active/` — if they are archived before next check, the Status.md reference will be stale. Verify Active/ contents on next run.
- The `bonsai validate` command added in Plan 40 can assist future doc checks by auditing frontmatter and lock-file integrity; consider using it as part of this routine going forward.
