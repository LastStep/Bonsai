---
tags: [log, session, plan-40]
description: Session log — Plan 40 dispatch (Odysseus integration, v0.5.0 Phases 1–3).
date: 2026-06-13
---

# Session — Plan 40 Dispatch (v0.5.0)

## Outcome
Plan 40 Phases 1–3 shipped to `main` (v0.5.0, **untagged**, additive). Phase 4 held, dogfood deferred, release tag held — all per user.

## Shipped
- **Phase 1** ([PR #114](https://github.com/LastStep/Bonsai/pull/114), `1e715c7`) — frozen v1 schemas, `RootRelative` scaffolding, `project-manifest` + `memory` catalog items, `catalog.Slugify`, NoteStandards memory-note schema.
- **Phase 2** ([PR #116](https://github.com/LastStep/Bonsai/pull/116), `a540fdd`) — `bonsai validate` project-level pass (manifest + memory-note lint, adversarial-grade note-target resolution, bounded walk).
- **Phase 3** ([PR #115](https://github.com/LastStep/Bonsai/pull/115), `2aef7fd`) — memory-routing protocol, `howToWorkLines` fix, `bonsai guide formats` page.
- CHANGELOG `## [0.5.0] - Unreleased` (`70a31f9`).

## Process
- Dispatch: P1 first (freezes schemas) → P2 + P3 in parallel off main. Each phase = own worktree PR; independent review agent per PR.
- **P2 review caught a BLOCKING sec-bug:** traversing `memory_dir` (`../escape`) was flagged `invalid_manifest` but `auditProject` still walked the resolved out-of-tree dir + read files → out-of-tree disclosure. Fixed in-branch (`memoryDirInvalid` blank-and-skip + regression test proven fail→pass).
- **P3 review caught a doc-vs-reality bug:** `formats.md` documented the held Phase-4 `bonsai update` delivery path. Fixed → `bonsai init` re-run.
- Post-merge audit green: build / `go test ./...` / Windows cross-compile / vet.

## Decisions (user)
- **Phase 4 HELD** → superseded by headless-CLI parity workstream ([Backlog P1](../Playbook/Backlog.md)).
- **Dogfood deferred** for v0.5.0 — blocked by held-Phase-4 (no CLI delivery to existing repos: `bonsai init --non-interactive` refuses an existing config by design) + repo gitignores `.bonsai-lock.yaml` (validate can't pass here, pre-existing 38-issue orphan wall). Feature proven via tests + direct `generate.Scaffolding()` (created=4, skipped=12, zero churn, manifest at repo root).
- **v0.5.0 tag HELD** — CHANGELOG prepped; cut later (triggers binaries + Homebrew).

## Follow-ups (Backlog)
- 3 review nits: manifest `created` via `yamlScalar`; stale `manifestRel` comment; `valid_from` unvalidated.
- `.bonsai-lock.yaml` gitignored → validate unusable on this repo (lock-policy decision).
- **Infra flag:** `isolation:"worktree"` agents leaked `Edit`/`Write` to the MAIN checkout repeatedly this session (P3 even clobbered P2's WIP). Handled each time, main stayed clean. Mitigation captured in [memory.md](../agent/Core/memory.md) Notes.

## Next session
`/plan` the headless-CLI parity workstream — unblocks Phase 4 + the dogfood.
