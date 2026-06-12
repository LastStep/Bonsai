---
tags: [plan, integration, odysseus]
description: Build plan — everything bonsai must provide for the Odysseus personal-platform integration. Source decisions in odysseus repo PLATFORM.md §11.
status: active
source: odysseus-design-session-2026-06-12
---

# Plan — Odysseus Platform Integration

## Context

Odysseus (`/home/rohan/Apps/odysseus`, private fork) is being rebuilt into a personal platform: module system, categorized sidebar, two-tier AI memory, unified Projects, custom tools. Full spec: `PLATFORM.md` in that repo. The division of responsibilities between bonsai and Odysseus was decided 2026-06-12 (PLATFORM.md §11):

> **Bonsai = repo-side standard.** Defines, scaffolds, and lints everything that lives inside a project repo. Host-agnostic — must keep working with bare Claude Code, no Odysseus required.
> **Odysseus = hub runtime.** Indexes repos, aggregates cross-project, renders UI, owns global memory + orchestration. Writes into repos only via defined bridge actions.

Confirmed boundary decisions that drive this plan:
1. **Project memory lives in-repo** (`station/Memory/`), bonsai-scaffolded — NOT centrally in Odysseus. Odysseus only builds derived indexes from it.
2. **Bonsai is the schema authority** for all repo-resident formats (memory note schema, project manifest, Playbook formats). Odysseus implements parsers against bonsai's published standard.
3. **Plan write-back**: Odysseus writes approved chat plans into `station/Playbook/Plans/Active/` — bonsai defines the file format so those plans validate.

## Workstreams

### A. Project manifest standard (`project.yaml`)
Define and own the schema for a per-repo project manifest consumed by the Odysseus Projects module:
- Location: `.odysseus/project.yaml` (works for non-bonsai repos too; Odysseus creates it there for plain repos using the same schema).
- Fields (minimum): `name`, `slug`, `status` (idea|active|paused|done|archived), `next_action`, `tags`, `description`, `links` (repo/docs/issues URLs), `created`, optional `memory_dir` override.
- Bonsai: new scaffolding item generates it on `init`; `bonsai update` refreshes derivable fields without clobbering hand-edits (lock-file provenance, same as existing generated files); `bonsai validate` lints it.
- Document the schema in bonsai docs as the canonical reference.

### B. Memory scaffolding (`station/Memory/`)
New scaffolding item `memory`:
- Structure: `station/Memory/MEMORY.md` (capped index, ~200 lines — always-in-context budget) + `decisions/` + `notes/` + `log/YYYY-MM.md`.
- Canonical note schema (this becomes the bonsai standard; Odysseus indexes exactly this):

```markdown
---
title: <note title>
type: decision | note | fact | log
permalink: <stable-slug>
tags: []
scope: project/<slug>
valid_from: 2026-06-12        # optional event time
superseded_by: null            # set instead of deleting — never hard-delete facts
---
## Observations
- [category] one fact per bullet #tag (optional context)

## Relations
- relation_type [[Target Note]]
```

- Notes = graph nodes; typed `[[wikilinks]]` = edges; observations are individually indexable facts. Forward references to not-yet-existing notes are legal.
- Update `NoteStandards.md` template to document this schema as THE project note standard.
- Decide migration story for existing `Logs/KeyDecisionLog.md` + `FieldNotes.md`: either evolve those templates to emit schema-conformant entries, or document them as legacy free-form (recommend: evolve — decisions belong in `Memory/decisions/`).

### C. Schema authority documentation
- Bonsai docs/guides become the canonical format spec for: note schema (B), project manifest (A), Playbook file formats, plan file format (G). Ship via `bonsai guide` + templates.
- This folds into the existing "documentation suite" backlog item — the integration formats are now part of that scope.

### D. `bonsai validate` extensions
Lint, in addition to current checks:
- Memory notes: required frontmatter fields, `scope` matches project, unresolved `[[relations]]` reported (warning, not error), `superseded_by` targets exist.
- `MEMORY.md` index over budget (>200 lines) → warning.
- `project.yaml`: schema validation.
- Plan files in `Plans/Active/`: format conformance (G).

### E. graphify integration (repo side)
- New catalog item (sensor or routine): installs a git post-commit hook running `graphify --update` + a config stub; documents running `python -m graphify.serve` (MCP) so the hub and agents can query the code graph.
- graphify = github.com/safishamsi/graphify. Bonsai only wires the repo side; Odysseus consumes the MCP/`graph.json`.

### F. Export bridge (`bonsai export`)
New command emitting stable JSON for hub ingest (supersedes the older "export bridge + catalog item" backlog idea):
- Contents: parsed `project.yaml`, agents/roles with workspaces + abilities (from `.bonsai.yaml` + lock), paths to memory dir / Playbook files / plans, scaffolding inventory, validate summary.
- Contract: versioned (`"export_version": 1`), additive evolution only. Odysseus's project re-sync calls this instead of parsing `.bonsai.yaml` ad hoc (its current `_roles_from_bonsai_yaml` approach).

### G. Plan file format (write-back contract)
Define the md format for `station/Playbook/Plans/Active/*.md` so externally-written plans (Odysseus chat plan-mode) validate:
- Frontmatter: `title`, `status` (active|done|abandoned), `created`, `source` (agent|odysseus-chat|human), optional `session` (chat session id), `project`.
- Body: goal + checklist steps (`- [ ]`). Archive convention: move to `Plans/Archive/` on completion.

## Out of scope for bonsai (Odysseus owns — do not build)
UI of any kind; cross-project indexes/search; global (user-level) memory tier; embeddings/graph index infrastructure; dispatch runtime; secrets; connections; scheduling.

## Sequencing (hub dependency order)
1. **B + C + D (memory schema + docs + lint)** — blocks Odysseus phase 2 (Memory module).
2. **A + F (manifest + export)** — blocks Odysseus phase 3 (Projects module + self-hosting).
3. **E (graphify)** — wanted by phase 3 project pages.
4. **G (plan format)** — blocks Odysseus phase 4 (Plans view write-back).

## Open items for this workspace to resolve
- Exact `project.yaml` field list + types (propose, then sync with Odysseus side before freezing).
- KeyDecisionLog/FieldNotes evolution vs legacy status.
- Whether `bonsai init` gains a `--minimal` profile for non-code projects (pure note/PM repos).
