---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Memory Consolidation"
date: 2026-07-21
status: partial
---

# Routine Report — Memory Consolidation

## Overview
- **Routine:** Memory Consolidation
- **Frequency:** Every 5 days
- **Last Ran:** 2026-05-07
- **Triggered By:** loop.md autonomous dispatch

## Execution Metadata
- **Status:** partial
- **Duration:** ~8 min
- **Files Read:** 5
  - `station/agent/Routines/memory-consolidation.md`
  - `station/agent/Core/memory.md`
  - `station/agent/Core/routines.md`
  - `station/Logs/RoutineLog.md`
  - `station/Playbook/Status.md`
- **Files Modified:** 2
  - `station/agent/Core/memory.md` — stale markers added to References section
  - `station/agent/Core/routines.md` — dashboard Last Ran / Next Due updated
- **Tools Used:** Bash (ls, grep, sed, git log, git tag, find, wc -l, git -C log)
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Read auto-memory sources
- **Action:** Noted that this project explicitly does NOT use Claude Code's auto-memory system per project instructions. No `~/.claude/projects/*/memory/MEMORY.md` files were scanned.
- **Result:** Step skipped — auto-memory is disabled by design. This is the documented steady state: auto-memory has been canonical-stub-only since at least 2026-04-25 per RoutineLog history.
- **Issues:** None — expected behavior per project CLAUDE.md.

### Step 2: Read current agent memory
- **Action:** Read `station/agent/Core/memory.md` in full.
- **Result:** Memory file has 5 sections: Flags (empty), Work State (Plan 41 context), Notes (20 durable gotchas), Feedback (durable UX prefs), References (6 research doc pointers). All sections loaded.
- **Issues:** None.

### Step 3: Apply consolidation decisions for auto-memory entries
- **Action:** No auto-memory entries to process (Step 1 was a no-op).
- **Result:** Zero keep/update/archive/insert_new decisions — same as prior runs.
- **Issues:** None.

### Step 4: Validate agent memory against codebase
- **Action:** Systematically checked file paths, functions, and architecture references in memory.md against the current codebase.
- **Result:** See detailed findings below. Most references are accurate; the References section has 6 stale file pointers.
- **Issues:** Research docs directory (`station/Research/`) does not exist. None of the 6 RESEARCH-*.md files exist in the repository or git history.

**Validated (accurate):**
- `docs/agent-interface.md` — exists at `/home/user/Bonsai/docs/agent-interface.md`
- `internal/nonint/` package — exists with all expected files (runner.go, nonint.go, remove.go, etc.)
- `ExitWrongCWDForInit = 4` (exit 4) — confirmed in `internal/nonint/runner.go:42`
- `ExitConflict = 5` — confirmed in `internal/nonint/runner.go:46`
- `catalog_snapshot_unix.go` + `catalog_snapshot_windows.go` — platform split confirmed (Note re: `syscall.O_NOFOLLOW` is accurate)
- `os.ReadDir` in `internal/generate/scan.go` — at line 45 (note says line 44, minor drift, accurate)
- `cmd/guide.go` — glamour dependency confirmed active
- `station/Playbook/Standards/NoteStandards.md` — exists
- `station/Playbook/Standards/SecurityStandards.md` — exists
- Plans 40 + 41 files in `Plans/Active/` — confirmed (note re: archiving is still valid)
- `.bonsai-lock.yaml` in `.gitignore` — confirmed (the gitignore policy is already implemented)
- Both `.md` and `.md.tmpl` suffix support in `internal/catalog/catalog.go` — confirmed at lines ~427, 522 (note says 361, 456 — line numbers drifted but behavior accurate)
- workspace-guide special case in `internal/generate/generate.go` — confirmed at line 976 (note says 782 — line numbers drifted but behavior accurate)

**Stale (file not found):**
- `station/Research/RESEARCH-landscape-analysis.md` — missing
- `station/Research/RESEARCH-concept-decisions.md` — missing
- `station/Research/RESEARCH-eval-system.md` — missing
- `station/Research/RESEARCH-trigger-system.md` — missing
- `station/Research/RESEARCH-uiux-overhaul.md` — missing
- `station/Research/RESEARCH-proof-of-bonsai-effectiveness.md` — missing

