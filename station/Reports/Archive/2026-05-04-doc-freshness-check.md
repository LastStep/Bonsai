---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Doc Freshness Check"
date: 2026-05-04
status: success
---

# Routine Report — Doc Freshness Check

## Overview
- **Routine:** Doc Freshness Check
- **Frequency:** Every 7 days
- **Last Ran:** 2026-04-21
- **Triggered By:** TL maintenance dispatch

## Execution Metadata
- **Status:** success
- **Duration:** ~7 min
- **Files Read:** 14
  - `station/agent/Routines/doc-freshness-check.md`
  - `station/Reports/Archive/2026-04-21-doc-freshness-check.md`
  - `station/INDEX.md`
  - `station/CLAUDE.md`
  - `Bonsai/CLAUDE.md` (project root)
  - `station/Playbook/Status.md`
  - `station/Playbook/Roadmap.md`
  - `station/Playbook/Backlog.md`
  - `station/code-index.md`
  - `station/agent/Core/routines.md`
  - `.bonsai.yaml`
  - `catalog/skills/bonsai-model/meta.yaml`
  - Directory listings for `cmd/`, `catalog/`, `internal/`, `internal/tui/`, `station/agent/{Skills,Routines,Sensors,Workflows,Protocols,Core}/`
  - `station/Playbook/Plans/{Active,Archive}/`
