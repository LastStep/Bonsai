---
tags: [plan, integration, odysseus]
description: Bonsai-side build plan for the Odysseus personal-platform integration. Rung 1 = B+C+D+A. Grilled rounds 1–2 (2026-06-13).
status: active
source: odysseus-design-session-2026-06-12; decisions-grill 2026-06-13
tier: 2
---

# Plan 40 — Odysseus Platform Integration

**Tier:** 2 · **Status:** Phases 1–3 SHIPPED on main (v0.5.0, untagged — PRs #114/#116/#115). Phase 4 HELD (→ headless-CLI plan). Dogfood deferred + tag held (2026-06-13). · **Agent:** code agents (gp) via worktree

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
| Manifest writer = **plain `writeFile`, lock-tracked, write-once** | (R2) NOT O_NOFOLLOW — no existing scaffolding write uses it; `root_relative` items must bypass `Scaffolding()`'s pre-`os.Stat` skip so the write reaches `writeFile`+`lock.Track`. Pre-existing untracked manifest → conflict prompt. On re-run, reuse the existing manifest's `created` (emit live date only when absent) so bytes stay stable → `Unchanged`. Broader symlink hardening → Backlog. |
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

> Order: **Phase 1** first (freezes schemas). Then **Phases 2 + 3 in parallel** (`internal/validate/` vs docs). **Phase 4 (delivery, minimal opt-in) lands after Phase 1 — grilled, dispatchable.** All `catalog/scaffolding/` edits live in Phase 1's worktree.

### Phase 1 — Generator + catalog (A + B) — `gp`, one worktree
**1a. `RootRelative` support + manifest item.**
- Add `RootRelative bool \`yaml:"root_relative"\`` to `ScaffoldingItem` (`internal/catalog/catalog.go`).
- Rework `Scaffolding()` (`internal/generate/generate.go`): replace the flat `allowedFiles map[string]bool` with `map[string]*catalog.ScaffoldingItem`; in the walk choose the prefix from `item.RootRelative` (root-relative → skip the `cfg.DocsPath` join); apply the same per-item prefix in the empty-dir creation loop. For root-relative items, **bypass the pre-`os.Stat` write-once skip** so the write reaches `writeFile` (lock-tracked, write-once via lock, conflict-on-untracked).
- Add `project-manifest` to `catalog/scaffolding/manifest.yaml`: `required: false`, `root_relative: true`, **host-agnostic `affects:`** (e.g. "downstream hub ingest / repo indexers"; never "Odysseus"), `files: [.bonsai/project.yaml.tmpl]`. Plain `writeFile`; no O_NOFOLLOW.
- Template renders frozen manifest. `name`←`{{ .ProjectName }}`, `slug`←new slugify helper, `description`←`{{ .ProjectDescription }}`, `created`←injectable date source (funcMap/context field; `YYYY-MM-DD`). **On re-run, if the manifest exists, read+reuse its `created`** (live date only when absent) — bytes stay stable → `Unchanged`, preserves created=first-seen. Emit user scalars via `yaml.Marshal`. (`missingkey=error` is net-new on the shared `renderTemplate` — optional belt-and-suspenders; the real backstop is validate's non-empty `name`/`slug`.)

**1b. `memory` scaffolding item.**
- `manifest.yaml`: `required: false`, host-agnostic `affects:`, `files:` for: `MEMORY.md.tmpl`, `Memory/decisions/.gitkeep`, `Memory/notes/.gitkeep` (trailing-slash dir entries or explicit `.gitkeep` paths to satisfy `isAllowedScaffoldingFile`). **No `log/`.**
- `MEMORY.md.tmpl` (≤200 lines): `# {{ .ProjectName }} — Memory Index` + purpose blurb + empty `## Decisions`/`## Notes` index + link to the note-schema guide.

**1c. NoteStandards + tests.**
- **Extend the existing** `catalog/scaffolding/Playbook/Standards/NoteStandards.md.tmpl` (shipped by `playbook` — edit, don't create) with the memory-note schema as THE project note standard.
- Tests (`internal/generate/generate_test.go`): manifest render (fixed date) + slug fixtures incl. edge cases `"Café Foo!"`→`caf-foo` (or chosen ASCII rule), `"!!!"`→error, `"123 Go"`→`123-go`; root-relative path asserts repo-root `.bonsai/project.yaml` **NOT** `station/.bonsai/`; lock entry present with source `scaffolding:.bonsai/project.yaml.tmpl`; memory tree; idempotent re-run; YAML-injection negative (`description: 'evil: "}\n!!x'` round-trips via `yaml.Unmarshal`).

### Phase 2 — Validate project-level pass (D) — `gp`, depends on Phase 1
- Add a **project-level pass** to `validate.Run()` run **regardless of `agentFilter`** on the non-error path (an unknown-agent filter still errors early); its `Issue`s carry empty `AgentName` (`omitempty`) and are intentionally absent from `AgentsScanned`. Register new **additive** `Category` constants. New typed structs for note frontmatter + manifest (`yaml.Unmarshal`; do NOT extend `CustomItemMeta`).
- Parse `.bonsai/project.yaml` for `memory_dir`+`slug`. Manifest absent but `Memory/` present → **warning** (scope unverifiable), skip scope-match, **still lint frontmatter**.
- Recursive walk `{memory_dir}/**` (bounded file-count/size). Per note: missing required frontmatter → **error**; `scope` ≠ `project/<slug>` → **error**; bad `schema_version` (≠1) → **error**; out-of-charset `permalink` → **error**; dangling non-null `superseded_by` → **error**; unresolved `[[relation]]` → **warning**. Target resolution: `EvalSymlinks` + abs-prefix-under-`memory_dir`, reject escapes; index by sanitized permalink.
- `MEMORY.md` > 200 lines → **warning**. Manifest schema: required fields, `status` enum, `schema_version == 1` → **error**; `memory_dir` traversal (`../…`, absolute) → **error**.
- Tests (`internal/validate/validate_test.go`): a **fixture↔rule↔(category,severity) table**, one negative control per rule above incl. `schema_version:2`, `status:bogus`, `memory_dir:../escape`, manifest-absent-warning (assert warning fires AND frontmatter errors still surface), **plus a valid-tree fixture → zero issues**. Unit tests assert `report` contents (category/severity), not exit code.

### Phase 3 — Docs (C) — `gp`, parallel with Phase 2
- `agent/Protocols/memory.md`: document three-surface routing (working memory / durable graph / Logs).
- `internal/generate/generate.go` `howToWorkLines` (~L597): update the "Decision logging → KeyDecisionLog" line so it doesn't contradict the new routing.
- `bonsai guide`: one combined "Formats" page (note schema + manifest); gate asserts body contains `schema_version` + `permalink`.

### Phase 4 — Delivery via `bonsai update` (minimal opt-in) — `gp`, depends on Phase 1
> Grilled 2026-06-13 (5 critics) → re-scoped to minimal. **Dispatchable.** No TUI picker, no nonint surgery.
- Add `generate.Scaffolding(cwd, cfg, cat, lock, &wr, false)` to update's sync action (`internal/tui/updateflow/run.go` `buildSyncAction`), reading `cfg.Scaffolding` (drop-in: same locals already in the sync closure). Mutate-then-render ordering.
- **Opt-in is manual + explicit:** the user adds the item name to `.bonsai.yaml`'s `scaffolding:` list, then runs `bonsai update`. Set `ConfigChanged=true` when `cfg.Scaffolding` changes so `cmd/update.go` persists it (today nothing flips it for a scaffolding-only delta). Document the one-line opt-in in the Phase-3 guide.
- **No `RunStatic` auto-delivery:** `RunStatic` renders only items already in the on-disk `.bonsai.yaml`; it must never auto-add newly-available catalog items (→ no silent write into user repos in cron/CI/piped). **Drop the `nonint/runner.go:222` guard reconciliation** — that guard is `add --from-config` only; `update` never reaches it.
- Conflict/skip is inherited: root-relative **manifest** → conflict-on-untracked (`writeFile`); **memory tree** → write-once `os.Stat` skip. `Scaffolding(..., force=false)` always (no `--force` on update).
- Config + lock writes prefer temp-file+rename; don't mark an item installed if its write errored (avoid orphan lock/config on partial run).
- Tests: (1) unit on `generate.Scaffolding` (tempdir; item in `cfg.Scaffolding`, absent on disk → file written + lock entry source `scaffolding:.bonsai/project.yaml.tmpl`); (2) item in `.bonsai.yaml` → `update` delivers + `ConfigChanged` persists + re-read `.bonsai.yaml` shows it; (3) idempotent re-run → Unchanged, no dup lock entry; (4) collision fixture = **untracked manifest** w/ sentinel bytes → `RunStatic` returns `SyncErr`/`HasConflicts`, bytes byte-preserved.

## Dependencies
Phase 1 freezes schemas → Phases 2/3/4 depend on it. No new Go module deps; no external services.

## Rollback
- Each phase a revertable PR; pre-tag v0.5.0 is `git revert`-able. Dogfood scaffold in a separate PR (reverts independently).
- Manifest is lock-tracked → **regenerates on delete** (unlike os.Stat-skipped scaffolding): a broken manifest → delete + re-run regenerates it; a user edit → conflict, not clobber.
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
- [ ] `bonsai init` (interactive + `--non-interactive`) scaffolds `.bonsai/project.yaml` at **repo root** (assert NOT `station/.bonsai/`) + `station/Memory/{decisions,notes}/`; both **lock-tracked**; idempotent re-run → Skipped/Unchanged (manifest stays Unchanged across days via `created`-preservation), exit 0.
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

### Round 3 (Phases 1–3 only; 3 critics: Architecture, Risk, Reality)
**Reality: pass** (9 mechanisms verified feasible against source). **Risk + Architecture: concerns, no block** — both caught one real defect: live `created` re-render breaks idempotency (later-day re-run rewrites the timestamp + fails the Unchanged claim). → **Fix: reuse the existing manifest's `created` on re-run.** Wording tightened: validate project-pass runs regardless of `agentFilter` (non-error path), intentionally absent from `AgentsScanned`; rollback notes manifest regenerates-on-delete. R2 blocks confirmed closed.

**Converged: 0 blocks across the round; single concern resolved + Reality-verified.** Phases 1–3 **LOCKED, ready for dispatch.**

### Phase 4 grill (5 critics: Architecture, Risk, Simplicity, Verification, Reality)
Risk + Verification **block**; Simplicity scope-**block**. All 5 converged: (1) the cited `nonint/runner.go:222` guard is `add --from-config` only — `update` has no nonint path (`RunStatic`), so **drop it**; (2) wiring delivery into `RunStatic` = silent un-opted-in write → **forbid auto-add**; (3) the cinematic picker is **gold-plating** for a ~2-repo migration. **Re-scoped to minimal** (user): one `generate.Scaffolding()` call in update's sync action + manual `.bonsai.yaml` opt-in + `ConfigChanged` flip + inherited conflict/skip + 4 small tests. Reality verified the minimal path (cfg.Save already wired, `generate.Scaffolding` drop-in callable, `SoilStage` exists if a picker is ever wanted). **Phase 4 LOCKED, dispatchable.**

> **Note (2026-06-13):** user then raised a larger goal — make **init/update/add/remove all fully agent-drivable (non-interactive)** so agents drive Bonsai without the TUI. That supersedes Phase 4's narrow update-delivery slice and is captured as its own workstream (Backlog P1 → next-session `/plan`). Plan 40's scope is unchanged; the headless-CLI plan will own the broader non-interactive parity.

> **Dispatch decision (2026-06-13):** user chose to **HOLD Phase 4** at dispatch time — ship **v0.5.0 = Phases 1–3 only**; the headless-CLI plan will own update-delivery (avoids throwaway). Phase 1 dispatched first (freezes schemas); Phases 2 + 3 fan out in parallel after Phase 1 merges.
