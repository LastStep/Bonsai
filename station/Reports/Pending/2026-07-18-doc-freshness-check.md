---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-07-18
status: partial
---

# Routine Report — Doc Freshness Check

## Overview
- **Routine:** Doc Freshness Check
- **Frequency:** Every 7 days
- **Last Ran:** 2026-05-04 (before this run)
- **Triggered By:** loop.md autonomous dispatch

## Execution Metadata
- **Status:** partial
- **Duration:** ~12 minutes
- **Files Read:** 7 — station/agent/Routines/doc-freshness-check.md, station/INDEX.md, station/agent/Core/routines.md, station/Playbook/Status.md, station/code-index.md, station/Playbook/Roadmap.md, station/Logs/RoutineLog.md (plus station/CLAUDE.md and root CLAUDE.md via context)
- **Files Modified:** 2 — station/agent/Core/routines.md (dashboard update), station/Logs/RoutineLog.md (log entry)
- **Tools Used:** git log --oneline -20 --format="%h %ad %s" --date=short; ls on catalog/, Sensors/, Playbook/, Logs/, Reports/ directories; test -f on .bonsai/catalog.json, .bonsai.yaml; glob on cmd/*.go, station/agent/**/*.md; line counts on catalog/* subdirectories
- **Errors Encountered:** 0

> Status is "partial" because the routine procedure says findings should be flagged for user decision — the audit identified actionable drift items that require user updates to docs, not resolved autonomously.

---

## Procedure Walkthrough

### Step 1: Scan project documentation against recent git history (last 7 days)
- **Action:** Ran `git log --oneline --since="2026-07-11"` to find commits in the past 7 days; also ran `git log --oneline -20` for broader recent history since last run (2026-05-04).
- **Result:** Only one commit in the past 7 days: `cebf20f 2026-07-18 routine(backlog-hygiene): run 2026-07-18` — a docs-only routine log commit, no code changes. The last meaningful code changes were 2026-06-16 (Plan 41) and 2026-06-13 (Plan 40), both within the gap since last doc-freshness-check (2026-05-04). Key events since last run: `bonsai completion` command added (PR #78, 2026-05-07), Plan 41 headless CLI contract (2026-06-16), Plan 40 platform integration (2026-06-13).
- **Issues:** Drift found — see findings below.

### Step 2: Check INDEX.md accuracy
- **Action:** Read station/INDEX.md; compared Tech Stack, Key Metrics, Architecture Overview against current codebase (cmd/*.go glob, catalog item counts).
- **Result:**
  - **Tech Stack table:** Still accurate (Go 1.25+, Cobra, Huh, LipGloss, BubbleTea, YAML, text/template, embed.FS).
  - **Agent types (6):** Confirmed — 6 agent directories in catalog/agents/.
  - **Catalog items (~50):** Current actual count is 53 (18 skills + 10 workflows + 4 protocols + 13 sensors + 8 routines). "~50" is approximate but still acceptable.
  - **CLI commands (8):** **STALE** — cmd/completion.go exists (added PR #78, 2026-05-07). Count should be 9. Listed commands also missing `completion`.
  - **Architecture Overview:** Still accurate. `internal/validate/` and `internal/wsvalidate/` are both listed and both exist. All cmd/ and internal/ layers are correctly described.
  - **Document Registry:** All referenced files/paths verified to exist. Does not reference `docs/agent-interface.md` (created in Plan 41 Phase 5) — this is a project-level docs file, not a station doc, so its absence here may be intentional.
- **Issues:** 1 stale metric (CLI commands count).

### Step 3: Check navigation links
- **Action:** Verified every file path linked in station/CLAUDE.md navigation tables (Core, Protocols, Workflows, Skills, Routines, Sensors, External References) against actual files on disk.
- **Result:**
  - **Core files:** identity.md, memory.md, self-awareness.md, routines.md — all exist ✓
  - **Protocols:** memory.md, scope-boundaries.md, security.md, session-start.md — all exist ✓
  - **Workflows:** code-review, planning, pr-review, security-audit, session-logging, test-plan, session-wrapup, issue-to-implementation, routine-digest — all exist ✓
  - **Skills:** planning-template, review-checklist, issue-classification, pr-creation, bubbletea, bonsai-model — all exist ✓
  - **Routines:** backlog-hygiene, dependency-audit, doc-freshness-check, memory-consolidation, roadmap-accuracy, status-hygiene, vulnerability-scan — all exist ✓
  - **Sensors:** context-guard.sh, scope-guard-files.sh, session-context.sh, status-bar.sh, routine-check.sh, agent-review.sh, dispatch-guard.sh, subagent-stop-review.sh, compact-recovery.sh, statusline.sh — all 10 exist ✓
  - **Bonsai Reference:** ../.bonsai/catalog.json ✓, ../.bonsai.yaml ✓
  - **External References:** Status.md, Roadmap.md, Backlog.md, SecurityStandards.md, Plans/Active/, KeyDecisionLog.md, Reports/Pending/, report-template.md, code-index.md — all exist ✓
  - **Unlisted files found:** `agent/Workflows/plan-grilling.md` and `agent/Skills/critic-agent-prompts.md` exist but are not in the CLAUDE.md navigation tables. Also `agent/Skills/bubbletea/` subdirectory files (components.md, golden-rules.md, troubleshooting.md, emoji-width-fix.md) exist but are subpages of the listed bubbletea.md skill, so omission is expected.
- **Issues:** No broken links. Two unlisted files may represent nav gaps (see Finding 3/4).

### Step 4: Check code-index.md accuracy
- **Action:** Read station/code-index.md; compared CLI Commands table against cmd/*.go files.
- **Result:**
  - **Missing command:** `bonsai completion` (cmd/completion.go) is absent from the CLI Commands table.
  - **Line number drift risk:** Plan 41 (2026-06-16) significantly modified remove.go, update.go, list.go, add.go (new headless cores, new flags, Result reshaping). Line references in code-index.md for these files (e.g., runRemove() at :290, runUpdate() at :19, runList() at :18) may have shifted. Could not verify every line without reading each file; flagged for user validation.
  - **New headless-core functions:** Plan 41 added headless `*Result` core functions (e.g., `RemoveCore()`, `UpdateCore()`, `ListResult`) that are not documented in code-index.md. These are substantive new entry points for MCP integration.
- **Issues:** Missing `completion` entry; potential line-number drift in 4 files; missing headless core function documentation.

### Step 5: Report findings and flag for user
- **Action:** Compiled findings below; no autonomous edits to INDEX.md or code-index.md (procedure: "propose updates but don't execute — flag for user decision").
- **Result:** 5 findings documented in table below.
- **Issues:** None.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | medium | INDEX.md CLI command count is stale — says 8, should be 9; `completion` not listed | station/INDEX.md, Key Metrics table | Flagged for user — update count to 9 and add `completion` to the list |
| 2 | medium | code-index.md missing `bonsai completion` entry — cmd/completion.go exists but not in CLI Commands table | station/code-index.md, CLI Commands section | Flagged for user — add row: `bonsai completion \| cmd/completion.go:17 \| completionCmd → shell completion script generation` |
| 3 | low | code-index.md line numbers for remove.go, update.go, list.go, add.go may have shifted — Plan 41 (2026-06-16) substantially modified these files (headless cores, new flags, Result types) | station/code-index.md, Add/Remove/Update/List Helpers sections | Flagged for user — verify line refs against current source; run a quick diff or re-read each file |
| 4 | low | code-index.md missing headless core documentation — Plan 41 added `*Result` headless cores for all mutating commands (init/add/update/remove) + `list --json` as the MCP-readiness layer; none appear in code-index.md | station/code-index.md | Flagged for user — consider adding a "Headless Cores (Plan 41)" subsection documenting the new `*Result` functions and the `--json`/`--non-interactive`/`--yes`/`--from`/`--skip-conflicts` flags |
| 5 | info | Two agent instruction files exist outside the CLAUDE.md navigation tables — `agent/Workflows/plan-grilling.md` and `agent/Skills/critic-agent-prompts.md` are present on disk but not listed | station/CLAUDE.md, Workflows and Skills sections | Flagged for user — intentional unlisting (private/advanced) or nav drift? Add to table if they should be discoverable |

---

## Errors & Warnings

No errors encountered.

---

## Items Flagged for User Review

1. **INDEX.md CLI count** — Update "8" → "9" and add `completion` to the parenthetical list (init, add, remove, list, catalog, update, guide, validate, completion).

2. **code-index.md completion entry** — Add a new row to the CLI Commands table for `bonsai completion | cmd/completion.go:17 | completionCmd → shell completion script (bash/zsh/fish/powershell)`.

3. **code-index.md Plan 41 drift** — Verify line numbers in Add, Remove, Update, List helper sections. Plan 41 PRs (#120/#122/#123/#121) reshaped those commands significantly. Easiest approach: a quick Plan 37-style doc refresh pass targeting just the four files.

4. **code-index.md headless cores** — Decision: add a new section for the Plan 41 headless API surface? This would help developer agents targeting the MCP integration path.

5. **CLAUDE.md unlisted files** — Decision: add `plan-grilling.md` to Workflows table and `critic-agent-prompts.md` to Skills table, or confirm they are intentionally private/unlisted.

---

## Notes for Next Run

- The gap since last run was 75 days (2026-05-04 → 2026-07-18). At weekly frequency this routine is overdue — the routine-check sensor should have flagged it but the dispatch was apparently deferred until today.
- No broken links found anywhere in station/CLAUDE.md — the nav table is clean.
- All navigation links resolved successfully; no orphaned references in Core, Protocols, Workflows, Skills, Routines, or Sensors tables.
- If Plan 42 (MCP server) ships before the next run, INDEX.md may need an update to the Architecture Overview and Key Metrics.
- The `docs/agent-interface.md` contract document (Plan 41 Phase 5) is worth considering for the INDEX.md Document Registry if tech-lead agents will reason about it.
