---
tags: [playbook, roadmap]
description: High-level project phases and goals. This is the long-term plan, not a task tracker.
---

# Bonsai — Roadmap

> [!note]
> This is the long-term project plan — think milestones, not daily tasks.
> For current active work, see Status.md.

---

## Current Phase

### Phase 1 — Foundation & Polish

**Goal:** Bonsai is a reliable, polished CLI that produces high-quality agent workspaces. Dogfooding validates the full loop.

- [x] Go rewrite — Cobra, Huh, LipGloss, BubbleTea
- [x] Full catalog — 6 agent types, skills, workflows, protocols, sensors, routines
- [x] Lock file conflict handling — detect user-modified files, prevent silent overwrites
- [x] Awareness Framework — status-bar + context-guard sensors
- [x] Dogfooding — Bonsai manages itself via station/ workspace
- [ ] Better trigger sections — clearer activation conditions for catalog items
- [ ] UI overhaul — polish TUI for public-facing quality
- [ ] Usage instructions — best practices, first-run guidance, catalog item explanations
- [x] Release pipeline — GoReleaser + GitHub Actions + Homebrew Tap for cross-platform binary distribution
- [x] Community health files — ISSUE_TEMPLATE, CONTRIBUTING.md, CODE_OF_CONDUCT.md

---

## Future Phases

### Phase 2 — Extensibility

**Goal:** Users can create custom catalog items, extend existing ones, and share them.

- [x] Custom item detection — recognize user-created skills/workflows/protocols outside the catalog
- [ ] Self-update mechanism — catalog items can flag when they're stale or have issues
- [ ] Template variables expansion — richer context available in templates
- [ ] Micro-task fast path — lightweight protocol bypass for trivial changes

### Phase 3 — Cloud & Orchestration

**Goal:** Bonsai workspaces can deploy to Claude's Managed Agents platform for autonomous execution.

- [ ] Managed Agents integration — `bonsai deploy`, session management, outcome rubrics
- [ ] Greenhouse companion app — desktop app for managing projects + observing AI agents (Tauri v2 + Svelte 5 + SQLite)

### Phase 4 — Ecosystem

**Goal:** Bonsai is a platform with community-contributed agents, skills, and workflows.

- [ ] Catalog marketplace — discover, install, and share catalog items
- [ ] Plugin system — third-party extensions
- [ ] Cross-project coordination — agents working across multiple repositories
