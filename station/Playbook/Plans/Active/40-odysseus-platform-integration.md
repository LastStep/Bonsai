---
tags: [plan, integration, odysseus]
description: Bonsai-side build plan for the Odysseus personal-platform integration. Rung 1 (B+C+D+A) decisions locked 2026-06-13.
status: active
source: odysseus-design-session-2026-06-12; decisions-grill 2026-06-13
---

# Plan 40 — Odysseus Platform Integration

**Tier:** 2
**Status:** Active (Rung 1 specced + locked)
**Agent:** code agents (gp) via worktree, tech-lead orchestrates

## Goal

Ship the bonsai-side repo standards the Odysseus hub consumes: in-repo memory graph (`station/Memory/`), per-repo project manifest (`.bonsai/project.yaml`), the canonical schemas documented as the bonsai standard, and `validate` lint for both. "Done" for Rung 1 = a bonsai project scaffolds a schema-conformant memory tree + manifest, `bonsai validate` lints them, `bonsai guide` documents the formats, and our own station dogfoods it — shipped as **v0.5.0**.

## Context

Odysseus (`/home/rohan/Apps/odysseus`, private fork) is being rebuilt into a personal platform. Boundary decided 2026-06-12 (odysseus `PLATFORM.md` §11):

> **Bonsai = repo-side standard.** Defines, scaffolds, lints everything inside a project repo. Host-agnostic — keeps working with bare Claude Code, no Odysseus required.
> **Odysseus = hub runtime.** Indexes repos, aggregates cross-project, renders UI, owns global memory + orchestration. Writes into repos only via defined bridge actions.

Bonsai is the schema authority for all repo-resident formats; Odysseus implements parsers against the published standard. The schema in PLATFORM §4 is the **frozen, agreed contract** — the user owns the Odysseus implementation; this workspace builds the bonsai side directly against it (no handshake loop).

## Decisions Locked (2026-06-13 grill)

| # | Decision | Rationale |
|---|----------|-----------|
| Scope | **Rung 1 = B + C + D + A.** Slug requirement pulled A (manifest) into rung 1. | Note `scope: project/<slug>` needs an authoritative slug source; manifest is it. |
| Release | **v0.5.0** minor bump at end of rung 1. | Additive: 2 new scaffolding items + validate extensions + guide pages. |
| Dispatch | **PR-flow code agents via worktree.** | Schema/lint = correctness-heavy; not UX-iteration. |
| Dogfood | **Yes — our station gets `station/Memory/` + `.bonsai/project.yaml`.** | Bonsai is project #1, mirrors Odysseus self-hosting. |
| Manifest location | **`.bonsai/project.yaml`** (NOT `.odysseus/` as the source plan said). | Host-agnostic hard rule — bonsai never hardcodes a consumer's name. Bonsai owns `.bonsai/` (catalog.json precedent). Hub checks two paths: `.bonsai/` for bonsai repos, `.odysseus/` for plain ones. |
| Manifest refresh | **Write-once + validate-only.** No field-merge machinery. | File barely changes; bonsai never re-edits after scaffold. You/hub make rare edits; validate lints. |
| Config split | **`.bonsai.yaml` and `project.yaml` stay separate.** | Different owners/lifecycles/consumers; AI team is opt-in, manifest is universal. |
| Manifest fields | Frozen v1 (below). **`next_action` REMOVED.** | Only volatile field + duplicated `Status.md`; hub reads status from manifest, next-action context from `Status.md` for bonsai repos. |
| Memory routing | **Decisions → `Memory/decisions/` graph; FieldNotes stays legacy free-form; `agent/Core/memory.md` untouched.** | Decisions are high-value queryable facts the hub should index. FieldNotes = human scratch, schema buys nothing. Working memory = always-loaded runtime scratchpad, different purpose. |
| Migration | **Scaffold now, migrate existing KeyDecisionLog entries later** (separate follow-up). | Don't block the release on hand-conversion. New decisions go to the graph; back-fill is its own task. |
| Note schema | **Freeze §4 shape + add `schema_version`.** | Migration path when schema evolves; validate pins lint rules to declared version. |
| Versioning | `schema_version` on **both** note frontmatter and manifest. | Bonsai-as-authority → every published format is versioned. |
| MEMORY.md budget | **Warning at >200 lines** (not hard error). | Context-window guideline, not a correctness rule. |
| Docs target | **`bonsai guide` pages + scaffolded templates.** No website. | Agent + hub-implementer audience; ships in binary; no Astro/MDX work. |
| graphify (E) | **Deferred behind rungs 1–2 + trust-vet first** (inspect_swe playbook: pin, fork, bus-factor). | Third-party, Python dep, new git-hook install mechanism. |
| Plan format (G) | **Deferred** — rung 4 (Odysseus phase 4 dependency). | Not needed for memory/manifest. |
| `--minimal` profile | **Deferred.** | Odysseus creates manifests for plain repos itself; low urgency. |

