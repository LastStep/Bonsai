---
tags: [plan, integration, odysseus]
description: Bonsai-side build plan for the Odysseus personal-platform integration. Rung 1 (B+C+D+A) decisions locked 2026-06-13; grilled (round 1) 2026-06-13.
status: active
source: odysseus-design-session-2026-06-12; decisions-grill 2026-06-13
tier: 2
---

# Plan 40 — Odysseus Platform Integration

**Tier:** 2
**Status:** Active — grilling in progress (round 1 applied)
**Agent:** code agents (gp) via worktree, tech-lead orchestrates

## Goal

Ship the bonsai-side repo standards the Odysseus hub consumes: in-repo memory graph (`station/Memory/`), per-repo project manifest (`.bonsai/project.yaml`), the canonical schemas documented as the bonsai standard, and `validate` lint for both. "Done" for Rung 1 = a bonsai project scaffolds a schema-conformant memory tree + manifest, existing projects receive them via `bonsai update`/`add`, `bonsai validate` lints them, `bonsai guide` documents the formats, and our own station dogfoods it — shipped as **v0.5.0**.

## Context

Odysseus (`/home/rohan/Apps/odysseus`, private fork) is being rebuilt into a personal platform. Boundary decided 2026-06-12 (odysseus `PLATFORM.md` §11):

> **Bonsai = repo-side standard.** Defines, scaffolds, lints everything inside a project repo. Host-agnostic — keeps working with bare Claude Code, no Odysseus required.
> **Odysseus = hub runtime.** Indexes repos, aggregates cross-project, renders UI, owns global memory + orchestration. Writes into repos only via defined bridge actions.

Bonsai is the schema authority for all repo-resident formats; Odysseus implements parsers against the published standard. The schema in PLATFORM §4 is the **frozen, agreed contract** — the user owns the Odysseus implementation; this workspace builds the bonsai side directly against it (no handshake loop).

## Decisions Locked (2026-06-13 grill)

