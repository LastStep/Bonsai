---
tags: [routine]
description: Detect documentation drift — docs that are stale relative to recent code changes.
frequency: 7 days
---

# Routine: Doc Freshness Check

**Frequency:** Every 7 days

## Purpose

Detect documentation drift — docs that are stale relative to recent code changes. Ensures architecture docs, project files, and navigation links stay accurate.

## Procedure

1. **Scan project documentation:**
   - Read docs in `station/` (INDEX.md, Playbook/, Architecture/ if present)
   - Compare against recent git history (last 7 days of commits)
   - Flag any new features, services, or config not reflected in docs

2. **Check INDEX.md accuracy:**
   - Verify tech stack, folder structure, and project description still match reality

3. **Check navigation links:**
   - Verify all links in `station/CLAUDE.md` navigation tables resolve to real files
   - Verify all links in `agent/Core/`, `agent/Protocols/`, `agent/Workflows/`, `agent/Skills/` resolve

4. **Report findings:**
   - List stale docs with specific drift description
   - Propose updates (but don't execute — flag for user decision)
   - Log to `station/Logs/RoutineLog.md`

5. **Update dashboard** — set `last_ran` to today's date in `agent/Core/routines.md`
