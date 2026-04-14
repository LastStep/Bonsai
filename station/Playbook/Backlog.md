---
tags: [playbook, backlog]
description: Prioritized backlog — bugs, features, debt, research, and improvement ideas. Self-maintained by agents via the backlog-hygiene routine.
---

# Bonsai — Backlog

> [!note]
> This is the intake queue for all work not yet in `Status.md`. Items flow from here into active work.
> For current active work, see `Playbook/Status.md`. For long-term direction, see `Playbook/Roadmap.md`.

---

## How This Works

**Capture:** When you discover a bug, improvement opportunity, tech debt, or idea during a session that is outside your current task scope — add it here instead of fixing it inline. Use the item format below.

**Promote:** When capacity opens, move P0/P1 items into `Playbook/Status.md` as Pending or In Progress. Remove the item from this file when it appears in Status.

**Resolve:** Items completed via Status.md are cleaned up by the backlog-hygiene routine. Items abandoned or made irrelevant should be removed with a note in `Logs/RoutineLog.md`.

**Review:** The backlog-hygiene routine runs periodically to flag stale items, escalate misplaced P0s, remove duplicates, and cross-reference with Status.md and Roadmap.md.

### Item Format

```
- **[category] Short description** — Context or rationale. *(added YYYY-MM-DD, source: routine|session|user)*
```

**Categories:** `bug`, `feature`, `debt`, `security`, `research`, `improvement`

### Priority Guide

| Priority | Meaning | Action |
|----------|---------|--------|
| **P0** | Blocking current work or broken functionality | Must be in Status.md. If a P0 is here, escalate it immediately |
| **P1** | Next up when current work completes | Promote to Status.md when capacity opens |
| **P2** | Planned but not urgent | Review at phase boundaries |
| **P3** | Ideas, nice-to-haves, research topics | Review during roadmap updates |

---

## P0 — Critical

## P1 — High

- **[improvement] Better trigger sections** — Each workflow/protocol/skill needs clearer trigger conditions with examples so they actually fire reliably in agent sessions. *(added 2026-04-13, source: user)*

## P2 — Medium

- **[improvement] UI overhaul** — Current TUI is rough. Needs polish pass. *(added 2026-04-13, source: user)*
- **[feature] Usage instructions** — Add guidance on best practices: running routines at first start to fix references, how to trigger specific effects, proper explanation of each catalog item. *(added 2026-04-13, source: user)*
- **[feature] ~~Custom item detection~~** — Done. Shipped as `bonsai update`. *(added 2026-04-13, completed 2026-04-14)*
- **[improvement] Self-update mechanism** — Skills and workflows should be able to self-heal or flag when they have issues. *(added 2026-04-13, source: user)*
- **[feature] `bonsai guide` command** — Surface the custom files guide (`docs/custom-files.md`) as a terminal command. Keep doc and command output in sync. *(added 2026-04-14, source: user)*
- **[feature] Custom item creator** — Interactive TUI for creating custom items (skill, workflow, protocol, sensor, routine) with frontmatter scaffolding — similar to Claude's skill creator. *(added 2026-04-14, source: user)*
- **[improvement] Catalog display_name audit** — Add explicit `display_name` to all catalog `meta.yaml` files. Research other metadata fields that could be useful (e.g., `version`, `tags`, `dependencies`, `examples`). *(added 2026-04-14, source: user)*

## P3 — Ideas & Research

- **[research] Archon analysis** — https://github.com/coleam00/Archon — research what it does, use cases, overlap with Bonsai, what we can learn. *(added 2026-04-13, source: user)*
- **[feature] Managed Agents integration** — Cloud deployment via `bonsai deploy`, session management, outcome rubrics in catalog. Build after local foundation is stable. *(added 2026-04-13, source: user)*
- **[feature] Greenhouse companion app** — Desktop app for managing projects + observing AI agents. Design doc: DESIGN-companion-app.md. Stack: Tauri v2 + Svelte 5 + SQLite. Status: Design phase, decisions locked. *(added 2026-04-13, source: user)*