| # | Decision | Rationale |
|---|----------|-----------|
| Scope | **Rung 1 = B + C + D + A.** | Note `scope: project/<slug>` needs an authoritative slug source; manifest is it. |
| Release | **v0.5.0** minor bump at end of rung 1. | Additive scaffolding + validate extensions + guide pages. |
| Dispatch | **PR-flow code agents via worktree.** | Schema/lint = correctness-heavy. |
| Dogfood | **Yes** — our station gets `station/Memory/` + `.bonsai/project.yaml`, in a **separate commit/PR** from the generator code (so it reverts independently). | Bonsai is project #1. |
| Manifest location | **`.bonsai/project.yaml`** (not `.odysseus/`). | Host-agnostic — bonsai never hardcodes a consumer's name. Hub checks both paths. |
| **Manifest writer** | **Lock-aware `writeFile` + new `root_relative: true` flag on `ScaffoldingItem`** that skips the `docs_path` prefix. Borrow only the `.bonsai/` dir + `openSnapshotFile` O_NOFOLLOW build-tag idiom from `catalog_snapshot.go` — do NOT call `WriteCatalogSnapshot`. | (Round-1 grill) `Scaffolding()` unconditionally prefixes `docs_path`; `WriteCatalogSnapshot` is a separate, NOT-lock-tracked writer. The `writeFile` path gives lock-tracking + write-once + conflict-handling for free. |
| **Delivery path** | **Teach `bonsai update` + `bonsai add` to deliver newly-selected scaffolding items** into existing projects (with the standard conflict handling). | (Round-1 grill) `Scaffolding()` is called only from `init`; `update` never delivers new items. Existing v0.4.x projects must get the memory/manifest items via `bonsai update`. |
| Manifest refresh | **Write-once + validate-only.** No field-merge. | File barely changes; bonsai never re-edits after scaffold. |
| Config split | **`.bonsai.yaml` and `.bonsai/project.yaml` stay separate.** `project.yaml` is the **project-identity authority** (name/slug/description); `.bonsai.yaml`'s copies are independent/cosmetic (no sync rule needed). | Different owners/lifecycles/consumers; AI team is opt-in, manifest is universal. |
| Manifest fields | Frozen v1 (table below). **`next_action` removed.** | Only volatile field + duplicated `Status.md`. |
| Memory routing | **Decisions → `Memory/decisions/`; FieldNotes stays legacy; `agent/Core/memory.md` untouched.** Phase 3 also updates `generate.go` `howToWorkLines` so the emitted "Decision logging → KeyDecisionLog" heuristic doesn't contradict the new routing. | (Round-1 grill) avoids a dual decision-system interim with contradictory agent guidance. |
| Migration | **Scaffold now, migrate existing KeyDecisionLog entries later.** | Don't block release on hand-conversion. |
| Note schema | **Freeze §4 shape + `schema_version`.** | Migration path; validate pins lint rules to the version. |
| Versioning | `schema_version: 1` (int) on note + manifest. Deliberately distinct from `CatalogSnapshot.Version` (a build-version string). | Bonsai-as-authority → every published format versioned. Post-v0.5.0, field changes need a `schema_version` bump + hub coordination even on a bonsai minor. |
| **Input validation** | `slug` constrained to `[a-z0-9-]`; `memory_dir` validated as repo-relative, non-traversing via `internal/wsvalidate`; note/manifest target resolution (`superseded_by`, `[[relations]]`) anchored under `memory_dir`, escapes rejected, resolved by sanitized `permalink` in an in-memory index (never by treating link text as a path); user scalars (`ProjectName`/`Description`) emitted via YAML-quoting with `missingkey=error`. | (Round-1 grill, Security) path-traversal + YAML-injection + downstream-hub-injection surfaces. |
| Catalog content | `affects:` strings phrased **host-agnostically** ("downstream hub ingest / repo indexers", "memory-graph consumers") — never "Odysseus". | (Round-1 grill, Architecture) `affects:` ships in the binary; naming a consumer re-introduces the coupling `.bonsai/` avoided. |
| MEMORY.md budget | **Warning at >200 lines.** | Context-window guideline, not correctness. |
| Docs target | **`bonsai guide` (one combined "Formats" page) + scaffolded templates.** No website. | Agent + hub-implementer audience; ships in binary. |
| graphify (E) | **Deferred** behind rungs 1–2 + trust-vet first. | Third-party, Python dep, new git-hook mechanism. |
| Plan format (G), `--minimal` | **Deferred** (rung 4 / Odysseus owns plain-repo manifests). | Not needed for memory/manifest. |
| Scaffolding required-ness | Both new items ship **`required: false`** (opt-in). | Non-breaking; we opt in for dogfood. |

## Frozen schemas (v1)

### Memory note — `{memory_dir}/{decisions,notes}/<permalink>.md`

```markdown
---
schema_version: 1
title: <note title>
type: decision | note | fact | log
permalink: <stable-slug>
tags: []
scope: project/<slug>
valid_from: 2026-06-13
superseded_by: null
---
## Observations
- [category] one fact per bullet #tag (optional context)

## Relations
- relation_type [[Target Note]]
```

| field | type | required | rule |
|-------|------|----------|------|
| `schema_version` | int | ✅ | `== 1` |
| `title` | string | ✅ | non-empty |
| `type` | enum | ✅ | `decision\|note\|fact\|log` |
| `permalink` | string | ✅ | `[a-z0-9-]`; unique within tree |
| `tags` | list[string] | ⬜ | default `[]` |
| `scope` | string | ✅ | `project/<slug>` matching manifest `slug` |
| `valid_from` | date | ⬜ | `YYYY-MM-DD` if present |
| `superseded_by` | string\|null | ⬜ | if non-null, must resolve to an existing `permalink` |

Notes = graph nodes; typed `[[wikilinks]]` = edges; observations individually indexable. Forward refs to not-yet-existing notes are legal (→ warning, not error).

### Project manifest — `.bonsai/project.yaml` (repo root)

```yaml
schema_version: 1
name: <project name>
slug: <stable-slug>
status: idea | active | paused | done | archived
tags: []
description: <one-line>
links:
  repo: <url>
  docs: <url>
  issues: <url>
created: 2026-06-13
memory_dir: station/Memory
```

