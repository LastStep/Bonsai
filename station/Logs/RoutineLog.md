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

### 2026-05-03 — Backlog Hygiene
- **Outcome:** success
- **Execution mode:** subagent (loop.md dispatch)
- **Duration:** ~6 min
- **Changes:** no changes to Backlog.md (audit-only run) — dashboard `last_ran` updated to 2026-05-03
- **Flags:** 4 items flagged for user — (1) P1 "stale worktrees" item appears partially self-resolved (one-time sweep done; remaining: add worktree-prune routine); (2) P2 "plan archiving" item partially resolved (Plans/Archive/ directory exists + in use since 2026-04-23; remaining: scaffolding manifest + workflow updates); (3) P2 "changelog generation skill" may be duplicative if tracked as a GH good-first-issue (added via Plan 24 Step E); (4) P2 "golang.org/x/net bump" is now unblocked — Go toolchain upgrade (prerequisite) landed.
- **Report:** `Reports/Pending/2026-05-03-backlog-hygiene.md`

---

### 2026-04-22 — Plan 28 Phase 1 — cinematic `bonsai catalog` + RenderHeader extension + hide `completion` (PR #68, issue-to-implementation)
- **Outcome:** success
- **Plan:** Plans/Active/28-view-cmds-cinematic.md (Tier 2, 3-phase — Phase 1 of 3)
- **Iterations:** 1 execute-review cycle + 1 fix-agent pass (4 minors, 0 majors) + 1 rebase onto landed Plan 27 PR1
- **Changes:** PR #68 squash `1b0b50a` — net +1322/−46 across 25 files. (A) New `internal/tui/catalogflow/` package: `BrowserStage` BubbleTea model embedding `initflow.Stage` w/ rail hidden; 7 tabs (Agents/Skills/Workflows/Protocols/Sensors/Routines/Scaffolding); `← → / h l` tab wraps, `↑ ↓ / j k` focus clamp, `?` inline-expand, `q / esc / ctrl-c / enter` quit; `-a <agent>` filter greys empty-for-agent tabs `(0)` suffix but keeps them in strip; narrow-width (<96 cols) switches to 5-char short labels (AGENT/SKILL/FLOWS/PROTO/SENSE/RTNES/SCAFF) + 1-space separator to fit 70-col min floor. (B) `initflow.RenderHeader` signature extended: `(version, projectDir, action, rightLabel string, width int, safe bool)`; empty `rightLabel` collapses right-block row 1. `Stage` struct gains `headerAction` + `headerRightLabel` fields + `ApplyContextHeader(ctx)` setter; `StageContext` gains `HeaderAction` + `HeaderRightLabel`. All 12 existing stage ctors (init + add) add `base.ApplyContextHeader(ctx)`; `cmd/init_flow.go` + `cmd/add.go` stamp `"INIT"` + `"PLANTING INTO"` on `StageContext`. (C) `cmd/catalog.go` rewired: factored body into `renderCatalogStatic(cat, agentFilter)`; TTY invocation launches `catalogflow.NewBrowser(cat, agentFilter, projectDir)` via `tea.NewProgram(stage, tea.WithAltScreen()).Run()` with error wrap `"catalog browser: %w"`; non-TTY falls back to static (piped behavior unchanged). (D) `cmd/root.go` adds `rootCmd.CompletionOptions.HiddenDefaultCmd = true` — `bonsai --help` drops `completion` from listing, `bonsai completion zsh` still emits script.
- **Flags:** (1) `git commit` leak — path-restricted `git add <plan-file>` + `git commit` grabbed parallel Plan 27 session's pre-staged `graft.go → branches.go` rename from shared git index (`d9ddee4`). Pure 100%-similarity rename, no caller breakage; left in place and flagged to parallel session. Fix-prevention: memory already documents this; use `git commit -- <paths>` or `git diff --staged` pre-commit when parallel-session WIP exists. (2) Mid-Phase-1 rebase — Plan 27 PR1 (`0885981`) landed on main while Phase 1 was in review, making PR #68 `DIRTY`. Rebased onto origin/main from the agent worktree; single content conflict in `internal/tui/addflow/ground.go` resolved by keeping BOTH lines (peer-stage pattern: `ApplyContextHeader(ctx)` then `SetRailLabels(StageLabels)`). Force-pushed `c1fea7e`; CI re-ran green. (3) Worktree-held-branch post-merge — pattern hit 8× this month; manual `git worktree remove -f -f` + `git branch -D` + `git push origin --delete` required after `gh pr merge --delete-branch`.
- **Notes:** Plan pre-committed to origin/main (`d9ddee4`) before dispatch so agent worktree saw current spec. Independent code-review returned PASS-WITH-MINORS (0 major / 4 minor / 6 nit). Fix-agent dispatched on same branch for all 4 minors (projectDir threading through ctor, short-label tab strip under 96 cols + regression tests, routed `View()` through `Stage.RenderFrame(body, keys)` path A consuming the stored header setters, `tea.Run` error wrap) — commit `64119e5`. 6 NITs filed to Backlog Group B as `[Plan-28 cosmetic]`. 6/6 CI green post-rebase (test/lint/Analyze-Go/govulncheck/CodeQL/GitGuardian). Parallel-session coordination: Plan 27 session's rebase onto my earlier `d9ddee4` rename commit went clean (zero-content overlap). Stash + ff-pull + unstash on main post-merge preserved parallel session's 4-file WIP (`conflicts.go`/`ground.go`/`grow.go`/`initflow/generate.go`) untouched.
- **Status updates:** Status.md — Plan 28 row updated (Phase 1 shipped, Phases 2+3 remain); Plan 28 Phase 1 added to Recently Done. Backlog.md — 6 NITs filed as single bundled Plan-28-cosmetic entry in Group B. Plan 28 file remains in `Plans/Active/` (Phases 2+3 not yet shipped).

