---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-06-26
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
- **Duration:** ~10 minutes
- **Files Read:** 9 — `station/agent/Routines/doc-freshness-check.md`, `station/INDEX.md`, `station/CLAUDE.md`, `station/Playbook/Status.md`, `station/Playbook/Roadmap.md`, `station/agent/Core/routines.md`, `station/agent/Core/memory.md`, `station/code-index.md`, `station/Logs/RoutineLog.md`
- **Files Modified:** 4 — `station/INDEX.md`, `station/code-index.md`, `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`
- **Tools Used:** Read, Edit, Write, Bash (git log, file existence checks, grep)
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1 — Scan project documentation against recent git history

Ran `git log --since="2026-05-04"` (since last doc freshness run). Identified 40+ commits across two major plans:

- **Plan 40 (Odysseus, v0.5.0)** — Phases 1–3 merged (PRs #114/#115/#116): frozen v1 schemas, root-relative scaffolding (`catalog/scaffolding/.bonsai/project.yaml.tmpl`, `MEMORY.md.tmpl`), project-level `validate` pass, memory-routing protocol + guide Formats page.
- **Plan 41 (Headless CLI Contract)** — all 5 phases merged (PRs #120–#125): new `internal/nonint/` package, `internal/generate/list_snapshot.go`, `docs/agent-interface.md` contract doc, `list --json`, exit codes 0/2/3/4/5.

Also noted: external contribution `bonsai completion` subcommand (PR #78, `2eae9d4`), Dependabot bumps (codeql-action v3→v4, checkout v4→v6, etc.).

### Step 2 — Check INDEX.md accuracy

- Tech stack: accurate (Go 1.25+, Cobra, Huh, LipGloss, BubbleTea, yaml.v3, text/template, embed.FS).
- CLI command count: accurate (8 commands).
- Architecture overview: **STALE** — `internal/nonint/` package missing. Added.
- Document Registry: **STALE** — `docs/agent-interface.md` (Plan 41 contract doc) not listed. Added.

### Step 3 — Check navigation links

Verified all links in `station/CLAUDE.md` navigation tables. Checked 49 link targets.

- **`../.bonsai/catalog.json`** — resolves to `/home/user/Bonsai/.bonsai/catalog.json`. EXISTS.
- **`../.bonsai.yaml`** — resolves to `/home/user/Bonsai/.bonsai.yaml`. EXISTS.
- All other 47 links: all resolve to existing files/directories.
- **Result: 0 broken links.**

Verified links in `station/agent/Core/` (memory, identity, self-awareness, routines), `station/agent/Protocols/` (4 files), `station/agent/Workflows/` (8 files), `station/agent/Skills/` (6 files). All present.

### Step 4 — Report findings

Found 3 doc drift items. Two resolved inline (clear factual additions, not judgment calls). One flagged for user.

**Inline fixes applied:**
1. `station/INDEX.md`: added `internal/nonint/` line to architecture diagram.
2. `station/INDEX.md`: added `docs/agent-interface.md` row to Document Registry.
3. `station/code-index.md`: added `list_snapshot.go` section and full `internal/nonint/` package section (Plan 41).

**Flagged for user:**
- Plan 41 file at `Plans/Active/41-headless-cli-contract.md` remains in Active/ — memory.md noted it should be archived but it was not actioned. Out of scope for this routine; flagged below.

### Step 5 — Dashboard update

Updated `station/agent/Core/routines.md`: Doc Freshness Check row → Last Ran `2026-06-26`, Next Due `2026-07-03`, Status `done`.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | medium | `internal/nonint/` package (Plan 41, shipped 2026-06-16) missing from architecture overview | `station/INDEX.md` architecture diagram | Fixed — added `internal/nonint/` line |
| 2 | low | `docs/agent-interface.md` headless CLI contract doc not in Document Registry | `station/INDEX.md` Document Registry table | Fixed — added row |
| 3 | medium | `internal/nonint/` package and `list_snapshot.go` not documented in code index | `station/code-index.md` | Fixed — added `list_snapshot.go` section + `internal/nonint/` package section |
| 4 | low | Plan 41 file still in `Plans/Active/` (should be archived per memory.md) | `station/Playbook/Plans/Active/41-headless-cli-contract.md` | Flagged — not archived; out of scope for this routine |

## Errors & Warnings

No errors encountered.

Note: Initial navigation link check misread relative paths (`../.bonsai/*`) as station-relative, but both links are valid — they correctly point one level up to the project root where `.bonsai.yaml` and `.bonsai/catalog.json` live.

## Items Flagged for User Review

**Plan 41 archive:** `station/Playbook/Plans/Active/41-headless-cli-contract.md` is stale in Active/ — memory.md notes it should be moved to `Plans/Archive/`. Can be done in the next wrap-up session or on-demand. No blocking impact.

## Notes for Next Run

- Plan 41 is shipped and `internal/nonint/` + `list_snapshot.go` are now documented. No further drift expected unless Plan 42 (MCP server) ships.
- The `docs/` directory now holds the agent-interface contract (`agent-interface.md`) and user-facing docs (`cli.md`, `concepts.md`, `custom-files.md`, `formats.md`, `quickstart.md`). If Plan 42 (MCP server) ships, expect `docs/` additions — check at next run.
- Consider adding a root-CLAUDE.md check sub-step to this routine (already in Backlog P3 per 2026-05-04 digest). The root `CLAUDE.md` project structure section still lacks `internal/nonint/` — deferred since it's developer-facing and updated less frequently.
