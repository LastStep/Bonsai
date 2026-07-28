---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-07-28
status: partial
---

# Routine Report — Doc Freshness Check

## Overview
- **Routine:** Doc Freshness Check
- **Frequency:** Every 7 days
- **Last Ran:** 2026-05-04 (previous value from dashboard)
- **Triggered By:** loop.md autonomous dispatch

## Execution Metadata
- **Status:** partial (audit complete — 5 findings flagged, none self-remediated per procedure)
- **Duration:** ~10 minutes
- **Files Read:** 12
  - `station/agent/Routines/doc-freshness-check.md`
  - `station/CLAUDE.md`
  - `station/INDEX.md`
  - `station/agent/Core/routines.md`
  - `station/Playbook/Status.md`
  - `station/Playbook/Roadmap.md`
  - `station/agent/Core/memory.md`
  - `station/agent/Workflows/issue-to-implementation.md`
  - `station/agent/Skills/critic-agent-prompts.md`
  - `station/agent/Workflows/plan-grilling.md`
  - `station/Logs/RoutineLog.md`
  - `station/agent/Core/self-awareness.md` (partial check)
- **Files Modified:** 2
  - `station/agent/Core/routines.md` (dashboard update)
  - `station/Logs/RoutineLog.md` (log entry)
- **Tools Used:** Read, Bash (git log, file existence checks, link extraction), Write, Edit
- **Errors Encountered:** 0

---

## Procedure Walkthrough

### Step 1: Scan project documentation vs recent git history

- **Action:** Retrieved last 7 days of git commits via `git log --since="7 days ago" --stat`; read INDEX.md, Status.md, Roadmap.md, and Backlog.md.
- **Result:** Only 2 commits in the last 7 days — both routine maintenance (status-hygiene and backlog-hygiene), touching only `station/` scaffolding files. No source code, catalog, or architecture changes. No new features, commands, or config additions that would create doc drift in INDEX.md or architecture docs.
- **Issues:** None from git history. All recent changes are station-internal (routine log entries, Status.md archival).

### Step 2: Check INDEX.md accuracy

- **Action:** Verified tech stack table, agent/catalog counts, and CLI command list against actual codebase.
- **Result:**
  - Tech stack (Go 1.25+, Cobra, Huh, LipGloss, BubbleTea, YAML, text/template, embed.FS): all accurate.
  - Agent types: says 6 — confirmed (backend, devops, frontend, fullstack, security, tech-lead).
  - Catalog items: says ~50 — actual is 53 non-agent items (18 skills + 10 workflows + 4 protocols + 13 sensors + 8 routines). Approximation is acceptable.
  - CLI commands: says 8 (init, add, remove, list, catalog, update, guide, validate) — `cmd/completion.go` provides a 9th command `bonsai completion` not listed. Minor inaccuracy (see Finding 5).
- **Issues:** One minor count inaccuracy (CLI commands: 8 vs 9).

### Step 3: Check navigation links

- **Action:** Extracted all markdown links from `station/CLAUDE.md`, `agent/Core/*.md`, `agent/Protocols/*.md`, `agent/Workflows/*.md`, and `agent/Skills/*.md`. Verified each link resolves relative to its source file.
- **Result:**
  - All 49 links in `station/CLAUDE.md` resolve correctly (including external references to .bonsai.yaml, catalog.json, Playbook/, Logs/, agent/ subtrees).
  - `agent/Core/memory.md`: 6 links in the References section point to `../../Research/RESEARCH-*.md`. The `Research/` directory does not exist at the repo root or anywhere in the workspace (see Finding 2).
  - `agent/Workflows/issue-to-implementation.md`: References `[agent/Skills/dispatch.md](../Skills/dispatch.md)` in 3 places (Prerequisites, Phase 7 Triage, Phase 8 Execute). File does not exist (see Finding 1).
  - `agent/Workflows/plan-grilling.md`: File exists at `agent/Workflows/` but is NOT listed in the CLAUDE.md Workflows navigation table (see Finding 3).
  - `agent/Skills/critic-agent-prompts.md`: File exists at `agent/Skills/` but is NOT listed in the CLAUDE.md Skills navigation table (see Finding 4).
  - All other Core, Protocol, Workflow, and Skill navigation links are valid.
