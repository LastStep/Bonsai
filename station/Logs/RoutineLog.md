---
tags: [logs, routines]
description: Append-only audit trail for routine executions. Each entry records outcome, changes, and notes.
---

# Routine Log

> [!note]
> Agents append to this log after completing a routine. Do not edit existing entries — this is an audit trail.

**Format:**

```
### YYYY-MM-DD — Routine Name
- **Outcome:** success | partial | skipped | deferred
- **Changes:** what was modified
- **Flags:** issues found
- **Notes:** context for future runs
```

---

### 2026-04-20 — Plan 18: `bonsai guide` multi-topic + legacy docs cleanup
- **Plan:** Playbook/Plans/Active/18-bonsai-guide-multi-topic.md
- **PR:** #25 (squash `e448140`)
- **Iterations:** 1 execute-review cycle (no fix dispatch needed)
- **Issues found:** none — all 4 agent-flagged decisions were reasonable interpretations
- **Scope delivered:** 3 terminal cheatsheets (quickstart 93L, concepts 113L, cli 119L) + Huh picker + direct-arg allowlist; dropped 3 orphan docs (1,213 lines); CLAUDE.md doc-drift fix rolled in
- **Verification:** CI green (test + lint + GitGuardian); post-merge `make build` + `go test ./...` + `./bonsai guide {quickstart,unknown,a b}` smoke all pass
- **Result:** completed

### 2026-04-20 — Routine Digest
- **Outcome:** success
- **Reports processed:** 2 — Memory Consolidation, Status Hygiene (both 2026-04-20)
- **Quick fixes applied:** 3 — rewrote status-hygiene Step 3 (Plans Index → Plan files vs Status rows) in catalog template + installed copy; created `Playbook/StatusArchive.md` stub; acknowledged References-section-drift warning (watch next memory-consolidation run).
- **Backlog items added:** 0
- **Plan report written:** no
- **Warnings acknowledged:** 1 — References section emptied between 2026-04-14 and 2026-04-20 memory-consolidation runs; re-populated today; watch next run to confirm it persists.

