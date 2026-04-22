<!-- Logo placeholder — artwork coming. Keep this comment so we remember to slot it in. -->

<div align="center">

# Bonsai

**A workspace for your coding agent.**

[![GitHub Release](https://img.shields.io/github/v/release/LastStep/Bonsai?style=flat&color=blue)](https://github.com/LastStep/Bonsai/releases)
[![License: MIT](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![CI](https://img.shields.io/github/actions/workflow/status/LastStep/Bonsai/ci.yml?label=CI)](https://github.com/LastStep/Bonsai/actions?query=workflow%3ACI)
[![Docs](https://img.shields.io/badge/docs-laststep.github.io%2FBonsai-blue)](https://laststep.github.io/Bonsai/)
[![Status: early-stage](https://img.shields.io/badge/status-early--stage-orange.svg)](https://github.com/LastStep/Bonsai/issues)

[Documentation](https://laststep.github.io/Bonsai/) · [Install](#install) · [Quick Start](#quick-start) · [Contributing](CONTRIBUTING.md)

</div>

<br>

---

## Who Bonsai is for

**Solo devs and small teams who want to give their coding agent real responsibility** — not just faster autocomplete.

Claude Code out of the box gets you a smart assistant. But the moment you want it to pick up where it left off last week, stay inside scope across sessions, or follow your team's standards without re-briefing every time, a single `CLAUDE.md` hits its ceiling fast.

Bonsai generates a structured workspace under `station/` and wires Claude Code hooks that enforce it.

**What that looks like in practice:**

- **Every session starts from the same context.** A `SessionStart` hook injects identity, memory, active plans, and health warnings before the agent's first reply. No re-briefing.
- **The project is navigable, not just searchable.** An indexed codebase, cross-linked plans, and Obsidian-compatible markdown mean the agent reads a map of the project — not a grep output.
- **Rules live in files, not prompts.** Protocols (`security.md`, `scope-boundaries.md`, `memory.md`) are version-controlled. Sensors fire on `PreToolUse` / `Stop` to block scope violations at the tool call, not the transcript.
- **Plans before it acts.** The agent writes a plan to `Playbook/Plans/Active/NN-*.md` before any dispatch. Reviews run from `agent/Skills/review-checklist.md` — not whatever the agent last remembered.
- **Everything is auditable.** Decisions go to `Logs/KeyDecisionLog.md`. Out-of-scope findings go to `Playbook/Backlog.md`. Dispatched agent reports go to `Reports/`. `git log` is your audit trail.

**Not just CLAUDE.md with extra steps.** `CLAUDE.md` is one markdown file, reloaded each session. Bonsai is a workspace: dozens of cross-linked files, version-controlled, enforced by hooks. The agent comes back to the same place every time — and so does the next contributor.

Here's the shape of what lands in your repo after `bonsai init`:

```
station/
├── CLAUDE.md                ← workspace navigation
├── INDEX.md                 ← project snapshot
├── Playbook/                ← Status · Roadmap · Backlog · Plans · Standards
├── Logs/                    ← decisions · field notes · routine log
├── Reports/                 ← pending · archive
└── agent/
    ├── Core/                ← identity · memory · self-awareness
    ├── Protocols/           ← security · scope · memory · session-start
    ├── Skills/              ← review checklist · planning template · domain standards
    ├── Workflows/           ← planning · code review · PR review · security audit
    ├── Sensors/             ← Claude Code hooks (auto-run on session / tool use / stop)
    └── Routines/            ← scheduled self-maintenance tasks
```

---

## Install

**Homebrew:**

```bash
brew install LastStep/tap/bonsai
```

**Binary download** — [GitHub Releases](https://github.com/LastStep/Bonsai/releases):

```bash
curl -sL https://github.com/LastStep/Bonsai/releases/latest/download/bonsai_Linux_amd64.tar.gz | tar xz
sudo mv bonsai /usr/local/bin/
```

**From source** (Go 1.24+):

```bash
go install github.com/LastStep/Bonsai/cmd/bonsai@latest
```

---

## Quick Start

```bash
cd your-project
bonsai init          # set up station + Tech Lead agent
bonsai add           # add a code agent (backend, frontend, etc.)
```

Open the project in Claude Code and say "hi, get started" — the agent self-orients: reads its identity, checks memory, scans active plans, and reports status.

> **[Your First Workspace](https://laststep.github.io/Bonsai/guides/your-first-workspace/)** — walkthrough with screenshots and explanations.

---

## See it in action

<div align="center">

<img src="assets/demos/init.gif" alt="bonsai init — cinematic flow from Vessel through Planted, ~28s end-to-end" width="900">

<sub><code>bonsai init</code> end-to-end — name your project, tend the soil, shape the branches, observe, plant.</sub>

</div>

---

## How it works

Bonsai treats agent instructions as a layered system. Each layer has clear semantics:

```
  Layer 6 │ Sensors       │ Automated enforcement via Claude Code hooks
  Layer 5 │ Routines      │ Periodic self-maintenance on a schedule
  Layer 4 │ Skills        │ Domain knowledge — standards, patterns, conventions
  Layer 3 │ Workflows     │ Step-by-step procedures — planning, review, audit
  Layer 2 │ Protocols     │ Hard rules — security, scope, memory, startup
  Layer 1 │ Core          │ Identity, memory, self-awareness
```

You pick the components at `bonsai init` / `bonsai add` time. Bonsai writes a complete, cross-linked workspace — navigation, hook wiring, file tracking — in one pass. Open it in any editor. Open it in Obsidian for a live graph.

<div align="center">

<img src="docs/assets/graph-view.png" alt="Bonsai workspace visualized in Obsidian graph view" width="700">

<sub>A Bonsai workspace in Obsidian — every node is a generated file, every edge is a live cross-reference.</sub>

</div>

### Six agent types

| Agent | Role |
|:------|:-----|
| **Tech Lead** | Architects, plans, reviews — never writes application code |
| **Backend** | API, database, server-side logic |
| **Frontend** | UI components, state management, styling |
| **Full-Stack** | End-to-end — UI, API, database, auth, tests |
| **DevOps** | Infrastructure-as-code, CI/CD, containers |
| **Security** | Vulnerability audits, auth review, dependency scanning |

The Tech Lead orchestrates — plans, dispatches work via worktree-isolated subagents, reviews output. Code agents implement. You talk to the Tech Lead.

### The catalog

Bonsai ships with **58 catalog items**, mix-and-match, filtered by agent compatibility:

- **17 skills** — coding standards, API design, auth patterns, testing, infrastructure
- **10 workflows** — planning, code review, security audit, PR review, session logging
- **4 protocols** — memory, security, scope boundaries, session startup (all required)
- **12 sensors** — scope guards, dispatch validation, context injection, code quality checks
- **8 routines** — backlog hygiene, dependency audit, doc freshness, vulnerability scan

> **[Browse the full catalog](https://laststep.github.io/Bonsai/catalog/overview/)** — descriptions, compatibility tables, defaults.

### Extensible

After generation, the files are yours. Add custom skills, workflows, or sensors — `bonsai update` detects them, tracks them in your config, and wires them into navigation. Lock-aware conflict resolution means your edits are never silently overwritten.

> **[Customizing Abilities](https://laststep.github.io/Bonsai/guides/customizing-abilities/)** · **[Creating Custom Sensors](https://laststep.github.io/Bonsai/guides/creating-custom-sensors/)** · **[Creating Custom Routines](https://laststep.github.io/Bonsai/guides/creating-custom-routines/)**

---

## Commands

| Command | What it does |
|:--------|:------------|
| `bonsai init` | Initialize project — station, scaffolding, Tech Lead |
| `bonsai add` | Add a code agent or abilities to an existing agent |
| `bonsai remove` | Remove an agent or individual ability |
| `bonsai list` | Show installed agents and components |
| `bonsai catalog` | Browse all available abilities |
| `bonsai update` | Detect custom files, sync workspace |
| `bonsai guide` | View bundled guides: quickstart, concepts, cli, custom-files |

> **[Command reference](https://laststep.github.io/Bonsai/commands/init/)** — flags, flows, examples.

---

<div align="center">

**[Documentation](https://laststep.github.io/Bonsai/)** · **[Catalog](https://laststep.github.io/Bonsai/catalog/overview/)** · **[Contributing](CONTRIBUTING.md)** · **[Releases](https://github.com/LastStep/Bonsai/releases)** · **[MIT License](LICENSE)**

Built with [Cobra](https://github.com/spf13/cobra), [Huh](https://github.com/charmbracelet/huh), [LipGloss](https://github.com/charmbracelet/lipgloss), and [BubbleTea](https://github.com/charmbracelet/bubbletea). Developed with [Claude Code](https://claude.ai/code).

</div>
