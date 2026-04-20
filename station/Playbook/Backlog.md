---
tags: [playbook, backlog]
description: Prioritized backlog — bugs, features, debt, research, and improvement ideas. Self-maintained by agents via the backlog-hygiene routine.
---

# Bonsai — Backlog

> [!note]
> This is the intake queue for all work not yet in `Status.md`. Items flow from here into active work.
> For current active work, see `Playbook/Status.md`. For long-term direction, see `Playbook/Roadmap.md`.

---

## How This Works

**Capture:** When you discover a bug, improvement opportunity, tech debt, or idea during a session that is outside your current task scope — add it here instead of fixing it inline. Use the item format below.

**Promote:** When capacity opens, move P0/P1 items into `Playbook/Status.md` as Pending or In Progress. Remove the item from this file when it appears in Status.

**Resolve:** Items completed via Status.md are cleaned up by the backlog-hygiene routine. Items abandoned or made irrelevant should be removed with a note in `Logs/RoutineLog.md`.

**Review:** The backlog-hygiene routine runs periodically to flag stale items, escalate misplaced P0s, remove duplicates, and cross-reference with Status.md and Roadmap.md.

### Item Format

```markdown
- **[category] Short description** — Context or rationale. *(added YYYY-MM-DD, source: routine|session|user)*
```

**Categories:** `bug`, `feature`, `debt`, `security`, `research`, `improvement`

### Priority Guide

| Priority | Meaning | Action |
|----------|---------|--------|
| **P0** | Blocking current work or broken functionality | Must be in Status.md. If a P0 is here, escalate it immediately |
| **P1** | Next up when current work completes | Promote to Status.md when capacity opens |
| **P2** | Planned but not urgent | Review at phase boundaries |
| **P3** | Ideas, nice-to-haves, research topics | Review during roadmap updates |

### Groups

Items that should be worked together are tagged with a group letter. See the group index at the bottom of this file for phasing and dependency info.

---

## P0 — Critical

(none)

## P1 — High

