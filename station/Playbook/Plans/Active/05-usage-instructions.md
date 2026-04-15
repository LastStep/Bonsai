# Plan 05 — Usage Instructions: AI Operational Intelligence

**Tier:** 2 (Feature)
**Status:** Active — PR #7 ready for review
**Agent:** tech-lead (orchestration), general-purpose (implementation)

## Goal

Every AI agent working in a Bonsai-generated workspace understands how to operate effectively from session one — when to use workflows vs. skills, how to leverage the Playbook, how to evolve the workspace — via a compact "How to Work" section in the generated CLAUDE.md and a detailed `workspace-guide` skill for deeper reference.

## Context

Research revealed that Bonsai workspaces tell agents *where things are* (navigation tables in CLAUDE.md) but not *how to use them effectively*. Agents don't know when to reach for specific workflows, how routines work, or how to leverage the Playbook. The fix is a two-layer approach:

1. **Layer 1:** Compact decision heuristics (~15-25 lines) injected into every agent's generated CLAUDE.md — always loaded, near-zero context cost
2. **Layer 2:** Detailed operational skill (`workspace-guide`) loaded on-demand — covers the full system in depth, discoverable via pointer in Layer 1

## Steps

### Step 1: Create the `workspace-guide` skill catalog entry

**1a.** Create `catalog/skills/workspace-guide/meta.yaml`:

```yaml
name: workspace-guide
description: Bonsai workspace operational guide — how to use workflows, skills, routines, sensors, and the Playbook effectively
agents: all
```

**1b.** Create `catalog/skills/workspace-guide/workspace-guide.md.tmpl` with the following structure. Use `{{ .AgentName }}` conditionals to tailor content per agent type.

The file should have this structure (all sections mandatory):

```
---
tags: [skill, workspace]
description: How to operate effectively in a Bonsai workspace — decision patterns, system mechanics, workspace evolution.
---

# Workspace Guide

## How This Workspace Is Organized
- Explain the workspace structure: Core → Protocols → Skills/Workflows → Sensors/Routines
- Core + Protocols load every session (foundational)
- Skills + Workflows load on-demand (reference + procedures)
- Sensors run automatically (event-driven hooks)
- Routines are periodic maintenance (opt-in per session)

## When to Load What
- Workflow: starting a multi-step activity (planning, reviewing, auditing, reporting). Follow end-to-end.
- Skill: need reference standards for a specific domain. Load only while relevant.
- Both: check the CLAUDE.md navigation tables for the full list.

## Using the Playbook
- Status.md: your task tracker — check for assigned work at session start
- Backlog.md: intake queue — add out-of-scope findings here, never fix inline
- Roadmap.md: long-term direction — check before proposing new features
- Plans/Active/: your assigned implementation plans
- KeyDecisionLog.md: significant architectural decisions — log yours here
- SecurityStandards.md: hard security constraints for all work

## How Routines Work
(Only render this section if agent has routines: {{ if .Routines }})
- Dashboard at agent/Core/routines.md tracks schedule
- Session-start hook flags overdue routines — user decides to run or defer
- Execution: read the routine file in agent/Routines/, follow procedure, log results to Logs/RoutineLog.md, update dashboard
- All routines must be idempotent

## How Sensors Work
- Auto-enforced via .claude/settings.json hooks
- You don't invoke sensors — they fire on events (SessionStart, PreToolUse, PostToolUse, Stop, UserPromptSubmit)
- Some block actions (scope guards), some inject context (session-context), some provide feedback (status-bar)
- Don't fight sensors — if a sensor blocks you, it's enforcing a rule. Read the rule.

## Working with Memory
- agent/Core/memory.md is your persistent memory between sessions
- Structure: Flags (temporary signals), Work State (current task), Notes (stable facts), Feedback (user corrections), References (external pointers)
- Write to memory when you learn something future sessions need
- Read memory at session start to restore context

## Evolving the Workspace
- `bonsai add` — install new agents or add abilities to existing agents
- `bonsai remove` — remove agents or individual abilities
- `bonsai update` — sync workspace after creating custom files
- `bonsai guide` — view the custom files guide
- `bonsai list` — see what's installed
- `bonsai catalog` — browse all available abilities
- Custom files: create files with YAML frontmatter in the right directory, run `bonsai update` to track them

## Agent-Specific Patterns

### For Tech Lead ({{ if eq .AgentName "tech-lead" }})
- You plan and orchestrate — never write application code directly
- Dispatch implementation to code agents via isolated worktrees
- Check Backlog before creating new work items — avoid duplicates
- After completing work: update Status.md, log to RoutineLog.md
- Use the issue-to-implementation workflow for end-to-end orchestration
- Review all agent output against the plan before marking done

### For Code Agents (backend, frontend, fullstack) ({{ if or (eq ...) }})
- Read your assigned plan in Playbook/Plans/Active/ before writing any code
- Follow the plan exactly — if ambiguous, stop and ask rather than guessing
- Stay within your workspace boundary (enforced by scope-guard sensor)
- Report completion via the reporting workflow
- Don't make architectural decisions — that's the Tech Lead's job

### For DevOps ({{ if eq .AgentName "devops" }})
- Always plan before apply — never auto-approve destructive operations
- Require explicit user confirmation for infrastructure changes
- The iac-safety-guard sensor blocks dangerous commands — work with it

### For Security ({{ if eq .AgentName "security" }})
- You audit and report — you don't implement application features
- Every finding needs: specific file, line number, and standard reference (OWASP, CWE, CVE)
- Only modify security-owned files
- Use the security-audit workflow for structured audits
```

