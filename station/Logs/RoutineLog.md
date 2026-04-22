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

### 2026-04-22 — Plan 23 Phase 3: cinematic `bonsai add` cutover + 7 bundled cleanups (issue-to-implementation)
- **Outcome:** success
- **Plan:** `Playbook/Plans/Archive/23-uiux-phase2-add.md` (status: Complete)
- **PR:** [#64](https://github.com/LastStep/Bonsai/pull/64) squash `788fa6c` (+ post-review fix `44e6874` for 2 minor consistency findings before merge)
- **Iterations:** 1 execute + 1 fix-agent on same branch (2 cycles total, well below the 3-cycle escalation threshold)
- **Execution mode:** tech-lead supervised, 2 general-purpose worktree dispatches (1 implementation + 1 minor-fix) + 2 parallel review-agent dispatches (code review + security review)
- **Changes:** Cutover — `BONSAI_ADD_REDESIGN` env gate deleted; `runAddRedesign` body merged into `cmd/add.go` as new `runAdd`; `cmd/add_redesign.go` deleted; `cmd/init_redesign.go` → `cmd/init_flow.go` via `git mv` (96% rename); legacy `runAdd`/`runAddSpinner`/`buildNewAgentSteps`/`buildAddItemsSteps`/`addOutcome`/`normaliseWorkspace`/`workspaceUniqueValidator`/`newDescriber`/`userSensorOptions`/`NewYieldAddItemsDeferred`/`yieldModeAddItemsDeferred` all removed (zero grep hits across `cmd/` + `internal/`). Bundled cleanups: dead post-harness Generate-error warning deleted from BOTH init_flow.go AND add.go (symmetric path caught by code review on top of plan); 3 harness composition tests for `NewConditional(NewLazy(...))`; `growSucceeded` predicate uses `outcome.SpinnerErr` closure capture; new `addflow.NewYieldUnknownAgent` variant + test; `.bak` write-error silent-discard fixed in BOTH conflict-apply helpers (PR #62 carry-forward security regression closed in both flows); conflicts post-harness slot resolved by type-scan; `cmd/add_test.go` new with 9 table tests using real OS perms.
- **Flags:**
  - **6th occurrence of `gh pr merge --delete-branch` worktree gotcha** — remote branch not deleted, manual `git push origin --delete` + `git worktree remove -f -f` + `git branch -D`. Memory pattern is well-documented; cleanup now reflexive.
  - **New gotcha (potential) — main worktree switched to detached HEAD mid-session** — Suspect fix-agent's `git checkout -B worktree-agent-ad57b8f4 origin/worktree-agent-ad57b8f4` leaked into main worktree despite `isolation: "worktree"`. Caught at PR merge step (`gh pr merge` errored with "could not determine current branch"). Easy recovery via `git checkout main`. Logging here so a pattern can build before adding to memory.md.
  - **Bundling 7 same-file backlog items into Phase 3 was the right call** — Each item touches a file the cutover already opens. Net diff is +947/−990 across 9 files; if each item shipped as its own PR it'd be 7-8 round trips with the same review surface. Code-review agent gave PASS in one pass.
- **Reviews:** Code review PASS + 4 minors (2 real fixed in same-branch follow-up commit `44e6874`: dead post-harness warning in add.go + test rename `DroppedListSorted`→`DroppedListContainsAll`; 2 cosmetic backlogged: `selected[:0]` aliasing pattern + `installedSet` shadowing). Security review PASS no findings — `.bak` regression closed in both helpers, no new file I/O, env-flag fully removed, path-safety verified (keys come from generator-internal `wr.Conflicts()`, not user-controllable).
- **CI:** 6/6 green (test, lint, Analyze Go, govulncheck, CodeQL, GitGuardian) on both `7a1ae2d` (initial) and `44e6874` (post-fix).
- **Notes:** First end-to-end run of issue-to-implementation workflow with Plan 23 as a multi-phase plan. Plan archive flow proven end-to-end (Active → Archive after final phase merge). Phase 3 self-review caught one architectural decision worth documenting: closure capture of `outcome.SpinnerErr` over a sentinel struct — simpler, no new types, predicate doc explicitly justifies safety from harness ordering guarantees.

---

### 2026-04-21 — Plan 20: Security Scanning Infrastructure (session)
- **Outcome:** success
- **Execution mode:** tech-lead supervised, 5 general-purpose worktree dispatches + 1 local gitleaks one-shot
- **Duration:** ~1 hr wall
- **Changes:** 6 PRs merged — #30 (pre-flight station chore), #29 (golangci-lint v1→v2 migration), #28 (Go 1.24.13 → 1.25.8 + lint pin v2.1→v2.11.4, bundled), #31 (Dependabot config: gomod + github-actions, weekly), #40 (govulncheck CI job), #41 (CodeQL workflow for Go SAST, PR + push + weekly)
- **Flags:**
  - CI lint gotcha rediscovered: golangci-lint v2.1 binary built on Go 1.24 still can't parse Go 1.25 targets; must pin to v2.11.4 (or later v2.x built on Go 1.25). Memory updated.
  - gh pr merge --delete-branch silently leaves remote branch behind when branch is checked out in locked worktree — had to `git push origin --delete` + force-remove worktree every time. 4 rounds this session.
  - Dependabot fired immediately after PR #31 merged — 8 auto-PRs (#32-#39) opened. Added to Backlog as [debt] Group G for next-session triage.
  - gitleaks one-shot history audit: 0 findings across 156 commits (6.58 MB). History tag-ready.
  - govulncheck on main (post Go 1.25.8): 0 reachable findings. 2 previously-reachable CVEs (GO-2026-4602, GO-2026-4601) cleared.
- **Plan:** `Playbook/Plans/Active/20-security-scanning-infra.md` (status: Complete)
- **Notes:** Plan originally 4 PRs (#1-#4). Grew to 6 PRs mid-execution because of coupled blockers: PR #0 (pre-flight was only local, needed its own PR) and PR #1a (lint v1→v2 migration required before Go 1.25 bump could pass CI). Lint pin bump got bundled into #28 as paired change. All sequential, all green, all squash-merged.

---

### 2026-04-21 — Vulnerability Scan
- **Outcome:** success
- **Execution mode:** subagent (loop.md dispatch)
- **Duration:** ~4 min
- **Changes:** no changes made (audit-only routine) — dashboard `last_ran` updated to 2026-04-21
- **Flags:** 3 findings (1 new, 2 persistent, 2 resolved since last scan). NEW: GO-2026-4601 (net/url IPv6 parsing, reachable HIGH, `cmd/guide.go:92` via glamour). PERSISTENT (7d): `.env`/`.env.*` still absent from `.gitignore`; GO-2026-4602 (os.ReadDir in `internal/generate/scan.go:44`) still reachable. RESOLVED: GO-2025-3956 and GO-2025-3750 (cleared by go1.24.13). No SAST issues, no hardcoded secrets found. Cross-referenced with today's dependency-audit report — same stdlib CVEs, single fix: upgrade Go 1.24.13 → 1.25.8+. Missing tools: semgrep, gitleaks, trufflehog (Grep fallback used).
- **Report:** `Reports/Pending/2026-04-21-vulnerability-scan.md`

### 2026-04-21 — Doc Freshness Check
- **Outcome:** success
- **Execution mode:** subagent (loop.md dispatch)
- **Duration:** ~5 min
- **Changes:** no changes made (audit-only routine) — dashboard `last_ran` updated to 2026-04-21
- **Flags:** 5 items flagged — (1) [low] `station/INDEX.md` architecture diagram doesn't mention BubbleTea harness added by Plan 15; (2) [info] `bubbletea` custom skill in `agent/Skills/` not listed in CLAUDE.md Skills nav table; (3) [medium] root `Bonsai/CLAUDE.md` project-structure tree's `internal/tui/` block missing `harness/` subdir + `styles_test.go` (Plan 15 drift); (4) [low] Routines table in CLAUDE.md + routines.md dashboard has a broken blank row splitting it into fragments; (5) [info] 10+ stale `.bak` files across `agent/` subdirectories from 2026-04-15 marker migration
- **Report:** `Reports/Pending/2026-04-21-doc-freshness-check.md`

### 2026-04-21 — Backlog Hygiene
- **Outcome:** success
- **Execution mode:** subagent (loop.md dispatch)
- **Duration:** ~4 min
- **Changes:** no changes made (audit-only routine) — no Backlog entries removed, no promotions, dashboard `last_ran` updated
- **Flags:** 4 items flagged for user — (1) Status.md Pending row for "Better trigger sections — Phase C" has stale Blocked By (UI/UX Phase 3 shipped via Plan 14 / PR #24 on 2026-04-17); (2) Roadmap.md Phase 1 checkboxes for "UI overhaul" and "Usage instructions" appear stale (plans shipped); (3) near-duplicate between Group C "CHANGELOG.md + richer release notes" (line 91) and Group D "Changelog generation skill" (line 101); (4) "No Plans Index file" finding from 2026-04-20 Status Hygiene not captured in Backlog
- **Report:** `Reports/Pending/2026-04-21-backlog-hygiene.md`

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

### 2026-04-21 — Dependency Audit
- **Outcome:** partial
- **Execution mode:** subagent (loop.md dispatch)
- **Duration:** ~2 minutes
- **Changes:** no changes made (audit-only routine)
- **Flags:** 11 Go vulns (2 reachable stdlib in `os`/`net/url`, 3 unreachable via `golang.org/x/net` v0.38.0 + stdlib TOCTOU, 6 unreachable stdlib-module-level) — all cleared by Go 1.24.13 → 1.25.9 toolchain bump + `golang.org/x/net@latest`. npm audit on `website/` clean (0 vulns, 455 deps). Python/Rust/Ruby ecosystems N/A (no manifest).
- **Report:** `Reports/Pending/2026-04-21-dependency-audit.md`

### 2026-04-21 — Routine Digest
- **Outcome:** success
- **Reports processed:** 4 — backlog-hygiene, doc-freshness-check, dependency-audit, vulnerability-scan
- **Quick fixes applied:** 4 — added `.env`/`.env.*` to root `.gitignore`, checked "UI overhaul" + "Usage instructions" boxes in Roadmap Phase 1, deleted 16 stale `.bak` files across `agent/` subdirs, added `bubbletea` row to Skills nav table in `station/CLAUDE.md`
- **Backlog items added:** 9 — P1 Go toolchain upgrade (rewrote existing GO-2026-4602 watch item to full upgrade scope); P2: routines-dashboard table fix (Group B), CHANGELOG consolidation decision (Group C), root `Bonsai/CLAUDE.md` tree drift (Group E), Plans Index decision (Group E), `golang.org/x/net` bump (Ungrouped), re-plan "Better trigger sections — Phase C" (Ungrouped); P3: batch Go module refresh (Research), root-CLAUDE.md check sub-step for doc-freshness routine (Routine Enhancements), reduce npm audit cadence (Routine Enhancements)
- **Plan report written:** no (P1 Go upgrade added to Backlog as agreed rather than plan-worthy digest report)
- **Warnings acknowledged:** 2 — scan tool gaps (semgrep/gitleaks/trufflehog — already on Backlog since 2026-04-16); `.env`/.gitignore persistence (subsumed by A4 quick fix)
- **Incidental flag:** GitHub PAT in `~/.claude/settings.json` leaked to conversation context — advised user to rotate (not from routines)