### Open defaults (flag to redirect, else taken)

- **Both new scaffolding items (`memory`, `project-manifest`) ship `required: false` (opt-in via init picker).** Non-breaking for existing projects; we opt in for the dogfood. Promote to required in a later version if desired.

## Frozen schemas (v1)

### Memory note — `station/Memory/{decisions,notes}/<permalink>.md`

```markdown
---
schema_version: 1
title: <note title>
type: decision | note | fact | log
permalink: <stable-slug>          # permanent node id; survives title change
tags: []
scope: project/<slug>             # <slug> from project.yaml (A)
valid_from: 2026-06-13            # optional event time
superseded_by: null               # set instead of deleting — never hard-delete facts
---
## Observations
- [category] one fact per bullet #tag (optional context)

## Relations
- relation_type [[Target Note]]
```

Notes = graph nodes; typed `[[wikilinks]]` = edges; observations individually indexable. Forward refs to not-yet-existing notes are legal.

### Project manifest — `.bonsai/project.yaml` (repo root)

```yaml
schema_version: 1
name: <project name>
slug: <stable-slug>               # authoritative slug source for note scope
status: idea | active | paused | done | archived
tags: []
description: <one-line>
links:
  repo: <url>
  docs: <url>
  issues: <url>
created: 2026-06-13
memory_dir: station/Memory        # optional override; default station/Memory
```

(No `next_action` — removed. Hub reads `status` here; richer next-action lives in `Status.md` for bonsai repos.)

## Rung 1 — Implementation Phases

> Dispatch order: **Phase 1 first** (it freezes the schemas), then **Phases 2 + 3 in parallel** (file-disjoint: `internal/validate` vs catalog docs). Phase 1 bundles A+B because both touch `catalog/scaffolding/manifest.yaml` + `internal/generate` — one agent avoids the shared-file conflict.

### Phase 1 — Scaffolding (A + B) — `gp`, one worktree

1. **`project-manifest` scaffolding item (A).**
   - Add item to `catalog/scaffolding/manifest.yaml`: `name: project-manifest`, `required: false`, `affects:` (Odysseus hub ingest, validate, project identity), `files: [.bonsai/project.yaml.tmpl]`.
   - Template renders the frozen manifest above. `name`←`{{ .ProjectName }}`, `slug`←slugified ProjectName, `description`←`{{ .ProjectDescription }}`, `created`←render date, others empty/placeholder.
   - **Generator root-relative write:** scaffolding files normally land under `docs_path`; this one targets repo-root `.bonsai/`. Reuse the path mechanism `internal/generate/catalog_snapshot.go` uses to write `.bonsai/catalog.json`. Write-once (skip if exists) — same as existing scaffolding write-once rule.
2. **`memory` scaffolding item (B).**
   - Add item to `catalog/scaffolding/manifest.yaml`: `name: memory`, `required: false`, `affects:` (Odysseus Memory module index, validate, NoteStandards), `files:` for the tree below.
   - Scaffold `{docs_path}/Memory/`: `MEMORY.md.tmpl` (capped index starter, ~200-line budget noted in a header comment), `decisions/.gitkeep`, `notes/.gitkeep`, `log/.gitkeep`.
   - `MEMORY.md.tmpl` documents its own purpose + links the note schema inline so it's self-describing.
3. Verify `bonsai init` (non-interactive + interactive) emits both items when selected; `.bonsai-lock.yaml` tracks them; re-run is conflict-clean.

### Phase 2 — Validate extensions (D) — `gp`, depends on Phase 1 schemas

