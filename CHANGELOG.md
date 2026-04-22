# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.3] — 2026-04-15
### Changed
- Apply `gofmt -s` across all Go source files for canonical formatting.

## [0.1.2] — 2026-04-15
### Fixed
- Restore Go Reference badge now that case-collision module-proxy issue is resolved.
- Retract v0.1.0 release notice (see v0.1.1 fix).

## [0.1.1] — 2026-04-15
### Added
- Obsidian-compatible markdown links across generated workspace files (`[[link]]` syntax in nav tables and cross-references).
- `workspace-guide` skill — installed with every agent to onboard humans into the generated workspace.
- "How to Work" decision heuristics section in generated `CLAUDE.md`.
- PR-triggered CI workflow — tests + `go vet` as required status checks.
### Fixed
- Case-insensitive file collision on Windows/macOS — `station/index.md` renamed to `station/code-index.md` to avoid collision with `INDEX.md`.
- CI badge and Go version badge corrections in README.
- Reverted stray `INSTRUCT` reference to backtick in memory protocol.

## [0.1.0] — 2026-04-15

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

[Unreleased]: https://github.com/LastStep/Bonsai/compare/v0.1.3...HEAD
[0.1.3]: https://github.com/LastStep/Bonsai/compare/v0.1.2...v0.1.3
[0.1.2]: https://github.com/LastStep/Bonsai/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/LastStep/Bonsai/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/LastStep/Bonsai/releases/tag/v0.1.0
