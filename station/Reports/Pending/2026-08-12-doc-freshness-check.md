---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-08-12
status: partial
---

# Routine Report — Doc Freshness Check

## Overview
- **Routine:** Doc Freshness Check
- **Frequency:** Every 7 days
- **Last Ran:** 2026-05-04
- **Triggered By:** loop.md autonomous dispatch

## Execution Metadata
- **Status:** partial
- **Duration:** ~10 minutes
- **Files Read:** 12 — `station/agent/Routines/doc-freshness-check.md`, `station/INDEX.md`, `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`, `station/CLAUDE.md` (via system-reminder), `station/agent/Workflows/plan-grilling.md`, `station/agent/Skills/critic-agent-prompts.md`, `station/Playbook/Roadmap.md`, `station/code-index.md`, `cmd/completion.go`, `station/agent/Core/identity.md` (dir listing), `station/agent/Protocols/` (dir listing)
- **Files Modified:** 2 — `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`
- **Tools Used:** `git log --since="7 days ago" --oneline`, `git log --since="7 days ago" --stat`, `git log --since="30 days ago" --oneline`, `ls` on catalog/, cmd/, agent/Workflows/, agent/Skills/, agent/Sensors/, agent/Routines/, Playbook/
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Scan project documentation against recent git history
- **Action:** Read `station/INDEX.md`, `station/Playbook/Roadmap.md`, and `station/code-index.md`. Ran `git log --since="7 days ago"` and `git log --since="30 days ago"`.
- **Result:** Only one commit in the last 7 days — `cd1949d` (2026-08-12, backlog-hygiene routine). No new features, services, or config changes in recent code history that would require doc updates. Last 30 days also showed only that single commit.
- **Issues:** The maintenance system has been dormant since May 2026 (confirmed by backlog-hygiene report); no feature drift from code changes to document.

### Step 2: Check INDEX.md accuracy
- **Action:** Verified agent types count, catalog item count, CLI command count, and tech stack against actual codebase.
- **Result:**
  - Agent types: INDEX.md says 6 (tech-lead, fullstack, backend, frontend, devops, security) → catalog/agents/ has 6 directories. **Accurate.**
  - Catalog items: INDEX.md says "~50" → actual count is 53 (18 skills + 10 workflows + 4 protocols + 13 sensors + 8 routines). **Accurate (within approximation).**
  - Tech stack (Go 1.25+, Cobra, Huh, LipGloss, BubbleTea, YAML, text/template): **Accurate.**
  - CLI commands: INDEX.md says 8 (init, add, remove, list, catalog, update, guide, validate) → `cmd/` contains 9 commands including `completion`. **Minor gap — see Finding #3.**
- **Issues:** Minor: `completion` command not counted (see findings).

### Step 3: Check navigation links
- **Action:** Verified every file linked in `station/CLAUDE.md` navigation tables (Core, Protocols, Workflows, Skills, Sensors, Routines) against the actual files on disk. Also listed installed files in `agent/Workflows/` and `agent/Skills/` to detect installed-but-unlisted items.
- **Result:**
  - **Core (3 links):** All resolve — identity.md ✓, memory.md ✓, self-awareness.md ✓
  - **Protocols (4 links):** All resolve — memory.md ✓, scope-boundaries.md ✓, security.md ✓, session-start.md ✓
  - **Workflows (9 links):** All 9 listed links resolve ✓. However, `plan-grilling.md` is installed in `agent/Workflows/` but **not listed** in the navigation table. **Gap — see Finding #1.**
  - **Skills (6 links):** All 6 listed links resolve ✓. However, `critic-agent-prompts.md` is installed in `agent/Skills/` but **not listed** in the navigation table. **Gap — see Finding #2.**
  - **Sensors (10 links):** All resolve — context-guard.sh ✓, scope-guard-files.sh ✓, session-context.sh ✓, status-bar.sh ✓, routine-check.sh ✓, agent-review.sh ✓, dispatch-guard.sh ✓, subagent-stop-review.sh ✓, compact-recovery.sh ✓, statusline.sh ✓
  - **Routines (7 links):** All resolve ✓
