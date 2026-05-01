---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-05-01
status: partial
---

# Routine Report — Doc Freshness Check

## Overview
- **Routine:** Doc Freshness Check
- **Frequency:** Every 7 days
- **Last Ran:** 2026-04-21
- **Triggered By:** loop.md autonomous dispatch

## Execution Metadata
- **Status:** partial
- **Duration:** ~8 min
- **Files Read:** 18 — `station/agent/Routines/doc-freshness-check.md`, `station/INDEX.md`, `station/agent/Core/routines.md`, `station/Playbook/Status.md`, `station/Playbook/Roadmap.md`, `station/Playbook/Backlog.md`, `station/Playbook/Standards/NoteStandards.md`, `station/Logs/RoutineLog.md`, `station/CLAUDE.md`, `station/agent/Core/identity.md`, `station/agent/Core/memory.md`, `station/agent/Protocols/memory.md`, `station/agent/Protocols/scope-boundaries.md`, `station/agent/Workflows/issue-to-implementation.md`, `station/agent/Workflows/session-wrapup.md`, `station/code-index.md`, `/home/user/Bonsai/go.mod`, `/home/user/Bonsai/.bonsai.yaml`
- **Files Modified:** 2 — `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`
- **Tools Used:** `git log --since="7 days ago"`, `find` on catalog directories, bash link resolution loop, `grep` patterns
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Scan project documentation vs recent git history
- **Action:** Retrieved git log for last 7 days (14 commits). Identified files changed in the period and cross-referenced with documentation.
- **Result:** Key changes in last 7 days: (1) `internal/wsvalidate/` package extracted (Plan 32), (2) `internal/generate/catalog_snapshot.go` added (Plan 32), (3) `internal/config/config.go` Validate() function added (Plan 32), (4) `catalog/scaffolding/Playbook/Standards/NoteStandards.md.tmpl` added + manifest updated (2026-04-25 feat commit), (5) NoteStandards wired into session-logging, memory protocol, session workflows. Four of these are not yet reflected in `station/code-index.md`.
- **Issues:** `code-index.md` does not document `internal/wsvalidate/`, `internal/generate/catalog_snapshot.go`, or the new `Validate()` function on `ProjectConfig`.

### Step 2: Check INDEX.md accuracy
- **Action:** Read `station/INDEX.md` and compared tech stack, metrics, and project description against codebase reality.
- **Result:** Found two stale entries:
  1. **Go version** — INDEX.md says `Go 1.24+`; `go.mod` shows `go 1.25.0` with `toolchain go1.25.8`. README also already says `Go 1.25+`. INDEX.md lags.
  2. **Catalog items** — INDEX.md says `~50`; actual count is 53 (skills: 18, workflows: 10, protocols: 4, sensors: 13, routines: 8). The tilde makes this defensible but approximate.
  3. Project description, tech stack entries (Cobra, Huh, LipGloss, BubbleTea, YAML, text/template), distribution method, and architecture overview all accurate.
  4. Agent types count (6) accurate. CLI commands count (7) accurate.
- **Issues:** Go version stale in INDEX.md. Catalog count approximate but still within "~50" range with 53 items.

### Step 3: Check navigation links
- **Action:** Extracted all relative links from `station/CLAUDE.md` and verified each resolved to a real file or directory. Then checked links in `agent/Core/`, `agent/Protocols/`, `agent/Workflows/`, and `agent/Skills/` files.
- **Result:**
  - All 44 links in `station/CLAUDE.md` resolve correctly — CLEAN.
  - `agent/Core/memory.md` References section has 6 broken links to `../../Research/RESEARCH-*.md` — the `Research/` directory does not exist anywhere in the repo.
  - `agent/Core/memory.md` also has one false-positive `url` link (inside a code example block) — not a real broken link.
  - `agent/Workflows/issue-to-implementation.md` references `../Skills/dispatch.md` in three places. The `dispatch` skill exists in the catalog (`catalog/skills/dispatch/`) but is NOT installed for the tech-lead agent (not in `.bonsai.yaml` skills list, not present in `station/agent/Skills/`). The workflow references a skill file that doesn't exist at the path.
  - `agent/Workflows/session-wrapup.md` has `path` and `url` as apparent link targets — these are literal placeholder text in a code example block, not broken links (false positive).
  - All `agent/Protocols/` links resolve correctly.