- **Files Modified:** 0 — audit-only; findings flagged for user review, no autonomous edits
- **Tools Used:** Read, Bash (ls + git log + grep)
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Scan project documentation against recent commits
- **Action:** Read all listed docs. Ran `git log --since='2026-04-21' --oneline` (43 commits in window) and `--stat` for files touched. Cross-referenced with archived plans 30/31/32/33/34/35.
- **Result:** Recent shipped work in window: Plan 30 (guide perf + view polish, PR #74), Plan 31 (v0.3 release readiness — 3 PRs #75/#76/#77; introduced `internal/tui/{removeflow,updateflow,hints}/`, `internal/generate/catalog_snapshot.go`), Plan 32 (followup bundle — extracted `internal/wsvalidate/`, PR #80), Plan 33 (website rewrite, PR #79), Plan 34 (custom-ability discovery bug bundle, PR #92), Plan 35 (`bonsai validate` command — added `cmd/validate.go` + `internal/validate/`, PR #93). NoteStandards skill added to catalog scaffolding. Plus `cmd/init_flow.go` separated from `cmd/init.go`.
- **Issues:** 5 drift items found (see Findings).

### Step 2: Check INDEX.md accuracy
- **Action:** Compared `station/INDEX.md` Tech Stack, Key Metrics, Architecture Overview, Document Registry against actual codebase.
- **Result:**
  - **Key Metrics — CLI commands stale.** Says "7 (init, add, remove, list, catalog, update, guide)" but `cmd/` now contains 8 user-facing commands — `validate` shipped 2026-05-04 via Plan 35.
  - **Key Metrics — Catalog items "~50".** Actual count: skills 18 + workflows 10 + protocols 4 + sensors 13 + routines 8 = 53. Still in the right ballpark; no change needed.
  - **Architecture Overview ASCII diagram still says** `internal/tui/   ← Huh forms + LipGloss styled output` — same drift as 2026-04-21 (BubbleTea + harness/ + 7 cinematic flow packages not surfaced). Now further out of date: `internal/validate/` and `internal/wsvalidate/` are entirely missing from the `internal/` tree shown in the diagram.
  - Tech Stack accurate. Agent-types count (6) accurate.
- **Issues:** 2 drift items (findings #1, #2).

### Step 3: Check navigation links in station/CLAUDE.md
- **Action:** Programmatically extracted all 47 file links from `station/CLAUDE.md` and verified each resolves on disk.
- **Result:** **46/47 resolve. 1 broken link.**
  - `[agent/Skills/bonsai-model.md](agent/Skills/bonsai-model.md)` (Bonsai Reference table, line 31) → file does NOT exist in `station/agent/Skills/`. The skill exists in catalog (`catalog/skills/bonsai-model/bonsai-model.md`) but `bonsai-model` is **not in `.bonsai.yaml` tech-lead skills list**, so it was never installed to the workspace. The link was added by commit `8e21b75` (frontmatter backfill) when the Bonsai Reference table was introduced.
- **Issues:** 1 drift item (finding #3).

### Step 4: Check navigation table coverage vs actual files
- **Action:** Cross-referenced `station/CLAUDE.md` Skills/Routines/Sensors tables against `ls station/agent/{Skills,Routines,Sensors}/`.
- **Result:**
  - **Skills table** lists 5 (planning-template, review-checklist, issue-classification, pr-creation, bubbletea) — matches 5 files on disk. Last run flagged this as a gap; `bubbletea` is now wired into the table (commit `8e21b75`). **Resolved.**
  - **Routines table** — 7 entries match 7 files exactly.
  - **Sensors table** — 10 entries match 10 files exactly. (Last run dashboard table-split issue is also resolved — `agent/Core/routines.md` now renders as one table.)
  - `bubbletea/` subdirectory (4 topic files: components.md, emoji-width-fix.md, golden-rules.md, troubleshooting.md) is referenced inside `bubbletea.md` but not surfaced in the nav table — same as 2026-04-21, still a user decision; not re-flagging.
- **Issues:** none new (one item from prior run resolved).

### Step 5: Check root Bonsai/CLAUDE.md project-structure tree
- **Action:** Compared the Project Structure ASCII tree in `Bonsai/CLAUDE.md:14-79` against actual repo layout.
- **Result:** Significant drift since last run, partly improved but still incomplete:
  - `cmd/` block lists 8 files but missing `validate.go` (Plan 35) and `init_flow.go` (Plan 22 split). Actual: `add.go, catalog.go, guide.go, init.go, init_flow.go, list.go, remove.go, root.go, update.go, validate.go` — 10 user-facing files.
  - `internal/` listing **missing entire `internal/validate/` directory** (Plan 35: validate.go + validate_test.go, 1216 LOC) and **`internal/wsvalidate/` directory** (Plan 32 extract). Both are first-class packages.
  - `internal/tui/` block: still only lists `styles.go`, `prompts.go`, `filetree.go`, `harness/`, `initflow/` — missing `addflow/`, `catalogflow/`, `guideflow/`, `hints/`, `listflow/`, `removeflow/`, `updateflow/` (7 packages totaling 50+ files). The cinematic-flow architecture established by Plans 23/27/28/30/31 is invisible in the root orientation doc.
  - `internal/generate/` listing missing `catalog_snapshot.go` + tests (Plan 31), `bonsai_reference_test.go`, `refresh_peer_awareness_test.go`.
- **Issues:** 1 drift item (finding #4) — most significant drift this run.

### Step 6: Check code-index.md accuracy
- **Action:** Spot-checked CLI Commands table line numbers with `grep -n "var .*Cmd"` against `cmd/*.go`.
- **Result:** Multiple line numbers stale due to file growth/refactors:
  - `cmd/list.go:19` → actually `:18` (off by 1)
  - `cmd/guide.go:34` → actually `:27` (off by 7)
  - `cmd/remove.go:32` → actually `:34`
  - `cmd/update.go:22` → actually `:19`
  - `cmd/catalog.go:16` → actually `:23`
  - **`bonsai validate` command not listed at all** (Plan 35 — `cmd/validate.go:23`).
  - "TUI" section (`internal/tui/`) only documents `styles.go`, `prompts.go`, `harness/`, `initflow/`, `addflow/`. Missing entire packages: `catalogflow/`, `guideflow/`, `hints/`, `listflow/`, `removeflow/`, `updateflow/` (the heart of v0.3 release).
  - "Catalog" section (`internal/catalog/`) — `New(fsys)` listed at `:220` not verified, but consider full pass after a refactor.
  - "Generator" section missing `catalog_snapshot.go` and `WriteCatalogSnapshot`.
- **Issues:** 1 drift item (finding #5).

### Step 7: Check Roadmap.md against shipped plans
- **Action:** Compared Roadmap Phase 1 + Phase 2 checkboxes against archived plans 30-35.
- **Result:**
  - **Resolved since 2026-04-21:** Phase 1 "UI overhaul" now `[x]` (was unchecked in last run flag), "Usage instructions" now `[x]`, "Release pipeline" `[x]`, "Community health files" `[x]`. Phase 2 "Custom item detection" now `[x]` — wired since Plan 34. Excellent progress.
  - **Still unchecked, still accurate:** "Better trigger sections" remains `[ ]` (Plan 08 Phase C still paused).
  - No new drift here.
- **Issues:** none new (last run's flagged unchecked boxes are now correctly checked).

### Step 8: Check for stale `.bak` files
- **Action:** `find station/agent -name '*.bak'`.
- **Result:** **Zero `.bak` files in station/agent/** — last run flagged 10+. Cleared since 2026-04-21. **Resolved.**
- **Issues:** none.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | medium | `station/INDEX.md` Key Metrics row says **CLI commands: 7** but `bonsai validate` (Plan 35, shipped 2026-05-04 via PR #93) brings the count to **8**. List should read: `init, add, remove, list, catalog, update, guide, validate`. | `station/INDEX.md:33` | flagged for user |
| 2 | low | `station/INDEX.md` Architecture Overview ASCII diagram describes `internal/tui/` as "Huh forms + LipGloss styled output" — still no mention of BubbleTea harness or the 7 cinematic flow packages (initflow, addflow, removeflow, updateflow, listflow, catalogflow, guideflow, hints). Also missing `internal/validate/` and `internal/wsvalidate/` entirely from the `internal/` block. Same drift class as 2026-04-21 #1, but compounded. | `station/INDEX.md:58-75` | flagged for user |
| 3 | medium | **Broken nav link in station/CLAUDE.md**: `Bonsai Reference` table links to `agent/Skills/bonsai-model.md` but that file does NOT exist (skill is not installed in `.bonsai.yaml` tech-lead skills list — only `bubbletea`, `planning-template`, `review-checklist`, `issue-classification`, `pr-creation`). Either `bonsai-model` should be added to the workspace via `bonsai add` (it's a tech-lead skill in the catalog) or the link removed from CLAUDE.md. | `station/CLAUDE.md:31` | flagged for user |
| 4 | high | **Root `Bonsai/CLAUDE.md` project-structure tree significantly out of date.** Three classes of drift: (a) `cmd/` block missing `init_flow.go` (Plan 22) and `validate.go` (Plan 35); (b) `internal/` block missing entire `validate/` and `wsvalidate/` packages; (c) `internal/tui/` block missing 7 cinematic flow packages introduced across Plans 23/27/28/30/31 and the `hints/` package. This is the codebase orientation doc — first thing a contributor reads — and currently misrepresents 70% of the TUI subsystem. P3 backlog item from last cycle ("Add root CLAUDE.md check to doc-freshness-check") was filed but the underlying drift was not fixed. | `Bonsai/CLAUDE.md:14-79` | flagged for user (outside station/ scope) |
| 5 | medium | **`station/code-index.md` stale on multiple axes.** (a) `bonsai validate` command not listed at all. (b) CLI Commands table line numbers off (list.go, guide.go, remove.go, update.go, catalog.go all drifted). (c) TUI section missing `catalogflow/`, `guideflow/`, `hints/`, `listflow/`, `removeflow/`, `updateflow/` packages — major undocumented coverage gap. (d) Generator section missing `catalog_snapshot.go` (Plan 31). | `station/code-index.md:17-41`, `:212-287` | flagged for user |

## Errors & Warnings

No errors encountered.

## Resolutions Since 2026-04-21 (last run)

These prior-run flags are now closed:

1. **bubbletea skill nav-table visibility** — added to `station/CLAUDE.md` Skills table via commit `8e21b75`. Resolved.
2. **Routines table formatting (split table)** — `station/agent/Core/routines.md` dashboard now renders as one continuous 7-row table. Plan 26 / PR #66 fix held. Resolved.
3. **Stale `.bak` files in agent/ subdirectories** — all cleaned. Resolved.
4. **Roadmap Phase 1 unchecked boxes** ("UI overhaul", "Usage instructions") — now correctly checked. Resolved.
5. **Status.md "Plan 08 Phase C blocker" stale wording** — Status.md no longer carries the "paused while UI/UX Phase 3 ships" line; current Pending table is empty. Resolved.

## Items Flagged for User Review

> Top 3 in priority order:

1. **[high] Root `Bonsai/CLAUDE.md` project-structure tree drift (finding #4)** — the contributor orientation doc is materially wrong. Quickest path: a single edit pass adding `cmd/init_flow.go`, `cmd/validate.go`, `internal/validate/`, `internal/wsvalidate/`, and the 7 missing `internal/tui/{flow}/` packages. Outside my `station/` scope — needs user edit or code-agent dispatch. Existing Backlog P3 item ("Add root Bonsai/CLAUDE.md check to doc-freshness-check") covers the *process* fix; this finding is the *content* fix.

2. **[medium] `station/CLAUDE.md` broken link to `bonsai-model.md` (finding #3)** — decide: install the skill via `bonsai add tech-lead --skill bonsai-model` (catalog already has it), or remove the broken link. The skill is genuinely useful for the TL agent; install path is recommended.

3. **[medium] `station/code-index.md` stale (finding #5)** — `bonsai validate` invisible, 6 TUI packages undocumented, several line numbers off. Larger refresh needed; consider scheduling alongside a code-index sweep when next plan touches `cmd/` or `internal/tui/`.

Lower-priority items: INDEX.md CLI count bump (finding #1, 1-line edit), INDEX.md architecture diagram polish (finding #2, ongoing slow drift).

## Notes for Next Run

- **Process improvement for routine itself:** existing Backlog P3 item ("Add root Bonsai/CLAUDE.md check to doc-freshness-check routine") is correct; today's finding #4 confirms the recurring pattern (now ~3 cycles in a row with this exact drift class). Worth promoting from P3 → P2 and folding into the routine's `content.md.tmpl` formally.
- **code-index.md drift is a new pattern** (first time flagged) — every plan that adds a `cmd/*.go` or `internal/tui/{flow}/` package compounds the gap. Consider whether code-index.md should be a generated artifact or whether a "code-index sweep" should be a step in the planning workflow's Phase 10 (close-out).
- **Cleanup-worthy:** archive 2026-04-21 reports from `station/Reports/Archive/` are still the latest in archive — that's correct, since 2026-05-04 reports remain in Pending. Routine digest will fold today's findings into one list.