**Important template implementation notes:**
- Use Go template conditionals: `{{ if eq .AgentName "tech-lead" }}...{{ end }}`
- For code agents group: `{{ if or (eq .AgentName "backend") (eq .AgentName "frontend") (eq .AgentName "fullstack") }}`
- Sections that apply to all agents render unconditionally
- The Routines section uses `{{ if .Routines }}` to only appear when routines are installed
- Keep total rendered output to ~150-200 lines per agent type (it's a skill, not a novel)

### Step 2: Add "How to Work" section to generated CLAUDE.md

Modify `internal/generate/generate.go` — the `WorkspaceClaudeMD` function.

**2a.** Add a new helper function `howToWorkLines` that generates the compact heuristics:

```go
func howToWorkLines(agentName string, docsPrefix string, hasRoutines bool, hasWorkspaceGuide bool) []string {
```

Parameters:
- `agentName` — from `agentDef.Name` (e.g., "tech-lead", "backend")
- `docsPrefix` — the docs path prefix (e.g., "station/")
- `hasRoutines` — `len(installed.Routines) > 0`
- `hasWorkspaceGuide` — whether workspace-guide is in the installed skills list

Returns `[]string` of lines to append.

**2b.** The function generates these lines:

Section header:
```markdown
### How to Work

> Decision heuristics — how to use this workspace effectively.
```

**Shared heuristics (all agents):**
```markdown
- **Before starting work:** Check `{docsPrefix}Playbook/Status.md` for assigned tasks and `{docsPrefix}Playbook/Plans/Active/` for your current plan.
- **When to load a Workflow:** You are starting a multi-step activity (planning, reviewing, auditing). Load the matching workflow from the table above and follow it end-to-end.
- **When to load a Skill:** You need reference standards for a specific domain (coding style, API design, test strategy). Load it, use it, move on.
- **Decision logging:** When you make or observe a significant architectural decision, append it to `{docsPrefix}Logs/KeyDecisionLog.md`.
- **Out-of-scope findings:** Don't fix bugs, debt, or improvements outside your current task. Add them to `{docsPrefix}Playbook/Backlog.md`.
- **Workspace evolution:** `bonsai add` (new abilities), `bonsai remove` (uninstall), `bonsai update` (sync custom files), `bonsai list` (see installed), `bonsai catalog` (browse available).
```

**Tech Lead only (when agentName == "tech-lead"):**
```markdown
- **You orchestrate, not implement.** Plan features, dispatch to code agents via worktrees, review their output. Never write application code directly.
- **Check Backlog first:** Before creating new work items, check `{docsPrefix}Playbook/Backlog.md` for existing entries.
- **After completing work:** Update `{docsPrefix}Playbook/Status.md` and log results.
```

**Code agents (backend, frontend, fullstack):**
```markdown
- **Plan first:** Read your assigned plan in `{docsPrefix}Playbook/Plans/Active/` before writing any code. Follow it exactly.
- **When stuck:** If the plan is ambiguous, stop and report — don't guess or make design decisions.
- **Stay in scope:** Only modify files within your workspace boundary.
```

**DevOps:**
```markdown
- **Plan before apply:** Never auto-approve destructive infrastructure operations. Require explicit user confirmation.
- **Stay in scope:** Only modify infrastructure and deployment files within your workspace boundary.
```

**Security:**
```markdown
- **Audit and report.** You read the entire codebase but only modify security-owned files.
- **Evidence-based findings:** Every finding must reference a specific file, line, and standard (OWASP, CWE, CVE).
```

**Workspace guide pointer (when hasWorkspaceGuide is true):**
```markdown
- **New to this workspace?** Load `agent/Skills/workspace-guide.md` for a full operational reference.
```

**2c.** In `WorkspaceClaudeMD`, call `howToWorkLines` and insert the returned lines **after the sensors section and before the Memory section**. Specifically, insert between current line 624 (end of sensors block) and line 626 (start of `"---", "", "## Memory"`).

Current code at insertion point (around line 624-626):
```go
    // ... end of sensors section ...
    lines = append(lines, "",
        "> Sensors run automatically — they are configured in `.claude/settings.json`.", "")
    }

    lines = append(lines,
        "---", "",
        "## Memory", "",
```

Insert:
```go
    // How to Work section
    hasWorkspaceGuide := false
    for _, s := range installed.Skills {
        if s == "workspace-guide" {
            hasWorkspaceGuide = true
            break
        }
    }
    htw := howToWorkLines(agentDef.Name, docsPrefix, len(installed.Routines) > 0, hasWorkspaceGuide)
    lines = append(lines, htw...)

    lines = append(lines,
        "---", "",
        "## Memory", "",
```

**2d.** Add a "How to Work" entry to the Core navigation table (around line 518-524) so agents can see it in the table of contents:

After the self-awareness row, add:
```
| (see below) | **How to Work** — decision heuristics for using this workspace |
```

Actually, don't add it to the Core table — it's a section, not a file. The section heading `### How to Work` is visible in the navigation flow naturally. No table entry needed.

### Step 3: Update generate_test.go

**3a.** Add a test case to `internal/generate/generate_test.go` that verifies:
- The "How to Work" section appears in generated CLAUDE.md output
- Tech Lead gets the orchestration heuristics ("You orchestrate, not implement")
- Code agents get the plan-first heuristics ("Read your assigned plan")
- The workspace-guide pointer appears when the skill is installed
- The workspace-guide pointer does NOT appear when the skill is not installed

Use the existing test patterns in the file — read the file to understand the test structure before writing.

### Step 4: Verify end-to-end

**4a.** Run `make build` — must pass with no errors.

**4b.** Run `go test ./...` — all tests pass.

**4c.** CLI smoke test in a temp directory:
```bash
mkdir /tmp/bonsai-test && cd /tmp/bonsai-test && git init
/path/to/bonsai init   # complete the flow
```
Then verify:
- The generated `CLAUDE.md` contains the "How to Work" section
- Tech Lead workspace has tech-lead-specific heuristics
- The workspace-guide skill appears in the skill picker during init

**4d.** Add a second agent:
```bash
/path/to/bonsai add   # add backend
```
Then verify:
- Backend workspace CLAUDE.md has code-agent heuristics (not tech-lead ones)
- workspace-guide skill appears if selected

## Dependencies

- No external dependencies — pure catalog content + generator modification
- No changes to config structs, lock file, or CLI commands
- The `WorkspaceClaudeMD` function signature does not change

## Security

> [!warning]
> Refer to `Playbook/Standards/SecurityStandards.md` for all security requirements.

- No user input flows into the How to Work section — all content is hardcoded in the generator
- Template rendering for workspace-guide.md.tmpl uses the existing safe `text/template` pipeline
- No new dependencies introduced

## Verification

- [ ] `make build` passes
- [ ] `go test ./...` passes (including new test cases)
- [ ] Tech Lead CLAUDE.md contains "How to Work" section with orchestration heuristics
- [ ] Backend/Frontend/Fullstack CLAUDE.md contains plan-first heuristics
- [ ] DevOps CLAUDE.md contains safety-first heuristics
- [ ] Security CLAUDE.md contains audit-and-report heuristics
- [ ] workspace-guide pointer appears when skill is installed
- [ ] workspace-guide pointer does NOT appear when skill is not installed
- [ ] workspace-guide.md.tmpl renders correctly for each agent type
- [ ] `bonsai catalog` shows workspace-guide skill
- [ ] Smoke test: `bonsai init` + `bonsai add` in temp directory produces correct output