- **Issues:** 2 installed abilities not listed in navigation table (see findings).

### Step 4: Report findings
- **Action:** Collated all findings from Steps 1–3. Per procedure, flagging for user decision — no doc edits executed.
- **Result:** 3 findings identified (2 medium, 1 low). No broken links. No architectural drift from code changes.
- **Issues:** None procedural.

### Step 5: Update dashboard
- **Action:** Updated `agent/Core/routines.md` dashboard row for Doc Freshness Check — Last Ran → 2026-08-12, Next Due → 2026-08-19, Status → done.
- **Result:** Dashboard updated.
- **Issues:** None.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Medium | `plan-grilling.md` workflow is installed in `agent/Workflows/` but has no entry in the CLAUDE.md Workflows navigation table — the Tech Lead cannot discover it via the nav table. Its trigger phrases ("grill the plan", "review plan NN", "critic pass") are also missing from Quick Triggers. | `station/CLAUDE.md` — Workflows table and Quick Triggers table | Flagged for user decision — proposed addition below |
| 2 | Medium | `critic-agent-prompts.md` skill is installed in `agent/Skills/` but has no entry in the CLAUDE.md Skills navigation table — it's consumed by `plan-grilling.md` but undiscoverable via nav. | `station/CLAUDE.md` — Skills table | Flagged for user decision — proposed addition below |
| 3 | Low | `bonsai completion` is a user-facing CLI command (shell completion setup) but `station/INDEX.md` and `station/code-index.md` count 8 commands and don't include it. | `station/INDEX.md` (Key Metrics table), `station/code-index.md` (CLI Commands table) | Flagged for user decision |

---

## Proposed Updates (for user approval — not auto-applied)

### Finding #1: Add plan-grilling to CLAUDE.md Workflows table

In `station/CLAUDE.md`, add a row to the Workflows table:

```markdown
| Adversarial pre-dispatch review of a drafted plan; Running the 6-critic grill loop before agent dispatch | [agent/Workflows/plan-grilling.md](agent/Workflows/plan-grilling.md) |
```

And add trigger phrases to the Quick Triggers table:

```markdown
| Run the plan-grilling critic loop | "grill the plan", "review plan NN", or `/plan-grilling` |
```

### Finding #2: Add critic-agent-prompts to CLAUDE.md Skills table

In `station/CLAUDE.md`, add a row to the Skills table:

```markdown
| Launching plan-grilling critic agents; The 6 critic prompt templates consumed verbatim by plan-grilling.md | [agent/Skills/critic-agent-prompts.md](agent/Skills/critic-agent-prompts.md) |
```

### Finding #3: Update CLI command count in INDEX.md

In `station/INDEX.md`, update the Key Metrics table:

```
| CLI commands | 9 (init, add, remove, list, catalog, update, guide, validate, completion) |
```

And in `station/code-index.md`, add a row to the CLI Commands table for `bonsai completion`.

---

## Errors & Warnings

No errors encountered.

---

## Items Flagged for User Review

1. **plan-grilling.md and critic-agent-prompts.md missing from CLAUDE.md navigation** — these are actively installed abilities that the Tech Lead cannot discover through the standard nav table. Recommend adding them (see proposed updates above). Low effort fix.

2. **CLI command count in INDEX.md** — minor accuracy gap, can be updated at any time.

---

## Notes for Next Run

- All navigation links are currently valid — no broken links to watch for.
- The only recent commit was a backlog-hygiene routine; no feature work introduced doc drift.
- If `plan-grilling.md` gets added to the Bonsai catalog (currently it's a custom installation from ZenGarden, per its frontmatter), the CLAUDE.md entry will already be in place.
- The 3 findings here are all documentation table gaps — safe to fix in one pass.
