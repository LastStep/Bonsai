<div align="center">

# Bonsai

**A structured language for working with AI agents.**

[![GitHub Release](https://img.shields.io/github/v/release/LastStep/Bonsai?style=flat&color=blue)](https://github.com/LastStep/Bonsai/releases)
[![License: MIT](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![CI](https://img.shields.io/github/actions/workflow/status/LastStep/Bonsai/ci.yml?label=CI)](https://github.com/LastStep/Bonsai/actions?query=workflow%3ACI)
[![Go Reference](https://pkg.go.dev/badge/github.com/LastStep/Bonsai.svg)](https://pkg.go.dev/github.com/LastStep/Bonsai)
[![Go Report Card](https://goreportcard.com/badge/github.com/LastStep/Bonsai)](https://goreportcard.com/report/github.com/LastStep/Bonsai)

Give your Claude Code agents identity, memory, protocols, and purpose —<br>
so they work like teammates, not tools.

[Install](#install) · [Quick Start](#quick-start) · [Handbook](HANDBOOK.md) · [Custom Files](docs/custom-files.md)

<br>

<img src="docs/assets/graph-view.png" alt="Bonsai workspace visualized in Obsidian graph view" width="700">

<sub>A Bonsai workspace visualized in Obsidian — every node is a generated file, every edge is a live cross-reference.</sub>

<!-- GIF showcase coming soon: bonsai init, bonsai add, agent dispatching -->

</div>

<br>

---

## The Problem

Claude Code is powerful out of the box. But the moment you need agents to **coordinate**, **stay consistent across sessions**, or **follow your team's standards** — you're writing walls of markdown by hand. Every project rediscovers the same patterns: identity files, memory protocols, security rules, scope boundaries, review checklists.

And if you want multiple agents working in the same codebase? You're maintaining parallel instruction sets that drift apart.

## What Bonsai Does

Bonsai treats agent instructions as a **structured language** — not a pile of markdown files, but a layered system with clear semantics:

```
  Layer 6 │ Sensors       │ Automated enforcement via Claude Code hooks
  Layer 5 │ Routines      │ Periodic self-maintenance on a schedule
  Layer 4 │ Skills        │ Domain knowledge — standards, patterns, conventions
  Layer 3 │ Workflows     │ Step-by-step procedures — planning, review, audit
  Layer 2 │ Protocols     │ Hard rules — security, scope, memory, startup
  Layer 1 │ Core          │ Identity, memory, self-awareness
```

You pick the components. Bonsai generates a complete, wired-up workspace — with cross-linked navigation, auto-enforced hooks, and a shared project scaffold that keeps every agent on the same page.

One binary. No runtime. Works with any project.

## Who It's For

- **Solo developers** who want their Claude Code agent to behave consistently, remember context between sessions, and follow project-specific standards
- **Teams** coordinating multiple agents across a codebase — backend, frontend, devops — with a shared source of truth
- **Anyone** who's tired of copy-pasting agent instructions between projects

---

## Install

**Homebrew** (macOS / Linux):

```bash
brew install LastStep/tap/bonsai
```

**Binary download** — grab the latest from [GitHub Releases](https://github.com/LastStep/Bonsai/releases):

```bash
# Example for Linux amd64
curl -sL https://github.com/LastStep/Bonsai/releases/latest/download/bonsai_Linux_amd64.tar.gz | tar xz
sudo mv bonsai /usr/local/bin/
```

**From source** (requires Go 1.24+):

```bash
go install github.com/LastStep/Bonsai@latest
```

---

## Quick Start

```bash
cd your-project
bonsai init          # set up station + Tech Lead agent
bonsai add           # add a code agent (backend, frontend, etc.)
```

That's it. Your project now has a full agent workspace. Open it in Claude Code and start working.

<details>
<summary><strong>What happens during <code>bonsai init</code></strong></summary>

<br>

You'll be prompted for:
- Project name and description
- Station directory (default `station/`)
- Agent type (Tech Lead is always first)
- Components to install — skills, workflows, sensors, routines

Bonsai generates everything: identity files, protocol docs, hook scripts, navigation tables, project scaffolding — all cross-linked with relative markdown paths that work in GitHub, VS Code, and Obsidian.

</details>

<details>
<summary><strong>All commands</strong></summary>

<br>

| Command | What it does |
|:--------|:------------|
| `bonsai init` | Initialize project — station, scaffolding, Tech Lead agent |
| `bonsai add` | Add a code agent (interactive picker) |
| `bonsai remove <agent>` | Remove an agent (`-d` to delete files) |
| `bonsai list` | Show installed agents and components |
| `bonsai catalog` | Browse all available abilities |
| `bonsai update` | Detect custom files, re-render abilities, sync workspace |
| `bonsai guide` | Render the custom files guide in the terminal |

</details>

---

## How It Works

### The Station

Every project gets a **station** — a shared command center with project scaffolding (status tracker, roadmap, plans, decision log) and the Tech Lead agent. Code agents get their own workspaces but reference the station for coordination.

### Agents as Team Members

| What a new hire needs | Bonsai equivalent |
|:---------------------|:-----------------|
| Role description | **Identity** — job title, mindset, relationships |
| Company handbook | **Protocols** — security policy, scope boundaries |
| Playbooks | **Workflows** — how to plan, review, report |
| Domain training | **Skills** — coding standards, API patterns |
| Automated guardrails | **Sensors** — catch mistakes in real time |
| Maintenance duties | **Routines** — periodic audits on a schedule |

### Agent Types

| Agent | Role | How to install |
|:------|:-----|:--------------|
| **Tech Lead** | Architects, plans, reviews — never writes application code | `bonsai init` |
| **Backend** | API, database, server-side logic | `bonsai add` |
| **Frontend** | UI components, state management, styling | `bonsai add` |
| **Full-Stack** | End-to-end — UI, API, database, auth, tests | `bonsai add` |
| **DevOps** | Infrastructure-as-code, CI/CD, containers | `bonsai add` |
| **Security** | Vulnerability audits, auth review, dependency scanning | `bonsai add` |

> You talk to the Tech Lead. It writes plans and dispatches work to code agents. You can also work directly with code agents for quick fixes.

---

## The Catalog

Everything is mix-and-match. Bonsai filters by agent compatibility automatically.

<details>
<summary><strong>Skills</strong> — domain knowledge and standards (16 items)</summary>

<br>

| Skill | Description | Agents |
|:------|:-----------|:-------|
| API Design Standards | REST conventions — URLs, methods, errors, pagination | backend, fullstack, tech-lead |
| Auth Patterns | Authentication flows, tokens, sessions, access control | backend, fullstack, security |
| CLI Conventions | CLI structure, flags, output, exit codes | backend, fullstack |
| Coding Standards | Language-specific formatting, linting, patterns | backend, devops, frontend, fullstack, security |
| Container Standards | Docker builds, image security, Kubernetes manifests | devops |
| Database Conventions | SQL naming, migrations, schema design | backend, fullstack, tech-lead |
| Design Guide | UI/UX patterns, components, accessibility | frontend, fullstack |
| Dispatch | How to triage and dispatch work to code agents | tech-lead |
| IaC Conventions | Terraform naming, state management, modules | devops |
| Issue Classification | Issue types, importance levels, domain labels | tech-lead |
| Mobile Patterns | State management, offline-first, navigation | fullstack |
| Planning Template | Plan format, tier rules, implementation templates | tech-lead |
| PR Creation | Branch naming, title conventions, body template | tech-lead, fullstack, backend, frontend |
| Review Checklist | Code review — correctness, security, performance | reviewer, security, tech-lead |
| Test Strategy | Test pyramid, coverage, anti-patterns | backend, frontend, fullstack |
| Testing | Frameworks, patterns, coverage requirements | backend, frontend, fullstack, security |

</details>

<details>
<summary><strong>Workflows</strong> — step-by-step task procedures (10 items)</summary>

<br>

| Workflow | Description | Agents |
|:---------|:-----------|:-------|
| API Development | Spec-first — define contract, scaffold, implement, test | backend, fullstack |
| Code Review | Review agent output against the plan | security, tech-lead |
| Issue to Implementation | End-to-end autonomous — intake to shipped code | tech-lead |
| Plan Execution | Execute an assigned plan step by step | backend, devops, frontend, fullstack |
| Planning | From request to dispatch-ready plan | tech-lead |
| PR Review | Review a pull request — scope, correctness, security | tech-lead |
| Reporting | Structured completion reports | backend, devops, frontend, fullstack, security |
| Security Audit | Secrets scan, SAST, dependency audit, config review | devops, security, tech-lead |
| Session Logging | End-of-session log — decisions, open items | all |
| Test Plan | Design a structured test plan for a feature | tech-lead |

</details>

<details>
<summary><strong>Sensors</strong> — automated enforcement via hooks (12 items)</summary>

<br>

| Sensor | Event | What it does |
|:-------|:------|:------------|
| Agent Review | PostToolUse (Agent) | Review checklist after agent completes |
| API Security Check | PreToolUse (Edit\|Write) | Detects SQL injection, hardcoded secrets, eval |
| Context Guard | UserPromptSubmit | Injects behavioral constraints per prompt |
| Dispatch Guard | PreToolUse (Agent) | Validates worktree isolation + plan before dispatch |
| IaC Safety Guard | PreToolUse (Bash) | Blocks terraform destroy, kubectl delete, etc. |
| Routine Check | SessionStart | Flags overdue maintenance routines |
| Scope Guard Commands | PreToolUse (Bash) | Blocks app execution commands (tests, builds) |
| Scope Guard Files | PreToolUse (Edit\|Write) | Blocks edits outside agent's workspace |
| Session Context | SessionStart | Injects identity, memory, protocols at startup |
| Status Bar | Stop | Shows context usage, health, git state |
| Subagent Stop Review | SubagentStop | Review checklist when dispatched agent finishes |
| Test Integrity Guard | PreToolUse (Edit\|Write) | Catches removed assertions, .skip, empty tests |

</details>

<details>
<summary><strong>Routines</strong> — periodic self-maintenance (8 items, Tech Lead only)</summary>

<br>

| Routine | Frequency | What it does |
|:--------|:----------|:------------|
| Backlog Hygiene | 7 days | Flag stale items, escalate P0s |
| Dependency Audit | 7 days | Scan for vulnerabilities and unmaintained packages |
| Doc Freshness Check | 7 days | Detect docs that drifted from code |
| Infra Drift Check | 7 days | Compare IaC state against actual resources |
| Memory Consolidation | 5 days | Validate and clean agent memory |
| Roadmap Accuracy | 14 days | Ensure roadmap reflects reality |
| Status Hygiene | 5 days | Archive done items, validate pending |
| Vulnerability Scan | 7 days | SAST, secrets scan, dependency audit |

</details>

<details>
<summary><strong>Protocols</strong> — hard rules enforced every session (4 items, all required)</summary>

<br>

| Protocol | What it enforces |
|:---------|:----------------|
| Memory | How to read/write working memory between sessions |
| Scope Boundaries | What you own, what you never touch |
| Security | Hard stops — secrets, credentials, dangerous operations |
| Session Start | Ordered startup sequence — what to read and check |

</details>

> Run `bonsai catalog` to see the full list with descriptions and compatibility info.

---

## What Gets Generated

After `bonsai init` + `bonsai add` (backend agent):

```
your-project/
├── .bonsai.yaml                    # project config
├── .bonsai-lock.yaml               # file tracking (hashes for conflict detection)
├── .claude/settings.json           # auto-wired sensor hooks
├── CLAUDE.md                       # root navigation
│
├── station/                        # Tech Lead workspace + project scaffolding
│   ├── CLAUDE.md                   #   navigation (cross-linked)
│   ├── INDEX.md                    #   project snapshot
│   ├── agent/
│   │   ├── Core/                   #     identity, memory, self-awareness
│   │   ├── Protocols/              #     security, scope, startup sequence
│   │   ├── Workflows/              #     planning, code-review, audit, etc.
│   │   ├── Skills/                 #     standards, patterns, checklists
│   │   ├── Sensors/                #     hook scripts (.sh)
│   │   └── Routines/              #     maintenance procedures
│   ├── Playbook/                   #   status, roadmap, plans, standards
│   ├── Logs/                       #   field notes, decisions, routine log
│   └── Reports/                    #   completion reports
│
└── backend/                        # code agent workspace
    ├── CLAUDE.md                   #   agent-specific navigation
    └── agent/
        ├── Core/                   #   identity, memory, self-awareness
        ├── Protocols/              #   hard rules
        ├── Workflows/              #   task procedures
        ├── Skills/                 #   domain knowledge
        └── Sensors/                #   hook scripts
```

Every file reference is a clickable markdown link — open the workspace in Obsidian to see the full knowledge graph, or browse it on GitHub.

---

## Extending Bonsai

Bonsai workspaces are designed to be extended. After generation, you own the files.

**Custom abilities** — add your own skills, workflows, protocols, or sensors directly to `agent/`. Run `bonsai update` and Bonsai detects them, tracks them in your config, and includes them in navigation tables.

**Lock-aware updates** — Bonsai tracks file hashes. When you re-run generation, unmodified files update silently. Files you've edited trigger a conflict prompt — skip, overwrite, or backup.

**Templates** — catalog files use Go templates (`{{ .ProjectName }}`, `{{ .AgentDisplayName }}`). Create `.tmpl` files and they're rendered at generation time.

> Run `bonsai guide` for the full custom files reference.

---

## Guides

| Guide | What you'll learn |
|:------|:-----------------|
| **[Handbook](HANDBOOK.md)** | Mental model, interaction patterns, sensor/routine deep dives, best practices |
| **[Working With Agents](docs/working-with-agents.md)** | Communication patterns, framing techniques, and collaboration rhythms that get the best results |
| **[Custom Files](docs/custom-files.md)** | How to create your own abilities — meta.yaml format, frontmatter, templates |

---

## Contributing

Bonsai is early-stage and evolving fast. If you have ideas, find bugs, or want to add catalog items:

1. Open an [issue](https://github.com/LastStep/Bonsai/issues) to discuss
2. Fork, branch, and submit a PR
3. Keep PRs focused — one feature or fix per PR

---

<div align="center">

**[GitHub](https://github.com/LastStep/Bonsai)** · **[Releases](https://github.com/LastStep/Bonsai/releases)** · **[MIT License](LICENSE)**

Built with [Cobra](https://github.com/spf13/cobra), [Huh](https://github.com/charmbracelet/huh), [LipGloss](https://github.com/charmbracelet/lipgloss), and [BubbleTea](https://github.com/charmbracelet/bubbletea).

</div>
