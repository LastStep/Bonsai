# Bonsai

**CLI tool for scaffolding Claude Code agent workspaces.**

Bonsai sets up the file structure, instructions, and automation that Claude Code agents need to work effectively in your project. Pick an agent type, select its skills, workflows, protocols, and sensors — Bonsai generates the workspace with everything wired up.

## What it does

- Generates agent workspaces with identity, memory, and self-awareness templates
- Installs skills (coding standards, testing), workflows (planning, code review), and protocols (security, scope boundaries)
- Sets up **sensors** — auto-enforced hooks that inject context at session start and block out-of-scope actions
- Creates project management scaffolding (status tracking, plans, logs, reports)
- Wires everything into `CLAUDE.md` navigation files and `.claude/settings.json`

## Install

```bash
go install github.com/LastStep/Bonsai@latest
```

Requires Go 1.24+.

## Usage

```bash
# Initialize in your project
bonsai init

# Add an agent — interactive selection of type, skills, workflows, protocols, sensors
bonsai add

# See what's installed
bonsai list

# Browse the full catalog
bonsai catalog

# Remove an agent
bonsai remove backend
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

Components are mix-and-match per agent:

- **Skills** — coding-standards, testing, database-conventions, design-guide, planning-template, api-design-standards, auth-patterns, and more
- **Workflows** — planning, plan-execution, code-review, reporting, session-logging, security-audit, and more
- **Protocols** — session-start, security, scope-boundaries, memory
- **Sensors** — session-context, scope-guard-files, scope-guard-commands, dispatch-guard, subagent-stop-review, api-security-check, and more
- **Routines** — doc-freshness-check, status-hygiene, roadmap-accuracy, memory-consolidation, and more

## Generated Structure

After `bonsai init` + `bonsai add` (backend agent), your project gets:

```
your-project/
├── .bonsai.yaml              # Project config
├── .claude/
│   └── settings.json         # Auto-generated hook wiring for sensors
├── CLAUDE.md                 # Root router — directs agents to their workspace
├── docs/
│   ├── INDEX.md
│   ├── Playbook/             # Status, Roadmap, Plans, SecurityStandards
│   ├── Logs/                 # FieldNotes, KeyDecisionLog
│   └── Reports/              # Pending reports
└── backend/
    ├── CLAUDE.md             # Agent-specific navigation
    └── agent/
        ├── Core/             # identity.md, memory.md, self-awareness.md
        ├── Skills/           # Selected skill files
        ├── Workflows/        # Selected workflow files
        ├── Protocols/        # Selected protocol files
        └── Sensors/          # Rendered hook scripts
```

## License

MIT