- **Issues:** Research/ links in memory.md are broken (stale references to files that were never created or were deleted). dispatch.md referenced by issue-to-implementation workflow is not installed.

### Step 4: Report findings
- **Action:** Compiled all findings. Per procedure, flagging for user decision — no content edits to docs.
- **Result:** 4 findings documented below. Flagging Research/ links and dispatch skill gap for user review.
- **Issues:** none

### Step 5: Update dashboard
- **Action:** Updated `station/agent/Core/routines.md` dashboard row for Doc Freshness Check.
- **Result:** Last Ran set to 2026-05-01, Next Due set to 2026-05-08, Status set to done.
- **Issues:** none

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Medium | `code-index.md` does not document new `internal/wsvalidate/` package, `catalog_snapshot.go`, or `config.Validate()` — all shipped in Plan 32 (2026-04-25) | `station/code-index.md` | Flagged for user; no edit (code-index update is non-trivial) |
| 2 | Low | Go version stale — INDEX.md says `Go 1.24+`, go.mod is `go 1.25.0` | `station/INDEX.md` (Tech Stack table) | Flagged for user |
| 3 | Medium | 6 broken `../../Research/RESEARCH-*.md` links in memory.md — Research/ directory does not exist anywhere in repo | `station/agent/Core/memory.md` (References section) | Flagged for user — either Research/ files need to be created or links need removal |
| 4 | Medium | `issue-to-implementation.md` references `../Skills/dispatch.md` in 3 places but `dispatch` skill is not installed for the tech-lead agent | `station/agent/Workflows/issue-to-implementation.md` | Flagged for user — either install dispatch skill via `bonsai add` or remove references |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

1. **Research/ directory missing** — `station/agent/Core/memory.md` References section links to 6 `RESEARCH-*.md` files (`RESEARCH-landscape-analysis.md`, `RESEARCH-concept-decisions.md`, `RESEARCH-eval-system.md`, `RESEARCH-trigger-system.md`, `RESEARCH-uiux-overhaul.md`, `RESEARCH-proof-of-bonsai-effectiveness.md`) that don't exist. Decide: (a) remove the links, (b) create placeholder stubs, or (c) if the files were deleted, scrub the references.

2. **dispatch skill not installed** — `issue-to-implementation.md` cites `agent/Skills/dispatch.md` three times (Prerequisites, Phase 7, Phase 8). The dispatch skill catalog entry exists but is not installed. Decide: (a) run `bonsai add` to install `dispatch` for the tech-lead agent, or (b) remove the three references from the workflow file.

3. **code-index.md needs update** — Plan 32 (2026-04-25) added `internal/wsvalidate/` (Normalise, InvalidReason functions), `internal/generate/catalog_snapshot.go` (CatalogSnapshot struct, WriteCatalogSnapshot function), and `internal/config/config.go` Validate() method. None are in code-index.md. Recommend adding entries; can be done as a small autonomous bundle.

4. **INDEX.md Go version** — Update `Go 1.24+` → `Go 1.25+` to match go.mod and README.

## Notes for Next Run

- Research/ link breakage in memory.md is long-standing (files were referenced but apparently never existed in the repo). If still present next run, escalate.
- dispatch skill installation gap: watch whether it gets installed or the references get removed.
- code-index.md drift: if Plan 33+ added more new packages without code-index updates, drift will compound.
- Catalog item count crossed 53 this run — consider updating INDEX.md metric from `~50` to `~55` or exact count after next major catalog addition.
