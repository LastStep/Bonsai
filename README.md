# Bonsai

**CLI tool for scaffolding Claude Code agent workspaces.**

Bonsai generates the file structure, instructions, and automation that Claude Code agents need to work effectively in your project. Pick an agent type, choose its skills, workflows, protocols, sensors, and routines — Bonsai wires everything up into a ready-to-use workspace.

One binary. No runtime dependencies. Works with any project.

> **New to Bonsai?** Read the [Handbook](HANDBOOK.md) for a walkthrough of how the agent system works and how to get the best results from it.

## Install

```bash
go install github.com/LastStep/Bonsai@latest
```

Requires Go 1.24+.

## Quick Start

### 1. Initialize your project

```bash
cd your-project
bonsai init
```

You'll be prompted for your project name, description, and docs directory. Then you'll select which scaffolding to include (status tracking, plans, logs, reports). Some scaffolding is required — Bonsai will tell you which.

### 2. Add an agent

```bash
bonsai add
```

Interactive walkthrough:
1. **Pick an agent type** — tech-lead, backend, frontend, fullstack, devops, or security
2. **Set the workspace directory** — where the agent's files live (e.g. `backend/`)
3. **Select components** — skills, workflows, protocols, sensors, and routines. Each agent type comes with smart defaults; you just confirm or customize.
4. **Review and confirm** — see a summary tree before anything is generated

Repeat `bonsai add` for each agent you want.

### 3. See what's installed

```bash
bonsai list
```

Shows all installed agents with their workspace paths and selected components.

### 4. Browse the catalog

```bash
# Full catalog
bonsai catalog

# Filter by agent type
bonsai catalog --agent backend
```

Shows every available agent, skill, workflow, protocol, sensor, and routine in formatted tables.

### 5. Remove an agent

```bash
# Remove from config (keeps generated files)
bonsai remove backend

# Remove config and delete generated files
bonsai remove backend --delete-files
```

## Agent Types

| Agent | Role |
|-------|------|
| **Tech Lead** | Architects the system, writes plans, reviews agent output — never writes application code |
| **Backend** | Executes backend plans — API, database, server-side logic |
| **Frontend** | Executes frontend plans — UI components, state management, styling |
| **Full-Stack** | Implements full-stack features end-to-end — UI, API routes, database, auth, tests |
| **DevOps** | Manages infrastructure-as-code, CI/CD pipelines, containers, and deployment automation |
| **Security** | Audits code for vulnerabilities, reviews auth patterns, scans dependencies |

## Catalog

Every component is mix-and-match. Each has an agent compatibility list — Bonsai only shows you what's relevant.

| Category | What it is | Examples |
|----------|-----------|----------|
| **Skills** | Domain knowledge and coding standards | coding-standards, testing, database-conventions, api-design-standards, auth-patterns, design-guide |
| **Workflows** | Step-by-step procedures for specific tasks | planning, plan-execution, code-review, reporting, security-audit, session-logging |
| **Protocols** | Hard rules enforced every session | session-start, security, scope-boundaries, memory |
| **Sensors** | Automated hooks that run on Claude Code events | session-context, scope-guard-files, dispatch-guard, api-security-check, test-integrity-guard |
| **Routines** | Periodic self-maintenance tasks | dependency-audit, vulnerability-scan, doc-freshness-check, status-hygiene, memory-consolidation |

Run `bonsai catalog` to see the full list with descriptions.

## What Gets Generated

After `bonsai init` + `bonsai add` for a backend agent with docs at `docs/`:

```
your-project/
├── .bonsai.yaml                  # Project config — tracks all installed agents
├── .claude/
│   └── settings.json             # Auto-wired sensor hooks
├── CLAUDE.md                     # Root navigation — routes agents to their workspace
├── docs/
│   ├── INDEX.md                  # Project snapshot and document registry
│   ├── Playbook/
│   │   ├── Status.md             # Live task tracker
│   │   ├── Roadmap.md            # Long-term milestones
│   │   ├── Plans/Active/         # Implementation plans go here
│   │   └── Standards/
│   │       └── SecurityStandards.md
│   ├── Logs/
│   │   ├── FieldNotes.md         # Notes from work outside agent sessions
│   │   ├── KeyDecisionLog.md     # Settled architectural decisions
│   │   └── RoutineLog.md         # Routine execution history
│   └── Reports/
│       ├── report-template.md
│       └── Pending/              # Unreviewed agent completion reports
└── backend/
    ├── CLAUDE.md                 # Agent-specific navigation
    └── agent/
        ├── Core/
        │   ├── identity.md       # Role, mindset, relationships
        │   ├── memory.md         # Working memory across sessions
        │   └── self-awareness.md # Context window monitoring
        ├── Skills/               # Domain knowledge files
        ├── Workflows/            # Task procedure files
        ├── Protocols/            # Hard-enforced rule files
        └── Sensors/              # Rendered hook scripts (.sh)
```

Agents with routines also get `agent/Routines/` and a managed dashboard at `agent/Core/routines.md`.

## CLI Reference

| Command | Description | Flags |
|---------|-------------|-------|
| `bonsai init` | Initialize Bonsai in the current project | — |
| `bonsai add` | Add an agent (interactive) | — |
| `bonsai remove <agent>` | Remove an installed agent | `--delete-files`, `-d` |
| `bonsai list` | Show installed agents and components | — |
| `bonsai catalog` | Browse available catalog items | `--agent <type>`, `-a` |

## Learn More

- **[Handbook](HANDBOOK.md)** — How the agent system works, interaction patterns, and tips for getting the best results
- **[License](LICENSE)** — MIT