In `internal/validate/validate.go` (+ tests):
1. **Memory notes** (`{memory_dir}/**/*.md`): required frontmatter present (`schema_version,title,type,permalink,scope`); `scope` matches `project/<slug>` from manifest; `superseded_by` target resolves to an existing note; unresolved `[[relations]]` → **warning** (not error — forward refs are legal).
2. **`MEMORY.md`** > 200 lines → **warning**.
3. **`.bonsai/project.yaml`**: schema validation — required fields, `status` enum, `schema_version` recognized.
4. Honor `--json` + `--agent` flag shapes already in validate. New issue categories registered with severities (error vs warning) consistent with existing ones.

### Phase 3 — Docs / standards (C) — `gp`, parallel with Phase 2

1. **`NoteStandards.md.tmpl`** (`catalog/scaffolding/...`): add the memory note schema as THE project note standard. Keep the existing tracker brevity rule; add a "Memory notes" section documenting frontmatter + Observations/Relations + supersession.
2. **`agent/Protocols/memory.md` (catalog source):** document the three-surface routing — working memory (`agent/Core/memory.md`, always-loaded scratch) vs durable graph (`station/Memory/`, decisions+facts, hub-indexed) vs Logs (`KeyDecisionLog` legacy/read-only, `FieldNotes` free-form). State where each kind of write goes.
3. **`bonsai guide` page(s):** add canonical-format guide content for the note schema + project manifest (rendered by guideflow). One combined "Formats" page or one per schema.
4. Update root `CLAUDE.md` + `station/CLAUDE.md` nav only if new files need routing entries.

## Dependencies

- Phase 1 freezes the on-disk schemas → Phases 2 and 3 depend on it.
- No new Go module deps. No external services.
- graphify (E), export bridge (F), plan format (G) are later rungs — not blockers.

## Security

> [!warning]
> Refer to [Playbook/Standards/SecurityStandards.md](../../Standards/SecurityStandards.md) for all requirements.

- No secrets in templates, schemas, or examples.
- Manifest/notes are plain YAML/markdown — validate must not execute or eval any content.
- Root-relative write (`.bonsai/`) must stay inside the repo; reuse the existing path-safety of `catalog_snapshot.go` (no traversal outside project root).

## Verification

- [ ] `bonsai init` (interactive + `--non-interactive`) scaffolds `.bonsai/project.yaml` + `station/Memory/` when items selected; both tracked in lock file; re-run conflict-clean.
- [ ] Generated `project.yaml` validates against the frozen v1 schema; `created`/`name`/`slug`/`description` populated from context.
- [ ] `bonsai validate` flags: missing note frontmatter (error), scope mismatch (error), dangling `superseded_by` (error), unresolved relation (warning), MEMORY.md >200 lines (warning), malformed manifest (error). `--json` output well-formed.
- [ ] `bonsai guide` renders the note-schema + manifest format pages.
- [ ] Dogfood: run on this repo — `.bonsai/project.yaml` + `station/Memory/` created, `./bonsai validate` exits clean (0 errors).
- [ ] `GOOS=windows GOARCH=amd64 go build ./...` passes (cross-compile gate — root-relative write must not use POSIX-only syscalls; recall v0.4.0 `O_NOFOLLOW` class).
- [ ] CI green; CHANGELOG entry; v0.5.0 tag.

## Out of scope for bonsai (Odysseus owns — do not build)

UI; cross-project indexes/search; global (user-level) memory tier; embeddings/graph index infrastructure; dispatch runtime; secrets; connections; scheduling.

## Later rungs (not this release)

- **Rung 2:** F (`bonsai export` stable JSON contract — supersedes `_roles_from_bonsai_yaml` in odysseus `workshop/ingest.py:110`). Manifest already lands in rung 1, so F is mostly the export command.
- **Rung 2.5:** Multi-agent `--from-config` for `bonsai init` (odysseus `workshop/lifecycle.py` currently splits init + sequential `add` as a workaround). File as backlog.
- **Rung 3:** E (graphify repo-side wiring) — vet first.
- **Rung 4:** G (plan write-back format) — Odysseus phase 4.

## Cross-repo follow-ups (outside this workspace)

- odysseus `PLATFORM.md` §11 last line points the plan at the `bonsai-design` workspace — stale; now `Bonsai/station/Playbook/Plans/Active/40-...`. Fix next odysseus session.
- odysseus memory doctrine §4 says manifest at `.odysseus/project.yaml`; bonsai now writes `.bonsai/project.yaml`. Hub discovery must check both paths. Reconcile §4/§5 wording odysseus-side.