All 6 files and the `station/Research/` directory itself are absent from the filesystem and have no trace in git history. Last validated as present: 2026-05-07 (per RoutineLog). These may have been local-only files that were never committed to the repo.

### Step 5: Check memory protocol compliance
- **Action:** Reviewed all sections for entries persisting 3+ sessions without action, and flags without resolution paths.
- **Result:**
  - **Flags section:** Empty (`(none)`) — clean.
  - **Work State:** Plan 41 archiving (archive Plans/Active/40 + 41 to Archive/) noted as open since June 2026 — escalated to findings. The note "dogfood still needs `.bonsai-lock.yaml` gitignore policy" appears to be resolved (`.bonsai-lock.yaml` IS in `.gitignore`); this portion of the Work State note is stale — flagged for user cleanup.
  - **Notes section:** All 20 durable gotchas reviewed. These are intended to persist indefinitely as operational knowledge; none triggered the 3-session escalation criterion (they are reference notes, not action items).
  - **Feedback section:** Durable UX preferences and confirmed approaches — appropriate for long-term retention.
- **Issues:** Two Work State items flagged (see above).

### Step 6: Clean auto-memory
- **Action:** Not applicable — no auto-memory to clean.
- **Result:** Skipped.
- **Issues:** None.

### Step 7: Log results
- **Action:** Appended entry to `station/Logs/RoutineLog.md`.
- **Result:** Entry added.
- **Issues:** None.

### Step 8: Update dashboard
- **Action:** Updated `Memory Consolidation` row in `station/agent/Core/routines.md`.
- **Result:** Last Ran → 2026-07-21, Next Due → 2026-07-26, Status → done.
- **Issues:** None.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | high | All 6 RESEARCH-*.md files missing — `station/Research/` directory does not exist in repo or git history; last validated present 2026-05-07 | `station/agent/Core/memory.md` — References section | Marked each entry `(stale — file not found)` with explanation; flagged for user: confirm if files exist elsewhere or remove block |
| 2 | medium | "dogfood still needs `.bonsai-lock.yaml` gitignore policy" in Work State — `.bonsai-lock.yaml` IS already in `.gitignore` | `station/agent/Core/memory.md` — Work State | Flagged for user to clean up the resolved note from Work State compound sentence |
| 3 | low | Plans 40 + 41 remain in `Plans/Active/` since June 2026 — noted in Work State as needing archiving at "next wrap-up" | `station/Playbook/Plans/Active/` | No action taken (plan archiving is a Tech Lead session task, not routine scope) — existing Backlog Group E item covers this |
| 4 | info | Line numbers in Notes drifted for `catalog.go:361,456` (now ~427, 522) and `generate.go:782` (now ~976) | `station/agent/Core/memory.md` — Notes | Not updated — behavior described is accurate; line numbers are illustrative references and shift with development. No action needed. |

## Errors & Warnings

No errors encountered.

## Items Flagged for User Review

- **Research docs missing (high):** The entire `station/Research/` directory and all 6 `RESEARCH-*.md` files referenced in `agent/Core/memory.md` are absent from the filesystem and have no git history. These were validated as present on 2026-05-07. They may have been local-only files never committed. User action needed: (a) if the files exist on another machine or backup, commit them; (b) if lost, remove the References block from memory.md or replace with whatever docs now serve that purpose.

- **Work State cleanup (medium):** The Work State sentence "Plan 40 P1-3 still untagged/tag-held; dogfood still needs `.bonsai-lock.yaml` gitignore policy" contains a resolved item — `.bonsai-lock.yaml` IS in `.gitignore`. User may want to remove/update that clause during next session wrap-up.

## Notes for Next Run

- If Research docs remain missing, the stale markers added today will still be present. Next run should confirm whether user resolved the issue (committed or removed the block) and clean up stale markers accordingly.
- Auto-memory remains in canonical-stub steady state — auto-memory bridging steps continue to be no-ops.
- Plans 40 + 41 archiving is a standing item; if still in Active/ at next run, escalate.
