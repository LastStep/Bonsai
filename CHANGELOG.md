# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.3.0] - 2026-04-24

### Added
- `bonsai-model` skill — pi.dev-style mental model doc installed by default on tech-lead. Covers catalog shape, customization model, decision heuristics. Agent-readable on-demand; non-tech-lead agents reference it via relative path. Part of the "make Bonsai agent-consumable" track.
- `.bonsai/catalog.json` on-disk snapshot — filesystem-discoverable catalog listing written at project root on every `init`/`add`/`update`. Agent can read it directly without invoking the CLI. Stable JSON schema via `generate.SerializeCatalog`.
- `bonsai catalog --json` flag — machine-readable catalog output to stdout, respects `-a <agent>` filter. Shares the same `SerializeCatalog` source of truth as the on-disk snapshot.
- "Bonsai Reference" block in generated workspace `CLAUDE.md` — pointer table to `bonsai-model.md`, `.bonsai/catalog.json`, and `.bonsai.yaml`. Path computed via `filepath.Rel` from workspace root (tech-lead self-reference, peers via `../station/...`).
- Cinematic `bonsai remove` flow — new `internal/tui/removeflow/` package with 4-stage rail (択 SELECT → 観 OBSERVE → 確 CONFIRM → 結 YIELD) + chromeless CONFLICT splicing. Confirm stage defaults focus to BACK (destructive opt-in). Covers both agent-removal and per-category item-removal paths.
- Cinematic `bonsai update` flow — new `internal/tui/updateflow/` package with 5-stage rail (探 DISCOVER → 択 SELECT → 同 SYNC → 衝 CONFLICT → 結 YIELD). Invalid custom-file warnings moved from pre-harness stdout into the DISCOVER stage panel. Non-TTY `RunStatic` fallback auto-accepts discoveries and returns non-zero exit on unresolved conflicts.
- Hints 3-layer overhaul — new `internal/tui/hints/` renderer + `catalog/agents/*/hints.yaml` content files for all 6 agent types. Yield stages now show NEXT STEPS (mechanical CLI), TRY THIS (workflow prompts), and ASK YOUR AGENT (bordered copy-paste AI-prompt boxes) per init/add/remove/update. Template substitution via `{{ .DocsPath }}`, `{{ .AgentName }}`, `{{ .ProjectName }}`.
- Explicit `NO_COLOR` + `TERM=dumb` honoring in `internal/tui/styles.go` — `shouldDisableColor(fd)` helper with test coverage locking the no-ANSI-escape contract.

### Changed
- `bonsai add` now refreshes peer awareness after install — every already-installed agent's `identity.md`, `scope-guard-files.sh`, and `dispatch-guard.sh` is re-rendered with the new agent in their `{{ range .OtherAgents }}` list. Lock-aware via existing `writeFile` pathway so user-edited copies still trigger the conflict resolver.
- `AgentWorkspace` now builds its template context via the shared `buildAgentTemplateContext` helper (same path as `RefreshPeerAwareness`) — prevents divergence if `identity.md.tmpl` ever references `Workspace`/`DocsPath`/`Skills`/etc.
- `cmd/update.go` shrunk from 313L to ~85L — all TUI logic delegated to `internal/tui/updateflow/`. Business-logic helpers (`buildCustomFileOptions`, `applyCustomFileSelection`, `appendUnique`) preserved.
- `cmd/remove.go` rewired to delegate both `runRemove` (agent) and `runRemoveItem` (per-category) to `removeflow.Run(...)`. Cobra wiring, flag parsing, and business-logic helpers preserved.

### Fixed
- Cross-agent `OtherAgents` template staleness on `bonsai add` — tech-lead's `scope-guard-files.sh` previously did NOT get `# Block writes to <new-agent>/` entries when a new agent was added; same for `dispatch-guard.sh`'s workspace→agent map. Silent correctness bug: scope-guards failed open, dispatch validation silently skipped. Now re-renders peer awareness automatically.
- `cmd/update.go` non-TTY path now returns non-zero exit code when `RunStatic` encounters unresolved conflicts or sync errors. Previously printed a warning and returned `nil` — invisible to CI.

