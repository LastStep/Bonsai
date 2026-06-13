---
tags: [plan, integration, odysseus]
description: Bonsai-side build plan for the Odysseus personal-platform integration. Rung 1 = B+C+D+A. Grilled rounds 1–2 (2026-06-13).
status: active
source: odysseus-design-session-2026-06-12; decisions-grill 2026-06-13
tier: 2
---

# Plan 40 — Odysseus Platform Integration

**Tier:** 2 · **Status:** Active — grilling (round 2 applied; round 3 pending on Phases 1–3) · **Agent:** code agents (gp) via worktree

## Goal

Ship the bonsai-side repo standards the Odysseus hub consumes: in-repo memory graph (`station/Memory/`), per-repo project manifest (`.bonsai/project.yaml`), the canonical schemas documented as the bonsai standard, `validate` lint for both, and delivery of the new items to existing projects via `bonsai update`. Shipped as **v0.5.0**, dogfooded on our own station.

## Context

Odysseus (`/home/rohan/Apps/odysseus`, private fork) → personal platform. Boundary (odysseus `PLATFORM.md` §11): **Bonsai = repo-side standard** (defines/scaffolds/lints repo-resident formats; host-agnostic, works with bare Claude Code); **Odysseus = hub runtime** (indexes/aggregates/renders; writes into repos only via bridge actions). Bonsai is schema authority; the PLATFORM §4 schema is the frozen, agreed contract; the user owns the Odysseus side.

## Decisions Locked (2026-06-13)

