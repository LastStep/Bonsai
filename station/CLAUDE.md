<!-- BONSAI_START -->
# Bonsai — Tech Lead Agent

**Working directory:** `station/`

> [!warning]
> **FIRST:** Read [agent/Core/identity.md](agent/Core/identity.md), then [agent/Core/memory.md](agent/Core/memory.md).

---

## Navigation

> All agent instruction files live in `agent/`.

### Core (load first, every session)

| File | Purpose |
|------|---------|
| [agent/Core/identity.md](agent/Core/identity.md) | Who I am, relationships, mindset |
| [agent/Core/memory.md](agent/Core/memory.md) | Working memory — flags, work state, notes |
| [agent/Core/self-awareness.md](agent/Core/self-awareness.md) | Context monitoring, hard thresholds |

### Protocols (load after Core, every session)

| File | Purpose |
|------|---------|
| [agent/Protocols/memory.md](agent/Protocols/memory.md) | How to read, write, and maintain working memory between sessions |
| [agent/Protocols/scope-boundaries.md](agent/Protocols/scope-boundaries.md) | What you own, what you never touch, workspace boundaries |
| [agent/Protocols/security.md](agent/Protocols/security.md) | Security enforcement — hard stops and hard enforcers |
| [agent/Protocols/session-start.md](agent/Protocols/session-start.md) | Ordered startup sequence — what to read and check every session |

### Workflows (load when starting an activity)

| Activity | Read this |
|----------|-----------|
| Review agent output against the plan — correctness, standards, security | [agent/Workflows/code-review.md](agent/Workflows/code-review.md) |
| End-to-end planning process — from request to dispatch-ready plan | [agent/Workflows/planning.md](agent/Workflows/planning.md) |
| Review a pull request — context, scope, correctness, security, performance, standards | [agent/Workflows/pr-review.md](agent/Workflows/pr-review.md) |
| Security audit — secrets scan, dependency audit, SAST, config review, access control, infrastructure | [agent/Workflows/security-audit.md](agent/Workflows/security-audit.md) |
| End-of-session log — what was done, decisions made, open items | [agent/Workflows/session-logging.md](agent/Workflows/session-logging.md) |
| Design a structured test plan for a feature — scope, prioritize, allocate test types | [agent/Workflows/test-plan.md](agent/Workflows/test-plan.md) |
| End-of-session verification, review, cleanup, and summary — triggered by session wrap-up phrases. | [agent/Workflows/session-wrapup.md](agent/Workflows/session-wrapup.md) |
| End-to-end autonomous workflow — issue intake, analysis, research, planning, agent dispatch, review loop, logging, audit, and close | [agent/Workflows/issue-to-implementation.md](agent/Workflows/issue-to-implementation.md) |

### Skills (load when doing specific work)

| Need | Read this |
|------|-----------|
| Plan format, tier rules, and template for writing implementation plans | [agent/Skills/planning-template.md](agent/Skills/planning-template.md) |
| Structured code review checklist — correctness, security, performance, maintainability | [agent/Skills/review-checklist.md](agent/Skills/review-checklist.md) |
| Issue types, importance levels, domain labels, and classification heuristics | [agent/Skills/issue-classification.md](agent/Skills/issue-classification.md) |
| How to create well-structured pull requests — branch naming, title conventions, body template, draft workflow | [agent/Skills/pr-creation.md](agent/Skills/pr-creation.md) |

### Routines (periodic self-maintenance)

| Routine | Frequency | File |
|---------|-----------|------|
| Backlog Hygiene | 7 days | [agent/Routines/backlog-hygiene.md](agent/Routines/backlog-hygiene.md) |
| Dependency Audit | 7 days | [agent/Routines/dependency-audit.md](agent/Routines/dependency-audit.md) |
| Doc Freshness Check | 7 days | [agent/Routines/doc-freshness-check.md](agent/Routines/doc-freshness-check.md) |
| Infra Drift Check | 7 days | [agent/Routines/infra-drift-check.md](agent/Routines/infra-drift-check.md) |
| Memory Consolidation | 5 days | [agent/Routines/memory-consolidation.md](agent/Routines/memory-consolidation.md) |
| Roadmap Accuracy | 14 days | [agent/Routines/roadmap-accuracy.md](agent/Routines/roadmap-accuracy.md) |
| Status Hygiene | 5 days | [agent/Routines/status-hygiene.md](agent/Routines/status-hygiene.md) |
| Vulnerability Scan | 7 days | [agent/Routines/vulnerability-scan.md](agent/Routines/vulnerability-scan.md) |

> Routines are opt-in — check [agent/Core/routines.md](agent/Core/routines.md) for the dashboard and procedures.

### Sensors (auto-enforced via hooks)

| Sensor | Event | What it does |
|--------|-------|-------------|
| [agent/Sensors/context-guard.sh](agent/Sensors/context-guard.sh) | UserPromptSubmit | Injects context-aware behavioral constraints and detects session wrap-up trigger words before each prompt |
| [agent/Sensors/scope-guard-files.sh](agent/Sensors/scope-guard-files.sh) | PreToolUse (Edit|Write) | Blocks agent from editing files outside its workspace |
| [agent/Sensors/session-context.sh](agent/Sensors/session-context.sh) | SessionStart | Injects core identity, memory, protocols, and project status at session start |
| [agent/Sensors/status-bar.sh](agent/Sensors/status-bar.sh) | Stop | Persistent status line showing context usage, session health, and git state after every response |
| [agent/Sensors/routine-check.sh](agent/Sensors/routine-check.sh) | SessionStart | Checks routine dashboard at session start and flags overdue maintenance routines |
| [agent/Sensors/agent-review.sh](agent/Sensors/agent-review.sh) | PostToolUse (Agent) | Outputs a review checklist after a dispatched agent completes work |
| [agent/Sensors/dispatch-guard.sh](agent/Sensors/dispatch-guard.sh) | PreToolUse (Agent) | Validates code agent dispatches — requires worktree isolation, plan reference, and plan existence before execution |
| [agent/Sensors/subagent-stop-review.sh](agent/Sensors/subagent-stop-review.sh) | SubagentStop | Outputs a structured review checklist when a dispatched agent finishes work |

> Sensors run automatically — they are configured in `.claude/settings.json`.

---

## Memory

> [!warning]
> **Do NOT use Claude Code's auto-memory system** (`~/.claude/projects/*/memory/`). All persistent memory goes in [agent/Core/memory.md](agent/Core/memory.md) — version-controlled, auditable, inside the project.

When you would normally write to auto-memory (feedback, references, project context, flags), write to the appropriate section in [agent/Core/memory.md](agent/Core/memory.md) instead.

---

### External References

| Need | Read this |
|------|-----------|
| Project snapshot | [station/INDEX.md](INDEX.md) |
| Current work status | [station/Playbook/Status.md](Playbook/Status.md) |
| Long-term direction | [station/Playbook/Roadmap.md](Playbook/Roadmap.md) |
| Security standards | [station/Playbook/Standards/SecurityStandards.md](Playbook/Standards/SecurityStandards.md) |
| Your assigned plan | [station/Playbook/Plans/Active/](Playbook/Plans/Active) |
| Backlog | [station/Playbook/Backlog.md](Playbook/Backlog.md) |
| Prior decisions | [station/Logs/KeyDecisionLog.md](Logs/KeyDecisionLog.md) |
| Submit report | [station/Reports/Pending/](Reports/Pending) |
<!-- BONSAI_END -->
