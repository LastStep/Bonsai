# Bonsai

**Scaffold Claude Code agent workspaces from the command line.**

Pick your components and Bonsai generates the full instruction layer — identity, protocols, workflows, skills, sensors, routines, and project scaffolding — all wired up and ready to go.

One binary. No runtime dependencies. Works with any project.

> [!TIP]
> **First time?** Read the **[Handbook](HANDBOOK.md)** for a deep dive into how the agent system works and how to get the best results.

---

## Install

```bash
go install github.com/LastStep/Bonsai@latest
```

> Requires Go 1.24+

---

## Quick Start

**1. Initialize** — set up the station (project scaffolding + Tech Lead agent):

```bash
cd your-project
bonsai init
```

This creates a `station/` directory with project management infrastructure and installs the Tech Lead agent — your project's primary agent that architects the system and dispatches work to other agents.

**2. Add code agents** — pick a type, select components, review, and generate:

```bash
bonsai add
```

**3. Repeat** for each code agent you need. That's it.

<details>
<summary><strong>Other commands</strong></summary>

```bash
bonsai list                    # see what's installed
bonsai catalog                 # browse all available items
bonsai catalog --agent backend # filter by agent type
bonsai remove backend          # remove agent from config
bonsai remove backend -d       # also delete generated files
```

</details>

---

## Agent Types

| Agent | Role | Installed via |
|:------|:-----|:-------------|
| **Tech Lead** | Architects the system, writes plans, reviews agent output — _never writes application code_ | `bonsai init` (required) |
| **Backend** | Executes backend plans — API, database, server-side logic | `bonsai add` |
| **Frontend** | Executes frontend plans — UI components, state management, styling | `bonsai add` |
| **Full-Stack** | Implements features end-to-end — UI, API routes, database, auth, tests | `bonsai add` |
| **DevOps** | Manages infrastructure-as-code, CI/CD pipelines, containers, deployment | `bonsai add` |
| **Security** | Audits code for vulnerabilities, reviews auth patterns, scans dependencies | `bonsai add` |

> Tech Lead is always installed first. You talk to the Tech Lead; it dispatches work to code agents. You can still work directly with code agents for quick fixes.

---

## Catalog

Every component is mix-and-match. Bonsai filters by agent compatibility automatically.

| Category | What it is | Examples |
|:---------|:----------|:---------|
| **Skills** | Domain knowledge and standards | `coding-standards` `testing` `database-conventions` `api-design-standards` `auth-patterns` `design-guide` |
| **Workflows** | Step-by-step task procedures | `planning` `plan-execution` `code-review` `reporting` `security-audit` `session-logging` |
| **Protocols** | Hard rules, enforced every session | `session-start` `security` `scope-boundaries` `memory` |
| **Sensors** | Auto-enforced hooks on Claude Code events | `session-context` `scope-guard-files` `dispatch-guard` `api-security-check` `test-integrity-guard` |
| **Routines** | Periodic self-maintenance tasks (Tech Lead only) | `dependency-audit` `vulnerability-scan` `doc-freshness-check` `status-hygiene` |

> Run `bonsai catalog` to see the full list with descriptions, agent compatibility, and frequencies.

---

## What Gets Generated

After `bonsai init` + `bonsai add` (backend agent):

```
your-project/
├── .bonsai.yaml                    # project config
├── .claude/settings.json           # auto-wired sensor hooks
├── CLAUDE.md                       # root navigation — routes to agent workspaces
│
├── station/                        # Tech Lead workspace + project scaffolding
│   ├── CLAUDE.md                   #   Tech Lead agent navigation
│   ├── INDEX.md                    #   project snapshot + document registry
│   ├── agent/                      #   Tech Lead agent instructions
│   │   ├── Core/                   #     identity.md, memory.md, self-awareness.md, routines.md
│   │   ├── Protocols/              #     hard-enforced rules
│   │   ├── Workflows/              #     task procedures (planning, code-review, etc.)
│   │   ├── Skills/                 #     domain knowledge
│   │   ├── Sensors/                #     rendered hook scripts (.sh)
│   │   └── Routines/              #     maintenance procedures
│   ├── Playbook/
│   │   ├── Status.md               #   live task tracker
│   │   ├── Roadmap.md              #   long-term milestones
│   │   ├── Plans/Active/           #   implementation plans
│   │   └── Standards/
│   │       └── SecurityStandards.md
│   ├── Logs/
│   │   ├── FieldNotes.md           #   notes from outside sessions
│   │   ├── KeyDecisionLog.md       #   settled architectural decisions
│   │   └── RoutineLog.md           #   routine execution history
│   └── Reports/
│       ├── report-template.md
│       └── Pending/                #   unreviewed completion reports
│
└── backend/                        # code agent workspace
    ├── CLAUDE.md                   #   agent-specific navigation
    └── agent/
        ├── Core/                   #   identity.md, memory.md, self-awareness.md
        ├── Protocols/              #   hard rules
        ├── Workflows/              #   task procedures
        ├── Skills/                 #   domain knowledge
        └── Sensors/                #   rendered hook scripts (.sh)
```

---

## CLI Reference

| Command | Description | Flags |
|:--------|:-----------|:------|
| `bonsai init` | Initialize project — station, scaffolding, and Tech Lead agent | |
| `bonsai add` | Add a code agent (interactive) | |
| `bonsai remove <agent>` | Remove an installed agent | `-d` `--delete-files` |
| `bonsai list` | Show installed agents + components | |
| `bonsai catalog` | Browse available catalog items | `-a` `--agent <type>` |

---

**[Handbook](HANDBOOK.md)** — mental model, interaction patterns, sensor/routine deep dives, setup recommendations, best practices

**[License](LICENSE)** — MIT