| Decision | Note |
|----------|------|
| Scope = **B + C + D + A** | slug needs an authoritative source → manifest in rung 1 |
| Release = **v0.5.0** | additive |
| Dispatch = **PR-flow worktree agents** | correctness-heavy |
| Dogfood = **yes**, separate commit/PR | reverts independently |
| Manifest location = **`.bonsai/project.yaml`** | host-agnostic; hub checks both paths |
| Manifest writer = **plain `writeFile`, lock-tracked, write-once** | (R2) NOT O_NOFOLLOW — no existing scaffolding write uses it; `root_relative` items must bypass `Scaffolding()`'s pre-`os.Stat` skip so the write reaches `writeFile`+`lock.Track`. Pre-existing untracked manifest → conflict prompt (normal lock behavior). Broader symlink hardening of all scaffolding writes → Backlog. |
| RootRelative wiring = **`map[string]*ScaffoldingItem`** | (R2) `Scaffolding()`'s flat `map[string]bool` discards item identity; walk must read `item.RootRelative` to skip the `docs_path` join. Dir-creation loop uses the same per-item prefix. |
| Delivery = **`bonsai update`, own phase (Phase 4)** | (R2) net-new feature: new scaffolding-picker stage + `cfg.Scaffolding` mutation + `Scaffolding()` call site + non-interactive guard reconciliation. `add` is agent-scoped → wrong home. **Phase 4 needs its own grilling pass before dispatch.** |
| Config split = **separate**; `project.yaml` is hub-facing identity, `.bonsai.yaml` generator-facing | (R2) seeded once at init, **never reconciled** — documented known drift, not an enforced "authority". Optional divergence warning → Backlog. |
| Manifest fields | frozen v1 (table below); `next_action` removed |
| Memory routing | decisions → `Memory/decisions/`; FieldNotes legacy; `agent/Core/memory.md` untouched. Phase 3 also fixes `generate.go` `howToWorkLines` (L597) so it stops pointing decisions at KeyDecisionLog. |
| Migration | scaffold now, migrate KeyDecisionLog later |
| Schemas | freeze §4 + `schema_version: 1` (int, distinct from `CatalogSnapshot.Version` build-string) |
| Input validation | `slug`/`permalink` = `[a-z0-9-]` (out-of-charset permalink = **error**, never indexed); `memory_dir` via `wsvalidate.InvalidReason` (accidental-grade; trim `Normalise`'s trailing `/`); **note-target resolution is adversarial-grade** → `filepath.EvalSymlinks` + abs-prefix-under-`memory_dir` check, refuse symlinked notes/dirs; relations/`superseded_by` resolved by sanitized permalink in a trivial in-memory `map[string]bool` (never link-text-as-path); user scalars emitted via `yaml.Marshal` (not hand-rolled quoting). |
| MEMORY.md budget | warning > 200 lines |
| Docs | `bonsai guide` one combined "Formats" page + templates; no website |
| Deferred | graphify (E), export (F), plan-format (G), `--minimal` |
| Both new scaffolding items | `required: false` (opt-in) |

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
- [category] one fact per bullet #tag
## Relations
- relation_type [[Target Note]]
```

| field | type | required | rule |
|-------|------|----------|------|
| `schema_version` | int | ✅ | `== 1` |
| `title` | string | ✅ | non-empty |
| `type` | enum | ✅ | `decision\|note\|fact\|log` |
| `permalink` | string | ✅ | `[a-z0-9-]`; out-of-charset = error; unique in tree |
| `tags` | list[str] | ⬜ | default `[]` |
| `scope` | string | ✅ | `project/<slug>` == manifest slug |
| `valid_from` | date | ⬜ | `YYYY-MM-DD` if present; scaffolded notes emit it, hand-authored may omit |
| `superseded_by` | string\|null | ⬜ | **absent ≡ null ≡ not-superseded** (no missing-key error); if non-null must resolve to an existing permalink |

Forward refs to not-yet-existing notes are legal → **warning**, not error.

### Project manifest — `.bonsai/project.yaml` (repo root)

| field | type | required | rule |
|-------|------|----------|------|
| `schema_version` | int | ✅ | `== 1` |
| `name` | string | ✅ | non-empty |
| `slug` | string | ✅ | `[a-z0-9-]` |
| `status` | enum | ✅ | `idea\|active\|paused\|done\|archived` |
| `tags` | list[str] | ⬜ | |
| `description` | string | ⬜ | |
| `links` | map | ⬜ | present-optional; each of `repo\|docs\|issues` individually optional; **unknown keys ignored, not error**; URL content unvalidated |
| `created` | date | ✅ | `YYYY-MM-DD` |
| `memory_dir` | string | ⬜ | repo-relative, non-traversing; default `station/Memory` |

**Slug algorithm (frozen):** ASCII-only (strip non-ASCII), lowercase, replace each `[^a-z0-9]` run with single `-`, trim leading/trailing `-`. **Empty result = generator error** (`name` is required-non-empty; a name slugifying to empty is a real conflict). Leading digits allowed. Net-new helper in `internal/catalog` (no existing slugify; `DisplayNameFrom` is the inverse).

## Rung 1 — Implementation Phases

> Order: **Phase 1** first (freezes schemas). Then **Phases 2 + 3 in parallel** (`internal/validate/` vs docs). **Phase 4 (delivery) is separate, lands after Phase 1, and gets its own grilling pass before dispatch.** All `catalog/scaffolding/` edits live in Phase 1's worktree.

### Phase 1 — Generator + catalog (A + B) — `gp`, one worktree
**1a. `RootRelative` support + manifest item.**
- Add `RootRelative bool \`yaml:"root_relative"\`` to `ScaffoldingItem` (`internal/catalog/catalog.go`).
- Rework `Scaffolding()` (`internal/generate/generate.go`): replace the flat `allowedFiles map[string]bool` with `map[string]*catalog.ScaffoldingItem`; in the walk choose the prefix from `item.RootRelative` (root-relative → skip the `cfg.DocsPath` join); apply the same per-item prefix in the empty-dir creation loop. For root-relative items, **bypass the pre-`os.Stat` write-once skip** so the write reaches `writeFile` (lock-tracked, write-once via lock, conflict-on-untracked).
- Add `project-manifest` to `catalog/scaffolding/manifest.yaml`: `required: false`, `root_relative: true`, **host-agnostic `affects:`** (e.g. "downstream hub ingest / repo indexers"; never "Odysseus"), `files: [.bonsai/project.yaml.tmpl]`. Plain `writeFile`; no O_NOFOLLOW.
- Template renders frozen manifest. `name`←`{{ .ProjectName }}`, `slug`←new slugify helper, `description`←`{{ .ProjectDescription }}`, `created`←new **injectable** date source (funcMap/context field, fixed in tests; emits `YYYY-MM-DD` only). Emit user scalars via `yaml.Marshal`. (`missingkey=error` is net-new on the shared `renderTemplate` — optional belt-and-suspenders; the real backstop is validate's non-empty `name`/`slug`.)

**1b. `memory` scaffolding item.**
- `manifest.yaml`: `required: false`, host-agnostic `affects:`, `files:` for: `MEMORY.md.tmpl`, `Memory/decisions/.gitkeep`, `Memory/notes/.gitkeep` (trailing-slash dir entries or explicit `.gitkeep` paths to satisfy `isAllowedScaffoldingFile`). **No `log/`.**
- `MEMORY.md.tmpl` (≤200 lines): `# {{ .ProjectName }} — Memory Index` + purpose blurb + empty `## Decisions`/`## Notes` index + link to the note-schema guide.

**1c. NoteStandards + tests.**
- **Extend the existing** `catalog/scaffolding/Playbook/Standards/NoteStandards.md.tmpl` (shipped by `playbook` — edit, don't create) with the memory-note schema as THE project note standard.
- Tests (`internal/generate/generate_test.go`): manifest render (fixed date) + slug fixtures incl. edge cases `"Café Foo!"`→`caf-foo` (or chosen ASCII rule), `"!!!"`→error, `"123 Go"`→`123-go`; root-relative path asserts repo-root `.bonsai/project.yaml` **NOT** `station/.bonsai/`; lock entry present with source `scaffolding:.bonsai/project.yaml.tmpl`; memory tree; idempotent re-run; YAML-injection negative (`description: 'evil: "}\n!!x'` round-trips via `yaml.Unmarshal`).

### Phase 2 — Validate project-level pass (D) — `gp`, depends on Phase 1
- Add a **project-level pass** to `validate.Run()` run **unconditionally** (agent `--filter` narrows only the per-agent loop); its `Issue`s carry empty `AgentName` (`omitempty`). Register new **additive** `Category` constants. New typed structs for note frontmatter + manifest (`yaml.Unmarshal`; do NOT extend `CustomItemMeta`).
- Parse `.bonsai/project.yaml` for `memory_dir`+`slug`. Manifest absent but `Memory/` present → **warning** (scope unverifiable), skip scope-match, **still lint frontmatter**.
- Recursive walk `{memory_dir}/**` (bounded file-count/size). Per note: missing required frontmatter → **error**; `scope` ≠ `project/<slug>` → **error**; bad `schema_version` (≠1) → **error**; out-of-charset `permalink` → **error**; dangling non-null `superseded_by` → **error**; unresolved `[[relation]]` → **warning**. Target resolution: `EvalSymlinks` + abs-prefix-under-`memory_dir`, reject escapes; index by sanitized permalink.
- `MEMORY.md` > 200 lines → **warning**. Manifest schema: required fields, `status` enum, `schema_version == 1` → **error**; `memory_dir` traversal (`../…`, absolute) → **error**.
- Tests (`internal/validate/validate_test.go`): a **fixture↔rule↔(category,severity) table**, one negative control per rule above incl. `schema_version:2`, `status:bogus`, `memory_dir:../escape`, manifest-absent-warning (assert warning fires AND frontmatter errors still surface), **plus a valid-tree fixture → zero issues**. Unit tests assert `report` contents (category/severity), not exit code.

### Phase 3 — Docs (C) — `gp`, parallel with Phase 2
- `agent/Protocols/memory.md`: document three-surface routing (working memory / durable graph / Logs).
- `internal/generate/generate.go` `howToWorkLines` (~L597): update the "Decision logging → KeyDecisionLog" line so it doesn't contradict the new routing.
- `bonsai guide`: one combined "Formats" page (note schema + manifest); gate asserts body contains `schema_version` + `permalink`.

### Phase 4 — Delivery via `bonsai update` (own phase, own grill) — `gp`
> Specced but **NOT dispatchable until it passes its own grilling pass.**
- New scaffolding-item selection stage in `internal/tui/updateflow/` (today the Select stage is a custom-file promoter, agent-keyed — this is a new project-level picker).
- Mutate + persist `cfg.Scaffolding`; add a `generate.Scaffolding()` call in the sync action (`updateflow/run.go`); reconcile the non-interactive exact-match guard (`internal/nonint/runner.go:222-228`, which currently rejects a scaffolding diff vs disk).
- Conflict path: delivering into a project with a pre-existing untracked file at a target path → conflict prompt, **not silent clobber**.
- Tests: first-time delivery (item **absent→present** + new lock entry, assert source string); pre-existing-collision negative path; non-interactive delivery.

## Dependencies
Phase 1 freezes schemas → Phases 2/3/4 depend on it. No new Go module deps; no external services.

## Rollback
- Each phase a revertable PR; pre-tag v0.5.0 is `git revert`-able. Dogfood scaffold in a separate PR (reverts independently).
- Write-once: a broken scaffolded manifest needs manual delete to regenerate (skip-if-exists). Documented.
- Delivery (Phase 4): reverting the binary does **not** un-mutate a user's persisted `.bonsai.yaml`/lock; `required: false` makes the data inert; a partial write leaves orphan lock entries that `bonsai validate` flags. Documented.
- Deliberate asymmetry: manifest is lock-tracked; `catalog.json` is not (regenerated). Noted so a future reader doesn't "fix" it.

## Security
> Refer to [SecurityStandards.md](../../Standards/SecurityStandards.md).
- No secrets in templates/schemas/examples.
- Manifest via plain `writeFile` (same symlink exposure as all existing scaffolding — write-once skip mitigates overwrite; broader hardening + `MkdirAll` parent-dir TOCTOU → Backlog, not rung 1).
- Path safety: `slug`/`permalink` charset; `memory_dir` via `wsvalidate` (accidental-grade); **note-target resolution adversarial-grade** (`EvalSymlinks` + prefix check, refuse symlinks).
- Template injection: user scalars via `yaml.Marshal`, never bare `{{ . }}` into YAML.
- Parser safety: typed structs only, no eval, bounded walk.

## Verification
- [ ] `bonsai init` (interactive + `--non-interactive`) scaffolds `.bonsai/project.yaml` at **repo root** (assert NOT `station/.bonsai/`) + `station/Memory/{decisions,notes}/`; both **lock-tracked**; idempotent re-run → Skipped/Unchanged, exit 0.
- [ ] Generated manifest: `./bonsai validate --json` zero issues; `schema_version: 1`, non-empty `name`/`slug`, `created` matches `^\d{4}-\d{2}-\d{2}$`, `status` ∈ enum. Generator test asserts deterministic bytes (fixed date) + slug fixtures (incl. edge cases) + YAML-injection round-trip.
- [ ] `bonsai validate` fixture↔rule table: each rule's negative control asserts category+severity (incl. `schema_version`, `status` enum, `memory_dir` traversal, manifest-absent warning + frontmatter-still-linted); **valid-tree → zero issues**. `--json` well-formed.
- [ ] `bonsai guide` Formats page renders; body contains `schema_version` + `permalink`.
- [ ] **Phase 4:** `bonsai update` delivers a previously-unselected item into an existing project (absent→present + new lock entry w/ scaffolding source); pre-existing-collision → conflict path, not clobber.
- [ ] Dogfood (separate PR): this repo gets `.bonsai/project.yaml` + `station/Memory/`; `./bonsai validate` exits 0.
- [ ] `GOOS=windows GOARCH=amd64 go build ./...` passes (no inline POSIX syscall — plain `writeFile`).
- [ ] CI green; CHANGELOG; v0.5.0 tag.

## Out of scope (Odysseus owns)
UI; cross-project indexes/search; global memory tier; embeddings/graph infra; dispatch runtime; secrets; connections; scheduling.

## Later rungs
Rung 2: F (`bonsai export` — supersedes `_roles_from_bonsai_yaml`, odysseus `workshop/ingest.py:110`). Rung 2.5: multi-agent `--from-config` (odysseus `workshop/lifecycle.py`). Rung 3: E (graphify, vet first). Rung 4: G (plan write-back).

## Cross-repo follow-ups (odysseus side)
- PLATFORM.md §11 last line points at `bonsai-design` workspace — stale.
- §4 says `.odysseus/project.yaml`; bonsai writes `.bonsai/project.yaml` — hub discovery checks both.

---

## Grilling Pass — 2026-06-13

### Round 1 (6 critics) — Architecture + Risk block; 4 concerns
Root causes → resolutions: false "reuse catalog_snapshot" → `root_relative` + lock-aware writeFile; `update` never delivers → teach update/add; agent-scoped validate → project-level pass; path-traversal/YAML-injection → input-validation decision; `affects:` leaked Odysseus → host-agnostic; `log/` no producer → dropped; `howToWorkLines` contradiction → Phase 3; verification rigor → folded in.

### Round 2 (6 critics, on the round-1-revised plan) — Architecture + Risk block; 4 concerns
2 root causes:
- **A. O_NOFOLLOW vs writeFile (4 critics):** the two are mutually exclusive write paths; "borrow O_NOFOLLOW + route through writeFile" shipped dead code. → **Manifest via plain writeFile, no O_NOFOLLOW** (consistent with all existing scaffolding; write-once mitigates); `root_relative` bypasses the pre-`os.Stat` skip to reach `lock.Track`; broader hardening → Backlog.
- **B. RootRelative wiring + delivery (Arch+Risk+Reality):** flat `allowedFiles` map discards item identity → `map[string]*ScaffoldingItem`; `update`/`add` have no scaffolding picker/`Scaffolding()` call (net-new feature) → **delivery split to Phase 4, own grill** (user: keep delivery, own phase).
- Plus: `project.yaml` "authority" → documented as known drift (not enforced); gate-tightening (fixture↔rule table, `created` regex, slug edge cases, `superseded_by` absent≡null, `links` per-key, EvalSymlinks, permalink-as-error, unit-tests-assert-report-not-exit-code) → folded in.
Reality verified clean: wsvalidate.InvalidReason, ScaffoldingItem tags additive, funcMap date/missingkey net-new, isAllowedScaffoldingFile permits new entries, howToWorkLines line, lockfile Track, NoteStandards edit-not-create, Windows baseline green, no slugify, TemplateContext vars.

### Round 3 — pending on Phases 1–3 (Phase 4 grilled separately before dispatch).
