---
description: The mental model — station, instruction stack, agents, sensors, routines, scaffolding.
---

# Concepts

Bonsai is a **generator**, not a runtime. It runs once to set up the
instruction layer that Claude Code agents read when they work in your
project. After that, the agents operate using the files it created.

## The Station

Every Bonsai project has one **station** — a single directory (default
`station/`) that is the command center. It holds two things:

- **Project scaffolding** — shared state every agent reads (Status,
  Roadmap, Backlog, Plans, Standards, Logs, Reports).
- **The Tech Lead agent** — the primary orchestrator. Plans, reviews,
  runs routines. Never writes application code.

Code agents (backend, frontend, etc.) live in their own workspaces but
reference the station for plans, status, and standards.

## The Instruction Stack

Each agent's instructions are layered, from foundational to automated:

```
  Layer 6  Sensors    Automated enforcement (hook scripts)
  Layer 5  Routines   Periodic self-maintenance (Tech Lead, opt-in)
  Layer 4  Skills     Domain knowledge and standards
  Layer 3  Workflows  Step-by-step task procedures
  Layer 2  Protocols  Hard rules — security, scope, startup
  Layer 1  Core       Identity, memory, self-awareness
```

| Layer     | Loaded when                               | Override? |
|-----------|-------------------------------------------|-----------|
| Core      | First, every session                      | No        |
| Protocols | Every session, via startup                | No        |
| Workflows | When starting a specific task             | Followed  |
| Skills    | Referenced on demand during work          | Reference |
| Routines  | Flagged at session start if overdue       | On demand |
| Sensors   | Automatically, on Claude Code events      | Can't bypass |

## Agents as Teammates

The Tech Lead orchestrates. Code agents implement.

```
  You (human)
   └── Tech Lead (station/)       plans, reviews, orchestrates, runs routines
        ├── Backend   (backend/)   API, database, server logic
        ├── Frontend  (frontend/)  UI, components, styling
        ├── DevOps    (devops/)    infra-as-code, CI/CD
        └── Security  (security/)  audits, scanning
```

You can talk to code agents directly for quick fixes. The Tech Lead
earns its keep on work that needs planning, coordination, or
cross-agent review.

## Sensors — Auto-enforced via Hooks

Sensors are shell scripts wired into Claude Code hook events
(`SessionStart`, `PreToolUse`, `PostToolUse`, `UserPromptSubmit`,
`Stop`, `SubagentStop`). They run automatically — the agent cannot
bypass them. Examples:

- `session-context` injects identity + memory at session start
- `scope-guard-files` blocks edits outside an agent's workspace
- `dispatch-guard` validates worktree isolation before dispatching
- `status-bar` prints a live status line after every response

## Routines — Periodic Self-maintenance

Routines are scheduled procedures the Tech Lead runs on request.
Each has a `frequency` (e.g. `7 days`). The `routine-check` sensor
parses `agent/Core/routines.md` at session start and flags overdue
items. Ship-along routines include backlog hygiene, dependency
audit, doc freshness, memory consolidation, roadmap accuracy,
status hygiene, vulnerability scan.

## Scaffolding — Shared Project State

Scaffolding is shared, human-readable state that outlives any
single session:

| Artifact     | Purpose |
|--------------|---------|
| `INDEX.md`   | Project snapshot |
| `Playbook/Status.md`    | Active + pending work |
| `Playbook/Roadmap.md`   | Long-term direction |
| `Playbook/Backlog.md`   | Intake queue for unscheduled items |
| `Playbook/Plans/`       | Implementation plans |
| `Playbook/Standards/`   | Security, coding, review standards |
| `Logs/FieldNotes.md`    | Working notes |
| `Logs/KeyDecisionLog.md`| Architecture decisions |
| `Logs/RoutineLog.md`    | Routine execution history |
| `Reports/`              | Completion reports + templates |

## When to Use What

| You are about to... | Load this layer |
|---------------------|-----------------|
| Start a session                  | Core + Protocols (auto) |
| Plan a feature                   | Workflow: planning |
| Review a PR                      | Workflow: pr-review |
| Apply a coding standard          | Skill: coding-standards / design-guide |
| Edit a file outside your scope   | Blocked by sensor: scope-guard-files |
| Run weekly maintenance           | Routine dashboard at `agent/Core/routines.md` |

> Full guide: https://laststep.github.io/Bonsai/concepts/how-bonsai-works/
