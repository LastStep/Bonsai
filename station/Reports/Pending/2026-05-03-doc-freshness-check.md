---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-05-03
status: success
---

# Routine Report — Doc Freshness Check

## Overview
- **Routine:** Doc Freshness Check
- **Frequency:** Every 7 days
- **Last Ran:** 2026-04-21 (before this run)
- **Triggered By:** loop.md autonomous dispatch

## Execution Metadata
- **Status:** success
- **Duration:** ~8 minutes
- **Files Read:** 18 — `/home/user/Bonsai/station/agent/Routines/doc-freshness-check.md`, `/home/user/Bonsai/station/INDEX.md`, `/home/user/Bonsai/station/CLAUDE.md`, `/home/user/Bonsai/station/Playbook/Status.md`, `/home/user/Bonsai/station/Playbook/Roadmap.md`, `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/agent/Core/memory.md`, `/home/user/Bonsai/station/agent/Core/self-awareness.md`, `/home/user/Bonsai/station/code-index.md`, `/home/user/Bonsai/.bonsai.yaml`, `/home/user/Bonsai/station/agent/Workflows/issue-to-implementation.md`, `/home/user/Bonsai/station/agent/Workflows/session-wrapup.md`, `/home/user/Bonsai/station/agent/Sensors/statusline.sh`, `/home/user/Bonsai/station/agent/Sensors/status-bar.sh`, `/home/user/Bonsai/.claude/settings.json`, `/home/user/Bonsai/station/.claude/settings.json`, `/home/user/Bonsai/internal/wsvalidate/wsvalidate.go`, `/home/user/Bonsai/station/Playbook/Standards/NoteStandards.md`
- **Files Modified:** 2 — `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`
- **Tools Used:** `git log`, `ls`, `find`, `grep`, `python3 -m json.tool`, bash link-checker script
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Scan project documentation against recent git history
- **Action:** Ran `git log --oneline -20 --name-only --format="%h %ad %s"` to review commits since last run (2026-04-21). Read INDEX.md, Status.md, Roadmap.md.
- **Result:** Only 1 commit landed since last doc-freshness run (2026-04-26 to 2026-05-03): `c5ea838` — backlog-hygiene routine run. Prior week (2026-04-22–2026-04-25) had significant commits including Plan 32 (wsvalidate package extraction) and Plan 33 (website rewrite). These predate the last doc-freshness check window but the wsvalidate package appears undocumented.
- **Issues:** Found one documentation drift: `internal/wsvalidate/` package was introduced in Plan 32 (2026-04-25) and is not reflected in INDEX.md architecture overview, code-index.md, or root CLAUDE.md project structure.

### Step 2: Check INDEX.md accuracy
- **Action:** Read INDEX.md and compared tech stack, folder structure, and project description against current codebase. Verified CLI command count, agent type count, and catalog item count.
- **Result:** INDEX.md is largely accurate:
  - Tech stack: correct (Go 1.24+, Cobra, Huh, LipGloss, BubbleTea, YAML, text/template, embed.FS)
  - CLI commands: "7 (init, add, remove, list, catalog, update, guide)" — verified, correct
  - Agent types: "6 (tech-lead, fullstack, backend, frontend, devops, security)" — verified, correct
  - Catalog items: "~50" — actual count is 53 (18 skills + 10 workflows + 4 protocols + 13 sensors + 8 routines), approximate is acceptable
  - Architecture diagram: lists `catalog`, `config`, `generate`, `tui` as internal packages but misses `wsvalidate` — drift confirmed
- **Issues:** `internal/wsvalidate` package not reflected in INDEX.md architecture section.

