---
description: Post-install 5-step walkthrough for new Bonsai users.
---

# Quickstart

You just ran `bonsai init`. Here are the five things to do next to get
from an empty workspace to a working agent session.

## 1. Open `CLAUDE.md`

The root `CLAUDE.md` is the navigation entry point that Claude Code reads
when it starts a session in your project. It lists:

- The Tech Lead station (default `station/`)
- Any code agent workspaces you added
- A link to each agent's own `CLAUDE.md` nav table

Skim `station/CLAUDE.md` too — that's the Tech Lead's routing table
into its Core, Protocols, Workflows, Skills, Sensors, and Routines.

## 2. Start a session: "Hi, get started"

Open your project in Claude Code and greet the agent:

```
You:   Hi, get started
Agent: [reads identity.md, memory.md, protocols, project status,
        flags any overdue routines, reports back]
```

Always greet before handing over a task. The greeting triggers the full
session-start sequence — identity load, memory recall, protocol
enforcement, status review. Jumping straight to work skips that context.

## 3. Add a code agent

The Tech Lead plans and reviews — it does not write application code.
Add a code agent for the actual implementation:

```bash
bonsai add
```

Pick one of: `backend`, `frontend`, `fullstack`, `devops`, `security`.
Accept the defaults, review the summary, confirm. Bonsai creates a new
workspace (e.g. `backend/`) with its own `CLAUDE.md`, identity, and
abilities. Sensor hooks are auto-wired into `.claude/settings.json`.

## 4. Understand the Status / Plans / Reports flow

Every non-trivial task flows through the Playbook:

| File | Role |
|------|------|
| `station/Playbook/Status.md`         | Current work — what's active, what's pending |
| `station/Playbook/Backlog.md`        | Unscheduled ideas, bugs, debt, research |
| `station/Playbook/Roadmap.md`        | Long-term direction |
| `station/Playbook/Plans/Active/NN-*` | Implementation plans the Tech Lead writes |
| `station/Reports/Pending/`           | Completion reports from dispatched agents |
| `station/Logs/KeyDecisionLog.md`     | Architectural decisions log |

Typical loop: **Backlog** → **Plan** → **Status (active)** → **dispatch to
code agent** → **Report in Pending/** → **Status (done)**.

## 5. Run routines when prompted

If you installed routines during `bonsai init`, the `routine-check`
sensor runs at every session start. If any routine is overdue, it
surfaces in the session-start output. Ask the Tech Lead to run it:

```
You:   Run the backlog hygiene routine.
Lead:  [follows station/agent/Routines/backlog-hygiene.md step by step,
        writes a report to station/Reports/Pending/]
```

Common routines: `backlog-hygiene` (7d), `dependency-audit` (7d),
`doc-freshness-check` (7d), `status-hygiene` (5d), `memory-consolidation`
(5d), `roadmap-accuracy` (14d), `vulnerability-scan` (7d).

## Cheat keys

| Want | Do |
|------|----|
| See what's installed   | `bonsai list` |
| Browse the catalog     | `bonsai catalog` |
| Add more abilities     | `bonsai add` (picks uninstalled) |
| Remove an ability      | `bonsai remove <type> <name>` |
| Sync after edits       | `bonsai update` |
| Read another topic     | `bonsai guide concepts` / `cli` / `custom-files` |

> Full guide: https://laststep.github.io/Bonsai/guides/your-first-workspace/