- **Issues:** 4 link/navigation findings (see Findings 1–4).

### Step 4: Report findings

- **Action:** Compiled findings table (below). Per procedure, findings are flagged for user decision — no edits made to the affected files.
- **Result:** 5 findings logged (1 high, 1 medium, 2 low, 1 info).
- **Issues:** None.

### Step 5: Update dashboard

- **Action:** Updated `agent/Core/routines.md` dashboard — set `Last Ran` to 2026-07-28, `Next Due` to 2026-08-04, `Status` to `done` for Doc Freshness Check row.
- **Result:** Dashboard updated.
- **Issues:** None.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | **High** | `dispatch.md` skill file referenced 3× in issue-to-implementation workflow does not exist | `agent/Workflows/issue-to-implementation.md` lines 36, 175, 204 | Flagged — not self-remediated (user decision required) |
| 2 | **Medium** | 6 links in memory.md References section point to `Research/` directory that does not exist anywhere in the repo | `agent/Core/memory.md` lines 87–92 | Flagged — not self-remediated |
| 3 | **Low** | `plan-grilling.md` workflow exists in `agent/Workflows/` but is absent from CLAUDE.md Workflows navigation table | `station/CLAUDE.md` Workflows section | Flagged — not self-remediated |
| 4 | **Low** | `critic-agent-prompts.md` skill exists in `agent/Skills/` but is absent from CLAUDE.md Skills navigation table | `station/CLAUDE.md` Skills section | Flagged — not self-remediated |
| 5 | **Info** | INDEX.md states "8" CLI commands; `cmd/completion.go` provides a 9th (`bonsai completion`) not listed | `station/INDEX.md` Key Metrics table | Flagged — minor accuracy item |

---

## Errors & Warnings

No errors encountered during execution.

---

## Items Flagged for User Review

- **[HIGH] Create or restore `dispatch.md`** — `agent/Workflows/issue-to-implementation.md` references `agent/Skills/dispatch.md` in 3 critical places (Prerequisites, Triage, Execute sections). The file is missing. Options: (a) create the file with dispatch triage rules and agent prompt structure, or (b) if this content was folded into another skill, update the references to point there (e.g., `issue-classification.md` or `planning-template.md`). This breaks a primary workflow.

- **[MEDIUM] Resolve Research/ links in memory.md** — The References section links to 6 research documents (`RESEARCH-landscape-analysis.md`, `RESEARCH-concept-decisions.md`, `RESEARCH-eval-system.md`, `RESEARCH-trigger-system.md`, `RESEARCH-uiux-overhaul.md`, `RESEARCH-proof-of-bonsai-effectiveness.md`) under a non-existent `Research/` directory. Options: (a) create the `Research/` directory at the repo root and add these docs, (b) update links if the documents live elsewhere, or (c) remove the References section if these are truly gone.

- **[LOW] Add `plan-grilling.md` to CLAUDE.md Workflows table** — The plan-grilling workflow (adversarial plan review via 6 critic agents) is not in the navigation table. If the user intends to use it, adding a row to the Workflows section would make it discoverable. Suggested activate-when: "Running adversarial review of a drafted plan before dispatch."

- **[LOW] Add or note `critic-agent-prompts.md` in CLAUDE.md** — Supporting skill for `plan-grilling.md`. May intentionally be omitted from the navigation table (consumed verbatim by plan-grilling, not a standalone skill). No action needed if this is deliberate.

- **[INFO] Update INDEX.md CLI command count** — Change "8 (init, add, remove, list, catalog, update, guide, validate)" to "9 (init, add, remove, list, catalog, update, guide, validate, completion)" if completeness matters. Low priority.

---

## Notes for Next Run

- No code changes shipped in the 7-day window — doc freshness was clean on the architecture side. Primary findings were workspace navigation gaps (missing skill file, broken research links).
- Plan 41 file remains in `Plans/Active/` per memory.md note ("archive to Plans/Archive/ at next wrap-up"). Not a doc-freshness issue but worth noting.
- If `dispatch.md` is created before the next run, verify the 3 links in `issue-to-implementation.md` resolve correctly.
- The Research/ directory links in memory.md are long-standing (pre-existing before this run's 7-day window). If not resolved, they will appear again next run.