### Step 3: Check navigation links
- **Action:** Ran bash link-checker against all relative links in `station/CLAUDE.md`, `station/INDEX.md`, `agent/Core/*.md`, `agent/Protocols/*.md`, `agent/Workflows/*.md`, and `agent/Skills/*.md`.
- **Result:**
  - `station/CLAUDE.md`: All 44 relative links resolve — OK
  - `station/INDEX.md`: All links resolve — OK
  - `agent/Core/identity.md`: No relative links — OK
  - `agent/Core/self-awareness.md`: Link to `memory.md` resolves — OK
  - `agent/Core/memory.md`: 6 broken links to `../../Research/RESEARCH-*.md` files (directory `station/Research/` does not exist). "url" flagged by checker is a false positive (inline code example, not a link).
  - `agent/Protocols/*.md`: All links resolve — OK
  - `agent/Workflows/issue-to-implementation.md`: 3 broken links to `../Skills/dispatch.md` — file not installed in `station/agent/Skills/` (skill exists in catalog but not in `.bonsai.yaml` installed skills for tech-lead)
  - `agent/Workflows/session-wrapup.md`: "path" and "url" flagged as broken links are false positives — they appear inside inline code in a brevity-rule note
  - `agent/Skills/*.md`: All links resolve — OK
- **Issues:** 2 real broken-link issues found (Research files missing, dispatch.md not installed).

### Step 4: Report findings
- **Action:** Compiled all findings into this report.
- **Result:** 3 findings requiring user review (2 broken-link issues + 1 documentation drift). No autonomous fixes executed — all flagged for user decision per procedure.
- **Issues:** none

### Step 5: Update dashboard
- **Action:** Updated `agent/Core/routines.md` dashboard row for "Doc Freshness Check" — Last Ran → 2026-05-03, Next Due → 2026-05-10, Status → done.
- **Result:** Dashboard updated.
- **Issues:** none

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Medium | 6 broken links to `station/Research/RESEARCH-*.md` files that do not exist — directory never created | `station/agent/Core/memory.md` Lines 78–83 (References section) | Flagged for user decision — options: (a) create the Research directory and files, (b) remove stale references, (c) update links to wherever the research content lives |
| 2 | Low | 3 broken links to `agent/Skills/dispatch.md` — dispatch skill exists in catalog but not installed for tech-lead in `.bonsai.yaml` | `station/agent/Workflows/issue-to-implementation.md` Lines 35, 175, 204 | Flagged for user decision — options: (a) `bonsai add` to install dispatch skill, (b) inline the dispatch guidance into the workflow, (c) remove the cross-links |
| 3 | Low | `internal/wsvalidate` package (added Plan 32, 2026-04-25) not documented in INDEX.md architecture overview, code-index.md, or root CLAUDE.md | `station/INDEX.md` architecture section; `station/code-index.md`; root `CLAUDE.md` | Flagged for user decision — straightforward doc update, can be added to backlog or done inline |

---

## Errors & Warnings

No errors encountered.

---

## Items Flagged for User Review

**Finding 1 — Research file references in memory.md**
The References section of `station/agent/Core/memory.md` (lines 78–83) contains 6 links to research documents in `station/Research/` that do not exist. Git history shows the directory was never created. These appear to be intended future documents or references to work that hasn't been produced yet. Options:
- Remove the broken references if research was never done
- Create `station/Research/` and stub the files if the research is planned
- Update links if the content moved elsewhere

**Finding 2 — dispatch skill not installed**
`station/agent/Workflows/issue-to-implementation.md` links to `agent/Skills/dispatch.md` in 3 places. The dispatch skill is available in the catalog (`catalog/skills/dispatch/`) and targets `tech-lead` agents, but is not installed in `.bonsai.yaml`. Running `bonsai add` and selecting the dispatch skill for the tech-lead agent would install it and resolve these broken links.

**Finding 3 — wsvalidate undocumented**
`internal/wsvalidate/` is a real package shipped in Plan 32 but not mentioned in any documentation. It centralises workspace-path validation (Normalise, InvalidReason functions). Low urgency — add to INDEX.md architecture table and code-index.md at next opportunity.

---

## Notes for Next Run

- `station/Research/` references have been present since at least `1b49963` (v0.2.0 bookkeeping) — they are clearly long-lived open items, not accidental drift. Include in next run to check if resolved.
- The session-wrapup.md and memory.md `[path]`/`[url]`/`url` false positives: the link-checker extracting text from markdown inline code is a known limitation. Future runs can pre-filter with `grep -v '^\`'`.
- statusline.sh in `station/agent/Sensors/` is intentionally unlisted in the CLAUDE.md Sensors nav table — it's a `statusLine` renderer registered in `station/.claude/settings.json`, not a hook sensor. This is correct.
