---
description: The frozen v1 repo formats — the memory-note schema and the project manifest.
---

# Formats

Bonsai defines two repo-resident formats as a stable standard. They are
**frozen at v1**: the keys are fixed, so downstream hub ingest / repo indexers
can parse them directly. Fill in the values — don't rename the keys.

## Memory note

A durable memory note lives at `{memory_dir}/{decisions,notes}/<permalink>.md`
(default `memory_dir` is `station/Memory`). Each note is one markdown file with
frozen frontmatter plus an Observations / Relations body:

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
| `schema_version` | int | yes | must be `1` |
| `title` | string | yes | non-empty |
| `type` | enum | yes | one of `decision`, `note`, `fact`, `log` |
| `permalink` | string | yes | `[a-z0-9-]` only; out-of-charset is an error, never indexed; unique within the tree |
| `tags` | list[str] | no | default `[]` |
| `scope` | string | yes | `project/<slug>` — `<slug>` must match the manifest `slug` |
| `valid_from` | date | no | `YYYY-MM-DD` if present; scaffolded notes emit it, hand-authored may omit |
| `superseded_by` | string \| null | no | absent ≡ `null` ≡ not-superseded; if non-null must resolve to an existing permalink |

The `permalink` is the note's stable identity — keep it fixed even when the
title changes, since relations and `superseded_by` resolve by permalink.
A relation (or `superseded_by`) pointing at a not-yet-existing note is a
**warning**, not an error: forward references are legal while the graph is built
out.

## Project manifest

The project manifest lives at `.bonsai/project.yaml` (repo root) and is the
canonical repo-identity record:

```yaml
schema_version: 1
name: My Project
slug: my-project
status: idea
tags: []
description: One-line summary.
links: {}
created: 2026-06-13
memory_dir: station/Memory
```

| field | type | required | rule |
|-------|------|----------|------|
| `schema_version` | int | yes | must be `1` |
| `name` | string | yes | non-empty |
| `slug` | string | yes | `[a-z0-9-]` only |
| `status` | enum | yes | one of `idea`, `active`, `paused`, `done`, `archived` |
| `tags` | list[str] | no | |
| `description` | string | no | |
| `links` | map | no | each of `repo`, `docs`, `issues` optional; unknown keys ignored, not an error |
| `created` | date | yes | `YYYY-MM-DD` |
| `memory_dir` | string | no | repo-relative, non-traversing; default `station/Memory` |

The `slug` is the project's machine identity: it anchors the memory-note `scope`
(`project/<slug>`), so the two formats stay linked.

## Delivering the formats to an existing project

Both formats ship as **opt-in scaffolding** (`project-manifest` and `memory`),
selected during `bonsai init`. When you run `bonsai init` and tick them in the
scaffolding picker, they're written: the manifest at the repo root
(`.bonsai/project.yaml`) and the memory tree under `memory_dir`
(`{memory_dir}/decisions/`, `{memory_dir}/notes/`, default `station/Memory`).

To add them to a project that predates them, **re-run `bonsai init`** and select
`project-manifest` / `memory` in the scaffolding picker. A re-run is safe and
idempotent: the manifest is lock-tracked and the memory tree is write-once, so
any files that already exist are left untouched.

Run `bonsai validate` to lint both formats — it reports schema violations
(missing required fields, bad `schema_version`, scope mismatch, out-of-charset
permalinks) before downstream tooling ever sees them.

> Full reference: <https://laststep.github.io/Bonsai/>
