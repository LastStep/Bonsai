---
tags: [core, memory]
description: Tech Lead Agent working memory — flags, work state, notes.
---

# Working Memory

## Flags

<!-- Active flags go here. Format: - [FLAG] description -->

(none)

## Work State

**Current task:** Idle. Plan 24 (pre-launch polish bundle) shipped 2026-04-22 via PR #58 squash `4ef8271`. Pre-OSS-launch punch list now: (remaining) demo GIF/asciinema for README hero (user-recording, not agent-able), pre-release docs audit (user-flagged as final gate before announce), 5 good-first-issues now open as contributor on-ramp.
**Main at:** `4ef8271` (Plan 24 merge). Working tree has stashed Plan 23 WIP — `stash@{0}` contains internal/tui/initflow/{chrome,design,enso,stage}.go edits + internal/tui/addflow/ (new) + Backlog/StatusArchive edits for Plan 23 UI/UX Phase 2 "port bonsai add to cinematic flow". Parallel session's track; don't cross-pollinate.
**In flight (other tracks):** Plan 23 (UI/UX Phase 2 — port bonsai add to cinematic flow) — WIP in stash@{0}, parallel session owns it.
**Blocked on:** Nothing.
**Loose ends (uncommitted):** Plan 23 WIP in stash@{0} (not mine to commit).
**Last completed:** 2026-04-22 Plan 24 — pre-launch polish bundle (PR #58 squash `4ef8271`). Ships: (A) `CHANGELOG.md` keep-a-changelog 1.1.0 curated v0.1.0–v0.1.3 backfill + `[Unreleased]` stub + link refs; (B) `.github/workflows/docs.yml` adds `pull_request` trigger + `if: github.event_name == 'push'` guards on deploy job + upload-pages-artifact step (Option 1 — same "Deploy Docs" failure surface for PR and main); (C) root `Bonsai/CLAUDE.md` `internal/tui/` tree refresh (styles_test.go, filetree*, harness/, initflow/ — subdirs only, no per-file enumeration); (D) Backlog Group C/D consolidation (2 entries removed + removal comments, Group D suffixed as "refiled as good-first-issue"); (E) 5 GitHub issues filed (#53 relabeled + #54/#55/#56/#57 new) all `good first issue` + `help wanted`. CI gotcha: pre-existing gofmt drift in `internal/tui/initflow/observe.go:466` + `planted.go:423` (from earlier direct-to-main polish) caught on PR lint — fixed on PR branch via `gofmt -s -w` before merge. Prior: 2026-04-22 post-Plan-22 12-point dogfood polish + statusLine redesign (commits `018966d` `4a8fea9` `975e15d`, all direct-to-main). Prior: 2026-04-21 Plan 22 Phase 5B (PR #52 squash `5916e05`) — wired `NewConditional(NewLazy(GenerateStage))` + LazyGroup conflicts + `NewConditional(NewLazy(PlantedStage))` into `runInit` at `cmd/init_redesign.go`; renamed `runInitRedesign` → `runInit`; deleted legacy 245-line `runInit` body + `buildReviewPanel` + `scaffoldingOptions` + `BONSAI_REDESIGN` env branch from `cmd/init.go`. Harness additions: `LazyStep.Chromeless()` + `ConditionalStep.Chromeless()` delegation, `ConditionalStep.Init` auto-builds nested LazyStep — necessary because `GenerateAction = func() error` needs prev-capture at build time (plan's literal inline ctor wouldn't fire). Split single `plantedConfirmed` predicate from plan into `observeConfirmed(prev[3] bool)` + `generateSucceeded(observeConfirmed && prev[4] non-error)` so Generate-failure skips Planted. `buildGenerateAction` closes over 10 params; `plantedSummary(installed)` lazy so post-`EnsureRoutineCheckSensor` counts render. +260/−297 across 8 files. 6/6 CI green. Independent review PASS + 2 minors (composition test + dead-path warning — filed to Backlog). Prior 2026-04-21: Plan 22 Phase 5A PR #51 squash `6baaf8e` + lint fix `3e31967` — responsive resize foundation (`MinTerminalWidth=70`, `ClampColumns(120)=(24,44,12)` regression anchor, hand-rolled Viewport, `RenderMinSizeFloor`), Branches/Vessel/Soil retrofit, ObserveStage wired, GenerateStage + PlantedStage packaged. Prior: 2026-04-21 Phase 4 dogfood polish run (direct-to-main, no PRs per fast-iter UX convention): `413e360` header left block split (`[ 盆 ]` row + INITIALIZE row) + per-tab 2-line intro + DETAILS box moved below list (fixed-height, no jitter) + right-aligned DEFAULT tag; `399fe08` density cut (row 60 cells) + word-wrapped details (2 rows) + centered tab counts via `lipgloss.PlaceHorizontal(colW, Center, ...)` ; `eaee416` widened everything back (nameColW 24, descColW 44, tab colW 16, row 84 cells) + rune-aware name truncation so DEFAULT/(required) always stays in its rightmost column; `6bb74e5` ABOUT/FILE values → `ColorAccent` (white) + 3-row wrap × 70 cells (210 absorbs every catalog description; dispatch-guard ~111 chars previously clipped); `fa0ae64` kanji centered via `[ 盆 ]` padding + extra blank line between DETAILS and counter. Helper `wrapToWidth` (word-break, rune-fallback hard-wrap) added to `branches.go`. Prior: PR #50 `89c21ba` — Plan 22 Phase 4: `BranchesStage` (branches.go 582L, branches_test.go 380L) — 5-tab picker across Skills/Workflows/Protocols/Sensors/Routines with per-category `branchCat` + `branchItem`; `← → / h l` tab cycle wraps; `↑ ↓ / j k` focus clamp no-wrap; `␣` toggle (no-op on required); `?` global `expanded` flag toggles inline-expand on focused row; `↵` advance; `esc/shift+tab` propagate. `BranchesResult{Skills,Workflows,Protocols,Sensors,Routines []string}` shape; catalog-sorted machine names; required items always in Result. Constructor `NewBranchesStage(ctx, cat, agentDef)` — builds `[]branchCat` internally via `cat.SkillsFor`/`WorkflowsFor`/`ProtocolsFor`/`SensorsFor`/`RoutinesFor("tech-lead")`; required via `item.Required.CompatibleWith("tech-lead")`; defaults from `agentDef.DefaultSkills`/etc. Inline-expand renders ABOUT + FILE only (catalog lacks `Affects`/`CrossLinks` today — fields reserved on struct for future expansion per plan note). Kanji 技/流/律/感/習 with ASCII fallback. Focus: leaf `│ ` border-left, no forced bg (Soil precedent). `Reset()` clears `done` only — preserves `categories`/`catIdx`/`itemIdx`/`selected`/`expanded` (plan §verification explicit). Wire: `cmd/init_redesign.go:78` stub→real; Observe at idx=3 still stub. 12 tests: tab cycle both dirs with wrap, focus clamp, toggle non-required, required no-op, `?` expand, Result per-tab, Result catalog order, Enter completes, Reset preservation, cross-tab selection persistence, defaults applied, required always in Result. Independent review PASS on all 10 locked decisions + 7 keybindings. Net 963 adds / 1 del across 3 files. CI green (test/Analyze Go/lint/govulncheck/CodeQL/GitGuardian). Process: pushed local-only `380dbc5` (Phase 3.5) to origin/main pre-dispatch so agent worktree saw current palette — per memory's "Agent worktrees base off origin/main" gotcha. Prior: PR #49 `971ee44` — Plan 22 Phase 3: `VesselStage` + `SoilStage` + header strip. Prior: PR #48 `2e2a08c` — Plan 22 Phase 2: initflow package + `harness.Chromeless` + env-flag routing. Prior: PR #47 `7553d43` — Plan 22 Phase 1: `RenderFileTree` widget + palette tokens.

## Notes

<!-- Session-to-session notes. Keep concise. -->

- **Plan 14 code shipped to main via PR #24 bundle (2026-04-17).** Iterations 1 + 2 (palette tokens, banner, answered-prompt persistence, required-only feedback, category counts, prompt polish, mustCwd error surfacing, width-aware TitledPanel) merged along with Plan 17's release-prep work. Phase 4+ scope (screen lifecycle, progressive disclosure, go-back nav, flow redesign) still deferred — Plan 15's BubbleTea harness is the likely vehicle for those.
- **Parallel session convention.** `ui-ux-testing` branch is where Plan 15 iterations land (separate session). Don't cross-pollinate planning-doc edits between branches — each branch's Status.md/memory.md should stay scoped to its own track. Merging to main collapses everything cleanly.
- **CI lint gotcha.** `golangci/golangci-lint-action@v6` with `version: latest` resolves to v1.x, not v2. `.golangci.yml` must use v1 schema (`linters.disable-all: true`, no `version:`/`formatters:` keys). If we ever move to v2, update both the action pin (`version: v2.x`) and the config schema together. Learned from PR #24 `66a6304`.
- **Plan 08 Phase C (new sensors) paused** — moved back to Pending while Plan 14 ships. Resume once UI/UX overhaul series wraps or explicitly requested.
- **Pre-flight learning:** Worktrees inherit only committed HEAD — uncommitted plans/docs in main tree are invisible to dispatched agents. Commit station/ planning artifacts before dispatch.
- **PR review memory hygiene:** "both reviews APPROVE" from prior session was dispatched review agents, not GitHub reviews. `gh pr view --json reviews` returned empty. When noting review status, distinguish agent-dispatched reviews (in `Reports/`) from GitHub formal reviews.
- **Squash-bundle rebase pattern.** When main has absorbed a feature branch's commits via squash-merge (e.g., Plan 14 individual commits → bundled into Plan 17's `bc565bf` squash), straight `git rebase main feature-branch` will conflict on every doc snapshot because main has moved past those states even though the *code* matches. Cleaner approach: `git reset --hard main && git cherry-pick <only-the-truly-new-commits>` after listing each commit and asking "is this code/content already on main via the squash?" Always preserve old tip as `<branch>-pre-rebase` first. Used 2026-04-20 to rebase `ui-ux-testing` (13 commits → 3 commits actually carried forward).
- **Worktree cwd gotcha from Claude Code.** `git worktree add ../<name> <branch>` runs in the bash tool's cwd, which is often a subdirectory (for us: `station/`), not the repo root. So `../Bonsai-uiux` resolves to `/home/rohan/ZenGarden/Bonsai/Bonsai-uiux` (nested inside the main worktree — violates the no-nesting rule), not the intended sibling `/home/rohan/ZenGarden/Bonsai-uiux`. **Always use absolute paths when creating worktrees from this session.** Hit this 2026-04-20 setting up the Plan 15 parallel-session worktree; caught, removed, recreated at the correct sibling path.
- **Subagent tool inheritance is flaky.** On 2026-04-17, the original executing agent (worktree, `gh` authed) successfully created draft PR #23, but every subsequent subagent I spawned for merge/verification reported `gh: command not found` — the environment wasn't inherited from the spawning agent or from my own shell (Windows Git Bash over WSL UNC path, no `gh` on PATH). Implication: for PR-flow tasks, bundle **all** gh operations (push, create PR, mark ready, merge, delete branch) into a **single** agent dispatch rather than splitting across agents. Second-best: ask the user to click-merge via web UI when a subagent-less step is needed.
- **PR CI ≠ main CI.** The `CI` workflow (test + lint + GitGuardian) runs on pull_request, but `Deploy Docs` only runs on `push` to main. So a broken `website/**` change can pass PR CI and then fail on main — happened 2026-04-20 with PR #25's MDX autolink. When reviewing PRs that touch `website/`, run `cd website && npm run build` locally before merging, or add a `website/**` path trigger to Deploy Docs on pull_request (separate backlog item).
- **MDX autolink gotcha.** `<https://example.com>` is valid GitHub-flavored markdown but MDX parses `<` as JSX — Astro build fails with "Unexpected character `/`". Always use `[label](url)` inside `.mdx` files. Incident: PR #25 `guide.mdx:20`, fixed in `e336ccb`.
- **Local `go build` + `go test` miss `golangci-lint unused`.** Plan 15 iter 3 left `resolveConflicts` in `cmd/root.go` as dead code after the harness migration. Local verification (`make build`, `go test ./...`, `go vet ./...`) all passed; only the CI `lint` step (golangci-lint v1.64.8 with the `unused` linter) caught it. Local golangci-lint on this machine is v2.11.4 and the repo config is v1, so `golangci-lint run` errors out locally with "unsupported version of the configuration". Implication: for PRs that heavily refactor/delete, trust CI to catch dead code; OR install a v1 golangci-lint binary locally (e.g. `go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8`) for pre-push. PR #26 hit this; fixed in `070bcd0`.
- **Post-squash-merge local cleanup gotcha.** `gh pr merge --squash --delete-branch` successfully squashed to main and deleted the remote branch, but the local-branch-delete step failed with `'main' is already checked out at '/home/rohan/ZenGarden/Bonsai'` because the CLI tried to fast-forward the local sibling worktree's main and that worktree was on main. Remote branch deletion also failed silently in the same call — needed a follow-up `git push origin --delete ui-ux-testing`. Check `gh pr view --json state` + `git ls-remote origin | grep <branch>` after any squash-merge from a non-main worktree to confirm both halves completed. Confirmed again 2026-04-21 × 4 during Plan 20 (#28, #29, #31, #40, #41 all hit the same pattern) — always need manual `git push origin --delete <br>` + `git worktree remove -f -f <path>` + `git branch -D <br>`.
- **Claude Code `statusLine.command` runs with different `$PWD` than hook commands.** The project-wide walk-up-to-`.bonsai.yaml` wrapper used for hooks (`bash -c 'r="$PWD"; while [ "$r" != "/" ] && [ ! -f "$r/.bonsai.yaml" ]; do r=$(dirname "$r"); done; bash "$r/<sensor>.sh"'`) **fails silently** when used for statusLine — walk terminates at `/` with no match. Use an **absolute path at install time** for statusLine (`"command": "bash /abs/path/to/statusline.sh"`). Confirmed 2026-04-22 when debugging why the project-level statusLine wasn't rendering. Also: statusLine config isn't hot-reloaded — requires `/clear` or restart to pick up settings.json changes.
- **Subdirectory launch determines which `.claude/settings.json` is live.** When Claude Code is launched from `station/`, the effective project settings file is `station/.claude/settings.json`, not the repo-root `.../.claude/settings.json`. The repo-root one is ignored. Hit this 2026-04-22 while wiring the statusLine stanza — put it in the wrong file and spent a cycle diagnosing. Check `pwd` at the time of the Claude session start before deciding which `.claude/settings.json` to edit.
- **golangci-lint binary Go-version coupling.** The `version: X.Y` input on `golangci/golangci-lint-action` resolves to a specific pre-built binary. That binary must have been built with a Go version ≥ the repo's target. Moving a repo to Go 1.25 required golangci-lint v2.11.4+ (older v2.1 binary was still built on Go 1.24 and errors with `can't load config: the Go language version (go1.24) used to build golangci-lint is lower than the targeted Go version (1.25.8)`). Rule: when bumping Go major in go.mod, also pin golangci-lint-action to a version whose release notes say "built with Go ≥ target". Track latest v2.x: https://github.com/golangci/golangci-lint/releases. Hit this 2026-04-21 across PR #28 / #29 during Plan 20.
- **Agent worktrees base off origin/main, not local main.** `Agent(isolation: "worktree")` fetches origin/main as the worktree base — ANY uncommitted work on the local-only main branch is invisible to the agent and won't be in its PR. To ship local-only work, create a dedicated branch from the local tip, push, PR, merge first — THEN dispatch agents. Learned 2026-04-21 during Plan 20 pre-flight: my 65cfdd6 chore commit was on local main only; first Go-bump PR #28 came back clean (2 lines) precisely because agent worktree didn't see it. Had to ship as separate PR #30 before resuming.
- **Direct-to-main polish commits accumulate silent gofmt/lint drift.** Fast-iter UX convention ships polish commits direct-to-main without PR — but `ci.yml` only runs on `pull_request`, not `push`. So gofmt/lint regressions introduced by direct-to-main commits sit silently on main until the next PR (from any branch) trips over them. Hit 2026-04-22 on Plan 24 PR #58 lint: `observe.go:466` + `planted.go:423` had been unformatted on main since the 2026-04-22 dogfood polish run (commits `018966d`/`4a8fea9`/`975e15d`). Fix: run `gofmt -s -w ./...` locally before every direct-to-main polish batch, OR add `push: branches: [main]` trigger to CI (makes main runs visible), OR run a local-git-hook pre-commit gofmt. Also: when a PR lint-fails on files the PR didn't touch, first-instinct should be "is this pre-existing on main?" — saves a diagnose cycle.

## Feedback

<!-- User corrections and confirmed approaches that persist across sessions. Only record what isn't already in CLAUDE.md, workflows, or protocols. -->

- **Backlog.md at session start is a P0 *scan*, not a full read.**
    - **Why:** During the 2026-04-17 context audit, the CLI session-start payload was ~4k heavier than the app's, and the delta was traced to my having `Read` Backlog.md in full (~200 lines, ~4k tokens) when session-start protocol explicitly says "scan, look for P0 items only." Full backlog review is the backlog-hygiene routine's job. On a 200k window that's ~4% of headroom burned for zero benefit.
    - **How to apply:** At session start, use `Grep -n "^## P0" Playbook/Backlog.md` (or read with `limit:` bracketed to the P0 section) instead of `Read` on the whole file. Only full-read Backlog when explicitly running backlog-hygiene or planning across multiple groups.

### Durable UX preferences (2026-04-17 dogfooding)

> Durable learnings captured during dogfooding sessions. These describe what the user values in work you deliver and how they prefer to iterate. Extend this section when new patterns emerge — don't let it calcify.

#### On visual and UI work

- **Palette first, visuals second.**
    - **Why:** During the 2026-04-17 `bonsai init` review, user pushed back on "overall coloring" as the first complaint. Before any re-skinning, semantic tokens must exist — otherwise every color change becomes a codebase-wide find-replace.
    - **How to apply:** For any TUI change that touches color or style, check that semantic tokens (`ColorPrimary`, `ColorAccent`, etc.) exist and are used. Introduce them as a prereq step if missing.

- **Sleek and minimal over ornate.**
    - **Why:** User called the prior UI "bulky", "thrown together", and the flat `B O N S A I` banner "unprofessional". Default ASCII art, excessive padding, spaced-letter wordmarks, and busy layouts are taste-negative signals.
    - **How to apply:** Prefer tight wordmarks, compact boxes, meaningful glyphs over decoration. When in doubt, strip and measure — don't ornament.

- **Visible state over hidden state.**
    - **Why:** User noticed their project name disappeared after advancing to the description prompt, and that required-only ability sections silently skipped. Both broke their sense of "I did something — where did it go?"
    - **How to apply:** Any answered prompt, selected item, or auto-processed section should leave visible evidence on screen. Don't rely on Huh's default clear-on-submit behavior — print a summary line afterward.

- **Rich guidance, not cramped.**
    - **Why:** User said "the next to do hints should be more rich, own their own space, and can be bit more verbose" — terse one-liners at the bottom of panels feel disrespectful of the moment.
    - **How to apply:** Next-steps, file structure views, and any screen that showcases core value should get dedicated visual real estate and substantive copy. Don't default to a single `Hint()` call when the user just finished something significant.

- **TUI should redraw, not stack.**
    - **Why:** User complained that the review panel, generate confirmation, and success messages all piled on top of each other during init. Their mental model is that each major step should feel like a new canvas.
    - **How to apply:** On major step transitions (review, generate, complete), plan for AltScreen or explicit clear/redraw. Don't treat TUI output as a scrollable log.

#### On planning and iteration

- **Fast iteration beats process for UX work.**
    - **Why:** User said "we don't need PRs for this, first we test locally" — for taste-heavy design work, the PR review loop is too slow to be useful until the visual direction is settled.
    - **How to apply:** When user signals "test locally," still use worktrees for isolation but commit directly to main (or merge to main locally). Skip PR creation. Save the PR flow for code-correctness-heavy work.

- **Pick scope pragmatically, foundations first.**
    - **Why:** User dropped 11 items at once (Group F) and said "pick some of these up." They trust me to sequence — which means wrong sequencing wastes their time.
    - **How to apply:** Group by dependency (foundations before consumers — e.g., palette before banner) and visible-win density. Defer architectural rewrites until taste has settled on smaller items. State the deferred scope explicitly so the user can redirect.

- **Propose scope before writing the plan.**
    - **Why:** Plan documents are verbose; scope disagreements at the plan stage cost a rewrite. A scope summary is ~10x cheaper to revise.
    - **How to apply:** For any taste-heavy or ambiguous task, send a scope summary (picked / deferred / rationale) and ask "OK to proceed?" before drafting the plan file.

- **Log findings as they surface, don't batch.**
    - **Why:** User said "add things to backlog as we go along, under ui ux testing category" — they want a running tally in real time, not a post-hoc summary.
    - **How to apply:** Maintain Group F (or equivalent) in `Playbook/Backlog.md` during any testing session. Each finding: category tag, Group F tag, specific fix options if known, source attribution. Don't fix inline and don't wait until the session ends.

#### On communication

- **Concise and direct wins.**
    - **Why:** User makes fast decisions with minimal elaboration ("b, but i will init myself"). Long hedged answers waste their attention.
    - **How to apply:** Short options, direct recommendations, no preamble. Mirror their energy — if they write two sentences, don't respond with five paragraphs.

- **Surface incidental findings proactively.**
    - **Why:** The `go install` binary-name bug was discovered while setting up their test environment, not while looking for it. If I'd silently renamed the binary, the bug would have stayed hidden.
    - **How to apply:** When you hit a workaround while doing setup/chores, explicitly flag it as a finding. Don't normalize broken behavior into your flow.

## References

<!-- Pointers to external resources not documented elsewhere in the project. -->

- **Foundational research docs** — Anchor for methodology/concept decisions. Reference these when improvements need grounding in the core philosophy (ambient behavior, meta-layer, talents, evals).
    - [Research/RESEARCH-landscape-analysis.md](../../Research/RESEARCH-landscape-analysis.md) — How Bonsai compares to GSD/ECC/others; Bonsai's identity/coordination layer positioning
    - [Research/RESEARCH-concept-decisions.md](../../Research/RESEARCH-concept-decisions.md) — Ambient vs. command-driven, authority hierarchy, catalog ownership, talents taxonomy
    - [Research/RESEARCH-eval-system.md](../../Research/RESEARCH-eval-system.md) — Eval system concept: scenarios, evaluators, benchmarks for methodology rigor
    - [Research/RESEARCH-trigger-system.md](../../Research/RESEARCH-trigger-system.md) — Trigger section design research
    - [Research/RESEARCH-uiux-overhaul.md](../../Research/RESEARCH-uiux-overhaul.md) — UI/UX overhaul research (Plan 14/15 origin)