### 2026-04-20 — Status Hygiene
- **Outcome:** success
- **Execution mode:** subagent (loop.md dispatch)
- **Duration:** ~4 min
- **Changes:** Minor fix to `Playbook/Backlog.md` Group B intro text — removed stale "triggerSection frontmatter" reference (bug was fixed in Plan 17 / PR #24). No Status.md archival needed (oldest Done item is 8 days old).
- **Flags:** 3 items flagged — (1) no `StatusArchive.md` exists yet (will need creation when items age past 14 days, earliest 2026-04-26), (2) no Plans Index file exists at all (routine procedure references it), (3) 17 plan files sit in `Plans/Active/` but most are merged — known P2 backlog item (Group E plan archiving) still pending.
- **Report:** `Reports/Pending/2026-04-20-status-hygiene.md`

### 2026-04-20 — Memory Consolidation
- **Outcome:** success
- **Execution mode:** subagent (loop.md dispatch)
- **Duration:** ~4 min
- **Changes:** Added 5 Research doc pointers to `agent/Core/memory.md` References section with corrected file paths (prior entry used stale RESEARCH.md/RESEARCH-concepts.md/RESEARCH-evals.md naming). Archived 2 stale auto-memory files (`project_go_rewrite.md` — redundant with root CLAUDE.md; `project_research_phase.md` — stale file names superseded by References section). Trimmed auto-memory MEMORY.md index to minimal pointer.
- **Flags:** none
- **Report:** `Reports/Pending/2026-04-20-memory-consolidation.md`

### 2026-04-16 — Routine Digest
- **Outcome:** success
- **Reports processed:** 8 (Backlog Hygiene, Status Hygiene, Vulnerability Scan, Dependency Audit, Memory Consolidation, Doc Freshness Check, Infra Drift Check, Roadmap Accuracy)
- **Quick fixes applied:** 3 (Roadmap checkboxes for "Custom item detection" and "Community health files", root CLAUDE.md project structure update)
- **Backlog items added:** 5 (2 P1: code index drift + Go toolchain upgrade; 3 P2: remove infra-drift-check, install semgrep/gitleaks, consolidate usage instructions)
- **Plan report written:** no
- **Warnings acknowledged:** 2 (.env gitignore, "Usage instructions" tracking gap)

### 2026-04-16 — PR #8: Fix case-insensitive file collision (issue-to-implementation)
- **Plan:** Plans/Active/06-case-insensitive-file-collision.md
- **Iterations:** 1 execute-review cycle
- **Issues found:** none
- **Result:** completed — merged via squash, PR #8

### 2026-04-15 — Selective file update in conflict resolution (issue-to-implementation)
- **Outcome:** completed
- **Plan:** Tier 1 (Patch) — no plan file needed
- **Iterations:** 1 execute cycle, 0 review rejections
- **Changes:** `internal/generate/generate.go` — added `ForceSelected()` method for per-file conflict forcing; `cmd/root.go` — rewrote `resolveConflicts()` from all-or-nothing skip/overwrite/backup to multi-select file picker with per-selection backup option
- **Issues found:** None
- **Result:** Users now see a multi-select picker listing each conflicted file, can choose which to update and which to skip, with optional .bak backup for selected files

### 2026-04-15 — Doubled path prefix fix (issue-to-implementation)
- **Outcome:** completed
- **Plan:** Tier 1 (Patch) — no plan file needed
- **Iterations:** 1 execute cycle, 0 review rejections
- **Changes:** `cmd/root.go` — strip rootLabel prefix from RelPath in `showWriteResults` before building file trees; added `strings` import
- **Issues found:** None
- **Result:** Display bug fixed. Created/Updated/Skipped panels now show `station/agent/...` instead of `station/station/agent/...`

### 2026-04-14 — Backlog Hygiene
- **Outcome:** success
- **Execution mode:** subagent (loop.md dispatch)
- **Duration:** ~3 min
- **Changes:** Removed 3 items from Backlog.md (2 duplicated in Status.md Pending, 1 already completed). Replaced with HTML comments for audit trail.
- **Flags:** Roadmap.md Phase 2 has stale "Custom item detection" checkbox (done but unchecked). P2 item "bonsai guide command" may warrant promotion to P1 (aligns with current Phase 1 milestone).
- **Report:** `Reports/Pending/2026-04-14-backlog-hygiene.md`

### 2026-04-14 — Status Hygiene
- **Outcome:** success
- **Execution mode:** subagent (loop.md dispatch)
- **Duration:** ~2 min
- **Changes:** no changes made (audit-only routine) — all Status.md items are current, no archiving needed, no stale Pending items, no plan index issues
- **Flags:** none
- **Report:** `Reports/Pending/2026-04-14-status-hygiene.md`

### 2026-04-14 — Doc Freshness Check
- **Outcome:** partial
- **Execution mode:** subagent (loop.md dispatch)
- **Duration:** ~5 min
- **Changes:** no changes made (audit-only routine)
- **Flags:** 5 findings — INDEX.md key metrics stale (CLI commands 5->6, catalog items ~30->46); root CLAUDE.md project structure missing cmd/update.go and 4 generate/ files; code index missing bonsai update + 2 new source files; 11 line number references drifted in generate.go section
- **Report:** `Reports/Pending/2026-04-14-doc-freshness-check.md`

### 2026-04-14 — Roadmap Accuracy
- **Outcome:** success
- **Execution mode:** subagent (loop.md dispatch)
- **Duration:** ~4 min
- **Changes:** no changes made (audit-only routine)
- **Flags:** 2 findings — Phase 2 "Custom item detection" checkbox stale (done but unchecked); "Usage instructions" Phase 1 item not tracked in Status.md Pending
- **Report:** `Reports/Pending/2026-04-14-roadmap-accuracy.md`

### 2026-04-14 — Memory Consolidation
- **Outcome:** success
- **Execution mode:** subagent (loop.md dispatch)
- **Duration:** ~4 min
- **Changes:** Updated security-audit agents note (added tech-lead); added 3 foundational RESEARCH docs to References section (RESEARCH.md, RESEARCH-concepts.md, RESEARCH-evals.md)
- **Flags:** none
- **Report:** `Reports/Pending/2026-04-14-memory-consolidation.md`

### 2026-04-14 — Infra Drift Check
- **Outcome:** success
- **Execution mode:** subagent (loop.md dispatch)
- **Duration:** ~2 min
- **Changes:** no changes made (audit-only routine) — no IaC files (Terraform, Pulumi, CloudFormation, Docker, CI/CD) exist in the project
- **Flags:** Routine has no work to do for Bonsai (local Go CLI with no cloud infrastructure). Consider removing `infra-drift-check` from `.bonsai.yaml` unless cloud infra is planned.
- **Report:** `Reports/Pending/2026-04-14-infra-drift-check.md`

### 2026-04-14 — Dependency Audit
- **Outcome:** success
- **Execution mode:** subagent (loop.md dispatch)
- **Duration:** ~4 min
- **Changes:** no changes made (audit-only routine)
- **Flags:** Go toolchain go1.24.3 is behind latest patches — 3 symbol-level stdlib vulns (GO-2025-3956, GO-2025-3750, GO-2026-4602), 4 package-level, 23 module-level. Upgrading to go1.24.13 resolves 2 of 3 symbol-level and all module-level vulns. No third-party dependency vulnerabilities found.
- **Report:** `Reports/Pending/2026-04-14-dependency-audit.md`

### 2026-04-14 — Vulnerability Scan
- **Outcome:** success
- **Execution mode:** subagent (loop.md dispatch)
- **Duration:** ~4 min
- **Changes:** no changes made (audit-only routine)
- **Flags:** 4 findings — `.env` missing from `.gitignore` (low); 3 Go stdlib vulns via govulncheck (GO-2025-3956 medium, GO-2025-3750 low/Windows-only, GO-2026-4602 low). No SAST issues, no hardcoded secrets, no third-party dependency vulns. Stdlib findings cross-referenced with dependency-audit routine.
- **Report:** `Reports/Pending/2026-04-14-vulnerability-scan.md`

### 2026-04-15 — `bonsai guide` command (issue-to-implementation)
- **Plan:** Plans/Active/02-bonsai-guide-command.md
- **Iterations:** 1 execute-review cycle
- **Issues found:** none
- **Result:** completed — PR #3 (draft)

### 2026-04-15 — CLAUDE.md marker migration (issue-to-implementation)
- **Plan:** Plans/Active/01-claudemd-marker-migration.md
- **Iterations:** 1 execute-review cycle
- **Issues found:** none
- **Result:** completed — PR #1 (draft)

### 2026-04-15 — Rename "catalog items" to "abilities" (issue-to-implementation)
- **Plan:** No plan — Tier 1 patch
- **Iterations:** 1 execute-review cycle
- **Issues found:** none
- **Result:** completed — PR #2, reviewed and marked ready

### 2026-04-15 — CI workflow (issue-to-implementation)
- **Plan:** No plan — Tier 1 patch
- **Iterations:** 1 execute-review cycle
- **Issues found:** none
- **Result:** completed — PR #6, reviewed and marked ready

### 2026-04-15 — Release pipeline (issue-to-implementation)
- **Plan:** Plans/Active/04-release-pipeline.md
- **Iterations:** 1 execute-review cycle
- **Issues found:** none (reviewer noted 2 minor observations: unused exported `Version` var in cmd, Actions pinned to major versions not SHAs — neither blocking)
- **Result:** completed — PR #5 (draft)

### 2026-04-16 — Plan 05: AI Operational Intelligence (issue-to-implementation)
- **Plan:** Plans/Active/05-usage-instructions.md
- **Iterations:** 1 execute-review cycle
- **Issues found:** Agent created PR on `feat/usage-instructions` branch but also left a stale `worktree-agent-ad11a6d4` branch ref — cleaned up during review
- **Result:** completed — PR #7 (ready for review)