### Security
- Peer awareness files now always reflect installed-agent state, closing a class of silent scope-guard bypasses where tech-lead could edit other agents' workspaces without the guard firing.

## [0.2.0] - 2026-04-22

### Added
- Cinematic `bonsai init` flow — Vessel (project name + docs path) → Soil (scaffolding) → Branches (tabbed ability picker) → Observe (review) → Generate → Planted summary, with kanji/kana stage rail, semantic palette, responsive resize, and ASCII fallback for terminals without wide-char support.
- Cinematic `bonsai add` flow — Select → [Ground] → Graft → Observe → Grow → [Conflicts] → Yield, with a per-file conflict picker (Skip / Overwrite / Backup) and four terminal Yield variants (success, all-installed, tech-lead-required, unknown-agent).
- Persistent statusline at `station/agent/Sensors/statusline.sh` — context %, 5h and 7d budget %, model, branch (with dirty marker), elapsed, cost, plus optional caveman-mode badge; sage/sand/rose 256-color tiers with `NO_COLOR` and `BONSAI_STATUSLINE_HIDE` honoured.
- `compact-recovery` sensor — re-injects Quick Triggers and work state after `/compact` so the agent doesn't lose its operating context.
- `context-guard` patterns expanded — verify and plan triggers in addition to existing tier-based context-percentage injections.
- Trigger metadata system — `triggers:` blocks on skills/workflows feed both generated CLAUDE.md tables and `.claude/skills/{name}/SKILL.md` slash commands.
- Starlight documentation site under `website/` — concepts, command reference, catalog browser (auto-generated), LLM-friendly llms.txt layer, deploy workflow.
- `bonsai guide` extended from a single topic to four — `quickstart`, `concepts`, `cli`, `custom-files`.
- `RenderFileTree` widget and palette chrome tokens (`ColorLeafDim`, `ColorRule`, `ColorRule2`, `ColorAccent`) used across the cinematic flows.

### Changed
- BubbleTea step/reducer harness in `internal/tui/harness/` is now the foundation for `init`, `add`, `remove`, and `update` — replaces the per-command Huh form chains.
- README rewritten audience-first — new tagline "A workspace for your coding agent", `Who Bonsai is for` section, mechanism-over-personality bullets with concrete file references, `station/` tree snippet, demo gif.
- `bonsai init` and `bonsai add` cinematic flows are now the default — no env flag required.
- Adaptive color palette across the TUI — automatic light/dark detection, `NO_COLOR` honoured, `FatalPanel` and version banner consistency, structured error display via `ErrorDetail`.
- Phase 2 UI consistency pass — ordering, item counts, key hints, "Up to date" no-op detection across pickers.
- CI workflow widened to run on `push: branches: [main]` in addition to `pull_request` — closes the "gofmt drift on main silently hides because CI is PR-only" pattern.
- `docs.yml` deploy workflow gains `pull_request` trigger (with deploy job guarded to push events) so broken MDX fails at PR time instead of post-merge.
- `bonsai` binary now builds from `cmd/bonsai/main.go` so `go install github.com/LastStep/Bonsai/cmd/bonsai@latest` produces the correct lowercase binary name.
- Go toolchain bumped to 1.25.8.
- golangci-lint migrated from v1 to v2.

### Removed
- `BONSAI_REDESIGN` env gate — cinematic `bonsai init` is the only path; legacy harness body deleted.
- `BONSAI_ADD_REDESIGN` env gate — cinematic `bonsai add` is the only path; legacy `runAddSpinner`, `buildNewAgentSteps`, `buildAddItemsSteps`, and `addOutcome` deleted.
- Three orphan legacy guide topics (1,213 lines) replaced by the four-topic `bonsai guide` set.

