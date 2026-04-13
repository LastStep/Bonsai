# Bonsai Handbook

A guide to understanding and using the agent workspaces Bonsai generates. If you haven't installed Bonsai yet, start with the [README](README.md).

---

**Table of Contents**

- [What Bonsai Actually Does](#what-bonsai-actually-does)
- [The Mental Model](#the-mental-model)
- [How the Pieces Fit Together](#how-the-pieces-fit-together)
- [Interaction Patterns](#interaction-patterns) — _the most important section_
- [Understanding Sensors](#understanding-sensors)
- [Understanding Routines](#understanding-routines)
- [Choosing the Right Setup](#choosing-the-right-setup)
- [Tips and Best Practices](#tips-and-best-practices)

---

## What Bonsai Actually Does

Bonsai is a **generator**, not a runtime. It runs once to set up the instruction layer that Claude Code agents read when they work in your project. After that, the agents operate using the files it created.

| Generated layer | Purpose |
|:----------------|:--------|
| **Identity + Memory** | Who the agent is, what it remembers from last session |
| **Protocols** | Hard rules — security, scope boundaries, startup sequence |
| **Workflows** | Step-by-step procedures — planning, code review, reporting |
| **Skills** | Domain knowledge — coding standards, API design, auth patterns |
| **Sensors** | Shell scripts wired into Claude Code hooks — auto-enforce boundaries |
| **Routines** | Periodic maintenance with a tracked dashboard (Tech Lead only) |
| **Scaffolding** | Shared docs — status tracker, roadmap, plans, decision log, reports |

> [!NOTE]
> Claude Code agents are only as good as the instructions they have access to. Bonsai gives them a structured, layered instruction set so they behave consistently, stay in scope, and coordinate with each other.

---

## The Mental Model

### The Station

Every Bonsai project has a **station** — a single directory (default `station/`) that serves as the command center. It contains:

- **Project scaffolding** — status tracker, roadmap, plans, logs, reports
- **Tech Lead agent** — the primary agent that orchestrates everything

The station is where all coordination happens. Code agents (backend, frontend, etc.) get their own separate workspaces, but they all reference the station for plans, status, and standards.

### Agents as Team Members

Think of each agent as a new hire. They need:

| What they need | Bonsai equivalent |
|:---------------|:-----------------|
| Role description | **Identity** — job, mindset, relationships |
| Company rules | **Protocols** — security policy, boundaries |
| Playbooks | **Workflows** — how to plan, review, report |
| Domain training | **Skills** — standards, patterns, conventions |
| Automated guardrails | **Sensors** — catch mistakes before they happen |
| Maintenance duties | **Routines** — periodic tasks on a schedule (Tech Lead only) |

When an agent starts a session, it reads its identity, checks its memory, follows its startup protocol, and gets context injected by sensors — all before you ask it anything.

### The Instruction Stack

Each agent's instructions are layered, from foundational to automated:

```
  Layer 6 │ Sensors       │ Automated enforcement (hook scripts)
  Layer 5 │ Routines      │ Periodic maintenance procedures (Tech Lead)
  Layer 4 │ Skills        │ Domain knowledge and standards
  Layer 3 │ Workflows     │ Step-by-step task procedures
  Layer 2 │ Protocols     │ Hard rules (security, scope, startup)
  Layer 1 │ Core          │ Identity + memory + self-awareness
```

| Layer | Loaded when | Can override? |
|:------|:-----------|:-------------|
| **Core** | Every session, first thing | No — this is the foundation |
| **Protocols** | Every session, via startup | No — non-negotiable rules |
| **Workflows** | When performing a specific task | Agent follows the procedure |
| **Skills** | Referenced as needed during work | Knowledge, not instructions |
| **Routines** | Flagged at session start if overdue | Tech Lead runs on request |
| **Sensors** | Automatically, on Claude Code events | Agent can't bypass them |

### How It Works

`bonsai init` sets up the station with the Tech Lead. You talk to the Tech Lead; it plans, orchestrates, and reviews. Code agents execute.

```
  You (human)
   └── Tech Lead (station/)      plans, reviews, orchestrates, runs routines
        ├── Backend (backend/)    API, database, server logic
        ├── Frontend (frontend/)  UI, components, styling
        ├── DevOps (devops/)      infrastructure, CI/CD
        └── Security (security/)  audits, scanning
```

> [!IMPORTANT]
> The Tech Lead **never writes application code**. It creates plans, dispatches them to code agents via worktree-isolated subagents, and reviews the output. The `dispatch-guard` sensor enforces that every dispatch has a plan and uses isolation. The `subagent-stop-review` sensor triggers a review checklist when a code agent finishes.

> [!NOTE]
> **You don't have to go through the Tech Lead for everything.** Quick fixes and small tasks work fine directly with a code agent. The Tech Lead shines when work needs planning, coordination, or cross-agent review.

---

## How the Pieces Fit Together

### The Station — Command Center

Created by `bonsai init`. The Tech Lead lives here alongside the project scaffolding:

| File | Purpose | Used by |
|:-----|:--------|:--------|
| `station/CLAUDE.md` | Tech Lead agent navigation | Tech Lead |
| `station/agent/` | Tech Lead instruction layer (Core, Protocols, Workflows, etc.) | Tech Lead |
| `station/INDEX.md` | Project snapshot — tech stack, phase, document registry | Every agent at session start |
| `station/Playbook/Status.md` | Live task tracker — in progress, pending, done | Agents check before starting work |
| `station/Playbook/Roadmap.md` | Long-term milestones and phases | Context for where tasks fit |
| `station/Playbook/Plans/Active/` | Numbered implementation plans | Tech Lead writes, code agents execute |
| `station/Playbook/Standards/SecurityStandards.md` | Hard security rules across all agents | Security protocol references this |
| `station/Logs/FieldNotes.md` | **Your** notes to the agents | Read every session start |
| `station/Logs/KeyDecisionLog.md` | Settled architectural decisions | Checked before proposing alternatives |
| `station/Logs/RoutineLog.md` | Routine execution history | Updated after each routine run |
| `station/Reports/Pending/` | Unreviewed completion reports | Tech Lead reviews during code review |

### Code Agent Workspaces — The Private Layer

Each code agent gets its own directory. This is what it owns and is allowed to modify:

```
backend/
├── CLAUDE.md              # agent-specific navigation
└── agent/
    ├── Core/              # identity.md, memory.md, self-awareness.md
    ├── Protocols/         # hard rules
    ├── Workflows/         # task procedures
    ├── Skills/            # domain knowledge
    └── Sensors/           # rendered hook scripts (.sh)
```

### Config and Hooks

| File | What it does | Managed by |
|:-----|:------------|:-----------|
| `.bonsai.yaml` | Single source of truth — all agents, components, paths | `bonsai init` / `bonsai add` / `bonsai remove` |
| `.claude/settings.json` | Hook entries for every sensor | Auto-generated by Bonsai |

---

## Interaction Patterns

> [!TIP]
> This is the most important section. These patterns help you get consistent, high-quality output from the agent system.

### Starting a Session

**Always open with a greeting before your actual task.**

```
You:    Hi, get started
Agent:  [reads identity, memory, protocols, checks status, reviews warnings...]
You:    Great. Now here's what I need...
```

> [!WARNING]
> If you skip the greeting and jump straight into a task, the agent may not fully process its startup context. It might miss overdue routines, pending reports, or status changes.

**What happens on startup:**

| Step | What fires | What it does |
|:-----|:----------|:-------------|
| 1 | `session-context` sensor | Injects identity, memory, protocols, INDEX, status, field notes |
| 2 | `session-start` protocol | Agent reads `Core/identity.md` and `Core/memory.md` |
| 3 | Agent initiative | Checks `Status.md`, `FieldNotes.md`, reviews warnings |
| 4 | `routine-check` sensor | Flags overdue routines (Tech Lead only) |

After this, the agent knows who it is, what happened last session, and what the current state looks like.

---

### Planning Work

> _Use with: **Tech Lead**_

Be specific about _what_ you want. Leave the _how_ to the agent.

```
You:    Hi, get started
Lead:   [processes startup context]
You:    I need user authentication — email/password login, JWT tokens,
        and a protected dashboard route.
Lead:   [researches codebase, asks clarifying questions, writes plan, self-reviews]
```

**Good planning prompts:**

| Pattern | Example |
|:--------|:--------|
| New feature | _"I need [feature]. Here's what it should do: [requirements]."_ |
| Bug fix | _"We have a bug where [symptom]. Investigate and plan a fix."_ |
| Refactor | _"Refactor [area] to [goal]. Keep backward compatibility."_ |

The plan ends up in `station/Playbook/Plans/Active/` as a numbered document. **Review it before dispatching** — it's cheaper to fix a plan than to fix code.

---

### Dispatching to Code Agents

> _Use with: **Tech Lead**_

```
You:    Dispatch plan 003 to the backend agent
Lead:   [dispatch-guard validates -> dispatches with worktree isolation]
```

The `dispatch-guard` sensor enforces three things:

| Check | What it validates |
|:------|:-----------------|
| **Isolation** | Dispatch uses worktree isolation (work on a branch, not main) |
| **Plan reference** | Prompt mentions a plan number or path |
| **Plan exists** | The plan file is actually in `Plans/Active/` |

**When the code agent finishes**, the `subagent-stop-review` sensor auto-triggers a checklist:

1. Review output against the plan — every step followed, nothing improvised
2. Check security against `SecurityStandards.md`
3. Verify that verification steps from the plan passed
4. Log results, update status, process pending reports

---

### Executing Plans

> _Use with: **Code agents** (backend, frontend, fullstack, devops)_

```
You:    Hi, get started
Agent:  [processes startup context]
You:    Execute plan 003
Agent:  [reads plan -> implements step by step -> tests -> reports]
```

The agent will:
1. Read and understand all steps in the plan
2. Implement each step, checking off as it goes
3. Run tests
4. Write a completion report to `station/Reports/Pending/`
5. Update `station/Playbook/Status.md`

**For partial execution:** `"Execute steps 1-3 of plan 003"` or `"Just do the database migration part"`

---

### Reviewing Work

> _Use with: **Tech Lead**_

```
You:    Hi, get started
Lead:   [notices pending reports in Reports/Pending/]
You:    Review the backend agent's work on plan 003
Lead:   [loads plan -> checks completeness -> reviews quality + security -> verdict]
```

**Review verdicts:**

| Verdict | Meaning |
|:--------|:--------|
| **Pass** | Merge the work |
| **Revise** | Specific changes needed — sends back to code agent |
| **Escalate** | Architectural issue requiring human decision |

**Targeted review prompts:**
- _"Review just the API endpoints from plan 003"_
- _"Check if the auth implementation follows our security standards"_
- _"Are there any test coverage gaps?"_

---

### Security Audits

> _Use with: **Security** agent (or any agent with the `security-audit` workflow)_

```
You:    Hi, get started
Agent:  [processes startup context]
You:    Run a security audit on the authentication module
Agent:  [secrets scan -> dependency audit -> SAST -> config review -> access control]
```

> [!TIP]
> Scoped audits work better than broad ones:
> - _"Audit the API routes for injection vulnerabilities"_
> - _"Check our dependency tree for known CVEs"_
> - _"Review session management against OWASP guidelines"_

---

### Routines

> _Use with: **Tech Lead**_

Routines are periodic maintenance tasks that only the Tech Lead runs. They keep the project healthy — documentation, dependencies, security, infrastructure.

```
You:    Hi, get started
Lead:   [startup flags: "OVERDUE: dependency-audit (last ran 12 days ago, due every 7)"]
You:    Run the dependency audit
Lead:   [follows procedure -> logs to RoutineLog.md -> updates dashboard]
```

**Proactive prompts:**
- _"Are any routines overdue?"_
- _"Run all overdue routines"_
- _"Run the vulnerability scan"_

Each routine has a concrete, step-by-step procedure — the agent doesn't improvise.

---

### Project Status

```
You:    What's the current project status?
Lead:   [reads Status.md, Roadmap.md, checks pending reports, reviews logs]
```

| What you want | What to say |
|:-------------|:-----------|
| Current work | _"What's in progress right now?"_ |
| Pending reviews | _"Are there any pending reports?"_ |
| What's next | _"What's next on the roadmap?"_ |
| Past decisions | _"Show me recent entries from the decision log"_ |

---

### Context Between Sessions

The agent's memory resets each conversation. Three things bridge the gap:

| Bridge | Who writes it | What it does |
|:-------|:-------------|:-------------|
| `Core/memory.md` | The agent | Working memory — current state, flags, notes |
| `station/Logs/FieldNotes.md` | **You** | Your notes — changes outside sessions, new requirements |
| `station/Playbook/Status.md` | The agent | Task tracker — done, in progress, pending |

> [!TIP]
> If the agent seems to have lost context:
> - _"Re-read your memory file"_
> - _"Check FieldNotes — I added some notes"_
> - _"Review the current status"_

---

### Ending a Session

```
You:    Let's wrap up this session
Agent:  [updates memory.md -> writes session log -> files report -> updates Status.md]
```

This ensures the next session starts with full context. If you forget, startup still works — but the context will be less precise.

---

### Multi-Agent Workflow

A typical feature cycle with the Tech Lead orchestrating:

```
Session 1 — Planning
  You:    Hi, get started
  You:    I need a REST API for user profiles — CRUD + pagination
  Lead:   [asks questions, writes plan 004]
  You:    Looks good, dispatch to backend

Session 2 — Review + Next Phase
  You:    Hi, get started
  Lead:   [sees pending report from backend agent]
  You:    Review the backend work on plan 004
  Lead:   [verdict: pass]
  You:    Plan the frontend — profile page with edit form
  Lead:   [writes plan 005, dispatches to frontend]

Session 3 — Integration
  You:    Hi, get started
  You:    Review frontend work, then security review both
```

---

## Understanding Sensors

Sensors are shell scripts wired into Claude Code's hook system. They fire automatically — you don't invoke them.

### Events

| Event | When it fires | Can block? |
|:------|:-------------|:-----------|
| `SessionStart` | Beginning of a conversation | No |
| `UserPromptSubmit` | Before Claude processes a user message | **Yes** (exit code 2) |
| `PreToolUse` | Before a tool executes | **Yes** (exit code 2) |
| `PostToolUse` | After a tool executes | No |
| `Stop` | After every Claude response | **Yes** (exit code 2) |
| `SubagentStop` | When a dispatched subagent finishes | No |

### Sensor Reference

| Sensor | Event | Matcher | Agents | What it does |
|:-------|:------|:--------|:-------|:-------------|
| **status-bar** | `Stop` | — | all | Persistent status: context %, turns, tools, git state, memory/routine health |
| **context-guard** | `UserPromptSubmit` | — | all | Injects tiered behavioral constraints + detects "session done" trigger words |
| **session-context** | `SessionStart` | — | all | Injects identity, memory, protocols, INDEX, status, field notes, health warnings |
| **scope-guard-files** | `PreToolUse` | `Edit\|Write` | all | **Blocks** edits outside workspace + `.env` files |
| **scope-guard-commands** | `PreToolUse` | `Bash` | tech-lead, security | **Blocks** app execution commands (tests, builds, servers) |
| **dispatch-guard** | `PreToolUse` | `Agent` | tech-lead | **Blocks** dispatches without worktree isolation or plan reference |
| **subagent-stop-review** | `SubagentStop` | — | tech-lead | Outputs review checklist when subagent finishes |
| **api-security-check** | `PreToolUse` | `Edit\|Write` | backend, fullstack, security | Detects SQL injection, hardcoded secrets, `eval()`, CORS wildcards |
| **test-integrity-guard** | `PreToolUse` | `Edit\|Write` | backend, frontend, fullstack | Catches `.skip()`, `.only()`, removed assertions, empty tests |
| **iac-safety-guard** | `PreToolUse` | `Bash` | devops | **Blocks** `terraform destroy`, `kubectl delete namespace`, unsafe docker ops |

> [!IMPORTANT]
> `PreToolUse` sensors can **block actions** by exiting with code 2. The tool call is rejected before it happens. This is how `scope-guard-files` prevents cross-workspace edits — the agent can't bypass it.

### Awareness Sensors

Two sensors work as a pair to give agents real-time self-awareness:

- **status-bar** fires after every response — shows the user a compact status line with context usage %, turn count, and session health warnings (uncommitted files, stale memory, overdue routines)
- **context-guard** fires before every prompt — reads session state and injects behavioral constraints the agent must follow. At 30% context, it nudges toward conciseness. At 70%+, it restricts the agent to current-task-only. At 85%+, it tells the agent to refuse new work.

**Session wrap-up:** When the user says "session done" (or similar), context-guard injects a structured checklist — commit check, memory update, backlog review, status update, session notes.

---

## Understanding Routines

Periodic maintenance tasks that keep the project healthy. **Routines are managed exclusively by the Tech Lead** — code agents don't run them.

### How They Work

```
1. Each routine has a frequency and a concrete procedure
2. Dashboard at station/agent/Core/routines.md tracks last-ran and next-due
3. routine-check sensor flags overdue items at session start
4. You tell the Tech Lead to run it -> follows procedure -> updates dashboard
```

> The `routine-check` sensor is auto-managed — Bonsai adds it when you install routines, removes it when you remove all routines.

### Routine Reference

| Routine | Freq | What it does |
|:--------|:-----|:-------------|
| **dependency-audit** | 7d | Scan dependencies for known CVEs + unmaintained packages |
| **vulnerability-scan** | 7d | SAST scan, secrets scan, dependency cross-reference |
| **doc-freshness-check** | 7d | Find docs that drifted from recent code changes |
| **infra-drift-check** | 7d | Compare declared IaC state vs actual cloud resources |
| **status-hygiene** | 5d | Archive done items, validate pending items in Status.md |
| **memory-consolidation** | 5d | Clean up and validate working memory entries |
| **roadmap-accuracy** | 14d | Ensure Roadmap.md matches what's actually built + planned |

---

## Choosing the Right Setup

Tech Lead is always installed during `bonsai init`. After that, add the code agents your project needs:

| Scenario | Code agents to add | Why |
|:---------|:-------------------|:----|
| **Solo developer** | `fullstack` | One code agent covers UI, API, DB, auth, tests |
| **Separated concerns** | `backend` + `frontend` | Domain-specific conventions and skills |
| **With infrastructure** | + `devops` | IaC sensors (`iac-safety-guard`) + infra-drift routine |
| **Security-conscious** | + `security` | Vulnerability scanning, dependency audits, security-audit workflow |

<details>
<summary><strong>Full team setup</strong></summary>

```bash
bonsai init       # station + Tech Lead
bonsai add        # backend     — API, database, server logic
bonsai add        # frontend    — UI, state management, styling
bonsai add        # devops      — infrastructure, CI/CD, containers
bonsai add        # security    — audits, scanning, compliance
```

</details>

> [!TIP]
> **Start small.** You don't need every code agent from day one. Begin with one, add more as complexity demands. `bonsai add` is designed to be run incrementally.

---

## Tips and Best Practices

| Practice | Why it matters |
|:---------|:--------------|
| **Always say "Hi, get started"** | Ensures the agent processes full startup context before working |
| **Keep `FieldNotes.md` updated** | The agent reads this every session — your bridge for out-of-session context |
| **Plan before executing** | A 5-min planning session catches edge cases and produces a clear spec |
| **Review plans before dispatching** | Cheaper to fix a plan than to fix implemented code |
| **Reference plans by number** | `"Execute plan 003"` is unambiguous; `"do the auth work"` isn't |
| **Use the decision log** | Settled decisions stop agents from re-proposing alternatives |
| **End with "let's wrap up"** | Agent updates memory + writes session log for next time |
| **Run overdue routines** | They catch real issues — stale docs, vulnerable deps, drifted infra |
| **Check `Reports/Pending/`** | Review queue for code agent completion reports |
| **Don't fight scope boundaries** | If a sensor blocks it, it's usually right — plan cross-workspace work through the Tech Lead |
