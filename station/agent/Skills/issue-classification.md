---
tags: [skill, issues]
description: Issue types, importance levels, domain labels, and classification heuristics for intake triage.
---

# Skill: Issue Classification

---

## Types

| Type | Description | Examples |
|------|-------------|----------|
| bug | Something is broken or behaves incorrectly | Template rendering error, wrong file paths in output, lock file hash mismatch |
| feature | New capability that doesn't exist yet | New CLI command, new catalog item type, new agent type |
| change | Modification to existing behavior | Change picker defaults, rename config fields, update template vars |
| debt | Technical cleanup with no user-facing change | Refactor generate.go, add tests for catalog loader, remove dead code |
| research | Investigation that produces findings, not code | Evaluate Managed Agents integration, benchmark TUI frameworks, design companion app |

---

## Importance

| Level | Meaning | Response |
|-------|---------|----------|
| critical | Blocks users, data loss risk, security vulnerability | Drop current work, fix immediately |
| high | Significant impact, degraded functionality | Next up, address promptly |
| medium | Important but not urgent, workaround exists | Schedule in current cycle |
| low | Nice to have, minor improvement | Backlog, address when convenient |

---

## Domain Labels

Domains represent which part of the Bonsai codebase is affected:

- `cli` — Cobra commands in `cmd/` (init, add, remove, list, catalog, update)
- `catalog` — catalog item definitions in `catalog/` (agents, skills, workflows, protocols, sensors, routines, scaffolding)
- `generator` — template rendering and file generation in `internal/generate/`
- `config` — project config and lock file handling in `internal/config/` (`.bonsai.yaml`, `.bonsai-lock.yaml`)
- `tui` — TUI forms, styling, and display in `internal/tui/` (Huh, LipGloss, BubbleTea)
- `catalog-loader` — catalog metadata loading in `internal/catalog/`
- `sensors` — hook scripts in `catalog/sensors/` (scope guards, context injection, status bar)
- `routines` — maintenance procedure templates in `catalog/routines/`
- `scaffolding` — project infrastructure templates in `catalog/scaffolding/`
- `docs` — station workspace docs, research files, design docs

An issue can span multiple domains. When it does, the plan should have separate step sections per domain with explicit sequencing.

---

## Classification Process

1. **Read the full issue** — title, body, comments, linked issues
2. **Identify the symptom** — what the user sees vs. what should happen
3. **Trace to root cause** — which code path is responsible
4. **Map domains** — which parts of the codebase need changes
5. **Assess importance** — based on user impact, not implementation effort
6. **Check for duplicates** — is this already tracked in Backlog.md or Status.md?
7. **Check for related items** — are there Backlog items that should be bundled?

---

## GitHub Label Mapping

**Repo:** `LastStep/Bonsai`

When working with GitHub Issues, map classifications to labels:

- **Type** → `bug`, `feature`, `change`, `debt`, `research`
- **Importance** → `critical`, `high`, `medium`, `low`
- **Domain** → `cli`, `catalog`, `generator`, `config`, `tui`, `sensors`, `routines`, `scaffolding`, `docs`
