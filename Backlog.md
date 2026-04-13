---
tags: [playbook, backlog]
description: Prioritized backlog — bugs, features, debt, research, and improvement ideas for Bonsai.
---

# Bonsai — Backlog

> [!note]
> This is the intake queue for all work not yet active. Items flow from here into active work.
> For current work state, see `agent/Core/memory.md` → Work State.

---

## How This Works

**Capture:** When you discover a bug, improvement opportunity, tech debt, or idea during a session that is outside your current task scope — add it here instead of fixing it inline. Use the item format below.

**Promote:** When starting work on an item, move it to Work State in `agent/Core/memory.md` and remove it from here.

**Resolve:** When an item is completed, remove it. Done items are tracked in git history.

### Item Format

```
- **[category] Short description** — Context or rationale. *(added YYYY-MM-DD, source: routine|session|user)*
```

**Categories:** `bug`, `feature`, `debt`, `security`, `research`, `improvement`

### Priority Guide

| Priority | Meaning | Action |
|----------|---------|--------|
| **P0** | Blocking current work or broken functionality | Work on it now |
| **P1** | Next up when current work completes | Promote when capacity opens |
| **P2** | Planned but not urgent | Review at phase boundaries |
| **P3** | Ideas, nice-to-haves, research topics | Review periodically |

---

## P0 — Critical

## P1 — High

- **[improvement] Better trigger sections** — Each workflow/protocol/skill needs clearer trigger conditions with examples so they actually fire reliably in agent sessions. *(added 2026-04-13, source: user)*

- **[feature] Session-done trigger** — End-of-session checklist triggered by human input (e.g. "session done"). Two layers: (1) a trigger word in CLAUDE.md that runs a wrap-up procedure (commit check, memory update, backlog review, etc.), (2) a `Stop` event sensor as a safety net for critical checks (uncommitted changes, stale memory). The trigger word approach is a workflow/routine; the `Stop` hook is a sensor. Could ship both. *(added 2026-04-13, source: user)*

## P2 — Medium

- **[improvement] UI overhaul** — Current TUI is rough. Needs polish pass. *(added 2026-04-13, source: user)*
- **[feature] Usage instructions** — Add guidance on best practices: running routines at first start to fix references, how to trigger specific effects, proper explanation of each catalog item. *(added 2026-04-13, source: user)*
- **[feature] Custom item detection** — Detect user-created skills/workflows/protocols that aren't from the catalog. *(added 2026-04-13, source: user)*
- **[improvement] Self-update mechanism** — Skills and workflows should be able to self-heal or flag when they have issues. *(added 2026-04-13, source: user)*
- **[feature] Context-aware self-awareness** — Improve self-awareness.md concept with real context consumption warnings. Warn user at 40%, 60%, 80% context usage before starting large tasks. Needs a hook/sensor approach — possibly a PostToolUse or periodic sensor that tracks conversation size and injects warnings. Research how Claude Code exposes context usage to hooks. Agent should refuse ambitious work at 80%+ and suggest a new session instead. *(added 2026-04-13, source: user)*

## P3 — Ideas & Research

- **[research] Archon analysis** — https://github.com/coleam00/Archon — research what it does, use cases, overlap with Bonsai, what we can learn. *(added 2026-04-13, source: user)*
- **[feature] Managed Agents integration** — Cloud deployment via `bonsai deploy`, session management, outcome rubrics in catalog. Build after local foundation is stable. *(added 2026-04-13, source: user)*
- **[feature] Greenhouse companion app** — Desktop app for managing projects + observing AI agents. Design doc: DESIGN-companion-app.md. Stack: Tauri v2 + Svelte 5 + SQLite. Status: Design phase, decisions locked. *(added 2026-04-13, source: user)*