| field | type | required | rule |
|-------|------|----------|------|
| `schema_version` | int | ✅ | `== 1` |
| `name` | string | ✅ | non-empty |
| `slug` | string | ✅ | `[a-z0-9-]` |
| `status` | enum | ✅ | `idea\|active\|paused\|done\|archived` |
| `tags` | list[string] | ⬜ | |
| `description` | string | ⬜ | |
| `links` | map | ⬜ | keys `repo\|docs\|issues`, each a URL; unvalidated content (no schema for the URL) |
| `created` | date | ✅ | `YYYY-MM-DD` |
| `memory_dir` | string | ⬜ | repo-relative, non-traversing; default `station/Memory` |

**Slug algorithm (frozen):** lowercase; replace each run of `[^a-z0-9]` with a single `-`; trim leading/trailing `-`. (Mirrors the `[a-z0-9-]` naming standard in `CLAUDE.md`.) Net-new helper in `internal/catalog` (no existing slugify — `DisplayNameFrom` is the inverse).

## Rung 1 — Implementation Phases

> Dispatch order: **Phase 1 first** (generator/catalog — freezes schemas + delivery). Then **Phases 2 + 3 in parallel** (file-disjoint: `internal/validate/` vs docs). All `catalog/scaffolding/` edits — including the `NoteStandards.md.tmpl` extension and `manifest.yaml` registration — live in **Phase 1's worktree** so they don't collide with Phase 3.

### Phase 1 — Generator + catalog (A + B + delivery) — `gp`, one worktree

**1a. `root_relative` scaffolding support + project-manifest item (A).**
- Add `RootRelative bool` to `ScaffoldingItem` (`internal/catalog/catalog.go`) + manifest field. In `Scaffolding()` (`internal/generate/generate.go` ~L405), when `RootRelative`, skip the `cfg.DocsPath` join so the file lands at repo root.
- Route the write through the existing lock-aware `writeFile` (lock-tracked, write-once, conflict-handled). For the `.bonsai/` directory creation + file open, factor `openSnapshotFile` into a shared `openRootFile` (or reuse it) so the O_NOFOLLOW build-tag split (`*_unix.go`/`*_windows.go`) is preserved — do NOT add a raw `syscall.O_NOFOLLOW` inline.
- Add the `project-manifest` item to `catalog/scaffolding/manifest.yaml`: `required: false`, `root_relative: true`, host-agnostic `affects:`, `files: [.bonsai/project.yaml.tmpl]`.
- Template renders the frozen manifest. `name`←`{{ .ProjectName }}`, `slug`←new slugify helper, `description`←`{{ .ProjectDescription }}`, `created`←new date source (add a `now`/`RenderDate` to the template funcMap/context — none exists today). Emit `ProjectName`/`ProjectDescription` via a YAML-quoting func; set `Option("missingkey=error")` on the manifest render.
- Validate `memory_dir` (if user-set) via `internal/wsvalidate`; reject non-repo-relative/traversing values.

**1b. `memory` scaffolding item (B).**
- Add to `manifest.yaml`: `required: false`, host-agnostic `affects:`, `files:` for the tree. Scaffold `{docs_path}/Memory/`: `MEMORY.md.tmpl` + `decisions/.gitkeep` + `notes/.gitkeep`. **No `log/`** (no rung-1 producer; would be force-linted against the note schema).
- Use trailing-slash dir entries or explicit `.gitkeep` paths to satisfy `isAllowedScaffoldingFile`.
- `MEMORY.md.tmpl` starter (≤200 lines): `# {{ .ProjectName }} — Memory Index` heading, a one-paragraph purpose blurb, a `## Decisions` + `## Notes` index section (empty), and a link to the note-schema guide. Self-describing.

**1c. Delivery via update/add.**
- Teach `bonsai update` (`cmd/update.go` + `internal/tui/updateflow/`) and `bonsai add` (`cmd/add.go` + `addflow/`) to scaffold newly-selected scaffolding items into existing projects, reusing `Scaffolding()` + the lock-aware conflict path. Existing projects opt in (select the item) and receive it on `update`/`add`.

**1d. NoteStandards + tests.**
- Extend the **existing** `catalog/scaffolding/Playbook/Standards/NoteStandards.md.tmpl` (it's already shipped by the `playbook` item — edit, don't create) with the memory-note schema as THE project note standard.
- Tests in `internal/generate/generate_test.go` (or new `*_test.go`): manifest render + slug fixtures (name→expected slug) + root-relative path (asserts file at repo-root `.bonsai/project.yaml`, NOT `station/.bonsai/`) + lock tracking + memory-tree scaffold + idempotent re-run + update/add delivery.