---

### 2026-04-22 — Plan 26 — P2 knock-off bundle (PR #66)
- **Outcome:** success
- **Plan:** Plans/Archive/26-p2-knockoff-bundle.md (Tier 1 patch, single dispatch)
- **Iterations:** 1 execute-review cycle (zero findings on code review)
- **Changes:** PR #66 squash `a9df552` — 5 files, +149/−14. (1) `station/agent/Sensors/context-guard.sh` lines 156-157 swapped `os.path.join(root, "")` → `docs_path` so planning-reminder paths include `station/` prefix (matches wrap-up checklist at lines 122-124). (2) `station/agent/Core/routines.md` dropped 2 blank lines (orig 38 + 55) splitting Dashboard + Definitions markdown tables — generator itself was already correct, file was stale output of older generator. (3) `cmd/root.go:173` + `cmd/add.go:329` replaced `filtered := slice[:0]` aliasing with `make([]string, 0, len(slice)-len(dropped))` pre-allocation — two sites; capacity arithmetic safe (`dropped ⊆ toBackup ⊆ toOverwrite` in add, `dropped ⊆ selected` in root). (4) `cmd/add.go` renamed inner closures `installedSet` → `installedItems` at lines 532 + 619 + 6 internal call sites (541/550/559/628/639/650); file-scope `func installedSet` at 349 preserved + its caller at 119 untouched. (5) `internal/generate/generate_test.go` new `TestRoutineDashboardNoBlankRows` — builds catalog with 7 tech-lead-default routines at mixed 5/7/14-day frequencies, calls `RoutineDashboard`, asserts no blank row splits body rows inside `ROUTINE_DASHBOARD_START`/`END` markers OR inside Routine Definitions table body; tolerates generator's expected blanks immediately after START/before END.
- **Flags:** Agent's initial edits accidentally landed in main worktree (Edit tool writes via absolute path); agent self-recovered via `git apply` to worktree + `git checkout --` revert on main, pre-merge main verified clean. Lockfile hash updated in main worktree only (`.bonsai-lock.yaml` gitignored) so `bonsai update` won't flag spurious conflict.
- **Notes:** All 4 items diagnosed + planned pre-dispatch. Plan pre-dispatched to origin/main (`fb6ac8c`) so agent worktree saw current spec. Independent code-review agent returned PASS with zero findings (no major/minor/nit). 6/6 CI green (test/lint/Analyze-Go/govulncheck/CodeQL/GitGuardian). Post-merge cleanup: worktree held branch → manual `git worktree remove -f -f` + `git branch -D` + `git push origin --delete` (documented pattern hits again, ~7th time this month).
- **Status updates:** Plan 26 archived. Backlog 4 items removed (context-guard path prefix, routines dashboard split, selected[:0] aliasing, installedSet shadowing).

---