- **[bug] Silent error swallowing in spinner callbacks** `[Group B]` — 41 errors are discarded in production code because the Huh spinner callback signature is `func()` — it can't return errors. Generation failures during `bonsai init/add/remove/update` are invisible to users. A corrupted template, missing file, or permission error succeeds silently and leaves the workspace broken. Fix: collect errors inside the callback via closure (e.g., append to `[]error`), then check after the spinner completes. Affects: `cmd/add.go` (10), `cmd/init.go` (5), `cmd/update.go` (5), `cmd/remove.go` (14), `internal/generate/generate.go` (6), `cmd/root.go` (1). *Scope note:* Plan 15's BubbleTea harness migration will centralize spinner handling — fix there rather than patching 34+ callsites that are about to be deleted. *(added 2026-04-16, source: repo-analytics)*
- **[security] Monitor GO-2026-4602 (FileInfo Root escape, needs go1.25.8)** — GO-2025-3956 + GO-2025-3750 resolved in Plan 17 (toolchain bumped to go1.24.13, PR #24). GO-2026-4602 still unresolved in 1.24.x series; requires go1.25.8. Revisit once go1.25 is stable. *(added 2026-04-16, updated 2026-04-17 post-Plan-17, source: routine-digest)*
<!-- "triggerSection() prepends before YAML frontmatter" — fixed 2026-04-17 via injectTriggerSection helper, Plan 17 / PR #24 -->
<!-- "Upgrade Go toolchain from 1.24.3 to 1.24.13+" — fixed 2026-04-17, Plan 17 / PR #24 -->
<!-- "Spinner error swallowing" — scope-deferred to Plan 15 harness work, 2026-04-17 -->
<!-- ".golangci.yml + test/lint/fmt Makefile targets" — shipped 2026-04-17, Plan 17 / PR #24 (demo GIF item remains, see Group C below) -->
- **[debt] Testing infrastructure for triggers and sensors** `[Group B]` — No testing infrastructure exists for hook-based triggers, prompt hooks, context-guard regex patterns, path-scoped rules, or skill auto-invocation. Need: (1) unit tests for context-guard regex patterns (positive/negative cases), (2) integration test harness for sensor scripts (mock stdin, verify stdout/exit codes), (3) end-to-end test framework for trigger activation (simulate user prompts, verify correct ability loads), (4) prompt hook evaluation testing (verify Haiku correctly classifies intents). The trigger system is expanding significantly — without test infra, regressions will be invisible. *(added 2026-04-16, source: user)*
<!-- "go install . installs binary as Bonsai (capital B)" — fixed 2026-04-20 via Plan 16 / PR #23 (option 2: main.go → cmd/bonsai/main.go + root embed.go) -->
- **[debt] `Bonsai/CLAUDE.md` ProjectStructure tree references stale `main.go` location** `[doc-drift]` — After Plan 16 merge, lines 19 (`├── main.go ← entry point, embeds catalog/ via embed.FS`) and 109 (`Catalog is embedded via embed.FS in main.go`) still describe pre-move layout. Linter/user already fixed line 118 install command, but the structural tree + key-concept bullet were missed. Should now read: `├── cmd/bonsai/main.go` + `├── embed.go ← root embed package (CatalogFS, GuideContent)` and "embedded via embed.FS in `embed.go`". Trivial fix; surfaced post-merge. *(added 2026-04-20, source: session — found during Plan 16 close-out)*
- **[debt] Stale agent worktrees accumulating under `.claude/worktrees/`** `[housekeeping]` — `git worktree list` shows ~15 leftover worktrees from prior plan dispatches (some on UNC `//wsl.localhost/...` paths from cross-OS sessions, some on Linux paths). Branches mostly already merged. The UNC ones can't be removed from this side; the Linux-side ones are safely prunable. Risk: confused state when re-dispatching to a worktree-named branch, occasional `branch -D` failures (hit during PR #23 merge). Suggested: add a station routine or `bonsai` chore to prune merged worktrees periodically, plus a one-time manual sweep. *(added 2026-04-20, source: session — found during Plan 16 close-out)*
- **[bug] Installed sensor scripts have CRLF line endings, breaking bash hooks** `[Group F]` — All 8 sensor scripts in `station/agent/Sensors/*.sh` were installed with CRLF line terminators despite the catalog source (`catalog/sensors/*/*.sh.tmpl`) being LF-only. Symptom: `Stop` hook etc. fail with `$'\r': command not found` and `syntax error: unexpected end of file`. Workaround applied this session: `sed -i 's/\r$//'` on the installed copies. **Investigation needed:** (1) does the generator's template renderer emit CRLF on some platforms? (2) is `git config core.autocrlf` rewriting on checkout? (3) was it a one-off from a manual edit through a Windows tool? **Fix options:** force LF in the generator's file-write path regardless of host platform; add a generation-time test that asserts emitted `.sh` files have no `\r`; add a `.gitattributes` rule pinning `*.sh` and `*.sh.tmpl` to `text eol=lf`. *(added 2026-04-17, source: session — Stop hook error in Plan 15 work)*
<!-- "Review panel borders break in non-fullscreen terminals" — fixed 2026-04-17 via width-aware TitledPanel (ansi.Truncate + term.GetSize) -->
<!-- "Explicit feedback for required-only sections" — fixed 2026-04-17 via collapsed chip line in PickItems -->
<!-- Implicit fix: Getwd error now surfaces via FatalPanel instead of becoming a confusing "open .bonsai.yaml: no such file or directory" — see mustCwd() in cmd/root.go -->


## P2 — Medium

### Group A: Documentation Suite

> Resolves Roadmap Phase 1 "Usage instructions". The multi-topic command is the delivery mechanism; the three docs are the content. Ship incrementally — content first, CLI wiring last.

- **[feature] Quickstart guide** — Write `docs/quickstart.md` as a post-install walkthrough: what `bonsai init` generated, what to read first, how to add your first code agent, running your first session with the Tech Lead, understanding the generated CLAUDE.md, and when to run routines. Distinct from the README quick start (which is 2 commands) — this is the "now what?" guide for after installation. *(added 2026-04-16, source: user)*
- **[feature] Concepts guide** — Write `docs/concepts.md` explaining Bonsai's mental model for newcomers: station vs workspace, the 6-layer instruction stack, agents as team members, sensors as automated enforcement, routines as self-maintenance, the Playbook as project state, and how everything connects. Extract and reorganize content from HANDBOOK.md into a standalone conceptual overview aimed at someone evaluating whether to adopt Bonsai. *(added 2026-04-16, source: user)*
- **[feature] CLI usage guide** — Write `docs/cli-usage.md` covering every command in depth: `init` walkthrough (what each prompt means, scaffolding choices, agent defaults), `add` (component selection, compatibility filtering, what gets generated), `remove` (clean vs file-preserving removal), `update` (conflict resolution flow, custom file detection), `list` and `catalog` (reading the output). Include practical scenarios: first-time setup, adding a second agent, upgrading after a Bonsai version bump, recovering from a bad update. Link from README guides table and `bonsai guide` multi-topic command. *(added 2026-04-16, source: user)*
- **[feature] `bonsai guide` multi-topic command** — Expand `bonsai guide` from a single-doc renderer into a multi-topic CLI guide with an interactive Huh topic picker. Topics: **quickstart**, **concepts**, **catalog**, **custom-files** (existing). Implementation: each topic is a separate `docs/{topic}.md` file; `cmd/guide.go` adds a Huh select form when called without args; `bonsai guide <topic>` skips the picker. *(added 2026-04-16, source: plan-05 split)*

### Group B: Code Quality & Testing

> Logical ordering: split the big file first (makes testing easier), then add tests, then fix error handling. The two P1 bugs (triggerSection frontmatter, spinner error swallowing) can be fixed independently at any time.

- **[debt] Break up `generate.go` — 1,357 lines, highest churn file** — `internal/generate/generate.go` is both the largest Go file and the most frequently modified. It handles file writing, template rendering, conflict resolution, lock management, sensor/routine wiring, and scaffolding — too many responsibilities in one file. Split along natural seams: (1) template rendering, (2) file writing + conflict resolution, (3) lock management, (4) sensor/routine wiring. Would improve testability and reduce merge friction for agent dispatches. *(added 2026-04-16, source: repo-analytics)*
- **[debt] `internal/catalog/` test coverage — 496 lines, 0%** — Catalog loading (`LoadCatalog()`, `DisplayNameFrom()`, meta.yaml parsing) is the bridge between embedded YAML and the rest of the system. A malformed `meta.yaml` in the catalog would break at runtime with no test to catch it. Basic tests for catalog loading, display name derivation, and agent compatibility filtering would catch regressions cheaply. *(added 2026-04-16, source: repo-analytics)*
- **[debt] CLI command test coverage — `cmd/` package at 0%** — The `cmd/` package contains all user-facing CLI logic (init, add, remove, update, list, catalog, guide) — 1,691 lines across 8 files, zero tests. Priority targets: (1) `cmd/init.go` — happy path e2e test (temp dir, verify output structure), (2) `cmd/add.go` — test that abilities land correctly, (3) `cmd/remove.go` — test clean removal (472 lines, 4th largest file). Table-driven tests with temp dir setup would cover the most ground. *(added 2026-04-16, source: repo-analytics)*
- **[debt] PTY smoke test for harness-driven CLI commands** — `internal/tui/harness/` reducer tests are TTY-free (`fakeStep` + message injection) which catches logic bugs but can't drive a real `bonsai init`/`add`/`remove`/`update` end-to-end. Add a PTY-based smoke test using `creack/pty` or similar: spawn the built binary, send scripted keystrokes, assert the post-exit filesystem state (config written, workspace generated, lockfile valid). Would catch regressions unit tests miss — huh state transitions, AltScreen entry/exit, embedded form focus. Scope covers iter 1's `bonsai init` + iter 2's `bonsai add` + iter 3's `remove`/`update`. *(added 2026-04-20, source: Plan 15 iter 1 report — out-of-scope followup)*
<!-- "ActionUnchanged test coverage gaps" promoted to Status.md as Plan 13 — removed 2026-04-17 -->
<!-- "writeFileChmod skips chmod on ActionUnchanged" promoted to Status.md as Plan 13 — removed 2026-04-17 -->

### Group C: OSS Readiness

> All support the public repo being contributor-friendly. Small, independent — could knock both out in one session.

- **[improvement] OSS polish — demo GIF/asciinema for README** — Last remaining OSS readiness item after Plan 17. Linter config (`.golangci.yml` with errcheck/govet/unused/misspell/gofmt/goimports) and Makefile `test`/`lint`/`fmt`/`tidy` targets shipped in PR #24. Still need a demo GIF or asciinema recording to add under README hero image — requires user recording (not agent-able). When recording: show `bonsai init` flow, add a skill, run `bonsai list`. *(added 2026-04-16, narrowed 2026-04-17 post-Plan-17, source: RESEARCH-oss-readiness.md cleanup)*
- **[improvement] Seed GitHub Issues for contributor on-ramp** — The repo is public with community health files and a polished README, but has 0 issues, 0 stars, 0 forks — no entry point for potential contributors. File 3-5 well-scoped issues labeled `good first issue` (e.g., add `test`/`lint` Makefile targets, add `.golangci.yml`, add catalog tests). Also consider adding a `help wanted` label for medium-complexity items. *(added 2026-04-16, source: repo-analytics)*

### Group D: Catalog Expansion

> Research first (concept-decisions), then build. The concept-decisions review informs which of the others to prioritize.

- **[research] Revisit concept-decisions research** — Review `station/Research/concept-decisions.md` for unbuilt concepts that may be worth promoting: (1) **Talents** — a new catalog category for innate behavioral aptitudes, (2) **Meta-layer** — runtime system-wide observation layer, (3) **Three-layer catalog ownership model**, (4) **Loading gradient** reasoning. Decide which to build, which to backlog properly, which to discard. *(added 2026-04-16, source: research doc cleanup)*
- **[feature] Unbuilt catalog items — 3 agents, 1 skill, 4 routines** — From the catalog expansion research, 8 items were never built: **agents** `qa`, `reviewer`, `docs`; **skill** `documentation-standards` (blocks `docs` agent); **routines** `test-coverage-check` (qa), `changelog-maintenance` (docs), `api-docs-drift` (docs), `standards-drift` (reviewer). Build order: `documentation-standards` skill first (unblocks `docs`), then agents, then routines. *(added 2026-04-16, source: RESEARCH-catalog-expansion.md cleanup)*
- **[feature] Changelog generation skill + release changelogs** — Add a changelog generation skill that: (1) parses conventional commit messages between tags to generate structured changelogs, (2) outputs CHANGELOG.md, (3) generates release notes for `gh release create --notes`. Current releases (v0.1.0-v0.1.3) shipped with no changelogs — backfill them. Also consider a `bonsai changelog` CLI command. *(added 2026-04-16, source: user)*
- **[feature] Research scaffolding item + abilities** — Add an optional `Research/` folder to project scaffolding for storing landscape analysis, concept decisions, and design research. Add associated abilities (tech-lead only): a research workflow and/or a research-template skill. *(added 2026-04-16, source: user)*

### Group E: Workspace Improvements

> Small, independent quality-of-life items. Can be done in any order.

- **[improvement] Plan archiving — Active/Archive folder structure** — Plans currently all live in `Plans/Active/`. Completed plans should move to `Plans/Archive/` after merge. Requires: create `Plans/Archive/` in scaffolding manifest, update issue-to-implementation workflow (Phase 10), update planning workflow and planning-template skill, update session-start protocol if it scans for active plans, update CLAUDE.md nav table. *(added 2026-04-16, source: user)*
- **[improvement] Consolidate FieldNotes usage** — The current `Logs/FieldNotes.md` file has unclear purpose and overlaps with other state files (memory.md, Status.md, KeyDecisionLog.md). Rethink what it's for, whether it should be merged into another artifact, and how it fits into the session-start context injection. *(added 2026-04-15, source: user)*
- **[improvement] Post-update backup merge hint** — After `bonsai update` creates `.bak` backups during conflict resolution, print a hint telling the user to ask their agent to reconcile customizations. Small change to `cmd/update.go` after `resolveConflicts()` returns. *(added 2026-04-16, source: user)*

### Group F: UI/UX Testing Findings

> Dogfooding findings from the 2026-04-17 `bonsai init` walk-through. Mostly `init` flow polish — visual identity, prompt flow ergonomics, and the review→generate→complete flow. Tackle as a coherent design push rather than piecemeal to avoid inconsistency. Start with the palette (prerequisite), then visual elements can consume it.

- **[improvement] Define a canonical color palette for the whole TUI** — The current coloring feels ad-hoc and doesn't establish a visual identity. Define a palette in `internal/tui/styles.go` with named tokens (primary, accent, muted, success, warning, error, border, hint) and refactor all LipGloss styles to reference tokens instead of inline hex. Prerequisite for all other visual-identity items below. *(added 2026-04-17, source: session — UI/UX testing)*
- **[improvement] Redesign the "B O N S A I / agent scaffolder" banner** — Current banner feels unprofessional and wastes vertical space without conveying useful info. Redesign with richer content: version number, tagline, maybe project context if available (e.g., "initializing new project" vs "managing existing"). More polish, less "hello world ASCII art" feel. *(added 2026-04-17, source: session — UI/UX testing)*
- **[improvement] Sleeker input cursor + tighter prompt-label/input grouping** — In prompts like `Project name: / >`, the label and input feel stacked together rather than cleanly grouped. The default typing cursor is also plain. Upgrade to a sleeker cursor style (e.g., Huh's theme options) and adjust indentation/spacing so prompts read as one tight unit. *(added 2026-04-17, source: session — UI/UX testing)*
- **[improvement] Persist answered prompt values on screen through multi-step init** — After typing the project name and moving to the description prompt, the project name disappears. Users lose track of what's been entered. Each step should render a running summary of prior answers so the user can see cumulative progress. *(added 2026-04-17, source: session — UI/UX testing)*
- **[feature] Go-back navigation in multi-step init flow** — No way to correct a mistake mid-flow (e.g., typo in project name after advancing). Add a way to step back to prior prompts — either a key binding (Esc/Shift-Tab) or an explicit "back" option at each step. Huh supports this via form navigation but it's not wired up currently. *(added 2026-04-17, source: session — UI/UX testing)*
- **[improvement] Progressive disclosure for project scaffolding step** — The scaffolding + agent + abilities selection currently dumps all text at once, overwhelming the user. Break into layered chunks: scaffolding first (with explanation), then agent type (with explanation), then abilities per category. Consider collapsed-by-default sections, tabs, or sequential focused screens instead of one long form. *(added 2026-04-17, source: session — UI/UX testing)*
<!-- "Explicit feedback when required-only section auto-accepted" — fixed 2026-04-17 via collapsed chip line in PickItems -->
- **[improvement] Show counts alongside ability category headers** — The review panel shows categories like "Skills" then lists items. Add counts to the header: "Skills (3)", "Workflows (5)", "Protocols (4 required)". Gives the user a quick at-a-glance read on the shape of their selection. *(added 2026-04-17, source: session — UI/UX testing)* — *(partial: ItemTree now shows `(N)` counts; reconsider if further category polish wanted)*
- **[improvement] TUI screen lifecycle — clear prior step output on major transitions** — When the review panel appears, all prior step output is still visible above it (banner, scaffolding list, agent intro, ability categories). Same problem after clicking "generate" — the review and generate confirmation stay onscreen above the success message. Screens should "take over" on major transitions: review clears prior content, generate clears the review, success clears the generate. Probably means using Bubbletea's `AltScreen` or explicit clear/redraw on step transitions. *(added 2026-04-17, source: session — UI/UX testing)*
- **[improvement] Modernize the review → generate → complete flow** — Three related touchpoints all feel dated: (1) **File structure preview** is plain text but is the centerpiece of the review — deserves a proper tree visualization with icons/colors, (2) **Generate Project confirmation** feels like a legacy yes/no prompt — modernize as a focused confirm panel with context (what will be written, where, how many files), (3) **Post-generate "next steps" hints** are cramped and terse — deserve their own space with verbose, inviting guidance on what to do next (cd into project, open CLAUDE.md, run first session, etc.). Likely one coherent design pass touching all three. *(added 2026-04-17, source: session — UI/UX testing)*
- **[bug] Spinner Ctrl-C partial-write window (init + add)** — Both `cmd/init.go` and `cmd/add.go` run `huh/spinner` → `generate.*` calls *outside* the BubbleTea harness (after `harness.Run` exits AltScreen). Ctrl-C during generation can leave partial files on disk with no cleanup — the harness's clean-exit guarantee only covers the interactive portion. Plan 15 iter 3 introduces a `SpinnerStep` adapter that folds generation into the program so a single cancellation path covers the whole flow; this entry is a waypoint so the issue isn't forgotten if iter 3 slips. *(added 2026-04-20, source: Plan 15 iter 2.1 review)*
- **[improvement] Workspace validator normalization** — `workspaceUniqueValidator` in `cmd/add.go` compares raw user-typed input against existing workspace paths, so `backend`, `backend/`, and `./backend/` all treat as distinct even though they resolve to the same directory. Normalize before compare (clean + trailing-slash strip + relative-path resolution). Low-impact in practice since the post-harness pipeline also calls `filepath.Clean`, but the validator should match. *(added 2026-04-20, source: Plan 15 iter 2.1 review)*
- **[improvement] Panic recovery around harness Splice/Build** — A `LazyStep` or `LazyGroup` builder that panics today propagates up through `tea.Program` and crashes the whole CLI with a stacktrace mid-AltScreen (often leaving the terminal in a weird state). Wrap the `Splice()`/`Build()` invocations in `expandSplicer` / the lazy-entry path with `recover()` and convert to a harness-level error that exits AltScreen cleanly and prints a `tui.FatalPanel`. *(added 2026-04-20, source: Plan 15 iter 2.1 review)*
- **[improvement] `ConditionalStep` adapter for empty-picker skip** — `buildAddItemsSteps` in `cmd/add.go` manually filters zero-item categories out of the step list; the logic belongs in a reusable harness adapter (`ConditionalStep`) that wraps another step and `Done()`s immediately with an empty result when its predicate returns false. Would also simplify iter 3's `update` flow where custom-file pickers per agent may be empty. *(added 2026-04-20, source: Plan 15 iter 2.1 review)*
- **[docs] Document AltScreen behavior change in release notes** — Plan 15 migrates `bonsai init` and `bonsai add` into AltScreen, which means the interactive flow no longer accretes to scrollback line-by-line — users who rely on copy/pasting partially-filled screens out of their terminal history will notice. Add a release-note bullet when iter 3 ships and the whole branch merges. *(added 2026-04-20, source: Plan 15 iter 2.1 review)*
- **[docs] Fill "Deviations from Plan" in completion reports more aggressively** — The iter-2 completion report at `Reports/Archive/2026-04-20-plan-15-iter-2-add-migration.md` listed three deviations but missed noting that the iter-2 "pre-harness tech-lead gate" design choice in the plan was itself a regression (fixed in 2.1). Implementing agents should err toward over-documenting plan↔implementation divergence, since the gaps are where post-ship reviewers find bugs. Tweak `planning-template` or `review-checklist` skill to prompt for this. *(added 2026-04-20, source: Plan 15 iter 2.1 review)*

### Ungrouped P2

- **[feature] Developer guide for Bonsai contributors** — Write a `DEVELOPMENT.md` (or docs site page) covering the internal dev workflow: how to build and test locally, `npm run generate:catalog` usage and when to run it, release checklist, catalog structure conventions, testing against a temp dir, and other commands/processes a contributor needs to know. Currently this info is scattered across CLAUDE.md and tribal knowledge. *(added 2026-04-17, source: user)*
- **[feature] Routine report template** — Add a `routine-report-template.md` to `station/Reports/` alongside the existing `report-template.md`. Routine reports have a different shape than plan completion reports. The template in `loop.md` defines the format; this makes it a first-class project artifact. *(added 2026-04-14, source: user)*
- **[improvement] Split design-guide: generic catalog skill + Bonsai-specific station override** — Plan 11 replaces `catalog/skills/design-guide` with Bonsai-specific TUI/CLI rules (paths like `internal/tui/**`, palette tokens). That's useful for dogfooding but irrelevant when external users install the skill. Follow-up: restore a generic Charm/Go CLI design-guide to the catalog (palette patterns, NO_COLOR support, panel vocabulary as principles — not specific paths), and move Bonsai's own rules into `station/agent/Skills/design-guide.md` as a local override. *(added 2026-04-17, source: plan-11 audit)*
- **[improvement] Install semgrep and/or gitleaks for better scanning** — Vulnerability scan and secrets scan routines currently use manual pattern-based Grep scanning. Installing semgrep (SAST) and/or gitleaks (secrets) would improve coverage and reduce false negatives. *(added 2026-04-16, source: routine-digest)*

## P3 — Ideas & Research

### Future Platform (Roadmap Phase 2+)

- **[feature] Integration scaffolding variants** — Support alternative backends for all PM artifacts (backlog, status, roadmap, reports). During `bonsai init`, user picks a backend per artifact: markdown (default), GitHub Issues, Notion, Jira, etc. Affects: scaffolding manifest, agent instructions, protocols, any sensor/workflow that references PM files. *(added 2026-04-15, source: user)*
- **[feature] Enhanced session-start sensor — project pulse** — Expand `session-context.sh` to inject a project-state summary at session start. Phase 1: check markdown PM files directly. Phase 2: auto-detect from integration variant in `.bonsai.yaml`. *Depends on: integration scaffolding variants.* *(added 2026-04-15, source: user)*
- **[feature] Custom item creator** — Interactive TUI for creating custom items (skill, workflow, protocol, sensor, routine) with frontmatter scaffolding — similar to Claude's skill creator. *(added 2026-04-14, source: user)*
- **[improvement] Self-update mechanism** — Skills and workflows should be able to self-heal or flag when they have issues. *(added 2026-04-13, source: user)*
- **[improvement] Micro-task fast path** — Define an explicit lightweight protocol for trivial changes (< 50 LOC, no architectural impact). Could be a sensor that auto-detects task weight, or a protocol clause agents check before entering full planning mode. *(added 2026-04-15, source: architectural audit)*

### Routine System Enhancements

- **[feature] Scheduled task generation for routines** — Auto-generate Claude scheduled task configs from routine metadata. Maps routine `frequency` to cron expressions. Two tiers: file-only routines → Desktop local tasks; heavier routines → Cloud routines. *Note: Cloud routines are in research preview — API/limits may change.* *(added 2026-04-15, source: user)*
- **[feature] Routine GitHub issue creation** — Routines can create GitHub issues for actionable findings. Per-routine opt-in via `creates_issues: true` in `meta.yaml`. Issues get a `bonsai-routine` label. Requires `gh` CLI. *Dependency: prerequisite for auto-fixer routine.* *(added 2026-04-15, source: user)*
- **[feature] Auto-fixer routine** — New routine that polls GitHub issues labeled `bonsai-routine`, attempts autonomous fixes. *Depends on: routine GitHub issue creation + scheduled task generation.* *(added 2026-04-15, source: user)*

### Outreach

- **[feature] README case study / blog post from session-transcript metrics** — The 2026-04-16 transcript analysis (archived at `station/Reports/Archive/2026-04-16-session-transcript-analysis.md`) contains compelling quantitative data: 20 sessions over 6 days, ~1,186 user messages (~90% silent tool approvals), ~2,000 substantive words total drove an entire CLI tool from Go rewrite through OSS release. Specific hooks: "75-message silent approval streak," "10+ deliverables in 48 minutes," "14 of 20 sessions had zero user-initiated rework." Could seed a README "Real-World Usage" section, a standalone case study page on the docs site, or a blog post. Parts 5 (session typology) and 7 (metrics) are the primary source material. *(added 2026-04-17, source: session — pending-report review)*

### Research

- **[research] Session-start payload optimization** — Investigate whether the session-context sensor payload can be made leaner. Current payload is ~600-700 lines — could free ~200-300 tokens with formatting cleanup. Low priority since layered loading already defers most content. *(added 2026-04-15, source: architectural audit)*
- **[research] Parallel agent coordination in shared repos** — Research how multiple code agents can work simultaneously on different tasks in the same repository. Key questions: git workflow, file contention, lock/claim protocol, orchestration model, state coherence, tooling. *(added 2026-04-16, source: user)*
- **[research] Archon analysis** — <https://github.com/coleam00/Archon> — research what it does, use cases, overlap with Bonsai, what we can learn. *(added 2026-04-13, source: user)*

### Big Bets

- **[feature] Managed Agents integration** — Cloud deployment via `bonsai deploy`, session management, outcome rubrics in catalog. Build after local foundation is stable. *(added 2026-04-13, source: user)*
- **[feature] Greenhouse companion app** — Desktop app for managing projects + observing AI agents. Design doc: DESIGN-companion-app.md. Stack: Tauri v2 + Svelte 5 + SQLite. Status: Design phase, decisions locked. *(added 2026-04-13, source: user)*
- **[improvement] Catalog display_name audit** — Add explicit `display_name` to all catalog `meta.yaml` files. Research other metadata fields that could be useful (e.g., `version`, `tags`, `dependencies`, `examples`). *(added 2026-04-14, source: user)*

---

## Removed Items

<!-- Items resolved or removed during backlog hygiene. Keep for audit trail. -->
<!-- "Case-insensitive file collision" fixed — removed 2026-04-16, issue-to-implementation workflow, PR #8 -->
<!-- "Code index line number drift" fixed — removed 2026-04-16, issue-to-implementation workflow, PR #12 -->
<!-- "CI workflow + branch protection" implemented — removed 2026-04-15, issue-to-implementation workflow, PR #6 -->
<!-- "Release pipeline" implemented — removed 2026-04-15, issue-to-implementation workflow, PR #5 -->
<!-- "Better trigger sections" promoted to Status.md Pending — removed 2026-04-14, backlog-hygiene routine -->
<!-- "Selective file update" implemented — removed 2026-04-15, issue-to-implementation workflow -->
<!-- "Doubled path prefix" fixed — removed 2026-04-15, issue-to-implementation workflow -->
<!-- "Workspace artifact sync" — marker migration implemented in PR #1, removed 2026-04-15, issue-to-implementation workflow -->
<!-- "Rename catalog items to abilities" implemented — removed 2026-04-15, issue-to-implementation workflow, PR #2 -->
<!-- "Custom item detection" completed and shipped as bonsai update — removed 2026-04-14, backlog-hygiene routine -->
<!-- "UI overhaul" promoted to Status.md Pending — removed 2026-04-14, backlog-hygiene routine -->
<!-- "Usage instructions" partially implemented as AI operational intelligence (Plan 05, PR pending) — split to guide items in Group A -->
<!-- "Human-AI interaction guide" implemented as docs/working-with-agents.md — removed 2026-04-16, session work -->
<!-- "bonsai guide command" implemented — removed 2026-04-15, issue-to-implementation workflow, PR #3 -->
<!-- "Community health files" implemented — removed 2026-04-16, issue-to-implementation workflow, PR #9 -->
<!-- "Routine report digest" implemented as custom workflow agent/Workflows/routine-digest.md — removed 2026-04-16, manual creation -->
<!-- "Clean up stale remote branches" — done 2026-04-16, backlog hygiene (deleted 8 merged branches, enabled prune) -->
<!-- "Remove infra-drift-check routine" — done 2026-04-16, backlog hygiene (no cloud infra to check) -->
<!-- "Consolidate Usage instructions roadmap item" — resolved by grouping guide items into Group A, 2026-04-16 -->
<!-- "ActionUnchanged test coverage gaps" + "writeFileChmod ActionUnchanged bug" — promoted to Status.md as Plan 13, 2026-04-17 -->

---

## Group Index

| Group | Theme | Phase Order | Notes |
|-------|-------|-------------|-------|
| **A** | Documentation Suite | Quickstart → Concepts → CLI Usage → Multi-topic command | Resolves Roadmap "Usage instructions". Content first, CLI wiring last. |
| **B** | Code Quality & Testing | Split generate.go → catalog tests → cmd tests → trigger test infra → spinner error fix | P1 bugs (frontmatter, spinners) can be fixed independently at any time. |
| **C** | OSS Readiness | Linter + Makefile → seed GitHub Issues | Small, one-session effort. |
| **D** | Catalog Expansion | Concept-decisions review → documentation-standards skill → agents → routines → changelog | Research informs build order. |
| **E** | Workspace Improvements | Any order | Independent quality-of-life items. |
| **F** | UI/UX Testing | Any order | Findings from dogfooding session on 2026-04-17 — CLI polish, install UX, prompt flow issues. |
