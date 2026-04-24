# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.3.0] - 2026-04-24

> **The "your AI agent can work with Bonsai itself" release.** Every new workspace now carries an agent-readable mental model of Bonsai, a filesystem-discoverable catalog snapshot, and copy-paste prompts to get your agent started. Plus a scope-guard correctness fix and cinematic polish on the last two commands that needed it.

### Added

- **Your agent can now reason about Bonsai directly.** Every new workspace ships a `bonsai-model` skill at `agent/Skills/bonsai-model.md` — a concise mental model covering how the catalog is shaped, when to add vs customize, and how the commands fit together. Point your agent at it and it can propose changes without you hand-feeding docs.
- **Machine-readable catalog.** `bonsai init`, `add`, and `update` now write `.bonsai/catalog.json` at your project root — a filesystem-discoverable listing of every agent type, skill, workflow, protocol, sensor, and routine the installed binary ships. Same data available on demand via `bonsai catalog --json` for CI or scripted consumption (respects `-a <agent>` filter).
- **"Bonsai Reference" in every generated CLAUDE.md.** A pointer table near the top of each workspace CLAUDE.md tells your agent where to find the model doc, the catalog snapshot, and the installed-state file. Pi.dev-inspired progressive-disclosure pattern: zero tokens spent until the agent actually needs to know.
- **Post-command hints you'll actually use.** `init`, `add`, `remove`, and `update` yield screens now show three labeled sections:
  1. **Next steps** — the mechanical CLI commands to run next
  2. **Try this** — workflow suggestions inside your new workspace
  3. **Ask your agent** — bordered copy-paste prompt boxes you can hand straight to Claude / your editor / wherever
  Hints are agent-type aware — a backend agent gets backend-shaped prompts, a devops agent gets devops-shaped ones.
- **Cinematic `bonsai remove` and `bonsai update`** — the last two commands without a dedicated flow package got the same treatment as `init` / `add` / `list` / `catalog` / `guide`. Stage rails, chromeless prompts for destructive actions, per-file conflict resolver, responsive resize, ASCII fallback.
- **`NO_COLOR` and `TERM=dumb` honored explicitly.** Scripts piping Bonsai output and color-averse terminals now get clean text with zero ANSI escapes. Covered by regression tests so it stays that way.

### Changed

- **`bonsai add` keeps every agent's awareness up to date automatically.** After you add a new agent, Bonsai now refreshes the peer-awareness files on every already-installed agent — you no longer need to run `bonsai update` after each `add` to keep scope-guards honest. Lock-aware, so any hand-edits you've made to those files still trigger the conflict resolver.
- **`bonsai update` non-TTY path returns a proper exit code.** When run in a script and conflicts can't be auto-resolved, `bonsai update` now exits non-zero so your CI actually notices. Previously it printed a warning and silently returned success.
- **Cleaner internals in `cmd/update.go`** (313 lines → ~85). All TUI logic delegated to the new flow package; business logic preserved.

### Fixed

- **Silent cross-agent scope-guard bypass.** Before 0.3, running `bonsai add backend` in a project that already had a tech-lead did *not* update tech-lead's `scope-guard-files.sh` with a block-rule for the new `backend/` workspace. Net effect: dispatching tech-lead to edit a backend file quietly worked when scope-guard was supposed to exit 2 and block it. Same class of silent staleness on `dispatch-guard.sh`'s workspace→agent map. Now refreshed automatically on every `add`.

### Security

- **Peer-awareness state is now always correct after `add`.** Closes the class of silent scope-guard bypasses described in Fixed. No more "tech-lead can edit any agent's files" once a project grows past one agent.

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