### Phase 2 — Validate project-level pass (D) — `gp`, depends on Phase 1

- Add a **new project-level audit pass** to `validate.Run()` (`internal/validate/validate.go`), run once (not inside the per-agent `auditAgent` loop, which is top-level-only and would skip `decisions/`/`notes/`). Register new **additive** `Category` constants (appending to the stable JSON contract is non-breaking).
- Resolve `memory_dir` + `slug` by parsing `.bonsai/project.yaml` (`yaml.Unmarshal` into a new typed struct — do NOT extend `CustomItemMeta`). If manifest absent but `Memory/` exists → **warning** ("no project.yaml; scope unverifiable"), skip scope-match, still lint frontmatter.
- Recursively walk `{memory_dir}/**/*.md` with a sane file-count/size bound (avoid CI OOM on a pathological tree). Per note: required frontmatter (`schema_version,title,type,permalink,scope`) → **error** if missing; `scope` ≠ `project/<slug>` → **error**; `superseded_by` non-null target unresolved → **error**; unresolved `[[relation]]` → **warning**. Target resolution anchored under `memory_dir`, `../` escapes rejected, lookups by sanitized `permalink` in an in-memory index.
- `MEMORY.md` > 200 lines → **warning**.
- `.bonsai/project.yaml` schema: required fields present, `status` enum, `schema_version == 1` → **error** on violation. (Reuses `--json`/`--agent` flag shapes already present.)
- Tests: table-driven fixtures in `internal/validate/validate_test.go` — one per error/warning rule asserting category+severity, **plus a valid-tree fixture asserting zero issues** (no false positives). Assert exit code (any issue → existing convention).

### Phase 3 — Docs (C) — `gp`, parallel with Phase 2

- **`agent/Protocols/memory.md` (catalog source):** document the three-surface routing — working memory (`agent/Core/memory.md`, always-loaded scratch) vs durable graph (`station/Memory/`, decisions+facts, hub-indexed) vs Logs (`KeyDecisionLog` legacy/read-only, `FieldNotes` free-form). State where each write goes.
- **`internal/generate/generate.go` `howToWorkLines` (~L597):** update the emitted "Decision logging → KeyDecisionLog" line so it doesn't contradict the new routing.
- **`bonsai guide`:** one combined "Formats" page (rendered by guideflow) covering the note schema + project manifest; gate asserts the rendered body contains `schema_version` and `permalink`.
- Update root `CLAUDE.md` / `station/CLAUDE.md` nav only if new files need routing entries.

## Dependencies

- Phase 1 freezes on-disk schemas + delivery → Phases 2 and 3 depend on it. No new Go module deps. No external services.

## Rollback

