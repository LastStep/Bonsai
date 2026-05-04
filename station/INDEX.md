---
tags: [meta, index]
description: Master lookup table — read this first every session.
---

# Bonsai — Project Index

## Project Snapshot

Bonsai is a CLI tool that scaffolds Claude Code agent workspaces — structured isolation, memory, protocols, workflows, and sensors for multi-agent projects. Single binary, `go install`.

**Current phase:** Dogfooding & Polish (see Playbook/Status.md)

### Tech Stack

| Layer | Technology |
|-------|-----------|
| Language | Go 1.24+ |
| CLI framework | Cobra |
| TUI forms | Huh (Charm) |
| TUI styling | LipGloss (Charm) |
| TUI components | BubbleTea (Charm) |
| Config format | YAML (`gopkg.in/yaml.v3`) |
| Template engine | Go `text/template` (stdlib) |
| Distribution | Single binary — Homebrew, GitHub Releases, `go install` / `embed.FS` |

### Key Metrics

| Metric | Value |
|--------|-------|
| Agent types | 6 (tech-lead, fullstack, backend, frontend, devops, security) |
| Catalog items | ~50 (skills, workflows, protocols, sensors, routines) |
| CLI commands | 8 (init, add, remove, list, catalog, update, guide, validate) |

---

## Document Registry

| Path | What it contains | When to use it |
|------|-----------------|----------------|
| `INDEX.md` | This file — project snapshot and document map | Read first, every session |
| `CLAUDE.md` | Navigation table — routes to agent instructions | First file loaded every session |
| `Playbook/Status.md` | Live task tracker (in-progress / pending / done) | Start of every session; after completing work |
| `Playbook/Roadmap.md` | Phases and milestones | When planning next steps |
| `Playbook/Backlog.md` | Prioritized intake queue — bugs, features, debt, ideas | When discovering out-of-scope issues; before promoting to Status.md |
| `Playbook/Plans/Active/` | Numbered implementation plans for agents | When handing off work to an agent |
| `Playbook/Standards/SecurityStandards.md` | Hard security rules — all agents | Every session, every plan, every code review |
| `Logs/FieldNotes.md` | User-maintained notes on work done outside sessions | Read every session |
| `Logs/KeyDecisionLog.md` | Settled architectural decisions | When planning or when a topic comes up |
| `Reports/Pending/` | Unprocessed agent completion reports | Check every session start |
| `Reports/report-template.md` | Structured report format for agents | When submitting a completion report |
| `code-index.md` | Code index — quick-nav to Go source functions | When navigating the codebase |

---

## Architecture Overview

```
User runs bonsai CLI
    |
    v
cmd/ (Cobra)          ← CLI commands: init, add, remove, list, catalog
    |
    v
internal/catalog/     ← loads embedded YAML metadata + templates from catalog/
internal/config/      ← reads/writes .bonsai.yaml + .bonsai-lock.yaml
internal/generate/    ← renders templates, writes lock-aware files
internal/tui/         ← Huh forms + LipGloss styled output
    |
    v
catalog/ (embed.FS)   ← bundled agents, skills, workflows, protocols, sensors, routines, scaffolding
    |
    v
Target project        ← generated workspace: CLAUDE.md, agent/, Playbook/, Logs/, etc.
```

---

## Agent Handoff Notes

> [!note]
> This project currently runs a single tech-lead agent (dogfooding). When code agents are added, they will operate in separate workspace directories with their own CLAUDE.md routing tables.