### 2026-04-22 — v0.2.0 release ship — pre-release docs audit + tag + Homebrew tap recovery
- **Outcome:** success
- **Plan:** none (scope-only execution per user choice Q-A; Plan 26 candidates filed to Backlog instead of formal plan doc)
- **Changes:** PR #65 squash `e9b7cba` "docs(audit): pre-release v0.2.0 cut" — 15 files, +224/−96. P0+P1 mechanical (commit `1be622a`, self): sensor count 12→13 everywhere via `website/scripts/generate-catalog.mjs` + manual README; tagline mechanical replace "structured language" → "workspace" across `index.mdx` + `astro.config.mjs` + `why-bonsai.mdx`; invalid agent types stripped (`reviewer` from `review-checklist`, `qa` from `test-strategy`); LICENSE 2025 → 2025-2026; Go 1.24+ → 1.25+ in README + CONTRIBUTING; `station/CLAUDE.md.bak` deleted; `Plans/Archive/23-uiux-phase2-add.md` frontmatter Draft → Complete. P2 + CHANGELOG (3 commits `ff24fce` + `c7d9442` + `c05fca4`, dispatched general-purpose agent): `code-index.md` refresh post-Plan-22+23, `docs/README.md` clarify embedded-vs-website split, `CHANGELOG.md` `[Unreleased]` → `[0.2.0] - 2026-04-22` with curated Added/Changed/Removed/Fixed/Security covering Plans 19-25 + statusLine + push-CI widening + binary path fix + Go bump; v0.1.x dates backfilled `2026-04-15` → `2026-04-16`. Tag `v0.2.0` annotated + pushed; GoReleaser Release workflow built+uploaded 6 platform binaries + checksums (release v0.2.0 published with all assets). Manual Homebrew formula push: `gh api -X PUT /repos/LastStep/homebrew-tap/contents/Formula/bonsai.rb` with SHAs from `checksums.txt` (commit `29fd0da` on tap; v0.2.0 + 4 platform SHAs verified live via API).
- **Flags:** **HOMEBREW_TAP_TOKEN PAT expired** — GoReleaser brew step failed at 401 Bad credentials. User rotated PAT + ran `gh secret set HOMEBREW_TAP_TOKEN --repo LastStep/Bonsai`. New PAT works for next release. Filed Backlog P1: workflow_dispatch trigger on release.yml + PAT expiry calendar (~2026-07-15).
- **Notes:** Q-verification before edits caught 2 false positives in audit reports (workspace-guide `.md.tmpl` works as designed per `catalog.go:361,456` both-suffix support + `generate.go:782` special-case; skills frontmatter missing on 13 of 17 is cosmetic since loader reads triggers from meta.yaml not file frontmatter — both deferred to Plan 26). Parallel session pushed 2 unrelated commits mid-work to main (`d5a7dc9` + `d25eef1`, non-overlapping files, clean merge forward). Recovery path for failed brew step chose manual API push over workflow retry — safer for one-off because no `workflow_dispatch:` trigger exists AND `goreleaser release --clean` on existing tag risks recreating GH Release with new timestamps. **All pre-OSS-launch gates cleared.**
- **Status updates:** Status.md Recently Done table reset to today's items only; pre-2026-04-22 items moved to StatusArchive.md (45 entries archived). memory.md Work State + Notes refreshed with v0.2.0 ship + 4 new gotchas (Homebrew PAT, session-start catalog generator diffs are load-bearing, GoReleaser `--clean` is destructive, workflow_dispatch should be added to release workflows pre-launch).

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
- **Post-merge verification audit:** Dispatched 2 parallel verification agents (code state + bookkeeping). Code PASS + 1 finding: 4 stale doc-comment refs to deleted `cmd/add_redesign.go` in `internal/tui/addflow/{addflow,ground,conflicts}.go` and `internal/config/lockfile.go` — fixed in `d5a7dc9`. Bookkeeping PASS + 1 unrelated finding: Plans 15 + 19 completed 2026-04-21 but never archived — bundled into `d5a7dc9` while audit attention was on plan-archive correctness. Confirmed pattern: post-merge verification catches small stragglers worth a follow-up commit; cheap when done in same session.

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

### 2026-04-25 — Memory Consolidation
- **Outcome:** success
- **Execution mode:** main agent (session-start)
- **Auto-memory scanned:** `~/.claude/projects/-home-rohan-ZenGarden-Bonsai/memory/MEMORY.md` (stub only) + `-home-rohan-ZenGarden-Bonsai-station/memory/MEMORY.md` (stub only) — no facts to bridge.
- **Agent memory validated:** Notes (14 gotchas) + Feedback (durable UX prefs 2026-04-17) + References (6 research docs at `station/Research/RESEARCH-*.md`, all exist) — all current, no stale entries marked.
- **Work State refreshed:** updated from "Plan 30 shipped 2026-04-23" → "v0.3.0 shipped 2026-04-24 via Plan 31 (PRs #75/#76/#77)".
- **Flags:** none active.
- **Decisions applied:** 0 keep-as-stale, 0 update, 0 archive, 0 insert_new — auto-memory consolidation is a no-op when auto-memory is empty stubs (intended steady state per Bonsai memory model).

### 2026-04-25 — Status Hygiene
- **Outcome:** success
- **Execution mode:** main agent (session-start)
- **Archive sweep:** none — oldest Recently Done row is 2026-04-22 (3d), well under 14d threshold.
- **Pending validated:** only standing comment "Plan 26 candidates filed in Backlog P2 Group C" — still relevant, no rows to flag.
- **Plan/Status cross-ref:** Plans/Active/ empty (no orphans); Status Recently Done rows for plans 23–31 all map to files in Plans/Archive/ (31 plan files total).
- **Backlog cross-ref:** "Re-archive Plan 29" already commented out in Backlog (line 58, resolved 2026-04-23 via archive-reconcile sweep). No new resolutions to remove.
- **Net changes:** 0 archived, 0 flagged, 0 demoted.