### Fixed
- Plan 19 OSS-blocker bug sweep — CRLF handling on Windows checkouts, cross-workspace tree rendering, ability dedup across agents, spinner-step `errors.Join` so concurrent failures surface, and assorted harness polish.
- `bonsai init` Planted stage shows the correct station path and the Observe stage no longer reports a misleading file count.
- `chmod` is now re-applied on `ActionUnchanged` so sensor scripts stay executable across re-runs.
- `.bak` write failures no longer silently discard files in the conflict picker — failed-backup paths are dropped from the overwrite list and a single warning is emitted.
- Doubled path prefix in `bonsai add` output panels.

### Security
- CodeQL workflow added for Go SAST.
- `govulncheck` job added to CI.
- Dependabot configured for `gomod` and `github-actions`.
- Two low-severity CodeQL `useless-assignment-to-local` alerts silenced after audit.
- Astro XSS advisory (CVE-2026-41067) resolved by lockfile bump to `astro@6.1.7`.

## [0.1.3] — 2026-04-16
### Changed
- Apply `gofmt -s` across all Go source files for canonical formatting.

## [0.1.2] — 2026-04-16
### Fixed
- Restore Go Reference badge now that case-collision module-proxy issue is resolved.
- Retract v0.1.0 release notice (see v0.1.1 fix).

## [0.1.1] — 2026-04-16
### Added
- Obsidian-compatible markdown links across generated workspace files (`[[link]]` syntax in nav tables and cross-references).
- `workspace-guide` skill — installed with every agent to onboard humans into the generated workspace.
- "How to Work" decision heuristics section in generated `CLAUDE.md`.
- PR-triggered CI workflow — tests + `go vet` as required status checks.
### Fixed
- Case-insensitive file collision on Windows/macOS — `station/index.md` renamed to `station/code-index.md` to avoid collision with `INDEX.md`.
- CI badge and Go version badge corrections in README.
- Reverted stray `INSTRUCT` reference to backtick in memory protocol.

## [0.1.0] — 2026-04-16

First public release. Go CLI for scaffolding Claude Code agent workspaces.

### Added
- Core commands — `bonsai init`, `add`, `remove`, `update`, `list`, `catalog`, `guide`.
- Six agent types — tech-lead (required), frontend, backend, fullstack, devops, security.
- Catalog of skills, workflows, protocols, sensors, routines — each with `meta.yaml` + content file; `required` and `agents` compatibility fields.
- Routines — periodic self-maintenance with auto-managed `routine-check` sensor + dashboard at `agent/Core/routines.md`.
- Sensors — auto-enforced hook scripts wired into `.claude/settings.json` at generation.
- Awareness framework — `status-bar` and `context-guard` sensors for live context monitoring and guardrails.
- Lock file (`.bonsai-lock.yaml`) — content hashing with conflict detection on re-run; user-modified files trigger skip/overwrite/backup prompt.
- Selective file update with multi-select conflict picker via `bonsai update`.
- Individual item removal — `bonsai remove <kind> <name>`.
- `display_name` catalog field — decouples human-readable labels from machine identifiers; auto-derived from `name` when omitted.
- Backlog system across catalog — intake queue scaffolding for every agent.
- PR-driven subagent workflow — dispatched agents create draft PRs in isolated worktrees; tech-lead reviews and merges.
- Release pipeline — GoReleaser v2, GitHub Actions tag trigger, Homebrew tap (`LastStep/homebrew-tap`).
- Dogfooding — Bonsai generates its own `station/` workspace with tech-lead agent.

[Unreleased]: https://github.com/LastStep/Bonsai/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/LastStep/Bonsai/compare/v0.1.3...v0.2.0
[0.1.3]: https://github.com/LastStep/Bonsai/compare/v0.1.2...v0.1.3
[0.1.2]: https://github.com/LastStep/Bonsai/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/LastStep/Bonsai/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/LastStep/Bonsai/releases/tag/v0.1.0