- Each phase ships as a revertable PR; pre-tag, v0.5.0 is `git revert`-able.
- The dogfood scaffold (our repo's `.bonsai/project.yaml` + `station/Memory/`) lands in a **separate commit/PR** from generator code, reverts independently.
- Write-once caveat: a broken scaffolded manifest must be hand-deleted to regenerate (skip-if-exists blocks an in-place fix). Acceptable; documented.
- Runtime rollback for the scaffolding items: `required: false` → a project simply doesn't select them.

## Security

> [!warning]
> Refer to [Playbook/Standards/SecurityStandards.md](../../Standards/SecurityStandards.md).

- No secrets in templates/schemas/examples.
- **Path safety:** root-relative write stays inside repo root (reuse the `openRootFile`/O_NOFOLLOW idiom — symlink-substitution defense, v0.4.0 class). `memory_dir`, `slug`, `permalink`, relation/`superseded_by` targets are all path-ish untrusted inputs — constrain charset, anchor resolution under `memory_dir`, reject `../` escapes (see Input-validation decision).
- **Template injection:** user scalars rendered via YAML-quoting + `missingkey=error`, never bare `{{ . }}` into YAML.
- **Parser safety:** validate parses YAML/frontmatter into typed structs only (no `map[string]interface{}` to anything executable); never evals content; bounded file walk.
- Optional defense-in-depth (backlog, not rung 1): warn if a `links` value contains `://user:pass@`.

## Verification

- [ ] `bonsai init` (interactive + `--non-interactive`) scaffolds `.bonsai/project.yaml` at **repo root** (assert NOT `station/.bonsai/project.yaml`) + `station/Memory/{decisions,notes}/` when items selected; both **lock-tracked** in `.bonsai-lock.yaml`.
- [ ] `bonsai update` / `bonsai add` delivers a newly-selected memory/manifest item into an existing project; re-run reports `Skipped`/`Unchanged`, exits 0, no conflict prompt, lock hashes unchanged.
- [ ] Generated `project.yaml`: `./bonsai validate --json` reports zero issues; assert `schema_version: 1`, non-empty `name`/`slug`/`created`, `status` ∈ enum. Generator test asserts rendered bytes + slug fixtures.
- [ ] `bonsai validate` negative controls (table-driven fixtures): missing note frontmatter → error; scope mismatch → error; dangling `superseded_by` → error; unresolved relation → warning; MEMORY.md >200 → warning; malformed manifest → error. **Valid-tree fixture → zero issues.** `--json` well-formed.
- [ ] `bonsai guide` Formats page renders; body contains `schema_version` + `permalink`.
- [ ] Dogfood (separate PR): this repo gets `.bonsai/project.yaml` + `station/Memory/`; `./bonsai validate` exits clean.
- [ ] `GOOS=windows GOARCH=amd64 go build ./...` passes (baseline green today; root-relative write uses the build-tag-split open, no inline POSIX syscall).
- [ ] CI green; CHANGELOG entry; v0.5.0 tag.

## Out of scope for bonsai (Odysseus owns)

UI; cross-project indexes/search; global memory tier; embeddings/graph infra; dispatch runtime; secrets; connections; scheduling.

## Later rungs

- **Rung 2:** F (`bonsai export` JSON contract — supersedes `_roles_from_bonsai_yaml`, odysseus `workshop/ingest.py:110`).
- **Rung 2.5:** Multi-agent `--from-config` for `bonsai init` (odysseus `workshop/lifecycle.py` splits init + sequential `add` today). Backlog.
- **Rung 3:** E (graphify repo-side wiring) — vet first.
- **Rung 4:** G (plan write-back format).

## Cross-repo follow-ups (outside this workspace)

- odysseus `PLATFORM.md` §11 last line points the plan at the `bonsai-design` workspace — stale; now `Bonsai/station/Playbook/Plans/Active/40-...`.
- odysseus §4 says manifest at `.odysseus/project.yaml`; bonsai writes `.bonsai/project.yaml`. Hub discovery must check both paths.

---

## Grilling Pass — 2026-06-13

### Round 1 (6 critics: security, architecture, simplicity, risk, verification, reality)

| Critic | Verdict | Highest |
|--------|---------|---------|
| Security | concerns | high |
| Architecture | **block** | block |
| Simplicity | concerns | concern |
| Risk | **block** | block |
| Verification | concerns | near-block |
| Reality | concerns | concern |

**Root causes → resolutions:**
1. "Reuse `catalog_snapshot.go`" was false (it's a separate, not-lock-tracked writer; `Scaffolding()` forces `docs_path` prefix). → **`root_relative` flag + lock-aware `writeFile`**, borrow only the O_NOFOLLOW idiom. (user fork)
2. `bonsai update` never delivers new scaffolding items. → **teach update/add to deliver** (Phase 1c). (user fork)
3. `validate.Run()` is agent-scoped + top-level-only. → **new project-level recursive pass** (Phase 2 re-scoped). (auto)
4. Path-traversal / YAML-injection on `slug`/`memory_dir`/relations/scalars. → **input-validation decision** + Security section. (auto)
5. `affects:` leaked "Odysseus" into shipped catalog. → **host-agnostic phrasing**. (auto)
6. `log/` had no producer + contradicted "FieldNotes legacy". → **dropped from rung 1**. (auto)
7. `generate.go` howToWorkLines contradicted new routing. → **Phase 3 updates it**. (auto)
8. Verification rigor (slug algo, MEMORY.md content, negative-control + valid-tree fixtures, rollback, field tables, date func, one guide page, lock-tracking wording). → **all folded in**. (auto)

**Round 2:** pending re-grill on this edited plan (convergence loop — resolutions are unreviewed design work).
