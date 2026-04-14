---
tags: [logs, decisions]
description: Architectural decisions grouped by relevance tier. Read Structural when planning; consult Domain-Specific and Settled only when that topic comes up.
---

# Bonsai — Key Decision Log

> [!note]
> Decisions recorded here are settled. Do not relitigate them unless the user explicitly reopens the topic.

---

## Structural

Decisions that shape how new features should be designed. Read when planning.

- **2026-04-12** — Rewrite from Python to Go. Rationale: single binary distribution via `go install`, proper TUI support via Charm ecosystem (BubbleTea, Huh, LipGloss), no runtime dependencies for users.
- **2026-04-12** — Catalog embedded via `embed.FS`. Rationale: ships inside the binary, no external file dependencies, version-locked to the build.
- **2026-04-12** — Templates use Go `text/template`, not Jinja2. Rationale: stdlib, no dependencies, sufficient for the rendering we need (variable substitution + conditionals).
- **2026-04-13** — Each workspace owns its own `CLAUDE.md` — no root-level generated CLAUDE.md. Rationale: root CLAUDE.md is hand-maintained project instructions; generated workspace CLAUDE.md is the routing table for that agent. Avoids conflicts between project docs and generated content.
- **2026-04-13** — Lock file (`.bonsai-lock.yaml`) tracks generated files with SHA-256 hashes. Rationale: enables conflict detection on re-run — distinguishes "user modified this file" from "file is unchanged since generation", supports skip/overwrite/backup workflow.
- **2026-04-13** — Tech-lead is required, always installed first during `bonsai init`, always lives in the station directory. Rationale: every project needs an architect agent; centralizing it in the docs directory keeps scaffolding co-located with the planning agent.

---

## Domain-Specific

Only matter when working in that area.

### Catalog Design

- **2026-04-13** — `required` field uses same format as `agents` (list or `"all"`). Rationale: consistency — same struct, same parsing logic, same compatibility check.
- **2026-04-13** — Sensors use `.sh.tmpl` scripts (not `.md`), routines use `.md.tmpl` content files. Rationale: sensors are executable hooks that run in the shell; routines are instruction documents the agent reads and follows.
- **2026-04-13** — `routine-check` sensor is auto-managed. Rationale: auto-installed when routines are present, auto-removed when the last routine is removed. No manual wiring needed — generator handles it via `EnsureRoutineCheckSensor()`.
- **2026-04-13** — Display names are derived, never stored in config. Rationale: `display_name` is purely cosmetic, auto-generated from `name` via `DisplayNameFrom()`. Storing it would create sync issues.

### Agent Design

- **2026-04-13** — Six agent types: tech-lead, fullstack, backend, frontend, devops, security. Rationale: covers the common roles in a software project without being so granular that users face paradox of choice.
- **2026-04-13** — Shared core files in `catalog/core/` with agent-level overrides. Rationale: memory.md and self-awareness.md are identical across agents; identity.md is agent-specific. Layered resolution (agent dir first, then shared) avoids duplication while allowing customization.

### Awareness Framework

- **2026-04-13** — Two-sensor design: status-bar (Stop) writes state file, context-guard (UserPromptSubmit) reads it. Rationale: decouples measurement from enforcement. Status bar runs after every response and writes metrics to `/tmp/`. Context guard reads those metrics before the next prompt and injects behavioral constraints.
- **2026-04-13** — Tiered context injection: advisory (30%), warning (50%), alert (70%), critical (85%). Rationale: graduated response prevents sudden behavior changes. Early tiers suggest efficiency; late tiers restrict new work.

---

## Settled

Past decisions that are locked in and should not be revisited.

- **2026-04-02** — Bonsai is a scaffolding tool, not a runtime orchestrator. It generates files and steps away. Rationale: see RESEARCH.md section 7 — "Where does Bonsai end?"
- **2026-04-02** — Catalog items are opinionated, not generic templates. Rationale: the opinionation is the value — users get working patterns out of the box, not blank templates they have to fill in.
- **2026-04-13** — Defer Managed Agents cloud integration until local foundation is stable. Rationale: the local CLI workflow needs to be solid before adding cloud deployment. See DESIGN-companion-app.md for Greenhouse design.
